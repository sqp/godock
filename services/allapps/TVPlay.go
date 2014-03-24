// +build TVPlay

package allapps

import (
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/godock/libs/dock" // Connection to cairo-dock.
	"github.com/sqp/godock/services/TVPlay"
)

var once bool

func init() {
	AddService("TVPlay", startTVPlay)
	AddOnStop("TVPlay", gtk.MainQuit)
}

func startTVPlay() dock.AppletInstance {
	if !once {
		once = true
		TVPlay.GRRTHREADS()
		gtk.Init(&[]string{"TVPlay"}) // check that there is not better way to set the class, or that there is no drawback to just use this setting like this.
	}
	go gtk.Main()
	return TVPlay.NewApplet()
}
