// +build Notifications || all

package allapps

import "github.com/sqp/godock/services/Notifications"

func init() {
	AddService("Notifications", Notifications.NewApplet)
}
