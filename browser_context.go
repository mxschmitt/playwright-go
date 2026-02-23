package playwright

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/playwright-community/playwright-go/internal/safe"
)

type browserContextImpl struct {
	channelOwner
	timeoutSettings *timeoutSettings
	closeWasCalled  atomic.Bool
	options         *BrowserNewContextOptions
	pages           []Page
	routes          []*routeHandlerEntry
	webSocketRoutes []*webSocketRouteHandler
	ownedPage       Page
	browser         *browserImpl
	serviceWorkers  []Worker
	backgroundPages []Page
	bindings        *safe.SyncMap[string, BindingCallFunction]
	tracing         *tracingImpl
	request         *apiRequestContextImpl
	harRecorders    map[string]harRecordingMetadata
	closed          chan struct{}
	closeReason     *string
	harRouters      []*harRouter
	clock           Clock
}

func (b *browserContextImpl) Clock() Clock {
	return b.clock
}

func (b *browserContextImpl) SetDefaultNavigationTimeout(timeout float64) {
	b.setDefaultNavigationTimeoutImpl(&timeout)
}

func (b *browserContextImpl) setDefaultNavigationTimeoutImpl(timeout *float64) {
	b.timeoutSettings.SetDefaultNavigationTimeout(timeout)
	b.channel.SendNoReplyInternal("setDefaultNavigationTimeoutNoReply", map[string]any{
		"timeout": timeout,
	})
}

func (b *browserContextImpl) SetDefaultTimeout(timeout float64) {
	b.setDefaultTimeoutImpl(&timeout)
}

func (b *browserContextImpl) setDefaultTimeoutImpl(timeout *float64) {
	b.timeoutSettings.SetDefaultTimeout(timeout)
	b.channel.SendNoReplyInternal("setDefaultTimeoutNoReply", map[string]any{
		"timeout": timeout,
	})
}

func (b *browserContextImpl) Pages() []Page {
	b.Lock()
	defer b.Unlock()
	return b.pages
}

func (b *browserContextImpl) Browser() Browser {
	return b.browser
}

func (b *browserContextImpl) Tracing() Tracing {
	return b.tracing
}

func (b *browserContextImpl) NewCDPSession(page any) (CDPSession, error) {
	params := map[string]any{}

	if p, ok := page.(*pageImpl); ok {
		params["page"] = p.channel
	} else if f, ok := page.(*frameImpl); ok {
		params["frame"] = f.channel
	} else {
		return nil, fmt.Errorf("not page or frame: %v", page)
	}

	channel, err := b.channel.Send("newCDPSession", params)
	if err != nil {
		return nil, err
	}

	cdpSession := fromChannel(channel).(*cdpSessionImpl)

	return cdpSession, nil
}

func (b *browserContextImpl) NewPage() (Page, error) {
	if b.ownedPage != nil {
		return nil, errors.New("Please use browser.NewContext()")
	}
	channel, err := b.channel.Send("newPage")
	if err != nil {
		return nil, err
	}
	return fromChannel(channel).(*pageImpl), nil
}

func (b *browserContextImpl) Cookies(urls ...string) ([]Cookie, error) {
	result, err := b.channel.Send("cookies", map[string]any{
		"urls": urls,
	})
	if err != nil {
		return nil, err
	}
	cookies := make([]Cookie, len(result.([]any)))
	for i, item := range result.([]any) {
		cookie := &Cookie{}
		remapMapToStruct(item, cookie)
		cookies[i] = *cookie
	}
	return cookies, nil
}

func (b *browserContextImpl) AddCookies(cookies []OptionalCookie) error {
	_, err := b.channel.Send("addCookies", map[string]any{
		"cookies": cookies,
	})
	return err
}

func (b *browserContextImpl) ClearCookies(options ...BrowserContextClearCookiesOptions) error {
	params := map[string]any{}
	if len(options) == 1 {
		if options[0].Domain != nil {
			switch t := options[0].Domain.(type) {
			case string:
				params["domain"] = t
			case *string:
				params["domain"] = t
			case *regexp.Regexp:
				pattern, flag := convertRegexp(t)
				params["domainRegexSource"] = pattern
				params["domainRegexFlags"] = flag
			default:
				return errors.New("invalid type for domain, expected string or *regexp.Regexp")
			}
		}
		if options[0].Name != nil {
			switch t := options[0].Name.(type) {
			case string:
				params["name"] = t
			case *string:
				params["name"] = t
			case *regexp.Regexp:
				pattern, flag := convertRegexp(t)
				params["nameRegexSource"] = pattern
				params["nameRegexFlags"] = flag
			default:
				return errors.New("invalid type for name, expected string or *regexp.Regexp")
			}
		}
		if options[0].Path != nil {
			switch t := options[0].Path.(type) {
			case string:
				params["path"] = t
			case *string:
				params["path"] = t
			case *regexp.Regexp:
				pattern, flag := convertRegexp(t)
				params["pathRegexSource"] = pattern
				params["pathRegexFlags"] = flag
			default:
				return errors.New("invalid type for path, expected string or *regexp.Regexp")
			}
		}
	}
	_, err := b.channel.Send("clearCookies", params)
	return err
}

func (b *browserContextImpl) GrantPermissions(permissions []string, options ...BrowserContextGrantPermissionsOptions) error {
	_, err := b.channel.Send("grantPermissions", map[string]any{
		"permissions": permissions,
	}, options)
	return err
}

func (b *browserContextImpl) ClearPermissions() error {
	_, err := b.channel.Send("clearPermissions")
	return err
}

func (b *browserContextImpl) SetGeolocation(geolocation *Geolocation) error {
	_, err := b.channel.Send("setGeolocation", map[string]any{
		"geolocation": geolocation,
	})
	return err
}

func (b *browserContextImpl) ResetGeolocation() error {
	_, err := b.channel.Send("setGeolocation", map[string]any{})
	return err
}

func (b *browserContextImpl) SetExtraHTTPHeaders(headers map[string]string) error {
	_, err := b.channel.Send("setExtraHTTPHeaders", map[string]any{
		"headers": serializeMapToNameAndValue(headers),
	})
	return err
}

func (b *browserContextImpl) SetOffline(offline bool) error {
	_, err := b.channel.Send("setOffline", map[string]any{
		"offline": offline,
	})
	return err
}

func (b *browserContextImpl) AddInitScript(script Script) error {
	var source string
	if script.Content != nil {
		source = *script.Content
	}
	if script.Path != nil {
		content, err := os.ReadFile(*script.Path)
		if err != nil {
			return err
		}
		source = string(content)
	}
	_, err := b.channel.Send("addInitScript", map[string]any{
		"source": source,
	})
	return err
}

func (b *browserContextImpl) ExposeBinding(name string, binding BindingCallFunction, handle ...bool) error {
	needsHandle := false
	if len(handle) == 1 {
		needsHandle = handle[0]
	}
	for _, page := range b.Pages() {
		if _, ok := page.(*pageImpl).bindings.Load(name); ok {
			return fmt.Errorf("Function '%s' has been already registered in one of the pages", name)
		}
	}
	if _, ok := b.bindings.Load(name); ok {
		return fmt.Errorf("Function '%s' has been already registered", name)
	}
	_, err := b.channel.Send("exposeBinding", map[string]any{
		"name":        name,
		"needsHandle": needsHandle,
	})
	if err != nil {
		return err
	}
	b.bindings.Store(name, binding)
	return err
}

func (b *browserContextImpl) ExposeFunction(name string, binding ExposedFunction) error {
	return b.ExposeBinding(name, func(source *BindingSource, args ...any) any {
		return binding(args...)
	})
}

func (b *browserContextImpl) Route(url any, handler routeHandler, times ...int) error {
	b.Lock()
	defer b.Unlock()
	b.routes = slices.Insert(b.routes, 0, newRouteHandlerEntry(newURLMatcher(url, b.options.BaseURL), handler, times...))
	return b.updateInterceptionPatterns()
}

func (b *browserContextImpl) Unroute(url any, handlers ...routeHandler) error {
	removed, remaining, err := unroute(b.routes, url, handlers...)
	if err != nil {
		return err
	}
	return b.unrouteInternal(removed, remaining, UnrouteBehaviorDefault)
}

func (b *browserContextImpl) unrouteInternal(removed []*routeHandlerEntry, remaining []*routeHandlerEntry, behavior *UnrouteBehavior) error {
	b.Lock()
	defer b.Unlock()
	b.routes = remaining
	if err := b.updateInterceptionPatterns(); err != nil {
		return err
	}
	if behavior == nil || behavior == UnrouteBehaviorDefault {
		return nil
	}
	wg := &sync.WaitGroup{}
	for _, entry := range removed {
		wg.Add(1)
		go func(entry *routeHandlerEntry) {
			defer wg.Done()
			entry.Stop(string(*behavior))
		}(entry)
	}
	wg.Wait()
	return nil
}

func (b *browserContextImpl) UnrouteAll(options ...BrowserContextUnrouteAllOptions) error {
	var behavior *UnrouteBehavior
	if len(options) == 1 {
		behavior = options[0].Behavior
	}
	defer b.disposeHarRouters()
	return b.unrouteInternal(b.routes, []*routeHandlerEntry{}, behavior)
}

func (b *browserContextImpl) disposeHarRouters() {
	for _, router := range b.harRouters {
		router.dispose()
	}
	b.harRouters = make([]*harRouter, 0)
}

func (b *browserContextImpl) Request() APIRequestContext {
	return b.request
}

func (b *browserContextImpl) RouteFromHAR(har string, options ...BrowserContextRouteFromHAROptions) error {
	opt := BrowserContextRouteFromHAROptions{}
	if len(options) == 1 {
		opt = options[0]
	}
	if opt.Update != nil && *opt.Update {
		var updateContent *HarContentPolicy
		switch opt.UpdateContent {
		case RouteFromHarUpdateContentPolicyAttach:
			updateContent = HarContentPolicyAttach
		case RouteFromHarUpdateContentPolicyEmbed:
			updateContent = HarContentPolicyEmbed
		}
		return b.recordIntoHar(har, browserContextRecordIntoHarOptions{
			URL:           opt.URL,
			UpdateContent: updateContent,
			UpdateMode:    opt.UpdateMode,
		})
	}
	notFound := opt.NotFound
	if notFound == nil {
		notFound = HarNotFoundAbort
	}
	router := newHarRouter(b.connection.localUtils, har, *notFound, opt.URL)
	b.harRouters = append(b.harRouters, router)
	return router.addContextRoute(b)
}

func (b *browserContextImpl) WaitForEvent(event string, options ...BrowserContextWaitForEventOptions) (any, error) {
	return b.waiterForEvent(event, options...).Wait()
}

func (b *browserContextImpl) waiterForEvent(event string, options ...BrowserContextWaitForEventOptions) *waiter {
	timeout := b.timeoutSettings.Timeout()
	var predicate any = nil
	if len(options) == 1 {
		if options[0].Timeout != nil {
			timeout = *options[0].Timeout
		}
		predicate = options[0].Predicate
	}
	waiter := newWaiter().WithTimeout(timeout)
	waiter.RejectOnEvent(b, "close", ErrTargetClosed)
	return waiter.WaitForEvent(b, event, predicate)
}

func (b *browserContextImpl) ExpectConsoleMessage(cb func() error, options ...BrowserContextExpectConsoleMessageOptions) (ConsoleMessage, error) {
	var w *waiter
	if len(options) == 1 {
		w = b.waiterForEvent("console", BrowserContextWaitForEventOptions{
			Predicate: options[0].Predicate,
			Timeout:   options[0].Timeout,
		})
	} else {
		w = b.waiterForEvent("console")
	}
	ret, err := w.RunAndWait(cb)
	if err != nil {
		return nil, err
	}
	return ret.(ConsoleMessage), nil
}

func (b *browserContextImpl) ExpectEvent(event string, cb func() error, options ...BrowserContextExpectEventOptions) (any, error) {
	if len(options) == 1 {
		return b.waiterForEvent(event, BrowserContextWaitForEventOptions(options[0])).RunAndWait(cb)
	}
	return b.waiterForEvent(event).RunAndWait(cb)
}

func (b *browserContextImpl) ExpectPage(cb func() error, options ...BrowserContextExpectPageOptions) (Page, error) {
	var w *waiter
	if len(options) == 1 {
		w = b.waiterForEvent("page", BrowserContextWaitForEventOptions{
			Predicate: options[0].Predicate,
			Timeout:   options[0].Timeout,
		})
	} else {
		w = b.waiterForEvent("page")
	}
	ret, err := w.RunAndWait(cb)
	if err != nil {
		return nil, err
	}
	return ret.(Page), nil
}

func (b *browserContextImpl) Close(options ...BrowserContextCloseOptions) error {
	if b.closeWasCalled.Load() {
		return nil
	}
	if len(options) == 1 {
		b.closeReason = options[0].Reason
	}
	b.closeWasCalled.Store(true)

	_, err := b.channel.connection.WrapAPICall(func() (any, error) {
		return nil, b.request.Dispose(APIRequestContextDisposeOptions{
			Reason: b.closeReason,
		})
	}, true)
	if err != nil {
		return err
	}

	innerClose := func() (any, error) {
		for harId, harMetaData := range b.harRecorders {
			overrides := map[string]any{}
			if harId != "" {
				overrides["harId"] = harId
			}
			response, err := b.channel.Send("harExport", overrides)
			if err != nil {
				return nil, err
			}
			artifact := fromChannel(response).(*artifactImpl)
			// Server side will compress artifact if content is attach or if file is .zip.
			needCompressed := strings.HasSuffix(strings.ToLower(harMetaData.Path), ".zip")
			if !needCompressed && harMetaData.Content == HarContentPolicyAttach {
				tmpPath := harMetaData.Path + ".tmp"
				if err := artifact.SaveAs(tmpPath); err != nil {
					return nil, err
				}
				err = b.connection.localUtils.HarUnzip(tmpPath, harMetaData.Path)
				if err != nil {
					return nil, err
				}
			} else {
				if err := artifact.SaveAs(harMetaData.Path); err != nil {
					return nil, err
				}
			}
			if err := artifact.Delete(); err != nil {
				return nil, err
			}
		}
		return nil, nil
	}

	_, err = b.channel.connection.WrapAPICall(innerClose, true)
	if err != nil {
		return err
	}

	_, err = b.channel.Send("close", map[string]any{
		"reason": b.closeReason,
	})
	if err != nil {
		return err
	}
	<-b.closed
	return err
}

type browserContextRecordIntoHarOptions struct {
	Page          Page
	URL           any
	UpdateContent *HarContentPolicy
	UpdateMode    *HarMode
}

func (b *browserContextImpl) recordIntoHar(har string, options ...browserContextRecordIntoHarOptions) error {
	overrides := map[string]any{}
	harOptions := recordHarInputOptions{
		Path:    har,
		Content: HarContentPolicyAttach,
		Mode:    HarModeMinimal,
	}
	if len(options) == 1 {
		if options[0].UpdateContent != nil {
			harOptions.Content = options[0].UpdateContent
		}
		if options[0].UpdateMode != nil {
			harOptions.Mode = options[0].UpdateMode
		}
		harOptions.URL = options[0].URL
		overrides["options"] = prepareRecordHarOptions(harOptions)
		if options[0].Page != nil {
			overrides["page"] = options[0].Page.(*pageImpl).channel
		}
	}
	harId, err := b.channel.Send("harStart", overrides)
	if err != nil {
		return err
	}
	b.harRecorders[harId.(string)] = harRecordingMetadata{
		Path:    har,
		Content: harOptions.Content,
	}
	return nil
}

func (b *browserContextImpl) StorageState(paths ...string) (*StorageState, error) {
	result, err := b.channel.SendReturnAsDict("storageState")
	if err != nil {
		return nil, err
	}
	if len(paths) == 1 {
		file, err := os.Create(paths[0])
		if err != nil {
			return nil, err
		}
		if err := json.NewEncoder(file).Encode(result); err != nil {
			return nil, err
		}
		if err := file.Close(); err != nil {
			return nil, err
		}
	}
	var storageState StorageState
	remapMapToStruct(result, &storageState)
	return &storageState, nil
}

func (b *browserContextImpl) onBinding(binding *bindingCallImpl) {
	function, ok := b.bindings.Load(binding.initializer["name"].(string))
	if !ok || function == nil {
		return
	}
	go binding.Call(function)
}

func (b *browserContextImpl) onClose() {
	if b.browser != nil {
		contexts := make([]BrowserContext, 0)
		b.browser.Lock()
		for _, context := range b.browser.contexts {
			if context != b {
				contexts = append(contexts, context)
			}
		}
		b.browser.contexts = contexts
		b.browser.Unlock()
	}
	b.disposeHarRouters()
	b.Emit("close", b)
}

func (b *browserContextImpl) onPage(page Page) {
	b.Lock()
	b.pages = append(b.pages, page)
	b.Unlock()
	b.Emit("page", page)
	opener, _ := page.Opener()
	if opener != nil && !opener.IsClosed() {
		opener.Emit("popup", page)
	}
}

func (b *browserContextImpl) onRoute(route *routeImpl) {
	b.Lock()
	route.context = b
	page := route.Request().(*requestImpl).safePage()
	routes := make([]*routeHandlerEntry, len(b.routes))
	copy(routes, b.routes)
	b.Unlock()

	checkInterceptionIfNeeded := func() {
		b.Lock()
		defer b.Unlock()
		if len(b.routes) == 0 {
			_, err := b.connection.WrapAPICall(func() (any, error) {
				err := b.updateInterceptionPatterns()
				return nil, err
			}, true)
			if err != nil {
				logger.Error("could not update interception patterns", "error", err)
			}
		}
	}

	url := route.Request().URL()
	for _, handlerEntry := range routes {
		// If the page or the context was closed we stall all requests right away.
		if (page != nil && page.closeWasCalled.Load()) || b.closeWasCalled.Load() {
			return
		}
		if !handlerEntry.Matches(url) {
			continue
		}
		if !slices.ContainsFunc(b.routes, func(entry *routeHandlerEntry) bool {
			return entry == handlerEntry
		}) {
			continue
		}
		if handlerEntry.WillExceed() {
			b.routes = slices.DeleteFunc(b.routes, func(rhe *routeHandlerEntry) bool {
				return rhe == handlerEntry
			})
		}
		handled := handlerEntry.Handle(route)
		checkInterceptionIfNeeded()
		yes := <-handled
		if yes {
			return
		}
	}
	// If the page is closed or unrouteAll() was called without waiting and interception disabled,
	// the method will throw an error - silence it.
	_ = route.internalContinue(true)
}

func (b *browserContextImpl) updateInterceptionPatterns() error {
	patterns := prepareInterceptionPatterns(b.routes)
	_, err := b.channel.Send("setNetworkInterceptionPatterns", map[string]any{
		"patterns": patterns,
	})
	return err
}

func (b *browserContextImpl) pause() <-chan error {
	ret := make(chan error, 1)
	go func() {
		_, err := b.channel.Send("pause")
		ret <- err
	}()
	return ret
}

func (b *browserContextImpl) onBackgroundPage(ev map[string]any) {
	b.Lock()
	p := fromChannel(ev["page"]).(*pageImpl)
	p.browserContext = b
	b.backgroundPages = append(b.backgroundPages, p)
	b.Unlock()
	b.Emit("backgroundpage", p)
}

func (b *browserContextImpl) onServiceWorker(worker *workerImpl) {
	worker.context = b
	b.serviceWorkers = append(b.serviceWorkers, worker)
	b.Emit("serviceworker", worker)
}

func (b *browserContextImpl) setOptions(options *BrowserNewContextOptions, tracesDir *string) {
	if options == nil {
		options = &BrowserNewContextOptions{}
	}
	b.options = options
	if tracesDir != nil {
		b.tracing.tracesDir = *tracesDir
	}
}

// initializeHarFromOptions starts HAR recording if RecordHarPath is set in options.
// This must be called after context creation to properly register the HAR recorder on the server.
func (b *browserContextImpl) initializeHarFromOptions() error {
	if b.options == nil || b.options.RecordHarPath == nil {
		return nil
	}
	path := *b.options.RecordHarPath
	// Determine default content policy based on file extension
	var content *HarContentPolicy
	if strings.HasSuffix(strings.ToLower(path), ".zip") {
		content = HarContentPolicyAttach
	} else {
		content = HarContentPolicyEmbed
	}
	if b.options.RecordHarContent != nil {
		content = b.options.RecordHarContent
	} else if b.options.RecordHarOmitContent != nil && *b.options.RecordHarOmitContent {
		content = HarContentPolicyOmit
	}
	mode := HarModeFull
	if b.options.RecordHarMode != nil {
		mode = b.options.RecordHarMode
	}
	return b.recordIntoHar(path, browserContextRecordIntoHarOptions{
		URL:           b.options.RecordHarURLFilter,
		UpdateContent: content,
		UpdateMode:    mode,
	})
}

func (b *browserContextImpl) BackgroundPages() []Page {
	b.Lock()
	defer b.Unlock()
	return b.backgroundPages
}

func (b *browserContextImpl) ServiceWorkers() []Worker {
	b.Lock()
	defer b.Unlock()
	return b.serviceWorkers
}

func (b *browserContextImpl) OnBackgroundPage(fn func(Page)) {
	b.On("backgroundpage", fn)
}

func (b *browserContextImpl) OnClose(fn func(BrowserContext)) {
	b.On("close", fn)
}

func (b *browserContextImpl) OnConsole(fn func(ConsoleMessage)) {
	b.On("console", fn)
}

func (b *browserContextImpl) OnDialog(fn func(Dialog)) {
	b.On("dialog", fn)
}

func (b *browserContextImpl) OnPage(fn func(Page)) {
	b.On("page", fn)
}

func (b *browserContextImpl) OnRequest(fn func(Request)) {
	b.On("request", fn)
}

func (b *browserContextImpl) OnRequestFailed(fn func(Request)) {
	b.On("requestfailed", fn)
}

func (b *browserContextImpl) OnRequestFinished(fn func(Request)) {
	b.On("requestfinished", fn)
}

func (b *browserContextImpl) OnResponse(fn func(Response)) {
	b.On("response", fn)
}

func (b *browserContextImpl) OnWebError(fn func(WebError)) {
	b.On("weberror", fn)
}

func (b *browserContextImpl) RouteWebSocket(url any, handler func(WebSocketRoute)) error {
	b.Lock()
	defer b.Unlock()
	b.webSocketRoutes = slices.Insert(b.webSocketRoutes, 0, newWebSocketRouteHandler(newURLMatcher(url, b.options.BaseURL), handler))

	return b.updateWebSocketInterceptionPatterns()
}

func (b *browserContextImpl) onWebSocketRoute(wr WebSocketRoute) {
	b.Lock()
	index := slices.IndexFunc(b.webSocketRoutes, func(r *webSocketRouteHandler) bool {
		return r.Matches(wr.URL())
	})
	if index == -1 {
		b.Unlock()
		_, err := wr.ConnectToServer()
		if err != nil {
			logger.Error("could not connect to WebSocket server", "error", err)
		}
		return
	}
	handler := b.webSocketRoutes[index]
	b.Unlock()
	handler.Handle(wr)
}

func (b *browserContextImpl) updateWebSocketInterceptionPatterns() error {
	patterns := prepareWebSocketRouteHandlerInterceptionPatterns(b.webSocketRoutes)
	_, err := b.channel.Send("setWebSocketInterceptionPatterns", map[string]any{
		"patterns": patterns,
	})
	return err
}

func (b *browserContextImpl) effectiveCloseReason() *string {
	b.Lock()
	defer b.Unlock()
	if b.closeReason != nil {
		return b.closeReason
	}
	if b.browser != nil {
		return b.browser.closeReason
	}
	return nil
}

func newBrowserContext(parent *channelOwner, objectType string, guid string, initializer map[string]any) *browserContextImpl {
	bt := &browserContextImpl{
		timeoutSettings: newTimeoutSettings(nil),
		pages:           make([]Page, 0),
		backgroundPages: make([]Page, 0),
		routes:          make([]*routeHandlerEntry, 0),
		bindings:        safe.NewSyncMap[string, BindingCallFunction](),
		harRecorders:    make(map[string]harRecordingMetadata),
		closed:          make(chan struct{}, 1),
		harRouters:      make([]*harRouter, 0),
	}
	bt.createChannelOwner(bt, parent, objectType, guid, initializer)
	if parent.objectType == "Browser" {
		bt.browser = fromChannel(parent.channel).(*browserImpl)
		bt.browser.contexts = append(bt.browser.contexts, bt)
	}
	bt.tracing = fromChannel(initializer["tracing"]).(*tracingImpl)
	bt.request = fromChannel(initializer["requestContext"]).(*apiRequestContextImpl)
	bt.clock = newClock(bt)

	// Register this context with the selectors manager for custom selector engines
	if bt.browser != nil && bt.browser.browserType != nil {
		if browserType, ok := bt.browser.browserType.(*browserTypeImpl); ok && browserType.playwright != nil {
			browserType.playwright.Selectors.(*selectorsImpl).addContext(bt)
		}
	}

	bt.channel.On("bindingCall", func(params map[string]any) {
		bt.onBinding(fromChannel(params["binding"]).(*bindingCallImpl))
	})

	bt.channel.On("close", func() {
		// Unregister this context from the selectors manager
		if bt.browser != nil && bt.browser.browserType != nil {
			if browserType, ok := bt.browser.browserType.(*browserTypeImpl); ok && browserType.playwright != nil {
				browserType.playwright.Selectors.(*selectorsImpl).removeContext(bt)
			}
		}
		bt.onClose()
	})
	bt.channel.On("page", func(payload map[string]any) {
		bt.onPage(fromChannel(payload["page"]).(*pageImpl))
	})
	bt.channel.On("route", func(params map[string]any) {
		bt.channel.CreateTask(func() {
			bt.onRoute(fromChannel(params["route"]).(*routeImpl))
		})
	})
	bt.channel.On("webSocketRoute", func(params map[string]any) {
		bt.channel.CreateTask(func() {
			bt.onWebSocketRoute(fromChannel(params["webSocketRoute"]).(*webSocketRouteImpl))
		})
	})
	bt.channel.On("backgroundPage", bt.onBackgroundPage)
	bt.channel.On("serviceWorker", func(params map[string]any) {
		bt.onServiceWorker(fromChannel(params["worker"]).(*workerImpl))
	})
	bt.channel.On("console", func(ev map[string]any) {
		message := newConsoleMessage(ev)
		bt.Emit("console", message)
		if message.page != nil {
			message.page.Emit("console", message)
		}
	})
	bt.channel.On("dialog", func(params map[string]any) {
		dialog := fromChannel(params["dialog"]).(*dialogImpl)
		go func() {
			hasListeners := bt.Emit("dialog", dialog)
			page := dialog.page
			if page != nil {
				if page.Emit("dialog", dialog) {
					hasListeners = true
				}
			}
			if !hasListeners {
				// Although we do similar handling on the server side, we still need this logic
				// on the client side due to a possible race condition between two async calls:
				// a) removing "dialog" listener subscription (client->server)
				// b) actual "dialog" event (server->client)
				if dialog.Type() == "beforeunload" {
					_ = dialog.Accept()
				} else {
					_ = dialog.Dismiss()
				}
			}
		}()
	})
	bt.channel.On(
		"pageError", func(ev map[string]any) {
			pwErr := &Error{}
			remapMapToStruct(ev["error"].(map[string]any)["error"], pwErr)
			err := parseError(*pwErr)
			page := fromNullableChannel(ev["page"])
			if page != nil {
				bt.Emit("weberror", newWebError(page.(*pageImpl), err))
				page.(*pageImpl).Emit("pageerror", err)
			} else {
				bt.Emit("weberror", newWebError(nil, err))
			}
		},
	)
	bt.channel.On("request", func(ev map[string]any) {
		request := fromChannel(ev["request"]).(*requestImpl)
		page := fromNullableChannel(ev["page"])
		bt.Emit("request", request)
		if page != nil {
			page.(*pageImpl).Emit("request", request)
		}
	})
	bt.channel.On("requestFailed", func(ev map[string]any) {
		request := fromChannel(ev["request"]).(*requestImpl)
		failureText := ev["failureText"]
		if failureText != nil {
			request.failureText = failureText.(string)
		}
		page := fromNullableChannel(ev["page"])
		request.setResponseEndTiming(ev["responseEndTiming"].(float64))
		bt.Emit("requestfailed", request)
		if page != nil {
			page.(*pageImpl).Emit("requestfailed", request)
		}
	})

	bt.channel.On("requestFinished", func(ev map[string]any) {
		request := fromChannel(ev["request"]).(*requestImpl)
		response := fromNullableChannel(ev["response"])
		page := fromNullableChannel(ev["page"])
		request.setResponseEndTiming(ev["responseEndTiming"].(float64))
		bt.Emit("requestfinished", request)
		if page != nil {
			page.(*pageImpl).Emit("requestfinished", request)
		}
		if response != nil {
			close(response.(*responseImpl).finished)
		}
	})
	bt.channel.On("response", func(ev map[string]any) {
		response := fromChannel(ev["response"]).(*responseImpl)
		page := fromNullableChannel(ev["page"])
		bt.Emit("response", response)
		if page != nil {
			page.(*pageImpl).Emit("response", response)
		}
	})
	bt.Once("close", func() {
		bt.closed <- struct{}{}
	})
	bt.setEventSubscriptionMapping(map[string]string{
		"console":         "console",
		"dialog":          "dialog",
		"request":         "request",
		"response":        "response",
		"requestfinished": "requestFinished",
		"responsefailed":  "responseFailed",
	})
	return bt
}
