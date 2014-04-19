// Package packages list and act on cairo-dock packages.
package packages

import (
	"github.com/sqp/godock/libs/log" // Display info in terminal.

	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/config"

	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"
)

const (
	DistantUrl  = "http://download.tuxfamily.org/glxdock/themes/third-party/"
	DistantList = "list.conf"
)

// Types of packagess.
//
type PackageType int

const (
	TypeLocal   PackageType = iota // package installed as root on the machine (in a sub-folder /usr).
	TypeUser                       // package located in the user's home
	TypeDistant                    // package present on the server
	TypeNew                        // package newly present on the server (for less than 1 month)
	TypeUpdated                    // package present locally but with a more recent version on the server, or distant package that has been updated in the past month.
	TypeInDev                      // package present locally but not on server. It's a user special applet we must not alter.
	TypeAny                        // joker (the search path function will search locally first, and on the server then).
)

//---------------------------------------------------------[ APPLET PACKAGES ]--

type AppletPackages []*AppletPackage

func (list AppletPackages) Len() int      { return len(list) }
func (list AppletPackages) Swap(i, j int) { list[i], list[j] = list[j], list[i] }

func (list AppletPackages) Exist(applet string) bool {
	return list.Get(applet) != nil
}

func (list AppletPackages) Get(applet string) *AppletPackage {
	for _, pack := range list { // Check if package exist in server list.
		if applet == pack.DisplayedName {
			return pack
		}
	}
	return nil
}

// Package list sort methods.
type ByName struct{ AppletPackages }

func (list ByName) Less(i, j int) bool {
	return list.AppletPackages[i].DisplayedName < list.AppletPackages[j].DisplayedName
}

// TODO:
// local size

//-----------------------------------------------------------[ LIST DOWNLOAD ]--

// Get the merged list of external packages in local and distant sources with downloable state.
//
func ListDownload(version string) AppletPackages {

	external, _ := ListDistant(version)
	// log.Err(e, "")
	// if e != nil {
	// }
	// Prepare the index with start size = nb packages on server.
	filled := make(map[string]*AppletPackage, len(external))
	for _, pack := range external {
		// log.Info("distant", pack.DisplayedName)
		filled[pack.DisplayedName] = pack
	}

	// Get applets dir.
	dir, eDir := DirExternal()
	if eDir == nil {
		local := ListExternalUser(dir, "applet")
		for _, pack := range local {
			if _, ok := filled[pack.DisplayedName]; !ok {
				// fmt.Println("found unknown package", pack.DisplayedName)
				pack.Type = TypeInDev
			}
			filled[pack.DisplayedName] = pack
		}
	}
	// Prepare output slice with good size and fill it with packages.
	sorted := make(AppletPackages, 0, len(filled))
	for _, pack := range filled {
		sorted = append(sorted, pack)
	}

	sort.Sort(ByName{sorted}) // Easy to get the list sorted the way we want.
	return sorted
}

//-----------------------------------------------------------------[ DISTANT ]--

// Get list of packages available on the server applets market for given version.
//
func ListDistant(version string) (AppletPackages, error) {
	url := DistantUrl + version

	// Download list from packages server.
	resp, e := http.Get(url + "/" + DistantList)
	if log.Err(e, "Connect to package server") {
		return nil, e
	}
	defer resp.Body.Close()

	// Parse distant list.
	conf, e := config.NewFromReader(resp.Body) // Special conf reflector around the config file parser.
	if log.Err(e, "Read distant applets info"); e != nil {
		return nil, e
	}

	// Create AppletPackages from parsed data.
	names := conf.Sections() // Sections names are applet names.
	list := make(AppletPackages, 0, len(names))
	for _, name := range names {
		if name != "DEFAULT" && name != "locale" { // The parser add a DEFAULT group we don't need.

			pack := &AppletPackage{}
			conf.UnmarshalGroup(reflect.ValueOf(pack).Elem(), name, config.GetTag)

			pack.DisplayedName = name
			pack.Type = TypeDistant
			pack.Path = url + "/" + name

			list = append(list, pack)
		}
	}

	return list, nil
}

//-----------------------------------------------------------[ USER EXTERNAL ]--

// Get list of packages in external applets dir.
//
func ListExternalUser(dir string, typ string) (list AppletPackages) {
	files, e := ioutil.ReadDir(dir) // ([]os.FileInfo, error)
	if e != nil {
		log.Debug("ReadDir:", e)
		return
	}

	for _, info := range files {
		if info.Name() == "po" || info.Name() == "locale" { // Drop crap.
			continue
		}

		fullpath := filepath.Join(dir, info.Name())
		info = fileGetLink(fullpath, info) // Get real dir if it is a link.
		if info.IsDir() {
			if pack := NewAppletPackageUser(dir, info.Name(), typ); pack != nil {
				list = append(list, pack)
			}
		}
	}
	return list
}

func ReadPackageFile(dir, applet, typ string) (*AppletPackage, error) {
	var file, group string
	switch typ {
	case "applet":
		file = "auto-load.conf"
		group = "Register"
	case "theme":
		file = "theme.conf"
		group = "Description"
	}
	filename := filepath.Join(dir, applet, file)
	conf, e := config.NewFromFile(filename)

	if log.Err(e, "Package description") {
		return nil, e
	}
	pack := &AppletPackage{}
	conf.UnmarshalGroup(reflect.ValueOf(pack).Elem(), group, config.GetTag)

	return pack, nil
}

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
	Title           string `conf:"Name"`
	Preview         string
	IsMultiInstance bool
	Instances       []string
	ModuleType      int
}

// NewAppletPackageUser try to read package info from dir for an external applet.
//
func NewAppletPackageUser(dir, name string, typ string) *AppletPackage {
	fullpath := filepath.Join(dir, name)

	if pack, e := ReadPackageFile(dir, name, typ); !log.Err(e, "ReadPackageFile") {
		pack.DisplayedName = name
		pack.Path = fullpath
		pack.Type = TypeUser
		pack.Size = float64(dirSize(fullpath)) / float64(MB)

		modif, e := ioutil.ReadFile(filepath.Join(fullpath, "last-modif"))
		if !log.Err(e, "Get last-modif") {
			pack.LastModifDate = strings.Replace(string(modif), "\n", "", -1) // strip \n. Check to use trimInt from Update.
		}
		return pack
	}
	return nil
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

// FormatCategory returns the human readable category for the applet.
//
func (pack *AppletPackage) FormatCategory() string {
	switch pack.Category {
	case 2:
		return "Files"
	case 3:
		return "Internet"
	case 4:
		return "Desktop"
	case 5:
		return "Accessory"
	case 6:
		return "System"
	case 7:
		return "Fun"
	}
	return ""
}

func (pack *AppletPackage) FormatState() string {
	switch pack.Type {
	case TypeLocal:
		return "Local"
	case TypeUser:
		return "User"
	case TypeDistant:
		return "Net"
	case TypeNew:
		return "New"
	case TypeUpdated:
		return "Updated"
	case TypeInDev:
		return "TypeInDev"
	}
	return ""
}

func (pack *AppletPackage) FormatSize() string {
	return ByteSize(pack.Size * float64(MB)).String()
}

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

func (pack *AppletPackage) GetDescription() string {
	if pack.Description == "" {
		switch pack.Type {
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

//-----------------------------------------------------------[ DOWNLOAD EXTERNAL ]--

// Download and extract external archive to package dir. Optional tar settings can
// be passed.
//
func (pack *AppletPackage) Install(version, options string) error {
	// func DownloadPackage(version, name, options string) error {
	// Get applets dir.
	dir, eDir := DirExternal()
	if eDir != nil {
		return eDir
	}
	// Connect a reader to the archive on server.
	resp, eNet := http.Get(DistantUrl + version + "/" + pack.DisplayedName + "/" + pack.DisplayedName + ".tar.gz")
	if eNet != nil {
		return eNet
	}
	defer resp.Body.Close()

	// Connect http reader to command tar.
	cmd := exec.Command("tar", "xz"+options) // Tar extract zip.
	cmd.Dir = dir                            // Extract in external applet directory.
	cmd.Stdin = resp.Body                    // Input is the http stream.
	cmd.Stdout = os.Stdout                   // Display output and error to console.
	cmd.Stderr = os.Stderr

	eRun := cmd.Run()
	if eRun == nil {
		lastModif := time.Now().Format("20060102")
		file := filepath.Join(dir, pack.DisplayedName, "last-modif")
		log.Err(ioutil.WriteFile(file, []byte(lastModif), 0644), "Write last-modif")

		newpack := NewAppletPackageUser(dir, pack.DisplayedName, "applet")

		pack.Type = TypeUser
		pack.Path = newpack.Path
		pack.Description = newpack.Description
		pack.Version = newpack.Version
		pack.ActAsLauncher = newpack.ActAsLauncher

		// modif, e := ioutil.ReadFile(filepath.Join(fullpath, "last-modif"))
		// if !log.Err(e, "Get last-modif") {
		// 	pack.LastModifDate = strings.Replace(string(modif), "\n", "", -1) // strip \n. Check to use trimInt from Update.
		// }

	}
	return nil
}

func (pack *AppletPackage) Uninstall(version string) error {
	if pack.Type != TypeUser && pack.Type != TypeUpdated {
		return errors.New("Wrong package type " + pack.FormatState())
	}
	dir, eDir := DirExternal()
	if eDir == nil && dir != "" && dir != "/" && pack.DisplayedName != "" {
		pack.Type = TypeDistant
		pack.Path = DistantUrl + version + "/" + pack.DisplayedName

		return os.RemoveAll(filepath.Join(dir, pack.DisplayedName))
	}
	return eDir
}

// Get Cairo-Dock external applets location.
//
func DirExternal() (dir string, e error) {
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".config", "cairo-dock", cdtype.AppletsDirName), nil
	}
	return "", errors.New("Can't get HOME directory")
}

func DirTheme(themeType string) (dir string, e error) {
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".config", "cairo-dock", "extras", themeType), nil
	}
	return "", errors.New("Can't get HOME directory")
}

// Get Cairo-Dock lanchers location.
//
func DirLaunchers() (dir string, e error) {
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".config", "cairo-dock", "current_theme", "launchers"), nil
	}
	return "", errors.New("Can't get HOME directory")
}

// Get Cairo-Dock main config file location.
//
func MainConf() (filepat string, e error) {
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".config", "cairo-dock", "current_theme", "cairo-dock.conf"), nil
	}
	return "", errors.New("Can't get HOME directory")
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

//-----------------------------------------------------------[ BYTESIZE ]--

type ByteSize float64

const (
	_           = iota // ignore first value by assigning to blank identifier
	KB ByteSize = 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
	ZB
	YB
)

// Format value to an usable range (1-999) with related unit. Three digits are provided : 123 or 12.3 or 1.23.
// No decimal under MB.
//
func (b ByteSize) String() string {
	val, unit := b.unit()
	digits := "2"
	switch {
	case b < MB:
		digits = "0"
	case val >= 100:
		digits = "0"
	case val >= 10:
		digits = "1"
	}

	return fmt.Sprintf("%."+digits+"f %s", val, unit)
}

func (b ByteSize) unit() (ByteSize, string) {
	switch {
	case b >= YB:
		return b / YB, "YB"
	case b >= ZB:
		return b / ZB, "ZB"
	case b >= EB:
		return b / EB, "EB"
	case b >= PB:
		return b / PB, "PB"
	case b >= TB:
		return b / TB, "TB"
	case b >= GB:
		return b / GB, "GB"
	case b >= MB:
		return b / MB, "MB"
	case b >= KB:
		return b / KB, "KB"
	}
	return b, "B"
}
