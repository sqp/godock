/*
Package dock is the main Cairo-Dock applet manager, currently tightly linked to
the DBus implementation.

See libs/dbus for direct actions on the applet icons.
*/
package dock

import (
	apiDbus "github.com/guelfey/go.dbus"

	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/config"
	"github.com/sqp/godock/libs/dbus" // Connection to cairo-dock.
	"github.com/sqp/godock/libs/log"  // Display info in terminal.
	"github.com/sqp/godock/libs/poller"

	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"
)

//
//------------------------------------------------------------[ START APPLET ]--

// Methods an applet must implement to use the StartApplet func.
//
type AppletInstance interface {
	// Need to be defined in user applet.
	Init(loadConf bool)
	DefineEvents()

	// Defined by CDApplet
	AddPoller(call func()) *poller.Poller
	Poller() *poller.Poller
	setEventReload(initFunc func(loadConf bool)) // Forward the init callback from interface to the reload event.

	// Defined by CdDbus
	ConnectToBus() (<-chan *apiDbus.Signal, error)
	OnSignal(*apiDbus.Signal)
	GetCloseChan() chan bool
}

// StartApplet will prepare and launch a cairo-dock applet. If you have provided
// events, they will respond when needed, and you have nothing more to worry
// about your applet management. It can handle only one poller for now.
//
// List of the steps, and their effect:
//   * Load applet events definition = DefineEvents().
//   * Connect the applet to cairo-dock with DBus. This also activate events callbacks.
//   * Initialise applet with option load config activated = Init(true).
//   * Start and run the polling loop if needed. This start a instant check, and
//     manage regular and manual timer refresh.
//   * Wait for the dock End signal to close the applet.
//
func StartApplet(app AppletInstance) {
	log.Debug("Applet started")
	defer log.Debug("Applet stopped")

	// Define and connect events to the dock.
	app.DefineEvents()
	app.setEventReload(func(loadConf bool) { app.Init(loadConf) })
	dbusEvent, e := app.ConnectToBus()
	log.Fatal(e, "ConnectToBus") // Mandatory.

	// Initialise applet: Load config and apply user settings.
	app.Init(true)

	// Prepare signals channels.
	close := app.GetCloseChan()

	if poller := app.Poller(); poller != nil {

		restart := poller.ChanRestart() // Restart chan for user events.
		action := true                  // Launch the poller check action directly at start.

		for { // Start main loop and handle events until the End signal is received from the dock.

			if action { // Launch the poller check action.
				go poller.Action()
				action = false
			}

			select { // Wait for events. Until the End signal is received from the dock.

			case s := <-dbusEvent: // Wait for DBus events.
				app.OnSignal(s)

			case <-poller.ChanEndTimer(): // Wait for end of timer. Reloop and check.
				action = true

			case action = <-restart: // Wait for manual restart event. Reloop and check.

			case <-close: // Received End event. That's all folks. We're closing.
				return
			}
		}

	} else { // Just handle DBus events until close event.

		for { // Start main loop and handle events until the End signal is received from the dock.

			select { // Wait for events. Until the End signal is received from the dock.

			case s := <-dbusEvent: // Wait for DBus events.
				app.OnSignal(s)

			case <-close: // Received End event. That's all folks. We're closing.
				return
			}
		}
	}
}

//
//----------------------------------------------------------------[ CDAPPLET ]--

const AppletsDir = "third-party"

// CDApplet is the base Cairo-Dock applet manager that will handle all your
// communications with the dock and provide some methods commonly needed by
// applets.
//
type CDApplet struct {
	AppletName    string // Applet name as known by the dock. As an external app = dir name.
	ConfFile      string // Config file location.
	ParentAppName string // Application launching the applet.
	ShareDataDir  string // Location of applet data files. As an external applet, it is the same as binary dir.
	RootDataDir   string //
	//~ _cMenuIconId string

	Templates map[string]*template.Template // Templates for text formating.
	Actions   Actions                       // Actions handler. Where events callbacks must be declared.
	commands  cdtype.Commands               // Programs and locations configured by the user, including application monitor.
	poller    *poller.Poller                // Poller loop. Need to provide a way to use more than one.

	*dbus.CdDbus // Dbus connector.
}

// Create a new applet manager with arguments received from command line.
//
func Applet() *CDApplet {
	args := os.Args
	name := args[0][2:] // Strip ./ in the beginning.
	cda := &CDApplet{
		AppletName:    name,
		ConfFile:      args[3],
		RootDataDir:   args[4],
		ParentAppName: args[5],
		ShareDataDir:  path.Join(args[4], AppletsDir, name),
		CdDbus:        dbus.New(args[2]),

		Templates: make(map[string]*template.Template),
	}

	log.SetPrefix(name)
	return cda
}

// Set defaults icon settings in one call. Empty fields will be reset, so this
// is better used in the Init() call.
//
func (cda *CDApplet) SetDefaults(def cdtype.Defaults) {
	icon := def.Icon
	if icon == "" {
		icon = cda.FileLocation("icon")
	}
	cda.SetIcon(icon)
	cda.SetQuickInfo(def.QuickInfo)
	cda.SetLabel(def.Label)
	cda.BindShortkey(def.Shortkeys...)

	cda.commands = def.Commands
	cda.ControlAppli(cda.commands.FindMonitor())

	if poller := cda.Poller(); poller != nil {
		poller.SetInterval(def.PollerInterval)
	}

	cda.LoadTemplate(def.Templates...)
	log.SetDebug(def.Debug)
}

// Forward the init callback from applet interface to the reload event.
//
func (cda *CDApplet) setEventReload(appInit func(loadConf bool)) {
	if cda.Events.Reload == nil {
		cda.Events.Reload = func(confChanged bool) {
			log.Debug("Reload module")
			appInit(confChanged)
			if cda.poller != nil {
				cda.poller.Restart() // send our restart event.
			}
		}
	}
}

//
//---------------------------------------------------------------[ TEMPLATES ]--

// Load template files. If error, it will just be be logged, so you must check
// that the template is valid. Map entry will still be created, just check it
// isn't nil. *CDapplet.ExecuteTemplate does it for you.
//
// Templates must be in a subdir called templates in applet dir. If you really
// need a way to change this, ask for a new method.
//
func (cda *CDApplet) LoadTemplate(names ...string) {
	for _, name := range names {
		fileloc := cda.FileLocation("templates", name+".tmpl")
		template, e := template.ParseFiles(fileloc)
		log.Err(e, "Template")
		cda.Templates[name] = template
	}
}

// Execute a pre-loaded template with given data.
//
func (cda *CDApplet) ExecuteTemplate(file, name string, data interface{}) (string, error) {
	if cda.Templates[file] == nil {
		return "", fmt.Errorf("Missing template %s", file)
	}

	buff := bytes.NewBuffer([]byte(""))
	if e := cda.Templates[file].ExecuteTemplate(buff, name, data); log.Err(e, "FormatDialog") {
		return "", e
	}
	return buff.String(), nil
}

//
//------------------------------------------------------------------[ POLLER ]--

// Add a poller to the main loop. Only one can be active in the main loop ATM.
// API will certainly change for the sub functions.
//
func (cda *CDApplet) AddPoller(call func()) *poller.Poller {
	cda.poller = poller.New(call)
	return cda.poller
}

func (cda *CDApplet) Poller() *poller.Poller {
	return cda.poller
}

//
//----------------------------------------------------------------[ COMMANDS ]--

// HaveMonitor gives informations about the state of the monitored application.
// Those are usefull is this option is enabled. A monitored application, if
// opened, is supposed to have its visibility state toggled by the user event.
//
//  haveApp: true if the monitored application is opened. (Xid > 0)
//  HaveFocus: true if the monitored application is the one with the focus.
//
func (cda *CDApplet) HaveMonitor() (haveApp bool, haveFocus bool) {
	Xid, e := cda.Get("Xid")
	log.Err(e, "Xid")

	if id, ok := Xid.(uint64); ok {
		haveApp = id > 0
	}
	HasFocus, _ := cda.Get("has_focus")
	return haveApp, HasFocus.(bool)
}

// Launch one of the configured command by its reference.
//
func (cda *CDApplet) LaunchCommand(name string) {
	if cmd, ok := cda.commands[name]; ok {
		if cmd.Monitored {
			haveMonitor, hasFocus := cda.HaveMonitor()
			if haveMonitor { // Application monitored and opened.
				cda.ShowAppli(!hasFocus)
				return
			}
		}
		if cmd.UseOpen {
			exec.Command("xdg-open", cmd.Name).Start()
		} else {
			cmd := strings.Split(cmd.Name, " ")
			exec.Command(cmd[0], cmd[1:]...).Start()
		}
	}
}

// Returns a callback to a configured command to bind on event OnClick or OnMiddleClick.
//
func (cda *CDApplet) LaunchFunc(name string) func() {
	return func() { cda.LaunchCommand(name) }
}

//
//-----------------------------------------------------------------[ HELPERS ]--

// fieldKey config.GetFieldKey

func (cda *CDApplet) LoadConfig(v interface{}) error {
	return config.Load(cda.ConfFile, v, config.GetBoth)
}

// Get full path to a file in applet data dir.
//
func (cda *CDApplet) FileLocation(filename ...string) string {
	args := append([]string{cda.ShareDataDir}, filename...)
	return path.Join(args...)
}

func PollerInterval(val ...int) int {
	for _, d := range val {
		if d > 0 {
			return d
		}
	}
	return 3600 * 24 // Failed to provide a valid value. Set check interval to one day.
}
