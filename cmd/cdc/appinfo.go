package main

import (
	"github.com/sqp/godock/libs/cdglobal"
	"github.com/sqp/godock/libs/packages/editpack"
)

var cmdAppInfo = &Command{
	UsageLine: "appinfo [-d path] appletname...",
	Short:     "appinfo edits applets information",
	Long: `
AppInfo edits applets registration information.

Common flags:
  -d path      Use a custom config directory. Default: ~/.config/cairo-dock
`,
}

func init() {
	cmdAppInfo.Run = runAppInfo // break init cycle
}

var appinfoConfPath = cmdAppInfo.Flag.String("d", "", "")

func runAppInfo(cmd *Command, args []string) {
	setPathAbsolute(appinfoConfPath) // Ensure we have an absolute path for the config dir.

	externalUserDir, e := cdglobal.DirAppletsExternal("") // option config dir
	exitIfFail(e, "start termui")

	packs, e := editpack.PacksExternal(logger, externalUserDir)
	exitIfFail(e, "get external packages")
	editpack.Start(logger, packs)
	return
}
