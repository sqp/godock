// Package cdglobal defines application and backend global consts and data.
package cdglobal

import (
	"errors"
	"go/build"
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

// AppName defines the readable application name.
//
var AppName = "NewDock" // TODO: set real value.

// Version variables also set at build time.
var (
	BuildMode     = "manual" // "manual" for custom builds. Let's also use "makefile", "package", "dockrepo"
	BuildDate     = ""       // Full readable date of build.
	BuildNbEdited = ""       // Number of files edited. Format: "not cached + already cached + untracked"
	GitHash       = ""       // Git commit hash.
)

// Original Cairo-Dock system dir locations. Can be overridden at build time.
// May be useful for distro packagers.
var (
	// CairoDockShareThemesDir defines system dock themes dir (CAIRO_DOCK_SHARE_THEMES_DIR).
	CairoDockShareThemesDir = "/usr/share/cairo-dock/themes"

	// CairoDockLocaleDir defines system dock locale dir (CAIRO_DOCK_LOCALE_DIR).
	CairoDockLocaleDir = "/usr/share/locale"

	// DownloadServerURL defines the URL of the distant theme server.
	DownloadServerURL = "http://download.tuxfamily.org/glxdock/themes" // CAIRO_DOCK_THEME_SERVER
)

// Download server locations.
const (
	DockThemeServerTag     = "themes3.4" // CAIRO_DOCK_DISTANT_THEMES_DIR
	DownloadServerListFile = "list.conf" // Name of the applets list file on the server.
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
	ConfigDirDockImages   = "images"        // Name of dir for dock images    (CAIRO_DOCK_LOCAL_IMAGES_DIR).

	// Only in DirShareData (/usr/share/cairo-dock).
	ConfigDirAppletsGo = "appletsgo" // Name of dir for go applets
	ConfigDirDefaults  = "defaults"  // Name of dir for dock default config files. TODO: need to move all old files here.
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
	GettextPackageCairoDock = "cairo-dock"         // CAIRO_DOCK_GETTEXT_PACKAGE
	GettextPackagePlugins   = "cairo-dock-plugins" // GETTEXT_NAME_EXTRAS
)

// Config files.
const (
	FileHiddenConfig  = ".cairo-dock"
	FileBuildSource   = "build.conf" // with globals.DirUserAppData
	FileChangelog     = "ChangeLog.txt"
	FileConfigThemes  = "themes.conf"    // in DirShareData
	FileCairoDockIcon = "cairo-dock.svg" // CAIRO_DOCK_ICON
	FileCairoDockLogo = "cairo-dock-logo.png"
)

// FileMode defines the default rights when writing a new file.
const FileMode os.FileMode = 0644

// CmdOpen defines the default open file/dir/url system command.
// May not be portable (lol), but it's still grouped here.
//
const CmdOpen = "xdg-open"

// AppBuildPath defines the application build location inside GOPATH.
var AppBuildPath = []string{"github.com", "sqp", "godock"}

// AppBuildPathFull returns the full path to the application build directory.
//
func AppBuildPathFull(path ...string) string {
	gopath := build.Default.GOPATH
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

// DirAppletsExternal returns external applets location.
//
func DirAppletsExternal(configDir string) (string, error) {
	configDir = ConfigDirDock(configDir)
	if configDir == "" {
		return "", errors.New("missing config dir in DirAppletsExternal")
	}
	return filepath.Join(configDir, AppletsDirName), nil
}

// DisplayMode defines the dock display backend.
type DisplayMode int

// Key display based on the display mode.
const (
	DisplayModeAll DisplayMode = iota
	DisplayModeCairo
	DisplayModeOpenGL
)

// DeskEnvironment represents a desktop environment.
//
type DeskEnvironment int

// Desktop environment backends.
//
const (
	DeskEnvGnome DeskEnvironment = iota
	DeskEnvKDE
	DeskEnvXFCE
	DeskEnvUnknown
)

// DesktEnvFromString finds the desktop environment backend from the string.
//
func DesktEnvFromString(envstr string) DeskEnvironment {
	switch envstr {
	case "gnome":
		return DeskEnvGnome
	case "kde":
		return DeskEnvKDE
	case "xfce":
		return DeskEnvXFCE
	}
	return DeskEnvUnknown
}
