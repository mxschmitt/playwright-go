package playwright

import (
	"encoding/json"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// captureTransport records messages sent by the connection without a real driver.
type captureTransport struct {
	mu       sync.Mutex
	messages []map[string]any
	// replies maps request id -> result payload delivered on Poll.
	replies chan map[string]any
	closed  chan struct{}
}

// failingSendTransport deterministically exercises the error path after a
// protocol callback has been registered but before a request reaches a driver.
type failingSendTransport struct {
	err error
}

func (t *failingSendTransport) Send(map[string]any) error {
	return t.err
}

func (*failingSendTransport) Poll() (*message, error) {
	return nil, ErrTargetClosed
}

func (*failingSendTransport) Close() error {
	return nil
}

func newCaptureTransport() *captureTransport {
	return &captureTransport{
		replies: make(chan map[string]any, 16),
		closed:  make(chan struct{}),
	}
}

func (t *captureTransport) Send(message map[string]any) error {
	t.mu.Lock()
	// Deep-ish copy via JSON so later mutations don't affect the capture.
	b, _ := json.Marshal(message)
	var copy map[string]any
	_ = json.Unmarshal(b, &copy)
	t.messages = append(t.messages, copy)
	t.mu.Unlock()
	// Auto-reply with empty result so SendWithTimeout unblocks.
	id, _ := copy["id"].(float64)
	t.replies <- map[string]any{"id": id, "result": map[string]any{}}
	return nil
}

func (t *captureTransport) Poll() (*message, error) {
	select {
	case msg := <-t.replies:
		return &message{
			ID:     int(msg["id"].(float64)),
			Result: msg["result"].(map[string]any),
		}, nil
	case <-t.closed:
		return nil, ErrTargetClosed
	}
}

func (t *captureTransport) Close() error {
	select {
	case <-t.closed:
	default:
		close(t.closed)
	}
	return nil
}

func (t *captureTransport) last() map[string]any {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.messages) == 0 {
		return nil
	}
	return t.messages[len(t.messages)-1]
}

func newTestConnection(t *testing.T) (*connection, *captureTransport, *channelOwner) {
	t.Helper()
	tr := newCaptureTransport()
	conn := newConnection(tr)
	// Create a root-like owner so Send has a guid to target.
	owner := &channelOwner{
		guid:       "test-guid",
		connection: conn,
		objects:    map[string]*channelOwner{},
	}
	owner.channel = newChannel(owner, owner)
	conn.objects.Store(owner.guid, owner)
	// Start the receive loop so blocking Sends complete.
	go func() {
		for {
			if !conn.pollOnce() {
				return
			}
		}
	}()
	t.Cleanup(func() { _ = tr.Close() })
	return conn, tr, owner
}

func TestFailedSendDoesNotOrphanProtocolCallback(t *testing.T) {
	sendErr := errors.New("cannot serialize protocol message")
	conn := newConnection(&failingSendTransport{err: sendErr})
	owner := &channelOwner{
		guid:       "failing-send-guid",
		connection: conn,
		objects:    map[string]*channelOwner{},
	}
	owner.channel = newChannel(owner, owner)

	_, err := owner.channel.Send("failingSend")
	require.ErrorIs(t, err, sendErr)
	require.Zero(t, conn.callbacks.Len(), "a failed transport send must not retain its callback")
}

func TestSendWithTimeoutPlacesTimeoutInMetadata(t *testing.T) {
	_, tr, owner := newTestConnection(t)
	timeout := Float(1234)
	_, err := owner.channel.SendWithTimeout("click", timeout, map[string]any{"selector": "button"})
	require.NoError(t, err)
	msg := tr.last()
	require.NotNil(t, msg)
	meta, ok := msg["metadata"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, float64(1234), meta["timeout"])
	params, ok := msg["params"].(map[string]any)
	require.True(t, ok)
	_, hasParamsTimeout := params["timeout"]
	require.False(t, hasParamsTimeout, "timeout must not appear in params for migrated commands")
}

func TestSendWithTimeoutPreservesExplicitZero(t *testing.T) {
	_, tr, owner := newTestConnection(t)
	timeout := Float(0)
	_, err := owner.channel.SendWithTimeout("click", timeout, map[string]any{"selector": "button", "timeout": float64(999)})
	require.NoError(t, err)
	msg := tr.last()
	meta := msg["metadata"].(map[string]any)
	require.Equal(t, float64(0), meta["timeout"])
	params := msg["params"].(map[string]any)
	_, has := params["timeout"]
	require.False(t, has)
}

func TestSendWithTimeoutNilStillStripsParamsTimeout(t *testing.T) {
	_, tr, owner := newTestConnection(t)
	_, err := owner.channel.SendWithTimeout("click", nil, map[string]any{
		"selector": "button",
		"timeout":  float64(999),
	})
	require.NoError(t, err)
	msg := tr.last()
	meta := msg["metadata"].(map[string]any)
	_, hasMetadataTimeout := meta["timeout"]
	require.False(t, hasMetadataTimeout)
	params := msg["params"].(map[string]any)
	_, hasParamsTimeout := params["timeout"]
	require.False(t, hasParamsTimeout)
}

func TestSendWithoutTimeoutOmitsMetadataTimeout(t *testing.T) {
	_, tr, owner := newTestConnection(t)
	_, err := owner.channel.Send("futureProtocolMethod", map[string]any{"timeout": float64(321)})
	require.NoError(t, err)
	msg := tr.last()
	meta := msg["metadata"].(map[string]any)
	_, has := meta["timeout"]
	require.False(t, has)
	params := msg["params"].(map[string]any)
	require.Equal(t, float64(321), params["timeout"], "ordinary Send must preserve a legitimate protocol timeout param")
}

func TestMetadataUsesInternalNotIsInternal(t *testing.T) {
	_, tr, owner := newTestConnection(t)
	_, err := owner.channel.Send("title")
	require.NoError(t, err)
	msg := tr.last()
	meta := msg["metadata"].(map[string]any)
	_, hasLegacy := meta["isInternal"]
	require.False(t, hasLegacy)
	_, hasInternal := meta["internal"]
	require.True(t, hasInternal)
}

func TestMetadataZonesAreIsolatedBetweenConcurrentCalls(t *testing.T) {
	conn, tr, owner := newTestConnection(t)
	ready := make(chan struct{}, 2)
	release := make(chan struct{})
	errs := make(chan error, 2)

	call := func(method string, internal bool) {
		_, err := conn.WrapAPICall(func() (any, error) {
			ready <- struct{}{}
			<-release
			return owner.channel.Send(method)
		}, internal)
		errs <- err
	}
	go call("publicCall", false)
	go call("internalCall", true)
	<-ready
	<-ready
	close(release)
	require.NoError(t, <-errs)
	require.NoError(t, <-errs)

	tr.mu.Lock()
	messages := append([]map[string]any(nil), tr.messages...)
	tr.mu.Unlock()
	require.Len(t, messages, 2)
	for _, msg := range messages {
		metadata := msg["metadata"].(map[string]any)
		switch msg["method"] {
		case "publicCall":
			require.Equal(t, false, metadata["internal"])
		case "internalCall":
			require.Equal(t, true, metadata["internal"])
		default:
			t.Fatalf("unexpected method %v", msg["method"])
		}
	}
}

func TestMetadataZoneIsClearedWhenSendStopsBeforeTransport(t *testing.T) {
	conn, tr, owner := newTestConnection(t)
	conn.err.Set(errors.New("pending listener error"))
	_, err := conn.WrapAPICall(func() (any, error) {
		return owner.channel.Send("failedInternalCall")
	}, true)
	require.ErrorContains(t, err, "pending listener error")
	require.Empty(t, tr.messages)

	_, err = owner.channel.Send("nextPublicCall")
	require.NoError(t, err)
	metadata := tr.last()["metadata"].(map[string]any)
	require.Equal(t, false, metadata["internal"])
}

func TestTransformOptionsDoesNotInjectDefaultTimeout(t *testing.T) {
	type opts struct {
		Timeout *float64 `json:"timeout"`
		Force   *bool    `json:"force"`
	}
	got := transformOptions([]opts{})
	_, has := got["timeout"]
	require.False(t, has, "empty options slice must not invent params.timeout")
	got = transformOptions(opts{Force: Bool(true)})
	_, has = got["timeout"]
	require.False(t, has)
	require.Equal(t, Bool(true), got["force"])
}

func TestAPIResponseTimingResponseEndIndependentOfTimingObject(t *testing.T) {
	// responseEndTiming present without a timing object → only ResponseEnd set.
	resp := newAPIResponse(nil, map[string]any{
		"status":            float64(200),
		"statusText":        "OK",
		"url":               "https://example.com/",
		"headers":           []any{},
		"fetchUid":          "u1",
		"responseEndTiming": float64(42),
	})
	timing := resp.Timing()
	require.Equal(t, float64(-1), timing.StartTime)
	require.Equal(t, float64(-1), timing.RequestStart)
	require.Equal(t, float64(42), timing.ResponseEnd)
}

func TestAPIResponseTimingAllMinusOneWhenAbsent(t *testing.T) {
	// HAR-style initializer with no timing fields.
	resp := newAPIResponse(nil, map[string]any{
		"status":     float64(200),
		"statusText": "OK",
		"url":        "https://example.com/",
		"headers":    []any{},
		"fetchUid":   "u2",
	})
	timing := resp.Timing()
	require.Equal(t, float64(-1), timing.StartTime)
	require.Equal(t, float64(-1), timing.DomainLookupStart)
	require.Equal(t, float64(-1), timing.DomainLookupEnd)
	require.Equal(t, float64(-1), timing.ConnectStart)
	require.Equal(t, float64(-1), timing.SecureConnectionStart)
	require.Equal(t, float64(-1), timing.ConnectEnd)
	require.Equal(t, float64(-1), timing.RequestStart)
	require.Equal(t, float64(-1), timing.ResponseStart)
	require.Equal(t, float64(-1), timing.ResponseEnd)
}

func TestAPIResponseTimingOverlaysTimingObject(t *testing.T) {
	resp := newAPIResponse(nil, map[string]any{
		"status":     float64(200),
		"statusText": "OK",
		"url":        "https://example.com/",
		"headers":    []any{},
		"fetchUid":   "u3",
		"timing": map[string]any{
			"startTime":     float64(1000),
			"requestStart":  float64(10),
			"responseStart": float64(20),
		},
		"responseEndTiming": float64(30),
	})
	timing := resp.Timing()
	require.Equal(t, float64(1000), timing.StartTime)
	require.Equal(t, float64(10), timing.RequestStart)
	require.Equal(t, float64(20), timing.ResponseStart)
	require.Equal(t, float64(30), timing.ResponseEnd)
	require.Equal(t, float64(-1), timing.DomainLookupStart)
}

func TestScreencastAcksAndContinuesWhenOnFramePanics(t *testing.T) {
	_, tr, owner := newTestConnection(t)
	// Build a minimal pageImpl-like owner that emits screencastFrame.
	pageOwner := &channelOwner{
		guid:       "page-guid",
		connection: owner.connection,
		objects:    map[string]*channelOwner{},
		parent:     owner,
	}
	pageOwner.channel = newChannel(pageOwner, pageOwner)
	owner.connection.objects.Store(pageOwner.guid, pageOwner)

	page := &pageImpl{}
	page.createChannelOwner(page, owner, "Page", "page-guid", map[string]any{})
	// Rebind to our pageOwner channel (createChannelOwner may replace).
	// Simpler approach: invoke the handler logic via screencastImpl after Start.
	sc := &screencastImpl{page: page}
	// Manually install the same listener Start would register.
	// Monkey-patch by calling On with our own wrapper that mirrors production.
	// Use production Start path with panicking OnFrame.
	// page.channel must work for Send - use pageOwner
	page.channel = pageOwner.channel

	var callbackCalls atomic.Int32
	require.NoError(t, sc.Start(ScreencastStartOptions{
		OnFrame: func(OnFrame) {
			if callbackCalls.Add(1) == 1 {
				panic("handler boom")
			}
		},
	}))

	page.channel.Emit("screencastFrame", map[string]any{
		"frameId":        float64(7),
		"data":           "",
		"timestamp":      float64(1),
		"viewportWidth":  float64(100),
		"viewportHeight": float64(50),
	})
	require.Eventually(t, func() bool {
		return owner.connection.err.Get() != nil
	}, time.Second, time.Millisecond)
	require.Equal(t, int32(1), callbackCalls.Load())

	// Listener errors are reported once through the next public API call, while
	// the dispatcher remains alive and can deliver later frames.
	_, err := owner.channel.Send("afterCallbackPanic")
	require.ErrorContains(t, err, "handler boom")
	_, err = owner.channel.Send("afterReportedPanic")
	require.NoError(t, err)
	page.channel.Emit("screencastFrame", map[string]any{
		"frameId":        float64(8),
		"data":           "",
		"timestamp":      float64(2),
		"viewportWidth":  float64(100),
		"viewportHeight": float64(50),
	})
	require.Eventually(t, func() bool {
		return callbackCalls.Load() == 2
	}, time.Second, time.Millisecond)

	// An invalid frame id must not be acknowledged.
	sc.mu.Lock()
	sc.onFrame = nil
	sc.mu.Unlock()
	page.channel.Emit("screencastFrame", map[string]any{"frameId": "invalid"})

	var msgs []map[string]any
	require.Eventually(t, func() bool {
		tr.mu.Lock()
		defer tr.mu.Unlock()
		ackCount := 0
		for _, msg := range tr.messages {
			if msg["method"] == "screencastFrameAck" {
				ackCount++
			}
		}
		if ackCount != 2 {
			return false
		}
		msgs = append([]map[string]any{}, tr.messages...)
		return true
	}, time.Second, time.Millisecond)
	var acks []map[string]any
	for _, msg := range msgs {
		if msg["method"] == "screencastFrameAck" {
			acks = append(acks, msg)
		}
	}
	require.Len(t, acks, 2, "each valid frame must produce exactly one ACK")
	params := acks[0]["params"].(map[string]any)
	require.Equal(t, float64(7), params["frameId"])
	params = acks[1]["params"].(map[string]any)
	require.Equal(t, float64(8), params["frameId"])
	metadata := acks[0]["metadata"].(map[string]any)
	require.Equal(t, true, metadata["internal"])
	require.Equal(t, float64(0), metadata["timeout"])
}

func TestScreencastAcksWhenOnFrameCallsGoexit(t *testing.T) {
	_, tr, owner := newTestConnection(t)
	page := &pageImpl{}
	page.createChannelOwner(page, owner, "Page", "goexit-page-guid", map[string]any{})
	sc := &screencastImpl{page: page}

	callbackStarted := make(chan struct{})
	require.NoError(t, sc.Start(ScreencastStartOptions{
		OnFrame: func(OnFrame) {
			close(callbackStarted)
			runtime.Goexit()
		},
	}))

	emitReturned := make(chan struct{})
	go func() {
		page.channel.Emit("screencastFrame", map[string]any{
			"frameId": float64(9),
			"data":    "",
		})
		close(emitReturned)
	}()
	select {
	case <-emitReturned:
	case <-time.After(time.Second):
		t.Fatal("screencast callback blocked or terminated the event dispatcher")
	}
	select {
	case <-callbackStarted:
	case <-time.After(time.Second):
		t.Fatal("screencast callback did not run")
	}

	require.Eventually(t, func() bool {
		tr.mu.Lock()
		defer tr.mu.Unlock()
		for _, msg := range tr.messages {
			if msg["method"] == "screencastFrameAck" {
				return true
			}
		}
		return false
	}, time.Second, time.Millisecond)
	require.NoError(t, owner.connection.err.Get())
	_, err := owner.channel.Send("afterGoexit")
	require.NoError(t, err)
}

func TestScreencastRestartKeepsOneListenerAndAcksWithoutActiveCallback(t *testing.T) {
	_, tr, owner := newTestConnection(t)
	page := &pageImpl{}
	page.createChannelOwner(page, owner, "Page", "restart-page-guid", map[string]any{})
	sc := &screencastImpl{page: page}

	var firstCalls atomic.Int32
	require.NoError(t, sc.Start(ScreencastStartOptions{OnFrame: func(OnFrame) { firstCalls.Add(1) }}))
	require.Equal(t, 1, page.channel.ListenerCount("screencastFrame"))
	require.NoError(t, sc.Stop())

	// A frame already in flight after Stop still needs its ACK, but the inactive
	// callback must not run.
	page.channel.Emit("screencastFrame", map[string]any{
		"frameId": float64(1),
		"data":    "%%% not base64 %%%",
	})
	require.Zero(t, firstCalls.Load())

	var secondCalls atomic.Int32
	require.NoError(t, sc.Start(ScreencastStartOptions{OnFrame: func(OnFrame) { secondCalls.Add(1) }}))
	require.Equal(t, 1, page.channel.ListenerCount("screencastFrame"))
	page.channel.Emit("screencastFrame", map[string]any{
		"frameId": float64(2),
		"data":    "",
	})
	require.Eventually(t, func() bool {
		return secondCalls.Load() == 1
	}, time.Second, time.Millisecond)
	require.Eventually(t, func() bool {
		tr.mu.Lock()
		defer tr.mu.Unlock()
		ackCount := 0
		for _, msg := range tr.messages {
			if msg["method"] == "screencastFrameAck" {
				ackCount++
			}
		}
		return ackCount == 2
	}, time.Second, time.Millisecond)
}
