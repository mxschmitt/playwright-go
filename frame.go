package playwright

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
)

type frameImpl struct {
	channelOwner
	detached    bool
	page        *pageImpl
	name        string
	url         string
	parentFrame Frame
	childFrames []Frame
	loadStates  mapset.Set[string]
}

func newFrame(parent *channelOwner, objectType string, guid string, initializer map[string]any) *frameImpl {
	loadStates := mapset.NewSet[string]()
	// Initializers are JSON-decoded, so an array arrives as []any, never []string.
	if ls, ok := initializer["loadStates"].([]any); ok {
		for _, state := range ls {
			if s, ok := state.(string); ok {
				loadStates.Add(s)
			}
		}
	}
	f := &frameImpl{
		name:        initializer["name"].(string),
		url:         initializer["url"].(string),
		loadStates:  loadStates,
		childFrames: make([]Frame, 0),
	}
	f.createChannelOwner(f, parent, objectType, guid, initializer)

	channelOwner := fromNullableChannel(initializer["parentFrame"])
	if channelOwner != nil {
		f.parentFrame = channelOwner.(*frameImpl)
		f.parentFrame.(*frameImpl).childFrames = append(f.parentFrame.(*frameImpl).childFrames, f)
	}

	f.channel.On("navigated", f.onFrameNavigated)
	f.channel.On("loadstate", f.onLoadState)
	return f
}

func (f *frameImpl) URL() string {
	f.RLock()
	defer f.RUnlock()
	return f.url
}

func (f *frameImpl) Name() string {
	f.RLock()
	defer f.RUnlock()
	return f.name
}

func (f *frameImpl) SetContent(content string, options ...FrameSetContentOptions) error {
	overrides := map[string]any{
		"html": content,
	}
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	_, err := f.channel.SendWithTimeout("setContent", resolveNavigationTimeout(f.page.timeoutSettings, explicit), overrides, options)
	return err
}

func (f *frameImpl) Content() (string, error) {
	content, err := f.channel.Send("content")
	if content == nil {
		return "", err
	}
	return content.(string), err
}

func (f *frameImpl) Goto(url string, options ...FrameGotoOptions) (Response, error) {
	overrides := map[string]any{
		"url": url,
	}
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	channel, err := f.channel.SendWithTimeout("goto", resolveNavigationTimeout(f.page.timeoutSettings, explicit), overrides, options)
	if err != nil {
		return nil, fmt.Errorf("Frame.Goto %s: %w", url, err)
	}
	channelOwner := fromNullableChannel(channel)
	if channelOwner == nil {
		// navigation to about:blank or navigation to the same URL with a different hash
		return nil, nil
	}
	return channelOwner.(*responseImpl), nil
}

func (f *frameImpl) AddScriptTag(options FrameAddScriptTagOptions) (ElementHandle, error) {
	if options.Path != nil {
		file, err := os.ReadFile(*options.Path)
		if err != nil {
			return nil, err
		}
		// Append a sourceURL so the injected script is attributed to its file
		// path in DevTools/traces, matching upstream addSourceUrlToScript.
		options.Content = String(addSourceURLToScript(string(file), *options.Path))
		options.Path = nil
	}
	channel, err := f.channel.Send("addScriptTag", options)
	if err != nil {
		return nil, err
	}
	return fromChannel(channel).(*elementHandleImpl), nil
}

func (f *frameImpl) AddStyleTag(options FrameAddStyleTagOptions) (ElementHandle, error) {
	if options.Path != nil {
		file, err := os.ReadFile(*options.Path)
		if err != nil {
			return nil, err
		}
		options.Content = String(string(file) + "/*# sourceURL=" + strings.ReplaceAll(*options.Path, "\n", "") + "*/")
		options.Path = nil
	}
	channel, err := f.channel.Send("addStyleTag", options)
	if err != nil {
		return nil, err
	}
	return fromChannel(channel).(*elementHandleImpl), nil
}

func (f *frameImpl) Page() Page {
	return f.page
}

func (f *frameImpl) WaitForLoadState(options ...FrameWaitForLoadStateOptions) error {
	option := FrameWaitForLoadStateOptions{}
	if len(options) == 1 {
		option = options[0]
	}
	if option.State == nil {
		option.State = LoadStateLoad
	}
	return f.waitForLoadStateImpl(string(*option.State), option.Timeout, nil)
}

func (f *frameImpl) waitForLoadStateImpl(state string, timeout *float64, cb func() error) error {
	if f.loadStates.ContainsOne(state) {
		return nil
	}
	waiter, err := f.setNavigationWaiter(timeout)
	if err != nil {
		return err
	}
	waiter.WaitForEvent(f, "loadstate", func(payload any) bool {
		gotState := payload.(string)
		return gotState == state
	})
	if cb == nil {
		_, err := waiter.Wait()
		return err
	} else {
		_, err := waiter.RunAndWait(cb)
		return err
	}
}

func (f *frameImpl) WaitForURL(url any, options ...FrameWaitForURLOptions) error {
	if f.page == nil {
		return errors.New("frame is detached")
	}
	matcher := newURLMatcher(url, f.page.browserContext.options.BaseURL)
	if matcher.Matches(f.URL()) {
		state := "load"
		timeout := Float(f.page.timeoutSettings.NavigationTimeout())
		if len(options) == 1 {
			if options[0].WaitUntil != nil {
				state = string(*options[0].WaitUntil)
			}
			if options[0].Timeout != nil {
				timeout = options[0].Timeout
			}
		}
		return f.waitForLoadStateImpl(state, timeout, nil)
	}
	navigationOptions := FrameExpectNavigationOptions{URL: url}
	if len(options) > 0 {
		navigationOptions.Timeout = options[0].Timeout
		navigationOptions.WaitUntil = options[0].WaitUntil
	}
	if _, err := f.ExpectNavigation(nil, navigationOptions); err != nil {
		return err
	}
	return nil
}

func (f *frameImpl) ExpectNavigation(cb func() error, options ...FrameExpectNavigationOptions) (Response, error) {
	if f.page == nil {
		return nil, errors.New("frame is detached")
	}
	option := FrameExpectNavigationOptions{}
	if len(options) == 1 {
		option = options[0]
	}
	if option.WaitUntil == nil {
		option.WaitUntil = WaitUntilStateLoad
	}
	if option.Timeout == nil {
		option.Timeout = Float(f.page.timeoutSettings.NavigationTimeout())
	}
	// A zero timeout disables the timeout. For positive values, use one deadline
	// across both the navigation event and the requested load state.
	var deadline *time.Time
	if *option.Timeout > 0 {
		d := time.Now().Add(time.Duration(*option.Timeout) * time.Millisecond)
		deadline = &d
	}
	var matcher *urlMatcher
	if option.URL != nil {
		matcher = newURLMatcher(option.URL, f.page.browserContext.options.BaseURL)
	}
	predicate := func(events ...any) bool {
		ev := events[0].(map[string]any)
		err, ok := ev["error"]
		if ok {
			// Any failed navigation results in a rejection.
			logger.Error("navigation error", "url", ev["url"].(string), "error", err)
			return true
		}
		return matcher == nil || matcher.Matches(ev["url"].(string))
	}
	waiter, err := f.setNavigationWaiter(option.Timeout)
	if err != nil {
		return nil, err
	}

	eventData, err := waiter.WaitForEvent(f, "navigated", predicate).RunAndWait(cb)
	if err != nil || eventData == nil {
		return nil, err
	}

	event := eventData.(map[string]any)
	if errVal, ok := event["error"]; ok {
		// Any failed navigation results in a rejection.
		return nil, errors.New(errVal.(string))
	}

	remaining := option.Timeout
	if deadline != nil {
		ms := float64(time.Until(*deadline).Milliseconds())
		// The waiter interprets zero as unlimited, so use the smallest positive
		// timeout when the shared budget has just been exhausted.
		if ms <= 0 {
			ms = 1
		}
		remaining = Float(ms)
	}
	if err = f.waitForLoadStateImpl(string(*option.WaitUntil), remaining, nil); err != nil {
		return nil, err
	}
	if event["newDocument"] != nil && event["newDocument"].(map[string]any)["request"] != nil {
		request := fromChannel(event["newDocument"].(map[string]any)["request"]).(*requestImpl)
		// The response lives on the final request after following any redirects.
		return request.finalRequest().Response()
	}
	return nil, nil
}

func (f *frameImpl) setNavigationWaiter(timeout *float64) (*waiter, error) {
	if f.page == nil {
		return nil, errors.New("page does not exist")
	}
	waiter := newWaiter()
	if timeout != nil {
		waiter.WithTimeout(*timeout)
	} else {
		waiter.WithTimeout(f.page.timeoutSettings.NavigationTimeout())
	}
	// If the page is already closed, fail immediately rather than waiting for the
	// (already-fired) close event or the navigation timeout, matching upstream's
	// rejectImmediately guard.
	if f.page.IsClosed() {
		waiter.reject(f.page.closeErrorWithReason())
	}
	waiter.RejectOnEvent(f.page, "close", f.page.closeErrorWithReason())
	waiter.RejectOnEvent(f.page, "crash", fmt.Errorf("Navigation failed because page crashed!"))
	waiter.RejectOnEvent(f.page, "framedetached", fmt.Errorf("Navigating frame was detached!"), func(payload any) bool {
		frame, ok := payload.(*frameImpl)
		if ok && frame == f {
			return true
		}
		return false
	})
	return waiter, nil
}

func (f *frameImpl) onFrameNavigated(ev map[string]any) {
	f.Lock()
	f.url = ev["url"].(string)
	f.name = ev["name"].(string)
	f.Unlock()
	f.Emit("navigated", ev)
	_, ok := ev["error"]
	if !ok && f.page != nil {
		f.page.Emit("framenavigated", f)
		if f.page.browserContext != nil {
			f.page.browserContext.Emit("framenavigated", f)
		}
	}
}

func (f *frameImpl) onLoadState(ev map[string]any) {
	if ev["add"] != nil {
		add := ev["add"].(string)
		f.loadStates.Add(add)
		f.Emit("loadstate", add)
		if f.parentFrame == nil && f.page != nil {
			if add == "load" || add == "domcontentloaded" {
				f.Page().Emit(add, f.page)
				if add == "load" && f.page.browserContext != nil {
					f.page.browserContext.Emit("pageload", f.page)
				}
			}
		}
	} else if ev["remove"] != nil {
		remove := ev["remove"].(string)
		f.loadStates.Remove(remove)
	}
}

func (f *frameImpl) QuerySelector(selector string, options ...FrameQuerySelectorOptions) (ElementHandle, error) {
	params := map[string]any{
		"selector": selector,
	}
	if len(options) == 1 {
		params["strict"] = options[0].Strict
	}
	channel, err := f.channel.Send("querySelector", params)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, nil
	}
	return fromChannel(channel).(*elementHandleImpl), nil
}

func (f *frameImpl) QuerySelectorAll(selector string) ([]ElementHandle, error) {
	channels, err := f.channel.Send("querySelectorAll", map[string]any{
		"selector": selector,
	})
	if err != nil {
		return nil, err
	}
	elements := make([]ElementHandle, 0)
	for _, channel := range channels.([]any) {
		elements = append(elements, fromChannel(channel).(*elementHandleImpl))
	}
	return elements, nil
}

func (f *frameImpl) Evaluate(expression string, options ...any) (any, error) {
	var arg any
	if len(options) == 1 {
		arg = options[0]
	}
	result, err := f.channel.Send("evaluateExpression", map[string]any{
		"expression": expression,
		"arg":        serializeArgument(arg),
	})
	if err != nil {
		return nil, err
	}
	return parseResult(result), nil
}

func (f *frameImpl) EvalOnSelector(selector string, expression string, arg any, options ...FrameEvalOnSelectorOptions) (any, error) {
	params := map[string]any{
		"selector":   selector,
		"expression": expression,
		"arg":        serializeArgument(arg),
	}
	if len(options) == 1 && options[0].Strict != nil {
		params["strict"] = *options[0].Strict
	}

	result, err := f.channel.Send("evalOnSelector", params)
	if err != nil {
		return nil, err
	}
	return parseResult(result), nil
}

func (f *frameImpl) EvalOnSelectorAll(selector string, expression string, options ...any) (any, error) {
	var arg any
	if len(options) == 1 {
		arg = options[0]
	}
	result, err := f.channel.Send("evalOnSelectorAll", map[string]any{
		"selector":   selector,
		"expression": expression,
		"arg":        serializeArgument(arg),
	})
	if err != nil {
		return nil, err
	}
	return parseResult(result), nil
}

func (f *frameImpl) EvaluateHandle(expression string, options ...any) (JSHandle, error) {
	var arg any
	if len(options) == 1 {
		arg = options[0]
	}
	result, err := f.channel.Send("evaluateExpressionHandle", map[string]any{
		"expression": expression,
		"arg":        serializeArgument(arg),
	})
	if err != nil {
		return nil, err
	}
	channelOwner := fromChannel(result)
	if channelOwner == nil {
		return nil, nil
	}
	return channelOwner.(JSHandle), nil
}

func (f *frameImpl) Click(selector string, options ...FrameClickOptions) error {
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	_, err := f.channel.SendWithTimeout("click", resolveTimeout(f.page.timeoutSettings, explicit), map[string]any{
		"selector": selector,
	}, options)
	return err
}

func (f *frameImpl) WaitForSelector(selector string, options ...FrameWaitForSelectorOptions) (ElementHandle, error) {
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	channel, err := f.channel.SendWithTimeout("waitForSelector", resolveTimeout(f.page.timeoutSettings, explicit), map[string]any{
		"selector": selector,
	}, options)
	if err != nil {
		return nil, err
	}
	channelOwner := fromNullableChannel(channel)
	if channelOwner == nil {
		return nil, nil
	}
	return channelOwner.(*elementHandleImpl), nil
}

func (f *frameImpl) DispatchEvent(selector, typ string, eventInit any, options ...FrameDispatchEventOptions) error {
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	_, err := f.channel.SendWithTimeout("dispatchEvent", resolveTimeout(f.page.timeoutSettings, explicit), map[string]any{
		"selector":  selector,
		"type":      typ,
		"eventInit": serializeArgument(eventInit),
	}, options)
	return err
}

func (f *frameImpl) InnerText(selector string, options ...FrameInnerTextOptions) (string, error) {
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	innerText, err := f.channel.SendWithTimeout("innerText", resolveTimeout(f.page.timeoutSettings, explicit), map[string]any{
		"selector": selector,
	}, options)
	if innerText == nil {
		return "", err
	}
	return innerText.(string), err
}

func (f *frameImpl) InnerHTML(selector string, options ...FrameInnerHTMLOptions) (string, error) {
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	innerHTML, err := f.channel.SendWithTimeout("innerHTML", resolveTimeout(f.page.timeoutSettings, explicit), map[string]any{
		"selector": selector,
	}, options)
	if innerHTML == nil {
		return "", err
	}
	return innerHTML.(string), err
}

func (f *frameImpl) GetAttribute(selector string, name string, options ...FrameGetAttributeOptions) (string, error) {
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	attribute, err := f.channel.SendWithTimeout("getAttribute", resolveTimeout(f.page.timeoutSettings, explicit), map[string]any{
		"selector": selector,
		"name":     name,
	}, options)
	if attribute == nil {
		return "", err
	}
	return attribute.(string), err
}

func (f *frameImpl) Hover(selector string, options ...FrameHoverOptions) error {
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	_, err := f.channel.SendWithTimeout("hover", resolveTimeout(f.page.timeoutSettings, explicit), map[string]any{
		"selector": selector,
	}, options)
	return err
}

func (f *frameImpl) SetInputFiles(selector string, files any, options ...FrameSetInputFilesOptions) error {
	params, err := convertInputFiles(files, f.page.browserContext)
	if err != nil {
		return err
	}
	params.Selector = &selector
	var option FrameSetInputFilesOptions
	if len(options) == 1 {
		option = options[0]
	}
	_, err = f.channel.SendWithTimeout("setInputFiles", resolveTimeout(f.page.timeoutSettings, option.Timeout), params, option)
	return err
}

func (f *frameImpl) Type(selector, text string, options ...FrameTypeOptions) error {
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	_, err := f.channel.SendWithTimeout("type", resolveTimeout(f.page.timeoutSettings, explicit), map[string]any{
		"selector": selector,
		"text":     text,
	}, options)
	return err
}

func (f *frameImpl) Press(selector, key string, options ...FramePressOptions) error {
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	_, err := f.channel.SendWithTimeout("press", resolveTimeout(f.page.timeoutSettings, explicit), map[string]any{
		"selector": selector,
		"key":      key,
	}, options)
	return err
}

func (f *frameImpl) Check(selector string, options ...FrameCheckOptions) error {
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	_, err := f.channel.SendWithTimeout("check", resolveTimeout(f.page.timeoutSettings, explicit), map[string]any{
		"selector": selector,
	}, options)
	return err
}

func (f *frameImpl) Uncheck(selector string, options ...FrameUncheckOptions) error {
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	_, err := f.channel.SendWithTimeout("uncheck", resolveTimeout(f.page.timeoutSettings, explicit), map[string]any{
		"selector": selector,
	}, options)
	return err
}

func (f *frameImpl) WaitForTimeout(timeout float64) {
	_, _ = f.channel.SendWithTimeout("waitForTimeout", Float(0), map[string]any{
		"waitTimeout": timeout,
	})
}

func (f *frameImpl) WaitForFunction(expression string, arg any, options ...FrameWaitForFunctionOptions) (JSHandle, error) {
	var option FrameWaitForFunctionOptions
	if len(options) == 1 {
		option = options[0]
	}
	overrides := map[string]any{
		"expression": expression,
		"arg":        serializeArgument(arg),
	}
	// The server expects a numeric `pollingInterval`; the string "raf" means
	// "poll on requestAnimationFrame" and is conveyed by omitting the interval.
	switch polling := option.Polling.(type) {
	case string:
		if polling != "raf" {
			return nil, fmt.Errorf("Unknown polling option: %s", polling)
		}
	case nil:
	default:
		overrides["pollingInterval"] = option.Polling
	}
	result, err := f.channel.SendWithTimeout("waitForFunction", resolveTimeout(f.page.timeoutSettings, option.Timeout), overrides)
	if err != nil {
		return nil, err
	}
	handle := fromChannel(result)
	if handle == nil {
		return nil, nil
	}
	return handle.(JSHandle), nil
}

func (f *frameImpl) Title() (string, error) {
	title, err := f.channel.Send("title")
	if title == nil {
		return "", err
	}
	return title.(string), err
}

func (f *frameImpl) ChildFrames() []Frame {
	return f.childFrames
}

func (f *frameImpl) Dblclick(selector string, options ...FrameDblclickOptions) error {
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	_, err := f.channel.SendWithTimeout("dblclick", resolveTimeout(f.page.timeoutSettings, explicit), map[string]any{
		"selector": selector,
	}, options)
	return err
}

func (f *frameImpl) Fill(selector string, value string, options ...FrameFillOptions) error {
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	_, err := f.channel.SendWithTimeout("fill", resolveTimeout(f.page.timeoutSettings, explicit), map[string]any{
		"selector": selector,
		"value":    value,
	}, options)
	return err
}

func (f *frameImpl) Focus(selector string, options ...FrameFocusOptions) error {
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	_, err := f.channel.SendWithTimeout("focus", resolveTimeout(f.page.timeoutSettings, explicit), map[string]any{
		"selector": selector,
	}, options)
	return err
}

func (f *frameImpl) FrameElement() (ElementHandle, error) {
	channel, err := f.channel.Send("frameElement")
	if err != nil {
		return nil, err
	}
	return fromChannel(channel).(*elementHandleImpl), nil
}

func (f *frameImpl) IsDetached() bool {
	return f.detached
}

func (f *frameImpl) ParentFrame() Frame {
	return f.parentFrame
}

func (f *frameImpl) TextContent(selector string, options ...FrameTextContentOptions) (string, error) {
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	textContent, err := f.channel.SendWithTimeout("textContent", resolveTimeout(f.page.timeoutSettings, explicit), map[string]any{
		"selector": selector,
	}, options)
	if textContent == nil {
		return "", err
	}
	return textContent.(string), err
}

func (f *frameImpl) Tap(selector string, options ...FrameTapOptions) error {
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	_, err := f.channel.SendWithTimeout("tap", resolveTimeout(f.page.timeoutSettings, explicit), map[string]any{
		"selector": selector,
	}, options)
	return err
}

func (f *frameImpl) SelectOption(selector string, values SelectOptionValues, options ...FrameSelectOptionOptions) ([]string, error) {
	opts := convertSelectOptionSet(values)

	m := make(map[string]any)
	m["selector"] = selector
	for k, v := range opts {
		m[k] = v
	}
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	selected, err := f.channel.SendWithTimeout("selectOption", resolveTimeout(f.page.timeoutSettings, explicit), m, options)
	if err != nil {
		return nil, err
	}

	return transformToStringList(selected), nil
}

func (f *frameImpl) IsChecked(selector string, options ...FrameIsCheckedOptions) (bool, error) {
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	checked, err := f.channel.SendWithTimeout("isChecked", resolveTimeout(f.page.timeoutSettings, explicit), map[string]any{
		"selector": selector,
	}, options)
	if err != nil {
		return false, err
	}
	return checked.(bool), nil
}

func (f *frameImpl) IsDisabled(selector string, options ...FrameIsDisabledOptions) (bool, error) {
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	disabled, err := f.channel.SendWithTimeout("isDisabled", resolveTimeout(f.page.timeoutSettings, explicit), map[string]any{
		"selector": selector,
	}, options)
	if err != nil {
		return false, err
	}
	return disabled.(bool), nil
}

func (f *frameImpl) IsEditable(selector string, options ...FrameIsEditableOptions) (bool, error) {
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	editable, err := f.channel.SendWithTimeout("isEditable", resolveTimeout(f.page.timeoutSettings, explicit), map[string]any{
		"selector": selector,
	}, options)
	if err != nil {
		return false, err
	}
	return editable.(bool), nil
}

func (f *frameImpl) IsEnabled(selector string, options ...FrameIsEnabledOptions) (bool, error) {
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	enabled, err := f.channel.SendWithTimeout("isEnabled", resolveTimeout(f.page.timeoutSettings, explicit), map[string]any{
		"selector": selector,
	}, options)
	if err != nil {
		return false, err
	}
	return enabled.(bool), nil
}

func (f *frameImpl) IsHidden(selector string, options ...FrameIsHiddenOptions) (bool, error) {
	// Timeout is deprecated and intentionally ignored for this immediate query.
	if len(options) == 1 {
		options[0].Timeout = nil
	}
	hidden, err := f.channel.SendWithTimeout("isHidden", Float(0), map[string]any{
		"selector": selector,
	}, options)
	if err != nil {
		return false, err
	}
	return hidden.(bool), nil
}

func (f *frameImpl) IsVisible(selector string, options ...FrameIsVisibleOptions) (bool, error) {
	// Timeout is deprecated and intentionally ignored for this immediate query.
	if len(options) == 1 {
		options[0].Timeout = nil
	}
	visible, err := f.channel.SendWithTimeout("isVisible", Float(0), map[string]any{
		"selector": selector,
	}, options)
	if err != nil {
		return false, err
	}
	return visible.(bool), nil
}

func (f *frameImpl) InputValue(selector string, options ...FrameInputValueOptions) (string, error) {
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	value, err := f.channel.SendWithTimeout("inputValue", resolveTimeout(f.page.timeoutSettings, explicit), map[string]any{
		"selector": selector,
	}, options)
	if value == nil {
		return "", err
	}
	return value.(string), err
}

func (f *frameImpl) DragAndDrop(source, target string, options ...FrameDragAndDropOptions) error {
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	_, err := f.channel.SendWithTimeout("dragAndDrop", resolveTimeout(f.page.timeoutSettings, explicit), map[string]any{
		"source": source,
		"target": target,
	}, options)
	return err
}

func (f *frameImpl) SetChecked(selector string, checked bool, options ...FrameSetCheckedOptions) error {
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	timeout := resolveTimeout(f.page.timeoutSettings, explicit)
	if checked {
		_, err := f.channel.SendWithTimeout("check", timeout, map[string]any{
			"selector": selector,
		}, options)
		return err
	}
	_, err := f.channel.SendWithTimeout("uncheck", timeout, map[string]any{
		"selector": selector,
	}, options)
	return err
}

func (f *frameImpl) Locator(selector string, options ...FrameLocatorOptions) Locator {
	var option LocatorOptions
	if len(options) == 1 {
		option = LocatorOptions{
			Has:        options[0].Has,
			HasNot:     options[0].HasNot,
			HasText:    options[0].HasText,
			HasNotText: options[0].HasNotText,
		}
	}
	return newLocator(f, selector, option)
}

func (f *frameImpl) GetByAltText(text any, options ...FrameGetByAltTextOptions) Locator {
	exact := false
	if len(options) == 1 {
		if options[0].Exact != nil && *options[0].Exact {
			exact = true
		}
	}
	selector, err := getByAltTextSelector(text, exact)
	if err != nil {
		return newErrorLocator(f, err)
	}
	return f.Locator(selector)
}

func (f *frameImpl) GetByLabel(text any, options ...FrameGetByLabelOptions) Locator {
	exact := false
	if len(options) == 1 {
		if options[0].Exact != nil && *options[0].Exact {
			exact = true
		}
	}
	selector, err := getByLabelSelector(text, exact)
	if err != nil {
		return newErrorLocator(f, err)
	}
	return f.Locator(selector)
}

func (f *frameImpl) GetByPlaceholder(text any, options ...FrameGetByPlaceholderOptions) Locator {
	exact := false
	if len(options) == 1 {
		if options[0].Exact != nil && *options[0].Exact {
			exact = true
		}
	}
	selector, err := getByPlaceholderSelector(text, exact)
	if err != nil {
		return newErrorLocator(f, err)
	}
	return f.Locator(selector)
}

func (f *frameImpl) GetByRole(role AriaRole, options ...FrameGetByRoleOptions) Locator {
	if len(options) == 1 {
		selector, err := getByRoleSelector(role, LocatorGetByRoleOptions(options[0]))
		if err != nil {
			return newErrorLocator(f, err)
		}
		return f.Locator(selector)
	}
	selector, err := getByRoleSelector(role)
	if err != nil {
		return newErrorLocator(f, err)
	}
	return f.Locator(selector)
}

func (f *frameImpl) GetByTestId(testId any) Locator {
	selector, err := getByTestIdSelector(getTestIdAttributeName(), testId)
	if err != nil {
		return newErrorLocator(f, err)
	}
	return f.Locator(selector)
}

func (f *frameImpl) GetByText(text any, options ...FrameGetByTextOptions) Locator {
	exact := false
	if len(options) == 1 {
		if options[0].Exact != nil && *options[0].Exact {
			exact = true
		}
	}
	selector, err := getByTextSelector(text, exact)
	if err != nil {
		return newErrorLocator(f, err)
	}
	return f.Locator(selector)
}

func (f *frameImpl) GetByTitle(text any, options ...FrameGetByTitleOptions) Locator {
	exact := false
	if len(options) == 1 {
		if options[0].Exact != nil && *options[0].Exact {
			exact = true
		}
	}
	selector, err := getByTitleSelector(text, exact)
	if err != nil {
		return newErrorLocator(f, err)
	}
	return f.Locator(selector)
}

func (f *frameImpl) FrameLocator(selector string) FrameLocator {
	return newFrameLocator(f, selector)
}

func (f *frameImpl) highlight(selector string) error {
	_, err := f.channel.Send("highlight", map[string]any{
		"selector": selector,
	})
	return err
}

func (f *frameImpl) queryCount(selector string) (int, error) {
	response, err := f.channel.Send("queryCount", map[string]any{
		"selector": selector,
	})
	if err != nil {
		return 0, err
	}
	return int(response.(float64)), nil
}
