package playwright_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestBrowserContextStorageStateShouldCaptureLocalStorage(t *testing.T) {
	BeforeEach(t)

	page1, err := context.NewPage()
	require.NoError(t, err)
	require.NoError(t, page1.Route("**/*", func(route playwright.Route) {
		require.NoError(t, route.Fulfill(playwright.RouteFulfillOptions{
			Body: "<html></html>",
		}))
	}))
	_, err = page1.Goto("https://www.example.com")
	require.NoError(t, err)
	_, err = page1.Evaluate("localStorage['name1'] = 'value1'")
	require.NoError(t, err)
	_, err = page1.Goto("https://www.domain.com")
	require.NoError(t, err)
	_, err = page1.Evaluate("localStorage['name2'] = 'value2'")
	require.NoError(t, err)

	state, err := context.StorageState()
	require.NoError(t, err)
	origins := state.Origins
	require.Equal(t, 2, len(origins))
	require.Equal(t, origins, []playwright.Origin{
		{
			Origin: "https://www.domain.com",
			LocalStorage: []playwright.NameValue{
				{
					Name:  "name2",
					Value: "value2",
				},
			},
		},
		{
			Origin: "https://www.example.com",
			LocalStorage: []playwright.NameValue{
				{
					Name:  "name1",
					Value: "value1",
				},
			},
		},
	})
}

func TestBrowserContextStorageStateSetLocalStorage(t *testing.T) {
	BeforeEach(t, playwright.BrowserNewContextOptions{
		StorageState: &playwright.OptionalStorageState{
			Origins: []playwright.Origin{
				{
					Origin: "https://www.example.com",
					LocalStorage: []playwright.NameValue{
						{
							Name:  "name1",
							Value: "value1",
						},
					},
				},
			},
		},
	})

	require.NoError(t, page.Route("**/*", func(route playwright.Route) {
		require.NoError(t, route.Fulfill(playwright.RouteFulfillOptions{
			Body: "<html></html>",
		}))
	}))
	_, err := page.Goto("https://www.example.com")
	require.NoError(t, err)
	localStorage, err := page.Evaluate("window.localStorage")
	require.NoError(t, err)
	require.Equal(t, map[string]interface{}{"name1": "value1"}, localStorage)
}

func TestBrowserContextStorageStateRoundTripThroughTheFile(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.Route("**/*", func(route playwright.Route) {
		require.NoError(t, route.Fulfill(playwright.RouteFulfillOptions{
			Body: "<html></html>",
		}))
	}))
	_, err := page.Goto("https://www.example.com")
	require.NoError(t, err)
	_, err = page.Evaluate(`
	() => {
		localStorage["name1"] = "value1"
		document.cookie = "username=John Doe"
		return document.cookie
	}
	`)
	require.NoError(t, err)
	tempfile, err := os.CreateTemp(os.TempDir(), "storage-state*.json")
	require.NoError(t, err)
	state, err := context.StorageState(tempfile.Name())
	require.NoError(t, err)
	stateWritten, err := os.ReadFile(tempfile.Name())
	require.NoError(t, err)
	var storageState *playwright.StorageState
	err = json.Unmarshal(stateWritten, &storageState)
	require.NoError(t, err)
	require.Equal(t, state, storageState)

	_, page2 := newBrowserContextAndPage(t, playwright.BrowserNewContextOptions{
		StorageStatePath: playwright.String(tempfile.Name()),
	})

	require.NoError(t, page2.Route("**/*", func(route playwright.Route) {
		require.NoError(t, route.Fulfill(playwright.RouteFulfillOptions{
			Body: "<html></html>",
		}))
	}))
	_, err = page2.Goto("https://www.example.com")
	require.NoError(t, err)
	cookie, err := page2.Evaluate("document.cookie")
	require.NoError(t, err)
	require.Equal(t, "username=John Doe", cookie)
	localStorage, err := page2.Evaluate("window.localStorage")
	require.NoError(t, err)
	require.Equal(t, map[string]interface{}{"name1": "value1"}, localStorage)
}

func TestBrowserContextStorageStateRoundTripThroughConvert(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.Route("**/*", func(route playwright.Route) {
		require.NoError(t, route.Fulfill(playwright.RouteFulfillOptions{
			Body: "<html></html>",
		}))
	}))
	_, err := page.Goto("https://www.example.com")
	require.NoError(t, err)
	_, err = page.Evaluate(`
	() => {
		localStorage["name1"] = "value1"
		document.cookie = "username=John Doe"
		return document.cookie
	}
	`)
	require.NoError(t, err)

	storageState, err := context.StorageState()
	require.NoError(t, err)

	_, page2 := newBrowserContextAndPage(t,
		playwright.BrowserNewContextOptions{
			StorageState: storageState.ToOptionalStorageState(),
		})
	require.NoError(t, page2.Route("**/*", func(route playwright.Route) {
		require.NoError(t, route.Fulfill(playwright.RouteFulfillOptions{
			Body: "<html></html>",
		}))
	}))
	_, err = page2.Goto("https://www.example.com")
	require.NoError(t, err)
	cookie, err := page2.Evaluate("document.cookie")
	require.NoError(t, err)
	require.Equal(t, "username=John Doe", cookie)
	localStorage, err := page2.Evaluate("window.localStorage")
	require.NoError(t, err)
	require.Equal(t, map[string]interface{}{"name1": "value1"}, localStorage)
}
