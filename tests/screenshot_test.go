package playwright_test

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestScreenshotShouldWorkWhileNavigating(t *testing.T) {
	BeforeEach(t)

	server.SetRoute("/redirectloop1.html", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<script>
			setTimeout(() => {
				const iteration = +(window.localStorage.iteration || 0);
				window.localStorage.iteration = iteration + 1;
				if (iteration < 10)
					window.location.href = "/redirectloop2.html";
			}, 1);
		</script>`))
	})
	server.SetRoute("/redirectloop2.html", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<script>setTimeout(() => window.location.href = "/redirectloop1.html", 1);</script>`))
	})

	require.NoError(t, page.SetViewportSize(500, 500))
	_, err := page.Goto(server.PREFIX + "/redirectloop1.html")
	require.NoError(t, err)
	successfulScreenshots := 0
	for range 10 {
		screenshot, err := page.Screenshot(playwright.PageScreenshotOptions{FullPage: playwright.Bool(true)})
		if err != nil && strings.Contains(err.Error(), "Cannot take a screenshot while page is navigating") {
			continue
		}
		require.NoError(t, err)
		require.NotNil(t, screenshot)
		successfulScreenshots++
	}
	require.Positive(t, successfulScreenshots, "all screenshot attempts failed while the page was navigating")
}

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

func TestScreenshotWebPTypeAndPath(t *testing.T) {
	BeforeEach(t)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`<div style="width:100px;height:100px;background:red"></div>`))

	// Explicit type.
	data, err := page.Screenshot(playwright.PageScreenshotOptions{
		Type: playwright.ScreenshotTypeWebp,
	})
	require.NoError(t, err)
	require.True(t, isWebP(data), "expected RIFF/WEBP signature")

	// Path inference.
	dir := t.TempDir()
	path := filepath.Join(dir, "shot.webp")
	_, err = page.Screenshot(playwright.PageScreenshotOptions{Path: playwright.String(path)})
	require.NoError(t, err)
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	require.True(t, isWebP(raw))

	// Quality variants.
	high, err := page.Screenshot(playwright.PageScreenshotOptions{
		Type:    playwright.ScreenshotTypeWebp,
		Quality: playwright.Int(100),
	})
	require.NoError(t, err)
	low, err := page.Screenshot(playwright.PageScreenshotOptions{
		Type:    playwright.ScreenshotTypeWebp,
		Quality: playwright.Int(1),
	})
	require.NoError(t, err)
	require.True(t, isWebP(high))
	require.True(t, isWebP(low))
	// Both quality settings must produce valid, non-empty WebP payloads.
	// Relative size is encoder-dependent for solid-color images, so we do not
	// assert low < high here (Firefox/libwebp may invert that for tiny frames).
	require.NotEmpty(t, high)
	require.NotEmpty(t, low)
}

func TestWebPScreenshotLocatorAndElementHandleHonorQuality(t *testing.T) {
	BeforeEach(t)
	require.NoError(t, page.SetContent(`<canvas id="target" width="192" height="192"></canvas>`))
	_, err := page.Evaluate(`() => {
		const canvas = document.querySelector('#target');
		const context = canvas.getContext('2d');
		const image = context.createImageData(canvas.width, canvas.height);
		for (let i = 0; i < image.data.length; i += 4) {
			const pixel = i / 4;
			image.data[i] = (pixel * 17) % 256;
			image.data[i + 1] = (pixel * 31) % 256;
			image.data[i + 2] = (pixel * 47) % 256;
			image.data[i + 3] = 255;
		}
		context.putImageData(image, 0, 0);
	}`)
	require.NoError(t, err)

	locator := page.Locator("#target")
	locatorLossless, err := locator.Screenshot(playwright.LocatorScreenshotOptions{
		Type:    playwright.ScreenshotTypeWebp,
		Quality: playwright.Int(100),
	})
	require.NoError(t, err)
	locatorLossy, err := locator.Screenshot(playwright.LocatorScreenshotOptions{
		Type:    playwright.ScreenshotTypeWebp,
		Quality: playwright.Int(1),
	})
	require.NoError(t, err)
	require.True(t, isWebP(locatorLossless))
	require.True(t, isWebP(locatorLossy))
	require.NotEqual(t, locatorLossless, locatorLossy, "WebP quality must reach the driver")

	//nolint:staticcheck // The new WebP format is explicitly exposed on the legacy ElementHandle API.
	handle, err := page.QuerySelector("#target")
	require.NoError(t, err)
	require.NotNil(t, handle)
	defer handle.Dispose() //nolint:errcheck
	//nolint:staticcheck // Required regression coverage for the ElementHandle API surface.
	handleLossless, err := handle.Screenshot(playwright.ElementHandleScreenshotOptions{
		Type:    playwright.ScreenshotTypeWebp,
		Quality: playwright.Int(100),
	})
	require.NoError(t, err)
	//nolint:staticcheck // Required regression coverage for the ElementHandle API surface.
	handleLossy, err := handle.Screenshot(playwright.ElementHandleScreenshotOptions{
		Type:    playwright.ScreenshotTypeWebp,
		Quality: playwright.Int(1),
	})
	require.NoError(t, err)
	require.True(t, isWebP(handleLossless))
	require.True(t, isWebP(handleLossy))
	require.NotEqual(t, handleLossless, handleLossy, "WebP quality must reach the driver")
}

func isWebP(data []byte) bool {
	if len(data) < 12 {
		return false
	}
	// RIFF....WEBP
	return string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP"
}
