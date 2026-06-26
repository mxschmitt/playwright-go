package playwright_test

import (
	"sync"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestScreencastOnFrameReceivesViewportSizeAndTimestamp(t *testing.T) {
	BeforeEach(t)

	ctx, err := browser.NewContext(playwright.BrowserNewContextOptions{
		Viewport: &playwright.Size{Width: 1000, Height: 400},
	})
	require.NoError(t, err)
	defer ctx.Close() //nolint:errcheck
	p, err := ctx.NewPage()
	require.NoError(t, err)

	var (
		mu       sync.Mutex
		received []playwright.OnFrame
	)
	sc, err := p.Screencast()
	require.NoError(t, err)
	require.NoError(t, sc.Start(playwright.ScreencastStartOptions{
		Size: &playwright.Size{Width: 500, Height: 400},
		OnFrame: func(frame playwright.OnFrame) {
			mu.Lock()
			received = append(received, frame)
			mu.Unlock()
		},
	}))

	_, err = p.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	_, err = p.Evaluate("() => document.body.style.backgroundColor = 'red'")
	require.NoError(t, err)
	for i := 0; i < 100; i++ {
		_, err = p.Evaluate("() => new Promise(f => requestAnimationFrame(() => requestAnimationFrame(f)))")
		require.NoError(t, err)
	}
	_, err = p.Screenshot()
	require.NoError(t, err)
	require.NoError(t, sc.Stop())

	mu.Lock()
	defer mu.Unlock()
	require.GreaterOrEqual(t, len(received), 1)
	for _, frame := range received {
		require.Equal(t, 400, frame.ViewportHeight)
		// Timestamp is milliseconds since the Unix epoch; just assert it was set.
		require.Greater(t, frame.Timestamp, float64(0))
	}
}

func TestScreencastShowActionsAcceptsCursorParam(t *testing.T) {
	BeforeEach(t)

	sc, err := page.Screencast()
	require.NoError(t, err)
	require.NoError(t, sc.Start(playwright.ScreencastStartOptions{
		OnFrame: func(playwright.OnFrame) {},
	}))
	defer sc.Stop() //nolint:errcheck

	require.NoError(t, sc.ShowActions(playwright.ScreencastShowActionsOptions{
		Duration: playwright.Float(100),
		Cursor:   playwright.ScreencastCursorPointer,
	}))
	require.NoError(t, sc.ShowActions(playwright.ScreencastShowActionsOptions{
		Duration: playwright.Float(100),
		Cursor:   playwright.ScreencastCursorNone,
	}))
}
