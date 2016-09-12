///bin/true; exec /usr/bin/env go run "$0" "$@"

package main

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"github.com/sqp/godock/libs/log"                  // Display info in terminal.
	"github.com/sqp/godock/widgets/cfbuild/cfprint"   // Print config file builder keys.
	"github.com/sqp/godock/widgets/cfbuild/cftype"    // Types for config file builder usage.
	"github.com/sqp/godock/widgets/cfbuild/vdata"     // Virtual data source.
	"github.com/sqp/godock/widgets/confgui/btnaction" // Set save button state.
	"github.com/sqp/godock/widgets/confmenu"          // Window menu bar with switcher and buttons.
	"github.com/sqp/godock/widgets/gtk/newgtk"        // Create widgets.
	"github.com/sqp/godock/widgets/pageswitch"        // Switcher for config pages.
)

const (
	winTitle  = "virtual config test (safe)"
	winWidth  = 700
	winHeight = 650
)

func main() {
	gtk.Init(nil)
	win, e := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if e != nil {
		println(e.Error())
		return
	}

	path, isTest := vdata.TestPathDefault()
	var saveCall func(cftype.Builder)
	if isTest {
		saveCall = cfprint.Updated
	} else {
		saveCall = func(build cftype.Builder) { cfprint.Default(build, true) }
	}
	source := vdata.New(log.NewLog(log.Logs), win, saveCall)
	build := vdata.TestInit(source, path)
	source.SetGrouper(build)

	glib.IdleAdd(packWindow(win, source, build))
	gtk.Main()
}

func packWindow(win *gtk.Window, source vdata.Sourcer, build cftype.Grouper) func() {
	return func() {
		win.SetDefaultSize(winWidth, winHeight)
		win.SetTitle(winTitle)
		win.SetIconFromFile(source.AppIcon())
		win.Connect("destroy", gtk.MainQuit)

		// widgets.
		box := newgtk.Box(gtk.ORIENTATION_VERTICAL, 0)
		source.SetBox(box)

		menu := confmenu.New(source)
		switcher := pageswitch.New()
		w := build.BuildAll(switcher)

		btnAdd := btnaction.New(menu.Save)
		btnAdd.SetTest()

		btnHack := newgtk.ButtonWithLabel("Hack")
		btnHack.Connect("clicked", func() { build.KeyWalk(vdata.TestValues) })

		menu.PackStart(switcher, false, false, 0)
		menu.PackEnd(btnHack, false, false, 0)
		win.Add(box)
		box.PackStart(menu, false, false, 0)
		box.PackStart(w, true, true, 0)
		win.ShowAll()
	}
}
