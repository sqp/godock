// Package globals provides access to the dock global variables.
package globals

/*
#cgo pkg-config: gldi

#include "cairo-dock-global-variables.h"           // g_pPrimaryContainer
#include "gldi-config.h"                           // GLDI_VERSION

// extern int g_iMajorVersion, g_iMinorVersion, g_iMicroVersion;

extern gchar *g_cCairoDockDataDir;
extern gchar *g_cCurrentThemePath;
extern gchar *g_cCurrentLaunchersPath;

extern gchar *g_cConfFile;

extern CairoDock *g_pMainDock;

*/
import "C"

import (
	"github.com/sqp/godock/libs/gldi"

	"os"
	"path/filepath"
	"unsafe"
)

const (
	dirAppData = "appdata"
)

// Version returns the current version of the dock.
//
func Version() string {
	return C.GLDI_VERSION // was CAIRO_DOCK_VERSION
}

func Maindock() *gldi.CairoDock {
	return gldi.NewDockFromNative(unsafe.Pointer(C.g_pMainDock))
}

func CurrentThemePath() string {
	return C.GoString((*C.char)(C.g_cCurrentThemePath))
}

func CurrentLaunchersPath() string {
	return C.GoString((*C.char)(C.g_cCurrentLaunchersPath))
}

func ConfigFile() string {
	return C.GoString((*C.char)(C.g_cConfFile))
}

func DirDockData() string {
	return C.GoString((*C.char)(C.g_cCairoDockDataDir))
}

func DirShareData() string {
	// return C.CAIRO_DOCK_SHARE_DATA_DIR
	return C.GLDI_SHARE_DATA_DIR
}

func DirAppdata() (string, error) {
	dir := filepath.Join(DirDockData(), dirAppData)

	_, e := os.Stat(dir)
	if e != nil {
		e = os.Mkdir(dir, 0700) // Create as private as it could contain passwords. TODO: Need to confirm.
	}

	return dir, e
}

func PrimaryContainer() *gldi.Container {
	return gldi.NewContainerFromNative(unsafe.Pointer(C.g_pPrimaryContainer))
}

// themes-manager.c
// gchar *g_cExtrasDirPath = NULL;  // le chemin vers le repertoire des extra.
// gchar *g_cThemesDirPath = NULL;  // le chemin vers le repertoire des themes.
// gchar *g_cCurrentIconsPath = NULL;  // le chemin vers le repertoire des icones du theme courant.
// gchar *g_cCurrentImagesPath = NULL;  // le chemin vers le repertoire des images ou autre du theme courant.
// gchar *g_cCurrentPlugInsPath = NULL;  // le chemin vers le repertoire des plug-ins du theme courant.

// #define CAIRO_DOCK_LOCAL_EXTRAS_DIR "extras"
// #define CAIRO_DOCK_LAUNCHERS_DIR "launchers"
// #define CAIRO_DOCK_PLUG_INS_DIR "plug-ins"
// #define CAIRO_DOCK_LOCAL_ICONS_DIR "icons"
// #define CAIRO_DOCK_LOCAL_IMAGES_DIR "images"
