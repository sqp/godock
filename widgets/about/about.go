// Package about creates the program about dialog.
package about

import (
	"github.com/gotk3/gotk3/gtk"

	"github.com/sqp/godock/libs/text/tran"

	"github.com/sqp/godock/widgets/gtk/newgtk"

	"fmt"
)

// Dialog window size.
const (
	AboutWidth  = 400
	AboutHeight = 500
)

// Project URLs for link buttons.
const (
	URLdockCode   = "https://github.com/Cairo-Dock"
	URLdockSite   = "http://glx-dock.org" // http://cairo-dock.vef.fr
	URLdockPayPal = "https://www.paypal.com/cgi-bin/webscr?cmd=_donations&business=UWQ3VVRB2ZTZS&lc=GB&item_name=Support%20Cairo%2dDock&currency_code=EUR&bn=PP%2dDonationsBF%3abtn_donate_LG%2egif%3aNonHosted"
	URLdockFlattr = "http://flattr.com/thing/370779/Support-Cairo-Dock-development"
)

var Img string

//
//------------------------------------------------------------[ ABOUT DIALOG ]--

// About defines the program about dialog.
//
type About gtk.Dialog

// New creates a GuiIcons widget to edit cairo-dock icons config.
//
func New() *About {
	// GtkWidget *pDialog = gtk_dialog_new_with_buttons (_(""),
	// 	GTK_WINDOW (pContainer->pWidget),
	// 	GTK_DIALOG_DESTROY_WITH_PARENT,
	// 	GLDI_ICON_NAME_CLOSE,
	// 	GTK_RESPONSE_CLOSE,
	// 	NULL);

	dialog := newgtk.Dialog()
	dialog.SetTitle(tran.Slate("About Cairo-Dock"))
	// dialog.SetParentWindow(parentWindow)
	dialog.AddButton(tran.Slate("_Close"), gtk.RESPONSE_CLOSE)
	dialog.Connect("response", dialog.Destroy)

	// Widgets.
	image := newgtk.ImageFromFile(Img)
	notebook := addNotebook()
	links := addLinks()

	// Packing.
	header := newgtk.Box(gtk.ORIENTATION_HORIZONTAL, 0)
	header.PackStart(image, false, false, 0)
	header.PackStart(links, false, false, 0)

	content, _ := dialog.GetContentArea()
	content.PackStart(header, false, false, 0)
	content.PackStart(notebook, true, true, 0)

	dialog.Resize(AboutWidth, AboutHeight) // had check min size compared to desktop size...
	dialog.ShowAll()

	//don't use gtk_dialog_run(), as we don't want to block the dock
	dialog.SetKeepAbove(true)

	return (*About)(dialog)
}

//
//-------------------------------------------------------------------[ LINKS ]--

func addLinks() *gtk.Box {
	linksData := []struct {
		uri     string
		label   string
		tooltip string
	}{{
		uri:   URLdockSite,
		label: "Cairo-Dock (2007-2014)", //GLDI_VERSION;
	}, {
		uri:     URLdockCode,
		label:   tran.Slate("Development site"),
		tooltip: tran.Slate("Find the latest version of Cairo-Dock here !"),
	}, {
		uri:     URLdockFlattr,
		label:   tran.Slate("Donate") + " (Flattr)",
		tooltip: tran.Slate("Support the people who spend countless hours to bring you the best dock ever."),
	}, {
		uri:     URLdockPayPal,
		label:   tran.Slate("Donate") + " (Paypal)",
		tooltip: tran.Slate("Support the people who spend countless hours to bring you the best dock ever."),
	}}

	box := newgtk.Box(gtk.ORIENTATION_VERTICAL, 0)
	for _, item := range linksData {
		link := newgtk.LinkButtonWithLabel(item.uri, item.label)
		box.PackStart(link, false, false, 0)
		if item.tooltip != "" {
			link.SetTooltipText(item.tooltip)
		}
	}
	return box
}

//
//----------------------------------------------------------[ NOTEBOOK BUILD ]--

func addNotebook() *gtk.Notebook {
	notebookData := []struct {
		title   string
		content string
		args    []interface{}
	}{{
		title:   "Development",
		content: tabDev,
		args: []interface{}{
			tran.Slate("Here is a list of the current developers and contributors"),
			tran.Slate("Developers"),
			tran.Slate("Main developer and project leader"),
			tran.Slate("Contributors / Hackers"),
		},
	}, {
		title:   "Support",
		content: tabSupport,
		args: []interface{}{
			tran.Slate("Website"),
			tran.Slate("Beta-testing / Suggestions / Forum animation"),
			tran.Slate("Translators for this language"),
			tran.Slate("translator-credits"),
		},
	}, {
		title:   "Thanks",
		content: tabThanks,
		args: []interface{}{
			tran.Slate("Thanks to all people that help us to improve the Cairo-Dock project.\n\n" +
				"Thanks to all current, former and future contributors."),
			tran.Slate("How to help us?"),
			tran.Slate("Don't hesitate to join the project, we need you ;)"),
			tran.Slate("Former contributors"),
			tran.Slate("For a complete list, please have a look to BZR logs"),
			tran.Slate("Users of our forum"),
			tran.Slate("List of our forum's members"),
			tran.Slate("Artwork"),
		},
	}}

	notebook := newgtk.Notebook()
	notebook.SetScrollable(true)

	for _, nb := range notebookData {
		notebookPage(notebook, tran.Slate(nb.title), nb.content, nb.args)
	}
	return notebook
}

func notebookPage(nb *gtk.Notebook, label, contentStr string, contentArgs []interface{}) {
	tabLabel := newgtk.Label(label)
	box := newgtk.Box(gtk.ORIENTATION_VERTICAL, 0)
	scroll := newgtk.ScrolledWindow(nil, nil)
	scroll.Add(box)
	nb.AppendPage(scroll, tabLabel)

	content := fmt.Sprintf(contentStr, contentArgs...)
	contentLabel := newgtk.Label(content)
	contentLabel.SetUseMarkup(true)
	box.PackStart(contentLabel, false, false, 15)

	nb.PopupEnable()
}

//
//-----------------------------------------------------------[ NOTEBOOK TEXT ]--

const tabDev = "%s\n\n" +
	"<span size=\"larger\" weight=\"bold\">%s</span>\n\n" +
	"  Fabounet (Fabrice Rey)\n" +
	"\t<span size=\"smaller\">%s</span>\n\n" +
	"  Matttbe (Matthieu Baerts)\n" +
	"\n\n<span size=\"larger\" weight=\"bold\">%s</span>\n\n" +
	"  Eduardo Mucelli\n" +
	"  Jesuisbenjamin\n" +
	"  SQP\n"

const tabSupport = "<span size=\"larger\" weight=\"bold\">%s</span>\n\n" +
	"  Matttbe\n" +
	"  Mav\n" +
	"  Necropotame\n" +
	"\n\n<span size=\"larger\" weight=\"bold\">%s</span>\n\n" +
	"  BobH\n" +
	"  Franksuse64\n" +
	"  Lylambda\n" +
	"  Ppmt\n" +
	"  Taiebot65\n" +
	"\n\n<span size=\"larger\" weight=\"bold\">%s</span>\n\n" +
	"%s"

const tabThanks = "%s\n" +
	"<a href=\"http://glx-dock.org/ww_page.php?p=How to help us\">%s</a>: %s\n\n" +
	"\n<span size=\"larger\" weight=\"bold\">%s</span>\n\n" +
	"  Augur\n" +
	"  ChAnGFu\n" +
	"  Ctaf\n" +
	"  Mav\n" +
	"  Necropotame\n" +
	"  Nochka85\n" +
	"  Paradoxxx_Zero\n" +
	"  Rom1\n" +
	"  Tofe\n" +
	"  Mac Slow (original idea)\n" +
	"\t<span size=\"smaller\">%s</span>\n" +
	"\n\n<span size=\"larger\" weight=\"bold\">%s</span>\n\n" +
	"\t<a href=\"http://glx-dock.org/userlist_messages.php\">%s</a>\n" +
	"\n\n<span size=\"larger\" weight=\"bold\">%s</span>\n\n" +
	"  Benoit2600\n" +
	"  Coz\n" +
	"  Fabounet\n" +
	"  Lord Northam\n" +
	"  Lylambda\n" +
	"  MastroPino\n" +
	"  Matttbe\n" +
	"  Nochka85\n" +
	"  Paradoxxx_Zero\n" +
	"  Taiebot65\n"
