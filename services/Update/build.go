package Update

import (
	"github.com/sqp/godock/libs/dbus"
	"github.com/sqp/godock/libs/packages/build"
	"github.com/sqp/godock/libs/ternary"

	"os"
	"path"
	"strconv"
	"strings"
	"text/template"
	"time"
)

//
//----------------------------------------------------------[ BUILDERS CONFIG ]--

// BuildTarget defines the common builder interface.
//
type BuildTarget interface {
	Label() string
	Icon() string
	Build() error
	SourceDir() string
	SetProgress(func(float64)) // Need values between 0 and 1 in the renderer.
	Progress(float64)
	SetIcon(icon string)
	SetDir(dir string)
}

//
//----------------------------------------------------------[ SOURCES BUILDER ]--

// Create the target renderer/builder.
//
func (app *AppletUpdate) setBuildTarget() {
	if !app.conf.UserMode { // Tester mode.
		app.target = &build.BuilderNull{}
	} else {
		list := app.conf.BuildTargets
		target := list[app.targetID]
		switch build.GetBuildType(target) {

		case build.Core:
			app.target = &build.BuilderCore{}
			app.target.SetDir(app.conf.SourceDir)

		case build.Applets:
			app.target = &build.BuilderApplets{}
			app.target.SetDir(app.conf.SourceDir)

		case build.AppletCompiled:
			pack := dbus.InfoApplet(target)
			if pack != nil {
				app.target = &build.BuilderCompiled{Module: target}
				app.target.SetIcon(pack.Icon)
				app.target.SetDir(pack.Dir())

			} else {
				app.Log.NewErr("applet not found: "+target, "set build target")
				app.target = &build.BuilderNull{}

				// app.target = &BuildCompiled{
				// 	module: target,
				// 	icon:   app.FileLocation("../", target, "/icon"),
				// 	dir:    app.FileLocation("../", target),
				// }
			}

		case build.AppletInternal:
			// Ask icon of module to the Dock as we can't guess its dir and icon name.
			icon := app.FileLocation("img", app.conf.IconMissing)
			if mod := dbus.InfoApplet(strings.Replace(target, "-", " ", -1)); mod != nil {
				icon = mod.Icon
			}

			app.target = &build.BuilderInternal{Module: target}
			app.target.SetIcon(icon)
			app.target.SetDir(path.Join(app.conf.SourceDir, "cairo-dock-plug-ins"))

		default: // ensure we have a valid target.
			app.target = &build.BuilderNull{}
		}
		app.target.SetProgress(func(f float64) { app.RenderValues(f) })
	}

	// app.Log.Info(app.conf.BuildTargets[app.targetID], app.buildType(app.conf.BuildTargets[app.targetID]))

	// Delayed display of emblem. 5ms seemed to be enough but 500 should do the job.
	go func() { time.Sleep(500 * time.Millisecond); app.showTarget() }()
}

func (app *AppletUpdate) showTarget() {
	app.SetEmblem(app.target.Icon(), EmblemTarget)
	app.SetLabel("Target: " + app.target.Label())
}

// Restart target if needed, with everything necessary (like the dock for an
// internal app).
//
func (app *AppletUpdate) restartTarget() {
	if !app.conf.BuildReload {
		return
	}
	target := app.conf.BuildTargets[app.targetID]
	app.Log.Info("restart", target)
	switch build.GetBuildType(target) {
	case build.AppletScript, build.AppletCompiled:
		if target == app.AppletName { // Don't eat the chicken, or you won't have any more eggs.
			logger.ExecAsync("make", "reload")
		} else {
			dbus.AppletRemove(target + ".conf")
			dbus.AppletAdd(target)
			// app.ActivateModule(target, false)
			// app.ActivateModule(target, true)
		}

	default:
		func() {
			logger.ExecAsync("cdc", "restart")
		}()

	}
}

//
//---------------------------------------------------------[ VERSION POLLING ]--

// Versions handles the version checking and result display method.
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

// Sources lists configured branches.
//
func (ver *Versions) Sources() []*Branch {
	return ver.sources
}

// Check is the callback for poller.
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

//
//-----------------------------------------------[ SOURCES BRANCH MANAGEMENT ]--

// Branch defines a sources branch information.
//
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

// NewBranch creates a source branch with name and dir.
//
func NewBranch(branch, dir string) *Branch {
	return &Branch{
		Branch: branch,
		Dir:    dir,
	}
}

// Get revisions informations.
//
func (branch *Branch) findNew() (new int, e error) {
	logger.Debug("Get version", branch.Branch)
	branch.GotData = false
	if logger.Err(os.Chdir(branch.Dir), "findNew Chdir") {
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

	logger.Debug("", "local=", current, "server=", server, "delta=", branch.Delta, "new=", branch.Delta-lastKnown)

	// logger.Info("versions", lastKnown, branch.Delta)

	// Get log info for new commits.
	logInfo := ""
	if branch.Delta-lastKnown > 0 {
		nb := ternary.Min(branch.Delta, 5)
		logInfo, e = logger.ExecSync(CmdBzr, "log", LocationLaunchpad+branch.Branch, "--line", "-l"+strconv.Itoa(nb))
		logger.Warn(e, "bzr log "+branch.Branch)
		// logger.Info("Cairo-Dock Commit", logInfo)
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

func getRevision(args ...string) int {
	args = append([]string{"revno"}, args...)
	rev, e := logger.ExecSync(CmdBzr, args...)

	if len(args) == 1 { // no args, adding local for error display.
		args = append(args, "local")
	}
	if !logger.Err(e, "Check revision: "+CmdBzr+" "+strings.Join(args, " ")) {
		version, e := trimInt(string(rev))
		if !logger.Err(e, "Check revision: type mismatch: "+string(rev)) {
			return version
		}
	}
	return 0
}

func (branch *Branch) update(dir string, progress func(float64)) (new int, e error) {
	if e = os.Chdir(dir); e != nil {
		return 0, e
	}
	ret, e := logger.ExecSync(CmdBzr, "up") // "pull", LocationLaunchpad+branch.Branch)
	logger.Info("PULL", ret)
	return 0, e
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
	logger.Println("Saved", path.Join(app.conf.ScriptLocation, app.conf.ScriptName))
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
