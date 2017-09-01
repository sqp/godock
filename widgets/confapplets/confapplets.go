// Package confapplets provides an applets list selection widget with preview.
//
package confapplets

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/text/tran"

	"github.com/sqp/godock/widgets/appletlist"
	"github.com/sqp/godock/widgets/appletpreview"
	"github.com/sqp/godock/widgets/cfbuild/datatype"
	"github.com/sqp/godock/widgets/gtk/newgtk"

	"os"
)

// ListMode defines the ConfApplet widget behaviour.
//
type ListMode int

// List widget behaviours.
const (
	ListCanAdd   ListMode = iota // Applets for the add page.
	ListExternal                 // Applets for the download page.
	ListThemes                   // Global dock themes.
)

//
//-------------------------------------------------[ WIDGET APPLETS DOWNLOAD ]--

// ListInterface is the interface to the applets list.
//
type ListInterface interface {
	ListInterfaceBase
	Selected() datatype.Appleter
	Delete(string)
}

// ListInterfaceBase is the interface to the applets list.
//
type ListInterfaceBase interface {
	gtk.IWidget
	Load(map[string]datatype.Appleter)
	Clear()
}

// UpdateModuleStater defines a widget able to update module state.
//
type UpdateModuleStater interface {
	UpdateModuleState(string, bool)
}

// GUIControl is the interface to the main GUI and data source.
//
type GUIControl interface {
	ListKnownApplets() map[string]datatype.Appleter
	ListDownloadApplets() (map[string]datatype.Appleter, error)
	ListDockThemeLoad() (map[string]datatype.Appleter, error)
	ListDockThemeSave() []datatype.Field
}

// SelectIconser defines the optional control item selection.
//
type SelectIconser interface {
	SelectIcons(string)
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

// New creates a widget to list cairo-dock applets and themes.
//
func New(control GUIControl, log cdtype.Logger, menu MenuDownloader, mode ListMode) *ConfApplet {
	mainbox := newgtk.Box(gtk.ORIENTATION_VERTICAL, 0)

	widget := &ConfApplet{
		Box:     *mainbox,
		menu:    menu,
		preview: appletpreview.New(log),
		control: control,
		log:     log,
		mode:    mode,
	}
	var preview gtk.IWidget = widget.preview

	switch widget.mode {
	case ListCanAdd:
		widget.applets = appletlist.NewListAdd(widget, log)
		widget.preview.HideState()
		widget.preview.HideSize()

	case ListExternal:
		widget.applets = appletlist.NewListExternal(widget, log)

		widget.preview.Load(&datatype.HandbookSimple{
			Title:  "Download applets page",
			Author: tran.Slate("Cairo-Dock contributors"),
			// Preview: "",
			Description: `Here, you can download external applets from the repository.
They will be directly activated after the download

You can also find them on the <a href="http://glx-dock.org/mc_album.php?a=12">Repository website</a>`,
		})

		if menu == nil { // Menu not provided, pack one above the preview.
			menu := NewMenuDownload(log)
			widget.menu = menu
			box := newgtk.Box(gtk.ORIENTATION_VERTICAL, 0)
			box.PackStart(menu, false, false, 2)
			box.PackStart(newgtk.Separator(gtk.ORIENTATION_HORIZONTAL), false, false, 2)
			box.PackStart(preview, true, true, 2)

			preview = box
		}

	case ListThemes:
		widget.applets = appletlist.NewListThemes(widget, log)

		widget.preview.Load(&datatype.HandbookSimple{
			Title:  "Download themes page",
			Author: tran.Slate("Cairo-Dock contributors"),
			// Preview: "",
			Description: "Here, you can download full dock themes from the repository.",
		})
	}

	inbox := newgtk.Box(gtk.ORIENTATION_HORIZONTAL, 0)

	widget.PackStart(inbox, true, true, 0)
	inbox.PackStart(widget.applets, false, false, 0)
	inbox.PackStart(preview, true, true, 4)
	widget.ShowAll()
	return widget
}

// NewLoaded creates a widget to list cairo-dock applets and themes and loads data.
//
func NewLoaded(control GUIControl, log cdtype.Logger, menu MenuDownloader, mode ListMode) *ConfApplet {
	w := New(control, log, menu, mode)
	w.Load()
	return w
}

// Load list of applets in the appletlist.
//
func (widget *ConfApplet) Load() {
	switch widget.mode {
	case ListCanAdd:
		applets := widget.control.ListKnownApplets()
		widget.applets.Load(applets)

	case ListExternal:
		applets, e := widget.control.ListDownloadApplets()
		if !widget.log.Err(e, "external applets list") {
			widget.applets.Load(applets)
		}

	case ListThemes:
		applets, e := widget.control.ListDockThemeLoad()
		if !widget.log.Err(e, "ListDockThemes") {
			widget.applets.Load(applets)
		}
	}
}

// Save enables the selected applet (user clicked the add button).
//
func (widget *ConfApplet) Save() {
	switch widget.mode {
	case ListCanAdd:
		app := widget.applets.Selected()
		newconf := app.Activate()

		selecter, ok := widget.control.(SelectIconser)
		if ok {
			selecter.SelectIcons(newconf)
		}
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

// Selected returns the name of the selected page.
//
func (widget *ConfApplet) Selected() datatype.Appleter {
	return widget.applets.Selected()
}

// UpdateModuleState updates the state of the given applet, from a dock event.
//
func (widget *ConfApplet) UpdateModuleState(name string, active bool) {
	switch widget.mode {
	case ListCanAdd:
		us := widget.applets.(UpdateModuleStater)
		us.UpdateModuleState(name, active)

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
	widget := &MenuDownload{
		Box:       *newgtk.Box(gtk.ORIENTATION_HORIZONTAL, 0),
		installed: newgtk.Switch(),
		active:    newgtk.Switch(),
		log:       log,
	}

	// Actions
	var e error
	widget.handlerInstalled, e = widget.installed.Connect("notify::active", widget.toggledInstalled)
	log.Err(e, "Connect installed button callback")
	widget.handlerActive, e = widget.active.Connect("notify::active", widget.toggledActive)
	log.Err(e, "Connect active button callback")

	widget.PackStart(newgtk.Label("Installed"), false, false, 4)
	widget.PackStart(widget.installed, false, false, 0)
	widget.PackStart(newgtk.Box(gtk.ORIENTATION_VERTICAL, 0), false, false, 8)
	widget.PackStart(newgtk.Label("Active"), false, false, 4)
	widget.PackStart(widget.active, false, false, 0)
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
	widget.active.SetSensitive(!state)
	widget.installed.HandlerBlock(widget.handlerInstalled)
	widget.installed.SetActive(state)
	widget.installed.HandlerUnblock(widget.handlerInstalled)
}

// SetActiveState sets the state of the 'active' switch.
//
func (widget *MenuDownload) SetActiveState(state bool) {
	widget.installed.SetSensitive(widget.current.CanUninstall())
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
		widget.installed.SetSensitive(widget.current.CanUninstall())
		widget.applets.SetActive(true)

	} else { // Uninstall
		if !widget.current.CanUninstall() {
			widget.SetInstalledState(true)
			return
		}

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
