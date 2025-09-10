#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "$0")/.." && pwd)"

echo "Building vectors CLI..." >&2
go -C "$repo_root/server" build ./cmd/vectors

mkdir -p "$repo_root/specs/goldens"

echo "Writing golden JSON..." >&2
"$repo_root/server/vectors" -fmt json > "$repo_root/specs/goldens/tx-bundle-v1.json"

echo "Golden updated: specs/goldens/tx-bundle-v1.json" >&2

