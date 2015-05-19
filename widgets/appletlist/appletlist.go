// Package appletlist provides an applets list treeview widget.
package appletlist

import (
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/godock/libs/log"
	"github.com/sqp/godock/widgets/common"
	"github.com/sqp/godock/widgets/confbuilder/datatype"
	"github.com/sqp/godock/widgets/gtk/buildhelp"
	"github.com/sqp/godock/widgets/gtk/gunvalue"
)

//-----------------------------------------------------[ WIDGET APPLETS LIST ]--

const iconSize = 24

// const gdkInterpol = 2 // 2 is GDK_INTERP_BILINEAR, supposed to be the default.

// Rows defines liststore rows. Must match the ListStore declaration type and order.
const (
	RowKey = iota
	RowIcon
	RowName
	RowCategory
	RowNameWeight // Display param
)

// ControlDownload is the interface to the main GUI for the download page.
//
type ControlDownload interface {
	OnSelect(datatype.Appleter)
	SetControlInstall(ControlInstall)
}

// ControlInstall is the interface to the download page for the main GUI.
//
type ControlInstall interface {
	Selected() datatype.Appleter
	SetActive(state bool)
}

// Row defines a pointer to link the package reference with its iter.
//
type Row struct {
	Iter *gtk.TreeIter
	Pack datatype.Appleter
}

// List defines an applets list widget.
//
type List struct {
	gtk.ScrolledWindow // Main widget is the container. The ScrolledWindow will handle list scrollbars.
	tree               *gtk.TreeView
	model              *gtk.ListStore
	control            ControlDownload

	rows map[string]*Row
}

// NewList creates a new applets list widget.
//
func NewList(control ControlDownload) *List {
	builder := buildhelp.New()

	builder.AddFromString(string(appletlistXML()))
	// builder.AddFromFile("appletlist.xml")

	widget := &List{
		ScrolledWindow: *builder.GetScrolledWindow("widget"),
		model:          builder.GetListStore("model"),
		tree:           builder.GetTreeView("tree"),
		control:        control,
		rows:           make(map[string]*Row),
	}

	if len(builder.Errors) > 0 {
		for _, e := range builder.Errors {
			log.DEV("build appletlist", e)
		}
		return nil
	}

	control.SetControlInstall(widget)

	// Action: Treeview Select line.
	if sel, e := widget.tree.GetSelection(); !log.Err(e, "appletlist TreeView.GetSelection") {
		sel.Connect("changed", widget.onSelectionChanged) // Changed is connected to TreeSelection
	}

	return widget
}

// SetActive sets the active state of selected line.
//
func (widget *List) SetActive(state bool) {
	sel, _ := widget.tree.GetSelection()
	var iter gtk.TreeIter
	var treeModel gtk.ITreeModel = widget.model
	sel.GetSelected(&treeModel, &iter)
	widget.setActiveIter(&iter, state)
}

// Set the active state of the iter argument.
//
func (widget *List) setActiveIter(iter *gtk.TreeIter, state bool) {
	weight := 400
	if state {
		weight += 200
	}
	widget.model.Set(iter, []int{RowNameWeight}, []interface{}{weight})
}

// Selected returns the applet package for the selected line.
//
func (widget *List) Selected() datatype.Appleter {
	sel, _ := widget.tree.GetSelection()
	var iter gtk.TreeIter
	var treeModel gtk.ITreeModel = widget.model
	sel.GetSelected(&treeModel, &iter)
	name, e := gunvalue.New(widget.model.GetValue(&iter, RowKey)).String()
	if log.Err(e, "Get selected iter name GoValue") {
		return nil
	}
	if row, ok := widget.rows[name]; ok {
		return row.Pack
	}
	return nil
}

func (widget *List) newIter(name string, pack datatype.Appleter) *gtk.TreeIter {
	iter := widget.model.Append()
	widget.rows[name] = &Row{iter, pack}
	return iter
}

// Selected line has changed. Forward the call to the controler.
//
func (widget *List) onSelectionChanged(obj *gtk.TreeSelection) {
	if widget.control != nil {
		widget.control.OnSelect(widget.Selected())
	}
}

// Delete deletes a row.
//
func (widget *List) Delete(key string) {
	row, ok := widget.rows[key]
	if !ok {
		return
	}
	widget.model.Remove(row.Iter)
	delete(widget.rows, key)
}

// Clear clears the widget data.
//
func (widget *List) Clear() {
	widget.rows = make(map[string]*Row)
	widget.model.Clear()
}

//
//---------------------------------------------------------[ LIST ADD APPLET ]--

// ListAdd defines an applet list widget with applets allowed to be enabled.
//
type ListAdd struct {
	List
}

// NewListAdd creates an applet list widget with applets allowed to be enabled.
//
func NewListAdd(control ControlDownload) *ListAdd {
	return &ListAdd{*NewList(control)}
}

// Load loads the applet list into the widget.
//
func (widget *ListAdd) Load(list map[string]datatype.Appleter) {
	for key, app := range list {
		if app.CanAdd() {
			iter := widget.newIter(key, app)
			widget.model.SetCols(iter, gtk.Cols{
				RowKey:      key,
				RowName:     app.GetTitle(),
				RowCategory: app.FormatCategory(),
			})

			// 	int iSize = cairo_dock_search_icon_size (GTK_ICON_SIZE_LARGE_TOOLBAR);
			// 	gchar *cIcon = cairo_dock_search_icon_s_path (pModule->pVisitCard->cIconFilePath, iSize);

			img := app.GetIconFilePath()
			if pix, e := common.PixbufNewFromFile(img, iconSize); !log.Err(e, "Load icon") {
				widget.model.SetValue(iter, RowIcon, pix)
			}

			widget.setActiveIter(iter, app.IsActive())
		}
	}
}

//
//---------------------------------------------------[ LIST EXTERNAL APPLETS ]--

// ListExternal defines an applet list widget with external applets to install.
//
type ListExternal struct {
	List
}

// NewListExternal creates an applet list widget with external applets to install.
//
func NewListExternal(control ControlDownload) *ListExternal {
	return &ListExternal{*NewList(control)}
}

// Load loads the applet list into the widget.
//
func (widget *ListExternal) Load(list map[string]datatype.Appleter) {
	for key, pack := range list {
		iter := widget.newIter(key, pack)
		// iter := widget.model.Append()
		widget.model.SetCols(iter, gtk.Cols{
			RowKey:      key,
			RowName:     pack.GetName(),
			RowCategory: pack.FormatCategory(),
		})

		if pack.IsInstalled() { // local packages.
			widget.setActiveIter(iter, true)

			img := pack.GetIconFilePath()
			if pix, e := common.PixbufNewFromFile(img, iconSize); !log.Err(e, "Load icon") {
				widget.model.SetValue(iter, RowIcon, pix)
			}
		} else {
			widget.setActiveIter(iter, false)
			// icon missing, can't set.
		}
	}
}

// func modelApplet() *gtk.ListStore {
// 	store, _ := gtk.ListStoreNew(
// 		glib.TYPE_STRING,  // AppletName
// 		glib.TYPE_STRING,  // AppletResult
// 		glib.TYPE_STRING,  // AppletDescriptionFile
// 		glib.TYPE_STRING,  // AppletImage
// 		glib.TYPE_BOOLEAN, // AppletActive
// 		glib.TYPE_INT,     // AppletOrder
// 		glib.TYPE_INT,     // AppletOrder2
// 		// gdkpixbuf.G_TYPE_PIXBUF, // AppletIcon
// 		glib.TYPE_INT,    // AppletState
// 		glib.TYPE_DOUBLE, // AppletSize
// 		glib.TYPE_STRING, // AppletAuthor
// 		glib.TYPE_STRING, // AppletCategory
// 		glib.TYPE_STRING, // AppletNameDisplayed
// 	)
// 	return store
// }
