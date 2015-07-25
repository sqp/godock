// Package confcore provides a cairo-dock core configuration widget.
package confcore

import (
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/text/tran"

	"github.com/sqp/godock/widgets/common"
	"github.com/sqp/godock/widgets/confapplets"
	"github.com/sqp/godock/widgets/confbuilder"
	"github.com/sqp/godock/widgets/confbuilder/datatype"
	"github.com/sqp/godock/widgets/confsettings"
	"github.com/sqp/godock/widgets/confshortkeys"
	"github.com/sqp/godock/widgets/docktheme"
	"github.com/sqp/godock/widgets/gtk/buildhelp"
	"github.com/sqp/godock/widgets/gtk/gunvalue"
	"github.com/sqp/godock/widgets/pageswitch"

	"errors"
	"path/filepath"
)

const iconSize = 24
const panedPosition = 200

const (
	// TabDownload is the name of the config download tab.
	TabDownload = "Download"
	// TabShortkeys is the name of the config shortkeys tab.
	TabShortkeys = "Shortkeys"
	// TabThemes is the name of the config themes tab.
	TabThemes = "Themes"
)

//
//-----------------------------------------------------------[ CONFCORE DATA ]--

// Item defines a core config page information.
//
type Item struct {
	Key      string
	Title    string
	Icon     string
	Tooltip  string
	Managers []string
	// File string
	// Group string // group in file.
}

var coreItems = []*Item{
	{
		Key:      "Position",
		Title:    "Position",
		Icon:     "icons/icon-position.svg",
		Tooltip:  "Set the position of the main dock.",
		Managers: []string{"Docks"}},

	{
		Key:      "Accessibility",
		Title:    "Visibility",
		Icon:     "icons/icon-visibility.svg",
		Tooltip:  "Do you like your dock to be always visible,\n or on the contrary unobtrusive?\nConfigure the way you access your docks and sub-docks!",
		Managers: []string{"Docks"}},

	{
		Key:      "TaskBar",
		Title:    "TaskBar",
		Icon:     "icons/icon-taskbar.png",
		Tooltip:  "Display and interact with currently open windows.",
		Managers: []string{"Taskbar"}},

	{
		Key:      "Style",
		Title:    "Style",
		Icon:     "icons/icon-style.svg",
		Tooltip:  "Configure the global style.",
		Managers: []string{"Style"}},

	{
		Key:      "Background",
		Title:    "Background",
		Icon:     "icons/icon-background.svg",
		Tooltip:  "Configure docks appearance.",
		Managers: []string{"Backends", "Docks"}}, // -> "dock rendering"

	{
		Key:      "Views",
		Title:    "Views",
		Icon:     "icons/icon-views.svg",
		Tooltip:  "Configure docks appearance.",  // same as background (were grouped).
		Managers: []string{"Backends", "Docks"}}, // -> "dock rendering"

	{
		Key:      "Dialogs",
		Title:    "Dialog boxes and Menus",
		Icon:     "icons/icon-dialogs.svg",
		Tooltip:  "Configure text bubble appearance.",
		Managers: []string{"Dialogs"}}, // -> "dialog rendering"

	{
		Key:      "Desklets",
		Title:    "Desklets",
		Icon:     "icons/icon-desklets.svg",
		Tooltip:  "Applets can be displayed on your desktop as widgets.",
		Managers: []string{"Desklets"}}, // -> "desklet rendering"

	{
		Key:      "Icons",
		Title:    "Icons",
		Icon:     "icons/icon-icons.svg",
		Tooltip:  "All about icons:\n size, reflection, icon theme,...",
		Managers: []string{"Icons", "Indicators"}}, // indicators needed here too ?

	{
		Key:      "Labels",
		Title:    "Captions",
		Icon:     "icons/icon-labels.svg",
		Tooltip:  "Define icon caption and quick-info style.",
		Managers: []string{"Icons"}},

	{
		Key:     TabThemes,
		Title:   TabThemes,
		Icon:    "icons/icon-appearance.svg", // icon-controler.svg
		Tooltip: "Try new themes and save your theme."},

	{
		Key:     TabShortkeys,
		Title:   TabShortkeys,
		Icon:    "icons/icon-shortkeys.svg",
		Tooltip: "Define all the keyboard shortcuts currently available."},

	{
		Key:      "System",
		Title:    "System",
		Icon:     "icons/icon-system.svg",
		Tooltip:  "All of the parameters you will never want to tweak.",
		Managers: []string{"Backends", "Containers", "Connection", "Docks"}},

	{
		Key:     TabDownload,
		Title:   TabDownload,
		Icon:    "cairo-dock.svg",
		Tooltip: "Download additional applets."},

	{
		Key:   confsettings.GuiGroup, // custom page for the config own settings.
		Title: confsettings.GuiGroup,
		Icon:  "cairo-dock.svg"},

	// {
	// 	Key:     "Help",
	// 	Title:   "Help",
	// 	Icon:    "plug-ins/Help/icon.svg",
	// 	Tooltip: "Try new themes and save your theme."},

	// + icon effects*
	// _add_sub_group_to_group_button (pGroupDescription, "Indicators", "icon-indicators.svg", _("Indicators"));

}

//
//--------------------------------------------------------[ PAGE GUI ICONS ]--

// Controller defines methods used on the main widget / data source by this widget and its sons.
//
type Controller interface {
	datatype.Source
	GetWindow() *gtk.Window
	SetActionNone()
	SetActionSave()
	SetActionGrab()
	SetActionCancel()

	SelectIcons(string)
}

// configWidget extends the Widget interface with common needed actions.
//
type configWidget interface {
	gtk.IWidget
	Destroy()
	ShowAll()
}

type grabber interface {
	configWidget
	Load()
	Grab()
}

type saver interface {
	configWidget
	confbuilder.KeyFiler
	Save()
}

// ConfCore provides a configuration widget for the main cairo-dock config.
//
type ConfCore struct {
	gtk.Paned

	list     *List
	config   configWidget
	switcher *pageswitch.Switcher

	data Controller
	log  cdtype.Logger
}

// New creates a ConfCore widget to edit the main cairo-dock config.
//
func New(data Controller, log cdtype.Logger, switcher *pageswitch.Switcher) *ConfCore {
	paned, _ := gtk.PanedNew(gtk.ORIENTATION_HORIZONTAL)

	widget := &ConfCore{
		Paned:    *paned,
		switcher: switcher,
		data:     data,
		log:      log,
	}
	widget.list = NewList(widget, log)
	widget.Pack1(widget.list, true, true)

	widget.SetPosition(panedPosition)
	return widget
}

// Load loads config items in the widget.
//
func (widget *ConfCore) Load() {
	widget.list.Load(widget.data.DirShareData())
}

// Selected returns the selected page config.
//
// func (widget *ConfCore) Selected() (*Item, error) {
// 	return widget.list.Selected()
// }

// func (widget *ConfCore) Clean() {
// }

//--------------------------------------------------------[ SAVE CONFIG PAGE ]--

// Save saves the current page configuration
//
func (widget *ConfCore) Save() {
	if widget.config == nil {
		return
	}

	if grab := widget.grabber(); grab != nil {
		grab.Grab()
		return
	}

	if saver, ok := widget.config.(saver); ok {
		saver.Save()

		item, e := widget.list.Selected()
		if widget.log.Err(e, "Save: selection problem") {
			return
		}

		for _, manager := range item.Managers {
			widget.data.ManagerReload(manager, true, saver.KeyFile())
		}
	}
	// //\_____________ reload modules that are concerned by these changes
	// GldiManager *pManager;
	// if (bUpdateColors)
	// {
	// 	cd_reload ("Style");
	// 	cd_reload ("Indicators");
	// 	cd_reload ("Dialogs");
	// 	GldiModule *pModule;
	// 	pModule = gldi_module_get ("clock");
	// 	if (pModule)
	// 		gldi_object_reload (GLDI_OBJECT (pModule), TRUE);
	// 	pModule = gldi_module_get ("keyboard indicator");
	// 	if (pModule)
	// 		gldi_object_reload (GLDI_OBJECT (pModule), TRUE);
	// 	pModule = gldi_module_get ("dock rendering");
	// 	if (pModule)
	// 		gldi_object_reload (GLDI_OBJECT (pModule), TRUE);
	// }
	// cd_reload ("Docks");
	// cd_reload ("Taskbar");
	// cd_reload ("Icons");
	// cd_reload ("Backends");

	// if (pModuleInstanceAnim != NULL)
	// {
	// 	gldi_object_reload (GLDI_OBJECT(pModuleInstanceAnim), TRUE);
	// }
	// if (pModuleInstanceEffect != NULL)
	// {
	// 	gldi_object_reload (GLDI_OBJECT(pModuleInstanceEffect), TRUE);
	// }
	// if (pModuleInstanceIllusion != NULL)
	// {
	// 	gldi_object_reload (GLDI_OBJECT(pModuleInstanceIllusion), TRUE);
	// }
}

func (widget *ConfCore) grabber() grabber {
	if widget.config != nil {
		grab, ok := widget.config.(grabber)
		if ok {
			return grab
		}
	}
	return nil
}

//-------------------------------------------------------[ CONTROL CALLBACKS ]--

// onSelect acts as a row selected callback.
//
func (widget *ConfCore) onSelect(item *Item, e error) {
	if widget.log.Err(e, "onSelect: selection problem") {
		return
	}
	widget.switcher.Clear()

	if widget.config != nil {
		widget.config.Destroy()
		widget.config = nil
	}

	widget.SetAction()

	file := ""
	def := ""
	switch item.Key {
	case TabShortkeys:
		w := confshortkeys.New(widget.data, widget.log)
		w.Load()
		widget.Pack2(w, true, true)
		widget.config = w
		return

	case TabThemes:
		widget.config = docktheme.New(widget.data, widget.log, widget.switcher)
		widget.Pack2(widget.config, true, true)
		return

	case TabDownload: // download tab has a special widget.
		w := confapplets.New(widget.data, widget.log, nil, confapplets.ListExternal)
		w.Load()
		w.ShowAll()

		widget.Pack2(w, true, true)
		widget.config = w
		return

	case confsettings.GuiGroup: // own config has a special path.
		file = confsettings.PathFile()

	default:
		file = widget.data.MainConfigFile()
		def = widget.data.MainConfigDefault()
	}

	build, e := confbuilder.NewGrouper(widget.data, widget.log, widget.data.GetWindow(), file, def, "")
	if widget.log.Err(e, "Load Keyfile "+file) {
		return
	}
	widget.config = build.BuildSingle(item.Key)

	widget.Pack2(widget.config, true, true)
	widget.config.ShowAll()
}

// SetAction sets the action button name (save or grab).
//
func (widget *ConfCore) SetAction() {
	item, e := widget.list.Selected()
	switch {
	case e != nil: // Do nothing. Should be triggered only on load, before any user selection.

	case item.Key == TabShortkeys:
		widget.data.SetActionGrab()

	case item.Key == TabDownload:
		widget.data.SetActionNone()

	default:
		widget.data.SetActionSave()
	}
}

// UpdateModuleState updates the state of the given applet, from a dock event.
//
func (widget *ConfCore) UpdateModuleState(name string, active bool) {
	if widget.config == nil {
		return
	}
	confapp, ok := widget.config.(datatype.UpdateModuleStater)
	if !ok {
		return
	}

	confapp.UpdateModuleState(name, active)
}

// UpdateShortkeys updates the shortkey widget if it's loaded.
//
func (widget *ConfCore) UpdateShortkeys() {
	widget.log.Info("UpdateShortkeys")
	// conf, e := widget.list.Selected()
	// log.Err(e, "confcore selected")

	if grab := widget.grabber(); grab != nil {
		grab.Load()
		// if e == nil && conf.Key == TabShortkeys {
		// widget.config.Load()
	}
}

//-------------------------------------------------------[ WIDGET CORE LIST ]--

// Liststore rows. Must match the ListStore declaration type and order.
const (
	rowKey = iota
	rowIcon
	rowName
	rowTooltip
)

// controlItems forwards events to other widgets.
//
type controlItems interface {
	onSelect(*Item, error)
}

// row defines a pointer to link the icon name with its iter.
//
type row struct {
	Iter *gtk.TreeIter
	Conf *Item
}

// List is a widget to list and select cairo-dock main config pages references.
//
type List struct {
	gtk.ScrolledWindow // Main widget is the container. The ScrolledWindow will handle list scrollbars.
	tree               *gtk.TreeView
	model              *gtk.ListStore
	selection          *gtk.TreeSelection
	control            controlItems
	log                cdtype.Logger

	rows map[string]*row
}

// NewList creates cairo-dock main config pages list.
//
func NewList(control controlItems, log cdtype.Logger) *List {
	builder := buildhelp.NewFromBytes(confcoreXML())

	widget := &List{
		ScrolledWindow: *builder.GetScrolledWindow("widget"),
		model:          builder.GetListStore("model"),
		tree:           builder.GetTreeView("tree"),
		selection:      builder.GetTreeSelection("selection"),
		control:        control,
		log:            log,
		rows:           make(map[string]*row),
	}

	if len(builder.Errors) > 0 {
		for _, e := range builder.Errors {
			log.Err(e, "build confcore list")
		}
		return nil
	}

	// Action: Treeview Select line.
	widget.selection.Connect("changed", widget.onSelectionChanged)

	return widget
}

// Load loads the widget fields.
//
func (widget *List) Load(shareData string) {
	for _, item := range coreItems {
		iter := widget.model.Append()
		widget.rows[item.Key] = &row{iter, item}

		args := gtk.Cols{
			rowKey:     item.Key,
			rowName:    tran.Slate(item.Title),
			rowTooltip: item.Tooltip,
		}
		widget.model.SetCols(iter, args)

		if item.Icon != "" {
			path := filepath.Join(shareData, item.Icon)
			if pix, e := common.PixbufNewFromFile(path, iconSize); !widget.log.Err(e, "Load icon") {
				widget.model.SetValue(iter, rowIcon, pix)
			}
		}
	}
}

// Selected returns the data about the selected item.
//
func (widget *List) Selected() (*Item, error) {
	key, e := gunvalue.SelectedValue(widget.model, widget.selection, rowKey).String()
	if e != nil {
		return nil, e
	}
	conf, ok := widget.rows[key]
	if !ok {
		// TODO:  warn
		return nil, errors.New("no matching row found")
	}
	return conf.Conf, nil
}

//-------------------------------------------------------[ ACTIONS CALLBACKS ]--

// Selected line has changed. Forward the call to the controler.
//
func (widget *List) onSelectionChanged(obj *gtk.TreeSelection) {
	widget.control.onSelect(widget.Selected())
}

//

/*
		_("Icons"));
	_add_sub_group_to_group_button (pGroupDescription, "Icons", "icon-icons.svg", _("Icons"));
	_add_sub_group_to_group_button (pGroupDescription, "Indicators", "icon-indicators.svg", _("Indicators"));
	pGroupDescription->pManagers = g_list_prepend (pGroupDescription->pManagers, (gchar*)"Icons");
	pGroupDescription->pManagers = g_list_prepend (pGroupDescription->pManagers, (gchar*)"Indicators");  // -> "drop indicator"

*/
