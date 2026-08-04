# Dependency risk acceptance

## React Router RSC advisory (temporary)

| Field | Decision |
|---|---|
| Advisory | `GHSA-qwww-vcr4-c8h2` |
| Affected dependency chain | `react-router-dom` 7.18.2 → `react-router` 7.18.2 |
| Severity | High |
| Accepted scope | `client-web` BrowserRouter SPA only |
| Owner | TrueRepublic maintainers |
| Review deadline | 2026-09-04, or before any public rollout (whichever comes first) |
| Removal condition | React 19.2.7+ and `react-router` 8.3.0+ migration, or an upstream React-18-compatible patch |

The advisory affects React Server Components/framework-mode server-action
handling. `client-web` is a client-only declarative SPA. It uses
`BrowserRouter`, `Routes`, and `Route`; it has no React Server Components,
framework-mode data router, server actions, route `action` handlers, or server
runtime. The vulnerable execution path is therefore not reachable in the
maintained application architecture.

The CI exception is deliberately fail-closed. `npm run audit:high` accepts only
the exact GitHub advisory URL above for the `react-router` package and only
dependency chains that terminate in that advisory. It also rejects router APIs
outside the reviewed declarative-SPA surface, so adding data-router, server, or
RSC capabilities invalidates this acceptance automatically. Any other high or
critical advisory, malformed audit result, or cyclic/unresolved dependency chain
fails the gate. Policy behavior is covered by Node tests and the live npm audit
still runs on every client CI and security scan.

This is not a general waiver for React Router or npm audit findings. Remove the
allowance as soon as the removal condition can be met without an unsafe forced
upgrade.
