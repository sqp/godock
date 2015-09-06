// +build all Audio

package allapps

import "github.com/sqp/godock/services/Audio"

func init() {
	AddService("Audio", Audio.NewApplet)
}
