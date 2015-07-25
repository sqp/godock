// Package keyfile provides Go bindings for GLib's GKeyFile.
package keyfile

// #cgo pkg-config: glib-2.0
// #include <glib.h>
import "C"

// #include "glib.gen.h"
// #include <stdlib.h>
// #include <stdint.h>

import (
	"errors"
	"unsafe"
)

// Desktop files entries.
const (
	DesktopGroup             = "Desktop Entry"
	DesktopKeyCategories     = "Categories"
	DesktopKeyComment        = "Comment"
	DesktopKeyExec           = "Exec"
	DesktopKeyFullname       = "X-GNOME-FullName"
	DesktopKeyGenericName    = "GenericName"
	DesktopKeyGettextDomain  = "X-GNOME-Gettext-Domain"
	DesktopKeyHidden         = "Hidden"
	DesktopKeyIcon           = "Icon"
	DesktopKeyKeywords       = "Keywords"
	DesktopKeyMIMEType       = "MimeType"
	DesktopKeyName           = "Name"
	DesktopKeyNotShowIn      = "NotShowIn"
	DesktopKeyNoDisplay      = "NoDisplay"
	DesktopKeyOnlyShowIn     = "OnlyShowIn"
	DesktopKeyPath           = "Path"
	DesktopKeyStartupNotify  = "StartupNotify"
	DesktopKeyStartupWmClass = "StartupWMClass"
	DesktopKeyTerminal       = "Terminal"
	DesktopKeyTryExec        = "TryExec"
	DesktopKeyType           = "Type"
	DesktopKeyUrl            = "URL"
	DesktopKeyVersion        = "Version"
	DesktopTypeApplication   = "Application"
	DesktopTypeDirectory     = "Directory"
	DesktopTypeLink          = "Link"
)

// Error is a representation of GLib's KeyFileError.
type Error C.gint

const (
	ErrorUnknownEncoding Error = 0
	ErrorParse           Error = 1
	ErrorNotFound        Error = 2
	ErrorKeyNotFound     Error = 3
	ErrorGroupNotFound   Error = 4
	ErrorInvalidValue    Error = 5
)

// Flags is a representation of GLib's KeyFileFlags.
type Flags C.gint

// Keyfile loading flags.
const (
	FlagsNone             Flags = 0
	FlagsKeepComments     Flags = 1
	FlagsKeepTranslations Flags = 2
)

// KeyFile is a representation of Glib's GKeyFile.
type KeyFile struct {
	cKey *C.GKeyFile
}

// New is a wrapper around g_key_file_new().
func New() *KeyFile {
	kf := &KeyFile{cKey: C.g_key_file_new()}
	return kf
}

// NewFromNative wraps a pointer to a C keyfile.
func NewFromNative(p unsafe.Pointer) *KeyFile {
	if p == nil {
		return nil
	}
	return &KeyFile{(*C.GKeyFile)(p)}
}

// NewFromFile returns a loaded keyfile if possible.
func NewFromFile(file string, flags Flags) (*KeyFile, error) {
	kf := New()
	_, e := kf.LoadFromFile(file, flags)
	if e != nil {
		return nil, e
	}
	return kf, nil
}

// Free is a wrapper around g_key_file_free().
func (this *KeyFile) Free() {
	C.g_key_file_free(this.cKey)
}

// LoadFromFile is a wrapper around g_key_file_load_from_file().
func (this *KeyFile) LoadFromFile(file string, flags Flags) (bool, error) {
	var cstr *C.char = C.CString(file)
	defer C.g_free(C.gpointer(cstr))
	var cerr *C.GError
	c := C.g_key_file_load_from_file(this.cKey, (*C.gchar)(cstr), C.GKeyFileFlags(flags), &cerr)
	return c != 0, goError(cerr)
}

// ToData is a wrapper around g_key_file_to_data().
func (this *KeyFile) ToData() (uint64, string, error) {
	var clength C.gsize
	var cerr *C.GError
	c := C.g_key_file_to_data(this.cKey, &clength, &cerr)
	defer C.g_free(C.gpointer(c))
	return uint64(clength), C.GoString((*C.char)(c)), goError(cerr)
}

// ToNative returns the pointer to the underlying GKeyFile.
func (this *KeyFile) ToNative() *C.GKeyFile {
	return this.cKey
}

// HasKey is a wrapper around g_key_file_has_key().
func (this *KeyFile) HasKey(group string, key string) bool {
	cGroup := (*C.gchar)(C.CString(group))
	defer C.g_free(C.gpointer(cGroup))
	cKey := (*C.gchar)(C.CString(key))
	defer C.g_free(C.gpointer(cKey))
	return goBool(C.g_key_file_has_key(this.cKey, cGroup, cKey, nil))
}

//
//---------------------------------------------------------------------[ GET ]--

// GetGroups is a wrapper around g_key_file_get_groups().
func (this *KeyFile) GetGroups() (uint64, []string) {
	var length C.gsize
	c := C.g_key_file_get_groups(this.cKey, &length)
	var list []string
	for i := 0; i < int(length); i++ {
		if str := C.GoString((*(*[999999]*C.char)(unsafe.Pointer(c)))[i]); str != "" {
			list = append(list, str)
		} else {
			println("-----------------------------------------------dropped empty group")
		}
		C.g_free(C.gpointer((*(*[999999]*C.char)(unsafe.Pointer(c)))[i]))
	}
	return uint64(length), list
}

// GetKeys is a wrapper around g_key_file_get_keys().
func (this *KeyFile) GetKeys(group string) (uint64, []string, error) {
	cGroup := (*C.gchar)(C.CString(group))
	defer C.g_free(C.gpointer(cGroup))
	var length C.gsize
	var cErr *C.GError
	c := C.g_key_file_get_keys(this.cKey, cGroup, &length, &cErr)
	list := make([]string, uint64(length))
	for i := range list {
		list[i] = C.GoString((*(*[999999]*C.char)(unsafe.Pointer(c)))[i])
		C.g_free(C.gpointer((*(*[999999]*C.char)(unsafe.Pointer(c)))[i]))
	}
	return uint64(length), list, goError(cErr)
}

// GetBooleanList is a wrapper around g_key_file_get_boolean_list().
func (this *KeyFile) GetBooleanList(group string, key string) (length uint64, list []bool, e error) {
	var c C.gpointer
	length, c, e = this.getList(group, key, "bool")
	defer C.g_free(c)
	list = make([]bool, length)
	for i := range list {
		list[i] = (*(*[999999]C.int)(c))[i] != 0
	}
	return length, list, e
}

// GetDoubleList is a wrapper around g_key_file_get_double_list().
func (this *KeyFile) GetDoubleList(group string, key string) (length uint64, list []float64, e error) {
	var c C.gpointer
	length, c, e = this.getList(group, key, "float64")
	defer C.g_free(c)
	list = make([]float64, length)
	for i := range list {
		list[i] = float64((*(*[999999]C.double)(c))[i])
	}
	return length, list, e
}

// GetIntegerList is a wrapper around g_key_file_get_integer_list().
func (this *KeyFile) GetIntegerList(group string, key string) (length uint64, list []int, e error) {
	var c C.gpointer
	length, c, e = this.getList(group, key, "int")
	defer C.g_free(c)
	list = make([]int, length)
	for i := range list {
		list[i] = int((*(*[999999]C.int)(c))[i])
	}
	return length, list, e
}

// GetStringList is a wrapper around g_key_file_get_string_list().
func (this *KeyFile) GetStringList(group string, key string) (length uint64, list []string, e error) {
	var c C.gpointer
	length, c, e = this.getList(group, key, "string")
	defer C.g_free(c)
	list = make([]string, length)
	for i := range list {
		list[i] = C.GoString((*(*[999999]*C.char)(c))[i])
	}
	return length, list, e
}

func (this *KeyFile) getList(group, key, typ string) (uint64, C.gpointer, error) {
	cGroup := (*C.gchar)(C.CString(group))
	defer C.g_free(C.gpointer(cGroup))
	cKey := (*C.gchar)(C.CString(key))
	defer C.g_free(C.gpointer(cKey))
	var length C.gsize
	var cErr *C.GError
	var c C.gpointer
	switch typ {
	case "bool":
		c = C.gpointer(C.g_key_file_get_boolean_list(this.cKey, cGroup, cKey, &length, &cErr))
	case "int":
		c = C.gpointer(C.g_key_file_get_integer_list(this.cKey, cGroup, cKey, &length, &cErr))
	case "float64":
		c = C.gpointer(C.g_key_file_get_double_list(this.cKey, cGroup, cKey, &length, &cErr))
	case "string":
		c = C.gpointer(C.g_key_file_get_string_list(this.cKey, cGroup, cKey, &length, &cErr))

	}
	return uint64(length), c, goError(cErr)
}

func (this *KeyFile) getOne(group string, key string, typ string) (interface{}, error) {
	cGroup := (*C.gchar)(C.CString(group))
	defer C.g_free(C.gpointer(cGroup))
	cKey := (*C.gchar)(C.CString(key))
	defer C.g_free(C.gpointer(cKey))
	var cErr *C.GError

	var c interface{}
	switch typ {
	case "bool":
		c = goBool(C.g_key_file_get_boolean(this.cKey, cGroup, cKey, &cErr))

	case "int":
		c = int(C.g_key_file_get_integer(this.cKey, cGroup, cKey, &cErr))

	case "comment":
		cstr := C.g_key_file_get_comment(this.cKey, cGroup, cKey, &cErr)
		c = C.GoString((*C.char)(cstr))
		C.g_free(C.gpointer(cstr))

	case "string":
		cstr := C.g_key_file_get_string(this.cKey, cGroup, cKey, &cErr)
		c = C.GoString((*C.char)(cstr))
		C.g_free(C.gpointer(cstr))

	}
	return c, goError(cErr)
}

// GetBoolean is a wrapper around g_key_file_get_boolean().
func (this *KeyFile) GetBoolean(group string, key string) (bool, error) {
	ret, e := this.getOne(group, key, "bool")
	return ret.(bool), e
}

// GetInteger is a wrapper around g_key_file_get_integer().
func (this *KeyFile) GetInteger(group string, key string) (int, error) {
	ret, e := this.getOne(group, key, "int")
	return ret.(int), e
}

// GetComment is a wrapper around g_key_file_get_comment().
func (this *KeyFile) GetComment(group string, key string) (string, error) {
	ret, e := this.getOne(group, key, "comment")
	return ret.(string), e
}

// GetString is a wrapper around g_key_file_get_string().
func (this *KeyFile) GetString(group string, key string) (string, error) {
	ret, e := this.getOne(group, key, "string")
	return ret.(string), e
}

//
//---------------------------------------------------------------------[ SET ]--

// Set is a generic wrapper around g_key_file_set_xxx() with type assertion.
func (this *KeyFile) Set(group string, key string, uncasted interface{}) {
	cGroup := (*C.gchar)(C.CString(group))
	defer C.g_free(C.gpointer(cGroup))
	cKey := (*C.gchar)(C.CString(key))
	defer C.g_free(C.gpointer(cKey))

	switch value := uncasted.(type) {
	case bool:
		C.g_key_file_set_boolean(this.cKey, cGroup, cKey, cBool(value))

	case int:
		C.g_key_file_set_integer(this.cKey, cGroup, cKey, C.gint(value))

	case float64:
		C.g_key_file_set_double(this.cKey, cGroup, cKey, C.gdouble(value))

	case string:
		cstr := (*C.gchar)(C.CString(value))
		defer C.g_free(C.gpointer(cstr))
		C.g_key_file_set_string(this.cKey, cGroup, cKey, cstr)

	case []bool:
		clist := cListBool(value)
		defer C.g_free(C.gpointer(clist))
		C.g_key_file_set_boolean_list(this.cKey, cGroup, cKey, clist, C.gsize(len(value)))

	case []int:
		clist := cListInt(value)
		defer C.g_free(C.gpointer(clist))
		C.g_key_file_set_integer_list(this.cKey, cGroup, cKey, clist, C.gsize(len(value)))

	case []float64:
		clist := cListDouble(value)
		defer C.g_free(C.gpointer(clist))
		C.g_key_file_set_double_list(this.cKey, cGroup, cKey, clist, C.gsize(len(value)))

	case []string:
		clist := cListString(value)
		C.g_key_file_set_string_list(this.cKey, cGroup, cKey, clist, C.gsize(len(value)))
		for i := range value {
			C.g_free(C.gpointer((*(*[999999]*C.gchar)(unsafe.Pointer(clist)))[i]))
		}
		C.g_free(C.gpointer(clist))
	}
}

// SetBoolean is a wrapper around g_key_file_set_boolean().
func (this *KeyFile) SetBoolean(group string, key string, value bool) {
	this.Set(group, key, value)
}

// SetBooleanList is a wrapper around g_key_file_set_boolean_list().
func (this *KeyFile) SetBooleanList(group string, key string, value []bool) {
	this.Set(group, key, value)
}

// SetDouble is a wrapper around g_key_file_set_double().
func (this *KeyFile) SetDouble(group string, key string, value float64) {
	this.Set(group, key, value)
}

// SetDoubleList is a wrapper around g_key_file_set_double_list().
func (this *KeyFile) SetDoubleList(group string, key string, value []float64) {
	this.Set(group, key, value)
}

// func (this0 *KeyFile) SetInt64(group_name0 string, key0 string, value0 int64) {
// 	C.g_key_file_set_int64(this1, group_name1, key1, int64_t(value))
// }

// SetInteger is a wrapper around g_key_file_set_integer().
func (this *KeyFile) SetInteger(group string, key string, value int) {
	this.Set(group, key, value)
}

// SetIntegerList is a wrapper around g_key_file_set_integer_list().
func (this *KeyFile) SetIntegerList(group string, key string, value []int) {
	this.Set(group, key, value)
}

// SetString is a wrapper around g_key_file_set_string().
func (this *KeyFile) SetString(group string, key string, value string) {
	this.Set(group, key, value)
}

// SetStringList is a wrapper around g_key_file_set_string_list().
func (this *KeyFile) SetStringList(group string, key string, value []string) {
	this.Set(group, key, value)
}

//
//-----------------------------------------------------------------[ HELPERS ]--

func goError(err *C.GError) error {
	if err != nil {
		defer C.g_error_free(err)
		return errors.New(C.GoString((*C.char)(err.message)))
	}
	return nil
}

func goBool(b C.gboolean) bool {
	if b > 0 {
		return true
	}
	return false
}

func cBool(b bool) C.gboolean {
	if b {
		return 1
	}
	return 0
}

func cListBool(value []bool) *C.gboolean {
	var clist *C.gboolean
	clist = (*C.gboolean)(C.malloc(C.size_t(int(unsafe.Sizeof(*clist)) * len(value))))
	for i, e := range value {
		(*(*[999999]C.gboolean)(unsafe.Pointer(clist)))[i] = cBool(e)
	}
	return clist
}

func cListInt(value []int) *C.gint {
	var clist *C.gint
	clist = (*C.gint)(C.malloc(C.size_t(int(unsafe.Sizeof(*clist)) * len(value))))
	for k, v := range value {
		(*(*[999999]C.gint)(unsafe.Pointer(clist)))[k] = C.gint(v)
	}
	return clist
}

func cListDouble(value []float64) *C.gdouble {
	var clist *C.gdouble
	clist = (*C.gdouble)(C.malloc(C.size_t(int(unsafe.Sizeof(*clist)) * len(value))))
	for i, e := range value {
		(*(*[999999]C.gdouble)(unsafe.Pointer(clist)))[i] = C.gdouble(e)
	}
	return clist
}

func cListString(value []string) **C.gchar {
	var clist **C.gchar
	clist = (**C.gchar)(C.malloc(C.size_t(int(unsafe.Sizeof(*clist)) * (len(value) + 1))))
	for i, e := range value {
		(*(*[999999]*C.gchar)(unsafe.Pointer(clist)))[i] = (*C.gchar)(C.CString(e))
	}
	(*(*[999999]*C.gchar)(unsafe.Pointer(clist)))[len(value)] = nil
	return clist
}
