//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"

	"github.com/playwright-community/playwright-go"
)

func main() {
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not launch playwright: %v", err)
	}
	browser, err := pw.Chromium.Launch()
	if err != nil {
		log.Fatalf("could not launch Chromium: %v", err)
	}
	page, err := browser.NewPage(playwright.BrowserNewPageOptions{
		RecordVideo: &playwright.RecordVideo{
			Dir: playwright.String("videos/"),
		},
	})
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}
	gotoPage := func(url string) {
		fmt.Printf("Visiting %s\n", url)
		if _, err = page.Goto(url); err != nil {
			log.Fatalf("could not goto: %v", err)
		}
		fmt.Printf("Visited %s\n", url)
	}
	gotoPage("https://playwright.dev")
	gotoPage("https://github.com")
	gotoPage("https://microsoft.com")
	if err := page.Close(); err != nil {
		log.Fatalf("failed to close page: %v", err)
	}
	path, err := page.Video().Path()
	if err != nil {
		log.Fatalf("failed to get video path: %v", err)
	}
	fmt.Printf("Saved to %s\n", path)
	if err = browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}
}
