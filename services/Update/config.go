/* Update is an applet for Cairo-Dock to check for its new versions and do updates.

Copyright : (C) 2012 by SQP
E-mail : sqp@glx-dock.org

This program is free software; you can redistribute it and/or
modify it under the terms of the GNU General Public License
as published by the Free Software Foundation; either version 3
of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.
http://www.gnu.org/licenses/licenses.html#GPL
*/
package Update

import (
	"github.com/sqp/godock/libs/cdtype"

	"errors"
)

const (
	EmblemVersion = cdtype.EmblemBottomLeft
	EmblemTarget  = cdtype.EmblemTopLeft
	// EmblemAction  = cdtype.EmblemTopRight
)

const defaultVersionPollingTimer = 60

var (
	LocationLaunchpad string = "https+urllib://launchpad.net/"
	cmdSudo           string = "gksudo"
	cmdBzr            string = "bzr"

	inProgress error = errors.New("not finished")
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

	BuildTargets string // list of buildable targets.
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
var menuTester []int = []int{
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
var menuDev []int = []int{
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
