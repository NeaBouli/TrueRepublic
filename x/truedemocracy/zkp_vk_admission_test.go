package truedemocracy

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"truerepublic/x/truedemocracy/zkpadmission"
)

// GH300B4B: chain genesis validation still admits any canonical circuit-shaped
// VK (consensus unchanged), while the offline admission policy denies the frozen
// TEST-ONLY key and leaves a fresh single-party setup key unadmitted.
func TestGenesisVerifyingKeysClassifyForOfflineAdmission(t *testing.T) {
	policy := zkpadmission.DefaultPolicy()
	if zkpadmission.CircuitID != MembershipCircuitID {
		t.Fatalf("admission circuit %q != %q", zkpadmission.CircuitID, MembershipCircuitID)
	}

	frozen, err := os.ReadFile(filepath.Join(zkpFixtureDir, "membership_v2.vk"))
	if err != nil {
		t.Fatal(err)
	}
	fresh, err := SerializeVerifyingKey(getTestZKPKeys(t).VerifyingKey)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		key  []byte
		want zkpadmission.Class
	}{
		{"frozen test-only fixture", frozen, zkpadmission.TestOnlyDenied},
		{"fresh single-party setup", fresh, zkpadmission.Unadmitted},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fingerprint := VerifyingKeyFingerprint(tc.key)
			if _, err := ValidateMembershipVerifyingKey(tc.key, MembershipCircuitID, fingerprint); err != nil {
				t.Fatalf("consensus validation unexpectedly changed: %v", err)
			}
			class, err := policy.Classify(MembershipCircuitID, hex.EncodeToString(tc.key), fingerprint)
			if err != nil || class != tc.want {
				t.Fatalf("Classify = (%q, %v), want %q", class, err, tc.want)
			}
		})
	}
}
