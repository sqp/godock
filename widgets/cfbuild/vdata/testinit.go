package vdata

import (
	"github.com/sqp/godock/libs/cdglobal"          // Global consts.
	"github.com/sqp/godock/libs/cdtype"            // Logger type.
	"github.com/sqp/godock/libs/dock/confown"      // New dock own settings to set save as virtual.
	"github.com/sqp/godock/widgets/cfbuild"        // The config file builder.
	"github.com/sqp/godock/widgets/cfbuild/cftype" // Types for config file builder usage.

	"flag"
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
func TestInit(source Sourcer, log cdtype.Logger, file string) cftype.Grouper {
	confown.Current.SaveEditor = ""
	confown.Current.SaveEnabled = false

	var e error
	file, e = filepath.EvalSymlinks(file)
	if e != nil {
		println("check file", e.Error())
		os.Exit(1)
	}

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
func TestPathDefault(log cdtype.Logger) (string, bool) {
	flag.Parse() // ensure we don't use a possible flag as file path.
	if len(flag.Args()) > 1 {
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
