package build

import (
	"github.com/sqp/godock/libs/appdbus"
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/log/color"
	"github.com/sqp/godock/libs/srvdbus/dlogbus"

	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var cdcpath = []string{"src", "github.com", "sqp", "godock"}

// Log defines the package logger. Mandatory.
var Log cdtype.Logger

//
//---------------------------------------------------------[ BUILDERS CONFIG ]--

var (
	// CmdSudo defines the command used to get root access for installation.
	CmdSudo = "gksudo"

	errInProgress = errors.New("not finished")
)

// SourceType defines the type of a builder.
type SourceType int

// Applets build types.
const (
	Core           SourceType = iota // Dock core.
	Applets                          // Dock all internal applets (C).
	AppletInternal                   // Dock one internal applet (C).
	AppletScript                     // Dock one external applet script (bash, python, ruby).
	AppletCompiled                   // Dock one external applet compiled (go, mono, vala).
	Godock                           // New dock.
)

// GetSourceType try to detect an applet type based on its location and content.
// Could be improved.
//
func GetSourceType(name string) SourceType {
	switch name {
	case "core":
		return Core
	case "applets", "plug-ins":
		return Applets
	case "cdc":
		return Godock

	}

	pack := appdbus.InfoApplet(name)
	if pack != nil {
		dir := pack.Dir()

		if filepath.Base(filepath.Dir(dir)) != cdtype.AppletsDirName { // != "third-party" sounds hacky... (but work for current system and user external dirs)
			return AppletInternal
		}

		if _, e := os.Stat(filepath.Join(dir, "Makefile")); e == nil { // Got makefile => can build.
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

// SetIcon sets the icon name (or path) for the builder.
//
func (build *BuilderBase) SetIcon(icon string) {
	build.icon = icon
}

// SetDir sets the source path for the builder.
//
func (build *BuilderBase) SetDir(dir string) {
	build.dir = dir
}

// SourceDir returns the source path of the builder.
//
func (build *BuilderBase) SourceDir() string {
	return build.dir
}

// Icon returns the icon name (or path) of the builder.
//
func (build *BuilderBase) Icon() string {
	return build.icon
}

//
//------------------------------------------------------------[ BUILDER NULL ]--

// BuilderNull is an empty Build for fallback.
//
type BuilderNull struct {
	BuilderBase
	BuilderProgress
}

// Icon returns the icon name (or path) of the builder.
//
func (build *BuilderNull) Icon() string {
	return "none"
}

// Label returns the builder label.
//
func (build *BuilderNull) Label() string {
	return ""
}

// SourceDir returns the source path of the builder.
//
func (build *BuilderNull) SourceDir() string {
	return ""
}

// Build builds the source code.
//
func (build *BuilderNull) Build() error {
	return errInProgress
}

//
//----------------------------------------------------------[ BUILDER GODOCK ]--

// BuilderGodock builds the new go dock version.
//
type BuilderGodock struct {
	BuilderBase
	BuilderProgress
	MakeFlags string
}

// Icon returns the icon name (or path) of the builder.
//
func (build *BuilderGodock) Icon() string {
	return "/usr/share/cairo-dock/cairo-dock.svg"
}

// Label returns the builder label.
//
func (build *BuilderGodock) Label() string {
	return "cdc"
}

// SourceDir returns the source path of the builder.
//
func (build BuilderGodock) SourceDir() string {
	path := append([]string{os.Getenv("GOPATH")}, cdcpath...)
	return filepath.Join(path...)
}

// Build builds the source code.
//
func (build *BuilderGodock) Build() error {
	go dlogbus.Action((*dlogbus.Client).Restart) // No need to wait an answer, it blocks.
	return nil
}

//
//------------------------------------------------------------[ BUILDER CORE ]--

// BuilderCore builds the dock core sources.
//
type BuilderCore struct {
	BuilderBase
	BuilderProgress
	MakeFlags string
}

// Icon returns the icon name (or path) of the builder.
//
func (build *BuilderCore) Icon() string {
	return "/usr/share/cairo-dock/cairo-dock.svg"
}

// Label returns the builder label.
//
func (build *BuilderCore) Label() string {
	return "Core"
}

// SourceDir returns the source path of the builder.
//
func (build BuilderCore) SourceDir() string {
	return filepath.Join(build.dir, "cairo-dock-core")
}

// Build builds the source code.
//
func (build *BuilderCore) Build() error {
	return buildCmake(build.SourceDir(), build.BuilderProgress.f, build.MakeFlags)
}

//
//---------------------------------------------------------[ BUILDER APPLETS ]--

// BuilderApplets builds all Cairo-Dock plug-ins.
//
type BuilderApplets struct {
	BuilderBase
	BuilderProgress
	MakeFlags string
}

// Icon returns the icon name (or path) of the builder.
//
func (build *BuilderApplets) Icon() string {
	return "/usr/share/cairo-dock/icons/icon-extensions.svg"
}

// Label returns the builder label.
//
func (build *BuilderApplets) Label() string {
	return "Applets"
}

// SourceDir returns the source path of the builder.
//
func (build BuilderApplets) SourceDir() string {
	return filepath.Join(build.dir, "cairo-dock-plug-ins")
}

// Build builds the source code.
//
func (build *BuilderApplets) Build() error {
	return buildCmake(build.SourceDir(), build.BuilderProgress.f, build.MakeFlags)
}

//
//--------------------------------------------------------[ BUILDER INTERNAL ]--

// BuilderInternal builds a Cairo-Dock internal applet.
// As C applet are provided by the plug-ins package, they are only available
// after the plug-ins build has been done at least once.
//
type BuilderInternal struct {
	BuilderBase
	BuilderProgress
	Module string
}

// Label returns the builder label.
//
func (build *BuilderInternal) Label() string {
	return build.Module
}

// Build builds the source code.
//
func (build *BuilderInternal) Build() error {
	dir := filepath.Join(build.dir, "build", build.Module)
	if e := os.Chdir(dir); e != nil {
		return e
	}
	return actionMakeAndInstall(build.BuilderProgress.f)
}

//
//--------------------------------------------------------[ BUILDER COMPILED ]--

// BuilderCompiled builds external applet that must be compiled (golang or vala).
// A Makefile with default build will have to be provided in each applet dir.
//
type BuilderCompiled struct {
	BuilderBase
	BuilderProgress
	Module string
}

// Label returns the builder label.
//
func (build *BuilderCompiled) Label() string {
	return build.Module
}

// Build builds the source code.
//
func (build *BuilderCompiled) Build() error {
	if e := os.Chdir(build.dir); e != nil {
		return e
	}

	t := time.NewTicker(100 * time.Millisecond)
	quit := make(chan struct{})
	defer func() {
		t.Stop()
		quit <- struct{}{}
		close(quit)
	}()

	go activityBar(quit, t.C, build.BuilderProgress.f)

	return Log.ExecShow("make")
}

//
//----------------------------------------------------[ BUILD FROM C SOURCES ]--

func buildCmake(dir string, progress func(float64), makeFlags string) error {

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

		args := []string{"..", "-DCMAKE_INSTALL_PREFIX=/usr"}
		if makeFlags != "" {
			args = append(args, strings.Fields(makeFlags)...)
		}
		Log.Info("buildCmake args", args)
		e = Log.ExecShow("cmake", args...)
		if e != nil {
			return e
		}
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
			current, e := trimInt(line[1:4])
			if e == nil {
				progress(float64(current) / 100)
				fmt.Printf("%s %s", line[:6], color.Green(line[7:]))

				// Log.Info(line[1:4], current)

			}
		} else {
			fmt.Fprint(os.Stdout, line)
		}

		line, err = r.ReadString('\n')
	}

	cmd.Wait()

	if !cmd.ProcessState.Success() {
		if err != io.EOF {
			Log.Err(err, "make error ?")
			return err
		}
		return errors.New("build fail")
	}

	return Log.ExecShow(CmdSudo, "make", "install")
}

func activityBar(quit chan struct{}, c <-chan time.Time, render func(float64)) {
	var val, step float64
	step = 0.05
	for {
		select {
		case <-c:
			if val+step < 0 || 1 < val+step {
				step = -step
			}
			val += step
			render(val)

		case <-quit:
			return
		}
	}
}

//

//
func trimInt(str string) (int, error) {
	return strconv.Atoi(strings.Trim(str, " \n"))
}
