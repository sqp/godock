// Network activity monitoring applet for the Cairo-Dock project.
package NetActivity

/*

Benefit from original version:
-not using temp files.
-new image upload site: pix.toile-libre.org
-a lot of upload sites don't require external dependencies.

Dependencies:
xsel or xclip for clipboard interaction.

curl command is needed for those backends:
  Image: ImageShackUs, ImgurCom, UppixCom
  File:  FreeFr

Compile:
libcurl-dev  (I'm using libcurl4-gnutls-dev)
glib-2.0

Not implemented (yet):
- real icon for the applet.
- Upload raw text with FileForAll option. I'm trying to find a way to do it
   without the temp file option before falling back to this method.
- More menu options, due to lack of proper AddMenuItem method.
- Save image copy (and display)
- Custom upload scripts
- Url shortener (as I'm not fan of those, you better do it yourself if you want it)

Unsure:
-Dropbox and Ubuntu-One: copying files to a local folder and launching a sync tool
doesn't seem to fit in the applet description for me.

*/

import (
	"github.com/atotto/clipboard"

	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/dock" // Connection to cairo-dock.
	"github.com/sqp/godock/libs/log"
	"github.com/sqp/godock/libs/packages" // ByteSize.
	"github.com/sqp/godock/libs/sysinfo"
	"github.com/sqp/godock/libs/ternary" // Ternary operators.
	"github.com/sqp/godock/libs/uptoshare"

	"fmt"
	"path"
)

const EmblemAction = cdtype.EmblemTopRight

//
//------------------------------------------------------------------[ APPLET ]--

// Applet data and controlers.
//
type Applet struct {
	*dock.CDApplet
	conf        *appletConf
	service     *sysinfo.IOActivity
	up          *uptoshare.Uploader
	menuActions []func() // Menu callbacks are saved to be sure we launch the good action (history can change).
}

// Create a new applet instance.
//
func NewApplet() dock.AppletInstance {
	app := &Applet{CDApplet: dock.NewCDApplet()} // Icon controler and interface to cairo-dock.

	// Uptoshare actions
	app.up = uptoshare.New()
	app.up.SetPreCheck(func() { app.SetEmblem(app.FileLocation("icon"), EmblemAction) })
	app.up.SetPostCheck(func() { app.SetEmblem("none", EmblemAction) })
	app.up.SetOnResult(app.onUpload)

	// Network activity actions.
	app.service = sysinfo.NewIOActivity(app)
	app.service.FormatIcon = sysinfo.FormatIcon
	app.service.FormatLabel = formatLabel
	app.service.GetData = sysinfo.GetNetActivity

	app.AddPoller(app.service.Check)

	return app
}

// Load user configuration if needed and initialise applet.
//
func (app *Applet) Init(loadConf bool) {
	app.LoadConfig(loadConf, &app.conf) // Load config will crash if fail. Expected.

	// Uptoshare settings.
	app.up.SetHistoryFile(path.Join(app.RootDataDir, historyFile))
	app.up.SetHistorySize(app.conf.UploadHistory)
	app.up.LimitRate = app.conf.UploadRateLimit
	app.up.PostAnonymous = app.conf.PostAnonymous
	app.up.FileForAll = app.conf.FileForAll
	app.up.SiteImage(app.conf.SiteImage)
	app.up.SiteText(app.conf.SiteText)
	app.up.SiteVideo(app.conf.SiteVideo)
	app.up.SiteFile(app.conf.SiteFile)

	// Settings for poller and IOActivity (force renderer reset in case of reload).
	app.conf.UpdateDelay = dock.PollerInterval(app.conf.UpdateDelay, defaultUpdateDelay)
	app.service.Settings(uint64(app.conf.UpdateDelay), cdtype.InfoPosition(app.conf.DisplayText), app.conf.DisplayValues, app.conf.GraphType, app.conf.GaugeName, app.conf.Devices...)

	// Set defaults to dock icon: display and controls.
	app.SetDefaults(dock.Defaults{
		Label:          ternary.String(app.conf.Name != "", app.conf.Name, app.AppletName),
		PollerInterval: app.conf.UpdateDelay,
		Commands: dock.Commands{
			"left":   dock.NewCommandStd(app.conf.LeftAction, app.conf.LeftCommand, app.conf.LeftClass),
			"middle": dock.NewCommandStd(app.conf.MiddleAction, app.conf.MiddleCommand)},
		Debug: app.conf.Debug})
}

//
//------------------------------------------------------------------[ EVENTS ]--

// Define applet events callbacks.
//
func (app *Applet) DefineEvents() {
	app.Events.OnClick = app.LaunchFunc("left")
	app.Events.OnMiddleClick = app.LaunchFunc("middle")

	app.Events.OnDropData = app.up.Upload

	app.Events.OnBuildMenu = app.buildMenu
	app.Events.OnMenuSelect = app.clickedMenu
}

//
//--------------------------------------------------------------------[ MENU ]--

func (app *Applet) buildMenu() {
	menu := []string{}
	app.menuActions = nil

	if app.conf.LeftAction > 0 {
		menu = append(menu, "Action left click")
		app.menuActions = append(app.menuActions, func() { app.LaunchCommand("left") })
	}
	if app.conf.MiddleAction > 0 {
		menu = append(menu, "Action middle click")
		app.menuActions = append(app.menuActions, func() { app.LaunchCommand("middle") })
	}

	if len(menu) > 0 { // Add separator.
		menu = append(menu, "")
		app.menuActions = append(app.menuActions, nil)
	}

	for _, hist := range app.up.ListHistory() {
		menu = append(menu, path.Base(hist["file"]))
		app.addMenuPaste(hist["link"])
	}

	app.PopulateMenu(menu...)
}

func (app *Applet) clickedMenu(i int32) {
	app.menuActions[i]()
}

func (app *Applet) addMenuPaste(link string) {
	app.menuActions = append(app.menuActions, func() {
		log.Info(link)
		clipboard.WriteAll(link)
		// app.ShowDialog(link, 5)
	})
}

//
//----------------------------------------------------------------[ CALLBACK ]--

// onUpload is called with the list of links when an item has been uploaded.
//
func (app *Applet) onUpload(links uptoshare.Links) {
	if e, ok := links["error"]; ok {
		app.ShowDialog("Error: "+e, 10)
		return
	}

	if app.conf.DialogEnabled {
		if link, ok := links["link"]; ok {
			clipboard.WriteAll(link)
			app.ShowDialog(link, app.conf.DialogDuration)
		}
	}

	for k, v := range links { // TMP TO DEL
		log.Info(k, v)
	}
}

//
//-----------------------------------------------------------------[ DISPLAY ]--

// Label display callback. One line for each device. Format="eth0: ↓ 42 / ↑ 128".
//
func formatLabel(dev string, in, out uint64) string {
	return fmt.Sprintf("%s: %s %s / %s %s", dev, "↓", packages.ByteSize(in), "↑", packages.ByteSize(out))
}
