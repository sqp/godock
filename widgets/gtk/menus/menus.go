// Package menus builds gtk menus.
package menus

import (
	"github.com/conformal/gotk3/glib"
	"github.com/conformal/gotk3/gtk"

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
	gtkmenu, _ := gtk.MenuNew()
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
			item, _ := gtk.MenuItemNewWithLabel(label)
			menu.Append(item)
			return item
		},

		callNewSubMenu: func(menu *gtk.Menu, label, iconPath string) (*gtk.Menu, *gtk.MenuItem) {
			gtkmenu, _ := gtk.MenuNew()
			item, _ := gtk.MenuItemNewWithLabel(label)
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
	sep, _ := gtk.SeparatorMenuItemNew()
	menu.Append(sep)
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
	item, _ = gtk.CheckMenuItemNewWithLabel(label)
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
	group, _ := menu.groups[groupID]

	item, _ = gtk.RadioMenuItemNewWithLabel(group, label)
	if group == nil {
		var e error
		group, e = item.GetGroup()
		if e == nil {
			menu.groups[groupID] = group
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
