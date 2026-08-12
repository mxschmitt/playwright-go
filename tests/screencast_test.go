package playwright_test

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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

	// Drive distinct visual mutations until at least two frames arrive.
	// Without screencastFrameAck the stream stalls after the first frame.
	colors := []string{"red", "green", "blue", "yellow", "purple"}
	deadline := time.Now().Add(10 * time.Second)
	for i := 0; time.Now().Before(deadline); i++ {
		color := colors[i%len(colors)]
		_, err = p.Evaluate(fmt.Sprintf("() => { document.body.style.backgroundColor = '%s'; }", color))
		require.NoError(t, err)
		_, err = p.Evaluate("() => new Promise(f => requestAnimationFrame(() => requestAnimationFrame(f)))")
		require.NoError(t, err)
		mu.Lock()
		n := len(received)
		mu.Unlock()
		if n >= 2 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	require.NoError(t, sc.Stop())

	mu.Lock()
	defer mu.Unlock()
	require.GreaterOrEqual(t, len(received), 2, "expected multi-frame screencast; ACK may be missing")
	for _, frame := range received {
		require.GreaterOrEqual(t, len(frame.Data), 2)
		require.Equal(t, byte(0xff), frame.Data[0])
		require.Equal(t, byte(0xd8), frame.Data[1])
		require.Equal(t, 1000, frame.ViewportWidth)
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

func TestScreencastOnFrameCanReenterClient(t *testing.T) {
	BeforeEach(t)

	sc, err := page.Screencast()
	require.NoError(t, err)
	result := make(chan error, 1)
	var once sync.Once
	var callbackCount atomic.Int32
	require.NoError(t, sc.Start(playwright.ScreencastStartOptions{
		OnFrame: func(playwright.OnFrame) {
			_, callbackErr := page.Evaluate(`() => document.title`)
			callbackCount.Add(1)
			once.Do(func() { result <- callbackErr })
		},
	}))
	defer sc.Stop() //nolint:errcheck

	require.NoError(t, page.SetContent(`<title>reentrant</title><body>first</body>`))
	_, err = page.Evaluate(`() => { document.body.textContent = "second" }`)
	require.NoError(t, err)
	select {
	case callbackErr := <-result:
		require.NoError(t, callbackErr)
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for a reentrant screencast callback")
	}

	_, err = page.Evaluate(`() => { document.body.style.backgroundColor = "blue" }`)
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		return callbackCount.Load() > 1
	}, 10*time.Second, 20*time.Millisecond, "expected another frame after the reentrant callback completed")
}

func TestScreencastAppliesBackpressureWhileOnFrameCallbackPending(t *testing.T) {
	BeforeEach(t)

	releaseCallback := make(chan struct{})
	firstFrame := make(chan struct{})
	var firstFrameOnce sync.Once
	var callbackCount atomic.Int32
	var lastFrameAt atomic.Int64

	sc, err := page.Screencast()
	require.NoError(t, err)
	require.NoError(t, sc.Start(playwright.ScreencastStartOptions{
		OnFrame: func(playwright.OnFrame) {
			callbackCount.Add(1)
			lastFrameAt.Store(time.Now().UnixNano())
			firstFrameOnce.Do(func() { close(firstFrame) })
			<-releaseCallback
		},
	}))

	released := false
	stopped := false
	defer func() {
		if !released {
			close(releaseCallback)
		}
		if !stopped {
			_ = sc.Stop()
		}
	}()

	require.NoError(t, page.SetContent(`<body></body><script>
		const animate = () => {
			document.body.style.backgroundColor = document.body.style.backgroundColor === "red" ? "blue" : "red";
			requestAnimationFrame(animate);
		};
		requestAnimationFrame(animate);
	</script>`))
	select {
	case <-firstFrame:
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for the first screencast frame")
	}
	// A small number of frames can already be in flight when the first callback
	// starts. Wait until that initial queue has drained before taking a baseline.
	require.Eventually(t, func() bool {
		last := lastFrameAt.Load()
		return last != 0 && time.Since(time.Unix(0, last)) > time.Second
	}, 10*time.Second, 20*time.Millisecond, "screencast frames did not quiesce while the callback was blocked")

	framesWhileBlocked := callbackCount.Load()
	blockedUntil := time.Now().Add(1200 * time.Millisecond)
	for time.Now().Before(blockedUntil) {
		_, err = page.Evaluate(`() => new Promise(f => requestAnimationFrame(() => requestAnimationFrame(f)))`)
		require.NoError(t, err)
		require.Equal(t, framesWhileBlocked, callbackCount.Load(), "a new frame arrived before the callback returned")
	}
	require.Equal(t, framesWhileBlocked, callbackCount.Load(), "a new frame arrived before the callback returned")

	close(releaseCallback)
	released = true
	require.Eventually(t, func() bool {
		_, evaluateErr := page.Evaluate(`() => {
			document.body.style.backgroundColor = document.body.style.backgroundColor === "red" ? "blue" : "red";
		}`)
		return evaluateErr == nil && callbackCount.Load() > framesWhileBlocked
	}, 10*time.Second, 20*time.Millisecond, "expected frame delivery to resume after the callback returned")

	require.NoError(t, sc.Stop())
	stopped = true
}

func TestScreencastReportsCallbackPanicsAndContinuesDeliveringFrames(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`<body></body><script>
		const animate = () => {
			document.body.style.backgroundColor = document.body.style.backgroundColor === "red" ? "blue" : "red";
			requestAnimationFrame(animate);
		};
		requestAnimationFrame(animate);
	</script>`))

	sc, err := page.Screencast()
	require.NoError(t, err)
	var callbackCount atomic.Int32
	require.NoError(t, sc.Start(playwright.ScreencastStartOptions{
		OnFrame: func(playwright.OnFrame) {
			if callbackCount.Add(1) == 1 {
				panic(errors.New("screencast callback failed"))
			}
		},
	}))
	stopped := false
	defer func() {
		if !stopped {
			_ = sc.Stop()
		}
	}()

	var callbackErr error
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		_, evaluateErr := page.Evaluate(`() => new Promise(f => requestAnimationFrame(() => requestAnimationFrame(f)))`)
		if evaluateErr != nil {
			if strings.Contains(evaluateErr.Error(), "screencast callback failed") {
				callbackErr = evaluateErr
			} else {
				require.NoError(t, evaluateErr)
			}
		}
		if callbackErr != nil && callbackCount.Load() > 1 {
			break
		}
	}
	require.ErrorContains(t, callbackErr, "screencast callback failed")
	require.Greater(t, callbackCount.Load(), int32(1))

	require.NoError(t, sc.Stop())
	stopped = true
}
