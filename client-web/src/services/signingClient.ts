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
 */
export async function connectSigningClient(
  config: ChainConfig,
  signer: OfflineSigner
): Promise<SigningStargateClient> {
  return SigningStargateClient.connectWithSigner(config.rpc, signer, {
    registry: createTxRegistry(),
  });
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
