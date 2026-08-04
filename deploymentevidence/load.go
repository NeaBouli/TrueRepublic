package deploymentevidence

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"truerepublic/networkpolicy"
	"truerepublic/topologypolicy"
)

// LoadTopology reads the exact raw bytes of a local GH-89 topology contract,
// computes the SHA-256 of those exact bytes, parses and validates the
// contract through topologypolicy, and derives the secret-free facts a
// manifest must match. All errors are generic and fixed: topology parser or
// validator messages are never reflected.
func LoadTopology(path string) (topology Topology, err error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Topology{}, fmt.Errorf("open topology contract: file does not exist")
		}
		return Topology{}, fmt.Errorf("open topology contract: file is unavailable")
	}
	defer func() {
		if closeErr := file.Close(); err == nil && closeErr != nil {
			topology = Topology{}
			err = fmt.Errorf("close topology contract: file is unavailable")
		}
	}()

	raw, err := io.ReadAll(io.LimitReader(file, topologypolicy.MaxContractBytes+1))
	if err != nil {
		return Topology{}, fmt.Errorf("read topology contract: input is unavailable")
	}
	if len(raw) > topologypolicy.MaxContractBytes {
		return Topology{}, fmt.Errorf("topology contract exceeds %d bytes", topologypolicy.MaxContractBytes)
	}
	if len(bytes.TrimSpace(raw)) == 0 {
		return Topology{}, fmt.Errorf("topology contract is empty")
	}

	sum := sha256.Sum256(raw)

	contract, err := topologypolicy.Parse(bytes.NewReader(raw))
	if err != nil {
		return Topology{}, fmt.Errorf("invalid topology contract")
	}
	if report := topologypolicy.Validate(contract); !report.Valid {
		return Topology{}, fmt.Errorf("topology contract validation failed")
	}

	topology = Topology{
		SHA256:    hex.EncodeToString(sum[:]),
		ChainID:   contract.ChainID,
		NodeCount: len(contract.Nodes),
	}
	for i := range contract.Nodes {
		switch contract.Nodes[i].Role {
		case networkpolicy.RoleSeed:
			topology.RoleCounts.Seed++
		case networkpolicy.RoleSentry:
			topology.RoleCounts.Sentry++
		case networkpolicy.RoleValidator:
			topology.RoleCounts.Validator++
		case networkpolicy.RoleRPC:
			topology.RoleCounts.RPC++
		}
	}
	return topology, nil
}
