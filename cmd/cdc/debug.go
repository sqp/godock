package main

import "github.com/sqp/godock/libs/srvdbus"

var cmdDebug = &Command{
	Run:       runDebug,
	UsageLine: "debug appletname [false|no|0]",
	Short:     "debug change the debug state of an applet",
	Long: `
Debug change the debug state of an applet. 

The first argument must be the applet name.

Options:
  false, no, 0    Disable debug.
  (default)       Enable debug.
.`,
}

func runDebug(cmd *Command, args []string) {
	if len(args) == 0 { // Ensure we have some data.
		cmd.Usage()
	}

	state := true
	if len(args) > 1 {
		state = parseState(args[1])
	}

	e := srvdbus.Debug(args[0], state)
	logger.Err(e, "send debug")
}

func parseState(state string) bool {
	switch state {
	case "false", "False", "no", "0":
		return false
	}
	return true
}
