// Package devpage provides a dock developer tools widget.
//
// Use the config builder with a virtual setup.
//
package devpage

import (
	"github.com/sqp/godock/libs/cdglobal"     // Global consts.
	"github.com/sqp/godock/libs/cdtype"       // Logger type.
	"github.com/sqp/godock/libs/text/gtktext" // Format text GTK.
	"github.com/sqp/godock/libs/text/tran"    // Translate.

	"github.com/sqp/godock/widgets/cfbuild"         // The config file builder.
	"github.com/sqp/godock/widgets/cfbuild/cfprint" // Print config file builder keys.
	"github.com/sqp/godock/widgets/cfbuild/cftype"  // Types for config file builder usage.
	"github.com/sqp/godock/widgets/cfbuild/newkey"  // Create config file builder keys.
	"github.com/sqp/godock/widgets/pageswitch"      // Switcher for config pages.

	"path/filepath"
	"strings"
)

// GroupDev defines the name of the dev tools group.
const GroupDev = "Dev"

// New creates a dev tools widget with unstable (moving) custom options.
//
func New(source cftype.Source, log cdtype.Logger, switcher *pageswitch.Switcher) cftype.Grouper {
	build := cfbuild.NewVirtual(source, log, "", "", "")
	return build.BuildAll(switcher,
		PageDev(source, log),
	)
}

// PageDev prepares dev tools page for the config builder.
//
func PageDev(source cftype.Source, log cdtype.Logger) func(build cftype.Builder) {
	return cfbuild.TweakAddGroup(GroupDev, KeysDev(source, log)...)
}

// KeysDev prepares dev tools keys for the config builder.
//
func KeysDev(source cftype.Source, log cdtype.Logger) cftype.ListKey {
	var (
		title = tran.Slate("Developer tools.")

		// all packages in the application gopath.
		pathGoTest = strings.Join(append(cdglobal.AppBuildPath, "..."), "/")

		printConfig = func(showAll bool) {
			path := source.MainConfigFile()
			def := source.MainConfigDefault()

			otherSw := pageswitch.New()
			defer otherSw.Destroy()
			otherBuild, e := cfbuild.NewFromFile(source, log, path, def, "")
			if !log.Err(e, "load current dock config file") {
				// build.BuildSingle("TaskBar")
				otherBuild.BuildAll(otherSw)
				println("conf", path, def)
				cfprint.Default(otherBuild, showAll)
				otherBuild.Destroy()
			}
		}

		pathTestConfCmd = cdglobal.AppBuildPathFull("test", "confcmd", "confcmd.go")
		pathTestConfGUI = cdglobal.AppBuildPathFull("test", "confgui", "confgui.go")
	)
	pathTestConfCmd, _ = filepath.EvalSymlinks(pathTestConfCmd)
	pathTestConfGUI, _ = filepath.EvalSymlinks(pathTestConfGUI)

	keys := cftype.ListKey{
		newkey.TextLabel(GroupDev, "txt_title", gtktext.Bold(gtktext.Big(title))),
		newkey.Separator(GroupDev, "sep_title"),
		newkey.TextLabel(GroupDev, "txt_dev_page", "Test page, with useful tools for the developer."),
		newkey.CustomButton(GroupDev, "printConfigEdited", "Print configuration",
			newkey.Call{Label: "mainconf edited", Func: func() { printConfig(false) }},
			newkey.Call{Label: "mainconf all", Func: func() { printConfig(true) }},
		),
		newkey.Separator(GroupDev, "sep_go_area"),
		newkey.TextLabel(GroupDev, "txt_go_area", "<b>Those commands requires the application sources in their Go environment</b>."),
		newkey.Separator(GroupDev, "sep_tests_gui"),
		newkey.LaunchCommand(GroupDev, "testConfGUI", "Launch config GUI test", "go run "+pathTestConfGUI),
		newkey.Separator(GroupDev, "sep_tests_cmd"),
		newkey.LaunchCommand(GroupDev, "testConfGUI", "Launch config console test", "go run "+pathTestConfCmd),
		newkey.LaunchCommand(GroupDev, "testConfGUI", "Launch config console mainconf diff", "go run "+pathTestConfCmd+" "+source.MainConfigFile()),
		newkey.Separator(GroupDev, "sep_tests_go"),
		newkey.LaunchCommand(GroupDev, "gotest", "Launch go tests", "go test "+pathGoTest),
		newkey.LaunchCommand(GroupDev, "gocover", "Launch coverage tests", "make --directory="+cdglobal.AppBuildPathFull()+" cover"),
		newkey.LaunchCommand(GroupDev, "golint", "Launch go lint", "golint "+pathGoTest),
		newkey.LaunchCommand(GroupDev, "govet", "Launch go vet", "go vet "+pathGoTest),
	}

	return keys
}
