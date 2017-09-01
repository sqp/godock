package main

import (
	"github.com/sqp/godock/libs/appdbus" // External connection to cairo-dock.

	"github.com/sqp/godock/libs/cdtype/AppTmpl/src" // Package with your applet source.
)

func main() { appdbus.StandAlone(AppTmpl.NewApplet) }
