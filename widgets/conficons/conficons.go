// Package conficons provides an icons list and configuration widget.
package conficons

import (
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/godock/libs/gldi"
	"github.com/sqp/godock/libs/log"

	"github.com/sqp/godock/widgets/common"
	"github.com/sqp/godock/widgets/confbuilder"
	"github.com/sqp/godock/widgets/confbuilder/datatype"
	"github.com/sqp/godock/widgets/gtk/buildhelp"
	"github.com/sqp/godock/widgets/gtk/gunvalue"
	"github.com/sqp/godock/widgets/pageswitch"

	"errors"
	"path/filepath"
	"strings"
)

const iconSize = 24
const listIconsWidth = 200

//--------------------------------------------------------[ PAGE GUI ICONS ]--

// Controller defines methods used on the main widget / data source by this widget and its sons.
//
type Controller interface {
	datatype.Source
	GetWindow() *gtk.Window
}

// configWidget defines a GtkWidget with a Save method.
//
type configWidget interface {
	gtk.IWidget
	Save()
}

// GuiIcons defines Icons configuration widget for currently actived cairo-dock Icons.
//
type GuiIcons struct {
	gtk.Paned

	icons  *List
	config *confbuilder.Grouper
	// page     configWidget
	switcher *pageswitch.Switcher

	data Controller
}

// New creates a GuiIcons widget to edit cairo-dock icons config.
//
func New(data Controller, switcher *pageswitch.Switcher) *GuiIcons {
	paned, _ := gtk.PanedNew(gtk.ORIENTATION_HORIZONTAL)
	widget := &GuiIcons{
		Paned:    *paned,
		switcher: switcher,
		data:     data,
	}
	widget.icons = NewList(widget)

	up, _ := gtk.ButtonNewFromIconName("go-up", gtk.ICON_SIZE_BUTTON)
	down, _ := gtk.ButtonNewFromIconName("go-down", gtk.ICON_SIZE_BUTTON)
	remove, _ := gtk.ButtonNewFromIconName("list-remove", gtk.ICON_SIZE_BUTTON)

	boxLeft, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	boxBtns, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	boxLeft.PackStart(widget.icons, true, true, 0)
	boxLeft.PackStart(boxBtns, false, false, 0)
	boxBtns.PackStart(up, false, false, 0)
	boxBtns.PackStart(down, false, false, 0)
	boxBtns.PackEnd(remove, false, false, 0)

	widget.Pack1(boxLeft, true, true)

	widget.SetPosition(listIconsWidth) // Paned position = list icons width.

	up.Connect("clicked", func() {
		ic, e := widget.icons.SelectedIcon()
		if e == nil {
			ic.MoveBeforePrevious()
		}
	})

	down.Connect("clicked", func() {
		ic, e := widget.icons.SelectedIcon()
		if e == nil {
			ic.MoveAfterNext()
		}
	})

	remove.Connect("clicked", func() {
		ic, e := widget.icons.SelectedIcon()
		if e == nil {
			ic.RemoveFromDock()
		}
	})

	// widget.icons.Connect("row-inserted", func() { log.Info("row inserted") })
	// widget.icons.Connect("row-deleted", func() { log.Info("row deleted") })

	return widget
}

// Load loads the list of icons in the iconsList.
//
func (widget *GuiIcons) Load() {
	icons := widget.data.ListIcons()
	widget.icons.Load(icons)
}

// Selected returns the selected icon.
//
// func (widget *GuiIcons) Selected() datatype.Iconer {
// 	return widget.icons.Selected()
// }

// Select sets the selected icon based on its config file.
//
func (widget *GuiIcons) Select(conf string) bool {
	return widget.icons.Select(conf)
}

// Clear clears the widget data.
//
func (widget *GuiIcons) Clear() string {
	path, _ := widget.icons.SelectedConf()
	widget.switcher.Clear()

	if widget.config != nil {
		widget.config.Destroy()
		widget.config = nil
	}
	widget.icons.Clear()
	return path
}

//
//-------------------------------------------------------[ SAVE CONFIG APPLET ]--

// Save saves the current page configuration
//
func (widget *GuiIcons) Save() {
	if widget.config == nil {
		return
	}

	icon, e := widget.icons.SelectedIcon()

	// applet
	// 		// if the parent dock doesn't exist (new dock), add a conf file for it with a nominal name.
	// 		if (g_key_file_has_key (pKeyFile, "Icon", "dock name", NULL))
	// 		{
	// 			gchar *cDockName = g_key_file_get_string (pKeyFile, "Icon", "dock name", NULL);
	// 			gboolean bIsDetached = g_key_file_get_boolean (pKeyFile, "Desklet", "initially detached", NULL);
	// 			if (!bIsDetached)
	// 			{
	// 				CairoDock *pDock = gldi_dock_get (cDockName);
	// 				if (pDock == NULL)
	// 				{
	// 					gchar *cNewDockName = gldi_dock_add_conf_file ();
	// 					g_key_file_set_string (pKeyFile, "Icon", "dock name", cNewDockName);
	// 					g_free (cNewDockName);
	// 				}
	// 			}
	// 			g_free (cDockName);
	// 		}

	// icon
	// 		// if the parent dock doesn't exist (new dock), add a conf file for it with a nominal name.
	// 		if (g_key_file_has_key (pKeyFile, "Desktop Entry", "Container", NULL))
	// 		{
	// 			gchar *cDockName = g_key_file_get_string (pKeyFile, "Desktop Entry", "Container", NULL);
	// 			CairoDock *pDock = gldi_dock_get (cDockName);
	// 			if (pDock == NULL)
	// 			{
	// 				gchar *cNewDockName = gldi_dock_add_conf_file ();
	// 				g_key_file_set_string (pKeyFile, "Icon", "dock name", cNewDockName);
	// 				g_free (cNewDockName);
	// 			}
	// 			g_free (cDockName);
	// 		}

	// 		if (pModuleInstance->pModule->pInterface->save_custom_widget != NULL)
	// 			pModuleInstance->pModule->pInterface->save_custom_widget (pModuleInstance, pKeyFile, pWidgetList);

	widget.config.Save()
	if e != nil {
		icon.Reload()
	}

	// 	_items_widget_reload (CD_WIDGET (pItemsWidget));  // we reload in case the items place has changed (icon's container, dock orientation, etc).

}

//
//-------------------------------------------------------[ CONTROL CALLBACKS ]--

// onSelect reacts when a row is selected. Creates a new config for the icon.
//
func (widget *GuiIcons) onSelect(icon datatype.Iconer, ei error) {
	widget.switcher.Clear()

	if widget.config != nil {
		widget.config.Destroy()
		widget.config = nil
	}

	if ei != nil { // shouldn't match.
		return
	}

	build, e := confbuilder.NewGrouper(widget.data, widget.data.GetWindow(), icon.ConfigPath(), icon.GetGettextDomain())
	if log.Err(e, "Load Keyfile "+icon.ConfigPath()) {
		return
	}

	if icon.IsTaskbar() {
		widget.config = build.BuildSingle("TaskBar")
	} else {
		widget.config = build.BuildAll(widget.switcher)

		// Little hack for empty launchers, not sure it could go somewhere else.
		if icon.IsLauncher() {
			origins, _ := build.Conf.GetString("Desktop Entry", "Origin")
			widget.config.PackStart(launcherMagic(icon, origins), false, false, 10)
		}

	}
	widget.Pack2(widget.config, true, true)
	widget.config.ShowAll()

}

func launcherMagic(icon datatype.Iconer, origins string) gtk.IWidget {
	// println("comand", icon.GetCommand(), icon.GetDesktopFileName())

	apps := strings.Split(origins, ";")

	var dir string
	if len(apps) > 0 {
		dir = filepath.Dir(apps[0])
		apps[0] = filepath.Base(apps[0])
		// log.DEV("Launcher origin", apps[0])
		// log.DETAIL(apps[1:])
	}

	desktop := icon.GetClassInfo(gldi.ClassDesktopFile)
	for k, v := range apps {
		flag := false
		if filepath.Join(dir, v) == desktop {
			flag = true
		}
		apps[k] = strings.TrimSuffix(apps[k], ".desktop")
		if flag {
			apps[k] = common.Bold(apps[k])
		}
	}

	str := "Magic launcher :" +
		"\nName :\t\t" + icon.GetClassInfo(gldi.ClassName) + // Must not use gldi, those consts will have to move.
		"\nIcon :\t\t" + icon.GetClassInfo(gldi.ClassIcon) +
		"\nCommand :\t" + icon.GetClassInfo(gldi.ClassCommand) +
		"\nDesktop file :\t" + strings.Join(apps, ", ")

	label, _ := gtk.LabelNew(str)
	label.Set("use-markup", true)
	return label
}

//-------------------------------------------------------[ WIDGET ICONS LIST ]--

// Liststore rows. Must match the ListStore declaration type and order.
//
const (
	rowConf = iota
	rowIcon
	rowName
)

// controlItems forwards events to other widgets.
//
type controlItems interface {
	onSelect(datatype.Iconer, error)
}

// row defines a pointer to link the icon object with its iter.
//
type row struct {
	Iter *gtk.TreeIter
	Icon datatype.Iconer
}

// List defines a dock icons management widget.
//
type List struct {
	gtk.ScrolledWindow // Main widget is the container. The ScrolledWindow will handle list scrollbars.
	tree               *gtk.TreeView
	model              *gtk.ListStore
	selection          *gtk.TreeSelection
	control            controlItems

	rows map[string]*row // maybe need to make if map[string] to get a ref to configfile.
}

// NewList creates a dock icons management widget.
//
func NewList(control controlItems) *List {
	builder := buildhelp.New()

	builder.AddFromString(string(conficonsXML()))
	// builder.AddFromFile("conficons.xml")

	widget := &List{
		ScrolledWindow: *builder.GetScrolledWindow("widget"),
		model:          builder.GetListStore("model"),
		tree:           builder.GetTreeView("tree"),
		selection:      builder.GetTreeSelection("selection"),
		control:        control,
		rows:           make(map[string]*row),
	}

	if len(builder.Errors) > 0 {
		for _, e := range builder.Errors {
			log.DEV("build conficons", e)
		}
		return nil
	}

	// Action: on Treeview Selected line.
	widget.selection.Connect("changed", widget.onSelectionChanged)

	return widget
}

// Load loads icons list in the widget.
//
func (widget *List) Load(icons map[string][]datatype.Iconer) {
	widget.Clear()

	for container, byOrder := range icons { // container, byOrder
		if container != datatype.KeyMainDock {
			log.Info("container dropped", container, "size:", len(byOrder))
			continue
		}

		for _, icon := range byOrder {
			iter := widget.model.Append()
			confPath := icon.ConfigPath()
			widget.rows[confPath] = &row{iter, icon} // also save local reference to icon so it's not garbage collected.

			name, img := icon.DefaultNameIcon()

			widget.model.SetCols(iter, gtk.Cols{
				rowName: name,
				rowConf: confPath,
			})

			if img != "" {
				if pix, e := common.PixbufNewFromFile(img, iconSize); !log.Err(e, "Load icon") {
					widget.model.SetValue(iter, rowIcon, pix)
				}
			}
		}
	}
}

// SelectedIcon returns the iconer matching the selected row.
//
func (widget *List) SelectedIcon() (datatype.Iconer, error) {
	key, e := widget.SelectedConf()
	if e != nil {
		return nil, e
	}
	icon, ok := widget.rows[key]
	if !ok {
		// TODO:  warn
		return nil, errors.New("no matching row found")
	}
	return icon.Icon, nil
}

// SelectedConf returns the path to icon config file for the selected row.
//
func (widget *List) SelectedConf() (string, error) {
	return gunvalue.SelectedValue(widget.model, widget.selection, rowConf).String()
}

// Select sets the selected icon based on its config file.
//
func (widget *List) Select(conf string) bool {
	row, ok := widget.rows[conf]
	if !ok {
		return false
	}
	widget.selection.SelectIter(row.Iter)
	return true

	// widget.tree.ScrollToCell : lol arguments are a joke, we'll see later.
}

// Clear clears the widget data.
//
func (widget *List) Clear() {
	widget.rows = make(map[string]*row)
	widget.model.Clear()
}

//-------------------------------------------------------[ ACTIONS CALLBACKS ]--

// Selected line has changed. Forward the call to the controler.
//
func (widget *List) onSelectionChanged(obj *gtk.TreeSelection) {
	widget.control.onSelect(widget.SelectedIcon())
}
