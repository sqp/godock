package main

import (
	"github.com/sqp/godock/libs/dbus"
	"github.com/sqp/godock/libs/log"
	"github.com/sqp/godock/libs/srvdbus"
)

var cmdRestart = &Command{
	Run:       runRestart,
	UsageLine: "restart [appletname...]",
	Short:     "restart the dock or one or more applet",
	Long: `
Restart restarts the Cairo-Dock instance or external applets.

Without any argument, all your dock will be restarted.

If one or more applet name is provided, they will be restarted.

Note that only external applets will benefit from a simple applet restart if you modified the code.`,
}

func runRestart(cmd *Command, args []string) {

	switch {
	case len(args) < 1: // Restart dock.
		if srv, e := srvdbus.GetServer(); srv != nil && e == nil {
			srv.RestartDock()
		} else {
			//go
			log.Err(dbus.DockQuit(), "DockQuit")
			srvdbus.StartDock()
		}

	default: // Restart applet(s).
		for _, name := range args {
			log.Err(dbus.AppletRemove(name+".conf"), "AppletRemove")
			log.Err(dbus.AppletAdd(name), "AppletAdd")
		}
	}
}
