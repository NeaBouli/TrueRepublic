package networkpolicy

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// hostClass classifies a peer or listener host for policy decisions.
type hostClass int

const (
	hostInvalid     hostClass = iota
	hostUnspecified           // 0.0.0.0, ::, or empty — wildcard bind/peer
	hostLoopback              // 127.0.0.0/8, ::1, "localhost"
	hostRoutable              // DNS name or any non-loopback, non-unspecified IP
)

// peerEndpoint is a canonical CometBFT peer endpoint: <40 lowercase hex node
// ID>@<host>:<port>. The node ID is the peer identity; the address is only
// its current location.
type peerEndpoint struct {
	ID   string
	Host string
	Port int
}

// canonicalPeer parses and validates one CometBFT peer endpoint in its
// canonical form. It rejects anything that is not exactly
// "<40 lowercase hex>@<host>:<port 1..65535>".
func canonicalPeer(raw string) (peerEndpoint, error) {
	var ep peerEndpoint
	id, addr, found := strings.Cut(raw, "@")
	if !found {
		return ep, fmt.Errorf("expected <node-id>@<host>:<port>")
	}
	if strings.Contains(addr, "@") {
		return ep, fmt.Errorf("expected exactly one @ separator")
	}
	if err := validateNodeID(id); err != nil {
		return ep, err
	}
	host, port, err := splitHostPort(addr)
	if err != nil {
		return ep, err
	}
	ep.ID = id
	ep.Host = host
	ep.Port = port
	return ep, nil
}

// validateNodeID enforces the canonical CometBFT node ID form: exactly 40
// lowercase hexadecimal characters.
func validateNodeID(id string) error {
	if len(id) != 40 {
		return fmt.Errorf("node ID must be exactly 40 hexadecimal characters, got %d", len(id))
	}
	for _, c := range id {
		switch {
		case c >= '0' && c <= '9':
		case c >= 'a' && c <= 'f':
		case c >= 'A' && c <= 'F':
			return fmt.Errorf("node ID must be lowercase hexadecimal; uppercase is not canonical")
		default:
			return fmt.Errorf("node ID contains non-hexadecimal character %q", string(c))
		}
	}
	return nil
}

// splitHostPort splits host:port (IPv6 must be bracketed) and enforces a
// non-empty host and a numeric port in 1..65535.
func splitHostPort(addr string) (string, int, error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return "", 0, fmt.Errorf("expected <host>:<port>: %v", err)
	}
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return "", 0, fmt.Errorf("host must not be empty")
	}
	if portStr == "" {
		return "", 0, fmt.Errorf("port must not be empty")
	}
	for _, c := range portStr {
		if c < '0' || c > '9' {
			return "", 0, fmt.Errorf("port %q is not numeric", portStr)
		}
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return "", 0, fmt.Errorf("port %q is outside 1..65535", portStr)
	}
	return host, port, nil
}

// classifyHost classifies a listener or peer host. DNS names other than
// "localhost" are treated as routable.
func classifyHost(host string) hostClass {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return hostUnspecified
	}
	if host == "localhost" {
		return hostLoopback
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return hostRoutable // DNS name
	}
	if ip.IsUnspecified() {
		return hostUnspecified
	}
	if ip.IsLoopback() {
		return hostLoopback
	}
	return hostRoutable
}

// parseListenerAddress parses a CometBFT/SDK listener address such as
// "tcp://127.0.0.1:26657" or "localhost:9090" and returns its host, port,
// and classification.
func parseListenerAddress(raw string) (host string, port int, class hostClass, err error) {
	addr := strings.TrimSpace(raw)
	if addr == "" {
		return "", 0, hostInvalid, fmt.Errorf("listener address must not be empty")
	}
	if i := strings.Index(addr, "://"); i >= 0 {
		scheme := strings.ToLower(addr[:i])
		if scheme != "tcp" {
			return "", 0, hostInvalid, fmt.Errorf("unsupported listener scheme %q", scheme)
		}
		addr = addr[i+3:]
	}
	host, port, err = splitHostPort(addr)
	if err != nil {
		return "", 0, hostInvalid, err
	}
	return host, port, classifyHost(host), nil
}
