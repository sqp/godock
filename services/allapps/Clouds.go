// +build all Clouds

package allapps

import "github.com/sqp/godock/services/Clouds"

func init() {
	AddService("Clouds", Clouds.NewApplet)
}
