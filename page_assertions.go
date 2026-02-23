package playwright

import (
	"fmt"
	"net/url"
	"path"
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
	result, err := frame.channel.SendReturnAsDict("expect", options, overrides)
	if err != nil {
		return err
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
		u, _ := url.Parse(*baseURL)
		u.Path = path.Join(u.Path, urlPath)
		urlOrRegExp = u.String()
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

func (pa *pageAssertionsImpl) Not() PageAssertions {
	return newPageAssertions(pa.actualPage, true, pa.defaultTimeout)
}
