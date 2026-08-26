# GH-219 Repository License Decision Package

- **Issue:** [GH-219](https://github.com/NeaBouli/TrueRepublic/issues/219)
- **Machine-readable state:** `configs/legal/license-decision.json`
  (`truerepublic.license-decision/v1`), decision status: **decided**
- **Deterministic check:** `scripts/check-license-policy.sh`
  (fail-closed across pending and decided states)
- **Final engineering audit:** `docs/legal/GH-219-CODEX-AUDIT.md`
- **Governance record:**
  [GH-219 comment 5423337355](https://github.com/NeaBouli/TrueRepublic/issues/219#issuecomment-5423337355)

> **This document is not legal advice.** It is an engineering decision
> package prepared from repository evidence and primary upstream sources. The
> community has now selected Apache-2.0 for maintained source code and
> maintained documentation. Individual contributors retain copyright and the
> collective attribution is “TrueRepublic contributors.” Brand assets,
> artwork, historical PDFs, archived historical evidence, and third-party
> materials remain excluded unless a file-specific notice applies or their
> provenance and permission are documented.

## 0. Decision recorded and publication scope

TrueRepublic is a community-governed open-source project without a central
corporate owner. The exact GH-219 governance record approves:

- SPDX identifier `Apache-2.0`;
- collective attribution `TrueRepublic contributors`, without assignment of
  individual contributor rights;
- maintained source code and maintained documentation as the covered project
  scope; and
- brand assets, artwork, historical PDFs, archived historical evidence, and
  third-party materials as explicit exclusions unless a file-specific record
  later documents provenance and permission.

The root `LICENSE`, `NOTICE`, npm/Cargo metadata, public documentation, and
`configs/legal/license-decision.json` implement that decision. This governance
foundation earns no rollout credit, makes no production claim, and does not
replace per-component dependency review.

---

## 1. Scope and method

This package records the completed GH-219 decision and its publication. The
pre-decision analysis is retained so the community can audit why the chosen
scope and exclusions exist. It:

1. inventories what the repository actually contains and who authored it;
2. classifies every maintained component as repository-authored code,
   generated material, third-party code/assets, external dependency, or
   historical-only content;
3. audits current public license claims and reconciles contradictory
   maintained wording;
4. compares Apache-2.0, AGPL-3.0-only, and a dual-license approach against
   the dependency, distribution, contribution, network-copyleft, and patent
   dimensions of this specific repository;
5. records the recommended decision path and, in Section 8, the exact
   implementation steps executed for publication plus the retained post-merge
   instructions.

Evidence base: the tracked file census (`git ls-files`, 622 files at the
GH-219 base), authorship metadata (`git shortlog`), package manifests
(`go.mod`, `contracts/Cargo.toml`, `client-web/package.json`), the
documented Alpha/V4 dependency boundaries
(`docs/SOVEREIGN_ALPHA_ARCHITECTURE.md`, `docs/SOVEREIGN_V4_EDGE_ARCHITECTURE.md`),
and primary upstream license texts cited in Section 9. Dependencies are not
vendored; their licenses are properties of the pinned upstream releases, not
of this repository.

---

## 2. Inventory and provenance classification

### 2.1 Repository-authored code

The inventory below records the analyzed pre-decision base. The publication
implemented by this change adds the Apache-2.0 metadata summarized in the
final column; it does not rewrite the historical evidence used to reach the
decision.

| Component | Paths | Files (approx.) | Notes |
|---|---|---|---|
| Go daemon, modules, tools | root `*.go`, `x/dex`, `x/truedemocracy`, `treasury`, `token`, `migration`, `observability`, `healthcheck`, `networkpolicy`, `topologypolicy`, `incidentpolicy`, `capacitypolicy`, `genesisevidence`, `releaseevidence`, `deploymentevidence`, `installlifecycle`, `internal/zkpprover`, `sovereignv4/protocol`, `cmd/*` | 210 `.go` | module `truerepublic` (`go.mod`); consensus-critical |
| Rust/CosmWasm workspace | `contracts/` (`core`, `packages/bindings`, `packages/testing-utils`, `examples/*`) | 32 `.rs` | pre-decision manifests had no license field; every maintained crate now declares `Apache-2.0` |
| Maintained TypeScript client | `client-web/src`, `client-web/scripts` | 114 `.ts`/`.tsx` | remains private and now declares `Apache-2.0` in `client-web/package.json` |
| Shell/ops scripts, configs, CI | `scripts/`, `configs/`, `.github/`, `monitoring/`, `nginx/` | ~60 | repository-authored |
| Documentation, wiki, whitepapers (Markdown) | `docs/`, `wiki/`, root `*.md` | 150 `.md` | repository-authored text |
| C++ prototype UI | `ui/ui.cpp` plus tracked Mach-O artifact `ui/truerepublic_ui` | 2 | unmaintained prototype; the binary's origin is not documented well enough to classify it conclusively |

Authorship metadata (`git shortlog -sne`) shows the tracked history was
written essentially by the project principals (`True Republic <…NeaBouli…>`,
`Gio Mario`, `NeaBouli`, plus two local `Pingonaut` test identities and
`dependabot[bot]`).

**Formal copyright posture at the analyzed pre-decision base:**

- No maintained Go, Rust, TypeScript, or C++ source file carries a copyright
  header or project SPDX declaration; no
  `LICENSE`, `COPYING`, or `NOTICE` file exists at any level.
- No Contributor License Agreement (CLA) or Developer Certificate of Origin
  (DCO) is recorded in the repository.
- Commits are made under personal identities; no employer or work-for-hire
  assignment is recorded. Copyright therefore rests with the individual
  authors by default, concentrated in the project principals, but this is an
  inference from authorship metadata, not a documented legal fact.
- The GH-219 governance record now supplies the documented project decision:
  individual contributors retain their rights, the collective attribution is
  “TrueRepublic contributors,” and maintained source/documentation are
  Apache-2.0. Any later license *change* would still require the rights and
  consent appropriate to the affected contributions.

### 2.2 Generated material

- No generated source is committed: there are no `*.pb.go`, no
  protoc/prost output, and no "Code generated … DO NOT EDIT" files. The
  protobuf registry/codec work is hand-written and repository-authored
  (`Makefile` `proto-gen` is an explicit stub).
- Lockfiles and checksums are generated *metadata*:
  `client-web/package-lock.json`, `contracts/Cargo.lock`, `go.sum`. They
  do not replace the notice obligations of the dependencies they enumerate.
- The tracked `ui/truerepublic_ui` Mach-O executable appears related to
  `ui/ui.cpp`, but the repository does not contain enough provenance evidence
  to confirm how it was produced. Its presence in version control is a hygiene
  observation for a separate task; its origin must be confirmed before it is
  distributed.

### 2.3 Third-party code and assets inside the repository

- **No third-party source code is vendored.** `vendor/` is gitignored;
  `node_modules/`, `build/`, and `contracts/target/` are untracked build
  caches.
- **Assets of unknown provenance (flagged honestly):** `assets/*.png`,
  `assets/*.svg`, and their duplicates under `docs/assets/` (10 PNG, 4 SVG)
  are project logos/ticker artwork introduced by commit `d490bea` ("assets:
  integrate official logos throughout project"). No author, tool, or license
  note is recorded for them. They are presumed project-commissioned
  branding, but provenance is undocumented; before a license is applied to
  them, the community must document that they are original project artwork
  (or replace them with documented-original versions). Trademark/branding rights
  in the "TrueRepublic" and "PNYX" marks are also unregistered and
  undocumented.
- **PDFs:** `docs/WhitePaper_TR_eng Kopie 2.pdf` and
  `TrueRepublic_CI_CD_Security.pdf` are project documents distributed in
  binary form; their Markdown sources exist for the whitepaper
  (`docs/WhitePaper_TR_eng.md`). Provenance of the rendered PDFs is presumed
  project-authored but, as binaries, is not independently verifiable.

### 2.4 External dependencies (not distributed by this repository)

Dependencies are fetched by package managers from pinned manifests; the
repository does not copy their code. Their licenses nonetheless constrain
what the project license can be and what obligations arise when binaries are
distributed. Direct dependencies, classified by license family (primary
sources in Section 9):

**Go (`go.mod`, direct requirements):**

| Dependency | Version | License |
|---|---|---|
| cosmos-sdk (+ `cosmossdk.io/*`) | v0.50.15 | Apache-2.0 |
| CometBFT | v0.38.26 | Apache-2.0 |
| wasmd / wasmvm | v0.53.4 / v2.2.2 | Apache-2.0 |
| ibc-go v8 (+ capability) | v8.7.0 / v1.0.1 | **MIT** (verified against the pinned LICENSE) |
| gnark / gnark-crypto | v0.14.0 / v0.19.2 | Apache-2.0 (verified) |
| cosmos-db | v1.1.3 | Apache-2.0 |
| gogoproto | v1.7.2 | BSD-3-Clause |
| grpc / protobuf | v1.82.1 / v1.36.12 | Apache-2.0 / BSD-3-Clause |
| grpc-gateway v1 | v1.16.0 | BSD-3-Clause |
| prometheus client_golang | v1.21.1 | Apache-2.0 |
| cobra / viper / zerolog / testify | — | Apache-2.0 / MIT / MIT / MIT |
| `golang.org/x/mod` | v0.37.0 | BSD-3-Clause |
| `gopkg.in/yaml.v3` | v3.0.1 | MIT and Apache-2.0 |

**Rust (`contracts/`, workspace dependencies):** cosmwasm-std / cosmwasm-vm
3 (Apache-2.0), serde 1 (MIT OR Apache-2.0), thiserror 2 (MIT OR
Apache-2.0), schemars 0.8 (MIT).

**npm (`client-web/package.json`, runtime dependencies):** CosmJS
(`@cosmjs/*`, `cosmjs-types`) Apache-2.0; react, react-dom, react-router-dom,
zustand, tailwindcss, @headlessui/react, @heroicons/react, clsx, postcss and
autoprefixer — all MIT. The last two are build tooling currently placed in the
manifest's `dependencies` section rather than `devDependencies`.

**npm dev tooling:** typescript Apache-2.0; eslint, vite, vitest, happy-dom,
@testing-library/*, @types/* — MIT; playwright
Apache-2.0; **@axe-core/playwright / axe-core MPL-2.0** (verified) — a
weak, file-level copyleft dev-only accessibility tester; it is not linked
into the shipped client bundle and is compatible with every candidate
license.

**Finding:** every maintained dependency is permissive or weak-copyleft
(Apache-2.0, MIT, BSD-3-Clause, BSL-1.0 at the documented boundary, MPL-2.0
dev-only). **No GPL-family dependency exists in the maintained tree.**
Nothing therefore forces or forbids a copyleft project license: the choice
is genuinely open. All dependency licenses permit combination under either
Apache-2.0 or AGPL-3.0-only for the project's own code; Apache-2.0/MIT/BSD
dependencies only require notice preservation on redistribution, and MPL-2.0
keeps its file-level copyleft inside its own files.

### 2.5 Historical-only content

| Content | Status |
|---|---|
| `RELEASE_NOTES_v0.3.0.md` (root) | historical draft, quarantined with the marker "Historical draft only"; its "See LICENSE" line is historical wording covered by the explicit allowlist in `configs/legal/license-decision.json` |
| `docs/archive/releases/*` | quarantined historical release notes; the v0.3.0 draft is already annotated that no root license exists and the decision is tracked in GH-219 |
| `client-web/CHANGELOG.md` | historical pre-recovery snapshot (marker "Historical pre-recovery snapshot") |
| Retired `web-wallet` / `mobile-wallet` | removed under GH-112/GH-102; exist only in Git history, audit-only, never a merge source |
| Legacy preserved checkout | exists only outside this repository; never a merge source (per `AGENTS.md`) |

Historical records are preserved verbatim; the deterministic check permits
their stale wording only through the explicit per-path allowlist.

---

## 3. Pre-decision public-claim audit and resolution

The numbered findings below describe the GH-219 base before publication. Each
is retained as decision evidence and followed by its current resolution:

1. **Pre-decision finding:** no `LICENSE`/`COPYING`/`NOTICE` existed and no maintained source or
   package metadata carries a project SPDX declaration. This decision package,
   policy manifest and checker intentionally contain literal SPDX terms for
   analysis and enforcement; they do not declare the project license. Under
   default copyright law, publishing
   source without a license grants no reuse rights: others may view and
   fork per the GitHub Terms of Service, but may not legally copy, modify,
   or redistribute
   ([GitHub licensing documentation](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/licensing-a-repository),
   [choosealicense.com/no-permission](https://choosealicense.com/no-permission/)).
   **Resolution:** the canonical root `LICENSE` and `NOTICE`, package
   metadata, `REUSE.toml`, and the exact GH-219 decision record now establish
   the maintained project scope and contributor-retained attribution.
2. **Pre-decision contradiction:** `CONTRIBUTING.md` asserted contributions
   "will be licensed under the same license as the project (Apache 2.0)"
   although no such license was ever published. That wording was first
   corrected to the pending state in the pre-decision package. **Resolution:**
   maintained contributor guidance now states Apache-2.0
   inbound-equals-outbound terms and links the
   published root files. Stale wording survives only in explicitly allowlisted,
   clearly marked historical records.
3. **Pre-decision documentation state:** maintained public and architecture
   docs described a pending GH-219 gate. **Resolution:** README, landing page,
   roadmap, limitations, whitepapers, architecture docs, maintained-client
   docs and wiki sources now all describe the recorded decision and exclusions.
4. **Policy resolution:** `configs/legal/license-decision.json` is the
   machine-readable decided state. `scripts/check-license-policy.sh` now fails
   closed unless the canonical license text, NOTICE, attribution, scope,
   exclusions, REUSE boundary, and npm/Cargo metadata remain consistent.

---

## 4. Candidate comparison

### 4.1 Apache-2.0

- **Text/identity:** [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0)
  ([SPDX: Apache-2.0](https://spdx.org/licenses/Apache-2.0.html)).
- **Copyleft:** none. Derivatives may be closed and relicensed; notices and
  the license text must be preserved, modified files must carry change
  notices, and a `NOTICE` file must be propagated when present (§4).
- **Patent:** express patent grant from every contributor for claims
  necessarily infringed by their contribution, with retaliation termination
  if the licensee sues over the work (§3). That protection is material for a
  novel protocol stack, and several pinned upstreams use the same license
  (cosmos-sdk, CometBFT, wasmd, gnark).
- **Compatibility:** compatible into GPLv3/AGPLv3 (Apache-2.0 code can be
  combined and the combination distributed under (A)GPLv3), but
  **not compatible with GPL-2.0-only** per the
  [FSF license list](https://www.gnu.org/licenses/license-list.html#apache2).
  No GPL-2.0-only code exists or is planned in the maintained tree
  (Section 2.4), so this is informational, not blocking.
- **Contribution friction:** minimal; §5 defines inbound=outbound
  ("submission of contributions under the same terms"), which matches this
  repository's no-CLA posture.
- **Ecosystem fit:** identical to the chain's entire upstream stack; the
  least surprising choice for Cosmos/IBC operators, relayers, and exchanges.

### 4.2 AGPL-3.0-only

- **Text/identity:** [GNU Affero GPL v3](https://www.gnu.org/licenses/agpl-3.0.html)
  ([SPDX: AGPL-3.0-only](https://spdx.org/licenses/AGPL-3.0-only.html)).
- **Copyleft:** strong, including **network copyleft (§13)**: anyone running
  a modified version that users interact with over a network must offer the
  corresponding source. For a blockchain this binds operators offering
  modified daemons/clients as a network service; running an *unmodified*
  node creates no new obligation, and on-chain interaction by third parties
  does not "convey" their software.
- **Patent:** GPLv3 §11 contains patent-license and retaliation provisions,
  but their scope and operation differ from Apache-2.0 §3 and should not be
  treated as legally interchangeable without counsel.
- **Compatibility:** one-way compatible with GPLv3; AGPL-3.0-only project
  code can safely embed the permissive (Apache/MIT/BSD) dependency set in
  Section 2.4. Element's 2023 move of Synapse to AGPLv3/commercial
  ([Element, Dec 2023](https://element.io/blog/element-to-adopt-agplv3/)) is
  the documented precedent for exactly this model in an adjacent ecosystem.
- **Contribution friction:** higher. Some corporate validators, custodians,
  and integrators have policies against AGPL; the documented TrueRepublic
  governance model may *want* that pressure (public infrastructure should
  stay public), but it is a real adoption cost.
- **Ecosystem fit:** less aligned with the Apache-2.0/MIT licenses of the
  pinned Cosmos/IBC stack, but more aligned with the project's stated goal
  that modified public infrastructure should remain auditable.

### 4.3 Dual-license approach

Two distinct models are commonly conflated; they have opposite effects:

1. **Recipient's-choice dual license ("Apache-2.0 OR AGPL-3.0-only").**
   Each recipient picks either license. Because anyone can choose the
   permissive side, this **collapses in practice to Apache-2.0**: it does
   not preserve copyleft. Its only honest benefit is letting downstreams
   that require GPL-family terms (e.g. a GPLv3-only project) combine the
   code without doubt. Used by e.g. serde-style "MIT OR Apache-2.0" crates.
2. **Copyleft-preserving dual licensing ("AGPL-3.0-only, or a paid
   commercial license").** Requires a **single copyright holder or a CLA**
   granting the owner re-licensing rights over every contribution. This
   repository currently has no CLA/DCO (Section 2.1), so adopting this model
   later would demand retroactive contributor consent. It monetizes
   proprietary use but adds governance centralization and contribution
   friction, and no such commercial entity is documented for TrueRepublic.

### 4.4 Side-by-side summary

| Dimension | Apache-2.0 | AGPL-3.0-only | Dual (Apache OR AGPL) | Dual (AGPL + commercial) |
|---|---|---|---|---|
| Copyleft | none | strong + network (§13) | none in practice | strong + network |
| Patent grant | express (§3) | express (§11) | express either way | express (AGPL side) |
| GPLv2-only combinability | no | no | AGPL side: no | no |
| GPLv3 ecosystem combinability | one-way in | yes | yes | yes |
| Works with current deps (2.4) | yes | yes | yes | yes |
| Works with documented Alpha boundary (Telegram code remains rejected; Matrix/Synapse is only a gated fallback) | yes, subject to per-component review | yes, subject to per-component review | as Apache | subject to CLA/rights consolidation |
| Contribution friction (no CLA today) | lowest | low–medium | low | requires CLA/consent |
| Cosmos ecosystem norm | yes (sdk/comet/wasmd) | no | partial | no |
| Community maintenance effort post-decision | LICENSE + NOTICE + scope metadata | LICENSE + scope metadata | two LICENSE texts + notice | LICENSE + rights-consolidation process |

---

## 5. Distribution, contribution, network-use, and patent implications

- **Distribution of binaries** (daemon releases, Docker images, client
  bundles): under every candidate, the permissive dependency set requires
  preserving license texts/notices in distributions. The published root
  `NOTICE` covers the project's own attribution; bundling third-party license
  texts into distributed artifacts remains a retained pre-distribution step
  (Section 8, item 10).
  Apache-2.0 additionally requires change notices in modified files when
  redistributing modified upstream code — this repository does not modify
  vendored upstream code, so the obligation is trivial.
- **Contribution:** with no CLA/DCO today, inbound rights rest on whatever
  the chosen license says (Apache-2.0 §5 and the GPL family's
  inbound=outbound convention both suffice). Apache-2.0 §5 inbound=outbound
  terms were recorded in `CONTRIBUTING.md`. A DCO sign-off line would be a
  lightweight hardening compatible with all candidates; it was not adopted in
  this change and remains an optional community decision.
- **Network copyleft:** only AGPL-3.0-only (and the AGPL side of model 4.3.2)
  reaches hosted services. For TrueRepublic this matters for hosted RPC
  endpoints, explorers, and any future hosted client; it does not change
  obligations of ordinary validators running unmodified binaries.
- **Patents:** Apache-2.0 and AGPL-3.0 both give express grants with
  retaliation; a recipient's-choice dual license inherits whichever side the
  recipient uses. Given the ZKP/governance novelty of the codebase, an
  express patent grant is a substantive reason to prefer either candidate
  over MIT/BSD-style minimalism.

---

## 6. Honest unknowns and flagged risks

1. **Contributor-level provenance still matters** (2.1): the community record
   confirms contributor-retained copyright and Apache-2.0 for maintained work,
   but there is no central rights holder or CLA. A future license change must
   therefore respect the rights in every affected contribution.
2. **Asset provenance is undocumented** (2.3): logos/artwork and the two
   PDFs. Confirm they are original before the license covers them, or
   replace them; license selection for *code* is not blocked by this.
3. **Branding/trademark is unaddressed** by every candidate license
   (Apache-2.0 §6 and the GPL family both exclude trademarks). A separate
   naming/usage policy remains an open community-governance decision outside
   GH-219's scope.
4. **Future relicensing remains distributed:** Apache-2.0 is now published;
   because copyright remains with contributors, a future license change must
   obtain the consent or other valid authority required for affected work.
5. This package is engineering analysis, **not legal advice**; a
   jurisdiction-aware review by counsel remains available to the community.

---

## 7. Recorded decision and implemented path

The community has completed the decision gate. The evidence and recorded
governance decision establish this outcome:

1. **Apache-2.0 is selected** for maintained source and maintained
   documentation because it fits the Cosmos/IBC ecosystem and preserves the
   contributor-retained community model.
2. **The exact scope is recorded on GH-219** with the SPDX identifier,
   collective attribution, retained individual rights, and provenance
   exclusions.
3. **The repository implements the decision** through the root texts,
   package metadata, public documentation, machine policy, and adversarial
   fixtures.
4. **Asset and dependency boundaries remain open by design:** exclusions in
   Section 6 require provenance or per-component review before reuse or
   distribution.

---

## 8. Publication implementation and verification

The decision was implemented in one protected, reviewable change. Steps 1–8
below record what that change contains; steps 9–11 are retained instructions
that were not completed by this change and are phrased as such.

Completed in this change:

1. **LICENSE:** the root `LICENSE` carries the exact canonical text of the
   selected SPDX identifier from its primary source
   (Apache-2.0: `https://www.apache.org/licenses/LICENSE-2.0.txt`).
   The documented AGPL-3.0-only and recipient's-choice dual-license paths
   (two texts `LICENSE-APACHE`/`LICENSE-AGPL` plus a root `LICENSE`
   explaining the choice) were retained for the record and not taken.
2. **NOTICE (Apache-2.0):** the root `NOTICE` carries the project
   attribution line; Apache §4(d) then propagates it.
3. **Copyright line:** the recorded collective attribution
   `Copyright 2026 TrueRepublic contributors` is used consistently in
   `NOTICE` and scope metadata without assigning individual contributor
   rights to a central entity.
4. **SPDX scope declaration:** instead of per-file
   `SPDX-License-Identifier` headers, the change adopted the
   [REUSE](https://reuse.software/) `REUSE.toml` annotation convention for
   the maintained scope, so the deterministic check verifies the declared
   coverage mechanically. Artwork and PDFs carry no annotation and stay
   explicitly out of scope per the recorded decision until a file-specific
   record documents provenance and permission.
5. **Package metadata:** `"license": "Apache-2.0"` (with `"private": true`
   retained) in `client-web/package.json`; `license = "Apache-2.0"` in every
   `contracts/**/Cargo.toml` package manifest; Go modules have no license
   field — the root `LICENSE` governs them.
6. **README / CONTRIBUTING:** the license sections record the selected
   identifier, inbound=outbound contribution terms, and point to `LICENSE`
   and this decision record. A DCO sign-off requirement was left as an
   optional community hardening and was not adopted.
7. **Maintained docs:** `docs/LIMITATIONS.md`,
   `docs/SOVEREIGN_ALPHA_ARCHITECTURE.md` (DG-3),
   `docs/SOVEREIGN_V4_EDGE_ARCHITECTURE.md` (DG-V4-1), and
   `client-web/README.md` were moved from "pending" to the recorded
   decision; the allowlisted historical files were left untouched.
8. **Manifest and check:** `configs/legal/license-decision.json` records
   `status: decided`, `selected_spdx_id`, the copyright line, the exact
   GH-219 governance-comment URL in `decision_record`, and the decided
   scope; `scripts/check-license-policy.sh` enforces the decided branch
   (root license file present with the recorded SPDX identity, package
   metadata consistent, REUSE annotations representatively covering the
   declared maintained scope, pending-state contradictions still forbidden).
   **Retained conditional instruction:** if the decided metadata surface
   grows, `.github/workflows/docs-check.yml` coverage must be extended in
   the same reviewed change. The decided branch accepts the single-license
   candidates only; a dual-license decision requires extending this contract
   and its positive fixtures in the same reviewed change.

Retained instructions (not completed by this change):

9. **GitHub detection (post-merge):** after merge, confirm the repository
   page shows the license via GitHub's licensee-based detection
   ([licensing a repository](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/licensing-a-repository));
   no API write is needed beyond the committed `LICENSE`.
10. **Dependency notices for distributions (pre-distribution):** when
    binaries/images are next assembled, include the third-party license
    texts for the Section 2.4 dependency set in the distributed artifact
    (release tooling change, not part of the license decision itself).
11. **Downstream gates (standing condition):** the publication half of DG-3
    (Alpha) and DG-V4-1 (V4) is resolved by the recorded decision; adopting
    third-party code still requires the per-component compatibility review
    those gates specify before any adoption.

---

## 9. Primary sources

License texts and policies:

- Apache-2.0 text: <https://www.apache.org/licenses/LICENSE-2.0>
  (canonical text: <https://www.apache.org/licenses/LICENSE-2.0.txt>);
  SPDX record: <https://spdx.org/licenses/Apache-2.0.html>
- AGPL-3.0 text: <https://www.gnu.org/licenses/agpl-3.0.html>
  (canonical text: <https://www.gnu.org/licenses/agpl-3.0.txt>);
  SPDX record: <https://spdx.org/licenses/AGPL-3.0-only.html>
- FSF license list (Apache-2.0/GPL-2.0 incompatibility note):
  <https://www.gnu.org/licenses/license-list.html#apache2>
- GNU AGPL §13 rationale (Why the Affero GPL):
  <https://www.gnu.org/licenses/why-affero-gpl.html>
- GitHub licensing documentation (no-license default):
  <https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/licensing-a-repository>;
  <https://choosealicense.com/no-permission/>
- REUSE specification (SPDX header mechanics): <https://reuse.software/>

Dependency licenses (pinned upstream LICENSE files):

- cosmos-sdk (Apache-2.0): <https://github.com/cosmos/cosmos-sdk/blob/v0.50.15/LICENSE>
- CometBFT (Apache-2.0): <https://github.com/cometbft/cometbft/blob/v0.38.26/LICENSE>
- wasmd (Apache-2.0): <https://github.com/CosmWasm/wasmd/blob/v0.53.4/LICENSE>
- wasmvm (Apache-2.0): <https://github.com/CosmWasm/wasmvm/blob/v2.2.2/LICENSE>
- ibc-go v8.7.0 (MIT): <https://github.com/cosmos/ibc-go/blob/v8.7.0/LICENSE>
- gnark v0.14.0 (Apache-2.0): <https://github.com/consensys/gnark/blob/v0.14.0/LICENSE>
- gnark-crypto (Apache-2.0): <https://github.com/consensys/gnark-crypto/blob/v0.19.2/LICENSE>
- cosmos-db (Apache-2.0): <https://github.com/cosmos/cosmos-db/blob/v1.1.3/LICENSE>
- gogoproto (BSD-3-Clause): <https://github.com/cosmos/gogoproto/blob/v1.7.2/LICENSE>
- grpc-go (Apache-2.0): <https://github.com/grpc/grpc-go/blob/v1.82.1/LICENSE>
- protobuf-go (BSD-3-Clause): <https://github.com/protocolbuffers/protobuf-go/blob/v1.36.12/LICENSE>
- grpc-gateway v1 (BSD-3-Clause): <https://github.com/grpc-ecosystem/grpc-gateway/blob/v1.16.0/LICENSE>
- cobra (Apache-2.0): <https://github.com/spf13/cobra/blob/v1.9.1/LICENSE>
- viper (MIT): <https://github.com/spf13/viper/blob/v1.19.0/LICENSE>
- zerolog (MIT): <https://github.com/rs/zerolog/blob/v1.34.0/LICENSE>
- testify (MIT): <https://github.com/stretchr/testify/blob/v1.11.1/LICENSE>
- cosmwasm v3.0.4 (Apache-2.0): <https://github.com/CosmWasm/cosmwasm/blob/v3.0.4/LICENSE>
- serde (MIT OR Apache-2.0): <https://github.com/serde-rs/serde/blob/master/LICENSE-MIT>
- thiserror (MIT OR Apache-2.0): <https://github.com/dtolnay/thiserror/blob/master/LICENSE-MIT>
- schemars (MIT): <https://github.com/GREsau/schemars/blob/master/LICENSE>
- CosmJS (Apache-2.0): <https://github.com/cosmos/cosmjs/blob/v0.39.0/LICENSE>
- React (MIT): <https://github.com/facebook/react/blob/v18.2.0/LICENSE>
- React Router (MIT): <https://github.com/remix-run/react-router/blob/main/LICENSE>
- zustand (MIT): <https://github.com/pmndrs/zustand/blob/main/LICENSE>
- Tailwind CSS / Headless UI / Heroicons (MIT):
  <https://github.com/tailwindlabs/tailwindcss/blob/v3.4.1/LICENSE>
- Vite (MIT): <https://github.com/vitejs/vite/blob/main/LICENSE>
- Vitest (MIT): <https://github.com/vitest-dev/vitest/blob/main/LICENSE>
- TypeScript (Apache-2.0): <https://github.com/microsoft/TypeScript/blob/main/LICENSE.txt>
- Playwright (Apache-2.0): <https://github.com/microsoft/playwright/blob/main/LICENSE>
- axe-core (MPL-2.0): <https://github.com/dequelabs/axe-core/blob/develop/LICENSE>
- TDLib (Boost Software License 1.0):
  <https://github.com/tdlib/td/blob/master/LICENSE_1_0.txt>
- Element Synapse relicensing to AGPLv3/commercial:
  <https://element.io/blog/element-to-adopt-agplv3/>

Repository evidence cited inline: `go.mod`, `contracts/Cargo.toml`,
`client-web/package.json`, `configs/legal/license-decision.json`,
`docs/SOVEREIGN_ALPHA_ARCHITECTURE.md` (DG-3, upstream license table),
`docs/SOVEREIGN_V4_EDGE_ARCHITECTURE.md` (DG-V4-1/DG-V4-2),
`docs/LIMITATIONS.md`, `CONTRIBUTING.md`, `configs/release/compatibility.json`
(historical quarantine).
