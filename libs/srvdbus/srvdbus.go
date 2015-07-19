// Package srvdbus provides a Dbus service (and client) with dock applets management.
package srvdbus

import (
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"

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
		<method name="Upload">
			<arg direction="in" type="s"/>
		</method>
		<method name="Debug">
			<arg direction="in" type="s"/>
			<arg direction="in" type="b"/>
		</method>
		` + methods + `
	</interface>
` + introspect.IntrospectDataString + `
</node>`
}

// AppService defines common applets service actions to remotely interact with applets.
//
type AppService interface {
	CountActive() int
	GetApplets(name string) (list []cdtype.AppInstance)
	Tick()
}

// MgrDbus defines actions needed by the Dbus grouped applets manager.
//
type MgrDbus interface {
	IsActive(path string) bool
	OnSignal(path string, s *dbus.Signal) bool
}

// Loader is a multi applet manager.
//
type Loader struct {
	*dbuscommon.Server            // Dbus connection.
	apps               AppService // Applet actions (debug, upload).
	mgr                MgrDbus    // Applet activity forwarding (signals).
}

// NewLoader creates a loader with the given list of applets services.
//
func NewLoader(log cdtype.Logger) *Loader {
	srv := dbuscommon.NewServer(SrvObj, SrvPath, log)
	if srv == nil {
		return nil
	}
	return &Loader{Server: srv}
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

// StartLoop handle applets until there's none of them alive.
//
func (load *Loader) StartLoop() {
	defer load.Conn.ReleaseName(SrvObj)
	defer load.Log.Debug("Dbus service stopped")
	load.Log.Debug("Dbus service started")

	var waiter <-chan time.Time
	if load.apps != nil {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		waiter = ticker.C
	}

	for { // Main loop.

		select { // Wait for events, until the End signal is received from the dock.

		case s := <-load.Events: // Listen to DBus events.
			if load.dispatchDbusSignal(s) { // true if signal was Stop.

				// Keep service alive if we still manage some applets.
				if load.apps.CountActive() == 0 { // That's all folks. We're closing.
					return
				}
			}

		case <-waiter: // Tick every second to update pollers counters and launch actions.
			load.apps.Tick()
		}
	}
}

// Forward the Dbus signal to local manager or applet
//
func (load *Loader) dispatchDbusSignal(s *dbus.Signal) bool {
	path := strings.TrimSuffix(string(s.Path), "/sub_icons")

	switch {
	case s.Name == "org.freedesktop.DBus.NameAcquired": // Service started confirmed.

	case load.mgr != nil && load.mgr.IsActive(path): // Signal to applet.
		return load.mgr.OnSignal(path, s)

	default:
		load.Log.Info("unknown signal", s)

	}
	return false

}

//
//----------------------------------------------------------[ DBUS CALLBACKS ]--

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
	if len(uncasts) == 0 {
		return nil
	}
	app := uncasts[0].(uploader) // Send it to the first found. Should be safe for now, we can launch only one.
	app.Upload(data)

	return nil
}

// Debug change the debug state of an active applet.
//
func (load *Loader) Debug(applet string, state bool) *dbus.Error {
	if load.apps == nil {
		return nil
	}

	found := false
	for _, app := range load.apps.GetApplets(applet) {
		app.Log().SetDebug(state)
		found = true
	}
	if found {
		load.Log.Info("set debug", applet, state)
	} else {
		load.Log.NewWarn("applet not found = "+applet, "set debug")
	}

	return nil
}

//
//------------------------------------------------------[ DBUS SEND COMMANDS ]--

// Debug forwards action set debug to a remote applet.
//
func Debug(applet string, state bool) error {
	client, e := dbuscommon.GetClient(SrvObj, SrvPath)
	if e != nil {
		return e
	}
	return client.Call("Debug", applet, state)
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
