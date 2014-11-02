// Package gunvalue provides some help to work with glib.Value.
package gunvalue

import (
	"github.com/conformal/gotk3/glib"
	"github.com/conformal/gotk3/gtk"

	"errors"
	"unsafe"
)

//
//----------------------------------------------------------[ ICONS BY ORDER ]--

// Conv converts a glib.Value to a usable go type.
//
type Conv struct {
	data interface{}
	err  error
}

// New converts the glib.Value func result to a usable go value.
// can be used with TreeModel.GetValue
func New(gval *glib.Value, err error) Conv {
	if err != nil {
		return Conv{err: err}
	}
	val, ego := gval.GoValue()
	if ego != nil {
		return Conv{err: ego}
	}
	return Conv{data: val}
}

// Int converts the value to an int.
//
func (c Conv) Int() (int, error) {
	if c.err != nil {
		return 0, c.err
	}
	value, ok := c.data.(int)
	if !ok {
		return 0, errors.New("convert: not int type")
	}
	return value, nil
}

// Pointer converts the value to a pointer.
//
func (c Conv) Pointer() (unsafe.Pointer, error) {
	if c.err != nil {
		return nil, c.err
	}
	value, ok := c.data.(unsafe.Pointer)
	if !ok {
		return nil, errors.New("convert: not unsafe.Pointer type")
	}
	return value, nil
}

// String converts the value to a string.
//
func (c Conv) String() (string, error) {
	if c.err != nil {
		return "", c.err
	}
	value, ok := c.data.(string)
	if !ok {
		return "", errors.New("convert: not string type")
	}
	return value, nil
}

// SelectedIter returns the iter matching the selected line.
//
func SelectedIter(model *gtk.ListStore, selection *gtk.TreeSelection) (*gtk.TreeIter, error) {
	if selection.CountSelectedRows() == 0 {
		return nil, errors.New("no line selected")
	}
	var iter gtk.TreeIter
	var treeModel gtk.ITreeModel = model
	ok := selection.GetSelected(&treeModel, &iter)
	if !ok {
		return nil, errors.New("SelectedIter: GetSelected failed")
	}
	return &iter, nil
}

// SelectedValue returns the liststore row value for the selected line as converter.

func SelectedValue(model *gtk.ListStore, selection *gtk.TreeSelection, row int) Conv {
	iter, e := SelectedIter(model, selection)
	if e != nil {
		return Conv{err: e}
	}
	return New(model.GetValue(iter, row))
}
