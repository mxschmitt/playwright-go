package playwright

import (
	"runtime"
	"strconv"
)

// currentGoroutineID returns the ID of the calling goroutine. The Go runtime
// does not expose this, so it is parsed from the runtime stack header
// ("goroutine <id> [...]"). It is used only to detect whether a blocking server
// call is being made from the dispatch goroutine, in which case the reply must
// be awaited by re-entrantly pumping the receive loop rather than blocking on a
// channel that only the dispatch goroutine could ever signal.
//
// It is called once per blocking server call. runtime.Stack and the 64-byte
// buffer it fills are not free, but both are cheap relative to the protocol
// round-trip each such call already pays.
func currentGoroutineID() uint64 {
	// The header alone ("goroutine 123 [running]:") fits comfortably in 64
	// bytes; runtime.Stack truncates to the buffer, which is all we need.
	var buf [64]byte
	b := buf[:runtime.Stack(buf[:], false)]
	// Format: "goroutine 123 [running]:\n..."
	const prefix = "goroutine "
	if len(b) < len(prefix) {
		return 0
	}
	b = b[len(prefix):]
	i := 0
	for i < len(b) && b[i] >= '0' && b[i] <= '9' {
		i++
	}
	id, err := strconv.ParseUint(string(b[:i]), 10, 64)
	if err != nil {
		// Unexpected stack header format. Return 0, which is never a valid
		// goroutine id (the runtime numbers them from 1), so the caller is not
		// mistaken for the dispatch goroutine and safely takes the normal
		// blocking path.
		return 0
	}
	return id
}
