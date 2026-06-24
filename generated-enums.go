package playwright

func getMixedState(in string) *MixedState {
	v := MixedState(in)
	return &v
}

type MixedState string

var (
	MixedStateOn    *MixedState = getMixedState("On")
	MixedStateOff               = getMixedState("Off")
	MixedStateMixed             = getMixedState("Mixed")
)

func getElementState(in string) *ElementState {
	v := ElementState(in)
	return &v
}

type ElementState string

var (
	ElementStateVisible  *ElementState = getElementState("visible")
	ElementStateHidden                 = getElementState("hidden")
	ElementStateStable                 = getElementState("stable")
	ElementStateEnabled                = getElementState("enabled")
	ElementStateDisabled               = getElementState("disabled")
	ElementStateEditable               = getElementState("editable")
)

func getAriaRole(in string) *AriaRole {
	v := AriaRole(in)
	return &v
}

type AriaRole string

var (
	AriaRoleAlert            *AriaRole = getAriaRole("alert")
	AriaRoleAlertdialog                = getAriaRole("alertdialog")
	AriaRoleApplication                = getAriaRole("application")
	AriaRoleArticle                    = getAriaRole("article")
	AriaRoleBanner                     = getAriaRole("banner")
	AriaRoleBlockquote                 = getAriaRole("blockquote")
	AriaRoleButton                     = getAriaRole("button")
	AriaRoleCaption                    = getAriaRole("caption")
	AriaRoleCell                       = getAriaRole("cell")
	AriaRoleCheckbox                   = getAriaRole("checkbox")
	AriaRoleCode                       = getAriaRole("code")
	AriaRoleColumnheader               = getAriaRole("columnheader")
	AriaRoleCombobox                   = getAriaRole("combobox")
	AriaRoleComplementary              = getAriaRole("complementary")
	AriaRoleContentinfo                = getAriaRole("contentinfo")
	AriaRoleDefinition                 = getAriaRole("definition")
	AriaRoleDeletion                   = getAriaRole("deletion")
	AriaRoleDialog                     = getAriaRole("dialog")
	AriaRoleDirectory                  = getAriaRole("directory")
	AriaRoleDocument                   = getAriaRole("document")
	AriaRoleEmphasis                   = getAriaRole("emphasis")
	AriaRoleFeed                       = getAriaRole("feed")
	AriaRoleFigure                     = getAriaRole("figure")
	AriaRoleForm                       = getAriaRole("form")
	AriaRoleGeneric                    = getAriaRole("generic")
	AriaRoleGrid                       = getAriaRole("grid")
	AriaRoleGridcell                   = getAriaRole("gridcell")
	AriaRoleGroup                      = getAriaRole("group")
	AriaRoleHeading                    = getAriaRole("heading")
	AriaRoleImg                        = getAriaRole("img")
	AriaRoleInsertion                  = getAriaRole("insertion")
	AriaRoleLink                       = getAriaRole("link")
	AriaRoleList                       = getAriaRole("list")
	AriaRoleListbox                    = getAriaRole("listbox")
	AriaRoleListitem                   = getAriaRole("listitem")
	AriaRoleLog                        = getAriaRole("log")
	AriaRoleMain                       = getAriaRole("main")
	AriaRoleMarquee                    = getAriaRole("marquee")
	AriaRoleMath                       = getAriaRole("math")
	AriaRoleMeter                      = getAriaRole("meter")
	AriaRoleMenu                       = getAriaRole("menu")
	AriaRoleMenubar                    = getAriaRole("menubar")
	AriaRoleMenuitem                   = getAriaRole("menuitem")
	AriaRoleMenuitemcheckbox           = getAriaRole("menuitemcheckbox")
	AriaRoleMenuitemradio              = getAriaRole("menuitemradio")
	AriaRoleNavigation                 = getAriaRole("navigation")
	AriaRoleNone                       = getAriaRole("none")
	AriaRoleNote                       = getAriaRole("note")
	AriaRoleOption                     = getAriaRole("option")
	AriaRoleParagraph                  = getAriaRole("paragraph")
	AriaRolePresentation               = getAriaRole("presentation")
	AriaRoleProgressbar                = getAriaRole("progressbar")
	AriaRoleRadio                      = getAriaRole("radio")
	AriaRoleRadiogroup                 = getAriaRole("radiogroup")
	AriaRoleRegion                     = getAriaRole("region")
	AriaRoleRow                        = getAriaRole("row")
	AriaRoleRowgroup                   = getAriaRole("rowgroup")
	AriaRoleRowheader                  = getAriaRole("rowheader")
	AriaRoleScrollbar                  = getAriaRole("scrollbar")
	AriaRoleSearch                     = getAriaRole("search")
	AriaRoleSearchbox                  = getAriaRole("searchbox")
	AriaRoleSeparator                  = getAriaRole("separator")
	AriaRoleSlider                     = getAriaRole("slider")
	AriaRoleSpinbutton                 = getAriaRole("spinbutton")
	AriaRoleStatus                     = getAriaRole("status")
	AriaRoleStrong                     = getAriaRole("strong")
	AriaRoleSubscript                  = getAriaRole("subscript")
	AriaRoleSuperscript                = getAriaRole("superscript")
	AriaRoleSwitch                     = getAriaRole("switch")
	AriaRoleTab                        = getAriaRole("tab")
	AriaRoleTable                      = getAriaRole("table")
	AriaRoleTablist                    = getAriaRole("tablist")
	AriaRoleTabpanel                   = getAriaRole("tabpanel")
	AriaRoleTerm                       = getAriaRole("term")
	AriaRoleTextbox                    = getAriaRole("textbox")
	AriaRoleTime                       = getAriaRole("time")
	AriaRoleTimer                      = getAriaRole("timer")
	AriaRoleToolbar                    = getAriaRole("toolbar")
	AriaRoleTooltip                    = getAriaRole("tooltip")
	AriaRoleTree                       = getAriaRole("tree")
	AriaRoleTreegrid                   = getAriaRole("treegrid")
	AriaRoleTreeitem                   = getAriaRole("treeitem")
)

func getColorScheme(in string) *ColorScheme {
	v := ColorScheme(in)
	return &v
}

type ColorScheme string

var (
	ColorSchemeLight        *ColorScheme = getColorScheme("light")
	ColorSchemeDark                      = getColorScheme("dark")
	ColorSchemeNoPreference              = getColorScheme("no-preference")
	ColorSchemeNoOverride                = getColorScheme("no-override")
)

func getContrast(in string) *Contrast {
	v := Contrast(in)
	return &v
}

type Contrast string

var (
	ContrastNoPreference *Contrast = getContrast("no-preference")
	ContrastMore                   = getContrast("more")
	ContrastNoOverride             = getContrast("no-override")
)

func getForcedColors(in string) *ForcedColors {
	v := ForcedColors(in)
	return &v
}

type ForcedColors string

var (
	ForcedColorsActive     *ForcedColors = getForcedColors("active")
	ForcedColorsNone                     = getForcedColors("none")
	ForcedColorsNoOverride               = getForcedColors("no-override")
)

func getHarContentPolicy(in string) *HarContentPolicy {
	v := HarContentPolicy(in)
	return &v
}

type HarContentPolicy string

var (
	HarContentPolicyOmit   *HarContentPolicy = getHarContentPolicy("omit")
	HarContentPolicyEmbed                    = getHarContentPolicy("embed")
	HarContentPolicyAttach                   = getHarContentPolicy("attach")
)

func getHarMode(in string) *HarMode {
	v := HarMode(in)
	return &v
}

type HarMode string

var (
	HarModeFull    *HarMode = getHarMode("full")
	HarModeMinimal          = getHarMode("minimal")
)

func getReducedMotion(in string) *ReducedMotion {
	v := ReducedMotion(in)
	return &v
}

type ReducedMotion string

var (
	ReducedMotionReduce       *ReducedMotion = getReducedMotion("reduce")
	ReducedMotionNoPreference                = getReducedMotion("no-preference")
	ReducedMotionNoOverride                  = getReducedMotion("no-override")
)

func getServiceWorkerPolicy(in string) *ServiceWorkerPolicy {
	v := ServiceWorkerPolicy(in)
	return &v
}

type ServiceWorkerPolicy string

var (
	ServiceWorkerPolicyAllow *ServiceWorkerPolicy = getServiceWorkerPolicy("allow")
	ServiceWorkerPolicyBlock                      = getServiceWorkerPolicy("block")
)

func getSameSiteAttribute(in string) *SameSiteAttribute {
	v := SameSiteAttribute(in)
	return &v
}

type SameSiteAttribute string

var (
	SameSiteAttributeStrict *SameSiteAttribute = getSameSiteAttribute("Strict")
	SameSiteAttributeLax                       = getSameSiteAttribute("Lax")
	SameSiteAttributeNone                      = getSameSiteAttribute("None")
)

func getHarNotFound(in string) *HarNotFound {
	v := HarNotFound(in)
	return &v
}

type HarNotFound string

var (
	HarNotFoundAbort    *HarNotFound = getHarNotFound("abort")
	HarNotFoundFallback              = getHarNotFound("fallback")
)

func getRouteFromHarUpdateContentPolicy(in string) *RouteFromHarUpdateContentPolicy {
	v := RouteFromHarUpdateContentPolicy(in)
	return &v
}

type RouteFromHarUpdateContentPolicy string

var (
	RouteFromHarUpdateContentPolicyEmbed  *RouteFromHarUpdateContentPolicy = getRouteFromHarUpdateContentPolicy("embed")
	RouteFromHarUpdateContentPolicyAttach                                  = getRouteFromHarUpdateContentPolicy("attach")
)

func getUnrouteBehavior(in string) *UnrouteBehavior {
	v := UnrouteBehavior(in)
	return &v
}

type UnrouteBehavior string

var (
	UnrouteBehaviorWait         *UnrouteBehavior = getUnrouteBehavior("wait")
	UnrouteBehaviorIgnoreErrors                  = getUnrouteBehavior("ignoreErrors")
	UnrouteBehaviorDefault                       = getUnrouteBehavior("default")
)

func getMouseButton(in string) *MouseButton {
	v := MouseButton(in)
	return &v
}

type MouseButton string

var (
	MouseButtonLeft   *MouseButton = getMouseButton("left")
	MouseButtonRight               = getMouseButton("right")
	MouseButtonMiddle              = getMouseButton("middle")
)

func getKeyboardModifier(in string) *KeyboardModifier {
	v := KeyboardModifier(in)
	return &v
}

type KeyboardModifier string

var (
	KeyboardModifierAlt           *KeyboardModifier = getKeyboardModifier("Alt")
	KeyboardModifierControl                         = getKeyboardModifier("Control")
	KeyboardModifierControlOrMeta                   = getKeyboardModifier("ControlOrMeta")
	KeyboardModifierMeta                            = getKeyboardModifier("Meta")
	KeyboardModifierShift                           = getKeyboardModifier("Shift")
)

func getScreenshotAnimations(in string) *ScreenshotAnimations {
	v := ScreenshotAnimations(in)
	return &v
}

type ScreenshotAnimations string

var (
	ScreenshotAnimationsDisabled *ScreenshotAnimations = getScreenshotAnimations("disabled")
	ScreenshotAnimationsAllow                          = getScreenshotAnimations("allow")
)

func getScreenshotCaret(in string) *ScreenshotCaret {
	v := ScreenshotCaret(in)
	return &v
}

type ScreenshotCaret string

var (
	ScreenshotCaretHide    *ScreenshotCaret = getScreenshotCaret("hide")
	ScreenshotCaretInitial                  = getScreenshotCaret("initial")
)

func getScreenshotScale(in string) *ScreenshotScale {
	v := ScreenshotScale(in)
	return &v
}

type ScreenshotScale string

var (
	ScreenshotScaleCss    *ScreenshotScale = getScreenshotScale("css")
	ScreenshotScaleDevice                  = getScreenshotScale("device")
)

func getScreenshotType(in string) *ScreenshotType {
	v := ScreenshotType(in)
	return &v
}

type ScreenshotType string

var (
	ScreenshotTypePng  *ScreenshotType = getScreenshotType("png")
	ScreenshotTypeJpeg                 = getScreenshotType("jpeg")
)

func getWaitForSelectorState(in string) *WaitForSelectorState {
	v := WaitForSelectorState(in)
	return &v
}

type WaitForSelectorState string

var (
	WaitForSelectorStateAttached *WaitForSelectorState = getWaitForSelectorState("attached")
	WaitForSelectorStateDetached                       = getWaitForSelectorState("detached")
	WaitForSelectorStateVisible                        = getWaitForSelectorState("visible")
	WaitForSelectorStateHidden                         = getWaitForSelectorState("hidden")
)

func getWaitUntilState(in string) *WaitUntilState {
	v := WaitUntilState(in)
	return &v
}

type WaitUntilState string

var (
	WaitUntilStateLoad             *WaitUntilState = getWaitUntilState("load")
	WaitUntilStateDomcontentloaded                 = getWaitUntilState("domcontentloaded")
	WaitUntilStateNetworkidle                      = getWaitUntilState("networkidle")
	WaitUntilStateCommit                           = getWaitUntilState("commit")
)

func getLoadState(in string) *LoadState {
	v := LoadState(in)
	return &v
}

type LoadState string

var (
	LoadStateLoad             *LoadState = getLoadState("load")
	LoadStateDomcontentloaded            = getLoadState("domcontentloaded")
	LoadStateNetworkidle                 = getLoadState("networkidle")
)

func getAriaSnapshotMode(in string) *AriaSnapshotMode {
	v := AriaSnapshotMode(in)
	return &v
}

type AriaSnapshotMode string

var (
	AriaSnapshotModeAi      *AriaSnapshotMode = getAriaSnapshotMode("ai")
	AriaSnapshotModeDefault                   = getAriaSnapshotMode("default")
)

func getPseudoElement(in string) *PseudoElement {
	v := PseudoElement(in)
	return &v
}

type PseudoElement string

var (
	PseudoElementBefore *PseudoElement = getPseudoElement("before")
	PseudoElementAfter                 = getPseudoElement("after")
)

func getMedia(in string) *Media {
	v := Media(in)
	return &v
}

type Media string

var (
	MediaScreen     *Media = getMedia("screen")
	MediaPrint             = getMedia("print")
	MediaNoOverride        = getMedia("no-override")
)

func getConsoleMessagesFilter(in string) *ConsoleMessagesFilter {
	v := ConsoleMessagesFilter(in)
	return &v
}

type ConsoleMessagesFilter string

var (
	ConsoleMessagesFilterAll             *ConsoleMessagesFilter = getConsoleMessagesFilter("all")
	ConsoleMessagesFilterSinceNavigation                        = getConsoleMessagesFilter("since-navigation")
)

func getAnnotatePosition(in string) *AnnotatePosition {
	v := AnnotatePosition(in)
	return &v
}

type AnnotatePosition string

var (
	AnnotatePositionTopLeft     *AnnotatePosition = getAnnotatePosition("top-left")
	AnnotatePositionTop                           = getAnnotatePosition("top")
	AnnotatePositionTopRight                      = getAnnotatePosition("top-right")
	AnnotatePositionBottomLeft                    = getAnnotatePosition("bottom-left")
	AnnotatePositionBottom                        = getAnnotatePosition("bottom")
	AnnotatePositionBottomRight                   = getAnnotatePosition("bottom-right")
)

func getHttpCredentialsSend(in string) *HttpCredentialsSend {
	v := HttpCredentialsSend(in)
	return &v
}

type HttpCredentialsSend string

var (
	HttpCredentialsSendUnauthorized *HttpCredentialsSend = getHttpCredentialsSend("unauthorized")
	HttpCredentialsSendAlways                            = getHttpCredentialsSend("always")
)
