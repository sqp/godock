// Package allapps declares applets available for the applet loader service.
package allapps

import "github.com/sqp/godock/libs/dock"

// Common fields filled by declared applets.
var apps = make(map[string]func() dock.AppletInstance)
var onStop = make(map[string]func())

func AddService(name string, app func() dock.AppletInstance) {
	apps[name] = app
}

func List() map[string]func() dock.AppletInstance {
	return apps
}

func AddOnStop(name string, call func()) {
	onStop[name] = call
}

func OnStop() {
	for _, f := range onStop {
		f()
	}
}
