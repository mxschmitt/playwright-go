package main

import (
	"log"

	"github.com/mxschmitt/playwright-golang"
)

func exitIfError(message string, err error) {
	if err != nil {
		log.Fatalf(message, err)
	}
}

func main() {
	pw, err := playwright.Run()
	exitIfError("could not launch playwright: %v", err)
	browser, err := pw.Chromium.Launch()
	exitIfError("could not launch Chromium: %v", err)
	context, err := browser.NewContext()
	exitIfError("could not create context: %v", err)
	page, err := context.NewPage()
	exitIfError("could not create page: %v", err)
	err = page.Goto("https://github.com/microsoft/playwright")
	exitIfError("could not goto: %v", err)
	_, err = page.PDF(playwright.PagePdfOptions{
		Path: playwright.String("playwright-example.pdf"),
	})
	exitIfError("could not create PDF: %v", err)
	err = browser.Close()
	exitIfError("could not close browser: %v", err)
	err = pw.Stop()
	exitIfError("could not stop Playwright: %v", err)
}
