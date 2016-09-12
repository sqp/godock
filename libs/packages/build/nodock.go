// +build !dock

package build

import (
	"github.com/sqp/godock/libs/cdtype"          // Logger type.
	"github.com/sqp/godock/libs/srvdbus/dockbus" // Dock remote commands.
)

func init() {
	AppletInfo = func(log cdtype.Logger, name string) (dir, icon string) {
		pack := dockbus.InfoApplet(log, name)
		if pack == nil {
			return "", ""
		}
		return pack.Dir(), pack.Icon
	}

	AppletRestart = func(name string) {
		dockbus.AppletRemove(name + ".conf")
		dockbus.AppletAdd(name)
	}
}
