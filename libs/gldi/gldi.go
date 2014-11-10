// Package gldi provides Go bindings for gldi (cairo-dock).  Supports version 3.4
package gldi

/*
#cgo pkg-config: glib-2.0 gldi
#include <stdlib.h>                              // free
#include <dbus/dbus-glib.h>                      // dbus_g_thread_init
#include <glib/gkeyfile.h>                       // GKeyFile


#include "cairo-dock-core.h"
#include "cairo-dock-animations.h"               // cairo_dock_trigger_icon_removal_from_dock
#include "cairo-dock-applet-facility.h"      // cairo_dock_pop_up_about_applet
#include "cairo-dock-backends-manager.h"         // cairo_dock_foreach_dock_renderer
#include "cairo-dock-config.h"                   // cairo_dock_load_current_theme
#include "cairo-dock-class-manager.h"            // cairo_dock_get_class_command
#include "cairo-dock-class-icon-manager.h"       // myClassIconObjectMgr
#include "cairo-dock-desklet-manager.h"          // myDeskletObjectMgr
#include "cairo-dock-desktop-manager.h"          // g_desktopGeometry
#include "cairo-dock-dock-factory.h"             // CairoDock
#include "cairo-dock-dock-facility.h"            // cairo_dock_get_available_docks
#include "cairo-dock-dock-manager.h"             // gldi_dock_get_readable_name
#include "cairo-dock-file-manager.h"             // CAIRO_DOCK_GNOME...
#include "cairo-dock-icon-factory.h"             // Icon
#include "cairo-dock-icon-facility.h"        // Icon
#include "cairo-dock-keybinder.h"                // gldi_shortkeys_foreach
#include "cairo-dock-launcher-manager.h"         // CAIRO_DOCK_ICON_TYPE_IS_LAUNCHER
#include "cairo-dock-log.h"                      // cd_log_set_level_from_name
#include "cairo-dock-menu.h"  // ModuleInstance
#include "cairo-dock-module-instance-manager.h"  // ModuleInstance
#include "cairo-dock-module-manager.h"           // gldi_modules_new_from_directory
#include "cairo-dock-object.h"                   // Icon
#include "cairo-dock-opengl.h"                   // gldi_gl_backend_force_indirect_rendering
#include "cairo-dock-separator-manager.h"        // CAIRO_DOCK_ICON_TYPE_IS_SEPARATOR
#include "cairo-dock-struct.h"                   // CAIRO_DOCK_LAST_ORDER
#include "cairo-dock-stack-icon-manager.h"       // CAIRO_DOCK_ICON_TYPE_IS_CONTAINER
#include "cairo-dock-themes-manager.h"           // cairo_dock_set_paths
#include "cairo-dock-windows-manager.h"          // gldi_window_can_minimize_maximize_close


extern CairoDock *g_pMainDock;

extern CairoDockGLConfig g_openglConfig;
extern gboolean          g_bUseOpenGL;

extern gchar *g_cCurrentLaunchersPath;

extern GldiDesktopGeometry g_desktopGeometry;

static int screen_position_x(int i) { return g_desktopGeometry.pScreens[i].x; }
static int screen_position_y(int i) { return g_desktopGeometry.pScreens[i].y; }

static gboolean IconIsSeparator    (Icon *icon) { return CAIRO_DOCK_ICON_TYPE_IS_SEPARATOR(icon); }
static gboolean IconIsSeparatorAuto(Icon *icon) { return CAIRO_DOCK_IS_AUTOMATIC_SEPARATOR(icon); }
static gboolean IconIsLauncher     (Icon *icon) { return CAIRO_DOCK_ICON_TYPE_IS_LAUNCHER(icon); }
static gboolean IconIsStackIcon    (Icon *icon) { return CAIRO_DOCK_ICON_TYPE_IS_CONTAINER(icon); }


extern void onDialogAnswer (int iClickedButton, GtkWidget *pInteractiveWidget, gpointer data, CairoDialog *pDialog);



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



//
//---------------------------------------------------------[ LSTS FORWARDING ]--

extern void addShortkey   (gpointer, gpointer);
extern void addItemToList (gpointer, gchar*, gpointer);

static void     fwd_one (const gchar *name, gpointer *item, gpointer p) { addItemToList(p, g_strdup(name), item); }
static gboolean fwd_chk (      gchar *name, gpointer *item, gpointer p) { addItemToList(p, name,           item); return FALSE;}

static void list_shortkey           (gpointer p){ gldi_shortkeys_foreach                ((GFunc)addShortkey, p); }

static void list_animation          (gpointer p){ cairo_dock_foreach_animation          ((GHFunc)fwd_one, p); }
static void list_desklet_decoration (gpointer p){ cairo_dock_foreach_desklet_decoration ((GHFunc)fwd_one, p); }
static void list_dialog_decorator   (gpointer p){ cairo_dock_foreach_dialog_decorator   ((GHFunc)fwd_one, p); }
static void list_dock_renderer      (gpointer p){ cairo_dock_foreach_dock_renderer      ((GHFunc)fwd_one, p); }

static void list_dock_module        (gpointer p){ gldi_module_foreach                   ((GHRFunc)fwd_chk, p); }




*/
import "C"

import (
	"github.com/bradfitz/iter"
	"github.com/conformal/gotk3/gdk"
	"github.com/conformal/gotk3/glib"
	"github.com/conformal/gotk3/gtk"
	"github.com/gosexy/gettext"

	"github.com/sqp/godock/widgets/gtk/keyfile"

	"errors"
	"path/filepath"
	"unsafe"
)

const IconLastOrder = -1e9 // CAIRO_DOCK_LAST_ORDER

type RenderingMethod int

const (
	RenderingOpenGL  RenderingMethod = C.GLDI_OPENGL
	RenderingCairo   RenderingMethod = C.GLDI_CAIRO
	RenderingDefault RenderingMethod = C.GLDI_DEFAULT
)

type DesktopEnvironment C.CairoDockDesktopEnv

const (
	DesktopEnvGnome   DesktopEnvironment = C.CAIRO_DOCK_GNOME
	DesktopEnvKDE     DesktopEnvironment = C.CAIRO_DOCK_KDE
	DesktopEnvXFCE    DesktopEnvironment = C.CAIRO_DOCK_XFCE
	DesktopEnvUnknown DesktopEnvironment = C.CAIRO_DOCK_UNKNOWN_ENV
)

type ModuleCategory C.GldiModuleCategory

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

// LoadCurrentTheme loads the theme in the dock.
//
func LoadCurrentTheme() {
	C.cairo_dock_load_current_theme()
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

// GLBackendIsUsed returns whether the OpenGL backend is safely usable or not.
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

func LogSetLevelFromName(verbosity string) {
	cstr := (*C.gchar)(C.CString(verbosity))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	C.cd_log_set_level_from_name(cstr)
}

func LogForceUseColor() {
	C.cd_log_force_use_color()
}

func DbusGThreadInit() {
	C.dbus_g_thread_init() // it's a wrapper: it will use dbus_threads_init_default ();
}

func FreeAll() {
	C.gldi_free_all()
}

func XMLCleanupParser() {
	C.xmlCleanupParser()
}

//
//-------------------------------------------------------------[ GLDI COMMON ]--

// need params
// pParentDock	excluding this dock if not NULL
// pSubDock	excluding this dock and its children if not NUL
func GetAllAvailableDocks(parent, subdock *CairoDock) []*CairoDock {
	var cp, cs *C.CairoDock
	if parent != nil {
		cp = parent.Ptr
	}
	if subdock != nil {
		cs = subdock.Ptr
	}

	clist := (*glib.List)(unsafe.Pointer(C.cairo_dock_get_available_docks(cp, cs)))
	defer clist.Free()
	return goListDocks(clist)
}

func DockGet(containerName string) *CairoDock {
	cstr := (*C.gchar)(C.CString(containerName))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	return NewDockFromNative(unsafe.Pointer(C.gldi_dock_get(cstr)))
}

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

type IObject interface {
	ToNative() unsafe.Pointer
}

func ObjectReload(obj IObject) {
	C.gldi_object_reload((*C.GldiObject)(obj.ToNative()), C.gboolean(1))
}

func ObjectDelete(obj IObject) {
	C.gldi_object_delete((*C.GldiObject)(obj.ToNative()))
}

func ObjectIsManagerChild(obj IObject, ptr *C.GldiObjectManager) bool {
	return gobool(C.gldi_object_is_manager_child((*C.GldiObject)(obj.ToNative()), ptr))
}

func ObjectIsDock(obj IObject) bool {
	return ObjectIsManagerChild(obj, &C.myDockObjectMgr)
}

func ObjectNotify(container *Container, notif int, icon *Icon, dock *CairoDock, key gdk.ModifierType) {
	C.objectNotify(container.Ptr, C.int(notif), icon.Ptr, dock.Ptr, C.GdkModifierType(key))
}

func QuickHideAllDocks() {
	C.cairo_dock_quick_hide_all_docks()
}

func LauncherAddNew(uri string, dock *CairoDock, order float64) *Icon {
	var cstr *C.gchar
	if uri != "" {
		cstr = (*C.gchar)(C.CString(uri))
		defer C.free(unsafe.Pointer((*C.char)(cstr)))
	}
	c := C.gldi_launcher_add_new(cstr, dock.Ptr, C.double(order))
	return NewIconFromNative(unsafe.Pointer(c))
}

func SeparatorIconAddNew(dock *CairoDock, order float64) *Icon {
	c := C.gldi_separator_icon_add_new(dock.Ptr, C.double(order))
	return NewIconFromNative(unsafe.Pointer(c))
}

func StackIconAddNew(dock *CairoDock, order float64) *Icon {
	c := C.gldi_stack_icon_add_new(dock.Ptr, C.double(order))
	return NewIconFromNative(unsafe.Pointer(c))
}

func DialogShowGeneralMessage(str string, duration float64) {
	cstr := (*C.gchar)(C.CString(str))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	C.gldi_dialog_show_general_message(cstr, C.double(duration))
}

func DialogShowTemporaryWithIcon(str string, icon *Icon, container *Container, duration float64, iconPath string) {
	cstr := (*C.gchar)(C.CString(str))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	cpath := (*C.gchar)(C.CString(iconPath))
	defer C.free(unsafe.Pointer((*C.char)(cpath)))

	C.gldi_dialog_show_temporary_with_icon(cstr, icon.Ptr, container.Ptr, C.double(duration), cpath)
}

func DialogShowWithQuestion(str string, icon *Icon, container *Container, iconPath string, onAnswer func(int, *gtk.Widget)) {
	cstr := (*C.gchar)(C.CString(str))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	cpath := (*C.gchar)(C.CString(iconPath))
	defer C.free(unsafe.Pointer((*C.char)(cpath)))

	answer := &listForward{onAnswer}

	C.gldi_dialog_show_with_question(cstr, icon.Ptr, container.Ptr, cpath, C.CairoDockActionOnAnswerFunc(C.onDialogAnswer), C.gpointer(answer), nil)
}

//export onDialogAnswer
func onDialogAnswer(clickedButton C.int, widget *C.GtkWidget, data C.gpointer, dialog *C.CairoDialog) {

	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(widget))}
	w := &gtk.Widget{glib.InitiallyUnowned{obj}}

	uncast := (*listForward)(data)

	call := (uncast.p).(func(int, *gtk.Widget))
	if call != nil {
		call(int(clickedButton), w)
	}
}

// CairoDialog *gldi_dialog_show_with_question (const gchar *cText, Icon *pIcon, GldiContainer *pContainer, const gchar *cIconPath, CairoDockActionOnAnswerFunc pActionFunc, gpointer data, GFreeFunc pFreeDataFunc)

/** A convenient function to add a sub-menu to a given menu.
 * @param pMenu the menu
 * @param cLabel the label, or NULL
 * @param cImage the image path or name, or NULL
 * @param pMenuItemPtr pointer that will contain the new menu-item, or NULL
 * @return the new sub-menu that has been added.
 */
// TODO: add last option for GtkWidget **pMenuItemPtr
func MenuAddSubMenu(menu *gtk.Menu, label, iconPath string) (*gtk.Menu, *MenuItem) {
	cstr := (*C.gchar)(C.CString(label))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	var cpath *C.gchar
	if iconPath != "" {
		cpath = (*C.gchar)(C.CString(iconPath))
		defer C.free(unsafe.Pointer((*C.char)(cpath)))
	}
	var cmenuitem *C.GtkWidget
	c := C.gldi_menu_add_sub_menu_full((*C.GtkWidget)(unsafe.Pointer(menu.Native())), cstr, cpath, &cmenuitem)

	if c == nil {
		return nil, nil
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	submenu := &gtk.Menu{gtk.MenuShell{gtk.Container{gtk.Widget{glib.InitiallyUnowned{obj}}}}}

	obj = &glib.Object{glib.ToGObject(unsafe.Pointer(cmenuitem))}
	item := &gtk.MenuItem{gtk.Bin{gtk.Container{gtk.Widget{glib.InitiallyUnowned{obj}}}}}

	return submenu, &MenuItem{*item}
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
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	item := &gtk.MenuItem{gtk.Bin{gtk.Container{gtk.Widget{glib.InitiallyUnowned{obj}}}}}

	return item
}

type MenuItem struct {
	gtk.MenuItem
}

func PopupAboutApplet(mi *ModuleInstance) {
	C.cairo_dock_pop_up_about_applet(nil, mi.Ptr)
	// C.cairo_dock_pop_up_about_applet(nil, (*C.GldiModuleInstance)(unsafe.Pointer(icon.ModuleInstance().Ptr)))
}

//
//---------------------------------------------------------------[ CAIRODOCK ]--

type CairoDock struct {
	Ptr *C.CairoDock
}

func NewDockFromNative(p unsafe.Pointer) *CairoDock {
	if p == nil {
		return nil
	}
	return &CairoDock{(*C.CairoDock)(p)}
}

func (o *CairoDock) ToNative() unsafe.Pointer {
	return unsafe.Pointer(o.Ptr)
}

func (o *CairoDock) ToContainer() *Container {
	return NewContainerFromNative(unsafe.Pointer(o.Ptr))
}

func (dock *CairoDock) Icons() (list []*Icon) {
	clist := (*glib.List)(unsafe.Pointer(dock.Ptr.icons))
	return goListIcons(clist)
}

func (dock *CairoDock) GetDockName() string {
	return C.GoString((*C.char)(dock.Ptr.cDockName))
}

func (dock *CairoDock) GetReadableName() string {
	return C.GoString((*C.char)(C.gldi_dock_get_readable_name(dock.Ptr)))
}

func (dock *CairoDock) GetRefCount() int {
	return int(dock.Ptr.iRefCount)
}

func (dock *CairoDock) IsMainDock() bool {
	return gobool(dock.Ptr.bIsMainDock)
}

func (dock *CairoDock) IsAutoHide() bool {
	return gobool(dock.Ptr.bAutoHide)
}

func (dock *CairoDock) SearchIconPointingOnDock(unknown interface{}) *Icon { // TODO: add param CairoDock **pParentDock
	c := C.cairo_dock_search_icon_pointing_on_dock(dock.Ptr, nil)
	return NewIconFromNative(unsafe.Pointer(c))
}

func (dock *CairoDock) GetPointedIcon() *Icon {
	c := C.cairo_dock_get_pointed_icon(dock.Ptr.icons)
	return NewIconFromNative(unsafe.Pointer(c))
}

func (dock *CairoDock) Container() *Container {
	return NewContainerFromNative(unsafe.Pointer(&dock.Ptr.container))
}

func (o *CairoDock) GetNextIcon(icon *Icon) *Icon {
	c := C.cairo_dock_get_next_icon(o.Ptr.icons, icon.Ptr)
	return NewIconFromNative(unsafe.Pointer(c))
}

func (o *CairoDock) GetPreviousIcon(icon *Icon) *Icon {
	c := C.cairo_dock_get_previous_icon(o.Ptr.icons, icon.Ptr)
	return NewIconFromNative(unsafe.Pointer(c))
}

//
//-------------------------------------------------------[ CONTAINER ]--

type Container struct {
	Ptr *C.GldiContainer
}

func NewContainerFromNative(p unsafe.Pointer) *Container {
	return &Container{(*C.GldiContainer)(p)}
}

func (o *Container) ToNative() unsafe.Pointer {
	return unsafe.Pointer(o.Ptr)
}

func (o *Container) ToCairoDock() *CairoDock {
	return NewDockFromNative(unsafe.Pointer(o.Ptr))
}

func (o *Container) ToDesklet() *Desklet {
	return NewDeskletFromNative(unsafe.Pointer(o.Ptr))
}

func (o *Container) IsDesklet() bool {
	return ObjectIsManagerChild(o, &C.myDeskletObjectMgr)
}

func (o *Container) MouseX() int {
	return int(o.Ptr.iMouseX)
}

func (o *Container) MouseY() int {
	return int(o.Ptr.iMouseY)
}

//
//-----------------------------------------------------------------[ DESKLET ]--

type Desklet struct {
	Ptr *C.CairoDesklet
}

func NewDeskletFromNative(p unsafe.Pointer) *Desklet {
	return &Desklet{(*C.CairoDesklet)(p)}
}

func (o *Desklet) ToNative() unsafe.Pointer {
	return unsafe.Pointer(o.Ptr)
}

func (o *Desklet) IsSticky() bool {
	return gobool(C.gldi_desklet_is_sticky(o.Ptr))
}

func (o *Desklet) PositionLocked() bool {
	return gobool(o.Ptr.bPositionLocked)
}

func (o *Desklet) GetIcon() *Icon {
	return NewIconFromNative(unsafe.Pointer(o.Ptr.pIcon))
}

func (o *Desklet) SetSticky(b bool) {
	C.gldi_desklet_set_sticky(o.Ptr, cbool(b))
}

func (o *Desklet) LockPosition(b bool) {
	C.gldi_desklet_lock_position(o.Ptr, cbool(b))
}

//
//--------------------------------------------------------------------[ ICON ]--

type Icon struct {
	Ptr *C.Icon
}

func NewIconFromNative(p unsafe.Pointer) *Icon {
	if p == nil {
		return nil
	}
	return &Icon{(*C.Icon)(p)}
}

func IconsGetAnyWithoutDialog() *Icon {
	return NewIconFromNative(unsafe.Pointer(C._icons_get_any_without_dialog()))
}

func (o *Icon) ToNative() unsafe.Pointer {
	return unsafe.Pointer(o.Ptr)
}

func (icon *Icon) GetClass() string {
	return C.GoString((*C.char)(icon.Ptr.cClass))
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

func (o *Icon) Scale() float64 {
	return float64(o.Ptr.fScale)
}
func (o *Icon) Order() float64 {
	return float64(o.Ptr.fOrder)
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
	return icon.IsAppli() && !icon.IsLauncher()
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
	return &ModuleInstance{icon.Ptr.pModuleInstance}
}

func (icon *Icon) GetSubDock() *CairoDock {
	return NewDockFromNative(unsafe.Pointer(icon.Ptr.pSubDock))
}

func (o *Icon) Window() *WindowActor {
	return NewWindowActorFromNative(unsafe.Pointer(o.Ptr.pAppli))
}

const (
	ClassCommand int = iota
	ClassName
	ClassIcon
	ClassWMClass
	ClassDesktopFile
)

// const GList *cairo_dock_get_class_menu_items (const gchar *cClass)
// const gchar **cairo_dock_get_class_mimetypes (const gchar *cClass)

func (icon *Icon) GetClassInfo(typ int) string {
	var c *C.gchar
	switch typ {
	case ClassCommand:
		c = C.cairo_dock_get_class_command(icon.Ptr.cClass)

	case ClassName:
		c = C.cairo_dock_get_class_name(icon.Ptr.cClass)

	case ClassIcon:
		c = C.cairo_dock_get_class_icon(icon.Ptr.cClass)

	// case ClassMenuItems:
	// 	c = C.cairo_dock_get_class_menu_items(icon.Ptr.cClass)

	case ClassWMClass:
		c = C.cairo_dock_get_class_wm_class(icon.Ptr.cClass)

	case ClassDesktopFile:
		c = C.cairo_dock_get_class_desktop_file(icon.Ptr.cClass)

	default:
		return ""
	}
	// defer C.free(unsafe.Pointer((*C.char)(c)))
	return C.GoString((*C.char)(c))
}

func (icon *Icon) ClassIsInhibited() bool {
	return gobool(C.cairo_dock_class_is_inhibited(icon.Ptr.cClass))
}

func (o *Icon) RemoveIconsFromSubdock(dest *CairoDock) {
	C.cairo_dock_remove_icons_from_dock(o.Ptr.pSubDock, dest.Ptr)
}

// TODO: may have to move.
func (icon *Icon) SubDockIcons() []*Icon {
	if icon.Ptr == nil || icon.Ptr.pSubDock == nil {
		return nil
	}
	clist := (*glib.List)(unsafe.Pointer(icon.Ptr.pSubDock.icons))
	return goListIcons(clist)
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
	if c == nil {
		return nil
	}
	return &Module{(*C.GldiModule)(c)}
}

func ModuleList() map[string]*Module {
	list := make(map[string]*Module)
	C.list_dock_module(C.gpointer(&listForward{list}))
	return list
}

func NewModuleFromNative(p unsafe.Pointer) *Module {
	return &Module{(*C.GldiModule)(p)}
}

func (m *Module) ToNative() unsafe.Pointer {
	return unsafe.Pointer(m.Ptr)
}

func (m *Module) IsAutoLoaded() bool {
	return gobool(C._module_is_auto_loaded(m.Ptr))
}

func (m *Module) VisitCard() *VisitCard {
	return &VisitCard{(*C.GldiVisitCard)(m.Ptr.pVisitCard)}
}

func (m *Module) InstancesList() (list []*ModuleInstance) {
	clist := (*glib.List)(unsafe.Pointer(m.Ptr.pInstancesList))
	return goListModuleInstance(clist)
}

func (m *Module) Activate() {
	C.gldi_module_activate(m.Ptr)
}

func (m *Module) AddInstance() {
	C.gldi_module_add_instance(m.Ptr)
}

//
//----------------------------------------------------------[ MODULEINSTANCE ]--

type ModuleInstance struct {
	Ptr *C.GldiModuleInstance
}

func NewModuleInstanceFromNative(p unsafe.Pointer) *ModuleInstance {
	return &ModuleInstance{(*C.GldiModuleInstance)(p)}
}

func (mi *ModuleInstance) ToNative() unsafe.Pointer {
	return unsafe.Pointer(mi.Ptr)
}

func (mi *ModuleInstance) GetConfFilePath() string {
	return C.GoString((*C.char)(mi.Ptr.cConfFilePath))
}

func (mi *ModuleInstance) Module() *Module {
	return &Module{(*C.GldiModule)(mi.Ptr.pModule)}
}

func (mi *ModuleInstance) Detach() {
	C.gldi_module_instance_detach(mi.Ptr)
}

//
//---------------------------------------------------------------[ VISITCARD ]--

type VisitCard struct {
	Ptr *C.GldiVisitCard
}

// Module name (real one)
func (vc *VisitCard) GetModuleName() string {
	return C.GoString((*C.char)(vc.Ptr.cModuleName))
}

// Module name used as identifier.
//
func (vc *VisitCard) GetName() string {
	return C.GoString((*C.char)(vc.Ptr.cModuleName))
}

// Module name translated (as seen by the user).
//
func (vc *VisitCard) GetTitle() string {
	return C.GoString((*C.char)(vc.Ptr.cTitle))
}

func (vc *VisitCard) GetIconFilePath() string {
	return C.GoString((*C.char)(vc.Ptr.cIconFilePath))
}

func (vc *VisitCard) GetAuthor() string {
	return C.GoString((*C.char)(vc.Ptr.cAuthor))
}

// GetDescription returns the module description text translated.
//
func (vc *VisitCard) GetDescription() string {
	desc := C.GoString((*C.char)(vc.Ptr.cDescription))
	return gettext.DGettext(vc.GetGettextDomain(), desc)
}

func (vc *VisitCard) GetPreviewFilePath() string {
	return C.GoString((*C.char)(vc.Ptr.cPreviewFilePath))
}

func (vc *VisitCard) GetGettextDomain() string {
	return C.GoString((*C.char)(vc.Ptr.cGettextDomain))
}

func (vc *VisitCard) GetModuleVersion() string {
	return C.GoString((*C.char)(vc.Ptr.cModuleVersion))
}

func (vc *VisitCard) GetCategory() ModuleCategory {
	return ModuleCategory(vc.Ptr.iCategory)
}

// IsMultiInstance returns whether the module can be activated multiple times or not.
//
func (vc *VisitCard) IsMultiInstance() bool {
	return gobool(vc.Ptr.bMultiInstance)
}

func (vc *VisitCard) GetContainerType() int {
	return int(vc.Ptr.iContainerType)
}

//
//---------------------------------------------------------------[ WINDOWACTOR ]--

type WindowActor struct {
	Ptr *C.GldiWindowActor
}

func NewWindowActorFromNative(p unsafe.Pointer) *WindowActor {
	return &WindowActor{(*C.GldiWindowActor)(p)}
}

func (o *WindowActor) ToNative() unsafe.Pointer {
	return unsafe.Pointer(o.Ptr)
}

func (o *WindowActor) CanMinMaxClose() (bool, bool, bool) {
	var bCanMinimize, bCanMaximize, bCanClose C.gboolean
	C.gldi_window_can_minimize_maximize_close(o.Ptr, &bCanMinimize, &bCanMaximize, &bCanClose)
	return gobool(bCanMinimize), gobool(bCanMaximize), gobool(bCanClose)
}

func (o *WindowActor) IsActive() bool {
	return o.Ptr == C.gldi_windows_get_active()
}

func (o *WindowActor) IsAbove() bool { // could split OrBelow but seem unused.
	var isAbove, isBelow C.gboolean
	C.gldi_window_is_above_or_below(o.Ptr, &isAbove, &isBelow)
	return gobool(isAbove)
}

func (o *WindowActor) IsFullScreen() bool {
	return gobool(o.Ptr.bIsFullScreen)
}

func (o *WindowActor) IsHidden() bool {
	return gobool(o.Ptr.bIsHidden)
}

func (o *WindowActor) IsMaximized() bool {
	return gobool(o.Ptr.bIsMaximized)
}

func (o *WindowActor) IsOnCurrentDesktop() bool {
	return gobool(C.gldi_window_is_on_current_desktop(o.Ptr))
}

func (o *WindowActor) IsSticky() bool {
	return gobool(C.gldi_window_is_sticky(o.Ptr))
}

func (o *WindowActor) Close() {
	C.gldi_window_close(o.Ptr)
}

func (o *WindowActor) Kill() {
	C.gldi_window_kill(o.Ptr)
}

func (o *WindowActor) Lower() {
	C.gldi_window_lower(o.Ptr)
}

func (o *WindowActor) Minimize() {
	C.gldi_window_minimize(o.Ptr)
}

func (o *WindowActor) Maximize(full bool) {
	C.gldi_window_maximize(o.Ptr, cbool(full))
}

func (o *WindowActor) MoveToCurrentDesktop() {
	C.gldi_window_move_to_current_desktop(o.Ptr)
}

func (o *WindowActor) SetAbove(above bool) {
	C.gldi_window_set_above(o.Ptr, cbool(above))
}

func (o *WindowActor) SetFullScreen(full bool) {
	C.gldi_window_set_fullscreen(o.Ptr, cbool(full))
}

func (o *WindowActor) SetSticky(sticky bool) {
	C.gldi_window_set_sticky(o.Ptr, cbool(sticky))
}

func (o *WindowActor) Show() {
	C.gldi_window_show(o.Ptr)
}

//
//---------------------------------------------------------------[ ANIMATION ]--

func AnimationList() map[string]*Animation {
	list := make(map[string]*Animation)
	C.list_animation(C.gpointer(&listForward{list}))
	return list
}

type Animation struct {
	Ptr *C.CairoDockAnimationRecord
}

func NewAnimationFromNative(p unsafe.Pointer) *Animation {
	return &Animation{(*C.CairoDockAnimationRecord)(p)}
}

func (dr *Animation) GetDisplayedName() string {
	return C.GoString((*C.char)(dr.Ptr.cDisplayedName))
}

//
//--------------------------------------------------[ CAIRODESKLETDECORATION ]--

func CairoDeskletDecorationList() map[string]*CairoDeskletDecoration {
	list := make(map[string]*CairoDeskletDecoration)
	C.list_desklet_decoration(C.gpointer(&listForward{list}))
	return list
}

type CairoDeskletDecoration struct {
	Ptr *C.CairoDeskletDecoration
}

func NewCairoDeskletDecorationFromNative(p unsafe.Pointer) *CairoDeskletDecoration {
	return &CairoDeskletDecoration{(*C.CairoDeskletDecoration)(p)}
}

func (dr *CairoDeskletDecoration) GetDisplayedName() string {
	return C.GoString((*C.char)(dr.Ptr.cDisplayedName))
}

//
//---------------------------------------------------------[ DIALOGDECORATOR ]--

func DialogDecoratorList() map[string]*DialogDecorator {
	list := make(map[string]*DialogDecorator)
	C.list_dialog_decorator(C.gpointer(&listForward{list}))
	return list
}

type DialogDecorator struct {
	Ptr *C.CairoDialogDecorator
}

func NewDialogDecoratorFromNative(p unsafe.Pointer) *DialogDecorator {
	return &DialogDecorator{(*C.CairoDialogDecorator)(p)}
}

func (dr *DialogDecorator) GetDisplayedName() string {
	return C.GoString((*C.char)(dr.Ptr.cDisplayedName))
}

// cairo_dock_foreach_dialog_decorator

//
//-------------------------------------------------------[ CAIRODOCKRENDERER ]--

func CairoDockRendererList() map[string]*CairoDockRenderer {
	list := make(map[string]*CairoDockRenderer)
	C.list_dock_renderer(C.gpointer(&listForward{list}))
	return list
}

// AKA views.
type CairoDockRenderer struct {
	Ptr *C.CairoDockRenderer
}

func NewCairoDockRendererFromNative(p unsafe.Pointer) *CairoDockRenderer {
	return &CairoDockRenderer{(*C.CairoDockRenderer)(p)}
}

func (dr *CairoDockRenderer) GetDisplayedName() string {
	return C.GoString((*C.char)(dr.Ptr.cDisplayedName))
}

//
//-----------------------------------------------------------------[ MANAGER ]--

type Manager struct {
	Ptr *C.GldiManager
}

func NewManagerFromNative(p unsafe.Pointer) *Manager {
	return &Manager{(*C.GldiManager)(p)}
}

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

type DesktopGeometry struct {
	Obj C.GldiDesktopGeometry
}

func GetDesktopGeometry() *DesktopGeometry {
	return &DesktopGeometry{C.g_desktopGeometry}
}

func (dg *DesktopGeometry) NbScreens() int {
	return int(dg.Obj.iNbScreens)
}

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

// ListShortkeys returns the list of dock shortkeys.
//
func ShortkeyList() []*Shortkey {
	list := &listForward{[]*Shortkey{}}
	C.list_shortkey(C.gpointer(list))
	return list.p.([]*Shortkey)
}

//export addShortkey
func addShortkey(sk C.gpointer, l C.gpointer) {
	list := (*listForward)(l)
	list.p = append(list.p.([]*Shortkey), NewShortkeyFromNative(unsafe.Pointer(sk)))
}

type Shortkey struct {
	Ptr *C.GldiShortkey
}

func NewShortkeyFromNative(p unsafe.Pointer) *Shortkey {
	return &Shortkey{(*C.GldiShortkey)(p)}
}

func (dr *Shortkey) GetDemander() string {
	return C.GoString((*C.char)(dr.Ptr.cDemander))
}

func (dr *Shortkey) GetDescription() string {
	return C.GoString((*C.char)(dr.Ptr.cDescription))
}

// GetKeyString returns the shortkey key reference as string.
//
func (dr *Shortkey) GetKeyString() string {
	return C.GoString((*C.char)(dr.Ptr.keystring))
}

func (dr *Shortkey) GetIconFilePath() string {
	return C.GoString((*C.char)(dr.Ptr.cIconFilePath))
}

func (dr *Shortkey) GetConfFilePath() string {
	return C.GoString((*C.char)(dr.Ptr.cConfFilePath))
}

func (dr *Shortkey) GetGroupName() string {
	return C.GoString((*C.char)(dr.Ptr.cGroupName))
}

// GetKeyName returns the config key name.
//
func (dr *Shortkey) GetKeyName() string {
	return C.GoString((*C.char)(dr.Ptr.cKeyName))
}

func (dr *Shortkey) GetSuccess() bool {
	return gobool(dr.Ptr.bSuccess)
}

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
	return (*C.gchar)(C.CString(str))
}

// List forwarding from c callbacks.
type listForward struct{ p interface{} }

//export addItemToList
func addItemToList(p C.gpointer, cstr *C.gchar, cdr C.gpointer) {
	name := C.GoString((*C.char)(cstr))
	free := true

	list := (*listForward)(p)
	switch v := list.p.(type) {
	case map[string]*Animation:
		v[name] = NewAnimationFromNative(unsafe.Pointer(cdr))

	case map[string]*DialogDecorator:
		v[name] = NewDialogDecoratorFromNative(unsafe.Pointer(cdr))

	case map[string]*CairoDeskletDecoration:
		v[name] = NewCairoDeskletDecorationFromNative(unsafe.Pointer(cdr))

	case map[string]*Module:
		v[name] = NewModuleFromNative(unsafe.Pointer(cdr))
		free = false

	case map[string]*CairoDockRenderer:
		v[name] = NewCairoDockRendererFromNative(unsafe.Pointer(cdr))
	}

	if free {
		C.free(unsafe.Pointer((*C.char)(cstr)))
	}

}

func goListDocks(clist *glib.List) (list []*CairoDock) {
	for _ = range iter.N(int(clist.Length())) {
		item := NewDockFromNative(unsafe.Pointer(clist.Data))
		list = append(list, item)
		clist = clist.Next
	}
	return
}

func goListIcons(clist *glib.List) (list []*Icon) {
	for _ = range iter.N(int(clist.Length())) {
		item := NewIconFromNative(unsafe.Pointer(clist.Data))
		list = append(list, item)
		clist = clist.Next
	}
	return
}

func goListModuleInstance(clist *glib.List) (list []*ModuleInstance) {
	for _ = range iter.N(int(clist.Length())) {
		item := NewModuleInstanceFromNative(unsafe.Pointer(clist.Data))
		list = append(list, item)
		clist = clist.Next
	}
	return
}
