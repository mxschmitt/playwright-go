package playwright_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mxschmitt/playwright-go"
)

const writeOPFSScript = `async () => {
	const root = await navigator.storage.getDirectory();
	const nested = await root.getDirectoryHandle('nested', { create: true });
	await nested.getDirectoryHandle('empty', { create: true });
	const binary = await nested.getFileHandle('data.bin', { create: true });
	const binaryWritable = await binary.createWritable();
	await binaryWritable.write(new Uint8Array([0, 1, 2, 255]));
	await binaryWritable.close();
	const text = await root.getFileHandle('hello.txt', { create: true });
	const textWritable = await text.createWritable();
	await textWritable.write('Hello, world!');
	await textWritable.close();
}`

const readOPFSScript = `async () => {
	const root = await navigator.storage.getDirectory();
	const nested = await root.getDirectoryHandle('nested');
	const data = await (await nested.getFileHandle('data.bin')).getFile();
	const text = await (await root.getFileHandle('hello.txt')).getFile();
	await nested.getDirectoryHandle('empty');
	return { data: [...new Uint8Array(await data.arrayBuffer())], text: await text.text() };
}`

func TestBrowserContextStorageStateShouldRoundTripOPFS(t *testing.T) {
	if isWebKit {
		t.Skip("OPFS is unavailable in ephemeral WebKit contexts")
	}
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	_, err = page.Evaluate(writeOPFSScript)
	require.NoError(t, err)

	path := filepath.Join(t.TempDir(), "storage-state.json")
	_, err = context.StorageState(playwright.BrowserContextStorageStateOptions{Path: playwright.String(path)})
	require.NoError(t, err)
	written, err := os.ReadFile(path)
	require.NoError(t, err)
	require.JSONEq(t, `{"cookies":[],"origins":[]}`, string(written))

	_, err = context.StorageState(playwright.BrowserContextStorageStateOptions{
		Path: playwright.String(path),
		Opfs: playwright.Bool(true),
	})
	require.NoError(t, err)
	written, err = os.ReadFile(path)
	require.NoError(t, err)
	expected := `{"cookies":[],"origins":[{
		"origin": "` + server.PREFIX + `",
		"localStorage": [],
		"opfs": [
			{"path": "hello.txt", "type": "file", "base64": "SGVsbG8sIHdvcmxkIQ=="},
			{"path": "nested", "type": "directory"},
			{"path": "nested/data.bin", "type": "file", "base64": "AAEC/w=="},
			{"path": "nested/empty", "type": "directory"}
		]
	}]}`
	require.JSONEq(t, expected, string(written))

	requestPath := filepath.Join(t.TempDir(), "request-storage-state.json")
	_, err = context.Request().StorageState(playwright.APIRequestContextStorageStateOptions{
		Path: playwright.String(requestPath),
		Opfs: playwright.Bool(true),
	})
	require.NoError(t, err)
	requestWritten, err := os.ReadFile(requestPath)
	require.NoError(t, err)
	require.JSONEq(t, expected, string(requestWritten))

	context2, page2 := newBrowserContextAndPage(t, playwright.BrowserNewContextOptions{StorageStatePath: playwright.String(path)})
	_, err = page2.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	contents, err := page2.Evaluate(readOPFSScript)
	require.NoError(t, err)
	encoded, err := json.Marshal(contents)
	require.NoError(t, err)
	require.JSONEq(t, `{"data":[0,1,2,255],"text":"Hello, world!"}`, string(encoded))
	require.NoError(t, context2.Close())

	context3, page3 := newBrowserContextAndPage(t, playwright.BrowserNewContextOptions{})
	_, err = page3.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, context3.SetStorageState(path))
	contents, err = page3.Evaluate(readOPFSScript)
	require.NoError(t, err)
	encoded, err = json.Marshal(contents)
	require.NoError(t, err)
	require.JSONEq(t, `{"data":[0,1,2,255],"text":"Hello, world!"}`, string(encoded))
}

func TestBrowserContextStorageStateShouldRoundTripOPFSThroughTheReturnedValue(t *testing.T) {
	if isWebKit {
		t.Skip("OPFS is unavailable in ephemeral WebKit contexts")
	}
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	_, err = page.Evaluate(writeOPFSScript)
	require.NoError(t, err)

	state, err := context.StorageState(playwright.BrowserContextStorageStateOptions{Opfs: playwright.Bool(true)})
	require.NoError(t, err)
	require.Len(t, state.Origins, 1)
	require.Equal(t, []playwright.OPFSEntry{
		{Path: "hello.txt", Type: playwright.OPFSEntryTypeFile, Base64: playwright.String("SGVsbG8sIHdvcmxkIQ==")},
		{Path: "nested", Type: playwright.OPFSEntryTypeDirectory},
		{Path: "nested/data.bin", Type: playwright.OPFSEntryTypeFile, Base64: playwright.String("AAEC/w==")},
		{Path: "nested/empty", Type: playwright.OPFSEntryTypeDirectory},
	}, state.Origins[0].Opfs)

	requestState, err := context.Request().StorageState(playwright.APIRequestContextStorageStateOptions{Opfs: playwright.Bool(true)})
	require.NoError(t, err)
	require.Equal(t, state.Origins, requestState.Origins)

	_, page2 := newBrowserContextAndPage(t, playwright.BrowserNewContextOptions{StorageState: state.ToOptionalStorageState()})
	_, err = page2.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	contents, err := page2.Evaluate(readOPFSScript)
	require.NoError(t, err)
	encoded, err := json.Marshal(contents)
	require.NoError(t, err)
	require.JSONEq(t, `{"data":[0,1,2,255],"text":"Hello, world!"}`, string(encoded))
}
