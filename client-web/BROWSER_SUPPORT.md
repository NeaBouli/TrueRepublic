# Maintained browser support

The maintained `client-web` release candidate is qualified against the
Chromium, Firefox, and WebKit engine builds pinned by the lockfile's Playwright
version. Desktop and mobile-engine profiles are tested, with responsive
overflow checks at 320, 375, 768, and 1280 CSS pixels. Updating Playwright is
an explicit browser-support update and requires this complete gate to pass.

The automated boundary covers only the safe, unauthenticated `/unlock`,
`/create`, and `/import` surfaces. Every profile checks serious/critical WCAG
findings, responsive overflow, delayed lazy-route loading, and rejection of
third-party requests. Physical-keyboard reachability and visible focus run on
the desktop engine profiles; the mobile profiles intentionally skip that
desktop-input assertion. The suite does not create or import a wallet, unlock
stored material, sign or broadcast a transaction, or qualify an authenticated
production flow.

Run the reproducible gate from this directory:

```bash
npm ci
npx playwright install chromium firefox webkit
npm run test:browser
```

An engine/version outside the pinned Playwright engine builds is best-effort until
it is reproduced by this gate. The protected Linux CI run is authoritative for
all five projects; a local host that Playwright no longer supports cannot
substitute for or waive that matrix. Passing this browser check is quality
evidence; it is not a production-readiness or wallet-custody approval.
