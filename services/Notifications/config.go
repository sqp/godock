package Notifications

const (
	defaultNotifAltIcon = "img/active.png"
)

// Dialog types.
// const (
// 	dialogInternal = "Internal dialog"
// 	dialogNotify   = "Desktop notifications"
// )

//------------------------------------------------------------------------------
// Config
//------------------------------------------------------------------------------

type appletConf struct {
	groupIcon   `group:"Icon"`
	groupConfig `group:"Configuration"`
	// groupActions `group:"Actions"`
}

type groupIcon struct {
	Icon string `conf:"icon"`
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
