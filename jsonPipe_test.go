package playwright

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// newTestJsonPipe builds a jsonPipe without a backing connection so the queue
// behavior can be tested in isolation.
func newTestJsonPipe() *jsonPipe {
	return &jsonPipe{
		signal: make(chan struct{}, 1),
	}
}

// TestJsonPipeEnqueueNeverBlocks guards against the deadlock where the outer
// connection's dispatch goroutine blocks delivering a message because the queue
// is full. enqueue must always return promptly, even with no consumer polling.
func TestJsonPipeEnqueueNeverBlocks(t *testing.T) {
	j := newTestJsonPipe()

	done := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			j.enqueue(&message{ID: i})
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("enqueue blocked: producer did not finish without a consumer")
	}
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
