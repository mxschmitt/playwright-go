package playwright

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRawHeadersSupportsUntypedMaps(t *testing.T) {
	headers := newRawHeaders([]any{
		map[string]any{
			"name":  "Accept",
			"value": "text/html",
		},
		map[string]any{
			"name":  "Set-Cookie",
			"value": "a=b",
		},
		map[string]any{
			"name":  "Set-Cookie",
			"value": "c=d",
		},
	})

	require.Equal(t, "text/html", headers.Get("accept"))
	require.Equal(t, []NameValue{
		{Name: "Accept", Value: "text/html"},
		{Name: "Set-Cookie", Value: "a=b"},
		{Name: "Set-Cookie", Value: "c=d"},
	}, headers.HeadersArray())
	require.Equal(t, "a=b\nc=d", headers.Get("set-cookie"))
}

func TestNewRawHeadersSupportsTypedNameValues(t *testing.T) {
	headers := newRawHeaders([]NameValue{
		{Name: "Accept", Value: "text/html"},
		{Name: "Set-Cookie", Value: "a=b"},
		{Name: "Set-Cookie", Value: "c=d"},
	})

	require.Equal(t, "text/html", headers.Get("accept"))
	require.Equal(t, []NameValue{
		{Name: "Accept", Value: "text/html"},
		{Name: "Set-Cookie", Value: "a=b"},
		{Name: "Set-Cookie", Value: "c=d"},
	}, headers.HeadersArray())
	require.Equal(t, "a=b\nc=d", headers.Get("set-cookie"))
}

// Mirrors the request.Headers()/ActualHeaders() fallback-override path, which
// feeds serializeMapToNameAndValue's output straight into newRawHeaders. This
// is the shape that previously panicked (see issue #453).
func TestNewRawHeadersFromSerializedMap(t *testing.T) {
	headers := newRawHeaders(serializeMapToNameAndValue(map[string]string{
		"Accept": "text/html",
	}))

	require.Equal(t, "text/html", headers.Get("accept"))
	require.Equal(t, []NameValue{
		{Name: "Accept", Value: "text/html"},
	}, headers.HeadersArray())
}
