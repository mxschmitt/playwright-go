package playwright

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	mapset "github.com/deckarep/golang-set/v2"
)

type (
	routeHandler = func(Route)
)

func skipFieldSerialization(val reflect.Value) bool {
	typ := val.Type()
	return (typ.Kind() == reflect.Pointer ||
		typ.Kind() == reflect.Interface ||
		typ.Kind() == reflect.Map ||
		typ.Kind() == reflect.Slice) && val.IsNil() || (val.Kind() == reflect.Interface && val.Elem().Kind() == reflect.Pointer && val.Elem().IsNil())
}

func transformStructValues(in any) any {
	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if _, ok := in.(*channel); ok {
		return in
	}
	if v.Kind() == reflect.Map || v.Kind() == reflect.Struct {
		return transformStructIntoMapIfNeeded(in)
	}
	if v.Kind() == reflect.Slice {
		outSlice := []any{}
		for i := 0; i < v.Len(); i++ {
			if !skipFieldSerialization(v.Index(i)) {
				outSlice = append(outSlice, transformStructValues(v.Index(i).Interface()))
			}
		}
		return outSlice
	}
	if v.Interface() == Null() || (v.Kind() == reflect.String && v.String() == Null().(string)) {
		return "null"
	}
	return in
}

func transformStructIntoMapIfNeeded(inStruct any) map[string]any {
	out := make(map[string]any)
	v := reflect.ValueOf(inStruct)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	typ := v.Type()
	if v.Kind() == reflect.Struct {
		// Merge into the base map by the JSON struct tag
		for i := 0; i < v.NumField(); i++ {
			fi := typ.Field(i)
			tagv := fi.Tag.Get("json")
			key := strings.Split(tagv, ",")[0]
			if key == "" {
				key = fi.Name
			}
			// Special handling for timeout field: provide default value when nil
			// This is required in Playwright v1.57+ protocol where timeout is no longer optional
			if key == "timeout" && skipFieldSerialization(v.Field(i)) {
				out[key] = float64(30000) // default 30s
				continue
			}
			// Skip the values when the field is a pointer (like *string) and nil.
			if fi.IsExported() && !skipFieldSerialization(v.Field(i)) {
				// We use the JSON struct fields for getting the original names
				// out of the field.
				out[key] = transformStructValues(v.Field(i).Interface())
			}
		}
	} else if v.Kind() == reflect.Map {
		// Merge into the base map
		for _, key := range v.MapKeys() {
			if !skipFieldSerialization(v.MapIndex(key)) {
				out[key.String()] = transformStructValues(v.MapIndex(key).Interface())
			}
		}
	}
	return out
}

// transformOptions handles the parameter data transformation
func transformOptions(options ...any) map[string]any {
	var base map[string]any
	var option any
	// Case 1: No options are given
	if len(options) == 0 {
		return make(map[string]any)
	}
	if len(options) == 1 {
		// Case 2: a single value (either struct or map) is given.
		base = make(map[string]any)
		option = options[0]
	} else if len(options) == 2 {
		// Case 3: two values are given. The first one needs to be transformed
		// to a map, the sencond one will be then get merged into the first
		// base map.
		if reflect.ValueOf(options[0]).Kind() != reflect.Map {
			base = transformOptions(options[0])
		} else {
			base = transformStructIntoMapIfNeeded(options[0])
		}
		option = options[1]
	}
	v := reflect.ValueOf(option)
	if v.Kind() == reflect.Slice {
		if v.Len() == 0 {
			// Check if the slice element type has a Timeout field and add default if so
			// This is required in Playwright v1.57+ protocol where timeout is no longer optional
			elemType := v.Type().Elem()
			if elemType.Kind() == reflect.Struct {
				if _, hasTimeout := elemType.FieldByName("Timeout"); hasTimeout {
					if base["timeout"] == nil {
						base["timeout"] = float64(30000) // default 30s
					}
				}
			}
			return base
		}
		option = v.Index(0).Interface()
	}

	if option == nil {
		return base
	}
	v = reflect.ValueOf(option)

	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	optionMap := transformStructIntoMapIfNeeded(v.Interface())
	for key, value := range optionMap {
		base[key] = value
	}
	return base
}

func remapValue(inMapValue reflect.Value, outStructValue reflect.Value) {
	switch outStructValue.Type().Kind() {
	case reflect.Bool:
		outStructValue.SetBool(inMapValue.Bool())
	case reflect.String:
		outStructValue.SetString(inMapValue.String())
	case reflect.Float64:
		outStructValue.SetFloat(inMapValue.Float())
	case reflect.Int:
		outStructValue.SetInt(int64(inMapValue.Float()))
	case reflect.Slice:
		outStructValue.Set(reflect.MakeSlice(outStructValue.Type(), inMapValue.Len(), inMapValue.Cap()))
		for i := 0; i < inMapValue.Len(); i++ {
			remapValue(inMapValue.Index(i).Elem(), outStructValue.Index(i))
		}
	case reflect.Struct:
		structTyp := outStructValue.Type()
		for i := 0; i < outStructValue.NumField(); i++ {
			fi := structTyp.Field(i)
			key := strings.Split(fi.Tag.Get("json"), ",")[0]
			structField := outStructValue.Field(i)
			value := inMapValue.MapIndex(reflect.ValueOf(key))
			if !value.IsValid() {
				continue
			}
			if value.Kind() == reflect.Interface {
				if value.IsNil() {
					continue
				}
				value = value.Elem()
			}
			structFieldDeref := structField
			if structField.Type().Kind() == reflect.Pointer {
				structField.Set(reflect.New(structField.Type().Elem()))
				structFieldDeref = structField.Elem()
			}
			remapValue(value, structFieldDeref)
		}
	default:
		panic(inMapValue.Interface())
	}
}

func remapMapToStruct(inputMap any, outStruct any) {
	remapValue(reflect.ValueOf(inputMap), reflect.ValueOf(outStruct).Elem())
}

type urlMatcher struct {
	raw     any
	pattern *regexp.Regexp
	matchFn func(url string) bool
}

func newURLMatcher(urlOrPredicate any, baseURL *string, isWsUrl ...bool) *urlMatcher {
	switch v := urlOrPredicate.(type) {
	case *regexp.Regexp:
		return &urlMatcher{pattern: v, raw: urlOrPredicate}
	case string:
		return &urlMatcher{
			pattern: resolveGlobToRegex(baseURL, v, len(isWsUrl) > 0 && isWsUrl[0]),
			raw:     urlOrPredicate,
		}
	}
	fn, ok := urlOrPredicate.(func(string) bool)
	if ok {
		return &urlMatcher{
			matchFn: fn,
			raw:     urlOrPredicate,
		}
	}
	panic(fmt.Errorf("invalid urlOrPredicate: %v", urlOrPredicate))
}

func (u *urlMatcher) Matches(url string) bool {
	if u.matchFn != nil {
		return u.matchFn(url)
	}
	if u.pattern != nil {
		return u.pattern.MatchString(url)
	}
	return false
}

// SameWith compares String() if urlOrPredicate is *regexp.Regexp
func (u *urlMatcher) SameWith(urlOrPredicate any) bool {
	switch v := urlOrPredicate.(type) {
	case *regexp.Regexp:
		return u.pattern.String() == v.String()
	default:
		return u.raw == urlOrPredicate
	}
}

type routeHandlerInvocation struct {
	route    Route
	complete chan bool
}

type routeHandlerEntry struct {
	matcher           *urlMatcher
	handler           routeHandler
	times             int
	count             int32
	ignoreErrors      *atomic.Bool
	activeInvocations mapset.Set[*routeHandlerInvocation]
}

func (r *routeHandlerEntry) Matches(url string) bool {
	return r.matcher.Matches(url)
}

func (r *routeHandlerEntry) Handle(route Route) chan bool {
	handlerInvocation := &routeHandlerInvocation{
		route:    route,
		complete: make(chan bool, 1),
	}
	r.activeInvocations.Add(handlerInvocation)

	defer func() {
		handlerInvocation.complete <- true
		r.activeInvocations.Remove(handlerInvocation)
	}()
	defer func() {
		// If the handler was stopped (without waiting for completion), we ignore all exceptions.
		if r.ignoreErrors.Load() {
			_ = recover()
			route.(*routeImpl).reportHandled(false)
		} else {
			e := recover()
			if e != nil {
				err, ok := e.(error)
				if ok && errors.Is(err, ErrTargetClosed) {
					panic(fmt.Errorf("\"%w\" while running route callback.\nConsider awaiting `page.UnrouteAll(playwright.PageUnrouteAllOptions{Behavior: playwright.UnrouteBehaviorIgnoreErrors})`\nbefore the end of the test to ignore remaining routes in flight.", err))
				}
				panic(e)
			}
		}
	}()

	return r.handleInternal(route)
}

func (r *routeHandlerEntry) Stop(behavior string) {
	// When a handler is manually unrouted or its page/context is closed we either
	// - wait for the current handler invocations to finish
	// - or do not wait, if the user opted out of it, but swallow all exceptions
	//   that happen after the unroute/close.
	if behavior == string(*UnrouteBehaviorIgnoreErrors) {
		r.ignoreErrors.Store(true)
	} else {
		wg := &sync.WaitGroup{}
		r.activeInvocations.Each(func(activation *routeHandlerInvocation) bool {
			if !activation.route.(*routeImpl).didThrow {
				wg.Add(1)
				go func(complete chan bool) {
					<-complete
					wg.Done()
				}(activation.complete)
			}
			return false
		})
		wg.Wait()
	}
}

func (r *routeHandlerEntry) handleInternal(route Route) chan bool {
	handled := route.(*routeImpl).startHandling()
	atomic.AddInt32(&r.count, 1)
	r.handler(route)
	return handled
}

func (r *routeHandlerEntry) WillExceed() bool {
	if r.times == 0 {
		return false
	}
	return int(atomic.LoadInt32(&r.count)+1) >= r.times
}

func newRouteHandlerEntry(matcher *urlMatcher, handler routeHandler, times ...int) *routeHandlerEntry {
	n := 0
	if len(times) > 0 {
		n = times[0]
	}
	return &routeHandlerEntry{
		matcher:           matcher,
		handler:           handler,
		times:             n,
		count:             0,
		ignoreErrors:      &atomic.Bool{},
		activeInvocations: mapset.NewSet[*routeHandlerInvocation](),
	}
}

func prepareInterceptionPatterns(handlers []*routeHandlerEntry) []map[string]any {
	patterns := []map[string]any{}
	all := false
	for _, h := range handlers {
		switch h.matcher.raw.(type) {
		case *regexp.Regexp:
			pattern, flags := convertRegexp(h.matcher.raw.(*regexp.Regexp))
			patterns = append(patterns, map[string]any{
				"regexSource": pattern,
				"regexFlags":  flags,
			})
		case string:
			patterns = append(patterns, map[string]any{
				"glob": h.matcher.raw.(string),
			})
		default:
			all = true
		}
	}
	if all {
		return []map[string]any{
			{
				"glob": "**/*",
			},
		}
	}
	return patterns
}

const defaultTimeout = 30 * 1000

type timeoutSettings struct {
	sync.RWMutex
	parent                   *timeoutSettings
	defaultTimeout           *float64
	defaultNavigationTimeout *float64
}

func (t *timeoutSettings) SetDefaultTimeout(timeout *float64) {
	t.Lock()
	defer t.Unlock()
	t.defaultTimeout = timeout
}

func (t *timeoutSettings) DefaultTimeout() *float64 {
	t.RLock()
	defer t.RUnlock()
	return t.defaultTimeout
}

func (t *timeoutSettings) Timeout(timeout ...float64) float64 {
	t.RLock()
	defer t.RUnlock()
	if len(timeout) == 1 {
		return timeout[0]
	}
	if t.defaultTimeout != nil {
		return *t.defaultTimeout
	}
	if t.parent != nil {
		return t.parent.Timeout()
	}
	return defaultTimeout
}

func (t *timeoutSettings) DefaultNavigationTimeout() *float64 {
	t.RLock()
	defer t.RUnlock()
	return t.defaultNavigationTimeout
}

func (t *timeoutSettings) SetDefaultNavigationTimeout(navigationTimeout *float64) {
	t.Lock()
	defer t.Unlock()
	t.defaultNavigationTimeout = navigationTimeout
}

func (t *timeoutSettings) NavigationTimeout() float64 {
	t.RLock()
	defer t.RUnlock()
	if t.defaultNavigationTimeout != nil {
		return *t.defaultNavigationTimeout
	}
	if t.parent != nil {
		return t.parent.NavigationTimeout()
	}
	return defaultTimeout
}

func newTimeoutSettings(parent *timeoutSettings) *timeoutSettings {
	return &timeoutSettings{
		parent:                   parent,
		defaultTimeout:           nil,
		defaultNavigationTimeout: nil,
	}
}

// SelectOptionValues is the option struct for ElementHandle.Select() etc.
type SelectOptionValues struct {
	ValuesOrLabels *[]string
	Values         *[]string
	Indexes        *[]int
	Labels         *[]string
	Elements       *[]ElementHandle
}

func convertSelectOptionSet(values SelectOptionValues) map[string]any {
	out := make(map[string]any)
	if values == (SelectOptionValues{}) {
		return out
	}

	var o []map[string]any
	if values.ValuesOrLabels != nil {
		for _, v := range *values.ValuesOrLabels {
			m := map[string]any{"valueOrLabel": v}
			o = append(o, m)
		}
	}
	if values.Values != nil {
		for _, v := range *values.Values {
			m := map[string]any{"value": v}
			o = append(o, m)
		}
	}
	if values.Indexes != nil {
		for _, i := range *values.Indexes {
			m := map[string]any{"index": i}
			o = append(o, m)
		}
	}
	if values.Labels != nil {
		for _, l := range *values.Labels {
			m := map[string]any{"label": l}
			o = append(o, m)
		}
	}
	if o != nil {
		out["options"] = o
	}

	var e []*channel
	if values.Elements != nil {
		for _, eh := range *values.Elements {
			e = append(e, eh.(*elementHandleImpl).channel)
		}
	}
	if e != nil {
		out["elements"] = e
	}

	return out
}

func unroute(inRoutes []*routeHandlerEntry, url any, handlers ...routeHandler) ([]*routeHandlerEntry, []*routeHandlerEntry, error) {
	var handler routeHandler
	if len(handlers) == 1 {
		handler = handlers[0]
	}
	handlerPtr := reflect.ValueOf(handler).Pointer()

	removed := make([]*routeHandlerEntry, 0)
	remaining := make([]*routeHandlerEntry, 0)

	for _, route := range inRoutes {
		routeHandlerPtr := reflect.ValueOf(route.handler).Pointer()
		// note: compare regex expression if url is a regexp, not pointer
		if !route.matcher.SameWith(url) ||
			(handler != nil && routeHandlerPtr != handlerPtr) {
			remaining = append(remaining, route)
		} else {
			removed = append(removed, route)
		}
	}

	return removed, remaining, nil
}

func serializeMapToNameAndValue(headers map[string]string) []map[string]string {
	serialized := make([]map[string]string, 0)
	for name, value := range headers {
		serialized = append(serialized, map[string]string{
			"name":  name,
			"value": value,
		})
	}
	return serialized
}

// assignStructFields assigns fields from src to dest,
//
//	omitExtra determines whether to omit src's extra fields
func assignStructFields(dest, src any, omitExtra bool) error {
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Pointer || destValue.IsNil() {
		return fmt.Errorf("dest must be a non-nil pointer")
	}
	destValue = destValue.Elem()
	if destValue.Kind() != reflect.Struct {
		return fmt.Errorf("dest must be a struct")
	}

	srcValue := reflect.ValueOf(src)
	if srcValue.Kind() == reflect.Pointer {
		srcValue = srcValue.Elem()
	}
	if srcValue.Kind() != reflect.Struct {
		return fmt.Errorf("src must be a struct")
	}

	for i := 0; i < destValue.NumField(); i++ {
		destField := destValue.Field(i)
		destFieldType := destField.Type()
		destFieldName := destValue.Type().Field(i).Name

		if srcField := srcValue.FieldByName(destFieldName); srcField.IsValid() && srcField.Type() != destFieldType {
			return fmt.Errorf("mismatched field type for field %s", destFieldName)
		} else if srcField.IsValid() {
			destField.Set(srcField)
		}
	}

	if !omitExtra {
		for i := 0; i < srcValue.NumField(); i++ {
			srcFieldName := srcValue.Type().Field(i).Name

			if destField := destValue.FieldByName(srcFieldName); !destField.IsValid() {
				return fmt.Errorf("extra field %s in src", srcFieldName)
			}
		}
	}

	return nil
}

func deserializeNameAndValueToMap(headersArray []map[string]string) map[string]string {
	unserialized := make(map[string]string)
	for _, item := range headersArray {
		unserialized[item["name"]] = item["value"]
	}
	return unserialized
}

type recordHarOptions struct {
	Path           string            `json:"path"`
	Content        *HarContentPolicy `json:"content,omitempty"`
	Mode           *HarMode          `json:"mode,omitempty"`
	UrlGlob        *string           `json:"urlGlob,omitempty"`
	UrlRegexSource *string           `json:"urlRegexSource,omitempty"`
	UrlRegexFlags  *string           `json:"urlRegexFlags,omitempty"`
}

type recordHarInputOptions struct {
	Path        string
	URL         any
	Mode        *HarMode
	Content     *HarContentPolicy
	OmitContent *bool
}

type harRecordingMetadata struct {
	Path    string
	Content *HarContentPolicy
}

func prepareRecordHarOptions(option recordHarInputOptions) recordHarOptions {
	out := recordHarOptions{
		Path: option.Path,
	}
	if option.URL != nil {
		switch option.URL.(type) {
		case *regexp.Regexp:
			pattern, flags := convertRegexp(option.URL.(*regexp.Regexp))
			out.UrlRegexSource = String(pattern)
			out.UrlRegexFlags = String(flags)
		case string:
			out.UrlGlob = String(option.URL.(string))
		}
	}
	if option.Mode != nil {
		out.Mode = option.Mode
	}
	if option.Content != nil {
		out.Content = option.Content
	} else if option.OmitContent != nil && *option.OmitContent {
		out.Content = HarContentPolicyOmit
	}
	return out
}

type safeValue[T any] struct {
	sync.Mutex
	v T
}

func (s *safeValue[T]) Set(v T) {
	s.Lock()
	defer s.Unlock()
	s.v = v
}

func (s *safeValue[T]) Get() T {
	s.Lock()
	defer s.Unlock()
	return s.v
}
