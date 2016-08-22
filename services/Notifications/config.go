package Notifications

import "github.com/sqp/godock/libs/cdtype"

const (
	defaultNotifAltIcon = "img/active.png"
)

//
//------------------------------------------------------------------[ CONFIG ]--

type appletConf struct {
	cdtype.ConfGroupIconBoth `group:"Icon"`
	NotifConfig              `group:"Configuration"`
	groupConfig              `group:"Configuration"`
}

type groupConfig struct {
	NotifAltIcon   string
	DialogDuration int
	DialogTemplate cdtype.Template `default:"dialognotif"`
}

//
//----------------------------------------------------------[ ACTIONS & MENU ]--

// List of actions defined in this applet. Order must match defineActions
// declaration order.
//
const (
	ActionNone = iota
	ActionShowAll
	ActionClear
)

// Actions available in the menu.
//
var menuUser = []int{
	ActionShowAll,
	ActionClear,
}
