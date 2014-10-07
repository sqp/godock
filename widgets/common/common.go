// Package common provides simple gtk helpers.
package common

import (
	"github.com/conformal/gotk3/gdk"
	"github.com/conformal/gotk3/gtk"

	"errors"
	"math"
)

//
//------------------------------------------------------------------[ WINDOW ]--

// Special hack to prevent threads related crashs.
// http://stackoverflow.com/questions/18647475/threading-problems-with-gtk
// http://stackoverflow.com/questions/13351297/what-is-the-downside-of-xinitthreads

// #include <X11/Xlib.h>
// #cgo pkg-config: x11
import "C"

func GRRTHREADS() {
	C.XInitThreads()
}

func InitGtk() (onstart, onstop func()) {
	gtkStart := func() {

		GRRTHREADS()

		// runtime.LockOSThread()
		gtk.Init(nil)
		gtk.Main()
	}
	return gtkStart, gtk.MainQuit
}

// Create a new toplevel window, set size and pack the main widget.
//
func NewWindowMain(widget gtk.IWidget, w, h int) *gtk.Window {
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		// log.Fatal("Unable to create window:", err)
		return nil
	}
	win.SetDefaultSize(w, h)
	win.Add(widget)
	win.ShowAll()
	return win
}

/*
func PixbufAtSize(file string, maxW, maxH int) (*gdk.Pixbuf, error) {
	return nil, nil
}

func ImageNewFromFile(cIcon string, iSize int) (pImage *gtk.Image) {
	return nil
}

func PixbufNewFromFile(cIcon string, iSize int) (pixbuf *gdk.Pixbuf, e error) {
	return nil, nil
}
*/

func PixbufAtSize(file string, maxW, maxH int) (*gdk.Pixbuf, error) {
	if _, imgW, imgH := gdk.PixbufGetFileInfo(file); imgW > 0 {
		ratio := math.Min(math.Min(float64(maxW)/float64(imgW), float64(maxH)/float64(imgH)), 1)
		// log.Info("ratio", ratio)
		pix, e := gdk.PixbufNewFromFileAtSize(file, int(float64(imgW)*ratio), int(float64(imgH)*ratio))
		return pix, e
	}
	return nil, errors.New("Problem getting file info: " + file)
}

func ImageNewFromFile(cIcon string, iSize int) (pImage *gtk.Image) {
	switch {
	case len(cIcon) == 0:
		return nil

	case cIcon[0] != '/': // GTK stock icon
		//img, e := gtk.ImageNewFromStock(gtk.Stock(cIcon), gtk.IconSize(iSize))
		//log.Err(e, "Load image stock")
		//return img

	default: // Full path.

		// if iSize == GTK_ICON_SIZE_BUTTON { /// TODO: find a way to get a correct transposition...
		// 	iSize = CAIRO_DOCK_TAB_ICON_SIZE
		// } else if iSize == GTK_ICON_SIZE_MENU {
		// 	iSize = CAIRO_DOCK_FRAME_ICON_SIZE
		// }

		// if pixbuf, e := PixbufAtSize(cIcon, iSize, iSize); !log.Err(e, "Load image pixbuf") {
		// if img, e := gtk.ImageNewFromPixbuf(pixbuf); !log.Err(e, "Create preview image widget") {
		// 	pImage = img
		// }
		// }
	}
	// GdkPixbuf * pixbuf = gdk_pixbuf_new_from_file_at_size(cIcon, iSize, iSize, NULL)
	// if pixbuf != nil {
	// 	gtk_image_set_from_pixbuf(GTK_IMAGE(pImage), pixbuf)
	// 	g_object_unref(pixbuf)
	// }

	return pImage
}

func PixbufNewFromFile(cIcon string, iSize int) (pixbuf *gdk.Pixbuf, e error) {
	switch {
	case len(cIcon) == 0:
		return nil, errors.New("PixbufNewFromFile: empty name")

	case cIcon[0] != '/': // GTK stock icon
		// return gtk.PixbufNewFromStock(cIcon, iSize)

		// return gdk.PixbufNewFromResourceAtScale(cIcon, iSize, iSize, true)
		// img, e := gdk.PixbufNewFromResourceAtScale(cIcon, iSize, iSize, true)
		// log.Err(e, "Load image stock")
		// return img, nil

	default: // Full path.

		// if iSize == GTK_ICON_SIZE_BUTTON { /// TODO: find a way to get a correct transposition...
		// 	iSize = CAIRO_DOCK_TAB_ICON_SIZE
		// } else if iSize == GTK_ICON_SIZE_MENU {
		// 	iSize = CAIRO_DOCK_FRAME_ICON_SIZE
		// }

		return PixbufAtSize(cIcon, iSize, iSize)
		// if pix, e := PixbufAtSize(cIcon, iSize, iSize); !log.Err(e, "Load image pixbuf") {
		// 	return pix
		// }
	}

	return nil, errors.New("PixbufNewFromFile: no match for " + cIcon)
}

func Big(text string) string   { return "<big>" + text + "</big>" }
func Bold(text string) string  { return "<b>" + text + "</b>" }
func Small(text string) string { return "<small>" + text + "</small>" }
func Mono(text string) string  { return "<tt>" + text + "</tt>" }
