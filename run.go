package playwright

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const playwrightCliVersion = "1.60.0"

var (
	logger               = slog.Default()
	playwrightCDNMirrors = []string{
		"https://playwright.azureedge.net",
		"https://playwright-akamai.azureedge.net",
		"https://playwright-verizon.azureedge.net",
	}
)

// PlaywrightDriver wraps the Playwright CLI of upstream Playwright.
//
// It's required for playwright-go to work.
type PlaywrightDriver struct {
	Version string
	options *RunOptions
}

func NewDriver(options ...*RunOptions) (*PlaywrightDriver, error) {
	transformed, err := transformRunOptions(options...) // get default values
	if err != nil {
		return nil, err
	}
	return &PlaywrightDriver{
		options: transformed,
		Version: playwrightCliVersion,
	}, nil
}

func getDefaultCacheDirectory() (string, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get user home directory: %w", err)
	}
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(userHomeDir, "AppData", "Local"), nil
	case "darwin":
		return filepath.Join(userHomeDir, "Library", "Caches"), nil
	case "linux":
		return filepath.Join(userHomeDir, ".cache"), nil
	}
	return "", errors.New("could not determine cache directory")
}

func (d *PlaywrightDriver) isUpToDateDriver() (bool, error) {
	if _, err := os.Stat(d.options.DriverDirectory); os.IsNotExist(err) {
		if err := os.MkdirAll(d.options.DriverDirectory, 0o777); err != nil {
			return false, fmt.Errorf("could not create driver directory: %w", err)
		}
	}
	if _, err := os.Stat(getDriverCliJs(d.options.DriverDirectory)); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("could not check if driver is up2date: %w", err)
	}
	cmd := d.Command("--version")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("could not run driver: %w", err)
	}
	if bytes.Contains(output, []byte(d.Version)) {
		return true, nil
	}
	// avoid triggering downloads and accidentally overwriting files
	return false, fmt.Errorf("driver exists but version not %s in : %s", d.Version, d.options.DriverDirectory)
}

// Command returns an exec.Cmd for the driver.
func (d *PlaywrightDriver) Command(arg ...string) *exec.Cmd {
	cmd := exec.Command(getNodeExecutable(d.options.DriverDirectory), append([]string{getDriverCliJs(d.options.DriverDirectory)}, arg...)...)
	cmd.SysProcAttr = defaultSysProcAttr
	return cmd
}

// Install downloads the driver and the browsers depending on [RunOptions].
func (d *PlaywrightDriver) Install() error {
	if err := d.DownloadDriver(); err != nil {
		return fmt.Errorf("could not install driver: %w", err)
	}
	if d.options.SkipInstallBrowsers {
		return nil
	}

	d.log("Downloading browsers...")
	if err := d.installBrowsers(); err != nil {
		return fmt.Errorf("could not install browsers: %w", err)
	}
	d.log("Downloaded browsers successfully")

	return nil
}

// Uninstall removes the driver and the browsers.
func (d *PlaywrightDriver) Uninstall() error {
	d.log("Removing browsers...")
	if err := d.uninstallBrowsers(); err != nil {
		return fmt.Errorf("could not uninstall browsers: %w", err)
	}

	d.log("Removing driver...")
	if err := os.RemoveAll(d.options.DriverDirectory); err != nil {
		return fmt.Errorf("could not remove driver directory: %w", err)
	}

	d.log("Uninstall driver successfully")
	return nil
}

// DownloadDriver downloads the driver only
func (d *PlaywrightDriver) DownloadDriver() error {
	up2Date, err := d.isUpToDateDriver()
	if err != nil {
		return err
	}
	if up2Date {
		return d.patchDriverBundle()
	}

	d.log("Downloading driver", "path", d.options.DriverDirectory)

	body, err := downloadDriver(d.getDriverURLs())
	if err != nil {
		return err
	}
	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return fmt.Errorf("could not read zip content: %w", err)
	}

	for _, zipFile := range zipReader.File {
		zipFileDiskPath := filepath.Join(d.options.DriverDirectory, zipFile.Name)
		if zipFile.FileInfo().IsDir() {
			if err := os.MkdirAll(zipFileDiskPath, os.ModePerm); err != nil {
				return fmt.Errorf("could not create directory: %w", err)
			}
			continue
		}

		outFile, err := os.Create(zipFileDiskPath)
		if err != nil {
			return fmt.Errorf("could not create driver: %w", err)
		}
		file, err := zipFile.Open()
		if err != nil {
			return fmt.Errorf("could not open zip file: %w", err)
		}
		if _, err = io.Copy(outFile, file); err != nil {
			return fmt.Errorf("could not copy response body to file: %w", err)
		}
		if err := outFile.Close(); err != nil {
			return fmt.Errorf("could not close file (driver): %w", err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("could not close file (zip file): %w", err)
		}
		if zipFile.Mode().Perm()&0o100 != 0 && runtime.GOOS != "windows" {
			if err := makeFileExecutable(zipFileDiskPath); err != nil {
				return fmt.Errorf("could not make executable: %w", err)
			}
		}
	}

	d.log("Downloaded driver successfully")

	return d.patchDriverBundle()
}

func (d *PlaywrightDriver) patchDriverBundle() error {
	coreBundlePath := filepath.Join(d.options.DriverDirectory, "package", "lib", "coreBundle.js")
	data, err := os.ReadFile(coreBundlePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("could not read driver bundle: %w", err)
	}

	replacements := map[string]string{
		"pageError.location.url":          `pageError.location?.url || ""`,
		"pageError.location.lineNumber":   "pageError.location?.lineNumber || 0",
		"pageError.location.columnNumber": "pageError.location?.columnNumber || 0",
	}
	changed := false
	for original, patched := range replacements {
		originalBytes := []byte(original)
		patchedBytes := []byte(patched)
		if bytes.Contains(data, originalBytes) {
			data = bytes.ReplaceAll(data, originalBytes, patchedBytes)
			changed = true
		}
	}
	if !changed {
		alreadyPatched := true
		for _, patched := range replacements {
			if !bytes.Contains(data, []byte(patched)) {
				alreadyPatched = false
				break
			}
		}
		if alreadyPatched {
			return nil
		}
		return fmt.Errorf("could not patch driver bundle: pageError location pattern not found")
	}
	if err := os.WriteFile(coreBundlePath, data, 0o644); err != nil {
		return fmt.Errorf("could not write patched driver bundle: %w", err)
	}
	return nil
}

func (d *PlaywrightDriver) log(msg string, args ...any) {
	if d.options.Verbose {
		logger.Info(msg, args...)
	}
}

func (d *PlaywrightDriver) run() (*connection, error) {
	transport, err := newPipeTransport(d, d.options.Stderr)
	if err != nil {
		return nil, err
	}
	connection := newConnection(transport)
	return connection, nil
}

func (d *PlaywrightDriver) installBrowsers() error {
	additionalArgs := []string{"install"}
	if d.options.Browsers != nil {
		additionalArgs = append(additionalArgs, d.options.Browsers...)
	}

	if d.options.OnlyInstallShell {
		additionalArgs = append(additionalArgs, "--only-shell")
	}

	if d.options.NoInstallShell {
		additionalArgs = append(additionalArgs, "--no-shell")
	}

	if d.options.DryRun {
		additionalArgs = append(additionalArgs, "--dry-run")
	}

	if d.options.WithDeps {
		additionalArgs = append(additionalArgs, "--with-deps")
	}

	cmd := d.Command(additionalArgs...)
	cmd.Stdout = d.options.Stdout
	cmd.Stderr = d.options.Stderr
	return cmd.Run()
}

func (d *PlaywrightDriver) uninstallBrowsers() error {
	cmd := d.Command("uninstall")
	cmd.Stdout = d.options.Stdout
	cmd.Stderr = d.options.Stderr
	return cmd.Run()
}

// RunOptions are custom options to run the driver
type RunOptions struct {
	// DriverDirectory points to the playwright driver directory.
	// It should have two subdirectories: node and package.
	// You can also specify it using the environment variable PLAYWRIGHT_DRIVER_PATH.
	//
	// Default is user cache directory + "/ms-playwright-go/x.xx.xx":
	//  - Windows: %USERPROFILE%\AppData\Local
	//  - macOS: ~/Library/Caches
	//  - Linux: ~/.cache
	DriverDirectory string
	// OnlyInstallShell only downloads the headless shell. (For chromium browsers only)
	OnlyInstallShell bool
	// NoInstallShell does not install chromium headless shell. (For chromium browsers only)
	NoInstallShell      bool
	SkipInstallBrowsers bool
	// if not set and SkipInstallBrowsers is false, will download all browsers (chromium, firefox, webkit)
	Browsers []string
	// install system dependencies for browsers
	WithDeps bool
	Verbose  bool // default true
	Stdout   io.Writer
	Stderr   io.Writer
	Logger   *slog.Logger
	// DryRun does not install browser/dependencies. It will only print information.
	DryRun bool
}

// Install does download the driver and the browsers.
//
// Use this before playwright.Run() or use playwright cli to install the driver and browsers
func Install(options ...*RunOptions) error {
	driver, err := NewDriver(options...)
	if err != nil {
		return fmt.Errorf("could not get driver instance: %w", err)
	}
	if err := driver.Install(); err != nil {
		return fmt.Errorf("could not install driver: %w", err)
	}
	return nil
}

// Run starts a Playwright instance.
//
// Requires the driver and the browsers to be installed before.
// Either use Install() or use playwright cli.
func Run(options ...*RunOptions) (*Playwright, error) {
	driver, err := NewDriver(options...)
	if err != nil {
		return nil, fmt.Errorf("could not get driver instance: %w", err)
	}
	up2date, err := driver.isUpToDateDriver()
	if err != nil || !up2date {
		ferr := fmt.Errorf("please install the driver (v%s) first", playwrightCliVersion)
		if err != nil {
			ferr = fmt.Errorf("%w: %w", ferr, err)
		}
		return nil, ferr
	}
	connection, err := driver.run()
	if err != nil {
		return nil, err
	}
	playwright, err := connection.Start()
	return playwright, err
}

func transformRunOptions(options ...*RunOptions) (*RunOptions, error) {
	option := &RunOptions{
		Verbose: true,
	}
	if len(options) == 1 {
		option = options[0]
	}
	if option.OnlyInstallShell && option.NoInstallShell {
		return nil, fmt.Errorf("OnlyInstallShell and NoInstallShell cannot be set at the same time")
	}
	if option.DriverDirectory == "" { // if user did not set it, try to get it from env
		option.DriverDirectory = os.Getenv("PLAYWRIGHT_DRIVER_PATH")
	}
	if option.DriverDirectory == "" {
		cacheDirectory, err := getDefaultCacheDirectory()
		if err != nil {
			return nil, fmt.Errorf("could not get default cache directory: %w", err)
		}
		option.DriverDirectory = filepath.Join(cacheDirectory, "ms-playwright-go", playwrightCliVersion)
	}
	if option.Stdout == nil {
		option.Stdout = os.Stdout
	}
	if option.Stderr == nil {
		option.Stderr = os.Stderr
	} else if option.Logger == nil {
		log.SetOutput(option.Stderr)
	}
	if option.Logger != nil {
		logger = option.Logger
	}
	return option, nil
}

func getNodeExecutable(driverDirectory string) string {
	envPath := os.Getenv("PLAYWRIGHT_NODEJS_PATH")
	if envPath != "" {
		return envPath
	}

	node := "node"
	if runtime.GOOS == "windows" {
		node = "node.exe"
	}
	return filepath.Join(driverDirectory, node)
}

func getDriverCliJs(driverDirectory string) string {
	return filepath.Join(driverDirectory, "package", "cli.js")
}

func (d *PlaywrightDriver) getDriverURLs() []string {
	platform := ""
	switch runtime.GOOS {
	case "windows":
		platform = "win32_x64"
	case "darwin":
		if runtime.GOARCH == "arm64" {
			platform = "mac-arm64"
		} else {
			platform = "mac"
		}
	case "linux":
		if runtime.GOARCH == "arm64" {
			platform = "linux-arm64"
		} else {
			platform = "linux"
		}
	}

	baseURLs := []string{}
	pattern := "%s/builds/driver/playwright-%s-%s.zip"
	if !d.isReleaseVersion() {
		pattern = "%s/builds/driver/next/playwright-%s-%s.zip"
	}

	if hostEnv := os.Getenv("PLAYWRIGHT_DOWNLOAD_HOST"); hostEnv != "" {
		baseURLs = append(baseURLs, fmt.Sprintf(pattern, hostEnv, d.Version, platform))
	} else {
		for _, mirror := range playwrightCDNMirrors {
			baseURLs = append(baseURLs, fmt.Sprintf(pattern, mirror, d.Version, platform))
		}
	}
	return baseURLs
}

// isReleaseVersion checks if the version is not a beta or alpha release
// this helps to determine the url from where to download the driver
func (d *PlaywrightDriver) isReleaseVersion() bool {
	return !strings.Contains(d.Version, "beta") && !strings.Contains(d.Version, "alpha") && !strings.Contains(d.Version, "next")
}

func makeFileExecutable(path string) error {
	stats, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("could not stat driver: %w", err)
	}
	if err := os.Chmod(path, stats.Mode()|0x40); err != nil {
		return fmt.Errorf("could not set permissions: %w", err)
	}
	return nil
}

func downloadDriver(driverURLs []string) (body []byte, e error) {
	for _, driverURL := range driverURLs {
		resp, err := http.Get(driverURL)
		if err != nil {
			e = errors.Join(e, fmt.Errorf("could not download driver from %s: %w", driverURL, err))
			continue
		}
		defer resp.Body.Close() //nolint:errcheck
		if resp.StatusCode != http.StatusOK {
			e = errors.Join(e, fmt.Errorf("error: got non 200 status code: %d (%s) from %s", resp.StatusCode, resp.Status, driverURL))
			continue
		}
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			e = errors.Join(e, fmt.Errorf("could not read response body: %w", err))
			continue
		}
		return body, nil
	}
	return nil, e
}
