// Package Mem is a memory monitoring applet for the Cairo-Dock project.
package Mem

import (
	"github.com/cloudfoundry/gosigar" // System informations.

	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/dock" // Connection to cairo-dock.
	"github.com/sqp/godock/libs/sysinfo"
	"github.com/sqp/godock/libs/ternary" // Ternary operators.
)

// Applet data and controlers.
//
type Applet struct {
	*dock.CDApplet
	conf    *appletConf
	service sysinfo.RenderPercent
}

// NewApplet create a new Mem applet instance.
//
func NewApplet() dock.AppletInstance {
	app := &Applet{CDApplet: dock.NewCDApplet()} // Icon controler and interface to cairo-dock.
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
	app.service.SetSize(boolToInt(app.conf.ShowRam) + boolToInt(app.conf.ShowSwap))

	// Set defaults to dock icon: display and controls.
	app.SetDefaults(dock.Defaults{
		Label:          ternary.String(app.conf.Name != "", app.conf.Name, app.AppletName),
		PollerInterval: dock.PollerInterval(app.conf.UpdateDelay, defaultUpdateDelay),
		Commands: dock.Commands{
			"left":   dock.NewCommandStd(app.conf.LeftAction, app.conf.LeftCommand, app.conf.LeftClass),
			"middle": dock.NewCommandStd(app.conf.MiddleAction, app.conf.MiddleCommand)},
		Debug: app.conf.Debug})
}

//
//------------------------------------------------------------------[ EVENTS ]--

// DefineEvents set applet events callbacks.
//
func (app *Applet) DefineEvents() {

	// Left and middle clicks: launch configured command.
	app.Events.OnClick = app.LaunchFunc("left")
	app.Events.OnMiddleClick = app.LaunchFunc("middle")

	app.Events.OnBuildMenu = func() {
		menu := []string{}
		if app.conf.LeftAction > 0 && app.conf.LeftCommand != "" {
			menu = append(menu, "Action left click")
		}
		if app.conf.MiddleAction > 0 && app.conf.MiddleCommand != "" {
			menu = append(menu, "Action middle click")
		}
		app.PopulateMenu(menu...)
	}

	app.Events.OnMenuSelect = func(i int32) {
		list := []string{"left", "middle"}
		app.LaunchCommand(list[i])
	}
}

//
//---------------------------------------------------------------------[ MEM ]--

// GetMemActivity displays memory activity.
//
func (app *Applet) GetMemActivity() {
	app.service.Clear()

	if app.conf.ShowRam {
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

func boolToInt(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
