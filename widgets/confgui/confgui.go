/*
Package confgui is a configuration window for Cairo-Dock.


Using GTK 3.10 library https://github.com/conformal/gotk3

If you use GTK 3.8, you will have to add a flag to compile it:
  go get -tags gtk_3_8 github.com/conformal/gotk3/gtk


Gui xml files are compressed with github.com/jteeuwen/go-bindata

*/
package confgui

import (
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/godock/libs/cdtype" // Logger type.
	"github.com/sqp/godock/libs/tran"

	"github.com/sqp/godock/widgets/confapplets"
	"github.com/sqp/godock/widgets/confbuilder/datatype"
	"github.com/sqp/godock/widgets/confcore"
	"github.com/sqp/godock/widgets/conficons"
	"github.com/sqp/godock/widgets/confmenu"
	"github.com/sqp/godock/widgets/confsettings"
	"github.com/sqp/godock/widgets/pageswitch"
)

// Window settings.
const (
	WindowTitle  = "Cairo-Dock Config"
	WindowClass  = "cdc"
	WindowWidth  = 900
	WindowHeight = 600
)

// Main config groups.
const (
	GroupIcons  = "Icons"
	GroupAdd    = "Add"
	GroupConfig = "Config"
)

// NewStandalone creates a new config window to use as standalone application.
//
func NewStandalone(data datatype.Source, log cdtype.Logger, path ...string) {
	gtk.Init(nil)

	widget, win := NewConfigWindow(data, log)
	win.Connect("destroy", gtk.MainQuit)

	widget.Load()
	// widget.Menu.Switcher.Activate("Icons")

	if len(path) > 0 {
		widget.SelectIcons(path[0])
	}

	gtk.Main()

	// log.Info("GUI QUITTED OK !!")
	win.Destroy()
}

// NewConfigWindow creates a new config widget and window, ready to use.
//
func NewConfigWindow(data datatype.Source, log cdtype.Logger) (*GuiConfigure, *gtk.Window) {
	widget := NewGuiConfigure(data, log)
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		// log.Fatal("Unable to create window:", err)
		return nil, nil
	}
	win.SetDefaultSize(WindowWidth, WindowHeight)
	win.Add(widget)

	win.SetTitle(WindowTitle)
	win.SetWMClass(WindowClass, WindowTitle)

	win.SetIconFromFile(data.AppIcon())

	// win.Set("border-width", 4)

	win.ShowAll()
	widget.SetWindow(win)
	widget.OnQuit = win.Destroy

	return widget, win
}

//

// Page defines a switcher page.
//
type Page struct {
	Name   string
	Widget Saver
	OnShow func()
	OnHide func()
}

// Saver extends the Widget interface with a Save action.
//
type Saver interface {
	gtk.IWidget
	Load()
	Save()
}

// Selecter defines the interface to select an item in the config page.
//
type Selecter interface {
	Select(string) bool
}

// Clearer defines the interface to clear the data of a config page.
//
type Clearer interface {
	Saver
	Selecter
	Clear() string
}

// UpdateShortkeyser defines the interface to update shortkeys of a config page.
//
type UpdateShortkeyser interface {
	UpdateShortkeys()
}

//-----------------------------------------------------------[ MAIN DOCK GUI ]--

// GuiConfigure defines the main Cairo-Dock configuration widget.
//
type GuiConfigure struct {
	gtk.Box

	datatype.Source // embeds the data source.

	window *gtk.Window       // pointer to the parent window.
	Menu   *confmenu.MenuBar // GUI menu widget.
	stack  *gtk.Stack        // GUI main switcher (icons/add/config).

	OnQuit func() // On clicked Quit callback.

	iconToSelect string // Cache for the icon name to select as ReloadItems is called after the display (fix case new item).

	pages   map[string]*Page
	current *Page

	log cdtype.Logger
}

// NewGuiConfigure creates the main Cairo-Dock configuration widget.
//
func NewGuiConfigure(source datatype.Source, log cdtype.Logger) *GuiConfigure {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)

	widget := &GuiConfigure{
		Source: source,
		Box:    *box,
		pages:  make(map[string]*Page),
		log:    log,
	}

	// Load GUI own config page settings.
	e := confsettings.Init(source.DirAppData())
	log.Err(e, "Load ConfigSettings")

	// Create widgets.

	widget.Menu = confmenu.New(widget)
	menuIcons := pageswitch.New()
	menuIcons.Set("no-show-all", true)

	// Box for separator on left of menuIcons.
	boxIcons, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	sepIcons, _ := gtk.SeparatorNew(gtk.ORIENTATION_VERTICAL)
	boxIcons.PackStart(sepIcons, false, false, 6)
	boxIcons.PackStart(menuIcons, false, false, 0)

	widget.stack, _ = gtk.StackNew()

	sw, _ := gtk.StackSwitcherNew()
	sw.SetStack(widget.stack)
	sw.SetHomogeneous(false)

	icons := conficons.New(widget, log, menuIcons)
	core := confcore.New(widget, log, menuIcons)
	add := confapplets.New(widget, log, nil, confapplets.ListCanAdd)

	// Add pages to the switcher. This will pack the pages widgets to the gui box.

	widget.AddPage(icons, GroupIcons, func() { widget.SetActionSave(); menuIcons.Show() }, func() { widget.Menu.Save.Hide(); menuIcons.Hide() })
	widget.AddPage(add, GroupAdd, widget.SetActionAdd, widget.Menu.Save.Hide)
	widget.AddPage(core, GroupConfig, core.SetAction, widget.Menu.Save.Hide)

	widget.stack.Connect("notify::visible-child-name", func() {
		widget.OnSelectPage(widget.stack.GetVisibleChildName())
	})

	// Packing menu.

	sep, _ := gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)

	widget.Menu.PackStart(sw, false, false, 0)
	widget.Menu.PackStart(boxIcons, false, false, 0)

	widget.PackStart(widget.Menu, false, false, 2)
	widget.PackStart(sep, false, false, 0)

	widget.PackStart(widget.stack, true, true, 0)

	return widget
}

// AddPage adds a tab to the main config switcher with its widget.
//
func (widget *GuiConfigure) AddPage(saver Saver, name string, onShow, onHide func()) {
	widget.stack.AddTitled(saver, name, tran.Slate(name))
	widget.pages[name] = &Page{
		Name:   name,
		Widget: saver,
		OnShow: onShow,
		OnHide: onHide,
	}
}

// Load loads all pages data.
//
func (widget *GuiConfigure) Load() {
	for _, page := range widget.pages {
		page.Widget.Load()
	}

	// widget.stack.SetVisibleChildName(GroupConfig)
	// widget.Switch(GroupIcons)
	// widget.Menu.Switcher.Load()
}

// SetWindow sets the pointer to the parent window, used for some config
// callbacks (grab events).
//
func (widget *GuiConfigure) SetWindow(win *gtk.Window) {
	widget.window = win
}

// GetWindow returns the pointer to the parent window.
//
func (widget *GuiConfigure) GetWindow() *gtk.Window {
	return widget.window
}

//
//-----------------------------------------------------[ INTERFACE CALLBACKS ]--

// ClickedSave forwards the save event to the current widget.
//
func (widget *GuiConfigure) ClickedSave() {
	widget.current.Widget.Save()
}

// ClickedQuit launches the OnQuit event defined.
//
func (widget *GuiConfigure) ClickedQuit() {
	if widget.OnQuit != nil {
		go widget.OnQuit()
	}
}

// SelectIcons selects a specific icon in the Icons page (key = full path to config file).
// If the icon isn't found, the name is cached for the late ReloadItems callback.
//
func (widget *GuiConfigure) SelectIcons(item string) {
	// widget.log.Info("SelectIcons", item)
	b := widget.Select(GroupIcons, item)
	if b {
		widget.iconToSelect = "" // Found, clear cache.
	} else {
		widget.iconToSelect = item // Not found, set the name to the cache for the late ReloadItems call.
	}
}

// Select selects the given group page and may also select a specific item in the page.
//
func (widget *GuiConfigure) Select(page string, item ...string) bool {
	widget.log.Info("newpage displayed")

	widget.stack.SetVisibleChildName(page)

	// Select a specific item in the page.
	selecter, ok := interface{}(widget.current.Widget).(Selecter) // Detect if the widget can Select.
	if len(item) > 0 && ok {
		return selecter.Select(item[0])
	}
	return false
}

// OnSelectPage reacts when the page is changed to toggle OnHide and OnShow additional callbacks.
//
func (widget *GuiConfigure) OnSelectPage(page string) {
	if widget.current != nil && widget.current.OnHide != nil { // Hide previous page.
		widget.current.OnHide()
	}

	widget.log.Info("GuiConfigure OnSelectPage")

	current, ok := widget.pages[page] // Set new current.
	if !ok {
		return
	}
	widget.current = current

	if widget.current.OnShow != nil { // Show new page.
		widget.current.OnShow()
	}
}

// SetActionNone disables the action button.
//
func (widget *GuiConfigure) SetActionNone() {
	widget.Menu.Save.Hide()
}

// SetActionSave sets the action button with save text.
//
func (widget *GuiConfigure) SetActionSave() {
	widget.Menu.Save.SetLabel(tran.Slate("Save"))
	widget.Menu.Save.Show()
}

// SetActionAdd sets the action button with add text.
//
func (widget *GuiConfigure) SetActionAdd() {
	widget.Menu.Save.SetLabel(tran.Slate("Add"))
	widget.Menu.Save.Show()
}

// SetActionGrab sets the action button with grab text.
//
func (widget *GuiConfigure) SetActionGrab() {
	widget.Menu.Save.SetLabel(tran.Slate("Grab"))
	widget.Menu.Save.Show()
}

// SetActionCancel sets the action button with cancel text.
//
func (widget *GuiConfigure) SetActionCancel() {
	widget.Menu.Save.SetLabel(tran.Slate("Cancel"))
	widget.Menu.Save.Show()
}

//
//------------------------------------------------------[ DOCK GUI CALLBACKS ]--

// ReloadItems refreshes the icons page list (clear and reselect, or select cached).
//
func (widget *GuiConfigure) ReloadItems() {
	// sel := widget.pages[GroupIcons].Selected()
	icons := interface{}(widget.pages[GroupIcons].Widget).(Clearer)
	path := icons.Clear()
	// widget.log.DEV("ReloadItems path to reselect", path)
	if widget.iconToSelect != "" {
		path = widget.iconToSelect
		widget.iconToSelect = ""
	}
	icons.Load()
	icons.Select(path)
	// widget.log.DEV("ReloadItems finished")
}

// UpdateModulesList updates listed references of applets.
//
func (widget *GuiConfigure) UpdateModulesList() {
	widget.log.Info("UpdateModulesList test")
	// w, ok := widget.pages[GroupAdd].Widget.(*confapplets.ConfApplet)
	// if ok {
	// 	w.Clear()
	// 	w.Load()
	// }
}

// UpdateModuleState updates the state of the given applet.
//
func (widget *GuiConfigure) UpdateModuleState(name string, active bool) {
	widget.log.Info("TODO: UpdateModuleState", name, active)

	w, ok := widget.pages[GroupAdd].Widget.(datatype.UpdateModuleStater)
	if ok {
		w.UpdateModuleState(name, active)
	}

	w, ok = widget.pages[GroupConfig].Widget.(datatype.UpdateModuleStater)
	if ok {
		w.UpdateModuleState(name, active)
	}
}

// UpdateShortkeys updates the shortkeys widget.
//
func (widget *GuiConfigure) UpdateShortkeys() {
	w, ok := widget.pages[GroupConfig].Widget.(UpdateShortkeyser)
	if ok { // Use pages instead of current as the widget can be loaded but not visible.
		w.UpdateShortkeys()
	}
}
