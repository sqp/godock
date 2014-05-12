// Package srvdbus provides a Dbus service (and client) for running dock applets.
package srvdbus

import (
	"github.com/guelfey/go.dbus" // imported as dbus.
	"github.com/guelfey/go.dbus/introspect"

	"github.com/sqp/godock/libs/appdbus"
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/dock" // Connection to cairo-dock.
	"github.com/sqp/godock/libs/log/color"

	"errors"
	"strings"
	"time"
)

// SrvObj is the Dbus object name for the service.
const SrvObj = "org.cairodock.GoDock"

// SrvPath is the Dbus path name for the service.
const SrvPath = "/org/cairodock/GoDock"

const intro = `
<node>
	<interface name="` + SrvObj + `">
		<signal name="ListServices"></signal>
		<signal name="RestartDock"></signal>
		<signal name="StopDock"></signal>
		<signal name="LogWindow"></signal>

		<method name="StartApplet">
			<arg direction="in" type="s"/>
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

// Log provides a common logger for the Dbus service. It must be set to use the server.
var Log cdtype.Logger

// LogWindow provides an optional call to open the log window.
var LogWindow func()

// Loader is a multi applet manager.
//
type Loader struct {
	conn     *dbus.Conn // Dbus connection.
	plopers  plopers
	apps     map[string]dock.AppletInstance        // Active applets.    Key = applet name.
	services map[string]func() dock.AppletInstance // Available applets. Key = applet name.
	c        <-chan *dbus.Signal                   // Dbus incoming signals channel.
	restart  chan string                           // Poller restart request channel.
}

// NewLoader creates a loader with the given list of applets services.
//
func NewLoader(services map[string]func() dock.AppletInstance) *Loader {
	conn, c, e := appdbus.SessionBus()
	if Log.Err(e, "DBus Connect") {
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
	reply, e := load.conn.RequestName(SrvObj, dbus.NameFlagDoNotQueue)
	if e != nil {
		return false, e
	}

	if reply != dbus.RequestNameReplyPrimaryOwner {
		return false, nil
	}

	// Everything OK, we can register our Dbus methods.
	e = load.conn.Export(load, SrvPath, SrvObj)
	Log.Err(e, "register service object")

	e = load.conn.Export(introspect.Introspectable(intro), SrvPath, "org.freedesktop.DBus.Introspectable")
	Log.Err(e, "register service introspect")

	return true, nil
}

// StartLoop handle applets (and dock) until there's no nothing more to handle.
//
func (load *Loader) StartLoop() {
	defer load.conn.ReleaseName(SrvObj)
	defer Log.Info("StopServer")
	Log.Info("StartServer")

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
	Log.Info("StopApplet", name)

}

// Forward the Dbus signal to local manager or applet
//
func (load *Loader) dispatchDbusSignal(s *dbus.Signal) bool {
	appname := pathToName(string(s.Path))
	app, ok := load.apps[appname]

	// Log.Info("received", s)

	switch {

	case strings.HasPrefix(string(s.Name), SrvObj): // Signal to applet manager.

		if len(s.Name) > len(SrvObj) {
			switch s.Name[len(SrvObj)+1:] { // Forwarded from here too so they can be called easily as signal with dbus-send.
			case "ListServices":
				load.ListServices()

			case "RestartDock":
				load.RestartDock()

			case "StopDock":
				load.StopDock()

			default:
				Log.Info("unknown service request", s.Name[len(SrvObj)+1:], s.Body)
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
		Log.Info("unknown signal", s)

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
func (load *Loader) RestartDock() *dbus.Error {
	isrestart = true
	Log.Err(RestartDock(), "restart dock")
	isrestart = false
	return nil
}

// StopDock close the dock.
//
func (load *Loader) StopDock() *dbus.Error {
	StopDock()
	return nil
}

// ListServices displays active services.
//
func (load *Loader) ListServices() *dbus.Error {
	println("Cairo-Dock applets services: active ", len(load.apps), "/", len(load.services))
	for name := range load.services {
		if _, ok := load.apps[name]; ok {
			print(color.Green(" * "))
		} else {
			print("   ")
		}
		println(name)
	}
	return nil
}

type defineEventser interface {
	DefineEvents()
}

// StartApplet creates a new applet instance with args from command line.
//
func (load *Loader) StartApplet(a, b, c, d, e, f, g, h string) *dbus.Error {
	name := pathToName(c)
	Log.Info("StartApplet", name)

	a = "./" + name
	args := []string{a, b, c, d, e, f, g, h}

	// Create applet instance.
	fn, ok := load.services[name]
	if !ok {
		return nil
	}
	app := fn()

	load.apps[name] = app

	// Define and connect applet events to the dock.
	app.SetArgs(args)

	if d, ok := app.(defineEventser); ok { // Old events callback method.
		d.DefineEvents()
	}

	app.SetEventReload(func(loadConf bool) { app.Init(loadConf) })
	er := app.ConnectEvents(load.conn)
	Log.Err(er, "ConnectEvents") // TODO: Big problem, need to handle better?

	app.RegisterEvents(app) // New events callback method.

	// Initialise applet: Load config and apply user settings.
	app.Init(true)

	if poller := app.Poller(); poller != nil {
		poller.SetChanRestart(load.restart, name) // Restart chan for user events.

		load.plopers.Add(name, poller.GetInterval()) // Set current at max for a first check ASAP.
	}
	return nil
}

type uploader interface {
	Upload(string)
}

// Upload send data (raw text or file) to a one-click hosting service.
//
func (load *Loader) Upload(data string) *dbus.Error {
	if uncast := load.GetApplet("NetActivity"); uncast != nil {
		net := uncast.(uploader)
		net.Upload(data)
	}
	return nil
}

type debuger interface {
	SetDebug(bool)
}

// Debug change the debug state of an active applet.
//
func (load *Loader) Debug(applet string, state bool) *dbus.Error {
	if uncast := load.GetApplet(applet); uncast != nil {
		app := uncast.(debuger)
		app.SetDebug(state)
	}
	return nil
}

// LogWindow opens the log terminal.
//
func (load *Loader) LogWindow() *dbus.Error {
	if LogWindow != nil {
		LogWindow()
	} else {
		Log.NewErr("no log service available", "open log window")
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
	return Log.ExecAsync("cairo-dock") // TODO: create a dedicated logger to the dock when sender becomes used.
	// cmd := exec.Command("cairo-dock")
	// cmd.Stdout = log.Logs
	// cmd.Stderr = log.Logs // TODO: need to split std and err streams.
	// return cmd.Start()
}

// cmd := exec.Command("cairo-dock")
// cmd.Stdout = logHistory
// cmd.Stderr = logHistory //os.Stderr

// if err := cmd.Start(); err != nil {
// 	logger.Err(err, "start dock")
// }

// StopDock close the dock.
//
func StopDock() error {
	return appdbus.DockQuit()
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
	dbus.Object
}

// GetServer return a connection to the active instance of the internal Dbus
// service if any. Return nil, nil if none found.
//
func GetServer() (*Client, error) {
	conn, ec := dbus.SessionBus()
	if ec != nil {
		return nil, ec
	}

	reply, e := conn.RequestName(SrvObj, dbus.NameFlagDoNotQueue)
	if e != nil {
		return nil, e
	}
	conn.ReleaseName(SrvObj)

	if reply != dbus.RequestNameReplyPrimaryOwner { // Found active instance, return client.
		return &Client{*conn.Object(SrvObj, SrvPath)}, nil
	}

	// no active instance.
	return nil, nil
}

func (cl *Client) call(method string, args ...interface{}) error {
	return cl.Call(SrvObj+"."+method, 0, args...).Err
}

// func (cl *Client) Go(method string, args ...interface{}) error {
// 	return cl.Object.Go(SrvObj+"."+method, dbus.FlagNoReplyExpected, nil, args...).Err
// }

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

// LogWindow opens the log terminal.
//
func (cl *Client) LogWindow() error {
	return cl.call("LogWindow")
}

// Send command to the active server.
//
func (load *Loader) Send(method string, args ...interface{}) error {
	busD := load.conn.Object(SrvObj, SrvPath)
	if busD == nil {
		return errors.New("can't connect to active instance")
	}

	return busD.Call(SrvObj+"."+method, 0, args...).Err
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
	prefixSize := len(appdbus.DbusPathDock) + 1 // +1 for the trailing /
	if len(path) <= prefixSize {
		return ""
	}
	text := path[prefixSize:]
	if i := strings.Index(text, "/"); i > 0 { // remove "/sub_icons" at the end if any.
		return text[:i]
	}
	return text
}
