// +build gtk

package main

import (
	// "github.com/conformal/gotk3/gdk"
	// "github.com/conformal/gotk3/glib"
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/vte/vte.gtk3"

	"github.com/sqp/godock/libs/log" // Display info in terminal.
	"github.com/sqp/godock/libs/srvdbus"
	"github.com/sqp/godock/services/allapps"

	"strings"
)

const (
	WindowTitle  = "cairo-dock-console"
	WindowWidth  = 700
	WindowHeight = 400

	// to move
	TermFont = "monospace 8"
)

var window *gtk.Window

func init() {
	srvdbus.LogWindow = onOpen
	allapps.AddGtkNeeded()
}

func onOpen() {
	if window != nil {
		return
	}

	term := vte.NewTerminal()
	// term.Connect("child-exited", gtk.MainQuit)

	for _, msg := range log.Logs.List() {
		term.Feed(strings.Replace(msg.Text, "\n", "\r\n", -1))
	}
	log.Logs.SetTerminal(term)

	term.SetFontFromString(TermFont)

	// window = common.NewWindowMain(term, WindowWidth, WindowHeight)
	window, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	window.SetDefaultSize(WindowWidth, WindowHeight)
	window.Add(term)

	window.SetTitle(WindowTitle)
	// window.SetWMClass("cdc", WindowTitle)
	window.ShowAll()

	_, e := window.Connect("destroy", onDestroy)
	logger.Err(e, "connect destroy")
	// _, e = term.Connect("button-release-event", termClicked)
	// log.Err(e, "connect clicked")
}

func onDestroy() {
	log.Logs.SetTerminal(nil)
	window = nil
}

// func termClicked(widget *glib.Object, event *gdk.Event) {
// 	logger.Info("clicked", event.GetButtonID())
// }
