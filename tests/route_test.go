package playwright_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestPageRouteContinue(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetExtraHTTPHeaders(map[string]string{
		"extra-http": "42",
	}))
	intercepted := make(chan bool, 1)
	err := page.Route("**/empty.html", func(route playwright.Route) {
		request := route.Request()
		require.Contains(t, request.URL(), "empty.html")
		require.True(t, len(request.Headers()["user-agent"]) > 5)
		require.Equal(t, "42", request.Headers()["extra-http"])
		require.Equal(t, "GET", request.Method())

		postData, err := request.PostData()
		require.NoError(t, err)
		require.Equal(t, "", postData)
		require.True(t, request.IsNavigationRequest())
		require.Equal(t, "document", request.ResourceType())
		require.Equal(t, request.Frame(), page.MainFrame())
		require.Equal(t, "about:blank", request.Frame().URL())
		require.NoError(t, route.Continue())
		intercepted <- true
	})
	require.NoError(t, err)
	response, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.True(t, response.Ok())
	<-intercepted
}

func TestRouteContinueOverwrite(t *testing.T) {
	BeforeEach(t)

	serverRequestChan := server.WaitForRequestChan("/sleep.zzz")
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.Route("**/*", func(route playwright.Route) {
		headers := route.Request().Headers()
		headers["Foo"] = "bar"
		require.NoError(t, route.Continue(playwright.RouteContinueOptions{
			Method:   playwright.String("POST"),
			Headers:  headers,
			PostData: "foobar",
		}))
	}))
	_, err = page.Evaluate(`() => fetch("/sleep.zzz")`)
	require.NoError(t, err)
	serverRequest := <-serverRequestChan
	require.Equal(t, "POST", serverRequest.Method)
	require.Equal(t, "bar", serverRequest.Header.Get("Foo"))
	respData, err := io.ReadAll(serverRequest.Body)
	require.NoError(t, err)
	require.Equal(t, "foobar", string(respData))
}

func TestRouteContinueOverwriteBodyBytes(t *testing.T) {
	BeforeEach(t)

	serverRequestChan := server.WaitForRequestChan("/sleep.zzz")
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.Route("**/*", func(route playwright.Route) {
		require.NoError(t, route.Continue(playwright.RouteContinueOptions{
			Method:   playwright.String("POST"),
			PostData: []byte("foobar"),
		}))
	}))
	_, err = page.Evaluate(`() => fetch("/sleep.zzz")`)
	require.NoError(t, err)
	serverRequest := <-serverRequestChan
	require.Equal(t, "POST", serverRequest.Method)
	respData, err := io.ReadAll(serverRequest.Body)
	require.NoError(t, err)
	require.Equal(t, "foobar", string(respData))
}

func TestRouteFulfill(t *testing.T) {
	BeforeEach(t)

	requestsChan := make(chan playwright.Request, 1)
	err := page.Route("**/empty.html", func(route playwright.Route) {
		request := route.Request()
		require.Contains(t, request.URL(), "empty.html")
		require.True(t, len(request.Headers()["user-agent"]) > 5)
		require.Equal(t, "GET", request.Method())

		postData, err := request.PostData()
		require.NoError(t, err)
		require.Equal(t, "", postData)
		require.True(t, request.IsNavigationRequest())
		require.Equal(t, "document", request.ResourceType())
		require.Equal(t, page, page.MainFrame().Page())
		require.Equal(t, request.Frame(), page.MainFrame())
		require.Equal(t, "about:blank", request.Frame().URL())
		require.NoError(t, route.Fulfill(playwright.RouteFulfillOptions{
			Body:        "123",
			ContentType: playwright.String("text/plain"),
			Headers: map[string]string{
				"Foo": "bar",
			},
		}))
		requestsChan <- request
	})
	require.NoError(t, err)
	response, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.True(t, response.Ok())
	request := <-requestsChan
	require.Equal(t, request, response.Request())
	text, err := response.Text()
	require.NoError(t, err)
	require.Equal(t, "123", text)
	require.Equal(t, "bar", response.Headers()["foo"])
	require.Equal(t, "text/plain", response.Headers()["content-type"])
}

func TestRouteFulfillByteSlice(t *testing.T) {
	BeforeEach(t)

	intercepted := make(chan bool, 1)
	err := page.Route("**/empty.html", func(route playwright.Route) {
		require.NoError(t, route.Fulfill(playwright.RouteFulfillOptions{
			Body:        []byte("123"),
			ContentType: playwright.String("text/plain"),
		}))
		intercepted <- true
	})
	require.NoError(t, err)
	response, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.True(t, response.Ok())
	<-intercepted
	text, err := response.Text()
	require.NoError(t, err)
	require.Equal(t, "123", text)
	require.Equal(t, "3", response.Headers()["content-length"])
	require.Equal(t, "text/plain", response.Headers()["content-type"])
}

func TestRouteFulfillPath(t *testing.T) {
	BeforeEach(t)

	intercepted := make(chan bool, 1)
	err := page.Route("**/empty.html", func(route playwright.Route) {
		require.NoError(t, route.Fulfill(playwright.RouteFulfillOptions{
			Path: playwright.String(Asset("pptr.png")),
		}))
		intercepted <- true
	})
	require.NoError(t, err)
	response, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.True(t, response.Ok())
	<-intercepted
	body, err := response.Body()
	require.NoError(t, err)
	require.True(t, len(body) > 5000)
	require.Equal(t, "image/png", response.Headers()["content-type"])
}

// TestRouteFulfillPathContentTypeFromExtension verifies the content type is
// derived from the file extension (matching upstream getMimeTypeForPath) rather
// than from content sniffing, which would report text/plain for a .css file.
func TestRouteFulfillPathContentTypeFromExtension(t *testing.T) {
	BeforeEach(t)

	intercepted := make(chan bool, 1)
	err := page.Route("**/empty.html", func(route playwright.Route) {
		require.NoError(t, route.Fulfill(playwright.RouteFulfillOptions{
			Path: playwright.String(Asset("one-style.css")),
		}))
		intercepted <- true
	})
	require.NoError(t, err)
	response, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.True(t, response.Ok())
	<-intercepted
	require.Equal(t, "text/css", response.Headers()["content-type"])
}

func TestRequestFinished(t *testing.T) {
	BeforeEach(t)

	eventsStorage := newSyncSlice[string]()
	var request playwright.Request
	page.Once("request", func(r playwright.Request) {
		request = r
		eventsStorage.Append("request")
	})
	page.Once("response", func() {
		eventsStorage.Append("response")
	})
	response, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, response.Finished())
	eventsStorage.Append("requestfinished")
	require.Equal(t, []string{"request", "response", "requestfinished"}, eventsStorage.Get())
	require.Equal(t, response.Request(), request)
	require.Equal(t, response.Frame(), page.MainFrame())
}

func TestResponsePostData(t *testing.T) {
	BeforeEach(t)

	requestData := map[string]interface{}{
		"foo": "bar123",
		"kek": true,
	}
	server.SetRoute("/foobar", func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewEncoder(w).Encode(requestData))
	})
	response, err := page.Goto(server.PREFIX + "/foobar")
	require.NoError(t, err)
	var actualResponse map[string]interface{}
	require.NoError(t, response.JSON(&actualResponse))
	require.Equal(t, requestData, actualResponse)
}

func TestRouteAbort(t *testing.T) {
	BeforeEach(t)

	failedRequests := make(chan playwright.Request, 1)
	page.Once("requestfailed", func(request playwright.Request) {
		failedRequests <- request
	})
	err := page.Route("**/empty.html", func(route playwright.Route) {
		require.NoError(t, route.Abort("aborted"))
	})
	require.NoError(t, err)
	_, err = page.Goto(server.EMPTY_PAGE)
	require.Error(t, err)
	request := <-failedRequests
	require.True(t, len(request.Failure().Error()) > 5)
}

func TestRequestPostData(t *testing.T) {
	BeforeEach(t)

	server.SetRoute("/foobar", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.Route("**/foobar", func(route playwright.Route) {
		var postData map[string]interface{}
		request := route.Request()
		require.NoError(t, request.PostDataJSON(&postData))
		require.Equal(t, map[string]interface{}{
			"foo": true,
			"kek": float64(123),
		}, postData)
		raw, err := request.PostDataBuffer()
		require.NoError(t, err)
		require.Equal(t, []byte(`{"foo":true,"kek":123}`), raw)
		require.NoError(t, route.Continue())
	}))
	_, err = page.Evaluate(`url => fetch(url, {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify({
			"foo": true,
			"kek": 123,
		})
	})`, server.PREFIX+"/foobar")
	require.NoError(t, err)
}

// TestRequestPostDataJSONFormURLEncoded mirrors upstream's "should parse the
// data if content-type is application/x-www-form-urlencoded" test
// (tests/page/page-network-request.spec.ts).
func TestRequestPostDataJSONFormURLEncoded(t *testing.T) {
	BeforeEach(t)

	server.SetRoute("/post", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	requestChan := make(chan playwright.Request, 1)
	page.OnRequest(func(r playwright.Request) {
		requestChan <- r
	})
	require.NoError(t, page.SetContent(`<form method='POST' action='/post'><input type='text' name='foo' value='bar'><input type='number' name='baz' value='123'><input type='submit'></form>`))
	require.NoError(t, page.Locator("input[type=submit]").Click())
	request := <-requestChan
	var postData map[string]interface{}
	require.NoError(t, request.PostDataJSON(&postData))
	require.Equal(t, map[string]interface{}{
		"foo": "bar",
		"baz": "123",
	}, postData)
}

// TestRequestPostDataJSONFormURLEncodedDuplicateKey verifies that a repeated
// form key keeps the LAST value, matching upstream's URLSearchParams.entries()
// (last write wins) rather than url.Values.Get (first value).
func TestRequestPostDataJSONFormURLEncodedDuplicateKey(t *testing.T) {
	BeforeEach(t)

	server.SetRoute("/post", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	requestChan := make(chan playwright.Request, 1)
	page.OnRequest(func(r playwright.Request) {
		requestChan <- r
	})
	_, err = page.Evaluate(`url => fetch(url, {
		method: 'POST',
		headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
		body: 'foo=bar&foo=baz',
	})`, server.PREFIX+"/post")
	require.NoError(t, err)
	request := <-requestChan
	var postData map[string]interface{}
	require.NoError(t, request.PostDataJSON(&postData))
	require.Equal(t, map[string]interface{}{"foo": "baz"}, postData)
}

func TestFulfillWithURLOverride(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.Route("**/*.html", func(route playwright.Route) {
		response, err := route.Fetch(playwright.RouteFetchOptions{
			URL: playwright.String(server.PREFIX + "/one-style.html"),
		})
		require.NoError(t, err)
		headers := response.Headers() // require for webkit on linux
		headers["foo"] = "bar"
		require.NoError(t, route.Fulfill(playwright.RouteFulfillOptions{
			Response: response,
			Headers:  headers,
		}))
	}))
	response, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.True(t, response.Ok())
	body, err := response.Text()
	require.NoError(t, err)
	require.Contains(t, body, "one-style.css")
	headers, err := response.AllHeaders()
	require.NoError(t, err)
	require.Equal(t, "bar", headers["foo"])
}

func TestResponseSecurityDetails(t *testing.T) {
	BeforeEach(t)

	if isWebKit {
		t.Skip("https://github.com/microsoft/playwright/issues/6759")
	}
	tlsServer := newTestServer(true)
	defer tlsServer.testServer.Close()
	page2, err := browser.NewPage(playwright.BrowserNewPageOptions{
		IgnoreHttpsErrors: playwright.Bool(true),
	})
	require.NoError(t, err)
	response, err := page2.Goto(tlsServer.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, response.Finished())
	securityDetails, err := response.SecurityDetails()
	require.NoError(t, err)

	require.Equal(t, "TLS 1.3", *securityDetails.Protocol)
	require.NoError(t, page2.Close())
}

func TestResponseServerAddrReturnNilForFulfilledRoute(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.Route("**/*", func(route playwright.Route) {
		require.NoError(t, route.Fulfill(playwright.RouteFulfillOptions{
			Status:      playwright.Int(200),
			ContentType: playwright.String("text/html"),
			Body:        "<h1>fulfilled</h1>",
		}))
	}))
	response, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, response.Finished())
	// A fulfilled route has no real server, so Playwright returns null for the
	// server address. This must not panic (https://github.com/mxschmitt/playwright-go/issues/543).
	serverAddr, err := response.ServerAddr()
	require.NoError(t, err)
	require.Nil(t, serverAddr)
}

func TestRequestTimingShouldWork(t *testing.T) {
	BeforeEach(t)

	result, err := page.ExpectEvent("requestfinished", func() error {
		_, err := page.Goto(server.EMPTY_PAGE)
		return err
	})
	require.NoError(t, err)
	request, ok := result.(playwright.Request)
	require.True(t, ok)
	timing := request.Timing()
	require.GreaterOrEqual(t, timing.RequestStart, timing.ConnectEnd)
	require.GreaterOrEqual(t, timing.ResponseStart, timing.RequestStart)
	require.GreaterOrEqual(t, timing.ResponseEnd, timing.ResponseStart)
	require.Less(t, timing.ResponseEnd, 10000.0)
}

func TestShouldInterceptByGlob(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.Route("http://localhos**?*oo", func(route playwright.Route) {
		require.NoError(t, route.Fulfill(playwright.RouteFulfillOptions{
			Body:   "intercepted",
			Status: playwright.Int(200),
		}))
	}))

	ret, err := page.Evaluate(`url => fetch(url).then(r => r.text())`, fmt.Sprintf("%s/?foo", server.CROSS_PROCESS_PREFIX))
	require.NoError(t, err)
	require.Equal(t, "intercepted", ret)
}
