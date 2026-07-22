#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(git -C "$(dirname "${BASH_SOURCE[0]}")/.." rev-parse --show-toplevel)

collect_packages() {
  git -C "$ROOT_DIR" ls-files --cached --others --exclude-standard -- '*.go' |
    while IFS= read -r source; do
      case "/$source/" in
        */node_modules/* | */vendor/*)
          continue
          ;;
      esac

      directory=${source%/*}
      if [[ "$directory" == "$source" ]]; then
        printf '.\n'
      else
        printf './%s\n' "$directory"
      fi
    done |
    LC_ALL=C sort -u
}

packages=()
while IFS= read -r package; do
  [[ -n "$package" ]] && packages+=("$package")
done < <(collect_packages)

if [[ ${#packages[@]} -eq 0 ]]; then
  echo "no repository-owned Go packages found" >&2
  exit 1
fi

if [[ ${1:-} == "--list" ]]; then
  printf '%s\n' "${packages[@]}"
  exit 0
fi

if [[ $# -eq 0 ]]; then
  echo "usage: $0 --list | <command> [arguments...]" >&2
  exit 2
fi

exec "$@" "${packages[@]}"
