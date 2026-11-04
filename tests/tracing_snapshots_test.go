package playwright_test

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mxschmitt/playwright-go"
)

func parseTraceEvents(t *testing.T, path string) []map[string]any {
	t.Helper()
	reader, err := zip.OpenReader(path)
	require.NoError(t, err)
	defer reader.Close() //nolint:errcheck
	var events []map[string]any
	for _, file := range reader.File {
		if !strings.HasSuffix(file.Name, ".trace") {
			continue
		}
		rc, err := file.Open()
		require.NoError(t, err)
		scanner := bufio.NewScanner(rc)
		scanner.Buffer(make([]byte, 1024*1024), 64*1024*1024)
		for scanner.Scan() {
			if len(scanner.Bytes()) == 0 {
				continue
			}
			var event map[string]any
			require.NoError(t, json.Unmarshal(scanner.Bytes(), &event))
			events = append(events, event)
		}
		require.NoError(t, scanner.Err())
		require.NoError(t, rc.Close())
	}
	return events
}

func clickTrace(t *testing.T, options playwright.TracingStartOptions) []map[string]any {
	t.Helper()
	require.NoError(t, context.Tracing().Start(options))
	_, err := page.Goto(server.PREFIX + "/input/button.html")
	require.NoError(t, err)
	require.NoError(t, page.Locator("button").Click())
	path := filepath.Join(t.TempDir(), "trace.zip")
	require.NoError(t, context.Tracing().Stop(path))
	return parseTraceEvents(t, path)
}

func TestTracingShouldNotCollectAriaAndScreenSnapshotsByDefault(t *testing.T) {
	BeforeEach(t)

	events := clickTrace(t, playwright.TracingStartOptions{Snapshots: playwright.Bool(true)})
	for _, event := range events {
		require.NotEqual(t, "screenshot", event["type"])
		require.NotEqual(t, "aria-snapshot", event["type"])
	}
}

func TestTracingShouldCollectAriaAndScreenSnapshots(t *testing.T) {
	BeforeEach(t)

	events := clickTrace(t, playwright.TracingStartOptions{
		AriaSnapshots:   playwright.Bool(true),
		ScreenSnapshots: playwright.Bool(true),
	})
	var clickCallID any
	for _, event := range events {
		if event["type"] == "before" && event["method"] == "click" {
			clickCallID = event["callId"]
			break
		}
	}
	require.NotNil(t, clickCallID)
	for _, eventType := range []string{"screenshot", "aria-snapshot"} {
		var phases []any
		for _, event := range events {
			if event["type"] == eventType && event["callId"] == clickCallID {
				phases = append(phases, event["phase"])
			}
		}
		require.Equal(t, []any{"before", "action", "after"}, phases, eventType)
	}
}

func TestTracingShouldRecordStacksWithStringCallIDs(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, context.Tracing().Start(playwright.TracingStartOptions{Snapshots: playwright.Bool(true)}))
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	path := filepath.Join(t.TempDir(), "trace.zip")
	require.NoError(t, context.Tracing().Stop(path))

	reader, err := zip.OpenReader(path)
	require.NoError(t, err)
	defer reader.Close() //nolint:errcheck
	var stacks struct {
		Stacks [][]any `json:"stacks"`
	}
	for _, file := range reader.File {
		if file.Name != "trace.stacks" {
			continue
		}
		rc, err := file.Open()
		require.NoError(t, err)
		require.NoError(t, json.NewDecoder(rc).Decode(&stacks))
		require.NoError(t, rc.Close())
	}
	// Each recorded call carries the client stack, keyed by the call id.
	require.NotEmpty(t, stacks.Stacks)
	for _, stack := range stacks.Stacks {
		require.Regexp(t, `^[a-z]{4}@\d+$`, stack[0])
	}
}
