package playwright_test

import (
	"os"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestFileChooser(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.PREFIX + "/input/fileupload.html")
	require.NoError(t, err)
	//nolint:staticcheck
	input, err := page.QuerySelector("input")
	require.NoError(t, err)
	file, err := os.ReadFile(Asset("file-to-upload.txt"))
	require.NoError(t, err)
	//nolint:staticcheck
	require.NoError(t, input.SetInputFiles([]playwright.InputFile{
		{
			Name:     "file-to-upload.txt",
			MimeType: "text/plain",
			Buffer:   file,
		},
	}))
	fileName, err := page.Evaluate("e => e.files[0].name", input)
	require.NoError(t, err)
	require.Equal(t, "file-to-upload.txt", fileName)
	content, err := page.Evaluate(`e => {
        reader = new FileReader()
        promise = new Promise(fulfill => reader.onload = fulfill)
        reader.readAsText(e.files[0])
        return promise.then(() => reader.result)
    }`, input)
	require.NoError(t, err)
	require.Equal(t, "contents of the file", content)
}

func TestFileChooserShouldEmitEvent(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent("<input type=file>"))

	fileChooser, err := page.ExpectFileChooser(func() error {
		return page.Locator("input").Click()
	})
	require.NoError(t, err)
	require.False(t, fileChooser.IsMultiple())
	require.Equal(t, page, fileChooser.Page())
	//nolint:staticcheck
	elementHTML, err := fileChooser.Element().InnerHTML()
	require.NoError(t, err)
	//nolint:staticcheck
	inputElement, err := page.QuerySelector("input")
	require.NoError(t, err)
	//nolint:staticcheck
	inputElementHTML, err := inputElement.InnerHTML()
	require.NoError(t, err)

	require.Equal(t, elementHTML, inputElementHTML)

	require.NoError(t, fileChooser.SetFiles([]playwright.InputFile{
		{
			Name:     "file-to-upload.txt",
			MimeType: "text/plain",
			Buffer:   []byte("123"),
		},
	}))
	fileName, err := page.Evaluate("e => e.files[0].name", inputElement)
	require.NoError(t, err)
	require.Equal(t, "file-to-upload.txt", fileName)
}
