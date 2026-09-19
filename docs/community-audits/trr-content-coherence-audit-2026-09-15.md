# TrueRepublic — Content & Coherence Audit

- **Series:** Collateral Web3 Open Audits
- **Date:** 2026-09-15
- **Target:** NeaBouli/TrueRepublic @ `f5de5a150b9ff0edc51d68411b7829e78bcf78e6` + GitHub wiki clone + live landing
- **Scope:** README/badges/status.json, whitepapers (EN/DE/PDF), wiki triple (repo wiki/ ↔ GitHub wiki ↔ docs/), CODEX_AUDIT re-baseline, sovereign docs, release/install docs, license stack
- **Method:** deep-recon agent sweep + lead verification of every finding; headline numbers re-computed independently
- **Register:** TRR-29 … TRR-37 (this report) — **0 Critical / 0 High / 3 Medium / 5 Low / 1 Info**

---

## Executive summary

The headline discipline is genuinely strong — every deterministically checkable number reconciles: the 2,462-test decomposition (2,117+26+319), rollout percentages against live issue #29 (36/59→61%, 36/51→71%, 6/7→86%, 3/10→30%), client inventory (44 components/23 routes/9 stores/17 services), module versions, wiki↔repo-wiki sync (12 pages byte-identical, 20 mechanically different), live↔docs byte-identity, and the license stack (Apache-2.0 decision #219, NOTICE, REUSE.toml, machine-enforced policy). Much of it is CI-enforced (`check-consistency.sh`). The incoherence lives one level below, in surfaces the consistency gate does not pin: the whitepaper's present-tense overclaims (live contracts, absolute anonymity, THORChain TSS boilerplate, price promises), the PoD anti-whale rule that code doesn't implement, stale sub-counts, and un-quarantined planning docs.

## Severity table

| ID | Severity | Title |
|----|----------|-------|
| TRR-29 | Medium | Whitepaper overclaims: "smart contracts are live on TRChain", §4 anonymity as existing absolute, §6 GG20 TSS (zero in repo), §5 cross-chain AMM overstatement |
| TRR-30 | Medium | PoD whitepaper rule 1 not implemented (bought coins qualify) vs landing's anti-whale marketing |
| TRR-31 | Medium | Financial-performance promises in a public whitepaper ("secure store of value", "ensures a steady price increase") |
| TRR-32 | Low | Stale sub-counts: "614" truedemocracy cases on three surfaces; CODEX_AUDIT's own 690/656 numbers; QUICKSTART rot (Go 1.24+, "8 frontend tests", version output, `cargo wasm`) |
| TRR-33 | Low | Stale security-Known-Issues page understates closed work; missing recovery caveats on six high-risk wiki pages |
| TRR-34 | Low | Stale structure claims (contracts layout, DEX test counts, Makefile `install`, Node 18 vs 22, guide counts, "go mod tidy" advice) |
| TRR-35 | Low | Un-quarantined stale planning docs (V0.3.0_ROADMAP, OPTIONAL_INDEXER_STACK, SESSION_SUMMARY); German whitepaper fee model contradicts code; Trustee-Modell unimplemented and unlabeled |
| TRR-36 | Low | Hygiene cluster: "WhitePaper_TR_eng Kopie 2.pdf" live-served; committed Mach-O binary `ui/truerepublic_ui`; unreferenced 2.4 KB CI-CD PDF; bug-bounty template contradicts SECURITY.md; "Team (BTC multi-sig)" vs community framing |
| TRR-37 | Info | Headline discipline verified; GH-128 bundle contradiction resolved in favor of the supersede note; PNYX/pnyx naming handled correctly |

---

## TRR-29 — Medium — Whitepaper overclaims

**Evidence (lead-verified):**
- `docs/WhitePaper_TR_eng.md:348` — "**The smart contracts are live on TRChain. Anyone can interact with them directly.**" No public network exists; the README banner says the opposite.
- `:327` — "**All** user activities on the issue and suggestion list such as votes, ratings and suggestions **are anonymous**" — present-tense absolute with no boundary on the chapter, while the project's own register carries SRF-ZKP-001 (critical, open: no production prover), client submission is hard-disabled, and GH-209 direct payouts are publicly linkable. This page is linked from the landing footer as "Whitepaper".
- `:358` — "TRChain is a decentralised cross-chain liquidity protocol which uses … **GG20 Threshold Signature Scheme (TSS)**" — zero TSS anywhere in the repo (grep over x/, contracts/, client-web: empty); the section reads like inherited THORChain boilerplate. §5 "cross-chain AMM … on-chain vaults" overstates a single-chain AMM with registered IBC assets.

**Impact:** the whitepaper is the project's front door for technical due diligence; these claims contradict the recovery doctrine the project itself enforces everywhere else.

**Recommendation:** apply the same claim-audit treatment the code docs received (dated status banner + present-tense scrub), or version the whitepaper as a *vision* document with an explicit reality-gap section.

## TRR-30 — Medium — PoD rule 1 not implemented vs anti-whale marketing

**Evidence (lead-verified):** WP §7 (`:375`) — "Staking amounts must originate from a domain wallet directly… borrowed coins don't qualify." Code enforces only domain membership + the 10% domain-payout transfer limit (`x/truedemocracy/validator.go:347`); bought coins qualify (the wiki even documents this accurately: "Bought from exchange: No limit"). Meanwhile `docs/index.html:345` markets PoD as "ensures network control stays with active community members, not wealthy investors" — the anti-whale claim exceeds the mechanism.

## TRR-31 — Medium — Financial-performance promises

**Evidence (lead-verified):** `WhitePaper_TR_eng.md:220` "As a **secure store of value**"; `:225` "…**ensures a steady price increase, while reducing price volatility**." Investment-guarantee tone on a public whitepaper of a project with no live network — a consumer-protection and credibility liability (and the kind of sentence regulators quote).

**Recommendation:** remove or reframe as design intent with explicit no-guarantee language.

## TRR-32 — Low — Stale sub-counts and QUICKSTART rot

- `README.md:181` "614 standard-suite cases" vs `docs/status.json:85` 676 (= 614 + GH-297's 62, per README's own `:286-288`); same stale 614 in `wiki/develop/Module-Deep-Dive.md:29` and `Code-Structure.md:21` — the consistency script pins only headline totals.
- `CODEX_AUDIT.md:119` "one 690-test source of truth", `:146` "656 Go cases" predate the 2,462 baseline; `:134` cites retired `web-wallet/src/services/api.js` as evidence — covered by the historical banner, but sloppy for an audit artifact.
- `docs/QUICKSTART.md:8` "Install Go 1.24+" (vs 1.26.6 everywhere else); `:121` "Maintained frontend tests (8)" (vs 319); `:29-31` documents `truerepublicd version` output "v0.4.0" (Makefile injects `git describe --tags --always` → at f5de5a15 that's `v0.4.0-251-g…`); `:87` `cargo wasm` (undocumented plugin; INSTALLATION.md uses plain `cargo build --target`).

## TRR-33 — Low — Stale security-Known-Issues; missing recovery caveats

`security-Known-Issues.md:25-28` lists consensus-key rotation, network-policy drills and persisted-state migration recovery as open — GH-56/GH-71/GH-184 are closed (issue #29 Phase 1 checked, harness files exist); the page understates completed security work and contradicts the same wiki's Current-Status. Separately, no non-production boundary on `operations-Validator-Guide.md` (100k PNYX staking walkthrough), `operations-Deployment-Options.md`, `operations-Troubleshooting.md`, `users-How-It-Works.md`, `users-System-Overview.md`, `develop-Module-Deep-Dive.md` — while Home/Installation/status pages carry it.

## TRR-34 — Low — Stale structure claims

`develop-Code-Structure.md`: `contracts/src/governance.rs+treasury.rs` (actual: 7-crate workspace core/packages/examples); "dex keeper_test.go 24 DEX tests" (claimed current: 138); "Makefile … install" target (absent); "node-operators 9 guides" (27 non-README files), developers "8 guides" (10), user-manual "7 guides" (6+README). `users-Installation-Wizards.md`: Node 18 vs required 22+ (`client-web/package.json` engines `>=22`); "go mod tidy" advice vs the repo's frozen-module-graph discipline. Lifecycle zone durations: `users-How-It-Works.md:402-410` and `users-System-Overview.md:76-78` "GREEN 0-7 days / YELLOW 7-30 / RED 30+" vs code default 1 day per zone (`x/truedemocracy/types.go:27`, matching WP §3.1.2).

## TRR-35 — Low — Un-quarantined stale planning docs; German whitepaper; Trustee-Modell

`docs/V0.3.0_ROADMAP.md` ("Status: Planned", references retired `web-wallet/src/…` as current structure), `docs/V0.4.0_OPTIONAL_INDEXER_STACK.md` ("In Planning, Target Q2 2026" — elapsed; no indexer exists), `docs/SESSION_SUMMARY_2026-02-28.md` — none received the historical-banner treatment that RELEASE_NOTES_v0.3.0.md got. `WhitePaper_TR.md:75` (German) describes a "Maker/Taker + Treasury-Anteil" fee model vs the actual 0.3%-to-LP + 1% burn, and "PNYX/ATOM" pools vs EN "BTC/ETH/LUSD". The Trustee-Modell (§3.2) has no implementation anywhere (grep over x/, client-web, contracts: empty) and no deferred-work label.

## TRR-36 — Low — Hygiene cluster

- `docs/WhitePaper_TR_eng Kopie 2.pdf` — spaces + German "Kopie 2" in the public docs tree, **live-served** at the Pages URL; excluded from the Apache grant as "historical PDF".
- `ui/truerepublic_ui` — a checked-in compiled Mach-O x86_64 binary (with `ui.cpp`), excluded from the license grant as "unmaintained ui/** prototype", absent from README's repo-structure section. Committed binaries in a source repo are a supply-chain smell even when labeled unmaintained.
- `TrueRepublic_CI_CD_Security.pdf` (2.4 KB) at repo root — trivially small, unreferenced.
- `.github/ISSUE_TEMPLATE/bug_bounty.md` asks for "BTC or PNYX address (for Reward)" implying a paid bounty — `SECURITY.md:18` states no bug-bounty payment is promised; the template also lists the fictional API endpoint (TRR-25).
- `README.md:441` "Team (BTC multi-sig)" donation line sits oddly next to the "community-governed project without a central corporate owner" framing — worth a clarifying note.

## TRR-37 — Info — Verified headline discipline

Independently re-computed and reconciled: 2,462 = 2,117 Go + 26 Rust + 319 frontend (module table sums to 2,117); rollout caps respected (2117 ≤ 2462); live issue #29 has exactly 36 `[x]` + 25 `[ ]` = 61 boxes = 59 top-level + 2 subchecks, per-phase counts match; issue #4 closed 2026-08-16; `release_date: 2026-03-04` = actual v0.4.0 tag date; tech badges match manifests (Vite 8.2.2 lockfile, CosmJS 0.39, Go toolchain 1.26.6, Cosmos SDK v0.50.15, ibc-go v8.7.0, gnark Groth16); status.json "676 truedemocracy" = 614 + GH-297's 62 per README; GH-128 bundle-splitting closure is real (20 lazy routes + CI bundle budget) — resolving the apparent contradiction with CODEX_AUDIT's LOW note in favor of the supersede banner; all GH-supersession claims (GH-209/56/71/175/178/181/89/97/101/80/85/184) verified against repo evidence files; license stack consistent (Apache-2.0, NOTICE, REUSE.toml, machine-enforced `check-license-policy.sh`); PNYX-vs-NeaBouli/pnyx naming handled correctly (`SOVEREIGN_V4_EDGE_ARCHITECTURE.md:22-25` frames pnyx as external sibling; no surface conflates them).

---

## Verified strengths

- `check-consistency.sh` CI-greps headline numbers between status.json and index.html — the discipline that kept the headline layer clean; the findings above are all **outside** that gate's current pin set (root-cause recommendation: extend the pin set to sub-counts, wiki operations pages, and whitepaper status banners).
- The historical-banner convention (CODEX_AUDIT, RELEASE_NOTES) is applied honestly where applied at all — extend it to the planning docs in TRR-35.
- Wiki↔repo-wiki sync verified per-page (12 byte-identical, 20 mechanically different link syntax only) — no content fork between the two wikis.
