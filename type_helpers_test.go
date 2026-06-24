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
