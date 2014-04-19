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
	"github.com/sqp/godock/libs/dock"    // Connection to cairo-dock.
	"github.com/sqp/godock/libs/ternary" // Ternary operators.

	"strconv"
)

// Applet DiskUsage data and controlers.
//
type Applet struct {
	*dock.CDApplet
	disks *diskFree
	conf  *appletConf
}

// NewApplet create a new DiskUsage applet instance.
//
func NewApplet() dock.AppletInstance {
	app := &Applet{CDApplet: dock.NewCDApplet()} // Icon controler and interface to cairo-dock.

	app.disks = newDiskFree(app)
	app.disks.log = app.Log
	app.AddPoller(func() { app.disks.GetData(); app.disks.Display() })

	return app
}

// Init load user configuration if needed and initialise applet.
//
func (app *Applet) Init(loadConf bool) {
	app.LoadConfig(loadConf, &app.conf) // Load config will crash if fail. Expected.

	// Set defaults to dock icon: display and controls.
	app.SetDefaults(dock.Defaults{
		Label:          ternary.String(app.conf.Name != "", app.conf.Name, app.AppletName),
		PollerInterval: dock.PollerInterval(app.conf.UpdateDelay, defaultUpdateDelay),
		Commands: dock.Commands{
			"left":   dock.NewCommandStd(app.conf.LeftAction, app.conf.LeftCommand, app.conf.LeftClass),
			"middle": dock.NewCommandStd(app.conf.MiddleAction, app.conf.MiddleCommand)},
		Debug: app.conf.Debug})

	// Settings for diskFree and poller.
	app.disks.Settings(cdtype.InfoPosition(app.conf.DisplayText), app.conf.AutoDetect, app.conf.GaugeName, app.conf.Partitions...)
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

// Informations about a partition and its usage.
//
type fileSystem struct {
	info  sigar.FileSystem
	usage sigar.FileSystemUsage
}

// Data poller for disk usage monitoring.
//
type diskFree struct {
	listUser   map[string]*fileSystem // Data about user list of partitions.
	listFound  map[string]*fileSystem // Data about other partitions.
	autoDetect bool                   // Will autodetect mounted partitions.
	names      []string               // User provided list of partitions.
	nbValues   int

	textPosition cdtype.InfoPosition
	gaugeName    string
	app          dock.RenderSimple // Controler to the Cairo-Dock icon.
	log          cdtype.Logger
}

// Create a new data poller for disk usage monitoring.
//
func newDiskFree(app dock.RenderSimple) *diskFree {
	return &diskFree{app: app}
}

// Apply user settings from config.
//
func (disks *diskFree) Settings(textPosition cdtype.InfoPosition, autoDetect bool, gaugeName string, names ...string) {
	disks.textPosition = textPosition
	disks.autoDetect = autoDetect
	disks.gaugeName = gaugeName
	disks.names = names

	disks.app.AddDataRenderer("", 0, "")          // Remove renderer when settings changed to be sure.
	disks.listUser = make(map[string]*fileSystem) // Clear list. Nothing must remain.
	disks.clearData()

	disks.nbValues = len(disks.listUser) + disks.countFound()

	if disks.nbValues == 0 {
		disks.log.NewErr("none", "disk found")
		disks.app.SetLabel("No disks found.")
	}
	disks.setRenderer()
}

// Set the Cairo-Dock renderer.
//
func (disks *diskFree) setRenderer() {
	disks.app.AddDataRenderer("gauge", int32(disks.nbValues), disks.gaugeName)
}

// Clear internal before a new check.
//
func (disks *diskFree) clearData() {
	disks.listFound = make(map[string]*fileSystem)
	for _, name := range disks.names { // Fill user list with placeholders to detect missing ones.
		disks.listUser[name] = nil
	}
}

//
//----------------------------------------------------------------[ GET DATA ]--

// Get disk usage information from the system.
//
func (disks *diskFree) GetData() {
	disks.clearData()

	fullList := sigar.FileSystemList{}
	fullList.Get()
	for _, fs := range fullList.List {
		if isFsValid(fs) {
			part := &fileSystem{
				info:  fs,
				usage: sigar.FileSystemUsage{},
			}
			part.usage.Get(fs.DirName)

			if _, ok := disks.listUser[fs.DirName]; ok { // Fill user provided list with data.
				disks.listUser[fs.DirName] = part

			} else if disks.autoDetect { // Only fill second list if needed.
				disks.listFound[fs.DirName] = part
			}
		}
	}

	if newcount := len(disks.listUser) + len(disks.listFound); newcount != disks.nbValues {
		disks.log.Debug("Number of partitions changed. Resizing", disks.nbValues, "=>", newcount)
		disks.nbValues = newcount
		disks.setRenderer()
	}
}

// Count extra disks for the size of the renderer (only those not in the user list).
//
func (disks *diskFree) countFound() (count int) {
	if disks.autoDetect {
		fullList := sigar.FileSystemList{}
		fullList.Get()
		for _, fs := range fullList.List {
			if _, ok := disks.listUser[fs.DirName]; !ok && isFsValid(fs) {
				disks.log.Debug("found extra partition", fs.DirName)
				count++
			}
		}
	}
	return
}

// Check if the filesystem isn't in the banned list.
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

//
//-----------------------------------------------------------------[ DISPLAY ]--

// Display disk usage info on the Cairo-Dock icon (renderer, quickinfo, label).
//
func (disks *diskFree) Display() {
	var values []float64
	var text string

	// User defined partitions.
	// for name, fs := range disks.listUser {
	for _, fs := range disks.listUser {
		var value float64

		if fs != nil {
			value = float64(fs.usage.UsePercent())
		} else {
			// disks.log.Info("DISK NOT FOUND", name)
		}

		values = append(values, value/100)
		disks.appendText(&text, value, fs)
	}

	// Autodetected partitions.
	for _, fs := range disks.listFound {
		value := float64(fs.usage.UsePercent())
		values = append(values, value/100)
		disks.appendText(&text, value, fs)
	}

	if len(values) > 0 {
		disks.app.RenderValues(values...)
	}

	switch disks.textPosition {
	case cdtype.InfoOnIcon:
		disks.app.SetQuickInfo(text)

	case cdtype.InfoOnLabel:
		disks.app.SetLabel(text)
	}
}

// Append one value to the text info.
//
func (disks *diskFree) appendText(text *string, value float64, fs *fileSystem) {
	if *text != "" {
		*text += "\n"
	}

	if value > -1 && fs != nil {
		*text += strconv.FormatFloat(value, 'f', 0, 64) + "%"
	} else {
		*text += "N/A"
		// return
	}

	// disks.log.Debug(curText + " : " + fs.info.DirName)

	switch disks.textPosition {
	// case cdtype.InfoOnIcon:

	case cdtype.InfoOnLabel:
		*text += " : " + fs.info.DirName //  fmt.Sprintf("%s : %s", curText, fs.info.DirName)
	}
}
