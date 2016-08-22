// Package cdglobal defines application and backend global consts and data.
package cdglobal

import (
	"os"
	"os/user"
	"path/filepath"
)

// AppVersion defines the application version.
//
// The -git suffix is used to tag the default value, but it should be removed if
// the Makefile was used for the build, and the real version was set.
//
var AppVersion = "0.0.3.5-git"

// Version variables also set at build time.
var (
	GitHash   = ""
	BuildDate = ""
)

// Original Cairo-Dock system dir locations. Can be overridden at build time.
// May be useful for distro packagers.
var (
	// CairoDockShareThemesDir defines system dock themes dir (CAIRO_DOCK_SHARE_THEMES_DIR).
	CairoDockShareThemesDir = "/usr/share/cairo-dock/themes"

	// CairoDockLocaleDir defines system dock locale dir (CAIRO_DOCK_LOCALE_DIR).
	CairoDockLocaleDir = "/usr/share/locale"
)

// Download server.
const (
	DownloadServerURL  = "http://download.tuxfamily.org/glxdock/themes" // CAIRO_DOCK_THEME_SERVER
	DockThemeServerTag = "themes3.4"                                    // CAIRO_DOCK_DISTANT_THEMES_DIR
)

// External applets constants.
const (
	AppletsDirName   = "third-party"
	AppletsServerTag = "3.4.0"
)

// Config dir names.
const (
	ConfigDirBaseName     = "cairo-dock"    // Default config path in .config  (CAIRO_DOCK_DATA_DIR).
	ConfigDirCurrentTheme = "current_theme" // Name of dir for current theme   (CAIRO_DOCK_CURRENT_THEME_NAME).
	ConfigDirDockThemes   = "themes"        // Name of dir for saved themes    (CAIRO_DOCK_THEMES_DIR).
	ConfigDirExtras       = "extras"        // Name of dir for extra themes    (CAIRO_DOCK_EXTRAS_DIR).
	ConfigDirPlugIns      = "plug-ins"      // Name of dir for plugins         (CAIRO_DOCK_PLUG_INS_DIR).
	ConfigDirAppletsGo    = "appletsgo"     // Name of dir for go applets      (in /usr/share/cairo-dock)
	ConfigDirDockImages   = "images"        // Name of dir for dock images    (CAIRO_DOCK_LOCAL_IMAGES_DIR).
)

// User config dir names.
const (
	// #define CAIRO_DOCK_LOCAL_EXTRAS_DIR "extras"
	// #define CAIRO_DOCK_LAUNCHERS_DIR "launchers"
	// #define CAIRO_DOCK_LOCAL_ICONS_DIR "icons"

	DirUserAppData = "appdata" // store user common applets data in ~/.cairo-dock/
)

// Translation domain.
const (
	CairoDockGettextPackage = "cairo-dock" // CAIRO_DOCK_GETTEXT_PACKAGE
)

// Config files.
const (
	FileHiddenConfig  = ".cairo-dock"
	FileChangelog     = "ChangeLog.txt"
	FileConfigThemes  = "themes.conf"    // in DirShareData
	FileCairoDockIcon = "cairo-dock.svg" // CAIRO_DOCK_ICON
	FileCairoDockLogo = "cairo-dock-logo.png"
)

// CmdOpen defines the default open file/dir/url system command.
// May not be portable (lol), but it's still grouped here.
//
const CmdOpen = "xdg-open"

// AppBuildPath defines the application build location inside GOPATH.
var AppBuildPath = []string{"github.com", "sqp", "godock"}

// AppBuildPathFull returns the full path to the application build directory.
//
func AppBuildPathFull(path ...string) string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		println("GOPATH missing")
		return ""
	}
	basedir := append([]string{gopath, "src"}, AppBuildPath...)
	return filepath.Join(append(basedir, path...)...)
}

// ConfigDirDock returns a full path to the dock config dir, according to the given user option.
//
func ConfigDirDock(dir string) string {
	if len(dir) > 0 {
		switch dir[0] {
		case '/':
			return dir // Full path, used as is.

		case '~':
			usr, e := user.Current()
			if e == nil {
				return usr.HomeDir + dir[1:] // Relative path to the homedir.
			}
		}

		current, e := os.Getwd()
		if e == nil {
			return filepath.Join(current, dir) // Relative path to the current dir.
		}
	}

	usr, e := user.Current()
	if e == nil {
		return filepath.Join(usr.HomeDir, ".config", ConfigDirBaseName) // Default dock config path in .config.
	}

	return ""
}
