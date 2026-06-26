#!/bin/bash
# verify-parity.sh — Cross-check a completed (or in-progress) roll against the
# python / java / dotnet sibling clients, so nothing is silently missed.
#
# Usage:   bash .claude/skills/roll-playwright/verify-parity.sh <NEW_VERSION>
# Example: bash .claude/skills/roll-playwright/verify-parity.sh 1.61.1
#
# Run AFTER apply-patch + go generate, while the playwright/ submodule is on the
# new tag (the playwright-build branch is fine too). Reports three buckets:
#   1. New upstream APIs (`* since: vNEW`) whose `* langs:` still lacks `go`.
#   2. Test files each sibling roll added/changed — confirm you ported them.
#   3. New Go API symbols vs the sibling roll's added symbols (rough name diff).
#
# Requires: gh (authenticated), jq, git. Read-only.
set -euo pipefail

NEW="${1:?usage: verify-parity.sh <NEW_VERSION>  e.g. 1.61.1}"
NEW="${NEW#v}"
MINOR="$(echo "$NEW" | grep -oE '^[0-9]+\.[0-9]+')"
SINCE="v$MINOR"
ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
cd "$ROOT"

hr() { printf '\n========== %s ==========\n' "$1"; }

# Resolve the roll PR number for a sibling repo: newest merged roll PR whose
# title contains the minor (prefers a non-alpha/beta title, else the newest).
find_roll_pr() {
  local repo="$1"
  gh pr list --repo "$repo" --state all --search "in:title roll" --limit 80 \
    --json number,title,mergedAt \
    --jq "[.[] | select(.title | contains(\"$MINOR\"))]
          | (map(select(.title | (contains(\"alpha\") or contains(\"beta\")) | not)) + .)
          | .[0] | \"\(.number)\t\(.title)\"" 2>/dev/null || true
}

hr "1. New upstream APIs ($SINCE) still missing 'go' in langs"
echo "Each block below is new this version but NOT exposed to Go. Decide per item:"
echo "intentionally omitted (a sibling also skips it) OR a miss to fix in the patch."
MISSING=0
for f in playwright/docs/src/api/class-*.md playwright/docs/src/api/params.md; do
  [ -f "$f" ] || continue
  # Walk blocks: a header line (## / ###) starts a block; within it look for
  # `* since: vMINOR` and a `* langs:` line lacking 'go'. awk over the file.
  awk -v since="$SINCE" -v file="$(basename "$f")" '
    /^#+ (async )?(method|property|event|param|option):/ { hdr=$0; sincehit=0; langs="" }
    /^\* since:/ { if ($0 ~ since) sincehit=1 }
    /^\* langs:/ { langs=$0 }
    /^$/ {
      if (sincehit && langs != "" && langs !~ /[, ]go([, ]|$)/ && langs !~ /langs: go/) {
        printf "  %s\n    %s\n    %s\n", file, hdr, langs
      }
      sincehit=0; langs=""
    }
  ' "$f"
done
echo "(blocks with NO '* langs:' line default to all-languages incl. go — not listed)"

hr "2. Sibling roll PRs + the test files they added"
for repo in microsoft/playwright-python microsoft/playwright-java microsoft/playwright-dotnet; do
  pr="$(find_roll_pr "$repo")"
  num="$(echo "$pr" | cut -f1)"
  title="$(echo "$pr" | cut -f2-)"
  printf '\n# %s  ->  PR #%s  %s\n' "$repo" "${num:-?}" "$title"
  if [ -n "$num" ]; then
    gh pr view "$num" --repo "$repo" --json files \
      --jq '.files[].path | select(test("(?i)test|spec"))' 2>/dev/null | sed 's/^/    /' || true
  else
    echo "    (no roll PR found — search manually, see references/finding-changes.md)"
  fi
done
echo ""
echo ">> Confirm every NEW test file above has a counterpart under ./tests/ in this repo."

hr "3. New Go API symbols this roll (for eyeball diff vs siblings above)"
echo "Interface methods / structs added to generated files vs committed HEAD:"
git diff HEAD -- generated-interfaces.go generated-structs.go 2>/dev/null \
  | grep -E '^\+' | grep -vE '^\+\+\+' \
  | grep -oE '^\+\t([A-Z][A-Za-z0-9]+\(|type [A-Z][A-Za-z0-9]+ )' \
  | sed -E 's/^\+\t/  /; s/\($//' | sort -u | head -80
echo ""
echo ">> Compare these names against the sibling PR diffs. A Go symbol no sibling"
echo "   has is often a naming mistake; a sibling symbol missing here may be a gap."

hr "Done"
echo "Every remaining difference from python/java/dotnet should be deliberate and explainable."
