package playwright

import (
	"errors"
	"fmt"
	"strconv"
)

// anyFrameSelector starts the search in any frame of the subtree. It is used by
// Page.FrameLocator and Frame.FrameLocator when called without a selector.
const anyFrameSelector = "internal:control=any-frame"

var errNthFrameOnAnyFrame = errors.New("selecting the nth frame is not allowed on FrameLocator() without a selector")

type frameLocatorImpl struct {
	frame         *frameImpl
	frameSelector string
	// err is reported by every locator created from this frame locator, since
	// FrameLocator methods have no error return of their own.
	err error
}

func newFrameLocator(frame *frameImpl, frameSelector string) *frameLocatorImpl {
	return &frameLocatorImpl{frame: frame, frameSelector: frameSelector}
}

func (fl *frameLocatorImpl) childSelector(selector string) string {
	if fl.frameSelector == anyFrameSelector {
		return fl.frameSelector + " >> " + selector
	}
	return fl.frameSelector + " >> internal:control=enter-frame >> " + selector
}

func (fl *frameLocatorImpl) nth(nth string) *frameLocatorImpl {
	if fl.frameSelector == anyFrameSelector {
		return &frameLocatorImpl{frame: fl.frame, frameSelector: fl.frameSelector, err: errors.Join(fl.err, errNthFrameOnAnyFrame)}
	}
	return &frameLocatorImpl{frame: fl.frame, frameSelector: fl.frameSelector + " >> nth=" + nth, err: fl.err}
}

func (fl *frameLocatorImpl) withLocatorError(locator *locatorImpl) *locatorImpl {
	if fl.err == nil {
		return locator
	}
	return locator.withError(fl.err)
}

func (fl *frameLocatorImpl) First() FrameLocator {
	return fl.nth("0")
}

func (fl *frameLocatorImpl) FrameLocator(selector string) FrameLocator {
	return &frameLocatorImpl{frame: fl.frame, frameSelector: fl.childSelector(selector), err: fl.err}
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
		return fl.withLocatorError(newErrorLocator(fl.frame, err))
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
		return fl.withLocatorError(newErrorLocator(fl.frame, err))
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
		return fl.withLocatorError(newErrorLocator(fl.frame, err))
	}
	return fl.Locator(selector)
}

func (fl *frameLocatorImpl) GetByRole(role AriaRole, options ...FrameLocatorGetByRoleOptions) Locator {
	if len(options) == 1 {
		selector, err := getByRoleSelector(role, LocatorGetByRoleOptions(options[0]))
		if err != nil {
			return fl.withLocatorError(newErrorLocator(fl.frame, err))
		}
		return fl.Locator(selector)
	}
	selector, err := getByRoleSelector(role)
	if err != nil {
		return fl.withLocatorError(newErrorLocator(fl.frame, err))
	}
	return fl.Locator(selector)
}

func (fl *frameLocatorImpl) GetByTestId(testId any) Locator {
	selector, err := getByTestIdSelector(getTestIdAttributeName(), testId)
	if err != nil {
		return fl.withLocatorError(newErrorLocator(fl.frame, err))
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
		return fl.withLocatorError(newErrorLocator(fl.frame, err))
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
		return fl.withLocatorError(newErrorLocator(fl.frame, err))
	}
	return fl.Locator(selector)
}

func (fl *frameLocatorImpl) Last() FrameLocator {
	return fl.nth("-1")
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
		return fl.withLocatorError(newLocator(fl.frame, fl.childSelector(selector), option))
	}
	locator, ok := selectorOrLocator.(*locatorImpl)
	if ok {
		if fl.frame != locator.frame {
			return locator.withError(ErrLocatorNotSameFrame)
		}
		chained := fl.withLocatorError(newLocator(locator.frame, fl.childSelector(locator.selector), option))
		if locator.err != nil {
			return chained.withError(locator.err)
		}
		return chained
	}
	return &locatorImpl{
		frame:    fl.frame,
		selector: fl.frameSelector,
		err:      fmt.Errorf("invalid locator parameter: %v", selectorOrLocator),
	}
}

func (fl *frameLocatorImpl) Nth(index int) FrameLocator {
	return fl.nth(strconv.Itoa(index))
}

func (fl *frameLocatorImpl) Owner() Locator {
	return fl.withLocatorError(newLocator(fl.frame, fl.frameSelector))
}
