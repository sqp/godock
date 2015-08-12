/*
Package guibridge provides a bridge between the dock and its config window.

It can create the config widget and window connected to the data source, and
forward dock events to update the GUI.

Implements backendgui.GuiInterface.
*/
package guibridge

import (
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/godock/libs/cdtype" // Logger type.
	"github.com/sqp/godock/libs/gldi"
	"github.com/sqp/godock/libs/gldi/confdata"
	"github.com/sqp/godock/libs/gldi/globals"

	"github.com/sqp/godock/widgets/cfbuild/datatype"
	"github.com/sqp/godock/widgets/confgui"
)

//
//-------------------------------------------------------------[ GUI BACKEND ]--

// Bridge provides the internal dock interface to the GUI.
//
type Bridge struct {
	Widget *confgui.GuiConfigure
	Source datatype.Source
	Log    cdtype.Logger
}

// New creates a GUI bridge that will answer to all GUI events.
//
func New(log cdtype.Logger) *Bridge {
	return &Bridge{
		Source: &confdata.Data{},
		Log:    log,
	}
}

// Create creates the config window.
//
func (gc *Bridge) Create() {
	if gc.Widget != nil { // Already opened, give it the focus.
		if !gc.Window().HasFocus() {
			gc.Window().Present()
		}
		return
	}

	var e error
	gc.Widget, e = confgui.NewWindow(gc.Source, gc.Log)
	if gc.Log.Err(e, "create GUI") {
		return
	}

	gc.Widget.Load()

	gc.Widget.GetWindow().Connect("destroy", func() { // OnQuit is already connected to emit this.
		gc.Window().Destroy()
		gc.Widget = nil
	})
}

// GUI interface

// ShowMainGui opens the GUI and displays the dock config page.
//
func (gc *Bridge) ShowMainGui() {
	gc.Create()
	if gc.Widget != nil {
		gc.Widget.Select(confgui.GroupConfig)
	}
}

// ShowModuleGui opens the GUI and should display an item of the config page.
//
// TODO: maybe rename.
// (was to display an applets config and is used only for Help ATM)
//
func (gc *Bridge) ShowModuleGui(key string) {
	gc.Log.Info("ShowModuleGui", key)

	gc.Create()
	if gc.Widget != nil {
		gc.Widget.Select(confgui.GroupConfig, key)
	}
}

// ShowItems opens the GUI and displays the given item icon config.
//
func (gc *Bridge) ShowItems(icon *gldi.Icon, container *gldi.Container, moduleInstance *gldi.ModuleInstance, showPage int) {
	confpath := ""
	if icon != nil {
		confpath = icon.ConfigPath()

	} else if container != nil { // A main dock that is not the first one. Use the dedicated conf file.
		confpath = globals.CurrentThemePath(container.ToCairoDock().GetDockName() + ".conf")
	}

	if confpath == "" {
		gc.Log.Info("ShowGui unmatched", "icon", icon != nil, "- container", container != nil, "- moduleInstance", moduleInstance != nil, "- page", showPage)

	} else {
		gc.Create()
		if gc.Widget != nil {
			gc.Widget.SelectIcons(confpath)
		}
	}
	// cairo_dock_items_widget_select_item (ITEMS_WIDGET (pCategory->pCdWidget), pIcon, pContainer, pModuleInstance, iShowPage);
}

// ShowAddons opens the GUI and displays the applets page.
//
func (gc *Bridge) ShowAddons() {
	gc.Create()
	if gc.Widget != nil {
		gc.Widget.Select(confgui.GroupAdd)
	}
}

// ReloadItems forwards the dock event to the GUI if loaded.
//
func (gc *Bridge) ReloadItems() {
	if gc.Widget != nil {
		gc.Widget.ReloadItems()
	}
}

// func (gc *Bridge) // ReloadCategoryWidget(){}
// func (gc *Bridge) Reload() {
// 	gc.Log.Info("TODO: Reload GUI")
// }

// func (gc *Bridge) Close() {
// 	gc.Log.Info("TODO: Close GUI")
// }

// UpdateModulesList forwards the dock event to the GUI if loaded.
//
func (gc *Bridge) UpdateModulesList() {
	if gc.Widget != nil {
		gc.Widget.UpdateModulesList()
	}
}

// UpdateModuleState forwards the dock event to the GUI if loaded.
//
func (gc *Bridge) UpdateModuleState(name string, active bool) {
	if gc.Widget != nil {
		gc.Widget.UpdateModuleState(name, active)
	}
}

// UpdateModuleInstanceContainer forwards the dock event to the GUI. TODO.
//
func (gc *Bridge) UpdateModuleInstanceContainer(instance *gldi.ModuleInstance, detached bool) {
	gc.Log.Info("TODO: UpdateModuleInstanceContainer")
}

// UpdateShortkeys forwards the dock event to the GUI if loaded.
//
func (gc *Bridge) UpdateShortkeys() {
	if gc.Widget != nil {
		gc.Widget.UpdateShortkeys()
	}
}

// UpdateDeskletParams forwards the dock event to the GUI.
//
func (gc *Bridge) UpdateDeskletParams(desklet *gldi.Desklet) {
	if gc.Widget != nil && desklet != nil {
		gc.Widget.UpdateDeskletParams(&confdata.IconConf{Icon: *desklet.GetIcon()})
	}
}

// UpdateDeskletVisibility forwards the dock event to the GUI.
//
func (gc *Bridge) UpdateDeskletVisibility(desklet *gldi.Desklet) {
	if gc.Widget != nil && desklet != nil {
		gc.Widget.UpdateDeskletVisibility(&confdata.IconConf{Icon: *desklet.GetIcon()})
	}
}

// CORE BACKEND

// SetStatusMessage is unused. TODO.
//
func (gc *Bridge) SetStatusMessage(message string) {
	gc.Log.Info("TODO: SetStatusMessage", message)
	// GtkWidget *pStatusBar = g_object_get_data (G_OBJECT (s_pSimpleConfigWindow), "status-bar");
	// gtk_statusbar_pop (GTK_STATUSBAR (pStatusBar), 0);  // clear any previous message, underflow is allowed.
	// gtk_statusbar_push (GTK_STATUSBAR (pStatusBar), 0, cMessage);
}

// ReloadCurrentWidget is unused. TODO.
//
func (gc *Bridge) ReloadCurrentWidget(moduleInstance *gldi.ModuleInstance, showPage int) {
	gc.Log.Info("TODO: ReloadCurrentWidget")
	// cairo_dock_items_widget_reload_current_widget (ITEMS_WIDGET (pCategory->pCdWidget), pInstance, iShowPage);
}

// ShowModuleInstanceGui is unused. TODO.
//
func (gc *Bridge) ShowModuleInstanceGui(pModuleInstance *gldi.ModuleInstance, iShowPage int) {
	gc.Create()
	if gc.Widget != nil {
		gc.Widget.Select(confgui.GroupIcons)
	}
	gc.Log.Info("TODO: ShowModuleInstanceGui")
	// show_gui (pModuleInstance->pIcon, NULL, pModuleInstance, iShowPage);
}

// func (gc *Bridge) GetWidgetFromName(moduleInstance *gldi.ModuleInstance, group string, key string) {
// 	gc.Log.Info("GetWidgetFromName", group, key)
// }

// Window returns the pointer to the parent window.
//
func (gc *Bridge) Window() *gtk.Window {
	return gc.Widget.GetWindow()
}

// CloseGui closes the configuration window.
//
func (gc *Bridge) CloseGui() {
	if gc.Widget != nil && gc.Widget.GetWindow() != nil {
		gc.Widget.GetWindow().Destroy()
	}
}
