package playwright

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
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

const (
	playwrightCliVersion = "1.61.1"
	// nodeVersion is the Node.js runtime downloaded alongside the driver when no
	// PLAYWRIGHT_NODEJS_PATH is provided. It is kept in line with the Node.js
	// version upstream Playwright bundles in its own driver.
	nodeVersion = "24.18.0"

	// defaultNpmRegistry serves the platform-independent playwright-core package.
	// Override with the PLAYWRIGHT_GO_NPM_REGISTRY environment variable.
	defaultNpmRegistry = "https://registry.npmjs.org"
	// defaultNodejsDistHost serves the per-platform Node.js binaries.
	// Override with the NODE_MIRROR environment variable (nvm/n convention).
	defaultNodejsDistHost = "https://nodejs.org/dist"
)

var logger = slog.Default()

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

// DownloadDriver downloads the driver only.
//
// The driver is assembled from two upstream sources instead of the (now
// deprecated) Playwright CDN:
//   - the platform-independent playwright-core package from the npm registry,
//     extracted into <DriverDirectory>/package (this contains cli.js); and
//   - the matching per-platform Node.js binary from nodejs.org, placed at
//     <DriverDirectory>/node[.exe].
//
// When PLAYWRIGHT_NODEJS_PATH is set the Node.js download is skipped and the
// preinstalled Node.js is used instead, which also covers platforms for which
// nodejs.org has no prebuilt binary (e.g. linux/arm).
func (d *PlaywrightDriver) DownloadDriver() error {
	up2Date, err := d.isUpToDateDriver()
	if err != nil {
		return err
	}
	if up2Date {
		return d.patchDriverBundle()
	}

	d.log("Downloading driver", "path", d.options.DriverDirectory)

	if err := d.downloadPlaywrightPackage(); err != nil {
		return err
	}
	if err := d.downloadNode(); err != nil {
		return err
	}

	d.log("Downloaded driver successfully")

	return d.patchDriverBundle()
}

// downloadPlaywrightPackage downloads the platform-independent playwright-core
// package from the npm registry and extracts its "package/" contents into the
// driver directory, so that <DriverDirectory>/package/cli.js exists.
func (d *PlaywrightDriver) downloadPlaywrightPackage() error {
	url := fmt.Sprintf("%s/playwright-core/-/playwright-core-%s.tgz", npmRegistry(), d.Version)
	body, err := downloadWithRetry(url)
	if err != nil {
		return fmt.Errorf("could not download playwright-core: %w", err)
	}

	gzReader, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("could not read playwright-core archive: %w", err)
	}
	defer gzReader.Close() //nolint:errcheck

	tarReader := tar.NewReader(gzReader)
	extracted := false
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("could not read playwright-core archive: %w", err)
		}
		// npm tarballs nest everything under a top-level "package/" directory,
		// which is exactly the layout the driver expects on disk.
		if header.Typeflag != tar.TypeReg || !strings.HasPrefix(header.Name, "package/") {
			continue
		}
		diskPath, err := safeJoin(d.options.DriverDirectory, header.Name)
		if err != nil {
			return err
		}
		if err := writeFileFromReader(diskPath, tarReader, header.FileInfo().Mode()); err != nil {
			return err
		}
		extracted = true
	}
	if !extracted {
		return fmt.Errorf("no files extracted from playwright-core %s", d.Version)
	}
	return nil
}

// downloadNode downloads the per-platform Node.js binary from nodejs.org and
// places it at <DriverDirectory>/node[.exe]. It is a no-op when
// PLAYWRIGHT_NODEJS_PATH is set, since a preinstalled Node.js is used then.
func (d *PlaywrightDriver) downloadNode() error {
	if os.Getenv("PLAYWRIGHT_NODEJS_PATH") != "" {
		d.log("Skipping Node.js download, using PLAYWRIGHT_NODEJS_PATH")
		return nil
	}

	suffix, err := nodePlatformSuffix()
	if err != nil {
		return err
	}

	archiveDir := fmt.Sprintf("node-v%s-%s", nodeVersion, suffix)
	isWindows := runtime.GOOS == "windows"
	ext := "tar.gz"
	if isWindows {
		ext = "zip"
	}
	url := fmt.Sprintf("%s/v%s/%s.%s", nodejsDistHost(), nodeVersion, archiveDir, ext)

	body, err := downloadWithRetry(url)
	if err != nil {
		return fmt.Errorf("could not download Node.js: %w", err)
	}

	nodeDiskPath := getNodeExecutable(d.options.DriverDirectory)
	if isWindows {
		// The Windows archive is a zip with node.exe at "<archiveDir>/node.exe".
		return extractZipEntry(body, archiveDir+"/node.exe", nodeDiskPath)
	}
	// Unix archives are gzipped tars with the binary at "<archiveDir>/bin/node".
	return extractTarGzEntry(body, archiveDir+"/bin/node", nodeDiskPath)
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
	// The driver is assembled from the playwright-core npm package and the
	// matching Node.js release. The following environment variables tune this:
	//  - PLAYWRIGHT_NODEJS_PATH: use a preinstalled Node.js and skip downloading
	//    one. Required on platforms without a prebuilt Node.js binary (e.g.
	//    linux/arm).
	//  - PLAYWRIGHT_CLI_PATH: use a preinstalled cli.js directly, bypassing the
	//    assumed <DriverDirectory>/package/cli.js layout. Useful for
	//    distro-packaged drivers (e.g. the official NixOS playwright-driver,
	//    which keeps cli.js at the package root).
	//  - PLAYWRIGHT_GO_NPM_REGISTRY: npm registry mirror for playwright-core
	//    (default https://registry.npmjs.org).
	//  - NODE_MIRROR: Node.js distribution mirror (default https://nodejs.org/dist).
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
	// Allow pointing directly at cli.js, e.g. for distro-packaged drivers whose
	// layout differs from the assembled bundle (the official NixOS
	// playwright-driver keeps cli.js at the package root, not under package/).
	// This mirrors PLAYWRIGHT_NODEJS_PATH for the Node.js binary.
	if envPath := os.Getenv("PLAYWRIGHT_CLI_PATH"); envPath != "" {
		return envPath
	}
	return filepath.Join(driverDirectory, "package", "cli.js")
}

func npmRegistry() string {
	if host := os.Getenv("PLAYWRIGHT_GO_NPM_REGISTRY"); host != "" {
		return strings.TrimRight(host, "/")
	}
	return defaultNpmRegistry
}

func nodejsDistHost() string {
	if host := os.Getenv("NODE_MIRROR"); host != "" {
		return strings.TrimRight(host, "/")
	}
	return defaultNodejsDistHost
}

// nodePlatformSuffix maps the current GOOS/GOARCH to the suffix nodejs.org uses
// in its release archive names (e.g. "linux-x64", "darwin-arm64", "win-x64").
// Platforms without a prebuilt Node.js binary (such as linux/arm, 32-bit ARM)
// return an actionable error pointing at PLAYWRIGHT_NODEJS_PATH.
func nodePlatformSuffix() (string, error) {
	var os_ string
	switch runtime.GOOS {
	case "windows":
		os_ = "win"
	case "darwin":
		os_ = "darwin"
	case "linux":
		os_ = "linux"
	default:
		return "", unsupportedNodePlatformError()
	}

	var arch string
	switch runtime.GOARCH {
	case "amd64":
		arch = "x64"
	case "arm64":
		arch = "arm64"
	default:
		// Notably linux/arm (32-bit, e.g. Raspberry Pi armv7l): nodejs.org no
		// longer ships a prebuilt binary, so we cannot download one.
		return "", unsupportedNodePlatformError()
	}

	return fmt.Sprintf("%s-%s", os_, arch), nil
}

func unsupportedNodePlatformError() error {
	return fmt.Errorf("no prebuilt Node.js %s is available for %s/%s; "+
		"install Node.js yourself and set PLAYWRIGHT_NODEJS_PATH to its path",
		nodeVersion, runtime.GOOS, runtime.GOARCH)
}

// safeJoin joins an archive entry name onto root, guarding against path
// traversal ("zip slip"/"tar slip"): the resulting path must stay within root.
// This matters because PLAYWRIGHT_GO_NPM_REGISTRY / NODE_MIRROR allow arbitrary
// mirrors, so archive contents are not fully trusted.
func safeJoin(root, name string) (string, error) {
	diskPath := filepath.Join(root, filepath.FromSlash(name))
	prefix := filepath.Clean(root) + string(os.PathSeparator)
	if diskPath != filepath.Clean(root) && !strings.HasPrefix(diskPath, prefix) {
		return "", fmt.Errorf("invalid path in archive: %s", name)
	}
	return diskPath, nil
}

// writeFileFromReader writes the contents of r to diskPath, creating parent
// directories as needed and preserving the executable bit on non-Windows hosts.
func writeFileFromReader(diskPath string, r io.Reader, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(diskPath), 0o777); err != nil {
		return fmt.Errorf("could not create directory: %w", err)
	}
	outFile, err := os.Create(diskPath)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	if _, err := io.Copy(outFile, r); err != nil {
		outFile.Close() //nolint:errcheck
		return fmt.Errorf("could not write file: %w", err)
	}
	if err := outFile.Close(); err != nil {
		return fmt.Errorf("could not close file: %w", err)
	}
	if mode.Perm()&0o100 != 0 && runtime.GOOS != "windows" {
		if err := makeFileExecutable(diskPath); err != nil {
			return err
		}
	}
	return nil
}

// extractTarGzEntry extracts a single named entry from a gzipped tar archive to
// diskPath and marks it executable.
func extractTarGzEntry(archive []byte, entryName, diskPath string) error {
	gzReader, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return fmt.Errorf("could not read archive: %w", err)
	}
	defer gzReader.Close() //nolint:errcheck

	tarReader := tar.NewReader(gzReader)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("could not read archive: %w", err)
		}
		if header.Name != entryName {
			continue
		}
		// Force the executable bit: the node binary must be runnable.
		return writeFileFromReader(diskPath, tarReader, header.FileInfo().Mode()|0o100)
	}
	return fmt.Errorf("could not find %s in archive", entryName)
}

// extractZipEntry extracts a single named entry from a zip archive to diskPath.
func extractZipEntry(archive []byte, entryName, diskPath string) error {
	zipReader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return fmt.Errorf("could not read archive: %w", err)
	}
	for _, file := range zipReader.File {
		if file.Name != entryName {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("could not open zip entry: %w", err)
		}
		err = writeFileFromReader(diskPath, rc, file.Mode())
		rc.Close() //nolint:errcheck
		return err
	}
	return fmt.Errorf("could not find %s in archive", entryName)
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

// downloadWithRetry downloads url, retrying a few times on transient failures.
// It does not retry client errors (4xx), which are not transient.
func downloadWithRetry(url string) ([]byte, error) {
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		body, retryable, err := download(url)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retryable {
			break
		}
	}
	return nil, lastErr
}

// download fetches url. The returned bool reports whether a failure is worth
// retrying (network errors and 5xx are; 4xx are not).
func download(url string) ([]byte, bool, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, true, fmt.Errorf("could not download from %s: %w", url, err)
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		retryable := resp.StatusCode >= 500
		return nil, retryable, fmt.Errorf("got non 200 status code: %d (%s) from %s", resp.StatusCode, resp.Status, url)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, true, fmt.Errorf("could not read response body: %w", err)
	}
	return body, false, nil
}
