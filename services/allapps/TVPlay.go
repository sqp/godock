// +build all TVPlay

package allapps

import "github.com/sqp/godock/services/TVPlay"

func init() {
	AddService("TVPlay", TVPlay.NewApplet)
	AddGtkNeeded()
}

// var once bool

// func startTVPlay() dock.AppInstance {
// 	if !once {
// 		once = true
// 		TVPlay.GRRTHREADS()
// 		gtk.Init(&[]string{"TVPlay"}) // check that there is not better way to set the class, or that there is no drawback to just use this setting like this.
// 	}
// 	go gtk.Main()
// 	return TVPlay.NewApplet()
// }
