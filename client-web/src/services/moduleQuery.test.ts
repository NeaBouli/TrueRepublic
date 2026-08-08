// @vitest-environment node
import { Buffer } from 'node:buffer';
import { describe, expect, it, vi } from 'vitest';
import { DEFAULT_CHAIN } from '@/config/chains';
import { DEXService } from './dex';
import { GovernanceService } from './governance';
import { GovernanceTxService } from './governanceTx';
import { NetworkService } from './network';
import { ZKPService } from './zkp';
import {
  expectChainMerkleProof,
  ModuleQueryClient,
  ModuleQueryError,
  QUERY_PATHS,
} from './moduleQuery';

type TestFetch = (
  input: RequestInfo | URL,
  init?: RequestInit
) => Promise<Response>;

function varint(value: number): number[] {
  const bytes: number[] = [];
  do {
    let byte = value & 0x7f;
    value >>>= 7;
    if (value > 0) byte |= 0x80;
    bytes.push(byte);
  } while (value > 0);
  return bytes;
}

function rpcResult(value: unknown): Response {
  const json = Buffer.from(JSON.stringify(value));
  const protobuf = Buffer.from([0x0a, ...varint(json.length), ...json]);
  return Response.json({
    jsonrpc: '2.0',
    id: 1,
    result: { response: { code: 0, value: protobuf.toString('base64') } },
  });
}

describe('ModuleQueryClient', () => {
  it('encodes a registered typed request and decodes Result JSON bytes', async () => {
    const fetchImpl = vi.fn<TestFetch>(async () => rpcResult({ used: true }));
    const client = new ModuleQueryClient(DEFAULT_CHAIN, fetchImpl);

    await expect(
      client.query<{ used: boolean }>(QUERY_PATHS.truedemocracy.nullifier, [
        { number: 1, type: 'string', value: 'Citizen' },
        { number: 2, type: 'string', value: 'aabb' },
      ])
    ).resolves.toEqual({ used: true });

    const [url, init] = fetchImpl.mock.calls[0];
    expect(url).toBe(`${DEFAULT_CHAIN.rpc}/`);
    const body = JSON.parse(String(init?.body));
    expect(body).toMatchObject({
      method: 'abci_query',
      params: {
        path: '/truedemocracy.Query/Nullifier',
        prove: false,
      },
    });
    expect(body.params.data).toBe('0a07436974697a656e120461616262');
  });

  it('distinguishes an RPC failure from an authoritative empty result', async () => {
    const failedFetch = vi.fn<TestFetch>(async () =>
      Response.json({
        jsonrpc: '2.0',
        id: 1,
        result: { response: { code: 1, log: 'route unavailable', value: '' } },
      })
    );
    const client = new ModuleQueryClient(DEFAULT_CHAIN, failedFetch);
    const service = new DEXService(DEFAULT_CHAIN, client);

    await expect(service.listPools()).rejects.toMatchObject({
      name: 'ModuleQueryError',
      failure: 'remote',
    });
  });

  it('rejects malformed response protobuf instead of manufacturing data', async () => {
    const fetchImpl = vi.fn<TestFetch>(async () =>
      Response.json({
        result: { response: { code: 0, value: Buffer.from([0x0a, 0x05]).toString('base64') } },
      })
    );
    const client = new ModuleQueryClient(DEFAULT_CHAIN, fetchImpl);

    await expect(client.query('/dex.Query/Pools')).rejects.toBeInstanceOf(
      ModuleQueryError
    );

    const valid = Buffer.from([0x0a, 0x02, 0x5b, 0x5d]);
    const trailingFixed64 = Buffer.concat([valid, Buffer.from([0x11, 0x00])]);
    const trailingFetch = vi.fn<TestFetch>(async () =>
      Response.json({
        result: {
          response: { code: 0, value: trailingFixed64.toString('base64') },
        },
      })
    );
    await expect(
      new ModuleQueryClient(DEFAULT_CHAIN, trailingFetch).query('/dex.Query/Pools')
    ).rejects.toBeInstanceOf(ModuleQueryError);
  });
});

describe('typed module response validation', () => {
  it('rejects malformed Merkle proof paths at the browser boundary', () => {
    expect(() =>
      expectChainMerkleProof(QUERY_PATHS.truedemocracy.merkleProof, {
        root: '00'.repeat(32),
        commitment: '11'.repeat(32),
        path_indices: Array(20).fill(0).map((value, index) =>
          index === 19 ? 2 : value
        ),
        path_elements: Array(20).fill('22'.repeat(32)),
      })
    ).toThrowError(ModuleQueryError);
  });

  it('rejects malformed nested DEX payloads instead of trusting type casts', async () => {
    const fetchImpl = vi.fn<TestFetch>(async () =>
      rpcResult([{ asset_denom: 'uatom', pnyx_reserve: null }])
    );
    const client = new ModuleQueryClient(DEFAULT_CHAIN, fetchImpl);

    await expect(new DEXService(DEFAULT_CHAIN, client).listPools()).rejects.toMatchObject({
      name: 'ModuleQueryError',
      failure: 'decode',
    });

    const validatorFetch = vi.fn<TestFetch>(async () =>
      rpcResult([{ operator_addr: 'truerepublic1validator', power: '100' }])
    );
    await expect(
      new NetworkService(
        DEFAULT_CHAIN,
        new ModuleQueryClient(DEFAULT_CHAIN, validatorFetch)
      ).getValidators()
    ).rejects.toMatchObject({ name: 'ModuleQueryError', failure: 'decode' });
  });
});

describe('registered module query services', () => {
  it('derives governance projections from the canonical Domain query', async () => {
    const fetchImpl = vi.fn<TestFetch>(async () =>
      rpcResult({
        name: 'Citizen',
        admin: 'truerepublic1admin',
        members: ['truerepublic1admin'],
        treasury: [{ denom: 'upnyx', amount: '1234' }],
        permission_reg: [],
        identity_commits: [],
        merkle_root: '',
        issues: [
          {
            name: 'Transport',
            stones: 2,
            creation_date: 10,
            external_link: 'https://example.invalid/issue',
            suggestions: [
              {
                name: 'Typed',
                creator: 'truerepublic1admin',
                stones: 3,
                ratings: [{ value: 4 }, { value: 2 }],
                color: 'green',
                creation_date: 11,
                external_link: '',
              },
            ],
          },
        ],
      })
    );
    const client = new ModuleQueryClient(DEFAULT_CHAIN, fetchImpl);
    const service = new GovernanceService(DEFAULT_CHAIN, client);

    await expect(service.listSuggestions('Citizen', 'Transport')).resolves.toEqual([
      expect.objectContaining({
        suggestionId: 'Typed',
        issueId: 'Transport',
        avgRating: 3,
        ratingCount: 2,
        greenStones: 3,
        zone: 'green',
      }),
    ]);
    const body = JSON.parse(String(fetchImpl.mock.calls[0][1]?.body));
    expect(body.params.path).toBe('/truedemocracy.Query/Domain');
  });

  it('rejects malformed nullifier and economic payloads instead of defaulting', async () => {
    const nullifierFetch = vi.fn<TestFetch>(async () => rpcResult({}));
    const nullifierClient = new ModuleQueryClient(DEFAULT_CHAIN, nullifierFetch);
    await expect(
      new ZKPService(DEFAULT_CHAIN, nullifierClient).isNullifierUsed(
        'Citizen',
        'aabb'
      )
    ).rejects.toMatchObject({ failure: 'decode' });

    const priceFetch = vi.fn<TestFetch>(async () =>
      rpcResult({
        base_cost: '1000',
        domain_multiplier: 1,
        final_cost: null,
        formula: 'canonical',
      })
    );
    const priceClient = new ModuleQueryClient(DEFAULT_CHAIN, priceFetch);
    await expect(
      new GovernanceTxService(DEFAULT_CHAIN, priceClient).calculatePayToPut(
        'Citizen'
      )
    ).rejects.toMatchObject({ failure: 'decode' });
  });
});
