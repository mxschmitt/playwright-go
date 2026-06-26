package playwright_test

import (
	"regexp"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestPageAssertionsToHaveTitle(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`<title>new title</title>`))

	require.NoError(t, expect.Page(page).ToHaveTitle("new title"))
	require.NoError(t, expect.Page(page).ToHaveTitle(regexp.MustCompile("(?i)new title")))
	require.NoError(t, expect.Page(page).Not().ToHaveTitle("not the current title", playwright.PageAssertionsToHaveTitleOptions{
		Timeout: playwright.Float(750),
	}))

	_, err = page.Evaluate(`setTimeout(() => {
		document.title = 'great title';
	}, 300);
	`)
	require.NoError(t, err)
	require.NoError(t, expect.Page(page).ToHaveTitle("great title"))
	require.NoError(t, expect.Page(page).Not().ToHaveTitle("not the current title"))
}

func TestPageAssertionsToHaveURL(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)

	require.NoError(t, expect.Page(page).ToHaveURL(server.EMPTY_PAGE))
	require.NoError(t, expect.Page(page).ToHaveURL(regexp.MustCompile(`.*/empty\.html`), playwright.PageAssertionsToHaveURLOptions{
		Timeout: playwright.Float(750),
	}))
	require.NoError(t, expect.Page(page).Not().ToHaveURL("https://playwright.dev"))

	_, err = page.Goto("data:text/html,<div>A</div>")
	require.NoError(t, err)
	require.NoError(t, expect.Page(page).ToHaveURL("DATA:teXT/HTml,<div>a</div>", playwright.PageAssertionsToHaveURLOptions{
		IgnoreCase: playwright.Bool(true),
	}))
}

func TestPageAssertionsToHaveURLWithBaseURL(t *testing.T) {
	BeforeEach(t)

	page, err := browser.NewPage(playwright.BrowserNewPageOptions{
		BaseURL: &server.PREFIX,
	})
	require.NoError(t, err)
	_, err = page.Goto("/empty.html")
	require.NoError(t, err)
	require.NoError(t, expect.Page(page).ToHaveURL("/empty.html"))
	require.NoError(t, expect.Page(page).ToHaveURL(regexp.MustCompile(`.*/empty\.html`)))
	require.NoError(t, expect.Page(page).Not().ToHaveURL("https://playwright.dev"))
	// An absolute URL must resolve to itself against the base URL (new URL
	// semantics), not be mangled by naive path joining.
	require.NoError(t, expect.Page(page).ToHaveURL(server.EMPTY_PAGE))
	require.NoError(t, page.Close())
}

func TestPageAssertionsToHaveAccessibleErrorMessage(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`
		<form>
			<input role="textbox" aria-invalid="true" aria-errormessage="error-message" />
			<div id="error-message">Hello</div>
			<div id="irrelevant-error">This should not be considered.</div>
		</form>
	`))

	locator := page.Locator("input[role=\"textbox\"]")
	require.NoError(t, expect.Locator(locator).ToHaveAccessibleErrorMessage("Hello"))
	require.NoError(t, expect.Locator(locator).Not().ToHaveAccessibleErrorMessage("hello"))
	require.NoError(t, expect.Locator(locator).ToHaveAccessibleErrorMessage("hello", playwright.LocatorAssertionsToHaveAccessibleErrorMessageOptions{
		IgnoreCase: playwright.Bool(true),
	}))
	require.NoError(t, expect.Locator(locator).ToHaveAccessibleErrorMessage(regexp.MustCompile(`ell\w`)))
	require.NoError(t, expect.Locator(locator).Not().ToHaveAccessibleErrorMessage(regexp.MustCompile(`hello`)))
	require.NoError(t, expect.Locator(locator).ToHaveAccessibleErrorMessage(regexp.MustCompile(`hello`), playwright.LocatorAssertionsToHaveAccessibleErrorMessageOptions{
		IgnoreCase: playwright.Bool(true),
	}))
	require.NoError(t, expect.Locator(locator).Not().ToHaveAccessibleErrorMessage("This should not be considered."))
}
