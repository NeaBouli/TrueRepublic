#!/bin/sh
set -eu

node_home="${TRUEREPUBLIC_HOME:-${HOME}/.truerepublic}"
explicit_home=false
previous=""

for argument in "$@"; do
  if [ "${previous}" = "--home" ]; then
    node_home="${argument}"
    explicit_home=true
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
  truerepublicd init "${MONIKER:-truerepublic-node}" \
    --chain-id "${CHAIN_ID:-truerepublic-1}" \
    --home "${node_home}"
fi

if [ "${explicit_home}" = true ]; then
  exec truerepublicd "$@"
fi
exec truerepublicd "$@" --home "${node_home}"
