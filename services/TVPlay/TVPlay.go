// Package TVPlay is a UPnP control point for the Cairo-Dock project.
package TVPlay

import (
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/gupnp/mediacp"        // upnp control point.
	"github.com/sqp/gupnp/mediacp/guigtk" // upnp gui.
	"github.com/sqp/gupnp/upnpcp"

	"github.com/sqp/godock/libs/dock" // Connection to cairo-dock.
	"github.com/sqp/godock/libs/ternary"
)

// Applet data and controlers.
//
type Applet struct {
	*dock.CDApplet
	conf *appletConf
	cp   *mediacp.MediaControl
	gui  *guigtk.TVGui
	win  *gtk.Window
}

// NewApplet creates a new TVPlay applet instance.
//
func NewApplet() dock.AppletInstance {
	app := &Applet{CDApplet: dock.NewCDApplet()} // Icon controler and interface to cairo-dock.
	app.defineActions()

	var e error
	app.cp, e = mediacp.New()
	app.Log.Err(e, "temp Dir")

	app.gui, app.win = guigtk.NewGui(app.cp, false)
	if app.gui != nil {
		app.gui.Load()
	}

	// func (gui *TVGui) ConnectControl() {
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
	app.SetDefaults(dock.Defaults{
		// Label: "",
		Icon: app.conf.Icon,
		Shortkeys: []string{app.conf.ShortkeyMute, app.conf.ShortkeyVolumeDown, app.conf.ShortkeyVolumeUp,
			app.conf.ShortkeyPlayPause, app.conf.ShortkeyStop,
			app.conf.ShortkeySeekBackward, app.conf.ShortkeySeekForward},
		// Templates:      []string{"default"},
		Commands: dock.Commands{
			"left": dock.NewCommand(true, "", app.AppletName)},
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
	app.Log.Info("Renderer Found", r.Name, "", r.Udn)
}

func (app *Applet) onMediaServerFound(srv *upnpcp.Server) {
	app.Log.Info("Server Found", srv.Name, "", srv.Udn)
}

func (app *Applet) onMediaRendererLost(r *upnpcp.Renderer) {
	app.Log.Info("Renderer Lost", r.Name, "", r.Udn)
}

func (app *Applet) onMediaServerLost(srv *upnpcp.Server) {
	app.Log.Info("Server Lost", srv.Name, "", srv.Udn)
}

//
//------------------------------------------------------------------[ EVENTS ]--

// DefineEvents set applet events callbacks.
//
func (app *Applet) DefineEvents() {

	// Left click: open and manage the gui window.
	//
	app.Events.OnClick = func() {
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
	app.Events.OnMiddleClick = func() {
		app.Actions.Launch(app.Actions.ID(app.conf.ActionClickMiddle))
	}

	app.Events.OnScroll = func(scrollUp bool) {
		var key int
		switch app.conf.ActionMouseWheel {
		case "Change volume":
			key = ternary.Int(scrollUp, mediacp.ActionVolumeUp, mediacp.ActionVolumeDown)

		case "Seek in track":
			key = ternary.Int(scrollUp, mediacp.ActionSeekForward, mediacp.ActionSeekBackward)
		}

		app.Actions.Launch(key)
	}

	app.Events.OnBuildMenu = func() {
		app.BuildMenu(dockMenu)
	}

	// Menu entry selected. Launch the expected action.
	//
	app.Events.OnMenuSelect = func(numEntry int32) {
		app.Actions.Launch(dockMenu[numEntry])
	}

	app.Events.OnShortkey = func(key string) {
		switch key {
		case app.conf.ShortkeyMute:
			app.Actions.Launch(mediacp.ActionToggleMute)

		case app.conf.ShortkeyVolumeDown:
			app.Actions.Launch(mediacp.ActionVolumeDown)

		case app.conf.ShortkeyVolumeUp:
			app.Actions.Launch(mediacp.ActionVolumeUp)

		case app.conf.ShortkeyPlayPause:
			app.Actions.Launch(mediacp.ActionPlayPause)

		case app.conf.ShortkeyStop:
			app.Actions.Launch(mediacp.ActionStop)

		case app.conf.ShortkeySeekBackward:
			app.Actions.Launch(mediacp.ActionSeekBackward)

		case app.conf.ShortkeySeekForward:
			app.Actions.Launch(mediacp.ActionSeekForward)
		}
	}
}

//
//-----------------------------------------------------------------[ ACTIONS ]--

// Define applet actions.
//
func (app *Applet) defineActions() {
	app.Actions.Add(
		&dock.Action{
			ID: mediacp.ActionNone,
			// Icontype: 2,
		},
		&dock.Action{
			ID:   mediacp.ActionToggleMute,
			Name: "Mute volume",
			Icon: "gtk-dialog-info",
			Call: func() { app.cp.Action(mediacp.ActionToggleMute) },
		},
		&dock.Action{
			ID:   mediacp.ActionVolumeDown,
			Name: "Volume down",
			Icon: "gtk-go-down",
			Call: func() { app.cp.Action(mediacp.ActionVolumeDown) },
		},
		&dock.Action{
			ID:   mediacp.ActionVolumeUp,
			Name: "Volume up",
			Icon: "gtk-go-up",
			Call: func() { app.cp.Action(mediacp.ActionVolumeUp) },
		},
		&dock.Action{
			ID:   mediacp.ActionPlayPause,
			Name: "Play / pause",
			Icon: "gtk-media-play",
			Call: func() { app.cp.Action(mediacp.ActionPlayPause) },
		},
		&dock.Action{
			ID:   mediacp.ActionStop,
			Name: "Stop",
			Icon: "gtk-media-stop",
			Call: func() { app.cp.Action(mediacp.ActionStop) },
		},
		&dock.Action{
			ID:       mediacp.ActionSeekBackward,
			Name:     "Seek backward",
			Icon:     "gtk-go-backward",
			Call:     func() { app.cp.Action(mediacp.ActionSeekBackward) },
			Threaded: true,
		},
		&dock.Action{
			ID:       mediacp.ActionSeekForward,
			Name:     "Seek forward",
			Icon:     "gtk-go-forward",
			Call:     func() { app.cp.Action(mediacp.ActionSeekForward) },
			Threaded: true,
		},
	)
}
