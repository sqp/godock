// +build Mem

package allapps

import "github.com/sqp/godock/services/Mem"

func init() {
	AddService("Mem", Mem.NewApplet)
}
