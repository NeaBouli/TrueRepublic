import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';

const mocks = vi.hoisted(() => ({
  wallet: {
    currentWallet: { address: 'truerepublic1sender', name: 'One', createdAt: 1 },
    isLocked: false,
    balances: [{ denom: 'upnyx', amount: '5000000' }],
  },
  ibc: {
    channels: [] as Array<Record<string, string>>,
    channelStatus: 'ready',
    channelFailure: null as string | null,
    channelError: null as string | null,
    submissionStatus: 'idle',
    submissionPhase: null,
    submissionError: null as string | null,
    activeTxHash: null as string | null,
    reconcileStatus: 'idle',
    reconcileError: null as string | null,
    reconcileTxHash: null as string | null,
    recordsByScope: {} as Record<string, unknown[]>,
    sessionGeneration: 0,
    loadChannels: vi.fn().mockResolvedValue(undefined),
    submitTransfer: vi.fn(),
    reconcileTransfer: vi.fn(),
    invalidateSession: vi.fn(),
  },
}));

vi.mock('@/stores/walletStore', () => ({ useWalletStore: () => mocks.wallet }));
vi.mock('@/stores/ibcTransferStore', async (importActual) => {
  const actual = await importActual<typeof import('@/stores/ibcTransferStore')>();
  return { ...actual, useIbcTransferStore: () => mocks.ibc };
});

import { IbcTransferPage } from './IbcTransferPage';

const CHANNEL = {
  portId: 'transfer', channelId: 'channel-0', counterpartyPortId: 'transfer',
  counterpartyChannelId: 'channel-7', connectionId: 'connection-0', version: 'ics20-1',
};
const HASH = 'A'.repeat(64);

function renderPage(path = '/ibc/transfer') {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route path="/ibc/transfer" element={<IbcTransferPage />} />
        <Route path="/ibc/transfer/:txHash" element={<IbcTransferPage />} />
        <Route path="/wallet" element={<div>Wallet</div>} />
        <Route path="/unlock" element={<div>Unlock</div>} />
      </Routes>
    </MemoryRouter>
  );
}

describe('IbcTransferPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    Object.assign(mocks.ibc, {
      channels: [], channelStatus: 'ready', channelFailure: null, channelError: null,
      submissionStatus: 'idle', submissionPhase: null, submissionError: null,
      reconcileStatus: 'idle', reconcileError: null, recordsByScope: {},
    });
    mocks.ibc.loadChannels.mockResolvedValue(undefined);
  });

  it('renders query failure separately from an empty valid channel set', () => {
    Object.assign(mocks.ibc, { channelStatus: 'error', channelFailure: 'decode' });
    const { rerender } = renderPage();
    expect(screen.getByRole('alert')).toHaveTextContent('not an empty channel set');

    Object.assign(mocks.ibc, { channelStatus: 'ready', channelFailure: null });
    rerender(
      <MemoryRouter initialEntries={['/ibc/transfer']}>
        <Routes><Route path="/ibc/transfer" element={<IbcTransferPage />} /></Routes>
      </MemoryRouter>
    );
    expect(screen.getByText(/no selectable open ICS-20 transfer channel/i)).toBeInTheDocument();
  });

  it('requires explicit counterparty receiver verification before submission', async () => {
    const user = userEvent.setup();
    Object.assign(mocks.ibc, { channels: [CHANNEL], channelStatus: 'ready' });
    mocks.ibc.submitTransfer.mockResolvedValue({ txHash: HASH });
    renderPage();
    await user.type(screen.getByLabelText(/Counterparty receiver/i), 'cosmos1receiver');
    await user.type(screen.getByLabelText(/Amount/i), '1');
    await user.click(screen.getByRole('button', { name: /Validate & sign once/i }));
    expect(screen.getByRole('alert')).toHaveTextContent(/Confirm that you verified/i);
    expect(mocks.ibc.submitTransfer).not.toHaveBeenCalled();

    await user.click(screen.getByRole('checkbox'));
    await user.click(screen.getByRole('button', { name: /Validate & sign once/i }));
    expect(mocks.ibc.submitTransfer).toHaveBeenCalledOnce();
  });

  it('labels a committed packet as pending relay, never as delivered', () => {
    const scope = 'truerepublic-1:truerepublic1sender';
    mocks.ibc.recordsByScope = {
      [scope]: [{
        txHash: HASH, chainId: 'truerepublic-1', walletAddress: 'truerepublic1sender',
        channel: CHANNEL, receiver: 'cosmos1receiver', amount: '1000000', memo: '',
        height: 10, phase: 'committed_pending_relay',
        packet: { sequence: '9', sourcePort: 'transfer', sourceChannel: 'channel-0', destinationPort: 'transfer', destinationChannel: 'channel-7' },
        submittedAt: 1,
      }],
    };
    renderPage(`/ibc/transfer/${HASH}`);
    expect(screen.getAllByText(/Committed · pending relay/i).length).toBeGreaterThan(0);
    expect(screen.getByText(/Delivery is not yet proven/i)).toBeInTheDocument();
    expect(screen.queryByText(/^Delivered$/i)).not.toBeInTheDocument();
  });
});
