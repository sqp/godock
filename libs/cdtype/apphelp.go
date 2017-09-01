package cdtype

import (
	humanize "github.com/dustin/go-humanize"

	"github.com/sqp/godock/libs/cdglobal"     // Dock types.
	"github.com/sqp/godock/libs/files"        // Files operations.
	"github.com/sqp/godock/libs/net/download" // Network pull.
	"github.com/sqp/godock/libs/text/tran"    // Translate.

	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

// Applets resources paths, inside their data folder.
const (
	// SubDirTemplate defines the default templates dir name in applets dir.
	SubDirTemplate = "templates"

	// SubDirThemeExtra defines the default theme extra dir name in applets dir.
	SubDirThemeExtra = "themes"
)

//
//--------------------------------------------------------------[ NEW APPLET ]--

// Applets stores registered applets creation calls.
//
var Applets = make(ListApps)

// ListApps defines a list of applet creation func, indexed by applet name.
//
type ListApps map[string]NewAppletFunc

// NewAppletFunc defines an applet creation function.
//
type NewAppletFunc func(base AppBase, events *Events) AppInstance

// Register registers an applet name with its new func.
//
func (l ListApps) Register(name string, callnew NewAppletFunc) {
	l[name] = callnew
}

// Unregister unregisters an applet name.
//
func (l ListApps) Unregister(name string) {
	delete(l, name)
}

// GetNewFunc gets the applet creation func for the name.
//
func (l ListApps) GetNewFunc(name string) NewAppletFunc {
	return l[name]
}

//
//----------------------------------------------------------------[ DURATION ]--

// Duration converts a time duration to seconds.
//
// Really basic, so you have to recreate one every time.
// Used by the auto config parser with tags "unit", "default" and "min".
//
type Duration struct {
	value      int
	multiplier int
}

// NewDuration creates an time duration helper.
//
func NewDuration(value int) *Duration {
	if value < 1 {
		value = 1
	}
	return &Duration{value: value, multiplier: 1}
}

// Value gets the time duration in seconds.
//
func (i *Duration) Value() int {
	return i.value * i.multiplier
}

// SetDefault sets a default duration value.
//
func (i *Duration) SetDefault(def int) {
	if i.value == 0 {
		i.value = def
	}
}

// SetMin sets a min duration value.
//
func (i *Duration) SetMin(min int) {
	if min < 1 {
		min = 1
	}
	if i.value < min {
		i.value = min
	}
}

// SetUnit sets the time unit multiplier.
//
func (i *Duration) SetUnit(unitTime string) error {
	switch strings.ToLower(unitTime) {
	case "":
		return nil

	case "s", "second":
		i.multiplier = 1
		return nil

	case "m", "minute":
		i.multiplier = 60
		return nil

	case "h", "hour":
		i.multiplier = 3600
		return nil
	}
	return errors.New("Duration.SetUnit: unknown unit=" + unitTime)
}

//
//----------------------------------------------------------------[ TEMPLATE ]--

// Template defines a template formatter.
// Can be used in applet config struct with a "default" tag.
//
type Template struct {
	*template.Template // extends the basic template.
	FilePath           string
}

// NewTemplate creates a template with the given file path.
//
func NewTemplate(name, appdir string) (*Template, error) {
	paths := []string{
		filepath.Join(appdir, SubDirTemplate, name+".tmpl"), // short template name in templates dir.
		filepath.Join(appdir, SubDirTemplate, name),         // template name with ext in templates dir.
		filepath.Clean(name),                                // full path to exact filename.
	}
	for _, path := range paths {
		_, e := os.Stat(path)
		if e != nil && os.IsNotExist(e) {
			continue
		}
		path, e = filepath.EvalSymlinks(path)
		if e != nil {
			return nil, e
		}

		tmpl, e := template.New(name).Funcs(template.FuncMap{
			"fmtTime":  humanize.Time,
			"fmtBytes": humanize.IBytes,
		}).ParseFiles(path)

		if e != nil {
			return nil, e
		}

		return &Template{
			FilePath: path,
			Template: tmpl,
		}, nil
	}
	return nil, errors.New("template not found:" + name)
}

// ToString executes a template function with the given data.
//
func (t *Template) ToString(funcName string, data interface{}) (string, error) {
	if t.Template == nil {
		return "", errors.New("template not found: " + t.FilePath)
	}
	buff := bytes.NewBuffer(nil) // NewBuffer([]byte(""))
	e := t.ExecuteTemplate(buff, funcName, data)
	return buff.String(), e
}

//
//----------------------------------------------------------------[ SHORTKEY ]--

// Shortkey defines mandatory informations to register a shortkey.
// Can be used in applet config struct with a "desc" tag.
//
type Shortkey struct {
	ConfGroup string
	ConfKey   string
	Shortkey  string       // value
	Desc      string       // tag "desc"
	ActionID  int          // tag "action". Will be converted to Call at SetDefaults.
	Call      func()       // Simple callback when triggered.
	CallE     func() error // Error returned callback. If set, used first.
}

// TestKey tests if a shortkey registered the given key name.
// Launch the registered CallE or Call and returns true if found.
//
// Error returned are callbacks errors.
// It can only happen when the key was matched and CallE returned something.
//
func (sa *Shortkey) TestKey(key string) (matched bool, cberr error) {
	switch {
	case sa.Shortkey != key: // Wrong key.

	case sa.CallE != nil:
		cberr = sa.CallE()
		matched = true

	case sa.Call != nil:
		sa.Call()
		matched = true
	}
	return matched, cberr
}

//
//-------------------------------------------------------------[ EXTRA THEME ]--

// ThemeExtra defines a custom applet theme.
//
type ThemeExtra struct {
	path string
}

// NewThemeExtra creates a theme extra with the given file path.
//
func NewThemeExtra(log Logger,
	themeName, def, // user defined theme and default theme.
	confDir, appDir, // paths.
	dirThemeSyst, // from config file: full path to extras in system dir
	// and dir names for local (in user config) and distant (in theme server).
	hintLocal, hintDist string) (*ThemeExtra, error) {

	var (
		dirThemeHome string
		dirThemeApp  string
	)
	if hintLocal != "" {
		dirThemeHome = filepath.Join(confDir, cdglobal.ConfigDirExtras, hintLocal)
		dirThemeApp = filepath.Join(appDir, SubDirThemeExtra)
	}

	var typ PackageType
	if themeName != "" {
		themeName, typ = PackType(themeName)
	}

	var tests []func() string
	if themeName != "" {
		if typ != PackTypeUpdated {
			tests = append(tests,
				testThemeDir(dirThemeHome, themeName), // First try provided name in config extras dir.
				testThemeDir(dirThemeSyst, themeName), // Then in the system dir provided as config option one [xxx;;].
			)
		}
		tests = append(tests,
			testDownload(log, hintDist, dirThemeHome, themeName), // Try to download from server.
		)
	}

	tests = append(tests,
		testThemeDir(dirThemeHome, def), // Next, try same dirs with default theme name.
		testThemeDir(dirThemeSyst, def), //
		testThemeDir(dirThemeApp, def),  // Last chance is default in applet dir.
	)

	path := ""
	for _, call := range tests {
		path = call()
		if path != "" {
			break
		}
	}

	if path == "" {
		return nil, errors.New("theme not found:" + themeName)
	}

	return &ThemeExtra{
		path: path,
	}, nil
}

// Path returns the full path to the theme dir.
//
func (te *ThemeExtra) Path() string { return te.path }

// PackType finds the package name and type from the config key value.
//
func PackType(name string) (string, PackageType) {
	idx := strings.LastIndex(name, "[")
	if name == "" || idx < 1 || !strings.HasSuffix(name, "]") {
		return name, PackTypeAny
	}
	typ, e := strconv.Atoi(name[idx+1 : len(name)-1])
	if e != nil {
		println("PackType convert", name[idx+1:len(name)-1], "err=", e.Error())
	}
	return name[:idx], PackageType(typ)
}

func testThemeDir(dir, name string) func() string {
	return func() string {
		path := filepath.Join(dir, name)
		if dir == "" || !files.IsExist(path) { // TODO NEED ISDIR
			return ""
		}
		return path
	}
}

func testDownload(log Logger, hintDist, dir, name string) func() string {
	return func() string {
		if hintDist == "" {
			return ""
		}
		path, _ := DownloadPack(log, hintDist, dir, name)
		return path
	}
}

//
//----------------------------------------------------------------[ PACKAGES ]--

// PackageType defines the type of a package (maybe rename to state?).
//
type PackageType int

// Types of packages (location).
//
const (
	PackTypeLocal      PackageType = iota // package installed as root on the machine (in a sub-folder /usr).
	PackTypeUser                          // package located in the user's home
	PackTypeDistant                       // package present on the server
	PackTypeNew                           // package newly present on the server (for less than 1 month)
	PackTypeUpdated                       // package present locally but with a more recent version on the server, or distant package that has been updated in the past month.
	PackTypeInDev                         // package present locally but not on server. It's a user special applet we must not alter.
	PackTypeGoInternal                    // package included in the dock binary.
	PackTypeAny                           // joker (the search path function will search locally first, and on the server then).
)

// String returns a human readable package type.
//
func (pt PackageType) String() string { // keep in sync with Translated.
	return map[PackageType]string{
		PackTypeLocal:      "Local",
		PackTypeUser:       "User",
		PackTypeDistant:    "On server",
		PackTypeNew:        "New",
		PackTypeUpdated:    "Updated",
		PackTypeInDev:      "Dev by user",
		PackTypeGoInternal: "Go internal",
		PackTypeAny:        "",
	}[pt]
}

// Translated returns the translated package type.
//
func (pt PackageType) Translated() string { // redeclared for translation gen.
	switch pt {
	case PackTypeLocal:
		return tran.Slate("Local")
	case PackTypeUser:
		return tran.Slate("User")
	case PackTypeDistant:
		return tran.Slate("On server")
	case PackTypeNew:
		return tran.Slate("New")
	case PackTypeUpdated:
		return tran.Slate("Updated")
	case PackTypeInDev:
		return tran.Slate("Dev by user")
	case PackTypeGoInternal:
		return tran.Slate("Go internal")
	}
	return ""
}

// DownloadPack downloads a dock package from the remote server.
//
func DownloadPack(log Logger, hintDist, dir, name string) (string, error) {
	// Get downloader.
	uri := fmt.Sprintf("%s/%s/%s/%s.tar.gz",
		cdglobal.DownloadServerURL, hintDist, name, name)
	reader, e := download.Reader(uri)
	if log.Err(e, "download theme", hintDist, "from", uri) {
		return "", e
	}

	// Extract to disk.
	log.Info("downloading theme", hintDist, name)
	e = files.UnTarGz(dir, reader)
	if log.Err(e, "extract theme", hintDist, "to", dir) {
		return "", e
	}

	// Save download date.
	path := filepath.Join(dir, name)
	e = files.SetLastModif(path)
	log.Err(e, "save theme lastmodif", hintDist, "to", path)

	return path, e
}

//
//---------------------------------------------------------------[ INIT CONF ]--

// InitConf creates a config file from its default location.
//
func InitConf(log Logger, orig, dest string) error {
	e := files.CopyFile(orig, dest, os.FileMode(0644))
	if e != nil {
		log.Errorf("init conf copy default file", "%s\n%s\n  to:\n%s", e, orig, dest)
		return e
	}
	log.Info("created config file", dest)
	return nil
}
