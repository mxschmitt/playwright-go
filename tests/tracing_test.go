package playwright_test

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestBrowserContextOutputTrace(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, context.Tracing().Start(playwright.TracingStartOptions{
		Screenshots: playwright.Bool(true),
		Snapshots:   playwright.Bool(true),
	}))

	_, err := page.Goto(server.PREFIX + "/grid.html")
	require.NoError(t, err)
	dir := t.TempDir()
	err = context.Tracing().Stop(filepath.Join(dir, "trace.zip"))
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(dir, "trace.zip"))
}

func TestTracingStartStop(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, context.Tracing().Start())
	require.NoError(t, context.Tracing().Stop())
}

func TestBrowserContextShouldNoErrorWhenStoppingWithoutStart(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, context.Tracing().Stop())
}

func TestBrowserContextOutputTraceChunk(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, context.Tracing().Start(playwright.TracingStartOptions{
		Screenshots: playwright.Bool(true),
		Snapshots:   playwright.Bool(true),
	}))

	_, err := page.Goto(server.PREFIX + "/grid.html")
	require.NoError(t, err)
	dir := t.TempDir()

	button := page.Locator(".box").First()

	err = context.Tracing().StartChunk(playwright.TracingStartChunkOptions{
		Title: playwright.String("foo"),
	})
	require.NoError(t, err)
	err = button.Click()
	require.NoError(t, err)
	err = context.Tracing().StopChunk(filepath.Join(dir, "trace1.zip"))
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(dir, "trace1.zip"))

	err = context.Tracing().StartChunk(playwright.TracingStartChunkOptions{
		Title: playwright.String("foo"),
	})
	require.NoError(t, err)
	err = button.Click()
	require.NoError(t, err)
	err = context.Tracing().StopChunk(filepath.Join(dir, "trace2.zip"))
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(dir, "trace2.zip"))
}

func TestBrowserContextTracingOutputMultipleChunks(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, context.Tracing().Start(playwright.TracingStartOptions{
		Screenshots: playwright.Bool(true),
		Snapshots:   playwright.Bool(true),
	}))

	_, err := page.Goto(server.PREFIX + "/frames/frame.html")
	require.NoError(t, err)
	require.NoError(t, context.Tracing().StartChunk())
	require.NoError(t, page.SetContent("<button>Click</button>"))
	require.NoError(t, page.Locator("button").Click())
	dir := t.TempDir()
	require.NoError(t, context.Tracing().StopChunk(filepath.Join(dir, "trace.zip")))
	require.FileExists(t, filepath.Join(dir, "trace.zip"))
}

func TestBrowserContextTracingRemoteConnect(t *testing.T) {
	BeforeEach(t)

	remoteServer, err := newRemoteServer()
	require.NoError(t, err)
	defer remoteServer.Close()

	browser1, err := browserType.Connect(remoteServer.url)
	require.NoError(t, err)
	require.NotNil(t, browser1)
	defer browser1.Close() //nolint:errcheck

	context1, err := browser1.NewContext()
	require.NoError(t, err)
	require.NoError(t, context1.Tracing().Start(playwright.TracingStartOptions{
		Screenshots: playwright.Bool(true),
		Snapshots:   playwright.Bool(true),
	}))
	page1, err := context1.NewPage()
	require.NoError(t, err)
	_, err = page1.Goto(server.PREFIX + "/frames/frame.html")
	require.NoError(t, err)
	require.NoError(t, context1.Tracing().StartChunk())
	require.NoError(t, page1.SetContent("<button>Click</button>"))
	require.NoError(t, page1.Locator("button").Click())
	dir := t.TempDir()
	require.NoError(t, context1.Tracing().StopChunk(filepath.Join(dir, "trace.zip")))
	require.FileExists(t, filepath.Join(dir, "trace.zip"))
}

func TestShouldShowTracingGroupInActionList(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, context.Tracing().Start())
	page, err := context.NewPage()
	require.NoError(t, err)

	require.NoError(t, context.Tracing().Group("outer group"))
	_, err = page.Goto(`data:text/html,<!DOCTYPE html><body><div>Hello world</div></body>`)
	require.NoError(t, err)
	require.NoError(t, context.Tracing().Group("inner group 1"))
	require.NoError(t, page.Locator("body").Click())
	require.NoError(t, context.Tracing().GroupEnd())
	require.NoError(t, context.Tracing().Group("inner group 2"))
	visiable, err := page.GetByText("Hello").IsVisible()
	require.NoError(t, err)
	require.True(t, visiable)
	require.NoError(t, context.Tracing().GroupEnd())
	require.NoError(t, context.Tracing().GroupEnd())

	tracePath := filepath.Join(t.TempDir(), "trace.zip")
	require.NoError(t, context.Tracing().Stop(tracePath))
	require.FileExists(t, tracePath)

	_, events := parseTrace(t, tracePath)
	actions := getTraceActions(events)
	require.Equal(t,
		[]string{
			"BrowserContext.NewPage",
			"outer group",
			"Page.Goto",
			"inner group 1",
			"Locator.Click",
			"inner group 2",
			"Locator.IsVisible",
		}, actions)
}

// mapInternalAPIToPublic maps internal Playwright class.method names to public API names
func mapInternalAPIToPublic(class, method string) string {
	// Map internal classes to public API classes
	classMethodKey := class + "." + method

	// Common Frame methods that should map to Page
	frameToPageMethods := map[string]bool{
		"goto": true, "reload": true, "goBack": true, "goForward": true,
		"setContent": true, "waitForNavigation": true, "waitForURL": true,
		"waitForLoadState": true, "screenshot": true, "pdf": true,
		"close": true, "pause": true,
	}

	// Frame selector/locator methods that should map to Locator
	frameLocatorMethods := map[string]bool{
		"click": true, "dblclick": true, "fill": true, "press": true,
		"type": true, "hover": true, "check": true, "uncheck": true,
		"selectOption": true, "setInputFiles": true, "focus": true,
		"blur": true, "tap": true, "dispatchEvent": true, "evaluate": true,
		"isVisible": true, "isHidden": true, "isEnabled": true, "isDisabled": true,
		"isChecked": true, "isEditable": true, "textContent": true,
		"innerText": true, "innerHTML": true, "getAttribute": true,
	}

	if class == "Frame" {
		if frameToPageMethods[method] {
			class = "Page"
		} else if frameLocatorMethods[method] {
			class = "Locator"
		}
	}

	// Special case mappings
	specialMappings := map[string]string{
		"BrowserContext.newPage": "BrowserContext.NewPage",
		"Page.waitForTimeout":    "Page.WaitForTimeout",
	}

	// Convert method to title case (first letter uppercase)
	titleCaseMethod := strings.ToUpper(method[:1]) + method[1:]
	apiName := class + "." + titleCaseMethod

	// Check for special mappings
	if mapped, ok := specialMappings[classMethodKey]; ok {
		return mapped
	}
	if mapped, ok := specialMappings[apiName]; ok {
		return mapped
	}

	return apiName
}

func parseTrace(t *testing.T, tracePath string) (files map[string][]byte, events []interface{}) {
	t.Helper()
	// read and unzip trace
	r, err := zip.OpenReader(tracePath)
	require.NoError(t, err)
	defer r.Close() //nolint:errcheck

	files = make(map[string][]byte)
	events = make([]interface{}, 0)
	actionMap := make(map[string]interface{})
	for _, f := range r.File {
		rc, err := f.Open()
		require.NoError(t, err)
		defer rc.Close() //nolint:errcheck

		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, rc)
		require.NoError(t, err)

		files[f.Name] = buf.Bytes()
		if f.Name == "trace.trace" || f.Name == "trace.network" {
			// read lines
			for _, line := range bytes.Split(buf.Bytes(), []byte("\n")) {
				if len(line) == 0 {
					continue
				}

				var event map[string]interface{}
				err := json.Unmarshal(line, &event)
				require.NoError(t, err)
				eventType, _ := event["type"].(string)

				switch eventType {
				case "before":
					event["type"] = "action"
					// Compute apiName from class and method for regular actions
					// For tracing groups, use the title field
					class, _ := event["class"].(string)
					method, _ := event["method"].(string)
					title, hasTitle := event["title"].(string)

					if method == "tracingGroup" && hasTitle {
						event["apiName"] = title
					} else if class != "" && method != "" {
						event["apiName"] = mapInternalAPIToPublic(class, method)
					}

					actionMap[event["callId"].(string)] = event
					events = append(events, event)
				case "input":
				case "after":
				default:
					events = append(events, event)
				}
			}
		}
	}

	return
}

func getTraceActions(events []interface{}) []string {
	actions := make([]string, 0)
	actionEvents := slices.DeleteFunc(events, func(e interface{}) bool {
		event := e.(map[string]interface{})
		return event["type"].(string) != "action"
	})
	slices.SortFunc(actionEvents, func(a, b interface{}) int {
		eventA := a.(map[string]interface{})
		eventB := b.(map[string]interface{})
		t1 := eventA["startTime"].(float64)
		t2 := eventB["startTime"].(float64)
		if t1 < t2 {
			return -1
		}
		if t1 > t2 {
			return 1
		}
		return 0
	})
	for _, e := range actionEvents {
		event := e.(map[string]interface{})
		if apiName, ok := event["apiName"].(string); ok {
			actions = append(actions, apiName)
		}
	}
	return actions
}

// Ported from upstream tests/library/har.spec.ts "tracing.startHar" >
// "should record a HAR with options".
func TestTracingStartHarWithOptions(t *testing.T) {
	BeforeEach(t)

	harPath := filepath.Join(t.TempDir(), "tracing.har")
	require.NoError(t, context.Tracing().StartHar(harPath, playwright.TracingStartHarOptions{
		Mode:      playwright.HarModeMinimal,
		URLFilter: "**/one-style.css",
	}))
	_, err := page.Goto(server.PREFIX + "/one-style.html")
	require.NoError(t, err)
	require.NoError(t, context.Tracing().StopHar())
	require.FileExists(t, harPath)

	data, err := os.ReadFile(harPath)
	require.NoError(t, err)
	var har struct {
		Log struct {
			Entries []struct {
				Request struct {
					URL      string  `json:"url"`
					BodySize float64 `json:"bodySize"`
				} `json:"request"`
			} `json:"entries"`
		} `json:"log"`
	}
	require.NoError(t, json.Unmarshal(data, &har))
	urls := make([]string, 0, len(har.Log.Entries))
	for _, e := range har.Log.Entries {
		urls = append(urls, e.Request.URL)
	}
	require.Equal(t, []string{server.PREFIX + "/one-style.css"}, urls)
	// Minimal mode drops body sizes.
	require.Equal(t, float64(-1), har.Log.Entries[0].Request.BodySize)
}

// Ported from upstream "should record a zipped HAR for APIRequestContext",
// adapted to a BrowserContext (the Go API exposes Tracing on both).
func TestTracingStartHarZipped(t *testing.T) {
	BeforeEach(t)

	harPath := filepath.Join(t.TempDir(), "tracing.har.zip")
	require.NoError(t, context.Tracing().StartHar(harPath, playwright.TracingStartHarOptions{
		Content: playwright.HarContentPolicyAttach,
	}))
	_, err := page.Goto(server.PREFIX + "/one-style.html")
	require.NoError(t, err)
	require.NoError(t, context.Tracing().StopHar())
	require.FileExists(t, harPath)

	// The zip contains the har.har entry alongside attached resources.
	zr, err := zip.OpenReader(harPath)
	require.NoError(t, err)
	defer zr.Close() //nolint:errcheck
	var harEntry *zip.File
	for _, f := range zr.File {
		if f.Name == "har.har" {
			harEntry = f
			break
		}
	}
	require.NotNil(t, harEntry, "zip should contain har.har")
	rc, err := harEntry.Open()
	require.NoError(t, err)
	defer rc.Close() //nolint:errcheck
	data, err := io.ReadAll(rc)
	require.NoError(t, err)
	require.Contains(t, string(data), server.PREFIX+"/one-style.html")
}

// StopHar without a prior StartHar should error, mirroring upstream's
// "HAR recording has not been started" guard.
func TestTracingStopHarWithoutStart(t *testing.T) {
	BeforeEach(t)

	require.ErrorContains(t, context.Tracing().StopHar(), "HAR recording has not been started")
}
