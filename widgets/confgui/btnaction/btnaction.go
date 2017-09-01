// Package btnaction provides a button with different actions texts.
//
// You can use different btnaction to act on the same button and reapply
// different button state when switching between different pages.
//
package btnaction

import "github.com/sqp/godock/libs/text/tran"

// Tune provides a way to adapt a button display to the current action type.
//
type Tune interface {
	// Set... sets the button state.
	//
	SetNone()   // hide the action button.
	SetSave()   // show the action button with save text.
	SetApply()  // show the action button with apply text.
	SetAdd()    // show the action button with add text.
	SetGrab()   // show the action button with grab text.
	SetCancel() // show the action button with cancel text.
	SetTest()   // show the action button with test text.
	SetDelete() // show the action button with delete text.

	// Display displays (or hides) the button as configured.
	// This will reapply the last known state.
	//
	Display()

	IWidget // Extends a gtk.Button, or other gtk.IWidget with SetLabel.
}

// IWidget defines methods needed to act on the button.
//
type IWidget interface {
	SetLabel(string)
	Show()
	Hide()
}

//
//----------------------------------------------------------[ IMPLEMENTATION ]--

type btnAction struct {
	IWidget        // Extends a gtk.Button
	display func() // display callback, to reset the button to its last state.
}

// New creates a button with different actions texts.
//
func New(widget IWidget) Tune {
	return &btnAction{
		IWidget: widget,
		display: widget.Hide,
	}
}

func (o *btnAction) Display() { o.display() }

func (o *btnAction) SetNone() {
	o.display = o.Hide
	o.Display()
}

func (o *btnAction) setShow(label string) {
	o.display = func() { o.SetLabel(label); o.Show() }
	o.Display()
}

func (o *btnAction) SetSave()   { o.setShow(tran.Slate("Save")) }
func (o *btnAction) SetApply()  { o.setShow(tran.Slate("Apply")) }
func (o *btnAction) SetAdd()    { o.setShow(tran.Slate("Add")) }
func (o *btnAction) SetCancel() { o.setShow(tran.Slate("Cancel")) }
func (o *btnAction) SetDelete() { o.setShow(tran.Slate("Delete")) }
func (o *btnAction) SetGrab()   { o.setShow(tran.Slate("Grab")) }
func (o *btnAction) SetTest()   { o.setShow(tran.Slate("Test")) }
