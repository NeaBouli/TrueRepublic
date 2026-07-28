package networkpolicy

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/cometbft/cometbft/p2p"
)

const (
	peerIDA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	peerIDB = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	peerIDC = "cccccccccccccccccccccccccccccccccccccccc"
)

// makeHome writes a minimal but real initialized node home: config.toml,
// app.toml, and a freshly generated node key. The generated key is a
// throwaway test identity that never leaves the temp directory.
func makeHome(t *testing.T, configTOML, appTOML string) (home, selfID string) {
	t.Helper()
	home = t.TempDir()
	configDir := filepath.Join(home, "config")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.toml"), []byte(configTOML), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "app.toml"), []byte(appTOML), 0o600); err != nil {
		t.Fatal(err)
	}
	nodeKey, err := p2p.LoadOrGenNodeKey(filepath.Join(configDir, "node_key.json"))
	if err != nil {
		t.Fatal(err)
	}
	return home, string(nodeKey.ID())
}

func p2pSection(peers, privateIDs, unconditionalIDs string, seedMode, pex bool, inbound, outbound int) string {
	return "[p2p]\n" +
		"laddr = \"tcp://127.0.0.1:26656\"\n" +
		boolTOML("seed_mode", seedMode) +
		boolTOML("pex", pex) +
		"persistent_peers = \"" + peers + "\"\n" +
		"private_peer_ids = \"" + privateIDs + "\"\n" +
		"unconditional_peer_ids = \"" + unconditionalIDs + "\"\n" +
		"max_num_inbound_peers = " + itoa(inbound) + "\n" +
		"max_num_outbound_peers = " + itoa(outbound) + "\n"
}

func rpcSection(laddr string, unsafe bool, pprof string, cors string) string {
	return "[rpc]\n" +
		"laddr = \"" + laddr + "\"\n" +
		boolTOML("unsafe", unsafe) +
		"pprof_laddr = \"" + pprof + "\"\n" +
		"cors_allowed_origins = [" + cors + "]\n"
}

func appSection(apiEnable, grpcEnable, grpcWebEnable, unsafeCORS bool, apiAddr, grpcAddr string) string {
	return "[api]\n" +
		boolTOML("enable", apiEnable) +
		"address = \"" + apiAddr + "\"\n" +
		boolTOML("enabled-unsafe-cors", unsafeCORS) +
		"\n[grpc]\n" +
		boolTOML("enable", grpcEnable) +
		"address = \"" + grpcAddr + "\"\n" +
		"\n[grpc-web]\n" +
		boolTOML("enable", grpcWebEnable)
}

func boolTOML(key string, value bool) string {
	if value {
		return key + " = true\n"
	}
	return key + " = false\n"
}

func itoa(n int) string {
	return strconv.Itoa(n)
}

const (
	goodRPC    = "tcp://127.0.0.1:26657"
	loopbackGW = "tcp://localhost:1317"
	loopbackGR = "localhost:9090"
)

// goodValidatorConfig is a compliant validator home: sentry persistent peers,
// identity protection, no inbound, private client surface.
func goodValidatorConfig(selfID string) (string, string) {
	peers := peerIDA + "@203.0.113.10:26656," + peerIDB + "@203.0.113.11:26656"
	config := p2pSection(peers, "", peerIDA+","+peerIDB, false, false, 0, 10) +
		rpcSection(goodRPC, false, "", "")
	app := appSection(false, false, false, false, loopbackGW, loopbackGR)
	return config, app
}

func goodSentryConfig() (string, string) {
	peers := peerIDA + "@203.0.113.10:26656," + peerIDB + "@203.0.113.11:26656"
	config := p2pSection(peers, peerIDC, peerIDC, false, true, 40, 10) +
		rpcSection(goodRPC, false, "", "")
	config = publicP2PConfig(config, "203.0.113.20")
	app := appSection(false, false, false, false, loopbackGW, loopbackGR)
	return config, app
}

func goodSeedConfig() (string, string) {
	config := p2pSection("", "", "", true, true, 60, 20) +
		rpcSection(goodRPC, false, "", "")
	config = publicP2PConfig(config, "203.0.113.21")
	app := appSection(false, false, false, false, loopbackGW, loopbackGR)
	return config, app
}

func goodRPCConfig() (string, string) {
	peers := peerIDA + "@203.0.113.10:26656"
	config := p2pSection(peers, "", "", false, true, 40, 10) +
		rpcSection(goodRPC, false, "", "\"https://wallet.example.org\"")
	config = publicP2PConfig(config, "203.0.113.22")
	app := appSection(true, true, true, false, loopbackGW, loopbackGR)
	return config, app
}

func publicP2PConfig(config, host string) string {
	return strings.Replace(config,
		"laddr = \"tcp://127.0.0.1:26656\"\n",
		"laddr = \"tcp://"+host+":26656\"\nexternal_address = \""+host+":26656\"\n", 1)
}

func goodPrivateConfig() (string, string) {
	peers := peerIDA + "@203.0.113.10:26656"
	config := p2pSection(peers, "", "", false, false, 0, 10) +
		rpcSection(goodRPC, false, "localhost:6060", "")
	app := appSection(false, true, false, false, loopbackGW, loopbackGR)
	return config, app
}

func TestValidateRolesPass(t *testing.T) {
	cases := []struct {
		name string
		role Role
		home func(selfID string) (string, string)
	}{
		{"validator", RoleValidator, goodValidatorConfig},
		{"sentry", RoleSentry, func(string) (string, string) { return goodSentryConfig() }},
		{"seed", RoleSeed, func(string) (string, string) { return goodSeedConfig() }},
		{"rpc", RoleRPC, func(string) (string, string) { return goodRPCConfig() }},
		{"private", RolePrivate, func(string) (string, string) { return goodPrivateConfig() }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			configTOML, appTOML := tc.home("")
			home, selfID := makeHome(t, configTOML, appTOML)
			// Rebuild after generating the node identity so every role case
			// follows the same initialized-home path.
			configTOML, appTOML = tc.home(selfID)
			rewriteConfig(t, home, configTOML, appTOML)
			report := Validate(home, tc.role, Options{})
			if !report.Valid {
				t.Fatalf("expected valid %s home, got %+v", tc.role, report.Violations)
			}
		})
	}
}

func rewriteConfig(t *testing.T, home, configTOML, appTOML string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(home, "config", "config.toml"), []byte(configTOML), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "config", "app.toml"), []byte(appTOML), 0o600); err != nil {
		t.Fatal(err)
	}
}

// violationCase builds one home expected to fail with a violation whose Check
// field contains wantCheck.
type violationCase struct {
	name      string
	role      Role
	configApp func(selfID string) (string, string)
	wantCheck string
}

func TestValidateViolations(t *testing.T) {
	validatorBase := func(mutate func(*string, *string)) func(string) (string, string) {
		return func(selfID string) (string, string) {
			config, app := goodValidatorConfig(selfID)
			if mutate != nil {
				mutate(&config, &app)
			}
			return config, app
		}
	}
	replace := func(old, new string) func(*string, *string) {
		return func(config, app *string) {
			*config = strings.Replace(*config, old, new, 1)
			*app = strings.Replace(*app, old, new, 1)
		}
	}

	cases := []violationCase{
		// Shared client-facing boundary.
		{"rpc-wildcard", RoleValidator, validatorBase(replace(goodRPC, "tcp://0.0.0.0:26657")), "rpc.laddr"},
		{"rpc-public-ip", RoleValidator, validatorBase(replace(goodRPC, "tcp://203.0.113.5:26657")), "rpc.laddr"},
		{"rpc-unsafe", RoleValidator, validatorBase(replace("unsafe = false", "unsafe = true")), "rpc.unsafe"},
		{"rpc-cors-wildcard", RoleValidator, validatorBase(replace("cors_allowed_origins = []", "cors_allowed_origins = [\"*\"]")), "rpc.cors_allowed_origins"},
		{"pprof-wildcard", RoleValidator, validatorBase(replace("pprof_laddr = \"\"", "pprof_laddr = \"0.0.0.0:6060\"")), "rpc.pprof_laddr"},
		{"pprof-public", RoleValidator, validatorBase(replace("pprof_laddr = \"\"", "pprof_laddr = \"203.0.113.5:6060\"")), "rpc.pprof_laddr"},
		{"prometheus-wildcard", RoleValidator, validatorBase(replace(
			"[p2p]\n", "[instrumentation]\nprometheus = true\nprometheus_listen_addr = \":26660\"\n\n[p2p]\n")), "instrumentation.prometheus_listen_addr"},
		{"prometheus-public", RoleValidator, validatorBase(replace(
			"[p2p]\n", "[instrumentation]\nprometheus = true\nprometheus_listen_addr = \"203.0.113.5:26660\"\n\n[p2p]\n")), "instrumentation.prometheus_listen_addr"},
		{"api-unsafe-cors", RoleRPC, rpcMutated(replace("enabled-unsafe-cors = false", "enabled-unsafe-cors = true")), "api.enabled-unsafe-cors"},
		{"api-wildcard", RoleRPC, rpcMutated(replace(loopbackGW, "tcp://0.0.0.0:1317")), "api.address"},
		{"grpc-wildcard", RoleRPC, rpcMutated(replace(loopbackGR, "0.0.0.0:9090")), "grpc.address"},

		// Peer endpoint parsing and identity.
		{"peer-uppercase-id", RoleValidator, validatorBase(replace(peerIDA, strings.ToUpper(peerIDA))), "p2p.persistent_peers"},
		{"peer-short-id", RoleValidator, validatorBase(replace(peerIDA, "aaaa")), "p2p.persistent_peers"},
		{"peer-port-zero", RoleValidator, validatorBase(replace("203.0.113.10:26656", "203.0.113.10:0")), "p2p.persistent_peers"},
		{"peer-port-too-high", RoleValidator, validatorBase(replace("203.0.113.10:26656", "203.0.113.10:70000")), "p2p.persistent_peers"},
		{"peer-missing-port", RoleValidator, validatorBase(replace("203.0.113.10:26656", "203.0.113.10")), "p2p.persistent_peers"},
		{"peer-duplicate", RoleValidator, validatorBase(replace(
			peerIDA+"@203.0.113.10:26656,"+peerIDB+"@203.0.113.11:26656",
			peerIDA+"@203.0.113.10:26656,"+peerIDA+"@203.0.113.10:26656")), "p2p.persistent_peers"},
		{"peer-id-conflict-different-address", RoleValidator, validatorBase(replace(
			peerIDA+"@203.0.113.10:26656,"+peerIDB+"@203.0.113.11:26656",
			peerIDA+"@203.0.113.10:26656,"+peerIDA+"@203.0.113.99:26656")), "p2p.persistent_peers"},
		{"peer-self-conflict", RoleValidator, func(selfID string) (string, string) {
			return goodValidatorConfigForPeers(selfID, selfID+"@203.0.113.10:26656")
		}, "p2p.persistent_peers"},
		{"peer-loopback-production", RoleValidator, validatorBase(replace("203.0.113.10:26656", "127.0.0.1:26656")), "p2p.persistent_peers"},
		{"peer-wildcard", RoleValidator, validatorBase(replace("203.0.113.10:26656", "0.0.0.0:26656")), "p2p.persistent_peers"},
		{"external-address-loopback", RoleValidator, validatorBase(replace(
			"[p2p]\n", "[p2p]\nexternal_address = \"tcp://127.0.0.1:26656\"\n")), "p2p.external_address"},
		{"external-address-wildcard", RoleValidator, validatorBase(replace(
			"[p2p]\n", "[p2p]\nexternal_address = \"tcp://0.0.0.0:26656\"\n")), "p2p.external_address"},
		{"unconditional-id-invalid", RoleValidator, validatorBase(replace(
			"unconditional_peer_ids = \""+peerIDA+","+peerIDB+"\"",
			"unconditional_peer_ids = \"not-a-node-id\"")), "p2p.unconditional_peer_ids"},
		{"unconditional-id-duplicate", RoleValidator, validatorBase(replace(
			"unconditional_peer_ids = \""+peerIDA+","+peerIDB+"\"",
			"unconditional_peer_ids = \""+peerIDA+","+peerIDA+"\"")), "p2p.unconditional_peer_ids"},

		// Validator role profile.
		{"validator-seed-mode", RoleValidator, validatorBase(replace("seed_mode = false", "seed_mode = true")), "p2p.seed_mode"},
		{"validator-pex", RoleValidator, validatorBase(replace("pex = false", "pex = true")), "p2p.pex"},
		{"validator-seeds", RoleValidator, validatorBase(replace(
			"private_peer_ids = \"\"", "seeds = \""+peerIDC+"@203.0.113.12:26656\"\nprivate_peer_ids = \"\"")), "p2p.seeds"},
		{"validator-self-unconditional-id", RoleValidator, func(selfID string) (string, string) {
			config, app := goodValidatorConfig(selfID)
			config = strings.Replace(config,
				"unconditional_peer_ids = \""+peerIDA+","+peerIDB+"\"",
				"unconditional_peer_ids = \""+selfID+","+peerIDB+"\"", 1)
			return config, app
		}, "p2p.unconditional_peer_ids"},
		{"validator-no-peers", RoleValidator, validatorBase(replace(
			peerIDA+"@203.0.113.10:26656,"+peerIDB+"@203.0.113.11:26656", "")), "p2p.persistent_peers"},
		{"validator-one-sentry", RoleValidator, validatorBase(replace(
			peerIDA+"@203.0.113.10:26656,"+peerIDB+"@203.0.113.11:26656",
			peerIDA+"@203.0.113.10:26656")), "p2p.persistent_peers"},
		{"validator-public-p2p-listener", RoleValidator, validatorBase(replace(
			"tcp://127.0.0.1:26656", "tcp://203.0.113.30:26656")), "p2p.laddr"},
		{"validator-external-address", RoleValidator, validatorBase(replace(
			"[p2p]\n", "[p2p]\nexternal_address = \"203.0.113.30:26656\"\n")), "p2p.external_address"},
		{"validator-no-unconditional", RoleValidator, validatorBase(replace(
			"unconditional_peer_ids = \""+peerIDA+","+peerIDB+"\"", "unconditional_peer_ids = \"\"")), "p2p.unconditional_peer_ids"},
		{"validator-unconditional-not-persistent", RoleValidator, validatorBase(replace(
			"unconditional_peer_ids = \""+peerIDA+","+peerIDB+"\"", "unconditional_peer_ids = \""+peerIDA+","+peerIDC+"\"")), "p2p.unconditional_peer_ids"},
		{"validator-inbound", RoleValidator, validatorBase(replace("max_num_inbound_peers = 0", "max_num_inbound_peers = 8")), "p2p.max_num_inbound_peers"},
		{"validator-api-enabled", RoleValidator, validatorBase(replace("[api]\nenable = false", "[api]\nenable = true")), "api.enable"},
		{"validator-grpc-web-enabled", RoleValidator, validatorBase(replace("[grpc-web]\nenable = false", "[grpc-web]\nenable = true")), "grpc-web.enable"},

		// Sentry role profile.
		{"sentry-seed-mode", RoleSentry, sentryMutated(replace("seed_mode = false", "seed_mode = true")), "p2p.seed_mode"},
		{"sentry-loopback-p2p-listener", RoleSentry, sentryMutated(replace(
			"tcp://203.0.113.20:26656", "tcp://127.0.0.1:26656")), "p2p.laddr"},
		{"sentry-wildcard-p2p-listener", RoleSentry, sentryMutated(replace(
			"tcp://203.0.113.20:26656", "tcp://0.0.0.0:26656")), "p2p.laddr"},
		{"sentry-dns-p2p-listener", RoleSentry, sentryMutated(replace(
			"tcp://203.0.113.20:26656", "tcp://sentry.example.org:26656")), "p2p.laddr"},
		{"sentry-missing-external-address", RoleSentry, sentryMutated(replace(
			"external_address = \"203.0.113.20:26656\"\n", "")), "p2p.external_address"},
		{"sentry-pex-off", RoleSentry, sentryMutated(replace("pex = true", "pex = false")), "p2p.pex"},
		{"sentry-no-peers", RoleSentry, sentryMutated(replace(
			peerIDA+"@203.0.113.10:26656,"+peerIDB+"@203.0.113.11:26656", "")), "p2p.persistent_peers"},
		{"sentry-no-private-ids", RoleSentry, sentryMutated(replace(
			"private_peer_ids = \""+peerIDC+"\"", "private_peer_ids = \"\"")), "p2p.private_peer_ids"},
		{"sentry-self-private-id", RoleSentry, func(selfID string) (string, string) {
			config, app := goodSentryConfig()
			config = strings.Replace(config,
				"private_peer_ids = \""+peerIDC+"\"",
				"private_peer_ids = \""+selfID+"\"", 1)
			config = strings.Replace(config,
				"unconditional_peer_ids = \""+peerIDC+"\"",
				"unconditional_peer_ids = \""+selfID+"\"", 1)
			return config, app
		}, "p2p.private_peer_ids"},
		{"sentry-no-unconditional", RoleSentry, sentryMutated(replace(
			"unconditional_peer_ids = \""+peerIDC+"\"", "unconditional_peer_ids = \"\"")), "p2p.unconditional_peer_ids"},
		{"sentry-private-not-unconditional", RoleSentry, sentryMutated(replace(
			"private_peer_ids = \""+peerIDC+"\"", "private_peer_ids = \""+peerIDA+"\"")), "p2p.private_peer_ids"},
		{"sentry-protected-validator-not-unconditional", RoleSentry, sentryMutated(replace(
			"unconditional_peer_ids = \""+peerIDC+"\"", "unconditional_peer_ids = \""+peerIDA+"\"")), "p2p.private_peer_ids"},
		{"sentry-zero-inbound", RoleSentry, sentryMutated(replace("max_num_inbound_peers = 40", "max_num_inbound_peers = 0")), "p2p.max_num_inbound_peers"},
		{"sentry-zero-outbound", RoleSentry, sentryMutated(replace("max_num_outbound_peers = 10", "max_num_outbound_peers = 0")), "p2p.max_num_outbound_peers"},
		{"sentry-api-enabled", RoleSentry, sentryMutated(replace("[api]\nenable = false", "[api]\nenable = true")), "api.enable"},

		// Seed role profile.
		{"seed-mode-off", RoleSeed, seedMutated(replace("seed_mode = true", "seed_mode = false")), "p2p.seed_mode"},
		{"seed-pex-off", RoleSeed, seedMutated(replace("pex = true", "pex = false")), "p2p.pex"},
		{"seed-private-ids", RoleSeed, seedMutated(replace("private_peer_ids = \"\"", "private_peer_ids = \""+peerIDA+"\"")), "p2p.private_peer_ids"},
		{"seed-zero-inbound", RoleSeed, seedMutated(replace("max_num_inbound_peers = 60", "max_num_inbound_peers = 0")), "p2p.max_num_inbound_peers"},

		// RPC role profile.
		{"rpc-no-peers", RoleRPC, rpcMutated(replace(peerIDA+"@203.0.113.10:26656", "")), "p2p.persistent_peers"},
		{"rpc-zero-inbound", RoleRPC, rpcMutated(replace("max_num_inbound_peers = 40", "max_num_inbound_peers = 0")), "p2p.max_num_inbound_peers"},
		{"rpc-zero-outbound", RoleRPC, rpcMutated(replace("max_num_outbound_peers = 10", "max_num_outbound_peers = 0")), "p2p.max_num_outbound_peers"},
		{"rpc-pex-off", RoleRPC, rpcMutated(replace("pex = true", "pex = false")), "p2p.pex"},
		{"rpc-seed-mode", RoleRPC, rpcMutated(replace("seed_mode = false", "seed_mode = true")), "p2p.seed_mode"},

		// Private role profile.
		{"private-pex-on", RolePrivate, privateMutated(replace("pex = false", "pex = true")), "p2p.pex"},
		{"private-seeds", RolePrivate, privateMutated(replace(
			"private_peer_ids = \"\"", "seeds = \""+peerIDC+"@203.0.113.12:26656\"\nprivate_peer_ids = \"\"")), "p2p.seeds"},
		{"private-public-p2p-listener", RolePrivate, privateMutated(replace(
			"tcp://127.0.0.1:26656", "tcp://203.0.113.30:26656")), "p2p.laddr"},
		{"private-inbound", RolePrivate, privateMutated(replace("max_num_inbound_peers = 0", "max_num_inbound_peers = 4")), "p2p.max_num_inbound_peers"},
		{"private-no-peers", RolePrivate, privateMutated(replace(peerIDA+"@203.0.113.10:26656", "")), "p2p.persistent_peers"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			configTOML, appTOML := tc.configApp("")
			home, selfID := makeHome(t, configTOML, appTOML)
			configTOML, appTOML = tc.configApp(selfID)
			rewriteConfig(t, home, configTOML, appTOML)
			report := Validate(home, tc.role, Options{})
			if report.Valid {
				t.Fatalf("expected violations for %s", tc.name)
			}
			found := false
			for _, violation := range report.Violations {
				if strings.Contains(violation.Check, tc.wantCheck) {
					found = true
				}
				if strings.Contains(violation.Message, home) {
					t.Errorf("violation message leaks home path: %q", violation.Message)
				}
			}
			if !found {
				t.Fatalf("no violation with check %q in %+v", tc.wantCheck, report.Violations)
			}
		})
	}
}

func goodValidatorConfigForPeers(selfID, peers string) (string, string) {
	config := p2pSection(peers, "", peerIDA, false, false, 0, 10) +
		rpcSection(goodRPC, false, "", "")
	app := appSection(false, false, false, false, loopbackGW, loopbackGR)
	return config, app
}

func sentryMutated(mutate func(*string, *string)) func(string) (string, string) {
	return func(string) (string, string) {
		config, app := goodSentryConfig()
		mutate(&config, &app)
		return config, app
	}
}

func seedMutated(mutate func(*string, *string)) func(string) (string, string) {
	return func(string) (string, string) {
		config, app := goodSeedConfig()
		mutate(&config, &app)
		return config, app
	}
}

func rpcMutated(mutate func(*string, *string)) func(string) (string, string) {
	return func(string) (string, string) {
		config, app := goodRPCConfig()
		mutate(&config, &app)
		return config, app
	}
}

func privateMutated(mutate func(*string, *string)) func(string) (string, string) {
	return func(string) (string, string) {
		config, app := goodPrivateConfig()
		mutate(&config, &app)
		return config, app
	}
}

func TestValidateAllowLocalPeersTestOnly(t *testing.T) {
	peers := peerIDA + "@127.0.0.1:26656"
	config := p2pSection(peers, "", "", false, false, 0, 10) +
		rpcSection(goodRPC, false, "", "")
	app := appSection(false, false, false, false, loopbackGW, loopbackGR)
	home, _ := makeHome(t, config, app)

	strict := Validate(home, RolePrivate, Options{})
	if strict.Valid {
		t.Fatal("loopback peers accepted without the test-only flag")
	}
	relaxed := Validate(home, RolePrivate, Options{AllowLocalPeers: true})
	if !relaxed.Valid {
		t.Fatalf("loopback peers rejected with test-only flag: %+v", relaxed.Violations)
	}
}

func TestValidateMissingHomeFailsClosed(t *testing.T) {
	report := Validate(filepath.Join(t.TempDir(), "does-not-exist"), RoleValidator, Options{})
	if report.Valid {
		t.Fatal("uninitialized home passed validation")
	}
	checks := map[string]bool{}
	for _, violation := range report.Violations {
		checks[violation.Check] = true
	}
	for _, want := range []string{"config.toml", "app.toml", "node_key.json"} {
		if !checks[want] {
			t.Fatalf("missing %s violation in %+v", want, report.Violations)
		}
	}
}

func TestValidateInvalidLibraryRoleFailsClosed(t *testing.T) {
	config, app := goodPrivateConfig()
	home, _ := makeHome(t, config, app)
	report := Validate(home, Role("operator"), Options{})
	if report.Valid {
		t.Fatal("invalid library role passed validation")
	}
	for _, violation := range report.Violations {
		if violation.Check == "role" {
			return
		}
	}
	t.Fatalf("missing role violation in %+v", report.Violations)
}

func TestValidatePrometheusLoopbackPasses(t *testing.T) {
	config, app := goodPrivateConfig()
	config = strings.Replace(config, "[p2p]\n",
		"[instrumentation]\nprometheus = true\nprometheus_listen_addr = \"127.0.0.1:26660\"\n\n[p2p]\n", 1)
	home, _ := makeHome(t, config, app)
	report := Validate(home, RolePrivate, Options{})
	if !report.Valid {
		t.Fatalf("loopback Prometheus listener rejected: %+v", report.Violations)
	}
}

func TestValidateMalformedTOMLFailsClosed(t *testing.T) {
	home, _ := makeHome(t, "this is [ not toml", "[api]\nenable = false")
	report := Validate(home, RoleValidator, Options{})
	if report.Valid {
		t.Fatal("malformed config.toml passed validation")
	}
	found := false
	for _, violation := range report.Violations {
		if violation.Check == "config.toml" {
			found = true
		}
		if strings.Contains(violation.Message, home) || strings.Contains(violation.Message, "this is [ not toml") {
			t.Fatalf("parse violation leaks path or file contents: %q", violation.Message)
		}
	}
	if !found {
		t.Fatalf("no config.toml violation in %+v", report.Violations)
	}
}

func TestParseRole(t *testing.T) {
	for _, role := range []Role{RoleSeed, RoleSentry, RoleValidator, RoleRPC, RolePrivate} {
		parsed, err := ParseRole(string(role))
		if err != nil || parsed != role {
			t.Fatalf("ParseRole(%q) = %q, %v", role, parsed, err)
		}
	}
	if _, err := ParseRole("operator"); err == nil {
		t.Fatal("unknown role accepted")
	}
	if _, err := ParseRole(""); err == nil {
		t.Fatal("empty role accepted")
	}
}
