// +build GoGmail || all

package allapps

import "github.com/sqp/godock/services/GoGmail"

func init() {
	AddService("GoGmail", GoGmail.NewApplet)
}
