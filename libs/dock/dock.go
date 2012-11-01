/*
Package dock is the main Cairo-Dock applet manager, currently tightly linked to the DBus implementation.

*/
package dock

import (
	"os"
	"path"
	//~ "time"
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/dbus" // Connection to cairo-dock.
	"github.com/sqp/godock/libs/log"  // Display info in terminal.
	"github.com/sqp/godock/libs/poller"
)

// Methods an applet must implement to use the StartApplet func.
//
type AppletInstance interface {
	DefineEvents()
	ConnectToBus() error
	Init(loadConf bool)
	GetCloseChan() chan bool
}

//---------------------------------------------------------------[ MAIN CALL ]--

func NewApplet() *CDApplet {
	app := Applet()
	log.SetPrefix(app.AppletName)
	return app
}

const AppletsDir = "third-party"

type CDApplet struct {
	AppletName    string
	ConfFile      string
	ParentAppName string
	ShareDataDir  string
	RootDataDir   string
	//~ _cMenuIconId string

	Actions                     Actions // Actions handler.
	onActionStart, onActionStop func()  // before and after actions calls. Used to set display.

	*dbus.CdDbus // Dbus connector.
}

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

	//~ cda._cMenuIconId = "";
	return cda
}

func (cda *CDApplet) FileLocation(filename ...string) string {
	args := append([]string{cda.ShareDataDir}, filename...)
	return path.Join(args...)
}

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

func (cda *CDApplet) HaveMonitor() (bool, bool) {
	d, e := cda.GetAll()
	log.Err(e, "Got Monitor")
	return d.Xid > 0, d.HasFocus
}

// StartApplet will prepare and launch a cairo-dock applet. If you have provided
// events, they will respond when needed, and you have nothing more to worry about
// your applet management. One optional poller can be provided atm. 
// 
// List of the steps, and their effect:
//   * Load applet events definition = DefineEvents().
//   * Connect the applet to cairo-dock with DBus. This also activate events callbacks.
//   * Initialise applet with option load config activated = Init(true).
//   * Start and run the polling loop if needed. This start a instant check, and 
//     manage automatic or manual timer refresh.
//   * Wait for the dock End signal to close the applet.
//
//
func StartApplet(app AppletInstance, poller ...*poller.Poller) {
	log.Debug("Applet started")
	defer log.Debug("Applet stopped")

	// Define and connect events to the dock
	app.DefineEvents()
	app.ConnectToBus()

	// Load config and apply user settings.
	app.Init(true)

	// Prepare signals channels.
	close := app.GetCloseChan()

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
