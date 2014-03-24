package main

import (
	"github.com/sqp/godock/libs/log"
	"github.com/sqp/godock/libs/srvdbus"
	"github.com/sqp/godock/services/allapps"

	// "github.com/sqp/godock/gui/config"
)

var cmdService = &Command{
	Run:       runService,
	UsageLine: "service [applet arguments list]",
	Short:     "start service with the dock or an applet",
	Long: `
Service handle the loading of the dock or its own packed applets.

Options:
  dock        Start the dock. This allow to relaunch the dock with its output
              in the same location.
  list        List active applets instances handled by the service.
  stop        Stop the dock.

The service option can also be called with options to start an applet. Those
options are provided by the dock when starting an applet. It only work for 
applets actually packed in this program.

To enable it, use a shell script in place of the applet binary to forward them:
  !/bin/sh 
  cdc service $*
.`,
}

func runService(cmd *Command, args []string) {
	switch {
	case len(args) < 1:

	// case args[0] == "config":
	// config.NewGuiConfig()

	case args[0] == "dock": // Start the service with cairo-dock.
		srvdbus.StartDock()
		service(nil)

	case args[0] == "list": // List active instances.
		if srv, e := srvdbus.GetServer(); srv != nil && !log.Err(e, "List") {
			srv.ListServices()
		}

	case args[0] == "stop": // Stop the dock.
		if srv, e := srvdbus.GetServer(); srv != nil && !log.Err(e, "Stop") {
			log.Info("stop")
			log.Err(srv.StopDock(), "stop")
		}

	case len(args) == 6: // Start applet.
		service(args)

	default:
		log.Info("wrong arguments", toSliceInterface(args)...)
	}
}

// Start Loader with the list of applets and args received for the first applet.
func service(args []string) {
	defer allapps.OnStop()

	load := srvdbus.NewLoader(allapps.List())
	active, e := load.StartServer()
	if !log.Err(e, "StartServer") {
		if active {
			if len(args) > 0 {
				load.StartApplet("", args[0], args[1], args[2], args[3], args[4], args[5])
			}
			load.StartLoop()
		} else {
			if len(args) > 0 {
				// Wether the first program instance will handle successfully the request or
				// not, this isn't our problem anymore, we still must quit.
				listifs := append([]interface{}{""}, toSliceInterface(args)...)
				load.Send("StartApplet", listifs...)
			}
		}
	}
}

func toSliceInterface(list []string) []interface{} {
	listifs := make([]interface{}, len(list))
	for i, arg := range list {
		listifs[i] = arg
	}
	return listifs
}
