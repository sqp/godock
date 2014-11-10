// +build dock

package main

import (
	"github.com/sqp/godock/libs/appdbus"
	"github.com/sqp/godock/libs/gldi"
	"github.com/sqp/godock/libs/gldi/backendgui"
	"github.com/sqp/godock/libs/gldi/backendmenu"
	"github.com/sqp/godock/libs/gldi/globals"
	"github.com/sqp/godock/libs/gldi/gui"
	"github.com/sqp/godock/libs/gldi/maindock"
	"github.com/sqp/godock/libs/gldi/menu"

	// loader
	"github.com/sqp/godock/libs/srvdbus"
	"github.com/sqp/godock/services/allapps"

	// web inspection.
	"net/http"
	_ "net/http/pprof"

	// "github.com/sqp/godock/libs/gldi/maindock/views" // custom hacked view

	"fmt"
)

func init() {
	cmdDock = &Command{
		Run:       runDock,
		UsageLine: "dock",
		Short:     "dock starts a custom version of cairo-dock with a new config GUI.",
		Long: `
Dock starts a custom version of cairo-dock with a new GUI.

Options:
  -c          Use Cairo backend.
  -o          Use OpenGL backend.
  -O          Use OpenGL backend with indirect rendering.
              There are very few case where this option should be used.
  -A          Ask again on startup which backend to use.
  -e env      Force the dock to consider this environnement - use it with care.

  -d path     Use a custom config directory. Default: ~/.config/cairo-dock
  -S url      Address of a server with additional themes (overrides default).

  -w time     Wait for N seconds before starting; this is useful if you notice
              some problems when the dock starts with the session.
  -x appname  Exclude a given plug-in from activating (it is still loaded).
  -f          Safe mode: don't load any plug-ins.
  -W          Work around some bugs in Metacity Window-Manager
              (invisible dialogs or sub-docks)
  -l level    Log verbosity (debug,message,warning,critical,error).
              Default is warning.
  -F          Force to display some output messages with colors.
  -k          Lock the dock so that any modification is impossible for users.
  -a          Keep the dock above other windows whatever.
  -s          Don't make the dock appear on all desktops.
  -M path     Ask the dock to load additionnal modules from this directory.
              (though it is unsafe for your dock to load unnofficial modules).

  -N          Don't start the applets service.
  -H          Http debug server: http://localhost:6987/debug/pprof
  -D          Debug mode for the go part of the code (including applets).

  -v          Print version.

This version lacks a lot of options and may not be considered usable for
everybody at the moment.
.`,
	}

	userForceCairo := cmdDock.Flag.Bool("c", false, "")
	userForceOpenGL := cmdDock.Flag.Bool("o", false, "")
	userIndirectOpenGL := cmdDock.Flag.Bool("O", false, "")
	userAskBackend := cmdDock.Flag.Bool("A", false, "")
	userEnv := cmdDock.Flag.String("e", "", "")

	userDir := cmdDock.Flag.String("d", "", "")
	userThemeServer := cmdDock.Flag.String("S", "", "")

	userDelay := cmdDock.Flag.Int("w", 0, "")
	// maintenance
	userExclude := cmdDock.Flag.String("x", "", "")
	userSafeMode := cmdDock.Flag.Bool("f", false, "")
	userMetacityWorkaround := cmdDock.Flag.Bool("W", false, "")
	userVerbosity := cmdDock.Flag.String("l", "", "")
	userForceColor := cmdDock.Flag.Bool("F", false, "")
	userLocked := cmdDock.Flag.Bool("k", false, "")
	userKeepAbove := cmdDock.Flag.Bool("a", false, "")
	userNoSticky := cmdDock.Flag.Bool("s", false, "")
	userModulesDir := cmdDock.Flag.String("M", "", "")

	userVersion = cmdDock.Flag.Bool("v", false, "")
	srvAppletsDisable = cmdDock.Flag.Bool("N", false, "")
	srvHttpPprof = cmdDock.Flag.Bool("H", false, "")
	srvDebug = cmdDock.Flag.Bool("D", false, "")

	dockSettings = func() maindock.DockSettings {
		return maindock.DockSettings{
			ForceCairo:     *userForceCairo,
			ForceOpenGL:    *userForceOpenGL,
			IndirectOpenGL: *userIndirectOpenGL,
			AskBackend:     *userAskBackend,
			Env:            *userEnv,

			UserDefinedDataDir: *userDir,
			ThemeServer:        *userThemeServer,

			Delay:              *userDelay,
			Exclude:            *userExclude,
			SafeMode:           *userSafeMode,
			MetacityWorkaround: *userMetacityWorkaround,
			Verbosity:          *userVerbosity,
			ForceColor:         *userForceColor,
			Locked:             *userLocked,
			KeepAbove:          *userKeepAbove,
			NoSticky:           *userNoSticky,
			ModulesDir:         *userModulesDir,
		}
	}
}

// 	{"maintenance", 'm', G_OPTION_FLAG_IN_MAIN, G_OPTION_ARG_NONE,
// 		&bMaintenance,
// 		_("Allow to edit the config before the dock is started and show the config panel on start."), NULL},
// 	{"exclude", 'x', G_OPTION_FLAG_IN_MAIN, G_OPTION_ARG_STRING,
// 		&cExcludeModule,
// 		_("Exclude a given plug-in from activating (it is still loaded though)."), NULL},

// 	{"testing", 'T', G_OPTION_FLAG_IN_MAIN, G_OPTION_ARG_NONE,
// 		&bTesting,
// 		_("For debugging purpose only. The crash manager will not be started to hunt down the bugs."), NULL},
// 	{"easter-eggs", 'E', G_OPTION_FLAG_IN_MAIN, G_OPTION_ARG_NONE,
// 		&g_bEasterEggs,
// 		_("For debugging purpose only. Some hidden and still unstable options will be activated."), NULL},

var (
	// dockSettings returns maindock settings parsed from the command line.
	dockSettings func() maindock.DockSettings

	userVersion       *bool
	srvAppletsDisable *bool
	srvHttpPprof      *bool
	srvDebug          *bool
)

func runDock(cmd *Command, args []string) {
	if *userVersion {
		println(globals.Version()) // -v option only prints version.
		return
	}

	// Applets service. Enabled by default.
	if !*srvAppletsDisable {
		go appletService()
	}

	// HTTP listener for the pprof debug.
	if *srvHttpPprof {
		go func() { http.ListenAndServe("localhost:6987", nil) }()
	}

	// Logger debug state.
	logger.SetDebug(*srvDebug)

	maindock.SetLogger(logger)

	settings := dockSettings()
	settings.Init()

	gtkA, gtkB, gtkC := globals.GtkVersion()
	logger.Info("Custom Dock", cdcVersion)
	logger.Info("   gldi    ", globals.Version())
	// logger.Info("Compiled date      ", C.__DATE__, C.__TIME__)
	logger.Info("   GTK     ", fmt.Sprintf("%d.%d.%d", gtkA, gtkB, gtkC))
	logger.Info("  OpenGL   ", gldi.GLBackendIsUsed())

	gldi.LoadCurrentTheme() // was after registration but caused some problems with refresh on start.

	backendgui.Register(gui.NewConnector(logger))
	backendmenu.Register("dock", menu.BuildMenuContainer, menu.BuildMenuIcon)
	maindock.RegisterEvents()

	settings.Prepare()

	// views.RegisterPanel("spanel")

	settings.Start()

}

// Start Loader.
//
func appletService() {
	appdbus.DbusPathDock = "/org/cdc/Cdc" // TODO: improve to autodetect.

	loader := srvdbus.NewLoader(allapps.List(), logger)
	if loader == nil {
		return
	}

	active, e := loader.Start(loader, srvdbus.Introspec)
	if logger.Err(e, "Start Applets service") {
		return
	}

	logger.Info("appletService active", active)

	if active {
		// defer allapps.OnStop()
		loader.StartLoop(true)
	}
}
