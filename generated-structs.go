package playwright

type BrowserNewContextOptions struct {
	/// <summary>
	/// <para>
	/// Whether to automatically download all the attachments. Defaults to <c>false</c>
	/// where all the downloads are canceled.
	/// </para>
	/// </summary>
	AcceptDownloads *bool `json:"acceptDownloads"`
	/// <summary><para>Toggles bypassing page's Content-Security-Policy.</para></summary>
	BypassCSP *bool `json:"bypassCSP"`
	/// <summary>
	/// <para>
	/// Emulates <c>'prefers-colors-scheme'</c> media feature, supported values are <c>'light'</c>,
	/// <c>'dark'</c>, <c>'no-preference'</c>. See <see cref="Page.EmulateMedia"/> for more
	/// details. Defaults to <c>'light'</c>.
	/// </para>
	/// </summary>
	ColorScheme *ColorScheme `json:"colorScheme"`
	/// <summary><para>Specify device scale factor (can be thought of as dpr). Defaults to <c>1</c>.</para></summary>
	DeviceScaleFactor *float64 `json:"deviceScaleFactor"`
	/// <summary>
	/// <para>
	/// An object containing additional HTTP headers to be sent with every request. All
	/// header values must be strings.
	/// </para>
	/// </summary>
	ExtraHttpHeaders map[string]string                    `json:"extraHTTPHeaders"`
	Geolocation      *BrowserNewContextOptionsGeolocation `json:"geolocation"`
	/// <summary><para>Specifies if viewport supports touch events. Defaults to false.</para></summary>
	HasTouch *bool `json:"hasTouch"`
	/// <summary>
	/// <para>
	/// Credentials for <a href="https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication">HTTP
	/// authentication</a>.
	/// </para>
	/// </summary>
	HttpCredentials *BrowserNewContextOptionsHttpCredentials `json:"httpCredentials"`
	/// <summary><para>Whether to ignore HTTPS errors during navigation. Defaults to <c>false</c>.</para></summary>
	IgnoreHttpsErrors *bool `json:"ignoreHTTPSErrors"`
	/// <summary>
	/// <para>
	/// Whether the <c>meta viewport</c> tag is taken into account and touch events are
	/// enabled. Defaults to <c>false</c>. Not supported in Firefox.
	/// </para>
	/// </summary>
	IsMobile *bool `json:"isMobile"`
	/// <summary><para>Whether or not to enable JavaScript in the context. Defaults to <c>true</c>.</para></summary>
	JavaScriptEnabled *bool `json:"javaScriptEnabled"`
	/// <summary>
	/// <para>
	/// Specify user locale, for example <c>en-GB</c>, <c>de-DE</c>, etc. Locale will affect
	/// <c>navigator.language</c> value, <c>Accept-Language</c> request header value as
	/// well as number and date formatting rules.
	/// </para>
	/// </summary>
	Locale *string `json:"locale"`
	/// <summary><para>Whether to emulate network being offline. Defaults to <c>false</c>.</para></summary>
	Offline *bool `json:"offline"`
	/// <summary>
	/// <para>
	/// A list of permissions to grant to all pages in this context. See <see cref="BrowserContext.GrantPermissions"/>
	/// for more details.
	/// </para>
	/// </summary>
	Permissions []string `json:"permissions"`
	/// <summary>
	/// <para>
	/// Network proxy settings to use with this context. Note that browser needs to be launched
	/// with the global proxy for this option to work. If all contexts override the proxy,
	/// global proxy will be never used and can be any string, for example <c>launch({ proxy:
	/// { server: 'per-context' } })</c>.
	/// </para>
	/// </summary>
	Proxy *BrowserNewContextOptionsProxy `json:"proxy"`
	/// <summary>
	/// <para>
	/// Enables video recording for all pages into <c>recordVideo.dir</c> directory. If
	/// not specified videos are not recorded. Make sure to await <see cref="BrowserContext.Close"/>
	/// for videos to be saved.
	/// </para>
	/// </summary>
	RecordVideo *BrowserNewContextOptionsRecordVideo `json:"recordVideo"`
	/// <summary>
	/// <para>
	/// Changes the timezone of the context. See <a href="https://cs.chromium.org/chromium/src/third_party/icu/source/data/misc/metaZones.txt?rcl=faee8bc70570192d82d2978a71e2a615788597d1">ICU's
	/// metaZones.txt</a> for a list of supported timezone IDs.
	/// </para>
	/// </summary>
	TimezoneId *string `json:"timezoneId"`
	/// <summary><para>Specific user agent to use in this context.</para></summary>
	UserAgent *string `json:"userAgent"`
	/// <summary>
	/// <para>
	/// Sets a consistent viewport for each page. Defaults to an 1280x720 viewport. <c>no_viewport</c>
	/// disables the fixed viewport.
	/// </para>
	/// </summary>
	Viewport *BrowserNewContextOptionsViewport `json:"viewport"`
}
type BrowserGeolocation struct {
	/// <summary><para>Latitude between -90 and 90.</para></summary>
	Latitude *float64 `json:"latitude"`
	/// <summary><para>Longitude between -180 and 180.</para></summary>
	Longitude *float64 `json:"longitude"`
	/// <summary><para>Non-negative accuracy value. Defaults to <c>0</c>.</para></summary>
	Accuracy *float64 `json:"accuracy"`
}
type BrowserHttpCredentials struct {
	/// <summary><para></para></summary>
	Username *string `json:"username"`
	/// <summary><para></para></summary>
	Password *string `json:"password"`
}
type BrowserProxy struct {
	/// <summary>
	/// <para>
	/// Proxy to be used for all requests. HTTP and SOCKS proxies are supported, for example
	/// <c>http://myproxy.com:3128</c> or <c>socks5://myproxy.com:3128</c>. Short form <c>myproxy.com:3128</c>
	/// is considered an HTTP proxy.
	/// </para>
	/// </summary>
	Server *string `json:"server"`
	/// <summary>
	/// <para>
	/// Optional coma-separated domains to bypass proxy, for example <c>".com, chromium.org,
	/// .domain.com"</c>.
	/// </para>
	/// </summary>
	Bypass *string `json:"bypass"`
	/// <summary><para>Optional username to use if HTTP proxy requires authentication.</para></summary>
	Username *string `json:"username"`
	/// <summary><para>Optional password to use if HTTP proxy requires authentication.</para></summary>
	Password *string `json:"password"`
}
type BrowserRecordVideo struct {
	/// <summary><para>Path to the directory to put videos into.</para></summary>
	Dir *string `json:"dir"`
	/// <summary>
	/// <para>
	/// Optional dimensions of the recorded videos. If not specified the size will be equal
	/// to <c>viewport</c> scaled down to fit into 800x800. If <c>viewport</c> is not configured
	/// explicitly the video size defaults to 800x450. Actual picture of each page will
	/// be scaled down if necessary to fit the specified size.
	/// </para>
	/// </summary>
	Size *BrowserRecordVideoSize `json:"size"`
}
type BrowserViewport struct {
	/// <summary><para>page width in pixels.</para></summary>
	Width *int `json:"width"`
	/// <summary><para>page height in pixels.</para></summary>
	Height *int `json:"height"`
}
type BrowserNewPageOptions struct {
	/// <summary>
	/// <para>
	/// Whether to automatically download all the attachments. Defaults to <c>false</c>
	/// where all the downloads are canceled.
	/// </para>
	/// </summary>
	AcceptDownloads *bool `json:"acceptDownloads"`
	/// <summary><para>Toggles bypassing page's Content-Security-Policy.</para></summary>
	BypassCSP *bool `json:"bypassCSP"`
	/// <summary>
	/// <para>
	/// Emulates <c>'prefers-colors-scheme'</c> media feature, supported values are <c>'light'</c>,
	/// <c>'dark'</c>, <c>'no-preference'</c>. See <see cref="Page.EmulateMedia"/> for more
	/// details. Defaults to <c>'light'</c>.
	/// </para>
	/// </summary>
	ColorScheme *ColorScheme `json:"colorScheme"`
	/// <summary><para>Specify device scale factor (can be thought of as dpr). Defaults to <c>1</c>.</para></summary>
	DeviceScaleFactor *float64 `json:"deviceScaleFactor"`
	/// <summary>
	/// <para>
	/// An object containing additional HTTP headers to be sent with every request. All
	/// header values must be strings.
	/// </para>
	/// </summary>
	ExtraHttpHeaders map[string]string                 `json:"extraHTTPHeaders"`
	Geolocation      *BrowserNewPageOptionsGeolocation `json:"geolocation"`
	/// <summary><para>Specifies if viewport supports touch events. Defaults to false.</para></summary>
	HasTouch *bool `json:"hasTouch"`
	/// <summary>
	/// <para>
	/// Credentials for <a href="https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication">HTTP
	/// authentication</a>.
	/// </para>
	/// </summary>
	HttpCredentials *BrowserNewPageOptionsHttpCredentials `json:"httpCredentials"`
	/// <summary><para>Whether to ignore HTTPS errors during navigation. Defaults to <c>false</c>.</para></summary>
	IgnoreHttpsErrors *bool `json:"ignoreHTTPSErrors"`
	/// <summary>
	/// <para>
	/// Whether the <c>meta viewport</c> tag is taken into account and touch events are
	/// enabled. Defaults to <c>false</c>. Not supported in Firefox.
	/// </para>
	/// </summary>
	IsMobile *bool `json:"isMobile"`
	/// <summary><para>Whether or not to enable JavaScript in the context. Defaults to <c>true</c>.</para></summary>
	JavaScriptEnabled *bool `json:"javaScriptEnabled"`
	/// <summary>
	/// <para>
	/// Specify user locale, for example <c>en-GB</c>, <c>de-DE</c>, etc. Locale will affect
	/// <c>navigator.language</c> value, <c>Accept-Language</c> request header value as
	/// well as number and date formatting rules.
	/// </para>
	/// </summary>
	Locale *string `json:"locale"`
	/// <summary><para>Whether to emulate network being offline. Defaults to <c>false</c>.</para></summary>
	Offline *bool `json:"offline"`
	/// <summary>
	/// <para>
	/// A list of permissions to grant to all pages in this context. See <see cref="BrowserContext.GrantPermissions"/>
	/// for more details.
	/// </para>
	/// </summary>
	Permissions []string `json:"permissions"`
	/// <summary>
	/// <para>
	/// Network proxy settings to use with this context. Note that browser needs to be launched
	/// with the global proxy for this option to work. If all contexts override the proxy,
	/// global proxy will be never used and can be any string, for example <c>launch({ proxy:
	/// { server: 'per-context' } })</c>.
	/// </para>
	/// </summary>
	Proxy *BrowserNewPageOptionsProxy `json:"proxy"`
	/// <summary>
	/// <para>
	/// Enables video recording for all pages into <c>recordVideo.dir</c> directory. If
	/// not specified videos are not recorded. Make sure to await <see cref="BrowserContext.Close"/>
	/// for videos to be saved.
	/// </para>
	/// </summary>
	RecordVideo *BrowserNewPageOptionsRecordVideo `json:"recordVideo"`
	/// <summary>
	/// <para>
	/// Changes the timezone of the context. See <a href="https://cs.chromium.org/chromium/src/third_party/icu/source/data/misc/metaZones.txt?rcl=faee8bc70570192d82d2978a71e2a615788597d1">ICU's
	/// metaZones.txt</a> for a list of supported timezone IDs.
	/// </para>
	/// </summary>
	TimezoneId *string `json:"timezoneId"`
	/// <summary><para>Specific user agent to use in this context.</para></summary>
	UserAgent *string `json:"userAgent"`
	/// <summary>
	/// <para>
	/// Sets a consistent viewport for each page. Defaults to an 1280x720 viewport. <c>no_viewport</c>
	/// disables the fixed viewport.
	/// </para>
	/// </summary>
	Viewport *BrowserNewPageOptionsViewport `json:"viewport"`
}
type BrowserContextCookies struct {
	/// <summary><para></para></summary>
	Name *string `json:"name"`
	/// <summary><para></para></summary>
	Value *string `json:"value"`
	/// <summary><para>either url or domain / path are required. Optional.</para></summary>
	URL *string `json:"url"`
	/// <summary><para>either url or domain / path are required Optional.</para></summary>
	Domain *string `json:"domain"`
	/// <summary><para>either url or domain / path are required Optional.</para></summary>
	Path *string `json:"path"`
	/// <summary><para>Unix time in seconds. Optional.</para></summary>
	Expires *float64 `json:"expires"`
	/// <summary><para>Optional.</para></summary>
	HttpOnly *bool `json:"httpOnly"`
	/// <summary><para>Optional.</para></summary>
	Secure *bool `json:"secure"`
	/// <summary><para>Optional.</para></summary>
	SameSite *SameSiteAttribute `json:"sameSite"`
}
type BrowserContextAddInitScriptOptions struct {
	/// <summary><para>Optional Script source to be evaluated in all pages in the browser context.</para></summary>
	Script *string `json:"script"`
	/// <summary><para>Optional Script path to be evaluated in all pages in the browser context.</para></summary>
	Path *string `json:"path"`
}

// Result of calling <see cref="BrowserContext.Cookies" />.
type BrowserContextCookiesResult struct {
	/// <summary><para></para></summary>
	Name *string `json:"name"`
	/// <summary><para></para></summary>
	Value *string `json:"value"`
	/// <summary><para></para></summary>
	Domain *string `json:"domain"`
	/// <summary><para></para></summary>
	Path *string `json:"path"`
	/// <summary><para>Unix time in seconds.</para></summary>
	Expires *float64 `json:"expires"`
	/// <summary><para></para></summary>
	HttpOnly *bool `json:"httpOnly"`
	/// <summary><para></para></summary>
	Secure *bool `json:"secure"`
	/// <summary><para></para></summary>
	SameSite *SameSiteAttribute `json:"sameSite"`
}
type BrowserContextCookiesOptions struct {
	/// <summary><para>Optional list of URLs.</para></summary>
	Urls []string `json:"urls"`
}
type BrowserContextExposeBindingOptions struct {
	/// <summary>
	/// <para>
	/// Whether to pass the argument as a handle, instead of passing by value. When passing
	/// a handle, only one argument is supported. When passing by value, multiple arguments
	/// are supported.
	/// </para>
	/// </summary>
	Handle *bool `json:"handle"`
}
type BrowserContextGrantPermissionsOptions struct {
	/// <summary><para>The <see cref="origin"/> to grant permissions to, e.g. "https://example.com".</para></summary>
	Origin *string `json:"origin"`
}
type BrowserContextRouteOptions struct {
	/// <summary><para>handler function to route the request.</para></summary>
	Handler func(Route) `json:"handler"`
}
type BrowserContextGeolocation struct {
	/// <summary><para>Latitude between -90 and 90.</para></summary>
	Latitude *float64 `json:"latitude"`
	/// <summary><para>Longitude between -180 and 180.</para></summary>
	Longitude *float64 `json:"longitude"`
	/// <summary><para>Non-negative accuracy value. Defaults to <c>0</c>.</para></summary>
	Accuracy *float64 `json:"accuracy"`
}

// Result of calling <see cref="BrowserContext.StorageState" />.
type BrowserContextStorageStateResult struct {
	/// <summary><para></para></summary>
	Cookies []BrowserContextStorageStateResultCookies `json:"cookies"`
	/// <summary><para></para></summary>
	Origins []BrowserContextStorageStateResultOrigins `json:"origins"`
}
type BrowserContextStorageStateOptions struct {
	/// <summary>
	/// <para>
	/// The file path to save the storage state to. If <paramref name="path"/> is a relative
	/// path, then it is resolved relative to current working directory. If no path is provided,
	/// storage state is still returned, but won't be saved to the disk.
	/// </para>
	/// </summary>
	Path *string `json:"path"`
}
type BrowserContextUnrouteOptions struct {
	/// <summary><para>Optional handler function used to register a routing with <see cref="BrowserContext.Route"/>.</para></summary>
	Handler func(Route, Request) `json:"handler"`
}
type BrowserTypeLaunchOptions struct {
	/// <summary>
	/// <para>
	/// Additional arguments to pass to the browser instance. The list of Chromium flags
	/// can be found <a href="http://peter.sh/experiments/chromium-command-line-switches/">here</a>.
	/// </para>
	/// </summary>
	Args []string `json:"args"`
	/// <summary>
	/// <para>
	/// Browser distribution channel. Read more about using <a href="./browsers#google-chrome--microsoft-edge">Google
	/// Chrome and Microsoft Edge</a>.
	/// </para>
	/// </summary>
	Channel *BrowserChannel `json:"channel"`
	/// <summary><para>Enable Chromium sandboxing. Defaults to <c>false</c>.</para></summary>
	ChromiumSandbox *bool `json:"chromiumSandbox"`
	/// <summary>
	/// <para>
	/// **Chromium-only** Whether to auto-open a Developer Tools panel for each tab. If
	/// this option is <c>true</c>, the <paramref name="headless"/> option will be set <c>false</c>.
	/// </para>
	/// </summary>
	Devtools *bool `json:"devtools"`
	/// <summary>
	/// <para>
	/// If specified, accepted downloads are downloaded into this directory. Otherwise,
	/// temporary directory is created and is deleted when browser is closed.
	/// </para>
	/// </summary>
	DownloadsPath *string `json:"downloadsPath"`
	/// <summary>
	/// <para>
	/// Path to a browser executable to run instead of the bundled one. If <paramref name="executablePath"/>
	/// is a relative path, then it is resolved relative to the current working directory.
	/// Note that Playwright only works with the bundled Chromium, Firefox or WebKit, use
	/// at your own risk.
	/// </para>
	/// </summary>
	ExecutablePath *string `json:"executablePath"`
	/// <summary><para>Close the browser process on SIGHUP. Defaults to <c>true</c>.</para></summary>
	HandleSIGHUP *bool `json:"handleSIGHUP"`
	/// <summary><para>Close the browser process on Ctrl-C. Defaults to <c>true</c>.</para></summary>
	HandleSIGINT *bool `json:"handleSIGINT"`
	/// <summary><para>Close the browser process on SIGTERM. Defaults to <c>true</c>.</para></summary>
	HandleSIGTERM *bool `json:"handleSIGTERM"`
	/// <summary>
	/// <para>
	/// Whether to run browser in headless mode. More details for <a href="https://developers.google.com/web/updates/2017/04/headless-chrome">Chromium</a>
	/// and <a href="https://developer.mozilla.org/en-US/docs/Mozilla/Firefox/Headless_mode">Firefox</a>.
	/// Defaults to <c>true</c> unless the <paramref name="devtools"/> option is <c>true</c>.
	/// </para>
	/// </summary>
	Headless *bool `json:"headless"`
	/// <summary><para>Network proxy settings.</para></summary>
	Proxy *BrowserTypeLaunchOptionsProxy `json:"proxy"`
	/// <summary>
	/// <para>
	/// Slows down Playwright operations by the specified amount of milliseconds. Useful
	/// so that you can see what is going on.
	/// </para>
	/// </summary>
	SlowMo *float64 `json:"slowMo"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds to wait for the browser instance to start. Defaults
	/// to <c>30000</c> (30 seconds). Pass <c>0</c> to disable timeout.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type BrowserTypeProxy struct {
	/// <summary>
	/// <para>
	/// Proxy to be used for all requests. HTTP and SOCKS proxies are supported, for example
	/// <c>http://myproxy.com:3128</c> or <c>socks5://myproxy.com:3128</c>. Short form <c>myproxy.com:3128</c>
	/// is considered an HTTP proxy.
	/// </para>
	/// </summary>
	Server *string `json:"server"`
	/// <summary>
	/// <para>
	/// Optional coma-separated domains to bypass proxy, for example <c>".com, chromium.org,
	/// .domain.com"</c>.
	/// </para>
	/// </summary>
	Bypass *string `json:"bypass"`
	/// <summary><para>Optional username to use if HTTP proxy requires authentication.</para></summary>
	Username *string `json:"username"`
	/// <summary><para>Optional password to use if HTTP proxy requires authentication.</para></summary>
	Password *string `json:"password"`
}
type BrowserTypeLaunchPersistentContextOptions struct {
	/// <summary>
	/// <para>
	/// Whether to automatically download all the attachments. Defaults to <c>false</c>
	/// where all the downloads are canceled.
	/// </para>
	/// </summary>
	AcceptDownloads *bool `json:"acceptDownloads"`
	/// <summary>
	/// <para>
	/// Additional arguments to pass to the browser instance. The list of Chromium flags
	/// can be found <a href="http://peter.sh/experiments/chromium-command-line-switches/">here</a>.
	/// </para>
	/// </summary>
	Args []string `json:"args"`
	/// <summary><para>Toggles bypassing page's Content-Security-Policy.</para></summary>
	BypassCSP *bool `json:"bypassCSP"`
	/// <summary><para>Browser distribution channel.</para></summary>
	Channel *BrowserChannel `json:"channel"`
	/// <summary><para>Enable Chromium sandboxing. Defaults to <c>true</c>.</para></summary>
	ChromiumSandbox *bool `json:"chromiumSandbox"`
	/// <summary>
	/// <para>
	/// Emulates <c>'prefers-colors-scheme'</c> media feature, supported values are <c>'light'</c>,
	/// <c>'dark'</c>, <c>'no-preference'</c>. See <see cref="Page.EmulateMedia"/> for more
	/// details. Defaults to <c>'light'</c>.
	/// </para>
	/// </summary>
	ColorScheme *ColorScheme `json:"colorScheme"`
	/// <summary><para>Specify device scale factor (can be thought of as dpr). Defaults to <c>1</c>.</para></summary>
	DeviceScaleFactor *float64 `json:"deviceScaleFactor"`
	/// <summary>
	/// <para>
	/// **Chromium-only** Whether to auto-open a Developer Tools panel for each tab. If
	/// this option is <c>true</c>, the <paramref name="headless"/> option will be set <c>false</c>.
	/// </para>
	/// </summary>
	Devtools *bool `json:"devtools"`
	/// <summary>
	/// <para>
	/// If specified, accepted downloads are downloaded into this directory. Otherwise,
	/// temporary directory is created and is deleted when browser is closed.
	/// </para>
	/// </summary>
	DownloadsPath *string `json:"downloadsPath"`
	/// <summary>
	/// <para>
	/// Path to a browser executable to run instead of the bundled one. If <paramref name="executablePath"/>
	/// is a relative path, then it is resolved relative to the current working directory.
	/// **BEWARE**: Playwright is only guaranteed to work with the bundled Chromium, Firefox
	/// or WebKit, use at your own risk.
	/// </para>
	/// </summary>
	ExecutablePath *string `json:"executablePath"`
	/// <summary>
	/// <para>
	/// An object containing additional HTTP headers to be sent with every request. All
	/// header values must be strings.
	/// </para>
	/// </summary>
	ExtraHttpHeaders map[string]string                                     `json:"extraHTTPHeaders"`
	Geolocation      *BrowserTypeLaunchPersistentContextOptionsGeolocation `json:"geolocation"`
	/// <summary><para>Close the browser process on SIGHUP. Defaults to <c>true</c>.</para></summary>
	HandleSIGHUP *bool `json:"handleSIGHUP"`
	/// <summary><para>Close the browser process on Ctrl-C. Defaults to <c>true</c>.</para></summary>
	HandleSIGINT *bool `json:"handleSIGINT"`
	/// <summary><para>Close the browser process on SIGTERM. Defaults to <c>true</c>.</para></summary>
	HandleSIGTERM *bool `json:"handleSIGTERM"`
	/// <summary><para>Specifies if viewport supports touch events. Defaults to false.</para></summary>
	HasTouch *bool `json:"hasTouch"`
	/// <summary>
	/// <para>
	/// Whether to run browser in headless mode. More details for <a href="https://developers.google.com/web/updates/2017/04/headless-chrome">Chromium</a>
	/// and <a href="https://developer.mozilla.org/en-US/docs/Mozilla/Firefox/Headless_mode">Firefox</a>.
	/// Defaults to <c>true</c> unless the <paramref name="devtools"/> option is <c>true</c>.
	/// </para>
	/// </summary>
	Headless *bool `json:"headless"`
	/// <summary>
	/// <para>
	/// Credentials for <a href="https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication">HTTP
	/// authentication</a>.
	/// </para>
	/// </summary>
	HttpCredentials *BrowserTypeLaunchPersistentContextOptionsHttpCredentials `json:"httpCredentials"`
	/// <summary><para>Whether to ignore HTTPS errors during navigation. Defaults to <c>false</c>.</para></summary>
	IgnoreHttpsErrors *bool `json:"ignoreHTTPSErrors"`
	/// <summary>
	/// <para>
	/// Whether the <c>meta viewport</c> tag is taken into account and touch events are
	/// enabled. Defaults to <c>false</c>. Not supported in Firefox.
	/// </para>
	/// </summary>
	IsMobile *bool `json:"isMobile"`
	/// <summary><para>Whether or not to enable JavaScript in the context. Defaults to <c>true</c>.</para></summary>
	JavaScriptEnabled *bool `json:"javaScriptEnabled"`
	/// <summary>
	/// <para>
	/// Specify user locale, for example <c>en-GB</c>, <c>de-DE</c>, etc. Locale will affect
	/// <c>navigator.language</c> value, <c>Accept-Language</c> request header value as
	/// well as number and date formatting rules.
	/// </para>
	/// </summary>
	Locale *string `json:"locale"`
	/// <summary><para>Whether to emulate network being offline. Defaults to <c>false</c>.</para></summary>
	Offline *bool `json:"offline"`
	/// <summary>
	/// <para>
	/// A list of permissions to grant to all pages in this context. See <see cref="BrowserContext.GrantPermissions"/>
	/// for more details.
	/// </para>
	/// </summary>
	Permissions []string `json:"permissions"`
	/// <summary><para>Network proxy settings.</para></summary>
	Proxy *BrowserTypeLaunchPersistentContextOptionsProxy `json:"proxy"`
	/// <summary>
	/// <para>
	/// Enables video recording for all pages into <c>recordVideo.dir</c> directory. If
	/// not specified videos are not recorded. Make sure to await <see cref="BrowserContext.Close"/>
	/// for videos to be saved.
	/// </para>
	/// </summary>
	RecordVideo *BrowserTypeLaunchPersistentContextOptionsRecordVideo `json:"recordVideo"`
	/// <summary>
	/// <para>
	/// Slows down Playwright operations by the specified amount of milliseconds. Useful
	/// so that you can see what is going on. Defaults to 0.
	/// </para>
	/// </summary>
	SlowMo *float64 `json:"slowMo"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds to wait for the browser instance to start. Defaults
	/// to <c>30000</c> (30 seconds). Pass <c>0</c> to disable timeout.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
	/// <summary>
	/// <para>
	/// Changes the timezone of the context. See <a href="https://cs.chromium.org/chromium/src/third_party/icu/source/data/misc/metaZones.txt?rcl=faee8bc70570192d82d2978a71e2a615788597d1">ICU's
	/// metaZones.txt</a> for a list of supported timezone IDs.
	/// </para>
	/// </summary>
	TimezoneId *string `json:"timezoneId"`
	/// <summary><para>Specific user agent to use in this context.</para></summary>
	UserAgent *string `json:"userAgent"`
	/// <summary>
	/// <para>
	/// Sets a consistent viewport for each page. Defaults to an 1280x720 viewport. <c>no_viewport</c>
	/// disables the fixed viewport.
	/// </para>
	/// </summary>
	Viewport *BrowserTypeLaunchPersistentContextOptionsViewport `json:"viewport"`
}
type BrowserTypeGeolocation struct {
	/// <summary><para>Latitude between -90 and 90.</para></summary>
	Latitude *float64 `json:"latitude"`
	/// <summary><para>Longitude between -180 and 180.</para></summary>
	Longitude *float64 `json:"longitude"`
	/// <summary><para>Non-negative accuracy value. Defaults to <c>0</c>.</para></summary>
	Accuracy *float64 `json:"accuracy"`
}
type BrowserTypeHttpCredentials struct {
	/// <summary><para></para></summary>
	Username *string `json:"username"`
	/// <summary><para></para></summary>
	Password *string `json:"password"`
}
type BrowserTypeRecordVideo struct {
	/// <summary><para>Path to the directory to put videos into.</para></summary>
	Dir *string `json:"dir"`
	/// <summary>
	/// <para>
	/// Optional dimensions of the recorded videos. If not specified the size will be equal
	/// to <c>viewport</c> scaled down to fit into 800x800. If <c>viewport</c> is not configured
	/// explicitly the video size defaults to 800x450. Actual picture of each page will
	/// be scaled down if necessary to fit the specified size.
	/// </para>
	/// </summary>
	Size *BrowserTypeRecordVideoSize `json:"size"`
}
type BrowserTypeViewport struct {
	/// <summary><para>page width in pixels.</para></summary>
	Width *int `json:"width"`
	/// <summary><para>page height in pixels.</para></summary>
	Height *int `json:"height"`
}
type DialogAcceptOptions struct {
	/// <summary>
	/// <para>
	/// A text to enter in prompt. Does not cause any effects if the dialog's <c>type</c>
	/// is not prompt. Optional.
	/// </para>
	/// </summary>
	PromptText *string `json:"promptText"`
}

// Result of calling <see cref="ElementHandle.BoundingBox" />.
type ElementHandleBoundingBoxResult struct {
	/// <summary><para>the x coordinate of the element in pixels.</para></summary>
	X *float64 `json:"x"`
	/// <summary><para>the y coordinate of the element in pixels.</para></summary>
	Y *float64 `json:"y"`
	/// <summary><para>the width of the element in pixels.</para></summary>
	Width *float64 `json:"width"`
	/// <summary><para>the height of the element in pixels.</para></summary>
	Height *float64 `json:"height"`
}
type ElementHandleCheckOptions struct {
	/// <summary>
	/// <para>
	/// Whether to bypass the <a href="./actionability.md">actionability</a> checks. Defaults
	/// to <c>false</c>.
	/// </para>
	/// </summary>
	Force *bool `json:"force"`
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type ElementHandleClickOptions struct {
	/// <summary><para>Defaults to <c>left</c>.</para></summary>
	Button *MouseButton `json:"button"`
	/// <summary><para>defaults to 1. See <see cref="UIEvent.detail"/>.</para></summary>
	ClickCount *int `json:"clickCount"`
	/// <summary>
	/// <para>
	/// Time to wait between <c>mousedown</c> and <c>mouseup</c> in milliseconds. Defaults
	/// to 0.
	/// </para>
	/// </summary>
	Delay *float64 `json:"delay"`
	/// <summary>
	/// <para>
	/// Whether to bypass the <a href="./actionability.md">actionability</a> checks. Defaults
	/// to <c>false</c>.
	/// </para>
	/// </summary>
	Force *bool `json:"force"`
	/// <summary>
	/// <para>
	/// Modifier keys to press. Ensures that only these modifiers are pressed during the
	/// operation, and then restores current modifiers back. If not specified, currently
	/// pressed modifiers are used.
	/// </para>
	/// </summary>
	Modifiers []KeyboardModifier `json:"modifiers"`
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// A point to use relative to the top-left corner of element padding box. If not specified,
	/// uses some visible point of the element.
	/// </para>
	/// </summary>
	Position *ElementHandleClickOptionsPosition `json:"position"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type ElementHandlePosition struct {
	/// <summary><para></para></summary>
	X *float64 `json:"x"`
	/// <summary><para></para></summary>
	Y *float64 `json:"y"`
}
type ElementHandleDblclickOptions struct {
	/// <summary><para>Defaults to <c>left</c>.</para></summary>
	Button *MouseButton `json:"button"`
	/// <summary>
	/// <para>
	/// Time to wait between <c>mousedown</c> and <c>mouseup</c> in milliseconds. Defaults
	/// to 0.
	/// </para>
	/// </summary>
	Delay *float64 `json:"delay"`
	/// <summary>
	/// <para>
	/// Whether to bypass the <a href="./actionability.md">actionability</a> checks. Defaults
	/// to <c>false</c>.
	/// </para>
	/// </summary>
	Force *bool `json:"force"`
	/// <summary>
	/// <para>
	/// Modifier keys to press. Ensures that only these modifiers are pressed during the
	/// operation, and then restores current modifiers back. If not specified, currently
	/// pressed modifiers are used.
	/// </para>
	/// </summary>
	Modifiers []KeyboardModifier `json:"modifiers"`
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// A point to use relative to the top-left corner of element padding box. If not specified,
	/// uses some visible point of the element.
	/// </para>
	/// </summary>
	Position *ElementHandleDblclickOptionsPosition `json:"position"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type ElementHandleDispatchEventOptions struct {
	/// <summary><para>Optional event-specific initialization properties.</para></summary>
	EventInit interface{} `json:"eventInit"`
}
type ElementHandleEvalOnSelectorOptions struct {
	/// <summary><para>Optional argument to pass to <paramref name="expression"/>.</para></summary>
	Arg interface{} `json:"arg"`
}
type ElementHandleEvalOnSelectorAllOptions struct {
	/// <summary><para>Optional argument to pass to <paramref name="expression"/>.</para></summary>
	Arg interface{} `json:"arg"`
}
type ElementHandleFillOptions struct {
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type ElementHandleHoverOptions struct {
	/// <summary>
	/// <para>
	/// Whether to bypass the <a href="./actionability.md">actionability</a> checks. Defaults
	/// to <c>false</c>.
	/// </para>
	/// </summary>
	Force *bool `json:"force"`
	/// <summary>
	/// <para>
	/// Modifier keys to press. Ensures that only these modifiers are pressed during the
	/// operation, and then restores current modifiers back. If not specified, currently
	/// pressed modifiers are used.
	/// </para>
	/// </summary>
	Modifiers []KeyboardModifier `json:"modifiers"`
	/// <summary>
	/// <para>
	/// A point to use relative to the top-left corner of element padding box. If not specified,
	/// uses some visible point of the element.
	/// </para>
	/// </summary>
	Position *ElementHandleHoverOptionsPosition `json:"position"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type ElementHandlePressOptions struct {
	/// <summary>
	/// <para>
	/// Time to wait between <c>keydown</c> and <c>keyup</c> in milliseconds. Defaults to
	/// 0.
	/// </para>
	/// </summary>
	Delay *float64 `json:"delay"`
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type ElementHandleScreenshotOptions struct {
	/// <summary>
	/// <para>
	/// Hides default white background and allows capturing screenshots with transparency.
	/// Not applicable to <c>jpeg</c> images. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	OmitBackground *bool `json:"omitBackground"`
	/// <summary>
	/// <para>
	/// The file path to save the image to. The screenshot type will be inferred from file
	/// extension. If <paramref name="path"/> is a relative path, then it is resolved relative
	/// to the current working directory. If no path is provided, the image won't be saved
	/// to the disk.
	/// </para>
	/// </summary>
	Path *string `json:"path"`
	/// <summary><para>The quality of the image, between 0-100. Not applicable to <c>png</c> images.</para></summary>
	Quality *int `json:"quality"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
	/// <summary><para>Specify screenshot type, defaults to <c>png</c>.</para></summary>
	Type *ScreenshotType `json:"type"`
}
type ElementHandleScrollIntoViewIfNeededOptions struct {
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type ElementHandleSelectOptionOptions struct {
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type ElementHandleSelectTextOptions struct {
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type ElementHandleSetInputFilesOptions struct {
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type ElementHandleTapOptions struct {
	/// <summary>
	/// <para>
	/// Whether to bypass the <a href="./actionability.md">actionability</a> checks. Defaults
	/// to <c>false</c>.
	/// </para>
	/// </summary>
	Force *bool `json:"force"`
	/// <summary>
	/// <para>
	/// Modifier keys to press. Ensures that only these modifiers are pressed during the
	/// operation, and then restores current modifiers back. If not specified, currently
	/// pressed modifiers are used.
	/// </para>
	/// </summary>
	Modifiers []KeyboardModifier `json:"modifiers"`
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// A point to use relative to the top-left corner of element padding box. If not specified,
	/// uses some visible point of the element.
	/// </para>
	/// </summary>
	Position *ElementHandleTapOptionsPosition `json:"position"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type ElementHandleTypeOptions struct {
	/// <summary><para>Time to wait between key presses in milliseconds. Defaults to 0.</para></summary>
	Delay *float64 `json:"delay"`
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type ElementHandleUncheckOptions struct {
	/// <summary>
	/// <para>
	/// Whether to bypass the <a href="./actionability.md">actionability</a> checks. Defaults
	/// to <c>false</c>.
	/// </para>
	/// </summary>
	Force *bool `json:"force"`
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type ElementHandleWaitForElementStateOptions struct {
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type ElementHandleWaitForSelectorOptions struct {
	/// <summary>
	/// <para>Defaults to <c>'visible'</c>. Can be either:</para>
	/// <list type="bullet">
	/// <item><description><c>'attached'</c> - wait for element to be present in DOM.</description></item>
	/// <item><description><c>'detached'</c> - wait for element to not be present in DOM.</description></item>
	/// <item><description>
	/// <c>'visible'</c> - wait for element to have non-empty bounding box and no <c>visibility:hidden</c>.
	/// Note that element without any content or with <c>display:none</c> has an empty bounding
	/// box and is not considered visible.
	/// </description></item>
	/// <item><description>
	/// <c>'hidden'</c> - wait for element to be either detached from DOM, or have an empty
	/// bounding box or <c>visibility:hidden</c>. This is opposite to the <c>'visible'</c>
	/// option.
	/// </description></item>
	/// </list>
	/// </summary>
	State *WaitForSelectorState `json:"state"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type FileChooserSetFilesOptions struct {
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type FrameAddScriptTagOptions struct {
	/// <summary><para>Raw JavaScript content to be injected into frame.</para></summary>
	Content *string `json:"content"`
	/// <summary>
	/// <para>
	/// Path to the JavaScript file to be injected into frame. If <c>path</c> is a relative
	/// path, then it is resolved relative to the current working directory.
	/// </para>
	/// </summary>
	Path *string `json:"path"`
	/// <summary>
	/// <para>
	/// Script type. Use 'module' in order to load a Javascript ES6 module. See <a href="https://developer.mozilla.org/en-US/docs/Web/HTML/Element/script">script</a>
	/// for more details.
	/// </para>
	/// </summary>
	Type *string `json:"type"`
	/// <summary><para>URL of a script to be added.</para></summary>
	URL *string `json:"url"`
}
type FrameAddStyleTagOptions struct {
	/// <summary><para>Raw CSS content to be injected into frame.</para></summary>
	Content *string `json:"content"`
	/// <summary>
	/// <para>
	/// Path to the CSS file to be injected into frame. If <c>path</c> is a relative path,
	/// then it is resolved relative to the current working directory.
	/// </para>
	/// </summary>
	Path *string `json:"path"`
	/// <summary><para>URL of the <c>&lt;link&gt;</c> tag.</para></summary>
	URL *string `json:"url"`
}
type FrameCheckOptions struct {
	/// <summary>
	/// <para>
	/// Whether to bypass the <a href="./actionability.md">actionability</a> checks. Defaults
	/// to <c>false</c>.
	/// </para>
	/// </summary>
	Force *bool `json:"force"`
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type FrameClickOptions struct {
	/// <summary><para>Defaults to <c>left</c>.</para></summary>
	Button *MouseButton `json:"button"`
	/// <summary><para>defaults to 1. See <see cref="UIEvent.detail"/>.</para></summary>
	ClickCount *int `json:"clickCount"`
	/// <summary>
	/// <para>
	/// Time to wait between <c>mousedown</c> and <c>mouseup</c> in milliseconds. Defaults
	/// to 0.
	/// </para>
	/// </summary>
	Delay *float64 `json:"delay"`
	/// <summary>
	/// <para>
	/// Whether to bypass the <a href="./actionability.md">actionability</a> checks. Defaults
	/// to <c>false</c>.
	/// </para>
	/// </summary>
	Force *bool `json:"force"`
	/// <summary>
	/// <para>
	/// Modifier keys to press. Ensures that only these modifiers are pressed during the
	/// operation, and then restores current modifiers back. If not specified, currently
	/// pressed modifiers are used.
	/// </para>
	/// </summary>
	Modifiers []KeyboardModifier `json:"modifiers"`
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// A point to use relative to the top-left corner of element padding box. If not specified,
	/// uses some visible point of the element.
	/// </para>
	/// </summary>
	Position *FrameClickOptionsPosition `json:"position"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type FramePosition struct {
	/// <summary><para></para></summary>
	X *float64 `json:"x"`
	/// <summary><para></para></summary>
	Y *float64 `json:"y"`
}
type FrameDblclickOptions struct {
	/// <summary><para>Defaults to <c>left</c>.</para></summary>
	Button *MouseButton `json:"button"`
	/// <summary>
	/// <para>
	/// Time to wait between <c>mousedown</c> and <c>mouseup</c> in milliseconds. Defaults
	/// to 0.
	/// </para>
	/// </summary>
	Delay *float64 `json:"delay"`
	/// <summary>
	/// <para>
	/// Whether to bypass the <a href="./actionability.md">actionability</a> checks. Defaults
	/// to <c>false</c>.
	/// </para>
	/// </summary>
	Force *bool `json:"force"`
	/// <summary>
	/// <para>
	/// Modifier keys to press. Ensures that only these modifiers are pressed during the
	/// operation, and then restores current modifiers back. If not specified, currently
	/// pressed modifiers are used.
	/// </para>
	/// </summary>
	Modifiers []KeyboardModifier `json:"modifiers"`
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// A point to use relative to the top-left corner of element padding box. If not specified,
	/// uses some visible point of the element.
	/// </para>
	/// </summary>
	Position *FrameDblclickOptionsPosition `json:"position"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type FrameDispatchEventOptions struct {
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type FrameEvalOnSelectorOptions struct {
	/// <summary><para>Optional argument to pass to <paramref name="expression"/>.</para></summary>
	Arg interface{} `json:"arg"`
}
type FrameEvalOnSelectorAllOptions struct {
	/// <summary><para>Optional argument to pass to <paramref name="expression"/>.</para></summary>
	Arg interface{} `json:"arg"`
}
type FrameEvaluateOptions struct {
	/// <summary><para>Optional argument to pass to <paramref name="expression"/>.</para></summary>
	Arg interface{} `json:"arg"`
}
type FrameEvaluateHandleOptions struct {
	/// <summary><para>Optional argument to pass to <paramref name="expression"/>.</para></summary>
	Arg interface{} `json:"arg"`
}
type FrameFillOptions struct {
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type FrameFocusOptions struct {
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type FrameGetAttributeOptions struct {
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type FrameGotoOptions struct {
	/// <summary>
	/// <para>
	/// Referer header value. If provided it will take preference over the referer header
	/// value set by <see cref="Page.SetExtraHttpHeaders"/>.
	/// </para>
	/// </summary>
	Referer *string `json:"referer"`
	/// <summary>
	/// <para>
	/// Maximum operation time in milliseconds, defaults to 30 seconds, pass <c>0</c> to
	/// disable timeout. The default value can be changed by using the <see cref="BrowserContext.SetDefaultNavigationTimeout"/>,
	/// <see cref="BrowserContext.SetDefaultTimeout"/>, <see cref="Page.SetDefaultNavigationTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
	/// <summary>
	/// <para>When to consider operation succeeded, defaults to <c>load</c>. Events can be either:</para>
	/// <list type="bullet">
	/// <item><description>
	/// <c>'domcontentloaded'</c> - consider operation to be finished when the <c>DOMContentLoaded</c>
	/// event is fired.
	/// </description></item>
	/// <item><description>
	/// <c>'load'</c> - consider operation to be finished when the <c>load</c> event is
	/// fired.
	/// </description></item>
	/// <item><description>
	/// <c>'networkidle'</c> - consider operation to be finished when there are no network
	/// connections for at least <c>500</c> ms.
	/// </description></item>
	/// </list>
	/// </summary>
	WaitUntil *WaitUntilState `json:"waitUntil"`
}
type FrameHoverOptions struct {
	/// <summary>
	/// <para>
	/// Whether to bypass the <a href="./actionability.md">actionability</a> checks. Defaults
	/// to <c>false</c>.
	/// </para>
	/// </summary>
	Force *bool `json:"force"`
	/// <summary>
	/// <para>
	/// Modifier keys to press. Ensures that only these modifiers are pressed during the
	/// operation, and then restores current modifiers back. If not specified, currently
	/// pressed modifiers are used.
	/// </para>
	/// </summary>
	Modifiers []KeyboardModifier `json:"modifiers"`
	/// <summary>
	/// <para>
	/// A point to use relative to the top-left corner of element padding box. If not specified,
	/// uses some visible point of the element.
	/// </para>
	/// </summary>
	Position *FrameHoverOptionsPosition `json:"position"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type FrameInnerHTMLOptions struct {
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type FrameInnerTextOptions struct {
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type FrameIsCheckedOptions struct {
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type FrameIsDisabledOptions struct {
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type FrameIsEditableOptions struct {
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type FrameIsEnabledOptions struct {
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type FrameIsHiddenOptions struct {
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type FrameIsVisibleOptions struct {
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type FramePressOptions struct {
	/// <summary>
	/// <para>
	/// Time to wait between <c>keydown</c> and <c>keyup</c> in milliseconds. Defaults to
	/// 0.
	/// </para>
	/// </summary>
	Delay *float64 `json:"delay"`
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type FrameSelectOptionOptions struct {
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type FrameSetContentOptions struct {
	/// <summary>
	/// <para>
	/// Maximum operation time in milliseconds, defaults to 30 seconds, pass <c>0</c> to
	/// disable timeout. The default value can be changed by using the <see cref="BrowserContext.SetDefaultNavigationTimeout"/>,
	/// <see cref="BrowserContext.SetDefaultTimeout"/>, <see cref="Page.SetDefaultNavigationTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
	/// <summary>
	/// <para>When to consider operation succeeded, defaults to <c>load</c>. Events can be either:</para>
	/// <list type="bullet">
	/// <item><description>
	/// <c>'domcontentloaded'</c> - consider operation to be finished when the <c>DOMContentLoaded</c>
	/// event is fired.
	/// </description></item>
	/// <item><description>
	/// <c>'load'</c> - consider operation to be finished when the <c>load</c> event is
	/// fired.
	/// </description></item>
	/// <item><description>
	/// <c>'networkidle'</c> - consider operation to be finished when there are no network
	/// connections for at least <c>500</c> ms.
	/// </description></item>
	/// </list>
	/// </summary>
	WaitUntil *WaitUntilState `json:"waitUntil"`
}
type FrameSetInputFilesOptions struct {
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type FrameTapOptions struct {
	/// <summary>
	/// <para>
	/// Whether to bypass the <a href="./actionability.md">actionability</a> checks. Defaults
	/// to <c>false</c>.
	/// </para>
	/// </summary>
	Force *bool `json:"force"`
	/// <summary>
	/// <para>
	/// Modifier keys to press. Ensures that only these modifiers are pressed during the
	/// operation, and then restores current modifiers back. If not specified, currently
	/// pressed modifiers are used.
	/// </para>
	/// </summary>
	Modifiers []KeyboardModifier `json:"modifiers"`
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// A point to use relative to the top-left corner of element padding box. If not specified,
	/// uses some visible point of the element.
	/// </para>
	/// </summary>
	Position *FrameTapOptionsPosition `json:"position"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type FrameTextContentOptions struct {
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type FrameTypeOptions struct {
	/// <summary><para>Time to wait between key presses in milliseconds. Defaults to 0.</para></summary>
	Delay *float64 `json:"delay"`
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type FrameUncheckOptions struct {
	/// <summary>
	/// <para>
	/// Whether to bypass the <a href="./actionability.md">actionability</a> checks. Defaults
	/// to <c>false</c>.
	/// </para>
	/// </summary>
	Force *bool `json:"force"`
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type FrameWaitForFunctionOptions struct {
	/// <summary>
	/// <para>
	/// If <paramref name="polling"/> is <c>'raf'</c>, then <paramref name="expression"/>
	/// is constantly executed in <c>requestAnimationFrame</c> callback. If <paramref name="polling"/>
	/// is a number, then it is treated as an interval in milliseconds at which the function
	/// would be executed. Defaults to <c>raf</c>.
	/// </para>
	/// </summary>
	Polling interface{} `json:"polling"`
	/// <summary>
	/// <para>
	/// maximum time to wait for in milliseconds. Defaults to <c>30000</c> (30 seconds).
	/// Pass <c>0</c> to disable timeout. The default value can be changed by using the
	/// <see cref="BrowserContext.SetDefaultTimeout"/>.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type FrameWaitForLoadStateOptions struct {
	/// <summary>
	/// <para>
	/// Maximum operation time in milliseconds, defaults to 30 seconds, pass <c>0</c> to
	/// disable timeout. The default value can be changed by using the <see cref="BrowserContext.SetDefaultNavigationTimeout"/>,
	/// <see cref="BrowserContext.SetDefaultTimeout"/>, <see cref="Page.SetDefaultNavigationTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type FrameWaitForNavigationOptions struct {
	/// <summary>
	/// <para>
	/// Maximum operation time in milliseconds, defaults to 30 seconds, pass <c>0</c> to
	/// disable timeout. The default value can be changed by using the <see cref="BrowserContext.SetDefaultNavigationTimeout"/>,
	/// <see cref="BrowserContext.SetDefaultTimeout"/>, <see cref="Page.SetDefaultNavigationTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
	/// <summary>
	/// <para>
	/// A glob pattern, regex pattern or predicate receiving <see cref="URL"/> to match
	/// while waiting for the navigation.
	/// </para>
	/// </summary>
	URL interface{} `json:"url"`
	/// <summary>
	/// <para>When to consider operation succeeded, defaults to <c>load</c>. Events can be either:</para>
	/// <list type="bullet">
	/// <item><description>
	/// <c>'domcontentloaded'</c> - consider operation to be finished when the <c>DOMContentLoaded</c>
	/// event is fired.
	/// </description></item>
	/// <item><description>
	/// <c>'load'</c> - consider operation to be finished when the <c>load</c> event is
	/// fired.
	/// </description></item>
	/// <item><description>
	/// <c>'networkidle'</c> - consider operation to be finished when there are no network
	/// connections for at least <c>500</c> ms.
	/// </description></item>
	/// </list>
	/// </summary>
	WaitUntil *WaitUntilState `json:"waitUntil"`
}
type FrameWaitForSelectorOptions struct {
	/// <summary>
	/// <para>Defaults to <c>'visible'</c>. Can be either:</para>
	/// <list type="bullet">
	/// <item><description><c>'attached'</c> - wait for element to be present in DOM.</description></item>
	/// <item><description><c>'detached'</c> - wait for element to not be present in DOM.</description></item>
	/// <item><description>
	/// <c>'visible'</c> - wait for element to have non-empty bounding box and no <c>visibility:hidden</c>.
	/// Note that element without any content or with <c>display:none</c> has an empty bounding
	/// box and is not considered visible.
	/// </description></item>
	/// <item><description>
	/// <c>'hidden'</c> - wait for element to be either detached from DOM, or have an empty
	/// bounding box or <c>visibility:hidden</c>. This is opposite to the <c>'visible'</c>
	/// option.
	/// </description></item>
	/// </list>
	/// </summary>
	State *WaitForSelectorState `json:"state"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type JSHandleEvaluateOptions struct {
	/// <summary><para>Optional argument to pass to <paramref name="expression"/>.</para></summary>
	Arg interface{} `json:"arg"`
}
type JSHandleEvaluateHandleOptions struct {
	/// <summary><para>Optional argument to pass to <paramref name="expression"/>.</para></summary>
	Arg interface{} `json:"arg"`
}
type KeyboardPressOptions struct {
	/// <summary>
	/// <para>
	/// Time to wait between <c>keydown</c> and <c>keyup</c> in milliseconds. Defaults to
	/// 0.
	/// </para>
	/// </summary>
	Delay *float64 `json:"delay"`
}
type KeyboardTypeOptions struct {
	/// <summary><para>Time to wait between key presses in milliseconds. Defaults to 0.</para></summary>
	Delay *float64 `json:"delay"`
}
type MouseClickOptions struct {
	/// <summary><para>Defaults to <c>left</c>.</para></summary>
	Button *MouseButton `json:"button"`
	/// <summary><para>defaults to 1. See <see cref="UIEvent.detail"/>.</para></summary>
	ClickCount *int `json:"clickCount"`
	/// <summary>
	/// <para>
	/// Time to wait between <c>mousedown</c> and <c>mouseup</c> in milliseconds. Defaults
	/// to 0.
	/// </para>
	/// </summary>
	Delay *float64 `json:"delay"`
}
type MouseDblclickOptions struct {
	/// <summary><para>Defaults to <c>left</c>.</para></summary>
	Button *MouseButton `json:"button"`
	/// <summary>
	/// <para>
	/// Time to wait between <c>mousedown</c> and <c>mouseup</c> in milliseconds. Defaults
	/// to 0.
	/// </para>
	/// </summary>
	Delay *float64 `json:"delay"`
}
type MouseDownOptions struct {
	/// <summary><para>Defaults to <c>left</c>.</para></summary>
	Button *MouseButton `json:"button"`
	/// <summary><para>defaults to 1. See <see cref="UIEvent.detail"/>.</para></summary>
	ClickCount *int `json:"clickCount"`
}
type MouseMoveOptions struct {
	/// <summary><para>defaults to 1. Sends intermediate <c>mousemove</c> events.</para></summary>
	Steps *int `json:"steps"`
}
type MouseUpOptions struct {
	/// <summary><para>Defaults to <c>left</c>.</para></summary>
	Button *MouseButton `json:"button"`
	/// <summary><para>defaults to 1. See <see cref="UIEvent.detail"/>.</para></summary>
	ClickCount *int `json:"clickCount"`
}
type PageAddInitScriptOptions struct {
	/// <summary><para>Optional Script source to be evaluated in all pages in the browser context.</para></summary>
	Script *string `json:"script"`
	/// <summary><para>Optional Script path to be evaluated in all pages in the browser context.</para></summary>
	Path *string `json:"path"`
}
type PageAddScriptTagOptions struct {
	/// <summary><para>Raw JavaScript content to be injected into frame.</para></summary>
	Content *string `json:"content"`
	/// <summary>
	/// <para>
	/// Path to the JavaScript file to be injected into frame. If <c>path</c> is a relative
	/// path, then it is resolved relative to the current working directory.
	/// </para>
	/// </summary>
	Path *string `json:"path"`
	/// <summary>
	/// <para>
	/// Script type. Use 'module' in order to load a Javascript ES6 module. See <a href="https://developer.mozilla.org/en-US/docs/Web/HTML/Element/script">script</a>
	/// for more details.
	/// </para>
	/// </summary>
	Type *string `json:"type"`
	/// <summary><para>URL of a script to be added.</para></summary>
	URL *string `json:"url"`
}
type PageAddStyleTagOptions struct {
	/// <summary><para>Raw CSS content to be injected into frame.</para></summary>
	Content *string `json:"content"`
	/// <summary>
	/// <para>
	/// Path to the CSS file to be injected into frame. If <c>path</c> is a relative path,
	/// then it is resolved relative to the current working directory.
	/// </para>
	/// </summary>
	Path *string `json:"path"`
	/// <summary><para>URL of the <c>&lt;link&gt;</c> tag.</para></summary>
	URL *string `json:"url"`
}
type PageCheckOptions struct {
	/// <summary>
	/// <para>
	/// Whether to bypass the <a href="./actionability.md">actionability</a> checks. Defaults
	/// to <c>false</c>.
	/// </para>
	/// </summary>
	Force *bool `json:"force"`
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type PageClickOptions struct {
	/// <summary><para>Defaults to <c>left</c>.</para></summary>
	Button *MouseButton `json:"button"`
	/// <summary><para>defaults to 1. See <see cref="UIEvent.detail"/>.</para></summary>
	ClickCount *int `json:"clickCount"`
	/// <summary>
	/// <para>
	/// Time to wait between <c>mousedown</c> and <c>mouseup</c> in milliseconds. Defaults
	/// to 0.
	/// </para>
	/// </summary>
	Delay *float64 `json:"delay"`
	/// <summary>
	/// <para>
	/// Whether to bypass the <a href="./actionability.md">actionability</a> checks. Defaults
	/// to <c>false</c>.
	/// </para>
	/// </summary>
	Force *bool `json:"force"`
	/// <summary>
	/// <para>
	/// Modifier keys to press. Ensures that only these modifiers are pressed during the
	/// operation, and then restores current modifiers back. If not specified, currently
	/// pressed modifiers are used.
	/// </para>
	/// </summary>
	Modifiers []KeyboardModifier `json:"modifiers"`
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// A point to use relative to the top-left corner of element padding box. If not specified,
	/// uses some visible point of the element.
	/// </para>
	/// </summary>
	Position *PageClickOptionsPosition `json:"position"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type PagePosition struct {
	/// <summary><para></para></summary>
	X *float64 `json:"x"`
	/// <summary><para></para></summary>
	Y *float64 `json:"y"`
}
type PageCloseOptions struct {
	/// <summary>
	/// <para>
	/// Defaults to <c>false</c>. Whether to run the <a href="https://developer.mozilla.org/en-US/docs/Web/Events/beforeunload">before
	/// unload</a> page handlers.
	/// </para>
	/// </summary>
	RunBeforeUnload *bool `json:"runBeforeUnload"`
}
type PageDblclickOptions struct {
	/// <summary><para>Defaults to <c>left</c>.</para></summary>
	Button *MouseButton `json:"button"`
	/// <summary>
	/// <para>
	/// Time to wait between <c>mousedown</c> and <c>mouseup</c> in milliseconds. Defaults
	/// to 0.
	/// </para>
	/// </summary>
	Delay *float64 `json:"delay"`
	/// <summary>
	/// <para>
	/// Whether to bypass the <a href="./actionability.md">actionability</a> checks. Defaults
	/// to <c>false</c>.
	/// </para>
	/// </summary>
	Force *bool `json:"force"`
	/// <summary>
	/// <para>
	/// Modifier keys to press. Ensures that only these modifiers are pressed during the
	/// operation, and then restores current modifiers back. If not specified, currently
	/// pressed modifiers are used.
	/// </para>
	/// </summary>
	Modifiers []KeyboardModifier `json:"modifiers"`
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// A point to use relative to the top-left corner of element padding box. If not specified,
	/// uses some visible point of the element.
	/// </para>
	/// </summary>
	Position *PageDblclickOptionsPosition `json:"position"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type PageDispatchEventOptions struct {
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type PageEmulateMediaOptions struct {
	/// <summary>
	/// <para>
	/// Emulates <c>'prefers-colors-scheme'</c> media feature, supported values are <c>'light'</c>,
	/// <c>'dark'</c>, <c>'no-preference'</c>. Passing <c>null</c> disables color scheme
	/// emulation.
	/// </para>
	/// </summary>
	ColorScheme *ColorScheme `json:"colorScheme"`
	/// <summary>
	/// <para>
	/// Changes the CSS media type of the page. The only allowed values are <c>'screen'</c>,
	/// <c>'print'</c> and <c>null</c>. Passing <c>null</c> disables CSS media emulation.
	/// </para>
	/// </summary>
	Media *Media `json:"media"`
}
type PageEvalOnSelectorOptions struct {
	/// <summary><para>Optional argument to pass to <paramref name="expression"/>.</para></summary>
	Arg interface{} `json:"arg"`
}
type PageEvalOnSelectorAllOptions struct {
	/// <summary><para>Optional argument to pass to <paramref name="expression"/>.</para></summary>
	Arg interface{} `json:"arg"`
}
type PageEvaluateOptions struct {
	/// <summary><para>Optional argument to pass to <paramref name="expression"/>.</para></summary>
	Arg interface{} `json:"arg"`
}
type PageEvaluateHandleOptions struct {
	/// <summary><para>Optional argument to pass to <paramref name="expression"/>.</para></summary>
	Arg interface{} `json:"arg"`
}
type PageExposeBindingOptions struct {
	/// <summary>
	/// <para>
	/// Whether to pass the argument as a handle, instead of passing by value. When passing
	/// a handle, only one argument is supported. When passing by value, multiple arguments
	/// are supported.
	/// </para>
	/// </summary>
	Handle *bool `json:"handle"`
}
type PageFillOptions struct {
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type PageFocusOptions struct {
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type PageGetAttributeOptions struct {
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type PageGoBackOptions struct {
	/// <summary>
	/// <para>
	/// Maximum operation time in milliseconds, defaults to 30 seconds, pass <c>0</c> to
	/// disable timeout. The default value can be changed by using the <see cref="BrowserContext.SetDefaultNavigationTimeout"/>,
	/// <see cref="BrowserContext.SetDefaultTimeout"/>, <see cref="Page.SetDefaultNavigationTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
	/// <summary>
	/// <para>When to consider operation succeeded, defaults to <c>load</c>. Events can be either:</para>
	/// <list type="bullet">
	/// <item><description>
	/// <c>'domcontentloaded'</c> - consider operation to be finished when the <c>DOMContentLoaded</c>
	/// event is fired.
	/// </description></item>
	/// <item><description>
	/// <c>'load'</c> - consider operation to be finished when the <c>load</c> event is
	/// fired.
	/// </description></item>
	/// <item><description>
	/// <c>'networkidle'</c> - consider operation to be finished when there are no network
	/// connections for at least <c>500</c> ms.
	/// </description></item>
	/// </list>
	/// </summary>
	WaitUntil *WaitUntilState `json:"waitUntil"`
}
type PageGoForwardOptions struct {
	/// <summary>
	/// <para>
	/// Maximum operation time in milliseconds, defaults to 30 seconds, pass <c>0</c> to
	/// disable timeout. The default value can be changed by using the <see cref="BrowserContext.SetDefaultNavigationTimeout"/>,
	/// <see cref="BrowserContext.SetDefaultTimeout"/>, <see cref="Page.SetDefaultNavigationTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
	/// <summary>
	/// <para>When to consider operation succeeded, defaults to <c>load</c>. Events can be either:</para>
	/// <list type="bullet">
	/// <item><description>
	/// <c>'domcontentloaded'</c> - consider operation to be finished when the <c>DOMContentLoaded</c>
	/// event is fired.
	/// </description></item>
	/// <item><description>
	/// <c>'load'</c> - consider operation to be finished when the <c>load</c> event is
	/// fired.
	/// </description></item>
	/// <item><description>
	/// <c>'networkidle'</c> - consider operation to be finished when there are no network
	/// connections for at least <c>500</c> ms.
	/// </description></item>
	/// </list>
	/// </summary>
	WaitUntil *WaitUntilState `json:"waitUntil"`
}
type PageGotoOptions struct {
	/// <summary>
	/// <para>
	/// Referer header value. If provided it will take preference over the referer header
	/// value set by <see cref="Page.SetExtraHttpHeaders"/>.
	/// </para>
	/// </summary>
	Referer *string `json:"referer"`
	/// <summary>
	/// <para>
	/// Maximum operation time in milliseconds, defaults to 30 seconds, pass <c>0</c> to
	/// disable timeout. The default value can be changed by using the <see cref="BrowserContext.SetDefaultNavigationTimeout"/>,
	/// <see cref="BrowserContext.SetDefaultTimeout"/>, <see cref="Page.SetDefaultNavigationTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
	/// <summary>
	/// <para>When to consider operation succeeded, defaults to <c>load</c>. Events can be either:</para>
	/// <list type="bullet">
	/// <item><description>
	/// <c>'domcontentloaded'</c> - consider operation to be finished when the <c>DOMContentLoaded</c>
	/// event is fired.
	/// </description></item>
	/// <item><description>
	/// <c>'load'</c> - consider operation to be finished when the <c>load</c> event is
	/// fired.
	/// </description></item>
	/// <item><description>
	/// <c>'networkidle'</c> - consider operation to be finished when there are no network
	/// connections for at least <c>500</c> ms.
	/// </description></item>
	/// </list>
	/// </summary>
	WaitUntil *WaitUntilState `json:"waitUntil"`
}
type PageHoverOptions struct {
	/// <summary>
	/// <para>
	/// Whether to bypass the <a href="./actionability.md">actionability</a> checks. Defaults
	/// to <c>false</c>.
	/// </para>
	/// </summary>
	Force *bool `json:"force"`
	/// <summary>
	/// <para>
	/// Modifier keys to press. Ensures that only these modifiers are pressed during the
	/// operation, and then restores current modifiers back. If not specified, currently
	/// pressed modifiers are used.
	/// </para>
	/// </summary>
	Modifiers []KeyboardModifier `json:"modifiers"`
	/// <summary>
	/// <para>
	/// A point to use relative to the top-left corner of element padding box. If not specified,
	/// uses some visible point of the element.
	/// </para>
	/// </summary>
	Position *PageHoverOptionsPosition `json:"position"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type PageInnerHTMLOptions struct {
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type PageInnerTextOptions struct {
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type PageIsCheckedOptions struct {
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type PageIsDisabledOptions struct {
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type PageIsEditableOptions struct {
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type PageIsEnabledOptions struct {
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type PageIsHiddenOptions struct {
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type PageIsVisibleOptions struct {
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type PagePDFmarginOption struct {
	/// <summary><para>Top margin, accepts values labeled with units. Defaults to 0.</para></summary>
	Top *string `json:"top"`
	/// <summary><para>Right margin, accepts values labeled with units. Defaults to 0.</para></summary>
	Right *string `json:"right"`
	/// <summary><para>Bottom margin, accepts values labeled with units. Defaults to 0.</para></summary>
	Bottom *string `json:"bottom"`
	/// <summary><para>Left margin, accepts values labeled with units. Defaults to 0.</para></summary>
	Left *string `json:"left"`
}
type PagePdfOptions struct {
	/// <summary><para>Display header and footer. Defaults to <c>false</c>.</para></summary>
	DisplayHeaderFooter *bool `json:"displayHeaderFooter"`
	/// <summary>
	/// <para>
	/// HTML template for the print footer. Should use the same format as the <paramref
	/// name="headerTemplate"/>.
	/// </para>
	/// </summary>
	FooterTemplate *string `json:"footerTemplate"`
	/// <summary>
	/// <para>
	/// Paper format. If set, takes priority over <paramref name="width"/> or <paramref
	/// name="height"/> options. Defaults to 'Letter'.
	/// </para>
	/// </summary>
	Format *string `json:"format"`
	/// <summary>
	/// <para>
	/// HTML template for the print header. Should be valid HTML markup with following classes
	/// used to inject printing values into them:
	/// </para>
	/// <list type="bullet">
	/// <item><description><c>'date'</c> formatted print date</description></item>
	/// <item><description><c>'title'</c> document title</description></item>
	/// <item><description><c>'url'</c> document location</description></item>
	/// <item><description><c>'pageNumber'</c> current page number</description></item>
	/// <item><description><c>'totalPages'</c> total pages in the document</description></item>
	/// </list>
	/// </summary>
	HeaderTemplate *string `json:"headerTemplate"`
	/// <summary><para>Paper orientation. Defaults to <c>false</c>.</para></summary>
	Landscape *bool `json:"landscape"`
	/// <summary>
	/// <para>
	/// Paper ranges to print, e.g., '1-5, 8, 11-13'. Defaults to the empty string, which
	/// means print all pages.
	/// </para>
	/// </summary>
	PageRanges *string `json:"pageRanges"`
	/// <summary>
	/// <para>
	/// The file path to save the PDF to. If <paramref name="path"/> is a relative path,
	/// then it is resolved relative to the current working directory. If no path is provided,
	/// the PDF won't be saved to the disk.
	/// </para>
	/// </summary>
	Path *string `json:"path"`
	/// <summary>
	/// <para>
	/// Give any CSS <c>@page</c> size declared in the page priority over what is declared
	/// in <paramref name="width"/> and <paramref name="height"/> or <paramref name="format"/>
	/// options. Defaults to <c>false</c>, which will scale the content to fit the paper
	/// size.
	/// </para>
	/// </summary>
	PreferCSSPageSize *bool `json:"preferCSSPageSize"`
	/// <summary><para>Print background graphics. Defaults to <c>false</c>.</para></summary>
	PrintBackground *bool `json:"printBackground"`
	/// <summary>
	/// <para>
	/// Scale of the webpage rendering. Defaults to <c>1</c>. Scale amount must be between
	/// 0.1 and 2.
	/// </para>
	/// </summary>
	Scale *float64 `json:"scale"`
	/// <summary><para>Paper margins, defaults to none.</para></summary>
	Margin *PagePDFmarginOption `json:"margin"`
}
type PagePressOptions struct {
	/// <summary>
	/// <para>
	/// Time to wait between <c>keydown</c> and <c>keyup</c> in milliseconds. Defaults to
	/// 0.
	/// </para>
	/// </summary>
	Delay *float64 `json:"delay"`
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type PageReloadOptions struct {
	/// <summary>
	/// <para>
	/// Maximum operation time in milliseconds, defaults to 30 seconds, pass <c>0</c> to
	/// disable timeout. The default value can be changed by using the <see cref="BrowserContext.SetDefaultNavigationTimeout"/>,
	/// <see cref="BrowserContext.SetDefaultTimeout"/>, <see cref="Page.SetDefaultNavigationTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
	/// <summary>
	/// <para>When to consider operation succeeded, defaults to <c>load</c>. Events can be either:</para>
	/// <list type="bullet">
	/// <item><description>
	/// <c>'domcontentloaded'</c> - consider operation to be finished when the <c>DOMContentLoaded</c>
	/// event is fired.
	/// </description></item>
	/// <item><description>
	/// <c>'load'</c> - consider operation to be finished when the <c>load</c> event is
	/// fired.
	/// </description></item>
	/// <item><description>
	/// <c>'networkidle'</c> - consider operation to be finished when there are no network
	/// connections for at least <c>500</c> ms.
	/// </description></item>
	/// </list>
	/// </summary>
	WaitUntil *WaitUntilState `json:"waitUntil"`
}
type PageRouteOptions struct {
	/// <summary><para>handler function to route the request.</para></summary>
	Handler func(Route, Request) `json:"handler"`
}
type PageScreenshotOptions struct {
	/// <summary>
	/// <para>
	/// An object which specifies clipping of the resulting image. Should have the following
	/// fields:
	/// </para>
	/// </summary>
	Clip *PageScreenshotOptionsClip `json:"clip"`
	/// <summary>
	/// <para>
	/// When true, takes a screenshot of the full scrollable page, instead of the currently
	/// visible viewport. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	FullPage *bool `json:"fullPage"`
	/// <summary>
	/// <para>
	/// Hides default white background and allows capturing screenshots with transparency.
	/// Not applicable to <c>jpeg</c> images. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	OmitBackground *bool `json:"omitBackground"`
	/// <summary>
	/// <para>
	/// The file path to save the image to. The screenshot type will be inferred from file
	/// extension. If <paramref name="path"/> is a relative path, then it is resolved relative
	/// to the current working directory. If no path is provided, the image won't be saved
	/// to the disk.
	/// </para>
	/// </summary>
	Path *string `json:"path"`
	/// <summary><para>The quality of the image, between 0-100. Not applicable to <c>png</c> images.</para></summary>
	Quality *int `json:"quality"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
	/// <summary><para>Specify screenshot type, defaults to <c>png</c>.</para></summary>
	Type *ScreenshotType `json:"type"`
}
type PageClip struct {
	/// <summary><para>x-coordinate of top-left corner of clip area</para></summary>
	X *float64 `json:"x"`
	/// <summary><para>y-coordinate of top-left corner of clip area</para></summary>
	Y *float64 `json:"y"`
	/// <summary><para>width of clipping area</para></summary>
	Width *float64 `json:"width"`
	/// <summary><para>height of clipping area</para></summary>
	Height *float64 `json:"height"`
}
type PageSelectOptionOptions struct {
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type PageSetContentOptions struct {
	/// <summary>
	/// <para>
	/// Maximum operation time in milliseconds, defaults to 30 seconds, pass <c>0</c> to
	/// disable timeout. The default value can be changed by using the <see cref="BrowserContext.SetDefaultNavigationTimeout"/>,
	/// <see cref="BrowserContext.SetDefaultTimeout"/>, <see cref="Page.SetDefaultNavigationTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
	/// <summary>
	/// <para>When to consider operation succeeded, defaults to <c>load</c>. Events can be either:</para>
	/// <list type="bullet">
	/// <item><description>
	/// <c>'domcontentloaded'</c> - consider operation to be finished when the <c>DOMContentLoaded</c>
	/// event is fired.
	/// </description></item>
	/// <item><description>
	/// <c>'load'</c> - consider operation to be finished when the <c>load</c> event is
	/// fired.
	/// </description></item>
	/// <item><description>
	/// <c>'networkidle'</c> - consider operation to be finished when there are no network
	/// connections for at least <c>500</c> ms.
	/// </description></item>
	/// </list>
	/// </summary>
	WaitUntil *WaitUntilState `json:"waitUntil"`
}
type PageSetInputFilesOptions struct {
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type PageTapOptions struct {
	/// <summary>
	/// <para>
	/// Whether to bypass the <a href="./actionability.md">actionability</a> checks. Defaults
	/// to <c>false</c>.
	/// </para>
	/// </summary>
	Force *bool `json:"force"`
	/// <summary>
	/// <para>
	/// Modifier keys to press. Ensures that only these modifiers are pressed during the
	/// operation, and then restores current modifiers back. If not specified, currently
	/// pressed modifiers are used.
	/// </para>
	/// </summary>
	Modifiers []KeyboardModifier `json:"modifiers"`
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// A point to use relative to the top-left corner of element padding box. If not specified,
	/// uses some visible point of the element.
	/// </para>
	/// </summary>
	Position *PageTapOptionsPosition `json:"position"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type PageTextContentOptions struct {
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type PageTypeOptions struct {
	/// <summary><para>Time to wait between key presses in milliseconds. Defaults to 0.</para></summary>
	Delay *float64 `json:"delay"`
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type PageUncheckOptions struct {
	/// <summary>
	/// <para>
	/// Whether to bypass the <a href="./actionability.md">actionability</a> checks. Defaults
	/// to <c>false</c>.
	/// </para>
	/// </summary>
	Force *bool `json:"force"`
	/// <summary>
	/// <para>
	/// Actions that initiate navigations are waiting for these navigations to happen and
	/// for pages to start loading. You can opt out of waiting via setting this flag. You
	/// would only need this option in the exceptional cases such as navigating to inaccessible
	/// pages. Defaults to <c>false</c>.
	/// </para>
	/// </summary>
	NoWaitAfter *bool `json:"noWaitAfter"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type PageUnrouteOptions struct {
	/// <summary><para>Optional handler function to route the request.</para></summary>
	Handler func(Route, Request) `json:"handler"`
}

// Result of calling <see cref="Page.ViewportSize" />.
type PageViewportSizeResult struct {
	/// <summary><para>page width in pixels.</para></summary>
	Width *int `json:"width"`
	/// <summary><para>page height in pixels.</para></summary>
	Height *int `json:"height"`
}
type PageWaitForFunctionOptions struct {
	/// <summary>
	/// <para>
	/// If <paramref name="polling"/> is <c>'raf'</c>, then <paramref name="expression"/>
	/// is constantly executed in <c>requestAnimationFrame</c> callback. If <paramref name="polling"/>
	/// is a number, then it is treated as an interval in milliseconds at which the function
	/// would be executed. Defaults to <c>raf</c>.
	/// </para>
	/// </summary>
	Polling interface{} `json:"polling"`
	/// <summary>
	/// <para>
	/// maximum time to wait for in milliseconds. Defaults to <c>30000</c> (30 seconds).
	/// Pass <c>0</c> to disable timeout. The default value can be changed by using the
	/// <see cref="BrowserContext.SetDefaultTimeout"/>.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type PageWaitForLoadStateOptions struct {
	/// <summary>
	/// <para>
	/// Maximum operation time in milliseconds, defaults to 30 seconds, pass <c>0</c> to
	/// disable timeout. The default value can be changed by using the <see cref="BrowserContext.SetDefaultNavigationTimeout"/>,
	/// <see cref="BrowserContext.SetDefaultTimeout"/>, <see cref="Page.SetDefaultNavigationTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type PageWaitForNavigationOptions struct {
	/// <summary>
	/// <para>
	/// Maximum operation time in milliseconds, defaults to 30 seconds, pass <c>0</c> to
	/// disable timeout. The default value can be changed by using the <see cref="BrowserContext.SetDefaultNavigationTimeout"/>,
	/// <see cref="BrowserContext.SetDefaultTimeout"/>, <see cref="Page.SetDefaultNavigationTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
	/// <summary>
	/// <para>
	/// A glob pattern, regex pattern or predicate receiving <see cref="URL"/> to match
	/// while waiting for the navigation.
	/// </para>
	/// </summary>
	URL interface{} `json:"url"`
	/// <summary>
	/// <para>When to consider operation succeeded, defaults to <c>load</c>. Events can be either:</para>
	/// <list type="bullet">
	/// <item><description>
	/// <c>'domcontentloaded'</c> - consider operation to be finished when the <c>DOMContentLoaded</c>
	/// event is fired.
	/// </description></item>
	/// <item><description>
	/// <c>'load'</c> - consider operation to be finished when the <c>load</c> event is
	/// fired.
	/// </description></item>
	/// <item><description>
	/// <c>'networkidle'</c> - consider operation to be finished when there are no network
	/// connections for at least <c>500</c> ms.
	/// </description></item>
	/// </list>
	/// </summary>
	WaitUntil *WaitUntilState `json:"waitUntil"`
}
type PageWaitForRequestOptions struct {
	/// <summary>
	/// <para>
	/// Maximum wait time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable
	/// the timeout. The default value can be changed by using the <see cref="Page.SetDefaultTimeout"/>
	/// method.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type PageWaitForResponseOptions struct {
	/// <summary>
	/// <para>
	/// Maximum wait time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable
	/// the timeout. The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}
type PageWaitForSelectorOptions struct {
	/// <summary>
	/// <para>Defaults to <c>'visible'</c>. Can be either:</para>
	/// <list type="bullet">
	/// <item><description><c>'attached'</c> - wait for element to be present in DOM.</description></item>
	/// <item><description><c>'detached'</c> - wait for element to not be present in DOM.</description></item>
	/// <item><description>
	/// <c>'visible'</c> - wait for element to have non-empty bounding box and no <c>visibility:hidden</c>.
	/// Note that element without any content or with <c>display:none</c> has an empty bounding
	/// box and is not considered visible.
	/// </description></item>
	/// <item><description>
	/// <c>'hidden'</c> - wait for element to be either detached from DOM, or have an empty
	/// bounding box or <c>visibility:hidden</c>. This is opposite to the <c>'visible'</c>
	/// option.
	/// </description></item>
	/// </list>
	/// </summary>
	State *WaitForSelectorState `json:"state"`
	/// <summary>
	/// <para>
	/// Maximum time in milliseconds, defaults to 30 seconds, pass <c>0</c> to disable timeout.
	/// The default value can be changed by using the <see cref="BrowserContext.SetDefaultTimeout"/>
	/// or <see cref="Page.SetDefaultTimeout"/> methods.
	/// </para>
	/// </summary>
	Timeout *float64 `json:"timeout"`
}

// Result of calling <see cref="Request.Timing" />.
type RequestTimingResult struct {
	/// <summary><para>Request start time in milliseconds elapsed since January 1, 1970 00:00:00 UTC</para></summary>
	StartTime *float64 `json:"startTime"`
	/// <summary>
	/// <para>
	/// Time immediately before the browser starts the domain name lookup for the resource.
	/// The value is given in milliseconds relative to <c>startTime</c>, -1 if not available.
	/// </para>
	/// </summary>
	DomainLookupStart *float64 `json:"domainLookupStart"`
	/// <summary>
	/// <para>
	/// Time immediately after the browser starts the domain name lookup for the resource.
	/// The value is given in milliseconds relative to <c>startTime</c>, -1 if not available.
	/// </para>
	/// </summary>
	DomainLookupEnd *float64 `json:"domainLookupEnd"`
	/// <summary>
	/// <para>
	/// Time immediately before the user agent starts establishing the connection to the
	/// server to retrieve the resource. The value is given in milliseconds relative to
	/// <c>startTime</c>, -1 if not available.
	/// </para>
	/// </summary>
	ConnectStart *float64 `json:"connectStart"`
	/// <summary>
	/// <para>
	/// Time immediately before the browser starts the handshake process to secure the current
	/// connection. The value is given in milliseconds relative to <c>startTime</c>, -1
	/// if not available.
	/// </para>
	/// </summary>
	SecureConnectionStart *float64 `json:"secureConnectionStart"`
	/// <summary>
	/// <para>
	/// Time immediately before the user agent starts establishing the connection to the
	/// server to retrieve the resource. The value is given in milliseconds relative to
	/// <c>startTime</c>, -1 if not available.
	/// </para>
	/// </summary>
	ConnectEnd *float64 `json:"connectEnd"`
	/// <summary>
	/// <para>
	/// Time immediately before the browser starts requesting the resource from the server,
	/// cache, or local resource. The value is given in milliseconds relative to <c>startTime</c>,
	/// -1 if not available.
	/// </para>
	/// </summary>
	RequestStart *float64 `json:"requestStart"`
	/// <summary>
	/// <para>
	/// Time immediately after the browser starts requesting the resource from the server,
	/// cache, or local resource. The value is given in milliseconds relative to <c>startTime</c>,
	/// -1 if not available.
	/// </para>
	/// </summary>
	ResponseStart *float64 `json:"responseStart"`
	/// <summary>
	/// <para>
	/// Time immediately after the browser receives the last byte of the resource or immediately
	/// before the transport connection is closed, whichever comes first. The value is given
	/// in milliseconds relative to <c>startTime</c>, -1 if not available.
	/// </para>
	/// </summary>
	ResponseEnd *float64 `json:"responseEnd"`
}
type RouteAbortOptions struct {
	/// <summary>
	/// <para>Optional error code. Defaults to <c>failed</c>, could be one of the following:</para>
	/// <list type="bullet">
	/// <item><description><c>'aborted'</c> - An operation was aborted (due to user action)</description></item>
	/// <item><description>
	/// <c>'accessdenied'</c> - Permission to access a resource, other than the network,
	/// was denied
	/// </description></item>
	/// <item><description>
	/// <c>'addressunreachable'</c> - The IP address is unreachable. This usually means
	/// that there is no route to the specified host or network.
	/// </description></item>
	/// <item><description><c>'blockedbyclient'</c> - The client chose to block the request.</description></item>
	/// <item><description>
	/// <c>'blockedbyresponse'</c> - The request failed because the response was delivered
	/// along with requirements which are not met ('X-Frame-Options' and 'Content-Security-Policy'
	/// ancestor checks, for instance).
	/// </description></item>
	/// <item><description>
	/// <c>'connectionaborted'</c> - A connection timed out as a result of not receiving
	/// an ACK for data sent.
	/// </description></item>
	/// <item><description><c>'connectionclosed'</c> - A connection was closed (corresponding to a TCP FIN).</description></item>
	/// <item><description><c>'connectionfailed'</c> - A connection attempt failed.</description></item>
	/// <item><description><c>'connectionrefused'</c> - A connection attempt was refused.</description></item>
	/// <item><description><c>'connectionreset'</c> - A connection was reset (corresponding to a TCP RST).</description></item>
	/// <item><description><c>'internetdisconnected'</c> - The Internet connection has been lost.</description></item>
	/// <item><description><c>'namenotresolved'</c> - The host name could not be resolved.</description></item>
	/// <item><description><c>'timedout'</c> - An operation timed out.</description></item>
	/// <item><description><c>'failed'</c> - A generic failure occurred.</description></item>
	/// </list>
	/// </summary>
	ErrorCode *string `json:"errorCode"`
}
type RouteContinueOptions struct {
	/// <summary><para>If set changes the request HTTP headers. Header values will be converted to a string.</para></summary>
	Headers map[string]string `json:"headers"`
	/// <summary><para>If set changes the request method (e.g. GET or POST)</para></summary>
	Method *string `json:"method"`
	/// <summary><para>If set changes the post data of request</para></summary>
	PostData interface{} `json:"postData"`
	/// <summary><para>If set changes the request URL. New URL must have same protocol as original one.</para></summary>
	URL *string `json:"url"`
}
type RouteFulfillOptions struct {
	/// <summary><para>Response body.</para></summary>
	Body interface{} `json:"body"`
	/// <summary><para>If set, equals to setting <c>Content-Type</c> response header.</para></summary>
	ContentType *string `json:"contentType"`
	/// <summary><para>Response headers. Header values will be converted to a string.</para></summary>
	Headers map[string]string `json:"headers"`
	/// <summary>
	/// <para>
	/// File path to respond with. The content type will be inferred from file extension.
	/// If <c>path</c> is a relative path, then it is resolved relative to the current working
	/// directory.
	/// </para>
	/// </summary>
	Path *string `json:"path"`
	/// <summary><para>Response status code, defaults to <c>200</c>.</para></summary>
	Status *int `json:"status"`
}
type SelectorsRegisterOptions struct {
	/// <summary>
	/// <para>
	/// Whether to run this selector engine in isolated JavaScript environment. This environment
	/// has access to the same DOM, but not any JavaScript objects from the frame's scripts.
	/// Defaults to <c>false</c>. Note that running as a content script is not guaranteed
	/// when this engine is used together with other registered engines.
	/// </para>
	/// </summary>
	ContentScript *bool `json:"contentScript"`
}
type FrameReceivedPayload struct {
	/// <summary><para>frame payload</para></summary>
	Payload []byte `json:"payload"`
}
type FrameSentPayload struct {
	/// <summary><para>frame payload</para></summary>
	Payload []byte `json:"payload"`
}
type WorkerEvaluateOptions struct {
	/// <summary><para>Optional argument to pass to <paramref name="expression"/>.</para></summary>
	Arg interface{} `json:"arg"`
}
type WorkerEvaluateHandleOptions struct {
	/// <summary><para>Optional argument to pass to <paramref name="expression"/>.</para></summary>
	Arg interface{} `json:"arg"`
}
type BrowserNewContextOptionsGeolocation struct {
	/// <summary><para>Latitude between -90 and 90.</para></summary>
	Latitude *float64 `json:"latitude"`
	/// <summary><para>Longitude between -180 and 180.</para></summary>
	Longitude *float64 `json:"longitude"`
	/// <summary><para>Non-negative accuracy value. Defaults to <c>0</c>.</para></summary>
	Accuracy *float64 `json:"accuracy"`
}
type BrowserNewContextOptionsHttpCredentials struct {
	/// <summary><para></para></summary>
	Username *string `json:"username"`
	/// <summary><para></para></summary>
	Password *string `json:"password"`
}
type BrowserNewContextOptionsProxy struct {
	/// <summary>
	/// <para>
	/// Proxy to be used for all requests. HTTP and SOCKS proxies are supported, for example
	/// <c>http://myproxy.com:3128</c> or <c>socks5://myproxy.com:3128</c>. Short form <c>myproxy.com:3128</c>
	/// is considered an HTTP proxy.
	/// </para>
	/// </summary>
	Server *string `json:"server"`
	/// <summary>
	/// <para>
	/// Optional coma-separated domains to bypass proxy, for example <c>".com, chromium.org,
	/// .domain.com"</c>.
	/// </para>
	/// </summary>
	Bypass *string `json:"bypass"`
	/// <summary><para>Optional username to use if HTTP proxy requires authentication.</para></summary>
	Username *string `json:"username"`
	/// <summary><para>Optional password to use if HTTP proxy requires authentication.</para></summary>
	Password *string `json:"password"`
}
type BrowserNewContextOptionsRecordVideo struct {
	/// <summary><para>Path to the directory to put videos into.</para></summary>
	Dir *string `json:"dir"`
	/// <summary>
	/// <para>
	/// Optional dimensions of the recorded videos. If not specified the size will be equal
	/// to <c>viewport</c> scaled down to fit into 800x800. If <c>viewport</c> is not configured
	/// explicitly the video size defaults to 800x450. Actual picture of each page will
	/// be scaled down if necessary to fit the specified size.
	/// </para>
	/// </summary>
	Size *BrowserNewContextOptionsRecordVideoSize `json:"size"`
}
type BrowserNewContextOptionsViewport struct {
	/// <summary><para>page width in pixels.</para></summary>
	Width *int `json:"width"`
	/// <summary><para>page height in pixels.</para></summary>
	Height *int `json:"height"`
}
type BrowserRecordVideoSize struct {
	/// <summary><para>Video frame width.</para></summary>
	Width *int `json:"width"`
	/// <summary><para>Video frame height.</para></summary>
	Height *int `json:"height"`
}
type BrowserNewPageOptionsGeolocation struct {
	/// <summary><para>Latitude between -90 and 90.</para></summary>
	Latitude *float64 `json:"latitude"`
	/// <summary><para>Longitude between -180 and 180.</para></summary>
	Longitude *float64 `json:"longitude"`
	/// <summary><para>Non-negative accuracy value. Defaults to <c>0</c>.</para></summary>
	Accuracy *float64 `json:"accuracy"`
}
type BrowserNewPageOptionsHttpCredentials struct {
	/// <summary><para></para></summary>
	Username *string `json:"username"`
	/// <summary><para></para></summary>
	Password *string `json:"password"`
}
type BrowserNewPageOptionsProxy struct {
	/// <summary>
	/// <para>
	/// Proxy to be used for all requests. HTTP and SOCKS proxies are supported, for example
	/// <c>http://myproxy.com:3128</c> or <c>socks5://myproxy.com:3128</c>. Short form <c>myproxy.com:3128</c>
	/// is considered an HTTP proxy.
	/// </para>
	/// </summary>
	Server *string `json:"server"`
	/// <summary>
	/// <para>
	/// Optional coma-separated domains to bypass proxy, for example <c>".com, chromium.org,
	/// .domain.com"</c>.
	/// </para>
	/// </summary>
	Bypass *string `json:"bypass"`
	/// <summary><para>Optional username to use if HTTP proxy requires authentication.</para></summary>
	Username *string `json:"username"`
	/// <summary><para>Optional password to use if HTTP proxy requires authentication.</para></summary>
	Password *string `json:"password"`
}
type BrowserNewPageOptionsRecordVideo struct {
	/// <summary><para>Path to the directory to put videos into.</para></summary>
	Dir *string `json:"dir"`
	/// <summary>
	/// <para>
	/// Optional dimensions of the recorded videos. If not specified the size will be equal
	/// to <c>viewport</c> scaled down to fit into 800x800. If <c>viewport</c> is not configured
	/// explicitly the video size defaults to 800x450. Actual picture of each page will
	/// be scaled down if necessary to fit the specified size.
	/// </para>
	/// </summary>
	Size *BrowserNewPageOptionsRecordVideoSize `json:"size"`
}
type BrowserNewPageOptionsViewport struct {
	/// <summary><para>page width in pixels.</para></summary>
	Width *int `json:"width"`
	/// <summary><para>page height in pixels.</para></summary>
	Height *int `json:"height"`
}
type BrowserContextStorageStateResultCookies struct {
	/// <summary><para></para></summary>
	Name *string `json:"name"`
	/// <summary><para></para></summary>
	Value *string `json:"value"`
	/// <summary><para></para></summary>
	Domain *string `json:"domain"`
	/// <summary><para></para></summary>
	Path *string `json:"path"`
	/// <summary><para>Unix time in seconds.</para></summary>
	Expires *float64 `json:"expires"`
	/// <summary><para></para></summary>
	HttpOnly *bool `json:"httpOnly"`
	/// <summary><para></para></summary>
	Secure *bool `json:"secure"`
	/// <summary><para></para></summary>
	SameSite *SameSiteAttribute `json:"sameSite"`
}
type BrowserContextStorageStateResultOrigins struct {
	/// <summary><para></para></summary>
	Origin *string `json:"origin"`
	/// <summary><para></para></summary>
	LocalStorage []BrowserContextStorageStateResultOriginsLocalStorage `json:"localStorage"`
}
type BrowserTypeLaunchOptionsProxy struct {
	/// <summary>
	/// <para>
	/// Proxy to be used for all requests. HTTP and SOCKS proxies are supported, for example
	/// <c>http://myproxy.com:3128</c> or <c>socks5://myproxy.com:3128</c>. Short form <c>myproxy.com:3128</c>
	/// is considered an HTTP proxy.
	/// </para>
	/// </summary>
	Server *string `json:"server"`
	/// <summary>
	/// <para>
	/// Optional coma-separated domains to bypass proxy, for example <c>".com, chromium.org,
	/// .domain.com"</c>.
	/// </para>
	/// </summary>
	Bypass *string `json:"bypass"`
	/// <summary><para>Optional username to use if HTTP proxy requires authentication.</para></summary>
	Username *string `json:"username"`
	/// <summary><para>Optional password to use if HTTP proxy requires authentication.</para></summary>
	Password *string `json:"password"`
}
type BrowserTypeLaunchPersistentContextOptionsGeolocation struct {
	/// <summary><para>Latitude between -90 and 90.</para></summary>
	Latitude *float64 `json:"latitude"`
	/// <summary><para>Longitude between -180 and 180.</para></summary>
	Longitude *float64 `json:"longitude"`
	/// <summary><para>Non-negative accuracy value. Defaults to <c>0</c>.</para></summary>
	Accuracy *float64 `json:"accuracy"`
}
type BrowserTypeLaunchPersistentContextOptionsHttpCredentials struct {
	/// <summary><para></para></summary>
	Username *string `json:"username"`
	/// <summary><para></para></summary>
	Password *string `json:"password"`
}
type BrowserTypeLaunchPersistentContextOptionsProxy struct {
	/// <summary>
	/// <para>
	/// Proxy to be used for all requests. HTTP and SOCKS proxies are supported, for example
	/// <c>http://myproxy.com:3128</c> or <c>socks5://myproxy.com:3128</c>. Short form <c>myproxy.com:3128</c>
	/// is considered an HTTP proxy.
	/// </para>
	/// </summary>
	Server *string `json:"server"`
	/// <summary>
	/// <para>
	/// Optional coma-separated domains to bypass proxy, for example <c>".com, chromium.org,
	/// .domain.com"</c>.
	/// </para>
	/// </summary>
	Bypass *string `json:"bypass"`
	/// <summary><para>Optional username to use if HTTP proxy requires authentication.</para></summary>
	Username *string `json:"username"`
	/// <summary><para>Optional password to use if HTTP proxy requires authentication.</para></summary>
	Password *string `json:"password"`
}
type BrowserTypeLaunchPersistentContextOptionsRecordVideo struct {
	/// <summary><para>Path to the directory to put videos into.</para></summary>
	Dir *string `json:"dir"`
	/// <summary>
	/// <para>
	/// Optional dimensions of the recorded videos. If not specified the size will be equal
	/// to <c>viewport</c> scaled down to fit into 800x800. If <c>viewport</c> is not configured
	/// explicitly the video size defaults to 800x450. Actual picture of each page will
	/// be scaled down if necessary to fit the specified size.
	/// </para>
	/// </summary>
	Size *BrowserTypeLaunchPersistentContextOptionsRecordVideoSize `json:"size"`
}
type BrowserTypeLaunchPersistentContextOptionsViewport struct {
	/// <summary><para>page width in pixels.</para></summary>
	Width *int `json:"width"`
	/// <summary><para>page height in pixels.</para></summary>
	Height *int `json:"height"`
}
type BrowserTypeRecordVideoSize struct {
	/// <summary><para>Video frame width.</para></summary>
	Width *int `json:"width"`
	/// <summary><para>Video frame height.</para></summary>
	Height *int `json:"height"`
}
type ElementHandleClickOptionsPosition struct {
	/// <summary><para></para></summary>
	X *float64 `json:"x"`
	/// <summary><para></para></summary>
	Y *float64 `json:"y"`
}
type ElementHandleDblclickOptionsPosition struct {
	/// <summary><para></para></summary>
	X *float64 `json:"x"`
	/// <summary><para></para></summary>
	Y *float64 `json:"y"`
}
type ElementHandleHoverOptionsPosition struct {
	/// <summary><para></para></summary>
	X *float64 `json:"x"`
	/// <summary><para></para></summary>
	Y *float64 `json:"y"`
}
type ElementHandleTapOptionsPosition struct {
	/// <summary><para></para></summary>
	X *float64 `json:"x"`
	/// <summary><para></para></summary>
	Y *float64 `json:"y"`
}
type FrameClickOptionsPosition struct {
	/// <summary><para></para></summary>
	X *float64 `json:"x"`
	/// <summary><para></para></summary>
	Y *float64 `json:"y"`
}
type FrameDblclickOptionsPosition struct {
	/// <summary><para></para></summary>
	X *float64 `json:"x"`
	/// <summary><para></para></summary>
	Y *float64 `json:"y"`
}
type FrameHoverOptionsPosition struct {
	/// <summary><para></para></summary>
	X *float64 `json:"x"`
	/// <summary><para></para></summary>
	Y *float64 `json:"y"`
}
type FrameTapOptionsPosition struct {
	/// <summary><para></para></summary>
	X *float64 `json:"x"`
	/// <summary><para></para></summary>
	Y *float64 `json:"y"`
}
type PageClickOptionsPosition struct {
	/// <summary><para></para></summary>
	X *float64 `json:"x"`
	/// <summary><para></para></summary>
	Y *float64 `json:"y"`
}
type PageDblclickOptionsPosition struct {
	/// <summary><para></para></summary>
	X *float64 `json:"x"`
	/// <summary><para></para></summary>
	Y *float64 `json:"y"`
}
type PageHoverOptionsPosition struct {
	/// <summary><para></para></summary>
	X *float64 `json:"x"`
	/// <summary><para></para></summary>
	Y *float64 `json:"y"`
}
type PageScreenshotOptionsClip struct {
	/// <summary><para>x-coordinate of top-left corner of clip area</para></summary>
	X *float64 `json:"x"`
	/// <summary><para>y-coordinate of top-left corner of clip area</para></summary>
	Y *float64 `json:"y"`
	/// <summary><para>width of clipping area</para></summary>
	Width *float64 `json:"width"`
	/// <summary><para>height of clipping area</para></summary>
	Height *float64 `json:"height"`
}
type PageTapOptionsPosition struct {
	/// <summary><para></para></summary>
	X *float64 `json:"x"`
	/// <summary><para></para></summary>
	Y *float64 `json:"y"`
}
type BrowserNewContextOptionsRecordVideoSize struct {
	/// <summary><para>Video frame width.</para></summary>
	Width *int `json:"width"`
	/// <summary><para>Video frame height.</para></summary>
	Height *int `json:"height"`
}
type BrowserNewPageOptionsRecordVideoSize struct {
	/// <summary><para>Video frame width.</para></summary>
	Width *int `json:"width"`
	/// <summary><para>Video frame height.</para></summary>
	Height *int `json:"height"`
}
type BrowserContextStorageStateResultOriginsLocalStorage struct {
	/// <summary><para></para></summary>
	Name *string `json:"name"`
	/// <summary><para></para></summary>
	Value *string `json:"value"`
}
type BrowserTypeLaunchPersistentContextOptionsRecordVideoSize struct {
	/// <summary><para>Video frame width.</para></summary>
	Width *int `json:"width"`
	/// <summary><para>Video frame height.</para></summary>
	Height *int `json:"height"`
}
