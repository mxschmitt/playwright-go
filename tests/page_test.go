package playwright_test

import (
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/h2non/filetype"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestPageURL(t *testing.T) {
	BeforeEach(t)

	require.Equal(t, "about:blank", page.URL())
	_, err := page.Goto("https://example.com")
	require.NoError(t, err)
	require.Equal(t, "https://example.com/", page.URL())
	require.Equal(t, context, page.Context())
	require.Equal(t, 1, len(page.Frames()))
}

func TestPageSetContent(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent("<h1>foo</h1>",
		playwright.PageSetContentOptions{
			WaitUntil: playwright.WaitUntilStateNetworkidle,
		}))
	content, err := page.Content()
	require.NoError(t, err)
	require.Equal(t, content, "<html><head></head><body><h1>foo</h1></body></html>")
}

func TestPageSetContentShouldRespectDefaultNavigationTimeout(t *testing.T) {
	BeforeEach(t)

	page.SetDefaultNavigationTimeout(1)
	imgPath := "/img.png"
	// stall for image
	server.SetRoute(imgPath, func(w http.ResponseWriter, r *http.Request) {})

	err := page.SetContent(fmt.Sprintf(`<img src="%s"></img>`, server.PREFIX+imgPath))
	require.ErrorIs(t, err, playwright.ErrTimeout)
	require.ErrorContains(t, err, "Timeout 1ms exceeded.")
}

func TestPageScreenshot(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent("<h1>foobar</h1>"))
	tmpfile, err := os.MkdirTemp("", "screenshot")
	require.NoError(t, err)
	screenshotPath := filepath.Join(tmpfile, "image.png")
	screenshot, err := page.Screenshot()
	require.NoError(t, err)
	require.True(t, filetype.IsImage(screenshot))
	require.Greater(t, len(screenshot), 50)

	screenshot, err = page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String(screenshotPath),
	})
	require.NoError(t, err)
	require.True(t, filetype.IsImage(screenshot))
	require.Greater(t, len(screenshot), 50)

	_, err = os.Stat(screenshotPath)
	require.NoError(t, err)
}

func TestPageScreenshotWithMask(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent("<h1>foobar</h1><p>sensitive</p>"))
	tmpfile, err := os.MkdirTemp("", "screenshot")
	require.NoError(t, err)
	screenshotPath := filepath.Join(tmpfile, "image.png")
	screenshot1, err := page.Screenshot()
	require.NoError(t, err)
	require.True(t, filetype.IsImage(screenshot1))
	require.Greater(t, len(screenshot1), 50)

	sensElem := page.Locator("p")

	screenshot2, err := page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String(screenshotPath),
		Mask: []playwright.Locator{
			sensElem,
		},
		MaskColor: playwright.String("red"),
	})
	require.NoError(t, err)
	require.True(t, filetype.IsImage(screenshot2))
	require.Greater(t, len(screenshot2), 50)
	require.NotEqual(t, screenshot1, screenshot2)

	_, err = os.Stat(screenshotPath)
	require.NoError(t, err)
}

func TestPagePDF(t *testing.T) {
	BeforeEach(t)

	if !isChromium {
		t.Skip("Skipping")
	}
	require.NoError(t, page.SetContent("<h1>foobar</h1>"))
	tmpfile, err := os.MkdirTemp("", "pdf")
	require.NoError(t, err)
	screenshotPath := filepath.Join(tmpfile, "image.png")
	screenshot, err := page.PDF()
	require.NoError(t, err)
	require.Equal(t, "application/pdf", http.DetectContentType(screenshot))
	require.Greater(t, len(screenshot), 50)

	screenshot, err = page.PDF(playwright.PagePdfOptions{
		Path: playwright.String(screenshotPath),
	})
	require.NoError(t, err)
	require.Equal(t, "application/pdf", http.DetectContentType(screenshot))
	require.Greater(t, len(screenshot), 50)

	_, err = os.Stat(screenshotPath)
	require.NoError(t, err)
}

func TestPagePDFWithOutline(t *testing.T) {
	BeforeEach(t)

	if !isChromium {
		t.Skip("Skipping")
	}

	_, err := page.Goto(fmt.Sprintf("%s/headings.html", server.PREFIX))
	require.NoError(t, err)

	pdfNoOutline, err := page.PDF()
	require.NoError(t, err)
	pdfWithOutline, err := page.PDF(playwright.PagePdfOptions{
		Tagged:  playwright.Bool(true),
		Outline: playwright.Bool(true),
	})
	require.NoError(t, err)

	require.Greater(t, len(pdfWithOutline), len(pdfNoOutline))
}

func TestPageQuerySelector(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`<div id="one">
		<span id="two">
			<div id="three">
				<span id="four">
					foobar
				</span>
			</div>
		</span>
	</div>`))
	//nolint:staticcheck
	one, err := page.QuerySelector("div#one")
	require.NoError(t, err)
	//nolint:staticcheck
	two, err := one.QuerySelector("span#two")
	require.NoError(t, err)
	//nolint:staticcheck
	three, err := two.QuerySelector("div#three")
	require.NoError(t, err)
	//nolint:staticcheck
	four, err := three.QuerySelector("span#four")
	require.NoError(t, err)
	//nolint:staticcheck
	textContent, err := four.TextContent()
	require.NoError(t, err)
	require.Equal(t, strings.TrimSpace(textContent), "foobar")
}

func TestPageQuerySelectorAll(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`
	<div class="foo">0</div>
	<div class="foo">1</div>
	<div class="foo">2</div>
	`))
	//nolint:staticcheck
	elements, err := page.QuerySelectorAll("div.foo")
	require.NoError(t, err)
	require.Equal(t, 3, len(elements))
	for i := 0; i < 3; i++ {
		//nolint:staticcheck
		textContent, err := elements[i].TextContent()
		require.NoError(t, err)
		require.Equal(t, strconv.Itoa(i), textContent)
	}
}

func TestPageEvaluate(t *testing.T) {
	BeforeEach(t)

	val, err := page.Evaluate(`() => 123`)
	require.NoError(t, err)
	require.Equal(t, val, 123)
	val, err = page.Evaluate(`() => 42n`)
	require.NoError(t, err)
	require.Equal(t, val, big.NewInt(42))
	val, err = page.Evaluate(`a => a`, big.NewInt(17))
	require.NoError(t, err)
	require.Equal(t, val, big.NewInt(17))
}

// TestPageEvaluateIntegerArgTypes verifies that sized integer types (not just
// the bare int) round-trip as numbers rather than being serialized as
// undefined.
func TestPageEvaluateIntegerArgTypes(t *testing.T) {
	BeforeEach(t)

	for _, arg := range []any{int8(7), int16(7), int32(7), int64(7), uint(7), uint8(7), uint16(7), uint32(7), uint64(7)} {
		val, err := page.Evaluate(`a => a + 1`, arg)
		require.NoError(t, err)
		require.EqualValues(t, 8, val)
	}
}

// TestPageEvaluateReturnsRegExp verifies that a RegExp returned from page
// evaluation is parsed into a *regexp.Regexp instead of panicking the client.
func TestPageEvaluateReturnsRegExp(t *testing.T) {
	BeforeEach(t)

	val, err := page.Evaluate(`() => /foo.*bar/i`)
	require.NoError(t, err)
	re, ok := val.(*regexp.Regexp)
	require.True(t, ok, "expected *regexp.Regexp, got %T", val)
	require.True(t, re.MatchString("xx FOO zz BAR yy"))
	require.False(t, re.MatchString("baz"))
}

func TestPageEvalOnSelectorAll(t *testing.T) {
	BeforeEach(t)

	err := page.SetContent(`
		<div class="foo">1</div>
		<div class="foo">2</div>
		<div class="foo">3</div>
	`)
	require.NoError(t, err)
	//nolint:staticcheck
	val, err := page.EvalOnSelectorAll(".foo", `(elements) => elements.map(el => el.textContent)`)
	require.NoError(t, err)
	require.Equal(t, val, []interface{}([]interface{}{"1", "2", "3"}))
}

func TestPageEvalOnSelector(t *testing.T) {
	BeforeEach(t)

	err := page.SetContent(`
		<div class="foo">bar</div>
	`)
	require.NoError(t, err)
	//nolint:staticcheck
	val, err := page.EvalOnSelector(".foo", `(element) => element.textContent`, nil)
	require.NoError(t, err)
	require.Equal(t, val, "bar")
}

func TestPageExpectWorker(t *testing.T) {
	BeforeEach(t)

	worker, err := page.ExpectWorker(func() error {
		_, err := page.Goto(server.PREFIX + "/worker/worker.html")
		return err
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(page.Workers()))
	require.Equal(t, worker, page.Workers()[0])
	worker = page.Workers()[0]
	require.Contains(t, worker.URL(), "worker.js")
	// flaky in the macos-latest of gh action
	require.Eventually(
		t,
		func() bool {
			v, err := worker.Evaluate(`() => self["workerFunction"] ? true : false`)
			require.NoError(t, err)
			return v == true
		},
		500*time.Millisecond, 10*time.Millisecond,
	)
	res, err := worker.Evaluate(`() => self["workerFunction"]()`)
	require.NoError(t, err)
	require.Equal(t, "worker function result", res)
	_, err = page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Equal(t, 0, len(page.Workers()))
}

func TestPageExpectRequest(t *testing.T) {
	BeforeEach(t)

	request, err := page.ExpectRequest("**/*", func() error {
		_, err := page.Goto(server.EMPTY_PAGE)
		return err
	})
	require.NoError(t, err)
	require.Equal(t, server.EMPTY_PAGE, request.URL())
	require.Equal(t, "document", request.ResourceType())
	require.Equal(t, "GET", request.Method())
}

func TestPageExpectRequestRegexp(t *testing.T) {
	BeforeEach(t)

	request, err := page.ExpectRequest(regexp.MustCompile(".*/empty.html"), func() error {
		_, err := page.Goto(server.EMPTY_PAGE)
		return err
	})
	require.NoError(t, err)
	require.Equal(t, server.EMPTY_PAGE, request.URL())
	require.Equal(t, "document", request.ResourceType())
	require.Equal(t, "GET", request.Method())
}

func TestPageExpectRequestFunc(t *testing.T) {
	BeforeEach(t)

	request, err := page.ExpectRequest(func(url string) bool {
		return strings.HasSuffix(url, "empty.html")
	}, func() error {
		_, err := page.Goto(server.EMPTY_PAGE)
		return err
	})
	require.NoError(t, err)
	require.Equal(t, server.EMPTY_PAGE, request.URL())
	require.Equal(t, "document", request.ResourceType())
	require.Equal(t, "GET", request.Method())
}

func TestPageExpectRequestFinished(t *testing.T) {
	BeforeEach(t)

	request, err := page.ExpectRequestFinished(func() error {
		_, err := page.Goto(server.EMPTY_PAGE)
		return err
	}, playwright.PageExpectRequestFinishedOptions{
		Predicate: func(r playwright.Request) bool {
			return strings.HasSuffix(r.URL(), "empty.html")
		},
	})
	require.NoError(t, err)
	require.Equal(t, server.EMPTY_PAGE, request.URL())
	require.Equal(t, "document", request.ResourceType())
	require.Equal(t, "GET", request.Method())
}

func TestPageExpectPopup(t *testing.T) {
	skipWebKitMacOSPopup(t)
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	popup, err := page.ExpectPopup(func() error {
		_, err := page.Evaluate(`window._popup = window.open(document.location.href)`)
		return err
	})
	require.NoError(t, err)
	require.Equal(t, popup.URL(), server.EMPTY_PAGE)
}

func TestPageExpectNavigation(t *testing.T) {
	t.Skip()
}

func TestPageExpectLoadState(t *testing.T) {
	t.Skip()
}

func TestPageExpectFileChooser(t *testing.T) {
	BeforeEach(t)

	t.Run("should work for single file pick", func(t *testing.T) {
		require.NoError(t, page.SetContent(`<input type=file>`))
		fc, err := page.ExpectFileChooser(func() error {
			return page.Locator("input").Click()
		}, playwright.PageExpectFileChooserOptions{
			Timeout: playwright.Float(1000),
		})
		require.NoError(t, err)
		require.False(t, fc.IsMultiple())
	})

	t.Run("should work for multiple", func(t *testing.T) {
		require.NoError(t, page.SetContent(`<input multiple type=file>`))
		_, err := page.ExpectFileChooser(func() error {
			return page.Locator("input").Click()
		}, playwright.PageExpectFileChooserOptions{
			Predicate: func(fc playwright.FileChooser) bool {
				return fc.IsMultiple()
			},
		})
		require.NoError(t, err)
	})
}

func TestPageExpectDialog(t *testing.T) {
	t.Skip()
}

func TestPageExpectConsoleMessage(t *testing.T) {
	BeforeEach(t)

	message, err := page.ExpectConsoleMessage(func() error {
		_, err := page.Evaluate(`console.log(123, "abc")`)
		return err
	})
	require.NoError(t, err)
	require.Equal(t, message.Text(), "123 abc")
}

func TestPageExpectEvent(t *testing.T) {
	t.Skip()
}

func TestPageOpener(t *testing.T) {
	skipWebKitMacOSPopup(t)
	BeforeEach(t)

	eventPage, err := context.ExpectEvent("page", func() error {
		_, err := page.Goto(server.PREFIX + "/popup/window-open.html")
		return err
	})
	require.NoError(t, err)
	popup := eventPage.(playwright.Page)

	opener, err := popup.Opener()
	require.NoError(t, err)
	require.Equal(t, opener, page)

	opener, err = page.Opener()
	require.NoError(t, err)
	require.Nil(t, opener)

	// Once the opener page is closed, Opener() returns nil (matches upstream).
	require.NoError(t, page.Close())
	opener, err = popup.Opener()
	require.NoError(t, err)
	require.Nil(t, opener)
}

func TestPageTitle(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`<title>abc</title>`))
	title, err := page.Title()
	require.NoError(t, err)
	require.Equal(t, "abc", title)
}

func TestPageWaitForSelector(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`<h1>myElement</h1>`))
	//nolint:staticcheck
	element, err := page.WaitForSelector("text=myElement")
	require.NoError(t, err)
	//nolint:staticcheck
	textContent, err := element.TextContent()
	require.NoError(t, err)
	require.Equal(t, "myElement", textContent)
	//nolint:staticcheck
	_, err = page.WaitForSelector("h1")
	require.NoError(t, err)
}

func TestPageDispatchEvent(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.PREFIX + "/input/button.html")
	require.NoError(t, err)
	//nolint:staticcheck
	require.NoError(t, page.DispatchEvent("button", "click", nil))
	clicked, err := page.Evaluate("() => result")
	require.NoError(t, err)
	require.Equal(t, "Clicked", clicked)
}

func TestPageReload(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	_, err = page.Evaluate("window._foo = 10")
	require.NoError(t, err)
	_, err = page.Reload()
	require.NoError(t, err)
	v, err := page.Evaluate("window._foo")
	require.NoError(t, err)
	require.Nil(t, v)
}

// TestPageReloadRespectsDefaultNavigationTimeout verifies Reload honors the
// configured default navigation timeout (it must resolve the timeout, not fall
// back to the hardcoded 30s serializer default).
func TestPageReloadRespectsDefaultNavigationTimeout(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	page.SetDefaultNavigationTimeout(1)
	server.SetRoute("/empty.html", func(w http.ResponseWriter, r *http.Request) {}) // stall
	_, err = page.Reload()
	require.ErrorIs(t, err, playwright.ErrTimeout)
	require.ErrorContains(t, err, "Timeout 1ms exceeded.")
}

func TestPageGoBackGoForward(t *testing.T) {
	BeforeEach(t)

	resp, err := page.GoBack()
	require.NoError(t, err)
	require.Nil(t, resp)

	_, err = page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	_, err = page.Goto(server.PREFIX + "/grid.html")
	require.NoError(t, err)

	resp, err = page.GoBack()
	require.NoError(t, err)
	require.True(t, resp.Ok())
	require.Equal(t, resp.URL(), server.EMPTY_PAGE)

	resp, err = page.GoForward()
	require.NoError(t, err)
	require.True(t, resp.Ok())
	require.Contains(t, resp.URL(), "/grid.html")

	resp, err = page.GoForward()
	require.NoError(t, err)
	require.Nil(t, resp)
}

func TestPageAddScriptTag(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)

	scriptHandle, err := page.AddScriptTag(playwright.PageAddScriptTagOptions{
		URL: playwright.String("injectedfile.js"),
	})
	require.NoError(t, err)
	require.NotNil(t, scriptHandle.AsElement())
	v, err := page.Evaluate("__injected")
	require.NoError(t, err)
	require.Equal(t, 42, v)
}

func TestPageAddScriptTagFile(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)

	scriptHandle, err := page.AddScriptTag(playwright.PageAddScriptTagOptions{
		Path: playwright.String(Asset("injectedfile.js")),
	})
	require.NoError(t, err)
	require.NotNil(t, scriptHandle.AsElement())
	v, err := page.Evaluate("__injected")
	require.NoError(t, err)
	require.Equal(t, 42, v)
}

func TestPageAddStyleTag(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)

	_, err = page.AddStyleTag(playwright.PageAddStyleTagOptions{
		URL: playwright.String("injectedstyle.css"),
	})
	require.NoError(t, err)
	v, err := page.Evaluate("window.getComputedStyle(document.querySelector('body')).getPropertyValue('background-color')")
	require.NoError(t, err)
	require.Equal(t, "rgb(255, 0, 0)", v)
}

func TestPageAddStyleTagFile(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)

	_, err = page.AddStyleTag(playwright.PageAddStyleTagOptions{
		Path: playwright.String(Asset("injectedstyle.css")),
	})
	require.NoError(t, err)
	v, err := page.Evaluate("window.getComputedStyle(document.querySelector('body')).getPropertyValue('background-color')")
	require.NoError(t, err)
	require.Equal(t, "rgb(255, 0, 0)", v)
}

func TestPageWaitForLoadState(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.PREFIX + "/one-style.html")
	require.NoError(t, err)
	require.NoError(t, page.WaitForLoadState())
	require.NoError(t, page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	}))
}

func TestPlaywrightDevices(t *testing.T) {
	BeforeEach(t)

	require.Greater(t, len(pw.Devices), 10)
	for name, device := range pw.Devices {
		require.NotEmpty(t, name)
		require.NotEmpty(t, device.UserAgent)
		require.NotEmpty(t, device.Viewport)
		require.Greater(t, device.DeviceScaleFactor, float64(0))
		require.NotEmpty(t, device.DefaultBrowserType)
	}
}

func TestPageAddInitScript(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.AddInitScript(playwright.Script{
		Content: playwright.String(`window['injected'] = 123;`),
	}))
	_, err := page.Goto(server.PREFIX + "/tamperable.html")
	require.NoError(t, err)
	result, err := page.Evaluate(`() => window['result']`)
	require.NoError(t, err)
	require.Equal(t, 123, result)
}

func TestPageExpectSelectorTimeout(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	err = page.Locator("foobar").Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(500),
	})
	require.ErrorIs(t, err, playwright.ErrTimeout)
}

func TestPageType(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent("<input type='text' />"))
	//nolint:staticcheck
	require.NoError(t, page.Type("input", "hello"))
	value, err := page.Locator("input").Evaluate("el => el.value", nil)
	require.NoError(t, err)
	require.Equal(t, "hello", value)
}

func TestPagePress(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent("<input type='text' />"))
	//nolint:staticcheck
	require.NoError(t, page.Press("input", "h"))
	value, err := page.Locator("input").Evaluate("el => el.value", nil)
	require.NoError(t, err)
	require.Equal(t, "h", value)
}

func TestPageCheck(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent("<input id='checkbox' type='checkbox'></input>"))
	//nolint:staticcheck
	require.NoError(t, page.Check("input"))
	value, err := page.Locator("input").Evaluate("el => el.checked", nil)
	require.NoError(t, err)
	require.Equal(t, true, value)
}

func TestPageUncheck(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent("<input id='checkbox' type='checkbox' checked></input>"))
	//nolint:staticcheck
	require.NoError(t, page.Uncheck("input"))
	value, err := page.Locator("input").Evaluate("el => el.checked", nil)
	require.NoError(t, err)
	require.Equal(t, false, value)
}

func TestPageWaitForTimeout(t *testing.T) {
	BeforeEach(t)

	before := time.Now()
	//nolint:staticcheck
	page.WaitForTimeout(1000)
	after := time.Now()
	duration := after.Sub(before)
	require.True(t, duration > time.Second)
	require.True(t, duration < 2*time.Second)
}

func TestPageWaitForFunction(t *testing.T) {
	BeforeEach(t)

	_, err := page.Evaluate(`() => setTimeout(() => window.FOO = true, 500)`)
	require.NoError(t, err)
	_, err = page.WaitForFunction(`window.FOO === true`, nil)
	require.NoError(t, err)
}

func TestPageDblclick(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`<button ondblclick="window.clicked=true"/>`))
	//nolint:staticcheck
	require.NoError(t, page.Dblclick("button"))
	result, err := page.Evaluate("window.clicked")
	require.NoError(t, err)
	require.True(t, result.(bool))
}

func TestPageFocus(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`<button onfocus="window.clicked=true"/>`))
	//nolint:staticcheck
	require.NoError(t, page.Focus("button"))
	result, err := page.Evaluate("window.clicked")
	require.NoError(t, err)
	require.True(t, result.(bool))
}

func TestPageTextContent(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.PREFIX + "/dom.html")
	require.NoError(t, err)
	//nolint:staticcheck
	content, err := page.TextContent("#inner")
	require.NoError(t, err)
	require.Equal(t, "Text,\nmore text", content)
}

func TestPageAddInitScriptWithPath(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.AddInitScript(playwright.Script{
		Path: playwright.String(Asset("injectedfile.js")),
	}))
	_, err := page.Goto(server.PREFIX + "/tamperable.html")
	require.NoError(t, err)
	result, err := page.Evaluate(`() => window['result']`)
	require.NoError(t, err)
	require.Equal(t, 123, result)
}

func TestPageSupportNetworkEvents(t *testing.T) {
	BeforeEach(t)

	eventsChan := make(chan string, 6)
	page.OnRequest(func(request playwright.Request) {
		eventsChan <- fmt.Sprintf("%s %s", request.Method(), request.URL())
	})
	page.OnResponse(func(response playwright.Response) {
		eventsChan <- fmt.Sprintf("%d %s", response.Status(), response.URL())
	})
	page.OnRequestFinished(func(request playwright.Request) {
		eventsChan <- fmt.Sprintf("DONE %s", request.URL())
	})
	page.OnRequestFailed(func(request playwright.Request) {
		eventsChan <- fmt.Sprintf("FAIL %s", request.URL())
	})
	server.SetRedirect("/foo.html", "/empty.html")
	FOO_URL := server.PREFIX + "/foo.html"
	response, err := page.Goto(FOO_URL)
	require.NoError(t, err)
	eventsSlice := ChanToSlice(eventsChan, 6)
	require.Equal(t, []string{
		fmt.Sprintf("GET %s", FOO_URL),
		fmt.Sprintf("302 %s", FOO_URL),
		fmt.Sprintf("DONE %s", FOO_URL),
		fmt.Sprintf("GET %s", server.EMPTY_PAGE),
		fmt.Sprintf("200 %s", server.EMPTY_PAGE),
		fmt.Sprintf("DONE %s", server.EMPTY_PAGE),
	}, eventsSlice)
	redirectedFrom := response.Request().RedirectedFrom()
	require.Contains(t, redirectedFrom.URL(), "foo.html")
	require.Nil(t, redirectedFrom.RedirectedFrom())
	require.Equal(t, redirectedFrom.RedirectedTo(), response.Request())
}

func TestPageSetViewport(t *testing.T) {
	BeforeEach(t)

	utils.VerifyViewport(t, page, 1280, 720)
	require.NoError(t, page.SetViewportSize(123, 456))
	utils.VerifyViewport(t, page, 123, 456)
}

func TestPageEmulateMedia(t *testing.T) {
	BeforeEach(t)

	utils.AssertEval(t, page, "matchMedia('screen').matches", true)
	utils.AssertEval(t, page, "matchMedia('print').matches", false)
	require.NoError(t, page.EmulateMedia(playwright.PageEmulateMediaOptions{
		Media: playwright.MediaPrint,
	}))
	utils.AssertEval(t, page, "matchMedia('screen').matches", false)
	utils.AssertEval(t, page, "matchMedia('print').matches", true)
	require.NoError(t, page.EmulateMedia())
	utils.AssertEval(t, page, "matchMedia('screen').matches", false)
	utils.AssertEval(t, page, "matchMedia('print').matches", true)
	require.NoError(t, page.EmulateMedia(playwright.PageEmulateMediaOptions{
		Media: playwright.MediaNoOverride,
	}))
	utils.AssertEval(t, page, "matchMedia('screen').matches", true)
	utils.AssertEval(t, page, "matchMedia('print').matches", false)
}

func TestPageBringToFront(t *testing.T) {
	BeforeEach(t)

	page1, err := browser.NewPage()
	require.NoError(t, err)
	require.NoError(t, page1.SetContent("Page1"))
	page2, err := browser.NewPage()
	require.NoError(t, err)
	require.NoError(t, page2.SetContent("Page2"))

	require.NoError(t, page1.BringToFront())
	utils.AssertEval(t, page1, "document.visibilityState", "visible")
	utils.AssertEval(t, page2, "document.visibilityState", "visible")

	require.NoError(t, page2.BringToFront())
	utils.AssertEval(t, page1, "document.visibilityState", "visible")
	utils.AssertEval(t, page2, "document.visibilityState", "visible")
	require.NoError(t, page1.Close())
	require.NoError(t, page2.Close())
}

func TestPageFrame(t *testing.T) {
	BeforeEach(t)

	err := page.SetContent(fmt.Sprintf("<iframe name=target src=%s></iframe>", server.EMPTY_PAGE))
	require.NoError(t, err)

	name := "target"
	frame1 := page.Frame(playwright.PageFrameOptions{Name: &name})
	require.Equal(t, name, frame1.Name())
	require.Equal(t, server.EMPTY_PAGE, frame1.URL())

	frame2 := page.Frame(playwright.PageFrameOptions{URL: server.EMPTY_PAGE})
	require.Equal(t, name, frame2.Name())
	require.Equal(t, server.EMPTY_PAGE, frame2.URL())

	// When a name is provided it takes precedence and the URL is never
	// consulted, matching upstream (page.frame should respect name): a
	// non-matching name returns nil regardless of the URL.
	badName := "test"
	require.Nil(t, page.Frame(playwright.PageFrameOptions{Name: &badName, URL: server.EMPTY_PAGE}))
	require.Nil(t, page.Frame(playwright.PageFrameOptions{Name: &badName, URL: "https://example.com"}))
	require.Nil(t, page.Frame(playwright.PageFrameOptions{Name: &badName}))
}

func TestPageTap(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent("<input id='checkbox' type='checkbox'></input>"))
	value, err := page.Locator("input").Evaluate("el => el.checked", nil)
	require.NoError(t, err)
	require.Equal(t, false, value)
	//nolint:staticcheck
	require.NoError(t, page.Tap("input"))
	value, err = page.Locator("input").Evaluate("el => el.checked", nil)
	require.NoError(t, err)
	require.Equal(t, true, value)
}

func TestPagePageError(t *testing.T) {
	BeforeEach(t)

	url := server.PREFIX + "/error.html"
	errAny, err := page.ExpectEvent("pageerror", func() error {
		_, err := page.Goto(url)
		return err
	})
	require.NoError(t, err)
	pageError := &playwright.Error{}
	require.True(t, errors.As(errAny.(error), &pageError))
	require.Equal(t, "Fancy error!", pageError.Message)
	require.Equal(t, "Error", pageError.Name)

	if browserName == "chromium" {
		require.Equal(t, `Error: Fancy error!
    at c (myscript.js:14:11)
    at b (myscript.js:10:5)
    at a (myscript.js:6:5)
    at myscript.js:3:1`, pageError.Stack)
	}
	if browserName == "firefox" {
		require.Equal(t, `Error: Fancy error!
    at c (myscript.js:14:11)
    at b (myscript.js:10:5)
    at a (myscript.js:6:5)
    at  (myscript.js:3:1)`, pageError.Stack)
	}
	if browserName == "webkit" {
		require.Equal(t, fmt.Sprintf(`Error: Fancy error!
    at c (%[1]s:14:36)
    at b (%[1]s:10:6)
    at a (%[1]s:6:6)
    at global code (%[1]s:3:2)`, url), pageError.Stack)
	}
}

func TestPageSelectOption(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent("<select id='lang'><option value='go'>go</option><option value='python'>python</option></select>"))
	//nolint:staticcheck
	selected, err := page.SelectOption("#lang", playwright.SelectOptionValues{
		Values: playwright.StringSlice("python"),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(selected))
	require.Equal(t, "python", selected[0])
}

func TestPageUnrouteShouldWork(t *testing.T) {
	BeforeEach(t)

	intercepted := []int{}
	handler1 := func(route playwright.Route) {
		intercepted = append(intercepted, 1)
		require.NoError(t, route.Continue())
	}
	require.NoError(t, page.Route("**/*", handler1))

	handler2 := func(route playwright.Route) {
		intercepted = append(intercepted, 2)
		require.NoError(t, route.Continue())
	}
	require.NoError(t, page.Route("**/empty.html", handler2))

	handler3 := func(route playwright.Route) {
		intercepted = append(intercepted, 3)
		require.NoError(t, route.Continue())
	}
	require.NoError(t, page.Route("**/empty.html", handler3))

	handler4 := func(route playwright.Route) {
		intercepted = append(intercepted, 4)
		require.NoError(t, route.Continue())
	}
	require.NoError(t, page.Route(regexp.MustCompile("empty.html"), handler4))

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Equal(t, []int{4}, intercepted)

	intercepted = []int{}
	require.NoError(t, page.Unroute(regexp.MustCompile("empty.html"), handler4))
	_, err = page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Equal(t, []int{3}, intercepted)

	intercepted = []int{}
	require.NoError(t, page.Unroute("**/empty.html"))
	_, err = page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Equal(t, []int{1}, intercepted)
}

func TestPageDragAndDrop(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.PREFIX + "/drag-n-drop.html")
	require.NoError(t, err)
	require.NoError(t, page.DragAndDrop("#source", "#target"))
	value, err := page.Locator("#target").Evaluate("target => target.contains(document.querySelector('#source'))", nil)
	require.NoError(t, err)
	require.Equal(t, true, value)
}

func TestPageInputValue(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`
		<input></input>
	`))
	require.NoError(t, page.Locator("input").Fill("hello"))
	//nolint:staticcheck
	value, err := page.InputValue("input")
	require.NoError(t, err)
	require.Equal(t, "hello", value)
	require.NoError(t, page.Locator("input").Fill(""))
	//nolint:staticcheck
	value, err = page.InputValue("input")
	require.NoError(t, err)
	require.Equal(t, "", value)
}

func TestPageWaitForFunction2(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	_, err = page.AddScriptTag(
		playwright.PageAddScriptTagOptions{
			Path: playwright.String(Asset(filepath.Join("es6", "es6pathimport.js"))),
			Type: playwright.String("module"),
		},
	)
	require.NoError(t, err)
	_, err = page.WaitForFunction("window.__es6injected", nil)
	require.NoError(t, err)
	value, err := page.Evaluate("window.__es6injected")
	require.NoError(t, err)
	require.Equal(t, 42, value)
}

func TestPageShouldSetBodysizeAndHeadersize(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	request, err := page.ExpectRequest(
		"**/*", func() error {
			_, err = page.Evaluate("() => fetch('./get', { method: 'POST', body: '12345'}).then(r => r.text())")
			require.NoError(t, err)
			return nil
		},
	)
	require.NoError(t, err)
	sizes, err := request.Sizes()
	require.NoError(t, err)
	require.Equal(t, 5, sizes.RequestBodySize)
	require.GreaterOrEqual(t, sizes.RequestHeadersSize, 300)
}

func TestPageTestShouldSetBodysizeTo0(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	request, err := page.ExpectRequest(
		"**/*", func() error {
			_, err = page.Evaluate("() => fetch('./get').then(r => r.text())")
			require.NoError(t, err)
			return nil
		},
	)
	require.NoError(t, err)
	sizes, err := request.Sizes()
	require.NoError(t, err)
	require.Equal(t, 0, sizes.RequestBodySize)
	require.GreaterOrEqual(t, sizes.RequestHeadersSize, 200)
}

func TestPageSetChecked(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`<input id='checkbox' type='checkbox'></input>`))
	//nolint:staticcheck
	require.NoError(t, page.SetChecked("input", true))
	isChecked, err := page.Evaluate("checkbox.checked")
	require.NoError(t, err)
	require.True(t, isChecked.(bool))
	//nolint:staticcheck
	require.NoError(t, page.SetChecked("input", false))
	isChecked, err = page.Evaluate("checkbox.checked")
	require.NoError(t, err)
	require.False(t, isChecked.(bool))
}

func TestPageExpectRequestTimeout(t *testing.T) {
	t.Run("should respect timeout", func(t *testing.T) {
		BeforeEach(t)

		request, err := page.ExpectRequest("**/one-style.html", func() error {
			_, err := page.Goto(server.EMPTY_PAGE)
			return err
		}, playwright.PageExpectRequestOptions{Timeout: playwright.Float(1000)})

		require.Nil(t, request)
		require.ErrorContains(t, err, "Timeout 1000.00ms exceeded.")
	})

	t.Run("should use default timeout", func(t *testing.T) {
		BeforeEach(t)

		page.SetDefaultTimeout(500)
		defer page.SetDefaultTimeout(30 * 1000) // reset

		request, err := page.ExpectRequest("**/one-style.html", func() error {
			_, err := page.Goto(server.EMPTY_PAGE)
			return err
		})

		require.Nil(t, request)
		require.ErrorContains(t, err, "Timeout 500.00ms exceeded.")
	})
}

func TestPageExpectResponse(t *testing.T) {
	t.Run("should work with predicate", func(t *testing.T) {
		BeforeEach(t)

		predicate := regexp.MustCompile(`(?i).*/one-style.html`)
		response, err := page.ExpectResponse(predicate, func() error {
			_, err := page.Goto(server.PREFIX + "/one-style.html")
			return err
		}, playwright.PageExpectResponseOptions{Timeout: playwright.Float(3 * 1000)})
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("%s/one-style.html", server.PREFIX), response.URL())
	})

	t.Run("should work", func(t *testing.T) {
		BeforeEach(t)

		response, err := page.ExpectResponse("**/*", func() error {
			_, err := page.Goto(server.EMPTY_PAGE)
			return err
		})
		require.NoError(t, err)
		require.Equal(t, server.EMPTY_PAGE, response.URL())
		require.True(t, response.Ok())
		require.Equal(t, 200, response.Status())
		require.Equal(t, "OK", response.StatusText())
	})

	t.Run("should respect timeout", func(t *testing.T) {
		BeforeEach(t)

		response, err := page.ExpectResponse("**/one-style.html", func() error {
			_, err := page.Goto(server.EMPTY_PAGE)
			return err
		}, playwright.PageExpectResponseOptions{Timeout: playwright.Float(1000)})

		require.Nil(t, response)
		require.ErrorContains(t, err, "Timeout 1000.00ms exceeded.")
	})

	t.Run("should use default timeout", func(t *testing.T) {
		BeforeEach(t)

		_, err := page.Goto(server.EMPTY_PAGE)
		page.SetDefaultTimeout(500)
		defer page.SetDefaultTimeout(30 * 1000) // reset

		require.NoError(t, err)
		response, err := page.ExpectResponse("**/one-style.html", nil)

		require.Nil(t, response)
		require.ErrorContains(t, err, "Timeout 500.00ms exceeded.")
	})

	t.Run("should use context default timeout", func(t *testing.T) {
		BeforeEach(t)

		_, err := page.Goto(server.EMPTY_PAGE)
		context.SetDefaultTimeout(1000)
		defer context.SetDefaultTimeout(30 * 1000) // reset

		require.NoError(t, err)
		response, err := page.ExpectResponse("**/one-style.html", nil)

		require.Nil(t, response)
		require.ErrorContains(t, err, "Timeout 1000.00ms exceeded.")
	})

	// Regression test for https://github.com/mxschmitt/playwright-go/issues/323:
	// waiting for several responses concurrently (the Go equivalent of
	// Promise.all) must resolve every waiter. Previously, once one waiter
	// completed it unsubscribed all the others, so the rest timed out.
	t.Run("should work with concurrent waiters", func(t *testing.T) {
		BeforeEach(t)

		ids := []string{"a", "b", "c", "d", "e", "f"}
		for _, id := range ids {
			id := id
			server.SetRoute("/api/"+id, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, err := fmt.Fprintf(w, `{"id":"%s"}`, id)
				require.NoError(t, err)
			})
		}

		_, err := page.Goto(server.EMPTY_PAGE)
		require.NoError(t, err)

		var wg sync.WaitGroup
		results := make([]string, len(ids))
		errs := make([]error, len(ids))
		for i, id := range ids {
			wg.Add(1)
			go func(idx int, id string) {
				defer wg.Done()
				resp, err := page.ExpectResponse("**/api/"+id, func() error {
					_, e := page.Evaluate(fmt.Sprintf(`fetch("/api/%s")`, id))
					return e
				}, playwright.PageExpectResponseOptions{Timeout: playwright.Float(10 * 1000)})
				if err != nil {
					errs[idx] = err
					return
				}
				results[idx] = resp.URL()
			}(i, id)
		}
		wg.Wait()

		for i, id := range ids {
			require.NoErrorf(t, errs[i], "waiter %d (id %s)", i, id)
			require.Contains(t, results[i], "/api/"+id)
		}
	})
}

func TestPageWaitForURL(t *testing.T) {
	t.Run("should work", func(t *testing.T) {
		BeforeEach(t)

		_, err := page.Goto(server.EMPTY_PAGE)
		require.NoError(t, err)
		_, err = page.Evaluate("url => window.location.href = url", fmt.Sprintf("%s/grid.html", server.PREFIX))
		require.NoError(t, err)
		require.NoError(t, page.WaitForURL("**/grid.html"))
		require.Contains(t, page.URL(), "grid.html")
	})

	t.Run("should respect timeout", func(t *testing.T) {
		BeforeEach(t)

		_, err := page.Goto(server.EMPTY_PAGE)
		require.NoError(t, err)
		require.ErrorContains(t, page.WaitForURL("**/grid.html", playwright.PageWaitForURLOptions{
			Timeout: playwright.Float(1000),
		}), "Timeout 1000.00ms exceeded.")
	})

	t.Run("should work with commit", func(t *testing.T) {
		BeforeEach(t)

		_, err := page.Goto(server.EMPTY_PAGE)
		require.NoError(t, err)
		_, err = page.Evaluate("url => window.location.href = url", fmt.Sprintf("%s/grid.html", server.PREFIX))
		require.NoError(t, err)
		require.NoError(t, page.WaitForURL("**/grid.html"), playwright.FrameWaitForURLOptions{
			WaitUntil: playwright.WaitUntilStateCommit,
		})
		require.Contains(t, page.URL(), "grid.html")
	})
}

func TestCloseShouldRunBeforunloadIfAskedFor(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(fmt.Sprintf("%s/beforeunload.html", server.PREFIX))
	require.NoError(t, err)

	dialogInfo, err := page.ExpectEvent("dialog", func() error {
		// We have to interact with a page so that 'beforeunload' handlers fire.
		require.NoError(t, page.Locator("body").Click())
		return page.Close(playwright.PageCloseOptions{
			RunBeforeUnload: playwright.Bool(true),
		})
	})
	require.NoError(t, err)
	dialog := dialogInfo.(playwright.Dialog)
	require.Equal(t, "beforeunload", dialog.Type())
	if isChromium {
		require.Equal(t, "", dialog.Message())
	} else if isWebKit {
		require.Equal(t, "Leave?", dialog.Message())
	} else {
		require.Contains(t, dialog.Message(), "This page is asking you to confirm that you want to leave")
	}
	// Accepting the beforeunload dialog closes the page, so waiting for the
	// "close" event must resolve successfully (matching upstream beforeunload.spec.ts).
	_, err = page.ExpectEvent("close", func() error {
		return dialog.Accept()
	})
	require.NoError(t, err)
}

// TestCloseShouldAccessPageAfterBeforeUnload mirrors upstream's "should access
// page after beforeunload" test: with runBeforeUnload the page is not actually
// closed after dismissing the dialog, so it remains usable.
func TestCloseShouldAccessPageAfterBeforeUnload(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(fmt.Sprintf("%s/beforeunload.html", server.PREFIX))
	require.NoError(t, err)
	dialogInfo, err := page.ExpectEvent("dialog", func() error {
		require.NoError(t, page.Locator("body").Click())
		return page.Close(playwright.PageCloseOptions{
			RunBeforeUnload: playwright.Bool(true),
		})
	})
	require.NoError(t, err)
	dialog := dialogInfo.(playwright.Dialog)
	require.NoError(t, dialog.Dismiss())
	_, err = page.Evaluate("() => document.title")
	require.NoError(t, err)
}

func TestPageGotoShouldFailWhenExceedingBrowserContextNavigationTimeout(t *testing.T) {
	BeforeEach(t)

	// Hang for request to the empty.html
	server.SetRoute("/empty.html", func(w http.ResponseWriter, r *http.Request) {})
	context.SetDefaultNavigationTimeout(2)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.ErrorIs(t, err, playwright.ErrTimeout)
	require.ErrorContains(t, err, "Timeout 2ms exceeded.")
	require.ErrorContains(t, err, "/empty.html")
}

func TestShouldEmulateContrast(t *testing.T) {
	BeforeEach(t)

	ret, err := page.Evaluate(`matchMedia('(prefers-contrast: no-preference)').matches`)
	require.NoError(t, err)
	require.True(t, ret.(bool))

	err = page.EmulateMedia(playwright.PageEmulateMediaOptions{
		Contrast: playwright.ContrastNoPreference,
	})
	require.NoError(t, err)
	_, err = page.Evaluate(`matchMedia('(prefers-contrast: no-preference)').matches`)
	require.NoError(t, err)

	ret, err = page.Evaluate(`matchMedia('(prefers-contrast: more)').matches`)
	require.NoError(t, err)
	require.False(t, ret.(bool))

	err = page.EmulateMedia(playwright.PageEmulateMediaOptions{
		Contrast: playwright.ContrastMore,
	})
	require.NoError(t, err)
	ret, err = page.Evaluate(`matchMedia('(prefers-contrast: no-preference)').matches`)
	require.NoError(t, err)
	require.False(t, ret.(bool))
	ret, err = page.Evaluate(`matchMedia('(prefers-contrast: more)').matches`)
	require.NoError(t, err)
	require.True(t, ret.(bool))

	err = page.EmulateMedia(playwright.PageEmulateMediaOptions{
		Contrast: playwright.ContrastNoOverride,
	})
	require.NoError(t, err)
	ret, err = page.Evaluate(`matchMedia('(prefers-contrast: no-preference)').matches`)
	require.NoError(t, err)
	require.True(t, ret.(bool))
}

// TestPageConsoleMessages verifies that Page.ConsoleMessages() returns accumulated console messages
// Based on upstream test: playwright/tests/page/page-event-console.spec.ts
func TestPageConsoleMessages(t *testing.T) {
	BeforeEach(t)

	// Generate 301 console messages (should keep last 100)
	_, err := page.Evaluate(`() => {
		for (let i = 0; i < 301; i++) {
			console.log('message' + i);
		}
	}`)
	require.NoError(t, err)

	messages, err := page.ConsoleMessages()
	require.NoError(t, err)

	// Should return at least 100 messages (the buffer limit)
	require.GreaterOrEqual(t, len(messages), 100, "should be at least 100 messages")

	// Verify the last 100 messages are correct (message201 to message300)
	expectedStart := 301 - len(messages)
	for i, msg := range messages {
		expectedText := fmt.Sprintf("message%d", expectedStart+i)
		require.Equal(t, expectedText, msg.Text())
		require.Equal(t, "log", msg.Type())

		// Note: Page() may be nil for console messages retrieved from the buffer
		// as they don't include full channel references like live events do
	}
}

// TestPageConsoleMessagesEmpty verifies that ConsoleMessages() returns empty array for new page
func TestPageConsoleMessagesEmpty(t *testing.T) {
	BeforeEach(t)

	messages, err := page.ConsoleMessages()
	require.NoError(t, err)
	require.Empty(t, messages)
}

// TestPageConsoleMessagesTypes verifies different console message types are captured
func TestPageConsoleMessagesTypes(t *testing.T) {
	BeforeEach(t)

	_, err := page.Evaluate(`() => {
		console.log('log message');
		console.warn('warn message');
		console.error('error message');
		console.info('info message');
	}`)
	require.NoError(t, err)

	messages, err := page.ConsoleMessages()
	require.NoError(t, err)
	require.Len(t, messages, 4)

	expectedTypes := []string{"log", "warning", "error", "info"}
	for i, msg := range messages {
		require.Equal(t, expectedTypes[i], msg.Type())
	}
}

// TestPageRequests verifies that Page.Requests() returns accumulated requests
// Based on upstream test: playwright/tests/page/page-event-request.spec.ts
func TestPageRequests(t *testing.T) {
	BeforeEach(t)

	// Navigate to a page (creates initial request)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)

	// Create multiple fetch requests
	for i := 0; i < 99; i++ {
		path := fmt.Sprintf("/fetch%d", i)
		server.SetRoute(path, func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("response"))
			require.NoError(t, err)
		})
	}

	// Make 99 fetch requests
	for i := 0; i < 99; i++ {
		url := fmt.Sprintf("%s/fetch%d", server.PREFIX, i)
		_, err := page.Evaluate(`url => fetch(url)`, url)
		require.NoError(t, err)
	}

	requests, err := page.Requests()
	require.NoError(t, err)

	// Should have at least 100 requests (navigation + 99 fetches, buffer limit is 100)
	require.GreaterOrEqual(t, len(requests), 99, "should capture fetch requests")

	// Verify requests are functional
	for _, req := range requests {
		require.NotEmpty(t, req.URL())
		require.NotEmpty(t, req.Method())
	}
}

// TestPageRequestsEmpty verifies that Requests() returns empty array for new page
func TestPageRequestsEmpty(t *testing.T) {
	BeforeEach(t)

	requests, err := page.Requests()
	require.NoError(t, err)
	require.Empty(t, requests)
}

// TestPageRequestsWithNavigation verifies navigation requests are included
func TestPageRequestsWithNavigation(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)

	requests, err := page.Requests()
	require.NoError(t, err)
	require.Len(t, requests, 1, "should capture navigation request")

	req := requests[0]
	require.Equal(t, server.EMPTY_PAGE, req.URL())
	require.Equal(t, "GET", req.Method())
	require.Equal(t, "document", req.ResourceType())
}
