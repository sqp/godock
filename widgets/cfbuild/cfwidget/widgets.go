// Package cfwidget implements key widgets for the config file builder.
package cfwidget

import (
	"github.com/conformal/gotk3/gdk"
	"github.com/conformal/gotk3/glib"
	"github.com/conformal/gotk3/gtk"

	"github.com/google/shlex"

	"github.com/sqp/godock/libs/helper/cast"
	"github.com/sqp/godock/libs/ternary"
	"github.com/sqp/godock/libs/text/tran" // Translate.

	"github.com/sqp/godock/widgets/cfbuild/cftype"   // Types for config file builder usage.
	"github.com/sqp/godock/widgets/cfbuild/datatype" // Types for config file builder data source.
	"github.com/sqp/godock/widgets/common"
	"github.com/sqp/godock/widgets/gtk/gunvalue"
	"github.com/sqp/godock/widgets/gtk/indexiter" // References of iter to reselect.
	"github.com/sqp/godock/widgets/gtk/newgtk"    // Create widgets.
	"github.com/sqp/godock/widgets/handbook"

	"strconv"
)

// Warning, during the widget build, the key value must be get before you set
// the GetValue callback.

//------------------------------------------------------[ WIDGETS COLLECTION ]--

// CheckButton adds a check button widget.
//
func CheckButton(key *cftype.Key) {
	if key.NbElements > 1 { // TODO: remove temp test
		key.Log().Info("CheckButton multi", key.NbElements, key.Type.String())
	}

	values := key.Value().ListBool()

	var activers []widgetActiver

	for k := 0; k < key.NbElements; k++ {
		var value bool
		if k < len(values) {
			value = values[k]
		}
		w := newgtk.CheckButton()
		w.SetActive(value)
		key.PackSubWidget(w)
		activers = append(activers, w)

		if key.IsType(cftype.KeyBoolCtrl) {
			// 		_allocate_new_buffer;
			// 		data[0] = pKeyBox;
			// 		data[1] = (pFrameVBox != NULL ? pFrameVBox : pGroupBox);
			// 		if (pAuthorizedValuesList != NULL && pAuthorizedValuesList[0] != NULL)
			// 			NbControlled = g_ascii_strtod (pAuthorizedValuesList[0], NULL);
			// 		else
			// 			NbControlled = 1;
			// 		data[2] = GINT_TO_POINTER (NbControlled);
			// 		if (NbControlled < 0)  // a negative value means that the behavior is inverted.
			// 		{
			// 			bValue = !bValue;
			// 			NbControlled = -NbControlled;
			// 		}
			// 		g_signal_connect (G_OBJECT (pOneWidget), "toggled", G_CALLBACK(_cairo_dock_toggle_control_button), data);

			// 		g_object_set_data (G_OBJECT (pKeyBox), "nb-ctrl-widgets", GINT_TO_POINTER (NbControlled));
			// 		g_object_set_data (G_OBJECT (pKeyBox), "one-widget", pOneWidget);

			// 		if (! bValue)  // les widgets suivants seront inactifs.
			// 		{
			// 			CDControlWidget *cw = g_new0 (CDControlWidget, 1);
			// 			pControlWidgets = g_list_prepend (pControlWidgets, cw);
			// 			cw->iNbSensitiveWidgets = 0;
			// 			cw->NbControlled = NbControlled;
			// 			cw->iFirstSensitiveWidget = 1;
			// 			cw->pControlContainer = (pFrameVBox != NULL ? pFrameVBox : pGroupBox);
			// 		}  // sinon le widget suivant est sensitif, donc rien a faire.
		}
	}
	if key.NbElements == 1 {
		key.PackKeyWidget(key,
			func() interface{} { return activers[0].GetActive() },
			func(uncast interface{}) { activers[0].SetActive(uncast.(bool)) },
		)
	} else {
		key.PackKeyWidget(key,
			func() interface{} { return listActiverGet(activers) },
			func(uncast interface{}) { listActiverSet(activers, uncast.([]bool)) },
		)
	}
}

// IntegerScale adds an integer scale widget.
//
func IntegerScale(key *cftype.Key) {
	if key.NbElements > 1 { // TODO: remove temp test
		key.Log().Info("IntegerScale multi", key.NbElements, key.Type.String())
	}

	value := key.Value().Int()
	minValue, maxValue := minMaxValues(key)

	step := (maxValue - minValue) / 20
	if step < 1 {
		step = 1
	}
	adjustment := newgtk.Adjustment(float64(value), minValue, maxValue, 1, step, 0)
	w := newgtk.Scale(gtk.ORIENTATION_HORIZONTAL, adjustment)
	w.Set("digits", 0)

	PackValuerAsInt(key, WrapKeyScale(key, w), w, value)
}

// IntegerSpin adds an integer scale widget.
//
func IntegerSpin(key *cftype.Key) {
	if key.NbElements > 1 { // TODO: remove temp test
		key.Log().Info("integer multi ffs", key.Type.String())
	}

	value := key.Value().Int()
	minValue, maxValue := minMaxValues(key)

	w := newgtk.SpinButtonWithRange(minValue, maxValue, 1)
	w.SetValue(float64(value))

	PackValuerAsInt(key, w, w, value)
}

// IntegerSize adds an integer selector widget.
//
func IntegerSize(key *cftype.Key) {
	if key.NbElements > 1 { // TODO: remove temp test
		key.Log().Info("IntegerSize multi", key.NbElements, key.Type.String())
	}

	toggle := newgtk.ToggleButton()
	img := newgtk.ImageFromIconName("media-playback-pause", gtk.ICON_SIZE_MENU) // get better image.
	toggle.SetImage(img)

	values := key.Value().ListInt()
	minValue, maxValue := minMaxValues(key)

	var valuers []WidgetValuer

	key.NbElements *= 2 // Two widgets to add.

	// Value pair data.
	var firstValue int
	var firstWidget *gtk.SpinButton
	var cbBlock func() func()

	for k := 0; k < key.NbElements; k++ {
		var value int
		if k < len(values) {
			value = values[k]
		}

		w := newgtk.SpinButtonWithRange(minValue, maxValue, 1)
		w.SetValue(float64(value))
		key.PackSubWidget(w)
		valuers = append(valuers, w)

		if k&1 == 0 { // first value, separator
			label := newgtk.Label("x")
			key.PackSubWidget(label)

			firstWidget = w
			firstValue = value

		} else { // second value. connect both spin values.
			if firstValue == value {
				toggle.SetActive(true)
			}

			cb0, e := firstWidget.Connect("value-changed", onValuePairChanged, &valuePair{
				linked: w,
				toggle: toggle})
			key.Log().Err(e, "IntegerSize connect value-changed 1")
			cb1, e := w.Connect("value-changed", onValuePairChanged, &valuePair{
				linked: firstWidget,
				toggle: toggle})
			key.Log().Err(e, "IntegerSize connect value-changed 2")

			cbBlock = func() func() {
				firstWidget.HandlerBlock(cb0)
				w.HandlerBlock(cb1)
				return func() {
					firstWidget.HandlerUnblock(cb0)
					w.HandlerUnblock(cb1)
				}
			}
		}
	}

	setValue := func(uncast interface{}) {
		if cbBlock == nil {
			key.Log().NewErr("no valuePair callbacks", "IntegerSize", key.Name)
		} else {
			defer cbBlock()() // This disables now and reenables during defer.
		}
		values := uncast.([]int)
		if len(values) < 2 {
			key.Log().NewErr("not enough values provided", "IntegerSize set value", key.Name, values)
			values = []int{0, 0}
		}

		listValuerSet(valuers, cast.IntsToFloats(values))
		toggle.SetActive(values[0] == values[1])
	}

	key.PackKeyWidget(key,
		func() interface{} { return cast.FloatsToInts(listValuerGet(valuers)) },
		setValue,
		toggle,
	)

	oldval, _ := key.Storage().Default(key.Group, key.Name)
	PackReset(key, oldval.ListInt())
}

// Float adds a float selector widget. SpinButton or Horizontal Scale
//
func Float(key *cftype.Key) {
	if key.NbElements > 1 { // TODO: remove temp test
		key.Log().Info("Float multi", key.NbElements, key.Type.String())
	}

	values := key.Value().ListFloat()
	minValue, maxValue := minMaxValues(key)

	var valuers []WidgetValuer

	for k := 0; k < key.NbElements; k++ {
		var value float64
		if k < len(values) {
			value = values[k]
		}

		switch key.Type {
		case cftype.KeyFloatScale:
			adjustment := newgtk.Adjustment(
				value,
				minValue,
				maxValue,
				(maxValue-minValue)/20,
				(maxValue-minValue)/10, 0,
			)
			w := newgtk.Scale(gtk.ORIENTATION_HORIZONTAL, adjustment)
			w.Set("digits", 3)

			key.PackSubWidget(WrapKeyScale(key, w))
			valuers = append(valuers, w)

		case cftype.KeyFloatSpin:
			w := newgtk.SpinButtonWithRange(minValue, maxValue, 1)
			w.Set("digits", 3)
			w.SetValue(value)

			key.PackSubWidget(w)
			valuers = append(valuers, w)
		}
	}

	switch {
	case key.NbElements == 1:
		key.PackKeyWidget(key,
			func() interface{} { return valuers[0].GetValue() },
			func(uncast interface{}) { valuers[0].SetValue(uncast.(float64)) },
		)
		oldval, _ := key.Storage().Default(key.Group, key.Name)
		PackReset(key, oldval.Float())

	default:
		key.PackKeyWidget(key,
			func() interface{} { return listValuerGet(valuers) },
			func(uncast interface{}) { listValuerSet(valuers, uncast.([]float64)) },
		)
		oldval, _ := key.Storage().Default(key.Group, key.Name)
		PackReset(key, oldval.ListFloat())
	}
}

// ColorSelector adds a color selector widget.
//
func ColorSelector(key *cftype.Key) {
	values := key.Value().ListFloat()
	if len(values) == 3 {
		values = append(values, 1) // no transparency.
	}
	gdkColor := gdk.NewRGBA(values...)
	widget := newgtk.ColorButtonWithRGBA(gdkColor)

	var getValue func() interface{}
	switch key.Type {
	case cftype.KeyColorSelectorRGB:
		key.NbElements = 3
		getValue = func() interface{} { return widget.GetRGBA().Floats()[:3] } // Need to trunk ?

	case cftype.KeyColorSelectorRGBA:
		key.NbElements = 4
		getValue = func() interface{} { return widget.GetRGBA().Floats() }
	}

	widget.Set("use-alpha", key.IsType(cftype.KeyColorSelectorRGBA))

	key.PackKeyWidget(key,
		getValue,
		func(uncast interface{}) { widget.SetRGBA(gdk.NewRGBA(uncast.([]float64)...)) },
		widget,
	)
	oldval, _ := key.Storage().Default(key.Group, key.Name)
	PackReset(key, oldval.ListFloat())
}

// ListThemeApplet adds an theme list widget.
//
func ListThemeApplet(key *cftype.Key) {
	var index map[string]datatype.Handbooker
	if len(key.AuthorizedValues) > 2 {
		if key.AuthorizedValues[1] == "gauges" {
			index = key.Source().ListThemeXML(key.AuthorizedValues[0], key.AuthorizedValues[1], key.AuthorizedValues[2])

		} else {
			index = key.Source().ListThemeINI(key.AuthorizedValues[0], key.AuthorizedValues[1], key.AuthorizedValues[2])
		}
	}
	PackComboBoxWithIndexHandbooker(key, index)

	// 	// list local packages first.
	// 	_allocate_new_buffer;
	// 	data[0] = pOneWidget;
	// 	data[1] = pMainWindow;
	// 	data[2] = g_key_file_get_string (pKeyFile, cGroupName, cKeyName, NULL);  // freed in the callback '_got_themes_combo_list'.
	// 	data[3] = g_strdup (cHint);  // idem

	// 	GHashTable *pThemeTable = cairo_dock_list_packages (cShareThemesDir, cUserThemesDir, NULL, NULL);
	// 	_got_themes_combo_list (pThemeTable, (gpointer*)data);

	// 	// list distant packages asynchronously.
	// 	if (cDistantThemesDir != NULL)
	// 	{
	// 		cairo_dock_set_status_message_printf (pMainWindow, _("Listing themes in '%s' ..."), cDistantThemesDir);
	// 		data[2] = g_key_file_get_string (pKeyFile, cGroupName, cKeyName, NULL);  // freed in the callback '_got_themes_combo_list'.
	// 		data[3] = g_strdup (cHint);
	// 		CairoDockTask *pTask = cairo_dock_list_packages_async (NULL, NULL, cDistantThemesDir, (CairoDockGetPackagesFunc) _got_themes_combo_list, data, pThemeTable);  // the table will be freed along with the task.
	// 		g_object_set_data (G_OBJECT (pOneWidget), "cd-task", pTask);
	// 		g_signal_connect (G_OBJECT (pOneWidget), "destroy", G_CALLBACK (on_delete_async_widget), NULL);
	// 	}
	// 	else
	// 	{
	// 		g_hash_table_destroy (pThemeTable);
	// 	}
	// 	g_free (cUserThemesDir);
	// 	g_free (cShareThemesDir);

}

// ListView adds a view list widget.
//
func ListView(key *cftype.Key) {
	index := key.Source().ListViews()
	PackComboBoxWithIndexHandbooker(key, index)
}

// ListThemeDesktopIcon adds a desktop icon-themes list widget.
//
func ListThemeDesktopIcon(key *cftype.Key) {
	getList := fieldsPrepend(key.Source().ListThemeDesktopIcon(),
		datatype.Field{},
		datatype.Field{Key: "_Custom Icons_", Name: tran.Slate("_Custom Icons_")},
	)
	PackComboBoxWithListField(key, false, false, getList)
}

// ListAnimation adds an animation list widget.
//
func ListAnimation(key *cftype.Key) {
	getList := fieldsPrepend(key.Source().ListAnimations(),
		datatype.Field{},
	)
	PackComboBoxWithListField(key, false, false, getList)
}

// ListDialogDecorator adds an dialog decorator list widget.
//
func ListDialogDecorator(key *cftype.Key) {
	PackComboBoxWithListField(key, false, false, key.Source().ListDialogDecorator)
}

// ListDeskletDecoration adds a desklet decoration list widget.
//
func ListDeskletDecoration(key *cftype.Key) {
	current := key.Value().String()
	getList := key.Source().ListDeskletDecorations
	if key.IsType(cftype.KeyListDeskletDecoDefault) {
		getList = fieldsPrepend(getList(),
			datatype.Field{Key: "default", Name: tran.Slate("default")},
		)
	}
	PackComboBoxWithListField(key, false, false, getList)

	// 	gtk_tree_sortable_set_sort_column_id (GTK_TREE_SORTABLE (pListStore), CAIRO_DOCK_MODEL_NAME, GTK_SORT_ASCENDING);

	// _allocate_new_buffer;
	// data[0] = pKeyBox;
	// data[1] = (pFrameVBox != NULL ? pFrameVBox : pGroupBox);
	// NbControlled = 9;
	// data[2] = GINT_TO_POINTER (NbControlled);
	// NbControlled --;  // car dans cette fonction, on ne compte pas le separateur.
	// g_signal_connect (G_OBJECT (pOneWidget), "changed", G_CALLBACK (_cairo_dock_select_custom_item_in_combo), data);

	if current == "personnal" { // Disable the next widgets.
		// 		CDControlWidget *cw = g_new0 (CDControlWidget, 1);
		// 		pControlWidgets = g_list_prepend (pControlWidgets, cw);
		// 		cw->NbControlled = NbControlled;
		// 		cw->iNbSensitiveWidgets = 0;
		// 		cw->iFirstSensitiveWidget = 1;
		// 		cw->pControlContainer = (pFrameVBox != NULL ? pFrameVBox : pGroupBox);
	}
}

// ListScreens adds a screen selection widget.
//
func ListScreens(key *cftype.Key) {
	list := key.Source().ListScreens()
	combo := PackComboBoxWithListField(key, false, false, fieldsPrepend(list))
	if len(list) <= 1 {
		combo.SetSensitive(false)
	}

	// 	gldi_object_register_notification (&myDesktopMgr,
	// 		NOTIFICATION_DESKTOP_GEOMETRY_CHANGED,
	// 		(GldiNotificationFunc) _on_screen_modified,
	// 		GLDI_RUN_AFTER, pScreensListStore);
	// 	g_signal_connect (pOneWidget, "destroy", G_CALLBACK (_on_list_destroyed), NULL);
}

// ListDock adds a dock list widget.
//
func ListDock(key *cftype.Key) {
	// Get current Icon name if its a Subdock.
	iIconType, _ := key.Storage().Int(key.Group, "Icon Type")
	SubdockName := ""
	if iIconType == cftype.UserIconStack { // it's a stack-icon
		SubdockName, _ = key.Storage().String(key.Group, "Name") // It's a subdock, get its name to remove the selection of a recursive position (inside itself).
	}

	list := key.Source().ListDocks("", SubdockName) // Get the list of available docks. Keep parent, but remove itself from the list.
	list = append(list, datatype.Field{
		Key:  datatype.KeyNewDock,
		Name: tran.Slate("New main dock")},
	)

	model := newModelSimple()
	current := key.Value().String()

	if current == "" {
		current = datatype.KeyMainDock
	}

	model.SetSortColumnId(RowName, gtk.SORT_ASCENDING)
	widget := newgtk.ComboBoxWithModel(model)
	renderer := newgtk.CellRendererText()
	widget.PackStart(renderer, false)
	widget.AddAttribute(renderer, "text", RowName)

	saved := indexiter.NewByString(widget, key.Log())
	iter := fillModelWithFields(key, model, list, current, saved)
	widget.SetActiveIter(iter)

	key.PackKeyWidget(key,
		getValueListCombo(widget, model, key.Log()),
		func(uncast interface{}) { saved.SetActive(uncast.(string)) },
		widget)
}

// ListIconsMainDock adds an icon list widget.
//
func ListIconsMainDock(key *cftype.Key) {
	// {
	// 	if (g_pMainDock == NULL) // maintenance mode... no dock, no icons
	// 	{
	// 		cValue = g_key_file_get_string (pKeyFile, cGroupName, cKeyName, NULL);

	// 		pOneWidget = gtk_entry_new ();
	// 		gtk_entry_set_text (GTK_ENTRY (pOneWidget), cValue); // if there is a problem, we can edit it.
	// 		_pack_subwidget (pOneWidget);

	// 		g_free (cValue);
	// 		break;
	// 	}

	//

	//

	current := key.Value().String()
	model := newModelSimple()
	widget := newgtk.ComboBoxWithModel(model)

	rp := newgtk.CellRendererPixbuf()
	widget.PackStart(rp, false)
	widget.AddAttribute(rp, "pixbuf", RowIcon)

	renderer := newgtk.CellRendererText()
	widget.PackStart(renderer, true)
	widget.AddAttribute(renderer, "text", RowName)

	list := key.Source().ListIconsMainDock()
	saved := indexiter.NewByString(widget, key.Log())
	iter := fillModelWithFields(key, model, list, current, saved)
	widget.SetActiveIter(iter)

	//

	// 	// build the modele and combo
	// 	modele = _cairo_dock_gui_allocate_new_model ();
	// 	pOneWidget = gtk_combo_box_new_with_model (GTK_TREE_MODEL (modele));
	// 	rend = gtk_cell_renderer_pixbuf_new ();
	// 	gtk_cell_layout_pack_start (GTK_CELL_LAYOUT (pOneWidget), rend, FALSE);
	// 	gtk_cell_layout_set_attributes (GTK_CELL_LAYOUT (pOneWidget), rend, "pixbuf", CAIRO_DOCK_MODEL_ICON, NULL);
	// 	rend = gtk_cell_renderer_text_new ();
	// 	gtk_cell_layout_pack_start (GTK_CELL_LAYOUT (pOneWidget), rend, FALSE);
	// 	gtk_cell_layout_set_attributes (GTK_CELL_LAYOUT (pOneWidget), rend, "text", CAIRO_DOCK_MODEL_NAME, NULL);
	// 	_pack_subwidget (pOneWidget);

	// 	// get the dock
	// 	CairoDock *pDock = NULL;
	// 	if (pAuthorizedValuesList != NULL && pAuthorizedValuesList[0] != NULL)
	// 		pDock = gldi_dock_get (pAuthorizedValuesList[0]);
	// 	if (!pDock)
	// 		pDock = g_pMainDock;

	// 	// insert each icon
	// 	cValue = g_key_file_get_string (pKeyFile, cGroupName, cKeyName, NULL);
	// 	gint iDesiredIconSize = cairo_dock_search_icon_size (GTK_ICON_SIZE_LARGE_TOOLBAR); // 24 by default
	// 	GtkTreeIter iter;
	// 	Icon *pIcon;
	// 	gchar *cImagePath, *cID;
	// 	const gchar *cName;
	// 	GdkPixbuf *pixbuf;
	// 	GList *ic;
	// 	for (ic = pDock->icons; ic != NULL; ic = ic->next)
	// 	{
	// 		pIcon = ic->data;
	// 		if (pIcon->cDesktopFileName != NULL
	// 		|| pIcon->pModuleInstance != NULL)
	// 		{
	// 			pixbuf = NULL;
	// 			cImagePath = NULL;
	// 			cName = NULL;

	// 			// get the ID
	// 			if (pIcon->cDesktopFileName != NULL)
	// 				cID = pIcon->cDesktopFileName;
	// 			else
	// 				cID = pIcon->pModuleInstance->cConfFilePath;

	// 			// get the image
	// 			if (pIcon->cFileName != NULL)
	// 			{
	// 				cImagePath = cairo_dock_search_icon_s_path (pIcon->cFileName, iDesiredIconSize);
	// 			}
	// 			if (cImagePath == NULL || ! g_file_test (cImagePath, G_FILE_TEST_EXISTS))
	// 			{
	// 				g_free (cImagePath);
	// 				if (GLDI_OBJECT_IS_SEPARATOR_ICON (pIcon))
	// 				{
	// 					if (myIconsParam.cSeparatorImage)
	// 						cImagePath = cairo_dock_search_image_s_path (myIconsParam.cSeparatorImage);
	// 				}
	// 				else if (CAIRO_DOCK_IS_APPLET (pIcon))
	// 				{
	// 					cImagePath = g_strdup (pIcon->pModuleInstance->pModule->pVisitCard->cIconFilePath);
	// 				}
	// 				else
	// 				{
	// 					cImagePath = cairo_dock_search_image_s_path (CAIRO_DOCK_DEFAULT_ICON_NAME);
	// 					if (cImagePath == NULL || ! g_file_test (cImagePath, G_FILE_TEST_EXISTS))
	// 					{
	// 						g_free (cImagePath);
	// 						cImagePath = g_strdup (GLDI_SHARE_DATA_DIR"/icons/"CAIRO_DOCK_DEFAULT_ICON_NAME);
	// 					}
	// 				}
	// 			}
	// 			//g_print (" + %s\n", cImagePath);
	// 			if (cImagePath != NULL)
	// 			{
	// 				pixbuf = gdk_pixbuf_new_from_file_at_size (cImagePath, iDesiredIconSize, iDesiredIconSize, NULL);
	// 			}
	// 			//g_print (" -> %p\n", pixbuf);

	// 			// get the name
	// 			if (CAIRO_DOCK_IS_USER_SEPARATOR (pIcon))  // separator
	// 				cName = "---------";
	// 			else if (CAIRO_DOCK_IS_APPLET (pIcon))  // applet
	// 				cName = pIcon->pModuleInstance->pModule->pVisitCard->cTitle;
	// 			else  // launcher
	// 				cName = (pIcon->cInitialName ? pIcon->cInitialName : pIcon->cName);

	// 			// store the icon
	// 			memset (&iter, 0, sizeof (GtkTreeIter));
	// 			gtk_list_store_append (GTK_LIST_STORE (modele), &iter);
	// 			gtk_list_store_set (GTK_LIST_STORE (modele), &iter,
	// 				CAIRO_DOCK_MODEL_NAME, cName,
	// 				CAIRO_DOCK_MODEL_RESULT, cID,
	// 				CAIRO_DOCK_MODEL_ICON, pixbuf, -1);
	// 			g_free (cImagePath);
	// 			if (pixbuf)
	// 				g_object_unref (pixbuf);

	// 			if (cValue && strcmp (cValue, cID) == 0)
	// 				gtk_combo_box_set_active_iter (GTK_COMBO_BOX (pOneWidget), &iter);
	// 		}
	// 	}
	// 	g_free (cValue);
	// }

	key.PackKeyWidget(key,
		getValueListCombo(widget, model, key.Log()),
		func(uncast interface{}) { saved.SetActive(uncast.(string)) },
		widget)
}

// JumpToModule adds a redirect button widget.
// USED?
//
func JumpToModule(key *cftype.Key) {
	// if (pAuthorizedValuesList == NULL || pAuthorizedValuesList[0] == NULL || *pAuthorizedValuesList[0] == '\0')
	// 	break ;

	// gchar *cModuleName = NULL;
	// GldiModule *pModule = gldi_module_get (pAuthorizedValuesList[0]);
	// if (pModule != NULL)
	// 	cModuleName = (gchar*)pModule->pVisitCard->cModuleName;  // 'cModuleName' will not be freed
	// else
	// {
	// 	if (iElementType == CAIRO_DOCK_WidgetJumpToModuleIfExists)
	// 	{
	// 		gtk_widget_set_sensitive (pLabel, FALSE);
	// 		break ;
	// 	}
	// 	cd_warning ("module '%s' not found", pAuthorizedValuesList[0]);
	// 	cModuleName = g_strdup (pAuthorizedValuesList[0]);  // petite fuite memoire dans ce cas tres rare ...
	// }
	// pOneWidget = gtk_button_new_from_stock (GTK_STOCK_JUMP_TO);
	// g_signal_connect (G_OBJECT (pOneWidget),
	// 	"clicked",
	// 	G_CALLBACK (_cairo_dock_configure_module),
	// 	cModuleName);
	// _pack_subwidget (pOneWidget);

}

// LaunchCommand adds a launch command widget.
// HELP ONLY
//
func LaunchCommand(key *cftype.Key) {
	if len(key.AuthorizedValues) == 0 || key.AuthorizedValues[0] == "" {
		key.Log().NewErrf("command missing", "widget LaunchCommand: %s", key.Name)
		return
	}
	// log.Info(key.Name, key.AuthorizedValues)

	if key.IsType(cftype.KeyLaunchCmdIf) {

		key.Log().Info("KeyLaunchCmdIf : disabled for now")
		return

		if len(key.AuthorizedValues) < 2 {
			key.Label().SetSensitive(false)
			return
		}
		// key.Log().Info("test", key.AuthorizedValues[1])

		// key.Log().Err(key.Log().ExecShow(key.AuthorizedValues[1]), "exec test")

		// gchar *cSecondCommand = pAuthorizedValuesList[1];
		// gchar *cResult = cairo_dock_launch_command_sync (cSecondCommand);
		// cd_debug ("%s: %s => %s", __func__, cSecondCommand, cResult);
		// if (cResult == NULL || *cResult == '0' || *cResult == '\0')  // result is 'fail'
		// {
		// 	gtk_widget_set_sensitive (pLabel, FALSE);
		// 	g_free (cResult);
		// 	break ;
		// }
		// g_free (cResult);
	}

	args, e := shlex.Split(key.AuthorizedValues[0])
	if key.Log().Err(e, "widget LaunchCommand parse command", key.Name, ":", key.AuthorizedValues[0], "==>", args) {
		return
	}

	// key.Log().DEV("new args", len(args), args)

	spinner := newgtk.Spinner()
	spinner.SetNoShowAll(true)
	key.PackSubWidget(spinner)

	btn := newgtk.ButtonFromIconName("go-jump", gtk.ICON_SIZE_BUTTON)
	key.PackSubWidget(btn)

	btn.Connect("clicked", func() {
		cmd := key.Log().ExecCmd(args[0], args[1:]...)

		e := cmd.Start()
		if key.Log().Err(e, "widget LaunchCommand exec", key.AuthorizedValues[0]) {
			return
		}

		btn.Hide()
		spinner.Show()
		spinner.Start()

		// Wait the external program in a go routine.
		// When finished, restore buttons state in the gtk idle loop.
		go func() {
			cmd.Wait()

			glib.IdleAdd(func() {
				btn.Show()
				spinner.Hide()
				spinner.Stop()
			})
		}()
	})

	// 	G_CALLBACK (_cairo_dock_widget_launch_command),
}

// Lists adds a string list widget.
//
func Lists(key *cftype.Key) {
	if key.IsType(cftype.KeyListNbCtrlSimple, cftype.KeyListNbCtrlSelect) && len(key.AuthorizedValues) == 0 {
		key.Log().NewWarn("not enough values", "widget numbered control list:", key.Name)
		return
	}

	// Get full value with ';'.
	value := key.Value().String()

	// log.DEV("LIST "+string(key.Type), key.Name, value, key.AuthorizedValues)

	listIsNumbered := key.IsType(cftype.KeyListNumbered, cftype.KeyListNbCtrlSimple, cftype.KeyListNbCtrlSelect)

	iSelectedItem := -1
	current := ""

	// Control selective use 3 AuthorizedValues fields for each "value".
	step := ternary.Int(key.IsType(cftype.KeyListNbCtrlSelect), 3, 1)

	if key.IsType(cftype.KeyListEntry) {
		current = value
	}

	if listIsNumbered && value != "" {
		var e error
		iSelectedItem, e = strconv.Atoi(value)
		switch {
		case key.Log().Err(e, "selection problem", "[", value, "]", "[", iSelectedItem, "]", key.Name):
		case iSelectedItem < 0:

		case iSelectedItem < len(key.AuthorizedValues):
			current = key.AuthorizedValues[iSelectedItem*step]

		default:
			key.Log().NewWarn("selection out of range", "widget numbered list:", key.Name)
		}
	}

	var list []datatype.Field
	if len(key.AuthorizedValues) > 0 {
		// int iOrder1, iOrder2, iExcept;

		if key.IsType(cftype.KeyListNbCtrlSimple, cftype.KeyListNbCtrlSelect) {
			key.SetNbControlled(0)
		}

		// 	gchar *cResult = (listIsNumbered ? g_new0 (gchar , 10) : NULL);

		for k := 0; k < len(key.AuthorizedValues); k += step { // on ajoute toutes les chaines possibles a la combo.
			if !listIsNumbered && iSelectedItem == -1 && value == key.AuthorizedValues[k] {
				current = value
				iSelectedItem = k / step
			}

			// 		if (cResult != NULL)
			// 			snprintf (cResult, 9, "%d", k/dk); // dk becomes step

			// 		iExcept = 0;
			// 		if key.IsType(cftype.KeyListNbCtrlSelect) 		{
			// 			iOrder1 = atoi (key.AuthorizedValues[k+1]);
			// 			gchar *str = strchr (key.AuthorizedValues[k+2], ',');
			// 			if (str)  // Note: this mechanism is an addition to the original {first widget, number of widgets}; it's not very generic nor beautiful, but until we need more, it's well enough (currently, only the Dock background needs it).
			// 			{
			// 				*str = '\0';
			// 				iExcept = atoi (str+1);
			// 			}
			// 			iOrder2 = atoi (key.AuthorizedValues[k+2]);
			// 			NbControlled = MAX (NbControlled, iOrder1 + iOrder2 - 1);
			// 			//g_print ("iSelectedItem:%d ; k/dk:%d\n", iSelectedItem , k/dk);
			// 			if (iSelectedItem == (int)k/dk)
			// 			{
			// 				iFirstSensitiveWidget = iOrder1;
			// 				iNbSensitiveWidgets = iOrder2;
			// 				iNonSensitiveWidget = iExcept;
			// 				if (iNonSensitiveWidget != 0)
			// 					NbControlled ++;
			// 			}
			// 		}					else					{
			// 			iOrder1 = iOrder2 = k;
			// 		}

			//
			name := ""
			if key.IsType(cftype.KeyListEntry) {
				name = key.AuthorizedValues[k]
			} else {
				name = key.Translate(key.AuthorizedValues[k])
			}

			list = append(list, datatype.Field{
				Key:  key.AuthorizedValues[k],
				Name: name,
			})
			// 			CAIRO_DOCK_MODEL_ORDER, iOrder1,
			// 			CAIRO_DOCK_MODEL_ORDER2, iOrder2,
			// 			CAIRO_DOCK_MODEL_STATE, iExcept, -1);
		}
	}

	// Current choice wasn't in the list. Select first.
	if current == "" && len(list) > 0 {
		current = list[0].Key
		iSelectedItem = 0
	}

	// Build the combo widget.
	widget, _, getValue, setValue := NewComboBox(key, key.IsType(cftype.KeyListEntry), listIsNumbered, current, list)

	// gtk_tree_sortable_set_sort_column_id (GTK_TREE_SORTABLE (modele), GTK_TREE_SORTABLE_UNSORTED_SORT_COLUMN_ID, GTK_SORT_ASCENDING);

	// int iNonSensitiveWidget = 0;

	if len(key.AuthorizedValues) > 0 {
		if key.IsType(cftype.KeyListNbCtrlSimple, cftype.KeyListNbCtrlSelect) {
			// 		_allocate_new_buffer;
			// 		data[0] = pKeyBox;
			// 		data[1] = (pFrameVBox != NULL ? pFrameVBox : pGroupBox);
			if key.IsType(cftype.KeyListNbCtrlSimple) {
				// 			NbControlled = k;
				// 			data[2] = GINT_TO_POINTER (NbControlled);
				// 			g_signal_connect (G_OBJECT (pOneWidget), "changed", G_CALLBACK (_cairo_dock_select_one_item_in_control_combo), data);
				// 			iFirstSensitiveWidget = iSelectedItem+1;  // on decroit jusqu'a 0.
				// 			iNbSensitiveWidgets = 1;
				// 			//g_print ("CONTROL : %d,%d,%d\n", NbControlled, iFirstSensitiveWidget, iNbSensitiveWidgets);
			} else {
				// 			data[2] = GINT_TO_POINTER (NbControlled);
				// 			g_signal_connect (G_OBJECT (pOneWidget), "changed", G_CALLBACK (_cairo_dock_select_one_item_in_control_combo_selective), data);
				// 			//g_print ("CONTROL : %d,%d,%d\n", NbControlled, iFirstSensitiveWidget, iNbSensitiveWidgets);
			}
			// 		g_object_set_data (G_OBJECT (pKeyBox), "nb-ctrl-widgets", GINT_TO_POINTER (NbControlled));
			// 		g_object_set_data (G_OBJECT (pKeyBox), "one-widget", pOneWidget);
			// 		CDControlWidget *cw = g_new0 (CDControlWidget, 1);
			// 		pControlWidgets = g_list_prepend (pControlWidgets, cw);
			// 		cw->pControlContainer = (pFrameVBox != NULL ? pFrameVBox : pGroupBox);
			// 		cw->NbControlled = NbControlled;
			// 		cw->iFirstSensitiveWidget = iFirstSensitiveWidget;
			// 		cw->iNbSensitiveWidgets = iNbSensitiveWidgets;
			// 		cw->iNonSensitiveWidget = iNonSensitiveWidget;
			// 		//g_print (" pControlContainer:%x\n", pControlContainer);
		}
	}
	key.PackKeyWidget(key, getValue, setValue, widget)
}

// TreeView adds a treeview widget.
//
func TreeView(key *cftype.Key) {
	values := key.Value().ListString()

	// Build treeview.
	model := newgtk.ListStore(
		glib.TYPE_STRING,  /* RowKey*/
		glib.TYPE_STRING,  /* RowName*/
		glib.TYPE_STRING,  /* RowIcon*/
		glib.TYPE_STRING,  /* RowDesc*/
		glib.TYPE_BOOLEAN, // active
		glib.TYPE_INT)     // order

	widget := newgtk.TreeViewWithModel(model)
	widget.Set("headers-visible", false)

	getValue := func() interface{} { // Grab data from all iters.
		var list []string
		iter, ok := model.GetIterFirst()
		for ; ok; ok = model.IterNext(iter) {
			str, e := gunvalue.New(model.GetValue(iter, RowName)).String()
			if !key.Log().Err(e, "WidgetTreeView GetValue "+key.Name) {
				list = append(list, str)
			}
		}
		return list
	}

	// Add control buttons.
	if key.IsType(cftype.KeyTreeViewMultiChoice) {
		renderer := newgtk.CellRendererToggle()
		col := newgtk.TreeViewColumnWithAttribute("", renderer, "active", 4)
		widget.AppendColumn(col)
		// 	g_signal_connect (G_OBJECT (rend), "toggled", (GCallback) _cairo_dock_activate_one_element, modele);

		renderer.Set("active", 4)
	}

	renderer := newgtk.CellRendererText()
	col := newgtk.TreeViewColumnWithAttribute("", renderer, "text", RowName)
	widget.AppendColumn(col)

	// cValueList = g_key_file_get_string_list (pKeyFile, cGroupName, cKeyName, &length, NULL);

	model.SetSortColumnId(5, gtk.SORT_ASCENDING)

	scroll := newgtk.ScrolledWindow(nil, nil)

	//

	// if len(key.AuthorizedValues) > 0 {
	// 	key.Log().Info("WidgetTreeView AuthorizedValues", key.AuthorizedValues)
	// }

	// if (key.AuthorizedValues != NULL && key.AuthorizedValues[0] != NULL)
	// 	for (k = 0; key.AuthorizedValues[k] != NULL; k++);
	// else
	// 	k = 1;
	scroll.Set("height-request", 100) // key.IsType(cftype.KeyTreeViewSortModify) ? 100 : MIN (100, k * 25)
	scroll.Set("width-request", 250)
	scroll.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_AUTOMATIC)

	scroll.Add(widget)

	vboxItems := newgtk.Box(gtk.ORIENTATION_VERTICAL, cftype.MarginGUI)
	grid := newgtk.Grid()
	grid.Attach(vboxItems, 0, 0, 1, 1)
	grid.Attach(scroll, 1, 0, 1, 1)

	if key.IsType(cftype.KeyTreeViewSortSimple, cftype.KeyTreeViewSortModify) {

		buttonUp := newgtk.Button()
		buttonDn := newgtk.Button()
		imgUp := newgtk.ImageFromIconName("go-up", gtk.ICON_SIZE_SMALL_TOOLBAR)
		imgDn := newgtk.ImageFromIconName("go-down", gtk.ICON_SIZE_SMALL_TOOLBAR)

		buttonUp.SetImage(imgUp)
		buttonDn.SetImage(imgDn)

		data := treeViewData{key.Log(), model, widget, nil}

		buttonUp.Connect("clicked", onTreeviewMoveUp, data) // Move selection up and down callbacks.
		buttonDn.Connect("clicked", onTreeviewMoveDown, data)

		vboxItems.PackStart(buttonUp, false, false, 0)
		vboxItems.PackStart(buttonDn, false, false, 0)

		if key.IsType(cftype.KeyTreeViewSortModify) {

			buttonAdd := newgtk.Button()
			entry := newgtk.Entry()
			buttonRm := newgtk.Button()

			imgAdd := newgtk.ImageFromIconName("list-add", gtk.ICON_SIZE_SMALL_TOOLBAR)
			imgRm := newgtk.ImageFromIconName("list-remove", gtk.ICON_SIZE_SMALL_TOOLBAR)
			buttonAdd.SetImage(imgAdd)
			buttonRm.SetImage(imgRm)

			vboxItems.PackStart(buttonRm, false, false, 0)
			grid.Attach(newgtk.Separator(gtk.ORIENTATION_HORIZONTAL), 0, 1, 2, 1)
			grid.Attach(buttonAdd, 0, 2, 1, 1)
			grid.Attach(entry, 1, 2, 1, 1)

			data.entry = entry
			buttonAdd.Connect("clicked", onTreeviewAddText, data)   // Add new iter to model with the value of the entry widget. Clear entry widget.
			buttonRm.Connect("clicked", onTreeviewRemoveText, data) // Remove selected iter from model. Set its value to the entry widget.
		}
	}

	setValues := func(newvalues []string) {
		for i, val := range newvalues {
			iter := model.Append()
			model.SetValue(iter, RowKey, val)
			model.SetValue(iter, RowName, val)

			model.SetValue(iter, 4, true) // active
			model.SetValue(iter, 5, i)    // order
		}
	}

	// Fill model with values.
	switch key.Type {
	case cftype.KeyTreeViewSortModify, // add saved values.
		cftype.KeyTreeViewSortSimple: // TODO: TEMP to improve and maybe regroup this case as was with multichoice and not modify.

		setValues(values)

		key.PackKeyWidget(key,
			getValue,
			func(uncast interface{}) { model.Clear(); setValues(uncast.([]string)) },
			grid)

	case cftype.KeyTreeViewMultiChoice:
		if len(key.AuthorizedValues) > 0 {
			// var NbMax, order int

		}
		// else if (pAuthorizedValuesList != NULL)  // on liste les choix possibles dans l'ordre choisi. Pour CAIRO_DOCK_WidgetTreeViewMultiChoice, on complete avec ceux n'ayant pas ete selectionnes.
		// {
		// 	gint iNbPossibleValues = 0, iOrder = 0;
		// 	while (pAuthorizedValuesList[iNbPossibleValues] != NULL)
		// 		iNbPossibleValues ++;
		// 	guint l;
		// 	for (l = 0; l < length; l ++)
		// 	{
		// 		cValue = cValueList[l];
		// 		if (! g_ascii_isdigit (*cValue))  // ancien format.
		// 		{
		// 			cd_debug ("old format\n");
		// 			int k;
		// 			for (k = 0; k < iNbPossibleValues; k ++)  // on cherche la correspondance.
		// 			{
		// 				if (strcmp (cValue, pAuthorizedValuesList[k]) == 0)
		// 				{
		// 					cd_debug (" correspondance %s <-> %d", cValue, k);
		// 					g_free (cValueList[l]);
		// 					cValueList[l] = g_strdup_printf ("%d", k);
		// 					cValue = cValueList[l];
		// 					break ;
		// 				}
		// 			}
		// 			if (k < iNbPossibleValues)
		// 				iValue = k;
		// 			else
		// 				continue;
		// 		}
		// 		else
		// 			iValue = atoi (cValue);

		// 		if (iValue < iNbPossibleValues)
		// 		{
		// 			memset (&iter, 0, sizeof (GtkTreeIter));
		// 			gtk_list_store_append (modele, &iter);
		// 			gtk_list_store_set (modele, &iter,
		// 				CAIRO_DOCK_MODEL_ACTIVE, TRUE,
		// 				CAIRO_DOCK_MODEL_NAME, dgettext (cGettextDomain, pAuthorizedValuesList[iValue]),
		// 				CAIRO_DOCK_MODEL_RESULT, cValue,
		// 				CAIRO_DOCK_MODEL_ORDER, iOrder ++, -1);
		// 		}
	}

	// 	if (iOrder < iNbPossibleValues)  // il reste des valeurs a inserer (ce peut etre de nouvelles valeurs apparues lors d'une maj du fichier de conf, donc CAIRO_DOCK_WidgetTreeViewSortSimple est concerne aussi).
	// 	{
	// 		gchar cResult[10];
	// 		for (k = 0; pAuthorizedValuesList[k] != NULL; k ++)
	// 		{
	// 			cValue =  pAuthorizedValuesList[k];
	// 			for (l = 0; l < length; l ++)
	// 			{
	// 				iValue = atoi (cValueList[l]);
	// 				if (iValue == (int)k)  // a deja ete inseree.
	// 					break;
	// 			}

	// 			if (l == length)  // elle n'a pas encore ete inseree.
	// 			{
	// 				snprintf (cResult, 9, "%d", k);
	// 				memset (&iter, 0, sizeof (GtkTreeIter));
	// 				gtk_list_store_append (modele, &iter);
	// 				gtk_list_store_set (modele, &iter,
	// 					CAIRO_DOCK_MODEL_ACTIVE, (iElementType == CAIRO_DOCK_WidgetTreeViewSortSimple),
	// 					CAIRO_DOCK_MODEL_NAME, dgettext (cGettextDomain, cValue),
	// 					CAIRO_DOCK_MODEL_RESULT, cResult,
	// 					CAIRO_DOCK_MODEL_ORDER, iOrder ++, -1);
	// 			}
	// 		}
	// 	}
	// }
}

// FontSelector adds a font selector widget.
//
func FontSelector(key *cftype.Key) {
	value := key.Value().String()
	widget := newgtk.FontButtonWithFont(value)
	widget.Set("show-style", true)
	widget.Set("show-size", true)
	widget.Set("use-size", true)
	widget.Set("use-font", true)

	key.PackKeyWidget(key,
		func() interface{} { return widget.GetFontName() },
		func(val interface{}) { widget.SetFontName(val.(string)) },
		widget)
}

// Link adds a link widget.
//
func Link(key *cftype.Key) {
	var text, link string
	if len(key.AuthorizedValues) > 0 {
		text = key.AuthorizedValues[0]
	} else {
		text = tran.Slate("link")
	}

	if len(key.AuthorizedValues) > 1 { // Custom keys have to use this input way.
		link = key.AuthorizedValues[1]
	} else {
		link = key.Value().String()
	}

	widget := newgtk.LinkButtonWithLabel(link, text)

	key.PackKeyWidget(key,
		func() interface{} { return widget.GetUri() },
		func(val interface{}) { widget.SetUri(val.(string)) },
		widget)
}

// Strings adds a string widget. Many options included.
// TODO: need password cypher, G_USER_DIRECTORY_PICTURES, play sound.
//
func Strings(key *cftype.Key) {
	value := key.Value().String()
	widget := newgtk.Entry()
	widget.SetText(value)

	if key.IsType(cftype.KeyPasswordEntry) { // on cache le texte entre et on decrypte 'cValue'.
		widget.SetVisibility(false)
		// gchar *cDecryptedString = NULL;
		// cairo_dock_decrypt_string ( cValue, &cDecryptedString );
		// g_free (cValue);
		// cValue = cDecryptedString;
	}

	// Pack the widget before any other (in full size if needed).
	// So we do it now and fill the callbacks later
	key.PackKeyWidget(key, nil, nil, widget)

	// 	Add special buttons to fill the entry box.
	switch key.Type {
	case cftype.KeyFileSelector, cftype.KeyFolderSelector,
		cftype.KeySoundSelector, cftype.KeyImageSelector: // Add a file selector.

		fileChooser := newgtk.Button()
		img := newgtk.ImageFromIconName("document-open", gtk.ICON_SIZE_SMALL_TOOLBAR)
		fileChooser.SetImage(img)
		fileChooser.Connect("clicked", onFileChooserOpen, fileChooserData{widget, key})

		key.PackSubWidget(fileChooser)

		if key.IsType(cftype.KeySoundSelector) { //Sound Play Button
			play := newgtk.Button()
			imgPlay := newgtk.ImageFromIconName("media-playback-start", gtk.ICON_SIZE_SMALL_TOOLBAR)
			play.SetImage(imgPlay)

			// play.Connect("clicked", C._cairo_dock_play_a_sound, data)

			key.PackSubWidget(play)
		}

	case cftype.KeyShortkeySelector, cftype.KeyClassSelector: // Add a key/class grabber.
		grab := newgtk.ButtonWithLabel("Grab")
		key.PackSubWidget(grab)
		// 		gtk_widget_add_events(pMainWindow, GDK_KEY_PRESS_MASK);

		switch key.Type {
		case cftype.KeyClassSelector:
			grab.Connect("clicked", onClassGrabClicked)

		case cftype.KeyShortkeySelector:
			grab.Connect("clicked", onKeyGrabClicked, &textGrabData{entry: widget, win: key.Source().GetWindow()})
		}

		// for _, sk := range key.Source().ListShortkeys() {
		// 	if sk.GetConfFilePath() == key.Storage().FilePath() {
		// 		println("found file shortkey")
		// 	}
		// }

	}

	var setValue func(interface{})
	// Display a default value when empty.
	if len(key.AuthorizedValues) > 0 && key.AuthorizedValues[0] != "" {
		defaultText := key.Translate(key.AuthorizedValues[0])
		cbChanged, _ := widget.Connect("changed", onTextDefaultChanged, key)
		data := textDefaultData{key: key, text: defaultText, cbID: cbChanged}
		widget.Connect("focus-in-event", onTextDefaultFocusIn, data)
		widget.Connect("focus-out-event", onTextDefaultFocusOut, data)

		// TODO: check other use of those fields.
		// 	 g_object_set_data (G_OBJECT (pEntry), "ignore-value", GINT_TO_POINTER (TRUE));
		// 	 g_object_set_data (G_OBJECT (pOneWidget), "default-text", cDefaultText);

		setValue = func(uncast interface{}) {
			// if !key.IsDefault { // not sure why this was here.
			widget.SetText(uncast.(string))
			onTextDefaultFocusOut(widget, nil, data)
		}

		setValue(value) // will set IsDefault and button state.
	} else {
		setValue = func(uncast interface{}) { widget.SetText(uncast.(string)) }
	}
	getValue := func() (text interface{}) {
		if key.IsDefault {
			return ""
		}
		text, _ = widget.GetText()
		return text
	}

	key.PackKeyWidget(key, getValue, setValue)
}

// Handbook adds a handbook widget to show basic applet informations.
//
func Handbook(key *cftype.Key) {
	appletName := key.Value().String()

	widget := handbook.New(key.Log())
	widget.ShowVersion = true

	book := key.Source().Handbook(appletName)

	if widget == nil || book == nil {
		key.Log().NewErr("Handbook no widget")
		return
	}
	widget.SetPackage(book)
	key.BoxPage().PackStart(widget, true, true, 0)

	key.PackKeyWidget(key,
		func() interface{} { return appletName },
		func(uncast interface{}) {
			appletName = uncast.(string)
			book := key.Source().Handbook(appletName)
			widget.SetPackage(book)
		},
	)
}

// Frame adds a simple or expanded frame widget.
//
func Frame(key *cftype.Key) {
	if len(key.AuthorizedValues) == 0 {
		key.SetFrame(nil)
		key.SetFrameBox(nil)
		return
	}

	value, img := "", ""
	if key.AuthorizedValues[0] == "" {
		key.Log().Info("WidgetFrame, need value case 1")
		// value = g_key_file_get_string(pKeyFile, cGroupName, cKeyName, NULL) // utile ?
	} else {
		value = key.AuthorizedValues[0]
		if len(key.AuthorizedValues) > 1 {
			img = key.AuthorizedValues[1]
		}
	}

	// Create the frame label with the optional icon.
	label := newgtk.Label("")
	// key.SetLabel(label)
	label.SetMarkup(" " + common.Bold(key.Translate(value)) + " ")

	var labelContainer gtk.IWidget
	if img == "" {
		labelContainer = label
	} else {
		box := newgtk.Box(gtk.ORIENTATION_HORIZONTAL, cftype.MarginIcon/2)
		if icon, e := common.ImageNewFromFile(img, iconSizeFrame); !key.Log().Err(e, "Frame icon") { // TODO: fix size : int(gtk.ICON_SIZE_MENU)
			box.Add(icon)
		}
		box.Add(label)
		labelContainer = box
	}

	// Create the box that will contain next widgets (inside the frame).
	box := newgtk.Box(gtk.ORIENTATION_VERTICAL, cftype.MarginGUI)
	key.SetFrameBox(box)

	frame := newgtk.Frame("")
	key.SetFrame(frame)
	frame.SetBorderWidth(cftype.MarginGUI)
	frame.SetShadowType(gtk.SHADOW_OUT)
	frame.Add(box)

	// Set label and create the expander around the frame if needed.
	switch key.Type {
	case cftype.KeyFrame:
		frame.SetLabelWidget(labelContainer)
		key.BoxPage().PackStart(frame, false, false, 0)

	case cftype.KeyExpander:
		expand := newgtk.Expander("")
		expand.SetExpanded(false)
		expand.SetLabelWidget(labelContainer)

		expand.Add(frame)
		key.BoxPage().PackStart(expand, false, false, 0)
	}

	// SAME AS IN builder.go

	// 	if (pControlWidgets != NULL)
	// 	{
	// 		cd_debug ("ctrl\n");
	// 		CDControlWidget *cw = pControlWidgets->data;
	// 		if (cw->pControlContainer == key.Box)
	// 		{
	// 			cd_debug ("ctrl (NbControlled:%d, iFirstSensitiveWidget:%d, iNbSensitiveWidgets:%d)", cw->NbControlled, cw->iFirstSensitiveWidget, cw->iNbSensitiveWidgets);
	// 			cw->NbControlled --;
	// 			if (cw->iFirstSensitiveWidget > 0)
	// 				cw->iFirstSensitiveWidget --;
	// 			cw->iNonSensitiveWidget --;

	// 			GtkWidget *w = pExternFrame;
	// 			if (cw->iFirstSensitiveWidget == 0 && cw->iNbSensitiveWidgets > 0 && cw->iNonSensitiveWidget != 0)
	// 			{
	// 				cd_debug (" => sensitive\n");
	// 				cw->iNbSensitiveWidgets --;
	// 				if (GTK_IS_EXPANDER (w))
	// 					gtk_expander_set_expanded (GTK_EXPANDER (w), TRUE);
	// 			}
	// 			else
	// 			{
	// 				cd_debug (" => unsensitive\n");
	// 				if (!GTK_IS_EXPANDER (w))
	// 					gtk_widget_set_sensitive (w, FALSE);
	// 			}
	// 			if (cw->iFirstSensitiveWidget == 0 && cw->NbControlled == 0)
	// 			{
	// 				pControlWidgets = g_list_delete_link (pControlWidgets, pControlWidgets);
	// 				g_free (cw);
	// 			}
	// 		}
	// 	}
}

// Separator adds a simple horizontal separator.
//
func Separator(key *cftype.Key) {
	// GtkWidget *pAlign = gtk_alignment_new (.5, .5, 0.8, 1.);
	// g_object_set (pAlign, "height-request", 12, NULL);
	widget := newgtk.Separator(gtk.ORIENTATION_HORIZONTAL)
	// gtk_container_add (GTK_CONTAINER (pAlign), pOneWidget);
	key.PackWidget(widget, false, false, 0)
}

// Text just enlarges the widget label (nothing more intended).
//
func Text(key *cftype.Key) {
	key.Label().SetLineWrap(true)
	key.Label().SetJustify(gtk.JUSTIFY_FILL)
}
