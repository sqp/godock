/*
Package maindock is a wrapper to build a dock interface.

Missing

	_cairo_dock_successful_launch - Happy New Year message.
	extern gboolean g_bEasterEggs;
	crash tests and recovery. Not sure what to do about it.
	  static gint s_iNbCrashes = 0;
	  static gboolean s_bSucessfulLaunch = FALSE;
	  static GString *s_pLaunchCommand = NULL;

*/
package maindock

import (
	"github.com/gotk3/gotk3/gtk"

	"github.com/sqp/godock/libs/cdglobal"     // Global consts.
	"github.com/sqp/godock/libs/cdtype"       // Logger type.
	"github.com/sqp/godock/libs/config"       // Config parser.
	"github.com/sqp/godock/libs/files"        // Files operations.
	"github.com/sqp/godock/libs/gldi"         // Gldi access.
	"github.com/sqp/godock/libs/gldi/current" // Current theme settings.
	"github.com/sqp/godock/libs/gldi/dialog"  // Popup dialog.
	"github.com/sqp/godock/libs/gldi/globals" // Global variables.
	"github.com/sqp/godock/libs/ternary"      // Ternary operators.
	"github.com/sqp/godock/libs/text/tran"    // Translate.

	"github.com/sqp/godock/widgets/about"
	"github.com/sqp/godock/widgets/gtk/newgtk"

	"fmt"
	"os"
	"path/filepath"
	"time"
)

var log cdtype.Logger

// SetLogger provides a common logger for the dock. It must be set to start a dock.
//
func SetLogger(l cdtype.Logger) {
	log = l
}

//
//----------------------------------------------------------------[ MAINDOCK ]--

// DockSettings defines the dock settings provided as command flags.
//
type DockSettings struct {
	// --- Original Cairo-Dock settings ---
	//
	ForceCairo     bool
	ForceOpenGL    bool
	IndirectOpenGL bool
	AskBackend     bool
	Env            string

	UserDefinedDataDir string
	ThemeServer        string

	Delay              int
	Exclude            []string
	SafeMode           bool
	MetacityWorkaround bool
	Verbosity          string
	ForceColor         bool

	Locked     bool
	KeepAbove  bool
	NoSticky   bool
	ModulesDir string

	// --- New Dock settings ---
	//
	WebHost     string
	WebPort     int
	WebMonitor  bool
	DisableDBus bool
	Debug       bool

	// auto loaded from original dock hidden file.
	isFirstLaunch  bool
	isNewVersion   bool
	sessionWasUsed bool
}

// Init is the first step to initialize the dock.
//
func (settings *DockSettings) Init() {
	confdir := cdglobal.ConfigDirDock(settings.UserDefinedDataDir)
	settings.isFirstLaunch = !files.IsExist(confdir) // TODO: need test is dir.

	hidden := loadHidden(confdir)

	settings.isNewVersion = hidden.LastVersion != globals.Version()
	settings.sessionWasUsed = hidden.SessionWasUsed

	// MISSING
	// //\___________________ build the command line used to respawn, and check if we have been launched from another life.

	// mute all output messages if CD is not launched from a terminal
	// if (getenv("TERM") == NULL)  /// why not isatty(stdout) ?...
	// 	g_set_print_handler(PrintMuteFunc);

	gldi.DbusGThreadInit() // it's a wrapper: it will use dbus_threads_init_default ();

	gtk.Init(nil)

	//\___________________ C dock log.
	if settings.Verbosity != "" {
		gldi.LogSetLevelFromName(settings.Verbosity)
	}

	if settings.ForceColor {
		gldi.LogForceUseColor()
	}

	//\___________________ delay the startup if specified.
	if settings.Delay > 0 {
		<-time.After(time.Duration(settings.Delay) * time.Second)
	}

	//\___________________ initialize libgldi.

	rendering := gldi.RenderingDefault
	switch {
	case settings.ForceOpenGL:
		rendering = gldi.RenderingOpenGL

	case settings.ForceCairo:
		rendering = gldi.RenderingCairo
	}
	gldi.Init(rendering)

	//\___________________ set custom user options.

	if settings.KeepAbove {
		gldi.ForceDocksAbove()
	}

	if settings.NoSticky {
		gldi.SetContainersNonSticky()
	}

	if settings.MetacityWorkaround {
		gldi.DisableContainersOpacity()
	}

	if settings.Env != "" {
		env := cdglobal.DesktEnvFromString(settings.Env)
		if env == cdglobal.DeskEnvUnknown {
			log.NewWarn(settings.Env, "Unknown desktop environment (valid options: gnome, kde, xfce)")
		} else {
			gldi.FMForceDeskEnv(env)
		}
	}

	if settings.IndirectOpenGL {
		gldi.GLBackendForceIndirectRendering()
	}

	if settings.ThemeServer != "" { // Custom theme server.
		cdglobal.DownloadServerURL = settings.ThemeServer
	}

	gldi.SetPaths(confdir, // will later be available as DirDockData  (g_cCairoDockDataDir)
		cdglobal.ConfigDirExtras,
		cdglobal.ConfigDirDockThemes,
		cdglobal.ConfigDirCurrentTheme,
		cdglobal.CairoDockShareThemesDir,
		cdglobal.DockThemeServerTag,
		cdglobal.DownloadServerURL)

	about.Img = globals.DirShareData(cdglobal.ConfigDirDockImages, cdglobal.FileCairoDockLogo)

	//\___________________ Check that OpenGL is safely usable, if not ask the user what to do.

	// Unsafe OpenGL requires to be confirmed to use (need forced in conf or user validation).
	if settings.AskBackend || (gldi.GLBackendIsUsed() && !gldi.GLBackendIsSafe() && !settings.ForceOpenGL && !settings.IndirectOpenGL) {
		if settings.AskBackend || hidden.DefaultBackend == "" { // no backend defined.
			dialogAskBackend()
		} else if hidden.DefaultBackend != "opengl" { // disable opengl if unused, revert to cairo.
			gldi.GLBackendDeactivate()
		}
	}

	// Remove excluded applets.
	for _, name := range settings.Exclude {
		cdtype.Applets.Unregister(name)
	}

	//\___________________ load plug-ins (must be done after everything is initialized).
	if !settings.SafeMode {
		err := gldi.ModulesNewFromDirectory("")
		log.Err(err, "no module will be available")

		if settings.ModulesDir != "" {
			err := gldi.ModulesNewFromDirectory(settings.ModulesDir)
			log.Err(err, "no additionnal module will be available")
		}
	}
}

// Start starts the dock theme and apply last settings.
//
func (settings *DockSettings) Start() {
	//\___________________ handle crashes.
	// if (! bTesting)
	// 	_cairo_dock_set_signal_interception ();

	//\___________________ handle terminate signals to quit properly (especially when the system shuts down).
	// signal (SIGTERM, _cairo_dock_quit);  // Term // kill -15 (system)
	// signal (SIGHUP,  _cairo_dock_quit);  // sent to a process when its controlling terminal is closed

	// MISSING
	//\___________________ Disable modules that have crashed
	//\___________________ maintenance mode -> show the main config panel.

	// Copy the default theme if needed.
	if !files.IsExist(globals.ConfigFile()) {
		createConfigDir(settings)
	}

	// MISSING
	// The first time the Cairo-Dock session is used but not the first time the dock is launched: propose to use the Default-Panel theme
	// s_bCDSessionLaunched => settings.sessionWasUsed

	gldi.CurrentThemeLoad() // was moved before registration when I had some problems with refresh on start. Removed here for now.

	//\___________________ lock mode.

	if settings.Locked {
		println("Cairo-Dock will be locked.") // was cd_warning (so it was set just before Verbosity). TODO: improve
		// As the config window shouldn't be opened, those settings won't change.
		current.Docks.LockIcons(true)
		current.Docks.LockAll(true)
		globals.FullLock = true
	}

	if !settings.SafeMode && gldi.ModulesGetNb() <= 1 { // 1 including Help.
		dialogNoPlugins()
	}

	if settings.isNewVersion { // update the version in the file.
		updateHiddenFile("last version", globals.Version())

		// If any operation must be done on the user theme (like activating
		// a module by default, or disabling an option), it should be done
		// here once (when CAIRO_DOCK_VERSION matches the new version).
	}

	//\___________________ display the changelog in case of a new version.

	if settings.isFirstLaunch { // first launch => set up config
		time.AfterFunc(4*time.Second, firstLaunchSetup)

	} else if settings.isNewVersion { // new version -> changelog (if it's the first launch, useless to display what's new, we already have the Welcome message).
		dialogChangelog()
		// In case something has changed in Compiz/Gtk/others, we also run the script on a new version of the dock.
		time.AfterFunc(4*time.Second, firstLaunchSetup)
	}

	// else if (cExcludeModule != NULL && ! bMaintenance && s_iNbCrashes > 1) {
	// 	gchar *cMessage;
	// 	if (s_iNbCrashes == 2) // <=> second crash: display a dialogue
	// 		cMessage = g_strdup_printf (_("The module '%s' may have encountered a problem.\nIt has been restored successfully, but if it happens again, please report it at http://glx-dock.org"), cExcludeModule);
	// 	else // since the 3th crash: the applet has been disabled
	// 		cMessage = g_strdup_printf (_("The module '%s' has been deactivated because it may have caused some problems.\nYou can reactivate it, if it happens again thanks to report it at http://glx-dock.org"), cExcludeModule);

	// 	GldiModule *pModule = gldi_module_get (cExcludeModule);
	// 	Icon *icon = gldi_icons_get_any_without_dialog ();
	// 	gldi_dialog_show_temporary_with_icon (cMessage, icon, CAIRO_CONTAINER (g_pMainDock), 15000., (pModule ? pModule->pVisitCard->cIconFilePath : NULL));
	// 	g_free (cMessage);
	// }

	// if (! bTesting)
	// 	g_timeout_add_seconds (5, _cairo_dock_successful_launch, GINT_TO_POINTER (bFirstLaunch));
}

// Clean cleans the dock variables on stop.
//
func Clean() {

	// signal(SIGSEGV, NULL) // Segmentation violation
	// signal(SIGFPE, NULL)  // Floating-point exception
	// signal(SIGILL, NULL)  // Illegal instruction
	// signal(SIGABRT, NULL)
	// signal(SIGTERM, NULL)
	// signal(SIGHUP, NULL)

	gldi.FreeAll()

	// #if (LIBRSVG_MAJOR_VERSION == 2 && LIBRSVG_MINOR_VERSION < 36)
	// rsvg_term ();
	// #endif
	gldi.XMLCleanupParser()
}

//
//---------------------------------------------------------[ CONFIG SETTINGS ]--

func firstLaunchSetup() {
	log.Info("firstLaunchScript", globals.DirShareData("scripts", "initial-setup.sh"))
	log.ExecShow(globals.DirShareData("scripts", "initial-setup.sh"))
}

func createConfigDir(settings *DockSettings) {
	log.Info("creating configuration directory", globals.DirDockData())

	themeName := "Default-Single"

	if os.Getenv("DESKTOP_SESSION") == "cairo-dock" { // We're using the CD session for the first time
		themeName = "Default-Panel"
		settings.sessionWasUsed = true
		updateHiddenFile("cd session", true)
	}

	files.CopyDir(globals.DirShareData(cdglobal.ConfigDirDockThemes, themeName), globals.CurrentThemePath())
}

//
//-----------------------------------------------------------------[ DIALOGS ]--

func dialogAskBackend() {
	// Need to keep the string as it is for translation.
	str := "OpenGL allows you to use the hardware acceleration, reducing the CPU load to the minimum.\nIt also allows some pretty visual effects similar to Compiz.\nHowever, some cards and/or their drivers don't fully support it, which may prevent the dock from running correctly.\nDo you want to activate OpenGL ?\n (To not show this dialog, launch the dock from the Application menu,\n  or with the -o option to force OpenGL and -c to force cairo.)"

	dialog := newgtk.Dialog()
	dialog.SetTitle(tran.Slate("Use OpenGL in Cairo-Dock"))
	dialog.AddButton(tran.Slate("Yes"), gtk.RESPONSE_YES)
	dialog.AddButton(tran.Slate("No"), gtk.RESPONSE_NO)

	labelTxt := newgtk.Label(tran.Slate(str))

	content, _ := dialog.GetContentArea()
	content.PackStart(labelTxt, false, false, 0)

	askBox := newgtk.Box(gtk.ORIENTATION_HORIZONTAL, 3)
	content.PackStart(askBox, false, false, 0)

	labelSave := newgtk.Label(tran.Slate("Remember this choice"))
	check := newgtk.CheckButton()
	askBox.PackEnd(check, false, false, 0)
	askBox.PackEnd(labelSave, false, false, 0)

	dialog.ShowAll()

	answer := dialog.Run() // has its own main loop, so we can call it before gtk_main.
	remember := check.GetActive()
	dialog.Destroy()

	if answer == int(gtk.RESPONSE_NO) {
		gldi.GLBackendDeactivate()
	}

	if remember { // save user choice to file.
		value := ternary.String(gtk.ResponseType(answer) == gtk.RESPONSE_YES, "opengl", "cairo")
		updateHiddenFile("default backend", value)
	}
}

func dialogNoPlugins() {
	// Need to keep the string as it is for translation.
	str := "No plug-in were found.\nPlug-ins provide most of the functionalities (animations, applets, views, etc).\nSee http://glx-dock.org for more information.\nThere is almost no meaning in running the dock without them and it's probably due to a problem with the installation of these plug-ins.\nBut if you really want to use the dock without these plug-ins, you can launch the dock with the '-f' option to no longer have this message.\n"
	icon := gldi.IconsGetAnyWithoutDialog()
	container := globals.Maindock().ToContainer()
	iconpath := globals.FileCairoDockIcon()
	dialog.ShowTemporaryWithIcon(tran.Slate(str), icon, container, 0, iconpath)
}

func dialogChangelog() {
	str := getChangelog()
	if str == "" {
		return
	}
	log.Info("", str) // Also show it on console.
	// TODO: icon shouldn't be nil, grab first icon.
	// 			Icon *pFirstIcon = cairo_dock_get_first_icon (g_pMainDock->icons);
	dialog.NewDialog(nil, globals.Maindock().Container(), cdtype.DialogData{
		Message:   str,
		Icon:      globals.FileCairoDockIcon(),
		UseMarkup: true})
}

func getChangelog() string {
	changelogPath := globals.DirShareData(cdglobal.FileChangelog)
	msg := ""
	e := config.GetFromFile(log, changelogPath, func(cfg cdtype.ConfUpdater) {
		major, minor, micro := globals.VersionSplit()

		// Get first line
		strver := fmt.Sprintf("%d.%d.%d", major, minor, micro) // version without "alpha", "beta", "rc", etc.
		msg = cfg.Valuer("ChangeLog", strver).String()
		if msg == "" {
			log.Debug("changelog", "no info for version", globals.Version())
			return
		}
		msg = tran.Slate(msg)

		// Add all changelog lines for that version.
		i := 0
		for {
			strver := fmt.Sprintf("%d.%d.%d.%d", major, minor, micro, i)
			sub := cfg.Valuer("ChangeLog", strver).String()
			msg += "\n " + tran.Slate(sub)
			i++
		}
	})
	log.Err(e, "load changelog", changelogPath)
	return msg
}

//
//---------------------------------------------------------[ HIDDEN SETTINGS ]--

type hiddenConfig struct {
	LastVersion    string `conf:"last version"`
	DefaultBackend string `conf:"default backend"`
	LastYear       int    `conf:"last year"`
	SessionWasUsed bool   `conf:"cd session"`
}

const hiddenGroup = "Launch"

// loadHidden will try to load the hidden config data from the file.
//
func loadHidden(path string) *hiddenConfig {
	file := filepath.Join(path, cdglobal.FileHiddenConfig)
	hidden := &hiddenConfig{}

	// Load data from file.
	e := config.GetFromFile(log, file, func(cfg cdtype.ConfUpdater) {
		cfg.UnmarshalGroup(hidden, hiddenGroup, config.GetTag)
	})

	if e != nil { // File missing, create it.
		conf := config.NewEmpty(log, file)

		conf.MarshalGroup(hidden, hiddenGroup, config.GetTag)
		conf.Set("Gui", "mode", "0") // another section unused here.

		e = conf.Save()
		log.Err(e, "create hidden conf", file)
	}

	return hidden
}

func updateHiddenFile(key string, value interface{}) {
	file := globals.DirDockData(cdglobal.FileHiddenConfig)
	e := config.UpdateFile(log, file, hiddenGroup, key, value)
	log.Err(e, "updateHiddenFile", key, value)
}
