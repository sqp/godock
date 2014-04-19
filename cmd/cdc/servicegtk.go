// +build gtk

package main

import "github.com/sqp/godock/widgets/common"

func init() {
	gtkStart, gtkStop = common.InitGtk()
}
