// Package welcome provides a cairo-dock welcome widget.
//
// Use the config builder with a virtual setup.
//
package welcome

import (
	"github.com/sqp/godock/libs/cdtype"    // Logger type.
	"github.com/sqp/godock/libs/text/tran" // Translate.

	"github.com/sqp/godock/widgets/cfbuild"        // The config file builder.
	"github.com/sqp/godock/widgets/cfbuild/cftype" // Types for config file builder usage.
	"github.com/sqp/godock/widgets/cfbuild/newkey" // Create config file builder keys.
	"github.com/sqp/godock/widgets/common"         // Text format gtk.
)

//
const (
	URLwebsite       = "https://github.com/sqp/godock"
	URLdocumentation = "http://godoc.org/github.com/sqp/godock"
	URLdockInfo      = "http://glx-dock.org/bg_topic.php?t=7638"
)

type docLink struct{ Title, URL, Icon, Text string }

// extracted from doc.go. Need to auto regenerate.
var links = []docLink{{
	Title: "Usage",
	URL:   "http://godoc.org/github.com/sqp/godock/cmd/cdc",
	Icon:  "dialog-information",
	Text:  "Once installed, you can run the cdc command to start a dock (if enabled), or interact with the dock and its new applets.",
}, {
	Title: "Install",
	URL:   "http://godoc.org/github.com/sqp/godock/dist",
	Icon:  "drive-harddisk",
	Text:  "Documentation for install, build and package creation is on the dist page, with pre build packages repositories (Debian, Ubuntu, Archlinux).",
}, {
	Title: "Applet creation",
	URL:   "http://godoc.org/github.com/sqp/godock/libs/cdtype",
	Icon:  "list-add",
	Text:  "If you want to create a new applet or learn more about them, most of their work is defined in the cdtype package, with their common types, actions, events...",
},
}

// New creates a welcome widget with informations about the program.
//
func New(source cftype.Source, log cdtype.Logger) cftype.Grouper {
	const group = "Welcome"
	title := tran.Slate("Welcome to cairo-dock-rework")
	header := tran.Slate(`This is a reworked version of cairo-dock, with all the user interface rewritten in Go.
It's still under development, but should now be very close to the original version.`)
	warningSave := `Warning, save configuration is disabled by default.
As it shares its files with the original dock, it requires more tests to ensure nothing will be broken.
It's better to save your current theme, and check nothing wrong will be changed in your files.
Then, you can enable the save option under the "GUI Settings" config tab, at your own risks.`

	keys := []*cftype.Key{
		newkey.TextLabel(group, "title", common.Bold(common.Big(title))),
		newkey.TextLabel(group, "header", header),
		newkey.Separator(group, "sep_title"),
		newkey.TextLabel(group, "warningSave", warningSave),
		newkey.Separator(group, "sep_warning"),
		newkey.Link(group, "URLwebsite", "Project website", "github", URLwebsite),
		newkey.Link(group, "URLdocumentation", "Documentation", "godoc", URLdocumentation),
		newkey.Link(group, "URLdockInfo", "Cairo-Dock forum related thread", "glx-dock forum", URLdockInfo),
	}
	for _, link := range links {
		str := common.Big(common.Bold(common.URI(link.URL, link.Title)))
		keys = append(keys,
			// newkey.Frame(group, "F_"+title, str, link.Icon),
			newkey.Separator(group, "sep_"+link.Title),
			newkey.TextLabel(group, "T_"+link.Title, str),
			newkey.TextLabel(group, "L_"+link.Title, link.Text),
		)
	}

	build := cfbuild.NewVirtual(source, log, "", "", "")
	return build.BuildSingle(group, cfbuild.TweakAddGroup(group, keys...))
}
