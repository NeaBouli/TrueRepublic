# GH-364 — pinned build/security maintenance boundary

Owner authorization: Gio, 2026-10-10 Europe/Athens. Base ff63e9d; no product feature or chain-state
change. This scoped map extends the existing build/release foundation only; new implementation
paths, flags, wrappers and protocol changes are out of scope.

| Hop | Opened source edge | Bound datum / required change |
| --- | --- | --- |
| B1 | `.github/workflows/security-scan.yml` setup-go -> `scripts/check-go-vulnerabilities.sh` -> `scripts/go-packages.sh` -> govulncheck -> `configs/security/gates.json` | exact Go toolchain, maintained package set, reachable IDs, bounded no-fix policy; no scanner-policy weakening |
| B2 | `.github/workflows/reproducible-daemon.yml` -> `configs/build/deterministic-linux-daemon.json` -> `scripts/verify-deterministic-daemon.sh` -> `releaseevidence/verify.go::exactBuildContract` / `candidateevidence/verify.go` | exact Go version, unchanged source-ref/build flags/upgrade-plan and artifact evidence |
| B3 | `Dockerfile` pinned builder -> `go.mod`/`go.sum` -> readonly daemon build -> `configs/release/tool-platform.json` / `configs/build/reproducible-oci.json` -> `releaseevidence/verify.go::exactImages` | official multi-arch Go image digest; wasmvm remains 2.2.8; no unpinned base image |
| B4 | `.github/workflows/react-ci.yml` -> `scripts/test-zkp-wasm-client.sh` -> `cmd/zkp-prover-wasm` | exact Go version and synthetic WASM/native/keeper compatibility; artifacts remain test-only |
| B5 | binary/OCI contracts -> `configs/release/candidate-evidence.json` -> `configs/release/cross-run-rebuild.json` -> corresponding testdata/verifiers | deliberate digest propagation to the fixed point, never stale or ignored pins |
| D1 | verified dependency/CI evidence -> `docs/status.json::features.dependency_security` / `configs/security/threat-model.json` -> `docs/index.html` static status/Phase-5 card, Go badge and `docs/security/THREAT_MODEL.md` | truthful candidate/merged distinction and bounded residual risk; four-viewport render with no layout or rollout expansion |

The module graph update is minimal x/net 0.58.0 -> 0.60.0 plus only dependencies required by it.
Go remains in the supported 1.26 family (1.26.9). Change the coupled CI, Docker, strict verifiers,
fixtures and maintained toolchain documentation together; retain negative/mutation test strength.
Historical Bridge/audit narratives are not mechanically rewritten. Public readiness/test/rollout
claims remain uninflated. Visual inspection remains required for any rendered Landing badge update.

No modification to app composition, governance, token cap, custody, ZKP circuit/ceremony, wasmvm,
IBC state, live operators, production/Mainnet or deferred V3/V4. #355 permits evidence-backed
existing exception renewal/removal, not adding IDs: withdrawn GO-2026-4740 is removed; the three
remaining records require a fresh exact-scan no-fix result before a bounded renewal is recorded.
Codex Security NOT RUN until separately authorized; other safe maintenance proceeds.
