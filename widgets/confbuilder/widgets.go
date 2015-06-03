package confbuilder

// // #cgo pkg-config: gtk+-3.0
// // #include <gtk/gtk.h>
// import "C"

import (
	"github.com/conformal/gotk3/gdk"
	"github.com/conformal/gotk3/glib"
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/godock/libs/log"
	"github.com/sqp/godock/libs/tran"

	"github.com/sqp/godock/widgets/common"
	"github.com/sqp/godock/widgets/confbuilder/datatype"
	"github.com/sqp/godock/widgets/gtk/gunvalue"

	"strconv"
)

//------------------------------------------------------[ WIDGETS COLLECTION ]--

// WidgetCheckButton adds a check button widget.
//
func (build *Builder) WidgetCheckButton(key *Key) {
	nbval, values, _ := build.Conf.GetBooleanList(key.Group, key.Name)
	for k := 0; k < key.NbElements; k++ {
		var value bool
		if uint64(k) < nbval {
			value = values[k]
		}
		widget, _ := gtk.CheckButtonNew()
		widget.SetActive(value)

		if key.Type == WidgetCheckControlButton {
			// 		_allocate_new_buffer;
			// 		data[0] = pKeyBox;
			// 		data[1] = (pFrameVBox != NULL ? pFrameVBox : pGroupBox);
			// 		if (pAuthorizedValuesList != NULL && pAuthorizedValuesList[0] != NULL)
			// 			iNbControlledWidgets = g_ascii_strtod (pAuthorizedValuesList[0], NULL);
			// 		else
			// 			iNbControlledWidgets = 1;
			// 		data[2] = GINT_TO_POINTER (iNbControlledWidgets);
			// 		if (iNbControlledWidgets < 0)  // a negative value means that the behavior is inverted.
			// 		{
			// 			bValue = !bValue;
			// 			iNbControlledWidgets = -iNbControlledWidgets;
			// 		}
			// 		g_signal_connect (G_OBJECT (pOneWidget), "toggled", G_CALLBACK(_cairo_dock_toggle_control_button), data);

			// 		g_object_set_data (G_OBJECT (pKeyBox), "nb-ctrl-widgets", GINT_TO_POINTER (iNbControlledWidgets));
			// 		g_object_set_data (G_OBJECT (pKeyBox), "one-widget", pOneWidget);

			// 		if (! bValue)  // les widgets suivants seront inactifs.
			// 		{
			// 			CDControlWidget *cw = g_new0 (CDControlWidget, 1);
			// 			pControlWidgets = g_list_prepend (pControlWidgets, cw);
			// 			cw->iNbSensitiveWidgets = 0;
			// 			cw->iNbControlledWidgets = iNbControlledWidgets;
			// 			cw->iFirstSensitiveWidget = 1;
			// 			cw->pControlContainer = (pFrameVBox != NULL ? pFrameVBox : pGroupBox);
			// 		}  // sinon le widget suivant est sensitif, donc rien a faire.
		}

		getValue := func() interface{} { return widget.GetActive() }
		build.addKeyWidget(widget, key, getValue)
	}
}

// WidgetInteger adds an integer selector widget.
//
func (build *Builder) WidgetInteger(key *Key) {
	var fMinValue float64
	var fMaxValue float64 = 9999
	if len(key.AuthorizedValues) > 0 {
		fMinValue, _ = strconv.ParseFloat(key.AuthorizedValues[0], 32)
	}
	if len(key.AuthorizedValues) > 1 {
		fMaxValue, _ = strconv.ParseFloat(key.AuthorizedValues[1], 32)
	}

	var toggle *gtk.ToggleButton
	if key.Type == WidgetIntegerSize {
		key.NbElements *= 2
		toggle, _ = gtk.ToggleButtonNew()
		img, _ := gtk.ImageNewFromIconName("media-playback-pause", gtk.ICON_SIZE_MENU) // get better image.
		toggle.SetImage(img)
	}

	var value, prevValue int
	var prevWidget *gtk.SpinButton

	nbval, values, _ := build.Conf.GetIntegerList(key.Group, key.Name)
	for k := 0; k < key.NbElements; k++ {
		if uint64(k) < nbval {
			value = values[k]
		}

		switch key.Type {
		case WidgetIntegerScale:
			// log.DEV("FLOAT INT", value, key.AuthorizedValues, fMinValue, fMaxValue)
			step := (fMaxValue - fMinValue) / 20
			if step < 1 {
				step = 1
			}
			adjustment, _ := gtk.AdjustmentNew(float64(value), fMinValue, fMaxValue, 1, step, 0)
			widget, _ := gtk.ScaleNew(gtk.ORIENTATION_HORIZONTAL, adjustment)
			widget.Set("digits", 0)
			getValue := func() interface{} { return int(widget.GetValue()) }
			build.addKeyScale(widget, key, getValue)

		case WidgetIntegerSpin, WidgetIntegerSize:
			widget, _ := gtk.SpinButtonNewWithRange(fMinValue, fMaxValue, 1)
			widget.SetValue(float64(value))
			getValue := func() interface{} { return widget.GetValueAsInt() }

			build.addKeyWidget(widget, key, getValue)

			if key.Type == WidgetIntegerSize {
				if k&1 == 0 { // separator
					label, _ := gtk.LabelNew("x")
					build.addSubWidget(label)
				} else { // connect both spin values.
					if prevValue == value {
						toggle.SetActive(true)
					}

					widget.Connect("value-changed", onValuePairChanged, &valuePair{
						// updated: widget,
						linked: prevWidget,
						toggle: toggle})
					prevWidget.Connect("value-changed", onValuePairChanged, &valuePair{
						// updated: prevWidget,
						linked: widget,
						toggle: toggle})
				}
				prevWidget = widget
				prevValue = value
			}
		}
	}

	if key.Type == WidgetIntegerSize {
		build.addSubWidget(toggle)
	}
	// bAddBackButton = TRUE;
}

// WidgetFloat adds a float selector widget. SpinButton or Horizontal Scale
//
func (build *Builder) WidgetFloat(key *Key) {
	var fMinValue float64
	var fMaxValue float64 = 9999
	if len(key.AuthorizedValues) > 0 {
		fMinValue, _ = strconv.ParseFloat(key.AuthorizedValues[0], 32)
	}
	if len(key.AuthorizedValues) > 1 {
		fMaxValue, _ = strconv.ParseFloat(key.AuthorizedValues[1], 32)
	}

	nbval, values, _ := build.Conf.GetDoubleList(key.Group, key.Name)
	for k := 0; k < key.NbElements; k++ {
		var value float64
		if uint64(k) < nbval {
			value = values[k]
		}

		switch key.Type {
		case WidgetFloatScale:
			// log.DEV("FLOAT SCALE", value, key.AuthorizedValues)
			adjustment, _ := gtk.AdjustmentNew(value, fMinValue, fMaxValue, (fMaxValue-fMinValue)/20, (fMaxValue-fMinValue)/10, 0)
			widget, _ := gtk.ScaleNew(gtk.ORIENTATION_HORIZONTAL, adjustment)
			widget.Set("digits", 3)
			getValue := func() interface{} { return widget.GetValue() }
			build.addKeyScale(widget, key, getValue)

		case WidgetFloatSpin:
			widget, _ := gtk.SpinButtonNewWithRange(fMinValue, fMaxValue, 1)
			widget.Set("digits", 3)
			widget.SetValue(value)
			getValue := func() interface{} { return widget.GetValue() }
			build.addKeyWidget(widget, key, getValue)
		}

		// bAddBackButton = TRUE,
	}

}

// WidgetColorSelector adds a color selector widget.
//
func (build *Builder) WidgetColorSelector(key *Key) {
	switch key.Type {
	case WidgetColorSelectorRGB:
		key.NbElements = 3
	case WidgetColorSelectorRGBA:
		key.NbElements = 4
	}
	_, values, _ := build.Conf.GetDoubleList(key.Group, key.Name)
	gdkColor := gdk.NewRGBA(values...)

	// test if we need
	// 	if nbval > 3 && key.Type == WidgetColorSelectorRGBA {
	// 	} else {
	// 	 gdkColor.alpha = C.gdouble(1)
	// 	}

	widget, _ := gtk.ColorButtonNewWithRGBA(gdkColor)
	widget.Set("use-alpha", key.Type == WidgetColorSelectorRGBA)
	getValue := func() interface{} { return widget.GetRGBA() }
	build.addKeyWidget(widget, key, getValue)
	// bAddBackButton = TRUE,
}

// WidgetListTheme adds an theme list widget.
//
func (build *Builder) WidgetListTheme(key *Key) {
	current, _ := build.Conf.GetString(key.Group, key.Name)
	log.Info("theme", current, key.AuthorizedValues)

	model, _ := newModelSimple()
	model.SetSortColumnId(RowName, gtk.SORT_ASCENDING)
	combo, getValue := build.newComboBoxWithModel(model, true, false, true)

	details := NewHandbook()
	build.addKeyWidget(combo, key, getValue)
	build.addWidget(details, false, false, 0)

	// Connect the theme preview update on selection.
	var list map[string]datatype.Handbooker
	combo.Connect("changed", func() {
		pack, ok := list[getValue().(string)]
		if ok {
			details.SetPackage(pack)
		}
	})

	// Fill the list with known themes.
	if len(key.AuthorizedValues) > 2 {
		current, _ := build.Conf.GetString(key.Group, key.Name)
		if key.AuthorizedValues[1] == "gauges" {
			list = build.data.ListThemeXML(key.AuthorizedValues[0], key.AuthorizedValues[1], key.AuthorizedValues[2])
			iter := fillModelWithTheme(model, list, current)
			combo.SetActiveIter(iter)

		} else {
			list = build.data.ListThemeINI(key.AuthorizedValues[0], key.AuthorizedValues[1], key.AuthorizedValues[2])
			iter := fillModelWithTheme(model, list, current)
			combo.SetActiveIter(iter)
		}
	}

	// //\______________ On construit le widget de visualisation de themes.
	// modele = _cairo_dock_gui_allocate_new_model ();
	// gtk_tree_sortable_set_sort_column_id (GTK_TREE_SORTABLE (modele), CAIRO_DOCK_MODEL_NAME, GTK_SORT_ASCENDING);

	// _add_combo_from_modele (modele, TRUE, FALSE, TRUE);

	// add the state icon.
	// 	rend = gtk_cell_renderer_pixbuf_new ();
	// 	gtk_cell_layout_pack_start (GTK_CELL_LAYOUT (pOneWidget), rend, FALSE);
	// 	gtk_cell_layout_set_attributes (GTK_CELL_LAYOUT (pOneWidget), rend, "pixbuf", CAIRO_DOCK_MODEL_ICON, NULL);
	// 	gtk_cell_layout_reorder (GTK_CELL_LAYOUT (pOneWidget), rend, 0);

	// //\______________ On recupere les themes.
	// if (pAuthorizedValuesList != NULL)
	// {
	// 	// get the local, shared and distant paths.
	// 	gchar *cShareThemesDir = NULL, *cUserThemesDir = NULL, *cDistantThemesDir = NULL, *cHint = NULL;
	// 	if (pAuthorizedValuesList[0] != NULL)
	// 	{
	// 		cShareThemesDir = (*pAuthorizedValuesList[0] != '\0' ? cairo_dock_search_image_s_path (pAuthorizedValuesList[0]) : NULL);  // on autorise les ~/blabla.
	// 		if (pAuthorizedValuesList[1] != NULL)
	// 		{
	// 			cUserThemesDir = g_strdup_printf ("%s/%s", g_cExtrasDirPath, pAuthorizedValuesList[1]);
	// 			if (pAuthorizedValuesList[2] != NULL)
	// 			{
	// 				cDistantThemesDir = (*pAuthorizedValuesList[2] != '\0' ? pAuthorizedValuesList[2] : NULL);
	// 				cHint = pAuthorizedValuesList[3];  // NULL to not filter.
	// 			}
	// 		}
	// 	}

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

// WidgetIconThemeList adds a desktop icon-themes list widget.
//
func (build *Builder) WidgetIconThemeList(key *Key) {
	list := append([]datatype.Field{
		{},
		{Key: "_Custom Icons_", Name: tran.Slate("_Custom Icons_")}},
		build.data.ListIconTheme()...)
	build.newComboBoxFields(key, list)
}

// WidgetViewList adds a view list widget.
//
func (build *Builder) WidgetViewList(key *Key) {
	list := append([]datatype.Field{{Key: "", Name: "default"}}, build.data.ListViews()...)
	build.newComboBoxFields(key, list)

	// RowDesc: "none"}) // (pRenderer != NULL ? pRenderer->cReadmeFilePath : "none")
	// 		CAIRO_DOCK_MODEL_IMAGE, (pRenderer != NULL ? pRenderer->cPreviewFilePath : "none")

	// 	gtk_tree_sortable_set_sort_column_id (GTK_TREE_SORTABLE (pListStore), CAIRO_DOCK_MODEL_NAME, GTK_SORT_ASCENDING);
}

// WidgetAnimationList adds an animation list widget.
//
func (build *Builder) WidgetAnimationList(key *Key) {
	list := append([]datatype.Field{{}}, build.data.ListAnimations()...)
	build.newComboBoxFields(key, list)
}

// WidgetDialogDecoratorList adds an dialog decorator list widget.
//
func (build *Builder) WidgetDialogDecoratorList(key *Key) {
	list := build.data.ListDialogDecorator()
	build.newComboBoxFields(key, list)
}

// WidgetListDeskletDecoration adds a desklet decoration list widget.
//
func (build *Builder) WidgetListDeskletDecoration(key *Key) {
	list := build.data.ListDeskletDecorations()
	if key.Type == WidgetDeskletDecorationListWithDefault {
		list = append([]datatype.Field{{Key: "default", Name: "default"}}, list...) // prepend default.
	}
	build.newComboBoxFields(key, list)
	// 	gtk_tree_sortable_set_sort_column_id (GTK_TREE_SORTABLE (pListStore), CAIRO_DOCK_MODEL_NAME, GTK_SORT_ASCENDING);

	// _allocate_new_buffer;
	// data[0] = pKeyBox;
	// data[1] = (pFrameVBox != NULL ? pFrameVBox : pGroupBox);
	// iNbControlledWidgets = 9;
	// data[2] = GINT_TO_POINTER (iNbControlledWidgets);
	// iNbControlledWidgets --;  // car dans cette fonction, on ne compte pas le separateur.
	// g_signal_connect (G_OBJECT (pOneWidget), "changed", G_CALLBACK (_cairo_dock_select_custom_item_in_combo), data);

	current, _ := build.Conf.GetString(key.Group, key.Name)
	if current == "personnal" { // Disable the next widgets.
		// 		CDControlWidget *cw = g_new0 (CDControlWidget, 1);
		// 		pControlWidgets = g_list_prepend (pControlWidgets, cw);
		// 		cw->iNbControlledWidgets = iNbControlledWidgets;
		// 		cw->iNbSensitiveWidgets = 0;
		// 		cw->iFirstSensitiveWidget = 1;
		// 		cw->pControlContainer = (pFrameVBox != NULL ? pFrameVBox : pGroupBox);
	}
}

// WidgetScreensList adds a screen selection widget.
//
func (build *Builder) WidgetScreensList(key *Key) {
	list := build.data.ListScreens()
	combo := build.newComboBoxFields(key, list)
	if len(list) <= 1 {
		combo.SetSensitive(false)
	}

	// 	gldi_object_register_notification (&myDesktopMgr,
	// 		NOTIFICATION_DESKTOP_GEOMETRY_CHANGED,
	// 		(GldiNotificationFunc) _on_screen_modified,
	// 		GLDI_RUN_AFTER, pScreensListStore);
	// 	g_signal_connect (pOneWidget, "destroy", G_CALLBACK (_on_list_destroyed), NULL);
}

// WidgetDockList adds a dock list widget.
//
func (build *Builder) WidgetDockList(key *Key) {
	// Get current Icon name if its a Subdock.
	iIconType, _ := build.Conf.GetInteger(key.Group, "Icon Type")
	SubdockName := ""
	if iIconType == UserIconStack { // it's a stack-icon
		SubdockName, _ = build.Conf.GetString(key.Group, "Name") // It's a subdock, get its name to remove the selection of a recursive position (inside itself).
	}

	list := build.data.ListDocks("", SubdockName)                                 // Get the list of available docks. Keep parent, but remove itself from the list.
	list = append(list, datatype.Field{Key: "_New Dock_", Name: "New main dock"}) // append create new.

	model, _ := newModelSimple()
	current, _ := build.Conf.GetString(key.Group, key.Name)

	model.SetSortColumnId(RowName, gtk.SORT_ASCENDING)

	iter := fillModelWithFields(model, list, current)
	combo, _ := gtk.ComboBoxNewWithModel(model)
	renderer, _ := gtk.CellRendererTextNew()
	combo.PackStart(renderer, false)
	combo.AddAttribute(renderer, "text", RowName)
	combo.SetActiveIter(iter)

	getValue := func() interface{} {
		iter, _ := combo.GetActiveIter()
		text := GetActiveRowInCombo(model, iter)
		// text := combo.GetActive();
		return text
	}
	build.addKeyWidget(combo, key, getValue)
}

// WidgetIconsList adds an icon list widget.
//
func (build *Builder) WidgetIconsList(key *Key) {
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

	current, _ := build.Conf.GetString(key.Group, key.Name)

	model, _ := newModelSimple()
	widget, _ := gtk.ComboBoxNewWithModel(model)

	rp, _ := gtk.CellRendererPixbufNew()
	widget.PackStart(rp, false)
	widget.AddAttribute(rp, "pixbuf", RowIcon)

	renderer, _ := gtk.CellRendererTextNew()
	widget.PackStart(renderer, true)
	widget.AddAttribute(renderer, "text", RowName)

	widget.Set("id-column", RowName)
	getValue := func() interface{} {
		iter, _ := widget.GetActiveIter()
		text := GetActiveRowInCombo(model, iter)
		return text
	}

	build.addKeyWidget(widget, key, getValue)

	iconSize := 24
	// iconSize := int(gtk.ICON_SIZE_LARGE_TOOLBAR)

	for _, icon := range build.data.ListIconsMainDock() {
		configPath := icon.ConfigPath()
		iter := model.Append()
		name, img := icon.DefaultNameIcon()

		model.SetCols(iter, gtk.Cols{
			RowName: name,
			RowKey:  configPath,
		})

		if img != "" {
			if pix, e := common.PixbufNewFromFile(img, iconSize); !log.Err(e, "Load icon") {
				model.SetValue(iter, RowIcon, pix)
			}
		}

		if configPath == current {
			widget.SetActiveIter(iter)
		}
	}

	//

	//

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

}

// WidgetJumpToModule adds a redirect button widget.
// USED?
//
func (build *Builder) WidgetJumpToModule(key *Key) {
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

// WidgetLaunchCommand adds a launch command widget.
// HELP ONLY
//
func (build *Builder) WidgetLaunchCommand(key *Key) {
	// if (pAuthorizedValuesList == NULL || pAuthorizedValuesList[0] == NULL || *pAuthorizedValuesList[0] == '\0')
	// 	break ;
	// gchar *cFirstCommand = NULL;
	// cFirstCommand = pAuthorizedValuesList[0];
	// if (iElementType == CAIRO_DOCK_WidgetLaunchCommandIfCondition)
	// {
	// 	if (pAuthorizedValuesList[1] == NULL)
	// 	{ // condition without condition...
	// 		gtk_widget_set_sensitive (pLabel, FALSE);
	// 		break ;
	// 	}
	// 	gchar *cSecondCommand = pAuthorizedValuesList[1];
	// 	gchar *cResult = cairo_dock_launch_command_sync (cSecondCommand);
	// 	cd_debug ("%s: %s => %s", __func__, cSecondCommand, cResult);
	// 	if (cResult == NULL || *cResult == '0' || *cResult == '\0')  // result is 'fail'
	// 	{
	// 		gtk_widget_set_sensitive (pLabel, FALSE);
	// 		g_free (cResult);
	// 		break ;
	// 	}
	// 	g_free (cResult);
	// }
	// pOneWidget = gtk_button_new_from_stock (GTK_STOCK_JUMP_TO);
	// g_signal_connect (G_OBJECT (pOneWidget),
	// 	"clicked",
	// 	G_CALLBACK (_cairo_dock_widget_launch_command),
	// 	g_strdup (cFirstCommand));
	// _pack_subwidget (pOneWidget);
}

// WidgetLists adds a string list widget.
//
func (build *Builder) WidgetLists(key *Key) {
	if (key.Type == WidgetNumberedControlListSimple || key.Type == WidgetNumberedControlListSelective) && len(key.AuthorizedValues) == 0 {
		return
	}

	// Get full value with ';'.
	value, _ := build.Conf.GetString(key.Group, key.Name)

	// log.DEV("LIST "+string(key.Type), key.Name, value, key.AuthorizedValues)

	bNumberedList := (key.Type == WidgetNumberedList ||
		key.Type == WidgetNumberedControlListSimple ||
		key.Type == WidgetNumberedControlListSelective)

	// on construit la combo.
	widget, model, getValue := newComboBox(key.Type == WidgetListWithEntry, bNumberedList)

	// gtk_tree_sortable_set_sort_column_id (GTK_TREE_SORTABLE (modele), GTK_TREE_SORTABLE_UNSORTED_SORT_COLUMN_ID, GTK_SORT_ASCENDING);

	// int iNonSensitiveWidget = 0;

	if len(key.AuthorizedValues) > 0 {
		// 	k = 0;
		iSelectedItem := -1
		// int iOrder1, iOrder2, iExcept;
		if bNumberedList {
			iSelectedItem, _ = strconv.Atoi(value)
		}

		dk := 1
		if key.Type == WidgetNumberedControlListSelective {
			dk = 3
		}

		if key.Type == WidgetNumberedControlListSimple || key.Type == WidgetNumberedControlListSelective {
			build.iNbControlledWidgets = 0
		}

		// 	gchar *cResult = (bNumberedList ? g_new0 (gchar , 10) : NULL);
		k := 0
		for ; k < len(key.AuthorizedValues); k += dk { // on ajoute toutes les chaines possibles a la combo.
			if iSelectedItem == -1 && value == key.AuthorizedValues[k] {
				iSelectedItem = k / dk
			}

			// 		if (cResult != NULL)
			// 			snprintf (cResult, 9, "%d", k/dk);

			// 		iExcept = 0;
			// 		if key.Type == WidgetNumberedControlListSelective 		{
			// 			iOrder1 = atoi (key.AuthorizedValues[k+1]);
			// 			gchar *str = strchr (key.AuthorizedValues[k+2], ',');
			// 			if (str)  // Note: this mechanism is an addition to the original {first widget, number of widgets}; it's not very generic nor beautiful, but until we need more, it's well enough (currently, only the Dock background needs it).
			// 			{
			// 				*str = '\0';
			// 				iExcept = atoi (str+1);
			// 			}
			// 			iOrder2 = atoi (key.AuthorizedValues[k+2]);
			// 			iNbControlledWidgets = MAX (iNbControlledWidgets, iOrder1 + iOrder2 - 1);
			// 			//g_print ("iSelectedItem:%d ; k/dk:%d\n", iSelectedItem , k/dk);
			// 			if (iSelectedItem == (int)k/dk)
			// 			{
			// 				iFirstSensitiveWidget = iOrder1;
			// 				iNbSensitiveWidgets = iOrder2;
			// 				iNonSensitiveWidget = iExcept;
			// 				if (iNonSensitiveWidget != 0)
			// 					iNbControlledWidgets ++;
			// 			}
			// 		}					else					{
			// 			iOrder1 = iOrder2 = k;
			// 		}

			//
			name := ""
			if key.Type == WidgetListWithEntry {
				name = key.AuthorizedValues[k]
			} else {
				name = build.translate(key.AuthorizedValues[k])
			}

			result := ""
			if result == "" {
				result = key.AuthorizedValues[k]
			}
			// log.DEV(name, ":", result)
			model.SetCols(model.Append(), gtk.Cols{
				RowKey:  name,
				RowName: result})
			// 			CAIRO_DOCK_MODEL_ORDER, iOrder1,
			// 			CAIRO_DOCK_MODEL_ORDER2, iOrder2,
			// 			CAIRO_DOCK_MODEL_STATE, iExcept, -1);
		}

		if key.Type == WidgetListWithEntry { // Set text directly if it's an entry.
			if iSelectedItem == -1 {
				entry, _ := widget.GetChild()
				toEntry(entry).SetText(value)
			} else {
				widget.SetActive(iSelectedItem)
			}

		} else { // Select current
			if iSelectedItem == -1 { // Current choice wasn't in the list. Select first.
				iSelectedItem = 0
			}
			if k > 0 { // Check we have something to select.
				widget.SetActive(iSelectedItem)
			}
		}
		if key.Type == WidgetNumberedControlListSimple || key.Type == WidgetNumberedControlListSelective {
			// 		_allocate_new_buffer;
			// 		data[0] = pKeyBox;
			// 		data[1] = (pFrameVBox != NULL ? pFrameVBox : pGroupBox);
			if key.Type == WidgetNumberedControlListSimple {
				// 			iNbControlledWidgets = k;
				// 			data[2] = GINT_TO_POINTER (iNbControlledWidgets);
				// 			g_signal_connect (G_OBJECT (pOneWidget), "changed", G_CALLBACK (_cairo_dock_select_one_item_in_control_combo), data);
				// 			iFirstSensitiveWidget = iSelectedItem+1;  // on decroit jusqu'a 0.
				// 			iNbSensitiveWidgets = 1;
				// 			//g_print ("CONTROL : %d,%d,%d\n", iNbControlledWidgets, iFirstSensitiveWidget, iNbSensitiveWidgets);
			} else {
				// 			data[2] = GINT_TO_POINTER (iNbControlledWidgets);
				// 			g_signal_connect (G_OBJECT (pOneWidget), "changed", G_CALLBACK (_cairo_dock_select_one_item_in_control_combo_selective), data);
				// 			//g_print ("CONTROL : %d,%d,%d\n", iNbControlledWidgets, iFirstSensitiveWidget, iNbSensitiveWidgets);
			}
			// 		g_object_set_data (G_OBJECT (pKeyBox), "nb-ctrl-widgets", GINT_TO_POINTER (iNbControlledWidgets));
			// 		g_object_set_data (G_OBJECT (pKeyBox), "one-widget", pOneWidget);
			// 		CDControlWidget *cw = g_new0 (CDControlWidget, 1);
			// 		pControlWidgets = g_list_prepend (pControlWidgets, cw);
			// 		cw->pControlContainer = (pFrameVBox != NULL ? pFrameVBox : pGroupBox);
			// 		cw->iNbControlledWidgets = iNbControlledWidgets;
			// 		cw->iFirstSensitiveWidget = iFirstSensitiveWidget;
			// 		cw->iNbSensitiveWidgets = iNbSensitiveWidgets;
			// 		cw->iNonSensitiveWidget = iNonSensitiveWidget;
			// 		//g_print (" pControlContainer:%x\n", pControlContainer);
		}
	}
	build.addKeyWidget(widget, key, getValue)
}

// WidgetTreeView adds a treeview widget.
//
func (build *Builder) WidgetTreeView(key *Key) {

	// value, _ := build.Conf.GetString(key.Group, key.Name)
	_, values, e := build.Conf.GetStringList(key.Group, key.Name)
	log.Err(e, "WidgetTreeView conf.GetStringList")

	// Build treeview.
	model, _ := gtk.ListStoreNew(
		glib.TYPE_STRING,  /* RowKey*/
		glib.TYPE_STRING,  /* RowName*/
		glib.TYPE_STRING,  /* RowIcon*/
		glib.TYPE_STRING,  /* RowDesc*/
		glib.TYPE_BOOLEAN, // active
		glib.TYPE_INT)     // order

	widget, _ := gtk.TreeViewNewWithModel(model)
	widget.Set("headers-visible", false)

	getValue := func() interface{} { // Grab data from all iters.
		var list []string
		iter, ok := model.GetIterFirst()
		for ; ok; ok = model.IterNext(iter) {
			str, e := gunvalue.New(model.GetValue(iter, RowName)).String()
			if !log.Err(e, "WidgetTreeView GetValue "+key.Name) {
				list = append(list, str)
			}
		}
		return list
	}

	// Add control buttons.
	if key.Type == WidgetTreeViewMultiChoice {
		renderer, _ := gtk.CellRendererToggleNew()
		col, _ := gtk.TreeViewColumnNewWithAttribute("", renderer, "active", 4)
		widget.AppendColumn(col)
		// 	g_signal_connect (G_OBJECT (rend), "toggled", (GCallback) _cairo_dock_activate_one_element, modele);

		renderer.Set("active", 4)
	}

	renderer, _ := gtk.CellRendererTextNew()
	col, _ := gtk.TreeViewColumnNewWithAttribute("", renderer, "text", RowName)
	widget.AppendColumn(col)

	// cValueList = g_key_file_get_string_list (pKeyFile, cGroupName, cKeyName, &length, NULL);

	model.SetSortColumnId(5, gtk.SORT_ASCENDING)

	scroll, _ := gtk.ScrolledWindowNew(nil, nil)

	//

	if len(key.AuthorizedValues) > 0 {
		log.DEV("WidgetTreeView AuthorizedValues", key.AuthorizedValues)
	}

	// if (key.AuthorizedValues != NULL && key.AuthorizedValues[0] != NULL)
	// 	for (k = 0; key.AuthorizedValues[k] != NULL; k++);
	// else
	// 	k = 1;
	scroll.Set("height-request", 100) // key.Type == WidgetTreeViewSortAndModify ? 100 : MIN (100, k * 25)
	scroll.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_AUTOMATIC)

	scroll.Add(widget)

	if key.Type == WidgetTreeViewSortSimple || key.Type == WidgetTreeViewSortAndModify {
		vboxItems, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, MarginGUI)

		buttonUp, _ := gtk.ButtonNew()
		buttonDn, _ := gtk.ButtonNew()
		imgUp, _ := gtk.ImageNewFromIconName("go-up", gtk.ICON_SIZE_SMALL_TOOLBAR)
		imgDn, _ := gtk.ImageNewFromIconName("go-down", gtk.ICON_SIZE_SMALL_TOOLBAR)

		buttonUp.SetImage(imgUp)
		buttonDn.SetImage(imgDn)

		data := treeViewData{model, widget, nil}

		buttonUp.Connect("clicked", onTreeviewMoveUp, data) // Move selection up and down callbacks.
		buttonDn.Connect("clicked", onTreeviewMoveDown, data)

		vboxItems.PackStart(buttonUp, false, false, 0)
		vboxItems.PackStart(buttonDn, false, false, 0)

		if key.Type == WidgetTreeViewSortAndModify {

			vboxAdd, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, MarginGUI)
			sep, _ := gtk.SeparatorNew(gtk.ORIENTATION_VERTICAL)
			buttonAdd, _ := gtk.ButtonNew()
			entry, _ := gtk.EntryNew()
			buttonRm, _ := gtk.ButtonNew()

			imgAdd, _ := gtk.ImageNewFromIconName("list-add", gtk.ICON_SIZE_SMALL_TOOLBAR)
			imgRm, _ := gtk.ImageNewFromIconName("list-remove", gtk.ICON_SIZE_SMALL_TOOLBAR)
			buttonAdd.SetImage(imgAdd)
			buttonRm.SetImage(imgRm)

			build.addSubWidget(vboxAdd)
			build.addSubWidget(sep)
			vboxAdd.PackStart(buttonAdd, false, false, 0)
			vboxAdd.PackStart(entry, false, false, 0)
			vboxItems.PackStart(buttonRm, false, false, 0)

			data.entry = entry
			buttonAdd.Connect("clicked", onTreeviewAddText, data)   // Add new iter to model with the value of the entry widget. Clear entry widget.
			buttonRm.Connect("clicked", onTreeviewRemoveText, data) // Remove selected iter from model. Set its value to the entry widget.
		}

		build.addSubWidget(vboxItems)
	}

	build.addKeyWidget(scroll, key, getValue)

	// Fill model with values.
	switch key.Type {
	case WidgetTreeViewSortAndModify: // add saved values.
		for i, val := range values {
			iter := model.Append()
			model.SetValue(iter, RowKey, val)
			model.SetValue(iter, RowName, val)

			model.SetValue(iter, 4, true) // active
			model.SetValue(iter, 5, i)    // order
		}

	case WidgetTreeViewSortSimple, WidgetTreeViewMultiChoice:
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

// WidgetFontSelector adds a font selector widget.
//
func (build *Builder) WidgetFontSelector(key *Key) {
	value, _ := build.Conf.GetString(key.Group, key.Name)
	widget, _ := gtk.FontButtonNewWithFont(value)
	widget.Set("show-style", true)
	widget.Set("show-size", true)
	widget.Set("use-size", true)
	widget.Set("use-font", true)

	getValue := func() interface{} { return widget.GetFontName() }
	build.addKeyWidget(widget, key, getValue)
}

// WidgetLink adds a link widget.
//
func (build *Builder) WidgetLink(key *Key) {
	link, _ := build.Conf.GetString(key.Group, key.Name)

	text := ""
	if len(key.AuthorizedValues) > 0 {
		text = key.AuthorizedValues[0]
	} else {
		text = tran.Slate("link")
	}
	widget, e := gtk.LinkButtonNewWithLabel(link, text)
	log.Err(e, "linkbutton")
	build.addSubWidget(widget)
}

// WidgetStrings adds a string widget. Many options included.
// TODO: need password cypher, G_USER_DIRECTORY_PICTURES, play sound.
//
func (build *Builder) WidgetStrings(key *Key) {
	value, _ := build.Conf.GetString(key.Group, key.Name)
	widget, _ := gtk.EntryNew()

	widget.SetText(value)

	if key.Type == WidgetPasswordEntry { // on cache le texte entre et on decrypte 'cValue'.
		widget.SetVisibility(false)
		// gchar *cDecryptedString = NULL;
		// cairo_dock_decrypt_string ( cValue, &cDecryptedString );
		// g_free (cValue);
		// cValue = cDecryptedString;
	}

	getValue := func() interface{} { text, _ := widget.GetText(); return text }
	build.addKeyWidget(widget, key, getValue)

	// 	Add special buttons to fill the entry box.
	switch key.Type {
	case WidgetFileSelector, WidgetFolderSelector, WidgetSoundSelector, WidgetImageSelector: // we add a file selector
		fileChooser, _ := gtk.ButtonNew()
		img, _ := gtk.ImageNewFromIconName("document-open", gtk.ICON_SIZE_SMALL_TOOLBAR)
		fileChooser.SetImage(img)
		fileChooser.Connect("clicked", onFileChooserOpen, fileChooserData{widget, key})

		build.addSubWidget(fileChooser)

		if key.Type == WidgetSoundSelector { //Sound Play Button
			play, _ := gtk.ButtonNew()
			imgPlay, _ := gtk.ImageNewFromIconName("media-playback-start", gtk.ICON_SIZE_SMALL_TOOLBAR)
			play.SetImage(imgPlay)

			// 			g_signal_connect (G_OBJECT (pButtonPlay),
			// 				"clicked",
			// 				G_CALLBACK (_cairo_dock_play_a_sound),
			// 				data);
			build.addSubWidget(play)
		}

	case WidgetShortkeySelector, WidgetClassSelector: // on ajoute un selecteur de touches/classe.
		grab, _ := gtk.ButtonNewWithLabel("Grab")
		// 		gtk_widget_add_events(pMainWindow, GDK_KEY_PRESS_MASK);

		switch key.Type {
		case WidgetClassSelector:
			grab.Connect("clicked", onClassGrabClicked)

		case WidgetShortkeySelector:
			grab.Connect("clicked", onKeyGrabClicked, &textGrabData{entry: widget, win: build.win})
		}
		build.addSubWidget(grab)

		// for _, sk := range build.data.ListShortkeys() {
		// 	if sk.GetConfFilePath() == build.Conf.File {
		// 		println("found file shortkey")
		// 	}
		// }

	}

	// Display a default value when empty.
	if len(key.AuthorizedValues) > 0 && key.AuthorizedValues[0] != "" {
		defaultText := build.translate(key.AuthorizedValues[0])
		if value == "" {
			key.IsDefault = true

			widget.SetText(defaultText)

			color := gdk.NewRGBA(DefaultTextColor, DefaultTextColor, DefaultTextColor, 1)
			widget.OverrideColor(gtk.STATE_FLAG_NORMAL, color)
		}
		cbChanged, _ := widget.Connect("changed", onTextDefaultChanged, key)
		data := textDefaultData{key: key, text: defaultText, cbID: cbChanged}
		widget.Connect("focus-in-event", onTextDefaultFocusIn, data)
		widget.Connect("focus-out-event", onTextDefaultFocusOut, data)

		// TODO: check other use of those fields.
		// 	 g_object_set_data (G_OBJECT (pEntry), "ignore-value", GINT_TO_POINTER (TRUE));
		// 	 g_object_set_data (G_OBJECT (pOneWidget), "default-text", cDefaultText);

	}
}

// WidgetHandbook adds a handbook widget to show basic applet informations.
//
func (build *Builder) WidgetHandbook(key *Key) {
	appletName, e := build.Conf.GetString(key.Group, key.Name)
	widget := NewHandbook()
	widget.ShowVersion = true

	book := build.data.Handbook(appletName)
	// pack := build.data.AppletPackage(appletName)

	if !log.Err(e, "WIDGET_HANDBOOK no key") {
		if widget != nil {
			if book != nil {
				widget.SetPackage(book)
				build.pageBox.PackStart(widget, true, true, 0)
			}
		}
	}
}

// WidgetFrame adds a simple or expanded frame widget.
//
func (build *Builder) WidgetFrame(key *Key) {
	if len(key.AuthorizedValues) == 0 {
		build.pFrame = nil
		build.pFrameVBox = nil
		return
	}

	value, img := "", ""
	if key.AuthorizedValues[0] == "" {
		log.DEV("WidgetFrame, need value case 1")
		// value = g_key_file_get_string(pKeyFile, cGroupName, cKeyName, NULL) // utile ?
	} else {
		value = key.AuthorizedValues[0]
		if len(key.AuthorizedValues) > 1 {
			img = key.AuthorizedValues[1]
		}
	}

	// Create the frame label with the optional icon.
	build.pLabel, _ = gtk.LabelNew("")
	build.pLabel.SetMarkup(" " + common.Bold(build.translate(value)) + " ")
	if img == "" {
		build.pLabelContainer = build.pLabel
	} else {
		box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, MarginIcon/2)
		if icon, e := common.ImageNewFromFile(img, 20); !log.Err(e, "Frame icon") { // TODO: fix size : int(gtk.ICON_SIZE_MENU)
			box.Add(icon)
		}
		box.Add(build.pLabel)
		build.pLabelContainer = box
	}

	// Create the box that will contain next widgets (inside the frame).
	build.pFrameVBox, _ = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, MarginGUI)

	build.pFrame, _ = gtk.FrameNew("")
	build.pFrame.SetBorderWidth(MarginGUI)
	build.pFrame.SetShadowType(gtk.SHADOW_OUT)
	build.pFrame.Add(build.pFrameVBox)

	// Set label and create the expander around the frame if needed.
	switch key.Type {
	case WidgetFrame:
		build.pFrame.SetLabelWidget(build.pLabelContainer)
		build.pageBox.PackStart(build.pFrame, false, false, 0)

	case WidgetExpander:
		expand, _ := gtk.ExpanderNew("")
		expand.SetExpanded(false)
		expand.SetLabelWidget(build.pLabelContainer)

		expand.Add(build.pFrame)
		build.pageBox.PackStart(expand, false, false, 0)
	}

	// SAME AS IN builder.go

	// 	if (pControlWidgets != NULL)
	// 	{
	// 		cd_debug ("ctrl\n");
	// 		CDControlWidget *cw = pControlWidgets->data;
	// 		if (cw->pControlContainer == build.Box)
	// 		{
	// 			cd_debug ("ctrl (iNbControlledWidgets:%d, iFirstSensitiveWidget:%d, iNbSensitiveWidgets:%d)", cw->iNbControlledWidgets, cw->iFirstSensitiveWidget, cw->iNbSensitiveWidgets);
	// 			cw->iNbControlledWidgets --;
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
	// 			if (cw->iFirstSensitiveWidget == 0 && cw->iNbControlledWidgets == 0)
	// 			{
	// 				pControlWidgets = g_list_delete_link (pControlWidgets, pControlWidgets);
	// 				g_free (cw);
	// 			}
	// 		}
	// 	}
}

// WidgetSeparator adds a simple horizontal separator.
//
func (build *Builder) WidgetSeparator() {
	// GtkWidget *pAlign = gtk_alignment_new (.5, .5, 0.8, 1.);
	// g_object_set (pAlign, "height-request", 12, NULL);
	widget, _ := gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)
	// gtk_container_add (GTK_CONTAINER (pAlign), pOneWidget);
	build.addWidget(widget, false, false, 0)
}
