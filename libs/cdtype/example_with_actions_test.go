package cdtype_test

import (
	"github.com/sqp/godock/libs/cdtype" // Applet types.
)

//
//-------------------------------------------------------------[ src/demo.go ]--

// Applet data and controlers.
//
type AppletWithAction struct {
	cdtype.AppBase // Extends applet base and dock connection.

	conf *appletConf // Applet configuration data.
}

// NewApplet create a new applet instance.
//
func NewAppletWithAction(base cdtype.AppBase, events *cdtype.Events) cdtype.AppInstance {
	app := &AppletWithAction{AppBase: base}
	app.SetConfig(&app.conf, app.actions()...) // Actions are added here.

	// Events.
	// ...

	return app
}

// Set interval in Init with your config values.
//
func (app *AppletWithAction) Init(def *cdtype.Defaults, confLoaded bool) {
	// Defaults.
	// You can assign shortkeys callbacks manually.
	// This is only needed when the "action" tag is not set in the conf.
	app.conf.ShortkeyOpenThing.Call = func() {
		// ...
	}

	// A second version of the callback can return an error that will be logged.
	// CallE is the only one used if set.
	app.conf.ShortkeyEditThing.CallE = func() error {
		// ...
		return nil
	}

	// Shortkeys should have been found in the conf, listed in def.Shortkeys.
	// They will be added or refreshed at the SetDefaults after Init.
}

//
//-------------------------------------------------------[ action definition ]--

// Define applet actions.
// Actions order in this list must match the order of defined actions numbers.
//
func (app *AppletWithAction) actions() []*cdtype.Action {
	return []*cdtype.Action{
		{
			ID:   ActionNone,
			Menu: cdtype.MenuSeparator,
		}, {
			ID:   ActionOpenThing,
			Name: "Open that thing",
			Icon: "icon-name",
			Call: app.askText,
		}, {
			ID:   ActionEditThing,
			Name: "Edit that thing",
			Icon: "other-icon",
			Call: app.askText,
		},
	}
}

//
//--------------------------------------------------------[ action callbacks ]--

// This is used as action (no arg) to popup a string entry dialog.
//
func (app *AppletWithAction) askText() {
	app.PopupDialog(cdtype.DialogData{
		Message: "Enter some text",
		Buttons: "ok;cancel",
		Widget:  cdtype.DialogWidgetText{},
		Callback: cdtype.DialogCallbackValidString(func(data string) {
			app.Log().Info("user validated with ok or enter", data)
		}),
	})
}

//
//------------------------------------------------------------[ doc and test ]--

func Example_withActions() {
	testApplet(NewAppletWithAction)

	// Output:
	// true
}
