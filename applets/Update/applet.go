/* Update and build your dock from sources with this applet for the Cairo-Dock project.

Install go and get go environment: you need a valid $GOPATH var and directory.

Download, build and install to your Cairo-Dock external applets dir:
  go get -d github.com/sqp/godock/applets/Update  # download applet and dependencies.

  cd $GOPATH/src/github.com/sqp/godock/applets/Update
  make        # compile the applet.
  make link   # link the applet to your external applet directory.



TODO: Version checking:
  get a better bzr result than simple revno if the user is on a different branch
   with an other stack of patches. (need to get the split version to know the real number of missing patches)



Icons used:: some icons from the Oxygen pack:
  http://www.iconarchive.com/show/oxygen-icons-by-oxygen-icons.org.1.html


Copyright : (C) 2012-2014 by SQP
E-mail : sqp@glx-dock.org

*/
package main

import (
	"github.com/sqp/godock/libs/dock" // Connection to cairo-dock.
	"github.com/sqp/godock/services/Update"
)

//---------------------------------------------------------------[ MAIN CALL ]--

// Program launched. Create and activate applet.
//
func main() {
	dock.StartApplet(Update.NewApplet())
}
