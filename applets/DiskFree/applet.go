/*
Disk usage monitoring applet for the Cairo-Dock project.

Show disk usage for mounted partitions.
Partitions can be autodetected or you can provide a list of partitions.
You can also use autodetect with some partitions names to be listed first.

Partitions names are designed by their mount point like / or /home.
Use the df command to know your partitions list.


--- INSTALL ---

Install go and get go environment: you need a valid $GOPATH var and directory.

Download, build and install to your Cairo-Dock external applets dir:
  go get -d github.com/sqp/godock/applets/DiskFree  # download applet and dependencies.

  cd $GOPATH/src/github.com/sqp/godock/applets/DiskFree
  make        # compile the applet.
  make link   # link the applet to your external applet directory.

Copyright : (C) 2014 by SQP.
  E-mail : sqp@glx-dock.org

*/
package main

import (
	"github.com/sqp/godock/libs/dock" // Connection to cairo-dock.
	"github.com/sqp/godock/services/DiskFree"
)

//---------------------------------------------------------------[ MAIN CALL ]--

// Program launched. Create and activate applet.
//
func main() {
	dock.StartApplet(DiskFree.NewApplet())
}
