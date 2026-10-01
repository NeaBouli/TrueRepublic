//! QUARANTINED PROTOTYPE — NON-PRODUCTION, NOT DEPLOYABLE (issue #308).
//!
//! `truerepublic-contracts` is historical prototype code kept only for host-side tests.
//! Audit findings: TRR-13 (treasury.rs), TRR-16 (governance.rs), TRR-18. Disposition: archive. It must never be built for
//! `wasm32`, stored, instantiated or migrated on any chain; see
//! `contracts/QUARANTINE.md`.

#[cfg(target_arch = "wasm32")]
compile_error!(
    "QUARANTINED (TrueRepublic #308): `truerepublic-contracts` is a non-production prototype and must not be built for wasm32 or deployed; see contracts/QUARANTINE.md"
);

pub mod governance;
pub mod treasury;
