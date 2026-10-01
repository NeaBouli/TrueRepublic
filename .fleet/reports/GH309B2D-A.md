id: GH309B2D-A
status: ok
worker: claude (interactive stand-in, sole owner; read-only on exact d7e9249, no product diff, nothing executed, no browser storage read)
summary: Wiring map for identity custody. Today the plaintext identity is a global zustand-persisted value that hydrates on every reload regardless of wallet lock, is never cleared by lock/switch/delete, and can be created in three UI places. The vault (PR #338) and migration (PR #339) exist but nothing calls them. Recommended staging: B2D1 = store/session wiring with no visible UI change except "identity absent while locked"; B2D2 = visible recovery/notice UI, creation/import disabled, file export, visual gate; deletion workflow only if a delete UI is wanted (none exists today).
ownership (file::symbol -> current role -> owner block):
  - stores/walletStore.ts::beginWalletSessionTransition (:59, module-private walletSessionGeneration :57) -> bumps generation + WalletService.invalidateSigningSession; called synchronously at the start of createWallet :149, importWallet :189, switchWallet :229, lock :291, deleteWallet (current only) :271 -> B2D1 hook point
  - stores/walletStore.ts::createWallet/importWallet/switchWallet -> after `await WalletService.*` and a generation check, set({password, currentWallet, isLocked:false}) (:168-175, :208-215, :249-254); unlock :304 = switchWallet(wallets[0]) -> B2D1 post-unlock hook
  - stores/walletStore.ts persist partialize {wallets} (:500-503): password/currentWallet never persisted; reload starts locked -> unchanged
  - stores/walletStore.ts::deleteWallet (:266) sync, no password, no UI caller anywhere (grep: only the store/service) -> B2D3/decision
  - stores/identityStore.ts (:23-82) zustand persist key 'identity-store' partialize {identity} = plaintext secret; createIdentity :32 (ZKPService.generateIdentity), importIdentity :62, exportIdentity :56, clearIdentity :49 -> B2D1 rewrite to in-memory session state
  - services/identityVault.ts::IdentityVault (#338), services/identityMigration.ts::LegacyIdentityMigration.inspect/detect/migrate (#339) -> called from B2D1 session wiring
  - UI creators: components/zkp/IdentitySetup.tsx::handleCreate :25 / handleImport :37 / handleExport clipboard :30 (mounted by VotingPanel.tsx :153 and IdentityManager.tsx :50, route /identity routes.tsx:113); components/membership/OnboardingFlow.tsx::handleCreateIdentity :55 -> B2D2
  - UI readers: VotingPanel.tsx :29, MembershipBadge.tsx :18-26 (identity?.commitment into loadMembership), OnboardingFlow.tsx :23/:49 -> B2D2 (read the in-memory state; no behaviour change)
  - lock UI: components/wallet/AccountInfo.tsx::handleLock :27 (lock(); navigate('/unlock')); unlock UI: components/auth/UnlockWallet.tsx :17/:28 -> unchanged
call_sequence (target):
  1. unlock/create/import/switch: beginWalletSessionTransition() synchronously also calls identitySession.invalidate() (identity=null, status='locked', session token bumped) BEFORE the first await. No await may precede it.
  2. After the store sets {password, currentWallet, isLocked:false} with generation g: identitySession.load({address, password, generation:g}) (fire-and-forget, errors contained). It runs LegacyIdentityMigration.detect(isCurrent) (only DETECTED/QUARANTINED marker writes), then IdentityVault.openIdentity(address, password).
  3. isCurrent = () => walletSessionGeneration === g && currentWallet?.address === address && !isLocked; needs a new exported walletStore::isWalletSessionCurrent(g, address) because the generation is module-private.
  4. Before every set() after an await, identitySession re-checks isCurrent; stale results are dropped. The decrypted Identity lives only in non-persisted zustand state (no persist middleware on the new store); it is never written to localStorage.
  5. Legacy present + marker DETECTED/PENDING: status 'legacy-pending' and a notice; the user action "Encrypt into this wallet" calls LegacyIdentityMigration.migrate({address, password, isCurrent}); VERIFIED -> openIdentity -> 'ready'. Plaintext stays (policy).
  6. lock/switch/delete-current/reload: step 1 clears the reference synchronously; reload never hydrates an identity (persist removed); the legacy 'identity-store' key is left untouched on disk (policy) and is read only by the migration core.
state_error_matrix (status -> trigger -> visible message/actions):
  - locked -> wallet locked or reload -> "Unlock your wallet to use your preview identity." (no identity data shown)
  - absent -> unlocked, no vault record, no legacy -> "No preview identity. Creating identities is disabled in this preview (issue #309)."
  - legacy-pending -> legacy valid, marker DETECTED/PENDING -> "An unencrypted preview identity was found on this device. Encrypt it into this wallet?" [Encrypt] [Export file] [Not now]
  - quarantined(reason) -> marker QUARANTINED -> "A stored preview identity could not be verified (<reason>). It is kept unchanged and is not used." [Export raw file]
  - error(vault-conflict|legacy-changed|verify-mismatch|vault-corrupt) -> marker ERROR or vault 'corrupt' -> bounded text per code, nothing used, [Export raw file]; wallet stays unlocked
  - wrong-password/locked vault -> IdentityVaultError 'locked' -> "The identity could not be opened with this wallet's password." (wallet unlock itself unaffected)
  - storage/no Web Locks -> 'storage' -> "Secure identity storage is unavailable in this browser." no writes
  - ready -> VERIFIED or vault record opened -> identity available in memory; export only as file with "preview, non-canonical, keep offline" warning
  - session-changed -> any stale async result -> silently dropped (no message)
writable_sets:
  - B2D1 (no visible UI change beyond 'locked/absent'): client-web/src/stores/identityStore.ts (+ new identityStore.test.ts), stores/walletStore.ts (invalidate call in beginWalletSessionTransition, post-unlock load, exported isWalletSessionCurrent; + walletStore.test.ts), optional services/identitySession.ts (+test) if the orchestration should stay out of the store. No component edits.
  - B2D2 (visible): components/zkp/IdentitySetup.tsx (create/import disabled, file export), IdentityManager.tsx (status notices, no clearIdentity of plaintext), VotingPanel.tsx (no creation entry), new components/zkp/IdentityCustodyNotice.tsx (+tests), components/membership/OnboardingFlow.tsx identity step (after #337 lands), one browser-quality spec + screenshots.
  - B2D3 (only if decided): wallet deletion UI + async password-confirmed delete gated on vault.removeIdentity after a successful explicit export.
tests_required:
  - unit: lock/switch/create/import/delete-current clear identity synchronously (assert state null before any awaited promise resolves); stale load after lock/switch never sets identity (deferred-promise interleaving); reload (fresh store) shows locked and never reads 'identity-store' into state; localStorage contains no secret after load/migrate (scan all keys); legacy-pending -> migrate -> ready; quarantined/error codes map to the bounded messages; wrong password leaves wallet unlocked; no Web Locks -> storage message; creation/import entry points absent; ZKPService.generateIdentity never called from UI (spy); MembershipService.registerIdentity never called.
  - browser/visual (B2D2, 1440x1000, 1180x820, 820x1180, 390x844, Chromium locally + Firefox/WebKit hosted): states locked, absent, legacy-pending, quarantined, error, ready-with-export-warning; seeded only with synthetic legacy records in a fresh profile; no overflow/clipping, 44px controls, keyboard order, contrast >= 4.5, zero broadcast/forbidden RPC; screenshots inspected.
collisions:
  - #337 (OnboardingFlow.tsx, its test, membership.ts, browser spec): its visual contract reaches the submit step by clicking "Create Anonymous Identity" (OnboardingFlow :55). Disabling creation changes that flow, so the OnboardingFlow part of B2D2 must stack on #337 and rewrite that contract (target state becomes "identity creation unavailable"), or creation stays in onboarding until #337 is merged. d7e9249 is based on main and does NOT contain #337's OnboardingFlow changes.
  - #338 (custodyEnvelope, identityVault, wallet.ts): B2D1 uses only the public IdentityVault API; no edit expected.
  - #339 (identityMigration, identityVault runExclusive, previewIdentityHash, zkp.ts): B2D1 uses LegacyIdentityMigration as is; zkp.ts generateIdentity untouched (only no longer reachable from UI).
  - walletStore.ts is untouched by #337-#339 -> no textual collision for B2D1.
security: non-none — custody lifecycle of user-held preview secrets; mandatory Codex review before implementation.
risks: zustand persist removal must not write back an empty 'identity-store' value (verify no persist middleware remains on the key, otherwise a hydrate/rehydrate could overwrite legacy plaintext); fire-and-forget load must not leak rejections; same-origin script exposure while unlocked remains.
decisions:
  - Gio: (1) Should a wallet-deletion UI exist in this preview at all? None exists today; without it B2D3 is unnecessary and the store-level deleteWallet should simply refuse while a vault record exists. (2) Should identity import (pasting exported JSON) be disabled together with creation until GH309C? Proposed: yes — import introduces a new preview identity just like creation.
  - Codex (technical): clipboard export removed in favour of file download (proposed); migration prompt after unlock vs only on /identity (proposed: non-blocking notice after unlock, action on /identity); placement of the orchestration (identityStore vs services/identitySession.ts).
next: Codex review of this map and the two Gio decisions; then a B2D1 brief.
