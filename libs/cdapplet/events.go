package cdapplet

import (
	"github.com/sqp/godock/libs/cdtype"
)

//
//------------------------------------------------------------------[ EVENTS ]--

// OnEvent forward the received event to the registered event callback.
// Return true if the signal was quit applet.
//
func (cda *CDApplet) OnEvent(event string, data ...interface{}) (exit bool) {
	cda.log.Debug("Event "+event, data...)

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
			go cda.events.OnClick()
		}
		if cda.events.OnClickMod != nil {
			go cda.events.OnClickMod(data[0].(int))
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
		key := data[0].(string)
		for _, sk := range cda.shortkeys {
			test, e := sk.TestKey(key)
			cda.log.Err(e, "shortkey="+key)
			if test {
				return false
			}
		}
	case "on_change_focus":
		if cda.events.OnChangeFocus != nil {
			go cda.events.OnChangeFocus(data[0].(bool))
		}

	// SubEvents. (icon name is moved back to first arg as it made more sense in that order)

	case "on_click_sub_icon":
		if cda.events.OnSubClick != nil {
			go cda.events.OnSubClick(data[1].(string))
		}
		if cda.events.OnSubClickMod != nil {
			go cda.events.OnSubClickMod(data[1].(string), data[0].(int))
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
