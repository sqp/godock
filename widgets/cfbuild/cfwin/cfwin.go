// Package cfwin creates a dedicated config builder window.
package cfwin

import (
	"github.com/gotk3/gotk3/gtk"

	"github.com/sqp/godock/libs/cdtype"               // Logger type.
	"github.com/sqp/godock/widgets/cfbuild"           // The config file builder.
	"github.com/sqp/godock/widgets/cfbuild/cftype"    // Types for config file builder usage.
	"github.com/sqp/godock/widgets/cfbuild/vdata"     // Virtual data source.
	"github.com/sqp/godock/widgets/confgui/btnaction" // Set save button state.
	"github.com/sqp/godock/widgets/confmenu"          // Window menu bar with switcher and buttons.
	"github.com/sqp/godock/widgets/gtk/newgtk"        // Create widgets.
	"github.com/sqp/godock/widgets/pageswitch"        // Switcher for config pages.
)

// Win defines a config builder wrapper.
//
type Win struct {
	cftype.Grouper
}

// New creates a dedicated config builder application with window.
// Returns the builder and its init func.
//
func New(logger cdtype.Logger, onBuild, onSave func(cftype.Builder)) (*Win, func(win *gtk.ApplicationWindow)) {
	b := &Win{}
	return b, func(win *gtk.ApplicationWindow) {
		source := vdata.New(logger, win, onSave)
		switcher := pageswitch.New()
		b.Grouper = cfbuild.NewVirtual(source, logger, "", "", "")

		b.Grouper = b.Grouper.BuildAll(switcher, onBuild)
		b.Grouper.ShowAll()
		source.SetGrouper(b.Grouper)
		packWindow(win, source, b.Grouper, switcher)
	}
}

func packWindow(win *gtk.ApplicationWindow, source vdata.Sourcer, build cftype.Grouper, switcher *pageswitch.Switcher) {
	win.SetIconFromFile(source.AppIcon())

	// Widgets.
	box := newgtk.Box(gtk.ORIENTATION_VERTICAL, 0)
	source.SetBox(box)

	menu := confmenu.New(source)
	btnAdd := btnaction.New(menu.Save)
	btnAdd.SetDelete()

	// Packing.
	menu.PackStart(switcher, false, false, 0)
	win.Add(box)
	box.PackStart(menu, false, false, 0)
	box.PackStart(build, true, true, 0)
	win.ShowAll()
}
