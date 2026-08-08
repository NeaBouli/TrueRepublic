export interface SendParams {
  to: string;
  amount: string;
  denom: string;
  memo?: string;
}

export interface TransactionResult {
  hash: string;
  height: number;
  success: boolean;
  error?: string;
}

/**
 * Typed failures for the submitted-transaction history query (GH-131). Every
 * failure stays distinct from an authoritative empty page: an empty history is
 * only reported when the chain itself answers with a valid zero-total page.
 */
export type TransactionHistoryFailure =
  | 'unavailable'
  | 'timeout'
  | 'protocol'
  | 'decode';

/**
 * One message of an indexed transaction. The type URL is preserved verbatim;
 * unknown message types are valid data and must be rendered safely by the UI.
 */
export interface SubmittedTransactionMessage {
  typeUrl: string;
}

export interface SubmittedTransactionFee {
  denom: string;
  amount: string;
}

/**
 * One indexed transaction submitted by the wallet. Fields the chain response
 * does not carry stay `null` instead of being fabricated.
 */
export interface SubmittedTransaction {
  hash: string;
  height: number;
  /** Block time reported by the indexer; null when absent. */
  timestamp: string | null;
  /** Delivered transaction code; 0 means success. */
  code: number;
  status: 'success' | 'failed';
  /** Bounded plain-text chain log for committed failures; null on success. */
  error: string | null;
  memo: string | null;
  messages: SubmittedTransactionMessage[];
  /** Declared fee coins; null when the response carries no fee. */
  fee: SubmittedTransactionFee[] | null;
}

/**
 * One newest-first server-paginated page of transactions submitted by
 * `address`. Incoming-only account activity is never part of this result.
 */
export interface SubmittedTransactionsPage {
  address: string;
  page: number;
  pageSize: number;
  total: number;
  hasMore: boolean;
  transactions: SubmittedTransaction[];
}
