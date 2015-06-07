// +build !dock

package main

import (
	"github.com/sqp/godock/libs/srvdbus"
	"github.com/sqp/godock/libs/srvdbus/mgrdbus"
	"github.com/sqp/godock/services/allapps"
)

func init() {
	cmdService = &Command{
		Run:       runService,
		UsageLine: "service [applet arguments list]",
		Short:     "list applets services",
		Long: `
Service lists known and active applets.

Options:
  list        List available and active applets instances handled by the service.

The service option is used by the dock to remotely start an applet.
Without arguments, it will display the list of known and active applets.

Otherwise, it requires a lot of arguments needed by the new applet instance it's
supposed to launch, so this won't be directly useful for a user.
It only work for applets actually packed in this program.

The list of applets will change over time with versions and build options.

To enable the applet service for an applet, use a shell script in place of the
applet binary to forward the call:

  !/bin/sh 
  cdc service $* "$(pwd)" &
.`,
	}
}

var gtkStart func()
var gtkStop func()

func runService(cmd *Command, args []string) {
	switch len(args) {

	case 7: // Start applet. Need all arguments AFTER the command name.
		service(args)

	default:
		str, e := mgrdbus.ListServices()
		if !logger.Err(e, "List services") {
			println(str)
		}
	}
}

// Start Loader with the list of applets and args received for the first applet.
//
func service(args []string) {
	loader := srvdbus.NewLoader(logger)
	if loader == nil {
		return
	}

	introspect := srvdbus.Introspect(mgrdbus.IntrospectApplets)

	mgr := mgrdbus.NewManager(loader, logger, allapps.List())
	loader.SetManager(mgr)

	active, e := loader.Start(mgr, introspect)
	if logger.Err(e, "StartServer") {
		return
	}

	if !active { // Someone else is active, forward the start applet and quit.
		// Whether the first program instance will handle successfully the
		// request or not, this isn't our problem anymore, we still must quit.
		mgrdbus.StartApplet(args[0], args[1], args[2], args[3], args[4], args[5], args[6])
		return
	}

	// I am the chosen one. Let's create the first miracle, and keep it alive.
	mgr.StartApplet("", args[0], args[1], args[2], args[3], args[4], args[5], args[6])

	// defer allapps.OnStop()
	if gtkStart != nil && allapps.GtkNeeded() {
		go func() {
			loader.StartLoop()
			logger.Info("cdc stopped")
			gtkStop()
		}()
		gtkStart()

	} else {
		loader.StartLoop()
	}
	loader.Conn.Close()
}
