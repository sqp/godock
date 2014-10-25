// Package buildHelper is a small wrapper around gtk.Builder to load interfaces easily.
package buildHelper

import (
	"github.com/conformal/gotk3/gtk"

	"errors"
)

//
//-------------------------------------------------------------[ BUILDHELPER ]--

// BuildHelper is a small wrapper around gtk.Builder to load interfaces easily.
//
type BuildHelper struct {
	gtk.Builder
	Errors []error
}

// New creates a *BuildHelper to load gtk.Builder interfaces easily.
//
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

// GetAdjustment get the object named name as Adjustment.
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

// GetBox get the object named name as Box.
//
func (b *BuildHelper) GetBox(name string) *gtk.Box {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.Box); ok {
			return widget
		}
		b.badtype(name, "Box")
	}
	return nil
}

// GetButton get the object named name as Button.
//
func (b *BuildHelper) GetButton(name string) *gtk.Button {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.Button); ok {
			return widget
		}
		b.badtype(name, "Button")
	}
	return nil
}

// GetCheckButton get the object named name as CheckButton.
//
func (b *BuildHelper) GetCheckButton(name string) *gtk.CheckButton {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.CheckButton); ok {
			return widget
		}
		b.badtype(name, "CheckButton")
	}
	return nil
}

// GetComboBox get the object named name as ComboBox.
//
func (b *BuildHelper) GetComboBox(name string) *gtk.ComboBox {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.ComboBox); ok {
			return widget
		}
		b.badtype(name, "ComboBox")
	}
	return nil
}

// GetFrame get the object named name as Frame.
//
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

// GetImage get the object named name as Image.
//
func (b *BuildHelper) GetImage(name string) *gtk.Image {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.Image); ok {
			return widget
		}
		b.badtype(name, "Image")
	}
	return nil
}

// GetLabel get the object named name as Label.
//
func (b *BuildHelper) GetLabel(name string) *gtk.Label {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.Label); ok {
			return widget
		}
		b.badtype(name, "Label")
	}
	return nil
}

// GetListStore get the object named name as ListStore.
//
func (b *BuildHelper) GetListStore(name string) *gtk.ListStore {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.ListStore); ok {
			return widget
		}
		b.badtype(name, "ListStore")
	}
	return nil
}

// GetScrolledWindow get the object named name as ScrolledWindow.
//
func (b *BuildHelper) GetScrolledWindow(name string) *gtk.ScrolledWindow {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.ScrolledWindow); ok {
			return widget
		}
		b.badtype(name, "ScrolledWindow")
	}
	return nil
}

// GetScale get the object named name as Scale.
//
func (b *BuildHelper) GetScale(name string) *gtk.Scale {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.Scale); ok {
			return widget
		}
		b.badtype(name, "Scale")
	}
	return nil
}

// GetSwitch get the object named name as Switch.
//
func (b *BuildHelper) GetSwitch(name string) *gtk.Switch {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.Switch); ok {
			return widget
		}
		b.badtype(name, "Switch")
	}
	return nil
}

// GetToggleButton get the object named name as ToggleButton.
//
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

// GetTreeSelection get the object named name as TreeSelection.
//
func (b *BuildHelper) GetTreeSelection(name string) *gtk.TreeSelection {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.TreeSelection); ok {
			return widget
		}
		b.badtype(name, "TreeSelection")
	}
	return nil
}

// GetTreeStore get the object named name as TreeStore.
//
func (b *BuildHelper) GetTreeStore(name string) *gtk.TreeStore {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.TreeStore); ok {
			return widget
		}
		b.badtype(name, "TreeStore")
	}
	return nil
}

// GetTreeView get the object named name as TreeView.
//
func (b *BuildHelper) GetTreeView(name string) *gtk.TreeView {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.TreeView); ok {
			return widget
		}
		b.badtype(name, "TreeView")
	}
	return nil
}

// GetTreeViewColumn get the object named name as TreeViewColumn.
//
func (b *BuildHelper) GetTreeViewColumn(name string) *gtk.TreeViewColumn {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.TreeViewColumn); ok {
			return widget
		}
		b.badtype(name, "TreeViewColumn")
	}
	return nil
}
