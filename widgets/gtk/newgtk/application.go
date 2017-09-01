package newgtk

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"log"
)

// AppInfo defines application settings to run a GTK application.
//
// When the object is created, start the application with:
//
//   func main() { os.Exit(appInfo.Run()) }
//
// If OnActivateWin is set, a first window will be created, and forwarded to this callback.
//
// Call order
//
// Callbacks are run in this order:
//   - OnStart
//   - OnActivateApp
//   - OnActivateWin
//       .............. (application running)
//   - OnStop
//
// See the package example.
//
type AppInfo struct {
	// Window settings.
	ID     string // Format: "org.gtk.example"
	Title  string
	Width  int
	Height int
	Flags  glib.ApplicationFlags // See flags: https://godoc.org/github.com/gotk3/gotk3/glib#ApplicationFlags

	// Application callbacks (connected to signals).
	OnStart       func(*gtk.Application)       // sets up the application when it first starts
	OnActivateApp func(*gtk.Application)       // shows the default first window of the application (like a new document). This corresponds to the application being launched by the desktop environment.
	OnActivateWin func(*gtk.ApplicationWindow) // for the first window, created on OnActivateApp
	OnStop        func(*gtk.Application)
	// OnOpen        func(app *gtk.Application, files unsafe.Pointer, hint string, test string) // opens files and shows them in a new window. This corresponds to someone trying to open a document (or documents) using the application from the file browser, or similar.

	// Pointers, filled between OnStart and OnActivateApp.
	App *gtk.Application       // Set before OnActivateApp
	Win *gtk.ApplicationWindow // Set before OnActivateWin. Only set if OnActivateWin is defined.
}

// Run starts the application and creates the window if needed.
// Locks the thread until application release (windows closed?).
// Returns an error code.
func (appInfo *AppInfo) Run() int {
	e := Application(appInfo)
	if e != nil {
		println(e)
		return 1
	}

	return appInfo.App.Run(nil)
}

// Application creates a *gtk.Application and connects its callbacks.
func Application(appInfo *AppInfo) error {
	var e error
	appInfo.App, e = gtk.ApplicationNew(appInfo.ID, appInfo.Flags)
	if e != nil {
		return e
	}

	onActivate := func(app *gtk.Application) {
		appInfo.App = app
		if appInfo.OnActivateApp != nil {
			appInfo.OnActivateApp(appInfo.App)
		}

		// Only create a window if there is something to put inside.
		if appInfo.OnActivateWin == nil {
			return
		}

		// Create ApplicationWindow
		appInfo.Win, e = gtk.ApplicationWindowNew(appInfo.App)
		if e != nil {
			// TODO IMPROVE
			log.Fatal("Could not create application window:", e)
		}

		// Set ApplicationWindow Properties
		appInfo.Win.SetTitle(appInfo.Title)
		appInfo.Win.SetDefaultSize(appInfo.Width, appInfo.Height)

		// Custom user call.
		appInfo.OnActivateWin(appInfo.Win)

		appInfo.Win.Show()
	}

	calls := map[string]interface{}{
		"activate": onActivate,
	}

	if appInfo.OnStart != nil {
		calls["startup"] = appInfo.OnStart
	}

	if appInfo.OnStop != nil {
		calls["shutdown"] = appInfo.OnStop
	}

	for signal, call := range calls {
		_, e = appInfo.App.Connect(signal, call)
		if e != nil {
			return e
		}
	}

	return nil
}
