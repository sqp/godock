/*
Package dbus is the godock cairo-dock connector using DBus. Its goal is to connect the
main Cairo-Dock Golang API, godock/libs/dock, to its parent.
*/
package dbus

import (
	"github.com/guelfey/go.dbus"

	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/log"

	"errors"
	// "reflect"
)

const (
	DbusObject             = "org.cairodock.CairoDock"
	DbusPathDock           = "/org/cairodock/CairoDock"
	DbusInterfaceDock      = "org.cairodock.CairoDock"
	DbusInterfaceApplet    = "org.cairodock.CairoDock.applet"
	DbusInterfaceSubapplet = "org.cairodock.CairoDock.subapplet"
)

type CdDbus struct {
	Icons     map[string]*SubIcon
	Close     chan bool // will receive true when the applet is closed.
	Events    cdtype.Events
	SubEvents cdtype.SubEvents

	BusPath dbus.ObjectPath

	eavesDropMatch string              // Special key to filter events from other Dbus provider.
	eavesDropCall  func(*dbus.Message) // Callback when a message is matched.

	// private data
	// dbusDock *dbus.Object
	dbusIcon *dbus.Object
	dbusSub  *dbus.Object
}

func New(path string) *CdDbus {
	return &CdDbus{
		Icons: make(map[string]*SubIcon),
		Close: make(chan bool),

		BusPath: dbus.ObjectPath(path),
	}
}

//------------------------------------------------------------[ DBUS CONNECT ]--

func (cda *CdDbus) GetCloseChan() chan bool {
	return cda.Close
}

// Connect the applet manager to the Cairo-Dock core. Saves interfaces to the
// icon and subicon DBus interfaces and connects events callbacks.
//
func (cda *CdDbus) ConnectToBus() (<-chan *dbus.Signal, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		log.Info("DBus Connect", err)
		return nil, err
	}

	// if e = cda.dbus.Authenticate(); e != nil {
	// 	log.Info("Failed Connection.Authenticate:", e.Error())
	// 	return e
	// }

	// cda.dbusDock = conn.Object(DbusObject, DbusPathDock)
	cda.dbusIcon = conn.Object(DbusObject, cda.BusPath)
	cda.dbusSub = conn.Object(DbusObject, cda.BusPath+"/sub_icons")
	if cda.dbusIcon == nil || cda.dbusSub == nil {
		return nil, errors.New("No DBus interface")
	}

	// Listen to all events emitted for the icon.
	e := conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
		"type='signal',path='"+string(cda.BusPath)+"',interface='"+DbusInterfaceApplet+"',sender='"+DbusObject+"'").Err
	log.Err(e, "Connect to Icon DBus events")

	e = conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
		"type='signal',path='"+string(cda.BusPath)+"/sub_icons',interface='"+DbusInterfaceSubapplet+"',sender='"+DbusObject+"'").Err
	log.Err(e, "Connect to Subicons DBus events")

	c := make(chan *dbus.Signal, 10)
	conn.Signal(c)
	// go cda.pull(c)
	return c, e

	// if cda.eavesDropMatch != "" { // Nothing to EavesDrop, just get our signals.
	// 	if e := conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, cda.eavesDropMatch).Err; !log.Err(e, "DBus AddMatch") {
	// 		c := make(chan *dbus.Message, 10)
	// 		conn.Eavesdrop(c)
	// 		go cda.pullEaves(c)
	// 	}
	// }
	// return nil
}

// func (cda *CdDbus) pull(c chan *dbus.Signal) {
// 	for v := range c {
// 		cda.OnSignal(v)
// 	}
// }

func (cda *CdDbus) OnSignal(v *dbus.Signal) {
	// log.DEV("received", v)
	switch {
	case v.Name[len(DbusInterfaceApplet)] == '.':
		// log.DEV("Received", v.Name[len(DbusInterfaceApplet)+1:], v.Body)
		cda.receivedMainEvent(v.Name[len(DbusInterfaceApplet)+1:], v.Body)

	case v.Name[len(DbusInterfaceSubapplet)] == '.':
		// log.DEV("SUBICON", v.Name[len(DbusInterfaceSubapplet)+1:], v.Body)
		cda.receivedSubEvent(v.Name[len(DbusInterfaceSubapplet)+1:], v.Body)
	}
}

//------------------------------------------------------------[ TEST ]--

// func (cda *CdDbus) pullEaves(c chan *dbus.Message) {
// 	for msg := range c {
// 		switch msg.Type {
// 		case dbus.TypeSignal:
// 			log.Info("signal")
// 			cda.OnSignal(msg.ToSignal())

// 		case dbus.TypeMethodCall:
// 			log.Info("method")
// 			go cda.eavesDropCall(msg)
// 		}
// 	}
// }

func (cda *CdDbus) EavesDrop(match string, call func(*dbus.Message)) {
	cda.eavesDropMatch = match
	cda.eavesDropCall = call
}

// Call DBus method without returned values.
//
func (cda *CdDbus) launch(iface *dbus.Object, action string, args ...interface{}) error {
	return iface.Call(action, 0, args...).Err
}

func launch(iface *dbus.Object, action string, args ...interface{}) error {
	return iface.Call(action, 0, args...).Err
}

//
//------------------------------------------------------------[ ICON ACTIONS ]--

// Sets the quickinfo text displayed on our icon.
//
func (cda *CdDbus) SetQuickInfo(info string) error {
	return cda.launch(cda.dbusIcon, DbusInterfaceApplet+".SetQuickInfo", info)
}

// Sets the text label of our icon.
func (cda *CdDbus) SetLabel(label string) error {
	return cda.launch(cda.dbusIcon, DbusInterfaceApplet+".SetLabel", label)
}

// Sets the image of our icon, overwriting the previous one.
// You can refer to the image by either its name if it's an image from a icon theme, or by a path.
//   app.SetIcon("gimp")
//   app.SetIcon("gtk-go-up")
//   app.SetIcon("/path/to/image")
//
func (cda *CdDbus) SetIcon(icon string) error {
	return cda.launch(cda.dbusIcon, DbusInterfaceApplet+".SetIcon", icon)
}

// Sets an emblem on our icon. The emblem is drawn directly on the icon, so if you want to remove it, you have to use SetIcon with the original image.
//   The image is given by its path
//   See cdtype.EmblemPosition for valid emblem locations.
//
//   app.SetEmblem("./emblem-charged.png", cdtype.EmblemBottomLeft)
//
func (cda *CdDbus) SetEmblem(icon string, position cdtype.EmblemPosition) error {
	return cda.launch(cda.dbusIcon, DbusInterfaceApplet+".SetEmblem", icon, int32(position))
}

func (cda *CdDbus) Animate(animation string, rounds int32) error {
	return cda.launch(cda.dbusIcon, DbusInterfaceApplet+".Animate", animation, rounds)
}

func (cda *CdDbus) ShowDialog(message string, duration int32) error {
	return cda.dbusIcon.Go(DbusInterfaceApplet+".ShowDialog", 0, nil, message, duration).Err
}

func (cda *CdDbus) DemandsAttention(start bool, animation string) error {
	return cda.launch(cda.dbusIcon, DbusInterfaceApplet+".DemandsAttention", start, animation)
}

// Pops up a dialog box . The dialog can contain a message, an icon, some buttons, and a widget the user can act on.
// Adding buttons will trigger an on_answer_dialog signal when the user press one of them.
// "ok" and "cancel" are used as keywords for the defined by the dock.
//
// Dialog attributes:
//   message        string    dialog box text (default=empty).
//   icon           string    icon displayed next to the message (default=applet icon).
//   time-length    bool      duration of the dialog, in second (default=unlimited).
//   force-above    bool      true to force the dialog above. Use it with parcimony (default=false)
//   use-markup     bool      true to use Pango markup to add text decorations (default=false).
//   buttons        string    images of the buttons, separated by comma ";" (default=none).
//
// Widget attributes:
//   type          string    type of the widget: "text-entry" or "scale" or "list".
//
// Widget text-entry attributes:
//   multi-lines    bool      true to have a multi-lines text-entry, ie a text-view (default=false).
//   editable       bool      whether the user can modify the text or not (default=true).
//   visible        bool      whether the text will be visible or not (useful to type passwords) (default=true).
//   nb-chars       int32     maximum number of chars (the current number of chars will be displayed next to the entry) (default=infinite).
//   initial-value  string    text initially contained in the entry (default=empty).
//
// Widget scale attributes:
//   min-value      double    lower value (default=0).
//   max-value      double    upper value (default=100).
//   nb-digit       int32     number of digits after the dot (default=2).
//   initial-value  double    value initially set to the scale (default=0).
//   min-label      string    label displayed on the left of the scale (default=empty).
//   max-label      string    label displayed on the right of the scale (default=empty).
//
// Widget list attributes:
//   editable       bool      true if a non-existing choice can be entered by the user (in this case, the content of the widget will be the selected text, and not the number of the selected line) (false by default)
//   values         string    a list of values, separated by comma ";", used to fill the combo list.
//   initial-value  string or int32 depending on the "editable" attribute :
//        case editable=true:   string with the default text for the user entry of the widget (default=empty).
//        case editable=false:  int with the selected line number (default=0).
//
func (cda *CdDbus) PopupDialog(dialog map[string]interface{}, widget map[string]interface{}) error {
	return cda.launch(cda.dbusIcon, DbusInterfaceApplet+".PopupDialog", toMapVariant(dialog), toMapVariant(widget))
}

// Renderer types: gauge, graph, progressbar
// Themes for renderer Graph: "Line", "Plain", "Bar", "Circle", "Plain Circle"
//
func (cda *CdDbus) AddDataRenderer(typ string, nbval int32, theme string) error {
	return cda.launch(cda.dbusIcon, DbusInterfaceApplet+".AddDataRenderer", typ, nbval, theme)
}

//
//
func (cda *CdDbus) RenderValues(values ...float64) error {
	return cda.dbusIcon.Call("RenderValues", dbus.FlagNoAutoStart, values).Err
	// return cda.launch(cda.dbusIcon, DbusInterfaceApplet+".RenderValues", values)
}

// Send an action on the application controlled by the icon (see ControlAppli)
//
// "minimize"            to hide the window
// "show"                to show the window and give it focus
// "toggle-visibility"   to show or hide
// "maximize"            to maximize the window
// "restore"             to restore the window
// "toggle-size"         to maximize or restore
// "close"               to close the window (Note: some programs will just hide the window and stay in the systray)
// "kill"                to kill the X window
//
func (cda *CdDbus) ActOnAppli(action string) error {
	return cda.launch(cda.dbusIcon, DbusInterfaceApplet+".ActOnAppli", action)
}

// Makes your applet control the window of an external application. Steals its
// icon from the Taskbar. Use the xprop command find the class of the window you
// want to control. Use "none" if you want to reset application control.
// Controling an application enables the OnFocusChange callback.
//
func (cda *CdDbus) ControlAppli(applicationClass string) error {
	return cda.launch(cda.dbusIcon, DbusInterfaceApplet+".ControlAppli", applicationClass)
}

// Set the visible state of the application controlled by the icon.
//
func (cda *CdDbus) ShowAppli(show bool) error {
	return cda.launch(cda.dbusIcon, DbusInterfaceApplet+".ShowAppli", interface{}(show))
}

// aa{sv}

//~ func (cda *CdDbus) AddMenuItems(items... []map[string]interface{}) error {
func (cda *CdDbus) AddMenuItems() error {
	// menuitem := []map[string]interface{}{
	// 	// {
	// 	{"widget-type": cdtype.MenuEntry, //int32(0),
	// 		"label": "this is an entry of the main menu",
	// 		"icon":  "gtk-add",
	// 		// "menu":    int32(0),
	// 		// "id":      int32(1),
	// 		// "tooltip": "this is the tooltip that will appear when you hover this entry",
	// 	},
	// 	// },
	// }

	// menuitem := []map[string]interface{}{
	// 	"label": {
	// 		{"widget-type", 0},
	// 		{"label", "this is an entry of the main menu"},
	// 		{"icon", "gtk-add"},
	// 		{"menu", 0},
	// 		{"id", 1},
	// 		{"tooltip", "this is the tooltip that will appear when you hover this entry"},
	// 	},
	// }

	// var data []map[string]dbus.Variant
	// for _, interf := range menuitem {
	// 	// for _, interf2 := range interf {
	// 	data = append(data, toMapVariant(interf))
	// 	// data = append(data, append([]map[string]dbus.Variant{}, toMapVariant(interf)))
	// 	// }
	// }

	data := []map[string]dbus.Variant{
		{
		// "widget-type": dbus.MakeVariant(int32(0)),
		// // "widget-type": dbus.MakeVariant(int32(cdtype.MenuEntry)),
		// "label": dbus.MakeVariant("entry"),
		// // "icon":  dbus.MakeVariant("gtk-add"),
		// // "menu":    int32(0),
		// "menu": dbus.MakeVariant(int32(0)),

		// "id": dbus.MakeVariant(int32(1)),
		// // "id":      int32(1),
		// // "tooltip": "this is the tooltip that will appear when you hover this entry",
		},
	}

	// icon := map[string]dbus.Variant{

	// "widget-type": dbus.MakeVariant(int32(cdtype.MenuEntry)),
	// "label":       dbus.MakeVariant("this is an entry of the main menu"),
	// "icon":  dbus.MakeVariant("gtk-add"),
	// "menu":    int32(0),
	// "id":      int32(1),
	// "tooltip": "this is the tooltip that will appear when you hover this entry",
	// }

	log.DETAIL(data)
	// return cda.launch(cda.dbusIcon, DbusInterfaceApplet+"AddMenuItems", []interface{}{menuitem})
	go log.Err(cda.dbusIcon.Call(DbusInterfaceApplet+".AddMenuItems", dbus.FlagNoReplyExpected, data).Err, "additems")
	log.Info("grr")
	return nil
}

func (cda *CdDbus) PopulateMenu(items ...string) error {
	return cda.launch(cda.dbusIcon, DbusInterfaceApplet+".PopulateMenu", items)
}

// Bind one or more keyboard shortcuts to your applet. Only non empty shortkeys
// will be sent to the dock so you can use this method to directly add them from
// config.
//
func (cda *CdDbus) BindShortkey(shortkeys ...string) error {
	return cda.launch(cda.dbusIcon, DbusInterfaceApplet+".BindShortkey", shortkeys)
}

func (cda *CdDbus) AskText(message, initialText string) error {
	return cda.dbusIcon.Call("AskText", 0, message, initialText).Err
}

func (cda *CdDbus) AskValue(message string, initialValue, maxValue float64) error {
	return cda.dbusIcon.Call("AskValue", 0, message, initialValue, maxValue).Err
}

func (cda *CdDbus) AskQuestion(message string) error {
	return cda.dbusIcon.Call("AskQuestion", 0, message).Err
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
func (cda *CdDbus) Get(property string) (interface{}, error) {
	var v dbus.Variant
	e := cda.dbusIcon.Call("Get", 0, property).Store(&v)
	return v.Value(), e
}

// Get Module Icon Properties.
//
func (cda *CdDbus) GetAll() *cdtype.DockProperties {
	vars := make(map[string]dbus.Variant)
	if log.Err(cda.dbusIcon.Call("GetAll", 0).Store(&vars), "dbus GetAll") {
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

// SubIcons actions.
//
type SubIcon struct {
	dbusSub *dbus.Object
	id      string
}

// Add subicons by pack of 3 string : label, icon, id.
//
func (cda *CdDbus) AddSubIcon(fields ...string) error {
	for i := 0; i < len(fields)/3; i++ {
		id := fields[3*i+2]
		cda.Icons[id] = &SubIcon{cda.dbusSub, id}
	}
	return cda.launch(cda.dbusSub, DbusInterfaceSubapplet+".AddSubIcons", fields)
}

func (cda *CdDbus) RemoveSubIcon(id string) error {
	if _, ok := cda.Icons[id]; ok {
		e := cda.launch(cda.dbusSub, DbusInterfaceSubapplet+".RemoveSubIcon", id)
		if e == nil {
			delete(cda.Icons, id)
		}
		return e
	}
	return errors.New("RemoveSubIcon Icon missing: " + id)
}

func (cdi *SubIcon) SetQuickInfo(info string) error {
	return launch(cdi.dbusSub, DbusInterfaceSubapplet+".SetQuickInfo", info, cdi.id)
}

func (cdi *SubIcon) SetLabel(label string) error {
	return launch(cdi.dbusSub, DbusInterfaceSubapplet+".SetLabel", label, cdi.id)
}

func (cdi *SubIcon) SetIcon(icon string) error {
	return launch(cdi.dbusSub, DbusInterfaceSubapplet+".SetIcon", icon, cdi.id)
}

func (cdi *SubIcon) SetEmblem(icon string, position cdtype.EmblemPosition) error {
	return launch(cdi.dbusSub, DbusInterfaceSubapplet+".SetEmblem", icon, int32(position), cdi.id)
}

func (cdi *SubIcon) Animate(animation string, rounds int32) error {
	return launch(cdi.dbusSub, DbusInterfaceSubapplet+".Animate", animation, rounds, cdi.id)
}

func (cdi *SubIcon) ShowDialog(message string, duration int32) error {
	return launch(cdi.dbusSub, DbusInterfaceSubapplet+".ShowDialog", message, duration, cdi.id)
}

func (cdi *SubIcon) DemandsAttention(start bool, animation string) error {
	return launch(cdi.dbusSub, DbusInterfaceSubapplet+".DemandsAttention", start, animation)
}

//
//----------------------------------------------------------[ EVENT CALLBACK ]--

// Event receiver, dispatch it to the configured callback.
//
func (cda *CdDbus) receivedMainEvent(event string, data []interface{}) {
	switch event {
	case "on_stop_module":
		log.Debug("Received from dock", event)
		if cda.Events.End != nil {
			cda.Events.End()
		}
		cda.Close <- true // Send closing signal.
	case "on_reload_module":
		if cda.Events.Reload != nil {
			go cda.Events.Reload(data[0].(bool))
		}
	case "on_click":
		if cda.Events.OnClick != nil {
			go cda.Events.OnClick()
		}
	case "on_middle_click":
		if cda.Events.OnMiddleClick != nil {
			go cda.Events.OnMiddleClick()
		}
	case "on_build_menu":
		if cda.Events.OnBuildMenu != nil {
			go cda.Events.OnBuildMenu()
		}
	case "on_menu_select":
		if cda.Events.OnMenuSelect != nil {
			go cda.Events.OnMenuSelect(data[0].(int32))
		}
	case "on_scroll":
		if cda.Events.OnScroll != nil {
			go cda.Events.OnScroll(data[0].(bool))
		}
	case "on_drop_data":
		if cda.Events.OnDropData != nil {
			go cda.Events.OnDropData(data[0].(string))
		}
	case "on_answer":
		if cda.Events.OnAnswer != nil {
			go cda.Events.OnAnswer(data[0])
		}
	case "on_answer_dialog":
		if cda.Events.OnAnswerDialog != nil {
			go cda.Events.OnAnswerDialog(data[0].(int32), data[1])
		}
	case "on_shortkey":
		if cda.Events.OnShortkey != nil {
			go cda.Events.OnShortkey(data[0].(string))
		}
	case "on_change_focus":
		if cda.Events.OnChangeFocus != nil {
			go cda.Events.OnChangeFocus(data[0].(bool))
		}
	default:
		log.Info(event, data)
	}
}

func (cda *CdDbus) receivedSubEvent(event string, data []interface{}) {
	switch event {
	case "on_click_sub_icon":
		if cda.SubEvents.OnSubClick != nil {
			go cda.SubEvents.OnSubClick(data[0].(int32), data[1].(string))
		}
	case "on_middle_click_sub_icon":
		if cda.SubEvents.OnSubMiddleClick != nil {
			go cda.SubEvents.OnSubMiddleClick(data[0].(string))
		}
	case "on_scroll_sub_icon":
		if cda.SubEvents.OnSubScroll != nil {
			go cda.SubEvents.OnSubScroll(data[0].(bool), data[1].(string))
		}
	case "on_drop_data_sub_icon":
		if cda.SubEvents.OnSubDropData != nil {
			go cda.SubEvents.OnSubDropData(data[0].(string), data[1].(string))
		}
	case "on_build_menu_sub_icon":
		if cda.SubEvents.OnSubBuildMenu != nil {
			go cda.SubEvents.OnSubBuildMenu(data[0].(string))
		}
	case "on_menu_select_sub_icon":
		if cda.SubEvents.OnSubMenuSelect != nil {
			go cda.SubEvents.OnSubMenuSelect(data[0].(int32), data[1].(string))
		}
	default:
		log.Info(event, data)
	}
}

//
//------------------------------------------------------------------[ COMMON ]--

// Recast list of args to []interface as requested by the DBus API.
//
func toMapVariant(input map[string]interface{}) map[string]dbus.Variant {
	vars := make(map[string]dbus.Variant)
	for k, v := range input {
		vars[k] = dbus.MakeVariant(v)
	}
	// 	size := valuesVal.Len()
	// 	ret := make([]interface{}, size)
	// 	for i := 0; i < size; i++ {
	// 		ret[i] = valuesVal.Index(i).Interface()
	// 	}
	return vars
}

// Recast list of args to []interface as requested by the DBus API.
//
// func toInterface(valuesVal reflect.Value) []interface{} {
// 	size := valuesVal.Len()
// 	ret := make([]interface{}, size)
// 	for i := 0; i < size; i++ {
// 		ret[i] = valuesVal.Index(i).Interface()
// 	}
// 	return ret
// }

//
//---------------------------------------------------------[ UNUSED / BUGGED ]--

/*


	// Connect defined events callbacks.
	// typ := reflect.TypeOf(cda.Events)
	// elem := reflect.ValueOf(&cda.Events).Elem()
	// for i := 0; i < typ.NumField(); i++ { // Parsing all fields in type.
	// 	cda.connectEvent(elem.Field(i), typ.Field(i))
	// }


// Connect an event to the dock if a callback is defined.
//
func (cda *CdDbus) connectEvent(elem reflect.Value, structField reflect.StructField) {
	conn, _ := dbus.SessionBus()

	tag := structField.Tag.Get("event")                          // Field must have the event tag.
	if tag != "" && (!elem.IsNil() || tag == "on_stop_module") { // And a valid callback. stop module is mandatory for the close signal.
		log.Info("Binded event", tag)
		// 	rule := &dbus.MatchRule{
		// 		Type:      dbus.TypeSignal,
		// 		Interface: DbusInterfaceApplet,
		// 		Member:    tag,
		// 		Path:      cda.BusPath,

		var ret interface{}
		e := conn.BusObject().Call(
			"org.freedesktop.DBus.AddMatch",
			0,
			// "type='signal',sender='org.freedesktop.DBus'").Store()
			"type='signal',path='"+string(cda.BusPath)+"',interface='"+DbusInterfaceApplet+"',sender='"+DbusObject+"'").Store()
		log.DEV("omar", ret, e)
	}

	// 	cda.dbus.Handle(rule, func(msg *dbus.Message) { cda.receivedMainEvent(msg) })
	// }
}
*/

/*
func (cda *CdDbus) GetIconProperties() interface{} {
	base := cda.dbus.Object("org.cairodock.CairoDock", "/org/cairodock/CairoDock").Interface("org.cairodock.CairoDock")
	//~ return cda.call(base, "GetIconProperties", "container=_MainDock_")
	return cda.call(base, "GetIconProperties", interface{}("class=chromium-browser"))
	//~ return cda.call(base, "GetIconProperties")
}

func (cda *CdDbus) GetContainerProperties() []interface{} {
	//~ props := &DockProperties{}

	base := cda.dbus.Object("org.cairodock.CairoDock", "/org/cairodock/CairoDock").Interface("org.cairodock.CairoDock")
	data, _ := cda.call(base, "GetContainerProperties", "_MainDock_")
return data
	//~ var args []interface{}{}:= interface{}("_MainDock_")
	//~ args := []string{"_MainDock_"}
	//~ args := "_MainDock_"
	//~ return cda.call(base, "GetIconProperties", "container=_MainDock_")
	//~ return cda.call(base, "GetContainerProperties", "_MainDock_", "")
	//~ return cda.call(base, "GetIconProperties")
}


*/

// func call(connect *dbus.Connection, iface *dbus.Interface, action string, args ...interface{}) error {
// 	if iface == nil {
// 		return errors.New("no subicon interface")
// 	}
// 	method, e := iface.Method(action)
// 	if e != nil {
// 		return e
// 	}
// 	_, err := connect.Call(method, args...)
// 	//~ fmt.Println("ret", ret)
// 	return err
// }

// func dbusAsync(connect *dbus.Connection, iface *dbus.Interface, action string, args ...interface{}) error {
// 	if iface == nil {
// 		return errors.New("no subicon interface")
// 	}
// 	method, e := iface.Method(action)
// 	if e != nil {
// 		return e
// 	}
// 	return connect.CallAsync(method, args...)
// }
