// Package buildHelper is a small wrapper around gtk.Builder to load interfaces easily.
package buildHelper

import (
	"github.com/conformal/gotk3/gtk"

	"errors"
)

//
//-------------------------------------------------------------[ BUILDHELPER ]--

type BuildHelper struct {
	gtk.Builder
	Errors []error
}

func New() *BuildHelper {
	builder, _ := gtk.BuilderNew()
	return &BuildHelper{Builder: *builder}
}

func (b *BuildHelper) err(e error) bool {
	if e != nil {
		b.Errors = append(b.Errors, e)
		return true
	}
	return false
}

func (b *BuildHelper) badtype(name string, typ string) {
	b.Errors = append(b.Errors, errors.New("builder bad type for key "+name+" not a "+typ))
}

//
func (b *BuildHelper) GetAdjustment(name string) *gtk.Adjustment {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.Adjustment); ok {
			return widget
		}
		b.badtype(name, "Adjustment")
	}
	return nil
}

func (b *BuildHelper) GetBox(name string) *gtk.Box {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.Box); ok {
			return widget
		}
		b.badtype(name, "Box")
	}
	return nil
}

func (b *BuildHelper) GetButton(name string) *gtk.Button {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.Button); ok {
			return widget
		}
		b.badtype(name, "Button")
	}
	return nil
}

func (b *BuildHelper) GetCheckButton(name string) *gtk.CheckButton {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.CheckButton); ok {
			return widget
		}
		b.badtype(name, "CheckButton")
	}
	return nil
}

func (b *BuildHelper) GetComboBox(name string) *gtk.ComboBox {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.ComboBox); ok {
			return widget
		}
		b.badtype(name, "ComboBox")
	}
	return nil
}

func (b *BuildHelper) GetFrame(name string) *gtk.Frame {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.Frame); ok {
			return widget
		}
		b.badtype(name, "Frame")
	}
	return nil
}

// func (b *BuildHelper) GetIconView(name string) *gtk.IconView {
// 	if obj, e := b.GetObject(name); !b.err(e) {
// 		if widget, ok := obj.(*gtk.IconView); ok {
// 			return widget
// 		}
// 		b.badtype(name, "IconView")
// 	}
// 	return nil
// }

func (b *BuildHelper) GetImage(name string) *gtk.Image {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.Image); ok {
			return widget
		}
		b.badtype(name, "Image")
	}
	return nil
}

func (b *BuildHelper) GetLabel(name string) *gtk.Label {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.Label); ok {
			return widget
		}
		b.badtype(name, "Label")
	}
	return nil
}

func (b *BuildHelper) GetListStore(name string) *gtk.ListStore {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.ListStore); ok {
			return widget
		}
		b.badtype(name, "ListStore")
	}
	return nil
}

func (b *BuildHelper) GetScrolledWindow(name string) *gtk.ScrolledWindow {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.ScrolledWindow); ok {
			return widget
		}
		b.badtype(name, "ScrolledWindow")
	}
	return nil
}

func (b *BuildHelper) GetScale(name string) *gtk.Scale {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.Scale); ok {
			return widget
		}
		b.badtype(name, "Scale")
	}
	return nil
}

func (b *BuildHelper) GetSwitch(name string) *gtk.Switch {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.Switch); ok {
			return widget
		}
		b.badtype(name, "Switch")
	}
	return nil
}

func (b *BuildHelper) GetToggleButton(name string) *gtk.ToggleButton {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.ToggleButton); ok {
			return widget
		}
		b.badtype(name, "ToggleButton")
	}
	return nil
}

// func (b *BuildHelper) GetTreeModelFilter(name string) *gtk.TreeModelFilter {
// 	if obj, e := b.GetObject(name); !b.err(e) {
// 		if widget, ok := obj.(*gtk.TreeModelFilter); ok {
// 			return widget
// 		}
// 		b.badtype(name, "TreeModelFilter")
// 	}
// 	return nil
// }

// func (b *BuildHelper) GetTreeModelSort(name string) *gtk.TreeModelSort {
// 	if obj, e := b.GetObject(name); !b.err(e) {
// 		if widget, ok := obj.(*gtk.TreeModelSort); ok {
// 			return widget
// 		}
// 		b.badtype(name, "TreeModelSort")
// 	}
// 	return nil
// }

func (b *BuildHelper) GetTreeSelection(name string) *gtk.TreeSelection {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.TreeSelection); ok {
			return widget
		}
		b.badtype(name, "TreeSelection")
	}
	return nil
}

func (b *BuildHelper) GetTreeStore(name string) *gtk.TreeStore {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.TreeStore); ok {
			return widget
		}
		b.badtype(name, "TreeStore")
	}
	return nil
}

func (b *BuildHelper) GetTreeView(name string) *gtk.TreeView {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.TreeView); ok {
			return widget
		}
		b.badtype(name, "TreeView")
	}
	return nil
}

func (b *BuildHelper) GetTreeViewColumn(name string) *gtk.TreeViewColumn {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.TreeViewColumn); ok {
			return widget
		}
		b.badtype(name, "TreeViewColumn")
	}
	return nil
}
