package Notifications

const (
	defaultNotifAltIcon = "img/active.png"
)

// Dialog types.
// const (
// 	dialogInternal = "Internal dialog"
// 	dialogNotify   = "Desktop notifications"
// )

//
//------------------------------------------------------------------[ CONFIG ]--

type appletConf struct {
	groupIcon   `group:"Icon"`
	groupConfig `group:"Configuration"`
	// groupActions `group:"Actions"`
}

type groupIcon struct {
	Icon string `conf:"icon"`
	Name string `conf:"name"`
}

type groupConfig struct {
	// UpdateDelay            int
	// Renderer               string
	// DialogType             string
	// DialogTimer            int
	// DialogNbMailActionShow int

	// MonitorName    string
	// MonitorEnabled bool

	NotifSize int
	// clear=true
	NotifBlackList []string

	NotifAltIcon string

	Debug bool
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
