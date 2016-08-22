// Package packages lists and acts on cairo-dock packages.
package packages

import (
	"github.com/sqp/godock/libs/log" // Display info in terminal.

	"github.com/sqp/godock/libs/cdglobal"
	"github.com/sqp/godock/libs/config"
	"github.com/sqp/godock/libs/text/bytesize"

	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	// "time"
)

const (
	// DistantURL is the location of cairo-dock applet market.
	DistantURL = "http://download.tuxfamily.org/glxdock/themes/"

	// DistantList is the name of the applets list file on the server.
	DistantList = "list.conf"
)

// PackageType defines the type of a package (maybe rename to state?).
//
type PackageType int

// Types of packages (location).
//
const (
	TypeLocal      PackageType = iota // package installed as root on the machine (in a sub-folder /usr).
	TypeUser                          // package located in the user's home
	TypeDistant                       // package present on the server
	TypeNew                           // package newly present on the server (for less than 1 month)
	TypeUpdated                       // package present locally but with a more recent version on the server, or distant package that has been updated in the past month.
	TypeInDev                         // package present locally but not on server. It's a user special applet we must not alter.
	TypeGoInternal                    // package included in the dock binary.
	TypeAny                           // joker (the search path function will search locally first, and on the server then).
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
func ListDownloadIndex(srvTag, externalUserDir string, source PackageSource) (map[string]*AppletPackage, error) {
	filled := make(map[string]*AppletPackage) // index by name so local packages will replace distant ones.

	found, eRet := ListDistant(srvTag)
	if eRet == nil {
		for _, pack := range found {
			filled[pack.DisplayedName] = pack
			pack.SrvTag = srvTag
		}
	}

	// Get local applets.
	local, eUsr := ListFromDir(externalUserDir, TypeUser, source)
	if eUsr != nil {
		return filled, eUsr
	}

	for _, pack := range local {
		// Flag local packages that are unknown on the server as "dev by user"
		// to prevent deletion.
		if _, ok := filled[pack.DisplayedName]; !ok {
			// fmt.Println("found unknown package", pack.DisplayedName)
			pack.Type = TypeInDev
		}

		filled[pack.DisplayedName] = pack
		pack.SrvTag = srvTag
	}

	return filled, eRet
}

// ListDownloadApplets builds the full list of external applets packages.
//
func ListDownloadApplets(externalUserDir string) (map[string]*AppletPackage, error) {
	return ListDownloadIndex(cdglobal.AppletsDirName+"/"+cdglobal.AppletsServerTag, externalUserDir, SourceApplet)
}

// ListDownloadDockThemes builds the full list of dock themes packages.
//
func ListDownloadDockThemes(themeDir string) (map[string]*AppletPackage, error) {
	return ListDownloadIndex(cdglobal.DockThemeServerTag, themeDir, SourceDockTheme)
}

//
//-----------------------------------------------------------------[ DISTANT ]--

// ListDistant lists packages available on the server applets market for given version.
//
func ListDistant(version string) (AppletPackages, error) {
	url := DistantURL + version

	// Download list from packages server.
	resp, e := http.Get(url + "/" + DistantList)
	// if log.Err(e, "Connect to package server") {
	if e != nil {
		return nil, e
	}

	defer resp.Body.Close()

	// Parse distant list.
	conf, e := config.NewFromReader(resp.Body) // Special conf reflector around the config file parser.
	if e != nil {
		// if log.Err(e, "Read distant applets info"); e != nil {
		return nil, e
	}

	// Create AppletPackages from parsed data.
	names := conf.Sections() // Sections names are applet names.
	list := make(AppletPackages, 0, len(names))
	for _, name := range names {
		if name != "DEFAULT" && name != "locale" { // The parser add a DEFAULT group we don't need.

			pack := &AppletPackage{}
			conf.UnmarshalGroup(pack, name, config.GetTag)

			pack.DisplayedName = name
			pack.Type = TypeDistant
			pack.Path = url + "/" + name

			list = append(list, pack)
		}
	}

	return list, nil
}

//
//-----------------------------------------------------------[ USER EXTERNAL ]--

// ListFromDir lists packages in external applets dir.
//
func ListFromDir(dir string, typ PackageType, source PackageSource) (AppletPackages, error) {
	files, e := ioutil.ReadDir(dir) // ([]os.FileInfo, error)
	if e != nil {
		log.Debug("ReadDir:", e)
		return nil, e
	}

	var list AppletPackages
	for _, info := range files {
		if info.Name() == "po" || info.Name() == "locale" { // Drop crap.
			continue
		}

		fullpath := filepath.Join(dir, info.Name())
		info = fileGetLink(fullpath, info) // Get real dir if it is a link.
		if info.IsDir() {
			pack, e := NewAppletPackageUser(dir, info.Name(), typ, source)
			if e == nil {
				list = append(list, pack)
			} else {
				log.Debug("packages.ListFromDir", e.Error())
			}
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
	DisplayedName string      // name of the package
	Type          PackageType // type of package : installed, user, distant...
	Path          string      // complete path of the package.
	LastModifDate string      `conf:"last modif"` // date of latest changes in the package.
	SrvTag        string      // webserver version tag to download.

	// On server only.
	CreationDate int     `conf:"creation"` // date of creation of the package.
	Size         float64 `conf:"size"`     // size in Mo
	// Rating int

	Author      string `conf:"author"` // author(s)
	Description string `conf:"description"`
	Category    int    `conf:"category"`

	Version       string `conf:"version"`
	ActAsLauncher bool   `conf:"act as launcher"`

	// From Dbus only
	Icon            string
	Title           string `conf:"name"`
	Preview         string
	IsMultiInstance bool `conf:"multi-instance"`
	Instances       []string
	ModuleType      int
}

// NewAppletPackageUser try to read an external applet package info from dir.
//
func NewAppletPackageUser(dir, name string, typ PackageType, source PackageSource) (*AppletPackage, error) {

	pack, e := ReadPackageFile(dir, name, source)
	if e != nil {
		return nil, e
	}

	fullpath := filepath.Join(dir, name)
	pack.DisplayedName = name
	pack.Path = fullpath
	pack.Type = typ
	pack.Size = float64(dirSize(fullpath)) / float64(bytesize.MB)

	// modif, e := ioutil.ReadFile(filepath.Join(fullpath, "last-modif"))
	// if !log.Err(e, "Get last-modif") {
	// 	pack.LastModifDate = strings.Replace(string(modif), "\n", "", -1) // strip \n. Check to use trimInt from Update.
	// }
	return pack, nil
}

// ReadPackageFile loads a package from its config file on disk.
//
func ReadPackageFile(dir, applet string, source PackageSource) (*AppletPackage, error) {
	var file, group string
	switch source {
	case SourceApplet:
		file = "auto-load.conf"
		group = "Register"

	case SourceTheme:
		file = "theme.conf"
		group = "Description"

	case SourceDockTheme:
		return &AppletPackage{}, nil
	}
	filename := filepath.Join(dir, applet, file)
	conf, e := config.NewFromFile(filename)

	if e != nil {
		return nil, e
	}
	pack := &AppletPackage{}
	conf.UnmarshalGroup(pack, group, config.GetTag)

	return pack, nil
}

// IsInstalled return true if the package is installed on disk.
//
func (pack *AppletPackage) IsInstalled() bool {
	return pack.Type == TypeUser || pack.Type == TypeUpdated || pack.Type == TypeInDev
}

// Dir gives the location of the package on disk.
// FIXME: do not hope that the icon is in the same dir as the applet.
// Currently based on icon location, it could really be improved.
//
func (pack *AppletPackage) Dir() string {
	return filepath.Dir(pack.Icon)
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
	switch pack.Type {
	case TypeLocal:
		return "Local"
	case TypeUser:
		return "User"
	case TypeDistant:
		return "On server"
	case TypeNew:
		return "New"
	case TypeUpdated:
		return "Updated"
	case TypeInDev:
		return "Dev by user"
	}
	return ""
}

// IconState returns the icon location for the state for the applet.
//
func (pack *AppletPackage) IconState() string {
	switch pack.Type {
	case TypeLocal:
		return "icons/theme-local.svg"
	case TypeUser:
		return "icons/theme-user.svg"
	case TypeDistant:
		return "icons/theme-distant.svg"
	case TypeNew:
		return "icons/theme-new.svg"
	case TypeUpdated:
		return "icons/theme-updated.svg"
	case TypeInDev:
		// return "TypeInDev"
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
	case TypeDistant, TypeNew: // Applets not on disk.

		resp, eNet := http.Get(pack.Path + "/preview")
		if log.Err(eNet, "Get applet image") {
			return "", false
		}
		defer resp.Body.Close()

		result, eRead := ioutil.ReadAll(resp.Body)
		if log.Err(eRead, "Download applet image") {
			return "", false
		}

		// Open destination file.
		var f *os.File
		var e error
		if tmp == "" {
			f, e = ioutil.TempFile("", "cairo-dock-appletPreview-") // Need to create a new temp file
		} else {
			if e = os.Remove(tmp); !log.Err(e, "Delete temp preview") { // We already have a temp file. Recycle it.
				f, e = os.Create(tmp)
			}
		}

		// Write data to file.
		if !log.Err(e, "Create temp file") {
			defer f.Close()
			if _, e = f.Write(result); !log.Err(e, "Write temp file") {
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
		case TypeLocal, TypeUser:
			body, _ := ioutil.ReadFile(filepath.Join(pack.Path, "readme"))
			pack.Description = string(body)

		case TypeDistant, TypeNew: // Applets not on disk.

			resp, eNet := http.Get(pack.Path + "/readme")
			if log.Err(eNet, "Get applet readme") {
				return ""
			}
			defer resp.Body.Close()

			result, eRead := ioutil.ReadAll(resp.Body)
			if log.Err(eRead, "Download applet readme") {
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
func (pack *AppletPackage) Install(externalUserDir, options string) error {
	// Connect a reader to the archive on server.
	resp, eNet := http.Get(DistantURL + pack.SrvTag + "/" + pack.DisplayedName + "/" + pack.DisplayedName + ".tar.gz")
	if eNet != nil {
		return eNet
	}
	defer resp.Body.Close()

	// Connect http reader to tar command.
	cmd := exec.Command("tar", "xz"+options) // Tar extract zip.
	cmd.Dir = externalUserDir                // Extract in external applet externalUserDirectory.
	cmd.Stdin = resp.Body                    // Input is the http stream.
	cmd.Stdout = os.Stdout                   // Display output and error to console.
	cmd.Stderr = os.Stderr

	eRun := cmd.Run()
	if eRun != nil {
		return eRun
	}
	return pack.SetInstalled(externalUserDir)
}

// SetInstalled updates package data with info from disk after download.
//
func (pack *AppletPackage) SetInstalled(externalUserDir string) error {

	// lastModif := time.Now().Format("20060102")
	// file := filepath.Join(dir, pack.DisplayedName, "last-modif")
	// log.Err(ioutil.WriteFile(file, []byte(lastModif), 0644), "Write last-modif")

	newpack, e := NewAppletPackageUser(externalUserDir, pack.DisplayedName, TypeUser, SourceApplet)
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
	if pack.Type != TypeUser && pack.Type != TypeUpdated {
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
	pack.Type = TypeDistant
	pack.Path = DistantURL + pack.SrvTag + "/" + pack.DisplayedName
	return nil
}

//
//--------------------------------------------------------------------[ DIRS ]--

// DirAppletsExternal returns external applets location.
//
func DirAppletsExternal(configDir string) (string, error) {
	dir, e := dirUserConfig(configDir)
	if e != nil {
		return "", e
	}
	return filepath.Join(dir, cdglobal.AppletsDirName), nil

}

// DirTheme returns external theme location for the given theme type.
//
func DirTheme(themeType string) (dir string, e error) {
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".config", "cairo-dock", "extras", themeType), nil
	}
	return "", errors.New("can't get HOME directory")
}

// DirLaunchers returns launchers location.
//
func DirLaunchers() (dir string, e error) {
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".config", "cairo-dock", "current_theme", "launchers"), nil
	}
	return "", errors.New("can't get HOME directory")
}

// MainConf returns the location of the Cairo-Dock main config file.
//
func MainConf() (filepat string, e error) {
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".config", "cairo-dock", "current_theme", "cairo-dock.conf"), nil
	}
	return "", errors.New("can't get HOME directory")
}

// dirUserConfig gets cairo-dock config directory, using the alternate config dir if
// provided.
//
func dirUserConfig(configDir string) (dir string, e error) {
	if configDir != "" {
		return configDir, nil
	}

	home := os.Getenv("HOME")
	if home == "" {
		return "", errors.New("can't get HOME directory")
	}
	return filepath.Join(home, ".config", "cairo-dock"), nil
}

/*
func stripComments(l string) string {
	// Comments are preceded by space or TAB
	for _, c := range []string{" ;", "\t;", " #", "\t#"} {
		if i := strings.Index(l, c); i != -1 {
			l = l[0:i]
		}
	}
	return l
}
*/

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
	DirName string      // really = directory name (used as key).
	Title   string      `xml:"name"`
	Author  string      `xml:"author"` // author(s)
	Version string      `xml:"version"`
	Type    PackageType // type of package : installed, user, distant...
	path    string
}

// ListThemesDir lists themes in a given directory.
//
func ListThemesDir(dir string, typ PackageType) ([]Theme, error) {
	files, e := ioutil.ReadDir(dir) // ([]os.FileInfo, error)
	if e != nil {
		log.Debug("ReadDir:", e)
		return nil, e
	}

	var list []Theme
	for _, info := range files {
		info = fileGetLink(filepath.Join(dir, info.Name()), info) // Get real dir if it is a link.
		if info.IsDir() {
			fullpath := filepath.Join(dir, info.Name(), "theme.xml")
			body, _ := ioutil.ReadFile(fullpath)

			// Parse data.
			theme := Theme{Type: typ, DirName: info.Name(), path: dir}
			gauge := Gauge{Theme: theme}
			if e := xml.Unmarshal(body, &gauge); e == nil {
				list = append(list, gauge.Theme)
			}

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
func fileGetLink(filename string, info os.FileInfo) os.FileInfo {
	if link, e := filepath.EvalSymlinks(filename); e == nil { // No else case, we dont case if it wasn't a link.
		//~ log.Println("was link", link)
		info, e = os.Stat(link)
		log.Err(e, "Get link")
	}
	return info
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

// FormatCategory returns the human readable category for the applet.
//
func FormatCategory(cat int) (text, RGB string) {
	switch cat {
	case 0:
		return "Behavior", "888888"
	case 2:
		return "Files", "004EA1"
	case 3:
		return "Internet", "FF5555"
	case 4:
		return "Desktop", "116E08"
	case 5:
		return "Accessory", "900009"
	case 6:
		return "System", "A58B0D"
	case 7:
		return "Fun", "FF55FF"
	}
	return "", ""
}
