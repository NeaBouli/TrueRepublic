# TrueRepublic GH-85 — Audit
> Scope: monitoring configuration, Grafana provisioning/dashboard, Prometheus rules/tests, CI runtime gates, operator docs, wiki mirrors, repository contract tests  ·  Date: 2026-07-30  ·  Result: 0 FAIL / 0 WARN / 7 PASS

## Summary

GH-85 provides a coherent private-testnet observability baseline: a file-provisioned
Grafana dashboard, eleven actionable Prometheus rules, deterministic rule fixtures,
role-based escalation ownership, and explicitly non-production service objectives.
The first independent review found one medium and five low issues; all were remediated
before this audit and the focused Kimi K3 recheck found no remaining P0, P1, or P2
defects. Local Prometheus validation, repository tests, the full race/coverage gate,
and all eight multi-validator recovery scenarios pass. Docker is unavailable on the
review host, so the pinned Compose runtime proof remains a mandatory exact-head GitHub
check before merge.

## Findings by domain

### Signal isolation and correctness — PASS

- **[🟡 MEDIUM, remediated] Validator signals remain paired per node** — `monitoring/prometheus-alerts.yml`
  - What: Early review identified cross-instance aggregation that could combine consensus and application signals from different validators.
  - Path: Two validators with divergent heights could be aggregated independently, producing a reassuring value that did not describe either validator.
  - Fix: Invariant and PNYX arithmetic now use per-instance vector matching; divergence pairs the two scrape jobs with a bounded static `node` label and `on (node)`. Repository and promtool tests pin the contract.

### Alert behavior — PASS

- **[🟢 LOW, remediated] All eleven rules have deterministic lifecycle coverage** — `monitoring/prometheus-alerts.test.yml`
  - What: The required-metric rule initially lacked explicit recovery evidence.
  - Path: A rule could fire correctly but remain active after the missing series returned without a fixture detecting the regression.
  - Fix: Fixtures now prove inactive, firing, and recovered states and cover all eleven alert names.

### Dashboard and datasource provisioning — PASS

- **[🟢 LOW, remediated] Runtime and structural gates are exact** — `monitoring/grafana/dashboards/blockchain.json`
  - What: Initial checks accepted a panel-count floor and did not prove the datasource could execute a query.
  - Path: An accidentally duplicated or empty dashboard could pass, or a provisioned datasource could exist but be unusable.
  - Fix: CI and repository tests require exactly 16 panels and 17 PromQL targets, verify the stable datasource UID, and execute a live query through Grafana.

### Supply integrity — PASS

- **[🔴 BLOCKING, protected] PNYX fixed-cap arithmetic is explicit and per instance** — `monitoring/prometheus-alerts.yml`
  - What: Supply overflow, negative headroom, or inconsistent supply-plus-headroom must never be hidden by aggregation.
  - Path: A cap violation on one instance could otherwise be offset by a healthy value from another instance.
  - Fix: The critical rule evaluates each application target directly against 21,000,000,000,000 base units and has deterministic firing coverage.

### CI and reproducibility — PASS

- **[🟠 HIGH, protected] Monitoring images and validation tools are immutable** — `docker-compose.yml`
  - What: Mutable image references could change monitoring behavior without a repository diff.
  - Path: A later registry update could introduce an incompatible Prometheus or Grafana runtime into the same commit.
  - Fix: Prometheus 3.13.1 and Grafana 13.1.1 are pinned by tag and multi-architecture digest; CI uses the same Prometheus digest for config and rule validation.

### Operational ownership — PASS

- **[🟠 HIGH, protected] Alerts name roles without claiming paging readiness** — `docs/node-operators/operations/monitoring.md`
  - What: Alert rules without accountable ownership or a tested delivery path are not operational alerts.
  - Path: A critical condition could fire while every operator assumes another person owns acknowledgement.
  - Fix: Every rule has a role owner and severity; acknowledgement targets and escalation roles are documented. No email, chat, webhook, PagerDuty, or other Alertmanager destination is claimed or configured, and a paging drill remains a rollout gate.

### Documentation and public claims — PASS

- **[🟢 LOW, remediated] Current documentation uses one honest contract** — `wiki/operations/Monitoring.md`
  - What: Historical ten-rule wording and ambiguous image-pin language conflicted with the implemented eleven-rule baseline.
  - Path: Operators could follow stale instructions or mistake upstream image pinning for complete project release provenance.
  - Fix: The latest append-only Bridge entry supersedes the historical snapshot; canonical and wiki docs state eleven rules, distinguish upstream pins from project release artifacts, and label objectives as private-testnet targets rather than production SLOs.

## Priority matrix

### 🔴 BLOCKING

None open.

### 🟠 HIGH

None open.

### 🟡 MEDIUM

None open.

### 🟢 LOW

None open. The deprecated Grafana datasource-proxy compatibility route is intentionally
kept behind an exact runtime CI assertion; any removal by the pinned Grafana release
will fail the pull request instead of silently degrading evidence.

## Verification boundary

- Local: Prometheus configuration and eleven rules valid; deterministic promtool tests pass.
- Local: dashboard JSON, provisioning YAML, workflow YAML, and every dashboard/objective PromQL expression parse.
- Local: `make build` and `make verify` pass, including Go race detection and coverage.
- Local: all eight maintained multi-validator recovery scenarios pass.
- Independent review: Kimi K3 remediation recheck reports 0 P0 / 0 P1 / 0 P2.
- Pending before merge: exact-head GitHub Docker Compose runtime checks, required reviews, and all repository branch-protection gates.
