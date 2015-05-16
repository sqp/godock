// +build Update || all

package allapps

import "github.com/sqp/godock/services/Update"

func init() {
	AddService("Update", Update.NewApplet)
}
