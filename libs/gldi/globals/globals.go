// Package globals provides access to the dock global variables.
package globals

/*
#cgo pkg-config: gldi

#include "cairo-dock-applications-manager.h"       // myTaskbarParam
#include "cairo-dock-container.h"                  // NOTIFICATION_CLICK_ICON ...
#include "cairo-dock-desklet-manager.h"            // myDeskletObjectMgr
#include "cairo-dock-dock-manager.h"               // myDocksParam
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
	"github.com/sqp/godock/libs/gldi"

	"os"
	"path/filepath"
	"unsafe"
)

const (
	CairoDockIcon = "cairo-dock.svg" // CAIRO_DOCK_ICON

	dirAppData = "appdata" // store user common applets data in ~/.cairo-dock/

	// #define CAIRO_DOCK_LOCAL_EXTRAS_DIR "extras"
	// #define CAIRO_DOCK_LAUNCHERS_DIR "launchers"
	// #define CAIRO_DOCK_LOCAL_ICONS_DIR "icons"
	// #define CAIRO_DOCK_LOCAL_IMAGES_DIR "images"

	DirPlugIns   = "plug-ins"  // CAIRO_DOCK_PLUG_INS_DIR
	DirAppletsGo = "goapplets" // go internal applets data files in /usr/share/cairo-dock/plug-ins
)

var (
	FullLock bool // Full settings lock from -k option at start. Config must not be changed if true.
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

func GtkVersion() (int, int, int) {
	return C.GTK_MAJOR_VERSION, C.GTK_MINOR_VERSION, C.GTK_MICRO_VERSION
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

func DirShareData(path ...string) string {
	path = append([]string{C.GLDI_SHARE_DATA_DIR}, path...) // was CAIRO_DOCK_SHARE_DATA_DIR
	return filepath.Join(path...)
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

// gchar *g_cCurrentImagesPath = NULL;  // le chemin vers le repertoire des images ou autre du theme courant.
// gchar *g_cCurrentPlugInsPath = NULL;  // le chemin vers le repertoire des plug-ins du theme courant.

//
//-------------------------------------------------------------[ DOCKS PARAM ]--

var DocksParam DocksParamType

type DocksParamType struct{}

func (DocksParamType) IsLockAll() bool {
	return gobool(C.myDocksParam.bLockAll)
}

func (DocksParamType) IsLockIcons() bool {
	return gobool(C.myDocksParam.bLockIcons)
}

func (DocksParamType) SetLockAll(b bool) {
	C.myDocksParam.bLockAll = cbool(b)
}

func (DocksParamType) SetLockIcons(b bool) {
	C.myDocksParam.bLockIcons = cbool(b)
}

func DockIsLocked() bool {
	return DocksParam.IsLockAll() || FullLock
}

//
//-----------------------------------------------------------[ TASKBAR PARAM ]--

// struct _CairoTaskbarParam {
// 	gboolean bShowAppli;
// 	gboolean bGroupAppliByClass;
// 	gint iAppliMaxNameLength;
// 	gboolean bMinimizeOnClick;
// 	gboolean bPresentClassOnClick;
// // 	gint iActionOnMiddleClick;
// 	gboolean bHideVisibleApplis;
// 	gdouble fVisibleAppliAlpha;
// 	gboolean bAppliOnCurrentDesktopOnly;
// 	gboolean bDemandsAttentionWithDialog;
// 	gint iDialogDuration;
// 	gchar *cAnimationOnDemandsAttention;
// 	gchar *cAnimationOnActiveWindow;
// 	gboolean bOverWriteXIcons;
// 	gint iMinimizedWindowRenderType;
// 	gboolean bMixLauncherAppli;
// 	gchar *cOverwriteException;
// 	gchar *cGroupException;
// 	gchar *cForceDemandsAttention;
// 	CairoTaskbarPlacement iIconPlacement;
// 	gchar *cRelativeIconName;
// 	gboolean bSeparateApplis;
// 	} ;

// var TaskbarParam TaskbarParamType

var TaskbarParam = &TaskbarParamType{&C.myTaskbarParam}

type TaskbarParamType struct {
	Ptr *C.CairoTaskbarParam
}

func (o *TaskbarParamType) ActionOnMiddleClick() int {
	return int(o.Ptr.iActionOnMiddleClick)
}

func (o *TaskbarParamType) OverWriteXIcons() bool {
	return gobool(o.Ptr.bOverWriteXIcons)
}

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

//
//-----------------------------------------------------------[ NOTIFICATIONS ]--

type NotifRun int

const (
	RunAfter NotifRun = C.GLDI_RUN_AFTER
	RunFirst NotifRun = C.GLDI_RUN_FIRST
)

// type NotifContainer int

// Notifications object (common).
const (
	/// notification called when an object has been created. data : the object
	NotifNew int = C.NOTIFICATION_NEW
	/// notification called when the object is going to be destroyed. data : the object
	NotifDestroy int = C.NOTIFICATION_DESTROY
)

// Notifications container.
const (
	/// notification called when the menu is being built on a container. data : {Icon, GldiContainer, GtkMenu, gboolean*}
	NotifBuildContainerMenu int = C.NOTIFICATION_BUILD_CONTAINER_MENU
	/// notification called when the menu is being built on an icon (possibly NULL). data : {Icon, GldiContainer, GtkMenu}
	NotifBuildIconMenu int = C.NOTIFICATION_BUILD_ICON_MENU
	/// notification called when use clicks on an icon data : {Icon, CairoDock, int}
	NotifClickIcon int = C.NOTIFICATION_CLICK_ICON
	/// notification called when the user double-clicks on an icon. data : {Icon, CairoDock}
	NotifDoubleClickIcon int = C.NOTIFICATION_DOUBLE_CLICK_ICON
	/// notification called when the user middle-clicks on an icon. data : {Icon, CairoDock}
	NotifMiddleClickIcon int = C.NOTIFICATION_MIDDLE_CLICK_ICON
	/// notification called when the user scrolls on an icon. data : {Icon, CairoDock, int}
	NotifScrollIcon int = C.NOTIFICATION_SCROLL_ICON
	/// notification called when the mouse enters an icon. data : {Icon, CairoDock, gboolean*}
	NotifEnterIcon int = C.NOTIFICATION_ENTER_ICON
	/// notification called when the mouse enters a dock while dragging an object.
	NotifStartDragData int = C.NOTIFICATION_START_DRAG_DATA
	/// notification called when something is dropped inside a container. data : {gchar*, Icon, double*, CairoDock}
	NotifDropData int = C.NOTIFICATION_DROP_DATA
	/// notification called when the mouse has moved inside a container.
	NotifMouseMoved int = C.NOTIFICATION_MOUSE_MOVED
	/// notification called when a key is pressed in a container that has the focus.
	NotifKeyPressed int = C.NOTIFICATION_KEY_PRESSED
	/// notification called for the fast rendering loop on a container.
	NotifUpdate int = C.NOTIFICATION_UPDATE
	/// notification called for the slow rendering loop on a container.
	NotifUpdateSlow int = C.NOTIFICATION_UPDATE_SLOW
	/// notification called when a container is rendered.
	NotifRender int = C.NOTIFICATION_RENDER
)

// Notifications Dock.
const (
	/// notification called when the mouse enters a dock.
	NotifEnterDock int = C.NOTIFICATION_ENTER_DOCK
	/// notification called when the mouse leave a dock.
	NotifLeaveDock int = C.NOTIFICATION_LEAVE_DOCK
	/// notification called when an icon has just been inserted into a dock. data : {Icon, CairoDock}
	NotifInsertIcon int = C.NOTIFICATION_INSERT_ICON
	/// notification called when an icon is going to be removed from a dock. data : {Icon, CairoDock}
	NotifRemoveIcon int = C.NOTIFICATION_REMOVE_ICON
	/// notification called when an icon is moved inside a dock. data : {Icon, CairoDock}
	NotifIconMoved int = C.NOTIFICATION_ICON_MOVED
)

// Notifications Desklet.
const (
	/// notification called when the mouse enters a desklet.
	NotifEnterDesklet int = C.NOTIFICATION_ENTER_DESKLET
	/// notification called when the mouse leave a desklet.
	NotifLeaveDesklet int = C.NOTIFICATION_LEAVE_DESKLET
	/// notification called when a desklet is resized or moved on the screen.
	NotifConfigureDesklet int = C.NOTIFICATION_CONFIGURE_DESKLET
)

// Notifications module.
const (
	NotifModuleRegistered int = C.NOTIFICATION_MODULE_REGISTERED
	NotifModuleActivated  int = C.NOTIFICATION_MODULE_ACTIVATED
	NotifLogout           int = C.NOTIFICATION_LOGOUT
)

// Notifications module instance.
const (
	NotifModuleInstanceDetached int = C.NOTIFICATION_MODULE_INSTANCE_DETACHED
)

// Notifications shortkey.
const (
	NotifShortkeyChanged int = C.NOTIFICATION_SHORTKEY_CHANGED
)

// Notifications window.
const (
	NotifWindowCreated             int = C.NOTIFICATION_WINDOW_CREATED
	NotifWindowDestroyed           int = C.NOTIFICATION_WINDOW_DESTROYED
	NotifWindowNameChanged         int = C.NOTIFICATION_WINDOW_NAME_CHANGED
	NotifWindowIconChanged         int = C.NOTIFICATION_WINDOW_ICON_CHANGED
	NotifWindowAttentionChanged    int = C.NOTIFICATION_WINDOW_ATTENTION_CHANGED
	NotifWindowSizePositionChanged int = C.NOTIFICATION_WINDOW_SIZE_POSITION_CHANGED
	NotifWindowStateChanged        int = C.NOTIFICATION_WINDOW_STATE_CHANGED
	NotifWindowClassChanged        int = C.NOTIFICATION_WINDOW_CLASS_CHANGED
	NotifWindowZOrderChanged       int = C.NOTIFICATION_WINDOW_Z_ORDER_CHANGED
	NotifWindowActivated           int = C.NOTIFICATION_WINDOW_ACTIVATED
	NotifWindowDesktopChanged      int = C.NOTIFICATION_WINDOW_DESKTOP_CHANGED
)

// Other sub events
// src/gldit/cairo-dock-flying-container.h:	NB_NOTIFICATIONS_FLYING_CONTAINER = NB_NOTIFICATIONS_CONTAINER
// src/gldit/cairo-dock-dialog-manager.h:	NB_NOTIFICATIONS_DIALOG = NB_NOTIFICATIONS_CONTAINER

//
//-----------------------------------------------------------------[ MANAGER ]--

type ObjectManager struct {
	Ptr *C.GldiObjectManager
}

func NewObjectManagerFromNative(p unsafe.Pointer) *ObjectManager {
	return &ObjectManager{(*C.GldiObjectManager)(p)}
}

func (o *ObjectManager) RegisterNotification(typ int, call unsafe.Pointer, run NotifRun) {
	C.gldi_object_register_notification(C.gpointer(o.Ptr),
		C.GldiNotificationType(typ),
		C.GldiNotificationFunc(call),
		C.gboolean(run), nil)
}

var ContainerObjectMgr = &ObjectManager{&C.myContainerObjectMgr}
var DeskletObjectMgr = &ObjectManager{&C.myDeskletObjectMgr}
var DockObjectMgr = &ObjectManager{&C.myDockObjectMgr}
var ModuleObjectMgr = &ObjectManager{&C.myModuleObjectMgr}
var ModuleInstanceObjectMgr = &ObjectManager{&C.myModuleInstanceObjectMgr}
var ShortkeyObjectMgr = &ObjectManager{&C.myShortkeyObjectMgr}
var WindowObjectMgr = &ObjectManager{&C.myWindowObjectMgr}

//
//-----------------------------------------------------------------[ HELPERS ]--

func cbool(b bool) C.gboolean {
	if b {
		return C.gboolean(1)
	}
	return C.gboolean(0)
}
func gobool(b C.gboolean) bool {
	if b == 1 {
		return true
	}
	return false
}
