// Package menus builds gtk menus.
package menus

import (
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"github.com/sqp/godock/widgets/common"
	"github.com/sqp/godock/widgets/gtk/newgtk"

	"errors"
	"reflect"
)

//
//--------------------------------------------------------------------[ MENU ]--

// CallNewItem defines a func that creates a menu item and packs it to the menu.
//
type CallNewItem func(menu *gtk.Menu, label, iconPath string) *gtk.MenuItem

// CallNewSubMenu defines a func that creates a submenu and packs it to the menu.
//
type CallNewSubMenu func(menu *gtk.Menu, label, iconPath string) (*gtk.Menu, *gtk.MenuItem)

// Menu builds gtk menus. Can be loaded manually or using MenuNewAttached.
//
type Menu struct {
	gtk.Menu
	groups map[int]*glib.SList // Radio items groups reference. Indexed by user given group key.

	callNewItem    CallNewItem
	callNewSubMenu CallNewSubMenu
}

// NewMenu creates a gtk menu builder.
//
// Warning, don't forget to use ShowAll when you add entries manually.
//
func NewMenu(items ...interface{}) *Menu {
	gtkmenu := newgtk.Menu()
	menu := WrapMenu(gtkmenu)
	menu.AddList(items...)
	menu.ShowAll()
	return menu
}

// WrapMenu wraps a gtk menu as a menu builder.
//
func WrapMenu(menu *gtk.Menu) *Menu {
	return &Menu{
		Menu:   *menu,
		groups: make(map[int]*glib.SList),

		callNewItem: func(menu *gtk.Menu, label, iconPath string) *gtk.MenuItem {
			item := newgtk.MenuItemWithLabel(label)
			menu.Append(item)
			return item
		},

		callNewSubMenu: func(menu *gtk.Menu, label, iconPath string) (*gtk.Menu, *gtk.MenuItem) {
			gtkmenu := newgtk.Menu()
			item := newgtk.MenuItemWithLabel(label)
			menu.Append(item)
			item.SetSubmenu(gtkmenu)
			return gtkmenu, item
		},
	}
}

// SetCallNewItem overrides the menu item creation and packing method.
//
func (menu *Menu) SetCallNewItem(call CallNewItem) {
	menu.callNewItem = call
}

// SetCallNewSubMenu overrides the submenu creation and packing method.
//
func (menu *Menu) SetCallNewSubMenu(call CallNewSubMenu) {
	menu.callNewSubMenu = call
}

// AddSeparator adds a separator to the menu.
//
func (menu *Menu) AddSeparator() {
	menu.Append(newgtk.SeparatorMenuItem())
}

// AddEntry adds an item to the menu with its callback.
//
func (menu *Menu) AddEntry(label, iconPath string, call interface{}, userData ...interface{}) *gtk.MenuItem {
	item := menu.callNewItem(&menu.Menu, label, iconPath)

	if call == nil {
		item.SetSensitive(false)
	} else {
		item.Connect("activate", call, userData...)
	}
	return item
}

// AddCheckEntry adds a check entry to the menu.
//
func (menu *Menu) AddCheckEntry(label string, active bool, call interface{}, userData ...interface{}) (item *gtk.CheckMenuItem) {
	item = newgtk.CheckMenuItemWithLabel(label)
	item.SetActive(active)
	if call != nil {
		item.Connect("toggled", call, userData...)
	}
	menu.Append(item)
	return item
}

// AddRadioEntry adds a radio entry to the menu.
//
func (menu *Menu) AddRadioEntry(label string, active bool, groupID int, call interface{}, userData ...interface{}) (item *gtk.RadioMenuItem) {
	group := menu.groups[groupID]

	item = newgtk.RadioMenuItemWithLabel(group, label)
	if group == nil {
		var e error
		group, e = item.GetGroup()
		if e == nil {
			menu.groups[groupID] = group
		} else {
			println("Menu.AddRadioEntry", "GetGroup", e.Error())
		}
	}
	item.SetActive(active)
	if call != nil {
		item.Connect("toggled", func() {
			if item.GetActive() {
				switch f := call.(type) {
				case func():
					f()
				}
			}
		}) //  userData...)
	}
	menu.Append(item)
	return item
}

// AddSubMenu adds a submenu to the menu.
//
func (menu *Menu) AddSubMenu(title, iconPath string) *Menu {
	gtkmenu, _ := menu.callNewSubMenu(&menu.Menu, title, iconPath)
	return WrapMenu(gtkmenu)
}

// AddSubMenuWrapped adds the provided gtk menu as submenu.
//
func (menu *Menu) AddSubMenuWrapped(title string, gtkmenu *gtk.Menu) {
	gtkItem := menu.callNewItem(&menu.Menu, title, "") // iconPath
	gtkItem.SetSubmenu(gtkmenu)
}

// func MenuNewAttached(widget Clickable, items ...interface{}) *Menu {
// 	menu := MenuNew(items...)
// 	if widget != nil {
// 		// menu.AttachToWidget(widget, nil)
// 		// widget.Clicked(func() { menu.Popup(nil, nil, nil, nil, 0, 0) })
// 	}
// 	return menu
// }

//
//-------------------------------------------------------------------[ ITEMS ]--

// AddList adds a list of custom entries to the menu.
//
// items can be of type:
//   nil : a separator.
//   Check : a checkable entry.
//   Item : a simple entry with Title and an optional argument.
//
func (menu *Menu) AddList(items ...interface{}) {
	for _, uncast := range items {
		switch entry := uncast.(type) {

		case nil:
			menu.AddSeparator()

		case Check:
			menu.AddCheckEntry(entry.Title, entry.Active, entry.Call)

		case Item:
			switch call := entry.Call.(type) {

			case nil, func():
				menu.AddEntry(entry.Title, "", call)

			case []interface{}:
				menu.AddSubMenu(entry.Title, "")

			case *gtk.Menu:
				menu.AddSubMenuWrapped(entry.Title, call)

			default:
				println("menu add, unknown Call type", reflect.TypeOf(call))
			}

		default:
			println("menu add, unknown entry type", reflect.TypeOf(entry))
		}
	}
}

// Item defines a simple entry description with Title and an optional argument.
//
// Call can be of type:
//     nil              visible but disabled option (can't be clicked).
//     func()           callback to connect to the activate event.
//     []interface{}    list of items for a submenu. Recursive call so the same options are allowed.
//
type Item struct {
	Title string
	Call  interface{}
}

// NewItem creates a menu item description.
//
func NewItem(title string, call interface{}) Item {
	return Item{
		Title: title,
		Call:  call,
	}
}

// Check defines a check item description.
//
type Check struct {
	Title  string
	Call   interface{}
	Active bool
}

//
//-----------------------------------------------------------[ BUTTONS ENTRY ]--

// AddButtonsEntry adds a button entry to the menu.
//
func (menu *Menu) AddButtonsEntry(label string) *ButtonsEntry {
	entry := NewButtonsEntry(label)
	menu.Append(entry)
	return entry
}

// ButtonsEntry defines a menu entry with buttons inside.
//
type ButtonsEntry struct {
	gtk.MenuItem // extends MenuItem, the main widget.

	box   *gtk.Box      // main content box.
	label *gtk.Label    // widget label.
	list  []*gtk.Button // list of buttons.
	img   []*gtk.Image  // list of images inside buttons (same key).
}

// NewButtonsEntry creates a menu entry with buttons management.
//
func NewButtonsEntry(text string) *ButtonsEntry {
	be := &ButtonsEntry{
		MenuItem: *newgtk.MenuItem(),
		box:      newgtk.Box(gtk.ORIENTATION_HORIZONTAL, 1),
		label:    newgtk.Label(text),
	}

	// Packing.
	be.Add(be.box)
	be.box.PackStart(be.label, false, false, 0)

	// Forward click to inside buttons.
	be.Connect("button-press-event", be.onMenuItemPress)

	// Highlight pointed button.
	be.Connect("motion-notify-event", be.onMenuItemMotionNotify)

	// Turn off highlight pointed button when we leave the menu-item.
	// if we leave it quickly, a motion event won't be generated.
	be.Connect("leave-notify-event", be.onMenuItemLeave)

	// Force the label to not highlight.
	// it gets highlighted, even if we overwrite the motion_notify_event callback.
	be.Connect("enter-notify-event", be.onMenuItemEnter)

	// We don't want to higlighted the whole menu-item , but only the currently
	// pointed button; so we draw the menu-item ourselves, with a propagate to
	// childs and intercept signal.
	be.Connect("draw", func(_ *gtk.MenuItem, cr *cairo.Context) bool {
		be.PropagateDraw(be.box, cr)
		return true
	})

	return be
}

// AddButton adds a button to the entry.
//
func (o *ButtonsEntry) AddButton(tooltip, img string, call interface{}) *gtk.Button {
	btn := newgtk.Button()
	btn.SetTooltipText(tooltip)
	btn.Connect("clicked", call)
	o.box.PackEnd(btn, false, false, 0)

	if img != "" {

		// 		if (*gtkStock == '/')
		// 			int size = cairo_dock_search_icon_size (GTK_ICON_SIZE_MENU);

		image, e := common.ImageNewFromFile(img, 12) // TODO: icon size
		if e == nil {
			btn.SetImage(image)
		}
		o.img = append(o.img, image)
	} else {
		o.img = append(o.img, nil)
	}

	o.list = append(o.list, btn)
	return btn
}

func (o *ButtonsEntry) onMenuItemPress(_ *gtk.MenuItem, event *gdk.Event) bool { // GdkEventCrossing
	// Position of the mouse relatively to the menu-item.
	eventBtn := &gdk.EventButton{Event: event}
	mouseX, mouseY := int(eventBtn.X()), int(eventBtn.Y())

	sel, e := o.findButtonHovered(mouseX, mouseY)
	if e == nil {
		for i, btn := range o.list {
			if i == sel {
				o.setStateBtn(i, gtk.STATE_FLAG_ACTIVE)
				btn.Clicked()

			} else {
				o.setStateBtn(i, gtk.STATE_FLAG_NORMAL)
			}
		}
		o.QueueDraw()
	}
	return true
}

// Mouse entered the widget, force the label to be in a normal state.
func (o *ButtonsEntry) onMenuItemEnter(_ *gtk.MenuItem, event *gdk.Event) bool { // GdkEventCrossing
	o.label.SetStateFlags(gtk.STATE_FLAG_NORMAL, true)
	o.label.QueueDraw()
	return false
}

func (o *ButtonsEntry) onMenuItemLeave(_ *gtk.MenuItem, event *gdk.Event) bool { // GdkEventCrossing
	for i := range o.list {
		o.setStateBtn(i, gtk.STATE_FLAG_NORMAL)
	}
	o.box.QueueDraw()
	return false
}

func (o *ButtonsEntry) onMenuItemMotionNotify(_ *gtk.MenuItem, event *gdk.Event) bool {
	// Position of the mouse relatively to the menu-item.
	eventBtn := &gdk.EventButton{Event: event} // GdkEventMotion
	mouseX, mouseY := int(eventBtn.X()), int(eventBtn.Y())

	sel, e := o.findButtonHovered(mouseX, mouseY)
	if e != nil {
		for i := range o.list {
			if i == sel {
				// the mouse is inside the button -> select it
				o.setStateBtn(i, gtk.STATE_FLAG_PRELIGHT)

			} else {
				// else deselect it, in case it was selected
				o.setStateBtn(i, gtk.STATE_FLAG_NORMAL)
			}
		}

		// needed ?
		// force the label to be in a normal state
		o.label.SetStateFlags(gtk.STATE_FLAG_NORMAL, true)
		o.label.QueueDraw()
	}
	return false
}

func (o *ButtonsEntry) findButtonHovered(mouseX, mouseY int) (int, error) {
	for i, btn := range o.list {
		// Position of the top-left corner of the button relatively to the menu-item.
		x, y, e := btn.TranslateCoordinates(o, 0, 0)
		w, h := btn.GetAllocatedWidth(), btn.GetAllocatedHeight()
		if e != nil {
			// if logger.Err(e, "MenuItemMotionNotify btn", i) {
			continue
		}

		if x < mouseX && mouseX < x+w && y < mouseY && mouseY < y+h {
			// the mouse is inside the button -> select it
			return i, nil
		}
	}
	return -1, errors.New("button not found")
}

func (o *ButtonsEntry) setStateBtn(key int, state gtk.StateFlags) {
	o.list[key].SetStateFlags(state, true)
	if o.img[key] != nil {
		o.img[key].SetStateFlags(state, true)
	}
}
