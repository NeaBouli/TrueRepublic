# GH-356 candidate UI evidence

Baseline: `ff63e9d087f1bbf09c01b3182b896ce30ac9116c` (main, PR #357).
Issue #356 was open, with no implementation PR found before this work.
This is unmerged candidate evidence, not rollout or production qualification.

## Changes and boundaries

- `CreateWallet.tsx`: one/two/three responsive recovery columns, nonshrinking
  indices, contained word cells, and a 44px minimum confirmation target.
- `MobileNav.tsx`: omit the floating menu on `/create`, so it cannot cover
  recovery words. Navigation and wallet handlers on other routes are unchanged.
- `AccountInfo.tsx`: allow the displayed address to wrap; keep the copy target
  nonshrinking and at least 44x44px. The full-address clipboard handler is unchanged.
- `WalletDashboard.tsx`: wrap the header, bound the logo to 40x40px, provide
  44x44px header controls, and allow grid columns to shrink.
- `src/assets/logo.png`: byte-identical copy of the existing repository asset,
  bundled by Vite and included by the existing Docker `COPY src` instruction.
  SHA-256: `e7d0d76fb35464d6d429600d502f36c959047a534032d17fbabf8014be5611fc`.

No services, stores, authentication, cryptography, signing, custody, ZKP, dependency
locks, public status, TODO or Bridge state are changed. No merge or deploy is claimed.

## Regression evidence

The test-only HTML entry renders actual components and production styles, with
synthetic in-memory state and no chain requests. Its repeated 24-word fixture is
intentionally invalid and never a wallet key. Wallet creation and network actions
are stubbed only in this entry; no production test seam is added. Neither its HTML
nor its fixture state appears in `dist`, and Docker does not copy it.

On the original components, all eight Chromium width/layout tests failed:

| Defect | Before | After |
| --- | --- | --- |
| Recovery columns | Word exceeds its cell right edge by 47.11px at 320px; FAB is present during recovery | Every word contained, pairwise overlap absent, no recovery FAB |
| Address / copy | Address exceeds its row right edge by 21.58px at 320px | Address contained and separated from copy; copy target 44x44px |
| Header logo | Restoring the external URL makes the dedicated load test fail (`naturalWidth > 0` remains false under the external-request guard) | Local image decodes; production manifest references it and HTTP returns the exact PNG bytes |

The negative logo mutation was reverted before final checks. The final full
browser suite passes 88 tests with two existing mobile physical-keyboard skips.
All 45 new regression cases pass. Chromium, Firefox and WebKit each cover
320, 375, 768 and 1280px widths at 900px height; Chromium and WebKit mobile
profiles run the same cases additionally. Assertions cover document/element
horizontal overflow, parent containment, clipping ancestors, pairwise word and
address/control overlap, visible hit targets after scrolling, keyboard focus order
and 44x44px targets. Each layout case attaches a full-page screenshot and numeric
geometry JSON. Playwright's JSON report and CI upload preserve these attachments.
A measured WebKit 320px wallet has document width/scrollWidth 320/320, logo 40x40,
header targets 44x44, address width/scrollWidth 154/154 and copy target 44x44.

## Local checks, 2026-10-04

Node 22.22.2, npm 10.9.7; lockfile-driven `npm ci`.

- `npm run lint`: PASS.
- `./node_modules/.bin/tsc -b --noEmit`: PASS.
- `./node_modules/.bin/tsc -p browser-quality/tsconfig.json`: PASS.
- `npm test -- --run`: 19 Node + 310 Vitest PASS; four existing integration skips.
- `npm run build` / bundle budget: PASS.
- `npm run audit:high`: no high or critical advisories.
- `./node_modules/.bin/playwright test --workers=2`: 88 PASS, two existing skips.
- `./scripts/check-consistency.sh` (Go 1.26.6): PASS.
- `git diff --check`: PASS.

This managed environment required temporary extracted Linux libraries in the
Playwright browser cache. Firefox used `MOZ_DISABLE_CONTENT_SANDBOX=1` and
`MOZ_DISABLE_RDD_SANDBOX=1`; WebKit's host-library preflight was bypassed with
`PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS=1` after supplying the actual libraries.
The browser tests themselves ran and passed; no assertion was skipped or weakened.
These local accommodations are not committed to configuration or CI. Protected CI
uses its existing `playwright install --with-deps` on Ubuntu and still needs to pass
on the PR head. The PR remains draft pending that evidence and review.
