// Package shortkeys manages keyboard shortkeys.
package shortkeys

/*
#cgo pkg-config: gldi
#include <stdlib.h>                              // free

#include "cairo-dock-keybinder.h"
*/
import "C"

import (
	"github.com/sqp/godock/libs/cdglobal" // Dock types.

	"unsafe"
)

// Trigger triggers a send shortkey action.
//
func Trigger(keystring string) bool {
	ckey := (*C.gchar)(C.CString(keystring))
	c := C.cairo_dock_trigger_shortkey(ckey)
	C.free(unsafe.Pointer((*C.char)(ckey)))
	return c > 0
}

//
//----------------------------------------------------------------[ SHORTKEY ]--

// Shortkey defines a dock shortkey.
//
type shortkey struct {
	Ptr *C.GldiShortkey
}

// NewFromNative wraps a dock shortkey from C pointer.
//
func NewFromNative(p unsafe.Pointer) cdglobal.Shortkeyer {
	return &shortkey{(*C.GldiShortkey)(p)}
}

func (dr *shortkey) ConfFilePath() string { return C.GoString((*C.char)(dr.Ptr.cConfFilePath)) }
func (dr *shortkey) Demander() string     { return C.GoString((*C.char)(dr.Ptr.cDemander)) }
func (dr *shortkey) Description() string  { return C.GoString((*C.char)(dr.Ptr.cDescription)) }
func (dr *shortkey) Success() bool        { return dr.Ptr.bSuccess > 0 }
func (dr *shortkey) GroupName() string    { return C.GoString((*C.char)(dr.Ptr.cGroupName)) }
func (dr *shortkey) IconFilePath() string { return C.GoString((*C.char)(dr.Ptr.cIconFilePath)) }
func (dr *shortkey) KeyName() string      { return C.GoString((*C.char)(dr.Ptr.cKeyName)) }
func (dr *shortkey) KeyString() string    { return C.GoString((*C.char)(dr.Ptr.keystring)) }

func (dr *shortkey) Rebind(keystring, description string) bool {
	ckey := (*C.gchar)(C.CString(keystring))
	defer C.free(unsafe.Pointer((*C.char)(ckey)))
	var cdesc *C.gchar
	if description != "" {
		cdesc := (*C.gchar)(C.CString(description))
		defer C.free(unsafe.Pointer((*C.char)(cdesc)))
	}

	c := C.gldi_shortkey_rebind(dr.Ptr, ckey, cdesc)
	return c > 0
}
