package playwright_test

import "testing"

// TestPageFramesRace reproduces a data race between Page.Frames()/Page.Frame()
// and frame attach/detach events. The race detector is the assertion.
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
				for _, frame := range page.Frames() {
					_ = frame.URL()
				}
				_ = page.Frame()
			}
		}
	}()
	defer func() {
		close(done)
		<-stopped
	}()

	for i := 0; i < 200; i++ {
		_, err := page.Evaluate(`() => {
			const frame = document.createElement('iframe');
			frame.src = 'about:blank';
			document.body.appendChild(frame);
			frame.remove();
		}`)
		if err != nil {
			t.Fatal(err)
		}
	}
}
