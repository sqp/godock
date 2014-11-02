package maindock

// #include "cairo-dock-gui-manager.h"              // CairoDockGuiBackend
// #include "cairo-dock-gui-backend.h"              // local file
/*

//
//----------------------------------------------------[ GUI CALLBACK BACKEND ]--

// Go exported func redeclarations.

extern GtkWidget* showMainGui                   ();
extern GtkWidget* showModuleGui                 (gchar *cModuleName);
extern GtkWidget* showGui                       (Icon *pIcon, GldiContainer *pContainer, GldiModuleInstance *pModuleInstance, int iShowPage);
extern GtkWidget* showAddons                    ();

extern void       reloadItems                   ();
extern void       reload                        ();
extern void       closeGui                      ();
extern void       updateModulesList             ();
extern void       updateModuleState             (gchar*, gboolean);
extern void       updateModuleInstanceContainer (GldiModuleInstance*, gboolean);
extern void       updateShortkeys               ();

// GUI CORE BACKEND
extern void       showModuleInstanceGui         (GldiModuleInstance *pModuleInstance, int iShowPage);
extern void       setStatusMessage              (gchar *cMessage);
extern void       reloadCurrentWidget           (GldiModuleInstance *pInstance, int iShowPage);


// Wrappers around calls impossible from C to Go (const gchar* -> gchar*)

static void       _setStatusMessage  (const gchar *message)                       { setStatusMessage(g_strdup(message)); }
static GtkWidget* _showModuleGui     (const gchar *cModuleName)                   { return showModuleGui(g_strdup(cModuleName)); }
static void       _updateModuleState (const gchar *cModuleName, gboolean bActive) { updateModuleState(g_strdup(cModuleName), bActive); }


static void register_gui (void)
{
	CairoDockMainGuiBackend *pBackend = g_new0 (CairoDockMainGuiBackend, 1);
	cairo_dock_register_config_gui_backend (pBackend);

 	pBackend->show_main_gui 					= showMainGui;
 	pBackend->show_module_gui 					= _showModuleGui;
 	pBackend->show_gui 							= showGui;

	pBackend->close_gui 						= closeGui;
	pBackend->update_module_state 				= _updateModuleState;
	pBackend->update_module_instance_container 	= updateModuleInstanceContainer;
// 	pBackend->update_desklet_params 			= update_desklet_params;
// 	pBackend->update_desklet_visibility_params 	= update_desklet_visibility_params;
	pBackend->update_modules_list 				= updateModulesList;
	pBackend->update_shortkeys 					= updateShortkeys;
	pBackend->show_addons 						= showAddons;
	pBackend->reload_items 						= reloadItems;
	pBackend->reload 							= reload;
// 	pBackend->cDisplayedName 					= _("Advanced Mode");  // name of the other backend.
// 	pBackend->cTooltip 							= _("The advanced mode lets you tweak every single parameter of the dock. It is a powerful tool to customise your current theme.");


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
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/godock/libs/gldi"

	// "github.com/sqp/godock/libs/maindock/gui"

	"unsafe"
)

//
//-----------------------------------------------------------[ GUI CALLBACKS ]--

// GuiInterface defines the interface to the gldi GUI backend.
//
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

var dockGui GuiInterface

// RegisterGui registers a GUI to the backend, allowing it to receive GUI events.
//
func RegisterGui(gui GuiInterface) {
	C.register_gui()
	dockGui = gui
}

//export showMainGui
func showMainGui() *C.GtkWidget {
	if dockGui == nil {
		return nil
	}
	dockGui.ShowMainGui()
	return toCWindow(dockGui.Window())
}

//export showModuleGui
func showModuleGui(cModuleName *C.gchar) *C.GtkWidget {
	name := C.GoString((*C.char)(cModuleName))
	C.g_free(C.gpointer(cModuleName))

	if dockGui == nil {
		return nil
	}
	dockGui.ShowModuleGui(name)
	return toCWindow(dockGui.Window())
}

//export showGui
func showGui(icon *C.Icon, container *C.GldiContainer, moduleInstance *C.GldiModuleInstance, iShowPage C.int) *C.GtkWidget {
	if dockGui == nil {
		return nil
	}
	i := gldi.NewIconFromNative(unsafe.Pointer(icon))
	c := gldi.NewContainerFromNative(unsafe.Pointer(container))
	m := gldi.NewModuleInstanceFromNative(unsafe.Pointer(moduleInstance))

	dockGui.ShowGui(i, c, m, int(iShowPage))
	return toCWindow(dockGui.Window())
}

//export showAddons
func showAddons() *C.GtkWidget {
	if dockGui == nil {
		return nil
	}
	dockGui.ShowAddons()
	return toCWindow(dockGui.Window())
}

//export reloadItems
func reloadItems() {
	if dockGui == nil {
		return
	}
	dockGui.ReloadItems()
}

//export reload
func reload() {
	if dockGui == nil {
		return
	}
	dockGui.Reload()
}

//export closeGui
func closeGui() {
	if dockGui == nil {
		return
	}
	dockGui.Close()
}

//export updateModulesList
func updateModulesList() {
	if dockGui == nil {
		return
	}
	dockGui.UpdateModulesList()
}

//export updateModuleState
func updateModuleState(cModuleName *C.gchar, active C.gboolean) {
	name := C.GoString((*C.char)(cModuleName))
	C.g_free(C.gpointer(cModuleName))

	if dockGui == nil {
		return
	}
	dockGui.UpdateModuleState(name, gobool(active))
}

//export updateModuleInstanceContainer
func updateModuleInstanceContainer(moduleInstance *C.GldiModuleInstance, detached C.gboolean) {
	if dockGui == nil {
		return
	}
	m := gldi.NewModuleInstanceFromNative(unsafe.Pointer(moduleInstance))
	dockGui.UpdateModuleInstanceContainer(m, gobool(detached))
}

//export updateShortkeys
func updateShortkeys() {
	if dockGui == nil {
		return
	}
	dockGui.UpdateShortkeys()
}

// CORE BACKEND

//export showModuleInstanceGui
func showModuleInstanceGui(moduleInstance *C.GldiModuleInstance, showPage C.int) {
	if dockGui == nil {
		return
	}
	m := gldi.NewModuleInstanceFromNative(unsafe.Pointer(moduleInstance))
	dockGui.ShowModuleInstanceGui(m, int(showPage))

}

//export setStatusMessage
func setStatusMessage(cmessage *C.gchar) {
	message := C.GoString((*C.char)(cmessage))
	C.g_free(C.gpointer(cmessage))

	if dockGui == nil {
		return
	}
	dockGui.SetStatusMessage(message)
}

//export reloadCurrentWidget
func reloadCurrentWidget(moduleInstance *C.GldiModuleInstance, showPage C.int) {
	if dockGui == nil {
		return
	}
	m := gldi.NewModuleInstanceFromNative(unsafe.Pointer(moduleInstance))
	dockGui.ReloadCurrentWidget(m, int(showPage))
}

// //export GetWidgetFromName
// func GetWidgetFromName(moduleInstance *C.GldiModuleInstance, cgroup *C.gchar, ckey *C.gchar) {
// 	group := C.GoString((*C.char)(cgroup))
// 	C.free(unsafe.Pointer((*C.char)(cgroup)))
// 	key := C.GoString((*C.char)(ckey))
// 	C.free(unsafe.Pointer((*C.char)(ckey)))

// 	if dockGui == nil {
// 		return
// 	}
// 	m := gldi.NewModuleInstanceFromNative(unsafe.Pointer(moduleInstance))
// 	dockGui.GetWidgetFromName(m, group, key)
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
