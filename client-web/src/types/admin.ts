/**
 * Admin types for domain management UI.
 * Field names match Go x/truedemocracy types where applicable.
 */

/** Domain member from Domain.Members array */
export interface DomainMember {
  address: string;
  /** Unknown because on-chain commitments are intentionally address-unbound. */
  hasIdentityCommitment: null;
  /** Unknown because on-chain permission keys are intentionally address-unbound. */
  inPermissionReg: null;
}

/** Domain statistics computed from domain query data */
export interface DomainStats {
  domainName: string;
  totalMembers: number;
  totalIssues: number;
  totalSuggestions: number;
  treasuryBalance: string;
  identityCommitments: number;
  permissionRegCount: number;
  merkleRoot: string;
}

/** Params for MsgCreateDomain (Go: create_domain) */
export interface CreateDomainParams {
  name: string;
  initial_coins: string; // amount in upnyx
}

/** Params for MsgApproveOnboarding (Go: approve_onboarding) */
export interface ApproveOnboardingParams {
  domain_name: string;
  requester_addr: string;
}

/** Params for MsgAddMember (Go: add_member) */
export interface AddMemberParams {
  domain_name: string;
  new_member: string;
}
