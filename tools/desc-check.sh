#!/usr/bin/env bash
set -euo pipefail

# Usage: tools/desc-check.sh [base_ref] [head_ref]
# Default: compare last commit (HEAD~1..HEAD)

base_ref="${1:-HEAD~1}"
head_ref="${2:-HEAD}"

changed_files=$(git diff --name-only --diff-filter=ACMRT "$base_ref" "$head_ref")

fail=0
missing=()

is_source() {
  local f="$1"
  # Exclude tests, desc files, assets, generated dist
  [[ "$f" == *.desc.md ]] && return 1
  [[ "$f" == web/tests/* ]] && return 1
  [[ "$f" == web/dist/* ]] && return 1
  [[ "$f" == server/**/*.db* ]] && return 1
  # Source file extensions we care about
  [[ "$f" == web/src/*.ts || "$f" == web/src/**/*.ts || "$f" == web/src/*.tsx || "$f" == web/src/**/*.tsx || "$f" == server/*.go || "$f" == server/**/*.go ]]
}

needs_desc_update() {
  local f="$1"
  local desc="$f.desc.md"
  # If the source file changed and the .desc.md is not in the changed set, flag it
  if ! grep -qxF "$desc" <<< "$changed_files"; then
    # Allow if file truly has no desc file yet? No: policy requires presence and update.
    echo "$desc"
    return 0
  fi
  return 1
}

while IFS= read -r f; do
  [[ -z "$f" ]] && continue
  if is_source "$f"; then
    d=$(needs_desc_update "$f" || true)
    if [[ -n "$d" ]]; then
      missing+=("$f -> $d")
      fail=1
    fi
  fi
done <<< "$changed_files"

if [[ $fail -ne 0 ]]; then
  echo "Missing description updates for modified source files:" >&2
  for m in "${missing[@]}"; do
    echo "  - $m" >&2
  done
  echo "Please update the corresponding .desc.md files in the same commit." >&2
  exit 1
fi

echo "Description check passed: all modified source files have updated .desc.md files."

