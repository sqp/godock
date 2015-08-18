// Package cdtype defines main types and constants for Cairo-Dock applets.
/*

External applets are small programs spawned by the dock to manage dedicated icons.
Once started, the program can use his icon to display informations and receive
user interactions like mouse events and keyboard shortcuts.

This API now covers all the dock external features, so it shouldn't change much.
All the base actions and callbacks are provided by Cairo-Dock, and are in stable
state, so everything provided will remain in almost the same format.
Only the name, args, of way to access some methods may evolve in the Golang
implementation, but only to better suits the needs and provide a great
Cairo-Dock Golang API.
Those changes could be made to fix some issues you may encounter while using it,
so feel free to test it and post your comments on the forum.



Applet lifecycle

The applet loading is handled by the dock, then it's up the applet to act upon
events received to work as expected and quit when needed. Be aware that you
don't control your applet, the dock does. Most of your work is to connect events
and make them useful.
The default Reload event will launch an Init call with the given loadConfig value
and restart the poller if any (can be overriden).

	Applet is created and started by the dock when its icon is enabled.
	  DefineEvents().
	  connect to backend.
	  Init(loadConfigTrue)
	  start poller with instant check (if any).
	Applet is alive
	  can act on icon to display informations.
	  receives and answers to connected events.
	  event reload will trigger other Init event with or without loadConfig.
	  regular poller ticking (if any).
	Applet receives event quit when the icon is removed or the dock closed.
	  triggers applet End event before releasing the main loop.



Register applet in the dock

To register our "demo" applet in the dock, we have to link it as "demo" in the external applets dir.
 ~/.config/cairo-dock/third-party/demo

 This can be done easily once you have configured the SOURCE location in the
 Makefile with the "make link" command.

or use the external system directory

  /usr/share/cairo-dock/plug-ins/Dbus/third-party/



Applet sources

For easier maintenance, sources for the applet will be separated from the data,
So we're creating 2 directories for each applet.
You can use a src subdir to keep them grouped.

In the examples we will use demo and demo/src.
  The applet management file:     demo/src/demo.go
  The applet configuration file:  demo/src/config.go



Code examples

You can see real applets code and configuration in applets and services subdirs.
	Code repository (real applets)  http://github.com/sqp/godock/tree/master/services
	Data repository (config files)  http://github.com/sqp/godock/tree/master/applets



Notes

A lot more tips are to be found for each method, but here are a few to help you
start:
 *The dock will launch the applet command (applet binary or script) with
   a few options to help it know some dock settings. Those args will only be
   parsed and stored in the public fields during StartApplet, so they won't be
   available at applet creation (Name, FileLocation, FileDataDir...).
   You can only start using them in the first Init call. Only the logger is safe
   to use before that.
 *The current directory by default is the applet directory, but to be safe don't
   rely on it and use chdir or absolute paths for all your file operations.
 *If you want to publish the package or ask for help, it would be easier if the
   applet is located in a "go gettable" directory (just use the same location in
   GOPATH as you have in your public repo, like this API).
 *Applets, once stable, can be packed in an applet manager, currently handled by
   the cdc command. The manager can currently run applets with the Dbus client
   or with its own internal go applets service (with an active gldi dock).
   When used with Dbus, the start applet will have to be forwarded to the cdc
   loader. For this, we're creating a tocdc file and link it to the applet name
	ln -s tocdc demo

   You will just have to remove the link and rebuild the binary to switch back
   to a single instance mode. Or delete the binary and relink the tocdc file to
   reuse the loader.

	tocdc:
		#!/bin/sh
		cdc service $* "$(pwd)" &

*/
package cdtype
