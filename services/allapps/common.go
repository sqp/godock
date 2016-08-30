// Package allapps declares applets available for the applet loader service.
package allapps

// Common fields filled by declared applets.
var needgtk bool // true if an applet has some GTK dependency.

// AddGtkNeeded allow an applet to declare its gtk dependency.
// If used, the gtk main loop should lock the main thread to prevent later problems.
//
func AddGtkNeeded() {
	needgtk = true
}

// GtkNeeded returns true if an applet requires gtk.
func GtkNeeded() bool {
	return needgtk
}
