#!/bin/sh
set -eu

node_home="${TRUEREPUBLIC_HOME:-${HOME}/.truerepublic}"
explicit_home=false
previous=""

for argument in "$@"; do
  if [ "${previous}" = "--home" ]; then
    node_home="${argument}"
    explicit_home=true
    previous=""
    continue
  fi
  case "${argument}" in
    --home=*)
      node_home="${argument#--home=}"
      explicit_home=true
      ;;
  esac
  previous="${argument}"
done

if [ "${1:-}" = "start" ] && [ ! -f "${node_home}/config/genesis.json" ]; then
  if [ -z "${BOOTSTRAP_OPERATOR:-}" ]; then
    echo "Error: BOOTSTRAP_OPERATOR is required for first-start initialization." >&2
    echo "Set BOOTSTRAP_OPERATOR to the independently controlled account address." >&2
    exit 1
  fi
  truerepublicd init "${MONIKER:-truerepublic-node}" \
    --chain-id "${CHAIN_ID:-truerepublic-1}" \
    --bootstrap-operator "${BOOTSTRAP_OPERATOR}" \
    --home "${node_home}"
fi

if [ "${1:-}" = "start" ] && [ -f "${node_home}/config/config.toml" ]; then
  case "${PROMETHEUS_ENABLED:-false}" in
    true|1)
      sed -i.bak 's/^prometheus = false/prometheus = true/' \
        "${node_home}/config/config.toml"
      ;;
    false|0|"")
      sed -i.bak 's/^prometheus = true/prometheus = false/' \
        "${node_home}/config/config.toml"
      ;;
    *)
      echo "Error: PROMETHEUS_ENABLED must be true, false, 1, or 0." >&2
      exit 1
      ;;
  esac
  sed -i.bak 's|^[[:space:]]*prometheus_listen_addr = ".*"|prometheus_listen_addr = "127.0.0.1:26660"|' \
    "${node_home}/config/config.toml"
  rm -f "${node_home}/config/config.toml.bak"
fi

if [ "${explicit_home}" = true ]; then
  exec truerepublicd "$@"
fi
exec truerepublicd "$@" --home "${node_home}"
