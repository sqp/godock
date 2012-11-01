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

// List of actions defined in this applet.
//
const (
	ActionNone       = "none"
	ActionOpenClient = "Open mail client"
	ActionCheckMail  = "Check now"
	ActionShowMails  = "Show mail dialog"
)

//~ self.svgpath = self.path+'emblem.svg' # SVG emblem file

//------------------------------------------------------------------[ CONFIG ]--

// Global struct conf.
//
type mailConf struct {
	MailIcon `tab:"Icon"`
	MailConfig
	MailActions
	//~ gconfig
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
	FeedGmail      string // Url of the Atom feed source.
	//~ DebugLevel int // unused

	// Defaults are currently added to the last tab of config. This could evolve,
	// but atm, this sound like a sane choice to have something consistant. All
	// values that would be hardcoded are grouped here so we have a good overview
	// of what is used (const & var). And in the conf file, we have all possibly
	// tweakable or fixable values.
	DefaultMonitorName    string // Default application or webpage to open.
	DefaultAlertSoundFile string // 
}

//-----------------------------------------------------------------[ ACTIONS ]--

/*  UNUSED YET


ActionOpenClient = "Open mail client"
	ActionCheckMail  = "Check now"
	ActionShowMails

// Define applet actions.
//
func (app *AppletGmail) defineActions() {
	app.AddAction(
		&dock.Action{
			Id:       ActionNone,
			Icontype: 2,
		},
		&dock.Action{
			Id:   ActionOpenClient,
			Name: "Open mail client",
			Icon: "gtk-open",
			//~ Icontype:
			Call: func() { app.actionOpenClient() },
		},
		&dock.Action{
			Id:       ActionCheckMail,
			Name:     "Check now",
			Icon:     "gtk-refresh",
			Call:     func() { app.action(ActionCheckMail) },
			Threaded: true,
		},
		&dock.Action{
			Id:       ActionShowMails,
			Name:     "Show mail dialog",
			Icon:     "gtk-media-forward",
			Call:     func() { app.action(ActionShowMails) },
			Threaded: true,
		},
	)
}


//--------------------------------------------------------------------[ MENU ]--

// Actions available in menu.
//
var menu []int = []int{
}

*/
