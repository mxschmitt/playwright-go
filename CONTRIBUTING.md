# Contributing

## Code style
The Go code is linted with [golangci-lint](https://golangci-lint.run/) and formatted with [gofumpt](https://github.com/mvdan/gofumpt). Please configure your editor to run the tools while developing and make sure to run the tools before committing any code.

## Tests

### Test coverage

For every Pull Request on GitHub and on the main branch the coverage data will get sent over to [Coveralls](https://coveralls.io/github/playwright-community/playwright-go), this is helpful for finding functions that aren't covered by tests.

### Running tests

You can use the `BROWSER` environment variable to use a different browser than Chromium for the tests and use the `HEADLESS` environment variable which is useful for debugging.

```
BROWSER=chromium HEADLESS=1 go test -v --race ./...
```

### Roll

1. Find out to which upstream version you want to roll, and change the value of `playwrightCliVersion` in **run.go** to the new version.
2. Download the current version of the Playwright driver: `go run scripts/install-browsers/main.go`
3. Apply the patch: `bash scripts/apply-patch.sh` (this checks out the matching upstream tag in the `playwright` submodule and commits the patched result automatically).
4. Fix merge conflicts if any, otherwise skip this step. Once you are happy, commit the changes: `cd playwright && git commit -am "apply patch" && cd ..`
5. Regenerate the patch: `bash scripts/update-patch.sh`
6. Generate Go code: `go generate ./...`

To adapt to the new version of Playwright's protocol and feature updates, you may need to modify the patch. Refer to the following steps:

1. Apply patch `bash scripts/apply-patch.sh`
2. `cd playwright`
3. Revert the patch `git reset HEAD~1`
4. Modify the files under `docs/src/api`, etc. as needed. Available references:
    - Protocol `packages/protocol/src/protocol.yml`
    - [Playwright python](https://github.com/microsoft/playwright-python)
5. Commit the changes `git commit -am "apply patch"`
6. Regenerate a new patch `bash scripts/update-patch.sh`
7. Generate go code `go generate ./...`.
