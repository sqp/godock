// Package dockpath defines paths used by the main dock dbus service.
package dockpath

import "github.com/godbus/dbus"

// Dbus dock paths.
//
const (
	DbusObject             = "org.cairodock.CairoDock"
	DbusInterfaceDock      = "org.cairodock.CairoDock"
	DbusInterfaceApplet    = "org.cairodock.CairoDock.applet"
	DbusInterfaceSubapplet = "org.cairodock.CairoDock.subapplet"
)

// DbusPathDock is the Dbus path to the dock. It depends on the name the dock was started with.
var DbusPathDock dbus.ObjectPath = "/org/cairodock/CairoDock"
