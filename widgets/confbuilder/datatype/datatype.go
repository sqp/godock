// Package datatype defines the data source format for the config.
package datatype

import (
	"github.com/sqp/godock/libs/packages"
	"github.com/sqp/godock/widgets/gtk/keyfile"

	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

const (
	// KeyMainDock is the key name of the first main dock (the one with the taskbar).
	//
	KeyMainDock = "_MainDock_"

	// DirIconsSystem is the location of desktop icons themes installed on the system.
	DirIconsSystem = "/usr/share/icons"

	// DirIconsUser is the name of desktop icons themes dir in the user home dir.
	DirIconsUser = ".icons" // in $HOME
)

// Source defines external data needed by the config builder.
//
type Source interface {
	//MainConf returns the full path to the dock config file.
	//
	MainConf() string

	// AppIcon returns the application icon path.
	//
	AppIcon() string

	DirAppData() (string, error)

	DirShareData() string

	// ListIcons builds the list of all icons.
	//
	ListIcons() *ListIcon

	// ListKnownApplets builds the list of all applets.
	//
	ListKnownApplets() map[string]Appleter

	// ListDownloadApplets builds the list of downloadable user applets (installed or not).
	//
	ListDownloadApplets() map[string]Appleter

	// ListIconsMainDock builds the list of icons in the maindock.
	//
	ListIconsMainDock() []Iconer

	// ListShortkeys returns the list of dock shortkeys.
	//
	ListShortkeys() []Shortkeyer

	// ListScreens returns the list of screens (active monitors on the session).
	//
	ListScreens() []Field

	// ListViews returns the list of views.
	//
	ListViews() map[string]Handbooker

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
	ListThemeXML(localSystem, localUser, distant string) map[string]Handbooker

	// ListThemeINI builds a list of icon theme in system and user dir.
	//
	ListThemeINI(localSystem, localUser, distant string) map[string]Handbooker

	// ManagerReload reloads the manager matching the given name.
	//
	ManagerReload(name string, b bool, keyf *keyfile.KeyFile)
}

// SourceCommon provides common methods for dock config data source.
//
type SourceCommon struct{}

// ListThemeXML builds a list of icon theme in system and user dir.
//
func (SourceCommon) ListThemeXML(localSystem, localUser, distant string) map[string]Handbooker {
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
		more := dist.Get(v.DirName)
		if more != nil && more.Title != "" {
			list[k].Title = more.Title
		}
	}

	// TODO: Distant theme management will have to be moved into the download area.
	// dist, _ := packages.ListDistant(distant)
	// for _, v := range dist {
	// 	log.DEV("", v)
	// }

	out := make(map[string]Handbooker)
	for _, v := range list {
		out[v.GetName()] = v
	}

	return out
}

// ListThemeINI builds a list of icon theme in system and user dir.
//
func (SourceCommon) ListThemeINI(localSystem, localUser, distant string) map[string]Handbooker {
	// Themes installed in system dir.
	list, _ := packages.ListFromDir(localSystem, packages.TypeLocal, packages.SourceTheme)

	// Themes installed in user dir.
	if userDir, e := packages.DirTheme(localUser); e == nil {
		dist, _ := packages.ListFromDir(userDir, packages.TypeUser, packages.SourceTheme)
		list = append(list, dist...)
	}

	out := make(map[string]Handbooker)
	for _, v := range list {
		out[v.GetName()] = v
	}

	return out
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
	Install(options string) error
	Uninstall() error
	CanUninstall() bool
	IsActive() bool
	Activate() string
	Deactivate()
	CanAdd() bool

	GetTitle() string // module name translated for the user.
	GetName() string  // module name used as key.
	GetAuthor() string
	GetDescription() string
	GetPreviewFilePath() string
	GetIconFilePath() string
	IconState() string
	FormatState() string
	FormatSize() string
	FormatCategory() string
}

//
//----------------------------------------------------------------[ LISTICON ]--

// ListIcon defines data for icons list building.
//
//   Maindocks  list of container + icons. (maindocks, desklets, services)
//   Subdocks   index of SubdockName => list of icons.
//
type ListIcon struct {
	Maindocks []*ListIconContainer
	Subdocks  map[string][]Iconer
}

// NewListIcon creates a container to list dock icons.
//
func NewListIcon() *ListIcon {
	return &ListIcon{Subdocks: make(map[string][]Iconer)}
}

// Add adds a container with its icons in the list.
//
func (li *ListIcon) Add(container Iconer, icons []Iconer) {
	li.Maindocks = append(li.Maindocks, &ListIconContainer{
		Container: container,
		Icons:     icons})
}

// ListIconContainer defines a ListIcon container with its icons.
//
type ListIconContainer struct {
	Container Iconer
	Icons     []Iconer
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

	IsStackIcon() bool

	GetClassInfo(int) string
	GetCommand() string
	Reload()

	// MoveAfterNext swaps the icon position with the previous one.
	//
	MoveBeforePrevious()

	// MoveAfterNext swaps the icon position with the next one.
	//
	MoveAfterNext()

	// RemoveFromDock removes the icon from the dock.
	RemoveFromDock()

	// GetGettextDomain returns the translation domain for the applet.
	GetGettextDomain() string
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

// IconSimple provides a simple Iconer.
//
type IconSimple struct {
	Field
	Taskbar bool
}

// NewIconSimple creates a simple Iconer compatible object.
//
func NewIconSimple(key, name, icon string) *IconSimple {
	return &IconSimple{Field: Field{
		Key:  key,
		Name: name,
		Icon: icon}}
}

// ConfigPath returns the key.
//
func (is *IconSimple) ConfigPath() string {
	return is.Key
}

// IsTaskbar returns whether the icon belongs to the taskbar or not.
//
func (is *IconSimple) IsTaskbar() bool {
	return is.Taskbar
}

// IsLauncher returns whether the icon is a separator or not.
//
func (is *IconSimple) IsLauncher() bool {
	return false
}

// IsStackIcon returns whether the icon is a stack icon (subdock) or not.
//
func (is *IconSimple) IsStackIcon() bool {
	return false
}

// DefaultNameIcon returns improved name and image for the icon if possible.
//
func (is *IconSimple) DefaultNameIcon() (string, string) {
	return is.Name, is.Icon
}

// GetCommand is unused ATM.
func (is *IconSimple) GetCommand() string { return "" }

// GetClassInfo is unused ATM.
func (is *IconSimple) GetClassInfo(int) string { return "" }

// Reload is unused ATM.
func (is *IconSimple) Reload() {}

// MoveBeforePrevious is unused.
func (is *IconSimple) MoveBeforePrevious() {}

// MoveAfterNext is unused.
func (is *IconSimple) MoveAfterNext() {}

// RemoveFromDock is unused.
func (is *IconSimple) RemoveFromDock() {}

// GetGettextDomain is unused.
func (is *IconSimple) GetGettextDomain() string { return "" }

//
//------------------------------------------------------[ HANDBOOK INTERFACE ]--

// Handbooker defines the interface needed by handbook module data provided as config source.
//
type Handbooker interface {
	// GetName returns the book key.
	//
	GetName() string // name will be used as key.

	// GetTitle returns the book readable name.
	//
	GetTitle() string

	// GetAuthor returns the book author.
	//
	GetAuthor() string

	// GetDescription returns the book description.
	//
	GetDescription() string

	// GetDescription returns the book icon name or path.
	//
	GetPreviewFilePath() string

	// GetDescription returns the book gettext domain for translations.
	//
	GetGettextDomain() string

	// GetDescription returns the book version.
	//
	GetModuleVersion() string
}

// HandbookSimple provides a simple Handbooker.
//
type HandbookSimple struct {
	Key         string
	Title       string
	Author      string
	Description string
	Preview     string
}

// GetName returns the book key.
//
func (hs *HandbookSimple) GetName() string { return hs.Key }

// GetTitle returns the book readable name.
//
func (hs *HandbookSimple) GetTitle() string { return hs.Title }

// GetAuthor returns the book author.
//
func (hs *HandbookSimple) GetAuthor() string { return hs.Author }

// GetDescription returns the book description.
//
func (hs *HandbookSimple) GetDescription() string { return hs.Description }

// GetPreviewFilePath returns the book preview path.
//
func (hs *HandbookSimple) GetPreviewFilePath() string { return hs.Preview }

// GetGettextDomain is unused.
//
func (hs *HandbookSimple) GetGettextDomain() string { return "" }

// GetModuleVersion is unused.
//
func (hs *HandbookSimple) GetModuleVersion() string { return "" }

//
//------------------------------------------------------[ HANDBOOK DESC DISK ]--

// HandbookDescDisk improves Handbooker to read the description from disk,
// using the current description value as source path.
//
type HandbookDescDisk struct{ Handbooker }

// GetDescription returns the book icon name or path.
//
func (dv *HandbookDescDisk) GetDescription() string {
	body, _ := ioutil.ReadFile(dv.Handbooker.GetDescription())
	return string(body)
}

//
//-----------------------------------------------------[ HANDBOOK DESC SPLIT ]--

// HandbookDescSplit improves Handbooker by replacing \n to EOL in description.
//
type HandbookDescSplit struct{ Handbooker }

// GetDescription returns the book description.
//
func (dv *HandbookDescSplit) GetDescription() string {
	desc := dv.Handbooker.GetDescription()
	return strings.Replace(desc, "\\n", "\n", -1)
}

//
//------------------------------------------------------[ SHORTKEY INTERFACE ]--

// Shortkeyer defines the interface needed by shortkey data provided as config source.
//
type Shortkeyer interface {
	GetDemander() string
	GetDescription() string
	GetKeyString() string
	GetIconFilePath() string
	GetConfFilePath() string
	GetGroupName() string
	GetKeyName() string
	GetSuccess() bool
	Rebind(keystring, description string) bool
}

// UpdateModuleStater defines the UpdateModuleState single interface.
//
type UpdateModuleStater interface {
	UpdateModuleState(name string, active bool)
}
