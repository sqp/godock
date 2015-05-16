// +build Cpu || all

package allapps

import "github.com/sqp/godock/services/Cpu"

func init() {
	AddService("Cpu", Cpu.NewApplet)
}
