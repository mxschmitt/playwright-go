package playwright_test

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestShouldWork(t *testing.T) {
	harPath := filepath.Join(t.TempDir(), "log.har")
	BeforeEach(t, playwright.BrowserNewContextOptions{
		RecordHarPath: playwright.String(harPath),
	})
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, context.Close())
	require.FileExists(t, harPath)
	data, err := os.ReadFile(harPath)
	require.NoError(t, err)
	require.Contains(t, string(data), "log")
}

func TestShouldOmitContent(t *testing.T) {
	harPath := filepath.Join(t.TempDir(), "log.har")
	BeforeEach(t, playwright.BrowserNewContextOptions{
		RecordHarPath:    playwright.String(harPath),
		RecordHarContent: playwright.HarContentPolicyOmit,
	})

	_, err := page.Goto(server.PREFIX + "/har.html")
	require.NoError(t, err)
	require.NoError(t, context.Close())
	require.FileExists(t, harPath)
	data, err := os.ReadFile(harPath)
	require.NoError(t, err)
	result := gjson.GetBytes(data, "log.entries.0.response.content.text")
	require.False(t, result.Exists())
	result = gjson.GetBytes(data, "log.entries.0.response.content.encoding")
	require.False(t, result.Exists())
}

func TestShouldOmitContentLegacy(t *testing.T) {
	harPath := filepath.Join(t.TempDir(), "log.har")
	BeforeEach(t, playwright.BrowserNewContextOptions{
		RecordHarPath:        playwright.String(harPath),
		RecordHarOmitContent: playwright.Bool(true),
	})

	_, err := page.Goto(server.PREFIX + "/har.html")
	require.NoError(t, err)
	require.NoError(t, context.Close())
	require.FileExists(t, harPath)
	data, err := os.ReadFile(harPath)
	require.NoError(t, err)
	result := gjson.GetBytes(data, "log.entries.0.response.content.text")
	require.False(t, result.Exists())
	result = gjson.GetBytes(data, "log.entries.0.response.content.encoding")
	require.False(t, result.Exists())
}

func TestShouldAttachContent(t *testing.T) {
	harZipFile := filepath.Join(t.TempDir(), "log.har.zip")
	BeforeEach(t, playwright.BrowserNewContextOptions{
		RecordHarPath:    playwright.String(harZipFile),
		RecordHarContent: playwright.HarContentPolicyAttach,
	})

	_, err := page.Goto(server.PREFIX + "/har.html")
	require.NoError(t, err)
	_, err = page.Evaluate(`() => fetch('/pptr.png').then(r => r.arrayBuffer())`)
	require.NoError(t, err)
	require.NoError(t, context.Close())
	require.FileExists(t, harZipFile)
	data, err := readFromZip(harZipFile, "har.har")
	require.NoError(t, err)
	entries := gjson.GetBytes(data, "log.entries")
	require.True(t, entries.Exists())
	require.Equal(t, "text/html; charset=utf-8", entries.Get("0.response.content.mimeType").String())
	require.Contains(t, entries.Get("0.response.content._file").String(), "75841480e2606c03389077304342fac2c58ccb1b")
	require.GreaterOrEqual(t, entries.Get("0.response.content.size").Int(), int64(96))
	require.Zero(t, entries.Get("0.response.content.compression").Int())

	require.Equal(t, "text/css; charset=utf-8", entries.Get("1.response.content.mimeType").String())
	require.Contains(t, entries.Get("1.response.content._file").String(), "79f739d7bc88e80f55b9891a22bf13a2b4e18adb")
	require.GreaterOrEqual(t, entries.Get("1.response.content.size").Int(), int64(37))
	require.Zero(t, entries.Get("1.response.content.compression").Int())

	require.Equal(t, "image/png", entries.Get("2.response.content.mimeType").String())
	require.Contains(t, entries.Get("2.response.content._file").String(), "a4c3a18f0bb83f5d9fe7ce561e065c36205762fa")
	require.GreaterOrEqual(t, entries.Get("2.response.content.size").Int(), int64(6000))
	require.Zero(t, entries.Get("2.response.content.compression").Int())
}

func TestShouldNotOmitContent(t *testing.T) {
	harPath := filepath.Join(t.TempDir(), "log.har")
	BeforeEach(t, playwright.BrowserNewContextOptions{
		RecordHarPath:        playwright.String(harPath),
		RecordHarOmitContent: playwright.Bool(false),
	})

	_, err := page.Goto(server.PREFIX + "/har.html")
	require.NoError(t, err)
	require.NoError(t, context.Close())
	require.FileExists(t, harPath)
	data, err := os.ReadFile(harPath)
	require.NoError(t, err)
	result := gjson.GetBytes(data, "log.entries.0.response.content.text")
	require.True(t, result.Exists())
}

func TestShouldIncludeContent(t *testing.T) {
	harPath := filepath.Join(t.TempDir(), "log.har")
	BeforeEach(t, playwright.BrowserNewContextOptions{
		RecordHarPath: playwright.String(harPath),
	})

	_, err := page.Goto(server.PREFIX + "/har.html")
	require.NoError(t, err)
	require.NoError(t, context.Close())
	require.FileExists(t, harPath)
	data, err := os.ReadFile(harPath)
	require.NoError(t, err)
	content := gjson.GetBytes(data, "log.entries.0.response.content")
	require.Equal(t, "text/html; charset=utf-8", content.Get("mimeType").String())
	require.Contains(t, content.Get("text").String(), "HAR Page")
}

func TestShouldDefaultToFullMode(t *testing.T) {
	harPath := filepath.Join(t.TempDir(), "log.har")
	BeforeEach(t, playwright.BrowserNewContextOptions{
		RecordHarPath: playwright.String(harPath),
	})

	_, err := page.Goto(server.PREFIX + "/har.html")
	require.NoError(t, err)
	require.NoError(t, context.Close())
	require.FileExists(t, harPath)
	data, err := os.ReadFile(harPath)
	require.NoError(t, err)
	bodySize := gjson.GetBytes(data, "log.entries.0.request.bodySize")
	require.True(t, bodySize.Exists())
	require.GreaterOrEqual(t, bodySize.Int(), int64(0))
}

func TestShouldSupportMinimalMode(t *testing.T) {
	harPath := filepath.Join(t.TempDir(), "log.har")
	BeforeEach(t, playwright.BrowserNewContextOptions{
		RecordHarPath: playwright.String(harPath),
		RecordHarMode: playwright.HarModeMinimal,
	})

	_, err := page.Goto(server.PREFIX + "/har.html")
	require.NoError(t, err)
	require.NoError(t, context.Close())
	require.FileExists(t, harPath)
	data, err := os.ReadFile(harPath)
	require.NoError(t, err)
	bodySize := gjson.GetBytes(data, "log.entries.0.request.bodySize")
	require.True(t, bodySize.Exists())
	require.Equal(t, bodySize.Int(), int64(-1))
}

func TestShouldFilterByGlob(t *testing.T) {
	harPath := filepath.Join(t.TempDir(), "log.har")
	BeforeEach(t, playwright.BrowserNewContextOptions{
		BaseURL:            &server.PREFIX,
		RecordHarPath:      playwright.String(harPath),
		RecordHarURLFilter: "/*.css",
	})

	_, err := page.Goto(server.PREFIX + "/har.html")
	require.NoError(t, err)
	require.NoError(t, context.Close())
	require.FileExists(t, harPath)
	data, err := os.ReadFile(harPath)
	require.NoError(t, err)
	require.Equal(t, len(gjson.GetBytes(data, "log.entries").Array()), 1)
	url := gjson.GetBytes(data, "log.entries.0.request.url")
	require.True(t, strings.HasSuffix(url.String(), "one-style.css"))
}

func TestShouldFilterByRegexp(t *testing.T) {
	harPath := filepath.Join(t.TempDir(), "log.har")
	BeforeEach(t, playwright.BrowserNewContextOptions{
		BaseURL:            &server.PREFIX,
		RecordHarPath:      playwright.String(harPath),
		RecordHarURLFilter: regexp.MustCompile("(?i)HAR.X?HTML"),
		IgnoreHttpsErrors:  playwright.Bool(true),
	})

	_, err := page.Goto(server.PREFIX + "/har.html")
	require.NoError(t, err)
	require.NoError(t, context.Close())
	require.FileExists(t, harPath)
	data, err := os.ReadFile(harPath)
	require.NoError(t, err)
	require.Equal(t, len(gjson.GetBytes(data, "log.entries").Array()), 1)
	url := gjson.GetBytes(data, "log.entries.0.request.url")
	require.True(t, strings.HasSuffix(url.String(), "har.html"))
}

func TestShouldContextRouteFromHarMatchingTheMethodAndFollowingRedirects(t *testing.T) {
	BeforeEach(t)

	err := context.RouteFromHAR(Asset("har-fulfill.har"))
	require.NoError(t, err)
	_, err = page.Goto("http://no.playwright/")
	require.NoError(t, err)
	// HAR contains a redirect for the script that should be followed automatically.
	data, err := page.Evaluate(`window.value`)
	require.NoError(t, err)
	require.Equal(t, "foo", data)
	// HAR contains a POST for the css file that should not be used.
	require.NoError(t, expect.Locator(page.Locator("body")).ToHaveCSS("background-color", "rgb(255, 0, 0)"))
}

func TestShouldPageRouteFromHarMatchingTheMethodAndFollowingRedirects(t *testing.T) {
	BeforeEach(t)

	err := page.RouteFromHAR(Asset("har-fulfill.har"))
	require.NoError(t, err)
	_, err = page.Goto("http://no.playwright/")
	require.NoError(t, err)
	// HAR contains a redirect for the script that should be followed automatically.
	data, err := page.Evaluate(`window.value`)
	require.NoError(t, err)
	require.Equal(t, "foo", data)
	// HAR contains a POST for the css file that should not be used.
	require.NoError(t, expect.Locator(page.Locator("body")).ToHaveCSS("background-color", "rgb(255, 0, 0)"))
}

func TestFallbackContinueShouldContinueWhenNotFoundInHar(t *testing.T) {
	BeforeEach(t)

	err := context.RouteFromHAR(Asset("har-fulfill.har"), playwright.BrowserContextRouteFromHAROptions{
		NotFound: playwright.HarNotFoundFallback,
	})
	require.NoError(t, err)
	_, err = page.Goto(server.PREFIX + "/one-style.html")
	require.NoError(t, err)
	require.NoError(t, expect.Locator(page.Locator("body")).ToHaveCSS("background-color", "rgb(255, 192, 203)"))
}

func TestByDefaultShouldAbortRequestsNotFoundInHar(t *testing.T) {
	BeforeEach(t)

	err := context.RouteFromHAR(Asset("har-fulfill.har"))
	require.NoError(t, err)

	_, err = page.Goto(server.EMPTY_PAGE)
	if isChromium {
		require.ErrorContains(t, err, "net::ERR_FAILED")
	} else if isWebKit {
		require.ErrorContains(t, err, "Blocked by Web Inspector")
	} else {
		require.ErrorContains(t, err, "NS_ERROR_FAILURE")
	}
}

func TestFallbackContinueShouldContinueRequestsOnBadHar(t *testing.T) {
	BeforeEach(t)

	harPath := filepath.Join(t.TempDir(), "invalid.har")
	require.NoError(t, os.WriteFile(harPath, []byte(`{"log": {}}`), 0o644))
	err := context.RouteFromHAR(harPath, playwright.BrowserContextRouteFromHAROptions{
		NotFound: playwright.HarNotFoundFallback,
	})
	require.NoError(t, err)
	_, err = page.Goto(server.PREFIX + "/one-style.html")
	require.NoError(t, err)
	require.NoError(t, expect.Locator(page.Locator("body")).ToHaveCSS("background-color", "rgb(255, 192, 203)"))
}

func TestShouldOnlyHandleRequestsMatchingUrlFilter(t *testing.T) {
	BeforeEach(t)

	err := context.RouteFromHAR(Asset("har-fulfill.har"), playwright.BrowserContextRouteFromHAROptions{
		NotFound: playwright.HarNotFoundFallback,
		URL:      "**/*.js",
	})
	require.NoError(t, err)

	err = context.Route("http://no.playwright/", func(r playwright.Route) {
		require.Equal(t, r.Request().URL(), "http://no.playwright/")
		require.NoError(t, r.Fulfill(playwright.RouteFulfillOptions{
			Status:      playwright.Int(200),
			ContentType: playwright.String("text/html"),
			Body:        []byte(`<script src="./script.js"></script><div>hello</div>`),
		}))
	})
	require.NoError(t, err)
	_, err = page.Goto("http://no.playwright/")
	require.NoError(t, err)
	data, err := page.Evaluate(`window.value`)
	require.NoError(t, err)
	require.Equal(t, "foo", data)
	require.NoError(t, expect.Locator(page.Locator("body")).ToHaveCSS("background-color", "rgba(0, 0, 0, 0)"))
}

func TestShouldOnlyHandleRequestsMatchingUrlFilterNoFallback(t *testing.T) {
	BeforeEach(t)

	err := context.RouteFromHAR(Asset("har-fulfill.har"), playwright.BrowserContextRouteFromHAROptions{
		URL: "**/*.js",
	})
	require.NoError(t, err)

	err = context.Route("http://no.playwright/", func(r playwright.Route) {
		require.Equal(t, r.Request().URL(), "http://no.playwright/")
		require.NoError(t, r.Fulfill(playwright.RouteFulfillOptions{
			Status:      playwright.Int(200),
			ContentType: playwright.String("text/html"),
			Body:        []byte(`<script src="./script.js"></script><div>hello</div>`),
		}))
	})
	require.NoError(t, err)
	_, err = page.Goto("http://no.playwright/")
	require.NoError(t, err)
	data, err := page.Evaluate(`window.value`)
	require.NoError(t, err)
	require.Equal(t, "foo", data)
	require.NoError(t, expect.Locator(page.Locator("body")).ToHaveCSS("background-color", "rgba(0, 0, 0, 0)"))
}

func TestShouldOnlyHandleRequestsMatchingUrlFilterNoFallbackPage(t *testing.T) {
	BeforeEach(t)

	err := page.RouteFromHAR(Asset("har-fulfill.har"), playwright.PageRouteFromHAROptions{
		URL: "**/*.js",
	})
	require.NoError(t, err)

	err = page.Route("http://no.playwright/", func(r playwright.Route) {
		require.Equal(t, r.Request().URL(), "http://no.playwright/")
		require.NoError(t, r.Fulfill(playwright.RouteFulfillOptions{
			Status:      playwright.Int(200),
			ContentType: playwright.String("text/html"),
			Body:        []byte(`<script src="./script.js"></script><div>hello</div>`),
		}))
	})
	require.NoError(t, err)
	_, err = page.Goto("http://no.playwright/")
	require.NoError(t, err)
	data, err := page.Evaluate(`window.value`)
	require.NoError(t, err)
	require.Equal(t, "foo", data)
	require.NoError(t, expect.Locator(page.Locator("body")).ToHaveCSS("background-color", "rgba(0, 0, 0, 0)"))
}

func TestShouldSupportRegexFilter(t *testing.T) {
	BeforeEach(t)

	urlPattern := regexp.MustCompile(`.*(\.js|.*\.css|no.playwright\/)`)
	err := context.RouteFromHAR(Asset("har-fulfill.har"), playwright.BrowserContextRouteFromHAROptions{
		URL: urlPattern,
	})
	require.NoError(t, err)
	_, err = page.Goto("http://no.playwright/")
	require.NoError(t, err)
	data, err := page.Evaluate(`window.value`)
	require.NoError(t, err)
	require.Equal(t, "foo", data)
	require.NoError(t, expect.Locator(page.Locator("body")).ToHaveCSS("background-color", "rgb(255, 0, 0)"))
}

func TestShouldGoBackToRedirectedNavigation(t *testing.T) {
	BeforeEach(t)

	if isWebKit && runtime.GOOS == "linux" {
		t.Skip("flaky: webkit on linux")
	}
	urlPattern := regexp.MustCompile(`/.*theverge.*/`)
	err := context.RouteFromHAR(Asset("har-redirect.har"), playwright.BrowserContextRouteFromHAROptions{
		URL: urlPattern,
	})
	require.NoError(t, err)
	_, err = page.Goto("https://theverge.com/")
	require.NoError(t, err)
	response, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Equal(t, server.EMPTY_PAGE, response.URL())

	response, err = page.GoBack()
	require.NoError(t, err)
	require.Equal(t, "https://www.theverge.com/", response.URL())
	data, err := page.Evaluate("window.location.href")
	require.NoError(t, err)
	require.Equal(t, "https://www.theverge.com/", data)
}

func TestShouldGoForwardToRedirectedNavigation(t *testing.T) {
	BeforeEach(t)

	if isFirefox {
		// skipped upstream (https://github.com/microsoft/playwright/blob/6a8d835145e2f4002ee00b67a80a1f70af956703/tests/library/browsercontext-har.spec.ts#L214)
		t.Skip("skipped upstream")
	}
	if isWebKit && runtime.GOOS == "linux" {
		t.Skip("flaky: webkit on linux")
	}
	urlPattern := regexp.MustCompile(`/.*theverge.*/`)
	err := context.RouteFromHAR(Asset("har-redirect.har"), playwright.BrowserContextRouteFromHAROptions{
		URL: urlPattern,
	})
	require.NoError(t, err)
	_, err = page.Goto("https://theverge.com/")
	require.NoError(t, err)
	response, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Equal(t, server.EMPTY_PAGE, response.URL())
	response, err = page.Goto("https://theverge.com/")
	require.NoError(t, err)
	require.Equal(t, "https://www.theverge.com/", response.URL())
	response, err = page.GoBack()
	require.NoError(t, err)
	require.Equal(t, server.EMPTY_PAGE, response.URL())
	response, err = page.GoForward()
	require.NoError(t, err)
	require.Equal(t, "https://www.theverge.com/", response.URL())
	data, err := page.Evaluate("window.location.href")
	require.NoError(t, err)
	require.Equal(t, "https://www.theverge.com/", data)
}

func TestShouldReloadRedirectedNavigation(t *testing.T) {
	BeforeEach(t)

	if isWebKit && runtime.GOOS == "linux" {
		t.Skip("flaky: webkit on linux")
	}
	urlPattern := regexp.MustCompile(`/.*theverge.*/`)
	err := context.RouteFromHAR(Asset("har-redirect.har"), playwright.BrowserContextRouteFromHAROptions{
		URL: urlPattern,
	})
	require.NoError(t, err)
	response, err := page.Goto("https://theverge.com/")
	require.NoError(t, err)
	if response != nil { // flaky
		require.Equal(t, "https://www.theverge.com/", response.URL())
	}
	response, err = page.Reload()
	require.NoError(t, err)
	if response != nil {
		require.Equal(t, "https://www.theverge.com/", response.URL())
	}
	data, err := page.Evaluate("window.location.href")
	require.NoError(t, err)
	require.Equal(t, "https://www.theverge.com/", data)
}

func TestShouldFulfillFromHarWithContentInAFile(t *testing.T) {
	BeforeEach(t)

	err := context.RouteFromHAR(Asset("har-sha1.har"))
	require.NoError(t, err)
	_, err = page.Goto("http://no.playwright/")
	require.NoError(t, err)
	content, err := page.Content()
	require.NoError(t, err)
	require.Equal(t, "<html><head></head><body>Hello, world</body></html>", content)
}

func TestShouldRoundTripHarZip(t *testing.T) {
	harPath := filepath.Join(t.TempDir(), "har.zip")
	BeforeEach(t, playwright.BrowserNewContextOptions{
		RecordHarMode: playwright.HarModeMinimal,
		RecordHarPath: playwright.String(harPath),
	})

	_, err := page.Goto(server.PREFIX + "/one-style.html")
	require.NoError(t, err)
	require.NoError(t, context.Close())

	context2, page2 := newBrowserContextAndPage(t, playwright.BrowserNewContextOptions{})
	require.NoError(t, err)

	err = context2.RouteFromHAR(harPath, playwright.BrowserContextRouteFromHAROptions{
		NotFound: playwright.HarNotFoundAbort,
	})
	require.NoError(t, err)
	_, err = page2.Goto(server.PREFIX + "/one-style.html")
	require.NoError(t, err)
	content, err := page2.Content()
	require.NoError(t, err)
	require.Contains(t, content, "hello, world!")
}

func TestShouldRoundTripHarWithPostData(t *testing.T) {
	harPath := filepath.Join(t.TempDir(), "har.zip")
	BeforeEach(t, playwright.BrowserNewContextOptions{
		RecordHarMode: playwright.HarModeMinimal,
		RecordHarPath: playwright.String(harPath),
	})

	server.SetRoute("/echo", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		_, err = w.Write(body)
		require.NoError(t, err)
	})
	const fetchFunction = `async (body) => {
			const response = await fetch('/echo', { method: 'POST', body });
			return response.text();
		}`

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)

	data, err := page.Evaluate(fetchFunction, "1")
	require.NoError(t, err)
	require.Equal(t, "1", data)
	data, err = page.Evaluate(fetchFunction, "2")
	require.NoError(t, err)
	require.Equal(t, "2", data)
	data, err = page.Evaluate(fetchFunction, "3")
	require.NoError(t, err)
	require.Equal(t, "3", data)
	require.NoError(t, context.Close())

	context2, page2 := newBrowserContextAndPage(t, playwright.BrowserNewContextOptions{})
	err = context2.RouteFromHAR(harPath, playwright.BrowserContextRouteFromHAROptions{
		NotFound: playwright.HarNotFoundAbort,
	})
	require.NoError(t, err)

	_, err = page2.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	data, err = page2.Evaluate(fetchFunction, "1")
	require.NoError(t, err)
	require.Equal(t, "1", data)
	data, err = page2.Evaluate(fetchFunction, "2")
	require.NoError(t, err)
	require.Equal(t, "2", data)
	data, err = page2.Evaluate(fetchFunction, "3")
	require.NoError(t, err)
	require.Equal(t, "3", data)
	_, err = page2.Evaluate(fetchFunction, "4")
	require.Error(t, err)
}

func TestShouldDisambiguateByHeader(t *testing.T) {
	harPath := filepath.Join(t.TempDir(), "har.zip")
	BeforeEach(t, playwright.BrowserNewContextOptions{
		RecordHarMode: playwright.HarModeMinimal,
		RecordHarPath: playwright.String(harPath),
	})

	server.SetRoute("/echo", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(r.Header.Get("baz")))
		require.NoError(t, err)
	})
	const fetchFunction = `
		async (bazValue) => {
			const response = await fetch('/echo', {
			method: 'POST',
			body: '',
			headers: {
					foo: 'foo-value',
					bar: 'bar-value',
					baz: bazValue,
			}
			});
			return response.text();
		}`

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)

	data, err := page.Evaluate(fetchFunction, "baz1")
	require.NoError(t, err)
	require.Equal(t, "baz1", data)
	data, err = page.Evaluate(fetchFunction, "baz2")
	require.NoError(t, err)
	require.Equal(t, "baz2", data)
	data, err = page.Evaluate(fetchFunction, "baz3")
	require.NoError(t, err)
	require.Equal(t, "baz3", data)
	require.NoError(t, context.Close())

	context2, page2 := newBrowserContextAndPage(t, playwright.BrowserNewContextOptions{})
	require.NoError(t, context2.RouteFromHAR(harPath))
	_, err = page2.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	data, err = page2.Evaluate(fetchFunction, "baz1")
	require.NoError(t, err)
	require.Equal(t, "baz1", data)
	data, err = page2.Evaluate(fetchFunction, "baz2")
	require.NoError(t, err)
	require.Equal(t, "baz2", data)
	data, err = page2.Evaluate(fetchFunction, "baz3")
	require.NoError(t, err)
	require.Equal(t, "baz3", data)
	_, err = page2.Evaluate(fetchFunction, "baz4")
	require.NoError(t, err)
	// why does this equals baz1 in playwright-python?
	require.Equal(t, "baz3", data)
}

func TestShouldProduceExtractedZip(t *testing.T) {
	harPath := filepath.Join(t.TempDir(), "har.har")
	BeforeEach(t, playwright.BrowserNewContextOptions{
		RecordHarMode:    playwright.HarModeMinimal,
		RecordHarPath:    playwright.String(harPath),
		RecordHarContent: playwright.HarContentPolicyAttach,
	})

	_, err := page.Goto(server.PREFIX + "/one-style.html")
	require.NoError(t, err)
	require.NoError(t, context.Close())

	require.FileExists(t, harPath)
	content, err := os.ReadFile(harPath)
	require.NoError(t, err)
	require.Contains(t, string(content), "log")
	require.NotContains(t, string(content), "background-color")

	context2, page2 := newBrowserContextAndPage(t, playwright.BrowserNewContextOptions{})
	require.NoError(t, context2.RouteFromHAR(harPath, playwright.BrowserContextRouteFromHAROptions{
		NotFound: playwright.HarNotFoundAbort,
	}))

	response, err := page2.Goto(server.PREFIX + "/one-style.html")
	require.NoError(t, err)
	body, err := response.Body()
	require.NoError(t, err)
	require.Contains(t, string(body), "hello, world!")
	require.NoError(t, expect.Locator(page2.Locator("body")).ToHaveCSS("background-color", "rgb(255, 192, 203)"))
}

func TestShouldUpdateHarZipForContext(t *testing.T) {
	BeforeEach(t)

	harPath := filepath.Join(t.TempDir(), "har.zip")
	require.NoError(t, context.RouteFromHAR(harPath, playwright.BrowserContextRouteFromHAROptions{
		Update: playwright.Bool(true),
	}))
	_, err := page.Goto(server.PREFIX + "/one-style.html")
	require.NoError(t, err)
	require.NoError(t, context.Close())

	require.FileExists(t, harPath)

	context2, page2 := newBrowserContextAndPage(t, playwright.BrowserNewContextOptions{})
	require.NoError(t, err)
	require.NoError(t, context2.RouteFromHAR(harPath, playwright.BrowserContextRouteFromHAROptions{
		NotFound: playwright.HarNotFoundAbort,
	}))

	response, err := page2.Goto(server.PREFIX + "/one-style.html")
	require.NoError(t, err)
	body, err := response.Body()
	require.NoError(t, err)
	require.Contains(t, string(body), "hello, world!")
	require.NoError(t, expect.Locator(page2.Locator("body")).ToHaveCSS("background-color", "rgb(255, 192, 203)"))
}

func TestPageUnrouteAllShouldStopPageRouteFromHAR(t *testing.T) {
	BeforeEach(t)

	harPath := Asset("har-fulfill.har")
	// The har file contains requests for another domain, so the router
	// is expected to abort all requests.
	require.NoError(t, page.RouteFromHAR(harPath, playwright.PageRouteFromHAROptions{
		NotFound: playwright.HarNotFoundAbort,
	}))
	_, err := page.Goto(server.EMPTY_PAGE)
	require.Error(t, err)
	require.NoError(t, page.UnrouteAll(playwright.PageUnrouteAllOptions{
		Behavior: playwright.UnrouteBehaviorWait,
	}))
	response, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.True(t, response.Ok())
}

func TestContextUnrouteCallShouldStopContextRouteFromHAR(t *testing.T) {
	BeforeEach(t)

	harPath := Asset("har-fulfill.har")
	// The har file contains requests for another domain, so the router
	// is expected to abort all requests.
	require.NoError(t, context.RouteFromHAR(harPath, playwright.BrowserContextRouteFromHAROptions{
		NotFound: playwright.HarNotFoundAbort,
	}))
	_, err := page.Goto(server.EMPTY_PAGE)
	require.Error(t, err)
	require.NoError(t, context.UnrouteAll(playwright.BrowserContextUnrouteAllOptions{
		Behavior: playwright.UnrouteBehaviorWait,
	}))
	response, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.True(t, response.Ok())
}

func TestShouldUpdateHarZipForPage(t *testing.T) {
	BeforeEach(t)

	harPath := filepath.Join(t.TempDir(), "har.zip")
	require.NoError(t, page.RouteFromHAR(harPath, playwright.PageRouteFromHAROptions{
		Update: playwright.Bool(true),
	}))
	_, err := page.Goto(server.PREFIX + "/one-style.html")
	require.NoError(t, err)
	require.NoError(t, context.Close())

	require.FileExists(t, harPath)

	_, page2 := newBrowserContextAndPage(t, playwright.BrowserNewContextOptions{})
	require.NoError(t, page2.RouteFromHAR(harPath, playwright.PageRouteFromHAROptions{
		NotFound: playwright.HarNotFoundAbort,
	}))

	response, err := page2.Goto(server.PREFIX + "/one-style.html")
	require.NoError(t, err)
	body, err := response.Body()
	require.NoError(t, err)
	require.Contains(t, string(body), "hello, world!")
	require.NoError(t, expect.Locator(page2.Locator("body")).ToHaveCSS("background-color", "rgb(255, 192, 203)"))
}

func TestShouldUpdateHarZipForPageWithDifferentOptions(t *testing.T) {
	BeforeEach(t)

	harPath := filepath.Join(t.TempDir(), "har.zip")
	require.NoError(t, page.RouteFromHAR(harPath, playwright.PageRouteFromHAROptions{
		Update:        playwright.Bool(true),
		UpdateContent: playwright.RouteFromHarUpdateContentPolicyEmbed,
		UpdateMode:    playwright.HarModeFull,
	}))
	_, err := page.Goto(server.PREFIX + "/one-style.html")
	require.NoError(t, err)
	require.NoError(t, context.Close())

	require.FileExists(t, harPath)

	_, page2 := newBrowserContextAndPage(t, playwright.BrowserNewContextOptions{})
	require.NoError(t, err)
	err = page2.RouteFromHAR(harPath, playwright.PageRouteFromHAROptions{
		NotFound: playwright.HarNotFoundAbort,
	})
	require.NoError(t, err)
	response, err := page2.Goto(server.PREFIX + "/one-style.html")
	require.NoError(t, err)
	body, err := response.Body()
	require.NoError(t, err)
	require.Contains(t, string(body), "hello, world!")
	require.NoError(t, expect.Locator(page2.Locator("body")).ToHaveCSS("background-color", "rgb(255, 192, 203)"))
}

func TestShouldUpdateExtractedHarZipForPage(t *testing.T) {
	BeforeEach(t)

	harPath := filepath.Join(t.TempDir(), "har.har")
	require.NoError(t, page.RouteFromHAR(harPath, playwright.PageRouteFromHAROptions{
		Update: playwright.Bool(true),
	}))
	_, err := page.Goto(server.PREFIX + "/one-style.html")
	require.NoError(t, err)
	require.NoError(t, context.Close())

	require.FileExists(t, harPath)
	content, err := os.ReadFile(harPath)
	require.NoError(t, err)
	require.Contains(t, string(content), "log")
	require.NotContains(t, string(content), "background-color")

	_, page2 := newBrowserContextAndPage(t, playwright.BrowserNewContextOptions{})
	require.NoError(t, err)
	err = page2.RouteFromHAR(harPath, playwright.PageRouteFromHAROptions{
		NotFound: playwright.HarNotFoundAbort,
	})
	require.NoError(t, err)
	response, err := page2.Goto(server.PREFIX + "/one-style.html")
	require.NoError(t, err)
	body, err := response.Body()
	require.NoError(t, err)
	require.Contains(t, string(body), "hello, world!")
	require.NoError(t, expect.Locator(page2.Locator("body")).ToHaveCSS("background-color", "rgb(255, 192, 203)"))
}

func TestShouldErrorWhenWrongZipFile(t *testing.T) {
	BeforeEach(t)

	require.Error(t, page.RouteFromHAR(Asset("chromium-linux.zip")))
}
