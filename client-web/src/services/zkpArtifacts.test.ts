import { describe, expect, it, vi } from 'vitest';
import {
  MAX_ARTIFACT_PATH_LENGTH,
  MAX_ARTIFACT_SIZE_BYTES,
  parseZKPArtifactManifest,
  parseZKPArtifactManifestJSON,
  fetchZKPArtifacts,
  LoadedZKPArtifacts,
  ZKPArtifactError,
  ZKP_ARTIFACT_CLASSIFICATION,
  ZKP_ARTIFACT_MANIFEST_SCHEMA,
  type ZKPArtifactFetch,
  type ZKPArtifactFetchInit,
  type ZKPArtifactManifest,
  type ZKPArtifactResponse,
  type ZKPDigester,
} from './zkpArtifacts';
import { MEMBERSHIP_CIRCUIT_ID } from './zkpWasmProver';

const encoder = new TextEncoder();
const CS_BYTES = encoder.encode('TRUEREPUBLIC-TEST-CONSTRAINT-SYSTEM');
const PK_BYTES = encoder.encode('TRUEREPUBLIC-TEST-PROVING-KEY');
const VK_BYTES = encoder.encode('TRUEREPUBLIC-TEST-VERIFYING-KEY');
const WASM_BYTES = encoder.encode('TRUEREPUBLIC-TEST-WASM-BINARY');
const OTHER_BYTES = encoder.encode('TRUEREPUBLIC-TEST-OTHER');

const CS_DIGEST =
  '92f3644eee1250a9654ffea46bb79b19b2def4ba79a57655ece302ca09b019db';
const PK_DIGEST =
  '34b9c3ece3a83a28675cf191149804b00188b6cc982e78f77f7bb53e53ce956c';
const VK_DIGEST =
  '209e16a13bc9f2c4c47a8385d738d72ca86800e7bfe04aa706eb4cb86f124f2d';
const WASM_DIGEST =
  'fda5a1ab1a077b2727d9fd378d4ced10891edc858e2da3bb57753ee0ab6b57ab';
const OTHER_DIGEST =
  '4f3691747b23446ddd5fef0d9d6419bf06a0eb69608f5bb653b9649b1c910662';

const BASE_ORIGIN = 'https://wallet.truerepublic.test';

function validManifestPlain(): Record<string, unknown> {
  return {
    schema: ZKP_ARTIFACT_MANIFEST_SCHEMA,
    circuit_id: MEMBERSHIP_CIRCUIT_ID,
    classification: ZKP_ARTIFACT_CLASSIFICATION,
    production_allowed: false,
    constraint_system: {
      path: 'zkp/constraint-system.bin',
      size_bytes: CS_BYTES.byteLength,
      sha256_hex: CS_DIGEST,
    },
    proving_key: {
      path: 'zkp/proving-key.bin',
      size_bytes: PK_BYTES.byteLength,
      sha256_hex: PK_DIGEST,
    },
    verifying_key: {
      path: 'zkp/verifying-key.bin',
      size_bytes: VK_BYTES.byteLength,
      sha256_hex: VK_DIGEST,
    },
    wasm: {
      path: 'zkp/prover.wasm',
      size_bytes: WASM_BYTES.byteLength,
      sha256_hex: WASM_DIGEST,
    },
  };
}

function validManifest(): ZKPArtifactManifest {
  return parseZKPArtifactManifest(validManifestPlain());
}

interface RecordedCall {
  url: string;
  init: ZKPArtifactFetchInit;
}

interface FakeRouteOptions {
  chunkSize?: number;
  contentLength?: string | null;
  ok?: boolean;
  status?: number;
  streamBody?: boolean;
}

const arrayBufferTrap = vi.fn(() => Promise.reject(new Error('unbounded arrayBuffer read must not be used')));

function makeFetch(
  routes: Record<string, { bytes: Uint8Array; options?: FakeRouteOptions }>,
  calls: RecordedCall[]
): ZKPArtifactFetch {
  return (url, init) => {
    calls.push({ url, init });
    const route = routes[new URL(url).pathname.slice(1)];
    if (route === undefined) {
      return Promise.reject(new Error(`unexpected fetch of ${url}`));
    }
    const options = route.options ?? {};
    const ok = options.ok ?? true;
    const status = options.status ?? 200;
    const contentLength =
      options.contentLength === undefined
        ? String(route.bytes.byteLength)
        : options.contentLength;
    const response: ZKPArtifactResponse = {
      ok,
      status,
      headers: {
        get: (name: string) =>
          name.toLowerCase() === 'content-length' ? contentLength : null,
      },
      body:
        options.streamBody === false
          ? null
          : {
              getReader: () => {
                const chunkSize = options.chunkSize ?? route.bytes.byteLength;
                let offset = 0;
                return {
                  read: () => {
                    if (offset >= route.bytes.byteLength) {
                      return Promise.resolve({ done: true });
                    }
                    const end = Math.min(
                      offset + chunkSize,
                      route.bytes.byteLength
                    );
                    const value = route.bytes.slice(offset, end);
                    offset = end;
                    return Promise.resolve({ done: false, value });
                  },
                  cancel: () => Promise.resolve(undefined),
                };
              },
            },
    };
    // Trap: the loader must never fall back to an unbounded whole-body read.
    Object.assign(response, { arrayBuffer: arrayBufferTrap });
    return Promise.resolve(response);
  };
}

function standardRoutes(
  overrides: Record<string, { bytes: Uint8Array; options?: FakeRouteOptions }> = {}
): Record<string, { bytes: Uint8Array; options?: FakeRouteOptions }> {
  return {
    'zkp/constraint-system.bin': { bytes: CS_BYTES },
    'zkp/proving-key.bin': { bytes: PK_BYTES },
    'zkp/verifying-key.bin': { bytes: VK_BYTES },
    'zkp/prover.wasm': { bytes: WASM_BYTES },
    ...overrides,
  };
}

function makeCapturingDigester(captured: Uint8Array[]): ZKPDigester {
  return {
    digest: (_algorithm: 'SHA-256', data: BufferSource) => {
      if (data instanceof Uint8Array) captured.push(data);
      return globalThis.crypto.subtle.digest('SHA-256', data);
    },
  };
}

describe('GH-300 artifact manifest parsing', () => {
  it('parses a valid v1 manifest', () => {
    const manifest = validManifest();
    expect(manifest.schema).toBe(ZKP_ARTIFACT_MANIFEST_SCHEMA);
    expect(manifest.circuit_id).toBe(MEMBERSHIP_CIRCUIT_ID);
    expect(manifest.production_allowed).toBe(false);
    expect(manifest.wasm.path).toBe('zkp/prover.wasm');
  });

  it('parses from JSON text and rejects invalid JSON', () => {
    expect(
      parseZKPArtifactManifestJSON(JSON.stringify(validManifestPlain())).schema
    ).toBe(ZKP_ARTIFACT_MANIFEST_SCHEMA);
    expect(() => parseZKPArtifactManifestJSON('{not json')).toThrow(
      'not valid JSON'
    );
  });

  it('rejects unknown and missing top-level fields', () => {
    const extra = {
      ...validManifestPlain(),
      unexpected: true,
    };
    expect(() => parseZKPArtifactManifest(extra)).toThrow(
      'missing or unknown fields'
    );
    const missing = validManifestPlain();
    delete missing.wasm;
    expect(() => parseZKPArtifactManifest(missing)).toThrow(
      'missing or unknown fields'
    );
    expect(() => parseZKPArtifactManifest(null)).toThrow(
      'missing or unknown fields'
    );
    expect(() => parseZKPArtifactManifest([])).toThrow(
      'missing or unknown fields'
    );
  });

  it('rejects wrong schema, circuit, classification or production flag', () => {
    const wrongSchema = validManifestPlain();
    wrongSchema.schema = 'truerepublic/zkp-artifact-manifest/v2';
    expect(() => parseZKPArtifactManifest(wrongSchema)).toThrow(
      'unsupported schema'
    );

    const wrongCircuit = validManifestPlain();
    wrongCircuit.circuit_id = 'truerepublic/other-circuit/v1';
    expect(() => parseZKPArtifactManifest(wrongCircuit)).toThrow(
      'unsupported circuit'
    );

    const wrongClassification = validManifestPlain();
    wrongClassification.classification = 'production';
    expect(() => parseZKPArtifactManifest(wrongClassification)).toThrow(
      'classification'
    );

    const production = validManifestPlain();
    production.production_allowed = true;
    expect(() => parseZKPArtifactManifest(production)).toThrow(
      'production_allowed'
    );
  });

  it('rejects descriptors with unknown or missing fields', () => {
    const extra = validManifestPlain();
    extra.wasm = {
      ...(extra.wasm as Record<string, unknown>),
      label: 'extra',
    };
    expect(() => parseZKPArtifactManifest(extra)).toThrow(
      'missing or unknown fields'
    );
    const missing = validManifestPlain();
    missing.proving_key = { path: 'zkp/proving-key.bin' };
    expect(() => parseZKPArtifactManifest(missing)).toThrow(
      'missing or unknown fields'
    );
  });

  it('rejects malformed descriptor field types', () => {
    const wrongPath = validManifestPlain();
    wrongPath.wasm = {
      path: 42,
      size_bytes: WASM_BYTES.byteLength,
      sha256_hex: WASM_DIGEST,
    };
    expect(() => parseZKPArtifactManifest(wrongPath)).toThrow(
      'path must be a string'
    );
  });

  it.each([
    ['fractional', 1.5],
    ['zero', 0],
    ['negative', -4],
    ['unsafe integer', Number.MAX_SAFE_INTEGER + 1],
    ['over the size bound', MAX_ARTIFACT_SIZE_BYTES + 1],
  ])('rejects a %s size_bytes', (_label, size) => {
    const manifest = validManifestPlain();
    manifest.verifying_key = {
      path: 'zkp/verifying-key.bin',
      size_bytes: size,
      sha256_hex: VK_DIGEST,
    };
    expect(() => parseZKPArtifactManifest(manifest)).toThrow('size_bytes');
  });

  it.each([
    ['empty', ''],
    ['too long', 'a'.repeat(MAX_ARTIFACT_PATH_LENGTH + 1)],
    ['parent traversal', 'zkp/../secret.bin'],
    ['leading traversal', '../secret.bin'],
    ['dot segment', 'zkp/./prover.wasm'],
    ['empty segment', 'zkp//prover.wasm'],
    ['absolute path', '/zkp/prover.wasm'],
    ['backslash', 'zkp\\prover.wasm'],
    ['absolute URL', 'https://evil.example/prover.wasm'],
    ['protocol-relative URL', '//evil.example/prover.wasm'],
    ['scheme without slashes', 'javascript:alert(1)'],
    ['credentials', 'https://user:pw@wallet.truerepublic.test/prover.wasm'],
    ['query string', 'zkp/prover.wasm?v=1'],
    ['fragment', 'zkp/prover.wasm#frag'],
    ['control character', 'zkp/prover\n.wasm'],
  ])('rejects a path that is %s', (_label, path) => {
    const manifest = validManifestPlain();
    manifest.wasm = {
      path,
      size_bytes: WASM_BYTES.byteLength,
      sha256_hex: WASM_DIGEST,
    };
    expect(() => parseZKPArtifactManifest(manifest)).toThrow();
  });

  it.each([
    ['uppercase', WASM_DIGEST.toUpperCase()],
    ['too short', WASM_DIGEST.slice(0, 63)],
    ['non-hex', `${WASM_DIGEST.slice(0, 63)}x`],
  ])('rejects a %s digest', (_label, digest) => {
    const manifest = validManifestPlain();
    manifest.wasm = {
      path: 'zkp/prover.wasm',
      size_bytes: WASM_BYTES.byteLength,
      sha256_hex: digest,
    };
    expect(() => parseZKPArtifactManifest(manifest)).toThrow('sha256_hex');
  });

  it('rejects duplicate paths and duplicate digests', () => {
    const dupPath = validManifestPlain();
    dupPath.wasm = {
      path: 'zkp/proving-key.bin',
      size_bytes: WASM_BYTES.byteLength,
      sha256_hex: WASM_DIGEST,
    };
    expect(() => parseZKPArtifactManifest(dupPath)).toThrow('reuses');

    const dupDigest = validManifestPlain();
    dupDigest.wasm = {
      path: 'zkp/prover-copy.wasm',
      size_bytes: PK_BYTES.byteLength,
      sha256_hex: PK_DIGEST,
    };
    expect(() => parseZKPArtifactManifest(dupDigest)).toThrow('reuses');
  });
});

describe('GH-300 same-origin bounded artifact loading', () => {
  it('loads all artifacts with exact digests and safe fetch options', async () => {
    const calls: RecordedCall[] = [];
    const controller = new AbortController();
    const loaded = await fetchZKPArtifacts(validManifest(), {
      fetchImpl: makeFetch(standardRoutes(), calls),
      baseOrigin: BASE_ORIGIN,
      signal: controller.signal,
    });
    expect(calls).toHaveLength(4);
    for (const call of calls) {
      expect(call.url.startsWith(`${BASE_ORIGIN}/zkp/`)).toBe(true);
      expect(call.init.credentials).toBe('omit');
      expect(call.init.redirect).toBe('error');
      expect(call.init.cache).toBe('no-store');
      expect(call.init.signal).toBe(controller.signal);
    }
    expect(loaded.constraintSystem).toEqual(CS_BYTES);
    expect(loaded.provingKey).toEqual(PK_BYTES);
    expect(loaded.verifyingKey).toEqual(VK_BYTES);
    expect(loaded.wasm).toEqual(WASM_BYTES);
  });

  it('loads multi-chunk streams and fails closed for a response without a body stream', async () => {
    const calls: RecordedCall[] = [];
    const streamed = await fetchZKPArtifacts(validManifest(), {
      fetchImpl: makeFetch(
        standardRoutes({ 'zkp/prover.wasm': { bytes: WASM_BYTES, options: { chunkSize: 3 } } }),
        calls
      ),
      baseOrigin: BASE_ORIGIN,
    });
    expect(streamed.wasm).toEqual(WASM_BYTES);

    arrayBufferTrap.mockClear();
    const bodyless = fetchZKPArtifacts(validManifest(), {
      fetchImpl: makeFetch(
        standardRoutes({ 'zkp/prover.wasm': { bytes: WASM_BYTES, options: { streamBody: false } } }),
        []
      ),
      baseOrigin: BASE_ORIGIN,
    });
    await expect(bodyless).rejects.toBeInstanceOf(ZKPArtifactError);
    await expect(bodyless).rejects.toMatchObject({ code: 'network', message: 'wasm: response has no readable body stream' });
    expect(arrayBufferTrap).not.toHaveBeenCalled();
  });

  it('rejects a Content-Length that does not match the manifest', async () => {
    await expect(
      fetchZKPArtifacts(validManifest(), {
        fetchImpl: makeFetch(
          standardRoutes({
            'zkp/prover.wasm': {
              bytes: WASM_BYTES,
              options: { contentLength: String(WASM_BYTES.byteLength + 1) },
            },
          }),
          []
        ),
        baseOrigin: BASE_ORIGIN,
      })
    ).rejects.toThrow('Content-Length');

    await expect(
      fetchZKPArtifacts(validManifest(), {
        fetchImpl: makeFetch(
          standardRoutes({
            'zkp/prover.wasm': {
              bytes: WASM_BYTES,
              options: { contentLength: 'not-a-number' },
            },
          }),
          []
        ),
        baseOrigin: BASE_ORIGIN,
      })
    ).rejects.toThrow('Content-Length');
  });

  it('rejects a failed response status', async () => {
    await expect(
      fetchZKPArtifacts(validManifest(), {
        fetchImpl: makeFetch(
          standardRoutes({
            'zkp/prover.wasm': {
              bytes: WASM_BYTES,
              options: { ok: false, status: 404 },
            },
          }),
          []
        ),
        baseOrigin: BASE_ORIGIN,
      })
    ).rejects.toThrow('status 404');
  });

  it('rejects a body shorter or longer than the declared size', async () => {
    const short = WASM_BYTES.slice(0, WASM_BYTES.byteLength - 1);
    await expect(
      fetchZKPArtifacts(validManifest(), {
        fetchImpl: makeFetch(
          standardRoutes({
            'zkp/prover.wasm': {
              bytes: short,
              options: { contentLength: null },
            },
          }),
          []
        ),
        baseOrigin: BASE_ORIGIN,
      })
    ).rejects.toThrow('byte length');

    const longBytes = new Uint8Array(WASM_BYTES.byteLength + 1);
    longBytes.set(WASM_BYTES);
    await expect(
      fetchZKPArtifacts(validManifest(), {
        fetchImpl: makeFetch(
          standardRoutes({
            'zkp/prover.wasm': {
              bytes: longBytes,
              options: { contentLength: null },
            },
          }),
          []
        ),
        baseOrigin: BASE_ORIGIN,
      })
    ).rejects.toThrow('exceeds the declared artifact size');
  });

  it('zeroes the mismatched artifact and every earlier artifact', async () => {
    const wrongDigest = validManifestPlain();
    wrongDigest.wasm = {
      path: 'zkp/prover.wasm',
      size_bytes: WASM_BYTES.byteLength,
      sha256_hex: OTHER_DIGEST,
    };
    const captured: Uint8Array[] = [];
    await expect(
      fetchZKPArtifacts(parseZKPArtifactManifest(wrongDigest), {
        fetchImpl: makeFetch(standardRoutes(), []),
        baseOrigin: BASE_ORIGIN,
        digester: makeCapturingDigester(captured),
      })
    ).rejects.toThrow('SHA-256');
    expect(captured).toHaveLength(4);
    for (const bytes of captured) {
      expect(bytes.every((byte) => byte === 0)).toBe(true);
    }
  });

  it('zeroes earlier artifacts when a later artifact fails', async () => {
    const captured: Uint8Array[] = [];
    await expect(
      fetchZKPArtifacts(validManifest(), {
        fetchImpl: makeFetch(
          standardRoutes({
            'zkp/prover.wasm': {
              bytes: OTHER_BYTES,
              options: { contentLength: null },
            },
          }),
          []
        ),
        baseOrigin: BASE_ORIGIN,
        digester: makeCapturingDigester(captured),
      })
    ).rejects.toThrow('byte length');
    expect(captured).toHaveLength(3);
    for (const bytes of captured) {
      expect(bytes.every((byte) => byte === 0)).toBe(true);
    }
  });

  it('propagates an abort error and zeroes previously loaded artifacts', async () => {
    const controller = new AbortController();
    const captured: Uint8Array[] = [];
    const abortingFetch: ZKPArtifactFetch = (url, init) => {
      if (url.endsWith('verifying-key.bin')) {
        controller.abort();
        return Promise.reject(new Error('The operation was aborted'));
      }
      expect(init.signal?.aborted).toBe(false);
      return makeFetch(standardRoutes(), [])(url, init);
    };
    await expect(
      fetchZKPArtifacts(validManifest(), {
        fetchImpl: abortingFetch,
        baseOrigin: BASE_ORIGIN,
        signal: controller.signal,
        digester: makeCapturingDigester(captured),
      })
    ).rejects.toThrow('aborted');
    expect(captured).toHaveLength(2);
    for (const bytes of captured) {
      expect(bytes.every((byte) => byte === 0)).toBe(true);
    }
  });

  it('wraps a fetch transport failure as a typed network error keeping the cause', async () => {
    const transport = new Error('network down');
    const failing: ZKPArtifactFetch = () => Promise.reject(transport);
    const result = fetchZKPArtifacts(validManifest(), {
      fetchImpl: failing,
      baseOrigin: BASE_ORIGIN,
    });
    await expect(result).rejects.toBeInstanceOf(ZKPArtifactError);
    await expect(result).rejects.toMatchObject({ code: 'network', cause: transport });
  });

  it('destroy() zeroes every retained array and blocks further access', async () => {
    const loaded = await fetchZKPArtifacts(validManifest(), {
      fetchImpl: makeFetch(standardRoutes(), []),
      baseOrigin: BASE_ORIGIN,
    });
    const cs = loaded.constraintSystem;
    const pk = loaded.provingKey;
    const vk = loaded.verifyingKey;
    const wasm = loaded.wasm;
    loaded.destroy();
    expect(loaded.destroyed).toBe(true);
    for (const bytes of [cs, pk, vk, wasm]) {
      expect(bytes.every((byte) => byte === 0)).toBe(true);
    }
    expect(() => loaded.constraintSystem).toThrow('destroyed');
    loaded.destroy();
    expect(loaded.destroyed).toBe(true);
  });

  it('uses a digest-pinned route map and never leaves the origin', async () => {
    const calls: RecordedCall[] = [];
    await fetchZKPArtifacts(validManifest(), {
      fetchImpl: makeFetch(standardRoutes(), calls),
      baseOrigin: `${BASE_ORIGIN}/`,
    });
    const origins = calls.map((call) => new URL(call.url).origin);
    expect(new Set(origins)).toEqual(new Set([BASE_ORIGIN]));
  });
});

// GH300B1a: every public loader failure is a ZKPArtifactError with a closed code.
describe('GH-300 typed artifact loader failures', () => {
  async function failure(promise: Promise<unknown>): Promise<ZKPArtifactError> {
    const error = await promise.then(
      () => null,
      (caught: unknown) => caught
    );
    expect(error).toBeInstanceOf(ZKPArtifactError);
    return error as ZKPArtifactError;
  }

  const load = (fetchImpl: ZKPArtifactFetch, extra: Record<string, unknown> = {}) =>
    fetchZKPArtifacts(validManifest(), { fetchImpl, baseOrigin: BASE_ORIGIN, ...extra });

  it('classifies manifest, environment, http, size and integrity failures', async () => {
    expect((await failure(Promise.resolve().then(() => parseZKPArtifactManifestJSON('not json')))).code).toBe('manifest');
    expect(
      (await failure(Promise.resolve().then(() => parseZKPArtifactManifest({ ...validManifestPlain(), extra: 1 })))).code
    ).toBe('manifest');
    expect(
      (await failure(fetchZKPArtifacts(validManifest(), { fetchImpl: makeFetch(standardRoutes(), []), baseOrigin: 'null' }))).code
    ).toBe('environment');
    expect(
      (await failure(load(makeFetch(standardRoutes({ 'zkp/prover.wasm': { bytes: WASM_BYTES, options: { ok: false, status: 404 } } }), []))))
        .code
    ).toBe('http');
    expect(
      (await failure(load(makeFetch(standardRoutes({ 'zkp/prover.wasm': { bytes: WASM_BYTES, options: { contentLength: '1' } } }), []))))
        .code
    ).toBe('size');
    const tampered = new Uint8Array(WASM_BYTES);
    tampered[0] ^= 0x01;
    expect((await failure(load(makeFetch(standardRoutes({ 'zkp/prover.wasm': { bytes: tampered } }), []))) ).code).toBe('integrity');
  });

  it('wraps fetch rejections as network or abort with a safe cause', async () => {
    const transport = new TypeError('Failed to fetch');
    const network = await failure(load(() => Promise.reject(transport)));
    expect(network).toMatchObject({ code: 'network', message: 'constraint_system: fetch failed' });
    expect((network as Error & { cause?: unknown }).cause).toBe(transport);

    const controller = new AbortController();
    controller.abort();
    const aborted = await failure(
      load(() => Promise.reject(new DOMException('aborted', 'AbortError')), { signal: controller.signal })
    );
    expect(aborted.code).toBe('abort');
  });

  it('wraps stream read failures and zeroes partially read bytes', async () => {
    const streamFailure = new Error('connection reset');
    const fetchImpl: ZKPArtifactFetch = (url, init) =>
      makeFetch(standardRoutes(), [])(url, init).then((response) => ({
        ...response,
        body: { getReader: () => ({ read: () => Promise.reject(streamFailure), cancel: () => Promise.resolve(undefined) }) },
      }));
    const error = await failure(load(fetchImpl));
    expect(error).toMatchObject({ code: 'network', message: 'constraint_system: reading the response stream failed' });
    expect(error.message).not.toContain('TRUEREPUBLIC');
  });

  it('maps a failing digester to environment and destroyed artifacts to destroyed', async () => {
    const digester: ZKPDigester = { digest: () => Promise.reject(new Error('crypto unavailable')) };
    expect((await failure(load(makeFetch(standardRoutes(), []), { digester }))).code).toBe('environment');

    const loaded = await load(makeFetch(standardRoutes(), []));
    expect(loaded).toBeInstanceOf(LoadedZKPArtifacts);
    loaded.destroy();
    expect((await failure(Promise.resolve().then(() => loaded.wasm))).code).toBe('destroyed');
  });
});
