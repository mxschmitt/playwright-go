package playwright_test

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mxschmitt/playwright-go"
)

const webMCPAddTool = `
	modelContext.registerTool({
		name: 'add',
		description: 'Adds two numbers',
		inputSchema: { type: 'object', properties: { a: { type: 'number' }, b: { type: 'number' } }, required: ['a', 'b'] },
		annotations: { readOnlyHint: true },
		async execute(input) { return { content: [{ type: 'text', text: String(input.a + input.b) }] }; },
	});
`

func serveWebMCPTools(path, tools string, extraHTML ...string) {
	server.SetRoute(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = fmt.Fprintf(w, `<title>WebMCP</title><script>
			const modelContext = document.modelContext || navigator.modelContext;
			%s
		</script>%s`, tools, strings.Join(extraHTML, ""))
	})
}

// newWebMCPPage launches a browser with WebMCP enabled and returns a page in it.
func newWebMCPPage(t *testing.T) playwright.Page {
	t.Helper()
	if isWebKit {
		t.Skip("WebKit does not support WebMCP")
	}
	BeforeEach(t)
	options := playwright.BrowserTypeLaunchOptions{Headless: playwright.Bool(headless)}
	if isFirefox {
		options.FirefoxUserPrefs = map[string]any{
			"dom.modelcontext.enabled":         true,
			"dom.modelcontext.testing.enabled": true,
		}
	} else {
		options.Args = []string{"--enable-features=WebMCP"}
	}
	webMCPBrowser, err := browserType.Launch(options)
	require.NoError(t, err)
	t.Cleanup(func() { _ = webMCPBrowser.Close() })
	page, err := webMCPBrowser.NewPage()
	require.NoError(t, err)
	return page
}

func TestWebMCPShouldListToolsRegisteredByThePage(t *testing.T) {
	page := newWebMCPPage(t)
	serveWebMCPTools("/webmcp.html", webMCPAddTool+`
		modelContext.registerTool({
			name: 'noop',
			description: 'Does nothing',
			async execute() {},
		});
	`)
	_, err := page.Goto(server.PREFIX + "/webmcp.html")
	require.NoError(t, err)
	tools, err := page.WebMCP().Tools()
	require.NoError(t, err)
	// Browsers list tools in their own order.
	sort.Slice(tools, func(i, j int) bool { return tools[i].Name < tools[j].Name })
	require.Equal(t, []playwright.WebMCPTool{
		{
			Name:        "add",
			Description: "Adds two numbers",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{"a": map[string]any{"type": "number"}, "b": map[string]any{"type": "number"}},
				"required":   []any{"a", "b"},
			},
			Annotations: &playwright.WebMCPToolAnnotations{ReadOnly: playwright.Bool(true)},
		},
		{Name: "noop", Description: "Does nothing"},
	}, tools)
	frameTools, err := page.MainFrame().WebMCP().Tools()
	require.NoError(t, err)
	sort.Slice(frameTools, func(i, j int) bool { return frameTools[i].Name < frameTools[j].Name })
	require.Equal(t, tools, frameTools)
}

func TestWebMCPShouldReportNoToolsWhenThePageRegistersNone(t *testing.T) {
	page := newWebMCPPage(t)
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	tools, err := page.WebMCP().Tools()
	require.NoError(t, err)
	require.Empty(t, tools)
}

func TestWebMCPShouldCallATool(t *testing.T) {
	page := newWebMCPPage(t)
	serveWebMCPTools("/webmcp.html", webMCPAddTool)
	_, err := page.Goto(server.PREFIX + "/webmcp.html")
	require.NoError(t, err)
	result, err := page.WebMCP().CallTool("add", map[string]any{"a": 2, "b": 40})
	require.NoError(t, err)
	require.Equal(t, map[string]any{"content": []any{map[string]any{"type": "text", "text": "42"}}}, result)
}

func TestWebMCPShouldReturnIsErrorResultsAsIs(t *testing.T) {
	page := newWebMCPPage(t)
	serveWebMCPTools("/webmcp.html", `
		modelContext.registerTool({
			name: 'broken',
			description: 'Always fails',
			async execute() { return { content: [{ type: 'text', text: 'nope' }], isError: true }; },
		});
	`)
	_, err := page.Goto(server.PREFIX + "/webmcp.html")
	require.NoError(t, err)
	result, err := page.WebMCP().CallTool("broken", nil)
	require.NoError(t, err)
	require.Equal(t, map[string]any{"content": []any{map[string]any{"type": "text", "text": "nope"}}, "isError": true}, result)
}

func TestWebMCPShouldThrowWhenTheToolIsNotRegistered(t *testing.T) {
	page := newWebMCPPage(t)
	serveWebMCPTools("/webmcp.html", webMCPAddTool)
	_, err := page.Goto(server.PREFIX + "/webmcp.html")
	require.NoError(t, err)
	_, err = page.WebMCP().CallTool("missing", nil)
	require.ErrorContains(t, err, `No WebMCP tool named "missing"`)
}

func TestWebMCPShouldTimeOutWhenTheToolDoesNotSettle(t *testing.T) {
	page := newWebMCPPage(t)
	serveWebMCPTools("/webmcp.html", `
		modelContext.registerTool({
			name: 'stuck',
			description: 'Never resolves',
			execute() { return new Promise(() => {}); },
		});
	`)
	_, err := page.Goto(server.PREFIX + "/webmcp.html")
	require.NoError(t, err)
	_, err = page.WebMCP().CallTool("stuck", map[string]any{}, playwright.WebMCPCallToolOptions{Timeout: playwright.Float(500)})
	require.ErrorIs(t, err, playwright.ErrTimeout)
	require.ErrorContains(t, err, "Timeout 500ms exceeded")
}

func TestWebMCPShouldScopeToolsToTheirFrame(t *testing.T) {
	if isFirefox {
		t.Skip("Firefox does not support registering WebMCP tools in iframes")
	}
	page := newWebMCPPage(t)
	serveWebMCPTools("/webmcp.html", webMCPAddTool, `<iframe src="/webmcp-frame.html"></iframe>`)
	serveWebMCPTools("/webmcp-frame.html", `
		modelContext.registerTool({
			name: 'subscribe',
			description: 'Subscribes to the newsletter',
			async execute() { return { content: [{ type: 'text', text: 'subscribed' }] }; },
		});
	`)
	_, err := page.Goto(server.PREFIX + "/webmcp.html")
	require.NoError(t, err)

	childFrame := page.Frames()[1]
	mainTools, err := page.WebMCP().Tools()
	require.NoError(t, err)
	require.Len(t, mainTools, 1)
	require.Equal(t, "add", mainTools[0].Name)
	frameTools, err := childFrame.WebMCP().Tools()
	require.NoError(t, err)
	require.Len(t, frameTools, 1)
	require.Equal(t, "subscribe", frameTools[0].Name)
	result, err := childFrame.WebMCP().CallTool("subscribe", nil)
	require.NoError(t, err)
	require.Equal(t, map[string]any{"content": []any{map[string]any{"type": "text", "text": "subscribed"}}}, result)

	_, err = page.WebMCP().CallTool("subscribe", nil)
	require.ErrorContains(t, err, `No WebMCP tool named "subscribe"`)
	require.ErrorContains(t, err, "Available tools: add")
}

func TestWebMCPShouldThrowWhenTheToolThrows(t *testing.T) {
	page := newWebMCPPage(t)
	serveWebMCPTools("/webmcp.html", `
		modelContext.registerTool({
			name: 'throwing',
			description: 'Throws',
			async execute() { throw new Error('boom'); },
		});
	`)
	_, err := page.Goto(server.PREFIX + "/webmcp.html")
	require.NoError(t, err)
	_, err = page.WebMCP().CallTool("throwing", nil)
	// Chromium does not report the error that the tool threw.
	if isChromium {
		require.ErrorContains(t, err, "the invocation failed")
	} else {
		require.ErrorContains(t, err, "boom")
	}
}
