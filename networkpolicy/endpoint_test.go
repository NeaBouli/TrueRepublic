package networkpolicy

import (
	"strings"
	"testing"
)

const testNodeID = "0123456789abcdef0123456789abcdef01234567"

func TestCanonicalPeerValid(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		host string
		port int
	}{
		{"ipv4", testNodeID + "@203.0.113.7:26656", "203.0.113.7", 26656},
		{"hostname", testNodeID + "@sentry-1.example.org:26656", "sentry-1.example.org", 26656},
		{"hostname-uppercase-normalized", testNodeID + "@Sentry-1.EXAMPLE.org:26656", "sentry-1.example.org", 26656},
		{"ipv6", testNodeID + "@[2001:db8::1]:26656", "2001:db8::1", 26656},
		{"port-low-bound", testNodeID + "@203.0.113.7:1", "203.0.113.7", 1},
		{"port-high-bound", testNodeID + "@203.0.113.7:65535", "203.0.113.7", 65535},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ep, err := canonicalPeer(tc.raw)
			if err != nil {
				t.Fatalf("expected valid endpoint, got %v", err)
			}
			if ep.ID != testNodeID || ep.Host != tc.host || ep.Port != tc.port {
				t.Fatalf("got %+v", ep)
			}
		})
	}
}

func TestCanonicalPeerInvalid(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"missing-at", testNodeID + "203.0.113.7:26656", "expected <node-id>@<host>:<port>"},
		{"double-at", testNodeID + "@" + testNodeID + "@203.0.113.7:26656", "exactly one @"},
		{"id-too-short", "0123abcd@203.0.113.7:26656", "exactly 40 hexadecimal"},
		{"id-too-long", testNodeID + "ff@203.0.113.7:26656", "exactly 40 hexadecimal"},
		{"id-uppercase", strings.ToUpper(testNodeID) + "@203.0.113.7:26656", "lowercase"},
		{"id-non-hex", "zzzz56789abcdef0123456789abcdef01234567@203.0.113.7:26656", "hexadecimal"},
		{"missing-port", testNodeID + "@203.0.113.7", "<host>:<port>"},
		{"empty-host", testNodeID + "@:26656", "host must not be empty"},
		{"empty-port", testNodeID + "@203.0.113.7:", "port must not be empty"},
		{"port-zero", testNodeID + "@203.0.113.7:0", "outside 1..65535"},
		{"port-too-high", testNodeID + "@203.0.113.7:65536", "outside 1..65535"},
		{"port-non-numeric", testNodeID + "@203.0.113.7:p2p", "not numeric"},
		{"port-negative", testNodeID + "@203.0.113.7:-1", "not numeric"},
		{"empty", "", "expected <node-id>@<host>:<port>"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := canonicalPeer(tc.raw)
			if err == nil {
				t.Fatalf("expected error for %q", tc.raw)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error %q does not contain %q", err.Error(), tc.want)
			}
		})
	}
}

func TestParseListenerAddress(t *testing.T) {
	cases := []struct {
		name  string
		raw   string
		class hostClass
		valid bool
	}{
		{"tcp-loopback", "tcp://127.0.0.1:26657", hostLoopback, true},
		{"tcp-localhost", "tcp://localhost:26657", hostLoopback, true},
		{"tcp-ipv6-loopback", "tcp://[::1]:26657", hostLoopback, true},
		{"tcp-wildcard-v4", "tcp://0.0.0.0:26657", hostUnspecified, true},
		{"tcp-wildcard-v6", "tcp://[::]:26657", hostUnspecified, true},
		{"bare-loopback", "127.0.0.1:9090", hostLoopback, true},
		{"bare-localhost", "localhost:9090", hostLoopback, true},
		{"public-ip", "tcp://203.0.113.9:26657", hostRoutable, true},
		{"dns-name", "tcp://rpc.example.org:443", hostRoutable, true},
		{"empty", "", hostInvalid, false},
		{"no-port", "tcp://127.0.0.1", hostInvalid, false},
		{"bad-port", "tcp://127.0.0.1:99999", hostInvalid, false},
		{"unix-scheme", "unix:///var/run/sock", hostInvalid, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, class, err := parseListenerAddress(tc.raw)
			if tc.valid && err != nil {
				t.Fatalf("expected valid, got %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatalf("expected error for %q", tc.raw)
			}
			if tc.valid && class != tc.class {
				t.Fatalf("class %v, want %v", class, tc.class)
			}
		})
	}
}

func TestValidateNodeID(t *testing.T) {
	if err := validateNodeID(testNodeID); err != nil {
		t.Fatalf("canonical ID rejected: %v", err)
	}
	if err := validateNodeID(strings.ToUpper(testNodeID)); err == nil {
		t.Fatal("uppercase ID accepted")
	}
	if err := validateNodeID(testNodeID[:39]); err == nil {
		t.Fatal("short ID accepted")
	}
}
