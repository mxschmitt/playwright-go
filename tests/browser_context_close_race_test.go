package playwright_test

import (
	"os"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestBrowserContextCloseRace tests for a data race between Close() and route handlers.
// This reproduces a race condition where Close() writes to closeWasCalled while
// route handler goroutines read it during page navigation.
//
// See: https://github.com/playwright-community/playwright-go/issues/566
func TestBrowserContextCloseRace(t *testing.T) {
	// Create a minimal HAR file
	harContent := `{
  "log": {
    "version": "1.2",
    "creator": {"name": "test", "version": "1.0"},
    "entries": [
      {
        "request": {
          "method": "GET",
          "url": "https://example.com/",
          "httpVersion": "HTTP/2.0",
          "headers": [],
          "queryString": [],
          "cookies": [],
          "headersSize": -1,
          "bodySize": 0
        },
        "response": {
          "status": 200,
          "statusText": "OK",
          "httpVersion": "HTTP/2.0",
          "headers": [{"name": "Content-Type", "value": "text/html"}],
          "cookies": [],
          "content": {
            "size": 13,
            "mimeType": "text/html",
            "text": "Hello, World!"
          },
          "redirectURL": "",
          "headersSize": -1,
          "bodySize": 13
        },
        "cache": {},
        "timings": {"send": 0, "wait": 0, "receive": 0}
      }
    ]
  }
}`

	harFile, err := os.CreateTemp("", "test-*.har")
	require.NoError(t, err)
	defer os.Remove(harFile.Name()) //nolint:errcheck

	_, err = harFile.WriteString(harContent)
	require.NoError(t, err)
	harFile.Close() //nolint:errcheck

	// Create a new context for this test (don't use BeforeEach)
	testContext, err := browser.NewContext()
	require.NoError(t, err)
	defer testContext.Close() //nolint:errcheck

	// Set up HAR replay - registers internal route handlers
	err = testContext.RouteFromHAR(harFile.Name(), playwright.BrowserContextRouteFromHAROptions{
		NotFound: playwright.HarNotFoundAbort,
	})
	require.NoError(t, err)

	// Add custom route handler
	err = testContext.Route("**/version.json*", func(route playwright.Route) {
		time.Sleep(5 * time.Millisecond) // Increase race window
		_ = route.Fulfill(playwright.RouteFulfillOptions{
			Status:      playwright.Int(200),
			ContentType: playwright.String("application/json"),
			Body:        playwright.String(`{"version": "1.0"}`),
		})
	})
	require.NoError(t, err)

	testPage, err := testContext.NewPage()
	require.NoError(t, err)

	// Start navigation in background
	done := make(chan error, 1)
	go func() {
		_, err := testPage.Goto("https://example.com/")
		done <- err
	}()

	// Give route handlers time to start processing
	time.Sleep(20 * time.Millisecond)

	// Close context while route handlers are actively running
	// This triggers the race between Close() and the route handler goroutines
	// Without proper synchronization, this will be detected by -race flag
	err = testContext.Close()
	require.NoError(t, err)

	// Wait for navigation to complete
	<-done
}

// TestPageCloseRace covers the sibling of the browser-context race: Page.Close()
// writes pageImpl.closeWasCalled while the page's own route-handler goroutine reads
// it during navigation. Without an atomic field this is detected by the -race flag.
//
// See: https://github.com/playwright-community/playwright-go/issues/566
func TestPageCloseRace(t *testing.T) {
	BeforeEach(t)

	// Abort (rather than fulfill) the request so the navigation never commits.
	// This still exercises the close-vs-route-handler race this test guards
	// against (the handler goroutine runs concurrently with Close() and reads
	// pageImpl.closeWasCalled), but avoids a Chromium-level deadlock where
	// Page.Close() never returns if it races a navigation that commits at the
	// same instant.
	require.NoError(t, page.Route("**/*", func(route playwright.Route) {
		time.Sleep(5 * time.Millisecond) // increase race window
		_ = route.Abort()
	}))

	// Start navigation in background so route handlers run concurrently with Close().
	done := make(chan error, 1)
	go func() {
		_, err := page.Goto(server.EMPTY_PAGE)
		done <- err
	}()

	// Give the route handler time to start processing.
	time.Sleep(10 * time.Millisecond)

	// Close the page while its route handler goroutine is reading closeWasCalled.
	require.NoError(t, page.Close())

	<-done
}
