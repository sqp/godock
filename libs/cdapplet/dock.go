// Package cdapplet is the Cairo-Dock applet base object, using DBus or Gldi backend.
package cdapplet

import (
	"github.com/sqp/godock/libs/cdapplet/action"
	"github.com/sqp/godock/libs/cdglobal"
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/config"
	"github.com/sqp/godock/libs/files"
	"github.com/sqp/godock/libs/log"     // Display info in terminal.
	"github.com/sqp/godock/libs/poller"  // Polling counter.
	"github.com/sqp/godock/libs/ternary" // Ternary operators.

	"errors"
	"path/filepath"
	"strings"
)

//
//----------------------------------------------------------------[ CDAPPLET ]--

// CDApplet is the base Cairo-Dock applet manager that will handle all your
// communications with the dock and provide some methods commonly needed by
// applets.
//
type CDApplet struct {
	appletName string // Applet name as known by the dock. As an external app = dir name.
	confFile   string // Config file location.
	// ParentAppName string // Application launching the applet.
	shareDataDir string // Location of applet data files. As an external applet, it is the same as binary dir.
	rootDataDir  string // Path to the config root dir (~/.config/cairo-dock).

	events    cdtype.Events      // Applet events callbacks (if DefineEvents was used).
	hooker    *Hooker            // Applet events callbacks (for applet self implemented methods).
	action    cdtype.AppAction   // Actions handler. Where an applet can declare its list of actions.
	poller    cdtype.AppPoller   // Poller counter. If you want more than one, use a common denominator.
	command   cdtype.AppCommand  // Programs and locations configured by the user, including application monitor.
	log       cdtype.Logger      // Applet logger.
	shortkeys []*cdtype.Shortkey // Shortkeys and callbacks.
	confPtr   interface{}        // Pointer to applet config.

	cdtype.AppIcon // Dock applet connection, Can be Gldi or Dbus (will be Gldi with build tag dock).
}

// New creates a new applet manager.
//
func New(confPtr interface{}, actions ...*cdtype.Action) cdtype.AppBase {
	app := &CDApplet{
		confPtr: confPtr,
		action:  &action.Actions{},
		hooker:  NewHooker(dockCalls, dockTests),
		log:     log.NewLog(log.Logs),
	}

	app.command = &appCmd{
		commands: make(cdtype.Commands),
		app:      app,
	}
	app.Action().Add(actions...)

	return app
}

// SetBase sets the name, conf and dirs for the applet.
//
func (cda *CDApplet) SetBase(name, conf, rootdir, sharedir string) {
	cda.log.SetName(name)

	cda.appletName = name
	cda.confFile = conf
	cda.rootDataDir = rootdir
	cda.shareDataDir = sharedir
}

// SetBackend sets the applet backend and connects its OnEvent callback to the
// OnEvent method provided here.
//
//    Before           After
//   -------------    -------------
//   |           |    |           |<==\
//   |    -------|    |    -------|   |
//   |    |      |    |    | back |   | OnEvent
//   |    |      |    |    | end  |===/
//   -------------    -------------
//
func (cda *CDApplet) SetBackend(base cdtype.AppBackend) {
	cda.AppIcon = base
	base.SetOnEvent(cda.OnEvent) // connect the backend events to the dispatcher.
}

// SetEvents connects events defined by the applet to the dock.
// It calls the DefineEvents method if the applet provides it, AND also registers
// methods matching those of the API that are defined by the applet.
//
// Returns the first init call func.
//
func (cda *CDApplet) SetEvents(app cdtype.AppInstance) func() error {
	if d, ok := app.(cdtype.DefineEventser); ok { // Old events callback method.
		cda.events = cdtype.Events{
			Reload: func(loadConf bool) {
				cda.log.Debug("Reload module")

				// need pre init.

				def, e := app.LoadConfig(loadConf) // Load config will crash if fail. Expected.
				if e != nil {
					// TODO : NEED STOP APPLET
				}

				app.Init(&def, loadConf)
				app.SetDefaults(def)
				cda.Poller().Restart() // send our restart event. (safe on nil pollers).
			},
		}

		d.DefineEvents(&cda.events)
	}

	cda.RegisterEvents(app) // New events callback method.

	return func() error {
		// Initialise applet: Load config and apply user settings.
		def, e := app.LoadConfig(true) // Load config will crash if fail. Expected.
		if e != nil {
			return e
		}

		app.Init(&def, true)
		app.SetDefaults(def)
		return nil
	}
}

//
//----------------------------------------------------------------[ DEFAULTS ]--

// SetDefaults set basic defaults icon settings in one call.
// Empty fields will be reset, so this is better used in the Init() call.
//
func (cda *CDApplet) SetDefaults(def cdtype.Defaults) {
	// Apply defaults.
	cda.SetIcon(ternary.String(def.Icon != "", def.Icon, cda.FileLocation("icon")))
	cda.SetLabel(ternary.String(def.Label != "", def.Label, cda.Name()))
	cda.SetQuickInfo(def.QuickInfo)

	cda.shortkeys = def.Shortkeys
	for _, sk := range cda.shortkeys {
		switch {
		case sk.ActionID <= 0:

		case sk.ActionID >= cda.Action().Len():
			cda.Log().NewWarn("action not defined", "shortkey: ", sk.ConfGroup, sk.Shortkey) // sk.ActionID

		default:
			sk.Call = cda.Action().CallbackNoArg(sk.ActionID)
		}
	}
	cda.BindShortkey(cda.shortkeys...)

	cda.Command().Clear()
	for key, cmd := range def.Commands {
		cda.Command().Add(key, cmd)
	}
	cda.Window().SetAppliClass(cda.Command().FindMonitor())

	if cda.Poller().Exists() {
		cda.Poller().SetInterval(def.PollerInterval)
	}

	cda.log.SetDebug(def.Debug)
}

//
//------------------------------------------------------------------[ POLLER ]--

// Poller return the applet poller if any.
//
func (cda *CDApplet) Poller() cdtype.AppPoller {
	if cda.poller == nil {
		return poller.NewNil(func(call func()) cdtype.AppPoller {
			cda.poller = poller.New(call)
			return cda.poller
		})
	}
	return cda.poller
}

//
//------------------------------------------------------------------[ ACTION ]--

// Action returns a manager of launchable actions for applets
//
func (cda *CDApplet) Action() cdtype.AppAction {
	return cda.action
}

//
//----------------------------------------------------------------[ COMMANDS ]--

// Command returns a manager of launchable commands for applets
//
func (cda *CDApplet) Command() cdtype.AppCommand {
	return cda.command
}

// appCmd implements cdtype.AppCommand.
type appCmd struct {
	commands cdtype.Commands // Programs and locations configured by the user, including application monitor.
	app      interface {
		Log() cdtype.Logger
		Window() cdtype.IconWindow
	}
}

func (ac *appCmd) Add(key int, cmd *cdtype.Command) {
	ac.commands[key] = cmd
}

// Launch executes one of the configured command by its reference.
//
func (ac *appCmd) Launch(ID int) {
	if cmd, ok := ac.commands[ID]; ok {
		if cmd.Monitored {
			if ac.app.Window().IsOpened() { // Application monitored and opened.
				ac.app.Window().ToggleVisibility()
				// cda.ShowAppli(!hasFocus)
				return
			}
		}

		if cmd.Name == "" {
			ac.app.Log().NewErr("empty command", "CommandLaunch")
			return
		}

		splitted := strings.Split(cmd.Name, " ")

		if cmd.UseOpen {
			ac.app.Log().ExecAsync(cdglobal.CmdOpen, splitted...)
		} else {
			ac.app.Log().ExecAsync(splitted[0], splitted[1:]...)
		}
	}
}

// CallbackNoArg returns a callback to a configured command to bind with event
// OnMiddleClick.
//
func (ac *appCmd) CallbackNoArg(ID int) func() {
	return func() { ac.Launch(ID) }
}

func (ac *appCmd) CallbackInt(ID int) func(int) {
	return func(int) { ac.Launch(ID) }
}

// FindMonitor return the configured window class for the command.
//
func (ac *appCmd) FindMonitor() string {
	for _, cmd := range ac.commands {
		if cmd.Monitored {
			if cmd.Class != "" { // Class provided, use it.
				return cmd.Class
			}
			return cmd.Name // Else use program name.
		}
	}
	return "none" // None found, reset it.
}

func (ac *appCmd) Clear() {
	ac.commands = make(cdtype.Commands)
}

//
//------------------------------------------------------------------[ CONFIG ]--

// LoadConfig will try to create and fill the given config struct with data from
// the configuration file. Log error and crash if something went wrong.
// Won't do anything if loadConf is false.
//
func (cda *CDApplet) LoadConfig(loadConf bool) (cdtype.Defaults, error) {
	def := cdtype.Defaults{}
	if !loadConf {
		return def, nil
	}
	if cda.confPtr == nil {
		return def, errors.New("conf pointer missing")
	}

	files.Access.Lock()
	defer files.Access.Unlock()

	// Try to load config.
	def, liste, e := config.Load(cda.confFile, cda.FileLocation(), cda.confPtr, config.GetBoth)
	if cda.Log().Err(e, "LoadConfig") {
		return def, e
	}

	// Display non fatal errors.
	for _, e := range liste {
		cda.Log().Err(e, "LoadConfig")
	}
	return def, nil
}

// UpdateConfig opens the applet config file for edition.
//
// You must ensure that Save or Cancel is called, and fast to prevent memory
// leaks and deadlocks.
//
func (cda *CDApplet) UpdateConfig() (cdtype.ConfUpdater, error) {
	return files.NewConfUpdater(cda.confFile)
}

//
//-------------------------------------------------------------------[ FILES ]--

// FileDataDir returns the path to the config root dir (~/.config/cairo-dock).
//
func (cda *CDApplet) FileDataDir(filename ...string) string {
	args := append([]string{cda.rootDataDir}, filename...)
	return filepath.Join(args...)
}

// FileLocation return the full path to a file in the applet data dir.
//
func (cda *CDApplet) FileLocation(filename ...string) string {
	args := append([]string{cda.shareDataDir}, filename...)
	return filepath.Join(args...)
}

//
//-------------------------------------------------------------------[ DEBUG ]--

// Log gives access to the applet logger.
//
func (cda *CDApplet) Log() cdtype.Logger {
	return cda.log
}

//
//-----------------------------------------------------------------[ HELPERS ]--

// Name returns the applet name as known by the dock. As an external app = dir name.
//
func (cda *CDApplet) Name() string {
	return cda.appletName
}
