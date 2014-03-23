/* Disk activity monitoring applet for the Cairo-Dock project.

Install go and get go environment: you need a valid $GOPATH var and directory.

Download, build and install to your Cairo-Dock external applets dir:
  go get github.com/sqp/godock/applets/DiskSpeed  # download applet and dependencies.

  cd $GOPATH/src/github.com/sqp/godock/applets/DiskSpeed
  make build  # compile the applet.
  make link   # link the applet to your external applet directory.

Copyright : (C) 2012-2014 by SQP
E-mail : sqp@glx-dock.org

*/
package main

import (
	"github.com/sqp/godock/applets/DiskSpeed/src"
	"github.com/sqp/godock/libs/dock" // Connection to cairo-dock.
)

//---------------------------------------------------------------[ MAIN CALL ]--

// Program launched. Create and activate applet.
//
func main() {
	dock.StartApplet(src.NewApplet())
}