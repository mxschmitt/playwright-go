package playwright

import (
	"strings"
)

type rawHeaders struct {
	headersArray []NameValue
	headersMap   map[string][]string
}

func (r *rawHeaders) Get(name string) string {
	values := r.GetAll(name)
	if len(values) == 0 {
		return ""
	}
	sep := ", "
	if strings.ToLower(name) == "set-cookie" {
		sep = "\n"
	}
	return strings.Join(values, sep)
}

func (r *rawHeaders) GetAll(name string) []string {
	name = strings.ToLower(name)
	if _, ok := r.headersMap[name]; !ok {
		return []string{}
	}
	return r.headersMap[name]
}

func (r *rawHeaders) Headers() map[string]string {
	out := make(map[string]string)
	for key := range r.headersMap {
		out[key] = r.Get(key)
	}
	return out
}

func (r *rawHeaders) HeadersArray() []NameValue {
	return r.headersArray
}

func (r *rawHeaders) addHeader(name, value string) {
	r.headersArray = append(r.headersArray, NameValue{
		Name:  name,
		Value: value,
	})
	if _, ok := r.headersMap[strings.ToLower(name)]; !ok {
		r.headersMap[strings.ToLower(name)] = make([]string, 0)
	}
	r.headersMap[strings.ToLower(name)] = append(r.headersMap[strings.ToLower(name)], value)
}

func newRawHeaders(headers any) *rawHeaders {
	r := &rawHeaders{}
	r.headersArray = make([]NameValue, 0)
	r.headersMap = make(map[string][]string)
	switch typed := headers.(type) {
	case []any:
		// Headers materialized from the wire protocol arrive as a slice of
		// decoded JSON objects.
		for _, header := range typed {
			entry := header.(map[string]any)
			r.addHeader(entry["name"].(string), entry["value"].(string))
		}
	case []NameValue:
		// Headers produced internally (e.g. serializeMapToNameAndValue for
		// fallback overrides) are already typed.
		for _, entry := range typed {
			r.addHeader(entry.Name, entry.Value)
		}
	default:
		panic(headers)
	}
	return r
}
