#!/bin/bash
# Restore a sanitized TrueRepublic chain-data backup into an initialized home.
# Usage: ./scripts/restore.sh <backup-file> <target-home>
#
# The target home must already be initialized with `truerepublicd init`.
# Restore intentionally preserves local node/validator keys and keyring files.

set -euo pipefail

if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <backup-file> <target-home>" >&2
    exit 2
fi

BACKUP_FILE="$1"
TARGET_HOME="$2"

if [ ! -f "$BACKUP_FILE" ]; then
    echo "Backup file not found: $BACKUP_FILE" >&2
    exit 1
fi

if [ ! -f "$TARGET_HOME/config/genesis.json" ]; then
    echo "Target home must be initialized before restore: $TARGET_HOME" >&2
    exit 1
fi

TEMP_DIR="$(mktemp -d)"
cleanup() {
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

tar -xzf "$BACKUP_FILE" -C "$TEMP_DIR"

SOURCE_HOME_COUNT="$(find "$TEMP_DIR" -mindepth 1 -maxdepth 1 -type d | wc -l | tr -d ' ')"
if [ "$SOURCE_HOME_COUNT" != "1" ]; then
    echo "Backup must contain exactly one chain-home directory" >&2
    exit 1
fi

SOURCE_HOME="$(find "$TEMP_DIR" -mindepth 1 -maxdepth 1 -type d | head -n 1)"

tar -C "$SOURCE_HOME" \
    --exclude "./config/node_key.json" \
    --exclude "./config/priv_validator_key.json" \
    --exclude "./data/priv_validator_state.json" \
    --exclude "./keyring-file" \
    --exclude "./keyring-test" \
    --exclude "./keyring-test-*" \
    -cf - . | tar -C "$TARGET_HOME" -xf -

echo "[$(date)] Restore complete: $TARGET_HOME"
