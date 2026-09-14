package playwright

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// inertTransport never carries a message: these tests drive Dispatch directly.
type inertTransport struct{}

func (inertTransport) Send(map[string]any) error { return nil }
func (inertTransport) Poll() (*message, error)   { return nil, ErrTargetClosed }
func (inertTransport) Close() error              { return nil }

func newBoundOwner(conn *connection, guid, objectType string) *channelOwner {
	owner := &channelOwner{
		guid:       guid,
		objectType: objectType,
		connection: conn,
		objects:    map[string]*channelOwner{},
	}
	owner.channel = newChannel(owner, owner)
	conn.objects.Store(guid, owner)
	return owner
}

// A CDPSession "event" relays raw CDP JSON. Chrome's Page.downloadWillBegin and
// Page.downloadProgress carry a "guid" field that names a download, not a
// protocol object; the event must reach the listener with that field intact
// and the connection must stay open.
func TestDispatchEventLeavesUnknownGUIDUntouched(t *testing.T) {
	conn := newConnection(inertTransport{})
	session := newBoundOwner(conn, "cdp-session-1", "CDPSession")

	var got map[string]any
	session.channel.On("event", func(params map[string]any) { got = params })

	const downloadGUID = "de856c15-889c-4e8a-a5cd-82f77ee4ca64"
	conn.Dispatch(&message{
		GUID:   "cdp-session-1",
		Method: "event",
		Params: map[string]any{
			"method": "Page.downloadProgress",
			"params": map[string]any{"guid": downloadGUID, "state": "inProgress", "receivedBytes": 4096.0},
		},
	})

	require.NoError(t, conn.closedError.Get(), "connection must survive an unbound guid in event params")
	require.NotNil(t, got, "event must still be delivered")
	inner, ok := got["params"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, downloadGUID, inner["guid"])
	require.Equal(t, "inProgress", inner["state"])
}

// The same walk must keep resolving guids that DO name bound objects.
func TestDispatchEventStillResolvesBoundGUIDs(t *testing.T) {
	conn := newConnection(inertTransport{})
	session := newBoundOwner(conn, "cdp-session-1", "CDPSession")
	page := newBoundOwner(conn, "page-1", "Page")

	var got map[string]any
	session.channel.On("event", func(params map[string]any) { got = params })

	conn.Dispatch(&message{
		GUID:   "cdp-session-1",
		Method: "event",
		Params: map[string]any{"page": map[string]any{"guid": "page-1"}, "nested": []any{map[string]any{"guid": "page-1"}}},
	})

	require.NoError(t, conn.closedError.Get())
	require.Same(t, page.channel, got["page"])
	require.Same(t, page.channel, got["nested"].([]any)[0])
}
