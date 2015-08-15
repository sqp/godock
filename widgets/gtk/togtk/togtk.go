// Package togtk provides gtk recast wrapers.
package togtk

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"unsafe"
)

//
//-------------------------------------------------------------[ GLIB RECAST ]--

// gObject recast a pointer to *glib.Object.
func gObject(ptr unsafe.Pointer) *glib.Object {
	return &glib.Object{GObject: glib.ToGObject(ptr)}
}

//
//--------------------------------------------------------------[ GTK RECAST ]--

// Container recast a pointer to *gtk.Container.
func Container(ptr unsafe.Pointer) *gtk.Container {
	return &gtk.Container{
		Widget: *Widget(ptr),
	}
}

// Menu recast a pointer to *gtk.Menu.
func Menu(ptr unsafe.Pointer) *gtk.Menu {
	return &gtk.Menu{
		MenuShell: gtk.MenuShell{
			Container: *Container(ptr),
		},
	}
}

// MenuItem recast a pointer to *gtk.MenuItem.
func MenuItem(ptr unsafe.Pointer) *gtk.MenuItem {
	return &gtk.MenuItem{
		Bin: gtk.Bin{
			Container: *Container(ptr),
		},
	}
}

// Widget recast a pointer to *gtk.Widget.
func Widget(ptr unsafe.Pointer) *gtk.Widget {
	return &gtk.Widget{
		InitiallyUnowned: glib.InitiallyUnowned{
			Object: gObject(ptr),
		},
	}
}
