// Package cdtype defines main types and constants for Cairo-Dock applets.
/*

External applets are small programs spawned by the dock to manage dedicated icons.
Once started, the program can use his icon to display informations and receive
user interactions like mouse events and keyboard shortcuts.

This API now covers all the dock external features, and most common applets
needs, like the automatic config loading, a menu builder, with a really short
and simple code.

I tried to make this applet API as straightforward and simple to use as I could
so feel free to test it and post your comments on the forum.


Applet lifecycle

The applet loading is handled by the dock, then it's up the applet to act upon
events received to work as expected and quit when needed. Be aware that you
don't control your applet, the dock does. Most of your work is to define config
keys, connect events and make them useful with your config options.

Automatic applet creation steps, and their effect, when the dock icon is enabled:

  * Create the applet object with its dock backend (DBus or gldi) and events:
      app := NewApplet(base, events)
  * Connect the applet backend to cairo-dock to activate events callbacks.
  * Load config, fill the data struct and prepare a Defaults object for Init.
  * Initialize applet with the filled Defaults:
      app.Init(def, true).
  * Defaults are applied, restoring icons defaults to missing fields.
  * Start and run the polling loop if needed with a first instant check.

At this point, the applet is alive and running

Init (and PreInit) will be triggered again (with a load config or not) on some
more or less mysterious occasions, from an applet or dock config save, to a dock
move, or even a system package install.

While alive the applet can:

  * Act on icon to display informations.
  * Receive user/dock events.
  * Regular polling if set.

Stops when End event is received from the dock to close the applet.


Register applet in the dock

To register our "demo" applet in the dock, we have to link it as "demo" in the
external applets dir.

 ~/.config/cairo-dock/third-party/demo

 This can be done easily once you have configured the SOURCE location in the
 Makefile with the "make link" command.

or use the external system directory

  /usr/share/cairo-dock/plug-ins/Dbus/third-party/


Notes

A lot more tips are to be found for each method, but here are a few to help you
start:

  *The applet activation requires some arguments provided by the dock with the
    command line options. They will be parsed and used by the backend before
    the creation.

  *Only a few applet methods are available at creation (before Init) like Log
    and file access (Name, FileLocation, FileDataDir...). Others like icon
    actions can be set in callbacks that will only be used later when the
    backend is running.

  *The current directory by default is the applet directory, but to be safe
    don't rely on it and use chdir or absolute paths for all your file
    operations.

  *If you want to publish the package or ask for help, it would be easier if the
    applet is located in a "go gettable" directory (just use the same location
    in GOPATH as you have in your public repo, like this API).

  *The default Reload event will launch an Init call and restart the poller
    if any (but this can be overridden, see cdapplet.SetEvents).

Backends

Applets can use two different backends to interact with the dock:

gldi: This require the applet to be compiled and launched with the rework
version of the dock. For brave testers, but can have great rewards as this
unlock all dock and applets capabilities.

DBus: As external applets, they will be limited to basic dock features but
are easy to make and deploy.

    We have two loaders for the DBus backend:

    *StandAlone: Great to work on the applet as it can be reloaded easily.
    This is the default mode for most external applets but can take some disk
    space if you install a lot of applets with big dependencies.

    *Applet service: Packing multiple applets in the same executable reduce CPU,
    memory and disk usage when you have some of them enabled, but it leads to a
    binary choice (the big pack or nothing).
    see the launcher forwarder replacing applet binary: cmd/cdc/data/tocdc


Service registration

When in StandAlone mode, we need to create the binary in the applet data dir.
We're using a minimal main package there to start as external applet.

  func main() { appdbus.StandAlone(Mem.NewApplet) }

When not in StandAlone mode, we need to register the applet with a go init.
This allow the different backends to start the applet on demand.

  func init() { cdtype.Applets.Register("demo", NewApplet) }

Services are enabled with build tags in the services/allapps package.
They are just declared to load them and trigger the registration.
The build tag 'all' is used to enable all known services at once.


Evolutions

All the base actions and callbacks are provided by Cairo-Dock, and are in stable
state for years, so everything provided will remain in almost the same format.

We now reached a point where a few evolutions will be made to extend some
methods to get more benefit from the gldi backend, with a fallback as best as
possible on the DBus backend (like shortkeys registered as expected on gldi).
The goal here will be to reduce the gap between internal C applets and go
applets as we have numerous options still unavailable (ex on gauge/graph).
One of those method will set renderers label formating functions for vertical
and horizontal dock modes.

We could still introduce minor API changes to fix some issues you may encounter
while using it, but this should be mostly cosmetic to improve readability when
needed.


Real applets code and data

You can see real applets code and configuration in applets and services subdirs.
	Code repository (real applets)  http://github.com/sqp/godock/tree/master/services
	Data repository (config files)  http://github.com/sqp/godock/tree/master/applets

Small examples

For easier maintenance, sources for the applet will be separated from the data,
So we're creating 2 directories for each applet.
You can use a src subdir to keep them grouped.

In the examples we will use demo and demo/src.
  The applet management file:     demo/src/demo.go
  The applet configuration file:  demo/src/config.go


*/
package cdtype
