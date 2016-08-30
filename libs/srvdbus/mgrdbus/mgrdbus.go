// Package mgrdbus provides a Dbus service (and client) for external applets management.
package mgrdbus

import (
	"github.com/godbus/dbus"

	"github.com/sqp/godock/libs/appdbus"            // Dock actions.
	"github.com/sqp/godock/libs/cdtype"             // AppStarter.
	"github.com/sqp/godock/libs/srvdbus"            // Dbus paths.
	"github.com/sqp/godock/libs/srvdbus/dbuscommon" // Dbus service.
	"github.com/sqp/godock/libs/text/color"         // Colored text.

	"strconv"
	"strings"
)

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
	*srvdbus.Loader // Extends the applet service loader to provide its methods on the bus.

	actives map[string]*activeApp // Active applets.    Key = applet dbus path (/org/cairodock/CairoDock/appletName).
	log     cdtype.Logger
}

// NewManager creates a loader with the given list of applets services.
//
func NewManager(loader *srvdbus.Loader, log cdtype.Logger) *Manager {
	return &Manager{
		Loader:  loader,
		actives: make(map[string]*activeApp),
		log:     log}
}

// CountActive returns the number of managed applets.
//
func (load *Manager) CountActive() int {
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
		return dbuscommon.NewError("StartApplet: incorrect dbus path " + c)
	}
	name := split[4] //path is /org/cairodock/CairoDock/appletName or  /org/cairodock/CairoDock/appletName/sub_icons

	a = "./" + name // reformat the launcher name as if it was directly called from shell.
	args := []string{a, b, c, d, e, f, g, h}

	if _, ok := load.actives[c]; ok {
		load.log.NewErr("StartApplet: applet already started", name)
		return dbuscommon.NewError("StartApplet: applet already started " + name)
	}

	if cdtype.Applets.GetNewFunc(name) == nil {
		load.log.NewErr(strings.Join(args, " "), "StartApplet: applet unknown (maybe not enabled at compile)")
		return dbuscommon.NewError("StartApplet: applet unknown (maybe not enabled at compile) " + strings.Join(args, " "))
	}

	// Create applet instance.
	// name := args[0][2:] // Strip ./ in the beginning.
	callnew := cdtype.Applets.GetNewFunc(name)
	app, backend, callinit := appdbus.New(callnew, args, h)
	if app == nil {
		load.log.NewErr(name, "failed to create applet")
		return dbuscommon.NewError("failed to create applet" + name)
	}

	er := backend.ConnectEvents(load.Loader.Conn)
	if app.Log().Err(er, "ConnectEvents") {
		return dbuscommon.NewError("ConnectEvents: " + er.Error())
	}

	load.actives[c] = &activeApp{
		app:     app,
		name:    name,
		backend: backend,
	}

	if load.log.GetDebug() { // If the service debug is active, force it also on applets.
		app.Log().SetDebug(true)
	}
	load.log.Debug("StartApplet", name)

	// Initialise applet: Load config and apply user settings.
	// Find a way to unload the applet without the crash in dock and DBus service mode.
	er = callinit()
	if load.log.Err(er, "init applet") {
		return dbuscommon.NewError("failed to create applet" + name)
	}

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

// ListServices displays active services.
//
func (load *Manager) ListServices() (string, *dbus.Error) {
	list := make(map[string]int)
	for _, ref := range load.actives {
		list[ref.name]++
	}

	str := "Cairo-Dock applets services: active " + strconv.Itoa(len(list)) + "/" + strconv.Itoa(len(cdtype.Applets))
	for name := range cdtype.Applets {
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

// ListServices forwards action to get the list of active services.
//
func ListServices() (string, error) {
	client, e := dbuscommon.GetClient(srvdbus.SrvObj, srvdbus.SrvPath)
	if e != nil {
		return "", e
	}
	var str string
	e = client.Get("ListServices", []interface{}{&str})
	return str, e
}
