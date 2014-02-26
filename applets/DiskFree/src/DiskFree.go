/* Disk usage monitoring applet for the Cairo-Dock project.

Show disk usage for mounted partitions.
Partitions can be autodetected or you can provide a list of partitions.
You can also use autodetect with some partitions names to be listed first.

Partitions names are designed by their mount point like / or /home.
Use the df command to know your partitions list.
*/
package main

import (
	"github.com/cloudfoundry/gosigar" // Partitions and usage informations.

	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/dock"    // Connection to cairo-dock.
	"github.com/sqp/godock/libs/log"     // Display info in terminal.
	"github.com/sqp/godock/libs/ternary" // Ternary operators.
)

// Program launched. Create and activate applet.
//
func main() {
	dock.StartApplet(NewApplet())
}

//
//------------------------------------------------------------------[ APPLET ]--

// Applet data and controlers.
//
type Applet struct {
	*dock.CDApplet
	// poller *poller.Poller
	disks *diskFree
	conf  *appletConf
}

// Create a new applet instance.
//
func NewApplet() *Applet {
	app := &Applet{
		CDApplet: dock.Applet(), // Icon controler and interface to cairo-dock.
	}
	app.disks = newDiskFree(app)
	app.AddPoller(func() { app.disks.GetData(); app.disks.Display() })

	return app
}

// Load user configuration if needed and initialise applet.
//
func (app *Applet) Init(loadConf bool) {
	if loadConf { // Try to load config. Exit if not found.
		app.conf = &appletConf{}
		log.Fatal(app.LoadConfig(&app.conf, dock.GetBoth), "config")
	}

	// Set defaults to dock icon: display and controls.
	app.SetDefaults(cdtype.Defaults{
		Label:          ternary.String(app.conf.Name != "", app.conf.Name, app.AppletName),
		PollerInterval: dock.PollerInterval(app.conf.UpdateDelay, defaultUpdateDelay),
		Commands: cdtype.Commands{
			"left":   cdtype.NewCommandStd(app.conf.LeftAction, app.conf.LeftCommand, app.conf.LeftClass),
			"middle": cdtype.NewCommandStd(app.conf.MiddleAction, app.conf.MiddleCommand)},
		Debug: app.conf.Debug})

	// Settings for diskFree and poller.
	app.disks.Settings(cdtype.InfoPosition(app.conf.DisplayText), app.conf.AutoDetect, app.conf.GaugeName, app.conf.Partitions...)
}

//
//------------------------------------------------------------------[ EVENTS ]--

// Define applet events callbacks.
//
func (app *Applet) DefineEvents() {

	// Left and middle clicks: launch configured command.
	app.Events.OnClick = app.LaunchFunc("left")
	app.Events.OnMiddleClick = app.LaunchFunc("middle")

	app.Events.OnBuildMenu = func() {
		// menu := []string{"", "ok"} // First entry is a separator.
		// app.PopulateMenu(menu...)
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
	app          cdtype.RenderSimple // Controler to the Cairo-Dock icon.
}

// Create a new data poller for disk usage monitoring.
//
func newDiskFree(app cdtype.RenderSimple) *diskFree {
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
		log.DEV("no disks ffs")
		disks.app.SetLabel("No disks found.")
		// TODO: need to stop the timer ?
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
		log.Info("Number of partitions changed. Resizing", disks.nbValues, "=>", newcount)
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
				log.Info("found extra partition", fs.DirName)
				count++
			}
		}
	}
	return
}

// Check if the filesystem isn't in the banned list.
//
func isFsValid(fs sigar.FileSystem) bool {
	dropped := fs.DevName == "none" ||
		fs.SysTypeName == "proc" || fs.SysTypeName == "sysfs" || fs.SysTypeName == "cgroup" ||
		fs.SysTypeName == "tmpfs" || fs.SysTypeName == "devtmpfs" || fs.SysTypeName == "devpts" ||
		(len(fs.DirName) > 5 && fs.DirName[:5] == "/run/") ||
		(len(fs.DirName) > 6 && fs.DirName[:6] == "/proc/")
	return !dropped
}

//
//-----------------------------------------------------------------[ DISPLAY ]--

// Display disk usage info on the Cairo-Dock icon (renderer, quickinfo, label).
//
func (disks *diskFree) Display() {
	var values []float64
	var text string

	// User defined partitions.
	for name, fs := range disks.listUser {
		var value float64 = 0

		if fs != nil {
			value = float64(fs.usage.UsePercent())
		} else {
			log.Info("DISK NOT FOUND", name)
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

	curText := ""
	if value > -1 {
		curText = sigar.FormatPercent(value)
	} else {
		curText = "N/A"
	}

	log.Debug(curText + " : " + fs.info.DirName)

	switch disks.textPosition {
	case cdtype.InfoOnIcon:
		*text += curText

	case cdtype.InfoOnLabel:
		*text += curText + " : " + fs.info.DirName //  fmt.Sprintf("%s : %s", curText, fs.info.DirName)
	}
}
