// +build dock

package srvdbus

import (
	"github.com/godbus/dbus"

	"github.com/sqp/godock/libs/gldi/docklist"
	"github.com/sqp/godock/libs/srvdbus/dbuscommon" // Dbus service.

	"strings"
)

// AppletRenew resets applets conf. Returns the list of files changed.
//
func (load *Loader) AppletRenew(applets []string) ([]string, *dbus.Error) {
	var (
		files   []string
		missing []string
	)

	dockapps := docklist.Module()

	for _, name := range applets {
		app, ok := dockapps[name]
		if ok {
			files = append(files, app.ConfFileUpgrade()...)
		} else {
			missing = append(missing, name)
		}
	}
	load.Log.Debug("AppletRenew rewritten", files)
	if len(missing) > 0 {
		load.Log.NewErr("AppletRenew", "applet not found", missing)
		return files, dbuscommon.NewError("applet not found: " + strings.Join(missing, "; "))
	}
	return files, nil
}

// AppletRenew forwards action reset applets conf to the dock.
//
func AppletRenew(applets ...string) ([]string, error) {
	client, e := dbuscommon.GetClient(SrvObj, SrvPath)
	if e != nil {
		return nil, e
	}

	var link []string
	e = client.Get("AppletRenew", []interface{}{&link}, applets)
	return link, e
}
