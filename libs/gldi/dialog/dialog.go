// Package dialog provides access to gldi dialogs.
package dialog

// #cgo pkg-config: gldi
// #include <stdlib.h>                              // free
// #include "cairo-dock-backends-manager.h"         // gldi_dialog_new
/*

extern void onDialogAnswer    (int iClickedButton, GtkWidget *pInteractiveWidget, gpointer data, CairoDialog *pDialog);
extern void onDialogDestroyed (gpointer p);


static gpointer intToPointer(int i) {
	return ((gpointer) (glong) (i));
}


static const gchar** constListString(gchar** split) {
	return (const gchar **)split;
}

*/
import "C"

import (
	"github.com/conformal/gotk3/glib"
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/godock/libs/cdtype" // Applets types.
	"github.com/sqp/godock/libs/gldi"

	"unsafe"
)

//
//-----------------------------------------------------------------[ DIALOGS ]--

// dialogPtr is the reference to the current dialog. Only one is allowed.
var dialogPtr *Dialog

var dialogCall func(int, *gtk.Widget)

func removeDialog() {
	if dialogPtr != nil { // Only one dialog allowed. Triggered when opening a dialog on another icon.
		gldi.ObjectUnref(dialogPtr)
		dialogPtr = nil
	}
}

type Dialog struct {
	Ptr *C.CairoDialog
}

func NewDialogFromNative(p unsafe.Pointer) *Dialog {
	if p == nil {
		return nil
	}
	return &Dialog{(*C.CairoDialog)(p)}
}

func (o *Dialog) ToNative() unsafe.Pointer {
	return unsafe.Pointer(o.Ptr)
}

//
//----------------------------------------------------------[ COMMON DIALOGS ]--

func DialogShowGeneralMessage(str string, duration float64) {
	cstr := (*C.gchar)(C.CString(str))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	C.gldi_dialog_show_general_message(cstr, C.double(duration))
}

func DialogShowTemporaryWithIcon(str string, icon *gldi.Icon, container *gldi.Container, duration float64, iconPath string) {
	cstr := (*C.gchar)(C.CString(str))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	cpath := (*C.gchar)(C.CString(iconPath))
	defer C.free(unsafe.Pointer((*C.char)(cpath)))

	cicon := (*C.Icon)(unsafe.Pointer(icon.Ptr))
	ccontainer := (*C.GldiContainer)(unsafe.Pointer(container.Ptr))

	C.gldi_dialog_show_temporary_with_icon(cstr, cicon, ccontainer, C.double(duration), cpath)
}

func DialogShowWithQuestion(str string, icon *gldi.Icon, container *gldi.Container, iconPath string, onAnswer func(int, *gtk.Widget)) {
	cstr := (*C.gchar)(C.CString(str))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	cpath := (*C.gchar)(C.CString(iconPath))
	defer C.free(unsafe.Pointer((*C.char)(cpath)))

	cicon := (*C.Icon)(unsafe.Pointer(icon.Ptr))
	ccontainer := (*C.GldiContainer)(unsafe.Pointer(container.Ptr))

	dialogCall = onAnswer
	removeDialog()
	c := C.gldi_dialog_show_with_question(cstr, cicon, ccontainer, cpath, C.CairoDockActionOnAnswerFunc(C.onDialogAnswer), nil, nil)
	dialogPtr = NewDialogFromNative(unsafe.Pointer(c))
}

//
//-----------------------------------------------------------[ CUSTOM DIALOG ]--

// NewDialog creates a custom dialog

func NewDialog(icon *gldi.Icon, container *gldi.Container, dialog cdtype.DialogData) *Dialog {
	dialogCall = nil

	// Common dialog attributes.
	attr := new(C.CairoDialogAttr)

	attr.pIcon = (*C.Icon)(unsafe.Pointer(icon.Ptr))
	attr.pContainer = (*C.GldiContainer)(unsafe.Pointer(container.Ptr))

	if dialog.Icon != "" {
		// w,h :=
		// 		cairo_dock_get_icon_extent (pIcon, &w, &h);
		// 		cImageFilePath = cairo_dock_search_icon_s_path (g_value_get_string (v), MAX (w, h));
		attr.cImageFilePath = gchar(dialog.Icon)
	} else {
		attr.cImageFilePath = gchar("same icon")
	}

	if dialog.Message != "" {
		attr.cText = gchar(dialog.Message)
	}

	if dialog.Buttons != "" {
		cstr := gchar(dialog.Buttons)
		csep := gchar(";")
		clist := C.g_strsplit(cstr, csep, -1) // NULL-terminated
		C.free(unsafe.Pointer((*C.char)(cstr)))
		C.free(unsafe.Pointer((*C.char)(csep)))
		defer C.g_strfreev(clist)
		attr.cButtonsImage = C.constListString(clist)

		// Set the common C callback for all methods.
		attr.pActionFunc = C.CairoDockActionOnAnswerFunc(C.onDialogAnswer)
	}

	attr.iTimeLength = C.gint(1000 * dialog.TimeLength)
	attr.bForceAbove = cbool(dialog.ForceAbove)
	attr.bUseMarkup = cbool(dialog.UseMarkup)

	var widget *C.GtkWidget
	var getValue = func() interface{} { return nil }
	switch typed := dialog.Widget.(type) {

	case cdtype.DialogWidgetText:
		widget, getValue = dialogWidgetText(typed)

	case cdtype.DialogWidgetScale:
		widget, getValue = dialogWidgetScale(typed)

		// default:
		// return errors.New("PopupDialog: invalid widget type")
	}

	if dialog.Buttons != "" && dialog.Callback != nil {
		dialogCall = func(clickedButton int, widget *gtk.Widget) { // No special widget, return button ID.
			dialog.Callback(clickedButton, getValue())
		}
	}

	attr.pUserData = C.intToPointer(1) // unused, but it seems it must be set so the onDialogDestroyed can be called.
	attr.pFreeDataFunc = C.GFreeFunc(C.onDialogDestroyed)

	if widget != nil {
		attr.pInteractiveWidget = widget
		C.gtk_widget_grab_focus(widget)
	}

	removeDialog()
	c := C.gldi_dialog_new(attr)
	dialogPtr = NewDialogFromNative(unsafe.Pointer(c))
	return dialogPtr
}

//
//-------------------------------------------------------------[ WIDGET TEXT ]--

func dialogWidgetText(data cdtype.DialogWidgetText) (*C.GtkWidget, func() interface{}) {
	var widget *gtk.Widget
	var getValue func() interface{}

	if data.MultiLines {
		textview, _ := gtk.TextViewNew()
		scroll, _ := gtk.ScrolledWindowNew(nil, nil)
		scroll.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_AUTOMATIC)
		scroll.Add(textview)
		scroll.Set("width-request", 230)
		scroll.Set("height-request", 130)

		if data.InitialValue != "" {
			buffer, e := textview.GetBuffer()
			if e == nil {
				buffer.SetText(data.InitialValue)
			}
		}
		if !data.Editable {
			textview.SetEditable(false)
		}

		widget = &scroll.Widget
		getValue = func() interface{} {
			buffer, e := textview.GetBuffer()
			if e != nil {
				return ""
			}
			start, end := buffer.GetBounds()
			answer, _ := buffer.GetText(start, end, true)
			return answer
		}

	} else {
		entry, _ := gtk.EntryNew()
		entry.SetHasFrame(false)
		if data.InitialValue != "" {
			entry.SetText(data.InitialValue)
		}
		if !data.Editable {
			entry.SetEditable(false)
		}
		if !data.Visible {
			entry.SetVisibility(false)
		}

		widget = &entry.Widget
		getValue = func() interface{} {
			answer, _ := entry.GetText()
			return answer
		}

		// 	g_object_set (pOneWidget, "width-request", CAIRO_DIALOG_MIN_ENTRY_WIDTH, NULL);

	}

	// 					if (iNbCharsMax != 0)
	// 					{
	// 						gchar *cLabel = g_strdup_printf ("<b>%zd</b>", cInitialText ? strlen (cInitialText) : 0);
	// 						GtkWidget *pLabel = gtk_label_new (cLabel);
	// 						g_free (cLabel);
	// 						gtk_label_set_use_markup (GTK_LABEL (pLabel), TRUE);
	// 						GtkWidget *pBox = gtk_box_new (GTK_ORIENTATION_HORIZONTAL, 3);
	// 						gtk_box_pack_start (GTK_BOX (pBox), pInteractiveWidget, TRUE, TRUE, 0);
	// 						gtk_box_pack_start (GTK_BOX (pBox), pLabel, FALSE, FALSE, 0);
	// 						pInteractiveWidget = pBox;

	// 						if (bMultiLines)
	// 						{
	// 							GtkTextBuffer *pBuffer = gtk_text_view_get_buffer (GTK_TEXT_VIEW (pOneWidget));
	// 							g_signal_connect (pBuffer, "changed", G_CALLBACK (_on_text_changed), pLabel);
	// 							g_object_set_data (G_OBJECT (pBuffer), "nb-chars-max", GINT_TO_POINTER (iNbCharsMax));
	// 						}
	// 						else
	// 						{
	// 							g_signal_connect (pOneWidget, "changed", G_CALLBACK (_on_text_changed), pLabel);
	// 							g_object_set_data (G_OBJECT (pOneWidget), "nb-chars-max", GINT_TO_POINTER (iNbCharsMax));
	// 							gtk_entry_set_width_chars (GTK_ENTRY (pOneWidget), MIN (iNbCharsMax/2, 100));  // a rough estimate is: 140 chars ~ 1024 pixels
	// 						}
	// 					}

	// cstr := (*C.gchar)(C.CString("cd-widget"))
	// defer C.free(unsafe.Pointer((*C.char)(cstr)))
	// p := unsafe.Pointer(widget.GObject)
	// C.g_object_set_data((*C.GObject)(p), cstr, C.gpointer(p))

	return (*C.GtkWidget)(unsafe.Pointer(widget.Native())), getValue
}

//
//------------------------------------------------------------[ WIDGET SCALE ]--

func dialogWidgetScale(data cdtype.DialogWidgetScale) (*C.GtkWidget, func() interface{}) {
	step := (data.MaxValue - data.MinValue) / 100
	scale, _ := gtk.ScaleNewWithRange(gtk.ORIENTATION_HORIZONTAL, data.MinValue, data.MaxValue, step)
	scale.SetValue(data.InitialValue)
	scale.Set("digits", data.NbDigit)
	scale.Set("width-request", 150)

	// C.gldi_dialog_set_widget_text_color((*C.GtkWidget)(unsafe.Pointer(scale.Native()))) // WTF ???

	getValue := func() interface{} { return scale.GetValue() }

	if data.MinLabel == "" && data.MaxLabel == "" {
		return (*C.GtkWidget)(unsafe.Pointer(scale.Native())), getValue
	}

	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	min, _ := gtk.LabelNew(data.MinLabel)
	max, _ := gtk.LabelNew(data.MaxLabel)
	box.PackStart(min, false, false, 0)
	box.PackStart(scale, false, false, 0)
	box.PackStart(max, false, false, 0)

	// 	GtkWidget *pAlign = gtk_alignment_new (1., 1., 0., 0.); // used alignments for labels

	return (*C.GtkWidget)(unsafe.Pointer(box.Native())), getValue
}

//
//-------------------------------------------------------------[ WIDGET LIST ]--

func dialogWidgetList(data cdtype.DialogWidgetText) *C.GtkWidget {
	var widget *gtk.Widget

	// 					gboolean bEditable = FALSE;
	// 					const gchar *cValues = NULL;
	// 					gchar **cValuesList = NULL;
	// 					const gchar *cInitialText = NULL;
	// 					int iInitialValue = 0;

	// 					v = g_hash_table_lookup (hWidgetAttributes, "editable");
	// 					if (v && G_VALUE_HOLDS_BOOLEAN (v))
	// 						bEditable = g_value_get_boolean (v);

	// 					v = g_hash_table_lookup (hWidgetAttributes, "values");
	// 					if (v && G_VALUE_HOLDS_STRING (v))
	// 						cValues = g_value_get_string (v);

	// 					if (cValues != NULL)
	// 						cValuesList = g_strsplit (cValues, ";", -1);

	// 					if (bEditable)
	// 						pOneWidget = gtk_combo_box_text_new_with_entry ();
	// 					else
	// 						pOneWidget = gtk_combo_box_text_new ();
	// 					pInteractiveWidget = pOneWidget;

	// 					if (cValuesList != NULL)
	// 					{
	// 						int i;
	// 						for (i = 0; cValuesList[i] != NULL; i ++)
	// 						{
	// 							gtk_combo_box_text_append_text (GTK_COMBO_BOX_TEXT (pInteractiveWidget), cValuesList[i]);
	// 						}
	// 					}

	// 					v = g_hash_table_lookup (hWidgetAttributes, "initial-value");
	// 					if (bEditable)
	// 					{
	// 						if (v && G_VALUE_HOLDS_STRING (v))
	// 							cInitialText = g_value_get_string (v);
	// 						if (cInitialText != NULL)
	// 						{
	// 							GtkWidget *pEntry = gtk_bin_get_child (GTK_BIN (pInteractiveWidget));
	// 							gtk_entry_set_text (GTK_ENTRY (pEntry), cInitialText);
	// 						}
	// 					}
	// 					else
	// 					{
	// 						if (v && G_VALUE_HOLDS_INT (v))
	// 							iInitialValue = g_value_get_int (v);
	// 						gtk_combo_box_set_active (GTK_COMBO_BOX (pInteractiveWidget), iInitialValue);
	// 					}

	// 					if (attr.cButtonsImage != NULL)
	// 					{
	// 						if (bEditable)
	// 							attr.pActionFunc = (CairoDockActionOnAnswerFunc) cd_dbus_applet_emit_on_answer_combo_entry;
	// 						else
	// 							attr.pActionFunc = (CairoDockActionOnAnswerFunc) cd_dbus_applet_emit_on_answer_combo;
	// 					}

	return (*C.GtkWidget)(unsafe.Pointer(widget.Native()))
}

//

// answer := &listForward{onAnswer}

// 	uncast := (*listForward)(data)

// 	call := (uncast.p).(func(int, *gtk.Widget))
// 	if call != nil {
// 		call(int(clickedButton), w)
// 	}
// }

//
//-------------------------------------------------------------[ C CALLBACKS ]--

//export onDialogDestroyed
func onDialogDestroyed(p C.gpointer) {
	dialogPtr = nil
}

//export onDialogAnswer
func onDialogAnswer(clickedButton C.int, widget *C.GtkWidget, data C.gpointer, dialog *C.CairoDialog) {
	dialogPtr = nil
	if dialogCall == nil {
		return
	}

	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(widget))}
	w := &gtk.Widget{glib.InitiallyUnowned{obj}}

	dialogCall(int(clickedButton), w)
}

func cbool(b bool) C.gboolean {
	if b {
		return C.gboolean(1)
	}
	return C.gboolean(0)
}

func gchar(str string) *C.gchar {
	if str == "" {
		return nil
	}
	return (*C.gchar)(C.CString(str))
}
