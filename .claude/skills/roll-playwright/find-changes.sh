#!/bin/bash
# find-changes.sh — Discover everything that changed for a Playwright version roll.
#
# Usage:   bash .claude/skills/roll-playwright/find-changes.sh <NEW_VERSION> [OLD_VERSION]
# Example: bash .claude/skills/roll-playwright/find-changes.sh 1.61.0
#
# Prints, for the target version:
#   1. The upstream release notes link + whether the tag exists.
#   2. The exact upstream docs/src/api/*.md files that changed between OLD and NEW.
#   3. Every NEW API block (`* since: vNEW`) and its current `* langs:` line — these
#      are the candidates that may need `go` added in patches/main.patch.
#   4. The sibling roll PRs (python / java / dotnet) to read for parity.
#
# Requires: gh (authenticated), jq. Read-only — makes no changes to the repo.
set -euo pipefail

NEW="${1:?usage: find-changes.sh <NEW_VERSION> [OLD_VERSION]  e.g. 1.61.0}"
NEW="${NEW#v}" # tolerate a leading v

# Derive OLD from run.go (the currently-pinned version) unless explicitly given.
if [ -n "${2:-}" ]; then
  OLD="${2#v}"
else
  OLD="$(grep -oE 'playwrightCliVersion = "[0-9.]+"' run.go | grep -oE '[0-9.]+')"
fi
MINOR="$(echo "$NEW" | grep -oE '^[0-9]+\.[0-9]+')"

hr() { printf '\n=== %s ===\n' "$1"; }

hr "Rolling: v$OLD  ->  v$NEW   (minor $MINOR)"

hr "1. Upstream release"
if gh api "repos/microsoft/playwright/git/refs/tags/v$NEW" >/dev/null 2>&1; then
  echo "tag v$NEW exists."
  gh release view "v$NEW" --repo microsoft/playwright --json tagName,name,publishedAt --jq '"  \(.tagName)  published \(.publishedAt)"' 2>/dev/null || true
  echo "  release notes: https://github.com/microsoft/playwright/releases/tag/v$NEW"
else
  echo "WARNING: tag v$NEW does not exist upstream yet. Available recent tags:"
  gh release list --repo microsoft/playwright --limit 8 2>/dev/null | sed 's/^/    /'
fi

hr "2. Changed upstream API docs + protocol (v$OLD...v$NEW)"
CMP="repos/microsoft/playwright/compare/v$OLD...v$NEW"
CHANGED="$(gh api "$CMP" --jq '.files[].filename' 2>/dev/null | grep -E '^docs/src/api/|^packages/protocol/spec/|release-notes' || true)"
if [ -z "$CHANGED" ]; then
  echo "(no docs/api/protocol files changed, or compare failed — check both tags exist)"
else
  echo "$CHANGED" | sed 's/^/  /'
fi

hr "3. NEW API blocks gated to other languages (candidates for adding 'go')"
echo "For each changed class-*.md, lines added in this version with their langs line."
echo "Look for '* langs:' WITHOUT 'go' — those are not yet exposed to the Go client."
for f in $(echo "$CHANGED" | grep '^docs/src/api/class-' || true); do
  PATCH="$(gh api "$CMP" --jq ".files[] | select(.filename==\"$f\") | .patch" 2>/dev/null || true)"
  # Surface added method/property/option headers + their since/langs context.
  HITS="$(echo "$PATCH" | grep -E '^\+' | grep -iE 'method:|property:|## |### |since: v'"$MINOR"'|langs:' || true)"
  if [ -n "$HITS" ]; then
    printf '\n  --- %s ---\n' "$f"
    echo "$HITS" | sed 's/^/    /'
  fi
done

hr "4. Sibling roll PRs to read for parity"
# Match the minor as a literal substring (contains), NOT test() — test() is a regex
# where `.` is a wildcard, which produces false positives like "1.3.0-next.106133".
echo "# python (heaviest changes are usually in the -beta roll, not the final .0):"
gh pr list --repo microsoft/playwright-python --state merged --search "in:title roll" --limit 60 \
  --json number,title,url --jq ".[] | select(.title | contains(\"$MINOR\")) | \"  #\(.number)  \(.title)\"" 2>/dev/null || true
echo "# java:"
gh pr list --repo microsoft/playwright-java --state all --search "in:title roll" --limit 60 \
  --json number,title,url --jq ".[] | select(.title | contains(\"$MINOR\")) | \"  #\(.number)  \(.title)\"" 2>/dev/null || true
echo "# dotnet:"
gh pr list --repo microsoft/playwright-dotnet --state all --search "roll in:title" --limit 60 \
  --json number,title --jq ".[] | select(.title | contains(\"$MINOR\")) | \"  #\(.number)  \(.title)\"" 2>/dev/null || true

hr "Next"
cat <<EOF
  Read the roll PRs above (prefer the -beta/-alpha one — it carries the real API port;
  the bare X.Y.0 PR is often just a version bump):
    gh pr view <num> --repo microsoft/playwright-python   # body lists every ported API
    gh pr diff <num> --repo microsoft/playwright-python
  Then follow SKILL.md to bump run.go, edit patches/main.patch langs, and regenerate.
EOF
