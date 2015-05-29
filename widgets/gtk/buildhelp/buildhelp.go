// Package buildhelp is a small wrapper around gtk.Builder to load interfaces easily.
package buildhelp

import (
	"github.com/conformal/gotk3/gtk"

	"errors"
)

//
//-------------------------------------------------------------[ BUILDHELPER ]--

// BuildHelp is a small wrapper around gtk.Builder to load interfaces easily.
//
type BuildHelp struct {
	gtk.Builder
	Errors []error
}

// New creates a *BuildHelp to load gtk.Builder interfaces easily.
//
func New() *BuildHelp {
	builder, _ := gtk.BuilderNew()
	return &BuildHelp{Builder: *builder}
}

// NewFromBytes creates a *BuildHelp to load gtk.Builder interfaces easily from a slice of bytes.
//
func NewFromBytes(bytes []byte) *BuildHelp {
	builder, _ := gtk.BuilderNew()
	builder.AddFromString(string(bytes))
	return &BuildHelp{Builder: *builder}
}

func (b *BuildHelp) err(e error) bool {
	if e != nil {
		b.Errors = append(b.Errors, e)
		return true
	}
	return false
}

func (b *BuildHelp) badtype(name string, typ string) {
	b.Errors = append(b.Errors, errors.New("builder bad type for key "+name+" not a "+typ))
}

// GetAdjustment get the object named name as Adjustment.
//
func (b *BuildHelp) GetAdjustment(name string) *gtk.Adjustment {
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
func (b *BuildHelp) GetBox(name string) *gtk.Box {
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
func (b *BuildHelp) GetButton(name string) *gtk.Button {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.Button); ok {
			return widget
		}
		b.badtype(name, "Button")
	}
	return nil
}

// GetCellRendererText get the object named name as CellRendererText.
//
func (b *BuildHelp) GetCellRendererText(name string) *gtk.CellRendererText {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.CellRendererText); ok {
			return widget
		}
		b.badtype(name, "CellRendererText")
	}
	return nil
}

// GetCheckButton get the object named name as CheckButton.
//
func (b *BuildHelp) GetCheckButton(name string) *gtk.CheckButton {
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
func (b *BuildHelp) GetComboBox(name string) *gtk.ComboBox {
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
func (b *BuildHelp) GetFrame(name string) *gtk.Frame {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.Frame); ok {
			return widget
		}
		b.badtype(name, "Frame")
	}
	return nil
}

// func (b *BuildHelp) GetIconView(name string) *gtk.IconView {
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
func (b *BuildHelp) GetImage(name string) *gtk.Image {
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
func (b *BuildHelp) GetLabel(name string) *gtk.Label {
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
func (b *BuildHelp) GetListStore(name string) *gtk.ListStore {
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
func (b *BuildHelp) GetScrolledWindow(name string) *gtk.ScrolledWindow {
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
func (b *BuildHelp) GetScale(name string) *gtk.Scale {
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
func (b *BuildHelp) GetSwitch(name string) *gtk.Switch {
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
func (b *BuildHelp) GetToggleButton(name string) *gtk.ToggleButton {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.ToggleButton); ok {
			return widget
		}
		b.badtype(name, "ToggleButton")
	}
	return nil
}

// func (b *BuildHelp) GetTreeModelFilter(name string) *gtk.TreeModelFilter {
// 	if obj, e := b.GetObject(name); !b.err(e) {
// 		if widget, ok := obj.(*gtk.TreeModelFilter); ok {
// 			return widget
// 		}
// 		b.badtype(name, "TreeModelFilter")
// 	}
// 	return nil
// }

// func (b *BuildHelp) GetTreeModelSort(name string) *gtk.TreeModelSort {
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
func (b *BuildHelp) GetTreeSelection(name string) *gtk.TreeSelection {
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
func (b *BuildHelp) GetTreeStore(name string) *gtk.TreeStore {
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
func (b *BuildHelp) GetTreeView(name string) *gtk.TreeView {
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
func (b *BuildHelp) GetTreeViewColumn(name string) *gtk.TreeViewColumn {
	if obj, e := b.GetObject(name); !b.err(e) {
		if widget, ok := obj.(*gtk.TreeViewColumn); ok {
			return widget
		}
		b.badtype(name, "TreeViewColumn")
	}
	return nil
}
