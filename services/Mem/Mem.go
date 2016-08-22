// Package Mem is a memory monitoring applet for Cairo-Dock.
package Mem

import (
	"github.com/cloudfoundry/gosigar" // System informations.

	"github.com/sqp/godock/libs/cdapplet" // Applet base.
	"github.com/sqp/godock/libs/cdtype"   // Applet types.
	"github.com/sqp/godock/libs/sysinfo"
)

// Applet data and controlers.
//
type Applet struct {
	cdtype.AppBase // Applet base and dock connection.

	conf    *appletConf
	service sysinfo.RenderPercent
}

// NewApplet create a new Mem applet instance.
//
func NewApplet() cdtype.AppInstance {
	app := &Applet{}
	app.AppBase = cdapplet.New(&app.conf) // Icon controler and interface to cairo-dock.

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
//------------------------------------------------------------------[ EVENTS ]--

// OnClick launch the configured action on user click.
//
func (app *Applet) OnClick(int) {
	app.Command().Launch(cmdLeft)
}

// OnMiddleClick launch the configured action on user middle click.
//
func (app *Applet) OnMiddleClick() {
	app.Command().Launch(cmdMiddle)
}

// OnBuildMenu fills the menu with left and middle click actions if they're set.
//
func (app *Applet) OnBuildMenu(menu cdtype.Menuer) {
	if app.conf.LeftAction > 0 && app.conf.LeftCommand != "" {
		menu.AddEntry("Action left click", "system-run", app.OnClick)
	}
	if app.conf.MiddleAction > 0 && app.conf.MiddleCommand != "" {
		menu.AddEntry("Action middle click", "system-run", app.OnMiddleClick)
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
