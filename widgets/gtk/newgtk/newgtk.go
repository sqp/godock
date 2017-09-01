// Package newgtk creates gtk objects.
package newgtk

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

// SetOnError sets the error forwarder for widget creation.
func SetOnError(onErr func(error)) {
	onError = onErr
}

var onError = func(error) {}

//
//--------------------------------------------------------------[ GTK RECAST ]--

// Adjustment creates a *gtk.Adjustment.
func Adjustment(value, lower, upper, stepIncrement, pageIncrement, pageSize float64) *gtk.Adjustment {
	w, e := gtk.AdjustmentNew(value, lower, upper, stepIncrement, pageIncrement, pageSize)
	onError(e)
	return w
}

// Box creates a *gtk.Box.
func Box(orientation gtk.Orientation, spacing int) *gtk.Box {
	w, e := gtk.BoxNew(orientation, spacing)
	onError(e)
	return w
}

// Button creates a *gtk.Button.
func Button() *gtk.Button {
	w, e := gtk.ButtonNew()
	onError(e)
	return w
}

// ButtonFromIconName creates a *gtk.Button.
func ButtonFromIconName(label string, size gtk.IconSize) *gtk.Button {
	w, e := gtk.ButtonNewFromIconName(label, size)
	onError(e)
	return w
}

// ButtonWithLabel creates a *gtk.Button.
func ButtonWithLabel(label string) *gtk.Button {
	w, e := gtk.ButtonNewWithLabel(label)
	onError(e)
	return w
}

// ButtonWithMnemonic creates a *gtk.Button.
func ButtonWithMnemonic(label string) *gtk.Button {
	w, e := gtk.ButtonNewWithMnemonic(label)
	onError(e)
	return w
}

// CellRendererPixbuf creates a *gtk.CellRendererPixbuf.
func CellRendererPixbuf() *gtk.CellRendererPixbuf {
	w, e := gtk.CellRendererPixbufNew()
	onError(e)
	return w
}

// CellRendererText creates a *gtk.CellRendererText.
func CellRendererText() *gtk.CellRendererText {
	w, e := gtk.CellRendererTextNew()
	onError(e)
	return w
}

// CellRendererToggle creates a *gtk.CellRendererToggle.
func CellRendererToggle() *gtk.CellRendererToggle {
	w, e := gtk.CellRendererToggleNew()
	onError(e)
	return w
}

// CheckButton creates a *gtk.CheckButton.
func CheckButton() *gtk.CheckButton {
	w, e := gtk.CheckButtonNew()
	onError(e)
	return w
}

// CheckMenuItemWithLabel creates a *gtk.CheckMenuItem.
func CheckMenuItemWithLabel(label string) *gtk.CheckMenuItem {
	w, e := gtk.CheckMenuItemNewWithLabel(label)
	onError(e)
	return w
}

// ColorButtonWithRGBA creates a *gtk.ColorButton.
func ColorButtonWithRGBA(gdkColor *gdk.RGBA) *gtk.ColorButton {
	w, e := gtk.ColorButtonNewWithRGBA(gdkColor)
	onError(e)
	return w
}

// ComboBox creates a *gtk.ComboBox.
func ComboBox() *gtk.ComboBox {
	w, e := gtk.ComboBoxNew()
	onError(e)
	return w
}

// ComboBoxWithEntry creates a *gtk.ComboBox.
func ComboBoxWithEntry() *gtk.ComboBox {
	w, e := gtk.ComboBoxNewWithEntry()
	onError(e)
	return w
}

// ComboBoxWithModel creates a *gtk.ComboBox.
func ComboBoxWithModel(model gtk.ITreeModel) *gtk.ComboBox {
	w, e := gtk.ComboBoxNewWithModel(model)
	onError(e)
	return w
}

// ComboBoxText creates a *gtk.ComboBoxText.
func ComboBoxText() *gtk.ComboBoxText {
	w, e := gtk.ComboBoxTextNew()
	onError(e)
	return w
}

// Dialog creates a *gtk.Dialog.
func Dialog() *gtk.Dialog {
	w, e := gtk.DialogNew()
	onError(e)
	return w
}

// Entry creates a *gtk.Entry.
func Entry() *gtk.Entry {
	w, e := gtk.EntryNew()
	onError(e)
	return w
}

// Expander creates a *gtk.Expander.
func Expander(label string) *gtk.Expander {
	w, e := gtk.ExpanderNew(label)
	onError(e)
	return w
}

// FileChooserDialogWith2Buttons creates a *gtk.FileChooserDialog.
func FileChooserDialogWith2Buttons(title string, parent *gtk.Window, action gtk.FileChooserAction,
	firstText string, firstID gtk.ResponseType,
	secondText string, secondID gtk.ResponseType) *gtk.FileChooserDialog {

	w, e := gtk.FileChooserDialogNewWith2Buttons(title, parent, action, firstText, firstID, secondText, secondID)
	onError(e)
	return w
}

// FileFilter creates a *gtk.FileFilter.
func FileFilter() *gtk.FileFilter {
	w, e := gtk.FileFilterNew()
	onError(e)
	return w
}

// FontButtonWithFont creates a *gtk.FontButton.
func FontButtonWithFont(fontname string) *gtk.FontButton {
	w, e := gtk.FontButtonNewWithFont(fontname)
	onError(e)
	return w
}

// Frame creates a *gtk.Frame.
func Frame(label string) *gtk.Frame {
	w, e := gtk.FrameNew(label)
	onError(e)
	return w
}

// Grid creates a *gtk.Grid.
func Grid() *gtk.Grid {
	w, e := gtk.GridNew()
	onError(e)
	return w
}

// Image creates a *gtk.Image.
func Image() *gtk.Image {
	w, e := gtk.ImageNew()
	onError(e)
	return w
}

// ImageFromFile creates a *gtk.Image.
func ImageFromFile(file string) *gtk.Image {
	w, e := gtk.ImageNewFromFile(file)
	onError(e)
	return w
}

// ImageFromIconName creates a *gtk.Image.
func ImageFromIconName(iconName string, size gtk.IconSize) *gtk.Image {
	w, e := gtk.ImageNewFromIconName(iconName, size)
	onError(e)
	return w
}

// Label creates a *gtk.Label.
func Label(label string) *gtk.Label {
	w, e := gtk.LabelNew(label)
	onError(e)
	return w
}

// LinkButtonWithLabel creates a *gtk.LinkButton.
func LinkButtonWithLabel(uri, label string) *gtk.LinkButton {
	w, e := gtk.LinkButtonNewWithLabel(uri, label)
	onError(e)
	return w
}

// ListBox creates a *gtk.ListBox.
func ListBox() *gtk.ListBox {
	w, e := gtk.ListBoxNew()
	onError(e)
	return w
}

// ListBoxRow creates a *gtk.ListBoxRow.
func ListBoxRow() *gtk.ListBoxRow {
	w, e := gtk.ListBoxRowNew()
	onError(e)
	return w
}

// ListStore creates a *gtk.ListStore.
func ListStore(types ...glib.Type) *gtk.ListStore {
	w, e := gtk.ListStoreNew(types...)
	onError(e)
	return w
}

// Menu creates a *gtk.Menu.
func Menu() *gtk.Menu {
	w, e := gtk.MenuNew()
	onError(e)
	return w
}

// MenuItem creates a *gtk.MenuItem.
func MenuItem() *gtk.MenuItem {
	w, e := gtk.MenuItemNew()
	onError(e)
	return w
}

// MenuItemWithLabel creates a *gtk.MenuItem.
func MenuItemWithLabel(label string) *gtk.MenuItem {
	w, e := gtk.MenuItemNewWithLabel(label)
	onError(e)
	return w
}

// Notebook creates a *gtk.Notebook.
func Notebook() *gtk.Notebook {
	w, e := gtk.NotebookNew()
	onError(e)
	return w
}

// Paned creates a *gtk.Paned.
func Paned(orientation gtk.Orientation) *gtk.Paned {
	w, e := gtk.PanedNew(orientation)
	onError(e)
	return w
}

// RadioMenuItemWithLabel creates a *gtk.RadioMenuItem.
func RadioMenuItemWithLabel(group *glib.SList, label string) *gtk.RadioMenuItem {
	w, e := gtk.RadioMenuItemNewWithLabel(group, label)
	onError(e)
	return w
}

// Scale creates a *gtk.Scale.
func Scale(orientation gtk.Orientation, adjustment *gtk.Adjustment) *gtk.Scale {
	w, e := gtk.ScaleNew(orientation, adjustment)
	onError(e)
	return w
}

// ScaleWithRange creates a *gtk.Scale.
func ScaleWithRange(orientation gtk.Orientation, min, max, step float64) *gtk.Scale {
	w, e := gtk.ScaleNewWithRange(orientation, min, max, step)
	onError(e)
	return w
}

// ScrolledWindow creates a *gtk.ScrolledWindow.
func ScrolledWindow(hadjustment, vadjustment *gtk.Adjustment) *gtk.ScrolledWindow {
	w, e := gtk.ScrolledWindowNew(hadjustment, vadjustment)
	onError(e)
	return w
}

// Separator creates a *gtk.Separator.
func Separator(orientation gtk.Orientation) *gtk.Separator {
	w, e := gtk.SeparatorNew(orientation)
	onError(e)
	return w
}

// SeparatorMenuItem creates a *gtk.SeparatorMenuItem.
func SeparatorMenuItem() *gtk.SeparatorMenuItem {
	w, e := gtk.SeparatorMenuItemNew()
	onError(e)
	return w
}

// SpinButtonWithRange creates a *gtk.SpinButton.
func SpinButtonWithRange(min, max, step float64) *gtk.SpinButton {
	w, e := gtk.SpinButtonNewWithRange(min, max, step)
	onError(e)
	return w
}

// Spinner creates a *gtk.Spinner.
func Spinner() *gtk.Spinner {
	w, e := gtk.SpinnerNew()
	onError(e)
	return w
}

// Stack creates a *gtk.Stack.
func Stack() *gtk.Stack {
	w, e := gtk.StackNew()
	onError(e)
	return w
}

// StackSwitcher creates a *gtk.StackSwitcher.
func StackSwitcher() *gtk.StackSwitcher {
	w, e := gtk.StackSwitcherNew()
	onError(e)
	return w
}

// Switch creates a *gtk.Switch.
func Switch() *gtk.Switch {
	w, e := gtk.SwitchNew()
	onError(e)
	return w
}

// TextView creates a *gtk.TextView.
func TextView() *gtk.TextView {
	w, e := gtk.TextViewNew()
	onError(e)
	return w
}

// ToggleButton creates a *gtk.ToggleButton.
func ToggleButton() *gtk.ToggleButton {
	w, e := gtk.ToggleButtonNew()
	onError(e)
	return w
}

// ToggleButtonWithLabel creates a *gtk.ToggleButton.
func ToggleButtonWithLabel(label string) *gtk.ToggleButton {
	w, e := gtk.ToggleButtonNewWithLabel(label)
	onError(e)
	return w
}

// TreeViewWithModel creates a *gtk.TreeView.
func TreeViewWithModel(model gtk.ITreeModel) *gtk.TreeView {
	w, e := gtk.TreeViewNewWithModel(model)
	onError(e)
	return w
}

// TreeViewColumnWithAttribute creates a *gtk.TreeViewColumn.
func TreeViewColumnWithAttribute(title string, renderer gtk.ICellRenderer, attribute string, column int) *gtk.TreeViewColumn {
	w, e := gtk.TreeViewColumnNewWithAttribute(title, renderer, attribute, column)
	onError(e)
	return w
}

// Window creates a *gtk.Window.
func Window(t gtk.WindowType) *gtk.Window {
	w, e := gtk.WindowNew(t)
	onError(e)
	return w
}
