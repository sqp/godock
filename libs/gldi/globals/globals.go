// Package globals provides access to the dock global variables.
package globals

/*
#cgo pkg-config: gldi

#include "cairo-dock-container.h"                  // NOTIFICATION_CLICK_ICON ...
#include "cairo-dock-desklet-manager.h"            // myDeskletObjectMgr
#include "cairo-dock-dock-manager.h"               // NOTIFICATION_ENTER_DOCK
#include "cairo-dock-global-variables.h"           // g_pPrimaryContainer
#include "cairo-dock-keybinder.h"                  // NOTIFICATION_SHORTKEY_CHANGED
#include "cairo-dock-module-manager.h"             // myModuleObjectMgr
#include "cairo-dock-module-instance-manager.h"    // NOTIFICATION_MODULE_INSTANCE_DETACHED
#include "cairo-dock-windows-manager.h"            // myWindowObjectMgr
#include "gldi-icon-names.h"                       // GLDI_ICON_NAME_*
#include "gldi-config.h"                           // GLDI_VERSION


extern int g_iMajorVersion, g_iMinorVersion, g_iMicroVersion;

extern gchar *g_cCairoDockDataDir;
extern gchar *g_cCurrentThemePath;
extern gchar *g_cCurrentLaunchersPath;
extern gchar *g_cCurrentIconsPath;
extern gchar *g_cConfFile;

extern CairoDock *g_pMainDock;

*/
import "C"

import (
	"github.com/sqp/godock/libs/cdglobal" // Global consts.

	"github.com/sqp/godock/libs/gldi"

	"os"
	"path/filepath"
	"unsafe"
)

var (
	// FullLock represents the full settings lock from -k option at start.
	// Config must not be changed if true.
	FullLock bool
)

// Version returns the current version of the dock.
//
func Version() string {
	return C.GLDI_VERSION // was CAIRO_DOCK_VERSION
}

// VersionSplit returns the current version of the dock.
//
func VersionSplit() (int, int, int) {
	return int(C.g_iMajorVersion), int(C.g_iMinorVersion), int(C.g_iMicroVersion) // was CAIRO_DOCK_VERSION
}

// GtkVersion returns the GTK version used to compile gldi.
//
func GtkVersion() (int, int, int) {
	return C.GTK_MAJOR_VERSION, C.GTK_MINOR_VERSION, C.GTK_MICRO_VERSION
}

//
//-------------------------------------------------------------------[ PATHS ]--

// CurrentThemePath returns the path to the current theme (in use).
//
func CurrentThemePath(path ...string) string {
	dir := C.GoString((*C.char)(C.g_cCurrentThemePath))
	path = append([]string{dir}, path...)
	return filepath.Join(path...)
}

// CurrentLaunchersPath returns the path to the launchers dir.
//
func CurrentLaunchersPath() string {
	return C.GoString((*C.char)(C.g_cCurrentLaunchersPath))
}

// ConfigFile returns the path of the current config file (in use).
//
func ConfigFile() string {
	return C.GoString((*C.char)(C.g_cConfFile))
}

// ConfigFileDefault returns the path of the default config file (unchanged).
//
func ConfigFileDefault() string {
	return DirShareData(C.CAIRO_DOCK_CONF_FILE)
}

// DirDockData returns the path to the dock data dir.
//
func DirDockData(path ...string) string {
	dir := C.GoString((*C.char)(C.g_cCairoDockDataDir))
	path = append([]string{dir}, path...)
	return filepath.Join(path...)
}

// DirCurrentIcons returns the path to current theme icons dir.
//
func DirCurrentIcons() string {
	return C.GoString((*C.char)(C.g_cCurrentIconsPath))
}

// DirShareData returns the path to the share data dir.
//
func DirShareData(path ...string) string {
	path = append([]string{C.GLDI_SHARE_DATA_DIR}, path...) // was CAIRO_DOCK_SHARE_DATA_DIR
	return filepath.Join(path...)
}

// DirUserAppData returns the path to user applet common data in ~/.cairo-dock/
//
func DirUserAppData(path ...string) (string, error) {
	path = append([]string{cdglobal.DirUserAppData}, path...) // was CAIRO_DOCK_SHARE_DATA_DIR
	full := DirDockData(path...)
	dir := filepath.Dir(full)

	_, e := os.Stat(dir)
	if e != nil {
		e = os.MkdirAll(dir, 0700) // Create as private as it could contain passwords. TODO: Need to confirm.
	}

	return full, e
}

//
//------------------------------------------------------------[ MAIN OBJECTS ]--

// Maindock returns the primary main dock object.
//
func Maindock() *gldi.CairoDock {
	return gldi.NewDockFromNative(unsafe.Pointer(C.g_pMainDock))
}

// PrimaryContainer returns the main primary container object.
//
func PrimaryContainer() *gldi.Container {
	return gldi.NewContainerFromNative(unsafe.Pointer(C.g_pPrimaryContainer))
}

// themes-manager.c
// gchar *g_cExtrasDirPath = NULL;  // le chemin vers le repertoire des extra.
// gchar *g_cThemesDirPath = NULL;  // le chemin vers le repertoire des themes.

// gchar *g_cCurrentImagesPath = NULL;  // le chemin vers le repertoire des images ou autre du theme courant.
// gchar *g_cCurrentPlugInsPath = NULL;  // le chemin vers le repertoire des plug-ins du theme courant.

//
//------------------------------------------------------------[ GLOBAL FILES ]--

// FileCairoDockIcon gives the location of the cairo-dock icon.
//
func FileCairoDockIcon() string { return DirShareData(cdglobal.FileCairoDockIcon) }

//
//--------------------------------------------------------------[ ICON NAMES ]--

// Icon names.
const (
	IconNameAbout                = C.GLDI_ICON_NAME_ABOUT
	IconNameAdd                  = C.GLDI_ICON_NAME_ADD
	IconNameBold                 = C.GLDI_ICON_NAME_BOLD
	IconNameCapsLockWarning      = C.GLDI_ICON_NAME_CAPS_LOCK_WARNING
	IconNameCDROM                = C.GLDI_ICON_NAME_CDROM
	IconNameClear                = C.GLDI_ICON_NAME_CLEAR
	IconNameClose                = C.GLDI_ICON_NAME_CLOSE
	IconNameConnect              = C.GLDI_ICON_NAME_CONNECT
	IconNameCopy                 = C.GLDI_ICON_NAME_COPY
	IconNameCut                  = C.GLDI_ICON_NAME_CUT
	IconNameDelete               = C.GLDI_ICON_NAME_DELETE
	IconNameDialogAuthentication = C.GLDI_ICON_NAME_DIALOG_AUTHENTICATION
	IconNameDialogInfo           = C.GLDI_ICON_NAME_DIALOG_INFO
	IconNameDialogWarning        = C.GLDI_ICON_NAME_DIALOG_WARNING
	IconNameDialogError          = C.GLDI_ICON_NAME_DIALOG_ERROR
	IconNameDialogQuestion       = C.GLDI_ICON_NAME_DIALOG_QUESTION
	IconNameDirectory            = C.GLDI_ICON_NAME_DIRECTORY
	IconNameDisconnect           = C.GLDI_ICON_NAME_DISCONNECT
	IconNameEdit                 = C.GLDI_ICON_NAME_EDIT
	IconNameExecute              = C.GLDI_ICON_NAME_EXECUTE
	IconNameFile                 = C.GLDI_ICON_NAME_FILE
	IconNameFind                 = C.GLDI_ICON_NAME_FIND
	IconNameFindAndReplace       = C.GLDI_ICON_NAME_FIND_AND_REPLACE
	IconNameFloppy               = C.GLDI_ICON_NAME_FLOPPY
	IconNameFullScreen           = C.GLDI_ICON_NAME_FULLSCREEN
	IconNameGotoBottom           = C.GLDI_ICON_NAME_GOTO_BOTTOM
	IconNameGotoFirst            = C.GLDI_ICON_NAME_GOTO_FIRST
	IconNameGotoLast             = C.GLDI_ICON_NAME_GOTO_LAST
	IconNameGotoTop              = C.GLDI_ICON_NAME_GOTO_TOP
	IconNameGoBack               = C.GLDI_ICON_NAME_GO_BACK
	IconNameGoDown               = C.GLDI_ICON_NAME_GO_DOWN
	IconNameGoForward            = C.GLDI_ICON_NAME_GO_FORWARD
	IconNameGoUp                 = C.GLDI_ICON_NAME_GO_UP
	IconNameHardDisk             = C.GLDI_ICON_NAME_HARDDISK
	IconNameHelp                 = C.GLDI_ICON_NAME_HELP
	IconNameHome                 = C.GLDI_ICON_NAME_HOME
	IconNameIndent               = C.GLDI_ICON_NAME_INDENT
	IconNameInfo                 = C.GLDI_ICON_NAME_INFO
	IconNameItalic               = C.GLDI_ICON_NAME_ITALIC
	IconNameJumpTo               = C.GLDI_ICON_NAME_JUMP_TO
	IconNameJustifyCenter        = C.GLDI_ICON_NAME_JUSTIFY_CENTER
	IconNameJustifyFill          = C.GLDI_ICON_NAME_JUSTIFY_FILL
	IconNameJustifyLeft          = C.GLDI_ICON_NAME_JUSTIFY_LEFT
	IconNameJustifyRight         = C.GLDI_ICON_NAME_JUSTIFY_RIGHT
	IconNameLeaveFullScreen      = C.GLDI_ICON_NAME_LEAVE_FULLSCREEN
	IconNameMissingImage         = C.GLDI_ICON_NAME_MISSING_IMAGE
	IconNameMediaEject           = C.GLDI_ICON_NAME_MEDIA_EJECT
	IconNameMediaForward         = C.GLDI_ICON_NAME_MEDIA_FORWARD
	IconNameMediaNext            = C.GLDI_ICON_NAME_MEDIA_NEXT
	IconNameMediaPause           = C.GLDI_ICON_NAME_MEDIA_PAUSE
	IconNameMediaPlay            = C.GLDI_ICON_NAME_MEDIA_PLAY
	IconNameMediaPrevious        = C.GLDI_ICON_NAME_MEDIA_PREVIOUS
	IconNameMediaRecord          = C.GLDI_ICON_NAME_MEDIA_RECORD
	IconNameMediaRewind          = C.GLDI_ICON_NAME_MEDIA_REWIND
	IconNameMediaStop            = C.GLDI_ICON_NAME_MEDIA_STOP
	IconNameNetwork              = C.GLDI_ICON_NAME_NETWORK
	IconNameNew                  = C.GLDI_ICON_NAME_NEW
	IconNameOpen                 = C.GLDI_ICON_NAME_OPEN
	IconNameSetup                = C.GLDI_ICON_NAME_PAGE_SETUP
	IconNamePaste                = C.GLDI_ICON_NAME_PASTE
	IconNamePreferences          = C.GLDI_ICON_NAME_PREFERENCES
	IconNamePrint                = C.GLDI_ICON_NAME_PRINT
	IconNamePrintError           = C.GLDI_ICON_NAME_PRINT_ERROR
	IconNameProperties           = C.GLDI_ICON_NAME_PROPERTIES
	IconNameQuit                 = C.GLDI_ICON_NAME_QUIT
	IconNameRedo                 = C.GLDI_ICON_NAME_REDO
	IconNameRefresh              = C.GLDI_ICON_NAME_REFRESH
	IconNameRemove               = C.GLDI_ICON_NAME_REMOVE
	IconNameRevertToSaved        = C.GLDI_ICON_NAME_REVERT_TO_SAVED
	IconNameSave                 = C.GLDI_ICON_NAME_SAVE
	IconNameSaveAs               = C.GLDI_ICON_NAME_SAVE_AS
	IconNameSelectAll            = C.GLDI_ICON_NAME_SELECT_ALL
	IconNameSelectColor          = C.GLDI_ICON_NAME_SELECT_COLOR
	IconNameSelectFont           = C.GLDI_ICON_NAME_SELECT_FONT
	IconNameSortAscending        = C.GLDI_ICON_NAME_SORT_ASCENDING
	IconNameSortDescending       = C.GLDI_ICON_NAME_SORT_DESCENDING
	IconNameSpellCheck           = C.GLDI_ICON_NAME_SPELL_CHECK
	IconNameStop                 = C.GLDI_ICON_NAME_STOP
	IconNameUnderline            = C.GLDI_ICON_NAME_UNDERLINE
	IconNameUndo                 = C.GLDI_ICON_NAME_UNDO
	IconNameUnindent             = C.GLDI_ICON_NAME_UNINDENT
	IconNameZoom100              = C.GLDI_ICON_NAME_ZOOM_100
	IconNameZoomFit              = C.GLDI_ICON_NAME_ZOOM_FIT
	IconNameZoomIn               = C.GLDI_ICON_NAME_ZOOM_IN
	IconNameZoomOut              = C.GLDI_ICON_NAME_ZOOM_OUT
)
