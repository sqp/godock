/*
Package maindock is a cairo-dock C wrapper to build a dock interface.


Files in the src dir are the same as in the cairo-dock-core tree. (or should be)

*/
package maindock

// #cgo pkg-config: gldi
// #include "maindock.h"
/*

#define CAIRO_DOCK_SHARE_THEMES_DIR "/usr/share/cairo-dock/themes"
#define CAIRO_DOCK_LOCALE_DIR       "/usr/share/locale"

*/
import "C"
import (
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/godock/libs/cdtype"       // Logger type.
	"github.com/sqp/godock/libs/config"       // Config parser.
	"github.com/sqp/godock/libs/gldi"         // Gldi access.
	"github.com/sqp/godock/libs/gldi/dialog"  // Popup dialog.
	"github.com/sqp/godock/libs/gldi/globals" // Global variables.
	"github.com/sqp/godock/libs/tran"         // Translate.

	"os"
	"os/user"
	"path/filepath"
	"time"
)

const (
	CAIRO_DOCK_DATA_DIR = "cairo-dock"

	// Nom du repertoire des themes extras.
	CAIRO_DOCK_EXTRAS_DIR = "extras"

	CAIRO_DOCK_THEMES_DIR = "themes"

	// Nom du repertoire racine du theme courant.
	CAIRO_DOCK_CURRENT_THEME_NAME = "current_theme"

	CAIRO_DOCK_DISTANT_THEMES_DIR = "themes3.4"

	CAIRO_DOCK_THEME_SERVER = "http://download.tuxfamily.org/glxdock/themes"

	CAIRO_DOCK_GETTEXT_PACKAGE = "cairo-dock"

	// Unused AFAIK.
	// CAIRO_DOCK_BACKUP_THEME_SERVER ="http://fabounet03.free.fr"

	HiddenFile = ".cairo-dock"
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

	isFirstLaunch bool
	isNewVersion  bool
}

func (settings *DockSettings) Init() {
	confdir := ConfigDir(settings.UserDefinedDataDir)
	_, e := os.Stat(confdir)
	settings.isFirstLaunch = e != nil // TODO: need test is dir.

	hidden := LoadHidden(confdir)

	settings.isNewVersion = hidden.LastVersion == globals.Version()

	// log.DETAIL(LoadHidden(confdir)) // missing year (don't know why)

	// MISSING
	// //\___________________ build the command line used to respawn, and check if we have been launched from another life.

	// mute all output messages if CD is not launched from a terminal
	// if (getenv("TERM") == NULL)  /// why not isatty(stdout) ?...
	// 	g_set_print_handler(PrintMuteFunc);

	gldi.DbusGThreadInit() // it's a wrapper: it will use dbus_threads_init_default ();

	gtk.Init(nil)

	//\___________________ internationalize the app.
	tran.Scend(CAIRO_DOCK_GETTEXT_PACKAGE, C.CAIRO_DOCK_LOCALE_DIR, "UTF-8")

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
		settings.ThemeServer = CAIRO_DOCK_THEME_SERVER
	}
	gldi.SetPaths(confdir, // will later be available as DirDockData  (g_cCairoDockDataDir)
		CAIRO_DOCK_EXTRAS_DIR,
		CAIRO_DOCK_THEMES_DIR,
		CAIRO_DOCK_CURRENT_THEME_NAME,
		C.CAIRO_DOCK_SHARE_THEMES_DIR,
		CAIRO_DOCK_DISTANT_THEMES_DIR,
		settings.ThemeServer)

	//\___________________ Check that OpenGL is safely usable, if not ask the user what to do.

	// Unsafe OpenGL requires to be confirmed to use (need forced in conf or user validation).
	if settings.AskBackend || (gldi.GLBackendIsUsed() && !gldi.GLBackendIsSafe() && !settings.ForceOpenGL && !settings.IndirectOpenGL) {
		if settings.AskBackend || hidden.DefaultBackend == "" { // no backend defined.
			dialogAskBackend()
		} else if hidden.DefaultBackend != "opengl" { // un backend par defaut qui n'est pas OpenGL.
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

func (settings DockSettings) Prepare() {

	//\___________________ handle crashes.
	// if (! bTesting)
	// 	_cairo_dock_set_signal_interception ();

	//\___________________ handle terminate signals to quit properly (especially when the system shuts down).
	// signal (SIGTERM, _cairo_dock_quit);  // Term // kill -15 (system)
	// signal (SIGHUP,  _cairo_dock_quit);  // sent to a process when its controlling terminal is closed

	// MISSING
	//\___________________ Disable modules that have crashed
	//\___________________ maintenance mode -> show the main config panel.
	//\___________________ load the current theme. (create if missing)
	// The first time the Cairo-Dock session is used but not the first time the dock is launched: propose to use the Default-Panel theme

}

func (settings *DockSettings) Start() {

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

	if settings.isNewVersion {
		/// If any operation must be done on the user theme (like activating a module by default, or disabling an option),
		/// it should be done here once (when CAIRO_DOCK_VERSION matches the new version).
	}

	// MISSING end of
	//\___________________ display the changelog in case of a new version.
}

func (settings *DockSettings) Clean() {

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
//------------------------------------------------------------------[ EVENTS ]--

func RegisterEvents() {
	C.register_events()
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

// ConfigDir returns a full path to the dock theme, according to the given user option.
//
func ConfigDir(dir string) string {
	if len(dir) > 0 {
		switch dir[0] {
		case '/':
			return dir // Full path, used as is.

		case '~':
			usr, e := user.Current()
			if e == nil {
				return usr.HomeDir + dir[1:] // Relative path to the homedir.
			}

		default:
			current, e := os.Getwd()
			if e == nil {
				return filepath.Join(current, dir) // Relative path to the current dir.
			}
		}
	}

	usr, e := user.Current()
	if e == nil {
		return filepath.Join(usr.HomeDir, ".config", CAIRO_DOCK_DATA_DIR) // Default theme path in .config.
	}

	return ""
}

//
//-----------------------------------------------------------------[ DIALOGS ]--

func dialogAskBackend() {
	dialog, _ := gtk.DialogNew()
	dialog.SetTitle("Use OpenGL in Cairo-Dock")
	dialog.AddButton("Yes", gtk.RESPONSE_YES)
	dialog.AddButton("No", gtk.RESPONSE_NO)

	labelTxt, _ := gtk.LabelNew(
		`OpenGL allows you to use the hardware acceleration, reducing the CPU load to the minimum.
It also allows some pretty visual effects similar to Compiz.
However, some cards and/or their drivers don't fully support it, which may prevent the dock from running correctly.
Do you want to activate OpenGL ?
 (To not show this dialog, launch the dock from the Application menu,
   or with the -o option to force OpenGL and -c to force cairo.)`)

	content, _ := dialog.GetContentArea()
	content.PackStart(labelTxt, false, false, 0)

	askBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 3)
	content.PackStart(askBox, false, false, 0)

	labelSave, _ := gtk.LabelNew("Remember this choice")
	check, _ := gtk.CheckButtonNew()
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
		log.Info("answer not saved yet")
		// 	s_cDefaulBackend = g_strdup (iAnswer == GTK_RESPONSE_NO ? "cairo" : "opengl");
		// 	gchar *cConfFilePath = g_strdup_printf ("%s/.cairo-dock", g_cCairoDockDataDir);
		// 	cairo_dock_update_conf_file (cConfFilePath,
		// 		G_TYPE_STRING, "Launch", "default backend", s_cDefaulBackend,
		// 		G_TYPE_INVALID);
		// 	g_free (cConfFilePath);
	}
}

func dialogNoPlugins() {
	str := `No plug-in were found.

Plug-ins provide most of the functionalities (animations, applets, views, etc).
See http://glx-dock.org for more information.
There is almost no meaning in running the dock without them and it's probably due to a problem with the installation of these plug-ins.
But if you really want to use the dock without these plug-ins, you can launch the dock with the '-f' option to no longer have this message.
`

	icon := gldi.IconsGetAnyWithoutDialog()
	container := globals.Maindock().ToContainer()
	iconpath := globals.DirShareData(globals.CairoDockIcon)
	dialog.DialogShowTemporaryWithIcon(str, icon, container, 0, iconpath)
}

//
//---------------------------------------------------------[ HIDDEN SETTINGS ]--

type HiddenSettings struct {
	LastVersion    string `conf:"last version"`
	DefaultBackend string `conf:"default backend"`
	// TestComposite string
	LastYear int `conf:"last year"`
	// UseSession bool `conf:"cd session"`
}

// LoadHidden will try to load the hidden config data from the file.
//
func LoadHidden(path string) *HiddenSettings {
	file := filepath.Join(path, HiddenFile)

	conf, e := config.NewFromFile(file) // Special conf reflector around the config file parser.
	if log.Err(e, "load hidden config file") {
		// TODO: need to create the file.
		return nil
	}

	hidden := &HiddenSettings{}
	conf.UnmarshalGroup(hidden, "Launch", config.GetTag)

	// TODO: update file with new version
	// if (bNewVersion)
	// {
	// 	gchar *cConfFilePath = g_strdup_printf ("%s/.cairo-dock", g_cCairoDockDataDir);
	// 	cairo_dock_update_conf_file (cConfFilePath,
	// 		G_TYPE_STRING, "Launch", "last version", CAIRO_DOCK_VERSION,
	// 		G_TYPE_INVALID);
	// 	g_free (cConfFilePath);

	return hidden
}
