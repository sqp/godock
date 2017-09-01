// Package TVPlay is a UPnP control point for Cairo-Dock.
package TVPlay

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	// "github.com/sqp/gupnp/backendsonos" // go UPnP backend.
	"github.com/sqp/gupnp/backendgupnp" // gupnp backend.

	"github.com/sqp/gupnp"          // UPnP control point.
	"github.com/sqp/gupnp/guigtk"   // UPnP gui.
	"github.com/sqp/gupnp/upnptype" // UPnP common types.

	"github.com/sqp/godock/libs/cdtype" // Applet types.
	"github.com/sqp/godock/libs/ternary"

	"fmt"
)

//
//------------------------------------------------------------------[ APPLET ]--

func init() { cdtype.Applets.Register("TVPlay", NewApplet) }

// Applet defines a dock applet.
//
type Applet struct {
	cdtype.AppBase // Applet base and dock connection.

	conf *appletConf
	cp   upnptype.MediaControl
	gui  *guigtk.TVGui
	win  *gtk.Window
}

// NewApplet creates a new applet instance.
//
func NewApplet(base cdtype.AppBase, events *cdtype.Events) cdtype.AppInstance {
	app := &Applet{AppBase: base}
	app.SetConfig(&app.conf, app.actions()...)

	// Events.
	events.OnClick = func() { // Left click: open and manage the gui window.
		if app.Window().IsOpened() { // Window opened.
			app.Window().ToggleVisibility()
		} else {
			app.createGui(true, true)
		}
	}

	events.OnMiddleClick = func() { app.Action().Launch(app.Action().ID(app.conf.ActionClickMiddle)) }

	events.OnScroll = func(scrollUp bool) {
		var key int
		switch app.conf.ActionMouseWheel {
		case "Change volume":
			key = ternary.Int(scrollUp, int(upnptype.ActionVolumeUp), int(upnptype.ActionVolumeDown))

		case "Seek in track":
			key = ternary.Int(scrollUp, int(upnptype.ActionSeekForward), int(upnptype.ActionSeekBackward))
		}

		app.Action().Launch(key)
	}

	events.OnBuildMenu = app.Action().CallbackMenu(dockMenu...)

	events.End = func() {
		if app.win != nil {
			glib.IdleAdd(app.win.Destroy)
		}
	}

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

	return app
}

// Init load user configuration if needed and initialise applet.
//
func (app *Applet) Init(def *cdtype.Defaults, confLoaded bool) {
	// Defaults.
	def.Commands[0] = cdtype.NewCommand(true, "", app.Name()) // Declare monitoring for our GUI.

	// Create the control window if needed.
	if app.conf.WindowVisibility == 0 {
		if app.win != nil {
			app.Window().Close()
		}
	} else {
		app.createGui(true, app.conf.WindowVisibility == 2)
	}
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
	if app.gui != nil {
		glib.IdleAdd(func() {
			app.Window().SetVisibility(show)
		})
		return
	}

	glib.IdleAdd(func() {
		app.gui, app.win = guigtk.NewGui(app.cp)
		if app.gui == nil {
			return
		}
		app.gui.Load()

		app.win.SetIconFromFile(app.FileLocation("icon")) // TODO: debug  path.Join(localDir, "data/icon.png")
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
//-----------------------------------------------------------------[ ACTIONS ]--

// Define applet actions.
//
func (app *Applet) actions() []*cdtype.Action {
	return []*cdtype.Action{
		{
			ID:   int(upnptype.ActionNone),
			Menu: cdtype.MenuSeparator,
		}, {
			ID:   int(upnptype.ActionToggleMute),
			Name: "Mute volume",
			Icon: "dialog-information",
			Call: func() { app.cp.Action(upnptype.ActionToggleMute) },
		}, {
			ID:   int(upnptype.ActionVolumeDown),
			Name: "Volume down",
			Icon: "go-down",
			Call: func() { app.cp.Action(upnptype.ActionVolumeDown) },
		}, {
			ID:   int(upnptype.ActionVolumeUp),
			Name: "Volume up",
			Icon: "go-up",
			Call: func() { app.cp.Action(upnptype.ActionVolumeUp) },
		}, {
			ID:   int(upnptype.ActionPlayPause),
			Name: "Play / pause",
			Icon: "media-playback-start",
			Call: func() { app.cp.Action(upnptype.ActionPlayPause) },
		}, {
			ID:   int(upnptype.ActionStop),
			Name: "Stop",
			Icon: "media-playback-stop",
			Call: func() { app.cp.Action(upnptype.ActionStop) },
		}, {
			ID:       int(upnptype.ActionSeekBackward),
			Name:     "Seek backward",
			Icon:     "go-previous",
			Call:     func() { app.cp.Action(upnptype.ActionSeekBackward) },
			Threaded: true,
		}, {
			ID:       int(upnptype.ActionSeekForward),
			Name:     "Seek forward",
			Icon:     "go-next",
			Call:     func() { app.cp.Action(upnptype.ActionSeekForward) },
			Threaded: true,
		},
	}
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
