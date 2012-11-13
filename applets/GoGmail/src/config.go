package main

// Constants it's better not to have in conf.
//
const (
	defaultUpdateDelay = 5
	loginLocation      = "../../.Gmail_subscription"
	feedGmail          = "https://mail.google.com/mail/feed/atom/"
)

// Dialog types.
//
const (
	dialogInternal = "Internal dialog"
	dialogNotify   = "Desktop notifications"
)

// Renderers.
//
const (
	NoDisplay   = "no"
	QuickInfo   = "quickinfo"
	EmblemSmall = "small emblem"
	EmblemLarge = "large emblem"
)

//~ self.svgpath = self.path+'emblem.svg' # SVG emblem file

//------------------------------------------------------------------[ CONFIG ]--

// Global struct conf.
//
type mailConf struct {
	MailIcon    `group:"Icon"`
	MailConfig  `group:"Configuration"`
	MailActions `group:"Actions"`
}

// Tab Icon.
//
type MailIcon struct {
	Icon string
}

// Tab Configuration.
//
type MailConfig struct {
	UpdateDelay            int
	Renderer               string
	DialogType             string
	DialogTimer            int
	DialogNbMailActionShow int

	MonitorName    string
	MonitorEnabled bool
}

// Tab Actions.
//
type MailActions struct {
	AlertDialogEnabled   bool
	AlertDialogMaxNbMail int

	AlertAnimName     string
	AlertAnimDuration int
	AlertSoundEnabled bool
	AlertSoundFile    string

	ActionClickLeft   string
	ActionClickMiddle string
	ShortkeyOpen      string
	ShortkeyCheck     string

	// Still hidden.
	Debug          bool
	PollingEnabled bool
	// FeedGmail      string // Url of the Atom feed source. Unused yet. See const
	//~ DebugLevel int // unused

	// Defaults are currently added to the last tab of config. This could evolve,
	// but atm, this sound like a sane choice to have something consistant. All
	// values that would be hardcoded are grouped here so we have a good overview
	// of what is used (const & var). And in the conf file, we have all possibly
	// tweakable or fixable values.
	DefaultMonitorName    string // Default application or webpage to open.
	DefaultAlertSoundFile string // 
}

//----------------------------------------------------------[ ACTIONS & MENU ]--

// List of actions defined in this applet. Order must match defineActions
// declaration order.
//
const (
	ActionNone = iota
	ActionOpenClient
	ActionCheckMail
	ActionShowMails
	ActionRegister
)

// Actions available in full menu.
//
var menuFull []int = []int{
	ActionOpenClient,
	ActionCheckMail,
	ActionShowMails,
	ActionNone,
	ActionRegister,
}

// Actions available in register menu. Displayed when account isn't set.
//
var menuRegister []int = []int{
	ActionRegister,
}
