---
name: roll-playwright
description: Roll playwright-go to a new upstream Playwright version. Bumps the pinned CLI version, advances the vendored submodule, finds the API changes from the sibling clients (python/java/dotnet roll PRs) and the upstream docs diff, updates patches/main.patch, regenerates the Go API, and verifies the build. Use when asked to "roll to", "bump", "update to", or "upgrade Playwright" to a version like v1.61.0.
---

# Roll playwright-go to a new Playwright version

Roll the client from the currently-pinned version to a target version (e.g. `1.61.0`). Work autonomously; only stop for the decision points called out in **When to ask the user**.

## What a roll actually is

playwright-go vendors `microsoft/playwright` as a git submodule (`playwright/`) pinned to a tag. The public Go API (`generated-*.go`) is generated from the submodule's `docs/src/api/*.md` by `playwright/utils/doclint/generateGoApi.js`. Those `.md` files gate each API by language with a `* langs:` line; **the Go client only sees an API if `go` is in that line.** `patches/main.patch` is a diff applied onto the upstream submodule that (a) adds the Go generator and (b) adds `go` to the `* langs:` of every API the Go client exposes.

So a roll = bump the version, re-apply the patch onto the new submodule, add `go` to the langs of any **newly-added** APIs we want, regenerate, build, test.

## Inputs

- Target version, e.g. `1.61.0` (the `v` is optional). If not given, default to the latest stable from `gh release list --repo microsoft/playwright --limit 5` and confirm with the user.
- Current version is read from `run.go` (`const playwrightCliVersion`).

## Prerequisites (check, install if missing)

`gh` (authenticated), `node`, `go`, `gofumpt` (`go install mvdan.cc/gofumpt@latest`), `jq`. The submodule must be initialized: `git submodule update --init`.

## Step 1 — Discover what changed

Run the helper, which prints the upstream docs diff, every new `* since: vNEW` API with its `langs:` line, and the sibling roll PRs:

```bash
bash .claude/skills/roll-playwright/find-changes.sh <NEW_VERSION>
```

**Required: locate the actual roll PR in all three sibling clients** (python, java, dotnet) before porting — you need all three as references, not just one. They are the source of truth for which new APIs are real client surface, their exact method/param names, types, defaults, optionality, AND which upstream PRs were skipped. The helper prints candidates; for each repo open the right PR and confirm it's the one that ports the API surface (not a no-op bump):

```bash
gh pr view <num> --repo microsoft/playwright-python   # body enumerates ported APIs
gh pr view <num> --repo microsoft/playwright-java      # body links each upstream PR ported + what was skipped
gh pr view <num> --repo microsoft/playwright-dotnet
```

If a repo's helper line is missing or looks like a no-op, dig with the queries in `references/finding-changes.md` until you have a real roll PR for **each** of the three. Use python as the primary shape reference (closest to Go), java for the per-upstream-PR breakdown, dotnet for option/enum shapes.

> **Critical:** the heavy API port usually lands in the **`-beta` / `-alpha`** roll PR, not the final `X.Y.0` PR (which is often a one-line version bump — e.g. python v1.60.0 final roll was +1/−1, the beta roll +2194/−226). Read the beta/feature roll for the real changes.

Build a checklist of new/changed APIs to expose in Go, AND a checklist of the **new tests** each sibling added (you'll port these in Step 6b). The Go client deliberately omits some upstream APIs (e.g. some Electron/Android/JS-internal surface) — match python's choices unless there's reason not to.

> **Watch for return-type naming:** new return-object types carry a `* alias:` in the docs (e.g. `* alias: VirtualCredential`, `* alias: WebStorageItem`). The Go generator does NOT honor it by default and falls back to a generic method-derived name (`Create`, `Get`, `Item`). If you see a generic/duplicate generated struct name, add a guarded special-case in `generateGoApi.js` `generateNameDefault` (scoped by `parent.name`) mapping it to the documented alias — see `references/patch-editing.md`.
> **Watch for pure getters:** a new no-arg getter (e.g. `Page.LocalStorage()`, `BrowserContext.Credentials()`) generates with a trailing `error` return unless its method name is in `methodNoErrArray` in `generateGoApi.js`. Add it there so it matches sibling getters like `Clock()`/`Mouse()`.

## Step 2 — Bump the version

1. Edit `run.go`: set `const playwrightCliVersion = "<NEW_VERSION>"`.
2. Download the new driver (also advances browser versions for the README step):
   ```bash
   go run scripts/install-browsers/main.go
   ```

## Step 3 — Apply the patch onto the new submodule

```bash
bash scripts/apply-patch.sh
```

This checks out the new tag in `playwright/`, creates branch `playwright-build`, and runs `git apply --3way patches/*`. Two outcomes:

- **Clean apply** → continue to Step 4.
- **Merge conflicts** (`.rej` files, or `<<<<<<<` markers in `playwright/docs/src/api/*.md`) → upstream edited a line the patch also touched. Resolve them: keep the upstream change AND re-add `go` to the `* langs:` line. Then `cd playwright && git add -A && git commit -am "apply patch" && cd ..`.

## Step 4 — Add `go` to new APIs (the real work)

For each new API from your Step 1 checklist that you want in Go, edit the matching block in `playwright/docs/src/api/class-*.md` (and `params.md`) so its `* langs:` line includes `go`. The patch in Step 5 is regenerated from these edits.

```bash
cd playwright
git reset --soft HEAD~1   # un-commit the applied patch so edits fold into one diff (keeps working tree)
# ...edit docs/src/api/*.md: add `go` to the `* langs:` lines of the new APIs...
git add -A && git commit -m "apply patch"
cd ..
```

Rules for editing langs (see `references/patch-editing.md` for details and examples):
- `* langs: js, python, csharp` → `* langs: js, python, csharp, go` (append, comma+space).
- A Go-only addition uses `* langs: go`.
- Keep edits surgical — only touch lines for APIs you're intentionally exposing.
- Mirror the python client's method/param naming and optionality; the Go generator turns these docs into `generated-*.go`.

## Step 5 — Regenerate the patch and the Go API

```bash
bash scripts/update-patch.sh      # regenerates patches/main.patch from the playwright-build branch
go generate ./...                 # runs generate-api.sh: re-applies patch, runs the JS generator, gofumpt
```

`go generate` rewrites `generated-enums.go`, `generated-interfaces.go`, `generated-structs.go`. Review the diff — it should reflect exactly the APIs you added.

> **Gotcha:** `generate-api.sh` ends with `git submodule update`, which resets `playwright/` back to the parent repo's *currently pinned* (old) commit. After generation you MUST bump the gitlink yourself: `(cd playwright && git checkout v<NEW_VERSION>)`, then confirm with `git submodule status playwright`. Otherwise the committed submodule pointer still references the old version. Do this as part of Step 9.

## Step 6 — Hand-written wiring

The generator produces interfaces/structs, but most new APIs need a hand-written implementation in the corresponding `*.go` file (e.g. a new `Page` method goes in `page.go`, often with a channel send). Use the sibling clients' impl (python `_impl/*.py`, java/dotnet `Core`/`impl`) as the behavior reference. Look at a previous roll commit for the pattern:

```bash
git show $(git log --oneline --grep="[Rr]oll to" -1 --format=%H) --stat   # files a past roll touched
```

New thin wrapper classes (not channel owners) follow `clock.go`: a struct holding the parent `*pageImpl`/`*browserContextImpl`, constructed in the parent's `newPage`/`newBrowserContext`, exposed via a getter, methods call `parent.channel.Send(...)`. Wire the getter field into the parent struct too.

Watch for **protocol behavior changes** (not just new APIs) — these don't show as build errors but break at runtime. Check the sibling roll PR bodies and the upstream `validator.ts` diff (`gh api .../compare/vOLD...vNEW`) for renamed channel methods, split methods, or changed result shapes. Example from v1.61: assertion `expect` stopped returning `{matches}` and now throws a protocol error carrying `errorDetails` — every assertion path had to be updated. **Grep for duplicated logic** (e.g. there were TWO expect call-sites: `locatorImpl.expect` and `pageAssertionsImpl.expectOnFrame`) so you fix all of them.

### Step 6b — Port the sibling tests

This is required, not optional. For every new API, port the tests the siblings added (from your Step 1 test checklist) into `tests/`, following Go conventions (`BeforeEach(t)`, `require`, `server.EMPTY_PAGE`). Mirror python's test cases 1:1 where possible:

```bash
# See exactly which test files each sibling roll added/changed:
gh pr view <num> --repo microsoft/playwright-python --json files --jq '.files[].path' | grep -i test
gh pr view <num> --repo microsoft/playwright-java   --json files --jq '.files[].path' | grep -i [Tt]est
gh pr view <num> --repo microsoft/playwright-dotnet --json files --jq '.files[].path' | grep -i [Tt]est
# Then read the diff of a specific test file to port its cases:
gh pr diff <num> --repo microsoft/playwright-python  # find the new test_*.py blocks
```

These tests are also your verification that the impl works — run them in Step 7.

## Step 7 — Build, lint, test

```bash
gofumpt -l -w .
go build ./...
go vet ./...
go test -race ./...                 # or target a package; CI runs full -race on 3 OSes x 3 browsers
```

The `verify_type_generation` CI job runs `go generate` and fails if it produces a diff — so committed `generated-*.go` must exactly match a fresh generation. Run `go generate ./...` once more and confirm `git diff --ignore-submodules` is clean for the generated files.

## Step 8 — Update README versions

Already handled inside `scripts/generate-api.sh` (it runs `update-readme-versions/main.go`), but if you ran steps manually, run `go run scripts/update-readme-versions/main.go` to refresh the browser-version badges.

## Step 9 — Verify parity against the sibling clients

Before committing, run the parity verifier. It cross-checks this roll's diff against what python/java/dotnet did, so you catch anything missed: a new `* since: vNEW` API the siblings expose but you didn't add `go` to, an API you exposed that no sibling did, or a sibling test you haven't ported.

```bash
bash .claude/skills/roll-playwright/verify-parity.sh <NEW_VERSION>
```

It reports three buckets:
1. **New upstream APIs not in Go** — each `* since: vNEW` block whose `* langs:` still lacks `go`. For each, decide: intentionally omitted (matches a sibling skipping it) or a miss to fix in `patches/main.patch`.
2. **Sibling test files added this roll** — so you can confirm you ported each one into `tests/` (Step 6b).
3. **Go API symbols added vs sibling API surface** — a rough name-level diff to spot a method you exposed that no sibling did (often a naming mistake) or vice-versa.

Treat each flagged item as needing an explicit decision: fix it, or note why Go intentionally differs. The goal is that any remaining difference from python/java/dotnet is deliberate and explainable.

## Step 10 — Commit & PR

Branch name should start with `roll/` (CI triggers on `roll/*`). Match the existing commit style:

```bash
# Bump the submodule gitlink to the new tag (generate-api.sh reset it — see Step 5 gotcha):
(cd playwright && git checkout v<NEW_VERSION>)
git submodule status playwright          # confirm it shows v<NEW_VERSION>
git diff --submodule=log -- playwright   # confirm the pointer moves OLD -> NEW

git checkout -b roll/v<NEW_VERSION>
git add -A && git commit -m "chore: roll to Playwright v<NEW_VERSION>"
```

Open a PR only if the user asks. Summarize new APIs added in the PR body.

## Existing tests may need updating for behavior changes

A roll can change the behavior of an *existing* API, breaking a test that asserts the old behavior — this is not a port bug. When a pre-existing test fails:
1. Confirm you didn't break it (does it fail because of a signature/impl change you made?).
2. If not, check whether upstream changed that behavior: read the relevant sibling roll-PR test diffs and the upstream source diff (`gh api .../compare/vOLD...vNEW --jq '.files[].filename'` then inspect the changed server file). Example from v1.61: `socksClientCertificatesInterceptor.ts` was rewritten so a client cert is only presented to its *matching* origin — so the client-cert test's origin-mismatch navigation went from "503, no cert" to "TLS handshake aborted (`net::ERR_BAD_SSL_CLIENT_AUTH_CERT`)". Update the assertion to the new behavior.
3. New error/abort behaviors can introduce async navigation races (an aborted nav triggers a `chrome-error://` navigation that interrupts the next `Goto` with "is interrupted by another navigation"). Retry the follow-up navigation in a small loop (see `gotoSettled` in `browser_context_client_certificates_test.go`) rather than asserting once.

## When to ask the user

- Target version is ambiguous or unreleased upstream.
- A new upstream API is non-trivial (new class, breaking signature change) and you're unsure whether Go should expose it or how to name it.
- A pre-existing test fails and after investigation you're genuinely unsure whether it's an intended upstream behavior change (update the test) or a port bug (fix the code). If the evidence is conclusive (e.g. an upstream source rewrite + the Go client doesn't implement that logic), proceed and flag it; otherwise ask.
- Before pushing or opening a PR.

## Files & references

- `find-changes.sh` — discovery helper (read-only): upstream docs diff, new `* since:` APIs, sibling roll PRs.
- `verify-parity.sh` — parity verifier (read-only, Step 9): new APIs missing `go`, sibling test files to port, Go-symbol diff.
- `references/finding-changes.md` — exact `gh` queries + roll-PR conventions for python/java/dotnet + upstream compare.
- `references/patch-editing.md` — anatomy of `patches/main.patch`, `* langs:` gating, editing the generator (`methodNoErrArray`, return-type aliases), conflict resolution.

## Gotchas

- `CONTRIBUTING.md` references `packages/protocol/src/protocol.yml` — **stale**. The protocol is now `packages/protocol/spec/*.yml`, and it rarely changes between minors.
- GitHub search tokenizes versions oddly: searching a bare minor like `1.61` matches nothing; use the full `1.61.0`, or list all roll PRs and filter locally (what the helper does). When filtering with `jq`, use `contains("1.61")` not `test("1.61")` — `test` is a regex where `.` matches any char (false positives like `1.3.0-next.106133`).
- Single-quote any `gh api` URL containing `?` (zsh globs on `?`).
- `apply-patch.sh` deletes and recreates the `playwright-build` branch each run — safe to re-run. (It now also always checks out the pinned tag first — a fix for the first-roll-on-fresh-submodule case where it used to patch the old version.)
- `generate-api.sh` ends with `git submodule update`, resetting `playwright/` to the OLD pinned commit. Re-checkout the new tag before committing (Step 10).
- The `.0` sibling roll PR is often just a version bump; the API port is in the preceding `-beta`/`-alpha` PR.
- New API blocks with **no `* langs:` line at all** default to all languages (incl. go) and are generated without any patch edit — the patch edits are only for blocks that explicitly exclude go.
- Run `go test -race` before finishing: a new test that exercises a previously-untested API (e.g. an event callback) can surface a **latent data race in existing impl code**, not just your test. Fix the impl (guard shared fields with a mutex), don't just adjust the test.
