// +build all Update

package allapps

import "github.com/sqp/godock/services/Update"

func init() {
	AddService("Update", Update.NewApplet)
}
