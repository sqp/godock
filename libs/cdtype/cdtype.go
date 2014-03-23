/*
Package cdype defines main types and constants for Cairo-Dock applets.
*/
package cdtype

// Events you can connect with the cairo-dock applet. They are better set in the
// mandatory DefineEvents call of your applet.
//
// Use with something like:
//    app.Events.OnClick = func () {app.onClick()}
//    app.Events.OnDropData = func (data string) {app.openWebpage(data)}
//
// Reload event is optional. Here is the default call if you want to override it.
//
// 	app.Events.Reload = func(confChanged bool) {
// 		log.Debug("Reload module")
// 		app.Init(confChanged)
// 		app.poller.Restart() // send our restart event.
// 	}
//
type Events struct {
	OnClick        func()
	OnMiddleClick  func()
	OnBuildMenu    func()
	OnMenuSelect   func(itemid int32)
	OnScroll       func(scrollUp bool)
	OnDropData     func(data string)
	OnAnswer       func(data interface{})
	OnAnswerDialog func(button int32, data interface{})
	OnShortkey     func(key string)
	OnChangeFocus  func(active bool)

	Reload func(bool)
	End    func()
}

// SubEvents work the same as main event with an additional argument for the id
// of the clicked icon.
//
type SubEvents struct {
	OnSubClick       func(state int32, icon string)
	OnSubMiddleClick func(icon string)
	OnSubBuildMenu   func(icon string)
	OnSubMenuSelect  func(numEntry int32, icon string)
	OnSubScroll      func(scrollUp bool, icon string)
	OnSubDropData    func(data string, icon string)
}

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

type InfoPosition int32

const (
	InfoNone    = iota // don't display anything.
	InfoOnIcon         // display info on the icon (as quick-info).
	InfoOnLabel        // display on the label of the icon.
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
	MenuEntry MenuItemType = iota
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
