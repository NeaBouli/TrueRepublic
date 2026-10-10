import { MEMBERSHIP_CIRCUIT_ID } from './zkpWasmProver';

/**
 * GH-300: strict, dependency-free v1 browser artifact contract and bounded
 * same-origin loader for the test-only Go/WASM prover. Every artifact is
 * synthetic and test-only; nothing here may load production artifacts.
 */
export const ZKP_ARTIFACT_MANIFEST_SCHEMA =
  'truerepublic/zkp-artifact-manifest/v1';
export const ZKP_ARTIFACT_CLASSIFICATION =
  'TEST-ONLY SINGLE-PARTY TOXIC WASTE';
export const MAX_ARTIFACT_PATH_LENGTH = 256;
export const MAX_ARTIFACT_SIZE_BYTES = 128 * 1024 * 1024;

/** Closed set of loader failure classes; every public loader failure uses one. */
export type ZKPArtifactErrorCode =
  | 'environment'
  | 'network'
  | 'http'
  | 'abort'
  | 'size'
  | 'integrity'
  | 'manifest'
  | 'destroyed';

/**
 * The only error type the artifact loader throws. Messages are deterministic
 * and never contain artifact bytes or response bodies; `cause` keeps the
 * original transport error where it helps diagnosis.
 */
export class ZKPArtifactError extends Error {
  readonly code: ZKPArtifactErrorCode;

  constructor(code: ZKPArtifactErrorCode, message: string, options?: { cause?: unknown }) {
    super(message);
    this.name = 'ZKPArtifactError';
    this.code = code;
    if (options && 'cause' in options) {
      Object.defineProperty(this, 'cause', { value: options.cause, enumerable: false });
    }
  }
}

function isAbort(error: unknown, signal: AbortSignal | undefined): boolean {
  return (
    signal?.aborted === true ||
    (typeof error === 'object' && error !== null && (error as { name?: unknown }).name === 'AbortError')
  );
}

/** Converts any non-loader failure into a typed transport error. */
function toArtifactError(label: string, action: string, error: unknown, signal: AbortSignal | undefined): ZKPArtifactError {
  if (error instanceof ZKPArtifactError) return error;
  return isAbort(error, signal)
    ? new ZKPArtifactError('abort', `${label}: ${action} was aborted`, { cause: error })
    : new ZKPArtifactError('network', `${label}: ${action} failed`, { cause: error });
}

export interface ZKPArtifactDescriptor {
  path: string;
  size_bytes: number;
  sha256_hex: string;
}

export interface ZKPArtifactManifest {
  schema: typeof ZKP_ARTIFACT_MANIFEST_SCHEMA;
  circuit_id: typeof MEMBERSHIP_CIRCUIT_ID;
  classification: typeof ZKP_ARTIFACT_CLASSIFICATION;
  production_allowed: false;
  constraint_system: ZKPArtifactDescriptor;
  proving_key: ZKPArtifactDescriptor;
  verifying_key: ZKPArtifactDescriptor;
  wasm: ZKPArtifactDescriptor;
}

const MANIFEST_KEYS = [
  'circuit_id',
  'classification',
  'constraint_system',
  'production_allowed',
  'proving_key',
  'schema',
  'verifying_key',
  'wasm',
];

const DESCRIPTOR_KEYS = ['path', 'sha256_hex', 'size_bytes'];

/** Parses and fully validates a v1 manifest; rejects anything malformed. */
export function parseZKPArtifactManifest(value: unknown): ZKPArtifactManifest {
  if (!isRecord(value) || !hasExactKeys(value, MANIFEST_KEYS)) {
    throw new ZKPArtifactError('manifest', 'artifact manifest contains missing or unknown fields');
  }
  if (value.schema !== ZKP_ARTIFACT_MANIFEST_SCHEMA) {
    throw new ZKPArtifactError('manifest', 'artifact manifest has an unsupported schema');
  }
  if (value.circuit_id !== MEMBERSHIP_CIRCUIT_ID) {
    throw new ZKPArtifactError('manifest', 'artifact manifest binds an unsupported circuit');
  }
  if (value.classification !== ZKP_ARTIFACT_CLASSIFICATION) {
    throw new ZKPArtifactError('manifest', 'artifact manifest has an unexpected classification');
  }
  if (value.production_allowed !== false) {
    throw new ZKPArtifactError('manifest', 'artifact manifest must set production_allowed to false');
  }
  const constraintSystem = parseDescriptor(
    value.constraint_system,
    'constraint_system'
  );
  const provingKey = parseDescriptor(value.proving_key, 'proving_key');
  const verifyingKey = parseDescriptor(value.verifying_key, 'verifying_key');
  const wasm = parseDescriptor(value.wasm, 'wasm');
  const descriptors = [constraintSystem, provingKey, verifyingKey, wasm];
  const paths = new Set<string>();
  const digests = new Set<string>();
  for (const descriptor of descriptors) {
    if (paths.has(descriptor.path)) {
      throw new ZKPArtifactError('manifest', 'artifact manifest reuses an artifact path');
    }
    if (digests.has(descriptor.sha256_hex)) {
      throw new ZKPArtifactError('manifest', 'artifact manifest reuses an artifact digest');
    }
    paths.add(descriptor.path);
    digests.add(descriptor.sha256_hex);
  }
  return {
    schema: ZKP_ARTIFACT_MANIFEST_SCHEMA,
    circuit_id: MEMBERSHIP_CIRCUIT_ID,
    classification: ZKP_ARTIFACT_CLASSIFICATION,
    production_allowed: false,
    constraint_system: constraintSystem,
    proving_key: provingKey,
    verifying_key: verifyingKey,
    wasm,
  };
}

/** JSON.parse wrapper that fails closed on invalid JSON. */
export function parseZKPArtifactManifestJSON(text: string): ZKPArtifactManifest {
  let value: unknown;
  try {
    value = JSON.parse(text);
  } catch {
    throw new ZKPArtifactError('manifest', 'artifact manifest is not valid JSON');
  }
  return parseZKPArtifactManifest(value);
}

function parseDescriptor(value: unknown, label: string): ZKPArtifactDescriptor {
  if (!isRecord(value) || !hasExactKeys(value, DESCRIPTOR_KEYS)) {
    throw new ZKPArtifactError('manifest', `${label}: descriptor contains missing or unknown fields`);
  }
  if (typeof value.path !== 'string') {
    throw new ZKPArtifactError('manifest', `${label}: path must be a string`);
  }
  validateArtifactPath(value.path, label);
  if (
    typeof value.size_bytes !== 'number' ||
    !Number.isSafeInteger(value.size_bytes) ||
    value.size_bytes <= 0 ||
    value.size_bytes > MAX_ARTIFACT_SIZE_BYTES
  ) {
    throw new ZKPArtifactError('manifest',
      `${label}: size_bytes must be a positive safe integer within the size bound`
    );
  }
  if (
    typeof value.sha256_hex !== 'string' ||
    !/^[0-9a-f]{64}$/u.test(value.sha256_hex)
  ) {
    throw new ZKPArtifactError('manifest', `${label}: sha256_hex must be 64 lowercase hex characters`);
  }
  return {
    path: value.path,
    size_bytes: value.size_bytes,
    sha256_hex: value.sha256_hex,
  };
}

/**
 * Only canonical relative same-origin paths are allowed: no traversal, no
 * backslashes, no absolute or scheme URLs, no credentials, no query or hash.
 */
function validateArtifactPath(path: string, label: string): void {
  if (path.length === 0 || path.length > MAX_ARTIFACT_PATH_LENGTH) {
    throw new ZKPArtifactError('manifest', `${label}: path length is outside the allowed bound`);
  }
  for (let index = 0; index < path.length; index += 1) {
    const code = path.charCodeAt(index);
    if (code < 0x20 || code === 0x7f) {
      throw new ZKPArtifactError('manifest', `${label}: path contains a control character`);
    }
  }
  if (
    path.startsWith('/') ||
    path.includes('://') ||
    path.includes(':') ||
    path.includes('\\') ||
    path.includes('?') ||
    path.includes('#') ||
    path.includes('@')
  ) {
    throw new ZKPArtifactError('manifest',
      `${label}: path must be a relative same-origin path without URL syntax`
    );
  }
  for (const segment of path.split('/')) {
    if (segment === '' || segment === '.' || segment === '..') {
      throw new ZKPArtifactError('manifest', `${label}: path contains a traversal or empty segment`);
    }
  }
}

export interface ZKPArtifactFetchInit {
  credentials: 'omit';
  redirect: 'error';
  cache: 'no-store';
  signal?: AbortSignal;
}

export interface ZKPArtifactReader {
  read(): Promise<{ done: boolean; value?: Uint8Array }>;
  cancel(): Promise<unknown>;
}

export interface ZKPArtifactResponse {
  ok: boolean;
  status: number;
  headers: { get(name: string): string | null };
  body: { getReader(): ZKPArtifactReader } | null;
}

export type ZKPArtifactFetch = (
  url: string,
  init: ZKPArtifactFetchInit
) => Promise<ZKPArtifactResponse>;

/** Minimal SHA-256 digester; `globalThis.crypto.subtle` satisfies it. */
export interface ZKPDigester {
  digest(algorithm: 'SHA-256', data: BufferSource): Promise<ArrayBuffer>;
}

export interface FetchZKPArtifactsOptions {
  fetchImpl?: ZKPArtifactFetch;
  baseOrigin?: string;
  signal?: AbortSignal;
  digester?: ZKPDigester;
}

/** Loaded artifacts that can be explicitly zeroed once no longer needed. */
export class LoadedZKPArtifacts {
  private destroyedFlag = false;

  constructor(
    private readonly constraintSystemBytes: Uint8Array,
    private readonly provingKeyBytes: Uint8Array,
    private readonly verifyingKeyBytes: Uint8Array,
    private readonly wasmBytes: Uint8Array
  ) {}

  get destroyed(): boolean {
    return this.destroyedFlag;
  }

  get constraintSystem(): Uint8Array {
    this.assertLive();
    return this.constraintSystemBytes;
  }

  get provingKey(): Uint8Array {
    this.assertLive();
    return this.provingKeyBytes;
  }

  get verifyingKey(): Uint8Array {
    this.assertLive();
    return this.verifyingKeyBytes;
  }

  get wasm(): Uint8Array {
    this.assertLive();
    return this.wasmBytes;
  }

  /** Zeroes every retained artifact byte array; idempotent. */
  destroy(): void {
    if (this.destroyedFlag) return;
    this.destroyedFlag = true;
    this.constraintSystemBytes.fill(0);
    this.provingKeyBytes.fill(0);
    this.verifyingKeyBytes.fill(0);
    this.wasmBytes.fill(0);
  }

  private assertLive(): void {
    if (this.destroyedFlag) {
      throw new ZKPArtifactError('destroyed', 'loaded artifacts have been destroyed');
    }
  }
}

/**
 * Loads all four artifacts sequentially from the same origin with bounded
 * reads and exact SHA-256 enforcement. Any failure zeroes every byte array
 * that was already loaded before the error is rethrown.
 */
export async function fetchZKPArtifacts(
  manifest: ZKPArtifactManifest,
  options: FetchZKPArtifactsOptions = {}
): Promise<LoadedZKPArtifacts> {
  const fetchImpl = options.fetchImpl ?? defaultFetch;
  const digester = options.digester ?? globalThis.crypto?.subtle;
  if (digester === undefined) {
    throw new ZKPArtifactError('environment', 'WebCrypto SHA-256 is unavailable in this environment');
  }
  const base = resolveBaseOrigin(options.baseOrigin);
  const loaded: Uint8Array[] = [];
  try {
    const constraintSystem = await loadArtifact(
      'constraint_system',
      manifest.constraint_system,
      base,
      fetchImpl,
      digester,
      options.signal
    );
    loaded.push(constraintSystem);
    const provingKey = await loadArtifact(
      'proving_key',
      manifest.proving_key,
      base,
      fetchImpl,
      digester,
      options.signal
    );
    loaded.push(provingKey);
    const verifyingKey = await loadArtifact(
      'verifying_key',
      manifest.verifying_key,
      base,
      fetchImpl,
      digester,
      options.signal
    );
    loaded.push(verifyingKey);
    const wasm = await loadArtifact(
      'wasm',
      manifest.wasm,
      base,
      fetchImpl,
      digester,
      options.signal
    );
    loaded.push(wasm);
    return new LoadedZKPArtifacts(
      constraintSystem,
      provingKey,
      verifyingKey,
      wasm
    );
  } catch (error: unknown) {
    for (const bytes of loaded) bytes.fill(0);
    throw toArtifactError('artifacts', 'loading', error, options.signal);
  }
}

async function loadArtifact(
  label: string,
  descriptor: ZKPArtifactDescriptor,
  base: URL,
  fetchImpl: ZKPArtifactFetch,
  digester: ZKPDigester,
  signal: AbortSignal | undefined
): Promise<Uint8Array> {
  const url = new URL(descriptor.path, `${base.origin}/`);
  if (url.origin !== base.origin) {
    throw new ZKPArtifactError('manifest', `${label}: path escapes the same-origin boundary`);
  }
  let response: ZKPArtifactResponse;
  try {
    response = await fetchImpl(url.href, {
      credentials: 'omit',
      redirect: 'error',
      cache: 'no-store',
      signal,
    });
  } catch (error: unknown) {
    throw toArtifactError(label, 'fetch', error, signal);
  }
  if (!response.ok) {
    throw new ZKPArtifactError('http', `${label}: fetch failed with status ${response.status}`);
  }
  enforceContentLength(label, descriptor, response.headers.get('content-length'));
  // No unbounded fallback: without a readable stream nothing is allocated.
  if (response.body === null || typeof response.body?.getReader !== 'function') {
    throw new ZKPArtifactError('network', `${label}: response has no readable body stream`);
  }
  const bytes = await readViaStream(label, descriptor, response.body, signal);
  if (bytes.byteLength !== descriptor.size_bytes) {
    bytes.fill(0);
    throw new ZKPArtifactError('size', `${label}: byte length does not match the manifest size`);
  }
  let digest: Uint8Array;
  try {
    digest = new Uint8Array(await digester.digest('SHA-256', bytes));
  } catch (error: unknown) {
    bytes.fill(0);
    throw new ZKPArtifactError('environment', `${label}: SHA-256 computation failed`, { cause: error });
  }
  if (toHex(digest) !== descriptor.sha256_hex) {
    bytes.fill(0);
    digest.fill(0);
    throw new ZKPArtifactError('integrity', `${label}: SHA-256 digest does not match the manifest`);
  }
  digest.fill(0);
  return bytes;
}

function enforceContentLength(
  label: string,
  descriptor: ZKPArtifactDescriptor,
  contentLength: string | null
): void {
  if (contentLength === null) return;
  if (!/^\d+$/u.test(contentLength)) {
    throw new ZKPArtifactError('size', `${label}: Content-Length header is malformed`);
  }
  const declared = Number(contentLength);
  if (
    !Number.isSafeInteger(declared) ||
    declared !== descriptor.size_bytes
  ) {
    throw new ZKPArtifactError('size',
      `${label}: Content-Length does not match the manifest size`
    );
  }
}

/** Allocates exactly the declared size and cancels the stream on overflow. */
async function readViaStream(
  label: string,
  descriptor: ZKPArtifactDescriptor,
  body: { getReader(): ZKPArtifactReader },
  signal: AbortSignal | undefined
): Promise<Uint8Array<ArrayBuffer>> {
  // Obtain the reader before allocating: a failing getReader() leaves no
  // artifact-sized buffer outside the zeroing boundary and is typed.
  let reader: ZKPArtifactReader;
  try {
    reader = body.getReader();
  } catch (error: unknown) {
    throw toArtifactError(label, 'opening the response stream', error, signal);
  }
  const buffer = new Uint8Array(descriptor.size_bytes);
  let offset = 0;
  try {
    for (;;) {
      const chunk = await reader.read();
      if (chunk.done) break;
      if (chunk.value === undefined || chunk.value.byteLength === 0) continue;
      if (offset + chunk.value.byteLength > buffer.byteLength) {
        try {
          await reader.cancel();
        } catch {
          // cancellation is best-effort; the overflow error is authoritative
        }
        throw new ZKPArtifactError('size', `${label}: stream exceeds the declared artifact size`);
      }
      buffer.set(chunk.value, offset);
      offset += chunk.value.byteLength;
    }
  } catch (error: unknown) {
    buffer.fill(0);
    throw toArtifactError(label, 'reading the response stream', error, signal);
  }
  if (offset !== buffer.byteLength) {
    buffer.fill(0);
    throw new ZKPArtifactError('size', `${label}: byte length does not match the manifest size`);
  }
  return buffer;
}

function resolveBaseOrigin(baseOrigin: string | undefined): URL {
  const candidate = baseOrigin ?? globalThis.location?.origin;
  if (candidate === undefined || candidate === 'null') {
    throw new ZKPArtifactError('environment', 'no same-origin base URL is available');
  }
  let parsed: URL;
  try {
    parsed = new URL(candidate);
  } catch {
    throw new ZKPArtifactError('environment', 'the base origin is not a valid URL');
  }
  if (parsed.origin === 'null') {
    throw new ZKPArtifactError('environment', 'the base origin must be an absolute origin');
  }
  return parsed;
}

function defaultFetch(
  url: string,
  init: ZKPArtifactFetchInit
): Promise<ZKPArtifactResponse> {
  return globalThis.fetch(url, init);
}

function toHex(bytes: Uint8Array): string {
  let out = '';
  for (const byte of bytes) {
    out += byte.toString(16).padStart(2, '0');
  }
  return out;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function hasExactKeys(value: Record<string, unknown>, keys: string[]): boolean {
  return JSON.stringify(Object.keys(value).sort()) === JSON.stringify(keys);
}
