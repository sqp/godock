// Package backendgui provides a GUI interface to interact with the dock GUI.
package backendgui

// #cgo pkg-config: gldi
// #include "cairo-dock-gui-manager.h"              // CairoDockGuiBackend
/*

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
	"github.com/sqp/godock/libs/gldi/notif" // Dock notifs.

	"unsafe"
)

//
//-----------------------------------------------------------[ GUI CALLBACKS ]--

// GuiInterface defines the interface to the gldi GUI backend.
//
type GuiInterface interface {
	ShowMainGui()                                                    //
	ShowModuleGui(appletName string)                                 //
	ShowItems(gldi.Icon, *gldi.Container, *gldi.ModuleInstance, int) //

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

	notif.RegisterDockIconMoved(onIconMoved)
	notif.RegisterDockInsertIcon(onIconAddRemove)
	notif.RegisterDockRemoveIcon(onIconAddRemove)
	notif.RegisterDockDestroy(func(*gldi.CairoDock) { ReloadItems() })

	notif.RegisterDeskletNew(func(*gldi.Desklet) { ReloadItems() })
	notif.RegisterDeskletDestroy(func(*gldi.Desklet) { ReloadItems() })
	notif.RegisterDeskletConfigure(UpdateDeskletParams)

	notif.RegisterModuleRegistered(func(string, bool) { UpdateModulesList() })
	notif.RegisterModuleActivated(func(name string, active bool) {
		UpdateModuleState(name, active)
		ReloadItems() // for plug-ins that don't have an applet, like Cairo-Pinguin.
	})

	notif.RegisterModuleInstanceDetached(func(mi *gldi.ModuleInstance, isDetached bool) {
		UpdateModuleInstanceContainer(mi, isDetached)
		ReloadItems()
	})

	notif.RegisterShortkeyChanged(UpdateShortkeys)
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

// TODO: maybe need to use the next func, I copied the exact behavior but need to report to confirm this func.
func onIconMoved(icon gldi.Icon, _ *gldi.CairoDock) {
	if icon.IsApplet() ||
		((icon.IsLauncher() || icon.IsStackIcon() || icon.IsSeparator()) && icon.GetDesktopFileName() != "") {
		ReloadItems()
	}
}

func onIconAddRemove(icon gldi.Icon, _ *gldi.CairoDock) {
	if icon.IsApplet() ||
		((icon.IsLauncher() || icon.IsStackIcon() || icon.IsSeparator()) && icon.GetDesktopFileName() != "") {
		ReloadItems()
	}
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
func ShowItems(icon gldi.Icon, container *gldi.Container, moduleInstance *gldi.ModuleInstance, showPage int) *gtk.Window {
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
