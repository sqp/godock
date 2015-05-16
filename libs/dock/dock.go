// Package dock is the Cairo-Dock applet manager, using DBus or Gldi backend.
package dock

import (
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/config"
	"github.com/sqp/godock/libs/log"     // Display info in terminal.
	"github.com/sqp/godock/libs/poller"  // Polling counter.
	"github.com/sqp/godock/libs/ternary" // Ternary operators.

	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
)

// RenderSimple is a small interface to the Dock icon for simple renderers like data pollers.
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
//----------------------------------------------------------------[ CDAPPLET ]--

// CDApplet is the base Cairo-Dock applet manager that will handle all your
// communications with the dock and provide some methods commonly needed by
// applets.
//
type CDApplet struct {
	Actions // Actions handler. Where an applet can declare its list of actions.

	appletName string // Applet name as known by the dock. As an external app = dir name.
	confFile   string // Config file location.
	// ParentAppName string // Application launching the applet.
	shareDataDir string // Location of applet data files. As an external applet, it is the same as binary dir.
	rootDataDir  string // Path to the config root dir (~/.config/cairo-dock).

	events    cdtype.Events                 // Applet events callbacks (if DefineEvents was used).
	hooker    *Hooker                       // Applet events callbacks (for applet self implemented methods).
	poller    *poller.Poller                // Poller counter. If you want more than one, use a common denominator.
	commands  cdtype.Commands               // Programs and locations configured by the user, including application monitor.
	templates map[string]*template.Template // Templates for text formating.
	log       cdtype.Logger                 // Applet logger.

	cdtype.AppIcon // Dock applet connection, Can be Gldi or Dbus (will be Gldi with build dock tag).
}

// NewCDApplet creates a new applet manager.
//
func NewCDApplet() cdtype.AppBase {
	return &CDApplet{
		hooker:    NewHooker(dockCalls, dockTests),
		templates: make(map[string]*template.Template),
		log:       log.NewLog(log.Logs)}
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
func (cda *CDApplet) SetEvents(app cdtype.AppInstance) {

	if d, ok := app.(cdtype.DefineEventser); ok { // Old events callback method.
		cda.events = cdtype.Events{
			Reload: func(loadConf bool) {
				cda.log.Debug("Reload module")
				app.Init(loadConf)
				cda.poller.Restart() // send our restart event. (safe on nil pollers).
			},
		}

		d.DefineEvents(&cda.events)
	}

	cda.RegisterEvents(app) // New events callback method.
}

//
//----------------------------------------------------------------[ DEFAULTS ]--

// SetDefaults set basic defaults icon settings in one call. Empty fields will
// be reset, so this is better used in the Init() call.
//
func (cda *CDApplet) SetDefaults(def cdtype.Defaults) {
	cda.SetIcon(ternary.String(def.Icon != "", def.Icon, cda.FileLocation("icon")))
	cda.SetLabel(ternary.String(def.Label != "", def.Label, cda.Name()))
	cda.SetQuickInfo(def.QuickInfo)
	cda.BindShortkey(def.Shortkeys...)
	cda.commands = def.Commands
	cda.ControlAppli(cda.commands.FindMonitor())

	if poller := cda.Poller(); poller != nil {
		poller.SetInterval(def.PollerInterval)
	}

	cda.LoadTemplate(def.Templates...)

	cda.log.SetDebug(def.Debug)
}

//
//---------------------------------------------------------------[ TEMPLATES ]--

// LoadTemplate load the provided list of template files. If error, it will just be be logged, so you must check
// that the template is valid. Map entry will still be created, just check if it
// isn't nil. *CDApplet.ExecuteTemplate does it for you.
//
// Templates must be in a subdir called templates in applet dir. If you really
// need a way to change this, ask for a new method.
//
func (cda *CDApplet) LoadTemplate(names ...string) {
	for _, name := range names {
		fileloc := cda.FileLocation("templates", name+".tmpl")
		template, e := template.ParseFiles(fileloc)
		cda.log.Err(e, "Template")
		cda.templates[name] = template
	}
}

// ExecuteTemplate will run a pre-loaded template with the given data.
//
func (cda *CDApplet) ExecuteTemplate(file, name string, data interface{}) (string, error) {
	if cda.templates[file] == nil {
		return "", fmt.Errorf("missing template %s", file)
	}

	buff := bytes.NewBuffer([]byte(""))
	if e := cda.templates[file].ExecuteTemplate(buff, name, data); cda.log.Err(e, "FormatDialog") {
		return "", e
	}
	return buff.String(), nil
}

//
//------------------------------------------------------------------[ POLLER ]--

// AddPoller add a poller to handle in the main loop. Only one can be active ATM.
// API will almost guaranteed to change for the sub functions.
//
func (cda *CDApplet) AddPoller(call func()) *poller.Poller {
	cda.poller = poller.New(call)
	return cda.poller
}

// Poller return the applet poller if any.
//
func (cda *CDApplet) Poller() *poller.Poller {
	return cda.poller
}

//
//----------------------------------------------------------------[ COMMANDS ]--

// CommandLaunch executes one of the configured command by its reference.
//
func (cda *CDApplet) CommandLaunch(name string) {
	if cmd, ok := cda.commands[name]; ok {
		if cmd.Monitored {
			haveMonitor, hasFocus := cda.HaveMonitor()
			if haveMonitor { // Application monitored and opened.
				cda.ShowAppli(!hasFocus)
				return
			}
		}
		splitted := strings.Split(cmd.Name, " ")
		if cmd.UseOpen {
			cda.log.ExecAsync("xdg-open", splitted...)
		} else {
			cda.log.ExecAsync(splitted[0], splitted[1:]...)
		}
	}
}

// CommandCallback returns a callback to a configured command to bind with event
// OnClick or OnMiddleClick.
//
func (cda *CDApplet) CommandCallback(name string) func() {
	return func() { cda.CommandLaunch(name) }
}

//
//------------------------------------------------------------------[ CONFIG ]--

// ConfFile returns the config file location.
//
func (cda *CDApplet) ConfFile() string {
	return cda.confFile
}

// LoadConfig will try to create and fill the given config struct with data from
// the configuration file. Log error and crash if something went wrong.
// Won't do anything if loadConf is false.
//
func (cda *CDApplet) LoadConfig(loadConf bool, v interface{}) {
	if loadConf { // Try to load config. Exit if not found.
		log.Fatal(config.Load(cda.confFile, v, config.GetBoth), "config")
	}
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

// SetDebug set the state of the debug reporting flood.
//
func (cda *CDApplet) SetDebug(debug bool) {
	cda.log.SetDebug(debug)
}

//
//-----------------------------------------------------------------[ HELPERS ]--

// Name returns the applet name as known by the dock. As an external app = dir name.
//
func (cda *CDApplet) Name() string {
	return cda.appletName
}
