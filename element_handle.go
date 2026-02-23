package playwright

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
)

type elementHandleImpl struct {
	jsHandleImpl
}

func (e *elementHandleImpl) AsElement() ElementHandle {
	return e
}

func (e *elementHandleImpl) OwnerFrame() (Frame, error) {
	channel, err := e.channel.Send("ownerFrame")
	if err != nil {
		return nil, err
	}
	channelOwner := fromNullableChannel(channel)
	if channelOwner == nil {
		return nil, nil
	}
	return channelOwner.(*frameImpl), nil
}

func (e *elementHandleImpl) ContentFrame() (Frame, error) {
	channel, err := e.channel.Send("contentFrame")
	if err != nil {
		return nil, err
	}
	channelOwner := fromNullableChannel(channel)
	if channelOwner == nil {
		return nil, nil
	}
	return channelOwner.(*frameImpl), nil
}

func (e *elementHandleImpl) GetAttribute(name string) (string, error) {
	attribute, err := e.channel.Send("getAttribute", map[string]any{
		"name": name,
	})
	if attribute == nil {
		return "", err
	}
	return attribute.(string), err
}

func (e *elementHandleImpl) TextContent() (string, error) {
	textContent, err := e.channel.Send("textContent")
	if textContent == nil {
		return "", err
	}
	return textContent.(string), err
}

func (e *elementHandleImpl) InnerText() (string, error) {
	innerText, err := e.channel.Send("innerText")
	if innerText == nil {
		return "", err
	}
	return innerText.(string), err
}

func (e *elementHandleImpl) InnerHTML() (string, error) {
	innerHTML, err := e.channel.Send("innerHTML")
	if innerHTML == nil {
		return "", err
	}
	return innerHTML.(string), err
}

func (e *elementHandleImpl) DispatchEvent(typ string, initObjects ...any) error {
	var initObject any
	if len(initObjects) == 1 {
		initObject = initObjects[0]
	}
	_, err := e.channel.Send("dispatchEvent", map[string]any{
		"type":      typ,
		"eventInit": serializeArgument(initObject),
	})
	return err
}

func (e *elementHandleImpl) Hover(options ...ElementHandleHoverOptions) error {
	_, err := e.channel.Send("hover", options)
	return err
}

func (e *elementHandleImpl) Click(options ...ElementHandleClickOptions) error {
	_, err := e.channel.Send("click", options)
	return err
}

func (e *elementHandleImpl) Dblclick(options ...ElementHandleDblclickOptions) error {
	_, err := e.channel.Send("dblclick", options)
	return err
}

func (e *elementHandleImpl) QuerySelector(selector string) (ElementHandle, error) {
	channel, err := e.channel.Send("querySelector", map[string]any{
		"selector": selector,
	})
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, nil
	}
	return fromChannel(channel).(*elementHandleImpl), nil
}

func (e *elementHandleImpl) QuerySelectorAll(selector string) ([]ElementHandle, error) {
	channels, err := e.channel.Send("querySelectorAll", map[string]any{
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

func (e *elementHandleImpl) EvalOnSelector(selector string, expression string, options ...any) (any, error) {
	var arg any
	if len(options) == 1 {
		arg = options[0]
	}
	result, err := e.channel.Send("evalOnSelector", map[string]any{
		"selector":   selector,
		"expression": expression,
		"arg":        serializeArgument(arg),
	})
	if err != nil {
		return nil, err
	}
	return parseResult(result), nil
}

func (e *elementHandleImpl) EvalOnSelectorAll(selector string, expression string, options ...any) (any, error) {
	var arg any
	if len(options) == 1 {
		arg = options[0]
	}
	result, err := e.channel.Send("evalOnSelectorAll", map[string]any{
		"selector":   selector,
		"expression": expression,
		"arg":        serializeArgument(arg),
	})
	if err != nil {
		return nil, err
	}
	return parseResult(result), nil
}

func (e *elementHandleImpl) ScrollIntoViewIfNeeded(options ...ElementHandleScrollIntoViewIfNeededOptions) error {
	_, err := e.channel.Send("scrollIntoViewIfNeeded", options)
	if err != nil {
		return err
	}
	return err
}

func (e *elementHandleImpl) SetInputFiles(files any, options ...ElementHandleSetInputFilesOptions) error {
	frame, err := e.OwnerFrame()
	if err != nil {
		return err
	}
	if frame == nil {
		return errors.New("Cannot set input files to detached element")
	}

	params, err := convertInputFiles(files, frame.(*frameImpl).page.browserContext)
	if err != nil {
		return err
	}
	_, err = e.channel.Send("setInputFiles", params, options)
	return err
}

func (e *elementHandleImpl) BoundingBox() (*Rect, error) {
	boundingBox, err := e.channel.Send("boundingBox")
	if err != nil {
		return nil, err
	}

	if boundingBox == nil {
		return nil, nil
	}

	out := &Rect{}
	remapMapToStruct(boundingBox, out)
	return out, nil
}

func (e *elementHandleImpl) Check(options ...ElementHandleCheckOptions) error {
	_, err := e.channel.Send("check", options)
	return err
}

func (e *elementHandleImpl) Uncheck(options ...ElementHandleUncheckOptions) error {
	_, err := e.channel.Send("uncheck", options)
	return err
}

func (e *elementHandleImpl) Press(key string, options ...ElementHandlePressOptions) error {
	_, err := e.channel.Send("press", map[string]any{
		"key": key,
	}, options)
	return err
}

func (e *elementHandleImpl) Fill(value string, options ...ElementHandleFillOptions) error {
	_, err := e.channel.Send("fill", map[string]any{
		"value": value,
	}, options)
	return err
}

func (e *elementHandleImpl) Type(value string, options ...ElementHandleTypeOptions) error {
	_, err := e.channel.Send("type", map[string]any{
		"text": value,
	}, options)
	return err
}

func (e *elementHandleImpl) Focus() error {
	_, err := e.channel.Send("focus")
	return err
}

func (e *elementHandleImpl) SelectText(options ...ElementHandleSelectTextOptions) error {
	_, err := e.channel.Send("selectText", options)
	return err
}

func (e *elementHandleImpl) Screenshot(options ...ElementHandleScreenshotOptions) ([]byte, error) {
	var path *string
	overrides := map[string]any{}
	if len(options) == 1 {
		path = options[0].Path
		options[0].Path = nil
		if options[0].Mask != nil {
			masks := make([]map[string]any, 0)
			for _, m := range options[0].Mask {
				if m.Err() != nil { // ErrLocatorNotSameFrame
					return nil, m.Err()
				}
				l, ok := m.(*locatorImpl)
				if ok {
					masks = append(masks, map[string]any{
						"selector": l.selector,
						"frame":    l.frame.channel,
					})
				}
			}
			overrides["mask"] = masks
			options[0].Mask = nil
		}
	}
	data, err := e.channel.Send("screenshot", options, overrides)
	if err != nil {
		return nil, err
	}
	image, err := base64.StdEncoding.DecodeString(data.(string))
	if err != nil {
		return nil, fmt.Errorf("could not decode base64 :%w", err)
	}
	if path != nil {
		if err := os.WriteFile(*path, image, 0o644); err != nil {
			return nil, err
		}
	}
	return image, nil
}

func (e *elementHandleImpl) Tap(options ...ElementHandleTapOptions) error {
	_, err := e.channel.Send("tap", options)
	return err
}

func (e *elementHandleImpl) SelectOption(values SelectOptionValues, options ...ElementHandleSelectOptionOptions) ([]string, error) {
	opts := convertSelectOptionSet(values)
	selected, err := e.channel.Send("selectOption", opts, options)
	if err != nil {
		return nil, err
	}

	return transformToStringList(selected), nil
}

func (e *elementHandleImpl) IsChecked() (bool, error) {
	checked, err := e.channel.Send("isChecked")
	if err != nil {
		return false, err
	}
	return checked.(bool), nil
}

func (e *elementHandleImpl) IsDisabled() (bool, error) {
	disabled, err := e.channel.Send("isDisabled")
	if err != nil {
		return false, err
	}
	return disabled.(bool), nil
}

func (e *elementHandleImpl) IsEditable() (bool, error) {
	editable, err := e.channel.Send("isEditable")
	if err != nil {
		return false, err
	}
	return editable.(bool), nil
}

func (e *elementHandleImpl) IsEnabled() (bool, error) {
	enabled, err := e.channel.Send("isEnabled")
	if err != nil {
		return false, err
	}
	return enabled.(bool), nil
}

func (e *elementHandleImpl) IsHidden() (bool, error) {
	hidden, err := e.channel.Send("isHidden")
	if err != nil {
		return false, err
	}
	return hidden.(bool), nil
}

func (e *elementHandleImpl) IsVisible() (bool, error) {
	visible, err := e.channel.Send("isVisible")
	if err != nil {
		return false, err
	}
	return visible.(bool), nil
}

func (e *elementHandleImpl) WaitForElementState(state ElementState, options ...ElementHandleWaitForElementStateOptions) error {
	_, err := e.channel.Send("waitForElementState", map[string]any{
		"state": state,
	}, options)
	if err != nil {
		return err
	}
	return nil
}

func (e *elementHandleImpl) WaitForSelector(selector string, options ...ElementHandleWaitForSelectorOptions) (ElementHandle, error) {
	ch, err := e.channel.Send("waitForSelector", map[string]any{
		"selector": selector,
	}, options)
	if err != nil {
		return nil, err
	}

	channelOwner := fromNullableChannel(ch)
	if channelOwner == nil {
		return nil, nil
	}
	return channelOwner.(*elementHandleImpl), nil
}

func (e *elementHandleImpl) InputValue(options ...ElementHandleInputValueOptions) (string, error) {
	result, err := e.channel.Send("inputValue", options)
	if result == nil {
		return "", err
	}
	return result.(string), err
}

func (e *elementHandleImpl) SetChecked(checked bool, options ...ElementHandleSetCheckedOptions) error {
	if checked {
		_, err := e.channel.Send("check", options)
		return err
	} else {
		_, err := e.channel.Send("uncheck", options)
		return err
	}
}

func newElementHandle(parent *channelOwner, objectType string, guid string, initializer map[string]any) *elementHandleImpl {
	bt := &elementHandleImpl{}
	bt.createChannelOwner(bt, parent, objectType, guid, initializer)
	return bt
}

func transformToStringList(in any) []string {
	s := in.([]any)

	var out []string
	for _, v := range s {
		out = append(out, v.(string))
	}
	return out
}
