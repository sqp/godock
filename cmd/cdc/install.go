package main

import (
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/packages"

	"strings"
)

var cmdInstall = &Command{
	UsageLine: "install [-v] appletname [appletname...]",
	Short:     "install external applet",
	Long: `
Install download and install a Cairo-Dock external applets from the repository.

Flags:
  -v               Verbose output for files extraction.
	`,
}

func init() {
	cmdInstall.Run = runInstall // break init cycle
}

var verbose = cmdInstall.Flag.Bool("v", false, "")

func runInstall(cmd *Command, args []string) {
	if len(args) == 0 { // Ensure we have a target.
		cmd.Usage()
	}

	distant, e := packages.ListDistant(cdtype.AppletsDirName + "/" + appVersion)
	exitIfFail(e, "List distant applets") // Ensure we have the server list.

	options := ""
	if *verbose {
		options = "v" // Tar command verbose option.
	}

	failed := false
	for _, applet := range args {
		pack := distant.Get(strings.Title(applet)) // Applets are using a CamelCase format. This will help lazy users
		failed = failed || !installApplet(pack, options)
	}
	if failed || len(args) == 0 {
		logger.NewErr("use list command to get the list of valid applets names", "applet name needed")
	}
}

// installApplet download and install an applet.
//   pack can be provided untested.
//   options are tar command options.
//
func installApplet(pack *packages.AppletPackage, options string) bool {
	if pack == nil {
		logger.NewErr(pack.DisplayedName, "unknown applet")
		return false
	}
	if logger.Err(pack.Install(appVersion, options), "install") {
		return false
	}
	logger.Info("Applet installed", pack.DisplayedName)
	return true
}
