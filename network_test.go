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

func TestNewRawHeadersSupportsTypedMaps(t *testing.T) {
	headers := newRawHeaders([]map[string]string{
		{
			"name":  "Accept",
			"value": "text/html",
		},
		{
			"name":  "Set-Cookie",
			"value": "a=b",
		},
		{
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
