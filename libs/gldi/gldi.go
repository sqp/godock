// Package gldi provides Go bindings for gldi (cairo-dock).  Supports version 3.4
package gldi

/*
#cgo pkg-config: glib-2.0 gldi
#include <stdlib.h>                              // free
#include <dbus/dbus-glib.h>                      // dbus_g_thread_init
#include <glib/gkeyfile.h>                       // GKeyFile

#include "cairo-dock-core.h"
#include "cairo-dock-backends-manager.h"         // cairo_dock_foreach_dock_renderer
#include "cairo-dock-config.h"                   // cairo_dock_load_current_theme
#include "cairo-dock-class-manager.h"            // cairo_dock_get_class_command
#include "cairo-dock-dock-factory.h"             // CairoDock
#include "cairo-dock-dock-facility.h"            // cairo_dock_get_available_docks
#include "cairo-dock-dock-manager.h"             // gldi_dock_get_readable_name
#include "cairo-dock-file-manager.h"             // CAIRO_DOCK_GNOME...
#include "cairo-dock-icon-factory.h"             // Icon
#include "cairo-dock-icon-facility.h"        // Icon
#include "cairo-dock-launcher-manager.h"         // CAIRO_DOCK_ICON_TYPE_IS_LAUNCHER
#include "cairo-dock-log.h"                      // cd_log_set_level_from_name
#include "cairo-dock-module-instance-manager.h"  // ModuleInstance
#include "cairo-dock-module-manager.h"           // gldi_modules_new_from_directory
#include "cairo-dock-object.h"                   // Icon
#include "cairo-dock-opengl.h"                   // gldi_gl_backend_force_indirect_rendering
#include "cairo-dock-separator-manager.h"        // CAIRO_DOCK_ICON_TYPE_IS_SEPARATOR
#include "cairo-dock-stack-icon-manager.h"       // CAIRO_DOCK_ICON_TYPE_IS_CONTAINER
#include "cairo-dock-themes-manager.h"           // cairo_dock_set_paths

extern CairoDock *g_pMainDock;

extern CairoDockGLConfig g_openglConfig;
extern gboolean          g_bUseOpenGL;

extern gchar *g_cCurrentLaunchersPath;


static gboolean IconIsSeparator    (Icon *icon) { return CAIRO_DOCK_ICON_TYPE_IS_SEPARATOR(icon); }
static gboolean IconIsSeparatorAuto(Icon *icon) { return CAIRO_DOCK_IS_AUTOMATIC_SEPARATOR(icon); }
static gboolean IconIsLauncher     (Icon *icon) { return CAIRO_DOCK_ICON_TYPE_IS_LAUNCHER(icon); }
static gboolean IconIsStackIcon    (Icon *icon) { return CAIRO_DOCK_ICON_TYPE_IS_CONTAINER(icon); }

// static void icon_reload(Icon *icon) { gldi_object_reload (GLDI_OBJECT(ic), TRUE);}



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

extern void addItemToList (gpointer, gchar*, gpointer);

static void     fwd_one (const gchar *name, gpointer *item, gpointer p) { addItemToList(p, g_strdup(name), item); }
static gboolean fwd_chk (      gchar *name, gpointer *item, gpointer p) { addItemToList(p, name,           item); return FALSE;}


static void list_dock_module        (gpointer p){ gldi_module_foreach                   ((GHRFunc)fwd_chk, p); }

static void list_animation          (gpointer p){ cairo_dock_foreach_animation          ((GHFunc)fwd_one, p); }
static void list_desklet_decoration (gpointer p){ cairo_dock_foreach_desklet_decoration ((GHFunc)fwd_one, p); }
static void list_dialog_decorator   (gpointer p){ cairo_dock_foreach_dialog_decorator   ((GHFunc)fwd_one, p); }
static void list_dock_renderer      (gpointer p){ cairo_dock_foreach_dock_renderer      ((GHFunc)fwd_one, p); }




*/
import "C"

import (
	"github.com/bradfitz/iter"
	"github.com/conformal/gotk3/glib"
	"github.com/gosexy/gettext"

	"github.com/sqp/godock/widgets/gtk/keyfile"

	"errors"
	"path/filepath"
	"unsafe"
)

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
func GetAllAvailableDocks(parent, subdock *CairoDock) (list []*CairoDock) {
	var cp, cs *C.CairoDock
	if parent != nil {
		cp = parent.Ptr
	}
	if subdock != nil {
		cs = subdock.Ptr
	}

	clist := (*glib.List)(unsafe.Pointer(C.cairo_dock_get_available_docks(cp, cs)))

	for i := range iter.N(int(clist.Length())) {
		_ = i
		item := NewDockFromNative(unsafe.Pointer(clist.Data))
		list = append(list, item)
		clist = clist.Next
	}
	return
}

func DockGet(containerName string) *CairoDock {
	cstr := (*C.gchar)(C.CString(containerName))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	return NewDockFromNative(unsafe.Pointer(C.gldi_dock_get(cstr)))
}

type IObject interface {
	ToNative() unsafe.Pointer
}

func ObjectReload(obj IObject) {
	C.gldi_object_reload((*C.GldiObject)(obj.ToNative()), C.gboolean(1))
}

func ObjectIsManagerChild(obj IObject, ptr *C.GldiObjectManager) bool {
	return gobool(C.gldi_object_is_manager_child((*C.GldiObject)(obj.ToNative()), ptr))
}

func ObjectIsDock(obj IObject) bool {
	return ObjectIsManagerChild(obj, &C.myDockObjectMgr)
}

func DialogShowTemporaryWithIcon(str string, icon *Icon, container *Container, duration float64, iconPath string) {
	cstr := (*C.gchar)(C.CString(str))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	cpath := (*C.gchar)(C.CString(iconPath))
	defer C.free(unsafe.Pointer((*C.char)(cpath)))

	C.gldi_dialog_show_temporary_with_icon(cstr, icon.Ptr, container.Ptr, C.double(duration), cpath)
}

//
//---------------------------------------------------------------[ CAIRODOCK ]--

type CairoDock struct {
	Ptr *C.CairoDock
}

func NewDockFromNative(p unsafe.Pointer) *CairoDock {
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

	for i := range iter.N(int(clist.Length())) {
		_ = i
		item := NewIconFromNative(unsafe.Pointer(clist.Data))
		list = append(list, item)
		clist = clist.Next
	}
	return
}

func (dock *CairoDock) GetDockName() string {
	return C.GoString((*C.char)(dock.Ptr.cDockName))
}

func (dock *CairoDock) GetReadableName() string {
	return C.GoString((*C.char)(C.gldi_dock_get_readable_name(dock.Ptr)))
}

func (dock *CairoDock) IsMainDock() bool {
	return gobool(dock.Ptr.bIsMainDock)
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

//
//--------------------------------------------------------------------[ ICON ]--

type Icon struct {
	Ptr *C.Icon
}

func NewIconFromNative(p unsafe.Pointer) *Icon {
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

// func (icon *Icon) GetIconType() int {
// 	return int(C.cairo_dock_get_icon_type(icon.Ptr))
// }

func (icon *Icon) IsApplet() bool {
	return icon.Ptr != nil && icon.Ptr.pModuleInstance != nil
}

// IsTaskbar returns whether the icon belongs to the taskbar or not.
//
func (icon *Icon) IsTaskbar() bool { // IS_APPLI
	return icon.Ptr != nil && !icon.IsLauncher() && icon.Ptr.pAppli != nil
}

// #define CAIRO_DOCK_IS_APPLI(icon) (icon != NULL && (icon)->pAppli != NULL)
// ./../Cairo-Dock/cairo-dock-core/src/gldit/cairo-dock-icon-factory.h:#define CAIRO_DOCK_IS_APPLI(icon) (icon != NULL && (icon)->pAppli != NULL)

func (icon *Icon) IsSeparator() bool {
	return gobool(C.IconIsSeparator(icon.Ptr))
}
func (icon *Icon) IsSeparatorAuto() bool {
	return gobool(C.IconIsSeparatorAuto(icon.Ptr))
}

func (icon *Icon) IsLauncher() bool {
	return gobool(C.IconIsLauncher(icon.Ptr))
}

func (icon *Icon) IsStackIcon() bool {
	return gobool(C.IconIsStackIcon(icon.Ptr))
}

// cairo-dock-core/src/gldit/cairo-dock-icon-factory.h
// #define CAIRO_DOCK_IS_AUTOMATIC_SEPARATOR(icon) (CAIRO_DOCK_ICON_TYPE_IS_SEPARATOR (icon) && (icon)->cDesktopFileName == NULL)
// #define CAIRO_DOCK_IS_USER_SEPARATOR(icon) (CAIRO_DOCK_ICON_TYPE_IS_SEPARATOR (icon) && (icon)->cDesktopFileName != NULL)

func (icon *Icon) ModuleInstance() *ModuleInstance {
	if !icon.IsApplet() {
		return nil
	}
	return &ModuleInstance{icon.Ptr.pModuleInstance}
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

	for i := range iter.N(int(clist.Length())) {
		_ = i
		item := NewModuleInstanceFromNative(unsafe.Pointer(clist.Data))
		list = append(list, item)
		clist = clist.Next
	}
	return
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
func ManagerReload(name string, b bool, keyf keyfile.KeyFile) {
	manager := ManagerGet(name)
	if manager == nil {
		return
	}
	C.manager_reload(manager.Ptr, cbool(b), (*C.GKeyFile)(unsafe.Pointer(keyf.ToNative())))
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
