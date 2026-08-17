package playwright

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

var (
	testIdAttributeName    = "data-testid"
	ErrLocatorNotSameFrame = errors.New("inner 'has' or 'hasNot' locator must belong to the same frame")
)

type locatorImpl struct {
	frame    *frameImpl
	selector string
	options  *LocatorOptions
	err      error
}

type LocatorOptions LocatorFilterOptions

func newLocator(frame *frameImpl, selector string, options ...LocatorOptions) *locatorImpl {
	option := &LocatorOptions{}
	if len(options) == 1 {
		option = &options[0]
	}
	locator := &locatorImpl{frame: frame, selector: selector, options: option, err: nil}
	if option.HasText != nil {
		escaped, err := escapeForTextSelector(option.HasText, false)
		if err != nil {
			locator.err = errors.Join(locator.err, fmt.Errorf("HasText: %w", err))
		} else {
			selector += fmt.Sprintf(` >> internal:has-text=%s`, escaped)
		}
	}
	if option.HasNotText != nil {
		escaped, err := escapeForTextSelector(option.HasNotText, false)
		if err != nil {
			locator.err = errors.Join(locator.err, fmt.Errorf("HasNotText: %w", err))
		} else {
			selector += fmt.Sprintf(` >> internal:has-not-text=%s`, escaped)
		}
	}
	if option.Has != nil {
		has := option.Has.(*locatorImpl)
		if frame != has.frame {
			locator.err = errors.Join(locator.err, ErrLocatorNotSameFrame)
		} else {
			selector += fmt.Sprintf(` >> internal:has=%s`, escapeText(has.selector))
		}
	}
	if option.HasNot != nil {
		hasNot := option.HasNot.(*locatorImpl)
		if frame != hasNot.frame {
			locator.err = errors.Join(locator.err, ErrLocatorNotSameFrame)
		} else {
			selector += fmt.Sprintf(` >> internal:has-not=%s`, escapeText(hasNot.selector))
		}
	}
	if option.Visible != nil {
		selector += fmt.Sprintf(` >> visible=%s`, strconv.FormatBool(*option.Visible))
	}

	locator.selector = selector

	return locator
}

// newErrorLocator returns a Locator that only carries err. The GetBy* methods
// return a Locator with no error slot of their own, so an unsupported argument
// type is reported the same way an invalid selector parameter already is: the
// error surfaces via Err() and from every action on the locator, rather than
// panicking. Upstream needs no equivalent because its `string | RegExp`
// parameter types reject these arguments at compile time.
func newErrorLocator(frame *frameImpl, err error) *locatorImpl {
	return &locatorImpl{frame: frame, options: &LocatorOptions{}, err: err}
}

func (l *locatorImpl) equals(locator Locator) bool {
	return l.frame == locator.(*locatorImpl).frame && l.err == locator.(*locatorImpl).err && l.selector == locator.(*locatorImpl).selector
}

// withError returns a copy of the locator carrying an additional error, without
// mutating the receiver. Upstream throws on these conditions; the Go API surfaces
// the error lazily via the returned locator's Err().
func (l *locatorImpl) withError(err error) *locatorImpl {
	return &locatorImpl{
		frame:    l.frame,
		selector: l.selector,
		options:  l.options,
		err:      errors.Join(l.err, err),
	}
}

func (l *locatorImpl) Err() error {
	return l.err
}

func (l *locatorImpl) Describe(description string) Locator {
	// Embed the description into the selector via the internal:describe engine so
	// it reaches the server (traces, error messages, call logs), matching upstream.
	return newLocator(l.frame, l.selector+" >> internal:describe="+escapeText(description))
}

func (l *locatorImpl) Description() (string, error) {
	return locatorCustomDescription(l.selector), nil
}

func (l *locatorImpl) All() ([]Locator, error) {
	result := make([]Locator, 0)
	count, err := l.Count()
	if err != nil {
		return nil, err
	}
	for i := range count {
		result = append(result, l.Nth(i))
	}
	return result, nil
}

func (l *locatorImpl) AllInnerTexts() ([]string, error) {
	if l.err != nil {
		return nil, l.err
	}
	innerTexts, err := l.frame.EvalOnSelectorAll(l.selector, "ee => ee.map(e => e.innerText)")
	if err != nil {
		return nil, err
	}
	texts := innerTexts.([]any)
	result := make([]string, len(texts))
	for i := range texts {
		result[i] = texts[i].(string)
	}
	return result, nil
}

func (l *locatorImpl) AllTextContents() ([]string, error) {
	if l.err != nil {
		return nil, l.err
	}
	textContents, err := l.frame.EvalOnSelectorAll(l.selector, "ee => ee.map(e => e.textContent || '')")
	if err != nil {
		return nil, err
	}
	texts := textContents.([]any)
	result := make([]string, len(texts))
	for i := range texts {
		result[i] = texts[i].(string)
	}
	return result, nil
}

func (l *locatorImpl) And(locator Locator) Locator {
	if l.frame != locator.(*locatorImpl).frame {
		return l.withError(ErrLocatorNotSameFrame)
	}
	return newLocator(l.frame, l.selector+` >> internal:and=`+escapeText(locator.(*locatorImpl).selector))
}

func (l *locatorImpl) Or(locator Locator) Locator {
	if l.frame != locator.(*locatorImpl).frame {
		return l.withError(ErrLocatorNotSameFrame)
	}
	return newLocator(l.frame, l.selector+` >> internal:or=`+escapeText(locator.(*locatorImpl).selector))
}

func (l *locatorImpl) Blur(options ...LocatorBlurOptions) error {
	if l.err != nil {
		return l.err
	}
	params := map[string]any{
		"selector": l.selector,
		"strict":   true,
	}
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	_, err := l.frame.channel.SendWithTimeout("blur", resolveTimeout(l.frame.page.timeoutSettings, explicit), params)
	return err
}

func (l *locatorImpl) AriaSnapshot(options ...LocatorAriaSnapshotOptions) (string, error) {
	if l.err != nil {
		return "", l.err
	}
	var option LocatorAriaSnapshotOptions
	if len(options) == 1 {
		option = options[0]
	}
	var explicit *float64
	if option.Timeout != nil {
		explicit = option.Timeout
	}
	// ariaSnapshot may use LocatorAriaSnapshotOptions
	ret, err := l.frame.channel.SendWithTimeout("ariaSnapshot", resolveTimeout(l.frame.page.timeoutSettings, explicit), option,
		map[string]any{"selector": l.selector})
	if err != nil {
		return "", err
	}
	return ret.(string), nil
}

func (l *locatorImpl) BoundingBox(options ...LocatorBoundingBoxOptions) (*Rect, error) {
	if l.err != nil {
		return nil, l.err
	}
	var option FrameWaitForSelectorOptions
	if len(options) == 1 {
		option.Timeout = options[0].Timeout
	}

	result, err := l.withElement(func(handle ElementHandle, _ *float64) (any, error) {
		return handle.BoundingBox()
	}, option)
	if err != nil {
		return nil, err
	}

	return result.(*Rect), nil
}

func (l *locatorImpl) Check(options ...LocatorCheckOptions) error {
	if l.err != nil {
		return l.err
	}
	opt := FrameCheckOptions{
		Strict: Bool(true),
	}
	if len(options) == 1 {
		if err := assignStructFields(&opt, options[0], false); err != nil {
			return err
		}
	}
	return l.frame.Check(l.selector, opt)
}

func (l *locatorImpl) Clear(options ...LocatorClearOptions) error {
	if l.err != nil {
		return l.err
	}
	if len(options) == 1 {
		return l.Fill("", LocatorFillOptions{
			Force:   options[0].Force,
			Timeout: options[0].Timeout,
		})
	} else {
		return l.Fill("")
	}
}

func (l *locatorImpl) Click(options ...LocatorClickOptions) error {
	if l.err != nil {
		return l.err
	}
	opt := FrameClickOptions{
		Strict: Bool(true),
	}
	if len(options) == 1 {
		if err := assignStructFields(&opt, options[0], false); err != nil {
			return err
		}
	}
	return l.frame.Click(l.selector, opt)
}

func (l *locatorImpl) ContentFrame() FrameLocator {
	return newFrameLocator(l.frame, l.selector)
}

func (l *locatorImpl) Count() (int, error) {
	if l.err != nil {
		return 0, l.err
	}
	return l.frame.queryCount(l.selector)
}

func (l *locatorImpl) Dblclick(options ...LocatorDblclickOptions) error {
	if l.err != nil {
		return l.err
	}
	opt := FrameDblclickOptions{
		Strict: Bool(true),
	}
	if len(options) == 1 {
		if err := assignStructFields(&opt, options[0], false); err != nil {
			return err
		}
	}
	return l.frame.Dblclick(l.selector, opt)
}

func (l *locatorImpl) DispatchEvent(typ string, eventInit any, options ...LocatorDispatchEventOptions) error {
	if l.err != nil {
		return l.err
	}
	opt := FrameDispatchEventOptions{
		Strict: Bool(true),
	}
	if len(options) == 1 {
		if err := assignStructFields(&opt, options[0], false); err != nil {
			return err
		}
	}
	return l.frame.DispatchEvent(l.selector, typ, eventInit, opt)
}

func (l *locatorImpl) DragTo(target Locator, options ...LocatorDragToOptions) error {
	if l.err != nil {
		return l.err
	}
	opt := FrameDragAndDropOptions{
		Strict: Bool(true),
	}
	if len(options) == 1 {
		if err := assignStructFields(&opt, options[0], false); err != nil {
			return err
		}
	}
	return l.frame.DragAndDrop(l.selector, target.(*locatorImpl).selector, opt)
}

func (l *locatorImpl) Drop(payload Payload, options ...LocatorDropOptions) error {
	if l.err != nil {
		return l.err
	}
	params := map[string]any{
		"selector": l.selector,
		"strict":   true,
	}
	if payload.Files != nil {
		converted, err := convertInputFiles(payload.Files, l.frame.page.browserContext)
		if err != nil {
			return err
		}
		if converted.Payloads != nil {
			params["payloads"] = converted.Payloads
		}
		if converted.LocalPaths != nil {
			params["localPaths"] = converted.LocalPaths
		}
		if converted.Streams != nil {
			params["streams"] = converted.Streams
		}
	}
	// The protocol expects data as an array of {mimeType, value} entries.
	if payload.Data != nil {
		data := make([]map[string]string, 0, len(payload.Data))
		for mimeType, value := range payload.Data {
			data = append(data, map[string]string{"mimeType": mimeType, "value": value})
		}
		params["data"] = data
	}
	if len(options) == 1 {
		// params is a map, so assignStructFields (which requires a struct dest)
		// cannot be used here; set the option keys directly.
		if options[0].Position != nil {
			params["position"] = options[0].Position
		}
	}
	var dropTimeout *float64
	if len(options) == 1 {
		dropTimeout = options[0].Timeout
	}
	_, err := l.frame.channel.SendWithTimeout("drop", resolveTimeout(l.frame.page.timeoutSettings, dropTimeout), params)
	return err
}

func (l *locatorImpl) ElementHandle(options ...LocatorElementHandleOptions) (ElementHandle, error) {
	if l.err != nil {
		return nil, l.err
	}
	option := FrameWaitForSelectorOptions{
		State:  WaitForSelectorStateAttached,
		Strict: Bool(true),
	}
	if len(options) == 1 {
		if err := assignStructFields(&option, options[0], false); err != nil {
			return nil, err
		}
	}
	return l.frame.WaitForSelector(l.selector, option)
}

func (l *locatorImpl) ElementHandles() ([]ElementHandle, error) {
	if l.err != nil {
		return nil, l.err
	}
	return l.frame.QuerySelectorAll(l.selector)
}

func (l *locatorImpl) Evaluate(expression string, arg any, options ...LocatorEvaluateOptions) (any, error) {
	if l.err != nil {
		return nil, l.err
	}
	var option FrameWaitForSelectorOptions
	if len(options) == 1 {
		option.Timeout = options[0].Timeout
	}

	return l.withElement(func(handle ElementHandle, _ *float64) (any, error) {
		return handle.Evaluate(expression, arg)
	}, option)
}

func (l *locatorImpl) EvaluateAll(expression string, options ...any) (any, error) {
	if l.err != nil {
		return nil, l.err
	}
	return l.frame.EvalOnSelectorAll(l.selector, expression, options...)
}

func (l *locatorImpl) EvaluateHandle(expression string, arg any, options ...LocatorEvaluateHandleOptions) (JSHandle, error) {
	if l.err != nil {
		return nil, l.err
	}
	var option FrameWaitForSelectorOptions
	if len(options) == 1 {
		option.Timeout = options[0].Timeout
	}

	h, err := l.withElement(func(handle ElementHandle, _ *float64) (any, error) {
		return handle.EvaluateHandle(expression, arg)
	}, option)
	if err != nil {
		return nil, err
	}
	return h.(JSHandle), nil
}

func (l *locatorImpl) Fill(value string, options ...LocatorFillOptions) error {
	if l.err != nil {
		return l.err
	}
	opt := FrameFillOptions{
		Strict: Bool(true),
	}
	if len(options) == 1 {
		if err := assignStructFields(&opt, options[0], false); err != nil {
			return err
		}
	}
	return l.frame.Fill(l.selector, value, opt)
}

func (l *locatorImpl) Filter(options ...LocatorFilterOptions) Locator {
	if len(options) == 1 {
		return newLocator(l.frame, l.selector, LocatorOptions(options[0]))
	}
	return newLocator(l.frame, l.selector)
}

func (l *locatorImpl) First() Locator {
	return newLocator(l.frame, l.selector+" >> nth=0")
}

func (l *locatorImpl) Focus(options ...LocatorFocusOptions) error {
	if l.err != nil {
		return l.err
	}
	opt := FrameFocusOptions{
		Strict: Bool(true),
	}
	if len(options) == 1 {
		if err := assignStructFields(&opt, options[0], false); err != nil {
			return err
		}
	}
	return l.frame.Focus(l.selector, opt)
}

func (l *locatorImpl) FrameLocator(selector string) FrameLocator {
	return newFrameLocator(l.frame, l.selector+" >> "+selector)
}

func (l *locatorImpl) GetAttribute(name string, options ...LocatorGetAttributeOptions) (string, error) {
	if l.err != nil {
		return "", l.err
	}
	opt := FrameGetAttributeOptions{
		Strict: Bool(true),
	}
	if len(options) == 1 {
		if err := assignStructFields(&opt, options[0], false); err != nil {
			return "", err
		}
	}
	return l.frame.GetAttribute(l.selector, name, opt)
}

func (l *locatorImpl) GetByAltText(text any, options ...LocatorGetByAltTextOptions) Locator {
	exact := false
	if len(options) == 1 {
		if options[0].Exact != nil && *options[0].Exact {
			exact = true
		}
	}
	selector, err := getByAltTextSelector(text, exact)
	if err != nil {
		return l.withError(err)
	}
	return l.Locator(selector)
}

func (l *locatorImpl) GetByLabel(text any, options ...LocatorGetByLabelOptions) Locator {
	exact := false
	if len(options) == 1 {
		if options[0].Exact != nil && *options[0].Exact {
			exact = true
		}
	}
	selector, err := getByLabelSelector(text, exact)
	if err != nil {
		return l.withError(err)
	}
	return l.Locator(selector)
}

func (l *locatorImpl) GetByPlaceholder(text any, options ...LocatorGetByPlaceholderOptions) Locator {
	exact := false
	if len(options) == 1 {
		if options[0].Exact != nil && *options[0].Exact {
			exact = true
		}
	}
	selector, err := getByPlaceholderSelector(text, exact)
	if err != nil {
		return l.withError(err)
	}
	return l.Locator(selector)
}

func (l *locatorImpl) GetByRole(role AriaRole, options ...LocatorGetByRoleOptions) Locator {
	selector, err := getByRoleSelector(role, options...)
	if err != nil {
		return l.withError(err)
	}
	return l.Locator(selector)
}

func (l *locatorImpl) GetByTestId(testId any) Locator {
	selector, err := getByTestIdSelector(getTestIdAttributeName(), testId)
	if err != nil {
		return l.withError(err)
	}
	return l.Locator(selector)
}

func (l *locatorImpl) GetByText(text any, options ...LocatorGetByTextOptions) Locator {
	exact := false
	if len(options) == 1 {
		if options[0].Exact != nil && *options[0].Exact {
			exact = true
		}
	}
	selector, err := getByTextSelector(text, exact)
	if err != nil {
		return l.withError(err)
	}
	return l.Locator(selector)
}

func (l *locatorImpl) GetByTitle(text any, options ...LocatorGetByTitleOptions) Locator {
	exact := false
	if len(options) == 1 {
		if options[0].Exact != nil && *options[0].Exact {
			exact = true
		}
	}
	selector, err := getByTitleSelector(text, exact)
	if err != nil {
		return l.withError(err)
	}
	return l.Locator(selector)
}

func (l *locatorImpl) HideHighlight() error {
	if l.err != nil {
		return l.err
	}
	_, err := l.frame.channel.Send("hideHighlight", map[string]any{
		"selector": l.selector,
	})
	return err
}

func (l *locatorImpl) Highlight() error {
	if l.err != nil {
		return l.err
	}
	return l.frame.highlight(l.selector)
}

func (l *locatorImpl) Hover(options ...LocatorHoverOptions) error {
	if l.err != nil {
		return l.err
	}
	opt := FrameHoverOptions{
		Strict: Bool(true),
	}
	if len(options) == 1 {
		if err := assignStructFields(&opt, options[0], false); err != nil {
			return err
		}
	}
	return l.frame.Hover(l.selector, opt)
}

func (l *locatorImpl) InnerHTML(options ...LocatorInnerHTMLOptions) (string, error) {
	if l.err != nil {
		return "", l.err
	}
	opt := FrameInnerHTMLOptions{
		Strict: Bool(true),
	}
	if len(options) == 1 {
		if err := assignStructFields(&opt, options[0], false); err != nil {
			return "", err
		}
	}
	return l.frame.InnerHTML(l.selector, opt)
}

func (l *locatorImpl) InnerText(options ...LocatorInnerTextOptions) (string, error) {
	if l.err != nil {
		return "", l.err
	}
	opt := FrameInnerTextOptions{
		Strict: Bool(true),
	}
	if len(options) == 1 {
		if err := assignStructFields(&opt, options[0], false); err != nil {
			return "", err
		}
	}
	return l.frame.InnerText(l.selector, opt)
}

func (l *locatorImpl) InputValue(options ...LocatorInputValueOptions) (string, error) {
	if l.err != nil {
		return "", l.err
	}
	opt := FrameInputValueOptions{
		Strict: Bool(true),
	}
	if len(options) == 1 {
		if err := assignStructFields(&opt, options[0], false); err != nil {
			return "", err
		}
	}
	return l.frame.InputValue(l.selector, opt)
}

func (l *locatorImpl) IsChecked(options ...LocatorIsCheckedOptions) (bool, error) {
	if l.err != nil {
		return false, l.err
	}
	opt := FrameIsCheckedOptions{
		Strict: Bool(true),
	}
	if len(options) == 1 {
		if err := assignStructFields(&opt, options[0], false); err != nil {
			return false, err
		}
	}
	return l.frame.IsChecked(l.selector, opt)
}

func (l *locatorImpl) IsDisabled(options ...LocatorIsDisabledOptions) (bool, error) {
	if l.err != nil {
		return false, l.err
	}
	opt := FrameIsDisabledOptions{
		Strict: Bool(true),
	}
	if len(options) == 1 {
		if err := assignStructFields(&opt, options[0], false); err != nil {
			return false, err
		}
	}
	return l.frame.IsDisabled(l.selector, opt)
}

func (l *locatorImpl) IsEditable(options ...LocatorIsEditableOptions) (bool, error) {
	if l.err != nil {
		return false, l.err
	}
	opt := FrameIsEditableOptions{
		Strict: Bool(true),
	}
	if len(options) == 1 {
		if err := assignStructFields(&opt, options[0], false); err != nil {
			return false, err
		}
	}
	return l.frame.IsEditable(l.selector, opt)
}

func (l *locatorImpl) IsEnabled(options ...LocatorIsEnabledOptions) (bool, error) {
	if l.err != nil {
		return false, l.err
	}
	opt := FrameIsEnabledOptions{
		Strict: Bool(true),
	}
	if len(options) == 1 {
		if err := assignStructFields(&opt, options[0], false); err != nil {
			return false, err
		}
	}
	return l.frame.IsEnabled(l.selector, opt)
}

func (l *locatorImpl) IsHidden(options ...LocatorIsHiddenOptions) (bool, error) {
	if l.err != nil {
		return false, l.err
	}
	opt := FrameIsHiddenOptions{
		Strict: Bool(true),
	}
	if len(options) == 1 {
		if err := assignStructFields(&opt, options[0], false); err != nil {
			return false, err
		}
	}
	return l.frame.IsHidden(l.selector, opt)
}

func (l *locatorImpl) IsVisible(options ...LocatorIsVisibleOptions) (bool, error) {
	if l.err != nil {
		return false, l.err
	}
	opt := FrameIsVisibleOptions{
		Strict: Bool(true),
	}
	if len(options) == 1 {
		if err := assignStructFields(&opt, options[0], false); err != nil {
			return false, err
		}
	}
	return l.frame.IsVisible(l.selector, opt)
}

func (l *locatorImpl) Last() Locator {
	return newLocator(l.frame, l.selector+" >> nth=-1")
}

func (l *locatorImpl) Locator(selectorOrLocator any, options ...LocatorLocatorOptions) Locator {
	var option LocatorOptions
	if len(options) == 1 {
		option = LocatorOptions{
			Has:        options[0].Has,
			HasNot:     options[0].HasNot,
			HasText:    options[0].HasText,
			HasNotText: options[0].HasNotText,
		}
	}

	selector, ok := selectorOrLocator.(string)
	if ok {
		return newLocator(l.frame, l.selector+" >> "+selector, option)
	}
	locator, ok := selectorOrLocator.(*locatorImpl)
	if ok {
		if l.frame != locator.frame {
			return l.withError(ErrLocatorNotSameFrame)
		}
		return newLocator(
			l.frame,
			l.selector+" >> internal:chain="+escapeText(locator.selector),
			option,
		)
	}
	return l.withError(fmt.Errorf("invalid locator parameter: %v", selectorOrLocator))
}

func (l *locatorImpl) Nth(index int) Locator {
	return newLocator(l.frame, l.selector+" >> nth="+strconv.Itoa(index))
}

func (l *locatorImpl) Page() (Page, error) {
	if l.err != nil {
		return nil, l.err
	}
	return l.frame.Page(), nil
}

func (l *locatorImpl) Press(key string, options ...LocatorPressOptions) error {
	if l.err != nil {
		return l.err
	}
	opt := FramePressOptions{
		Strict: Bool(true),
	}
	if len(options) == 1 {
		if err := assignStructFields(&opt, options[0], false); err != nil {
			return err
		}
	}
	return l.frame.Press(l.selector, key, opt)
}

func (l *locatorImpl) PressSequentially(text string, options ...LocatorPressSequentiallyOptions) error {
	if l.err != nil {
		return l.err
	}
	var option LocatorTypeOptions
	if len(options) == 1 {
		option = LocatorTypeOptions(options[0])
	}
	return l.Type(text, option)
}

func (l *locatorImpl) Screenshot(options ...LocatorScreenshotOptions) ([]byte, error) {
	if l.err != nil {
		return nil, l.err
	}
	var option FrameWaitForSelectorOptions
	if len(options) == 1 {
		option.Timeout = options[0].Timeout
	}

	result, err := l.withElement(func(handle ElementHandle, timeout *float64) (any, error) {
		var screenshotOption ElementHandleScreenshotOptions
		if len(options) == 1 {
			screenshotOption = ElementHandleScreenshotOptions(options[0])
		}
		screenshotOption.Timeout = timeout
		return handle.Screenshot(screenshotOption)
	}, option)
	if err != nil {
		return nil, err
	}

	return result.([]byte), nil
}

func (l *locatorImpl) ScrollIntoViewIfNeeded(options ...LocatorScrollIntoViewIfNeededOptions) error {
	if l.err != nil {
		return l.err
	}
	var option FrameWaitForSelectorOptions
	if len(options) == 1 {
		option.Timeout = options[0].Timeout
	}

	_, err := l.withElement(func(handle ElementHandle, timeout *float64) (any, error) {
		var opt ElementHandleScrollIntoViewIfNeededOptions
		opt.Timeout = timeout
		return nil, handle.ScrollIntoViewIfNeeded(opt)
	}, option)

	return err
}

func (l *locatorImpl) SelectOption(values SelectOptionValues, options ...LocatorSelectOptionOptions) ([]string, error) {
	if l.err != nil {
		return nil, l.err
	}
	opt := FrameSelectOptionOptions{
		Strict: Bool(true),
	}
	if len(options) == 1 {
		if err := assignStructFields(&opt, options[0], false); err != nil {
			return nil, err
		}
	}
	return l.frame.SelectOption(l.selector, values, opt)
}

func (l *locatorImpl) SelectText(options ...LocatorSelectTextOptions) error {
	if l.err != nil {
		return l.err
	}
	var option FrameWaitForSelectorOptions
	if len(options) == 1 {
		option.Timeout = options[0].Timeout
	}

	_, err := l.withElement(func(handle ElementHandle, timeout *float64) (any, error) {
		var opt ElementHandleSelectTextOptions
		if len(options) == 1 {
			opt = ElementHandleSelectTextOptions(options[0])
		}
		opt.Timeout = timeout
		return nil, handle.SelectText(opt)
	}, option)

	return err
}

func (l *locatorImpl) SetChecked(checked bool, options ...LocatorSetCheckedOptions) error {
	if l.err != nil {
		return l.err
	}
	opt := FrameSetCheckedOptions{
		Strict: Bool(true),
	}
	if len(options) == 1 {
		if err := assignStructFields(&opt, options[0], false); err != nil {
			return err
		}
	}
	return l.frame.SetChecked(l.selector, checked, opt)
}

func (l *locatorImpl) SetInputFiles(files any, options ...LocatorSetInputFilesOptions) error {
	if l.err != nil {
		return l.err
	}
	opt := FrameSetInputFilesOptions{
		Strict: Bool(true),
	}
	if len(options) == 1 {
		if err := assignStructFields(&opt, options[0], false); err != nil {
			return err
		}
	}
	return l.frame.SetInputFiles(l.selector, files, opt)
}

func (l *locatorImpl) Tap(options ...LocatorTapOptions) error {
	if l.err != nil {
		return l.err
	}
	opt := FrameTapOptions{
		Strict: Bool(true),
	}
	if len(options) == 1 {
		if err := assignStructFields(&opt, options[0], false); err != nil {
			return err
		}
	}
	return l.frame.Tap(l.selector, opt)
}

func (l *locatorImpl) TextContent(options ...LocatorTextContentOptions) (string, error) {
	if l.err != nil {
		return "", l.err
	}
	opt := FrameTextContentOptions{
		Strict: Bool(true),
	}
	if len(options) == 1 {
		if err := assignStructFields(&opt, options[0], false); err != nil {
			return "", err
		}
	}
	return l.frame.TextContent(l.selector, opt)
}

func (l *locatorImpl) Type(text string, options ...LocatorTypeOptions) error {
	if l.err != nil {
		return l.err
	}
	opt := FrameTypeOptions{
		Strict: Bool(true),
	}
	if len(options) == 1 {
		if err := assignStructFields(&opt, options[0], false); err != nil {
			return err
		}
	}
	return l.frame.Type(l.selector, text, opt)
}

func (l *locatorImpl) Uncheck(options ...LocatorUncheckOptions) error {
	if l.err != nil {
		return l.err
	}
	opt := FrameUncheckOptions{
		Strict: Bool(true),
	}
	if len(options) == 1 {
		if err := assignStructFields(&opt, options[0], false); err != nil {
			return err
		}
	}
	return l.frame.Uncheck(l.selector, opt)
}

func (l *locatorImpl) WaitFor(options ...LocatorWaitForOptions) error {
	if l.err != nil {
		return l.err
	}
	opt := FrameWaitForSelectorOptions{
		Strict: Bool(true),
	}
	if len(options) == 1 {
		if err := assignStructFields(&opt, options[0], false); err != nil {
			return err
		}
	}
	_, err := l.frame.WaitForSelector(l.selector, opt)
	return err
}

func (l *locatorImpl) withElement(
	callback func(handle ElementHandle, timeout *float64) (any, error),
	options ...FrameWaitForSelectorOptions,
) (any, error) {
	if l.err != nil {
		return nil, l.err
	}
	option := FrameWaitForSelectorOptions{
		State:  WaitForSelectorStateAttached,
		Strict: Bool(true),
	}
	if len(options) == 1 {
		option.Timeout = options[0].Timeout
	}
	// Mirror upstream `_withElement`: resolve the effective timeout once and use
	// one budget for waitForSelector plus the inner action. A zero timeout means
	// unlimited and must remain zero for both stages.
	option.Timeout = resolveTimeout(l.frame.page.timeoutSettings, option.Timeout)
	var deadline *time.Time
	if *option.Timeout > 0 {
		d := time.Now().Add(time.Duration(*option.Timeout) * time.Millisecond)
		deadline = &d
	}
	handle, err := l.frame.WaitForSelector(l.selector, option)
	if err != nil {
		return nil, err
	}

	remaining := option.Timeout
	if deadline != nil {
		ms := float64(time.Until(*deadline).Milliseconds())
		// Floor at 1ms, not 0: the protocol treats timeout 0 as "disable timeout"
		// (infinite), so an exhausted budget must still fail fast rather than hang.
		if ms <= 0 {
			ms = 1
		}
		remaining = Float(ms)
	}
	result, err := callback(handle, remaining)
	if err != nil {
		_ = handle.Dispose()
		return nil, err
	}
	if err := handle.Dispose(); err != nil {
		return nil, err
	}
	return result, nil
}

func (l *locatorImpl) WaitForFunction(expression string, arg any, options ...LocatorWaitForFunctionOptions) error {
	if l.err != nil {
		return l.err
	}
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	params := map[string]any{
		"selector":   l.selector,
		"strict":     true,
		"expression": expression,
		"arg":        serializeArgument(arg),
	}
	_, err := l.frame.channel.SendWithTimeout(
		"waitForFunction",
		resolveTimeout(l.frame.page.timeoutSettings, explicit),
		params,
	)
	return err
}

func (l *locatorImpl) expect(expression string, options frameExpectOptions) (*frameExpectResult, error) {
	if l.err != nil {
		return nil, l.err
	}
	overrides := map[string]any{
		"selector":   l.selector,
		"expression": expression,
	}
	if options.ExpectedValue != nil {
		overrides["expectedValue"] = serializeArgument(options.ExpectedValue)
		options.ExpectedValue = nil
	}
	_, err := l.frame.channel.SendReturnAsDictWithTimeout("expect", options.Timeout, options, overrides)
	if err != nil {
		// Since v1.61 a failed assertion is reported as a server error carrying
		// structured errorDetails rather than a `{ matches: false }` result.
		// Unwrap it into the negative result; any other error propagates.
		var detailed *errorWithDetails
		if errors.As(err, &detailed) {
			result := &frameExpectResult{
				Matches:  options.IsNot,
				Received: parseExpectReceived(detailed.details["received"]),
				Log:      detailed.log,
			}
			if v, ok := detailed.details["timedOut"].(bool); ok {
				result.TimedOut = &v
			}
			return result, nil
		}
		return nil, err
	}
	// No error means the assertion matched.
	return &frameExpectResult{Matches: !options.IsNot}, nil
}

func (l *locatorImpl) Normalize() Locator {
	result, err := l.frame.channel.Send("resolveSelector", map[string]any{
		"selector": l.selector,
	})
	if err != nil {
		return l
	}
	if selector, ok := result.(string); ok {
		return newLocator(l.frame, selector)
	}
	return l
}
