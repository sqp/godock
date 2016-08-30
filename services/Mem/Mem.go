// Package Mem is a memory monitoring applet for Cairo-Dock.
package Mem

import (
	"github.com/cloudfoundry/gosigar" // System informations.

	"github.com/sqp/godock/libs/cdtype" // Applet types.
	"github.com/sqp/godock/libs/sysinfo"
)

func init() { cdtype.Applets.Register("Mem", NewApplet) }

// Applet defines a dock applet.
//
type Applet struct {
	cdtype.AppBase // Applet base and dock connection.

	conf    *appletConf
	service sysinfo.RenderPercent
}

// NewApplet creates a new applet instance.
//
func NewApplet(base cdtype.AppBase, events *cdtype.Events) cdtype.AppInstance {
	app := &Applet{AppBase: base}
	app.SetConfig(&app.conf)

	// Events.
	events.OnClick = app.Command().Callback(cmdLeft) // Left and middle click: launch the configured action.
	events.OnMiddleClick = app.Command().Callback(cmdMiddle)
	events.OnBuildMenu = func(menu cdtype.Menuer) {
		if app.conf.LeftAction > 0 && app.conf.LeftCommand != "" {
			menu.AddEntry("Action left click", "system-run", app.Command().Callback(cmdLeft))
		}
		if app.conf.MiddleAction > 0 && app.conf.MiddleCommand != "" {
			menu.AddEntry("Action middle click", "system-run", app.Command().Callback(cmdMiddle))
		}
	}

	// Memory service.
	app.Poller().Add(app.GetMemActivity)

	app.service.App = app
	app.service.Texts = map[cdtype.InfoPosition]sysinfo.RenderOne{
		cdtype.InfoNone:    {},
		cdtype.InfoOnIcon:  {Sep: "\n", ShowPre: false},
		cdtype.InfoOnLabel: {Sep: " - ", ShowPre: true},
	}
	return app
}

// Init load user configuration if needed and initialise applet.
//
func (app *Applet) Init(def *cdtype.Defaults, confLoaded bool) {
	// Settings for service.
	app.service.Settings(app.conf.DisplayText, app.conf.DisplayValues, app.conf.GraphType, app.conf.GaugeName)
	app.service.SetSize(countTrue(app.conf.ShowRAM, app.conf.ShowSwap))

	// Defaults.
	def.PollerInterval = app.conf.UpdateDelay.Value()
	def.Commands = cdtype.Commands{
		cmdLeft:   cdtype.NewCommandStd(app.conf.LeftAction, app.conf.LeftCommand, app.conf.LeftClass),
		cmdMiddle: cdtype.NewCommandStd(app.conf.MiddleAction, app.conf.MiddleCommand),
	}
}

//
//---------------------------------------------------------------------[ MEM ]--

// GetMemActivity displays current memory usage.
//
func (app *Applet) GetMemActivity() {
	app.service.Clear()

	if app.conf.ShowRAM {
		mem := sigar.Mem{}
		mem.Get()

		app.service.Append("RAM", float64(mem.ActualUsed)/float64(mem.Total))
	}

	if app.conf.ShowSwap {
		swap := sigar.Swap{}
		swap.Get()

		app.service.Append("Swap", float64(swap.Used)/float64(swap.Total))
	}

	app.service.Display()
}

func countTrue(bools ...bool) (count int) {
	for _, b := range bools {
		if b {
			count++
		}
	}
	return count
}
