// Package confmenu provides a menu widget for the GUI.
//
// Only contains save and close buttons but can embed more widgets (as a box).
package confmenu

import (
	"github.com/gotk3/gotk3/gtk"

	"github.com/sqp/godock/libs/text/tran"
	"github.com/sqp/godock/widgets/about"
	"github.com/sqp/godock/widgets/gtk/menus"
	"github.com/sqp/godock/widgets/gtk/newgtk"
)

// IconSize defines the default icon size.
//
var IconSize = gtk.ICON_SIZE_SMALL_TOOLBAR

// Controller defines methods used on the main widget / data source by this widget and its sons.
//
type Controller interface {
	ClickedSave()
	ClickedQuit()
	Select(page string, item ...string) bool
}

//
//----------------------------------------------------------------[ GUI MENU ]--

// MenuBar is the config window menu.
//
type MenuBar struct {
	gtk.Box // Container is first level. Act as (at least) a GtkBox.

	Save *gtk.Button

	control Controller // interface to controler
}

// New creates the config menu with add or save buttons.
//
func New(control Controller) *MenuBar {
	wmb := &MenuBar{
		Box:     *newgtk.Box(gtk.ORIENTATION_HORIZONTAL, 0),
		Save:    newgtk.ButtonWithMnemonic("_Save"),
		control: control,
	}

	wmb.Save.Set("no-show-all", true)

	sep := newgtk.Box(gtk.ORIENTATION_HORIZONTAL, 0)

	/// Actions
	wmb.Save.Connect("clicked", wmb.control.ClickedSave)

	mainBtn, _ := gtk.MenuButtonNew()

	img := newgtk.ImageFromIconName("preferences-system", IconSize)
	mainBtn.SetImage(img)

	mainBtn.SetPopup(menus.NewMenu(
		menus.NewItem(tran.Slate("Help"), func() { control.Select("Config", "Help") }),
		menus.NewItem(tran.Slate("About"), func() { about.New() }),
		nil,
		menus.NewItem(tran.Slate("Close"), wmb.control.ClickedQuit),
	))

	/// Packing: End list = reversed.

	wmb.PackEnd(mainBtn, false, false, 0)
	wmb.PackEnd(sep, false, false, 3) // separator add 3x2px.
	wmb.PackEnd(wmb.Save, false, false, 0)
	return wmb
}
