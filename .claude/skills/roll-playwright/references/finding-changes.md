# Finding the changes for a roll

Four sources tell you what changed. All commands below are verified against live data
(example: rolling `v1.60.0` → `v1.61.0`). `find-changes.sh` automates most of this; this
file is the manual fallback and the per-source detail.

The discovery flow across all three sibling clients is the same: **the PR _title_ carries the
target version, branch names don't** (they're inconsistent and sometimes disagree with the
title). GitHub full-text search tokenizes a dotted version oddly — a bare minor like `1.61`
matches nothing, but the full `1.61.0` matches. The robust move is to list all `roll` PRs and
filter locally on the version.

> **Read the `-beta`/`-alpha` roll, not just the `.0`.** Each minor typically gets 2–3 roll
> PRs (alpha → beta → stable). The stable `X.Y.0` PR is frequently a one-line driver-version
> bump; the **API port lands in the preceding beta/alpha PR**. Verified: python `1.60.0` final
> roll #3079 was +1/−1 (one line), while the beta roll #3069 was +2194/−226 across 36 files.

---

## 1. microsoft/playwright (upstream monorepo) — the diff

Upstream has no roll PRs; it ships `vX.Y.Z` tags + GitHub Releases. `docs/src/api/*.md` is the
source of truth every client generates from.

```bash
# Confirm the tag/release exists and read notes
gh release list --repo microsoft/playwright --limit 10
gh release view v1.61.0 --repo microsoft/playwright --json tagName,name,publishedAt,body

# Currently-pinned version (the OLD side of the diff)
grep playwrightCliVersion run.go

# Changed API docs + protocol between OLD and NEW
gh api repos/microsoft/playwright/compare/v1.60.0...v1.61.0 --jq '.files[].filename' \
  | grep -E '^docs/src/api/|^packages/protocol/spec/|release-notes'

# Pull the hunk for one changed file and find newly-added APIs
gh api 'repos/microsoft/playwright/compare/v1.60.0...v1.61.0' \
  --jq '.files[] | select(.filename=="docs/src/api/class-apiresponse.md") | .patch'
```

**Reading langs annotations.** Each method/option/property block carries:
- `* since: vX.Y` — version it appeared.
- `* langs: js, python, csharp` — which clients expose it (comma+space separated).

A new API gated to other languages (e.g. `* langs: js, python`) is **not** in Go until you add
`go`. Look in each hunk for added blocks containing `* since: v<NEW>` and inspect their
`* langs:` line — those are the candidates. Variants:
- sub-list form for per-language naming: `* langs:\n  - alias-python: and_`
- Go-only block: `* langs: go`

**Protocol** lives under `packages/protocol/spec/*.yml` (`api.yml`, `page.yml`, `network.yml`,
…). There is **no** `packages/protocol/src/protocol.yml` (the `CONTRIBUTING.md` reference is
stale). Protocol rarely changes between minors.

**zsh gotcha:** single-quote any `gh api` URL containing `?` (e.g. `'...?ref=v1.61.0'`).

---

## 2. microsoft/playwright-python

Closest in shape to Go — use it as the primary parity reference.

```bash
# Robust: list roll PRs, filter on the version locally
gh pr list --repo microsoft/playwright-python --state merged \
  --search "in:title roll" --limit 60 --json number,title,url \
  --jq '.[] | select(.title | test("1\\.61\\.")) | "#\(.number)\t\(.title)\t\(.url)"'

# Read it — the body enumerates ported APIs
gh pr view <num> --repo microsoft/playwright-python
gh pr view <num> --repo microsoft/playwright-python --json files --jq '.files[].path'
gh pr diff <num> --repo microsoft/playwright-python
```

- **Title:** `chore: roll to X.Y.Z` / `chore: roll driver to X.Y.Z` / `chore(roll): vX.Y.Z`;
  beta rolls carry `-beta-<epochms>`.
- **Extract:** PR body; new files under `playwright/_impl/` (new classes); diffs of `_impl/*.py`
  and the regenerated `async_api/_generated.py` / `sync_api/_generated.py` for exact method
  names, param names/types, defaults, optionality, and removed/deprecated params.

---

## 3. microsoft/playwright-java

Best for a per-upstream-PR breakdown — java roll bodies link each `microsoft/playwright` PR
ported and note which needed no client change / were skipped.

```bash
# Most robust: list all roll PRs, grep the minor locally (surfaces alpha/beta/stable)
gh pr list --repo microsoft/playwright-java --state merged --search "roll" --limit 100 | grep '1.61'

# Exact-match shortcut — must use the FULL version incl. patch
gh search prs --repo microsoft/playwright-java "1.61.0" --limit 10

gh pr view <num> --repo microsoft/playwright-java
gh pr diff <num> --repo microsoft/playwright-java
```

- **Title:** `chore: roll driver to X.Y.Z` / `chore: roll X.Y.Z`; occasionally `feat:` for big
  feature rolls. Version marker file: `scripts/DRIVER_VERSION`.
- **Extract:** PR body (per-PR port list); new public interfaces under
  `playwright/src/main/java/com/microsoft/playwright/` and option/enum types under `.../options/`
  map ~1:1 to Go bindings.

---

## 4. microsoft/playwright-dotnet

```bash
# Search by title + version token (drop the patch to the minor if patch-level is empty)
gh pr list --repo microsoft/playwright-dotnet --state all \
  --search "roll in:title 1.61.0" --json number,title,headRefName,state,mergedAt

# List all roll PRs and eyeball
gh pr list --repo microsoft/playwright-dotnet --state all --search "roll in:title" \
  --limit 50 --json number,title,state,mergedAt

# Confirm a candidate actually bumps the driver (some "roll"-matching PRs aren't rolls)
gh pr diff <num> --repo microsoft/playwright-dotnet | grep DriverVersion
```

- **Title:** `chore: roll driver to X.Y.Z` / `chore(roll): roll Playwright to vX.Y.Z`; pre-release
  rolls embed `-alpha-<ts>`/`-beta-<ts>`. **Match on title, not branch.**
- **Version marker:** `src/Common/Version.props` `<DriverVersion>` (changed in every roll).
- **Extract:** ported API surface in `src/Playwright/API/Generated/**` (new `I*.cs` members,
  `Options/*Options.cs`, `Types/`, `Enums/`); wire changes in
  `src/Playwright/Transport/Protocol/Generated/**`. PR body is the best changelog.

---

## How these map to the Go patch

- The upstream `v<OLD>...v<NEW>` docs diff tells you **which API/option blocks gained
  `* since: v<NEW>`** and whether their `* langs:` already reaches Go.
- The **sibling PRs tell you which of those are real client surface** (names, types, defaults,
  optionality, what was skipped). Use python/java for parity decisions.
- The **Go roll mirrors this by adding `go` to the `* langs:`** of each desired new API in
  `docs/src/api/*.md`, recorded as one-line edits in `patches/main.patch`. Then regenerate.

The three sibling repos also keep their own roll runbooks at `.claude/skills/playwright-roll/SKILL.md`
— worth cross-referencing if a roll gets hairy.
