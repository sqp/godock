// Copyright : (C) 2014-2016 by SQP
// E-mail    : sqp@glx-dock.org

/*
TVPlay is an applet for Cairo-Dock to control DLNA/UPnP devices like network TV and radio.

Install

Dependencies:
  libgupnp-1.0-dev libgupnp-av-1.0-dev libgssdp-1.0-dev
  libgtk-3-dev
  and I guess libglib2.0-dev

Install go and set go environment: you need a valid $GOPATH var and directory.

Download, build and install to your Cairo-Dock external applets dir:
  go get -d -u github.com/sqp/godock/applets/TVPlay  # download applet and dependencies.

  cd $GOPATH/src/github.com/sqp/godock/applets/TVPlay
  make        # compile the applet.
  make link   # link the applet to your external applet directory.


Icons used:: :

*/
package main

import (
	"github.com/gotk3/gotk3/gtk"

	"github.com/sqp/godock/libs/appdbus"    // Connection to cairo-dock.
	"github.com/sqp/godock/services/TVPlay" // Applet service.
)

func main() {
	gtk.Init(nil)
	go func() {
		appdbus.StandAlone(TVPlay.NewApplet)
		gtk.MainQuit()
	}()
	gtk.Main()
}
