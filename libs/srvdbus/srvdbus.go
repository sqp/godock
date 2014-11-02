// Package srvdbus provides a Dbus service (and client) for running dock applets.
package srvdbus

import (
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"

	"github.com/sqp/godock/libs/appdbus"   // Dock actions.
	"github.com/sqp/godock/libs/cdtype"    // Logger type.
	"github.com/sqp/godock/libs/dock"      // Connection to cairo-dock.
	"github.com/sqp/godock/libs/log/color" // Colored text.

	"errors"
	"strconv"
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
		<method name="ListServices">
			<arg direction="out" type="s"/>
		</method>
	</interface>` +
	introspect.IntrospectDataString + `
</node> `

var log cdtype.Logger

// SetLogger provides a common logger for the Dbus service. It must be set to use the server.
//
func SetLogger(l cdtype.Logger) {
	log = l
}

// LogWindow provides an optional call to open the log window.
var LogWindow func()

// Loader is a multi applet manager.
//
type Loader struct {
	conn     *dbus.Conn // Dbus connection.
	plopers  plopers
	actives  map[string]*activeApp                 // Active applets.    Key = applet dbus path (/org/cairodock/CairoDock/appletName).
	services map[string]func() dock.AppletInstance // Available applets. Key = applet name.
	c        <-chan *dbus.Signal                   // Dbus incoming signals channel.
	restart  chan string                           // Poller restart request channel.
}

type activeApp struct {
	app  dock.AppletInstance
	name string
}

// NewLoader creates a loader with the given list of applets services.
//
func NewLoader(services map[string]func() dock.AppletInstance) *Loader {
	conn, c, e := appdbus.SessionBus()
	if log.Err(e, "DBus Connect") {
		return nil
	}

	load := &Loader{
		conn:     conn,
		c:        c,
		restart:  make(chan string, 1),
		services: services,
		plopers:  newPlopers(),
		actives:  make(map[string]*activeApp)}

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
	log.Err(e, "register service object")

	e = load.conn.Export(introspect.Introspectable(intro), SrvPath, "org.freedesktop.DBus.Introspectable")
	log.Err(e, "register service introspect")

	return true, nil
}

// StartLoop handle applets (and dock) until there's no nothing more to handle.
//
func (load *Loader) StartLoop(withdock bool) {
	defer load.conn.ReleaseName(SrvObj)
	defer log.Debug("Applets service stopped")
	log.Debug("Applets service started")

	action := true
	var waiter <-chan time.Time

	for { // Start main loop and handle events until the End signal is received from the dock.
		if action {
			for path, ref := range load.actives {
				poller := ref.app.Poller()
				if poller != nil && load.plopers.Test(path, poller.GetInterval()) {
					go poller.Action()
				}
			}

			action = false
			waiter = time.After(time.Second)
		}

		select { // Wait for events. Until the End signal is received from the dock.

		case s := <-load.c: // Listen to DBus events.
			if load.dispatchDbusSignal(s) { // true if signal was Stop.

				// Keep service alive if: any app alive, or we manage the dock and launched a restart manually. => false
				if len(load.actives) == 0 && !(withdock && isrestart) { // That's all folks. We're closing.
					return
				}

				// return
			}

		case path := <-load.restart: // Wait for manual restart poller event. Reloop and check.
			go load.actives[path].app.Poller().Action()
			load.plopers.Add(path, 0) // reset timer for given poller.

		case <-waiter: // Wait for the end of the timer. Reloop and check.
			action = true
		}
	}
}

// StopApplet close the applet instance.
//
func (load *Loader) StopApplet(path string) {

	// unregister events?

	name := load.actives[path].name
	load.plopers.Remove(path)
	delete(load.actives, path)
	log.Debug("StopApplet", name)
}

// Forward the Dbus signal to local manager or applet
//
func (load *Loader) dispatchDbusSignal(s *dbus.Signal) bool {
	path := strings.TrimSuffix(string(s.Path), "/sub_icons")

	ref, ok := load.actives[path]

	switch {
	case s.Name == "org.freedesktop.DBus.NameAcquired": // Service started confirmed.

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
				log.Info("unknown service request", s.Name[len(SrvObj)+1:], s.Body)
			}
		}

	case ok: // Signal to applet.
		if ref.app.OnSignal(s) {
			load.StopApplet(path) // Signal was stop_module.
			return true
		}

	default:
		log.Info("unknown signal", s)

	}
	return false

}

// GetApplets return an applet instance.
//
func (load *Loader) GetApplets(name string) (list []dock.AppletInstance) {
	for _, ref := range load.actives {
		if ref.name == name && ref.app != nil {
			list = append(list, ref.app)
		}
	}
	return
}

//
//----------------------------------------------------------[ DBUS CALLBACKS ]--

// RestartDock is a full restart of the dock, respawned in the same location if
// it was managed.
//
func (load *Loader) RestartDock() *dbus.Error {
	isrestart = true
	log.Err(RestartDock(), "restart dock")
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
func (load *Loader) ListServices() (string, *dbus.Error) {
	list := make(map[string]int)
	for _, ref := range load.actives {
		list[ref.name]++
	}

	str := "Cairo-Dock applets services: active " + strconv.Itoa(len(list)) + "/" + strconv.Itoa(len(load.services))
	for name := range load.services {
		count := list[name]
		switch {
		case count > 1:
			str += "\n" + color.Green(" * ") + name + ":" + color.Green(strconv.Itoa(count))
		case count == 1:
			str += "\n" + color.Green(" * ") + name
		default:
			str += "\n" + "   " + name
		}
	}
	if len(load.actives) > len(list) {
		str += "\n" + "Total applets started: " + strconv.Itoa(len(load.actives))
	}
	return str, nil
}

type defineEventser interface {
	DefineEvents()
}

// StartApplet creates a new applet instance with args from command line.
//
func (load *Loader) StartApplet(a, b, c, d, e, f, g, h string) *dbus.Error {
	split := strings.Split(c, "/")
	if len(split) < 4 {
		log.NewErr("StartApplet: incorrect dbus path", c)
		return nil
	}
	name := split[4] //path is /org/cairodock/CairoDock/appletName or  /org/cairodock/CairoDock/appletName/sub_icons

	a = "./" + name // reformat the launcher name as if it was directly called from shell.
	args := []string{a, b, c, d, e, f, g, h}

	if _, ok := load.actives[c]; ok {
		log.NewErr("StartApplet: applet already started", name)
		return nil
	}

	fn, ok := load.services[name]
	if !ok {
		log.NewErr(strings.Join(args, " "), "StartApplet: applet unknown (maybe not enabled at compile)")
		return nil
	}

	// Create applet instance.
	log.Debug("StartApplet", name)
	app := fn()
	load.actives[c] = &activeApp{app: app, name: name}

	app.(debugger).SetDebug(log.GetDebug()) // If the service debug is active, forward it to all applets.

	// Define and connect applet events to the dock.
	app.SetArgs(args)

	if d, ok := app.(defineEventser); ok { // Old events callback method.
		d.DefineEvents()
	}

	app.SetEventReload(func(loadConf bool) { app.Init(loadConf) })
	er := app.ConnectEvents(load.conn)
	log.Err(er, "ConnectEvents") // TODO: Big problem, need to handle better?

	app.RegisterEvents(app) // New events callback method.

	// Initialise applet: Load config and apply user settings.
	app.Init(true)

	if poller := app.Poller(); poller != nil {
		poller.SetChanRestart(load.restart, c) // Restart chan for user events.

		load.plopers.Add(c, poller.GetInterval()) // Set current at max for a first check ASAP.
	}
	return nil
}

type uploader interface {
	Upload(string)
}

// Upload send data (raw text or file) to a one-click hosting service.
//
func (load *Loader) Upload(data string) *dbus.Error {
	uncasts := load.GetApplets("NetActivity")
	if len(uncasts) == 0 {
		return nil
	}

	app := uncasts[0].(uploader)
	app.Upload(data)
	return nil
}

type debugger interface {
	SetDebug(bool)
}

// Debug change the debug state of an active applet.
//
func (load *Loader) Debug(applet string, state bool) *dbus.Error {
	for _, uncast := range load.GetApplets(applet) {
		app := uncast.(debugger)
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
		log.NewErr("no log service available", "open log window")
	}
	return nil
}

//
//------------------------------------------------------------[ DOCK CONTROL ]--

// current state.
var isrestart bool

// StartDock launch the dock.
//
func StartDock() error {
	return log.ExecAsync("cairo-dock") // TODO: create a dedicated logger to the dock when sender becomes used.
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
func (cl *Client) ListServices() (string, error) {
	str := ""
	call := cl.Call(SrvObj+"."+"ListServices", 0)
	if call.Err != nil {
		return "", call.Err
	}
	e := call.Store(&str)
	return str, e
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
	if pl[name] >= interval {
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
