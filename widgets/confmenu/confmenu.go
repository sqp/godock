// Package confmenu provides a menu widget for the GUI.
//
// Only contains save and close buttons but can embed more widgets (as a box).
package confmenu

import (
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/godock/widgets/gtk/newgtk"
)

// Controller defines methods used on the main widget / data source by this widget and its sons.
//
type Controller interface {
	ClickedSave()
	ClickedQuit()
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
	buttonQuit := newgtk.ButtonWithMnemonic("_Close")

	/// Actions
	wmb.Save.Connect("clicked", wmb.control.ClickedSave)
	buttonQuit.Connect("clicked", wmb.control.ClickedQuit)

	/// Packing: End list = reversed.

	wmb.PackEnd(buttonQuit, false, false, 0)
	wmb.PackEnd(sep, false, false, 3) // separator add 3x2px.
	wmb.PackEnd(wmb.Save, false, false, 0)
	return wmb
}
