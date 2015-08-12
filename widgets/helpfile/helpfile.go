// Package helpfile provides a cairo-dock help widget.
//
// Still use the old help builder from file with some hacks.
//
package helpfile

import (
	"github.com/sqp/godock/libs/cdglobal"
	"github.com/sqp/godock/libs/cdtype"

	"github.com/sqp/godock/widgets/cfbuild"
	"github.com/sqp/godock/widgets/cfbuild/cftype"
	"github.com/sqp/godock/widgets/pageswitch"
)

const groupGeneral = "General"

// New creates a Help widget with more informations about the program.
//
func New(data cftype.Source, log cdtype.Logger, switcher *pageswitch.Switcher) (cftype.Grouper, bool) {
	file := data.DirShareData(cdglobal.ConfigDirPlugIns, "Help", "Help.conf")
	build, ok := cfbuild.NewFromFileSafe(data, log, file, "", "")
	if !ok {
		return build, false
	}

	// Hack packs the Docks and Desklets pages into the first group to save space.

	hack := cfbuild.TweakAddKeys(groupGeneral, append(
		build.Storage().List("Docks"),
		build.Storage().List("Desklets")...,
	)...)

	// dropped: The Project, Icon, Desklet
	groups := []string{"General", "Icons", "Taskbar", "Useful Features", "Troubleshooting"}
	return build.BuildGroups(switcher, groups, hack), true
}
