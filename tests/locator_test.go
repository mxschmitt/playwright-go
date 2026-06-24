package playwright_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestLocatorAllInnerTexts(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`<div>A</div><div>B</div><div>C</div>`))

	innerHTML, err := page.Locator("div").AllInnerTexts()
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"A", "B", "C"}, innerHTML)
}

func TestLocatorAllTextContents(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`<div>A</div><div>B</div><div>C</div>`))

	innerHTML, err := page.Locator("div").AllTextContents()
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"A", "B", "C"}, innerHTML)
}

func TestLocatorFill(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.PREFIX + "/dom.html")
	require.NoError(t, err)

	locator := page.Locator("#input")
	require.NoError(t, locator.Fill("input value"))
	result, err := locator.InputValue()
	require.NoError(t, err)
	require.Equal(t, "input value", result)
}

func TestLocatorGetAttribute(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.PREFIX + "/dom.html")
	require.NoError(t, err)

	result, err := page.Locator("#outer").GetAttribute("name")
	require.NoError(t, err)
	require.Equal(t, "value", result)
	result, err = page.Locator("#outer").GetAttribute("foo")
	require.NoError(t, err)
	require.Empty(t, result)
}

func TestLocatorInnerHTML(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.PREFIX + "/dom.html")
	require.NoError(t, err)

	result, err := page.Locator("#outer").InnerHTML()
	require.NoError(t, err)
	require.Equal(t, "<div id=\"inner\">Text,\nmore text</div>", result)
}

func TestLocatorInnerText(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.PREFIX + "/dom.html")
	require.NoError(t, err)

	result, err := page.Locator("#inner").InnerHTML()
	require.NoError(t, err)
	require.Equal(t, "Text,\nmore text", result)
}

func TestLocatorInputValue(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.PREFIX + "/dom.html")
	require.NoError(t, err)
	require.NoError(t, page.Locator("#input").Fill("input value"))

	result, err := page.Locator("#input").InputValue()
	require.NoError(t, err)
	require.Equal(t, "input value", result)
}

func TestLocatorIsChecked(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent("<input type='checkbox' checked><div>Not a checkbox</div>"))

	result, err := page.Locator("input").IsChecked()
	require.NoError(t, err)
	require.True(t, result)
}

func TestLocatorIsDisabled(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	err = page.SetContent(`
	<button disabled>button1</button>
	<button>button2</button>
	<div>div</div>
	`)
	require.NoError(t, err)

	result, err := page.Locator("div").IsDisabled()
	require.NoError(t, err)
	require.False(t, result)

	result, err = page.Locator(":text(\"button1\")").IsDisabled()
	require.NoError(t, err)
	require.True(t, result)

	result, err = page.Locator(":text(\"button2\")").IsDisabled()
	require.NoError(t, err)
	require.False(t, result)
}

func TestLocatorIsEditable(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	err = page.SetContent(`<input id=input1 disabled><textarea></textarea><input id=input2>
	`)
	require.NoError(t, err)

	result, err := page.Locator("#input1").IsEditable()
	require.NoError(t, err)
	require.False(t, result)

	result, err = page.Locator("#input2").IsEditable()
	require.NoError(t, err)
	require.True(t, result)

	result, err = page.Locator("textarea").IsEditable()
	require.NoError(t, err)
	require.True(t, result)
}

func TestLocatorIsEnabled(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	err = page.SetContent(`
	<button disabled>button1</button>
	<button>button2</button>
	<div>div</div>
	`)
	require.NoError(t, err)

	result, err := page.Locator("div").IsEnabled()
	require.NoError(t, err)
	require.True(t, result)

	result, err = page.Locator(":text(\"button1\")").IsEnabled()
	require.NoError(t, err)
	require.False(t, result)

	result, err = page.Locator(":text(\"button2\")").IsEnabled()
	require.NoError(t, err)
	require.True(t, result)
}

func TestLocatorIsHidden(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	err = page.SetContent(`
	<details>
		<summary>click to open</summary>
		<ul>
			<li>hidden item 1</li>
			<li>hidden item 2</li>
			<li>hidden item 3</li>
		</ul>
	</details>`)
	require.NoError(t, err)

	result, err := page.Locator("ul").IsHidden()
	require.NoError(t, err)
	require.True(t, result)

	result, err = page.Locator("summary").IsHidden()
	require.NoError(t, err)
	require.False(t, result)
}

func TestLocatorIsVisible(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	err = page.SetContent(`
	<details>
		<summary>click to open</summary>
		<ul>
			<li>hidden item 1</li>
			<li>hidden item 2</li>
			<li>hidden item 3</li>
		</ul>
	</details>`)
	require.NoError(t, err)

	result, err := page.Locator("ul").IsVisible()
	require.NoError(t, err)
	require.False(t, result)

	result, err = page.Locator("summary").IsVisible()
	require.NoError(t, err)
	require.True(t, result)
}

func TestLocatorLocatorHas(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)

	expText := " First item 1 First item 1A1"

	err = page.SetContent(`
	<section>
		<ul>
			<li>
				<div>
					<input class="r1" name="r1" type="checkbox"/>
					<span> First item 1</span>
					<span> First item 1A1</span>
				</div>
			</li>
			<li>
				<div>
					<input name="r2" type="checkbox"/>
					<span> Second item 1</span>
					<span> Second item 1<span>A1</span></span>
				</div>
			</li>
		</ul>
	</section>`)
	require.NoError(t, err)

	inputLocator := page.Locator("input[name='r1']")
	require.NoError(t, inputLocator.Err())

	listLocator := page.Locator("ul", playwright.PageLocatorOptions{Has: inputLocator})
	spanLocator := page.Locator("span", playwright.PageLocatorOptions{HasText: "First item 1A"})

	targetText, err := listLocator.Locator("li div", playwright.LocatorLocatorOptions{Has: spanLocator}).InnerText()
	require.NoError(t, err)
	require.Equal(t, expText, targetText)
}

func TestLocatorLocatorHasText(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)

	expText := "A1 B2"

	err = page.SetContent(`
	<section>
		<ul>
			<li>
				<div>
					<input name="r2" type="checkbox"/>
					<span> Second item a1</span>
					<span> Second item 1<span>A1 B2</span></span>
				</div>
			</li>
		</ul>
	</section>`)
	require.NoError(t, err)

	inputLocator := page.Locator("input[name='r2']")
	require.NoError(t, inputLocator.Err())

	listLocator := page.Locator("ul", playwright.PageLocatorOptions{Has: inputLocator})
	require.NoError(t, listLocator.Err())

	count, err := listLocator.Locator("li div span", playwright.LocatorLocatorOptions{HasText: "A1"}).Count()
	require.NoError(t, err)
	require.Equal(t, 3, count, "Locator count should be equal 3")

	targetText, err := listLocator.Locator("li div span", playwright.LocatorLocatorOptions{HasText: "1A1"}).
		Locator("span").InnerText()
	require.NoError(t, err)
	require.Equal(t, expText, targetText)
}

func TestLocatorSelectOption(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.PREFIX + "/dom.html")
	require.NoError(t, err)

	values := []string{"foo"}
	result, err := page.Locator("#select").SelectOption(playwright.SelectOptionValues{Values: &values})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"foo"}, result)
}

func TestLocatorTextContent(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.PREFIX + "/dom.html")
	require.NoError(t, err)

	result, err := page.Locator("#inner").TextContent()
	require.NoError(t, err)
	require.Equal(t, "Text,\nmore text", result)
}

func TestLocatorShouldFocusAndBlurButton(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.PREFIX + "/input/button.html")
	require.NoError(t, err)
	button := page.Locator("button")
	require.NoError(t, button.Err())
	ret, err := button.Evaluate(`button => document.activeElement === button`, nil)
	require.NoError(t, err)
	require.False(t, ret.(bool))

	var (
		focused = false
		blurred = false
	)
	require.NoError(t, page.ExposeFunction("focusEvent", func(args ...interface{}) interface{} {
		focused = true
		return nil
	}))
	require.NoError(t, page.ExposeFunction("blurEvent", func(args ...interface{}) interface{} {
		blurred = true
		return nil
	}))
	_, err = button.Evaluate(`button => {
		button.addEventListener('focus', window['focusEvent']);
		button.addEventListener('blur', window['blurEvent']);
}`, nil)
	require.NoError(t, err)

	require.NoError(t, button.Focus())
	ret, err = button.Evaluate(`button => document.activeElement === button`, nil)
	require.NoError(t, err)
	require.True(t, ret.(bool))
	require.True(t, focused)
	require.False(t, blurred)

	require.NoError(t, button.Blur())
	ret, err = button.Evaluate(`button => document.activeElement === button`, nil)
	require.NoError(t, err)
	require.False(t, ret.(bool))
	require.True(t, focused)
	require.True(t, blurred)
}

func TestLocatorAllShouldWork(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, page.SetContent(`<div><p>A</p><p>B</p><p>C</p></div>`))
	expected := []string{"A", "B", "C"}
	texts := make([]string, 0)

	locators, err := page.Locator("p").All()
	require.NoError(t, err)
	for _, locator := range locators {
		content, err := locator.TextContent()
		require.NoError(t, err)
		texts = append(texts, content)
	}
	require.ElementsMatch(t, expected, texts)
}

func TestLocatorsShouldReturnBoundingBox(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetViewportSize(500, 500))
	_, err := page.Goto(fmt.Sprintf("%s/grid.html", server.PREFIX))
	require.NoError(t, err)
	box, err := page.Locator(".box:nth-of-type(13)").BoundingBox()
	require.NoError(t, err)
	require.Equal(t, &playwright.Rect{
		X:      100,
		Y:      50,
		Width:  50,
		Height: 50,
	}, box)
}

func TestLocatorsCheckShouldWork(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`<input id='checkbox' type='checkbox'></input>`))
	require.NoError(t, page.Locator("input").Check())
	ret, err := page.Evaluate("checkbox.checked")
	require.NoError(t, err)
	require.True(t, ret.(bool))
	require.NoError(t, page.Locator("input").Uncheck())
	ret, err = page.Evaluate("checkbox.checked")
	require.NoError(t, err)
	require.False(t, ret.(bool))
}

func TestLocatorsClearShouldWork(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(fmt.Sprintf("%s/input/textarea.html", server.PREFIX))
	require.NoError(t, err)
	button := page.Locator("input")
	require.NoError(t, button.Fill("some value"))
	ret, err := page.Evaluate(`result`)
	require.NoError(t, err)
	require.Equal(t, "some value", ret)
	require.NoError(t, button.Clear(playwright.LocatorClearOptions{
		Timeout: playwright.Float(1000),
	}))
	ret, err = page.Evaluate(`result`)
	require.NoError(t, err)
	require.Equal(t, "", ret)
}

func TestLocatorsClickShouldWorkForTextNodes(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(fmt.Sprintf("%s/input/button.html", server.PREFIX))
	require.NoError(t, err)
	_, err = page.Evaluate(`
		() => {
			window['double'] = false;
			const button = document.querySelector('button');
			button.addEventListener('dblclick', event => {
				window['double'] = true;
			});
		}`)
	require.NoError(t, err)
	require.NoError(t, page.Locator("button").Dblclick())

	ret, err := page.Evaluate(`double`)
	require.NoError(t, err)
	require.True(t, ret.(bool))
	ret, err = page.Evaluate(`result`)
	require.NoError(t, err)
	require.Equal(t, "Clicked", ret)
}

func TestLocatorsDispatchEventShouldWork(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(fmt.Sprintf("%s/input/button.html", server.PREFIX))
	require.NoError(t, err)
	require.NoError(t, page.Locator("button").DispatchEvent("click", nil))
	ret, err := page.Evaluate(`result`)
	require.NoError(t, err)
	require.Equal(t, "Clicked", ret)
}

func TestLocatorsDragToShouldWork(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(fmt.Sprintf("%s/drag-n-drop.html", server.PREFIX))
	require.NoError(t, err)
	require.NoError(t, page.Locator("#source").DragTo(page.Locator("#target")))
	ret, err := page.Locator("#target").Evaluate("target => target.contains(document.querySelector('#source'))", nil)
	require.NoError(t, err)
	require.True(t, ret.(bool))
}

func TestLocatorDrop(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`
		<div id="target" style="width:200px;height:200px;border:1px solid"
		   ondrop="window.__dropped=event.dataTransfer.getData('text/plain');event.preventDefault()"
		   ondragover="event.preventDefault()"></div>`))

	// Drop a clipboard-style data payload. The serialized data must be an
	// array of {mimeType, value} entries for the protocol to accept it.
	require.NoError(t, page.Locator("#target").Drop(playwright.Payload{
		Data: map[string]string{"text/plain": "hello-drop"},
	}))
	dropped, err := page.Evaluate("() => window.__dropped")
	require.NoError(t, err)
	require.Equal(t, "hello-drop", dropped)

	// Drop a file payload, exercising the files -> payloads/localPaths conversion.
	require.NoError(t, page.Locator("#target").Drop(playwright.Payload{
		Files: []playwright.InputFile{{Name: "a.txt", MimeType: "text/plain", Buffer: []byte("hi")}},
	}))
}

func TestLocatorsShouldUploadFile(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(fmt.Sprintf("%s/input/fileupload.html", server.PREFIX))
	require.NoError(t, err)
	input := page.Locator("input[type=file]")
	require.NoError(t, input.SetInputFiles(Asset("file-to-upload.txt")))
	//nolint:staticcheck
	elm, err := input.ElementHandle()
	require.NoError(t, err)
	ret, err := page.Evaluate(`e => e.files[0].name`, elm)
	require.NoError(t, err)
	require.Equal(t, "file-to-upload.txt", ret)
}

func TestLocatorsShouldUploadFileRemote(t *testing.T) {
	BeforeEach(t)

	remoteServer, err := newRemoteServer()
	require.NoError(t, err)
	defer remoteServer.Close()

	browser1, err := browserType.Connect(remoteServer.url)
	require.NoError(t, err)
	require.NotNil(t, browser1)
	defer browser1.Close() //nolint:errcheck

	browser_context, err := browser1.NewContext()
	require.NoError(t, err)
	page1, err := browser_context.NewPage()
	require.NoError(t, err)
	_, err = page1.Goto(fmt.Sprintf("%s/input/fileupload.html", server.PREFIX))
	require.NoError(t, err)
	input := page1.Locator("input[type=file]")
	require.NoError(t, input.SetInputFiles(Asset("file-to-upload.txt")))
	//nolint:staticcheck
	elm, err := input.ElementHandle()
	require.NoError(t, err)
	ret, err := page1.Evaluate(`e => e.files[0].name`, elm)
	require.NoError(t, err)
	require.Equal(t, "file-to-upload.txt", ret)
}

func TestLocatorsShouldUploadFileUseBuffer(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(fmt.Sprintf("%s/input/fileupload.html", server.PREFIX))
	require.NoError(t, err)
	input := page.Locator("input[type=file]")
	file, err := os.ReadFile(Asset("file-to-upload.txt"))
	require.NoError(t, err)
	require.NoError(t, input.SetInputFiles([]playwright.InputFile{
		{
			Name:     "file-to-upload.txt",
			MimeType: "text/plain",
			Buffer:   file,
		},
	}))
	//nolint:staticcheck
	elm, err := input.ElementHandle()
	require.NoError(t, err)
	ret, err := page.Evaluate(`e => e.files[0].name`, elm)
	require.NoError(t, err)
	require.Equal(t, "file-to-upload.txt", ret)
}

func TestLocatorsShouldQueryExistingElements(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`<html><body><div>A</div><br/><div>B</div></body></html>`))
	//nolint:staticcheck
	elements, err := page.Locator("html").Locator("div").ElementHandles()
	require.NoError(t, err)
	require.Equal(t, 2, len(elements))
	results := make([]string, 0)
	for _, element := range elements {
		//nolint:staticcheck
		content, err := element.TextContent()
		require.NoError(t, err)
		results = append(results, content)
	}
	require.Equal(t, []string{"A", "B"}, results)
}

func TestLocatorsEvaluateAllShouldWork(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`<html><body><div class="tweet"><div class="like">100</div><div class="like">10</div></div></body></html>`))
	content, err := page.Locator(".tweet .like").EvaluateAll(`nodes => nodes.map(n => n.innerText)`)
	require.NoError(t, err)
	require.Equal(t, []interface{}{"100", "10"}, content)
}

func TestShouldSupportLocatorFilter(t *testing.T) {
	BeforeEach(t)

	err := page.SetContent(`<section><div><span>hello</span></div><div><span>world</span></div></section>`)
	require.NoError(t, err)
	locator := page.Locator("div").Filter(playwright.LocatorFilterOptions{
		HasText: "hello",
	})

	require.NoError(t, expect.Locator(locator).ToHaveCount(1))
}

func TestShouldSupportLocatorAnd(t *testing.T) {
	BeforeEach(t)

	err := page.SetContent(`
		<div data-testid=foo>hello</div><div data-testid=bar>world</div>
    <span data-testid=foo>hello2</span><span data-testid=bar>world2</span>`)
	require.NoError(t, err)
	require.NoError(t, expect.Locator(page.Locator("div").And(page.Locator("div"))).ToHaveCount(2))
	require.NoError(t, expect.Locator(page.Locator("div").And(page.GetByTestId("foo"))).ToHaveText([]string{"hello"}))
	require.NoError(t, expect.Locator(page.Locator("div").And(page.GetByTestId("bar"))).ToHaveText([]string{"world"}))
	require.NoError(t, expect.Locator(page.GetByTestId("foo").And(page.Locator("div"))).ToHaveText([]string{"hello"}))
	require.NoError(t, expect.Locator(page.GetByTestId("bar").And(page.Locator("span"))).ToHaveText([]string{"world2"}))
}

func TestShouldSupportLocatorOr(t *testing.T) {
	BeforeEach(t)

	err := page.SetContent(`<div>hello</div><span>world</span>`)
	require.NoError(t, err)
	require.NoError(t, expect.Locator(page.Locator("div").Or(page.Locator("span"))).ToHaveCount(2))
	require.NoError(t, expect.Locator(page.Locator("div").Or(page.Locator("span"))).ToHaveText([]string{"hello", "world"}))
	require.NoError(t, expect.Locator(
		page.Locator("span").Or(page.Locator("article")).Or(page.Locator("div")),
	).ToHaveText([]string{"hello", "world"}))

	require.NoError(t, expect.Locator(page.Locator("article").Or(page.Locator("something"))).ToHaveCount(0))
	require.NoError(t, expect.Locator(page.Locator("article").Or(page.Locator("div"))).ToHaveText("hello"))
	require.NoError(t, expect.Locator(page.Locator("article").Or(page.Locator("span"))).ToHaveText("world"))
	require.NoError(t, expect.Locator(page.Locator("div").Or(page.Locator("article"))).ToHaveText("hello"))
	require.NoError(t, expect.Locator(page.Locator("span").Or(page.Locator("article"))).ToHaveText("world"))
}

func TestLocatorAndFrameLocatorShouldAcceptLocator(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`
		<div><input value=outer></div>
    <iframe srcdoc="<div><input value=inner></div>"></iframe>
	`))
	input := page.Locator("input")
	require.NoError(t, expect.Locator(input).ToHaveValue("outer"))
	require.NoError(t, expect.Locator(page.Locator("div").Locator(input)).ToHaveValue("outer"))
	require.NoError(t, expect.Locator(page.FrameLocator("iframe").Locator(input)).ToHaveValue("inner"))
	require.NoError(t, expect.Locator(page.FrameLocator("iframe").Locator("div").Locator(input)).ToHaveValue("inner"))

	div := page.Locator("div")
	require.NoError(t, expect.Locator(page.FrameLocator("iframe").Locator(div).Locator("input")).ToHaveValue("inner"))
}

func TestLocatorShouldSupportLocatorWithAndOr(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`
		<div>one <span>two</span> <button>three</button> </div>
		<span>four</span>
		<button>five</button>
	`))

	require.NoError(t, expect.Locator(page.Locator("div").Locator(page.Locator("button"))).ToHaveText([]string{"three"}))
	require.NoError(t, expect.Locator(page.Locator("div").Locator(page.Locator("button").Or(page.Locator("span")))).
		ToHaveText([]string{"two", "three"}))
	require.NoError(t, expect.Locator(page.Locator("button").Or(page.Locator("span"))).
		ToHaveText([]string{"two", "three", "four", "five"}))

	require.NoError(t, expect.Locator(page.Locator("div").Locator(
		page.Locator("button").And(page.GetByRole("button")),
	)).ToHaveText([]string{"three"}))
	require.NoError(t, expect.Locator(page.Locator("button").And(page.GetByRole("button"))).
		ToHaveText([]string{"three", "five"}))
}

func TestLocatorHighlightShoudWork(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.PREFIX + "/grid.html")
	require.NoError(t, err)
	require.NoError(t, page.Locator(".box").Nth(3).Highlight())
	yes, err := page.Locator("x-pw-glass").IsVisible()
	require.NoError(t, err)
	require.True(t, yes)
}

func TestLocatorShouldType(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`<input type='text' />`))
	//nolint:staticcheck
	require.NoError(t, page.Locator("input").Type("hello"))
	utils.AssertResult(t, func() (interface{}, error) {
		//nolint:staticcheck
		return page.EvalOnSelector("input", "e => e.value", nil)
	}, "hello")
}

func TestLocatorShouldPressSequentially(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`<input type='text' />`))
	require.NoError(t, page.Locator("input").PressSequentially("hello"))
	utils.AssertResult(t, func() (interface{}, error) {
		//nolint:staticcheck
		return page.EvalOnSelector("input", "e => e.value", nil)
	}, "hello")
}

func TestLocatorShouldSupportFilterVisible(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`
		<div>
			<div class="item" style="display: none">Hidden data0</div>
			<div class="item">visible data1</div>
			<div class="item" style="display: none">Hidden data1</div>
			<div class="item">visible data2</div>
			<div class="item" style="display: none">Hidden data2</div>
			<div class="item">visible data3</div>
		</div>
		`))

	locator := page.Locator(".item").Filter(playwright.LocatorFilterOptions{
		Visible: playwright.Bool(true),
	}).Nth(1)

	require.NoError(t, expect.Locator(locator).ToHaveText("visible data2"))
	require.NoError(t, expect.Locator(page.Locator(".item").Filter(playwright.LocatorFilterOptions{
		Visible: playwright.Bool(true),
	}).GetByText("data3")).ToHaveText("visible data3"))
	require.NoError(t, expect.Locator(page.Locator(".item").Filter(playwright.LocatorFilterOptions{
		Visible: playwright.Bool(false),
	}).GetByText("data1")).ToHaveText("Hidden data1"))
}

// TestLocatorDescribe verifies that Locator.Describe() sets a description
// Based on upstream test: playwright/tests/page/locator-convenience.spec.ts
func TestLocatorDescribe(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`<button>Submit</button>`))

	locator := page.Locator("button")

	// Locator without description should return empty string
	desc, err := locator.Description()
	require.NoError(t, err)
	require.Empty(t, desc, "description should be empty for locator without description")

	// Set description
	describedLocator := locator.Describe("Submit button")
	desc, err = describedLocator.Description()
	require.NoError(t, err)
	require.Equal(t, "Submit button", desc)

	// Original locator should still have no description
	desc, err = locator.Description()
	require.NoError(t, err)
	require.Empty(t, desc, "original locator should remain unchanged")
}

// TestLocatorDescribeSpecialCharacters verifies descriptions with special characters
func TestLocatorDescribeSpecialCharacters(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`<div>Test</div>`))

	locator := page.Locator("div").Describe(`Button with "quotes" and 'apostrophes'`)
	desc, err := locator.Description()
	require.NoError(t, err)
	require.Equal(t, `Button with "quotes" and 'apostrophes'`, desc)
}

// TestLocatorDescribeChained verifies descriptions work with chained locators
func TestLocatorDescribeChained(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`
		<form>
			<input type="text" />
		</form>
	`))

	locator := page.Locator("form").Locator("input").Describe("Form input field")
	desc, err := locator.Description()
	require.NoError(t, err)
	require.Equal(t, "Form input field", desc)
}

// TestLocatorDescribeMultipleCalls verifies multiple describe calls override each other
func TestLocatorDescribeMultipleCalls(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.SetContent(`<button><span>Click me</span></button>`))

	// First description
	locator1 := page.Locator("button").Describe("First description")
	desc, err := locator1.Description()
	require.NoError(t, err)
	require.Equal(t, "First description", desc)

	// Second description on chained locator
	locator2 := locator1.Locator("span").Describe("Second description")
	desc, err = locator2.Description()
	require.NoError(t, err)
	require.Equal(t, "Second description", desc)

	// Chained locator without describe should have no description
	locator3 := locator2.Locator("span")
	desc, err = locator3.Description()
	require.NoError(t, err)
	require.Empty(t, desc, "chained locator without describe should have empty description")
}
