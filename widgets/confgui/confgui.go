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
	"github.com/sqp/godock/libs/text/tran"

	"github.com/sqp/godock/widgets/cfbuild/datatype"
	"github.com/sqp/godock/widgets/confapplets"
	"github.com/sqp/godock/widgets/confcore"
	"github.com/sqp/godock/widgets/confgui/btnaction"
	"github.com/sqp/godock/widgets/conficons"
	"github.com/sqp/godock/widgets/confmenu"
	"github.com/sqp/godock/widgets/confsettings"
	"github.com/sqp/godock/widgets/gtk/newgtk"
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

//
//------------------------------------------------------[ WIDGETS INTERFACES ]--

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

// ShowWelcomer defines the optional interface to show a placeholder page.
//
type ShowWelcomer interface {
	ShowWelcome(setBtn bool)
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

// UpdateDeskletVisibilityer defines the interface to update shortkeys of a config page.
//
type UpdateDeskletVisibilityer interface {
	UpdateDeskletVisibility(icon datatype.Iconer)
}

// UpdateDeskletParamser defines the interface to update shortkeys of a config page.
//
type UpdateDeskletParamser interface {
	UpdateDeskletParams(icon datatype.Iconer)
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

	btnAction map[string]btnaction.Tune
	pages     map[string]*Page
	current   *Page

	log cdtype.Logger
}

// NewWidget creates the main Cairo-Dock configuration widget.
//
func NewWidget(source datatype.Source, log cdtype.Logger) *GuiConfigure {
	box := newgtk.Box(gtk.ORIENTATION_VERTICAL, 0)

	widget := &GuiConfigure{
		Box:    *box,
		Source: source,

		stack:     newgtk.Stack(),
		btnAction: make(map[string]btnaction.Tune),
		pages:     make(map[string]*Page),
		log:       log,
	}

	// Load GUI own config page settings.
	e := confsettings.Init(source.DirUserAppData(confsettings.GuiFilename))
	log.Err(e, "Load ConfigSettings")

	// Create widgets.

	widget.Menu = confmenu.New(widget)
	menuIcons := pageswitch.New()
	menuIcons.Set("no-show-all", true)
	menuCore := pageswitch.New()
	menuCore.Set("no-show-all", true)

	// Box for separator on left of menuIcons.
	boxIcons := newgtk.Box(gtk.ORIENTATION_HORIZONTAL, 0)
	sepIcons := newgtk.Separator(gtk.ORIENTATION_VERTICAL)
	boxIcons.PackStart(sepIcons, false, false, 6)
	boxIcons.PackStart(menuIcons, false, false, 0)
	boxIcons.PackStart(menuCore, false, false, 0)

	sw := newgtk.StackSwitcher()
	sw.SetStack(widget.stack)
	sw.SetHomogeneous(false)

	btnIcons := btnaction.New(widget.Menu.Save)
	btnCore := btnaction.New(widget.Menu.Save)
	btnAdd := btnaction.New(widget.Menu.Save)
	btnAdd.SetAdd()

	icons := conficons.New(widget, log, menuIcons, btnIcons)
	core := confcore.New(widget, log, menuCore, btnCore)
	add := confapplets.New(widget, log, nil, confapplets.ListCanAdd)
	add.Hide() // TODO: REMOVE THE NEED OF THAT.

	// Add pages to the switcher. This will pack the pages widgets to the gui box.

	widget.AddPage(GroupIcons, icons, btnIcons, menuIcons.Show, menuIcons.Hide)
	widget.AddPage(GroupAdd, add, btnAdd, nil, nil)
	widget.AddPage(GroupConfig, core, btnCore, menuCore.Show, menuCore.Hide)

	// widget.stack.ChildSetProperty(core, "icon-name", "cairo-dock")
	widget.stack.ChildSetProperty(add, "icon-name", "list-add")

	widget.stack.Connect("notify::visible-child-name", func() {
		widget.OnSelectPage(widget.stack.GetVisibleChildName())
	})

	// Packing menu.

	sep := newgtk.Separator(gtk.ORIENTATION_HORIZONTAL)

	widget.Menu.PackStart(sw, false, false, 0)
	widget.Menu.PackStart(boxIcons, false, false, 0)

	widget.PackStart(widget.Menu, false, false, 2)
	widget.PackStart(sep, false, false, 0)

	widget.PackStart(widget.stack, true, true, 0)

	return widget
}

//
//-------------------------------------------------------------[ CONFIG PAGE ]--

// Page defines a switcher page.
//
type Page struct {
	Name   string
	Widget Saver
	OnShow func()
	OnHide func()
	btn    btnaction.Tune
}

// AddPage adds a tab to the main config switcher with its widget.
//
func (widget *GuiConfigure) AddPage(name string, saver Saver, btn btnaction.Tune, onShow, onHide func()) {
	widget.stack.AddTitled(saver, name, tran.Slate(name))
	widget.pages[name] = &Page{
		Name:   name,
		Widget: saver,
		OnShow: onShow,
		OnHide: onHide,
		btn:    btn,
	}
}

// Load loads all pages data.
//
func (widget *GuiConfigure) Load() {
	for _, page := range widget.pages {
		page.Widget.Load()
	}
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
	// Show placeholders if needed.
	defer widget.pages[GroupIcons].Widget.(ShowWelcomer).ShowWelcome(false)
	defer widget.pages[GroupConfig].Widget.(ShowWelcomer).ShowWelcome(false)

	// widget.log.Info("newpage selected", page, item)

	child, ok := widget.pages[page]
	if !ok {
		widget.log.Info("config select, no matching page:", page, item)
		return false
	}
	widget.stack.SetVisibleChild(child.Widget)

	if widget.current == nil {
		widget.log.Info("config select", "no page selected", page)
		return false
	}

	// Select a specific item in the page.
	if len(item) > 0 {
		selecter, ok := widget.current.Widget.(Selecter) // Detect if the widget can Select.
		if ok {
			return selecter.Select(item[0])
		}
		widget.log.Info("config select: no selecter", page, item)
	}

	return false
}

// OnSelectPage reacts when the page is changed to set the button state and
// trigger OnHide and OnShow additional callbacks.
//
func (widget *GuiConfigure) OnSelectPage(page string) {
	if widget.current != nil {
		widget.current.btn.Hide()
		if widget.current.OnHide != nil { // Hide previous page.
			widget.current.OnHide()
		}
	}

	widget.log.Debug("OnSelectPage", page)

	current, ok := widget.pages[page] // Set new current.
	if !ok {
		return
	}
	widget.current = current

	widget.current.btn.Display()
	if widget.current.OnShow != nil { // Show new page.
		widget.current.OnShow()
	}
}

//
//------------------------------------------------------[ DOCK GUI CALLBACKS ]--

// WARNING: Callbacks must use pages instead of current.
// Widgets can be loaded but not visible. You must act on every needed page.

// ReloadItems refreshes the icons page list (clear and reselect, or select cached).
//
func (widget *GuiConfigure) ReloadItems() {
	icons := widget.pages[GroupIcons].Widget.(Clearer)
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
	if ok {
		w.UpdateShortkeys()
	}
}

// UpdateDeskletParams updates applet desklet settings if it's loaded.
//
func (widget *GuiConfigure) UpdateDeskletParams(icon datatype.Iconer) {
	widget.log.Info("TODO: UpdateDeskletParams")
	w, ok := widget.pages[GroupIcons].Widget.(UpdateDeskletParamser)
	if ok {
		w.UpdateDeskletParams(icon)
	}
}

// UpdateDeskletVisibility updates applet desklet settings if it's loaded.
//
func (widget *GuiConfigure) UpdateDeskletVisibility(icon datatype.Iconer) {
	widget.log.Info("TODO: UpdateDeskletVisibility")
	w, ok := widget.pages[GroupIcons].Widget.(UpdateDeskletVisibilityer)
	if ok {
		w.UpdateDeskletVisibility(icon)
	}
}

//
//------------------------------------------------------------------[ WINDOW ]--

// SetWindow sets the pointer to the parent window, used for some config
// callbacks (grab events).
//
func (widget *GuiConfigure) SetWindow(win *gtk.Window) {
	widget.window = win
	widget.OnQuit = win.Destroy
}

// GetWindow returns the pointer to the parent window.
//
func (widget *GuiConfigure) GetWindow() *gtk.Window {
	return widget.window
}

// NewWindow creates a new config window with its widget, ready to use.
//
func NewWindow(data datatype.Source, log cdtype.Logger) (*GuiConfigure, error) {
	win, e := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if e != nil {
		return nil, e
	}
	win.SetDefaultSize(WindowWidth, WindowHeight)
	win.SetTitle(WindowTitle)
	win.SetWMClass(WindowClass, WindowTitle)

	win.SetIconFromFile(data.AppIcon())

	widget := NewWidget(data, log)
	widget.SetWindow(win)

	win.Add(widget)
	win.ShowAll()

	return widget, nil
}

// NewStandalone creates a new config window to use as standalone application.
//
// func NewStandalone(data datatype.Source, log cdtype.Logger, path ...string) {
// 	gtk.Init(nil)

// 	widget, win := NewWindow(data, log)
// 	win.Connect("destroy", gtk.MainQuit)

// 	widget.Load()
// 	// widget.Menu.Switcher.Activate("Icons")

// 	if len(path) > 0 {
// 		widget.SelectIcons(path[0])
// 	}

// 	gtk.Main()

// 	// log.Info("GUI QUITTED OK !!")
// 	win.Destroy()
// }
