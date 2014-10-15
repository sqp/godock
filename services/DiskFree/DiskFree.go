// Package DiskFree is a monitoring applet for the Cairo-Dock project.
/*
Show disk usage for mounted partitions.
Partitions can be autodetected or you can provide a list of partitions.
You can also use autodetect with some partitions names to be listed first.

Partitions names are designed by their mount point like / or /home.
Use the df command to know your partitions list.
*/
package DiskFree

import (
	"github.com/cloudfoundry/gosigar" // Partitions and usage informations.

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
	service DiskFree
}

// NewApplet creates a new DiskFree applet instance.
//
func NewApplet() dock.AppletInstance {
	app := &Applet{CDApplet: dock.NewCDApplet()} // Icon controler and interface to cairo-dock.
	app.AddPoller(app.service.Check)

	app.service.App = app
	app.service.log = app.Log
	app.service.Texts = map[cdtype.InfoPosition]sysinfo.RenderOne{
		cdtype.InfoNone:    {},
		cdtype.InfoOnIcon:  {Sep: "\n", ShowPost: false},
		cdtype.InfoOnLabel: {Sep: "\n", ShowPost: true},
	}

	return app
}

// Init loads user configuration if needed and initialise applet.
//
func (app *Applet) Init(loadConf bool) {
	app.LoadConfig(loadConf, &app.conf) // Load config will crash if fail. Expected.

	// Settings for DiskFree.
	app.service.Settings(app.conf.DisplayText, 0, 0, app.conf.GaugeName)
	app.service.SetParts(app.conf.Partitions, app.conf.AutoDetect)

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

// DefineEvents sets applet events callbacks.
//
func (app *Applet) DefineEvents() {

	// Left and middle clicks: launch configured command.
	app.Events.OnClick = app.LaunchFunc("left")
	app.Events.OnMiddleClick = app.LaunchFunc("middle")

	app.Events.OnBuildMenu = func() {
		menu := []string{}
		if app.conf.LeftAction > 0 {
			menu = append(menu, "Action left click")
		}
		if app.conf.MiddleAction > 0 {
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
//----------------------------------------------------------------[ DISKFREE ]--

// DiskFree is a data poller for disk usage monitoring.
//
type DiskFree struct {
	sysinfo.RenderPercent

	autoDetect bool     // Will autodetect mounted partitions.
	names      []string // User provided list of partitions.
	nbValues   int

	log cdtype.Logger
}

// SetParts sets the user monitored pertitions.
//
func (disks *DiskFree) SetParts(parts []string, autoDetect bool) {
	disks.names = parts
	disks.autoDetect = autoDetect

	disks.nbValues = len(parts) + len(disks.findOthers())
	disks.SetSize(int32(disks.nbValues))

	if disks.nbValues == 0 {
		disks.log.NewErr("none", "disk found")
		disks.App.SetLabel("No disks found.")
	}
}

// Check updates disk usage information from the system.
//
func (disks *DiskFree) Check() {
	disks.Clear()

	parts := append(disks.names, disks.findOthers()...)

	for _, name := range parts {
		usage := sigar.FileSystemUsage{}
		value := float64(-1)
		if usage.Get(name) == nil { // no error
			value = float64(usage.UsePercent()) / 100
		}
		disks.Append(name, value)
	}

	if newcount := len(parts); newcount != disks.nbValues {
		disks.log.Debug("Number of partitions changed. Resizing", disks.nbValues, "=>", newcount)
		disks.nbValues = newcount
		disks.SetSize(int32(newcount))
	}

	disks.Display()
}

// findOthers returns the list of partitions found and not in user list.
//
func (disks *DiskFree) findOthers() (list []string) {
	if disks.autoDetect {
		all := sigar.FileSystemList{}
		all.Get()
		for _, fs := range all.List {
			if isFsValid(fs) && !disks.isListed(fs.DirName) {
				list = append(list, fs.DirName)
			}
		}
	}
	return
}

// isListed returns whether the provided partition is already in the user list or not.
//
func (disks *DiskFree) isListed(name string) bool {
	for _, diskName := range disks.names {
		if diskName == name {
			return true
		}
	}
	return false
}

// isFsValid returns whether the filesystem isn't in the banned list or not.
//
func isFsValid(fs sigar.FileSystem) bool {
	if fs.DevName == "none" ||
		fs.SysTypeName == "proc" || fs.SysTypeName == "sysfs" || fs.SysTypeName == "cgroup" ||
		fs.SysTypeName == "tmpfs" || fs.SysTypeName == "devtmpfs" || fs.SysTypeName == "devpts" {
		return false
	}

	if len(fs.DirName) > 5 {
		switch fs.DirName[:5] {
		case "/dev/", "/run/", "/sys/":
			return false
		}
		if fs.DirName[:6] == "/proc/" {
			return false
		}
	}

	return true
}
