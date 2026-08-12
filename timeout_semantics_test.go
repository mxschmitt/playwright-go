package playwright

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// timeoutSemanticsTransport records protocol envelopes and returns the small
// canned responses needed by the timeout semantics tests below.
type timeoutSemanticsTransport struct {
	mu          sync.Mutex
	messages    []map[string]any
	replies     chan *message
	closed      chan struct{}
	delays      map[string]time.Duration
	elementGUID string
}

func newTimeoutSemanticsTransport() *timeoutSemanticsTransport {
	return &timeoutSemanticsTransport{
		replies: make(chan *message, 16),
		closed:  make(chan struct{}),
		delays:  make(map[string]time.Duration),
	}
}

func (t *timeoutSemanticsTransport) Send(outgoing map[string]any) error {
	encoded, _ := json.Marshal(outgoing)
	var captured map[string]any
	_ = json.Unmarshal(encoded, &captured)

	t.mu.Lock()
	t.messages = append(t.messages, captured)
	delay := t.delays[captured["method"].(string)]
	t.mu.Unlock()
	if delay > 0 {
		time.Sleep(delay)
	}

	result := map[string]any{}
	switch captured["method"] {
	case "isHidden":
		result["value"] = true
	case "isVisible":
		result["value"] = false
	case "inputValue":
		result["value"] = "input value"
	case "waitForSelector":
		result["element"] = map[string]any{"guid": t.elementGUID}
	}
	t.replies <- &message{ID: int(captured["id"].(float64)), Result: result}
	return nil
}

func (t *timeoutSemanticsTransport) Poll() (*message, error) {
	select {
	case reply := <-t.replies:
		return reply, nil
	case <-t.closed:
		return nil, ErrTargetClosed
	}
}

func (t *timeoutSemanticsTransport) Close() error {
	select {
	case <-t.closed:
	default:
		close(t.closed)
	}
	return nil
}

func (t *timeoutSemanticsTransport) messageFor(method string) map[string]any {
	t.mu.Lock()
	defer t.mu.Unlock()
	for i := len(t.messages) - 1; i >= 0; i-- {
		if t.messages[i]["method"] == method {
			return t.messages[i]
		}
	}
	return nil
}

func newTimeoutSemanticsFixture(t *testing.T, defaultTimeout float64) (*timeoutSemanticsTransport, *frameImpl, *elementHandleImpl) {
	t.Helper()
	transport := newTimeoutSemanticsTransport()
	connection := newConnection(transport)
	root := &channelOwner{
		guid:       "timeout-test-root",
		connection: connection,
		objects:    make(map[string]*channelOwner),
	}
	root.channel = newChannel(root, root)
	connection.objects.Store(root.guid, root)

	frame := newFrame(root, "Frame", "timeout-test-frame", map[string]any{
		"name":       "",
		"url":        "about:blank",
		"loadStates": []any{},
	})
	settings := newTimeoutSettings(nil)
	settings.SetDefaultTimeout(Float(defaultTimeout))
	settings.SetDefaultNavigationTimeout(Float(defaultTimeout))
	frame.page = &pageImpl{
		timeoutSettings: settings,
		browserContext:  &browserContextImpl{},
	}

	handle := newElementHandle(&frame.channelOwner, "ElementHandle", "timeout-test-element", map[string]any{
		"preview": "JSHandle@node",
	})
	transport.elementGUID = handle.guid

	go func() {
		for connection.pollOnce() {
		}
	}()
	t.Cleanup(func() { _ = transport.Close() })
	return transport, frame, handle
}

func requireProtocolNoTimeout(t *testing.T, envelope map[string]any) {
	t.Helper()
	require.NotNil(t, envelope)
	metadata := envelope["metadata"].(map[string]any)
	require.Equal(t, float64(0), metadata["timeout"])
	params := envelope["params"].(map[string]any)
	_, hasParamsTimeout := params["timeout"]
	require.False(t, hasParamsTimeout)
}

func TestImmediateQueriesIgnoreDeprecatedTimeout(t *testing.T) {
	transport, frame, handle := newTimeoutSemanticsFixture(t, 5000)

	hidden, err := frame.IsHidden("#hidden", FrameIsHiddenOptions{
		Strict:  Bool(true),
		Timeout: Float(1234),
	})
	require.NoError(t, err)
	require.True(t, hidden)
	hiddenMessage := transport.messageFor("isHidden")
	requireProtocolNoTimeout(t, hiddenMessage)
	require.Equal(t, "#hidden", hiddenMessage["params"].(map[string]any)["selector"])
	require.Equal(t, true, hiddenMessage["params"].(map[string]any)["strict"])

	visible, err := frame.IsVisible("#visible", FrameIsVisibleOptions{
		Strict:  Bool(true),
		Timeout: Float(2345),
	})
	require.NoError(t, err)
	require.False(t, visible)
	visibleMessage := transport.messageFor("isVisible")
	requireProtocolNoTimeout(t, visibleMessage)
	require.Equal(t, "#visible", visibleMessage["params"].(map[string]any)["selector"])
	require.Equal(t, true, visibleMessage["params"].(map[string]any)["strict"])

	value, err := handle.InputValue(ElementHandleInputValueOptions{Timeout: Float(3456)})
	require.NoError(t, err)
	require.Equal(t, "input value", value)
	requireProtocolNoTimeout(t, transport.messageFor("inputValue"))
}

func TestWaitForTimeoutUsesProtocolWaitTimeout(t *testing.T) {
	transport, frame, _ := newTimeoutSemanticsFixture(t, 5000)

	frame.WaitForTimeout(42.5)

	envelope := transport.messageFor("waitForTimeout")
	requireProtocolNoTimeout(t, envelope)
	require.Equal(t, float64(42.5), envelope["params"].(map[string]any)["waitTimeout"])
}

func TestLocatorWithElementPreservesExplicitZeroTimeout(t *testing.T) {
	transport, frame, _ := newTimeoutSemanticsFixture(t, 5000)
	locator := newLocator(frame, "button")

	_, err := locator.withElement(func(handle ElementHandle, timeout *float64) (any, error) {
		require.NotNil(t, timeout)
		require.Zero(t, *timeout)
		return nil, handle.ScrollIntoViewIfNeeded(ElementHandleScrollIntoViewIfNeededOptions{Timeout: timeout})
	}, FrameWaitForSelectorOptions{Timeout: Float(0)})
	require.NoError(t, err)

	waitMetadata := transport.messageFor("waitForSelector")["metadata"].(map[string]any)
	require.Equal(t, float64(0), waitMetadata["timeout"])
	actionMessage := transport.messageFor("scrollIntoViewIfNeeded")
	actionMetadata := actionMessage["metadata"].(map[string]any)
	require.Equal(t, float64(0), actionMetadata["timeout"])
	_, hasParamsTimeout := actionMessage["params"].(map[string]any)["timeout"]
	require.False(t, hasParamsTimeout)
}

func TestLocatorWithElementSharesResolvedDefaultTimeout(t *testing.T) {
	transport, frame, _ := newTimeoutSemanticsFixture(t, 1000)
	transport.delays["waitForSelector"] = 30 * time.Millisecond
	locator := newLocator(frame, "button")

	_, err := locator.withElement(func(handle ElementHandle, timeout *float64) (any, error) {
		require.NotNil(t, timeout)
		return nil, handle.ScrollIntoViewIfNeeded(ElementHandleScrollIntoViewIfNeededOptions{Timeout: timeout})
	})
	require.NoError(t, err)

	waitTimeout := transport.messageFor("waitForSelector")["metadata"].(map[string]any)["timeout"].(float64)
	actionTimeout := transport.messageFor("scrollIntoViewIfNeeded")["metadata"].(map[string]any)["timeout"].(float64)
	require.Equal(t, float64(1000), waitTimeout)
	require.Positive(t, actionTimeout)
	require.Less(t, actionTimeout, waitTimeout)
}

func TestExpectNavigationZeroTimeoutStillWaitsForLoadState(t *testing.T) {
	_, frame, _ := newTimeoutSemanticsFixture(t, 1000)

	started := time.Now()
	_, err := frame.ExpectNavigation(func() error {
		frame.Emit("navigated", map[string]any{
			"url":         "https://example.test/",
			"newDocument": nil,
		})
		go func() {
			time.Sleep(25 * time.Millisecond)
			frame.Emit("loadstate", "load")
		}()
		return nil
	}, FrameExpectNavigationOptions{
		Timeout:   Float(0),
		WaitUntil: WaitUntilStateLoad,
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, time.Since(started), 20*time.Millisecond)
}

func TestExpectNavigationUsesOnePositiveTimeoutBudget(t *testing.T) {
	_, frame, _ := newTimeoutSemanticsFixture(t, 1000)

	started := time.Now()
	_, err := frame.ExpectNavigation(func() error {
		time.Sleep(70 * time.Millisecond)
		frame.Emit("navigated", map[string]any{
			"url":         "https://example.test/",
			"newDocument": nil,
		})
		return nil
	}, FrameExpectNavigationOptions{
		Timeout:   Float(120),
		WaitUntil: WaitUntilStateLoad,
	})
	require.ErrorIs(t, err, ErrTimeout)
	require.Less(t, time.Since(started), 170*time.Millisecond)
}
