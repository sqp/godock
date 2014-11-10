// Package gui provides a connection between the config window and the gldi backend and data.
package gui

import (
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/godock/libs/cdtype" // Logger type.
	"github.com/sqp/godock/libs/gldi"

	"github.com/sqp/godock/widgets/confbuilder/datagldi"
	"github.com/sqp/godock/widgets/confbuilder/datatype"
	"github.com/sqp/godock/widgets/confgui"
)

//
//-------------------------------------------------------------[ GUI BACKEND ]--

// GuiConnector connects the config widget and window to the data source and
// provides an inferface for backendgui.
//
type GuiConnector struct {
	Widget *confgui.GuiConfigure
	Win    *gtk.Window
	Source datatype.Source
	Log    cdtype.Logger
}

func NewConnector(log cdtype.Logger) *GuiConnector {
	return &GuiConnector{
		Source: &datagldi.Data{},
		Log:    log}
}

// Create creates the config window.
//
func (gc *GuiConnector) Create() {
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

func (gc *GuiConnector) ShowMainGui() {
	gc.Create()
	gc.Widget.Select(confgui.GroupConfig)
}

func (gc *GuiConnector) ShowModuleGui(appletName string) {
	gc.Log.Info("ShowModuleGui", appletName)

	gc.Create()
	gc.Widget.Select(confgui.GroupIcons)
}

func (gc *GuiConnector) ShowItems(icon *gldi.Icon, container *gldi.Container, moduleInstance *gldi.ModuleInstance, showPage int) {
	gc.Log.Info("ShowGui", "icon", icon != nil, "- container", container != nil, "- moduleInstance", moduleInstance != nil, "- page", showPage)

	gc.Create()

	if icon != nil {
		confPath := icon.ConfigPath()
		gc.Widget.SelectIcons(confPath)
	}
	// cairo_dock_items_widget_select_item (ITEMS_WIDGET (pCategory->pCdWidget), pIcon, pContainer, pModuleInstance, iShowPage);
}

func (gc *GuiConnector) ShowAddons() {
	gc.Create()
	gc.Widget.Select(confgui.GroupAdd)
}

func (gc *GuiConnector) ReloadItems() {
	if gc.Widget != nil {
		gc.Widget.ReloadItems()
	}
}

// func (gc *GuiConnector) // ReloadCategoryWidget(){}
// func (gc *GuiConnector) Reload() {
// 	gc.Log.Info("TODO: Reload GUI")
// }

// func (gc *GuiConnector) Close() {
// 	gc.Log.Info("TODO: Close GUI")
// }

func (gc *GuiConnector) UpdateModulesList() {
	if gc.Widget != nil {
		gc.Widget.UpdateModulesList()
	}
}

func (gc *GuiConnector) UpdateModuleState(name string, active bool) {
	if gc.Widget != nil {
		gc.Widget.UpdateModuleState(name, active)
	}
}

func (gc *GuiConnector) UpdateModuleInstanceContainer(instance *gldi.ModuleInstance, detached bool) {
	gc.Log.Info("TODO: UpdateModuleInstanceContainer")
}

func (gc *GuiConnector) UpdateShortkeys() {
	if gc.Widget != nil {
		gc.Widget.UpdateShortkeys()
	}
}

// func (gc *GuiConnector) UpdateDeskletParams(*gldi.Desklet)                                          {gc.Log.Info("UpdateDeskletParams")}
// func (gc *GuiConnector) UpdateDeskletVisibility(*gldi.Desklet)                                      {gc.Log.Info("UpdateDeskletVisibility")}

// CORE BACKEND
func (gc *GuiConnector) SetStatusMessage(message string) {
	gc.Log.Info("TODO: SetStatusMessage", message)
	// GtkWidget *pStatusBar = g_object_get_data (G_OBJECT (s_pSimpleConfigWindow), "status-bar");
	// gtk_statusbar_pop (GTK_STATUSBAR (pStatusBar), 0);  // clear any previous message, underflow is allowed.
	// gtk_statusbar_push (GTK_STATUSBAR (pStatusBar), 0, cMessage);
}

func (gc *GuiConnector) ReloadCurrentWidget(moduleInstance *gldi.ModuleInstance, showPage int) {
	gc.Log.Info("TODO: ReloadCurrentWidget")
	// cairo_dock_items_widget_reload_current_widget (ITEMS_WIDGET (pCategory->pCdWidget), pInstance, iShowPage);
}

func (gc *GuiConnector) ShowModuleInstanceGui(pModuleInstance *gldi.ModuleInstance, iShowPage int) {
	gc.Create()
	gc.Widget.Select(confgui.GroupIcons)
	gc.Log.Info("TODO: ShowModuleInstanceGui")
	// show_gui (pModuleInstance->pIcon, NULL, pModuleInstance, iShowPage);
}

// func (gc *GuiConnector) GetWidgetFromName(moduleInstance *gldi.ModuleInstance, group string, key string) {
// 	gc.Log.Info("GetWidgetFromName", group, key)
// }

// Window returns the pointer to the parent window.
//
func (gc *GuiConnector) Window() *gtk.Window { return gc.Win }
