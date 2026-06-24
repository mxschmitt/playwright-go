package playwright_test

import (
	"testing"

	"github.com/playwright-community/playwright-go"
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
