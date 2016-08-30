// Package Cpu is a CPU monitoring applet for Cairo-Dock.
package Cpu

import (
	"github.com/cloudfoundry/gosigar" // System informations.

	"github.com/sqp/godock/libs/cdtype" // Applet types.
	"github.com/sqp/godock/libs/sysinfo"
)

//
//------------------------------------------------------------------[ APPLET ]--

func init() { cdtype.Applets.Register("Cpu", NewApplet) }

// Applet defines a dock applet.
//
type Applet struct {
	cdtype.AppBase // Applet base and dock connection.

	conf    *appletConf
	service *CPU
}

// NewApplet creates a new applet instance.
//
func NewApplet(base cdtype.AppBase, events *cdtype.Events) cdtype.AppInstance {
	app := &Applet{AppBase: base, service: NewCPU()}
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

	app.Poller().Add(app.service.Check)

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
func (app *Applet) Init(def *cdtype.Defaults, confLoaded bool) {
	// Settings for poller and renderer.
	app.service.Settings(app.conf.DisplayText, app.conf.DisplayValues, app.conf.GraphType, app.conf.GaugeName)
	app.service.SetSize(1)
	app.service.interval = app.conf.UpdateDelay.Value()

	// Defaults.
	def.PollerInterval = app.conf.UpdateDelay.Value()
	def.Commands[cmdLeft] = cdtype.NewCommandStd(app.conf.LeftAction, app.conf.LeftCommand, app.conf.LeftClass)
	def.Commands[cmdMiddle] = cdtype.NewCommandStd(app.conf.MiddleAction, app.conf.MiddleCommand)
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
