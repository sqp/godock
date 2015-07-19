// Package build builds cairo-dock or applets from sources.
package build

import (
	"github.com/sqp/godock/libs/cdglobal"
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/srvdbus/dlogbus"
	"github.com/sqp/godock/libs/text/color"
	"github.com/sqp/godock/libs/text/linesplit" // Parse command output.

	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	// IconMissing defines the optional path to the default icon package emblem.
	//
	IconMissing string

	// CmdSudo defines the command used to get root access for installation.
	//
	CmdSudo = "gksudo"

	labelCore    = "Core"
	labelApplets = "Applets"
	labelGodock  = "cdc"

	errInProgress = errors.New("not finished")
)

// Set by the dock or external backend.
var (
	// AppletInfo returns an applet location and icon.
	//
	AppletInfo func(string) (dir, icon string)

	// AppletRestart restarts an applet.
	//
	AppletRestart func(name string)

	CloseGui = func() {}

	// dirDockData is the location of dock data for icons (overriden by gldi backend with real values).
	//
	dirShareData = "/usr/share/cairo-dock"

	// iconCore defines the name of the dock icon (overriden by gldi backend with real values).
	//
	iconCore = "cairo-dock.svg"
)

// SourceType defines the type of a builder.
type SourceType int

// Applets build types.
const (
	TypeNull           SourceType = iota // Do nothing.
	TypeCore                             // Dock core.
	TypeApplets                          // Dock all internal applets (C).
	TypeAppletInternal                   // Dock one internal applet (C).
	TypeAppletScript                     // Dock one external applet script (bash, python, ruby).
	TypeAppletCompiled                   // Dock one external applet compiled (go, mono, vala).
	TypeGodock                           // New dock.
)

// GetSourceType try to detect an applet type based on its location and content.
// Could be improved.
//
func GetSourceType(name string) SourceType {
	switch name {
	case "core":
		return TypeCore
	case "applets", "plug-ins":
		return TypeApplets
	case "cdc":
		return TypeGodock
	}

	dir, _ := AppletInfo(name)
	if dir != "" {
		if filepath.Base(filepath.Dir(dir)) != cdglobal.AppletsDirName { // AppletsDirName is used for system and user external dirs.
			return TypeAppletInternal
		}

		if _, e := os.Stat(filepath.Join(dir, "Makefile")); e == nil { // Got makefile => can build.
			return TypeAppletCompiled
		}

		return TypeAppletScript // An external without makefile, nothing to do so consider it as a scripted applet.
	}

	return TypeAppletInternal // Finally it must be an internal one (non instanciable one like dock-rendering).
}

//
//------------------------------------------------------------[ BUILD TARGET ]--

// Builder defines the common builder interface.
//
type Builder interface {
	Label() string
	Icon() string
	Build() error
	SourceDir() string
	SetProgress(func(float64)) // Need values between 0 and 1 in the renderer.
	Progress(float64)
	SetIcon(icon string)
	SetDir(dir string)
}

// NewBuilder creates the target renderer/builder.
//
func NewBuilder(target string, log cdtype.Logger) Builder {
	switch GetSourceType(target) {

	case TypeGodock:
		build := &BuilderGodock{}
		build.SetLogger(log)
		return build

	case TypeCore:
		build := &BuilderCore{}
		build.SetLogger(log)
		return build

	case TypeApplets:
		build := &BuilderApplets{}
		build.SetLogger(log)
		return build

	case TypeAppletCompiled:
		dir, icon := AppletInfo(target)
		if dir == "" {
			log.NewErr("applet not found: "+target, "set build target")
			return &BuilderNull{}
		}

		build := &BuilderCompiled{Module: target}
		build.SetLogger(log)
		build.SetIcon(icon)
		build.SetDir(dir)
		return build

	case TypeAppletInternal:
		// Ask icon of module to the Dock as we can't guess its dir and icon name.
		_, icon := AppletInfo(strings.Replace(target, "-", " ", -1))
		if icon == "" {
			icon = IconMissing
		}

		build := &BuilderInternal{Module: target}
		build.SetLogger(log)
		build.SetIcon(icon)
		return build
	}

	// ensure we have a valid target.
	return &BuilderNull{}
}

//
//--------------------------------------------------------[ BUILDER PROGRESS ]--

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

//
//------------------------------------------------------------[ BUILDER BASE ]--

// BuilderBase provides basic informations about a build.
//
type BuilderBase struct {
	icon string
	dir  string
	log  cdtype.Logger
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

// SetLogger sets the builder logger.
//
func (build *BuilderBase) SetLogger(log cdtype.Logger) {
	build.log = log
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
func (build *BuilderNull) Icon() string { return "none" }

// Label returns the builder label.
//
func (build *BuilderNull) Label() string { return "" }

// SourceDir returns the source path of the builder.
//
func (build *BuilderNull) SourceDir() string { return "" }

// Build builds the source code.
//
func (build *BuilderNull) Build() error { return errInProgress }

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
	return filepath.Join(dirShareData, iconCore) // TODO: improve with a dedicated icon.
}

// Label returns the builder label.
//
func (build *BuilderGodock) Label() string {
	return labelGodock
}

// SourceDir returns the source path of the builder.
//
func (build BuilderGodock) SourceDir() string {
	return filepath.Join(cdglobal.AppBuildPathFull()...)
}

// Build builds the source code.
//
func (build *BuilderGodock) Build() error {
	path := build.SourceDir()
	if path == "" {
		return errors.New("GOPATH is not set")
	}

	cmd := build.log.ExecCmd("make", "dock")
	cmd.Dir = path
	e := cmd.Run()
	if e != nil {
		return e
	}

	CloseGui()
	go build.log.Err(dlogbus.Action((*dlogbus.Client).Restart), "restart") // No need to wait an answer, it blocks.
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
	return filepath.Join(dirShareData, iconCore)
}

// Label returns the builder label.
//
func (build *BuilderCore) Label() string {
	return labelCore
}

// Build builds the source code.
//
func (build *BuilderCore) Build() error {
	return build.buildCSources(build.SourceDir(), build.BuilderProgress.f, build.MakeFlags)
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
	return filepath.Join(dirShareData, "icons", "icon-extensions.svg")
}

// Label returns the builder label.
//
func (build *BuilderApplets) Label() string {
	return labelApplets
}

// Build builds the source code.
//
func (build *BuilderApplets) Build() error {
	return build.buildCSources(build.SourceDir(), build.BuilderProgress.f, build.MakeFlags)
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
	e := build.makeBuild(dir, build.BuilderProgress.f)
	if e != nil {
		return e
	}
	return build.makeInstall(dir)
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
	onStop := startActivityBar(build.BuilderProgress.f)
	defer onStop()

	cmd := build.log.ExecCmd("make")
	cmd.Dir = build.dir
	return cmd.Run()
}

//
//----------------------------------------------------[ BUILD FROM C SOURCES ]--

// buildCSources builds C sources from cmake to install.
//
// Progress is sent to the update callback provided.
//
func (build *BuilderBase) buildCSources(dir string, progress func(float64), makeFlags string) error {
	if _, e := os.Stat(dir); e != nil { // basedir must exist.
		return e
	}

	// Initialise build subdir. Create or clean previous compile.
	dir = filepath.Join(dir, "build")

	e := build.cmake(dir, makeFlags)
	if e != nil {
		return e
	}

	e = build.makeBuild(dir, progress)
	if e != nil {
		return e
	}
	return build.makeInstall(dir)
}

// cmake creates the build subdir and launch cmake (like a ./configure).
//
// If the build subdir already exists, launch make clean.
//
//   dir must be the build subdir full path.
//
func (build *BuilderBase) cmake(dir string, makeFlags string) error {

	if _, e := os.Stat(dir); e == nil { // Subdir exists. Clean if needed.
		build.makeClean(dir)
		return nil // Ignore clean error, can fail because it's already too clean.
	}

	// Create build dir and launch cmake.
	build.log.Info("Create build directory")
	if e := os.Mkdir(dir, os.ModePerm); e != nil {
		return e
	}

	args := []string{"..", "-DCMAKE_INSTALL_PREFIX=/usr"}
	if makeFlags != "" {
		args = append(args, strings.Fields(makeFlags)...)
	}
	build.log.Info("cmake", args)
	cmd := build.log.ExecCmd("cmake", args...)
	cmd.Dir = dir
	return cmd.Run()

	//~ PARAM_PLUG_INS:="-Denable-scooby-do=yes -Denable-mail=no -Denable-impulse=yes"
}

// makeClean cleans the build subdir.
//
func (build *BuilderBase) makeClean(dir string) error {
	// if !*buildKeep { // Default is to clean build dir.
	build.log.Debug("Clean build directory")

	cmd := build.log.ExecCmd("make", "clean")
	cmd.Dir = dir
	return cmd.Run()
}

// makeBuild builds sources in the build subdir.
//
// Progress is sent to the update callback provided.
//
func (build *BuilderBase) makeBuild(dir string, progress func(float64)) error {
	jobs := strconv.Itoa(runtime.NumCPU())

	build.log.Info("make", "-j", jobs)
	cmd := build.log.ExecCmd("make", "-j", jobs)
	cmd.Dir = dir

	lastvalue := 0
	cmd.Stdout = linesplit.NewWriter(func(line string) {
		curvalue, curstr, text := trimInt(line)
		if curvalue > -1 {
			if curvalue > lastvalue {
				progress(float64(curvalue) / 100)
				lastvalue = curvalue
			}
			fmt.Printf("%s %s\n", curstr, color.Green(text))

		} else {
			println(line)
		}
	})

	e := cmd.Run()
	if e != nil {
		return e
	}

	return build.makeInstall(dir)
}

// makeInstall installs sources from the build subdir.
//
func (build *BuilderBase) makeInstall(dir string) error {
	cmd := build.log.ExecCmd(CmdSudo, "make", "install")
	cmd.Dir = dir
	return cmd.Run()
}

//
//-----------------------------------------------------------------[ HELPERS ]--

func startActivityBar(render func(float64)) func() {
	t := time.NewTicker(100 * time.Millisecond)
	quit := make(chan struct{})

	go activityBar(quit, t.C, render)

	return func() {
		t.Stop()
		quit <- struct{}{}
		close(quit)
	}
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

// trimInt parses build output to find the progress percent.
//
// Recursive to return the last value if more than one has been found.
//
// input: "[ 42%] Building..."
// output: 42, "[ 42%]", "Building..."
func trimInt(line string) (curvalue int, curstr, text string) {
	if len(line) < 6 || line[0] != '[' {
		return -1, "", ""
	}

	current, e := strconv.Atoi(strings.Trim(line[1:4], " "))
	if e != nil {
		return -1, "", ""
	}

	if len(line) > 7 && line[7] == '[' {
		curvalue, curstr, text := trimInt(line[7:])
		if curvalue > -1 {
			return curvalue, curstr, text
		}
	}

	return current, line[:6], strings.TrimLeft(line[7:], " ")
}
