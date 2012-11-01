/*
Package cdype defines main types and constants for Cairo-Dock applets.
*/
package cdtype

// Events you can connect with cairo-dock applet. Use with something like:
//    app.Events.OnClick = func () {app.onClick()}
//    app.Events.OnDropData = func (data string) {app.openWebpage(data)}
//
type Events struct {
	OnClick        func() // state int
	OnMiddleClick  func()
	OnBuildMenu    func()
	OnMenuSelect   func(itemid int32)
	OnScroll       func(scrollUp bool)
	OnDropData     func(data string)
	OnAnswer       func(data interface{})
	OnAnswerDialog func(button int32, data interface{})
	OnShortkey     func(key string)
	OnChangeFocus  func(active bool)

	// TODO: to use
	OnSubClick       func(icon string, state int)
	OnSubMiddleClick func(icon string)
	OnSubBuildMenu   func(icon string)
	OnSubMenuSelect  func(icon string, numEntry int)
	OnSubScroll      func(icon string, scrollUp bool)
	OnSubDropData    func(icon string, data string)

	Reload func(bool)
	End    func()
}

// Defaults settings can be set in one call with something like:
//    app.SetDefaults(dock.Defaults{
//        Label:      "No data",
//        QuickInfo:  "?",
//    })
//
type Defaults struct {
	Icon        string
	Label       string
	QuickInfo   string
	Shortkeys   []string
	MonitorName string
	//~ MonitorClass string
}

type ScreenPosition int32

const (
	ScreenBottom ScreenPosition = iota
	ScreenTop
	ScreenRight
	ScreenLeft
)

type ContainerType int32

const (
	ContainerDock ContainerType = iota
	ContainerDesklet
)

type EmblemPosition int32

const (
	EmblemTopLeft EmblemPosition = iota
	EmblemBottomLeft
	EmblemBottomRight
	EmblemTopRight
	EmblemMiddle
	EmblemBottom
	EmblemTop
	EmblemRight
	EmblemLeft
)

type EmblemModifier int32

const (
	EmblemPersistent EmblemModifier = 0
	EmblemPrint      EmblemModifier = 9
)

type MenuItemType int32

const (
	MenuEntry MenuItemType = 0
	MenuSubMenu
	MenuSeparator
	MenuCheckBox
	MenuRadioButton
)

// MenuItemId
const (
	MainMenuId = 0
)

// DialogKey
const (
	DialogKeyEnter  = -1
	DialogKeyEscape = -2
)

type DockProperties struct {
	Xid         uint64
	X           int32
	Y           int32
	Orientation uint32
	Container   uint32
	Width       int32
	Height      int32
	HasFocus    bool
}
