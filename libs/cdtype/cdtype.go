// Package cdtype defines main types and constants for Cairo-Dock applets.
package cdtype

import "github.com/sqp/godock/libs/poller"

// Dock constants.
const (
	AppletsDirName = "third-party"
)

// Events represents the list of events you can receive with a cairo-dock applet.
// They can be set in the optional DefineEvents call of your applet (see
// DefineEventser).
//   events.OnClick = func() { }
//
// Or they can be declared directly as methods of your applet.
//   func (app *Applet) OnClick() { }
//
// All those events are optional but it's better to find something meaningful to
// assign to them to improve your applet utility.
//
// Use with something like:
//    app.Events.OnClick = func () {app.onClick()}
//    app.Events.OnDropData = func (data string) {app.openWebpage(data)}
//
// Reload event is optional. Here is the default call if you want to override it.
//
// 	app.Events.Reload = func(confChanged bool) {
// 		app.Log.Debug("Reload module")
// 		app.Init(confChanged)
//  	if app.Poller() != nil {
// 			app.Poller().Restart() // send our restart event.
//  	}
// 	}
//
type Events struct {
	// Action when the user clicks on the icon.
	OnClick func()

	// Action when the user use the middle-click on the icon.
	OnMiddleClick func()

	// Action when the user use the right-click on the icon, opening the menu with full options.
	OnBuildMenu func(Menuer)

	// Action when the user use the mouse wheel on the icon.
	OnScroll func(scrollUp bool)

	// Action when the user drop data on the icon.
	OnDropData func(data string)

	// Action when the user triggers a registered shortkey.
	OnShortkey func(key string)

	// Action when the focus of the managed window change.
	OnChangeFocus func(active bool)

	// Action when a reload applet event is triggered from the dock.
	Reload func(bool)

	// Action when the quit applet event is triggered from the dock.
	End func()

	// SubEvents actions work the same as main event with an additional argument
	// for the id of the clicked icon.
	//
	// Action when the user clicks on the subicon.
	OnSubClick func(icon string, state int32)

	// Action when the user use the middle-click on the subicon.
	OnSubMiddleClick func(icon string)

	// Action when the user use the right-click on the subicon, opening the menu.
	OnSubBuildMenu func(icon string, menu Menuer)

	// Action when the user use the mouse wheel on the subicon.
	OnSubScroll func(icon string, scrollUp bool)

	// Action when the user drop data on the subicon.
	OnSubDropData func(icon string, data string)
}

// ListStarter defines a list of applet creation func, indexed by applet name.
//
type ListStarter map[string]AppStarter

// AppStarter represents an applet creation func.
//
type AppStarter func() AppInstance

// AppInstance defines methods an applet must implement to use StartApplet.
//
type AppInstance interface {
	// Need to be defined in user applet.
	Init(loadConf bool)

	// Provided by extending
	AppBase
}

// AppBase groups methods provided to the applet with methods needed on start.
//
type AppBase interface {
	AppIcon

	// Actions.
	ActionAdd(acts ...*Action)
	ActionCallback(ID int) func()
	ActionCount() int
	ActionID(name string) int
	ActionLaunch(ID int)
	ActionSetBool(ID int, boolPointer *bool)
	ActionSetMax(max int)
	ActionSetIndicators(onStart, onStop func())
	BuildMenu(menu Menuer, actionIds []int)
	BuildMenuCallback(actionIds []int) func(menu Menuer)

	// Commands.
	CommandCallback(name string) func()
	CommandLaunch(name string)

	// Poller.
	AddPoller(call func()) *poller.Poller
	Poller() *poller.Poller

	// Log.
	Log() Logger
	SetDebug(debug bool)

	// Templates.
	ExecuteTemplate(file, name string, data interface{}) (string, error)
	LoadTemplate(names ...string)

	// Config.
	LoadConfig(loadConf bool, v interface{})
	ConfFile() string // ConfFile returns the config file location.

	// Files
	FileLocation(filename ...string) string
	FileDataDir(filename ...string) string // RootDataDir returns the path to the config root dir (~/.config/cairo-dock).

	// Common.
	Name() string // Name returns the applet name as known by the dock. As an external app = dir name.
	SetDefaults(def Defaults)

	// also include applet management methods as hidden for the doc.
	appManagement
}

// AppIcon defines all methods that can be used on a dock icon (with IconBase too).
//
type AppIcon interface {
	IconBase // Main icon also has the same actions as subicons.

	// SubIcon returns the subicon object you can act on for the given key.
	//
	SubIcon(key string) IconBase

	// RemoveSubIcons removes all subicons from the subdock.
	//
	RemoveSubIcons()

	// HaveMonitor gives informations about the state of the monitored application.
	// Those are usefull if this option is enabled. A monitored application, if
	// opened, is supposed to have its visibility state toggled by the user event.
	//
	//  haveApp:   true if the monitored application is opened. (Xid > 0)
	//  HaveFocus: true if the monitored application is the one with the focus.
	//
	HaveMonitor() (haveApp bool, haveFocus bool)

	// DemandsAttention is like the Animate method, but will animate the icon
	// endlessly, and the icon will be visible even if the dock is hidden. If the
	// animation is an empty string, or "default", the animation used when an
	// application demands the attention will be used.
	// The first argument is true to start animation, or false to stop it.
	//
	DemandsAttention(start bool, animation string) error

	// PopupDialog open a dialog box .
	// The dialog can contain a message, an icon, some buttons, and a widget the
	// user can act on.
	//
	// Adding buttons will trigger the provided callback when the user use them.
	//
	// See DialogData for the list of options.
	//
	PopupDialog(DialogData) error

	// AddDataRenderer add a graphic data renderer to the icon.
	//
	//  Renderer types: gauge, graph, progressbar.
	//  Themes for renderer Graph: "Line", "Plain", "Bar", "Circle", "Plain Circle"
	//
	AddDataRenderer(typ string, nbval int32, theme string) error

	// RenderValues render new values on the icon.
	//   * You must have added a data renderer before with AddDataRenderer.
	//   * The number of values sent must match the number declared before.
	//   * Values are given between 0 and 1.
	//
	RenderValues(values ...float64) error

	// ActOnAppli send an action on the application controlled by the icon (see ControlAppli).
	//
	//   "minimize"            to hide the window
	//   "show"                to show the window and give it focus
	//   "toggle-visibility"   to show or hide
	//   "maximize"            to maximize the window
	//   "restore"             to restore the window
	//   "toggle-size"         to maximize or restore
	//   "close"               to close the window (Note: some programs will just hide the window and stay in the systray)
	//   "kill"                to kill the X window
	//
	ActOnAppli(action string) error

	// ControlAppli allow your applet to control the window of an external
	// application and can steal its icon from the Taskbar.
	//  *Use the xprop command find the class of the window you want to control.
	//  *Use "none" if you want to reset application control.
	//  *Controling an application enables the OnFocusChange callback.
	//
	ControlAppli(applicationClass string) error

	// ShowAppli set the visible state of the application controlled by the icon.
	//
	ShowAppli(show bool) error

	// BindShortkey binds any number of keyboard shortcuts to your applet.
	//
	BindShortkey(shortkeys ...Shortkey) error

	Get(property string) (interface{}, error)

	// AddSubIcon adds subicons by pack of 3 strings : label, icon, ID.
	//
	AddSubIcon(fields ...string) error

	// RemoveSubIcon only need the ID to remove the SubIcon.
	// The option to remove all icons is "any".
	//
	RemoveSubIcon(id string) error
}

// IconBase defines common actions for icons and subicons.
//
type IconBase interface {

	// SetQuickInfo change the quickinfo text displayed on the subicon.
	//
	SetQuickInfo(info string) error

	// SetLabel change the text label next to the subicon.
	//
	SetLabel(label string) error

	// SetIcon set the image of the icon, overwriting the previous one.
	// A lot of image formats are supported, including SVG.
	// You can refer to the image by either its name if it's an image from a icon theme, or by a path.
	//   app.SetIcon("gimp")
	//   app.SetIcon("gtk-go-up")
	//   app.SetIcon("/path/to/image")
	//
	SetIcon(icon string) error

	// SetEmblem set an emblem image on the icon.
	// To remove it, you have to use SetEmblem again with an empty string.
	//
	//   app.SetEmblem(app.FileLocation("img", "emblem-work.png"), cdtype.EmblemBottomLeft)
	//
	SetEmblem(iconPath string, position EmblemPosition) error

	// Animate animates the icon, with a given animation and for a given number of
	// rounds.
	//
	Animate(animation string, rounds int32) error

	// ShowDialog pops up a simple dialog bubble on the icon.
	// The dialog can be closed by clicking on it.
	//
	ShowDialog(message string, duration int32) error
}

// DefineEventser defines the optional DefineEvents call to group applets events
// definition.
//
type DefineEventser interface {
	DefineEvents(*Events)
}

// appManagement defines methods needed to start the applet and connect the
// different layers of callbacks.
// They are not supposed to be used by applets, this work is done by the launcher.
//
type appManagement interface {
	OnEvent(event string, data ...interface{}) bool
	SetBase(name, conf, rootdir, sharedir string)
	SetBackend(AppBackend)
	SetEvents(AppInstance)
}

// AppBackend extends AppIcon with SetOnEvent used for internal connection.
//
type AppBackend interface {
	AppIcon

	SetOnEvent(func(string, ...interface{}) bool)
}

// Logger defines the interface for reporting information.
//
type Logger interface {
	SetDebug(debug bool)
	GetDebug() bool
	SetName(name string)
	SetLogOut(LogOut)
	Debug(msg string, more ...interface{})
	Info(msg string, more ...interface{})
	Render(color, msg string, more ...interface{})
	Warn(e error, msg ...string) (fail bool)
	NewWarn(e string, msg ...string)
	Err(e error, msg ...interface{}) (fail bool)
	NewErr(e string, msg ...interface{})
	GetErr(e error, msg ...interface{}) error
	ExecShow(command string, args ...string) error
	ExecSync(command string, args ...string) (string, error)
	ExecAsync(command string, args ...string) error
}

// LogOut defines the interface for log forwarding.
//
type LogOut interface {
	Raw(sender, msg string)
	Info(sender, msg string, more ...interface{})
	Debug(sender, msg string, more ...interface{})
	Err(e string, sender string, msg ...interface{})
}

// Menuer provides a common interface to build applets menu.
//
type Menuer interface {
	SubMenu(label, iconPath string) Menuer
	Separator()
	AddEntry(label, iconPath string, call interface{}, userData ...interface{}) MenuWidgeter
	AddCheckEntry(label string, active bool, call interface{}, userData ...interface{}) MenuWidgeter
}

// MenuWidgeter provides a common interface to apply settings on menu entries.
//
type MenuWidgeter interface {
	SetTooltipText(string)
}

//
//----------------------------------------------------------------[ DEFAULTS ]--

// Defaults settings that can be set in one call with something like:
//    app.SetDefaults(dock.Defaults{
//        Label:      "No data",
//        QuickInfo:  "?",
//    })
//
type Defaults struct {
	Icon      string
	Label     string
	QuickInfo string
	Shortkeys []Shortkey

	PollerInterval int
	Commands       Commands

	Templates []string
	Debug     bool // Enable debug flood.
}

// Action is an applet internal actions that can be used for callbacks or menu.
//
type Action struct {
	ID   int
	Name string
	Call func()
	Icon string
	Menu MenuItemType
	Bool *bool // reference to active value for checkitems.

	// in fact all actions are threaded in the go version, but we could certainly
	// use this as a "add to actions queue" to prevent problems with settings
	// changed while working, or double launch.
	//
	Threaded bool
}

// Shortkey defines mandatory informations to register a shortkey.
//
type Shortkey struct {
	Group    string
	Key      string
	Desc     string
	Shortkey string
}

//
//----------------------------------------------------------------[ COMMANDS ]--

// Commands handles a list of Command.
//
type Commands map[string]*Command

// FindMonitor return the configured window class for the command.
//
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

// Command is the description of a standard command launcher.
//
type Command struct {
	Name      string // Command or location to open.
	UseOpen   bool   // If true, open with xdg-open.
	Monitored bool   // If true, the window will be monitored by the dock. (don't work wit UseOpen)
	Class     string // Window class if needed.
}

// NewCommand creates a standard command launcher.
//
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

// NewCommandStd creates a command launcher from configuration options.
//
func NewCommandStd(action int, name string, class ...string) *Command {
	cmd := NewCommand(action == 3, name, class...)
	cmd.UseOpen = (action == 1)
	return cmd
}

//
//-----------------------------------------------------------------[ DIALOGS ]--

// DialogData defines options for a dialog popup.
//
type DialogData struct {
	// Dialog box text.
	Message string

	// Icon displayed next to the message (default=applet icon).
	Icon string

	// Duration of the dialog, in second (0=unlimited).
	TimeLength int

	// True to force the dialog above. Use it with parcimony.
	ForceAbove bool

	// True to use Pango markup to add text decorations.
	UseMarkup bool

	// Images of the buttons, separated by comma ";".
	// "ok" and "cancel" are used as keywords defined by the dock.
	Buttons string

	// Callback when buttons are used.
	//   The type of data depends on the widget provided:
	//     - nil if no widget is provided.
	//     - string for a DialogWidgetText.
	//     - float64 for a DialogWidgetScale.
	Callback func(button int, data interface{})

	// Optional custom widget.
	// Can be of type DialogWidgetText, DialogWidgetScale
	Widget interface{}
}

// DialogWidgetText defines options for the text widget of a dialog.
//
type DialogWidgetText struct {
	MultiLines   bool   // True to have a multi-lines text-entry, ie a text-view.
	Editable     bool   // Whether the user can modify the text or not.
	Visible      bool   // Whether the text will be visible or not (useful to type passwords).
	NbChars      int32  // Maximum number of chars (the current number of chars will be displayed next to the entry) (0=infinite).
	InitialValue string // Text initially contained in the entry.
}

// DialogWidgetScale defines options for the scale widget of a dialog.
//
type DialogWidgetScale struct {
	MinValue     float64 // Lower value.
	MaxValue     float64 // Upper value.
	NbDigit      int32   // Number of digits after the dot.
	InitialValue float64 // Value initially set to the scale.
	MinLabel     string  // Label displayed on the left of the scale.
	MaxLabel     string  // Label displayed on the right of the scale.
}

// Widget list attributes:
//   editable       bool      true if a non-existing choice can be entered by the user (in this case, the content of the widget will be the selected text, and not the number of the selected line) (false by default)
//   values         string    a list of values, separated by comma ";", used to fill the combo list.
//   initial-value  string or int32 depending on the "editable" attribute :
//        case editable=true:   string with the default text for the user entry of the widget (default=empty).
//        case editable=false:  int with the selected line number (default=0).

// Dialog answer keys.
const (
	DialogButtonFirst = 0  // Answer when the user press the first button (will often be ok).
	DialogKeyEnter    = -1 // Answer when the user press enter.
	DialogKeyEscape   = -2 // Answer when the user press escape.
)

// DialogCallbackIsOK prepares a dialog callback launched only on user confirmation.
// The provided call is triggerred if the user pressed the enter key or the first button.
//
func DialogCallbackIsOK(call func()) func(int, interface{}) {
	return func(clickedButton int, _ interface{}) {
		if clickedButton == 0 || clickedButton == -1 {
			call()
		}
	}
}

//
//--------------------------------------------------------------[ PROPERTIES ]--

// DockProperties defines basic informations about a dock icon.
//
type DockProperties struct {
	X      int32 // Distance from the left of the screen.
	Y      int32 // Distance from the bottom of the screen.
	Width  int32 // Width of the icon.
	Height int32 // Height of the icon.

	Orientation uint32 // Dock orientation.
	Container   uint32 // Container type

	HasFocus bool   // True if the monitored window has the cursor focus.
	Xid      uint64 // Xid of the monitored window. Value > 0 if a window is monitored.
}

//
//---------------------------------------------------------------[ CONSTANTS ]--

// ScreenPosition refers to the border of the screen the dock is attached to.
type ScreenPosition int32

// Dock position on screen.
const (
	ScreenBottom ScreenPosition = iota // Dock in the bottom.
	ScreenTop                          // Dock in the top.
	ScreenRight                        // Dock in the right.
	ScreenLeft                         // Dock in the left.
)

// ContainerType is the type of container that manages the applet.
//
type ContainerType int32

// Applet container type.
const (
	ContainerDock    ContainerType = iota // Applet in a dock.
	ContainerDesklet                      // Applet in a desklet.
)

// InfoPosition is the location to render text data for an applet.
//
type InfoPosition int32

// Applet text info position.
const (
	InfoNone    InfoPosition = iota // don't display anything.
	InfoOnIcon                      // display info on the icon (as quick-info).
	InfoOnLabel                     // display on the label of the icon.
)

// EmblemPosition is the location where an emblem is displayed.
//
type EmblemPosition int32

// Applet emblem position.
const (
	EmblemTopLeft     EmblemPosition = iota // Emblem in top left.
	EmblemBottomLeft                        // Emblem in bottom left.
	EmblemBottomRight                       // Emblem in bottom right.
	EmblemTopRight                          // Emblem in top right.
	EmblemMiddle                            // Emblem in the middle.
	EmblemBottom                            // Emblem in the bottom.
	EmblemTop                               // Emblem in the top.
	EmblemRight                             // Emblem in the right.
	EmblemLeft                              // Emblem in the left.
	EmblemCount                             // Number of emblem positions.
)

// type EmblemModifier int32

// const (
// 	EmblemPersistent EmblemModifier = 0
// 	EmblemPrint      EmblemModifier = 9
// )

// MenuItemType is the type of menu entry to create.
//
type MenuItemType int32

// Applet menu entry type.
const (
	MenuEntry       MenuItemType = iota // Simple menu text entry.
	MenuSubMenu                         // Not working for Dbus.
	MenuSeparator                       // Menu separator.
	MenuCheckBox                        // Menu checkbox.
	MenuRadioButton                     // Not working.
)

// MenuItemId
// const (
// 	MainMenuId = 0
// )

// DialogKey

// PollerInterval returns a valid poller check interval.
//
func PollerInterval(val ...int) int {
	for _, d := range val {
		if d > 0 {
			return d
		}
	}
	return 3600 * 24 // Failed to provide a valid value. Set check interval to one day.
}
