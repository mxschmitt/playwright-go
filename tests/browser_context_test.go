package playwright_test

import (
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"runtime"
	"sync/atomic"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestBrowserContextNewPage(t *testing.T) {
	BeforeEach(t)

	require.Equal(t, context.Browser(), browser)
}

func TestBrowserContextNewContext(t *testing.T) {
	BeforeEach(t)

	require.Equal(t, 1, len(browser.Contexts()))
	context2, err := browser.NewContext()
	require.NoError(t, err)
	require.Equal(t, 2, len(browser.Contexts()))
	require.Equal(t, browser.Contexts()[1], context2)
	require.Equal(t, context2.Browser(), browser)
	require.NoError(t, context2.Close())
	require.Equal(t, 1, len(browser.Contexts()))
	require.Equal(t, context.Browser(), browser)
}

func TestBrowserContextClose(t *testing.T) {
	BeforeEach(t)

	context2, err := browser.NewContext()
	require.NoError(t, err)
	require.Equal(t, 2, len(browser.Contexts()))
	require.Equal(t, context, browser.Contexts()[0])
	require.Equal(t, context2, browser.Contexts()[1])
	require.NoError(t, context.Close())
	require.Equal(t, 1, len(browser.Contexts()))
	require.NoError(t, context2.Close())
	require.Equal(t, 0, len(browser.Contexts()))
}

func TestBrowserContextCloseWithHarDownload(t *testing.T) {
	BeforeEach(t)

	tmpFile := filepath.Join(t.TempDir(), "test.har")
	context2, err := browser.NewContext(playwright.BrowserNewContextOptions{
		RecordHarPath:        playwright.String(tmpFile),
		RecordHarOmitContent: playwright.Bool(true),
	})
	require.NoError(t, err)
	require.NoError(t, context.Close())
	require.NoFileExists(t, tmpFile)
	require.NoError(t, context2.Close())
	require.FileExists(t, tmpFile)
}

func TestBrowserContextOffline(t *testing.T) {
	BeforeEach(t)

	offline, err := page.Evaluate("window.navigator.onLine")
	require.NoError(t, err)
	require.True(t, offline.(bool))

	require.NoError(t, context.SetOffline(true))
	offline, err = page.Evaluate("window.navigator.onLine")
	require.NoError(t, err)
	require.False(t, offline.(bool))

	require.NoError(t, context.SetOffline(false))
	offline, err = page.Evaluate("window.navigator.onLine")
	require.NoError(t, err)
	require.True(t, offline.(bool))
}

func TestBrowserContextSetExtraHTTPHeaders(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, context.SetExtraHTTPHeaders(map[string]string{
		"extra-http": "42",
	}))
	intercepted := make(chan bool, 1)
	err := page.Route("**/empty.html", func(route playwright.Route) {
		require.NoError(t, route.Continue())
		intercepted <- true
	})
	require.NoError(t, err)
	response, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.True(t, response.Ok())
	<-intercepted
}

func TestBrowserContextSetHttpCredentials(t *testing.T) {
	BeforeEach(t)

	server.SetBasicAuth("/empty.html", "user", "pass")

	response, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Equal(t, 401, response.Status())
	context.Close()

	context, page = newBrowserContextAndPage(t, playwright.BrowserNewContextOptions{
		AcceptDownloads: playwright.Bool(true),
		HasTouch:        playwright.Bool(true),
		HttpCredentials: &playwright.HttpCredentials{
			Username: "user",
			Password: "pass",
		},
	})
	response, err = page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Equal(t, 200, response.Status())
}

func TestBrowserContextNewCDPSession(t *testing.T) {
	BeforeEach(t)

	cdpSession, err := page.Context().NewCDPSession(page)
	if isChromium {
		require.NoError(t, err)
		require.NoError(t, cdpSession.Detach())
	} else {
		require.Error(t, err)
	}
}

func TestBrowserContextSetGeolocation(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, context.GrantPermissions([]string{"geolocation"}))
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, context.SetGeolocation(&playwright.Geolocation{
		Longitude: 10,
		Latitude:  10,
	}))
	geolocation, err := page.Evaluate(`() => new Promise(resolve => navigator.geolocation.getCurrentPosition(position => {
      resolve({latitude: position.coords.latitude, longitude: position.coords.longitude});
    }))`)
	require.NoError(t, err)
	require.Equal(t, geolocation, map[string]interface{}{
		"latitude":  10,
		"longitude": 10,
	})
	require.NoError(t, context.ClearPermissions())
}

func TestBrowserContextAddCookies(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, context.AddCookies([]playwright.OptionalCookie{
		{
			Name:  "password",
			Value: "123456",
			URL:   playwright.String(server.EMPTY_PAGE),
		},
	}))
	cookie, err := page.Evaluate("() => document.cookie")
	require.NoError(t, err)
	require.Equal(t, "password=123456", cookie)

	cookies, err := context.Cookies()
	require.NoError(t, err)
	sameSite := playwright.SameSiteAttributeNone
	if isChromium || (isWebKit && runtime.GOOS == "linux") {
		sameSite = playwright.SameSiteAttributeLax
	}

	require.Equal(t, []playwright.Cookie{
		{
			Name:         "password",
			Value:        "123456",
			Domain:       "127.0.0.1",
			Path:         "/",
			Expires:      -1,
			HttpOnly:     false,
			Secure:       false,
			SameSite:     sameSite,
			PartitionKey: nil,
		},
	}, cookies)

	require.NoError(t, page.Context().ClearCookies())
	_, err = page.Reload()
	require.NoError(t, err)
	cookie, err = page.Evaluate("() => document.cookie")
	require.NoError(t, err)
	require.Equal(t, "", cookie)
}

func TestBrowserContextAddInitScript(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, context.AddInitScript(playwright.Script{
		Content: playwright.String(`window['injected'] = 123;`),
	}))
	_, err := page.Goto(server.PREFIX + "/tamperable.html")
	require.NoError(t, err)
	result, err := page.Evaluate(`() => window['result']`)
	require.NoError(t, err)
	require.Equal(t, 123, result)
}

func TestBrowserContextAddInitScriptWithPath(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, context.AddInitScript(playwright.Script{
		Path: playwright.String(Asset("injectedfile.js")),
	}))
	_, err := page.Goto(server.PREFIX + "/tamperable.html")
	require.NoError(t, err)
	result, err := page.Evaluate(`() => window['result']`)
	require.NoError(t, err)
	require.Equal(t, 123, result)
}

func TestBrowserContextWindowOpenshouldUseParentTabContext(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	popupEvent, err := page.ExpectEvent("popup", func() error {
		_, err := page.Evaluate("url => window.open(url)", server.EMPTY_PAGE)
		return err
	})
	require.NoError(t, err)
	popup := popupEvent.(playwright.Page)
	require.Equal(t, popup.Context(), context)
}

func TestBrowserContextUnrouteShouldWork(t *testing.T) {
	BeforeEach(t)

	intercepted := []int{}
	handler1 := func(route playwright.Route) {
		intercepted = append(intercepted, 1)
		require.NoError(t, route.Continue())
	}
	require.NoError(t, context.Route("**/*", handler1))
	require.NoError(t, context.Route("**/empty.html", func(route playwright.Route) {
		intercepted = append(intercepted, 2)
		require.NoError(t, route.Continue())
	}))
	require.NoError(t, context.Route("**/empty.html", func(route playwright.Route) {
		intercepted = append(intercepted, 3)
		require.NoError(t, route.Continue())
	}))

	handler4 := func(route playwright.Route) {
		intercepted = append(intercepted, 4)
		require.NoError(t, route.Continue())
	}
	require.NoError(t, context.Route(regexp.MustCompile("empty.html"), handler4))

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Equal(t, []int{4}, intercepted)

	intercepted = []int{}
	require.NoError(t, context.Unroute(regexp.MustCompile("empty.html"), handler4))
	_, err = page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Equal(t, []int{3}, intercepted)

	intercepted = []int{}
	require.NoError(t, context.Unroute("**/empty.html"))
	_, err = page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Equal(t, []int{1}, intercepted)
}

func TestBrowserContextShouldReturnBackgroundPage(t *testing.T) {
	// Background pages have been removed from Chromium together with Manifest V2 extensions
	t.Skip("Background pages are deprecated - Manifest V2 extensions no longer supported in Chromium")
}

func TestPageEventShouldHaveURL(t *testing.T) {
	BeforeEach(t)

	context.OnPage(func(p playwright.Page) {
		require.Equal(t, server.EMPTY_PAGE, p.URL())
	})
	newPage, err := context.ExpectPage(func() error {
		_, err := page.Evaluate("url => window.open(url)", server.EMPTY_PAGE)
		return err
	})
	require.NoError(t, err)
	require.Equal(t, server.EMPTY_PAGE, newPage.URL())
}

func TestConsoleEventShouldWork(t *testing.T) {
	BeforeEach(t)

	context.OnConsole(func(message playwright.ConsoleMessage) {
		require.Equal(t, "hello", message.Text())
	})
	message, err := context.ExpectConsoleMessage(func() error {
		_, err := page.Evaluate(`() => console.log("hello")`)
		return err
	})
	require.NoError(t, err)
	require.Equal(t, "hello", message.Text())
	require.Equal(t, page, message.Page())
}

func TestBrowserContextEventsRequest(t *testing.T) {
	BeforeEach(t)

	var requests []playwright.Request
	context.OnRequest(func(request playwright.Request) {
		requests = append(requests, request)
	})
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`<a target=_blank rel=noopener href="/one-style.html">yo</a>`))

	ret, err := context.ExpectEvent("page", func() error {
		return page.Locator("a").Click()
	})
	require.NoError(t, err)
	page1 := ret.(playwright.Page)
	require.NoError(t, page1.WaitForLoadState())
	require.Len(t, requests, 3)
	require.Equal(t, server.EMPTY_PAGE, requests[0].URL())
	require.Equal(t, server.PREFIX+"/one-style.html", requests[1].URL())
	require.Equal(t, server.PREFIX+"/one-style.css", requests[2].URL())
}

func TestBrowserContextEventsResponse(t *testing.T) {
	BeforeEach(t)

	var responses []playwright.Response
	context.OnResponse(func(response playwright.Response) {
		responses = append(responses, response)
	})
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`<a target=_blank rel=noopener href="/one-style.html">yo</a>`))

	ret, err := context.ExpectEvent("page", func() error {
		return page.Locator("a").Click()
	})
	require.NoError(t, err)
	page1 := ret.(playwright.Page)
	require.NoError(t, page1.WaitForLoadState())
	require.Len(t, responses, 3)
	require.Equal(t, server.EMPTY_PAGE, responses[0].URL())
	require.Equal(t, server.PREFIX+"/one-style.html", responses[1].URL())
	require.Equal(t, server.PREFIX+"/one-style.css", responses[2].URL())
}

func TestBrowserContextEventsRequestFailed(t *testing.T) {
	BeforeEach(t)

	server.SetRoute("/one-style.css", func(w http.ResponseWriter, r *http.Request) {
		hw, ok := w.(http.Hijacker)
		if ok {
			conn, _, err := hw.Hijack()
			if err == nil {
				conn.Close()
			}
		}
	})
	var failedRequests []playwright.Request
	context.OnRequestFailed(func(request playwright.Request) {
		failedRequests = append(failedRequests, request)
	})
	_, err := page.Goto(server.PREFIX + "/one-style.css")
	require.Error(t, err)
	require.Len(t, failedRequests, 1)
	require.Equal(t, server.PREFIX+"/one-style.css", failedRequests[0].URL())
}

func TestBrowerContextEventsShouldFireInProperOrder(t *testing.T) {
	BeforeEach(t)

	var events []string
	context.OnRequest(func(request playwright.Request) {
		events = append(events, "request")
	})
	context.OnResponse(func(response playwright.Response) {
		events = append(events, "response")
	})
	context.OnRequestFinished(func(request playwright.Request) {
		events = append(events, "requestfinished")
	})
	_, err := context.ExpectEvent("requestfinished", func() error {
		_, err := page.Goto(server.EMPTY_PAGE)
		return err
	})
	require.NoError(t, err)
	require.Equal(t, []string{"request", "response", "requestfinished"}, events)
}

func TestBrowserContextShouldFireCloseEvent(t *testing.T) {
	BeforeEach(t)

	browser1, err := browserType.Launch()
	require.NoError(t, err)
	defer browser1.Close()
	context1, err := browser1.NewContext()
	require.NoError(t, err)
	closed := false
	context1.OnClose(func(bc playwright.BrowserContext) {
		closed = true
	})
	require.NoError(t, browser1.Close())
	require.True(t, closed)
}

func TestDialogEventShouldWorkInImmdiatelyClosedPopup(t *testing.T) {
	BeforeEach(t)

	if isFirefox {
		t.Skip("flaky on firefox")
	}
	var popup playwright.Page
	page.OnPopup(func(p playwright.Page) {
		popup = p
	})
	msg, err := page.Context().ExpectConsoleMessage(func() error {
		_, err := page.Evaluate(`() => {
			const win = window.open();
			win.console.log('hello');
			win.close();
		}`)
		return err
	})
	require.NoError(t, err)
	require.Equal(t, "hello", msg.Text())
	require.Equal(t, popup, msg.Page())
}

func TestBrowserContextCloseShouldAbortWaitForEvent(t *testing.T) {
	BeforeEach(t)

	_, err := context.ExpectPage(func() error {
		return context.Close()
	})
	require.ErrorIs(t, err, playwright.ErrTargetClosed)
}

func TestBrowserContextCloseShouldBeCallableTwice(t *testing.T) {
	BeforeEach(t)

	countClose := atomic.Int32{}
	context.OnClose(func(bc playwright.BrowserContext) {
		countClose.Add(1)
	})
	require.NoError(t, context.Close())
	require.NoError(t, context.Close())
	require.NoError(t, context.Close())
	require.Equal(t, int32(1), countClose.Load())
}

func TestPageErrorEventShouldWork(t *testing.T) {
	BeforeEach(t)

	ret, err := page.Context().ExpectEvent("weberror", func() error {
		return page.SetContent(`<script>throw new Error("boom")</script>`)
	})
	require.NoError(t, err)
	require.NotNil(t, ret)
	weberror, ok := ret.(playwright.WebError)
	require.True(t, ok)
	require.Equal(t, page, weberror.Page())
	require.ErrorContains(t, weberror.Error(), "boom")
}

func TestBrowserContextOnResponse(t *testing.T) {
	BeforeEach(t)

	responseChan := make(chan playwright.Response, 1)
	context.OnResponse(func(response playwright.Response) {
		responseChan <- response
	})
	_, err := page.Goto(fmt.Sprintf("%s/title.html", server.PREFIX))
	require.NoError(t, err)
	response := <-responseChan
	body, err := response.Body()
	require.NoError(t, err)
	require.Equal(t, "<!DOCTYPE html>\n<title>Woof-Woof</title>\n", string(body))
}

func TestBrowserContextGetSecureCookies(t *testing.T) {
	BeforeEach(t, playwright.BrowserNewContextOptions{
		IgnoreHttpsErrors: playwright.Bool(true), // webkit requires https to support secure cookies
	})

	tlsServer := newTestServer(true)
	defer tlsServer.testServer.Close()

	tlsServer.SetRoute("/cookie.html", func(w http.ResponseWriter, r *http.Request) {
		// set secure cookie
		cookie := http.Cookie{
			Name:     "foo",
			Value:    "bar",
			Secure:   true,
			HttpOnly: true,
		}
		http.SetCookie(w, &cookie)
	})

	_, err := page.Goto(fmt.Sprintf("%s/cookie.html", tlsServer.PREFIX))
	require.NoError(t, err)
	cookies, err := context.Cookies()
	require.NoError(t, err)
	require.Equal(t, 1, len(cookies))
	require.Equal(t, "foo", cookies[0].Name)
	require.Equal(t, "bar", cookies[0].Value)
	require.True(t, cookies[0].Secure)
}

func TestBrowserContextClearCookies(t *testing.T) {
	t.Run("should remove cookies by name", func(t *testing.T) {
		BeforeEach(t)
		require.NoError(t, context.AddCookies([]playwright.OptionalCookie{
			{
				Name:   "cookie1",
				Value:  "1",
				Domain: mustGetHostname(server.PREFIX),
				Path:   playwright.String("/"),
			},
			{
				Name:   "cookie2",
				Value:  "2",
				Domain: mustGetHostname(server.PREFIX),
				Path:   playwright.String("/"),
			},
		}))

		_, err := page.Goto(server.PREFIX)
		require.NoError(t, err)
		expectPageCookies(t, "cookie1=1; cookie2=2")
		require.NoError(t, context.ClearCookies(playwright.BrowserContextClearCookiesOptions{
			Name: "cookie1",
		}))
		expectPageCookies(t, "cookie2=2")
	})

	t.Run("should remove cookies by name regex", func(t *testing.T) {
		BeforeEach(t)
		require.NoError(t, context.AddCookies([]playwright.OptionalCookie{
			{
				Name:   "cookie1",
				Value:  "1",
				Domain: mustGetHostname(server.PREFIX),
				Path:   playwright.String("/"),
			},
			{
				Name:   "cookie2",
				Value:  "2",
				Domain: mustGetHostname(server.PREFIX),
				Path:   playwright.String("/"),
			},
		}))

		_, err := page.Goto(server.PREFIX)
		require.NoError(t, err)
		expectPageCookies(t, "cookie1=1; cookie2=2")
		require.NoError(t, context.ClearCookies(playwright.BrowserContextClearCookiesOptions{
			Name: regexp.MustCompile(`coo.*1`),
		}))
		expectPageCookies(t, "cookie2=2")
	})

	t.Run("should remove cookies by domain", func(t *testing.T) {
		BeforeEach(t)
		require.NoError(t, context.AddCookies([]playwright.OptionalCookie{
			{
				Name:   "cookie1",
				Value:  "1",
				Domain: mustGetHostname(server.PREFIX),
				Path:   playwright.String("/"),
			},
			{
				Name:   "cookie2",
				Value:  "2",
				Domain: mustGetHostname(server.CROSS_PROCESS_PREFIX),
				Path:   playwright.String("/"),
			},
		}))

		_, err := page.Goto(server.PREFIX)
		require.NoError(t, err)
		expectPageCookies(t, "cookie1=1")

		_, err = page.Goto(server.CROSS_PROCESS_PREFIX)
		require.NoError(t, err)
		expectPageCookies(t, "cookie2=2")

		require.NoError(t, context.ClearCookies(playwright.BrowserContextClearCookiesOptions{
			Domain: mustGetHostname(server.CROSS_PROCESS_PREFIX),
		}))
		expectPageCookies(t, "")

		_, err = page.Goto(server.PREFIX)
		require.NoError(t, err)
		expectPageCookies(t, "cookie1=1")
	})

	t.Run("should remove cookies by path", func(t *testing.T) {
		BeforeEach(t)
		require.NoError(t, context.AddCookies([]playwright.OptionalCookie{
			{
				Name:   "cookie1",
				Value:  "1",
				Domain: mustGetHostname(server.PREFIX),
				Path:   playwright.String("/api/v1"),
			},
			{
				Name:   "cookie2",
				Value:  "2",
				Domain: mustGetHostname(server.PREFIX),
				Path:   playwright.String("/api/v2"),
			},
			{
				Name:   "cookie3",
				Value:  "3",
				Domain: mustGetHostname(server.PREFIX),
				Path:   playwright.String("/"),
			},
		}))

		_, err := page.Goto(fmt.Sprintf("%s/api/v1", server.PREFIX))
		require.NoError(t, err)

		ret, err := page.Evaluate(`document.cookie`)
		require.NoError(t, err)
		require.Equal(t, "cookie1=1; cookie3=3", ret)
		require.NoError(t, context.ClearCookies(playwright.BrowserContextClearCookiesOptions{
			Path: "/api/v1",
		}))
		expectPageCookies(t, "cookie3=3")

		_, err = page.Goto(fmt.Sprintf("%s/api/v2", server.PREFIX))
		require.NoError(t, err)
		expectPageCookies(t, "cookie2=2; cookie3=3")

		_, err = page.Goto(server.PREFIX)
		require.NoError(t, err)
		expectPageCookies(t, "cookie3=3")
	})

	t.Run("should remove cookies by name and domain", func(t *testing.T) {
		BeforeEach(t)
		require.NoError(t, context.AddCookies([]playwright.OptionalCookie{
			{
				Name:   "cookie1",
				Value:  "1",
				Domain: mustGetHostname(server.PREFIX),
				Path:   playwright.String("/"),
			},
			{
				Name:   "cookie1",
				Value:  "1",
				Domain: mustGetHostname(server.CROSS_PROCESS_PREFIX),
				Path:   playwright.String("/"),
			},
		}))

		_, err := page.Goto(server.PREFIX)
		require.NoError(t, err)
		expectPageCookies(t, "cookie1=1")
		require.NoError(t, context.ClearCookies(playwright.BrowserContextClearCookiesOptions{
			Name:   "cookie1",
			Domain: mustGetHostname(server.PREFIX),
		}))
		expectPageCookies(t, "")

		_, err = page.Goto(server.CROSS_PROCESS_PREFIX)
		require.NoError(t, err)
		expectPageCookies(t, "cookie1=1")
	})
}

func expectPageCookies(t *testing.T, cookie string) {
	t.Helper()
	ret, err := page.Evaluate(`document.cookie`)
	require.NoError(t, err)
	require.Equal(t, cookie, ret)
}

func TestBrowserContextShouldRetryECONNRESET(t *testing.T) {
	BeforeEach(t)

	requestCount := atomic.Int32{}
	server.SetRoute("/test", func(w http.ResponseWriter, r *http.Request) {
		if requestCount.Add(1) <= 3 {
			server.CloseClientConnections()
			return
		}
		w.Header().Add("Content-Type", "text/plain")
		_, _ = w.Write([]byte("Hello!"))
	})

	response, err := context.Request().Fetch(server.PREFIX+"/test", playwright.APIRequestContextFetchOptions{
		MaxRetries: playwright.Int(3),
	})
	require.NoError(t, err)
	require.Equal(t, 200, response.Status())
	body, err := response.Body()
	require.NoError(t, err)
	require.Equal(t, []byte("Hello!"), body)
	require.Equal(t, int32(4), requestCount.Load())
}

func TestBrowserContextShouldShowErrorAfterFulfill(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.Route("**/*", func(route playwright.Route) {
		require.NoError(t, route.Continue())
		panic("Exception text!?")
	}))

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	// Any next API call should throw because handler did throw during previous goto()
	_, err = page.Goto(server.EMPTY_PAGE)
	require.ErrorContains(t, err, "Exception text!?")
}
