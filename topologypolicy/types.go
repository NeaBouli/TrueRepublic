// Package topologypolicy validates a secret-free, multi-node deployment
// contract. It correlates node roles and permitted flows but never resolves
// DNS, reads a node home, mutates a firewall, or deploys infrastructure.
package topologypolicy

import "truerepublic/networkpolicy"

const (
	ContractVersion     = "truerepublic.topology/v1"
	MaxContractBytes    = 256 * 1024
	MaxRPCRatePerSecond = 10
	MaxRPCBurst         = 20
	MaxAPIRatePerSecond = 30
	MaxAPIBurst         = 50
	MaxRequestBodyBytes = 1 << 20
	MaxTimeoutSeconds   = 30
	MaxConcurrent       = 1024
)

type Contract struct {
	Version  string   `json:"version"`
	ChainID  string   `json:"chain_id"`
	Defaults Defaults `json:"defaults"`
	Nodes    []Node   `json:"nodes"`
	Flows    []Flow   `json:"flows"`
	Ingress  Ingress  `json:"ingress"`
}

type Defaults struct {
	Inbound  string `json:"inbound"`
	Outbound string `json:"outbound"`
}

type Node struct {
	Name      string             `json:"name"`
	Role      networkpolicy.Role `json:"role"`
	Zone      string             `json:"zone"`
	NodeID    string             `json:"node_id"`
	PublicP2P *Endpoint          `json:"public_p2p,omitempty"`
	Peers     []string           `json:"peers"`
	Protects  []string           `json:"protects"`
}

type Endpoint struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type Flow struct {
	From     string   `json:"from"`
	To       string   `json:"to"`
	Services []string `json:"services"`
}

type Ingress struct {
	RPC QueryIngress `json:"rpc"`
	API QueryIngress `json:"api"`
}

type QueryIngress struct {
	Enabled        bool     `json:"enabled"`
	TLSOnly        bool     `json:"tls_only"`
	ProxyOnly      bool     `json:"proxy_only"`
	RatePerSecond  int      `json:"rate_per_second"`
	Burst          int      `json:"burst"`
	MaxBodyBytes   int      `json:"max_body_bytes"`
	TimeoutSeconds int      `json:"timeout_seconds"`
	MaxConcurrent  int      `json:"max_concurrent"`
	AllowedMethods []string `json:"allowed_methods"`
	AllowedRoutes  []string `json:"allowed_routes"`
	WebSocketRoute string   `json:"websocket_route"`
	MetricsEnabled bool     `json:"metrics_enabled"`
	AdminEnabled   bool     `json:"admin_enabled"`
	UnsafeEnabled  bool     `json:"unsafe_enabled"`
}

type Violation struct {
	Check   string `json:"check"`
	Message string `json:"message"`
}

type Report struct {
	Version    string      `json:"version"`
	ChainID    string      `json:"chain_id"`
	NodeCount  int         `json:"node_count"`
	Valid      bool        `json:"valid"`
	Violations []Violation `json:"violations"`
}
