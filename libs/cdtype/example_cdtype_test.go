package cdtype_test

import (
	"github.com/sqp/godock/libs/cdtype" // Applet types.
)

//
//-------------------------------------------------------------[ src/demo.go ]--

// Applet defines a dock applet.
//
// We start with the applet declaration with at least two mandatory items:
//   -Extend an AppBase.
//   -Provide the config struct (see the config example).
//
type Applet struct {
	cdtype.AppBase             // Extends applet base and dock connection.
	conf           *appletConf // Applet configuration data.
}

// NewApplet create a new applet instance.
//
// It will be trigered remotely by the dock when the applet icon is created.
// Only called once for each running instance.
//
// The goal is to:
//   -Create and fill the applet struct.
//   -Set the config pointer, and maybe some actions.
//   -Register (dock to) applet events (See cdtype.Events for the full list).
//   -Declare and load permanent items required during all the applet lifetime.
//    (those who do not require a config setting, unless it's in a callback).
//
func NewApplet(base cdtype.AppBase, events *cdtype.Events) cdtype.AppInstance {
	app := &Applet{AppBase: base}
	app.SetConfig(&app.conf)

	// Events.

	// Forward click events to the defined command launcher.
	events.OnClick = app.Command().Callback(cmdClickLeft)
	events.OnMiddleClick = app.Command().Callback(cmdClickMiddle)

	// For simple callbacks event, use a closure.
	events.OnDropData = func(data string) {
		app.Log().Info("dropped", data)
	}

	events.OnBuildMenu = func(menu cdtype.Menuer) { // Another closure.
		menu.AddEntry("disabled entry", "icon-name", nil)
		menu.AddSeparator()
		menu.AddEntry("my action", "system-run", func() {
			app.Log().Info("clicked")
		})
	}

	events.OnScroll = app.myonScroll // Or use another function with the same args.

	//
	// Create and set your other permanent items here...

	return app
}

// Applet init
//
// Then we will have to declare the mandatory Init(loadConf bool) method:
//
// Init is called at startup and every time the applet is moved or asked to
// reload by the dock.
// When called, the config struct has already been filled from the config file,
// and some Defaults fields may already be set.
// You may still have to set the poller interval, some commands or declare
// shortkeys callbacks.
//
// Defaults fields not set will be reset to icon defaults (apply even if blank).
//
//   -The config data isn't available before the first Init.
//   -Use PreInit event if you need to act on the previous config before deletion.
//   -cdtype.ConfGroupIconBoth sets those def fields: Label, Icon and Debug.
//   -Don't forget that you may have to clean up some old display or data.
//
func (app *Applet) Init(def *cdtype.Defaults, confLoaded bool) {
	// Set defaults to dock icon: display and controls.
	def.Commands = cdtype.Commands{
		cmdClickLeft: cdtype.NewCommandStd(
			app.conf.LeftAction,
			app.conf.LeftCommand,
			app.conf.LeftClass),

		cmdClickMiddle: cdtype.NewCommandStd(
			app.conf.MiddleAction,
			app.conf.MiddleCommand),
	}

	// Set your other variable settings here...
}

//
//------------------------------------------------------------[ doc and test ]--

// This is an event callback function to be filled with your on scroll actions.
//
func (app *Applet) myonScroll(scrollUp bool) {
	// ...
}

//
//------------------------------------------------------------[ doc and test ]--

func Example_applet() {
	testApplet(NewApplet)

	// Output:
	// true
}
