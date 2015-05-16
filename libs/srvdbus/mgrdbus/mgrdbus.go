package mgrdbus

import (
	"github.com/godbus/dbus"

	"github.com/sqp/godock/libs/appdbus"            // Dock actions.
	"github.com/sqp/godock/libs/cdtype"             // AppStarter.
	"github.com/sqp/godock/libs/log/color"          // Colored text.
	"github.com/sqp/godock/libs/srvdbus"            // Dbus paths.
	"github.com/sqp/godock/libs/srvdbus/dbuscommon" // Dbus service.

	"strconv"
	"strings"
)

var IntrospectApplets = `
	<method name="StartApplet">
			<arg direction="in" type="s"/>
			<arg direction="in" type="s"/>
			<arg direction="in" type="s"/>
			<arg direction="in" type="s"/>
			<arg direction="in" type="s"/>
			<arg direction="in" type="s"/>
			<arg direction="in" type="s"/>
			<arg direction="in" type="s"/>
		</method>`

// activeApp holds a reference to an active applet instance.
//
type activeApp struct {
	app     cdtype.AppInstance // active instance.
	name    string             // applet name.
	backend *appdbus.CDDbus
}

// Manager is an external applets manager for cairo-dock.
//
type Manager struct {
	actives  map[string]*activeApp // Active applets.    Key = applet dbus path (/org/cairodock/CairoDock/appletName).
	services cdtype.ListStarter    // Available applets. Key = applet name.
	conn     *dbus.Conn
	log      cdtype.Logger
}

// NewManager creates a loader with the given list of applets services.
//
func NewManager(services cdtype.ListStarter, conn *dbus.Conn, log cdtype.Logger) *Manager {
	return &Manager{
		services: services,
		actives:  make(map[string]*activeApp),
		conn:     conn,
		log:      log}
}

// Count returns the number of managed applets.
//
func (load *Manager) Count() int {
	return len(load.actives)
}

// IsActive returns whether the given applet path is active or not.
//
func (load *Manager) IsActive(path string) bool {
	_, ok := load.actives[path]
	return ok
}

// Tick ticks all applets pollers.
//
func (load *Manager) Tick() {
	for _, ref := range load.actives {
		ref.app.Poller().Plop() // Safe to use on nil poller.
	}
}

// OnSignal forwards a signal event to the applet backend.
//
func (load *Manager) OnSignal(path string, s *dbus.Signal) bool {
	ref := load.actives[path]

	if ref.backend.OnSignal(s) {
		load.StopApplet(path) // Signal was stop_module.
		return true
	}
	return false
}

// StartApplet creates a new applet instance with args from command line.
//
func (load *Manager) StartApplet(a, b, c, d, e, f, g, h string) *dbus.Error {
	split := strings.Split(c, "/")
	if len(split) < 4 {
		load.log.NewErr("StartApplet: incorrect dbus path", c)
		return nil
	}
	name := split[4] //path is /org/cairodock/CairoDock/appletName or  /org/cairodock/CairoDock/appletName/sub_icons

	a = "./" + name // reformat the launcher name as if it was directly called from shell.
	args := []string{a, b, c, d, e, f, g, h}

	if _, ok := load.actives[c]; ok {
		load.log.NewErr("StartApplet: applet already started", name)
		return nil
	}

	fn, ok := load.services[name]
	if !ok {
		load.log.NewErr(strings.Join(args, " "), "StartApplet: applet unknown (maybe not enabled at compile)")
		return nil
	}

	// Create applet instance.
	load.log.Debug("StartApplet", name)
	app := fn()

	backend := appdbus.NewWithApp(app, args, h)

	er := backend.ConnectEvents(load.conn)
	load.log.Err(er, "ConnectEvents") // TODO: Big problem, need to handle better?

	load.actives[c] = &activeApp{
		app:     app,
		name:    name,
		backend: backend,
	}

	// Initialise applet: Load config and apply user settings.
	app.Init(true)

	if load.log.GetDebug() { // If the service debug is active, force it also on applets.
		app.SetDebug(true)
	}
	app.Poller().Restart() // check poller now if it exists. Safe to use on nil poller.
	return nil
}

// StopApplet close the applet instance.
//
func (load *Manager) StopApplet(path string) {

	// unregister events?

	name := load.actives[path].name
	// load.plopers.Remove(path)
	delete(load.actives, path)
	load.log.Debug("StopApplet", name)
}

// GetApplets return an applet instance.
//
func (load *Manager) GetApplets(name string) (list []cdtype.AppInstance) {
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
// func (load *Manager) RestartDock() *dbus.Error {
// 	load.log.Err(RestartDock(), "restart dock")
// 	return nil
// }

// ListServices displays active services.
//
func (load *Manager) ListServices() (string, *dbus.Error) {
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

// LogWindow provides an optional call to open the log window.
// var LogWindow func()

// LogWindow opens the log terminal.
//
// func (load *Manager) LogWindow() *dbus.Error {
// 	if LogWindow != nil {
// 		LogWindow()
// 	} else {
// 		load.log.NewErr("no log service available", "open log window")
// 	}
// 	return nil
// }

//
//------------------------------------------------------------[ DOCK CONTROL ]--

// StartApplet forwards action to start a new applet.
// Args are those sent by the dock in the started applet command line.
//
func StartApplet(a, b, c, d, e, f, g string) error {
	client, err := dbuscommon.GetClient(srvdbus.SrvObj, srvdbus.SrvPath)
	if err != nil {
		return err
	}
	return client.Call("StartApplet", "", a, b, c, d, e, f, g)
}

// StartDock launch the dock.
//
// func StartDock() error {
// 	// TODO: use loader logger.
// 	return log.ExecAsync("cairo-dock") // TODO: create a dedicated logger to the dock when sender becomes used.

// 	// cmd := exec.Command("cairo-dock")
// 	// cmd.Stdout = log.Logs
// 	// cmd.Stderr = log.Logs // TODO: need to split std and err streams.
// 	// return cmd.Start()
// }

// // cmd := exec.Command("cairo-dock")
// // cmd.Stdout = logHistory
// // cmd.Stderr = logHistory //os.Stderr

// // if err := cmd.Start(); err != nil {
// // 	logger.Err(err, "start dock")
// // }

// StopDock close the dock.
//
// func StopDock() error {
// 	return appdbus.DockQuit()
// }

// // RestartDock close and relaunch the dock.
// //
// func RestartDock() error {
// 	if e := StopDock(); e != nil {
// 		return e
// 	}
// 	return StartDock()
// }

//
//-----------------------------------------------------------------[ HELPERS ]--
