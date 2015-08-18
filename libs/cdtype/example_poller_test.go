package cdtype_test

import (
	"github.com/sqp/godock/libs/cdapplet" // Applet base.
	"github.com/sqp/godock/libs/cdtype"   // Applet types.

	"fmt"
)

//
//-----------------------------------------------------------[ src/config.go ]--

const defaultInterval = 60                 // Interval in seconds. This set a default to one minute.
const emblemAction = cdtype.EmblemTopRight // Emblem position for polling activity.

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
func NewPollerApplet() cdtype.AppInstance {
	app := &PollerApplet{AppBase: cdapplet.New()} // Icon controler and interface to cairo-dock.

	app.service = newService(app)                 // This will depend on your applet polling service.
	poller := app.Poller().Add(app.service.Check) // Create poller and set action callback.

	// The poller can trigger other actions before and after its call.
	//
	// This can be used for example to display an activity emblem on the icon.
	//
	imgPath := "/usr/share/cairo-dock/icons/icon-movment.png"
	poller.SetPreCheck(func() { app.SetEmblem(imgPath, emblemAction) })
	poller.SetPostCheck(func() { app.SetEmblem("none", emblemAction) })

	// The example use constant paths, but you can get a path relative to your applet:
	// imgPath := app.FileLocation("icon", "iconname.png")

	// Or you can define custom calls.

	// app.Poller().SetPreCheck(onStarted)
	// app.Poller().SetPostCheck(onFinished)

	// Create and set your permanent items here...

	return app
}

// Set interval in Init when you have config values.
//
func (app *PollerApplet) Init(loadConf bool) {
	app.LoadConfig(loadConf, &app.conf) // Load config will crash if fail. Expected.

	// Ensure we have a valid time for the poller with a default constant.
	delay := cdtype.PollerInterval(app.conf.UpdateInterval, defaultInterval)

	// Set the poller interval directly.
	app.Poller().SetInterval(delay)

	// or use the defaults (it's better as SetDefaults is a must).

	app.SetDefaults(cdtype.Defaults{
		PollerInterval: delay,

		// Set your other defaults here...
	})

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
}

//
//------------------------------------------------------------[ doc and test ]--

func Example_poller() {
	app := NewApplet()
	fmt.Println(app != nil)

	// Output:
	// true
}
