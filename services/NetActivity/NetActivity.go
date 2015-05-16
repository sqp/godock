// Package NetActivity is a monitoring and upload applet for Cairo-Dock.
/*

Benefit from original version:
  Not using temp files.
  New image upload site: http://pix.toile-libre.org
  A lot of upload sites don't require external dependencies.

Dependencies:
	xsel or xclip for clipboard interaction.

	curl command is needed for those backends:
	  Image: ImageShackUs, ImgurCom, UppixCom
	  File:  FreeFr

Compile:
  libcurl-dev  (I'm using libcurl4-gnutls-dev)
  glib-2.0

Not implemented (yet):
  Icon for the applet.
  Upload raw text with FileForAll option. I'm trying to find a way to do it
    without the temp file option before falling back to this method.
  More menu options, due to lack of proper AddMenuItem method.
  Save image copy (and display).
  Custom upload scripts.
  Url shortener (as I'm not fan of those, you better do it yourself if you want it).

Unsure:
  Dropbox and Ubuntu-One: copying files to a local folder and launching a sync tool
    doesn't seem to fit in the applet description for me.
    Ubuntu-one problem solved.

*/
package NetActivity

import (
	"github.com/atotto/clipboard"

	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/cdtype/bytesize"
	"github.com/sqp/godock/libs/dock"      // Connection to cairo-dock.
	"github.com/sqp/godock/libs/sysinfo"   // IOActivity.
	"github.com/sqp/godock/libs/uptoshare" // Uploader service.

	"fmt"
)

// EmblemAction is the position of the "upload in progress" emblem.
const EmblemAction = cdtype.EmblemTopRight

//
//------------------------------------------------------------------[ APPLET ]--

// Applet data and controlers.
//
type Applet struct {
	cdtype.AppBase // Applet base and dock connection.

	conf    *appletConf
	service *sysinfo.IOActivity
	up      *uptoshare.Uploader
}

// NewApplet create a new applet instance.
//
func NewApplet() cdtype.AppInstance {
	app := &Applet{AppBase: dock.NewCDApplet()} // Icon controler and interface to cairo-dock.

	// Uptoshare actions
	app.up = uptoshare.New()
	app.up.Log = app.Log()
	app.up.SetPreCheck(func() { app.SetEmblem(app.FileLocation("icon"), EmblemAction) })
	app.up.SetPostCheck(func() { app.SetEmblem("none", EmblemAction) })
	app.up.SetOnResult(app.onUpload)

	// Network activity actions.
	app.service = sysinfo.NewIOActivity(app)
	app.service.Log = app.Log()
	app.service.FormatIcon = sysinfo.FormatIcon
	app.service.FormatLabel = formatLabel
	app.service.GetData = sysinfo.GetNetActivity

	app.AddPoller(app.service.Check)

	return app
}

// Init load user configuration if needed and initialise applet.
//
func (app *Applet) Init(loadConf bool) {
	app.LoadConfig(loadConf, &app.conf) // Load config will crash if fail. Expected.

	// Uptoshare settings.
	app.up.SetHistoryFile(app.FileDataDir(historyFile))
	app.up.SetHistorySize(app.conf.UploadHistory)
	app.up.LimitRate = app.conf.UploadRateLimit
	app.up.PostAnonymous = app.conf.PostAnonymous
	app.up.FileForAll = app.conf.FileForAll
	app.up.SiteImage(app.conf.SiteImage)
	app.up.SiteText(app.conf.SiteText)
	app.up.SiteVideo(app.conf.SiteVideo)
	app.up.SiteFile(app.conf.SiteFile)

	// Settings for poller and IOActivity (force renderer reset in case of reload).
	app.conf.UpdateDelay = cdtype.PollerInterval(app.conf.UpdateDelay, defaultUpdateDelay)
	app.service.Settings(uint64(app.conf.UpdateDelay), cdtype.InfoPosition(app.conf.DisplayText), app.conf.DisplayValues, app.conf.GraphType, app.conf.GaugeName, app.conf.Devices...)

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
	needSep := false
	if app.conf.LeftAction > 0 && app.conf.LeftCommand != "" {
		menu.AddEntry("Action left click", "gtk-execute", app.OnClick)
		needSep = true
	}
	if app.conf.MiddleAction > 0 && app.conf.MiddleCommand != "" {
		menu.AddEntry("Action middle click", "gtk-execute", app.OnMiddleClick)
		needSep = true
	}
	if needSep {
		menu.Separator()
	}
	for _, hist := range app.up.ListHistory() {
		hist := hist
		menu.AddEntry(hist["file"], "", func() {
			app.Log().Info(hist["link"])
			clipboard.WriteAll(hist["link"])
			// app.ShowDialog(link, 5)
		})
	}

}

func (app *Applet) OnDropData(data string) {
	app.Upload(data)
}

// Upload data to a one-click site: file location or text.
//
func (app *Applet) Upload(data string) {
	app.up.Upload(data)
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
		app.Log().Info(k, v)
	}
}

//
//-----------------------------------------------------------------[ DISPLAY ]--

// Label display callback. One line for each device. Format="eth0: ↓ 42 / ↑ 128".
//
func formatLabel(dev string, in, out uint64) string {
	return fmt.Sprintf("%s: %s %s / %s %s", dev, "↓", bytesize.ByteSize(in), "↑", bytesize.ByteSize(out))
}
