// Package confapplets provides an applets list selection widget with preview.
//
// Can actually display the "can add" applets list.
// The download page is yet to fix to the new data source.
package confapplets

import (
	"github.com/conformal/gotk3/glib"
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/godock/libs/appdbus"
	"github.com/sqp/godock/libs/log"
	// "github.com/sqp/godock/libs/packages"
	"github.com/sqp/godock/widgets/appletlist"
	"github.com/sqp/godock/widgets/appletpreview"
	"github.com/sqp/godock/widgets/confbuilder/datatype"

	"os"
)

// const version = "3.3.0"

// ListMode defines the ConfApplet widget behaviour.
//
type ListMode int

// List widget behaviours.
const (
	ListCanAdd   ListMode = iota // Applets for the add page.
	ListExternal ListMode = iota // Applets for the download page.
)

//
//-------------------------------------------------[ WIDGET APPLETS DOWNLOAD ]--

// ListInterface is the interface to the applets list.
//
type ListInterface interface {
	gtk.IWidget
	Load(map[string]datatype.Appleter)
	Selected() datatype.Appleter
	Delete(string)
}

// GUIControl is the interface to the main GUI and data source.
//
type GUIControl interface {
	SelectIcons(string)
	ListApplets() map[string]datatype.Appleter
}

// ConfApplet provides an applet downloader widget.
// It's connected to a control menu to allow actions on the modules listed.
//
type ConfApplet struct {
	gtk.Box
	menu    MenuDownloader
	control GUIControl
	applets ListInterface
	preview *appletpreview.Preview

	mode    ListMode
	Applets *map[string]datatype.Appleter // List of applets known by the Dock.
}

// New creates a widget to download cairo-dock applets.
//
func New(control GUIControl, menu MenuDownloader, mode ListMode) *ConfApplet {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)

	widget := &ConfApplet{
		Box:     *box,
		preview: appletpreview.New(),
		mode:    mode,
		control: control,
		menu:    menu,
	}
	switch widget.mode {
	case ListCanAdd:
		widget.applets = appletlist.NewListAdd(widget)

	case ListExternal:
		// widget.applets = appletlist.NewListExternal(widget, version)
	}

	widget.PackStart(widget.applets, false, false, 0)
	widget.PackStart(widget.preview, true, true, 4)
	return widget
}

// Load list of applets in the appletlist.
//
func (widget *ConfApplet) Load() {
	list := widget.control.ListApplets()
	switch widget.mode {
	case ListCanAdd:
		widget.applets.Load(list)
		widget.preview.HideState() // Hide must be done onLoad, after the 1st ShowAll.
		widget.preview.HideSize()

	case ListExternal:
		widget.applets.Load(list)
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

// Clean resets the download widget.
//
func (widget *ConfApplet) Clean() {
	if widget.preview.TmpFile != "" {
		os.Remove(widget.preview.TmpFile)
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
}

// NewMenuDownload creates the menu to control the selected applet.
//
func NewMenuDownload() *MenuDownload {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	installed, _ := gtk.SwitchNew()
	active, _ := gtk.SwitchNew()

	widget := &MenuDownload{
		Box:       *box,
		installed: installed,
		active:    active,
	}

	sep, _ := gtk.SeparatorNew(gtk.ORIENTATION_VERTICAL)
	sep2, _ := gtk.SeparatorNew(gtk.ORIENTATION_VERTICAL)
	lblInstalled, _ := gtk.LabelNew("Installed")
	lblActive, _ := gtk.LabelNew("Active")

	// Actions
	var e error
	widget.handlerInstalled, e = widget.installed.Connect("notify::active", widget.toggledInstalled)
	log.Err(e, "Connect installed button callback")
	widget.handlerActive, e = widget.active.Connect("notify::active", widget.toggledActive)
	log.Err(e, "Connect active button callback")

	widget.PackStart(sep, false, false, 8)
	widget.PackStart(lblInstalled, false, false, 8)
	widget.PackStart(installed, false, false, 0)
	widget.PackStart(sep2, false, false, 4)
	widget.PackStart(lblActive, false, false, 0)
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
	// widget.installed.SetSensitive(pack.Type != packages.TypeInDev) // Disable uninstall button if it's a user special applet.
	widget.installed.HandlerBlock(widget.handlerInstalled)
	widget.installed.SetActive(pack.IsInstalled())
	widget.installed.HandlerUnblock(widget.handlerInstalled)

	// Set installed button state and disable it if the package isn't installed yet.
	moduleActive := (len(appdbus.AppletInstances(pack.GetTitle())) > 0)
	widget.active.SetSensitive(pack.IsInstalled())
	widget.active.HandlerBlock(widget.handlerActive)
	widget.active.SetActive(moduleActive)
	widget.active.HandlerUnblock(widget.handlerActive)
}

//-------------------------------------------------------[ ACTIONS CALLBACKS ]--

// Action on Installed button. Need to install or delete the applet.
//
func (widget *MenuDownload) toggledInstalled(switc *gtk.Switch) {
	// name := widget.current.DisplayedName
	// if widget.installed.GetActive() { // Install
	// 	e := widget.current.Install(version, "")

	// 	if log.Err(e, "Installing "+name) { // Install failed. Force unchecked state of widget.
	// 		widget.installed.HandlerBlock(widget.handlerInstalled)
	// 		widget.installed.SetActive(false)
	// 		widget.installed.HandlerUnblock(widget.handlerInstalled)
	// 		return
	// 	}

	// 	log.Info("Installed applet", name)
	// 	widget.active.SetSensitive(true)
	// 	widget.applets.SetActive(true)

	// } else { // Uninstall
	// 	e := widget.current.Uninstall(version)
	// 	if log.Err(e, "Uninstall package") { // Uninstall failed. Force checked state of widget.
	// 		widget.installed.HandlerBlock(widget.handlerInstalled)
	// 		widget.installed.SetActive(true)
	// 		widget.installed.HandlerUnblock(widget.handlerInstalled)
	// 		return
	// 	}

	// 	log.Info("Removed applet", name)

	// 	widget.applets.SetActive(false)
	// 	widget.active.SetSensitive(false)
	// }
}

// Action on Active button. Need to activate or unload the applet.
//
func (widget *MenuDownload) toggledActive(switc *gtk.Switch) {
	if widget.active.GetActive() {
		appdbus.AppletAdd(widget.current.GetTitle())
	} else {
		appdbus.AppletRemove(widget.current.GetTitle() + ".conf")
	}
}
