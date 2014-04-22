package build

import (
	"github.com/sqp/godock/libs/appdbus"
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/log"
	// "github.com/sqp/godock/libs/ternary"

	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
	// "text/template"
	"time"
)

var Log cdtype.Logger

//
//---------------------------------------------------------[ BUILDERS CONFIG ]--

var (
	CmdSudo string = "gksudo" // Default command to get root access for installation.

	inProgress error = errors.New("not finished")
)

// BuildType defines the type of a builder.
//
type BuildType int

const (
	Core           BuildType = iota // Dock core.
	Applets                         // Dock all internal applets (C).
	AppletInternal                  // Dock one internal applet (C).
	AppletScript                    // Dock one external applet script (bash, python, ruby).
	AppletCompiled                  // Dock one external applet compiled (go, mono, vala).
)

// GetBuildType try to detect an applet type based on its location and content.
// Could be improved.
//
func GetBuildType(name string) BuildType {
	switch name {
	case "core":
		return Core
	case "applets", "plug-ins":
		return Applets
	}

	pack := appdbus.InfoApplet(name)
	if pack != nil {
		dir := pack.Dir()

		if path.Base(path.Dir(dir)) != cdtype.AppletsDirName { // != "third-party" sounds hacky... (but work for current system and user external dirs)
			return AppletInternal
		}

		if _, e := os.Stat(path.Join(dir, "Makefile")); e == nil { // Got makefile => can build.
			return AppletCompiled
		}
		return AppletScript // An external without makefile, nothing to do so consider it as a scripted applet.
	}

	// Log.Info("buildType set to Internal for applet", name)
	return AppletInternal // Finally it must be an internal one (non instanciable one like dock-rendering).

	// if _, e := os.Stat(app.FileLocation("..", name)); e == nil { // Applet is in external dir.
	// 	if _, e := os.Stat(app.FileLocation("..", name, "Makefile")); e == nil { // Got makefile => can build.
	// 		return AppletCompiled
	// 	}
	// 	// Else it's a scripted (atm the only other module not scripted is in vala and is the previous version of this module).
	// 	return AppletScript
	// }

}

//
//----------------------------------------------------------------[ BUILDERS ]--

// BuilderProgress provides a callback handler for Builders.
//
type BuilderProgress struct {
	f func(float64)
}

// Progress forwards data to the provided callback.
//
func (build *BuilderProgress) Progress(data float64) {
	if build.f != nil {
		build.f(data)
	}
}

// SetProgress sets the callback to forward progress data.
//
func (build *BuilderProgress) SetProgress(f func(float64)) {
	build.f = f
}

// BuilderBase provides basic informations about a build.
//
type BuilderBase struct {
	icon string
	dir  string
}

func (build *BuilderBase) SetIcon(icon string) {
	build.icon = icon
}

func (build *BuilderBase) SetDir(dir string) {
	build.dir = dir
}

func (build *BuilderBase) SourceDir() string {
	return build.dir
}

func (build *BuilderBase) Icon() string {
	return build.icon
}

// BuilderNull is an empty Build for fallback.
//
type BuilderNull struct {
	BuilderBase
	BuilderProgress
}

func (build *BuilderNull) Icon() string                { return "none" }
func (build *BuilderNull) Label() string               { return "" }
func (build *BuilderNull) SourceDir() string           { return "" }
func (build *BuilderNull) Build() error                { return inProgress }
func (build *BuilderNull) SetProgress(f func(float64)) {}

// BuilderCore is the dock core sources builder.
//
type BuilderCore struct {
	BuilderBase
	BuilderProgress
	// Dir string
	// app *AppletUpdate
}

func (build *BuilderCore) Icon() string {
	return "/usr/share/cairo-dock/cairo-dock.svg"
}

func (build *BuilderCore) Label() string {
	return "Core"
}

func (build BuilderCore) SourceDir() string {
	return path.Join(build.dir, "cairo-dock-core")
}

func (build *BuilderCore) Build() error {
	return buildCmake(build.SourceDir(), build.BuilderProgress.f)
}

// Build all Cairo-Dock plug-ins.
//
type BuilderApplets struct {
	BuilderBase
	BuilderProgress
	// Dir string
}

func (build *BuilderApplets) Icon() string {
	return "/usr/share/cairo-dock/icons/icon-extensions.svg"
}

func (build *BuilderApplets) Label() string {
	return "Applets"
}

func (build BuilderApplets) SourceDir() string {
	return path.Join(build.dir, "cairo-dock-plug-ins")
}

func (build *BuilderApplets) Build() error {
	return buildCmake(build.SourceDir(), build.BuilderProgress.f)
}

// Internal applet = C applet provided by the plug-ins package.
//
type BuilderInternal struct {
	BuilderBase
	BuilderProgress
	Module string
	// Icon   string
	// Dir    string
}

func (build *BuilderInternal) Label() string {
	return build.Module
}

func (build *BuilderInternal) Build() error {
	dir := path.Join(build.dir, "build", build.Module)
	if e := os.Chdir(dir); e != nil {
		return e
	}
	return actionMakeAndInstall(build.BuilderProgress.f)
}

// External applet that must be compiled (golang or vala)
//
type BuilderCompiled struct {
	BuilderBase
	BuilderProgress
	Module string
	// Icon   string
	// Dir    string
}

func (build *BuilderCompiled) Label() string {
	return build.Module
}

func (build *BuilderCompiled) Build() error {
	if e := os.Chdir(build.dir); e != nil {
		return e
	}

	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()
	go activityBar(t.C, build.BuilderProgress.f)

	return Log.ExecShow("make")
}

//
//----------------------------------------------------[ BUILD FROM C SOURCES ]--

func buildCmake(dir string, progress func(float64)) error {

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
		Log.Info("Create build directory")
		if e = os.Mkdir(dir, os.ModePerm); e != nil {
			return e
		}
		if e = os.Chdir(dir); e != nil {
			return e
		}
		Log.ExecShow("cmake", "..", "-DCMAKE_INSTALL_PREFIX=/usr", "-Denable-gmenu=no")
	}

	return actionMakeAndInstall(progress)

	//~ PARAM_PLUG_INS:="-Denable-scooby-do=yes -Denable-mail=no -Denable-impulse=yes"
}

// Make clean in build dir.
//
func actionMakeClean() {
	// if !*buildKeep { // Default is to clean build dir.
	Log.Info("Clean build directory")
	Log.ExecShow("make", "clean")
	// }
}

// Compile and install a Cairo-Dock source dir.
// Progress is sent to the update callback provided.
//
func actionMakeAndInstall(progress func(float64)) error {
	jobs := runtime.NumCPU()

	cmd := exec.Command("make", "-j", strconv.Itoa(jobs))
	stdout, _ := cmd.StdoutPipe()
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		Log.Err(err, "Execute make")
		return err
	}
	r := bufio.NewReader(stdout)

	line, err := r.ReadString('\n')
	for err == nil {
		if len(line) > 3 && line[0] == '[' {
			current, e := TrimInt(line[1:4])
			if e == nil {
				progress(float64(current) / 100)
				fmt.Printf("[%3d%%] %s", progress, log.Green(line[7:]))
			}
		} else {
			fmt.Fprint(os.Stdout, line)
		}

		line, err = r.ReadString('\n')
	}

	return Log.ExecShow(CmdSudo, "make", "install")
}

func activityBar(c <-chan time.Time, render func(float64)) {
	var val, step float64
	step = 0.05
	for _ = range c {
		if val+step < 0 || 1 < val+step {
			step = -step
		}
		val += step
		render(val)
	}
}

//

//

func TrimInt(str string) (int, error) {
	return strconv.Atoi(strings.Trim(str, "  \n"))
}
