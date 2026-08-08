import { fromBech32 } from '@cosmjs/encoding';
import type { DirectSecp256k1HdWallet } from '@cosmjs/proto-signing';
import type { SigningStargateClient } from '@cosmjs/stargate';
import type { ChainConfig } from '@/types/chain';
import type {
  SendParams,
  SubmittedTransaction,
  SubmittedTransactionFee,
  SubmittedTransactionMessage,
  SubmittedTransactionsPage,
  TransactionHistoryFailure,
  TransactionResult,
} from '@/types/transaction';
import { connectSigningClient, deliverMessages } from './signingClient';

export const HISTORY_DEFAULT_PAGE_SIZE = 20;
export const HISTORY_MAX_PAGE_SIZE = 50;
const HISTORY_TIMEOUT_MS = 10_000;
const HISTORY_ERROR_MAX_LENGTH = 200;

/**
 * Typed history failure. Every variant stays distinct from an authoritative
 * empty page: callers only see an empty result when the chain itself returns
 * a valid zero-total page.
 */
export class TransactionHistoryError extends Error {
  public readonly originalCause?: unknown;

  constructor(
    public readonly failure: TransactionHistoryFailure,
    message: string,
    cause?: unknown
  ) {
    super(`Transaction history query failed: ${message}`);
    this.name = 'TransactionHistoryError';
    this.originalCause = cause;
  }
}

type Fetch = (
  input: RequestInfo | URL,
  init?: RequestInit
) => Promise<Response>;

function isTimeoutError(error: unknown, signal: AbortSignal): boolean {
  if (!signal.aborted) return false;
  if (error === signal.reason) return true;
  return (
    error instanceof DOMException &&
    (error.name === 'TimeoutError' || error.name === 'AbortError')
  );
}

function expectRecord(value: unknown, detail: string): Record<string, unknown> {
  if (typeof value !== 'object' || value === null || Array.isArray(value)) {
    throw new TransactionHistoryError('decode', `${detail} must be an object`);
  }
  return value as Record<string, unknown>;
}

function expectDigits(value: unknown, detail: string): string {
  if (typeof value !== 'string' || !/^\d+$/.test(value)) {
    throw new TransactionHistoryError(
      'decode',
      `${detail} must be an unsigned decimal string`
    );
  }
  return value;
}

function expectOptionalString(
  value: unknown,
  detail: string
): string | null {
  if (value === undefined || value === null) return null;
  if (typeof value !== 'string') {
    throw new TransactionHistoryError('decode', `${detail} must be a string`);
  }
  return value;
}

function parseCoin(value: unknown, detail: string): SubmittedTransactionFee {
  const coin = expectRecord(value, detail);
  return {
    denom: (() => {
      const denom = coin.denom;
      if (typeof denom !== 'string' || denom.length === 0) {
        throw new TransactionHistoryError(
          'decode',
          `${detail}.denom must be a non-empty string`
        );
      }
      return denom;
    })(),
    amount: expectDigits(coin.amount, `${detail}.amount`),
  };
}

/**
 * Bound an untrusted chain log to plain single-line text so a committed
 * failure can be shown without reflecting control sequences or unbounded
 * payloads into the UI.
 */
export function boundChainError(log: string, code: number): string {
  const plain = log
    // eslint-disable-next-line no-control-regex -- control characters are exactly what must be stripped from untrusted chain logs
    .replace(/[\x00-\x1f\x7f]/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
    .slice(0, HISTORY_ERROR_MAX_LENGTH)
    .trim();
  return plain || `Transaction failed with code ${code}`;
}

function parseMessage(value: unknown, detail: string): SubmittedTransactionMessage {
  const message = expectRecord(value, detail);
  const typeUrl = message['@type'];
  if (typeof typeUrl !== 'string' || typeUrl.length === 0) {
    throw new TransactionHistoryError(
      'decode',
      `${detail}.@type must be a non-empty string`
    );
  }
  return { typeUrl };
}

function parseSubmittedTransaction(value: unknown): SubmittedTransaction {
  const entry = expectRecord(value, 'tx_responses entry');

  const hash = entry.txhash;
  if (typeof hash !== 'string' || !/^[0-9A-Fa-f]{64}$/.test(hash)) {
    throw new TransactionHistoryError(
      'decode',
      'tx_responses entry txhash must be 64 hex characters'
    );
  }

  const heightText = expectDigits(entry.height, 'tx_responses entry height');
  const height = Number(heightText);
  if (!Number.isSafeInteger(height)) {
    throw new TransactionHistoryError(
      'decode',
      'tx_responses entry height exceeds the safe integer range'
    );
  }

  const codeValue = entry.code;
  if (codeValue !== undefined && codeValue !== null) {
    if (
      typeof codeValue !== 'number' ||
      !Number.isInteger(codeValue) ||
      codeValue < 0
    ) {
      throw new TransactionHistoryError(
        'decode',
        'tx_responses entry code must be a non-negative integer'
      );
    }
  }
  const code = typeof codeValue === 'number' ? codeValue : 0;

  const timestamp = expectOptionalString(
    entry.timestamp,
    'tx_responses entry timestamp'
  );
  const rawLog = expectOptionalString(
    entry.raw_log,
    'tx_responses entry raw_log'
  );

  const tx = expectRecord(entry.tx, 'tx_responses entry tx');
  const body = expectRecord(tx.body, 'tx_responses entry tx.body');

  let messages: SubmittedTransactionMessage[] = [];
  if (body.messages !== undefined && body.messages !== null) {
    if (!Array.isArray(body.messages)) {
      throw new TransactionHistoryError(
        'decode',
        'tx_responses entry tx.body.messages must be an array'
      );
    }
    messages = body.messages.map((message, index) =>
      parseMessage(message, `tx_responses entry message[${index}]`)
    );
  }

  const memo = expectOptionalString(body.memo, 'tx_responses entry tx.body.memo');

  let fee: SubmittedTransactionFee[] | null = null;
  if (tx.auth_info !== undefined && tx.auth_info !== null) {
    const authInfo = expectRecord(tx.auth_info, 'tx_responses entry tx.auth_info');
    if (authInfo.fee !== undefined && authInfo.fee !== null) {
      const feeRecord = expectRecord(
        authInfo.fee,
        'tx_responses entry tx.auth_info.fee'
      );
      if (feeRecord.amount !== undefined && feeRecord.amount !== null) {
        if (!Array.isArray(feeRecord.amount)) {
          throw new TransactionHistoryError(
            'decode',
            'tx_responses entry fee amount must be an array'
          );
        }
        fee = feeRecord.amount.map((coin, index) =>
          parseCoin(coin, `tx_responses entry fee amount[${index}]`)
        );
      }
    }
  }

  const status = code === 0 ? 'success' : 'failed';
  return {
    hash,
    height,
    timestamp,
    code,
    status,
    error: status === 'failed' ? boundChainError(rawLog ?? '', code) : null,
    memo,
    messages,
    fee,
  };
}

export class TransactionService {
  private readonly fetchImpl: Fetch;

  constructor(
    private readonly config: ChainConfig,
    fetchImpl: Fetch = (input, init) => globalThis.fetch(input, init),
    private readonly historyTimeoutMs: number = HISTORY_TIMEOUT_MS
  ) {
    if (
      !Number.isSafeInteger(historyTimeoutMs) ||
      historyTimeoutMs <= 0
    ) {
      throw new Error('history timeout must be a positive safe integer');
    }
    this.fetchImpl = fetchImpl;
  }

  /**
   * Send tokens via the standard bank message retained through the canonical
   * registry's default types.
   */
  async send(
    wallet: DirectSecp256k1HdWallet,
    params: SendParams
  ): Promise<TransactionResult> {
    const [account] = await wallet.getAccounts();

    let client: SigningStargateClient | undefined;

    try {
      client = await connectSigningClient(this.config, wallet);
      const msg = {
        typeUrl: '/cosmos.bank.v1beta1.MsgSend',
        value: {
          fromAddress: account.address,
          toAddress: params.to,
          amount: [
            {
              denom: params.denom,
              amount: params.amount,
            },
          ],
        },
      };

      return await deliverMessages(
        client,
        account.address,
        [msg],
        this.config.gasPrice,
        params.memo || ''
      );
    } catch (error: unknown) {
      const message =
        error instanceof Error ? error.message : 'Transaction failed';
      return {
        hash: '',
        height: 0,
        success: false,
        error: message,
      };
    } finally {
      client?.disconnect();
    }
  }

  /**
   * Newest-first server-paginated history of transactions submitted by
   * `address`, read from the configured Cosmos REST transaction query
   * (`/cosmos/tx/v1beta1/txs`) with the CometBFT query
   * `tx.acc_seq CONTAINS '<address>/'`. Only transactions signed by this
   * address are returned; incoming-only account activity is excluded by
   * construction.
   *
   * The response is schema-validated and fails closed with a typed
   * unavailable/timeout/protocol/decode error. An empty page is only returned
   * for a valid zero-result chain answer.
   */
  async getSubmittedTransactions(
    address: string,
    page = 1,
    pageSize = HISTORY_DEFAULT_PAGE_SIZE,
    callerSignal?: AbortSignal
  ): Promise<SubmittedTransactionsPage> {
    let normalizedAddress: string;
    try {
      const decoded = fromBech32(address);
      if (
        decoded.prefix !== this.config.bech32Prefix ||
        decoded.data.length !== 20
      ) {
        throw new Error('wrong address network or length');
      }
      normalizedAddress = address;
    } catch (error) {
      throw new TransactionHistoryError(
        'protocol',
        `address must be a valid ${this.config.bech32Prefix} bech32 address with 20 bytes`,
        error
      );
    }

    const requestedPage = Number.isSafeInteger(page) && page >= 1 ? page : 1;
    const requestedPageSize =
      Number.isSafeInteger(pageSize) && pageSize >= 1
        ? Math.min(pageSize, HISTORY_MAX_PAGE_SIZE)
        : HISTORY_DEFAULT_PAGE_SIZE;

    // Cosmos SDK v0.50 GetTxsEvent: the deprecated events/pagination request
    // fields are ignored server-side; the raw CometBFT query plus 1-based
    // page/limit and ORDER_BY_DESC give newest-first server pagination.
    const query = new URLSearchParams();
    query.set('query', `tx.acc_seq CONTAINS '${normalizedAddress}/'`);
    query.set('page', String(requestedPage));
    query.set('limit', String(requestedPageSize));
    query.set('order_by', 'ORDER_BY_DESC');
    const url = `${this.config.rest}/cosmos/tx/v1beta1/txs?${query.toString()}`;

    const timeoutSignal = AbortSignal.timeout(this.historyTimeoutMs);
    const signal = callerSignal
      ? AbortSignal.any([timeoutSignal, callerSignal])
      : timeoutSignal;

    let response: Response;
    try {
      response = await this.fetchImpl(url, {
        signal,
        headers: { accept: 'application/json' },
      });
    } catch (error) {
      if (callerSignal?.aborted && !timeoutSignal.aborted) {
        throw error;
      }
      if (isTimeoutError(error, timeoutSignal)) {
        throw new TransactionHistoryError(
          'timeout',
          `request timed out after ${this.historyTimeoutMs} ms`,
          error
        );
      }
      throw new TransactionHistoryError(
        'unavailable',
        'transaction query request failed',
        error
      );
    }
    if (!response.ok) {
      throw new TransactionHistoryError(
        'unavailable',
        `transaction query returned HTTP ${response.status}`
      );
    }

    let body: unknown;
    try {
      body = await response.json();
    } catch (error) {
      if (callerSignal?.aborted && !timeoutSignal.aborted) {
        throw error;
      }
      if (isTimeoutError(error, timeoutSignal)) {
        throw new TransactionHistoryError(
          'timeout',
          `request timed out after ${this.historyTimeoutMs} ms`,
          error
        );
      }
      throw new TransactionHistoryError(
        'decode',
        'transaction query response is not JSON',
        error
      );
    }

    if (typeof body !== 'object' || body === null || Array.isArray(body)) {
      throw new TransactionHistoryError(
        'protocol',
        'transaction query response is not a tx service envelope'
      );
    }
    const envelope = body as Record<string, unknown>;

    if (envelope.total === undefined || envelope.total === null) {
      throw new TransactionHistoryError(
        'protocol',
        'transaction query response is missing total'
      );
    }
    const totalText = expectDigits(envelope.total, 'total');
    const total = Number(totalText);
    if (!Number.isSafeInteger(total)) {
      throw new TransactionHistoryError(
        'decode',
        'total exceeds the safe integer range'
      );
    }

    let txResponses: unknown[] = [];
    if (envelope.tx_responses !== undefined && envelope.tx_responses !== null) {
      if (!Array.isArray(envelope.tx_responses)) {
        throw new TransactionHistoryError(
          'decode',
          'tx_responses must be an array'
        );
      }
      txResponses = envelope.tx_responses;
    }

    const transactions = txResponses.map(parseSubmittedTransaction);
    if (transactions.length > requestedPageSize || transactions.length > total) {
      throw new TransactionHistoryError(
        'protocol',
        'transaction query response count exceeds its requested page or total'
      );
    }

    return {
      address: normalizedAddress,
      page: requestedPage,
      pageSize: requestedPageSize,
      total,
      hasMore: requestedPage * requestedPageSize < total,
      transactions,
    };
  }
}
