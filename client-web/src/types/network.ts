/**
 * Network types for the blockchain explorer UI.
 * Validator fields match Go x/truedemocracy Validator struct (PoD consensus).
 * Block queries use CometBFT RPC (no cosmos/base/tendermint/v1beta1 registered).
 */

/** Network overview from CometBFT /status + bank supply */
export interface NetworkInfo {
  chainId: string;
  latestBlockHeight: number;
  latestBlockTime: string;
  totalValidators: number;
  nodeInfo: {
    moniker: string;
    version: string;
    network: string;
  };
}

/**
 * Validator mirrors Go truedemocracy.Validator (Proof-of-Domain consensus).
 * No commission/delegatorShares — those are Cosmos x/staking concepts.
 */
export interface Validator {
  operator_addr: string;
  stake: { denom: string; amount: string }[];
  domains: string[];
  power: number;
  jailed: boolean;
  jailed_until: number;
  missed_blocks: number;
}

/** Block from CometBFT RPC /block?height=N */
export interface Block {
  height: number;
  hash: string;
  time: string;
  proposer: string;
  txCount: number;
}

/** IBC channel from ibc-go (ibc/core/channel/v1/channels) */
export interface IBCChannel {
  channel_id: string;
  port_id: string;
  state: string;
  counterparty: {
    channel_id: string;
    port_id: string;
  };
  connection_hops: string[];
  version: string;
}
