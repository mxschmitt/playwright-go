package playwright

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

func convertRegexp(reg *regexp.Regexp) (pattern, flags string) {
	matches := regexp.MustCompile(`\(\?([imsU]+)\)(.+)`).FindStringSubmatch(reg.String())

	if len(matches) == 3 {
		pattern = matches[2]
		flags = matches[1]
	} else {
		pattern = reg.String()
	}
	return
}

func escapeForAttributeSelector(text any, exact bool) string {
	switch text := text.(type) {
	case *regexp.Regexp:
		return escapeRegexForSelector(text)
	default:
		suffix := "i"
		if exact {
			suffix = "s"
		}
		return fmt.Sprintf(`"%s"%s`, strings.Replace(strings.Replace(text.(string), `\`, `\\`, -1), `"`, `\"`, -1), suffix)
	}
}

func escapeForTextSelector(text any, exact bool) string {
	switch text := text.(type) {
	case *regexp.Regexp:
		return escapeRegexForSelector(text)
	default:
		if exact {
			return fmt.Sprintf(`%ss`, escapeText(text.(string)))
		}
		return fmt.Sprintf(`%si`, escapeText(text.(string)))
	}
}

func escapeRegexForSelector(re *regexp.Regexp) string {
	pattern, flag := convertRegexp(re)
	return fmt.Sprintf(`/%s/%s`, strings.ReplaceAll(pattern, `>>`, `\>\>`), flag)
}

func escapeText(s string) string {
	builder := &strings.Builder{}
	encoder := json.NewEncoder(builder)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(s)
	return strings.TrimSpace(builder.String())
}

func getByAltTextSelector(text any, exact bool) string {
	return getByAttributeTextSelector("alt", text, exact)
}

func getByAttributeTextSelector(attrName string, text any, exact bool) string {
	return fmt.Sprintf(`internal:attr=[%s=%s]`, attrName, escapeForAttributeSelector(text, exact))
}

func getByLabelSelector(text any, exact bool) string {
	return fmt.Sprintf(`internal:label=%s`, escapeForTextSelector(text, exact))
}

func getByPlaceholderSelector(text any, exact bool) string {
	return getByAttributeTextSelector("placeholder", text, exact)
}

func getByRoleSelector(role AriaRole, options ...LocatorGetByRoleOptions) string {
	props := make(map[string]string)
	if len(options) == 1 {
		if options[0].Checked != nil {
			props["checked"] = fmt.Sprintf("%t", *options[0].Checked)
		}
		if options[0].Disabled != nil {
			props["disabled"] = fmt.Sprintf("%t", *options[0].Disabled)
		}
		if options[0].Selected != nil {
			props["selected"] = fmt.Sprintf("%t", *options[0].Selected)
		}
		if options[0].Expanded != nil {
			props["expanded"] = fmt.Sprintf("%t", *options[0].Expanded)
		}
		if options[0].IncludeHidden != nil {
			props["include-hidden"] = fmt.Sprintf("%t", *options[0].IncludeHidden)
		}
		if options[0].Level != nil {
			props["level"] = fmt.Sprintf("%d", *options[0].Level)
		}
		if options[0].Name != nil {
			exact := false
			if options[0].Exact != nil {
				exact = *options[0].Exact
			}
			props["name"] = escapeForAttributeSelector(options[0].Name, exact)
		}
		if options[0].Pressed != nil {
			props["pressed"] = fmt.Sprintf("%t", *options[0].Pressed)
		}
	}
	var propsStr strings.Builder
	for k, v := range props {
		propsStr.WriteString("[" + k + "=" + v + "]")
	}
	return fmt.Sprintf("internal:role=%s%s", role, propsStr.String())
}

func getByTextSelector(text any, exact bool) string {
	return fmt.Sprintf(`internal:text=%s`, escapeForTextSelector(text, exact))
}

func getByTestIdSelector(testIdAttributeName string, testId any) string {
	return fmt.Sprintf(`internal:testid=[%s=%s]`, testIdAttributeName, escapeForAttributeSelector(testId, true))
}

func getByTitleSelector(text any, exact bool) string {
	return getByAttributeTextSelector("title", text, exact)
}

func getTestIdAttributeName() string {
	return testIdAttributeName
}

func setTestIdAttributeName(name string) {
	testIdAttributeName = name
}
