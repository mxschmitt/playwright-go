package playwright

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// newTestJsonPipe builds a jsonPipe without a backing connection so the queue
// behavior can be tested in isolation.
func newTestJsonPipe() *jsonPipe {
	j := &jsonPipe{}
	j.cond = sync.NewCond(&j.mu)
	return j
}

// TestJsonPipeEnqueueNeverBlocks reproduces the deadlock the unbounded queue
// fixes: the dispatch goroutine enqueues a burst while a Send() awaits its
// reply on the same goroutine. enqueue must never block (regardless of how many
// messages queue up before a consumer drains them), otherwise the reply can
// never be delivered. The Poll consumer starts only after the burst, mirroring
// a consumer that briefly falls behind a producer spike.
func TestJsonPipeEnqueueNeverBlocks(t *testing.T) {
	j := newTestJsonPipe()

	const burst = 1000
	done := make(chan struct{})
	go func() {
		for i := 0; i < burst; i++ {
			j.enqueue(&message{ID: i})
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("enqueue blocked: producer did not finish without a consumer draining")
	}

	// The burst must still be fully drained in order once a consumer catches up.
	for i := 0; i < burst; i++ {
		msg, err := j.Poll()
		require.NoError(t, err)
		require.Equal(t, i, msg.ID)
	}
}

// TestJsonPipeMultipleConsumers verifies the cond-based design supports more
// than one concurrent Poll() consumer: every enqueued message is delivered
// exactly once and none are stranded.
func TestJsonPipeMultipleConsumers(t *testing.T) {
	j := newTestJsonPipe()

	const total = 200
	got := make(chan int, total)
	var wg sync.WaitGroup
	for c := 0; c < 4; c++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				msg, err := j.Poll()
				if err != nil {
					return
				}
				got <- msg.ID
			}
		}()
	}

	for i := 0; i < total; i++ {
		j.enqueue(&message{ID: i})
	}
	j.markClosed()
	wg.Wait()
	close(got)

	seen := make(map[int]bool, total)
	for id := range got {
		require.False(t, seen[id], "message %d delivered more than once", id)
		seen[id] = true
	}
	require.Len(t, seen, total, "every enqueued message must be delivered exactly once")
}

// TestJsonPipePreservesOrder verifies messages are delivered in the order they
// were enqueued.
func TestJsonPipePreservesOrder(t *testing.T) {
	j := newTestJsonPipe()
	for i := 0; i < 100; i++ {
		j.enqueue(&message{ID: i})
	}
	for i := 0; i < 100; i++ {
		msg, err := j.Poll()
		require.NoError(t, err)
		require.Equal(t, i, msg.ID)
	}
}

// TestJsonPipePollBlocksUntilMessage verifies Poll waits for a message that is
// enqueued concurrently.
func TestJsonPipePollBlocksUntilMessage(t *testing.T) {
	j := newTestJsonPipe()
	got := make(chan *message, 1)
	go func() {
		msg, err := j.Poll()
		require.NoError(t, err)
		got <- msg
	}()

	// Give Poll a moment to start waiting, then deliver.
	time.Sleep(50 * time.Millisecond)
	j.enqueue(&message{ID: 42})

	select {
	case msg := <-got:
		require.Equal(t, 42, msg.ID)
	case <-time.After(5 * time.Second):
		t.Fatal("Poll did not return after enqueue")
	}
}

// TestJsonPipeCloseUnblocksPoll verifies that closing the pipe wakes a waiting
// Poll with an error.
func TestJsonPipeCloseUnblocksPoll(t *testing.T) {
	j := newTestJsonPipe()
	errCh := make(chan error, 1)
	go func() {
		_, err := j.Poll()
		errCh <- err
	}()

	time.Sleep(50 * time.Millisecond)
	j.markClosed()

	select {
	case err := <-errCh:
		require.Error(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Poll did not return after close")
	}
}

// TestJsonPipeDrainsQueueBeforeClose verifies buffered messages are still
// delivered after the pipe is marked closed, before the close error.
func TestJsonPipeDrainsQueueBeforeClose(t *testing.T) {
	j := newTestJsonPipe()
	for i := 0; i < 3; i++ {
		j.enqueue(&message{ID: i})
	}
	j.markClosed()

	for i := 0; i < 3; i++ {
		msg, err := j.Poll()
		require.NoError(t, err, fmt.Sprintf("message %d should drain before close", i))
		require.Equal(t, i, msg.ID)
	}
	_, err := j.Poll()
	require.Error(t, err)
}
