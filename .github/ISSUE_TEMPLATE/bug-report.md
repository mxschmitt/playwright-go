---
name: "Bug report \U0001FAB2"
about: Create a bug report to help us improve
title: "[Bug]: "
labels: p2-bug
assignees: ''

---

<!--
Before posting a bug report, please make sure it is a software defect related to "playwright-go"(or playwright).

The issue tracker is not a help forum. Consider asking for help on for example:
- [Github Discussions](/mxschmitt/playwright-go/discussions)
- upstream [Playwright Discord channel](https://aka.ms/playwright/discord)
-->
**Environments**
<!--pls always use latest since playwright-go still `v0.x`-->
- playwright-go Version: [e.g. v0.4201.1]   
- Browser: [e.g. firefox]
- OS and version: [e.g. macOS / Windows 11/ Ubuntu 22.04]

**Bug description**
<!-- A clear and concise description of what the bug is. Describe the expected bahavior
  and how it is currently different.-->

**To Reproduce**
Please provide a mini reproduction rather than just a description. For example:

```go
package main

import "github.com/mxschmitt/playwright-go"

func main() {
	// ignore unnecessary error handling code
	pw, _ := playwright.Run()
	browser, _ := pw.Chromium.Launch()
	context, _ := browser.NewContext()
	page, _ := context.NewPage()

	_, _ = page.Goto("https://playwright.dev")

	_, err := page.PDF(playwright.PagePdfOptions{
		Path: playwright.String("playwright-example.pdf"),
	})
	// should no error
	if err != nil {
		panic(err)
	}
}

```

**Additional context**
Add any other context about the problem here.
