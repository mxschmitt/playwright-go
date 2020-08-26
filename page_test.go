package playwright

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/h2non/filetype"
	"github.com/stretchr/testify/require"
)

func TestPageURL(t *testing.T) {
	helper := NewTestHelper(t)
	require.Equal(t, "about:blank", helper.Page.URL())
	require.NoError(t, helper.Page.Goto("https://example.com"))
	require.Equal(t, "https://example.com/", helper.Page.URL())
	helper.Close(t)
}

func TestPageSetContent(t *testing.T) {
	helper := NewTestHelper(t)
	require.NoError(t, helper.Page.SetContent("<h1>foo</h1>",
		PageSetContentOptions{
			WaitUntil: String("networkidle"),
		}))
	content, err := helper.Page.Content()
	require.NoError(t, err)
	require.Equal(t, content, "<html><head></head><body><h1>foo</h1></body></html>")
	helper.Close(t)
}

func TestPageScreenshot(t *testing.T) {
	helper := NewTestHelper(t)

	require.NoError(t, helper.Page.SetContent("<h1>foobar</h1>"))
	tmpfile, err := ioutil.TempDir("", "screenshot")
	require.NoError(t, err)
	screenshotPath := filepath.Join(tmpfile, "image.png")
	screenshot, err := helper.Page.Screenshot()
	require.NoError(t, err)
	require.True(t, filetype.IsImage(screenshot))
	require.Greater(t, len(screenshot), 50)

	screenshot, err = helper.Page.Screenshot(PageScreenshotOptions{
		Path: String(screenshotPath),
	})
	require.NoError(t, err)
	require.True(t, filetype.IsImage(screenshot))
	require.Greater(t, len(screenshot), 50)

	_, err = os.Stat(screenshotPath)
	require.NoError(t, err)
	helper.Close(t)
}

func TestPagePDF(t *testing.T) {
	helper := NewTestHelper(t)
	if !helper.IsChromium {
		helper.Close(t)
		t.Skip("Skipping")
	}
	require.NoError(t, helper.Page.SetContent("<h1>foobar</h1>"))
	tmpfile, err := ioutil.TempDir("", "pdf")
	require.NoError(t, err)
	screenshotPath := filepath.Join(tmpfile, "image.png")
	screenshot, err := helper.Page.PDF()
	require.NoError(t, err)
	require.Equal(t, "application/pdf", http.DetectContentType(screenshot))
	require.Greater(t, len(screenshot), 50)

	screenshot, err = helper.Page.PDF(PagePdfOptions{
		Path: String(screenshotPath),
	})
	require.NoError(t, err)
	require.Equal(t, "application/pdf", http.DetectContentType(screenshot))
	require.Greater(t, len(screenshot), 50)

	_, err = os.Stat(screenshotPath)
	require.NoError(t, err)
	helper.Close(t)
}

func TestPageQuerySelector(t *testing.T) {
	helper := NewTestHelper(t)
	require.NoError(t, helper.Page.SetContent(`<div id="one">
		<span id="two">
			<div id="three">
				<span id="four">
					foobar
				</span>
			</div>
		</span>
	</div>`))
	one, err := helper.Page.QuerySelector("div#one")
	require.NoError(t, err)
	two, err := one.QuerySelector("span#two")
	require.NoError(t, err)
	three, err := two.QuerySelector("div#three")
	require.NoError(t, err)
	four, err := three.QuerySelector("span#four")
	require.NoError(t, err)
	textContent, err := four.TextContent()
	require.NoError(t, err)
	require.Equal(t, strings.TrimSpace(textContent), "foobar")
	helper.Close(t)
}

func TestPageEvaluate(t *testing.T) {
	helper := NewTestHelper(t)
	val, err := helper.Page.Evaluate(`() => 123`)
	require.NoError(t, err)
	require.Equal(t, val, 123)
	helper.Close(t)
}

func TestPageEvaluateOnSelectorAll(t *testing.T) {
	helper := NewTestHelper(t)
	err := helper.Page.SetContent(`
		<div class="foo">1</div>
		<div class="foo">2</div>
		<div class="foo">3</div>
	`)
	require.NoError(t, err)
	val, err := helper.Page.EvaluateOnSelectorAll(".foo", `(elements) => elements.map(el => el.textContent)`)
	require.NoError(t, err)
	require.Equal(t, val, []interface{}([]interface{}{"1", "2", "3"}))
	helper.Close(t)
}

func TestPageEvaluateOnSelector(t *testing.T) {
	helper := NewTestHelper(t)
	err := helper.Page.SetContent(`
		<div class="foo">bar</div>
	`)
	require.NoError(t, err)
	val, err := helper.Page.EvaluateOnSelector(".foo", `(element) => element.textContent`)
	require.NoError(t, err)
	require.Equal(t, val, "bar")
	helper.Close(t)
}
