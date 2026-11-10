# GH364 refresh triage — 2026-11-10 (Claude Code)

Base: origin/agent/codex/GH364-go-advisories @ 6d1c79d (#365). Branch agent/claude/GH364-refresh-20261110.
Local: `GOVULNCHECK_BIN=~/go/bin/govulncheck ./scripts/check-go-vulnerabilities.sh` (go1.26.9, govulncheck v1.6.0) -> exit 1.

Reachable IDs:
| ID | Package | Fix | Action |
|---|---|---|---|
| GO-2026-6809 | stdlib crypto/x509 | go1.26.10 | bump toolchain go1.26.9 -> go1.26.10 (go.mod, gates.json, CI setup-go) |
| GO-2026-6865 | github.com/klauspost/compress (fse/huff0) | v1.20.2 | `go get github.com/klauspost/compress@v1.20.2`, go mod tidy, full Go suite |
| GO-2026-5932 | golang.org/x/crypto/openpgp | x/crypto pseudo 0.58.1-0.20261009181718 (openpgp subpkg: no fix) | re-evaluate; allowance expired 2026-10-20 |
| GO-2023-1881 | cosmos-sdk x/crisis ConstantFee | none | re-evaluate no-fix allowance with fresh date (owner/lead decision) |
| GO-2023-1821 | cosmos-sdk x/crisis halt | none | same |

Not yet done: rust-audit (rustls >=0.23.46, ruzstd >=0.9.1), node-audit-client high advisory, required verification suite.
No allowance renewed silently; no merge.
