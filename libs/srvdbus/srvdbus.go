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
	</interface>` + introspect.IntrospectDataString + `</node> `

type Loader struct {
	conn     *apiDbus.Conn // Dbus connection.
	plopers  plopers
	apps     map[string]dock.AppletInstance        // Active applets.    Key = applet name.
	services map[string]func() dock.AppletInstance // Available applets. Key = applet name.
	c        <-chan *apiDbus.Signal                // Dbus incoming signals channel.
	restart  chan string                           // Poller restart request channel.
}

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
		plopers:  NewPlopers(),
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

func (load *Loader) StartLoop() {
	defer load.conn.ReleaseName(localPath)
	defer log.Info("StopServer")
	log.Info("StartServer")

	action := true
	name := ""
	for { // Start main loop and handle events until the End signal is received from the dock.
		if action {
			if name != "" {
				load.plopers.Add(name, load.apps[name].Poller().GetInterval()) // hack?
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

func (load *Loader) StopApplet(name string) {

	// unregister events?

	load.plopers.Remove(name)

	delete(load.apps, name)
	log.Info("StopApplet", name)

}

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

//
//----------------------------------------------------------[ DBUS CALLBACKS ]--

func (load *Loader) RestartDock() *apiDbus.Error {
	isrestart = true
	StopDock()
	StartDock()
	isrestart = false
	return nil
}

func (load *Loader) StopDock() *apiDbus.Error {
	StopDock()
	return nil
}

func (load *Loader) ListServices() *apiDbus.Error {
	log.Info("Active services", len(load.apps), "/", len(load.services))
	for name, _ := range load.apps {
		log.Info("  " + name)
	}
	return nil
}

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

//
//------------------------------------------------------[ DBUS SEND COMMANDS ]--

// current state.
var withdock bool
var isrestart bool

func StartDock() {
	withdock = true
	// exec.Command("nohup", "cairo-dock").Start()
	cmd := exec.Command("cairo-dock")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Err(cmd.Start(), "Launch dock")
}

func StopDock() {
	log.Err(dbus.DockQuit(), "DockQuit")
}

type Client struct {
	apiDbus.Object
}

// Get connection to active instance.
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

func (cl *Client) ListServices() error {
	return cl.call("ListServices")
}

func (cl *Client) RestartDock() error {
	return cl.call("RestartDock")
}

func (cl *Client) StopDock() error {
	return cl.call("StopDock")
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

func NewPlopers() plopers {
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

func (pl plopers) Add(name string, interval int) {
	pl[name] = interval
}

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
