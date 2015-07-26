// Package helpfile provides a cairo-dock help widget.
//
// Still use the old help builder with some hacks.
//
package helpfile

import (
	"github.com/sqp/godock/libs/cdglobal"
	"github.com/sqp/godock/libs/cdtype"

	"github.com/sqp/godock/widgets/confbuilder"
	"github.com/sqp/godock/widgets/pageswitch"
)

const groupGeneral = "General"

// New creates a Help widget with more informations about the program.
//
func New(data confbuilder.Source, log cdtype.Logger, switcher *pageswitch.Switcher) *confbuilder.Grouper {
	file := data.DirShareData(cdglobal.ConfigDirPlugIns, "Help", "Help.conf")
	build, e := confbuilder.NewGrouper(data, log, file, "", "")
	if log.Err(e, "Load Keyfile "+file) {
		return nil
	}

	// Hack packs the Docks and Desklets pages into the first group to save space.
	hack := func(build *confbuilder.Builder) {
		gid, e := build.FindGroupÎD(groupGeneral)
		if log.Err(e, "hack help groups") {
			return
		}

		build.AddGroupKey(gid, build.Conf.List("Docks")...)
		build.AddGroupKey(gid, build.Conf.List("Desklets")...)
	}

	// dropped: The Project, Icon, Desklet
	groups := []string{"General", "Icons", "Taskbar", "Useful Features", "Troubleshooting"}
	return build.BuildGroups(switcher, groups, hack)
}
