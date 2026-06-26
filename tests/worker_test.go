package playwright_test

import (
	"testing"
	"time"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestWorkerShouldWork(t *testing.T) {
	BeforeEach(t)

	worker, err := page.ExpectWorker(func() error {
		_, err := page.Goto(server.PREFIX + "/worker/worker.html")
		return err
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(page.Workers()))
	require.Equal(t, worker, page.Workers()[0])
	worker = page.Workers()[0]
	require.Contains(t, worker.URL(), "worker.js")
	// flaky in the macos-latest of gh action
	require.Eventually(
		t,
		func() bool {
			v, err := worker.Evaluate(`() => self["workerFunction"] ? true : false`)
			require.NoError(t, err)
			return v == true
		},
		500*time.Millisecond, 10*time.Millisecond,
	)
	res, err := worker.Evaluate(`() => self["workerFunction"]()`)
	require.NoError(t, err)
	require.Equal(t, "worker function result", res)
	_, err = page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Equal(t, 0, len(page.Workers()))
}

func TestWorkerShouldEmitCreatedAndDestroyedEvents(t *testing.T) {
	BeforeEach(t)

	var workerObj playwright.JSHandle
	worker, err := page.ExpectWorker(func() error {
		workerObjInterface, err := page.EvaluateHandle("() => new Worker(URL.createObjectURL(new Blob(['1'], {type: 'application/javascript'})))")
		workerObj = workerObjInterface
		return err
	})
	require.NoError(t, err)
	workerThisObj, err := worker.EvaluateHandle(`() => this`)
	require.NoError(t, err)
	eventFired := make(chan bool, 1)
	worker.OnClose(func(w playwright.Worker) {
		eventFired <- true
	})
	_, err = page.Evaluate("workerObj => workerObj.terminate()", workerObj)
	require.NoError(t, err)
	require.True(t, <-eventFired)
	_, err = workerThisObj.GetProperty("self")
	require.Error(t, err)
}

func TestWorkerShouldReportConsoleLogs(t *testing.T) {
	BeforeEach(t)

	message, err := page.ExpectEvent("console", func() error {
		_, err := page.Evaluate("() => new Worker(URL.createObjectURL(new Blob(['console.log(1)'], {type: 'application/javascript'})))")
		return err
	})
	require.NoError(t, err)
	require.Equal(t, message.(playwright.ConsoleMessage).Text(), "1")
}

func TestWorkerShouldHaveJSHandlesForConsoleLogs(t *testing.T) {
	BeforeEach(t)

	message, err := page.ExpectEvent("console", func() error {
		_, err := page.Evaluate("() => new Worker(URL.createObjectURL(new Blob(['console.log(1,2,3,this)'], {type: 'application/javascript'})))")
		return err
	})
	require.NoError(t, err)
	log := message.(playwright.ConsoleMessage)
	if !isFirefox {
		require.Equal(t, "1 2 3 DedicatedWorkerGlobalScope", log.Text())
	} else {
		require.Equal(t, "1 2 3 JSHandle@object", log.Text())
	}
	require.Equal(t, 4, len(log.Args()))
	origin, err := log.Args()[3].GetProperty("origin")
	require.NoError(t, err)
	val, err := origin.JSONValue()
	require.NoError(t, err)
	require.Equal(t, "null", val)
}

func TestWorkerShouldEvaluate(t *testing.T) {
	BeforeEach(t)

	worker, err := page.ExpectWorker(func() error {
		_, err := page.Evaluate("() => new Worker(URL.createObjectURL(new Blob(['console.log(1)'], {type: 'application/javascript'})))")
		return err
	})
	require.NoError(t, err)
	result, err := worker.Evaluate("1+1")
	require.NoError(t, err)
	require.Equal(t, 2, result)
}

func TestWorkershouldReportErrors(t *testing.T) {
	BeforeEach(t)

	errChan := make(chan error, 1)
	page.OnPageError(func(err error) {
		errChan <- err
	})

	_, err := page.Evaluate("() => new Worker(URL.createObjectURL(new Blob([`\n" +
		"  setTimeout(() => {\n" +
		"    // Do a console.log just to check that we do not confuse it with an error.\n" +
		"    console.log('hey');\n" +
		"    throw new Error('this is my error');\n" +
		"  })\n" +
		"`], {type: 'application/javascript'})))")
	require.NoError(t, err)
	pageError := <-errChan
	require.ErrorContains(t, pageError, "this is my error")
}

func TestWorkerShouldClearUponCrossProcessNavigation(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	worker, err := page.ExpectWorker(func() error {
		_, err := page.Evaluate("() => new Worker(URL.createObjectURL(new Blob(['console.log(1)'], {type: 'application/javascript'})))")
		return err
	})
	require.NoError(t, err)
	require.Equal(t, len(page.Workers()), 1)
	destroyed := false
	_ = worker
	worker.OnClose(func(w playwright.Worker) {
		destroyed = true
	})
	_, err = page.Goto(server.CROSS_PROCESS_PREFIX + "/empty.html")
	require.NoError(t, err)
	require.True(t, destroyed)
	require.Equal(t, 0, len(page.Workers()))
}

// TestConsoleMessageWorker verifies that ConsoleMessage.Worker() returns the worker
// Based on upstream test: playwright/tests/page/workers.spec.ts
func TestConsoleMessageWorker(t *testing.T) {
	BeforeEach(t)

	// Create a worker and capture console message from it
	workerChan := make(chan playwright.Worker, 1)
	page.Once("worker", func(worker playwright.Worker) {
		workerChan <- worker
	})

	consoleChan := make(chan playwright.ConsoleMessage, 1)
	page.Once("console", func(message playwright.ConsoleMessage) {
		consoleChan <- message
	})

	// Create a worker that logs a message
	_, err := page.Evaluate(`() => {
		const workerCode = 'console.log("hello from worker")';
		const blob = new Blob([workerCode], { type: 'application/javascript' });
		const worker = new Worker(URL.createObjectURL(blob));
	}`)
	require.NoError(t, err)

	// Wait for worker creation
	worker := <-workerChan
	require.NotNil(t, worker)

	// Wait for console message
	message := <-consoleChan
	require.NotNil(t, message)

	// Verify the message is from the worker
	require.Equal(t, "hello from worker", message.Text())

	msgWorker, err := message.Worker()
	require.NoError(t, err)
	require.Equal(t, worker, msgWorker, "console message should reference the worker")

	// Worker console messages also have a page reference (they're emitted to both)
	msgPage := message.Page()
	require.Equal(t, page, msgPage, "worker console messages are also associated with the page")
}

// TestConsoleMessageWorkerNil verifies that page console messages have nil worker
func TestConsoleMessageWorkerNil(t *testing.T) {
	BeforeEach(t)

	consoleChan := make(chan playwright.ConsoleMessage, 1)
	page.Once("console", func(message playwright.ConsoleMessage) {
		consoleChan <- message
	})

	_, err := page.Evaluate(`() => console.log('hello from page')`)
	require.NoError(t, err)

	message := <-consoleChan
	require.NotNil(t, message)
	require.Equal(t, "hello from page", message.Text())

	// Page console messages should not have a worker
	msgWorker, err := message.Worker()
	require.NoError(t, err)
	require.Nil(t, msgWorker, "page console messages should not have a worker")

	// But should have a page
	msgPage := message.Page()
	require.Equal(t, page, msgPage)
}
