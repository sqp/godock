// Package pageswitch is a custom page switcher with a buttons box.
package pageswitch

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"github.com/sqp/godock/widgets/gtk/newgtk"
)

//-------------------------------------------------[  ]--

// Page defines a switcher page.
//
type Page struct {
	Key     string // key reference.
	Name    string // Visible name.
	Icon    string // Or Visible Icon. Optional, but replace text when set.
	OnLoad  func()
	OnShow  func()
	OnHide  func()
	OnClear func()
	// Widget  gtk.IWidget

	// internal
	btn     *gtk.ToggleButton // Save/apply button.
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
	box := newgtk.Box(gtk.ORIENTATION_HORIZONTAL, 0)
	context, _ := box.GetStyleContext()
	context.AddClass("linked")

	return &Switcher{
		Box:   *box,
		pages: make(map[string]*Page),
	}
}

// AddPage connects a page to a new button.
//
func (widget *Switcher) AddPage(page *Page) {
	if page.Icon != "" {
		img := newgtk.ImageFromIconName(page.Icon, gtk.ICON_SIZE_SMALL_TOOLBAR)
		if img != nil {
			page.btn = newgtk.ToggleButton()
			page.btn.SetTooltipText(page.Name)
			page.btn.SetImage(img)
		}
	}

	if page.btn == nil {
		page.btn = newgtk.ToggleButtonWithLabel(page.Name)
	}

	context, _ := page.btn.GetStyleContext()
	context.RemoveClass("text-button")

	widget.PackStart(page.btn, false, false, 0)
	page.handler, _ = page.btn.Connect("clicked", func() { widget.clickedBtn(page.Key) })

	widget.pages[page.Key] = page
	page.btn.Show()
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
