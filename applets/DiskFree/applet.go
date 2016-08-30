// Copyright : (C) 2014-2016 by SQP
// E-mail    : sqp@glx-dock.org

/*
Disk usage monitoring applet for the Cairo-Dock project.

Install

Show disk usage for mounted partitions.
Partitions can be autodetected or you can provide a list of partitions.
You can also use autodetect with some partitions names to be listed first.

Partitions names are designed by their mount point like / or /home.
Use the df command to know your partitions list.

Install go and get go environment: you need a valid $GOPATH var and directory.

Download, build and install to your Cairo-Dock external applets dir:
  go get -d -u github.com/sqp/godock/applets/DiskFree  # download applet and dependencies.

  cd $GOPATH/src/github.com/sqp/godock/applets/DiskFree
  make        # compile the applet.
  make link   # link the applet to your external applet directory.

*/
package main

import (
	"github.com/sqp/godock/libs/appdbus"      // Connection to cairo-dock.
	"github.com/sqp/godock/services/DiskFree" // Applet service.
)

func main() { appdbus.StandAlone(DiskFree.NewApplet) }
