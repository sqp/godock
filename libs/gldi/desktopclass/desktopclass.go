// Package desktopclass defines a desktop class informations source.
package desktopclass

// #cgo pkg-config: glib-2.0 gldi
// #include <stdlib.h>                              // free
// #include "cairo-dock-class-manager.h"            // cairo_dock_get_class_command
import "C"

import (
	"github.com/conformal/gotk3/glib"
	"unsafe"
)

// Info defines a desktop class informations source.
//
type Info string

// String returns the desktop class as a string.
//
func (class Info) String() string {
	return string(class)
}

// Name returns the desktop class application name.
//
func (class Info) Name() string {
	cClass := (*C.gchar)(C.CString(class.String()))
	defer C.free(unsafe.Pointer((*C.char)(cClass)))
	return C.GoString((*C.char)(C.cairo_dock_get_class_name(cClass)))
}

// Command returns the desktop class command.
//
func (class Info) Command() string {
	cClass := (*C.gchar)(C.CString(class.String()))
	defer C.free(unsafe.Pointer((*C.char)(cClass)))
	return C.GoString((*C.char)(C.cairo_dock_get_class_command(cClass)))
}

// Icon returns the desktop class icon.
//
func (class Info) Icon() string {
	cClass := (*C.gchar)(C.CString(class.String()))
	defer C.free(unsafe.Pointer((*C.char)(cClass)))
	return C.GoString((*C.char)(C.cairo_dock_get_class_icon(cClass)))
}

// return C.GoString((*C.char)(C.cairo_dock_get_class_wm_class(cClass)))
// return C.GoString((*C.char)(C.cairo_dock_get_class_desktop_file(cClass)))

// MenuItems returns the list of extra commands for the class, by packs of 3
// strings: Name, Command, Icon.
//
func (class Info) MenuItems() (ret [][]string) {
	cClass := (*C.gchar)(C.CString(class.String()))
	defer C.free(unsafe.Pointer((*C.char)(cClass)))
	c := C.cairo_dock_get_class_menu_items(cClass) // do not free.
	list := (*glib.List)(unsafe.Pointer(c))

	for list != nil {
		chars := (*[3]*C.gchar)(unsafe.Pointer(list.Data))
		ret = append(ret, []string{
			C.GoString((*C.char)(chars[0])),
			C.GoString((*C.char)(chars[1])),
			C.GoString((*C.char)(chars[2])),
		})

		list = list.Next
	}
	return ret
}

// const gchar **cairo_dock_get_class_mimetypes (const gchar *cClass)
