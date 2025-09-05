#!/usr/bin/env bash
set -euo pipefail

# move-expanded-step-to-done.sh
# Moves a delimited step from Next in blueprint/implementation.md to blueprint/done/
# and adds a one-liner link under the Done section.
#
# Requirements:
# - Steps in Next are delimited by a single line with only dashes (e.g., "-----")
#   immediately before the step header line:
#     -----
#     25. **Rate limiting & limits**
# - The step block ends at the next delimiter line (-----) or EOF.
#
# Usage:
#   tools/move-expanded-step-to-done.sh <step-number>
#
# Behavior:
# - Extracts the step block for the given number
# - Detects the step title and enclosing Phase header
# - Writes full block to blueprint/done/<phase>-step-<n>-<kebab>.md
# - Removes the block from implementation.md
# - Adds a one-liner link to the Done section

PLAN="blueprint/implementation.md"
DONE_DIR="blueprint/done"
DELIM='^-----\s*$'

die() { echo "Error: $*" >&2; exit 1; }

[ $# -ge 1 ] || die "Usage: $0 <step-number>"
STEP="$1"
[[ "$STEP" =~ ^[0-9]+$ ]] || die "Step must be a number"

# Use a repo-local temp file instead of mktemp (sandbox safe)
tmpplan="tools/.move-step-tmp.$$"
cp "$PLAN" "$tmpplan" || die "cannot copy plan"

# Work only in the Next section
next_start=$(awk '/^## Next/{print NR; exit}' "$tmpplan") || true
[ -n "$next_start" ] || die "Could not find '## Next' section"

# Find all delimiter lines (three or more dashes) in Next
delims=$(awk -v S="$next_start" 'NR>=S && $0 ~ /^-+\s*$/{print NR}' "$tmpplan")
[ -n "$delims" ] || die "No delimiters found in Next. Add '-----' before the step."

start_line=""
title_line=""

while read -r d; do
  # The next non-empty line after delimiter should be the step header
  hdr=$(awk -v D="$d" 'NR>D { if ($0 ~ /\S/) {print NR ":" $0; exit} }' "$tmpplan")
  [ -n "$hdr" ] || continue
  hdr_line=${hdr%%:*}
  hdr_text=${hdr#*:}
  if echo "$hdr_text" | grep -q "^${STEP}\. \*\*"; then
    start_line=$d
    title_line=$hdr
    break
  fi
done <<< "$delims"

[ -n "$start_line" ] || die "Could not find step $STEP after a delimiter in Next."

# Determine end line = last line before next delimiter after header, or EOF
hdr_line=${title_line%%:*}
end_line=$(awk -v S="$hdr_line" 'NR>S && $0 ~ /^-+\s*$/{print NR-1; exit} END{print NR}' "$tmpplan")
[ -n "$end_line" ] || die "Could not determine end of step block"

# Extract block (exclude the delimiter, start from header line)
block=$(awk -v A="$hdr_line" -v B="$end_line" 'NR>=A && NR<=B{print}' "$tmpplan")

# Parse title and kebab
raw_title=${title_line#*:}              # e.g., "25. **Rate limiting & limits**"
title=$(echo "$raw_title" | sed -E 's/^([0-9]+)\. \*\*(.*)\*\*.*/\2/')
kebab=$(echo "$title" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9]+/-/g; s/^-+//; s/-+$//')

# Detect phase from nearest above "### Phase X" header
phase=$(awk -v A="$start_line" 'NR<A && $0 ~ /^### Phase [A-Z]/{ph=$0} END{print ph}' "$tmpplan" | sed -E 's/^### Phase ([A-Z]).*/\1/' | tr '[:upper:]' '[:lower:]')
[ -n "$phase" ] || phase="x"
phase_tag="phase-${phase}"

mkdir -p "$DONE_DIR"
donefile="$DONE_DIR/${phase_tag}-step-${STEP}-${kebab}.md"

# Write Done file with date header + original block
date_str=$(date +%F)
{
  echo "### Step ${STEP} — ${title} (Done: ${date_str})"
  echo ""
  echo "$block"
} > "$donefile"

# Remove the original block including its leading delimiter
awk -v A="$start_line" -v B="$end_line" 'NR<A || NR>B{print}' "$tmpplan" > "$PLAN" || die "failed to update plan"

# Insert one-liner link under Done section if missing
link_line="- [Step ${STEP} — ${title}](${donefile})"
if ! grep -Fq "$link_line" "$PLAN"; then
  # Find end of existing Done bullet list (after the last '- [Step ...]')
  awk -v LL="$link_line" '
    BEGIN{added=0}
    {
      print $0
      if (!added && $0 ~ /^- \[Step [0-9]+/){last=NR}
    }
    END{
      # Post-process not supported here; do a second pass
    }
  ' "$PLAN" > "$PLAN.tmp"
  # Insert after last existing bullet; if none, place after the Done section header
last_bullet=$(awk '/^## Done/{d=1} d && $0 ~ /^- \[Step /{lb=NR} END{print lb+0}' "$PLAN")
  if [ "$last_bullet" -gt 0 ]; then
    awk -v L="$last_bullet" -v LL="$link_line" 'NR==L{print; print LL; next} {print}' "$PLAN" > "$PLAN.tmp2"
    mv "$PLAN.tmp2" "$PLAN"
  else
    # Insert right after the Done section header line
    done_hdr=$(awk '/^## Done/{print NR; exit}' "$PLAN")
    if [ -n "$done_hdr" ]; then
      awk -v H="$done_hdr" -v LL="$link_line" 'NR==H{print; print LL; next} {print}' "$PLAN" > "$PLAN.tmp2"
      mv "$PLAN.tmp2" "$PLAN"
    fi
  fi
fi

echo "Moved step ${STEP} to ${donefile} and updated ${PLAN}."
