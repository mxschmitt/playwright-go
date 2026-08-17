package playwright

import (
	"fmt"
	"strconv"
)

type frameLocatorImpl struct {
	frame         *frameImpl
	frameSelector string
}

func newFrameLocator(frame *frameImpl, frameSelector string) *frameLocatorImpl {
	return &frameLocatorImpl{frame: frame, frameSelector: frameSelector}
}

func (fl *frameLocatorImpl) First() FrameLocator {
	return newFrameLocator(fl.frame, fl.frameSelector+" >> nth=0")
}

func (fl *frameLocatorImpl) FrameLocator(selector string) FrameLocator {
	return newFrameLocator(fl.frame, fl.frameSelector+" >> internal:control=enter-frame >> "+selector)
}

func (fl *frameLocatorImpl) GetByAltText(text any, options ...FrameLocatorGetByAltTextOptions) Locator {
	exact := false
	if len(options) == 1 {
		if options[0].Exact != nil && *options[0].Exact {
			exact = true
		}
	}
	selector, err := getByAltTextSelector(text, exact)
	if err != nil {
		return newErrorLocator(fl.frame, err)
	}
	return fl.Locator(selector)
}

func (fl *frameLocatorImpl) GetByLabel(text any, options ...FrameLocatorGetByLabelOptions) Locator {
	exact := false
	if len(options) == 1 {
		if options[0].Exact != nil && *options[0].Exact {
			exact = true
		}
	}
	selector, err := getByLabelSelector(text, exact)
	if err != nil {
		return newErrorLocator(fl.frame, err)
	}
	return fl.Locator(selector)
}

func (fl *frameLocatorImpl) GetByPlaceholder(text any, options ...FrameLocatorGetByPlaceholderOptions) Locator {
	exact := false
	if len(options) == 1 {
		if options[0].Exact != nil && *options[0].Exact {
			exact = true
		}
	}
	selector, err := getByPlaceholderSelector(text, exact)
	if err != nil {
		return newErrorLocator(fl.frame, err)
	}
	return fl.Locator(selector)
}

func (fl *frameLocatorImpl) GetByRole(role AriaRole, options ...FrameLocatorGetByRoleOptions) Locator {
	if len(options) == 1 {
		selector, err := getByRoleSelector(role, LocatorGetByRoleOptions(options[0]))
		if err != nil {
			return newErrorLocator(fl.frame, err)
		}
		return fl.Locator(selector)
	}
	selector, err := getByRoleSelector(role)
	if err != nil {
		return newErrorLocator(fl.frame, err)
	}
	return fl.Locator(selector)
}

func (fl *frameLocatorImpl) GetByTestId(testId any) Locator {
	selector, err := getByTestIdSelector(getTestIdAttributeName(), testId)
	if err != nil {
		return newErrorLocator(fl.frame, err)
	}
	return fl.Locator(selector)
}

func (fl *frameLocatorImpl) GetByText(text any, options ...FrameLocatorGetByTextOptions) Locator {
	exact := false
	if len(options) == 1 {
		if options[0].Exact != nil && *options[0].Exact {
			exact = true
		}
	}
	selector, err := getByTextSelector(text, exact)
	if err != nil {
		return newErrorLocator(fl.frame, err)
	}
	return fl.Locator(selector)
}

func (fl *frameLocatorImpl) GetByTitle(text any, options ...FrameLocatorGetByTitleOptions) Locator {
	exact := false
	if len(options) == 1 {
		if options[0].Exact != nil && *options[0].Exact {
			exact = true
		}
	}
	selector, err := getByTitleSelector(text, exact)
	if err != nil {
		return newErrorLocator(fl.frame, err)
	}
	return fl.Locator(selector)
}

func (fl *frameLocatorImpl) Last() FrameLocator {
	return newFrameLocator(fl.frame, fl.frameSelector+" >> nth=-1")
}

func (fl *frameLocatorImpl) Locator(selectorOrLocator any, options ...FrameLocatorLocatorOptions) Locator {
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
		return newLocator(fl.frame, fl.frameSelector+" >> internal:control=enter-frame >> "+selector, option)
	}
	locator, ok := selectorOrLocator.(*locatorImpl)
	if ok {
		if fl.frame != locator.frame {
			return locator.withError(ErrLocatorNotSameFrame)
		}
		return newLocator(
			locator.frame,
			fmt.Sprintf("%s >> internal:control=enter-frame >> %s", fl.frameSelector, locator.selector),
			option,
		)
	}
	return &locatorImpl{
		frame:    fl.frame,
		selector: fl.frameSelector,
		err:      fmt.Errorf("invalid locator parameter: %v", selectorOrLocator),
	}
}

func (fl *frameLocatorImpl) Nth(index int) FrameLocator {
	return newFrameLocator(fl.frame, fl.frameSelector+" >> nth="+strconv.Itoa(index))
}

func (fl *frameLocatorImpl) Owner() Locator {
	return newLocator(fl.frame, fl.frameSelector)
}
