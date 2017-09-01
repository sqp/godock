// Package AppTmpl is a new template applet for Cairo-Dock.
package AppTmpl

import "github.com/sqp/godock/libs/cdtype" // Applet types.

//
//------------------------------------------------------------------[ APPLET ]--

func init() { cdtype.Applets.Register("AppTmpl", NewApplet) }

// Applet defines a dock applet.
//
type Applet struct {
	cdtype.AppBase // Applet base and dock connection.

	conf *appletConf
}

// NewApplet creates a new applet instance.
//
func NewApplet(base cdtype.AppBase, events *cdtype.Events) cdtype.AppInstance {
	app := &Applet{AppBase: base}
	app.SetConfig(&app.conf)

	// Events.
	events.OnClick = app.onClick
	events.OnMiddleClick = app.onMiddleClick
	events.OnScroll = app.onScroll
	events.OnBuildMenu = app.onBuildMenu
	events.OnSubClick = app.onSubClick
	events.OnSubMiddleClick = app.onSubMiddleClick
	events.OnSubScroll = app.onSubScroll
	events.OnSubBuildMenu = app.onSubBuildMenu

	// https://godoc.org/github.com/sqp/godock/libs/cdtype#Events
	//
	// events.OnDropData = onDropData
	// events.OnChangeFocus = onChangeFocus
	// events.PreInit = onPreInit
	// events.Reload = onReload
	// events.End = onEnd
	// events.OnSubDropData = onSubDropData
	// events.OnClickMod = onClickMod
	// events.OnSubClickMod = onSubClickMod
	//
	// https://godoc.org/github.com/sqp/godock/libs/cdtype#pkg-examples

	return app
}

// Init load user configuration if needed and initialise applet.
//
func (app *Applet) Init(def *cdtype.Defaults, confLoaded bool) {

	// Defaults
	def.Commands[cmdLeft] = cdtype.NewCommand(app.conf.LeftAction == 1, app.conf.LeftCommand, app.conf.LeftClass)

	// It's often better to reset and readd your sub icons during init if they aren't constant.
	// app.RemoveSubIcons()
}

//
//-------------------------------------------------------------[ DOCK EVENTS ]--

func (app *Applet) onClick() {
	switch app.conf.LeftAction {
	case 1:
		if app.conf.LeftCommand != "" {
			app.Command().Launch(cmdLeft)
		}
	}
}

func (app *Applet) onMiddleClick() {
	switch app.conf.MiddleAction {
	case 1: // TODO: need more actions and constants to define them.
		app.Log().Info("MiddleClick")
	}
}

func (app *Applet) onScroll(up bool) {
	app.Log().Info("Scroll")
}

// onBuildMenu fills the menu with device actions: mute, mixer, select device.
//
func (app *Applet) onBuildMenu(menu cdtype.Menuer) { // device actions menu: mute, mixer, select device.
	menu.AddCheckEntry("checkbox", true || false, func() {})
	if app.conf.LeftCommand != "" {
		menu.AddEntry("Open command", "multimedia-volume-control", app.Command().Callback(cmdLeft))
	}
}

func (app *Applet) onSubClick(icon string) {
}

func (app *Applet) onSubMiddleClick(icon string) {
}

func (app *Applet) onSubScroll(icon string, up bool) {
}

// onSubBuildMenu fills the menu with stream actions: select device.
//
func (app *Applet) onSubBuildMenu(icon string, menu cdtype.Menuer) { // stream actions menu: select device.
	menu.AddCheckEntry("checkbox", true || false, func() {
	})
}

// func (app *Applet) onDropData(data string)                  {}
// func (app *Applet) onChangeFocus(active bool)               {}
// func (app *Applet) onPreInit(loadConf bool)                 {}
// func (app *Applet) onReload(bool)                           {}
// func (app *Applet) onEnd()                                  {}
// func (app *Applet) onSubDropData(icon string, data string)  {}
// func (app *Applet) onClickMod(btnState int)                 {}
// func (app *Applet) onSubClickMod(icon string, btnState int) {}

// openMixer opens the mixer if found.
//
func (app *Applet) openMixer() {
	if app.conf.LeftCommand != "" {
		app.Command().Launch(cmdLeft)
	}
}
