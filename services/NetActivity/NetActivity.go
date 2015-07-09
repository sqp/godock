// Package NetActivity is a monitoring and upload applet for Cairo-Dock.
/*

Improvements since original DropToShare version:
  Not using temp files.
  New image upload sites:
    - http://pix.toile-libre.org
    - http://postimage.org
  Code simple and maintainable (350 lines for 12 backends).

Dependencies:
  glib-2.0

  xsel or xclip command for clipboard interaction when build without gtk.

  curl is needed for those backends:
    Text:  none
    Image: all
    Video: all
    File:  all

Not implemented (yet):
  Icon for the applet.
  Upload raw text with FileForAll option. I'm trying to find a way to do it
    without the temp file option before falling back to this method.
  More menu options, due to lack of proper AddMenuItem method.
  Save image copy (and display).
  Custom upload scripts.
  Url shortener (as I'm not fan of those, you better do it yourself if you want it).

*/
package NetActivity

import (
	"github.com/sqp/godock/libs/cdtype"        // Applets types.
	"github.com/sqp/godock/libs/clipboard"     // Set clipboard content.
	"github.com/sqp/godock/libs/dock"          // Connection to cairo-dock.
	"github.com/sqp/godock/libs/net/uptoshare" // Uploader service.
	"github.com/sqp/godock/libs/net/videodl"   // Video downloader service.
	"github.com/sqp/godock/libs/sysinfo"       // IOActivity.
	"github.com/sqp/godock/libs/text/bytesize" // Human readable bytes.

	"fmt"
	"strings"
)

const (
	// EmblemAction is the position of the "upload in progress" emblem.
	EmblemAction = cdtype.EmblemTopRight

	// EmblemDownload is the position of the "download in progress" emblem.
	EmblemDownload = cdtype.EmblemTopLeft
)

//
//------------------------------------------------------------------[ APPLET ]--

// Applet data and controlers.
//
type Applet struct {
	cdtype.AppBase // Applet base and dock connection.

	conf    *appletConf
	service *sysinfo.IOActivity
	up      *uptoshare.Uploader
	video   *videodl.Manager
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
	app.up.SetOnResult(app.onUploadDone)

	// Video download actions.
	ActionsVideoDL := 0
	app.video = videodl.NewManager(app, app.Log())
	app.video.SetBackend(videodl.BackendYoutubeDL)
	app.video.SetPreCheck(func() error { return app.SetEmblem(app.FileLocation("img", "go-down.svg"), EmblemDownload) })
	app.video.SetPostCheck(func() error { return app.SetEmblem("none", EmblemDownload) })
	// app.up.SetOnResult(app.onUploadDone)
	app.video.Actions(ActionsVideoDL, app.ActionAdd)

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

	// Video download settings.
	app.video.SetPath(app.conf.VideoDLPath)
	app.video.SetQuality(videodl.Quality(app.conf.VideoDLQuality))
	app.video.SetBlacklist(app.conf.VideoDLBlacklist)

	app.Log().Info("init", app.video.Path)

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
		menu.AddEntry("Action left click", "system-run", app.OnClick)
		needSep = true
	}
	if app.conf.MiddleAction > 0 && app.conf.MiddleCommand != "" {
		menu.AddEntry("Action middle click", "system-run", app.OnMiddleClick)
		needSep = true
	}
	if needSep {
		menu.Separator()
	}
	for _, hist := range app.up.ListHistory() {
		hist := hist
		menu.AddEntry(hist["file"], "", func() {
			app.Log().Info(hist["link"])
			clipboard.Write(hist["link"])
			// app.ShowDialog(link, 5)
		})
	}

	// app.video.MenuQuality(menu)
}

// OnDropData uploads file(s) or text dropped on the icon.
//
func (app *Applet) OnDropData(data string) {
	if strings.HasPrefix(data, "http://") || strings.HasPrefix(data, "https://") {
		if app.conf.VideoDLEnabled {
			app.DownloadVideo(data)
		}
	} else {
		app.Upload(data)
	}
}

// DefineEvents set applet events callbacks.
//
func (app *Applet) DefineEvents(events *cdtype.Events) {
	events.OnBuildMenu = app.video.MenuQuality
}

//
//-----------------------------------------------------------------[ SERVICE ]--

// Upload uploads data to a one-click site: file location or text.
//
func (app *Applet) Upload(data string) {
	app.up.Upload(data)
}

// DownloadVideo downloads the video from url.
//
func (app *Applet) DownloadVideo(data string) {
	app.video.Download(data)
}

//
//----------------------------------------------------------------[ CALLBACK ]--

// onUploadDone is called with the list of links when an item has been uploaded.
//
func (app *Applet) onUploadDone(links uptoshare.Links) {
	if e, ok := links["error"]; ok {
		app.ShowDialog("Error: "+e, 10)
		return
	}

	if app.conf.DialogEnabled {
		if link, ok := links["link"]; ok {
			clipboard.Write(link)
			app.ShowDialog(link, app.conf.DialogDuration)
		}
	}

	for k, v := range links { // TODO: to improve.
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
