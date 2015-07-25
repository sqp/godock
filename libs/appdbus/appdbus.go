/*
Package appdbus is the godock cairo-dock connector using DBus.

// Its goal is to connect the main Cairo-Dock Golang applet object,
// godock/libs/cdapplet, to its parent, the dock.
*/
package appdbus

import (
	"github.com/godbus/dbus"

	"github.com/sqp/godock/libs/cdtype"             // Applets types.
	"github.com/sqp/godock/libs/srvdbus/dbuscommon" // Dbus session.
	"github.com/sqp/godock/libs/srvdbus/dockpath"   // Path to main dock dbus service.

	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

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
	dbusIcon   *dbuscommon.Client                // Icon remote actions object.
	dbusSub    *dbuscommon.Client                // Subicon remote actions object.
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
func (cda *CDDbus) ConnectEvents(conn *dbus.Conn) (e error) {
	cda.dbusIcon, e = dbuscommon.GetClient(dockpath.DbusObject, string(cda.busPath), dockpath.DbusInterfaceApplet)
	if e != nil {
		return e
	}

	cda.dbusSub, e = dbuscommon.GetClient(dockpath.DbusObject, string(cda.busPath)+"/sub_icons", dockpath.DbusInterfaceSubapplet)
	if e != nil {
		return e
	}

	if cda.dbusIcon == nil || cda.dbusSub == nil {
		return errors.New("missing Dbus interface")
	}

	// Listen to all events emitted for the icon.
	matchIcon := "type='signal',path='" + string(cda.busPath) + "',interface='" + dockpath.DbusInterfaceApplet + "',sender='" + dockpath.DbusObject + "'"
	matchSubs := "type='signal',path='" + string(cda.busPath) + "/sub_icons',interface='" + dockpath.DbusInterfaceSubapplet + "',sender='" + dockpath.DbusObject + "'"

	e = conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, matchIcon).Err
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

	name := strings.TrimPrefix(string(s.Name), dockpath.DbusInterfaceApplet+".")
	if name != s.Name { // dbus interface matched.
		switch name {
		case "on_click": // Recast to int.
			return cda.onEvent(name, int(s.Body[0].(int32)))

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

	name = strings.TrimPrefix(string(s.Name), dockpath.DbusInterfaceSubapplet+".")
	if name != s.Name { // dbus subicons interface matched.
		switch name {
		case "on_click_sub_icon": // Recast to int.
			return cda.onEvent(name, int(s.Body[0].(int32)), s.Body[1])

		case "on_build_menu_sub_icon": // Provide the simple menu builder.
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

//
//------------------------------------------------------------[ ICON ACTIONS ]--

// SetQuickInfo change the quickinfo text displayed on the icon.
//
func (cda *CDDbus) SetQuickInfo(info string) error {
	return cda.dbusIcon.Call("SetQuickInfo", info)
}

// SetLabel change the text label next to the icon.
//
func (cda *CDDbus) SetLabel(label string) error {
	return cda.dbusIcon.Call("SetLabel", label)
}

// SetIcon set the image of the icon, overwriting the previous one.
//
func (cda *CDDbus) SetIcon(icon string) error {
	return cda.dbusIcon.Call("SetIcon", icon)
}

// SetEmblem set an emblem image on the icon. To remove it, you have to use
// SetEmblem again with an empty string.
//
func (cda *CDDbus) SetEmblem(icon string, position cdtype.EmblemPosition) error {
	return cda.dbusIcon.Call("SetEmblem", icon, int32(position))
}

// Animate animates the icon for a given number of rounds.
//
func (cda *CDDbus) Animate(animation string, rounds int32) error {
	return cda.dbusIcon.Call("Animate", animation, rounds)
}

// DemandsAttention is an endless Animate method.
//
func (cda *CDDbus) DemandsAttention(start bool, animation string) error {
	return cda.dbusIcon.Call("DemandsAttention", start, animation)
}

// ShowDialog pops up a simple dialog bubble on the icon.
//
func (cda *CDDbus) ShowDialog(message string, duration int32) error {
	return cda.dbusIcon.Go(dockpath.DbusInterfaceApplet+".ShowDialog", 0, nil, message, duration).Err
}

// PopupDialog open a dialog box . See cdtype.AppIcon.
//
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

	case cdtype.DialogWidgetList:
		widget = map[string]interface{}{
			"widget-type": "list",
			"editable":    dw.Editable,
			"values":      dw.Values,
		}

		// Recast interface to real type so it won't crash in ToMapVariant.
		switch v := dw.InitialValue.(type) {
		case int32, string:
			widget["initial-value"] = v
			// case int:
			// 	widget["initial-value"] = int32(v)
		}

	default:
		widget = make(map[string]interface{})
	}
	cda.dialogCall = data.Callback

	return cda.dbusIcon.Call("PopupDialog", dbuscommon.ToMapVariant(dialog), dbuscommon.ToMapVariant(widget))
}

// AddMenuItems adds a list of items to the menu triggered by OnBuildMenu.
//
func (cda *CDDbus) AddMenuItems(items ...map[string]interface{}) error {
	var data []map[string]dbus.Variant
	for _, interf := range items {
		data = append(data, dbuscommon.ToMapVariant(interf))
	}

	return cda.dbusIcon.Call("AddMenuItems", data)
}

// BindShortkey binds one or more keyboard shortcuts to your applet.
//
func (cda *CDDbus) BindShortkey(shortkeys ...cdtype.Shortkey) error {
	var list []string
	for _, sk := range shortkeys {
		// if sk != "" {
		list = append(list, sk.Shortkey)
		// }
	}
	return cda.dbusIcon.Call("BindShortkey", list)
}

//
//-----------------------------------------------------------[ DATA RENDERER ]--

// DataRenderer manages the graphic data renderer of the icon.
//
func (cda *CDDbus) DataRenderer() cdtype.IconRenderer {
	return &dataRend{icon: cda}
}

// datarend implements cdtype.IconRenderer.
//
type dataRend struct {
	icon *CDDbus
}

func (o *dataRend) Gauge(nbval int, themeName string) error {
	return o.icon.dbusIcon.Call("AddDataRenderer", "gauge", int32(nbval), themeName)
}

func (o *dataRend) Graph(nbval int, typ cdtype.RendererGraphType) error {
	return o.icon.dbusIcon.Call("AddDataRenderer", "graph", int32(nbval), strconv.Itoa(int(typ)))
}

func (o *dataRend) Progress(nbval int) error {
	return o.icon.dbusIcon.Call("AddDataRenderer", "progress", int32(nbval), "")
}

func (o *dataRend) Remove() error {
	return o.icon.dbusIcon.Call("AddDataRenderer", "", int32(0), "")
}

func (o *dataRend) Render(values ...float64) error {
	return o.icon.dbusIcon.Call("RenderValues", values)
}

func (o *dataRend) GraphLine(nb int) error        { return o.Graph(nb, cdtype.RendererGraphLine) }
func (o *dataRend) GraphPlain(nb int) error       { return o.Graph(nb, cdtype.RendererGraphPlain) }
func (o *dataRend) GraphBar(nb int) error         { return o.Graph(nb, cdtype.RendererGraphBar) }
func (o *dataRend) GraphCircle(nb int) error      { return o.Graph(nb, cdtype.RendererGraphCircle) }
func (o *dataRend) GraphPlainCircle(nb int) error { return o.Graph(nb, cdtype.RendererGraphPlainCircle) }

//
//----------------------------------------------------------[ WINDOW ACTIONS ]--

// Window gives access to actions on the controlled window.
//
func (cda *CDDbus) Window() cdtype.IconWindow { return &winAction{icon: cda} }

// winAction implements cdtype.IconWindow
//
type winAction struct {
	icon *CDDbus
}

func (o *winAction) SetAppliClass(applicationClass string) error {
	return o.icon.dbusIcon.Call("ControlAppli", applicationClass)
}

// actOnAppli sends an action to the application controlled by the icon.
//
func (o *winAction) actOnAppli(action string) error {
	return o.icon.dbusIcon.Call("ActOnAppli", action)
}

func (o *winAction) IsOpened() bool          { xid, _ := o.icon.IconProperty().Xid(); return xid > 0 }
func (o *winAction) Minimize() error         { return o.actOnAppli("minimize") }
func (o *winAction) Show() error             { return o.actOnAppli("show") }
func (o *winAction) ToggleVisibility() error { return o.actOnAppli("toggle-visibility") }
func (o *winAction) Maximize() error         { return o.actOnAppli("maximize") }
func (o *winAction) Restore() error          { return o.actOnAppli("restore") }
func (o *winAction) ToggleSize() error       { return o.actOnAppli("toggle-size") }
func (o *winAction) Close() error            { return o.actOnAppli("close") }
func (o *winAction) Kill() error             { return o.actOnAppli("kill") }

func (o *winAction) SetVisibility(show bool) error {
	return o.icon.dbusIcon.Call("ShowAppli", interface{}(show))
}

//
//---------------------------------------------------------[ SINGLE PROPERTY ]--

// IconProperty returns applet icon properties one by one.
//
func (cda *CDDbus) IconProperty() cdtype.IconProperty {
	return &iconProp{*cda}
}

// iconProp returns icon properties one by one, implements cdtype.IconProperty
type iconProp struct {
	CDDbus
}

func (o *iconProp) get(property string) (interface{}, error) {
	var v dbus.Variant
	e := o.dbusIcon.Object.Call("Get", 0, property).Store(&v)
	return v.Value(), e
}

func (o *iconProp) X() (int, error) {
	v, e := o.get("x")
	if e != nil {
		return 0, e
	}
	return int(v.(int32)), e
}

func (o *iconProp) Y() (int, error) {
	v, e := o.get("y")
	if e != nil {
		return 0, e
	}
	return int(v.(int32)), e
}

func (o *iconProp) Width() (int, error) {
	v, e := o.get("width")
	if e != nil {
		return 0, e
	}
	return int(v.(int32)), e
}

func (o *iconProp) Height() (int, error) {
	v, e := o.get("height")
	if e != nil {
		return 0, e
	}
	return int(v.(int32)), e
}

func (o *iconProp) ContainerPosition() (cdtype.ContainerPosition, error) {
	v, e := o.get("orientation")
	if e != nil {
		return 0, e
	}
	return cdtype.ContainerPosition(v.(uint32)), e
}

func (o *iconProp) ContainerType() (cdtype.ContainerType, error) {
	v, e := o.get("container")
	if e != nil {
		return cdtype.ContainerUnknown, e
	}
	return cdtype.ContainerType(v.(uint32) + 1), e // +1 as we have an unknown as 0 in this version.
}

func (o *iconProp) Xid() (uint64, error) {
	v, e := o.get("Xid")
	if e != nil {
		return 0, e
	}
	return v.(uint64), e
}

func (o *iconProp) HasFocus() (bool, error) {
	v, e := o.get("has_focus")
	if e != nil {
		return false, e
	}
	return v.(bool), e
}

//
//----------------------------------------------------------[ ALL PROPERTIES ]--

// IconProperties returns all applet icon properties at once.
//
func (cda *CDDbus) IconProperties() (cdtype.IconProperties, error) {
	vars := make(map[string]dbus.Variant)
	e := cda.dbusIcon.Object.Call("GetAll", 0).Store(&vars)
	if e != nil {
		return nil, e
	}

	props := &iconProps{}
	for k, v := range vars {
		switch k {
		case "Xid":
			props.xid = v.Value().(uint64)
		case "x":
			props.x = int(v.Value().(int32))
		case "y":
			props.y = int(v.Value().(int32))
		case "orientation":
			props.containerPosition = cdtype.ContainerPosition(v.Value().(uint32))
		case "container":
			props.containerType = cdtype.ContainerType(v.Value().(uint32) + 1) // +1 as we have an unknown as 0 in this version.
		case "width":
			props.width = int(v.Value().(int32))
		case "height":
			props.height = int(v.Value().(int32))
		case "has_focus":
			props.hasFocus = v.Value().(bool)
		}
	}
	return props, nil
}

// iconProps returns all icon properties at once, implements cdtype.IconProperties
//
type iconProps struct {
	x      int // Distance from the left of the screen.
	y      int // Distance from the bottom of the screen.
	width  int // Width of the icon.
	height int // Height of the icon.

	containerPosition cdtype.ContainerPosition // bottom, top, right, left.
	containerType     cdtype.ContainerType     // Dock, desklet...

	xid      uint64 // Xid of the monitored window. Value > 0 if a window is monitored.
	hasFocus bool   // True if the monitored window has the cursor focus.
}

func (o *iconProps) X() int                                      { return o.x }
func (o *iconProps) Y() int                                      { return o.y }
func (o *iconProps) Width() int                                  { return o.width }
func (o *iconProps) Height() int                                 { return o.height }
func (o *iconProps) Xid() uint64                                 { return o.xid }
func (o *iconProps) HasFocus() bool                              { return o.hasFocus }
func (o *iconProps) ContainerPosition() cdtype.ContainerPosition { return o.containerPosition }
func (o *iconProps) ContainerType() cdtype.ContainerType         { return o.containerType }

//
//--------------------------------------------------------[ SUBICONS ACTIONS ]--

// AddSubIcon adds subicons by pack of 3 strings : label, icon, id.
//
func (cda *CDDbus) AddSubIcon(fields ...string) error {
	for i := 0; i < len(fields)/3; i++ {
		id := fields[3*i+2]
		cda.icons[id] = &SubIcon{cda.dbusSub, id}
	}
	return cda.dbusSub.Call("AddSubIcons", fields)
}

// RemoveSubIcon only need the ID to remove the SubIcon.
//
func (cda *CDDbus) RemoveSubIcon(id string) error {
	if _, ok := cda.icons[id]; !ok {
		return errors.New("RemoveSubIcon Icon missing: " + id)
	}

	e := cda.dbusSub.Call("RemoveSubIcon", id)
	if e == nil {
		delete(cda.icons, id)
	}
	return e
}

// SubIcon defines a connection to the subdock icon.
//
type SubIcon struct {
	dbusSub *dbuscommon.Client
	id      string
}

// SetQuickInfo change the quickinfo text displayed on the subicon.
//
func (cdi *SubIcon) SetQuickInfo(info string) error {
	return cdi.dbusSub.Call("SetQuickInfo", info, cdi.id)
}

// SetLabel change the text label next to the subicon.
//
func (cdi *SubIcon) SetLabel(label string) error {
	return cdi.dbusSub.Call("SetLabel", label, cdi.id)
}

// SetIcon set the image of the subicon, overwriting the previous one. See Icon.
//
func (cdi *SubIcon) SetIcon(icon string) error {
	return cdi.dbusSub.Call("SetIcon", icon, cdi.id)
}

// SetEmblem set an emblem image on the subicon. See Icon.
//
func (cdi *SubIcon) SetEmblem(icon string, position cdtype.EmblemPosition) error {
	return cdi.dbusSub.Call("SetEmblem", icon, int32(position), cdi.id)
}

// Animate animates the subicon, with a given animation and for a given number of
// rounds. See Icon.
//
func (cdi *SubIcon) Animate(animation string, rounds int32) error {
	return cdi.dbusSub.Call("Animate", animation, rounds, cdi.id)
}

// ShowDialog pops up a simple dialog bubble on the subicon. See Icon.
//
func (cdi *SubIcon) ShowDialog(message string, duration int32) error {
	return cdi.dbusSub.Call("ShowDialog", message, duration, cdi.id)
}

//
//-------------------------------------------------------------[ MENU SIMPLE ]--

// MenuData stores the menu data to send, and callbacks to launch on return.
//
type MenuData struct {
	actions []interface{} // Menu callbacks are saved to be sure we launch the good action (options can change).
	items   []map[string]interface{}
}

// Menu is a menu builder, storing callbacks at creation to be sure the answer
// match the user request.
//
type Menu struct {
	*MenuData
	MenuID int32
}

// AddEntry adds an item to the menu with its callback.
//
func (menu *Menu) AddEntry(label, iconPath string, call interface{}, userData ...interface{}) cdtype.MenuWidgeter {
	return menu.addOne(call, map[string]interface{}{
		"type":  cdtype.MenuEntry,
		"label": label,
		"icon":  iconPath,
		"menu":  menu.MenuID,
		"id":    int32(len(menu.actions)),
	})
}

// AddSeparator adds a separator to the menu.
//
func (menu *Menu) AddSeparator() {
	menu.addOne(nil, map[string]interface{}{
		"type": cdtype.MenuSeparator,
		"menu": menu.MenuID,
		"id":   int32(len(menu.actions)),
	})
}

// AddSubMenu adds a submenu to the menu.
//
// TODO: test if first entry (ID=0) really can't be a submenu.
//
func (menu *Menu) AddSubMenu(label, iconPath string) cdtype.Menuer {
	menu.addOne(nil, map[string]interface{}{
		"type":  cdtype.MenuSubMenu,
		"label": label,
		"icon":  iconPath,
		"menu":  menu.MenuID, //      int32(0),
		"id":    int32(len(menu.actions)),
	})

	return &Menu{
		MenuData: menu.MenuData,
		MenuID:   int32(len(menu.actions) - 1), // MenuID is current item ID. Can't be 0 (main menu).
	}
}

// AddCheckEntry adds a check entry to the menu.
//
func (menu *Menu) AddCheckEntry(label string, active bool, call interface{}, userData ...interface{}) cdtype.MenuWidgeter {
	return menu.addOne(call, map[string]interface{}{
		"type":  cdtype.MenuCheckBox,
		"label": label,
		"menu":  menu.MenuID,
		"state": active,
		"id":    int32(len(menu.actions)),
	})
}

// AddRadioEntry adds a radio entry to the menu.
//
func (menu *Menu) AddRadioEntry(label string, active bool, group int, call interface{}, userData ...interface{}) cdtype.MenuWidgeter {
	return menu.addOne(call, map[string]interface{}{
		"type":  cdtype.MenuRadioButton,
		"label": label,
		"menu":  menu.MenuID,
		"state": active,
		"group": int32(group),
		"id":    int32(len(menu.actions)),
	})
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

func (menu *Menu) addOne(call interface{}, item map[string]interface{}) cdtype.MenuWidgeter {
	menu.items = append(menu.items, item)
	menu.actions = append(menu.actions, call)
	return tooltiper(item)
}

//
//---------------------------------------------------------------[ TOOLTIPER ]--

// tooltiper provides a MenuWidgeter interface to set more menu options.
//
type tooltiper map[string]interface{}

func (tt tooltiper) SetTooltipText(str string) {
	tt["tooltip"] = str
}
