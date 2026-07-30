package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

const grafanaPrometheusUID = "truerepublic-prometheus"

func TestMonitoringArtifactsAreProvisionedAndBounded(t *testing.T) {
	t.Parallel()

	var raw map[string]json.RawMessage
	dashboardJSON := readRepositoryFile(t, "monitoring/grafana/dashboards/blockchain.json")
	if err := json.Unmarshal([]byte(dashboardJSON), &raw); err != nil {
		t.Fatalf("decode Grafana dashboard: %v", err)
	}
	for _, apiWrapperField := range []string{"dashboard", "overwrite"} {
		if _, exists := raw[apiWrapperField]; exists {
			t.Fatalf("file-provisioned dashboard must not contain API wrapper field %q", apiWrapperField)
		}
	}

	var dashboard struct {
		UID           string `json:"uid"`
		Title         string `json:"title"`
		SchemaVersion int    `json:"schemaVersion"`
		Version       int    `json:"version"`
		Editable      bool   `json:"editable"`
		Panels        []struct {
			ID         int    `json:"id"`
			Title      string `json:"title"`
			Type       string `json:"type"`
			Datasource struct {
				UID string `json:"uid"`
			} `json:"datasource"`
			Targets []struct {
				RefID      string `json:"refId"`
				Expression string `json:"expr"`
				Datasource struct {
					UID string `json:"uid"`
				} `json:"datasource"`
			} `json:"targets"`
		} `json:"panels"`
	}
	if err := json.Unmarshal([]byte(dashboardJSON), &dashboard); err != nil {
		t.Fatalf("decode native Grafana dashboard: %v", err)
	}
	if dashboard.UID != "truerepublic-blockchain" || dashboard.Title == "" {
		t.Fatalf("unexpected dashboard identity: uid=%q title=%q", dashboard.UID, dashboard.Title)
	}
	if dashboard.SchemaVersion <= 0 || dashboard.Version <= 0 {
		t.Fatalf("dashboard schema/version must be explicit: schema=%d version=%d", dashboard.SchemaVersion, dashboard.Version)
	}
	if dashboard.Editable {
		t.Fatal("provisioned dashboard must be immutable to prevent UI drift")
	}
	if len(dashboard.Panels) != 16 {
		t.Fatalf("dashboard must expose the reviewed operations surface; got %d panels", len(dashboard.Panels))
	}

	panelIDs := make(map[int]string, len(dashboard.Panels))
	var expressions strings.Builder
	for _, panel := range dashboard.Panels {
		if panel.ID <= 0 || panel.Title == "" || panel.Type == "" {
			t.Fatalf("panel has incomplete identity: %+v", panel)
		}
		if prior, exists := panelIDs[panel.ID]; exists {
			t.Fatalf("panels %q and %q reuse id %d", prior, panel.Title, panel.ID)
		}
		panelIDs[panel.ID] = panel.Title
		if panel.Type == "row" {
			continue
		}
		if len(panel.Targets) == 0 {
			t.Fatalf("panel %q has no Prometheus target", panel.Title)
		}
		refIDs := make(map[string]struct{}, len(panel.Targets))
		for _, target := range panel.Targets {
			if target.RefID == "" || target.Expression == "" {
				t.Fatalf("panel %q has an incomplete target: %+v", panel.Title, target)
			}
			if _, exists := refIDs[target.RefID]; exists {
				t.Fatalf("panel %q reuses refId %q", panel.Title, target.RefID)
			}
			refIDs[target.RefID] = struct{}{}
			datasourceUID := target.Datasource.UID
			if datasourceUID == "" {
				datasourceUID = panel.Datasource.UID
			}
			if datasourceUID != grafanaPrometheusUID {
				t.Fatalf("panel %q target %q uses datasource %q", panel.Title, target.RefID, datasourceUID)
			}
			fmt.Fprintln(&expressions, target.Expression)
		}
	}

	expressionSet := expressions.String()
	for _, metric := range []string{
		"cometbft_consensus_height",
		"cometbft_p2p_peers",
		"cometbft_mempool_size",
		"cometbft_consensus_missing_validators",
		"truerepublic_app_last_successful_block_height",
		"truerepublic_app_completed_blocks_total",
		"truerepublic_app_last_successful_invariant_cycle_height",
		"truerepublic_token_pnyx_supply_base_units",
		"truerepublic_token_pnyx_supply_headroom_base_units",
		"go_goroutines",
		"process_resident_memory_bytes",
	} {
		if !strings.Contains(expressionSet, metric) {
			t.Fatalf("dashboard is missing reviewed metric %q", metric)
		}
	}

	prometheusConfig := readRepositoryFile(t, "monitoring/prometheus.yml")
	if !strings.Contains(prometheusConfig, "/etc/prometheus/prometheus-alerts.yml") {
		t.Fatal("Prometheus config must load the repository alert rules")
	}
	if strings.Count(prometheusConfig, "node: 'local'") != 2 {
		t.Fatal("both local scrape jobs must share the bounded static node label")
	}
	alerts := readRepositoryFile(t, "monitoring/prometheus-alerts.yml")
	for _, required := range []string{
		"groups:",
		"severity:",
		"owner:",
		"summary:",
		"description:",
		"runbook_url:",
		"truerepublic-node",
		"truerepublic-app",
		"21000000000000",
	} {
		if !strings.Contains(alerts, required) {
			t.Fatalf("alert contract must contain %q", required)
		}
	}
	for _, alertName := range []string{
		"TrueRepublicNodeTargetDown",
		"TrueRepublicApplicationTargetDown",
		"TrueRepublicRequiredMetricMissing",
		"TrueRepublicConsensusStalled",
		"TrueRepublicApplicationStalled",
		"TrueRepublicApplicationConsensusDivergence",
		"TrueRepublicInvariantLag",
		"TrueRepublicMissingValidators",
		"TrueRepublicLowPeerCount",
		"TrueRepublicMempoolPressure",
		"TrueRepublicPNYXCapIntegrityBreach",
	} {
		if !strings.Contains(alerts, "- alert: "+alertName) {
			t.Fatalf("alert contract is missing %q", alertName)
		}
	}
	if strings.Count(alerts, "\n      - alert:") != 11 ||
		strings.Count(alerts, "\n        for:") != 11 {
		t.Fatal("every reviewed alert must have a bounded activation window")
	}
	if !strings.Contains(alerts, "on (node)") {
		t.Fatal("consensus/application divergence must pair sources by the bounded node label")
	}

	ruleTests := readRepositoryFile(t, "monitoring/prometheus-alerts.test.yml")
	for _, required := range []string{
		"rule_files:",
		"evaluation_interval:",
		"alert_rule_test:",
		"exp_alerts:",
	} {
		if !strings.Contains(ruleTests, required) {
			t.Fatalf("promtool test contract must contain %q", required)
		}
	}

	datasource := readRepositoryFile(t, "monitoring/grafana/provisioning/datasources/datasource.yml")
	for _, required := range []string{
		"uid: " + grafanaPrometheusUID,
		"url: http://truerepublic-node:9090",
		"editable: false",
	} {
		if !strings.Contains(datasource, required) {
			t.Fatalf("Grafana datasource must contain %q", required)
		}
	}
	provider := readRepositoryFile(t, "monitoring/grafana/provisioning/dashboards/dashboard.yml")
	for _, required := range []string{"disableDeletion: true", "editable: false"} {
		if !strings.Contains(provider, required) {
			t.Fatalf("Grafana provider must contain %q", required)
		}
	}

	compose := readRepositoryFile(t, "docker-compose.yml")
	for _, required := range []string{
		"image: prom/prometheus:v3.13.1@sha256:3c42b892cf723fa54d2f262c37a0e1f80aa8c8ddb1da7b9b0df9455a35a7f893",
		"image: grafana/grafana:13.1.1@sha256:7cb8c64c4d57a57e734073f3cc94620adb24a0acb929bd80ba9f14017e3a975b",
		"./monitoring/prometheus-alerts.yml:/etc/prometheus/prometheus-alerts.yml:ro",
		"./monitoring/grafana/dashboards:/var/lib/grafana/dashboards:ro",
		"./monitoring/grafana/provisioning:/etc/grafana/provisioning:ro",
	} {
		if !strings.Contains(compose, required) {
			t.Fatalf("Compose monitoring contract must contain %q", required)
		}
	}

	workflow := readRepositoryFile(t, ".github/workflows/go-ci.yml")
	for _, required := range []string{
		"Validate Prometheus configuration and alert behavior",
		"prom/prometheus:v3.13.1@sha256:3c42b892cf723fa54d2f262c37a0e1f80aa8c8ddb1da7b9b0df9455a35a7f893",
		"test rules /etc/prometheus/prometheus-alerts.test.yml",
		"Verify provisioned dashboard, rules, and panel queries",
		"/api/dashboards/uid/truerepublic-blockchain",
		"/api/datasources/uid/truerepublic-prometheus",
		"/api/datasources/proxy/uid/truerepublic-prometheus/api/v1/query",
		"/api/v1/rules?type=alert",
		`--data-urlencode "query=${expression}"`,
	} {
		if !strings.Contains(workflow, required) {
			t.Fatalf("GitHub runtime evidence must contain %q", required)
		}
	}

	monitoringWiki := readRepositoryFile(t, "wiki/operations/Monitoring.md")
	normalizedMonitoringWiki := strings.Join(strings.Fields(monitoringWiki), " ")
	for _, required := range []string{
		"eleven reviewed rules",
		"other Alertmanager destination",
		"not a production SLO commitment",
		"21,000,000,000,000 base units",
	} {
		if !strings.Contains(normalizedMonitoringWiki, required) {
			t.Fatalf("monitoring wiki must contain %q", required)
		}
	}
	for _, forbidden := range []string{
		"prom/prometheus:latest",
		"grafana/grafana:latest",
		"tendermint_consensus",
		"monitoring/alerts.yml",
		"alertmanager:9093",
	} {
		if strings.Contains(monitoringWiki, forbidden) {
			t.Fatalf("monitoring wiki contains obsolete contract %q", forbidden)
		}
	}

	installationWizard := readRepositoryFile(t, "wiki/users/Installation-Wizards.md")
	for _, forbidden := range []string{"monitoring/alerts.yml", "tendermint_consensus"} {
		if strings.Contains(installationWizard, forbidden) {
			t.Fatalf("installation wizard contains obsolete monitoring contract %q", forbidden)
		}
	}
}
