// Package mgrgldi manages go applets as internal applets.
package mgrgldi

// #cgo pkg-config: gldi
// #include "cairo-dock-applications-manager.h"       // myTaskbarParam
// #include "cairo-dock-container.h"                  // NOTIFICATION_CLICK_ICON ...
// #include "cairo-dock-desklet-manager.h"            // myDeskletObjectMgr
// #include "cairo-dock-dock-manager.h"               // myDocksParam
// #include "cairo-dock-global-variables.h"           // g_pPrimaryContainer
// #include "cairo-dock-keybinder.h"                  // NOTIFICATION_SHORTKEY_CHANGED
// #include "cairo-dock-module-manager.h"             // myModuleObjectMgr
// #include "cairo-dock-module-instance-manager.h"    // NOTIFICATION_MODULE_INSTANCE_DETACHED
// #include "gldi-icon-names.h"                       // GLDI_ICON_NAME_*
// #include "gldi-config.h"                           // GLDI_VERSION
/*

extern void onAppletInit (GldiModuleInstance *pModuleInstance, GKeyFile *pKeyFile);
extern void onAppletStop (GldiModuleInstance *pModuleInstance);
extern gboolean onAppletReload (GldiModuleInstance *pModuleInstance, GldiContainer *pOldContainer, GKeyFile *pKeyFile);


static GldiModule* newModule (gpointer vc) {
	GldiModuleInterface *pInterface = g_new0 (GldiModuleInterface, 1);
	pInterface->initModule = onAppletInit;
	pInterface->stopModule = onAppletStop;
	pInterface->reloadModule = onAppletReload;
	// GldiModule* mod = gldi_module_new ((GldiVisitCard*)(vc), pInterface);
	return gldi_module_new ((GldiVisitCard*)(vc), pInterface);
}


*/
import "C"

import (
	"github.com/sqp/godock/libs/cdglobal" // Global consts.
	"github.com/sqp/godock/libs/cdtype"   // Applets types.

	"github.com/sqp/godock/libs/gldi"
	"github.com/sqp/godock/libs/gldi/appgldi"
	"github.com/sqp/godock/libs/gldi/backendevents"
	"github.com/sqp/godock/libs/gldi/backendmenu"
	"github.com/sqp/godock/libs/gldi/globals"
	"github.com/sqp/godock/libs/packages"
	"github.com/sqp/godock/widgets/gtk/keyfile"

	"path/filepath"
	"time"
	"unsafe"
)

// Apps is the active applet manager.
//
var Apps *AppManager

// LogWindow provides an optional call to open the log window.
// var LogWindow func()

// Register starts the applets manager service to use go internal applets in the dock.
//
func Register(services cdtype.ListStarter, log cdtype.Logger) *AppManager {
	Apps = NewAppManager(services, log)
	backendevents.Register(Apps)

	Apps.registerApplets()

	backendmenu.Register("applets", nil, Apps.BuildMenu)
	return Apps
}

//
//--------------------------------------------------------------[ APPMANAGER ]--

// AppManager is a multi applet manager.
//
type AppManager struct {
	services cdtype.ListStarter                    // Available applets. Key = applet name.
	actives  map[unsafe.Pointer]cdtype.AppInstance // Active services. Key = pointer to dock C Icon.

	visitCards []*gldi.VisitCard // Keep reference to registered modules visit cards. Must not free.

	// menu *backendmenu.DockMenu

	activeWin  *gldi.WindowActor
	activeIcon *gldi.Icon

	stop chan struct{} // Manual exit chan.
	log  cdtype.Logger
}

// NewAppManager creates an applets manager with the given list of applets services.
//
func NewAppManager(services cdtype.ListStarter, log cdtype.Logger) *AppManager {
	load := &AppManager{
		services: services,
		actives:  make(map[unsafe.Pointer]cdtype.AppInstance), //*AppGldi),
		log:      log}

	return load
}

// CountActive returns the number of managed applets.
//
func (o *AppManager) CountActive() int {
	return len(o.actives)
}

// GetApplets return an applet instance.
//
func (o *AppManager) GetApplets(name string) (list []cdtype.AppInstance) {
	for _, app := range o.actives {
		if app.Name() == name {
			list = append(list, app)
		}
	}
	return
}

// Tick ticks all applets pollers.
//
func (o *AppManager) Tick() {
	for _, app := range o.actives {
		app.Poller().Plop() // Safe to use on nil poller.
	}
}

// StartLoop starts the polling loop for applets.
//
func (o *AppManager) StartLoop() {
	// o.Log.Debug("Applets service started")
	// defer o.Log.Debug("Applets service stopped")

	o.stop = make(chan struct{})
	waiter := time.NewTicker(time.Second)
	defer waiter.Stop()

	for { // Start pollers loop and handle pollings until StopLoop is called.

		select {

		case <-waiter.C: // Tick every second to update pollers counters and launch actions.
			for _, ref := range o.actives {
				ref.Poller().Plop() // Safe to use on nil poller.
			}

		case <-o.stop:
			return
		}
	}
}

// StopLoop stops the polling loop.
//
func (o *AppManager) StopLoop() {
	o.stop <- struct{}{}
}

//
//----------------------------------------------------[ APPLETS REGISTRATION ]--

func (o *AppManager) registerApplets() {
	dir := globals.DirShareData(cdglobal.ConfigDirAppletsGo)
	packs, e := packages.ListFromDir(dir, packages.TypeGoInternal, packages.SourceApplet)
	if o.log.Err(e, "registerapplets") {
		return
	}
	for _, pack := range packs {
		if call, ok := o.services[pack.DisplayedName]; ok {
			o.registerOneApplet(pack, call)
		}
	}
}

func (o *AppManager) registerOneApplet(pack *packages.AppletPackage, call cdtype.AppStarter) {
	if gldi.ModuleGet(pack.DisplayedName) != nil {
		o.log.Debug("register applet, already known = dropped", pack.DisplayedName)
		return
	}

	vc := gldi.NewVisitCardFromPackage(pack)
	o.visitCards = append(o.visitCards, vc)
	c := C.newModule(C.gpointer(vc.Ptr))
	mod := gldi.NewModuleFromNative(unsafe.Pointer(c))
	o.log.Debug("register applet", mod != nil, vc.GetName(), vc.GetShareDataDir())
}

//
//------------------------------------------------------[ APPLETS LIFE CYCLE ]--

// startApplet starts a new applet instance connected to its dock icon and instance.
//
func (o *AppManager) startApplet(mi *gldi.ModuleInstance, kf *keyfile.KeyFile) {
	icon := mi.Icon()
	vc := mi.Module().VisitCard()
	name := vc.GetName()

	call, ok := o.services[name]
	if !ok {
		o.log.NewErr(name, "StartApplet: applet unknown (maybe not enabled at compile)")
		return
	}

	// Default desklet renderer.
	if desklet := mi.Desklet(); desklet != nil {
		desklet.SetRendererByName("Simple")
	}

	// Default icon image.
	if icon != nil && icon.GetFileName() == "" {
		icon.SetIcon(mi.Module().VisitCard().GetIconFilePath())

		// 		gtk_widget_queue_draw (pModuleInstance->pContainer->pWidget);
	}

	// Upgrade configuration file if needed.
	if kf != nil && gldi.ConfFileNeedUpdate(kf, vc.GetModuleVersion()) {
		original := filepath.Join(vc.GetShareDataDir(), vc.GetConfFileName())

		o.log.Info("Conf file upgrade", mi.GetConfFilePath(), original)
		// gldi.ConfFileUpgrade(kf, mi.GetConfFilePath(), original, true)

		// 			gchar *cTemplate = g_strdup_printf ("%s/%s", pModuleInstance->pModule->pVisitCard->cShareDataDir, pModuleInstance->pModule->pVisitCard->cConfFileName);
		// 			cairo_dock_upgrade_conf_file (pModuleInstance->cConfFilePath, pKeyFile, cTemplate);
		// 			g_free (cTemplate);
	}

	// Create applet instance and set its core data.
	app := call()

	if app == nil {
		o.log.NewErr(name, "failed to start applet")
		return
	}

	o.actives[unsafe.Pointer(icon.Ptr)] = app

	app.SetBase(name, mi.GetConfFilePath(), globals.DirDockData(), vc.GetShareDataDir()) // TODO: need rootdir
	app.SetBackend(appgldi.New(mi))
	app.SetEvents(app)

	o.log.Debug("Applet started", name)

	// Initialise applet: Load config and apply user settings.
	app.Init(true)

	if o.log.GetDebug() { // If the service debug is active, force it also on applets.
		app.Log().SetDebug(true)
	}
	app.Poller().Restart() // check poller now if it exists. Safe to use on nil poller.
}

// StopApplet close the applet instance.
//
func (o *AppManager) stopApplet(mi *gldi.ModuleInstance) {
	icon := mi.Icon()
	o.sendApp(icon, "on_stop_module")

	if subdock := icon.GetSubDock(); subdock != nil {
		gldi.ObjectUnref(subdock)
		// mi.Icon().pSubDock = nil
	}

	icon.RemoveDataRenderer()

	desklet := mi.Desklet()
	if desklet != nil && desklet.HasIcons() {
		desklet.RemoveIcons()
	}

	o.log.Debug("Applet stopped", mi.Module().VisitCard().GetName())

	delete(o.actives, unsafe.Pointer(icon.Ptr))
}

func (o *AppManager) reloadApplet(mi *gldi.ModuleInstance, oldContainer *gldi.Container, kf *keyfile.KeyFile) bool {
	o.log.Info("reload", mi.Module().VisitCard().GetName())

	icon := mi.Icon()
	o.sendApp(icon, "on_reload_module", kf != nil)

	if o.log.GetDebug() { // If the service debug is active, force it also on applets.
		app := o.actives[unsafe.Pointer(icon.Ptr)]
		if app != nil {
			app.Log().SetDebug(true)
		}
	}

	// Default desklet renderer.
	if desklet := mi.Desklet(); desklet != nil {
		if desklet.HasIcons() {
			desklet.SetRendererByNameData("Caroussel", true, false)
		} else {
			desklet.SetRendererByName("Simple")
		}
	}

	// Default icon image.
	if icon != nil && icon.GetFileName() == "" {
		icon.SetIcon(mi.Module().VisitCard().GetIconFilePath())

		// 		gtk_widget_queue_draw (pModuleInstance->pContainer->pWidget);
	}

	// Update data renderer size.
	if kf == nil {
		// 		CairoDataRenderer *pDataRenderer = cairo_dock_get_icon_data_renderer (pIcon);
		// 		if (pDataRenderer != NULL)
		// 		{
		// 			CairoDataToRenderer *pData = cairo_data_renderer_get_data (pDataRenderer);
		// 			if (pData->iMemorySize > 2)
		// 				cairo_dock_resize_data_renderer_history (pIcon, pIcon->fWidth);
		// 		}
	}
	return true
}

//

/*
	//\____________ internationalize the applets (we need to do that before registering applets).
	gchar *cLocaleDir = g_strdup_printf ("%s/"CD_DBUS_APPLETS_FOLDER"/"LOCALE_DIR_NAME, g_cCairoDockDataDir);  // user version of /usr/share/locale
	if (! g_file_test (cLocaleDir, G_FILE_TEST_EXISTS))  // translations not downloaded yet.
	{
		gchar *cUserAppletsFolder = g_strdup_printf ("%s/"CD_DBUS_APPLETS_FOLDER, g_cCairoDockDataDir);
		if (! g_file_test (cUserAppletsFolder, G_FILE_TEST_EXISTS))
		{
			if (g_mkdir (cUserAppletsFolder, 7*8*8+7*8+5) != 0)  // create an empty folder; since there is no date file, the "locale" package will be seen as "to be updated" by the package manager, which will therefore download it.
				cd_warning ("couldn't create '%s'; third-party applets can't be added", cUserAppletsFolder);
		}
		g_free (cUserAppletsFolder);
*/

/*
 * create an empty folder; since there is no date file, the "locale"
 * package will be seen as "to be updated" by the package manager, and
 * will therefore download it.
 * But if last-modif file is not available, it will be seen as "to be
 * updated" only if the external package is younger than one month:
 *  => cairo-dock-packages.c:_cairo_dock_parse_package_list
 * Solution: added a file with "0" to force the download
 */
/*
		if (g_mkdir (cLocaleDir, 7*8*8+7*8+5) != 0)
			cd_warning ("couldn't create '%s'; applets won't be translated", cLocaleDir);
		else
		{
			gchar *cVersionFile = g_strdup_printf ("%s/last-modif", cLocaleDir);
			g_file_set_contents (cVersionFile,
					"0",
					-1,
					NULL);
			g_free (cVersionFile);
		}
	}
	bindtextdomain (GETTEXT_NAME_EXTRAS, cLocaleDir);  // bind the applets' domain to the user locale folder.
	bind_textdomain_codeset (GETTEXT_NAME_EXTRAS, "UTF-8");
	g_free (cLocaleDir);
*/

//
//------------------------------------------------------------------[ EVENTS ]--

// OnLeftClick forwards a click event to the applet.
//
func (o *AppManager) OnLeftClick(icon *gldi.Icon, container *gldi.Container, btnState uint) bool {
	return o.sendIconOrSub(icon, container, "on_click", "on_click_sub_icon", int(btnState))
}

// OnMiddleClick forwards a click event to the applet.
//
func (o *AppManager) OnMiddleClick(icon *gldi.Icon, container *gldi.Container) bool {
	return o.sendIconOrSub(icon, container, "on_middle_click", "on_middle_click_sub_icon")
}

// OnMouseScroll forwards a mouse event to the applet.
//
func (o *AppManager) OnMouseScroll(icon *gldi.Icon, container *gldi.Container, scrollUp bool) bool {
	return o.sendIconOrSub(icon, container, "on_scroll", "on_scroll_sub_icon", scrollUp)
}

// OnDropData forwards a drop event to the applet.
//
func (o *AppManager) OnDropData(icon *gldi.Icon, container *gldi.Container, data string) bool {
	return o.sendIconOrSub(icon, container, "on_drop_data", "on_drop_data_sub_icon", data)
}

// OnChangeFocus forwards a window focus event to the applet.
//
func (o *AppManager) OnChangeFocus(win *gldi.WindowActor) bool {
	// Emit signal on the applet that had focus.
	if o.activeIcon != nil {
		o.sendApp(o.activeIcon, "on_change_focus", false)
		o.activeIcon = nil
	}

	// Emit signal on the applet that now has focus.
	if win != nil {
		icon := win.GetAppliIcon()
		if icon != nil {
			icon = icon.GetInhibitor(false)
			if icon != nil && icon.IsApplet() {
				o.sendApp(icon, "on_change_focus", true)
				o.activeIcon = icon
			}
		}
	}
	return false
}

// BuildMenu forwards a build menu event to the applet.
//
func (o *AppManager) BuildMenu(m *backendmenu.DockMenu) int {
	o.sendIconOrSub(m.Icon, m.Container, "on_build_menu", "on_build_menu_sub_icon", &MenuerLike{*m})
	return 0 // don't intercept menu. (to check)
}

// sendIconOrSub sends an event to the applet matching the icon or subicon.
//
func (o *AppManager) sendIconOrSub(icon *gldi.Icon, container *gldi.Container, mainEvent, subEvent string, data ...interface{}) bool {
	var appIcon *gldi.Icon
	switch { // Find the base icon of the icon that was clicked on (for subdock or desklets).
	case container.IsDesklet():
		appIcon = container.ToDesklet().GetIcon()

	case gldi.ObjectIsDock(container) && container.ToCairoDock().GetRefCount() != 0 && !icon.IsApplet():
		appIcon = container.ToCairoDock().SearchIconPointingOnDock(nil)

	default:
		appIcon = icon
	}

	if appIcon == nil || icon == nil || icon.Ptr == nil { // TODO: need to check why.
		return false
	}

	if appIcon.Ptr == icon.Ptr {
		return o.sendApp(appIcon, mainEvent, data...) // Main Icon event.
	}
	data = append(data, icon.GetCommand())       // add reference to subicon key.
	return o.sendApp(appIcon, subEvent, data...) // SubIcon event.
}

// sendApp sends an event to the applet matching the icon.
//
func (o *AppManager) sendApp(icon *gldi.Icon, event string, data ...interface{}) bool {
	app := o.actives[unsafe.Pointer(icon.Ptr)]
	if app == nil {
		return false
	}
	o.log.Debug(event, data...)
	app.OnEvent(event, data...)
	return true // an app received the event (even if unused). intercept it.
}

// func (dock *CairoDock) SearchIconPointingOnDock(unknown interface{}) *Icon { // TODO: add param CairoDock **pParentDock
// 	c := C.cairo_dock_search_icon_pointing_on_dock(dock.Ptr, nil)
// 	return NewIconFromNative(unsafe.Pointer(c))
// }

// static inline Icon *_get_main_icon_from_clicked_icon (Icon *pIcon, GldiContainer *pContainer)
// {
// 	Icon *pMainIcon = NULL;

// 	return pMainIcon;
// }

// 	if (pClickedIcon == pAppletIcon)
// 	{
// 		//g_print ("emit clic on main icon\n");
// 		g_signal_emit (pDbusApplet, s_iSignals[CLIC], 0, iButtonState);
// 	}
// 	else if (pDbusApplet->pSubApplet != NULL)
// 	{
// 		//g_print ("emit clic on sub icon\n");
// 		g_signal_emit (pDbusApplet->pSubApplet, s_iSubSignals[CLIC], 0, iButtonState, pClickedIcon->cCommand);
// 	}

// 	// if the applet acts as a launcher, assume it launches the program it controls on click
// 	// Note: if one day it poses a problem, we can either make a new attribute, or add a dbus method (or even reuse the "Animate" method with a pseudo "launching" animation).
// 	if (pAppletIcon->pModuleInstance->pModule->pVisitCard->bActAsLauncher
// 	&& pClickedIcon->pAppli == NULL)  // if the icon already controls a window, don't notify; most probably, the action the applet will take is to show/minimize this window
// 		gldi_class_startup_notify (pClickedIcon);
// 	return GLDI_NOTIFICATION_INTERCEPT;
// }

// void cd_dbus_emit_on_menu_select (GtkMenuItem *pMenuItem, gpointer data)
// {
// 	g_return_if_fail (myData.pCurrentMenuDbusApplet != NULL);
// 	if (GTK_IS_RADIO_MENU_ITEM (pMenuItem))
// 	{
// 		if (!gtk_check_menu_item_get_active (GTK_CHECK_MENU_ITEM (pMenuItem)))
// 			return ;
// 	}

// 	int iNumEntry = GPOINTER_TO_INT (data);
// 	g_signal_emit (myData.pCurrentMenuDbusApplet, s_iSignals[MENU_SELECT], 0, iNumEntry);  // since there can only be 1 menu at once, and the applet knows when the menu is raised, there is no need to pass the icon in the signal: the applet can remember the clicked icon when it received the 'build-menu' event.
// }

//
//-------------------------------------------------------------[ C CALLBACKS ]--

// struct _GldiModuleInterface {
// 	gboolean	(* read_conf_file)		(GldiModuleInstance *pInstance, GKeyFile *pKeyFile);
// 	void		(* reset_config)		(GldiModuleInstance *pInstance);
// 	void		(* reset_data)			(GldiModuleInstance *pInstance);
// 	void		(* load_custom_widget)	(GldiModuleInstance *pInstance, GKeyFile *pKeyFile, GSList *pWidgetList);
// 	void		(* save_custom_widget)	(GldiModuleInstance *pInstance, GKeyFile *pKeyFile, GSList *pWidgetList);
// };

//export onAppletInit
func onAppletInit(cInstance *C.GldiModuleInstance, cKeyfile *C.GKeyFile) {
	mi := gldi.NewModuleInstanceFromNative(unsafe.Pointer(cInstance))
	kf := keyfile.NewFromNative(unsafe.Pointer(cKeyfile))
	Apps.startApplet(mi, kf)
}

//export onAppletStop
func onAppletStop(cInstance *C.GldiModuleInstance) {
	mi := gldi.NewModuleInstanceFromNative(unsafe.Pointer(cInstance))
	Apps.stopApplet(mi)
}

//export onAppletReload
func onAppletReload(cInstance *C.GldiModuleInstance, oldContainer *C.GldiContainer, cKeyfile *C.GKeyFile) C.gboolean {
	mi := gldi.NewModuleInstanceFromNative(unsafe.Pointer(cInstance))
	cont := gldi.NewContainerFromNative(unsafe.Pointer(oldContainer))
	kf := keyfile.NewFromNative(unsafe.Pointer(cKeyfile))
	if Apps.reloadApplet(mi, cont, kf) { // if applet matched, which should always be true.
		return 1
	}
	return 0
}

//
//--------------------------------------------------------------[ MENUERLIKE ]--

// MenuerLike converts the backend menu to match the applets Menuer interface.
//
type MenuerLike struct {
	backendmenu.DockMenu
}

// AddSubMenu adds a submenu to the menu.
//
func (m *MenuerLike) AddSubMenu(label, iconPath string) cdtype.Menuer {
	return &MenuerLike{*m.DockMenu.AddSubMenu(label, iconPath)}
}

// AddEntry adds an item to the menu with its callback.
//
func (m *MenuerLike) AddEntry(label, iconPath string, call interface{}, userData ...interface{}) cdtype.MenuWidgeter {
	return m.DockMenu.AddEntry(label, iconPath, call, userData...)
}

// AddCheckEntry adds a check entry to the menu.
//
func (m *MenuerLike) AddCheckEntry(label string, active bool, call interface{}, userData ...interface{}) cdtype.MenuWidgeter {
	return m.DockMenu.AddCheckEntry(label, active, call, userData)
}

// AddRadioEntry adds a radio entry to the menu.
//
func (m *MenuerLike) AddRadioEntry(label string, active bool, group int, call interface{}, userData ...interface{}) cdtype.MenuWidgeter {
	return m.DockMenu.AddRadioEntry(label, active, group, call, userData)
}
