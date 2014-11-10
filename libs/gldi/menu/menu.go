// Package menu fills the main menu for cairo-dock.
package menu

import (
	"github.com/sqp/godock/libs/gldi"             // Gldi access.
	"github.com/sqp/godock/libs/gldi/backendmenu" // Menu types.
	"github.com/sqp/godock/libs/gldi/globals"     // Global variables.
	"github.com/sqp/godock/libs/ternary"          // Helpers.
	"github.com/sqp/godock/libs/tran"             // Translate.
)

func BuildMenuContainer(m *backendmenu.DockMenu) int {

	if m.Container.IsDesklet() && m.Icon != nil && !m.Icon.IsApplet() { // not on the icons of a desklet, except the applet icon (on a desklet, it's easy to click out of any icon).
		return backendmenu.LetPass
	}

	if m.Dock != nil && m.Dock.GetRefCount() > 0 { // not on the sub-docks, except user sub-docks.
		pointingIcon := m.Dock.SearchIconPointingOnDock(nil)
		if pointingIcon != nil && !pointingIcon.IsStackIcon() {
			return backendmenu.LetPass
		}
	}

	if m.Dock != nil && (m.Icon == nil || m.Icon.IsSeparatorAuto()) {
		return backendmenu.LetPass
	}

	//\_________________________ First item is the Cairo-Dock sub-menu.
	dockmenu, _ := m.SubMenu("Cairo-Dock", globals.DirShareData(globals.CairoDockIcon))

	if !globals.DockIsLocked() {
		dockmenu.Entry(backendmenu.MenuConfigure)

		if m.Dock != nil && !m.Dock.IsMainDock() && m.Dock.GetRefCount() == 0 { // root dock settings
			dockmenu.Entry(backendmenu.MenuEditDock)
			dockmenu.Entry(backendmenu.MenuDeleteDock)
		}

		// 	if backendgui.CanManageThemes () {// themes. Still to do.
		// dockmenu.Entry(backendmenu.MenuThemes)
		// 	}

		// add new item
		if m.Dock != nil {
			sub, _ := dockmenu.SubMenu(tran.Slate("Add"), globals.IconNameAdd)
			sub.Entry(backendmenu.MenuAddSubDock)
			sub.Entry(backendmenu.MenuAddMainDock)
			sub.Entry(backendmenu.MenuAddSeparator)
			sub.Entry(backendmenu.MenuAddLauncher)
			sub.Entry(backendmenu.MenuAddApplet)
		}

		dockmenu.Separator()
		dockmenu.Entry(backendmenu.MenuLockIcons)
	}

	if m.Dock != nil && !m.Dock.IsAutoHide() {
		dockmenu.Entry(backendmenu.MenuQuickHide)
	}

	if !globals.FullLock {
		// dockmenu.Entry(backendmenu.MenuAutostart) // removed crap.
		// dockmenu.Entry(backendmenu.MenuThirdParty)

		dockmenu.Entry(backendmenu.MenuHelp) // Don't show if locked, because it would open the configuration window.
	}

	dockmenu.Entry(backendmenu.MenuAbout)

	if !globals.FullLock {
		dockmenu.Entry(backendmenu.MenuQuit)
	}

	// //\_________________________ Second item is the Icon sub-menu.

	// Icon *pIcon = icon;
	// if (pIcon == NULL && CAIRO_DOCK_IS_DESKLET (pContainer))  // on a desklet, the main applet icon may not be drawn; therefore we add the applet sub-menu if we clicked outside of an icon.
	// {
	// 	pIcon = CAIRO_DESKLET (pContainer)->pIcon;
	// }

	// pIcon := GetIconForDesklet(icon, container)

	if m.Icon != nil && !m.Icon.IsSeparatorAuto() {

		items, _ := m.SubMenu(DefaultNameIcon(m.Icon))

		// 	GtkWidget *pItemSubMenu = _add_item_sub_menu (pIcon, menu);

		if globals.DockIsLocked() {
			switch {
			case m.Icon.IsAppli() && m.Icon.GetCommand() != "":
				items.Entry(backendmenu.MenuLaunchNew)

			case m.Icon.IsApplet():
				items.Entry(backendmenu.MenuHandbook)

			default:
				items.SetSensitive(false) // empty, the submenu is added for consistency between icons, but disabled.
			}

		} else {
			if m.Icon.IsAppli() && m.Icon.GetCommand() != "" {
				items.Entry(backendmenu.MenuLaunchNew)
			}

			switch {
			case m.Icon.GetDesktopFileName() != "" &&
				(m.Icon.IsLauncher() || m.Icon.IsStackIcon() || m.Icon.IsSeparator()):

				items.Entry(backendmenu.MenuEditIcon)
				items.Entry(backendmenu.MenuRemoveIcon)
				items.Entry(backendmenu.MenuMoveToDock)

			case m.Icon.IsAppli() || m.Icon.IsStackIcon(): // appli with no launcher

				if !m.Icon.ClassIsInhibited() { // if the class doesn't already have an inhibator somewhere.
					items.Entry(backendmenu.MenuMakeLauncher)

					if !globals.DocksParam.IsLockAll() && m.Icon.IsAppli() {

						if globals.TaskbarParam.OverWriteXIcons() {
							items.Entry(backendmenu.MenuCustomIconRemove)
						}

						items.Entry(backendmenu.MenuCustomIconSet)
					}
				}

			case m.Icon.IsApplet(): // applet (icon or desklet) (the sub-icons have been filtered before and won't have this menu).
				items.Entry(backendmenu.MenuEditApplet)

				if m.Icon.IsDetachableApplet() {
					items.Entry(backendmenu.MenuDetachApplet)
				}

				items.Entry(backendmenu.MenuRemoveApplet)

				if m.Icon.ModuleInstance().Module().VisitCard().IsMultiInstance() {
					items.Entry(backendmenu.MenuDuplicateApplet)
				}

				if m.Dock != nil && m.Icon.GetContainer() != nil {
					items.Entry(backendmenu.MenuMoveToDock)
				}

				items.Separator()

				items.Entry(backendmenu.MenuHandbook)

			}
		}
	}

	return backendmenu.LetPass
}

func BuildMenuIcon(m *backendmenu.DockMenu) int {

	//\_________________________ Clic en-dehors d'une icone utile => on s'arrete la.
	if m.Dock != nil && (m.Icon == nil || m.Icon.IsSeparatorAuto()) {
		return backendmenu.LetPass
	}

	needSeparator := true

	if m.Container.IsDesklet() && m.Icon != nil && !m.Icon.IsApplet() { // not on the icons of a desklet, except the applet icon (on a desklet, it's easy to click out of any icon).
		needSeparator = false
	}

	if m.Dock != nil && m.Dock.GetRefCount() > 0 { // not on the sub-docks, except user sub-docks.
		pointingIcon := m.Dock.SearchIconPointingOnDock(nil)
		if pointingIcon != nil && !pointingIcon.IsStackIcon() {
			needSeparator = false
		}
	}

	// 	//\_________________________ class actions.
	if m.Icon != nil && m.Icon.GetClass() != "" && !m.Icon.GetIgnoreQuickList() {
		m.Entry(backendmenu.MenuClassItems)
	}

	//\_________________________ Actions on applications.
	if m.Icon.IsAppli() {
		if needSeparator {
			m.Separator()
		}
		needSeparator = true

		appli := m.Icon.Window()
		canMin, canMax, canClose := appli.CanMinMaxClose()

		m.AddButtonsEntry(tran.Slate("Window"))
		if canClose {
			m.Button(backendmenu.MenuWindowClose)
		}

		if !appli.IsHidden() {
			if canMax {
				m.Button(backendmenu.MenuWindowMax)
			}
			if canMin {
				m.Button(backendmenu.MenuWindowMin)
			}
		}

		if appli.IsHidden() || !appli.IsActive() {
			// 			 || !gldi_window_is_on_current_desktop (pAppli)))

			m.Button(backendmenu.MenuWindowShow)
		}

		//\_________________________ Other actions

		otherActions, _ := m.SubMenu(tran.Slate("Other actions"), "")

		otherActions.Entry(backendmenu.MenuWindowMoveHere)
		otherActions.Entry(backendmenu.MenuWindowFullScreen)
		otherActions.Entry(backendmenu.MenuWindowBelow)
		otherActions.Entry(backendmenu.MenuWindowAbove)
		otherActions.Entry(backendmenu.MenuWindowSticky)

		// _add_desktops_entry (otherActions.Menu, FALSE, data)

		otherActions.Separator()

		otherActions.Entry(backendmenu.MenuWindowKill)

	} else if m.Icon.IsMultiAppli() { // Window management
		if needSeparator {
			m.Separator()
		}
		needSeparator = true

		m.AddButtonsEntry("Windows")
		m.Button(backendmenu.MenuWindowCloseAll)
		m.Button(backendmenu.MenuWindowMinAll)
		m.Button(backendmenu.MenuWindowShowAll)

		otherActions, _ := m.SubMenu(tran.Slate("Other actions"), "")
		otherActions.Entry(backendmenu.MenuWindowMoveAllHere)

		// 		_add_desktops_entry (pSubMenuOtherActions, TRUE, data);
	}

	//\_________________________ Desklet positioning actions.

	if !globals.DockIsLocked() && m.Container.IsDesklet() {
		if needSeparator {
			m.Separator()
		}
		needSeparator = true

		m.Entry(backendmenu.MenuDeskletVisibility)
		m.Entry(backendmenu.MenuDeskletSticky)
		m.Entry(backendmenu.MenuDeskletLock)
	}

	return backendmenu.LetPass
}

//
//-------------------------------------------------------------[ MENU ENTRIES]--

// func (menu *Menu) addItemSubMenu(icon *gldi.Icon) {
// name := ""
// switch {
// case icon.IsLauncher(), icon.IsStackIcon():
// 	name = ternary.String(icon.GetInitialName()!= "", icon.GetInitialName(), icon.GetName())

// case icon.IsAppli(), icon.IsClassIcon():
// 	name = icon.GetClassInfo(gldi.ClassName)
// 	if name == "" {
// 		name = icon.GetClass()
// 	}

// case icon.IsApplet():
// 	name = icon.ModuleInstance().Module().VisitCard().GetTitle()

// case icon.IsSeparator():
// 	name = tran.Slate("Separator")

// default:
// 	name = icon.GetName()
// }

// const gchar *cName = NULL;
// if (CAIRO_DOCK_ICON_TYPE_IS_LAUNCHER (icon) || CAIRO_DOCK_ICON_TYPE_IS_CONTAINER (icon))
// {
// 	cName = (icon->cInitialName ? icon->cInitialName : icon->cName);
// }
// else if (CAIRO_DOCK_ICON_TYPE_IS_APPLI (icon) || CAIRO_DOCK_ICON_TYPE_IS_CLASS_CONTAINER (icon))
// {
// 	cName = cairo_dock_get_class_name (icon->cClass);  // better than the current window title.
// 	if (cName == NULL)
// 		cName = icon->cClass;
// }
// else if (CAIRO_DOCK_IS_APPLET (icon))
// {
// 	cName = icon->pModuleInstance->pModule->pVisitCard->cTitle;
// }
// else if (CAIRO_DOCK_ICON_TYPE_IS_SEPARATOR (icon))
// {
// 	cName = _("Separator");
// }
// else
// 	cName = icon->cName;

// img := ""
// switch {
// case icon.IsApplet():
// }

// gchar *cIconFile = NULL;
// if (CAIRO_DOCK_IS_APPLET (icon))
// {
// 	if (icon->cFileName != NULL)  // if possible, use the actual icon
// 		cIconFile = cairo_dock_search_icon_s_path (icon->cFileName, cairo_dock_search_icon_size (GTK_ICON_SIZE_LARGE_TOOLBAR));
// 	if (!cIconFile)  // else, use the default applet's icon.
// 		cIconFile = cairo_dock_search_icon_s_path (icon->pModuleInstance->pModule->pVisitCard->cIconFilePath, cairo_dock_search_icon_size (GTK_ICON_SIZE_LARGE_TOOLBAR));
// }
// else if (CAIRO_DOCK_ICON_TYPE_IS_SEPARATOR (icon))
// {
// 	if (myIconsParam.cSeparatorImage)
// 		cIconFile = cairo_dock_search_image_s_path (myIconsParam.cSeparatorImage);
// }
// else if (icon->cFileName != NULL)
// {
// 	cIconFile = cairo_dock_search_icon_s_path (icon->cFileName, cairo_dock_search_icon_size (GTK_ICON_SIZE_LARGE_TOOLBAR));
// }
// if (cIconFile == NULL && icon->cClass != NULL)
// {
// 	const gchar *cClassIcon = cairo_dock_get_class_icon (icon->cClass);
// 	if (cClassIcon)
// 		cIconFile = cairo_dock_search_icon_s_path (cClassIcon, cairo_dock_search_icon_size (GTK_ICON_SIZE_LARGE_TOOLBAR));
// }

// GtkWidget *pItemSubMenu;
// GdkPixbuf *pixbuf = NULL;

// if (!cIconFile)  // no icon file (for instance a class that has no icon defined in its desktop file, like gnome-setting-daemon) => use its buffer directly.
// {
// 	pixbuf = cairo_dock_icon_buffer_to_pixbuf (icon);
// }

// if (pixbuf)
// {
// 	GtkWidget *pMenuItem = NULL;
// 	pItemSubMenu = gldi_menu_add_sub_menu_full (pMenu, cName, "", &pMenuItem);

// 	GtkWidget *image = gtk_image_new_from_pixbuf (pixbuf);
// 	gldi_menu_item_set_image (pMenuItem, image);
// 	g_object_unref (pixbuf);
// }
// else
// {
// 	pItemSubMenu = cairo_dock_create_sub_menu (cName, pMenu, cIconFile);
// }

// g_free (cIconFile);
// return pItemSubMenu;
// }

// DefaultNameIcon returns improved name and image for the icon if possible.
// TODO: facto with version in datagldi
//
func DefaultNameIcon(icon *gldi.Icon) (name, img string) {
	switch {
	case icon.IsApplet():
		vc := icon.ModuleInstance().Module().VisitCard()
		return vc.GetTitle(), vc.GetIconFilePath()

	case icon.IsSeparator():
		return "--------", ""

	case icon.IsLauncher(), icon.IsStackIcon(), icon.IsAppli(), icon.IsClassIcon():
		name := icon.GetClassInfo(gldi.ClassName)
		if name != "" {
			return name, icon.GetFileName() // icon.GetClassInfo(ClassIcon)
		}
		return ternary.String(icon.GetInitialName() != "", icon.GetInitialName(), icon.GetName()), icon.GetFileName()

	}
	return icon.GetName(), icon.GetFileName()
}
