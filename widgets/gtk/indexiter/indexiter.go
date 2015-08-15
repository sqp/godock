package indexiter

import (
	"github.com/gotk3/gotk3/gtk"

	"github.com/sqp/godock/libs/cdtype" // Logger type.
)

//
//-------------------------------------------------------------------[ BY ID ]--

// ByID defines a simple list of gtk iter indexed by the position they were added.
// Use with a list of your useful data referenced in the same order.
//
type ByID interface {
	Append(iter *gtk.TreeIter)
	FindIter(int) (*gtk.TreeIter, bool)
	SetActive(int) bool
	Log() cdtype.Logger
}

// Widget defines a gtk widget with the SetActiveIter method.
type Widget interface {
	SetActiveIter(*gtk.TreeIter)
}

// idxID stores a list of iter to find the one to reselect.
//
type idxID struct {
	widget Widget
	list   []*gtk.TreeIter
	log    cdtype.Logger
}

// NewByID creates a list of gtk iter indexed by their position.
//
func NewByID(widget Widget, log cdtype.Logger) ByID {
	return &idxID{
		widget: widget,
		log:    log,
	}
}

func (o *idxID) Append(iter *gtk.TreeIter) { o.list = append(o.list, iter) }
func (o *idxID) Log() cdtype.Logger        { return o.log }

func (o *idxID) SetActive(id int) bool {
	iter, ok := o.FindIter(id)
	o.widget.SetActiveIter(iter)
	return ok
}

func (o *idxID) FindIter(id int) (*gtk.TreeIter, bool) {
	if id >= len(o.list) {
		o.log.NewErrf("out of range", "indexiter FindIter=%d size=%d", id, len(o.list))
		return nil, false
	}
	return o.list[id], true
}

//
//---------------------------------------------------------------[ BY STRING ]--

// ByString creates a list of gtk iter indexed by a key string.
//
type ByString interface {
	Append(*gtk.TreeIter, string)
	FindID(string) (int, bool)
	SetActive(string) bool
}

// idxStr stores a list of strings to find the one matching the iter to reselect.
//
type idxStr struct {
	ByID // extends the real iter index.

	list []string
}

// NewByString creates a list of gtk iter indexed by their key string.
//
func NewByString(widget Widget, log cdtype.Logger) ByString {
	return &idxStr{
		ByID: NewByID(widget, log),
		// model:  model,
	}
}

func (o *idxStr) Append(iter *gtk.TreeIter, str string) {
	o.list = append(o.list, str)
	o.ByID.Append(iter)
}

func (o *idxStr) FindID(search string) (int, bool) {
	for i, str := range o.list {
		if str == search {
			return i, true
		}
	}
	o.Log().NewErrf("out of range", "indexiter FindID=%s", search)
	return 0, false
}

func (o *idxStr) SetActive(str string) bool {
	id, ok := o.FindID(str)
	if ok {
		o.ByID.SetActive(id)
	}
	return ok
}
