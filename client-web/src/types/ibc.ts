/**
 * ICS-20 IBC transfer types for the maintained client (GH-190).
 *
 * These types are the pure contract between the fail-closed transfer core in
 * `services/ibcTransfer.ts` and the store/UI layer. They deliberately carry
 * no implementation details: every value here has already passed strict
 * validation, and every lifecycle phase is derived from source-chain
 * evidence only — never from wall time or from a bare broadcast result.
 */

/**
 * A schema-validated, selectable ICS-20 transfer channel. A channel only
 * qualifies when it is STATE_OPEN, ORDER_UNORDERED, bound to the `transfer`
 * port with version `ics20-1`, has exactly one non-empty connection hop, and
 * names a non-empty counterparty port and channel.
 */
export interface TransferChannel {
  /** Always `transfer` for ICS-20. */
  portId: string;
  channelId: string;
  counterpartyPortId: string;
  counterpartyChannelId: string;
  /** The single connection hop of the channel. */
  connectionId: string;
  /** Always `ics20-1` for ICS-20. */
  version: string;
}

/** Latest height of the home-chain IBC light client for the counterparty. */
export interface IbcClientLatestHeight {
  revisionNumber: bigint;
  revisionHeight: bigint;
}

/**
 * Absolute packet timeout derived from the home-chain client-state latest
 * height plus explicit bounded margins. Nanoseconds are always BigInt; no
 * float arithmetic is involved anywhere in the timeout path.
 */
export interface IbcTransferTimeouts {
  timeoutHeight: IbcClientLatestHeight;
  /** Absolute nanoseconds since the Unix epoch. */
  timeoutTimestamp: bigint;
}

/** User-supplied ICS-20 transfer input, validated before any signing. */
export interface IbcTransferParams {
  channel: TransferChannel;
  /**
   * Bech32 receiver on the counterparty chain. Only the bech32 syntax is
   * validated; a syntactically valid prefix never proves that the address
   * belongs to the intended counterparty network.
   */
  receiver: string;
  /** Strict positive decimal amount in whole PNYX, e.g. "1.25". */
  amount: string;
  memo?: string;
}

/** Typed failure kinds of the IBC transfer core. */
export type IbcTransferFailureKind =
  | 'invalid_amount'
  | 'invalid_denom'
  | 'invalid_receiver'
  | 'invalid_channel'
  | 'invalid_client_state'
  | 'invalid_timeout'
  | 'transport'
  | 'decode'
  | 'broadcast'
  | 'evidence_unavailable'
  | 'evidence_contradictory';

/**
 * Packet lifecycle phases. Broadcast is never equated with delivery: a
 * committed transfer stays `committed_pending_relay` until source-chain
 * evidence of an acknowledgement or timeout exists. Contradictory or
 * insufficient evidence always reduces to `unknown`.
 */
export type IbcTransferPhase =
  | 'validating'
  | 'signing'
  | 'broadcasting'
  | 'committed_pending_relay'
  | 'acknowledged'
  | 'acknowledged_error_or_refunded'
  | 'timed_out_refunded'
  | 'unknown';

/** send_packet evidence extracted from a committed source-chain transaction. */
export interface SendPacketEvidence {
  /** Packet sequence on the source channel. */
  sequence: bigint;
  sourcePort: string;
  sourceChannel: string;
  destinationPort: string;
  destinationChannel: string;
}

/**
 * Source-chain evidence about one escrowed packet. Every field is evidence,
 * never inference: `timedOut` may only be `true` when the source chain shows
 * the packet timed out, not because wall time passed.
 */
export interface PacketLifecycleEvidence {
  /** Committed send_packet on the source chain; null when not evidenced. */
  sendPacket: SendPacketEvidence | null;
  /** Acknowledgement outcome evidenced on the source chain; null when none. */
  acknowledgement: 'success' | 'error' | null;
  /** Source-chain timeout evidence; null when none exists. */
  timedOut: boolean | null;
}

/** Result of a committed ICS-20 transfer broadcast. */
export interface IbcTransferSubmission {
  /**
   * `committed_pending_relay` only when the committed transaction carries
   * exactly one matching send_packet event; otherwise `unknown`.
   */
  phase: IbcTransferPhase;
  txHash: string;
  height: number;
  packet: SendPacketEvidence | null;
}

/** Persistable packet evidence: BigInt sequence is an exact decimal string. */
export interface PersistedSendPacketEvidence {
  sequence: string;
  sourcePort: string;
  sourceChannel: string;
  destinationPort: string;
  destinationChannel: string;
}

/** Non-secret, chain-and-wallet-scoped recovery record. */
export interface IbcTransferRecord {
  txHash: string;
  chainId: string;
  walletAddress: string;
  channel: TransferChannel;
  receiver: string;
  /** Exact base-denom amount in upnyx. */
  amount: string;
  memo: string;
  height: number;
  phase: IbcTransferPhase;
  packet: PersistedSendPacketEvidence | null;
  submittedAt: number;
}
