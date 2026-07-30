#!/usr/bin/env bash
# Initialize a TrueRepublic PoD node from its generated CometBFT key.
# Usage: ./scripts/init-node.sh

set -euo pipefail

BINARY="${BINARY:-truerepublicd}"
CHAIN_ID="${CHAIN_ID:-truerepublic-1}"
MONIKER="${MONIKER:-truerepublic-node}"
CHAIN_HOME="${CHAIN_HOME:-$HOME/.truerepublic}"
DENOM="upnyx"
BOOTSTRAP_OPERATOR="${BOOTSTRAP_OPERATOR:-}"

echo "============================================"
echo "  TrueRepublic Node Initialization"
echo "  Chain ID:  ${CHAIN_ID}"
echo "  Moniker:   ${MONIKER}"
echo "  Home:      ${CHAIN_HOME}"
echo "  Bootstrap Operator: ${BOOTSTRAP_OPERATOR}"
echo "============================================"

if [ -z "${BOOTSTRAP_OPERATOR}" ]; then
  echo "Error: BOOTSTRAP_OPERATOR is required for node initialization." >&2
  echo "Set BOOTSTRAP_OPERATOR to a valid bech32 account address for bootstrap authority." >&2
  exit 1
fi

# The daemon's init command is the only supported bootstrap boundary. It binds
# the generated CometBFT Ed25519 key to a positive-power PoD validator and
# creates its exact cap-checked bank backing. TrueRepublic does not wire
# x/staking, so this wrapper must never create accounts or gentxs itself.
"$BINARY" init "$MONIKER" --chain-id "$CHAIN_ID" --home "$CHAIN_HOME" --bootstrap-operator "$BOOTSTRAP_OPERATOR"

# Set minimum gas price in app.toml
if [ -f "${CHAIN_HOME}/config/app.toml" ]; then
    sed -i.bak "s/minimum-gas-prices = \"\"/minimum-gas-prices = \"1000${DENOM}\"/" \
        "${CHAIN_HOME}/config/app.toml"
    # Application metrics share the existing loopback API listener. Hostname
    # and service labels stay disabled so node identity and deployment topology
    # do not become metric cardinality.
    sed -i.bak \
        -e '/^\[telemetry\]/,/^\[/{s/^service-name = .*/service-name = "truerepublic"/;}' \
        -e '/^\[telemetry\]/,/^\[/{s/^enabled = .*/enabled = true/;}' \
        -e '/^\[telemetry\]/,/^\[/{s/^enable-hostname = .*/enable-hostname = false/;}' \
        -e '/^\[telemetry\]/,/^\[/{s/^enable-hostname-label = .*/enable-hostname-label = false/;}' \
        -e '/^\[telemetry\]/,/^\[/{s/^enable-service-label = .*/enable-service-label = false/;}' \
        -e '/^\[telemetry\]/,/^\[/{s/^prometheus-retention-time = .*/prometheus-retention-time = 60/;}' \
        "${CHAIN_HOME}/config/app.toml"
    if ! awk '
        $0 == "[telemetry]" { in_telemetry = 1; next }
        in_telemetry && /^\[/ { in_telemetry = 0 }
        in_telemetry && $0 == "service-name = \"truerepublic\"" { service_name = 1 }
        in_telemetry && $0 == "enabled = true" { enabled = 1 }
        in_telemetry && $0 == "enable-hostname = false" { hostname = 1 }
        in_telemetry && $0 == "enable-hostname-label = false" { hostname_label = 1 }
        in_telemetry && $0 == "enable-service-label = false" { service_label = 1 }
        in_telemetry && $0 == "prometheus-retention-time = 60" { retention = 1 }
        END {
            exit !(service_name && enabled && hostname && hostname_label &&
                service_label && retention)
        }
    ' "${CHAIN_HOME}/config/app.toml"; then
        echo "Error: failed to configure the complete application telemetry block." >&2
        exit 1
    fi
    rm -f "${CHAIN_HOME}/config/app.toml.bak"
fi

# Enable Prometheus metrics in config.toml
if [ -f "${CHAIN_HOME}/config/config.toml" ]; then
    sed -i.bak 's/prometheus = false/prometheus = true/' \
        "${CHAIN_HOME}/config/config.toml"
    sed -i.bak 's/prometheus_listen_addr = ":[0-9][0-9]*"/prometheus_listen_addr = "127.0.0.1:26660"/' \
        "${CHAIN_HOME}/config/config.toml"
fi

echo ""
echo "Node initialized with generated-key, bank-backed PoD genesis."
echo "Start the node with: $BINARY start --home $CHAIN_HOME"
