//nolint:staticcheck
package playwright_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestSelectorsRegisterShouldWork(t *testing.T) {
	tagSelector := `(() => ({
		query(root, selector) {
			return root.querySelector(selector);
		},
		queryAll(root, selector) {
			return Array.from(root.querySelectorAll(selector));
		}
	}))()`
	// Use unique names to avoid conflicts when running tests multiple times
	uniqueSuffix := fmt.Sprintf("%s_%d", t.Name(), time.Now().UnixNano())
	selectorName := "tag_" + browserName + "_" + uniqueSuffix
	selector2Name := "tag2_" + browserName + "_" + uniqueSuffix

	// Register one engine before creating context.
	err := pw.Selectors.Register(selectorName, playwright.Script{
		Content: &tagSelector,
	})
	require.NoError(t, err)

	BeforeEach(t)

	// Register another engine after creating context.
	err = pw.Selectors.Register(selector2Name, playwright.Script{
		Content: &tagSelector,
	})
	require.NoError(t, err)

	require.NoError(t, page.SetContent(`<div><span></span></div><div></div>`))

	ret, err := page.EvalOnSelector(selectorName+"=DIV", `e => e.nodeName`, nil)
	require.NoError(t, err)
	require.Equal(t, "DIV", ret)
	ret, err = page.EvalOnSelector(selectorName+"=SPAN", `e => e.nodeName`, nil)
	require.NoError(t, err)
	require.Equal(t, "SPAN", ret)
	ret, err = page.EvalOnSelectorAll(selectorName+"=DIV", `es => es.length`)
	require.NoError(t, err)
	require.Equal(t, 2, ret)

	ret, err = page.EvalOnSelector(selector2Name+"=DIV", `e => e.nodeName`, nil)
	require.NoError(t, err)
	require.Equal(t, "DIV", ret)
	ret, err = page.EvalOnSelector(selector2Name+"=SPAN", `e => e.nodeName`, nil)
	require.NoError(t, err)
	require.Equal(t, "SPAN", ret)
	ret, err = page.EvalOnSelectorAll(selector2Name+"=DIV", `es => es.length`)
	require.NoError(t, err)
	require.Equal(t, 2, ret)

	// Selector names are case-sensitive.
	_, err = page.Locator("tAG=DIV").All()
	require.ErrorContains(t, err, `Unknown engine "tAG" while parsing selector tAG=DIV`)

	// Registering the same engine name again errors upfront (matches upstream).
	err = pw.Selectors.Register(selectorName, playwright.Script{Content: &tagSelector})
	require.ErrorContains(t, err, "has been already registered")

	require.NoError(t, context.Close())
}

func TestSelectorsShouldUseDataTestIdInStrictErrors(t *testing.T) {
	BeforeEach(t)

	pw.Selectors.SetTestIdAttribute("data-custom-id")
	require.NoError(t, page.SetContent(`
	<div>
		<div></div>
		<div>
			<div></div>
			<div></div>
		</div>
	</div>
	<div>
		<div class='foo bar:0' data-custom-id='One'>
		</div>
		<div class='foo bar:1' data-custom-id='Two'>
		</div>
	</div>`))

	err := page.Locator(".foo").Hover(playwright.LocatorHoverOptions{
		Timeout: playwright.Float(500),
	})
	require.ErrorContains(t, err, "strict mode violation")
	require.ErrorContains(t, err, `<div class="foo bar:0`)
	require.ErrorContains(t, err, `<div class="foo bar:1`)
	require.ErrorContains(t, err, `aka getByTestId('One')`)
}

func TestSelectorsShouldWorkWithPath(t *testing.T) {
	BeforeEach(t)

	// Use unique name to avoid conflicts when running tests multiple times
	selectorName := fmt.Sprintf("foo_%s_%d", t.Name(), time.Now().UnixNano())
	require.NoError(t, pw.Selectors.Register(selectorName, playwright.Script{
		Path: playwright.String(Asset("sectionselectorengine.js")),
	}))
	require.NoError(t, page.SetContent(`<section></section>`))

	ret, err := page.EvalOnSelector(selectorName+"=whatever", `e => e.nodeName`, nil)
	require.NoError(t, err)
	require.Equal(t, "SECTION", ret)
}
