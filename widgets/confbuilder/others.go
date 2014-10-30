package confbuilder

import (
	"github.com/conformal/gotk3/gdk"
	"github.com/conformal/gotk3/glib"
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/godock/libs/log"
	"github.com/sqp/godock/libs/packages"

	"github.com/sqp/godock/widgets/common"
	"github.com/sqp/godock/widgets/confbuilder/datatype"
	"github.com/sqp/godock/widgets/gtk/buildhelp"
	"github.com/sqp/godock/widgets/gtk/gunvalue"

	"fmt"
	"os"
	"path/filepath"
)

//
//----------------------------------------------------------------[ HANDBOOK ]--

// Handbook defines a handbook widget (applet info).
//
type Handbook struct {
	gtk.Frame    // Main widget is the container.
	title        *gtk.Label
	author       *gtk.Label
	description  *gtk.Label
	previewFrame *gtk.Frame
	previewImage *gtk.Image
}

// NewHandbook creates a handbook widget (applet info).
//
func NewHandbook() *Handbook {
	builder := buildhelp.New()

	builder.AddFromString(string(handbookXML()))
	// builder.AddFromFile("handbook.xml")

	widget := &Handbook{
		Frame:        *builder.GetFrame("handbook"),
		title:        builder.GetLabel("title"),
		author:       builder.GetLabel("author"),
		description:  builder.GetLabel("description"),
		previewFrame: builder.GetFrame("previewFrame"),
		previewImage: builder.GetImage("previewImage"),
	}

	if len(builder.Errors) > 0 {
		for _, e := range builder.Errors {
			log.DEV("build handbook", e)
		}
		return nil
	}

	return widget
}

// SetPackage fills the handbook data with a package.
//
func (widget *Handbook) SetPackage(book datatype.Handbooker) {
	title := common.Bold(common.Big(book.GetTitle() + " "))
	widget.title.SetMarkup(title + "v" + book.GetModuleVersion())
	widget.author.SetMarkup(common.Small(common.Mono(fmt.Sprintf("by %s", book.GetAuthor()))))
	widget.description.SetMarkup("<span rise='8000'>" + book.GetDescription() + "</span>")

	previewFound := false
	defer widget.previewFrame.SetVisible(previewFound)

	file := book.GetPreviewFilePath()
	if file == "" {
		return
	}
	_, w, h := gdk.PixbufGetFileInfo(file)

	var pixbuf *gdk.Pixbuf
	var e error
	if w > PreviewSizeMax || h > PreviewSizeMax {
		pixbuf, e = gdk.PixbufNewFromFileAtScale(file, PreviewSizeMax, PreviewSizeMax, true)
	} else {
		pixbuf, e = gdk.PixbufNewFromFile(file)
	}

	if !log.Err(e, "Handbook image: "+file) && pixbuf != nil {
		previewFound = true
		widget.previewImage.SetFromPixbuf(pixbuf)
	}
}

//
//-------------------------------------------------------------[ MODELS DATA ]--

func fillModelWithViews(model *gtk.ListStore, list []datatype.Field, current string) (toSelect int) {
	i := 0
	for _, field := range list {
		model.SetCols(model.Append(), gtk.Cols{
			RowKey:  field.Key,
			RowName: field.Name, // GETTEXT TRANSLATE
			// RowIcon: "none",
			RowDesc: "none"}) // (pRenderer != NULL ? pRenderer->cReadmeFilePath : "none")
		// 		CAIRO_DOCK_MODEL_IMAGE, (pRenderer != NULL ? pRenderer->cPreviewFilePath : "none")

		if field.Key == current {
			toSelect = i
		}
		i++
	}
	return
}

func fillModelWithTheme(model *gtk.ListStore, list []packages.Theme, current string) (toSelect *gtk.TreeIter) {
	for _, theme := range list {
		key := fmt.Sprintf("%s[%d]", theme.DirName, theme.Type)
		iter := model.Append()
		model.SetCols(iter, gtk.Cols{
			RowKey:  key,
			RowName: theme.Title,
			// RowIcon: "none",
			RowDesc: "none"})

		if key == current {
			toSelect = iter
		}
	}
	return
}

func fillModelWithThemeINI(model *gtk.ListStore, list packages.AppletPackages, current string) (toSelect *gtk.TreeIter) {
	for _, theme := range list {
		// key := fmt.Sprintf("%s[%d]", theme.DisplayedName, theme.Type)
		key := theme.DisplayedName
		iter := model.Append()
		model.SetCols(iter, gtk.Cols{
			RowKey:  key,
			RowName: theme.DisplayedName,
			// RowIcon: "none",
			RowDesc: "none"})

		log.DEV(key, theme.DisplayedName)

		if key == current {
			toSelect = iter
		}
	}
	return
}

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

func fillModelWithAnimation(model *gtk.ListStore, list []datatype.Field, current string) (toSelect *gtk.TreeIter) {
	for _, field := range list {
		iter := model.Append()
		model.SetCols(iter, gtk.Cols{
			RowKey:  field.Key,
			RowName: field.Name,
			// RowIcon: "none",
			RowDesc: "none"})

		if field.Key == current {
			toSelect = iter
		}
	}
	return
}

func fillModelWithDeskletDecoration(model *gtk.ListStore, list []datatype.Field, current string) (toSelect *gtk.TreeIter) {
	for _, field := range list {
		iter := model.Append()
		model.SetCols(iter, gtk.Cols{
			RowKey:  field.Key,
			RowName: field.Name, // GETTEXT TRANSLATE  pDecoration->cDisplayedName if available before name
			// RowIcon: "none",
			RowDesc: "none"})

		if field.Key == current {
			toSelect = iter
		}
	}
	return
}

func fillModelWithDocks(model *gtk.ListStore, list []datatype.Field, current string) (toSelect *gtk.TreeIter) {
	for _, field := range list {
		iter := model.Append()
		model.SetCols(iter, gtk.Cols{
			RowKey:  field.Key,
			RowName: field.Name,
			// RowIcon: "none",
			RowDesc: "none"})

		if field.Key == current {
			toSelect = iter
		}
	}
	model.SetCols(model.Append(), gtk.Cols{
		RowKey:  "_New Dock_", // special new dock key.
		RowName: "New main dock",
		// RowIcon: "none",
		RowDesc: "none",
	})
	return
}

//-------------------------------------------------------[ COMMON LISTSTORES ]--

// Rows defines liststore rows. Must match the ListStore declaration type and order.
const (
	RowKey = iota
	RowName
	RowIcon
	RowDesc
)

func newModelSimple() (*gtk.ListStore, error) {
	return gtk.ListStoreNew(
		glib.TYPE_STRING,    /* RowKey*/
		glib.TYPE_STRING,    /* RowName*/
		gdk.PixbufGetType(), /* RowIcon*/
		glib.TYPE_STRING)    /* RowDesc*/
}

//
func newComboBox(withEntry, numbered bool) (widget *gtk.ComboBox, model *gtk.ListStore, getValue func() interface{}) {
	model, _ = newModelSimple()
	// gtk_tree_sortable_set_sort_column_id(GTK_TREE_SORTABLE(modele), CAIRO_DOCK_MODEL_NAME, GTK_SORT_ASCENDING)

	if withEntry {
		widget, _ = gtk.ComboBoxNewWithEntry()
		widget.SetModel(model)
		widget.Set("entry-text-column", RowName)

		getValue = func() interface{} { // return selected as position int
			entry, _ := widget.GetChild()
			e := toEntry(entry)
			v, _ := e.GetText()
			return v
		}

	} else {
		widget, _ = gtk.ComboBoxNewWithModel(model)
		renderer, _ := gtk.CellRendererTextNew()
		widget.PackStart(renderer, true)
		widget.AddAttribute(renderer, "text", RowName)

		if numbered {
			getValue = func() interface{} { // return selected as position int
				return widget.GetActive()
			}
		} else {
			widget.Set("id-column", RowName)
			getValue = func() interface{} { // return selected as content string
				return widget.GetActiveID()
			}
		}
	}
	return
}

// _add_combo_from_modele
// used do/while. find why
func (build *Builder) newComboBoxWithModel(model *gtk.ListStore, bAddPreviewWidgets, bWithEntry, bHorizontalPackaging bool) (widget *gtk.ComboBox, getValue func() interface{}) {
	if model == nil {
		// TODO: need the one with entry.
		combo, _ := gtk.ComboBoxNew()
		getValue = func() interface{} { v := combo.GetActive(); return v }

		widget = combo
		return
	}
	if bWithEntry {
		widget, _ := gtk.ComboBoxNewWithEntry()
		widget.SetModel(model)
	} else {
		combo, _ := gtk.ComboBoxNewWithModel(model)
		renderer, _ := gtk.CellRendererTextNew()
		combo.PackStart(renderer, false)
		combo.AddAttribute(renderer, "text", RowName)
		getValue = func() interface{} {
			iter, _ := combo.GetActiveIter()
			text := GetActiveRowInCombo(model, iter)
			return text
		}

		widget = combo
	}
	if bAddPreviewWidgets {
		// pPreviewBox = cairo_dock_gui_make_preview_box(pMainWindow, pOneWidget, bHorizontalPackaging, 1, NULL, NULL, pDataGarbage)
		// bFullSize := bWithEntry || bHorizontalPackaging
		// gtk_box_pack_start (GTK_BOX (pAdditionalItemsVBox ? pAdditionalItemsVBox : pKeyBox), pPreviewBox, bFullSize, bFullSize, 0);
	}
	// cValue = g_key_file_get_string(pKeyFile, cGroupName, cKeyName, NULL)
	// if _cairo_dock_find_iter_from_name(model, cValue, &iter) {
	// 	gtk_combo_box_set_active_iter(GTK_COMBO_BOX(pOneWidget), &iter)
	// }
	return
}

// GetActiveRowInCombo gets the value of the current RowKey in the store.
//
func GetActiveRowInCombo(model *gtk.ListStore, iter *gtk.TreeIter) string {
	if iter != nil {
		str, e := gunvalue.New(model.GetValue(iter, RowKey)).String()
		if !log.Err(e, "GetActiveRowInCombo") {
			return str
		}
	}

	// if (cValue == NULL && GTK_IS_COMBO_BOX (pOneWidget) && gtk_combo_box_get_has_entry (GTK_COMBO_BOX (pOneWidget)))
	// {
	// 	GtkWidget *pEntry = gtk_bin_get_child (GTK_BIN (pOneWidget));
	// 	cValue = g_strdup (gtk_entry_get_text (GTK_ENTRY (pEntry)));
	// }
	return ""
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
	key  *Key
	text string // default text to use.
	cbID glib.SignalHandle
}

// text changed by the user. Restore color and the ability to save the value.
func onTextDefaultChanged(entry *gtk.Entry, key *Key) {
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
	if text == "" {
		data.key.IsDefault = true

		widget.HandlerBlock(data.cbID)
		widget.SetText(data.text)
		widget.HandlerUnblock(data.cbID)

		color := gdk.NewRGBA(DefaultTextColor, DefaultTextColor, DefaultTextColor, 1)
		widget.OverrideColor(gtk.STATE_FLAG_NORMAL, color)
	}
}

type fileChooserData struct {
	entry *gtk.Entry
	key   *Key
}

func onFileChooserOpen(obj *gtk.Button, data fileChooserData) {
	var parent *gtk.Window

	var title string
	var action gtk.FileChooserAction

	switch data.key.Type {
	case WidgetFolderSelector:
		title = "Pick up a directory" // GETTEXT TRANSLATION
		action = gtk.FILE_CHOOSER_ACTION_SELECT_FOLDER

	case WidgetImageSelector:
		title = "Pick up an image" // GETTEXT TRANSLATION
		action = gtk.FILE_CHOOSER_ACTION_OPEN

	default:
		title = "Pick up a file" // GETTEXT TRANSLATION
		action = gtk.FILE_CHOOSER_ACTION_OPEN
	}

	dialog, _ := gtk.FileChooserDialogNewWith2Buttons(title, parent, action,
		"_OK", gtk.RESPONSE_OK, "_Cancel", gtk.RESPONSE_CANCEL)

	// Set the current folder to the current value in conf.
	value, _ := data.entry.GetText()
	if value == "" || value[0] != '/' {
		if data.key.Type == WidgetImageSelector {
			println("need dir pictures")
			// dialog.SetCurrentFolder(filepath.Dir(value)) // g_get_user_special_dir (G_USER_DIRECTORY_PICTURES) :
		} else {
			println(os.Getenv("HOME"))
			dialog.SetCurrentFolder(os.Getenv("HOME"))
		}
	} else {
		dialog.SetCurrentFolder(filepath.Dir(value))
	}

	if data.key.Type == WidgetImageSelector { // Add shortcuts to icons of the system.
		dialog.AddShortcutFolder("/usr/share/icons")
		dialog.AddShortcutFolder("/usr/share/pixmaps")
	}

	if data.key.Type == WidgetFileSelector || data.key.Type == WidgetSoundSelector { // Add shortcuts to system icons directories.
		filter, _ := gtk.FileFilterNew()
		filter.SetName("All") // GETTEXT TRANSLATION
		filter.AddPattern("*")
		dialog.AddFilter(filter)
	}

	if data.key.Type == WidgetFileSelector || data.key.Type == WidgetImageSelector { // Preview and images filter.
		filter, _ := gtk.FileFilterNew()
		filter.SetName("Image") // GETTEXT TRANSLATION
		filter.AddPixbufFormats()
		dialog.AddFilter(filter)

		img, _ := gtk.ImageNew()
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
	model  *gtk.ListStore
	widget *gtk.TreeView
	entry  *gtk.Entry
}

func onTreeviewMoveUp(_ *gtk.Button, data treeViewData) {
	var treeModel gtk.ITreeModel = data.model
	var iter gtk.TreeIter
	sel, e := data.widget.GetSelection()
	if log.Err(e, "WidgetTreeView widget.GetSelection") || !sel.GetSelected(&treeModel, &iter) {
		return
	}

	order, e := gunvalue.New(data.model.GetValue(&iter, 5)).Int()
	if log.Err(e, "WidgetTreeView model.GetValue order") {
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
	if log.Err(e, "WidgetTreeView widget.GetSelection") || !sel.GetSelected(&treeModel, &iter) {
		return
	}
	order, e := gunvalue.New(data.model.GetValue(&iter, 5)).Int()
	if log.Err(e, "WidgetTreeView model.GetValue order") {
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
	if val != "" && !log.Err(e, "WidgetTreeView entry.GetText") {
		data.entry.SetText("")

		// Add new iter to model.
		iter := data.model.Append()
		data.model.SetValue(iter, RowKey, val)
		data.model.SetValue(iter, RowName, val)
		data.model.SetValue(iter, 4, true)                            // active
		data.model.SetValue(iter, 5, data.model.IterNChildren(nil)-1) // order

		// Select new iter.
		sel, e := data.widget.GetSelection()
		if !log.Err(e, "WidgetTreeView widget.GetSelection") {
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
	if !log.Err(e, "WidgetTreeView widget.GetSelection") {
		if !sel.GetSelected(&treeModel, &iter) {
			return
		}
	}

	name, e := gunvalue.New(data.model.GetValue(&iter, RowName)).String()
	if !log.Err(e, "WidgetTreeView model.GetValue RowName") {
		data.entry.SetText(name)
	}

	order, e := gunvalue.New(data.model.GetValue(&iter, 5)).Int()
	data.model.Remove(&iter)
	if log.Err(e, "WidgetTreeView model.GetValue order") { // no order nor iters. can't do shit.
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
	key := &gdk.EventKey{event}

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
	log.DEV("grab class is still to do")

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

func toEntry(w *gtk.Widget) *gtk.Entry {
	return &gtk.Entry{Widget: *w, Editable: gtk.Editable{w.Object}}
	// e := &gtk.Entry{Widget: gtk.Widget{glib.InitiallyUnowned{entry.Object}}, Editable: gtk.Editable{entry.Object}}
}
