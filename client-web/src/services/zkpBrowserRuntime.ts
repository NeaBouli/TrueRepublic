import type {
  TestOnlyZKPArtifacts,
  TestOnlyZKPRuntime,
} from './zkpWasmProver';

/**
 * GH-300: deterministic coordinator that drives the test-only Go/WASM prover
 * inside an injected Worker-like port. The Worker is test-only with fail-closed
 * suppression of an enumerated set of capabilities and pinned trusted Go glue;
 * it is not a general sandbox for hostile code. No Worker global is touched at module
 * import or construction time; the factory runs lazily on the first prove.
 * The coordinator proves only: it has no signing, RPC, or broadcast surface.
 */
export const ZKP_WORKER_REQUEST_SCHEMA =
  'truerepublic/zkp-worker-request/v1';
export const ZKP_WORKER_RESULT_SCHEMA = 'truerepublic/zkp-worker-result/v1';

/** At most one in-flight prove plus one queued successor; overflow rejects. */
export const ZKP_WORKER_MAX_PENDING_PROOFS = 2;
export const ZKP_WORKER_PROOF_TIMEOUT_MS = 120_000;
/** Bounds shared with browser-quality/zkp-browser-worker.js. */
export const ZKP_WORKER_MAX_REQUEST_JSON_CHARS = 64 * 1024;
export const ZKP_WORKER_MAX_RESULT_JSON_CHARS = 64 * 1024;
export const ZKP_WORKER_MAX_ERROR_CHARS = 512;

export interface ZKPWorkerMessageEvent {
  readonly data: unknown;
}

export interface ZKPWorkerErrorEvent {
  readonly message?: string;
}

/** Minimal Worker-like port; a real Worker is adapted to this at the seam. */
export interface ZKPWorkerPort {
  onmessage: ((event: ZKPWorkerMessageEvent) => void) | null;
  onerror: ((event: ZKPWorkerErrorEvent) => void) | null;
  postMessage(message: unknown): void;
  terminate(): void;
}

export type ZKPWorkerFactory = () => ZKPWorkerPort;

export interface ZKPWorkerRequestEnvelope {
  schema: typeof ZKP_WORKER_REQUEST_SCHEMA;
  request_id: number;
  request_json: string;
  constraint_system: Uint8Array;
  proving_key: Uint8Array;
  verifying_key: Uint8Array;
}

export interface ZKPWorkerResultEnvelope {
  schema: typeof ZKP_WORKER_RESULT_SCHEMA;
  request_id: number;
  ok: boolean;
  result_json: string;
  error: string;
}

const RESULT_ENVELOPE_KEYS = [
  'error',
  'ok',
  'request_id',
  'result_json',
  'schema',
];

interface PendingProve {
  requestId: number;
  requestJSON: string;
  resolve: (resultJSON: string) => void;
  reject: (error: Error) => void;
  timeout: ReturnType<typeof setTimeout> | null;
}

export class ZKPWorkerProverRuntime implements TestOnlyZKPRuntime {
  private worker: ZKPWorkerPort | null = null;
  private destroyed = false;
  private nextRequestId = 0;
  private active: PendingProve | null = null;
  private queued: PendingProve | null = null;

  constructor(
    private readonly factory: ZKPWorkerFactory,
    private readonly artifacts: TestOnlyZKPArtifacts,
    private readonly timeoutMilliseconds = ZKP_WORKER_PROOF_TIMEOUT_MS,
    signal?: AbortSignal
  ) {
    if (
      !Number.isSafeInteger(timeoutMilliseconds) ||
      timeoutMilliseconds <= 0
    ) {
      throw new Error('worker proof timeout must be a positive safe integer');
    }
    if (signal?.aborted) {
      this.destroyWith(new Error('worker proof generation was aborted'));
    } else {
      // Aborting terminates the worker and rejects every pending call deterministically.
      signal?.addEventListener('abort', () => this.destroyWith(new Error('worker proof generation was aborted')), {
        once: true,
      });
    }
  }

  get isDestroyed(): boolean {
    return this.destroyed;
  }

  async prove(requestJSON: string): Promise<string> {
    if (this.destroyed) {
      throw new Error('runtime has been destroyed');
    }
    if (
      typeof requestJSON !== 'string' ||
      requestJSON.length === 0 ||
      requestJSON.length > ZKP_WORKER_MAX_REQUEST_JSON_CHARS
    ) {
      throw new Error('prove request JSON must be a bounded non-empty string');
    }
    if (this.active !== null && this.queued !== null) {
      throw new Error('runtime is busy: pending prove limit exceeded');
    }
    return new Promise<string>((resolve, reject) => {
      const call: PendingProve = {
        requestId: ++this.nextRequestId,
        requestJSON,
        resolve,
        reject,
        timeout: null,
      };
      if (this.active !== null) {
        this.queued = call;
        return;
      }
      this.active = call;
      this.dispatchActive();
    });
  }

  /** Rejects pending calls, terminates the worker and detaches listeners. */
  destroy(): void {
    this.destroyWith(new Error('runtime has been destroyed'));
  }

  private destroyWith(error: Error): void {
    if (this.destroyed) return;
    this.destroyed = true;
    const active = this.active;
    this.active = null;
    this.clearTimeout(active);
    active?.reject(error);
    const queued = this.queued;
    this.queued = null;
    queued?.reject(error);
    this.recycleWorker();
  }

  private dispatchActive(): void {
    const call = this.active;
    if (call === null) return;
    let worker: ZKPWorkerPort;
    try {
      worker = this.ensureWorker();
    } catch (error: unknown) {
      this.active = null;
      call.reject(
        error instanceof Error
          ? error
          : new Error('worker factory failed with an unknown error')
      );
      this.pumpQueue();
      return;
    }
    const envelope: ZKPWorkerRequestEnvelope = {
      schema: ZKP_WORKER_REQUEST_SCHEMA,
      request_id: call.requestId,
      request_json: call.requestJSON,
      constraint_system: this.artifacts.constraintSystem,
      proving_key: this.artifacts.provingKey,
      verifying_key: this.artifacts.verifyingKey,
    };
    try {
      worker.postMessage(envelope);
      call.timeout = setTimeout(() => {
        if (this.active !== call) return;
        this.active = null;
        this.recycleWorker();
        call.reject(new Error('worker proof generation timed out'));
        this.pumpQueue();
      }, this.timeoutMilliseconds);
    } catch (error: unknown) {
      this.active = null;
      this.recycleWorker();
      call.reject(
        new Error(
          `posting the prove request failed: ${
            error instanceof Error ? error.message : 'unknown error'
          }`
        )
      );
      this.pumpQueue();
    }
  }

  private ensureWorker(): ZKPWorkerPort {
    if (this.destroyed) {
      throw new Error('runtime has been destroyed');
    }
    if (this.worker !== null) return this.worker;
    let candidate: ZKPWorkerPort;
    try {
      candidate = this.factory();
    } catch (error: unknown) {
      throw new Error(
        `worker factory failed: ${
          error instanceof Error ? error.message : 'unknown error'
        }`
      );
    }
    if (
      candidate === null ||
      typeof candidate !== 'object' ||
      typeof candidate.postMessage !== 'function' ||
      typeof candidate.terminate !== 'function'
    ) {
      throw new Error('worker factory returned an invalid port');
    }
    // Callbacks are bound to this exact instance: a late message or error from
    // a recycled worker is inert and can never touch a newer request.
    candidate.onmessage = (event) => {
      if (this.worker !== candidate) return;
      this.handleMessage(event);
    };
    candidate.onerror = (event) => {
      if (this.worker !== candidate) return;
      this.handleWorkerError(event);
    };
    this.worker = candidate;
    return candidate;
  }

  private handleMessage(event: ZKPWorkerMessageEvent): void {
    const data = event.data;
    if (
      !isRecord(data) ||
      JSON.stringify(Object.keys(data).sort()) !==
        JSON.stringify(RESULT_ENVELOPE_KEYS) ||
      data.schema !== ZKP_WORKER_RESULT_SCHEMA ||
      typeof data.request_id !== 'number' ||
      !Number.isSafeInteger(data.request_id) ||
      typeof data.ok !== 'boolean' ||
      typeof data.result_json !== 'string' ||
      typeof data.error !== 'string' ||
      data.result_json.length > ZKP_WORKER_MAX_RESULT_JSON_CHARS ||
      data.error.length > ZKP_WORKER_MAX_ERROR_CHARS
    ) {
      this.failActiveProtocol('worker returned a malformed result envelope');
      return;
    }
    const active = this.active;
    if (active === null) {
      // An unsolicited or duplicate reply makes the worker untrustworthy.
      this.recycleWorker();
      return;
    }
    if (data.request_id !== active.requestId) {
      this.failActiveProtocol(
        'worker reply does not match the in-flight request'
      );
      return;
    }
    this.active = null;
    this.clearTimeout(active);
    // Always terminate after a proof attempt. This releases the isolated Go
    // runtime and its copies of the proving material before the next request.
    this.recycleWorker();
    if (!data.ok) {
      active.reject(
        new Error(data.error !== '' ? data.error : 'worker prover rejected the request')
      );
    } else if (data.error !== '' || data.result_json === '') {
      this.recycleWorker();
      active.reject(
        new Error('worker returned an inconsistent result envelope')
      );
    } else {
      active.resolve(data.result_json);
    }
    this.pumpQueue();
  }

  private handleWorkerError(event: ZKPWorkerErrorEvent): void {
    const active = this.active;
    this.active = null;
    this.clearTimeout(active);
    this.recycleWorker();
    active?.reject(
      new Error(
        `worker error: ${
          typeof event.message === 'string' && event.message !== ''
            ? event.message
            : 'unknown error'
        }`
      )
    );
    this.pumpQueue();
  }

  private failActiveProtocol(message: string): void {
    const active = this.active;
    if (active === null) return;
    this.active = null;
    this.clearTimeout(active);
    this.recycleWorker();
    active.reject(new Error(message));
    this.pumpQueue();
  }

  private pumpQueue(): void {
    if (this.destroyed || this.active !== null || this.queued === null) return;
    this.active = this.queued;
    this.queued = null;
    this.dispatchActive();
  }

  private recycleWorker(): void {
    const worker = this.worker;
    this.worker = null;
    if (worker === null) return;
    worker.onmessage = null;
    worker.onerror = null;
    try {
      worker.terminate();
    } catch {
      // termination is best-effort during teardown
    }
  }

  private clearTimeout(call: PendingProve | null): void {
    if (call?.timeout === null || call === null) return;
    clearTimeout(call.timeout);
    call.timeout = null;
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}
