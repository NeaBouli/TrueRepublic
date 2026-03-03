export interface DomainInvite {
  domainId: string;
  domainName: string;
  inviter: string;
  signature?: string;
  expiresAt?: number;
}

export interface MembershipStatus {
  domainId: string;
  address: string;
  isMember: boolean;
  hasIdentityCommitment: boolean;
  inMerkleTree: boolean;
  /** Step 1: MsgOnboardToDomain submitted (domain key registration) */
  step1Complete: boolean;
  /** Step 2: MsgApproveOnboarding by admin */
  step2Complete: boolean;
  joinedAt?: string;
}

export interface OnboardingStep {
  step: 1 | 2;
  status: 'pending' | 'complete';
  timestamp?: string;
}

/**
 * Parameters for MsgOnboardToDomain (Go: onboard_to_domain).
 * Two-step onboarding: user submits domain key pair, admin approves.
 */
export interface OnboardParams {
  domain_name: string;
  domain_pub_key_hex: string;
  global_pub_key_hex: string;
  signature_hex: string;
}

/**
 * Parameters for MsgRegisterIdentity (Go: register_identity).
 * Registers a MiMC identity commitment for ZKP anonymous voting.
 * Requires the user to already be a domain member.
 */
export interface RegisterIdentityParams {
  domain_name: string;
  commitment: string; // 64 hex chars (32 bytes MiMC hash)
}
