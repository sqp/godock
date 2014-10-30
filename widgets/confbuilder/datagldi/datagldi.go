// Package datagldi provides a data source for the config, based on the gldi backend.
package datagldi

import (
	"github.com/sqp/godock/libs/gldi"
	"github.com/sqp/godock/libs/gldi/globals"
	// "github.com/sqp/godock/libs/maindock"
	"github.com/sqp/godock/widgets/confbuilder/datatype"
	"github.com/sqp/godock/widgets/gtk/keyfile"

	// "path/filepath"
)

//--------------------------------------------------------[ ICONER GLDI ICON ]--

// IconConf wraps a dock icon as Iconer for config data source.
//
type IconConf struct {
	gldi.Icon
}

// DefaultNameIcon returns improved name and image for the icon if possible.
//
func (icon *IconConf) DefaultNameIcon() (name, img string) {
	switch {
	case icon.IsApplet():
		vc := icon.ModuleInstance().Module().VisitCard()
		return vc.GetTitle(), vc.GetIconFilePath()

	case icon.IsSeparator():
		return "--------", ""

	case icon.IsLauncher():
		name := icon.GetClassInfo(gldi.ClassName)
		if name != "" {
			return name, icon.GetFileName() // icon.GetClassInfo(ClassIcon)
		}

	}
	return icon.GetName(), icon.GetFileName()
}

// Reload reloads the icon to apply the new configuration.
//
func (icon *IconConf) Reload() {
	// container := icon.GetContainer()
	switch {
	case icon.IsApplet():
		gldi.ObjectReload(icon.Icon.ModuleInstance())

		// case container != nil && gldi.ObjectIsDock(container):
		// 	dock := container.ToCairoDock()

		// 	// if !dock.IsMainDock() { // pour l'instant le main dock n'a pas de fichier de conf

		// 	path := filepath.Join(maindock.CurrentThemePath(), dock.GetDockName())

		// 	println(dock.IsMainDock(), path, dock.GetDockName()) //, container.IsDock()) //gldi.OjectIsManagerChild(container))

		// 	// 			// reload dock's config.
		// 	gldi.ObjectReload(dock)
		// 	// }

	default: // else if (pIcon)

		// prend tout en compte, y compris le redessin et declenche le rechargement de l'IHM.
		gldi.ObjectReload(&icon.Icon)

	}
}

// AppletConf wraps a dock module and visitcard as Appleter for config data source.
//
type AppletConf struct {
	gldi.VisitCard
	app *gldi.Module
}

// func (v *AppletConf) DefaultNameIcon() (string, string) { return v.GetTitle(), v.GetIconFilePath() }

// FormatCategory formats the applet category text.
//
func (v *AppletConf) FormatCategory() string {
	return datatype.FormatCategory(int(v.GetCategory()))
}

// IsInstalled returns whether the applet is installed or not.
//
func (v *AppletConf) IsInstalled() bool { return true }

// IsActive returns whether there is at least one active instance of the applet or not.
//
func (v *AppletConf) IsActive() bool { return len(v.app.InstancesList()) > 0 }

// CanAdd returns whether the applet can be activated (again).
//
func (v *AppletConf) CanAdd() bool {
	return (v.IsMultiInstance() || !v.IsActive()) && v.GetCategory() != gldi.CategoryTheme // don't display the animations plug-ins
}

// Activate activates the applet and returns the path to the config file of the new instance.
//
func (v *AppletConf) Activate() string {
	if globals.PrimaryContainer().Ptr == nil {
		println("NIL CONTAINER")
	}

	// 	GldiModule *pModule = gldi_module_get (cModuleName);
	// 	if (g_pPrimaryContainer == NULL)
	// 	{
	// 		cairo_dock_add_remove_element_to_key (g_cConfFile, "System", "modules", cModuleName, bState);
	// 	}
	// 	else if (pModule->pInstancesList == NULL)
	// 	{

	instances := v.app.InstancesList() // Save current instances.

	if len(v.app.InstancesList()) == 0 {
		v.app.Activate()
	} else {
		v.app.AddInstance()
	}

	return findNewInstance(instances, v.app.InstancesList())
}

func findNewInstance(listold, listnew []*gldi.ModuleInstance) string {
	for _, appnew := range listnew {
		found := false
		path := appnew.GetConfFilePath()
		for _, appold := range listold {
			if path == appold.GetConfFilePath() {
				found = true
			}
		}

		if !found {
			return path
		}
	}
	return ""
}

//
//-------------------------------------------------------------[ DATA SOURCE ]--

// Data provides a config Source interface based on the dock gldi backend.
//
type Data struct{ datatype.SourceCommon }

//MainConf returns the full path to the dock config file.
//
func (Data) MainConf() string {
	return globals.ConfigFile()
}

// DirShareData returns the path to the shared data dir.
func (Data) DirShareData() string {
	return globals.DirShareData()
}

// DirAppData returns the path to the applications data dir (user saved data).
//
func (Data) DirAppData() (string, error) {
	return globals.DirAppdata()
}

// ListApplets builds the list of all user applets.
//
func (Data) ListApplets() map[string]datatype.Appleter {
	list := make(map[string]datatype.Appleter)
	for name, app := range gldi.ModuleList() {
		if !app.IsAutoLoaded() { // don't display modules that can't be disabled
			list[name] = &AppletConf{*app.VisitCard(), app}
		}
	}
	return list
}

// ListIcons builds the list of all icons.
//
func (Data) ListIcons() map[string][]datatype.Iconer {
	list := make(map[string][]datatype.Iconer)
	icons := globals.Maindock().Icons()
	taskbar := false
	for _, icon := range icons {
		parent := icon.GetParentDockName()

		// Group taskbar icons and separators.
		if icon.IsTaskbar() || icon.IsSeparatorAuto() {
			if !taskbar {
				ic := &datatype.IconSeparator{
					Field:   datatype.Field{Key: globals.ConfigFile(), Name: "--[ Taskbar ]--"},
					Taskbar: true}

				list[parent] = append(list[parent], ic)
				taskbar = true
			}
			continue
		}

		list[parent] = append(list[parent], &IconConf{*icon})
	}
	return list
}

// ListViews returns the list of views.
//
func (Data) ListViews() (list []datatype.Field) {
	for key, rend := range gldi.CairoDockRendererList() {
		list = append(list, displayerField(key, rend))
	}
	return list
}

// ListAnimations returns the list of animations.
//
func (Data) ListAnimations() (list []datatype.Field) {
	for key, rend := range gldi.AnimationList() {
		list = append(list, displayerField(key, rend))
	}
	return list
}

// ListDeskletDecorations returns the list of desklet decorations.
//
func (Data) ListDeskletDecorations() (list []datatype.Field) {
	for key, rend := range gldi.CairoDeskletDecorationList() {
		list = append(list, displayerField(key, rend))
	}
	return list
}

// ListDialogDecorator returns the list of dialog decorators.
//
func (Data) ListDialogDecorator() (list []datatype.Field) {
	for key, rend := range gldi.DialogDecoratorList() {
		list = append(list, displayerField(key, rend))
	}
	return list
}

type displayer interface {
	GetDisplayedName() string
}

func displayerField(key string, data displayer) datatype.Field {
	name := data.GetDisplayedName()
	if name == "" {
		name = key
	}
	return datatype.Field{Key: key, Name: name}
}

// ListDocks builds the list of docks with readable name.
// Both options are docks to remove from the list. Subdock childrens are removed too.
//
func (Data) ListDocks(parent, subdock string) []datatype.Field {
	var list []datatype.Field

	// count := make([]int, 4)
	// for _, dock := range appdbus.DockProperties("type=Dock") {

	sub := gldi.DockGet(subdock)

	for _, dock := range gldi.GetAllAvailableDocks(nil, sub) { // nil because we want the current parent dock in the list.
		field := datatype.Field{Key: dock.GetDockName()}
		if dock.IsMainDock() {
			field.Name = dock.GetReadableName()
		} else {
			field.Name = dock.GetDockName()
		}
		list = append(list, field)

		// orientation := dock["orientation"].Value().(uint32)
		// count[orientation]++ // Count the number of docks on each position.

		// name := dock["name"].Value().(string)
		// if name != parent && name != subdock {
		// 	text := ""
		// 	switch orientation {
		// 	case 0:
		// 		text = "Bottom dock"
		// 	case 1:
		// 		text = "Top dock"
		// 	case 2:
		// 		text = "Right dock"
		// 	case 3:
		// 		text = "Left dock"
		// 	}

		// 	if count[orientation] > 1 {
		// 		text = fmt.Sprintf("%s (%d)", text, count[orientation])
		// 	}

		// 	list = append(list, datatype.Field{Key: name, Name: text})
		// }
	}

	// for _, dock := range appdbus.DockProperties("type=Stack-icon") {
	// 	name := dock["name"].Value().(string)

	// 	if name != parent && name != subdock {
	// 		list = append(list, datatype.Field{Key: name, Name: name})
	// 	}
	// }
	return list
}

// AppletPackage returns the package of the requested applet.
//
// func (Data) AppletPackage(appletName string) *packages.AppletPackage {
// 	// return appdbus.InfoApplet(appletName)
// 	return nil
// }

// ListIconsMainDock builds the list of icons in the maindock.
//
func (Data) ListIconsMainDock() (list []datatype.Iconer) {
	for _, icon := range globals.Maindock().Icons() {
		if !icon.IsTaskbar() && !icon.IsSeparatorAuto() && icon.GetParentDockName() == datatype.KeyMainDock {
			list = append(list, &IconConf{*icon})
		}
	}
	return list
}

// Handbook wraps a dock module visit card as Handbooker for config data source.
//
func (Data) Handbook(name string) datatype.Handbooker {
	mod := gldi.ModuleGet(name)
	if mod == nil {
		return nil
	}
	return mod.VisitCard()
}

// ManagerReload reloads the manager matching the given name.
//
func (Data) ManagerReload(name string, b bool, keyf keyfile.KeyFile) {
	gldi.ManagerReload(name, b, keyf)
}

//

//

//

// case CAIRO_DOCK_WIDGET_SCREENS_LIST :
// {
// 	GHashTable *pHashTable = _cairo_dock_build_screens_list ();

// 	GtkListStore *pScreensListStore = _cairo_dock_build_screens_list_for_gui (pHashTable);

// 	_add_combo_from_modele (pScreensListStore, FALSE, FALSE, FALSE);

// 	g_object_unref (pScreensListStore);
// 	g_hash_table_destroy (pHashTable);

// 	gldi_object_register_notification (&myDesktopMgr,
// 		NOTIFICATION_DESKTOP_GEOMETRY_CHANGED,
// 		(GldiNotificationFunc) _on_screen_modified,
// 		GLDI_RUN_AFTER, pScreensListStore);
// 	g_signal_connect (pOneWidget, "destroy", G_CALLBACK (_on_list_destroyed), NULL);

// 	if (g_desktopGeometry.iNbScreens <= 1)
// 		gtk_widget_set_sensitive (pOneWidget, FALSE);
// }
