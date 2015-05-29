// Package TVPlay is a UPnP control point for Cairo-Dock.
package TVPlay

import (
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/gupnp/mediacp"        // upnp control point.
	"github.com/sqp/gupnp/mediacp/guigtk" // upnp gui.
	"github.com/sqp/gupnp/upnpcp"

	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/dock" // Connection to cairo-dock.
	"github.com/sqp/godock/libs/ternary"
)

// Applet data and controlers.
//
type Applet struct {
	cdtype.AppBase // Applet base and dock connection.

	conf *appletConf
	cp   *mediacp.MediaControl
	gui  *guigtk.TVGui
	win  *gtk.Window
}

// NewApplet creates a new TVPlay applet instance.
//
func NewApplet() cdtype.AppInstance {
	app := &Applet{AppBase: dock.NewCDApplet()} // Icon controler and interface to cairo-dock.
	app.defineActions()

	var e error
	app.cp, e = mediacp.New()
	app.Log().Err(e, "temp Dir")

	app.gui, app.win = guigtk.NewGui(app.cp, false)
	if app.gui != nil {
		app.gui.Load()
	}

	hook := app.cp.SubscribeHook("applet")
	hook.OnRendererFound = app.onMediaRendererFound
	hook.OnServerFound = app.onMediaServerFound
	hook.OnRendererLost = app.onMediaRendererLost
	hook.OnServerLost = app.onMediaServerLost

	// hook.OnRendererSelected = gui.SetRenderer
	// hook.OnServerSelected = gui.SetServer

	// hook.OnTransportState = func(r *upnpcp.Renderer, state upnpcp.PlaybackState) { gui.SetPlaybackState(state) }
	// hook.OnCurrentTrackDuration = func(r *upnpcp.Renderer, dur int) { gui.SetDuration(mediacp.TimeToString(dur)) }
	// hook.OnCurrentTrackMetaData = func(r *upnpcp.Renderer, item *upnpcp.Item) { gui.SetTitle(item.Title) }
	// hook.OnMute = func(r *upnpcp.Renderer, muted bool) { gui.SetMuted(muted) }
	// hook.OnVolume = func(r *upnpcp.Renderer, vol uint) { gui.SetVolume(int(vol)) }
	// hook.OnCurrentTime = func(r *upnpcp.Renderer, secs int, f float64) { gui.SetCurrentTime(secs, f*100) }
	// hook.OnSetVolumeDelta = func(delta int) { gui.SetVolumeDelta(delta) }
	// }

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

func (app *Applet) onMediaRendererFound(r *upnpcp.Renderer) {
	app.Log().Info("Renderer Found", r.Name, "", r.Udn)
}

func (app *Applet) onMediaServerFound(srv *upnpcp.Server) {
	app.Log().Info("Server Found", srv.Name, "", srv.Udn)
}

func (app *Applet) onMediaRendererLost(r *upnpcp.Renderer) {
	app.Log().Info("Renderer Lost", r.Name, "", r.Udn)
}

func (app *Applet) onMediaServerLost(srv *upnpcp.Server) {
	app.Log().Info("Server Lost", srv.Name, "", srv.Udn)
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
			// guigtk.NewGui(app.cp, true).Load() // shouldn't be reached ATM.
			// app.cpInit()
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
			key = ternary.Int(scrollUp, mediacp.ActionVolumeUp, mediacp.ActionVolumeDown)

		case "Seek in track":
			key = ternary.Int(scrollUp, mediacp.ActionSeekForward, mediacp.ActionSeekBackward)
		}

		app.ActionLaunch(key)
	}

	events.OnBuildMenu = func(menu cdtype.Menuer) {
		app.BuildMenu(menu, dockMenu)
	}

	events.OnShortkey = func(key string) {
		switch key {
		case app.conf.ShortkeyMute:
			app.ActionLaunch(mediacp.ActionToggleMute)

		case app.conf.ShortkeyVolumeDown:
			app.ActionLaunch(mediacp.ActionVolumeDown)

		case app.conf.ShortkeyVolumeUp:
			app.ActionLaunch(mediacp.ActionVolumeUp)

		case app.conf.ShortkeyPlayPause:
			app.ActionLaunch(mediacp.ActionPlayPause)

		case app.conf.ShortkeyStop:
			app.ActionLaunch(mediacp.ActionStop)

		case app.conf.ShortkeySeekBackward:
			app.ActionLaunch(mediacp.ActionSeekBackward)

		case app.conf.ShortkeySeekForward:
			app.ActionLaunch(mediacp.ActionSeekForward)
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
			ID:   mediacp.ActionNone,
			Menu: cdtype.MenuSeparator,
		},
		&cdtype.Action{
			ID:   mediacp.ActionToggleMute,
			Name: "Mute volume",
			Icon: "dialog-information",
			Call: func() { app.cp.Action(mediacp.ActionToggleMute) },
		},
		&cdtype.Action{
			ID:   mediacp.ActionVolumeDown,
			Name: "Volume down",
			Icon: "go-down",
			Call: func() { app.cp.Action(mediacp.ActionVolumeDown) },
		},
		&cdtype.Action{
			ID:   mediacp.ActionVolumeUp,
			Name: "Volume up",
			Icon: "go-up",
			Call: func() { app.cp.Action(mediacp.ActionVolumeUp) },
		},
		&cdtype.Action{
			ID:   mediacp.ActionPlayPause,
			Name: "Play / pause",
			Icon: "media-playback-start",
			Call: func() { app.cp.Action(mediacp.ActionPlayPause) },
		},
		&cdtype.Action{
			ID:   mediacp.ActionStop,
			Name: "Stop",
			Icon: "media-playback-stop",
			Call: func() { app.cp.Action(mediacp.ActionStop) },
		},
		&cdtype.Action{
			ID:       mediacp.ActionSeekBackward,
			Name:     "Seek backward",
			Icon:     "go-previous",
			Call:     func() { app.cp.Action(mediacp.ActionSeekBackward) },
			Threaded: true,
		},
		&cdtype.Action{
			ID:       mediacp.ActionSeekForward,
			Name:     "Seek forward",
			Icon:     "go-next",
			Call:     func() { app.cp.Action(mediacp.ActionSeekForward) },
			Threaded: true,
		},
	)
}
