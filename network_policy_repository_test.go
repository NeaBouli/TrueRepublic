package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContainerNetworkDefaultsFailClosed(t *testing.T) {
	t.Parallel()

	dockerfile := readRepositoryFile(t, "Dockerfile")
	for _, required := range []string{
		"EXPOSE 26656",
		"--rpc.laddr=tcp://127.0.0.1:26657",
		"--grpc.enable=false",
		"--api.enable=false",
		"--log_format=json",
		`CMD ["truerepublicd", "healthcheck", "live", "--timeout", "2s"]`,
	} {
		if !strings.Contains(dockerfile, required) {
			t.Fatalf("Dockerfile must contain %q", required)
		}
	}
	for _, forbidden := range []string{
		"EXPOSE 26656 26657",
		"--rpc.laddr=tcp://0.0.0.0:26657",
		"--api.enable=true",
	} {
		if strings.Contains(dockerfile, forbidden) {
			t.Fatalf("Dockerfile contains unsafe default %q", forbidden)
		}
	}

	compose := readRepositoryFile(t, "docker-compose.yml")
	for _, required := range []string{
		`127.0.0.1:${P2P_PORT:-26656}:26656`,
		"127.0.0.1:8080:80",
		"127.0.0.1:9091:9090",
		"127.0.0.1:3000:3000",
		"--rpc.laddr=tcp://127.0.0.1:26657",
		"--api.address=tcp://127.0.0.1:1317",
		"--grpc.enable=false",
		"--log_format=json",
		`network_mode: "service:truerepublic-node"`,
		"PROMETHEUS_ENABLED=${PROMETHEUS_ENABLED:-false}",
		"${GRAFANA_PASSWORD:?set GRAFANA_PASSWORD}",
	} {
		if !strings.Contains(compose, required) {
			t.Fatalf("docker-compose.yml must contain %q", required)
		}
	}
	if strings.Contains(compose, "${GRAFANA_PASSWORD:-admin}") {
		t.Fatal("docker-compose.yml must not provide a default Grafana admin password")
	}
	prometheus := readRepositoryFile(t, "monitoring/prometheus.yml")
	if !strings.Contains(prometheus, "targets: ['127.0.0.1:26660']") {
		t.Fatal("Prometheus must scrape the node's loopback metrics listener from the shared namespace")
	}
	for _, required := range []string{
		"job_name: 'truerepublic-app'",
		"metrics_path: '/metrics'",
		"format: ['prometheus']",
		"targets: ['127.0.0.1:1317']",
		"node: 'local'",
	} {
		if !strings.Contains(prometheus, required) {
			t.Fatalf("Prometheus application scrape must contain %q", required)
		}
	}
	if strings.Count(prometheus, "node: 'local'") != 2 {
		t.Fatal("both local Prometheus targets must share exactly one bounded node pairing label")
	}
	if strings.Contains(prometheus, "chain_id: 'truerepublic-1'") {
		t.Fatal("Prometheus must not attach a false static chain ID to operator-selected chains")
	}
	entrypoint := readRepositoryFile(t, "scripts/docker-entrypoint.sh")
	for _, required := range []string{
		"TRUEREPUBLICD_TELEMETRY_ENABLED=true",
		"TRUEREPUBLICD_TELEMETRY_ENABLED=false",
		"TRUEREPUBLICD_TELEMETRY_PROMETHEUS_RETENTION_TIME=60",
		"TRUEREPUBLICD_TELEMETRY_PROMETHEUS_RETENTION_TIME=0",
		"TRUEREPUBLICD_TELEMETRY_SERVICE_NAME=truerepublic",
		"TRUEREPUBLICD_TELEMETRY_ENABLE_HOSTNAME=false",
		"TRUEREPUBLICD_TELEMETRY_ENABLE_HOSTNAME_LABEL=false",
		"TRUEREPUBLICD_TELEMETRY_ENABLE_SERVICE_LABEL=false",
	} {
		if !strings.Contains(entrypoint, required) {
			t.Fatalf("container telemetry policy must contain %q", required)
		}
	}
	datasource := readRepositoryFile(t, "monitoring/grafana/provisioning/datasources/datasource.yml")
	if !strings.Contains(datasource, "url: http://truerepublic-node:9090") {
		t.Fatal("Grafana must reach Prometheus through the node's shared network endpoint")
	}

	nginx := readRepositoryFile(t, "nginx/nginx.conf")
	for _, forbidden := range []string{
		`Access-Control-Allow-Origin "*"`,
		"listen 443",
	} {
		if strings.Contains(nginx, forbidden) {
			t.Fatalf("local nginx configuration contains unsupported exposure %q", forbidden)
		}
	}
	for _, required := range []string{
		"limit_req zone=rpc burst=20 nodelay",
		"limit_req zone=api burst=50 nodelay",
		"location ~ ^/api/metrics(?:/|$)",
		"return 404",
		"client_max_body_size 1m",
		"proxy_connect_timeout 5s",
		"server 127.0.0.1:26657",
		"server 127.0.0.1:1317",
	} {
		if !strings.Contains(nginx, required) {
			t.Fatalf("local nginx configuration must contain %q", required)
		}
	}

	workflow := readRepositoryFile(t, ".github/workflows/go-ci.yml")
	if strings.Contains(workflow, "--publish") {
		t.Fatal("Docker smoke must not publish a host RPC port")
	}
	for _, required := range []string{
		"truerepublicd healthcheck live --timeout 2s",
		"truerepublicd healthcheck ready --timeout 2s",
		"docker exec truerepublic-node-ci wget -qO- http://127.0.0.1:26657/status",
		"Verify structured logs after restart",
		"docker compose up --detach --build truerepublic-node web-wallet nginx prometheus grafana",
		"- 'nginx/**'",
		"http://127.0.0.1:8080/rpc/status",
		`select(.labels.job == "truerepublic-node" and .health == "up")`,
		`select(.labels.job == "truerepublic-app" and .health == "up")`,
		"http://127.0.0.1:8080/api/metrics?format=prometheus",
		`metrics_proxy_blocked=true`,
		"http://127.0.0.1:8080/api/metrics/",
		`metrics_descendant_proxy_blocked=true`,
		"Verify metrics contract and restart progression",
		"name == metric",
		"cometbft_consensus_block_interval_seconds_sum",
		"cometbft_consensus_block_interval_seconds_count",
		"truerepublic_app_last_successful_block_height",
		"truerepublic_app_last_successful_invariant_cycle_height",
		"truerepublic_token_pnyx_supply_base_units",
		"truerepublic_token_pnyx_supply_headroom_base_units",
		"http://127.0.0.1:3000/api/health",
		"docker compose down --volumes --remove-orphans",
	} {
		if !strings.Contains(workflow, required) {
			t.Fatalf("Docker smoke must exercise the loopback Compose stack with %q", required)
		}
	}

	monitoring := readRepositoryFile(t, "docs/node-operators/operations/monitoring.md")
	for _, required := range []string{
		"truerepublicd healthcheck live",
		"truerepublicd healthcheck ready",
		"does not restart a healthy syncing node",
		"literal loopback address only",
		"`truerepublic-node`",
		"`truerepublic-app`",
		"`truerepublic_app_last_successful_invariant_cycle_height`",
		"`truerepublic_token_pnyx_supply_headroom_base_units`",
		"does not prove a production topology",
		"Initial Service Objectives",
		"Escalation Ownership",
		"does not configure email, chat, webhook, PagerDuty",
		"not a production SLO commitment",
		"Every supported node log line is JSON",
		"[REDACTED]",
		"rejects `--trace-store`",
	} {
		if !strings.Contains(monitoring, required) {
			t.Fatalf("operator monitoring guide must explain health semantics with %q", required)
		}
	}
	deploymentWiki := readRepositoryFile(t, "wiki/operations/Deployment-Options.md")
	if !strings.Contains(deploymentWiki, "- healthcheck\n            - live") ||
		!strings.Contains(deploymentWiki, "- healthcheck\n            - ready") {
		t.Fatal("Kubernetes example must use distinct exec liveness/readiness probes")
	}
}

func TestConfigIndependentRootCommandsDoNotInitializeHome(t *testing.T) {
	t.Parallel()

	type testCase struct {
		home string
		args []string
	}
	policyHome := filepath.Join(t.TempDir(), "policy-home")
	healthHome := filepath.Join(t.TempDir(), "health-home")
	tests := map[string]testCase{
		"network policy": {
			home: policyHome,
			args: []string{
				"network-policy", "validate",
				"--role", "validator",
				"--home", policyHome,
				"--output", "json",
			},
		},
		"healthcheck": {
			home: healthHome,
			args: []string{
				"--home", healthHome,
				"healthcheck", "live",
				"--rpc-url", "https://127.0.0.1:26657",
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			root := newRootCmd()
			output := new(bytes.Buffer)
			root.SetOut(output)
			root.SetErr(output)
			root.SetArgs(tc.args)
			if err := root.Execute(); err == nil {
				t.Fatal("invalid read-only command unexpectedly succeeded")
			}
			if _, err := os.Stat(tc.home); !os.IsNotExist(err) {
				t.Fatalf("read-only command created or touched its target home: %v", err)
			}
			if strings.Contains(output.String(), tc.home) {
				t.Fatalf("command output leaked the local home path: %s", output)
			}
		})
	}
}

func readRepositoryFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}
