package Update

import (
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/text/color"
)

const (
	// EmblemVersion defines the position of the "new version" emblem.
	EmblemVersion = cdtype.EmblemBottomLeft

	// EmblemTarget defines the position of the "current target" emblem.
	EmblemTarget = cdtype.EmblemTopLeft
)

// Commands references.
const (
	cmdShowDiff = iota
)

var (
	grepCmdArgs = []string{
		"-r", // recursive
		"-I", // ignore binaries.
	}

	// Grep text format.
	grepTitlePattern   = "\n   ---[ grep %s  -  %s ]---\n"
	grepTitleFormatter = color.Yellow
	grepFileFormatter  = color.Green
	grepQueryFormatter = color.Yellow
)

//
//------------------------------------------------------------------[ CONFIG ]--

// The full config data. One struct for each tab.
//
type updateConf struct {
	cdtype.ConfGroupIconBoth `group:"Icon"`
	groupConfiguration       `group:"Configuration"`
	groupActions             `group:"Actions"`
}

type groupConfiguration struct {
	UserMode bool // false = tester / true = developer

	VersionPollingEnabled bool
	VersionPollingTimer   cdtype.Duration `unit:"minute" default:"60" min:"5"`
	DialogDuration        int

	SourceDir   string // base cairo-dock sources directory. Must contain core and plug-ins folders, same as bzr script.
	BuildReload bool   `action:"8"` // true if the reload action should be triggered after build

	DiffCommand   string // Command to launch on Show diff action.
	DiffMonitored bool   // true if the diff command application should be monitored (like a launcher).
	DiffStash     bool   `action:"9"` // true to show the diff vs stash (use git difftool -d).
	CmdOpenSource string

	BuildTargets []string // list of buildable targets.

	VersionDialogTemplate cdtype.Template `default:"dialogversion"`
	VersionEmblemWork     string          `default:"EmblemWork.svg"`
	VersionEmblemNew      string          `default:"EmblemNew.svg"`
	IconMissing           string          `default:"IconMissing.svg"`

	CommandSudo  string
	FlagsApplets string // for the full applets pack, to help enable or disable them.

	DirCore    string
	DirApplets string

	SourceExtra []string // additional repos to version check, separated by \n.
}

type groupActions struct {
	TesterClickLeft   string // tester action.
	TesterClickMiddle string // tester action.
	DevClickLeft      string // dev action.
	DevClickMiddle    string // dev action.
	DevMouseWheel     string // dev action.

	ShortkeyShowDiff       *cdtype.Shortkey `action:"1"`
	ShortkeyShowVersions   *cdtype.Shortkey `action:"2"`
	ShortkeyGrepTarget     *cdtype.Shortkey `action:"4"`
	ShortkeyNextTarget     *cdtype.Shortkey `action:"5"`
	ShortkeyOpenFileTarget *cdtype.Shortkey `action:"6"`
	ShortkeyBuildTarget    *cdtype.Shortkey `action:"10"`
}

//
//------------------------------------------------------[ ACTIONS DEFINITION ]--

// List of actions defined in this applet.
// Actions order in this list must match the order in defineActions.
// The reference in shortkey declaration must also match.
//
const (
	ActionNone = iota
	ActionShowDiff
	ActionShowVersions
	ActionCheckVersions
	ActionGrepTarget
	ActionCycleTarget
	ActionOpenFileTarget
	ActionToggleUserMode
	ActionToggleReload
	ActionToggleDiffStash
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
	ActionToggleDiffStash,
	// ActionToggleUserMode,
	ActionUpdateAll,
}
