//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mxschmitt/playwright-go"
)

func assertErrorToNilf(message string, err error) {
	if err != nil {
		log.Fatalf(message, err)
	}
}

func main() {
	startHttpServer()

	pw, err := playwright.Run()
	assertErrorToNilf("could not launch playwright: %w", err)
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	assertErrorToNilf("could not launch Chromium: %w", err)
	context, err := browser.NewContext()
	assertErrorToNilf("could not create context: %w", err)
	page, err := context.NewPage()
	assertErrorToNilf("could not create page: %w", err)
	_, err = page.Goto("http://localhost:1234")
	assertErrorToNilf("could not goto: %w", err)
	assertErrorToNilf("could not set content: %w", page.SetContent(`<a href="/download" download>download</a>`))
	download, err := page.ExpectDownload(func() error {
		return page.Locator("text=download").Click()
	})
	assertErrorToNilf("could not download: %w", err)
	fmt.Println(download.SuggestedFilename())
	assertErrorToNilf("could not close browser: %w", browser.Close())
	assertErrorToNilf("could not stop Playwright: %w", pw.Stop())
}

func startHttpServer() {
	http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/octet-stream")
		w.Header().Add("Content-Disposition", "attachment; filename=file.txt")
		if _, err := w.Write([]byte("foobar")); err != nil {
			log.Printf("could not write: %v", err)
		}
	})
	go func() {
		log.Fatal(http.ListenAndServe(":1234", nil))
	}()
}
