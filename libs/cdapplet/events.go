package cdapplet

import (
	"github.com/sqp/godock/libs/cdtype"

	"reflect"
)

//
//------------------------------------------------------------------[ EVENTS ]--

// OnEvent forward the received event to the registered event callback.
// Return true if the signal was quit applet.
//
func (cda *CDApplet) OnEvent(event string, data ...interface{}) (exit bool) {
	cda.log.Debug("Event "+event, data...)

	cda.hooker.Call(event, data...)

	switch event {
	case "on_stop_module":
		cda.log.Debug("Received from dock", event)
		if cda.events.End != nil {
			cda.events.End()
		}
		return true

	case "on_reload_module":
		if cda.events.Reload != nil {
			cda.events.Reload(data[0].(bool)) // no async, need to set debug after for applets services.
		}
	case "on_click":
		if cda.events.OnClick != nil {
			go cda.events.OnClick(data[0].(int))
		}
	case "on_middle_click":
		if cda.events.OnMiddleClick != nil {
			go cda.events.OnMiddleClick()
		}
	case "on_build_menu":
		if cda.events.OnBuildMenu != nil {
			cda.events.OnBuildMenu(data[0].(cdtype.Menuer)) // no async for menus, we need to populate it after on dbus backend.
		}
	case "on_scroll":
		if cda.events.OnScroll != nil {
			go cda.events.OnScroll(data[0].(bool))
		}
	case "on_drop_data":
		if cda.events.OnDropData != nil {
			go cda.events.OnDropData(data[0].(string))
		}
	case "on_shortkey":
		if cda.events.OnShortkey != nil {
			go cda.events.OnShortkey(data[0].(string))
		}
	case "on_change_focus":
		if cda.events.OnChangeFocus != nil {
			go cda.events.OnChangeFocus(data[0].(bool))
		}

	// SubEvents. (icon name is moved back to first arg as it made more sense in that order)

	case "on_click_sub_icon":
		if cda.events.OnSubClick != nil {
			go cda.events.OnSubClick(data[1].(string), data[0].(int))
		}
	case "on_middle_click_sub_icon":
		if cda.events.OnSubMiddleClick != nil {
			go cda.events.OnSubMiddleClick(data[0].(string))
		}
	case "on_scroll_sub_icon":
		if cda.events.OnSubScroll != nil {
			go cda.events.OnSubScroll(data[1].(string), data[0].(bool))
		}
	case "on_drop_data_sub_icon":
		if cda.events.OnSubDropData != nil {
			go cda.events.OnSubDropData(data[1].(string), data[0].(string))
		}
	case "on_build_menu_sub_icon":
		if cda.events.OnSubBuildMenu != nil {
			cda.events.OnSubBuildMenu(data[1].(string), data[0].(cdtype.Menuer)) // no async for menus.
		}

	default:
		cda.log.NewWarn("unknown event", event)
	}

	return false
}

//
//-----------------------------------------------------------[ NEW CALLBACKS ]--

// RegisterEvents connects an object to the dock events hooks it implements.
// If the object declares any of the method in the Define... interfaces list, it
// will be registered to receive those events.
//
func (cda *CDApplet) RegisterEvents(obj interface{}) {
	tolisten := cda.hooker.Register(obj)
	if len(tolisten) > 0 {
		cda.log.Debug("listened events", tolisten)
	}
}

// UnregisterEvents disconnects an object from the dock events hooks.
//
func (cda *CDApplet) UnregisterEvents(obj interface{}) {
	tounlisten := cda.hooker.Unregister(obj)
	cda.log.Info("unlistened events", tounlisten)
}

// DefineOnClick is an interface to the OnClick method.
type DefineOnClick interface {
	OnClick(int)
}

// DefineOnMiddleClick is an interface to the OnMiddleClick method.
type DefineOnMiddleClick interface {
	OnMiddleClick()
}

// DefineOnBuildMenu is an interface to the OnBuildMenu method.
type DefineOnBuildMenu interface {
	OnBuildMenu(cdtype.Menuer)
}

// DefineOnScroll is an interface to the OnScroll method.
type DefineOnScroll interface {
	OnScroll(up bool)
}

// DefineOnDropData is an interface to the OnDropData method.
type DefineOnDropData interface {
	OnDropData(string)
}

// DefineOnShortkey is an interface to the OnShortkey method.
type DefineOnShortkey interface {
	OnShortkey(string)
}

// DefineOnChangeFocus is an interface to the OnChangeFocus method.
type DefineOnChangeFocus interface {
	OnChangeFocus(bool)
}

// DefineOnReload is an interface to the OnReload method.
type DefineOnReload interface {
	OnReload(bool)
}

// DefineOnStopModule is an interface to the OnStopModule method.
type DefineOnStopModule interface {
	OnStopModule()
}

// DefineOnSubClick is an interface to the OnSubClick method.
type DefineOnSubClick interface {
	OnSubClick(string, int)
}

// DefineOnSubMiddleClick is an interface to the OnSubMiddleClick method.
type DefineOnSubMiddleClick interface {
	OnSubMiddleClick(string)
}

// DefineOnSubScroll is an interface to the OnSubScroll method.
type DefineOnSubScroll interface {
	OnSubScroll(string, bool)
}

// DefineOnSubDropData is an interface to the OnSubDropData method.
type DefineOnSubDropData interface {
	OnSubDropData(string, string)
}

// DefineOnSubBuildMenu is an interface to the OnSubBuildMenu method.
type DefineOnSubBuildMenu interface {
	OnSubBuildMenu(string, cdtype.Menuer)
}

//
//--------------------------------------------------------[ CALLBACK METHODS ]--

// dockCalls defines callbacks methods for matching objects with type-asserted arguments.
//
var dockCalls = Calls{
	"on_click":         func(m Msg) { m.O.(DefineOnClick).OnClick(m.D[0].(int)) },
	"on_middle_click":  func(m Msg) { m.O.(DefineOnMiddleClick).OnMiddleClick() },
	"on_build_menu":    func(m Msg) { m.O.(DefineOnBuildMenu).OnBuildMenu(m.D[0].(cdtype.Menuer)) },
	"on_scroll":        func(m Msg) { m.O.(DefineOnScroll).OnScroll(m.D[0].(bool)) },
	"on_drop_data":     func(m Msg) { m.O.(DefineOnDropData).OnDropData(m.D[0].(string)) },
	"on_shortkey":      func(m Msg) { m.O.(DefineOnShortkey).OnShortkey(m.D[0].(string)) },
	"on_change_focus":  func(m Msg) { m.O.(DefineOnChangeFocus).OnChangeFocus(m.D[0].(bool)) },
	"on_reload_module": func(m Msg) { m.O.(DefineOnReload).OnReload(m.D[0].(bool)) },
	"on_stop_module":   func(m Msg) { m.O.(DefineOnStopModule).OnStopModule() },

	"on_click_sub_icon":        func(m Msg) { m.O.(DefineOnSubClick).OnSubClick(m.D[1].(string), m.D[0].(int)) },
	"on_middle_click_sub_icon": func(m Msg) { m.O.(DefineOnSubMiddleClick).OnSubMiddleClick(m.D[0].(string)) },
	"on_scroll_sub_icon":       func(m Msg) { m.O.(DefineOnSubScroll).OnSubScroll(m.D[1].(string), m.D[0].(bool)) },
	"on_drop_data_sub_icon":    func(m Msg) { m.O.(DefineOnSubDropData).OnSubDropData(m.D[1].(string), m.D[0].(string)) },
	"on_build_menu_sub_icon":   func(m Msg) { m.O.(DefineOnSubBuildMenu).OnSubBuildMenu(m.D[1].(string), m.D[0].(cdtype.Menuer)) },
}

// dockTests defines callbacks to test if objects are implementing the callback interface.
//
var dockTests = Types{
	"on_click":         reflect.TypeOf((*DefineOnClick)(nil)).Elem(),
	"on_middle_click":  reflect.TypeOf((*DefineOnMiddleClick)(nil)).Elem(),
	"on_build_menu":    reflect.TypeOf((*DefineOnBuildMenu)(nil)).Elem(),
	"on_scroll":        reflect.TypeOf((*DefineOnScroll)(nil)).Elem(),
	"on_drop_data":     reflect.TypeOf((*DefineOnDropData)(nil)).Elem(),
	"on_shortkey":      reflect.TypeOf((*DefineOnShortkey)(nil)).Elem(),
	"on_change_focus":  reflect.TypeOf((*DefineOnChangeFocus)(nil)).Elem(),
	"on_reload_module": reflect.TypeOf((*DefineOnReload)(nil)).Elem(),
	"on_stop_module":   reflect.TypeOf((*DefineOnStopModule)(nil)).Elem(),

	"on_click_sub_icon":        reflect.TypeOf((*DefineOnSubClick)(nil)).Elem(),
	"on_middle_click_sub_icon": reflect.TypeOf((*DefineOnSubMiddleClick)(nil)).Elem(),
	"on_scroll_sub_icon":       reflect.TypeOf((*DefineOnSubScroll)(nil)).Elem(),
	"on_drop_data_sub_icon":    reflect.TypeOf((*DefineOnSubDropData)(nil)).Elem(),
	"on_build_menu_sub_icon":   reflect.TypeOf((*DefineOnSubBuildMenu)(nil)).Elem(),
}

//
//------------------------------------------------------------------[ HOOKER ]--

// Msg defines an event message.
//
type Msg struct {
	O interface{}   // hook client.
	P string        // event name.
	D []interface{} // event data.
}

// Calls defines a list of event callback methods indexed by dbus method name.
//
type Calls map[string]func(Msg)

// Types defines a list of interfaces types indexed by dbus method name.
//
type Types map[string]reflect.Type

// Hooker defines a list of objects indexed by the methods they implement.
// An object can be referenced multiple times.
// If an object declares all methods, he will be referenced in every field.
//   hooker:= NewHooker(myCalls, myTypes)
//
//   // create a type with some of your callback methods and register it.
//   tolisten := hooker.Register(obj) // tolisten is the list of events you may have to listen.
//
//   // add the signal forwarder in your events listening loop.
//   matched := Call(signalName, dbusSignal)
//
type Hooker struct {
	Hooks map[string][]interface{}
	Calls Calls
	Types Types
}

// NewHooker handles a loosely coupled hook interface to forward events.
// to registered clients.
//
func NewHooker(calls Calls, tests Types) *Hooker {
	return &Hooker{
		Hooks: make(map[string][]interface{}),
		Calls: calls,
		Types: tests,
	}
}

// Call forwards an event to registered clients for this event.
//
func (hook Hooker) Call(event string, data ...interface{}) bool {
	call, ok := hook.Calls[event]
	if !ok { // Signal name not defined.
		return false
	}
	if list, ok := hook.Hooks[event]; ok { // Hook clients found.
		for _, obj := range list {
			call(Msg{obj, event, data})
		}
	}
	return true
}

// Register connects an object to the events hooks it implements.
// If the object implements any of the interfaces types declared, it will be
// registered to receive the matching events.
// //
func (hook Hooker) Register(obj interface{}) (tolisten []string) {
	t := reflect.ValueOf(obj).Type()
	for name, modelType := range hook.Types {
		if t.Implements(modelType) {
			hook.Hooks[name] = append(hook.Hooks[name], obj)
			if len(hook.Hooks[name]) == 1 { // First client registered for this event. need to listen.
				tolisten = append(tolisten, name)
			}
		}
	}
	return tolisten
}

// Unregister disconnects an object from the events hooks.
//
func (hook Hooker) Unregister(obj interface{}) (tounlisten []string) {
	for name, list := range hook.Hooks {
		hook.Hooks[name] = hook.remove(list, obj)
		if len(hook.Hooks[name]) == 0 {
			delete(hook.Hooks, name)
			tounlisten = append(tounlisten, name) // No more clients, need to unlisten.
		}
	}
	return tounlisten
}

// AddCalls registers a list of callback methods.
//
// func (hook Hooker) AddCalls(calls *Calls) {
// 	for name, call := range calls {
// 		hook.Calls[name] = call
// 	}
// }

// // AddTypes registers a list of interfaces types.
// //
// func (hook Hooker) AddTypes(tests *Types) {
// 	for name, test := range tests {
// 		hook.Types[name] = test
// 	}
// }

// remove removes an object from the list if found.
//
func (hook Hooker) remove(list []interface{}, obj interface{}) []interface{} {
	for i, test := range list {
		if obj == test {
			return append(list[:i], list[i+1:]...)
		}
	}
	return list
}
