/*
Package dock is the main Cairo-Dock applet manager, currently tightly linked to 
the DBus implementation.

See libs/dbus for direct actions on the applet icons.
*/
package dock

import (
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/dbus" // Connection to cairo-dock.
	"github.com/sqp/godock/libs/log"  // Display info in terminal.
	"github.com/sqp/godock/libs/poller"
	"os"
	"path"
)

//------------------------------------------------------------[ START APPLET ]--

// Methods an applet must implement to use the StartApplet func.
//
type AppletInstance interface {
	DefineEvents()
	Init(loadConf bool)
	Reload(confChanged bool)
}

// StartApplet will prepare and launch a cairo-dock applet. If you have provided
// events, they will respond when needed, and you have nothing more to worry
// about your applet management. One optional poller can be provided atm.
//
// List of the steps, and their effect:
//   * Load applet events definition = DefineEvents().
//   * Connect the applet to cairo-dock with DBus. This also activate events callbacks.
//   * Initialise applet with option load config activated = Init(true).
//   * Start and run the polling loop if needed. This start a instant check, and 
//     manage regular and manual timer refresh.
//   * Wait for the dock End signal to close the applet.
//
func StartApplet(cda *CDApplet, app AppletInstance, poller ...*poller.Poller) {
	log.Debug("Applet started")
	defer log.Debug("Applet stopped")

	// Define and connect events to the dock
	cda.Events.Reload = func(loadConf bool) {
		app.Reload(loadConf)
	}

	app.DefineEvents()
	cda.ConnectToBus()

	// Load config and apply user settings.
	app.Init(true)

	// Prepare signals channels.
	close := cda.GetCloseChan()

	if len(poller) == 0 {
		<-close // Just wait for End event signal.

	} else { // With only one poller currently managed.
		var restart chan bool
		restart = poller[0].GetRestart()

		for { // Start main loop and wait for events until the End signal is received from the dock.
			//~ var tick *time.Ticker
			// Start a timer if needed. The poller will do the first check right now.
			tick := poller[0].Start()

			select { // Wait for events. Until the End signal is received from the dock.
			case <-close: // Received End event. That's all folks.
				return

			case <-tick.C: // End of timer. Reset timer.

			case <-restart: // Rechecked manually. Reset timer.
			}
		}
	}
}

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

	Actions                     Actions // Actions handler. Where events callbacks must be declared.
	onActionStart, onActionStop func()  // Before and after actions calls. Used to set display.

	*dbus.CdDbus // Dbus connector.
}

// Create a new applet manager with arguments received from command line.
//
func Applet() *CDApplet {
	//~ localdir, _ := os.Getwd()
	args := os.Args
	name := args[0][2:] // Strip ./ in the beginning.
	cda := &CDApplet{
		AppletName:    name,
		ConfFile:      args[3],
		RootDataDir:   args[4],
		ParentAppName: args[5],
		//~ ShareDataDir:  localdir,
		ShareDataDir: path.Join(args[4], AppletsDir, name),
		CdDbus:       dbus.New(args[2]),
	}

	log.SetPrefix(name)

	//~ cda._cMenuIconId = "";
	return cda
}

func (cda *CDApplet) FileLocation(filename ...string) string {
	args := append([]string{cda.ShareDataDir}, filename...)
	return path.Join(args...)
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
	cda.ControlAppli(def.MonitorName)
}

// HaveMonitor gives informations about the state of the monitored application.
// Those are usefull is this option is enabled. A monitored application, if 
// opened, is supposed to have its visibility state toggled by the user event.
// 
//  * haveApp: true if the monitored application is opened. (Xid > 0)
//  * HaveFocus: true if the monitored application is the one with the focus.
//
func (cda *CDApplet) HaveMonitor() (haveApp bool, haveFocus bool) {
	d, e := cda.GetAll()
	log.Err(e, "Got Monitor")
	return d.Xid > 0, d.HasFocus
}
