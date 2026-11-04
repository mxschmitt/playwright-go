package playwright_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/mxschmitt/playwright-go"
)

func routeAnyFramePage(t *testing.T, url, body string) {
	t.Helper()
	require.NoError(t, page.Route("**/"+url, func(route playwright.Route) {
		_ = route.Fulfill(playwright.RouteFulfillOptions{
			Body:        body,
			ContentType: playwright.String("text/html"),
		})
	}))
}

// waitForAllFrames waits until the page has frameCount frames and every child
// frame has selector attached, so that searches deterministically see all of them.
func waitForAllFrames(t *testing.T, frameCount int, selector string) {
	t.Helper()
	require.Eventually(t, func() bool { return len(page.Frames()) == frameCount }, 5*time.Second, 50*time.Millisecond)
	for _, frame := range page.Frames() {
		if frame == page.MainFrame() {
			continue
		}
		require.NoError(t, frame.Locator(selector).First().WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateAttached,
		}))
	}
}

func waitForFrameCount(t *testing.T, frameCount int) {
	t.Helper()
	require.Eventually(t, func() bool { return len(page.Frames()) == frameCount }, 5*time.Second, 50*time.Millisecond)
}

func findFrameByURL(t *testing.T, part string) playwright.Frame {
	t.Helper()
	for _, frame := range page.Frames() {
		if strings.Contains(frame.URL(), part) {
			return frame
		}
	}
	require.Failf(t, "frame not found", "no frame with URL containing %q", part)
	return nil
}

func TestFrameLocatorAnyFrameShouldClickAButtonInsideAnIframe(t *testing.T) {
	BeforeEach(t)
	routeAnyFramePage(t, "empty.html", `<iframe src="a.html"></iframe>`)
	routeAnyFramePage(t, "a.html", `<button onclick="window.__clicked = true">Click me</button>`)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.FrameLocator().GetByRole("button", playwright.FrameLocatorGetByRoleOptions{Name: "Click me"}).Click())
	clicked, err := page.Frames()[1].Evaluate(`() => window.__clicked`)
	require.NoError(t, err)
	require.Equal(t, true, clicked)
}

func TestFrameLocatorAnyFrameShouldClickAButtonInTheMainFrame(t *testing.T) {
	BeforeEach(t)
	routeAnyFramePage(t, "empty.html", `<iframe src="a.html"></iframe><button onclick="window.__clicked = true">Click me</button>`)
	routeAnyFramePage(t, "a.html", `<div>No buttons here</div>`)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.FrameLocator().Locator("button").Click())
	clicked, err := page.Evaluate(`() => window.__clicked`)
	require.NoError(t, err)
	require.Equal(t, true, clicked)
}

func TestFrameLocatorAnyFrameShouldSearchAnyFrameWithSrcdoc(t *testing.T) {
	BeforeEach(t)
	require.NoError(t, page.SetContent(`<div>main</div><iframe srcdoc="<button>Click me</button>"></iframe>`))
	require.NoError(t, expect.Locator(page.FrameLocator().GetByRole("button")).ToHaveText("Click me"))
	require.NoError(t, expect.Locator(page.MainFrame().FrameLocator().Locator("button")).ToHaveCount(1))

	require.NoError(t, page.SetContent(`<button>Main</button><iframe srcdoc="<div>No buttons here</div>"></iframe>`))
	require.NoError(t, expect.Locator(page.FrameLocator().Locator("button")).ToHaveText("Main"))
}

func TestFrameLocatorAnyFrameShouldFailClickWhenElementsMatchInMultipleFrames(t *testing.T) {
	BeforeEach(t)
	routeAnyFramePage(t, "empty.html", `<iframe src="a.html"></iframe><iframe src="b.html"></iframe>`)
	routeAnyFramePage(t, "a.html", `<button>one</button>`)
	routeAnyFramePage(t, "b.html", `<button>two</button>`)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	waitForAllFrames(t, 3, "button")
	err = page.FrameLocator().Locator("button").Click(playwright.LocatorClickOptions{Timeout: playwright.Float(3000)})
	require.ErrorContains(t, err, "frameLocator() matched elements in multiple frames")
}

func TestFrameLocatorAnyFrameShouldFailClickUponStrictModeViolationInsideASingleFrame(t *testing.T) {
	BeforeEach(t)
	routeAnyFramePage(t, "empty.html", `<iframe src="a.html"></iframe>`)
	routeAnyFramePage(t, "a.html", `<button>one</button><button>two</button>`)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	waitForAllFrames(t, 2, "button")
	err = page.FrameLocator().Locator("button").Click(playwright.LocatorClickOptions{Timeout: playwright.Float(3000)})
	require.ErrorContains(t, err, "strict mode violation")
}

func TestFrameLocatorAnyFrameShouldTimeOutOnClickWhenThereAreNoMatches(t *testing.T) {
	BeforeEach(t)
	routeAnyFramePage(t, "empty.html", `<iframe src="a.html"></iframe>`)
	routeAnyFramePage(t, "a.html", `<div>Nothing here</div>`)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	err = page.FrameLocator().Locator("button").Click(playwright.LocatorClickOptions{Timeout: playwright.Float(1000)})
	require.ErrorContains(t, err, "Timeout 1000ms exceeded")
}

func TestFrameLocatorAnyFrameShouldCountElementsInASingleFrame(t *testing.T) {
	BeforeEach(t)
	routeAnyFramePage(t, "empty.html", `<iframe src="a.html"></iframe>`)
	routeAnyFramePage(t, "a.html", `<div>1</div><div>2</div><div>3</div>`)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	waitForAllFrames(t, 2, "div")
	count, err := page.FrameLocator().Locator("div").Count()
	require.NoError(t, err)
	require.Equal(t, 3, count)
	count, err = page.FrameLocator().Locator("button").Count()
	require.NoError(t, err)
	require.Equal(t, 0, count)
}

func TestFrameLocatorAnyFrameShouldFailCountWhenElementsMatchInMultipleFrames(t *testing.T) {
	BeforeEach(t)
	routeAnyFramePage(t, "empty.html", `<div>main</div><iframe src="a.html"></iframe>`)
	routeAnyFramePage(t, "a.html", `<div>child</div>`)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	waitForAllFrames(t, 2, "div")
	_, err = page.FrameLocator().Locator("div").Count()
	require.ErrorContains(t, err, "frameLocator() matched elements in multiple frames")
}

func TestFrameLocatorAnyFrameShouldSupportToHaveCount(t *testing.T) {
	BeforeEach(t)
	routeAnyFramePage(t, "empty.html", `<iframe src="a.html"></iframe>`)
	routeAnyFramePage(t, "a.html", `<span>one</span><span>two</span>`)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, expect.Locator(page.FrameLocator().Locator("span")).ToHaveCount(2))
	require.NoError(t, expect.Locator(page.FrameLocator().Locator("button")).ToHaveCount(0))
}

func TestFrameLocatorAnyFrameShouldWaitForAFrameToAppearWithToHaveCount(t *testing.T) {
	BeforeEach(t)
	routeAnyFramePage(t, "empty.html", `<div>No frames yet</div>`)
	routeAnyFramePage(t, "a.html", `<span>one</span><span>two</span>`)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	_, err = page.Evaluate(`() => {
		setTimeout(() => {
			const iframe = document.createElement('iframe');
			iframe.src = 'a.html';
			document.body.appendChild(iframe);
		}, 500);
	}`)
	require.NoError(t, err)
	require.NoError(t, expect.Locator(page.FrameLocator().Locator("span")).ToHaveCount(2))
}

func TestFrameLocatorAnyFrameShouldFailToHaveCountWhenElementsMatchInMultipleFrames(t *testing.T) {
	BeforeEach(t)
	routeAnyFramePage(t, "empty.html", `<iframe src="a.html"></iframe><iframe src="b.html"></iframe>`)
	routeAnyFramePage(t, "a.html", `<span>one</span>`)
	routeAnyFramePage(t, "b.html", `<span>two</span>`)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	waitForAllFrames(t, 3, "span")
	err = expect.Locator(page.FrameLocator().Locator("span")).ToHaveCount(2, playwright.LocatorAssertionsToHaveCountOptions{Timeout: playwright.Float(3000)})
	require.ErrorContains(t, err, "frameLocator() matched elements in multiple frames")
}

func TestFrameLocatorAnyFrameShouldSupportToHaveText(t *testing.T) {
	BeforeEach(t)
	routeAnyFramePage(t, "empty.html", `<iframe src="a.html"></iframe>`)
	routeAnyFramePage(t, "a.html", `<div>Hello iframe</div>`)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, expect.Locator(page.FrameLocator().Locator("div")).ToHaveText("Hello iframe"))
}

func TestFrameLocatorAnyFrameShouldSupportToHaveTextWithAnArray(t *testing.T) {
	BeforeEach(t)
	routeAnyFramePage(t, "empty.html", `<iframe src="a.html"></iframe>`)
	routeAnyFramePage(t, "a.html", `<span>one</span><span>two</span>`)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, expect.Locator(page.FrameLocator().Locator("span")).ToHaveText([]string{"one", "two"}))
}

func TestFrameLocatorAnyFrameShouldFailToHaveTextWhenElementsMatchInMultipleFrames(t *testing.T) {
	BeforeEach(t)
	routeAnyFramePage(t, "empty.html", `<iframe src="a.html"></iframe><iframe src="b.html"></iframe>`)
	routeAnyFramePage(t, "a.html", `<div>one</div>`)
	routeAnyFramePage(t, "b.html", `<div>two</div>`)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	waitForAllFrames(t, 3, "div")
	err = expect.Locator(page.FrameLocator().Locator("div")).ToHaveText("one", playwright.LocatorAssertionsToHaveTextOptions{Timeout: playwright.Float(3000)})
	require.ErrorContains(t, err, "frameLocator() matched elements in multiple frames")
}

func TestFrameLocatorAnyFrameShouldSupportEvaluate(t *testing.T) {
	BeforeEach(t)
	routeAnyFramePage(t, "empty.html", `<iframe src="a.html"></iframe>`)
	routeAnyFramePage(t, "a.html", `<div data-foo="bar">Hello</div>`)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	value, err := page.FrameLocator().Locator("div").Evaluate(`e => e.getAttribute('data-foo')`, nil)
	require.NoError(t, err)
	require.Equal(t, "bar", value)
}

func TestFrameLocatorAnyFrameShouldSupportEvaluateAll(t *testing.T) {
	BeforeEach(t)
	routeAnyFramePage(t, "empty.html", `<iframe src="a.html"></iframe>`)
	routeAnyFramePage(t, "a.html", `<span>one</span><span>two</span>`)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	waitForAllFrames(t, 2, "span")
	texts, err := page.FrameLocator().Locator("span").EvaluateAll(`els => els.map(e => e.textContent)`)
	require.NoError(t, err)
	require.Equal(t, []any{"one", "two"}, texts)
	length, err := page.FrameLocator().Locator("button").EvaluateAll(`els => els.length`)
	require.NoError(t, err)
	require.Equal(t, 0, length)
}

func TestFrameLocatorAnyFrameShouldSupportHasTextFilter(t *testing.T) {
	BeforeEach(t)
	routeAnyFramePage(t, "empty.html", `<iframe src="a.html"></iframe>`)
	routeAnyFramePage(t, "a.html", `<div>foo</div><div>bar</div>`)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, expect.Locator(page.FrameLocator().Locator("div", playwright.FrameLocatorLocatorOptions{HasText: "bar"})).ToHaveText("bar"))
}

func TestFrameLocatorAnyFrameShouldSupportFirstLastNth(t *testing.T) {
	BeforeEach(t)
	routeAnyFramePage(t, "empty.html", `<iframe src="a.html"></iframe>`)
	routeAnyFramePage(t, "a.html", `<span>one</span><span>two</span><span>three</span>`)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	waitForAllFrames(t, 2, "span")
	require.NoError(t, expect.Locator(page.FrameLocator().Locator("span").First()).ToHaveText("one"))
	require.NoError(t, expect.Locator(page.FrameLocator().Locator("span").Last()).ToHaveText("three"))
	require.NoError(t, expect.Locator(page.FrameLocator().Locator("span").Nth(1)).ToHaveText("two"))
}

func TestFrameLocatorAnyFrameShouldSupportCompositeLocators(t *testing.T) {
	BeforeEach(t)
	routeAnyFramePage(t, "empty.html", `<iframe src="a.html"></iframe>`)
	routeAnyFramePage(t, "a.html", `<div><span>foo</span></div><div><i>bar</i></div>`)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, expect.Locator(page.FrameLocator().Locator("div", playwright.FrameLocatorLocatorOptions{Has: page.Locator("span")})).ToHaveText("foo"))
}

func TestFrameLocatorAnyFrameShouldOnlySearchFramesInsideTheStartingFrame(t *testing.T) {
	BeforeEach(t)
	routeAnyFramePage(t, "empty.html", `<iframe src="a.html"></iframe><button>main</button>`)
	routeAnyFramePage(t, "a.html", `<iframe src="b.html"></iframe>`)
	routeAnyFramePage(t, "b.html", `<button>deep</button>`)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	waitForFrameCount(t, 3)
	require.NoError(t, findFrameByURL(t, "b.html").Locator("button").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateAttached,
	}))
	middleFrame := findFrameByURL(t, "a.html")
	require.NoError(t, expect.Locator(middleFrame.FrameLocator().Locator("button")).ToHaveText("deep"))
}

func TestFrameLocatorAnyFrameShouldEnterAFrameFoundInANestedFrame(t *testing.T) {
	BeforeEach(t)
	routeAnyFramePage(t, "empty.html", `<iframe src="a.html"></iframe><button>main</button>`)
	routeAnyFramePage(t, "a.html", `<iframe id="target" src="b.html"></iframe><button>decoy</button>`)
	routeAnyFramePage(t, "b.html", `<button>inside</button>`)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, expect.Locator(page.FrameLocator().FrameLocator("#target").Locator("button")).ToHaveText("inside"))
}

func TestFrameLocatorAnyFrameShouldClickInsideAnEnteredFrame(t *testing.T) {
	BeforeEach(t)
	routeAnyFramePage(t, "empty.html", `<iframe src="a.html"></iframe>`)
	routeAnyFramePage(t, "a.html", `<iframe id="target" src="b.html"></iframe><button>Click me</button>`)
	routeAnyFramePage(t, "b.html", `<button onclick="window.__clicked = true">Click me</button>`)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.FrameLocator().FrameLocator("#target").GetByRole("button", playwright.FrameLocatorGetByRoleOptions{Name: "Click me"}).Click())
	clicked, err := findFrameByURL(t, "b.html").Evaluate(`() => window.__clicked`)
	require.NoError(t, err)
	require.Equal(t, true, clicked)
}

func TestFrameLocatorAnyFrameShouldNotSearchNestedFramesAfterEnteringAFrame(t *testing.T) {
	BeforeEach(t)
	routeAnyFramePage(t, "empty.html", `<iframe id="target" src="a.html"></iframe><button>main</button>`)
	routeAnyFramePage(t, "a.html", `<iframe src="b.html"></iframe>`)
	routeAnyFramePage(t, "b.html", `<button>deep</button>`)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	waitForFrameCount(t, 3)
	// The entered frame itself has no button, and its nested frames are not searched.
	require.NoError(t, expect.Locator(page.FrameLocator().FrameLocator("#target").Locator("button")).ToHaveCount(0))
}

func TestFrameLocatorAnyFrameShouldSupportTwoFrameLocators(t *testing.T) {
	BeforeEach(t)
	routeAnyFramePage(t, "empty.html", `<iframe src="a.html"></iframe>`)
	routeAnyFramePage(t, "a.html", `<iframe id="x" src="b.html"></iframe>`)
	routeAnyFramePage(t, "b.html", `<iframe id="y" src="c.html"></iframe><button>decoy</button>`)
	routeAnyFramePage(t, "c.html", `<button>bottom</button>`)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, expect.Locator(page.FrameLocator().FrameLocator("#x").FrameLocator("#y").Locator("button")).ToHaveText("bottom"))
}

func TestFrameLocatorAnyFrameShouldSupportLocatorBeforeFrameLocator(t *testing.T) {
	BeforeEach(t)
	routeAnyFramePage(t, "empty.html", `<iframe src="a.html"></iframe>`)
	routeAnyFramePage(t, "a.html", `<section><iframe src="b.html"></iframe></section><iframe src="c.html"></iframe>`)
	routeAnyFramePage(t, "b.html", `<button>in-section</button>`)
	routeAnyFramePage(t, "c.html", `<button>outside</button>`)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, expect.Locator(page.FrameLocator().Locator("section").FrameLocator("iframe").Locator("button")).ToHaveText("in-section"))
}

func TestFrameLocatorAnyFrameShouldSupportOwnerOfAFrameLocator(t *testing.T) {
	BeforeEach(t)
	routeAnyFramePage(t, "empty.html", `<iframe src="a.html"></iframe>`)
	routeAnyFramePage(t, "a.html", `<iframe id="target" src="b.html"></iframe>`)
	routeAnyFramePage(t, "b.html", `<button>inside</button>`)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	id, err := page.FrameLocator().FrameLocator("#target").Owner().GetAttribute("id")
	require.NoError(t, err)
	require.Equal(t, "target", id)
}

func TestFrameLocatorAnyFrameShouldSupportContentFrame(t *testing.T) {
	BeforeEach(t)
	routeAnyFramePage(t, "empty.html", `<iframe src="a.html"></iframe>`)
	routeAnyFramePage(t, "a.html", `<iframe id="target" src="b.html"></iframe><button>decoy</button>`)
	routeAnyFramePage(t, "b.html", `<button>inside</button>`)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, expect.Locator(page.FrameLocator().Locator("#target").ContentFrame().Locator("button")).ToHaveText("inside"))
}

func TestFrameLocatorAnyFrameShouldNotAllowFrameLocatorInsideACompositeLocator(t *testing.T) {
	BeforeEach(t)
	routeAnyFramePage(t, "empty.html", `<button>main</button><iframe src="a.html"></iframe>`)
	routeAnyFramePage(t, "a.html", `<a href="#">link</a>`)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	waitForAllFrames(t, 2, "a")

	_, err = page.Locator("button").Or(page.FrameLocator().Locator("a")).Count()
	require.ErrorContains(t, err, "frameLocator() is not allowed inside composite locators")

	_, err = page.Locator("button").Filter(playwright.LocatorFilterOptions{Has: page.FrameLocator().Locator("a")}).Count()
	require.ErrorContains(t, err, "frameLocator() is not allowed inside composite locators")

	// With frameLocator() first, it applies to the whole locator, so both operands are searched in every frame.
	_, err = page.FrameLocator().Locator("a").Or(page.Locator("button")).Count()
	require.ErrorContains(t, err, "frameLocator() matched elements in multiple frames")
}

func TestFrameLocatorAnyFrameShouldSupportACompositeLocatorUnderFrameLocator(t *testing.T) {
	BeforeEach(t)
	routeAnyFramePage(t, "empty.html", `<iframe src="a.html"></iframe>`)
	routeAnyFramePage(t, "a.html", `<div class="classname">first</div><button>second</button>`)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	waitForAllFrames(t, 2, "button")
	require.NoError(t, expect.Locator(page.FrameLocator().Locator(".classname").Or(page.GetByRole("button"))).ToHaveText([]string{"first", "second"}))
}

func TestFrameLocatorAnyFrameShouldNotAllowFirstLastNth(t *testing.T) {
	BeforeEach(t)
	for _, frameLocator := range []playwright.FrameLocator{
		page.FrameLocator().First(), //nolint:staticcheck // the deprecated API must report the error
		page.FrameLocator().Last(),  //nolint:staticcheck // the deprecated API must report the error
		page.FrameLocator().Nth(1),  //nolint:staticcheck // the deprecated API must report the error
	} {
		_, err := frameLocator.Locator("button").Count()
		require.ErrorContains(t, err, "selecting the nth frame is not allowed")
		_, err = frameLocator.Owner().Count()
		require.ErrorContains(t, err, "selecting the nth frame is not allowed")
	}
}

func TestFrameLocatorAnyFrameShouldNotAllowOwner(t *testing.T) {
	BeforeEach(t)
	_, err := page.FrameLocator().Owner().Count()
	require.ErrorContains(t, err, "Selector cannot be empty after frameLocator()")
}

func TestFrameLocatorAnyFrameNthErrorSurvivesDerivedLocators(t *testing.T) {
	BeforeEach(t)
	// A single frame with a single button: without the error the actions below would succeed.
	routeAnyFramePage(t, "empty.html", `<iframe src="a.html"></iframe>`)
	routeAnyFramePage(t, "a.html", `<button onclick="window.__clicked = true">one</button>`)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	waitForAllFrames(t, 2, "button")

	nth := page.FrameLocator().First() //nolint:staticcheck // the deprecated API must report the error
	button := nth.Locator("button")
	for _, tc := range []struct {
		name    string
		locator playwright.Locator
	}{
		{"First", button.First()},
		{"Last", button.Last()},
		{"Nth", button.Nth(0)},
		{"Filter", button.Filter(playwright.LocatorFilterOptions{HasText: "one"})},
		{"Visible", button.Visible()},
		{"Locator", nth.Locator("body").Locator("button")},
		{"GetByRole", nth.Locator("body").GetByRole("button")},
		{"GetByText", nth.GetByText("one")},
		{"Describe", button.Describe("the button")},
		{"Or", page.Locator("button").Or(button)},
		{"And", page.Locator("button").And(button)},
		{"Has", page.Locator("body").Filter(playwright.LocatorFilterOptions{Has: button})},
		{"Within", page.Locator("button").Within(nth.Locator("body"))},
		{"ContentFrame", nth.Locator("body").ContentFrame().Locator("button")},
		{"FrameLocator", nth.FrameLocator("iframe").Locator("button")},
		{"Chained twice", button.First().Locator("xpath=.").First()},
	} {
		name, locator := tc.name, tc.locator
		require.ErrorContains(t, locator.Click(playwright.LocatorClickOptions{Timeout: playwright.Float(1000)}), "selecting the nth frame is not allowed", name)
		_, err := locator.Count()
		require.ErrorContains(t, err, "selecting the nth frame is not allowed", name)
		require.ErrorContains(t, expect.Locator(locator).ToHaveCount(1), "selecting the nth frame is not allowed", name)
	}
	clicked, err := page.Frames()[1].Evaluate(`() => window.__clicked`)
	require.NoError(t, err)
	require.Nil(t, clicked)
}

func TestLocatorInvalidArgumentErrorSurvivesDerivedLocators(t *testing.T) {
	BeforeEach(t)
	require.NoError(t, page.SetContent(`<button>42</button>`))

	invalid := page.GetByText(42)
	require.Error(t, invalid.Err())
	for _, tc := range []struct {
		name    string
		locator playwright.Locator
	}{
		{"First", invalid.First()},
		{"Nth", invalid.Nth(0)},
		{"Filter", invalid.Filter()},
		{"Locator", invalid.Locator("xpath=.")},
		{"Or", page.Locator("button").Or(invalid)},
	} {
		_, err := tc.locator.Count()
		require.ErrorIs(t, err, invalid.Err(), tc.name)
	}
}
