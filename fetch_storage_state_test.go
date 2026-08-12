package playwright

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSanitizeStorageStateForAPIRequestStripsOnlyCredentials(t *testing.T) {
	state := &StorageState{
		Cookies: []Cookie{{
			Name:   "session",
			Value:  "secret",
			Domain: "example.com",
			Path:   "/",
		}},
		Origins: []Origin{{
			Origin: "https://example.com",
			LocalStorage: []NameValue{{
				Name:  "token",
				Value: "value",
			}},
		}},
		Credentials: []VirtualCredential{{Id: "credential-id"}},
	}

	sanitized := sanitizeStorageStateForAPIRequest(state)
	require.NotSame(t, state, sanitized)
	require.Equal(t, state.Cookies, sanitized.Cookies)
	require.Equal(t, state.Origins, sanitized.Origins)
	require.Empty(t, sanitized.Credentials)
	// Sanitizing the transport payload must not mutate the caller's state.
	require.Len(t, state.Credentials, 1)
}

func TestSanitizeStorageStateForAPIRequestAcceptsNil(t *testing.T) {
	require.Nil(t, sanitizeStorageStateForAPIRequest(nil))
}
