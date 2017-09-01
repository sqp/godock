// +build dock

package main

import (
	"github.com/sqp/godock/libs/srvdbus"

	"fmt"
)

func init() {
	cmdRemote.Long += `
With new dock started

  ar  AppletRenew appletName...
        Rewrite applets configuration, for devs upgrading applets.
        The file is renewed from default, and current values reapplied.
`
	remoteDockArgs = func(cmd *Command, args []string) bool {
		switch args[0] {
		case "ar", "AppletRenew":
			if len(args) < 2 {
				cmd.Usage()
			}
			files, e := srvdbus.AppletRenew(args[1:]...)
			for _, f := range files {
				fmt.Println(f)
			}
			logger.Err(e, "remote call %s", "AppletRenew")
			return true
		}
		return false
	}
}
