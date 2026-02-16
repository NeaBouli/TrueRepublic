#!/bin/bash
# Initialize a TrueRepublic blockchain node.
# Usage: ./scripts/init-node.sh

set -euo pipefail

BINARY="${BINARY:-truerepublicd}"
CHAIN_ID="${CHAIN_ID:-truerepublic-1}"
MONIKER="${MONIKER:-truerepublic-node}"
CHAIN_HOME="${CHAIN_HOME:-$HOME/.truerepublic}"
DENOM="pnyx"
GENESIS_AMOUNT="1000000000${DENOM}"
STAKING_AMOUNT="100000${DENOM}"

echo "============================================"
echo "  TrueRepublic Node Initialization"
echo "  Chain ID:  ${CHAIN_ID}"
echo "  Moniker:   ${MONIKER}"
echo "  Home:      ${CHAIN_HOME}"
echo "============================================"

# Initialize the node
$BINARY init "$MONIKER" --chain-id "$CHAIN_ID" --home "$CHAIN_HOME"

# Set minimum gas price in app.toml
if [ -f "${CHAIN_HOME}/config/app.toml" ]; then
    sed -i.bak "s/minimum-gas-prices = \"\"/minimum-gas-prices = \"0.001${DENOM}\"/" \
        "${CHAIN_HOME}/config/app.toml"
fi

# Enable Prometheus metrics in config.toml
if [ -f "${CHAIN_HOME}/config/config.toml" ]; then
    sed -i.bak 's/prometheus = false/prometheus = true/' \
        "${CHAIN_HOME}/config/config.toml"
fi

# Add genesis account
$BINARY keys add genesis-validator \
    --home "$CHAIN_HOME" \
    --keyring-backend test 2>&1 | tee "${CHAIN_HOME}/genesis-key.txt"

GENESIS_ADDR=$($BINARY keys show genesis-validator -a \
    --home "$CHAIN_HOME" \
    --keyring-backend test)

$BINARY genesis add-genesis-account "$GENESIS_ADDR" "$GENESIS_AMOUNT" \
    --home "$CHAIN_HOME"

# Create genesis transaction
$BINARY genesis gentx genesis-validator "$STAKING_AMOUNT" \
    --chain-id "$CHAIN_ID" \
    --home "$CHAIN_HOME" \
    --keyring-backend test \
    --moniker "$MONIKER"

# Collect genesis transactions
$BINARY genesis collect-gentxs --home "$CHAIN_HOME"

# Validate genesis
$BINARY genesis validate-genesis --home "$CHAIN_HOME"

echo ""
echo "Node initialized successfully!"
echo "Genesis validator address: ${GENESIS_ADDR}"
echo "Start the node with: ./scripts/start-node.sh"
