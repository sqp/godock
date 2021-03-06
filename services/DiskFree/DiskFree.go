// Package DiskFree is a monitoring applet for Cairo-Dock.
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

	"github.com/sqp/godock/libs/cdtype" // Applet types.
	"github.com/sqp/godock/libs/sysinfo"
)

func init() { cdtype.Applets.Register("DiskFree", NewApplet) }

// Applet defines a dock applet.
//
type Applet struct {
	cdtype.AppBase // Applet base and dock connection.

	conf    *appletConf
	service DiskFree
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

	// Disk free service.
	app.Poller().Add(app.service.Check)

	app.service.App = app
	app.service.log = app.Log()
	app.service.Texts = map[cdtype.InfoPosition]sysinfo.RenderOne{
		cdtype.InfoNone:    {},
		cdtype.InfoOnIcon:  {Sep: "\n", ShowPost: false},
		cdtype.InfoOnLabel: {Sep: "\n", ShowPost: true},
	}

	return app
}

// Init loads user configuration if needed and initialise applet.
//
func (app *Applet) Init(def *cdtype.Defaults, confLoaded bool) {
	// Settings for DiskFree.
	app.service.Settings(app.conf.DisplayText, 0, 0, app.conf.GaugeName)
	app.service.SetParts(app.conf.Partitions, app.conf.AutoDetect)

	// Defaults.
	def.PollerInterval = app.conf.UpdateDelay.Value()
	def.Commands = cdtype.Commands{
		cmdLeft:   cdtype.NewCommandStd(app.conf.LeftAction, app.conf.LeftCommand, app.conf.LeftClass),
		cmdMiddle: cdtype.NewCommandStd(app.conf.MiddleAction, app.conf.MiddleCommand),
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
	disks.SetSize(disks.nbValues)

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
		disks.SetSize(newcount)
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
