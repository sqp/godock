/*
Update and build your dock from sources with this applet for Cairo-Dock.

Play with cairo-dock sources:
download/update, compile, restart dock... Usefull for developers, testers and
users who want to stay up to date, or maybe on a distro without packages.

This file will build the applet as standalone.

Install

Install go and get go environment: you need a valid $GOPATH var and directory.

Download, build and install to your Cairo-Dock external applets dir:
  go get -d github.com/sqp/godock/applets/Update  # download applet and dependencies.

  cd $GOPATH/src/github.com/sqp/godock/applets/Update
  make        # compile the applet.
  make link   # link the applet to your external applet directory.



TODO: Version checking:
  check branch, work in progress and new local commits.


Icons used:: some icons from the Oxygen pack:
  http://www.iconarchive.com/show/oxygen-icons-by-oxygen-icons.org.1.html


Copyright : (C) 2012-2016 by SQP
E-mail : sqp@glx-dock.org

*/
package main

import (
	"github.com/sqp/godock/libs/appdbus" // Connection to cairo-dock.
	"github.com/sqp/godock/services/Update"
)

//---------------------------------------------------------------[ MAIN CALL ]--

// Program launched. Create and activate applet.
//
func main() {
	appdbus.StartApplet(Update.NewApplet())
}
