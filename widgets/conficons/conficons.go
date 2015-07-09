// Package conficons provides an icons list and configuration widget.
package conficons

import (
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/gldi"

	"github.com/sqp/godock/widgets/common"
	"github.com/sqp/godock/widgets/confbuilder"
	"github.com/sqp/godock/widgets/confbuilder/datatype"
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
	log  cdtype.Logger
}

// New creates a GuiIcons widget to edit cairo-dock icons config.
//
func New(data Controller, log cdtype.Logger, switcher *pageswitch.Switcher) *GuiIcons {
	paned, _ := gtk.PanedNew(gtk.ORIENTATION_HORIZONTAL)
	widget := &GuiIcons{
		Paned:    *paned,
		switcher: switcher,
		data:     data,
		log:      log,
	}
	widget.icons = NewList(widget, log)

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

	up.Connect("clicked", widget.actionSelected(datatype.Iconer.MoveBeforePrevious))
	down.Connect("clicked", widget.actionSelected(datatype.Iconer.MoveAfterNext))
	remove.Connect("clicked", widget.actionSelected(datatype.Iconer.RemoveFromDock))

	// widget.icons.Connect("row-inserted", func() { log.Info("row inserted") })
	// widget.icons.Connect("row-deleted", func() { log.Info("row deleted") })

	return widget
}

// actionSelected prepares a callback to act on the icon of the selected row.
//
func (widget *GuiIcons) actionSelected(call func(datatype.Iconer)) func() {
	return func() {
		ic, e := widget.icons.SelectedIcon()
		if e == nil {
			call(ic)
		}
	}
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
	path := widget.icons.SelectedConf()
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

	if e == nil {
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

	build, e := confbuilder.NewGrouper(widget.data, widget.log, widget.data.GetWindow(), icon.ConfigPath(), icon.GetGettextDomain())
	if widget.log.Err(e, "Load Keyfile "+icon.ConfigPath()) {
		return
	}
	name, _ := icon.DefaultNameIcon()
	switch {
	case icon.IsTaskbar():
		widget.config = build.BuildSingle("TaskBar")

	case name == "_desklets_":
		widget.config = build.BuildSingle("Desklets")

	default:
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

	if len(apps) > 0 {
		// Extract the path from the first item.
		dir := filepath.Dir(apps[0])
		apps[0] = filepath.Base(apps[0])

		// Remove suffix from apps names and highlight the active one.
		desktop := icon.GetClassInfo(gldi.ClassDesktopFile)
		for k, v := range apps {
			apps[k] = strings.TrimSuffix(apps[k], ".desktop")
			if filepath.Join(dir, v) == desktop {
				apps[k] = common.Bold(apps[k])
			}
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

// controlItems forwards events to other widgets.
//
type controlItems interface {
	onSelect(datatype.Iconer, error)
}

// List defines a dock icons management widget.
//
type List struct {
	gtk.ScrolledWindow // Main widget is the container. The ScrolledWindow will handle list scrollbars.
	list               *gtk.ListBox

	index map[*gtk.ListBoxRow]datatype.Iconer

	control controlItems // link to higher level widgets.
	log     cdtype.Logger
}

// NewList creates a dock icons management widget.
//
func NewList(control controlItems, log cdtype.Logger) *List {
	scroll, _ := gtk.ScrolledWindowNew(nil, nil)
	widget := &List{
		ScrolledWindow: *scroll,
		control:        control,
		log:            log,
		index:          make(map[*gtk.ListBoxRow]datatype.Iconer),
	}

	widget.list, _ = gtk.ListBoxNew()
	widget.list.Connect("row-selected", widget.onSelectionChanged)
	widget.Add(widget.list)

	return widget
}

// Load loads icons list in the widget.
//
func (widget *List) Load(icons *datatype.ListIcon) {
	widget.Clear()

	for _, iconContainer := range icons.Maindocks {
		widget.addBoxItem(iconContainer.Container, 0, true)
		widget.iconsDock(iconContainer.Icons, icons.Subdocks, 1)
	}

	widget.ShowAll()
}

// iconsDock adds a list of icons, and fills subdocks content if needed (recursive).
//
func (widget *List) iconsDock(icons []datatype.Iconer, subdocks map[string][]datatype.Iconer, indent int) {
	for _, icon := range icons {
		widget.addBoxItem(icon, indent, icon.IsStackIcon())

		if icon.IsStackIcon() {
			name, _ := icon.DefaultNameIcon()
			subicons, ok := subdocks[name]
			if ok {
				widget.iconsDock(subicons, subdocks, indent+1)
			}
		}
	}
}

// addBoxItem adds an item to the list.
//
func (widget *List) addBoxItem(icon datatype.Iconer, indent int, bold bool) *gtk.ListBoxRow {
	row, _ := gtk.ListBoxRowNew()
	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	row.Add(box)

	name, img := icon.DefaultNameIcon()

	box.Set("margin-start", 15*indent)
	if bold {
		name = common.Bold(name)
	}
	if img != "" {
		if pix, e := common.ImageNewFromFile(img, iconSize); !widget.log.Err(e, "Load icon") {
			box.PackStart(pix, false, false, 0)
		}
	}
	lbl, _ := gtk.LabelNew(name)
	lbl.SetUseMarkup(true)
	box.PackStart(lbl, false, false, 0)

	widget.list.Add(row)
	widget.index[row] = icon

	return row
}

// Clear clears the widget data.
//
func (widget *List) Clear() {
	for box := range widget.index {
		widget.list.Remove(box)
	}
	widget.index = make(map[*gtk.ListBoxRow]datatype.Iconer)
}

//---------------------------------------------------------------[ SELECTION ]--

// SelectedIcon returns the iconer matching the selected row.
//
func (widget *List) SelectedIcon() (datatype.Iconer, error) {
	sel := widget.list.GetSelectedRow()
	if sel == nil {
		return nil, errors.New("no selection")
	}
	for box, icon := range widget.index {
		if box.Native() == sel.Native() {
			return icon, nil
		}
	}
	return nil, errors.New("no matching icon for selection")
}

// SelectedConf returns the path to icon config file for the selected row.
//
func (widget *List) SelectedConf() string {
	icon, e := widget.SelectedIcon()
	if e != nil {
		return ""
	}
	return icon.ConfigPath()
}

// Select sets the selected icon based on its config file.
//
func (widget *List) Select(conf string) bool {
	if conf != "" {
		for box, icon := range widget.index {
			if icon.ConfigPath() == conf {
				box.Activate()
				return true
			}
		}
	}
	return false
}

//-------------------------------------------------------[ ACTIONS CALLBACKS ]--

// Selected line has changed. Forward the call to the controler.
//
func (widget *List) onSelectionChanged(box *gtk.ListBox, row *gtk.ListBoxRow) {
	widget.control.onSelect(widget.SelectedIcon())
}
