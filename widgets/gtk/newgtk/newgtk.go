// Package newgtk creates gtk objects.
package newgtk

import (
	"github.com/conformal/gotk3/gdk"
	"github.com/conformal/gotk3/glib"
	"github.com/conformal/gotk3/gtk"

	"unsafe"
)

//
//-------------------------------------------------------------[ GLIB RECAST ]--

// gObject recast a pointer to *glib.Object.
func gObject(ptr unsafe.Pointer) *glib.Object {
	return &glib.Object{GObject: glib.ToGObject(ptr)}
}

//
//--------------------------------------------------------------[ GTK RECAST ]--

// Adjustment recast a pointer to *gtk.Adjustment.
func Adjustment(value, lower, upper, stepIncrement, pageIncrement, pageSize float64) *gtk.Adjustment {
	w, _ := gtk.AdjustmentNew(value, lower, upper, stepIncrement, pageIncrement, pageSize)
	return w
}

// Box recast a pointer to *gtk.Box.
func Box(orientation gtk.Orientation, spacing int) *gtk.Box {
	w, _ := gtk.BoxNew(orientation, spacing)
	return w
}

// Button recast a pointer to *gtk.Button.
func Button() *gtk.Button {
	w, _ := gtk.ButtonNew()
	return w
}

// ButtonFromIconName recast a pointer to *gtk.Button.
func ButtonFromIconName(label string, size gtk.IconSize) *gtk.Button {
	w, _ := gtk.ButtonNewFromIconName(label, size)
	return w
}

// ButtonWithLabel recast a pointer to *gtk.Button.
func ButtonWithLabel(label string) *gtk.Button {
	w, _ := gtk.ButtonNewWithLabel(label)
	return w
}

// ButtonWithMnemonic recast a pointer to *gtk.Button.
func ButtonWithMnemonic(label string) *gtk.Button {
	w, _ := gtk.ButtonNewWithMnemonic(label)
	return w
}

// CellRendererPixbuf creates a *gtk.CellRendererPixbuf.
func CellRendererPixbuf() *gtk.CellRendererPixbuf {
	w, _ := gtk.CellRendererPixbufNew()
	return w
}

// CellRendererText creates a *gtk.CellRendererText.
func CellRendererText() *gtk.CellRendererText {
	w, _ := gtk.CellRendererTextNew()
	return w
}

// CellRendererToggle creates a *gtk.CellRendererToggle.
func CellRendererToggle() *gtk.CellRendererToggle {
	w, _ := gtk.CellRendererToggleNew()
	return w
}

// CheckButton creates a *gtk.CheckButton.
func CheckButton() *gtk.CheckButton {
	w, _ := gtk.CheckButtonNew()
	return w
}

// CheckMenuItemWithLabel creates a *gtk.CheckMenuItem.
func CheckMenuItemWithLabel(label string) *gtk.CheckMenuItem {
	w, _ := gtk.CheckMenuItemNewWithLabel(label)
	return w
}

// ColorButtonWithRGBA creates a *gtk.ColorButton.
func ColorButtonWithRGBA(gdkColor *gdk.RGBA) *gtk.ColorButton {
	w, _ := gtk.ColorButtonNewWithRGBA(gdkColor)
	return w
}

// ComboBox creates a *gtk.ComboBox.
func ComboBox() *gtk.ComboBox {
	w, _ := gtk.ComboBoxNew()
	return w
}

// ComboBoxWithEntry creates a *gtk.ComboBox.
func ComboBoxWithEntry() *gtk.ComboBox {
	w, _ := gtk.ComboBoxNewWithEntry()
	return w
}

// ComboBoxWithModel creates a *gtk.ComboBox.
func ComboBoxWithModel(model gtk.ITreeModel) *gtk.ComboBox {
	w, _ := gtk.ComboBoxNewWithModel(model)
	return w
}

// ComboBoxText creates a *gtk.ComboBoxText.
func ComboBoxText() *gtk.ComboBoxText {
	w, _ := gtk.ComboBoxTextNew()
	return w
}

// Dialog creates a *gtk.Dialog.
func Dialog() *gtk.Dialog {
	w, _ := gtk.DialogNew()
	return w
}

// Entry creates a *gtk.Entry.
func Entry() *gtk.Entry {
	w, _ := gtk.EntryNew()
	return w
}

// Expander creates a *gtk.Expander.
func Expander(label string) *gtk.Expander {
	w, _ := gtk.ExpanderNew(label)
	return w
}

// FileChooserDialogWith2Buttons creates a *gtk.FileChooserDialog.
func FileChooserDialogWith2Buttons(title string, parent *gtk.Window, action gtk.FileChooserAction,
	firstText string, firstID gtk.ResponseType,
	secondText string, secondID gtk.ResponseType) *gtk.FileChooserDialog {

	w, _ := gtk.FileChooserDialogNewWith2Buttons(title, parent, action, firstText, firstID, secondText, secondID)
	return w
}

// FileFilter creates a *gtk.FileFilter.
func FileFilter() *gtk.FileFilter {
	w, _ := gtk.FileFilterNew()
	return w
}

// FontButtonWithFont creates a *gtk.FontButton.
func FontButtonWithFont(fontname string) *gtk.FontButton {
	w, _ := gtk.FontButtonNewWithFont(fontname)
	return w
}

// Frame creates a *gtk.Frame.
func Frame(label string) *gtk.Frame {
	w, _ := gtk.FrameNew(label)
	return w
}

// Grid creates a *gtk.Grid.
func Grid() *gtk.Grid {
	w, _ := gtk.GridNew()
	return w
}

// Image creates a *gtk.Image.
func Image() *gtk.Image {
	w, _ := gtk.ImageNew()
	return w
}

// ImageFromFile creates a *gtk.Image.
func ImageFromFile(file string) *gtk.Image {
	w, _ := gtk.ImageNewFromFile(file)
	return w
}

// ImageFromIconName creates a *gtk.Image.
func ImageFromIconName(iconName string, size gtk.IconSize) *gtk.Image {
	w, _ := gtk.ImageNewFromIconName(iconName, size)
	return w
}

// Label creates a *gtk.Label.
func Label(label string) *gtk.Label {
	w, _ := gtk.LabelNew(label)
	return w
}

// LinkButtonWithLabel creates a *gtk.LinkButton.
func LinkButtonWithLabel(uri, label string) *gtk.LinkButton {
	w, _ := gtk.LinkButtonNewWithLabel(uri, label)
	return w
}

// ListBox creates a *gtk.ListBox.
func ListBox() *gtk.ListBox {
	w, _ := gtk.ListBoxNew()
	return w
}

// ListBoxRow creates a *gtk.ListBoxRow.
func ListBoxRow() *gtk.ListBoxRow {
	w, _ := gtk.ListBoxRowNew()
	return w
}

// ListStore creates a *gtk.ListStore.
func ListStore(types ...glib.Type) *gtk.ListStore {
	w, _ := gtk.ListStoreNew(types...)
	return w
}

// Menu creates a *gtk.Menu.
func Menu() *gtk.Menu {
	w, _ := gtk.MenuNew()
	return w
}

// MenuItem creates a *gtk.MenuItem.
func MenuItem() *gtk.MenuItem {
	w, _ := gtk.MenuItemNew()
	return w
}

// MenuItemWithLabel creates a *gtk.MenuItem.
func MenuItemWithLabel(label string) *gtk.MenuItem {
	w, _ := gtk.MenuItemNewWithLabel(label)
	return w
}

// Notebook creates a *gtk.Notebook.
func Notebook() *gtk.Notebook {
	w, _ := gtk.NotebookNew()
	return w
}

// Paned creates a *gtk.Paned.
func Paned(orientation gtk.Orientation) *gtk.Paned {
	w, _ := gtk.PanedNew(orientation)
	return w
}

// RadioMenuItemWithLabel creates a *gtk.RadioMenuItem.
func RadioMenuItemWithLabel(group *glib.SList, label string) *gtk.RadioMenuItem {
	w, _ := gtk.RadioMenuItemNewWithLabel(group, label)
	return w
}

// Scale creates a *gtk.Scale.
func Scale(orientation gtk.Orientation, adjustment *gtk.Adjustment) *gtk.Scale {
	w, _ := gtk.ScaleNew(orientation, adjustment)
	return w
}

// ScaleWithRange creates a *gtk.Scale.
func ScaleWithRange(orientation gtk.Orientation, min, max, step float64) *gtk.Scale {
	w, _ := gtk.ScaleNewWithRange(orientation, min, max, step)
	return w
}

// ScrolledWindow creates a *gtk.ScrolledWindow.
func ScrolledWindow(hadjustment, vadjustment *gtk.Adjustment) *gtk.ScrolledWindow {
	w, _ := gtk.ScrolledWindowNew(hadjustment, vadjustment)
	return w
}

// Separator creates a *gtk.Separator.
func Separator(orientation gtk.Orientation) *gtk.Separator {
	w, _ := gtk.SeparatorNew(orientation)
	return w
}

// SeparatorMenuItem creates a *gtk.SeparatorMenuItem.
func SeparatorMenuItem() *gtk.SeparatorMenuItem {
	w, _ := gtk.SeparatorMenuItemNew()
	return w
}

// SpinButtonWithRange creates a *gtk.SpinButton.
func SpinButtonWithRange(min, max, step float64) *gtk.SpinButton {
	w, _ := gtk.SpinButtonNewWithRange(min, max, step)
	return w
}

// Spinner creates a *gtk.Spinner.
func Spinner() *gtk.Spinner {
	w, _ := gtk.SpinnerNew()
	return w
}

// Stack creates a *gtk.Stack.
func Stack() *gtk.Stack {
	w, _ := gtk.StackNew()
	return w
}

// StackSwitcher creates a *gtk.StackSwitcher.
func StackSwitcher() *gtk.StackSwitcher {
	w, _ := gtk.StackSwitcherNew()
	return w
}

// Switch creates a *gtk.Switch.
func Switch() *gtk.Switch {
	w, _ := gtk.SwitchNew()
	return w
}

// TextView creates a *gtk.TextView.
func TextView() *gtk.TextView {
	w, _ := gtk.TextViewNew()
	return w
}

// ToggleButton creates a *gtk.ToggleButton.
func ToggleButton() *gtk.ToggleButton {
	w, _ := gtk.ToggleButtonNew()
	return w
}

// ToggleButtonWithLabel creates a *gtk.ToggleButton.
func ToggleButtonWithLabel(label string) *gtk.ToggleButton {
	w, _ := gtk.ToggleButtonNewWithLabel(label)
	return w
}

// TreeViewWithModel creates a *gtk.TreeView.
func TreeViewWithModel(model gtk.ITreeModel) *gtk.TreeView {
	w, _ := gtk.TreeViewNewWithModel(model)
	return w
}

// TreeViewColumnWithAttribute creates a *gtk.TreeViewColumn.
func TreeViewColumnWithAttribute(title string, renderer gtk.ICellRenderer, attribute string, column int) *gtk.TreeViewColumn {
	w, _ := gtk.TreeViewColumnNewWithAttribute(title, renderer, attribute, column)
	return w
}

// Window creates a *gtk.Window.
func Window(t gtk.WindowType) *gtk.Window {
	w, _ := gtk.WindowNew(t)
	return w
}
