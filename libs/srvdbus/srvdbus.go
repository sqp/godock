// Package srvdbus provides a Dbus service (and client) for running dock applets.
package srvdbus

import (
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"

	"github.com/sqp/godock/libs/appdbus"            // Dock actions.
	"github.com/sqp/godock/libs/cdtype"             // Logger type.
	"github.com/sqp/godock/libs/srvdbus/dbuscommon" // Dbus service.

	"strings"
	"time"
)

// SrvObj is the Dbus object name for the service.
const SrvObj = "org.cairodock.GoDock"

// SrvPath is the Dbus path name for the service.
const SrvPath = "/org/cairodock/GoDock"

// Introspect returns the introspect text with methods provided on the Dbus service.
func Introspect(methods string) string {
	return `
<node>
	<interface name="` + SrvObj + `">
		<signal name="StopDock"></signal>
		<signal name="LogWindow"></signal>
		<method name="Upload">
			<arg direction="in" type="s"/>
		</method>
		<method name="Debug">
			<arg direction="in" type="s"/>
			<arg direction="in" type="b"/>
		</method>
		<method name="ListServices">
			<arg direction="out" type="s"/>
		</method>` +
		methods + `
	</interface>` +
		introspect.IntrospectDataString + `
</node> `
}

// 		<signal name="RestartDock"></signal>

var log cdtype.Logger

// SetLogger provides a common logger for the Dbus service. It must be set to use the server.
//
func SetLogger(l cdtype.Logger) {
	log = l
}

// AppService defines common applets service actions to remotely interact with applets.
//
type AppService interface {
	Count() int
	GetApplets(name string) (list []cdtype.AppInstance)
	Tick()
}

// MgrDbus defines actions needed by the Dbus grouped applets manager.
//
type MgrDbus interface {
	ListServices() (string, *dbus.Error)
	IsActive(path string) bool
	OnSignal(path string, s *dbus.Signal) bool
	StartApplet(a, b, c, d, e, f, g, h string) *dbus.Error
	// RestartDock() *dbus.Error
}

// Loader is a multi applet manager.
//
type Loader struct {
	*dbuscommon.Server               // Dbus connection.
	restart            chan string   // Poller restart request channel.
	quit               chan struct{} // Manual exit chan.
	apps               AppService
	mgr                MgrDbus
	isrestart          bool // current state.

}

// NewLoader creates a loader with the given list of applets services.
//
func NewLoader(log cdtype.Logger) *Loader {
	srv := dbuscommon.NewServer(SrvObj, SrvPath, log)
	if srv == nil {
		return nil
	}
	return &Loader{
		Server:  srv,
		restart: make(chan string, 1)}
}

// SetManager sets the applet manager service.
//
func (load *Loader) SetManager(mgr AppService) {
	load.apps = mgr
	if db, ok := mgr.(MgrDbus); ok {
		load.mgr = db
	}
}

//
//--------------------------------------------------------------------[ LOOP ]--

// StartLoop handle applets (and dock) until there's no nothing more to handle.
//
func (load *Loader) StartLoop(withdock bool) {
	defer load.Conn.ReleaseName(SrvObj)
	defer load.Log.Debug("Dbus service stopped")
	load.Log.Debug("Dbus service started")

	load.quit = make(chan struct{})

	var waiter <-chan time.Time
	if load.apps != nil {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		waiter = ticker.C
	}

	for { // Start main loop and handle events until the End signal is received from the dock.

		select { // Wait for events. Until the End signal is received from the dock.

		case s := <-load.Events: // Listen to DBus events.
			if load.dispatchDbusSignal(s) { // true if signal was Stop.

				// Keep service alive if: any app alive, or we manage the dock and launched a restart manually. => false
				if load.apps.Count() == 0 && !(withdock && load.isrestart) { // That's all folks. We're closing.
					return
				}
			}

		case <-waiter: // Tick every second to update pollers counters and launch actions.
			load.apps.Tick()

		case <-load.quit:
			return
		}
	}
}

// Forward the Dbus signal to local manager or applet
//
func (load *Loader) dispatchDbusSignal(s *dbus.Signal) bool {
	path := strings.TrimSuffix(string(s.Path), "/sub_icons")

	switch {
	case s.Name == "org.freedesktop.DBus.NameAcquired": // Service started confirmed.

	case strings.HasPrefix(string(s.Name), SrvObj): // Signal to applet manager.

		if len(s.Name) > len(SrvObj) {
			switch s.Name[len(SrvObj)+1:] { // Forwarded from here too so they can be called easily as signal with dbus-send.
			case "ListServices":
				if load.mgr != nil {
					load.mgr.ListServices()
				}

			// case "RestartDock":
			// 	load.isrestart = true
			// 	load.apps.RestartDock()
			// 	load.isrestart = false

			case "StopDock":
				load.StopDock()

			default:
				load.Log.Info("unknown service request", s.Name[len(SrvObj)+1:], s.Body)
			}
		}

	case load.mgr != nil && load.mgr.IsActive(path): // Signal to applet.
		return load.mgr.OnSignal(path, s)

	default:
		load.Log.Info("unknown signal", s)

	}
	return false

}

//
//----------------------------------------------------------[ DBUS CALLBACKS ]--

// StartApplet creates a new applet instance with args from command line.
//
func (load *Loader) StartApplet(a, b, c, d, e, f, g, h string) *dbus.Error {
	if load.mgr != nil {
		return load.mgr.StartApplet(a, b, c, d, e, f, g, h)
	}
	return nil
}

// StopDock close the dock.
//
func (load *Loader) StopDock() *dbus.Error {
	load.quit <- struct{}{} // Release the Dbus service ASAP so it could be captured by a restarted dock.
	// load.Conn.Close()

	appdbus.DockQuit()
	return nil
}

type uploader interface {
	Upload(string)
}

// Upload send data (raw text or file) to a one-click hosting service.
//
func (load *Loader) Upload(data string) *dbus.Error {
	if load.apps == nil {
		return nil
	}

	uncasts := load.apps.GetApplets("NetActivity")
	if len(uncasts) > 0 {
		app := uncasts[0].(uploader) // Send it to the first found. Should be safe for now, we can launch only one.
		app.Upload(data)
	}
	return nil
}

// Debug change the debug state of an active applet.
//
func (load *Loader) Debug(applet string, state bool) *dbus.Error {
	if load.apps == nil {
		return nil
	}

	for _, app := range load.apps.GetApplets(applet) {
		app.SetDebug(true)
	}
	return nil
}

//
//------------------------------------------------------[ DBUS SEND COMMANDS ]--

// Client is a Dbus client to connect to the internal Dbus server.
//
type Client struct {
	*dbuscommon.Client
}

// Action forwards a simple client action to the active applets service.
//
func Action(action func(*Client) error) error {
	client, e := dbuscommon.GetClient(SrvObj, SrvPath)
	if e != nil {
		return e
	}
	return action(&Client{client}) // we have a server, launch the provided action.
}

// RestartDock forwards action to restart the dock.
//
// func (cl *Client) RestartDock() error {
// 	return cl.Call("RestartDock")
// }

// StopDock forwards action to stop the dock.
//
func (cl *Client) StopDock() error {
	return cl.Call("StopDock")
}

// LogWindow forwards action to opens the log terminal.
//
func (cl *Client) LogWindow() error {
	return cl.Call("LogWindow")
}

// Debug forwards action set debug to a remote applet.
//
func Debug(applet string, state bool) error {
	client, e := dbuscommon.GetClient(SrvObj, SrvPath)
	if e != nil {
		return e
	}
	return client.Call("Debug", applet, state)
}

// ListServices forwards action to get the list of active services.
//
func ListServices() (string, error) {
	client, e := dbuscommon.GetClient(SrvObj, SrvPath)
	if e != nil {
		return "", e
	}

	call := client.Object.Call(SrvObj+"."+"ListServices", 0)
	if call.Err != nil {
		return "", call.Err
	}
	str := ""
	e = call.Store(&str)
	return str, e
}

// Upload forwards action upload data to the dock.
//
func Upload(data string) error {
	client, e := dbuscommon.GetClient(SrvObj, SrvPath)
	if e != nil {
		return e
	}
	return client.Call("Upload", data)
}
