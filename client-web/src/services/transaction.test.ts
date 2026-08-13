// @vitest-environment node
import { describe, expect, it, vi } from 'vitest';
import { toBech32 } from '@cosmjs/encoding';
import { DEFAULT_CHAIN } from '@/config/chains';
import {
  boundChainError,
  HISTORY_DEFAULT_PAGE_SIZE,
  HISTORY_MAX_PAGE_SIZE,
  TransactionService,
} from './transaction';

type TestFetch = (
  input: RequestInfo | URL,
  init?: RequestInit
) => Promise<Response>;

const ADDRESS = toBech32(
  DEFAULT_CHAIN.bech32Prefix,
  Uint8Array.from({ length: 20 }, (_, index) => index + 1)
);
const FOREIGN_ADDRESS = toBech32('cosmos', new Uint8Array(20).fill(9));
const LONG_ADDRESS = toBech32(
  DEFAULT_CHAIN.bech32Prefix,
  new Uint8Array(32).fill(3)
);

const HASH_A = 'AB'.repeat(32);
const HASH_B = 'CD'.repeat(32);

interface TxResponseOverrides {
  [key: string]: unknown;
}

function txResponse(overrides: TxResponseOverrides = {}): Record<string, unknown> {
  return {
    height: '42',
    txhash: HASH_A,
    code: 0,
    timestamp: '2026-08-08T12:34:56.000000000Z',
    raw_log: '[]',
    tx: {
      '@type': '/cosmos.tx.v1beta1.Tx',
      body: {
        messages: [
          {
            '@type': '/cosmos.bank.v1beta1.MsgSend',
            from_address: ADDRESS,
            to_address: FOREIGN_ADDRESS,
            amount: [{ denom: 'upnyx', amount: '7' }],
          },
        ],
        memo: 'hello chain',
      },
      auth_info: {
        fee: {
          amount: [{ denom: 'upnyx', amount: '12500' }],
          gas_limit: '200000',
        },
      },
    },
    ...overrides,
  };
}

function envelope(
  txResponses: unknown,
  total: string
): Record<string, unknown> {
  return {
    txs: [],
    tx_responses: txResponses,
    pagination: null,
    total,
  };
}

function jsonFetch(body: unknown): TestFetch {
  return vi.fn<TestFetch>(async () => Response.json(body));
}

describe('TransactionService.getSubmittedTransactions', () => {
  it('queries the REST tx service with the address query and newest-first pagination', async () => {
    const fetchImpl = jsonFetch(envelope([], '0'));
    const service = new TransactionService(DEFAULT_CHAIN, fetchImpl);

    await service.getSubmittedTransactions(ADDRESS);

    const [input, init] = (fetchImpl as ReturnType<typeof vi.fn>).mock.calls[0];
    const url = new URL(String(input));
    expect(url.origin + url.pathname).toBe(
      `${DEFAULT_CHAIN.rest}/cosmos/tx/v1beta1/txs`
    );
    expect(url.searchParams.get('query')).toBe(
      `tx.acc_seq CONTAINS '${ADDRESS}/'`
    );
    expect(url.searchParams.get('page')).toBe('1');
    expect(url.searchParams.get('limit')).toBe(
      String(HISTORY_DEFAULT_PAGE_SIZE)
    );
    expect(url.searchParams.get('order_by')).toBe('ORDER_BY_DESC');
    expect(init?.signal).toBeInstanceOf(AbortSignal);

    const globalFetch = vi
      .spyOn(globalThis, 'fetch')
      .mockImplementation(async function (this: unknown) {
        expect(this).toBe(globalThis);
        return Response.json(envelope([], '0'));
      });
    try {
      await new TransactionService(DEFAULT_CHAIN).getSubmittedTransactions(
        ADDRESS
      );
    } finally {
      globalFetch.mockRestore();
    }
  });

  it('maps pages to server page numbers and clamps the page size to the maximum', async () => {
    const fetchImpl = jsonFetch(envelope([], '0'));
    const service = new TransactionService(DEFAULT_CHAIN, fetchImpl);

    await service.getSubmittedTransactions(ADDRESS, 3, HISTORY_MAX_PAGE_SIZE);
    let url = new URL(
      String((fetchImpl as ReturnType<typeof vi.fn>).mock.calls[0][0])
    );
    expect(url.searchParams.get('page')).toBe('3');
    expect(url.searchParams.get('limit')).toBe(String(HISTORY_MAX_PAGE_SIZE));

    await service.getSubmittedTransactions(ADDRESS, 1, 500);
    url = new URL(
      String((fetchImpl as ReturnType<typeof vi.fn>).mock.calls[1][0])
    );
    expect(url.searchParams.get('limit')).toBe(String(HISTORY_MAX_PAGE_SIZE));

    await service.getSubmittedTransactions(ADDRESS, Number.NaN, 0);
    url = new URL(
      String((fetchImpl as ReturnType<typeof vi.fn>).mock.calls[2][0])
    );
    expect(url.searchParams.get('page')).toBe('1');
    expect(url.searchParams.get('limit')).toBe(
      String(HISTORY_DEFAULT_PAGE_SIZE)
    );
  });

  it('preserves successful and committed-failed entries with real fields', async () => {
    const failed = txResponse({
      height: '41',
      txhash: HASH_B,
      code: 5,
      raw_log: 'failed to execute message: insufficient funds',
      timestamp: '2026-08-08T12:30:00.000000000Z',
    });
    const service = new TransactionService(
      DEFAULT_CHAIN,
      jsonFetch(envelope([txResponse(), failed], '11'))
    );

    const page = await service.getSubmittedTransactions(ADDRESS, 1, 5);

    expect(page.address).toBe(ADDRESS);
    expect(page.page).toBe(1);
    expect(page.pageSize).toBe(5);
    expect(page.total).toBe(11);
    expect(page.hasMore).toBe(true);
    expect(page.transactions).toHaveLength(2);

    const [success, failure] = page.transactions;
    expect(success).toEqual({
      hash: HASH_A,
      height: 42,
      timestamp: '2026-08-08T12:34:56.000000000Z',
      code: 0,
      status: 'success',
      error: null,
      memo: 'hello chain',
      messages: [{ typeUrl: '/cosmos.bank.v1beta1.MsgSend' }],
      fee: [{ denom: 'upnyx', amount: '12500' }],
    });
    expect(failure.status).toBe('failed');
    expect(failure.code).toBe(5);
    expect(failure.error).toBe('failed to execute message: insufficient funds');
    expect(failure.hash).toBe(HASH_B);
    expect(failure.timestamp).toBe('2026-08-08T12:30:00.000000000Z');
  });

  it('treats a valid zero-total response as authoritative empty', async () => {
    const service = new TransactionService(
      DEFAULT_CHAIN,
      jsonFetch(envelope([], '0'))
    );

    const page = await service.getSubmittedTransactions(ADDRESS);

    expect(page).toEqual({
      address: ADDRESS,
      page: 1,
      pageSize: HISTORY_DEFAULT_PAGE_SIZE,
      total: 0,
      hasMore: false,
      transactions: [],
    });
  });

  it('reports unavailable for transport and HTTP failures', async () => {
    const offline = new TransactionService(
      DEFAULT_CHAIN,
      vi.fn<TestFetch>(async () => {
        throw new TypeError('fetch failed');
      })
    );
    await expect(offline.getSubmittedTransactions(ADDRESS)).rejects.toMatchObject({
      name: 'TransactionHistoryError',
      failure: 'unavailable',
    });

    const httpError = new TransactionService(
      DEFAULT_CHAIN,
      vi.fn<TestFetch>(async () => new Response('', { status: 503 }))
    );
    await expect(
      httpError.getSubmittedTransactions(ADDRESS)
    ).rejects.toMatchObject({
      name: 'TransactionHistoryError',
      failure: 'unavailable',
    });
  });

  it('reports a bounded timeout distinctly from unavailability', async () => {
    const hangingFetch = vi.fn<TestFetch>(
      async (_input, init) =>
        await new Promise<Response>((_resolve, reject) => {
          init?.signal?.addEventListener(
            'abort',
            () => reject(init.signal?.reason),
            { once: true }
          );
        })
    );
    const service = new TransactionService(DEFAULT_CHAIN, hangingFetch, 1);

    await expect(service.getSubmittedTransactions(ADDRESS)).rejects.toMatchObject({
      name: 'TransactionHistoryError',
      failure: 'timeout',
    });
  });

  it('honors a caller abort signal without reporting a timeout', async () => {
    const controller = new AbortController();
    controller.abort();
    const fetchImpl = vi.fn<TestFetch>(async (_input, init) => {
      throw (init?.signal?.reason ?? new Error('aborted')) as unknown;
    });
    const service = new TransactionService(DEFAULT_CHAIN, fetchImpl);

    await expect(
      service.getSubmittedTransactions(ADDRESS, 1, 20, controller.signal)
    ).rejects.toMatchObject({ name: 'AbortError' });
  });

  it('rejects malformed envelopes as protocol failures', async () => {
    const missingTotal = new TransactionService(
      DEFAULT_CHAIN,
      jsonFetch({ tx_responses: [] })
    );
    await expect(
      missingTotal.getSubmittedTransactions(ADDRESS)
    ).rejects.toMatchObject({ failure: 'protocol' });

    const notAnEnvelope = new TransactionService(DEFAULT_CHAIN, jsonFetch([]));
    await expect(
      notAnEnvelope.getSubmittedTransactions(ADDRESS)
    ).rejects.toMatchObject({ failure: 'protocol' });

    const countExceedsTotal = new TransactionService(
      DEFAULT_CHAIN,
      jsonFetch(envelope([txResponse()], '0'))
    );
    await expect(
      countExceedsTotal.getSubmittedTransactions(ADDRESS)
    ).rejects.toMatchObject({ failure: 'protocol' });

    const countExceedsPage = new TransactionService(
      DEFAULT_CHAIN,
      jsonFetch(envelope([txResponse(), txResponse({ txhash: HASH_B })], '2'))
    );
    await expect(
      countExceedsPage.getSubmittedTransactions(ADDRESS, 1, 1)
    ).rejects.toMatchObject({ failure: 'protocol' });
  });

  it('rejects malformed entries as decode failures', async () => {
    const cases: Record<string, unknown>[] = [
      { tx_responses: {}, total: '0' },
      envelope([txResponse({ txhash: 'zz' })], '1'),
      envelope([txResponse({ height: 'abc' })], '1'),
      envelope([txResponse({ code: '0' })], '1'),
      envelope([txResponse({ timestamp: 12 })], '1'),
      envelope(
        [
          txResponse({
            tx: { body: { messages: [{ no_type: true }] } },
          }),
        ],
        '1'
      ),
      envelope(
        [
          txResponse({
            tx: {
              body: { messages: [] },
              auth_info: { fee: { amount: [{ denom: 'upnyx', amount: 'x' }] } },
            },
          }),
        ],
        '1'
      ),
      envelope([], 'not-a-number'),
    ];

    for (const body of cases) {
      const service = new TransactionService(DEFAULT_CHAIN, jsonFetch(body));
      await expect(
        service.getSubmittedTransactions(ADDRESS)
      ).rejects.toMatchObject({
        name: 'TransactionHistoryError',
        failure: 'decode',
      });
    }
  });

  it('never fabricates absent timestamp, memo, or fee and keeps unknown types verbatim', async () => {
    const sparse = txResponse({ timestamp: undefined, raw_log: undefined });
    const tx = sparse.tx as Record<string, unknown>;
    const body = tx.body as Record<string, unknown>;
    delete body.memo;
    body.messages = [{ '@type': '/example.MsgUnsupported', payload: 1 }];
    delete tx.auth_info;

    const service = new TransactionService(
      DEFAULT_CHAIN,
      jsonFetch(envelope([sparse], '1'))
    );
    const page = await service.getSubmittedTransactions(ADDRESS);

    expect(page.transactions[0]).toEqual({
      hash: HASH_A,
      height: 42,
      timestamp: null,
      code: 0,
      status: 'success',
      error: null,
      memo: null,
      messages: [{ typeUrl: '/example.MsgUnsupported' }],
      fee: null,
    });
  });

  it('rejects invalid wallet addresses before any network request', async () => {
    const fetchImpl = jsonFetch(envelope([], '0'));
    const service = new TransactionService(DEFAULT_CHAIN, fetchImpl);

    for (const bad of [FOREIGN_ADDRESS, LONG_ADDRESS, 'not-an-address']) {
      await expect(
        service.getSubmittedTransactions(bad)
      ).rejects.toMatchObject({
        name: 'TransactionHistoryError',
        failure: 'protocol',
      });
    }
    expect(fetchImpl).not.toHaveBeenCalled();
  });
});

describe('boundChainError', () => {
  it('strips control characters and bounds the length', () => {
    const noisy = `line one\nline two\t[31m${'x'.repeat(500)}`;
    const bound = boundChainError(noisy, 7);

    expect(bound.length).toBeLessThanOrEqual(200);
    // eslint-disable-next-line no-control-regex
    expect(bound).not.toMatch(/[\x00-\x1f\x7f]/);
    expect(bound.startsWith('line one line two')).toBe(true);
  });

  it('falls back to the code when the log carries no readable text', () => {
    expect(boundChainError('\n\t', 11)).toBe('Transaction failed with code 11');
  });
});

describe('send input validation', () => {
  const stubWallet = {
    getAccounts: async () => [
      { address: ADDRESS, algo: 'secp256k1', pubkey: new Uint8Array(33) },
    ],
  } as unknown as import('@cosmjs/proto-signing').DirectSecp256k1HdWallet;

  it('rejects a foreign or malformed recipient before any network work', async () => {
    const service = new TransactionService(DEFAULT_CHAIN);

    for (const to of [FOREIGN_ADDRESS, LONG_ADDRESS, 'not-an-address']) {
      const result = await service.send(stubWallet, {
        to,
        amount: '1',
        denom: 'upnyx',
      });
      expect(result.success).toBe(false);
      expect(result.error).toContain('Recipient must be a valid');
    }
  });

  it('rejects non-positive or non-integer amounts before any network work', async () => {
    const service = new TransactionService(DEFAULT_CHAIN);

    for (const amount of ['0', '1.5', '-3', 'abc', '']) {
      const result = await service.send(stubWallet, {
        to: ADDRESS,
        amount,
        denom: 'upnyx',
      });
      expect(result.success).toBe(false);
      expect(result.error).toContain('positive integer');
    }
  });

  it('rejects an empty denom before any network work', async () => {
    const service = new TransactionService(DEFAULT_CHAIN);

    const result = await service.send(stubWallet, {
      to: ADDRESS,
      amount: '1',
      denom: '',
    });
    expect(result.success).toBe(false);
    expect(result.error).toContain('Denom is required');
  });
});
