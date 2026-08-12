#!/usr/bin/env bash
# Read-only helper for release-playwright: prints the pinned driver version,
# the latest published release, and the candidate next tag.
#
# Usage: bash .claude/skills/release-playwright/compute-version.sh
set -euo pipefail

REPO="mxschmitt/playwright-go"

driver_version=$(grep -oE 'playwrightCliVersion[[:space:]]*=[[:space:]]*"[0-9.]+"' run.go | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' || true)
if [[ -z "$driver_version" ]]; then
  echo "Could not find playwrightCliVersion in run.go" >&2
  exit 1
fi

minor=$(cut -d. -f2 <<<"$driver_version")
patch=$(cut -d. -f3 <<<"$driver_version")
mm=$(printf "%02d" "$minor")
pp=$(printf "%02d" "$patch")
base="v0.${mm}${pp}"

latest_tag=$(gh release list --repo "$REPO" --limit 1 --json tagName --jq '.[0].tagName')

echo "Pinned driver version (run.go): $driver_version"
echo "Latest published release:       ${latest_tag:-<none found>}"
echo

if [[ "$latest_tag" == "${base}."* ]]; then
  last_g=$(cut -d. -f3 <<<"$latest_tag")
  next_g=$((last_g + 1))
  echo "Same driver version as the latest release -> this looks like a go-only patch release (no driver bump)."
  echo "Candidate tag: ${base}.${next_g}"
else
  echo "Driver version differs from the latest release's -> this looks like a new roll release."
  echo "Candidate tag (mm=minor, pp=driver patch, following the v0.<mm><pp>.0 convention): ${base}.0"
fi

echo
echo "--- Sanity check against history (does the mm+pp digit pattern actually track the driver patch?) ---"
gh release list --repo "$REPO" --limit 8 --json tagName --jq '.[].tagName' | while read -r tag; do
  dv=$(git show "${tag}:run.go" 2>/dev/null | grep -oE 'playwrightCliVersion[[:space:]]*=[[:space:]]*"[0-9.]+"' | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' || true)
  echo "$tag -> driver $dv"
done
echo
echo "NOTE: v0.6100.0 is a known one-off exception (pinned driver 1.61.1 but used '00' not '01')."
echo "Do not treat it as the new pattern without asking the user first — see SKILL.md 'When to ask the user'."
