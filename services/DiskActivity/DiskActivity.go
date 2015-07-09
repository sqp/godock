// Package DiskActivity is a monitoring applet for Cairo-Dock.
package DiskActivity

import (
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/dock" // Connection to cairo-dock.
	"github.com/sqp/godock/libs/sysinfo"
	"github.com/sqp/godock/libs/text/bytesize"

	"fmt"
)

//
//------------------------------------------------------------------[ APPLET ]--

// Applet data and controlers.
//
type Applet struct {
	cdtype.AppBase // Applet base and dock connection.

	conf    *appletConf
	service *sysinfo.IOActivity
}

// NewApplet create a new applet instance.
//
func NewApplet() cdtype.AppInstance {
	app := &Applet{AppBase: dock.NewCDApplet()} // Icon controler and interface to cairo-dock.

	app.service = sysinfo.NewIOActivity(app)
	app.service.Log = app.Log()
	app.service.FormatIcon = formatIcon
	app.service.FormatLabel = formatLabel
	app.service.GetData = sysinfo.GetDiskActivity

	app.AddPoller(app.service.Check)

	return app
}

// Init load user configuration if needed and initialise applet.
//
func (app *Applet) Init(loadConf bool) {
	app.LoadConfig(loadConf, &app.conf) // Load config will crash if fail. Expected.

	// Settings for poller and IOActivity (force renderer reset in case of reload).
	app.conf.UpdateDelay = cdtype.PollerInterval(app.conf.UpdateDelay, defaultUpdateDelay)
	app.service.Settings(uint64(app.conf.UpdateDelay), cdtype.InfoPosition(app.conf.DisplayText), app.conf.DisplayValues, app.conf.GraphType, app.conf.GaugeName, app.conf.Disks...)

	// Set defaults to dock icon: display and controls.
	app.SetDefaults(cdtype.Defaults{
		Label:          app.conf.Name,
		PollerInterval: app.conf.UpdateDelay,
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
//-----------------------------------------------------------------[ DISPLAY ]--

// Quick-info display callback. One line for each value. Zero are replaced by empty string.
//
func formatIcon(dev string, in, out uint64) string {
	return sysinfo.FormatRate(in*BlockSize) + "\n" + sysinfo.FormatRate(out*BlockSize)
}

// Label display callback. One line for each device. Format="eth0: ↓ 42 / ↑ 128".
//
func formatLabel(dev string, in, out uint64) string {
	return fmt.Sprintf("%s: %s %s / %s %s", dev, "r", bytesize.ByteSize(in*BlockSize), "w", bytesize.ByteSize(out*BlockSize))
}
