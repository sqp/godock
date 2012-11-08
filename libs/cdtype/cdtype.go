/*
Package cdype defines main types and constants for Cairo-Dock applets.
*/
package cdtype

// Events you can connect with the cairo-dock applet. Use with something like:
//    app.Events.OnClick = func () {app.onClick()}
//    app.Events.OnDropData = func (data string) {app.openWebpage(data)}
//
type Events struct {
	OnClick        func()                               `event:"on_click"` // state int
	OnMiddleClick  func()                               `event:"on_middle_click"`
	OnBuildMenu    func()                               `event:"on_build_menu"`
	OnMenuSelect   func(itemid int32)                   `event:"on_menu_select"`
	OnScroll       func(scrollUp bool)                  `event:"on_scroll"`
	OnDropData     func(data string)                    `event:"on_drop_data"`
	OnAnswer       func(data interface{})               `event:"on_answer"`
	OnAnswerDialog func(button int32, data interface{}) `event:"on_answer_dialog"`
	OnShortkey     func(key string)                     `event:"on_shortkey"`
	OnChangeFocus  func(active bool)                    `event:"on_change_focus"`

	Reload func(bool) `event:"on_reload_module"` // Automatically bind by StartApplet
	End    func()     `event:"on_stop_module"`
}

// TODO: to repair
type SubEvents struct {
	OnSubClick       func(icon string, state int)     `event:"on_click_sub_icon"`
	OnSubMiddleClick func(icon string)                `event:"on_middle_click_sub_icon"`
	OnSubBuildMenu   func(icon string)                `event:"on_build_menu_sub_icon"`
	OnSubMenuSelect  func(icon string, numEntry int)  `event:"on_menu_select_sub_icon"`
	OnSubScroll      func(icon string, scrollUp bool) `event:"on_scroll_sub_icon"`
	OnSubDropData    func(icon string, data string)   `event:"on_drop_data_sub_icon"`
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
	Templates   []string
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
