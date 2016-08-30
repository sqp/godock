// Package NetActivity is a monitoring, upload and download applet for Cairo-Dock.
/*

Improvements since original DropToShare version:
  Not using temp files.
  Many new upload sites.
  Code simple and maintainable (400 lines for 18 backends).

Dependencies:
  xsel or xclip command for clipboard interaction when build without gtk.

Not implemented (yet):
  Icon for the applet.
  More menu options.
  Save image copy (and display).
  Custom upload scripts.

*/
package NetActivity

import (
	"github.com/sqp/godock/libs/cdglobal"      // Global consts.
	"github.com/sqp/godock/libs/cdtype"        // Applet types.
	"github.com/sqp/godock/libs/clipboard"     // Set clipboard content.
	"github.com/sqp/godock/libs/net/uptoshare" // Uploader service.
	"github.com/sqp/godock/libs/net/videodl"   // Video downloader service.
	"github.com/sqp/godock/libs/sysinfo"       // IOActivity.
	"github.com/sqp/godock/libs/text/bytesize" // Human readable bytes.

	"fmt"
	"strings"
)

//
//------------------------------------------------------------------[ APPLET ]--

func init() { cdtype.Applets.Register("NetActivity", NewApplet) }

// Applet defines a dock applet.
//
type Applet struct {
	cdtype.AppBase // Applet base and dock connection.

	conf    *appletConf
	service *sysinfo.IOActivity
	up      *uptoshare.Uploader
	video   videodl.Downloader
}

// NewApplet creates a new applet instance.
//
func NewApplet(base cdtype.AppBase, events *cdtype.Events) cdtype.AppInstance {
	app := &Applet{AppBase: base}
	app.SetConfig(&app.conf)

	// Events.
	events.OnClick = app.Command().Callback(cmdLeft)
	events.OnMiddleClick = app.Command().Callback(cmdMiddle)
	events.OnBuildMenu = app.buildMenu
	events.End = func() { app.video.WebUnregister() }

	events.OnDropData = func(data string) {
		if strings.HasPrefix(data, "http://") || strings.HasPrefix(data, "https://") {
			if app.conf.VideoDLEnabled {
				app.DownloadVideo(data)
			}
		} else {
			app.UpToShareUpload(data)
		}
	}

	// Uptoshare actions
	app.up = uptoshare.New()
	app.up.Log = app.Log()
	app.up.SetPreCheck(func() { app.SetEmblem(app.FileLocation("icon"), EmblemAction) })
	app.up.SetPostCheck(func() { app.SetEmblem("none", EmblemAction) })
	app.up.SetOnResult(app.onUploadDone)

	// Network activity actions.
	app.service = sysinfo.NewIOActivity(app)
	app.service.Log = app.Log()
	app.service.FormatIcon = sysinfo.FormatIcon
	app.service.FormatLabel = formatLabel
	app.service.GetData = sysinfo.GetNetActivity

	app.Poller().Add(app.service.Check)

	// Video download actions.
	ActionsVideoDL := 0

	hist := videodl.NewHistoryVideo(app, videodl.HistoryFile)
	app.video = videodl.NewManager(app, app.Log(), hist)

	app.video.SetPreCheck(func() error { return app.SetEmblem(app.FileLocation("img", "go-down.svg"), EmblemDownload) })
	app.video.SetPostCheck(func() error { return app.SetEmblem("none", EmblemDownload) })
	app.video.Actions(ActionsVideoDL, app.Action().Add)

	app.video.WebRegister()

	hist.Load()

	return app
}

// Init load user configuration if needed and initialise applet.
//
func (app *Applet) Init(def *cdtype.Defaults, confLoaded bool) {
	// Uptoshare settings.
	app.up.SetHistoryFile(app.FileDataDir(cdglobal.DirUserAppData, uptoshare.HistoryFile))
	app.up.SetHistorySize(app.conf.UploadHistory)
	app.up.LimitRate = app.conf.UploadRateLimit
	app.up.PostAnonymous = app.conf.PostAnonymous
	app.up.FileForAll = app.conf.FileForAll
	app.up.SiteFile(app.conf.SiteFile)
	app.up.SiteImage(app.conf.SiteImage)
	app.up.SiteText(app.conf.SiteText)
	app.up.SiteVideo(app.conf.SiteVideo)

	// Video download settings.
	app.video.SetConfig(&app.conf.Config)
	app.video.SetEnabledWeb(videodl.WebState(app.conf.EnabledWeb))

	// Settings for poller and IOActivity (force renderer reset in case of reload).
	app.service.Settings(uint64(app.conf.UpdateDelay.Value()), app.conf.DisplayText,
		app.conf.DisplayValues, app.conf.GraphType, app.conf.GaugeName, app.conf.Devices...)

	// Defaults.
	def.PollerInterval = app.conf.UpdateDelay.Value()
	def.Commands = cdtype.Commands{
		cmdLeft:   cdtype.NewCommandStd(app.conf.LeftAction, app.conf.LeftCommand, app.conf.LeftClass),
		cmdMiddle: cdtype.NewCommandStd(app.conf.MiddleAction, app.conf.MiddleCommand),
	}
}

//
//--------------------------------------------------------------------[ MENU ]--

// buildMenu fills the menu with left and middle click actions if they're set.
//
func (app *Applet) buildMenu(menu cdtype.Menuer) {
	needSep := false
	if app.conf.LeftAction > 0 && app.conf.LeftCommand != "" {
		menu.AddEntry("Action left click", "system-run", app.Command().Callback(cmdLeft))
		needSep = true
	}
	if app.conf.MiddleAction > 0 && app.conf.MiddleCommand != "" {
		menu.AddEntry("Action middle click", "system-run", app.Command().Callback(cmdMiddle))
		needSep = true
	}
	if needSep {
		menu.AddSeparator()
	}
	subup := menu.AddSubMenu("Upload", "")
	for _, hist := range app.up.ListHistory() {
		hist := hist
		subup.AddEntry(hist["file"], "", func() {
			app.Log().Info(hist["link"])
			clipboard.Write(hist["link"])
			// app.ShowDialog(link, 5)
		})
	}

	menu.AddSeparator()
	app.video.Menu(menu)
}

//
//-----------------------------------------------------------[ PUBLIC REMOTE ]--

// UpToShareUpload uploads data to a one-click site: file location or text.
//
func (app *Applet) UpToShareUpload(data string) {
	go app.up.Upload(data)
}

// UpToShareLastLink uploads data to a one-click site: file location or text.
//
func (app *Applet) UpToShareLastLink() string {
	hists := app.up.ListHistory()
	if len(hists) == 0 {
		return ""
	}
	return hists[0]["link"]
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

	if msg, ok := links["support"]; ok {
		app.Log().Info("support message", msg)
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
