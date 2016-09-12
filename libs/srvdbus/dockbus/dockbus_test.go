package dockbus_test

import (
	"github.com/stretchr/testify/assert"

	"github.com/sqp/godock/libs/log"              // Display info in terminal.
	"github.com/sqp/godock/libs/srvdbus/dockbus"  // Dock remote commands.
	"github.com/sqp/godock/libs/srvdbus/dockpath" // Path to main dock dbus service.

	"fmt"
	"os"
	"testing"
)

// To print the results, launch the test with any option:
// go test -- anything

// Need an active dock with a clock applet for a successful test.

var logger = log.NewLog(log.Logs)

func TestDockbus(t *testing.T) {
	args := os.Args
	show := len(args) > 1 && args[1] == "--"

	dockpath.DbusPathDock = "/org/cdc/Cdc"

	info := dockbus.InfoApplet(logger, "clock")
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

	applets, e := dockbus.ListKnownApplets(logger)
	if assert.NoError(t, e, "DockProperties") && assert.NotEmpty(t, applets, "ListKnownApplets") {
		app, ok := applets["clock"]
		if assert.True(t, ok, "ListKnownApplets field clock") {
			assert.Equal(t, app.DisplayedName, "clock", "ListKnownApplets name clock")
		}
	}

	icons := dockbus.ListIcons(logger)
	assert.NotEmpty(t, icons, "ListIcons")

	if show {
		log.Info("AppletInstances clock")
		fmt.Println(instances)

		log.Info("DockProperties", "found", len(props))
		if len(props) > 0 {
			fmt.Println("first:", props[0])
		}

		log.Info("ListKnownApplets", "found", len(applets))
		for _, app := range applets {
			fmt.Println("first:", app)
			break
		}

		log.Info("ListIcons: found", len(icons))
		if len(icons) > 0 {
			fmt.Println("first:", icons[0])
		}

		log.Info("InfoApplet clock")
		fmt.Println(info)
	}

	// dockbus.DockQuit()
}
