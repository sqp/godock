// Package confshortkeys provides a dock shortkey configuration widget.
package confshortkeys

import (
	"github.com/conformal/gotk3/gdk"
	"github.com/conformal/gotk3/glib"
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/godock/libs/cdtype"

	"github.com/sqp/godock/widgets/common"
	"github.com/sqp/godock/widgets/confbuilder/datatype"
	"github.com/sqp/godock/widgets/confsettings"
	"github.com/sqp/godock/widgets/gtk/buildhelp"
	"github.com/sqp/godock/widgets/gtk/gunvalue"
	"github.com/sqp/godock/widgets/gtk/keyfile"
)

//--------------------------------------------------------[ WIDGET SHORTKEYS ]--

// Rows defines liststore rows. Must match the ListStore declaration type and order.
//
const (
	rowIcon = iota
	rowDemander
	rowDescription
	rowShortkey
	rowColor
	rowEditable
)

// Controller defines methods used on the main widget / data source by the shortkeys widget.
//
type Controller interface {
	GetWindow() *gtk.Window
	ListShortkeys() []datatype.Shortkeyer
	SetActionGrab()
	SetActionCancel()
}

// Shortkeys defines a dock shortkeys management widget.
//
type Shortkeys struct {
	gtk.ScrolledWindow // Main widget is the container. The ScrolledWindow will handle list scrollbars.
	tree               *gtk.TreeView
	model              *gtk.ListStore
	selection          *gtk.TreeSelection
	control            Controller
	log                cdtype.Logger

	cbID glib.SignalHandle // Grab callback id.

	rows map[*gtk.TreeIter]datatype.Shortkeyer // index of iter -> shortkey.
}

// New creates a dock shortkeys management widget.
//
func New(control Controller, log cdtype.Logger) *Shortkeys {
	builder := buildhelp.NewFromBytes(confshortkeysXML())

	widget := &Shortkeys{
		ScrolledWindow: *builder.GetScrolledWindow("widget"),
		model:          builder.GetListStore("model"),
		tree:           builder.GetTreeView("tree"),
		selection:      builder.GetTreeSelection("selection"),
		control:        control,
		log:            log,
		rows:           make(map[*gtk.TreeIter]datatype.Shortkeyer),
	}

	rend := builder.GetCellRendererText("cellrenderertextShortkey")

	if len(builder.Errors) > 0 {
		for _, e := range builder.Errors {
			log.Err(e, "build confshortkeys")
		}
		return nil
	}

	// The user is allowed to edit the shortcut text. This will handle the new text.
	rend.Connect("edited", widget.onManualEdit)

	widget.Load()
	return widget
}

// Clear clears the widget data.
//
func (widget *Shortkeys) Clear() {
	if widget.cbID > 0 {
		widget.onKeyGrabFinish() // Was grabbing, cancel it, not sure what event was triggerred since (refresh was asked).
	}

	widget.rows = make(map[*gtk.TreeIter]datatype.Shortkeyer)
	widget.model.Clear()
}

// Load loads the list of dock shortkeys in the widget.
//
func (widget *Shortkeys) Load() {
	widget.Clear()

	for _, sk := range widget.control.ListShortkeys() {
		iter := widget.model.Append()
		widget.rows[iter] = sk

		widget.model.SetCols(iter, gtk.Cols{
			rowDemander:    sk.GetDemander(),
			rowDescription: sk.GetDescription(),
			rowShortkey:    sk.GetKeyString(),
			rowColor:       getColor(sk),
			rowEditable:    true}) // Editable forced for all shortkey cells.

		img := sk.GetIconFilePath()
		if pix, e := common.PixbufNewFromFile(img, 24); !widget.log.Err(e, "Load icon") {
			widget.model.SetValue(iter, rowIcon, pix)
		}
	}
}

// getColor returns the color displayed for the shortcut cell.
//
func getColor(sk datatype.Shortkeyer) string {
	switch {
	case sk.GetSuccess():
		return "#116E08"

	case sk.GetKeyString() != "": // defined but failed.
		return "#B00000"
	}
	return "#000000" // unused, who cares what color an empty text can be (still prevents logged errors).
}

// Grab starts (or stops) the grab key mode to help the user assign a new
// shortcut for the selected line, if any.
//
func (widget *Shortkeys) Grab() {
	if widget.cbID > 0 {
		widget.onKeyGrabFinish() // Was grabbing, it's a cancel.
		widget.updateDisplay()
		return
	}

	_, iter := widget.selectedShortkey()
	if iter == nil {
		return
	}

	widget.control.SetActionCancel()
	widget.SetSensitive(false)
	widget.cbID, _ = widget.control.GetWindow().Connect("key-press-event", widget.onKeyGrabReceived)

	widget.model.SetValue(iter, rowShortkey, "...press key...")
	widget.model.SetValue(iter, rowColor, "#888888")
}

// updateShortkey updates the shortkey with user input, (forwarding to the dock).
// This will trigger the UpdateShortkeys event that will take care of refreshing
// the widget display.
//
func (widget *Shortkeys) updateShortkey(accel string) {
	sk, _ := widget.selectedShortkey()
	if sk == nil {
		return
	}
	widget.log.Debug("Set new shortkey", accel)

	sk.Rebind(accel, "")

	file := sk.GetConfFilePath()
	if file == "" {
		widget.log.NewErr("shortkeys wrong filepath", sk.GetConfFilePath())
		return
	}

	if sk.GetDescription() == "-" {
		widget.log.Info("shortkeys update aren't saved to file from this page for external applets.", "You muse use the applet config page.")
		return
	}

	// TODO: improve code. need to use files.UpdateConfFile(file, sk.GetGroupName(), sk.GetKeyName(), accel)
	pKeyF, e := keyfile.NewFromFile(file, keyfile.FlagsKeepComments|keyfile.FlagsKeepTranslations)
	if widget.log.Err(e, "Update shortkey in file") {
		return
	}
	defer pKeyF.Free()
	pKeyF.Set(sk.GetGroupName(), sk.GetKeyName(), accel)

	_, str, _ := pKeyF.ToData()

	confsettings.SaveFile(file, str)

	// 				cairo_dock_update_conf_file (binding->cConfFilePath,
	// 					G_TYPE_STRING, binding->cGroupName, binding->cKeyName, key,
	// 					G_TYPE_INVALID);

}

// updateDisplay sets shortcut text after a cancelled edit.
//
func (widget *Shortkeys) updateDisplay() {
	sk, iter := widget.selectedShortkey()
	if sk == nil {
		return
	}
	widget.model.SetValue(iter, rowShortkey, sk.GetKeyString())
	widget.model.SetValue(iter, rowColor, getColor(sk))
}

// selectedShortkey returns the Shortkeyer matching the selected line.
//
func (widget *Shortkeys) selectedShortkey() (datatype.Shortkeyer, *gtk.TreeIter) {
	iter, e := gunvalue.SelectedIter(widget.model, widget.selection)
	if widget.log.Err(e, "selectedShortkey") {
		return nil, nil
	}
	for it, sk := range widget.rows {
		if it.GtkTreeIter == iter.GtkTreeIter {
			return sk, iter
		}
	}
	return nil, nil
}

//-------------------------------------------------------[ ACTIONS CALLBACKS ]--

// onManualEdit is called when the user enterred a new shortcut value manually.
//
func (widget *Shortkeys) onManualEdit(oo *gtk.CellRendererText, path, accel string) {
	widget.updateShortkey(accel)
}

// onKeyGrabReceived is called when a grabbed keyboard events is received.
//
func (widget *Shortkeys) onKeyGrabReceived(win *gtk.Window, event *gdk.Event) {
	key := &gdk.EventKey{Event: event}

	if !gtk.AcceleratorValid(key.KeyVal(), gdk.ModifierType(key.State())) {
		return
	}

	// This lets us ignore all ignorable modifier keys, including NumLock and many others. :)
	// The logic is: keep only the important modifiers that were pressed for this event.
	state := gdk.ModifierType(key.State()) & gtk.AcceleratorGetDefaultModMask()

	accel := gtk.AcceleratorName(key.KeyVal(), state)

	if accel != "Escape" { // TODO: FIX
		widget.updateShortkey(accel)

	} else {
		widget.updateDisplay()

	}
	widget.onKeyGrabFinish()
}

// onKeyGrabFinish cleans after grab or cancel (restore display).
//
func (widget *Shortkeys) onKeyGrabFinish() {
	widget.control.GetWindow().HandlerDisconnect(widget.cbID)
	widget.cbID = 0

	widget.control.SetActionGrab()
	widget.SetSensitive(true)
}

// 	int iSize = cairo_dock_search_icon_size (GTK_ICON_SIZE_LARGE_TOOLBAR);
// 	gchar *cIcon = cairo_dock_search_icon_s_path (binding->cIconFilePath, iSize);
// 	GdkPixbuf *pixbuf = gdk_pixbuf_new_from_file_at_size (cIcon, iSize, iSize, NULL);
