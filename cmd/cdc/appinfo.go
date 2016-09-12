package main

import (
	"github.com/sqp/godock/libs/packages"
	"github.com/sqp/godock/libs/packages/editpack"
)

var cmdAppInfo = &Command{Run: runAppInfo,
	UsageLine: "appinfo [-d path] appletname...",
	Short:     "appinfo edits applets information",
	Long: `
AppInfo edits applets registration information.

Common flags:
  -d path      Use a custom config directory. Default: ~/.config/cairo-dock
`,
}

var appinfoConfPath = cmdReset.Flag.String("d", "", "")

func runAppInfo(cmd *Command, args []string) {
	setPathAbsolute(appinfoConfPath) // Ensure we have an absolute path for the config dir.

	externalUserDir, e := packages.DirAppletsExternal("") // option config dir
	exitIfFail(e, "start termui")

	packs, e := editpack.PacksExternal(logger, externalUserDir)
	exitIfFail(e, "get external packages")
	editpack.Start(logger, packs)
	return
}
