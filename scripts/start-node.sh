#!/bin/bash
# Start a TrueRepublic blockchain node.
# Usage: ./scripts/start-node.sh

set -euo pipefail

BINARY="${BINARY:-truerepublicd}"
CHAIN_HOME="${CHAIN_HOME:-$HOME/.truerepublic}"
LOG_LEVEL="${LOG_LEVEL:-info}"

# Set seeds and persistent peers from environment if provided
if [ -n "${SEEDS:-}" ] && [ -f "${CHAIN_HOME}/config/config.toml" ]; then
    sed -i.bak "s/^seeds = .*/seeds = \"${SEEDS}\"/" \
        "${CHAIN_HOME}/config/config.toml"
fi

if [ -n "${PERSISTENT_PEERS:-}" ] && [ -f "${CHAIN_HOME}/config/config.toml" ]; then
    sed -i.bak "s/^persistent_peers = .*/persistent_peers = \"${PERSISTENT_PEERS}\"/" \
        "${CHAIN_HOME}/config/config.toml"
fi

echo "Starting TrueRepublic node..."
echo "  Home:      ${CHAIN_HOME}"
echo "  Log level: ${LOG_LEVEL}"

exec $BINARY start \
    --home "$CHAIN_HOME" \
    --log_level "$LOG_LEVEL"
