// Package datatype defines the data source format for the config.
package datatype

import (
	"github.com/sqp/godock/libs/packages"
	"github.com/sqp/godock/widgets/gtk/keyfile"

	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
)

// KeyMainDock is the key name of the first main dock (the one with the taskbar).
//
const (
	KeyMainDock    = "_MainDock_"
	DirIconsSystem = "/usr/share/icons"
	DirIconsUser   = ".icons" // in $HOME
)

// Source defines external data needed by the config builder.
//
type Source interface {
	//MainConf returns the full path to the dock config file.
	//
	MainConf() string

	DirAppData() (string, error)

	DirShareData() string

	// ListIcons builds the list of all icons.
	//
	ListIcons() map[string][]Iconer

	// ListApplets builds the list of all applets.
	//
	ListApplets() map[string]Appleter

	// ListIconsMainDock builds the list of icons in the maindock.
	//
	ListIconsMainDock() []Iconer

	// ListViews returns the list of views.
	//
	ListViews() []Field

	// ListAnimations returns the list of animations.
	//
	ListAnimations() []Field

	// ListDeskletDecorations returns the list of desklet decorations.
	//
	ListDeskletDecorations() []Field

	// ListDialogDecorator returns the list of dialog decorators.
	//
	ListDialogDecorator() []Field

	// ListDocks builds the list of docks with a readable name.
	// Both options are docks to remove from the list. Subdock childrens are removed too.
	//
	ListDocks(parent, subdock string) []Field

	// ListIconTheme builds a list of desktop icon-themes in system and user dir.
	//
	ListIconTheme() []Field

	Handbook(appletName string) Handbooker

	// ListThemeXML builds a list of icon theme in system and user dir.
	//
	ListThemeXML(localSystem, localUser, distant string) []packages.Theme

	// ListThemeINI builds a list of icon theme in system and user dir.
	//
	ListThemeINI(localSystem, localUser, distant string) packages.AppletPackages

	// ManagerReload reloads the manager matching the given name.
	//
	ManagerReload(name string, b bool, keyf keyfile.KeyFile)
}

// SourceCommon provides common methods for dock config data source.
//
type SourceCommon struct{}

// ListThemeXML builds a list of icon theme in system and user dir.
//
func (SourceCommon) ListThemeXML(localSystem, localUser, distant string) []packages.Theme {
	// list, _ := packages.ListExternalUser(localSystem, "theme")
	list, _ := packages.ListThemesDir(localSystem, packages.TypeLocal)

	if userDir, e := packages.DirTheme(localUser); e == nil {
		users, _ := packages.ListThemesDir(userDir, packages.TypeUser)

		list = append(list, users...)

	}

	// Rename theme title with the online list.
	// TODO: maybe need to use hint here.
	dist, _ := packages.ListDistant(distant)
	for k, v := range list {
		if more := dist.Get(v.DirName); more != nil && more.Title != "" {
			list[k].Title = more.Title
		}
	}

	// TODO: Distant theme management will have to be moved into the download area.
	// dist, _ := packages.ListDistant(distant)
	// for _, v := range dist {
	// 	log.DEV("", v)
	// }

	return list
}

// ListThemeINI builds a list of icon theme in system and user dir.
//
func (SourceCommon) ListThemeINI(localSystem, localUser, distant string) packages.AppletPackages {
	// Themes installed in system dir.
	list, _ := packages.ListFromDir(localSystem, packages.TypeLocal, packages.SourceTheme)

	// Themes installed in user dir.
	if userDir, e := packages.DirTheme(localUser); e == nil {
		dist, _ := packages.ListFromDir(userDir, packages.TypeUser, packages.SourceTheme)
		list = append(list, dist...)
	}
	return list
}

// ListIconTheme builds a list of desktop icon-themes in system and user dir.
//
func (SourceCommon) ListIconTheme() []Field {

	dirs := []string{DirIconsSystem}
	usr, e := user.Current()
	if e == nil {
		dirs = append([]string{filepath.Join(usr.HomeDir, DirIconsUser)}, dirs...) // prepend ~/.icons
	}

	var list []Field
	for _, dir := range dirs {

		files, e := ioutil.ReadDir(dir) // Get all files in the given directories.
		if e != nil {
			continue
		}

		for _, info := range files {
			fullpath := filepath.Join(dir, info.Name()) // and only keep dirs.
			if !info.IsDir() {
				continue
			}

			file := filepath.Join(fullpath, "index.theme") // Check if a theme index file exists.
			if _, e = os.Stat(file); e != nil {
				continue
			}

			kf := keyfile.New()
			ok, _ := kf.LoadFromFile(file, keyfile.FlagsNone) // Keyfile required.
			if !ok {
				continue
			}

			hidden, _ := kf.GetBoolean("Icon Theme", "Hidden")
			hasdirs := kf.HasKey("Icon Theme", "Directories")
			name, _ := kf.GetString("Icon Theme", "Name")
			if hidden || !hasdirs || name == "" { // Check theme settings.
				continue
			}

			list = append(list, Field{Key: info.Name(), Name: name})
		}
	}
	return list
}

//
//--------------------------------------------------------[ APPLET INTERFACE ]--

// Appleter defines the interface needed by applets provided as config source.
//
type Appleter interface {
	// DefaultNameIcon() (string, string)

	// Icon() string
	IsInstalled() bool
	IsActive() bool
	Activate() string
	CanAdd() bool

	GetTitle() string // module name translated for the user.
	GetName() string  // module name used as key.
	GetAuthor() string
	GetDescription() string
	GetPreviewFilePath() string
	GetIconFilePath() string
	FormatCategory() string
}

// FormatCategory returns the human readable category for the applet.
//
func FormatCategory(cat int) string {
	switch cat {
	case 0:
		return "Behavior"
	case 2:
		return "<span fgcolor='#004EA1'>Files</span>" // blue.
	case 3:
		return "<span fgcolor='#FF5555'>Internet</span>" // orange.
	case 4:
		return "<span fgcolor='#116E08'>Desktop</span>" // green.
	case 5:
		return "<span fgcolor='#900009'>Accessory</span>" // red.
	case 6:
		return "<span fgcolor='#A58B0D'>System</span>" // yellow.
	case 7:
		return "<span fgcolor='#FF55FF'>Fun</span>" // pink.
	}
	return ""
}

//
//----------------------------------------------------------[ ICON INTERFACE ]--

// Iconer defines the interface needed by icons provided as config source.
//
type Iconer interface {
	ConfigPath() string
	DefaultNameIcon() (string, string) //applets map[string]*packages.AppletPackage) (string, string)
	IsTaskbar() bool
	IsLauncher() bool
	GetClassInfo(int) string
	GetCommand() string
	Reload()
}

/* An icon can either be:
* - a launcher (it has a command, a class, and possible an X window ID)
* - an appli (it has a X window ID and a class, no command)
* - an applet (it has a module instance and no command, possibly a class)
* - a container (it has a sub-dock and no class nor command)
* - a class icon (it has a bsub-dock and a class, but no command nor X ID)
* - a separator (it has nothing)
 */
// type IconType int

// const (
// 	IconTypeLauncher IconType = iota
// 	IconTypeTaskbar
// 	IconTypeApplet
// 	IconTypeContainer
// 	IconTypeClass // ???
// 	IconTypeSeparatorUser
// 	IconTypeSeparatorAuto
// )

// Field defines a simple data field for dock queries.
//
type Field struct {
	Key  string
	Name string
	Icon string
}

// IconSeparator is a simple Iconer used for separators. (taskbar only ATM).
//
type IconSeparator struct {
	Field
	Taskbar bool
}

// ConfigPath returns the key.
//
func (icon *IconSeparator) ConfigPath() string {
	return icon.Key
}

// IsTaskbar returns whether the icon belongs to the taskbar or not.
//
func (icon *IconSeparator) IsTaskbar() bool {
	return icon.Taskbar
}

// IsLauncher returns whether the icon is a separator or not.
//
func (icon *IconSeparator) IsLauncher() bool {
	return false
}

// DefaultNameIcon returns improved name and image for the icon if possible.
//
func (icon *IconSeparator) DefaultNameIcon() (string, string) {
	return icon.Name, icon.Icon
}

// GetCommand is unused ATM.
func (icon *IconSeparator) GetCommand() string { return "" }

// GetClassInfo is unused ATM.
func (icon *IconSeparator) GetClassInfo(int) string { return "" }

// Reload is unused ATM.
func (icon *IconSeparator) Reload() {}

//--------------------------------------------------------[ HANDBOOK INTERFACE ]--

// Handbooker defines the interface needed by handbook module data provided as config source.
//
type Handbooker interface {
	GetTitle() string
	GetAuthor() string
	GetDescription() string
	GetPreviewFilePath() string
	GetGettextDomain() string
	GetModuleVersion() string
}
