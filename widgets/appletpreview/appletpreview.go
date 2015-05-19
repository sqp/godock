// Package appletpreview provides an applet preview widget.
package appletpreview

import (
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/godock/libs/log"
	"github.com/sqp/godock/widgets/common"
	"github.com/sqp/godock/widgets/confbuilder/datatype"
	"github.com/sqp/godock/widgets/gtk/buildhelp"

	"fmt"
)

// const DataDirCRAP = "/usr/share/cairo-dock/"

// Preview image settings.
const (
	MaxPreviewWidth  = 350
	MaxPreviewHeight = 250
)

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
}

// New creates an applet preview widget.
//
func New() *Preview {
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
	}

	if len(builder.Errors) > 0 {
		for _, e := range builder.Errors {
			log.DEV("build appletpreview", e)
		}
		return nil
	}

	widget.Connect("destroy", widget.RemoveTmpFile)

	return widget
}

// Load loads an applet in the preview.
//
func (widget *Preview) Load(pack datatype.Appleter) {
	widget.title.SetMarkup(common.Big(common.Bold(pack.GetTitle())))
	widget.author.SetMarkup(common.Small(common.Mono(fmt.Sprintf("by %s", pack.GetAuthor()))))
	// widget.stateText.SetMarkup(pack.FormatState())
	// widget.size.SetMarkup(common.Small(pack.FormatSize()))

	//DataDirCRAP+pack.IconState()
	// if pixbuf, e := common.PixbufAtSize(pack.IconState(), 24, 24); !log.Err(e, "Load image pixbuf") {
	// 	widget.stateIcon.SetFromPixbuf(pixbuf)
	// 	widget.stateIcon.Show()
	// }

	// widget.RemoveTmpFile()
	widget.previewFrame.Hide() // Hide the preview frame until we have an image.

	// Async calls for description and image. They can have to be downloaded and be slow at it.
	go widget.description.SetMarkup(pack.GetDescription())
	go widget.setImage(pack)
}

func (widget *Preview) setImage(pack datatype.Appleter) {
	imageLocation := pack.GetPreviewFilePath()
	// imageLocation, isTemp := pack.GetPreview(widget.TmpFile) // reuse the same tmp location if needed.
	if imageLocation != "" {
		if pixbuf, e := common.PixbufAtSize(imageLocation, MaxPreviewWidth, MaxPreviewHeight); !log.Err(e, "Load image pixbuf") {
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
	// widget.stateText.Hide()
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
