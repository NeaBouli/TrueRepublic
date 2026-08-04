package deploymentevidence

import (
	"fmt"
	"regexp"
	"time"
)

var (
	seatPattern      = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,62}[a-z0-9])?$`)
	hexDigestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
	timestampPattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)
)

// Verify checks a manifest against the derived facts of the exact local
// topology contract at an explicit evaluation time. It is deterministic:
// the same manifest, topology, and evaluation time always yield the same
// report. All violation messages are fixed and generic.
func Verify(manifest Manifest, topology Topology, evaluation time.Time) Report {
	v := verifier{evaluation: evaluation.UTC().Unix()}
	v.checkHeader(manifest, topology)
	preparedAt, preparedOK := v.checkPrepared(manifest)
	v.checkGates(manifest, preparedAt, preparedOK)
	v.checkApprovals(manifest, topology, preparedAt, preparedOK)

	violations := make([]Violation, len(v.violations))
	copy(violations, v.violations)
	return Report{
		Version:    ManifestVersion,
		GateCount:  len(GateIDs),
		NodeCount:  topology.NodeCount,
		Valid:      len(violations) == 0,
		Violations: violations,
	}
}

type verifier struct {
	violations []Violation
	evaluation int64 // evaluation time as Unix seconds; second arithmetic is overflow-safe
}

func (v *verifier) fail(check, message string) {
	v.violations = append(v.violations, Violation{Check: check, Message: message})
}

func (v *verifier) checkHeader(manifest Manifest, topology Topology) {
	if manifest.Version != ManifestVersion {
		v.fail("version", "manifest version is not supported")
	}
	if !hexDigestPattern.MatchString(manifest.TopologySHA256) {
		v.fail("topology_sha256.format", "topology digest must be 64 lowercase hexadecimal characters")
	} else if manifest.TopologySHA256 != topology.SHA256 {
		v.fail("topology_sha256.binding", "topology digest does not match the supplied topology contract")
	}
	if manifest.ChainID != topology.ChainID {
		v.fail("chain_id", "manifest chain does not match the topology contract")
	}
	if manifest.NodeCount != topology.NodeCount {
		v.fail("node_count", "node count does not match the topology contract")
	}
	if manifest.RoleCounts != topology.RoleCounts {
		v.fail("role_counts", "role counts do not match the topology contract")
	}
	if !seatPattern.MatchString(manifest.PreparedBy) {
		v.fail("prepared_by", "prepared_by must be a canonical logical seat")
	}
}

// checkPrepared validates the manifest preparation timestamp and returns its
// Unix seconds so gate and approval ordering can be checked against it.
func (v *verifier) checkPrepared(manifest Manifest) (int64, bool) {
	preparedAt, ok := parseStrictTimestamp(manifest.PreparedAt)
	if !ok {
		v.fail("prepared_at.format", "prepared_at must be a strict UTC timestamp")
		return 0, false
	}
	v.checkFresh("prepared_at", preparedAt)
	return preparedAt, true
}

func (v *verifier) checkGates(manifest Manifest, preparedAt int64, preparedOK bool) {
	if len(manifest.Gates) != len(GateIDs) {
		v.fail("gates.count", "manifest must declare exactly the canonical gate set")
		return
	}
	seenDigests := make(map[string]bool)
	for i, gate := range manifest.Gates {
		prefix := fmt.Sprintf("gates[%d]", i)
		if gate.ID != GateIDs[i] {
			v.fail(prefix+".id", "gates must appear exactly once in canonical order")
		}
		if gate.Result != GateResultPassed {
			v.fail(prefix+".result", "gate result must be passed")
		}
		startedAt, startedOK := parseStrictTimestamp(gate.StartedAt)
		if !startedOK {
			v.fail(prefix+".started_at", "gate started_at must be a strict UTC timestamp")
		}
		completedAt, completedOK := parseStrictTimestamp(gate.CompletedAt)
		if !completedOK {
			v.fail(prefix+".completed_at", "gate completed_at must be a strict UTC timestamp")
		}
		if startedOK && completedOK && preparedOK {
			if startedAt > completedAt || completedAt > preparedAt {
				v.fail(prefix+".order", "gate times must satisfy started_at <= completed_at <= prepared_at")
			}
		}
		if completedOK {
			v.checkFresh(prefix+".completed_at", completedAt)
		}
		if !hexDigestPattern.MatchString(gate.EvidenceSHA256) {
			v.fail(prefix+".evidence_sha256", "gate evidence digest must be 64 lowercase hexadecimal characters")
		} else if seenDigests[gate.EvidenceSHA256] {
			v.fail(prefix+".evidence_sha256", "gate evidence digest must be unique")
		} else {
			seenDigests[gate.EvidenceSHA256] = true
		}
	}
}

func (v *verifier) checkApprovals(manifest Manifest, topology Topology, preparedAt int64, preparedOK bool) {
	if len(manifest.Approvals) != 2 {
		v.fail("approvals.count", "manifest requires exactly two approvals")
		return
	}
	seats := make(map[string]bool)
	roles := make(map[string]bool)
	for i, approval := range manifest.Approvals {
		prefix := fmt.Sprintf("approvals[%d]", i)
		if !seatPattern.MatchString(approval.Seat) {
			v.fail(prefix+".seat", "approval seat must be a canonical logical seat")
		} else if approval.Seat == manifest.PreparedBy {
			v.fail(prefix+".seat", "approval seat must differ from the preparer seat")
		} else if seats[approval.Seat] {
			v.fail(prefix+".seat", "approval seats must be distinct")
		} else {
			seats[approval.Seat] = true
		}
		if approval.Role != ApprovalRoleOperator && approval.Role != ApprovalRoleIndependentReviewer {
			v.fail(prefix+".role", "approval role must be operator or independent-reviewer")
		} else if roles[approval.Role] {
			v.fail(prefix+".role", "approvals must pair one operator with one independent reviewer")
		} else {
			roles[approval.Role] = true
		}
		approvedAt, approvedOK := parseStrictTimestamp(approval.ApprovedAt)
		if !approvedOK {
			v.fail(prefix+".approved_at", "approval approved_at must be a strict UTC timestamp")
		} else {
			if preparedOK && approvedAt < preparedAt {
				v.fail(prefix+".approved_at", "approval must not precede manifest preparation")
			}
			v.checkFresh(prefix+".approved_at", approvedAt)
		}
		if !hexDigestPattern.MatchString(approval.TopologySHA256) {
			v.fail(prefix+".topology_sha256", "approval topology digest must be 64 lowercase hexadecimal characters")
		} else if approval.TopologySHA256 != topology.SHA256 {
			v.fail(prefix+".topology_sha256", "approval must be bound to the topology digest")
		}
	}
}

// checkFresh rejects evidence older than the maximum age or beyond the
// allowed future clock skew. Comparisons use Unix seconds, which stay far
// inside int64 range for every representable time.Time.
func (v *verifier) checkFresh(check string, at int64) {
	if at < v.evaluation-maxEvidenceAgeSec {
		v.fail(check, "evidence exceeds the maximum age of 30 days")
	}
	if at > v.evaluation+maxClockSkewSec {
		v.fail(check, "evidence is beyond the allowed clock skew of 5 minutes")
	}
}

// parseStrictTimestamp accepts only the canonical UTC form
// YYYY-MM-DDTHH:MM:SSZ and round-trips the parse to reject non-canonical or
// out-of-range components.
func parseStrictTimestamp(value string) (int64, bool) {
	if !timestampPattern.MatchString(value) {
		return 0, false
	}
	parsed, err := time.Parse(strictTimestampLayout, value)
	if err != nil || parsed.Format(strictTimestampLayout) != value {
		return 0, false
	}
	return parsed.Unix(), true
}
