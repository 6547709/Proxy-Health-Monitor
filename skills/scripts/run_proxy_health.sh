#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BIN_DEFAULT="$SKILL_DIR/bin/proxy-monitor"
BIN_PATH="${PROXY_MONITOR_BIN:-$BIN_DEFAULT}"

if [[ ! -x "$BIN_PATH" ]]; then
  echo "ERROR: proxy-monitor binary not found or not executable: $BIN_PATH" >&2
  echo "Set PROXY_MONITOR_BIN to the binary path, or place the binary at $BIN_DEFAULT" >&2
  exit 2
fi

URL="${1:-${SUBSCRIBE_URL:-}}"
if [[ -z "$URL" ]]; then
  echo "ERROR: SUBSCRIBE_URL is not set and no URL argument was provided" >&2
  exit 3
fi

exec "$BIN_PATH" -url "$URL" -json
