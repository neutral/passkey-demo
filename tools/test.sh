#!/usr/bin/env bash
set -euo pipefail

mode="${1:-short}"
repo_root="$(cd "$(dirname "$0")/.." && pwd)"
server_dir="$repo_root/server"

all_packages=( $(go -C "$server_dir" list ./...) )

# Curated set for fast feedback in short mode
short_packages=(
  "github.com/neutral/passkey-demo/internal/config"
  "github.com/neutral/passkey-demo/internal/encoding"
  "github.com/neutral/passkey-demo/internal/crypto"
  "github.com/neutral/passkey-demo/internal/types"
  "github.com/neutral/passkey-demo/internal/http"
  "github.com/neutral/passkey-demo/internal/vectors"
)

idx=0

now_ms() {
  if command -v python3 >/dev/null 2>&1; then
    python3 - <<'PY'
import time
print(int(time.time()*1000))
PY
  else
    # Fallback to seconds if ms not available
    date +%s000
  fi
}

run_pkg() {
  local pkg="$1"; shift
  idx=$((idx+1))
  start_ms=$(now_ms)
  echo "➡️  [$idx/$total] $pkg"
  go -C "$server_dir" test -count=1 "$@" "$pkg"
  end_ms=$(now_ms)
  dur=$((end_ms-start_ms))
  echo "   ⏱  ${dur}ms"
}

case "$mode" in
  short)
    echo "Running SHORT test suite (-short, curated packages)";
    packages=("${short_packages[@]}")
    total=${#packages[@]}
    for p in "${packages[@]}"; do
      run_pkg "$p" -short
    done
    ;;
  all)
    echo "Running FULL test suite (package-by-package)";
    packages=("${all_packages[@]}")
    total=${#packages[@]}
    for p in "${packages[@]}"; do
      run_pkg "$p"
    done
    ;;
  pkgs)
    shift || true
    if [[ $# -eq 0 ]]; then
      echo "Usage: $0 pkgs <pkg1> [pkg2 ...]" >&2; exit 2;
    fi
    packages=("$@")
    total=${#packages[@]}
    echo "Running selected packages: ${packages[*]}";
    for p in "${packages[@]}"; do
      run_pkg "$p"
    done
    ;;
  *)
    echo "Usage: $0 [short|all]" >&2
    exit 2
    ;;
esac

echo "✅ Done"
