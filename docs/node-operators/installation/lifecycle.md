# Verified Artifact Lifecycle

GH-222 provides a repository-owned, fail-closed lifecycle for a native
`truerepublicd` artifact. It covers isolated install, identity-bound status and
pre-start checks, compatible binary upgrade, one-step rollback, and safe
uninstall. It does not publish a release, approve production, migrate chain
state, deploy a service, or authorize real keys or funds.

The machine contract is
`configs/release/install-lifecycle.json`. It permits only `linux-amd64` and
`linux-arm64`, both with the `linux-glibc` runtime. The CLI also requires the
declared target to match the host running it. Use an absolute artifact path and
the SHA-256 and 40-character source commit from independently verified release
evidence.

## Operator state boundary

The managed prefix contains only the daemon binary and lifecycle metadata.
Pass the real operator state directory separately with `--operator-state`; it
must be outside, and must not contain, the managed prefix. The tool never reads,
moves, restores, or deletes chain data, configuration, keyrings, node keys,
validator keys, or `priv_validator_state.json`.

These examples use `/opt/truerepublic` for managed files and
`/var/lib/truerepublic` for operator state. Replace the artifact, target and
identity values with verified evidence:

```bash
export LIFECYCLE_CONTRACT="$PWD/configs/release/install-lifecycle.json"
export INSTALL_PREFIX=/opt/truerepublic
export OPERATOR_STATE=/var/lib/truerepublic
export TARGET=linux-amd64
export RUNTIME=linux-glibc
export SOURCE_REF=<40-lowercase-hex-commit>
export ARTIFACT=/absolute/path/to/truerepublicd-linux-amd64
export ARTIFACT_SHA256=<64-lowercase-hex-sha256>
```

Do not point `INSTALL_PREFIX` at `/usr`, `/usr/local`, an operator home, or any
shared directory. Install requires an absent or completely empty prefix and
rejects symlinks, traversal, partial state, unsafe executable permissions,
empty or oversized artifacts, and identity mismatches before mutation.
The lifecycle process must have exclusive access to a root-owned prefix parent;
do not run it where another local account can replace path components while a
check or atomic rename is in progress.

## Clean install

```bash
go run ./cmd/install-lifecycle \
  --contract "$LIFECYCLE_CONTRACT" \
  --prefix "$INSTALL_PREFIX" \
  --operator-state "$OPERATOR_STATE" \
  --sha256 "$ARTIFACT_SHA256" \
  --source-ref "$SOURCE_REF" \
  --target "$TARGET" \
  --runtime "$RUNTIME" \
  install --artifact "$ARTIFACT"
```

The binary is installed at `/opt/truerepublic/bin/truerepublicd`. Same-directory
temporary files, file and directory synchronization, atomic rename, and a final
digest check prevent a complete-looking unverified install.

## Status and pre-start gate

Use the exact identity currently expected to be installed:

```bash
go run ./cmd/install-lifecycle \
  --contract "$LIFECYCLE_CONTRACT" --prefix "$INSTALL_PREFIX" \
  --operator-state "$OPERATOR_STATE" --sha256 "$ARTIFACT_SHA256" \
  --source-ref "$SOURCE_REF" --target "$TARGET" --runtime "$RUNTIME" \
  status

go run ./cmd/install-lifecycle \
  --contract "$LIFECYCLE_CONTRACT" --prefix "$INSTALL_PREFIX" \
  --operator-state "$OPERATOR_STATE" --sha256 "$ARTIFACT_SHA256" \
  --source-ref "$SOURCE_REF" --target "$TARGET" --runtime "$RUNTIME" \
  pre-start
```

`pre-start` must succeed immediately before a service manager starts the
binary. Missing or malformed metadata, digest drift, a different source,
target or runtime, symlinks, unsafe permissions, and an interrupted transaction
marker all fail closed.

## Compatible upgrade

Record the current identity, then bind the command to the different candidate:

```bash
export OLD_SHA256="$ARTIFACT_SHA256"
export OLD_SOURCE_REF="$SOURCE_REF"
export ARTIFACT=/absolute/path/to/new/truerepublicd-linux-amd64
export ARTIFACT_SHA256=<new-64-lowercase-hex-sha256>
export SOURCE_REF=<new-40-lowercase-hex-commit>

go run ./cmd/install-lifecycle \
  --contract "$LIFECYCLE_CONTRACT" --prefix "$INSTALL_PREFIX" \
  --operator-state "$OPERATOR_STATE" --sha256 "$ARTIFACT_SHA256" \
  --source-ref "$SOURCE_REF" --target "$TARGET" --runtime "$RUNTIME" \
  upgrade --artifact "$ARTIFACT" \
  --expected-current-sha256 "$OLD_SHA256" \
  --expected-current-source-ref "$OLD_SOURCE_REF"

go run ./cmd/install-lifecycle \
  --contract "$LIFECYCLE_CONTRACT" --prefix "$INSTALL_PREFIX" \
  --operator-state "$OPERATOR_STATE" --sha256 "$ARTIFACT_SHA256" \
  --source-ref "$SOURCE_REF" --target "$TARGET" --runtime "$RUNTIME" \
  pre-start
```

The current binary is copied to the single rollback slot before replacement.
A later successful upgrade replaces that slot with the immediately preceding
verified binary. State migrations are not performed by this tool; if a
candidate may have opened or changed chain state, stop and follow the governed
upgrade and incident-command procedure instead of attempting binary rollback.
The post-upgrade `pre-start` invocation above must succeed before any service
restart.

## Rollback

Bind the command to the currently installed candidate identity:

```bash
go run ./cmd/install-lifecycle \
  --contract "$LIFECYCLE_CONTRACT" --prefix "$INSTALL_PREFIX" \
  --operator-state "$OPERATOR_STATE" --sha256 "$ARTIFACT_SHA256" \
  --source-ref "$SOURCE_REF" --target "$TARGET" --runtime "$RUNTIME" \
  rollback
```

Rollback verifies the snapshot against the manifest, restores it atomically,
and consumes the snapshot. Replaying the command fails. Run `pre-start` again
with the restored digest and source commit before starting the service.

## Safe uninstall

Uninstall first verifies the current identity and walks the full prefix. Any
unknown directory or file—including configuration, chain data or key-like
material—causes rejection before deletion. Only the exact managed binary,
manifest and optional rollback binary are removed; operator state and unrelated
paths remain untouched.

```bash
go run ./cmd/install-lifecycle \
  --contract "$LIFECYCLE_CONTRACT" --prefix "$INSTALL_PREFIX" \
  --operator-state "$OPERATOR_STATE" --sha256 "$ARTIFACT_SHA256" \
  --source-ref "$SOURCE_REF" --target "$TARGET" --runtime "$RUNTIME" \
  uninstall --expected-current-sha256 "$ARTIFACT_SHA256"
```

Empty managed directories may remain and can be removed only after an operator
confirms they are empty. The lifecycle tool never recursively deletes the
prefix.

## Interrupted operation

Every mutation writes a transaction marker before changing managed files and
removes it only after synchronized completion. If a process or host stops in
between, `status`, `pre-start`, and all mutations refuse to continue. Preserve
the prefix and logs for incident review; do not delete the marker or guess which
step completed. Prepare a new empty prefix from verified evidence, or follow an
explicit reviewed recovery plan.

See also [Offline Release Evidence](release-evidence.md),
[Governed Application Upgrades and Rollback](../operations/upgrades.md), and
[Backup & Recovery](../operations/backup-recovery.md).
