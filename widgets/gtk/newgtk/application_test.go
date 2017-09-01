///bin/true; exec /usr/bin/env go run "$0" "$@"

// Shebang to directly run as script on unix like.

package newgtk_test

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"github.com/sqp/godock/widgets/gtk/newgtk" // Create gtk widgets.

	"fmt"
	"os"
	"time"
)

// How to create a simple Gtk3 Application in go.

// Define your application informations.
//
var appInfo = newgtk.AppInfo{
	ID:            "org.gtk.example",
	Title:         "Basic Application",
	Width:         400,
	Height:        400,
	Flags:         glib.APPLICATION_FLAGS_NONE, // See flags: https://godoc.org/github.com/gotk3/gotk3/glib#ApplicationFlags
	OnStart:       onStart,                     // test
	OnActivateApp: onActivateApp,               // test
	OnStop:        onStop,                      // test
}

func init() { appInfo.OnActivateWin = onActivateWin } // set here only to prevents initialization loop with our test.

// Start application and return error code if any.
//
func main() { os.Exit(appInfo.Run()) }

// Fill the window at creation.
//
// Move this and everything else in other packages.
// This help the main package stay clean and not requiring impossible tests.
//
func onActivateWin(win *gtk.ApplicationWindow) {
	win.Add(newgtk.Label("Hello, gotk3!"))
	win.ShowAll() // Don't forget to show your widgets. Most common problem for GTK beginners

	go time.AfterFunc(3*time.Second, win.Close) // Autoclose window for the test.
	title, _ := win.GetTitle()
	fmt.Println("win.Title      :", title)
	fmt.Println("win.Size       :", win.GetAllocatedWidth(), "x", win.GetAllocatedHeight())
	fmt.Println("app.App != nil :", appInfo.App != nil)
	fmt.Println("app.Win != nil :", appInfo.Win != nil)
}

// test
func onStart(app *gtk.Application)       { fmt.Println("--[ started ]--") }
func onActivateApp(app *gtk.Application) { fmt.Println("app.ID         :", app.GetApplicationID()) }
func onStop(app *gtk.Application)        { fmt.Print("exit code      : ") }

func Example() { // Auto tested example.
	fmt.Println(appInfo.Run())
	// Output:
	// --[ started ]--
	// app.ID         : org.gtk.example
	// win.Title      : Basic Application
	// win.Size       : 400 x 400
	// app.App != nil : true
	// app.Win != nil : true
	// exit code      : 0
}
