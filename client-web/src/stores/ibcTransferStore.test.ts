import { beforeEach, describe, expect, it, vi } from 'vitest';
import { IbcChannelError } from '@/services/network';

const mocks = vi.hoisted(() => {
  const walletState = {
    currentWallet: { address: 'truerepublic1sender', name: 'One', createdAt: 1 } as
      | { address: string; name: string; createdAt: number }
      | null,
    password: 'secret-password',
    isLocked: false,
    balances: [{ denom: 'upnyx', amount: '5000000' }],
    refreshBalance: vi.fn().mockResolvedValue(undefined),
  };
  return {
    walletState,
    getTransferChannels: vi.fn(),
    getWalletForSigning: vi.fn(),
    getBalance: vi.fn(),
    sendIbcTransfer: vi.fn(),
    reconcilePacketLifecycle: vi.fn(),
    connect: vi.fn(),
  };
});

vi.mock('./walletStore', () => ({
  useWalletStore: { getState: () => mocks.walletState },
}));

vi.mock('@/services/wallet', () => ({
  WalletService: { getWalletForSigning: mocks.getWalletForSigning },
}));

vi.mock('@/services/blockchain', () => ({
  BlockchainService: class {
    getBalance = mocks.getBalance;
  },
}));

vi.mock('@/services/network', async (importActual) => {
  const actual = await importActual<typeof import('@/services/network')>();
  return {
    ...actual,
    NetworkService: class {
      getTransferChannels = mocks.getTransferChannels;
    },
  };
});

vi.mock('@/services/ibcTransfer', async (importActual) => {
  const actual = await importActual<typeof import('@/services/ibcTransfer')>();
  return {
    ...actual,
    sendIbcTransfer: mocks.sendIbcTransfer,
    reconcilePacketLifecycle: mocks.reconcilePacketLifecycle,
  };
});

vi.mock('@cosmjs/stargate', async (importActual) => {
  const actual = await importActual<typeof import('@cosmjs/stargate')>();
  return {
    ...actual,
    StargateClient: { connect: mocks.connect },
  };
});

import {
  ibcTransferScopeKey,
  IbcTransferStoreError,
  useIbcTransferStore,
} from './ibcTransferStore';

const CHANNEL = {
  portId: 'transfer',
  channelId: 'channel-0',
  counterpartyPortId: 'transfer',
  counterpartyChannelId: 'channel-7',
  connectionId: 'connection-0',
  version: 'ics20-1',
};
const HASH = 'A'.repeat(64);
const PACKET = {
  sequence: 9n,
  sourcePort: 'transfer',
  sourceChannel: 'channel-0',
  destinationPort: 'transfer',
  destinationChannel: 'channel-7',
};

function resetStore() {
  useIbcTransferStore.setState({
    channels: [],
    channelStatus: 'idle',
    channelFailure: null,
    channelError: null,
    submissionStatus: 'idle',
    submissionPhase: null,
    submissionError: null,
    activeTxHash: null,
    reconcileStatus: 'idle',
    reconcileError: null,
    reconcileTxHash: null,
    recordsByScope: {},
    sessionGeneration: 0,
  });
}

describe('IBC transfer store', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
    Object.assign(mocks.walletState, {
      currentWallet: { address: 'truerepublic1sender', name: 'One', createdAt: 1 },
      password: 'secret-password',
      isLocked: false,
      balances: [{ denom: 'upnyx', amount: '5000000' }],
    });
    mocks.walletState.refreshBalance.mockResolvedValue(undefined);
    mocks.getWalletForSigning.mockResolvedValue({ signer: true });
    mocks.getBalance.mockResolvedValue([{ denom: 'upnyx', amount: '5000000' }]);
    resetStore();
  });

  it('distinguishes an authoritative empty channel set from query failure', async () => {
    mocks.getTransferChannels.mockResolvedValueOnce([]);
    await useIbcTransferStore.getState().loadChannels();
    expect(useIbcTransferStore.getState()).toMatchObject({
      channelStatus: 'ready',
      channels: [],
      channelFailure: null,
    });

    mocks.getTransferChannels.mockRejectedValueOnce(
      new IbcChannelError('decode', 'bad schema')
    );
    await expect(useIbcTransferStore.getState().loadChannels()).rejects.toMatchObject({
      failure: 'channel_query_failed',
    });
    expect(useIbcTransferStore.getState()).toMatchObject({
      channelStatus: 'error',
      channels: [],
      channelFailure: 'decode',
    });
  });

  it('rejects an amount that consumes the reserve before signer access or signing', async () => {
    mocks.getBalance.mockResolvedValueOnce([{ denom: 'upnyx', amount: '1009999' }]);
    await expect(
      useIbcTransferStore.getState().submitTransfer({
        channel: CHANNEL,
        receiver: 'cosmos1receiver',
        amount: '1',
      })
    ).rejects.toMatchObject({ failure: 'insufficient_balance' });
    expect(mocks.getWalletForSigning).not.toHaveBeenCalled();
    expect(mocks.sendIbcTransfer).not.toHaveBeenCalled();
  });

  it('submits exactly once, retains an unknown hash, and persists no secrets', async () => {
    mocks.sendIbcTransfer.mockResolvedValueOnce({
      txHash: HASH.toLowerCase(),
      height: 42,
      phase: 'unknown',
      packet: null,
    });
    const record = await useIbcTransferStore.getState().submitTransfer({
      channel: CHANNEL,
      receiver: 'cosmos1receiver',
      amount: '1.25',
      memo: 'public memo',
    });
    expect(record).toMatchObject({ txHash: HASH, amount: '1250000', phase: 'unknown' });
    expect(mocks.sendIbcTransfer).toHaveBeenCalledOnce();
    expect(mocks.getWalletForSigning).toHaveBeenCalledWith(
      'truerepublic1sender',
      'secret-password'
    );
    expect(mocks.walletState.refreshBalance).toHaveBeenCalledOnce();

    const persisted = localStorage.getItem('ibc-transfer-store') ?? '';
    expect(persisted).toContain(HASH);
    expect(persisted).not.toContain('secret-password');
    expect(persisted).not.toContain('mnemonic');
    expect(persisted).not.toContain('signer');
  });

  it('isolates records by chain and wallet scope', async () => {
    mocks.sendIbcTransfer
      .mockResolvedValueOnce({ txHash: HASH, height: 1, phase: 'unknown', packet: null })
      .mockResolvedValueOnce({ txHash: 'B'.repeat(64), height: 2, phase: 'unknown', packet: null });
    await useIbcTransferStore.getState().submitTransfer({ channel: CHANNEL, receiver: 'cosmos1a', amount: '1' });
    mocks.walletState.currentWallet = { address: 'truerepublic1other', name: 'Two', createdAt: 2 };
    await useIbcTransferStore.getState().submitTransfer({ channel: CHANNEL, receiver: 'cosmos1b', amount: '1' });

    expect(
      useIbcTransferStore.getState().recordsByScope[
        ibcTransferScopeKey('truerepublic-1', 'truerepublic1sender')
      ]
    ).toHaveLength(1);
    expect(
      useIbcTransferStore.getState().recordsByScope[
        ibcTransferScopeKey('truerepublic-1', 'truerepublic1other')
      ]
    ).toHaveLength(1);
  });

  it('invalidates stale completion while retaining a scoped committed recovery record', async () => {
    let resolveSubmission!: (value: unknown) => void;
    mocks.sendIbcTransfer.mockReturnValueOnce(
      new Promise((resolve) => { resolveSubmission = resolve; })
    );
    const pending = useIbcTransferStore.getState().submitTransfer({
      channel: CHANNEL,
      receiver: 'cosmos1receiver',
      amount: '1',
    });
    await vi.waitFor(() => expect(mocks.sendIbcTransfer).toHaveBeenCalledOnce());
    useIbcTransferStore.getState().invalidateSession();
    mocks.walletState.isLocked = true;
    mocks.walletState.currentWallet = null;
    resolveSubmission({ txHash: HASH, height: 8, phase: 'unknown', packet: null });
    await pending;

    expect(useIbcTransferStore.getState()).toMatchObject({
      submissionStatus: 'idle',
      activeTxHash: null,
      sessionGeneration: 1,
    });
    expect(
      useIbcTransferStore.getState().recordsByScope[
        ibcTransferScopeKey('truerepublic-1', 'truerepublic1sender')
      ]
    ).toHaveLength(1);
  });

  it('manually reconciles exact packet evidence without resubmitting', async () => {
    const disconnect = vi.fn();
    mocks.connect.mockResolvedValueOnce({ disconnect, searchTx: vi.fn() });
    mocks.reconcilePacketLifecycle.mockResolvedValueOnce({
      sendPacket: PACKET,
      acknowledgement: 'success',
      timedOut: null,
    });
    const scope = ibcTransferScopeKey('truerepublic-1', 'truerepublic1sender');
    useIbcTransferStore.setState({
      recordsByScope: {
        [scope]: [{
          txHash: HASH,
          chainId: 'truerepublic-1',
          walletAddress: 'truerepublic1sender',
          channel: CHANNEL,
          receiver: 'cosmos1receiver',
          amount: '1000000',
          memo: '',
          height: 8,
          phase: 'committed_pending_relay',
          packet: { ...PACKET, sequence: '9' },
          submittedAt: 1,
        }],
      },
    });

    await expect(useIbcTransferStore.getState().reconcileTransfer(HASH)).resolves.toBe('acknowledged');
    expect(useIbcTransferStore.getState().recordsByScope[scope][0].phase).toBe('acknowledged');
    expect(mocks.reconcilePacketLifecycle).toHaveBeenCalledOnce();
    expect(mocks.sendIbcTransfer).not.toHaveBeenCalled();
    expect(disconnect).toHaveBeenCalledOnce();
  });

  it('retains a record when recovery transport fails and never retries', async () => {
    const disconnect = vi.fn();
    mocks.connect.mockResolvedValueOnce({ disconnect, searchTx: vi.fn() });
    mocks.reconcilePacketLifecycle.mockRejectedValueOnce(new Error('offline'));
    const scope = ibcTransferScopeKey('truerepublic-1', 'truerepublic1sender');
    const record = {
      txHash: HASH, chainId: 'truerepublic-1', walletAddress: 'truerepublic1sender',
      channel: CHANNEL, receiver: 'cosmos1receiver', amount: '1', memo: '', height: 1,
      phase: 'committed_pending_relay' as const,
      packet: { ...PACKET, sequence: '9' }, submittedAt: 1,
    };
    useIbcTransferStore.setState({ recordsByScope: { [scope]: [record] } });

    await expect(useIbcTransferStore.getState().reconcileTransfer(HASH)).rejects.toThrow('offline');
    expect(useIbcTransferStore.getState().recordsByScope[scope]).toEqual([record]);
    expect(mocks.reconcilePacketLifecycle).toHaveBeenCalledOnce();
    expect(mocks.sendIbcTransfer).not.toHaveBeenCalled();
    expect(disconnect).toHaveBeenCalledOnce();
  });

  it('rejects recovery without packet evidence before connecting', async () => {
    const scope = ibcTransferScopeKey('truerepublic-1', 'truerepublic1sender');
    useIbcTransferStore.setState({ recordsByScope: { [scope]: [{
      txHash: HASH, chainId: 'truerepublic-1', walletAddress: 'truerepublic1sender',
      channel: CHANNEL, receiver: 'cosmos1receiver', amount: '1', memo: '', height: 1,
      phase: 'unknown', packet: null, submittedAt: 1,
    }] } });
    await expect(useIbcTransferStore.getState().reconcileTransfer(HASH)).rejects.toBeInstanceOf(IbcTransferStoreError);
    expect(mocks.connect).not.toHaveBeenCalled();
  });
});
