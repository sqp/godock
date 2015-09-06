// +build all DiskFree

package allapps

import "github.com/sqp/godock/services/DiskFree"

func init() {
	AddService("DiskFree", DiskFree.NewApplet)
}
