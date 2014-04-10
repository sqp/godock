package main

import (
	"github.com/sqp/godock/libs/dbus"
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
	if len(args) == 0 { // Restart dock.
		if clientSend(restartDock, args) != nil { // Try to forward to an active instance.
			logger.Err(srvdbus.RestartDock(), "restart dock") // Nobody else wants to, I'll do it myself!
		}
		return
	}

	for _, name := range args { // Restart applet(s).
		logger.Err(dbus.AppletRemove(name+".conf"), "AppletRemove")
		logger.Err(dbus.AppletAdd(name), "AppletAdd")
	}
}

func restartDock(srv *srvdbus.Client, args []string) error {
	return srv.RestartDock()
}
