import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { describe, expect, it } from 'vitest';
import {
  bytesToHex,
  computeVoteNullifierScope,
  computeVoteSignal,
  computeVoteSignalV2,
  hexToBytes,
  mimcBn254,
} from './zkpEncoding';

interface GoldenVector {
  chain_id: string;
  domain_name: string;
  issue_name: string;
  suggestion_name: string;
  rating: number;
  synthetic_witness_hex: string;
  commitment_hex: string;
  external_nullifier_hex: string;
  signal_hash_hex: string;
  nullifier_hash_hex: string;
  synthetic_and_test_only: boolean;
}

interface CircuitSpecV2Vector {
  chain_id: string;
  domain_name: string;
  issue_name: string;
  suggestion_name: string;
  rating: number;
  reward_recipient: string;
  signal_hash_hex: string;
}

interface CircuitSpec {
  schema: string;
  vote_context_v2: {
    domain_separator: string;
    vector: CircuitSpecV2Vector;
  };
}

const vector = JSON.parse(
  readFileSync(
    resolve(
      process.cwd(),
      '../x/truedemocracy/testdata/zkp/golden_vector.json'
    ),
    'utf8'
  )
) as GoldenVector;

const spec = JSON.parse(
  readFileSync(
    resolve(process.cwd(), '../configs/security/zkp-circuit.json'),
    'utf8'
  )
) as CircuitSpec;

describe('ZKP Go/client encoding compatibility', () => {
  it('matches the synthetic Go MiMC commitment and nullifier vector', () => {
    const emptyElement = bytesToHex(mimcBn254([new Uint8Array()]));
    const explicitZero = bytesToHex(mimcBn254([new Uint8Array(32)]));

    expect(emptyElement).toBe(explicitZero);
    expect(emptyElement).toHaveLength(64);
    expect(vector.synthetic_and_test_only).toBe(true);
    const secret = hexToBytes(vector.synthetic_witness_hex);
    const external = hexToBytes(vector.external_nullifier_hex);

    expect(bytesToHex(mimcBn254([secret]))).toBe(vector.commitment_hex);
    expect(bytesToHex(mimcBn254([secret, external]))).toBe(
      vector.nullifier_hash_hex
    );
  });

  it('matches Go chain-scoped context and exact-rating signal encoding', () => {
    const scope = computeVoteNullifierScope(
      vector.chain_id,
      vector.domain_name,
      vector.issue_name,
      vector.suggestion_name
    );
    const signal = computeVoteSignal(
      vector.chain_id,
      vector.domain_name,
      vector.issue_name,
      vector.suggestion_name,
      vector.rating
    );

    expect(bytesToHex(scope)).toBe(vector.external_nullifier_hex);
    expect(bytesToHex(signal)).toBe(vector.signal_hash_hex);
    expect(
      bytesToHex(
        computeVoteSignal(
          vector.chain_id,
          vector.domain_name,
          vector.issue_name,
          vector.suggestion_name,
          vector.rating + 1
        )
      )
    ).not.toBe(vector.signal_hash_hex);
  });

  it('matches the GH-209 recipient-bound v2 signal vector', () => {
    expect(spec.schema).toBe('truerepublic/zkp-circuit/v2');
    expect(spec.vote_context_v2.domain_separator).toBe('TrueRepublic/vote/v2');
    const v2 = spec.vote_context_v2.vector;
    const signal = computeVoteSignalV2(
      v2.chain_id,
      v2.domain_name,
      v2.issue_name,
      v2.suggestion_name,
      v2.rating,
      v2.reward_recipient
    );
    expect(bytesToHex(signal)).toBe(v2.signal_hash_hex);

    // The v2 signal binds the recipient and never collides with the v1 or
    // recipient-shifted encodings.
    expect(
      bytesToHex(
        computeVoteSignalV2(
          v2.chain_id,
          v2.domain_name,
          v2.issue_name,
          v2.suggestion_name,
          v2.rating,
          `${v2.reward_recipient}x`
        )
      )
    ).not.toBe(v2.signal_hash_hex);
    expect(
      bytesToHex(
        computeVoteSignal(
          v2.chain_id,
          v2.domain_name,
          v2.issue_name,
          v2.suggestion_name,
          v2.rating
        )
      )
    ).not.toBe(v2.signal_hash_hex);
  });
});
