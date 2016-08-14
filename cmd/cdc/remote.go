package main

import (
	"github.com/sqp/godock/libs/srvdbus"

	"fmt"
	"strings"
)

var cmdRemote = &Command{
	Run:       runRemote,
	UsageLine: "remote command [args...]",
	Short:     "remote control of the dock",
	Long: `
Remote sends a remote command to the active dock.

  d   Debug appletName [state]
        Set the debug state of a go applet.
          To disable the debug, use state: false, no, 0.
          All other state value will enable the debug for the applet.

  sb  SourceCodeBuildTarget
        Build the current source code target.
  sg  SourceCodeGrepTarget grepString
        Grep text in the current source code target dir.
  so  SourceCodeOpenFile filePath
        Open a file in the current source code target dir (if relative).

  ul  UpToShareLastLink
        Print the link of the last uploaded item.
`,
}

func runRemote(cmd *Command, args []string) {
	if len(args) == 0 { // Ensure we have some data.
		cmd.Usage()
	}

	var e error
	switch args[0] {
	case "d", "debug", "Debug", "AppletDebug":
		if len(args) == 1 {
			cmd.Usage()
		}

		state := true
		if len(args) > 2 {
			state = parseState(args[2])
		}

		e = srvdbus.AppletDebug(args[1], state)

	case "sb", "SourceCodeBuildTarget":
		e = srvdbus.SourceCodeBuildTarget()

	case "sg", "SourceCodeGrepTarget":
		e = srvdbus.SourceCodeGrepTarget(args[1])

	case "so", "SourceCodeOpenFile":
		e = srvdbus.SourceCodeOpenFile(args[1])

	case "ul", "UpToShareLastLink":
		var link string
		link, e = srvdbus.UpToShareLastLink()
		if e == nil {
			fmt.Println(link)
		}

	default:
		logger.NewErrf("unknown remote command", "%v", args)
		cmd.Usage()
	}

	if e != nil {
		logger.NewErrf(e.Error(), "remote call %v  %s", args, e.Error())
	}
}

func parseState(state string) bool {
	switch strings.ToLower(state) {
	case "false", "no", "0":
		return false
	}
	return true
}
