package main

import (
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
)

var cmdBuild = &Command{
	UsageLine: "build [-k] [-r] [-h] target",
	Short:     "build a cairo-dock package",
	Long: `
Build builds and install a Cairo-Dock package.

Targets:
  c or core        Build the dock core.
  p or plug-ins    Build all plug-ins.
  applet name      Use the name of the applet directory in cairo-dock-plug-ins.
                   Plug-ins must have been installed first.

Flags:
  -k               Keep build dir unchanged before build (no make clean).
  -r               Reload your target after build.
  -h               Hide the make install flood if any.
  -g               Graphical mode. Use gksudo to request password.

Options:
  -s               Sources directory. Default is current dir.
  -j               Specifies the number of jobs (commands) to run simultaneously.
                   Default = all availables processors. 
`,
}

var buildKeep = cmdBuild.Flag.Bool("k", false, "")
var buildHide = cmdBuild.Flag.Bool("h", false, "")
var buildReload = cmdBuild.Flag.Bool("r", false, "")
var buildGui = cmdBuild.Flag.Bool("g", false, "")
var buildSource = cmdBuild.Flag.String("s", "", "")
var buildJobs = cmdBuild.Flag.Int("j", 0, "")

// Needed
//~ Dest: goopt.Flag([]string{"-d"}, "Dest", "Installation dir. /usr by default."),

func init() {
	cmdBuild.Run = runBuild // break init cycle
}

func runBuild(cmd *Command, args []string) {

	if *buildJobs == 0 { // Default number of cores = max.
		*buildJobs = runtime.NumCPU()
	}

	if *buildSource == "" { // Default dir = current.
		dir, e := os.Getwd()
		exitIfFail(e, "Get dir")
		*buildSource = dir
	} else {
		dir, e := filepath.Abs(*buildSource)
		exitIfFail(e, "Get dir")
		*buildSource = dir
	}

	// localDir = "/home/sqp/Documents/projets/cairo-3.1/"

	if len(args) == 0 { // Ensure we have a build target.
		cmd.Usage()
	}
	message := "Build %s (" + strconv.Itoa(*buildJobs) + " jobs)\n"

	// Main targets.
	switch args[0] {
	case "c", "core":
		log.Printf(message, "core")

		build(path.Join(*buildSource, "cairo-dock-core"))
	case "p", "plug-ins":
		log.Printf(message, "plug-ins")

		build(path.Join(*buildSource, "cairo-dock-plug-ins"))
	}

	// Not a main target. Try to build a module.
	// 	buildModule(localDir, goopt.Args[0])
}

// Test build dir. If found try to clean, else launch cmake. Then build.
//
func build(dir string) {
	chdirFatal(dir)
	// if _, e := os.Stat(dir); e != nil { // No basedir . Quit.
	// 	log.Fatal(e)
	// }
	build := dir + "/build"

	if _, e := os.Stat(build); e == nil { // Dir exists. Clean if needed.
		exitIfFail(os.Chdir(build), "Build dir")
		actionMakeClean()
	} else { // Create build dir and launch cmake.
		log.Println("Create build directory")
		exitIfFail(os.Mkdir(build, os.ModePerm), "Creation failed")
		os.Chdir(build)
		execShow("cmake", "..", "-DCMAKE_INSTALL_PREFIX=/usr", "-Denable-disks=yes", "-Denable-impulse=yes")
	}

	actionMakeAndInstall()
	actionReload()

	//~ PARAM_PLUG_INS:="-Denable-scooby-do=yes -Denable-disks=yes -Denable-mail=no -Denable-impulse=yes"

	//~ PARAM_CORE="-Dforce-gtk2=yes -Denable_gtk_grip=yes"
	//~ #NB_PROC=$(grep -c ^processor /proc/cpuinfo)
}

func buildModule(dir, module string) {
	build := dir + "cairo-dock-plug-ins/build/" + module
	chdirFatal(build)
	actionMakeAndInstall()
	actionReload()
}

// Make clean in build dir.
//
func actionMakeClean() {
	if !*buildKeep { // Default is to clean build dir.
		log.Println("Clean build directory")
		execShow("make", "clean")
	}
}

// 
func actionMakeAndInstall() {
	execShow("make", "-j", strconv.Itoa(*buildJobs))
	args := []string{"make", "install"}
	command := "sudo"
	if *buildGui {
		command = "gksudo"
	}

	if *buildHide {
		exec.Command(command, args...).Run()
	} else {
		execShow(command, args...)
	}
}

// Kill and reload dock.
//
func actionReload() {
	if *buildReload {
		exec.Command("killall", "cairo-dock").Run()
		exec.Command("cairo-dock").Start()
	}
}

func chdirFatal(dir string) {
	exitIfFail(os.Chdir(dir), "Build dir")
}

// Run command with output forwarded to console.
//
func execShow(command string, args ...string) {
	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if e := cmd.Run(); e != nil {
		log.Fatal(e)
	}
}
