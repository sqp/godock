// Package gldi provides Go bindings for gldi (cairo-dock).  Supports version 3.4
package gldi

/*
#cgo pkg-config: glib-2.0 gldi
#include <stdlib.h>                              // free
#include <dbus/dbus-glib.h>                      // dbus_g_thread_init
#include <glib/gkeyfile.h>                       // GKeyFile
#include <glib.h>                       // GKeyFile


#include "cairo-dock-core.h"
#include "cairo-dock-animations.h"               // cairo_dock_trigger_icon_removal_from_dock
#include "cairo-dock-applications-manager.h"     // cairo_dock_get_appli_icon
#include "cairo-dock-applet-facility.h"          // cairo_dock_insert_icons_in_applet
#include "cairo-dock-backends-manager.h"         // cairo_dock_foreach_dock_renderer
#include "cairo-dock-config.h"                   // cairo_dock_load_current_theme
#include "cairo-dock-class-manager.h"            // cairo_dock_get_class_command
#include "cairo-dock-class-icon-manager.h"       // myClassIconObjectMgr
#include "cairo-dock-desklet-manager.h"          // myDeskletObjectMgr
#include "cairo-dock-desktop-manager.h"          // g_desktopGeometry
#include "cairo-dock-data-renderer.h"            // cairo_dock_render_new_data_on_icon
#include "cairo-dock-dock-factory.h"             // CairoDock
#include "cairo-dock-dock-facility.h"            // cairo_dock_get_available_docks
#include "cairo-dock-dock-manager.h"             // gldi_dock_get_readable_name
#include "cairo-dock-draw-opengl.h"              // cairo_dock_set_container_orientation_opengl
#include "cairo-dock-file-manager.h"             // CAIRO_DOCK_GNOME...
#include "cairo-dock-flying-container.h"         // myFlyingObjectMgr
#include "cairo-dock-gauge.h"                    // CairoGaugeAttribute
#include "cairo-dock-graph.h"                    // CairoGraphAttribute
#include "cairo-dock-icon-factory.h"             // Icon
#include "cairo-dock-icon-facility.h"        // Icon
#include "cairo-dock-keybinder.h"                // gldi_shortkey_new
#include "cairo-dock-keyfile-utilities.h"        // cairo_dock_conf_file_needs_update
#include "cairo-dock-launcher-manager.h"         // CAIRO_DOCK_ICON_TYPE_IS_LAUNCHER
#include "cairo-dock-log.h"                      // cd_log_set_level_from_name
#include "cairo-dock-overlay.h"                  // cairo_dock_add_overlay_from_image
#include "cairo-dock-menu.h"  // ModuleInstance
#include "cairo-dock-module-instance-manager.h"  // ModuleInstance
#include "cairo-dock-module-manager.h"           // gldi_modules_new_from_directory
#include "cairo-dock-object.h"                   // Icon
#include "cairo-dock-opengl.h"                   // gldi_gl_backend_force_indirect_rendering
#include "cairo-dock-progressbar.h"                    // CairoGraphAttribute
#include "cairo-dock-separator-manager.h"        // CAIRO_DOCK_ICON_TYPE_IS_SEPARATOR
#include "cairo-dock-struct.h"                   // CAIRO_DOCK_LAST_ORDER
#include "cairo-dock-stack-icon-manager.h"       // CAIRO_DOCK_ICON_TYPE_IS_CONTAINER
#include "cairo-dock-style-manager.h"            // gldi_style_colors_freeze
#include "cairo-dock-themes-manager.h"           // cairo_dock_set_paths
// #include "cairo-dock-utils.h"                    // cairo_dock_launch_command               TODO try remove;
#include "cairo-dock-windows-manager.h"          // gldi_window_can_minimize_maximize_close


extern CairoDock *g_pMainDock;

extern CairoDockGLConfig g_openglConfig;
extern gboolean          g_bUseOpenGL;

extern gchar *g_cCurrentLaunchersPath;

extern GldiDesktopGeometry g_desktopGeometry;

static int screen_position_x(int i)	{ return g_desktopGeometry.pScreens[i].x; }
static int screen_position_y(int i)	{ return g_desktopGeometry.pScreens[i].y; }
static int screen_width_i(int i) 	{ return g_desktopGeometry.pScreens[i].width; }
static int screen_height_i(int i) 	{ return g_desktopGeometry.pScreens[i].height; }

static gboolean IconIsSeparator    (Icon *icon) { return CAIRO_DOCK_ICON_TYPE_IS_SEPARATOR(icon); }
static gboolean IconIsSeparatorAuto(Icon *icon) { return CAIRO_DOCK_IS_AUTOMATIC_SEPARATOR(icon); }
static gboolean IconIsLauncher     (Icon *icon) { return CAIRO_DOCK_ICON_TYPE_IS_LAUNCHER(icon); }
static gboolean IconIsStackIcon    (Icon *icon) { return CAIRO_DOCK_ICON_TYPE_IS_CONTAINER(icon); }


static void emitSignalDropData(GldiContainer* container, gchar* data, Icon* icon, double order) {
	gldi_object_notify(container,NOTIFICATION_DROP_DATA, data, icon, order, container);
}


static void objectNotify(GldiContainer* container, int notif,  Icon* icon,  CairoDock* dock, GdkModifierType key) {
	gldi_object_notify(container, notif, icon, dock, key);
}


// from where this shit belongs.
static void manager_reload(GldiManager* manager, gboolean flag, GKeyFile* keyfile) {
 	GLDI_OBJECT(manager)->mgr->reload_object (GLDI_OBJECT(manager), flag, keyfile); // that's quite hacky, but we already have the keyfile, so...
 }



// from cairo-dock-icon-facility.h
static Icon* _icons_get_any_without_dialog() {
	return gldi_icons_get_without_dialog (g_pMainDock?g_pMainDock->icons:NULL);
}

// from cairo-dock-module-manager.h
static gboolean _module_is_auto_loaded(GldiModule *module) {
	return (module->pInterface->initModule == NULL || module->pInterface->stopModule == NULL || module->pVisitCard->cInternalModule != NULL);
}


static gpointer  intToPointer(int i)  { return GINT_TO_POINTER(i); }


static CairoDockRenderer*  newDockRenderer()  { return g_new0 (CairoDockRenderer, 1); }


static void render_new_data_on_icon (Icon *pIcon, double *pNewValues) {
	GldiContainer *pContainer = pIcon->pModuleInstance->pContainer;

	cairo_t *pDrawContext = cairo_create (pIcon->image.pSurface);
	cairo_dock_render_new_data_on_icon (pIcon, pContainer, pDrawContext, pNewValues);
	cairo_destroy (pDrawContext);
}



//-----------------------------------------------------------------[ C TO GO ]--

extern void onShortkey(gchar*, gpointer);


*/
import "C"

import (
	// "github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"github.com/sqp/godock/libs/cdtype" // Dock types.
	"github.com/sqp/godock/libs/gldi/desktopclass"
	"github.com/sqp/godock/libs/packages"
	"github.com/sqp/godock/libs/ternary"
	"github.com/sqp/godock/libs/text/tran"

	"github.com/sqp/godock/widgets/gtk/keyfile"
	"github.com/sqp/godock/widgets/gtk/togtk"

	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"unsafe"
)

// IconLastOrder defines the last icon position ??
//
const IconLastOrder = -1e9 // CAIRO_DOCK_LAST_ORDER

// RenderingMethod represents a screen display backend.
//
type RenderingMethod int

// Screen display backends.
const (
	RenderingOpenGL  RenderingMethod = C.GLDI_OPENGL
	RenderingCairo   RenderingMethod = C.GLDI_CAIRO
	RenderingDefault RenderingMethod = C.GLDI_DEFAULT
)

// DesktopEnvironment represents a desktop environment.
//
type DesktopEnvironment C.CairoDockDesktopEnv

// Desktop environment backends.
//
const (
	DesktopEnvGnome   DesktopEnvironment = C.CAIRO_DOCK_GNOME
	DesktopEnvKDE     DesktopEnvironment = C.CAIRO_DOCK_KDE
	DesktopEnvXFCE    DesktopEnvironment = C.CAIRO_DOCK_XFCE
	DesktopEnvUnknown DesktopEnvironment = C.CAIRO_DOCK_UNKNOWN_ENV
)

// ModuleCategory represents a module category.
//
type ModuleCategory C.GldiModuleCategory

// Module categories.
//
const (
	CategoryBehavior        ModuleCategory = C.CAIRO_DOCK_CATEGORY_BEHAVIOR
	CategoryTheme           ModuleCategory = C.CAIRO_DOCK_CATEGORY_THEME
	CategoryAppletFiles     ModuleCategory = C.CAIRO_DOCK_CATEGORY_APPLET_FILES
	CategoryAppletInternet  ModuleCategory = C.CAIRO_DOCK_CATEGORY_APPLET_INTERNET
	CategoryAppletDesktop   ModuleCategory = C.CAIRO_DOCK_CATEGORY_APPLET_DESKTOP
	CategoryAppletAccessory ModuleCategory = C.CAIRO_DOCK_CATEGORY_APPLET_ACCESSORY
	CategoryAppletSystem    ModuleCategory = C.CAIRO_DOCK_CATEGORY_APPLET_SYSTEM
	CategoryAppletFun       ModuleCategory = C.CAIRO_DOCK_CATEGORY_APPLET_FUN
)

//
//-----------------------------------------------------------[ DOCK BUILDING ]--

// Init initialises the gldi backend with the rendering method provided.
//
func Init(rendering int) {
	C.gldi_init(C.GldiRenderingMethod(rendering))
}

// SetPaths sets the default paths for the gldi backend.
//
func SetPaths(dataDir string, extra, themes, current, shareTheme, distantTheme, serverTheme string) {
	DirExtra := filepath.Join(dataDir, extra)
	DirThemes := filepath.Join(dataDir, themes)
	DirCurrent := filepath.Join(dataDir, current)
	C.cairo_dock_set_paths(gchar(dataDir), gchar(DirExtra), gchar(DirThemes), gchar(DirCurrent),
		gchar(shareTheme), gchar(distantTheme), gchar(serverTheme))
}

// ModulesNewFromDirectory loads internal modules from the given directory.
// Use with an empty dir to load from the default directory.
//
func ModulesNewFromDirectory(dir string) error {
	var cerr *C.GError
	var cstr *C.gchar
	if dir != "" {
		cstr = (*C.gchar)(C.CString(dir))
		defer C.free(unsafe.Pointer((*C.char)(cstr)))
	}
	C.gldi_modules_new_from_directory(cstr, &cerr) // load gldi-based plug-ins
	if cerr != nil {
		defer C.g_error_free(cerr)
		return errors.New(C.GoString((*C.char)(cerr.message)))
	}
	return nil
}

// ModulesGetNb returns the number of internal modules defined.
//
func ModulesGetNb() int {
	return int(C.gldi_module_get_nb())
}

// CurrentThemeLoad loads the theme in the dock.
//
func CurrentThemeLoad() {
	C.cairo_dock_load_current_theme()
}

// CurrentThemeNeedSave loads the theme in the dock.
//
func CurrentThemeNeedSave() bool {
	return gobool(C.cairo_dock_current_theme_need_save())
}

// CurrentThemeImport loads the theme in the dock.
//
func CurrentThemeImport(themeName string, useBehaviour, useLaunchers bool) error {
	if themeName == "" {
		return errors.New("theme name empty")
	}
	cTheme := (*C.gchar)(C.CString(themeName))
	defer C.free(unsafe.Pointer((*C.char)(cTheme)))

	c := C.cairo_dock_import_theme(cTheme, cbool(useBehaviour), cbool(useLaunchers))
	if !gobool(c) {
		return errors.New("import theme failed")
	}
	return nil
}

// CurrentThemeExport loads the theme in the dock.
//
func CurrentThemeExport(themeName string, saveBehaviour, saveLaunchers bool) error {
	if themeName == "" {
		return errors.New("theme name empty")
	}
	cstr := (*C.gchar)(C.CString(themeName))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))

	c := C.cairo_dock_export_current_theme(cstr, cbool(saveBehaviour), cbool(saveLaunchers))
	if !gobool(c) {
		return errors.New("export theme failed")
	}
	return nil
}

// CurrentThemePackage creates a package with the current theme of the dock.
//
func CurrentThemePackage(themeName, dirPath string) error {
	if themeName == "" {
		return errors.New("theme name empty")
	}
	cTheme := (*C.gchar)(C.CString(themeName))
	defer C.free(unsafe.Pointer((*C.char)(cTheme)))
	cDirPath := (*C.gchar)(C.CString(dirPath))
	defer C.free(unsafe.Pointer((*C.char)(cDirPath)))
	c := C.cairo_dock_package_current_theme(cTheme, cDirPath)
	if !gobool(c) {
		return errors.New("package theme failed")
	}
	return nil
}

// FMForceDesktopEnv forces the dock to use the given desktop environment backend.
//
func FMForceDesktopEnv(env DesktopEnvironment) {
	C.cairo_dock_fm_force_desktop_env(C.CairoDockDesktopEnv(env))
}

// ForceDocksAbove forces all docks to appear on top of other windows.
//
func ForceDocksAbove() {
	C.cairo_dock_force_docks_above()
}

// SetContainersNonSticky sets the non sticky mode for containers.
//
func SetContainersNonSticky() {
	C.cairo_dock_set_containers_non_sticky()
}

// DisableContainersOpacity disables the opacity ability on containers.
//
func DisableContainersOpacity() {
	C.cairo_dock_disable_containers_opacity()
}

// GLBackendForceIndirectRendering forces the indirect rendering on the OpenGL backend.
//
func GLBackendForceIndirectRendering() {
	C.gldi_gl_backend_force_indirect_rendering()
}

// GLBackendDeactivate prevents the dock from activating the OpenGL backend.
//
func GLBackendDeactivate() {
	C.gldi_gl_backend_deactivate()
}

// GLBackendIsSafe returns whether the OpenGL backend is safely usable or not.
//
func GLBackendIsSafe() bool {
	return GLBackendIsUsed() &&
		!gobool(C.g_openglConfig.bIndirectRendering) &&
		gobool(C.g_openglConfig.bAlphaAvailable) &&
		gobool(C.g_openglConfig.bStencilBufferAvailable)
}

// GLBackendIsUsed returns whether the dock use OpenGL backend or not.
//
func GLBackendIsUsed() bool {
	return gobool(C.g_bUseOpenGL)
}

// CanSetOnWidgetLayer returns whether desklets can be set on the widget layer.
//
func CanSetOnWidgetLayer() bool {
	return gobool(C.gldi_desktop_can_set_on_widget_layer())
}

// LogSetLevelFromName sets the C dock debug level.
//
func LogSetLevelFromName(verbosity string) {
	cstr := (*C.gchar)(C.CString(verbosity))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	C.cd_log_set_level_from_name(cstr)
}

// LogForceUseColor sets the colored output for the C dock debug.
func LogForceUseColor() {
	C.cd_log_force_use_color()
}

// StyleColorsFreeze TODO FIND USAGE.
//
func StyleColorsFreeze() {
	C.gldi_style_colors_freeze()
}

// DbusGThreadInit inits C threads.
//
func DbusGThreadInit() {
	C.dbus_g_thread_init() // it's a wrapper: it will use dbus_threads_init_default ();
}

// FreeAll frees all C dock memory.
func FreeAll() {
	C.gldi_free_all()
}

// XMLCleanupParser TODO FIND USAGE.
//
func XMLCleanupParser() {
	C.xmlCleanupParser()
}

//
//-------------------------------------------------------------[ GLDI COMMON ]--

// DockIsLoading returns whether the dock is still loading or not.
//
func DockIsLoading() bool {
	return gobool(C.cairo_dock_is_loading())
}

// GetAllAvailableDocks returns a filtered list of docks.
//
//   pParentDock    if not nil: exclude this dock.
//   pSubDock       if not nil: exclude this dock and its children.
//
func GetAllAvailableDocks(parent, subdock *CairoDock) []*CairoDock {
	var cp, cs *C.CairoDock
	if parent != nil {
		cp = parent.Ptr
	}
	if subdock != nil {
		cs = subdock.Ptr
	}

	c := C.cairo_dock_get_available_docks(cp, cs)
	clist := glib.WrapList(uintptr(unsafe.Pointer(c)))
	defer clist.Free()
	return goListDocks(clist)
}

// DockGet returns the dock with the given name.
//
func DockGet(containerName string) *CairoDock {
	cstr := (*C.gchar)(C.CString(containerName))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	return NewDockFromNative(unsafe.Pointer(C.gldi_dock_get(cstr)))
}

// DockNew creates a new dock with the given name.
//
func DockNew(name string) *CairoDock {
	cstr := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	return NewDockFromNative(unsafe.Pointer(C.gldi_dock_new(cstr)))
}

// DockAddConfFile adds a config file for a new root dock.
// Does not create the dock (use gldi_dock_new for that).
// Returns the unique name for the new dock, to be passed to gldi_dock_new.
//
func DockAddConfFile() string {
	c := C.gldi_dock_add_conf_file()
	defer C.free(unsafe.Pointer((*C.char)(c)))
	return C.GoString((*C.char)(c))
}

// ConfFileNeedUpdate returns whether the config file needs update or not.
//
func ConfFileNeedUpdate(kf *keyfile.KeyFile, version string) bool {
	cKeyfile := (*C.GKeyFile)(unsafe.Pointer(kf.ToNative()))
	cstr := (*C.gchar)(C.CString(version))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	return gobool(C.cairo_dock_conf_file_needs_update(cKeyfile, cstr))
}

// unused, see files.UpdateConfFile
// func ConfFileUpgradeFull(kf *keyfile.KeyFile, current, original string, updateKeys bool) {
// 	cKeyfile := (*C.GKeyFile)(unsafe.Pointer(kf.ToNative()))
// 	cCurrent := (*C.gchar)(C.CString(current))
// 	defer C.free(unsafe.Pointer((*C.char)(cCurrent)))
// 	cOriginal := (*C.gchar)(C.CString(original))
// 	defer C.free(unsafe.Pointer((*C.char)(cOriginal)))
// 	C.cairo_dock_upgrade_conf_file_full(cCurrent, cKeyfile, cOriginal, cbool(updateKeys))
// }

// EmitSignalDropData emits the signal on the container.
//
func EmitSignalDropData(container *Container, data string, icon *Icon, order float64) {
	cstr := (*C.gchar)(C.CString(data))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	var iconPtr *C.Icon
	if icon != nil {
		iconPtr = icon.Ptr
	}
	C.emitSignalDropData(container.Ptr, cstr, iconPtr, C.double(order))
}

// QuickHideAllDocks hides all dock at once.
//
func QuickHideAllDocks() {
	C.cairo_dock_quick_hide_all_docks()
}

// LauncherAddNew adds a new launcher to the dock.
//
func LauncherAddNew(uri string, dock *CairoDock, order float64) *Icon {
	var cstr *C.gchar
	if uri != "" {
		cstr = (*C.gchar)(C.CString(uri))
		defer C.free(unsafe.Pointer((*C.char)(cstr)))
	}
	c := C.gldi_launcher_add_new(cstr, dock.Ptr, C.double(order))
	return NewIconFromNative(unsafe.Pointer(c))
}

// SeparatorIconAddNew adds a separator to the dock.
//
func SeparatorIconAddNew(dock *CairoDock, order float64) *Icon {
	c := C.gldi_separator_icon_add_new(dock.Ptr, C.double(order))
	return NewIconFromNative(unsafe.Pointer(c))
}

// StackIconAddNew adds a stack icon to the dock (subdock).
//
func StackIconAddNew(dock *CairoDock, order float64) *Icon {
	c := C.gldi_stack_icon_add_new(dock.Ptr, C.double(order))
	return NewIconFromNative(unsafe.Pointer(c))
}

// MenuAddSubMenu adds a sub-menu to a given menu.
//
func MenuAddSubMenu(menu *gtk.Menu, label, iconPath string) (*gtk.Menu, *gtk.MenuItem) {
	cstr := (*C.gchar)(C.CString(label))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	var cpath *C.gchar
	if iconPath != "" {
		cpath = (*C.gchar)(C.CString(iconPath))
		defer C.free(unsafe.Pointer((*C.char)(cpath)))
	}
	var cmenuitem *C.GtkWidget
	cmenu := (*C.GtkWidget)(C.gpointer(menu.Native()))
	c := C.gldi_menu_add_sub_menu_full(cmenu, cstr, cpath, &cmenuitem)

	if c == nil {
		return nil, nil
	}
	return togtk.Menu(unsafe.Pointer(c)), togtk.MenuItem(unsafe.Pointer(cmenuitem))
}

/** A convenient function to add an item to a given menu.
 * @param pMenu the menu
 * @param cLabel the label, or NULL
 * @param cImage the image path or name, or NULL
 * @param pFunction the callback
 * @param pData the data passed to the callback
 * @return the new menu-entry that has been added.
 */

// f must be
// a function with a signaure matching the callback signature for
// detailedSignal.  userData must either 0 or 1 elements which can
// be optionally passed to f.

// MenuAddItem adds an item to the menu.
//
func MenuAddItem(menu *gtk.Menu, label, iconPath string) *gtk.MenuItem {
	cstr := (*C.gchar)(C.CString(label))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	var cpath *C.gchar
	if iconPath != "" {
		cpath = (*C.gchar)(C.CString(iconPath))
		defer C.free(unsafe.Pointer((*C.char)(cpath)))
	}
	c := C.gldi_menu_add_item((*C.GtkWidget)(unsafe.Pointer(menu.Native())), cstr, cpath, nil, nil)
	if c == nil {
		return nil
	}
	return togtk.MenuItem(unsafe.Pointer(c))
}

// unused, see cdtype.Logger.Exec...
// func LaunchCommand(cmd string) {
// 	cstr := (*C.gchar)(C.CString(cmd))
// 	defer C.free(unsafe.Pointer((*C.char)(cstr)))
// 	C.cairo_dock_launch_command_full(cstr, nil)
// }

//
//-----------------------------------------------------------------[ IOBJECT ]--

// IObject represents a basic dock object.
//
type IObject interface {
	ToNative() unsafe.Pointer
}

// ObjectReload reloads the given object.
//
func ObjectReload(obj IObject) {
	C.gldi_object_reload((*C.GldiObject)(obj.ToNative()), C.gboolean(1))
}

// ObjectUnref unrefs the given object.
//
func ObjectUnref(obj IObject) {
	C.gldi_object_unref((*C.GldiObject)(obj.ToNative()))
}

// ObjectDelete deletes the given object.
//
func ObjectDelete(obj IObject) {
	C.gldi_object_delete((*C.GldiObject)(obj.ToNative()))
}

// ObjectIsManagerChild returns whether the given object is a manager child.
//
func ObjectIsManagerChild(obj IObject, ptr *C.GldiObjectManager) bool {
	return gobool(C.gldi_object_is_manager_child((*C.GldiObject)(obj.ToNative()), ptr))
}

// ObjectIsDock returns whether the given object is a dock or not.
//
func ObjectIsDock(obj IObject) bool {
	return ObjectIsManagerChild(obj, &C.myDockObjectMgr)
}

// ObjectNotify notifies the given object.
//
func ObjectNotify(container *Container, notif int, icon *Icon, dock *CairoDock, key gdk.ModifierType) {
	C.objectNotify(container.Ptr, C.int(notif), icon.Ptr, dock.Ptr, C.GdkModifierType(key))
}

//
//---------------------------------------------------------------[ CAIRODOCK ]--

// CairoDock defines a gldi dock.
//
type CairoDock struct {
	Ptr          *C.CairoDock
	RendererData interface{} // View rendering data.
}

// NewDockFromNative wraps a gldi dock from C pointer.
//
func NewDockFromNative(p unsafe.Pointer) *CairoDock {
	if p == nil {
		return nil
	}
	return &CairoDock{
		Ptr:          (*C.CairoDock)(p),
		RendererData: nil,
	}
}

// ToNative returns the pointer to the native object.
//
func (o *CairoDock) ToNative() unsafe.Pointer {
	return unsafe.Pointer(o.Ptr)
}

// ToContainer returns the dock as a Container object.
//
func (o *CairoDock) ToContainer() *Container {
	return NewContainerFromNative(unsafe.Pointer(o.Ptr))
}

// Icons returns the list of icons in the dock.
//
func (o *CairoDock) Icons() (list []*Icon) {
	clist := glib.WrapList(uintptr(unsafe.Pointer(o.Ptr.icons)))
	return goListIcons(clist)
}

// FirstDrawnElementLinear TODO FIND USAGE.
//
func (o *CairoDock) FirstDrawnElementLinear() []*Icon {
	c := C.cairo_dock_get_first_drawn_element_linear(o.Ptr.icons)
	clist := glib.WrapList(uintptr(unsafe.Pointer(c)))
	return goListIcons(clist)
}

// GetDockName returns the key name of the dock.
//
func (o *CairoDock) GetDockName() string {
	return C.GoString((*C.char)(o.Ptr.cDockName))
}

// GetReadableName returns the readable name of the dock.
//
func (o *CairoDock) GetReadableName() string {
	return C.GoString((*C.char)(C.gldi_dock_get_readable_name(o.Ptr)))
}

// GetRefCount gives the number of icons pointing on the dock.
// 0 means it is a root dock, >0 a sub-dock.
//
func (o *CairoDock) GetRefCount() int {
	return int(o.Ptr.iRefCount)
}

// IsMainDock returns whether the dock is a main dock or not.
//
func (o *CairoDock) IsMainDock() bool {
	return gobool(o.Ptr.bIsMainDock)
}

// IsAutoHide returns whether the dock was auto hidden (TODO ensure).
//
func (o *CairoDock) IsAutoHide() bool {
	return gobool(o.Ptr.bAutoHide)
}

// SearchIconPointingOnDock TODO FIND USAGE AND COMPLETE.
//
func (o *CairoDock) SearchIconPointingOnDock(unknown interface{}) *Icon { // TODO: add param CairoDock **pParentDock
	c := C.cairo_dock_search_icon_pointing_on_dock(o.Ptr, nil)
	return NewIconFromNative(unsafe.Pointer(c))
}

// GetPointedIcon returns the pointed icon (mouse on the icon).
//
func (o *CairoDock) GetPointedIcon() *Icon {
	c := C.cairo_dock_get_pointed_icon(o.Ptr.icons)
	return NewIconFromNative(unsafe.Pointer(c))
}

// Container returns the inside container ?? TODO ENSURE.
//
func (o *CairoDock) Container() *Container {
	return NewContainerFromNative(unsafe.Pointer(&o.Ptr.container))
}

// GetNextIcon returns the icon after the given one in the dock.
//
func (o *CairoDock) GetNextIcon(icon *Icon) *Icon {
	c := C.cairo_dock_get_next_icon(o.Ptr.icons, icon.Ptr)
	return NewIconFromNative(unsafe.Pointer(c))
}

// GetPreviousIcon returns the icon before the given one in the dock.
//
func (o *CairoDock) GetPreviousIcon(icon *Icon) *Icon {
	c := C.cairo_dock_get_previous_icon(o.Ptr.icons, icon.Ptr)
	return NewIconFromNative(unsafe.Pointer(c))
}

// ScreenNumber TODO FIND USAGE.
//
func (o *CairoDock) ScreenNumber() int {
	return int(o.Ptr.iNumScreen)
}

// ScreenWidth returns the dock allocated screen width.
//
// cairo_dock_get_max_authorized_dock_width
func (o *CairoDock) ScreenWidth() int {
	geo := GetDesktopGeometry()
	i := o.ScreenNumber()
	isHoriz := o.Container().IsHorizontal()

	switch {
	case isHoriz && 0 <= i && i < geo.NbScreens():
		return geo.WidthScreen(i)

	case isHoriz:
		return geo.WidthAll()

	case 0 <= i && i < geo.NbScreens():
		return geo.HeightScreen(i)
	}

	return geo.HeightAll()
}

// MaxIconHeight returns the max icon height.
//
func (o *CairoDock) MaxIconHeight() int {
	return int(o.Ptr.iMaxIconHeight)
}

// Align returns the dock alignment (left/right or top/bottom).
//
func (o *CairoDock) Align() float64 {
	return float64(o.Ptr.fAlign)
}

// SetDecorationsHeight sets decorations height.
//
func (o *CairoDock) SetDecorationsHeight(i int) {
	o.Ptr.iDecorationsHeight = C.gint(i)
}

// MinDockWidth returns the min dock width.
//
func (o *CairoDock) MinDockWidth() int {
	return int(o.Ptr.iMinDockWidth)
}

// MaxDockWidth returns the max dock width.
//
func (o *CairoDock) MaxDockWidth() int {
	return int(o.Ptr.iMaxDockWidth)
}

// DecorationsHeight returns the dock decorations height.
//
func (o *CairoDock) DecorationsHeight() int {
	return int(o.Ptr.iDecorationsHeight)
}

// SetMinDockWidth sets the min dock width.
//
func (o *CairoDock) SetMinDockWidth(val int) {
	o.Ptr.iMinDockWidth = C.gint(val)
}

// SetMaxDockWidth sets the max dock width.
//
func (o *CairoDock) SetMaxDockWidth(val int) {
	o.Ptr.iMaxDockWidth = C.gint(val)
}

// SetActiveWidth sets the dock active width.
//
func (o *CairoDock) SetActiveWidth(val int) {
	o.Ptr.iActiveWidth = C.gint(val)
}

// SetDecorationsWidth sets decorations width.
//
func (o *CairoDock) SetDecorationsWidth(val int) {
	o.Ptr.iDecorationsWidth = C.gint(val)
}

// MinDockHeight returns the min dock height.
//
func (o *CairoDock) MinDockHeight() int {
	return int(o.Ptr.iMinDockHeight)
}

// MaxDockHeight returns the max dock height.
//
func (o *CairoDock) MaxDockHeight() int {
	return int(o.Ptr.iMaxDockHeight)
}

// SetMinDockHeight sets the dock min height.
//
func (o *CairoDock) SetMinDockHeight(val int) {
	o.Ptr.iMinDockHeight = C.gint(val)
}

// SetActiveHeight sets the dock active height.
//
func (o *CairoDock) SetActiveHeight(val int) {
	o.Ptr.iActiveHeight = C.gint(val)
}

// SetMaxDockHeight sets the dock max height.
//
func (o *CairoDock) SetMaxDockHeight(val int) {
	o.Ptr.iMaxDockHeight = C.gint(val)
}

// SetFlatDockWidth sets the dock flat width.
//
func (o *CairoDock) SetFlatDockWidth(val float64) {
	o.Ptr.fFlatDockWidth = C.gdouble(val)
}

// SetMagnitudeMax sets the max dock magnitude.
//
func (o *CairoDock) SetMagnitudeMax(val float64) {
	o.Ptr.fMagnitudeMax = C.gdouble(val)
}

// GlobalIconSize returns the global icon size.
//
func (o *CairoDock) GlobalIconSize() bool {
	return gobool(o.Ptr.bGlobalIconSize)
}

// IconSize returns the dock icon size.
//
func (o *CairoDock) IconSize() int {
	return int(o.Ptr.iIconSize)
}

// IsExtendedDock returns whether the dock is extended.
//
func (o *CairoDock) IsExtendedDock() bool {
	return gobool(o.Ptr.bExtendedMode) && o.GetRefCount() == 0
}

// HasShapeBitmap returns whether the dock has shape bitmap.
//
func (o *CairoDock) HasShapeBitmap() bool {
	return o.Ptr.pShapeBitmap != nil
}

// GetCurrentWidthLinear returns the dock current width linear.
//
func (o *CairoDock) GetCurrentWidthLinear() float64 {
	return float64(C.cairo_dock_get_current_dock_width_linear(o.Ptr))
}

// ShapeBitmapSubstractRectangleInt TODO FIND USAGE
func (o *CairoDock) ShapeBitmapSubstractRectangleInt(x, y, w, h int) {
	var rect C.cairo_rectangle_int_t
	rect.x = C.int(x)
	rect.y = C.int(y)
	rect.width = C.int(w)
	rect.height = C.int(h)
	C.cairo_region_subtract_rectangle(o.Ptr.pShapeBitmap, &rect)
}

// CheckIfMouseInsideLinear checks if the mouse is inside the linear area.
//
func (o *CairoDock) CheckIfMouseInsideLinear() {
	C.cairo_dock_check_if_mouse_inside_linear(o.Ptr)
}

// CheckCanDropLinear checks if we can drop linear. TODO CHECK USAGE.
func (o *CairoDock) CheckCanDropLinear() {
	C.cairo_dock_check_can_drop_linear(o.Ptr)
}

// CalculateMagnitude calculates the dock magnitude.
//
func (o *CairoDock) CalculateMagnitude() float64 {
	return float64(C.cairo_dock_calculate_magnitude(o.Ptr.iMagnitudeIndex))
}

// BackgroundBufferTexture TODO FIND USAGE.
//
func (o *CairoDock) BackgroundBufferTexture() uint32 {
	return uint32(o.Ptr.backgroundBuffer.iTexture)
}

//
//-------------------------------------------------------[ CONTAINER ]--

// Container defines a gldi container.
//
type Container struct {
	Ptr *C.GldiContainer
}

// NewContainerFromNative wraps a gldi container from C pointer.
//
func NewContainerFromNative(p unsafe.Pointer) *Container {
	return &Container{(*C.GldiContainer)(p)}
}

// ToNative returns the pointer to the native object.
//
func (o *Container) ToNative() unsafe.Pointer {
	return unsafe.Pointer(o.Ptr)
}

func (o *Container) ToCairoDock() *CairoDock {
	return NewDockFromNative(unsafe.Pointer(o.Ptr))
}

func (o *Container) ToDesklet() *Desklet {
	return NewDeskletFromNative(unsafe.Pointer(o.Ptr))
}

func (o *Container) Type() cdtype.ContainerType {
	switch {
	case ObjectIsDock(o):
		return cdtype.ContainerDock

	case o.IsDesklet():
		return cdtype.ContainerDesklet

	case o.IsDialog():
		return cdtype.ContainerDialog

	case o.IsFlyingContainer():
		return cdtype.ContainerFlying
	}

	return cdtype.ContainerUnknown
}

func (o *Container) IsDesklet() bool {
	return ObjectIsManagerChild(o, &C.myDeskletObjectMgr)
}

func (o *Container) IsDialog() bool {
	return ObjectIsManagerChild(o, &C.myDialogObjectMgr)
}

func (o *Container) IsFlyingContainer() bool {
	return ObjectIsManagerChild(o, &C.myFlyingObjectMgr)
}

func (o *Container) Width() int {
	return int(o.Ptr.iWidth)
}

func (o *Container) Height() int {
	return int(o.Ptr.iHeight)
}

func (o *Container) Ratio() float64 {
	return float64(o.Ptr.fRatio)
}

func (o *Container) MouseX() int {
	return int(o.Ptr.iMouseX)
}

func (o *Container) MouseY() int {
	return int(o.Ptr.iMouseY)
}

func (o *Container) WindowPositionX() int {
	return int(o.Ptr.iWindowPositionX)
}

func (o *Container) WindowPositionY() int {
	return int(o.Ptr.iWindowPositionY)
}

func (o *Container) IsHorizontal() bool {
	return gobool(C.gboolean(o.Ptr.bIsHorizontal))
}

func (o *Container) IsDirectionUp() bool {
	return gobool(C.gboolean(o.Ptr.bDirectionUp))
}

func (o *Container) ScreenBorder() cdtype.ContainerPosition {
	switch {
	case o.IsHorizontal() && o.IsDirectionUp():
		return cdtype.ContainerPositionBottom

	case o.IsHorizontal() && !o.IsDirectionUp():
		return cdtype.ContainerPositionTop

	case !o.IsHorizontal() && !o.IsDirectionUp():
		return cdtype.ContainerPositionLeft
	}

	return cdtype.ContainerPositionRight
}

//
//-----------------------------------------------------------------[ DESKLET ]--

// Desklet defines a gldi desklet.
//
type Desklet struct {
	Ptr *C.CairoDesklet
}

// NewDeskletFromNative wraps a gldi desklet from C pointer.
//
func NewDeskletFromNative(p unsafe.Pointer) *Desklet {
	if p == nil {
		return nil
	}
	return &Desklet{(*C.CairoDesklet)(p)}
}

// ToNative returns the pointer to the native object.
//
func (o *Desklet) ToNative() unsafe.Pointer {
	return unsafe.Pointer(o.Ptr)
}

// IsSticky returns whether the desklet is sticky.
//
func (o *Desklet) IsSticky() bool {
	return gobool(C.gldi_desklet_is_sticky(o.Ptr))
}

func (o *Desklet) PositionLocked() bool {
	return gobool(o.Ptr.bPositionLocked)
}

func (o *Desklet) GetIcon() *Icon {
	return NewIconFromNative(unsafe.Pointer(o.Ptr.pIcon))
}

func (o *Desklet) HasIcons() bool {
	return o.Ptr.icons != nil
}

func (o *Desklet) Icons() []*Icon {
	clist := glib.WrapList(uintptr(unsafe.Pointer(o.Ptr.icons)))
	return goListIcons(clist)
}

func (o *Desklet) RemoveIcons() {
	for _, ic := range o.Icons() {
		ObjectUnref(ic)
	}
	C.g_list_free(o.Ptr.icons)
	o.Ptr.icons = nil
}

func (o *Desklet) SetSticky(b bool) {
	println("SetSticky", b)
	C.gldi_desklet_set_sticky(o.Ptr, cbool(b))
}

func (o *Desklet) LockPosition(b bool) {
	C.gldi_desklet_lock_position(o.Ptr, cbool(b))
}

func (o *Desklet) Visibility() cdtype.DeskletVisibility {
	return cdtype.DeskletVisibility(o.Ptr.iVisibility)
}

// TRUE <=> save state in conf.
func (o *Desklet) SetVisibility(vis cdtype.DeskletVisibility, save bool) {
	C.gldi_desklet_set_accessibility(o.Ptr, C.CairoDeskletVisibility(vis), cbool(save))
}

func (o *Desklet) SetRendererByName(name string) {
	cstr := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	C.cairo_dock_set_desklet_renderer_by_name(o.Ptr, cstr, nil)
}

func (o *Desklet) SetRendererByNameData(name string, unk1, unk2 bool) {
	cstr := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))

	data := [2]C.gpointer{CIntPointer(int(cbool(unk1))), CIntPointer(int(cbool(unk2)))}
	C.cairo_dock_set_desklet_renderer_by_name(o.Ptr, cstr, C.CairoDeskletRendererConfigPtr(unsafe.Pointer(&data)))
}

//
//--------------------------------------------------------------------[ ICON ]--

// Icon defines a gldi icon.
//
type Icon struct {
	Ptr *C.Icon

	dataRendererText DataRendererText // optional data renderer (when set, this will replace the C data rendering)
}

// NewIconFromNative wraps a gldi icon from C pointer.
//
func NewIconFromNative(p unsafe.Pointer) *Icon {
	if p == nil {
		return nil
	}
	return &Icon{
		Ptr:              (*C.Icon)(p),
		dataRendererText: nil,
	}
}

func CreateDummyLauncher(name, iconPath, command, quickinfo string, order float64) *Icon {
	var qi *C.gchar
	if quickinfo != "" {
		qi = gchar(quickinfo)
	}

	c := C.cairo_dock_create_dummy_launcher(gchar(name), gchar(iconPath), gchar(command), qi, C.double(order))
	return NewIconFromNative(unsafe.Pointer(c))
}

func IconsGetAnyWithoutDialog() *Icon {
	return NewIconFromNative(unsafe.Pointer(C._icons_get_any_without_dialog()))
}

// ToNative returns the pointer to the native object.
//
func (o *Icon) ToNative() unsafe.Pointer {
	return unsafe.Pointer(o.Ptr)
}

// DefaultNameIcon returns improved name and image for the icon if possible.
//
func (icon *Icon) DefaultNameIcon() (name, img string) {
	switch {
	case icon.IsApplet():
		vc := icon.ModuleInstance().Module().VisitCard()
		return vc.GetTitle(), vc.GetIconFilePath()

	case icon.IsSeparator():
		return "--------", ""

	case icon.IsLauncher(), icon.IsStackIcon(), icon.IsAppli(), icon.IsClassIcon():
		name := icon.GetClass().Name()
		if name != "" {
			return name, icon.GetFileName() // icon.GetClassInfo(ClassIcon)
		}
		return ternary.String(icon.GetInitialName() != "", icon.GetInitialName(), icon.GetName()), icon.GetFileName()

	}
	return icon.GetName(), icon.GetFileName()
}

func (icon *Icon) GetClass() desktopclass.Info {
	return desktopclass.Info(C.GoString((*C.char)(icon.Ptr.cClass)))
}

func (icon *Icon) GetName() string {
	return C.GoString((*C.char)(icon.Ptr.cName))
}

func (icon *Icon) GetInitialName() string {
	return C.GoString((*C.char)(icon.Ptr.cInitialName))
}

func (icon *Icon) GetFileName() string {
	return C.GoString((*C.char)(icon.Ptr.cFileName))
}

func (icon *Icon) GetDesktopFileName() string {
	return C.GoString((*C.char)(icon.Ptr.cDesktopFileName))
}

func (icon *Icon) GetParentDockName() string {
	return C.GoString((*C.char)(icon.Ptr.cParentDockName))
}

func (icon *Icon) GetCommand() string {
	return C.GoString((*C.char)(icon.Ptr.cCommand))
}

func (icon *Icon) GetContainer() *Container {
	if icon.Ptr == nil || icon.Ptr.pContainer == nil {
		return nil
	}
	return NewContainerFromNative(unsafe.Pointer(icon.Ptr.pContainer))
}

// ConfigPath gives the full path to the icon config file.
//
func (icon *Icon) ConfigPath() string {
	switch {
	case icon.IsApplet():
		return icon.ModuleInstance().GetConfFilePath()

	case icon.IsStackIcon(), icon.IsLauncher() || icon.IsSeparator():
		dir := C.GoString((*C.char)(C.g_cCurrentLaunchersPath))
		// dir := globals.CurrentLaunchersPath()
		return filepath.Join(dir, icon.GetDesktopFileName())
	}
	return ""
}

func (icon *Icon) GetIgnoreQuickList() bool {
	return gobool(icon.Ptr.bIgnoreQuicklist)
}

func (o *Icon) X() float64 {
	return float64(o.Ptr.fX)
}

func (o *Icon) Y() float64 {
	return float64(o.Ptr.fY)
}

func (o *Icon) DrawX() float64 {
	return float64(o.Ptr.fDrawX)
}

func (o *Icon) DrawY() float64 {
	return float64(o.Ptr.fDrawY)
}

func (o *Icon) Width() float64 {
	return float64(o.Ptr.fWidth)
}

func (o *Icon) Height() float64 {
	return float64(o.Ptr.fHeight)
}

func (o *Icon) RequestedWidth() int {
	return int(o.Ptr.iRequestedWidth)
}

func (o *Icon) RequestedHeight() int {
	return int(o.Ptr.iRequestedHeight)
}

func (o *Icon) RequestedDisplayWidth() int {
	return int(o.Ptr.iRequestedDisplayWidth)
}

func (o *Icon) RequestedDisplayHeight() int {
	return int(o.Ptr.iRequestedDisplayHeight)
}

func (o *Icon) IsPointed() bool {
	return gobool(o.Ptr.bPointed)
}

func (o *Icon) SetPointed(val bool) {
	o.Ptr.bPointed = cbool(val)
}

func (o *Icon) SetWidth(val float64) {
	o.Ptr.fWidth = C.gdouble(val)
}

func (o *Icon) SetHeight(val float64) {
	o.Ptr.fHeight = C.gdouble(val)
}

func (o *Icon) SetScale(val float64) {
	o.Ptr.fScale = C.gdouble(val)
}

func (o *Icon) InsertRemoveFactor() float64 {
	return float64(o.Ptr.fInsertRemoveFactor)
}

func (o *Icon) Scale() float64 {
	return float64(o.Ptr.fScale)
}

func (o *Icon) XAtRest() float64 {
	return float64(o.Ptr.fXAtRest)
}

func (o *Icon) SetX(f float64) {
	o.Ptr.fX = C.gdouble(f)
}

func (o *Icon) SetY(f float64) {
	o.Ptr.fY = C.gdouble(f)
}

func (o *Icon) SetDrawX(f float64) {
	o.Ptr.fDrawX = C.gdouble(f)
}

func (o *Icon) SetDrawY(f float64) {
	o.Ptr.fDrawY = C.gdouble(f)
}

func (o *Icon) SetXAtRest(f float64) {
	o.Ptr.fXAtRest = C.gdouble(f)
}

func (o *Icon) SetAllocatedSize(w, h int) {
	o.Ptr.iAllocatedWidth = C.gint(w)
	o.Ptr.iAllocatedHeight = C.gint(h)
}

func (o *Icon) SetWidthFactor(f float64) {
	o.Ptr.fWidthFactor = C.gdouble(f)
}

func (o *Icon) SetHeightFactor(f float64) {
	o.Ptr.fHeightFactor = C.gdouble(f)
}

func (o *Icon) SetOrientation(f float64) {
	o.Ptr.fOrientation = C.gdouble(f)
}

func (o *Icon) SetAlpha(f float64) {
	o.Ptr.fAlpha = C.gdouble(f)
}

func (o *Icon) Order() float64 {
	return float64(o.Ptr.fOrder)
}

func (o *Icon) IconExtent() (int, int) {
	var width, height C.int
	C.cairo_dock_get_icon_extent(o.Ptr, &width, &height)
	return int(width), int(height)
}

// func (icon *Icon) GetIconType() int {
// 	return int(C.cairo_dock_get_icon_type(icon.Ptr))
// }

func (icon *Icon) IsApplet() bool {
	return icon.Ptr != nil && icon.Ptr.pModuleInstance != nil
}

// IsAppli returns whether the icon manages an application. CAIRO_DOCK_IS_APPLI
//
func (icon *Icon) IsAppli() bool {
	return icon.Ptr != nil && icon.Ptr.pAppli != nil
}

// IsClassIcon returns whether the icon .
// GLDI_OBJECT_IS_CLASS_ICON / CAIRO_DOCK_ICON_TYPE_IS_CLASS_CONTAINER
//
func (o *Icon) IsClassIcon() bool {
	return ObjectIsManagerChild(o, &C.myClassIconObjectMgr)
}

//
func (o *Icon) IsDetachableApplet() bool {
	return o.IsApplet() &&
		o.ModuleInstance().Module().VisitCard().GetContainerType()&C.CAIRO_DOCK_MODULE_CAN_DESKLET > 0
}

// IsMultiAppli returns whether the icon manages multiple applications. CAIRO_DOCK_IS_MULTI_APPLI
//
func (icon *Icon) IsMultiAppli() bool {
	return icon.Ptr.pSubDock != nil &&
		(icon.IsLauncher() ||
			icon.IsClassIcon() ||
			(icon.IsApplet() && icon.GetClass() != ""))
}

// IsTaskbar returns whether the icon belongs to the taskbar or not.
//
func (icon *Icon) IsTaskbar() bool {
	return icon.IsAppli() && !icon.IsLauncher() && !icon.IsApplet()
}

func (icon *Icon) IsSeparator() bool {
	return gobool(C.IconIsSeparator(icon.Ptr))
}
func (icon *Icon) IsSeparatorAuto() bool {
	return gobool(C.IconIsSeparatorAuto(icon.Ptr))
}

func (icon *Icon) IsLauncher() bool {
	return gobool(C.IconIsLauncher(icon.Ptr))
}

func (icon *Icon) IsStackIcon() bool { // CAIRO_DOCK_ICON_TYPE_IS_CONTAINER
	return gobool(C.IconIsStackIcon(icon.Ptr))
}

// cairo-dock-core/src/gldit/cairo-dock-icon-factory.h
// #define CAIRO_DOCK_IS_AUTOMATIC_SEPARATOR(icon) (CAIRO_DOCK_ICON_TYPE_IS_SEPARATOR (icon) && (icon)->cDesktopFileName == NULL)
// #define CAIRO_DOCK_IS_USER_SEPARATOR(icon) (CAIRO_DOCK_ICON_TYPE_IS_SEPARATOR (icon) && (icon)->cDesktopFileName != NULL)

func (icon *Icon) RemoveFromDock() {
	C.cairo_dock_trigger_icon_removal_from_dock(icon.Ptr)
}

func (icon *Icon) WriteContainerNameInConfFile(newdock string) {
	cstr := (*C.gchar)(C.CString(newdock))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	C.gldi_theme_icon_write_container_name_in_conf_file(icon.Ptr, cstr)
}

func (icon *Icon) ModuleInstance() *ModuleInstance {
	if !icon.IsApplet() {
		return nil
	}
	return NewModuleInstanceFromNative(unsafe.Pointer(icon.Ptr.pModuleInstance))
}

func (icon *Icon) GetSubDock() *CairoDock {
	return NewDockFromNative(unsafe.Pointer(icon.Ptr.pSubDock))
}

func (o *Icon) RemoveSubdockEmpty() {
	if o.Ptr.pSubDock != nil && o.Ptr.pSubDock.icons == nil {
		o.Ptr.pSubDock = nil
	}
}

func (o *Icon) RemoveDialogs() {
	C.gldi_dialogs_remove_on_icon(o.Ptr)
}

func (o *Icon) IsDemandingAttention() bool {
	return gobool(o.Ptr.bIsDemandingAttention)
}

func (o *Icon) RequestAttention(animation string, nbRounds int) {
	cstr := (*C.gchar)(C.CString(animation))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	C.gldi_icon_request_attention(o.Ptr, cstr, C.int(nbRounds))
}

func (o *Icon) StopAttention() {
	C.gldi_icon_stop_attention(o.Ptr)
}

func (icon *Icon) ClassIsInhibited() bool {
	return gobool(C.cairo_dock_class_is_inhibited(icon.Ptr.cClass))
}

func (o *Icon) InhibiteClass(class string) {
	cstr := (*C.gchar)(C.CString(class))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	C.cairo_dock_inhibite_class(cstr, o.Ptr)
}

func (o *Icon) DeinhibiteClass() {
	C.cairo_dock_deinhibite_class(o.Ptr.cClass, o.Ptr)
}

func (o *Icon) Window() *WindowActor {
	return NewWindowActorFromNative(unsafe.Pointer(o.Ptr.pAppli))
}

// GetInhibitor returns the icon that inhibits the current one (has registered the class).
//
func (icon *Icon) GetInhibitor(b bool) *Icon {
	c := C.cairo_dock_get_inhibitor(icon.Ptr, cbool(b))
	return NewIconFromNative(unsafe.Pointer(c))
	// 			pNewActiveIcon = cairo_dock_get_inhibitor (pNewActiveIcon, FALSE);
}

func (o *Icon) RemoveIconsFromSubdock(dest *CairoDock) {
	C.cairo_dock_remove_icons_from_dock(o.Ptr.pSubDock, dest.Ptr)
}

// TODO: may have to move.
func (icon *Icon) SubDockIcons() []*Icon {
	if icon.Ptr == nil || icon.Ptr.pSubDock == nil {
		return nil
	}
	clist := glib.WrapList(uintptr(unsafe.Pointer(icon.Ptr.pSubDock.icons)))
	return goListIcons(clist)
}

// SetLabel sets the label of an icon.
// If it has a sub-dock, it is renamed (the name is possibly altered to stay unique).
// The label buffer is updated too.
//
func (icon *Icon) SetLabel(str string) {
	var cstr *C.gchar
	if str != "" {
		cstr = (*C.gchar)(C.CString(str))
		defer C.free(unsafe.Pointer((*C.char)(cstr)))
	}
	C.gldi_icon_set_name(icon.Ptr, cstr)
}

// SetQuickInfo sets the quick-info of an icon.
// This is a small text (a few characters) that is superimposed on the icon.
//
func (icon *Icon) SetQuickInfo(str string) {
	var cstr *C.gchar
	if str != "" {
		cstr = (*C.gchar)(C.CString(str))
		defer C.free(unsafe.Pointer((*C.char)(cstr)))
	}
	C.gldi_icon_set_quick_info(icon.Ptr, cstr)
}

func (icon *Icon) SetIcon(str string) error {
	if icon.Ptr.image.pSurface == nil {
		return errors.New("icon has no image.pSurface")
	}

	var cstr *C.gchar
	if str != "" {
		cstr = (*C.gchar)(C.CString(str))
		defer C.free(unsafe.Pointer((*C.char)(cstr)))
	}
	ctx := C.cairo_create(icon.Ptr.image.pSurface)
	C.cairo_dock_set_image_on_icon(ctx, cstr, icon.Ptr, icon.GetContainer().Ptr) // returns gboolean
	C.cairo_destroy(ctx)

	return nil
}

func (icon *Icon) AddNewDataRenderer(attr DataRendererAttributer) {
	cAttr, free := attr.ToAttribute()
	C.cairo_dock_add_new_data_renderer_on_icon(icon.Ptr, icon.GetContainer().Ptr, cAttr)
	free()
}

func (icon *Icon) AddDataRendererWithText(attr DataRendererAttributer, dataRendererText DataRendererText) {
	cAttr, free := attr.ToAttribute()
	C.cairo_dock_add_new_data_renderer_on_icon(icon.Ptr, icon.GetContainer().Ptr, cAttr)
	icon.dataRendererText = dataRendererText
	free()
}

func (icon *Icon) Render(values ...float64) error {
	if icon.Ptr.image.pSurface == nil {
		return errors.New("Render: icon has no image.pSurface")
	}
	if icon.GetContainer() == nil {
		return errors.New("Render: icon has no container")
	}

	list := make([]C.double, len(values))
	for i, val := range values {
		list[i] = C.double(val)
	}

	C.render_new_data_on_icon(icon.Ptr, &list[0]) // points to the first element of the slice array (good enough for C).
	return nil
}

type DataRendererText func(...float64) []string

func (icon *Icon) DataRendererTextPercent(values ...float64) []string {
	list := make([]string, len(values))
	for i, val := range values {
		list[i] = fmt.Sprintf("%.1f%%%%", val*100)

		//snprintf (cFormatBuffer, iBufferLength, fValue < .0995 ? "%.1f%%" : (fValue < 1 ? " %.0f%%" : "%.0f%%"), fValue * 100.);

	}
	return list
}

func (icon *Icon) RemoveDataRenderer() {
	C.cairo_dock_remove_data_renderer_on_icon(icon.Ptr)
	icon.dataRendererText = nil
}

// AddOverlayFromImage adds an overlay on the icon.
//
func (icon *Icon) AddOverlayFromImage(iconPath string, position cdtype.EmblemPosition) {
	var cstr *C.gchar
	if iconPath != "" {
		cstr = (*C.gchar)(C.CString(iconPath))
		defer C.free(unsafe.Pointer((*C.char)(cstr)))
	}
	// last arg was 'myApplet' to identify the overlays set by the Dbus plug-in (since the plug-in can't be deactivated, 'myApplet' is constant).
	C.cairo_dock_add_overlay_from_image(icon.Ptr, cstr, C.CairoOverlayPosition(position), C.gpointer(icon.Ptr))
}

// RemoveOverlayAtPosition removes an overlay on the icon.
//
func (icon *Icon) RemoveOverlayAtPosition(position cdtype.EmblemPosition) {
	C.cairo_dock_remove_overlay_at_position(icon.Ptr, C.CairoOverlayPosition(position), C.gpointer(icon.Ptr))
}

func (icon *Icon) RequestAnimation(animation string, rounds int) {
	cstr := (*C.gchar)(C.CString(animation))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	C.gldi_icon_request_animation(icon.Ptr, cstr, C.int(rounds))
}

func (icon *Icon) Redraw() {
	C.cairo_dock_redraw_icon(icon.Ptr)
}

// MoveAfterIcon moves the icon position after the given icon.
//
func (icon *Icon) MoveAfterIcon(container *CairoDock, target *Icon) {
	C.cairo_dock_move_icon_after_icon(container.Ptr, icon.Ptr, target.Ptr)
}

// CallbackActionWindow returns a func to use as gtk callback.
// On event, it will test if the icon still has a valid window and launch the
// provided action on this window.
//
func (icon *Icon) CallbackActionWindow(call func(*WindowActor)) func() {
	return func() {
		if icon.IsAppli() {
			call(icon.Window())
		}
	}
}

// CallbackActionSubWindows is the same as CallbackActionWindow but launch the
// action on all subdock windows.
//
func (icon *Icon) CallbackActionSubWindows(call func(*WindowActor)) func() {
	return func() {
		for _, ic := range icon.SubDockIcons() {
			if ic.IsAppli() {
				call(ic.Window())
			}
		}
	}
}

func (icon *Icon) CallbackActionWindowToggle(call func(*WindowActor, bool), getvalue func(*WindowActor) bool) func() {
	return icon.CallbackActionWindow(func(win *WindowActor) {
		v := getvalue(win)
		call(win, !v)
	})
}

//
//--------------------------------------------------[ DATARENDERERATTRIBUTES ]--

type DataRendererAttributer interface {
	ToAttribute() (attr *C.CairoDataRendererAttribute, free func())
}

type DataRendererAttributeCommon struct {
	ModelName    string
	LatencyTime  int
	NbValues     int
	WriteValues  bool
	UpdateMinMax bool // for graph
	MemorySize   int  // for graph
	RotateTheme  int  // for gauge

	cType *C.gchar
}

func (o *DataRendererAttributeCommon) parseCommon(p unsafe.Pointer) *C.CairoDataRendererAttribute {
	attr := (*C.CairoDataRendererAttribute)(p)

	attr.cModelName = o.cType
	attr.iLatencyTime = C.gint(o.LatencyTime)
	attr.iNbValues = C.gint(o.NbValues)
	attr.iMemorySize = C.gint(o.MemorySize)
	attr.iRotateTheme = C.RendererRotateTheme(o.RotateTheme)
	attr.bUpdateMinMax = cbool(o.UpdateMinMax)
	attr.bWriteValues = cbool(o.WriteValues)
	return attr
}

func (o *DataRendererAttributeCommon) Free() {
	C.free(unsafe.Pointer((*C.char)(o.cType)))
}

type DataRendererAttributeProgressBar struct {
	DataRendererAttributeCommon
}

func NewDataRendererAttributeProgressBar() *DataRendererAttributeProgressBar {
	return &DataRendererAttributeProgressBar{
		DataRendererAttributeCommon: DataRendererAttributeCommon{
			cType: (*C.gchar)(C.CString("progressbar")),
		},
	}
}

func (o *DataRendererAttributeProgressBar) ToAttribute() (*C.CairoDataRendererAttribute, func()) {
	aGaugeAttr := new(C.CairoProgressBarAttribute)
	return o.parseCommon(unsafe.Pointer(aGaugeAttr)), o.Free
}

type DataRendererAttributeGauge struct {
	DataRendererAttributeCommon
	Theme string
}

func NewDataRendererAttributeGauge() *DataRendererAttributeGauge {
	return &DataRendererAttributeGauge{
		DataRendererAttributeCommon: DataRendererAttributeCommon{
			cType: (*C.gchar)(C.CString("gauge")),
		},
	}
}

func (o *DataRendererAttributeGauge) ToAttribute() (*C.CairoDataRendererAttribute, func()) {
	aGaugeAttr := new(C.CairoGaugeAttribute)
	cTheme := (*C.gchar)(C.CString(o.Theme))
	defer C.free(unsafe.Pointer((*C.char)(cTheme)))

	aGaugeAttr.cThemePath = C.cairo_dock_get_data_renderer_theme_path(o.cType, cTheme, C.CAIRO_DOCK_ANY_PACKAGE)
	free := func() {
		// C.free(unsafe.Pointer(aGaugeAttr.cThemePath)) // don't free this one ?
		o.Free() // free internal type text.
	}
	return o.parseCommon(unsafe.Pointer(aGaugeAttr)), free
}

type DataRendererAttributeGraph struct {
	DataRendererAttributeCommon
	Type            cdtype.RendererGraphType
	MixGraphs       bool
	HighColor       []float64
	LowColor        []float64
	BackGroundColor [4]float64
}

func NewDataRendererAttributeGraph() *DataRendererAttributeGraph {
	return &DataRendererAttributeGraph{
		DataRendererAttributeCommon: DataRendererAttributeCommon{
			cType: (*C.gchar)(C.CString("graph")),
		},
	}
}

func (o *DataRendererAttributeGraph) ToAttribute() (*C.CairoDataRendererAttribute, func()) {
	attr := new(C.CairoGraphAttribute)

	attr.iType = C.CairoDockTypeGraph(int(o.Type))
	attr.bMixGraphs = cbool(o.MixGraphs)
	attr.fHighColor = cListGdouble(o.HighColor)
	attr.fLowColor = cListGdouble(o.LowColor)

	for i := 0; i < 4; i++ { // copy background colors.
		attr.fBackGroundColor[i] = C.gdouble(o.BackGroundColor[i])
	}
	free := func() {
		C.free(unsafe.Pointer(attr.fHighColor))
		C.free(unsafe.Pointer(attr.fLowColor))
		o.Free() // free internal type text.
	}
	return o.parseCommon(unsafe.Pointer(attr)), free
}

//
//------------------------------------------------------------------[ MODULE ]--

type Module struct {
	Ptr *C.GldiModule
}

func ModuleGet(name string) *Module {
	cstr := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	c := C.gldi_module_get(cstr)
	return NewModuleFromNative(unsafe.Pointer(c))
}

func NewModuleFromNative(p unsafe.Pointer) *Module {
	if p == nil {
		return nil
	}
	return &Module{(*C.GldiModule)(p)}
}

// func NewModule(vc *VisitCard) (*Module, *C.GldiModuleInterface) {
// 	interf := new(C.GldiModuleInterface)
// 	c := C.gldi_module_new(vc.Ptr, interf)
// 	return NewModuleFromNative(unsafe.Pointer(c)), interf
// }

func (m *Module) ToNative() unsafe.Pointer {
	return unsafe.Pointer(m.Ptr)
}

func (m *Module) IsAutoLoaded() bool {
	return gobool(C._module_is_auto_loaded(m.Ptr))
}

func (m *Module) VisitCard() *VisitCard {
	return NewVisitCardFromNative(unsafe.Pointer(m.Ptr.pVisitCard))
}

func (m *Module) InstancesList() (list []*ModuleInstance) {
	clist := glib.WrapList(uintptr(unsafe.Pointer(m.Ptr.pInstancesList)))
	return goListModuleInstance(clist)
}

func (m *Module) Activate() {
	C.gldi_module_activate(m.Ptr)
}

func (m *Module) Deactivate() {
	C.gldi_module_deactivate(m.Ptr)
}

func (m *Module) AddInstance() {
	C.gldi_module_add_instance(m.Ptr)
}

//
//----------------------------------------------------------[ MODULEINSTANCE ]--

var (
	shortkeyMapStatic = make(map[int]func(string))
	shortkeyMapMutex  = &sync.Mutex{}
)

type ModuleInstance struct {
	Ptr *C.GldiModuleInstance

	shortkeys map[string]int // index = group+key, value = shortkeyMapStatic reference.
}

// /// container of the icon.
// GldiContainer *pContainer;
// /// this field repeats the 'pContainer' field if the container is a dock, and is NULL otherwise.
// CairoDock *pDock;
// /// this field repeats the 'pContainer' field if the container is a desklet, and is NULL otherwise.
// CairoDesklet *pDesklet;
// /// a drawing context on the icon.
// cairo_t *pDrawContext;

func NewModuleInstanceFromNative(p unsafe.Pointer) *ModuleInstance {
	return &ModuleInstance{
		Ptr:       (*C.GldiModuleInstance)(p),
		shortkeys: make(map[string]int),
	}
}

func (mi *ModuleInstance) ToNative() unsafe.Pointer {
	return unsafe.Pointer(mi.Ptr)
}

// GetConfFilePath returns the path to the config file of the instance.
//
func (mi *ModuleInstance) GetConfFilePath() string {
	return C.GoString((*C.char)(mi.Ptr.cConfFilePath))
}

func (mi *ModuleInstance) Dock() *CairoDock {
	return NewDockFromNative(unsafe.Pointer(mi.Ptr.pDock))
}

func (mi *ModuleInstance) Module() *Module {
	return NewModuleFromNative(unsafe.Pointer(mi.Ptr.pModule))
}

func (mi *ModuleInstance) Desklet() *Desklet {
	return NewDeskletFromNative(unsafe.Pointer(mi.Ptr.pDesklet))
}

// Icon returns the icon holding the instance.
//
func (mi *ModuleInstance) Icon() *Icon {
	return NewIconFromNative(unsafe.Pointer(mi.Ptr.pIcon))
}

func (mi *ModuleInstance) Detach() {
	C.gldi_module_instance_detach(mi.Ptr)
}

func (mi *ModuleInstance) PopupAboutApplet() {
	C.gldi_module_instance_popup_description(mi.Ptr)
}

func (mi *ModuleInstance) PrepareNewIcons(fields []string) (map[string]*Icon, *C.GList) {
	var list *C.GList
	switch {
	case mi.Ptr.pDock == nil: // In desklet mode.
		list = mi.Desklet().Ptr.icons

	case mi.Icon().GetSubDock() != nil: // In dock with a subdock. Reuse current list.
		list = mi.Icon().GetSubDock().Ptr.icons
	}

	lastIcon := C.cairo_dock_get_last_icon(list)
	n := 0
	if lastIcon != nil {
		n = int(lastIcon.fOrder)
	}

	// GList *pCurrentIconsList = (pInstance->pDock ? (pIcon->pSubDock ? pIcon->pSubDock->icons : NULL) : pInstance->pDesklet->icons);
	// Icon *pLastIcon = cairo_dock_get_last_icon (pCurrentIconsList);
	// int n = (pLastIcon ? pLastIcon->fOrder + 1 : 0);

	icons := make(map[string]*Icon)
	var clist *C.GList
	for i := 0; i < len(fields)/3; i++ {
		id := fields[3*i+2]
		icon := CreateDummyLauncher(fields[3*i], fields[3*i+1], fields[3*i+2], "", float64(i+n))
		clist = C.g_list_append(clist, C.gpointer(icon.Ptr))
		icons[id] = icon
	}

	return icons, clist
}

func (mi *ModuleInstance) InsertIcons(clist *C.GList, dockRenderer, deskletRenderer string) {
	data := [3]C.gpointer{CIntPointer(0), CIntPointer(1), nil}

	C.cairo_dock_insert_icons_in_applet(mi.Ptr,
		clist,
		gchar(dockRenderer), // check if really not to free.
		gchar(deskletRenderer),
		C.gpointer(unsafe.Pointer(&data))) // CairoDeskletRendererConfigPtr
}

func (mi *ModuleInstance) RemoveAllIcons() {
	C.cairo_dock_remove_all_icons_from_applet(mi.Ptr)
}

// NewShortkey is a helper to create a shortkey related to a module instance.
//
func (mi *ModuleInstance) NewShortkey(group, key, desc, shortkey string, call func(string)) *Shortkey {
	vc := mi.Module().VisitCard()

	cGroup := (*C.gchar)(C.CString(group))
	defer C.free(unsafe.Pointer((*C.char)(cGroup)))
	cKey := (*C.gchar)(C.CString(key))
	defer C.free(unsafe.Pointer((*C.char)(cKey)))
	cDesc := (*C.gchar)(C.CString(desc))
	defer C.free(unsafe.Pointer((*C.char)(cDesc)))

	var cShortkey *C.gchar // can be null.
	if shortkey != "" {
		cShortkey = (*C.gchar)(C.CString(shortkey))
		defer C.free(unsafe.Pointer((*C.char)(cShortkey)))
	}

	// find a new shortkeyMapStatic ID.
	max := 0
	shortkeyMapMutex.Lock()

	for k := range shortkeyMapStatic {
		if k > max {
			max = k
		}
	}

	ID := max + 1
	shortkeyMapStatic[ID] = call

	shortkeyMapMutex.Unlock()

	// Save callback reference.
	mi.shortkeys[group+key] = ID

	// Create and return the shortkey.
	c := C.gldi_shortkey_new(cShortkey,
		vc.Ptr.cTitle,
		cDesc,
		vc.Ptr.cIconFilePath, // original conf.
		mi.Ptr.cConfFilePath, // current conf.
		cGroup, cKey,
		C.CDBindkeyHandler(C.onShortkey),
		CIntPointer(ID),
	)

	return NewShortkeyFromNative(unsafe.Pointer(c))
}

//export onShortkey
func onShortkey(cShortkey *C.gchar, cid C.gpointer) {
	id := int(uintptr(cid))
	call := shortkeyMapStatic[id]
	call(C.GoString((*C.char)(cShortkey)))
}

func (mi *ModuleInstance) removeShortkeyMap() {
	// 	shortkeyMapMutex.Lock()
	// 	delete(shortkeyMapStatic, id)
	// 	shortkeyMapMutex.Unlock()
}

//
//---------------------------------------------------------------[ VISITCARD ]--

// VisitCard defines a dock visit card.
//
type VisitCard struct {
	Ptr *C.GldiVisitCard
}

// NewVisitCardFromNative wraps a dock visit card from C pointer.
//
func NewVisitCardFromNative(p unsafe.Pointer) *VisitCard {
	return &VisitCard{(*C.GldiVisitCard)(p)}
}

// NewVisitCardFromPackage wraps a dock visit card from an applet package.
//
func NewVisitCardFromPackage(pack *packages.AppletPackage) *VisitCard {
	vc := new(C.GldiVisitCard)

	vc.cModuleName = gchar(pack.DisplayedName)

	// 		pVisitCard->iMajorVersionNeeded = 2;
	// 		pVisitCard->iMinorVersionNeeded = 1;
	// 		pVisitCard->iMicroVersionNeeded = 1;

	// cShareDataDir ? g_strdup_printf ("%s/preview", cShareDataDir) : NULL;
	vc.cPreviewFilePath = gchar(filepath.Join(pack.Path, "preview"))
	vc.cGettextDomain = gchar("cairo-dock-plugins") // TODO: NEED-GETTEXT-FFS:  GETTEXT_NAME_EXTRAS ... no, need another domain for go applets.
	vc.cUserDataDir = gchar(pack.DisplayedName)
	vc.cShareDataDir = gchar(pack.Path) // TODO: check cShareDataDir
	vc.cConfFileName = gchar(pack.DisplayedName + ".conf")
	vc.cModuleVersion = gchar(pack.Version)
	vc.cAuthor = gchar(pack.Author)
	vc.iCategory = C.GldiModuleCategory(pack.Category)

	if pack.Icon == "" {
		// (cShareDataDir ? g_strdup_printf ("%s/icon", cShareDataDir) : NULL);
		vc.cIconFilePath = gchar(filepath.Join(pack.Path, "icon"))
	} else {
		vc.cIconFilePath = gchar(pack.Icon) // take the filename as it is, the path will be searched when needed only.
	}

	// 		pVisitCard->iSizeOfConfig = 4;  // au cas ou ...
	// 		pVisitCard->iSizeOfData = 4;  // au cas ou ...

	vc.cDescription = gchar(pack.Description)

	// 		pVisitCard->cTitle = cTitle ?
	// 			g_strdup (dgettext (pVisitCard->cGettextDomain, cTitle)) :
	// 			g_strdup (cModuleName);

	vc.cTitle = gchar(pack.DisplayedName) // TODO:  improve

	vc.iContainerType = C.CAIRO_DOCK_MODULE_CAN_DOCK | C.CAIRO_DOCK_MODULE_CAN_DESKLET
	vc.bMultiInstance = cbool(pack.IsMultiInstance)
	vc.bActAsLauncher = cbool(pack.ActAsLauncher) // ex.: XChat controls xchat/xchat-gnome, but it does that only after initializing; we need to know if it's a launcher before the taskbar is loaded, hence this parameter.
	return NewVisitCardFromNative(unsafe.Pointer(vc))
}

// GetModuleName returns the module name (real one)
func (vc *VisitCard) GetModuleName() string {
	return C.GoString((*C.char)(vc.Ptr.cModuleName))
}

// GetName returns the module name used as identifier.
//
func (vc *VisitCard) GetName() string {
	return C.GoString((*C.char)(vc.Ptr.cModuleName))
}

// GetTitle returns the module name translated (as seen by the user).
//
func (vc *VisitCard) GetTitle() string {
	return C.GoString((*C.char)(vc.Ptr.cTitle))
}

// GetIconFilePath returns the module icon file path.
//
func (vc *VisitCard) GetIconFilePath() string {
	return C.GoString((*C.char)(vc.Ptr.cIconFilePath))
}

// GetAuthor returns the module author name.
//
func (vc *VisitCard) GetAuthor() string {
	return C.GoString((*C.char)(vc.Ptr.cAuthor))
}

// GetDescription returns the module description text translated.
//
func (vc *VisitCard) GetDescription() string {
	desc := C.GoString((*C.char)(vc.Ptr.cDescription))
	return tran.Sloc(vc.GetGettextDomain(), desc)
}

// GetPreviewFilePath returns the module preview file path.
//
func (vc *VisitCard) GetPreviewFilePath() string {
	return C.GoString((*C.char)(vc.Ptr.cPreviewFilePath))
}

// GetShareDataDir returns the module share data dir.
//
func (vc *VisitCard) GetShareDataDir() string {
	return C.GoString((*C.char)(vc.Ptr.cShareDataDir))
}

// GetGettextDomain returns the module gettext domain.
//
func (vc *VisitCard) GetGettextDomain() string {
	return C.GoString((*C.char)(vc.Ptr.cGettextDomain))
}

// GetModuleVersion returns the module version.
//
func (vc *VisitCard) GetModuleVersion() string {
	return C.GoString((*C.char)(vc.Ptr.cModuleVersion))
}

// GetCategory returns the module category.
//
func (vc *VisitCard) GetCategory() ModuleCategory {
	return ModuleCategory(vc.Ptr.iCategory)
}

// GetConfFileName returns the module container file name.
//
func (vc *VisitCard) GetConfFileName() string {
	return C.GoString((*C.char)(vc.Ptr.cConfFileName))
}

// IsMultiInstance returns whether the module can be activated multiple times or not.
//
func (vc *VisitCard) IsMultiInstance() bool {
	return gobool(vc.Ptr.bMultiInstance)
}

// GetContainerType returns the module icon file path.
//
func (vc *VisitCard) GetContainerType() int {
	return int(vc.Ptr.iContainerType)
}

//
//---------------------------------------------------------------[ WINDOWACTOR ]--

// WindowActor defines a dock window actor.
//
type WindowActor struct {
	Ptr *C.GldiWindowActor
}

// NewWindowActorFromNative wraps a dock window actor from C pointer.
//
func NewWindowActorFromNative(p unsafe.Pointer) *WindowActor {
	if p == nil {
		return nil
	}
	return &WindowActor{(*C.GldiWindowActor)(p)}
}

// ToNative returns the pointer to the native object.
//
func (o *WindowActor) ToNative() unsafe.Pointer {
	return unsafe.Pointer(o.Ptr)
}

// CanMinMaxClose returns whether the window can do those actions.
//
func (o *WindowActor) CanMinMaxClose() (bool, bool, bool) {
	var bCanMinimize, bCanMaximize, bCanClose C.gboolean
	C.gldi_window_can_minimize_maximize_close(o.Ptr, &bCanMinimize, &bCanMaximize, &bCanClose)
	return gobool(bCanMinimize), gobool(bCanMaximize), gobool(bCanClose)
}

// IsActive returns whether the window is active.
//
func (o *WindowActor) IsActive() bool {
	return o.Ptr == C.gldi_windows_get_active()
}

// IsAbove returns whether the window is above.
//
func (o *WindowActor) IsAbove() bool { // could split OrBelow but seem unused.
	var isAbove, isBelow C.gboolean
	C.gldi_window_is_above_or_below(o.Ptr, &isAbove, &isBelow)
	return gobool(isAbove)
}

// IsFullScreen returns whether the window is full screen.
//
func (o *WindowActor) IsFullScreen() bool {
	return gobool(o.Ptr.bIsFullScreen)
}

// IsHidden returns whether the window is hidden.
//
func (o *WindowActor) IsHidden() bool {
	return gobool(o.Ptr.bIsHidden)
}

// IsMaximized returns whether the window is maximized.
//
func (o *WindowActor) IsMaximized() bool {
	return gobool(o.Ptr.bIsMaximized)
}

// IsOnCurrentDesktop returns whether the window is on current desktop.
//
func (o *WindowActor) IsOnCurrentDesktop() bool {
	return gobool(C.gldi_window_is_on_current_desktop(o.Ptr))
}

// IsOnDesktop returns whether the window is the given desktop and viewport number.
//
func (o *WindowActor) IsOnDesktop(desktopNumber, viewPortX, viewPortY int) bool {
	return gobool(C.gldi_window_is_on_desktop(o.Ptr, C.int(desktopNumber), C.int(viewPortX), C.int(viewPortY)))
}

// IsSticky returns whether the window is sticky.
//
func (o *WindowActor) IsSticky() bool {
	return gobool(C.gldi_window_is_sticky(o.Ptr))
}

// Close closes window.
//
func (o *WindowActor) Close() {
	C.gldi_window_close(o.Ptr)
}

// Kill kills the window.
//
func (o *WindowActor) Kill() {
	C.gldi_window_kill(o.Ptr)
}

// Lower lowers the window.
//
func (o *WindowActor) Lower() {
	C.gldi_window_lower(o.Ptr)
}

// Minimize minimizes the window.
//
func (o *WindowActor) Minimize() {
	C.gldi_window_minimize(o.Ptr)
}

// Maximize maximizes the window.
//
func (o *WindowActor) Maximize(full bool) {
	C.gldi_window_maximize(o.Ptr, cbool(full))
}

// MoveToCurrentDesktop moves the window to the current desktop.
//
func (o *WindowActor) MoveToCurrentDesktop() {
	C.gldi_window_move_to_current_desktop(o.Ptr)
}

// MoveToDesktop move the window to the given desktop and viewport number.
//
func (o *WindowActor) MoveToDesktop(desktopNumber, viewPortX, viewPortY int) {
	C.gldi_window_move_to_desktop(o.Ptr, C.int(desktopNumber), C.int(viewPortX), C.int(viewPortY))
}

// SetAbove sets the window above.
//
func (o *WindowActor) SetAbove(above bool) {
	C.gldi_window_set_above(o.Ptr, cbool(above))
}

// SetFullScreen sets the window full screen state.
//
func (o *WindowActor) SetFullScreen(full bool) {
	C.gldi_window_set_fullscreen(o.Ptr, cbool(full))
}

// SetSticky sets the window sticky state.
//
func (o *WindowActor) SetSticky(sticky bool) {
	C.gldi_window_set_sticky(o.Ptr, cbool(sticky))
}

// Show shows the window.
//
func (o *WindowActor) Show() {
	C.gldi_window_show(o.Ptr)
}

// SetVisibility sets the window visibility.
//
func (o *WindowActor) SetVisibility(show bool) {
	if show {
		o.Show()
	}
	o.Minimize()
}

// ToggleVisibility toggles the window visibility.
//
func (o *WindowActor) ToggleVisibility() {
	if o.IsActive() {
		o.Minimize()
	} else {
		o.Show()
	}
}

// GetAppliIcon returns the icon managing the window.
//
func (o *WindowActor) GetAppliIcon() *Icon {
	c := C.cairo_dock_get_appli_icon(o.Ptr)
	return NewIconFromNative(unsafe.Pointer(c))
}

//
//---------------------------------------------------------------[ ANIMATION ]--

// Animation defines a dock animation.
//
type Animation struct {
	Ptr *C.CairoDockAnimationRecord
}

// NewAnimationFromNative wraps a dock animation from C pointer.
//
func NewAnimationFromNative(p unsafe.Pointer) *Animation {
	return &Animation{(*C.CairoDockAnimationRecord)(p)}
}

// GetDisplayedName returns the animation displayed name.
//
func (dr *Animation) GetDisplayedName() string {
	return C.GoString((*C.char)(dr.Ptr.cDisplayedName))
}

//
//--------------------------------------------------[ CAIRODESKLETDECORATION ]--

// CairoDeskletDecoration defines a dock desklet decoration.
//
type CairoDeskletDecoration struct {
	Ptr *C.CairoDeskletDecoration
}

// NewCairoDeskletDecorationFromNative wraps a dock desklet decoration from C pointer.
//
func NewCairoDeskletDecorationFromNative(p unsafe.Pointer) *CairoDeskletDecoration {
	return &CairoDeskletDecoration{(*C.CairoDeskletDecoration)(p)}
}

// GetDisplayedName returns the desklet decoration displayed name.
//
func (dr *CairoDeskletDecoration) GetDisplayedName() string {
	return C.GoString((*C.char)(dr.Ptr.cDisplayedName))
}

//
//---------------------------------------------------------[ DIALOGDECORATOR ]--

// DialogDecorator defines a dock dialog decorator.
//
type DialogDecorator struct {
	Ptr *C.CairoDialogDecorator
}

// NewDialogDecoratorFromNative wraps a dock dialog decorator from C pointer.
//
func NewDialogDecoratorFromNative(p unsafe.Pointer) *DialogDecorator {
	return &DialogDecorator{(*C.CairoDialogDecorator)(p)}
}

// GetDisplayedName returns the dialog decorator displayed name.
//
func (dr *DialogDecorator) GetDisplayedName() string {
	return C.GoString((*C.char)(dr.Ptr.cDisplayedName))
}

//
//-------------------------------------------------------[ CAIRODOCKRENDERER ]--

// CairoDockRenderer defines a dock renderer, AKA views.
//
type CairoDockRenderer struct {
	Ptr *C.CairoDockRenderer
}

// NewCairoDockRenderer creates a new dock renderer.
//
func NewCairoDockRenderer(title, readmePath, previewPath string) *CairoDockRenderer {
	c := C.newDockRenderer()

	// do not free strings.
	c.cDisplayedName = (*C.gchar)(C.CString(title))
	c.cReadmeFilePath = (*C.gchar)(C.CString(readmePath))
	c.cPreviewFilePath = (*C.gchar)(C.CString(previewPath))

	// pRenderer->cDisplayedName = _ (name);

	// parameters
	c.bUseReflect = 0 // false

	return NewCairoDockRendererFromNative(unsafe.Pointer(c))
}

// ToNative returns the pointer to the native object.
//
func (dr *CairoDockRenderer) ToNative() unsafe.Pointer {
	return unsafe.Pointer(dr.Ptr)
}

// NewCairoDockRendererFromNative wraps a dock renderer from C pointer.
//
func NewCairoDockRendererFromNative(p unsafe.Pointer) *CairoDockRenderer {
	return &CairoDockRenderer{(*C.CairoDockRenderer)(p)}
}

// GetDisplayedName returns the dock renderer displayed name.
//
func (dr *CairoDockRenderer) GetDisplayedName() string {
	return C.GoString((*C.char)(dr.Ptr.cDisplayedName))
}

// GetReadmeFilePath returns the dock renderer readme file path.
//
func (dr *CairoDockRenderer) GetReadmeFilePath() string {
	return C.GoString((*C.char)(dr.Ptr.cReadmeFilePath))
}

// GetPreviewFilePath returns the dock renderer preview file path.
//
func (dr *CairoDockRenderer) GetPreviewFilePath() string {
	return C.GoString((*C.char)(dr.Ptr.cPreviewFilePath))
}

// Register registers the dock renderer in the dock.
//
func (dr *CairoDockRenderer) Register(name string) {
	cname := (*C.gchar)(C.CString(name))
	C.cairo_dock_register_renderer(cname, dr.Ptr)
}

//
//-----------------------------------------------------------------[ MANAGER ]--

// Manager defines a dock manager.
//
type Manager struct {
	Ptr *C.GldiManager
}

// NewManagerFromNative wraps a dock manager from C pointer.
//
func NewManagerFromNative(p unsafe.Pointer) *Manager {
	return &Manager{(*C.GldiManager)(p)}
}

// ManagerGet gets the manager with the given name.
//
func ManagerGet(name string) *Manager {
	cstr := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	c := C.gldi_manager_get(cstr)
	if c == nil {
		return nil
	}

	return NewManagerFromNative(unsafe.Pointer(c))
}

// ManagerReload reloads the manager matching the given name.
//
func ManagerReload(name string, b bool, keyf *keyfile.KeyFile) {
	manager := ManagerGet(name)
	if manager == nil {
		return
	}
	C.manager_reload(manager.Ptr, cbool(b), (*C.GKeyFile)(unsafe.Pointer(keyf.ToNative())))
}

//
//---------------------------------------------------------[ DESKTOPGEOMETRY ]--

// DesktopGeometry defines the dock desktop geometry.
//
type DesktopGeometry struct {
	Ptr C.GldiDesktopGeometry
}

// GetDesktopGeometry gets desktop geometry settings.
//
func GetDesktopGeometry() *DesktopGeometry {
	return &DesktopGeometry{C.g_desktopGeometry}
}

// NbDesktops returns the number of desktops.
//
func (dg *DesktopGeometry) NbDesktops() int {
	return int(dg.Ptr.iNbDesktops)
}

// NbScreens returns the number of desktop screens.
//
func (dg *DesktopGeometry) NbScreens() int {
	return int(dg.Ptr.iNbScreens)
}

// NbViewportX returns the number of horizontal viewports.
//
func (dg *DesktopGeometry) NbViewportX() int {
	return int(dg.Ptr.iNbViewportX)
}

// NbViewportY returns the number of vertical viewports.
//
func (dg *DesktopGeometry) NbViewportY() int {
	return int(dg.Ptr.iNbViewportY)
}

// WidthAll returns the desktop total width.
//
func (dg *DesktopGeometry) WidthAll() int {
	return int(dg.Ptr.Xscreen.width)
}

// WidthScreen returns the width of a desktop screen.
//
func (dg *DesktopGeometry) WidthScreen(screenNb int) int {
	return int(C.screen_width_i(C.int(screenNb)))
}

// HeightAll returns the desktop total height.
//
func (dg *DesktopGeometry) HeightAll() int {
	return int(dg.Ptr.Xscreen.height)
}

// HeightScreen returns the height of a desktop screen.
//
func (dg *DesktopGeometry) HeightScreen(screenNb int) int {
	return int(C.screen_height_i(C.int(screenNb)))
}

// int iCurrentDesktop, iCurrentViewportX, iCurrentViewportY;

// ScreenPosition returns the desktop screen position.
//
func (dg *DesktopGeometry) ScreenPosition(i int) (int, int) {
	if 0 > i || i >= dg.NbScreens() {
		return 0, 0
	}
	x := int(C.screen_position_x(C.int(i)))
	y := int(C.screen_position_y(C.int(i)))
	return x, y
}

//
//----------------------------------------------------------------[ SHORTKEY ]--

// Shortkey defines a dock shortkey.
//
type Shortkey struct {
	Ptr *C.GldiShortkey
}

// NewShortkeyFromNative wraps a dock shortkey from C pointer.
//
func NewShortkeyFromNative(p unsafe.Pointer) *Shortkey {
	return &Shortkey{(*C.GldiShortkey)(p)}
}

// GetDemander returns the shortkey Demander.
//
func (dr *Shortkey) GetDemander() string {
	return C.GoString((*C.char)(dr.Ptr.cDemander))
}

// GetDescription returns the shortkey description.
//
func (dr *Shortkey) GetDescription() string {
	return C.GoString((*C.char)(dr.Ptr.cDescription))
}

// GetKeyString returns the shortkey key reference as string.
//
func (dr *Shortkey) GetKeyString() string {
	return C.GoString((*C.char)(dr.Ptr.keystring))
}

// GetIconFilePath returns the shortkey icon file path.
//
func (dr *Shortkey) GetIconFilePath() string {
	return C.GoString((*C.char)(dr.Ptr.cIconFilePath))
}

// GetConfFilePath returns the shortkey conf file path.
//
func (dr *Shortkey) GetConfFilePath() string {
	return C.GoString((*C.char)(dr.Ptr.cConfFilePath))
}

// GetGroupName returns the shortkey group name.
//
func (dr *Shortkey) GetGroupName() string {
	return C.GoString((*C.char)(dr.Ptr.cGroupName))
}

// GetKeyName returns the config key name.
//
func (dr *Shortkey) GetKeyName() string {
	return C.GoString((*C.char)(dr.Ptr.cKeyName))
}

// GetSuccess returns the shortkey success.
//
func (dr *Shortkey) GetSuccess() bool {
	return gobool(dr.Ptr.bSuccess)
}

// Rebind rebinds the shortkey.
//
func (dr *Shortkey) Rebind(keystring, description string) bool {
	ckey := (*C.gchar)(C.CString(keystring))
	defer C.free(unsafe.Pointer((*C.char)(ckey)))
	var cdesc *C.gchar
	if description != "" {
		cdesc := (*C.gchar)(C.CString(description))
		defer C.free(unsafe.Pointer((*C.char)(cdesc)))
	}

	c := C.gldi_shortkey_rebind(dr.Ptr, ckey, cdesc)
	return gobool(c)
}

//
//-----------------------------------------------------------------[ HELPERS ]--

func cbool(b bool) C.gboolean {
	if b {
		return C.gboolean(1)
	}
	return C.gboolean(0)
}

func gobool(b C.gboolean) bool {
	if b == 1 {
		return true
	}
	return false
}

func gchar(str string) *C.gchar {
	if str == "" {
		return nil
	}
	return (*C.gchar)(C.CString(str))
}

// CIntPointer returns a C gpointer to an int value.
//
func CIntPointer(i int) C.gpointer {
	return C.intToPointer(C.int(i))
}

// //

//
//---------------------------------------------------------[ LISTS TO NATIVE ]--

func goListDocks(clist *glib.List) (list []*CairoDock) {
	clist.Foreach(func(data interface{}) {
		list = append(list, NewDockFromNative(data.(unsafe.Pointer)))
	})
	return
}

func goListIcons(clist *glib.List) (list []*Icon) {
	clist.Foreach(func(data interface{}) {
		list = append(list, NewIconFromNative(data.(unsafe.Pointer)))
	})
	return
}

func goListModuleInstance(clist *glib.List) (list []*ModuleInstance) {
	clist.Foreach(func(data interface{}) {
		list = append(list, NewModuleInstanceFromNative(data.(unsafe.Pointer)))
	})
	return
}

func cListGdouble(value []float64) *C.gdouble {
	var clist *C.gdouble
	clist = (*C.gdouble)(C.malloc(C.size_t(int(unsafe.Sizeof(*clist)) * len(value))))
	for i, e := range value {
		(*(*[999999]C.gdouble)(unsafe.Pointer(clist)))[i] = C.gdouble(e)
	}
	return clist
}
