// +build gtk dock

package clipboard

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"errors"
	"time"
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

		defer func() {
			close(cs)
			close(ce)
		}()
		done := false
		glib.IdleAdd(func() { // Synced in the GTK loop to prevent thread crashs.
			str, e := clip.WaitForText()
			if !done {
				done = true
				cs <- str
				ce <- e
			}
		})
		go func() {
			<-time.After(time.Second * 3)
			if !done {
				done = true
				cs <- ""
				ce <- errors.New("clipboard read timeout")
			}
		}()
		return <-cs, <-ce
	}
}
