package playwright_test

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestBrowserTypeBrowserName(t *testing.T) {
	BeforeEach(t)

	require.Equal(t, pw.Chromium.Name(), "chromium")
	require.Equal(t, pw.Firefox.Name(), "firefox")
	require.Equal(t, pw.WebKit.Name(), "webkit")
}

func TestBrowserTypeExecutablePath(t *testing.T) {
	BeforeEach(t)

	require.Greater(t, len(pw.Chromium.ExecutablePath()), 0)
}

func TestBrowserTypeLaunchPersistentContext(t *testing.T) {
	BeforeEach(t)

	user_data_dir1 := t.TempDir()
	browser_context, err := browserType.LaunchPersistentContext(user_data_dir1)
	require.NoError(t, err)
	page, err := browser_context.NewPage()
	require.NoError(t, err)
	_, err = page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	_, err = page.Evaluate("() => localStorage.hey = 'hello'")
	require.NoError(t, err)
	require.NoError(t, browser_context.Close())

	browser_context2, err := browserType.LaunchPersistentContext(user_data_dir1)
	require.NoError(t, err)
	page2, err := browser_context2.NewPage()
	require.NoError(t, err)
	_, err = page2.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	result, err := page2.Evaluate("() => localStorage.hey")
	require.NoError(t, err)
	require.Equal(t, "hello", result)
	require.NoError(t, browser_context2.Close())

	user_data_dir2 := t.TempDir()
	browser_context3, err := browserType.LaunchPersistentContext(user_data_dir2)
	require.NoError(t, err)
	page3, err := browser_context3.NewPage()
	require.NoError(t, err)
	_, err = page3.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	result, err = page3.Evaluate("() => localStorage.hey")
	require.NoError(t, err)
	require.NotEqual(t, "hello", result)
	require.NoError(t, browser_context3.Close())
}

// TestLaunchPersistentContextAppliesSelectorEngines verifies that a custom
// selector engine registered before LaunchPersistentContext is applied to the
// persistent context (previously addContext was skipped for persistent contexts).
func TestLaunchPersistentContextAppliesSelectorEngines(t *testing.T) {
	BeforeEach(t)

	tagSelector := `(() => ({
		query(root, selector) { return root.querySelector(selector); },
		queryAll(root, selector) { return Array.from(root.querySelectorAll(selector)); }
	}))()`
	engineName := fmt.Sprintf("ptag_%s_%d", browserName, time.Now().UnixNano())
	require.NoError(t, pw.Selectors.Register(engineName, playwright.Script{Content: &tagSelector}))

	ctx, err := browserType.LaunchPersistentContext(t.TempDir())
	require.NoError(t, err)
	defer ctx.Close() //nolint:errcheck
	p, err := ctx.NewPage()
	require.NoError(t, err)
	require.NoError(t, p.SetContent(`<div><span></span></div>`))
	name, err := p.Locator(engineName+"=DIV").Evaluate("e => e.nodeName", nil)
	require.NoError(t, err)
	require.Equal(t, "DIV", name)
}

func TestBrowserTypeConnect(t *testing.T) {
	BeforeEach(t)

	remoteServer, err := newRemoteServer()
	require.NoError(t, err)
	defer remoteServer.Close()

	browser1, err := browserType.Connect(remoteServer.url)
	require.NoError(t, err)
	require.NotNil(t, browser1)
	defer browser1.Close() //nolint:errcheck

	browser_context, err := browser1.NewContext()
	require.NoError(t, err)
	page, err := browser_context.NewPage()
	require.NoError(t, err)
	result, err := page.Evaluate("11 * 11")
	require.NoError(t, err)
	require.Equal(t, result, 121)
	require.NoError(t, browser1.Close())
}

func TestBrowserTypeConnectShouldBeAbleToReconnectToBrowser(t *testing.T) {
	BeforeEach(t)

	remoteServer, err := newRemoteServer()
	require.NoError(t, err)
	defer remoteServer.Close()

	browser1, err := browserType.Connect(remoteServer.url)
	require.NoError(t, err)
	require.NotNil(t, browser1)
	defer browser1.Close() //nolint:errcheck

	require.Len(t, browser1.Contexts(), 0)
	browser_context, err := browser1.NewContext()
	require.NoError(t, err)
	require.Len(t, browser1.Contexts(), 1)
	require.Len(t, browser_context.Pages(), 0)
	page, err := browser_context.NewPage()
	require.Len(t, browser_context.Pages(), 1)
	require.NoError(t, err)
	result, err := page.Evaluate("11 * 11")
	require.NoError(t, err)
	require.Equal(t, result, 121)
	require.NoError(t, browser1.Close())

	browser1, err = browserType.Connect(remoteServer.url)
	require.NoError(t, err)
	require.NotNil(t, browser1)
	defer browser1.Close() //nolint:errcheck

	require.Len(t, browser1.Contexts(), 0)
	browser_context, err = browser1.NewContext()
	require.NoError(t, err)
	require.Len(t, browser1.Contexts(), 1)
	require.Len(t, browser_context.Pages(), 0)
	page, err = browser_context.NewPage()
	require.Len(t, browser_context.Pages(), 1)
	require.NoError(t, err)
	result, err = page.Evaluate("11 * 11")
	require.NoError(t, err)
	require.Equal(t, result, 121)
	require.NoError(t, browser1.Close())
}

func TestBrowserTypeConnectShouldEmitDisconnectedEvent(t *testing.T) {
	BeforeEach(t)

	remoteServer, err := newRemoteServer()
	require.NoError(t, err)
	disconnected1 := newSyncSlice[bool]()
	disconnected2 := newSyncSlice[bool]()
	browser1, err := browserType.Connect(remoteServer.url)
	require.NoError(t, err)
	require.NotNil(t, browser1)
	browser2, err := browserType.Connect(remoteServer.url)
	require.NoError(t, err)
	require.NotNil(t, browser2)
	browser1.OnDisconnected(func(playwright.Browser) {
		disconnected1.Append(true)
	})
	browser2.OnDisconnected(func(playwright.Browser) {
		disconnected2.Append(true)
	})
	page, err := browser2.NewPage()
	require.NoError(t, err)
	require.NoError(t, browser1.Close())
	require.False(t, browser1.IsConnected())
	require.Len(t, disconnected1.Get(), 1)
	require.Len(t, disconnected2.Get(), 0)
	remoteServer.Close()

	_, err = page.Title()
	require.Error(t, err)
	require.False(t, browser2.IsConnected())
	require.Len(t, disconnected2.Get(), 1)
}

func TestBrowserTypeConnectSlowMo(t *testing.T) {
	BeforeEach(t)

	remoteServer, err := newRemoteServer()
	require.NoError(t, err)
	defer remoteServer.Close()

	browser1, err := browserType.Connect(remoteServer.url, playwright.BrowserTypeConnectOptions{
		SlowMo: playwright.Float(100),
	})
	require.NoError(t, err)
	require.NotNil(t, browser1)
	defer browser1.Close() //nolint:errcheck

	browser_context, err := browser1.NewContext()
	require.NoError(t, err)
	t1 := time.Now()
	page, err := browser_context.NewPage()
	require.NoError(t, err)
	result, err := page.Evaluate("11 * 11")
	require.NoError(t, err)
	require.Equal(t, result, 121)
	require.GreaterOrEqual(t, time.Since(t1), time.Duration(time.Millisecond*200))
	require.NoError(t, browser1.Close())
}

func TestBrowserTypeConnectArtifactPath(t *testing.T) {
	BeforeEach(t)

	remoteServer, err := newRemoteServer()
	require.NoError(t, err)
	defer remoteServer.Close()

	browser1, err := browserType.Connect(remoteServer.url)
	require.NoError(t, err)
	require.NotNil(t, browser1)
	defer browser1.Close() //nolint:errcheck

	recordVideoDir := t.TempDir()
	browserContext, err := browser1.NewContext(playwright.BrowserNewContextOptions{
		RecordVideo: &playwright.RecordVideo{
			Dir: playwright.String(recordVideoDir),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, browserContext)
	defer browserContext.Close() //nolint:errcheck
	page, err := browserContext.NewPage()
	require.NoError(t, err)
	defer page.Close() //nolint:errcheck
	_, err = page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	_, err = page.Video().Path()
	require.Error(t, err)
	require.Equal(t, err, errors.New("Path is not available when connecting remotely. Use SaveAs() to save a local copy."))
}

func TestBrowserTypeConnectOverCDP(t *testing.T) {
	if !isChromium {
		t.Skip("CDP is only supported on Chromium")
	}
	BeforeEach(t)

	port, err := getFreePort()
	require.NoError(t, err)
	browserServer, err := browserType.Launch(playwright.BrowserTypeLaunchOptions{
		Args: []string{fmt.Sprintf("--remote-debugging-port=%d", port)},
	})
	require.NoError(t, err)
	defer browserServer.Close() //nolint:errcheck
	// Register a custom selector engine before connecting; it must reach the
	// CDP default context (it was previously never registered there).
	tagSelector := `(() => ({
		query(root, selector) { return root.querySelector(selector); },
		queryAll(root, selector) { return Array.from(root.querySelectorAll(selector)); }
	}))()`
	engineName := fmt.Sprintf("cdptag_%d", time.Now().UnixNano())
	require.NoError(t, pw.Selectors.Register(engineName, playwright.Script{Content: &tagSelector}))

	browser, err := browserType.ConnectOverCDP(fmt.Sprintf("http://localhost:%d", port))
	require.NoError(t, err)
	require.NotNil(t, browser)
	defer browser.Close() //nolint:errcheck
	require.Len(t, browser.Contexts(), 1)

	ctx := browser.Contexts()[0]
	p, err := ctx.NewPage()
	require.NoError(t, err)
	require.NoError(t, p.SetContent(`<div><span></span></div>`))
	name, err := p.Locator(engineName+"=DIV").Evaluate("e => e.nodeName", nil)
	require.NoError(t, err)
	require.Equal(t, "DIV", name)
}

func TestBrowserTypeConnectOverCDPTwice(t *testing.T) {
	if !isChromium {
		t.Skip("CDP is only supported on Chromium")
	}
	BeforeEach(t)

	port, err := getFreePort()
	require.NoError(t, err)
	browserServer, err := browserType.Launch(playwright.BrowserTypeLaunchOptions{
		Args: []string{fmt.Sprintf("--remote-debugging-port=%d", port)},
	})
	require.NoError(t, err)
	defer browserServer.Close() //nolint:errcheck
	browser1, err := browserType.ConnectOverCDP(fmt.Sprintf("http://localhost:%d", port))
	require.NoError(t, err)
	require.NotNil(t, browser1)
	browser2, err := browserType.ConnectOverCDP(fmt.Sprintf("http://localhost:%d", port))
	require.NoError(t, err)
	require.NotNil(t, browser2)
	defer browser1.Close() //nolint:errcheck
	defer browser2.Close() //nolint:errcheck
	require.Len(t, browser1.Contexts(), 1)
	page1, err := browser1.Contexts()[0].NewPage()
	require.NoError(t, err)
	_, err = page1.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Len(t, browser2.Contexts(), 1)
	page2, err := browser2.Contexts()[0].NewPage()
	require.NoError(t, err)
	_, err = page2.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)

	require.Len(t, browser1.Contexts()[0].Pages(), 2)
	require.Len(t, browser2.Contexts()[0].Pages(), 2)
}

func TestSetInputFilesShouldPreserveLastModifiedTimestamp(t *testing.T) {
	BeforeEach(t)

	remoteServer, err := newRemoteServer()
	require.NoError(t, err)
	defer remoteServer.Close()

	browser1, err := browserType.Connect(remoteServer.url)
	require.NoError(t, err)
	require.NotNil(t, browser1)
	defer browser1.Close() //nolint:errcheck

	browser_context, err := browser1.NewContext()
	require.NoError(t, err)
	page, err := browser_context.NewPage()
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`<input type=file multiple=true/>`))
	input := page.Locator("input")
	filenames := []string{
		"file-to-upload.txt",
		"file-to-upload-2.txt",
	}
	files := make([]string, len(filenames))
	for i, filename := range filenames {
		files[i] = Asset(filename)
	}
	require.NoError(t, input.SetInputFiles(files))
	result, err := input.Evaluate(`input => [...input.files].map(f => f.name)`, nil)
	require.NoError(t, err)
	names, ok := result.([]interface{})
	require.True(t, ok)
	for i, name := range names {
		require.Equal(t, filenames[i], name.(string))
	}

	results, err := input.Evaluate(`input => [...input.files].map(f => f.lastModified)`, nil)
	require.NoError(t, err)
	timestamps, ok := results.([]interface{})
	require.True(t, ok)
	for i, timestamp := range timestamps {
		expected, err := getFileLastModifiedTimeMs(files[i])
		require.NoError(t, err)
		// On Linux browser sometimes reduces the timestamp by 1ms: 1696272058110.0715  -> 1696272058109 or even
		// rounds it to seconds in WebKit: 1696272058110 -> 1696272058000.
		require.Less(t, math.Abs(float64(expected-int64(timestamp.(int)))), 1000.0)
	}
}

func TestShouldUploadAFolderRemote(t *testing.T) {
	BeforeEach(t)

	remoteServer, err := newRemoteServer()
	require.NoError(t, err)
	defer remoteServer.Close()

	browser1, err := browserType.Connect(remoteServer.url)
	require.NoError(t, err)
	require.NotNil(t, browser1)
	defer browser1.Close() //nolint:errcheck

	browser_context, err := browser1.NewContext()
	require.NoError(t, err)
	page, err := browser_context.NewPage()
	require.NoError(t, err)
	// WebKit folder upload over a remote connection (file streaming + browser
	// ingestion) can exceed the default 30s action timeout on loaded CI runners.
	// flakiness-go does not retry (unlike upstream's CI retries: 3), so raise the
	// default action timeout to give the slow WebKit step room.
	page.SetDefaultTimeout(90 * 1000)

	_, err = page.Goto(fmt.Sprintf("%s%s", server.PREFIX, "/input/folderupload.html"))
	require.NoError(t, err)

	//nolint:staticcheck
	input, err := page.QuerySelector("input")
	require.NoError(t, err)

	dir := filepath.Join(t.TempDir(), "file-upload-test")
	require.NoError(t, os.MkdirAll(dir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("file1 content"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file2"), []byte("file2 content"), 0o600))
	require.Nil(t, os.Mkdir(filepath.Join(dir, "sub-dir"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sub-dir", "really.txt"), []byte("sub-dir file content"), 0o600))
	//nolint:staticcheck
	require.NoError(t, input.SetInputFiles(dir))

	ret, err := input.Evaluate(`e => [...e.files].map(f => f.webkitRelativePath)`)
	require.NoError(t, err)

	expectResult := []interface{}{"file-upload-test/file1.txt", "file-upload-test/file2"}
	// https://issues.chromium.org/issues/345393164
	if !isChromium || !headless || !chromiumVersionLessThan(browser.Version(), "127.0.6533.0") {
		expectResult = append(expectResult, "file-upload-test/sub-dir/really.txt")
	}
	slices.SortFunc(ret.([]interface{}), func(i, j interface{}) int {
		return strings.Compare(i.(string), j.(string))
	})
	require.Equal(t, expectResult, ret.([]interface{}))

	webkitRelativePaths, err := input.Evaluate(`e => [...e.files].map(f => f.webkitRelativePath)`)
	require.NoError(t, err)
	for i, path := range webkitRelativePaths.([]interface{}) {
		content, err := input.Evaluate(`(e, i) => {
			const reader = new FileReader();
			const promise = new Promise(fulfill => reader.onload = fulfill);
			reader.readAsText(e.files[i]);
			return promise.then(() => reader.result);
    }`, i)
		require.NoError(t, err)
		b, err := os.ReadFile(filepath.Join(dir, "..", path.(string)))
		require.NoError(t, err)
		require.Equal(t, string(b), content.(string))
	}
}

func TestBrowserBind(t *testing.T) {
	BeforeEach(t)

	result, err := browser.Bind("test-server")
	require.NoError(t, err)
	require.NotEmpty(t, result.Endpoint)

	// Connect from another Playwright instance
	pw2, err := playwright.Run()
	require.NoError(t, err)
	defer func() { _ = pw2.Stop() }()

	browser2, err := pw2.Chromium.Connect(result.Endpoint)
	require.NoError(t, err)
	require.NotNil(t, browser2)

	page2, err := browser2.NewPage()
	require.NoError(t, err)
	_, err = page2.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Equal(t, server.EMPTY_PAGE, page2.URL())

	require.NoError(t, browser2.Close())
	require.NoError(t, browser.Unbind())
}
