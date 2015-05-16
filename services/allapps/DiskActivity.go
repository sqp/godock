// +build DiskActivity || all

package allapps

import "github.com/sqp/godock/services/DiskActivity"

func init() {
	AddService("DiskActivity", DiskActivity.NewApplet)
}
