import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import { toBech32 } from '@cosmjs/encoding';
import type { SubmittedTransaction } from '@/types/transaction';
import type { Wallet } from '@/types/wallet';
import { useWalletStore } from '@/stores/walletStore';
import { TransactionHistory } from './TransactionHistory';

const ADDRESS = toBech32(
  'truerepublic',
  Uint8Array.from({ length: 20 }, (_, index) => index + 1)
);

const wallet: Wallet = { address: ADDRESS, name: 'Primary', createdAt: 1 };

const originalActions = {
  loadHistoryPage: useWalletStore.getState().loadHistoryPage,
  nextHistoryPage: useWalletStore.getState().nextHistoryPage,
  prevHistoryPage: useWalletStore.getState().prevHistoryPage,
};

function tx(overrides: Partial<SubmittedTransaction> = {}): SubmittedTransaction {
  return {
    hash: 'AB'.repeat(32),
    height: 42,
    timestamp: '2026-08-08T12:34:56Z',
    code: 0,
    status: 'success',
    error: null,
    memo: null,
    messages: [{ typeUrl: '/cosmos.bank.v1beta1.MsgSend' }],
    fee: [{ denom: 'upnyx', amount: '12500' }],
    ...overrides,
  };
}

type StorePatch = Exclude<
  Parameters<typeof useWalletStore.setState>[0],
  // eslint-disable-next-line @typescript-eslint/no-unsafe-function-type
  Function
>;

function setHistoryState(overrides: StorePatch = {}) {
  useWalletStore.setState({
    currentWallet: { ...wallet },
    isLocked: false,
    historyTransactions: [],
    historyStatus: 'idle',
    historyFailure: null,
    historyError: null,
    historyPage: 1,
    historyPageSize: 20,
    historyTotal: 0,
    historyHasMore: false,
    historyAddress: ADDRESS,
    ...originalActions,
    ...overrides,
  });
}

describe('TransactionHistory', () => {
  beforeEach(() => {
    setHistoryState();
  });

  it('is explicitly titled and scopes itself to submitted transactions only', () => {
    render(<TransactionHistory />);

    expect(
      screen.getByRole('heading', { name: 'Submitted transactions' })
    ).toBeInTheDocument();
    expect(
      screen.getByText(/Incoming transfers to this wallet are not included/i)
    ).toBeInTheDocument();
  });

  it('loads page 1 on mount for an address without history', () => {
    const loadHistoryPage = vi.fn(async () => {});
    setHistoryState({ historyAddress: null, loadHistoryPage });

    render(<TransactionHistory />);

    expect(loadHistoryPage).toHaveBeenCalledWith(1);
  });

  it('does not reload when the address history is already loaded', () => {
    const loadHistoryPage = vi.fn(async () => {});
    setHistoryState({ historyAddress: ADDRESS, historyStatus: 'ready', loadHistoryPage });

    render(<TransactionHistory />);

    expect(loadHistoryPage).not.toHaveBeenCalled();
  });

  it('announces loading while the first page is in flight', () => {
    setHistoryState({ historyStatus: 'loading' });

    render(<TransactionHistory />);

    expect(screen.getByRole('status')).toHaveTextContent(
      'Loading submitted transactions'
    );
  });

  it('shows a typed error with retry instead of an empty list', () => {
    const loadHistoryPage = vi.fn(async () => {});
    setHistoryState({
      historyStatus: 'error',
      historyFailure: 'timeout',
      historyTransactions: [tx()],
      loadHistoryPage,
    });

    render(<TransactionHistory />);

    const alert = screen.getByRole('alert');
    expect(alert).toHaveTextContent('timed out');
    expect(
      screen.queryByText(/No submitted transactions found/i)
    ).not.toBeInTheDocument();
    expect(screen.queryByTestId('history-row')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Retry' }));
    expect(loadHistoryPage).toHaveBeenCalledWith(1);
  });

  it('renders the authoritative empty state without an error', () => {
    setHistoryState({ historyStatus: 'ready' });

    render(<TransactionHistory />);

    expect(
      screen.getByText('No submitted transactions found for this wallet.')
    ).toBeInTheDocument();
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('renders successful and failed rows with real fields', () => {
    setHistoryState({
      historyStatus: 'ready',
      historyTotal: 2,
      historyTransactions: [
        tx({ memo: 'rent' }),
        tx({
          hash: 'CD'.repeat(32),
          height: 41,
          code: 5,
          status: 'failed',
          error: 'failed to execute message: slippage exceeded',
          messages: [{ typeUrl: '/dex.MsgSwapExact' }],
        }),
      ],
    });

    render(<TransactionHistory />);

    const rows = screen.getAllByTestId('history-row');
    expect(rows).toHaveLength(2);
    expect(rows[0]).toHaveTextContent('Send');
    expect(rows[0]).toHaveTextContent('Success');
    expect(rows[0]).toHaveTextContent('Memo: rent');
    expect(rows[0]).toHaveTextContent('0.0125 PNYX');
    expect(rows[1]).toHaveTextContent('Swap');
    expect(rows[1]).toHaveTextContent('Failed (code 5)');
    expect(rows[1]).toHaveTextContent('slippage exceeded');
    // The full hash stays available without dumping 64 characters into the row.
    expect(
      screen.getByLabelText(`Transaction hash ${'AB'.repeat(32)}`)
    ).toBeInTheDocument();
  });

  it('renders unknown message types and hostile text safely', () => {
    const hostile = '<img src=x onerror=alert(1)>';
    setHistoryState({
      historyStatus: 'ready',
      historyTotal: 1,
      historyTransactions: [
        tx({
          messages: [{ typeUrl: '/attacker.MsgAnything' }],
          memo: hostile,
        }),
      ],
    });

    const { container } = render(<TransactionHistory />);

    expect(screen.getByText(/Unknown message/)).toBeInTheDocument();
    expect(screen.getByText(`Memo: ${hostile}`)).toBeInTheDocument();
    expect(container.querySelector('img')).toBeNull();
  });

  it('does not fabricate missing time or fee', () => {
    setHistoryState({
      historyStatus: 'ready',
      historyTotal: 1,
      historyTransactions: [tx({ timestamp: null, fee: null })],
    });

    render(<TransactionHistory />);

    expect(screen.getByText('Time unavailable')).toBeInTheDocument();
    expect(screen.getByText('Fee unavailable')).toBeInTheDocument();
  });

  it('paginates within bounds through the store actions', () => {
    const nextHistoryPage = vi.fn(async () => {});
    const prevHistoryPage = vi.fn(async () => {});
    setHistoryState({
      historyStatus: 'ready',
      historyPage: 2,
      historyTotal: 45,
      historyHasMore: true,
      historyTransactions: [tx()],
      nextHistoryPage,
      prevHistoryPage,
    });

    render(<TransactionHistory />);

    expect(screen.getByText(/Page 2 of 3/)).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Next page' }));
    expect(nextHistoryPage).toHaveBeenCalled();
    fireEvent.click(screen.getByRole('button', { name: 'Previous page' }));
    expect(prevHistoryPage).toHaveBeenCalled();
  });

  it('disables pagination at the bounds', () => {
    setHistoryState({
      historyStatus: 'ready',
      historyPage: 1,
      historyTotal: 10,
      historyHasMore: false,
      historyTransactions: [tx()],
    });

    render(<TransactionHistory />);

    expect(screen.getByRole('button', { name: 'Previous page' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'Next page' })).toBeDisabled();
  });
});
