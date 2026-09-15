# TrueRepublic — AI-Readiness & Landing Quality Audit

- **Series:** Collateral Web3 Open Audits
- **Date:** 2026-09-15
- **Target:** live `neabouli.github.io/TrueRepublic/` + `docs/` source @ `f5de5a150b9ff0edc51d68411b7829e78bcf78e6`
- **Scope:** landing metadata, sitemap/robots/llms, JSON-LD, accessibility, mobile UX, design-system quality, machine discoverability
- **Method:** deep-recon agent + lead verification (live probes of sitemap/robots/llms URLs — all 404; CSS/DOM inspection; landing byte-identical to source)
- **Register:** TRR-38 … TRR-42 (this report) — **0 Critical / 0 High / 0 Medium / 4 Low / 1 Info**

---

## Executive summary

The landing is content-honest (recovery banner disclaiming production/real funds in the hero, careful ZKP wording matching SRF-ZKP-001, all headline numbers CI-pinned to status.json) and privacy-clean (zero scripts at all), with passing contrast spot-checks. Its machine discoverability is minimal: **no Open Graph/Twitter/canonical/JSON-LD anywhere, and no sitemap.xml, robots.txt or llms.txt** (all live-404) — so social unfurls are bare and AI/search crawlers get no guidance. Mobile users lose the entire navigation at ≤768px with no hamburger alternative.

## Severity table

| ID | Severity | Title |
|----|----------|-------|
| TRR-38 | Low | No social/structured metadata: no OG, Twitter card, canonical, or JSON-LD on the landing |
| TRR-39 | Low | No sitemap.xml, robots.txt, or llms.txt (all live-404) |
| TRR-40 | Low | Mobile navigation disappears at ≤768px with no alternative; minor a11y nits |
| TRR-41 | Low | Inline styles across roadmap/hero; single ad-hoc `<style>` block; no design system |
| TRR-42 | Info | Positives: zero third-party scripts, AA contrast spot-checks pass, sane heading structure, CI-grep status discipline |

---

## TRR-38 — Low — No social/structured metadata

**Evidence:** `docs/index.html:4-10` — the head carries only description/keywords/title/icon: no Open Graph, no Twitter card, no canonical, no JSON-LD anywhere on the page. Link unfurls from Telegram/socials render bare. **Recommendation:** add OG/Twitter/canonical plus a `SoftwareApplication` JSON-LD block (the project's status.json already holds every field such a block needs).

## TRR-39 — Low — No sitemap, robots, or llms.txt

**Evidence (lead live-probed 2026-09-15):** `https://neabouli.github.io/TrueRepublic/sitemap.xml`, `/robots.txt`, `/llms.txt` → all **404**. The project's audience includes AI-assisted researchers (the repo itself is agent-maintained); a small llms.txt mirroring the recovery status + canonical links would prevent stale pre-training summaries from dominating. Sitemap.xml is a 6-line file for a single-page site.

## TRR-40 — Low — Mobile navigation disappears; a11y nits

**Evidence:** `@media (max-width:768px) { .nav-links { display:none } }` (`docs/index.html:284`) with no hamburger/alternative — mobile users lose all navigation. Minor: footer `h4` skips `h3` in the outline; no `<main>` landmark; emoji feature icons lack `role="img"`/`aria-hidden`. Positives verified: single h1, working anchor ids, keyboard-focus styles present.

## TRR-41 — Low — Inline styles; no design system

**Evidence:** the entire roadmap section, hero banner and image sizing use `style=` attributes (`docs/index.html:310,318,635-674`); one ad-hoc `<style>` block carries the whole page; no shared CSS file (only one page exists — but the same tokens repeat in client-web's separate design). Hygiene, not dysfunction — unlike the Prometheus landing, no duplicate-rule contrast destruction was found here (contrast spot-checks pass AA: `#666`/white 5.7:1 etc.).

## TRR-42 — Info — Verified positives

- **Zero `<script>` tags** on the landing — zero third-party JS, the cleanest possible supply-chain posture for a project front page.
- Recovery banner in the hero (`index.html:310-317`) explicitly disclaims production/real funds; the ZKP card (`:349-351`) is carefully worded to match the open SRF-ZKP-001.
- `check-consistency.sh:417-445` CI-pins totals and rollout percentages between status.json and index.html — stale headline numbers fail the build.
- Heading structure sane (single h1, h2 per section, working anchors); contrast spot-checks pass WCAG AA.

---

## Verified strengths (series context)

TrueRepublic's public surface is the smallest of the four projects in this audit series — one static page + status.json — and correspondingly has the smallest AI-readiness gap: the deficiencies are omissions (TRR-38/39), not errors. The content that *is* there is honest and CI-pinned, which is more than the IFR, Ekklesia, Stealth or Prometheus surfaces could claim across the board (fabricated dashboard, GA contradictions, stale wikis elsewhere). Adding the three missing files (OG/JSON-LD block, sitemap, llms.txt) is a one-hour fix that closes this entire report.
