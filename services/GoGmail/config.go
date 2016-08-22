package GoGmail

import "github.com/sqp/godock/libs/cdtype"

// Constants it's better not to have in conf.
//
const (
	loginLocation = ".Gmail_subscription"
	feedGmail     = "https://mail.google.com/mail/feed/atom/"
)

// Renderers.
//
const (
	NoDisplay   = "no"
	QuickInfo   = "quickinfo"
	EmblemSmall = "small emblem"
	EmblemLarge = "large emblem"
)

// Mail client action settings.
//
const (
	MailClientLocation = iota // Open mail client as location, with cdglobal.CmdOpen.
	MailClientProgram         // Open mail client as command.
	MailClientMonitor         // Open mail client as command and monitor it.
)

// Commands references.
const (
	cmdMailClient = iota
)

//------------------------------------------------------------------[ CONFIG ]--

// Global struct conf.
//
type mailConf struct {
	cdtype.ConfGroupIconBoth `group:"Icon"`
	groupConfig              `group:"Configuration"`
	groupActions             `group:"Actions"`
}

type groupConfig struct {
	UpdateDelay  cdtype.Duration `unit:"minute" default:"15"`
	Renderer     string
	DialogTimer  int
	DialogNbMail int

	AlertDialogEnabled bool
	// AlertDialogMaxNbMail int

	AlertAnimName     string
	AlertAnimDuration int
	AlertSoundEnabled bool
	AlertSoundFile    string          `default:"snd/pop.wav"`
	DialogTemplate    cdtype.Template `default:"dialogmail"`
}

type groupActions struct {
	ActionClickLeft   string
	ActionClickMiddle string

	ShortkeyOpenClient *cdtype.Shortkey `action:"1" desc:"Open mail client"`
	ShortkeyShowMails  *cdtype.Shortkey `action:"2" desc:"Show last mails dialog"`
	ShortkeyCheck      *cdtype.Shortkey `action:"3" desc:"Check now"`

	MailClientAction int
	MailClientName   string
	MailClientClass  string

	// Still hidden.
	PollingEnabled bool

	// Defaults are currently added to the last tab of config. This could evolve,
	// but atm, this sound like a sane choice to have something consistent. All
	// values that would be hardcoded are grouped here so we have a good overview
	// of what is used (const & var). And in the conf file, we have all possibly
	// tweakable or fixable values.
	DefaultMonitorName string // Default application or webpage to open.
}

//----------------------------------------------------------[ ACTIONS & MENU ]--

// List of actions defined in this applet.
// Actions order in this list must match the order in defineActions.
// The reference in shortkey declaration must also match.
//
const (
	ActionNone = iota
	ActionOpenClient
	ActionShowMails
	ActionCheckMail
	ActionRegister
)

// Actions available in full menu.
//
var menuFull = []int{
	ActionOpenClient,
	ActionShowMails,
	ActionCheckMail,
	ActionNone,
	ActionRegister,
}

// Actions available in register menu. Displayed when account isn't set.
//
var menuRegister = []int{
	ActionRegister,
}
