/* TVPlay is an applet for Cairo-Dock to control DLNA/UPnP devices like network TV and radio.

Dependencies: libgupnp-1.0-dev libgupnp-av-1.0-dev libgssdp-1.0-dev
  libgtk-3-dev
and I guess libglib2.0-dev

Install go and set go environment: you need a valid $GOPATH var and directory.

Download, build and install to your Cairo-Dock external applets dir:
  go get -d github.com/sqp/godock/applets/TVPlay  # download applet and dependencies.

  cd $GOPATH/src/github.com/sqp/godock/applets/TVPlay
  make        # compile the applet.
  make link   # link the applet to your external applet directory.


Icons used:: :


Copyright : (C) 2014 by SQP
E-mail : sqp@glx-dock.org

*/
package main

import (
	"github.com/gotk3/gotk3/gtk"

	"github.com/sqp/godock/services/TVPlay"

	"github.com/sqp/godock/libs/appdbus" // Connection to cairo-dock.
)

// Special hack to prevent threads related crashs.
// http://stackoverflow.com/questions/18647475/threading-problems-with-gtk
// http://stackoverflow.com/questions/13351297/what-is-the-downside-of-xinitthreads

// #include <X11/Xlib.h>
// #cgo pkg-config: x11
import "C"

func GRRTHREADS() {
	C.XInitThreads()
}

//---------------------------------------------------------------[ MAIN CALL ]--

// Program launched. Create and activate applet.
//
func main() {
	GRRTHREADS()
	gtk.Init(nil)
	// app :=  // need to build the gui before launching GTK. Works better without threads related crashes.
	go gtk.Main()
	defer gtk.MainQuit()

	appdbus.StartApplet(TVPlay.NewApplet())
}
