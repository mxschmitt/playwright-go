# 🎭 [Playwright](https://playwright.dev) for <img src="https://user-images.githubusercontent.com/17984549/91302719-343a1d80-e7a7-11ea-8d6a-9448ef598420.png" alt="Go" height="35" />

[![PkgGoDev](https://pkg.go.dev/badge/github.com/mxschmitt/playwright-go)](https://pkg.go.dev/github.com/mxschmitt/playwright-go)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](http://opensource.org/licenses/MIT)
[![Go](https://github.com/mxschmitt/playwright-go/actions/workflows/build.yml/badge.svg)](https://github.com/mxschmitt/playwright-go/actions/workflows/build.yml)
[![Tests](https://img.shields.io/endpoint?url=https%3A%2F%2Fflakiness.io%2Fapi%2Fbadge%3Finput%3D%257B%2522badgeToken%2522%253A%2522badge-6g4pNCL3d8qZbDdJEqFhSI%2522%257D)](https://flakiness.io/playwright-community/playwright-go)
[![Coverage Status](https://coveralls.io/repos/github/mxschmitt/playwright-go/badge.svg?branch=main)](https://coveralls.io/github/mxschmitt/playwright-go?branch=main)
[![Join Discord](https://img.shields.io/badge/join-discord-informational)](https://aka.ms/playwright/discord)
<!-- GEN:chromium-version-badge -->[![Chromium version](https://img.shields.io/badge/chromium-151.0.7922.34-blue.svg?logo=google-chrome)](https://www.chromium.org/Home)<!-- GEN:stop --> <!-- GEN:firefox-version-badge -->[![Firefox version](https://img.shields.io/badge/firefox-153.0-blue.svg?logo=mozilla-firefox)](https://www.mozilla.org/en-US/firefox/new/)<!-- GEN:stop --> <!-- GEN:webkit-version-badge -->[![WebKit version](https://img.shields.io/badge/webkit-26.5-blue.svg?logo=safari)](https://webkit.org/)<!-- GEN:stop -->

#### [Website](https://playwright.dev) | [API reference](https://pkg.go.dev/github.com/mxschmitt/playwright-go) | [Example recipes](https://github.com/mxschmitt/playwright-go/tree/main/examples)

Playwright is a Go library to automate [Chromium](https://www.chromium.org/Home), [Firefox](https://www.mozilla.org/en-US/firefox/new/) and [WebKit](https://webkit.org/) with a single API. Playwright is built to enable cross-browser web automation that is **ever-green**, **capable**, **reliable** and **fast**.

|          | Linux | macOS | Windows |
|   :---   | :---: | :---: | :---:   |
| Chromium <!-- GEN:chromium-version -->151.0.7922.34<!-- GEN:stop --> | :white_check_mark: | :white_check_mark: | :white_check_mark: |
| WebKit <!-- GEN:webkit-version -->26.5<!-- GEN:stop --> | :white_check_mark: | :white_check_mark: | :white_check_mark: |
| Firefox <!-- GEN:firefox-version -->153.0<!-- GEN:stop --> | :white_check_mark: | :white_check_mark: | :white_check_mark: |

Headless and headed execution is supported for all the browsers on all platforms.

## Installation

```shell
go get -u github.com/mxschmitt/playwright-go
```

Install the Playwright driver and browsers (add `--with-deps` to also install the OS dependencies). **Note** that you should replace the version number `0.xxxx.x` with the version used in your current `go.mod`. Each minor version upgrade requires a specific Playwright driver version.

```shell
go run github.com/mxschmitt/playwright-go/cmd/playwright@v0.xxxx.x install --with-deps
# Or
go install github.com/mxschmitt/playwright-go/cmd/playwright@v0.xxxx.x
playwright install --with-deps
```

Alternatively, you can download the driver and browsers from your code. If your operating system lacks the browser dependencies you still need to install them manually, because installing system dependencies requires privileges.

```go
err := playwright.Install()
```

## Documentation

[https://playwright.dev/docs/intro](https://playwright.dev/docs/intro)

The guides, concepts and API semantics are shared across all Playwright languages — only the code samples on that site are written in JavaScript.

## API Reference

[https://pkg.go.dev/github.com/mxschmitt/playwright-go](https://pkg.go.dev/github.com/mxschmitt/playwright-go)

## Example

The following example crawls the current top voted items from [Hacker News](https://news.ycombinator.com).

```go
package main

import (
	"fmt"
	"log"

	"github.com/mxschmitt/playwright-go"
)

func main() {
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	browser, err := pw.Chromium.Launch()
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}
	if _, err = page.Goto("https://news.ycombinator.com"); err != nil {
		log.Fatalf("could not goto: %v", err)
	}
	entries, err := page.Locator(".athing").All()
	if err != nil {
		log.Fatalf("could not get entries: %v", err)
	}
	for i, entry := range entries {
		title, err := entry.Locator("td.title > span > a").TextContent()
		if err != nil {
			log.Fatalf("could not get text content: %v", err)
		}
		fmt.Printf("%d: %s\n", i+1, title)
	}
	if err = browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}
}
```

## Capabilities

* **Resilient locators** — find elements the way a user sees the page with `GetByRole`, `GetByLabel`, `GetByPlaceholder`, `GetByText` and `GetByTestId` instead of brittle CSS paths.
* **Auto-wait** — actions such as `Click` and `Fill` wait for the element to be actionable, and web-first assertions created via `playwright.NewPlaywrightAssertions()` retry until the condition is met. No arbitrary sleeps.
* **Full isolation** — every `BrowserContext` is the equivalent of a brand new browser profile at near-zero overhead. Save the authentication state once with `context.StorageState()` and reuse it everywhere.
* **Trace Viewer** — record a trace via `context.Tracing()` and inspect DOM snapshots, network traffic and console logs afterwards with `playwright show-trace`.
* **Network interception** — stub and mock requests with `page.Route()`, or monitor all traffic of a page.
* **Emulation** — mobile devices, geolocation, permissions, color scheme, locale and timezone.
* **Beyond the DOM** — scenarios that span multiple pages, domains and iframes, shadow-piercing selectors, native mouse and keyboard input, file uploads and downloads.

## Docker

Refer to the [Dockerfile.example](./Dockerfile.example) to build your own Docker image.

## More examples

* Refer to [helper_test.go](./tests/helper_test.go) for End-To-End testing
* [Downloading files](./examples/download/main.go)
* [End-To-End testing a website](./examples/end-to-end-testing/main.go)
* [Executing JavaScript in the browser](./examples/javascript/main.go)
* [Emulate mobile and geolocation](./examples/mobile-and-geolocation/main.go)
* [Monitor network activity](./examples/network-monitoring/main.go)
* [Parallel scraping using a WaitGroup](./examples/parallel-scraping/main.go)
* [Rendering a PDF of a website](./examples/pdf/main.go)
* [Scraping HackerNews](./examples/scraping/main.go)
* [Take a screenshot](./examples/screenshot/main.go)
* [Using a locally installed Chrome](./examples/use-local-chrome/main.go)
* [Record a video](./examples/video/main.go)

## How does it work?

Playwright is a Node.js library which uses:

* Chrome DevTools Protocol to communicate with Chromium
* Patched Firefox to communicate with Firefox
* Patched WebKit to communicate with WebKit

These patches are based on the original sources of the browsers and don't modify the browser behaviour, so the browsers are basically the same (see [here](https://github.com/microsoft/playwright/tree/main/browser_patches)) as you see them in the wild. The support for different programming languages is based on exposing a RPC server in the Node.js land which can be used to allow other languages to use Playwright without implementing all the custom logic.

The bridge between Node.js and the other languages is basically a Node.js runtime combined with Playwright which gets shipped for each of these languages (around 50MB) and then communicates over stdio to send the relevant commands. This will also download the pre-compiled browsers.

## Other languages

More comfortable in another programming language? [Playwright](https://playwright.dev) is also available in

* [Node.js (JavaScript / TypeScript)](https://playwright.dev/docs/intro),
* [Python](https://playwright.dev/python/docs/intro),
* [.NET](https://playwright.dev/dotnet/docs/intro),
* [Java](https://playwright.dev/java/docs/intro).
