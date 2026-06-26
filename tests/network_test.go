package playwright_test

import (
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestRequestFulfill(t *testing.T) {
	BeforeEach(t)

	requests := make(chan playwright.Request, 1)
	routes := make(chan playwright.Route, 1)
	err := page.Route("**/empty.html", func(route playwright.Route) {
		requests <- route.Request()
		routes <- route
		err := route.Fulfill(playwright.RouteFulfillOptions{
			Body: "Hello World",
		})
		require.NoError(t, err)
	})
	require.NoError(t, err)
	response, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, response.Finished())
	require.NoError(t, err)
	require.True(t, response.Ok())
	text, err := response.Text()
	require.NoError(t, err)
	require.Equal(t, "Hello World", text)
	request := <-requests
	route := <-routes
	headers, err := route.Request().AllHeaders()
	require.NoError(t, err)
	require.Contains(t, headers, "accept")
	require.Equal(t, route.Request(), request)
	require.Contains(t, request.URL(), "empty.html")
	require.Equal(t, "GET", request.Method())
	require.Contains(t, headers, "user-agent")
	postData, err := request.PostData()
	require.NoError(t, err)
	require.Equal(t, "", postData)
	require.True(t, request.IsNavigationRequest())
	require.Equal(t, request.ResourceType(), "document")
	require.Equal(t, request.Frame(), page.MainFrame())
}

func TestRequestContinue(t *testing.T) {
	BeforeEach(t)

	intercepted := make(chan bool, 1)
	err := page.Route("**/empty.html", func(route playwright.Route) {
		intercepted <- true
		err := route.Continue()
		require.NoError(t, err)
	})
	require.NoError(t, err)
	response, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, response.Finished())
	require.NoError(t, err)
	require.True(t, response.Ok())
	require.True(t, <-intercepted)
}

func TestRequestShouldFireForNavigationRequests(t *testing.T) {
	BeforeEach(t)

	requests := make(chan playwright.Request, 1)
	page.OnRequest(func(request playwright.Request) {
		requests <- request
	})
	response, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, response.Finished())
	require.True(t, response.Ok())
	<-requests
}

func TestShouldReportRequestHeadersArray(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	request, err := page.ExpectRequest("*/**", func() error {
		_, err := page.Evaluate(`
		() => fetch('/headers', {
            headers: [
                ['header-a', 'value-a'],
                ['header-a', 'value-a-1'],
                ['header-a', 'value-a-2'],
                ['header-b', 'value-b'],
            ]
            })
		`)
		return err
	})
	require.NoError(t, err)
	headers, err := request.AllHeaders()
	require.NoError(t, err)
	require.Contains(t, headers, "header-a")
	valueA, err := request.HeaderValue("header-a")
	require.NoError(t, err)
	require.Equal(t, "value-a, value-a-1, value-a-2", valueA)
	valueA, err = request.HeaderValue("not-there")
	require.NoError(t, err)
	require.Equal(t, "", valueA)
}

func TestShouldReportResponseHeadersArray(t *testing.T) {
	if isWebKit && runtime.GOOS == "windows" {
		t.Skip("libcurl does not support non-set-cookie multivalue headers")
	}
	BeforeEach(t)

	server.SetRoute("/headers", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("header-a", "value-a")
		rw.Header().Add("header-a", "value-a-1")
		rw.Header().Add("header-a", "value-a-2")
		rw.Header().Add("header-b", "value-b")
		rw.Header().Add("set-cookie", "a=b")
		rw.Header().Add("set-cookie", "c=d")
	})
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	response, err := page.ExpectResponse("*/**", func() error {
		_, err := page.Evaluate(`
		() => fetch('/headers')
		`)
		return err
	})
	require.NoError(t, err)
	actual_headers := make(map[string][]string)
	pw_headers, err := response.HeadersArray()
	require.NoError(t, err)
	for _, header := range pw_headers {
		if _, ok := actual_headers[strings.ToLower(header.Name)]; !ok {
			actual_headers[strings.ToLower(header.Name)] = make([]string, 0)
		}
		actual_headers[strings.ToLower(header.Name)] = append(actual_headers[strings.ToLower(header.Name)], header.Value)
	}
	for _, k := range []string{"Keep-Alive", "Connection", "Date", "Transfer-Encoding"} {
		delete(actual_headers, strings.ToLower(k))
	}
	require.Equal(t, actual_headers, map[string][]string{
		"header-a":       {"value-a", "value-a-1", "value-a-2"},
		"header-b":       {"value-b"},
		"set-cookie":     {"a=b", "c=d"},
		"content-length": {"0"},
	})
	header, err := response.HeaderValue("header-b")
	require.NoError(t, err)
	require.Equal(t, "value-b", header)
	header, err = response.HeaderValue("set-cookie")
	require.NoError(t, err)
	require.Equal(t, "a=b\nc=d", header)
	header, err = response.HeaderValue("not-there")
	require.NoError(t, err)
	require.Equal(t, "", header)
	headers, err := response.HeaderValues("header-a")
	require.NoError(t, err)
	sort.Strings(headers)
	require.Equal(t, []string{"value-a", "value-a-1", "value-a-2"}, headers)
	headers, err = response.HeaderValues("header-b")
	require.NoError(t, err)
	sort.Strings(headers)
	require.Equal(t, []string{"value-b"}, headers)
	headers, err = response.HeaderValues("set-cookie")
	require.NoError(t, err)
	sort.Strings(headers)
	require.Equal(t, []string{"a=b", "c=d"}, headers)
	headers, err = response.HeaderValues("not-there")
	require.NoError(t, err)
	require.Equal(t, []string{}, headers)
}

func TestShouldReportResponseServerAddr(t *testing.T) {
	BeforeEach(t)

	response, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	server_addr, err := response.ServerAddr()
	require.NoError(t, err)
	require.NotNil(t, server_addr)
	url, err := url.Parse(server.PREFIX)
	require.NoError(t, err)
	require.Equal(t, url.Port(), strconv.Itoa(server_addr.Port))
}

func TestShouldReportIfRequestWasFromServiceWorker(t *testing.T) {
	BeforeEach(t)

	response, err := page.Goto(fmt.Sprintf("%s/serviceworkers/fetch/sw.html", server.PREFIX))
	require.NoError(t, err)
	require.False(t, response.FromServiceWorker())

	_, err = page.Evaluate(`() => window.activationPromise`)
	require.NoError(t, err)
	response, err = page.ExpectResponse("**/example.txt", func() error {
		_, err := page.Evaluate(`() => fetch('/example.txt')`)
		return err
	})
	require.NoError(t, err)
	require.True(t, response.FromServiceWorker())
	if isChromium {
		require.NotZero(t, len(context.ServiceWorkers()))
	}
}

func TestNetworkEventsRequest(t *testing.T) {
	BeforeEach(t)

	requests := []playwright.Request{}
	page.OnRequest(func(request playwright.Request) {
		requests = append(requests, request)
	})
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Len(t, requests, 1)
	request := requests[0]
	require.Equal(t, "GET", request.Method())
	require.Equal(t, server.EMPTY_PAGE, request.URL())
	require.Equal(t, page.MainFrame(), request.Frame())
	_, err = request.Response()
	require.NoError(t, err)
}

func TestNetworkEventsResponse(t *testing.T) {
	BeforeEach(t)

	responses := []playwright.Response{}
	page.OnResponse(func(response playwright.Response) {
		responses = append(responses, response)
	})
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Len(t, responses, 1)
	response := responses[0]
	require.Equal(t, 200, response.Status())
	require.Equal(t, "OK", response.StatusText())
	require.Equal(t, page.MainFrame(), response.Frame())
}

func TestNetworkEventsShouldFireEventsInProperOrder(t *testing.T) {
	BeforeEach(t)

	events := []string{}
	page.OnRequest(func(request playwright.Request) {
		events = append(events, "request")
	})
	page.OnResponse(func(response playwright.Response) {
		events = append(events, "response")
	})
	page.OnRequestFinished(func(r playwright.Request) {
		events = append(events, "requestfinished")
	})
	response, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, response.Finished())
	require.Equal(t, []string{"request", "response", "requestfinished"}, events)
}

// TestBlockingServerCallInsideEventHandler reproduces the deadlock cluster
// (#391, #481, #398, #524, #574): making a blocking server round-trip from
// inside an event handler must not freeze the connection dispatcher.
func TestBlockingServerCallInsideEventHandler(t *testing.T) {
	BeforeEach(t)

	type result struct {
		headers []playwright.NameValue
		body    []byte
		err     error
	}
	done := make(chan result, 1)
	page.OnResponse(func(response playwright.Response) {
		// HeadersArray() and Body() each issue a blocking server call.
		// Before the fix these run on the dispatcher goroutine and deadlock.
		headers, err := response.HeadersArray()
		if err != nil {
			done <- result{err: err}
			return
		}
		body, err := response.Body()
		done <- result{headers: headers, body: body, err: err}
	})

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)

	select {
	case res := <-done:
		require.NoError(t, res.err)
		require.NotEmpty(t, res.headers)
	case <-time.After(10 * time.Second):
		t.Fatal("blocking server call inside an event handler deadlocked the dispatcher")
	}
}

// TestRequestHeadersArrayInsideRouteHandler reproduces #519: calling
// Request().HeadersArray() (or any header accessor) from inside a route handler
// must not stall the request. The accessor previously waited on Response(),
// which only resolves after the route is continued, deadlocking the handler.
func TestRequestHeadersArrayInsideRouteHandler(t *testing.T) {
	BeforeEach(t)

	headersCh := make(chan []playwright.NameValue, 1)
	err := page.Route("**/*", func(route playwright.Route) {
		headers, err := route.Request().HeadersArray()
		require.NoError(t, err)
		headersCh <- headers
		require.NoError(t, route.Continue())
	})
	require.NoError(t, err)

	done := make(chan error, 1)
	go func() {
		_, err := page.Goto(server.EMPTY_PAGE)
		done <- err
	}()

	select {
	case err := <-done:
		require.NoError(t, err)
		require.NotEmpty(t, <-headersCh)
	case <-time.After(15 * time.Second):
		t.Fatal("HeadersArray() inside a route handler deadlocked the request (#519)")
	}
}

func TestRequestExistingResponse(t *testing.T) {
	BeforeEach(t)

	response, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)

	// After the response arrives, the originating request exposes it
	// synchronously via ExistingResponse.
	existing, err := response.Request().ExistingResponse()
	require.NoError(t, err)
	require.NotNil(t, existing)
	require.Equal(t, response, existing)
}

func TestResponseTimingShouldPopulateTimingFromInitializer(t *testing.T) {
	BeforeEach(t)

	response, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, response.Finished())
	timing := response.Request().Timing()
	require.NotNil(t, timing)
	// StartTime defaults to 0 and is populated for real responses, while fields
	// the server does not report retain their -1 default (mirroring upstream).
	require.GreaterOrEqual(t, timing.StartTime, 0.0)
	require.GreaterOrEqual(t, timing.ResponseStart, timing.RequestStart)
}
