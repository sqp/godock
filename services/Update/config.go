package Update

import (
	"github.com/sqp/godock/libs/cdtype"
)

const (
	// EmblemVersion defines the position of the "new version" emblem.
	EmblemVersion = cdtype.EmblemBottomLeft

	// EmblemTarget defines the position of the "current target" emblem.
	EmblemTarget = cdtype.EmblemTopLeft
)

const defaultVersionPollingTimer = 60

var (
	// LocationLaunchpad defines the launchpad url that doesn't require auth to download.
	LocationLaunchpad = "https+urllib://launchpad.net/"

	// CmdBzr defines the command to get dock sources versions and upgrade.
	CmdBzr = "bzr"
)

//
//------------------------------------------------------------------[ CONFIG ]--

// The full config data. One struct for each tab.
//
type updateConf struct {
	groupIcon          `group:"Icon"`
	groupConfiguration `group:"Configuration"`
	groupActions       `group:"Actions"`
}

type groupIcon struct {
	Icon string `conf:"icon"`
	Name string `conf:"name"`
}

type groupConfiguration struct {
	UserMode bool // false = tester / true = developer

	VersionPollingEnabled bool
	VersionPollingTimer   int

	SourceDir string // base cairo-dock sources directory. Must contain core and plug-ins folders, same as bzr script.

	BuildAppletName string // single applet target. Must be the dir name of the applet.
	BuildOneMode    bool   // false = core / true = applet(s)
	BuildReload     bool   // true if the reload action should be triggered after build

	DiffCommand   string // Command to launch on Show diff action.
	DiffMonitored bool   // true if the diff command application should be monitored (like a launcher).

	BuildTargets []string // list of buildable targets.
}

type groupActions struct {
	TesterClickLeft   string // tester action.
	TesterClickMiddle string // tester action.
	DevClickLeft      string // dev action.
	DevClickMiddle    string // dev action.
	DevMouseWheel     string // dev action.

	ShortkeyOneAction string // shortcut, all actions provided.
	ShortkeyOneKey    string // key binded.

	ShortkeyTwoAction string // shortcut, all actions provided.
	ShortkeyTwoKey    string // key binded.

	// still hidden
	VersionDialogTimer    int
	VersionDialogTemplate string // both file name and template name inside.
	VersionEmblemWork     string
	VersionEmblemNew      string
	IconMissing           string

	CommandSudo string

	DirCore           string
	DirApplets        string
	BranchCore        string
	BranchApplets     string
	ScriptName        string
	ScriptLocation    string
	LocationLaunchpad string

	Debug bool
	// Debug                 int // debug level. Still unused.
}

//
//------------------------------------------------------[ ACTIONS DEFINITION ]--

// List of actions defined in this applet.
// Actions order in this list must match the order in defineActions.
//
const (
	ActionNone = iota
	ActionShowDiff
	ActionShowVersions
	ActionCheckVersions
	ActionCycleTarget
	ActionToggleUserMode
	ActionToggleReload
	ActionBuildTarget
	//~ 	GENERATE_REPORT // TODO
	// ActionBuildAll
	// ActionDownloadCore
	// ActionDownloadApplets
	// ActionDownloadAll
	// ActionUpdateAll
)

//
//--------------------------------------------------------------[ USER MENUS ]--

// Actions available in tester menu.
//
var menuTester = []int{
	ActionShowVersions,
	// ActionNone,
	// ActionUpdateAll,
	// ActionNone,
	// ActionDownloadAll,
	// ActionBuildAll,
	ActionNone,
	ActionToggleUserMode,
}

// Actions available in developer menu.
//
var menuDev = []int{
	ActionCycleTarget,
	// ActionSetAppletName,
	ActionNone,
	ActionShowDiff,
	ActionShowVersions,
	ActionNone,
	ActionBuildTarget,
	// ActionNone,
	// ActionDownloadCore,
	// ActionDownloadApplets,
	// ActionDownloadAll,
	ActionNone,
	ActionToggleReload,
	ActionToggleUserMode,
}
