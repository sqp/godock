// Package devpage provides a dock developer tools widget.
//
// Use the config builder with a virtual setup.
//
package devpage

import (
	"github.com/sqp/godock/libs/cdglobal"  // Global consts.
	"github.com/sqp/godock/libs/cdtype"    // Logger type.
	"github.com/sqp/godock/libs/text/tran" // Translate.

	"github.com/sqp/godock/widgets/cfbuild"         // The config file builder.
	"github.com/sqp/godock/widgets/cfbuild/cfprint" // Print config file builder keys.
	"github.com/sqp/godock/widgets/cfbuild/cftype"  // Types for config file builder usage.
	"github.com/sqp/godock/widgets/cfbuild/newkey"  // Create config file builder keys.
	"github.com/sqp/godock/widgets/common"          // Text format gtk.
	"github.com/sqp/godock/widgets/pageswitch"      // Switcher for config pages.

	"path/filepath"
	"strings"
)

// New creates a welcome widget with informations about the program.
//
func New(source cftype.Source, log cdtype.Logger, switcher *pageswitch.Switcher) cftype.Grouper {
	const group = "Dev"
	title := tran.Slate("hi")

	// all packages in the application gopath.
	pathGoTest := strings.Join(append(cdglobal.AppBuildPath, "..."), "/")

	pathTestConfCmd := cdglobal.AppBuildPathFull("test", "confcmd", "confcmd.go")
	pathTestConfGUI := cdglobal.AppBuildPathFull("test", "confgui", "confgui.go")
	pathTestConfCmd, _ = filepath.EvalSymlinks(pathTestConfCmd)
	pathTestConfGUI, _ = filepath.EvalSymlinks(pathTestConfGUI)

	printConfig := func(showAll bool) {
		path := source.MainConfigFile()
		def := source.MainConfigDefault()

		otherSw := pageswitch.New()
		defer otherSw.Destroy()
		build, e := cfbuild.NewFromFile(source, log, path, def, "")
		if !log.Err(e, "load current dock config file") {
			// build.BuildSingle("TaskBar")
			build.BuildAll(otherSw)
			println("conf", path, def)
			cfprint.Default(build, showAll)
			build.Destroy()
		}
	}

	buildInfo := cfbuild.TweakAddGroup(group,
		newkey.TextLabel(group, "txt_title", common.Bold(common.Big(title))),
		newkey.Separator(group, "sep_title"),
		newkey.TextLabel(group, "txt_dev_page", "Test page, with useful tools for the developer."),
		newkey.CustomButtonLabel(group, "printConfig", "Print configuration", "show mainconf edited", func() { printConfig(false) }),
		newkey.CustomButtonLabel(group, "printConfig", "Print configuration", "show mainconf all", func() { printConfig(true) }),
		newkey.Separator(group, "sep_go_area"),
		newkey.TextLabel(group, "txt_go_area", "<b>Those commands requires the application sources in their Go environment</b>."),
		newkey.Separator(group, "sep_tests_gui"),
		newkey.LaunchCommand(group, "testConfGUI", "Launch config GUI test", "go run "+pathTestConfGUI),
		newkey.Separator(group, "sep_tests_cmd"),
		newkey.LaunchCommand(group, "testConfGUI", "Launch config console test", "go run "+pathTestConfCmd),
		newkey.LaunchCommand(group, "testConfGUI", "Launch config console mainconf diff", "go run "+pathTestConfCmd+" "+source.MainConfigFile()),
		newkey.Separator(group, "sep_tests_go"),
		newkey.LaunchCommand(group, "gotest", "Launch go tests", "go test "+pathGoTest),
		newkey.LaunchCommand(group, "golint", "Launch go lint", "golint "+pathGoTest),
		newkey.LaunchCommand(group, "govet", "Launch go vet", "go vet "+pathGoTest),
	)

	build := cfbuild.NewVirtual(source, log, "", "", "").BuildAll(switcher, buildInfo)

	build.ShowAll()
	return build
}
