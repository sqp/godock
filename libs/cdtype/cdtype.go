package cdtype

import (
	"github.com/sqp/godock/libs/poller"

	"os/exec"
	"text/template"
)

// Dock constants.
const (
	AppletsDirName   = "third-party"
	AppletsServerTag = "3.4.0"
)

// Events represents the list of events you can receive as a cairo-dock applet.
//
// They can be set in the optional DefineEvents call of your applet (see
// DefineEventser).
//
// All those events are optional but it's better to find something meaningful to
// assign to them to improve your applet utility.
//
// Use with something like:
//    app.Events.OnClick = func app.myActionLeftClick
//    app.Events.OnDropData = func (data string) {app.openWebpage(data, ...)}
//
// They can also be declared directly as methods of your applet.
//   func (app *Applet) OnClick(btnState uint) { }
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

	// Extends AppIcon, so subicons actions are used the same way as the main icon.
	// Those are the most common actions for an icon (label...)
	AppIcon

	// Actions.

	// ActionAdd adds actions to the list.
	//
	ActionAdd(acts ...*Action)

	// ActionCallback returns a callback to the given action ID.
	//
	ActionCallback(ID int) func()

	// ActionCount returns the number of started actions.
	//
	ActionCount() int

	// ActionID finds the ID matching given action name.
	//
	ActionID(name string) int

	// ActionLaunch starts the desired action by ID.
	//
	ActionLaunch(ID int)

	// ActionSetBool sets the pointer to the boolean value for a checkentry menu field.
	//
	ActionSetBool(ID int, boolPointer *bool)

	// ActionSetMax sets the maximum number of actions that can be started at the same time.
	//
	ActionSetMax(max int)

	// ActionSetIndicators set the pre and post action callbacks.
	//
	ActionSetIndicators(onStart, onStop func())

	// BuildMenu fills the menu with the given actions list.
	//
	// MenuCheckBox: If Call isn't set, a default toggle callback will be used.
	//
	BuildMenu(menu Menuer, actionIds []int)

	// BuildMenuCallback provides a fill menu callback with the given actions list.
	//
	BuildMenuCallback(actionIds []int) func(menu Menuer)

	// Commands.

	// CommandCallback returns a callback to a configured command to bind with event
	// OnClick or OnMiddleClick.
	//
	CommandCallback(name string) func()

	// CommandLaunch executes one of the configured command by its reference.
	//
	CommandLaunch(name string)

	// Poller.

	// AddPoller add a poller to handle in the main loop. Only one can be active ATM.
	// API will almost guaranteed to change for the sub functions.
	//
	AddPoller(call func()) *poller.Poller

	// Poller return the applet poller if any.
	//
	Poller() *poller.Poller

	// Log.

	// Log gives access to the applet logger.
	//
	Log() Logger

	// SetDebug set the state of the debug reporting flood.
	//
	SetDebug(debug bool)

	// Templates.

	// LoadTemplate load the provided list of template files. If error, it will just be be logged, so you must check
	// that the template is valid. Map entry will still be created, just check if it
	// isn't nil. *CDApplet.ExecuteTemplate does it for you.
	//
	// Templates must be in a subdir called templates in applet dir. If you really
	// need a way to change this, ask for a new method.
	//
	LoadTemplate(names ...string)

	// Template gives access to a loaded template by its name.
	//
	Template(file string) *template.Template

	// ExecuteTemplate will run a pre-loaded template with the given data.
	//
	ExecuteTemplate(file, name string, data interface{}) (string, error)

	// Config.

	// LoadConfig will try to create and fill the given config struct with data from
	// the configuration file. Log error and crash if something went wrong.
	// Won't do anything if loadConf is false.
	//
	LoadConfig(loadConf bool, v interface{})

	// ConfFile returns the config file location.
	//
	ConfFile() string // ConfFile returns the config file location.

	// Files

	// FileLocation return the full path to a file in the applet data dir.
	//
	FileLocation(filename ...string) string

	// FileDataDir returns the path to the config root dir (~/.config/cairo-dock).
	//
	FileDataDir(filename ...string) string // RootDataDir.

	// Common.

	// Name returns the applet name as known by the dock. As an external app = dir name.
	//
	Name() string // Name returns the applet name as known by the dock. As an external app = dir name.

	// SetDefaults set basic defaults icon settings in one call. Empty fields will
	// be reset, so this is better used in the Init() call.
	//
	SetDefaults(def Defaults)

	// also include applet management methods as hidden for the doc.
	appManagement
}

// AppIcon defines all methods that can be used on a dock icon (with IconBase too).
//
type AppIcon interface {
	IconBase // Main icon also has the same actions as subicons.

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
	//
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
	//
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

	// Get gets a property of the icon. Current available properties are :
	//
	//   x            int32     x position of the icon's center on the screen (starting from 0 on the left)
	//   y            int32     y position of the icon's center on the screen (starting from 0 at the top of the screen)
	//   width        int32     width of the icon, in pixels (this is the maximum width, when the icon is zoomed)
	//   height       int32     height of the icon, in pixels (this is the maximum height, when the icon is zoomed)
	//   container    uint32   type of container of the applet (DOCK, DESKLET)
	//   orientation  uint32   position of the container on the screen (BOTTOM, TOP, RIGHT, LEFT). A desklet has always an orientation of BOTTOM.
	//   Xid          uint64   ID of the application's window which is controlled by the applet, or 0 if none (this parameter can only be non nul if you used the method ControlAppli beforehand).
	//   has_focus    bool     Whether the application's window which is controlled by the applet is the current active window (it has the focus) or not. E.g.:
	//
	Get(property string) (interface{}, error)

	// GetAll returns all applet icon properties.
	//
	GetAll() *DockProperties

	// AddSubIcon adds subicons by pack of 3 strings : label, icon, ID.
	//
	AddSubIcon(fields ...string) error

	// RemoveSubIcon only need the ID to remove the SubIcon.
	// The option to remove all icons is "any".
	//
	RemoveSubIcon(id string) error

	// RemoveSubIcons removes all subicons from the subdock.
	//
	RemoveSubIcons()

	// SubIcon returns the subicon object you can act on for the given key.
	//
	SubIcon(key string) IconBase
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
	//   app.SetIcon("go-up")
	//   app.SetIcon("/path/to/image")
	//
	SetIcon(icon string) error

	// SetEmblem set an emblem image on the icon.
	// To remove it, you have to use SetEmblem again with an empty string.
	//
	//   app.SetEmblem(app.FileLocation("img", "emblem-work.png"), cdtype.EmblemBottomLeft)
	//
	SetEmblem(iconPath string, position EmblemPosition) error

	// Animate animates the icon for a given number of rounds.
	//
	Animate(animation string, rounds int32) error

	// ShowDialog pops up a simple dialog bubble on the icon.
	// The dialog can be closed by clicking on it.
	//
	ShowDialog(message string, duration int32) error
}

// RenderSimple defines a subset of AppBase for simple renderers like data pollers.
//
type RenderSimple interface {
	AddDataRenderer(string, int32, string) error
	FileLocation(...string) string
	RenderValues(...float64) error
	SetIcon(string) error
	SetLabel(string) error
	SetQuickInfo(string) error
}

//
//----------------------------------------------------------[ APP MANAGEMENT ]--

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

//
//--------------------------------------------------------------------[ LOGS ]--

// Logger defines the interface for reporting information.
//
type Logger interface {
	// SetDebug change the debug state of the logger.
	// Only enable or disable messages send with the Debug command.
	//
	SetDebug(debug bool)

	// GetDebug gets the debug state of the logger.
	//
	GetDebug() bool

	// SetName set the displayed and forwarded name for the logger.
	//
	SetName(name string) Logger

	// SetLogOut connects the optional forwarder to the logger.
	//
	SetLogOut(LogOut)

	// Debug is to be used every time a usefull step is reached in your module
	// activity. It will display the flood to the user only when the debug flag is
	// enabled.
	//
	Debug(msg string, more ...interface{})

	// Info displays normal informations on the standard output, with the first param in green.
	//
	Info(msg string, more ...interface{})

	// Render displays the msg argument in the given color.
	// The colored message is passed with others to classic println.
	//
	Render(color, msg string, more ...interface{})

	// Warn test and log the error as warning type. Return true if an error was found.
	//
	Warn(e error, msg ...string) (fail bool)

	// NewWarn log a new warning.
	//
	NewWarn(e string, msg ...string)

	// Err test and log the error as Error type. Return true if an error was found.
	//
	Err(e error, msg ...interface{}) (fail bool)

	// NewErr log a new error.
	//
	NewErr(e string, msg ...interface{})

	// GetErr test and logs the error, and return it for later use.
	//
	GetErr(e error, msg ...interface{}) error

	// ExecShow run a command with output forwarded to console and wait.
	//
	ExecShow(command string, args ...string) error

	// ExecSync run a command with and grab the output to return it when finished.
	//
	ExecSync(command string, args ...string) (string, error)

	// ExecAsync run a command with output forwarded to console but don't wait for its completion.
	// Errors will be logged.
	//
	ExecAsync(command string, args ...string) error

	// ExecCmd provides a generic command with output forwarded to console.
	//
	ExecCmd(command string, args ...string) *exec.Cmd

	// DEV is like Info, but to be used by the dev for his temporary tests.
	//
	DEV(msg string, more ...interface{})
}

// LogOut defines the interface for log forwarding.
//
type LogOut interface {
	Raw(sender, msg string)
	Info(sender, msg string, more ...interface{})
	Debug(sender, msg string, more ...interface{})
	Err(e string, sender string, msg ...interface{})
}

//
//--------------------------------------------------------------------[ MENU ]--

// Menuer provides a common interface to build applets menu.
//
type Menuer interface {
	// AddSubMenu adds a submenu to the menu.
	//
	AddSubMenu(label, iconPath string) Menuer

	// AddSeparator adds a separator to the menu.
	//
	AddSeparator()

	// AddEntry adds an item to the menu with its callback.
	//
	AddEntry(label, iconPath string, call interface{}, userData ...interface{}) MenuWidgeter

	// AddCheckEntry adds a check entry to the menu.
	//
	AddCheckEntry(label string, active bool, call interface{}, userData ...interface{}) MenuWidgeter

	// AddRadioEntry adds a radio entry to the menu.
	//
	AddRadioEntry(label string, active bool, group int, call interface{}, userData ...interface{}) MenuWidgeter
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
	ID      int
	Name    string
	Call    func()
	Icon    string
	Menu    MenuItemType
	Bool    *bool  // reference to active value for checkitems.
	Group   int    // Radio item group.
	Tooltip string // Entry tooltip.

	// in fact all actions are threaded in the go version, but we could certainly
	// use this as a "add to actions queue" to prevent problems with settings
	// changed while working, or double launch.
	//
	Threaded bool
}

// Shortkey defines mandatory informations to register a shortkey.
//
type Shortkey struct {
	ConfGroup string
	ConfKey   string
	Desc      string
	Shortkey  string
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
	//
	// The type of data depends on the widget provided:
	//     nil         if no widget is provided.
	//     string      for a DialogWidgetText.
	//     float64     for a DialogWidgetScale.
	//     int32       for a DialogWidgetList with Editable=false.
	//     string      for a DialogWidgetList with Editable=true.
	//
	// Callback can be tested and asserted with one of the following functions.
	// The callback will be triggered only if the first button or the enter key
	// is pressed.
	//     DialogCallbackValidNoArg       callback without arguments.
	//     DialogCallbackValidInt         callback with an int.
	//     DialogCallbackValidString      callback with a string.
	//
	Callback func(button int, data interface{})

	// Optional custom widget.
	// Can be of type DialogWidgetText, DialogWidgetScale, DialogWidgetList.
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

// DialogWidgetList defines options for the string list widget of a dialog.
//
type DialogWidgetList struct {
	// Editable represents whether the user can enter a custom choice or not.
	// If true, InitialValue and returned value are string.
	// If false, InitialValue and returned value are int32.
	Editable bool

	// Values represents the combo list values, separated by comma ";".
	Values string

	// InitialValue defines the default value, presented to the user.
	// Type:  string if editable=true, or int32 if editable=false.
	InitialValue interface{}
}

// Dialog answer keys.
const (
	DialogButtonFirst = 0  // Answer when the user press the first button (will often be ok).
	DialogKeyEnter    = -1 // Answer when the user press enter.
	DialogKeyEscape   = -2 // Answer when the user press escape.
)

// DialogCallbackValidNoArg prepares a dialog callback launched only on user confirmation.
// The provided call is triggerred if the user pressed the enter key or the first button.
//
func DialogCallbackValidNoArg(call func()) func(int, interface{}) {
	return func(clickedButton int, _ interface{}) {
		if clickedButton == 0 || clickedButton == -1 {
			call()
		}
	}
}

// DialogCallbackValidInt checks user answer of a dialog to launch the callback
// with the value asserted as an int.
// Will be triggered only by the first button pressed or the enter key.
//
func DialogCallbackValidInt(call func(data int)) func(int, interface{}) {
	return func(clickedButton int, data interface{}) {
		id, ok := data.(int32)
		if ok && (clickedButton == DialogButtonFirst || clickedButton == DialogKeyEnter) {
			call(int(id))
		}
	}
}

// DialogCallbackValidString checks user answer of a dialog to launch the callback
// with the value asserted as a string.
// Will be triggered only by the first button pressed or the enter key.
//
func DialogCallbackValidString(call func(data string)) func(int, interface{}) {
	return func(clickedButton int, data interface{}) {
		str, ok := data.(string)
		if ok && (clickedButton == DialogButtonFirst || clickedButton == DialogKeyEnter) {
			call(str)
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

// ContainerPosition refers to the border of the screen the dock is attached to.
//
// A desklet has always an orientation of BOTTOM.
//
type ContainerPosition int32

// Dock position on screen.
const (
	ContainerPositionBottom ContainerPosition = iota // Dock in the bottom.
	ContainerPositionTop                             // Dock in the top.
	ContainerPositionRight                           // Dock in the right.
	ContainerPositionLeft                            // Dock in the left.
)

// ContainerType is the type of container that manages the applet.
//
type ContainerType int32

// Applet container type.
const (
	ContainerDock    ContainerType = iota // Applet in a dock.
	ContainerDesklet                      // Applet in a desklet.
)

// DeskletVisibility defines the visibility of a desklet.
type DeskletVisibility int

// Desklet visibility settings.
const (
	DeskletVisibilityNormal       DeskletVisibility = iota // Normal, like normal window
	DeskletVisibilityKeepAbove                             // always above
	DeskletVisibilityKeepBelow                             // always below
	DeskletVisibilityWidgetLayer                           // on the Compiz widget layer
	DeskletVisibilityReserveSpace                          // prevent other windows form overlapping it
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
