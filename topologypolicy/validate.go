package topologypolicy

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"truerepublic/networkpolicy"
)

var (
	logicalNamePattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,62}[a-z0-9])?$`)
	nodeIDPattern      = regexp.MustCompile(`^[0-9a-f]{40}$`)
	chainIDPattern     = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{2,63}$`)
)

type validator struct {
	violations []Violation
	nodes      map[string]*Node
	flows      map[string]map[string]bool
}

func Validate(contract Contract) Report {
	v := validator{
		nodes: make(map[string]*Node),
		flows: make(map[string]map[string]bool),
	}
	v.checkHeader(contract)
	v.checkDefaults(contract.Defaults)
	v.checkNodes(contract.Nodes)
	v.checkComposition(contract.Nodes)
	v.checkPeerGraph(contract.Nodes)
	v.checkProtectionGraph(contract.Nodes)
	v.checkIngress(contract.Ingress)
	v.checkFlows(contract)

	violations := make([]Violation, len(v.violations))
	copy(violations, v.violations)
	return Report{
		Version:    contract.Version,
		ChainID:    contract.ChainID,
		NodeCount:  len(contract.Nodes),
		Valid:      len(v.violations) == 0,
		Violations: violations,
	}
}

func (v *validator) fail(check, format string, args ...any) {
	v.violations = append(v.violations, Violation{
		Check:   check,
		Message: fmt.Sprintf(format, args...),
	})
}

func (v *validator) checkHeader(contract Contract) {
	if contract.Version != ContractVersion {
		v.fail("version", "version must be %q", ContractVersion)
	}
	if !chainIDPattern.MatchString(contract.ChainID) {
		v.fail("chain_id", "chain_id must be a canonical, non-secret logical identifier")
	}
}

func (v *validator) checkDefaults(defaults Defaults) {
	if defaults.Inbound != "deny" {
		v.fail("defaults.inbound", "default inbound policy must be deny")
	}
	if defaults.Outbound != "deny" {
		v.fail("defaults.outbound", "default outbound policy must be deny")
	}
}

func (v *validator) checkNodes(nodes []Node) {
	seenIDs := make(map[string]string)
	seenEndpoints := make(map[string]string)
	for i := range nodes {
		node := &nodes[i]
		prefix := fmt.Sprintf("nodes[%d]", i)
		if !logicalNamePattern.MatchString(node.Name) {
			v.fail(prefix+".name", "node name %q is not canonical", node.Name)
		} else if node.Name == "internet" {
			v.fail(prefix+".name", "node name %q is reserved for the external flow principal", node.Name)
		} else if _, exists := v.nodes[node.Name]; exists {
			v.fail(prefix+".name", "node name %q is duplicated", node.Name)
		} else {
			v.nodes[node.Name] = node
		}
		if !logicalNamePattern.MatchString(node.Zone) {
			v.fail(prefix+".zone", "zone %q is not canonical", node.Zone)
		}
		if !nodeIDPattern.MatchString(node.NodeID) {
			v.fail(prefix+".node_id", "node ID must contain exactly 40 lowercase hexadecimal characters")
		} else if owner, exists := seenIDs[node.NodeID]; exists {
			v.fail(prefix+".node_id", "node ID is already assigned to %q", owner)
		} else {
			seenIDs[node.NodeID] = node.Name
		}
		if !allowedRole(node.Role) {
			v.fail(prefix+".role", "role %q is not supported by the topology contract", node.Role)
		}

		switch node.Role {
		case networkpolicy.RoleValidator:
			if node.PublicP2P != nil {
				v.fail(prefix+".public_p2p", "validator must not declare a public P2P endpoint")
			}
		case networkpolicy.RoleSeed, networkpolicy.RoleSentry, networkpolicy.RoleRPC:
			if node.PublicP2P == nil {
				v.fail(prefix+".public_p2p", "%s role requires an explicit public P2P endpoint", node.Role)
			}
		}
		if node.PublicP2P != nil {
			if err := validatePublicEndpoint(*node.PublicP2P); err != nil {
				v.fail(prefix+".public_p2p", "%v", err)
			} else {
				key := canonicalEndpointKey(*node.PublicP2P)
				if owner, exists := seenEndpoints[key]; exists {
					v.fail(prefix+".public_p2p", "public endpoint is already assigned to %q", owner)
				} else {
					seenEndpoints[key] = node.Name
				}
			}
		}
	}
}

func (v *validator) checkComposition(nodes []Node) {
	counts := map[networkpolicy.Role]int{}
	for i := range nodes {
		counts[nodes[i].Role]++
	}
	for _, role := range []networkpolicy.Role{
		networkpolicy.RoleSeed,
		networkpolicy.RoleSentry,
		networkpolicy.RoleValidator,
		networkpolicy.RoleRPC,
	} {
		if counts[role] == 0 {
			v.fail("composition."+string(role), "topology requires at least one %s node", role)
		}
	}
}

func (v *validator) checkPeerGraph(nodes []Node) {
	for i := range nodes {
		node := &nodes[i]
		seen := make(map[string]bool)
		sentryCount := 0
		sentryZones := make(map[string]bool)
		for j, peerName := range node.Peers {
			check := fmt.Sprintf("nodes[%d].peers[%d]", i, j)
			if seen[peerName] {
				v.fail(check, "peer %q is duplicated", peerName)
				continue
			}
			seen[peerName] = true
			if peerName == node.Name {
				v.fail(check, "node must not peer with itself")
				continue
			}
			peer := v.nodes[peerName]
			if peer == nil {
				v.fail(check, "peer %q is not declared", peerName)
				continue
			}
			if !allowedPeerRole(node.Role, peer.Role) {
				v.fail(check, "%s node must not dial %s node %q", node.Role, peer.Role, peerName)
			}
			if node.Role == networkpolicy.RoleValidator && peer.Role == networkpolicy.RoleSentry {
				sentryCount++
				if sentryZones[peer.Zone] {
					v.fail(check, "validator sentries must be in distinct zones; %q is repeated", peer.Zone)
				}
				sentryZones[peer.Zone] = true
			}
		}
		if node.Role == networkpolicy.RoleValidator {
			if sentryCount < 2 {
				v.fail(fmt.Sprintf("nodes[%d].peers", i), "validator requires at least two declared sentry peers")
			}
			if len(node.Protects) != 0 {
				v.fail(fmt.Sprintf("nodes[%d].protects", i), "validator must not protect other nodes")
			}
		}
	}
}

func (v *validator) checkProtectionGraph(nodes []Node) {
	for i := range nodes {
		node := &nodes[i]
		seen := make(map[string]bool)
		for j, protectedName := range node.Protects {
			check := fmt.Sprintf("nodes[%d].protects[%d]", i, j)
			if node.Role != networkpolicy.RoleSentry {
				v.fail(check, "only sentry nodes may declare protected validators")
			}
			if seen[protectedName] {
				v.fail(check, "protected node %q is duplicated", protectedName)
				continue
			}
			seen[protectedName] = true
			protected := v.nodes[protectedName]
			if protected == nil {
				v.fail(check, "protected node %q is not declared", protectedName)
				continue
			}
			if protected.Role != networkpolicy.RoleValidator {
				v.fail(check, "protected node %q must have validator role", protectedName)
				continue
			}
			if !contains(protected.Peers, node.Name) {
				v.fail(check, "validator %q does not reciprocally declare sentry %q", protectedName, node.Name)
			}
			if contains(node.Peers, protectedName) {
				v.fail(check, "sentry %q must not dial protected validator %q", node.Name, protectedName)
			}
		}
	}

	for i := range nodes {
		validatorNode := &nodes[i]
		if validatorNode.Role != networkpolicy.RoleValidator {
			continue
		}
		for j, peerName := range validatorNode.Peers {
			peer := v.nodes[peerName]
			if peer != nil && peer.Role == networkpolicy.RoleSentry &&
				!contains(peer.Protects, validatorNode.Name) {
				v.fail(fmt.Sprintf("nodes[%d].peers[%d]", i, j),
					"sentry %q does not declare protection for validator %q", peerName, validatorNode.Name)
			}
		}
	}
}

func (v *validator) checkIngress(ingress Ingress) {
	v.checkQueryIngress("rpc", ingress.RPC, MaxRPCRatePerSecond, MaxRPCBurst)
	v.checkQueryIngress("api", ingress.API, MaxAPIRatePerSecond, MaxAPIBurst)
}

func (v *validator) checkQueryIngress(name string, ingress QueryIngress, maxRate, maxBurst int) {
	prefix := "ingress." + name
	if ingress.MetricsEnabled {
		v.fail(prefix+".metrics_enabled", "metrics must never be exposed by public query ingress")
	}
	if ingress.AdminEnabled {
		v.fail(prefix+".admin_enabled", "admin surfaces must never be exposed by public query ingress")
	}
	if ingress.UnsafeEnabled {
		v.fail(prefix+".unsafe_enabled", "unsafe surfaces must never be exposed by public query ingress")
	}
	if !ingress.Enabled {
		if ingress.TLSOnly || ingress.ProxyOnly || ingress.RatePerSecond != 0 ||
			ingress.Burst != 0 || ingress.MaxBodyBytes != 0 ||
			ingress.TimeoutSeconds != 0 || ingress.MaxConcurrent != 0 ||
			len(ingress.AllowedMethods) != 0 || len(ingress.AllowedRoutes) != 0 ||
			ingress.WebSocketRoute != "" {
			v.fail(prefix, "disabled ingress must not retain active methods, routes, limits, proxy, TLS, or WebSocket settings")
		}
		return
	}
	if !ingress.TLSOnly {
		v.fail(prefix+".tls_only", "enabled public ingress must require TLS")
	}
	if !ingress.ProxyOnly {
		v.fail(prefix+".proxy_only", "enabled public ingress must be proxy-only")
	}
	checkBound := func(check string, value, maximum int) {
		if value < 1 || value > maximum {
			v.fail(prefix+"."+check, "value must be between 1 and %d", maximum)
		}
	}
	checkBound("rate_per_second", ingress.RatePerSecond, maxRate)
	checkBound("burst", ingress.Burst, maxBurst)
	checkBound("max_body_bytes", ingress.MaxBodyBytes, MaxRequestBodyBytes)
	checkBound("timeout_seconds", ingress.TimeoutSeconds, MaxTimeoutSeconds)
	checkBound("max_concurrent", ingress.MaxConcurrent, MaxConcurrent)
	if len(ingress.AllowedMethods) == 0 {
		v.fail(prefix+".allowed_methods", "enabled public ingress requires an explicit HTTP method allowlist")
	}
	seenMethods := make(map[string]bool)
	for i, method := range ingress.AllowedMethods {
		check := fmt.Sprintf("%s.allowed_methods[%d]", prefix, i)
		if seenMethods[method] {
			v.fail(check, "HTTP method %q is duplicated", method)
			continue
		}
		seenMethods[method] = true
		switch method {
		case "GET", "HEAD", "POST", "OPTIONS":
		default:
			v.fail(check, "HTTP method %q is not allowed on public query ingress", method)
		}
	}
	if len(ingress.AllowedRoutes) == 0 {
		v.fail(prefix+".allowed_routes", "enabled public ingress requires an explicit route allowlist")
	}
	seen := make(map[string]bool)
	for i, route := range ingress.AllowedRoutes {
		check := fmt.Sprintf("%s.allowed_routes[%d]", prefix, i)
		if seen[route] {
			v.fail(check, "route %q is duplicated", route)
			continue
		}
		seen[route] = true
		if route == "/" || !strings.HasPrefix(route, "/") ||
			strings.ContainsAny(route, "*?%\\#;") ||
			strings.Contains(route, "..") || strings.Contains(route, "//") ||
			containsControlCharacter(route) {
			v.fail(check, "route %q is not an exact safe path prefix", route)
		}
		for _, forbidden := range []string{
			"/metrics", "/admin", "/unsafe", "/debug", "/pprof",
			"/dial_seeds", "/dial_peers", "/unsafe_flush_mempool",
		} {
			if route == forbidden || strings.HasPrefix(route, forbidden+"/") {
				v.fail(check, "route %q exposes forbidden surface %q", route, forbidden)
			}
		}
	}
	if name == "api" && ingress.WebSocketRoute != "" {
		v.fail(prefix+".websocket_route", "API ingress must not expose a WebSocket route")
	}
	if name == "rpc" && ingress.WebSocketRoute != "" {
		if ingress.WebSocketRoute != "/websocket" {
			v.fail(prefix+".websocket_route", "RPC WebSocket route must be exactly /websocket")
		} else if !seen[ingress.WebSocketRoute] {
			v.fail(prefix+".websocket_route", "RPC WebSocket route must be present in allowed_routes")
		}
	}
}

func (v *validator) checkFlows(contract Contract) {
	for i := range contract.Flows {
		flow := contract.Flows[i]
		pair := flow.From + "\x00" + flow.To
		if _, exists := v.flows[pair]; exists {
			v.fail(fmt.Sprintf("flows[%d]", i), "flow %q -> %q is duplicated", flow.From, flow.To)
			continue
		}
		services := make(map[string]bool)
		v.flows[pair] = services
		if flow.From != "internet" && v.nodes[flow.From] == nil {
			v.fail(fmt.Sprintf("flows[%d].from", i), "source %q is not declared", flow.From)
		}
		target := v.nodes[flow.To]
		if target == nil {
			v.fail(fmt.Sprintf("flows[%d].to", i), "destination %q is not declared", flow.To)
		}
		if flow.From == flow.To {
			v.fail(fmt.Sprintf("flows[%d]", i), "self-flow is forbidden")
		}
		if len(flow.Services) == 0 {
			v.fail(fmt.Sprintf("flows[%d].services", i), "flow requires at least one service")
		}
		for j, service := range flow.Services {
			check := fmt.Sprintf("flows[%d].services[%d]", i, j)
			if services[service] {
				v.fail(check, "service %q is duplicated", service)
				continue
			}
			services[service] = true
			if service != "p2p" && service != "rpc" && service != "api" {
				v.fail(check, "service %q is forbidden or unknown", service)
				continue
			}
			v.checkFlowService(check, contract, flow, service, target)
		}
	}

	for i := range contract.Nodes {
		node := &contract.Nodes[i]
		for j, peer := range node.Peers {
			if !v.hasFlow(node.Name, peer, "p2p") {
				v.fail(fmt.Sprintf("nodes[%d].peers[%d]", i, j), "peer relationship lacks an explicit P2P flow")
			}
		}
		for j, protected := range node.Protects {
			if !v.hasFlow(node.Name, protected, "p2p") {
				v.fail(fmt.Sprintf("nodes[%d].protects[%d]", i, j), "protected relationship lacks an explicit sentry-to-validator P2P flow")
			}
		}
		if node.PublicP2P != nil && !v.hasFlow("internet", node.Name, "p2p") {
			v.fail(fmt.Sprintf("nodes[%d].public_p2p", i), "public P2P endpoint lacks an explicit internet flow")
		}
	}
	if contract.Ingress.RPC.Enabled && !v.hasInternetQueryFlow("rpc") {
		v.fail("ingress.rpc", "enabled RPC ingress lacks an explicit internet-to-RPC flow")
	}
	if contract.Ingress.API.Enabled && !v.hasInternetQueryFlow("api") {
		v.fail("ingress.api", "enabled API ingress lacks an explicit internet-to-RPC flow")
	}
}

func (v *validator) checkFlowService(check string, contract Contract, flow Flow, service string, target *Node) {
	if target == nil {
		return
	}
	if flow.From == "internet" {
		if target.Role == networkpolicy.RoleValidator {
			v.fail(check, "internet must never reach a validator")
			return
		}
		switch service {
		case "p2p":
			if target.PublicP2P == nil {
				v.fail(check, "internet P2P flow requires a declared public P2P endpoint")
			}
		case "rpc":
			if target.Role != networkpolicy.RoleRPC || !contract.Ingress.RPC.Enabled {
				v.fail(check, "internet RPC flow requires enabled RPC ingress on an RPC node")
			}
		case "api":
			if target.Role != networkpolicy.RoleRPC || !contract.Ingress.API.Enabled {
				v.fail(check, "internet API flow requires enabled API ingress on an RPC node")
			}
		}
		return
	}

	source := v.nodes[flow.From]
	if source == nil {
		return
	}
	if service != "p2p" {
		v.fail(check, "node-to-node flow may carry P2P only")
		return
	}
	backed := contains(source.Peers, target.Name) ||
		(source.Role == networkpolicy.RoleSentry &&
			target.Role == networkpolicy.RoleValidator &&
			contains(source.Protects, target.Name))
	if !backed {
		v.fail(check, "flow is not backed by a declared peer or protection relationship")
	}
	if target.Role == networkpolicy.RoleValidator &&
		(source.Role != networkpolicy.RoleSentry || !contains(source.Protects, target.Name)) {
		v.fail(check, "only a protecting sentry may flow to a validator")
	}
}

func (v *validator) hasFlow(from, to, service string) bool {
	return v.flows[from+"\x00"+to][service]
}

func (v *validator) hasInternetQueryFlow(service string) bool {
	for name, node := range v.nodes {
		if node.Role == networkpolicy.RoleRPC && v.hasFlow("internet", name, service) {
			return true
		}
	}
	return false
}

func allowedRole(role networkpolicy.Role) bool {
	switch role {
	case networkpolicy.RoleSeed, networkpolicy.RoleSentry,
		networkpolicy.RoleValidator, networkpolicy.RoleRPC:
		return true
	default:
		return false
	}
}

func allowedPeerRole(source, target networkpolicy.Role) bool {
	switch source {
	case networkpolicy.RoleValidator:
		return target == networkpolicy.RoleSentry
	case networkpolicy.RoleSeed, networkpolicy.RoleSentry, networkpolicy.RoleRPC:
		return target == networkpolicy.RoleSeed || target == networkpolicy.RoleSentry
	default:
		return false
	}
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func validatePublicEndpoint(endpoint Endpoint) error {
	if endpoint.Port < 1 || endpoint.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if endpoint.Host == "" || endpoint.Host != strings.ToLower(endpoint.Host) ||
		strings.HasSuffix(endpoint.Host, ".") {
		return fmt.Errorf("host must be a canonical lowercase host without a trailing dot")
	}
	if ip := net.ParseIP(endpoint.Host); ip != nil {
		if ip.IsUnspecified() || ip.IsLoopback() || ip.IsPrivate() ||
			ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
			ip.IsMulticast() || ip.Equal(net.IPv4bcast) {
			return fmt.Errorf("host must be a routable unicast address")
		}
		return nil
	}
	if endpoint.Host == "localhost" || strings.HasSuffix(endpoint.Host, ".localhost") {
		return fmt.Errorf("host must not resolve through the localhost namespace")
	}
	if len(endpoint.Host) > 253 {
		return fmt.Errorf("host exceeds 253 characters")
	}
	for _, label := range strings.Split(endpoint.Host, ".") {
		if len(label) < 1 || len(label) > 63 ||
			label[0] == '-' || label[len(label)-1] == '-' {
			return fmt.Errorf("host contains an invalid DNS label")
		}
		for _, character := range label {
			if (character < 'a' || character > 'z') &&
				(character < '0' || character > '9') && character != '-' {
				return fmt.Errorf("host contains a non-canonical DNS character")
			}
		}
	}
	return nil
}

func canonicalEndpointKey(endpoint Endpoint) string {
	host := endpoint.Host
	if ip := net.ParseIP(host); ip != nil {
		host = ip.String()
	}
	return net.JoinHostPort(host, strconv.Itoa(endpoint.Port))
}

func containsControlCharacter(value string) bool {
	for _, character := range value {
		if character < 0x20 || character == 0x7f {
			return true
		}
	}
	return false
}
