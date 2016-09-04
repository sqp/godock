package cdtype_test

import (
	"github.com/sqp/godock/libs/cdtype" // Applet types.
)

//
//-------------------------------------------------------------[ src/demo.go ]--

// The applet main loop can handle a polling service to help you update data on a
// regular basis. The poller will be agnostic of what's happening, he will just
// trigger the provided call at the end of timer or in case of manual reset with
// app.Poller().Restart().

// Define an applet with a polling service.
//
type PollerApplet struct {
	cdtype.AppBase                 // Extends applet base and dock connection.
	conf           *appletConf     // Applet configuration data.
	service        *pollingService // The action triggered by the poller.
}

// Create your poller with the applet.
//
func NewPollerApplet(base cdtype.AppBase, events *cdtype.Events) cdtype.AppInstance {
	app := &PollerApplet{AppBase: base}
	app.SetConfig(&app.conf)

	// Events.
	// ...

	// Polling service.
	app.service = newService(app)                 // This will depend on your applet polling service.
	poller := app.Poller().Add(app.service.Check) // Create poller and set action callback.

	// The poller can trigger other actions before and after the real call.
	//
	// This can be used for example to display an activity emblem on the icon.
	//
	imgPath := "/usr/share/cairo-dock/icons/icon-bubble.png"
	poller.SetPreCheck(func() { app.SetEmblem(imgPath, emblemAction) })
	poller.SetPostCheck(func() { app.SetEmblem("none", emblemAction) })

	// The example use constant paths, but you can get a path relative to your applet:
	// imgPath := app.FileLocation("icon", "iconname.png")

	// Or you can define custom calls.

	// app.Poller().SetPreCheck(app.service.onStarted)
	// app.Poller().SetPostCheck(app.service.onFinished)

	// Create and set your other permanent items here...

	return app
}

// Init is called after load config. We'll set the poller interval.
//
func (app *PollerApplet) Init(def *cdtype.Defaults, confLoaded bool) {
	// In Init, we'll use the Defaults as it's already provided.
	def.PollerInterval = app.conf.UpdateInterval.Value()

	// But you can also set a poller interval directly anytime.
	app.Poller().SetInterval(app.conf.UpdateInterval.Value())

	// Set your other defaults here...

	// Set your other variable settings here...
}

//
//---------------------------------------------------------[ Polling Service ]--

// appletControl defines methods needed on the applet by the polling service.
// It's better to declare an interface to restrict the polling usage and keep
// the rest of this service as agnostic of the others as possible.
//
type appletControl interface {
	SetIcon(string) error
	Log() cdtype.Logger
}

// pollingService defines a polling service skelton.
//
type pollingService struct {
	app appletControl // polling services often need to interact with the applet.
}

// newService creates a polling service skelton.
//
func newService(app appletControl) *pollingService {
	return &pollingService{app}
}

// Check is the real polling action
//
// It's doing its routinely task, display or forward its result, and that's all.
//
func (p *pollingService) Check() {

	// This is where the polling service action takes place.
	p.app.Log().Info("checked")
}

//
//------------------------------------------------------------[ doc and test ]--

func Example_poller() {
	testApplet(NewPollerApplet)

	// Output:
	// true
}
