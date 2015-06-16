// -build dock

package build

import "github.com/sqp/godock/libs/srvdbus/dockbus"

func init() {
	AppletInfo = func(name string) (dir, icon string) {
		pack := dockbus.InfoApplet(name)
		if pack == nil {
			return "", ""
		}
		return pack.Dir(), pack.Icon
	}

	AppletRestart = func(name string) {
		dockbus.Send(dockbus.AppletRemove(name + ".conf"))
		dockbus.Send(dockbus.AppletAdd(name))
	}
}
