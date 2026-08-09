# TrueRepublic security policy

TrueRepublic v0.4 is a recovery project. It is not approved for production,
mainnet, real keys, or real funds. The remaining rollout gates are tracked in
[GH-29](https://github.com/NeaBouli/TrueRepublic/issues/29) and
[`docs/ROLLOUT_ROADMAP.md`](docs/ROLLOUT_ROADMAP.md).

## Reporting a vulnerability

Do not publish exploitable details, credentials, private keys, mnemonics, or
private infrastructure data in a public issue. Use GitHub's private
vulnerability-reporting flow for this repository when it is available. If that
flow is unavailable, open a public issue containing only a request for a private
contact channel and no sensitive detail.

The repository does not currently promise a bug-bounty payment or production
response SLA. Acknowledgement, triage, remediation evidence, and disclosure
timing must be agreed for each report.

## Maintained automated gates

The authoritative policy is
[`configs/security/gates.json`](configs/security/gates.json). The Security Scan
workflow runs for every pull request and push to `main`, on a weekly schedule,
and by manual dispatch. It blocks on:

- reachable Go vulnerabilities, except exact active no-fix IDs in the central
  policy;
- Rust advisories and maintained-client high/critical advisories;
- Go `staticcheck`, Rust `clippy`, and maintained-client lint/type checks;
- maintained-tree secret scanning, including a planted-credential failure test;
- the repository security-gate contract and dependency lock integrity;
- retired client and custom-query surfaces reappearing.

All third-party GitHub Actions use reviewed immutable commit SHAs. Scanner and
toolchain versions are exact values in the policy contract. Dependabot opens
bounded weekly grouped updates for GitHub Actions, Go modules, Cargo, and the
maintained npm client.

## Local verification

Install the exact tool versions from `configs/security/gates.json`, then run:

```bash
go test . -run '^TestSecurityGateRepositoryContract$' -count=1
STATICCHECK_BIN=/path/to/staticcheck ./scripts/check-static-analysis.sh
GITLEAKS_BIN=/path/to/gitleaks ./scripts/check-secret-scan.sh
GITLEAKS_BIN=/path/to/gitleaks ./scripts/test-secret-scan.sh
GOVULNCHECK_BIN=/path/to/govulncheck ./scripts/check-go-vulnerabilities.sh
./scripts/test-go-vulnerability-scan.sh

cd contracts
cargo fmt --check
cargo clippy --locked --workspace -- -D warnings
cargo build --locked --workspace
cargo test --locked --workspace
cargo audit

cd ../client-web
npm ci
npm run lint
npm test -- --run
npm run build
npm run audit:high
```

The complete Go integration baseline remains `make verify`. A green local run is
not a substitute for protected exact-head pull-request checks and final `main`
verification.

## Update and exception rules

Security dependencies are reviewed weekly. Updates must change the central
contract and immutable workflow pins together, pass the planted negative
fixtures, and receive normal pull-request review. Never replace a pin with a
mutable tag or `latest` reference.

An exception must identify the exact package or rule, affected surface, owner,
reason, compensating test, upstream reference, and an expiry no more than 30
days away. Broad path, commit, or stopword exclusions are forbidden for secret
scanning. Expired exceptions fail closed and must not be renewed without fresh
evidence.

The Go gate permits only the exact reachable IDs listed in the central policy,
only while they have no published fixed version and their dated exception is
active. A new, fixable, expired, missing, malformed, or stale finding/exception
fails the gate. Scanner execution and report parsing errors also fail closed.

Moderate/low npm findings and non-vulnerability RustSec warnings may remain
visible without blocking only when the maintained policy explicitly classifies
them. Critical/high findings, scanner errors, and malformed reports fail the
gate. Independent cryptographic/privacy review, release signing, SBOM,
provenance, production topology, and rollout approval remain separate GH-29
requirements.
