/* Update is an applet for Cairo-Dock to check for its new versions and do update.

Install go and get go environment: you need a valid $GOPATH var and directory.

Download, build and install to your Cairo-Dock external applets dir:
	go get github.com/sqp/godock/applets/update
	go build $GOPATH/src/github.com/sqp/godock/applets/Update
	ln -s $GOPATH/src/github.com/sqp/godock/applets/Update/  ~/.config/cairo-dock/third-party


TODO: Version checking:
* connect the restart to reset the timer if user asked for a recheck.
* get a better bzr result than simple revno if the user is on a different branch with an other stack of patches. (need to get the split version to know the real number of missing patches)



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

import (
	dock "github.com/sqp/godock/libs/dbus"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"term"
	"time"
)

//------------------------------------------------------------------------------
// Applet interfaces.
//------------------------------------------------------------------------------

//------------------------------------------------------------------------------
// MAIN CALL.
//------------------------------------------------------------------------------

// Program launched. Create and activate applet.
//
func main() {
	app := Update()

	log.SetPrefix(term.Yellow("[" + app.AppletName + "] "))
	log.Println(term.Green("Applet started"))
	defer log.Println(term.Yellow("Applet stopped"))

	// Start main loop and wait for events. Until the End signal is received from the dock.
	ticker, restart := app.poller.NewTicker()
	for { // Main loop, waiting for events.
		select {
		case <-app.Close: // That's all folks.
			return

		// not triggered yet.
		case <-restart: // Want to recheck now ?
			log.Println(term.Yellow("timer restart"))
			ticker, restart = app.poller.NewTicker()

		case <-ticker.C: // It's time to work !
			//~ log.Println(term.Yellow("timer"))
			app.poller.Check()
		}
	}
}

//------------------------------------------------------------------------------
// APPLET.
//------------------------------------------------------------------------------

type AppletUpdate struct {
	*dock.CDApplet // Dock interface.

	conf    *UpdateConf  // applet user configuration.
	poller  *dock.Poller // applet polling routine.
	version *Versions    // applet data.
	target  BuildTarget  // build from sources interface.

	// Local variables for the app.
	err error
}

// Create an instance of applet Update. Used to play with cairo-dock sources:
// download/update, compile, restart dock... Usefull for developers, testers
// and users who want to stay up to date, or maybe on a distro without packages.
//
func Update() *AppletUpdate {
	app := &AppletUpdate{
		CDApplet: dock.Applet(),
	}

	// Define and connect events to the dock
	app.defineEvents()
	app.ConnectToBus()

	// Action indicators: display emblem while busy..
	onStarted := func() { app.SetEmblem(app.FileLocation("img", app.conf.BuildEmblemWork), dock.EmblemTopRight) }
	onFinished := func() { app.SetEmblem("none", dock.EmblemTopRight) }
	app.SetActionIndicators(onStarted, onFinished)

	app.defineActions()
	app.getConfig()

	// Create a cairo-dock sources version checker.
	app.version = &Versions{
		callResult: func(new int, e error) { app.onGotVersions(new, e) },
		template:   LoadTemplate(app.FileLocation("templates", "ShowVersionsDialog.tmpl")),
		sources: []*Branch{
			NewBranch(app.conf.BranchCore, path.Join(app.conf.SourceDir, app.conf.DirCore)),
			NewBranch(app.conf.BranchApplets, path.Join(app.conf.SourceDir, app.conf.DirApplets)),
		},
	}

	// Prepare updates callbacks during and after polling. Display a small emblem
	// during the polling, result will be updated directly.
	onStarted = func() { app.SetEmblem(app.FileLocation("img", app.conf.VersionEmblemWork), dock.EmblemBottomLeft) }

	// Version checking polling timer.
	onTimer := func() { app.version.Check() }
	app.poller = dock.NewPoller(onTimer)
	app.poller.SetCallDisplay(onStarted, nil)

	//~ app.data.SetRestart(app.poller.restart)

	app.Init()
	return app
}

// Initialise applet with user configuration.
//
func (app *AppletUpdate) Init() {
	// Polling timer is mandatory.
	if app.conf.VersionPollingTimer > 0 {
		app.poller.SetInterval(app.conf.VersionPollingTimer)
	} else {
		app.poller.SetInterval(defaultVersionPollingTimer)
	}

	// need set [Icon]icon 

	// Build target renderer. Displays an emblem on top left.
	app.setBuildTarget()

	def := dock.Defaults{
		Shortkeys: []string{app.conf.ShortkeyOneKey, app.conf.ShortkeyTwoKey},
	}

	if app.conf.DiffMonitored {
		def.MonitorName = app.conf.DiffCommand
	}
	app.SetDefaults(def)

	// Delayed display of emblem. 5ms seemed to be enough but 500 should do the job.
	go func() { time.Sleep(500 * time.Millisecond); app.showTarget() }()
}

//------------------------------------------------------------------------------
// EVENTS.
//------------------------------------------------------------------------------

// Define applet events callbacks.
//
func (app *AppletUpdate) defineEvents() {

	// Left click: launch configured action for current user mode.
	//
	app.Events.OnClick = func() {
		if app.conf.UserMode {
			app.Launch(app.conf.DevClickLeft)
		} else {
			app.Launch(menuTester[app.conf.TesterClickLeft])
		}
	}

	// Middle click: launch configured action for current user mode.
	//
	app.Events.OnMiddleClick = func() {
		if app.conf.UserMode {
			app.Launch(app.conf.DevClickMiddle)
		} else {
			app.Launch(actionsClickTester[app.conf.TesterClickMiddle])
		}
	}

	// Right click menu: show menu for current user mode.
	//
	app.Events.OnBuildMenu = func() {
		menu := []string{""} // First entry is a separator.
		//~ if gmail.data.IsValid() {
		menu = append(menu, "ActionShowVersions")
		//~ } else { // No running loop =  no registration. User will do as expected !
		//~ menu = append(menu, "Set account")
		//~ }
		app.PopulateMenu(menu...)
		//~ build_menu ( this.config.bDevMode ? (CDCairoBzrAction[]) CAIROBZR_MENU_DEV : (CDCairoBzrAction[]) CAIROBZR_MENU_TESTER );

	}

	// Menu entry selected. Launch the expected action.
	//
	app.Events.OnMenuSelect = func(numEntry int32) {
		app.actionShowVersions()
		//~ switch numEntry {
		//~ case 1:
		//~ gmail.action(ActionOpenClient)
		//~ case 2:
		//~ gmail.action(ActionCheckMail)
		//~ case 3:
		//~ gmail.action(ActionShowMails)
		//~ case 5:
		//~ gmail.askLogin()
		//~ }
	}

	// Scroll event: launch configured action if in dev mode.
	//
	app.Events.OnScroll = func(scrollUp bool) {
		if app.conf.UserMode {
			app.Launch(actionsDevWheel[app.conf.DevMouseWheel])
		}
	}

	// Answer dialog: only question asked atm is the target applet => set new target.
	//
	app.Events.OnAnswerDialog = func(button int32, data interface{}) {
		// save target +
		//~ set_target ((string) answer);
	}

	app.Events.OnShortkey = func(key string) {
		if key == app.conf.ShortkeyOneKey {
			app.Launch(app.conf.ShortkeyOneAction)
		}
		//~ if key == app.conf.ShortkeyTwoKey {
		//~ app.Launch(app.conf.ShortkeyTwoAction)
		//~ }
	}

	/*


	*/

	// Reset all settings and restart timer.
	//
	app.Events.Reload = func(confChanged bool) {
		if confChanged {
			app.getConfig()
		}
		log.Println("should restart")
		app.Init()
		//~ 
		//~ log.Println("init ok")
		//~ 
		//~ app.data.Restart()
	}

	// Nothing to do ? Need to check DBus API.
	//~ app.Events.End = func() {}
}

//------------------------------------------------------------------------------
// BASIC ACTIONS CALLS.
//------------------------------------------------------------------------------

// Open diff command, or toggle window visibility if application is monitored and opened.
//
func (app *AppletUpdate) actionShowDiff() {
	d, e := app.GetAll()
	switch {
	case e == nil && d.Xid > 0 && app.conf.DiffMonitored: // Application monitored & opened.
		app.ShowAppli(!d.HasFocus)
	default: // Application

		// bugs: use target folder.
		// ~ this.spawn_async (this.config.bTarget ? this.config.sFolderPlugIns : this.config.sFolderCore, argv);
		if _, e := os.Stat(path.Join(app.conf.SourceDir, "cairo-dock-core")); e != nil {
			log.Println("invalid source directory")
		} else {
			execAsync(app.conf.DiffCommand, path.Join(app.conf.SourceDir, "cairo-dock-core"))
		}

	}
}

// Change target and display the new one.
//
func (app *AppletUpdate) actionToggleTarget() {
	app.conf.BuildOneMode = !app.conf.BuildOneMode
	app.setBuildTarget()
	app.showTarget()
}

func (app *AppletUpdate) actionToggleUserMode() {
	//~ this.config.bDevMode = this.config.bDevMode == true ? false : true;
	//~ set_icon_info ();
}

func (app *AppletUpdate) actionToggleReload() {
	//~ this.config.bReload = this.config.bReload == true ? false : true;
}

func (app *AppletUpdate) actionSetAppletName() {
	//~ var dialog_attributes = new HashTable<string,Variant>(str_hash, str_equal);
	//~ dialog_attributes.insert ("icon", "stock_properties");
	//~ dialog_attributes.insert ("message", "Set build plugin name");
	//~ dialog_attributes.insert ("buttons", "ok;cancel");
	//~ var widget_attributes = new HashTable<string,Variant>(str_hash, str_equal);
	//~ widget_attributes.insert ("widget-type", "text-entry");
	//~ widget_attributes.insert ("editable", true);
	//~ try { this.icon.PopupDialog (dialog_attributes, widget_attributes); }
	//~ catch (Error e) {}
}

//------------------------------------------------------------------------------
// THREADED ACTIONS CALLS.
//------------------------------------------------------------------------------

func (app *AppletUpdate) actionBuildTarget() {
	//~ action_launch (this.config.bTarget ? CDCairoBzrAction.BUILD_ONE : CDCairoBzrAction.BUILD_CORE);
}

// Build one applet.
//
func (app *AppletUpdate) actionBuildOne() {
	app.target.Build()
}

func (app *AppletUpdate) actionBuildCore()       {}
func (app *AppletUpdate) actionBuildApplets()    {}
func (app *AppletUpdate) actionBuildAll()        {}
func (app *AppletUpdate) actionDownloadCore()    {}
func (app *AppletUpdate) actionDownloadApplets() {}
func (app *AppletUpdate) actionDownloadAll()     {}
func (app *AppletUpdate) actionUpdateAll()       {}

//~ func (app *AppletUpdate) action(action string) {
// No running loop = no registration. User must comply !
//~ if !app.data.IsValid() {
//~ app.askLogin()
//~ return
//~ }
//~ 
//~ switch action {
//~ case ActionOpenClient:
//~ app.actionOpenClient()

//~ }
//~ }

//------------------------------------------------------------------------------
// COMMON.
//------------------------------------------------------------------------------

// Run command with output forwarded to console.
//
func execShow(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	e := cmd.Run()
	//~ if e != nil {
	//~ log.Println(e)
	//~ }
	return e
}

func execSync(command string, args ...string) string {
	out, e := exec.Command(command, args...).Output()
	if logE("execSync: "+command, e) {
		return ""
	}
	//~ term.Info(string(out))
	return string(out)
}

func execAsync(command string, args ...string) error {
	e := exec.Command(command, args...).Start()
	if e != nil {
		log.Println(term.Red("Launch failed error"), e, args)
	}
	return e
}

func logE(action string, e error) (wasErr bool) {
	if e != nil {
		wasErr = true
		log.Println(term.Red("Error"), e)
	}
	return wasErr
}

func testFatal(e error) {
	if e != nil {
		log.Println(term.Red("Applet load error"), e)
		os.Exit(2)
	}
}

func TrimInt(imdb string) (int, error) {
	//~ Replace, _ := regexp.CompilePOSIX("^.*([:digit:]*).*$")
	Replace, _ := regexp.Compile("[0-9]+")
	str := Replace.FindString(imdb)
	return strconv.Atoi(str)
}
