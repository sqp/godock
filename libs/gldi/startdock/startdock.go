// Package startdock wraps all backends and clients to start a dock.
package startdock

import (
	"github.com/sqp/godock/libs/cdglobal"
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/gldi"
	"github.com/sqp/godock/libs/gldi/backendgui"
	"github.com/sqp/godock/libs/gldi/backendmenu"
	"github.com/sqp/godock/libs/gldi/globals"
	"github.com/sqp/godock/libs/gldi/guibridge"
	"github.com/sqp/godock/libs/gldi/maindock"
	"github.com/sqp/godock/libs/gldi/menu"
	"github.com/sqp/godock/libs/gldi/mgrgldi"
	"github.com/sqp/godock/libs/net/websrv"
	"github.com/sqp/godock/libs/ternary"
	"github.com/sqp/godock/libs/text/color"
	"github.com/sqp/godock/libs/text/strhelp"
	"github.com/sqp/godock/services/allapps"

	// loader
	"github.com/sqp/godock/libs/srvdbus"
	"github.com/sqp/godock/libs/srvdbus/dockpath" // hack dock dbus path

	"errors"
	"fmt"
	"net/http/pprof"
)

var (
	// CustomHacks defines developer optional custom settings launched during init.
	CustomHacks = func() {}
)

// Run starts dock routines and locks the main thread with gtk.
//
// It wraps all backends and clients to start a dock :
// Dbus server, applets manager, GUI, menu and gldi.
//
// Dbus service is enabled by default. Mandatory if enabled, to provide remote control.
// This will prevent launching a 2nd instance without the disable Dbus service option.
//
// You can add custom changes, launched before the start, with CustomHacks.
//
// Run returns true if the dock is able to start. This can be done with:
//   maindock.Lock()  // alias for gtk_main.
//   maindock.Clean() // may be better with defer, but cause confused panic messages.
//
func Run(log cdtype.Logger, getSettings func() maindock.DockSettings) bool {
	settings := getSettings()

	// Logger debug state.
	log.SetDebug(settings.Debug)
	maindock.SetLogger(log)

	// Dock init.
	settings.Init()

	// dbus service is mandatory if enabled.
	if !settings.AppletsDisable {
		dbus, e := serviceDbus(log)
		if log.Err(e, "applets service") {
			return false
		}
		// TODO: run a ticking loop for applets when Dbus is disabled.
		appmgr := mgrgldi.Register(allapps.List(settings.Exclude), log)
		dbus.SetManager(appmgr)
		// go appmgr.StartLoop()
		go dbus.StartLoop()
	}

	// HTTP listener for the pprof debug.
	if settings.HTTPPprof {
		websrv.Service.Register("debug/pprof", pprof.Index, log)
		websrv.Service.Start("debug/pprof")
	}

	PrintVersions()

	CustomHacks()

	backendgui.Register(guibridge.New(log))
	backendmenu.Register("dock", menu.BuildMenuContainer, menu.BuildMenuIcon)
	backendmenu.SetLogger(log)

	settings.Prepare()
	settings.Start()

	return true
}

// PrintVersions prints all program and backends versions.
//
func PrintVersions() {
	gtkA, gtkB, gtkC := globals.GtkVersion()

	for _, line := range []struct{ k, v string }{
		{"Custom Dock", cdglobal.AppVersion},
		{"   gldi    ", globals.Version()},
		{"   GTK     ", fmt.Sprintf("%d.%d.%d", gtkA, gtkB, gtkC)},
		{"  OpenGL   ", ternary.String(gldi.GLBackendIsUsed(), "Yes", "No")},
		{"Build date ", cdglobal.BuildDate},
		{" Git Hash  ", cdglobal.GitHash},
	} {
		println(strhelp.Bracket(color.Colored(line.k, color.FgGreen)), line.v)
	}
}

// Start Loader.
//
func serviceDbus(log cdtype.Logger) (*srvdbus.Loader, error) {
	dockpath.DbusPathDock = "/org/cdc/Cdc" // TODO: improve to autodetect.

	loader := srvdbus.NewLoader(log)
	if loader == nil {
		return nil, errors.New("Dbus service failed to start")
	}

	active, e := loader.Connect()
	switch {
	case e != nil:
		return nil, e

	case !active:
		return nil, errors.New("service already active")
	}

	return loader, nil
}
