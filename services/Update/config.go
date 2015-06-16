package Update

import (
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/log/color"
)

const (
	// EmblemVersion defines the position of the "new version" emblem.
	EmblemVersion = cdtype.EmblemBottomLeft

	// EmblemTarget defines the position of the "current target" emblem.
	EmblemTarget = cdtype.EmblemTopLeft
)

const defaultVersionPollingTimer = 60

var (
	grepCmdArgs = []string{"-r", "-I"} // -r: recursive, -I: ignore binaries.

	// Grep text format.
	grepTitlePattern   = "\n   ---[ grep %s ]---\n"
	grepTitleFormatter = color.Yellow
	grepFileFormatter  = color.Green
	grepQueryFormatter = color.Yellow
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

	CommandSudo  string
	FlagsApplets string // for the full applets pack, to help enable or disable them.

	DirCore       string
	DirApplets    string
	BranchCore    string
	BranchApplets string

	SourceExtra string // additional repos to version check, separated by \n.

	Debug bool
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
	ActionGrepTarget
	ActionCycleTarget
	ActionToggleUserMode
	ActionToggleReload
	ActionBuildTarget
	ActionUpdateAll
	ActionDownloadOthers
	//~ 	GENERATE_REPORT // TODO
	// ActionBuildAll
	// ActionDownloadCore
	// ActionDownloadApplets
	// ActionDownloadAll
)

//
//--------------------------------------------------------------[ USER MENUS ]--

// Actions available in tester menu.
//
var menuTester = []int{
	ActionShowVersions,
	ActionNone,
	ActionUpdateAll,
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
	// ActionToggleUserMode,
	ActionUpdateAll,
}
