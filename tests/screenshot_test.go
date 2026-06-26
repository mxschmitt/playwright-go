package playwright_test

import (
	"path/filepath"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestLocatorScreenshotShouldWork(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetViewportSize(500, 500))
	_, err := page.Goto(server.PREFIX + "/grid.html")
	require.NoError(t, err)
	_, err = page.Evaluate(`window.scrollBy(50, 100)`)
	require.NoError(t, err)
	screenshot, err := page.Locator(".box:nth-of-type(3)").Screenshot()
	require.NoError(t, err)
	require.NotEmpty(t, screenshot)
	AssertToBeGolden(t, screenshot, "screenshot-element-bounding-box.png")
}

// TestScreenshotPathOptionShouldDetectJpeg mirrors upstream's "path option
// should detect jpeg" test (tests/page/page-screenshot.spec.ts): a .jpg path
// without an explicit type must produce a JPEG (FF D8 FF magic bytes).
func TestScreenshotPathOptionShouldDetectJpeg(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetViewportSize(300, 300))
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	outputPath := filepath.Join(t.TempDir(), "screenshot.jpg")
	screenshot, err := page.Screenshot(playwright.PageScreenshotOptions{
		OmitBackground: playwright.Bool(true),
		Path:           playwright.String(outputPath),
	})
	require.NoError(t, err)
	require.Equal(t, []byte{0xFF, 0xD8, 0xFF}, screenshot[:3])
}

// TestScreenshotPathOptionShouldThrowForUnsupportedMimeType mirrors upstream's
// "path option should throw for unsupported mime type" test.
func TestScreenshotPathOptionShouldThrowForUnsupportedMimeType(t *testing.T) {
	BeforeEach(t)

	_, err := page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String("file.txt"),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), `unsupported mime type "text/plain"`)
}

// TestScreenshotShouldPreferTypeOverExtension mirrors upstream's "should prefer
// type over extension" test: a .png path with an explicit jpeg type yields JPEG.
func TestScreenshotShouldPreferTypeOverExtension(t *testing.T) {
	BeforeEach(t)

	outputPath := filepath.Join(t.TempDir(), "file.png")
	screenshot, err := page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String(outputPath),
		Type: playwright.ScreenshotTypeJpeg,
	})
	require.NoError(t, err)
	require.Equal(t, []byte{0xFF, 0xD8, 0xFF}, screenshot[:3])
}

func TestShouldScreenshotWithMask(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetViewportSize(500, 500))
	_, err := page.Goto(server.PREFIX + "/grid.html")
	require.NoError(t, err)

	screenshot, err := page.Screenshot(playwright.PageScreenshotOptions{
		Mask: []playwright.Locator{
			page.Locator("div").Nth(5),
		},
	})
	require.NoError(t, err)
	AssertToBeGolden(t, screenshot, "mask-should-work.png")

	screenshot, err = page.Locator("body").Screenshot(playwright.LocatorScreenshotOptions{
		Mask: []playwright.Locator{
			page.Locator("div").Nth(5),
		},
	})
	require.NoError(t, err)
	AssertToBeGolden(t, screenshot, "mask-should-work-with-locator.png")

	//nolint:staticcheck
	element, err := page.QuerySelector("body")
	require.NoError(t, err)
	//nolint:staticcheck
	screenshot, err = element.Screenshot(playwright.ElementHandleScreenshotOptions{
		Mask: []playwright.Locator{
			page.Locator("div").Nth(5),
		},
	})
	require.NoError(t, err)
	AssertToBeGolden(t, screenshot, "mask-should-work-with-elementhandle.png")
}
