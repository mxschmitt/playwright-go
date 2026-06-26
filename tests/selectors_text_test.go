//nolint:staticcheck
package playwright_test

import (
	"regexp"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestHasTextAndInternalTextShouldMatchFullNodeInStrictMode(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`<div id=div1>hello<span>world</span></div>
		<div id=div2>hello</div>`))

	require.NoError(t, expect.Locator(page.GetByText("helloworld", playwright.PageGetByTextOptions{
		Exact: playwright.Bool(true),
	})).ToHaveId("div1"))
	require.NoError(t, expect.Locator(page.GetByText("hello", playwright.PageGetByTextOptions{
		Exact: playwright.Bool(true),
	})).ToHaveId("div2"))
	require.NoError(t, expect.Locator(page.Locator("div", playwright.PageLocatorOptions{
		HasText: regexp.MustCompile(`^helloworld$`),
	})).ToHaveId("div1"))
	require.NoError(t, expect.Locator(page.Locator("div", playwright.PageLocatorOptions{
		HasText: regexp.MustCompile(`^hello$`),
	})).ToHaveId("div2"))

	require.NoError(t, page.SetContent(`<div id=div1><span id=span1>hello</span>world</div>
		<div id=div2><span id=span2>hello</span></div>`))

	require.NoError(t, expect.Locator(page.GetByText("helloworld", playwright.PageGetByTextOptions{
		Exact: playwright.Bool(true),
	})).ToHaveId("div1"))

	utils.AssertResult(t, func() (interface{}, error) {
		return page.GetByText("hello", playwright.PageGetByTextOptions{
			Exact: playwright.Bool(true),
		}).EvaluateAll(`els => els.map(e => e.id)`)
	}, []interface{}{"span1", "span2"})

	require.NoError(t, expect.Locator(page.Locator("div", playwright.PageLocatorOptions{
		HasText: regexp.MustCompile(`^helloworld$`),
	})).ToHaveId("div1"))
	require.NoError(t, expect.Locator(page.Locator("div", playwright.PageLocatorOptions{
		HasText: regexp.MustCompile(`^hello$`),
	})).ToHaveId("div2"))
}

func TestSelectorsTextShouldWork(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent("<div>yo</div><div>ya</div><div>\nye  </div>"))

	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector("text=ya", "e => e.outerHTML", nil)
	}, "<div>ya</div>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector(`text="ya"`, "e => e.outerHTML", nil)
	}, "<div>ya</div>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector(`text=/^[ay]+$/`, "e => e.outerHTML", nil)
	}, "<div>ya</div>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector(`text=/Ya/i`, "e => e.outerHTML", nil)
	}, "<div>ya</div>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector("text=ye", "e => e.outerHTML", nil)
	}, "<div>\nye  </div>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.GetByText("ye").Evaluate("e => e.outerHTML", nil)
	}, "<div>\nye  </div>")

	require.NoError(t, page.SetContent("<div> ye </div><div>ye</div>"))

	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector(`text="ye"`, "e => e.outerHTML", nil)
	}, "<div> ye </div>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.GetByText("ye", playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).First().Evaluate("e => e.outerHTML", nil)
	}, "<div> ye </div>")

	require.NoError(t, page.SetContent(`<div>yo</div><div>"ya</div><div> hello world! </div>`))

	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector("text=\"\\\"ya\"", "e => e.outerHTML", nil)
	}, "<div>\"ya</div>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector(`text=/hello/`, "e => e.outerHTML", nil)
	}, "<div> hello world! </div>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector("text=/^\\s*heLLo/i", "e => e.outerHTML", nil)
	}, "<div> hello world! </div>")

	require.NoError(t, page.SetContent(`<div>yo<div>ya</div>hey<div>hey</div></div>`))

	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector("text=hey", "e => e.outerHTML", nil)
	}, "<div>hey</div>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector(`text=yo>>text="ya"`, "e => e.outerHTML", nil)
	}, "<div>ya</div>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector(`text=yo>> text="ya"`, "e => e.outerHTML", nil)
	}, "<div>ya</div>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector(`text=yo >>text='ya'`, "e => e.outerHTML", nil)
	}, "<div>ya</div>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector(`text=yo >> text='ya'`, "e => e.outerHTML", nil)
	}, "<div>ya</div>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector("'yo'>>\"ya\"", "e => e.outerHTML", nil)
	}, "<div>ya</div>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector("\"yo\" >> 'ya'", "e => e.outerHTML", nil)
	}, "<div>ya</div>")

	require.NoError(t, page.SetContent(`<div>yo<span id="s1"></span></div><div>yo<span id="s2"></span><span id="s3"></span></div>`))

	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelectorAll("text=yo", "es => es.map(e => e.outerHTML).join('\\n')", nil)
	}, "<div>yo<span id=\"s1\"></span></div>\n<div>yo<span id=\"s2\"></span><span id=\"s3\"></span></div>")

	require.NoError(t, page.SetContent("<div>'</div><div>\"</div><div>\\</div><div>x</div>"))

	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector("text='\\''", "e => e.outerHTML", nil)
	}, "<div>'</div>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector("text='\"'", "e => e.outerHTML", nil)
	}, "<div>\"</div>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector(`text="\""`, "e => e.outerHTML", nil)
	}, "<div>\"</div>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector(`text="'"`, "e => e.outerHTML", nil)
	}, "<div>'</div>")

	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector(`text="\x"`, "e => e.outerHTML", nil)
	}, "<div>x</div>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector("text='\\x'", "e => e.outerHTML", nil)
	}, "<div>x</div>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector("text='\\\\'", "e => e.outerHTML", nil)
	}, "<div>\\</div>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector(`text="\\"`, "e => e.outerHTML", nil)
	}, "<div>\\</div>")

	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector(`text="`, "e => e.outerHTML", nil)
	}, "<div>\"</div>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector("text='", "e => e.outerHTML", nil)
	}, "<div>'</div>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector(`"x"`, "e => e.outerHTML", nil)
	}, "<div>x</div>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector("'x'", "e => e.outerHTML", nil)
	}, "<div>x</div>")

	_, err := page.QuerySelectorAll(`"`)
	require.Error(t, err)
	_, err = page.QuerySelectorAll("'")
	require.Error(t, err)

	require.NoError(t, page.SetContent("<div> ' </div><div> \" </div>"))

	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector(`text="`, "e => e.outerHTML", nil)
	}, `<div> " </div>`)
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector("text='", "e => e.outerHTML", nil)
	}, `<div> ' </div>`)

	require.NoError(t, page.SetContent("<div>Hi'\"&gt;&gt;foo=bar</div>"))

	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector(`text="Hi'\">>foo=bar"`, "e => e.outerHTML", nil)
	}, "<div>Hi'\"&gt;&gt;foo=bar</div>")

	require.NoError(t, page.SetContent("<div>Hi&gt;&gt;<span></span></div>"))
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector(`text="Hi>>">>span`, "e => e.outerHTML", nil)
	}, "<span></span>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector("text=/Hi\\>\\>/ >> span", "e => e.outerHTML", nil)
	}, "<span></span>")

	require.NoError(t, page.SetContent("<div>a<br>b</div><div>a</div>"))
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector("text=a", "e => e.outerHTML", nil)
	}, "<div>a<br>b</div>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector("text=b", "e => e.outerHTML", nil)
	}, "<div>a<br>b</div>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector("text=ab", "e => e.outerHTML", nil)
	}, "<div>a<br>b</div>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.QuerySelector("text=abc")
	}, nil)
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelectorAll("text=a", "els => els.length")
	}, 2)
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelectorAll("text=b", "els => els.length")
	}, 1)
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelectorAll("text=ab", "els => els.length")
	}, 1)
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelectorAll("text=abc", "els => els.length")
	}, 0)

	require.NoError(t, page.SetContent("<div></div><span></span>"))
	_, err = page.EvalOnSelector("div", `div => {
		div.appendChild(document.createTextNode('hello'))
		div.appendChild(document.createTextNode('world'))
	}`, nil)
	require.NoError(t, err)
	_, err = page.EvalOnSelector("span", `span => {
		span.appendChild(document.createTextNode('hello'))
		span.appendChild(document.createTextNode('world'))
	}`, nil)
	require.NoError(t, err)
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector("text=lowo", "e => e.outerHTML", nil)
	}, "<div>helloworld</div>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelectorAll("text=lowo", "els => els.map(e => e.outerHTML).join('')")
	}, "<div>helloworld</div><span>helloworld</span>")

	require.NoError(t, page.SetContent("<span>Sign&nbsp;in</span><span>Hello\n \nworld</span>"))
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector("text=Sign in", "e => e.outerHTML", nil)
	}, "<span>Sign&nbsp;in</span>")
	ret, err := page.QuerySelectorAll("text=Sign \tin")
	require.NoError(t, err)
	require.Len(t, ret, 1)
	ret, err = page.QuerySelectorAll(`text="Sign in"`)
	require.NoError(t, err)
	require.Len(t, ret, 1)
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector("text=lo wo", "e => e.outerHTML", nil)
	}, "<span>Hello\n \nworld</span>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector(`text="Hello world"`, "e => e.outerHTML", nil)
	}, "<span>Hello\n \nworld</span>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.QuerySelector(`text="lo wo"`)
	}, nil)
	ret, err = page.QuerySelectorAll("text=lo \nwo")
	require.NoError(t, err)
	require.Len(t, ret, 1)
	ret, err = page.QuerySelectorAll("text=\"lo \nwo\"")
	require.NoError(t, err)
	require.Len(t, ret, 0)

	require.NoError(t, page.SetContent("<div>let's<span>hello</span></div>"))
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector("text=/let's/i >> span", "e => e.outerHTML", nil)
	}, "<span>hello</span>")
	utils.AssertResult(t, func() (interface{}, error) {
		return page.EvalOnSelector("text=/let\\'s/i >> span", "e => e.outerHTML", nil)
	}, "<span>hello</span>")
}
