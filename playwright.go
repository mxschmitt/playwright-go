// Package playwright is a library to automate Chromium, Firefox and WebKit with
// a single API. Playwright is built to enable cross-browser web automation that
// is ever-green, capable, reliable and fast.
package playwright

type DeviceDescriptor struct {
	UserAgent          string                     `json:"userAgent"`
	Viewport           *BrowserNewContextViewport `json:"viewport"`
	DeviceScaleFactor  int                        `json:"deviceScaleFactor"`
	IsMobile           bool                       `json:"isMobile"`
	HasTouch           bool                       `json:"hasTouch"`
	DefaultBrowserType string                     `json:"defaultBrowserType"`
}

type Playwright struct {
	channelOwner
	Chromium BrowserType
	Firefox  BrowserType
	WebKit   BrowserType
	Devices  map[string]*DeviceDescriptor
}

func (p *Playwright) Stop() error {
	return p.connection.Stop()
}

func newPlaywright(parent *channelOwner, objectType string, guid string, initializer map[string]interface{}) *Playwright {
	pw := &Playwright{
		Chromium: fromChannel(initializer["chromium"]).(*browserTypeImpl),
		Firefox:  fromChannel(initializer["firefox"]).(*browserTypeImpl),
		WebKit:   fromChannel(initializer["webkit"]).(*browserTypeImpl),
		Devices:  make(map[string]*DeviceDescriptor),
	}
	for _, dd := range initializer["deviceDescriptors"].([]interface{}) {
		entry := dd.(map[string]interface{})
		pw.Devices[entry["name"].(string)] = &DeviceDescriptor{
			Viewport: &BrowserNewContextViewport{},
		}
		remapMapToStruct(entry["descriptor"], pw.Devices[entry["name"].(string)])
	}
	pw.createChannelOwner(pw, parent, objectType, guid, initializer)
	return pw
}
