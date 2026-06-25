package playwright

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var escapedChars = map[rune]bool{
	'$':  true,
	'^':  true,
	'+':  true,
	'.':  true,
	'*':  true,
	'(':  true,
	')':  true,
	'|':  true,
	'\\': true,
	'?':  true,
	'{':  true,
	'}':  true,
	'[':  true,
	']':  true,
}

func globMustToRegex(glob string) *regexp.Regexp {
	tokens := []string{"^"}
	inGroup := false

	for i := 0; i < len(glob); i++ {
		c := rune(glob[i])
		if c == '\\' && i+1 < len(glob) {
			char := rune(glob[i+1])
			if _, ok := escapedChars[char]; ok {
				tokens = append(tokens, "\\"+string(char))
			} else {
				tokens = append(tokens, string(char))
			}
			i++
		} else if c == '*' {
			charBefore := rune(0)
			if i > 0 {
				charBefore = rune(glob[i-1])
			}
			starCount := 1
			for i+1 < len(glob) && glob[i+1] == '*' {
				starCount++
				i++
			}
			if starCount > 1 {
				charAfter := rune(0)
				if i+1 < len(glob) {
					charAfter = rune(glob[i+1])
				}
				// Match either /..something../ or /.
				if charAfter == '/' {
					if charBefore == '/' {
						tokens = append(tokens, "((.+/)|)")
					} else {
						tokens = append(tokens, "(.*/)")
					}
					i++
				} else {
					tokens = append(tokens, "(.*)")
				}
			} else {
				tokens = append(tokens, "([^/]*)")
			}
		} else {
			switch c {
			case '{':
				inGroup = true
				tokens = append(tokens, "(")
			case '}':
				inGroup = false
				tokens = append(tokens, ")")
			case ',':
				if inGroup {
					tokens = append(tokens, "|")
				} else {
					tokens = append(tokens, "\\"+string(c))
				}
			default:
				if _, ok := escapedChars[c]; ok {
					tokens = append(tokens, "\\"+string(c))
				} else {
					tokens = append(tokens, string(c))
				}
			}
		}
	}

	tokens = append(tokens, "$")
	return regexp.MustCompile(strings.Join(tokens, ""))
}

func resolveGlobToRegex(baseURL *string, glob string, isWebSocketUrl bool) *regexp.Regexp {
	if isWebSocketUrl {
		baseURL = toWebSocketBaseURL(baseURL)
	}
	glob = resolveGlobBase(baseURL, glob)
	return globMustToRegex(glob)
}

func resolveGlobBase(baseURL *string, match string) string {
	if strings.HasPrefix(match, "*") {
		return match
	}

	tokenMap := make(map[string]string)
	mapToken := func(original string, replacement string) string {
		if len(original) == 0 {
			return ""
		}
		tokenMap[replacement] = original
		return replacement
	}
	// Escaped `\\?` behaves the same as `?` in our glob patterns.
	match = strings.ReplaceAll(match, `\\?`, "?")
	// Special case about:/data:/chrome:/edge:/file: URLs as they are not relative to baseURL.
	if strings.HasPrefix(match, "about:") || strings.HasPrefix(match, "data:") ||
		strings.HasPrefix(match, "chrome:") || strings.HasPrefix(match, "edge:") ||
		strings.HasPrefix(match, "file:") {
		return match
	}
	// Glob symbols may be escaped in the URL and some of them such as ? affect resolution,
	// so we replace them with safe components first.
	relativePath := strings.Split(match, "/")
	for i, token := range relativePath {
		if token == "." || token == ".." || token == "" {
			continue
		}
		// Handle special case of http*://, note that the new schema has to be
		// a web schema so that slashes are properly inserted after domain.
		if i == 0 && strings.HasSuffix(token, ":") {
			// Replace any pattern with http:; preserve an explicit schema as-is as
			// it may affect trailing slashes after the domain.
			if strings.ContainsAny(token, "*{") {
				relativePath[i] = mapToken(token, "http:")
			}
		} else {
			questionIndex := strings.Index(token, "?")
			if questionIndex == -1 {
				relativePath[i] = mapToken(token, "$_"+strconv.Itoa(i)+"_$")
			} else {
				newPrefix := mapToken(token[:questionIndex], "$_"+strconv.Itoa(i)+"_$")
				newSuffix := mapToken(token[questionIndex:], "?$"+strconv.Itoa(i)+"_$")
				relativePath[i] = newPrefix + newSuffix
			}
		}
	}
	resolved := constructURLBasedOnBaseURL(baseURL, strings.Join(relativePath, "/"))
	for token, original := range tokenMap {
		resolved = strings.ReplaceAll(resolved, token, original)
	}
	return resolved
}

func constructURLBasedOnBaseURL(baseURL *string, givenURL string) string {
	u, err := url.Parse(givenURL)
	if err != nil {
		return givenURL
	}
	if baseURL != nil {
		base, err := url.Parse(*baseURL)
		if err != nil {
			return givenURL
		}
		u = base.ResolveReference(u)
	}
	if u.Path == "" { // In Node.js, new URL('http://localhost') returns 'http://localhost/'.
		u.Path = "/"
	}
	return u.String()
}

func toWebSocketBaseURL(baseURL *string) *string {
	if baseURL == nil {
		return nil
	}

	// Allow http(s) baseURL to match ws(s) urls.
	re := regexp.MustCompile(`(?m)^http(s?://)`)
	wsBaseURL := re.ReplaceAllString(*baseURL, "ws$1")

	return &wsBaseURL
}
