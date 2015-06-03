// Package datagldi provides a data source for the config, based on the gldi backend.
package datagldi

import (
	"github.com/bradfitz/iter"

	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/gldi"
	"github.com/sqp/godock/libs/gldi/globals"
	"github.com/sqp/godock/libs/packages"
	"github.com/sqp/godock/libs/ternary"
	"github.com/sqp/godock/libs/tran"
	// "github.com/sqp/godock/libs/maindock"
	"github.com/sqp/godock/widgets/confbuilder/datatype"
	"github.com/sqp/godock/widgets/gtk/keyfile"

	"errors"
	"path/filepath"
	"strconv"
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

	case icon.IsLauncher(), icon.IsStackIcon(), icon.IsAppli(), icon.IsClassIcon():
		name := icon.GetClassInfo(gldi.ClassName)
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

// MoveAfterNext swaps the icon position with the previous one.
//
func (icon *IconConf) MoveBeforePrevious() {
	prev := icon.GetContainer().ToCairoDock().GetPreviousIcon(&icon.Icon)
	if prev == nil {
		return
	}
	prev.MoveAfterIcon(icon.GetContainer().ToCairoDock(), &icon.Icon)
}

// MoveAfterNext swaps the icon position with the next one.
//
func (icon *IconConf) MoveAfterNext() {
	next := icon.GetContainer().ToCairoDock().GetNextIcon(&icon.Icon)
	if next != nil {
		icon.MoveAfterIcon(icon.GetContainer().ToCairoDock(), next)
	}
}

// GetGettextDomain returns the translation domain for the applet.
//
func (v *IconConf) GetGettextDomain() string {
	mi := v.ModuleInstance()
	if mi == nil {
		return ""
	}
	return mi.Module().VisitCard().GetGettextDomain()
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

// AppletDownload wraps a dock module and visitcard as Appleter for config data source.
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
	return v.Type != packages.TypeInDev && v.Type != packages.TypeLocal
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
	url := packages.DistantURL + cdtype.AppletsDirName + "/" + v.SrvTag + "/" + v.DisplayedName + "/" + v.DisplayedName + ".tar.gz"
	gldi.EmitSignalDropData(globals.Maindock().Container(), url, nil, 0)

	v.app = gldi.ModuleGet(v.DisplayedName)
	if v.app == nil {
		return errors.New("install failed: v.DisplayedName")
	}

	dir, _ := packages.DirExternal()
	v.SetInstalled(dir)
	return nil

	// return v.AppletPackage.Install(options)
}

// Uninstall downloads and extract an external archive to package dir.
//
func (v *AppletDownload) Uninstall() error {
	e := v.AppletPackage.Uninstall()
	if e == nil {
		v.app = nil
	}
	return e
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

// ListKnownApplets builds the list of all user applets.
//
func (Data) ListKnownApplets() map[string]datatype.Appleter {
	list := make(map[string]datatype.Appleter)
	for name, app := range gldi.ModuleList() {
		if !app.IsAutoLoaded() { // don't display modules that can't be disabled
			list[name] = &AppletConf{*app.VisitCard(), app}
		}
	}
	return list
}

// ListDownloadApplets builds the list of downloadable user applets (installed or not).
//
func (Data) ListDownloadApplets() map[string]datatype.Appleter {
	packs, e := packages.ListDownloadIndex(cdtype.AppletsServerTag)
	// if log.Err(e, "ListExternal applets") {
	// 	return nil
	// }
	_ = e

	applets := gldi.ModuleList()
	list := make(map[string]datatype.Appleter)
	for k, v := range packs {
		list[k] = &AppletDownload{*v, applets[k]}
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
func (Data) ManagerReload(name string, b bool, keyf *keyfile.KeyFile) {
	gldi.ManagerReload(name, b, keyf)
}
