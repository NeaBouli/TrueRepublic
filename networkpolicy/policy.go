// Package networkpolicy implements a deterministic, fail-closed validation
// boundary for TrueRepublic node homes. It inspects the effective CometBFT
// config.toml and Cosmos app.toml values of an initialized home and judges
// them against an explicit operator role profile (seed, sentry, validator,
// rpc, private). Validation is read-only: it never mutates configuration and
// never prints secret material.
package networkpolicy

import (
	"fmt"
	"net"
	"strings"

	cmtcfg "github.com/cometbft/cometbft/config"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
)

// Role is an explicit node role profile.
type Role string

const (
	RoleSeed      Role = "seed"
	RoleSentry    Role = "sentry"
	RoleValidator Role = "validator"
	RoleRPC       Role = "rpc"
	RolePrivate   Role = "private"
)

// ParseRole parses a CLI role string.
func ParseRole(s string) (Role, error) {
	switch Role(strings.ToLower(strings.TrimSpace(s))) {
	case RoleSeed:
		return RoleSeed, nil
	case RoleSentry:
		return RoleSentry, nil
	case RoleValidator:
		return RoleValidator, nil
	case RoleRPC:
		return RoleRPC, nil
	case RolePrivate:
		return RolePrivate, nil
	}
	return "", fmt.Errorf("unknown role %q: expected one of seed, sentry, validator, rpc, private", s)
}

// Options tunes validation. The zero value is the strict production policy.
type Options struct {
	// AllowLocalPeers permits loopback peer addresses exclusively for direct
	// library use by repository process tests. The operator CLI never exposes
	// this option, and unspecified addresses remain invalid.
	AllowLocalPeers bool
}

// Violation is one actionable policy failure. Check names the configuration
// key; Message explains the expectation. Violations never contain file
// contents or secret material.
type Violation struct {
	Check   string `json:"check"`
	Message string `json:"message"`
}

// Report is the complete validation outcome for one node home and role.
type Report struct {
	Role       Role        `json:"role"`
	Valid      bool        `json:"valid"`
	Violations []Violation `json:"violations"`
}

// Validate inspects the initialized node home and judges it against the role
// profile. It is deterministic and read-only, and it always fails closed:
// unreadable or unparseable configuration is a violation, never a skip.
func Validate(home string, role Role, opts Options) Report {
	report := Report{Role: role}
	v := &policyValidator{opts: opts}
	if _, err := ParseRole(string(role)); err != nil {
		v.fail("role", "role must be one of seed, sentry, validator, rpc, or private")
	}
	nc := loadHome(home)
	v.violations = append(v.violations, nc.loadErrs...)
	if nc.cmt != nil {
		v.checkCometRPC(nc.cmt.RPC)
		v.checkInstrumentation(nc.cmt.Instrumentation)
		v.checkPeerTopology(nc)
	}
	if nc.app != nil {
		v.checkAppListeners(nc.app)
	}
	if nc.cmt != nil && nc.app != nil {
		v.checkRoleProfile(nc, role)
	}
	report.Violations = v.violations
	report.Valid = len(v.violations) == 0
	return report
}

type policyValidator struct {
	opts       Options
	violations []Violation
}

func (v *policyValidator) fail(check, format string, args ...interface{}) {
	v.violations = append(v.violations, Violation{Check: check, Message: fmt.Sprintf(format, args...)})
}

// requireLoopbackListener enforces the shared client-facing boundary: a
// listener must bind a loopback address. Wildcard/unspecified binding of
// RPC, API, gRPC, gRPC-web, or pprof always fails; any other non-loopback
// bind also fails because public client traffic is expected behind a
// reverse proxy on the same host.
func (v *policyValidator) requireLoopbackListener(check, raw string) {
	_, port, class, err := parseListenerAddress(raw)
	if err != nil {
		v.fail(check, "invalid listener address: %v", err)
		return
	}
	switch class {
	case hostUnspecified:
		v.fail(check, "refuses wildcard/unspecified bind address; bind tcp://127.0.0.1:%d and expose public traffic through a reverse proxy", port)
	case hostLoopback:
	default:
		v.fail(check, "binds a non-loopback host directly; bind a loopback address and expose public traffic through a reverse proxy")
	}
}

// checkCometRPC enforces the CometBFT RPC boundary shared by every role.
func (v *policyValidator) checkCometRPC(rpc *cmtcfg.RPCConfig) {
	v.requireLoopbackListener("rpc.laddr", rpc.ListenAddress)
	if rpc.Unsafe {
		v.fail("rpc.unsafe", "unsafe RPC commands must stay disabled on every role")
	}
	for _, origin := range rpc.CORSAllowedOrigins {
		if strings.TrimSpace(origin) == "*" {
			v.fail("rpc.cors_allowed_origins", "wildcard CORS origin \"*\" is never allowed; list explicit origins")
		}
	}
	if rpc.PprofListenAddress != "" {
		v.requireLoopbackListener("rpc.pprof_laddr", rpc.PprofListenAddress)
	}
}

func (v *policyValidator) checkInstrumentation(instrumentation *cmtcfg.InstrumentationConfig) {
	if instrumentation.Prometheus {
		v.requireLoopbackListener("instrumentation.prometheus_listen_addr", instrumentation.PrometheusListenAddr)
	}
}

// checkAppListeners enforces the Cosmos app.toml API/gRPC/gRPC-web boundary
// shared by every role: any enabled client listener is loopback-only.
func (v *policyValidator) checkAppListeners(app *serverconfig.Config) {
	if app.API.Enable {
		v.requireLoopbackListener("api.address", app.API.Address)
	}
	if app.API.EnableUnsafeCORS {
		v.fail("api.enabled-unsafe-cors", "unsafe API CORS must stay disabled on every role")
	}
	if app.GRPC.Enable {
		v.requireLoopbackListener("grpc.address", app.GRPC.Address)
	}
	// In SDK v0.50 gRPC-web is served through the API listener, so an
	// enabled gRPC-web inherits the api.address loopback check above; the
	// role profiles additionally require it explicitly disabled where
	// least privilege demands it.
}

// checkPeerTopology canonicalizes and validates every configured CometBFT
// peer endpoint and peer ID list. It rejects malformed endpoints, duplicate
// or self-conflicting node IDs, and — unless test-only local peers are
// explicitly allowed — loopback or wildcard peer addresses.
func (v *policyValidator) checkPeerTopology(nc *nodeConfig) {
	p2pCfg := nc.cmt.P2P

	seen := make(map[string]struct{})
	collect := func(check, list string) {
		for i, raw := range splitCSV(list) {
			ep, err := canonicalPeer(raw)
			if err != nil {
				v.fail(check, "entry %d is not a canonical peer endpoint: %v", i, err)
				continue
			}
			if _, dup := seen[ep.ID]; dup {
				v.fail(check, "entry %d repeats or conflicts with an earlier peer identity", i)
				continue
			}
			seen[ep.ID] = struct{}{}
			if nc.hasSelf && ep.ID == nc.selfID {
				v.fail(check, "entry %d lists this node's own node ID; a node must never peer with itself", i)
			}
			switch classifyHost(ep.Host) {
			case hostUnspecified:
				v.fail(check, "entry %d uses a wildcard/unspecified peer address; peers must have explicit routable addresses", i)
			case hostLoopback:
				if !v.opts.AllowLocalPeers {
					v.fail(check, "entry %d uses a loopback peer address; loopback peers are test-only and never valid production topology", i)
				}
			}
		}
	}
	collect("p2p.persistent_peers", p2pCfg.PersistentPeers)
	collect("p2p.seeds", p2pCfg.Seeds)

	checkIDs := func(check, list string) {
		seenIDs := make(map[string]struct{})
		for i, id := range splitCSV(list) {
			if err := validateNodeID(id); err != nil {
				v.fail(check, "entry %d: %v", i, err)
				continue
			}
			if _, dup := seenIDs[id]; dup {
				v.fail(check, "entry %d repeats an earlier node ID", i)
				continue
			}
			seenIDs[id] = struct{}{}
			if nc.hasSelf && id == nc.selfID {
				v.fail(check, "entry %d lists this node's own node ID", i)
			}
		}
	}
	checkIDs("p2p.private_peer_ids", p2pCfg.PrivatePeerIDs)
	checkIDs("p2p.unconditional_peer_ids", p2pCfg.UnconditionalPeerIDs)

	if ext := strings.TrimSpace(p2pCfg.ExternalAddress); ext != "" {
		_, _, class, err := parseListenerAddress(ext)
		if err != nil {
			v.fail("p2p.external_address", "invalid external address: %v", err)
		} else if class == hostUnspecified {
			v.fail("p2p.external_address", "external address must not be a wildcard/unspecified address")
		} else if class == hostLoopback && !v.opts.AllowLocalPeers {
			v.fail("p2p.external_address", "external address must not be loopback outside repository process tests")
		}
	}
}

// splitCSV splits a CometBFT comma-separated list, dropping empty entries.
func splitCSV(list string) []string {
	var out []string
	for _, item := range strings.Split(list, ",") {
		if item = strings.TrimSpace(item); item != "" {
			out = append(out, item)
		}
	}
	return out
}

// hasPersistentPeers reports whether at least one persistent peer is
// configured.
func hasPersistentPeers(nc *nodeConfig) bool {
	return len(splitCSV(nc.cmt.P2P.PersistentPeers)) > 0
}

func persistentPeerIDs(nc *nodeConfig) map[string]struct{} {
	ids := make(map[string]struct{})
	for _, raw := range splitCSV(nc.cmt.P2P.PersistentPeers) {
		if peer, err := canonicalPeer(raw); err == nil {
			ids[peer.ID] = struct{}{}
		}
	}
	return ids
}

// checkRoleProfile enforces the role-specific least-privilege topology.
func (v *policyValidator) checkRoleProfile(nc *nodeConfig, role Role) {
	p2pCfg := nc.cmt.P2P
	app := nc.app
	peerIDs := persistentPeerIDs(nc)
	persistentPeerCount := len(splitCSV(p2pCfg.PersistentPeers))

	requireDisabled := func(check string, enabled bool, what string) {
		if enabled {
			v.fail(check, "%s role must keep %s disabled; only rpc and private roles may expose local client services", role, what)
		}
	}
	requirePersistentIDs := func(check, list, purpose string) {
		for _, id := range splitCSV(list) {
			if validateNodeID(id) != nil {
				continue
			}
			if _, ok := peerIDs[id]; !ok {
				v.fail(check, "entry must also be an explicit persistent peer (%s)", purpose)
			}
		}
	}
	requireIDsInList := func(check, required, allowed, purpose string) {
		allowedIDs := make(map[string]struct{})
		for _, id := range splitCSV(allowed) {
			if validateNodeID(id) == nil {
				allowedIDs[id] = struct{}{}
			}
		}
		for _, id := range splitCSV(required) {
			if validateNodeID(id) != nil {
				continue
			}
			if _, ok := allowedIDs[id]; !ok {
				v.fail(check, "entry must also appear in %s", purpose)
			}
		}
	}
	requirePrivateP2P := func() {
		_, _, class, err := parseListenerAddress(p2pCfg.ListenAddress)
		if err != nil {
			v.fail("p2p.laddr", "invalid P2P listener address: %v", err)
		} else if class != hostLoopback {
			v.fail("p2p.laddr", "%s role must bind P2P to loopback and reach explicit peers outbound", role)
		}
		if strings.TrimSpace(p2pCfg.ExternalAddress) != "" {
			v.fail("p2p.external_address", "%s role must not advertise an external P2P address", role)
		}
	}
	requirePublicP2P := func() {
		host, _, class, err := parseListenerAddress(p2pCfg.ListenAddress)
		if err != nil {
			v.fail("p2p.laddr", "invalid P2P listener address: %v", err)
		} else if class != hostRoutable || net.ParseIP(host) == nil {
			v.fail("p2p.laddr", "%s role must bind P2P to one explicit non-loopback IP interface; wildcard, loopback, and DNS binds are rejected", role)
		}
		if strings.TrimSpace(p2pCfg.ExternalAddress) == "" {
			v.fail("p2p.external_address", "%s role requires an explicit advertised P2P address", role)
		}
	}

	switch role {
	case RoleValidator:
		// Validator identity protection: hide behind sentry peers only.
		requirePrivateP2P()
		if p2pCfg.SeedMode {
			v.fail("p2p.seed_mode", "validator must not run in seed mode")
		}
		if p2pCfg.PexReactor {
			v.fail("p2p.pex", "validator must disable peer exchange; it may only talk to its sentries")
		}
		if len(splitCSV(p2pCfg.Seeds)) != 0 {
			v.fail("p2p.seeds", "validator must not configure discovery seeds; it dials only reviewed sentry persistent peers")
		}
		if persistentPeerCount < 2 {
			v.fail("p2p.persistent_peers", "validator requires at least two explicit sentry persistent peers; direct public peering is forbidden")
		}
		if len(splitCSV(p2pCfg.UnconditionalPeerIDs)) == 0 {
			v.fail("p2p.unconditional_peer_ids", "validator must pin its sentry node IDs in unconditional_peer_ids so sentry links survive peer limits")
		}
		requirePersistentIDs("p2p.unconditional_peer_ids", p2pCfg.UnconditionalPeerIDs, "validator sentry links")
		if p2pCfg.MaxNumInboundPeers != 0 {
			v.fail("p2p.max_num_inbound_peers", "validator must accept zero inbound peers; sentries are dialed outbound only")
		}
		requireDisabled("api.enable", app.API.Enable, "the API server")
		requireDisabled("grpc-web.enable", app.GRPCWeb.Enable, "gRPC-web")

	case RoleSentry:
		requirePublicP2P()
		if p2pCfg.SeedMode {
			v.fail("p2p.seed_mode", "sentry must not run in seed mode; use the seed role for a dedicated seed")
		}
		if !p2pCfg.PexReactor {
			v.fail("p2p.pex", "sentry must enable peer exchange to shield its validator from the public mesh")
		}
		if !hasPersistentPeers(nc) {
			v.fail("p2p.persistent_peers", "sentry requires explicit persistent peers to its reviewed upstream sentry/seed set")
		}
		if len(splitCSV(p2pCfg.PrivatePeerIDs)) == 0 {
			v.fail("p2p.private_peer_ids", "sentry must list its protected validator node ID in private_peer_ids")
		}
		if len(splitCSV(p2pCfg.UnconditionalPeerIDs)) == 0 {
			v.fail("p2p.unconditional_peer_ids", "sentry must pin its protected validator in unconditional_peer_ids")
		}
		requireIDsInList("p2p.private_peer_ids", p2pCfg.PrivatePeerIDs, p2pCfg.UnconditionalPeerIDs, "unconditional_peer_ids")
		if p2pCfg.MaxNumInboundPeers < 1 {
			v.fail("p2p.max_num_inbound_peers", "sentry must accept inbound peers; a sentry with zero inbound capacity is useless")
		}
		if p2pCfg.MaxNumOutboundPeers < 1 {
			v.fail("p2p.max_num_outbound_peers", "sentry must allow outbound peers")
		}
		requireDisabled("api.enable", app.API.Enable, "the API server")
		requireDisabled("grpc-web.enable", app.GRPCWeb.Enable, "gRPC-web")

	case RoleSeed:
		requirePublicP2P()
		if !p2pCfg.SeedMode {
			v.fail("p2p.seed_mode", "seed role requires p2p.seed_mode = true")
		}
		if !p2pCfg.PexReactor {
			v.fail("p2p.pex", "seed must enable peer exchange; crawling and gossiping the address book is its purpose")
		}
		if p2pCfg.MaxNumInboundPeers < 1 {
			v.fail("p2p.max_num_inbound_peers", "seed must accept inbound peers")
		}
		if p2pCfg.MaxNumOutboundPeers < 1 {
			v.fail("p2p.max_num_outbound_peers", "seed must allow outbound peers to crawl the mesh")
		}
		if len(splitCSV(p2pCfg.PrivatePeerIDs)) != 0 {
			v.fail("p2p.private_peer_ids", "seed must not hide any peer ID; a seed gossips the full address book")
		}
		requireDisabled("api.enable", app.API.Enable, "the API server")
		requireDisabled("grpc-web.enable", app.GRPCWeb.Enable, "gRPC-web")

	case RoleRPC:
		requirePublicP2P()
		if p2pCfg.SeedMode {
			v.fail("p2p.seed_mode", "rpc role must not run in seed mode")
		}
		if !p2pCfg.PexReactor {
			v.fail("p2p.pex", "rpc role must enable peer exchange; use the private role for an explicit-peer-only node")
		}
		if !hasPersistentPeers(nc) {
			v.fail("p2p.persistent_peers", "rpc role requires explicit persistent peers to a trusted sentry/seed set")
		}
		if p2pCfg.MaxNumInboundPeers < 1 {
			v.fail("p2p.max_num_inbound_peers", "rpc role must accept inbound peers to stay synced with the mesh")
		}
		if p2pCfg.MaxNumOutboundPeers < 1 {
			v.fail("p2p.max_num_outbound_peers", "rpc role must allow outbound peers")
		}
		// API, gRPC, and gRPC-web may be enabled here, but the shared
		// listener checks already forced them onto loopback behind a
		// reverse proxy.

	case RolePrivate:
		requirePrivateP2P()
		if p2pCfg.SeedMode {
			v.fail("p2p.seed_mode", "private role must not run in seed mode")
		}
		if p2pCfg.PexReactor {
			v.fail("p2p.pex", "private role must disable peer exchange and dial only its explicit peers")
		}
		if len(splitCSV(p2pCfg.Seeds)) != 0 {
			v.fail("p2p.seeds", "private role must not configure discovery seeds")
		}
		if !hasPersistentPeers(nc) {
			v.fail("p2p.persistent_peers", "private role requires explicit persistent peers; it never joins the public mesh")
		}
		if p2pCfg.MaxNumInboundPeers != 0 {
			v.fail("p2p.max_num_inbound_peers", "private role must accept zero inbound peers")
		}
	}
}
