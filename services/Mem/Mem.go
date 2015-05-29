// Package Mem is a memory monitoring applet for Cairo-Dock.
package Mem

import (
	"github.com/cloudfoundry/gosigar" // System informations.

	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/dock" // Connection to cairo-dock.
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
	app := &Applet{AppBase: dock.NewCDApplet()} // Icon controler and interface to cairo-dock.
	app.AddPoller(app.GetMemActivity)

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
func (app *Applet) Init(loadConf bool) {
	app.LoadConfig(loadConf, &app.conf) // Load config will crash if fail. Expected.

	// Settings for service.
	app.service.Settings(app.conf.DisplayText, app.conf.DisplayValues, app.conf.GraphType, app.conf.GaugeName)
	app.service.SetSize(countTrue(app.conf.ShowRAM, app.conf.ShowSwap))

	// Set defaults to dock icon: display and controls.
	app.SetDefaults(cdtype.Defaults{
		Label:          app.conf.Name,
		PollerInterval: cdtype.PollerInterval(app.conf.UpdateDelay, defaultUpdateDelay),
		Commands: cdtype.Commands{
			"left":   cdtype.NewCommandStd(app.conf.LeftAction, app.conf.LeftCommand, app.conf.LeftClass),
			"middle": cdtype.NewCommandStd(app.conf.MiddleAction, app.conf.MiddleCommand)},
		Debug: app.conf.Debug})
}

//
//------------------------------------------------------------------[ EVENTS ]--

// OnClick launch the configured action on user click.
//
func (app *Applet) OnClick() {
	app.CommandLaunch("left")
}

// OnMiddleClick launch the configured action on user middle click.
//
func (app *Applet) OnMiddleClick() {
	app.CommandLaunch("middle")
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

func countTrue(bools ...bool) (count int32) {
	for _, b := range bools {
		if b {
			count++
		}
	}
	return count
}
