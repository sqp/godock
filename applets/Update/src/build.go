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

import (
	"bytes"
	"errors"
	dock "github.com/sqp/godock/libs/dbus"
	"os"
	"os/exec"
	"strconv"
	"term"
	"text/template"

//~ DEBUG "github.com/kr/pretty"
)

//~ func grr() { DEBUG.Printf("") }

var command = "bzr"

//------------------------------------------------------------------------------
// BUILDERS CONFIG.
//------------------------------------------------------------------------------

type BuildType int

const (
	Core BuildType = iota
	Applets
	AppletInternal
	AppletScript
	AppletCompiled
)

type BuildTarget interface {
	Label() string
	Icon() string
	Build()
}

type BuildNull struct{}

func (build *BuildNull) Icon() string  { return "" }
func (build *BuildNull) Label() string { return "" }
func (build *BuildNull) Build()        {}

type BuildCore struct{}

func (build *BuildCore) Icon() string {
	return "/usr/share/cairo-dock/cairo-dock.svg"
}

func (build *BuildCore) Label() string {
	return "Core"
}

func (build BuildCore) Dir() string {
	return "Core"
}

func (build *BuildCore) Build() {
}

type BuildCompiled struct {
	module string
	icon   string
	build  func()
}

func (build *BuildCompiled) Icon() string {
	return build.icon
}

func (build *BuildCompiled) Label() string {
	return build.module
}

func (build *BuildCompiled) Build() {
	build.build()
}

//------------------------------------------------------------------------------
// SOURCES BUILDER.
//------------------------------------------------------------------------------

func (app *AppletUpdate) setBuildTarget() {
	// Create the renderer.
	if app.conf.BuildOneMode { // One applet mode.
		switch app.appletType(app.conf.BuildAppletName) {

		case AppletCompiled:
			app.target = &BuildCompiled{
				module: app.conf.BuildAppletName,
				icon:   app.FileLocation("../", app.conf.BuildAppletName, "/icon"),
				build:  func() { app.buildApplet(app.conf.BuildAppletName) },
			}
		case AppletInternal:
			app.target = &BuildCompiled{module: app.conf.BuildAppletName}
		//~ case EmblemSmall, EmblemLarge:
		//~ app.render = NewRenderedSVG(app.CDApplet, app.config.renderer)
		//~ default: // NoDisplay case, but using default to be sure we have a valid renderer.
		//~ app.render = NewRenderedNone()
		default:
			app.target = &BuildNull{}
		}
	} else {
		app.target = &BuildCore{}
	}
}

// Try to detect applet type based on its location and content.
// Could be improved.
//
func (app *AppletUpdate) appletType(name string) BuildType {
	if _, e := os.Stat(app.FileLocation("..", name)); e == nil { // Applet is in external dir.
		if _, e := os.Stat(app.FileLocation("..", name, "Makefile")); e == nil { // Got makefile => can build.
			term.Info("ok")
			return AppletCompiled
		}
		// Else it's a scripted (atm the only other module not scripted is in vala and is the previous version of this one).
		return AppletScript
	}

	return AppletInternal // // Finally it must be an internal one. 
}

func (app *AppletUpdate) showTarget() {
	//~ term.Info(app.FileLocation("../", app.conf.BuildAppletName, "/icon"))
	//~ if len(app.conf.BuildAppletName) > 4 {
	//~ name = app.conf.BuildAppletName[:4]
	//~ } else {
	//~ name = app.conf.BuildAppletName
	//~ }
	//~ app.SetQuickInfo(name)

	app.SetEmblem(app.target.Icon(), dock.EmblemTopLeft)
	app.SetLabel("Target: " + app.target.Label())
}

func (app *AppletUpdate) buildApplet(name string) {
	var e error
	term.Info(app.conf.BuildAppletName, app.appletType(app.conf.BuildAppletName))
	switch app.appletType(app.conf.BuildAppletName) {
	case AppletCompiled:
		if logE("buildApplet Chdir: ", os.Chdir(app.FileLocation("..", name))) {
			return
		}

		e = execShow("make")
	case AppletInternal:
		//~ TODO
	default:
		e = errors.New("not finished")
	}
	// Finally we restart the app.
	if e == nil { // No errors => restart.
		if app.conf.BuildAppletName == app.AppletName { // Don't eat the chicken, ore you won't have any more eggs.
			exec.Command("./reload.sh", app.AppletName).Start()
		} else {
			app.ActivateModule(app.conf.BuildAppletName, false)
			app.ActivateModule(app.conf.BuildAppletName, true)
		}
	}
}

//------------------------------------------------------------------------------
// SHOW VERSIONS.
//------------------------------------------------------------------------------

// To improve : parse http://bazaar.launchpad.net/~cairo-dock-team/cairo-dock-core/cairo-dock/changes/
// and maybe see to use as download tool : http://golang.org/src/cmd/go/vcs.go
//
func (app *AppletUpdate) actionShowVersions() {
	app.version.Check()
}

// Got versions informations, Need to set a new emblem
func (app *AppletUpdate) onGotVersions(new int, e error) {
	if new > 0 {
		app.ShowDialog(app.version.FormatDialog(), int32(app.conf.VersionDialogTimer))
	}
	term.Info("img", app.FileLocation("img", app.conf.VersionEmblemNew))
	app.SetEmblem(app.FileLocation("img", app.conf.VersionEmblemNew), dock.EmblemBottomLeft)
}

//------------------------------------------------------------------------------
// VERSION POLLING.
//------------------------------------------------------------------------------

type Versions struct {
	sources  []*Branch
	template *template.Template

	// Polling data.
	restart    chan bool        // restart channel to forward user requests.
	callResult func(int, error) // Action to execute to send polling results.
}

func (ver *Versions) FormatDialog() string {
	buff := bytes.NewBuffer([]byte(""))
	logE("FormatDialog", ver.template.ExecuteTemplate(buff, "ShowVersionsDialog", ver.sources))
	return buff.String()
}

// Check for new mails. Return the mails count delta (change since last check).
// callback for poller.
// TODO: need to better handle errors.
//
func (ver *Versions) Check() {
	var n, cur int
	var e error
	for _, branch := range ver.sources {
		cur, e = branch.findNew()
		n += cur
	}
	ver.callResult(n, e)
}

/*

// Get ver data.
//
func (ver *Versions) SetRestart(restart chan bool) {
	ver.restart = restart
}

// Get number of unread mails.
//
//~ func (ver *Versions) Count() int {
	//~ return len(ver.Mail)
//~ }


func (ver *Versions) IsValid() bool {
	//~ return ver.login != ""
}


// Get feed data.
//
func (ver *Versions) Data() interface{} {
	return ver
}


// Restart mail polling ticker.
//
func (ver *Versions) Restart() {
	if ver.IsValid() {
		//~ ver.Mail = nil // Our renderer has been reset, counter need to be reset too to redisplay correct value.
		log.Println("should restart")
		ver.restart <- true // send our restart event.
	} else {
		ver.callResult(0, errors.New("No account informations provided."))
	}
}
*/

//------------------------------------------------------------------------------
// SOURCES BRANCH MANAGEMENT.
//------------------------------------------------------------------------------

type Branch struct {
	Branch  string // Branch name.
	Dir     string // Location of branch on the filesystem.
	GotData bool   // true if data was successfully pulled.

	Log       string // Commit messages for new commits.
	RevLocal  int    // Current local version.
	RevDist   int    // Current server version.
	Delta     int    // Delta of revisions between server and local (server - local).
	NewLocal  int    //  Number of unmerged patch in local dir.
	NewServer int    // Number of new commits on server.
	Zero      bool   // True if revisions are the same.
}

func NewBranch(branch, dir string) *Branch {
	return &Branch{
		Branch: branch,
		Dir:    dir,
	}
}

// Get revisions informations.
//
func (branch *Branch) findNew() (new int, e error) {
	branch.GotData = false
	previous := branch.RevDist
	if logE("RevNo Chdir", os.Chdir(branch.Dir)) {
		return
	}

	local := execSync(command, "revno")
	dist := execSync(command, "revno", branch.Branch)

	if local == "" || dist == "" {
		return
	}

	current, ok1 := TrimInt(local)
	server, ok2 := TrimInt(dist)

	if ok1 != nil || ok2 != nil {
		e = errors.New("Check versions failed.")
		logE("findNew", e)
		return
	}

	// We have valid data.
	branch.GotData = true
	branch.RevLocal = current
	branch.RevDist = server
	branch.Delta = server - current

	if delta := branch.RevDist - previous; delta > 0 {
		new = delta
	}

	// Get log info for new commits.
	log := ""
	if branch.Delta != 0 {
		log = execSync(command, "log", "--line", "-l"+strconv.Itoa(branch.Delta))
	}
	branch.Log = log

	switch { // Data for formatter.
	case branch.Delta > 0:
		branch.NewServer = branch.Delta
	case branch.Delta == 0:
		branch.Zero = true
	case branch.Delta < 0:
		branch.NewLocal = -branch.Delta
	}
	return
}

// Load template file. If error, it will just be be logged, so you must check 
// that the template is valid.
//
func LoadTemplate(filename string) *template.Template {
	template, e := template.ParseFiles(filename)
	logE("Template loading", e)
	return template
}
