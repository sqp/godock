/*
Package Update is an applet for Cairo-Dock to build and update the dock and applets.

Play with cairo-dock sources:
download/update, compile, restart dock... Usefull for developers, testers and
users who want to stay up to date, or maybe on a distro without packages.
*/
package Update

/*
Copyright : (C) 2012 by SQP
E-mail    : sqp@glx-dock.org

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

import (
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/dock" // Connection to cairo-dock.
	"github.com/sqp/godock/libs/packages/build"

	"os"
	"path"
	"strconv"
	"strings"
)

var log cdtype.Logger

//------------------------------------------------------------------[ APPLET ]--

// Applet data and controlers.
//
type Applet struct {
	cdtype.AppBase // Applet base and dock connection.

	conf    *updateConf // applet user configuration.
	version *Versions   // applet data.
	target  BuildTarget // build from sources interface.

	targetID int // position of current target in BuildTargets list.
	err      error
}

// NewApplet create an new Update applet instance.
//
func NewApplet() cdtype.AppInstance {
	app := &Applet{AppBase: dock.NewCDApplet()}
	app.defineActions()

	// Create a cairo-dock sources version checker.
	app.version = &Versions{
		callResult: app.onGotVersions,
		newCommits: -1,
	}

	// The poller will check for new versions on a timer.
	poller := app.AddPoller(app.version.Check)

	// Set "working" emblem during version check. It should be removed or changed by the check.
	poller.SetPreCheck(func() { app.SetEmblem(app.FileLocation("img", app.conf.VersionEmblemWork), EmblemVersion) })

	log = app.Log()
	build.Log = log

	return app
}

// Init load user configuration if needed and initialise applet.
//
func (app *Applet) Init(loadConf bool) {
	app.LoadConfig(loadConf, &app.conf) // Load config will crash if fail. Expected.

	// Icon default settings.
	app.SetDefaults(cdtype.Defaults{
		Icon:           app.conf.Icon,
		Label:          app.conf.Name,
		Templates:      []string{app.conf.VersionDialogTemplate},
		PollerInterval: cdtype.PollerInterval(app.conf.VersionPollingTimer*60, defaultVersionPollingTimer*60),
		Commands: cdtype.Commands{
			"showDiff": cdtype.NewCommand(app.conf.DiffMonitored, app.conf.DiffCommand)},
		Shortkeys: []cdtype.Shortkey{
			{"Actions", "ShortkeyOneKey", "Action one", app.conf.ShortkeyOneKey},
			{"Actions", "ShortkeyTwoKey", "Action two", app.conf.ShortkeyTwoKey}},
		Debug: app.conf.Debug})

	if app.conf.VersionPollingEnabled {
		app.Poller().Start()
	} else {
		app.Poller().Stop()
	}

	// Branches for versions checking.
	app.version.sources = []*Branch{
		NewBranch(app.conf.BranchCore, path.Join(app.conf.SourceDir, app.conf.DirCore)),
		NewBranch(app.conf.BranchApplets, path.Join(app.conf.SourceDir, app.conf.DirApplets))}

	// Build targets. Allow actions on sources and displays emblem on top left for togglable target.
	app.setBuildTarget()

	// Build globals.
	LocationLaunchpad = app.conf.LocationLaunchpad
	build.CmdSudo = app.conf.CommandSudo

	// Set booleans references for menu checkboxes.
	app.ActionSetBool(ActionToggleUserMode, &app.conf.UserMode)
	app.ActionSetBool(ActionToggleReload, &app.conf.BuildReload)
}

//------------------------------------------------------------------[ EVENTS ]--

// DefineEvents set applet events callbacks.
//
func (app *Applet) DefineEvents(events *cdtype.Events) {

	// Left click: launch configured action for current user mode.
	//
	events.OnClick = func() {
		if app.conf.UserMode {
			app.ActionLaunch(app.ActionID(app.conf.DevClickLeft))
		} else {
			app.ActionLaunch(app.ActionID(app.conf.TesterClickLeft))
		}
	}

	// Middle click: launch configured action for current user mode.
	//
	events.OnMiddleClick = func() {
		if app.conf.UserMode {
			app.ActionLaunch(app.ActionID(app.conf.DevClickMiddle))
		} else {
			app.ActionLaunch(app.ActionID(app.conf.TesterClickMiddle))
		}
	}

	// Right click menu: show menu for current user mode.
	//
	events.OnBuildMenu = func(menu cdtype.Menuer) {
		if app.conf.UserMode {
			app.BuildMenu(menu, menuDev)
		} else {
			app.BuildMenu(menu, menuTester)
		}
	}

	// Scroll event: launch configured action if in dev mode.
	//
	events.OnScroll = func(scrollUp bool) {
		log.Info("scroll", app.conf.UserMode, app.ActionCount(), app.ActionID(app.conf.DevMouseWheel))
		if !app.conf.UserMode || app.ActionCount() > 0 { // Wheel action only for dev and if no threaded tasks running.
			return
		}
		id := app.ActionID(app.conf.DevMouseWheel)
		if id == ActionCycleTarget { // Cycle depends on wheel direction.
			if scrollUp {
				app.cycleTarget(1)
			} else {
				app.cycleTarget(-1)
			}
		} else { // Other actions are simple toggle.
			app.ActionLaunch(id)
		}
	}

	// Shortkey event: launch configured action.
	//
	events.OnShortkey = func(key string) {
		if key == app.conf.ShortkeyOneKey {
			app.ActionLaunch(app.ActionID(app.conf.ShortkeyOneAction))
		}
		if key == app.conf.ShortkeyTwoKey {
			app.ActionLaunch(app.ActionID(app.conf.ShortkeyTwoAction))
		}
	}

	// Feature to test: rgrep of the dropped string on the source dir.
	//
	events.OnDropData = func(data string) {
		log.Info("Grep " + data)
		// log.ExecShow("grep", "-r", "--color", data, app.ShareDataDir)
	}
}

//----------------------------------------------------------------[ CALLBACK ]--

// Got versions informations, Need to set a new emblem
//
func (app *Applet) onGotVersions(new int, e error) {
	if new > 0 {
		app.SetEmblem(app.FileLocation("img", app.conf.VersionEmblemNew), EmblemVersion)

		if app.version.newCommits != -1 && new > app.version.newCommits { // Drop first message and only show others if number changed.
			app.actionShowVersions(false)
		}

	} else {
		app.SetEmblem("none", EmblemVersion)
	}
	app.version.newCommits = new
}

//-----------------------------------------------------------------[ ACTIONS ]--

// Define applet actions.
//
func (app *Applet) defineActions() {
	app.ActionSetMax(1)
	app.ActionAdd(
		&cdtype.Action{
			ID:   ActionNone,
			Menu: cdtype.MenuSeparator,
		},
		&cdtype.Action{
			ID:   ActionShowDiff,
			Name: "Show diff",
			Icon: "gtk-justify-fill",
			Call: app.actionShowDiff,
		},
		&cdtype.Action{
			ID:       ActionShowVersions,
			Name:     "Show versions",
			Icon:     "gtk-network", // to change
			Call:     func() { app.actionShowVersions(true) },
			Threaded: true,
		},

		&cdtype.Action{
			ID:       ActionCheckVersions,
			Name:     "Check versions",
			Icon:     "gtk-network",
			Call:     app.actionCheckVersions,
			Threaded: true,
		},
		&cdtype.Action{
			ID:       ActionCycleTarget,
			Name:     "Cycle target",
			Icon:     "gtk-refresh",
			Call:     func() { go app.cycleTarget(1) }, // async as it require a dbus query (need ask and answer in internal mode).
			Threaded: true,
		},
		&cdtype.Action{
			ID:   ActionToggleUserMode,
			Name: "Dev mode",
			Menu: cdtype.MenuCheckBox,
			Call: app.actionToggleUserMode,
		},

		&cdtype.Action{
			ID:   ActionToggleReload,
			Name: "Reload target",
			Menu: cdtype.MenuCheckBox,
			Call: app.actionToggleReload,
		},
		&cdtype.Action{
			ID:       ActionBuildTarget,
			Name:     "Build target",
			Icon:     "gtk-media-play",
			Call:     app.actionBuildTarget,
			Threaded: true,
		},
		//~ action_add(CDCairoBzrAction.GENERATE_REPORT, action_none, "", "gtk-refresh");

		// &cdtype.Action{
		// 	ID:       ActionBuildAll,
		// 	Name:     "Build All",
		// 	Icon:     "gtk-media-next",
		// 	Call:     func() { app.actionBuildAll() },
		// 	Threaded: true,
		// },
		// &cdtype.Action{
		// 	ID:       ActionDownloadCore,
		// 	Name:     "Download Core",
		// 	Icon:     "gtk-network",
		// 	Call:     func() { app.actionDownloadCore() },
		// 	Threaded: true,
		// },
		// &cdtype.Action{
		// 	ID:       ActionDownloadApplets,
		// 	Name:     "Download Plug-Ins",
		// 	Icon:     "gtk-network",
		// 	Call:     func() { app.actionDownloadApplets() },
		// 	Threaded: true,
		// },
		// &cdtype.Action{
		// 	ID:       ActionDownloadAll,
		// 	Name:     "Download All",
		// 	Icon:     "gtk-network",
		// 	Call:     func() { app.actionDownloadAll() },
		// 	Threaded: true,
		// },
		&cdtype.Action{
			ID:       ActionUpdateAll,
			Name:     "Update All",
			Icon:     "gtk-network",
			Call:     app.actionUpdateAll,
			Threaded: true,
		},
	)
}

//-----------------------------------------------------------[ BASIC ACTIONS ]--

// Open diff command, or toggle window visibility if application is monitored and opened.
//
func (app *Applet) actionShowDiff() {
	haveMonitor, hasFocus := app.HaveMonitor()
	switch {
	case app.conf.DiffMonitored && haveMonitor: // Application monitored and open.
		app.ShowAppli(!hasFocus)

	default: // Launch application.
		if _, e := os.Stat(app.target.SourceDir()); e != nil {
			log.NewWarn("Invalid source directory", "ShowDiff")
		} else {
			log.ExecAsync(app.conf.DiffCommand, app.target.SourceDir())
		}
	}
}

// Change target and display the new one.
//
func (app *Applet) cycleTarget(delta int) {
	app.targetID += delta
	switch {
	case app.targetID >= len(app.conf.BuildTargets):
		app.targetID = 0

	case app.targetID < 0:
		app.targetID = len(app.conf.BuildTargets) - 1
	}

	app.setBuildTarget()
	app.showTarget()
}

func (app *Applet) actionToggleUserMode() {
	app.conf.UserMode = !app.conf.UserMode
	app.setBuildTarget()
}

func (app *Applet) actionToggleReload() {
	app.conf.BuildReload = !app.conf.BuildReload
}

//--------------------------------------------------------[ THREADED ACTIONS ]--

// Check new versions now and reset timer.
//
func (app *Applet) actionCheckVersions() {
	app.Poller().Restart()
}

// To improve : parse http://bazaar.launchpad.net/~cairo-dock-team/cairo-dock-core/cairo-dock/changes/
// and maybe see to use as download tool : http://golang.org/src/cmd/go/vcs.go
//
func (app *Applet) actionShowVersions(force bool) {
	for _, v := range app.version.Sources() {
		if v.Delta > 0 {
			force = true
		}
	}
	if force {
		text, e := app.ExecuteTemplate(app.conf.VersionDialogTemplate, app.conf.VersionDialogTemplate, app.version.Sources())
		if log.Err(e, "template "+app.conf.VersionDialogTemplate) {
			return
		}

		app.PopupDialog(cdtype.DialogData{
			Message:    text,
			TimeLength: app.conf.VersionDialogTimer,
			UseMarkup:  true,
			// Buttons:    "gtk-open;cancel",
		})
		log.Err(e, "popup")
	}
}

// Build current target.
//
func (app *Applet) actionBuildTarget() {
	app.AddDataRenderer("progressbar", 1, "")
	defer app.AddDataRenderer("progressbar", 0, "")

	// app.Animate("busy", 200)
	if !log.Err(app.target.Build(), "Build") {
		log.Info("Build", app.target.Label())
		app.restartTarget()
	}
}

// func (app *Applet) actionBuildCore()       {}
// func (app *Applet) actionBuildApplets()    {}
func (app *Applet) actionBuildAll()        {}
func (app *Applet) actionDownloadCore()    {}
func (app *Applet) actionDownloadApplets() {}
func (app *Applet) actionDownloadAll()     {}

func (app *Applet) actionUpdateAll() {
	app.AddDataRenderer("progressbar", 1, "")
	defer app.AddDataRenderer("progressbar", 0, "")

	log.Info("downloading core")
	_, e := app.version.sources[0].update()
	if log.Err(e, "update core") {
		return
	}
	log.Info("updating core")
	core := &build.BuilderCore{}
	core.SetDir(app.conf.SourceDir)
	core.SetProgress(func(f float64) { app.RenderValues(f) })
	log.Err(core.Build(), "build core")

	log.Info("downloading applets")
	_, e = app.version.sources[1].update()
	if log.Err(e, "update applets") {
		return
	}
	log.Info("updating applets")
	applets := &build.BuilderApplets{}
	applets.SetDir(app.conf.SourceDir)
	applets.SetProgress(func(f float64) { app.RenderValues(f) })

	applets.MakeFlags = "-Denable-Logout=no" // "-Denable-gmenu=no"

	log.Err(applets.Build(), "build applets")

	app.Poller().Restart()
}

//------------------------------------------------------------------[ COMMON ]--

// Get numeric part of a string and convert it to int.
//
// func trimInt(imdb string) (int, error) {
// 	//~ Replace, _ := regexp.CompilePOSIX("^.*([:digit:]*).*$")
// 	Replace, _ := regexp.Compile("[0-9]+")
// 	str := Replace.FindString(imdb)
// 	return strconv.Atoi(str)
// }

func trimInt(str string) (int, error) {
	return strconv.Atoi(strings.Trim(str, " \n"))
}
