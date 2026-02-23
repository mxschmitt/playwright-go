package playwright

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	testIdAttributeName    = "data-testid"
	ErrLocatorNotSameFrame = errors.New("inner 'has' or 'hasNot' locator must belong to the same frame")
)

type locatorImpl struct {
	frame       *frameImpl
	selector    string
	options     *LocatorOptions
	err         error
	description *string
}

type LocatorOptions LocatorFilterOptions

func newLocator(frame *frameImpl, selector string, options ...LocatorOptions) *locatorImpl {
	option := &LocatorOptions{}
	if len(options) == 1 {
		option = &options[0]
	}
	locator := &locatorImpl{frame: frame, selector: selector, options: option, err: nil}
	if option.HasText != nil {
		selector += fmt.Sprintf(` >> internal:has-text=%s`, escapeForTextSelector(option.HasText, false))
	}
	if option.HasNotText != nil {
		selector += fmt.Sprintf(` >> internal:has-not-text=%s`, escapeForTextSelector(option.HasNotText, false))
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

func (l *locatorImpl) equals(locator Locator) bool {
	return l.frame == locator.(*locatorImpl).frame && l.err == locator.(*locatorImpl).err && l.selector == locator.(*locatorImpl).selector
}

func (l *locatorImpl) Err() error {
	return l.err
}

func (l *locatorImpl) Describe(description string) Locator {
	return &locatorImpl{
		frame:       l.frame,
		selector:    l.selector,
		options:     l.options,
		err:         l.err,
		description: &description,
	}
}

func (l *locatorImpl) Description() (string, error) {
	if l.description == nil {
		return "", nil
	}
	return *l.description, nil
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
	return newLocator(l.frame, l.selector+` >> internal:and=`+escapeText(locator.(*locatorImpl).selector))
}

func (l *locatorImpl) Or(locator Locator) Locator {
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
	if len(options) == 1 && options[0].Timeout != nil {
		params["timeout"] = options[0].Timeout
	} else {
		params["timeout"] = float64(30000) // default 30s, required in Playwright v1.57+
	}
	_, err := l.frame.channel.Send("blur", params)
	return err
}

func (l *locatorImpl) AriaSnapshot(options ...LocatorAriaSnapshotOptions) (string, error) {
	var option LocatorAriaSnapshotOptions
	if len(options) == 1 {
		option = options[0]
	}
	ret, err := l.frame.channel.Send("ariaSnapshot", option,
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

	result, err := l.withElement(func(handle ElementHandle) (any, error) {
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

	return l.withElement(func(handle ElementHandle) (any, error) {
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

	h, err := l.withElement(func(handle ElementHandle) (any, error) {
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
		if *options[0].Exact {
			exact = true
		}
	}
	return l.Locator(getByAltTextSelector(text, exact))
}

func (l *locatorImpl) GetByLabel(text any, options ...LocatorGetByLabelOptions) Locator {
	exact := false
	if len(options) == 1 {
		if *options[0].Exact {
			exact = true
		}
	}
	return l.Locator(getByLabelSelector(text, exact))
}

func (l *locatorImpl) GetByPlaceholder(text any, options ...LocatorGetByPlaceholderOptions) Locator {
	exact := false
	if len(options) == 1 {
		if *options[0].Exact {
			exact = true
		}
	}
	return l.Locator(getByPlaceholderSelector(text, exact))
}

func (l *locatorImpl) GetByRole(role AriaRole, options ...LocatorGetByRoleOptions) Locator {
	return l.Locator(getByRoleSelector(role, options...))
}

func (l *locatorImpl) GetByTestId(testId any) Locator {
	return l.Locator(getByTestIdSelector(getTestIdAttributeName(), testId))
}

func (l *locatorImpl) GetByText(text any, options ...LocatorGetByTextOptions) Locator {
	exact := false
	if len(options) == 1 {
		if *options[0].Exact {
			exact = true
		}
	}
	return l.Locator(getByTextSelector(text, exact))
}

func (l *locatorImpl) GetByTitle(text any, options ...LocatorGetByTitleOptions) Locator {
	exact := false
	if len(options) == 1 {
		if *options[0].Exact {
			exact = true
		}
	}
	return l.Locator(getByTitleSelector(text, exact))
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
			l.err = errors.Join(l.err, ErrLocatorNotSameFrame)
			return l
		}
		return newLocator(l.frame,
			l.selector+" >> internal:chain="+escapeText(locator.selector),
			option,
		)
	}
	l.err = errors.Join(l.err, fmt.Errorf("invalid locator parameter: %v", selectorOrLocator))
	return l
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

	result, err := l.withElement(func(handle ElementHandle) (any, error) {
		var screenshotOption ElementHandleScreenshotOptions
		if len(options) == 1 {
			screenshotOption = ElementHandleScreenshotOptions(options[0])
		}
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

	_, err := l.withElement(func(handle ElementHandle) (any, error) {
		var opt ElementHandleScrollIntoViewIfNeededOptions
		if len(options) == 1 {
			opt.Timeout = options[0].Timeout
		}
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

	_, err := l.withElement(func(handle ElementHandle) (any, error) {
		var opt ElementHandleSelectTextOptions
		if len(options) == 1 {
			opt = ElementHandleSelectTextOptions(options[0])
		}
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
	callback func(handle ElementHandle) (any, error),
	options ...FrameWaitForSelectorOptions,
) (any, error) {
	if l.err != nil {
		return nil, l.err
	}
	handle, err := l.frame.WaitForSelector(l.selector, options...)
	if err != nil {
		return nil, err
	}

	result, err := callback(handle)
	if err != nil {
		return nil, err
	}
	return result, nil
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
	result, err := l.frame.channel.SendReturnAsDict("expect", options, overrides)
	if err != nil {
		return nil, err
	}
	var (
		received any
		matches  bool
		log      []string
	)

	if v, ok := result["received"]; ok {
		received = parseResult(v)
	}
	if v, ok := result["matches"]; ok {
		matches = v.(bool)
	}
	if v, ok := result["log"]; ok {
		for _, l := range v.([]any) {
			log = append(log, l.(string))
		}
	}
	return &frameExpectResult{Received: received, Matches: matches, Log: log}, nil
}
