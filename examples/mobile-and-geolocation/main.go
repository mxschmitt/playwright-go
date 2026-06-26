//go:build ignore
// +build ignore

package main

import (
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
	device := pw.Devices["Pixel 5"]
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		Geolocation: &playwright.Geolocation{
			Longitude: 12.492507,
			Latitude:  41.889938,
		},
		Permissions:       []string{"geolocation"},
		Viewport:          device.Viewport,
		UserAgent:         playwright.String(device.UserAgent),
		DeviceScaleFactor: playwright.Float(device.DeviceScaleFactor),
		IsMobile:          playwright.Bool(device.IsMobile),
		HasTouch:          playwright.Bool(device.HasTouch),
	})
	if err != nil {
		log.Fatalf("could not create context: %v", err)
	}
	page, err := context.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}
	if _, err = page.Goto("https://www.openstreetmap.org"); err != nil {
		log.Fatalf("could not goto: %v", err)
	}
	if err = page.Locator("a[data-bs-original-title='Show My Location']").Click(); err != nil {
		log.Fatalf("could not click on location: %v", err)
	}
	if _, err = page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String("colosseum-iphone.png"),
	}); err != nil {
		log.Fatalf("could not make screenshot: %v", err)
	}
	if err = browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}
}
