package playwright_test

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestPageExposesWebStorageProperties(t *testing.T) {
	BeforeEach(t)

	require.NotNil(t, page.LocalStorage())
	require.NotNil(t, page.SessionStorage())
	// The same instance is returned on each access.
	require.Same(t, page.LocalStorage(), page.LocalStorage())
	require.Same(t, page.SessionStorage(), page.SessionStorage())
}

func TestPageLocalStorageSetGetAndItems(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)

	storage := page.LocalStorage()
	require.NoError(t, storage.SetItem("foo", "bar"))

	value, err := storage.GetItem("foo")
	require.NoError(t, err)
	require.Equal(t, "bar", value)

	jsValue, err := page.Evaluate("() => localStorage.getItem('foo')")
	require.NoError(t, err)
	require.Equal(t, "bar", jsValue)

	require.NoError(t, storage.SetItem("baz", "qux"))
	items, err := storage.Items()
	require.NoError(t, err)
	require.Contains(t, items, playwright.WebStorageItem{Name: "foo", Value: "bar"})
	require.Contains(t, items, playwright.WebStorageItem{Name: "baz", Value: "qux"})
}

func TestPageLocalStorageGetItemMissing(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)

	value, err := page.LocalStorage().GetItem("missing")
	require.NoError(t, err)
	require.Equal(t, "", value)
}

func TestPageLocalStorageRemoveAndClear(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)

	storage := page.LocalStorage()
	require.NoError(t, storage.SetItem("foo", "bar"))
	require.NoError(t, storage.RemoveItem("foo"))
	value, err := storage.GetItem("foo")
	require.NoError(t, err)
	require.Equal(t, "", value)

	require.NoError(t, storage.SetItem("a", "1"))
	require.NoError(t, storage.Clear())
	length, err := page.Evaluate("() => localStorage.length")
	require.NoError(t, err)
	require.Equal(t, 0, length)
}

func TestPageSessionStorageSetAndGet(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)

	storage := page.SessionStorage()
	require.NoError(t, storage.SetItem("foo", "bar"))
	value, err := storage.GetItem("foo")
	require.NoError(t, err)
	require.Equal(t, "bar", value)

	jsValue, err := page.Evaluate("() => sessionStorage.getItem('foo')")
	require.NoError(t, err)
	require.Equal(t, "bar", jsValue)
}

func TestPageSessionStorageItemsRemoveAndClear(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)

	storage := page.SessionStorage()
	require.NoError(t, storage.SetItem("a", "1"))
	require.NoError(t, storage.SetItem("b", "2"))

	items, err := storage.Items()
	require.NoError(t, err)
	require.Contains(t, items, playwright.WebStorageItem{Name: "a", Value: "1"})
	require.Contains(t, items, playwright.WebStorageItem{Name: "b", Value: "2"})

	require.NoError(t, storage.RemoveItem("a"))
	value, err := storage.GetItem("a")
	require.NoError(t, err)
	require.Equal(t, "", value)

	require.NoError(t, storage.Clear())
	length, err := page.Evaluate("() => sessionStorage.length")
	require.NoError(t, err)
	require.Equal(t, 0, length)
}

// localStorage and sessionStorage are independent stores (mirrors upstream).
func TestPageLocalAndSessionStorageAreIndependent(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)

	require.NoError(t, page.LocalStorage().SetItem("key", "local"))
	require.NoError(t, page.SessionStorage().SetItem("key", "session"))

	localValue, err := page.LocalStorage().GetItem("key")
	require.NoError(t, err)
	require.Equal(t, "local", localValue)

	sessionValue, err := page.SessionStorage().GetItem("key")
	require.NoError(t, err)
	require.Equal(t, "session", sessionValue)
}
