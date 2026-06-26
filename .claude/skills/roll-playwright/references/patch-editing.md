# Editing patches/main.patch

`patches/main.patch` is a git diff applied onto the vendored `playwright/` submodule. You never
hand-edit the `.patch` file directly — you edit the files **inside `playwright/`**, then
regenerate the patch with `scripts/update-patch.sh`. This file explains what the patch contains
and how to add/change entries.

## What the patch contains

24 file diffs, two kinds:

1. **`utils/doclint/generateGoApi.js`** — a new ~880-line file: the Go API generator. Vendored
   here because upstream doesn't ship a Go generator. You rarely touch this; only if the upstream
   doclint framework (`documentation.js`, `api_parser.js`, `Type` shapes) changes in a way that
   breaks generation, or to map a new Go type. The CI `verify_type_generation` job will surface
   such breakage.

2. **`docs/src/api/*.md` + `params.md`** — ~130 small edits, almost all of the form "add `go` to
   a `* langs:` line" or "add a Go-only block". This is the gate that decides what the Go client
   exposes.

## How langs gating works

The generator (`generateGoApi.js`) calls `documentation.filterForLanguage('go')`. Any API block
whose `* langs:` line does **not** include `go` is dropped before code generation. So:

- An API present in `generated-*.go` ⇒ its docs block has `go` in `* langs:` (or no `* langs:`
  line at all, meaning "all languages").
- A new upstream API gated to other languages (e.g. `* langs: js, python`) is invisible to Go
  until you add `go`.

## The three edit patterns (real examples from the current patch)

### 1. Append `go` to an existing langs line

The most common edit. Append `, go` (comma + space) to the end:

```diff
-* langs: js, python
+* langs: js, python, go
```
```diff
-* langs: java, js, csharp
+* langs: java, js, csharp, go
```

Applies to method blocks (`## method:` / `## async method:`), option blocks (`### option:`),
param blocks (`### param:`), and property blocks (`## property:`).

### 2. Add a langs line to a block that had none

If an option block under a method had no `* langs:` line but you need it Go-visible distinctly,
add one (often when the surrounding method's langs was narrowed). Example from the patch — a new
struct option exposed to all langs incl. go:

```diff
 * since: v1.51
+* langs: js, python, java, csharp, go
 - `indexedDB` ?<boolean>
```

### 3. Add a Go-only block

When Go needs a differently-shaped signature than the shared one (e.g. an option struct where
other languages take a positional param). The pattern: leave the original block restricted to the
other languages and add a parallel `* langs: go` block. Real example (`Browser.startTracing.page`):

```diff
 ### param: Browser.startTracing.page
 * since: v1.11
+* langs: js, java, python, csharp
 - `page` ?<[Page]>

 Optional, if specified, tracing includes screenshots of the given page.

+### option: Browser.startTracing.page
+* since: v1.11
+* langs: go
+- `page` <[Page]>
+
+Optional, if specified, tracing includes screenshots of the given page.
```

Here the upstream `param` is narrowed away from `go`, and a `go`-only `option` (struct field) is
added — because the Go client passes options as a struct, not a positional param.

## Workflow to change the patch

From `CONTRIBUTING.md`, adapted:

```bash
# 1. Apply the current patch onto the (newly-bumped) submodule
bash scripts/apply-patch.sh

# 2. Un-commit it so your edits fold into the same diff (working tree is preserved)
cd playwright
git reset --soft HEAD~1     # CONTRIBUTING says `git reset HEAD~1`; --soft keeps it staged. Either works.

# 3. Edit docs/src/api/*.md — add `go` to the langs of the new APIs (patterns above)

# 4. Re-commit
git add -A && git commit -m "apply patch"
cd ..

# 5. Regenerate the patch file and the Go code
bash scripts/update-patch.sh    # git diff playwright-build^1..playwright-build > patches/main.patch
go generate ./...               # re-applies patch, runs the JS generator, gofumpt-formats
```

## Resolving conflicts from `git apply --3way`

When upstream edits a line the patch also touches (common: upstream added a language to a
`* langs:` line, or reflowed a doc paragraph), `apply-patch.sh` leaves a conflict.

1. Find them: `cd playwright && git status` (look for `.rej`), or
   `grep -rn '<<<<<<<' docs/src/api/`.
2. Resolve by **keeping the upstream content AND ensuring `go` is in the langs line.** Example: if
   upstream changed `* langs: js, python` → `* langs: js, python, java` and the patch wanted
   `* langs: js, python, go`, the merged result is `* langs: js, python, java, go`.
3. `git add -A && git commit -am "apply patch"`, then regenerate (Step 5 above).

## Verifying your edits

- `go generate ./...` then `git diff --ignore-submodules generated-interfaces.go` — the generated
  diff should contain exactly the APIs you added, nothing spurious.
- The CI `verify_type_generation` job runs `go generate` and fails on any diff, so committed
  generated files must match a fresh run byte-for-byte (after `gofumpt`).
- `go build ./...` then `go vet ./...` — generated interfaces compile, but new methods usually
  need a hand-written impl in the matching `*.go` (see SKILL.md Step 6).

## Editing the generator (`generateGoApi.js`)

Sometimes the langs edits aren't enough and you must tweak the vendored generator itself (it's
part of the patch, so edit it inside `playwright/`, then `update-patch.sh`). Two cases recur on
rolls that add new APIs:

### Pure getters returning a spurious `error`

By default the generator gives every method a trailing `error` return. Pure no-arg getters
(`Clock()`, `Mouse()`, `Keyboard()`, …) are exempted via the `methodNoErrArray` list near the top
of the file. A new getter like `Page.LocalStorage()` / `BrowserContext.Credentials()` will
otherwise generate as `LocalStorage() (WebStorage, error)` and break the hand-written impl. Add
the method name to `methodNoErrArray` (keep it alphabetical):

```js
'Context',
'Contexts',
'Credentials',      // <- added: pure getter, no error
'DefaultValue',
...
'Keyboard',
'LocalStorage',     // <- added
...
'ServiceWorkers',
'SessionStorage',   // <- added
```

### Generic/duplicate return-type names (honor `* alias:`)

Inline return-object types get a name derived from the method (`Create`, `Get`, `Item`) which is
generic and can collide. The docs annotate the intended name with `* alias:` (e.g.
`* alias: VirtualCredential`, `* alias: WebStorageItem`), which java/csharp honor but the Go
generator does not. Add a **scoped** special-case in `generateNameDefault` alongside the existing
ones (`SecurityDetail` → `ResponseSecurityDetailsResult`, etc.), guarded by `parent.name` so it
can't rename an unrelated type:

```js
// WebAuthn: Credentials.Create/Get return the same shape; use the documented alias.
if (parent.name === 'Credentials' && (attemptedName === 'Create' || attemptedName === 'Get'))
  attemptedName = 'VirtualCredential';
if (parent.name === 'WebStorage' && attemptedName === 'Item')
  attemptedName = 'WebStorageItem';
```

Do NOT make the generator honor `langAliases` globally — 40+ existing types carry aliases and a
blanket change would rename stable public types. Scope every fix.

After either edit: `cd playwright && git commit -am "apply patch"` (or amend the Applied-patches
commit), `update-patch.sh`, `go generate ./...`, then confirm the generated signatures.

## Tips

- Keep edits surgical: only flip langs for APIs you intend to expose. A stray `go` produces a
  generated method with no impl → build break.
- Match python's naming/optionality; the Go method/struct names derive from these docs.
- `enumTypes`, `classNameMap`, and `methodNoErrArray` near the top of `generateGoApi.js` map names
  and mark which methods don't return `error` (see the generator-editing section above).
