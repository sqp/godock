// Package allapps declares applets available for the applet loader service.
package allapps

import "github.com/sqp/godock/libs/cdtype"

// Common fields filled by declared applets.
var apps = make(cdtype.ListStarter)
var needgtk bool // true if an applet has some GTK dependency.

// AddService is used to declare a service to the loader.
func AddService(name string, app cdtype.AppStarter) {
	apps[name] = app
}

// List returns the list of declared applets.
//
func List(exclude ...string) cdtype.ListStarter {
	list := make(cdtype.ListStarter)
	for name, instancer := range apps {
		drop := false
		for _, test := range exclude {
			if test == name {
				drop = true
			}
		}
		if !drop {
			list[name] = instancer
		}
	}
	return list
}

// AddGtkNeeded allow an applet to declare its gtk dependency.
// If used, the gtk main loop should lock the main thread to prevent later problems.
//
func AddGtkNeeded() {
	needgtk = true
}

// GtkNeeded returns true if an applet requires gtk.
func GtkNeeded() bool {
	return needgtk
}
