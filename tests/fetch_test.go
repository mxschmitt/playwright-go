package playwright_test

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestFetchShouldWork(t *testing.T) {
	BeforeEach(t)

	request, err := pw.Request.NewContext()
	require.NoError(t, err)
	url := server.PREFIX + "/simple.json"
	check := func(method string, response playwright.APIResponse) {
		require.NoError(t, err)
		require.Equal(t, 200, response.Status())
		require.Equal(t, "OK", response.StatusText())
		require.Equal(t, url, response.URL())
		require.Equal(t, "application/json", response.Headers()["content-type"])
		text, err := response.Text()
		require.NoError(t, err)
		if method == "HEAD" {
			require.Equal(t, "", text)
		} else {
			require.Equal(t, "{\"foo\": \"bar\"}\n", text)
		}
	}

	response, err := request.Fetch(url, playwright.APIRequestContextFetchOptions{
		Timeout: playwright.Float(500),
	})
	check("GET", response)
	response, err = request.Head(url, playwright.APIRequestContextHeadOptions{
		Timeout: playwright.Float(500),
	})
	check("HEAD", response)
	response, err = request.Post(url, playwright.APIRequestContextPostOptions{
		Timeout: playwright.Float(500),
	})
	check("PATCH", response)
	response, err = request.Patch(url, playwright.APIRequestContextPatchOptions{
		Timeout: playwright.Float(500),
	})
	check("POST", response)
	response, err = request.Put(url, playwright.APIRequestContextPutOptions{
		Timeout: playwright.Float(500),
	})
	check("PUT", response)
	response, err = request.Delete(url, playwright.APIRequestContextDeleteOptions{
		Timeout: playwright.Float(500),
	})
	check("DELETE", response)
}

func TestShouldDisposeGlobalRequest(t *testing.T) {
	BeforeEach(t)

	request, err := pw.Request.NewContext()
	require.NoError(t, err)
	response, err := request.Get(server.PREFIX + "/simple.json")
	require.NoError(t, err)
	var data map[string]string
	require.NoError(t, response.JSON(&data))
	require.Equal(t, map[string]string{"foo": "bar"}, data)
	require.NoError(t, request.Dispose())
	_, err = response.Body()
	require.Error(t, err, "response has been disposed")
}

func TestShouldDisposeWithCustomErrorMessage(t *testing.T) {
	BeforeEach(t)

	request, err := pw.Request.NewContext()
	require.NoError(t, err)
	require.NoError(t, request.Dispose(playwright.APIRequestContextDisposeOptions{
		Reason: playwright.String("My reason"),
	}))
	_, err = request.Get(server.EMPTY_PAGE)
	require.ErrorContains(t, err, "My reason")
}

func TestShouldWorkAfterContextDisposed(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, context.Close(playwright.BrowserContextCloseOptions{
		Reason: playwright.String("Test ended."),
	}))

	_, err := context.Request().Get(server.EMPTY_PAGE)
	require.ErrorContains(t, err, "Test ended.")
}

func TestShouldSupportGlobalUserAgentOption(t *testing.T) {
	BeforeEach(t)

	request, err := pw.Request.NewContext(playwright.APIRequestNewContextOptions{
		UserAgent: playwright.String("My Agent"),
	})
	require.NoError(t, err)
	chanRes := server.WaitForRequestChan("/empty.html")
	response, err := request.Get(server.PREFIX + "/empty.html")
	require.NoError(t, err)
	res := <-chanRes
	require.Equal(t, "My Agent", res.UserAgent())
	require.Equal(t, 200, response.Status())
}

func TestShoulSupportGlobalTimeoutOption(t *testing.T) {
	BeforeEach(t)

	request, err := pw.Request.NewContext(playwright.APIRequestNewContextOptions{
		Timeout: playwright.Float(100),
	})
	require.NoError(t, err)
	server.SetRoute("/empty.html", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
	})
	_, err = request.Get(server.PREFIX + "/empty.html")
	require.ErrorContains(t, err, `Timeout 100ms exceeded`)
}

func TestShouldPropagateExtraHttpHeadersWithRedirects(t *testing.T) {
	BeforeEach(t)

	server.SetRedirect("/a/redirect1", "/b/c/redirect2")
	server.SetRedirect("/b/c/redirect2", "/simple.json")
	request, err := pw.Request.NewContext(playwright.APIRequestNewContextOptions{
		ExtraHttpHeaders: map[string]string{
			"My-Secret": "Value",
		},
	})
	require.NoError(t, err)
	wg := &sync.WaitGroup{}
	for _, url := range []string{"/a/redirect1", "/b/c/redirect2", "/simple.json"} {
		wg.Add(1)
		chanRes := server.WaitForRequestChan(url)
		require.NoError(t, err)
		go func() {
			defer wg.Done()
			res := <-chanRes
			require.Equal(t, "Value", res.Header.Get("My-Secret"))
		}()
	}
	response, err := request.Get(server.PREFIX + "/a/redirect1")
	require.NoError(t, err)
	require.Equal(t, 200, response.Status())
	wg.Wait()
	require.NoError(t, request.Dispose())
}

func TestShouldSupportGlobalHttpCredentialsOption(t *testing.T) {
	BeforeEach(t)

	server.SetBasicAuth("/empty.html", "user", "pass")
	request, err := pw.Request.NewContext()
	require.NoError(t, err)
	response, err := request.Get(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Equal(t, 401, response.Status())
	require.NoError(t, request.Dispose())

	request1, err := pw.Request.NewContext(playwright.APIRequestNewContextOptions{
		HttpCredentials: &playwright.HttpCredentials{
			Username: "user",
			Password: "pass",
		},
	})
	require.NoError(t, err)
	response1, err := request1.Get(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Equal(t, 200, response1.Status())
	require.NoError(t, request1.Dispose())

	request2, err := pw.Request.NewContext(playwright.APIRequestNewContextOptions{
		HttpCredentials: &playwright.HttpCredentials{
			Username: "user",
			Password: "wrong",
		},
	})
	require.NoError(t, err)
	response2, err := request2.Get(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Equal(t, 401, response2.Status())
	require.NoError(t, request2.Dispose())
}

func TestShouldSupportGlobalIgnoreHttpsErrorsOption(t *testing.T) {
	BeforeEach(t)

	tlsServer := newTestServer(true)
	defer tlsServer.testServer.Close()
	request, err := pw.Request.NewContext(playwright.APIRequestNewContextOptions{
		IgnoreHttpsErrors: playwright.Bool(true),
	})
	require.NoError(t, err)
	response, err := request.Get(tlsServer.EMPTY_PAGE)
	require.NoError(t, err)
	require.Equal(t, 200, response.Status())
	require.Equal(t, tlsServer.EMPTY_PAGE, response.URL())
	require.NoError(t, request.Dispose())
}

func TestShouldResoleUrlRelativeToGlobalBaseUrlOption(t *testing.T) {
	BeforeEach(t)

	request, err := pw.Request.NewContext(playwright.APIRequestNewContextOptions{
		BaseURL: playwright.String(server.PREFIX),
	})
	require.NoError(t, err)
	response, err := request.Get("/empty.html")
	require.NoError(t, err)
	require.Equal(t, 200, response.Status())
	require.Equal(t, server.EMPTY_PAGE, response.URL())
	require.NoError(t, request.Dispose())
}

func TestShouldUsePlaywrightAsAUserAgent(t *testing.T) {
	BeforeEach(t)

	request, err := pw.Request.NewContext()
	require.NoError(t, err)
	resChan := server.WaitForRequestChan("/empty.html")
	_, err = request.Get(server.PREFIX + "/empty.html")
	require.NoError(t, err)
	res := <-resChan
	require.True(t, strings.HasPrefix(res.Header.Get("User-Agent"), "Playwright/"))
	require.NoError(t, request.Dispose())
}

func TestShouldReturnEmptyBody(t *testing.T) {
	BeforeEach(t)

	request, err := pw.Request.NewContext()
	require.NoError(t, err)
	response, err := request.Get(server.PREFIX + "/empty.html")
	require.NoError(t, err)
	body, err := response.Body()
	require.NoError(t, err)
	require.Equal(t, 0, len(body))
	text, err := response.Text()
	require.NoError(t, err)
	require.Equal(t, "", text)
	require.NoError(t, response.Dispose())
	_, err = response.Body()
	require.Error(t, err, "response has been disposed")
}

func TestStorageStateShouldRoundTripThroughFile(t *testing.T) {
	BeforeEach(t)

	storageState := &playwright.StorageState{
		Cookies: []playwright.Cookie{
			{
				Name:     "a",
				Value:    "b",
				Domain:   "a.b.one.com",
				Path:     "/",
				Secure:   false,
				HttpOnly: false,
				SameSite: playwright.SameSiteAttributeLax,
				Expires:  -1,
			},
		},
		Origins: []playwright.Origin{},
	}
	request, err := pw.Request.NewContext(playwright.APIRequestNewContextOptions{
		StorageState: storageState,
	})
	require.NoError(t, err)
	tempfile := filepath.Join(t.TempDir(), "storage-state.json")
	actual, err := request.StorageState(tempfile)
	require.NoError(t, err)
	require.Equal(t, storageState, actual)
	stateWritten, err := os.ReadFile(tempfile)
	require.NoError(t, err)
	var state *playwright.StorageState
	err = json.Unmarshal(stateWritten, &state)
	require.NoError(t, err)
	require.Equal(t, state, storageState)

	request2, err := pw.Request.NewContext(playwright.APIRequestNewContextOptions{
		StorageStatePath: playwright.String(tempfile),
	})
	require.NoError(t, err)
	actual2, err := request2.StorageState()
	require.NoError(t, err)
	require.Equal(t, storageState, actual2)
}

func TestShouldJsonStringifyBodyWhenContentTypeIsApplicationJson(t *testing.T) {
	BeforeEach(t)

	request, err := pw.Request.NewContext()
	require.NoError(t, err)
	serializationData := []interface{}{
		[]interface{}{map[string]string{"foo": "bar"}},
		[]interface{}{[]interface{}{"foo", "bar", 2021}},
		[]interface{}{"foo"},
		[]interface{}{true},
		[]interface{}{2021},
	}
	stringifiedValue, _ := json.Marshal(serializationData)
	server.SetRoute("/empty.html", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		_, err = w.Write(body)
		require.NoError(t, err)
	})
	response, err := request.Post(server.EMPTY_PAGE, playwright.APIRequestContextPostOptions{
		Headers: map[string]string{
			"content-type": "application/json",
		},
		Data: string(stringifiedValue),
	})
	require.NoError(t, err)
	body, err := response.Body()
	require.NoError(t, err)
	require.Equal(t, stringifiedValue, body)
	require.NoError(t, request.Dispose())
}

func TestShouldAcceptAlreadySerializedDataAsBytesWhenContentTypeIsApplicationJson(t *testing.T) {
	BeforeEach(t)

	request, err := pw.Request.NewContext()
	require.NoError(t, err)
	stringifiedValue, _ := json.Marshal(map[string]string{"foo": "bar"})
	server.SetRoute("/empty.html", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		_, err = w.Write(body)
		require.NoError(t, err)
	})
	response, err := request.Post(server.EMPTY_PAGE, playwright.APIRequestContextPostOptions{
		Headers: map[string]string{
			"content-type": "application/json",
		},
		Data: stringifiedValue,
	})
	require.NoError(t, err)
	body, err := response.Body()
	require.NoError(t, err)
	require.Equal(t, stringifiedValue, body)
	require.NoError(t, request.Dispose())
}

func TestShouldErrorWhenMaxRedirectsIsExceeded(t *testing.T) {
	BeforeEach(t)

	request, err := pw.Request.NewContext()
	require.NoError(t, err)
	server.SetRedirect("/a/redirect1", "/b/c/redirect2")
	server.SetRedirect("/b/c/redirect2", "/b/c/redirect3")
	server.SetRedirect("/b/c/redirect3", "/b/c/redirect4")
	server.SetRedirect("/b/c/redirect4", "/simple.json")
	for _, method := range []string{"GET", "PUT", "POST", "OPTIONS", "HEAD", "PATCH"} {
		for i := 1; i < 4; i++ {
			_, err := request.Fetch(server.PREFIX+"/a/redirect1", playwright.APIRequestContextFetchOptions{
				MaxRedirects: playwright.Int(i),
				Method:       playwright.String(method),
			})
			require.Error(t, err, "Max redirect count exceeded")
		}
	}
}

func TestShouldNotFollowRedirectsWhenMaxRedirectsIsZero(t *testing.T) {
	BeforeEach(t)

	request, err := pw.Request.NewContext()
	require.NoError(t, err)
	server.SetRedirect("/a/redirect1", "/b/c/redirect2")
	server.SetRedirect("/b/c/redirect2", "/simple.json")
	for _, method := range []string{"GET", "PUT", "POST", "OPTIONS", "HEAD", "PATCH"} {
		response, err := request.Fetch(server.PREFIX+"/a/redirect1", playwright.APIRequestContextFetchOptions{
			MaxRedirects: playwright.Int(0),
			Method:       playwright.String(method),
		})
		require.NoError(t, err)
		require.Equal(t, 302, response.Status())
		require.Equal(t, "/b/c/redirect2", response.Headers()["location"])
	}
}

func TestErrorWhenMaxRedirectsIsLessThanZero(t *testing.T) {
	BeforeEach(t)

	request, err := pw.Request.NewContext()
	require.NoError(t, err)
	for _, method := range []string{"GET", "PUT", "POST", "OPTIONS", "HEAD", "PATCH"} {
		_, err := request.Fetch(server.PREFIX+"/simple.json", playwright.APIRequestContextFetchOptions{
			MaxRedirects: playwright.Int(-1),
			Method:       playwright.String(method),
		})
		require.Error(t, err, "maxRedirects must be non-negative")
	}
}

func TestShouldSerializeNullValuesInJson(t *testing.T) {
	BeforeEach(t)

	request, err := pw.Request.NewContext()
	require.NoError(t, err)
	server.SetRoute("/echo", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		_, err = w.Write(body)
		require.NoError(t, err)
	})
	response, err := request.Post(server.PREFIX+"/echo", playwright.APIRequestContextPostOptions{
		Data: []byte("{\"foo\": null}"),
	})
	require.NoError(t, err)
	text, err := response.Text()
	require.NoError(t, err)
	require.Equal(t, `{"foo": null}`, text)
	require.NoError(t, request.Dispose())
}

func TestShouldSupportApplicationXWwwFormUrlencoded(t *testing.T) {
	BeforeEach(t)

	server.SetRoute("/empty.html", func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.Equal(t, "POST", r.Method)
		require.Contains(t, r.Header.Get("content-type"), "x-www-form-urlencoded")
		w.WriteHeader(200)
	})

	_, err := context.Request().Post(server.EMPTY_PAGE, playwright.APIRequestContextPostOptions{
		Form: map[string]interface{}{
			"firstName": "John",
			"lastName":  "Doe",
		},
	})
	require.NoError(t, err)
}

func TestShouldSupportMultipartFormData(t *testing.T) {
	BeforeEach(t)

	server.SetRoute("/empty.html", func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.Equal(t, "POST", r.Method)
		require.Contains(t, r.Header.Get("content-type"), "multipart")
		w.WriteHeader(200)
	})

	_, err := context.Request().Post(server.EMPTY_PAGE, playwright.APIRequestContextPostOptions{
		Multipart: map[string]interface{}{
			"firstName": "John",
			"lastName":  "Doe",
			"file": playwright.InputFile{
				Name:     "f.js",
				MimeType: "text/javascript",
				Buffer:   []byte("var x = 10;\r\n;console.log(x);"),
			},
		},
	})
	require.NoError(t, err)
}

func TestShouldSupportHttpCredentialsSendImmediatelyForBrowserContext(t *testing.T) {
	BeforeEach(t, playwright.BrowserNewContextOptions{
		HttpCredentials: &playwright.HttpCredentials{
			Username: "user",
			Password: "pass",
			Origin:   playwright.String(strings.ToUpper(server.PREFIX)),
			Send:     playwright.HttpCredentialsSendAlways,
		},
	})

	exptectedAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("user:pass"))

	// First request
	chanRes := server.WaitForRequestChan("/empty.html")
	response, err := context.Request().Get(server.PREFIX + "/empty.html")
	require.NoError(t, err)
	res := <-chanRes
	require.Equal(t, exptectedAuth, res.Header.Get("authorization"))
	require.Equal(t, 200, response.Status())

	// Second request
	response, err = context.Request().Get(server.CROSS_PROCESS_PREFIX + "/empty.html")
	require.NoError(t, err)
	res = <-chanRes
	// Not sent to another origin.
	require.Equal(t, "", res.Header.Get("authorization"))
	require.Equal(t, 200, response.Status())
}

func TestSupportHttpCredentialsSendImmediatelyForBrowserNewPage(t *testing.T) {
	BeforeEach(t)

	page1, err := browser.NewPage(playwright.BrowserNewPageOptions{
		HttpCredentials: &playwright.HttpCredentials{
			Username: "user",
			Password: "pass",
			Origin:   playwright.String(strings.ToUpper(server.PREFIX)),
			Send:     playwright.HttpCredentialsSendAlways,
		},
	})
	require.NoError(t, err)
	chanRes := server.WaitForRequestChan("/empty.html")
	response, err := page1.Request().Get(server.PREFIX + "/empty.html")
	require.NoError(t, err)
	res := <-chanRes
	require.Equal(t, "Basic "+base64.StdEncoding.EncodeToString([]byte("user:pass")), res.Header.Get("authorization"))
	require.Equal(t, 200, response.Status())

	response, err = page1.Request().Get(server.CROSS_PROCESS_PREFIX + "/empty.html")
	require.NoError(t, err)
	res = <-chanRes
	// Not sent to another origin.
	require.Equal(t, "", res.Header.Get("authorization"))
	require.Equal(t, 200, response.Status())

	require.NoError(t, page1.Close())
}

func TestFetchShouldRetryECONNRESET(t *testing.T) {
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

	request, err := pw.Request.NewContext()
	require.NoError(t, err)
	response, err := request.Get(server.PREFIX+"/test", playwright.APIRequestContextGetOptions{
		MaxRetries: playwright.Int(3),
	})
	require.NoError(t, err)
	require.Equal(t, 200, response.Status())
	body, err := response.Body()
	require.NoError(t, err)
	require.Equal(t, []byte("Hello!"), body)
	require.Equal(t, int32(4), requestCount.Load())
	require.NoError(t, request.Dispose())
}

func TestFetchShouldThrowWhenFailOnStatusCodeIsTrue(t *testing.T) {
	BeforeEach(t)

	server.SetRoute("/empty.html", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(404)
	})

	req, err := pw.Request.NewContext(playwright.APIRequestNewContextOptions{
		FailOnStatusCode: playwright.Bool(true),
	})
	require.NoError(t, err)

	_, err = req.Fetch(server.EMPTY_PAGE)
	require.ErrorContains(t, err, "404 Not Found")

	require.NoError(t, req.Dispose())
}

func TestFetchShouldNotThrowWhenFailOnStatusCodeIsFalse(t *testing.T) {
	BeforeEach(t)

	server.SetRoute("/empty.html", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(404)
	})

	req, err := pw.Request.NewContext(playwright.APIRequestNewContextOptions{
		FailOnStatusCode: playwright.Bool(false),
	})
	require.NoError(t, err)

	resp, err := req.Fetch(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Equal(t, 404, resp.Status())
	require.NoError(t, req.Dispose())
}

func TestShouldFollowMaxRedirects(t *testing.T) {
	BeforeEach(t)

	redirectCount := atomic.Int32{}
	server.SetRoute("/empty.html", func(w http.ResponseWriter, r *http.Request) {
		redirectCount.Add(1)
		w.Header().Add("Location", server.EMPTY_PAGE)
		w.WriteHeader(301)
	})

	request, err := pw.Request.NewContext(playwright.APIRequestNewContextOptions{
		MaxRedirects: playwright.Int(1),
	})
	require.NoError(t, err)
	_, err = request.Fetch(server.EMPTY_PAGE)
	require.ErrorContains(t, err, "Max redirect count exceeded")
	require.Equal(t, int32(2), redirectCount.Load())
	require.NoError(t, request.Dispose())
}
