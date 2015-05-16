// +build Audio || all

package allapps

import "github.com/sqp/godock/services/Audio"

func init() {
	AddService("Audio", Audio.NewApplet)
}
