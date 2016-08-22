// Package handbook provides an applet or theme description widget.
package handbook

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"github.com/sqp/godock/libs/cdtype"

	"github.com/sqp/godock/widgets/cfbuild/datatype"
	"github.com/sqp/godock/widgets/common"
	"github.com/sqp/godock/widgets/gtk/buildhelp"

	"fmt"
)

// PreviewSizeMax defines the preview widget size (image).
//
const PreviewSizeMax = 200

//
//----------------------------------------------------------------[ HANDBOOK ]--

// Handbook defines a handbook widget (applet info).
//
type Handbook struct {
	gtk.Frame    // Main widget is the container.
	title        *gtk.Label
	author       *gtk.Label
	description  *gtk.Label
	previewFrame *gtk.Frame
	previewImage *gtk.Image

	ShowVersion bool
}

// New creates a handbook widget (applet info).
//
func New(log cdtype.Logger) *Handbook {
	builder := buildhelp.New()

	builder.AddFromString(string(handbookXML()))
	// builder.AddFromFile("handbook.xml")

	widget := &Handbook{
		Frame:        *builder.GetFrame("handbook"),
		title:        builder.GetLabel("title"),
		author:       builder.GetLabel("author"),
		description:  builder.GetLabel("description"),
		previewFrame: builder.GetFrame("previewFrame"),
		previewImage: builder.GetImage("previewImage"),
	}

	if len(builder.Errors) > 0 {
		for _, e := range builder.Errors {
			log.Err(e, "build handbook")
		}
		return nil
	}

	return widget
}

// SetPackage fills the handbook data with a package.
//
func (widget *Handbook) SetPackage(book datatype.Handbooker) {
	title := common.Bold(common.Big(book.GetTitle()))
	if widget.ShowVersion {
		title += " v" + book.GetModuleVersion()
	}
	widget.title.SetMarkup(title)

	author := book.GetAuthor()
	if author != "" {
		author = fmt.Sprintf("by %s", author)
		widget.author.SetMarkup(common.Small(common.Mono(author)))
	}
	widget.author.SetVisible(author != "")

	widget.description.SetMarkup("<span rise='8000'>" + book.GetDescription() + "</span>")

	previewFound := false
	defer func() { widget.previewFrame.SetVisible(previewFound) }()

	file := book.GetPreviewFilePath()
	if file == "" {
		return
	}
	_, w, h := gdk.PixbufGetFileInfo(file)

	var pixbuf *gdk.Pixbuf
	var e error
	if w > PreviewSizeMax || h > PreviewSizeMax {
		pixbuf, e = gdk.PixbufNewFromFileAtScale(file, PreviewSizeMax, PreviewSizeMax, true)
	} else {
		pixbuf, e = gdk.PixbufNewFromFile(file)
	}

	if e == nil && pixbuf != nil {
		previewFound = true
		widget.previewImage.SetFromPixbuf(pixbuf)
	}
}
