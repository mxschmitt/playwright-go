package playwright

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type apiResponseAssertionsImpl struct {
	actual APIResponse
	isNot  bool
}

func newAPIResponseAssertions(actual APIResponse, isNot bool) *apiResponseAssertionsImpl {
	return &apiResponseAssertionsImpl{
		actual: actual,
		isNot:  isNot,
	}
}

func (ar *apiResponseAssertionsImpl) Not() APIResponseAssertions {
	return newAPIResponseAssertions(ar.actual, true)
}

func (ar *apiResponseAssertionsImpl) ToBeOK() error {
	if ar.isNot != ar.actual.Ok() {
		return nil
	}
	message := fmt.Sprintf(`Response status expected to be within [200..299] range, was %v`, ar.actual.Status())
	if ar.isNot {
		message = strings.ReplaceAll(message, "expected to", "expected not to")
	}
	logList, err := ar.actual.(*apiResponseImpl).fetchLog()
	if err != nil {
		return err
	}
	log := strings.Join(logList, "\n")
	if log != "" {
		message += "\nCall log:\n" + log
	}

	isTextEncoding := false
	contentType, ok := ar.actual.Headers()["content-type"]
	if ok {
		isTextEncoding = isTexualMimeType(contentType)
	}
	if isTextEncoding {
		text, err := ar.actual.Text()
		if err == nil {
			message += fmt.Sprintf(`\n Response Text:\n %s`, subString(text, 0, 1000))
		}
	}
	return errors.New(message)
}

func isTexualMimeType(mimeType string) bool {
	re := regexp.MustCompile(`^(text\/.*?|application\/(json|(x-)?javascript|xml.*?|ecmascript|graphql|x-www-form-urlencoded)|image\/svg(\+xml)?|application\/.*?(\+json|\+xml))(;\s*charset=.*)?$`)
	return re.MatchString(mimeType)
}

func subString(s string, start, length int) string {
	if start < 0 {
		start = 0
	}
	if length < 0 {
		length = 0
	}
	rs := []rune(s)
	end := min(start+length, len(rs))
	return string(rs[start:end])
}
