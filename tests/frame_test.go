package playwright_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestFrameWaitForNavigationShouldWork(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	//nolint:staticcheck
	response, err := page.ExpectNavigation(func() error {
		_, err := page.Evaluate("url => window.location.href = url", server.PREFIX+"/grid.html")
		return err
	})
	require.NoError(t, err)
	require.True(t, response.Ok())
	require.Contains(t, response.URL(), "grid.html")
}

func TestFrameWaitForNavigationShouldRespectTimeout(t *testing.T) {
	BeforeEach(t)

	timeout := 500.0
	//nolint:staticcheck
	_, err := page.ExpectNavigation(func() error {
		_, err := page.Evaluate("url => window.location.href = url", server.EMPTY_PAGE)
		return err
	}, playwright.PageExpectNavigationOptions{
		URL:     "**/frame.html",
		Timeout: playwright.Float(timeout),
	})
	require.ErrorContains(t, err, fmt.Sprintf(`Timeout %.2fms exceeded.`, timeout))
}

func TestFrameWaitForURLShouldWork(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)

	require.NoError(t, page.SetContent(`<a href="grid.html">foobar</a>`))
	go func() {
		time.Sleep(2 * time.Second)
		require.NoError(t, page.Locator("a").Click())
	}()

	err = page.MainFrame().WaitForURL(server.PREFIX + "/grid.html")
	require.NoError(t, err)
	require.Equal(t, server.PREFIX+"/grid.html", page.URL())
}

func TestFrameWaitForNavigationAnchorLinks(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`<a href="#foobar">foobar</a>`))
	//nolint:staticcheck
	response, err := page.ExpectNavigation(func() error {
		return page.Locator("a").Click()
	})
	require.NoError(t, err)
	require.Nil(t, response)
	require.Equal(t, server.EMPTY_PAGE+"#foobar", page.URL())
}

func TestFrameInnerHTML(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.PREFIX + "/dom.html")
	require.NoError(t, err)
	//nolint:staticcheck
	handle, err := page.QuerySelector("#outer")
	require.NoError(t, err)
	//nolint:staticcheck
	innerHTML, err := handle.InnerHTML()
	require.NoError(t, err)
	require.Equal(t, `<div id="inner">Text,
more text</div>`, innerHTML)
	//nolint:staticcheck
	innerHTML, err = page.InnerHTML("#outer")
	require.NoError(t, err)
	require.Equal(t, `<div id="inner">Text,
more text</div>`, innerHTML)
}

func TestFrameSetInputFiles(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent("<input type=file>"))
	//nolint:staticcheck
	require.NoError(t, page.SetInputFiles("input", []playwright.InputFile{
		{
			Name:     "file-to-upload.txt",
			MimeType: "text/plain",
			Buffer:   []byte("123"),
		},
	}))
	fileName, err := page.Locator("input").Evaluate("e => e.files[0].name", nil)
	require.NoError(t, err)
	require.Equal(t, "file-to-upload.txt", fileName)
}

func TestShouldReportDifferentFrameInstanceWhenFrameReattaches(t *testing.T) {
	BeforeEach(t)

	frame1, err := utils.AttachFrame(page, "frame1", server.EMPTY_PAGE)
	require.NoError(t, err)

	_, err = page.Evaluate(`() => {
			window.frame = document.querySelector('#frame1')
			window.frame.remove()
		}`)
	require.NoError(t, err)

	require.True(t, frame1.IsDetached())
	ret, err := page.ExpectEvent("frameattached", func() error {
		_, err := page.Evaluate(`() => document.body.appendChild(window.frame)`)
		return err
	})
	require.NoError(t, err)
	frame2 := ret.(playwright.Frame)
	require.False(t, frame2.IsDetached())
	require.NotEqual(t, frame1, frame2)
}

func TestShouldSendEventsWhenFramesAreManipulatedDynamically(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	// validate frameattached events
	attachedFrames := []playwright.Frame{}
	page.OnFrameAttached(func(frame playwright.Frame) {
		attachedFrames = append(attachedFrames, frame)
	})
	_, err = utils.AttachFrame(page, "frame1", "./assets/frame.html")
	require.NoError(t, err)
	require.Len(t, attachedFrames, 1)
	require.Contains(t, attachedFrames[0].URL(), "/assets/frame.html")

	// validate framenavigated evnents
	navigatedFrames := []playwright.Frame{}
	page.OnFrameNavigated(func(frame playwright.Frame) {
		navigatedFrames = append(navigatedFrames, frame)
	})
	_, err = page.Evaluate(`() => {
			frame = document.getElementById('frame1')
			frame.src = './empty.html'
			return new Promise(x => frame.onload = x)
		}`)
	require.NoError(t, err)
	require.Len(t, navigatedFrames, 1)
	require.Equal(t, navigatedFrames[0].URL(), server.EMPTY_PAGE)

	// validate framedetached events
	detachedFrames := []playwright.Frame{}
	page.OnFrameDetached(func(frame playwright.Frame) {
		detachedFrames = append(detachedFrames, frame)
	})
	require.NoError(t, utils.DetachFrame(page, "frame1"))
	require.Len(t, detachedFrames, 1)
	require.True(t, detachedFrames[0].IsDetached())
}

func TestFrameElement(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	frame1, err := utils.AttachFrame(page, "frame1", server.EMPTY_PAGE)
	require.NoError(t, err)
	_, err = utils.AttachFrame(page, "frame2", server.EMPTY_PAGE)
	require.NoError(t, err)
	frame3, err := utils.AttachFrame(page, "frame3", server.EMPTY_PAGE)
	require.NoError(t, err)
	//nolint:staticcheck
	frame1Handle1, err := page.QuerySelector("#frame1")
	require.NoError(t, err)
	frame1Handle2, err := frame1.FrameElement()
	require.NoError(t, err)
	//nolint:staticcheck
	frame3Handle1, err := page.QuerySelector("#frame3")
	require.NoError(t, err)
	frame3Handle2, err := frame3.FrameElement()
	require.NoError(t, err)
	ret, err := frame1Handle1.Evaluate(`(a, b) => a === b`, frame1Handle2)
	require.NoError(t, err)
	require.True(t, ret.(bool))
	ret, err = frame3Handle1.Evaluate(`(a, b) => a === b`, frame3Handle2)
	require.NoError(t, err)
	require.True(t, ret.(bool))
	ret, err = frame1Handle1.Evaluate(`(a, b) => a === b`, frame3Handle1)
	require.NoError(t, err)
	require.False(t, ret.(bool))
}

func TestFrameParent(t *testing.T) {
	BeforeEach(t)

	_, err := utils.AttachFrame(page, "frame1", server.EMPTY_PAGE)
	require.NoError(t, err)
	_, err = utils.AttachFrame(page, "frame2", server.EMPTY_PAGE)
	require.NoError(t, err)
	frames := page.Frames()
	require.Len(t, frames, 3)
	require.Nil(t, frames[0].ParentFrame())
	require.Equal(t, page.MainFrame(), frames[1].ParentFrame())
	require.Equal(t, page.MainFrame(), frames[2].ParentFrame())
}

func TestFrameShouldHandleNestedFrames(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.PREFIX + "/frames/nested-frames.html")
	require.NoError(t, err)
	dump := utils.DumpFrames(page.MainFrame(), "")
	require.Equal(t, []string{
		"http://localhost:<PORT>/frames/nested-frames.html",
		"    http://localhost:<PORT>/frames/frame.html (aframe)",
		"    http://localhost:<PORT>/frames/two-frames.html (2frames)",
		"        http://localhost:<PORT>/frames/frame.html (dos)",
		"        http://localhost:<PORT>/frames/frame.html (uno)",
	}, dump)
}
