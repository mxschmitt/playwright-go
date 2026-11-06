package playwright

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

// RemoteConnectOptions configures ConnectRemote.
type RemoteConnectOptions struct {
	// Browser is sent as the x-playwright-browser header ("chromium", "firefox"
	// or "webkit"). A server started with BrowserType.launchServer ignores it.
	Browser string
	// Headers are extra handshake headers, e.g. for authentication. User-Agent
	// is reserved: it carries the version check described on ConnectRemote.
	Headers map[string]string
}

// ConnectRemote attaches to the browser served by a remote Playwright server
// (BrowserType.launchServer) over WebSocket. Unlike BrowserType.Connect it
// needs no local Node.js driver: do not call Run or Install, nothing is
// downloaded or started on this machine.
//
// The server must run the same Playwright major.minor as this module, which
// its release tag encodes: v0.<MM><PP>.<G> speaks Playwright 1.MM and works
// with any 1.MM.x server. Otherwise the server refuses the handshake and
// ConnectRemote returns an error wrapping ErrVersionMismatch that names both
// versions.
//
// ctx bounds the handshake and initialization only; closing the returned
// Browser closes the connection.
//
// Features that need the local driver's LocalUtils are unavailable: Devices,
// RouteFromHAR, recording HAR to a non-zip file, and client-side call stacks in
// traces.
func ConnectRemote(ctx context.Context, wsEndpoint string, options ...RemoteConnectOptions) (Browser, error) {
	headers := map[string]string{}
	if len(options) == 1 {
		for k, v := range options[0].Headers {
			headers[k] = v
		}
		if options[0].Browser != "" {
			headers["x-playwright-browser"] = options[0].Browser
		}
	}
	transport, err := dialWSTransport(ctx, wsEndpoint, headers)
	if err != nil {
		return nil, err
	}
	connection := newConnection(transport)
	connection.isRemote = true
	// Installed before Start so the receive loop never races this assignment.
	// The browser is only known after initialization; see the end of this
	// function for a close that happens in between.
	var remoteBrowser atomic.Pointer[browserImpl]
	var closeOnce sync.Once
	markClosed := func() {
		browser := remoteBrowser.Load()
		if browser == nil {
			return
		}
		closeOnce.Do(func() {
			for _, context := range browser.Contexts() {
				for _, page := range context.Pages() {
					page.(*pageImpl).onClose()
				}
				context.(*browserContextImpl).onClose()
			}
			browser.onClose()
		})
	}
	connection.afterClose = markClosed

	type startResult struct {
		pw  *Playwright
		err error
	}
	started := make(chan startResult, 1)
	go func() {
		pw, err := connection.Start()
		started <- startResult{pw, err}
	}()
	var pw *Playwright
	select {
	case r := <-started:
		if r.err != nil {
			_ = transport.Close()
			connection.cleanup(r.err)
			return nil, r.err
		}
		pw = r.pw
	case <-ctx.Done():
		_ = transport.Close()
		connection.cleanup(ctx.Err())
		return nil, ctx.Err()
	}

	preLaunched := fromNullableChannel(pw.initializer["preLaunchedBrowser"])
	if preLaunched == nil {
		_ = connection.Stop()
		return nil, errors.New("malformed endpoint. Did you use BrowserType.LaunchServer method?")
	}
	browser := preLaunched.(*browserImpl)
	browser.shouldCloseConnectionOnClose = true
	// Use the remote Playwright's own BrowserType, whose playwright is set, so
	// contexts register with its Selectors (see registerContextSelectors).
	switch browser.browserType.Name() {
	case "chromium":
		browser.browserType = pw.Chromium
	case "firefox":
		browser.browserType = pw.Firefox
	case "webkit":
		browser.browserType = pw.WebKit
	}
	if bt, ok := browser.browserType.(*browserTypeImpl); ok {
		for _, context := range browser.Contexts() {
			bt.registerContextSelectors(context.(*browserContextImpl))
		}
	}
	remoteBrowser.Store(browser)
	// cleanup records closedError before running afterClose, so a close that
	// ran afterClose before the Store above is still seen here.
	if connection.closedError.Get() != nil {
		markClosed()
	}
	return browser, nil
}
