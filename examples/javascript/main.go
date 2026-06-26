//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"

	"github.com/mxschmitt/playwright-go"
)

func main() {
	pw, err := playwright.Run()
	if err != nil {
		log.Fatal(err)
	}
	browser, err := pw.Chromium.Launch()
	if err != nil {
		log.Fatalf("could not launch browser: %v\n", err)
	}
	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v\n", err)
	}
	if _, err = page.Goto("https://en.wikipedia.org/wiki/JavaScript"); err != nil {
		log.Fatalf("could not goto: %v\n", err)
	}
	// mw.config.values is the JS object where Wikipedia stores wiki metadata
	handle, err := page.EvaluateHandle("mw.config.values", struct{}{})
	if err != nil {
		log.Fatalf("could not acquire JSHandle: %v\n", err)
	}
	// mw.config.values.wgPageName is the name of the current page
	pageName, err := handle.GetProperty("wgPageName")
	if err != nil {
		log.Fatalf("could not get Wikipedia page name: %v\n", err)
	}

	fmt.Printf("Lots of type casting, brought to you by %s\n", pageName)

	if err := browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v\n", err)
	}
	if err := pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v\n", err)
	}
}
