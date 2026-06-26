//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/mxschmitt/playwright-go"
)

func main() {
	const readmePath = "README.md"
	readmeContent, err := os.ReadFile(readmePath)
	if err != nil {
		log.Fatalf("could not read readme: %v", err)
	}
	replaceValueInTemplate := func(name, value string) {
		re := regexp.MustCompile(fmt.Sprintf("<!-- GEN:%s -->([^<]+)<!-- GEN:stop -->", name))
		readmeContent = re.ReplaceAll(readmeContent, []byte(fmt.Sprintf("<!-- GEN:%s -->%s<!-- GEN:stop -->", name, value)))
	}

	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not run playwright: %v", err)
	}
	browserTypes := []playwright.BrowserType{pw.Chromium, pw.Firefox, pw.WebKit}
	for _, browserType := range browserTypes {
		browser, err := browserType.Launch()
		if err != nil {
			log.Fatalf("could not launch browser: %v", err)
		}
		replaceValueInTemplate(fmt.Sprintf("%s-version", browserType.Name()), browser.Version())

		var badgeFormat string
		switch browserType.Name() {
		case "chromium":
			badgeFormat = "[![Chromium version](https://img.shields.io/badge/chromium-%s-blue.svg?logo=google-chrome)](https://www.chromium.org/Home)"
		case "firefox":
			badgeFormat = "[![Firefox version](https://img.shields.io/badge/firefox-%s-blue.svg?logo=mozilla-firefox)](https://www.mozilla.org/en-US/firefox/new/)"
		case "webkit":
			badgeFormat = "[![WebKit version](https://img.shields.io/badge/webkit-%s-blue.svg?logo=safari)](https://webkit.org/)"
		}
		replaceValueInTemplate(browserType.Name()+"-version-badge", fmt.Sprintf(badgeFormat, browser.Version()))

		if err := browser.Close(); err != nil {
			log.Fatalf("could not close browser: %v", err)
		}
	}
	if err := os.WriteFile(readmePath, readmeContent, 0o644); err != nil {
		log.Fatalf("could not write readme: %v", err)
	}
	if err := pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}
}
