// Package datagldi provides a data source for the config, based on the gldi backend.
package datagldi

import (
	"github.com/bradfitz/iter"

	"github.com/sqp/godock/libs/cdglobal"
	"github.com/sqp/godock/libs/gldi"
	"github.com/sqp/godock/libs/gldi/desktopclass"
	"github.com/sqp/godock/libs/gldi/globals"
	"github.com/sqp/godock/libs/packages"
	"github.com/sqp/godock/libs/ternary"
	"github.com/sqp/godock/libs/text/tran"
	"github.com/sqp/godock/widgets/confbuilder/datatype"
	"github.com/sqp/godock/widgets/gtk/keyfile"

	"errors"
	"path/filepath"
	"strconv"
)

//
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

	case icon.IsLauncher(), icon.IsStackIcon(), icon.IsAppli(), icon.IsClassIcon():
		name := icon.GetClass().Name()
		if name != "" {
			return name, icon.GetFileName() // icon.GetClassInfo(ClassIcon)
		}
		return ternary.String(icon.GetInitialName() != "", icon.GetInitialName(), icon.GetName()), icon.GetFileName()

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

// MoveBeforePrevious swaps the icon position with the previous one.
//
func (icon *IconConf) MoveBeforePrevious() {
	if icon.GetContainer().IsDesklet() {
		return
	}
	prev := icon.GetContainer().ToCairoDock().GetPreviousIcon(&icon.Icon)
	if prev == nil {
		return
	}
	prev.MoveAfterIcon(icon.GetContainer().ToCairoDock(), &icon.Icon)
}

// MoveAfterNext swaps the icon position with the next one.
//
func (icon *IconConf) MoveAfterNext() {
	if icon.GetContainer().IsDesklet() {
		return
	}
	next := icon.GetContainer().ToCairoDock().GetNextIcon(&icon.Icon)
	if next != nil {
		icon.MoveAfterIcon(icon.GetContainer().ToCairoDock(), next)
	}
}

// GetGettextDomain returns the translation domain for the applet.
//
func (icon *IconConf) GetGettextDomain() string {
	mi := icon.ModuleInstance()
	if mi == nil {
		return ""
	}
	return mi.Module().VisitCard().GetGettextDomain()
}

// ConfigGroup is unused on this backend.
//
func (icon *IconConf) ConfigGroup() string {
	return ""
}

// GetClass returns the class defined for the icon, able to get all related
// desktop class informations.
//
func (icon *IconConf) GetClass() datatype.DesktopClasser {
	return icon.Icon.GetClass() // Just recast as common inferface.
}

// OriginalConfigPath gives the full path to the icon original config file.
// This is the default unchanged config file.
//
func (icon *IconConf) OriginalConfigPath() string {
	if !icon.IsApplet() {
		return ""
	}
	vc := icon.ModuleInstance().Module().VisitCard()
	return filepath.Join(vc.GetShareDataDir(), vc.GetConfFileName())
}

//
//--------------------------------------------------------------[ APPLETCONF ]--

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
	txt, color := packages.FormatCategory(int(v.GetCategory()))
	return "<span fgcolor='#" + color + "'>" + tran.Slate(txt) + "</span>"
}

// IsInstalled returns whether the applet is installed or not.
//
func (v *AppletConf) IsInstalled() bool { return true }

// CanUninstall returns whether the applet can be uninstalled or not.
//
func (v *AppletConf) CanUninstall() bool { return false }

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

// Install does nothing. Only here for compatibility with datatype.Appleter
//
func (v *AppletConf) Install(options string) error {
	return nil
}

// Uninstall does nothing. Only here for compatibility with datatype.Appleter
//
func (v *AppletConf) Uninstall() error {
	return nil
}

// Deactivate does nothing. Only here for compatibility with datatype.Appleter
//
func (v *AppletConf) Deactivate() {}

// IconState does nothing. Only here for compatibility with datatype.Appleter
//
func (v *AppletConf) IconState() string {
	return ""
}

// FormatState does nothing. Only here for compatibility with datatype.Appleter
//
func (v *AppletConf) FormatState() string {
	return ""
}

// FormatSize does nothing. Only here for compatibility with datatype.Appleter
//
func (v *AppletConf) FormatSize() string {
	return ""
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

//----------------------------------------------------------[ APPLETDOWNLOAD ]--

// AppletDownload wraps a dock module and an external applet package as Appleter
// for config data source.
//
type AppletDownload struct {
	packages.AppletPackage
	app *gldi.Module
}

// IsActive returns whether there is at least one active instance of the applet or not.
//
func (v *AppletDownload) IsActive() bool {
	return v.app != nil && len(v.app.InstancesList()) > 0
}

// CanAdd returns whether the applet can be activated.
//
func (v *AppletDownload) CanAdd() bool {
	return !v.IsActive()
}

// CanUninstall returns whether the applet can be uninstalled or not.
//
func (v *AppletDownload) CanUninstall() bool {
	return v.Type != packages.TypeInDev && v.Type != packages.TypeLocal && !v.IsActive()
}

// Activate activates the applet.
// TODO: May have to returns the path to the config file of the new instance.
//
func (v *AppletDownload) Activate() string {
	v.app.Activate()
	return ""
}

// Deactivate deactivates the applet.
//
func (v *AppletDownload) Deactivate() {
	v.app.Deactivate()
}

// GetTitle returns the applet readable name.
//
func (v *AppletDownload) GetTitle() string {
	return v.DisplayedName
}

// GetName returns the applet name to use as config key.
//
func (v *AppletDownload) GetName() string {
	return v.FormatName()
}

// GetAuthor returns the applet author.
//
func (v *AppletDownload) GetAuthor() string {
	return v.Author
}

// FormatCategory formats the applet category text.
//
func (v *AppletDownload) FormatCategory() string {
	txt, color := packages.FormatCategory(int(v.Category))
	return "<span fgcolor='#" + color + "'>" + tran.Slate(txt) + "</span>"
}

// GetIconFilePath returns the location of the applet icon.
// TODO: improve.
//
func (v *AppletDownload) GetIconFilePath() string {
	return filepath.Join(v.Path, "icon")
}

// IconState returns the icon location for the state for the applet.
//
func (v *AppletDownload) IconState() string {
	return globals.DirShareData(v.AppletPackage.IconState())
}

// Install downloads and extract an external archive to package dir.
//
func (v *AppletDownload) Install(options string) error {
	// Using the "drop data signal" trick to ask the Dbus applet to work for us.
	// Only way I found for now to interact with it and let it know it will have
	// a new applet to handle. As a bonus, it also activate the applet, which
	// will toggle the activated button with the UpdateModuleState signal.
	url := packages.DistantURL + v.SrvTag + "/" + v.DisplayedName + "/" + v.DisplayedName + ".tar.gz"
	gldi.EmitSignalDropData(globals.Maindock().Container(), url, nil, 0)

	v.app = gldi.ModuleGet(v.DisplayedName)
	if v.app == nil {
		return errors.New("install failed: v.DisplayedName")
	}

	externalUserDir := globals.DirDockData(cdglobal.AppletsDirName)
	v.SetInstalled(externalUserDir)
	return nil

	// return v.AppletPackage.Install(options)
}

// Uninstall removes the external applet package dir.
//
func (v *AppletDownload) Uninstall() error {
	externalUserDir := globals.DirDockData(cdglobal.AppletsDirName)
	e := v.AppletPackage.Uninstall(externalUserDir)
	if e == nil {
		v.app = nil
	}
	return e
}

//--------------------------------------------------------------[ DOCK THEME ]--

// dockTheme wraps an dock theme package as Appleter for config data source.
//
type dockTheme struct {
	packages.AppletPackage
}

func (v *dockTheme) IsActive() bool               { return false }
func (v *dockTheme) CanAdd() bool                 { return false }
func (v *dockTheme) Activate() string             { return "" }
func (v *dockTheme) Deactivate()                  {}
func (v *dockTheme) GetTitle() string             { return v.DisplayedName }
func (v *dockTheme) GetName() string              { return v.FormatName() }
func (v *dockTheme) GetAuthor() string            { return v.Author }
func (v *dockTheme) FormatCategory() string       { return "" }
func (v *dockTheme) GetIconFilePath() string      { return v.IconState() }
func (v *dockTheme) IconState() string            { return globals.DirShareData(v.AppletPackage.IconState()) }
func (v *dockTheme) Install(options string) error { return nil }
func (v *dockTheme) Uninstall() error             { return nil }

func (v *dockTheme) CanUninstall() bool {
	return v.Type != packages.TypeInDev && v.Type != packages.TypeLocal
}

//
//-------------------------------------------------[ HANDBOOK DESC TRANSLATE ]--

// HandbookDescTranslate improves Handbooker with a translated description.
//
type HandbookDescTranslate struct{ datatype.Handbooker }

// GetDescription returns the book description.
//
func (dv *HandbookDescTranslate) GetDescription() string {
	return tran.Slate(dv.Handbooker.GetDescription())
}

//
//-------------------------------------------------------------[ DATA SOURCE ]--

// Data provides a config Source interface based on the dock gldi backend.
//
type Data struct{ datatype.SourceCommon }

//MainConfigFile returns the full path to the dock config file.
//
func (Data) MainConfigFile() string {
	return globals.ConfigFile()
}

//MainConfigDefault returns the full path to the dock config file.
//
func (Data) MainConfigDefault() string {
	return globals.ConfigFileDefault()
}

// AppIcon returns the application icon path.
//
func (Data) AppIcon() string {
	return globals.FileCairoDockIcon()
}

// DirShareData returns the path to the shared data dir.
func (Data) DirShareData(path ...string) string {
	return globals.DirShareData(path...)
}

// DirUserAppData returns the path to the applications data dir (user saved data).
//
func (Data) DirUserAppData(path ...string) (string, error) {
	return globals.DirUserAppData(path...)
}

// ListKnownApplets builds the list of all user applets.
//
func (Data) ListKnownApplets() map[string]datatype.Appleter {
	list := make(map[string]datatype.Appleter)
	for name, app := range gldi.ModuleList() {
		if !app.IsAutoLoaded() { // don't display modules that can't be disabled
			list[name] = &AppletConf{
				VisitCard: *app.VisitCard(),
				app:       app,
			}
		}
	}
	return list
}

// ListDownloadApplets builds the list of downloadable user applets (installed or not).
//
func (Data) ListDownloadApplets() (map[string]datatype.Appleter, error) {
	externalUserDir := globals.DirDockData(cdglobal.AppletsDirName)
	packs, e := packages.ListDownloadApplets(externalUserDir)
	if e != nil {
		return nil, e
	}

	applets := gldi.ModuleList()
	list := make(map[string]datatype.Appleter)
	for k, v := range packs {
		list[k] = &AppletDownload{
			AppletPackage: *v,
			app:           applets[k],
		}
	}

	return list, nil
}

// ListIcons builds the list of all icons.
//
func (Data) ListIcons() *datatype.ListIcon {
	list := datatype.NewListIcon()

	// Add icons in docks.
	taskbarSet := false
	for _, dock := range gldi.GetAllAvailableDocks(nil, nil) {
		// for _, dock := range gldi.ListDocksRoot() {

		var found []datatype.Iconer
		for _, icon := range dock.Icons() {
			if dock.GetRefCount() == 0 { // Maindocks.

				// Group taskbar icons and separators.
				if icon.ConfigPath() == "" || icon.IsSeparatorAuto() {
					// if icon.IsTaskbar() || icon.IsSeparatorAuto() {
					if !taskbarSet {
						taskbarSet = true
						ic := datatype.NewIconSimple(
							globals.ConfigFile(),
							datatype.FieldTaskBar,
							datatype.TitleTaskBar,
							globals.DirShareData("icons/icon-taskbar.png"))

						ic.Taskbar = true

						found = append(found, ic)
					}

				} else {
					found = append(found, &IconConf{*icon})
				}

			} else { // Subdock.
				parentName := icon.GetParentDockName()
				list.Subdocks[parentName] = append(list.Subdocks[parentName], &IconConf{*icon})
			}
		}

		if len(found) > 0 {
			var file, group string

			if dock.IsMainDock() {
				// Only maindocks after the first one have a config file.
				// So the first maindock use a custom group.
				group = datatype.KeyMainDock

			} else {
				// Other maindocks have a dedicated config file.
				// So the group is empty to load all of them (auto find).
				file = globals.CurrentThemePath(dock.GetDockName() + ".conf")
			}
			container := datatype.NewIconSimple(
				file,
				group,
				dock.GetReadableName(),
				"") // TODO: maybe get an icon for the maindock.

			list.Add(container, found)
		}
	}

	// Add modules in desklets.
	var desklets []datatype.Iconer
	for _, desklet := range gldi.DeskletList() {
		icon := desklet.GetIcon()
		if icon != nil {
			desklets = append(desklets, &IconConf{*icon})
		}
	}

	if len(desklets) > 0 {
		container := datatype.NewIconSimple(
			globals.ConfigFile(),
			datatype.GroupDesklets,
			datatype.TitleDesklets,
			"") // TODO: maybe get an icon for the desklets group.

		list.Add(container, desklets)
	}

	// Add other modules (not in a dock or a desklet) : plug-in or detached applet.
	// We need to create custom icons for them.
	var services []datatype.Iconer
	for _, mod := range gldi.ModuleList() {
		cat := mod.VisitCard().GetCategory()

		if cat != gldi.CategoryBehavior && cat != gldi.CategoryTheme && !mod.IsAutoLoaded() {
			for _, mi := range mod.InstancesList() {

				if mi.Icon() == nil || (mi.Dock() != nil && mi.Icon().GetContainer() == nil) {
					icon := datatype.NewIconSimple(
						mi.GetConfFilePath(),
						"", // no group, we need all of them for an applet.
						mod.VisitCard().GetTitle(),
						mod.VisitCard().GetIconFilePath())

					services = append(services, icon)
				}
			}
		}
	}
	if len(services) > 0 {
		container := datatype.NewIconSimple(
			"", // no config file available.
			datatype.GroupServices, // so we set a custom group.
			tran.Slate(datatype.TitleServices),
			"") // TODO: maybe get an icon for the services group.

		list.Add(container, services)
	}

	return list
}

// ListShortkeys returns the list of dock shortkeys.
//
func (Data) ListShortkeys() (list []datatype.Shortkeyer) {
	for _, rend := range gldi.ShortkeyList() {
		list = append(list, datatype.Shortkeyer(rend))
	}
	return list
}

// ListScreens returns the list of screens.
//
func (Data) ListScreens() (list []datatype.Field) {
	geo := gldi.GetDesktopGeometry()
	nb := geo.NbScreens()
	if nb <= 1 {
		return []datatype.Field{{Key: "0", Name: "Use all screens"}}
	}

	var xmax, ymax int
	for i := range iter.N(nb) {
		x, y := geo.ScreenPosition(i)
		xmax = ternary.Max(x, xmax)
		ymax = ternary.Max(y, ymax)
	}

	for i := range iter.N(nb) {
		var xstr, ystr string
		x, y := geo.ScreenPosition(i)
		if xmax > 0 { // at least 2 screens horizontally
			switch {
			case x == 0:
				xstr = "left"
			case x == xmax:
				xstr = "right"
			default:
				xstr = "middle"
			}
		}

		if ymax > 0 { // at least 2 screens vertically
			switch {
			case y == 0:
				ystr = "top"
			case y == ymax:
				ystr = "bottom"
			default:
				ystr = "middle"
			}
		}
		key := strconv.Itoa(i)
		sep := ternary.String(xstr != "" && ystr != "", " - ", "")
		name := "Screen" + " " + key + " (" + xstr + sep + ystr + ")"

		list = append(list, datatype.Field{Key: key, Name: name})
	}
	return list
}

// ListViews returns the list of views.
//
func (Data) ListViews() map[string]datatype.Handbooker {
	list := make(map[string]datatype.Handbooker)
	for key, rend := range gldi.CairoDockRendererList() {
		list[key] = &HandbookDescTranslate{&datatype.HandbookDescDisk{&datatype.HandbookSimple{
			Key:         key,
			Title:       ternary.String(rend.GetDisplayedName() != "", rend.GetDisplayedName(), key),
			Description: rend.GetReadmeFilePath(),
			Preview:     rend.GetPreviewFilePath(),
		}}}
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
		if dock.GetRefCount() == 0 { // Any maindock.
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

// ListDockThemeLoad builds the list of dock themes for load widget (local and distant).
//
func (Data) ListDockThemeLoad() (map[string]datatype.Appleter, error) {
	dir := globals.DirDockData(cdglobal.ConfigDirDockThemes)
	packs, e := packages.ListDownloadDockThemes(dir)
	if e != nil {
		return nil, e
	}

	list := make(map[string]datatype.Appleter)
	for _, v := range packs {
		list[v.DisplayedName] = &dockTheme{AppletPackage: *v}
	}

	return list, nil
}

// ListDockThemeSave builds the list of dock themes for save widget (local).
//
func (Data) ListDockThemeSave() []datatype.Field {
	dir := globals.DirDockData(cdglobal.ConfigDirDockThemes)
	packs, e := packages.ListFromDir(dir, packages.TypeUser, packages.SourceDockTheme)
	if e != nil {
		println("ListDockThemeSave wrong dir:", dir) // TODO: use a logger.
		return nil
	}

	var list []datatype.Field
	for _, pack := range packs {
		list = append(list, datatype.Field{
			Key:  pack.DisplayedName,
			Name: pack.GetName(),
		})
	}
	return list
}

// CurrentThemeLoad imports and loads a dock theme.
//
func (Data) CurrentThemeLoad(themeName string, useBehaviour, useLaunchers bool) error {

	// if (pThemesWidget->pImportTask != NULL)
	// {
	// 	gldi_task_discard (pThemesWidget->pImportTask);
	// 	pThemesWidget->pImportTask = NULL;
	// }
	// //\___________________ On regarde si le theme courant est modifie.
	// gboolean bNeedSave = cairo_dock_current_theme_need_save ();
	// if (bNeedSave)
	// {

	if gldi.CurrentThemeNeedSave() {
		// 	Icon *pIcon = cairo_dock_get_current_active_icon ();  // it's most probably the icon corresponding to the configuration window
		// 	if (pIcon == NULL || cairo_dock_get_icon_container (pIcon) == NULL)  // if not available, get any icon
		// 		pIcon = gldi_icons_get_any_without_dialog ();
		// 	int iClickedButton = gldi_dialog_show_and_wait (_("You have made some changes to the current theme.\nYou will lose them if you don't save before choosing a new theme. Continue anyway?"),
		// 		pIcon, CAIRO_CONTAINER (g_pMainDock),
		// 		CAIRO_DOCK_SHARE_DATA_DIR"/"CAIRO_DOCK_ICON, NULL);
		// 	if (iClickedButton != 0 && iClickedButton != -1)  // cancel button or Escape.
		// 	{
		// 		return FALSE;
		// 	}
	}

	// //\___________________ On charge le nouveau theme choisi.
	// gchar *tmp = g_strdup (cNewThemeName);
	// CairoDockPackageType iType = cairo_dock_extract_package_type_from_name (tmp);
	// g_free (tmp);

	// gboolean bThemeImported = FALSE;
	// if (iType != CAIRO_DOCK_LOCAL_PACKAGE && iType != CAIRO_DOCK_USER_PACKAGE)
	// {
	// 	GtkWidget *pWaitingDialog = gtk_window_new (GTK_WINDOW_TOPLEVEL);
	// 	pThemesWidget->pWaitingDialog = pWaitingDialog;
	// 	gtk_window_set_decorated (GTK_WINDOW (pWaitingDialog), FALSE);
	// 	gtk_window_set_skip_taskbar_hint (GTK_WINDOW (pWaitingDialog), TRUE);
	// 	gtk_window_set_skip_pager_hint (GTK_WINDOW (pWaitingDialog), TRUE);
	// 	gtk_window_set_transient_for (GTK_WINDOW (pWaitingDialog), pMainWindow);
	// 	gtk_window_set_modal (GTK_WINDOW (pWaitingDialog), TRUE);

	// 	GtkWidget *pMainVBox = gtk_box_new (GTK_ORIENTATION_VERTICAL, CAIRO_DOCK_FRAME_MARGIN);
	// 	gtk_container_add (GTK_CONTAINER (pWaitingDialog), pMainVBox);

	// 	GtkWidget *pLabel = gtk_label_new (_("Please wait while importing the theme..."));
	// 	gtk_box_pack_start(GTK_BOX (pMainVBox), pLabel, FALSE, FALSE, 0);

	// 	GtkWidget *pBar = gtk_progress_bar_new ();
	// 	gtk_progress_bar_pulse (GTK_PROGRESS_BAR (pBar));
	// 	gtk_box_pack_start (GTK_BOX (pMainVBox), pBar, FALSE, FALSE, 0);
	// 	pThemesWidget->iSidPulse = g_timeout_add (100, (GSourceFunc)_pulse_bar, pBar);
	// 	g_signal_connect (G_OBJECT (pWaitingDialog),
	// 		"destroy",
	// 		G_CALLBACK (on_waiting_dialog_destroyed),
	// 		pThemesWidget);

	// 	GtkWidget *pCancelButton = gtk_button_new_with_label (_("Cancel"));
	// 	g_signal_connect (G_OBJECT (pCancelButton), "clicked", G_CALLBACK(on_cancel_dl), pWaitingDialog);
	// 	gtk_box_pack_start (GTK_BOX (pMainVBox), pCancelButton, FALSE, FALSE, 0);

	// 	gtk_widget_show_all (pWaitingDialog);

	// 	cd_debug ("start importation...");
	// 	pThemesWidget->pImportTask = cairo_dock_import_theme_async (cNewThemeName, bLoadBehavior, bLoadLaunchers, (GFunc)_load_theme, pThemesWidget);  // if 'pThemesWidget' is destroyed, the 'reset' callback will be called and will cancel the task.
	// }
	// else  // if the theme is already local and uptodate, there is really no need to show a progressbar, because only the download/unpacking is done asynchonously (and the copy of the files is fast enough).
	// {

	e := gldi.CurrentThemeImport(themeName, useBehaviour, useLaunchers)
	if e != nil {
		return e
	}
	gldi.CurrentThemeLoad()

	return nil
}

// CurrentThemeSave saves the current dock theme, and can also make an archive.
//
func (Data) CurrentThemeSave(themeName string, saveBehaviour, saveLaunchers, needPackage bool, dirPackage string) error {
	e := gldi.CurrentThemeExport(themeName, saveBehaviour, saveLaunchers)
	if e != nil {
		return e
	}
	if !needPackage {
		return nil
	}
	return gldi.CurrentThemePackage(themeName, dirPackage)
}

// Handbook wraps a dock module visit card as Handbooker for config data source.
//
func (Data) Handbook(name string) datatype.Handbooker {
	mod := gldi.ModuleGet(name)
	if mod == nil {
		return nil
	}
	return &datatype.HandbookDescSplit{mod.VisitCard()}
}

// ManagerReload reloads the manager matching the given name.
//
func (Data) ManagerReload(name string, b bool, keyf *keyfile.KeyFile) {
	gldi.ManagerReload(name, b, keyf)
}

// CreateMainDock creates a new main dock to store a moved icon.
//
func (Data) CreateMainDock() string {
	return gldi.DockAddConfFile()
}

// DesktopClasser allows to get desktop class informations for a given name.
//
func (Data) DesktopClasser(class string) datatype.DesktopClasser {
	return desktopclass.Info(class)
}
