package playwright

import (
	"fmt"
)

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
		overrides["timeout"] = float64(30000) // default 30s
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
	overrides := map[string]any{
		"userDataDir": userDataDir,
	}
	// timeout is required in Playwright v1.57+ protocol
	if len(options) == 0 || options[0].Timeout == nil {
		overrides["timeout"] = float64(30000) // default 30s
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
	b.didCreateContext(context, option, tracesDir)
	if err := context.initializeHarFromOptions(); err != nil {
		return nil, err
	}
	return context, nil
}

func (b *browserTypeImpl) Connect(wsEndpoint string, options ...BrowserTypeConnectOptions) (Browser, error) {
	overrides := map[string]any{
		"wsEndpoint": wsEndpoint,
		"headers": map[string]string{
			"x-playwright-browser":        b.Name(),
			"x-playwright-launch-options": "{}",
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
	browser := fromChannel(playwright.initializer["preLaunchedBrowser"]).(*browserImpl)
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
	return browser, nil
}

func (b *browserTypeImpl) ConnectOverCDP(endpointURL string, options ...BrowserTypeConnectOverCDPOptions) (Browser, error) {
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
		b.didCreateContext(context, nil, nil)
	}
	return browser, nil
}

func (b *browserTypeImpl) didCreateContext(context *browserContextImpl, contextOptions *BrowserNewContextOptions, tracesDir *string) {
	context.setOptions(contextOptions, tracesDir)
}

func (b *browserTypeImpl) didLaunchBrowser(browser *browserImpl) {
	browser.browserType = b
}

func newBrowserType(parent *channelOwner, objectType string, guid string, initializer map[string]any) *browserTypeImpl {
	bt := &browserTypeImpl{}
	bt.createChannelOwner(bt, parent, objectType, guid, initializer)
	return bt
}
