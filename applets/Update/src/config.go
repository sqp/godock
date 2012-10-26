/* Update is an applet for Cairo-Dock to check for its new versions and do update.

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
http://www.gnu.org/licenses/licenses.html#GPL */

package main

import dock "github.com/sqp/godock/libs/dbus"

const defaultVersionPollingTimer = 5

//------------------------------------------------------------------------------
// CONFIG.
//------------------------------------------------------------------------------

// The full config data. One struct for each tab.
//
type UpdateConf struct {
	updateConfig
	updateDevel
	updateHidden
}

// Tab Configuration.
//
type updateConfig struct {
	UserMode          bool /// false = tester / true = developer
	TesterClickLeft   int
	TesterClickMiddle int
	SourceDir         string
}

// Tab Developer.
//
type updateDevel struct {
	DevClickLeft    int
	DevClickMiddle  int
	DevMouseWheel   int
	BuildAppletName string
	BuildOneMode    bool   // false = core / true = applet(s)
	BuildReload     bool   // true if the reload action should be triggered after build
	DiffCommand     string // Command to launch on Show diff action.
	DiffMonitored   bool   // true if the diff command application should be monitored (like a launcher).

	ShortkeyOneAction int
	ShortkeyTwoAction int
	ShortkeyOneKey    string
	ShortkeyTwoKey    string
}

// Tab Configuration.
//
type updateHidden struct {
	Debug               int /// false = tester / true = developer
	VersionPollingTimer int
	VersionDialogTimer  int
	VersionEmblemWork   string
	VersionEmblemNew    string
	BuildEmblemWork     string
	DirCore             string
	DirApplets          string
	BranchCore          string
	BranchApplets       string
}

// Config loading.
//
func (app *AppletUpdate) getConfig() {
	app.conf = &UpdateConf{}
	loaded, e := dock.NewConfig(app.ConfFile)
	testFatal(e)
	loaded.Parse("Configuration", updateConfig{}, &app.conf.updateConfig)
	loaded.Parse("Developer", updateDevel{}, &app.conf.updateDevel)
	loaded.Parse("Hidden", updateHidden{}, &app.conf.updateHidden)
}

//------------------------------------------------------------------------------
// ACTIONS DEFINITION.
//------------------------------------------------------------------------------

// List of actions defined in this applet.
// The config options "dev left click" and "dev middle click" must match with this list.
//
const (
	ActionNone = iota
	ActionShowDiff
	ActionShowVersions
	ActionToggleTarget
	ActionToggleUserMode
	ActionToggleReload
	ActionSetAppletName
	//~ 	GENERATE_REPORT // TODO
	ActionBuildTarget
	ActionBuildOne
	ActionBuildCore
	ActionBuildApplets
	ActionBuildAll
	ActionDownloadCore
	ActionDownloadApplets
	ActionDownloadAll
	ActionUpdateAll
)

// Define applet actions.
//
func (app *AppletUpdate) defineActions() {
	app.AddAction(
		&dock.Action{
			Id:       ActionNone,
			Icontype: 2,
		},
		&dock.Action{
			Id:   ActionShowDiff,
			Name: "Show Diff",
			Icon: "gtk-justify-fill",
			//~ Icontype:
			Call: func() { app.actionShowDiff() },
		},
		&dock.Action{
			Id:       ActionShowVersions,
			Name:     "Show Versions",
			Icon:     "gtk-network",
			Call:     func() { app.actionShowVersions() },
			Threaded: true,
		},
		&dock.Action{
			Id:   ActionToggleTarget,
			Icon: "gtk-refresh",
			Call: func() { app.actionToggleTarget() },
		},
		&dock.Action{
			Id:       ActionToggleUserMode,
			Name:     "Use developer mode",
			Icontype: 3,
			Call:     func() { app.actionToggleUserMode() },
		},

		&dock.Action{
			Id:       ActionToggleReload,
			Name:     "Reload after build",
			Icontype: 3,
			Call:     func() { app.actionToggleReload() },
		},
		&dock.Action{
			Id:   ActionSetAppletName,
			Name: "Set target applet",
			Icon: "gtk-refresh",
			Call: func() { app.actionSetAppletName() },
		},
		&dock.Action{
			Id:       ActionBuildTarget,
			Icon:     "gtk-media-play",
			Call:     func() { app.actionBuildTarget() },
			Threaded: true,
		},
		//~ action_add(CDCairoBzrAction.GENERATE_REPORT, action_none, "", "gtk-refresh");

		&dock.Action{
			Id:       ActionBuildOne,
			Icon:     "gtk-media-play",
			Call:     func() { app.actionBuildOne() },
			Threaded: true,
		},
		&dock.Action{
			Id:       ActionBuildCore,
			Name:     "Build Core",
			Icon:     "gtk-media-forward",
			Call:     func() { app.actionBuildCore() },
			Threaded: true,
		},
		&dock.Action{Id: ActionBuildApplets,
			Name:     "Build Applets",
			Icon:     "gtk-media-next",
			Call:     func() { app.actionBuildApplets() },
			Threaded: true,
		},
		&dock.Action{
			Id:       ActionBuildAll,
			Name:     "Build All",
			Icon:     "gtk-media-next",
			Call:     func() { app.actionBuildAll() },
			Threaded: true,
		},
		&dock.Action{
			Id:       ActionDownloadCore,
			Name:     "Download Core",
			Icon:     "gtk-network",
			Call:     func() { app.actionDownloadCore() },
			Threaded: true,
		},
		&dock.Action{
			Id:       ActionDownloadApplets,
			Name:     "Download Plug-Ins",
			Icon:     "gtk-network",
			Call:     func() { app.actionDownloadApplets() },
			Threaded: true,
		},
		&dock.Action{
			Id:       ActionDownloadAll,
			Name:     "Download All",
			Icon:     "gtk-network",
			Call:     func() { app.actionDownloadAll() },
			Threaded: true,
		},
		&dock.Action{
			Id:       ActionUpdateAll,
			Name:     "Update All",
			Icon:     "gtk-network",
			Call:     func() { app.actionUpdateAll() },
			Threaded: true,
		},
	)
}

//------------------------------------------------------------------------------
// USER ACTIONS AND MENUS.
//------------------------------------------------------------------------------

// Actions available to tester on clicks.
// Devs don't need this, they get the full actions list in the menu.
// Must match with the config options "tester left click" and "tester middle click".
//
var actionsClickTester []int = []int{
	ActionNone,
	ActionShowVersions,
	ActionDownloadAll,
	ActionBuildAll,
	ActionUpdateAll,
}

// Actions available to dev on mouse wheel.
// Users don't have this as there is nothing to provide to them atm.
// Must match with the config option "dev mouse wheel".
//
var actionsDevWheel []int = []int{
	ActionNone,
	ActionToggleTarget,
}

// Actions available in tester menu.
//
var menuTester []int = []int{
	ActionShowVersions,
	ActionNone,
	ActionUpdateAll,
	ActionNone,
	ActionDownloadAll,
	ActionBuildAll,
	ActionNone,
	ActionToggleUserMode,
}

// Actions available in developer menu.
//
var menuDev []int = []int{
	ActionToggleTarget,
	ActionSetAppletName,
	ActionNone,
	ActionShowDiff,
	ActionShowVersions,
	ActionNone,
	ActionBuildOne,
	ActionBuildCore,
	ActionBuildApplets,
	ActionNone,
	ActionDownloadCore,
	ActionDownloadApplets,
	ActionDownloadAll,
	ActionNone,
	ActionToggleReload,
	ActionToggleUserMode,
}
