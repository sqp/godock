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
	"github.com/gotk3/gotk3/gtk"

	"github.com/sqp/godock/libs/cdtype" // Applets types.
	"github.com/sqp/godock/libs/gldi"

	"github.com/sqp/godock/widgets/gtk/newgtk"
	"github.com/sqp/godock/widgets/gtk/togtk"

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

// Dialog wraps a dock dialog widget.
//
type Dialog struct {
	Ptr *C.CairoDialog
}

// NewDialogFromNative wraps a pointer to a dock dialog.
//
func NewDialogFromNative(p unsafe.Pointer) *Dialog {
	if p == nil {
		return nil
	}
	return &Dialog{(*C.CairoDialog)(p)}
}

// ToNative returns a pointer to the native dock dialog.
//
func (o *Dialog) ToNative() unsafe.Pointer {
	return unsafe.Pointer(o.Ptr)
}

//
//----------------------------------------------------------[ COMMON DIALOGS ]--

// ShowGeneralMessage opens a simple dialog for general information message.
//
func ShowGeneralMessage(str string, duration float64) {
	cstr := (*C.gchar)(C.CString(str))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	C.gldi_dialog_show_general_message(cstr, C.double(duration))
}

// ShowTemporaryWithIcon opens a simple dialog with a timeout.
//
func ShowTemporaryWithIcon(str string, icon gldi.Icon, container *gldi.Container, duration float64, iconPath string) {
	cstr := (*C.gchar)(C.CString(str))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	cpath := (*C.gchar)(C.CString(iconPath))
	defer C.free(unsafe.Pointer((*C.char)(cpath)))

	cicon := (*C.Icon)(unsafe.Pointer(icon.ToNative()))
	ccontainer := (*C.GldiContainer)(unsafe.Pointer(container.Ptr))

	C.gldi_dialog_show_temporary_with_icon(cstr, cicon, ccontainer, C.double(duration), cpath)
}

// ShowWithQuestion opens a dialog with a question to answer.
//
func ShowWithQuestion(str string, icon gldi.Icon, container *gldi.Container, iconPath string, onAnswer func(int, *gtk.Widget)) {
	cstr := (*C.gchar)(C.CString(str))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	cpath := (*C.gchar)(C.CString(iconPath))
	defer C.free(unsafe.Pointer((*C.char)(cpath)))

	cicon := (*C.Icon)(unsafe.Pointer(icon.ToNative()))
	ccontainer := (*C.GldiContainer)(unsafe.Pointer(container.Ptr))

	dialogCall = onAnswer
	removeDialog()
	c := C.gldi_dialog_show_with_question(cstr, cicon, ccontainer, cpath, C.CairoDockActionOnAnswerFunc(C.onDialogAnswer), nil, nil)
	dialogPtr = NewDialogFromNative(unsafe.Pointer(c))
}

//
//-----------------------------------------------------------[ CUSTOM DIALOG ]--

// NewDialog creates a custom dialog.
//
func NewDialog(icon gldi.Icon, container *gldi.Container, dialog cdtype.DialogData) *Dialog {
	dialogCall = nil

	// Common dialog attributes.
	attr := new(C.CairoDialogAttr)
	if icon != nil {
		attr.pIcon = (*C.Icon)(unsafe.Pointer(icon.ToNative()))
	}
	attr.pContainer = (*C.GldiContainer)(unsafe.Pointer(container.Ptr))

	var clear func()
	if dialog.Icon != "" {
		// w,h :=
		// 		cairo_dock_get_icon_extent (pIcon, &w, &h);
		// 		cImageFilePath = cairo_dock_search_icon_s_path (g_value_get_string (v), MAX (w, h));
		attr.cImageFilePath, clear = gchar(dialog.Icon)
	} else {
		attr.cImageFilePath, clear = gchar("same icon")
	}
	defer clear()

	if dialog.Message != "" {
		var clear func()
		attr.cText, clear = gchar(dialog.Message)
		defer clear()
	}

	if dialog.Buttons != "" {
		cstr, freestr := gchar(dialog.Buttons)
		csep, freesep := gchar(";")
		clist := C.g_strsplit(cstr, csep, -1) // NULL-terminated
		freestr()
		freesep()
		defer C.g_strfreev(clist)
		attr.cButtonsImage = C.constListString(clist)

		// Set the common C callback for all methods.
		attr.pActionFunc = C.CairoDockActionOnAnswerFunc(C.onDialogAnswer)
	}

	attr.iTimeLength = C.gint(1000 * dialog.TimeLength)
	attr.bForceAbove = cbool(dialog.ForceAbove)
	attr.bUseMarkup = cbool(dialog.UseMarkup)

	attr.pUserData = C.intToPointer(1) // unused, but it seems it must be set so the onDialogDestroyed can be called.
	attr.pFreeDataFunc = C.GFreeFunc(C.onDialogDestroyed)

	var widget *gtk.Widget
	var getValue = func() interface{} { return nil }
	switch typed := dialog.Widget.(type) {

	case cdtype.DialogWidgetText:
		widget, getValue = dialogWidgetText(typed)

	case cdtype.DialogWidgetScale:
		widget, getValue = dialogWidgetScale(typed)

	case cdtype.DialogWidgetList:
		widget, getValue = dialogWidgetList(typed)

		// default:
		// return errors.New("PopupDialog: invalid widget type")
	}

	if widget != nil {
		attr.pInteractiveWidget = (*C.GtkWidget)(unsafe.Pointer(widget.Native()))
	}

	if dialog.Buttons != "" && dialog.Callback != nil {
		dialogCall = func(clickedButton int, _ *gtk.Widget) { // No special widget, return button ID.
			dialog.Callback(clickedButton, getValue())
		}
	}

	removeDialog()
	c := C.gldi_dialog_new(attr)
	dialogPtr = NewDialogFromNative(unsafe.Pointer(c))
	return dialogPtr
}

//
//-------------------------------------------------------------[ WIDGET TEXT ]--

func dialogWidgetText(data cdtype.DialogWidgetText) (*gtk.Widget, func() interface{}) {
	var widget *gtk.Widget
	var getValue func() interface{}

	if data.MultiLines {
		textview := newgtk.TextView()
		scroll := newgtk.ScrolledWindow(nil, nil)
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
		if data.Locked {
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
		entry := newgtk.Entry()
		entry.SetHasFrame(false)
		if data.InitialValue != "" {
			entry.SetText(data.InitialValue)
		}
		if data.Locked {
			entry.SetEditable(false)
		}
		if data.Hidden {
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

	widget.GrabFocus()
	return widget, getValue
}

//
//------------------------------------------------------------[ WIDGET SCALE ]--

func dialogWidgetScale(data cdtype.DialogWidgetScale) (*gtk.Widget, func() interface{}) {
	step := (data.MaxValue - data.MinValue) / 100
	scale := newgtk.ScaleWithRange(gtk.ORIENTATION_HORIZONTAL, data.MinValue, data.MaxValue, step)
	scale.SetValue(data.InitialValue)
	scale.Set("digits", data.NbDigit)
	scale.Set("width-request", 150)
	scale.GrabFocus()

	// C.gldi_dialog_set_widget_text_color((*C.GtkWidget)(unsafe.Pointer(scale.Native()))) // WTF ???

	getValue := func() interface{} { return scale.GetValue() }

	if data.MinLabel == "" && data.MaxLabel == "" {
		return &scale.Widget, getValue
	}

	box := newgtk.Box(gtk.ORIENTATION_HORIZONTAL, 0)

	box.PackStart(newgtk.Label(data.MinLabel), false, false, 0)
	box.PackStart(scale, false, false, 0)
	box.PackStart(newgtk.Label(data.MaxLabel), false, false, 0)

	// 	GtkWidget *pAlign = gtk_alignment_new (1., 1., 0., 0.); // used alignments for labels
	return &box.Widget, getValue
}

//
//-------------------------------------------------------------[ WIDGET LIST ]--

func dialogWidgetList(data cdtype.DialogWidgetList) (*gtk.Widget, func() interface{}) {
	var getValue func() interface{}
	widget := newgtk.ComboBoxText()

	// Fill the list with user choices.
	for _, val := range data.Values {
		widget.AppendText(val)
	}

	if data.Editable {
		// Add entry manually so we don't have to recast it after a GetChild.
		entry := newgtk.Entry()
		widget.Add(entry)
		widget.Connect("changed", func() { entry.SetText(widget.GetActiveText()) })

		getValue = func() interface{} { str, _ := entry.GetText(); return str }

		val, ok := data.InitialValue.(string)
		if ok && val != "" {
			entry.SetText(val)
		}

	} else {
		getValue = func() interface{} { return widget.GetActive() }

		val, _ := data.InitialValue.(int)
		widget.SetActive(val)
	}

	widget.GrabFocus()
	return &widget.Widget, getValue
}

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
	dialogCall(int(clickedButton), togtk.Widget(unsafe.Pointer(widget)))
}

func cbool(b bool) C.gboolean {
	if b {
		return C.gboolean(1)
	}
	return C.gboolean(0)
}

func gchar(str string) (*C.gchar, func()) {
	if str == "" {
		return nil, func() {}
	}
	cstr := (*C.gchar)(C.CString(str))
	return cstr, func() { C.free(unsafe.Pointer((*C.char)(cstr))) }
}
