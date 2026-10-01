import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import process from 'node:process';
import { describe, expect, it } from 'vitest';
import type { TestOnlyZKPArtifacts } from './zkpWasmProver';
import {
  ZKPWorkerProverRuntime,
  ZKP_WORKER_MAX_ERROR_CHARS,
  ZKP_WORKER_MAX_REQUEST_JSON_CHARS,
  ZKP_WORKER_MAX_RESULT_JSON_CHARS,
  ZKP_WORKER_REQUEST_SCHEMA,
  ZKP_WORKER_RESULT_SCHEMA,
  type ZKPWorkerErrorEvent,
  type ZKPWorkerMessageEvent,
  type ZKPWorkerPort,
} from './zkpBrowserRuntime';

const artifacts: TestOnlyZKPArtifacts = {
  constraintSystem: new Uint8Array([1, 2, 3]),
  provingKey: new Uint8Array([4, 5, 6]),
  verifyingKey: new Uint8Array([7, 8, 9]),
};

class FakeWorker implements ZKPWorkerPort {
  onmessage: ((event: ZKPWorkerMessageEvent) => void) | null = null;
  onerror: ((event: ZKPWorkerErrorEvent) => void) | null = null;
  readonly posted: unknown[] = [];
  terminated = false;

  postMessage(message: unknown): void {
    this.posted.push(message);
  }

  terminate(): void {
    this.terminated = true;
  }

  reply(data: unknown): void {
    this.onmessage?.({ data });
  }

  fail(message: string): void {
    this.onerror?.({ message });
  }

  lastRequest(): Record<string, unknown> {
    const last = this.posted[this.posted.length - 1];
    if (typeof last !== 'object' || last === null) {
      throw new Error('no posted request');
    }
    return last as Record<string, unknown>;
  }
}

function makeFactory(): { factory: () => ZKPWorkerPort; workers: FakeWorker[] } {
  const workers: FakeWorker[] = [];
  return {
    workers,
    factory: () => {
      const worker = new FakeWorker();
      workers.push(worker);
      return worker;
    },
  };
}

function okResult(requestId: number, resultJSON: string): unknown {
  return {
    schema: ZKP_WORKER_RESULT_SCHEMA,
    request_id: requestId,
    ok: true,
    result_json: resultJSON,
    error: '',
  };
}

describe('GH-300 worker prover runtime', () => {
  it('proves through the worker with an exact request envelope', async () => {
    const { factory, workers } = makeFactory();
    const runtime = new ZKPWorkerProverRuntime(factory, artifacts);
    expect(workers).toHaveLength(0);

    const pending = runtime.prove('{"request":1}');
    expect(workers).toHaveLength(1);
    const worker = workers[0];
    expect(worker.posted).toHaveLength(1);
    const envelope = worker.lastRequest();
    expect(Object.keys(envelope).sort()).toEqual([
      'constraint_system',
      'proving_key',
      'request_id',
      'request_json',
      'schema',
      'verifying_key',
    ]);
    expect(envelope.schema).toBe(ZKP_WORKER_REQUEST_SCHEMA);
    expect(envelope.request_id).toBe(1);
    expect(envelope.request_json).toBe('{"request":1}');
    expect(envelope.constraint_system).toBe(artifacts.constraintSystem);
    expect(envelope.proving_key).toBe(artifacts.provingKey);
    expect(envelope.verifying_key).toBe(artifacts.verifyingKey);

    worker.reply(okResult(1, '{"result":true}'));
    await expect(pending).resolves.toBe('{"result":true}');
    expect(worker.terminated).toBe(true);
  });

  it('assigns monotonic request IDs across sequential proves', async () => {
    const { factory, workers } = makeFactory();
    const runtime = new ZKPWorkerProverRuntime(factory, artifacts);
    const first = runtime.prove('a');
    workers[0].reply(okResult(1, 'ra'));
    await expect(first).resolves.toBe('ra');

    const second = runtime.prove('b');
    expect(workers).toHaveLength(2);
    expect(workers[1].lastRequest().request_id).toBe(2);
    workers[1].reply(okResult(2, 'rb'));
    await expect(second).resolves.toBe('rb');
  });

  it('serializes concurrent proves through one in-flight slot', async () => {
    const { factory, workers } = makeFactory();
    const runtime = new ZKPWorkerProverRuntime(factory, artifacts);
    const first = runtime.prove('a');
    const second = runtime.prove('b');
    expect(workers[0].posted).toHaveLength(1);

    workers[0].reply(okResult(1, 'ra'));
    await expect(first).resolves.toBe('ra');
    expect(workers).toHaveLength(2);
    expect(workers[1].posted).toHaveLength(1);
    expect(workers[1].lastRequest().request_id).toBe(2);
    expect(workers[1].lastRequest().request_json).toBe('b');

    workers[1].reply(okResult(2, 'rb'));
    await expect(second).resolves.toBe('rb');
  });

  it('rejects concurrency overflow instead of queueing unboundedly', async () => {
    const { factory, workers } = makeFactory();
    const runtime = new ZKPWorkerProverRuntime(factory, artifacts);
    const first = runtime.prove('a');
    const second = runtime.prove('b');
    await expect(runtime.prove('c')).rejects.toThrow('busy');

    workers[0].reply(okResult(1, 'ra'));
    await expect(first).resolves.toBe('ra');
    workers[1].reply(okResult(2, 'rb'));
    await expect(second).resolves.toBe('rb');
  });

  it('rejects an empty prove request before touching the worker', async () => {
    const { factory, workers } = makeFactory();
    const runtime = new ZKPWorkerProverRuntime(factory, artifacts);
    await expect(runtime.prove('')).rejects.toThrow('non-empty');
    expect(workers).toHaveLength(0);
  });

  it('recycles after an unsolicited or duplicate worker reply', async () => {
    const { factory, workers } = makeFactory();
    const runtime = new ZKPWorkerProverRuntime(factory, artifacts);
    const first = runtime.prove('a');
    workers[0].reply(okResult(1, 'ra'));
    await expect(first).resolves.toBe('ra');

    // Duplicate reply for the settled request and a stale lower ID are ignored.
    workers[0].reply(okResult(1, 'duplicate'));
    workers[0].reply(okResult(0, 'stale'));

    const second = runtime.prove('b');
    expect(workers).toHaveLength(2);
    workers[1].reply(okResult(2, 'rb'));
    await expect(second).resolves.toBe('rb');
  });

  it('times out, terminates the worker and permits a clean retry', async () => {
    const { factory, workers } = makeFactory();
    const runtime = new ZKPWorkerProverRuntime(factory, artifacts, 5);
    const pending = runtime.prove('a');
    await expect(pending).rejects.toThrow('timed out');
    expect(workers[0].terminated).toBe(true);

    const next = runtime.prove('b');
    expect(workers).toHaveLength(2);
    workers[1].reply(okResult(2, 'rb'));
    await expect(next).resolves.toBe('rb');
  });

  it.each([0, -1, 1.5, Number.MAX_SAFE_INTEGER + 1])(
    'rejects an unsafe timeout value %s',
    (timeout) => {
      const { factory } = makeFactory();
      expect(
        () => new ZKPWorkerProverRuntime(factory, artifacts, timeout)
      ).toThrow('positive safe integer');
    }
  );

  it('rejects the in-flight call on a mismatched or stale reply ID', async () => {
    const { factory, workers } = makeFactory();
    const runtime = new ZKPWorkerProverRuntime(factory, artifacts);
    const pending = runtime.prove('a');
    workers[0].reply(okResult(99, 'wrong'));
    await expect(pending).rejects.toThrow('does not match');

    // The protocol-violating worker is recycled; the next prove uses a new one.
    const next = runtime.prove('b');
    expect(workers).toHaveLength(2);
    workers[1].reply(okResult(2, 'rb'));
    await expect(next).resolves.toBe('rb');
  });

  it.each([
    ['an extra key', { extra: true }],
    ['a missing key', { error: undefined }],
    ['a wrong schema', { schema: 'truerepublic/zkp-worker-result/v2' }],
    ['a non-numeric request id', { request_id: '1' }],
    ['a non-boolean ok flag', { ok: 'yes' }],
  ])('rejects a malformed result envelope with %s', async (_label, mutation) => {
    const { factory, workers } = makeFactory();
    const runtime = new ZKPWorkerProverRuntime(factory, artifacts);
    const pending = runtime.prove('a');
    const base = okResult(1, 'ra') as Record<string, unknown>;
    const mutated: Record<string, unknown> = { ...base, ...mutation };
    for (const key of Object.keys(mutated)) {
      if (mutated[key] === undefined) delete mutated[key];
    }
    workers[0].reply(mutated);
    await expect(pending).rejects.toThrow('malformed result envelope');
  });

  it('rejects an inconsistent result envelope', async () => {
    const { factory, workers } = makeFactory();
    const runtime = new ZKPWorkerProverRuntime(factory, artifacts);
    const pending = runtime.prove('a');
    workers[0].reply({
      schema: ZKP_WORKER_RESULT_SCHEMA,
      request_id: 1,
      ok: true,
      result_json: '',
      error: '',
    });
    await expect(pending).rejects.toThrow('inconsistent result envelope');
  });

  it('rejects with the worker error message on a failed prove', async () => {
    const { factory, workers } = makeFactory();
    const runtime = new ZKPWorkerProverRuntime(factory, artifacts);
    const pending = runtime.prove('a');
    workers[0].reply({
      schema: ZKP_WORKER_RESULT_SCHEMA,
      request_id: 1,
      ok: false,
      result_json: '',
      error: 'prover exploded',
    });
    await expect(pending).rejects.toThrow('prover exploded');

    const again = runtime.prove('b');
    expect(workers).toHaveLength(2);
    workers[1].reply({
      schema: ZKP_WORKER_RESULT_SCHEMA,
      request_id: 2,
      ok: false,
      result_json: '',
      error: '',
    });
    await expect(again).rejects.toThrow('rejected the request');
  });

  it('rejects the in-flight call on a worker error event and recovers', async () => {
    const { factory, workers } = makeFactory();
    const runtime = new ZKPWorkerProverRuntime(factory, artifacts);
    const pending = runtime.prove('a');
    workers[0].fail('wasm trap');
    await expect(pending).rejects.toThrow('worker error: wasm trap');
    expect(workers[0].terminated).toBe(true);

    const next = runtime.prove('b');
    expect(workers).toHaveLength(2);
    workers[1].reply(okResult(2, 'rb'));
    await expect(next).resolves.toBe('rb');
  });

  it('fails closed when the worker factory throws or returns junk', async () => {
    const throwing = new ZKPWorkerProverRuntime(() => {
      throw new Error('no workers');
    }, artifacts);
    await expect(throwing.prove('a')).rejects.toThrow('worker factory failed');

    const junk = new ZKPWorkerProverRuntime(
      () => ({}) as unknown as ZKPWorkerPort,
      artifacts
    );
    await expect(junk.prove('a')).rejects.toThrow('invalid port');
  });

  it('rejects in-flight and queued calls on destroy and detaches the worker', async () => {
    const { factory, workers } = makeFactory();
    const runtime = new ZKPWorkerProverRuntime(factory, artifacts);
    const first = runtime.prove('a');
    const second = runtime.prove('b');
    runtime.destroy();

    await expect(first).rejects.toThrow('destroyed');
    await expect(second).rejects.toThrow('destroyed');
    expect(runtime.isDestroyed).toBe(true);
    expect(workers[0].terminated).toBe(true);
    expect(workers[0].onmessage).toBeNull();
    expect(workers[0].onerror).toBeNull();

    // Late replies after destroy are inert.
    workers[0].reply(okResult(1, 'late'));
    await expect(runtime.prove('c')).rejects.toThrow('destroyed');
    runtime.destroy();
    expect(runtime.isDestroyed).toBe(true);
  });

  it('exposes no signing, RPC or broadcast capability', () => {
    const { factory } = makeFactory();
    const runtime = new ZKPWorkerProverRuntime(factory, artifacts);
    const surface = [
      ...Object.getOwnPropertyNames(runtime),
      ...Object.getOwnPropertyNames(Object.getPrototypeOf(runtime)),
    ];
    for (const name of surface) {
      expect(name.toLowerCase()).not.toMatch(/sign|broadcast|rpc|submit/);
    }
  });
});

// GH300B2: bounds, abort and the worker's capability contract.
describe('GH-300 worker runtime bounds, abort and capabilities', () => {
  it('rejects an oversized prove request before touching the worker', async () => {
    const { factory, workers } = makeFactory();
    const runtime = new ZKPWorkerProverRuntime(factory, artifacts);
    await expect(runtime.prove('x'.repeat(ZKP_WORKER_MAX_REQUEST_JSON_CHARS + 1))).rejects.toThrow(
      'prove request JSON must be a bounded non-empty string'
    );
    expect(workers).toHaveLength(0);
  });

  it('fails closed on an oversized result or error string and recycles the worker', async () => {
    for (const reply of [
      (id: number) => okResult(id, 'r'.repeat(ZKP_WORKER_MAX_RESULT_JSON_CHARS + 1)),
      (id: number) => ({
        schema: ZKP_WORKER_RESULT_SCHEMA,
        request_id: id,
        ok: false,
        result_json: '',
        error: 'e'.repeat(ZKP_WORKER_MAX_ERROR_CHARS + 1),
      }),
    ]) {
      const { factory, workers } = makeFactory();
      const runtime = new ZKPWorkerProverRuntime(factory, artifacts);
      const pending = runtime.prove('{"x":1}');
      workers[0].reply(reply(Number(workers[0].lastRequest().request_id)));
      await expect(pending).rejects.toThrow('worker returned a malformed result envelope');
      expect(workers[0].terminated).toBe(true);
    }
  });

  it('terminates deterministically when aborted mid-flight and rejects queued calls', async () => {
    const { factory, workers } = makeFactory();
    const controller = new AbortController();
    const runtime = new ZKPWorkerProverRuntime(factory, artifacts, 60_000, controller.signal);
    const first = runtime.prove('{"x":1}');
    const second = runtime.prove('{"x":2}');
    controller.abort();
    await expect(first).rejects.toThrow('worker proof generation was aborted');
    await expect(second).rejects.toThrow('worker proof generation was aborted');
    expect(workers[0].terminated).toBe(true);
    expect(runtime.isDestroyed).toBe(true);
    await expect(runtime.prove('{"x":3}')).rejects.toThrow('runtime has been destroyed');
  });

  it('never creates a worker when constructed with an already aborted signal', async () => {
    const { factory, workers } = makeFactory();
    const controller = new AbortController();
    controller.abort();
    const runtime = new ZKPWorkerProverRuntime(factory, artifacts, 60_000, controller.signal);
    await expect(runtime.prove('{"x":1}')).rejects.toThrow('runtime has been destroyed');
    expect(workers).toHaveLength(0);
  });

  it('keeps the worker free of script loading, eval, network, storage and signing paths', () => {
    const source = readFileSync(resolve(process.cwd(), 'browser-quality/zkp-browser-worker.js'), 'utf8');
    const code = source.replace(/\/\/.*$/gmu, '');
    expect(code).toMatch(/^import '\/__zkp\/wasm_exec\.js';$/mu);
    expect((code.match(/^import /gmu) ?? []).length).toBe(1);
    for (const forbidden of [
      /importScripts\s*\(/u,
      /\beval\s*\(/u,
      /new Function\s*\(/u,
      /\bimport\s*\(/u,
      /fetch\s*\(/u,
      /XMLHttpRequest\s*\(/u,
      /new WebSocket/u,
      /localStorage|sessionStorage/u,
      /indexedDB\.open/u,
    ]) {
      expect(code).not.toMatch(forbidden);
    }
    // The capability names may appear only in the neutralization list; the
    // handling code after it must not mention signing, broadcast or wallets.
    const handling = code.slice(code.indexOf('let handled'));
    expect(handling.length).toBeGreaterThan(0);
    expect(handling).not.toMatch(/sign|broadcast|wallet|mnemonic/iu);
    // Network/storage globals are neutralized before any message is handled.
    expect(code.indexOf("'fetch'")).toBeLessThan(code.indexOf('self.onmessage'));
    expect(code).toContain('self.close()');
  });

  it('shares its request/result bounds with the worker source', () => {
    const source = readFileSync(resolve(process.cwd(), 'browser-quality/zkp-browser-worker.js'), 'utf8');
    expect(source).toContain('const MAX_REQUEST_JSON_CHARS = 64 * 1024;');
    expect(source).toContain('const MAX_RESULT_JSON_CHARS = 64 * 1024;');
    expect(source).toContain('const MAX_ERROR_CHARS = 512;');
    expect(ZKP_WORKER_MAX_REQUEST_JSON_CHARS).toBe(64 * 1024);
    expect(ZKP_WORKER_MAX_RESULT_JSON_CHARS).toBe(64 * 1024);
    expect(ZKP_WORKER_MAX_ERROR_CHARS).toBe(512);
  });
});
