package playwright

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_assignFloatIfPresent(t *testing.T) {
	t.Run("overrides target when key holds a float64", func(t *testing.T) {
		target := -1.0
		assignFloatIfPresent(map[string]any{"startTime": 123.0}, "startTime", &target)
		require.Equal(t, 123.0, target)
	})

	t.Run("leaves target untouched when key is missing", func(t *testing.T) {
		target := -1.0
		assignFloatIfPresent(map[string]any{}, "startTime", &target)
		require.Equal(t, -1.0, target)
	})

	t.Run("leaves target untouched when value is nil", func(t *testing.T) {
		target := -1.0
		assignFloatIfPresent(map[string]any{"startTime": nil}, "startTime", &target)
		require.Equal(t, -1.0, target)
	})

	t.Run("leaves target untouched when value is not a float64", func(t *testing.T) {
		target := -1.0
		assignFloatIfPresent(map[string]any{"startTime": "not-a-float"}, "startTime", &target)
		require.Equal(t, -1.0, target)
	})
}

func TestToOptionalStorageStatePreservesCredentials(t *testing.T) {
	state := StorageState{
		Cookies: []Cookie{{Name: "a", Value: "b", Domain: "ex.com", Path: "/"}},
		Origins: []Origin{{Origin: "https://ex.com", LocalStorage: []NameValue{{Name: "k", Value: "v"}}}},
		Credentials: []VirtualCredential{{
			Id: "id1", RpId: "ex.com", UserHandle: "u", PrivateKey: "pk", PublicKey: "pub",
		}},
	}
	opt := state.ToOptionalStorageState()
	require.Len(t, opt.Credentials, 1)
	require.Equal(t, "id1", opt.Credentials[0].Id)
	require.Equal(t, "a", opt.Cookies[0].Name)
}
