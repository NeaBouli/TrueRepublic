# October Go advisory and bounded exception review

Status: implementation in progress; no release or production approval.
Owner authorization: Gio, 2026-10-10 Europe/Athens (2026-10-09 UTC).
Tickets: [GH-364](https://github.com/NeaBouli/TrueRepublic/issues/364) and
[GH-355](https://github.com/NeaBouli/TrueRepublic/issues/355).

## Fixed-version maintenance scope

The October baseline scanner reported twelve new reachable IDs:
GO-2026-6599, GO-2026-6600, GO-2026-6603, GO-2026-6605, GO-2026-6607, GO-2026-6608,
GO-2026-6609, GO-2026-6610, GO-2026-6611, GO-2026-6612, GO-2026-6613 and GO-2026-6617.
The candidate selects the supported Go 1.26.9 security release and x/net 0.60.0,
with only its required minimal-version-selection graph. x/net itself requires Go 1.26.0.
See the [official Go release history](https://go.dev/doc/devel/release#go1.26.9)
and the linked ID records at `https://vuln.go.dev/ID/<ID>.json`.

All coupled toolchain, official image, build/verifier, release and synthetic fixture pins
must agree. The upgrade plan, wasmvm version, application logic and consensus state are unchanged.
No new advisory exception, scanner-policy relaxation or production artifact claim is permitted.

## Fresh upstream exception records

The official JSON records were fetched on 2026-10-09 UTC. SHA-256 binds the exact retrieved bytes.

| ID | Official record | SHA-256 | Fresh record state |
| --- | --- | --- | --- |
| GO-2023-1821 | [Go vulnerability database](https://vuln.go.dev/ID/GO-2023-1821.json) | `7f99523bd195ee09f336f24edc0db37e5f9c361dbafab59de892cf7ec65b55e0` | Cosmos SDK; no fixed event published |
| GO-2023-1881 | [Go vulnerability database](https://vuln.go.dev/ID/GO-2023-1881.json) | `43acdd490e5dda539f579ff7f4b5415555b1bf67695a42f7ca8ad7667d960dd8` | Cosmos SDK; no fixed event published |
| GO-2026-5932 | [Go vulnerability database](https://vuln.go.dev/ID/GO-2026-5932.json) | `f277b0400996200a7d5034c676661cd8cf18adeab1fdc5d0dc369e456ddf1fad` | x/crypto; no fixed event published |
| GO-2026-4740 | [Go vulnerability database](https://vuln.go.dev/ID/GO-2026-4740.json) | `3639d5ace70ea11fabc0adf026b8e13734ea67b78793fffd34a8df9e2a376615` | Withdrawn at 2026-10-01T16:21:21Z; last modified 2026-10-07T14:09:50Z |

Withdrawal is an upstream reclassification, not an application fix. The completed fresh scan
confirms GO-2026-4740 is absent and exactly GO-2023-1821, GO-2023-1881 and GO-2026-5932
remain reachable with empty fixed-version fields. The obsolete allowance is removed; the three
remaining IDs are renewed under Gio's exact authorization. UTC approval 2026-10-09 and expiry
2026-10-20 are eleven days, within the unchanged 30-day maximum. The existing seven-day
review cadence requires another review by 2026-10-16. No new ID or policy bypass is added.

The effective MVS graph is bounded by x/net 0.60.0 requiring x/crypto 0.57.0, x/sys 0.48.0,
x/term 0.46.0 and x/text 0.42.0; x/text additionally requires x/mod 0.41.0 and x/sync 0.23.0.
The downloaded verified module manifests all declare Go 1.26.0; tidy is idempotent.

## Verification and remaining acceptance

PASS: pinned govulncheck v1.6.0 with Go 1.26.9; completed symbol-level scan of all 33 maintained
package directories using database last-modified 2026-10-08T22:31:09Z. Exactly three no-fix IDs
remain and all twelve newly fixable IDs are absent. Raw scan SHA-256:
`be652dce938341c30cfcb132211bb4c4f511e97276fe12c7c95fbc34ea099799`.
Local policy replay consumes those completed bytes through the unmodified gate; it is not
another scanner run. Protected exact-head CI must independently run the real scanner again.

PASS: module verification/tidy idempotence, focused release/candidate/cross-run/tool-bootstrap
packages and six coupled script contracts before the final policy digest propagation. Test-only
Go WASM builds repeat byte-identically at 18,778,260 bytes, SHA-256
`d2cf62c436055de543562347f877554c631fbffbdafa26155c96c07d6dc67d66`.
Incoming GH300 test manifests are not imported; deliberate re-pin is needed when that stack restacks.

Final policy/fixture contracts, relevant full Go/race/runtime/WASM/OCI gates, independent review
and protected exact-head CI are not yet accepted. Codex Security NOT RUN; existing CI scans do
not substitute for that separately authorized gate. Rollout credit and production status are unchanged.
