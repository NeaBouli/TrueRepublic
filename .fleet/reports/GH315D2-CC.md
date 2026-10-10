# GH315D2 — report (Claude, 2026-10-10T19:05Z)

status: ok (candidate) — integration visual review owed to Codex; not live
branch: agent/claude/GH315D2-landing-semantics, base 5f99c2575edd9c78f6dd9a2c6962eec828ef405b (Draft #371), head 669e21c8db05d7be6b40cf41e4eea5a6541a98e3
PR: #372 (Draft, stacked on #371)
files: docs/index.html (+33/-23)
map: D2 header/nav -> existing anchors; content/footer -> landmarks/semantics (extends MAP §GH315D1; no new module/hop)

## Change
- nav wraps instead of display:none; wrap breakpoint <=1024px (first try 768px collided logo+nav at 769px — caught in screenshots, fixed); header static <=1024, sticky above
- nav links: 44px min-height, nowrap, :focus-visible 3px outline; nav aria-label="Primary"
- single <main id="main"> (hero..cta); main section[id] scroll-margin-top 88px
- footer h4->h2 (+ selector), decorative emoji aria-hidden (11 feature, 3 use-case), empty alt on logos beside text

## Tests (real)
- scripts/check-consistency.sh: PASSED (incl. Landing discovery contract). git diff --check: clean.
- Playwright Chromium headless shell 1243, script .fleet/reports/GH315D2-evidence/d2-assertions.js; only project raw-GitHub images allowed, all other network aborted.
  Viewports 320x700, 390x844, 768x1024, 769x1024, 820x1180, 1024x768, 1025x768, 1180x820, 1440x1000.
  Candidate PASS 9/9; base 5f99c25 FAIL 9/9 (negative control: hidden nav, no main, heading skip, footer h4, icons exposed, anchors).
- Screenshots (top/anchor/footer per viewport) opened and inspected: 320, 390, 769, 1025, 1440 top; 390+1440 footer; 1440 license anchor. Footer visuals unchanged.

## Findings (out of D2 scope, not fixed)
- F1 docs/index.html License section: secondary buttons "Community Notice" / "Governance Record" render white text on the near-white section background (invisible). Pre-existing on base 5f99c25 and main. Needs its own bounded ticket.

## NOT RUN / open
- GitHub CI on #372 (pending at report time).
- Kimi independent read-only review: dispatched (GH315D2-REVIEW-KIMI).
- Final human integration visual review: owed to Codex. No deploy/live PASS claimed. Full AA, design-system consolidation, live identity stay separate (#315).
