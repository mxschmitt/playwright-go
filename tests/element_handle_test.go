//nolint:staticcheck
package playwright_test

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestElementHandleInnerText(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.PREFIX + "/dom.html")
	require.NoError(t, err)
	handle, err := page.QuerySelector("#inner")
	require.NoError(t, err)
	t1, err := handle.InnerText()
	require.NoError(t, err)
	require.Equal(t, t1, "Text, more text")
	t2, err := page.InnerText("#inner")
	require.NoError(t, err)
	require.Equal(t, t2, "Text, more text")
}

func TestElementHandleOwnerFrame(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	_, err = utils.AttachFrame(page, "iframe1", server.EMPTY_PAGE)
	require.NoError(t, err)
	frame := page.Frames()[1]
	elementHandle, err := frame.EvaluateHandle("document.body")
	require.NoError(t, err)
	ownerFrame, err := elementHandle.(playwright.ElementHandle).OwnerFrame()
	require.NoError(t, err)
	require.Equal(t, ownerFrame, frame)
	require.Equal(t, "iframe1", ownerFrame.Name())
}

func TestElementHandleContentFrame(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	_, err = utils.AttachFrame(page, "frame1", server.EMPTY_PAGE)
	require.NoError(t, err)
	elementHandle, err := page.QuerySelector("#frame1")
	require.NoError(t, err)
	frame, err := elementHandle.ContentFrame()
	require.NoError(t, err)
	require.Equal(t, frame, page.Frames()[1])
}

func TestElementHandleGetAttribute(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.PREFIX + "/dom.html")
	require.NoError(t, err)
	handle, err := page.QuerySelector("#outer")
	require.NoError(t, err)
	a1, err := handle.GetAttribute("name")
	require.NoError(t, err)
	require.Equal(t, "value", a1)
	a2, err := page.GetAttribute("#outer", "name")
	require.NoError(t, err)
	require.Equal(t, "value", a2)
}

func TestElementHandleDispatchEvent(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.PREFIX + "/input/button.html")
	require.NoError(t, err)
	element, err := page.QuerySelector("button")
	require.NoError(t, err)
	require.NoError(t, element.DispatchEvent("click"))
	result, err := page.Evaluate("() => result")
	require.NoError(t, err)
	require.Equal(t, "Clicked", result)
}

func TestElementHandleDispatchEventInitObject(t *testing.T) {
	BeforeEach(t)

	err := page.SetContent(`
	<button onclick="window.eventBubbles = event.bubbles">ok</button>`)
	require.NoError(t, err)
	element, err := page.QuerySelector("button")
	require.NoError(t, err)
	require.NoError(t, element.DispatchEvent("click", map[string]interface{}{
		"bubbles": true,
	}))
	result, err := page.Evaluate("() => window.eventBubbles")
	require.NoError(t, err)
	require.Equal(t, true, result)
}

func TestElementHandleHover(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.PREFIX + "/input/scrollable.html")
	require.NoError(t, err)
	btn, err := page.QuerySelector("#button-6")
	require.NoError(t, err)
	require.NoError(t, btn.Hover())
	result, err := page.Evaluate(`document.querySelector("button:hover").id`)
	require.NoError(t, err)
	require.Equal(t, "button-6", result)
}

func TestElementHandleClick(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.PREFIX + "/input/button.html")
	require.NoError(t, err)
	btn, err := page.QuerySelector("button")
	require.NoError(t, err)
	require.NoError(t, btn.Click())
	result, err := page.Evaluate(`result`)
	require.NoError(t, err)
	require.Equal(t, "Clicked", result)
}

func TestElementHandleDblclick(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.PREFIX + "/input/button.html")
	require.NoError(t, err)
	_, err = page.Evaluate(`() => {
            window.double = false;
            button = document.querySelector('button');
            button.addEventListener('dblclick', event => {
            window.double = true;
            });
	}`)
	require.NoError(t, err)
	btn, err := page.QuerySelector("button")
	require.NoError(t, err)
	require.NoError(t, btn.Dblclick())
	result, err := page.Evaluate("double")
	require.NoError(t, err)
	require.Equal(t, true, result)

	result, err = page.Evaluate(`result`)
	require.NoError(t, err)
	require.Equal(t, "Clicked", result)
}

func TestElementBoundingBox(t *testing.T) {
	BeforeEach(t)

	testCases := []struct {
		name        string
		setup       func() error
		selector    string
		expectedBox *playwright.Rect
	}{
		{
			name: "Test element bounding box",
			setup: func() error {
				require.NoError(t, page.SetViewportSize(500, 500))
				_, err := page.Goto(server.PREFIX + "/grid.html")
				return err
			},
			selector:    ".box:nth-of-type(13)",
			expectedBox: &playwright.Rect{X: 100.0, Y: 50.0, Width: 50.0, Height: 50.0},
		},
		{
			name: "Bounding box of display:none element should be nil",
			setup: func() error {
				_, err := page.Goto(server.EMPTY_PAGE)
				require.NoError(t, err)
				return page.SetContent("<div id='hidden' style='display:none'>Hidden</div>")
			},
			selector:    "#hidden",
			expectedBox: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, tc.setup())

			elementHandle, err := page.QuerySelector(tc.selector)
			require.NoError(t, err)
			box, err := elementHandle.BoundingBox()
			require.NoError(t, err)

			require.Equal(t, tc.expectedBox, box)
		})
	}
}

func TestElementHandleTap(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent("<input id='checkbox' type='checkbox'></input>"))
	value, err := page.EvalOnSelector("input", "el => el.checked", nil)
	require.NoError(t, err)
	require.Equal(t, false, value)

	elemHandle, err := page.QuerySelector("#checkbox")
	require.NoError(t, err)
	require.NoError(t, elemHandle.Tap())
	value, err = page.EvalOnSelector("input", "el => el.checked", nil)
	require.NoError(t, err)
	require.Equal(t, true, value)
}

func TestElementHandleQuerySelectorNotExists(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`
	<div id="a1">
	</div>
	`))
	rootElement, err := page.QuerySelector("#a1")
	require.NoError(t, err)
	element, err := rootElement.QuerySelector(".foobar")
	require.NoError(t, err)
	require.Nil(t, element)
}

func TestElementHandleQuerySelectorAll(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`
	<div id="a1">
		<div class="foobar">
		</div>
		<div class="foobar">
		</div>
	</div>
	`))
	rootElement, err := page.QuerySelector("#a1")
	require.NoError(t, err)
	elements, err := rootElement.QuerySelectorAll(".foobar")
	require.NoError(t, err)
	require.Equal(t, 2, len(elements))
	className, err := elements[0].GetAttribute("class")
	require.NoError(t, err)
	require.Equal(t, "foobar", className)
}

func TestElementHandleEvalOnSelector(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`
	<div id="a1">
		<div id="a2">
			foobar
		</div>
	</div>
	`))
	rootElement, err := page.QuerySelector("#a1")
	require.NoError(t, err)
	innerText, err := rootElement.EvalOnSelector("#a2", "e => e.innerText")
	require.NoError(t, err)
	require.Equal(t, "foobar", innerText)
}

func TestElementHandleEvalOnSelectorAll(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`
	<div id="a1">
		<div class="foobar">
		</div>
		<div class="foobar">
		</div>
	</div>
	`))
	rootElement, err := page.QuerySelector("#a1")
	require.NoError(t, err)
	classNames, err := rootElement.EvalOnSelectorAll(".foobar", "elements => [...elements].map(e => e.getAttribute('class'))")
	require.NoError(t, err)
	require.Equal(t, []interface{}([]interface{}{"foobar", "foobar"}), classNames)
}

func TestElementHandleString(t *testing.T) {
	BeforeEach(t)

	numberHandle, err := page.EvaluateHandle("() => 2")
	require.NoError(t, err)
	require.Equal(t, "2", numberHandle.String())
	stringHandle, err := page.EvaluateHandle("() => 'a'")
	require.NoError(t, err)
	require.Equal(t, "a", stringHandle.String())

	// A real ElementHandle renders its server-provided preview (not an empty
	// string) — newElementHandle must initialize preview like newJSHandle.
	require.NoError(t, page.SetContent(`<button>Submit</button>`))
	//nolint:staticcheck
	elementHandle, err := page.QuerySelector("button")
	require.NoError(t, err)
	require.NotEmpty(t, elementHandle.String())
}

func TestElementHandleCheck(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`
		<input type="checkbox"/>
	`))
	inputElement, err := page.QuerySelector("input")
	require.NoError(t, err)
	isChecked, err := inputElement.Evaluate("e => e.checked")
	require.NoError(t, err)
	require.Equal(t, false, isChecked)
	require.NoError(t, inputElement.Check())
	isChecked, err = inputElement.Evaluate("e => e.checked")
	require.NoError(t, err)
	require.Equal(t, true, isChecked)
}

func TestElementHandleUnCheck(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`
		<input type="checkbox" checked/>
	`))
	inputElement, err := page.QuerySelector("input")
	require.NoError(t, err)
	require.NoError(t, inputElement.Uncheck())
	isChecked, err := inputElement.Evaluate("e => e.checked")
	require.NoError(t, err)
	require.Equal(t, false, isChecked)
}

func TestElementHandleSelectOption(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent("<select id='lang'><option value='go'>go</option><option value='python'>python</option></select>"))
	elemHandle, err := page.QuerySelector("#lang")
	require.NoError(t, err)
	selected, err := elemHandle.SelectOption(playwright.SelectOptionValues{
		Values: playwright.StringSlice("go"),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(selected))
	require.Equal(t, "go", selected[0])
}

func TestElementHandleSelectOptionOverElementHandle(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent("<select id='lang'><option value='go'>go</option><option value='python'>python</option></select>"))

	pythonOption, err := page.QuerySelector("option[value=python]")
	require.NoError(t, err)

	selected, err := page.SelectOption("#lang", playwright.SelectOptionValues{
		Elements: &[]playwright.ElementHandle{pythonOption},
	})
	require.NoError(t, err)
	selected2, err := page.Locator("#lang").SelectOption(
		playwright.SelectOptionValues{Values: playwright.StringSlice("python")},
	)
	require.NoError(t, err)
	require.Equal(t, 1, len(selected))
	require.Equal(t, "python", selected[0])
	require.Equal(t, selected2, selected)
}

func TestElementHandleIsVisibleAndIsHiddenShouldWork(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`<div>Hi</div><span></span>`))
	div, err := page.QuerySelector("div")
	require.NoError(t, err)
	isVisible, err := div.IsVisible()
	require.NoError(t, err)
	require.True(t, isVisible)
	isHidden, err := div.IsHidden()
	require.NoError(t, err)
	require.False(t, isHidden)

	isVisible, err = page.IsVisible("div")
	require.NoError(t, err)
	require.True(t, isVisible)
	isHidden, err = page.IsHidden("div")
	require.NoError(t, err)
	require.False(t, isHidden)

	span, err := page.QuerySelector("span")
	require.NoError(t, err)
	isVisible, err = span.IsVisible()
	require.NoError(t, err)
	require.False(t, isVisible)
	isHidden, err = span.IsHidden()
	require.NoError(t, err)
	require.True(t, isHidden)

	isVisible, err = page.Locator("span").IsVisible()
	require.NoError(t, err)
	require.False(t, isVisible)
	isHidden, err = page.Locator("span").IsHidden()
	require.NoError(t, err)
	require.True(t, isHidden)
}

func TestElementHandleIsEnabledAndIsDisabledshouldWork(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`
		<button disabled>button1</button>
		<button>button2</button>
		<div>div</div>
	`))
	div, err := page.QuerySelector("div")
	require.NoError(t, err)
	isEnabled, err := div.IsEnabled()
	require.NoError(t, err)
	require.True(t, isEnabled)
	isDisabled, err := div.IsDisabled()
	require.NoError(t, err)
	require.False(t, isDisabled)

	isEnabled, err = page.IsEnabled("div")
	require.NoError(t, err)
	require.True(t, isEnabled)
	isDisabled, err = page.IsDisabled("div")
	require.NoError(t, err)
	require.False(t, isDisabled)

	button1, err := page.QuerySelector(":text('button1')")
	require.NoError(t, err)
	isEnabled, err = button1.IsEnabled()
	require.NoError(t, err)
	require.False(t, isEnabled)
	isDisabled, err = button1.IsDisabled()
	require.NoError(t, err)
	require.True(t, isDisabled)

	isEnabled, err = page.Locator(":text('button1')").IsEnabled()
	require.NoError(t, err)
	require.False(t, isEnabled)
	isDisabled, err = page.Locator(":text('button1')").IsDisabled()
	require.NoError(t, err)
	require.True(t, isDisabled)

	button2, err := page.QuerySelector(":text('button2')")
	require.NoError(t, err)
	isEnabled, err = button2.IsEnabled()
	require.NoError(t, err)
	require.True(t, isEnabled)
	isDisabled, err = button2.IsDisabled()
	require.NoError(t, err)
	require.False(t, isDisabled)

	isEnabled, err = page.Locator(":text('button2')").IsEnabled()
	require.NoError(t, err)
	require.True(t, isEnabled)
	isDisabled, err = page.Locator(":text('button2')").IsDisabled()
	require.NoError(t, err)
	require.False(t, isDisabled)
}

func TestElementHandleIsEditableShouldWork(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`
		<input id=input1 disabled><textarea></textarea><input id=input2>
	`))
	_, err := page.Locator("textarea").Evaluate("t => t.readOnly = true", nil)
	require.NoError(t, err)
	input1, err := page.QuerySelector("#input1")
	require.NoError(t, err)
	isEditable, err := input1.IsEditable()
	require.NoError(t, err)
	require.False(t, isEditable)
	isEditable, err = page.IsEditable("#input1")
	require.NoError(t, err)
	require.False(t, isEditable)

	input2, err := page.QuerySelector("#input2")
	require.NoError(t, err)
	isEditable, err = input2.IsEditable()
	require.NoError(t, err)
	require.True(t, isEditable)
	isEditable, err = page.Locator("#input2").IsEditable()
	require.NoError(t, err)
	require.True(t, isEditable)

	textarea, err := page.QuerySelector("textarea")
	require.NoError(t, err)
	isEditable, err = textarea.IsEditable()
	require.NoError(t, err)
	require.False(t, isEditable)
	isEditable, err = page.Locator("textarea").IsEditable()
	require.NoError(t, err)
	require.False(t, isEditable)
}

func TestElementHandleIsCheckedShouldWork(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`
		<input type="checkbox" checked><div>Not a checkbox</div>
	`))
	handle, err := page.QuerySelector("input")
	require.NoError(t, err)
	isChecked, err := handle.IsChecked()
	require.NoError(t, err)
	require.True(t, isChecked)
	isChecked, err = page.IsChecked("input")
	require.NoError(t, err)
	require.True(t, isChecked)

	_, err = handle.Evaluate("input => input.checked = false")
	require.NoError(t, err)
	isChecked, err = handle.IsChecked()
	require.NoError(t, err)
	require.False(t, isChecked)
	isChecked, err = page.Locator("input").IsChecked()
	require.NoError(t, err)
	require.False(t, isChecked)

	_, err = page.Locator("div").IsChecked()
	require.Contains(t, err.Error(), "Not a checkbox or radio button")
}

func TestElementHandleWaitForElementState(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent("<div><p id='result'>test result</p></div>"))

	handle, err := page.QuerySelector("#result")
	require.NoError(t, err)
	err = handle.WaitForElementState("visible")
	require.NoError(t, err)
}

func TestElementHandleWaitForSelector(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent("<div><p id='result'>test result</p></div>"))

	div, err := page.QuerySelector("div")
	require.NoError(t, err)

	handle, err := div.WaitForSelector("#result", playwright.ElementHandleWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateAttached,
	})
	require.NoError(t, err)
	text, err := handle.InnerText()
	require.NoError(t, err)
	require.Equal(t, "test result", text)
}

func TestElemetHandleFocus(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`<button onfocus="window.clicked=true"/>`))
	buttonElement, err := page.QuerySelector("button")
	require.NoError(t, err)
	require.NoError(t, buttonElement.Focus())
	result, err := page.Evaluate("window.clicked")
	require.NoError(t, err)
	require.True(t, result.(bool))
}

func TestElementHandleInputValue(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`
	<input></input>
	`))
	inputElement, err := page.QuerySelector("input")
	require.NoError(t, err)
	require.NoError(t, inputElement.Fill("test"))
	value, err := inputElement.InputValue()
	require.NoError(t, err)
	require.Equal(t, "test", value)
	require.NoError(t, inputElement.Fill(""))
	value, err = inputElement.InputValue()
	require.NoError(t, err)
	require.Equal(t, "", value)
}

func TestElementHandleSetChecked(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`<input id='checkbox' type='checkbox'></input>`))
	selectElement, err := page.QuerySelector("input")
	require.NoError(t, err)
	require.NoError(t, selectElement.SetChecked(true))
	isChecked, err := page.Evaluate("checkbox.checked")
	require.NoError(t, err)
	require.True(t, isChecked.(bool))
	require.NoError(t, selectElement.SetChecked(false))
	isChecked, err = page.Evaluate("checkbox.checked")
	require.NoError(t, err)
	require.False(t, isChecked.(bool))
}
