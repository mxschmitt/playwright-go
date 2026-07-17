package playwright_test

import (
	"testing"
)

// TestPageFramesRace reproduces a data race between Page.Frames()/Page.Frame()
// (called from a user goroutine) and onFrameAttached/onFrameDetached (called
// from the connection dispatch goroutine as iframes are added/removed).
//
// Before the fix this fails under `go test -race` with multiple
// "WARNING: DATA RACE" reports on pageImpl.frames; there are no assertions
// here beyond that because the race detector itself is the check.
func TestPageFramesRace(t *testing.T) {
	BeforeEach(t)

	done := make(chan struct{})
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		for {
			select {
			case <-done:
				return
			default:
				for _, f := range page.Frames() {
					_ = f.URL()
				}
				_ = page.Frame()
			}
		}
	}()

	for i := 0; i < 200; i++ {
		_, err := page.Evaluate(`() => {
			const f = document.createElement('iframe');
			f.src = 'about:blank';
			document.body.appendChild(f);
			f.remove();
		}`)
		if err != nil {
			t.Fatal(err)
		}
	}

	close(done)
	<-stopped
}
