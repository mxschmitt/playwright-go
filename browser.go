package playwright

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type browserImpl struct {
	channelOwner
	isConnected                  bool
	shouldCloseConnectionOnClose bool
	contexts                     []BrowserContext
	browserType                  BrowserType
	chromiumTracingPath          *string
	closeReason                  *string
}

func (b *browserImpl) BrowserType() BrowserType {
	return b.browserType
}

func (b *browserImpl) IsConnected() bool {
	b.RLock()
	defer b.RUnlock()
	return b.isConnected
}

func (b *browserImpl) NewContext(options ...BrowserNewContextOptions) (BrowserContext, error) {
	overrides := map[string]any{}
	option := BrowserNewContextOptions{}
	if len(options) == 1 {
		option = options[0]
	}
	if option.AcceptDownloads != nil {
		if *option.AcceptDownloads {
			overrides["acceptDownloads"] = "accept"
		} else {
			overrides["acceptDownloads"] = "deny"
		}
		options[0].AcceptDownloads = nil
	}
	if option.ExtraHttpHeaders != nil {
		overrides["extraHTTPHeaders"] = serializeMapToNameAndValue(options[0].ExtraHttpHeaders)
		options[0].ExtraHttpHeaders = nil
	}
	if option.ClientCertificates != nil {
		certs, err := transformClientCertificate(option.ClientCertificates)
		if err != nil {
			return nil, err
		}
		overrides["clientCertificates"] = certs
		options[0].ClientCertificates = nil
	}
	if option.StorageStatePath != nil {
		var storageState *OptionalStorageState
		storageString, err := os.ReadFile(*options[0].StorageStatePath)
		if err != nil {
			return nil, fmt.Errorf("could not read storage state file: %w", err)
		}
		err = json.Unmarshal(storageString, &storageState)
		if err != nil {
			return nil, fmt.Errorf("could not parse storage state file: %w", err)
		}
		options[0].StorageState = storageState
		options[0].StorageStatePath = nil
	}
	if option.NoViewport != nil && *options[0].NoViewport {
		overrides["noDefaultViewport"] = true
		options[0].NoViewport = nil
	}
	if option.RecordHarPath != nil {
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
	channel, err := b.channel.Send("newContext", options, overrides)
	if err != nil {
		return nil, err
	}
	context := fromChannel(channel).(*browserContextImpl)
	context.browser = b
	b.browserType.(*browserTypeImpl).didCreateContext(context, &option, nil)
	if err := context.initializeHarFromOptions(); err != nil {
		return nil, err
	}
	return context, nil
}

func (b *browserImpl) NewPage(options ...BrowserNewPageOptions) (Page, error) {
	opts := make([]BrowserNewContextOptions, 0)
	if len(options) == 1 {
		opts = append(opts, BrowserNewContextOptions(options[0]))
	}
	context, err := b.NewContext(opts...)
	if err != nil {
		return nil, err
	}
	page, err := context.NewPage()
	if err != nil {
		return nil, err
	}
	page.(*pageImpl).ownedContext = context
	context.(*browserContextImpl).ownedPage = page
	return page, nil
}

func (b *browserImpl) NewBrowserCDPSession() (CDPSession, error) {
	channel, err := b.channel.Send("newBrowserCDPSession")
	if err != nil {
		return nil, err
	}

	cdpSession := fromChannel(channel).(*cdpSessionImpl)

	return cdpSession, nil
}

func (b *browserImpl) Contexts() []BrowserContext {
	b.Lock()
	defer b.Unlock()
	return b.contexts
}

func (b *browserImpl) Close(options ...BrowserCloseOptions) (err error) {
	if len(options) == 1 {
		b.closeReason = options[0].Reason
	}

	if b.shouldCloseConnectionOnClose {
		err = b.connection.Stop()
	} else if b.closeReason != nil {
		_, err = b.channel.Send("close", map[string]any{
			"reason": b.closeReason,
		})
	} else {
		_, err = b.channel.Send("close")
	}
	if err != nil && !errors.Is(err, ErrTargetClosed) {
		return fmt.Errorf("close browser failed: %w", err)
	}
	return nil
}

func (b *browserImpl) Version() string {
	return b.initializer["version"].(string)
}

func (b *browserImpl) StartTracing(options ...BrowserStartTracingOptions) error {
	overrides := map[string]any{}
	option := BrowserStartTracingOptions{}
	if len(options) == 1 {
		option = options[0]
	}
	if option.Page != nil {
		overrides["page"] = option.Page.(*pageImpl).channel
		option.Page = nil
	}
	if option.Path != nil {
		b.chromiumTracingPath = option.Path
		option.Path = nil
	}
	_, err := b.channel.Send("startTracing", option, overrides)
	return err
}

func (b *browserImpl) StopTracing() ([]byte, error) {
	channel, err := b.channel.Send("stopTracing")
	if err != nil {
		return nil, err
	}
	artifact := fromChannel(channel).(*artifactImpl)
	binary, err := artifact.ReadIntoBuffer()
	if err != nil {
		return nil, err
	}
	err = artifact.Delete()
	if err != nil {
		return binary, err
	}
	if b.chromiumTracingPath != nil {
		err := os.MkdirAll(filepath.Dir(*b.chromiumTracingPath), 0o777)
		if err != nil {
			return binary, err
		}
		err = os.WriteFile(*b.chromiumTracingPath, binary, 0o644)
		if err != nil {
			return binary, err
		}
	}
	return binary, nil
}

func (b *browserImpl) onClose() {
	b.Lock()
	if b.isConnected {
		b.isConnected = false
		b.Unlock()
		b.Emit("disconnected", b)
		return
	}
	b.Unlock()
}

func (b *browserImpl) OnDisconnected(fn func(Browser)) {
	b.On("disconnected", fn)
}

func newBrowser(parent *channelOwner, objectType string, guid string, initializer map[string]any) *browserImpl {
	b := &browserImpl{
		isConnected: true,
		contexts:    make([]BrowserContext, 0),
	}
	b.createChannelOwner(b, parent, objectType, guid, initializer)
	// convert parent to *browserTypeImpl
	b.browserType = newBrowserType(parent.parent, parent.objectType, parent.guid, parent.initializer)
	b.channel.On("close", b.onClose)
	return b
}

func transformClientCertificate(clientCertificates []ClientCertificate) ([]map[string]any, error) {
	results := make([]map[string]any, 0)

	for _, cert := range clientCertificates {
		data := map[string]any{
			"origin":     cert.Origin,
			"passphrase": cert.Passphrase,
		}
		if len(cert.Cert) > 0 {
			data["cert"] = base64.StdEncoding.EncodeToString(cert.Cert)
		} else if cert.CertPath != nil {
			content, err := os.ReadFile(*cert.CertPath)
			if err != nil {
				return nil, err
			}
			data["cert"] = base64.StdEncoding.EncodeToString(content)
		}

		if len(cert.Key) > 0 {
			data["key"] = base64.StdEncoding.EncodeToString(cert.Key)
		} else if cert.KeyPath != nil {
			content, err := os.ReadFile(*cert.KeyPath)
			if err != nil {
				return nil, err
			}
			data["key"] = base64.StdEncoding.EncodeToString(content)
		}

		if len(cert.Pfx) > 0 {
			data["pfx"] = base64.StdEncoding.EncodeToString(cert.Pfx)
		} else if cert.PfxPath != nil {
			content, err := os.ReadFile(*cert.PfxPath)
			if err != nil {
				return nil, err
			}
			data["pfx"] = base64.StdEncoding.EncodeToString(content)
		}

		results = append(results, data)
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results, nil
}
