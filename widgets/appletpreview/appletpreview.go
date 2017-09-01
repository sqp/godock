// Package appletpreview provides an applet preview widget.
package appletpreview

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"github.com/sqp/godock/libs/cdtype"       // Logger type.
	"github.com/sqp/godock/libs/text/gtktext" // Format text GTK.
	"github.com/sqp/godock/libs/text/tran"    // Translate.

	"github.com/sqp/godock/widgets/common"
	"github.com/sqp/godock/widgets/gtk/buildhelp" // Widget builder.

	"fmt"
)

// Preview image settings.
const (
	MaxPreviewWidth  = 350
	MaxPreviewHeight = 250
)

// Previewer defines the main data needed by the preview.
type Previewer interface {
	GetTitle() string
	GetAuthor() string
	GetDescription() string
	GetPreviewFilePath() string
}

// Appleter defines additional data that can be used by the preview.
type Appleter interface {
	FormatState() string
	FormatSize() string
	IconState() string
}

//-----------------------------------------------------[ WIDGET APPLET PREVIEW ]--

// Preview defines an applet preview widget.
//
type Preview struct {
	gtk.Box      // Main widget is the container.
	title        *gtk.Label
	author       *gtk.Label
	size         *gtk.Label
	stateText    *gtk.Label
	stateIcon    *gtk.Image
	description  *gtk.Label
	previewImage *gtk.Image // gtk.IWidget
	previewFrame *gtk.Frame

	TmpFile string

	log cdtype.Logger
}

// New creates an applet preview widget.
//
func New(log cdtype.Logger) *Preview {
	builder := buildhelp.New()
	builder.AddFromString(string(appletpreviewXML()))
	// builder.AddFromFile("appletpreview.xml")

	widget := &Preview{
		Box:          *builder.GetBox("widget"),
		title:        builder.GetLabel("title"),
		author:       builder.GetLabel("author"),
		size:         builder.GetLabel("size"),
		stateText:    builder.GetLabel("stateText"),
		stateIcon:    builder.GetImage("stateIcon"),
		description:  builder.GetLabel("description"),
		previewFrame: builder.GetFrame("previewFrame"),
		previewImage: builder.GetImage("previewImage"),
		log:          log,
	}

	if len(builder.Errors) > 0 {
		for _, e := range builder.Errors {
			log.Err(e, "build appletpreview")
		}
		return nil
	}

	widget.Connect("destroy", widget.RemoveTmpFile)

	return widget
}

// Load loads an applet or theme in the preview. Handbooker and Appleter can be used.
//
func (widget *Preview) Load(pack Previewer) {
	widget.title.SetMarkup(gtktext.Big(gtktext.Bold(pack.GetTitle())))
	author := pack.GetAuthor()
	if author != "" {
		author = fmt.Sprintf(tran.Slate("by %s"), author)
	}
	widget.author.SetMarkup(gtktext.Small(gtktext.Mono(author)))

	apl, ok := pack.(Appleter)
	if ok {
		widget.stateText.SetMarkup(apl.FormatState())
		widget.size.SetMarkup(gtktext.Small(apl.FormatSize()))

		if icon := apl.IconState(); icon != "" {
			if pixbuf, e := common.PixbufAtSize(icon, 24, 24); !widget.log.Err(e, "Load image pixbuf") {
				widget.stateIcon.SetFromPixbuf(pixbuf)
				widget.stateIcon.Show()
			}
		}
	}

	// widget.RemoveTmpFile()

	widget.previewFrame.Hide() // Hide the preview frame until we have an image.

	// Async calls for description and image. They can have to be downloaded and be slow at it.

	chDesc := make(chan (string))
	widget.log.GoTry(func() { // Go routines to get data.
		chDesc <- pack.GetDescription()
	})

	widget.log.GoTry(func() {
		imageLocation := pack.GetPreviewFilePath()
		// imageLocation, isTemp := pack.GetPreview(widget.TmpFile) // reuse the same tmp location if needed.

		desc := <-chDesc
		glib.IdleAdd(func() { // glib Idle to show the result.
			widget.description.SetMarkup(desc)
			widget.setImage(imageLocation)
		})
	})

}

func (widget *Preview) setImage(imageLocation string) {
	if imageLocation != "" {
		pixbuf, e := common.PixbufAtSize(imageLocation, MaxPreviewWidth, MaxPreviewHeight)
		if e == nil {
			widget.previewImage.SetFromPixbuf(pixbuf)
			widget.previewFrame.Show()
		}
	}

	// if isTemp {
	// 	widget.TmpFile = imageLocation
	// }

}

// HideState hides the state widget.
//
func (widget *Preview) HideState() {
	widget.stateText.Hide()
	widget.stateIcon.Hide()
}

// HideSize hides the size widget.
//
func (widget *Preview) HideSize() {
	widget.size.Hide()
}

// RemoveTmpFile delete the temporary file if used.
//
func (widget *Preview) RemoveTmpFile() {
	if widget.TmpFile != "" {
		println("need to delete temp", widget.TmpFile)
	}
}

// gboolean bHorizontalPackaging,
// int iAddInfoBar,
// const gchar *cInitialDescription,
// const gchar *cInitialImage,
