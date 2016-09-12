///bin/true; exec /usr/bin/env go run "$0" "$@"

package main

import (
	"github.com/gotk3/gotk3/gtk"

	"github.com/sqp/godock/libs/log"                // Display info in terminal.
	"github.com/sqp/godock/widgets/cfbuild/cfprint" // Print config file builder keys.
	"github.com/sqp/godock/widgets/cfbuild/vdata"   // Virtual data source.
	"github.com/sqp/godock/widgets/pageswitch"      // Switcher for config pages.
)

func main() {
	path, isTest := vdata.TestPathDefault()

	// path := cdglobal.ConfigDirDock("") + "/current_theme/cairo-dock.conf"

	// path := cdglobal.AppBuildPathFull(pathGoGmail...)

	gtk.Init(nil)

	source := vdata.New(log.NewLog(log.Logs), nil, nil)
	build := vdata.TestInit(source, path)
	if build == nil {
		return
	}
	build.BuildAll(pageswitch.New())
	if isTest {
		cfprint.Updated(build)
		build.KeyWalk(vdata.TestValues)
		cfprint.Updated(build)
	} else {
		cfprint.Default(build, true)
	}
}

//

// #V[List with entry;Returns the string;Selected;Even if not in the list] TreeView sort multi choice (V)
// KeyTreeViewMultiChoice=

// 	# KeyLaunchCmdIf KeyType = 'G' // a button to launch a specific command with a condition.

// [Dock]

// 	# KeyListAnimation           KeyType = 'a' // list of available animations.
// 	# KeyListDialogDecorator     KeyType = 't' // list of available dialog decorators.
// 	# KeyListDeskletDecoSimple   KeyType = 'O' // list of available desklet decorations.
// 	# KeyListDeskletDecoDefault  KeyType = 'o' // same but with the 'default' choice too.

// 	# KeyListScreens   KeyType = 'r' // list of screens

// 	# KeyJumpToModuleSimple   KeyType = 'm' // a button to jump to another module inside the config panel.
// 	# KeyJumpToModuleIfExists KeyType = 'M' // same but only if the module exists.
