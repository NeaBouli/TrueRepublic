/**
 * ZKP types for anonymous voting (Groth16 on BN254/MiMC).
 * Matches Go x/truedemocracy/zkp.go + merkle.go + anonymity.go.
 */

/** Merkle tree depth matching Go MerkleTreeDepth = 20 */
export const MERKLE_TREE_DEPTH = 20;

/** Rating range matching Go MsgRateWithProof validation (-5 to +5) */
export const RATING_MIN = -5;
export const RATING_MAX = 5;

/** Client-side ZKP identity (stored encrypted in localStorage) */
export interface Identity {
  secret: string;
  commitment: string;
  nullifier: string;
  createdAt: number;
}

/** Merkle proof for set membership (20 levels) */
export interface MerkleProof {
  root: string;
  pathIndices: number[];
  pathElements: string[];
  leaf: string;
}

/** Inputs for Groth16 proof generation */
export interface ProofInputs {
  chainId: string;
  identitySecret: string;
  merkleRoot: string;
  merkleProof: MerkleProof;
  externalNullifier: string;
  rating: number;
  domainName: string;
  issueName: string;
  suggestionName: string;
  /** GH-209: canonical bech32 payout address bound into the v2 signal */
  rewardRecipient: string;
}

/** Generated Groth16 proof ready for chain submission */
export interface GeneratedProof {
  proof: string;
  nullifierHash: string;
  merkleRoot: string;
  publicSignals: string[];
}

/** Test seam for replaying compatible proof fixtures without enabling UI submission. */
export interface Groth16Prover {
  generate(inputs: ProofInputs): Promise<GeneratedProof>;
}

/**
 * Params for MsgRateWithProof submission.
 * Field names match Go MsgRateWithProof json tags.
 */
export interface VoteWithProofParams {
  domain_name: string;
  issue_name: string;
  suggestion_name: string;
  rating: number;
  proof: string;
  nullifier_hash: string;
  merkle_root: string;
  /** GH-209: proof-bound canonical bech32 payout address */
  reward_recipient: string;
}

/** Proof generation progress tracking */
export interface ProofGenerationStatus {
  step:
    | 'idle'
    | 'loading_wasm'
    | 'fetching_proof'
    | 'generating'
    | 'complete'
    | 'error';
  progress: number;
  message: string;
  error?: string;
}
