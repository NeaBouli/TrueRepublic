// GH-300 test-only browser prover Worker (module worker).
//
// Scope: a test-only Worker with fail-closed suppression of the enumerated
// capabilities below and pinned trusted Go glue. It is NOT a general sandbox
// for hostile code. The Go WASM runtime glue is loaded through one fixed
// same-origin static import (no importScripts, eval or dynamic URL); it runs
// before this module body, so the test integration verifies its pinned
// SHA-256 before the Worker can start. The enumerated network, storage and
// script-loading globals are then replaced and verified; if any of them stays
// usable the Worker refuses every request. It proves exactly one request,
// zeroes its artifact copies and closes; it has no signing, wallet, storage,
// RPC or broadcast surface.
import '/__zkp/wasm_exec.js';

const REQUEST_SCHEMA = 'truerepublic/zkp-worker-request/v1';
const RESULT_SCHEMA = 'truerepublic/zkp-worker-result/v1';
const MAX_REQUEST_JSON_CHARS = 64 * 1024;
const MAX_RESULT_JSON_CHARS = 64 * 1024;
const MAX_ERROR_CHARS = 512;
const MAX_ARTIFACT_BYTES = 128 * 1024 * 1024;
const REQUEST_KEYS = [
  'constraint_system',
  'proving_key',
  'request_id',
  'request_json',
  'schema',
  'verifying_key',
  'wasm',
];

const SUPPRESSED_CAPABILITIES = [
  'fetch',
  'XMLHttpRequest',
  'WebSocket',
  'EventSource',
  'WebTransport',
  'BroadcastChannel',
  'indexedDB',
  'caches',
  'importScripts',
];

// Names that could not be replaced or that remain usable; any entry blocks all requests.
const unsuppressed = [];
for (const name of SUPPRESSED_CAPABILITIES) {
  try {
    Object.defineProperty(self, name, { value: undefined, writable: false, configurable: false });
  } catch {
    // Verified below; a capability that survives fails the Worker closed.
  }
  if (typeof self[name] !== 'undefined') unsuppressed.push(name);
}

let handled = false;

function result(requestId, ok, resultJSON, error) {
  self.postMessage({
    schema: RESULT_SCHEMA,
    request_id: requestId,
    ok,
    result_json: resultJSON,
    error: error.slice(0, MAX_ERROR_CHARS),
  });
}

function exactKeys(value, keys) {
  return (
    value !== null &&
    typeof value === 'object' &&
    !Array.isArray(value) &&
    JSON.stringify(Object.keys(value).sort()) === JSON.stringify(keys)
  );
}

function boundedBytes(value) {
  return value instanceof Uint8Array && value.byteLength > 0 && value.byteLength <= MAX_ARTIFACT_BYTES;
}

function zero(request) {
  for (const key of ['constraint_system', 'proving_key', 'verifying_key', 'wasm']) {
    if (request?.[key] instanceof Uint8Array) request[key].fill(0);
  }
}

self.onmessage = async (event) => {
  const request = event.data;
  const requestId = Number.isSafeInteger(request?.request_id) ? request.request_id : 0;
  try {
    if (unsuppressed.length !== 0) {
      throw new Error(`browser worker capability suppression failed: ${unsuppressed.join(', ')}`);
    }
    if (handled) {
      throw new Error('browser worker accepts exactly one request');
    }
    handled = true;
    if (
      !exactKeys(request, REQUEST_KEYS) ||
      request.schema !== REQUEST_SCHEMA ||
      requestId <= 0 ||
      typeof request.request_json !== 'string' ||
      request.request_json.length === 0 ||
      request.request_json.length > MAX_REQUEST_JSON_CHARS ||
      !boundedBytes(request.constraint_system) ||
      !boundedBytes(request.proving_key) ||
      !boundedBytes(request.verifying_key) ||
      !boundedBytes(request.wasm)
    ) {
      throw new Error('browser worker received a malformed request envelope');
    }
    const GoRuntime = self.Go;
    if (typeof GoRuntime !== 'function') {
      throw new Error('Go WASM runtime is unavailable');
    }
    const go = new GoRuntime();
    const instantiated = await WebAssembly.instantiate(request.wasm, go.importObject);
    void go.run(instantiated.instance);
    for (let attempt = 0; attempt < 100; attempt += 1) {
      const prove = self.trueRepublicTestOnlyGroth16Prove;
      if (typeof prove === 'function') {
        const response = prove(
          request.request_json,
          request.constraint_system,
          request.proving_key,
          request.verifying_key
        );
        if (
          !exactKeys(response, ['error', 'ok', 'result']) ||
          typeof response.ok !== 'boolean' ||
          typeof response.result !== 'string' ||
          typeof response.error !== 'string' ||
          response.result.length > MAX_RESULT_JSON_CHARS
        ) {
          throw new Error('Go WASM prover returned a malformed response');
        }
        result(requestId, response.ok, response.result, response.error);
        return;
      }
      await new Promise((resolve) => setTimeout(resolve, 0));
    }
    throw new Error('Go WASM prover did not register');
  } catch (error) {
    result(requestId, false, '', error instanceof Error ? error.message : 'unknown browser worker error');
  } finally {
    zero(request);
    self.close();
  }
};
