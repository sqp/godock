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

// Defaults settings can be set in one call with something like:
//    app.SetDefaults(dock.Defaults{
//        Label:      "No data",
//        QuickInfo:  "?",
//    })
//
type Defaults struct {
	Icon      string
	Label     string
	QuickInfo string
	Shortkeys []string

	// MonitorEnabled bool   // Steal icon from the taskbar.
	// MonitorName    string // Name for the application monitoring.
	// Monitor        Command
	//~ MonitorClass string

	PollerInterval int
	Commands       Commands

	Templates []string
	Debug     bool // Enable debug flood.
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

// Small interface to the Dock icon for simple renderers like data pollers.
//
type RenderSimple interface {
	AddDataRenderer(string, int32, string) error
	FileLocation(...string) string
	RenderValues(...float64) error
	SetIcon(string) error
	SetLabel(string) error
	SetQuickInfo(string) error
}

type Commands map[string]*Command

func (commands Commands) FindMonitor() string {
	for _, cmd := range commands {
		if cmd.Monitored {
			if cmd.Class != "" { // Class provided, use it.
				return cmd.Class
			}
			return cmd.Name // Else use program name.
		}
	}
	return "none" // None found, reset it.
}

type Command struct {
	Name      string
	UseOpen   bool
	Monitored bool
	Class     string
}

func NewCommand(monitored bool, name string, class ...string) *Command {
	cmd := &Command{
		Monitored: monitored,
		Name:      name,
	}
	if len(class) > 0 {
		cmd.Class = class[0]
	}
	return cmd
}

func NewCommandStd(action int, name string, class ...string) *Command {
	cmd := NewCommand(action == 3, name, class...)
	cmd.UseOpen = (action == 1)
	return cmd
}
