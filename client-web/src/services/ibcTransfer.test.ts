// @vitest-environment node
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { toBech32 } from '@cosmjs/encoding';
import type { SigningStargateClient } from '@cosmjs/stargate';
import { MsgTransfer } from 'cosmjs-types/ibc/applications/transfer/v1/tx';
import { DEFAULT_CHAIN } from '@/config/chains';
import type { TransferChannel } from '@/types/ibc';
import {
  assertBech32Receiver,
  assertNativeDenom,
  assertTransferChannel,
  buildMsgTransferEncodeObject,
  DEFAULT_TIMEOUT_HEIGHT_MARGIN,
  DEFAULT_TIMEOUT_TIMESTAMP_MARGIN_NS,
  deriveTransferTimeouts,
  extractSendPacket,
  fetchClientLatestHeight,
  getCommittedSendPacket,
  IbcTransferError,
  isSelectableTransferChannel,
  MAX_TIMEOUT_HEIGHT_MARGIN,
  parseTransferAmount,
  reconcilePacketLifecycle,
  reducePacketLifecycle,
  sendIbcTransfer,
} from './ibcTransfer';
import { connectSigningClient, deliverMessages } from './signingClient';
import { createTxRegistry } from './txRegistry';

vi.mock('./signingClient', () => ({
  connectSigningClient: vi.fn(),
  deliverMessages: vi.fn(),
}));

type TestFetch = (
  input: RequestInfo | URL,
  init?: RequestInit
) => Promise<Response>;

const CHANNEL: TransferChannel = {
  portId: 'transfer',
  channelId: 'channel-0',
  counterpartyPortId: 'transfer',
  counterpartyChannelId: 'channel-7',
  connectionId: 'connection-0',
  version: 'ics20-1',
};

const RAW_CHANNEL = {
  channel_id: 'channel-0',
  port_id: 'transfer',
  state: 'STATE_OPEN',
  ordering: 'ORDER_UNORDERED',
  counterparty: { channel_id: 'channel-7', port_id: 'transfer' },
  connection_hops: ['connection-0'],
  version: 'ics20-1',
};

const RECEIVER = toBech32('cosmos', new Uint8Array(20).fill(7));
const SENDER = toBech32(DEFAULT_CHAIN.bech32Prefix, new Uint8Array(20).fill(3));

const TIMEOUTS = {
  timeoutHeight: { revisionNumber: 1n, revisionHeight: 1500n },
  timeoutTimestamp: 1_600_000_000_000n,
};

function sendPacketEvent(sequence: string, dstChannel = 'channel-7') {
  return {
    type: 'send_packet',
    attributes: [
      { key: 'packet_src_port', value: 'transfer' },
      { key: 'packet_src_channel', value: 'channel-0' },
      { key: 'packet_sequence', value: sequence },
      { key: 'packet_dst_port', value: 'transfer' },
      { key: 'packet_dst_channel', value: dstChannel },
    ],
  };
}

function expectKind(error: unknown, kind: string): void {
  expect(error).toBeInstanceOf(IbcTransferError);
  expect((error as IbcTransferError).kind).toBe(kind);
}

describe('parseTransferAmount', () => {
  it('converts strict positive decimals with BigInt arithmetic', () => {
    expect(parseTransferAmount('1', 6)).toBe('1000000');
    expect(parseTransferAmount('1.25', 6)).toBe('1250000');
    expect(parseTransferAmount('0.000001', 6)).toBe('1');
    expect(parseTransferAmount('21000000', 6)).toBe('21000000000000');
    expect(parseTransferAmount('7', 0)).toBe('7');
  });

  it.each(['0', '0.0', '0.000000'])('rejects zero amount %j', (amount) => {
    expect(() => parseTransferAmount(amount, 6)).toThrowError(IbcTransferError);
  });

  it.each([
    '-1',
    '+1',
    '1e3',
    'abc',
    '',
    ' 1',
    '1 ',
    '1.',
    '.5',
    '1.5.2',
    '1,5',
  ])('rejects non-strict-decimal amount %j', (amount) => {
    expect(() => parseTransferAmount(amount, 6)).toThrowError(IbcTransferError);
  });

  it('rejects fractions beyond the configured decimals', () => {
    expect(() => parseTransferAmount('1.0000001', 6)).toThrowError(
      IbcTransferError
    );
    expect(() => parseTransferAmount('1.5', 0)).toThrowError(IbcTransferError);
  });

  it('rejects out-of-range decimals', () => {
    expect(() => parseTransferAmount('1', -1)).toThrowError(IbcTransferError);
    expect(() => parseTransferAmount('1', 19)).toThrowError(IbcTransferError);
    expect(() => parseTransferAmount('1', 1.5)).toThrowError(IbcTransferError);
  });
});

describe('assertNativeDenom', () => {
  it('accepts the native upnyx denom', () => {
    expect(assertNativeDenom({ coinMinimalDenom: 'upnyx' })).toBe('upnyx');
  });

  it('rejects any other denom', () => {
    try {
      assertNativeDenom({ coinMinimalDenom: 'uatom' });
      expect.unreachable();
    } catch (error) {
      expectKind(error, 'invalid_denom');
    }
  });
});

describe('assertBech32Receiver', () => {
  it('accepts syntactically valid bech32 regardless of prefix', () => {
    // Syntax-only: a cosmos prefix must validate even though it is not the
    // home prefix; the prefix never proves the counterparty network.
    expect(assertBech32Receiver(RECEIVER)).toBe(RECEIVER);
    expect(assertBech32Receiver(SENDER)).toBe(SENDER);
  });

  it.each(['not-an-address', '', 'cosmos1', 'COSMOS1' + 'q'.repeat(38)])(
    'rejects invalid bech32 %j',
    (receiver) => {
      try {
        assertBech32Receiver(receiver);
        expect.unreachable();
      } catch (error) {
        expectKind(error, 'invalid_receiver');
      }
    }
  );
});

describe('transfer channel selectability', () => {
  it('accepts exactly the strict ICS-20 criteria', () => {
    expect(isSelectableTransferChannel(RAW_CHANNEL)).toBe(true);
    expect(assertTransferChannel(RAW_CHANNEL)).toEqual(CHANNEL);
  });

  it.each([
    ['state', { state: 'STATE_TRYOPEN' }],
    ['ordering', { ordering: 'ORDER_ORDERED' }],
    ['port', { port_id: 'wasm' }],
    ['version', { version: 'ics20-0' }],
    ['empty channel id', { channel_id: '' }],
    ['two hops', { connection_hops: ['connection-0', 'connection-1'] }],
    ['no hops', { connection_hops: [] }],
    ['empty hop', { connection_hops: [''] }],
    [
      'empty counterparty channel',
      { counterparty: { channel_id: '', port_id: 'transfer' } },
    ],
    [
      'empty counterparty port',
      { counterparty: { channel_id: 'channel-7', port_id: '' } },
    ],
  ])('rejects a channel with invalid %s', (_label, override) => {
    const channel = { ...RAW_CHANNEL, ...override };
    expect(isSelectableTransferChannel(channel)).toBe(false);
    try {
      assertTransferChannel(channel);
      expect.unreachable();
    } catch (error) {
      expectKind(error, 'invalid_channel');
    }
  });
});

describe('deriveTransferTimeouts', () => {
  const latest = { revisionNumber: 1n, revisionHeight: 500n };

  it('derives bounded timeouts from the client-state latest height', () => {
    const timeouts = deriveTransferTimeouts(latest, {
      nowUnixNs: 1_000_000_000_000n,
    });
    expect(timeouts).toEqual({
      timeoutHeight: {
        revisionNumber: 1n,
        revisionHeight: 500n + DEFAULT_TIMEOUT_HEIGHT_MARGIN,
      },
      timeoutTimestamp: 1_000_000_000_000n + DEFAULT_TIMEOUT_TIMESTAMP_MARGIN_NS,
    });
  });

  it('honors explicit margins exactly', () => {
    const timeouts = deriveTransferTimeouts(latest, {
      heightMargin: 7n,
      timestampMarginNs: 61_000_000_000n,
      nowUnixNs: 5n,
    });
    expect(timeouts.timeoutHeight.revisionHeight).toBe(507n);
    expect(timeouts.timeoutTimestamp).toBe(61_000_000_005n);
  });

  it.each([0n, -1n, MAX_TIMEOUT_HEIGHT_MARGIN + 1n])(
    'rejects height margin %s',
    (heightMargin) => {
      try {
        deriveTransferTimeouts(latest, { heightMargin });
        expect.unreachable();
      } catch (error) {
        expectKind(error, 'invalid_timeout');
      }
    }
  );

  it.each([1n, 86_400_000_000_001n])(
    'rejects timestamp margin %s',
    (timestampMarginNs) => {
      try {
        deriveTransferTimeouts(latest, { timestampMarginNs });
        expect.unreachable();
      } catch (error) {
        expectKind(error, 'invalid_timeout');
      }
    }
  );

  it('rejects a non-positive client-state latest height', () => {
    try {
      deriveTransferTimeouts({ revisionNumber: 1n, revisionHeight: 0n });
      expect.unreachable();
    } catch (error) {
      expectKind(error, 'invalid_client_state');
    }
  });
});

describe('fetchClientLatestHeight', () => {
  const connectionResponse = () =>
    Response.json({
      connection: { client_id: '07-tendermint-0', state: 'STATE_OPEN' },
    });
  const clientStateResponse = () =>
    Response.json({
      client_state: {
        '@type': '/ibc.lightclients.tendermint.v1.ClientState',
        latest_height: { revision_number: '1', revision_height: '500' },
      },
    });

  it('resolves the latest height through the connection hop', async () => {
    const fetchImpl = vi
      .fn<TestFetch>()
      .mockResolvedValueOnce(connectionResponse())
      .mockResolvedValueOnce(clientStateResponse());
    await expect(
      fetchClientLatestHeight(DEFAULT_CHAIN.rest, 'connection-0', fetchImpl)
    ).resolves.toEqual({ revisionNumber: 1n, revisionHeight: 500n });
    expect(fetchImpl.mock.calls[0][0]).toBe(
      `${DEFAULT_CHAIN.rest}/ibc/core/connection/v1/connections/connection-0`
    );
    expect(fetchImpl.mock.calls[1][0]).toBe(
      `${DEFAULT_CHAIN.rest}/ibc/core/client/v1/client_states/07-tendermint-0`
    );
  });

  it('rejects a malformed connection identifier', async () => {
    const fetchImpl = vi.fn<TestFetch>();
    await expect(
      fetchClientLatestHeight(DEFAULT_CHAIN.rest, '../../etc', fetchImpl)
    ).rejects.toMatchObject({ kind: 'invalid_channel' });
    expect(fetchImpl).not.toHaveBeenCalled();
  });

  it('fails typed on HTTP failure', async () => {
    const fetchImpl = vi.fn<TestFetch>(
      async () => new Response('down', { status: 502 })
    );
    await expect(
      fetchClientLatestHeight(DEFAULT_CHAIN.rest, 'connection-0', fetchImpl)
    ).rejects.toMatchObject({ kind: 'transport' });
  });

  it('fails typed on a non-JSON body', async () => {
    const fetchImpl = vi.fn<TestFetch>(
      async () => new Response('not json', { status: 200 })
    );
    await expect(
      fetchClientLatestHeight(DEFAULT_CHAIN.rest, 'connection-0', fetchImpl)
    ).rejects.toMatchObject({ kind: 'decode' });
  });

  it('fails typed on a missing latest height', async () => {
    const fetchImpl = vi
      .fn<TestFetch>()
      .mockResolvedValueOnce(connectionResponse())
      .mockResolvedValueOnce(
        Response.json({ client_state: { '@type': '/x' } })
      );
    await expect(
      fetchClientLatestHeight(DEFAULT_CHAIN.rest, 'connection-0', fetchImpl)
    ).rejects.toMatchObject({ kind: 'decode' });
  });

  it('fails typed on a non-digit revision height', async () => {
    const fetchImpl = vi
      .fn<TestFetch>()
      .mockResolvedValueOnce(connectionResponse())
      .mockResolvedValueOnce(
        Response.json({
          client_state: {
            latest_height: { revision_number: '1', revision_height: 'high' },
          },
        })
      );
    await expect(
      fetchClientLatestHeight(DEFAULT_CHAIN.rest, 'connection-0', fetchImpl)
    ).rejects.toMatchObject({ kind: 'decode' });
  });

  it('fails typed on a zero revision height', async () => {
    const fetchImpl = vi
      .fn<TestFetch>()
      .mockResolvedValueOnce(connectionResponse())
      .mockResolvedValueOnce(
        Response.json({
          client_state: {
            latest_height: { revision_number: '1', revision_height: '0' },
          },
        })
      );
    await expect(
      fetchClientLatestHeight(DEFAULT_CHAIN.rest, 'connection-0', fetchImpl)
    ).rejects.toMatchObject({ kind: 'invalid_client_state' });
  });
});

describe('buildMsgTransferEncodeObject', () => {
  const params = { channel: CHANNEL, receiver: RECEIVER, amount: '1.5' };

  it('builds the canonical MsgTransfer EncodeObject', () => {
    const message = buildMsgTransferEncodeObject(
      DEFAULT_CHAIN,
      SENDER,
      params,
      TIMEOUTS
    );
    expect(message.typeUrl).toBe('/ibc.applications.transfer.v1.MsgTransfer');
    expect(message.value).toEqual({
      sourcePort: 'transfer',
      sourceChannel: 'channel-0',
      token: { denom: 'upnyx', amount: '1500000' },
      sender: SENDER,
      receiver: RECEIVER,
      timeoutHeight: { revisionNumber: 1n, revisionHeight: 1500n },
      timeoutTimestamp: 1_600_000_000_000n,
      memo: '',
      encoding: '',
    });

    // The canonical registry must encode it and the codec must round-trip.
    const bytes = createTxRegistry().encode(message);
    expect(MsgTransfer.decode(bytes)).toEqual(message.value);
  });

  it('rejects a non-native configured denom', () => {
    try {
      buildMsgTransferEncodeObject(
        { ...DEFAULT_CHAIN, coinMinimalDenom: 'uatom' },
        SENDER,
        params,
        TIMEOUTS
      );
      expect.unreachable();
    } catch (error) {
      expectKind(error, 'invalid_denom');
    }
  });

  it('rejects an invalid amount', () => {
    try {
      buildMsgTransferEncodeObject(
        DEFAULT_CHAIN,
        SENDER,
        { ...params, amount: '0' },
        TIMEOUTS
      );
      expect.unreachable();
    } catch (error) {
      expectKind(error, 'invalid_amount');
    }
  });

  it('rejects an invalid receiver', () => {
    try {
      buildMsgTransferEncodeObject(
        DEFAULT_CHAIN,
        SENDER,
        { ...params, receiver: 'nope' },
        TIMEOUTS
      );
      expect.unreachable();
    } catch (error) {
      expectKind(error, 'invalid_receiver');
    }
  });

  it('rejects non-positive timeouts', () => {
    try {
      buildMsgTransferEncodeObject(DEFAULT_CHAIN, SENDER, params, {
        timeoutHeight: { revisionNumber: 1n, revisionHeight: 0n },
        timeoutTimestamp: 1n,
      });
      expect.unreachable();
    } catch (error) {
      expectKind(error, 'invalid_timeout');
    }
  });
});

describe('extractSendPacket', () => {
  it('evidences exactly one reconciled send_packet event', () => {
    const result = extractSendPacket(
      [
        { type: 'coin_spent', attributes: [] },
        sendPacketEvent('42'),
        {
          type: 'send_packet',
          attributes: [
            { key: 'packet_src_port', value: 'transfer' },
            { key: 'packet_src_channel', value: 'channel-99' },
          ],
        },
      ],
      CHANNEL
    );
    expect(result).toEqual({
      kind: 'evidenced',
      packet: {
        sequence: 42n,
        sourcePort: 'transfer',
        sourceChannel: 'channel-0',
        destinationPort: 'transfer',
        destinationChannel: 'channel-7',
      },
    });
  });

  it('reports absent when no matching event exists', () => {
    expect(extractSendPacket([], CHANNEL)).toEqual({ kind: 'absent' });
    expect(
      extractSendPacket([{ type: 'send_packet', attributes: [] }], CHANNEL)
    ).toEqual({ kind: 'absent' });
  });

  it('reports contradictory for two distinct matching sequences', () => {
    expect(
      extractSendPacket([sendPacketEvent('1'), sendPacketEvent('2')], CHANNEL)
    ).toEqual({ kind: 'contradictory' });
  });

  it('reports contradictory for duplicate matching send_packet events', () => {
    expect(
      extractSendPacket([sendPacketEvent('1'), sendPacketEvent('1')], CHANNEL)
    ).toEqual({ kind: 'contradictory' });
  });

  it('reports contradictory for a malformed matching event', () => {
    const missing = {
      type: 'send_packet',
      attributes: [
        { key: 'packet_src_port', value: 'transfer' },
        { key: 'packet_src_channel', value: 'channel-0' },
      ],
    };
    expect(extractSendPacket([missing], CHANNEL)).toEqual({
      kind: 'contradictory',
    });
  });

  it('reports contradictory when the counterparty does not reconcile', () => {
    expect(
      extractSendPacket([sendPacketEvent('1', 'channel-8')], CHANNEL)
    ).toEqual({ kind: 'contradictory' });
  });

  it('reports contradictory for duplicate attribute keys', () => {
    const duplicated = {
      type: 'send_packet',
      attributes: [
        { key: 'packet_src_port', value: 'transfer' },
        { key: 'packet_src_port', value: 'transfer' },
        { key: 'packet_src_channel', value: 'channel-0' },
        { key: 'packet_sequence', value: '1' },
        { key: 'packet_dst_port', value: 'transfer' },
        { key: 'packet_dst_channel', value: 'channel-7' },
      ],
    };
    expect(extractSendPacket([duplicated], CHANNEL)).toEqual({
      kind: 'contradictory',
    });
  });
});

describe('getCommittedSendPacket', () => {
  const HASH = 'A'.repeat(64);

  it('rejects a malformed transaction hash', async () => {
    const client = { getTx: vi.fn() };
    await expect(
      getCommittedSendPacket(client as never, 'xyz', CHANNEL)
    ).rejects.toMatchObject({ kind: 'evidence_unavailable' });
    expect(client.getTx).not.toHaveBeenCalled();
  });

  it('reports absent evidence for an unknown transaction', async () => {
    const client = { getTx: vi.fn(async () => null) };
    await expect(
      getCommittedSendPacket(client as never, HASH, CHANNEL)
    ).resolves.toEqual({
      txHash: HASH,
      height: 0,
      code: 0,
      extraction: { kind: 'absent' },
    });
  });

  it('extracts evidence from a committed transaction', async () => {
    const client = {
      getTx: vi.fn(async () => ({
        height: 100,
        hash: HASH,
        code: 0,
        events: [sendPacketEvent('9')],
      })),
    };
    const result = await getCommittedSendPacket(client as never, HASH, CHANNEL);
    expect(result.code).toBe(0);
    expect(result.extraction).toMatchObject({
      kind: 'evidenced',
      packet: { sequence: 9n },
    });
  });

  it('fails typed when the lookup transport fails', async () => {
    const client = {
      getTx: vi.fn(async () => {
        throw new Error('rpc down');
      }),
    };
    await expect(
      getCommittedSendPacket(client as never, HASH, CHANNEL)
    ).rejects.toMatchObject({ kind: 'transport' });
  });
});

describe('reducePacketLifecycle', () => {
  const packet = {
    sequence: 1n,
    sourcePort: 'transfer',
    sourceChannel: 'channel-0',
    destinationPort: 'transfer',
    destinationChannel: 'channel-7',
  };

  it('reduces insufficient evidence to unknown', () => {
    expect(
      reducePacketLifecycle({
        sendPacket: null,
        acknowledgement: null,
        timedOut: null,
      })
    ).toBe('unknown');
    // An acknowledgement without a committed send_packet is inconsistent.
    expect(
      reducePacketLifecycle({
        sendPacket: null,
        acknowledgement: 'success',
        timedOut: null,
      })
    ).toBe('unknown');
  });

  it('keeps a committed packet without outcome evidence pending relay', () => {
    expect(
      reducePacketLifecycle({
        sendPacket: packet,
        acknowledgement: null,
        timedOut: null,
      })
    ).toBe('committed_pending_relay');
    // Explicit source-chain evidence of no timeout is still pending, never
    // a refund claim.
    expect(
      reducePacketLifecycle({
        sendPacket: packet,
        acknowledgement: null,
        timedOut: false,
      })
    ).toBe('committed_pending_relay');
  });

  it('reduces acknowledgement evidence to terminal phases', () => {
    expect(
      reducePacketLifecycle({
        sendPacket: packet,
        acknowledgement: 'success',
        timedOut: null,
      })
    ).toBe('acknowledged');
    expect(
      reducePacketLifecycle({
        sendPacket: packet,
        acknowledgement: 'error',
        timedOut: null,
      })
    ).toBe('acknowledged_error_or_refunded');
  });

  it('reduces source-chain timeout evidence to timed_out_refunded', () => {
    expect(
      reducePacketLifecycle({
        sendPacket: packet,
        acknowledgement: null,
        timedOut: true,
      })
    ).toBe('timed_out_refunded');
  });

  it('reduces contradictory evidence to unknown', () => {
    expect(
      reducePacketLifecycle({
        sendPacket: packet,
        acknowledgement: 'success',
        timedOut: true,
      })
    ).toBe('unknown');
    expect(
      reducePacketLifecycle({
        sendPacket: packet,
        acknowledgement: 'error',
        timedOut: true,
      })
    ).toBe('unknown');
  });
});

describe('reconcilePacketLifecycle', () => {
  const packet = {
    sequence: 9n,
    sourcePort: 'transfer',
    sourceChannel: 'channel-0',
    destinationPort: 'transfer',
    destinationChannel: 'channel-7',
  };

  function lifecycleEvent(
    type: 'acknowledge_packet' | 'timeout_packet',
    acknowledgement?: string
  ) {
    return {
      type,
      attributes: [
        { key: 'packet_src_port', value: 'transfer' },
        { key: 'packet_src_channel', value: 'channel-0' },
        { key: 'packet_dst_port', value: 'transfer' },
        { key: 'packet_dst_channel', value: 'channel-7' },
        { key: 'packet_sequence', value: '9' },
        ...(acknowledgement === undefined
          ? []
          : [{ key: 'packet_ack', value: acknowledgement }]),
      ],
    };
  }

  function searchClient(
    acknowledgements: readonly ReturnType<typeof lifecycleEvent>[],
    timeouts: readonly ReturnType<typeof lifecycleEvent>[] = []
  ) {
    return {
      searchTx: vi
        .fn()
        .mockResolvedValueOnce(
          acknowledgements.length === 0
            ? []
            : [{ height: 10, hash: 'A'.repeat(64), code: 0, events: acknowledgements }]
        )
        .mockResolvedValueOnce(
          timeouts.length === 0
            ? []
            : [{ height: 11, hash: 'B'.repeat(64), code: 0, events: timeouts }]
        ),
    };
  }

  it('keeps an indexed packet pending when no terminal evidence exists', async () => {
    const evidence = await reconcilePacketLifecycle(searchClient([]), packet);
    expect(reducePacketLifecycle(evidence)).toBe('committed_pending_relay');
  });

  it.each([
    ['{"result":"AQ=="}', 'acknowledged'],
    [globalThis.btoa('{"error":"destination rejected"}'), 'acknowledged_error_or_refunded'],
  ])('classifies exact acknowledgement evidence %j', async (ack, phase) => {
    const evidence = await reconcilePacketLifecycle(
      searchClient([lifecycleEvent('acknowledge_packet', ack)]),
      packet
    );
    expect(reducePacketLifecycle(evidence)).toBe(phase);
  });

  it('prefers the canonical packet_ack_hex attribute', async () => {
    const event = lifecycleEvent('acknowledge_packet');
    event.attributes.push({
      key: 'packet_ack_hex',
      value: Buffer.from('{"result":"AQ=="}').toString('hex'),
    });
    const evidence = await reconcilePacketLifecycle(searchClient([event]), packet);
    expect(reducePacketLifecycle(evidence)).toBe('acknowledged');
  });

  it('classifies exact timeout evidence as refunded', async () => {
    const evidence = await reconcilePacketLifecycle(
      searchClient([], [lifecycleEvent('timeout_packet')]),
      packet
    );
    expect(reducePacketLifecycle(evidence)).toBe('timed_out_refunded');
  });

  it('fails closed on contradictory acknowledgement and timeout evidence', async () => {
    const evidence = await reconcilePacketLifecycle(
      searchClient(
        [lifecycleEvent('acknowledge_packet', '{"result":"AQ=="}')],
        [lifecycleEvent('timeout_packet')]
      ),
      packet
    );
    expect(reducePacketLifecycle(evidence)).toBe('unknown');
  });

  it('fails closed on an undecodable acknowledgement and on transport failure', async () => {
    const evidence = await reconcilePacketLifecycle(
      searchClient([lifecycleEvent('acknowledge_packet', 'not-an-ack')]),
      packet
    );
    expect(reducePacketLifecycle(evidence)).toBe('unknown');

    await expect(
      reconcilePacketLifecycle(
        { searchTx: vi.fn().mockRejectedValue(new Error('offline')) },
        packet
      )
    ).rejects.toMatchObject({ kind: 'transport' });
  });
});

describe('sendIbcTransfer', () => {
  const HASH = 'B'.repeat(64);

  beforeEach(() => {
    vi.mocked(connectSigningClient).mockReset();
    vi.mocked(deliverMessages).mockReset();
  });

  function fakeSigner(address = SENDER) {
    return {
      getAccounts: vi.fn(async () => [
        { address, pubkey: new Uint8Array(33), algo: 'secp256k1' as const },
      ]),
    };
  }

  function fakeClient(events: unknown[] | null) {
    return {
      getTx: vi.fn(async () =>
        events === null
          ? null
          : { height: 55, hash: HASH, code: 0, events }
      ),
      disconnect: vi.fn(),
    };
  }

  it('commits to committed_pending_relay with reconciled evidence', async () => {
    const client = fakeClient([sendPacketEvent('3')]);
    vi.mocked(connectSigningClient).mockResolvedValue(
      client as unknown as SigningStargateClient
    );
    vi.mocked(deliverMessages).mockResolvedValue({
      hash: HASH,
      height: 55,
      success: true,
    });
    const phases: string[] = [];

    const submission = await sendIbcTransfer(
      DEFAULT_CHAIN,
      fakeSigner() as never,
      { channel: CHANNEL, receiver: RECEIVER, amount: '2' },
      { timeouts: TIMEOUTS, onPhase: (phase) => phases.push(phase) }
    );

    expect(submission).toEqual({
      phase: 'committed_pending_relay',
      txHash: HASH,
      height: 55,
      packet: {
        sequence: 3n,
        sourcePort: 'transfer',
        sourceChannel: 'channel-0',
        destinationPort: 'transfer',
        destinationChannel: 'channel-7',
      },
    });
    expect(phases).toEqual([
      'validating',
      'signing',
      'broadcasting',
      'committed_pending_relay',
    ]);
    expect(client.disconnect).toHaveBeenCalledOnce();

    // The signer-derived sender, never a caller-supplied string, is signed.
    const [, senderAddress, messages] = vi.mocked(deliverMessages).mock
      .calls[0];
    expect(senderAddress).toBe(SENDER);
    expect(messages[0].value).toMatchObject({
      sender: SENDER,
      token: { denom: 'upnyx', amount: '2000000' },
    });
  });

  it('derives timeouts from the fetched client state when not provided', async () => {
    const client = fakeClient([sendPacketEvent('4')]);
    vi.mocked(connectSigningClient).mockResolvedValue(
      client as unknown as SigningStargateClient
    );
    vi.mocked(deliverMessages).mockResolvedValue({
      hash: HASH,
      height: 55,
      success: true,
    });
    const fetchImpl = vi
      .fn<TestFetch>()
      .mockResolvedValueOnce(
        Response.json({ connection: { client_id: '07-tendermint-0' } })
      )
      .mockResolvedValueOnce(
        Response.json({
          client_state: {
            latest_height: { revision_number: '2', revision_height: '900' },
          },
        })
      );

    await sendIbcTransfer(
      DEFAULT_CHAIN,
      fakeSigner() as never,
      { channel: CHANNEL, receiver: RECEIVER, amount: '1' },
      { fetchImpl, nowUnixNs: 10_000_000_000_000n }
    );

    const [, , messages] = vi.mocked(deliverMessages).mock.calls[0];
    expect(messages[0].value).toMatchObject({
      timeoutHeight: {
        revisionNumber: 2n,
        revisionHeight: 900n + DEFAULT_TIMEOUT_HEIGHT_MARGIN,
      },
      timeoutTimestamp:
        10_000_000_000_000n + DEFAULT_TIMEOUT_TIMESTAMP_MARGIN_NS,
    });
  });

  it('reports unknown when committed evidence is unavailable', async () => {
    const client = fakeClient(null);
    vi.mocked(connectSigningClient).mockResolvedValue(
      client as unknown as SigningStargateClient
    );
    vi.mocked(deliverMessages).mockResolvedValue({
      hash: HASH,
      height: 55,
      success: true,
    });

    const submission = await sendIbcTransfer(
      DEFAULT_CHAIN,
      fakeSigner() as never,
      { channel: CHANNEL, receiver: RECEIVER, amount: '2' },
      { timeouts: TIMEOUTS }
    );

    expect(submission.phase).toBe('unknown');
    expect(submission.packet).toBeNull();
    // The committed hash is preserved for manual evidence recovery.
    expect(submission.txHash).toBe(HASH);
  });

  it('reports unknown on contradictory committed evidence', async () => {
    const client = fakeClient([sendPacketEvent('1'), sendPacketEvent('2')]);
    vi.mocked(connectSigningClient).mockResolvedValue(
      client as unknown as SigningStargateClient
    );
    vi.mocked(deliverMessages).mockResolvedValue({
      hash: HASH,
      height: 55,
      success: true,
    });

    const submission = await sendIbcTransfer(
      DEFAULT_CHAIN,
      fakeSigner() as never,
      { channel: CHANNEL, receiver: RECEIVER, amount: '2' },
      { timeouts: TIMEOUTS }
    );
    expect(submission.phase).toBe('unknown');
    expect(submission.packet).toBeNull();
  });

  it('throws typed on broadcast failure and never resubmits', async () => {
    const client = fakeClient(null);
    vi.mocked(connectSigningClient).mockResolvedValue(
      client as unknown as SigningStargateClient
    );
    vi.mocked(deliverMessages).mockRejectedValue(
      new Error('insufficient funds')
    );

    await expect(
      sendIbcTransfer(
        DEFAULT_CHAIN,
        fakeSigner() as never,
        { channel: CHANNEL, receiver: RECEIVER, amount: '2' },
        { timeouts: TIMEOUTS }
      )
    ).rejects.toMatchObject({
      name: 'IbcTransferError',
      kind: 'broadcast',
    });
    expect(vi.mocked(deliverMessages)).toHaveBeenCalledOnce();
    expect(client.disconnect).toHaveBeenCalledOnce();
    expect(client.getTx).not.toHaveBeenCalled();
  });

  it('fails during validation before any signing', async () => {
    await expect(
      sendIbcTransfer(
        DEFAULT_CHAIN,
        fakeSigner() as never,
        { channel: CHANNEL, receiver: RECEIVER, amount: '0' },
        { timeouts: TIMEOUTS }
      )
    ).rejects.toMatchObject({ kind: 'invalid_amount' });
    expect(vi.mocked(connectSigningClient)).not.toHaveBeenCalled();
    expect(vi.mocked(deliverMessages)).not.toHaveBeenCalled();
  });
});
