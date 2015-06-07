// Package confapplets provides an applets list selection widget with preview.
//
package confapplets

import (
	"github.com/conformal/gotk3/glib"
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/widgets/appletlist"
	"github.com/sqp/godock/widgets/appletpreview"
	"github.com/sqp/godock/widgets/confbuilder/datatype"

	"os"
)

// ListMode defines the ConfApplet widget behaviour.
//
type ListMode int

// List widget behaviours.
const (
	ListCanAdd   ListMode = iota // Applets for the add page.
	ListExternal                 // Applets for the download page.
)

//
//-------------------------------------------------[ WIDGET APPLETS DOWNLOAD ]--

// ListInterface is the interface to the applets list.
//
type ListInterface interface {
	gtk.IWidget
	Load(map[string]datatype.Appleter)
	Selected() datatype.Appleter
	UpdateModuleState(string, bool)
	Delete(string)
	Clear()
}

// GUIControl is the interface to the main GUI and data source.
//
type GUIControl interface {
	SelectIcons(string)
	ListKnownApplets() map[string]datatype.Appleter
	ListDownloadApplets() map[string]datatype.Appleter
}

// ConfApplet provides an applets list and preview widget.
// It can be connected to a control menu to allow actions on the modules listed.
//
type ConfApplet struct {
	gtk.Box
	menu    MenuDownloader
	applets ListInterface
	preview *appletpreview.Preview

	control GUIControl
	log     cdtype.Logger

	mode    ListMode
	Applets *map[string]datatype.Appleter // List of applets known by the Dock.
}

// New creates a widget to download cairo-dock applets.
//
func New(control GUIControl, log cdtype.Logger, menu MenuDownloader, mode ListMode) *ConfApplet {
	mainbox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)

	widget := &ConfApplet{
		Box:     *mainbox,
		menu:    menu,
		preview: appletpreview.New(log),
		control: control,
		log:     log,
		mode:    mode,
	}
	switch widget.mode {
	case ListCanAdd:
		widget.applets = appletlist.NewListAdd(widget, log)
		widget.preview.HideState()
		widget.preview.HideSize()

	case ListExternal:
		if menu == nil {
			menu := NewMenuDownload(log)
			widget.menu = menu
			widget.PackStart(menu, false, false, 0)
		}
		widget.applets = appletlist.NewListExternal(widget, log)
	}

	inbox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)

	widget.PackStart(inbox, true, true, 0)
	inbox.PackStart(widget.applets, false, false, 0)
	inbox.PackStart(widget.preview, true, true, 4)
	return widget
}

// Load list of applets in the appletlist.
//
func (widget *ConfApplet) Load() {
	switch widget.mode {
	case ListCanAdd:
		applets := widget.control.ListKnownApplets()
		widget.applets.Load(applets)

	case ListExternal:
		applets := widget.control.ListDownloadApplets()
		widget.applets.Load(applets)
	}
}

// Save enables the selected applet (user clicked the add button).
//
func (widget *ConfApplet) Save() {
	switch widget.mode {
	case ListCanAdd:
		app := widget.applets.Selected()
		newconf := app.Activate()
		widget.control.SelectIcons(newconf)
		if !app.CanAdd() {
			widget.applets.Delete(app.GetName())
		}
	}
}

// Clean removes temporary files.
//
func (widget *ConfApplet) Clean() {
	if widget.preview.TmpFile != "" {
		os.Remove(widget.preview.TmpFile)
	}
}

// Clear clears the widget data.
//
func (widget *ConfApplet) Clear() {
	widget.applets.Clear()
}

// UpdateModuleState updates the state of the given applet, from a dock event.
//
func (widget *ConfApplet) UpdateModuleState(name string, active bool) {
	switch widget.mode {
	case ListCanAdd:
		widget.applets.UpdateModuleState(name, active)

	case ListExternal:
		sel := widget.applets.Selected()
		if widget.menu != nil && sel != nil && sel.GetTitle() == name {
			widget.menu.SetActiveState(active)
		}
	}
}

//--------------------------------------------------[ LIST CONTROL CALLBACKS ]--

// OnSelect reacts when a row is selected. Show preview and set menu position.
//
func (widget *ConfApplet) OnSelect(pack datatype.Appleter) {
	if widget.menu != nil {
		widget.menu.OnSelect(pack)
	}
	widget.preview.Load(pack)
}

// SetControlInstall forwards the list controler to the menu for updates.
//
func (widget *ConfApplet) SetControlInstall(ctrl appletlist.ControlInstall) {
	if widget.menu != nil {
		widget.menu.SetControlInstall(ctrl)
	}
}

//
//----------------------------------------------------[ WIDGET MENU DOWNLOAD ]--

// MenuDownloader forwards events to other widgets.
//
type MenuDownloader interface {
	OnSelect(datatype.Appleter)
	SetControlInstall(appletlist.ControlInstall)
	SetActiveState(bool)
}

// MenuDownload provides install and active switches to control the selected applet.
//
type MenuDownload struct {
	gtk.Box

	installed        *gtk.Switch
	active           *gtk.Switch
	handlerInstalled glib.SignalHandle
	handlerActive    glib.SignalHandle

	applets appletlist.ControlInstall // still needed?
	current datatype.Appleter

	log cdtype.Logger
}

// NewMenuDownload creates the menu to control the selected applet.
//
func NewMenuDownload(log cdtype.Logger) *MenuDownload {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	installed, _ := gtk.SwitchNew()
	active, _ := gtk.SwitchNew()

	widget := &MenuDownload{
		Box:       *box,
		installed: installed,
		active:    active,
		log:       log,
	}

	sep, _ := gtk.SeparatorNew(gtk.ORIENTATION_VERTICAL)
	lblInstalled, _ := gtk.LabelNew("Installed")
	lblActive, _ := gtk.LabelNew("Active")

	// Actions
	var e error
	widget.handlerInstalled, e = widget.installed.Connect("notify::active", widget.toggledInstalled)
	log.Err(e, "Connect installed button callback")
	widget.handlerActive, e = widget.active.Connect("notify::active", widget.toggledActive)
	log.Err(e, "Connect active button callback")

	widget.PackStart(lblInstalled, false, false, 8)
	widget.PackStart(installed, false, false, 0)
	widget.PackStart(sep, false, false, 4)
	widget.PackStart(lblActive, false, false, 8)
	widget.PackStart(active, false, false, 0)
	return widget
}

//-------------------------------------------------------[ CONTROL CALLBACKS ]--

// SetControlInstall forwards the list controler to the menu for updates.
//
func (widget *MenuDownload) SetControlInstall(ctrl appletlist.ControlInstall) {
	widget.applets = ctrl
}

// OnSelect reacts when a row is selected.
// Set preview data and set installed and active buttons state.
//
func (widget *MenuDownload) OnSelect(pack datatype.Appleter) {
	widget.current = pack

	// Set installed button state.
	widget.installed.SetSensitive(pack.CanUninstall()) // Disable uninstall button if it's a user special applet.
	widget.SetInstalledState(pack.IsInstalled())

	// Set installed button state and disable it if the package isn't installed yet.
	widget.active.SetSensitive(pack.IsInstalled())
	widget.SetActiveState(pack.IsActive())
}

// SetInstalledState sets the state of the 'installed' switch.
//
func (widget *MenuDownload) SetInstalledState(state bool) {
	// widget.active.SetSensitive(!state)
	widget.installed.HandlerBlock(widget.handlerInstalled)
	widget.installed.SetActive(state)
	widget.installed.HandlerUnblock(widget.handlerInstalled)
}

// SetActiveState sets the state of the 'active' switch.
//
func (widget *MenuDownload) SetActiveState(state bool) {
	// widget.installed.SetSensitive(pack.CanUninstall())
	widget.active.HandlerBlock(widget.handlerActive)
	widget.active.SetActive(state)
	widget.active.HandlerUnblock(widget.handlerActive)
}

//-------------------------------------------------------[ ACTIONS CALLBACKS ]--

// Action on Installed button. Need to install or delete the applet.
//
func (widget *MenuDownload) toggledInstalled(switc *gtk.Switch) {
	name := widget.current.GetTitle()
	if widget.installed.GetActive() { // Install
		e := widget.current.Install("")

		if widget.log.Err(e, "Install failed: "+name) { // Failed => Force unchecked state of widget.
			widget.SetInstalledState(false)
			return
		}

		widget.log.Info("Installed applet", name)
		widget.active.SetSensitive(true)
		widget.applets.SetActive(true)

	} else { // Uninstall
		e := widget.current.Uninstall()
		if widget.log.Err(e, "Uninstall package") { // Uninstall failed. Force checked state of widget.
			widget.SetInstalledState(true)
			return
		}

		widget.log.Info("Removed applet", name)

		widget.applets.SetActive(false)
		widget.active.SetSensitive(false)
	}
}

// Action on Active button. Need to activate or unload the applet.
//
func (widget *MenuDownload) toggledActive(switc *gtk.Switch) {
	if widget.active.GetActive() {
		widget.current.Activate()
	} else {
		widget.current.Deactivate()
	}
}
