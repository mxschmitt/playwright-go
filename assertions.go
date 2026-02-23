package playwright

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

const assertionsDefaultTimeout = 5000 // 5s

type playwrightAssertionsImpl struct {
	defaultTimeout *float64
}

// NewPlaywrightAssertions creates a new instance of PlaywrightAssertions
//   - timeout: default value is 5000 (ms)
func NewPlaywrightAssertions(timeout ...float64) PlaywrightAssertions {
	if len(timeout) > 0 {
		return &playwrightAssertionsImpl{Float(timeout[0])}
	}
	return &playwrightAssertionsImpl{Float(assertionsDefaultTimeout)}
}

func (pa *playwrightAssertionsImpl) APIResponse(response APIResponse) APIResponseAssertions {
	return newAPIResponseAssertions(response, false)
}

func (pa *playwrightAssertionsImpl) Locator(locator Locator) LocatorAssertions {
	return newLocatorAssertions(locator, false, pa.defaultTimeout)
}

func (pa *playwrightAssertionsImpl) Page(page Page) PageAssertions {
	return newPageAssertions(page, false, pa.defaultTimeout)
}

type expectedTextValue struct {
	Str                 *string `json:"string,omitempty"`
	RegexSource         *string `json:"regexSource,omitempty"`
	RegexFlags          *string `json:"regexFlags,omitempty"`
	MatchSubstring      *bool   `json:"matchSubstring,omitempty"`
	IgnoreCase          *bool   `json:"ignoreCase,omitempty"`
	NormalizeWhiteSpace *bool   `json:"normalizeWhiteSpace,omitempty"`
}

type frameExpectOptions struct {
	ExpressionArg  any                 `json:"expressionArg,omitempty"`
	ExpectedText   []expectedTextValue `json:"expectedText,omitempty"`
	ExpectedNumber *float64            `json:"expectedNumber,omitempty"`
	ExpectedValue  any                 `json:"expectedValue,omitempty"`
	UseInnerText   *bool               `json:"useInnerText,omitempty"`
	IsNot          bool                `json:"isNot"`
	Timeout        *float64            `json:"timeout"`
}

type frameExpectResult struct {
	Matches  bool     `json:"matches"`
	Received any      `json:"received,omitempty"`
	TimedOut *bool    `json:"timedOut,omitempty"`
	Log      []string `json:"log,omitempty"`
}

type assertionsBase struct {
	actualLocator  Locator
	isNot          bool
	defaultTimeout *float64
}

func (b *assertionsBase) expect(
	expression string,
	options frameExpectOptions,
	expected any,
	message string,
) error {
	options.IsNot = b.isNot
	if options.Timeout == nil {
		options.Timeout = b.defaultTimeout
	}
	if options.IsNot {
		message = strings.ReplaceAll(message, "expected to", "expected not to")
	}
	result, err := b.actualLocator.(*locatorImpl).expect(expression, options)
	if err != nil {
		return err
	}

	if result.Matches == b.isNot {
		actual := result.Received
		log := strings.Join(result.Log, "\n")
		if log != "" {
			log = "\nCall log:\n" + log
		}
		if expected != nil {
			return fmt.Errorf("%s '%v'\nActual value: %v %s", message, expected, actual, log)
		}
		return fmt.Errorf("%s\nActual value: %v %s", message, actual, log)
	}

	return nil
}

func toExpectedTextValues(
	items []any,
	matchSubstring bool,
	normalizeWhiteSpace bool,
	ignoreCase *bool,
) ([]expectedTextValue, error) {
	var out []expectedTextValue
	for _, item := range items {
		switch item := item.(type) {
		case string:
			out = append(out, expectedTextValue{
				Str:                 String(item),
				MatchSubstring:      Bool(matchSubstring),
				NormalizeWhiteSpace: Bool(normalizeWhiteSpace),
				IgnoreCase:          ignoreCase,
			})
		case *regexp.Regexp:
			pattern, flags := convertRegexp(item)
			out = append(out, expectedTextValue{
				RegexSource:         String(pattern),
				RegexFlags:          String(flags),
				MatchSubstring:      Bool(matchSubstring),
				NormalizeWhiteSpace: Bool(normalizeWhiteSpace),
				IgnoreCase:          ignoreCase,
			})
		default:
			return nil, errors.New("value must be a string or regexp")
		}
	}
	return out, nil
}

func convertToInterfaceList(v any) []any {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return []any{v}
	}

	list := make([]any, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		list[i] = rv.Index(i).Interface()
	}
	return list
}
