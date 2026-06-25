package playwright_test

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestMouseMove(t *testing.T) {
	BeforeEach(t)

	if isWebKit {
		_, err := page.Evaluate(`() => new Promise(requestAnimationFrame)`)
		require.NoError(t, err)
	}
	require.NoError(t, page.Mouse().Move(100, 100))
	_, err := page.Evaluate(`() => {
    window['result'] = [];
    document.addEventListener('mousemove', event => {
      window['result'].push([event.clientX, event.clientY]);
    });
  }`)
	require.NoError(t, err)
	require.NoError(t, page.Mouse().Move(200, 300, playwright.MouseMoveOptions{
		Steps: playwright.Int(5),
	}))
	result, err := page.Evaluate("result")
	require.NoError(t, err)
	require.Equal(t, []interface{}([]interface{}{[]interface{}{120, 140}, []interface{}{140, 180}, []interface{}{160, 220}, []interface{}{180, 260}, []interface{}{200, 300}}), result)
}

func TestMouseDown(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`<button onmousedown="window.clicked=true"/>`))
	require.NoError(t, page.Locator("button").Hover())
	require.NoError(t, page.Mouse().Down())
	result, err := page.Evaluate("window.clicked")
	require.NoError(t, err)
	require.True(t, result.(bool))
}

func TestMouseUp(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`<button onmouseup="window.clicked=true"/>`))
	require.NoError(t, page.Locator("button").Hover())
	require.NoError(t, page.Mouse().Down())
	require.NoError(t, page.Mouse().Up())
	result, err := page.Evaluate("window.clicked")
	require.NoError(t, err)
	require.True(t, result.(bool))
}

func TestMouseClick(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`<button onclick="window.clicked=true" style="width: 500px; height: 500px;"/>`))
	require.NoError(t, page.Locator("button").Hover())
	require.NoError(t, page.Mouse().Click(100, 100))
	result, err := page.Evaluate("window.clicked")
	require.NoError(t, err)
	require.True(t, result.(bool))
}

func TestMouseDblclick(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`<button ondblclick="window.clicked=true" style="width: 500px; height: 500px;"/>`))
	require.NoError(t, page.Locator("button").Hover())
	require.NoError(t, page.Mouse().Dblclick(100, 100))
	result, err := page.Evaluate("window.clicked")
	require.NoError(t, err)
	require.True(t, result.(bool))
}

func TestMouseWheel(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`<div style="width: 5000px; height: 5000px;"></div>`))

	require.NoError(t, page.Mouse().Move(50, 60))
	require.NoError(t, page.Mouse().Wheel(0, 100))
	time.Sleep(500 * time.Millisecond)
	h, err := page.WaitForFunction(`window.scrollY === 100`, nil)
	require.NoError(t, err)
	value, err := h.JSONValue()
	require.NoError(t, err)
	require.True(t, value.(bool))
}

func TestKeyboardDown(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`<input onkeydown="window.clicked=true"/>`))
	require.NoError(t, page.Locator("input").Click())
	require.NoError(t, page.Keyboard().Down("Enter"))
	result, err := page.Evaluate("window.clicked")
	require.NoError(t, err)
	require.True(t, result.(bool))
}

func TestKeyboardUp(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`<input onkeyup="window.clicked=true"/>`))
	require.NoError(t, page.Locator("input").Click())
	require.NoError(t, page.Keyboard().Up("Enter"))
	result, err := page.Evaluate("window.clicked")
	require.NoError(t, err)
	require.True(t, result.(bool))
}

func TestKeyboardInsertText(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`<input oninput="window.clicked=true"/>`))
	require.NoError(t, page.Locator("input").Click())
	require.NoError(t, page.Keyboard().InsertText("abc123"))
	result, err := page.Evaluate("window.clicked")
	require.NoError(t, err)
	require.True(t, result.(bool))
}

func TestKeyboardType(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`<input oninput="window.clicked=true"/>`))
	require.NoError(t, page.Locator("input").Click())
	require.NoError(t, page.Keyboard().Type("abc123"))
	result, err := page.Evaluate("window.clicked")
	require.NoError(t, err)
	require.True(t, result.(bool))
}

func TestElementHandleType(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`<input oninput="window.clicked=true"/>`))
	require.NoError(t, page.Locator("input").Click())
	//nolint:staticcheck
	inputElement, err := page.QuerySelector("input")
	require.NoError(t, err)
	//nolint:staticcheck
	require.NoError(t, inputElement.Type("abc123"))
	result, err := page.Evaluate("window.clicked")
	require.NoError(t, err)
	require.True(t, result.(bool))
}

func TestElementHandleFill(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`<input oninput="window.clicked=true"/>`))
	require.NoError(t, page.Locator("input").Click())
	//nolint:staticcheck
	inputElement, err := page.QuerySelector("input")
	require.NoError(t, err)
	//nolint:staticcheck
	require.NoError(t, inputElement.Fill("abc123"))
	result, err := page.Evaluate("window.clicked")
	require.NoError(t, err)
	require.True(t, result.(bool))
}

func TestKeyboardInsertPress(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`<input onkeydown="window.clicked=true"/>`))
	require.NoError(t, page.Locator("input").Click())
	require.NoError(t, page.Keyboard().Press("A"))
	result, err := page.Evaluate("window.clicked")
	require.NoError(t, err)
	require.True(t, result.(bool))
}

func TestElementHandlePress(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`<input onkeydown="window.clicked=true"/>`))
	require.NoError(t, page.Locator("input").Click())
	//nolint:staticcheck
	inputElement, err := page.QuerySelector("input")
	require.NoError(t, err)
	//nolint:staticcheck
	require.NoError(t, inputElement.Press("A"))
	result, err := page.Evaluate("window.clicked")
	require.NoError(t, err)
	require.True(t, result.(bool))
}

func TestTouchscreenTap(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`<button onclick="window.clicked=true" style="width: 500px; height: 500px;"/>`))
	require.NoError(t, page.Touchscreen().Tap(100, 100))
	result, err := page.Evaluate("window.clicked")
	require.NoError(t, err)
	require.True(t, result.(bool))
}

func TestShouldUploadAFolder(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(fmt.Sprintf("%s%s", server.PREFIX, "/input/folderupload.html"))
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
	// WebKit can take longer than the default 30s action timeout to ingest a
	// directory upload on loaded CI runners. Upstream absorbs this flake via CI
	// retries (retries: 3); flakiness-go does not retry, so raise the per-action
	// timeout to give the slow WebKit step room instead.
	//nolint:staticcheck
	require.NoError(t, input.SetInputFiles(dir, playwright.ElementHandleSetInputFilesOptions{
		Timeout: playwright.Float(90 * 1000),
	}))

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

func TestShouldUploadAFolderAndThrowForMultipleDirectories(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(fmt.Sprintf("%s%s", server.PREFIX, "/input/folderupload.html"))
	require.NoError(t, err)

	input := page.Locator("input")
	dir := filepath.Join(t.TempDir(), "file-upload-test")
	require.NoError(t, os.MkdirAll(dir, 0o700))
	require.Nil(t, os.Mkdir(filepath.Join(dir, "folder1"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "folder1", "file1.txt"), []byte("file1 content"), 0o600))
	require.Nil(t, os.Mkdir(filepath.Join(dir, "folder2"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "folder2", "file2.txt"), []byte("file1content"), 0o600))

	err = input.SetInputFiles([]string{
		filepath.Join(dir, "folder1"),
		filepath.Join(dir, "folder2"),
	})

	require.ErrorContains(t, err, "Multiple directories are not supported")
}

func TestShouldThrowIfADirectoryAndFilesArePassed(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(fmt.Sprintf("%s%s", server.PREFIX, "/input/folderupload.html"))
	require.NoError(t, err)

	input := page.Locator("input")
	dir := filepath.Join(t.TempDir(), "file-upload-test")
	require.NoError(t, os.MkdirAll(dir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("file1 content"), 0o600))

	err = input.SetInputFiles([]string{
		dir,
		filepath.Join(dir, "file1.txt"),
	})

	require.ErrorContains(t, err, "File paths must be all files or a single directory")
}

func TestShouldThrowWhenUploadAFolderInANormalFileUploadInput(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(fmt.Sprintf("%s%s", server.PREFIX, "/input/fileupload.html"))
	require.NoError(t, err)

	input := page.Locator("input[type=file]")
	dir := filepath.Join(t.TempDir(), "file-upload-test")
	require.NoError(t, os.MkdirAll(dir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("file1 content"), 0o600))
	//nolint:staticcheck
	err = input.SetInputFiles(dir)
	require.ErrorContains(t, err, "File input does not support directories, pass individual files instead")
}

func TestSetInputFilesShouldRespectDefaultTimeout(t *testing.T) {
	BeforeEach(t)

	// Regression test: SetInputFiles must honor Page.SetDefaultTimeout instead of
	// always sending a hardcoded 30s timeout. The selector never resolves to an
	// element, so the call waits until the configured timeout elapses.
	page.SetDefaultTimeout(500)
	defer page.SetDefaultTimeout(30 * 1000) // reset

	file := filepath.Join(t.TempDir(), "file.txt")
	require.NoError(t, os.WriteFile(file, []byte("content"), 0o600))

	err := page.Locator("input#does-not-exist").SetInputFiles(file)
	require.ErrorContains(t, err, "Timeout 500ms exceeded")
}
