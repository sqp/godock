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

// Introspect defines other Dbus actions to provide with introspect.
var IntrospectApplets = `
	<method name="ListServices">
		<arg direction="out" type="s"/>
	</method>
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
	*srvdbus.Loader // Extends the applet service loader to provide its methods on the bus.

	actives  map[string]*activeApp // Active applets.    Key = applet dbus path (/org/cairodock/CairoDock/appletName).
	services cdtype.ListStarter    // Available applets. Key = applet name.
	log      cdtype.Logger
}

// NewManager creates a loader with the given list of applets services.
//
func NewManager(loader *srvdbus.Loader, log cdtype.Logger, services cdtype.ListStarter) *Manager {
	return &Manager{
		Loader:   loader,
		services: services,
		actives:  make(map[string]*activeApp),
		log:      log}
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

	if app == nil {
		load.log.NewErr(name, "failed to start applet")
		return nil
		// return &dbus.Error{Name: "start failed: " + name}
	}

	backend := appdbus.NewWithApp(app, args, h)

	er := backend.ConnectEvents(load.Loader.Conn)
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

	call := client.Object.Call("ListServices", 0)
	if call.Err != nil {
		return "", call.Err
	}
	str := ""
	e = call.Store(&str)
	return str, e
}
