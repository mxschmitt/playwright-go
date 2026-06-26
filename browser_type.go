package playwright

import (
	"errors"
	"fmt"
	"path/filepath"
)

// defaultLaunchTimeout matches DEFAULT_PLAYWRIGHT_LAUNCH_TIMEOUT upstream (3 minutes).
const defaultLaunchTimeout = 3 * 60 * 1000

type browserTypeImpl struct {
	channelOwner
	playwright *Playwright
}

func (b *browserTypeImpl) Name() string {
	return b.initializer["name"].(string)
}

func (b *browserTypeImpl) ExecutablePath() string {
	return b.initializer["executablePath"].(string)
}

func (b *browserTypeImpl) Launch(options ...BrowserTypeLaunchOptions) (Browser, error) {
	overrides := map[string]any{}
	// timeout is required in Playwright v1.57+ protocol
	if len(options) == 0 || options[0].Timeout == nil {
		overrides["timeout"] = float64(defaultLaunchTimeout) // default 3 min
	}
	if len(options) == 1 && options[0].Env != nil {
		overrides["env"] = serializeMapToNameAndValue(options[0].Env)
		options[0].Env = nil
	}
	channel, err := b.channel.Send("launch", options, overrides)
	if err != nil {
		return nil, err
	}
	browser := fromChannel(channel).(*browserImpl)
	b.didLaunchBrowser(browser)
	return browser, nil
}

func (b *browserTypeImpl) LaunchPersistentContext(userDataDir string, options ...BrowserTypeLaunchPersistentContextOptions) (BrowserContext, error) {
	// Resolve a relative userDataDir to an absolute path, matching upstream, so
	// the driver receives a path independent of the process working directory.
	if userDataDir != "" && !filepath.IsAbs(userDataDir) {
		if abs, err := filepath.Abs(userDataDir); err == nil {
			userDataDir = abs
		}
	}
	overrides := map[string]any{
		"userDataDir": userDataDir,
	}
	// timeout is required in Playwright v1.57+ protocol
	if len(options) == 0 || options[0].Timeout == nil {
		overrides["timeout"] = float64(defaultLaunchTimeout) // default 3 min
	}
	option := &BrowserNewContextOptions{}
	var tracesDir *string = nil
	if len(options) == 1 {
		tracesDir = options[0].TracesDir
		err := assignStructFields(option, options[0], true)
		if err != nil {
			return nil, fmt.Errorf("can not convert options: %w", err)
		}
		if options[0].AcceptDownloads != nil {
			if *options[0].AcceptDownloads {
				overrides["acceptDownloads"] = "accept"
			} else {
				overrides["acceptDownloads"] = "deny"
			}
			options[0].AcceptDownloads = nil
		}
		if options[0].ClientCertificates != nil {
			certs, err := transformClientCertificate(options[0].ClientCertificates)
			if err != nil {
				return nil, err
			}
			overrides["clientCertificates"] = certs
			options[0].ClientCertificates = nil
		}
		if options[0].ExtraHttpHeaders != nil {
			overrides["extraHTTPHeaders"] = serializeMapToNameAndValue(options[0].ExtraHttpHeaders)
			options[0].ExtraHttpHeaders = nil
		}
		if options[0].Env != nil {
			overrides["env"] = serializeMapToNameAndValue(options[0].Env)
			options[0].Env = nil
		}
		if options[0].NoViewport != nil && *options[0].NoViewport {
			overrides["noDefaultViewport"] = true
			options[0].NoViewport = nil
		}
		if options[0].RecordVideo != nil {
			if err := resolveRecordVideoDir(options[0].RecordVideo); err != nil {
				return nil, err
			}
		}
		if options[0].RecordHarPath != nil {
			overrides["recordHar"] = prepareRecordHarOptions(recordHarInputOptions{
				Path:        *options[0].RecordHarPath,
				URL:         options[0].RecordHarURLFilter,
				Mode:        options[0].RecordHarMode,
				Content:     options[0].RecordHarContent,
				OmitContent: options[0].RecordHarOmitContent,
			})
			options[0].RecordHarPath = nil
			options[0].RecordHarURLFilter = nil
			options[0].RecordHarMode = nil
			options[0].RecordHarContent = nil
			options[0].RecordHarOmitContent = nil
		}
	}
	response, err := b.channel.SendReturnAsDict("launchPersistentContext", options, overrides)
	if err != nil {
		return nil, err
	}
	context := fromChannel(response["context"]).(*browserContextImpl)
	// Wire the real browserType (whose playwright is non-nil) onto the persistent
	// context's browser, then register the context with the selectors manager.
	// newBrowserContext's own registration ran before this wiring (its browser's
	// browserType has a nil playwright), so custom selector engines / testId would
	// otherwise never reach a persistent context.
	if context.browser != nil {
		b.didLaunchBrowser(context.browser)
	}
	b.registerContextSelectors(context)
	b.didCreateContext(context, option, tracesDir)
	if err := context.initializeHarFromOptions(); err != nil {
		return nil, err
	}
	return context, nil
}

func (b *browserTypeImpl) Connect(wsEndpoint string, options ...BrowserTypeConnectOptions) (Browser, error) {
	overrides := map[string]any{
		"endpoint": wsEndpoint,
		"headers": map[string]string{
			"x-playwright-browser": b.Name(),
		},
	}
	// timeout is required in Playwright v1.57+ protocol
	if len(options) == 0 || options[0].Timeout == nil {
		overrides["timeout"] = float64(0) // default no timeout
	}
	if len(options) == 1 {
		if options[0].Headers != nil {
			for k, v := range options[0].Headers {
				overrides["headers"].(map[string]string)[k] = v
			}
			options[0].Headers = nil
		}
	}
	localUtils := b.connection.LocalUtils()
	pipe, err := localUtils.channel.SendReturnAsDict("connect", options, overrides)
	if err != nil {
		return nil, err
	}
	jsonPipe := fromChannel(pipe["pipe"]).(*jsonPipe)
	connection := newConnection(jsonPipe, localUtils)

	playwright, err := connection.Start()
	if err != nil {
		return nil, err
	}
	playwright.setSelectors(b.playwright.Selectors)
	preLaunchedBrowser := fromNullableChannel(playwright.initializer["preLaunchedBrowser"])
	if preLaunchedBrowser == nil {
		connection.cleanup()
		return nil, errors.New("malformed endpoint. Did you use BrowserType.LaunchServer method?")
	}
	browser := preLaunchedBrowser.(*browserImpl)
	browser.shouldCloseConnectionOnClose = true
	pipeClosed := func() {
		for _, context := range browser.Contexts() {
			pages := context.Pages()
			for _, page := range pages {
				page.(*pageImpl).onClose()
			}
			context.(*browserContextImpl).onClose()
		}
		browser.onClose()
		connection.cleanup()
	}
	jsonPipe.On("closed", pipeClosed)

	b.didLaunchBrowser(browser)
	// When connecting to a shared browser server, the browser's pre-existing
	// contexts are dispatched before didLaunchBrowser wires the real browserType
	// (whose playwright is non-nil), so newBrowserContext's own selectors
	// registration was skipped for them. Register them now, mirroring upstream's
	// _connectToBrowserType -> _setupBrowserContext loop over all existing
	// contexts, so custom selector engines / testId reach pre-existing contexts.
	for _, context := range browser.Contexts() {
		b.registerContextSelectors(context.(*browserContextImpl))
	}
	return browser, nil
}

func (b *browserTypeImpl) ConnectOverCDP(endpointURL string, options ...BrowserTypeConnectOverCDPOptions) (Browser, error) {
	if b.Name() != "chromium" {
		return nil, errors.New("connecting over CDP is only supported in Chromium")
	}
	overrides := map[string]any{
		"endpointURL": endpointURL,
	}
	// timeout is required in Playwright v1.57+ protocol
	if len(options) == 0 || options[0].Timeout == nil {
		overrides["timeout"] = float64(30000) // default 30s
	}
	if len(options) == 1 {
		if options[0].Headers != nil {
			overrides["headers"] = serializeMapToNameAndValue(options[0].Headers)
			options[0].Headers = nil
		}
	}
	response, err := b.channel.SendReturnAsDict("connectOverCDP", options, overrides)
	if err != nil {
		return nil, err
	}
	browser := fromChannel(response["browser"]).(*browserImpl)
	b.didLaunchBrowser(browser)
	if defaultContext, ok := response["defaultContext"]; ok {
		context := fromChannel(defaultContext).(*browserContextImpl)
		// The default context arrives during dispatch, before didLaunchBrowser
		// wires the real browserType (whose playwright is non-nil), so
		// newBrowserContext's own selectors registration was skipped. Register
		// it now, mirroring upstream's _connectToBrowserType -> _setupBrowserContext,
		// so custom selector engines / testId reach the CDP default context.
		b.registerContextSelectors(context)
		b.didCreateContext(context, nil, nil)
	}
	return browser, nil
}

func (b *browserTypeImpl) didCreateContext(context *browserContextImpl, contextOptions *BrowserNewContextOptions, tracesDir *string) {
	context.setOptions(contextOptions, tracesDir)
}

// registerContextSelectors registers a context with the selectors manager,
// mirroring upstream _setupBrowserContext. Used for contexts that arrive during
// dispatch (persistent/connect/CDP) before didLaunchBrowser wires the real
// browserType, so newBrowserContext's own registration was skipped. addContext
// is idempotent, so this is safe regardless of call ordering.
func (b *browserTypeImpl) registerContextSelectors(context *browserContextImpl) {
	if b.playwright != nil {
		b.playwright.Selectors.(*selectorsImpl).addContext(context)
	}
}

func (b *browserTypeImpl) didLaunchBrowser(browser *browserImpl) {
	browser.browserType = b
}

func newBrowserType(parent *channelOwner, objectType string, guid string, initializer map[string]any) *browserTypeImpl {
	bt := &browserTypeImpl{}
	bt.createChannelOwner(bt, parent, objectType, guid, initializer)
	return bt
}
