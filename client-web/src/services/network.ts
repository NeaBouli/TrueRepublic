import type { ChainConfig } from '@/types/chain';
import type { TransferChannel } from '@/types/ibc';
import type {
  NetworkInfo,
  Validator,
  Block,
  IBCChannel,
} from '@/types/network';
import {
  expectQueryArray,
  expectQueryBoolean,
  expectQueryNumber,
  expectQueryRecord,
  expectQueryString,
  expectQueryStringArray,
  ModuleQueryClient,
  ModuleQueryError,
  QUERY_PATHS,
} from './moduleQuery';

function expectValidator(value: unknown): Validator {
  const path = QUERY_PATHS.truedemocracy.validators;
  const validator = expectQueryRecord(path, value);
  expectQueryString(path, 'operator_addr', validator.operator_addr);
  expectQueryStringArray(path, 'domains', validator.domains);
  expectQueryNumber(path, 'power', validator.power);
  expectQueryBoolean(path, 'jailed', validator.jailed);
  expectQueryNumber(path, 'jailed_until', validator.jailed_until);
  expectQueryNumber(path, 'missed_blocks', validator.missed_blocks);
  for (const [index, value] of expectQueryArray<unknown>(path, validator.stake).entries()) {
    const coin = expectQueryRecord(path, value);
    expectQueryString(path, `stake[${index}].denom`, coin.denom);
    expectQueryString(path, `stake[${index}].amount`, coin.amount);
  }
  return validator as unknown as Validator;
}

/** Failure kinds for the IBC channel REST query (GH-190). */
export type IbcChannelFailure = 'transport' | 'timeout' | 'protocol' | 'decode';

/**
 * Typed IBC channel query failure. Transport, protocol, and schema failures
 * are always distinct from an authoritative empty channel set: callers only
 * see `[]` when the chain itself answers with a valid empty list.
 */
export class IbcChannelError extends Error {
  public readonly originalCause?: unknown;

  constructor(
    public readonly failure: IbcChannelFailure,
    message: string,
    cause?: unknown
  ) {
    super(`IBC channel query failed: ${message}`);
    this.name = 'IbcChannelError';
    this.originalCause = cause;
  }
}

const IBC_CHANNELS_TIMEOUT_MS = 15_000;

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

/**
 * Schema-validate one raw ibc-go channel entry. Only types are checked here;
 * ICS-20 selectability (state, ordering, port, version, hops) is a separate
 * decision in `toTransferChannel` so a valid non-transfer channel on the
 * chain never breaks the query.
 */
function expectIBCChannel(value: unknown, detail: string): IBCChannel {
  try {
    const channel = expectQueryRecord(detail, value);
    const counterparty = expectQueryRecord(
      `${detail}.counterparty`,
      channel.counterparty
    );
    return {
      channel_id: expectQueryString(detail, 'channel_id', channel.channel_id),
      port_id: expectQueryString(detail, 'port_id', channel.port_id),
      state: expectQueryString(detail, 'state', channel.state),
      ordering: expectQueryString(detail, 'ordering', channel.ordering),
      counterparty: {
        channel_id: expectQueryString(
          detail,
          'counterparty.channel_id',
          counterparty.channel_id
        ),
        port_id: expectQueryString(
          detail,
          'counterparty.port_id',
          counterparty.port_id
        ),
      },
      connection_hops: expectQueryStringArray(
        detail,
        'connection_hops',
        channel.connection_hops
      ),
      version: expectQueryString(detail, 'version', channel.version),
    };
  } catch (error) {
    if (error instanceof ModuleQueryError) {
      throw new IbcChannelError('decode', error.message, error);
    }
    throw error;
  }
}

/** Keep read-only network discovery independent from the signing bundle. */
function toSelectableTransferChannel(channel: IBCChannel): TransferChannel | null {
  if (
    channel.port_id !== 'transfer' ||
    channel.state !== 'STATE_OPEN' ||
    channel.ordering !== 'ORDER_UNORDERED' ||
    channel.version !== 'ics20-1' ||
    channel.channel_id.length === 0 ||
    channel.connection_hops.length !== 1 ||
    channel.connection_hops[0].length === 0 ||
    channel.counterparty.port_id.length === 0 ||
    channel.counterparty.channel_id.length === 0
  ) {
    return null;
  }
  return {
    portId: channel.port_id,
    channelId: channel.channel_id,
    counterpartyPortId: channel.counterparty.port_id,
    counterpartyChannelId: channel.counterparty.channel_id,
    connectionId: channel.connection_hops[0],
    version: channel.version,
  };
}

export class NetworkService {
  private readonly queries: ModuleQueryClient;
  private readonly fetchImpl: Fetch;

  constructor(
    private readonly config: ChainConfig,
    queries = new ModuleQueryClient(config),
    fetchImpl: Fetch = (input, init) => globalThis.fetch(input, init)
  ) {
    this.queries = queries;
    this.fetchImpl = fetchImpl;
  }

  /**
   * Get network info from CometBFT RPC /status.
   * No cosmos/base/tendermint/v1beta1 endpoint registered — use RPC directly.
   */
  async getNetworkInfo(): Promise<NetworkInfo | null> {
    try {
      const response = await fetch(`${this.config.rpc}/status`);
      if (!response.ok) return null;

      const data = await response.json();
      const result = data.result;

      return {
        chainId: result.node_info?.network || '',
        latestBlockHeight: parseInt(
          result.sync_info?.latest_block_height || '0',
          10
        ),
        latestBlockTime: result.sync_info?.latest_block_time || '',
        totalValidators: 0, // Updated separately via loadValidators
        nodeInfo: {
          moniker: result.node_info?.moniker || '',
          version: result.node_info?.version || '',
          network: result.node_info?.network || '',
        },
      };
    } catch {
      return null;
    }
  }

  /**
   * Get validators from truedemocracy module.
   * No cosmos/staking module — validators are Proof-of-Domain.
   * Uses the registered truedemocracy gRPC method over ABCI query RPC.
   */
  async getValidators(): Promise<Validator[]> {
    const result = await this.queries.query<unknown>(
      QUERY_PATHS.truedemocracy.validators
    );
    const validators = expectQueryArray<unknown>(
      QUERY_PATHS.truedemocracy.validators,
      result
    ).map(expectValidator);
    return validators.sort((a, b) => b.power - a.power);
  }

  /**
   * Get recent blocks from CometBFT RPC.
   * Uses /block?height=N (not cosmos/base/tendermint/v1beta1).
   */
  async getRecentBlocks(limit: number = 10): Promise<Block[]> {
    try {
      // Get latest block height
      const statusResponse = await fetch(`${this.config.rpc}/status`);
      if (!statusResponse.ok) return [];

      const statusData = await statusResponse.json();
      const latestHeight = parseInt(
        statusData.result.sync_info?.latest_block_height || '0',
        10
      );

      if (latestHeight === 0) return [];

      // Fetch blocks in parallel (limit to avoid too many requests)
      const count = Math.min(limit, latestHeight);
      const promises = Array.from({ length: count }, (_, i) =>
        this.getBlock(latestHeight - i)
      );

      const results = await Promise.allSettled(promises);
      return results
        .filter(
          (r): r is PromiseFulfilledResult<Block | null> =>
            r.status === 'fulfilled' && r.value !== null
        )
        .map((r) => r.value!);
    } catch {
      return [];
    }
  }

  /**
   * Get a single block by height from CometBFT RPC.
   */
  private async getBlock(height: number): Promise<Block | null> {
    try {
      const response = await fetch(
        `${this.config.rpc}/block?height=${height}`
      );
      if (!response.ok) return null;

      const data = await response.json();
      const block = data.result?.block;
      const blockId = data.result?.block_id;

      if (!block) return null;

      return {
        height: parseInt(block.header.height, 10),
        hash: blockId?.hash || '',
        time: block.header.time,
        proposer: block.header.proposer_address || '',
        txCount: block.data?.txs?.length || 0,
      };
    } catch {
      return null;
    }
  }

  /**
   * Get IBC channels (ibc-go is registered).
   *
   * Fail-closed (GH-190): HTTP/network failure, non-JSON bodies, and
   * schema-violating entries throw a typed IbcChannelError; `[]` is returned
   * only for a valid empty chain answer. A non-empty pagination next_key is
   * rejected as a protocol failure instead of silently truncating the set.
   */
  async getIBCChannels(): Promise<IBCChannel[]> {
    const path = '/ibc/core/channel/v1/channels';
    const signal = AbortSignal.timeout(IBC_CHANNELS_TIMEOUT_MS);

    let response: Response;
    try {
      response = await this.fetchImpl(`${this.config.rest}${path}`, {
        signal,
        headers: { accept: 'application/json' },
      });
    } catch (error) {
      if (isTimeoutError(error, signal)) {
        throw new IbcChannelError(
          'timeout',
          `request timed out after ${IBC_CHANNELS_TIMEOUT_MS} ms`,
          error
        );
      }
      throw new IbcChannelError('transport', 'channel request failed', error);
    }
    if (!response.ok) {
      throw new IbcChannelError(
        'transport',
        `channel query returned HTTP ${response.status}`
      );
    }

    let body: unknown;
    try {
      body = await response.json();
    } catch (error) {
      if (isTimeoutError(error, signal)) {
        throw new IbcChannelError(
          'timeout',
          `request timed out after ${IBC_CHANNELS_TIMEOUT_MS} ms`,
          error
        );
      }
      throw new IbcChannelError(
        'decode',
        'channel query response is not JSON',
        error
      );
    }

    let channels: IBCChannel[];
    let pagination: unknown;
    try {
      const envelope = expectQueryRecord(path, body);
      channels = expectQueryArray<unknown>(path, envelope.channels).map(
        (entry, index) => expectIBCChannel(entry, `channels[${index}]`)
      );
      pagination = envelope.pagination;
    } catch (error) {
      if (error instanceof IbcChannelError) throw error;
      if (error instanceof ModuleQueryError) {
        throw new IbcChannelError('decode', error.message, error);
      }
      throw error;
    }

    if (pagination !== undefined && pagination !== null) {
      let nextKey: unknown;
      try {
        nextKey = expectQueryRecord(`${path}.pagination`, pagination).next_key;
      } catch (error) {
        if (error instanceof ModuleQueryError) {
          throw new IbcChannelError('decode', error.message, error);
        }
        throw error;
      }
      if (
        nextKey !== undefined &&
        nextKey !== null &&
        (typeof nextKey !== 'string' || nextKey.length > 0)
      ) {
        throw new IbcChannelError(
          'protocol',
          'paginated channel responses are not supported'
        );
      }
    }

    return channels;
  }

  /**
   * Schema-validated channels narrowed to the strict ICS-20 selectability
   * criteria (STATE_OPEN, ORDER_UNORDERED, port `transfer`, version
   * `ics20-1`, exactly one connection hop, non-empty counterparty). Suitable
   * for the signing path: transport/schema failures throw, and an empty
   * result means the chain verifiably has no selectable transfer channel.
   */
  async getTransferChannels(): Promise<TransferChannel[]> {
    const channels = await this.getIBCChannels();
    return channels.flatMap((channel) => {
      const transferChannel = toSelectableTransferChannel(channel);
      return transferChannel === null ? [] : [transferChannel];
    });
  }
}
