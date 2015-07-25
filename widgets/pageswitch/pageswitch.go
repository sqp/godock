// Package pageswitch is a custom page switcher with a buttons box.
package pageswitch

import (
	"github.com/conformal/gotk3/glib"
	"github.com/conformal/gotk3/gtk"
)

//-------------------------------------------------[  ]--

// Page defines a switcher page.
//
type Page struct {
	Name    string
	Icon    string
	OnLoad  func()
	OnShow  func()
	OnHide  func()
	OnClear func()
	Widget  gtk.IWidget

	// internal
	btn     *gtk.ToggleButton
	handler glib.SignalHandle
}

// Switcher is a custom page switcher with a buttons box.
//
type Switcher struct {
	gtk.Box

	current string
	pages   map[string]*Page
}

// New creates a switcher box to handle page switching.
//
func New() *Switcher {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)

	widget := &Switcher{
		Box:   *box,
		pages: make(map[string]*Page),
	}
	return widget
}

// AddPage connects a page to a new button.
//
func (widget *Switcher) AddPage(page *Page) {
	page.btn, _ = gtk.ToggleButtonNewWithLabel(page.Name)
	widget.PackStart(page.btn, false, false, 0)
	page.btn.Show()
	page.handler, _ = page.btn.Connect("clicked", func() { widget.clickedBtn(page.Name) })

	widget.pages[page.Name] = page
}

// Activate selects a page.
//
func (widget *Switcher) Activate(name string) {
	widget.clickedBtn(name)
	widget.pages[name].btn.SetActive(true)
}

// Load loads data in all pages.
//
func (widget *Switcher) Load() {
	for _, page := range widget.pages {
		if page.OnLoad != nil {
			page.OnLoad()
		}
	}
}

// Selected returns the name of the selected page.
//
func (widget *Switcher) Selected() string {
	return widget.current
}

// Clear resets the switcher and all its pages.
//
func (widget *Switcher) Clear() {
	for _, page := range widget.pages {
		if page.OnClear != nil {
			page.OnClear()
		}
		page.btn.Destroy()
	}
	widget.pages = make(map[string]*Page)
	widget.current = ""
}

// ReloadCurrent forces a data reload on the current page.
//
func (widget *Switcher) ReloadCurrent() {
	if current, ok := widget.pages[widget.current]; ok {
		current.OnLoad()
	}
}

//-----------------------------------------------------[ INTERFACE CALLBACKS ]--

func (widget *Switcher) clickedBtn(name string) {
	if name == widget.current { // Same reclicked. Reselect.
		info := widget.pages[name]
		info.btn.HandlerBlock(info.handler)
		info.btn.SetActive(true)
		info.btn.HandlerUnblock(info.handler)
		return
	}

	for btnName, info := range widget.pages {
		if btnName == widget.current { // Old one. Deselect button.
			info.btn.HandlerBlock(info.handler)
			info.btn.SetActive(false)
			info.btn.HandlerUnblock(info.handler)
			if info.OnHide != nil {
				info.OnHide()
			}
		}
	}

	// 2nd loop to be sure we show the new one after the hide.
	for btnName, info := range widget.pages {
		if btnName == name { // New one. Show widget.
			info.OnShow()
		}
	}
	widget.current = name
}
