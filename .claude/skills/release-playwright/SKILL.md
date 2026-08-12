---
name: release-playwright
description: Cut a new GitHub release (git tag + release notes) for playwright-go. Computes the next version from the pinned driver version in run.go and the last published tag, builds release notes from merged PRs, and publishes the release. Use when asked to "release", "cut a release", "tag a release", "publish a new version", or right after a roll-playwright PR has merged to main.
---

# Release playwright-go

Publish a new tagged GitHub release of `mxschmitt/playwright-go`. This is usually the last step right after a [`roll-playwright`](../roll-playwright/SKILL.md) PR merges to `main`, but can also be run standalone to ship accumulated non-roll fixes.

## Versioning scheme

Releases are **not** semver-on-the-package; the tag encodes the upstream Playwright driver version: `v0.<MM><PP>.<G>`

- `MM` — upstream minor version, two digits (`1.62.x` → `62`).
- `PP` — upstream driver's own patch digit, two digits (`1.62.1` → `01`). This is **not** an independent counter — it is meant to literally equal the driver's patch number.
- `G` — a go-package-only patch counter. Starts at `0` for a new `MM`+`PP` combination; increments only when a release ships with **no driver version change** (e.g. a bugfix-only release), reusing the same `MM`+`PP`.

Examples from history: driver `1.51.1` → `v0.5101.0`; driver `1.52.0` → `v0.5200.0`, then a go-only fix on the same driver → `v0.5200.1`.

> **Known exception:** `v0.6100.0` was published for driver `1.61.1` but used `PP=00` instead of the expected `01` — a one-off inconsistency (coincided with the org migration to `mxschmitt`), not a scheme change. Don't treat it as precedent; if you find another tag that breaks the `PP == driver patch` rule, flag it to the user rather than silently following it.

## Step 1 — Compute the candidate version

```bash
bash .claude/skills/release-playwright/compute-version.sh
```

Prints the pinned driver version (`run.go`), the latest published release, the candidate next tag, and a sanity-check table of the last several releases' tags against the driver version each actually pinned. Read that table — if any entry (besides the known `v0.6100.0` exception above) breaks the `PP == driver patch` rule, the scheme has drifted further and you should ask the user how to proceed instead of guessing.

## Step 2 — Confirm the release point

1. `git fetch origin main && git status` — confirm your local `main` matches `origin/main` (fast-forward only, no divergence). If a `roll-playwright` PR just merged, that merge commit is normally the release point.
2. Confirm CI is green on `main` at that commit before tagging:
   ```bash
   RUN=$(gh run list --repo mxschmitt/playwright-go --branch main --limit 1 --json databaseId,headSha --jq '.[0]')
   echo "$RUN"   # confirm headSha matches `git rev-parse HEAD`, then:
   gh run view "$(jq -r .databaseId <<<"$RUN")" --repo mxschmitt/playwright-go --exit-status
   ```
   Do not release on a red or stale run.

## Step 3 — Write release notes

Two cases:

- **After a roll** (the common case): pull the "New APIs" / "Behavior changes" / "Fixes" sections straight from the merged roll PR body (`gh pr view <num> --repo mxschmitt/playwright-go --json body --jq .body`) and reshape into the established header format (see any past release for the shape, e.g. `gh release view v0.6201.0 --repo mxschmitt/playwright-go`). Pull browser versions from the README's `GEN:` markers:
  ```bash
  grep -oE '(chromium|firefox|webkit)-version -->[^<]+' README.md
  ```
  Include the `[!IMPORTANT]` reinstall callout with the new tag baked into the example command:
  ```
  go run github.com/mxschmitt/playwright-go/cmd/playwright@v<NEW_TAG> install --with-deps
  ```
- **Non-roll release** (accumulated fixes only): a short summary paragraph is enough; skip the driver-version/reinstall callout since the driver didn't change.

Write the custom header to a file. **Do not include a `## What's Changed` heading in it** — `--generate-notes` (Step 4) appends its own, and including one yourself produces a duplicate (a real mistake made the first time this skill's process was done by hand — the release had to be edited after publishing to fix it).

## Step 4 — Tag and publish

```bash
gh release create v<NEW_TAG> --repo mxschmitt/playwright-go --target main \
  --title "v<NEW_TAG>" \
  --notes-file <your-header-file> \
  --generate-notes \
  --latest
```

`--generate-notes` appends the auto-generated "What's Changed" PR list (and "New Contributors" if applicable) after your custom header, computed against the previous tag automatically. Omit `--latest` if this is a patch to an older line rather than the newest release.

## Step 5 — Verify

```bash
gh release view v<NEW_TAG> --repo mxschmitt/playwright-go
```

Confirm: no duplicate headings, the tag points at the intended commit (`git rev-parse v<NEW_TAG>` should equal the `main` HEAD from Step 2), it's marked `Latest` if intended, and the PR list in "What's Changed" matches what actually merged since the last tag.

## When to ask the user

- The version-number scheme doesn't cleanly resolve (Step 1's sanity check flags a second inconsistent tag, or the driver version in `run.go` doesn't obviously map to `MM.PP`).
- It's unclear whether a release should be cut now vs. batching further changes first.
- Whether to mark the release `--latest` (e.g. driver is a pre-release/beta version, or this is a patch to a non-latest line).
- CI on `main` isn't green at the intended release commit.

Otherwise this is a short, mechanical flow — work it autonomously once the version is confirmed.

## Files

- `compute-version.sh` — read-only helper for Step 1.
