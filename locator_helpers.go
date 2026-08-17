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

// errInvalidTextArgument reports an argument that is neither a string nor a
// regexp. Upstream relies on its `string | RegExp` types to rule this out at
// compile time; the Go signatures are `any`, so it has to be checked here.
func errInvalidTextArgument(text any) error {
	return fmt.Errorf("expected string or *regexp.Regexp, but got %T", text)
}

func escapeForAttributeSelector(text any, exact bool) (string, error) {
	switch text := text.(type) {
	case *regexp.Regexp:
		return escapeRegexForSelector(text), nil
	case string:
		suffix := "i"
		if exact {
			suffix = "s"
		}
		return fmt.Sprintf(`"%s"%s`, strings.ReplaceAll(strings.ReplaceAll(text, `\`, `\\`), `"`, `\"`), suffix), nil
	default:
		return "", errInvalidTextArgument(text)
	}
}

func escapeForTextSelector(text any, exact bool) (string, error) {
	switch text := text.(type) {
	case *regexp.Regexp:
		return escapeRegexForSelector(text), nil
	case string:
		if exact {
			return fmt.Sprintf(`%ss`, escapeText(text)), nil
		}
		return fmt.Sprintf(`%si`, escapeText(text)), nil
	default:
		return "", errInvalidTextArgument(text)
	}
}

func escapeRegexForSelector(re *regexp.Regexp) string {
	pattern, flag := convertRegexp(re)
	// Escape any unescaped quote characters so a regex containing a quote does
	// not break the selector tokenizer when the locator is chained (`>>`).
	// Mirrors upstream's /(^|[^\\])(\\\\)*(["'`])/g → '$1$2\\$3' replacement.
	pattern = escapeQuotesInRegexSource(pattern)
	return fmt.Sprintf(`/%s/%s`, strings.ReplaceAll(pattern, `>>`, `\>\>`), flag)
}

// escapeQuotesInRegexSource backslash-escapes ", ' and ` characters that are not
// already escaped by a preceding (even-length run of) backslashes.
func escapeQuotesInRegexSource(s string) string {
	var b strings.Builder
	backslashes := 0
	for _, r := range s {
		if r == '"' || r == '\'' || r == '`' {
			// Already escaped only if preceded by an odd number of backslashes.
			if backslashes%2 == 0 {
				b.WriteByte('\\')
			}
		}
		if r == '\\' {
			backslashes++
		} else {
			backslashes = 0
		}
		b.WriteRune(r)
	}
	return b.String()
}

// locatorCustomDescription extracts the description embedded by Describe via the
// trailing `internal:describe=<json>` selector segment, mirroring upstream
// locatorCustomDescription. Returns "" when no description is present.
func locatorCustomDescription(selector string) string {
	const marker = " >> internal:describe="
	idx := strings.LastIndex(selector, marker)
	if idx == -1 {
		return ""
	}
	// The describe segment is always the last one (Describe appends it).
	jsonValue := selector[idx+len(marker):]
	var description string
	if err := json.Unmarshal([]byte(jsonValue), &description); err != nil {
		return ""
	}
	return description
}

func escapeText(s string) string {
	builder := &strings.Builder{}
	encoder := json.NewEncoder(builder)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(s)
	return strings.TrimSpace(builder.String())
}

func getByAltTextSelector(text any, exact bool) (string, error) {
	return getByAttributeTextSelector("alt", text, exact)
}

func getByAttributeTextSelector(attrName string, text any, exact bool) (string, error) {
	escaped, err := escapeForAttributeSelector(text, exact)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`internal:attr=[%s=%s]`, attrName, escaped), nil
}

func getByLabelSelector(text any, exact bool) (string, error) {
	escaped, err := escapeForTextSelector(text, exact)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`internal:label=%s`, escaped), nil
}

func getByPlaceholderSelector(text any, exact bool) (string, error) {
	return getByAttributeTextSelector("placeholder", text, exact)
}

func getByRoleSelector(role AriaRole, options ...LocatorGetByRoleOptions) (string, error) {
	// Keep the property ordering stable and identical to upstream, since the
	// produced selector string must be deterministic.
	props := make([][2]string, 0)
	if len(options) == 1 {
		exact := false
		if options[0].Exact != nil {
			exact = *options[0].Exact
		}
		if options[0].Checked != nil {
			props = append(props, [2]string{"checked", fmt.Sprintf("%t", *options[0].Checked)})
		}
		if options[0].Disabled != nil {
			props = append(props, [2]string{"disabled", fmt.Sprintf("%t", *options[0].Disabled)})
		}
		if options[0].Selected != nil {
			props = append(props, [2]string{"selected", fmt.Sprintf("%t", *options[0].Selected)})
		}
		if options[0].Expanded != nil {
			props = append(props, [2]string{"expanded", fmt.Sprintf("%t", *options[0].Expanded)})
		}
		if options[0].IncludeHidden != nil {
			props = append(props, [2]string{"include-hidden", fmt.Sprintf("%t", *options[0].IncludeHidden)})
		}
		if options[0].Level != nil {
			props = append(props, [2]string{"level", fmt.Sprintf("%d", *options[0].Level)})
		}
		if options[0].Name != nil {
			escaped, err := escapeForAttributeSelector(options[0].Name, exact)
			if err != nil {
				return "", err
			}
			props = append(props, [2]string{"name", escaped})
		}
		if options[0].Description != nil {
			escaped, err := escapeForAttributeSelector(options[0].Description, exact)
			if err != nil {
				return "", err
			}
			props = append(props, [2]string{"description", escaped})
		}
		if options[0].Pressed != nil {
			props = append(props, [2]string{"pressed", fmt.Sprintf("%t", *options[0].Pressed)})
		}
	}
	var propsStr strings.Builder
	for _, p := range props {
		propsStr.WriteString("[" + p[0] + "=" + p[1] + "]")
	}
	return fmt.Sprintf("internal:role=%s%s", role, propsStr.String()), nil
}

func getByTextSelector(text any, exact bool) (string, error) {
	escaped, err := escapeForTextSelector(text, exact)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`internal:text=%s`, escaped), nil
}

func getByTestIdSelector(testIdAttributeName string, testId any) (string, error) {
	escaped, err := escapeForAttributeSelector(testId, true)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`internal:testid=[%s=%s]`, encodeTestIdAttributeName(testIdAttributeName), escaped), nil
}

// encodeTestIdAttributeName JSON-quotes the attribute name when it contains a
// comma so the engine treats a comma-separated list (e.g. "data-pw,data-ti") as
// multiple attribute names rather than malforming the selector. Mirrors upstream
// encodeTestIdAttributeName (packages/isomorphic/locatorUtils.ts).
func encodeTestIdAttributeName(testIdAttributeName string) string {
	if strings.Contains(testIdAttributeName, ",") {
		encoded, err := json.Marshal(testIdAttributeName)
		if err == nil {
			return string(encoded)
		}
	}
	return testIdAttributeName
}

func getByTitleSelector(text any, exact bool) (string, error) {
	return getByAttributeTextSelector("title", text, exact)
}

func getTestIdAttributeName() string {
	return testIdAttributeName
}

func setTestIdAttributeName(name string) {
	testIdAttributeName = name
}
