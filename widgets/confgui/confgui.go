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

	"github.com/sqp/godock/libs/log"

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
func NewStandalone(data datatype.Source, path ...string) {
	gtk.Init(nil)

	widget, win := NewConfigWindow(data)
	win.Connect("destroy", gtk.MainQuit)

	widget.Load()
	// widget.Menu.Switcher.Activate("Icons")

	if len(path) > 0 {
		widget.SelectIcons(path[0])
	}

	gtk.Main()

	log.Info("GUI QUITTED OK !!")
	win.Destroy()
}

// NewConfigWindow creates a new config widget and window, ready to use.
//
func NewConfigWindow(data datatype.Source) (*GuiConfigure, *gtk.Window) {
	widget := NewGuiConfigure()
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		// log.Fatal("Unable to create window:", err)
		return nil, nil
	}
	win.SetDefaultSize(WindowWidth, WindowHeight)
	win.Add(widget)

	win.SetTitle(WindowTitle)
	win.SetWMClass(WindowClass, WindowTitle)

	// win.Set("border-width", 4)

	win.ShowAll()
	widget.SetWindow(win)
	widget.SetDataSource(data)
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

// SetWindower defines the interface to set the pointer to the main window of a config page.
//
type SetWindower interface {
	SetWindow(*gtk.Window)
}

//-------------------------------------------------[  ]--

// GuiConfigure defines the main Cairo-Dock configuration widget.
//
type GuiConfigure struct {
	gtk.Box

	datatype.Source // embeds the data source.

	Menu *confmenu.MenuBar

	stack *gtk.Stack

	OnQuit func() // On clicked Quit callback.

	iconToSelect string // Cache for the icon name to select as ReloadItems is called after the display (fix case new item).

	pages   map[string]*Page
	current *Page
}

// NewGuiConfigure creates the main Cairo-Dock configuration widget.
//
func NewGuiConfigure() *GuiConfigure {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)

	widget := &GuiConfigure{
		Box:   *box,
		pages: make(map[string]*Page),
	}

	// Create widgets.

	widget.Menu = confmenu.New(widget)
	menuIcons := pageswitch.New()
	menuDownload := confapplets.NewMenuDownload()
	menuIcons.Set("no-show-all", true)
	menuDownload.ShowAll()
	menuDownload.Hide()
	menuDownload.Set("no-show-all", true)

	// Box for separator on left of menuIcons.
	boxIcons, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	sepIcons, _ := gtk.SeparatorNew(gtk.ORIENTATION_VERTICAL)
	boxIcons.PackStart(sepIcons, false, false, 6)
	boxIcons.PackStart(menuIcons, false, false, 0)

	widget.stack, _ = gtk.StackNew()

	sw, _ := gtk.StackSwitcherNew()
	sw.SetStack(widget.stack)
	sw.SetHomogeneous(false)

	icons := conficons.New(widget, menuIcons)
	core := confcore.New(widget, menuIcons)
	add := confapplets.New(widget, nil, confapplets.ListCanAdd)

	// dl := confapplets.New(widget, menuDownload, confapplets.ListExternal)

	// Add pages to the switcher. This will pack the pages widgets to the gui box.

	widget.AddPage(icons, GroupIcons, widget.Menu.Save.Show, widget.Menu.Save.Hide)
	widget.AddPage(add, GroupAdd, widget.Menu.Add.Show, widget.Menu.Add.Hide)
	widget.AddPage(core, GroupConfig, widget.Menu.Save.Show, widget.Menu.Save.Hide)

	widget.stack.Connect("notify::visible-child-name", func() {
		widget.OnSelectPage(widget.stack.GetVisibleChildName())
	})

	// Packing menu.

	sep, _ := gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)

	widget.Menu.PackStart(sw, false, false, 0)
	widget.Menu.PackStart(boxIcons, false, false, 0)
	widget.Menu.PackStart(menuDownload, false, false, 0)

	widget.PackStart(widget.Menu, false, false, 2)
	widget.PackStart(sep, false, false, 0)

	widget.PackStart(widget.stack, true, true, 0)

	return widget
}

func (widget *GuiConfigure) AddPage(saver Saver, name string, onShow, onHide func()) {
	widget.stack.AddTitled(saver, name, name)
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
	for _, page := range widget.pages {
		if windower, ok := interface{}(page.Widget).(SetWindower); ok { // Detect if the widget can SetWindow.
			windower.SetWindow(win)
		}
	}
}

// SetWindow sets the pointer to the data source, needed for every widget.
//
func (widget *GuiConfigure) SetDataSource(source datatype.Source) {
	widget.Source = source

	// Load GUI own config page settings.
	e := confsettings.Init(source.DirAppData())
	log.Err(e, "Load ConfigSettings")
}

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

// ReloadItems refreshes the icons page list (clear and reselect, or select cached).
//
func (widget *GuiConfigure) ReloadItems() {
	// sel := widget.pages[GroupIcons].Selected()
	icons := interface{}(widget.pages[GroupIcons].Widget).(Clearer)
	path := icons.Clear()
	if widget.iconToSelect != "" {
		path = widget.iconToSelect
		widget.iconToSelect = ""
	}
	icons.Load()
	icons.Select(path)
}

// SelectIcons selects a specific icon in the Icons page (key = full path to config file).
// If the icon isn't found, the name is cached for the late ReloadItems callback.
//
func (widget *GuiConfigure) SelectIcons(item string) {
	log.DEV("SelectIcons", item)
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
	log.DEV("newpage displayed")

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

	log.DEV("GuiConfigure OnSelectPage")

	current, ok := widget.pages[page] // Set new current.
	if !ok {
		return
	}
	widget.current = current

	if widget.current.OnShow != nil { // Show new page.
		widget.current.OnShow()
	}
}

// iconsPage := &pageswitch.Page{
// 	Widget: icons,
// 	Name:   "Icons",
// 	// icon:"",
// 	OnLoad: func() { icons.Load() },
// 	OnShow: func() {
// 		icons.Show()
// 		widget.Menu.SetSaveVisible(true)
// 		boxIcons.Show()
// 		widget.current = icons
// 	},
// 	OnHide: func() {
// 		icons.Hide()
// 		widget.Menu.SetSaveVisible(false)
// 		boxIcons.Hide()
// 	},
// }

// corePage := &pageswitch.Page{
// 	Widget: core,
// 	Name:   "Config",
// 	// icon:"",
// 	OnLoad: func() { core.Load() },
// 	OnShow: func() { core.Show(); widget.Menu.SetSaveVisible(true); widget.current = core },
// 	OnHide: func() { core.Hide(); widget.Menu.SetSaveVisible(false) },
// }

// addPage := &pageswitch.Page{
// 	Widget: add,
// 	Name:   "Add",
// 	// icon:"",
// 	OnLoad: func() { add.Load() },
// 	OnShow: func() { add.Show(); widget.Menu.SetAddVisible(true); widget.current = add },
// 	OnHide: func() { add.Hide(); widget.Menu.SetAddVisible(false) },
// }

// dlPage := &pageswitch.Page{
// 	Widget: dl,
// 	Name:   "Download",
// 	// icon:"",
// 	// OnLoad: func() { dl.Load() },
// 	OnShow: func() { dl.Show(); menuDownload.Show() }, // need to add widget.current = dl ??
// 	OnHide: func() { dl.Hide(); menuDownload.Hide() },

// widget.addPage(iconsPage, corePage, addPage, dlPage)

// func (widget *GuiConfigure) addPage(pages ...*pageswitch.Page) {
// 	for _, page := range pages {
// 		page.Widget.Set("no-show-all", true)
// 		widget.Menu.Switcher.AddPage(page)
// 		widget.PackStart(page.Widget, true, true, 0)
// 	}
// }
