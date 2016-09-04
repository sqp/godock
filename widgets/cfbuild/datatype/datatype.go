// Package datatype defines the data source format for the config.
package datatype

import (
	"github.com/sqp/godock/libs/cdglobal" // Dock types.
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/packages"
	"github.com/sqp/godock/widgets/gtk/keyfile"

	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
)

const (
	// KeyMainDock is the key name of the first main dock (the one with the taskbar).
	//
	KeyMainDock = "_MainDock_"

	// KeyNewDock is the key name for a new dock to create.
	//
	KeyNewDock = "_New Dock_"
)

// Custom config groups for the Icons GUI.
const (
	// GroupServices is the key name for the services group.
	//
	GroupServices = "_services_"

	// TitleServices is the displayed name for the services group (translatable).
	//
	TitleServices = "Services"

	// GroupDesklets is the key name for the desklets group.
	//
	GroupDesklets = "Desklets"

	// TitleDesklets is the displayed name for the desklets group (translatable).
	//
	TitleDesklets = "Desklets"

	// FieldTaskBar is the key name for the taskbar field.
	//
	FieldTaskBar = "TaskBar"

	// TitleTaskBar is the displayed name for the taskbar field (translatable).
	//
	TitleTaskBar = "--[ Taskbar ]--"
)

// Icons locations.
const (
	// DirIconsSystem is the location of desktop icons themes installed on the system.
	//
	DirIconsSystem = "/usr/share/icons"

	// DirIconsUser is the name of desktop icons themes dir in the user home dir.
	//
	DirIconsUser = ".icons" // in $HOME
)

// DisplayMode defines the dock display backend.
type DisplayMode int

// Key display based on the display mode.
const (
	DisplayModeAll DisplayMode = iota
	DisplayModeCairo
	DisplayModeOpenGL
)

// Source defines external data needed by the config builder.
//
type Source interface {
	cdglobal.Crypto // Encrypt and Decrypt string.

	//MainConfigFile returns the full path to the dock config file.
	//
	MainConfigFile() string

	//MainConfigDefault returns the full path to the dock config file.
	//
	MainConfigDefault() string

	// AppIcon returns the application icon path.
	//
	AppIcon() string

	// DirUserAppData returns the path to user applet common data in ~/.config/cairo-dock/
	//
	DirUserAppData(path ...string) (string, error)

	// DirShareData returns the path to the shared data dir (/usr/share/cairo-dock/).
	//
	DirShareData(path ...string) string

	// DesktopClasser allows to get desktop class informations for a given name.
	//
	DesktopClasser(class string) DesktopClasser

	// DisplayMode tells which renderer mode is used.
	//
	DisplayMode() DisplayMode

	// ListIcons builds the list of all icons.
	//
	ListIcons() *ListIcon

	// ListKnownApplets builds the list of all applets.
	//
	ListKnownApplets() map[string]Appleter

	// ListDownloadApplets builds the list of downloadable user applets (installed or not).
	//
	ListDownloadApplets() (map[string]Appleter, error)

	// ListIconsMainDock builds the list of icons in the maindock.
	//
	ListIconsMainDock() []Field

	// ListShortkeys returns the list of dock shortkeys.
	//
	ListShortkeys() []cdglobal.Shortkeyer

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

	// ListThemeDesktopIcon builds a list of desktop icon-themes in system and user dir.
	//
	ListThemeDesktopIcon() []Field

	// ListDockThemes builds the list of dock themes local and distant.
	//
	ListDockThemeLoad() (map[string]Appleter, error)

	// ListDockThemes builds the list of dock themes local only.
	//
	ListDockThemeSave() []Field

	// CurrentThemeSave saves the current dock theme.
	//
	CurrentThemeSave(themeName string, saveBehaviour, saveLaunchers, needPackage bool, dirPackage string) error

	// CurrentThemeLoad imports and loads a dock theme.
	//
	CurrentThemeLoad(themeName string, useBehaviour, useLaunchers bool) error

	// Handbook creates a handbook (description) for the given applet name.
	//
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

	// CreateMainDock creates a new main dock to store a moved icon.
	//
	CreateMainDock() string
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

// ListThemeDesktopIcon builds a list of desktop icon-themes in system and user dir.
//
func (SourceCommon) ListThemeDesktopIcon() []Field {

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

			kf, e := keyfile.NewFromFile(file, keyfile.FlagsNone) // Keyfile required.
			if e != nil {
				continue
			}

			hidden, _ := kf.Bool("Icon Theme", "Hidden")
			hasdirs := kf.HasKey("Icon Theme", "Directories")
			name, _ := kf.String("Icon Theme", "Name")
			kf.Free()
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
	// ConfigPath gives the full path to the icon config file.
	//
	ConfigPath() string

	// OriginalConfigPath gives the full path to the icon original config file.
	// This is the default unchanged config file.
	//
	OriginalConfigPath() string

	DefaultNameIcon() (string, string) //applets map[string]*packages.AppletPackage) (string, string)
	IsTaskbar() bool
	IsLauncher() bool

	IsStackIcon() bool

	// ConfigGroup gives the config group to build if any.
	// If no config file is set, it defines a special config key.
	//
	ConfigGroup() string

	// GetClass returns the class defined for the icon, able to get all related
	// desktop class informations.
	//
	GetClass() DesktopClasser

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
	Conf string
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
func NewIconSimple(conf, key, name, icon string) *IconSimple {
	return &IconSimple{Field: Field{
		Conf: conf,
		Key:  key,
		Name: name,
		Icon: icon}}
}

// ConfigPath returns the key.
//
func (is *IconSimple) ConfigPath() string {
	return is.Conf
}

// ConfigGroup gives the config group to build if any, or the special config key
// if no config file is defined.
//
func (is *IconSimple) ConfigGroup() string {
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

// OriginalConfigPath is unused.
func (is *IconSimple) OriginalConfigPath() string { return "" }

// GetCommand is unused ATM.
func (is *IconSimple) GetCommand() string { return "" }

// GetClass is unused ATM.
func (is *IconSimple) GetClass() DesktopClasser { return DesktopClassNil{} }

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
//-----------------------------------------------------[ UPDATE MODULE STATE ]--

// UpdateModuleStater defines the UpdateModuleState single interface.
//
type UpdateModuleStater interface {
	UpdateModuleState(name string, active bool)
}

//
//-----------------------------------------------------------[ DESKTOP CLASS ]--

// DesktopClasser defines methods to get informations about a desktop class.
//
type DesktopClasser interface {
	// String returns the desktop class as a string.
	//
	String() string

	// Name returns the desktop class application name.
	//
	Name() string

	// Command returns the desktop class command.
	//
	Command() string

	// Icon returns the desktop class icon.
	//
	Icon() string

	// MenuItems returns the list of extra commands for the class, by packs of 3
	// strings: Name, Command, Icon.
	//
	MenuItems() [][]string
}

// DesktopClassNil provides an empty DesktopClasser.
//
type DesktopClassNil struct{}

// String is unused.
func (DesktopClassNil) String() string { return "" }

// Name is unused.
func (DesktopClassNil) Name() string { return "" }

// Command is unused.
func (DesktopClassNil) Command() string { return "" }

// Icon is unused.
func (DesktopClassNil) Icon() string { return "" }

// MenuItems is unused.
func (DesktopClassNil) MenuItems() [][]string { return nil }

//
//-----------------------------------------------------------------[ HELPERS ]--

// ListFieldsKeys returns the list of row keys in a list of fields.
//
func ListFieldsKeys(fields []Field) []string {
	list := make([]string, len(fields))
	for i, v := range fields {
		list[i] = v.Key
	}
	return list
}

// ListFieldsIDByName searches the list of fields for the matching key.
// Returns the position of the field in the list.
// Returns 0 if not found, to have a valid entry to select.
//
func ListFieldsIDByName(fields []Field, key string, log cdtype.Logger) int {
	for i, field := range fields {
		if key == field.Key {
			return i
		}
	}
	if log != nil {
		log.NewErr("not found", "ListFieldsIDByName", key, fields)
	}
	return 0
}

// ListFieldsSortByName sorts the list of fields by their readable name.
//
func ListFieldsSortByName(fields []Field) {
	sort.Sort(byName(fields))
}

// byName implements sort.Interface for []Field based on the Key field.
type byName []Field

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i].Name < a[j].Name }

// ListHandbooksKeys returns the list of row keys in a list of handbooks.
//
func ListHandbooksKeys(books []Handbooker) []string {
	list := make([]string, len(books))
	for i, v := range books {
		list[i] = v.GetName()
	}
	return list
}

// IndexHandbooksKeys returns the list of row keys in an index of handbooks.
//
func IndexHandbooksKeys(books map[string]Handbooker) (list []string) {
	for _, v := range books {
		list = append(list, v.GetName())
	}
	return list
}

// IndexHandbooksToFields converts an index of handbooks to a list of fields.
//
func IndexHandbooksToFields(in map[string]Handbooker) (fields []Field) {
	for _, theme := range in {
		fields = append(fields, Field{
			Key:  theme.GetName(),
			Name: theme.GetTitle(),
		})
	}
	return
}
