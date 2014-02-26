/* Disk activity monitoring applet for the Cairo-Dock project.
 */
package main

import (
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/dock" // Connection to cairo-dock.
	"github.com/sqp/godock/libs/log"  // Display info in terminal.

	"github.com/sqp/godock/libs/packages" // ByteSize.
	"github.com/sqp/godock/libs/ternary"  // Ternary operators.

	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
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
	disks *DiskSpeed

	conf *appletConf
}

// Create a new applet instance.
//
func NewApplet() *Applet {
	app := &Applet{
		CDApplet: dock.Applet(), // Icon controler and interface to cairo-dock.
	}
	app.disks = NewDiskSpeed(app)
	app.AddPoller(func() { app.disks.GetData(); app.disks.Display() })

	return app
}

// Load user configuration if needed and initialise applet.
//
func (app *Applet) Init(loadConf bool) {
	if loadConf { // Try to load config. Exit if not found.
		app.conf = &mailConf{}
		log.Fatal(app.LoadConfig(&app.conf, dock.GetBoth), "config")
	}

	// Settings for poller and diskSpeed (force renderer reset in case of reload).
	app.conf.UpdateDelay = dock.PollerInterval(app.conf.UpdateDelay, defaultUpdateDelay)
	app.disks.Settings(uint64(app.conf.UpdateDelay), cdtype.InfoPosition(app.conf.DisplayText), app.conf.DisplayValues, app.conf.GraphType, app.conf.GaugeName, app.conf.Disks...)

	// Set defaults to dock icon: display and controls.
	app.SetDefaults(cdtype.Defaults{
		Label:          ternary.String(app.conf.Name != "", app.conf.Name, app.AppletName),
		PollerInterval: app.conf.UpdateDelay,
		Commands: cdtype.Commands{
			"left":   cdtype.NewCommandStd(app.conf.LeftAction, app.conf.LeftCommand, app.conf.LeftClass),
			"middle": cdtype.NewCommandStd(app.conf.MiddleAction, app.conf.MiddleCommand)},
		Debug: app.conf.Debug})

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
//---------------------------------------------------------------[ DISKSPEED ]--

type stat struct {
	rateReadNow  uint64
	rateWriteNow uint64

	rateReadMax  uint64
	rateWriteMax uint64

	blocksRead  uint64
	blocksWrite uint64

	bInitialized  bool // true after the 2nd data pull.
	acquisitionOK bool // true if data was found this pull.
}

// Create a new data poller for disk activity monitoring.
//
type DiskSpeed struct {
	list         map[string]*stat
	interval     uint64
	textPosition cdtype.InfoPosition
	app          cdtype.RenderSimple // Controler to the Cairo-Dock icon.
}

func NewDiskSpeed(app cdtype.RenderSimple) *DiskSpeed {
	return &DiskSpeed{
		list: make(map[string]*stat),
		app:  app,
	}
}

// Apply user settings from config.
//
func (disks *DiskSpeed) Settings(interval uint64, textPosition cdtype.InfoPosition, renderer, graphType int, gaugeTheme string, names ...string) {
	disks.interval = interval
	disks.textPosition = textPosition

	disks.list = make(map[string]*stat) // Clear list. Nothing must remain.
	disks.app.AddDataRenderer("", 0, "")

	if len(names) > 0 {
		for _, name := range names {
			disks.list[name] = &stat{}
		}

		switch renderer {
		case 0:
			disks.app.AddDataRenderer("gauge", 2*int32(len(disks.list)), gaugeTheme)
		case 1:
			disks.app.AddDataRenderer("graph", 2*int32(len(disks.list)), DockGraphType[graphType])
		}
	} else {
		log.DEV("no disks ffs")
		disks.app.SetLabel("No disks defined.")
	}
}

//
//----------------------------------------------------------------[ GET DATA ]--

// Get activity information for configured disks.
//
// Using Linux iostat : http://www.kernel.org/doc/Documentation/iostats.txt
// gathering fields 3, 6 and 10
//
func (disks *DiskSpeed) GetData() {
	if len(disks.list) == 0 {
		return
	}
	for _, stat := range disks.list { // Reset our acquisition status for every disk.
		stat.acquisitionOK = false
	}

	if file, e := os.Open(KernelDiskStats); !log.Err(e, "Your kernel doesn't support diskstat. (2.5.70 or above required)") {
		r := bufio.NewReader(file)
		line, err := r.ReadString('\n')
		for err == nil {
			if len(line) > 13 { // Drop first part with static size (2 fields).
				line = line[13:]

				data := strings.Fields(line) // Useful data is only separated by a blank space.
				if len(data) > 10 {
					if stat, ok := disks.list[data[0]]; ok {
						stat.acquisitionOK = true
						blocksReadNew, _ := strconv.ParseUint(data[3], 10, 64)
						blocksWriteNew, _ := strconv.ParseUint(data[7], 10, 64)

						if stat.bInitialized { // Drop first pull. Values are stupidly high: total since boot time.
							stat.rateReadNow = (blocksReadNew - stat.blocksRead) * BlockSize / disks.interval
							stat.rateWriteNow = (blocksWriteNew - stat.blocksWrite) * BlockSize / disks.interval
						}

						// Save our new values.
						stat.blocksRead = blocksReadNew
						stat.blocksWrite = blocksWriteNew
						stat.bInitialized = true
					}
				}
			}

			line, err = r.ReadString('\n')
		}
	}
}

//
//-----------------------------------------------------------------[ DISPLAY ]--

// Display disk activity info on the Cairo-Dock icon (renderer, quickinfo, label).
//
func (disks *DiskSpeed) Display() {
	if len(disks.list) == 0 {
		return
	}
	var values []float64
	var text string
	for name, stat := range disks.list {
		// Text separator.
		if text != "" && (disks.textPosition == cdtype.InfoOnIcon || disks.textPosition == cdtype.InfoOnLabel) {
			text += "\n"
		}

		if stat.acquisitionOK {
			values = append(values, currentRate(stat.rateReadNow, &stat.rateReadMax))
			values = append(values, currentRate(stat.rateWriteNow, &stat.rateWriteMax))

			switch disks.textPosition { // Add text renderer info.
			case cdtype.InfoOnIcon:
				text += fmt.Sprintf("%s\n%s", formatRate(stat.rateReadNow), formatRate(stat.rateWriteNow))

			case cdtype.InfoOnLabel:
				text += fmt.Sprintf("%s : %s %s / %s %s", name, "r", packages.ByteSize(stat.rateReadNow), "w", packages.ByteSize(stat.rateWriteNow)) // NEED TRANSLATE GETTEXT
			}

		} else {
			values = append(values, 0, 0)

			switch disks.textPosition {
			case cdtype.InfoOnIcon:
				text += "N/A" // NEED TRANSLATE GETTEXT

			case cdtype.InfoOnLabel:
				text += fmt.Sprintf("%s : %s", name, "N/A") // NEED TRANSLATE GETTEXT
			}
		}
	}

	if len(values) > 0 {
		disks.app.RenderValues(values...)
	}

	if disks.textPosition == cdtype.InfoOnIcon {
		disks.app.SetQuickInfo(text)
	}

	if disks.textPosition == cdtype.InfoOnLabel {
		disks.app.SetLabel(text)
	}
}

func currentRate(speed uint64, max *uint64) float64 {
	if speed > *max {
		*max = speed
	}
	if *max > 0 {
		return float64(speed) / float64(*max)
	}
	return 0
}

func formatRate(size uint64) string {
	if size > 0 {
		// *log.DEV("size", size)
		return packages.ByteSize(size).String()
	}
	return ""
}
