#!/usr/bin/env bash
# Initialize a TrueRepublic PoD node from its generated CometBFT key.
# Usage: ./scripts/init-node.sh

set -euo pipefail

BINARY="${BINARY:-truerepublicd}"
CHAIN_ID="${CHAIN_ID:-truerepublic-1}"
MONIKER="${MONIKER:-truerepublic-node}"
CHAIN_HOME="${CHAIN_HOME:-$HOME/.truerepublic}"
DENOM="upnyx"

echo "============================================"
echo "  TrueRepublic Node Initialization"
echo "  Chain ID:  ${CHAIN_ID}"
echo "  Moniker:   ${MONIKER}"
echo "  Home:      ${CHAIN_HOME}"
echo "============================================"

# The daemon's init command is the only supported bootstrap boundary. It binds
# the generated CometBFT Ed25519 key to a positive-power PoD validator and
# creates its exact cap-checked bank backing. TrueRepublic does not wire
# x/staking, so this wrapper must never create accounts or gentxs itself.
"$BINARY" init "$MONIKER" --chain-id "$CHAIN_ID" --home "$CHAIN_HOME"

# Set minimum gas price in app.toml
if [ -f "${CHAIN_HOME}/config/app.toml" ]; then
    sed -i.bak "s/minimum-gas-prices = \"\"/minimum-gas-prices = \"1000${DENOM}\"/" \
        "${CHAIN_HOME}/config/app.toml"
fi

# Enable Prometheus metrics in config.toml
if [ -f "${CHAIN_HOME}/config/config.toml" ]; then
    sed -i.bak 's/prometheus = false/prometheus = true/' \
        "${CHAIN_HOME}/config/config.toml"
fi

echo ""
echo "Node initialized with generated-key, bank-backed PoD genesis."
echo "Start the node with: $BINARY start --home $CHAIN_HOME"
