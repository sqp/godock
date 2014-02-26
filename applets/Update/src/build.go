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
	"bufio"
	// "errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/sqp/godock/libs/dbus"
	"github.com/sqp/godock/libs/log"

	// Applet still under work.
	// "github.com/kr/pretty"
)

// func grr(i interface{}) { pretty.Printf("%# v\n", i) } // libs unused are errors in go. Force using lib to prevent build fail when lib is activated.

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
	Build() error
	SourceDir() string
	SetProgress(func(float64)) // Need values between 0 and 1 in the renderer.
	Progress(float64)
}

// Progress callback handler for Builders.
//
type buildProgress struct {
	f func(float64)
}

func (build *buildProgress) Progress(data float64) {
	if build.f != nil {
		build.f(data)
	}
}
func (build *buildProgress) SetProgress(f func(float64)) {
	build.f = f
}

// Empty Build for fallback.
//
type BuildNull struct {
	buildProgress
}

func (build *BuildNull) Icon() string                { return "none" }
func (build *BuildNull) Label() string               { return "" }
func (build *BuildNull) SourceDir() string           { return "" }
func (build *BuildNull) Build() error                { return inProgress }
func (build *BuildNull) SetProgress(f func(float64)) {}

//
//
type BuildCore struct {
	buildProgress
	dir string
	app *AppletUpdate
}

func (build *BuildCore) Icon() string {
	return "/usr/share/cairo-dock/cairo-dock.svg"
}

func (build *BuildCore) Label() string {
	return "Core"
}

func (build BuildCore) SourceDir() string {
	return path.Join(build.dir, "cairo-dock-core")
}

func (build *BuildCore) Build() error {
	return buildCmake(build.SourceDir(), build.buildProgress.f)
}

// Build all Cairo-Dock plug-ins.
//
type BuildApplets struct {
	buildProgress
	dir string
}

func (build *BuildApplets) Icon() string {
	return "/usr/share/cairo-dock/icons/icon-extensions.svg"
}

func (build *BuildApplets) Label() string {
	return "Applets"
}

func (build BuildApplets) SourceDir() string {
	return path.Join(build.dir, "cairo-dock-plug-ins")
}

func (build *BuildApplets) Build() error {
	return buildCmake(build.SourceDir(), build.buildProgress.f)
}

// Internal applet = C applet provided by the plug-ins package.
//
type BuildInternal struct {
	buildProgress
	module string
	icon   string
	dir    string
}

func (build *BuildInternal) Icon() string {
	return build.icon
}

func (build *BuildInternal) Label() string {
	return build.module
}

func (build *BuildInternal) SourceDir() string {
	return build.dir
}

func (build *BuildInternal) Build() error {
	dir := path.Join(build.dir, "build", build.module)
	if e := os.Chdir(dir); e != nil {
		return e
	}
	return actionMakeAndInstall(build.buildProgress.f)
}

// External applet that must be compiled (golang or vala)
//
type BuildCompiled struct {
	buildProgress
	module string
	icon   string
	dir    string
}

func (build *BuildCompiled) Icon() string {
	return build.icon
}

func (build *BuildCompiled) Label() string {
	return build.module
}

func (build *BuildCompiled) SourceDir() string {
	return build.dir
}

func (build *BuildCompiled) Build() error {
	if e := os.Chdir(build.dir); e != nil {
		return e
	}

	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()
	go activityBar(t.C, build.buildProgress.f)

	return execShow("make")
}

//------------------------------------------------------------------------------
// SOURCES BUILDER.
//------------------------------------------------------------------------------

func (app *AppletUpdate) getBuildTargets() []string {
	s := app.conf.BuildTargets
	if s[len(s)-1] == ';' { // Drop last ; if any.
		s = s[:len(s)-1]
	}
	return strings.Split(s, ";") // And return splitted list.
}

// Create the target renderer/builder.
//
func (app *AppletUpdate) setBuildTarget() {
	if !app.conf.UserMode { // Tester mode.
		app.target = &BuildNull{}
	} else {
		list := app.getBuildTargets()
		target := list[app.targetId]
		switch app.buildType(target) {

		case Core:
			app.target = &BuildCore{
				dir: app.conf.SourceDir,
			}

		case Applets:
			app.target = &BuildApplets{
				dir: app.conf.SourceDir,
			}

		case AppletCompiled:
			app.target = &BuildCompiled{
				module: target,
				icon:   app.FileLocation("../", target, "/icon"),
				dir:    app.FileLocation("../", target),
			}

		case AppletInternal:
			// Ask icon of module to the Dock as we can't guess its dir and icon name.
			icon := app.FileLocation("img", app.conf.IconMissing)
			if mod := dbus.InfoApplet(strings.Replace(target, "-", " ", -1)); mod != nil {
				icon = mod.Icon
			}
			app.target = &BuildInternal{
				module: target,
				icon:   icon,
				dir:    path.Join(app.conf.SourceDir, "cairo-dock-plug-ins"),
			}

		default: // ensure we have a valid target.
			app.target = &BuildNull{}
		}
		app.target.SetProgress(func(f float64) { app.RenderValues(f) })
	}

	// Delayed display of emblem. 5ms seemed to be enough but 500 should do the job.
	go func() { time.Sleep(500 * time.Millisecond); app.showTarget() }()
}

// Try to detect applet type based on its location and content.
// Could be improved.
//
func (app *AppletUpdate) buildType(name string) BuildType {
	switch name {
	case "core":
		return Core
	case "applets", "plug-ins":
		return Applets
	}

	if _, e := os.Stat(app.FileLocation("..", name)); e == nil { // Applet is in external dir.
		if _, e := os.Stat(app.FileLocation("..", name, "Makefile")); e == nil { // Got makefile => can build.
			return AppletCompiled
		}
		// Else it's a scripted (atm the only other module not scripted is in vala and is the previous version of this module).
		return AppletScript
	}

	return AppletInternal // Finally it must be an internal one.
}

func (app *AppletUpdate) showTarget() {
	app.SetEmblem(app.target.Icon(), EmblemTarget)
	app.SetLabel("Target: " + app.target.Label())
}

//------------------------------------------------------------------------------
// BUILD FROM C SOURCES.
//------------------------------------------------------------------------------

func buildCmake(dir string, call func(float64)) error {

	if _, e := os.Stat(dir); e != nil { // No basedir.
		return e
	}

	// Initialise build subdir. Create or clean previous compile.
	dir += "/build"
	if _, e := os.Stat(dir); e == nil { // Dir exists. Clean if needed.
		if e = os.Chdir(dir); e != nil {
			return e
		}
		actionMakeClean()
	} else { // Create build dir and launch cmake.
		log.Info("Create build directory")
		if e = os.Mkdir(dir, os.ModePerm); e != nil {
			return e
		}
		os.Chdir(dir)
		execShow("cmake", "..", "-DCMAKE_INSTALL_PREFIX=/usr", "-Denable-disks=yes", "-Denable-impulse=yes", "-Denable-gmenu=no")
	}

	return actionMakeAndInstall(call)

	//~ PARAM_PLUG_INS:="-Denable-scooby-do=yes -Denable-disks=yes -Denable-mail=no -Denable-impulse=yes"
}

// Make clean in build dir.
//
func actionMakeClean() {
	// if !*buildKeep { // Default is to clean build dir.
	log.Info("Clean build directory")
	execShow("make", "clean")
	// }
}

// Compile and install a Cairo-Dock source dir.
//
func actionMakeAndInstall(call func(float64)) error {
	jobs := runtime.NumCPU()

	// execShow("make", "-j", strconv.Itoa(jobs))

	cmd := exec.Command("make", "-j", strconv.Itoa(jobs))
	stdout, _ := cmd.StdoutPipe()
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Err(err, "Execute make")
	}
	r := bufio.NewReader(stdout)

	line, err := r.ReadString('\n')
	for err == nil {
		if len(line) > 3 && line[0] == '[' {
			progress, e := trimInt(line[1:4])
			if e == nil {
				call(float64(progress) / 100)
				fmt.Printf("[%3d%%] %s", progress, log.Green(line[7:]))
			}
		} else {
			fmt.Fprint(os.Stdout, line)
		}

		line, err = r.ReadString('\n')
	}

	return execShow(cmdSudo, "make", "install")
}

// Restart target if needed, with everything necessary (like the dock for an
// internal app).
//
func (app *AppletUpdate) restartTarget() {
	if !app.conf.BuildReload {
		return
	}
	list := app.getBuildTargets()
	target := list[app.targetId]
	log.Info(target)
	switch app.buildType(target) {
	case AppletScript, AppletCompiled:
		if target == app.AppletName { // Don't eat the chicken, or you won't have any more eggs.
			exec.Command("make", "reload").Start()
		} else {
			dbus.AppletRemove(target + ".conf")
			dbus.AppletAdd(target)
			// app.ActivateModule(target, false)
			// app.ActivateModule(target, true)
			// app.DockRemove("type=Module-Instance & config-file=" + target + ".conf")
			// go app.DockAdd(map[string]interface{}{"type": "Module", "module": target})
		}

	default:
		func() {
			// exec.Command("killall", "cairo-dock").Start()
			// exec.Command("cairo", "reload").Start()
			execAsync("nohup", "cdc", "reload")
		}()

	}
}

//------------------------------------------------------------------------------
// VERSION POLLING.
//------------------------------------------------------------------------------

// Just handles the version checking and result display method.
//
type Versions struct {
	sources        []*Branch
	template       *template.Template
	dialogTemplate string
	fileTemplate   string
	newCommits     int

	// Polling data.
	restart    chan bool        // restart channel to forward user requests.
	callResult func(int, error) // Action to execute to send polling results.
}

func (ver *Versions) Sources() []*Branch {
	return ver.sources
}

// callback for poller.
// TODO: need to better handle errors.
//
func (ver *Versions) Check() {
	var nb, cur int
	var e error
	for _, branch := range ver.sources {
		cur, e = branch.findNew()
		nb += cur
	}
	ver.callResult(nb, e)
}

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
	log.Debug("Get version", branch.Branch)
	branch.GotData = false
	if logE("RevNo Chdir", os.Chdir(branch.Dir)) {
		return
	}

	current := getRevision()
	server := getRevision(LocationLaunchpad + branch.Branch)

	if current == 0 || server == 0 {
		return
	}

	// We have valid data.

	// Save versions.
	branch.GotData = true
	branch.RevLocal = current
	branch.RevDist = server

	lastKnown := branch.Delta
	branch.Delta = server - current

	// log.Info("versions", lastKnown, branch.Delta)

	// Get log info for new commits.
	logInfo := ""
	if branch.Delta-lastKnown > 0 {
		nb := testInt(branch.Delta > 5, 5, branch.Delta)
		logInfo, e = execSync(cmdBzr, "log", LocationLaunchpad+branch.Branch, "--line", "-l"+strconv.Itoa(nb))
		log.Warn(e, "bzr log "+branch.Branch)
		// log.Info("Cairo-Dock Commit", logInfo)
	}
	branch.Log = logInfo

	switch { // Data for formatter.
	case branch.Delta > 0:
		branch.NewServer = branch.Delta
	case branch.Delta == 0:
		branch.Zero = true
	case branch.Delta < 0:
		branch.NewLocal = -branch.Delta
	}

	return branch.Delta, e
}

func (branch *Branch) update() (new int, e error) {

	ret, e := execSync(cmdBzr, "pull", LocationLaunchpad+branch.Branch)
	log.Info("PULL", ret)
	return 0, e
}

// Load template file. If error, it will just be be logged, so you must check
// that the template is valid.
//
// func loadTemplate(filename string) *template.Template {
// 	template, e := template.ParseFiles(filename)
// 	logE("Template loading", e)
// 	return template
// }

func getRevision(args ...string) int {
	args = append([]string{"revno"}, args...)
	rev, e := exec.Command(cmdBzr, args...).Output()

	if len(args) == 1 { // no args, adding local for error display.
		args = append(args, "local")
	}
	if !log.Err(e, "Check revision: "+cmdBzr+" "+strings.Join(args, " ")) {
		version, e := trimInt(string(rev))
		if !log.Err(e, "Check revision: type mismatch: "+string(rev)) {
			return version
		}
	}
	return 0
}

func GetLog() {
}

/*
// Get and launch bzr update script. Unused.
//
func (app *AppletUpdate) updateByBzrScript() {
if logE("Wrong sources directory", os.Chdir(app.conf.SourceDir)) { // move to base dir.
	return
}
if _, e := os.Stat(app.conf.ScriptName); e != nil { // script missing ?.
	_, e := exec.Command("wget", app.conf.ScriptLocation).Output()
	if logE("Download script failed", e) {

	//~ if execSync("wget", app.conf.ScriptLocation) == "" { // Download.
		//~ logE("Download script failed", errors.New("Check the wget log"))
		return
	}
	log.Println("Saved", path.Join(app.conf.ScriptLocation, app.conf.ScriptName))
		if logE("Can't chmod", execShow("chmod", "a+x", app.conf.ScriptName)) { // allow execute.
			return
		}

}

if logE("Update failed", execShow("./" + app.conf.ScriptName, "-u")) {
	return
}

}
*/

/*


pdate(){
	if [ $LG -eq 0 ]; then
		LG_SEARCH_FOR="Recherche des mises à jour pour"
		LG_UP_FOUND="Une mise à jour a été détectée pour"
	else
		LG_SEARCH_FOR="Check if there is an update for"
		LG_UP_FOUND="An update has been detected for"
	fi

	echo -e "$BLEU""$LG_SEARCH_FOR Cairo-Dock"
	if test -e "$BZR_REV_FILE_CORE"; then
		ACTUAL_CORE_VERSION=`cat "$BZR_REV_FILE_CORE"`
	else
		echo 0 > "$BZR_REV_FILE_CORE"
		ACTUAL_CORE_VERSION=0
	fi

	if [ $BZR_DL_MODE -eq 1 ]; then
		cd $DIR/$CAIRO_DOCK_CORE_LP_BRANCH
		BZR_UP="pull"
		bzr $BZR_UP lp:$CAIRO_DOCK_CORE_LP_BRANCH
		NEW_CORE_VERSION=`bzr revno -q`
		cd $DIR/
	else
		BZR_UP="update -q"
		NEW_CORE_VERSION=`bzr revno -q $CAIRO_DOCK_CORE_LP_BRANCH`
		if [ $ACTUAL_CORE_VERSION -ne $NEW_CORE_VERSION ]; then
			bzr $BZR_UP $CAIRO_DOCK_CORE_LP_BRANCH
		fi
	fi

	echo $NEW_CORE_VERSION > $BZR_REV_FILE_CORE
	echo -e "\nCairo-Dock-Core : rev $ACTUAL_CORE_VERSION -> $NEW_CORE_VERSION \n"
	echo -e "\nCairo-Dock-Core : rev $ACTUAL_CORE_VERSION -> $NEW_CORE_VERSION \n" >> $LOG_CAIRO_DOCK

	if [ $ACTUAL_CORE_VERSION -ne $NEW_CORE_VERSION ]; then
		DIFF_CORE_VERSION=$(($NEW_CORE_VERSION-$ACTUAL_CORE_VERSION))
		if [ $DIFF_CORE_VERSION -le 10 ]; then
			bzr log -l$DIFF_CORE_VERSION --line $CAIRO_DOCK_CORE_LP_BRANCH
		else
			bzr log -l1 --line $CAIRO_DOCK_CORE_LP_BRANCH
		fi
		echo -e "$VERT""\n$LG_UP_FOUND Cairo-Dock"
		sleep 1
		install_cairo_dock
		UPDATE_CAIRO_DOCK=1
		UPDATE=1
		#echo -e "$VERT""Mise à jour et recompilation des plug-ins suite à la mise à jour de cairo-dock"
		echo -e "$NORMAL"""
	else
		echo -e "$NORMAL"""
	fi

	## PLUG-INS ##

	echo -e "$BLEU""\n$LG_SEARCH_FOR Plug-ins"
	if test -e "$BZR_REV_FILE_PLUG_INS"; then
		ACTUAL_PLUG_INS_VERSION=`cat "$BZR_REV_FILE_PLUG_INS"`
	else
		echo 0 > "$BZR_REV_FILE_PLUG_INS"
		ACTUAL_PLUG_INS_VERSION=0
	fi

	if [ $BZR_DL_MODE -eq 1 ]; then
		cd $DIR/$CAIRO_DOCK_PLUG_INS_LP_BRANCH
		bzr $BZR_UP lp:$CAIRO_DOCK_PLUG_INS_LP_BRANCH
		NEW_PLUG_INS_VERSION=`bzr revno -q`
		cd $DIR/
	else
		NEW_PLUG_INS_VERSION=`bzr revno -q $CAIRO_DOCK_PLUG_INS_LP_BRANCH`
		if [ $ACTUAL_PLUG_INS_VERSION -ne $NEW_PLUG_INS_VERSION ]; then
			bzr $BZR_UP $CAIRO_DOCK_PLUG_INS_LP_BRANCH
		fi
	fi

	echo $NEW_PLUG_INS_VERSION > "$BZR_REV_FILE_PLUG_INS"
	echo -e "\nCairo-Dock-Plug-Ins : rev $ACTUAL_PLUG_INS_VERSION -> $NEW_PLUG_INS_VERSION \n"
	echo -e "\nCairo-Dock-Plug-Ins : rev $ACTUAL_PLUG_INS_VERSION -> $NEW_PLUG_INS_VERSION \n" >> $LOG_CAIRO_DOCK

	if [ $ACTUAL_PLUG_INS_VERSION -ne $NEW_PLUG_INS_VERSION ]; then
		DIFF_PLUG_INS_VERSION=$(($NEW_PLUG_INS_VERSION-$ACTUAL_PLUG_INS_VERSION))
		if [ $DIFF_PLUG_INS_VERSION -le 10 ]; then
			bzr log -l$DIFF_PLUG_INS_VERSION --line $CAIRO_DOCK_PLUG_INS_LP_BRANCH
		else
			bzr log -l1 --line $CAIRO_DOCK_PLUG_INS_LP_BRANCH
		fi
		echo -e "$VERT""\n$LG_UP_FOUND Plug-Ins"
		install_cairo_dock_plugins
		UPDATE=1
	elif [ $UPDATE_CAIRO_DOCK -eq 1 ]; then
		if [ $LG -eq 0 ]; then
			echo -e "$VERT""Recompilation des plug-ins suite à la mise à jour de Cairo-Dock Core"
		else
			echo -e "$VERT""Recompilation due to some changes of Cairo-Dock API"
		fi
		install_cairo_dock_plugins
	fi
	echo -e "$NORMAL"

	## PLUG-INS EXTRAS ##

	echo -e "$BLEU""\n$LG_SEARCH_FOR Plug-ins Extras"
	if test -e "$BZR_REV_FILE_PLUG_INS_EXTRAS"; then # le fichier existe
		ACTUAL_PLUG_INS_EXTRAS_VERSION=`cat "$BZR_REV_FILE_PLUG_INS_EXTRAS"`
	else
		echo 0 > "$BZR_REV_FILE_PLUG_INS_EXTRAS"
		ACTUAL_PLUG_INS_EXTRAS_VERSION=0
	fi

	if [ $BZR_DL_MODE -eq 1 ]; then
		cd $DIR/$CAIRO_DOCK_PLUG_INS_EXTRAS_LP_BRANCH
		bzr $BZR_UP lp:$CAIRO_DOCK_PLUG_INS_EXTRAS_LP_BRANCH
		NEW_PLUG_INS_EXTRAS_VERSION=`bzr revno -q`
		cd $DIR/
	else
		NEW_PLUG_INS_EXTRAS_VERSION=`bzr revno -q $CAIRO_DOCK_PLUG_INS_EXTRAS_LP_BRANCH`
		if [ $ACTUAL_PLUG_INS_EXTRAS_VERSION -ne $NEW_PLUG_INS_EXTRAS_VERSION ]; then
			bzr $BZR_UP $CAIRO_DOCK_PLUG_INS_EXTRAS_LP_BRANCH
		fi
	fi

	echo $NEW_PLUG_INS_EXTRAS_VERSION > "$BZR_REV_FILE_PLUG_INS_EXTRAS"
	echo -e "\nCairo-Dock-Plug-Ins-Extras : rev $ACTUAL_PLUG_INS_EXTRAS_VERSION -> $NEW_PLUG_INS_EXTRAS_VERSION \n"
	echo -e "\nCairo-Dock-Plug-Ins-Extras : rev $ACTUAL_PLUG_INS_EXTRAS_VERSION -> $NEW_PLUG_INS_EXTRAS_VERSION \n" >> $LOG_CAIRO_DOCK

	if [ $ACTUAL_PLUG_INS_EXTRAS_VERSION -ne $NEW_PLUG_INS_EXTRAS_VERSION ]; then
		DIFF_PLUG_INS_EXTRAS_VERSION=$(($NEW_PLUG_INS_EXTRAS_VERSION-$ACTUAL_PLUG_INS_EXTRAS_VERSION))
		if [ $DIFF_PLUG_INS_EXTRAS_VERSION -le 10 ]; then
			bzr log -l$DIFF_PLUG_INS_EXTRAS_VERSION --line $CAIRO_DOCK_PLUG_INS_EXTRAS_LP_BRANCH
		else
			bzr log -l1 --line $CAIRO_DOCK_PLUG_INS_EXTRAS_LP_BRANCH
		fi
		echo -e "$VERT""\n$LG_UP_FOUND Plug-Ins Extras"
		install_cairo_dock_plugins_extras
		UPDATE=1
	fi

	## CAIRO-DESKLET ##

	echo -e "$BLEU""\n$LG_SEARCH_FOR Desklets"
	if test -e "$BZR_REV_FILE_DESKLET"; then # le fichier existe
		ACTUAL_DESKLET_VERSION=`cat "$BZR_REV_FILE_DESKLET"`
	else
		echo 0 > "$BZR_REV_FILE_DESKLET"
		ACTUAL_DESKLET_VERSION=0
	fi

	if [ ! -d $DIR/$CAIRO_DESKLET_LP_BRANCH ]; then # desklet a été ajouté après
		echo -e "$BLEU""$LG_DL_BG Desklets"
		if [ $BZR_DL_MODE -eq 1 ]; then
			bzr branch lp:$CAIRO_DESKLET_LP_BRANCH
		else
			bzr checkout --lightweight lp:$CAIRO_DESKLET_LP_BRANCH
		fi
		if [ $? -ne 0 ]; then
			echo -e "$ROUGE""$LG_DL_ERROR"
			read
			exit
		else
			NEW_DESKLET_VERSION=`bzr revno -q $CAIRO_DESKLET_LP_BRANCH`
			echo $NEW_DESKLET_VERSION > $BZR_REV_FILE_DESKLET
			echo -e "\nCairo-Desklets : rev $NEW_DESKLET_VERSION \n"
			echo -e "\nCairo-Desklets : rev $NEW_DESKLET_VERSION \n" >> $LOG_CAIRO_DOCK
		fi
	elif [ $BZR_DL_MODE -eq 1 ]; then
		cd $DIR/$CAIRO_DESKLET_LP_BRANCH
		bzr $BZR_UP lp:$CAIRO_DESKLET_LP_BRANCH
		NEW_DESKLET_VERSION=`bzr revno -q`
		cd $DIR/
	else
		NEW_DESKLET_VERSION=`bzr revno -q $CAIRO_DESKLET_LP_BRANCH`
		if [ $ACTUAL_DESKLET_VERSION -ne $NEW_DESKLET_VERSION ]; then
			bzr $BZR_UP $CAIRO_DESKLET_LP_BRANCH
		fi
	fi

	echo $NEW_DESKLET_VERSION > "$BZR_REV_FILE_DESKLET"
	echo -e "\nCairo-Desklet : rev $ACTUAL_DESKLET_VERSION -> $NEW_DESKLET_VERSION \n"
	echo -e "\nCairo-Desklet : rev $ACTUAL_DESKLET_VERSION -> $NEW_DESKLET_VERSION \n" >> $LOG_CAIRO_DOCK

	if [ $ACTUAL_DESKLET_VERSION -ne $NEW_DESKLET_VERSION ]; then
		DIFF_DESKLET_VERSION=$(($NEW_DESKLET_VERSION-$ACTUAL_DESKLET_VERSION))
		if [ $DIFF_DESKLET_VERSION -le 10 ]; then
			bzr log -l$DIFF_DESKLET_VERSION --line $CAIRO_DESKLET_LP_BRANCH
		else
			bzr log -l1 --line $CAIRO_DESKLET_LP_BRANCH
		fi
		echo -e "$VERT""\n$LG_UP_FOUND Desklet"
		install_cairo_desklet
		UPDATE=1
	elif [ $UPDATE_CAIRO_DOCK -eq 1 ]; then
		if [ $LG -eq 0 ]; then
			echo -e "$VERT""Recompilation suite à la mise à jour de l'API Cairo-Dock"
		else
			echo -e "$VERT""Recompilation due to some changes of Cairo-Dock API"
		fi
		install_cairo_desklet
	fi

	## CHECK ##

	echo -e "$NORMAL"

	if [ $UPDATE -eq 1 ]; then
	    check $LOG_CAIRO_DOCK "CD"
	else
		if [ $LG -eq 0 ]; then
			echo -e "$BLEU""Pas de mise à jour disponible"
			echo -e "$NORMAL"
			if test  `ps aux | grep -c " [c]airo-dock"` -gt 0; then
				dbus-send --session --dest=org.cairodock.CairoDock /org/cairodock/CairoDock org.cairodock.CairoDock.ShowDialog string:"Cairo-Dock: Pas de mise à jour" int32:8 string:"class=$COLORTERM"
			else
				zenity --info --title=Cairo-Dock --text="$LG_CLOSE"
			fi
		else
			echo -e "$BLEU""No update available"
			echo -e "$NORMAL"
			if test  `ps aux | grep -c " [c]airo-dock"` -gt 0; then
				dbus-send --session --dest=org.cairodock.CairoDock /org/cairodock/CairoDock org.cairodock.CairoDock.ShowDialog string:"Cairo-Dock: no update" int32:8 string:"class=$COLORTERM"
			else
				zenity --info --title=Cairo-Dock --text="$LG_CLOSE"
			fi
		fi
	fi
*/
