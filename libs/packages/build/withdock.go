// +build dock

package build

import (
	"github.com/sqp/godock/libs/gldi"
	"github.com/sqp/godock/libs/gldi/globals"
)

func init() {
	dirShareData = globals.DirShareData()

	iconCore = globals.CairoDockIcon

	AppletInfo = func(name string) (dir, icon string) {
		mod := gldi.ModuleGet(name)
		if mod == nil {
			return "", ""
		}
		vc := mod.VisitCard()
		return vc.GetShareDataDir(), vc.GetIconFilePath()
	}

	AppletRestart = func(name string) {
		mod := gldi.ModuleGet(name)
		if mod == nil {
			return
		}
		for _, mi := range mod.InstancesList() {
			gldi.ObjectDelete(mi)
		}
		mod.Activate()
	}
}
