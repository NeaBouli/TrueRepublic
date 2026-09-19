package truedemocracy

import (
	"bytes"
	"encoding/hex"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkaddress "github.com/cosmos/cosmos-sdk/types/address"
)

// TestDetectDomainAdminCorruption covers the read-only GH-304 detector:
// healthy domains are not reported, intentionally corrupted legacy admins are
// classified with stable reason codes, and the scan never mutates state.
func TestDetectDomainAdminCorruption(t *testing.T) {
	k, ctx := setupKeeper(t)
	addrs := realTestAddrs(4)

	// Healthy domain must not be reported.
	setupBech32ElectionDomain(t, k, ctx, "HealthyDomain", addrs)

	// GH-304-style legacy corruption: the stored admin bytes are the winner's
	// bech32 text instead of the decoded address.
	setupBech32ElectionDomain(t, k, ctx, "CorruptDomain", addrs)
	corrupt, _ := k.GetDomain(ctx, "CorruptDomain")
	corrupt.Admin = sdk.AccAddress([]byte(addrs[1].String()))
	saveDomain(t, k, ctx, corrupt)

	// Empty admin.
	setupBech32ElectionDomain(t, k, ctx, "EmptyAdminDomain", addrs)
	empty, _ := k.GetDomain(ctx, "EmptyAdminDomain")
	empty.Admin = nil
	saveDomain(t, k, ctx, empty)

	// Malformed admin: raw bytes exceed the SDK maximum address length.
	setupBech32ElectionDomain(t, k, ctx, "MalformedDomain", addrs)
	oversized := bytes.Repeat([]byte{0xab}, sdkaddress.MaxAddrLen+1)
	malformed, _ := k.GetDomain(ctx, "MalformedDomain")
	malformed.Admin = sdk.AccAddress(oversized)
	saveDomain(t, k, ctx, malformed)

	// Well-formed admin address that is not in the canonical member set.
	setupBech32ElectionDomain(t, k, ctx, "StrangerDomain", addrs)
	stranger, _ := k.GetDomain(ctx, "StrangerDomain")
	strangerBytes := bytes.Repeat([]byte{0xff}, 20)
	stranger.Admin = sdk.AccAddress(strangerBytes)
	saveDomain(t, k, ctx, stranger)

	snapshot := snapshotModuleStore(t, k, ctx)

	findings := k.DetectDomainAdminCorruption(ctx)

	// Deterministic order follows ascending domain-key iteration.
	want := []DomainAdminFinding{
		{
			DomainName:  "CorruptDomain",
			Reason:      DomainAdminReasonBech32Text,
			AdminHex:    hex.EncodeToString([]byte(addrs[1].String())),
			AdminBech32: sdk.AccAddress([]byte(addrs[1].String())).String(),
			MemberCount: 4,
		},
		{
			DomainName:  "EmptyAdminDomain",
			Reason:      DomainAdminReasonEmpty,
			AdminHex:    "",
			AdminBech32: "",
			MemberCount: 4,
		},
		{
			DomainName:  "MalformedDomain",
			Reason:      DomainAdminReasonMalformed,
			AdminHex:    hex.EncodeToString(oversized),
			AdminBech32: sdk.AccAddress(oversized).String(),
			MemberCount: 4,
		},
		{
			DomainName:  "StrangerDomain",
			Reason:      DomainAdminReasonNotMember,
			AdminHex:    hex.EncodeToString(strangerBytes),
			AdminBech32: sdk.AccAddress(strangerBytes).String(),
			MemberCount: 4,
		},
	}
	if len(findings) != len(want) {
		t.Fatalf("findings = %+v, want exactly %d", findings, len(want))
	}
	for i, w := range want {
		got := findings[i]
		if got.DomainName != w.DomainName || got.Reason != w.Reason ||
			got.AdminHex != w.AdminHex || got.AdminBech32 != w.AdminBech32 ||
			got.MemberCount != w.MemberCount {
			t.Errorf("finding %d = %+v, want %+v", i, got, w)
		}
	}

	// The scan is strictly read-only: the complete module KV store must be
	// byte-identical, not only the pre-existing domain values.
	assertModuleStoreUnchanged(t, k, ctx, snapshot)
}

// TestDetectDomainAdminCorruptionCleanAfterElection proves a valid elected
// admin is never flagged: no false positives after a correct real-bech32
// election.
func TestDetectDomainAdminCorruptionCleanAfterElection(t *testing.T) {
	k, ctx := setupKeeper(t)
	addrs := realTestAddrs(4)
	setupBech32ElectionDomain(t, k, ctx, "ElectDomain", addrs)
	electBobByBech32Stones(t, k, ctx, "ElectDomain", addrs[1], addrs[2], addrs[3])

	if err := k.ElectAdmin(ctx, "ElectDomain"); err != nil {
		t.Fatalf("ElectAdmin: %v", err)
	}

	if findings := k.DetectDomainAdminCorruption(ctx); len(findings) != 0 {
		t.Fatalf("valid elected admin must not be flagged: %+v", findings)
	}
}

// snapshotModuleStore captures every raw key/value pair in the module store —
// not only decodable domain values — so any detector write, including new,
// deleted, or non-domain keys, is detected.
func snapshotModuleStore(t *testing.T, k Keeper, ctx sdk.Context) map[string][]byte {
	t.Helper()
	store := ctx.KVStore(k.StoreKey)
	snapshot := map[string][]byte{}
	it := store.Iterator(nil, nil)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		snapshot[string(it.Key())] = append([]byte(nil), it.Value()...)
	}
	return snapshot
}

// assertModuleStoreUnchanged verifies the complete module store is still
// byte-identical to the snapshot: same keys, same values.
func assertModuleStoreUnchanged(t *testing.T, k Keeper, ctx sdk.Context, snapshot map[string][]byte) {
	t.Helper()
	after := snapshotModuleStore(t, k, ctx)
	if len(after) != len(snapshot) {
		t.Fatalf("module store key count = %d, want %d", len(after), len(snapshot))
	}
	for key, before := range snapshot {
		got, found := after[key]
		if !found {
			t.Fatalf("detector deleted store key %q", key)
		}
		if !bytes.Equal(got, before) {
			t.Fatalf("detector mutated store key %q", key)
		}
	}
}
