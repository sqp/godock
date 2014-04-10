package srvdbus

import (
	apiDbus "github.com/guelfey/go.dbus"
	"github.com/guelfey/go.dbus/introspect"

	"github.com/sqp/godock/libs/dbus" // Connection to cairo-dock.
	"github.com/sqp/godock/libs/dock" // Connection to cairo-dock.
	"github.com/sqp/godock/libs/log"  // Display info in terminal.
	// "github.com/sqp/godock/libs/poller"

	"errors"
	"os"
	"os/exec"
	"strings"
	"time"
)

const localPath = "org.cairodock.GoDock"
const grr = "/org/cairodock/GoDock"

const prefixName = len("/org/cairodock/CairoDock/")

const intro = `
<node>
	<interface name="` + localPath + `">
		<signal name="ListServices"></signal>
		<signal name="RestartDock"></signal>
		<signal name="StopDock"></signal>

		<method name="StartApplet">
			<arg direction="in" type="s"/>
			<arg direction="in" type="s"/>
			<arg direction="in" type="s"/>
			<arg direction="in" type="s"/>
			<arg direction="in" type="s"/>
			<arg direction="in" type="s"/>
			<arg direction="in" type="s"/>
		</method>
		<method name="Upload">
			<arg direction="in" type="s"/>
		</method>
		<method name="Debug">
			<arg direction="in" type="s"/>
			<arg direction="in" type="b"/>
		</method>
	</interface>` + introspect.IntrospectDataString + `</node> `

// Loader is a multi applet manager.
//
type Loader struct {
	conn     *apiDbus.Conn // Dbus connection.
	plopers  plopers
	apps     map[string]dock.AppletInstance        // Active applets.    Key = applet name.
	services map[string]func() dock.AppletInstance // Available applets. Key = applet name.
	c        <-chan *apiDbus.Signal                // Dbus incoming signals channel.
	restart  chan string                           // Poller restart request channel.
}

// NewLoader creates a loader with the given list of applets services.
//
func NewLoader(services map[string]func() dock.AppletInstance) *Loader {
	conn, c, e := dbus.SessionBus()
	if log.Err(e, "DBus Connect") {
		return nil
	}

	load := &Loader{
		conn:     conn,
		c:        c,
		restart:  make(chan string, 1),
		services: services,
		plopers:  newPlopers(),
		apps:     make(map[string]dock.AppletInstance)}

	return load
}

// StartServer will try to start and manage the applets server. You must provide
// the applet arguments used to launch the applet.
// If a server was already active, the applet start request is forwarded and
// no loop will be started, the function just return with the error if any.
//
func (load *Loader) StartServer() (bool, error) {
	reply, e := load.conn.RequestName(localPath, apiDbus.NameFlagDoNotQueue)
	if e != nil {
		return false, e
	}

	if reply != apiDbus.RequestNameReplyPrimaryOwner {
		return false, nil
	}

	// Everything OK, we can register our Dbus methods, start the first applet,
	// and start the loop to handle applets events and other.

	e = load.conn.Export(load, grr, localPath)
	log.Err(e, "reg")

	e = load.conn.Export(introspect.Introspectable(intro), grr, "org.freedesktop.DBus.Introspectable")
	log.Err(e, "introspect")

	return true, nil
}

// StartLoop handle applets (and dock) until there's no nothing more to handle.
//
func (load *Loader) StartLoop() {
	defer load.conn.ReleaseName(localPath)
	defer log.Info("StopServer")
	log.Info("StartServer")

	action := true
	name := ""
	for { // Start main loop and handle events until the End signal is received from the dock.
		if action {
			if name != "" {
				load.plopers.Add(name, load.apps[name].Poller().GetInterval()) // set time to max for given poller so it will be recheck. Can be considerer as a hack?
				name = ""
			}

			for name, app := range load.apps {
				poller := app.Poller()
				if poller != nil && load.plopers.Test(name, poller.GetInterval()) {
					go poller.Action()
				}
			}
			action = false
		}

		waiter := time.After(time.Second)

		select { // Wait for events. Until the End signal is received from the dock.

		case s := <-load.c: // Listen to DBus events.
			if load.dispatchDbusSignal(s) {
				return
			}

		case name = <-load.restart: // Wait for manual restart poller event. Reloop and check.
			action = true

		case <-waiter: // Wait for the end of the timer. Reloop and check.
			action = true
		}
	}
}

// StopApplet close the applet instance.
//
func (load *Loader) StopApplet(name string) {

	// unregister events?

	load.plopers.Remove(name)

	delete(load.apps, name)
	log.Info("StopApplet", name)

}

// Forward the Dbus signal to local manager or applet
//
func (load *Loader) dispatchDbusSignal(s *apiDbus.Signal) bool {
	appname := pathToName(string(s.Path))
	app, ok := load.apps[appname]

	// log.Info("received", s)

	switch {

	case strings.HasPrefix(string(s.Name), localPath): // Signal to applet manager.

		if len(s.Name) > len(localPath) {
			switch s.Name[len(localPath)+1:] { // Forwarded from here too so they can be called easily as signal with dbus-send.
			case "ListServices":
				load.ListServices()

			case "RestartDock":
				load.RestartDock()

			case "StopDock":
				load.StopDock()

			default:
				log.Info("unknown service request", s.Name[len(localPath)+1:], s.Body)
			}
		}

	case ok: // Signal to applet.
		if app.OnSignal(s) {
			load.StopApplet(appname) // Signal was stop_module.

			// Keep service alive if: any app alive, or we manage the dock and launched a restart manually. => false
			if len(load.apps) == 0 && !(withdock && isrestart) { // That's all folks. We're closing.
				return true
			}
		}

	default:
		log.Info("unknown", s)

	}
	return false

}

// GetApplet return an applet instance.
//
func (load *Loader) GetApplet(name string) dock.AppletInstance {
	return load.apps[name]
}

//
//----------------------------------------------------------[ DBUS CALLBACKS ]--

// RestartDock is a full restart of the dock, respawned in the same location if
// it was managed.
//
func (load *Loader) RestartDock() *apiDbus.Error {
	isrestart = true
	log.Err(RestartDock(), "restart dock")
	isrestart = false
	return nil
}

// StopDock close the dock.
//
func (load *Loader) StopDock() *apiDbus.Error {
	StopDock()
	return nil
}

// ListServices displays active services.
//
func (load *Loader) ListServices() *apiDbus.Error {
	println("Cairo-Dock applets services: active ", len(load.apps), "/", len(load.services))
	for name, _ := range load.services {
		if _, ok := load.apps[name]; ok {
			print(log.Colored(" * ", log.FgGreen))
		} else {
			print("   ")
		}
		println(name)
	}
	return nil
}

// StartApplet creates a new applet instance with args from command line.
//
func (load *Loader) StartApplet(a, b, c, d, e, f, g string) *apiDbus.Error {
	name := pathToName(c)
	log.Info("StartApplet", name)

	a = "./" + name
	args := []string{a, b, c, d, e, f, g}

	// Create applet instance.
	fn, ok := load.services[name]
	if !ok {
		return nil
	}
	app := fn()

	load.apps[name] = app

	// Define and connect applet events to the dock.
	app.SetArgs(args)
	app.DefineEvents()
	app.SetEventReload(func(loadConf bool) { app.Init(loadConf) })
	er := app.ConnectEvents(load.conn)
	log.Err(er, "ConnectEvents") // TODO: Big problem, need to handle better?

	// Initialise applet: Load config and apply user settings.
	app.Init(true)

	if poller := app.Poller(); poller != nil {
		poller.SetChanRestart(load.restart, name) // Restart chan for user events.

		load.plopers.Add(name, poller.GetInterval()) // Set current at max for a first check ASAP.
	}
	return nil
}

type Uploader interface {
	Upload(string)
}

// Upload send data (raw text or file) to a one-click hosting service.
//
func (load *Loader) Upload(data string) *apiDbus.Error {
	if uncast := load.GetApplet("NetActivity"); uncast != nil {
		net := uncast.(Uploader)
		net.Upload(data)
	}
	return nil
}

type Debuger interface {
	SetDebug(bool)
}

// Debug change the debug state of an active applet.
//
func (load *Loader) Debug(applet string, state bool) *apiDbus.Error {
	if uncast := load.GetApplet(applet); uncast != nil {
		app := uncast.(Debuger)
		app.SetDebug(state)
	}
	return nil
}

//
//------------------------------------------------------------[ DOCK CONTROL ]--

// current state.
var withdock bool
var isrestart bool

// StartDock launch the dock.
//
func StartDock() error {
	withdock = true
	// exec.Command("nohup", "cairo-dock").Start()
	cmd := exec.Command("cairo-dock")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Start()
}

// StopDock close the dock.
//
func StopDock() error {
	return dbus.DockQuit()
}

// RestartDock close and relaunch the dock.
//
func RestartDock() error {
	if e := StopDock(); e != nil {
		return e
	}
	return StartDock()
}

//
//------------------------------------------------------[ DBUS SEND COMMANDS ]--

// Client is a Dbus client to connect to the internal Dbus server.
//
type Client struct {
	apiDbus.Object
}

// GetServer return a connection to the active instance of the internal Dbus
// service if any. Return nil, nil if none found.
//
func GetServer() (*Client, error) {
	conn, _, e := dbus.SessionBus() // TODO: get better.
	// close(c)

	if e != nil {
		return nil, e
	}

	reply, e := conn.RequestName(localPath, apiDbus.NameFlagDoNotQueue)
	if e != nil {
		return nil, e
	}
	defer conn.ReleaseName(localPath)

	if reply != apiDbus.RequestNameReplyPrimaryOwner { // Found active instance, return client.
		return &Client{*conn.Object(localPath, grr)}, nil
	}

	// no active instance.
	return nil, nil
}

func (cl *Client) call(method string, args ...interface{}) error {
	return cl.Call(localPath+"."+method, 0, args...).Err
}

// ListServices send action to displays active services.
//
func (cl *Client) ListServices() error {
	return cl.call("ListServices")
}

// RestartDock send action to restart the dock.
//
func (cl *Client) RestartDock() error {
	return cl.call("RestartDock")
}

// StopDock send action to stop the dock.
//
func (cl *Client) StopDock() error {
	return cl.call("StopDock")
}

// Upload send action upload data to the dock.
//
func (cl *Client) Upload(data string) error {
	return cl.call("Upload", data)
}

// Debug send action upload data to the dock.
//
func (cl *Client) Debug(applet string, state bool) error {
	return cl.call("Debug", applet, state)
}

// Send command to the active server.
//
func (load *Loader) Send(method string, args ...interface{}) error {
	busD := load.conn.Object(localPath, grr)
	if busD == nil {
		return errors.New("Can't connect to active instance")
	}

	return busD.Call(localPath+"."+method, 0, args...).Err
}

//
//----------------------------------------------------------[ PLOPERS ]--

// plopers is a simple counter to know when to launch each polling action.
//
type plopers map[string]int // Active pollers.    Key = applet name.

func newPlopers() plopers {
	return make(plopers)
}

// Test increase the counter and return true if the counter reached interval.
// Otherwise the counter is increased.
//
func (pl plopers) Test(name string, interval int) bool {
	// if _, ok := pl[name]; !ok {
	// 	return false
	// }
	pl[name]++
	if pl[name] > interval {
		pl[name] = 0
		return true
	}
	return false
}

// Add or update a counter reference. Use max value for a check ASAP.
//
func (pl plopers) Add(name string, current int) {
	pl[name] = current
}

// Remove a counter reference.
//
func (pl plopers) Remove(name string) {
	if _, ok := pl[name]; ok {
		delete(pl, name)
	}
}

//
//-----------------------------------------------------------------[ HELPERS ]--

func pathToName(path string) string {
	if len(path) > prefixName {
		return path[prefixName:]
	}
	return ""
}
