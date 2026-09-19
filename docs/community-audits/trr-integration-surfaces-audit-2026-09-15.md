# TrueRepublic — Integration & Surfaces Audit

- **Series:** Collateral Web3 Open Audits
- **Date:** 2026-09-15
- **Target:** NeaBouli/TrueRepublic @ `f5de5a150b9ff0edc51d68411b7829e78bcf78e6` + live `neabouli.github.io/TrueRepublic/`
- **Scope:** client-web deployment surface (nginx/Docker), RPC trust model, wiki operations instructions, external infrastructure references, landing integration
- **Method:** deep-recon agent + lead verification incl. live probes (anonymous GET/HEAD/DNS only); landing verified byte-identical to `docs/index.html`
- **Register:** TRR-23 … TRR-28 (this report) — **0 Critical / 0 High / 3 Medium / 2 Low / 1 Info**

---

## Executive summary

The maintained client is fail-closed where it matters (ZKP submission disabled by omission from the tx registry; no admin-withdrawal message type at all; mnemonics AES-GCM-encrypted; BigInt money math; bundle budget CI-gated) — but its served origin ships **without a Content-Security-Policy** on a wallet-holding page, and the public wiki teaches setups that cannot work: non-existent `.env` keys, a 404 genesis download, **NXDOMAIN** seed/state-sync hosts, wrong CLI arguments, wrong units, a fictional Discord faucet, and a Grafana admin/admin instruction that contradicts the actual compose config. The fictional `truerepublic.network` domain is used as security contact and bug-bounty repro endpoint — an unregistered domain, i.e. a squatting hazard for a security channel.

## Severity table

| ID | Severity | Title |
|----|----------|-------|
| TRR-23 | Medium | client-web served without Content-Security-Policy (or Permissions-Policy) on a wallet-holding origin |
| TRR-24 | Medium | Wiki operations pages teach broken/insecure setups (6 independent defects) |
| TRR-25 | Medium | Fictional `truerepublic.network` (NXDOMAIN) used as seeds/state-sync host, security contact and bounty repro endpoint — squatting hazard |
| TRR-26 | Low | Single-RPC trust model (`prove:false`); plaintext localhost defaults; PBKDF2-only wallet KDF; session keeps plaintext password; no unlock throttling |
| TRR-27 | Low | Landing hotlinks raw.githubusercontent images despite local copies (privacy/availability, no SRI) |
| TRR-28 | Info | Verified client/deploy strengths (fail-closed ZKP, tx-registry curation, bundle budget, encrypted mnemonics, digest-pinned non-root images, loopback compose) |

---

## TRR-23 — Medium — No Content-Security-Policy on a wallet-holding origin

**Evidence (lead-verified):** `client-web/nginx.conf:40-43,96-110` sets `X-Frame-Options: DENY`, `X-Content-Type-Options: nosniff`, `Referrer-Policy` — but **no `Content-Security-Policy`**, no `Permissions-Policy`, no HSTS (TLS termination undocumented). The origin holds AES-GCM-encrypted mnemonics in localStorage plus the plaintext ZKP identity secret (TRR-19): any XSS is full wallet compromise, and CSP is the primary available mitigation.

**Recommendation:** add a strict CSP (`default-src 'self'`; verify Vite output for inline-script needs), `Permissions-Policy`, and document the TLS termination point.

## TRR-24 — Medium — Wiki operations pages teach broken/insecure setups

**Evidence (agent-swept, lead spot-verified):**
1. `operations-Node-Setup.md`: the `.env` example sets `EXTERNAL_IP`, `RPC_PORT`, `REST_PORT`, `GRPC_PORT`, `DB_BACKEND`, `PRUNING*`, `PROMETHEUS_PORT` — **none exist in `.env.example`** (7 real keys; the repo explicitly states "Peer topology is not accepted through environment substitution", `start-node.sh:14-18`); `wget …/main/genesis.json` — no genesis.json at repo root (404); "match official checksum" — none published; seeds/snapshot/state-sync at `*.truerepublic.network` — NXDOMAIN (TRR-25); "Docker 20.10+/docker-compose" vs INSTALLATION.md's Docker 24+/Compose v2.20+; "prometheus 9090, nginx 80" vs compose's loopback 9091/8080.
2. `operations-Validator-Guide.md:269-272,350-353` + `users-Installation-Wizards.md:255`: `query staking validators`, `tx distribution withdraw-validator-commission`, `query slashing signing-info` — **`x/staking`/`x/distribution` are not mounted** (fail-closed adapters, `app.go:297,337-338`); these commands fail as written.
3. `register-validator YOUR_DOMAIN 100000000000upnyx` (wiki, 2 args) vs CLI `register-validator [pubkey-hex] [stake] [domain]` (`cobra.ExactArgs(3)`, `x/truedemocracy/cli.go:150-152`, lead-verified).
4. Unit error ×10⁶: "Should show 100,000+ upnyx" (`operations-Validator-Guide.md:160`) — 100,000 PNYX = 100,000,000,000 upnyx (the same guide gets it right at `:202`).
5. `users-Installation-Wizards.md:151-154`: "visit `http://your-server-ip:3000`, Password: admin (change immediately)" — compose **requires** `GRAFANA_PASSWORD` (fails unset) and binds 127.0.0.1:3000; admin/admin never works and the URL is unreachable remotely.
6. `users-Installation-Wizards.md:48-54`: a **fictional Discord faucet** ("Join TrueRepublic Discord → #faucet → /faucet → 10,000 PNYX") — the only Discord mention anywhere; QUICKSTART explicitly says examples "are not a faucet flow".
7. `operations-Deployment-Options.md:335-351`: K8s `Service type: LoadBalancer` publishing RPC 26657/REST 1317/gRPC 9090 — contradicting the same page's "never publish the upstream ports directly" (`:246`) and the repo network policy; image `ghcr.io/neabouli/truerepublic:latest` is not anonymously pullable (project publishes no images per GH-212/258); `/root/.truerepublic` mounts vs the non-root image (`Dockerfile:58 USER truerepublic`); AWS user-data seds the nonexistent `EXTERNAL_IP`.

**Impact:** every operator following the public wiki hits a wall or opens an exposure — for a recovery project, these are the exact pages new validators will read first.

## TRR-25 — Medium — Fictional `truerepublic.network` infrastructure

**Evidence (lead-verified NXDOMAIN 2026-09-15):** the domain is referenced as seed/snapshot/state-sync RPC host (wiki Node-Setup), **security@ contact** (`wiki/security/Security-Architecture.md:650`), troubleshooting status page (`docs/user-manual/troubleshooting.md:133`), and bug-bounty repro endpoint (`.github/ISSUE_TEMPLATE/bug_bounty.md:15`, `https://api.truerepublic.network/transfer`). An unregistered domain used as a **security contact and seed host** is a theoretical squatting hazard: whoever registers it controls a channel users are told to trust.

**Recommendation:** register the domain or purge all references; point the security contact at the GitHub private-vulnerability flow that SECURITY.md already defines.

## TRR-26 — Low — Client trust/custody notes

**Evidence:** `client-web/src/config/chains.ts:11-33` — one configurable RPC/REST (build-time env, or same-origin `/rpc` `/api` reverse-proxying to 127.0.0.1); all module queries use `abci_query` with `prove:false` (`services/moduleQuery.ts:418`) — balances, pool prices, nullifier status all trust one node; the chain-id check at connect (`signingClient.ts:72-77`) prevents wrong-chain signing, but a malicious endpoint can display fabricated state or censor broadcasts; defaults are plaintext `http://localhost`. Wallet KDF is PBKDF2-SHA-256 600k (not memory-hard; agent-flagged), the session keeps the plaintext password in memory (`stores/walletStore.ts:139`), and `UnlockWallet.tsx` has no attempt throttling. `deliverMessages` surfaces raw chain `rawLog` into UI errors (public data).

## TRR-27 — Low — Landing hotlinks third-party images despite local copies

**Evidence:** favicon/logo/PNYX images load from `raw.githubusercontent.com` (`docs/index.html:9,293,318,461,746`) while local copies exist in `docs/assets/` — privacy (referrer leakage to GitHub), availability (raw.githubusercontent is not a CDN), no SRI. Serve from `docs/assets/`.

## TRR-28 — Info — Verified client/deploy strengths

- Mock proof submission **disabled**: `ZKPService.initialize`/`generateProof` throw unconditionally; `isSubmittable` hardcoded false; vote button disabled with warning banner; `MsgRateWithProof` commented out and **absent from the tx registry**; no `MsgWithdrawFromDomain` type at all.
- Mnemonics encrypted at rest (AES-GCM + PBKDF2) with address-derivation cross-check on signing; importWallet bounds error messages without wordlist leakage; session-generation proxy invalidates stale signers.
- Bundle governance: `chunkSizeWarningLimit: 1100` + CI-enforced budget (`scripts/bundle-budget.mjs`) + 20 lazy routes — the GH-128 closure is real (resolves the CODEX_AUDIT LOW note).
- Playwright/browser-quality suites cover axe WCAG, keyboard focus, viewport overflow on safe routes.
- Dockerfile enforces `npm ci`, digest-pinned base images, non-root nginx runtime; compose binds loopback, mandates `GRAFANA_PASSWORD`, disables gRPC publicly; no retired web-wallet/mobile-wallet references live anywhere (retirement-guard scripts enforce it).

---

## Verified strengths (beyond TRR-28)

- Zero `<script>` tags on the landing page — zero third-party JS anywhere on the public site.
- Telegram community link live and consistent (65 members, description links back to the site — not hijacked).
- Landing↔repo byte-identity verified; status.json live and internally consistent (2,117+26+319=2,462; module table sums; rollout caps respected; live issue #29 arithmetic matches per-phase counts).
- configs/ and monitoring/ are clean: no real IPs, hosts, or credentials anywhere (loopback scrape targets; synthetic examples; `threat-model.json:8` explicitly declares the repo secret-free).
