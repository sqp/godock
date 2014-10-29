// Package conficons provides an icons list and configuration widget.
package conficons

import (
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/godock/widgets/confbuilder"
	"github.com/sqp/godock/widgets/pageswitch"

	"github.com/sqp/godock/libs/gldi"
	"github.com/sqp/godock/libs/log"
	// "github.com/sqp/godock/libs/packages"

	"github.com/sqp/godock/widgets/common"
	"github.com/sqp/godock/widgets/confbuilder/datatype"
	"github.com/sqp/godock/widgets/gtk/buildhelp"
	"github.com/sqp/godock/widgets/gtk/gunvalue"

	"path/filepath"
	"strings"
)

const iconSize = 24
const listIconsWidth = 200

//--------------------------------------------------------[ PAGE GUI ICONS ]--

// ConfigWidget defines a GtkWidget with a Save method.
//
type ConfigWidget interface {
	gtk.IWidget
	Save()
}

// GuiIcons defines Icons configuration widget for currently actived cairo-dock Icons.
//
type GuiIcons struct {
	gtk.Paned

	icons    *List
	config   *confbuilder.Grouper
	page     ConfigWidget
	switcher *pageswitch.Switcher

	window *gtk.Window
	data   datatype.Source
}

// New creates a GuiIcons widget to edit cairo-dock icons config.
//
func New(data datatype.Source, switcher *pageswitch.Switcher) *GuiIcons {
	paned, _ := gtk.PanedNew(gtk.ORIENTATION_HORIZONTAL)
	widget := &GuiIcons{
		Paned:    *paned,
		switcher: switcher,
		data:     data,
	}
	widget.icons = NewList(widget)
	widget.Pack1(widget.icons, true, true)

	widget.SetPosition(listIconsWidth) // Paned position = list icons width.

	return widget
}

// SetWindow sets the pointer to the parent window, used for some config
// callbacks (grab events).
//
func (widget *GuiIcons) SetWindow(win *gtk.Window) {
	widget.window = win
}

// Load loads the list of icons in the iconsList.
//
func (widget *GuiIcons) Load() {
	icons := widget.data.ListIcons()
	widget.icons.Load(icons)
}

// Selected returns the selected icon.
//
func (widget *GuiIcons) Selected() datatype.Iconer {
	return widget.icons.Selected()
}

// Select sets the selected icon based on its config file.
//
func (widget *GuiIcons) Select(conf string) bool {
	return widget.icons.Select(conf)
}

// Clear clears the widget data.
//
func (widget *GuiIcons) Clear() string {
	path := ""
	sel := widget.Selected()
	if sel != nil {
		path = sel.ConfigPath()
	}
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

	icon := widget.Selected()

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
	icon.Reload()

	// 	_items_widget_reload (CD_WIDGET (pItemsWidget));  // we reload in case the items place has changed (icon's container, dock orientation, etc).

}

//
//-------------------------------------------------------[ CONTROL CALLBACKS ]--

// OnSelect reacts when a row is selected. Creates a new config for the icon.
//
func (widget *GuiIcons) OnSelect(icon datatype.Iconer) {
	widget.switcher.Clear()

	if widget.config != nil {
		widget.config.Destroy()
		widget.config = nil
	}

	if icon == nil { // shouldn't match.
		return
	}

	build, e := confbuilder.NewGrouper(widget.data, widget.window, icon.ConfigPath())
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
		"\nName :\t\t" + icon.GetClassInfo(gldi.ClassName) +
		"\nIcon :\t\t" + icon.GetClassInfo(gldi.ClassIcon) +
		"\nCommand :\t" + icon.GetClassInfo(gldi.ClassCommand) +
		"\nDesktop file :\t" + strings.Join(apps, ", ")

	label, _ := gtk.LabelNew(str)
	label.Set("use-markup", true)
	return label
}

//-------------------------------------------------------[ WIDGET ICONS LIST ]--

// Rows defines liststore rows. Must match the ListStore declaration type and order.
//
const (
	RowConf = iota
	RowIcon
	RowName
)

// ControlItems forwards events to other widgets.
//
type ControlItems interface {
	OnSelect(icon datatype.Iconer)
}

// Row defines a pointer to link the icon object with its iter.
//
type Row struct {
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
	control            ControlItems

	rows map[string]*Row // maybe need to make if map[string] to get a ref to configfile.
}

// NewList creates a dock icons management widget.
//
func NewList(control ControlItems) *List {
	builder := buildhelp.New()

	builder.AddFromString(string(conficonsXML()))
	// builder.AddFromFile("conficons.xml")

	widget := &List{
		ScrolledWindow: *builder.GetScrolledWindow("widget"),
		model:          builder.GetListStore("model"),
		tree:           builder.GetTreeView("tree"),
		selection:      builder.GetTreeSelection("selection"),
		control:        control,
		rows:           make(map[string]*Row),
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
	widget.model.Clear()
	widget.rows = make(map[string]*Row)

	for container, byOrder := range icons { // container, byOrder
		if container != datatype.KeyMainDock {
			log.Info("container dropped", container, "size:", len(byOrder))
			continue
		}

		for _, icon := range byOrder {
			iter := widget.model.Append()
			confPath := icon.ConfigPath()
			widget.rows[confPath] = &Row{iter, icon} // also save local reference to icon so it's not garbage collected.

			name, img := icon.DefaultNameIcon()

			widget.model.SetCols(iter, gtk.Cols{
				RowName: name,
				RowConf: confPath,
			})

			if img != "" {
				if pix, e := common.PixbufNewFromFile(img, iconSize); !log.Err(e, "Load icon") {
					widget.model.SetValue(iter, RowIcon, pix)
				}
			}
		}
	}
}

// Selected returns the Iconer matching the selected line.
//
func (widget *List) Selected() datatype.Iconer {
	if widget.selection.CountSelectedRows() == 0 {
		log.DEV("icons none selected")
		return nil
	}
	var iter gtk.TreeIter
	var treeModel gtk.ITreeModel = widget.model
	widget.selection.GetSelected(&treeModel, &iter)

	log.DEV("icons get selected")

	conf, e := gunvalue.New(widget.model.GetValue(&iter, RowConf)).String()
	if log.Err(e, "conficons.Selected") {
		return nil
	}
	icon, ok := widget.rows[conf]
	if !ok {
		// TODO:  warn
		return nil
	}
	return icon.Icon
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
	widget.rows = make(map[string]*Row)
	widget.model.Clear()

	// widget.tree.ScrollToCell : lol arguments are a joke, we'll see later.
}

//-------------------------------------------------------[ ACTIONS CALLBACKS ]--

// Selected line has changed. Forward the call to the controler.
//
func (widget *List) onSelectionChanged(obj *gtk.TreeSelection) {
	widget.control.OnSelect(widget.Selected())
}
