// Package backendgui provides a GUI interface to interact with the dock GUI.
package backendgui

// #cgo pkg-config: gldi
// #include "cairo-dock-gui-manager.h"              // CairoDockGuiBackend
/*

//----------------------------------------------[ GUI CALLBACK NOTIFICATIONS ]--

// Go exported func redeclarations for notifications.

extern gboolean notifIconMoved        (gpointer pUserData, Icon *pIcon, CairoDock *pDock);
extern gboolean notifIconAddRemove    (gpointer pUserData, Icon *pIcon, CairoDock *pDock);
extern gboolean notifReloadItems      (gpointer pUserData, gpointer unused);
extern gboolean notifConfigureDesklet (G_GNUC_UNUSED gpointer pUserData, CairoDesklet *pDesklet);
extern gboolean notifModuleActivated  (G_GNUC_UNUSED gpointer pUserData, gchar *cModuleName, gboolean bActivated);
extern gboolean notifModuleRegistered (G_GNUC_UNUSED gpointer pUserData, gchar *cModuleName, gboolean bActivated);
extern gboolean notifModuleInstanceDetached (G_GNUC_UNUSED gpointer pUserData, GldiModuleInstance *pInstance, gboolean bIsDetached);
extern gboolean notifShortkeyUpdate   (G_GNUC_UNUSED gpointer pUserData, G_GNUC_UNUSED gpointer shortkey);

//-----------------------------------------------[ GUI CALLBACK CORE BACKEND ]--

// Go exported func redeclarations for GUI backend.

extern void   showModuleInstanceGui (GldiModuleInstance *pModuleInstance, int iShowPage);
extern void   setStatusMessage      (gchar *cMessage);
extern void   reloadCurrentWidget   (GldiModuleInstance *pInstance, int iShowPage);


// Wrappers around calls impossible from C to Go (const gchar* -> gchar*)

static void  _setStatusMessage  (const gchar *message)  { setStatusMessage(g_strdup(message)); }



static void register_gui (void)
{
 	CairoDockGuiBackend *pConfigBackend = g_new0 (CairoDockGuiBackend, 1);

	pConfigBackend->set_status_message_on_gui 	= _setStatusMessage;
	pConfigBackend->reload_current_widget 		= reloadCurrentWidget;
 	pConfigBackend->show_module_instance_gui 	= showModuleInstanceGui;
// 	pConfigBackend->get_widget_from_name 		= get_widget_from_name;

 	cairo_dock_register_gui_backend (pConfigBackend);
}


*/
import "C"
import (
	"github.com/gotk3/gotk3/gtk"

	"github.com/sqp/godock/libs/gldi"
	"github.com/sqp/godock/libs/gldi/globals"

	"unsafe"
)

//
//-----------------------------------------------------------[ GUI CALLBACKS ]--

// GuiInterface defines the interface to the gldi GUI backend.
//
type GuiInterface interface {
	ShowMainGui()                                                     //
	ShowModuleGui(appletName string)                                  //
	ShowItems(*gldi.Icon, *gldi.Container, *gldi.ModuleInstance, int) //

	ShowAddons()
	ReloadItems()
	// ReloadCategoryWidget()
	// Reload()
	CloseGui()

	UpdateModulesList()
	UpdateModuleState(name string, active bool)
	UpdateModuleInstanceContainer(instance *gldi.ModuleInstance, detached bool)
	UpdateShortkeys()
	UpdateDeskletParams(*gldi.Desklet)
	UpdateDeskletVisibility(*gldi.Desklet)

	// CORE BACKEND
	SetStatusMessage(message string)
	ReloadCurrentWidget(moduleInstance *gldi.ModuleInstance, showPage int)
	ShowModuleInstanceGui(*gldi.ModuleInstance, int) //
	// GetWidgetFromName(moduleInstance *gldi.ModuleInstance, group string, key string)

	Window() *gtk.Window
}

var dockGui GuiInterface

// Register registers a GUI to the backend, allowing it to receive GUI events.
//
func Register(gui GuiInterface) {
	C.register_gui()
	dockGui = gui

	registerEvents()
}

// CanManageThemes returns if the backend can configure themes.
//
// Unused yet (but should be possible to activate now).
//
func CanManageThemes() bool {
	return false // TODO.
}

//
//-----------------------------------------------------------[ NOTIFICATIONS ]--

func registerEvents() {
	// Dock manager
	globals.DockObjectMgr.RegisterNotification(
		globals.NotifIconMoved,
		unsafe.Pointer(C.notifIconMoved),
		globals.RunAfter)

	globals.DockObjectMgr.RegisterNotification(
		globals.NotifInsertIcon,
		unsafe.Pointer(C.notifIconAddRemove),
		globals.RunAfter)

	globals.DockObjectMgr.RegisterNotification(
		globals.NotifRemoveIcon,
		unsafe.Pointer(C.notifIconAddRemove),
		globals.RunAfter)

	globals.DockObjectMgr.RegisterNotification(
		globals.NotifDestroy,
		unsafe.Pointer(C.notifReloadItems),
		globals.RunAfter)

	// Desklet manager.
	globals.DeskletObjectMgr.RegisterNotification(
		globals.NotifDestroy,
		unsafe.Pointer(C.notifReloadItems),
		globals.RunAfter)

	globals.DeskletObjectMgr.RegisterNotification(
		globals.NotifNew,
		unsafe.Pointer(C.notifReloadItems),
		globals.RunAfter)

	globals.DeskletObjectMgr.RegisterNotification(
		globals.NotifConfigureDesklet,
		unsafe.Pointer(C.notifConfigureDesklet),
		globals.RunAfter)

	// Module manager.
	globals.ModuleObjectMgr.RegisterNotification(
		globals.NotifModuleActivated,
		unsafe.Pointer(C.notifModuleActivated),
		globals.RunAfter)

	globals.ModuleObjectMgr.RegisterNotification(
		globals.NotifModuleRegistered,
		unsafe.Pointer(C.notifModuleRegistered),
		globals.RunAfter)

	// Module instance manager.
	globals.ModuleInstanceObjectMgr.RegisterNotification(
		globals.NotifModuleInstanceDetached,
		unsafe.Pointer(C.notifModuleInstanceDetached),
		globals.RunAfter)

	// Shortkey manager.
	globals.ShortkeyObjectMgr.RegisterNotification(
		globals.NotifNew,
		unsafe.Pointer(C.notifShortkeyUpdate),
		globals.RunAfter)

	globals.ShortkeyObjectMgr.RegisterNotification(
		globals.NotifDestroy,
		unsafe.Pointer(C.notifShortkeyUpdate),
		globals.RunAfter)

	globals.ShortkeyObjectMgr.RegisterNotification(
		globals.NotifShortkeyChanged,
		unsafe.Pointer(C.notifShortkeyUpdate),
		globals.RunAfter)

}

// TODO: maybe need to use the next func, I copied the exact behavior but need to report to confirm this func.
//export notifIconMoved
func notifIconMoved(_ C.gpointer, cicon *C.Icon, _ *C.CairoDock) C.gboolean {
	icon := gldi.NewIconFromNative(unsafe.Pointer(cicon))
	if (icon.IsLauncher() || icon.IsStackIcon() || icon.IsSeparator() && icon.GetDesktopFileName() != "") || icon.IsApplet() {
		ReloadItems()
	}
	return C.GLDI_NOTIFICATION_LET_PASS
}

//export notifIconAddRemove
func notifIconAddRemove(_ C.gpointer, cicon *C.Icon, _ *C.CairoDock) C.gboolean {
	icon := gldi.NewIconFromNative(unsafe.Pointer(cicon))
	if ((icon.IsLauncher() || icon.IsStackIcon() || icon.IsSeparator()) && icon.GetDesktopFileName() != "") || icon.IsApplet() {
		ReloadItems()
	}
	return C.GLDI_NOTIFICATION_LET_PASS
}

//export notifReloadItems
func notifReloadItems(_ C.gpointer, _ C.gpointer) C.gboolean {
	ReloadItems()
	return C.GLDI_NOTIFICATION_LET_PASS
}

//export notifConfigureDesklet
func notifConfigureDesklet(_ C.gpointer, cdesklet *C.CairoDesklet) C.gboolean {
	desklet := gldi.NewDeskletFromNative(unsafe.Pointer(cdesklet))
	UpdateDeskletParams(desklet)
	return C.GLDI_NOTIFICATION_LET_PASS
}

//export notifModuleActivated
func notifModuleActivated(_ C.gpointer, cModuleName *C.gchar, active C.gboolean) C.gboolean {
	name := C.GoString((*C.char)(cModuleName))
	UpdateModuleState(name, gobool(active))

	ReloadItems() // for plug-ins that don't have an applet, like Cairo-Pinguin.
	return C.GLDI_NOTIFICATION_LET_PASS
}

//export notifModuleRegistered
func notifModuleRegistered(_ C.gpointer, _ *C.gchar, _ C.gboolean) C.gboolean {
	UpdateModulesList()
	return C.GLDI_NOTIFICATION_LET_PASS
}

//export notifModuleInstanceDetached
func notifModuleInstanceDetached(_ C.gpointer, instance *C.GldiModuleInstance, isDetached C.gboolean) C.gboolean {
	mi := gldi.NewModuleInstanceFromNative(unsafe.Pointer(instance))

	UpdateModuleInstanceContainer(mi, gobool(isDetached))
	ReloadItems()
	return C.GLDI_NOTIFICATION_LET_PASS
}

//export notifShortkeyUpdate
func notifShortkeyUpdate(_ C.gpointer, _ C.gpointer) C.gboolean {
	UpdateShortkeys()
	return C.GLDI_NOTIFICATION_LET_PASS
}

//
//-----------------------------------------------------------------[ FROM GO ]--

// ShowMainGui shows the main config page of the GUI.
//
func ShowMainGui() *gtk.Window {
	if dockGui == nil {
		return nil
	}
	dockGui.ShowMainGui()
	return dockGui.Window()
}

// ShowModuleGui opens the icons page of the GUI for the specific applet.
//
func ShowModuleGui(appletName string) *gtk.Window {
	if dockGui == nil {
		return nil
	}
	dockGui.ShowModuleGui(appletName)
	return dockGui.Window()
}

// ShowItems opens the icons page of the GUI to configure the given item.
//
func ShowItems(icon *gldi.Icon, container *gldi.Container, moduleInstance *gldi.ModuleInstance, showPage int) *gtk.Window {
	if dockGui == nil {
		return nil
	}
	dockGui.ShowItems(icon, container, moduleInstance, showPage)
	return dockGui.Window()
}

// ShowAddons opens the addons page of the GUI.
//
func ShowAddons() *gtk.Window {
	if dockGui == nil {
		return nil
	}
	dockGui.ShowAddons()
	return dockGui.Window()
}

// ReloadItems reloads the items page.
//
//export ReloadItems
func ReloadItems() {
	if dockGui == nil {
		return
	}
	dockGui.ReloadItems()
}

// UpdateModulesList updates the list of applets.
//
//export UpdateModulesList
func UpdateModulesList() {
	if dockGui == nil {
		return
	}
	dockGui.UpdateModulesList()
}

// UpdateModuleState updates the state of an applet.
//
func UpdateModuleState(name string, active bool) {
	if dockGui == nil {
		return
	}
	dockGui.UpdateModuleState(name, active)
}

// UpdateModuleInstanceContainer updates the container widget of an applet.
//
func UpdateModuleInstanceContainer(moduleInstance *gldi.ModuleInstance, detached bool) {
	if dockGui == nil {
		return
	}
	dockGui.UpdateModuleInstanceContainer(moduleInstance, detached)
}

// UpdateShortkeys updates the shortkeys list.
//
//export UpdateShortkeys
func UpdateShortkeys() {
	if dockGui == nil {
		return
	}
	dockGui.UpdateShortkeys()
}

// UpdateDeskletParams updates desklets params of an applet.
//
func UpdateDeskletParams(desklet *gldi.Desklet) {
	if dockGui == nil {
		return
	}
	dockGui.UpdateDeskletParams(desklet)
}

// UpdateDeskletVisibility updates the desklet visibility widget of an applet.
//
func UpdateDeskletVisibility(desklet *gldi.Desklet) {
	if dockGui == nil {
		return
	}
	dockGui.UpdateDeskletVisibility(desklet)
}

// CORE BACKEND

// ShowModuleInstanceGui opens the icons page of the GUI to configure the given applet.
//
func ShowModuleInstanceGui(moduleInstance *gldi.ModuleInstance, showPage int) {
	if dockGui == nil {
		return
	}
	dockGui.ShowModuleInstanceGui(moduleInstance, showPage)

}

// SetStatusMessage is unused.
//
func SetStatusMessage(message string) {
	if dockGui == nil {
		return
	}
	dockGui.SetStatusMessage(message)
}

// ReloadCurrentWidget reloads the current widget page.
//
func ReloadCurrentWidget(moduleInstance *gldi.ModuleInstance, showPage int) {
	if dockGui == nil {
		return
	}
	dockGui.ReloadCurrentWidget(moduleInstance, showPage)
}

// CloseGui closes the configuration window.
//
func CloseGui() {
	if dockGui == nil {
		return
	}
	dockGui.CloseGui()
}

//
//-----------------------------------------------------[ CORE BACKEND FROM C ]--

//export showModuleInstanceGui
func showModuleInstanceGui(moduleInstance *C.GldiModuleInstance, showPage C.int) {
	m := gldi.NewModuleInstanceFromNative(unsafe.Pointer(moduleInstance))
	ShowModuleInstanceGui(m, int(showPage))

}

//export setStatusMessage
func setStatusMessage(cmessage *C.gchar) {
	message := C.GoString((*C.char)(cmessage))
	C.g_free(C.gpointer(cmessage))
	SetStatusMessage(message)
}

//export reloadCurrentWidget
func reloadCurrentWidget(moduleInstance *C.GldiModuleInstance, showPage C.int) {
	m := gldi.NewModuleInstanceFromNative(unsafe.Pointer(moduleInstance))
	ReloadCurrentWidget(m, int(showPage))
}

func gobool(b C.gboolean) bool {
	if b == 1 {
		return true
	}
	return false
}
