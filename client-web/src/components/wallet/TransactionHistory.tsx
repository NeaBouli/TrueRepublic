import { useEffect } from 'react';
import { useWalletStore } from '@/stores/walletStore';
import type {
  SubmittedTransaction,
  TransactionHistoryFailure,
} from '@/types/transaction';
import { formatPnyx } from '@/utils/format';
import { DEFAULT_CHAIN } from '@/config/chains';

/**
 * Human labels for the message types this client can submit. Any other type
 * URL is valid chain data and renders through the safe "Unknown message" path
 * below — it is never executed, interpreted, or hidden.
 */
const MESSAGE_LABELS: Record<string, string> = {
  '/cosmos.bank.v1beta1.MsgSend': 'Send',
  '/truedemocracy.MsgCreateDomain': 'Create domain',
  '/truedemocracy.MsgAddMember': 'Add member',
  '/truedemocracy.MsgSubmitProposal': 'Submit proposal',
  '/truedemocracy.MsgPlaceStoneOnIssue': 'Place stone on issue',
  '/truedemocracy.MsgPlaceStoneOnSuggestion': 'Place stone on suggestion',
  '/truedemocracy.MsgOnboardToDomain': 'Onboard to domain',
  '/truedemocracy.MsgApproveOnboarding': 'Approve onboarding',
  '/truedemocracy.MsgRegisterIdentity': 'Register identity',
  '/dex.MsgAddLiquidity': 'Add liquidity',
  '/dex.MsgRemoveLiquidity': 'Remove liquidity',
  '/dex.MsgSwapExact': 'Swap',
};

const FAILURE_MESSAGES: Record<TransactionHistoryFailure, string> = {
  unavailable: 'The transaction service is unavailable right now.',
  timeout: 'The transaction history request timed out.',
  protocol: 'The chain returned an unexpected transaction history response.',
  decode: 'The chain transaction history response could not be decoded.',
};

function messageLabel(typeUrl: string): string {
  return MESSAGE_LABELS[typeUrl] ?? 'Unknown message';
}

function formatTimestamp(timestamp: string | null): string {
  if (timestamp === null) return 'Time unavailable';
  const parsed = new Date(timestamp);
  if (Number.isNaN(parsed.getTime())) return timestamp;
  return parsed.toLocaleString();
}

function formatFee(fee: SubmittedTransaction['fee']): string {
  if (fee === null) return 'Fee unavailable';
  if (fee.length === 0) return 'No fee';
  return fee
    .map((coin) =>
      coin.denom === DEFAULT_CHAIN.coinMinimalDenom
        ? `${formatPnyx(coin.amount)} ${DEFAULT_CHAIN.coinDenom}`
        : `${coin.amount} ${coin.denom}`
    )
    .join(', ');
}

function summarizeMessages(tx: SubmittedTransaction): string {
  if (tx.messages.length === 0) return 'No message details';
  const labels = tx.messages.map((message) => messageLabel(message.typeUrl));
  const unique = [...new Set(labels)];
  const summary = unique.join(', ');
  return tx.messages.length > 1 ? `${summary} (${tx.messages.length} messages)` : summary;
}

function HistoryRow({ tx }: { tx: SubmittedTransaction }) {
  const shortHash = `${tx.hash.slice(0, 10)}…${tx.hash.slice(-6)}`;
  return (
    <li className="py-3" data-testid="history-row">
      <div className="flex items-center justify-between gap-3">
        <div className="min-w-0">
          <p className="text-sm font-medium text-gray-900 truncate">
            {summarizeMessages(tx)}
          </p>
          <p className="text-xs text-gray-500 mt-0.5">
            <span
              className="font-mono"
              title={tx.hash}
              aria-label={`Transaction hash ${tx.hash}`}
            >
              {shortHash}
            </span>
            {' · '}
            <span>{formatTimestamp(tx.timestamp)}</span>
            {' · '}
            <span>{formatFee(tx.fee)}</span>
          </p>
        </div>
        {tx.status === 'success' ? (
          <span className="shrink-0 inline-flex items-center rounded-full bg-green-50 px-2 py-1 text-xs font-medium text-green-700">
            Success
          </span>
        ) : (
          <span className="shrink-0 inline-flex items-center rounded-full bg-red-50 px-2 py-1 text-xs font-medium text-red-700">
            Failed (code {tx.code})
          </span>
        )}
      </div>
      {tx.memo !== null && tx.memo !== '' && (
        <p className="text-xs text-gray-600 mt-1 break-words">
          Memo: {tx.memo}
        </p>
      )}
      {tx.error !== null && (
        <p className="text-xs text-red-700 bg-red-50 rounded p-2 mt-1 break-words">
          {tx.error}
        </p>
      )}
    </li>
  );
}

/**
 * Newest-first, server-paginated history of transactions submitted by the
 * unlocked wallet. This list is honestly scoped to submissions: incoming-only
 * account activity is not part of the query and is never implied here.
 */
export function TransactionHistory() {
  const address = useWalletStore((state) => state.currentWallet?.address ?? null);
  const status = useWalletStore((state) => state.historyStatus);
  const failure = useWalletStore((state) => state.historyFailure);
  const transactions = useWalletStore((state) => state.historyTransactions);
  const page = useWalletStore((state) => state.historyPage);
  const pageSize = useWalletStore((state) => state.historyPageSize);
  const total = useWalletStore((state) => state.historyTotal);
  const hasMore = useWalletStore((state) => state.historyHasMore);
  const loadedAddress = useWalletStore((state) => state.historyAddress);
  const nextHistoryPage = useWalletStore((state) => state.nextHistoryPage);
  const prevHistoryPage = useWalletStore((state) => state.prevHistoryPage);

  useEffect(() => {
    if (!address) return;
    const state = useWalletStore.getState();
    if (state.historyAddress !== address && state.historyStatus !== 'loading') {
      void state.loadHistoryPage(1);
    }
  }, [address]);

  const pageCount = Math.max(1, Math.ceil(total / pageSize));

  return (
    <section aria-labelledby="submitted-transactions-heading" className="bg-white rounded-xl border border-gray-200 p-6">
      <h2
        id="submitted-transactions-heading"
        className="text-lg font-semibold text-gray-900"
      >
        Submitted transactions
      </h2>
      <p className="text-xs text-gray-500 mt-1">
        Only transactions submitted by this wallet are shown. Incoming
        transfers to this wallet are not included.
      </p>

      <div className="mt-4" aria-busy={status === 'loading'}>
        {status === 'loading' && transactions.length === 0 && (
          <p role="status" className="text-sm text-gray-600 py-4 text-center">
            Loading submitted transactions…
          </p>
        )}

        {status === 'error' && (
          <div role="alert" className="py-4 text-center">
            <p className="text-sm text-red-700">
              {failure !== null
                ? FAILURE_MESSAGES[failure]
                : 'Submitted transactions could not be loaded.'}
            </p>
            <button
              type="button"
              onClick={() => void useWalletStore.getState().loadHistoryPage(page)}
              className="mt-3 inline-flex items-center rounded-lg border border-gray-300 px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-50"
            >
              Retry
            </button>
          </div>
        )}

        {status === 'ready' && transactions.length === 0 && (
          <p className="text-sm text-gray-600 py-4 text-center">
            No submitted transactions found for this wallet.
          </p>
        )}

        {status === 'ready' && transactions.length > 0 && (
          <ul
            aria-label="Submitted transactions"
            aria-live="polite"
            className="divide-y divide-gray-100"
          >
            {transactions.map((tx) => (
              <HistoryRow key={tx.hash} tx={tx} />
            ))}
          </ul>
        )}
      </div>

      {loadedAddress !== null && status !== 'idle' && total > 0 && (
        <nav
          aria-label="Submitted transactions pages"
          className="mt-4 flex items-center justify-between"
        >
          <button
            type="button"
            aria-label="Previous page"
            onClick={() => void prevHistoryPage()}
            disabled={page <= 1 || status === 'loading'}
            className="inline-flex items-center rounded-lg border border-gray-300 px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Previous
          </button>
          <span className="text-xs text-gray-600">
            Page {page} of {pageCount} · {total} transaction{total === 1 ? '' : 's'}
          </span>
          <button
            type="button"
            aria-label="Next page"
            onClick={() => void nextHistoryPage()}
            disabled={!hasMore || status === 'loading'}
            className="inline-flex items-center rounded-lg border border-gray-300 px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Next
          </button>
        </nav>
      )}
    </section>
  );
}
