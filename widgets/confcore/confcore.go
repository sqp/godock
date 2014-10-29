// Package confcore provides a cairo-dock core configuration widget.
package confcore

import (
	"github.com/conformal/gotk3/gtk"
	// "github.com/gosexy/gettext"

	"github.com/sqp/godock/widgets/common"
	"github.com/sqp/godock/widgets/confbuilder"
	"github.com/sqp/godock/widgets/confbuilder/datatype"
	"github.com/sqp/godock/widgets/confsettings"
	"github.com/sqp/godock/widgets/gtk/buildhelp"
	"github.com/sqp/godock/widgets/gtk/gunvalue"
	"github.com/sqp/godock/widgets/pageswitch"

	"github.com/sqp/godock/libs/log"

	"path/filepath"
)

const iconSize = 24
const panedPosition = 200

//
//-----------------------------------------------------------[ CONFCORE DATA ]--

// ConfCore defines a core config page information.
//
type ConfCore struct {
	Key      string
	Title    string
	Icon     string
	Managers []string
	// File string
	// Group string // group in file.
}

var items = []*ConfCore{
	{
		Key:      "Position",
		Title:    "Position",
		Icon:     "icons/icon-position.svg",
		Managers: []string{"Docks"}},

	{
		Key:      "Accessibility",
		Title:    "Visibility",
		Icon:     "icons/icon-visibility.svg",
		Managers: []string{"Docks"}},

	{
		Key:      "TaskBar",
		Title:    "TaskBar",
		Icon:     "icons/icon-taskbar.png",
		Managers: []string{"Taskbar"}},

	{
		Key:      "Style",
		Title:    "Style",
		Icon:     "icons/icon-style.svg",
		Managers: []string{"Style"}},

	{
		Key:      "Background",
		Title:    "Background",
		Icon:     "icons/icon-background.svg",
		Managers: []string{"Docks"}},

	{
		Key:      "Views",
		Title:    "Views",
		Icon:     "icons/icon-views.svg",
		Managers: []string{"Docks", "Backends"}}, // -> "dock rendering"

	{
		Key:      "Dialogs",
		Title:    "Dialogs",
		Icon:     "icons/icon-dialogs.svg",
		Managers: []string{"Dialogs"}}, // -> "dialog rendering"

	{
		Key:      "Desklets",
		Title:    "Desklets",
		Icon:     "icons/icon-desklets.svg",
		Managers: []string{"Desklets"}}, // -> "desklet rendering"

	{
		Key:      "Icons",
		Title:    "Icons",
		Icon:     "icons/icon-icons.svg",
		Managers: []string{"Icons", "Indicators"}}, // indicators needed here too ?

	{
		Key:      "Labels",
		Title:    "Labels",
		Icon:     "icons/icon-labels.svg",
		Managers: []string{"Icons"}},

	// {
	// 	Key:   "Themes",
	// 	Title: "Themes",
	// 	Icon:  "icons/icon-controler.svg"},

	// {
	// 	Key:   "Shortkeys",
	// 	Title: "Shortkeys",
	// 	Icon:  "icons/icon-shortkeys.svg"},

	{
		Key:      "System",
		Title:    "System",
		Icon:     "icons/icon-system.svg",
		Managers: []string{"Docks", "Connection", "Containers", "Backends"}},

	{
		Key:   confsettings.GuiGroup, // custom page for the config own settings.
		Title: confsettings.GuiGroup,
		Icon:  "cairo-dock.svg"},

	// + icon effects*
	// _add_sub_group_to_group_button (pGroupDescription, "Indicators", "icon-indicators.svg", _("Indicators"));

}

//
//--------------------------------------------------------[ PAGE GUI ICONS ]--

// ConfigWidget extends the Widget interface with a Save action.
//
type ConfigWidget interface {
	gtk.IWidget
	Save()
}

// GuiMain provides a configuration widget for the main cairo-dock config.
//
type GuiMain struct {
	gtk.Paned

	// Applets *map[string]*packages.AppletPackage // List of applets known by the Dock.

	icons    *List
	config   *confbuilder.Grouper
	page     ConfigWidget
	switcher *pageswitch.Switcher

	window *gtk.Window
	data   datatype.Source
}

// New creates a GuiMain widget to edit the main cairo-dock config.
//
func New(data datatype.Source, switcher *pageswitch.Switcher) *GuiMain {
	paned, _ := gtk.PanedNew(gtk.ORIENTATION_HORIZONTAL)

	widget := &GuiMain{
		Paned:    *paned,
		switcher: switcher,
		data:     data,
	}
	widget.icons = NewList(widget)
	widget.Pack1(widget.icons, true, true)

	widget.SetPosition(panedPosition)
	return widget
}

// SetWindow sets the pointer to the parent window, used for some config
// callbacks (grab events).
//
func (widget *GuiMain) SetWindow(win *gtk.Window) {
	widget.window = win
}

// Load loads config items in the widget.
//
func (widget *GuiMain) Load() {
	widget.icons.Load(widget.data.DirShareData())
}

// Selected returns the selected page config.
//
func (widget *GuiMain) Selected() *ConfCore {
	return widget.icons.Selected()
}

// func (widget *GuiMain) Clean() {
// }

//--------------------------------------------------------[ SAVE CONFIG PAGE ]--

// Save saves the current page configuration
//
func (widget *GuiMain) Save() {
	// log.DEV("SAVE")
	if widget.config == nil {
		return
	}
	widget.config.Save()

	item := widget.Selected()
	for _, manager := range item.Managers {
		widget.data.ManagerReload(manager, true, widget.config.Conf.KeyFile)
	}

	// widget.data.ManagerReload("Style", true, widget.config.Conf.KeyFile)
	// widget.data.ManagerReload("Indicators", true, widget.config.Conf.KeyFile)
	// widget.data.ManagerReload("Dialogs", true, widget.config.Conf.KeyFile)
	// widget.data.ManagerReload("Docks", true, widget.config.Conf.KeyFile)
	// widget.data.ManagerReload("Taskbar", true, widget.config.Conf.KeyFile)
	// widget.data.ManagerReload("Icons", true, widget.config.Conf.KeyFile)
	// widget.data.ManagerReload("Backends", true, widget.config.Conf.KeyFile)

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

//-------------------------------------------------------[ CONTROL CALLBACKS ]--

// OnSelect acts as a row selected callback.
//
func (widget *GuiMain) OnSelect(item *ConfCore) {
	// widget.switcher.Clear() // unused yet

	if widget.config != nil {
		widget.config.Destroy()
		widget.config = nil
	}

	file := ""
	if item.Key == confsettings.GuiGroup {
		file = confsettings.PathFile()
	} else {
		file = widget.data.MainConf()
	}

	build, e := confbuilder.NewGrouper(widget.data, widget.window, file)
	if log.Err(e, "Load Keyfile "+file) {
		return
	}
	widget.config = build.BuildSingle(item.Key)

	widget.Pack2(widget.config, true, true)
	widget.config.ShowAll()
}

//-------------------------------------------------------[ WIDGET ICONS LIST ]--

// Rows defines liststore rows. Must match the ListStore declaration type and order.
const (
	RowKey = iota
	RowIcon
	RowName
	RowTooltip
)

// ControlItems forwards events to other widgets.
//
type ControlItems interface {
	OnSelect(item *ConfCore)
}

// Row defines a pointer to link the icon name with its iter.
//
type Row struct {
	Iter *gtk.TreeIter
	Conf *ConfCore
}

// List is a widget to select cairo-dock main config pages.
//
type List struct {
	gtk.ScrolledWindow // Main widget is the container. The ScrolledWindow will handle list scrollbars.
	tree               *gtk.TreeView
	model              *gtk.ListStore
	selection          *gtk.TreeSelection
	control            ControlItems

	rows map[string]*Row // maybe need to make if map[string] to get a ref to configfile.
}

// NewList creates cairo-dock main config pages list.
//
func NewList(control ControlItems) *List {
	builder := buildhelp.New()

	builder.AddFromString(string(confcoreXML()))
	// builder.AddFromFile("confcore.xml")

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
			log.DEV("build configMain", e)
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
	for _, item := range items {
		iter := widget.model.Append()
		widget.rows[item.Key] = &Row{iter, item}

		args := gtk.Cols{
			RowKey:  item.Key,
			RowName: item.Title,
		}
		widget.model.SetCols(iter, args)

		if item.Icon != "" {
			path := filepath.Join(shareData, item.Icon)
			if pix, e := common.PixbufNewFromFile(path, iconSize); !log.Err(e, "Load icon") {
				widget.model.SetValue(iter, RowIcon, pix)
			}
		}
	}
}

// Selected returns the name of the selected item.
//
func (widget *List) Selected() *ConfCore {
	var iter gtk.TreeIter
	var treeModel gtk.ITreeModel = widget.model
	widget.selection.GetSelected(&treeModel, &iter)

	key, e := gunvalue.New(widget.model.GetValue(&iter, RowKey)).String()
	if log.Err(e, "configMain.Selected") {
		return nil
	}
	conf, ok := widget.rows[key]
	if !ok {
		// TODO:  warn
		return nil
	}
	return conf.Conf
}

//-------------------------------------------------------[ ACTIONS CALLBACKS ]--

// Selected line has changed. Forward the call to the controler.
//
func (widget *List) onSelectionChanged(obj *gtk.TreeSelection) {
	widget.control.OnSelect(widget.Selected())
}

//

/*
	pGroupDescription = _add_one_main_group_button ("Position",
		CAIRO_DOCK_SHARE_DATA_DIR"/icons/icon-position.svg",
		CAIRO_DOCK_CATEGORY_BEHAVIOR,
		N_("Set the position of the main dock."),
		_("Position"));
	_add_sub_group_to_group_button (pGroupDescription, "Position", "icon-position.svg", _("Position"));
	pGroupDescription->pManagers = g_list_prepend (pGroupDescription->pManagers, (gchar*)"Docks");
	pGroupDescription->build_widget = _build_config_group_widget;

	pGroupDescription = _add_one_main_group_button ("Accessibility",
		CAIRO_DOCK_SHARE_DATA_DIR"/icons/icon-visibility.svg",
		CAIRO_DOCK_CATEGORY_BEHAVIOR,
		N_("Do you like your dock to be always visible,\n or on the contrary unobtrusive?\nConfigure the way you access your docks and sub-docks!"),
		_("Visibility"));
	_add_sub_group_to_group_button (pGroupDescription, "Accessibility", "icon-visibility.svg", _("Visibility"));
	pGroupDescription->pManagers = g_list_prepend (pGroupDescription->pManagers, (gchar*)"Docks");
	pGroupDescription->build_widget = _build_config_group_widget;

	pGroupDescription = _add_one_main_group_button ("TaskBar",
		CAIRO_DOCK_SHARE_DATA_DIR"/icons/icon-taskbar.png",
		CAIRO_DOCK_CATEGORY_BEHAVIOR,
		N_("Display and interact with currently open windows."),
		_("Taskbar"));
	_add_sub_group_to_group_button (pGroupDescription, "TaskBar", "icon-taskbar.png", _("Taskbar"));
	pGroupDescription->pManagers = g_list_prepend (pGroupDescription->pManagers, (gchar*)"Taskbar");
	pGroupDescription->build_widget = _build_config_group_widget;

	pGroupDescription = _add_one_main_group_button ("Shortkeys",
		CAIRO_DOCK_SHARE_DATA_DIR"/icons/icon-shortkeys.svg",
		CAIRO_DOCK_CATEGORY_BEHAVIOR,
		N_("Define all the keyboard shortcuts currently available."),
		_("Shortkeys"));
	pGroupDescription->build_widget = _build_shortkeys_widget;

	pGroupDescription = _add_one_main_group_button ("System",
		CAIRO_DOCK_SHARE_DATA_DIR"/icons/icon-system.svg",
		CAIRO_DOCK_CATEGORY_BEHAVIOR,
		N_("All of the parameters you will never want to tweak."),
		_("System"));
	_add_sub_group_to_group_button (pGroupDescription, "System", "icon-system.svg", _("System"));
	pGroupDescription->pManagers = g_list_prepend (pGroupDescription->pManagers, (gchar*)"Docks");
	pGroupDescription->pManagers = g_list_prepend (pGroupDescription->pManagers, (gchar*)"Connection");
	pGroupDescription->pManagers = g_list_prepend (pGroupDescription->pManagers, (gchar*)"Containers");
	pGroupDescription->pManagers = g_list_prepend (pGroupDescription->pManagers, (gchar*)"Backends");
	pGroupDescription->build_widget = _build_config_group_widget;

	pGroupDescription = _add_one_main_group_button ("Style",
		CAIRO_DOCK_SHARE_DATA_DIR"/icons/icon-style.svg",
		CAIRO_DOCK_CATEGORY_THEME,
		N_("Configure the global style."),
		_("Style"));
	_add_sub_group_to_group_button (pGroupDescription, "Style", "icon-style.svg", _("Style"));
	pGroupDescription->pManagers = g_list_prepend (pGroupDescription->pManagers, (gchar*)"Style");
	pGroupDescription->build_widget = _build_config_group_widget;

	pGroupDescription = _add_one_main_group_button ("Background",
		CAIRO_DOCK_SHARE_DATA_DIR"/icons/icon-docks.svg",
		CAIRO_DOCK_CATEGORY_THEME,
		N_("Configure docks appearance."),
		_("Docks"));
	_add_sub_group_to_group_button (pGroupDescription, "Background", "icon-background.svg", _("Background"));
	_add_sub_group_to_group_button (pGroupDescription, "Views", "icon-views.svg", _("Views"));
	pGroupDescription->pManagers = g_list_prepend (pGroupDescription->pManagers, (gchar*)"Docks");
	pGroupDescription->pManagers = g_list_prepend (pGroupDescription->pManagers, (gchar*)"Backends");  // -> "dock rendering"
	pGroupDescription->build_widget = _build_config_group_widget;

	pGroupDescription = _add_one_main_group_button ("Dialogs",
		CAIRO_DOCK_SHARE_DATA_DIR"/icons/icon-dialogs.svg",
		CAIRO_DOCK_CATEGORY_THEME,
		N_("Configure text bubble appearance."),
		_("Dialog boxes and Menus"));
	_add_sub_group_to_group_button (pGroupDescription, "Dialogs", "icon-dialogs.svg", _("Dialog boxes and Menus"));
	pGroupDescription->pManagers = g_list_prepend (pGroupDescription->pManagers, (gchar*)"Dialogs");  // -> "dialog rendering"
	pGroupDescription->build_widget = _build_config_group_widget;

	pGroupDescription = _add_one_main_group_button ("Desklets",
		CAIRO_DOCK_SHARE_DATA_DIR"/icons/icon-desklets.svg",
		CAIRO_DOCK_CATEGORY_THEME,
		N_("Applets can be displayed on your desktop as widgets."),
		_("Desklets"));
	_add_sub_group_to_group_button (pGroupDescription, "Desklets", "icon-desklets.svg", _("Desklets"));
	pGroupDescription->pManagers = g_list_prepend (pGroupDescription->pManagers, (gchar*)"Desklets");  // -> "desklet rendering"
	pGroupDescription->build_widget = _build_config_group_widget;

	pGroupDescription = _add_one_main_group_button ("Icons",
		CAIRO_DOCK_SHARE_DATA_DIR"/icons/icon-icons.svg",
		CAIRO_DOCK_CATEGORY_THEME,
		N_("All about icons:\n size, reflection, icon theme,..."),
		_("Icons"));
	_add_sub_group_to_group_button (pGroupDescription, "Icons", "icon-icons.svg", _("Icons"));
	_add_sub_group_to_group_button (pGroupDescription, "Indicators", "icon-indicators.svg", _("Indicators"));
	pGroupDescription->pManagers = g_list_prepend (pGroupDescription->pManagers, (gchar*)"Icons");
	pGroupDescription->pManagers = g_list_prepend (pGroupDescription->pManagers, (gchar*)"Indicators");  // -> "drop indicator"
	pGroupDescription->build_widget = _build_config_group_widget;

	pGroupDescription = _add_one_main_group_button ("Labels",
		CAIRO_DOCK_SHARE_DATA_DIR"/icons/icon-labels.svg",
		CAIRO_DOCK_CATEGORY_THEME,
		N_("Define icon caption and quick-info style."),
		_("Captions"));
	_add_sub_group_to_group_button (pGroupDescription, "Labels", "icon-labels.svg", _("Captions"));
	pGroupDescription->pManagers = g_list_prepend (NULL, (gchar*)"Icons");
	pGroupDescription->build_widget = _build_config_group_widget;

	pGroupDescription = _add_one_main_group_button ("Themes",
		CAIRO_DOCK_SHARE_DATA_DIR"/icons/icon-controler.svg",  /// TODO: find an icon...
		CAIRO_DOCK_CATEGORY_THEME,
		N_("Try new themes and save your theme."),
		_("Themes"));
	pGroupDescription->build_widget = _build_themes_widget;

	pGroupDescription = _add_one_main_group_button ("Items",
		CAIRO_DOCK_SHARE_DATA_DIR"/icons/icon-all.svg",  /// TODO: find an icon...
		CAIRO_DOCK_CATEGORY_THEME,
		N_("Current items in your dock(s)."),
		_("Current items"));
	pGroupDescription->build_widget = _build_items_widget;
}

*/
