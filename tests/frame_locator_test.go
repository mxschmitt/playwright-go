package playwright_test

import (
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mxschmitt/playwright-go"
)

func routeIframe(t *testing.T, page playwright.Page) {
	t.Helper()

	err := page.Route("**/empty.html", func(route playwright.Route) {
		err := route.Fulfill(playwright.RouteFulfillOptions{
			Body:        `<iframe src="iframe.html" name="frame1"></iframe>`,
			ContentType: playwright.String("text/html"),
		})
		require.NoError(t, err)
	})
	require.NoError(t, err)

	err = page.Route("**/iframe.html", func(route playwright.Route) {
		routeErr := route.Fulfill(playwright.RouteFulfillOptions{
			Body: `
				<html>
					<div>
					<button data-testid="buttonId">Hello iframe</button>
					<iframe src="iframe-2.html"></iframe>
				</div>
				<span>1</span>
				<span>2</span>
				<label for=target>Name</label><input id=target type=text placeholder=Placeholder title=Title alt=Alternative>
				</html>`,
			ContentType: playwright.String("text/html"),
		})
		require.NoError(t, routeErr)
	})
	require.NoError(t, err)

	err = page.Route("**/iframe-2.html", func(route playwright.Route) {
		routeErr := route.Fulfill(playwright.RouteFulfillOptions{
			Body:        "<html><button>Hello nested iframe</button></html>",
			ContentType: playwright.String("text/html"),
		})
		require.NoError(t, routeErr)
	})
	require.NoError(t, err)
}

func routeAmbiguous(t *testing.T, page playwright.Page) {
	t.Helper()

	err := page.Route("**/empty.html", func(route playwright.Route) {
		err := route.Fulfill(playwright.RouteFulfillOptions{
			Body: `<iframe src="iframe-1.html"></iframe>
             <iframe src="iframe-2.html"></iframe>
             <iframe src="iframe-3.html"></iframe>`,
			ContentType: playwright.String("text/html"),
		})
		require.NoError(t, err)
	})
	require.NoError(t, err)

	err = page.Route("**/iframe-*", func(route playwright.Route) {
		// const path = new URL(route.request().url()).pathname.slice(1);
		u, err := url.Parse(route.Request().URL())
		require.NoError(t, err)
		path := strings.TrimLeft(u.Path, "/")
		routeErr := route.Fulfill(playwright.RouteFulfillOptions{
			Body:        fmt.Sprintf("<html><button>Hello from %s</button></html>", path),
			ContentType: playwright.String("text/html"),
		})
		require.NoError(t, routeErr)
	})
	require.NoError(t, err)
}

func TestFrameLocatorShouldWorkForIframe(t *testing.T) {
	BeforeEach(t)

	routeIframe(t, page)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)

	button := page.FrameLocator("iframe").Locator("button")
	require.NoError(t, button.WaitFor())

	innerText, err := button.InnerText()
	require.NoError(t, err)
	require.Equal(t, "Hello iframe", innerText)
	require.NoError(t, expect.Locator(button).ToHaveText("Hello iframe"))
	require.NoError(t, button.Click())
}

func TestFrameLocatorShouldWorkForNestedIframe(t *testing.T) {
	BeforeEach(t)

	routeIframe(t, page)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)

	button := page.FrameLocator("iframe").FrameLocator("iframe").Locator("button")
	require.NoError(t, button.WaitFor())

	innerText, err := button.InnerText()
	require.NoError(t, err)
	require.Equal(t, "Hello nested iframe", innerText)
	require.NoError(t, expect.Locator(button).ToHaveText("Hello nested iframe"))
	require.NoError(t, button.Click())
}

func TestFrameLocatorShouldWorkForDollar(t *testing.T) {
	BeforeEach(t)

	routeIframe(t, page)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)

	button := page.FrameLocator("iframe").Locator("button")
	require.NoError(t, button.WaitFor())

	innerText, err := button.InnerText()
	require.NoError(t, err)
	require.Equal(t, "Hello iframe", innerText)

	spans := page.FrameLocator("iframe").Locator("span")
	require.NoError(t, expect.Locator(spans).ToHaveCount(2))
}

func TestFrameLocatorGetByCoverage(t *testing.T) {
	BeforeEach(t)

	routeIframe(t, page)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)

	button1 := page.FrameLocator("iframe").GetByRole("button")
	button2 := page.FrameLocator("iframe").GetByText("Hello")
	button3 := page.FrameLocator("iframe").GetByTestId("buttonId")
	require.NoError(t, expect.Locator(button1).ToHaveText("Hello iframe"))
	require.NoError(t, expect.Locator(button2).ToHaveText("Hello iframe"))
	require.NoError(t, expect.Locator(button3).ToHaveText("Hello iframe"))

	input1 := page.FrameLocator("iframe").GetByLabel("Name")
	input2 := page.FrameLocator("iframe").GetByPlaceholder("Placeholder")
	input3 := page.FrameLocator("iframe").GetByAltText("Alternative")
	input4 := page.FrameLocator("iframe").GetByTitle("Title")
	require.NoError(t, expect.Locator(input1).ToHaveValue(""))
	require.NoError(t, expect.Locator(input2).ToHaveValue(""))
	require.NoError(t, expect.Locator(input3).ToHaveValue(""))
	require.NoError(t, expect.Locator(input4).ToHaveValue(""))
}

func TestFrameLocatorFirst(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		BeforeEach(t)

		routeAmbiguous(t, page)
		_, err := page.Goto(server.EMPTY_PAGE)
		require.NoError(t, err)
		// nolint:staticcheck
		innerText, err := page.Locator("body").FrameLocator("iframe").First().GetByRole("button").InnerText()
		require.NoError(t, err)
		require.Equal(t, "Hello from iframe-1.html", innerText)
	})

	t.Run("ambiguous", func(t *testing.T) {
		BeforeEach(t)

		routeAmbiguous(t, page)
		_, err := page.Goto(server.EMPTY_PAGE)
		require.NoError(t, err)
		// nolint:staticcheck
		innerText, err := page.Locator("body").FrameLocator("iframe").Nth(1).Locator("button").InnerText()
		require.NoError(t, err)
		require.Equal(t, "Hello from iframe-2.html", innerText)
	})
}

func TestFrameLocatorNth(t *testing.T) {
	BeforeEach(t)

	routeAmbiguous(t, page)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	// nolint:staticcheck
	innerText, err := page.Locator("body").FrameLocator("iframe").Nth(1).Locator("button").InnerText()
	require.NoError(t, err)
	require.Equal(t, "Hello from iframe-2.html", innerText)
}

func TestFrameLocatorLast(t *testing.T) {
	BeforeEach(t)

	routeAmbiguous(t, page)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	// nolint:staticcheck
	innerText, err := page.Locator("body").FrameLocator("iframe").Last().Locator("button").InnerText()
	require.NoError(t, err)
	require.Equal(t, "Hello from iframe-3.html", innerText)
}

func TestFrameLocatorLocator(t *testing.T) {
	BeforeEach(t)

	routeIframe(t, page)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)

	innerText, err := page.Locator("body").FrameLocator("iframe").Locator("span").First().InnerText()
	require.NoError(t, err)
	require.Equal(t, "1", innerText)
}

func TestFrameLocatorContentFrameShouldWork(t *testing.T) {
	BeforeEach(t)

	routeIframe(t, page)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)

	locator := page.Locator("iframe")
	frameLocator := locator.ContentFrame()
	button := frameLocator.Locator("button")

	innerText, err := button.InnerText()
	require.NoError(t, err)
	require.Equal(t, "Hello iframe", innerText)
	require.NoError(t, expect.Locator(button).ToHaveText("Hello iframe"))
	require.NoError(t, button.Click())
}

func TestFrameLocatorOwnerShouldWork(t *testing.T) {
	BeforeEach(t)

	routeIframe(t, page)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)

	frameLocator := page.FrameLocator("iframe")
	locator := frameLocator.Owner()

	require.NoError(t, expect.Locator(locator).ToBeVisible())
	name, err := locator.GetAttribute("name")
	require.NoError(t, err)
	require.Equal(t, "frame1", name)
}
