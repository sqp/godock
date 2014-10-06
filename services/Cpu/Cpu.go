// Package Cpu is a CPU monitoring applet for the Cairo-Dock project.
package Cpu

import (
	"github.com/cloudfoundry/gosigar" // System informations.

	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/dock" // Connection to cairo-dock.
	"github.com/sqp/godock/libs/sysinfo"
	"github.com/sqp/godock/libs/ternary" // Ternary operators.
)

//
//------------------------------------------------------------------[ APPLET ]--

// Applet data and controlers.
//
type Applet struct {
	*dock.CDApplet
	conf    *appletConf
	service *CPU
}

// NewApplet create a new applet instance.
//
func NewApplet() dock.AppletInstance {
	app := &Applet{
		CDApplet: dock.NewCDApplet(), // Icon controler and interface to cairo-dock.
		service:  NewCPU(),
	}

	app.AddPoller(app.service.Check)

	app.service.App = app
	app.service.Texts = map[cdtype.InfoPosition]sysinfo.RenderOne{
		cdtype.InfoNone:    {},
		cdtype.InfoOnIcon:  {ShowPre: false},
		cdtype.InfoOnLabel: {ShowPre: true},
	}

	return app
}

// Init load user configuration if needed and initialise applet.
//
func (app *Applet) Init(loadConf bool) {
	app.LoadConfig(loadConf, &app.conf) // Load config will crash if fail. Expected.

	// Settings for poller and renderer.
	app.service.Settings(app.conf.DisplayText, app.conf.DisplayValues, app.conf.GraphType, app.conf.GaugeName)
	app.service.SetSize(1)
	app.service.interval = dock.PollerInterval(app.conf.UpdateDelay, defaultUpdateDelay)

	// Set defaults to dock icon: display and controls.
	app.SetDefaults(dock.Defaults{
		Label:          ternary.String(app.conf.Name != "", app.conf.Name, app.AppletName),
		PollerInterval: app.service.interval,
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
//---------------------------------------------------------------------[ CPU ]--

// CPU monitors CPU activity and rendering on icon.
//
type CPU struct {
	sysinfo.RenderPercent

	lastIdle uint64
	nbCPU    float64
	interval int
}

// NewCPU creates a new CPU monitoring service.
//
func NewCPU() *CPU {
	list := sigar.CpuList{}
	list.Get()
	return &CPU{nbCPU: float64(len(list.List))}
}

// Check displays current CPU activity (average since last interval).
//
func (cpu *CPU) Check() {
	cpu.Clear()

	procs := sigar.Cpu{}
	procs.Get()

	if cpu.lastIdle > 0 { // Initialized.
		delta := procs.Idle - cpu.lastIdle
		used := float64(delta) / cpu.nbCPU / float64(cpu.interval)

		cpu.Append("CPU", (100-used)/100)
		cpu.Display()
	}
	cpu.lastIdle = procs.Idle
}
