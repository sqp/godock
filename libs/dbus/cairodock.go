/* 
dbus is the godock cairo-dock connector using DBus. It's goal is to connect the
main Cairo-Dock Golang API, godock/libs/dock, to its parent.
*/
package dbus

import (
	"errors"
	"fmt"
	dbus "github.com/norisatir-message/go-dbus"
	"github.com/sqp/godock/libs/cdtype"
	"log"
	//~ "os"
	//~ "path"
)

// TODO: to use
//~ const (
//~ DBUS_NAME      = "org.cairodock.CairoDock"
//~ DBUS_INTERFACE = "org.cairodock.CairoDock"
//~ )

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

//------------------------------------------------------------------------------
// Actions on the icon.
//------------------------------------------------------------------------------

func (cda *CdDbus) SetQuickInfo(info string) error {
	return cda.dbusAction(cda.dbusIcon, "SetQuickInfo", info)
}

func (cda *CdDbus) SetLabel(label string) error {
	return cda.dbusAction(cda.dbusIcon, "SetLabel", label)
}

func (cda *CdDbus) SetIcon(icon string) error {
	return cda.dbusAction(cda.dbusIcon, "SetIcon", icon)
}

func (cda *CdDbus) SetEmblem(icon string, position cdtype.EmblemPosition) error {
	return cda.dbusAction(cda.dbusIcon, "SetEmblem", icon, int32(position))
}

func (cda *CdDbus) Animate(animation string, rounds int32) error {
	return cda.dbusAction(cda.dbusIcon, "Animate", animation, rounds)
}

func (cda *CdDbus) ShowDialog(message string, duration int32) error {
	return cda.dbusAction(cda.dbusIcon, "ShowDialog", message, duration)
}

func (cda *CdDbus) DemandsAttention(start bool, animation string) error {
	return cda.dbusAction(cda.dbusIcon, "DemandsAttention", start, animation)
}

//~ <arg name="hDialogAttributes" type="a{sv}" direction="in"/>
//~ <arg name="hWidgetAttributes" type="a{sv}" direction="in"/>
func (cda *CdDbus) PopupDialog(dialog, widget []interface{}) error {
	return cda.dbusAction(cda.dbusIcon, "PopupDialog", dialog, widget)
}

func (cda *CdDbus) AddDataRenderer(typ string, nbval int32, theme string) error {
	return cda.dbusAction(cda.dbusIcon, "AddDataRenderer", typ, nbval, theme)
}

//~ func (cda *CdDbus) RenderValues(values []float64) error {
//~ return dbusAsync(cda.dbus, cda.dbusIcon, "RenderValues", floats2interface(values))
//~ var uncast interface{} = values
//~ return dbusAsync(cda.dbus, cda.dbusIcon, "RenderValues", uncast)
//~ return cda.dbusAction(cda.dbusIcon, "RenderValues", values)
//~ }

// Makes your applet control the window of an external application. Steals its
// icon from the Taskbar. Use the xprop command find the class of the window you
// want to control. Use "none" if you want to reset application control.
// Controling an application enables the OnFocusChange callback.
//
func (cda *CdDbus) ControlAppli(applicationClass string) error {
	return cda.dbusAction(cda.dbusIcon, "ControlAppli", applicationClass)
}

// Set the visible state of the application controlled by the icon.
//
func (cda *CdDbus) ShowAppli(show bool) error {
	return cda.dbusAction(cda.dbusIcon, "ShowAppli", interface{}(show))
}

//~ func (cda *CdDbus) AddMenuItems(items... map[string]interface{}) error {
//~ func (cda *CdDbus) AddMenuItems(items... [][]interface{}) error {
//~ log.Println("menu", items)
//~ return cda.dbusAction(cda.dbusIcon, "AddMenuItems", items)
//~ }

func (cda *CdDbus) PopulateMenu(items ...string) error {
	return cda.dbusAction(cda.dbusIcon, "PopulateMenu", strings2interface(items))
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
	return cda.dbusAction(cda.dbusIcon, "BindShortkey", validkeys)
}

func (cda *CdDbus) AskText(message, initialText string) ([]interface{}, error) {
	return cda.dbusGet(cda.dbusIcon, "AskText", message, initialText)
}

func (cda *CdDbus) AskValue(message string, initialValue, maxValue float64) ([]interface{}, error) {
	return cda.dbusGet(cda.dbusIcon, "AskValue", message, initialValue, maxValue)
}

func (cda *CdDbus) AskQuestion(message string) ([]interface{}, error) {
	return cda.dbusGet(cda.dbusIcon, "AskQuestion", message)
}

func (cda *CdDbus) Get(property string) ([]interface{}, error) {
	return cda.dbusGet(cda.dbusIcon, "Get", property)
}

func (cda *CdDbus) GetAll() (*cdtype.DockProperties, error) {
	props := &cdtype.DockProperties{}
	data, e := cda.dbusGet(cda.dbusIcon, "GetAll")
	if e == nil {
		if args, ok := data[0].([]interface{}); ok {
			for _, uncast := range args {
				if arg, ok := uncast.([]interface{}); ok {
					switch arg[0] {
					case "Xid":
						props.Xid = arg[1].(uint64)
					case "x":
						props.X = arg[1].(int32)
					case "y":
						props.Y = arg[1].(int32)
					case "orientation":
						props.Orientation = arg[1].(uint32)
					case "container":
						props.Container = arg[1].(uint32)
					case "width":
						props.Width = arg[1].(int32)
					case "height":
						props.Height = arg[1].(int32)
					case "has_focus":
						props.HasFocus = arg[1].(bool)
					}

					//~ log.Printf("%s:%# v\n", args[0], args[1])

				}
			}
		}
	}
	return props, e
}

func (cda *CdDbus) ActivateModule(module string, state bool) interface{} {
	base := cda.dbus.Object("org.cairodock.CairoDock", "/org/cairodock/CairoDock").Interface("org.cairodock.CairoDock")
	return cda.dbusAction(base, "ActivateModule", interface{}(module), interface{}(state))
}

/*
func (cda *CdDbus) GetIconProperties() interface{} {
	base := cda.dbus.Object("org.cairodock.CairoDock", "/org/cairodock/CairoDock").Interface("org.cairodock.CairoDock")
	//~ return cda.dbusAction(base, "GetIconProperties", "container=_MainDock_")
	return cda.dbusAction(base, "GetIconProperties", interface{}("class=chromium-browser"))
	//~ return cda.dbusAction(base, "GetIconProperties")
}

func (cda *CdDbus) GetContainerProperties() []interface{} {
	//~ props := &DockProperties{}

	base := cda.dbus.Object("org.cairodock.CairoDock", "/org/cairodock/CairoDock").Interface("org.cairodock.CairoDock")
	data, _ := cda.dbusGet(base, "GetContainerProperties", "_MainDock_")
return data
	//~ var args []interface{}{}:= interface{}("_MainDock_")
	//~ args := []string{"_MainDock_"}
	//~ args := "_MainDock_"
	//~ return cda.dbusAction(base, "GetIconProperties", "container=_MainDock_")
	//~ return cda.dbusAction(base, "GetContainerProperties", "_MainDock_", "")
	//~ return cda.dbusAction(base, "GetIconProperties")
}
*/

// aa{sv}
func (cda *CdDbus) Test() error {
	//~ menuitem := map[string]interface{}{"widget-type" : 0,  
	//~ "label": "this is an entry of the main menu",  
	//~ "icon" : "gtk-add",  
	//~ "menu" : 0,  
	//~ "id" : 1,  
	//~ "tooltip" : "this is the tooltip that will appear when you hover this entry"}

	return cda.dbusAction(cda.dbusIcon, "PopulateMenu", []interface{}{"test", "cool"})

	menuitem := [][]interface{}{
		{"widget-type", 0},
		{"label", "this is an entry of the main menu"},
		{"icon", "gtk-add"},
		{"menu", 0},
		{"id", 1},
		{"tooltip", "this is the tooltip that will appear when you hover this entry"},
	}
	//~ demo.AddMenuItems(menuitem)
	return cda.dbusAction(cda.dbusIcon, "AddMenuItems", []interface{}{menuitem})
	//~ return cda.dbusAction(cda.dbusIcon, "AddMenuItems", menuitem)
}

//------------------------------------------------------------------------------
// Actions on sub icons.
//------------------------------------------------------------------------------

// Add subicons by pack of 3 string : label, icon, id.
//
func (cda *CdDbus) AddSubIcon(fields []string) error {
	for i := 0; i < len(fields)/3; i++ {
		log.Println("icon:", fields[3*i+2])
		id := fields[3*i+2]
		cda.Icons[id] = &SubIcon{cda.dbus, cda.dbusSub, id}
	}
	return cda.dbusAction(cda.dbusSub, "AddSubIcons", strings2interface(fields))
}

func (cda *CdDbus) RemoveSubIcon(id string) error {
	return cda.dbusAction(cda.dbusSub, "RemoveSubIcon", id)
}

// SubIcons actions.
//
type SubIcon struct {
	connect *dbus.Connection
	interf  *dbus.Interface
	id      string
}

func (cdi *SubIcon) SetQuickInfo(info string) error {
	return dbusAction(cdi.connect, cdi.interf, "SetQuickInfo", info, cdi.id)
}

func (cdi *SubIcon) SetLabel(label string) error {
	return dbusAction(cdi.connect, cdi.interf, "SetLabel", label, cdi.id)
}

func (cdi *SubIcon) SetIcon(icon string) error {
	return dbusAction(cdi.connect, cdi.interf, "SetIcon", icon, cdi.id)
}

func (cdi *SubIcon) SetEmblem(icon string, position cdtype.EmblemPosition) error {
	return dbusAction(cdi.connect, cdi.interf, "SetEmblem", icon, int32(position), cdi.id)
}

func (cdi *SubIcon) Animate(animation string, rounds int32) error {
	return dbusAction(cdi.connect, cdi.interf, "Animate", animation, rounds, cdi.id)
}

func (cdi *SubIcon) ShowDialog(message string, duration int32) error {
	return dbusAction(cdi.connect, cdi.interf, "ShowDialog", message, duration, cdi.id)
}

//------------------------------------------------------------------------------
// Applet Callback (user interaction)
//------------------------------------------------------------------------------

func (cda *CdDbus) receivedMainEvent(msg *dbus.Message) {
	switch msg.Member {
	case "on_stop_module":
		if cda.Events.End != nil {
			cda.Events.End()
		}
		cda.Close <- true // Send closing signal.
	case "on_reload_module":
		confChanged := msg.Params[0].(bool)
		//~ if confChanged {
		//~ cda.getConfig()
		//~ }
		if cda.Events.Reload != nil {
			go cda.Events.Reload(confChanged)
		}
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
		fmt.Println(msg.Member, msg.Params)
	}
}

func (cda *CdDbus) receivedSubEvent(msg *dbus.Message) {
	icon := msg.Params[0].(string)
	switch msg.Member {
	case "on_click_sub_icon":
		//~ fmt.Println("clicked nbparam=", len(msg.Params))
		go cda.Events.OnSubClick(icon, 0) // TODO debug : no param received.
	case "on_middle_click_sub_icon":
		go cda.Events.OnSubMiddleClick(icon)
	case "on_scroll_sub_icon":
		go cda.Events.OnSubScroll(icon, msg.Params[1].(bool))
	case "on_drop_data_sub_icon":
		go cda.Events.OnSubDropData(icon, msg.Params[1].(string))
	case "on_build_menu_sub_icon":
		go cda.Events.OnSubBuildMenu(icon)
	case "on_menu_select_sub_icon":
		go cda.Events.OnSubMenuSelect(icon, msg.Params[1].(int))
	default:
		fmt.Println(msg.Member, msg.Params)
	}
}

//------------------------------------------------------------------------------
// DBus actions.
//------------------------------------------------------------------------------

//~ 
//~ OnSubClick(icon string, state int)
//~ OnSubMiddleClick(icon string)
//~ OnSubBuildMenu(icon string)
//~ OnSubMenuSelect(icon string, numEntry int)
//~ OnSubScroll(icon string, scrollUp bool)
//~ OnSubDropData(icon string, data string)
//~ 

func (cda *CdDbus) ConnectToBus() (e error) {
	cda.dbus, e = dbus.Connect(dbus.SessionBus)
	if e != nil {
		log.Println("DBus Connect", e)
		return e
	}
	if e = cda.dbus.Authenticate(); e != nil {
		log.Println("Failed Connection.Authenticate:", e.Error())
		return e
	}

	cda.dbusIcon = cda.dbus.Object("org.cairodock.CairoDock", cda.BusPath).Interface("org.cairodock.CairoDock.applet")
	cda.dbusSub = cda.dbus.Object("org.cairodock.CairoDock", cda.BusPath+"/sub_icons").Interface("org.cairodock.CairoDock.subapplet")

	var events []string
	if cda.Events.OnClick != nil {
		events = append(events, "on_click")
	}
	if cda.Events.OnMiddleClick != nil {
		events = append(events, "on_middle_click")
	}
	if cda.Events.OnBuildMenu != nil {
		events = append(events, "on_build_menu")
	}
	if cda.Events.OnMenuSelect != nil {
		events = append(events, "on_menu_select")
	}
	if cda.Events.OnScroll != nil {
		events = append(events, "on_scroll")
	}
	if cda.Events.OnDropData != nil {
		events = append(events, "on_drop_data")
	}
	if cda.Events.OnAnswer != nil {
		events = append(events, "on_answer")
	}
	if cda.Events.OnAnswerDialog != nil {
		events = append(events, "on_answer_dialog")
	}
	if cda.Events.OnShortkey != nil {
		events = append(events, "on_shortkey")
	}
	if cda.Events.OnChangeFocus != nil {
		events = append(events, "on_change_focus")
	}

	// Mandatory.
	events = append(events, "on_reload_module")
	events = append(events, "on_stop_module")

	for _, event := range events {
		rule := &dbus.MatchRule{
			Type:      dbus.TypeSignal,
			Interface: "org.cairodock.CairoDock.applet",
			Member:    event,
			Path:      cda.BusPath,
		}
		cda.dbus.Handle(rule, func(msg *dbus.Message) { cda.receivedMainEvent(msg) })
	}

	for _, event := range []string{"on_click_sub_icon", "on_middle_click_sub_icon", "on_scroll_sub_icon", "on_drop_data_sub_icon", "on_build_menu_sub_icon", "on_menu_select_sub_icon"} {
		rule := &dbus.MatchRule{
			Type:      dbus.TypeSignal,
			Interface: "org.cairodock.CairoDock.subapplet",
			Member:    event,
			Path:      cda.BusPath + "/sub_icons",
		}
		cda.dbus.Handle(rule, func(msg *dbus.Message) { cda.receivedSubEvent(msg) })
	}

	return nil
}

// Connect event to the dock if a callback is defined.
//
//~ func (cda *CdDbus) addEvent(event string, call func) {
//~ if cda.Events.OnBuildMenu != nil {
//~ events = append(events, "on_build_menu")
//~ }
//~ }

func (cda *CdDbus) dbusGet(iface *dbus.Interface, action string, args ...interface{}) ([]interface{}, error) {
	if iface == nil {
		return nil, errors.New("no subicon interface")
	}
	method, e := iface.Method(action)
	if e != nil {
		return nil, e
	}
	return cda.dbus.Call(method, args...)
}

func (cda *CdDbus) dbusAction(iface *dbus.Interface, action string, args ...interface{}) error {
	if iface == nil {
		return errors.New("no subicon interface")
	}
	method, e := iface.Method(action)
	if e != nil {
		return e
	}
	_, err := cda.dbus.Call(method, args...)
	//~ fmt.Println("ret", ret)
	return err
}

func dbusAction(connect *dbus.Connection, iface *dbus.Interface, action string, args ...interface{}) error {
	if iface == nil {
		return errors.New("no subicon interface")
	}
	method, e := iface.Method(action)
	if e != nil {
		return e
	}
	_, err := connect.Call(method, args...)
	//~ fmt.Println("ret", ret)
	return err
}

func dbusAsync(connect *dbus.Connection, iface *dbus.Interface, action string, args ...interface{}) error {
	if iface == nil {
		return errors.New("no subicon interface")
	}
	method, e := iface.Method(action)
	if e != nil {
		return e
	}
	return connect.CallAsync(method, args...)
}

//------------------------------------------------------------------------------
// Common
//------------------------------------------------------------------------------

func strings2interface(strings []string) (uncasted []interface{}) {
	for _, str := range strings {
		uncasted = append(uncasted, str)
	}
	return
}

func floats2interface(floats []float64) (uncasted []interface{}) {
	for _, fl := range floats {
		uncasted = append(uncasted, fl)
	}
	return
}
