package playwright_test

import (
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/mxschmitt/playwright-go"
)

func cookieNames(cookies []playwright.Cookie) []string {
	names := make([]string, 0, len(cookies))
	for _, cookie := range cookies {
		names = append(names, cookie.Name)
	}
	slices.Sort(names)
	return names
}

func TestPageRequestAddCookiesShouldAddCookiesToTheBrowserContext(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.Request().AddCookies([]playwright.OptionalCookie{
		{Name: "a", Value: "b", URL: playwright.String(server.EMPTY_PAGE)},
		{
			Name:    "c",
			Value:   "d",
			Domain:  playwright.String("127.0.0.1"),
			Path:    playwright.String("/"),
			Expires: playwright.Float(float64(time.Now().Add(time.Hour).Unix())),
		},
	}))
	cookies, err := context.Cookies()
	require.NoError(t, err)
	require.Equal(t, []string{"a", "c"}, cookieNames(cookies))

	requests := server.WaitForRequestChan("/empty.html")
	_, err = context.Request().Get(server.EMPTY_PAGE)
	require.NoError(t, err)
	header := strings.Split((<-requests).Header.Get("Cookie"), ";")
	for i := range header {
		header[i] = strings.TrimSpace(header[i])
	}
	slices.Sort(header)
	require.Equal(t, []string{"a=b", "c=d"}, header)
}

func TestPageRequestCookiesShouldReturnBrowserContextCookies(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, context.AddCookies([]playwright.OptionalCookie{
		{Name: "a", Value: "b", URL: playwright.String(server.EMPTY_PAGE)},
		{Name: "c", Value: "d", Domain: playwright.String("example.com"), Path: playwright.String("/")},
	}))
	cookies, err := page.Request().Cookies()
	require.NoError(t, err)
	require.Equal(t, []string{"a", "c"}, cookieNames(cookies))
	requestCookies, err := page.Request().Cookies(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Equal(t, []string{"a"}, cookieNames(requestCookies))
	contextCookies, err := context.Cookies(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Equal(t, contextCookies, requestCookies)
}

func TestPageRequestClearCookiesShouldClearBrowserContextCookies(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, context.AddCookies([]playwright.OptionalCookie{
		{Name: "a", Value: "b", URL: playwright.String(server.EMPTY_PAGE)},
		{Name: "c", Value: "d", URL: playwright.String(server.EMPTY_PAGE)},
	}))
	require.NoError(t, page.Request().ClearCookies(playwright.APIRequestContextClearCookiesOptions{Name: "a"}))
	cookies, err := context.Cookies()
	require.NoError(t, err)
	require.Equal(t, []string{"c"}, cookieNames(cookies))
	require.NoError(t, page.Request().ClearCookies())
	cookies, err = context.Cookies()
	require.NoError(t, err)
	require.Empty(t, cookies)
}

func TestGlobalFetchAddCookiesShouldAddCookiesToTheCookieJar(t *testing.T) {
	BeforeEach(t)

	request, err := pw.Request.NewContext()
	require.NoError(t, err)
	defer request.Dispose() //nolint:errcheck
	require.NoError(t, request.AddCookies([]playwright.OptionalCookie{
		{Name: "a", Value: "b", URL: playwright.String(server.EMPTY_PAGE)},
		{
			Name:     "c",
			Value:    "d",
			Domain:   playwright.String("127.0.0.1"),
			Path:     playwright.String("/input"),
			HttpOnly: playwright.Bool(true),
			SameSite: playwright.SameSiteAttributeStrict,
		},
		{Name: "e", Value: "f", Domain: playwright.String("other.com"), Path: playwright.String("/")},
	}))
	requests := server.WaitForRequestChan("/input/button.html")
	_, err = request.Get(server.PREFIX + "/input/button.html")
	require.NoError(t, err)
	require.Equal(t, "a=b; c=d", (<-requests).Header.Get("Cookie"))

	state, err := request.StorageState()
	require.NoError(t, err)
	require.Equal(t, []playwright.Cookie{
		{Name: "a", Value: "b", Domain: "127.0.0.1", Path: "/", Expires: -1, SameSite: playwright.SameSiteAttributeLax},
		{Name: "c", Value: "d", Domain: "127.0.0.1", Path: "/input", Expires: -1, HttpOnly: true, SameSite: playwright.SameSiteAttributeStrict},
		{Name: "e", Value: "f", Domain: "other.com", Path: "/", Expires: -1, SameSite: playwright.SameSiteAttributeLax},
	}, state.Cookies)

	err = request.AddCookies([]playwright.OptionalCookie{{Name: "a", Value: "b"}})
	require.ErrorContains(t, err, "Cookie should have a url or a domain/path pair")
}

func TestGlobalFetchCookiesShouldReturnCookiesFilteredByUrls(t *testing.T) {
	BeforeEach(t)

	request, err := pw.Request.NewContext()
	require.NoError(t, err)
	defer request.Dispose() //nolint:errcheck
	require.NoError(t, request.AddCookies([]playwright.OptionalCookie{
		{Name: "a", Value: "b", Domain: playwright.String("127.0.0.1"), Path: playwright.String("/")},
		{Name: "c", Value: "d", Domain: playwright.String("127.0.0.1"), Path: playwright.String("/input")},
		{Name: "e", Value: "f", Domain: playwright.String("one.com"), Path: playwright.String("/")},
		{Name: "g", Value: "h", Domain: playwright.String("two.com"), Path: playwright.String("/"), Secure: playwright.Bool(true)},
	}))
	for _, tc := range []struct {
		urls []string
		want []string
	}{
		{nil, []string{"a", "c", "e", "g"}},
		{[]string{server.EMPTY_PAGE}, []string{"a"}},
		{[]string{server.PREFIX + "/input/button.html"}, []string{"a", "c"}},
		{[]string{"http://sub.one.com/", "https://two.com/"}, []string{"e", "g"}},
		{[]string{"http://other.com/"}, []string{}},
	} {
		cookies, err := request.Cookies(tc.urls...)
		require.NoError(t, err)
		require.Equal(t, tc.want, cookieNames(cookies), "urls: %v", tc.urls)
	}
}

func TestGlobalFetchClearCookiesShouldRemoveAllCookies(t *testing.T) {
	BeforeEach(t)

	request, err := pw.Request.NewContext()
	require.NoError(t, err)
	defer request.Dispose() //nolint:errcheck
	require.NoError(t, request.AddCookies([]playwright.OptionalCookie{
		{Name: "a", Value: "b", URL: playwright.String(server.EMPTY_PAGE)},
		{Name: "c", Value: "d", Domain: playwright.String("one.com"), Path: playwright.String("/")},
	}))
	require.NoError(t, request.ClearCookies())
	cookies, err := request.Cookies()
	require.NoError(t, err)
	require.Empty(t, cookies)
	requests := server.WaitForRequestChan("/empty.html")
	_, err = request.Get(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Empty(t, (<-requests).Header.Get("Cookie"))
}

func TestGlobalFetchClearCookiesShouldFilterByNameDomainAndPath(t *testing.T) {
	BeforeEach(t)

	request, err := pw.Request.NewContext()
	require.NoError(t, err)
	defer request.Dispose() //nolint:errcheck
	require.NoError(t, request.AddCookies([]playwright.OptionalCookie{
		{Name: "session", Value: "1", Domain: playwright.String("one.com"), Path: playwright.String("/")},
		{Name: "session", Value: "2", Domain: playwright.String("two.com"), Path: playwright.String("/")},
		{Name: "session", Value: "3", Domain: playwright.String("two.com"), Path: playwright.String("/api")},
		{Name: "other", Value: "4", Domain: playwright.String("one.com"), Path: playwright.String("/")},
	}))
	values := func() []string {
		cookies, err := request.Cookies()
		require.NoError(t, err)
		values := []string{}
		for _, cookie := range cookies {
			values = append(values, cookie.Value)
		}
		slices.Sort(values)
		return values
	}
	require.NoError(t, request.ClearCookies(playwright.APIRequestContextClearCookiesOptions{Name: "session", Domain: "two.com", Path: "/api"}))
	require.Equal(t, []string{"1", "2", "4"}, values())
	require.NoError(t, request.ClearCookies(playwright.APIRequestContextClearCookiesOptions{Domain: regexp.MustCompile(`one\.com$`)}))
	require.Equal(t, []string{"2"}, values())
	require.NoError(t, request.ClearCookies(playwright.APIRequestContextClearCookiesOptions{Name: regexp.MustCompile(`^sess`)}))
	require.Empty(t, values())
}
