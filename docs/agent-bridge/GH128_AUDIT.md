# GH-128 Client Route Splitting and Bundle Budget Audit

Date: 2026-08-08

## Scope and safety boundary

GH-128 preserves the canonical 21-route BrowserRouter contract while reducing
initial client delivery cost and making future bundle regressions fail CI. It
does not alter chain messages, the GH-115 transaction registry, simulation,
fees, signing, broadcast, RPC configuration, custody, keys, or production
deployment.

## Baseline

The exact `origin/main` baseline emitted one JavaScript entry of 1,719,256 raw
bytes and 322.63 kB gzip by Vite's reporter. Vite warned that the single chunk
exceeded 500 kB. Static page imports and the application-level MobileNav made
the route tree, wallet signing stack, Stargate transport, and protobuf registry
part of that entry.

## Implemented controls

- Nineteen page modules are dynamic `React.lazy` entries; the root and catch-all
  redirects remain eager and all 21 route paths, order, parameters, and query
  behavior are unchanged.
- A single accessible loading boundary is inside the existing outer
  `ErrorBoundary`; a rejected or stale route chunk produces the fail-closed
  recovery UI instead of a blank application.
- Wallet crypto, chain-query, and transaction services load only when their
  operations need them. This breaks the eager App/MobileNav import chain while
  preserving one cached service instance and the existing behavior.
- Vite emits a manifest. A dependency-free Node gate identifies the application
  and dynamic route entries independent of hashed filenames, measures raw and
  Node-zlib gzip bytes, and fails on missing artifacts, missing route entries,
  or entry/route/chunk/total regressions.
- Every `npm run build` runs the gate, so Client Web CI cannot bypass it. Docs CI
  now also runs for `client-web/**` changes.

## Measured result and budgets

| Measure | Result | Enforced ceiling |
|---|---:|---:|
| Initial entry raw | 234,322 B | 260,000 B |
| Initial entry gzip | 75,786 B | 85,000 B |
| Lazy page entries | 19 | exactly 19 |
| Largest direct route gzip | 5,031 B | 7,000 B |
| Largest individual chunk raw | 1,057,764 B | 1,100,000 B |
| Largest individual chunk gzip | 134,769 B | 150,000 B |
| Total JavaScript raw | 1,733,603 B | 1,900,000 B |
| Total JavaScript gzip | 349,422 B | 390,000 B |

The largest raw chunk is deferred generated protobuf/signing code and compresses
to 134,769 bytes. Both raw and gzip ceilings remain enforced; Vite's reporting
threshold is aligned with the stricter project gate rather than being the only
control.

## Verification

- `npm run lint` — PASS.
- `npm test -- --run` — PASS: 94 Vitest cases, one explicitly skipped external
  case, and 10 Node audit/budget policy cases.
- `npm run build` — PASS: TypeScript, Vite, 19-route manifest, and all bundle
  budgets.
- `npm run audit:high` — PASS: no live High or Critical npm advisory.

## Independent review

Kimi K3 performed a bounded read-only import-graph and CI review. It identified
the otherwise cosmetic MobileNav-to-CosmJS eager path and recommended the
service-level split, emitted-artifact gate, and Docs CI trigger. A Spark worker
independently confirmed the minimal Suspense/ErrorBoundary regression strategy.
Sol owns and reviewed all edits and verification.

## Remaining risk and rollout boundary

- First visits to deferred routes require a network fetch. Missing/stale chunks
  fail closed through the recovery UI; no automatic reload loop is introduced.
- Total compressed JavaScript is slightly larger than the old single file due
  to splitting overhead, but the initial entry is approximately 76% smaller and
  total growth is bounded.
- Accessibility, broader low-bandwidth/browser qualification, wallet-key review,
  and production deployment remain separate open rollout work.
- This evidence is local until protected exact-head GitHub checks, merge, Pages,
  and final-main verification pass. Production readiness remains false.
