// Notifications is a desktop notifications history applet for the Cairo-Dock project.
/*
Install go and get go environment: you need a valid $GOPATH var and directory.

Download, build and install to your Cairo-Dock external applets dir:
  go get -d github.com/sqp/godock/applets/Notifications  # download applet and dependencies.

  cd $GOPATH/src/github.com/sqp/godock/applets/Notifications
  make        # compile the applet.
  make link   # link the applet to your external applet directory.

Copyright : (C) 2012-2014 by SQP
E-mail : sqp@glx-dock.org

*/
package main

import (
	"github.com/sqp/godock/libs/appdbus" // Connection to cairo-dock.
	"github.com/sqp/godock/services/Notifications"
)

//---------------------------------------------------------------[ MAIN CALL ]--

// Program launched. Create and activate applet.
//
func main() {
	appdbus.StartApplet(Notifications.NewApplet())
}
