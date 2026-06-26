//nolint:staticcheck
package playwright_test

import (
	"regexp"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestSelectorsGetByEscaping(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`
	<label id=label for=control>Hello my
wo"rld</label><input id=control />`))
	_, err := page.EvalOnSelector("input", `input => {
		input.setAttribute('placeholder', 'hello my\nwo"rld');
		input.setAttribute('title', 'hello my\nwo"rld');
		input.setAttribute('alt', 'hello my\nwo"rld');
	}`, nil)
	require.NoError(t, err)
	require.NoError(t, expect.Locator(page.GetByText("hello my\nwo\"rld")).ToHaveAttribute("id", "label"))
	require.NoError(t, expect.Locator(page.GetByText("hello       my     wo\"rld")).ToHaveAttribute("id", "label"))
	require.NoError(t, expect.Locator(page.GetByLabel("hello my\nwo\"rld")).ToHaveAttribute("id", "control"))
	require.NoError(t, expect.Locator(page.GetByPlaceholder("hello my\nwo\"rld")).ToHaveAttribute("id", "control"))
	require.NoError(t, expect.Locator(page.GetByAltText("hello my\nwo\"rld")).ToHaveAttribute("id", "control"))
	require.NoError(t, expect.Locator(page.GetByTitle("hello my\nwo\"rld")).ToHaveAttribute("id", "control"))

	require.NoError(t, page.SetContent(`
	<label id=label for=control>Hello my
world</label><input id=control />`))
	_, err = page.EvalOnSelector("input", `input => {
		input.setAttribute('placeholder', 'hello my\nworld');
		input.setAttribute('title', 'hello my\nworld');
		input.setAttribute('alt', 'hello my\nworld');
	}`, nil)
	require.NoError(t, err)
	require.NoError(t, expect.Locator(page.GetByText("hello my\nworld")).ToHaveAttribute("id", "label"))
	require.NoError(t, expect.Locator(page.GetByText("hello       my     world")).ToHaveAttribute("id", "label"))
	require.NoError(t, expect.Locator(page.GetByLabel("hello my\nworld")).ToHaveAttribute("id", "control"))
	require.NoError(t, expect.Locator(page.GetByPlaceholder("hello my\nworld")).ToHaveAttribute("id", "control"))
	require.NoError(t, expect.Locator(page.GetByAltText("hello my\nworld")).ToHaveAttribute("id", "control"))
	require.NoError(t, expect.Locator(page.GetByTitle("hello my\nworld")).ToHaveAttribute("id", "control"))

	require.NoError(t, page.SetContent(`<div id=target title="my title">Text here</div>`))
	require.NoError(t, expect.Locator(page.GetByTitle("my title", playwright.PageGetByTitleOptions{
		Exact: playwright.Bool(true),
	})).ToHaveCount(1, playwright.LocatorAssertionsToHaveCountOptions{
		Timeout: playwright.Float(500),
	}))
	require.NoError(t, expect.Locator(page.GetByTitle("my t\\itle", playwright.PageGetByTitleOptions{
		Exact: playwright.Bool(true),
	})).ToHaveCount(0, playwright.LocatorAssertionsToHaveCountOptions{
		Timeout: playwright.Float(500),
	}))
	require.NoError(t, expect.Locator(page.GetByTitle("my t\\\\itle", playwright.PageGetByTitleOptions{
		Exact: playwright.Bool(true),
	})).ToHaveCount(0, playwright.LocatorAssertionsToHaveCountOptions{
		Timeout: playwright.Float(500),
	}))

	require.NoError(t, page.SetContent(`<label for=target>foo &gt;&gt; bar</label><input id=target>`))
	_, err = page.EvalOnSelector("input", `input => {
		input.setAttribute('placeholder', 'foo >> bar');
		input.setAttribute('title', 'foo >> bar');
		input.setAttribute('alt', 'foo >> bar');
	}`, nil)
	require.NoError(t, err)

	utils.AssertResult(t, func() (interface{}, error) {
		return page.GetByText("foo >> bar").TextContent()
	}, "foo >> bar")
	require.NoError(t, expect.Locator(page.Locator("label")).ToHaveText("foo >> bar"))
	require.NoError(t, expect.Locator(page.GetByText("foo >> bar")).ToHaveText("foo >> bar"))
	utils.AssertResult(t, func() (interface{}, error) {
		return page.GetByText(regexp.MustCompile("foo >> bar")).TextContent()
	}, "foo >> bar")
	require.NoError(t, expect.Locator(page.GetByLabel("foo >> bar")).ToHaveAttribute("id", "target"))
	require.NoError(t, expect.Locator(page.GetByLabel(regexp.MustCompile("foo >> bar"))).ToHaveAttribute("id", "target"))
	require.NoError(t, expect.Locator(page.GetByPlaceholder("foo >> bar")).ToHaveAttribute("id", "target"))

	require.NoError(t, expect.Locator(page.GetByAltText("foo >> bar")).ToHaveAttribute("id", "target"))
	require.NoError(t, expect.Locator(page.GetByTitle("foo >> bar")).ToHaveAttribute("id", "target"))
	require.NoError(t, expect.Locator(page.GetByPlaceholder(regexp.MustCompile("foo >> bar"))).ToHaveAttribute("id", "target"))
	require.NoError(t, expect.Locator(page.GetByAltText(regexp.MustCompile("foo >> bar"))).ToHaveAttribute("id", "target"))
	require.NoError(t, expect.Locator(page.GetByTitle(regexp.MustCompile("foo >> bar"))).ToHaveAttribute("id", "target"))
}

func TestSelectorsGetByRoleEscaping(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`
		<a href="https://playwright.dev">issues 123</a>
		<a href="https://playwright.dev">he llo 56</a>
		<button>Click me</button>
	`))
	utils.AssertResult(t, func() (interface{}, error) {
		return page.GetByRole("button").EvaluateAll("els => els.map(e => e.outerHTML)")
	}, []interface{}{"<button>Click me</button>"})
	utils.AssertResult(t, func() (interface{}, error) {
		return page.GetByRole("link").EvaluateAll("els => els.map(e => e.outerHTML)")
	}, []interface{}{
		`<a href="https://playwright.dev">issues 123</a>`,
		`<a href="https://playwright.dev">he llo 56</a>`,
	})
	utils.AssertResult(t, func() (interface{}, error) {
		return page.GetByRole("link", playwright.PageGetByRoleOptions{
			Name: "issues 123",
		}).EvaluateAll("els => els.map(e => e.outerHTML)")
	}, []interface{}{`<a href="https://playwright.dev">issues 123</a>`})
	utils.AssertResult(t, func() (interface{}, error) {
		return page.GetByRole("link", playwright.PageGetByRoleOptions{
			Name: "sues",
		}).EvaluateAll("els => els.map(e => e.outerHTML)")
	}, []interface{}{`<a href="https://playwright.dev">issues 123</a>`})
	utils.AssertResult(t, func() (interface{}, error) {
		return page.GetByRole("link", playwright.PageGetByRoleOptions{
			Name: "  he    \n  llo ",
		}).EvaluateAll("els => els.map(e => e.outerHTML)")
	}, []interface{}{`<a href="https://playwright.dev">he llo 56</a>`})
	utils.AssertResult(t, func() (interface{}, error) {
		return page.GetByRole("button", playwright.PageGetByRoleOptions{
			Name: "issues",
		}).EvaluateAll("els => els.map(e => e.outerHTML)")
	}, []interface{}{})
	utils.AssertResult(t, func() (interface{}, error) {
		return page.GetByRole("link", playwright.PageGetByRoleOptions{
			Name:  "sues",
			Exact: playwright.Bool(true),
		}).EvaluateAll("els => els.map(e => e.outerHTML)")
	}, []interface{}{})
	utils.AssertResult(t, func() (interface{}, error) {
		return page.GetByRole("link", playwright.PageGetByRoleOptions{
			Name:  "   he \n llo 56 ",
			Exact: playwright.Bool(true),
		}).EvaluateAll("els => els.map(e => e.outerHTML)")
	}, []interface{}{`<a href="https://playwright.dev">he llo 56</a>`})
	utils.AssertResult(t, func() (interface{}, error) {
		return page.GetByRole("button", playwright.PageGetByRoleOptions{
			Name:  "Click me",
			Exact: playwright.Bool(true),
		}).EvaluateAll("els => els.map(e => e.outerHTML)")
	}, []interface{}{`<button>Click me</button>`})
	utils.AssertResult(t, func() (interface{}, error) {
		return page.GetByRole("button", playwright.PageGetByRoleOptions{
			Name:  "Click \\me",
			Exact: playwright.Bool(true),
		}).EvaluateAll("els => els.map(e => e.outerHTML)")
	}, []interface{}{})
	utils.AssertResult(t, func() (interface{}, error) {
		return page.GetByRole("button", playwright.PageGetByRoleOptions{
			Name:  "Click \\\\me",
			Exact: playwright.Bool(true),
		}).EvaluateAll("els => els.map(e => e.outerHTML)")
	}, []interface{}{})
}

func TestSelectorsGetByRoleWithDescription(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`
		<div role="alert" aria-label="Upload successful" aria-description="File doc-2025.pdf was uploaded successfully">Alert 1</div>
		<div role="alert" aria-label="Upload successful" aria-description="File report-2026.pdf was uploaded successfully">Alert 2</div>
		<div role="alert" aria-label="Invalid file" aria-description="File demo.doc has an invalid file format">Alert 3</div>
	`))

	// description substring match (default)
	utils.AssertResult(t, func() (interface{}, error) {
		return page.GetByRole("alert", playwright.PageGetByRoleOptions{
			Description: "doc-2025",
		}).EvaluateAll(`els => els.map(e => e.textContent)`)
	}, []interface{}{"Alert 1"})
	utils.AssertResult(t, func() (interface{}, error) {
		return page.GetByRole("alert", playwright.PageGetByRoleOptions{
			Description: "report-2026",
		}).EvaluateAll(`els => els.map(e => e.textContent)`)
	}, []interface{}{"Alert 2"})

	// description combined with name
	utils.AssertResult(t, func() (interface{}, error) {
		return page.GetByRole("alert", playwright.PageGetByRoleOptions{
			Name:        "Upload successful",
			Description: "doc-2025",
		}).EvaluateAll(`els => els.map(e => e.textContent)`)
	}, []interface{}{"Alert 1"})
	utils.AssertResult(t, func() (interface{}, error) {
		return page.GetByRole("alert", playwright.PageGetByRoleOptions{
			Name:        "Invalid file",
			Description: "doc-2025",
		}).EvaluateAll(`els => els.map(e => e.textContent)`)
	}, []interface{}{})

	// description exact match
	utils.AssertResult(t, func() (interface{}, error) {
		return page.GetByRole("alert", playwright.PageGetByRoleOptions{
			Description: "doc-2025",
			Exact:       playwright.Bool(true),
		}).EvaluateAll(`els => els.map(e => e.textContent)`)
	}, []interface{}{})
}

func TestSelectorsIncludeHiddenShouldWork(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`<button style="display: none">Hidden</button>`))
	utils.AssertResult(t, func() (interface{}, error) {
		return page.GetByRole("button", playwright.PageGetByRoleOptions{
			Name: "Hidden",
		}).EvaluateAll(`els => els.map(e => e.outerHTML)`)
	}, []interface{}{})
	utils.AssertResult(t, func() (interface{}, error) {
		return page.GetByRole("button", playwright.PageGetByRoleOptions{
			Name:          "Hidden",
			IncludeHidden: playwright.Bool(true),
		}).EvaluateAll(`els => els.map(e => e.outerHTML)`)
	}, []interface{}{`<button style="display: none">Hidden</button>`})
}
