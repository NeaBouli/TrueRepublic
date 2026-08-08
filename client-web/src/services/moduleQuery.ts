import type { ChainConfig } from '@/types/chain';
import type { ChainDomain } from '@/types/chainData';
import type { ChainMerkleProof } from '@/types/chainData';

export const QUERY_PATHS = {
  truedemocracy: {
    domain: '/truedemocracy.Query/Domain',
    domains: '/truedemocracy.Query/Domains',
    validators: '/truedemocracy.Query/Validators',
    nullifier: '/truedemocracy.Query/Nullifier',
    merkleProof: '/truedemocracy.Query/MerkleProof',
    payToPut: '/truedemocracy.Query/PayToPut',
  },
  dex: {
    pool: '/dex.Query/Pool',
    pools: '/dex.Query/Pools',
    registeredAssets: '/dex.Query/RegisteredAssets',
    estimateSwap: '/dex.Query/EstimateSwap',
    poolStats: '/dex.Query/PoolStats',
    spotPrice: '/dex.Query/SpotPrice',
    lpPosition: '/dex.Query/LPPosition',
  },
} as const;

export type ModuleQueryField =
  | { number: number; type: 'string'; value: string }
  | { number: number; type: 'int64'; value: string };

export type ModuleQueryFailure =
  | 'transport'
  | 'remote'
  | 'protocol'
  | 'decode';

export class ModuleQueryError extends Error {
  public readonly originalCause?: unknown;

  constructor(
    public readonly path: string,
    public readonly failure: ModuleQueryFailure,
    message: string,
    cause?: unknown
  ) {
    super(`Module query ${path} failed: ${message}`);
    this.name = 'ModuleQueryError';
    this.originalCause = cause;
  }
}

export function expectQueryRecord(
  path: string,
  value: unknown
): Record<string, unknown> {
  if (typeof value !== 'object' || value === null || Array.isArray(value)) {
    throw new ModuleQueryError(path, 'decode', 'result must be an object');
  }
  return value as Record<string, unknown>;
}

export function expectQueryArray<T>(path: string, value: unknown): T[] {
  if (!Array.isArray(value)) {
    throw new ModuleQueryError(path, 'decode', 'result must be an array');
  }
  return value as T[];
}

export function expectQueryStringArray(
  path: string,
  field: string,
  value: unknown
): string[] {
  const result = expectQueryArray<unknown>(path, value);
  if (!result.every((entry) => typeof entry === 'string')) {
    throw new ModuleQueryError(path, 'decode', `${field} must contain strings`);
  }
  return result as string[];
}

export function expectQueryNumberArray(
  path: string,
  field: string,
  value: unknown
): number[] {
  const result = expectQueryArray<unknown>(path, value);
  if (!result.every((entry) => typeof entry === 'number' && Number.isFinite(entry))) {
    throw new ModuleQueryError(path, 'decode', `${field} must contain finite numbers`);
  }
  return result as number[];
}

export function expectQueryString(
  path: string,
  field: string,
  value: unknown
): string {
  if (typeof value !== 'string') {
    throw new ModuleQueryError(path, 'decode', `${field} must be a string`);
  }
  return value;
}

export function expectQueryNumber(
  path: string,
  field: string,
  value: unknown
): number {
  if (typeof value !== 'number' || !Number.isFinite(value)) {
    throw new ModuleQueryError(path, 'decode', `${field} must be a finite number`);
  }
  return value;
}

export function expectQueryBoolean(
  path: string,
  field: string,
  value: unknown
): boolean {
  if (typeof value !== 'boolean') {
    throw new ModuleQueryError(path, 'decode', `${field} must be a boolean`);
  }
  return value;
}

function expectNullableArray(
  path: string,
  field: string,
  value: unknown
): unknown[] {
  if (value === null) return [];
  if (!Array.isArray(value)) {
    throw new ModuleQueryError(path, 'decode', `${field} must be an array or null`);
  }
  return value;
}

export function expectChainDomain(path: string, value: unknown): ChainDomain {
  const domain = expectQueryRecord(path, value);
  expectQueryString(path, 'name', domain.name);
  expectQueryString(path, 'admin', domain.admin);
  expectQueryString(path, 'merkle_root', domain.merkle_root);
  for (const field of ['members', 'permission_reg', 'identity_commits']) {
    if (domain[field] !== null) expectQueryStringArray(path, field, domain[field]);
  }
  for (const [index, coinValue] of expectNullableArray(
    path,
    'treasury',
    domain.treasury
  ).entries()) {
    const coin = expectQueryRecord(path, coinValue);
    expectQueryString(path, `treasury[${index}].denom`, coin.denom);
    expectQueryString(path, `treasury[${index}].amount`, coin.amount);
  }
  for (const [issueIndex, issueValue] of expectNullableArray(
    path,
    'issues',
    domain.issues
  ).entries()) {
    const issue = expectQueryRecord(path, issueValue);
    expectQueryString(path, `issues[${issueIndex}].name`, issue.name);
    expectQueryNumber(path, `issues[${issueIndex}].stones`, issue.stones);
    expectQueryNumber(
      path,
      `issues[${issueIndex}].creation_date`,
      issue.creation_date
    );
    expectQueryString(
      path,
      `issues[${issueIndex}].external_link`,
      issue.external_link
    );
    for (const [suggestionIndex, suggestionValue] of expectNullableArray(
      path,
      `issues[${issueIndex}].suggestions`,
      issue.suggestions
    ).entries()) {
      const suggestion = expectQueryRecord(path, suggestionValue);
      for (const field of ['name', 'creator', 'color', 'external_link']) {
        expectQueryString(
          path,
          `issues[${issueIndex}].suggestions[${suggestionIndex}].${field}`,
          suggestion[field]
        );
      }
      expectQueryNumber(
        path,
        `issues[${issueIndex}].suggestions[${suggestionIndex}].stones`,
        suggestion.stones
      );
      expectQueryNumber(
        path,
        `issues[${issueIndex}].suggestions[${suggestionIndex}].creation_date`,
        suggestion.creation_date
      );
      for (const [ratingIndex, ratingValue] of expectNullableArray(
        path,
        `issues[${issueIndex}].suggestions[${suggestionIndex}].ratings`,
        suggestion.ratings
      ).entries()) {
        const rating = expectQueryRecord(path, ratingValue);
        expectQueryNumber(
          path,
          `issues[${issueIndex}].suggestions[${suggestionIndex}].ratings[${ratingIndex}].value`,
          rating.value
        );
      }
    }
  }
  return domain as unknown as ChainDomain;
}

export function expectChainMerkleProof(
  path: string,
  value: unknown
): ChainMerkleProof {
  const result = expectQueryRecord(path, value);
  const proof: ChainMerkleProof = {
    root: expectQueryString(path, 'root', result.root),
    commitment: expectQueryString(path, 'commitment', result.commitment),
    path_indices: expectQueryNumberArray(
      path,
      'path_indices',
      result.path_indices
    ),
    path_elements: expectQueryStringArray(
      path,
      'path_elements',
      result.path_elements
    ),
  };
  if (!/^[0-9a-f]{64}$/.test(proof.root)) {
    throw new ModuleQueryError(path, 'decode', 'root must be 64 lowercase hex characters');
  }
  if (!/^[0-9a-f]{64}$/.test(proof.commitment)) {
    throw new ModuleQueryError(
      path,
      'decode',
      'commitment must be 64 lowercase hex characters'
    );
  }
  if (
    proof.path_indices.length !== 20 ||
    !proof.path_indices.every((index) => Number.isInteger(index) && (index === 0 || index === 1))
  ) {
    throw new ModuleQueryError(path, 'decode', 'path_indices must contain 20 binary indices');
  }
  if (
    proof.path_elements.length !== 20 ||
    !proof.path_elements.every((element) => /^[0-9a-f]{64}$/.test(element))
  ) {
    throw new ModuleQueryError(
      path,
      'decode',
      'path_elements must contain 20 lowercase 32-byte hashes'
    );
  }
  return proof;
}

type Fetch = (
  input: RequestInfo | URL,
  init?: RequestInit
) => Promise<Response>;

interface RPCEnvelope {
  error?: { code?: number; message?: string; data?: string };
  result?: {
    response?: {
      code?: number | string;
      log?: string;
      value?: string;
    };
  };
}

function encodeVarint(value: bigint): number[] {
  if (value < 0n) throw new Error('negative protobuf varint');
  const encoded: number[] = [];
  do {
    let next = Number(value & 0x7fn);
    value >>= 7n;
    if (value !== 0n) next |= 0x80;
    encoded.push(next);
  } while (value !== 0n);
  return encoded;
}

function encodeRequest(fields: ModuleQueryField[]): Uint8Array {
  const encoded: number[] = [];
  for (const field of fields) {
    if (!Number.isInteger(field.number) || field.number < 1) {
      throw new Error(`invalid protobuf field number ${field.number}`);
    }
    if (field.type === 'string') {
      const value = new TextEncoder().encode(field.value);
      encoded.push(...encodeVarint(BigInt((field.number << 3) | 2)));
      encoded.push(...encodeVarint(BigInt(value.length)), ...value);
      continue;
    }
    if (!/^\d+$/.test(field.value)) {
      throw new Error(`field ${field.number} must be an unsigned decimal int64`);
    }
    const value = BigInt(field.value);
    if (value > 9_223_372_036_854_775_807n) {
      throw new Error(`field ${field.number} exceeds signed int64`);
    }
    encoded.push(...encodeVarint(BigInt(field.number << 3)));
    encoded.push(...encodeVarint(value));
  }
  return Uint8Array.from(encoded);
}

function readVarint(bytes: Uint8Array, offset: number): [bigint, number] {
  let result = 0n;
  let shift = 0n;
  for (let index = offset; index < bytes.length && index < offset + 10; index++) {
    const current = bytes[index];
    result |= BigInt(current & 0x7f) << shift;
    if ((current & 0x80) === 0) return [result, index + 1];
    shift += 7n;
  }
  throw new Error('invalid protobuf varint');
}

function decodeResult(bytes: Uint8Array): Uint8Array {
  let offset = 0;
  let result: Uint8Array | undefined;
  while (offset < bytes.length) {
    const [tag, next] = readVarint(bytes, offset);
    offset = next;
    const field = Number(tag >> 3n);
    const wire = Number(tag & 7n);
    if (wire === 0) {
      [, offset] = readVarint(bytes, offset);
      continue;
    }
    if (wire === 1) {
      if (offset + 8 > bytes.length) throw new Error('truncated fixed64 field');
      offset += 8;
      continue;
    }
    if (wire === 5) {
      if (offset + 4 > bytes.length) throw new Error('truncated fixed32 field');
      offset += 4;
      continue;
    }
    if (wire !== 2) throw new Error(`unsupported protobuf wire type ${wire}`);
    const [length, valueOffset] = readVarint(bytes, offset);
    const end = valueOffset + Number(length);
    if (end > bytes.length) throw new Error('truncated protobuf field');
    if (field === 1) result = bytes.slice(valueOffset, end);
    offset = end;
  }
  if (!result) throw new Error('query response has no result field');
  return result;
}

function bytesToHex(bytes: Uint8Array): string {
  return Array.from(bytes, (byte) => byte.toString(16).padStart(2, '0')).join('');
}

function base64ToBytes(value: string): Uint8Array {
  const decoded = globalThis.atob(value);
  return Uint8Array.from(decoded, (char) => char.charCodeAt(0));
}

export class ModuleQueryClient {
  private readonly fetchImpl: Fetch;

  constructor(
    private readonly config: Pick<ChainConfig, 'rpc'>,
    fetchImpl: Fetch = globalThis.fetch
  ) {
    this.fetchImpl = fetchImpl;
  }

  async query<T>(path: string, fields: ModuleQueryField[] = []): Promise<T> {
    let requestData: Uint8Array;
    try {
      requestData = encodeRequest(fields);
    } catch (error) {
      throw new ModuleQueryError(path, 'protocol', 'invalid request fields', error);
    }

    let response: Response;
    try {
      response = await this.fetchImpl(`${this.config.rpc}/`, {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({
          jsonrpc: '2.0',
          id: 1,
          method: 'abci_query',
          params: { path, data: bytesToHex(requestData), prove: false },
        }),
      });
    } catch (error) {
      throw new ModuleQueryError(path, 'transport', 'RPC request failed', error);
    }
    if (!response.ok) {
      throw new ModuleQueryError(
        path,
        'transport',
        `RPC returned HTTP ${response.status}`
      );
    }

    let envelope: RPCEnvelope;
    try {
      envelope = (await response.json()) as RPCEnvelope;
    } catch (error) {
      throw new ModuleQueryError(path, 'decode', 'RPC response is not JSON', error);
    }
    if (envelope.error) {
      const detail = envelope.error.data || envelope.error.message || 'unknown RPC error';
      throw new ModuleQueryError(path, 'remote', detail);
    }
    const abci = envelope.result?.response;
    if (!abci) {
      throw new ModuleQueryError(path, 'protocol', 'missing ABCI response');
    }
    const code = Number(abci.code ?? 0);
    if (!Number.isFinite(code) || code !== 0) {
      throw new ModuleQueryError(
        path,
        'remote',
        abci.log || `ABCI returned code ${String(abci.code)}`
      );
    }
    if (typeof abci.value !== 'string' || abci.value.length === 0) {
      throw new ModuleQueryError(path, 'protocol', 'missing protobuf response value');
    }

    try {
      const jsonBytes = decodeResult(base64ToBytes(abci.value));
      return JSON.parse(new TextDecoder().decode(jsonBytes)) as T;
    } catch (error) {
      throw new ModuleQueryError(path, 'decode', 'invalid typed query response', error);
    }
  }
}
