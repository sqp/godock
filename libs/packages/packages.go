// Package packages lists and acts on cairo-dock packages.
package packages

import (
	"github.com/sqp/godock/libs/cdglobal"      // Dock types.
	"github.com/sqp/godock/libs/cdtype"        // Logger type.
	"github.com/sqp/godock/libs/config"        // Config parser.
	"github.com/sqp/godock/libs/text/bytesize" // Human readable bytes.
	"github.com/sqp/godock/libs/text/tran"     // Translate.

	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// PackageSource defines whether the loaded package is an applet or a theme.
//
type PackageSource int

// Source of package to load (applet or theme).
//
const (
	SourceApplet PackageSource = iota
	SourceTheme
	SourceDockTheme
)

// File returns the file name for the package source.
//
func (ps PackageSource) File() string {
	return map[PackageSource]string{
		SourceApplet: "auto-load.conf",
		SourceTheme:  "theme.conf",
	}[ps]
}

// Group returns the group name to parse for the package source.
//
func (ps PackageSource) Group() string {
	return map[PackageSource]string{
		SourceApplet: "Register",
		SourceTheme:  "Description",
	}[ps]
}

// AppInfoField defines edit applet info fields.
//
type AppInfoField int

// Applet info fields.
const (
	AppInfoUnknown       AppInfoField = iota // empty (or unused).
	AppInfoVersion                           // string
	AppInfoCategory                          // int
	AppInfoAuthor                            // string
	AppInfoDescription                       // string
	AppInfoActAsLauncher                     // bool
	AppInfoMultiInstance                     // bool
	AppInfoTitle                             // string
	AppInfoIcon                              // string
)

// Translated returns the translated name for the field.
//
func (field AppInfoField) Translated() string {
	switch field {
	case AppInfoVersion:
		return tran.Slate("Version")
	case AppInfoCategory:
		return tran.Slate("Category")
	case AppInfoAuthor:
		return tran.Slate("Author")
	case AppInfoDescription:
		return tran.Slate("Description")
	case AppInfoActAsLauncher:
		return tran.Slate("Act as launcher")
	case AppInfoMultiInstance:
		return tran.Slate("Multi instance")
	case AppInfoTitle:
		return tran.Slate("Alternate title")
	case AppInfoIcon:
		return tran.Slate("Icon")
	}
	return ""
}

// Key returns the config file key for the field.
//
func (field AppInfoField) Key() string {
	return map[AppInfoField]string{
		AppInfoVersion:       "version",
		AppInfoCategory:      "category",
		AppInfoAuthor:        "author",
		AppInfoDescription:   "description",
		AppInfoActAsLauncher: "act as launcher",
		AppInfoMultiInstance: "multi-instance",
		AppInfoTitle:         "title",
		AppInfoIcon:          "icon",
	}[field]
}

// Comment returns the config comment for the field.
//
func (field AppInfoField) Comment() string {
	switch field {
	case AppInfoVersion:
		return "# Version of the applet; change it everytime you change something in the config file. Don't forget to update the version both in this file and in the config file."
	case AppInfoCategory:
		return "# Category of the applet : 2 = files, 3 = internet, 4 = Desktop, 5 = accessory, 6 = system, 7 = fun"
	case AppInfoAuthor:
		return "# Author of the applet"
	case AppInfoDescription:
		return "# A short description of the applet and how to use it."
	case AppInfoActAsLauncher:
		return `# The applet is a "smart launcher"; it will behave as a launcher in the taskbar.`
	case AppInfoMultiInstance:
		return "# Whether the applet can be instanciated several times or not."
	case AppInfoTitle:
		return "# Rename the applet: useful if the name can be translated or if it contains spaces"
	case AppInfoIcon:
		return `# Default icon to use if no icon has been defined by the user. If not specified, or if the file is not found, the "icon" file will be used.`
	}
	return ""
}

//
//---------------------------------------------------------[ APPLET PACKAGES ]--

// AppletPackages defines a list of AppletPackage.
//
type AppletPackages []*AppletPackage

// Len returns the number of packages in the list.
//
func (list AppletPackages) Len() int { return len(list) }

// Swap exchanges the position of two packages.
//
func (list AppletPackages) Swap(i, j int) { list[i], list[j] = list[j], list[i] }

// Exist returns true if the package was found in the list.
//
func (list AppletPackages) Exist(applet string) bool {
	return list.Get(applet) != nil
}

// Get returns the package matching the name provided if found.
//
func (list AppletPackages) Get(applet string) *AppletPackage {
	for _, pack := range list { // Check if package exist in server list.
		if applet == pack.DisplayedName {
			return pack
		}
	}
	return nil
}

// ByName sorts the list of packages by name.
//
type ByName struct{ AppletPackages }

// Less compares packages names for the sort.
//
func (list ByName) Less(i, j int) bool {
	return list.AppletPackages[i].DisplayedName < list.AppletPackages[j].DisplayedName
}

// TODO:
// local size

//
//-----------------------------------------------------------[ LIST DOWNLOAD ]--

// ListDownloadSort sorts a list of applet packages.
//
func ListDownloadSort(list map[string]*AppletPackage) (sorted AppletPackages) {
	for _, pack := range list {
		sorted = append(sorted, pack)
	}
	sort.Sort(ByName{sorted}) // Easy to get the list sorted the way we want.
	return sorted
}

// ListDownloadIndex builds a merged list of external packages in local and distant
// sources with downloadable state, indexed by applet name.
// In case of multiple errors, the last one is returned.
// (local access errors are more important than network errors)
//
func ListDownloadIndex(log cdtype.Logger, srvTag, externalUserDir string, source PackageSource) (map[string]*AppletPackage, error) {
	filled := make(map[string]*AppletPackage) // index by name so local packages will replace distant ones.

	found, eRet := ListDistant(log, srvTag)
	if eRet == nil {
		for _, pack := range found {
			filled[pack.DisplayedName] = pack
			pack.SrvTag = srvTag
		}
	}

	// Get local applets.
	local, eUsr := ListFromDir(log, externalUserDir, cdtype.PackTypeUser, source)
	if eUsr != nil {
		return filled, eUsr
	}

	for _, pack := range local {
		// Flag local packages that are unknown on the server as "dev by user"
		// to prevent deletion.
		if _, ok := filled[pack.DisplayedName]; !ok {
			// fmt.Println("found unknown package", pack.DisplayedName)
			pack.Type = cdtype.PackTypeInDev
		}

		filled[pack.DisplayedName] = pack
		pack.SrvTag = srvTag
	}

	return filled, eRet
}

// ListDownloadApplets builds the full list of external applets packages.
//
func ListDownloadApplets(log cdtype.Logger, externalUserDir string) (map[string]*AppletPackage, error) {
	return ListDownloadIndex(log, cdglobal.AppletsDirName+"/"+cdglobal.AppletsServerTag, externalUserDir, SourceApplet)
}

// ListDownloadDockThemes builds the full list of dock themes packages.
//
func ListDownloadDockThemes(log cdtype.Logger, themeDir string) (map[string]*AppletPackage, error) {
	return ListDownloadIndex(log, cdglobal.DockThemeServerTag, themeDir, SourceDockTheme)
}

//
//-----------------------------------------------------------------[ DISTANT ]--

// ListDistant lists packages available on the server applets market for given version.
//
func ListDistant(log cdtype.Logger, version string) (AppletPackages, error) {
	url := cdglobal.DownloadServerURL + "/" + version

	// Download list from packages server.
	resp, e := http.Get(url + "/" + cdglobal.DownloadServerListFile)
	if e != nil {
		return nil, e
	}

	defer resp.Body.Close()

	// Parse distant list.
	cfg, e := config.NewFromReader(resp.Body) // Special conf reflector around the config file parser.
	if e != nil {
		return nil, e
	}

	// Create AppletPackages from parsed data.
	names := cfg.SectionStrings() // Sections names are applet names.
	list := make(AppletPackages, 0, len(names))
	for _, name := range names {
		if name == "locale" {
			continue
		}

		pack := NewAppletPackage(log)
		cfg.UnmarshalGroup(pack, name, config.GetTag)

		pack.DisplayedName = name
		pack.Type = cdtype.PackTypeDistant
		pack.Path = url + "/" + name

		list = append(list, pack)
	}

	return list, nil
}

//
//-----------------------------------------------------------[ USER EXTERNAL ]--

// ListFromDir lists packages in external applets dir.
//
func ListFromDir(log cdtype.Logger, dir string, typ cdtype.PackageType, source PackageSource) (AppletPackages, error) {
	files, e := ioutil.ReadDir(dir) // ([]os.FileInfo, error)
	if e != nil {
		return nil, e
	}

	var list AppletPackages
	for _, info := range files {
		if info.Name() == "po" || info.Name() == "locale" { // Drop translations.
			continue
		}

		// Get real dir if it is a link.
		fullpath := filepath.Join(dir, info.Name())
		info, e = fileGetLink(fullpath, info)
		if log.Err(e, "packages.fileGetLink") || !info.IsDir() {
			continue
		}

		// Load package.
		pack, e := NewAppletPackageUser(log, dir, info.Name(), typ, source)
		if e == nil {
			list = append(list, pack)
		} else {
			log.Debug("packages.ListFromDir", e.Error())
		}
	}
	return list, nil
}

//
//-----------------------------------------------------------[ APPLET PACKAGE ]--

//~ gchar *cPackagePath //

//~ gchar *cHint // hint of the package, for instance "sound" or "battery" for a gauge, "internet" or "desktop" for a third-party applet.
//~ gint iSobriety // sobriety/simplicity of the package.

// AppletPackage defines a generic cairo-dock applet package.
//
type AppletPackage struct {
	DisplayedName string             // name of the package
	Type          cdtype.PackageType // type of package : installed, user, distant...
	Path          string             // complete path of the package.
	LastModifDate string             `conf:"last modif"` // date of latest changes in the package.
	SrvTag        string             // webserver version tag to download.
	Source        PackageSource      // applet or theme.

	Author          string              `conf:"author"` // author(s)
	Description     string              `conf:"description"`
	Category        cdtype.CategoryType `conf:"category"`
	Title           string              `conf:"title"` // From file: alt applet name to use. From DBus: also translated.
	Icon            string              `conf:"icon"`  // From file: alt icon name to use.   From DBus: with full path?.
	Version         string              `conf:"version"`
	ActAsLauncher   bool                `conf:"act as launcher"`
	IsMultiInstance bool                `conf:"multi-instance"`

	// On server only.
	CreationDate int     `conf:"creation"` // date of creation of the package.
	Size         float64 `conf:"size"`     // size in Mo
	// Rating int

	// From Dbus only
	Preview    string
	Instances  []string
	ModuleType int

	log cdtype.Logger
}

// NewAppletPackage creates an empty AppletPackage.
//
func NewAppletPackage(log cdtype.Logger) *AppletPackage {
	return &AppletPackage{
		log: log,
	}
}

// NewAppletPackageUser try to read an external applet package info from dir.
//
func NewAppletPackageUser(log cdtype.Logger, dir, name string, typ cdtype.PackageType, source PackageSource) (*AppletPackage, error) {

	pack, e := ReadPackageFile(log, dir, name, source)
	if e != nil {
		return nil, e
	}

	fullpath := filepath.Join(dir, name)
	pack.DisplayedName = name
	pack.Path = fullpath
	pack.Type = typ
	pack.Size = float64(dirSize(fullpath)) / float64(bytesize.MB)
	pack.Source = source

	// modif, e := ioutil.ReadFile(filepath.Join(fullpath, "last-modif"))
	// if !log.Err(e, "Get last-modif") {
	// 	pack.LastModifDate = strings.Replace(string(modif), "\n", "", -1) // strip \n. Check to use trimInt from Update.
	// }
	return pack, nil
}

// ReadPackageFile loads a package from its config file on disk.
//
func ReadPackageFile(log cdtype.Logger, dir, applet string, source PackageSource) (*AppletPackage, error) {
	pack := NewAppletPackage(log)
	if source == SourceDockTheme {
		return pack, nil
	}
	filename := filepath.Join(dir, applet, source.File())
	e := config.GetFromFile(log, filename, func(cfg cdtype.ConfUpdater) {
		cfg.UnmarshalGroup(pack, source.Group(), config.GetTag)
	})
	return pack, e
}

// IsInstalled return true if the package is installed on disk.
//
func (pack *AppletPackage) IsInstalled() bool {
	return pack.Type == cdtype.PackTypeUser || pack.Type == cdtype.PackTypeUpdated || pack.Type == cdtype.PackTypeInDev
}

// Dir gives the location of the package on disk.
// TODO: confirm same dir as the applet.
//
func (pack *AppletPackage) Dir() string {
	return filepath.Dir(pack.Path)
}

// FormatName returns the best available package name to display.
// (translated if possible).
//
func (pack *AppletPackage) FormatName() string {
	if pack.Title != "" {
		return pack.Title
	}
	return pack.DisplayedName
}

// FormatState returns the human readable state for the applet.
//
func (pack *AppletPackage) FormatState() string {
	return pack.Type.Translated()
}

// IconState returns the icon location for the state for the applet.
//
func (pack *AppletPackage) IconState() string {
	switch pack.Type {
	case cdtype.PackTypeLocal:
		return "icons/theme-local.svg"
	case cdtype.PackTypeUser:
		return "icons/theme-user.svg"
	case cdtype.PackTypeDistant:
		return "icons/theme-distant.svg"
	case cdtype.PackTypeNew:
		return "icons/theme-new.svg"
	case cdtype.PackTypeUpdated:
		return "icons/theme-updated.svg"
	case cdtype.PackTypeInDev:
		return "icons/theme-local.svg" // TODO: improve !

	}
	return ""
}

// FormatSize returns the human readable size for the applet.
//
func (pack *AppletPackage) FormatSize() string {
	return bytesize.ByteSize(pack.Size * float64(bytesize.MB)).String()
}

// GetPreview returns the location of the applet preview on disk.
// The preview will be downloaded from the server for non installed applets.
// If a temp file location is provided, it will be used, otherwise, the
// Returned values are the file location and the boolean indicates if a temp
// file was used and need to be removed when no more useful.
//
func (pack *AppletPackage) GetPreview(tmp string) (string, bool) {
	if pack.Preview != "" {
		return pack.Preview, false
	}

	switch pack.Type {
	case cdtype.PackTypeDistant, cdtype.PackTypeNew: // Applets not on disk.

		resp, e := http.Get(pack.Path + "/preview")
		if pack.log.Err(e, "Get applet image") {
			return "", false
		}
		defer resp.Body.Close()

		result, e := ioutil.ReadAll(resp.Body)
		if pack.log.Err(e, "Download applet image") {
			return "", false
		}

		// Open destination file.
		var f *os.File
		if tmp == "" {
			f, e = ioutil.TempFile("", "cairo-dock-appletPreview-") // Need to create a new temp file
		} else {
			if e = os.Remove(tmp); !pack.log.Err(e, "Delete temp preview") { // We already have a temp file. Recycle it.
				f, e = os.Create(tmp)
			}
		}

		// Write data to file.
		if !pack.log.Err(e, "Create temp file") {
			defer f.Close()
			if _, e = f.Write(result); !pack.log.Err(e, "Write temp file") {
				return f.Name(), true
			}
		}

		return "", false // Problem with temp file.

	default:
		return filepath.Join(pack.Path, "preview"), false
	}
}

// GetDescription returns the package description text.
// Can be slow if it needs to download the file (non installed package).
//
func (pack *AppletPackage) GetDescription() string {
	if pack.Description == "" {
		switch pack.Type {
		case cdtype.PackTypeLocal, cdtype.PackTypeUser:
			body, _ := ioutil.ReadFile(filepath.Join(pack.Path, "readme"))
			pack.Description = string(body)

		case cdtype.PackTypeDistant, cdtype.PackTypeNew: // Applets not on disk.

			resp, e := http.Get(pack.Path + "/readme")
			if pack.log.Err(e, "Get applet readme") {
				return ""
			}
			defer resp.Body.Close()

			result, e := ioutil.ReadAll(resp.Body)
			if pack.log.Err(e, "Download applet readme") {
				return ""
			}
			pack.Description = string(result)
		}
	}
	return strings.Replace(pack.Description, "\\n", "\n", -1)
}

//
// Handbooker interface.

// GetTitle returns the package readable name.
//
func (pack *AppletPackage) GetTitle() string { return pack.DisplayedName }

// GetAuthor returns the package author.
//
func (pack *AppletPackage) GetAuthor() string { return pack.Author }

// GetGettextDomain is a stub. TODO: expand and use.
//
func (pack *AppletPackage) GetGettextDomain() string { return "" }

// GetModuleVersion returns the version of the package.
//
func (pack *AppletPackage) GetModuleVersion() string { return pack.Version }

// GetName returns the package name to use as config key.
//
func (pack *AppletPackage) GetName() string { return pack.DisplayedName }

// GetPreviewFilePath returns the location of the preview file.
// Can be slow if it needs to download the file (non installed package).
//
func (pack *AppletPackage) GetPreviewFilePath() string {
	file, _ := pack.GetPreview("")
	return file
}

//
//-----------------------------------------------------------[ DOWNLOAD EXTERNAL ]--

// Install downloads and extract an external archive to package dir.
// Optional tar settings can be passed.
//
func (pack *AppletPackage) Install(externalUserDir string) error {
	_, e := cdtype.DownloadPack(pack.log, pack.SrvTag, externalUserDir, pack.DisplayedName)
	if e != nil {
		return e
	}
	return pack.SetInstalled(externalUserDir)
}

// SetInstalled updates package data with info from disk after download.
//
func (pack *AppletPackage) SetInstalled(externalUserDir string) error {

	// lastModif := time.Now().Format("20060102")
	// file := filepath.Join(dir, pack.DisplayedName, "last-modif")
	// log.Err(ioutil.WriteFile(file, []byte(lastModif), 0644), "Write last-modif")

	newpack, e := NewAppletPackageUser(pack.log, externalUserDir, pack.DisplayedName, cdtype.PackTypeUser, SourceApplet)
	if e != nil {
		return e
	}

	pack.Path = newpack.Path
	pack.Type = newpack.Type
	pack.Description = newpack.Description
	pack.Version = newpack.Version
	pack.ActAsLauncher = newpack.ActAsLauncher

	// modif, e := ioutil.ReadFile(filepath.Join(fullpath, "last-modif"))
	// if !log.Err(e, "Get last-modif") {
	// 	pack.LastModifDate = strings.Replace(string(modif), "\n", "", -1) // strip \n. Check to use trimInt from Update.
	// }

	return nil
}

// Uninstall removes an external applet from disk.
//
func (pack *AppletPackage) Uninstall(externalUserDir string) error {
	if pack.Type != cdtype.PackTypeUser && pack.Type != cdtype.PackTypeUpdated {
		return errors.New("wrong package type " + pack.FormatState())
	}
	appdir := filepath.Join(externalUserDir, pack.DisplayedName)
	if externalUserDir == "" || externalUserDir == "/" || pack.DisplayedName == "" {
		return errors.New("wrong package dir " + appdir)
	}

	// return errors.New("DISABLED TEMP", appdir, pack.DisplayedName)

	e := os.RemoveAll(appdir)
	if e != nil {
		return e
	}
	pack.Type = cdtype.PackTypeDistant
	pack.Path = cdglobal.DownloadServerURL + "/" + pack.SrvTag + "/" + pack.DisplayedName
	return nil
}

// SaveUpdated updates the package, then saves and returns a reloaded package.
//
func (pack *AppletPackage) SaveUpdated(edits map[AppInfoField]interface{}) (*AppletPackage, error) {
	if len(edits) == 0 {
		return nil, errors.New("package save: no fields to update")
	}

	group := pack.Source.Group()
	newver := ""

	// Update version first in applet config file.
	// When an upgrade is requested, we have to change versions in both files,
	// so the config upgrade becomes mandatory.
	if delta, upVer := edits[AppInfoVersion]; upVer {
		var e error
		newver, e = FormatNewVersion(pack.Version, delta.(int))
		if e != nil {
			return nil, e
		}
		filename := filepath.Join(pack.Path, pack.DisplayedName+".conf")
		e = config.SetFileVersion(pack.log, filename, "Icon", pack.Version, newver)
		if e != nil {
			return nil, e
		}
	}

	// Update edited fields.
	filename := filepath.Join(pack.Path, pack.Source.File())
	e := config.SetToFile(pack.log, filename, func(cfg cdtype.ConfUpdater) (e error) {
		for k, v := range edits {
			switch k {
			case AppInfoVersion:
				e = cfg.Set(group, k.Key(), newver)

			case AppInfoCategory:
				e = cfg.Set(group, k.Key(), int(v.(cdtype.CategoryType)))

			case AppInfoActAsLauncher, AppInfoMultiInstance:
				e = cfg.Set(group, k.Key(), v.(bool))

			case AppInfoAuthor, AppInfoDescription, AppInfoTitle, AppInfoIcon:
				e = cfg.Set(group, k.Key(), v.(string))
			}

			if e != nil {
				return e
			}
			comment, e := cfg.GetComment(group, k.Key())
			if !pack.log.Err(e, "package update get comment") && comment == "" {
				cfg.SetComment(group, k.Key(), k.Comment())
			}
		}
		return nil
	})
	if e != nil {
		return nil, e
	}

	// Reload data from disk.
	return NewAppletPackageUser(pack.log,
		filepath.Dir(pack.Path), filepath.Base(pack.Path),
		pack.Type, pack.Source)
}

//
//----------------------------------------------------------[ APPLETS THEMES ]--

// Gauge is an icon theme.
//
type Gauge struct {
	XMLName xml.Name `xml:"gauge"`
	Theme
}

// Theme represents an icon theme (gauge, clock...).
//
type Theme struct {
	// Name    string      `xml:"name"`   // name of the package
	DirName string             // really = directory name (used as key).
	Title   string             `xml:"name"`
	Author  string             `xml:"author"` // author(s)
	Version string             `xml:"version"`
	Type    cdtype.PackageType // type of package : installed, user, distant...
	path    string
}

// ListThemesDir lists themes in a given directory.
//
func ListThemesDir(log cdtype.Logger, dir string, typ cdtype.PackageType) ([]Theme, error) {
	files, e := ioutil.ReadDir(dir) // ([]os.FileInfo, error)
	if e != nil {
		return nil, e
	}

	var list []Theme
	for _, info := range files {
		info, e = fileGetLink(filepath.Join(dir, info.Name()), info) // Get real dir if it is a link.
		if log.Err(e, "packages.fileGetLink") || !info.IsDir() {
			continue
		}
		fullpath := filepath.Join(dir, info.Name(), "theme.xml")
		body, _ := ioutil.ReadFile(fullpath)

		// Parse data.
		theme := Theme{Type: typ, DirName: info.Name(), path: dir}
		gauge := Gauge{Theme: theme}
		if e := xml.Unmarshal(body, &gauge); e == nil {
			list = append(list, gauge.Theme)
		}
	}
	return list, nil
}

//
// Handbooker interface.

// GetTitle returns the package readable name.
//
func (t Theme) GetTitle() string { return t.Title }

// GetAuthor returns the package author.
//
func (t Theme) GetAuthor() string { return t.Author }

// GetGettextDomain is a stub. TODO: expand and use.
//
func (t Theme) GetGettextDomain() string { return "" }

// GetModuleVersion returns the version of the package.
//
func (t Theme) GetModuleVersion() string { return t.Version }

// GetName returns the package name to use as config key.
//
func (t Theme) GetName() string { return fmt.Sprintf("%s[%d]", t.DirName, t.Type) }

// GetDescription returns the package description text.
// Can be slow if it needs to download the file (non installed package).
//
func (t Theme) GetDescription() string {
	body, _ := ioutil.ReadFile(filepath.Join(t.path, t.DirName, "readme"))
	return string(body)
}

// GetPreviewFilePath returns the location of the preview file.
// Can be slow if it needs to download the file (non installed package).
//
func (t Theme) GetPreviewFilePath() string { return filepath.Join(t.path, t.DirName, "preview") }

//
//-----------------------------------------------------------[ HELPER ]--

// Follow link if needed to get real file or dir. Give it your current FileInfo
// and it will be replaced by the link target it was a link (meaning if it is not
// a link the provided FileInfo will not change).
//
func fileGetLink(filename string, info os.FileInfo) (os.FileInfo, error) {
	link, e := filepath.EvalSymlinks(filename)
	if e != nil { // No else case, we dont care if it wasn't a link.
		return nil, e
	}
	return os.Stat(link)
}

func dirSize(location string) (size int64) {
	dir, err := os.Open(location)
	if err != nil {
		return
	}
	defer dir.Close()

	if fileInfos, err := dir.Readdir(-1); err == nil {
		for _, info := range fileInfos {
			if info.IsDir() {
				size += dirSize(filepath.Join(location, info.Name()))
			} else {
				size += info.Size()
				// fmt.Println(info.Name(), info.Size())
			}
		}
	}
	return
}

// FormatNewVersion formats an upgraded X.Y.Z version string.
//
// Value for delta:
//   1: increase micro  X.Y.Z+1
//   2: increase minor  X.Y+1.0
//   3: increase major  X+1.0.0
//
func FormatNewVersion(str string, delta int) (string, error) {
	var vers [3]int
	_, e := fmt.Sscanf(str, "%d.%d.%d", &vers[0], &vers[1], &vers[2])
	if e != nil {
		return "", errors.New("scan version failed (" + str + ") : " + e.Error())
	}

	switch delta {
	case 1:
		vers[2]++

	case 2:
		vers[1]++
		vers[2] = 0

	case 3:
		vers[0]++
		vers[1] = 0
		vers[2] = 0

	default:
		return "", errors.New("bad delta value:" + strconv.Itoa(delta))
	}
	return strconv.Itoa(vers[0]) + "." + strconv.Itoa(vers[1]) + "." + strconv.Itoa(vers[2]), nil
}
