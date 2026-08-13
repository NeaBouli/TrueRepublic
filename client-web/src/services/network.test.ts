// @vitest-environment node
import { describe, expect, it, vi } from 'vitest';
import { DEFAULT_CHAIN } from '@/config/chains';
import { IbcChannelError, NetworkService } from './network';

type TestFetch = (
  input: RequestInfo | URL,
  init?: RequestInit
) => Promise<Response>;

const OPEN_TRANSFER_CHANNEL = {
  channel_id: 'channel-0',
  port_id: 'transfer',
  state: 'STATE_OPEN',
  ordering: 'ORDER_UNORDERED',
  counterparty: { channel_id: 'channel-7', port_id: 'transfer' },
  connection_hops: ['connection-0'],
  version: 'ics20-1',
};

function channelsResponse(channels: unknown, pagination?: unknown): Response {
  return Response.json({ channels, pagination: pagination ?? null });
}

function serviceWith(fetchImpl: TestFetch): NetworkService {
  return new NetworkService(DEFAULT_CHAIN, undefined, fetchImpl);
}

describe('NetworkService.getIBCChannels (GH-190 fail-closed)', () => {
  it('returns schema-validated channels for a valid response', async () => {
    const fetchImpl = vi.fn<TestFetch>(async () =>
      channelsResponse([OPEN_TRANSFER_CHANNEL])
    );
    const channels = await serviceWith(fetchImpl).getIBCChannels();
    expect(channels).toEqual([OPEN_TRANSFER_CHANNEL]);
    expect(fetchImpl.mock.calls[0][0]).toBe(
      `${DEFAULT_CHAIN.rest}/ibc/core/channel/v1/channels`
    );
  });

  it('returns [] only for a valid empty chain answer', async () => {
    const fetchImpl = vi.fn<TestFetch>(async () => channelsResponse([]));
    await expect(serviceWith(fetchImpl).getIBCChannels()).resolves.toEqual([]);
  });

  it('throws a typed transport error on HTTP failure instead of []', async () => {
    const fetchImpl = vi.fn<TestFetch>(
      async () => new Response('nope', { status: 503 })
    );
    await expect(serviceWith(fetchImpl).getIBCChannels()).rejects.toMatchObject(
      { name: 'IbcChannelError', failure: 'transport' }
    );
  });

  it('throws a typed transport error on network failure instead of []', async () => {
    const fetchImpl = vi.fn<TestFetch>(async () => {
      throw new TypeError('fetch failed');
    });
    await expect(serviceWith(fetchImpl).getIBCChannels()).rejects.toMatchObject(
      { name: 'IbcChannelError', failure: 'transport' }
    );
  });

  it('throws a typed decode error on a non-JSON body instead of []', async () => {
    const fetchImpl = vi.fn<TestFetch>(
      async () => new Response('not json', { status: 200 })
    );
    await expect(serviceWith(fetchImpl).getIBCChannels()).rejects.toMatchObject(
      { name: 'IbcChannelError', failure: 'decode' }
    );
  });

  it('throws a typed decode error when channels is not an array', async () => {
    const fetchImpl = vi.fn<TestFetch>(async () =>
      Response.json({ channels: null })
    );
    await expect(serviceWith(fetchImpl).getIBCChannels()).rejects.toMatchObject(
      { name: 'IbcChannelError', failure: 'decode' }
    );
  });

  it('throws a typed decode error for a schema-violating entry', async () => {
    const fetchImpl = vi.fn<TestFetch>(async () =>
      channelsResponse([{ ...OPEN_TRANSFER_CHANNEL, ordering: 7 }])
    );
    await expect(serviceWith(fetchImpl).getIBCChannels()).rejects.toMatchObject(
      { name: 'IbcChannelError', failure: 'decode' }
    );
  });

  it('throws a typed decode error for non-string connection hops', async () => {
    const fetchImpl = vi.fn<TestFetch>(async () =>
      channelsResponse([{ ...OPEN_TRANSFER_CHANNEL, connection_hops: [1] }])
    );
    await expect(serviceWith(fetchImpl).getIBCChannels()).rejects.toMatchObject(
      { name: 'IbcChannelError', failure: 'decode' }
    );
  });

  it('rejects a paginated response instead of silently truncating', async () => {
    const fetchImpl = vi.fn<TestFetch>(async () =>
      channelsResponse([OPEN_TRANSFER_CHANNEL], { next_key: 'bW9yZQ==', total: '2' })
    );
    await expect(serviceWith(fetchImpl).getIBCChannels()).rejects.toMatchObject(
      { name: 'IbcChannelError', failure: 'protocol' }
    );
  });

  it('exposes a typed error class', () => {
    expect(new IbcChannelError('transport', 'x').name).toBe('IbcChannelError');
  });
});

describe('NetworkService.getTransferChannels (GH-190)', () => {
  it('narrows to the strict ICS-20 selectability criteria', async () => {
    const fetchImpl = vi.fn<TestFetch>(async () =>
      channelsResponse([
        OPEN_TRANSFER_CHANNEL,
        { ...OPEN_TRANSFER_CHANNEL, channel_id: 'channel-1', state: 'STATE_INIT' },
        { ...OPEN_TRANSFER_CHANNEL, channel_id: 'channel-2', ordering: 'ORDER_ORDERED' },
        { ...OPEN_TRANSFER_CHANNEL, channel_id: 'channel-3', port_id: 'wasm' },
        { ...OPEN_TRANSFER_CHANNEL, channel_id: 'channel-4', version: 'ics20-2' },
        {
          ...OPEN_TRANSFER_CHANNEL,
          channel_id: 'channel-5',
          connection_hops: ['connection-0', 'connection-1'],
        },
        {
          ...OPEN_TRANSFER_CHANNEL,
          channel_id: 'channel-6',
          counterparty: { channel_id: '', port_id: 'transfer' },
        },
      ])
    );
    const channels = await serviceWith(fetchImpl).getTransferChannels();
    expect(channels).toEqual([
      {
        portId: 'transfer',
        channelId: 'channel-0',
        counterpartyPortId: 'transfer',
        counterpartyChannelId: 'channel-7',
        connectionId: 'connection-0',
        version: 'ics20-1',
      },
    ]);
  });

  it('propagates the typed failure instead of returning []', async () => {
    const fetchImpl = vi.fn<TestFetch>(
      async () => new Response('down', { status: 500 })
    );
    await expect(
      serviceWith(fetchImpl).getTransferChannels()
    ).rejects.toMatchObject({ name: 'IbcChannelError', failure: 'transport' });
  });
});
