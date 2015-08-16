package dockbus_test

import (
	"github.com/stretchr/testify/assert"

	"github.com/kr/pretty"

	"github.com/sqp/godock/libs/log"
	"github.com/sqp/godock/libs/srvdbus/dockbus"
	"github.com/sqp/godock/libs/srvdbus/dockpath"

	"os"
	"testing"
)

// To print the results, launch the test with any option:
// go test -- anything

// Need an active dock with a clock applet for a successful test.

func TestDockbus(t *testing.T) {
	args := os.Args
	show := len(args) > 1 && args[1] == "--"

	dockpath.DbusPathDock = "/org/cdc/Cdc"

	info := dockbus.InfoApplet("clock")
	if assert.NotNil(t, info, "InfoApplet") {
		assert.Equal(t, info.DisplayedName, "clock", "InfoApplet")
	}

	instances, e := dockbus.AppletInstances("clock")
	assert.NoError(t, e, "AppletInstances")
	assert.NotEmpty(t, instances, "AppletInstances, found no active instance of applet clock")

	test := "type=Module"
	props, e := dockbus.DockProperties(test)
	assert.NoError(t, e, "DockProperties")
	assert.NotEmpty(t, props, "DockProperties")

	applets := dockbus.ListKnownApplets()
	if assert.NotEmpty(t, applets, "ListKnownApplets") {
		app, ok := applets["clock"]
		if assert.True(t, ok, "ListKnownApplets field clock") {
			assert.Equal(t, app.DisplayedName, "clock", "ListKnownApplets name clock")
		}
	}

	icons := dockbus.ListIcons()
	assert.NotEmpty(t, icons, "ListIcons")

	if show {
		log.Info("AppletInstances clock")
		pretty.Println(instances)

		log.Info("DockProperties", "found", len(props))
		if len(props) > 0 {
			pretty.Println("first:", props[0])
		}

		log.Info("ListKnownApplets", "found", len(applets))
		for _, app := range applets {
			pretty.Println("first:", app)
			break
		}

		log.Info("ListIcons: found", len(icons))
		if len(icons) > 0 {
			pretty.Println("first:", icons[0])
		}

		log.Info("InfoApplet clock")
		pretty.Println(info)
	}
}
