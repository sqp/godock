/* 
Package dbus is the godock cairo-dock connector using DBus. It's goal is to connect the
main Cairo-Dock Golang API, godock/libs/dock, to its parent.
*/
package dbus

import (
	"errors"
	"reflect"

	dbus "github.com/remyoudompheng/go-dbus"

	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/log"
)

const (
	DbusObject             = "org.cairodock.CairoDock"
	DBusPathDock           = "/org/cairodock/CairoDock"
	DbusInterfaceDock      = "org.cairodock.CairoDock"
	DbusInterfaceApplet    = "org.cairodock.CairoDock.applet"
	DbusInterfaceSubapplet = "org.cairodock.CairoDock.subapplet"
)

type CdDbus struct {
	Icons   map[string]*SubIcon
	Close   chan bool // will receive true when the applet is closed.
	Events  cdtype.Events
	BusPath string

	// private data
	dbus     *dbus.Connection
	dbusIcon *dbus.Interface
	dbusSub  *dbus.Interface
}

func New(path string) *CdDbus {
	return &CdDbus{
		Icons: make(map[string]*SubIcon),
		Close: make(chan bool),

		BusPath: path,
	}
}

func (cda *CdDbus) GetCloseChan() chan bool {
	return cda.Close
}

//
//------------------------------------------------------------[ DOCK ACTIONS ]--

// Set the active state for the provided module name.
//
func (cda *CdDbus) ActivateModule(module string, state bool) interface{} {
	base := cda.dbus.Object(DbusObject, DBusPathDock).Interface(DbusInterfaceDock)
	return cda.launch(base, "ActivateModule", interface{}(module), interface{}(state))
}

//
//------------------------------------------------------------[ ICON ACTIONS ]--

func (cda *CdDbus) SetQuickInfo(info string) error {
	return cda.launch(cda.dbusIcon, "SetQuickInfo", info)
}

func (cda *CdDbus) SetLabel(label string) error {
	return cda.launch(cda.dbusIcon, "SetLabel", label)
}

// Sets the image of our icon, overwriting the previous one. 
// You can refer to the image by either its name if it's an image from a icon theme, or by a path.
//   app.SetIcon("gimp")  
//   app.SetIcon("gtk-go-up")  
//   app.SetIcon("/path/to/image") 
//
func (cda *CdDbus) SetIcon(icon string) error {
	return cda.launch(cda.dbusIcon, "SetIcon", icon)
}

// Sets an emblem on our icon. The emblem is drawn directly on the icon, so if you want to remove it, you have to use SetIcon with the original image. 
//   The image is given by its path
//   See cdtype.EmblemPosition for valid emblem locations.
//
//   app.SetEmblem("./emblem-charged.png", cdtype.EmblemBottomLeft) 
//
func (cda *CdDbus) SetEmblem(icon string, position cdtype.EmblemPosition) error {
	return cda.launch(cda.dbusIcon, "SetEmblem", icon, int32(position))
}

func (cda *CdDbus) Animate(animation string, rounds int32) error {
	return cda.launch(cda.dbusIcon, "Animate", animation, rounds)
}

func (cda *CdDbus) ShowDialog(message string, duration int32) error {
	return cda.launch(cda.dbusIcon, "ShowDialog", message, duration)
}

func (cda *CdDbus) DemandsAttention(start bool, animation string) error {
	return cda.launch(cda.dbusIcon, "DemandsAttention", start, animation)
}

// PopupDialog

func (cda *CdDbus) AddDataRenderer(typ string, nbval int32, theme string) error {
	return cda.launch(cda.dbusIcon, "AddDataRenderer", typ, nbval, theme)
}

// 
//
func (cda *CdDbus) RenderValues(values []float64) error {
	return cda.launch(cda.dbusIcon, "RenderValues", toInterface(reflect.ValueOf(values)))
}

// Makes your applet control the window of an external application. Steals its
// icon from the Taskbar. Use the xprop command find the class of the window you
// want to control. Use "none" if you want to reset application control.
// Controling an application enables the OnFocusChange callback.
//
func (cda *CdDbus) ControlAppli(applicationClass string) error {
	return cda.launch(cda.dbusIcon, "ControlAppli", applicationClass)
}

// Set the visible state of the application controlled by the icon.
//
func (cda *CdDbus) ShowAppli(show bool) error {
	return cda.launch(cda.dbusIcon, "ShowAppli", interface{}(show))
}

// AddMenuItems

func (cda *CdDbus) PopulateMenu(items ...string) error {
	return cda.launch(cda.dbusIcon, "PopulateMenu", toInterface(reflect.ValueOf(items)))
}

// Bind one or more keyboard shortcuts to your applet. Only non empty shortkeys
// will be sent to the dock so you can use this method to directly add them from
// config.
//
func (cda *CdDbus) BindShortkey(shortkeys ...string) error {
	validkeys := []interface{}{}
	for _, key := range shortkeys {
		if key != "" {
			validkeys = append(validkeys, key)
		}
	}
	return cda.launch(cda.dbusIcon, "BindShortkey", validkeys)
}

func (cda *CdDbus) AskText(message, initialText string) ([]interface{}, error) {
	return cda.call(cda.dbusIcon, "AskText", message, initialText)
}

func (cda *CdDbus) AskValue(message string, initialValue, maxValue float64) ([]interface{}, error) {
	return cda.call(cda.dbusIcon, "AskValue", message, initialValue, maxValue)
}

func (cda *CdDbus) AskQuestion(message string) ([]interface{}, error) {
	return cda.call(cda.dbusIcon, "AskQuestion", message)
}

// Get a property of the icon of your applet. Current available properties are : 
//   x            int?     x position of the icon's center on the screen (starting from 0 on the left)
//   y            int?     y position of the icon's center on the screen (starting from 0 at the top of the screen)
//   width        int?     width of the icon, in pixels (this is the maximum width, when the icon is zoomed)
//   height       int?     height of the icon, in pixels (this is the maximum height, when the icon is zoomed)
//   container    ??       type of container of the applet (DOCK, DESKLET)
//   orientation  ??       position of the container on the screen (BOTTOM, TOP, RIGHT, LEFT). A desklet has always an orientation of BOTTOM.
//   Xid          int64    ID of the application's window which is controlled by the applet, or 0 if none (this parameter can only be non nul if you used the method ControlAppli beforehand).
//   has_focus    bool     Wether the application's window which is controlled by the applet is the current active window (it has the focus) or not. E.g.:
//
func (cda *CdDbus) Get(property string) (interface{}, error) {
	data, e := cda.call(cda.dbusIcon, "Get", property)
	if len(data) > 0 {
		return data[0], e
	}
	return nil, e
}

// invalid signature "a{sv}"
//
// func (cda *CdDbus) GetAll() {
// 	// func GetAll() (*cdtype.DockProperties, error) {
// 	data, e := cda.get(cda.dbusIcon, "GetAll")
// 	log.Info("all", data, e)
// 	// 	if e == nil {
// 	// 		if args, ok := data[0].([]interface{}); ok {
// 	// 			for _, uncast := range args {
// 	// 				if arg, ok := uncast.([]interface{}); ok {
// 	// 					switch arg[0] {
// 	// 					case "Xid":
// 	// 						props.Xid = arg[1].(uint64)
// 	// 					case "x":
// 	// 						props.X = arg[1].(int32)
// 	// 					case "y":
// 	// 						props.Y = arg[1].(int32)
// 	// 					case "orientation":
// 	// 						props.Orientation = arg[1].(uint32)
// 	// 					case "container":
// 	// 						props.Container = arg[1].(uint32)
// 	// 					case "width":
// 	// 						props.Width = arg[1].(int32)
// 	// 					case "height":
// 	// 						props.Height = arg[1].(int32)
// 	// 					case "has_focus":
// 	// 						props.HasFocus = arg[1].(bool)
// 	// 					}

// 	// 					//~ log.Printf("%s:%# v\n", args[0], args[1])

// 	// 				}
// 	// 			}
// 	// 		}
// 	// 	}
// 	// return props, e
// }

//
//--------------------------------------------------------[ SUBICONS ACTIONS ]--

// Add subicons by pack of 3 string : label, icon, id.
//
func (cda *CdDbus) AddSubIcon(fields []string) error {
	for i := 0; i < len(fields)/3; i++ {
		log.Info("icon:", fields[3*i+2])
		id := fields[3*i+2]
		cda.Icons[id] = &SubIcon{cda.dbus, cda.dbusSub, id}
	}
	return cda.launch(cda.dbusSub, "AddSubIcons", toInterface(reflect.ValueOf(fields)))
}

func (cda *CdDbus) RemoveSubIcon(id string) error {
	return cda.launch(cda.dbusSub, "RemoveSubIcon", id)
}

//
//------------------------------------------------------------[ DBUS CONNECT ]--

// Connect the applet manager to the Cairo-Dock core. Saves interfaces to the
// icon and subicon DBus interfaces and connects events callbacks.
//
func (cda *CdDbus) ConnectToBus() (e error) {
	cda.dbus, e = dbus.Connect(dbus.SessionBus)
	if e != nil {
		log.Info("DBus Connect", e)
		return e
	}
	if e = cda.dbus.Authenticate(); e != nil {
		log.Info("Failed Connection.Authenticate:", e.Error())
		return e
	}

	cda.dbusIcon = cda.dbus.Object(DbusObject, cda.BusPath).Interface(DbusInterfaceApplet)
	cda.dbusSub = cda.dbus.Object(DbusObject, cda.BusPath+"/sub_icons").Interface(DbusInterfaceSubapplet)
	if cda.dbusIcon == nil || cda.dbusSub == nil {
		return errors.New("No DBus interface")
	}

	// Connect defined events callbacks.
	typ := reflect.TypeOf(cda.Events)
	elem := reflect.ValueOf(&cda.Events).Elem()
	for i := 0; i < typ.NumField(); i++ { // Parsing all fields in type.
		cda.connectEvent(elem.Field(i), typ.Field(i))
	}
	return nil
}

// Connect an event to the dock if a callback is defined.
//
func (cda *CdDbus) connectEvent(elem reflect.Value, structField reflect.StructField) {
	tag := structField.Tag.Get("event")                          // Field must have the event tag.
	if tag != "" && (!elem.IsNil() || tag == "on_stop_module") { // And a valid callback. stop module is mandatory for the close signal.
		// log.Info("Binded event", tag)
		rule := &dbus.MatchRule{
			Type:      dbus.TypeSignal,
			Interface: DbusInterfaceApplet,
			Member:    tag,
			Path:      cda.BusPath,
		}
		cda.dbus.Handle(rule, func(msg *dbus.Message) { cda.receivedMainEvent(msg) })
	}
}

//
//----------------------------------------------------------[ EVENT CALLBACK ]--

// Event receiver, dispatch it to the configured callback.
//
func (cda *CdDbus) receivedMainEvent(msg *dbus.Message) {
	switch msg.Member {
	case "on_stop_module":
		if cda.Events.End != nil {
			cda.Events.End()
		}
		cda.Close <- true // Send closing signal.
	case "on_reload_module":
		go cda.Events.Reload(msg.Params[0].(bool))
	case "on_click":
		go cda.Events.OnClick() // should use msg.Params[0].(int32) ?
	case "on_middle_click":
		go cda.Events.OnMiddleClick()
	case "on_build_menu":
		go cda.Events.OnBuildMenu()
	case "on_menu_select":
		go cda.Events.OnMenuSelect(msg.Params[0].(int32))
	case "on_scroll":
		go cda.Events.OnScroll(msg.Params[0].(bool))
	case "on_drop_data":
		go cda.Events.OnDropData(msg.Params[0].(string))
	case "on_answer":
		go cda.Events.OnAnswer(msg.Params[0])
	case "on_answer_dialog":
		go cda.Events.OnAnswerDialog(msg.Params[0].(int32), msg.Params[1])
	case "on_shortkey":
		go cda.Events.OnShortkey(msg.Params[0].(string))
	case "on_change_focus":
		go cda.Events.OnChangeFocus(msg.Params[0].(bool))
	default:
		log.Info(msg.Member, msg.Params)
	}
}

//
//------------------------------------------------------------[ DBUS ACTIONS ]--

// Call DBus method.
//
func (cda *CdDbus) call(iface *dbus.Interface, action string, args ...interface{}) (ret []interface{}, err error) {
	method, e := iface.Method(action)
	if e != nil {
		return nil, e
	}
	ret, err = cda.dbus.Call(method, args...)
	return ret, e
}

// Call DBus method and only get the error if any.
//
func (cda *CdDbus) launch(iface *dbus.Interface, action string, args ...interface{}) error {
	_, e := cda.call(iface, action, args...)
	return e
}

// func (cda *CdDbus) get(iface *dbus.Interface, action string, args ...interface{}) (ret []interface{}, eror error) {
// 	method, e := iface.Method(action)
// 	if e != nil {
// 		return nil, e
// 	}

// 	r, err := cda.dbus.Invoke(method, args...)
// 	log.Info("ret", err)
// 	pretty.Printf("%# v\n", r.Params)

// 	r2, _ := cda.call(iface, action, args...)
// 	log.Info("r2", err)
// 	pretty.Printf("%# v\n", r2)

// 	return ret, err
// }

//
//------------------------------------------------------------------[ COMMON ]--

// Recast list of args to []interface as requested by the DBus API.
//
func toInterface(valuesVal reflect.Value) []interface{} {
	size := valuesVal.Len()
	ret := make([]interface{}, size)
	for i := 0; i < size; i++ {
		ret[i] = valuesVal.Index(i).Interface()
	}
	return ret
}

//
//---------------------------------------------------------[ UNUSED / BUGGED ]--

// SubIcons actions.
//
type SubIcon struct {
	connect *dbus.Connection
	interf  *dbus.Interface
	id      string
}

// func (cdi *SubIcon) SetQuickInfo(info string) error {
// 	return call(cdi.connect, cdi.interf, "SetQuickInfo", info, cdi.id)
// }

// func (cdi *SubIcon) SetLabel(label string) error {
// 	return call(cdi.connect, cdi.interf, "SetLabel", label, cdi.id)
// }

// func (cdi *SubIcon) SetIcon(icon string) error {
// 	return call(cdi.connect, cdi.interf, "SetIcon", icon, cdi.id)
// }

// func (cdi *SubIcon) SetEmblem(icon string, position cdtype.EmblemPosition) error {
// 	return call(cdi.connect, cdi.interf, "SetEmblem", icon, int32(position), cdi.id)
// }

// func (cdi *SubIcon) Animate(animation string, rounds int32) error {
// 	return call(cdi.connect, cdi.interf, "Animate", animation, rounds, cdi.id)
// }

// func (cdi *SubIcon) ShowDialog(message string, duration int32) error {
// 	return call(cdi.connect, cdi.interf, "ShowDialog", message, duration, cdi.id)
// }

// func (cda *CdDbus) receivedSubEvent(msg *dbus.Message) {
// 	icon := msg.Params[0].(string)
// 	switch msg.Member {
// 	case "on_click_sub_icon":
// 		//~ fmt.Println("clicked nbparam=", len(msg.Params))
// 		go cda.Events.OnSubClick(icon, 0) // TODO debug : no param received.
// 	case "on_middle_click_sub_icon":
// 		go cda.Events.OnSubMiddleClick(icon)
// 	case "on_scroll_sub_icon":
// 		go cda.Events.OnSubScroll(icon, msg.Params[1].(bool))
// 	case "on_drop_data_sub_icon":
// 		go cda.Events.OnSubDropData(icon, msg.Params[1].(string))
// 	case "on_build_menu_sub_icon":
// 		go cda.Events.OnSubBuildMenu(icon)
// 	case "on_menu_select_sub_icon":
// 		go cda.Events.OnSubMenuSelect(icon, msg.Params[1].(int))
// 	default:
// 		fmt.Println(msg.Member, msg.Params)
// 	}
// }

//

//~ OnSubClick(icon string, state int)
//~ OnSubMiddleClick(icon string)
//~ OnSubBuildMenu(icon string)
//~ OnSubMenuSelect(icon string, numEntry int)
//~ OnSubScroll(icon string, scrollUp bool)
//~ OnSubDropData(icon string, data string)

// for _, event := range []string{"on_click_sub_icon", "on_middle_click_sub_icon", "on_scroll_sub_icon", "on_drop_data_sub_icon", "on_build_menu_sub_icon", "on_menu_select_sub_icon"} {
// 	rule := &dbus.MatchRule{
// 		Type:      dbus.TypeSignal,
// 		Interface: DbusInterfaceSubapplet,
// 		Member:    event,
// 		Path:      cda.BusPath + "/sub_icons",
// 	}
// 	cda.dbus.Handle(rule, func(msg *dbus.Message) { cda.receivedSubEvent(msg) })
// }

//

//~ func (cda *CdDbus) AddMenuItems(items... map[string]interface{}) error {
//~ func (cda *CdDbus) AddMenuItems(items... [][]interface{}) error {
//~ log.Println("menu", items)
//~ return cda.call(cda.dbusIcon, "AddMenuItems", items)
//~ }

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

// aa{sv}
func (cda *CdDbus) Test() error {
	//~ menuitem := map[string]interface{}{"widget-type" : 0,  
	//~ "label": "this is an entry of the main menu",  
	//~ "icon" : "gtk-add",  
	//~ "menu" : 0,  
	//~ "id" : 1,  
	//~ "tooltip" : "this is the tooltip that will appear when you hover this entry"}

	return cda.call(cda.dbusIcon, "PopulateMenu", []interface{}{"test", "cool"})

	menuitem := [][]interface{}{
		{"widget-type", 0},
		{"label", "this is an entry of the main menu"},
		{"icon", "gtk-add"},
		{"menu", 0},
		{"id", 1},
		{"tooltip", "this is the tooltip that will appear when you hover this entry"},
	}
	//~ demo.AddMenuItems(menuitem)
	return cda.call(cda.dbusIcon, "AddMenuItems", []interface{}{menuitem})
	//~ return cda.call(cda.dbusIcon, "AddMenuItems", menuitem)
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
