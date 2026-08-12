package playwright_test

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func Unshift(snapshot string) string {
	lines := strings.Split(snapshot, "\n")
	whitespacePrefixLength := 100
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		// replace each tab with 2 spaces
		match := regexp.MustCompile(`^(\t*)`).FindStringSubmatch(line)
		if len(match) > 1 {
			lines[i] = regexp.MustCompile(`^(\t*)`).ReplaceAllString(line, strings.Repeat("  ", len(match[1])))
			if len(match[1]) < whitespacePrefixLength {
				whitespacePrefixLength = len(match[1]) * 2
			}
		}
	}

	for i, line := range lines {
		if line == "" {
			continue
		}
		lines[i] = regexp.MustCompile(fmt.Sprintf(`^(\s{0,%d})`, whitespacePrefixLength)).ReplaceAllString(line, "")
	}
	return strings.Join(slices.DeleteFunc(lines, func(line string) bool { return strings.TrimSpace(line) == "" }), "\n")
}

func checkAndMatchSnapshot(t *testing.T, locator playwright.Locator, snapshot string) {
	t.Helper()
	snapshot = Unshift(snapshot)
	ariaSnapshot, err := locator.AriaSnapshot()
	require.NoError(t, err)
	require.Equal(t, snapshot, ariaSnapshot)
	require.NoError(t, expect.Locator(locator).ToMatchAriaSnapshot(snapshot))
}

func TestShouldSnapshot(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`<h1>title</h1>`))
	checkAndMatchSnapshot(t, page.Locator("body"), `
	- heading "title" [level=1]
	`)
}

func TestShouldSnapshotList(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`<h1>title</h1><h1>title 2</h1>`))
	checkAndMatchSnapshot(t, page.Locator("body"), `
	- heading "title" [level=1]
	- heading "title 2" [level=1]
	`)
}

func TestShouldSnapshotListWithList(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`<ul><li>one</li><li>two</li></ul>`))
	checkAndMatchSnapshot(t, page.Locator("body"), `
	- list:
		- listitem: one
		- listitem: two
	`)
}

func TestShouldSnapshotListWithAccessibleName(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`<ul aria-label="my list"><li>one</li><li>two</li></ul>`))
	checkAndMatchSnapshot(t, page.Locator("body"), `
	- list "my list":
		- listitem: one
		- listitem: two
	`)
}

func TestShouldSnapshotComplex(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`<ul><li><a href="about:blank">link</a></li></ul>`))
	checkAndMatchSnapshot(t, page.Locator("body"), `
	- list:
		- listitem:
			- link "link":
				- /url: about:blank
	`)
}

func TestShouldSnapshotWithRef(t *testing.T) {
	t.Skip("the Ref option was removed in Playwright v1.53")
}

func TestShouldSnapshotWithUnexpectedChildrenEqual(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`
		<ul>
			<li>One</li>
			<li>Two</li>
			<li>Three</li>
		</ul>
	`))
	require.NoError(t, expect.Locator(page.Locator("body")).ToMatchAriaSnapshot(Unshift(`
	- list:
		- listitem: One
		- listitem: Three
	`)))
	require.Error(t, expect.Locator(page.Locator("body")).ToMatchAriaSnapshot(Unshift(`
	- list:
		- /children: equal
		- listitem: One
		- listitem: Three
	`), playwright.LocatorAssertionsToMatchAriaSnapshotOptions{Timeout: playwright.Float(1000)}))
}

func TestShouldSnapshotWithUnexpectedChildrenDeepEqual(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`
      <ul>
        <li>
          <ul>
            <li>1.1</li>
            <li>1.2</li>
          </ul>
        </li>
      </ul>
	`))
	require.NoError(t, expect.Locator(page.Locator("body")).ToMatchAriaSnapshot(`
      - list:
        - listitem:
          - list:
            - listitem: 1.1
	`))
	require.NoError(t, expect.Locator(page.Locator("body")).ToMatchAriaSnapshot(`
        - list:
          - /children: equal
          - listitem:
            - list:
              - listitem: 1.1
	`))
	require.Error(t, expect.Locator(page.Locator("body")).ToMatchAriaSnapshot(`
          - list:
            - /children: deep-equal
            - listitem:
              - list:
                - listitem: 1.1
	`, playwright.LocatorAssertionsToMatchAriaSnapshotOptions{Timeout: playwright.Float(1000)}))
}

func TestPageAssertionsToMatchAriaSnapshotInvalidAttribute(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`
		<input type="text" aria-label="Email" aria-invalid="true" value="not-an-email">
		<input type="text" aria-label="Name" value="Alice">
	`))
	require.NoError(t, expect.Page(page).ToMatchAriaSnapshot(Unshift(`
		- textbox "Email" [invalid]: not-an-email
		- textbox "Name": Alice
	`)))
	require.NoError(t, expect.Page(page).ToMatchAriaSnapshot(Unshift(`
		- textbox "Email" [invalid=true]: not-an-email
		- textbox "Name" [invalid=false]: Alice
	`)))

	// `grammar` and `spelling` retain their semantic value instead of being
	// collapsed into the generic invalid state.
	require.NoError(t, page.SetContent(`
		<input type="text" aria-label="Bio" aria-invalid="grammar">
		<input type="text" aria-label="Note" aria-invalid="spelling">
	`))
	require.NoError(t, expect.Page(page).ToMatchAriaSnapshot(Unshift(`
		- textbox "Bio" [invalid=grammar]
		- textbox "Note" [invalid=spelling]
	`)))
	err := expect.Page(page).ToMatchAriaSnapshot(Unshift(`
		- textbox "Bio" [invalid]
	`), playwright.PageAssertionsToMatchAriaSnapshotOptions{Timeout: playwright.Float(1000)})
	require.ErrorContains(t, err, "[invalid=grammar]")

	// Any non-false value other than grammar/spelling is the generic true state.
	require.NoError(t, page.SetContent(`<input type="text" aria-label="Zip" aria-invalid="garbage">`))
	require.NoError(t, expect.Page(page).ToMatchAriaSnapshot(`- textbox "Zip" [invalid]`))
}

// Covers PageAssertions.ToMatchAriaSnapshot, which must use the
// "to.match.aria" expect expression (mirrors the Locator variant).
func TestPageAssertionsToMatchAriaSnapshot(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`<h1>title</h1>`))
	require.NoError(t, expect.Page(page).ToMatchAriaSnapshot(`- heading "title" [level=1]`))
	require.Error(t, expect.Page(page).ToMatchAriaSnapshot(
		`- heading "wrong"`,
		playwright.PageAssertionsToMatchAriaSnapshotOptions{Timeout: playwright.Float(1000)},
	))
}
