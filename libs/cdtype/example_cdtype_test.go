package cdtype_test

import (
	"github.com/sqp/godock/libs/cdapplet" // Applet base.
	"github.com/sqp/godock/libs/cdtype"   // Applet types.

	"fmt"
)

//
//-------------------------------------------------------------[ src/demo.go ]--

// Applet data and controlers.
//
// We start with the applet declaration with at least two mandatory items:
//   -Extend an AppBase.
//   -Provide the config struct (see later for the config).
//
type Applet struct {
	cdtype.AppBase             // Extends applet base and dock connection.
	conf           *appletConf // Applet configuration data.
}

// Applet creation.
//
// As this will only be called once for each running instance, this is the place
// to declare all your permanent data that will be required during all the
// applet lifetime.
//
// This will be called remotely by the dock when the applet icon is created.
//
func NewApplet() cdtype.AppInstance {
	app := &Applet{AppBase: cdapplet.New()} // Icon controler and interface to cairo-dock.

	// Create and set your permanent items here...

	return app
}

// Applet init
//
// Then we will have to declare the mandatory Init(loadConf bool) method:
//
// Init is called at startup and every time the applet is asked to reload by the dock.
// First, it reload the config file if asked to, and then it should (re)initialise
// everything needed. Don't forget that you may have to clean some things up.
// SetDefaults will help you as it reset everything it handles, even if not set
// (default blank value is used).
//
// Load user configuration if needed and initialise applet.
//
func (app *Applet) Init(loadConf bool) {
	app.LoadConfig(loadConf, &app.conf) // Load config will crash if fail. Expected.

	// Set defaults to dock icon: display and controls.
	app.SetDefaults(cdtype.Defaults{
		Label: app.conf.Name,
		Icon:  app.conf.Icon,
		Commands: cdtype.Commands{
			0: cdtype.NewCommandStd(app.conf.LeftAction, app.conf.LeftCommand, app.conf.LeftClass),
			1: cdtype.NewCommandStd(app.conf.MiddleAction, app.conf.MiddleCommand)},
		Debug: app.conf.Debug,
	})

	// Set your other variable settings here...
}

// Applet events
//
// DefineEvents is called by the backend only once at startup.
// Its goal is to get a common place for events callback configuration.
// See cdtype.Events for the full list with arguments.
//
// Define events is now optional as you can also provide your events callbacks
// as methods of your applet, with the same name and arguments.
//
func (app *Applet) DefineEvents(events cdtype.Events) {
	// To forward click events to the defined command launcher.
	// It would be better to use consts instead of 0 and 1 like in this example.

	events.OnClick = app.Command().CallbackInt(0)
	events.OnMiddleClick = app.Command().CallbackNoArg(1)

	events.OnDropData = func(data string) { // For simple callbacks event, use a closure.
		app.Log().Info("dropped", data)
	}

	events.OnBuildMenu = func(menu cdtype.Menuer) { // Another closure.
		menu.AddEntry("disabled entry", "", nil)
		menu.AddSeparator()
		menu.AddEntry("my action", "system-run", func() { app.Log().Info("clicked") })
	}

	events.OnScroll = app.myonScroll // Or use another function with the same args.
}

// This is an event callback function to be filled with your on scroll actions.
//
func (app *Applet) myonScroll(scrollUp bool) {
	// ...
}

// Here, the same events is declared directly as methods of the applet.
//
// There is no way to document it directly, and no safety check at compilation,
// as the hook service just use matching methods.
//
// Note that both methods are available and enabled at the same time, and you may
// be able to use them both in the same applet, but this isn't really tested.
// (and its not sure this will remain. we'll see what are the needs).
//
func (app *Applet) OnBuildMenu(menu cdtype.Menuer) {
	// ...
}

//
//------------------------------------------------------------[ doc and test ]--

func Example_applet() {
	app := NewApplet()
	fmt.Println(app != nil)

	// Output:
	// true
}
