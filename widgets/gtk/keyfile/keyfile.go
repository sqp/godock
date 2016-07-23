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
	DesktopKeyURL            = "URL"
	DesktopKeyVersion        = "Version"
	DesktopTypeApplication   = "Application"
	DesktopTypeDirectory     = "Directory"
	DesktopTypeLink          = "Link"
)

// Error is a representation of GLib's KeyFileError.
type Error C.gint

// Error type list.
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
func (kf *KeyFile) Free() {
	C.g_key_file_free(kf.cKey)
}

// LoadFromFile is a wrapper around g_key_file_load_from_file().
func (kf *KeyFile) LoadFromFile(file string, flags Flags) (bool, error) {
	var cstr = C.CString(file)
	defer C.g_free(C.gpointer(cstr))
	var cerr *C.GError
	c := C.g_key_file_load_from_file(kf.cKey, (*C.gchar)(cstr), C.GKeyFileFlags(flags), &cerr)
	return c != 0, goError(cerr)
}

// ToData is a wrapper around g_key_file_to_data().
func (kf *KeyFile) ToData() (uint64, string, error) {
	var clength C.gsize
	var cerr *C.GError
	c := C.g_key_file_to_data(kf.cKey, &clength, &cerr)
	defer C.g_free(C.gpointer(c))
	return uint64(clength), C.GoString((*C.char)(c)), goError(cerr)
}

// ToNative returns the pointer to the underlying GKeyFile.
func (kf *KeyFile) ToNative() *C.GKeyFile {
	return kf.cKey
}

// HasKey is a wrapper around g_key_file_has_key().
func (kf *KeyFile) HasKey(group string, key string) bool {
	cGroup := (*C.gchar)(C.CString(group))
	defer C.g_free(C.gpointer(cGroup))
	cKey := (*C.gchar)(C.CString(key))
	defer C.g_free(C.gpointer(cKey))
	return goBool(C.g_key_file_has_key(kf.cKey, cGroup, cKey, nil))
}

//
//---------------------------------------------------------------------[ GET ]--

// GetGroups is a wrapper around g_key_file_get_groups().
func (kf *KeyFile) GetGroups() (uint64, []string) {
	var length C.gsize
	c := C.g_key_file_get_groups(kf.cKey, &length)
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
func (kf *KeyFile) GetKeys(group string) (uint64, []string, error) {
	cGroup := (*C.gchar)(C.CString(group))
	defer C.g_free(C.gpointer(cGroup))
	var length C.gsize
	var cErr *C.GError
	c := C.g_key_file_get_keys(kf.cKey, cGroup, &length, &cErr)
	list := make([]string, uint64(length))
	for i := range list {
		list[i] = C.GoString((*(*[999999]*C.char)(unsafe.Pointer(c)))[i])
		C.g_free(C.gpointer((*(*[999999]*C.char)(unsafe.Pointer(c)))[i]))
	}
	return uint64(length), list, goError(cErr)
}

// ListBool is a wrapper around g_key_file_get_boolean_list().
func (kf *KeyFile) ListBool(group string, key string) ([]bool, error) {
	// var c C.gpointer
	length, c, e := kf.getList(group, key, "bool")
	defer C.g_free(c)
	list := make([]bool, length)
	for i := range list {
		list[i] = (*(*[999999]C.int)(c))[i] != 0
	}
	return list, e
}

// ListInt is a wrapper around g_key_file_get_integer_list().
func (kf *KeyFile) ListInt(group string, key string) ([]int, error) {
	length, c, e := kf.getList(group, key, "int")
	defer C.g_free(c)
	list := make([]int, length)
	for i := range list {
		list[i] = int((*(*[999999]C.int)(c))[i])
	}
	return list, e
}

// ListFloat is a wrapper around g_key_file_get_double_list().
func (kf *KeyFile) ListFloat(group string, key string) ([]float64, error) {
	// var c C.gpointer
	length, c, e := kf.getList(group, key, "float64")
	defer C.g_free(c)
	list := make([]float64, length)
	for i := range list {
		list[i] = float64((*(*[999999]C.double)(c))[i])
	}
	return list, e
}

// ListString is a wrapper around g_key_file_get_string_list().
func (kf *KeyFile) ListString(group string, key string) ([]string, error) {
	length, c, e := kf.getList(group, key, "string")
	defer C.g_free(c)
	list := make([]string, length)
	for i := range list {
		list[i] = C.GoString((*(*[999999]*C.char)(c))[i])
	}
	return list, e
}

func (kf *KeyFile) getList(group, key, typ string) (uint64, C.gpointer, error) {
	cGroup := (*C.gchar)(C.CString(group))
	defer C.g_free(C.gpointer(cGroup))
	cKey := (*C.gchar)(C.CString(key))
	defer C.g_free(C.gpointer(cKey))
	var length C.gsize
	var cErr *C.GError
	var c C.gpointer
	switch typ {
	case "bool":
		c = C.gpointer(C.g_key_file_get_boolean_list(kf.cKey, cGroup, cKey, &length, &cErr))
	case "int":
		c = C.gpointer(C.g_key_file_get_integer_list(kf.cKey, cGroup, cKey, &length, &cErr))
	case "float64":
		c = C.gpointer(C.g_key_file_get_double_list(kf.cKey, cGroup, cKey, &length, &cErr))
	case "string":
		c = C.gpointer(C.g_key_file_get_string_list(kf.cKey, cGroup, cKey, &length, &cErr))

	}
	return uint64(length), c, goError(cErr)
}

// Get gets a value from the keyfile. Must be used with a pointer to value.
//
func (kf *KeyFile) Get(group string, key string, val interface{}) error {
	switch ptr := val.(type) {
	case *[]bool:
		cast, e := kf.ListBool(group, key)
		*ptr = cast
		return e

	case *[]int:
		cast, e := kf.ListInt(group, key)
		*ptr = cast
		return e

	case *[]float64:
		cast, e := kf.ListFloat(group, key)
		*ptr = cast
		return e

	case *[]string:
		cast, e := kf.ListString(group, key)
		*ptr = cast
		return e
	}

	cGroup := (*C.gchar)(C.CString(group))
	defer C.g_free(C.gpointer(cGroup))
	cKey := (*C.gchar)(C.CString(key))
	defer C.g_free(C.gpointer(cKey))
	var cErr *C.GError

	switch ptr := val.(type) {
	case *bool:
		*ptr = goBool(C.g_key_file_get_boolean(kf.cKey, cGroup, cKey, &cErr))

	case *int:
		*ptr = int(C.g_key_file_get_integer(kf.cKey, cGroup, cKey, &cErr))

	case *float64:
		*ptr = float64(C.g_key_file_get_double(kf.cKey, cGroup, cKey, &cErr))

	case *string:
		cstr := C.g_key_file_get_string(kf.cKey, cGroup, cKey, &cErr)
		*ptr = C.GoString((*C.char)(cstr))
		C.g_free(C.gpointer(cstr))

	}
	return goError(cErr)
}

// GetOne returns a key value as interface.
//
// valid types are:
//   bool, int, float64, string, comment
//   listbool, listint, listfloat64, liststring,
//
func (kf *KeyFile) GetOne(group string, key string, typ string) (interface{}, error) {
	switch typ {
	case "listbool":
		return kf.ListBool(group, key)

	case "listint":
		return kf.ListInt(group, key)

	case "listfloat64":
		return kf.ListFloat(group, key)

	case "liststring":
		return kf.ListString(group, key)
	}

	cGroup := (*C.gchar)(C.CString(group))
	defer C.g_free(C.gpointer(cGroup))
	cKey := (*C.gchar)(C.CString(key))
	defer C.g_free(C.gpointer(cKey))
	var cErr *C.GError

	var c interface{}
	switch typ {
	case "bool":
		c = goBool(C.g_key_file_get_boolean(kf.cKey, cGroup, cKey, &cErr))

	case "int":
		c = int(C.g_key_file_get_integer(kf.cKey, cGroup, cKey, &cErr))

	case "float64":
		c = float64(C.g_key_file_get_double(kf.cKey, cGroup, cKey, &cErr))

	case "comment":
		cstr := C.g_key_file_get_comment(kf.cKey, cGroup, cKey, &cErr)
		c = C.GoString((*C.char)(cstr))
		C.g_free(C.gpointer(cstr))

	case "string":
		cstr := C.g_key_file_get_string(kf.cKey, cGroup, cKey, &cErr)
		c = C.GoString((*C.char)(cstr))
		C.g_free(C.gpointer(cstr))

	}
	return c, goError(cErr)
}

// Bool is a wrapper around g_key_file_get_boolean().
func (kf *KeyFile) Bool(group string, key string) (bool, error) {
	ret, e := kf.GetOne(group, key, "bool")
	return ret.(bool), e
}

// Int is a wrapper around g_key_file_get_integer().
func (kf *KeyFile) Int(group string, key string) (int, error) {
	ret, e := kf.GetOne(group, key, "int")
	return ret.(int), e
}

// Float is a wrapper around g_key_file_get_double().
func (kf *KeyFile) Float(group string, key string) (float64, error) {
	ret, e := kf.GetOne(group, key, "float64")
	return ret.(float64), e
}

// String is a wrapper around g_key_file_get_string().
func (kf *KeyFile) String(group string, key string) (string, error) {
	ret, e := kf.GetOne(group, key, "string")
	return ret.(string), e
}

// GetComment is a wrapper around g_key_file_get_comment().
func (kf *KeyFile) GetComment(group string, key string) (string, error) {
	ret, e := kf.GetOne(group, key, "comment")
	return ret.(string), e
}

//
//---------------------------------------------------------------------[ SET ]--

// Set is a generic wrapper around g_key_file_set_xxx() with type assertion.
func (kf *KeyFile) Set(group string, key string, uncasted interface{}) error {
	cGroup := (*C.gchar)(C.CString(group))
	defer C.g_free(C.gpointer(cGroup))
	cKey := (*C.gchar)(C.CString(key))
	defer C.g_free(C.gpointer(cKey))

	switch value := uncasted.(type) {
	case bool:
		C.g_key_file_set_boolean(kf.cKey, cGroup, cKey, cBool(value))

	case int:
		C.g_key_file_set_integer(kf.cKey, cGroup, cKey, C.gint(value))

	case float64:
		C.g_key_file_set_double(kf.cKey, cGroup, cKey, C.gdouble(value))

	case string:
		cstr := (*C.gchar)(C.CString(value))
		defer C.g_free(C.gpointer(cstr))
		C.g_key_file_set_string(kf.cKey, cGroup, cKey, cstr)

	case []bool:
		clist := cListBool(value)
		defer C.g_free(C.gpointer(clist))
		C.g_key_file_set_boolean_list(kf.cKey, cGroup, cKey, clist, C.gsize(len(value)))

	case []int:
		clist := cListInt(value)
		defer C.g_free(C.gpointer(clist))
		C.g_key_file_set_integer_list(kf.cKey, cGroup, cKey, clist, C.gsize(len(value)))

	case []float64:
		clist := cListDouble(value)
		defer C.g_free(C.gpointer(clist))
		C.g_key_file_set_double_list(kf.cKey, cGroup, cKey, clist, C.gsize(len(value)))

	case []string:
		clist := cListString(value)
		C.g_key_file_set_string_list(kf.cKey, cGroup, cKey, clist, C.gsize(len(value)))
		for i := range value {
			C.g_free(C.gpointer((*(*[999999]*C.gchar)(unsafe.Pointer(clist)))[i]))
		}
		C.g_free(C.gpointer(clist))

	default:
		return errors.New("type unknown")
	}
	return nil
}

// SetBool is a wrapper around g_key_file_set_boolean().
func (kf *KeyFile) SetBool(group string, key string, value bool) {
	kf.Set(group, key, value)
}

// SetListBool is a wrapper around g_key_file_set_boolean_list().
func (kf *KeyFile) SetListBool(group string, key string, value []bool) {
	kf.Set(group, key, value)
}

// SetFloat is a wrapper around g_key_file_set_double().
func (kf *KeyFile) SetFloat(group string, key string, value float64) {
	kf.Set(group, key, value)
}

// SetListFloat is a wrapper around g_key_file_set_double_list().
func (kf *KeyFile) SetListFloat(group string, key string, value []float64) {
	kf.Set(group, key, value)
}

// func (this0 *KeyFile) SetInt64(group_name0 string, key0 string, value0 int64) {
// 	C.g_key_file_set_int64(this1, group_name1, key1, int64_t(value))
// }

// SetInt is a wrapper around g_key_file_set_integer().
func (kf *KeyFile) SetInt(group string, key string, value int) {
	kf.Set(group, key, value)
}

// SetListInt is a wrapper around g_key_file_set_integer_list().
func (kf *KeyFile) SetListInt(group string, key string, value []int) {
	kf.Set(group, key, value)
}

// SetString is a wrapper around g_key_file_set_string().
func (kf *KeyFile) SetString(group string, key string, value string) {
	kf.Set(group, key, value)
}

// SetListString is a wrapper around g_key_file_set_string_list().
func (kf *KeyFile) SetListString(group string, key string, value []string) {
	kf.Set(group, key, value)
}

//
//------------------------------------------------------------------[ VALUER ]--

// Valuer gives access to a storage group/key value. Implements cftype.Valuer
//
type Valuer struct {
	kf    *KeyFile
	group string
	name  string
}

// NewValuer creates a valuer for the key matching group and name.
//
func NewValuer(kf *KeyFile, group, name string) *Valuer {
	return &Valuer{
		kf:    kf,
		group: group,
		name:  name,
	}
}

// Get assigns the value to the given pointer to value (of the matching type).
//
func (o *Valuer) Get(v interface{}) { o.kf.Get(o.group, o.name, v) }

// Bool returns the value as bool.
func (o *Valuer) Bool() (v bool) {
	o.kf.Get(o.group, o.name, &v)
	return
}

// Int returns the value as bool.
func (o *Valuer) Int() (v int) {
	o.kf.Get(o.group, o.name, &v)
	return
}

// Float returns the value as bool.
func (o *Valuer) Float() (v float64) {
	o.kf.Get(o.group, o.name, &v)
	return
}

// String returns the value as bool.
func (o *Valuer) String() (v string) {
	o.kf.Get(o.group, o.name, &v)
	return
}

// ListBool returns the value as bool.
func (o *Valuer) ListBool() (v []bool) {
	o.kf.Get(o.group, o.name, &v)
	return
}

// ListInt returns the value as bool.
func (o *Valuer) ListInt() (v []int) {
	o.kf.Get(o.group, o.name, &v)
	return
}

// ListFloat returns the value as bool.
func (o *Valuer) ListFloat() (v []float64) {
	o.kf.Get(o.group, o.name, &v)
	return
}

// ListString returns the value as bool.
func (o *Valuer) ListString() (v []string) {
	o.kf.Get(o.group, o.name, &v)
	return
}

// Set sets the pointed keyfile key value.
func (o *Valuer) Set(v interface{}) {
	o.kf.Set(o.group, o.name, v)
}

// Sprint returns the value as printable text.
func (o *Valuer) Sprint() string {
	return o.String()
}

// SprintI returns the value as printable text of the element at position I in
// the list if possible.
//
func (o *Valuer) SprintI(id int) string {
	list := o.ListString()
	if id >= len(list) {
		println("valuer SprintI. out of range:", id, list)
		return ""
	}
	return list[id]
}

// Count returns the number of elements in the list.
//
func (o *Valuer) Count() int { return len(o.ListString()) } // unsure.

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
