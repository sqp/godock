// +build all DiskActivity

package allapps

import "github.com/sqp/godock/services/DiskActivity"

func init() {
	AddService("DiskActivity", DiskActivity.NewApplet)
}
