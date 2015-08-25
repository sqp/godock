package cfwidget

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"github.com/sqp/godock/libs/cdtype"    // Logger type.
	"github.com/sqp/godock/libs/text/tran" // Translate.

	"github.com/sqp/godock/widgets/cfbuild/cftype"   // Types for config file builder usage.
	"github.com/sqp/godock/widgets/cfbuild/datatype" // Types for config file builder data source.
	"github.com/sqp/godock/widgets/common"           // PixbufNewFromFile.
	"github.com/sqp/godock/widgets/gtk/gunvalue"     // Extract gvalue data.
	"github.com/sqp/godock/widgets/gtk/indexiter"    // References of iter to reselect.
	"github.com/sqp/godock/widgets/gtk/newgtk"       // Create widgets.
	"github.com/sqp/godock/widgets/handbook"         // Package preview.

	"errors"
	"os"
	"path/filepath"
	"strconv"
)

var (
	iconSizeCombo = 24
	iconSizeFrame = 20
)

// iconSize = int(gtk.ICON_SIZE_LARGE_TOOLBAR)

//
//----------------------------------------------------------[ GET/SET ACTIVE ]--

type widgetActiver interface {
	GetActive() bool
	SetActive(bool)
}

func listActiverGet(btns []widgetActiver) (vals []bool) {
	for _, btn := range btns {
		vals = append(vals, btn.GetActive())
	}
	return
}

func listActiverSet(btns []widgetActiver, vals []bool) {
	for i, btn := range btns {
		btn.SetActive(vals[i])
	}
}

//
//-----------------------------------------------------[ GET/SET FLOAT VALUE ]--

// PackValuerAsInt packs a valuer widget with its reset button (to given value).
// Values are get and set as int.
//
func PackValuerAsInt(key *cftype.Key, w gtk.IWidget, valuer WidgetValuer, value int) {
	key.PackKeyWidget(key,
		func() interface{} { return int(valuer.GetValue()) },
		func(uncast interface{}) { valuer.SetValue(float64(uncast.(int))) },
		w)

	oldval, e := key.Storage().Default(key.Group, key.Name)
	if e == nil {
		PackReset(key, oldval.Int())
	}
}

// WidgetValuer defines a widget with GetValue and SetValue methods.
//
type WidgetValuer interface {
	GetValue() float64
	SetValue(float64)
}

func listValuerGet(btns []WidgetValuer) (vals []float64) {
	for _, btn := range btns {
		vals = append(vals, btn.GetValue())
	}
	return
}

func listValuerSet(btns []WidgetValuer, vals []float64) {
	for i, btn := range btns {
		btn.SetValue(vals[i])
	}
}

//
//-------------------------------------------------------------[ MODELS DATA ]--

// static void _fill_model_with_one_theme (const gchar *cThemeName, CairoDockPackage *pTheme, gpointer *data)
// {
// 	GtkListStore *pModele = data[0];
// 	gchar *cHint = data[1];
// 	if ( ! cHint  // no hint is specified => take all themes
// 		|| ! pTheme->cHint  // the theme has no hint => it's a generic theme, take it
// 		|| strcmp (cHint, pTheme->cHint) == 0 )  // hints match, take it
// 	{
// 		GtkTreeIter iter;
// 		memset (&iter, 0, sizeof (GtkTreeIter));
// 		gtk_list_store_append (GTK_LIST_STORE (pModele), &iter);
// 		gchar *cReadmePath = g_strdup_printf ("%s/readme", pTheme->cPackagePath);
// 		gchar *cPreviewPath = g_strdup_printf ("%s/preview", pTheme->cPackagePath);
// 		gchar *cResult = g_strdup_printf ("%s[%d]", cThemeName, pTheme->iType);

// 		GdkPixbuf *pixbuf = _cairo_dock_gui_get_package_state_icon (pTheme->iType);
// 		gtk_list_store_set (GTK_LIST_STORE (pModele), &iter,
// 			CAIRO_DOCK_MODEL_NAME, pTheme->cDisplayedName,
// 			CAIRO_DOCK_MODEL_RESULT, cResult,
// 			CAIRO_DOCK_MODEL_DESCRIPTION_FILE, cReadmePath,
// 			CAIRO_DOCK_MODEL_IMAGE, cPreviewPath,
// 			CAIRO_DOCK_MODEL_ORDER, pTheme->iRating,
// 			CAIRO_DOCK_MODEL_ORDER2, pTheme->iSobriety,
// 			CAIRO_DOCK_MODEL_STATE, pTheme->iType,
// 			CAIRO_DOCK_MODEL_SIZE, pTheme->fSize,
// 			CAIRO_DOCK_MODEL_ICON, pixbuf,
// 			CAIRO_DOCK_MODEL_AUTHOR, pTheme->cAuthor, -1);
// 		g_free (cReadmePath);
// 		g_free (cPreviewPath);
// 		g_free (cResult);
// 		g_object_unref (pixbuf);
// 	}
// }

// func (widget *ListAdd) Load() {
// 	for _, pack := range widget.control.ListApplets() {
// 		if len(pack.Instances) > 0 && !pack.IsMultiInstance {
// 			// log.DEV("dropped", pack.Title)
// 		} else {
// 			iter := widget.newIter(pack)
// 			widget.model.SetCols(iter, gtk.Cols{
// 				RowData:     pack,
// 				RowName:     pack.Title,
// 				RowCategory: pack.FormatCategory(),
// 			})

// 			if pack.Icon[0] != '/' {
// 				log.DEV("no icon", pack.Title, pack.Icon)
// 			} else if pix, e := gdk.PixbufNewFromFileAtScale(pack.Icon, iconSize, iconSize, true); !log.Err(e, "Load icon") {
// 				widget.model.SetValue(iter, RowIcon, pix)
// 			}
// 		}
// 	}
// }

// Rows defines liststore rows. Must match the ListStore declaration type and order.
const (
	RowKey = iota
	RowName
	RowIcon
	RowDesc
)

func newModelSimple() *gtk.ListStore {
	return newgtk.ListStore(
		glib.TYPE_STRING,    // RowKey
		glib.TYPE_STRING,    // RowName
		gdk.PixbufGetType(), // RowIcon
		glib.TYPE_STRING,    // RowDesc
	)
}

// fill the model and can save references of fields+iter.
//
func fillModelWithFields(key *cftype.Key, model *gtk.ListStore, list []datatype.Field, current string, ro indexiter.ByString) (toSelect *gtk.TreeIter) {
	for _, field := range list {
		iter := modelAddField(key, model, field, ro)

		if field.Key == current {
			toSelect = iter
		}
	}
	return
}

// modelAddField adds one field to the model and can save reference of fields+iter.
//
func modelAddField(key *cftype.Key, model *gtk.ListStore, field datatype.Field, ro indexiter.ByString) *gtk.TreeIter {
	iter := model.Append()
	model.SetCols(iter, gtk.Cols{
		RowKey:  field.Key,
		RowName: field.Name,
		RowDesc: "none",
	})
	if field.Icon != "" {
		pix, e := common.PixbufNewFromFile(field.Icon, iconSizeCombo)
		if !key.Log().Err(e, "Load icon") {
			model.SetValue(iter, RowIcon, pix)
		}
	}

	if ro != nil {
		ro.Append(iter, field.Key)
	}
	return iter
}

//
//-------------------------------------------------------[ COMMON LISTSTORES ]--

// NewComboBox creates a combo box.
//
func NewComboBox(key *cftype.Key, withEntry, numbered bool, current string, list []datatype.Field) (
	widget *gtk.ComboBox, model *gtk.ListStore, getValue func() interface{}, setValue func(interface{})) {

	model = newModelSimple()
	// gtk_tree_sortable_set_sort_column_id(GTK_TREE_SORTABLE(modele), CAIRO_DOCK_MODEL_NAME, GTK_SORT_ASCENDING)

	widget = newgtk.ComboBoxWithModel(model)
	renderer := newgtk.CellRendererText()
	widget.PackStart(renderer, true)
	widget.AddAttribute(renderer, "text", RowName)

	// Fill and set current.
	iter := fillModelWithFields(key, model, list, current, nil)
	widget.SetActiveIter(iter)

	switch {
	case withEntry: // get and set the entry content string.
		entry := newgtk.Entry() // Add entry manually so we don't have to recast a GetChild
		entry.SetText(current)
		widget.Add(entry)
		widget.Set("id-column", RowName)
		widget.Connect("changed", func() { entry.SetText(widget.GetActiveID()) })
		getValue = func() interface{} { v, _ := entry.GetText(); return v }
		setValue = func(uncast interface{}) { entry.SetText(uncast.(string)) }

	case numbered: // get and set selected as position int
		getValue = func() interface{} { return widget.GetActive() }
		setValue = func(uncast interface{}) { widget.SetActive(uncast.(int)) }

	default: // get and set selected as content string
		widget.Set("id-column", RowKey)
		getValue = func() interface{} { return widget.GetActiveID() }
		setValue = func(uncast interface{}) {
			newID := datatype.ListFieldsIDByName(list, uncast.(string), key.Log())
			widget.SetActive(newID)
		}
	}

	return
}

// NewComboBoxWithModel adds a combo box with the given model (can be nil).
//
// _add_combo_from_modele
// used do/while. find why
func NewComboBoxWithModel(model *gtk.ListStore, log cdtype.Logger, bAddPreviewWidgets, bWithEntry, bHorizontalPackaging bool) (
	widget *gtk.ComboBox, getValue func() interface{}) {

	if model == nil {
		// TODO: need the one with entry.
		combo := newgtk.ComboBox()
		getValue = func() interface{} { v := combo.GetActive(); return v }

		widget = combo
		return
	}
	if bWithEntry {
		widget := newgtk.ComboBoxWithEntry()
		widget.SetModel(model)
	} else {

		combo := newgtk.ComboBoxWithModel(model)
		renderer := newgtk.CellRendererText()
		combo.PackStart(renderer, false)
		combo.AddAttribute(renderer, "text", RowName)

		getValue = getValueListCombo(combo, model, log)

		widget = combo
	}
	if bAddPreviewWidgets {
		// pPreviewBox = cairo_dock_gui_make_preview_box(pMainWindow, pOneWidget, bHorizontalPackaging, 1, NULL, NULL, pDataGarbage)
		// fullSize := bWithEntry || bHorizontalPackaging
		// gtk_box_pack_start (GTK_BOX (pAdditionalItemsVBox ? pAdditionalItemsVBox : pKeyBox), pPreviewBox, fullSize, fullSize, 0);
	}
	// cValue = g_key_file_get_string(pKeyFile, cGroupName, cKeyName, NULL)
	// if _cairo_dock_find_iter_from_name(model, cValue, &iter) {
	// 	gtk_combo_box_set_active_iter(GTK_COMBO_BOX(pOneWidget), &iter)
	// }
	return
}

func getValueListCombo(widget *gtk.ComboBox, model *gtk.ListStore, log cdtype.Logger) func() interface{} {
	return func() interface{} {
		iter, _ := widget.GetActiveIter()
		text, e := getActiveRowInCombo(model, iter)
		log.Err(e, "ListIcons")
		return text
	}
}

// getActiveRowInCombo gets the value of the current RowKey in the store.
//
func getActiveRowInCombo(model *gtk.ListStore, iter *gtk.TreeIter) (string, error) {
	if iter == nil {
		return "", errors.New("getActiveRowInCombo: no selection")
	}
	str, e := gunvalue.New(model.GetValue(iter, RowKey)).String()
	if e != nil {
		return "", e
	}

	// if (cValue == NULL && GTK_IS_COMBO_BOX (pOneWidget) && gtk_combo_box_get_has_entry (GTK_COMBO_BOX (pOneWidget)))
	// {
	// 	GtkWidget *pEntry = gtk_bin_get_child (GTK_BIN (pOneWidget));
	// 	cValue = g_strdup (gtk_entry_get_text (GTK_ENTRY (pEntry)));
	// }
	return str, nil
}

//
//----------------------------------------------------------[ COMMON WIDGETS ]--

// PackComboBoxWithListField creates a combo box filled with the getList call.
//
func PackComboBoxWithListField(key *cftype.Key, withEntry, numbered bool, getList func() []datatype.Field) *gtk.ComboBox {
	var list []datatype.Field
	if getList != nil {
		list = getList()
	}
	current, _ := key.Storage().String(key.Group, key.Name)
	widget, _, getValue, setValue := NewComboBox(key, withEntry, numbered, current, list)

	key.PackKeyWidget(key, getValue, setValue, widget)
	return widget
}

// PackComboBoxWithIndexHandbooker creates a combo box filled with the getList call.
//
func PackComboBoxWithIndexHandbooker(key *cftype.Key, index map[string]datatype.Handbooker) {
	model := newModelSimple()
	model.SetSortColumnId(RowName, gtk.SORT_ASCENDING)
	widget, getValue := NewComboBoxWithModel(model, key.Log(), false, false, false)

	value := key.Value().String()

	details := handbook.New(key.Log())
	key.PackWidget(details, false, false, 0)

	widget.Connect("changed", func() {
		name := key.Value().String()
		pack, ok := index[name]
		if ok {
			details.SetPackage(pack)
		} else {
			key.Log().NewErr("key missing", "ComboHandbook select preview:", name)
		}
	})

	fields := datatype.IndexHandbooksToFields(index)
	saved := indexiter.NewByString(widget, key.Log())
	iter := fillModelWithFields(key, model, fields, value, saved)
	widget.SetActiveIter(iter) // Set iter after connect, to update with current value.

	key.PackKeyWidget(key,
		getValue,
		func(uncast interface{}) { saved.SetActive(uncast.(string)) },
		widget,
	)
}

// fieldsPrepend prepends one or more fields to a list of fields.
//
func fieldsPrepend(list []datatype.Field, fields ...datatype.Field) func() []datatype.Field {
	return func() []datatype.Field {
		return append(fields, list...) // prepend defaults.
	}
}

//
//----------------------------------------------------------[ COMMON PACKING ]--

// PackReset adds a reset value button.
//
func PackReset(key *cftype.Key, value interface{}) *gtk.Button {
	fileDefault := key.Storage().FileDefault()
	if fileDefault == "" {
		return nil
	}

	back := newgtk.ButtonFromIconName("edit-clear", gtk.ICON_SIZE_MENU)
	back.Connect("clicked", func() { key.ValueSet(value) })
	key.PackSubWidget(back)
	return back
}

// WrapKeyScale wraps a key scale with its information labels if needed (enough values).
//
// (was _pack_hscale).
func WrapKeyScale(key *cftype.Key, child *gtk.Scale) gtk.IWidget {
	child.Set("width-request", 150)
	if len(key.AuthorizedValues) >= 4 {

		child.Set("value-pos", gtk.POS_TOP)
		// log.DEV("MISSING SubScale options", string(key.Type), key.AuthorizedValues)
		box := newgtk.Box(gtk.ORIENTATION_HORIZONTAL, 0)
		// 	GtkWidget * pAlign = gtk_alignment_new(1., 1., 0., 0.)
		labelLeft := newgtk.Label(key.Translate(key.AuthorizedValues[2]))
		// 	pAlign = gtk_alignment_new(1., 1., 0., 0.)
		labelRight := newgtk.Label(key.Translate(key.AuthorizedValues[3]))

		box.PackStart(labelLeft, false, false, 0)
		box.PackStart(child, false, false, 0)
		box.PackStart(labelRight, false, false, 0)
		return box
	}
	child.Set("value-pos", gtk.POS_LEFT)
	return child
}

//
//-----------------------------------------------------------------[ HELPERS ]--

func minMaxValues(key *cftype.Key) (float64, float64) {
	var fMinValue float64
	var fMaxValue float64 = 9999
	if len(key.AuthorizedValues) > 0 {
		fMinValue, _ = strconv.ParseFloat(key.AuthorizedValues[0], 32)
	}
	if len(key.AuthorizedValues) > 1 {
		fMaxValue, _ = strconv.ParseFloat(key.AuthorizedValues[1], 32)
	}
	return fMinValue, fMaxValue
}

//
//---------------------------------------------------------------[ CALLBACKS ]--

type valuePair struct {
	linked *gtk.SpinButton
	toggle *gtk.ToggleButton
}

func onValuePairChanged(updated *gtk.SpinButton, vp *valuePair) {
	if !vp.toggle.GetActive() {
		return
	}
	newval := updated.GetValue()
	if vp.linked.GetValue() != newval {
		vp.linked.SetValue(newval)
	}
}

type textDefaultData struct {
	key  *cftype.Key
	text string // default text to use.
	cbID glib.SignalHandle
}

// text changed by the user. Restore color and the ability to save the value.
func onTextDefaultChanged(entry *gtk.Entry, key *cftype.Key) {
	key.IsDefault = false
	entry.OverrideColor(gtk.STATE_FLAG_NORMAL, nil)
}

// got focus, removing default text if any.
func onTextDefaultFocusIn(widget *gtk.Entry, _ *gdk.Event, data textDefaultData) {
	if data.key.IsDefault {
		widget.HandlerBlock(data.cbID)
		widget.SetText("")
		widget.HandlerUnblock(data.cbID)
	}
}

// lost focus, setting back default text and color if needed.
func onTextDefaultFocusOut(widget *gtk.Entry, _ *gdk.Event, data textDefaultData) {
	text, _ := widget.GetText()
	data.key.IsDefault = text == ""
	if data.key.IsDefault {
		widget.HandlerBlock(data.cbID)
		widget.SetText(data.text)
		widget.HandlerUnblock(data.cbID)

		color := gdk.NewRGBA(cftype.DefaultTextColor, cftype.DefaultTextColor, cftype.DefaultTextColor, 1)
		widget.OverrideColor(gtk.STATE_FLAG_NORMAL, color)
	}
}

type fileChooserData struct {
	entry *gtk.Entry
	key   *cftype.Key
}

func onFileChooserOpen(obj *gtk.Button, data fileChooserData) {
	var parent *gtk.Window

	var title string
	var action gtk.FileChooserAction

	switch data.key.Type {
	case cftype.KeyFolderSelector:
		title = tran.Slate("Pick up a directory")
		action = gtk.FILE_CHOOSER_ACTION_SELECT_FOLDER

	case cftype.KeyImageSelector:
		title = tran.Slate("Pick up an image")
		action = gtk.FILE_CHOOSER_ACTION_OPEN

	default:
		title = tran.Slate("Pick up a file")
		action = gtk.FILE_CHOOSER_ACTION_OPEN
	}

	dialog, _ := gtk.FileChooserDialogNewWith2Buttons(title, parent, action,
		"_OK", gtk.RESPONSE_OK, "_Cancel", gtk.RESPONSE_CANCEL)

	// Set the current folder to the current value in conf.
	value, _ := data.entry.GetText()
	if value == "" || value[0] != '/' {
		if data.key.IsType(cftype.KeyImageSelector) {
			println("need dir pictures")
			// dialog.SetCurrentFolder(filepath.Dir(value)) // g_get_user_special_dir (G_USER_DIRECTORY_PICTURES) :
		} else {
			println(os.Getenv("HOME"))
			dialog.SetCurrentFolder(os.Getenv("HOME"))
		}
	} else {
		dialog.SetCurrentFolder(filepath.Dir(value))
	}

	if data.key.IsType(cftype.KeyImageSelector) { // Add shortcuts to icons of the system.
		dialog.AddShortcutFolder("/usr/share/icons")
		dialog.AddShortcutFolder("/usr/share/pixmaps")
	}

	if data.key.IsType(cftype.KeyFileSelector, cftype.KeySoundSelector) { // Add shortcuts to system icons directories.
		filter := newgtk.FileFilter()
		filter.SetName(tran.Slate("All"))
		filter.AddPattern("*")
		dialog.AddFilter(filter)
	}

	if data.key.IsType(cftype.KeyFileSelector, cftype.KeyImageSelector) { // Preview and images filter.
		filter := newgtk.FileFilter()
		filter.SetName(tran.Slate("Image"))
		filter.AddPixbufFormats()
		dialog.AddFilter(filter)

		img := newgtk.Image()
		dialog.SetPreviewWidget(img)
		dialog.Connect("update-preview", onFileChooserUpdatePreview, img)
	}

	dialog.Show()
	answer := dialog.Run()
	if gtk.ResponseType(answer) == gtk.RESPONSE_OK {
		data.entry.SetText(dialog.GetFilename())
	}
	dialog.Destroy()
}

func onFileChooserUpdatePreview(dialog *gtk.FileChooserDialog, img *gtk.Image) {
	filename := dialog.GetPreviewFilename()
	pixbuf, _ := gdk.PixbufNewFromFileAtSize(filename, 64, 64)
	if pixbuf == nil {
		dialog.SetPreviewWidgetActive(false)
	} else {
		dialog.SetPreviewWidgetActive(true)
		img.SetFromPixbuf(pixbuf)
		// 		g_object_unref (pixbuf); // need unref ??
	}
}

type treeViewData struct {
	log    cdtype.Logger
	model  *gtk.ListStore
	widget *gtk.TreeView
	entry  *gtk.Entry
}

func onTreeviewMoveUp(_ *gtk.Button, data treeViewData) {
	var treeModel gtk.ITreeModel = data.model
	var iter gtk.TreeIter
	sel, e := data.widget.GetSelection()
	if data.log.Err(e, "WidgetTreeView widget.GetSelection") || !sel.GetSelected(&treeModel, &iter) {
		return
	}

	order, e := gunvalue.New(data.model.GetValue(&iter, 5)).Int()
	if data.log.Err(e, "WidgetTreeView model.GetValue order") {
		return
	}

	data.model.SetValue(&iter, 5, order-1) // Set the new order value.

	if !data.model.IterPrevious(&iter) { // Get previous iter.
		return
	}
	data.model.SetValue(&iter, 5, order) // Set it to its new order.
}

// Move treeview selection down.
//
func onTreeviewMoveDown(_ *gtk.Button, data treeViewData) {
	var treeModel gtk.ITreeModel = data.model
	var iter gtk.TreeIter
	sel, e := data.widget.GetSelection()
	if data.log.Err(e, "WidgetTreeView widget.GetSelection") || !sel.GetSelected(&treeModel, &iter) {
		return
	}
	order, e := gunvalue.New(data.model.GetValue(&iter, 5)).Int()
	if data.log.Err(e, "WidgetTreeView model.GetValue order") {
		return
	}

	if order >= data.model.IterNChildren(nil) { // Check we aren't in last position.
		return
	}

	data.model.SetValue(&iter, 5, order+1) // Set the new order value.

	if !data.model.IterNext(&iter) { // Get next iter.
		return
	}
	data.model.SetValue(&iter, 5, order) // Set it to its new order.
}

func onTreeviewAddText(_ *gtk.Button, data treeViewData) {

	// Add new iter to model with the value of the entry widget. Clear entry widget.
	val, e := data.entry.GetText()
	if val != "" && !data.log.Err(e, "WidgetTreeView entry.GetText") {
		data.entry.SetText("")

		// Add new iter to model.
		iter := data.model.Append()
		data.model.SetValue(iter, RowKey, val)
		data.model.SetValue(iter, RowName, val)
		data.model.SetValue(iter, 4, true)                            // active
		data.model.SetValue(iter, 5, data.model.IterNChildren(nil)-1) // order

		// Select new iter.
		sel, e := data.widget.GetSelection()
		if !data.log.Err(e, "WidgetTreeView widget.GetSelection") {
			sel.SelectIter(iter)
		}
	}
}

// Remove selected iter from model. Set its value to the entry widget.
//
func onTreeviewRemoveText(_ *gtk.Button, data treeViewData) {

	var treeModel gtk.ITreeModel = data.model
	var iter gtk.TreeIter
	sel, e := data.widget.GetSelection()
	if !data.log.Err(e, "WidgetTreeView widget.GetSelection") {
		if !sel.GetSelected(&treeModel, &iter) {
			return
		}
	}

	name, e := gunvalue.New(data.model.GetValue(&iter, RowName)).String()
	if !data.log.Err(e, "WidgetTreeView model.GetValue RowName") {
		data.entry.SetText(name)
	}

	order, e := gunvalue.New(data.model.GetValue(&iter, 5)).Int()
	data.model.Remove(&iter)
	if data.log.Err(e, "WidgetTreeView model.GetValue order") { // no order nor iters. can't do shit.
		return
	}

	// Decrease order for iters after the deleted one.
	it, ok := data.model.GetIterFirst()
	for ok {
		current, _ := gunvalue.New(data.model.GetValue(it, 5)).Int()
		if current > order {
			data.model.SetValue(it, 5, current-1)
		}
		ok = data.model.IterNext(it)
	}

}

type textGrabData struct {
	entry *gtk.Entry
	win   *gtk.Window
	cbID  glib.SignalHandle
}

// Start listening for keyboard events until a valid shortcut is found.
//
func onKeyGrabClicked(_ *gtk.Button, data *textGrabData) {
	if data.cbID > 0 { // Already waiting.
		return
	}
	data.entry.SetSensitive(false)
	data.cbID, _ = data.win.Connect("key-press-event", onKeyGrabReceived, data)
}

// Receives keyboard events until a valid shortcut is found, to set it as entry text.
//
func onKeyGrabReceived(_ *gtk.Window, event *gdk.Event, data *textGrabData) {
	key := &gdk.EventKey{Event: event}

	if !gtk.AcceleratorValid(key.KeyVal(), gdk.ModifierType(key.State())) {
		return
	}
	data.entry.SetSensitive(true)
	data.win.HandlerDisconnect(data.cbID)
	data.cbID = 0

	// This lets us ignore all ignorable modifier keys, including NumLock and many others. :)
	// The logic is: keep only the important modifiers that were pressed for this event.
	state := gdk.ModifierType(key.State()) & gtk.AcceleratorGetDefaultModMask()

	accel := gtk.AcceleratorName(key.KeyVal(), state)
	data.entry.SetText(accel)
}

func onClassGrabClicked(obj *gtk.Button) {
	println("grab class is still to do")

	// 	GtkEntry *pEntry = data[0];
	// 	GtkWindow *pParentWindow = data[1];

	// 	//set widget insensitive
	// 	gtk_widget_set_sensitive (GTK_WIDGET(pEntry), FALSE);

	// 	g_signal_connect (G_OBJECT(pParentWindow), "key-press-event", G_CALLBACK(_cairo_dock_key_grab_cb), pEntry);
}

// static void _cairo_dock_key_grab_class (G_GNUC_UNUSED GtkButton *button, gpointer *data)
// {
// 	GtkEntry *pEntry = data[0];
// 	// GtkWindow *pParentWindow = data[1];

// 	cd_debug ("clicked");
// 	gtk_widget_set_sensitive (GTK_WIDGET(pEntry), FALSE);  // lock the widget during the grab (it makes it more comprehensive).

// 	const gchar *cResult = NULL;
// 	GldiWindowActor *actor = gldi_window_pick ();

// 	if (actor && actor->bIsTransientFor)
// 		actor = gldi_window_get_transient_for (actor);

// 	if (actor)
// 		cResult = actor->cClass;
// 	else
// 		cd_warning ("couldn't get a window actor");

// 	gtk_widget_set_sensitive (GTK_WIDGET(pEntry), TRUE);  // unlock the widget
// 	gtk_entry_set_text (pEntry, cResult);  // write the result in the entry-box
// }

// //Sound Callback
// static void _cairo_dock_play_a_sound (G_GNUC_UNUSED GtkButton *button, gpointer *data)
// {
// 	GtkWidget *pEntry = data[0];
// 	const gchar *cSoundPath = gtk_entry_get_text (GTK_ENTRY (pEntry));
// 	cairo_dock_play_sound (cSoundPath);
// }
