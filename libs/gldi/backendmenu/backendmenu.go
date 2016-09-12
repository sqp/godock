// Package backendmenu provides menu events and widgets to build the menu.
package backendmenu

// #cgo pkg-config: gldi
/*
#include "backendmenu.h"       // to remake/remove.

*/
import "C"
import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"github.com/sqp/godock/libs/cdglobal"        // Dock types.
	"github.com/sqp/godock/libs/cdtype"          // Applet types.
	"github.com/sqp/godock/libs/config"          // Config parser.
	"github.com/sqp/godock/libs/gldi"            // Gldi access.
	"github.com/sqp/godock/libs/gldi/backendgui" // GUI callbacks.
	"github.com/sqp/godock/libs/gldi/current"    // Current theme settings.
	"github.com/sqp/godock/libs/gldi/desktops"   // Desktop and screens info.
	"github.com/sqp/godock/libs/gldi/dialog"     // Popup dialog.
	"github.com/sqp/godock/libs/gldi/globals"    // Global variables.
	"github.com/sqp/godock/libs/gldi/notif"      // Dock notifs.
	"github.com/sqp/godock/libs/ternary"         // Helpers.
	"github.com/sqp/godock/libs/text/tran"       // Translate.

	"github.com/sqp/godock/widgets/about"
	"github.com/sqp/godock/widgets/gtk/menus"

	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"
)

var log cdtype.Logger

// Register registers menu callbacks, to receive menu events.
//
func Register(l cdtype.Logger, menucontainer, menuicon func(*DockMenu) bool) {
	log = l

	notif.RegisterContainerMenuContainer(WrapDockMenuCallback(menucontainer))
	notif.RegisterContainerMenuIcon(WrapDockMenuCallback(menuicon))
}

//
//-------------------------------------------------------------[ MENU ENTRIES]--

// MenuEntry represents common menu item entries type.
//
type MenuEntry int

// Common menu item entries.
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
	MenuEditTarget
	MenuHandbook
	MenuHelp
	MenuLaunchNew
	MenuLockIcons
	MenuMakeLauncher
	MenuMoveToDesktopClass
	MenuMoveToDesktopWindow
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

// MenuBtn represents common menu button entries type.
//
type MenuBtn int

// Common menu button entries.
const (
	MenuWindowClose MenuBtn = iota
	MenuWindowCloseAll
	MenuWindowMax
	MenuWindowMin
	MenuWindowMinAll
	MenuWindowShow
	MenuWindowShowAll
)

// DockMenu represents a dock menu.
//
type DockMenu struct {
	menus.Menu
	Icon      gldi.Icon
	Container *gldi.Container
	Dock      *gldi.CairoDock // just a pointer to container with type dock.
}

// WrapDockMenu wraps a menu as DockMenu.
//
func WrapDockMenu(icon gldi.Icon, container *gldi.Container, dock *gldi.CairoDock, menu *gtk.Menu) *DockMenu {
	usermenu := menus.WrapMenu(menu)
	usermenu.SetCallNewItem(gldi.MenuAddItem)
	usermenu.SetCallNewSubMenu(gldi.MenuAddSubMenu)

	return &DockMenu{
		Menu:      *usermenu,
		Icon:      icon,
		Container: container,
		Dock:      dock}
}

// WrapDockMenuCallback returns a DockMenu wrapper function.
//
func WrapDockMenuCallback(call func(dm *DockMenu) bool) func(icon gldi.Icon, container *gldi.Container, dock *gldi.CairoDock, menu *gtk.Menu) bool {
	return func(icon gldi.Icon, container *gldi.Container, dock *gldi.CairoDock, menu *gtk.Menu) bool {
		return call(WrapDockMenu(icon, container, dock, menu))
	}
}

// AddSubMenu adds a submenu to the menu.
//
func (m *DockMenu) AddSubMenu(label, iconPath string) *DockMenu {
	submenu := m.Menu.AddSubMenu(label, iconPath)
	return WrapDockMenu(m.Icon, m.Container, m.Dock, &submenu.Menu)
}

// Entry adds a defined entry to the menu. Returns if a separator is needed.
//
func (m *DockMenu) Entry(entry MenuEntry) bool {
	switch entry {

	case MenuAbout:
		m.AddEntry(
			tran.Slate("About"),
			globals.IconNameAbout,
			about.New,
		)

	case MenuAddApplet:
		m.AddEntry(
			tran.Slate("Applet"),
			globals.IconNameAdd,
			backendgui.ShowAddons)

	case MenuAddLauncher:
		m.AddEntry(
			tran.Slate("Custom launcher"),
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
		).SetTooltipText(tran.Slate("Usually you would drag a launcher from the menu and drop it on the dock."))

	case MenuAddMainDock:
		m.AddEntry(
			tran.Slate("Main dock"),
			globals.IconNameAdd,
			func() {
				key := gldi.DockAddConfFile()
				gldi.DockNew(key)

				backendgui.ReloadItems()

				// g_timeout_add_seconds (1, (GSourceFunc)_show_new_dock_msg, cDockName);  // delai, car sa fenetre n'est pas encore bien placee (0,0).
			})

	case MenuAddSeparator:
		m.AddEntry(
			tran.Slate("Separator"),
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
			tran.Slate("Sub-dock"),
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
		items := m.Icon.GetClass().MenuItems()
		for _, it := range items {
			cmd := strings.Fields(it[1])
			m.AddEntry(
				it[0], // name
				it[2], // icon
				func() { log.ExecAsync(cmd[0], cmd[1:]...) }) // was gldi.LaunchCommand(cmd)
		}
		return len(items) > 0

	case MenuConfigure:
		m.AddEntry(
			tran.Slate("Configure"),
			globals.IconNamePreferences,
			backendgui.ShowMainGui,
		).SetTooltipText(tran.Slate("Configure behaviour, appearance, and applets."))

	case MenuCustomIconRemove:
		classIcon := m.Icon.GetClass().Icon()
		if classIcon == "" {
			classIcon = m.Icon.GetClass().String()
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
				tran.Slate("Remove custom icon"),
				globals.IconNameRemove,
				func() {
					C._cairo_dock_remove_custom_appli_icon((*C.Icon)(unsafe.Pointer(m.Icon.ToNative())), (*C.CairoDock)(unsafe.Pointer(m.Dock.Ptr)))
				})
		}

	case MenuCustomIconSet:
		m.AddEntry(
			tran.Slate("Set a custom icon"),
			globals.IconNameSelectColor,
			func() {
				C._cairo_dock_set_custom_appli_icon((*C.Icon)(unsafe.Pointer(m.Icon.ToNative())), (*C.CairoDock)(unsafe.Pointer(m.Dock.Ptr)))
			})

	case MenuDeleteDock:
		m.AddEntry(
			tran.Slate("Delete this dock"),
			globals.IconNameDelete,
			func() {
				if m.Dock.GetRefCount() != 0 {
					dialog.ShowWithQuestion("Delete this dock?",
						m.Dock.GetPointedIcon(),
						m.Container,
						globals.FileCairoDockIcon(),
						cbDialogIsOK(func() {
							gldi.ObjectDelete(m.Dock)
						}))
				}
			})

	case MenuDeskletLock:
		desklet := m.Container.ToDesklet()
		m.AddCheckEntry(
			tran.Slate("Lock position"),
			desklet.PositionLocked(),
			func(c *gtk.CheckMenuItem) {
				desklet.LockPosition(c.GetActive())
				backendgui.UpdateDeskletVisibility(desklet)
			})

	case MenuDeskletSticky:
		desklet := m.Container.ToDesklet()
		m.AddCheckEntry(
			tran.Slate("On all desktops"),
			m.Container.ToDesklet().IsSticky(),
			func(c *gtk.CheckMenuItem) {
				desklet.SetSticky(c.GetActive())

				// TODO: check why SetSticky isn't working...
				println("get stick", desklet.IsSticky())

				backendgui.UpdateDeskletVisibility(desklet)
			})

	case MenuDeskletVisibility:
		submenu := m.AddSubMenu(tran.Slate("Visibility"), globals.IconNameFind)

		desklet := m.Container.ToDesklet()

		callback := func(radio *gtk.RadioMenuItem, val cdtype.DeskletVisibility) {
			if radio.GetActive() {
				desklet.SetVisibility(val, true) // with true, also save to conf.
				backendgui.UpdateDeskletVisibility(desklet)
			}
		}

		group := 42
		visibility := desklet.Visibility()
		addRadio := func(vis cdtype.DeskletVisibility, label string) {
			radio := submenu.AddRadioEntry(label, visibility == vis, group, nil)
			radio.Connect("toggled", callback, vis)
		}

		addRadio(cdtype.DeskletVisibilityNormal, tran.Slate("Normal"))
		addRadio(cdtype.DeskletVisibilityKeepAbove, tran.Slate("Always on top"))
		addRadio(cdtype.DeskletVisibilityKeepBelow, tran.Slate("Always below"))
		if gldi.CanSetOnWidgetLayer() {
			addRadio(cdtype.DeskletVisibilityWidgetLayer, tran.Slate("Widget Layer"))
		}
		addRadio(cdtype.DeskletVisibilityReserveSpace, tran.Slate("Reserve space"))

	case MenuDetachApplet:
		m.AddEntry(
			ternary.String(m.Dock != nil, tran.Slate("Detach"), tran.Slate("Return to the dock")),
			ternary.String(m.Dock != nil, globals.IconNameGotoTop, globals.IconNameGotoBottom),
			func() {
				// if (CAIRO_DOCK_IS_DESKLET (pContainer))
				// 	icon = (CAIRO_DESKLET (pContainer))->pIcon;  // l'icone cliquee du desklet n'est pas forcement celle qui contient le module !
				// g_return_if_fail (CAIRO_DOCK_IS_APPLET (icon));

				m.Icon.ModuleInstance().Detach()
			})

	case MenuDuplicateApplet:
		m.AddEntry(
			tran.Slate("Duplicate"),
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
			tran.Slate("Edit"),
			globals.IconNameEdit,
			func() {
				// if (CAIRO_DOCK_IS_DESKLET (pContainer))
				// 	icon = (CAIRO_DESKLET (pContainer))->pIcon;  // l'icone cliquee du desklet n'est pas forcement celle qui contient le module.
				// g_return_if_fail (CAIRO_DOCK_IS_APPLET (icon));
				backendgui.ShowItems(m.Icon, nil, nil, -1)
			})

	case MenuEditDock:
		m.AddEntry(
			tran.Slate("Configure this dock"),
			globals.IconNameExecute,
			func() { backendgui.ShowItems(nil, m.Container, nil, 0) },
		).SetTooltipText(tran.Slate("Customize the position, visibility and appearance of this main dock."))

	case MenuEditIcon:
		m.AddEntry(tran.Slate(
			tran.Slate("Edit")),
			globals.IconNameEdit,
			func() {
				switch m.Icon.GetDesktopFileName() {
				case "", "none":
					dialog.ShowTemporaryWithIcon("Sorry, this icon doesn't have a configuration file.", m.Icon, m.Container, 4000, "same icon")

				default:
					backendgui.ShowItems(m.Icon, nil, nil, -1)
				}
			})

	case MenuEditTarget:
		_, img := m.Icon.DefaultNameIcon()
		m.AddEntry(
			tran.Slate("Edit icon"),
			img,
			func() {
				// if (CAIRO_DOCK_IS_DESKLET (pContainer))
				// 	icon = (CAIRO_DESKLET (pContainer))->pIcon;  // l'icone cliquee du desklet n'est pas forcement celle qui contient le module.
				// g_return_if_fail (CAIRO_DOCK_IS_APPLET (icon));
				backendgui.ShowItems(m.Icon, nil, nil, -1)
			})

	case MenuHandbook:
		m.AddEntry(
			tran.Slate("Applet's handbook"),
			globals.IconNameAbout,
			m.Icon.ModuleInstance().PopupAboutApplet)

		// handbook + 5x to facto
		// picon := icon
		// if container.IsDesklet() {
		// 	picon = container.Icon()
		// 	icon = (CAIRO_DESKLET (pContainer))->pIcon;  // l'icone cliquee du desklet n'est pas forcement celle qui contient le module !
		// }

	case MenuHelp:
		m.AddEntry(
			tran.Slate("Help"),
			globals.IconNameHelp,
			func() { backendgui.ShowModuleGui("Help") },
		).SetTooltipText(tran.Slate("There are no problems, only solutions (and a lot of useful hints!)"))

	case MenuLaunchNew:
		m.AddEntry(
			tran.Slate("Launch a new (Shift+clic)"),
			globals.IconNameAdd,
			func() { m.Icon.LaunchCommand(log) })

	case MenuLockIcons:
		m.AddCheckEntry(
			tran.Slate("Lock icons position"),
			current.Docks.LockIcons(),
			func() {
				current.Docks.LockIcons(!current.Docks.LockIcons())
				config.UpdateFile(log, globals.ConfigFile(), "Accessibility", "lock icons", current.Docks.LockIcons())
			},
		).SetTooltipText(tran.Slate("This will (un)lock the position of the icons."))

	case MenuMakeLauncher:
		m.AddEntry(
			tran.Slate("Make it a launcher"),
			globals.IconNameNew,
			func() {
				C._cairo_dock_make_launcher_from_appli((*C.Icon)(unsafe.Pointer(m.Icon.ToNative())), (*C.CairoDock)(unsafe.Pointer(m.Dock.Ptr)))
			})

	case MenuMoveToDesktopClass:
		m.moveToDesktop(true)

	case MenuMoveToDesktopWindow:
		m.moveToDesktop(false)

	case MenuMoveToDock:
		sub := m.AddSubMenu(tran.Slate("Move to another dock"), globals.IconNameJumpTo)

		docks := gldi.GetAllAvailableDocks(m.Icon.GetContainer().ToCairoDock(), m.Icon.GetSubDock())
		docks = append(docks, nil)
		for _, dock := range docks {
			name := ""
			key := ""
			icon := ""
			if dock == nil {
				name = tran.Slate("New main dock")
				icon = globals.IconNameAdd // globals.IconNameNew
			} else {
				name = ternary.String(dock.GetReadableName() != "", dock.GetReadableName(), dock.GetDockName())
				key = dock.GetDockName()

				if dock.GetRefCount() == 0 { // Maindocks icon
					switch dock.Container().ScreenBorder() {
					case cdtype.ContainerPositionBottom:
						icon = "go-down"
					case cdtype.ContainerPositionTop:
						icon = "go-up"
					case cdtype.ContainerPositionLeft:
						icon = "go-previous"
					case cdtype.ContainerPositionRight:
						icon = "go-next"
					}
				} else { // Subdocks icon.
					dockicon := dock.SearchIconPointingOnDock(nil)
					if dockicon != nil {
						icon = dockicon.GetFileName()
					}
				}

			}
			sub.AddEntry(
				name,
				icon,
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
						dialog.ShowGeneralMessage(str, 8000) // we don't show it on the new dock as its window isn't positioned yet (0,0).
					}
				})
		}

	case MenuQuickHide:
		m.AddEntry(
			tran.Slate("Quick-Hide"),
			globals.IconNameGotoBottom,
			gldi.QuickHideAllDocks,
		).SetTooltipText(tran.Slate("This will hide the dock until you hover over it with the mouse."))

	case MenuQuit:
		item := m.AddEntry(
			tran.Slate("Quit"),
			globals.IconNameQuit,
			func() {
				backendgui.CloseGui()

				// gtk.MainQuit() // TODO: remove SQP HACK, easy quit no confirm for tests.

				dialog.ShowWithQuestion("Quit Cairo-Dock?",
					GetIconForDesklet(m.Icon, m.Container),
					m.Container,
					globals.FileCairoDockIcon(),
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
			key := &gdk.EventKey{Event: event}

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

				dialog.ShowWithQuestion(
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
				dialog.ShowWithQuestion(
					fmt.Sprintf(tran.Slate("You're about to remove this icon (%s) from the dock. Are you sure?"), name),
					m.Icon,
					m.Container,
					"same icon",
					cbDialogIsOK(func() {
						if m.Icon.IsStackIcon() && m.Icon.GetSubDock() != nil && len(m.Icon.GetSubDock().Icons()) > 0 {
							dialog.ShowWithQuestion(
								tran.Slate("Do you want to re-dispatch the icons contained inside this container into the dock?\n(otherwise they will be destroyed)"),
								m.Icon,
								m.Container,
								globals.FileCairoDockIcon(),
								cbDialogIsOK(func() {
									m.Icon.RemoveIconsFromSubdock(m.Dock)
								}))
						}

						m.Icon.RemoveFromDock()
					}))
			}).SetTooltipText(tran.Slate("You can remove a launcher by dragging it out of the dock with the mouse ."))

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
			ternary.String(flag, tran.Slate("Don't keep above"), tran.Slate("Keep above")),
			ternary.String(flag, globals.IconNameGotoBottom, globals.IconNameGotoTop),
			m.Icon.CallbackActionWindowToggle((cdglobal.Window).SetAbove, (cdglobal.Window).IsAbove))

	case MenuWindowBelow:
		if !m.Icon.Window().IsHidden() { // this could be a button in the menu, if we find an icon that doesn't look too much like the "minimise" one
			m.AddEntry(
				tran.Slate("Below other windows")+actionMiddleClick(m.Icon, 4),
				globals.DirShareData("icons", "icon-lower.svg"),
				m.Icon.CallbackActionWindow((cdglobal.Window).Lower))
		}

	case MenuWindowFullScreen:
		flag := m.Icon.Window().IsFullScreen()
		m.AddEntry(
			ternary.String(flag, tran.Slate("Not Fullscreen"), tran.Slate("Fullscreen")),
			ternary.String(flag, globals.IconNameLeaveFullScreen, globals.IconNameFullScreen),
			m.Icon.CallbackActionWindowToggle((cdglobal.Window).SetFullScreen, (cdglobal.Window).IsFullScreen))

	case MenuWindowKill:
		m.AddEntry(
			tran.Slate("Kill"),
			globals.IconNameClose,
			m.Icon.CallbackActionWindow((cdglobal.Window).Kill))

	case MenuWindowMoveAllHere:
		m.AddEntry(
			tran.Slate("Move all to this desktop"),
			globals.IconNameJumpTo,
			m.Icon.CallbackActionSubWindows((cdglobal.Window).MoveToCurrentDesktop))

	case MenuWindowMoveHere:
		var callback func()
		if !m.Icon.Window().IsOnCurrentDesktop() {
			callback = m.Icon.CallbackActionWindow(func(win cdglobal.Window) {
				win.MoveToCurrentDesktop()
				if !win.IsHidden() {
					win.Show()
				}
			})
		}
		m.AddEntry(tran.Slate("Move to this desktop"), globals.IconNameJumpTo, callback)

	case MenuWindowSticky:
		m.AddEntry(
			ternary.String(m.Icon.Window().IsSticky(), tran.Slate("Visible only on this desktop"), tran.Slate("Visible on all desktops")),
			globals.IconNameJumpTo,
			m.Icon.CallbackActionWindowToggle((cdglobal.Window).SetSticky, (cdglobal.Window).IsSticky))

	}
	return notif.AnswerLetPass
}

//
//-----------------------------------------------------------[ BUTTONS ENTRY ]--

// AddButtonsEntry adds an item with button entries to the menu.
//
func (m *DockMenu) AddButtonsEntry(label string, btns ...MenuBtn) *menus.ButtonsEntry {
	entry := m.Menu.AddButtonsEntry(label)

	for _, btnID := range btns {
		switch btnID {
		case MenuWindowClose:
			entry.AddButton(
				tran.Slate("Close")+actionMiddleClick(m.Icon, 1),
				globals.DirShareData("icons", "icon-close.svg"),
				m.Icon.CallbackActionWindow((cdglobal.Window).Close))

		case MenuWindowCloseAll:
			entry.AddButton(
				tran.Slate("Close all")+actionMiddleClick(m.Icon, 1),
				globals.DirShareData("icons", "icon-close.svg"),
				m.Icon.CallbackActionSubWindows((cdglobal.Window).Close))

		case MenuWindowMax:
			max := m.Icon.Window().IsMaximized()
			entry.AddButton(
				ternary.String(max, tran.Slate("Unmaximise"), tran.Slate("Maximise")),
				globals.DirShareData("icons", ternary.String(max, "icon-restore.svg", "icon-maximize.svg")),
				m.Icon.CallbackActionWindowToggle((cdglobal.Window).Maximize, (cdglobal.Window).IsMaximized))

		case MenuWindowMin:
			entry.AddButton(
				tran.Slate("Minimise")+actionMiddleClick(m.Icon, 2),
				globals.DirShareData("icons", "icon-minimize.svg"),
				m.Icon.CallbackActionWindow((cdglobal.Window).Minimize))

		case MenuWindowMinAll:
			entry.AddButton(
				tran.Slate("Minimise all")+actionMiddleClick(m.Icon, 2),
				globals.DirShareData("icons", "icon-minimize.svg"),
				m.Icon.CallbackActionSubWindows((cdglobal.Window).Minimize))

		case MenuWindowShow:
			entry.AddButton(
				tran.Slate("Show"),
				globals.IconNameFind,
				m.Icon.CallbackActionWindow((cdglobal.Window).Show))

		case MenuWindowShowAll:
			entry.AddButton(
				tran.Slate("Show all"),
				globals.IconNameFind,
				m.Icon.CallbackActionSubWindows((cdglobal.Window).Show))

		default:
			log.NewWarn(fmt.Sprintf("invalid id: %d", btnID), "AddButtonsEntry")
		}
	}
	return entry
}

//
//--------------------------------------------------------[ DOCKMENU HELPERS ]--

// GetIconForDesklet will return the correct icon if clicked on a desklet.
//
func GetIconForDesklet(icon gldi.Icon, container *gldi.Container) gldi.Icon {
	if container.IsDesklet() {
		return container.ToDesklet().GetIcon()
	}
	return icon
}

// actionMiddleClick returns the middle-click string to add to the action button
// if it's matching the action ID provided.
//
func actionMiddleClick(icon gldi.Icon, id int) string {
	if current.Taskbar.ActionOnMiddleClick() != id || icon.IsApplet() {
		return ""
	}
	return " (" + tran.Slate("middle-click") + ")"
}

// nextOrder calculates an order position for a new icon, based on the mouse
// position on current icon, and the order of the current icon and the next one.
//
func nextOrder(icon gldi.Icon, dock *gldi.CairoDock) float64 {
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

// cbDialogIsOK prepares a callback for DialogWithQuestion to launch on user confirmation.
//
func cbDialogIsOK(call func()) func(int, *gtk.Widget) {
	return func(clickedButton int, widget *gtk.Widget) {
		if clickedButton == cdtype.DialogButtonFirst || clickedButton == cdtype.DialogKeyEnter {
			call()
		}
	}
}

//
//----------------------------------------------------------[ DESKTOP ENTRIES]--

func (m *DockMenu) moveToDesktop(useAll bool) {
	if desktops.NbDesktops() < 2 && desktops.NbViewportX() < 2 && desktops.NbViewportY() < 2 {
		return
	}

	m.AddSeparator()
	win := m.Icon.Window()

	moveto := newMenuMoveToDesktop(m.Icon, useAll)

	for i := 0; i < desktops.NbDesktops(); i++ { // sort by desktop

		for j := 0; j < desktops.NbViewportY(); j++ { // and by columns.

			for k := 0; k < desktops.NbViewportX(); k++ { // then rows.

				entry := m.AddEntry(moveto.Label(i, j, k), "", moveto.Callback(i, j, k))

				if win != nil && win.IsOnDesktop(i, j, k) {
					entry.SetSensitive(false)
				}
			}
		}
	}
}

type menuMoveToDesktop struct {
	format   string
	mode     int
	nbx      int
	cbAction func(call func(cdglobal.Window)) func()
}

func newMenuMoveToDesktop(icon gldi.Icon, useAll bool) *menuMoveToDesktop {
	// Create object and set work mode.
	desk := &menuMoveToDesktop{nbx: desktops.NbViewportX()}
	switch {
	case desktops.NbDesktops() > 1 && (desktops.NbViewportX() > 1 || desktops.NbViewportY() > 1):
		desk.mode = 2

	case desktops.NbDesktops() > 1:
		desk.mode = 1
	}

	// Set label format string.
	switch {
	case desk.mode == 2 && useAll:
		desk.format = tran.Slate("Move all to desktop %d - face %d")

	case desk.mode == 2:
		desk.format = tran.Slate("Move to desktop %d - face %d")

	case desk.mode == 1 && useAll:
		desk.format = tran.Slate("Move all to desktop %d")

	case desk.mode == 1:
		desk.format = tran.Slate("Move to desktop %d")

	case desk.mode == 0 && useAll:
		desk.format = tran.Slate("Move all to face %d")

	case desk.mode == 0:
		desk.format = tran.Slate("Move to face %d")
	}

	// Set the window action method.
	if useAll {
		desk.cbAction = icon.CallbackActionSubWindows // All windows of class.
	} else {
		desk.cbAction = icon.CallbackActionWindow // One window.
	}

	return desk
}

// Label returns the formated text for the label (move to desktop x).
//
func (g *menuMoveToDesktop) Label(i, j, k int) string {
	var args []interface{}
	switch g.mode {
	case 2:
		args = []interface{}{i + 1, g.nbx*j + k + 1}

	case 1:
		args = []interface{}{i + 1}

	case 0:
		args = []interface{}{g.nbx*j + k + 1}
	}

	return fmt.Sprintf(g.format, args...)
}

// Callback returns a move to action callback.
//
func (g *menuMoveToDesktop) Callback(i, j, k int) func() {
	return g.cbAction(func(win cdglobal.Window) { win.MoveToDesktop(i, j, k) })
}
