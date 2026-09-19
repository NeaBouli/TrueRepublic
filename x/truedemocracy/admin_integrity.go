package truedemocracy

import (
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Read-only domain-admin integrity scan (GH-304). Legacy ElectAdmin builds
// persisted the elected member's bech32 text bytes as the raw admin address,
// which corrupts every admin-gated operation (member onboarding, permission
// register purges, treasury withdrawals) and makes exported genesis fail
// import validation. The detector below only reads state; it never writes.

// Stable reason codes reported by DetectDomainAdminCorruption.
const (
	// DomainAdminReasonEmpty: the stored admin address is empty.
	DomainAdminReasonEmpty = "admin_empty"
	// DomainAdminReasonMalformed: the stored admin bytes fail SDK address
	// format verification.
	DomainAdminReasonMalformed = "admin_malformed"
	// DomainAdminReasonBech32Text: the stored admin bytes are themselves
	// bech32 address text — the double-encoding signature of the GH-304
	// ElectAdmin corruption.
	DomainAdminReasonBech32Text = "admin_bytes_are_bech32_text"
	// DomainAdminReasonNotMember: the stored admin is not present in the
	// domain's canonical member set.
	DomainAdminReasonNotMember = "admin_not_in_members"
)

// DomainAdminFinding describes one domain whose stored admin failed the
// read-only integrity scan. All fields are stable, secret-free operator
// context.
type DomainAdminFinding struct {
	DomainName  string `json:"domain_name"`
	Reason      string `json:"reason"`
	AdminHex    string `json:"admin_hex"`    // raw stored admin bytes, hex-encoded
	AdminBech32 string `json:"admin_bech32"` // bech32 rendering of the stored bytes ("" when empty)
	MemberCount int    `json:"member_count"`
}

// DetectDomainAdminCorruption scans every domain and reports stored admins
// that are empty, malformed, double-encoded bech32 text (GH-304), or absent
// from the canonical member set. It is strictly read-only and deterministic:
// findings follow ascending domain-key iteration order.
func (k Keeper) DetectDomainAdminCorruption(ctx sdk.Context) []DomainAdminFinding {
	var findings []DomainAdminFinding
	k.IterateDomains(ctx, func(domain Domain) bool {
		if reason := classifyDomainAdmin(domain); reason != "" {
			findings = append(findings, DomainAdminFinding{
				DomainName:  domain.Name,
				Reason:      reason,
				AdminHex:    hex.EncodeToString(domain.Admin),
				AdminBech32: domain.Admin.String(),
				MemberCount: len(domain.Members),
			})
		}
		return false
	})
	return findings
}

// classifyDomainAdmin returns the stable reason code for a corrupted stored
// admin, or "" when the admin is intact.
func classifyDomainAdmin(domain Domain) string {
	switch {
	case domain.Admin.Empty():
		return DomainAdminReasonEmpty
	case sdk.VerifyAddressFormat(domain.Admin) != nil:
		return DomainAdminReasonMalformed
	case isMember(domain, domain.Admin.String()):
		return ""
	}
	// The admin is outside the member set. Distinguish the GH-304 signature —
	// raw bytes that are themselves bech32 address text — from other drift.
	if _, err := sdk.AccAddressFromBech32(string(domain.Admin)); err == nil {
		return DomainAdminReasonBech32Text
	}
	return DomainAdminReasonNotMember
}
