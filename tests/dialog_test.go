package playwright_test

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestDialog(t *testing.T) {
	BeforeEach(t)

	page.OnDialog(func(dialog playwright.Dialog) {
		require.Equal(t, "alert", dialog.Type())
		require.Equal(t, "", dialog.DefaultValue())
		require.Equal(t, "yo", dialog.Message())
		require.NoError(t, dialog.Accept())
	})
	_, err := page.Evaluate("alert('yo')")
	require.NoError(t, err)
}

func TestDialogDismiss(t *testing.T) {
	BeforeEach(t)

	page.OnDialog(func(dialog playwright.Dialog) {
		require.NoError(t, dialog.Dismiss())
	})
	result, err := page.Evaluate("prompt('question?')")
	require.NoError(t, err)
	require.Equal(t, result, nil)
}

func TestDialogAcceptWithText(t *testing.T) {
	BeforeEach(t)

	page.OnDialog(func(dialog playwright.Dialog) {
		require.NoError(t, dialog.Accept("hey foobar"))
	})
	result, err := page.Evaluate("prompt('question?')")
	require.NoError(t, err)
	require.Equal(t, result, "hey foobar")
}

func TestDialogShouldWorkInPopup(t *testing.T) {
	skipWebKitMacOSPopup(t)
	BeforeEach(t)

	var d playwright.Dialog
	context.OnDialog(func(dialog playwright.Dialog) {
		d = dialog
		require.NoError(t, dialog.Accept("hello"))
	})

	popup, err := page.ExpectPopup(func() error {
		ret, err := page.Evaluate("() => window.open('').prompt('hey?')")
		require.NoError(t, err)
		require.Equal(t, "hello", ret)
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, "hey?", d.Message())
	require.Equal(t, d.Page(), popup)
}

func TestDialogClosedEventShouldWork(t *testing.T) {
	BeforeEach(t)

	page.OnDialog(func(dialog playwright.Dialog) {
		require.NoError(t, dialog.Accept("hello"))
	})
	var result any
	var pageEvent any
	contextEvent, err := context.ExpectEvent("dialogclosed", func() error {
		var err error
		pageEvent, err = page.ExpectEvent("dialogclosed", func() error {
			var err error
			result, err = page.Evaluate("() => prompt('hey?')")
			return err
		})
		return err
	})
	require.NoError(t, err)
	require.Equal(t, "hello", result)
	dialog := pageEvent.(playwright.Dialog)
	require.Equal(t, dialog, contextEvent)
	require.Equal(t, "hey?", dialog.Message())
	require.Equal(t, page, dialog.Page())
}

func TestDialogClosedEventShouldFireWhenDialogIsDismissed(t *testing.T) {
	BeforeEach(t)

	page.OnDialog(func(dialog playwright.Dialog) {
		require.NoError(t, dialog.Dismiss())
	})
	closed, err := page.ExpectEvent("dialogclosed", func() error {
		_, err := page.Evaluate("() => confirm('sure?')")
		return err
	})
	require.NoError(t, err)
	require.Equal(t, "sure?", closed.(playwright.Dialog).Message())
}

func TestDialogClosedEventShouldFireForAutoDismissedDialogs(t *testing.T) {
	BeforeEach(t)

	closed, err := page.ExpectEvent("dialogclosed", func() error {
		_, err := page.Evaluate("() => alert('yo')")
		return err
	})
	require.NoError(t, err)
	require.Equal(t, "yo", closed.(playwright.Dialog).Message())
}

func TestDialogClosedEventShouldBeDeliveredToOnDialogClosedHandlers(t *testing.T) {
	BeforeEach(t)

	pageDialogs := make(chan playwright.Dialog, 1)
	contextDialogs := make(chan playwright.Dialog, 1)
	page.OnDialogClosed(func(dialog playwright.Dialog) { pageDialogs <- dialog })
	context.OnDialogClosed(func(dialog playwright.Dialog) { contextDialogs <- dialog })
	_, err := page.Evaluate("() => alert('yo')")
	require.NoError(t, err)
	require.Equal(t, "yo", (<-pageDialogs).Message())
	require.Equal(t, "yo", (<-contextDialogs).Message())
}
