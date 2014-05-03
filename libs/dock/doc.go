/*
Package dock is the Cairo-Dock applet manager, using the DBus connection.

External applets are small programs spawned by the dock to manage dedicated icons.
Once started, the program can use his icon to display informations and receive
user interactions like mouse events and keyboard shortcuts.

This API is still fairly new, so it will evolve, but not that much. You can
already consider starting an applet using it. All the base actions and callbacks
are provided by Cairo-Dock, and are in stable state, so everything provided will
remain in almost the same format. Only the name, args, of way to access methods
may evolve in the Golang implementation, but only to better suits the needs and
provide a great Cairo-Dock Golang API. Those changes could be made to fix some
issues you may encounter while using it, so feel free to use it and post your
comments on the forum.

Applet sources

Sources for the applet will better be separated from applet data, for easier
maintenance. So we will start by creating 2 directories.
You can use a src subdir to keep them grouped.
In this example we will use demo and demo/src.

The applet management file will be demo/src/demo.go


Applet lifecycle

The applet loading is handled by the dock, then it's up the applet to act upon
events received to work as expected and quit when needed. Be aware that you
don't control your applet, the dock does. Most of your work is to connect events
and make them useful.
The default Reload event will launch an Init call with the given loadConfig value
and restart the poller if any (can be overriden).

	Applet is created by the dock when its icon is enabled.
	Applet is started. Triggers the first Init event with loadConfig=true
	Applet is alive
	  can act on icon to display informations.
	  receives and answers to connected events.
	  event reload will trigger other Init event with or without loadConfig.
	Applet receives event quit when the icon is removed or the dock closed.
	  Triggers End event before releasing the main loop.

Applet creation

First, we need to declare and create our applet. For this we will extend a
CDApplet object and declare our config struct (see later for the config).
As this will only be called once, this is the place to declare all your permanent
data that will be required during all the applet lifetime.

	package demo

	import "github.com/sqp/godock/libs/dock" // Connection to cairo-dock.

	// Applet data and controlers.
	//
	type Applet struct {
		*dock.CDApplet
		conf *appletConf
	}

	// NewApplet creates a new applet instance.
	//
	func NewApplet() dock.AppletInstance {
		app := &Applet{CDApplet: dock.NewCDApplet()} // Icon controler and interface to cairo-dock.
		return app
	}

Applet events

Then we will have to declare the two mandatory methods Init(bool) and DefineEvents():

Init is called at startup and every time the applet is asked to reload by the dock.
First, it reload the config file if asked to, and then it should (re)initialise
everything needed. Don't forget that you may have to clean some things up.
SetDefaults will help you as it reset everything it handles, even if not set
(default blank value is used).

	// Load user configuration if needed and initialise applet.
	//
	func (app *Applet) Init(loadConf bool) {
		app.LoadConfig(loadConf, &app.conf) // Load config will crash if fail. Expected.

		// Set defaults to dock icon: display and controls.
		app.SetDefaults(dock.Defaults{
			Label:          ternary.String(app.conf.Name != "", app.conf.Name, app.AppletName),
			Commands: dock.Commands{
				"left":   dock.NewCommandStd(app.conf.LeftAction, app.conf.LeftCommand, app.conf.LeftClass),
				"middle": dock.NewCommandStd(app.conf.MiddleAction, app.conf.MiddleCommand)},
			Debug: app.conf.Debug})
	}

DefineEvents is called only once at startup. Its goal is to get a common place
for events callback configuration.
See cdtype.Events and cdtype.SubEvents for the full list with arguments.


	// Define applet events callbacks.
	//
	func (app *Applet) DefineEvents() {
		app.Events.OnClick = app.LaunchFunc("left")          // If you want to forward a click event to the optional command launcher.
		app.Events.OnMiddleClick = app.LaunchFunc("middle")

		app.Events.OnDropData = func(data string) { // For simple callbacks event, use a closure.
			app.Log.Info("dropped", data)
		}

		app.Events.OnBuildMenu = func() {
			menu := []string{"", "ok"} // First entry is a separator.
			app.PopulateMenu(menu...)
		}

		app.Events.OnMenuSelect = app.clickedMenu // If you have more work to do, use a dedicated func.
	}

Poller

The applet main loop can handle a polling service to help you update data on a
regular basis. The poller will be agnostic of what's happening, he will just
trigger the provided call at the end of timer or in case of manual reset with
app.Poller().Restart().

Create your poller with the applet.
	func NewApplet() dock.AppletInstance {
		...
		app.service = NewService(app)    // This will depend on your applet polling service.
		app.AddPoller(app.service.Check) // Create poller and set action callback.
		...
	}

Set interval in Init when you have config values.
	app.Poller().SetInterval(app.conf.UpdateInterval)
or
	app.SetDefaults(dock.Defaults{
		...
		PollerInterval: app.conf.UpdateInterval,
		...
		})


Config

The configuration loading can be fully handled by LoadConfig if the config
struct is created accordingly.

Main conf struct: one nested struct for each config page with its name refered
to in the "group" tag.

Sub conf structs: one field for each setting you need with the keys and types
refering to those in your .conf file loaded by the dock. As fields are filled by
reflection, they must be public and start with an uppercase letter.

If the name of you field is the same as in the config file, just declare it.

If you have different names, or need to refer to a lowercase field in the file,
you must add a "conf" tag with the name of the field to match in the file.

The applet configuration file will be demo/src/config.go

	type appletConf struct {
		groupIcon                       `group:"Icon"`
		groupConfiguration              `group:"Configuration"`
	}

	type groupIcon struct {
		Name             string         `conf:"name"`
	}

	type groupConfiguration struct {
		GaugeName        string
		UpdateInterval   int
		Devices          []string
	}

Documentation about the dock config file:
http://www.glx-dock.org/ww_page.php?p=Documentation&lang=en#27-Building%20the%20Applet%27s%20Configuration%20Window



Applet data files

To register our "demo" applet in the dock, we have to link it as "demo" in the external applets dir.
 ~/.config/cairo-dock/third-party/demo
 This can be done easily once you have configured the SOURCE location in the
 Makefile with the "make link" command.

Files needed in the "demo" directory:
   demo            the executable binary or script, without extension (use the shebang on the first line).
   demo.conf       the default config file (see above for the options syntax).
   auto-load.conf  the file describing our applet.
   icon            the default icon of the applet (optional, without extension, can be any image format).
   preview         an image preview of the applet (optional, without extension, can be any image format).
   applet.go       source to load and start the applet as standalone.
   Makefile        shortcuts to build and link the applet.

auto-load.conf
	[Register]

	# Author of the applet
	author = AuthorName

	# A short description of the applet and how to use it.
	description = This is the description of the applet\nIt can be on several lines.

	# Category of the applet : 2 = files, 3 = internet, 4 = Desktop, 5 = accessory, 6 = system, 7 = fun
	category = 5

	# Version of the applet; change it everytime you change something in the config file. Don't forget to update the version both in this file and in the config file.
	version = 0.0.1

	# Default icon to use if no icon has been defined by the user. If not specified, or if the file is not found, the "icon" file will be used.
	icon =

	# Whether the applet will act as a launcher or not (like Pidgin or Transmission)
	act as launcher = false

Makefile

	TARGET=demo
	SOURCE=github.com/sqp/godock/applets

	# Default is standard build for current arch.

	%: build

	build:
		go build -o $(TARGET) $(SOURCE)/$(TARGET)

	# make a link to the user external directory for easy install.
	link:
		ln -s $(GOPATH)/src/$(SOURCE)/$(TARGET) $(HOME)/.config/cairo-dock/third-party/$(TARGET)

applet.go
	package main

	import (
		"demo/src" // Package with your applet source
		"github.com/sqp/godock/libs/dock"     // Connection to cairo-dock.
	)

	func main() {
		dock.StartApplet(demo.NewApplet())
	}


More documentation
	DBus methods to interact with the dock   http://godoc.org/github.com/sqp/godock/libs/appdbus
	Common types and applet events           http://godoc.org/github.com/sqp/godock/libs/cdtype
	Applets code repository (real examples)  http://github.com/sqp/godock/tree/master/services
	Applets data repository (config files)   http://github.com/sqp/godock/tree/master/applets




Notes:
 *The dock will launch the applet command (applet binary or script) with
   a few options to help him know some dock settings. Those args will only be
   parsed and stored in the public fields during StartApplet, so they won't be
   available at applet creation.
   You can only start using them at the first Init call.
     dock fields: AppletName, ConfFile, RootDataDir, ParentAppName, ShareDataDir
     and related methods like FileLocation
 *The current directory by default is the applet directory, but to be safe don't
   rely on it and use chdir or absolute paths for all your file operations.
 *If you want to publish the package or ask for help, it would be easier if the
   applet is located in a "go gettable" directory (just use the same location in
   GOPATH as you have in your public repo, like this API).
 *Applets, once stable, can be packed in an applet manager, currently handled by
   the cdc command. To forward the applet start to the cdc loader, create a tocdc
   file and link it to the applet name (ln -s tocdc demo).
   You will just have to remove the link and rebuild the binary to switch back
   to a single instance mode. Or delete the binary and relink to reuse the loader.

	tocdc:
		#!/bin/sh
		cdc service $* &




*/
package dock
