// Package keyfile provides Go bindings for GLib's GKeyFile.
package keyfile

// #cgo pkg-config: glib-2.0
// #include "glib.gen.h"
import "C"

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
type Error C.uint32_t

const (
	ErrorUnknownEncoding Error = 0
	ErrorParse           Error = 1
	ErrorNotFound        Error = 2
	ErrorKeyNotFound     Error = 3
	ErrorGroupNotFound   Error = 4
	ErrorInvalidValue    Error = 5
)

// Flags is a representation of GLib's KeyFileFlags.
type Flags C.uint32_t

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
	return &KeyFile{C.g_key_file_new()}
}

// LoadFromFile is a wrapper around g_key_file_load_from_file().
func (this *KeyFile) LoadFromFile(file string, flags Flags) (bool, error) {
	var cstr *C.char = C.CString(file)
	defer C.free(unsafe.Pointer(cstr))
	var cerr *C.GError
	c := C.g_key_file_load_from_file(this.cKey, cstr, C.GKeyFileFlags(flags), &cerr)
	return c != 0, goError(cerr)
}

// ToData is a wrapper around g_key_file_to_data().
func (this *KeyFile) ToData() (uint64, string, error) {
	var clength C.uint64_t
	var cerr *C.GError
	c := C.g_key_file_to_data(this.cKey, &clength, &cerr)
	defer C.g_free(unsafe.Pointer(c))
	return uint64(clength), C.GoString(c), goError(cerr)
}

// ToNative returns the pointer to the underlying GKeyFile.
func (this *KeyFile) ToNative() *C.GKeyFile {
	return this.cKey
}

//
//---------------------------------------------------------------------[ GET ]--

// GetGroups is a wrapper around g_key_file_get_groups().
func (this *KeyFile) GetGroups() (uint64, []string) {
	var length C.uint64_t
	c := C.g_key_file_get_groups(this.cKey, &length)
	var list []string
	for i := 0; i < int(length); i++ {
		if str := C.GoString((*(*[999999]*C.char)(unsafe.Pointer(c)))[i]); str != "" {
			list = append(list, str)
		} else {
			println("-----------------------------------------------dropped empty group")
		}
		C.g_free(unsafe.Pointer((*(*[999999]*C.char)(unsafe.Pointer(c)))[i]))
	}
	return uint64(length), list
}

// GetKeys is a wrapper around g_key_file_get_keys().
func (this *KeyFile) GetKeys(group string) (uint64, []string, error) {
	var cGroup *C.char = C.CString(group)
	var length C.uint64_t
	var cErr *C.GError
	defer C.free(unsafe.Pointer(cGroup))
	c := C.g_key_file_get_keys(this.cKey, cGroup, &length, &cErr)
	list := make([]string, uint64(length))
	for i := range list {
		list[i] = C.GoString((*(*[999999]*C.char)(unsafe.Pointer(c)))[i])
		C.g_free(unsafe.Pointer((*(*[999999]*C.char)(unsafe.Pointer(c)))[i]))
	}
	return uint64(length), list, goError(cErr)
}

// GetBooleanList is a wrapper around g_key_file_get_boolean_list().
func (this *KeyFile) GetBooleanList(group string, key string) (length uint64, list []bool, e error) {
	var c unsafe.Pointer
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
	var c unsafe.Pointer
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
	var c unsafe.Pointer
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
	var c unsafe.Pointer
	length, c, e = this.getList(group, key, "string")
	defer C.g_free(c)
	list = make([]string, length)
	for i := range list {
		list[i] = C.GoString((*(*[999999]*C.char)(c))[i])
	}
	return length, list, e
}

func (this *KeyFile) getList(group, key, typ string) (uint64, unsafe.Pointer, error) {
	var cGroup *C.char = C.CString(group)
	defer C.free(unsafe.Pointer(cGroup))
	var cKey *C.char = C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	var length1 C.uint64_t
	var cErr *C.GError
	var c unsafe.Pointer
	switch typ {
	case "bool":
		c = unsafe.Pointer(C.g_key_file_get_boolean_list(this.cKey, cGroup, cKey, &length1, &cErr))
	case "int":
		c = unsafe.Pointer(C.g_key_file_get_integer_list(this.cKey, cGroup, cKey, &length1, &cErr))
	case "float64":
		c = unsafe.Pointer(C.g_key_file_get_double_list(this.cKey, cGroup, cKey, &length1, &cErr))
	case "string":
		c = unsafe.Pointer(C.g_key_file_get_string_list(this.cKey, cGroup, cKey, &length1, &cErr))

	}
	return uint64(length1), c, goError(cErr)
}

func (this *KeyFile) getOne(group string, key string, typ string) (interface{}, error) {
	var cGroup *C.char = C.CString(group)
	defer C.free(unsafe.Pointer(cGroup))
	var cKey *C.char = C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	var cErr *C.GError

	var c interface{}
	switch typ {
	case "int":
		c = int(C.g_key_file_get_integer(this.cKey, cGroup, cKey, &cErr))

	case "comment":
		cstr := C.g_key_file_get_comment(this.cKey, cGroup, cKey, &cErr)
		c = C.GoString(cstr)
		C.g_free(unsafe.Pointer(cstr))

	case "string":
		cstr := C.g_key_file_get_string(this.cKey, cGroup, cKey, &cErr)
		c = C.GoString(cstr)
		C.g_free(unsafe.Pointer(cstr))

	}
	return c, goError(cErr)
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
	var cGroup *C.char = C.CString(group)
	defer C.free(unsafe.Pointer(cGroup))
	var cKey *C.char = C.CString(key)
	defer C.free(unsafe.Pointer(cKey))

	switch value := uncasted.(type) {
	case bool:
		C.g_key_file_set_boolean(this.cKey, cGroup, cKey, cBool(value))

	case int:
		C.g_key_file_set_integer(this.cKey, cGroup, cKey, C.int32_t(value))

	case float64:
		C.g_key_file_set_double(this.cKey, cGroup, cKey, C.double(value))

	case string:
		cstr := C.CString(value)
		defer C.free(unsafe.Pointer(cstr))
		C.g_key_file_set_string(this.cKey, cGroup, cKey, cstr)

	case []bool:
		clist := cListBool(value)
		defer C.free(unsafe.Pointer(clist))
		C.g_key_file_set_boolean_list(this.cKey, cGroup, cKey, clist, C.uint64_t(len(value)))

	case []int:
		clist := cListInt(value)
		defer C.free(unsafe.Pointer(clist))
		C.g_key_file_set_integer_list(this.cKey, cGroup, cKey, clist, C.uint64_t(len(value)))

	case []float64:
		clist := cListDouble(value)
		defer C.free(unsafe.Pointer(clist))
		C.g_key_file_set_double_list(this.cKey, cGroup, cKey, clist, C.uint64_t(len(value)))

	case []string:
		clist := cListString(value)
		defer C.free(unsafe.Pointer(clist))
		C.g_key_file_set_string_list(this.cKey, cGroup, cKey, clist, C.uint64_t(len(value)))
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

func goError(cErr *C.GError) error {
	if cErr != nil {
		defer C.g_error_free(cErr)
		return errors.New(C.GoString(((*_GError)(unsafe.Pointer(cErr))).message))
	}
	return nil
}

type _GError struct {
	domain  uint32
	code    int32
	message *C.char
}

func cBool(x bool) C.int {
	if x {
		return 1
	}
	return 0
}

func cListBool(value []bool) *C.int {
	var clist *C.int
	clist = (*C.int)(C.malloc(C.size_t(int(unsafe.Sizeof(*clist)) * len(value))))
	for i, e := range value {
		(*(*[999999]C.int)(unsafe.Pointer(clist)))[i] = cBool(e)
	}
	return clist
}

func cListInt(value []int) *C.int32_t {
	var clist *C.int32_t
	clist = (*C.int32_t)(C.malloc(C.size_t(int(unsafe.Sizeof(*clist)) * len(value))))
	for k, v := range value {
		(*(*[999999]C.int32_t)(unsafe.Pointer(clist)))[k] = C.int32_t(v)
	}
	return clist
}

func cListDouble(value []float64) *C.double {
	var clist *C.double
	clist = (*C.double)(C.malloc(C.size_t(int(unsafe.Sizeof(*clist)) * len(value))))
	for i, e := range value {
		(*(*[999999]C.double)(unsafe.Pointer(clist)))[i] = C.double(e)
	}
	return clist
}

func cListString(value []string) **C.char {
	var clist **C.char
	clist = (**C.char)(C.malloc(C.size_t(int(unsafe.Sizeof(*clist)) * (len(value) + 1))))
	defer C.free(unsafe.Pointer(clist))
	for i, e := range value {
		(*(*[999999]*C.char)(unsafe.Pointer(clist)))[i] = C.CString(e)
		defer C.free(unsafe.Pointer((*(*[999999]*C.char)(unsafe.Pointer(clist)))[i]))
	}
	(*(*[999999]*C.char)(unsafe.Pointer(clist)))[len(value)] = nil
	return clist
}
