package docklist

/*
#cgo pkg-config: glib-2.0 gldi
#include <stdlib.h>		// free

#include "cairo-dock-animations.h"			// cairo_dock_foreach_animation
#include "cairo-dock-backends-manager.h"	// cairo_dock_foreach_dock_renderer
#include "cairo-dock-desklet-manager.h"		// cairo_dock_foreach_desklet_decoration
#include "cairo-dock-dock-facility.h"		// cairo_dock_foreach_dock_renderer
#include "cairo-dock-keybinder.h"			// gldi_shortkeys_foreach
#include "cairo-dock-module-manager.h"		// gldi_module_foreach

//
//--------------------------------------------------------[ LISTS FORWARDING ]--

extern void addItemToList (gpointer, gpointer);
extern void addItemToMap (gpointer, gchar*, gpointer);

static void     fwd_one (const gchar *name, gpointer *item, gpointer p) { addItemToMap(p, g_strdup(name), item); }
static gboolean fwd_chk (      gchar *name, gpointer *item, gpointer p) { addItemToMap(p, name,           item); return FALSE;}

static void list_shortkey           (gpointer p){ gldi_shortkeys_foreach(                (GFunc)addItemToList, p); }
static void list_desklets           (gpointer p){ gldi_desklets_foreach(                 (GldiDeskletForeachFunc)addItemToList, p); }

static void list_animation          (gpointer p){ cairo_dock_foreach_animation(          (GHFunc)fwd_one, p); }
static void list_desklet_decoration (gpointer p){ cairo_dock_foreach_desklet_decoration( (GHFunc)fwd_one, p); }
static void list_dialog_decorator   (gpointer p){ cairo_dock_foreach_dialog_decorator(   (GHFunc)fwd_one, p); }
static void list_dock_renderer      (gpointer p){ cairo_dock_foreach_dock_renderer(      (GHFunc)fwd_one, p); }

static void list_dock_module        (gpointer p){ gldi_module_foreach                   ((GHRFunc)fwd_chk, p); }

*/
import "C"

import (
	"sync"
	"unsafe"

	"github.com/sqp/godock/libs/cdglobal" // Dock types.
	"github.com/sqp/godock/libs/gldi"
	"github.com/sqp/godock/libs/gldi/shortkeys" // Keyboard shortkeys.
)

//
//------------------------------------------------------------------[ PUBLIC ]--

// Desklet returns the list of active desklets.
//
func Desklet() []*gldi.Desklet {
	return newItemList(itemListTypeDesklet).Value().([]*gldi.Desklet)
}

// Shortkey returns the list of dock shortkeys.
//
func Shortkey() []cdglobal.Shortkeyer {
	return newItemList(itemListTypeShortkey).Value().([]cdglobal.Shortkeyer)
}

// Animation returns the list of dock animations.
//
func Animation() map[string]*gldi.Animation {
	return newItemMap(itemMapTypeAnimation).Value().(map[string]*gldi.Animation)
}

// CairoDeskletDecoration returns the list of dock desklet decorations.
//
func CairoDeskletDecoration() map[string]*gldi.CairoDeskletDecoration {
	return newItemMap(itemMapTypeCairoDeskletDecoration).Value().(map[string]*gldi.CairoDeskletDecoration)
}

// CairoDockRenderer returns the list of dock renderers.
//
func CairoDockRenderer() map[string]*gldi.CairoDockRenderer {
	return newItemMap(itemMapTypeCairoDockRenderer).Value().(map[string]*gldi.CairoDockRenderer)
}

// DialogDecorator returns the list of dock dialog decorators.
//
func DialogDecorator() map[string]*gldi.DialogDecorator {
	return newItemMap(itemMapTypeDialogDecorator).Value().(map[string]*gldi.DialogDecorator)
}

// Module returns the list of dock modules.
//
func Module() map[string]*gldi.Module {
	return newItemMap(itemMapTypeModule).Value().(map[string]*gldi.Module)
}

//
//-------------------------------------------------------------[ C CALLBACKS ]--

//export addItemToMap
func addItemToMap(cid C.gpointer, cstr *C.gchar, cdr C.gpointer) {
	id := int(uintptr(cid))
	itemMapStatic[id].Add(cstr, cdr)
}

//export addItemToList
func addItemToList(citem C.gpointer, cid C.gpointer) {
	id := int(uintptr(cid))
	itemListStatic[id].Add(citem)
}

//
//-----------------------------------------------------------------[ ITEMMAP ]--

var (
	itemMapStatic = make(map[int]*itemMap)
	itemMapMutex  = &sync.Mutex{}
)

// itemMapType defines usable types for itemMap.
type itemMapType int

const (
	itemMapTypeAnimation itemMapType = iota
	itemMapTypeCairoDeskletDecoration
	itemMapTypeCairoDockRenderer
	itemMapTypeDialogDecorator
	itemMapTypeModule
)

// itemMap stores dock items references in a map from a C callback.
//
type itemMap struct {
	ID      int
	Release func()
	typ     itemMapType
	list    map[string]interface{}
}

// newItemMap creates an item map to collect items references from C callbacks.
//
func newItemMap(typ itemMapType) *itemMap {
	max := 0
	itemMapMutex.Lock()
	for k := range itemMapStatic {
		if k > max {
			max = k
		}
	}

	id := max + 1

	itemMapStatic[id] = &itemMap{
		ID:   id,
		list: make(map[string]interface{}),
		typ:  typ,
		Release: func() {
			itemMapMutex.Lock()
			delete(itemMapStatic, id)
			itemMapMutex.Unlock()
		},
	}
	itemMapMutex.Unlock()

	return itemMapStatic[id]
}

// Add adds an item to the map (from C callback).
//
func (im *itemMap) Add(cstr *C.gchar, cdr C.gpointer) {
	name := C.GoString((*C.char)(cstr))
	free := true

	switch im.typ {
	case itemMapTypeAnimation:
		im.list[name] = gldi.NewAnimationFromNative(unsafe.Pointer(cdr))

	case itemMapTypeCairoDeskletDecoration:
		im.list[name] = gldi.NewCairoDeskletDecorationFromNative(unsafe.Pointer(cdr))

	case itemMapTypeCairoDockRenderer:
		im.list[name] = gldi.NewCairoDockRendererFromNative(unsafe.Pointer(cdr))

	case itemMapTypeDialogDecorator:
		im.list[name] = gldi.NewDialogDecoratorFromNative(unsafe.Pointer(cdr))

	case itemMapTypeModule:
		im.list[name] = gldi.NewModuleFromNative(unsafe.Pointer(cdr))
		free = false
	}

	if free {
		C.free(unsafe.Pointer((*C.char)(cstr)))
	}
}

// Value release the index reference and returns the list content.
// You just have to assert to a map[string]*ExpectedType
//
func (im *itemMap) Value() interface{} {
	defer im.Release()
	switch im.typ {
	case itemMapTypeAnimation:
		C.list_animation(cIntPointer(im.ID))

		ret := make(map[string]*gldi.Animation)
		for name, mod := range im.list {
			ret[name] = mod.(*gldi.Animation)
		}
		return ret
	// return im.Value().(map[string]*gldi.Animation)

	case itemMapTypeCairoDeskletDecoration:
		C.list_desklet_decoration(cIntPointer(im.ID))

		ret := make(map[string]*gldi.CairoDeskletDecoration)
		for name, mod := range im.list {
			ret[name] = mod.(*gldi.CairoDeskletDecoration)
		}
		return ret

	case itemMapTypeCairoDockRenderer:
		C.list_dock_renderer(cIntPointer(im.ID))

		ret := make(map[string]*gldi.CairoDockRenderer)
		for name, mod := range im.list {
			ret[name] = mod.(*gldi.CairoDockRenderer)
		}
		return ret

	case itemMapTypeDialogDecorator:
		C.list_dialog_decorator(cIntPointer(im.ID))

		ret := make(map[string]*gldi.DialogDecorator)
		for name, mod := range im.list {
			ret[name] = mod.(*gldi.DialogDecorator)
		}
		return ret

	case itemMapTypeModule:
		C.list_dock_module(cIntPointer(im.ID))

		ret := make(map[string]*gldi.Module)
		for name, mod := range im.list {
			ret[name] = mod.(*gldi.Module)
		}
		return ret
	}
	return nil
}

//
//----------------------------------------------------------------[ ITEMLIST ]--

var (
	itemListStatic = make(map[int]*itemList)
	itemListMutex  = &sync.Mutex{}
)

// itemListType defines usable types for itemList.
type itemListType int

const (
	itemListTypeDesklet itemListType = iota
	itemListTypeShortkey
)

// itemList stores dock items references in a slice from a C callback.
//
type itemList struct {
	ID      int
	Release func()
	typ     itemListType
	list    []interface{}
}

// newItemList creates an item list to collect items references from C callbacks.
//
func newItemList(typ itemListType) *itemList {
	max := 0
	itemListMutex.Lock()
	for k := range itemListStatic {
		if k > max {
			max = k
		}
	}

	id := max + 1

	itemListStatic[id] = &itemList{
		ID:  id,
		typ: typ,
		Release: func() {
			itemListMutex.Lock()
			delete(itemListStatic, id)
			itemListMutex.Unlock()
		},
	}
	itemListMutex.Unlock()

	return itemListStatic[id]
}

// Add adds an item to the list (from C callback).
//
func (il *itemList) Add(cdr C.gpointer) {
	switch il.typ {
	case itemListTypeDesklet:
		item := gldi.NewDeskletFromNative(unsafe.Pointer(cdr))
		il.list = append(il.list, item)

	case itemListTypeShortkey:
		item := shortkeys.NewFromNative(unsafe.Pointer(cdr))
		il.list = append(il.list, item)
	}
}

// Value release the index reference and returns the list content.
// You just have to assert to a []*ExpectedType
//
func (il *itemList) Value() interface{} {
	defer il.Release()

	switch il.typ {
	case itemListTypeDesklet:
		C.list_desklets(cIntPointer(il.ID))

		var ret []*gldi.Desklet
		for _, mod := range il.list {
			ret = append(ret, mod.(*gldi.Desklet))
		}
		return ret

	case itemListTypeShortkey:
		C.list_shortkey(cIntPointer(il.ID))

		var ret []cdglobal.Shortkeyer
		for _, mod := range il.list {
			ret = append(ret, mod.(cdglobal.Shortkeyer))
		}
		return ret
	}
	return nil
}

// cIntPointer returns a C gpointer to an int value.
//
func cIntPointer(i int) C.gpointer {
	return C.gpointer(gldi.CIntPointer(i))
}
