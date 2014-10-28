// Package confmenu provides a menu widget for the GUI.
//
// Embeds the switcher and add and save buttons.
package confmenu

import (
	"github.com/conformal/gotk3/gtk"
	// "github.com/sqp/godock/widgets/pageswitch"
)

// GUIControl defines external control callbacks for the GUI.
//
type GUIControl interface {
	ClickedSave()
	ClickedQuit()
}

//
//----------------------------------------------------------------[ GUI MENU ]--

// MenuBar is the config window menu.
//
type MenuBar struct {
	gtk.Box // Container is first level. Act as (at least) a GtkBox.

	// Switcher *pageswitch.Switcher

	Save *gtk.Button
	Add  *gtk.Button

	control GUIControl // interface to controler
}

// New creates the config menu with add or save buttons.
//
func New(control GUIControl) *MenuBar {

	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	if box == nil {
		return nil
	}

	wmb := &MenuBar{
		Box:     *box,
		control: control,
	}

	wmb.Save, _ = gtk.ButtonNewWithMnemonic("_Save")
	wmb.Add, _ = gtk.ButtonNewWithMnemonic("_Add")

	wmb.Add.Set("no-show-all", true)
	wmb.Add.Hide()

	sep, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	buttonQuit, _ := gtk.ButtonNewWithMnemonic("_Close")

	// wmb.Switcher = pageswitch.New()

	/// Actions
	wmb.Save.Connect("clicked", wmb.control.ClickedSave)
	wmb.Add.Connect("clicked", wmb.control.ClickedSave)
	buttonQuit.Connect("clicked", wmb.control.ClickedQuit)

	/// Packing: End list = reversed.
	// wmb.PackStart(wmb.Switcher, false, false, 0)

	wmb.PackEnd(buttonQuit, false, false, 0)
	wmb.PackEnd(sep, false, false, 3) // separator add 3x2px.
	wmb.PackEnd(wmb.Save, false, false, 0)
	wmb.PackEnd(wmb.Add, false, false, 0)
	return wmb
}

// SetSaveVisible sets the visibility of the save button.
//
// func (wmb *MenuBar) SetSaveVisible(visible bool) {
// 	wmb.Save.SetVisible(visible)
// }

// // SetAddVisible sets the visibility of the add button.
// //
// func (wmb *MenuBar) SetAddVisible(visible bool) {
// 	wmb.Add.SetVisible(visible)
// }
