package cdapplet_test

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

type Applet struct {
	cdtype.AppBase                 // Extends applet base and dock connection.
	conf           *appletConf     // Applet configuration data.
	service        *pollingService // The action triggered by the poller.
}

func NewApplet() cdtype.AppInstance {
	app := &Applet{AppBase: cdapplet.New()} // Icon controler and interface to cairo-dock.

	app.service = newService(app)                 // This will depend on your applet polling service.
	poller := app.Poller().Add(app.service.Check) // Create poller and set action callback.

	imgPath := "/usr/share/cairo-dock/icons/icon-movment.png"
	poller.SetPreCheck(func() { app.SetEmblem(imgPath, emblemAction) })
	poller.SetPostCheck(func() { app.SetEmblem("none", emblemAction) })

	return app
}

// Set interval in Init when you have config values.
//
func (app *Applet) Init(loadConf bool) {
	app.LoadConfig(loadConf, &app.conf) // Load config will crash if fail. Expected.

	delay := cdtype.PollerInterval(app.conf.UpdateInterval, defaultInterval)

	app.SetDefaults(cdtype.Defaults{
		PollerInterval: delay,
	})
}

//
//---------------------------------------------------------[ Polling Service ]--

type appletControl interface {
	SetIcon(string) error
}

type pollingService struct {
	app appletControl
}

func newService(app appletControl) *pollingService {
	return &pollingService{app}
}

func (p *pollingService) Check() {}

//
//------------------------------------------------------------[ doc and test ]--

func Example_poller() {
	app := NewApplet()
	fmt.Println(app != nil)

	// Output:
	// true
}

type appletConf struct {
	cdtype.ConfGroupIconBoth `group:"Icon"`
	groupConfiguration       `group:"Configuration"`
}

type groupConfiguration struct {
	UpdateInterval int
}
