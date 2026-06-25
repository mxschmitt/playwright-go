package playwright

import (
	"runtime"
	"strconv"
	"sync"
)

var goidStackPool = sync.Pool{
	New: func() any {
		b := make([]byte, 64)
		return &b
	},
}

// currentGoroutineID returns the ID of the calling goroutine. The Go runtime
// does not expose this, so it is parsed from the runtime stack header
// ("goroutine <id> [...]"). It is used only to detect whether a blocking server
// call is being made from the dispatch goroutine, in which case the reply must
// be awaited by re-entrantly pumping the receive loop rather than blocking on a
// channel that only the dispatch goroutine could ever signal.
func currentGoroutineID() uint64 {
	bp := goidStackPool.Get().(*[]byte)
	defer goidStackPool.Put(bp)
	b := (*bp)[:cap(*bp)]
	b = b[:runtime.Stack(b, false)]
	// Format: "goroutine 123 [running]:\n..."
	const prefix = "goroutine "
	b = b[len(prefix):]
	i := 0
	for i < len(b) && b[i] >= '0' && b[i] <= '9' {
		i++
	}
	id, _ := strconv.ParseUint(string(b[:i]), 10, 64)
	return id
}
