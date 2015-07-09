// Package TVPlay is a UPnP control point for Cairo-Dock.
package TVPlay

import (
	"github.com/conformal/gotk3/glib"
	"github.com/conformal/gotk3/gtk"

	// "github.com/sqp/gupnp/backendsonos" // go UPnP backend.
	"github.com/sqp/gupnp/backendgupnp" // gupnp backend.

	"github.com/sqp/gupnp"          // UPnP control point.
	"github.com/sqp/gupnp/guigtk"   // UPnP gui.
	"github.com/sqp/gupnp/upnptype" // UPnP common types.

	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/dock" // Connection to cairo-dock.
	"github.com/sqp/godock/libs/ternary"

	"fmt"
)

// Applet data and controlers.
//
type Applet struct {
	cdtype.AppBase // Applet base and dock connection.

	conf *appletConf
	cp   upnptype.MediaControl
	gui  *guigtk.TVGui
	win  *gtk.Window
}

// NewApplet creates a new TVPlay applet instance.
//
func NewApplet() cdtype.AppInstance {
	app := &Applet{AppBase: dock.NewCDApplet()} // Icon controler and interface to cairo-dock.
	app.defineActions()

	// Create the UPnP device manager.
	var e error
	app.cp, e = gupnp.New(&logger{app.Log()})
	app.Log().Err(e, "temp Dir")

	// Connect local tests.
	hook := app.cp.SubscribeHook("applet")
	hook.OnRendererFound = app.onMediaRendererFound
	hook.OnServerFound = app.onMediaServerFound
	hook.OnRendererLost = app.onMediaRendererLost
	hook.OnServerLost = app.onMediaServerLost

	// hook.OnRendererSelected = gui.SetRenderer
	// hook.OnServerSelected = gui.SetServer

	// hook.OnTransportState = func(r upnptype.Renderer, state upnptype.PlaybackState) { gui.SetPlaybackState(state) }
	// hook.OnCurrentTrackDuration = func(r upnptype.Renderer, dur int) { gui.SetDuration(mediacp.TimeToString(dur)) }
	// hook.OnCurrentTrackMetaData = func(r upnptype.Renderer, item upnptype.Item) { gui.SetTitle(item.Title) }
	// hook.OnMute = func(r upnptype.Renderer, muted bool) { gui.SetMuted(muted) }
	// hook.OnVolume = func(r upnptype.Renderer, vol uint) { gui.SetVolume(int(vol)) }
	// hook.OnCurrentTime = func(r upnptype.Renderer, secs int, f float64) { gui.SetCurrentTime(secs, f*100) }
	// hook.OnSetVolumeDelta = func(delta int) { gui.SetVolumeDelta(delta) }
	// }

	// Connect an UPnP backend to the manager.
	// mgr := backendsonos.NewManager(&logger{app.Log()})
	// mgr.SetEvents(app.cp.DefineEvents())
	// go mgr.Start(true)

	cp := backendgupnp.NewControlPoint()
	cp.SetEvents(app.cp.DefineEvents())

	// Create the control window.
	// guigtk.WindowTitle = "Test"
	app.createGui(false, true)

	return app
}

// Init load user configuration if needed and initialise applet.
//
func (app *Applet) Init(loadConf bool) {
	app.LoadConfig(loadConf, &app.conf) // Load config will crash if fail. Expected.

	if app.win != nil {
		app.win.SetIconFromFile(app.FileLocation("icon")) // TODO: debug  path.Join(localDir, "data/icon.png")
	}

	// Set defaults to dock icon: display and controls.
	app.SetDefaults(cdtype.Defaults{
		Icon: app.conf.Icon,
		Commands: cdtype.Commands{
			"left": cdtype.NewCommand(true, "", app.Name())},
		Shortkeys: []cdtype.Shortkey{
			{"Actions", "ShortkeyMute", "Mute volume", app.conf.ShortkeyMute},
			{"Actions", "ShortkeyVolumeDown", "Lower volume", app.conf.ShortkeyVolumeDown},
			{"Actions", "ShortkeyVolumeUp", "Increase volume", app.conf.ShortkeyVolumeUp},
			{"Actions", "ShortkeyPlayPause", "Play / Pause", app.conf.ShortkeyPlayPause},
			{"Actions", "ShortkeyStop", "Stop", app.conf.ShortkeyStop},
			{"Actions", "ShortkeySeekBackward", "Seek backward", app.conf.ShortkeySeekBackward},
			{"Actions", "ShortkeySeekForward", "Seek forward", app.conf.ShortkeySeekForward}},
		Debug: app.conf.Debug})

	app.cpInit()
}

// initialise the control point.
//
func (app *Applet) cpInit() {
	app.cp.SetVolumeDelta(app.conf.VolumeDelta)
	app.cp.SetSeekDelta(app.conf.SeekDelta)
	app.cp.SetPreferredRenderer(app.conf.PreferredRenderer)
	app.cp.SetPreferredServer(app.conf.PreferredServer)
}

func (app *Applet) createGui(init, show bool) {
	glib.IdleAdd(func() {
		app.gui, app.win = guigtk.NewGui(app.cp)
		if app.gui == nil {
			return
		}
		app.gui.Load()

		app.win.Connect("delete-event", func() bool { app.gui, app.win = nil, nil; return false })
		// app.win.Connect("delete-event", func() bool { window.Iconify(); return true })

		if init {
			app.cpInit()
		}
		if !show {
			app.win.Iconify()
		}
	})
}

func (app *Applet) onMediaRendererFound(r upnptype.Renderer) {
	app.Log().Info("Renderer Found", r.Name(), "", r.UDN())
}

func (app *Applet) onMediaServerFound(srv upnptype.Server) {
	app.Log().Info("Server Found", srv.Name(), "", srv.UDN())
}

func (app *Applet) onMediaRendererLost(r upnptype.Renderer) {
	app.Log().Info("Renderer Lost", r.Name(), "", r.UDN())
}

func (app *Applet) onMediaServerLost(srv upnptype.Server) {
	app.Log().Info("Server Lost", srv.Name(), "", srv.UDN())
}

//
//------------------------------------------------------------------[ EVENTS ]--

// DefineEvents set applet events callbacks.
//
func (app *Applet) DefineEvents(events *cdtype.Events) {

	// Left click: open and manage the gui window.
	//
	events.OnClick = func() {
		haveMonitor, hasFocus := app.HaveMonitor()
		if haveMonitor { // Window opened.
			app.ShowAppli(!hasFocus)
		} else {
			app.createGui(true, true)
		}
	}

	// Middle click: launch configured action.
	//
	events.OnMiddleClick = func() {
		app.ActionLaunch(app.ActionID(app.conf.ActionClickMiddle))
	}

	events.OnScroll = func(scrollUp bool) {
		var key int
		switch app.conf.ActionMouseWheel {
		case "Change volume":
			key = ternary.Int(scrollUp, int(upnptype.ActionVolumeUp), int(upnptype.ActionVolumeDown))

		case "Seek in track":
			key = ternary.Int(scrollUp, int(upnptype.ActionSeekForward), int(upnptype.ActionSeekBackward))
		}

		app.ActionLaunch(key)
	}

	events.OnBuildMenu = func(menu cdtype.Menuer) {
		app.BuildMenu(menu, dockMenu)
	}

	events.OnShortkey = func(key string) {
		switch key {
		case app.conf.ShortkeyMute:
			app.ActionLaunch(int(upnptype.ActionToggleMute))

		case app.conf.ShortkeyVolumeDown:
			app.ActionLaunch(int(upnptype.ActionVolumeDown))

		case app.conf.ShortkeyVolumeUp:
			app.ActionLaunch(int(upnptype.ActionVolumeUp))

		case app.conf.ShortkeyPlayPause:
			app.ActionLaunch(int(upnptype.ActionPlayPause))

		case app.conf.ShortkeyStop:
			app.ActionLaunch(int(upnptype.ActionStop))

		case app.conf.ShortkeySeekBackward:
			app.ActionLaunch(int(upnptype.ActionSeekBackward))

		case app.conf.ShortkeySeekForward:
			app.ActionLaunch(int(upnptype.ActionSeekForward))
		}
	}
}

//
//-----------------------------------------------------------------[ ACTIONS ]--

// Define applet actions.
//
func (app *Applet) defineActions() {
	app.ActionAdd(
		&cdtype.Action{
			ID:   int(upnptype.ActionNone),
			Menu: cdtype.MenuSeparator,
		},
		&cdtype.Action{
			ID:   int(upnptype.ActionToggleMute),
			Name: "Mute volume",
			Icon: "dialog-information",
			Call: func() { app.cp.Action(upnptype.ActionToggleMute) },
		},
		&cdtype.Action{
			ID:   int(upnptype.ActionVolumeDown),
			Name: "Volume down",
			Icon: "go-down",
			Call: func() { app.cp.Action(upnptype.ActionVolumeDown) },
		},
		&cdtype.Action{
			ID:   int(upnptype.ActionVolumeUp),
			Name: "Volume up",
			Icon: "go-up",
			Call: func() { app.cp.Action(upnptype.ActionVolumeUp) },
		},
		&cdtype.Action{
			ID:   int(upnptype.ActionPlayPause),
			Name: "Play / pause",
			Icon: "media-playback-start",
			Call: func() { app.cp.Action(upnptype.ActionPlayPause) },
		},
		&cdtype.Action{
			ID:   int(upnptype.ActionStop),
			Name: "Stop",
			Icon: "media-playback-stop",
			Call: func() { app.cp.Action(upnptype.ActionStop) },
		},
		&cdtype.Action{
			ID:       int(upnptype.ActionSeekBackward),
			Name:     "Seek backward",
			Icon:     "go-previous",
			Call:     func() { app.cp.Action(upnptype.ActionSeekBackward) },
			Threaded: true,
		},
		&cdtype.Action{
			ID:       int(upnptype.ActionSeekForward),
			Name:     "Seek forward",
			Icon:     "go-next",
			Call:     func() { app.cp.Action(upnptype.ActionSeekForward) },
			Threaded: true,
		},
	)
}

//
//------------------------------------------------------------------[ LOGGER ]--

type logger struct{ cdtype.Logger }

func (l *logger) Infof(pattern string, args ...interface{}) {
	str := fmt.Sprintf(pattern, args...)
	l.Info("WTF", str)
}

func (l *logger) Warningf(pattern string, args ...interface{}) {
	str := fmt.Sprintf(pattern, args...)
	l.Info("WTF", str)
}
