// Package gui provides a connection between the config window and the gldi backend and data.
package gui

import (
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/godock/libs/cdtype" // Logger type.
	"github.com/sqp/godock/libs/gldi"
	"github.com/sqp/godock/libs/gldi/globals"

	"github.com/sqp/godock/widgets/confbuilder/datagldi"
	"github.com/sqp/godock/widgets/confbuilder/datatype"
	"github.com/sqp/godock/widgets/confgui"
)

//
//-------------------------------------------------------------[ GUI BACKEND ]--

// Connector connects the config widget and window to the data source and
// provides an inferface for backendgui.
//
type Connector struct {
	Widget *confgui.GuiConfigure
	Win    *gtk.Window
	Source datatype.Source
	Log    cdtype.Logger
}

// NewConnector creates a GUI connector that will answer to all GUI events.
//
func NewConnector(log cdtype.Logger) *Connector {
	return &Connector{
		Source: &datagldi.Data{},
		Log:    log,
	}
}

// Create creates the config window.
//
func (gc *Connector) Create() {
	if gc.Widget != nil || gc.Win != nil {
		gc.Log.Info("create GUI, found: widget", gc.Widget != nil, " window", gc.Win != nil)
		return
	}
	gc.Widget, gc.Win = confgui.NewConfigWindow(gc.Source, gc.Log)

	gc.Win.Connect("destroy", func() { // OnQuit is already connected to emit this.
		gc.Widget.Destroy()
		gc.Widget = nil
		gc.Win.Destroy()
		gc.Win = nil
	})

	gc.Widget.Load()
}

// GUI interface

// ShowMainGui opens the GUI and displays the dock config page.
//
func (gc *Connector) ShowMainGui() {
	gc.Create()
	gc.Widget.Select(confgui.GroupConfig)
}

// ShowModuleGui opens the GUI and should display an applets config. TODO: Fix. (used only for Help ATM)
//
func (gc *Connector) ShowModuleGui(appletName string) {
	gc.Log.Info("ShowModuleGui", appletName)

	gc.Create()
	gc.Widget.Select(confgui.GroupIcons)
}

// ShowItems opens the GUI and displays the given item icon config.
//
func (gc *Connector) ShowItems(icon *gldi.Icon, container *gldi.Container, moduleInstance *gldi.ModuleInstance, showPage int) {
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
		gc.Widget.SelectIcons(confpath)
	}
	// cairo_dock_items_widget_select_item (ITEMS_WIDGET (pCategory->pCdWidget), pIcon, pContainer, pModuleInstance, iShowPage);
}

// ShowAddons opens the GUI and displays the applets page.
//
func (gc *Connector) ShowAddons() {
	gc.Create()
	gc.Widget.Select(confgui.GroupAdd)
}

// ReloadItems forwards the dock event to the GUI if loaded.
//
func (gc *Connector) ReloadItems() {
	if gc.Widget != nil {
		gc.Widget.ReloadItems()
	}
}

// func (gc *Connector) // ReloadCategoryWidget(){}
// func (gc *Connector) Reload() {
// 	gc.Log.Info("TODO: Reload GUI")
// }

// func (gc *Connector) Close() {
// 	gc.Log.Info("TODO: Close GUI")
// }

// UpdateModulesList forwards the dock event to the GUI if loaded.
//
func (gc *Connector) UpdateModulesList() {
	if gc.Widget != nil {
		gc.Widget.UpdateModulesList()
	}
}

// UpdateModuleState forwards the dock event to the GUI if loaded.
//
func (gc *Connector) UpdateModuleState(name string, active bool) {
	if gc.Widget != nil {
		gc.Widget.UpdateModuleState(name, active)
	}
}

// UpdateModuleInstanceContainer forwards the dock event to the GUI. TODO.
//
func (gc *Connector) UpdateModuleInstanceContainer(instance *gldi.ModuleInstance, detached bool) {
	gc.Log.Info("TODO: UpdateModuleInstanceContainer")
}

// UpdateShortkeys forwards the dock event to the GUI if loaded.
//
func (gc *Connector) UpdateShortkeys() {
	if gc.Widget != nil {
		gc.Widget.UpdateShortkeys()
	}
}

// UpdateDeskletParams forwards the dock event to the GUI.
//
func (gc *Connector) UpdateDeskletParams(desklet *gldi.Desklet) {
	if gc.Widget != nil && desklet != nil {
		gc.Widget.UpdateDeskletParams(&datagldi.IconConf{*desklet.GetIcon()})
	}
}

// UpdateDeskletVisibility forwards the dock event to the GUI.
//
func (gc *Connector) UpdateDeskletVisibility(desklet *gldi.Desklet) {
	if gc.Widget != nil && desklet != nil {
		gc.Widget.UpdateDeskletVisibility(&datagldi.IconConf{*desklet.GetIcon()})
	}
}

// CORE BACKEND

// SetStatusMessage is unused. TODO.
//
func (gc *Connector) SetStatusMessage(message string) {
	gc.Log.Info("TODO: SetStatusMessage", message)
	// GtkWidget *pStatusBar = g_object_get_data (G_OBJECT (s_pSimpleConfigWindow), "status-bar");
	// gtk_statusbar_pop (GTK_STATUSBAR (pStatusBar), 0);  // clear any previous message, underflow is allowed.
	// gtk_statusbar_push (GTK_STATUSBAR (pStatusBar), 0, cMessage);
}

// ReloadCurrentWidget is unused. TODO.
//
func (gc *Connector) ReloadCurrentWidget(moduleInstance *gldi.ModuleInstance, showPage int) {
	gc.Log.Info("TODO: ReloadCurrentWidget")
	// cairo_dock_items_widget_reload_current_widget (ITEMS_WIDGET (pCategory->pCdWidget), pInstance, iShowPage);
}

// ShowModuleInstanceGui is unused. TODO.
//
func (gc *Connector) ShowModuleInstanceGui(pModuleInstance *gldi.ModuleInstance, iShowPage int) {
	gc.Create()
	gc.Widget.Select(confgui.GroupIcons)
	gc.Log.Info("TODO: ShowModuleInstanceGui")
	// show_gui (pModuleInstance->pIcon, NULL, pModuleInstance, iShowPage);
}

// func (gc *Connector) GetWidgetFromName(moduleInstance *gldi.ModuleInstance, group string, key string) {
// 	gc.Log.Info("GetWidgetFromName", group, key)
// }

// Window returns the pointer to the parent window.
//
func (gc *Connector) Window() *gtk.Window {
	return gc.Win
}

// CloseGui closes the configuration window.
//
func (gc *Connector) CloseGui() {
	if gc.Win != nil {
		gc.Win.Destroy()
	}
}
