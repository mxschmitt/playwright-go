package playwright

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestSelectorsAddContextIsIdempotent verifies that addContext registers a
// context at most once: re-adding the same context (same guid) is a no-op,
// mirroring upstream's _contextsForSelectors Set semantics. This guards against
// double registration if a context-creation path ever calls addContext more
// than once for the same context.
func TestSelectorsAddContextIsIdempotent(t *testing.T) {
	// Clear the testId attribute for this test so addContext has no engines and
	// no testId to push, and therefore never touches context.channel — only the
	// LoadOrStore membership logic runs. Restore the global afterwards.
	prevTestID := testIdAttributeName
	testIdAttributeName = ""
	t.Cleanup(func() { testIdAttributeName = prevTestID })

	s := newSelectorsImpl()
	ctx := &browserContextImpl{channelOwner: channelOwner{guid: "context@idempotent"}}

	s.addContext(ctx)
	first, ok := s.contexts.Load("context@idempotent")
	require.True(t, ok)
	require.Same(t, ctx, first)

	// Re-adding the same context must be a no-op and must not replace the stored
	// value (LoadOrStore keeps the first).
	other := &browserContextImpl{channelOwner: channelOwner{guid: "context@idempotent"}}
	s.addContext(other)
	stored, ok := s.contexts.Load("context@idempotent")
	require.True(t, ok)
	require.Same(t, ctx, stored, "second addContext must not overwrite the stored context")

	count := 0
	s.contexts.Range(func(_, _ any) bool { count++; return true })
	require.Equal(t, 1, count)
}
