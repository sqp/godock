/*
Package maindock is a cairo-dock C wrapper to build a dock interface.

C files in the dir are the same as in the cairo-dock-core tree, or should be a
stripped version of them. They are supposed to be rewritten.

*/
package maindock

// Missing:
// _cairo_dock_successful_launch - Happy New Year message.
// extern gboolean g_bEasterEggs;
// crash tests and recovery. Not sure what to do about it.
//   static gint s_iNbCrashes = 0;
//   static gboolean s_bSucessfulLaunch = FALSE;
//   static GString *s_pLaunchCommand = NULL;

//

// #cgo pkg-config: gldi
// #include "cairo-dock-user-interaction.h"
/*

gboolean g_bLocked;   // TODO: To remove (1 more use in interaction)

*/
import "C"
import (
	"github.com/gotk3/gotk3/gtk"

	"github.com/sqp/godock/libs/cdglobal"     // Global consts.
	"github.com/sqp/godock/libs/cdtype"       // Logger type.
	"github.com/sqp/godock/libs/config"       // Config parser.
	"github.com/sqp/godock/libs/files"        // Files operations.
	"github.com/sqp/godock/libs/gldi"         // Gldi access.
	"github.com/sqp/godock/libs/gldi/dialog"  // Popup dialog.
	"github.com/sqp/godock/libs/gldi/globals" // Global variables.
	"github.com/sqp/godock/libs/ternary"      // Ternary operators.
	"github.com/sqp/godock/libs/text/tran"    // Translate.

	"github.com/sqp/godock/widgets/gtk/newgtk"

	"fmt"
	"os"
	"path/filepath"
	"time"
	"unsafe"
)

var log cdtype.Logger

// SetLogger provides a common logger for the Dbus service. It must be set to start a dock.
//
func SetLogger(l cdtype.Logger) {
	log = l
}

//
//----------------------------------------------------------------[ MAINDOCK ]--

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
	Exclude            string
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
	HttpPprof      bool
	AppletsDisable bool
	Debug          bool

	// auto loaded from original dock hidden file.
	isFirstLaunch  bool
	isNewVersion   bool
	sessionWasUsed bool
}

func (settings *DockSettings) Init() {
	confdir := cdglobal.ConfigDirDock(settings.UserDefinedDataDir)
	_, e := os.Stat(confdir)
	settings.isFirstLaunch = e != nil // TODO: need test is dir.

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

	//\___________________ internationalize the app.
	tran.Scend(cdglobal.CairoDockGettextPackage, cdglobal.CairoDockLocaleDir, "UTF-8")

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

	var rendering gldi.RenderingMethod
	switch {
	case settings.ForceOpenGL:
		rendering = gldi.RenderingOpenGL

	case settings.ForceCairo:
		rendering = gldi.RenderingCairo

	default:
		rendering = gldi.RenderingDefault
	}
	gldi.Init(int(rendering))

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

	env := DesktopEnvironment(settings.Env)
	if env != gldi.DesktopEnvUnknown {
		gldi.FMForceDesktopEnv(env)
	}

	if settings.IndirectOpenGL {
		gldi.GLBackendForceIndirectRendering()
	}

	if settings.ThemeServer == "" {
		settings.ThemeServer = cdglobal.DownloadServerURL
	}
	gldi.SetPaths(confdir, // will later be available as DirDockData  (g_cCairoDockDataDir)
		cdglobal.ConfigDirExtras,
		cdglobal.ConfigDirDockThemes,
		cdglobal.ConfigDirCurrentTheme,
		cdglobal.CairoDockShareThemesDir,
		cdglobal.DockThemeServerTag,
		settings.ThemeServer)

	//\___________________ Check that OpenGL is safely usable, if not ask the user what to do.

	// Unsafe OpenGL requires to be confirmed to use (need forced in conf or user validation).
	if settings.AskBackend || (gldi.GLBackendIsUsed() && !gldi.GLBackendIsSafe() && !settings.ForceOpenGL && !settings.IndirectOpenGL) {
		if settings.AskBackend || hidden.DefaultBackend == "" { // no backend defined.
			dialogAskBackend()
		} else if hidden.DefaultBackend != "opengl" { // disable opengl if unused, revert to cairo.
			gldi.GLBackendDeactivate()
		}
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

// Prepare is the last step before starting the dock, creating the config files.
//
func (settings DockSettings) Prepare() {

	// Register events.
	globals.ContainerObjectMgr.RegisterNotification(
		globals.NotifClickIcon,
		unsafe.Pointer(C.cairo_dock_notification_click_icon),
		globals.RunAfter)

	globals.ContainerObjectMgr.RegisterNotification(
		globals.NotifDropData,
		unsafe.Pointer(C.cairo_dock_notification_drop_data),
		globals.RunAfter)

	globals.ContainerObjectMgr.RegisterNotification(
		globals.NotifMiddleClickIcon,
		unsafe.Pointer(C.cairo_dock_notification_middle_click_icon),
		globals.RunAfter)

	globals.ContainerObjectMgr.RegisterNotification(
		globals.NotifScrollIcon,
		unsafe.Pointer(C.cairo_dock_notification_scroll_icon),
		globals.RunAfter)

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
	_, e := os.Stat(globals.ConfigFile())
	// if e == os.ErrNotExist {
	if e != nil {
		log.Info("creating configuration directory", globals.DirDockData())

		themeName := "Default-Single"

		if os.Getenv("DESKTOP_SESSION") == "cairo-dock" { // We're using the CD session for the first time
			themeName = "Default-Panel"
			settings.sessionWasUsed = true
			files.UpdateConfFile(globals.DirDockData(cdglobal.FileHiddenConfig), "Launch", "cd session", true)
		}

		files.CopyDir(globals.DirShareData(cdglobal.ConfigDirDockThemes, themeName), globals.CurrentThemePath())
	}

	// MISSING
	// The first time the Cairo-Dock session is used but not the first time the dock is launched: propose to use the Default-Panel theme
	// s_bCDSessionLaunched => settings.sessionWasUsed
}

func (settings *DockSettings) Start() {
	gldi.CurrentThemeLoad() // was moved before registration when I had some problems with refresh on start. Removed here for now.

	//\___________________ lock mode.

	// comme on ne pourra pas ouvrir le panneau de conf, ces 2 variables resteront tel quel.
	if settings.Locked {
		println("Cairo-Dock will be locked.") // was cd_warning (so it was set just before Verbosity). TODO: improve
		globals.DocksParam.SetLockIcons(true)
		globals.DocksParam.SetLockAll(true)
		globals.FullLock = true
		C.g_bLocked = C.gboolean(1) // forward for interaction
	}

	if !settings.SafeMode && gldi.ModulesGetNb() <= 1 { // 1 including Help.
		dialogNoPlugins()
	}

	if settings.isNewVersion { // update the version in the file.
		files.UpdateConfFile(globals.DirDockData(cdglobal.FileHiddenConfig), "Launch", "last version", globals.Version())

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

func Lock() {
	C.gtk_main()
}

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

// DesktopEnvironment converts the desktop environment backend type from the string.
//
func DesktopEnvironment(envstr string) gldi.DesktopEnvironment {
	env := gldi.DesktopEnvUnknown
	switch envstr {
	case "gnome":
		env = gldi.DesktopEnvGnome
	case "kde":
		env = gldi.DesktopEnvKDE
	case "xfce":
		env = gldi.DesktopEnvXFCE
	default:
		if envstr != "" {
			log.NewWarn(envstr, "Unknown desktop environment (valid options: gnome, kde, xfce)")
		}
	}
	return env
}

func firstLaunchSetup() {
	log.Info("firstLaunchScript", globals.DirShareData("scripts", "initial-setup.sh"))
	log.ExecShow(globals.DirShareData("scripts", "initial-setup.sh"))
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
		files.UpdateConfFile(globals.DirDockData(cdglobal.FileHiddenConfig), "Launch", "default backend", value)
	}
}

func dialogNoPlugins() {
	// Need to keep the string as it is for translation.
	str := "No plug-in were found.\nPlug-ins provide most of the functionalities (animations, applets, views, etc).\nSee http://glx-dock.org for more information.\nThere is almost no meaning in running the dock without them and it's probably due to a problem with the installation of these plug-ins.\nBut if you really want to use the dock without these plug-ins, you can launch the dock with the '-f' option to no longer have this message.\n"
	icon := gldi.IconsGetAnyWithoutDialog()
	container := globals.Maindock().ToContainer()
	iconpath := globals.FileCairoDockIcon()
	dialog.DialogShowTemporaryWithIcon(tran.Slate(str), icon, container, 0, iconpath)
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

	conf, e := config.NewFromFile(changelogPath)
	if e != nil {
		log.Debug("changelog not found", changelogPath, e.Error())
		return ""
	}

	major, minor, micro := globals.VersionSplit()

	// Get first line
	strver := fmt.Sprintf("%d.%d.%d", major, minor, micro) // version without "alpha", "beta", "rc", etc.
	msg, e := conf.String("ChangeLog", strver)
	if e != nil {
		log.Debug("changelog", "no info for version", globals.Version())
		return ""
	}
	msg = tran.Slate(msg)

	// Add all changelog lines for that version.
	i := 0
	for {
		strver := fmt.Sprintf("%d.%d.%d.%d", major, minor, micro, i)
		sub, e := conf.String("ChangeLog", strver)
		if e != nil {
			break
		}
		msg += "\n " + tran.Slate(sub)
		i++
	}

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

// loadHidden will try to load the hidden config data from the file.
//
func loadHidden(path string) *hiddenConfig {
	file := filepath.Join(path, cdglobal.FileHiddenConfig)
	conf, e := config.NewFromFile(file)

	if e != nil { // File missing, create it.
		conf = config.New()
		conf.AddSection("Launch")
		conf.AddOption("Launch", "last version", "")
		conf.AddOption("Launch", "default backend", "")
		conf.AddOption("Launch", "last year", "0")
		conf.AddSection("Gui")
		conf.AddOption("Gui", "mode", "0")

		log.Err(conf.WriteFile(file, 0644, ""), "create hidden conf", file)
		return &hiddenConfig{}
	}

	// Load data from file.
	hidden := &hiddenConfig{}
	conf.UnmarshalGroup(hidden, "Launch", config.GetTag)
	return hidden
}
