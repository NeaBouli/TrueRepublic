#!/bin/bash
# Start a TrueRepublic blockchain node.
# Usage: ./scripts/start-node.sh

set -euo pipefail

BINARY="${BINARY:-truerepublicd}"
CHAIN_HOME="${CHAIN_HOME:-$HOME/.truerepublic}"
LOG_LEVEL="${LOG_LEVEL:-info}"
NETWORK_ROLE="${NETWORK_ROLE:-}"

# Topology is reviewed state, not a shell-substitution boundary. Edit the
# generated TOML through the operator change procedure, then validate it.
if [ -n "${SEEDS:-}" ] || [ -n "${PERSISTENT_PEERS:-}" ]; then
    echo "Error: SEEDS/PERSISTENT_PEERS environment mutation is disabled." >&2
    echo "Edit the reviewed config.toml, then run network-policy validation." >&2
    exit 1
fi

if [ -z "${NETWORK_ROLE}" ]; then
    echo "Error: NETWORK_ROLE is required (seed, sentry, validator, rpc, or private)." >&2
    exit 1
fi

"${BINARY}" network-policy validate \
    --role "${NETWORK_ROLE}" \
    --home "${CHAIN_HOME}"

echo "Starting TrueRepublic node..."
echo "  Home:      ${CHAIN_HOME}"
echo "  Log level: ${LOG_LEVEL}"
echo "  Role:      ${NETWORK_ROLE}"

exec "${BINARY}" start \
    --home "$CHAIN_HOME" \
    --log_level "$LOG_LEVEL"
