/**
 * Fail-closed ICS-20 IBC transfer core (GH-190).
 *
 * Pure validation, timeout derivation, canonical MsgTransfer construction,
 * send_packet evidence extraction, and evidence-only packet lifecycle
 * reduction for the maintained client. Everything here is repository-local:
 * no relayer, counterparty, or network trust is assumed, broadcast is never
 * equated with delivery, and nothing is ever resubmitted automatically.
 *
 * The sender identity always comes from the authenticated signer, never from
 * a caller-supplied address string.
 */
import { fromBech32 } from '@cosmjs/encoding';
import type { EncodeObject, OfflineSigner } from '@cosmjs/proto-signing';
import type { SigningStargateClient } from '@cosmjs/stargate';
import type { MsgTransfer } from 'cosmjs-types/ibc/applications/transfer/v1/tx';
import type { ChainConfig } from '@/types/chain';
import type {
  IbcClientLatestHeight,
  IbcTransferFailureKind,
  IbcTransferParams,
  IbcTransferPhase,
  IbcTransferSubmission,
  IbcTransferTimeouts,
  PacketLifecycleEvidence,
  SendPacketEvidence,
  TransferChannel,
} from '@/types/ibc';
import type { IBCChannel } from '@/types/network';
import { connectSigningClient, deliverMessages } from './signingClient';

/** Native minimal denom; the only denom this client escrows over IBC. */
export const NATIVE_DENOM = 'upnyx';
export const ICS20_PORT = 'transfer';
export const ICS20_VERSION = 'ics20-1';
export const MSG_TRANSFER_TYPE_URL =
  '/ibc.applications.transfer.v1.MsgTransfer';

export const MIN_TIMEOUT_HEIGHT_MARGIN = 1n;
export const MAX_TIMEOUT_HEIGHT_MARGIN = 1_000_000n;
export const DEFAULT_TIMEOUT_HEIGHT_MARGIN = 1_000n;
export const MIN_TIMEOUT_TIMESTAMP_MARGIN_NS = 60_000_000_000n; // 1 minute
export const MAX_TIMEOUT_TIMESTAMP_MARGIN_NS = 86_400_000_000_000n; // 24 hours
export const DEFAULT_TIMEOUT_TIMESTAMP_MARGIN_NS = 600_000_000_000n; // 10 minutes

const IBC_QUERY_TIMEOUT_MS = 15_000;

/** Typed failure of the IBC transfer core. */
export class IbcTransferError extends Error {
  public readonly originalCause?: unknown;

  constructor(
    public readonly kind: IbcTransferFailureKind,
    message: string,
    cause?: unknown
  ) {
    super(`IBC transfer failed (${kind}): ${message}`);
    this.name = 'IbcTransferError';
    this.originalCause = cause;
  }
}

/* ------------------------------------------------------------------------ */
/* amount and denom validation (BigInt only, never floats)                   */
/* ------------------------------------------------------------------------ */

/** Assert that the configured minimal denom is the native upnyx denom. */
export function assertNativeDenom(
  config: Pick<ChainConfig, 'coinMinimalDenom'>
): typeof NATIVE_DENOM {
  if (config.coinMinimalDenom !== NATIVE_DENOM) {
    throw new IbcTransferError(
      'invalid_denom',
      `native denom must be ${NATIVE_DENOM}, got ${config.coinMinimalDenom}`
    );
  }
  return NATIVE_DENOM;
}

/**
 * Parse a strict positive decimal PNYX amount into its canonical base-denom
 * (upnyx) decimal string. Rejects non-decimal input, zero, fractions beyond
 * the configured decimals, and any notation (exponent, sign, whitespace)
 * that would require coercion. All arithmetic is BigInt.
 */
export function parseTransferAmount(amount: string, decimals: number): string {
  if (!Number.isSafeInteger(decimals) || decimals < 0 || decimals > 18) {
    throw new IbcTransferError(
      'invalid_amount',
      'decimals must be an integer between 0 and 18'
    );
  }
  if (typeof amount !== 'string' || !/^[0-9]+(\.[0-9]+)?$/.test(amount)) {
    throw new IbcTransferError(
      'invalid_amount',
      'amount must be a positive decimal string'
    );
  }
  const [whole, fraction = ''] = amount.split('.');
  if (fraction.length > decimals) {
    throw new IbcTransferError(
      'invalid_amount',
      `amount exceeds ${decimals} decimal places`
    );
  }
  const scale = 10n ** BigInt(decimals);
  const base =
    BigInt(whole) * scale +
    (fraction.length > 0 ? BigInt(fraction.padEnd(decimals, '0')) : 0n);
  if (base <= 0n) {
    throw new IbcTransferError('invalid_amount', 'amount must be positive');
  }
  return base.toString();
}

/* ------------------------------------------------------------------------ */
/* receiver validation                                                       */
/* ------------------------------------------------------------------------ */

/**
 * Validate generic bech32 receiver syntax. This is syntax-only: a valid
 * prefix does NOT prove the address belongs to the intended counterparty
 * network, and callers must not present it as such.
 */
export function assertBech32Receiver(receiver: string): string {
  try {
    const { prefix, data } = fromBech32(receiver);
    if (prefix.length === 0 || data.length === 0) {
      throw new Error('empty bech32 prefix or data');
    }
  } catch (error) {
    throw new IbcTransferError(
      'invalid_receiver',
      'receiver must be a syntactically valid bech32 address',
      error
    );
  }
  return receiver;
}

/* ------------------------------------------------------------------------ */
/* channel selectability                                                     */
/* ------------------------------------------------------------------------ */

/**
 * Strict ICS-20 selectability for a schema-validated channel: STATE_OPEN,
 * ORDER_UNORDERED, port `transfer`, version `ics20-1`, exactly one non-empty
 * connection hop, and a non-empty counterparty port/channel.
 */
export function isSelectableTransferChannel(channel: IBCChannel): boolean {
  return (
    channel.port_id === ICS20_PORT &&
    channel.state === 'STATE_OPEN' &&
    channel.ordering === 'ORDER_UNORDERED' &&
    channel.version === ICS20_VERSION &&
    channel.channel_id.length > 0 &&
    channel.connection_hops.length === 1 &&
    channel.connection_hops[0].length > 0 &&
    channel.counterparty.port_id.length > 0 &&
    channel.counterparty.channel_id.length > 0
  );
}

/**
 * Narrow a schema-validated channel to a TransferChannel. Only call for
 * channels that satisfy `isSelectableTransferChannel`.
 */
export function toTransferChannel(channel: IBCChannel): TransferChannel {
  return {
    portId: channel.port_id,
    channelId: channel.channel_id,
    counterpartyPortId: channel.counterparty.port_id,
    counterpartyChannelId: channel.counterparty.channel_id,
    connectionId: channel.connection_hops[0],
    version: channel.version,
  };
}

/**
 * Narrow a channel for the signing path or fail closed with a typed
 * invalid_channel error naming the unmet criterion.
 */
export function assertTransferChannel(channel: IBCChannel): TransferChannel {
  if (!isSelectableTransferChannel(channel)) {
    throw new IbcTransferError(
      'invalid_channel',
      `channel ${channel.channel_id || '(unnamed)'} is not a selectable ICS-20 channel ` +
        '(requires STATE_OPEN, ORDER_UNORDERED, port transfer, version ics20-1, ' +
        'exactly one connection hop, non-empty counterparty)'
    );
  }
  return toTransferChannel(channel);
}

/* ------------------------------------------------------------------------ */
/* timeout derivation (bounded, BigInt only)                                 */
/* ------------------------------------------------------------------------ */

export interface DeriveTransferTimeoutOptions {
  /** Blocks added to the client-state latest height. */
  heightMargin?: bigint;
  /** Nanoseconds added to the current time. */
  timestampMarginNs?: bigint;
  /** Current time in nanoseconds since the Unix epoch (tests inject this). */
  nowUnixNs?: bigint;
}

function expectMargin(
  value: bigint,
  min: bigint,
  max: bigint,
  field: string
): bigint {
  if (value < min || value > max) {
    throw new IbcTransferError(
      'invalid_timeout',
      `${field} must be within [${min}, ${max}]`
    );
  }
  return value;
}

/**
 * Derive bounded absolute packet timeouts from the schema-validated
 * home-chain IBC client-state latest height plus explicit bounded margins.
 * The height timeout stays on the client's own revision; the timestamp
 * timeout is wall-clock now plus the margin. All values are BigInt.
 */
export function deriveTransferTimeouts(
  latestHeight: IbcClientLatestHeight,
  options: DeriveTransferTimeoutOptions = {}
): IbcTransferTimeouts {
  if (
    latestHeight.revisionNumber < 0n ||
    latestHeight.revisionHeight <= 0n
  ) {
    throw new IbcTransferError(
      'invalid_client_state',
      'client-state latest height must have a non-negative revision number and a positive revision height'
    );
  }
  const heightMargin = expectMargin(
    options.heightMargin ?? DEFAULT_TIMEOUT_HEIGHT_MARGIN,
    MIN_TIMEOUT_HEIGHT_MARGIN,
    MAX_TIMEOUT_HEIGHT_MARGIN,
    'height margin'
  );
  const timestampMarginNs = expectMargin(
    options.timestampMarginNs ?? DEFAULT_TIMEOUT_TIMESTAMP_MARGIN_NS,
    MIN_TIMEOUT_TIMESTAMP_MARGIN_NS,
    MAX_TIMEOUT_TIMESTAMP_MARGIN_NS,
    'timestamp margin'
  );
  const nowUnixNs =
    options.nowUnixNs ?? BigInt(Date.now()) * 1_000_000n;
  if (nowUnixNs <= 0n) {
    throw new IbcTransferError(
      'invalid_timeout',
      'current time must be positive nanoseconds since the Unix epoch'
    );
  }
  return {
    timeoutHeight: {
      revisionNumber: latestHeight.revisionNumber,
      revisionHeight: latestHeight.revisionHeight + heightMargin,
    },
    timeoutTimestamp: nowUnixNs + timestampMarginNs,
  };
}

/* ------------------------------------------------------------------------ */
/* home-chain IBC client-state query (schema-validated from unknown)         */
/* ------------------------------------------------------------------------ */

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
    throw new IbcTransferError('decode', `${detail} must be an object`);
  }
  return value as Record<string, unknown>;
}

function expectDigits(value: unknown, detail: string): string {
  if (typeof value !== 'string' || !/^[0-9]+$/.test(value)) {
    throw new IbcTransferError(
      'decode',
      `${detail} must be an unsigned decimal string`
    );
  }
  return value;
}

async function fetchJson(
  fetchImpl: Fetch,
  url: string,
  detail: string
): Promise<unknown> {
  const signal = AbortSignal.timeout(IBC_QUERY_TIMEOUT_MS);
  let response: Response;
  try {
    response = await fetchImpl(url, {
      signal,
      headers: { accept: 'application/json' },
    });
  } catch (error) {
    if (isTimeoutError(error, signal)) {
      throw new IbcTransferError(
        'transport',
        `${detail} request timed out after ${IBC_QUERY_TIMEOUT_MS} ms`,
        error
      );
    }
    throw new IbcTransferError(
      'transport',
      `${detail} request failed`,
      error
    );
  }
  if (!response.ok) {
    throw new IbcTransferError(
      'transport',
      `${detail} query returned HTTP ${response.status}`
    );
  }
  try {
    return await response.json();
  } catch (error) {
    throw new IbcTransferError(
      'decode',
      `${detail} response is not JSON`,
      error
    );
  }
}

/**
 * Resolve the home-chain IBC light-client latest height behind a channel's
 * connection hop via the ibc-go REST gateway. Both responses are parsed from
 * `unknown` with explicit validation; any transport or schema failure is a
 * typed error, never a fabricated height.
 */
export async function fetchClientLatestHeight(
  rest: string,
  connectionId: string,
  fetchImpl: Fetch = (input, init) => globalThis.fetch(input, init)
): Promise<IbcClientLatestHeight> {
  if (!/^connection-[0-9]+$/.test(connectionId)) {
    throw new IbcTransferError(
      'invalid_channel',
      'connection hop must be a connection identifier'
    );
  }

  const connectionBody = await fetchJson(
    fetchImpl,
    `${rest}/ibc/core/connection/v1/connections/${encodeURIComponent(connectionId)}`,
    'connection'
  );
  const connection = expectRecord(
    expectRecord(connectionBody, 'connection response').connection,
    'connection response connection'
  );
  const clientId = connection.client_id;
  if (typeof clientId !== 'string' || clientId.length === 0) {
    throw new IbcTransferError(
      'invalid_client_state',
      'connection response carries no client identifier'
    );
  }

  const clientBody = await fetchJson(
    fetchImpl,
    `${rest}/ibc/core/client/v1/client_states/${encodeURIComponent(clientId)}`,
    'client state'
  );
  const clientState = expectRecord(
    expectRecord(clientBody, 'client state response').client_state,
    'client state response client_state'
  );
  const latestHeight = expectRecord(
    clientState.latest_height,
    'client state latest_height'
  );
  const revisionNumber = BigInt(
    expectDigits(latestHeight.revision_number, 'latest_height.revision_number')
  );
  const revisionHeight = BigInt(
    expectDigits(latestHeight.revision_height, 'latest_height.revision_height')
  );
  if (revisionHeight <= 0n) {
    throw new IbcTransferError(
      'invalid_client_state',
      'client-state latest revision height must be positive'
    );
  }
  return { revisionNumber, revisionHeight };
}

/* ------------------------------------------------------------------------ */
/* canonical MsgTransfer construction                                        */
/* ------------------------------------------------------------------------ */

/**
 * Validate the transfer input end to end and build the canonical MsgTransfer
 * EncodeObject for the fail-closed registry. The sender must be the
 * authenticated signer's address. Amount and denom are validated with
 * BigInt-only arithmetic; the receiver is syntax-validated bech32; the
 * channel must satisfy the strict ICS-20 criteria; the timeouts must be
 * positive.
 */
export function buildMsgTransferEncodeObject(
  config: Pick<ChainConfig, 'coinMinimalDenom' | 'coinDecimals'>,
  sender: string,
  params: IbcTransferParams,
  timeouts: IbcTransferTimeouts
): EncodeObject {
  const denom = assertNativeDenom(config);
  const amount = parseTransferAmount(params.amount, config.coinDecimals);
  const receiver = assertBech32Receiver(params.receiver);
  if (
    timeouts.timeoutHeight.revisionNumber < 0n ||
    timeouts.timeoutHeight.revisionHeight <= 0n ||
    timeouts.timeoutTimestamp <= 0n
  ) {
    throw new IbcTransferError(
      'invalid_timeout',
      'timeout height and timestamp must be positive'
    );
  }
  const value: MsgTransfer = {
    sourcePort: params.channel.portId,
    sourceChannel: params.channel.channelId,
    token: { denom, amount },
    sender,
    receiver,
    timeoutHeight: {
      revisionNumber: timeouts.timeoutHeight.revisionNumber,
      revisionHeight: timeouts.timeoutHeight.revisionHeight,
    },
    timeoutTimestamp: timeouts.timeoutTimestamp,
    memo: params.memo ?? '',
    encoding: '',
  };
  return { typeUrl: MSG_TRANSFER_TYPE_URL, value };
}

/* ------------------------------------------------------------------------ */
/* send_packet evidence extraction and reconciliation                        */
/* ------------------------------------------------------------------------ */

/** Minimal committed-transaction event shape used for evidence extraction. */
export interface TxEventLike {
  readonly type: string;
  readonly attributes: readonly { readonly key: string; readonly value: string }[];
}

/** Minimal indexed-transaction shape used for evidence extraction. */
export interface IndexedTxLike {
  readonly height: number;
  readonly hash: string;
  readonly code: number;
  readonly events: readonly TxEventLike[];
}

/** Minimal read-only transaction search surface used for manual recovery. */
export interface PacketSearchClient {
  searchTx: (
    query: readonly { readonly key: string; readonly value: string }[]
  ) => Promise<readonly IndexedTxLike[]>;
}

/** Result of reconciling send_packet events for one committed transfer. */
export type SendPacketExtraction =
  | { readonly kind: 'evidenced'; readonly packet: SendPacketEvidence }
  | { readonly kind: 'absent' }
  | { readonly kind: 'contradictory' };

/**
 * Extract the send_packet evidence of a single committed MsgTransfer from
 * its transaction events, reconciled against the expected source channel and
 * counterparty. Zero matching events is `absent` (insufficient evidence);
 * multiple or malformed matching events is `contradictory`.
 */
export function extractSendPacket(
  events: readonly TxEventLike[],
  expected: TransferChannel
): SendPacketExtraction {
  const matching: SendPacketEvidence[] = [];
  let malformed = false;

  for (const event of events) {
    if (event.type !== 'send_packet') continue;
    const attributes = new Map<string, string>();
    let duplicate = false;
    for (const attribute of event.attributes) {
      if (attributes.has(attribute.key)) duplicate = true;
      attributes.set(attribute.key, attribute.value);
    }
    if (duplicate) {
      malformed = true;
      continue;
    }
    if (
      attributes.get('packet_src_port') !== expected.portId ||
      attributes.get('packet_src_channel') !== expected.channelId
    ) {
      continue;
    }
    const sequenceText = attributes.get('packet_sequence');
    const destinationPort = attributes.get('packet_dst_port');
    const destinationChannel = attributes.get('packet_dst_channel');
    if (
      typeof sequenceText !== 'string' ||
      !/^[0-9]+$/.test(sequenceText) ||
      typeof destinationPort !== 'string' ||
      typeof destinationChannel !== 'string'
    ) {
      malformed = true;
      continue;
    }
    if (
      destinationPort !== expected.counterpartyPortId ||
      destinationChannel !== expected.counterpartyChannelId
    ) {
      malformed = true;
      continue;
    }
    matching.push({
      sequence: BigInt(sequenceText),
      sourcePort: expected.portId,
      sourceChannel: expected.channelId,
      destinationPort,
      destinationChannel,
    });
  }

  if (malformed) return { kind: 'contradictory' };
  if (matching.length === 0) return { kind: 'absent' };
  if (matching.length !== 1) return { kind: 'contradictory' };
  return { kind: 'evidenced', packet: matching[0] };
}

/** Evidence lookup result for one committed transaction hash. */
export interface CommittedSendPacket {
  readonly txHash: string;
  readonly height: number;
  readonly code: number;
  readonly extraction: SendPacketExtraction;
}

/**
 * Fetch a committed transaction by hash and reconcile its send_packet
 * evidence. A missing transaction is `absent` evidence; transport failure is
 * a typed error. The transaction code is reported verbatim — a non-zero code
 * means the chain rejected the message and no packet was escrowed.
 */
export async function getCommittedSendPacket(
  client: Pick<SigningStargateClient, 'getTx'>,
  txHash: string,
  expected: TransferChannel
): Promise<CommittedSendPacket> {
  if (!/^[0-9A-Fa-f]{64}$/.test(txHash)) {
    throw new IbcTransferError(
      'evidence_unavailable',
      'transaction hash must be 64 hex characters'
    );
  }
  let tx: IndexedTxLike | null;
  try {
    tx = (await client.getTx(txHash.toUpperCase())) as IndexedTxLike | null;
  } catch (error) {
    throw new IbcTransferError(
      'transport',
      'committed transaction lookup failed',
      error
    );
  }
  if (tx === null) {
    return {
      txHash: txHash.toUpperCase(),
      height: 0,
      code: 0,
      extraction: { kind: 'absent' },
    };
  }
  return {
    txHash: tx.hash,
    height: tx.height,
    code: tx.code,
    extraction: extractSendPacket(tx.events, expected),
  };
}

/* ------------------------------------------------------------------------ */
/* evidence-only packet lifecycle reduction                                  */
/* ------------------------------------------------------------------------ */

/**
 * Reduce source-chain evidence to a packet lifecycle phase. Evidence is the
 * only input: wall time never implies a timeout, and contradictory or
 * insufficient evidence always reduces to `unknown`. An acknowledgement or
 * timeout without a committed send_packet is inconsistent evidence.
 */
export function reducePacketLifecycle(
  evidence: PacketLifecycleEvidence
): IbcTransferPhase {
  const { sendPacket, acknowledgement, timedOut } = evidence;
  if (acknowledgement !== null && timedOut === true) return 'unknown';
  if (sendPacket === null) return 'unknown';
  if (acknowledgement === 'success') return 'acknowledged';
  if (acknowledgement === 'error') return 'acknowledged_error_or_refunded';
  if (timedOut === true) return 'timed_out_refunded';
  return 'committed_pending_relay';
}

function eventAttributes(event: TxEventLike): Map<string, string> | null {
  const attributes = new Map<string, string>();
  for (const attribute of event.attributes) {
    if (attributes.has(attribute.key)) return null;
    attributes.set(attribute.key, attribute.value);
  }
  return attributes;
}

function packetEventMatches(
  event: TxEventLike,
  eventType: 'acknowledge_packet' | 'timeout_packet',
  packet: SendPacketEvidence
): boolean {
  if (event.type !== eventType) return false;
  const attributes = eventAttributes(event);
  return (
    attributes !== null &&
    attributes.get('packet_src_port') === packet.sourcePort &&
    attributes.get('packet_src_channel') === packet.sourceChannel &&
    attributes.get('packet_dst_port') === packet.destinationPort &&
    attributes.get('packet_dst_channel') === packet.destinationChannel &&
    attributes.get('packet_sequence') === packet.sequence.toString()
  );
}

function parseAcknowledgement(value: string): 'success' | 'error' | null {
  const candidates = [value];
  try {
    const bytes = Uint8Array.from(globalThis.atob(value), (character) =>
      character.charCodeAt(0)
    );
    candidates.push(new TextDecoder().decode(bytes));
  } catch {
    // The event may already expose acknowledgement JSON. Unknown encodings
    // stay unknown instead of being guessed.
  }
  for (const candidate of candidates) {
    try {
      const parsed: unknown = JSON.parse(candidate);
      if (typeof parsed !== 'object' || parsed === null || Array.isArray(parsed)) {
        continue;
      }
      const acknowledgement = parsed as Record<string, unknown>;
      if (
        typeof acknowledgement.result === 'string' &&
        acknowledgement.result.length > 0 &&
        acknowledgement.error === undefined
      ) {
        return 'success';
      }
      if (
        typeof acknowledgement.error === 'string' &&
        acknowledgement.error.length > 0 &&
        acknowledgement.result === undefined
      ) {
        return 'error';
      }
    } catch {
      // Try the next exact representation.
    }
  }
  return null;
}

function parseAcknowledgementHex(value: string): 'success' | 'error' | null {
  if (!/^[0-9A-Fa-f]+$/.test(value) || value.length % 2 !== 0) return null;
  const bytes = new Uint8Array(value.length / 2);
  for (let index = 0; index < bytes.length; index++) {
    bytes[index] = Number.parseInt(value.slice(index * 2, index * 2 + 2), 16);
  }
  return parseAcknowledgement(new TextDecoder().decode(bytes));
}

/**
 * Query source-chain acknowledgement and timeout events for a committed
 * packet. This is a read-only, user-triggered recovery check: absence of an
 * indexed event means pending, malformed/contradictory evidence means
 * unknown, and the transfer is never resubmitted.
 */
export async function reconcilePacketLifecycle(
  client: PacketSearchClient,
  packet: SendPacketEvidence
): Promise<PacketLifecycleEvidence> {
  let acknowledgementTxs: readonly IndexedTxLike[];
  let timeoutTxs: readonly IndexedTxLike[];
  try {
    [acknowledgementTxs, timeoutTxs] = await Promise.all([
      client.searchTx([
        { key: 'acknowledge_packet.packet_src_port', value: packet.sourcePort },
        { key: 'acknowledge_packet.packet_src_channel', value: packet.sourceChannel },
        { key: 'acknowledge_packet.packet_sequence', value: packet.sequence.toString() },
      ]),
      client.searchTx([
        { key: 'timeout_packet.packet_src_port', value: packet.sourcePort },
        { key: 'timeout_packet.packet_src_channel', value: packet.sourceChannel },
        { key: 'timeout_packet.packet_sequence', value: packet.sequence.toString() },
      ]),
    ]);
  } catch (error) {
    throw new IbcTransferError(
      'transport',
      'packet lifecycle evidence lookup failed',
      error
    );
  }
  const acknowledgementEvents = acknowledgementTxs.flatMap((tx) =>
    tx.code === 0
      ? tx.events.filter((event) =>
          packetEventMatches(event, 'acknowledge_packet', packet)
        )
      : []
  );
  const timeoutEvents = timeoutTxs.flatMap((tx) =>
    tx.code === 0
      ? tx.events.filter((event) =>
          packetEventMatches(event, 'timeout_packet', packet)
        )
      : []
  );

  const malformedIndexedEvent = [...acknowledgementTxs, ...timeoutTxs].some(
    (tx) =>
      tx.code === 0 &&
      tx.events.some(
        (event) =>
          (event.type === 'acknowledge_packet' ||
            event.type === 'timeout_packet') &&
          eventAttributes(event) === null
      )
  );

  if (
    malformedIndexedEvent ||
    acknowledgementEvents.length > 1 ||
    timeoutEvents.length > 1
  ) {
    return { sendPacket: null, acknowledgement: null, timedOut: null };
  }
  if (acknowledgementEvents.length === 1 && timeoutEvents.length === 1) {
    return { sendPacket: null, acknowledgement: null, timedOut: null };
  }
  if (timeoutEvents.length === 1) {
    return { sendPacket: packet, acknowledgement: null, timedOut: true };
  }
  if (acknowledgementEvents.length === 1) {
    const attributes = eventAttributes(acknowledgementEvents[0]);
    const acknowledgement = attributes?.get('packet_ack');
    const acknowledgementHex = attributes?.get('packet_ack_hex');
    const outcome =
      acknowledgementHex === undefined
        ? acknowledgement === undefined
          ? null
          : parseAcknowledgement(acknowledgement)
        : parseAcknowledgementHex(acknowledgementHex);
    return {
      sendPacket: outcome === null ? null : packet,
      acknowledgement: outcome,
      timedOut: null,
    };
  }
  return { sendPacket: packet, acknowledgement: null, timedOut: null };
}

/* ------------------------------------------------------------------------ */
/* transfer submission                                                       */
/* ------------------------------------------------------------------------ */

export interface SendIbcTransferOptions {
  /** Lifecycle observer; intermediate phases are best-effort notifications. */
  onPhase?: (phase: IbcTransferPhase) => void;
  /** Caller-derived timeouts; fetched and derived when absent. */
  timeouts?: IbcTransferTimeouts;
  heightMargin?: bigint;
  timestampMarginNs?: bigint;
  nowUnixNs?: bigint;
  fetchImpl?: Fetch;
}

/**
 * Validate, sign, and broadcast one ICS-20 transfer through the canonical
 * signing path (connectSigningClient + deliverMessages). The sender is the
 * authenticated signer's account. Broadcast is never reported as delivery:
 * the result is `committed_pending_relay` only when the committed
 * transaction carries exactly one reconciled send_packet event, and
 * `unknown` otherwise. A failed broadcast throws and is never retried.
 */
export async function sendIbcTransfer(
  config: ChainConfig,
  signer: OfflineSigner,
  params: IbcTransferParams,
  options: SendIbcTransferOptions = {}
): Promise<IbcTransferSubmission> {
  const { onPhase } = options;
  onPhase?.('validating');

  assertNativeDenom(config);
  const channel = params.channel;
  const timeouts =
    options.timeouts ??
    deriveTransferTimeouts(
      await fetchClientLatestHeight(
        config.rest,
        channel.connectionId,
        options.fetchImpl
      ),
      {
        heightMargin: options.heightMargin,
        timestampMarginNs: options.timestampMarginNs,
        nowUnixNs: options.nowUnixNs,
      }
    );

  const [account] = await signer.getAccounts();
  if (!account) {
    throw new IbcTransferError(
      'broadcast',
      'signer exposes no account to sign with'
    );
  }
  const message = buildMsgTransferEncodeObject(
    config,
    account.address,
    params,
    timeouts
  );

  let client: SigningStargateClient | undefined;
  try {
    try {
      client = await connectSigningClient(config, signer);
    } catch (error) {
      throw new IbcTransferError(
        'transport',
        'signing client connection failed',
        error
      );
    }

    onPhase?.('signing');
    // deliverMessages simulates, signs, and broadcasts as one step; both
    // phases are in flight for the duration of the await.
    onPhase?.('broadcasting');
    let result;
    try {
      result = await deliverMessages(
        client,
        account.address,
        [message],
        config.gasPrice,
        ''
      );
    } catch (error) {
      throw new IbcTransferError(
        'broadcast',
        error instanceof Error ? error.message : 'broadcast failed',
        error
      );
    }

    let packet: SendPacketEvidence | null = null;
    let phase: IbcTransferPhase = 'unknown';
    try {
      const committed = await getCommittedSendPacket(
        client,
        result.hash,
        channel
      );
      if (committed.extraction.kind === 'evidenced') {
        packet = committed.extraction.packet;
        phase = 'committed_pending_relay';
      }
    } catch {
      // Evidence lookup failure after a committed broadcast must not throw:
      // the transaction hash stays with the caller and the phase is unknown.
      phase = 'unknown';
    }
    onPhase?.(phase);
    return { phase, txHash: result.hash, height: result.height, packet };
  } finally {
    client?.disconnect();
  }
}
