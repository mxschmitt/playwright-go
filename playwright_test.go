package playwright_test

import (
	"log"

	"github.com/mxschmitt/playwright-go"
)

func exitIfErrorf(message string, err error) {
	if err != nil {
		log.Fatalf(message, err)
	}
}

func ExampleRun() {
	pw, err := playwright.Run()
	exitIfErrorf("could not launch playwright: %v", err)
	browser, err := pw.Chromium.Launch()
	exitIfErrorf("could not launch Chromium: %v", err)
	context, err := browser.NewContext()
	exitIfErrorf("could not create context: %v", err)
	page, err := context.NewPage()
	exitIfErrorf("could not create page: %v", err)
	_, err = page.Goto("http://whatsmyuseragent.org/")
	exitIfErrorf("could not goto: %v", err)
	_, err = page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String("foo.png"),
	})
	exitIfErrorf("could not create screenshot: %v", err)
	exitIfErrorf("could not close browser: %v", browser.Close())
	exitIfErrorf("could not stop Playwright: %v", pw.Stop())
}
