package playwright_test

import (
	"bytes"
	"image"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	_ "image/png"

	"github.com/orisano/pixelmatch"
	"github.com/playwright-community/playwright-go"
)

// global variables, can be used in any tests
var (
	pw          *playwright.Playwright
	browser     playwright.Browser
	context     playwright.BrowserContext
	page        playwright.Page
	expect      playwright.PlaywrightAssertions
	headless    = os.Getenv("HEADFUL") == ""
	isChromium  bool
	isFirefox   bool
	isWebKit    bool
	browserName = getBrowserName()
	browserType playwright.BrowserType
)

// default context options for most tests
var DEFAULT_CONTEXT_OPTIONS = playwright.BrowserNewContextOptions{
	AcceptDownloads: playwright.Bool(true),
	HasTouch:        playwright.Bool(true),
}

// TestMain is used to setup and teardown the tests
func TestMain(m *testing.M) {
	BeforeAll()
	code := m.Run()
	AfterAll()
	os.Exit(code)
}

func skipWebKitMacOSPopup(t *testing.T) {
	t.Helper()
	if isWebKit && runtime.GOOS == "darwin" {
		// Keep this scoped to popup tests: GitHub macOS WebKit closes the browser
		// while opening these popups, which cascades into unrelated test failures.
		t.Skip("WebKit closes the browser on macOS when opening this popup")
	}
}

// BeforeAll prepares the environment, including
//   - start Playwright driver
//   - launch browser depends on BROWSER env
//   - init web-first assertions, alias as `expect`
func BeforeAll() {
	var err error
	pw, err = playwright.Run()
	if err != nil {
		log.Fatalf("could not start Playwright: %v", err)
	}
	switch browserName {
	case "chromium", "":
		browserType = pw.Chromium
	case "firefox":
		browserType = pw.Firefox
	case "webkit":
		browserType = pw.WebKit
	}
	launchOptions := playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
	}
	if browserType == pw.Chromium {
		launchOptions.Args = []string{"--disable-features=LocalNetworkAccessChecks"}
	}
	// launch browser, headless or not depending on HEADFUL env
	browser, err = browserType.Launch(launchOptions)
	if err != nil {
		log.Fatalf("could not launch: %v", err)
	}
	// init web-first assertions with 1s timeout instead of default 5s
	expect = playwright.NewPlaywrightAssertions(1000)
	isChromium = browserName == "chromium" || browserName == ""
	isFirefox = browserName == "firefox"
	isWebKit = browserName == "webkit"

	// for playwright-go tests
	server = newTestServer()
	utils = &testUtils{}
}

// AfterAll does cleanup, e.g. stop playwright driver
func AfterAll() {
	if server != nil {
		server.testServer.Close()
	}
	if err := pw.Stop(); err != nil {
		log.Fatalf("could not start Playwright: %v", err)
	}
}

// BeforeEach creates a new context and page for each test,
// so each test has isolated environment. Usage:
//
//	Func TestFoo(t *testing.T) {
//	  BeforeEach(t)
//	  // your test code
//	}
func BeforeEach(t *testing.T, contextOptions ...playwright.BrowserNewContextOptions) {
	t.Helper()
	opt := DEFAULT_CONTEXT_OPTIONS
	if len(contextOptions) == 1 {
		opt = contextOptions[0]
	}
	context, page = newBrowserContextAndPage(t, opt)

	t.Cleanup(func() {
		server.AfterEach()
	})
}

func getBrowserName() string {
	browserName, hasEnv := os.LookupEnv("BROWSER")
	if hasEnv {
		return browserName
	}
	return "chromium"
}

func newBrowserContextAndPage(t *testing.T, options playwright.BrowserNewContextOptions) (playwright.BrowserContext, playwright.Page) {
	t.Helper()
	context, err := browser.NewContext(options)
	if err != nil {
		t.Fatalf("could not create new context: %v", err)
	}
	t.Cleanup(func() {
		if err := context.Close(); err != nil {
			t.Errorf("could not close context: %v", err)
		}
	})
	page, err := context.NewPage()
	if err != nil {
		t.Fatalf("could not create new page: %v", err)
	}
	return context, page
}

// AssertToBeGolden compares the given image with a golden file and asserts that they are equal.
//
// Notes:
// - Golden files are stored in the "*-snapshots" directory in the same directory as the test file. e.g. "page_test.go" lead to "page-snapshots".
// - If the golden file does not exist, creates the golden file with the given image.
// - If the UPDATE_SNAPSHOTS environment variable is set, updates the golden file with the given image.
// - Use `pixelmatch.MatchOptions` to configure the pixelmatch algorithm.
func AssertToBeGolden(t *testing.T, img []byte, filename string, matchOptions ...pixelmatch.MatchOption) {
	t.Helper()

	actual, _, err := image.Decode(bytes.NewReader(img))
	if err != nil {
		t.Errorf("could not decode actual image: %v", err)
	}

	_, srcFile, _, _ := runtime.Caller(1)

	goldenDir := strings.Replace(srcFile, "_test.go", "-snapshots", 1)

	goldenPath := filepath.Join(goldenDir, getGoldenFilename(filename))

	if os.Getenv("UPDATE_SNAPSHOTS") != "" {
		if err := writeGoldenFile(goldenPath, img); err != nil {
			t.Errorf("could not write golden file: %v", err)
		} else {
			t.Logf("updated golden file: %s", goldenPath)
		}
		return
	}

	goldenRaw, err := os.ReadFile(goldenPath)
	if err != nil {
		// create golden file if it does not exist
		if os.IsNotExist(err) {
			if err := writeGoldenFile(goldenPath, img); err != nil {
				t.Errorf("could not write golden file: %v", err)
			} else {
				t.Logf("created golden file: %s", goldenPath)
			}
			return
		}
		t.Errorf("could not read golden file: %v", err)
	}

	golden, _, err := image.Decode(bytes.NewReader(goldenRaw))
	if err != nil {
		t.Errorf("could not decode golden: %v", err)
	}
	if actual.Bounds().Size() != golden.Bounds().Size() {
		t.Errorf("actual and golden have different sizes: %v != %v", actual.Bounds().Size(), golden.Bounds().Size())
	}
	diff, err := pixelmatch.MatchPixel(actual, golden, matchOptions...)
	if err != nil {
		t.Errorf("could not match pixel: %v", err)
	}
	if diff > 0 {
		t.Errorf("diff: %v", diff)
	}
}

func writeGoldenFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, data, 0o644)
}

func getGoldenFilename(name string) string {
	ext := filepath.Ext(name)
	return strings.TrimSuffix(name, ext) + "-" + browserName + ext
}
