package vdata

import (
	"github.com/sqp/godock/libs/cdglobal" // Global consts.
	"github.com/sqp/godock/libs/log"      // Display info in terminal.

	"github.com/sqp/godock/widgets/cfbuild"        // The config file builder.
	"github.com/sqp/godock/widgets/cfbuild/cftype" // Types for config file builder usage.
	"github.com/sqp/godock/widgets/confsettings"   // New dock own settings to set save as virtual.

	"os"
	"path/filepath"
)

//
//---------------------------------------------------[ CONF TEST COMMON DATA ]--

var testLogName = "dock-test-conf"

var (
	// inside Godock
	pathTestConf = []string{"test", "test.conf"}
	pathGoGmail  = []string{"applets", "GoGmail", "GoGmail.conf"}
)

// TestInit inits the test builder.
//
func TestInit(source Sourcer, file string) cftype.Grouper {
	confsettings.Settings.SaveEditor = ""
	confsettings.Settings.SaveEnabled = false

	var e error
	file, e = filepath.EvalSymlinks(file)
	if e != nil {
		println("check file", e)
		os.Exit(1)
	}

	log := &log.Log{}
	log.SetName(testLogName)

	def := file
	if filepath.Base(file) == "cairo-dock.conf" {
		def = "/usr/share/cairo-dock/cairo-dock.conf" // TODO: improve (use sourcer).
	}

	build, e := cfbuild.NewFromFile(source, log, file, def, "") // default domain.
	if log.Err(e, "load builder") {
		os.Exit(1)
	}

	build.Log().Info("config file:", file)
	return build
}

// TestPathDefault returns the first command line arg or the default test path.
// returns true if the default test file is used.
//
func TestPathDefault() (string, bool) {
	if len(os.Args) > 1 {
		newpath, e := filepath.Abs(os.Args[1])
		log.Err(e, "filepath.Abs")
		return newpath, false
	}
	return PathTestConf(), true
}

// PathTestConf returns the default config test path.
//
func PathTestConf() string {
	return cdglobal.AppBuildPathFull(pathTestConf...)
}
