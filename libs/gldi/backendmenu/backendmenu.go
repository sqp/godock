// Package backendmenu provides menu events and widgets to build the menu.
package backendmenu

// #cgo pkg-config: gldi
/*

#include "cairo-dock-dock-manager.h"       // NOTIFICATION_CLICK_ICON

#include "backendmenu.h"       // to remake.


extern gboolean buildMenuIcon(gpointer, Icon*, GldiContainer*, GtkWidget*);
extern gboolean buildMenuContainer(gpointer, Icon*, GldiContainer*, GtkWidget*, gboolean);


*/
import "C"
import (
	"github.com/conformal/gotk3/cairo"
	"github.com/conformal/gotk3/gdk"
	"github.com/conformal/gotk3/glib"
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/godock/libs/gldi"            // Gldi access.
	"github.com/sqp/godock/libs/gldi/backendgui" // GUI callbacks.
	"github.com/sqp/godock/libs/gldi/globals"    // Global variables.
	"github.com/sqp/godock/libs/ternary"         // Helpers.
	"github.com/sqp/godock/libs/tran"            // Translate.

	"github.com/sqp/godock/widgets/common"

	"fmt"
	"os"
	"path/filepath"
	"unsafe"
)

const (
	LetPass   = C.GLDI_NOTIFICATION_LET_PASS
	Intercept = C.GLDI_NOTIFICATION_INTERCEPT
)

//
//------------------------------------------------------------[ MENU BACKEND ]--

var (
	menuContainer map[string]func(*DockMenu) int
	menuIcon      map[string]func(*DockMenu) int
)

func init() {
	menuContainer = make(map[string]func(*DockMenu) int)
	menuIcon = make(map[string]func(*DockMenu) int)
}

func Register(name string, menucontainer, menuicon func(*DockMenu) int) {
	if menucontainer != nil {
		if len(menuContainer) == 0 {
			C.gldi_object_register_notification(C.gpointer(&C.myContainerObjectMgr),
				C.NOTIFICATION_BUILD_CONTAINER_MENU,
				C.GldiNotificationFunc(C.buildMenuContainer),
				C.GLDI_RUN_FIRST, nil)
		}

		menuContainer[name] = menucontainer
	}

	if menuicon != nil {
		if len(menuIcon) == 0 {
			C.gldi_object_register_notification(C.gpointer(&C.myContainerObjectMgr),
				C.NOTIFICATION_BUILD_ICON_MENU,
				C.GldiNotificationFunc(C.buildMenuIcon),
				C.GLDI_RUN_AFTER, nil)
		}

		menuIcon[name] = menuicon
	}
}

//export buildMenuIcon
func buildMenuIcon(_ C.gpointer, ic *C.Icon, cont *C.GldiContainer, cmenu *C.GtkWidget) C.gboolean {
	for _, call := range menuIcon {
		if call(convert(ic, cont, cmenu)) == Intercept {
			return C.gboolean(Intercept)
		}
	}
	return C.gboolean(LetPass)
}

//export buildMenuContainer
func buildMenuContainer(_ C.gpointer, ic *C.Icon, cont *C.GldiContainer, cmenu *C.GtkWidget, _ C.gboolean) C.gboolean {
	for _, call := range menuContainer {
		if call(convert(ic, cont, cmenu)) == Intercept {
			return C.gboolean(Intercept)
		}
	}
	return C.gboolean(LetPass)
}

func convert(ic *C.Icon, cont *C.GldiContainer, cmenu *C.GtkWidget) *DockMenu {
	icon := gldi.NewIconFromNative(unsafe.Pointer(ic))
	container := gldi.NewContainerFromNative(unsafe.Pointer(cont))

	var dock *gldi.CairoDock
	if gldi.ObjectIsDock(container) {
		dock = container.ToCairoDock()
	}

	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(cmenu))}
	menu := &gtk.Menu{gtk.MenuShell{gtk.Container{gtk.Widget{glib.InitiallyUnowned{obj}}}}}

	return WrapDockMenu(icon, container, dock, menu)

}

//
//-------------------------------------------------------------[ MENU ENTRIES]--

type MenuEntry int

const (
	MenuAbout MenuEntry = iota
	MenuAddApplet
	MenuAddLauncher
	MenuAddMainDock
	MenuAddSeparator
	MenuAddSubDock
	MenuClassItems
	MenuConfigure
	MenuCustomIconRemove
	MenuCustomIconSet
	MenuDeleteDock
	MenuDeskletLock
	MenuDeskletSticky
	MenuDeskletVisibility
	MenuDetachApplet
	MenuEditDock
	MenuDuplicateApplet
	MenuEditApplet
	MenuEditIcon
	MenuHandbook
	MenuHelp
	MenuLaunchNew
	MenuLockIcons
	MenuMakeLauncher
	MenuMoveToDock
	MenuQuickHide
	MenuQuit
	MenuRemoveApplet
	MenuRemoveIcon
	MenuThemes
	MenuWindowAbove
	MenuWindowBelow
	MenuWindowFullScreen
	MenuWindowKill
	MenuWindowMoveAllHere
	MenuWindowMoveHere
	MenuWindowSticky
)

type MenuBtn int

const (
	MenuWindowClose MenuBtn = iota
	MenuWindowCloseAll
	MenuWindowMax
	MenuWindowMin
	MenuWindowMinAll
	MenuWindowShow
	MenuWindowShowAll
)

type DockMenu struct {
	Menu
	Icon      *gldi.Icon
	Container *gldi.Container
	Dock      *gldi.CairoDock // just a pointer to container with type dock.

	btns *ButtonsEntry
}

func WrapDockMenu(icon *gldi.Icon, container *gldi.Container, dock *gldi.CairoDock, menu *gtk.Menu) *DockMenu {
	return &DockMenu{
		Menu:      Menu{*menu},
		Icon:      icon,
		Container: container,
		Dock:      dock}
}

func (m *DockMenu) Entry(entry MenuEntry) {
	switch entry {

	case MenuAbout:
		m.AddEntry(
			"About",
			globals.IconNameAbout,
			func() {
				C._cairo_dock_about((*C.GldiContainer)(unsafe.Pointer(m.Container.Ptr)))
			},
		).SetTooltipText("This will hide the dock until you hover over it with the mouse.")

	case MenuAddApplet:
		m.AddEntry(
			"Applet",
			globals.IconNameAdd,
			backendgui.ShowAddons)

	case MenuAddLauncher:
		m.AddEntry(
			"Custom launcher",
			globals.IconNameAdd,
			func() {
				newIcon := gldi.LauncherAddNew("", m.Dock, nextOrder(m.Icon, m.Dock))
				if newIcon == nil {
					// TODO: use log
					println("Couldn't create create the icon.\nCheck that you have writing permissions on ~/.config/cairo-dock and its sub-folders")
				} else {
					backendgui.ShowItems(newIcon, nil, nil, -1)
				}
			},
		).SetTooltipText("Usually you would drag a launcher from the menu and drop it on the dock.")

	case MenuAddMainDock:
		m.AddEntry(
			"Main dock",
			globals.IconNameAdd,
			func() {
				key := gldi.DockAddConfFile()
				gldi.DockNew(key)

				backendgui.ReloadItems()

				// g_timeout_add_seconds (1, (GSourceFunc)_show_new_dock_msg, cDockName);  // delai, car sa fenetre n'est pas encore bien placee (0,0).
			})

	case MenuAddSeparator:
		m.AddEntry(
			"Separator",
			globals.IconNameAdd,
			func() {
				newIcon := gldi.SeparatorIconAddNew(m.Dock, nextOrder(m.Icon, m.Dock))
				if newIcon == nil {
					// TODO: use log
					println("Couldn't create create the icon.\nCheck that you have writing permissions on ~/.config/cairo-dock and its sub-folders")
				}
			})

	case MenuAddSubDock:
		m.AddEntry(
			"Sub-dock",
			globals.IconNameAdd,
			func() {
				newIcon := gldi.StackIconAddNew(m.Dock, nextOrder(m.Icon, m.Dock))
				if newIcon == nil {
					// TODO: use log
					println("Couldn't create create the icon.\nCheck that you have writing permissions on ~/.config/cairo-dock and its sub-folders")
				}
			})

	case MenuClassItems:
	//\_________________________ class actions.

	// const GList *pClassMenuItems = cairo_dock_get_class_menu_items (icon->cClass);
	// if (pClassMenuItems != NULL)
	// {
	// 	if (bAddSeparator)
	// 	{
	// 		pMenuItem = gtk_separator_menu_item_new ();
	// 		gtk_menu_shell_append (GTK_MENU_SHELL (menu), pMenuItem);
	// 	}
	// 	bAddSeparator = TRUE;
	// 	gchar **pClassItem;
	// 	const GList *m;
	// 	for (m = pClassMenuItems; m != NULL; m = m->next)
	// 	{
	// 		pClassItem = m->data;
	// 		pMenuItem = cairo_dock_add_in_menu_with_stock_and_data (pClassItem[0], pClassItem[2], G_CALLBACK (_cairo_dock_launch_class_action), menu, pClassItem[1]);
	// 	}
	// }

	case MenuConfigure:
		m.AddEntry(
			"Configure",
			globals.IconNamePreferences,
			backendgui.ShowMainGui,
		).SetTooltipText("Configure behaviour, appearance, and applets.")

	case MenuCustomIconRemove:
		classIcon := m.Icon.GetClassInfo(gldi.ClassIcon)
		if classIcon == "" {
			classIcon = m.Icon.GetClass()
		}

		path := filepath.Join(globals.DirCurrentIcons(), classIcon+".png")
		if _, e := os.Stat(path); e != nil {
			path = filepath.Join(globals.DirCurrentIcons(), classIcon+".svg")
			if _, e := os.Stat(path); e != nil {
				path = ""
			}
		}

		// println("custom icon", path)

		if path != "" {
			m.AddEntry(
				"Remove custom icon",
				globals.IconNameRemove,
				func() {
					C._cairo_dock_remove_custom_appli_icon((*C.Icon)(unsafe.Pointer(m.Icon.Ptr)), (*C.CairoDock)(unsafe.Pointer(m.Dock.Ptr)))
				})
		}

	case MenuCustomIconSet:
		m.AddEntry(
			"Set a custom icon",
			globals.IconNameSelectColor,
			func() {
				C._cairo_dock_set_custom_appli_icon((*C.Icon)(unsafe.Pointer(m.Icon.Ptr)), (*C.CairoDock)(unsafe.Pointer(m.Dock.Ptr)))
			})

	case MenuDeleteDock:
		m.AddEntry(
			"Delete this dock",
			globals.IconNameDelete,
			func() {
				if m.Dock.GetRefCount() != 0 {
					gldi.DialogShowWithQuestion("Delete this dock?",
						m.Dock.GetPointedIcon(),
						m.Container,
						globals.DirShareData(globals.CairoDockIcon),
						cbDialogIsOK(func() {
							gldi.ObjectDelete(m.Dock)
						}))
				}
			})

	case MenuDeskletLock:
		m.AddCheckEntry(
			"Lock position",
			m.Container.ToDesklet().PositionLocked(),
			func(c *gtk.CheckMenuItem) {
				m.Container.ToDesklet().LockPosition(c.GetActive())
				// cairo_dock_gui_update_desklet_visibility (pDesklet);
			})

	case MenuDeskletSticky:
		m.AddCheckEntry(
			"On all desktops",
			m.Container.ToDesklet().IsSticky(),
			func(c *gtk.CheckMenuItem) {
				m.Container.ToDesklet().SetSticky(c.GetActive())

				// TODO: check why SetSticky isn't working...
				println("set stick", c.GetActive())
				println("get stick", m.Container.ToDesklet().IsSticky())
				// 	desklet.UpdateVisibility()
				// s_pMainGuiBackend->update_desklet_visibility_params (pDesklet);
				// \--> cairo_dock_update_desklet_visibility_widgets  gui-commons.c

			})

	case MenuDeskletVisibility:
		// accessibility := menu.AddSubMenu(tran.Slate("Visibility")) // GLDI_ICON_NAME_FIND

		// desklet := container.ToDesklet()

		// 	static gpointer data[3];
		// 	data[0] = icon;
		// 	data[1] = pContainer;
		// 	data[2] = menu;
		// 	GtkWidget *pMenuItem;

		// 		GSList *group = NULL;

		// 		CairoDesklet *pDesklet = CAIRO_DESKLET (pContainer);
		// 		CairoDeskletVisibility iVisibility = pDesklet->iVisibility;

		// 		pMenuItem = gtk_radio_menu_item_new_with_label(group, _("Normal"));
		// 		group = gtk_radio_menu_item_get_group(GTK_RADIO_menu_ITEM(pMenuItem));
		// 		gtk_menu_shell_append(GTK_menu_SHELL(pSubMenuAccessibility), pMenuItem);
		// 		gtk_check_menu_item_set_active(GTK_CHECK_menu_ITEM(pMenuItem), iVisibility == CAIRO_DESKLET_NORMAL/*bIsNormal*/);  // on coche celui-ci par defaut, il sera decoche par les suivants eventuellement.
		// 		g_signal_connect(G_OBJECT(pMenuItem), "toggled", G_CALLBACK(_cairo_dock_keep_normal), data);

		// 		pMenuItem = gtk_radio_menu_item_new_with_label(group, _("Always on top"));
		// 		group = gtk_radio_menu_item_get_group(GTK_RADIO_MENU_ITEM(pMenuItem));
		// 		gtk_menu_shell_append(GTK_MENU_SHELL(pSubMenuAccessibility), pMenuItem);
		// 		if (iVisibility == CAIRO_DESKLET_KEEP_ABOVE)
		// 			gtk_check_menu_item_set_active(GTK_CHECK_MENU_ITEM(pMenuItem), TRUE);
		// 		g_signal_connect(G_OBJECT(pMenuItem), "toggled", G_CALLBACK(_cairo_dock_keep_above), data);

		// 		pMenuItem = gtk_radio_menu_item_new_with_label(group, _("Always below"));
		// 		group = gtk_radio_menu_item_get_group(GTK_RADIO_MENU_ITEM(pMenuItem));
		// 		gtk_menu_shell_append(GTK_MENU_SHELL(pSubMenuAccessibility), pMenuItem);
		// 		if (iVisibility == CAIRO_DESKLET_KEEP_BELOW)
		// 			gtk_check_menu_item_set_active(GTK_CHECK_MENU_ITEM(pMenuItem), TRUE);
		// 		g_signal_connect(G_OBJECT(pMenuItem), "toggled", G_CALLBACK(_cairo_dock_keep_below), data);

		// 		if (gldi_desktop_can_set_on_widget_layer ())
		// 		{
		// 			pMenuItem = gtk_radio_menu_item_new_with_label(group, "Widget Layer");
		// 			group = gtk  _radio_menu_item_get_group(GTK_RADIO_MENU_ITEM(pMenuItem));
		// 			gtk_menu_shell_append(GTK_MENU_SHELL(pSubMenuAccessibility), pMenuItem);
		// 			if (iVisibility == CAIRO_DESKLET_ON_WIDGET_LAYER)
		// 				gtk_check_menu_item_set_active(GTK_CHECK_MENU_ITEM(pMenuItem), TRUE);
		// 			g_signal_connect(G_OBJECT(pMenuItem), "toggled", G_CALLBACK(_cairo_dock_keep_on_widget_layer), data);
		// 		}

		// 		pMenuItem = gtk_radio_menu_item_new_with_label(group, _("Reserve space"));
		// 		group = gtk_radio_menu_item_get_group(GTK_RADIO_MENU_ITEM(pMenuItem));
		// 		gtk_menu_shell_append(GTK_MENU_SHELL(pSubMenuAccessibility), pMenuItem);
		// 		if (iVisibility == CAIRO_DESKLET_RESERVE_SPACE)
		// 			gtk_check_menu_item_set_active(GTK_CHECK_MENU_ITEM(pMenuItem), TRUE);
		// 		g_signal_connect(G_OBJECT(pMenuItem), "toggled", G_CALLBACK(_cairo_dock_keep_space), data);

	case MenuDetachApplet:
		m.AddEntry(
			ternary.String(m.Dock != nil, "Detach", "Return to the dock"),
			ternary.String(m.Dock != nil, globals.IconNameGotoTop, globals.IconNameGotoBottom),
			func() {
				// if (CAIRO_DOCK_IS_DESKLET (pContainer))
				// 	icon = (CAIRO_DESKLET (pContainer))->pIcon;  // l'icone cliquee du desklet n'est pas forcement celle qui contient le module !
				// g_return_if_fail (CAIRO_DOCK_IS_APPLET (icon));

				m.Icon.ModuleInstance().Detach()
			})

	case MenuDuplicateApplet:
		m.AddEntry(
			"Duplicate",
			globals.IconNameAdd,
			func() {
				// if (CAIRO_DOCK_IS_DESKLET (pContainer))
				// 	icon = (CAIRO_DESKLET (pContainer))->pIcon;  // l'icone cliquee du desklet n'est pas forcement celle qui contient le module !
				// g_return_if_fail (CAIRO_DOCK_IS_APPLET (icon));

				m.Icon.ModuleInstance().Module().AddInstance()
				// gldi_module_add_instance (icon->pModuleInstance->pModule);
			})

	case MenuEditApplet:
		m.AddEntry(
			"Edit",
			globals.IconNameEdit,
			func() {
				// if (CAIRO_DOCK_IS_DESKLET (pContainer))
				// 	icon = (CAIRO_DESKLET (pContainer))->pIcon;  // l'icone cliquee du desklet n'est pas forcement celle qui contient le module.
				// g_return_if_fail (CAIRO_DOCK_IS_APPLET (icon));
				backendgui.ShowItems(m.Icon, nil, nil, -1)
			})

	case MenuEditDock:
		m.AddEntry(
			"Configure this dock",
			globals.IconNameExecute,
			func() {
				if m.Dock.GetRefCount() != 0 {
					backendgui.ShowItems(nil, m.Container, nil, 0)
				}
			},
		).SetTooltipText("Customize the position, visibility and appearance of this main dock.")

	case MenuEditIcon:
		m.AddEntry(tran.Slate(
			"Edit"),
			globals.IconNameEdit,
			func() {
				switch m.Icon.GetDesktopFileName() {
				case "", "none":
					gldi.DialogShowTemporaryWithIcon("Sorry, this icon doesn't have a configuration file.", m.Icon, m.Container, 4000, "same icon")

				default:
					backendgui.ShowItems(m.Icon, nil, nil, -1)
				}
			})

	case MenuHandbook:
		m.AddEntry(
			tran.Slate("Applet's handbook"),
			globals.IconNameAbout,
			func() { gldi.PopupAboutApplet(m.Icon.ModuleInstance()) })

		// handbook + 5x to facto
		// picon := icon
		// if container.IsDesklet() {
		// 	picon = container.Icon()
		// 	icon = (CAIRO_DESKLET (pContainer))->pIcon;  // l'icone cliquee du desklet n'est pas forcement celle qui contient le module !
		// }

	case MenuHelp:
		m.AddEntry(
			"Help",
			globals.IconNameHelp,
			func() { backendgui.ShowModuleGui("Help") },
		).SetTooltipText("There are no problems, only solutions (and a lot of useful hints!)")

	case MenuLaunchNew:
		m.AddEntry(
			"Launch a new (Shift+clic)",
			globals.IconNameAdd,
			func() {
				gldi.ObjectNotify(m.Container, C.NOTIFICATION_CLICK_ICON, m.Icon, m.Dock, gdk.GDK_SHIFT_MASK) // emit a shift click on the icon.
			})

	case MenuLockIcons:
		m.AddCheckEntry(
			"Lock icons position",
			globals.DocksParam.IsLockIcons(),
			func() {
				globals.DocksParam.SetLockIcons(!globals.DocksParam.IsLockIcons())
				// cairo_dock_update_conf_file (g_cConfFile,
				// 	G_TYPE_BOOLEAN, "Accessibility", "lock icons", myDocksParam.bLockIcons,
				// 	G_TYPE_INVALID);
			},
		).SetTooltipText("This will (un)lock the position of the icons.")

	case MenuMakeLauncher:
		m.AddEntry(
			"Make it a launcher",
			globals.IconNameNew,
			func() {
				C._cairo_dock_make_launcher_from_appli((*C.Icon)(unsafe.Pointer(m.Icon.Ptr)), (*C.CairoDock)(unsafe.Pointer(m.Dock.Ptr)))
			})

	case MenuMoveToDock:
		sub, _ := m.SubMenu(tran.Slate("Move to another dock"), globals.IconNameJumpTo)

		docks := gldi.GetAllAvailableDocks(m.Icon.GetContainer().ToCairoDock(), m.Icon.GetSubDock())
		docks = append(docks, nil)
		for _, dock := range docks {
			name := ""
			key := ""
			icon := ""
			if dock == nil {
				name = tran.Slate("New main dock")
				icon = globals.IconNameNew
			} else {
				name = ternary.String(dock.GetReadableName() != "", dock.GetReadableName(), dock.GetDockName())
				key = dock.GetDockName()
			}
			sub.AddEntry(
				name,
				icon, // TODO: maybe add something at least for subdocks.
				func() {
					if key == "" {
						key = gldi.DockAddConfFile()
					}

					m.Icon.WriteContainerNameInConfFile(key) // Update icon conf file.

					if m.Icon.IsLauncher() || m.Icon.IsStackIcon() || m.Icon.IsSeparator() { // Reload icon (creating the dock).
						gldi.ObjectReload(m.Icon)

					} else if m.Icon.IsApplet() {
						gldi.ObjectReload(m.Icon.ModuleInstance())
					}

					newdock := gldi.DockGet(key)
					if newdock != nil && newdock.GetRefCount() == 0 && len(newdock.Icons()) == 1 {
						str := tran.Slate("The new dock has been created.\nYou can customize it by right-clicking on it -> cairo-dock -> configure this dock.")
						gldi.DialogShowGeneralMessage(str, 8000) // we don't show it on the new dock as its window isn't positioned yet (0,0).
					}
				})
		}

	case MenuQuickHide:
		m.AddEntry(
			"Quick-Hide",
			globals.IconNameGotoBottom,
			gldi.QuickHideAllDocks,
		).SetTooltipText("This will hide the dock until you hover over it with the mouse.")

	case MenuQuit:
		item := m.AddEntry(
			"Quit",
			globals.IconNameQuit,
			func() {
				gtk.MainQuit() // TODO: remove SQP HACK, easy quit no confirm for tests.

				gldi.DialogShowWithQuestion("Quit Cairo-Dock?",
					GetIconForDesklet(m.Icon, m.Container),
					m.Container,
					globals.DirShareData(globals.CairoDockIcon),
					cbDialogIsOK(gtk.MainQuit))
			})

		// const gchar *cDesktopSession = g_getenv ("DESKTOP_SESSION");
		// gboolean bIsCairoDockSession = cDesktopSession && g_str_has_prefix (cDesktopSession, "cairo-dock");

		// 	// if we're using a Cairo-Dock session and we quit the dock we have... nothing to relaunch it!
		// 	if bIsCairoDockSession {

		// item.SetSensitive(false)
		// item.SetTooltipText("You're using a Cairo-Dock Session!\nIt's not advised to quit the dock but you can press Shift to unlock this menu entry.")

		// static void _cairo_dock_set_sensitive_quit_menu (G_GNUC_UNUSED GtkWidget *pMenuItem, GdkEventKey *pKey, GtkWidget *pQuitEntry)
		// {
		// pMenuItem not used because we want to only modify one entry
		// 	if (pKey->type == GDK_KEY_PRESS &&
		// 		(pKey->keyval == GDK_KEY_Shift_L ||
		// 		pKey->keyval == GDK_KEY_Shift_R)) // pressed
		// 		gtk_widget_set_sensitive (pQuitEntry, TRUE); // unlocked
		// 	else if (pKey->state & GDK_SHIFT_MASK) // released
		// 		gtk_widget_set_sensitive (pQuitEntry, FALSE); // locked)
		// }
		cbSensitive := func(_ *gtk.CheckMenuItem, event *gdk.Event) {
			key := &gdk.EventKey{event}

			println("key", key.KeyVal())

			if key.KeyVal() == uint(C.GDK_KEY_Shift_R) { // pressed.
				item.SetSensitive(true)

			} else if gdk.ModifierType(key.State())&gdk.GDK_SHIFT_MASK > 0 { // released.
				item.SetSensitive(false)
			}
		}

		_, e := item.Connect("key-press-event", cbSensitive)
		if e != nil {
			println((e.Error()))
		}
		item.Connect("key-release-event", cbSensitive)

		// 		gtk_widget_set_sensitive (pMenuItem, FALSE); // locked
		// 		gtk_widget_set_tooltip_text (pMenuItem, _("You're using a Cairo-Dock Session!\nIt's not advised to quit the dock but you can press Shift to unlock this menu entry."));
		// 		// signal to unlock the entry (signal monitored only in the submenu)
		// 		g_signal_connect (pSubMenu, "key-press-event", G_CALLBACK (_cairo_dock_set_sensitive_quit_menu), pMenuItem);
		// 		g_signal_connect (pSubMenu, "key-release-event", G_CALLBACK (_cairo_dock_set_sensitive_quit_menu), pMenuItem);
		// 	}

		// case MenuAutostart:

		// 	gchar *cCairoAutoStartDirPath = g_strdup_printf ("%s/.config/autostart", g_getenv ("HOME"));
		// 	gchar *cCairoAutoStartEntryPath = g_strdup_printf ("%s/cairo-dock.desktop", cCairoAutoStartDirPath);
		// 	gchar *cCairoAutoStartEntryPath2 = g_strdup_printf ("%s/cairo-dock-cairo.desktop", cCairoAutoStartDirPath);
		// 	if (! bIsCairoDockSession && ! g_file_test (cCairoAutoStartEntryPath, G_FILE_TEST_EXISTS) && ! g_file_test (cCairoAutoStartEntryPath2, G_FILE_TEST_EXISTS))
		// 	{
		// 		cairo_dock_add_in_menu_with_stock_and_data (_("Launch Cairo-Dock on startup"),
		// 			GLDI_ICON_NAME_ADD,
		// 			G_CALLBACK (_cairo_dock_add_autostart),
		// 			pSubMenu,
		// 			NULL);
		// 	}
		// 	g_free (cCairoAutoStartEntryPath);
		// 	g_free (cCairoAutoStartEntryPath2);
		// 	g_free (cCairoAutoStartDirPath);

		// case MenuThirdParty:

		// 	pMenuItem = cairo_dock_add_in_menu_with_stock_and_data (_("Get more applets!"),
		// 		GLDI_ICON_NAME_ADD,
		// 		G_CALLBACK (_cairo_dock_show_third_party_applets),
		// 		pSubMenu,
		// 		NULL);
		// 	gtk_widget_set_tooltip_text (pMenuItem, _("Third-party applets provide integration with many programs, like Pidgin"));

	case MenuRemoveApplet:
		m.AddEntry(
			tran.Slate("Remove"),
			globals.IconNameRemove,
			func() {
				// if (CAIRO_DOCK_IS_DESKLET (pContainer))
				// 	icon = (CAIRO_DESKLET (pContainer))->pIcon;  // l'icone cliquee du desklet n'est pas forcement celle qui contient le module !
				// g_return_if_fail (CAIRO_DOCK_IS_APPLET (icon));

				gldi.DialogShowWithQuestion(
					fmt.Sprintf("You're about to remove this applet (%s) from the dock. Are you sure?", m.Icon.ModuleInstance().Module().VisitCard().GetTitle()),
					m.Icon,
					m.Container,
					"same icon",
					cbDialogIsOK(func() {
						gldi.ObjectDelete(m.Icon.ModuleInstance())
					}))
			})

	case MenuRemoveIcon:
		m.AddEntry(
			tran.Slate("Remove"),
			globals.IconNameRemove,
			func() {
				name := ternary.String(m.Icon.GetInitialName() != "", m.Icon.GetInitialName(), m.Icon.GetName())
				if name == "" {
					name = ternary.String(m.Icon.IsSeparator(), tran.Slate("separator"), "no name")
				}
				gldi.DialogShowWithQuestion(
					fmt.Sprintf("You're about to remove this icon (%s) from the dock. Are you sure?", name),
					m.Icon,
					m.Container,
					"same icon",
					cbDialogIsOK(func() {
						if m.Icon.IsStackIcon() && m.Icon.GetSubDock() != nil && len(m.Icon.GetSubDock().Icons()) > 0 {
							gldi.DialogShowWithQuestion(
								"Do you want to re-dispatch the icons contained inside this container into the dock?\n(otherwise they will be destroyed)",
								m.Icon,
								m.Container,
								globals.DirShareData(globals.CairoDockIcon),
								cbDialogIsOK(func() {
									m.Icon.RemoveIconsFromSubdock(m.Dock)
								}))
						}

						m.Icon.RemoveFromDock()
					}))
			}).SetTooltipText("You can remove a launcher by dragging it out of the dock with the mouse .")

	case MenuThemes:
		// 		pMenuItem = cairo_dock_add_in_menu_with_stock_and_data (_("Manage themes"),
		// 			CAIRO_DOCK_SHARE_DATA_DIR"/icons/icon-appearance.svg",
		// 			G_CALLBACK (_cairo_dock_initiate_theme_management),
		// 			pSubMenu,
		// 			NULL);
		// 		gtk_widget_set_tooltip_text (pMenuItem, _("Choose from amongst many themes on the server or save your current theme."));

	case MenuWindowAbove:
		flag := m.Icon.Window().IsAbove()
		m.AddEntry(
			ternary.String(flag, "Don't keep above", "Keep above"),
			ternary.String(flag, globals.IconNameGotoBottom, globals.IconNameGotoTop),
			cbActionWindowToggle(m.Icon, (*gldi.WindowActor).SetAbove, (*gldi.WindowActor).IsAbove))

	case MenuWindowBelow:
		if !m.Icon.Window().IsHidden() { // this could be a button in the menu, if we find an icon that doesn't look too much like the "minimise" one
			m.AddEntry(
				"Below other windows"+actionMiddleClick(m.Icon, 4),
				globals.DirShareData("icons", "icon-lower.svg"),
				cbActionWindow(m.Icon, (*gldi.WindowActor).Lower))
		}

	case MenuWindowFullScreen:
		flag := m.Icon.Window().IsFullScreen()
		m.AddEntry(
			ternary.String(flag, "Not Fullscreen", "Fullscreen"),
			ternary.String(flag, globals.IconNameLeaveFullScreen, globals.IconNameFullScreen),
			cbActionWindowToggle(m.Icon, (*gldi.WindowActor).SetFullScreen, (*gldi.WindowActor).IsFullScreen))

	case MenuWindowKill:
		m.AddEntry(
			"Kill",
			globals.IconNameClose,
			cbActionWindow(m.Icon, (*gldi.WindowActor).Kill))

	case MenuWindowMoveAllHere:
		m.AddEntry(
			"Move all to this desktop",
			globals.IconNameJumpTo,
			cbActionSubWindows(m.Icon, (*gldi.WindowActor).MoveToCurrentDesktop))

	case MenuWindowMoveHere:
		var callback func()
		if !m.Icon.Window().IsOnCurrentDesktop() {
			callback = cbActionWindow(m.Icon, func(win *gldi.WindowActor) {
				win.MoveToCurrentDesktop()
				if !win.IsHidden() {
					win.Show()
				}
			})
		}
		m.AddEntry("Move to this desktop", globals.IconNameJumpTo, callback)

	case MenuWindowSticky:
		m.AddEntry(
			ternary.String(m.Icon.Window().IsSticky(), "Visible only on this desktop", "Visible on all desktops"),
			globals.IconNameJumpTo,
			cbActionWindowToggle(m.Icon, (*gldi.WindowActor).SetSticky, (*gldi.WindowActor).IsSticky))

	}
}

func (m *DockMenu) Button(btn MenuBtn) {
	switch btn {
	case MenuWindowClose:
		m.btns.AddButton(
			"Close"+actionMiddleClick(m.Icon, 1),
			globals.DirShareData("icons", "icon-close.svg"),
			cbActionWindow(m.Icon, (*gldi.WindowActor).Close))

	case MenuWindowCloseAll:
		m.btns.AddButton(
			"Close all"+actionMiddleClick(m.Icon, 1),
			globals.DirShareData("icons", "icon-close.svg"),
			cbActionSubWindows(m.Icon, (*gldi.WindowActor).Close))

	case MenuWindowMax:
		max := m.Icon.Window().IsMaximized()
		m.btns.AddButton(
			ternary.String(max, "Unmaximise", "Maximise"),
			globals.DirShareData("icons", ternary.String(max, "icon-restore.svg", "icon-maximize.svg")),
			cbActionWindowToggle(m.Icon, (*gldi.WindowActor).Maximize, (*gldi.WindowActor).IsMaximized))

	case MenuWindowMin:
		m.btns.AddButton(
			"Minimise"+actionMiddleClick(m.Icon, 2),
			globals.DirShareData("icons", "icon-minimize.svg"),
			cbActionWindow(m.Icon, (*gldi.WindowActor).Minimize))

	case MenuWindowMinAll:
		m.btns.AddButton(
			"Minimise all"+actionMiddleClick(m.Icon, 2),
			globals.DirShareData("icons", "icon-minimize.svg"),
			cbActionSubWindows(m.Icon, (*gldi.WindowActor).Minimize))

	case MenuWindowShow:
		m.btns.AddButton(
			"Show",
			globals.IconNameFind,
			cbActionWindow(m.Icon, (*gldi.WindowActor).Show))

	case MenuWindowShowAll:
		m.btns.AddButton(
			"Show all",
			globals.IconNameFind,
			cbActionSubWindows(m.Icon, (*gldi.WindowActor).Show))

	}
}

//
//-----------------------------------------------------------[ DOCKMENU FILL ]--

func (menu *DockMenu) SubMenu(label, iconPath string) (*DockMenu, *gldi.MenuItem) {
	submenu, item := gldi.MenuAddSubMenu(&menu.Menu.Menu, label, iconPath)
	return WrapDockMenu(menu.Icon, menu.Container, menu.Dock, submenu), item
}

func (m *DockMenu) AddButtonsEntry(str string) *ButtonsEntry {
	item, _ := gtk.MenuItemNew()
	m.Append(item)

	item.Connect("button-press-event", onMenuItemPress)         // Forward click to inside buttons.
	item.Connect("motion-notify-event", onMenuItemMotionNotify) // Highlight pointed button.
	item.Connect("leave-notify-event", onMenuItemLeaveNotify)   // Highlight pointed button.
	item.Connect("enter-notify-event", onMenuItemEnterNotify)   // Highlight pointed button.
	item.Connect("draw", onMenuItemDraw)                        // Highlight pointed button.

	// 	g_signal_connect (G_OBJECT (pMenuItem), "motion-notify-event",
	// 		G_CALLBACK(_on_motion_notify_menu_item),
	// 		NULL);  // we need to manually higlight the currently pointed button
	// 	g_signal_connect (G_OBJECT (pMenuItem),
	// 		"leave-notify-event",
	// 		G_CALLBACK (_on_leave_menu_item),
	// 		NULL);  // to turn off the highlighted button when we leave the menu-item (if we leave it quickly, a motion event won't be generated)
	// 	g_signal_connect (G_OBJECT (pMenuItem),
	// 		"enter-notify-event",
	// 		G_CALLBACK (_on_enter_menu_item),
	// 		NULL);  // to force the label to not highlight (it gets highlighted, even if we overwrite the motion_notify_event callback)
	// 	g_signal_connect (G_OBJECT (pMenuItem),
	// 		"draw",
	// 		G_CALLBACK (_draw_menu_item),
	// 		NULL);  // we don't want the whole menu-item to be higlighted, but only the currently pointed button; so we draw the menu-item ourselves.

	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 1)
	item.Add(box)

	label, _ := gtk.LabelNew(str)
	box.PackStart(label, false, false, 0)

	m.btns = &ButtonsEntry{*box, nil}

	return &ButtonsEntry{*box, nil}
}

// onPressMenuItem forwards the click on the menuitem to the buttons inside if needed.
//
func onMenuItemPress(m *gtk.MenuItem, ev *gdk.Event) bool {
	return gobool(C._on_press_menu_item((*C.GtkWidget)(unsafe.Pointer(m.Native())), (*C.GdkEventButton)(unsafe.Pointer(ev.Native())), nil))
}

// onMenuItemMotionNotify
//
func onMenuItemMotionNotify(m *gtk.MenuItem, ev *gdk.Event) bool {
	return gobool(C._on_motion_notify_menu_item((*C.GtkWidget)(unsafe.Pointer(m.Native())), (*C.GdkEventMotion)(unsafe.Pointer(ev.Native())), nil))
}

// onMenuItemLeaveNotify
//
func onMenuItemLeaveNotify(m *gtk.MenuItem, _ *gdk.Event) bool {
	return gobool(C._on_leave_menu_item((*C.GtkWidget)(unsafe.Pointer(m.Native())), nil, nil))
}

// onMenuItemEnterNotify
//
func onMenuItemEnterNotify(m *gtk.MenuItem, _ *gdk.Event) bool {
	return gobool(C._on_enter_menu_item((*C.GtkWidget)(unsafe.Pointer(m.Native())), nil, nil))
}

// onMenuItemDraw
//
func onMenuItemDraw(m *gtk.MenuItem, cr *cairo.Context) bool {
	return gobool(C._draw_menu_item((*C.GtkWidget)(unsafe.Pointer(m.Native())), (*C.cairo_t)(unsafe.Pointer(cr.Native()))))
}

func gobool(b C.gboolean) bool {
	if b == 1 {
		return true
	}
	return false
}

//
//--------------------------------------------------------[ DOCKMENU HELPERS ]--

// GetIconForDesklet will return the correct icon if clicked on a desklet.
//
func GetIconForDesklet(icon *gldi.Icon, container *gldi.Container) *gldi.Icon {
	if container.IsDesklet() {
		return container.ToDesklet().GetIcon()
	}
	return icon
}

// actionMiddleClick returns the middle-click string to add to the action button
// if it's matching the action ID provided.
//
func actionMiddleClick(icon *gldi.Icon, id int) string {
	if globals.TaskbarParam.ActionOnMiddleClick() != id || icon.IsApplet() {
		return ""
	}
	return " (" + tran.Slate("middle-click") + ")"
}

// nextOrder calculates an order position for a new icon, based on the mouse
// position on current icon, and the order of the current icon and the next one.
//
func nextOrder(icon *gldi.Icon, dock *gldi.CairoDock) float64 {
	if icon == nil {
		return gldi.IconLastOrder
	}

	if float64(dock.Container().MouseX()) < icon.DrawX()+icon.Width()*icon.Scale()/2 { // on the left.
		prev := dock.GetPreviousIcon(icon)
		if prev == nil {
			return icon.Order() - 1
		}
		return (icon.Order() + prev.Order()) / 2
	}

	next := dock.GetNextIcon(icon) // on the right.
	if next == nil {
		return icon.Order() + 1
	}
	return (icon.Order() + next.Order()) / 2
}

//
//-------------------------------------------------------[ PREPARE CALLBACKS ]--

// Returns a func to use as gtk callback. On event, it will test the icon still
// has a valid window and launch the provided action on this window.
//
func cbActionWindow(icon *gldi.Icon, call func(*gldi.WindowActor)) func() {
	return func() {
		if icon.IsAppli() {
			call(icon.Window())
		}
	}
}

func cbActionWindowToggle(icon *gldi.Icon, call func(*gldi.WindowActor, bool), getvalue func(*gldi.WindowActor) bool) func() {
	return cbActionWindow(icon, func(win *gldi.WindowActor) {
		v := getvalue(win)
		call(win, !v)
	})
}

// same as cbActionWindow but launch the action on all subdock windows.
//
func cbActionSubWindows(icon *gldi.Icon, call func(*gldi.WindowActor)) func() {
	return func() {
		for _, ic := range icon.SubDockIcons() {
			if ic.IsAppli() {
				call(ic.Window())
			}
		}
	}
}

// Prepare a callback for DialogWithQuestion to launch on user confirmation.
//
func cbDialogIsOK(call func()) func(int, *gtk.Widget) {
	return func(clickedButton int, widget *gtk.Widget) {
		if clickedButton == 0 || clickedButton == -1 {
			call()
		}
	}
}

//
//------------------------------------------------------------[ MENU BUILDER ]--

// Menu : a GtkMenu builder.
//
type Menu struct {
	gtk.Menu
}

func NewMenu() *Menu {
	gtkmenu, _ := gtk.MenuNew()
	menu := &Menu{*gtkmenu}
	return menu
}

func (menu *Menu) Separator() {
	sep, _ := gtk.SeparatorMenuItemNew()
	menu.Append(sep)
}

func (menu *Menu) AddEntry(label, iconPath string, call interface{}, userData ...interface{}) *gtk.MenuItem {
	item := gldi.MenuAddItem(&menu.Menu, label, iconPath)
	if call == nil {
		item.SetSensitive(false)
	} else {
		item.Connect("activate", call, userData...)
	}
	return item
}

func (menu *Menu) AddCheckEntry(label string, active bool, call interface{}) (item *gtk.CheckMenuItem) {
	item, _ = gtk.CheckMenuItemNewWithLabel(label)
	item.SetActive(active)
	if call != nil {
		item.Connect("toggled", call)
	}
	menu.Append(item)
	return item
}

// func (menu *Menu) AddSubMenu(title string) *Menu {
// 	gtkmenu := NewMenu()
// 	gtkItem, _ := gtk.MenuItemNewWithLabel(title)
// 	gtkItem.SetSubmenu(gtkmenu)
// 	menu.Append(gtkItem)
// 	return gtkmenu
// }

type ButtonsEntry struct {
	gtk.Box
	list []*gtk.Button
}

func (box *ButtonsEntry) AddButton(tooltip, img string, call interface{}) *gtk.Button {
	btn, _ := gtk.ButtonNew()
	btn.SetTooltipText(tooltip)
	btn.Connect("clicked", call)
	box.PackEnd(btn, false, false, 0)

	if img != "" {

		// 		if (*gtkStock == '/')
		// 			int size = cairo_dock_search_icon_size (GTK_ICON_SIZE_MENU);

		image, e := common.ImageNewFromFile(img, 12) // TODO: icon size
		if e == nil {
			btn.SetImage(image)
		}
	}

	box.list = append(box.list, btn)
	return btn
}

func (box *ButtonsEntry) onMenuItemMotionNotify(unused *gtk.Box, event *gdk.Event) {

	// GdkEventMotion

	// for _, btn := range box.list {
	// }

	// GtkWidget *hbox = gtk_bin_get_child (GTK_BIN (pWidget));
	// GList *children = gtk_container_get_children (GTK_CONTAINER (hbox));
	// int x = pEvent->x, y = pEvent->y;  // position of the mouse relatively to the menu-item
	// int xb, yb;  // position of the top-left corner of the button relatively to the menu-item
	// GtkWidget* pButton;
	// GList* c;
	// for (c = children->next; c != NULL; c = c->next)  // skip the label
	// {
	// 	pButton = GTK_WIDGET (c->data);
	// 	GtkAllocation alloc;
	// 	gtk_widget_get_allocation (pButton, &alloc);
	// 	gtk_widget_translate_coordinates (pButton, pWidget,
	// 		0, 0, &xb, &yb);
	// 	if (x >= xb && x < (xb + alloc.width)
	// 	&& y >= yb && y < (yb + alloc.height))  // the mouse is inside the button -> select it
	// 	{
	// 		gtk_widget_set_state_flags (pButton, GTK_STATE_FLAG_PRELIGHT, TRUE);
	// 		gtk_widget_set_state_flags (
	// 			gtk_bin_get_child(GTK_BIN(pButton)),
	// 			GTK_STATE_FLAG_PRELIGHT, TRUE);
	// 	}
	// 	else  // else deselect it, in case it was selected
	// 	{
	// 		gtk_widget_set_state_flags (pButton, GTK_STATE_FLAG_NORMAL, TRUE);
	// 		gtk_widget_set_state_flags (
	// 			gtk_bin_get_child(GTK_BIN(pButton)),
	// 			GTK_STATE_FLAG_NORMAL, TRUE);
	// 	}
	// }
	// GtkWidget *pLabel = children->data;  // force the label to be in a normal state
	// gtk_widget_set_state_flags (pLabel, GTK_STATE_FLAG_NORMAL, TRUE);
	// g_list_free (children);
	// gtk_widget_queue_draw (pWidget);  // and redraw everything
	// return FALSE;
}
