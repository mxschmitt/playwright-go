package playwright

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testDriverHelperEnv  = "GO_WANT_PLAYWRIGHT_DRIVER_HELPER"
	testDriverVersionEnv = "PLAYWRIGHT_GO_TEST_DRIVER_VERSION"
)

func TestMain(m *testing.M) {
	if os.Getenv(testDriverHelperEnv) != "1" {
		os.Exit(m.Run())
	}
	version := os.Getenv(testDriverVersionEnv)
	validArgs := len(os.Args) >= 3 && filepath.Base(os.Args[len(os.Args)-2]) == "cli.js" && os.Args[len(os.Args)-1] == "--version"
	if version == "" || !validArgs {
		os.Exit(2)
	}
	fmt.Printf("Version %s\n", version)
	os.Exit(0)
}

func TestRunOptionsRedirectStderr(t *testing.T) {
	r, w := io.Pipe()
	var output string
	wg := &sync.WaitGroup{}
	readIOAsyncTilEOF(t, r, wg, &output)

	driverPath := t.TempDir()
	options := &RunOptions{
		Stderr:          w,
		DriverDirectory: driverPath,
		Browsers:        []string{},
		Verbose:         true,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer ts.Close()

	t.Setenv("PLAYWRIGHT_GO_NPM_REGISTRY", ts.URL)
	driver, err := NewDriver(options)
	require.NoError(t, err)
	err = driver.Install()
	require.Error(t, err)
	require.NoError(t, w.Close())
	wg.Wait()

	assert.Contains(t, output, "Downloading driver")
	require.Contains(t, output, fmt.Sprintf("path=%s", driverPath))
}

func TestRunOptions_OnlyInstallShell(t *testing.T) {
	if getBrowserName() != "chromium" {
		t.Skip("chromium only")
		return
	}

	r, w := io.Pipe()
	var output string
	wg := &sync.WaitGroup{}
	readIOAsyncTilEOF(t, r, wg, &output)

	driverPath := t.TempDir()
	driver, err := NewDriver(&RunOptions{
		Stdout:           w,
		DriverDirectory:  driverPath,
		Browsers:         []string{getBrowserName()},
		Verbose:          true,
		OnlyInstallShell: true,
		DryRun:           true,
	})
	require.NoError(t, err)
	browserPath := t.TempDir()

	t.Setenv("PLAYWRIGHT_BROWSERS_PATH", browserPath)

	err = driver.Install()
	require.NoError(t, err)
	require.NoError(t, w.Close())
	wg.Wait()

	assert.Contains(t, output, "chromium-headless-shell")
	assert.NotContains(t, output, "Chrome for Testing")
}

func TestDriverInstall(t *testing.T) {
	if _, err := nodePlatformSuffix(); err != nil {
		t.Skipf("bundled Node.js is not available on this platform: %v", err)
	}
	t.Setenv("PLAYWRIGHT_NODEJS_PATH", "")
	t.Setenv("PLAYWRIGHT_CLI_PATH", "")
	driverPath := t.TempDir()
	driver, err := NewDriver(&RunOptions{
		DriverDirectory: driverPath,
		Browsers:        []string{getBrowserName()},
		Verbose:         true,
	})
	if err != nil {
		t.Fatalf("could not start driver: %v", err)
	}
	browserPath := t.TempDir()
	t.Setenv("PLAYWRIGHT_BROWSERS_PATH", browserPath)
	err = driver.Install()
	if err != nil {
		t.Fatalf("could not install driver: %v", err)
	}
	requireDriverPackageArtifacts(t, driver, driverPath)
	err = driver.Uninstall()
	if err != nil {
		t.Fatalf("could not uninstall driver: %v", err)
	}
}

func TestNpmRegistryEnv(t *testing.T) {
	driverPath := t.TempDir()
	driver, err := NewDriver(&RunOptions{
		DriverDirectory:     driverPath,
		SkipInstallBrowsers: true,
	})
	if err != nil {
		t.Fatalf("could not start driver: %v", err)
	}
	uri := ""
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uri = r.URL.String()
		w.WriteHeader(404)
	}))
	defer ts.Close()

	t.Setenv("PLAYWRIGHT_GO_NPM_REGISTRY", ts.URL)
	err = driver.Install()
	if err == nil || !strings.Contains(err.Error(), "404 Not Found") || !strings.Contains(uri, "playwright-core") {
		t.Fatalf("PLAYWRIGHT_GO_NPM_REGISTRY does not work: %v", err)
	}
}

func TestNodePlatformSuffix(t *testing.T) {
	suffix, err := nodePlatformSuffix()
	switch runtime.GOARCH {
	case "amd64", "arm64":
		require.NoError(t, err)
		assert.NotEmpty(t, suffix)
	default:
		// e.g. linux/arm has no prebuilt Node.js binary on nodejs.org.
		require.Error(t, err)
		assert.Contains(t, err.Error(), "PLAYWRIGHT_NODEJS_PATH")
	}
}

func TestPatchDriverBundleMakesPageErrorLocationOptional(t *testing.T) {
	t.Setenv("PLAYWRIGHT_CLI_PATH", "")
	driverPath := t.TempDir()
	bundlePath := filepath.Join(driverPath, "package", "lib", "coreBundle.js")
	require.NoError(t, os.MkdirAll(filepath.Dir(bundlePath), 0o755))
	require.NoError(t, os.WriteFile(bundlePath, []byte(`location:{
url:pageError.location.url,
line: pageError.location.lineNumber,
column:    pageError.location.columnNumber
}`), 0o644))

	driver, err := NewDriver(&RunOptions{DriverDirectory: driverPath})
	require.NoError(t, err)
	require.NoError(t, driver.patchDriverBundle())
	require.NoError(t, driver.patchDriverBundle())

	data, err := os.ReadFile(bundlePath)
	require.NoError(t, err)
	require.Contains(t, string(data), `pageError.location?.url || ""`)
	require.Contains(t, string(data), `pageError.location?.lineNumber || 0`)
	require.Contains(t, string(data), `pageError.location?.columnNumber || 0`)
}

func TestPatchDriverBundleAcceptsMixedOriginalAndPatchedPatterns(t *testing.T) {
	t.Setenv("PLAYWRIGHT_CLI_PATH", "")
	driverPath := t.TempDir()
	bundlePath := filepath.Join(driverPath, "package", "lib", "coreBundle.js")
	require.NoError(t, os.MkdirAll(filepath.Dir(bundlePath), 0o755))
	require.NoError(t, os.WriteFile(bundlePath, []byte(`location:{
	url:pageError.location?.url || "",
	line: pageError.location.lineNumber,
	column: pageError.location?.columnNumber || 0
}`), 0o644))

	driver, err := NewDriver(&RunOptions{DriverDirectory: driverPath})
	require.NoError(t, err)
	require.NoError(t, driver.patchDriverBundle())

	data, err := os.ReadFile(bundlePath)
	require.NoError(t, err)
	require.Contains(t, string(data), `pageError.location?.url || ""`)
	require.Contains(t, string(data), `pageError.location?.lineNumber || 0`)
	require.Contains(t, string(data), `pageError.location?.columnNumber || 0`)
}

func TestPatchDriverBundleRequiresCoreBundle(t *testing.T) {
	t.Setenv("PLAYWRIGHT_CLI_PATH", "")
	driver, err := NewDriver(&RunOptions{DriverDirectory: t.TempDir()})
	require.NoError(t, err)

	err = driver.patchDriverBundle()
	require.Error(t, err)
	require.ErrorContains(t, err, "could not read driver bundle")
}

func TestPatchDriverBundleRejectsPartialPatternMismatch(t *testing.T) {
	t.Setenv("PLAYWRIGHT_CLI_PATH", "")
	driverPath := t.TempDir()
	bundlePath := filepath.Join(driverPath, "package", "lib", "coreBundle.js")
	require.NoError(t, os.MkdirAll(filepath.Dir(bundlePath), 0o755))
	original := []byte(`location:{
	url:pageError.location.url,
	line: pageError.location.lineNumber
}`)
	require.NoError(t, os.WriteFile(bundlePath, original, 0o644))

	driver, err := NewDriver(&RunOptions{DriverDirectory: driverPath})
	require.NoError(t, err)
	err = driver.patchDriverBundle()
	require.Error(t, err)
	require.ErrorContains(t, err, "pageError.location.columnNumber")

	data, readErr := os.ReadFile(bundlePath)
	require.NoError(t, readErr)
	require.Equal(t, original, data, "an incompatible bundle must not be partially rewritten")
}

func TestDownloadDriverExternalCLIDoesNotRequireManagedBundle(t *testing.T) {
	driverPath := filepath.Join(t.TempDir(), "driver-cache")
	externalCLI := filepath.Join(t.TempDir(), "cli.js")
	require.NoError(t, os.WriteFile(externalCLI, []byte("// externally managed"), 0o644))
	configureTestDriverRuntime(t, playwrightCliVersion)
	t.Setenv("PLAYWRIGHT_CLI_PATH", externalCLI)

	var registryRequests atomic.Int32
	registry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		registryRequests.Add(1)
		http.Error(w, "unexpected download", http.StatusInternalServerError)
	}))
	defer registry.Close()
	t.Setenv("PLAYWRIGHT_GO_NPM_REGISTRY", registry.URL)

	driver, err := NewDriver(&RunOptions{DriverDirectory: driverPath})
	require.NoError(t, err)
	require.NoError(t, driver.DownloadDriver())
	require.Zero(t, registryRequests.Load())
	_, err = os.Stat(driverPath)
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestDownloadDriverMissingExternalCLIFailsWithoutDownload(t *testing.T) {
	missingCLI := filepath.Join(t.TempDir(), "missing-cli.js")
	t.Setenv("PLAYWRIGHT_CLI_PATH", missingCLI)

	var registryRequests atomic.Int32
	registry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		registryRequests.Add(1)
		http.Error(w, "unexpected download", http.StatusInternalServerError)
	}))
	defer registry.Close()
	t.Setenv("PLAYWRIGHT_GO_NPM_REGISTRY", registry.URL)

	driver, err := NewDriver(&RunOptions{DriverDirectory: t.TempDir()})
	require.NoError(t, err)
	err = driver.DownloadDriver()
	require.EqualError(t, err, fmt.Sprintf("PLAYWRIGHT_CLI_PATH %q does not exist", missingCLI))
	require.Zero(t, registryRequests.Load())
}

func TestDownloadDriverExternalCLIWrongVersionFailsWithoutDownload(t *testing.T) {
	externalCLI := filepath.Join(t.TempDir(), "cli.js")
	require.NoError(t, os.WriteFile(externalCLI, []byte("// externally managed"), 0o644))
	configureTestDriverRuntime(t, "1.61.1")
	t.Setenv("PLAYWRIGHT_CLI_PATH", externalCLI)

	var registryRequests atomic.Int32
	registry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		registryRequests.Add(1)
		http.Error(w, "unexpected download", http.StatusInternalServerError)
	}))
	defer registry.Close()
	t.Setenv("PLAYWRIGHT_GO_NPM_REGISTRY", registry.URL)

	driver, err := NewDriver(&RunOptions{DriverDirectory: t.TempDir()})
	require.NoError(t, err)
	err = driver.DownloadDriver()
	require.ErrorContains(t, err, "driver exists but version not "+playwrightCliVersion)
	require.Zero(t, registryRequests.Load())
}

func TestDownloadDriverManagedCLIRequiresCoreBundle(t *testing.T) {
	driverPath := t.TempDir()
	cliPath := filepath.Join(driverPath, "package", "cli.js")
	require.NoError(t, os.MkdirAll(filepath.Dir(cliPath), 0o755))
	require.NoError(t, os.WriteFile(cliPath, []byte("// managed"), 0o644))
	configureTestDriverRuntime(t, playwrightCliVersion)
	t.Setenv("PLAYWRIGHT_CLI_PATH", "")

	driver, err := NewDriver(&RunOptions{DriverDirectory: driverPath})
	require.NoError(t, err)
	err = driver.DownloadDriver()
	require.ErrorContains(t, err, "could not read driver bundle")
}

func configureTestDriverRuntime(t *testing.T, version string) {
	t.Helper()
	testExecutable, err := os.Executable()
	require.NoError(t, err)
	t.Setenv("PLAYWRIGHT_NODEJS_PATH", testExecutable)
	t.Setenv(testDriverHelperEnv, "1")
	t.Setenv(testDriverVersionEnv, version)
}

func TestShouldNotHangWhenPlaywrightUnexpectedExit(t *testing.T) {
	if getBrowserName() != "chromium" {
		t.Skip("chromium only")
		return
	}

	pw, err := Run()
	require.NoError(t, err)
	defer func() {
		_ = pw.Stop()
	}()
	browser, err := pw.Chromium.Launch()
	require.NoError(t, err)
	context, err := browser.NewContext()
	require.NoError(t, err)

	// Get the process ID directly from Playwright
	pid := pw.Pid()
	require.NotZero(t, pid, "Playwright process PID should not be zero")

	// Kill the process
	err = killProcessByPid(pid)
	require.NoError(t, err)

	_, err = context.NewPage()
	require.Error(t, err)
}

func TestGetNodeExecutable(t *testing.T) {
	// When PLAYWRIGHT_NODEJS_PATH is set, use that path.
	err := os.Setenv("PLAYWRIGHT_NODEJS_PATH", "envDir/node.exe")
	require.NoError(t, err)

	executable := getNodeExecutable("testDirectory")
	assert.Equal(t, "envDir/node.exe", executable)

	err = os.Unsetenv("PLAYWRIGHT_NODEJS_PATH")
	require.NoError(t, err)

	executable = getNodeExecutable("testDirectory")
	assert.Contains(t, executable, "testDirectory")
}

func TestGetDriverCliJs(t *testing.T) {
	// When PLAYWRIGHT_CLI_PATH is set, use that path directly.
	t.Setenv("PLAYWRIGHT_CLI_PATH", "/custom/cli.js")
	assert.Equal(t, "/custom/cli.js", getDriverCliJs("testDirectory"))

	// Otherwise fall back to the assumed <DriverDirectory>/package/cli.js layout.
	require.NoError(t, os.Unsetenv("PLAYWRIGHT_CLI_PATH"))
	cliJs := getDriverCliJs("testDirectory")
	assert.Contains(t, cliJs, "testDirectory")
	assert.Contains(t, cliJs, "cli.js")
}

func killProcessByPid(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	if err := process.Kill(); err != nil {
		return err
	}
	return nil
}

func getBrowserName() string {
	browserName, hasEnv := os.LookupEnv("BROWSER")
	if hasEnv {
		return browserName
	}
	return "chromium"
}

func readIOAsyncTilEOF(t *testing.T, r *io.PipeReader, wg *sync.WaitGroup, output *string) {
	t.Helper()
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := bufio.NewReader(r)
		for {
			line, _, err := buf.ReadLine()
			if err == io.EOF {
				break
			}
			*output += string(line)
		}
		_ = r.Close()
	}()
}

func requireDriverPackageArtifacts(t *testing.T, driver *PlaywrightDriver, driverPath string) {
	t.Helper()
	cliPath := filepath.Join(driverPath, "package", "cli.js")
	coreBundlePath := filepath.Join(driverPath, "package", "lib", "coreBundle.js")
	codecPath := filepath.Join(driverPath, "package", "lib", "webp_codec.wasm")
	for _, expectedPath := range []string{cliPath, coreBundlePath, codecPath} {
		info, statErr := os.Stat(expectedPath)
		require.NoError(t, statErr, "installed driver must contain %s", expectedPath)
		require.False(t, info.IsDir(), "installed driver artifact must be a file: %s", expectedPath)
	}

	nodeOutput, err := exec.Command(getNodeExecutable(driverPath), "--version").CombinedOutput()
	require.NoError(t, err, "could not execute installed Node.js: %s", nodeOutput)
	require.Equal(t, "v"+nodeVersion, strings.TrimSpace(string(nodeOutput)))

	cliOutput, err := driver.Command("--version").CombinedOutput()
	require.NoError(t, err, "could not execute installed Playwright CLI: %s", cliOutput)
	require.Equal(t, "Version "+playwrightCliVersion, strings.TrimSpace(string(cliOutput)))
}
