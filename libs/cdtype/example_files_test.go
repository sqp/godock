package cdtype_test

// Files needed in the "demo" directory
//
//    demo            the executable binary or script, without extension (use the shebang on the first line).
//    demo.conf       the default config file (see above for the options syntax).
//    auto-load.conf  the file describing our applet.
//    applet.go       source to load and start the applet as standalone.
//
// Files naming convention:
//
//   demo              must match the directory name, this is used as applet name.
//   demo.conf         same as applet name with extention .conf.
//   auto-load.conf    constant.
//   applet.go         only used by convention, can be customized.

//
//----------------------------------------------------------[ auto-load.conf ]--

const AutoLoadDotConf = `

[Register]

# Author of the applet
author = AuthorName

# A short description of the applet and how to use it.
description = This is the description of the applet.\nIt can be on several lines.

# Category of the applet : 2 = files, 3 = internet, 4 = Desktop, 5 = accessory, 6 = system, 7 = fun
category = 5

# Version of the applet; change it everytime you change something in the config file. Don't forget to update the version both in this file and in the config file.
version = 0.0.1

# Default icon to use if no icon has been defined by the user. If not specified, or if the file is not found, the "icon" file will be used.
icon =

# Whether the applet will act as a launcher or not (like Pidgin or Transmission)
act as launcher = false


`

//
//---------------------------------------------------------------[ applet.go ]--

const AppletDotGo = `

package main

import (
	demo "demo/src"                      // Package with your applet source.
	"github.com/sqp/godock/libs/appdbus" // Connection to cairo-dock.
)

func main() { appdbus.StandAlone(demo.NewApplet) }

`

//
//----------------------------------------------------------------[ OPTIONAL ]--

// Files optional in the "demo" directory:
//
//    icon            the default icon of the applet (without extension, can be any image format).
//    preview         an image preview of the applet (without extension, can be any image format).
//    Makefile        shortcuts to build and link the applet.

//
//----------------------------------------------------------------[ Makefile ]--

// A Makefile with a default build is only required if you want to use the
// build applet action of the Update applet (really helpful to build/restart).
//
const Makefile = `

TARGET=demo
SOURCE=github.com/sqp/godock/applets

# Default is standard build for current arch.

%: build

build:
	go build -o $(TARGET) $(SOURCE)/$(TARGET)

# make a link to the user external directory for easy install.
link:
	ln -s $(GOPATH)/src/$(SOURCE)/$(TARGET) $(HOME)/.config/cairo-dock/third-party/$(TARGET)

`

//
//---------------------------------------------------------------------[ doc ]--

func Example_files() {}
