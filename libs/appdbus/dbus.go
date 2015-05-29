/*
Package appdbus is the godock cairo-dock connector using DBus.

Its goal is to connect the main Cairo-Dock Golang API, godock/libs/dock, to its parent.

Actions on the main icon

Examples:
	app.SetQuickInfo("OK")
	app.SetLabel("label changed")
	app.SetIcon("/usr/share/icons/gnome/32x32/actions/media-playback-pause.png")
	app.SetEmblem("/usr/share/icons/gnome/32x32/actions/go-down.png", cdtype.EmblemTopRight)
	app.Animate("fire", 10)
	app.DemandsAttention(true, "default")
	app.ShowDialog("dialog string\n with time in second", 8)

	app.BindShortkey("<Control><Shift>Y", "<Alt>K")
	app.AddDataRenderer("gauge", 2, "Turbo-night-fuel")
	app.RenderValues(0.2, 0.7})

	app.AskText("Enter your name", "<my name>")
	app.AskValue("How many?", 0, 42)
	app.AskQuestion("Why?")

	app.ControlAppli("devhelp")
	app.ShowAppli(true)

	properties, e := app.GetAll()

Add SubIcons

Some of the actions to play with SubIcons:

	app.AddSubIcon(
		"icon 1", "firefox-3.0",      "id1",
		"text 2", "chromium-browser", "id2",
		"1 more", "geany",            "id3",
	)
	app.RemoveSubIcon("id1")

	app.SubIcon("id3").SetQuickInfo("OK")
	app.SubIcon("id2").SetLabel("label changed")
	app.SubIcon("id3").Animate("fire", 3)

Still to do;
	* Icon Actions missing: PopupDialog, AddMenuItems
*/
package appdbus

import (
	"github.com/godbus/dbus"

	"github.com/sqp/godock/libs/cdtype"             // Applets types.
	"github.com/sqp/godock/libs/srvdbus/dbuscommon" // Dbus session.

	"errors"
	"os"
	"strings"
	"time"
)

// Dbus dock paths.
//
const (
	DbusObject             = "org.cairodock.CairoDock"
	DbusInterfaceDock      = "org.cairodock.CairoDock"
	DbusInterfaceApplet    = "org.cairodock.CairoDock.applet"
	DbusInterfaceSubapplet = "org.cairodock.CairoDock.subapplet"
)

// DbusPathDock is the Dbus path to the dock. It depends on the name the dock was started with.
var DbusPathDock dbus.ObjectPath = "/org/cairodock/CairoDock"

//
//------------------------------------------------------------[ START APPLET ]--

// StartApplet will prepare and launch a standalone cairo-dock applet.
// If you have provided events, they will respond when needed, and you have
// nothing more to worry about your applet management.
// It can handle only one poller for now.
//
// List of the steps, and their effect:
//   * Load applet events definition = DefineEvents().
//   * Connect the applet to cairo-dock with DBus. This also activate events callbacks.
//   * Initialise applet with option load config activated = Init(true).
//   * Start and run the polling loop if needed. This start a instant check, and
//     manage regular and manual timer refresh.
//   * Wait for the dock End signal to close the applet.
//
func StartApplet(app cdtype.AppInstance) {
	if app == nil {
		println("dock applet failed to create")
		os.Exit(1)
	}

	// Define and connect events to the dock.
	args := os.Args
	appDir, _ := os.Getwd() // standalone applet, using current dir.
	backend := NewWithApp(app, args, appDir)

	dbusEvent, e := backend.ConnectToBus()
	log := app.Log()
	if log.Err(e, "ConnectToBus") { // Mandatory.
		os.Exit(1)
	}

	// Initialise applet: Load config and apply user settings.
	app.Init(true)

	log.Debug("Applet started")
	defer log.Debug("Applet stopped")

	var waiter <-chan time.Time
	poller := app.Poller()
	if poller != nil {
		poller.Restart() // Check poller directly on start.
		waiter = poller.Wait()
	}

	// Start main loop and handle events until the End signal is received from the dock.
	for {
		select { // Wait for events. Until the End signal is received from the dock.

		case s := <-dbusEvent: // Listen to DBus events.
			if backend.OnSignal(s) {
				return // Signal was stop_module. That's all folks. We're closing.
			}

		case <-waiter: // Wait for the end of the timer. Reloop and check.
			poller.Restart() // recheck poller.
			waiter = poller.Wait()
		}
	}
}

//
//------------------------------------------------------------------[ CDDBUS ]--

// CDDbus is an applet connection to Cairo-Dock using Dbus.
//
type CDDbus struct {
	log cdtype.Logger // Applet logger.

	busPath    dbus.ObjectPath                   // Dbus path to the dock (depends on the program name at launch).
	dbusIcon   *dbus.Object                      // Icon remote actions object.
	dbusSub    *dbus.Object                      // Subicon remote actions object.
	icons      map[string]*SubIcon               // SubIcons index (by ID).
	onEvent    func(string, ...interface{}) bool // Callback to dock.OnEvent to forward.
	dialogCall func(int, interface{})            // Dialog callback action.
	menu       *Menu                             // Opened menu titles and callbacks.
}

// New creates a CDDbus connection.
//
func New(path string) *CDDbus {
	return &CDDbus{
		icons:   make(map[string]*SubIcon),
		busPath: dbus.ObjectPath(path),
		menu:    &Menu{},
	}
}

// NewWithApp creates a CDDbus connection and binds it to an applet instance.
//
func NewWithApp(app cdtype.AppInstance, args []string, dir string) *CDDbus {
	name := args[0][2:] // Strip ./ in the beginning.

	app.SetBase(name, args[3], args[4], dir)
	// app.ParentAppName = args[5]

	backend := New(args[2])
	app.SetBackend(backend)
	app.SetEvents(app)

	backend.log = app.Log()
	return backend
}

// SetOnEvent sets the OnEvent callback to forwards events.
//
func (cda *CDDbus) SetOnEvent(onEvent func(string, ...interface{}) bool) {
	cda.onEvent = onEvent
}

// SubIcon returns the subicon object matching the given key.
//
func (cda *CDDbus) SubIcon(key string) cdtype.IconBase {
	return cda.icons[key]
}

// RemoveSubIcons removes all subicons from the applet. (To be called in init).
//
func (cda *CDDbus) RemoveSubIcons() {
	for icon := range cda.icons { // Remove old subicons.
		cda.RemoveSubIcon(icon)
	}
}

// HaveMonitor gives the state of the monitored application. See cdtype.AppIcon.
//
func (cda *CDDbus) HaveMonitor() (haveApp bool, haveFocus bool) {
	Xid, e := cda.Get("Xid")
	cda.log.Err(e, "Xid")

	if id, ok := Xid.(uint64); ok {
		haveApp = id > 0
	}

	HasFocus, _ := cda.Get("has_focus")
	return haveApp, HasFocus.(bool)
}

//
//------------------------------------------------------------[ DBUS CONNECT ]--

// ConnectToBus connects the applet manager to the dock and register events callbacks.
//
func (cda *CDDbus) ConnectToBus() (<-chan *dbus.Signal, error) {
	conn, c, e := dbuscommon.SessionBus()
	if e != nil {
		close(c)
		return nil, e
	}
	return c, cda.ConnectEvents(conn)
}

// ConnectEvents registers to receive Dbus applet events.
//
func (cda *CDDbus) ConnectEvents(conn *dbus.Conn) error {

	cda.dbusIcon = conn.Object(DbusObject, cda.busPath)
	cda.dbusSub = conn.Object(DbusObject, cda.busPath+"/sub_icons")
	if cda.dbusIcon == nil || cda.dbusSub == nil {
		return errors.New("no Dbus interface")
	}

	// Listen to all events emitted for the icon.
	matchIcon := "type='signal',path='" + string(cda.busPath) + "',interface='" + DbusInterfaceApplet + "',sender='" + DbusObject + "'"
	matchSubs := "type='signal',path='" + string(cda.busPath) + "/sub_icons',interface='" + DbusInterfaceSubapplet + "',sender='" + DbusObject + "'"

	e := conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, matchIcon).Err
	cda.log.Err(e, "connect to icon DBus events")
	e = conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, matchSubs).Err
	cda.log.Err(e, "connect to subicons DBus events")

	return e
}

// OnSignal forward the received signal to the registered event callback.
// Return true if the signal was quit applet.
//
func (cda *CDDbus) OnSignal(s *dbus.Signal) (exit bool) {
	if s == nil {
		return false
	}

	name := strings.TrimPrefix(string(s.Name), DbusInterfaceApplet+".")
	if name != s.Name { // dbus interface matched.
		switch name {
		case "on_answer_dialog": // Callback is already provided.
			if cda.dialogCall != nil && len(s.Body) > 1 {
				value := s.Body[1].(dbus.Variant).Value()
				cda.dialogCall(int(s.Body[0].(int32)), value)
			}
			return false

		case "on_build_menu": // Provide the simple menu builder.
			cda.menu.Clear()
			cda.onEvent(name, cdtype.Menuer(cda.menu))
			cda.AddMenuItems(cda.menu.items...)
			return false

		case "on_menu_select": // Callback is already provided.
			cda.menu.Launch(s.Body[0].(int32))
			return false
		}

		return cda.onEvent(name, s.Body...) // New and old callbacks methods.
	}

	name = strings.TrimPrefix(string(s.Name), DbusInterfaceSubapplet+".")
	if name != s.Name { // dbus subicons interface matched.
		if name == "on_build_menu_sub_icon" { // Provide the simple menu builder.
			cda.menu.Clear()
			cda.onEvent(name, cdtype.Menuer(cda.menu), s.Body[0].(string))
			cda.AddMenuItems(cda.menu.items...)
			return false
		}

		cda.onEvent(name, s.Body...)
		return false
	}

	cda.log.Info("unknown signal", s)
	return false
}

// Call DBus method without returned values.
//
func (cda *CDDbus) launch(iface *dbus.Object, action string, args ...interface{}) error {
	return iface.Call(action, 0, args...).Err
}

func launch(iface *dbus.Object, action string, args ...interface{}) error {
	return iface.Call(action, 0, args...).Err
}

// EavesDrop allow to register to Dbus events for custom parsing.
//
func EavesDrop(match string) (chan *dbus.Message, error) {
	conn, e := dbus.SessionBus()
	if e != nil {
		return nil, e
	}
	e = conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, match).Err
	if e != nil {
		return nil, e
	}
	c := make(chan *dbus.Message, 10)
	conn.Eavesdrop(c)
	return c, nil
}

//
//------------------------------------------------------------[ ICON ACTIONS ]--

// SetQuickInfo change the quickinfo text displayed on the icon.
//
func (cda *CDDbus) SetQuickInfo(info string) error {
	return cda.launch(cda.dbusIcon, DbusInterfaceApplet+".SetQuickInfo", info)
}

// SetLabel change the text label next to the icon.
//
func (cda *CDDbus) SetLabel(label string) error {
	return cda.launch(cda.dbusIcon, DbusInterfaceApplet+".SetLabel", label)
}

// SetIcon set the image of the icon, overwriting the previous one.
// A lot of image formats are supported, including SVG.
// You can refer to the image by either its name if it's an image from a icon theme, or by a path.
//   app.SetIcon("gimp")
//   app.SetIcon("go-up")
//   app.SetIcon("/path/to/image")
//
func (cda *CDDbus) SetIcon(icon string) error {
	return cda.launch(cda.dbusIcon, DbusInterfaceApplet+".SetIcon", icon)
}

// SetEmblem set an emblem image on the icon. To remove it, you have to use
// SetEmblem again with an empty string.
//
//   app.SetEmblem(app.FileLocation("img", "emblem-work.png"), cdtype.EmblemBottomLeft)
//
func (cda *CDDbus) SetEmblem(icon string, position cdtype.EmblemPosition) error {
	return cda.launch(cda.dbusIcon, DbusInterfaceApplet+".SetEmblem", icon, int32(position))
}

// Animate animates the icon, with a given animation and for a given number of
// rounds.
//
func (cda *CDDbus) Animate(animation string, rounds int32) error {
	return cda.launch(cda.dbusIcon, DbusInterfaceApplet+".Animate", animation, rounds)
}

// DemandsAttention is like the Animate method, but will animate the icon
// endlessly, and the icon will be visible even if the dock is hidden. If the
// animation is an empty string, or "default", the animation used when an
// application demands the attention will be used.
// The first argument is true to start animation, or false to stop it.
//
func (cda *CDDbus) DemandsAttention(start bool, animation string) error {
	return cda.launch(cda.dbusIcon, DbusInterfaceApplet+".DemandsAttention", start, animation)
}

// ShowDialog pops up a simple dialog bubble on the icon.
// The dialog can be closed by clicking on it.
//
func (cda *CDDbus) ShowDialog(message string, duration int32) error {
	return cda.dbusIcon.Go(DbusInterfaceApplet+".ShowDialog", 0, nil, message, duration).Err
}

// PopupDialog open a dialog box . See cdtype.AppIcon.
//
// func (cda *CDDbus) PopupDialog(dialog map[string]interface{}, widget map[string]interface{}) error {
func (cda *CDDbus) PopupDialog(data cdtype.DialogData) error {
	dialog := map[string]interface{}{
		"message":     data.Message,
		"icon":        data.Icon,
		"time-length": int32(data.TimeLength),
		"force-above": data.ForceAbove,
		"use-markup":  data.UseMarkup,
		"buttons":     data.Buttons,
	}

	var widget map[string]interface{}
	switch dw := data.Widget.(type) {

	case cdtype.DialogWidgetText:
		widget = map[string]interface{}{
			"widget-type":   "text-entry",
			"multi-lines":   dw.MultiLines,
			"editable":      dw.Editable,
			"visible":       dw.Visible,
			"nb-chars":      int32(dw.NbChars),
			"initial-value": dw.InitialValue,
		}

	case cdtype.DialogWidgetScale:
		widget = map[string]interface{}{
			"widget-type":   "scale",
			"min-value":     dw.MinValue,
			"max-value":     dw.MaxValue,
			"nb-digit":      int32(dw.NbDigit),
			"initial-value": dw.InitialValue,
			"min-label":     dw.MinLabel,
			"max-label":     dw.MaxLabel,
		}

	default:
		widget = make(map[string]interface{})
	}
	cda.dialogCall = data.Callback

	return cda.launch(cda.dbusIcon, DbusInterfaceApplet+".PopupDialog", toMapVariant(dialog), toMapVariant(widget))
}

// AddDataRenderer add a graphic data renderer to the icon.
//
//  Renderer types: gauge, graph, progressbar.
//  Themes for renderer Graph: "Line", "Plain", "Bar", "Circle", "Plain Circle"
//
func (cda *CDDbus) AddDataRenderer(typ string, nbval int32, theme string) error {
	return cda.launch(cda.dbusIcon, DbusInterfaceApplet+".AddDataRenderer", typ, nbval, theme)
}

// RenderValues render new values on the icon.
//   * You must have added a data renderer before with AddDataRenderer.
//   * The number of values sent must match the number declared before.
//   * Values are given between 0 and 1.
//
func (cda *CDDbus) RenderValues(values ...float64) error {
	// return cda.dbusIcon.Call("RenderValues", dbus.FlagNoAutoStart, values).Err
	return cda.launch(cda.dbusIcon, DbusInterfaceApplet+".RenderValues", values)
}

// ActOnAppli send an action on the application controlled by the icon (see ControlAppli).
//
//   "minimize"            to hide the window
//   "show"                to show the window and give it focus
//   "toggle-visibility"   to show or hide
//   "maximize"            to maximize the window
//   "restore"             to restore the window
//   "toggle-size"         to maximize or restore
//   "close"               to close the window (Note: some programs will just hide the window and stay in the systray)
//   "kill"                to kill the X window
//
func (cda *CDDbus) ActOnAppli(action string) error {
	return cda.launch(cda.dbusIcon, DbusInterfaceApplet+".ActOnAppli", action)
}

// ControlAppli allow your applet to control the window of an external
// application and can steal its icon from the Taskbar.
//  *Use the xprop command find the class of the window you want to control.
//  *Use "none" if you want to reset application control.
//  *Controling an application enables the OnFocusChange callback.
//
func (cda *CDDbus) ControlAppli(applicationClass string) error {
	return cda.launch(cda.dbusIcon, DbusInterfaceApplet+".ControlAppli", applicationClass)
}

// ShowAppli set the visible state of the application controlled by the icon.
//
func (cda *CDDbus) ShowAppli(show bool) error {
	return cda.launch(cda.dbusIcon, DbusInterfaceApplet+".ShowAppli", interface{}(show))
}

// AddMenuItems adds a list of items to the menu triggered by OnBuildMenu.
//
func (cda *CDDbus) AddMenuItems(items ...map[string]interface{}) error {
	var data []map[string]dbus.Variant
	for _, interf := range items {
		data = append(data, toMapVariant(interf))
	}

	return cda.launch(cda.dbusIcon, DbusInterfaceApplet+".AddMenuItems", data)
}

// PopulateMenu adds a list of entry to the default menu. An empty string will
// add a separator. Can only be used in the OnBuildMenu callback.
//
// func (cda *CDDbus) PopulateMenu(items ...string) error {
// 	return cda.launch(cda.dbusIcon, DbusInterfaceApplet+".PopulateMenu", items)
// }

// BindShortkey binds one or more keyboard shortcuts to your applet.
//
func (cda *CDDbus) BindShortkey(shortkeys ...cdtype.Shortkey) error {
	var list []string
	for _, sk := range shortkeys {
		// if sk != "" {
		list = append(list, sk.Shortkey)
		// }
	}
	return cda.launch(cda.dbusIcon, DbusInterfaceApplet+".BindShortkey", list)
}

// Get a property of the icon of your applet. Current available properties are :
//   x            int32     x position of the icon's center on the screen (starting from 0 on the left)
//   y            int32     y position of the icon's center on the screen (starting from 0 at the top of the screen)
//   width        int32     width of the icon, in pixels (this is the maximum width, when the icon is zoomed)
//   height       int32     height of the icon, in pixels (this is the maximum height, when the icon is zoomed)
//   container    uint32   type of container of the applet (DOCK, DESKLET)
//   orientation  uint32   position of the container on the screen (BOTTOM, TOP, RIGHT, LEFT). A desklet has always an orientation of BOTTOM.
//   Xid          uint64   ID of the application's window which is controlled by the applet, or 0 if none (this parameter can only be non nul if you used the method ControlAppli beforehand).
//   has_focus    bool     Whether the application's window which is controlled by the applet is the current active window (it has the focus) or not. E.g.:
//
func (cda *CDDbus) Get(property string) (interface{}, error) {
	var v dbus.Variant
	e := cda.dbusIcon.Call("Get", 0, property).Store(&v)
	return v.Value(), e
}

// GetAll returns applet icon properties.
//
func (cda *CDDbus) GetAll() *cdtype.DockProperties {
	vars := make(map[string]dbus.Variant)
	if cda.log.Err(cda.dbusIcon.Call("GetAll", 0).Store(&vars), "dbus GetAll") {
		return nil
	}

	props := &cdtype.DockProperties{}
	for k, v := range vars {
		switch k {
		case "Xid":
			props.Xid = v.Value().(uint64)
		case "x":
			props.X = v.Value().(int32)
		case "y":
			props.Y = v.Value().(int32)
		case "orientation":
			props.Orientation = v.Value().(uint32)
		case "container":
			props.Container = v.Value().(uint32)
		case "width":
			props.Width = v.Value().(int32)
		case "height":
			props.Height = v.Value().(int32)
		case "has_focus":
			props.HasFocus = v.Value().(bool)
		}
	}
	return props
}

//
//--------------------------------------------------------[ SUBICONS ACTIONS ]--

// AddSubIcon adds subicons by pack of 3 strings : label, icon, id.
//
func (cda *CDDbus) AddSubIcon(fields ...string) error {
	for i := 0; i < len(fields)/3; i++ {
		id := fields[3*i+2]
		cda.icons[id] = &SubIcon{cda.dbusSub, id}
	}
	return cda.launch(cda.dbusSub, DbusInterfaceSubapplet+".AddSubIcons", fields)
}

// RemoveSubIcon only need the ID to remove the SubIcon.
//
func (cda *CDDbus) RemoveSubIcon(id string) error {
	if _, ok := cda.icons[id]; !ok {
		return errors.New("RemoveSubIcon Icon missing: " + id)
	}

	e := cda.launch(cda.dbusSub, DbusInterfaceSubapplet+".RemoveSubIcon", id)
	if e == nil {
		delete(cda.icons, id)
	}
	return e
}

// SubIcon defines a connection to the subdock icon.
//
type SubIcon struct {
	dbusSub *dbus.Object
	id      string
}

// SetQuickInfo change the quickinfo text displayed on the subicon.
//
func (cdi *SubIcon) SetQuickInfo(info string) error {
	return launch(cdi.dbusSub, DbusInterfaceSubapplet+".SetQuickInfo", info, cdi.id)
}

// SetLabel change the text label next to the subicon.
//
func (cdi *SubIcon) SetLabel(label string) error {
	return launch(cdi.dbusSub, DbusInterfaceSubapplet+".SetLabel", label, cdi.id)
}

// SetIcon set the image of the subicon, overwriting the previous one. See Icon.
//
func (cdi *SubIcon) SetIcon(icon string) error {
	return launch(cdi.dbusSub, DbusInterfaceSubapplet+".SetIcon", icon, cdi.id)
}

// SetEmblem set an emblem image on the subicon. See Icon.
//
func (cdi *SubIcon) SetEmblem(icon string, position cdtype.EmblemPosition) error {
	return launch(cdi.dbusSub, DbusInterfaceSubapplet+".SetEmblem", icon, int32(position), cdi.id)
}

// Animate animates the subicon, with a given animation and for a given number of
// rounds. See Icon.
//
func (cdi *SubIcon) Animate(animation string, rounds int32) error {
	return launch(cdi.dbusSub, DbusInterfaceSubapplet+".Animate", animation, rounds, cdi.id)
}

// ShowDialog pops up a simple dialog bubble on the subicon. See Icon.
//
func (cdi *SubIcon) ShowDialog(message string, duration int32) error {
	return launch(cdi.dbusSub, DbusInterfaceSubapplet+".ShowDialog", message, duration, cdi.id)
}

//
//------------------------------------------------------------------[ COMMON ]--

// Recast list of args to map[string]dbus.Variant as requested by the DBus API.
//
func toMapVariant(input map[string]interface{}) map[string]dbus.Variant {
	vars := make(map[string]dbus.Variant)
	for k, v := range input {
		vars[k] = dbus.MakeVariant(v)
	}
	return vars
}

//
//-------------------------------------------------------------[ MENU SIMPLE ]--

// Menu is a menu builder to store callbacks at creation to be sure the answer
// match the user request.
//
type Menu struct {
	*MenuData
	MenuID int32
}

// AddEntry adds an item to the menu with its callback.
//
func (menu *Menu) AddEntry(label, iconPath string, call interface{}, userData ...interface{}) cdtype.MenuWidgeter {
	menu.items = append(menu.items, map[string]interface{}{
		"widget-type": cdtype.MenuEntry, //int32(0),
		"label":       label,
		"icon":        iconPath,
		"menu":        menu.MenuID, //int32(0),
		// "id":          int32(1),
		"id": int32(len(menu.actions)),
		// 	"tooltip":     "this is the tooltip that will appear when you hover this entry",
	})

	menu.actions = append(menu.actions, call)
	return nil
}

// Separator adds a separator to the menu.
//
func (menu *Menu) Separator() {
	menu.items = append(menu.items, map[string]interface{}{
		"widget-type": cdtype.MenuSeparator,
		"id":          int32(len(menu.actions)),
	})

	menu.actions = append(menu.actions, nil) // func() {})
}

// SubMenu adds a submenu to the menu.
//
func (menu *Menu) SubMenu(label, iconPath string) cdtype.Menuer {
	menu.items = append(menu.items, map[string]interface{}{
		"widget-type": int32(cdtype.MenuSubMenu),
		"label":       label,
		"icon":        iconPath,
		// "menu":        int32(0),
		"id": int32(len(menu.actions)),
		// 	"tooltip":     "this is the tooltip that will appear when you hover this entry",
	})

	menu.actions = append(menu.actions, nil)
	return &Menu{menu.MenuData, int32(-1)}
	// return &Menu{menu.MenuData, int32(len(menu.actions) - 1)}
}

// AddCheckEntry adds a check entry to the menu.
//
func (menu *Menu) AddCheckEntry(label string, active bool, call interface{}, userData ...interface{}) cdtype.MenuWidgeter {
	menu.items = append(menu.items, map[string]interface{}{
		"widget-type": cdtype.MenuCheckBox,
		"label":       label,
		"state":       active,
		"id":          int32(len(menu.actions)),
	})
	menu.actions = append(menu.actions, call)

	// menu.items = append(menu.items, map[string]interface{}{
	// 	"widget-type": cdtype.MenuRadioButton,
	// 	"label":       "1",
	// 	"state":       true,
	// 	"group":       int32(100),
	// 	"id":          int32(len(menu.actions)),
	// })

	// menu.items = append(menu.items, map[string]interface{}{
	// 	"widget-type": cdtype.MenuRadioButton,
	// 	"label":       "2",
	// 	"group":       int32(100),
	// 	"id":          int32(len(menu.actions) + 1),
	// })

	// menu.actions = append(menu.actions, call)
	// menu.actions = append(menu.actions, call)
	return nil
}

// Launch calls the action referenced by its id.
//
func (menu *Menu) Launch(id int32) {
	if int(id) < len(menu.actions) {
		switch call := menu.actions[id].(type) {
		case func():
			call()
		}
	}
}

// Clear resets the menu items list.
//
func (menu *Menu) Clear() {
	menu.MenuData = &MenuData{}
}

// MenuData stores the menu data to send, and callbacks to launch on return.
//
type MenuData struct {
	actions []interface{} // Menu callbacks are saved to be sure we launch the good action (options can change).
	items   []map[string]interface{}
}
