package topologypolicy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"truerepublic/networkpolicy"
)

func exampleContractPath() string {
	return filepath.Join("..", "configs", "topology", "qualification.example.json")
}

func exampleContractBytes(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile(exampleContractPath())
	if err != nil {
		t.Fatalf("read example contract: %v", err)
	}
	return b
}

func exampleContract(t *testing.T) Contract {
	t.Helper()
	contract, err := Parse(strings.NewReader(string(exampleContractBytes(t))))
	if err != nil {
		t.Fatalf("parse example contract: %v", err)
	}
	return contract
}

func cloneContract(contract Contract) Contract {
	b, err := json.Marshal(contract)
	if err != nil {
		panic(err)
	}
	var copy Contract
	if err := json.Unmarshal(b, &copy); err != nil {
		panic(err)
	}
	return copy
}

func mustContainViolations(t *testing.T, violations []Violation, fragments ...string) {
	t.Helper()
	for _, fragment := range fragments {
		found := false
		for _, violation := range violations {
			target := violation.Check + ":" + violation.Message
			if strings.Contains(target, fragment) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected violation containing %q; got %d violation(s): %+v", fragment, len(violations), violations)
		}
	}
}

func deepNestedJSON() string {
	var b strings.Builder
	b.WriteString(`{"root":`)
	for i := 0; i < maxJSONDepth+1; i++ {
		b.WriteString("[")
	}
	b.WriteString("0")
	for i := 0; i < maxJSONDepth+1; i++ {
		b.WriteString("]")
	}
	b.WriteString("}")
	return b.String()
}

func TestParse(t *testing.T) {
	cases := []struct {
		name    string
		payload string
		wantErr string
	}{
		{name: "valid example", payload: string(exampleContractBytes(t))},
		{name: "empty", payload: "  \n", wantErr: "topology contract is empty"},
		{name: "trailing value", payload: "{}{}", wantErr: "trailing value"},
		{name: "duplicate key", payload: `{"version":"a","version":"b"}`, wantErr: "duplicate object key \"version\""},
		{name: "unknown field", payload: `{"version":"truerepublic.topology/v1","chain_id":"truerepublic-qualification-1","defaults":{"inbound":"deny","outbound":"deny"},"nodes":[],"flows":[],"ingress":{"rpc":{"enabled":false,"tls_only":false,"proxy_only":false,"rate_per_second":0,"burst":0,"max_body_bytes":0,"timeout_seconds":0,"max_concurrent":0,"allowed_routes":[],"metrics_enabled":false,"admin_enabled":false,"unsafe_enabled":false},"api":{"enabled":false,"tls_only":false,"proxy_only":false,"rate_per_second":0,"burst":0,"max_body_bytes":0,"timeout_seconds":0,"max_concurrent":0,"allowed_routes":[],"metrics_enabled":false,"admin_enabled":false,"unsafe_enabled":false}},"secret": "forbidden"}`, wantErr: "invalid topology schema"},
		{name: "secret-shaped unknown field", payload: `{"version":"truerepublic.topology/v1","chain_id":"truerepublic-qualification-1","defaults":{"inbound":"deny","outbound":"deny"},"nodes":[],"flows":[],"ingress":{"rpc":{"enabled":false,"tls_only":false,"proxy_only":false,"rate_per_second":0,"burst":0,"max_body_bytes":0,"timeout_seconds":0,"max_concurrent":0,"allowed_routes":[],"metrics_enabled":false,"admin_enabled":false,"unsafe_enabled":false},"api":{"enabled":false,"tls_only":false,"proxy_only":false,"rate_per_second":0,"burst":0,"max_body_bytes":0,"timeout_seconds":0,"max_concurrent":0,"allowed_routes":[],"metrics_enabled":false,"admin_enabled":false,"unsafe_enabled":false}},"api_key":"${TOP_SECRET}"}`, wantErr: "invalid topology schema"},
		{name: "depth bound", payload: deepNestedJSON(), wantErr: "maximum JSON depth"},
		{name: "size cap", payload: strings.Repeat("a", MaxContractBytes+1), wantErr: "exceeds 262144 bytes"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse(strings.NewReader(tc.payload))
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected parse error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestParseTrailingValueDoesNotLeakIt(t *testing.T) {
	const secret = "must-not-appear-in-error"
	_, err := Parse(strings.NewReader(`{} "` + secret + `"`))
	if err == nil {
		t.Fatal("trailing value unexpectedly passed")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("trailing-value error leaked rejected content: %v", err)
	}
}

func TestLoad(t *testing.T) {
	cases := []struct {
		name    string
		writeFn func() (string, func(), error)
		wantErr string
	}{
		{
			name: "valid file",
			writeFn: func() (string, func(), error) {
				tmp, err := os.CreateTemp(t.TempDir(), "topology-*.json")
				if err != nil {
					return "", nil, err
				}
				if _, err := tmp.Write(exampleContractBytes(t)); err != nil {
					return "", nil, err
				}
				name := tmp.Name()
				if err := tmp.Close(); err != nil {
					return "", nil, err
				}
				return name, func() { _ = os.Remove(name) }, nil
			},
			wantErr: "",
		},
		{name: "missing file", writeFn: func() (string, func(), error) { return "does-not-exist.json", func() {}, nil }, wantErr: "open topology contract"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path, cleanup, err := tc.writeFn()
			if err != nil {
				t.Fatalf("setup: %v", err)
			}
			defer cleanup()

			_, err = Load(path)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	base := exampleContract(t)

	cases := []struct {
		name      string
		mutate    func(Contract) Contract
		wantValid bool
		fragments []string
	}{
		{name: "valid example", mutate: func(c Contract) Contract { return c }, wantValid: true},
		{
			name:      "deterministic repeat",
			mutate:    func(c Contract) Contract { return c },
			wantValid: true,
		},
		{
			name: "unique names, IDs, and endpoints",
			mutate: func(c Contract) Contract {
				c.Nodes[1].Name = c.Nodes[0].Name
				c.Nodes[2].NodeID = c.Nodes[1].NodeID
				c.Nodes[3].Name = c.Nodes[4].Name
				c.Nodes[1].PublicP2P.Host = c.Nodes[2].PublicP2P.Host
				return c
			},
			wantValid: false,
			fragments: []string{"duplicated", "already assigned"},
		},
		{
			name: "equivalent IPv6 endpoints collide",
			mutate: func(c Contract) Contract {
				c.Nodes[0].PublicP2P.Host = "2001:db8::1"
				c.Nodes[1].PublicP2P.Host = "2001:0db8:0:0:0:0:0:1"
				return c
			},
			wantValid: false,
			fragments: []string{"public endpoint is already assigned"},
		},
		{
			name: "internet principal is reserved",
			mutate: func(c Contract) Contract {
				c.Nodes[0].Name = "internet"
				return c
			},
			wantValid: false,
			fragments: []string{"reserved for the external flow principal"},
		},
		{
			name: "required roles",
			mutate: func(c Contract) Contract {
				c.Nodes = c.Nodes[:2]
				c.Flows = c.Flows[:1]
				c.Ingress.RPC.Enabled = false
				c.Ingress.API.Enabled = false
				return c
			},
			wantValid: false,
			fragments: []string{"composition.validator", "composition.rpc"},
		},
		{
			name: "validator with public P2P",
			mutate: func(c Contract) Contract {
				for i := range c.Nodes {
					if c.Nodes[i].Role == networkpolicy.RoleValidator {
						c.Nodes[i].PublicP2P = c.Nodes[0].PublicP2P
						break
					}
				}
				return c
			},
			wantValid: false,
			fragments: []string{"must not declare a public P2P endpoint"},
		},
		{
			name: "validator one sentry",
			mutate: func(c Contract) Contract {
				for i := range c.Nodes {
					if c.Nodes[i].Role == networkpolicy.RoleValidator {
						c.Nodes[i].Peers = c.Nodes[i].Peers[:1]
					}
				}
				return c
			},
			wantValid: false,
			fragments: []string{"at least two declared sentry peers"},
		},
		{
			name: "validator two sentries same zone",
			mutate: func(c Contract) Contract {
				for i := range c.Nodes {
					if c.Nodes[i].Name == "sentry-2" {
						c.Nodes[i].Zone = "zone-a"
					}
				}
				return c
			},
			wantValid: false,
			fragments: []string{"validator sentries must be in distinct zones"},
		},
		{
			name: "validator peers to non-sentry",
			mutate: func(c Contract) Contract {
				for i := range c.Nodes {
					if c.Nodes[i].Role == networkpolicy.RoleValidator {
						c.Nodes[i].Peers = []string{"seed-1"}
					}
				}
				return c
			},
			wantValid: false,
			fragments: []string{"must not dial"},
		},
		{
			name: "sentry protection non-reciprocal",
			mutate: func(c Contract) Contract {
				for i := range c.Nodes {
					if c.Nodes[i].Role == networkpolicy.RoleSentry && c.Nodes[i].Name == "sentry-1" {
						c.Nodes[i].Protects = nil
					}
				}
				return c
			},
			wantValid: false,
			fragments: []string{"does not declare protection"},
		},
		{
			name: "sentry dials protected validator",
			mutate: func(c Contract) Contract {
				for i := range c.Nodes {
					if c.Nodes[i].Role == networkpolicy.RoleSentry && c.Nodes[i].Name == "sentry-1" {
						c.Nodes[i].Peers = append(c.Nodes[i].Peers, "validator-1")
					}
				}
				c.Flows = append(c.Flows, Flow{From: "sentry-1", To: "validator-1", Services: []string{"p2p"}})
				return c
			},
			wantValid: false,
			fragments: []string{"must not dial protected validator"},
		},
		{
			name: "undeclared flow source",
			mutate: func(c Contract) Contract {
				c.Flows = append(c.Flows, Flow{From: "ghost", To: "seed-1", Services: []string{"p2p"}})
				return c
			},
			wantValid: false,
			fragments: []string{"source \"ghost\" is not declared"},
		},
		{
			name: "unsupported service",
			mutate: func(c Contract) Contract {
				c.Flows[0].Services = []string{"admin"}
				return c
			},
			wantValid: false,
			fragments: []string{"service \"admin\" is forbidden or unknown"},
		},
		{
			name: "deny defaults",
			mutate: func(c Contract) Contract {
				c.Defaults.Inbound = "allow"
				c.Defaults.Outbound = "allow"
				return c
			},
			wantValid: false,
			fragments: []string{"default inbound policy must be deny", "default outbound policy must be deny"},
		},
		{
			name: "internet to validator",
			mutate: func(c Contract) Contract {
				for i, flow := range c.Flows {
					if flow.From == "internet" && flow.To == "rpc-1" {
						c.Flows[i].To = "validator-1"
					}
				}
				return c
			},
			wantValid: false,
			fragments: []string{"internet must never reach a validator"},
		},
		{
			name: "ingress tls", mutate: func(c Contract) Contract { c.Ingress.RPC.TLSOnly = false; return c }, wantValid: false, fragments: []string{"ingress.rpc.tls_only"},
		},
		{
			name: "ingress proxy", mutate: func(c Contract) Contract { c.Ingress.RPC.ProxyOnly = false; return c }, wantValid: false, fragments: []string{"ingress.rpc.proxy_only"},
		},
		{
			name: "ingress rpc rate", mutate: func(c Contract) Contract { c.Ingress.RPC.RatePerSecond = 0; return c }, wantValid: false, fragments: []string{"ingress.rpc.rate_per_second"},
		},
		{
			name: "ingress api burst", mutate: func(c Contract) Contract { c.Ingress.API.Burst = MaxAPIBurst + 1; return c }, wantValid: false, fragments: []string{"ingress.api.burst"},
		},
		{
			name: "ingress body", mutate: func(c Contract) Contract { c.Ingress.RPC.MaxBodyBytes = 0; return c }, wantValid: false, fragments: []string{"ingress.rpc.max_body_bytes"},
		},
		{
			name: "ingress timeout", mutate: func(c Contract) Contract { c.Ingress.API.TimeoutSeconds = MaxTimeoutSeconds + 1; return c }, wantValid: false, fragments: []string{"ingress.api.timeout_seconds"},
		},
		{
			name: "ingress concurrency", mutate: func(c Contract) Contract { c.Ingress.API.MaxConcurrent = 0; return c }, wantValid: false, fragments: []string{"ingress.api.max_concurrent"},
		},
		{
			name: "ingress methods required", mutate: func(c Contract) Contract { c.Ingress.RPC.AllowedMethods = nil; return c }, wantValid: false, fragments: []string{"ingress.rpc.allowed_methods"},
		},
		{
			name: "ingress method forbidden", mutate: func(c Contract) Contract { c.Ingress.API.AllowedMethods = []string{"DELETE"}; return c }, wantValid: false, fragments: []string{"is not allowed on public query ingress"},
		},
		{
			name: "ingress root route forbidden", mutate: func(c Contract) Contract { c.Ingress.RPC.AllowedRoutes = []string{"/"}; return c }, wantValid: false, fragments: []string{"is not an exact safe path prefix"},
		},
		{
			name: "dangerous CometBFT route forbidden", mutate: func(c Contract) Contract { c.Ingress.RPC.AllowedRoutes = []string{"/dial_seeds"}; return c }, wantValid: false, fragments: []string{"exposes forbidden surface"},
		},
		{
			name: "encoded forbidden route rejected", mutate: func(c Contract) Contract { c.Ingress.RPC.AllowedRoutes = []string{"/%6detrics"}; return c }, wantValid: false, fragments: []string{"is not an exact safe path prefix"},
		},
		{
			name: "ingress route forbidden", mutate: func(c Contract) Contract { c.Ingress.RPC.AllowedRoutes = []string{"/metrics"}; return c }, wantValid: false, fragments: []string{"exposes forbidden surface"},
		},
		{
			name: "RPC WebSocket route must be allowlisted", mutate: func(c Contract) Contract {
				c.Ingress.RPC.AllowedRoutes = []string{"/status"}
				return c
			}, wantValid: false, fragments: []string{"RPC WebSocket route must be present"},
		},
		{
			name: "API WebSocket forbidden", mutate: func(c Contract) Contract {
				c.Ingress.API.WebSocketRoute = "/websocket"
				return c
			}, wantValid: false, fragments: []string{"API ingress must not expose a WebSocket route"},
		},
		{
			name: "ingress metrics exposed", mutate: func(c Contract) Contract { c.Ingress.RPC.MetricsEnabled = true; return c }, wantValid: false, fragments: []string{"metrics must never be exposed"},
		},
		{
			name: "ingress admin exposed", mutate: func(c Contract) Contract { c.Ingress.API.AdminEnabled = true; return c }, wantValid: false, fragments: []string{"admin surfaces must never be exposed"},
		},
		{
			name: "ingress unsafe exposed", mutate: func(c Contract) Contract { c.Ingress.API.UnsafeEnabled = true; return c }, wantValid: false, fragments: []string{"unsafe surfaces must never be exposed"},
		},
		{
			name: "disabled ingress is empty",
			mutate: func(c Contract) Contract {
				c.Ingress.API.Enabled = false
				return c
			},
			wantValid: false,
			fragments: []string{"disabled ingress must not retain active"},
		},
		{
			name: "environment-looking host", mutate: func(c Contract) Contract { c.Nodes[0].PublicP2P.Host = "${SEED_HOST}"; return c }, wantValid: false, fragments: []string{"canonical lowercase host"},
		},
		{
			name: "link-local public host", mutate: func(c Contract) Contract { c.Nodes[0].PublicP2P.Host = "169.254.169.254"; return c }, wantValid: false, fragments: []string{"routable unicast address"},
		},
		{
			name: "localhost namespace", mutate: func(c Contract) Contract { c.Nodes[0].PublicP2P.Host = "seed.localhost"; return c }, wantValid: false, fragments: []string{"localhost namespace"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			contract := tc.mutate(cloneContract(base))
			report := Validate(contract)
			if report.Valid != tc.wantValid {
				t.Fatalf("valid mismatch: got %v want %v with %d violations", report.Valid, tc.wantValid, len(report.Violations))
			}
			if tc.name == "deterministic repeat" {
				repeat := Validate(contract)
				if !reflect.DeepEqual(report, repeat) {
					t.Fatalf("report mismatch\nfirst=%+v\nsecond=%+v", report, repeat)
				}
			}
			if tc.wantValid {
				if len(report.Violations) != 0 {
					t.Fatalf("expected no violations, got %+v", report.Violations)
				}
				return
			}
			mustContainViolations(t, report.Violations, tc.fragments...)
		})
	}

	for _, node := range base.Nodes {
		if node.PublicP2P == nil {
			continue
		}
		if !strings.HasSuffix(node.PublicP2P.Host, ".invalid") {
			t.Fatalf("public host %s must be .invalid", node.Name)
		}
		if strings.ContainsAny(node.PublicP2P.Host, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
			t.Fatalf("node public host %s must be synthetic lowercase", node.Name)
		}
	}
}
