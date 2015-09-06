// +build all Cpu

package allapps

import "github.com/sqp/godock/services/Cpu"

func init() {
	AddService("Cpu", Cpu.NewApplet)
}
