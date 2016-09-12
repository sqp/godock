// Package eventmouse registers and handles icon click events.
package eventmouse

import (
	"github.com/gotk3/gotk3/gdk"

	"github.com/sqp/godock/libs/cdglobal"     // Dock types.
	"github.com/sqp/godock/libs/cdtype"       // Applet types.
	"github.com/sqp/godock/libs/dock/confown" // New dock own settings.
	"github.com/sqp/godock/libs/gldi"
	"github.com/sqp/godock/libs/gldi/current"  // Current theme settings.
	"github.com/sqp/godock/libs/gldi/desktops" // Desktop and screens info.
	"github.com/sqp/godock/libs/gldi/notif"    // Dock notifs.

	"strings"
)

var log cdtype.Logger

// Register registers mouse events and sets the logger.
//
func Register(l cdtype.Logger) {
	log = l
	notif.RegisterContainerLeftClick(OnLeftClick)
	notif.RegisterContainerMiddleClick(OnMiddleClick)
	notif.RegisterContainerMouseScroll(OnMouseScroll)
	notif.RegisterContainerDropData(OnDropData)
}

// OnLeftClick triggers a dock left click action on the icon.
//
func OnLeftClick(icon gldi.Icon, container *gldi.Container, btnState uint) bool {
	switch {
	case icon == nil || !gldi.ObjectIsDock(container):
		log.Debug("notifClickIcon", "ignored: no icon or dock target")
		return notif.AnswerLetPass

	// With shift or ctrl on an icon that is linked to a program => re-launch this program.
	case gdk.ModifierType(btnState)&(gdk.GDK_SHIFT_MASK|gdk.GDK_CONTROL_MASK) > 0:
		if icon.IsLauncher() || icon.IsAppli() || icon.IsStackIcon() {
			icon.LaunchCommand(log)
		}
		return notif.AnswerLetPass

	// scale on an icon holding a class sub-dock (fallback: show all windows).
	case icon.IsMultiAppli() &&
		current.Taskbar.PresentClassOnClick() && // if we want to use this feature
		(!current.Docks.ShowSubDockOnClick() || // if sub-docks are shown on mouse over
			icon.SubDockIsVisible() && // or this sub-dock is already visible
				icon.DesktopPresentClass()): // we use the scale plugin if it's possible

		icon.CallbackActionSubWindows((cdglobal.Window).Show)()

		// in case the dock is visible or about to be visible, hide it, as it would confuse the user to have both.
		// cairo_dock_emit_leave_signal (CAIRO_CONTAINER (icon->pSubDock));
		return notif.AnswerIntercept

		// 	// else handle sub-docks showing on click, applis and launchers (not applets).
		// icon pointing to a sub-dock with either "sub-dock activation on click" option enabled,
		// or sub-dock not visible -> open the sub-dock
	case icon.GetSubDock() != nil && (current.Docks.ShowSubDockOnClick() || !icon.SubDockIsVisible()):
		icon.ShowSubdock(container.ToCairoDock())
		return notif.AnswerIntercept

		// icon holding an appli, but not being an applet -> show/hide the window.
	case icon.IsAppli() && !icon.IsApplet():

		// ne marche que si le dock est une fenêtre de type 'dock', sinon il prend le focus.
		if icon.Window().IsActive() && current.Taskbar.MinimizeOnClick() && !icon.Window().IsHidden() && icon.Window().IsOnCurrentDesktop() {
			icon.Window().Minimize()
		} else {
			icon.Window().Show()
		}
		return notif.AnswerIntercept

		// icon holding a class sub-dock -> show/hide the windows of the class.
	case icon.IsMultiAppli():
		if current.Docks.ShowSubDockOnClick() {
			hideShowInClassSubdock(icon)
		}
		return notif.AnswerIntercept

		// finally, launcher being none of the previous cases -> launch the command
	case icon.IsLauncher():
		// 			if (! gldi_class_is_starting (icon->cClass) && ! gldi_icon_is_launching (icon))  {// do not launch it twice (avoid wrong double click) => if we want to launch it 2 times in a row, we have to use Shift + Click
		icon.LaunchCommand(log)

		return notif.AnswerIntercept // wasn't there in real dock.
	}

	// for applets and their sub-icons, let the module-instance handles the click; for separators, no action.
	// 			cd_debug ("no action here");
	return notif.AnswerLetPass
}

// OnMiddleClick triggers a dock liddle click action on the icon.
//
func OnMiddleClick(icon gldi.Icon, container *gldi.Container) bool {
	if icon == nil || !gldi.ObjectIsDock(container) {
		log.Debug("notifMiddleClick", "ignored: no icon or dock target")
		return notif.AnswerLetPass
	}

	actmid := current.Taskbar.ActionOnMiddleClick()
	simple := icon.IsAppli() && !icon.IsApplet()
	multi := icon.IsMultiAppli()

	switch {
	case (simple || multi) && actmid == 3: // Launch new.
		if icon.GetCommand() != "" {
			// 	if (! gldi_class_is_starting (icon->cClass) && ! gldi_icon_is_launching (icon))  // do not launch it twice
			icon.LaunchCommand(log)
		}
		return notif.AnswerIntercept

	case simple && actmid == 1: // Close one.
		icon.Window().Close()
		return notif.AnswerIntercept

	case simple && actmid == 2: // Minimise one.
		if !icon.Window().IsHidden() {
			icon.Window().Minimize()
		}
		return notif.AnswerIntercept

	case multi && actmid == 1: // Close all.
		icon.CallbackActionSubWindows((cdglobal.Window).Close)()
		return notif.AnswerIntercept

	case multi && actmid == 2: // Minimise all.
		hideShowInClassSubdock(icon)
		return notif.AnswerIntercept
	}

	return notif.AnswerLetPass
}

// OnMouseScroll triggers a dock mouse scroll action on the icon.
//
func OnMouseScroll(icon gldi.Icon, _ *gldi.Container, scrollUp bool) bool {
	switch {
	case icon == nil:

	case icon.IsSeparator(): // Cycle between desktops.
		log.Debug("SeparatorWheelChangeDesktop", confown.Settings.SeparatorWheelChangeDesktop)
		switch confown.Settings.SeparatorWheelChangeDesktop {
		case confown.SeparatorWheelChangeRange:
			desktops.Cycle(scrollUp, false)

		case confown.SeparatorWheelChangeLoop:
			desktops.Cycle(scrollUp, true)
		}
	// SeparatorDesktopLoop   bool

	case icon.IsMultiAppli() || icon.IsStackIcon(): // Cycle between subdock applets list.
		showPrevNextInSubdock(icon, scrollUp)

	case icon.IsAppli() && icon.HasClass():
		next := icon.GetPrevNextClassMateIcon(!scrollUp)
		if next != nil {
			next.Window().Show()
		}
	}
	return notif.AnswerLetPass
}

// OnDropData triggers a dock drop data action on the icon.
//
func OnDropData(icon gldi.Icon, container *gldi.Container, data string, order float64) bool {
	if !gldi.ObjectIsDock(container) {
		log.Debug("notifDropData", "ignored: container is not a dock")
		return notif.AnswerLetPass
	}

	receivingDock := container.ToCairoDock()

	switch {
	case strings.HasSuffix(data, ".desktop"): // -> add a new launcher if dropped on or amongst launchers.

		if !current.Taskbar.MixLauncherAppli() && icon.IsAppli() {
			log.Debug("notifDropData", "ignored: desktop file found but maybe bad location in dock (or need mix launchers)")
			return notif.AnswerLetPass
		}

		// drop onto a container icon.
		if order == gldi.IconLastOrder && icon.IsStackIcon() && icon.GetSubDock() != nil {
			// add into the pointed sub-dock.
			receivingDock = icon.GetSubDock()
		}

		// else, still try to consider it a file?

	case icon == nil || order != gldi.IconLastOrder:
		//  dropped between 2 icons -> try to add it (for instance a script).

	case icon.IsStackIcon(): // sub-dock -> propagate to the sub-dock.
		receivingDock = icon.GetSubDock()

		// dropped on an icon

	case icon.IsLauncher() || icon.IsAppli() || icon.IsClassIcon():
		// launcher/appli -> fire the command with this file.
		cmd := icon.GetCommand()
		if cmd == "" {
			log.Debug("notifDropData", "ignored: no command to trigger")
			return notif.AnswerLetPass
		}

		// Some programs doesn't handle URI. Convert it to local path.
		if strings.HasPrefix(data, "file://") {

			// gchar *cPath = g_filename_from_uri (cReceivedData, NULL, NULL);
			// data = bouine(data)
		}

		ok := icon.LaunchCommand(log, data)
		if ok {
			icon.RequestAnimation("blink", 2)
			log.Debug("notifDropData", "opened with", icon.GetName(), "::", data)
		}

		return notif.AnswerIntercept

	default: // skip any other case.
		log.Debug("notifDropData", "ignored: nothing to do with icon type", icon.GetName())
		return notif.AnswerLetPass
	}

	if current.DockIsLocked() || current.Docks.LockAll() {
		log.Debug("notifDropData", "ignored: dock is locked, can't add icon")
		return notif.AnswerLetPass
	}

	// Still here ? Try to add to target dock.

	newicon := gldi.LauncherAddNew(data, receivingDock, order)
	if newicon == nil {
		log.Debug("notifDropData", "add icon failed ::", data)
		return notif.AnswerIntercept
	}

	log.Debug("notifDropData", "icon added:", icon.GetName())
	return notif.AnswerLetPass
}

//
//---------------------------------------------------------[ C CALLBACKS ]--

// activate the previous or next window managed by the icon.
//
func showPrevNextInSubdock(icon gldi.Icon, next bool) {
	apps, found := iconWindowActive(icon)
	switch {
	case len(apps) == 0: // No apps, nothing to do.
		return

	case found < 0: // None selected. Use first.
		found = 0

	case next: // Wheel up, use next.
		found++

	case !next: // Wheel down, use previous.
		found--
	}

	log.Info("showPrevNextInSubdock found", found)

	if 0 <= found && found < len(apps) {
		apps[found].Window().Show()
	}
}

func hideShowInClassSubdock(icon gldi.Icon) {
	apps, found := iconWindowActive(icon)

	log.Debug("hideShowInClassSubdock", "winNum", found, " wins count", len(apps))

	switch {
	case len(apps) == 0:
		return

	case found >= 0: // one of the windows of the appli has the focus -> hide.
		apps[found].Window().Minimize()

	default:

		// Get list of windows by desktop: current and others.
		var onCurrent []gldi.Icon
		var onOthers []gldi.Icon
		for _, app := range apps {
			win := app.Window()
			switch {
			case win == nil:

			case win.IsOnCurrentDesktop():
				onCurrent = append(onCurrent, app)

			default:
				onOthers = append(onOthers, app)
			}
		}

		log.Debug("hideShowInClassSubdock", "wins on current desktop", len(onCurrent), " on others", len(onOthers))
		switch {
		case len(onCurrent) > 0: // show app windows on current desktop, in Z order.
			// need sort by StackOrder

			for _, app := range onCurrent {
				app.Window().Show()
			}
			return

		case len(onOthers) > 0: // no window on the current desktop -> take the first desktop
			// first := onOthers[0].Window()
			// desk := first.NumDesktop()
			// vx := first.ViewPortX()
			// vy := first.ViewPortY()

			// 	for (ic = pZOrderList; ic != NULL; ic = ic->next){
			// 		pIcon = ic->data;
			// 		if (gldi_window_is_on_desktop (pIcon->pAppli, iNumDesktop, iViewPortX, iViewPortY))
			// 			gldi_window_show (pIcon->pAppli);
			// 	}
		}
	}
}

func iconWindowActive(icon gldi.Icon) ([]gldi.Icon, int) {
	found := -1
	if icon.GetSubDock() == nil {
		return nil, found
	}
	apps := icon.GetSubDock().Icons()
	for i, app := range apps {
		if app.Window().IsActive() {
			found = i
		}
	}
	return apps, found
}
