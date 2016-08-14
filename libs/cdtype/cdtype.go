package cdtype

import (
	"os/exec"
	"text/template"
	"time"
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
//   func (app *Applet) OnClick(btnState int) { }
//
// Reload event is optional. Here is the default call if you want to override it.
//
// 	app.Events.Reload = func(loadConf bool) {
// 		app.Log().Debug("Reload module")
// 		app.Init(loadConf)
// 		app.Poller().Restart() // send our restart event. (safe on nil pollers).
// 	}
//
type Events struct {
	// Action when the user clicks on the icon.
	OnClick func(btnState int)

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
	OnSubClick func(icon string, btnState int)

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

	// Extends AppIcon, for all interactions with the dock icon.
	//
	AppIcon

	// --- Common ---
	//
	// Name returns the applet name as known by the dock. As an external app = dir name.
	//
	Name() string // Name returns the applet name as known by the dock. As an external app = dir name.

	// SetDefaults set basic defaults icon settings in one call. Empty fields will
	// be reset, so this is better used in the Init() call.
	//
	SetDefaults(def Defaults)

	// --- Files ---
	//
	// FileLocation returns the full path to a file in the applet data dir.
	//
	FileLocation(filename ...string) string

	// FileDataDir returns the path to the config root dir (~/.config/cairo-dock).
	//
	FileDataDir(filename ...string) string // RootDataDir.

	// --- Config ---
	//
	// LoadConfig will try to create and fill the given config struct with data from
	// the configuration file. Log error and crash if something went wrong.
	// Won't do anything if loadConf is false.
	//
	LoadConfig(loadConf bool, v interface{})

	// UpdateConfig opens the applet config file for edition.

	// You must ensure that Save or Cancel is called, and fast to prevent memory
	// leaks and deadlocks.
	//
	UpdateConfig() (ConfUpdater, error)

	// ConfFile returns the config file location.
	//
	// ConfFile() string // ConfFile returns the config file location.

	// --- Grouped Interfaces ---
	//
	// Action returns a manager of launchable actions for applets
	//
	Action() AppAction

	// Command returns a manager of launchable commands for applets
	//
	Command() AppCommand

	// Poller returns the applet poller if any.
	//
	Poller() AppPoller

	// Template returns a manager of go text templates for applets
	//
	Template() AppTemplate

	// Log gives access to the applet logger.
	//
	Log() Logger

	// also include applet management methods as hidden for the doc.
	appManagement
}

// AppIcon defines all methods that can be used on a dock icon (with IconBase too).
//
type AppIcon interface {
	// Extends IconBase. The main icon also has the same actions as subicons,
	// so subicons are used the same way as the main icon.
	//
	IconBase

	// DemandsAttention is like the Animate method, but will animate the icon
	// endlessly, and the icon will be visible even if the dock is hidden. If the
	// animation is an empty string, or "default", the animation used when an
	// application demands the attention will be used.
	// The first argument is true to start animation, or false to stop it.
	//
	DemandsAttention(start bool, animation string) error

	// PopupDialog opens a dialog box.
	// The dialog can contain a message, an icon, some buttons, and a widget the
	// user can act on.
	//
	// Adding buttons will trigger the provided callback when the user use them.
	//
	// See DialogData for the list of options.
	//
	PopupDialog(DialogData) error

	// DataRenderer manages the graphic data renderer of the icon.
	//
	// You must add a renderer (Gauge, Graph, Progress) before you can Render
	// a list of floats values (from 0 lowest, to 1 highest).
	//
	DataRenderer() IconRenderer

	// Window gives access to actions on the controlled window.
	//
	Window() IconWindow

	// BindShortkey binds any number of keyboard shortcuts to your applet.
	//
	BindShortkey(shortkeys ...Shortkey) error

	// IconProperties gets all applet icon properties at once.
	//
	IconProperties() (IconProperties, error)

	// IconProperty gets applet icon properties one by one.
	//
	IconProperty() IconProperty

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
	Animate(animation string, rounds int) error

	// ShowDialog pops up a simple dialog bubble on the icon.
	// The dialog can be closed by clicking on it.
	//
	ShowDialog(message string, duration int) error
}

// RenderSimple defines a subset of AppBase for simple renderers like data pollers.
//
type RenderSimple interface {
	DataRenderer() IconRenderer
	FileLocation(...string) string
	SetIcon(string) error
	SetLabel(string) error
	SetQuickInfo(string) error
}

//
//-----------------------------------------------------------[ ICON RENDERER ]--

// RendererGraphType defines the type of display for a renderer graph.
type RendererGraphType int

// Types of display for renderer graph.
const (
	RendererGraphLine RendererGraphType = iota
	RendererGraphPlain
	RendererGraphBar
	RendererGraphCircle
	RendererGraphPlainCircle
)

// IconRenderer defines interactions with the graphic data renderer of the icon.
//
type IconRenderer interface {
	Gauge(nbval int, themeName string) error // sets a gauge data renderer.
	Progress(nbval int) error                // sets a progress data renderer.

	Graph(nbval int, typ RendererGraphType) error // sets a graph data renderer.
	GraphLine(nbval int) error                    // sets a graph of type RendererGraphLine.
	GraphPlain(nbval int) error                   // sets a graph of type RendererGraphPlain.
	GraphBar(nbval int) error                     // sets a graph of type RendererGraphBar.
	GraphCircle(nbval int) error                  // sets a graph of type RendererGraphCircle.
	GraphPlainCircle(nbval int) error             // sets a graph of type RendererGraphPlainCircle.

	Remove() error // removes the data renderer.

	// Render renders new values on the icon.
	//
	//   * You must have set a data renderer before.
	//   * The number of values sent must match the number declared before.
	//   * Values are given between 0 and 1.
	//
	Render(values ...float64) error
}

//
//-----------------------------------------------------------[ WINDOW ACTION ]--

// IconWindow defines interactions with the controlled window.
//
type IconWindow interface {
	// SetAppliClass allow your applet to control the window of an external
	// application and to steal its icon from the Taskbar.
	//
	//  *Use the xprop command find the class of the window you want to control.
	//  *Use "none" if you want to reset application control.
	//  *Controling an application enables the OnChangeFocus callback.
	//
	SetAppliClass(applicationClass string) error // Sets the monitored class name.

	IsOpened() bool // Returns true if the monitored application is opened.

	Minimize() error               // Hide the window.
	Show() error                   // Show the window and give it focus.
	SetVisibility(show bool) error // Set the visible state of the window.
	ToggleVisibility() error       // Send Show or Minimize.

	Maximize() error   // Sets the window in full size.
	Restore() error    // Removes the maximized size of the window.
	ToggleSize() error // Send Maximize or Restore.

	Close() error // Close the window (some programs will just hide in the systray).
	Kill() error  // Kill the X window.
}

//
//--------------------------------------------------------------[ PROPERTIES ]--

// IconProperty defines properties of an applet icon.
//
type IconProperty interface {
	// X gets the position of the icon's on the horizontal axis.
	// Starting from 0 on the left.
	//
	X() (int, error)

	// Y gets the position of the icon's on the vertical axis.
	// Starting from 0 at the top of the screen.
	//
	Y() (int, error)

	// Width gets the width of the icon, in pixels.
	// This is the maximum width, when the icon is zoomed.
	//
	Width() (int, error)

	// Height gets the height of the icon, in pixels.
	// This is the maximum height, when the icon is zoomed.
	//
	Height() (int, error)

	//ContainerPosition gets the position of the container on the screen.
	// (bottom, top, right, left). A desklet has always an orientation of bottom.
	//
	ContainerPosition() (ContainerPosition, error)

	// ContainerType gets the type of the applet's container (DOCK, DESKLET).
	//
	ContainerType() (ContainerType, error)

	// Xid gets the ID of the application's window controlled by the applet,
	//  or 0 if none (this parameter can only be non nul if you used the method
	// ControlAppli beforehand).
	//
	Xid() (uint64, error)

	// HasFocus gets whether the application's window which is controlled by the
	// applet is the current active window (it has the focus) or not.
	//
	HasFocus() (bool, error)
}

// IconProperties defines basic informations about a dock icon.
//
type IconProperties interface {
	// X gets the position of the icon's on the horizontal axis.
	//
	X() int

	// Y gets the position of the icon's on the vertical axis.
	//
	Y() int

	// Width gets the width of the icon, in pixels.
	//
	Width() int

	// Height gets the height of the icon, in pixels.
	//
	Height() int

	//ContainerPosition gets the position of the container on the screen.
	//
	ContainerPosition() ContainerPosition

	// ContainerType gets the type of the applet's container (DOCK, DESKLET).
	//
	ContainerType() ContainerType

	// Xid gets the ID of the application's window controlled by the applet,
	//
	Xid() uint64

	// HasFocus gets whether the application's window has the focus or not.
	//
	HasFocus() bool
}

//
//----------------------------------------------------------------[ COMMANDS ]--

// AppTemplate defines interactions with common templates actions.
//
type AppTemplate interface {

	// Load loads the provided list of template files. If error, it will just be be logged, so you must check
	// that the template is valid. Map entry will still be created, just check if it
	// isn't nil. *CDApplet.ExecuteTemplate does it for you.
	//
	// Templates must be in a subdir called templates in applet dir. If you really
	// need a way to change this, ask for a new method.
	//
	Load(names ...string)

	// Get gives access to a loaded template by its name.
	//
	Get(file string) *template.Template

	// Execute runs a pre-loaded template with the given data.
	//
	Execute(file, name string, data interface{}) (string, error)

	// Clear clears the templates list.
	//
	Clear()
}

//
//-----------------------------------------------------------------[ ACTIONS ]--

// AppAction defines a launcher of manageable actions for applets.
//
type AppAction interface {
	// Add adds actions to the list.
	//
	Add(acts ...*Action)

	// CallbackNoArg returns a callback to the given action ID.
	//
	CallbackNoArg(ID int) func()

	// CallbackInt returns a callback to the given action ID with an int input,
	// for left click events.
	//
	CallbackInt(ID int) func(int)

	// Count returns the number of started actions.
	//
	Count() int

	// ID finds the ID matching given action name.
	//
	ID(name string) int

	// Launch starts the desired action by ID.
	//
	Launch(ID int)

	// SetBool sets the pointer to the boolean value for a checkentry menu field.
	//
	SetBool(ID int, boolPointer *bool)

	// SetMax sets the maximum number of actions that can be started at the same time.
	//
	SetMax(max int)

	// SetIndicators set the pre and post action callbacks.
	//
	SetIndicators(onStart, onStop func())

	// BuildMenu fills the menu with the given actions list.
	//
	// MenuCheckBox: If Call isn't set, a default toggle callback will be used.
	//
	BuildMenu(menu Menuer, actionIds []int)

	// CallbackMenu provides a fill menu callback with the given actions list.
	//
	CallbackMenu(actionIds []int) func(menu Menuer)
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

//
//----------------------------------------------------------------[ COMMANDS ]--

// AppCommand defines a launcher of manageable commands for applets.
//
type AppCommand interface {
	// Add adds a command to the list.
	//
	Add(key int, cmd *Command)

	// CallbackNoArg returns a callback to the given command ID.
	// To bind with event OnMiddleClick.
	//
	CallbackNoArg(ID int) func()

	// CallbackInt returns a callback to the given command ID with an int input,
	// To bind with event OnClick.
	//
	CallbackInt(ID int) func(int)

	// Launch executes one of the configured command by its reference.
	//
	Launch(ID int)

	// FindMonitor return the configured window class for the command.
	//
	FindMonitor() string

	// Clear clears the commands list.
	//
	Clear()
}

// Commands handles a list of Command.
//
type Commands map[int]*Command

// Command is the description of a standard command launcher.
//
type Command struct {
	Name      string // Command or location to open.
	UseOpen   bool   // If true, open with the cdglobal.CmdOpen command.
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
//   action: 0=open location, 1=open program, 2=monitor program.
//
func NewCommandStd(action int, name string, class ...string) *Command {
	cmd := NewCommand(action == 3, name, class...)
	cmd.UseOpen = (action == 1)
	return cmd
}

//
//------------------------------------------------------------------[ POLLER ]--

// AppPoller defines an optional applet regular polling actions.
//
type AppPoller interface {
	// Exists returns true if the poller exists (isn't nil).
	//
	Exists() bool

	// Add adds a poller to handle in the main loop. Only one can be active.
	// Multiple calls will just override the action called.
	//
	Add(call func()) AppPoller

	// SetPreCheck callback actions to launch before the polling job.
	//
	SetPreCheck(onStarted func())

	// SetPostCheck callback actions to launch after the polling job.
	//
	SetPostCheck(onFinished func())

	// SetInterval sets the polling interval time, in seconds. You can add a default
	// value as a second argument to be sure you will have a valid value (> 0).
	//
	SetInterval(delay ...int) int

	// Start enables the polling ticker.
	//
	Start()

	// Restart resets the counter and launch Action in a goroutine.
	// Safe to use on nil poller.
	//
	Restart()

	// Stop disables the polling ticker.
	//
	Stop()

	// Wait return a channel that will be triggered after the defined poller interval.
	// You will have to call it on every loop as it not a real ticker.
	// It's just a single use chan.
	//
	Wait() <-chan time.Time

	// Plop increase the counter and launch the action if it reached the interval.
	// The counter is also reset if the action is launched.
	// Safe to use on nil poller.
	//
	Plop() bool
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

	// NewErrf log a new error with arguments formatting.
	//
	NewErrf(title, format string, args ...interface{})

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

	// DEV is like Info, but to be used by the dev for his temporary tests (easier to grep).
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
	Icon            string
	Label           string
	QuickInfo       string
	Shortkeys       []Shortkey
	ShortkeyActions []ShortkeyAction

	PollerInterval int
	Commands       Commands

	Templates []string
	Debug     bool // Enable debug flood.
}

//
//-------------------------------------------------------------[ CONF STRUCT ]--

// ConfGroupIconBoth defines a common config struct for the Icon tab.
//
type ConfGroupIconBoth struct {
	Icon  string `conf:"icon"`
	Name  string `conf:"name"`
	Debug bool
}

// ConfGroupIconName defines a special config struct for the Icon tab.
//
// This if the version without Icon, which shouldn't be the first choice, unless
// you know what you're doing (like a poller locked in gauge/graph).
//
type ConfGroupIconName struct {
	Name  string `conf:"name"`
	Debug bool
}

// ConfUpdater updates a config file.
//
// You must ensure that Save or Cancel is called, and fast to prevent memory
// leaks and deadlocks.
//
type ConfUpdater interface {
	// Set sets a new value for the group/key reference.
	//
	Set(group, key string, value interface{}) error

	// Save saves the edited config to disk, and release locks and memory.
	//
	Save() error

	// Cancel releases locks and memory.
	//
	Cancel()
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

// ContainerType is the type of container that manages the icon.
//
type ContainerType int32

// Icon container type.
const (
	ContainerUnknown ContainerType = iota // just in case.
	ContainerDock                         // Applet in a dock.
	ContainerDesklet                      // Applet in a desklet.
	ContainerDialog
	ContainerFlying
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
