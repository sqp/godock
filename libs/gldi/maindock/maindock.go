/*
Package maindock is a cairo-dock C wrapper to build a dock interface.


Files in the src dir are the same as in the cairo-dock-core tree. (or should be)

*/
package maindock

// #cgo pkg-config: gldi
// #include "maindock.h"
/*

// #define CAIRO_DOCK_SHARE_DATA_DIR   "/usr/share/cairo-dock"
#define CAIRO_DOCK_SHARE_THEMES_DIR "/usr/share/cairo-dock/themes"
#define CAIRO_DOCK_LOCALE_DIR       "/usr/share/locale"

#define CAIRO_DOCK_ICON "cairo-dock.svg"
#define CAIRO_DOCK_LOGO "cairo-dock-logo.png"

*/
import "C"
import (
	"github.com/conformal/gotk3/gtk"
	"github.com/gosexy/gettext"

	"github.com/sqp/godock/libs/config"
	"github.com/sqp/godock/libs/gldi"
	"github.com/sqp/godock/libs/gldi/globals"
	// "github.com/sqp/godock/libs/gldi/maindock/views"
	"github.com/sqp/godock/libs/log"
	// "github.com/sqp/godock/libs/maindock/gui"

	"os"
	"os/user"
	"path/filepath"
	"time"
	"unsafe"
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

func DockMore(settings DockSettings) {
	settings.Init()
	C.register_gui()
	C.register_events()

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

	// views.RegisterPanel("spanel")

	gldi.LoadCurrentTheme()
	settings.Start()
}

func (settings *DockSettings) Init() {
	confdir := ConfigDir(settings.UserDefinedDataDir)
	_, e := os.Stat(confdir)
	settings.isFirstLaunch = e != nil // TODO: need test is dir.

	// C._cairo_dock_get_global_config((*C.gchar)(C.CString(confdir)))

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
	gettext.BindTextdomain(CAIRO_DOCK_GETTEXT_PACKAGE, C.CAIRO_DOCK_LOCALE_DIR)
	gettext.BindTextdomainCodeset(CAIRO_DOCK_GETTEXT_PACKAGE, "UTF-8")
	gettext.Textdomain(CAIRO_DOCK_GETTEXT_PACKAGE)

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

	log.Info("Cairo-Dock version ", globals.Version())
	// log.Info("Compiled date      ", C.__DATE__, C.__TIME__)
	log.Info("Built with GTK     ", C.GTK_MAJOR_VERSION, C.GTK_MINOR_VERSION)
	log.Info("Running with OpenGL", gldi.GLBackendIsUsed())

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

func (settings *DockSettings) Start() {

	//\___________________ lock mode.

	// comme on ne pourra pas ouvrir le panneau de conf, ces 2 variables resteront tel quel.
	if settings.Locked {
		println("Cairo-Dock will be locked.") // was cd_warning (so it was set just before Verbosity). TODO: improve
		C.myDocksParam.bLockIcons = C.gboolean(1)
		C.myDocksParam.bLockAll = C.gboolean(1)

		C.g_bLocked = C.gboolean(1) // forward for the menu (to remove once menu is redone)
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

	// Start Mainloop
	defer settings.Clean()
	gtk.Main()
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
			log.Info("Unknown desktop environment", envstr, " (valid options: gnome, kde, xfce)")
		}
	}
	return env
}

// ConfigDir defines the current dock theme path, according to the given user option.
//
func ConfigDir(dir string) string {
	if len(dir) > 0 {
		switch dir[0] {
		case '/':
			return dir

		case '~':
			usr, e := user.Current()
			if e == nil {
				return usr.HomeDir + dir[1:]
			}

		default:
			current, e := os.Getwd()
			if e == nil {
				return filepath.Join(current, dir)
			}
		}
	}

	usr, e := user.Current()
	if e == nil {
		return filepath.Join(usr.HomeDir, ".config", CAIRO_DOCK_DATA_DIR)
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
	log.Info("answer", answer)

	if answer == int(gtk.RESPONSE_NO) {
		gldi.GLBackendDeactivate()
	}

	if remember { // save user choice to file.
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
	iconpath := globals.DirShareData() + "/" + C.CAIRO_DOCK_ICON
	gldi.DialogShowTemporaryWithIcon(str, icon, container, 0, iconpath)
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

// LoadHidden will try to load the own config data from the file.
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

//
//-----------------------------------------------------------[ GUI CALLBACKS ]--

type GuiInterface interface {
	ShowMainGui()                                                   //
	ShowModuleGui(appletName string)                                //
	ShowGui(*gldi.Icon, *gldi.Container, *gldi.ModuleInstance, int) //

	ShowAddons()
	ReloadItems()
	// ReloadCategoryWidget()
	Reload()
	Close()

	UpdateModulesList()
	UpdateModuleState(name string, active bool)
	UpdateModuleInstanceContainer(instance *gldi.ModuleInstance, detached bool)
	UpdateShortkeys()
	// UpdateDeskletParams(*gldi.Desklet)
	// UpdateDeskletVisibility(*gldi.Desklet)

	// CORE BACKEND
	SetStatusMessage(message string)
	ReloadCurrentWidget(moduleInstance *gldi.ModuleInstance, showPage int)
	ShowModuleInstanceGui(*gldi.ModuleInstance, int) //
	// GetWidgetFromName(moduleInstance *gldi.ModuleInstance, group string, key string)

	Window() *gtk.Window
}

var GuiInstance GuiInterface

//export ShowMainGui
func ShowMainGui() *C.GtkWidget {
	if GuiInstance == nil {
		return nil
	}
	GuiInstance.ShowMainGui()
	return toCWindow(GuiInstance.Window())
}

//export ShowModuleGui
func ShowModuleGui(cModuleName *C.gchar) *C.GtkWidget {
	name := C.GoString((*C.char)(cModuleName))
	C.g_free(C.gpointer(cModuleName))

	if GuiInstance == nil {
		return nil
	}
	GuiInstance.ShowModuleGui(name)
	return toCWindow(GuiInstance.Window())
}

//export ShowGui
func ShowGui(icon *C.Icon, container *C.GldiContainer, moduleInstance *C.GldiModuleInstance, iShowPage C.int) *C.GtkWidget {
	if GuiInstance == nil {
		return nil
	}
	i := gldi.NewIconFromNative(unsafe.Pointer(icon))
	c := gldi.NewContainerFromNative(unsafe.Pointer(container))
	m := gldi.NewModuleInstanceFromNative(unsafe.Pointer(moduleInstance))

	GuiInstance.ShowGui(i, c, m, int(iShowPage))
	return toCWindow(GuiInstance.Window())
}

//export ShowAddons
func ShowAddons() *C.GtkWidget {
	if GuiInstance == nil {
		return nil
	}
	GuiInstance.ShowAddons()
	return toCWindow(GuiInstance.Window())
}

//export ReloadItems
func ReloadItems() {
	if GuiInstance == nil {
		return
	}
	GuiInstance.ReloadItems()
}

//export Reload
func Reload() {
	if GuiInstance == nil {
		return
	}
	GuiInstance.Reload()
}

//export Close
func Close() {
	if GuiInstance == nil {
		return
	}
	GuiInstance.Close()
}

//export UpdateModulesList
func UpdateModulesList() {
	if GuiInstance == nil {
		return
	}
	GuiInstance.UpdateModulesList()
}

//export UpdateModuleState
func UpdateModuleState(cModuleName *C.gchar, active C.gboolean) {
	name := C.GoString((*C.char)(cModuleName))
	C.g_free(C.gpointer(cModuleName))

	if GuiInstance == nil {
		return
	}
	GuiInstance.UpdateModuleState(name, gobool(active))
}

//export UpdateModuleInstanceContainer
func UpdateModuleInstanceContainer(moduleInstance *C.GldiModuleInstance, detached C.gboolean) {
	if GuiInstance == nil {
		return
	}
	m := gldi.NewModuleInstanceFromNative(unsafe.Pointer(moduleInstance))
	GuiInstance.UpdateModuleInstanceContainer(m, gobool(detached))
}

//export UpdateShortkeys
func UpdateShortkeys() {
	if GuiInstance == nil {
		return
	}
	GuiInstance.UpdateShortkeys()
}

// CORE BACKEND

//export ShowModuleInstanceGui
func ShowModuleInstanceGui(moduleInstance *C.GldiModuleInstance, showPage C.int) {
	if GuiInstance == nil {
		return
	}
	m := gldi.NewModuleInstanceFromNative(unsafe.Pointer(moduleInstance))
	GuiInstance.ShowModuleInstanceGui(m, int(showPage))

}

//export SetStatusMessage
func SetStatusMessage(cmessage *C.gchar) {
	message := C.GoString((*C.char)(cmessage))
	C.g_free(C.gpointer(cmessage))

	if GuiInstance == nil {
		return
	}
	GuiInstance.SetStatusMessage(message)
}

//export ReloadCurrentWidget
func ReloadCurrentWidget(moduleInstance *C.GldiModuleInstance, showPage C.int) {
	if GuiInstance == nil {
		return
	}
	m := gldi.NewModuleInstanceFromNative(unsafe.Pointer(moduleInstance))
	GuiInstance.ReloadCurrentWidget(m, int(showPage))
}

// //export GetWidgetFromName
// func GetWidgetFromName(moduleInstance *C.GldiModuleInstance, cgroup *C.gchar, ckey *C.gchar) {
// 	group := C.GoString((*C.char)(cgroup))
// 	C.free(unsafe.Pointer((*C.char)(cgroup)))
// 	key := C.GoString((*C.char)(ckey))
// 	C.free(unsafe.Pointer((*C.char)(ckey)))

// 	if GuiInstance == nil {
// 		return
// 	}
// 	m := gldi.NewModuleInstanceFromNative(unsafe.Pointer(moduleInstance))
// 	GuiInstance.GetWidgetFromName(m, group, key)
// }

func toCWindow(win *gtk.Window) *C.GtkWidget {
	if win == nil {
		return nil
	}
	return (*C.GtkWidget)(unsafe.Pointer(win.Native()))
}

func gobool(b C.gboolean) bool {
	if b == 1 {
		return true
	}
	return false
}
