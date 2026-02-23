package playwright

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/playwright-community/playwright-go/internal/safe"
)

type pageImpl struct {
	channelOwner
	isClosed        bool
	closedOrCrashed chan error
	video           *videoImpl
	mouse           *mouseImpl
	keyboard        *keyboardImpl
	touchscreen     *touchscreenImpl
	timeoutSettings *timeoutSettings
	browserContext  *browserContextImpl
	frames          []Frame
	workers         []Worker
	mainFrame       Frame
	routes          []*routeHandlerEntry
	webSocketRoutes []*webSocketRouteHandler
	viewportSize    *Size
	ownedContext    BrowserContext
	bindings        *safe.SyncMap[string, BindingCallFunction]
	closeReason     *string
	closeWasCalled  atomic.Bool
	harRouters      []*harRouter
	locatorHandlers map[float64]*locatorHandlerEntry
}

type locatorHandlerEntry struct {
	locator *locatorImpl
	handler func(Locator)
	times   *int
}

func (p *pageImpl) AddLocatorHandler(locator Locator, handler func(Locator), options ...PageAddLocatorHandlerOptions) error {
	if locator == nil || handler == nil {
		return errors.New("locator or handler must not be nil")
	}
	if locator.Err() != nil {
		return locator.Err()
	}

	var option PageAddLocatorHandlerOptions
	if len(options) == 1 {
		option = options[0]
		if option.Times != nil && *option.Times == 0 {
			return nil
		}
	}

	loc := locator.(*locatorImpl)
	if loc.frame != p.mainFrame {
		return errors.New("locator must belong to the main frame of this page")
	}
	uid, err := p.channel.Send("registerLocatorHandler", map[string]any{
		"selector":    loc.selector,
		"noWaitAfter": option.NoWaitAfter,
	})
	if err != nil {
		return err
	}
	p.locatorHandlers[uid.(float64)] = &locatorHandlerEntry{locator: loc, handler: handler, times: option.Times}
	return nil
}

func (p *pageImpl) onLocatorHandlerTriggered(uid float64) {
	var remove *bool
	handler, ok := p.locatorHandlers[uid]
	if !ok {
		return
	}
	if handler.times != nil {
		*handler.times--
		if *handler.times == 0 {
			remove = Bool(true)
		}
	}
	defer func() {
		if remove != nil && *remove {
			delete(p.locatorHandlers, uid)
		}
		_, _ = p.connection.WrapAPICall(func() (any, error) {
			_, err := p.channel.Send("resolveLocatorHandlerNoReply", map[string]any{
				"uid":    uid,
				"remove": remove,
			})
			return nil, err
		}, true)
	}()

	handler.handler(handler.locator)
}

func (p *pageImpl) RemoveLocatorHandler(locator Locator) error {
	for uid := range p.locatorHandlers {
		if p.locatorHandlers[uid].locator.equals(locator) {
			delete(p.locatorHandlers, uid)
			p.channel.SendNoReply("unregisterLocatorHandler", map[string]any{
				"uid": uid,
			})
			return nil
		}
	}
	return nil
}

func (p *pageImpl) Context() BrowserContext {
	return p.browserContext
}

func (b *pageImpl) Clock() Clock {
	return b.browserContext.clock
}

func (p *pageImpl) Close(options ...PageCloseOptions) error {
	if len(options) == 1 {
		p.closeReason = options[0].Reason
	}
	p.closeWasCalled.Store(true)
	_, err := p.channel.Send("close", options)
	if err == nil && p.ownedContext != nil {
		err = p.ownedContext.Close()
	}
	if errors.Is(err, ErrTargetClosed) || (len(options) == 1 && options[0].RunBeforeUnload != nil && *(options[0].RunBeforeUnload)) {
		return nil
	}
	return err
}

func (p *pageImpl) InnerText(selector string, options ...PageInnerTextOptions) (string, error) {
	if len(options) == 1 {
		return p.mainFrame.InnerText(selector, FrameInnerTextOptions(options[0]))
	}
	return p.mainFrame.InnerText(selector)
}

func (p *pageImpl) InnerHTML(selector string, options ...PageInnerHTMLOptions) (string, error) {
	if len(options) == 1 {
		return p.mainFrame.InnerHTML(selector, FrameInnerHTMLOptions(options[0]))
	}
	return p.mainFrame.InnerHTML(selector)
}

func (p *pageImpl) Opener() (Page, error) {
	channel := p.initializer["opener"]
	channelOwner := fromNullableChannel(channel)
	if channelOwner == nil {
		// not popup page or opener has been closed
		return nil, nil
	}
	return channelOwner.(*pageImpl), nil
}

func (p *pageImpl) MainFrame() Frame {
	return p.mainFrame
}

func (p *pageImpl) Frame(options ...PageFrameOptions) Frame {
	option := PageFrameOptions{}
	if len(options) == 1 {
		option = options[0]
	}
	var matcher *urlMatcher
	if option.URL != nil {
		matcher = newURLMatcher(option.URL, p.browserContext.options.BaseURL)
	}

	for _, f := range p.frames {
		if option.Name != nil && f.Name() == *option.Name {
			return f
		}

		if option.URL != nil && matcher != nil && matcher.Matches(f.URL()) {
			return f
		}
	}

	return nil
}

func (p *pageImpl) Frames() []Frame {
	return p.frames
}

func (p *pageImpl) SetDefaultNavigationTimeout(timeout float64) {
	p.timeoutSettings.SetDefaultNavigationTimeout(&timeout)
	p.channel.SendNoReplyInternal("setDefaultNavigationTimeoutNoReply", map[string]any{
		"timeout": timeout,
	})
}

func (p *pageImpl) SetDefaultTimeout(timeout float64) {
	p.timeoutSettings.SetDefaultTimeout(&timeout)
	p.channel.SendNoReplyInternal("setDefaultTimeoutNoReply", map[string]any{
		"timeout": timeout,
	})
}

func (p *pageImpl) QuerySelector(selector string, options ...PageQuerySelectorOptions) (ElementHandle, error) {
	if len(options) == 1 {
		return p.mainFrame.QuerySelector(selector, FrameQuerySelectorOptions(options[0]))
	}
	return p.mainFrame.QuerySelector(selector)
}

func (p *pageImpl) QuerySelectorAll(selector string) ([]ElementHandle, error) {
	return p.mainFrame.QuerySelectorAll(selector)
}

func (p *pageImpl) WaitForSelector(selector string, options ...PageWaitForSelectorOptions) (ElementHandle, error) {
	if len(options) == 1 {
		return p.mainFrame.WaitForSelector(selector, FrameWaitForSelectorOptions(options[0]))
	}
	return p.mainFrame.WaitForSelector(selector)
}

func (p *pageImpl) DispatchEvent(selector string, typ string, eventInit any, options ...PageDispatchEventOptions) error {
	if len(options) == 1 {
		return p.mainFrame.DispatchEvent(selector, typ, eventInit, FrameDispatchEventOptions(options[0]))
	}
	return p.mainFrame.DispatchEvent(selector, typ, eventInit)
}

func (p *pageImpl) Evaluate(expression string, arg ...any) (any, error) {
	return p.mainFrame.Evaluate(expression, arg...)
}

func (p *pageImpl) EvaluateHandle(expression string, arg ...any) (JSHandle, error) {
	return p.mainFrame.EvaluateHandle(expression, arg...)
}

func (p *pageImpl) EvalOnSelector(selector string, expression string, arg any, options ...PageEvalOnSelectorOptions) (any, error) {
	if len(options) == 1 {
		return p.mainFrame.EvalOnSelector(selector, expression, arg, FrameEvalOnSelectorOptions(options[0]))
	}
	return p.mainFrame.EvalOnSelector(selector, expression, arg)
}

func (p *pageImpl) EvalOnSelectorAll(selector string, expression string, arg ...any) (any, error) {
	return p.mainFrame.EvalOnSelectorAll(selector, expression, arg...)
}

func (p *pageImpl) AddScriptTag(options PageAddScriptTagOptions) (ElementHandle, error) {
	return p.mainFrame.AddScriptTag(FrameAddScriptTagOptions(options))
}

func (p *pageImpl) AddStyleTag(options PageAddStyleTagOptions) (ElementHandle, error) {
	return p.mainFrame.AddStyleTag(FrameAddStyleTagOptions(options))
}

func (p *pageImpl) SetExtraHTTPHeaders(headers map[string]string) error {
	_, err := p.channel.Send("setExtraHTTPHeaders", map[string]any{
		"headers": serializeMapToNameAndValue(headers),
	})
	return err
}

func (p *pageImpl) URL() string {
	return p.mainFrame.URL()
}

func (p *pageImpl) Unroute(url any, handlers ...routeHandler) error {
	p.Lock()
	defer p.Unlock()

	removed, remaining, err := unroute(p.routes, url, handlers...)
	if err != nil {
		return err
	}
	return p.unrouteInternal(removed, remaining, UnrouteBehaviorDefault)
}

func (p *pageImpl) unrouteInternal(removed []*routeHandlerEntry, remaining []*routeHandlerEntry, behavior *UnrouteBehavior) error {
	p.routes = remaining
	err := p.updateInterceptionPatterns()
	if err != nil {
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

func (p *pageImpl) disposeHarRouters() {
	for _, router := range p.harRouters {
		router.dispose()
	}
	p.harRouters = make([]*harRouter, 0)
}

func (p *pageImpl) UnrouteAll(options ...PageUnrouteAllOptions) error {
	var behavior *UnrouteBehavior
	if len(options) == 1 {
		behavior = options[0].Behavior
	}
	p.Lock()
	defer p.Unlock()
	defer p.disposeHarRouters()
	return p.unrouteInternal(p.routes, []*routeHandlerEntry{}, behavior)
}

func (p *pageImpl) Content() (string, error) {
	return p.mainFrame.Content()
}

func (p *pageImpl) SetContent(html string, options ...PageSetContentOptions) error {
	if len(options) == 1 {
		return p.mainFrame.SetContent(html, FrameSetContentOptions(options[0]))
	}
	return p.mainFrame.SetContent(html)
}

func (p *pageImpl) Goto(url string, options ...PageGotoOptions) (Response, error) {
	if len(options) == 1 {
		return p.mainFrame.Goto(url, FrameGotoOptions(options[0]))
	}
	return p.mainFrame.Goto(url)
}

func (p *pageImpl) Reload(options ...PageReloadOptions) (Response, error) {
	channel, err := p.channel.Send("reload", options)
	if err != nil {
		return nil, err
	}
	channelOwner := fromNullableChannel(channel)
	if channelOwner == nil {
		return nil, nil
	}
	return channelOwner.(*responseImpl), nil
}

func (p *pageImpl) WaitForLoadState(options ...PageWaitForLoadStateOptions) error {
	if len(options) == 1 {
		return p.mainFrame.WaitForLoadState(FrameWaitForLoadStateOptions(options[0]))
	}
	return p.mainFrame.WaitForLoadState()
}

func (p *pageImpl) GoBack(options ...PageGoBackOptions) (Response, error) {
	channel, err := p.channel.Send("goBack", options)
	if err != nil {
		return nil, err
	}
	channelOwner := fromNullableChannel(channel)
	if channelOwner == nil {
		// can not go back
		return nil, nil
	}
	return channelOwner.(*responseImpl), nil
}

func (p *pageImpl) GoForward(options ...PageGoForwardOptions) (Response, error) {
	channel, err := p.channel.Send("goForward", options)
	if err != nil {
		return nil, err
	}
	channelOwner := fromNullableChannel(channel)
	if channelOwner == nil {
		// can not go forward
		return nil, nil
	}
	return channelOwner.(*responseImpl), nil
}

func (p *pageImpl) EmulateMedia(options ...PageEmulateMediaOptions) error {
	_, err := p.channel.Send("emulateMedia", options)
	if err != nil {
		return err
	}
	return err
}

func (p *pageImpl) SetViewportSize(width, height int) error {
	_, err := p.channel.Send("setViewportSize", map[string]any{
		"viewportSize": map[string]any{
			"width":  width,
			"height": height,
		},
	})
	if err != nil {
		return err
	}
	p.viewportSize.Width = width
	p.viewportSize.Height = height
	return nil
}

func (p *pageImpl) ViewportSize() *Size {
	return p.viewportSize
}

func (p *pageImpl) BringToFront() error {
	_, err := p.channel.Send("bringToFront")
	return err
}

func (p *pageImpl) Type(selector, text string, options ...PageTypeOptions) error {
	if len(options) == 1 {
		return p.mainFrame.Type(selector, text, FrameTypeOptions(options[0]))
	}
	return p.mainFrame.Type(selector, text)
}

func (p *pageImpl) Fill(selector, text string, options ...PageFillOptions) error {
	if len(options) == 1 {
		return p.mainFrame.Fill(selector, text, FrameFillOptions(options[0]))
	}
	return p.mainFrame.Fill(selector, text)
}

func (p *pageImpl) Press(selector, key string, options ...PagePressOptions) error {
	if len(options) == 1 {
		return p.mainFrame.Press(selector, key, FramePressOptions(options[0]))
	}
	return p.mainFrame.Press(selector, key)
}

func (p *pageImpl) Title() (string, error) {
	return p.mainFrame.Title()
}

func (p *pageImpl) Workers() []Worker {
	return p.workers
}

func (p *pageImpl) Request() APIRequestContext {
	return p.Context().Request()
}

func (p *pageImpl) Screenshot(options ...PageScreenshotOptions) ([]byte, error) {
	var path *string
	overrides := map[string]any{}
	if len(options) == 1 {
		path = options[0].Path
		options[0].Path = nil
		if options[0].Mask != nil {
			masks := make([]map[string]any, 0)
			for _, m := range options[0].Mask {
				if m.Err() != nil { // ErrLocatorNotSameFrame
					return nil, m.Err()
				}
				l, ok := m.(*locatorImpl)
				if ok {
					masks = append(masks, map[string]any{
						"selector": l.selector,
						"frame":    l.frame.channel,
					})
				}
			}
			overrides["mask"] = masks
		}
	}
	data, err := p.channel.Send("screenshot", options, overrides)
	if err != nil {
		return nil, err
	}
	image, err := base64.StdEncoding.DecodeString(data.(string))
	if err != nil {
		return nil, fmt.Errorf("could not decode base64 :%w", err)
	}
	if path != nil {
		if err := os.WriteFile(*path, image, 0o644); err != nil {
			return nil, err
		}
	}
	return image, nil
}

func (p *pageImpl) PDF(options ...PagePdfOptions) ([]byte, error) {
	var path *string
	if len(options) == 1 {
		path = options[0].Path
	}
	data, err := p.channel.Send("pdf", options)
	if err != nil {
		return nil, err
	}
	pdf, err := base64.StdEncoding.DecodeString(data.(string))
	if err != nil {
		return nil, fmt.Errorf("could not decode base64 :%w", err)
	}
	if path != nil {
		if err := os.WriteFile(*path, pdf, 0o644); err != nil {
			return nil, err
		}
	}
	return pdf, nil
}

func (p *pageImpl) Click(selector string, options ...PageClickOptions) error {
	if len(options) == 1 {
		return p.mainFrame.Click(selector, FrameClickOptions(options[0]))
	}
	return p.mainFrame.Click(selector)
}

func (p *pageImpl) WaitForEvent(event string, options ...PageWaitForEventOptions) (any, error) {
	return p.waiterForEvent(event, options...).Wait()
}

func (p *pageImpl) waiterForEvent(event string, options ...PageWaitForEventOptions) *waiter {
	timeout := p.timeoutSettings.Timeout()
	var predicate any = nil
	if len(options) == 1 {
		if options[0].Timeout != nil {
			timeout = *options[0].Timeout
		}
		predicate = options[0].Predicate
	}
	waiter := newWaiter().WithTimeout(timeout)
	waiter.RejectOnEvent(p, "close", p.closeErrorWithReason())
	waiter.RejectOnEvent(p, "crash", errors.New("page crashed"))
	return waiter.WaitForEvent(p, event, predicate)
}

func (p *pageImpl) waiterForRequest(url any, options ...PageExpectRequestOptions) *waiter {
	option := PageExpectRequestOptions{}
	if len(options) == 1 {
		option = options[0]
	}
	if option.Timeout == nil {
		option.Timeout = Float(p.timeoutSettings.Timeout())
	}
	var matcher *urlMatcher
	if url != nil {
		matcher = newURLMatcher(url, p.browserContext.options.BaseURL)
	}
	predicate := func(req *requestImpl) bool {
		if matcher != nil {
			return matcher.Matches(req.URL())
		}
		return true
	}

	waiter := newWaiter().WithTimeout(*option.Timeout)
	return waiter.WaitForEvent(p, "request", predicate)
}

func (p *pageImpl) waiterForResponse(url any, options ...PageExpectResponseOptions) *waiter {
	option := PageExpectResponseOptions{}
	if len(options) == 1 {
		option = options[0]
	}
	if option.Timeout == nil {
		option.Timeout = Float(p.timeoutSettings.Timeout())
	}
	var matcher *urlMatcher
	if url != nil {
		matcher = newURLMatcher(url, p.browserContext.options.BaseURL)
	}
	predicate := func(req *responseImpl) bool {
		if matcher != nil {
			return matcher.Matches(req.URL())
		}
		return true
	}

	waiter := newWaiter().WithTimeout(*option.Timeout)
	return waiter.WaitForEvent(p, "response", predicate)
}

func (p *pageImpl) ExpectEvent(event string, cb func() error, options ...PageExpectEventOptions) (any, error) {
	if len(options) == 1 {
		return p.waiterForEvent(event, PageWaitForEventOptions(options[0])).RunAndWait(cb)
	}
	return p.waiterForEvent(event).RunAndWait(cb)
}

func (p *pageImpl) ExpectNavigation(cb func() error, options ...PageExpectNavigationOptions) (Response, error) {
	if len(options) == 1 {
		return p.mainFrame.ExpectNavigation(cb, FrameExpectNavigationOptions(options[0]))
	}
	return p.mainFrame.ExpectNavigation(cb)
}

func (p *pageImpl) ExpectConsoleMessage(cb func() error, options ...PageExpectConsoleMessageOptions) (ConsoleMessage, error) {
	option := PageWaitForEventOptions{}
	if len(options) == 1 {
		option.Timeout = options[0].Timeout
		option.Predicate = options[0].Predicate
	}
	ret, err := p.waiterForEvent("console", option).RunAndWait(cb)
	if ret == nil {
		return nil, err
	}
	return ret.(*consoleMessageImpl), err
}

func (p *pageImpl) ExpectDownload(cb func() error, options ...PageExpectDownloadOptions) (Download, error) {
	option := PageWaitForEventOptions{}
	if len(options) == 1 {
		option.Timeout = options[0].Timeout
		option.Predicate = options[0].Predicate
	}
	ret, err := p.waiterForEvent("download", option).RunAndWait(cb)
	if ret == nil {
		return nil, err
	}
	return ret.(*downloadImpl), err
}

func (p *pageImpl) ExpectFileChooser(cb func() error, options ...PageExpectFileChooserOptions) (FileChooser, error) {
	option := PageWaitForEventOptions{}
	if len(options) == 1 {
		option.Timeout = options[0].Timeout
		option.Predicate = options[0].Predicate
	}
	ret, err := p.waiterForEvent("filechooser", option).RunAndWait(cb)
	if ret == nil {
		return nil, err
	}
	return ret.(*fileChooserImpl), err
}

func (p *pageImpl) ExpectPopup(cb func() error, options ...PageExpectPopupOptions) (Page, error) {
	option := PageWaitForEventOptions{}
	if len(options) == 1 {
		option.Timeout = options[0].Timeout
		option.Predicate = options[0].Predicate
	}
	ret, err := p.waiterForEvent("popup", option).RunAndWait(cb)
	if ret == nil {
		return nil, err
	}
	return ret.(*pageImpl), err
}

func (p *pageImpl) ExpectResponse(url any, cb func() error, options ...PageExpectResponseOptions) (Response, error) {
	ret, err := p.waiterForResponse(url, options...).RunAndWait(cb)
	if ret == nil {
		return nil, err
	}
	return ret.(*responseImpl), err
}

func (p *pageImpl) ExpectRequest(url any, cb func() error, options ...PageExpectRequestOptions) (Request, error) {
	ret, err := p.waiterForRequest(url, options...).RunAndWait(cb)
	if ret == nil {
		return nil, err
	}
	return ret.(*requestImpl), err
}

func (p *pageImpl) ExpectRequestFinished(cb func() error, options ...PageExpectRequestFinishedOptions) (Request, error) {
	option := PageWaitForEventOptions{}
	if len(options) == 1 {
		option.Timeout = options[0].Timeout
		option.Predicate = options[0].Predicate
	}
	ret, err := p.waiterForEvent("requestfinished", option).RunAndWait(cb)
	if ret == nil {
		return nil, err
	}
	return ret.(*requestImpl), err
}

func (p *pageImpl) ExpectWebSocket(cb func() error, options ...PageExpectWebSocketOptions) (WebSocket, error) {
	option := PageWaitForEventOptions{}
	if len(options) == 1 {
		option.Timeout = options[0].Timeout
		option.Predicate = options[0].Predicate
	}
	ret, err := p.waiterForEvent("websocket", option).RunAndWait(cb)
	if ret == nil {
		return nil, err
	}
	return ret.(*webSocketImpl), err
}

func (p *pageImpl) ExpectWorker(cb func() error, options ...PageExpectWorkerOptions) (Worker, error) {
	option := PageWaitForEventOptions{}
	if len(options) == 1 {
		option.Timeout = options[0].Timeout
		option.Predicate = options[0].Predicate
	}
	ret, err := p.waiterForEvent("worker", option).RunAndWait(cb)
	if ret == nil {
		return nil, err
	}
	return ret.(*workerImpl), err
}

func (p *pageImpl) Route(url any, handler routeHandler, times ...int) error {
	p.Lock()
	defer p.Unlock()
	p.routes = slices.Insert(p.routes, 0, newRouteHandlerEntry(newURLMatcher(url, p.browserContext.options.BaseURL), handler, times...))
	return p.updateInterceptionPatterns()
}

func (p *pageImpl) GetAttribute(selector string, name string, options ...PageGetAttributeOptions) (string, error) {
	if len(options) == 1 {
		return p.mainFrame.GetAttribute(selector, name, FrameGetAttributeOptions(options[0]))
	}
	return p.mainFrame.GetAttribute(selector, name)
}

func (p *pageImpl) Hover(selector string, options ...PageHoverOptions) error {
	if len(options) == 1 {
		return p.mainFrame.Hover(selector, FrameHoverOptions(options[0]))
	}
	return p.mainFrame.Hover(selector)
}

func (p *pageImpl) IsClosed() bool {
	return p.isClosed
}

func (p *pageImpl) AddInitScript(script Script) error {
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
	_, err := p.channel.Send("addInitScript", map[string]any{
		"source": source,
	})
	return err
}

func (p *pageImpl) Keyboard() Keyboard {
	return p.keyboard
}

func (p *pageImpl) ConsoleMessages() ([]ConsoleMessage, error) {
	result, err := p.channel.Send("consoleMessages", nil)
	if err != nil {
		return nil, err
	}
	messages := result.([]any)
	consoleMessages := make([]ConsoleMessage, len(messages))
	for i, m := range messages {
		consoleMessages[i] = newConsoleMessage(m.(map[string]any))
	}
	return consoleMessages, nil
}

func (p *pageImpl) Requests() ([]Request, error) {
	result, err := p.channel.Send("requests", nil)
	if err != nil {
		return nil, err
	}
	requests := result.([]any)
	reqs := make([]Request, len(requests))
	for i, r := range requests {
		reqs[i] = fromChannel(r).(*requestImpl)
	}
	return reqs, nil
}

func (p *pageImpl) Mouse() Mouse {
	return p.mouse
}

func (p *pageImpl) RouteFromHAR(har string, options ...PageRouteFromHAROptions) error {
	opt := PageRouteFromHAROptions{}
	if len(options) == 1 {
		opt = options[0]
	}
	if opt.Update != nil && *opt.Update {
		return p.browserContext.recordIntoHar(har, browserContextRecordIntoHarOptions{
			Page: p,
			URL:  opt.URL,
		})
	}
	notFound := opt.NotFound
	if notFound == nil {
		notFound = HarNotFoundAbort
	}
	router := newHarRouter(p.connection.localUtils, har, *notFound, opt.URL)
	p.harRouters = append(p.harRouters, router)
	return router.addPageRoute(p)
}

func (p *pageImpl) Touchscreen() Touchscreen {
	return p.touchscreen
}

func newPage(parent *channelOwner, objectType string, guid string, initializer map[string]any) *pageImpl {
	viewportSize := &Size{}
	if _, ok := initializer["viewportSize"].(map[string]any); ok {
		viewportSize.Height = int(initializer["viewportSize"].(map[string]any)["height"].(float64))
		viewportSize.Width = int(initializer["viewportSize"].(map[string]any)["width"].(float64))
	}
	bt := &pageImpl{
		workers:         make([]Worker, 0),
		routes:          make([]*routeHandlerEntry, 0),
		bindings:        safe.NewSyncMap[string, BindingCallFunction](),
		viewportSize:    viewportSize,
		harRouters:      make([]*harRouter, 0),
		locatorHandlers: make(map[float64]*locatorHandlerEntry, 0),
	}
	bt.createChannelOwner(bt, parent, objectType, guid, initializer)
	bt.browserContext = fromChannel(parent.channel).(*browserContextImpl)
	bt.timeoutSettings = newTimeoutSettings(bt.browserContext.timeoutSettings)
	mainframe := fromChannel(initializer["mainFrame"]).(*frameImpl)
	mainframe.page = bt
	bt.mainFrame = mainframe
	bt.frames = []Frame{mainframe}
	bt.mouse = newMouse(bt.channel)
	bt.keyboard = newKeyboard(bt.channel)
	bt.touchscreen = newTouchscreen(bt.channel)
	bt.channel.On("bindingCall", func(params map[string]any) {
		bt.onBinding(fromChannel(params["binding"]).(*bindingCallImpl))
	})
	bt.channel.On("close", bt.onClose)
	bt.channel.On("crash", func() {
		bt.Emit("crash", bt)
	})
	bt.channel.On("domcontentloaded", func() {
		bt.Emit("domcontentloaded", bt)
	})
	bt.channel.On("fileChooser", func(ev map[string]any) {
		bt.Emit("filechooser", newFileChooser(bt, fromChannel(ev["element"]).(*elementHandleImpl), ev["isMultiple"].(bool)))
	})
	bt.channel.On("frameAttached", func(ev map[string]any) {
		bt.onFrameAttached(fromChannel(ev["frame"]).(*frameImpl))
	})
	bt.channel.On("frameDetached", func(ev map[string]any) {
		bt.onFrameDetached(fromChannel(ev["frame"]).(*frameImpl))
	})
	bt.channel.On("locatorHandlerTriggered", func(ev map[string]any) {
		bt.channel.CreateTask(func() {
			bt.onLocatorHandlerTriggered(ev["uid"].(float64))
		})
	})
	bt.channel.On(
		"load", func(ev map[string]any) {
			bt.Emit("load", bt)
		},
	)
	bt.channel.On("popup", func(ev map[string]any) {
		bt.Emit("popup", fromChannel(ev["page"]).(*pageImpl))
	})
	bt.channel.On("route", func(ev map[string]any) {
		bt.channel.CreateTask(func() {
			bt.onRoute(fromChannel(ev["route"]).(*routeImpl))
		})
	})
	bt.channel.On("download", func(ev map[string]any) {
		url := ev["url"].(string)
		suggestedFilename := ev["suggestedFilename"].(string)
		artifact := fromChannel(ev["artifact"]).(*artifactImpl)
		bt.Emit("download", newDownload(bt, url, suggestedFilename, artifact))
	})
	bt.channel.On("video", func(params map[string]any) {
		artifact := fromChannel(params["artifact"]).(*artifactImpl)
		bt.Video().(*videoImpl).artifactReady(artifact)
	})
	bt.channel.On("webSocket", func(ev map[string]any) {
		bt.Emit("websocket", fromChannel(ev["webSocket"]).(*webSocketImpl))
	})
	bt.channel.On("webSocketRoute", func(ev map[string]any) {
		bt.channel.CreateTask(func() {
			bt.onWebSocketRoute(fromChannel(ev["webSocketRoute"]).(*webSocketRouteImpl))
		})
	})

	bt.channel.On("worker", func(ev map[string]any) {
		bt.onWorker(fromChannel(ev["worker"]).(*workerImpl))
	})
	bt.closedOrCrashed = make(chan error, 1)
	bt.OnClose(func(Page) {
		select {
		case bt.closedOrCrashed <- bt.closeErrorWithReason():
		default:
		}
	})
	bt.OnCrash(func(Page) {
		select {
		case bt.closedOrCrashed <- ErrTargetClosed:
		default:
		}
	})
	bt.setEventSubscriptionMapping(map[string]string{
		"console":         "console",
		"dialog":          "dialog",
		"request":         "request",
		"response":        "response",
		"requestfinished": "requestFinished",
		"responsefailed":  "responseFailed",
		"filechooser":     "fileChooser",
	})

	return bt
}

func (p *pageImpl) closeErrorWithReason() error {
	if p.closeReason != nil {
		return targetClosedError(p.closeReason)
	}
	return targetClosedError(p.browserContext.effectiveCloseReason())
}

func (p *pageImpl) onBinding(binding *bindingCallImpl) {
	function, ok := p.bindings.Load(binding.initializer["name"].(string))
	if !ok || function == nil {
		return
	}
	go binding.Call(function)
}

func (p *pageImpl) onFrameAttached(frame *frameImpl) {
	frame.page = p
	p.frames = append(p.frames, frame)
	p.Emit("frameattached", frame)
}

func (p *pageImpl) onFrameDetached(frame *frameImpl) {
	frame.detached = true
	frames := make([]Frame, 0)
	for i := 0; i < len(p.frames); i++ {
		if p.frames[i] != frame {
			frames = append(frames, frame)
		}
	}
	if len(frames) != len(p.frames) {
		p.frames = frames
	}
	p.Emit("framedetached", frame)
}

func (p *pageImpl) onRoute(route *routeImpl) {
	p.Lock()
	route.context = p.browserContext
	routes := make([]*routeHandlerEntry, len(p.routes))
	copy(routes, p.routes)
	p.Unlock()

	checkInterceptionIfNeeded := func() {
		p.Lock()
		defer p.Unlock()
		if len(p.routes) == 0 {
			_, err := p.connection.WrapAPICall(func() (any, error) {
				err := p.updateInterceptionPatterns()
				return nil, err
			}, true)
			if err != nil {
				logger.Error("could not update interception patterns", "error", err)
			}
		}
	}

	url := route.Request().URL()
	for _, handlerEntry := range routes {
		// If the page was closed we stall all requests right away.
		if p.closeWasCalled.Load() || p.browserContext.closeWasCalled.Load() {
			return
		}
		if !handlerEntry.Matches(url) {
			continue
		}
		if !slices.ContainsFunc(p.routes, func(entry *routeHandlerEntry) bool {
			return entry == handlerEntry
		}) {
			continue
		}
		if handlerEntry.WillExceed() {
			p.routes = slices.DeleteFunc(p.routes, func(rhe *routeHandlerEntry) bool {
				return rhe == handlerEntry
			})
		}
		handled := handlerEntry.Handle(route)
		checkInterceptionIfNeeded()

		if <-handled {
			return
		}
	}
	p.browserContext.onRoute(route)
}

func (p *pageImpl) updateInterceptionPatterns() error {
	patterns := prepareInterceptionPatterns(p.routes)
	_, err := p.channel.Send("setNetworkInterceptionPatterns", map[string]any{
		"patterns": patterns,
	})
	return err
}

func (p *pageImpl) onWorker(worker *workerImpl) {
	p.workers = append(p.workers, worker)
	worker.page = p
	p.Emit("worker", worker)
}

func (p *pageImpl) onClose() {
	p.isClosed = true
	newPages := []Page{}
	newBackgoundPages := []Page{}
	if p.browserContext != nil {
		p.browserContext.Lock()
		for _, page := range p.browserContext.pages {
			if page != p {
				newPages = append(newPages, page)
			}
		}
		for _, page := range p.browserContext.backgroundPages {
			if page != p {
				newBackgoundPages = append(newBackgoundPages, page)
			}
		}
		p.browserContext.pages = newPages
		p.browserContext.backgroundPages = newBackgoundPages
		p.browserContext.Unlock()
	}
	p.disposeHarRouters()
	p.Emit("close", p)
}

func (p *pageImpl) SetInputFiles(selector string, files any, options ...PageSetInputFilesOptions) error {
	if len(options) == 1 {
		return p.mainFrame.SetInputFiles(selector, files, FrameSetInputFilesOptions(options[0]))
	}
	return p.mainFrame.SetInputFiles(selector, files)
}

func (p *pageImpl) Check(selector string, options ...PageCheckOptions) error {
	if len(options) == 1 {
		return p.mainFrame.Check(selector, FrameCheckOptions(options[0]))
	}
	return p.mainFrame.Check(selector)
}

func (p *pageImpl) Uncheck(selector string, options ...PageUncheckOptions) error {
	if len(options) == 1 {
		return p.mainFrame.Uncheck(selector, FrameUncheckOptions(options[0]))
	}
	return p.mainFrame.Uncheck(selector)
}

func (p *pageImpl) WaitForTimeout(timeout float64) {
	p.mainFrame.WaitForTimeout(timeout)
}

func (p *pageImpl) WaitForFunction(expression string, arg any, options ...PageWaitForFunctionOptions) (JSHandle, error) {
	if len(options) == 1 {
		return p.mainFrame.WaitForFunction(expression, arg, FrameWaitForFunctionOptions(options[0]))
	}
	return p.mainFrame.WaitForFunction(expression, arg)
}

func (p *pageImpl) Dblclick(expression string, options ...PageDblclickOptions) error {
	if len(options) == 1 {
		return p.mainFrame.Dblclick(expression, FrameDblclickOptions(options[0]))
	}
	return p.mainFrame.Dblclick(expression)
}

func (p *pageImpl) Focus(expression string, options ...PageFocusOptions) error {
	if len(options) == 1 {
		return p.mainFrame.Focus(expression, FrameFocusOptions(options[0]))
	}
	return p.mainFrame.Focus(expression)
}

func (p *pageImpl) TextContent(selector string, options ...PageTextContentOptions) (string, error) {
	if len(options) == 1 {
		return p.mainFrame.TextContent(selector, FrameTextContentOptions(options[0]))
	}
	return p.mainFrame.TextContent(selector)
}

func (p *pageImpl) Video() Video {
	p.Lock()
	defer p.Unlock()

	if p.video == nil {
		p.video = newVideo(p)
	}
	return p.video
}

func (p *pageImpl) Tap(selector string, options ...PageTapOptions) error {
	if len(options) == 1 {
		return p.mainFrame.Tap(selector, FrameTapOptions(options[0]))
	}
	return p.mainFrame.Tap(selector)
}

func (p *pageImpl) ExposeFunction(name string, binding ExposedFunction) error {
	return p.ExposeBinding(name, func(source *BindingSource, args ...any) any {
		return binding(args...)
	})
}

func (p *pageImpl) ExposeBinding(name string, binding BindingCallFunction, handle ...bool) error {
	needsHandle := false
	if len(handle) == 1 {
		needsHandle = handle[0]
	}
	if _, ok := p.bindings.Load(name); ok {
		return fmt.Errorf("Function '%s' has been already registered", name)
	}
	if _, ok := p.browserContext.bindings.Load(name); ok {
		return fmt.Errorf("Function '%s' has been already registered in the browser context", name)
	}
	_, err := p.channel.Send("exposeBinding", map[string]any{
		"name":        name,
		"needsHandle": needsHandle,
	})
	if err != nil {
		return err
	}
	p.bindings.Store(name, binding)
	return nil
}

func (p *pageImpl) SelectOption(selector string, values SelectOptionValues, options ...PageSelectOptionOptions) ([]string, error) {
	if len(options) == 1 {
		return p.mainFrame.SelectOption(selector, values, FrameSelectOptionOptions(options[0]))
	}
	return p.mainFrame.SelectOption(selector, values)
}

func (p *pageImpl) IsChecked(selector string, options ...PageIsCheckedOptions) (bool, error) {
	if len(options) == 1 {
		return p.mainFrame.IsChecked(selector, FrameIsCheckedOptions(options[0]))
	}
	return p.mainFrame.IsChecked(selector)
}

func (p *pageImpl) IsDisabled(selector string, options ...PageIsDisabledOptions) (bool, error) {
	if len(options) == 1 {
		return p.mainFrame.IsDisabled(selector, FrameIsDisabledOptions(options[0]))
	}
	return p.mainFrame.IsDisabled(selector)
}

func (p *pageImpl) IsEditable(selector string, options ...PageIsEditableOptions) (bool, error) {
	if len(options) == 1 {
		return p.mainFrame.IsEditable(selector, FrameIsEditableOptions(options[0]))
	}
	return p.mainFrame.IsEditable(selector)
}

func (p *pageImpl) IsEnabled(selector string, options ...PageIsEnabledOptions) (bool, error) {
	if len(options) == 1 {
		return p.mainFrame.IsEnabled(selector, FrameIsEnabledOptions(options[0]))
	}
	return p.mainFrame.IsEnabled(selector)
}

func (p *pageImpl) IsHidden(selector string, options ...PageIsHiddenOptions) (bool, error) {
	if len(options) == 1 {
		return p.mainFrame.IsHidden(selector, FrameIsHiddenOptions(options[0]))
	}
	return p.mainFrame.IsHidden(selector)
}

func (p *pageImpl) IsVisible(selector string, options ...PageIsVisibleOptions) (bool, error) {
	if len(options) == 1 {
		return p.mainFrame.IsVisible(selector, FrameIsVisibleOptions(options[0]))
	}
	return p.mainFrame.IsVisible(selector)
}

func (p *pageImpl) DragAndDrop(source, target string, options ...PageDragAndDropOptions) error {
	if len(options) == 1 {
		return p.mainFrame.DragAndDrop(source, target, FrameDragAndDropOptions(options[0]))
	}
	return p.mainFrame.DragAndDrop(source, target)
}

func (p *pageImpl) Pause() (err error) {
	defaultNavigationTimout := p.browserContext.timeoutSettings.DefaultNavigationTimeout()
	defaultTimeout := p.browserContext.timeoutSettings.DefaultTimeout()
	p.browserContext.SetDefaultNavigationTimeout(0)
	p.browserContext.SetDefaultTimeout(0)
	select {
	case err = <-p.closedOrCrashed:
	case err = <-p.browserContext.pause():
	}
	if err != nil {
		return err
	}
	p.browserContext.setDefaultNavigationTimeoutImpl(defaultNavigationTimout)
	p.browserContext.setDefaultTimeoutImpl(defaultTimeout)
	return
}

func (p *pageImpl) InputValue(selector string, options ...PageInputValueOptions) (string, error) {
	if len(options) == 1 {
		return p.mainFrame.InputValue(selector, FrameInputValueOptions(options[0]))
	}
	return p.mainFrame.InputValue(selector)
}

func (p *pageImpl) WaitForURL(url any, options ...PageWaitForURLOptions) error {
	if len(options) == 1 {
		return p.mainFrame.WaitForURL(url, FrameWaitForURLOptions(options[0]))
	}
	return p.mainFrame.WaitForURL(url)
}

func (p *pageImpl) SetChecked(selector string, checked bool, options ...PageSetCheckedOptions) error {
	if len(options) == 1 {
		return p.mainFrame.SetChecked(selector, checked, FrameSetCheckedOptions(options[0]))
	}
	return p.mainFrame.SetChecked(selector, checked)
}

func (p *pageImpl) Locator(selector string, options ...PageLocatorOptions) Locator {
	var option FrameLocatorOptions
	if len(options) == 1 {
		option = FrameLocatorOptions(options[0])
	}
	return p.mainFrame.Locator(selector, option)
}

func (p *pageImpl) GetByAltText(text any, options ...PageGetByAltTextOptions) Locator {
	exact := false
	if len(options) == 1 {
		if *options[0].Exact {
			exact = true
		}
	}
	return p.Locator(getByAltTextSelector(text, exact))
}

func (p *pageImpl) GetByLabel(text any, options ...PageGetByLabelOptions) Locator {
	exact := false
	if len(options) == 1 {
		if *options[0].Exact {
			exact = true
		}
	}
	return p.Locator(getByLabelSelector(text, exact))
}

func (p *pageImpl) GetByPlaceholder(text any, options ...PageGetByPlaceholderOptions) Locator {
	exact := false
	if len(options) == 1 {
		if *options[0].Exact {
			exact = true
		}
	}
	return p.Locator(getByPlaceholderSelector(text, exact))
}

func (p *pageImpl) GetByRole(role AriaRole, options ...PageGetByRoleOptions) Locator {
	if len(options) == 1 {
		return p.Locator(getByRoleSelector(role, LocatorGetByRoleOptions(options[0])))
	}
	return p.Locator(getByRoleSelector(role))
}

func (p *pageImpl) GetByTestId(testId any) Locator {
	return p.Locator(getByTestIdSelector(getTestIdAttributeName(), testId))
}

func (p *pageImpl) GetByText(text any, options ...PageGetByTextOptions) Locator {
	exact := false
	if len(options) == 1 {
		if *options[0].Exact {
			exact = true
		}
	}
	return p.Locator(getByTextSelector(text, exact))
}

func (p *pageImpl) GetByTitle(text any, options ...PageGetByTitleOptions) Locator {
	exact := false
	if len(options) == 1 {
		if *options[0].Exact {
			exact = true
		}
	}
	return p.Locator(getByTitleSelector(text, exact))
}

func (p *pageImpl) FrameLocator(selector string) FrameLocator {
	return p.mainFrame.FrameLocator(selector)
}

func (p *pageImpl) OnClose(fn func(Page)) {
	p.On("close", fn)
}

func (p *pageImpl) OnConsole(fn func(ConsoleMessage)) {
	p.On("console", fn)
}

func (p *pageImpl) OnCrash(fn func(Page)) {
	p.On("crash", fn)
}

func (p *pageImpl) OnDialog(fn func(Dialog)) {
	p.On("dialog", fn)
}

func (p *pageImpl) OnDOMContentLoaded(fn func(Page)) {
	p.On("domcontentloaded", fn)
}

func (p *pageImpl) OnDownload(fn func(Download)) {
	p.On("download", fn)
}

func (p *pageImpl) OnFileChooser(fn func(FileChooser)) {
	p.On("filechooser", fn)
}

func (p *pageImpl) OnFrameAttached(fn func(Frame)) {
	p.On("frameattached", fn)
}

func (p *pageImpl) OnFrameDetached(fn func(Frame)) {
	p.On("framedetached", fn)
}

func (p *pageImpl) OnFrameNavigated(fn func(Frame)) {
	p.On("framenavigated", fn)
}

func (p *pageImpl) OnLoad(fn func(Page)) {
	p.On("load", fn)
}

func (p *pageImpl) OnPageError(fn func(error)) {
	p.On("pageerror", fn)
}

func (p *pageImpl) OnPopup(fn func(Page)) {
	p.On("popup", fn)
}

func (p *pageImpl) OnRequest(fn func(Request)) {
	p.On("request", fn)
}

func (p *pageImpl) OnRequestFailed(fn func(Request)) {
	p.On("requestfailed", fn)
}

func (p *pageImpl) OnRequestFinished(fn func(Request)) {
	p.On("requestfinished", fn)
}

func (p *pageImpl) OnResponse(fn func(Response)) {
	p.On("response", fn)
}

func (p *pageImpl) OnWebSocket(fn func(WebSocket)) {
	p.On("websocket", fn)
}

func (p *pageImpl) OnWorker(fn func(Worker)) {
	p.On("worker", fn)
}

func (p *pageImpl) RequestGC() error {
	_, err := p.channel.Send("requestGC")
	return err
}

func (p *pageImpl) RouteWebSocket(url any, handler func(WebSocketRoute)) error {
	p.Lock()
	defer p.Unlock()
	p.webSocketRoutes = slices.Insert(p.webSocketRoutes, 0, newWebSocketRouteHandler(newURLMatcher(url, p.browserContext.options.BaseURL, true), handler))

	return p.updateWebSocketInterceptionPatterns()
}

func (p *pageImpl) onWebSocketRoute(wr WebSocketRoute) {
	p.Lock()
	index := slices.IndexFunc(p.webSocketRoutes, func(r *webSocketRouteHandler) bool {
		return r.Matches(wr.URL())
	})
	if index == -1 {
		p.Unlock()
		p.browserContext.onWebSocketRoute(wr)
		return
	}
	handler := p.webSocketRoutes[index]
	p.Unlock()
	handler.Handle(wr)
}

func (p *pageImpl) updateWebSocketInterceptionPatterns() error {
	patterns := prepareWebSocketRouteHandlerInterceptionPatterns(p.webSocketRoutes)
	_, err := p.channel.Send("setWebSocketInterceptionPatterns", map[string]any{
		"patterns": patterns,
	})
	return err
}
