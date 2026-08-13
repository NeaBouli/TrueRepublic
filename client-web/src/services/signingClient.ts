/**
 * Canonical signing-client factory and delivery helper (GH-115).
 *
 * Every browser signing path must obtain its SigningStargateClient through
 * `connectSigningClient` so that exactly the canonical registry from
 * `txRegistry.ts` (default types plus the fail-closed custom type set) is
 * used, and must deliver transactions through `deliverMessages` so that
 * simulation, gas, fee, and disconnect behavior stay identical everywhere.
 *
 * Read-only queries may continue to use plain `StargateClient.connect`.
 */
import {
  calculateFee,
  GasPrice,
  SigningStargateClient,
} from '@cosmjs/stargate';
import { fromBech32 } from '@cosmjs/encoding';
import type { EncodeObject } from '@cosmjs/proto-signing';
import type { OfflineSigner } from '@cosmjs/proto-signing';
import type { ChainConfig } from '@/types/chain';
import type { TransactionResult } from '@/types/transaction';
import { createTxRegistry } from './txRegistry';

// Store-heavy module messages consume materially more gas during delivery
// than the SDK simulation reports. Keep one conservative, tested margin for
// every maintained browser transaction family.
export const GAS_ADJUSTMENT = 2;

/**
 * Connect a signing client bound to the canonical custom transaction
 * registry. The gas price stays config-driven.
 *
 * Scoping is fail-closed: the signer must expose at least one account and
 * every account must carry this chain's bech32 prefix, and the connected
 * endpoint must report the configured chain ID. A prefix or chain mismatch
 * rejects before any message is signed, so a wallet from another network or
 * a misconfigured endpoint can never produce a signature for this chain.
 */
export async function connectSigningClient(
  config: ChainConfig,
  signer: OfflineSigner
): Promise<SigningStargateClient> {
  const accounts = await signer.getAccounts();
  if (accounts.length === 0) {
    throw new Error('Signer exposes no account to sign with');
  }
  for (const account of accounts) {
    let prefix: string;
    let byteLength: number;
    try {
      const decoded = fromBech32(account.address);
      prefix = decoded.prefix;
      byteLength = decoded.data.length;
    } catch {
      throw new Error('Signer account address is not valid bech32');
    }
    if (prefix !== config.bech32Prefix || byteLength !== 20) {
      throw new Error(
        `Signer account does not belong to the ${config.bech32Prefix} network`
      );
    }
  }

  const client = await SigningStargateClient.connectWithSigner(
    config.rpc,
    signer,
    {
      registry: createTxRegistry(),
    }
  );
  try {
    const chainId = await client.getChainId();
    if (chainId !== config.chainId) {
      throw new Error(
        `Connected endpoint reports chain ${chainId}, expected ${config.chainId}`
      );
    }
  } catch (error) {
    client.disconnect();
    throw error;
  }
  return client;
}

/**
 * Simulate, sign, and broadcast messages. Simulation failure aborts the
 * delivery — there is deliberately no fallback gas that would mask it. The
 * fee stays derived from the configured gas price. A non-zero delivery code
 * throws with the chain's raw log, or with the code itself when the log is
 * empty. The caller owns `client.disconnect()` (finally).
 */
export async function deliverMessages(
  client: SigningStargateClient,
  senderAddress: string,
  messages: EncodeObject[],
  gasPrice: string,
  memo = ''
): Promise<TransactionResult> {
  const gasEstimate = await client.simulate(senderAddress, messages, memo);
  const gas = Math.ceil(gasEstimate * GAS_ADJUSTMENT);

  const result = await client.signAndBroadcast(
    senderAddress,
    messages,
    calculateFee(gas, GasPrice.fromString(gasPrice)),
    memo
  );

  if (result.code !== 0) {
    throw new Error(
      result.rawLog || `Transaction failed with code ${result.code}`
    );
  }

  return {
    hash: result.transactionHash,
    height: result.height,
    success: true,
  };
}
