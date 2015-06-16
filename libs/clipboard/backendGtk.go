// +build gtk || dock

package clipboard

import (
	"github.com/conformal/gotk3/gdk"
	"github.com/conformal/gotk3/glib"
	"github.com/conformal/gotk3/gtk"
)

func init() {
	Write = func(text string) error {
		clip, e := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)
		if e != nil {
			return e
		}
		glib.IdleAdd(func() {
			clip.SetText(text)
		})
		return nil
	}

	Read = func() (string, error) {
		clip, e := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)
		if e != nil {
			return "", e
		}
		cs := make(chan (string))
		ce := make(chan (error))
		glib.IdleAdd(func() { // Synced in the GTK loop to prevent thread crashs.
			str, e := clip.WaitForText()
			cs <- str
			close(cs)
			ce <- e
			close(ce)
		})
		return <-cs, <-ce
	}
}
