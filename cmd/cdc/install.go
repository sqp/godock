package main

import (
	"github.com/sqp/godock/libs/packages"

	"strings"
)

var cmdInstall = &Command{
	UsageLine: "install [-d] [-l] [-f format] [-json]",
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

	distant, e := packages.ListDistant(version)
	exitIfFail(e, "List distant applets") // Ensure we have the server list.

	options := ""
	if *verbose {
		options = "v" // Tar command verbose option.
	}

	for _, applet := range args {
		applet = strings.Title(applet) // Applets are using a CamelCase format. This will help lazy users
		if pack := distant.Get(applet); pack != nil {
			if !logger.Err(pack.Install(version, options), "install") {
				logger.Info("Applet installed", applet)
			}
			return
		} else {
			logger.NewErr(applet, "unknown applet")
		}
	}
	logger.NewErr("use cdc list to get the list of valid applets names", "applet name needed")
}

// func downloadOne(applet, options string) {
// 	pack := &packages.AppletPackage{DisplayedName: applet}

// }
