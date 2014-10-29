// +build dock

package main

import (
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/godock/libs/gldi"
	"github.com/sqp/godock/libs/gldi/globals"
	"github.com/sqp/godock/libs/gldi/maindock"
	"github.com/sqp/godock/widgets/confbuilder/datagldi"
	"github.com/sqp/godock/widgets/confbuilder/datatype"
	"github.com/sqp/godock/widgets/confgui"

	// loader
	"github.com/sqp/godock/libs/srvdbus"
	"github.com/sqp/godock/services/allapps"

	// web inspection.
	// "net/http"
	// _ "net/http/pprof"
)

func init() {
	cmdDock = &Command{
		Run:       runDock,
		UsageLine: "dock",
		Short:     "dock starts a custom version of cairo-dock with a new config GUI.",
		Long: `
Dock starts a custom version of cairo-dock with a new GUI.

Options:
  -c          Use Cairo backend.
  -o          Use OpenGL backend.
  -O          Use OpenGL backend with indirect rendering.
              There are very few case where this option should be used.
  -A          Ask again on startup which backend to use.
  -e env      Force the dock to consider this environnement - use it with care.

  -d path     Use a custom config directory. Default: ~/.config/cairo-dock
  -S url      Address of a server with additional themes (overrides default).

  -w time     Wait for N seconds before starting; this is useful if you notice
              some problems when the dock starts with the session.
  -x appname  Exclude a given plug-in from activating (it is still loaded).
  -f          Safe mode: don't load any plug-ins.
  -W          Work around some bugs in Metacity Window-Manager
              (invisible dialogs or sub-docks)
  -l level    Log verbosity (debug,message,warning,critical,error).
              Default is warning.
  -F          Force to display some output messages with colors.
  -k          Lock the dock so that any modification is impossible for users.
  -a          Keep the dock above other windows whatever.
  -s          Don't make the dock appear on all desktops.
  -M path     Ask the dock to load additionnal modules from this directory.
              (though it is unsafe for your dock to load unnofficial modules).

  -v          Print version.

This version lacks a lot of options and may not be considered usable for
everybody at the moment.
.`,
	}
	userForceCairo = cmdDock.Flag.Bool("c", false, "")
	userForceOpenGL = cmdDock.Flag.Bool("o", false, "")
	userIndirectOpenGL = cmdDock.Flag.Bool("O", false, "")
	userAskBackend = cmdDock.Flag.Bool("A", false, "")
	userEnv = cmdDock.Flag.String("e", "", "")

	userDir = cmdDock.Flag.String("d", "", "")
	userThemeServer = cmdDock.Flag.String("S", "", "")

	userDelay = cmdDock.Flag.Int("w", 0, "")
	// maintenance
	userExclude = cmdDock.Flag.String("x", "", "")
	userSafeMode = cmdDock.Flag.Bool("f", false, "")
	userMetacityWorkaround = cmdDock.Flag.Bool("W", false, "")
	userVerbosity = cmdDock.Flag.String("l", "", "")
	userForceColor = cmdDock.Flag.Bool("F", false, "")
	userVersion = cmdDock.Flag.Bool("v", false, "")
	userLocked = cmdDock.Flag.Bool("k", false, "")
	userKeepAbove = cmdDock.Flag.Bool("a", false, "")
	userNoSticky = cmdDock.Flag.Bool("s", false, "")
	userModulesDir = cmdDock.Flag.String("M", "", "")
}

// 	{"maintenance", 'm', G_OPTION_FLAG_IN_MAIN, G_OPTION_ARG_NONE,
// 		&bMaintenance,
// 		_("Allow to edit the config before the dock is started and show the config panel on start."), NULL},
// 	{"exclude", 'x', G_OPTION_FLAG_IN_MAIN, G_OPTION_ARG_STRING,
// 		&cExcludeModule,
// 		_("Exclude a given plug-in from activating (it is still loaded though)."), NULL},

// 	{"testing", 'T', G_OPTION_FLAG_IN_MAIN, G_OPTION_ARG_NONE,
// 		&bTesting,
// 		_("For debugging purpose only. The crash manager will not be started to hunt down the bugs."), NULL},
// 	{"easter-eggs", 'E', G_OPTION_FLAG_IN_MAIN, G_OPTION_ARG_NONE,
// 		&g_bEasterEggs,
// 		_("For debugging purpose only. Some hidden and still unstable options will be activated."), NULL},

var (
	// GLDI options: cairo, opengl, indirect-opengl, env, keep-above, no-sticky
	userForceCairo     *bool
	userForceOpenGL    *bool
	userIndirectOpenGL *bool
	userAskBackend     *bool
	userEnv            *string

	userDir         *string
	userThemeServer *string

	userDelay              *int
	userExclude            *string
	userSafeMode           *bool
	userMetacityWorkaround *bool
	userVerbosity          *string
	userForceColor         *bool
	userVersion            *bool
	userLocked             *bool
	userKeepAbove          *bool
	userNoSticky           *bool
	userModulesDir         *string
)

func runDock(cmd *Command, args []string) {
	if *userVersion {
		println(globals.Version()) // -v option only prints version.
		return
	}

	settings := maindock.DockSettings{
		ForceCairo:     *userForceCairo,
		ForceOpenGL:    *userForceOpenGL,
		IndirectOpenGL: *userIndirectOpenGL,
		AskBackend:     *userAskBackend,
		Env:            *userEnv,

		UserDefinedDataDir: *userDir,
		ThemeServer:        *userThemeServer,

		Delay:              *userDelay,
		Exclude:            *userExclude,
		SafeMode:           *userSafeMode,
		MetacityWorkaround: *userMetacityWorkaround,
		Verbosity:          *userVerbosity,
		ForceColor:         *userForceColor,
		Locked:             *userLocked,
		KeepAbove:          *userKeepAbove,
		NoSticky:           *userNoSticky,
		ModulesDir:         *userModulesDir,
	}

	maindock.GuiInstance = &GuiConnector{Source: &datagldi.Data{}}

	// go appletService()

	// HTTP listener for the pprof debug:
	// Access with http://localhost:6060/debug/pprof/
	// go func() { http.ListenAndServe("localhost:6060", nil) }()

	maindock.DockMore(settings)
}

// Start Loader.
//
func appletService() {
	srvdbus.Log = logger

	loader := srvdbus.NewLoader(allapps.List())
	active, e := loader.StartServer()
	if logger.Err(e, "Start Applets service") {
		return
	}

	if active {
		// defer allapps.OnStop()
		loader.StartLoop()
	}
}

// GuiConnector connects the config window to maindock GUI callbacks.
//
type GuiConnector struct {
	Widget *confgui.GuiConfigure
	Win    *gtk.Window
	Source datatype.Source
}

// Create creates the config window.
//
func (gc *GuiConnector) Create() {
	if gc.Widget != nil || gc.Win != nil {
		logger.Info("create GUI, found: widget", gc.Widget != nil, " window", gc.Win != nil)
		return
	}
	gc.Widget, gc.Win = confgui.NewConfigWindow(gc.Source)

	gc.Win.Connect("destroy", func() { // OnQuit is already connected to emit this.
		gc.Widget.Destroy()
		gc.Widget = nil
		gc.Win.Destroy()
		gc.Win = nil
	})

	gc.Widget.Load()
}

// GUI interface

func (gc *GuiConnector) ShowMainGui() {
	gc.Create()
	gc.Widget.Select(confgui.GroupConfig)
}

func (gc *GuiConnector) ShowModuleGui(appletName string) {
	logger.Info("ShowModuleGui", appletName)

	gc.Create()
	gc.Widget.Select(confgui.GroupIcons)
}

func (gc *GuiConnector) ShowGui(icon *gldi.Icon, container *gldi.Container, moduleInstance *gldi.ModuleInstance, showPage int) {
	logger.Info("ShowGui", "icon", icon != nil, "- container", container != nil, "- moduleInstance", moduleInstance != nil, "- page", showPage)

	gc.Create()

	if icon != nil {
		confPath := icon.ConfigPath()
		gc.Widget.SelectIcons(confPath)
	}
	// cairo_dock_items_widget_select_item (ITEMS_WIDGET (pCategory->pCdWidget), pIcon, pContainer, pModuleInstance, iShowPage);
}

func (gc *GuiConnector) ShowAddons() {
	gc.Create()
	gc.Widget.Select(confgui.GroupAdd)
}

func (gc *GuiConnector) ReloadItems() {
	if gc.Widget != nil {
		gc.Widget.ReloadItems()
	}
	logger.Info("ReloadItems")
}

// func (gc *GuiConnector) // ReloadCategoryWidget(){}
func (gc *GuiConnector) Reload() { logger.Info("Reload") }
func (gc *GuiConnector) Close()  { logger.Info("Close") }

func (gc *GuiConnector) UpdateModulesList() { logger.Info("UpdateModulesList") }

func (gc *GuiConnector) UpdateModuleState(name string, active bool) {
	logger.Info("UpdateModuleState "+name, active)
	// cairo_dock_widget_plugins_update_module_state (PLUGINS_WIDGET (pCategory->pCdWidget), cModuleName, bActive);
}

func (gc *GuiConnector) UpdateModuleInstanceContainer(instance *gldi.ModuleInstance, detached bool) {
	logger.Info("UpdateModuleInstanceContainer")
}

func (gc *GuiConnector) UpdateShortkeys() { logger.Info("UpdateShortkeys") }

// func (gc *GuiConnector) UpdateDeskletParams(*gldi.Desklet)                                          {logger.Info("UpdateDeskletParams")}
// func (gc *GuiConnector) UpdateDeskletVisibility(*gldi.Desklet)                                      {logger.Info("UpdateDeskletVisibility")}

// CORE BACKEND
func (gc *GuiConnector) SetStatusMessage(message string) { logger.Info("SetStatusMessage", message) }

func (gc *GuiConnector) ReloadCurrentWidget(moduleInstance *gldi.ModuleInstance, showPage int) {
	logger.Info("ReloadCurrentWidget")
}

func (gc *GuiConnector) ShowModuleInstanceGui(pModuleInstance *gldi.ModuleInstance, iShowPage int) {
	gc.Create()
	gc.Widget.Select(confgui.GroupIcons)
	// show_gui (pModuleInstance->pIcon, NULL, pModuleInstance, iShowPage);
}

// func (gc *GuiConnector) GetWidgetFromName(moduleInstance *gldi.ModuleInstance, group string, key string) {
// 	logger.Info("GetWidgetFromName", group, key)
// }

func (gc *GuiConnector) Window() *gtk.Window { return gc.Win }

/*

static void update_modules_list (void)
{
	if (s_pSimpleConfigWindow == NULL)
		return;
	CDCategory *pCategory = _get_category (CD_CATEGORY_PLUGINS);
	if (pCategory->pCdWidget != NULL)  // category is built
	{
		cairo_dock_widget_reload (pCategory->pCdWidget);
	}
}

static void update_shortkeys (void)
{
	if (s_pSimpleConfigWindow == NULL)
		return;
	CDCategory *pCategory = _get_category (CD_CATEGORY_CONFIG);
	if (pCategory->pCdWidget != NULL)  // category is built
	{
		cairo_dock_widget_config_update_shortkeys (CONFIG_WIDGET (pCategory->pCdWidget));
	}
}

static void update_desklet_params (CairoDesklet *pDesklet)
{
	if (s_pSimpleConfigWindow == NULL || pDesklet == NULL || pDesklet->pIcon == NULL)
		return;

	CDCategory *pCategory = _get_category (CD_CATEGORY_ITEMS);
	if (pCategory->pCdWidget != NULL)  // category is built
	{
		cairo_dock_items_widget_update_desklet_params (ITEMS_WIDGET (pCategory->pCdWidget), pDesklet);
	}
}

static void update_desklet_visibility_params (CairoDesklet *pDesklet)
{
	if (s_pSimpleConfigWindow == NULL || pDesklet == NULL || pDesklet->pIcon == NULL)
		return;

	CDCategory *pCategory = _get_category (CD_CATEGORY_ITEMS);
	if (pCategory->pCdWidget != NULL)  // category is built
	{
		cairo_dock_items_widget_update_desklet_visibility_params (ITEMS_WIDGET (pCategory->pCdWidget), pDesklet);
	}
}

static void update_module_instance_container (GldiModuleInstance *pInstance, gboolean bDetached)
{
	if (s_pSimpleConfigWindow == NULL || pInstance == NULL)
		return;

	CDCategory *pCategory = _get_category (CD_CATEGORY_ITEMS);
	if (pCategory->pCdWidget != NULL)  // category is built
	{
		cairo_dock_items_widget_update_module_instance_container (ITEMS_WIDGET (pCategory->pCdWidget), pInstance, bDetached);
	}
}


static void _reload_category_widget (CDCategoryEnum iCategory)
{
	CDCategory *pCategory = _get_category (iCategory);
	g_return_if_fail (pCategory != NULL);
	if (pCategory->pCdWidget != NULL)  // the category is built, reload it
	{
		GtkWidget *pPrevWidget = pCategory->pCdWidget->pWidget;
		cairo_dock_widget_reload (pCategory->pCdWidget);
		cd_debug ("%s (%p -> %p)", __func__, pPrevWidget, pCategory->pCdWidget->pWidget);

		if (pPrevWidget != pCategory->pCdWidget->pWidget)  // the widget has been rebuilt, let's re-pack it in its container
		{
			GtkWidget *pNoteBook = g_object_get_data (G_OBJECT (s_pSimpleConfigWindow), "notebook");
			GtkWidget *page = gtk_notebook_get_nth_page (GTK_NOTEBOOK (pNoteBook), iCategory);
			gtk_box_pack_start (GTK_BOX (page), pCategory->pCdWidget->pWidget, TRUE, TRUE, 0);
			gtk_widget_show_all (pCategory->pCdWidget->pWidget);
		}
	}
}

static void reload (void)
{
	if (s_pSimpleConfigWindow == NULL)
		return;

	_reload_category_widget (CD_CATEGORY_ITEMS);

	_reload_category_widget (CD_CATEGORY_CONFIG);

	_reload_category_widget (CD_CATEGORY_PLUGINS);
}

////////////////////
/// CORE BACKEND ///
////////////////////

static void set_status_message_on_gui (const gchar *cMessage)
{
	if (s_pSimpleConfigWindow == NULL)
		return;
	GtkWidget *pStatusBar = g_object_get_data (G_OBJECT (s_pSimpleConfigWindow), "status-bar");
	gtk_statusbar_pop (GTK_STATUSBAR (pStatusBar), 0);  // clear any previous message, underflow is allowed.
	gtk_statusbar_push (GTK_STATUSBAR (pStatusBar), 0, cMessage);
}

static void reload_current_widget (GldiModuleInstance *pInstance, int iShowPage)
{
	g_return_if_fail (s_pSimpleConfigWindow != NULL);

	CDCategory *pCategory = _get_category (CD_CATEGORY_ITEMS);
	if (pCategory->pCdWidget != NULL)  // category is built
	{
		cairo_dock_items_widget_reload_current_widget (ITEMS_WIDGET (pCategory->pCdWidget), pInstance, iShowPage);
	}
}


static CairoDockGroupKeyWidget *get_widget_from_name (G_GNUC_UNUSED GldiModuleInstance *pInstance, const gchar *cGroupName, const gchar *cKeyName)
{
	g_return_val_if_fail (s_pSimpleConfigWindow != NULL, NULL);
	cd_debug ("%s (%s, %s)", __func__, cGroupName, cKeyName);
	CDCategory *pCategory = _get_current_category ();
	CDWidget *pCdWidget = pCategory->pCdWidget;
	if (pCdWidget)  /// check that the widget represents the given instance...
		return cairo_dock_gui_find_group_key_widget_in_list (pCdWidget->pWidgetList, cGroupName, cKeyName);
	else
		return NULL;
}



*/
