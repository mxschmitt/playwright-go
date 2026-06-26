package playwright

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type pageAssertionsImpl struct {
	assertionsBase
	actualPage Page
}

func newPageAssertions(page Page, isNot bool, defaultTimeout *float64) *pageAssertionsImpl {
	return &pageAssertionsImpl{
		assertionsBase: assertionsBase{
			actualLocator:  page.Locator(":root"),
			isNot:          isNot,
			defaultTimeout: defaultTimeout,
		},
		actualPage: page,
	}
}

// expectOnFrame calls the frame's expect method directly without a selector.
// This is needed for page-level assertions like ToHaveTitle and ToHaveURL
// which should not be bound to a specific element.
func (pa *pageAssertionsImpl) expectOnFrame(
	expression string,
	options frameExpectOptions,
	expected any,
	message string,
) error {
	options.IsNot = pa.isNot
	if options.Timeout == nil {
		options.Timeout = pa.defaultTimeout
	}
	if options.IsNot {
		message = strings.ReplaceAll(message, "expected to", "expected not to")
	}

	frame := pa.actualPage.MainFrame().(*frameImpl)
	overrides := map[string]any{
		"expression": expression,
	}

	var (
		received any
		matches  bool
		log      []string
	)
	_, err := frame.channel.SendReturnAsDict("expect", options, overrides)
	if err != nil {
		// Since v1.61 a failed assertion is reported as a server error carrying
		// structured errorDetails rather than a `{ matches: false }` result.
		var detailed *errorWithDetails
		if !errors.As(err, &detailed) {
			return err
		}
		matches = options.IsNot
		received = parseExpectReceived(detailed.details["received"])
		log = detailed.log
	} else {
		// No error means the assertion matched.
		matches = !options.IsNot
	}

	if matches == pa.isNot {
		actual := received
		logStr := strings.Join(log, "\n")
		if logStr != "" {
			logStr = "\nCall log:\n" + logStr
		}
		if expected != nil {
			return fmt.Errorf("%s '%v'\nActual value: %v %s", message, expected, actual, logStr)
		}
		return fmt.Errorf("%s\nActual value: %v %s", message, actual, logStr)
	}

	return nil
}

func (pa *pageAssertionsImpl) ToHaveTitle(titleOrRegExp any, options ...PageAssertionsToHaveTitleOptions) error {
	var timeout *float64
	if len(options) == 1 {
		timeout = options[0].Timeout
	}
	expectedValues, err := toExpectedTextValues([]any{titleOrRegExp}, false, true, nil)
	if err != nil {
		return err
	}
	return pa.expectOnFrame(
		"to.have.title",
		frameExpectOptions{ExpectedText: expectedValues, Timeout: timeout},
		titleOrRegExp,
		"Page title expected to be",
	)
}

func (pa *pageAssertionsImpl) ToHaveURL(urlOrRegExp any, options ...PageAssertionsToHaveURLOptions) error {
	var timeout *float64
	var ignoreCase *bool
	if len(options) == 1 {
		timeout = options[0].Timeout
		ignoreCase = options[0].IgnoreCase
	}

	baseURL := pa.actualPage.Context().(*browserContextImpl).options.BaseURL
	if urlPath, ok := urlOrRegExp.(string); ok && baseURL != nil {
		// Resolve against the base URL the same way the browser's `new URL(given,
		// base)` does, rather than naive path joining (matching upstream
		// constructURLBasedOnBaseURL).
		if base, err := url.Parse(*baseURL); err == nil {
			if given, err := url.Parse(urlPath); err == nil {
				urlOrRegExp = base.ResolveReference(given).String()
			}
		}
	}

	expectedValues, err := toExpectedTextValues([]any{urlOrRegExp}, false, false, ignoreCase)
	if err != nil {
		return err
	}
	return pa.expectOnFrame(
		"to.have.url",
		frameExpectOptions{ExpectedText: expectedValues, Timeout: timeout},
		urlOrRegExp,
		"Page URL expected to be",
	)
}

func (pa *pageAssertionsImpl) ToMatchAriaSnapshot(expected string, options ...PageAssertionsToMatchAriaSnapshotOptions) error {
	var timeout *float64
	if len(options) == 1 {
		timeout = options[0].Timeout
	}
	return pa.expect(
		"to.match.aria",
		frameExpectOptions{ExpectedValue: expected, Timeout: timeout},
		expected,
		"Page expected to match ARIA snapshot",
	)
}

func (pa *pageAssertionsImpl) Not() PageAssertions {
	return newPageAssertions(pa.actualPage, true, pa.defaultTimeout)
}
