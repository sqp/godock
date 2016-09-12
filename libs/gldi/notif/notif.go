// Package notif registers dock notifications.
package notif

/*
#cgo pkg-config: gldi

#include "cairo-dock-container.h"                  // NOTIFICATION_CLICK_ICON ...
#include "cairo-dock-desklet-manager.h"            // myDeskletObjectMgr
#include "cairo-dock-dock-manager.h"               // NOTIFICATION_ENTER_DOCK
#include "cairo-dock-keybinder.h"                  // NOTIFICATION_SHORTKEY_CHANGED
#include "cairo-dock-module-manager.h"             // myModuleObjectMgr
#include "cairo-dock-module-instance-manager.h"    // NOTIFICATION_MODULE_INSTANCE_DETACHED
#include "cairo-dock-windows-manager.h"            // myWindowObjectMgr


// Go exported func redeclarations.

extern gboolean onContainerLeftClick   (gpointer data, Icon *pClickedIcon, GldiContainer *pClickedContainer, guint iButtonState);
extern gboolean onContainerMiddleClick (gpointer data, Icon *pClickedIcon, GldiContainer *pClickedContainer);
extern gboolean onContainerMouseScroll (gpointer data, Icon *pClickedIcon, GldiContainer *pClickedContainer, int iDirection);
extern gboolean onContainerDropData    (gpointer data, gchar* cReceivedData, Icon *pClickedIcon, double fPosition, GldiContainer *pClickedContainer);
extern gboolean onContainerMenuIcon(gpointer, Icon*, GldiContainer*, GtkWidget*);
extern gboolean onContainerMenuContainer(gpointer, Icon*, GldiContainer*, GtkWidget*, gboolean);

extern gboolean onDeskletNew       (GldiObject *desklet);
extern gboolean onDeskletDestroy   (GldiObject *desklet);
extern gboolean onDeskletConfigure (G_GNUC_UNUSED gpointer pUserData, CairoDesklet *pDesklet);

extern gboolean onDockIconMoved    (gpointer pUserData, Icon *pIcon, CairoDock *pDock);
extern gboolean onDockInsertIcon   (gpointer pUserData, Icon *pIcon, CairoDock *pDock);
extern gboolean onDockRemoveIcon   (gpointer pUserData, Icon *pIcon, CairoDock *pDock);
extern gboolean onDockDestroy      (GldiObject *obj);

extern gboolean onModuleActivated  (G_GNUC_UNUSED gpointer pUserData, gchar *cModuleName, gboolean bActivated);
extern gboolean onModuleRegistered (G_GNUC_UNUSED gpointer pUserData, gchar *cModuleName, gboolean bActivated);

extern gboolean onModuleInstanceDetached (G_GNUC_UNUSED gpointer pUserData, GldiModuleInstance *pInstance, gboolean bIsDetached);

extern gboolean onShortkeyUpdate   (G_GNUC_UNUSED gpointer pUserData, G_GNUC_UNUSED gpointer shortkey);

extern gboolean onWindowChangeFocus (gpointer data, GldiWindowActor *window);

*/
import "C"

import (
	"github.com/gotk3/gotk3/gtk"

	"github.com/sqp/godock/libs/cdglobal" // Dock types.
	"github.com/sqp/godock/libs/gldi"
	"github.com/sqp/godock/libs/gldi/window" // Desktop windows control.
	"github.com/sqp/godock/widgets/gtk/togtk"

	"unsafe"
)

// Signal intercept or let pass.
const (
	// ActionLetPass is returned when the event is continued to be processed (C).
	ActionLetPass = C.GLDI_NOTIFICATION_LET_PASS

	// ActionLetPass is returned when the event is intercepted and blocked (C).
	ActionIntercept = C.GLDI_NOTIFICATION_INTERCEPT
)

const (
	// AnswerLetPass is returned when the event is continued to be processed.
	AnswerLetPass = false

	// AnswerIntercept is returned when the event is intercepted and blocked.
	AnswerIntercept = true
)

// RunType defines if an event is received before or after others.
type RunType int

// Event registration call time.
const (
	RunFirst RunType = C.GLDI_RUN_FIRST
	RunAfter RunType = C.GLDI_RUN_AFTER
)

//
//----------------------------------------------------------[ CALLBACK TYPES ]--

// OnContainerLeftClickFunc defines a container left click callback.
//
type OnContainerLeftClickFunc func(icon gldi.Icon, container *gldi.Container, btnState uint) bool

// OnContainerMiddleClickFunc defines a container middle click callback.
//
type OnContainerMiddleClickFunc func(icon gldi.Icon, container *gldi.Container) bool

// OnContainerMouseScrollFunc defines a container mouse scroll callback.
//
type OnContainerMouseScrollFunc func(icon gldi.Icon, container *gldi.Container, scrollUp bool) bool

// OnContainerDropDataFunc defines a container drop data callback.
//
type OnContainerDropDataFunc func(icon gldi.Icon, container *gldi.Container, data string, position float64) bool

// OnContainerMenuFunc defines a container menu callback.
//
type OnContainerMenuFunc func(gldi.Icon, *gldi.Container, *gldi.CairoDock, *gtk.Menu) bool

// OnDeskletFunc defines a desklet callback.
//
type OnDeskletFunc func(*gldi.Desklet)

// OnDockIconMovedFunc defines a dock icon moved callback.
//
type OnDockIconMovedFunc func(gldi.Icon, *gldi.CairoDock)

// OnDockInsertIconFunc defines a dock insert icon callback.
//
type OnDockInsertIconFunc func(gldi.Icon, *gldi.CairoDock)

// OnDockRemoveIconFunc defines a dock remove icon callback.
//
type OnDockRemoveIconFunc func(gldi.Icon, *gldi.CairoDock)

// OnDockDestroyFunc defines a dock destroy callback.
//
type OnDockDestroyFunc func(*gldi.CairoDock)

// OnModuleFunc defines a module callback.
//
type OnModuleFunc func(name string, active bool)

// OnModuleInstanceDetachedFunc defines a module instance detached callback.
//
type OnModuleInstanceDetachedFunc func(mi *gldi.ModuleInstance, isDetached bool)

// OnShortkeyFunc defines a shortkey callback.
//
type OnShortkeyFunc func()

// OnWindowChangeFocusFunc defines a window focus callback.
//
type OnWindowChangeFocusFunc func(cdglobal.Window) bool

//
//------------------------------------------------------[ REGISTER CONTAINER ]--

// MISSING
// // Notification called when the user double-clicks on an icon. data : {Icon, CairoDock}
// containerDoubleClickIcon int = C.NOTIFICATION_DOUBLE_CLICK_ICON
// // Notification called when the mouse enters an icon. data : {Icon, CairoDock, gboolean*}
// containerEnterIcon int = C.NOTIFICATION_ENTER_ICON
// // Notification called when the mouse enters a dock while dragging an object.
// containerStartDragData int = C.NOTIFICATION_START_DRAG_DATA
// // Notification called when the mouse has moved inside a container.
// containerMouseMoved int = C.NOTIFICATION_MOUSE_MOVED
// // Notification called when a key is pressed in a container that has the focus.
// containerKeyPressed int = C.NOTIFICATION_KEY_PRESSED
// // Notification called for the fast rendering loop on a container.
// containerUpdate int = C.NOTIFICATION_UPDATE
// // Notification called for the slow rendering loop on a container.
// containerUpdateSlow int = C.NOTIFICATION_UPDATE_SLOW
// // Notification called when a container is rendered.
// containerRender int = C.NOTIFICATION_RENDER

// RegisterContainerLeftClick registers notification called when use clicks on an icon
//
func RegisterContainerLeftClick(call OnContainerLeftClickFunc) {
	// data : {Icon, CairoDock, int}
	if len(eventsContainerLeftClick) == 0 {
		mgrContainer.register(C.NOTIFICATION_CLICK_ICON, C.onContainerLeftClick, RunAfter)
	}
	eventsContainerLeftClick = append(eventsContainerLeftClick, call)
}

// RegisterContainerMiddleClick registers notification called when the user middle-clicks on an icon.
//
func RegisterContainerMiddleClick(call OnContainerMiddleClickFunc) {
	// data : {Icon, CairoDock}
	if len(eventsContainerMiddleClick) == 0 {
		mgrContainer.register(C.NOTIFICATION_MIDDLE_CLICK_ICON, C.onContainerMiddleClick, RunAfter)
	}
	eventsContainerMiddleClick = append(eventsContainerMiddleClick, call)
}

// RegisterContainerMouseScroll registers notification called when the user scrolls on an icon.
//
func RegisterContainerMouseScroll(call OnContainerMouseScrollFunc) {
	// data : {Icon, CairoDock, int}
	if len(eventsContainerMouseScroll) == 0 {
		mgrContainer.register(C.NOTIFICATION_SCROLL_ICON, C.onContainerMouseScroll, RunAfter)
	}
	eventsContainerMouseScroll = append(eventsContainerMouseScroll, call)
}

// RegisterContainerDropData registers notification called when something is dropped inside a container.
//
func RegisterContainerDropData(call OnContainerDropDataFunc) {
	// data : {gchar*, Icon, double*, CairoDock}
	if len(eventsContainerDropData) == 0 {
		mgrContainer.register(C.NOTIFICATION_DROP_DATA, C.onContainerDropData, RunAfter)
	}
	eventsContainerDropData = append(eventsContainerDropData, call)
}

// RegisterContainerMenuContainer registers notification called when the menu is being built on a container.
//
func RegisterContainerMenuContainer(call OnContainerMenuFunc) {
	// data : {Icon, GldiContainer, GtkMenu, gboolean*}
	if len(eventsContainerMenuContainer) == 0 {
		mgrContainer.register(C.NOTIFICATION_BUILD_CONTAINER_MENU, C.onContainerMenuContainer, RunFirst)
	}
	eventsContainerMenuContainer = append(eventsContainerMenuContainer, call)
}

// RegisterContainerMenuIcon registers notification called when the menu is being built on an icon (possibly NULL).
//
func RegisterContainerMenuIcon(call OnContainerMenuFunc) {
	// data : {Icon, GldiContainer, GtkMenu}
	if len(eventsContainerMenuIcon) == 0 {
		mgrContainer.register(C.NOTIFICATION_BUILD_ICON_MENU, C.onContainerMenuIcon, RunAfter)
	}
	eventsContainerMenuIcon = append(eventsContainerMenuIcon, call)
}

//
//--------------------------------------------------------[ REGISTER DESKLET ]--

// MISSING
// // Notification called when the mouse enters a desklet.
// deskletEnter int = C.NOTIFICATION_ENTER_DESKLET
// // Notification called when the mouse leave a desklet.
// deskletLeave int = C.NOTIFICATION_LEAVE_DESKLET

// RegisterDeskletNew registers notification called when a desklet is created.
//
func RegisterDeskletNew(call OnDeskletFunc) {
	if len(eventsDeskletNew) == 0 {
		mgrDesklet.register(C.NOTIFICATION_NEW, C.onDeskletNew, RunAfter)
	}
	eventsDeskletNew = append(eventsDeskletNew, call)
}

// RegisterDeskletDestroy registers notification called when a desklet is going to be destroyed.
//
func RegisterDeskletDestroy(call OnDeskletFunc) {
	// data : the object
	if len(eventsDeskletDestroy) == 0 {
		mgrDesklet.register(C.NOTIFICATION_DESTROY, C.onDeskletDestroy, RunAfter)
	}
	eventsDeskletDestroy = append(eventsDeskletDestroy, call)
}

// RegisterDeskletConfigure registers notification called when a desklet is resized or moved on the screen.
//
func RegisterDeskletConfigure(call OnDeskletFunc) {
	if len(eventsDeskletConfigure) == 0 {
		mgrDesklet.register(C.NOTIFICATION_CONFIGURE_DESKLET, C.onDeskletConfigure, RunAfter)
	}
	eventsDeskletConfigure = append(eventsDeskletConfigure, call)
}

//
//-----------------------------------------------------------[ REGISTER DOCK ]--

// MISSING
// // Notification called when the mouse enters a dock.
// dockEnterDock int = C.NOTIFICATION_ENTER_DOCK
// // Notification called when the mouse leave a dock.
// dockLeaveDock int = C.NOTIFICATION_LEAVE_DOCK

// RegisterDockIconMoved registers notification called when an icon is moved inside a dock.
//
func RegisterDockIconMoved(call OnDockIconMovedFunc) {
	// data : {Icon, CairoDock}
	if len(eventsDockIconMoved) == 0 {
		mgrDock.register(C.NOTIFICATION_ICON_MOVED, C.onDockIconMoved, RunAfter)
	}
	eventsDockIconMoved = append(eventsDockIconMoved, call)
}

// RegisterDockInsertIcon registers notification called when an icon has just been inserted into a dock.
//
func RegisterDockInsertIcon(call OnDockInsertIconFunc) {
	// data : {Icon, CairoDock}
	if len(eventsDockInsertIcon) == 0 {
		mgrDock.register(C.NOTIFICATION_INSERT_ICON, C.onDockInsertIcon, RunAfter)
	}
	eventsDockInsertIcon = append(eventsDockInsertIcon, call)
}

// RegisterDockRemoveIcon registers notification called when an icon is going to be removed from a dock.
//
func RegisterDockRemoveIcon(call OnDockRemoveIconFunc) {
	// data : {Icon, CairoDock}
	if len(eventsDockRemoveIcon) == 0 {
		mgrDock.register(C.NOTIFICATION_REMOVE_ICON, C.onDockRemoveIcon, RunAfter)
	}
	eventsDockRemoveIcon = append(eventsDockRemoveIcon, call)
}

// RegisterDockDestroy registers notification called when a dock is going to be destroyed.
//
func RegisterDockDestroy(call OnDockDestroyFunc) {
	// data : the object
	if len(eventsDockDestroy) == 0 {
		mgrDock.register(C.NOTIFICATION_DESTROY, C.onDockDestroy, RunAfter)
	}
	eventsDockDestroy = append(eventsDockDestroy, call)
}

//
//---------------------------------------------------------[ REGISTER MODULE ]--

// MISSING
// moduleLogout int = C.NOTIFICATION_LOGOUT

// RegisterModuleActivated registers notification called when a module is activated.
//
func RegisterModuleActivated(call OnModuleFunc) {
	if len(eventsModuleActivated) == 0 {
		mgrModule.register(C.NOTIFICATION_MODULE_ACTIVATED, C.onModuleActivated, RunAfter)
	}
	eventsModuleActivated = append(eventsModuleActivated, call)
}

// RegisterModuleRegistered registers notification called when a module is registered.
//
func RegisterModuleRegistered(call OnModuleFunc) {
	if len(eventsModuleRegistered) == 0 {
		mgrModule.register(C.NOTIFICATION_MODULE_REGISTERED, C.onModuleRegistered, RunAfter)
	}
	eventsModuleRegistered = append(eventsModuleRegistered, call)
}

//
//------------------------------------------------[ REGISTER MODULE INSTANCE ]--

// RegisterModuleInstanceDetached registers notification called when a module instance is detached.
//
func RegisterModuleInstanceDetached(call OnModuleInstanceDetachedFunc) {
	if len(eventsModuleInstanceDetached) == 0 {
		mgrModuleInstance.register(C.NOTIFICATION_MODULE_INSTANCE_DETACHED, C.onModuleInstanceDetached, RunAfter)
	}
	eventsModuleInstanceDetached = append(eventsModuleInstanceDetached, call)
}

//
//-------------------------------------------------------[ REGISTER SHORTKEY ]--

// RegisterShortkeyChanged registers notification called when shortkeys are added, removed or changed.
//
func RegisterShortkeyChanged(call OnShortkeyFunc) {
	if len(eventsShortkeyChanged) == 0 {
		mgrShortkey.register(C.NOTIFICATION_NEW, C.onShortkeyUpdate, RunAfter)
		mgrShortkey.register(C.NOTIFICATION_DESTROY, C.onShortkeyUpdate, RunAfter)
		mgrShortkey.register(C.NOTIFICATION_SHORTKEY_CHANGED, C.onShortkeyUpdate, RunAfter)
	}
	eventsShortkeyChanged = append(eventsShortkeyChanged, call)
}

//
//---------------------------------------------------------[ REGISTER WINDOW ]--

// MISSING
// windowCreated             int = C.NOTIFICATION_WINDOW_CREATED
// windowDestroyed           int = C.NOTIFICATION_WINDOW_DESTROYED
// windowNameChanged         int = C.NOTIFICATION_WINDOW_NAME_CHANGED
// windowIconChanged         int = C.NOTIFICATION_WINDOW_ICON_CHANGED
// windowAttentionChanged    int = C.NOTIFICATION_WINDOW_ATTENTION_CHANGED
// windowSizePositionChanged int = C.NOTIFICATION_WINDOW_SIZE_POSITION_CHANGED
// windowStateChanged        int = C.NOTIFICATION_WINDOW_STATE_CHANGED
// windowClassChanged        int = C.NOTIFICATION_WINDOW_CLASS_CHANGED
// windowZOrderChanged       int = C.NOTIFICATION_WINDOW_Z_ORDER_CHANGED
// windowDesktopChanged      int = C.NOTIFICATION_WINDOW_DESKTOP_CHANGED

// RegisterWindowChangeFocus registers notification called when the focused window changed.
//
func RegisterWindowChangeFocus(call OnWindowChangeFocusFunc) {
	if len(eventsWindowChangeFocus) == 0 {
		mgrWindow.register(C.NOTIFICATION_WINDOW_ACTIVATED, C.onWindowChangeFocus, RunAfter)
	}
	eventsWindowChangeFocus = append(eventsWindowChangeFocus, call)
}

//
//-------------------------------------------------------------[ C CALLBACKS ]--

//export onContainerLeftClick
func onContainerLeftClick(data C.gpointer, pClickedIcon *C.Icon, pClickedContainer *C.GldiContainer, iButtonState C.guint) C.gboolean {
	icon := gldi.NewIconFromNative(unsafe.Pointer(pClickedIcon))
	container := gldi.NewContainerFromNative(unsafe.Pointer(pClickedContainer))
	for _, call := range eventsContainerLeftClick {
		ans := call(icon, container, uint(iButtonState))
		if ans == AnswerIntercept {
			return ActionIntercept
		}
	}
	return ActionLetPass
}

//export onContainerMiddleClick
func onContainerMiddleClick(data C.gpointer, pClickedIcon *C.Icon, pClickedContainer *C.GldiContainer) C.gboolean {
	icon := gldi.NewIconFromNative(unsafe.Pointer(pClickedIcon))
	container := gldi.NewContainerFromNative(unsafe.Pointer(pClickedContainer))
	for _, call := range eventsContainerMiddleClick {
		ans := call(icon, container)
		if ans == AnswerIntercept {
			return ActionIntercept
		}
	}
	return ActionLetPass
}

//export onContainerMouseScroll
func onContainerMouseScroll(data C.gpointer, pClickedIcon *C.Icon, pClickedContainer *C.GldiContainer, iDirection C.int) C.gboolean {
	icon := gldi.NewIconFromNative(unsafe.Pointer(pClickedIcon))
	container := gldi.NewContainerFromNative(unsafe.Pointer(pClickedContainer))
	for _, call := range eventsContainerMouseScroll {
		ans := call(icon, container, iDirection == C.GDK_SCROLL_UP)
		if ans == AnswerIntercept {
			return ActionIntercept
		}
	}
	return ActionLetPass
}

//export onContainerDropData
func onContainerDropData(data C.gpointer, cReceivedData *C.gchar, pClickedIcon *C.Icon, fPosition C.double, pClickedContainer *C.GldiContainer) C.gboolean {
	icon := gldi.NewIconFromNative(unsafe.Pointer(pClickedIcon))
	container := gldi.NewContainerFromNative(unsafe.Pointer(pClickedContainer))
	str := C.GoString((*C.char)(cReceivedData)) // don't free it's a const.
	for _, call := range eventsContainerDropData {
		ans := call(icon, container, str, float64(fPosition))
		if ans == AnswerIntercept {
			return ActionIntercept
		}
	}
	return ActionLetPass
}

//export onWindowChangeFocus
func onWindowChangeFocus(data C.gpointer, cwin *C.GldiWindowActor) C.gboolean {
	win := window.NewFromNative(unsafe.Pointer(cwin))
	for _, call := range eventsWindowChangeFocus {
		ans := call(win)
		if ans == AnswerIntercept {
			return ActionIntercept
		}
	}
	return ActionLetPass
}

//export onContainerMenuContainer
func onContainerMenuContainer(_ C.gpointer, ic *C.Icon, cont *C.GldiContainer, cmenu *C.GtkWidget, _ C.gboolean) C.gboolean {
	for _, call := range eventsContainerMenuContainer {
		ans := call(menuConvert(ic, cont, cmenu))
		if ans == AnswerIntercept {
			return ActionIntercept
		}
	}
	return ActionLetPass
}

//export onContainerMenuIcon
func onContainerMenuIcon(_ C.gpointer, ic *C.Icon, cont *C.GldiContainer, cmenu *C.GtkWidget) C.gboolean {
	for _, call := range eventsContainerMenuIcon {
		ans := call(menuConvert(ic, cont, cmenu))
		if ans == AnswerIntercept {
			return ActionIntercept
		}
	}
	return ActionLetPass
}

func menuConvert(ic *C.Icon, cont *C.GldiContainer, cmenu *C.GtkWidget) (gldi.Icon, *gldi.Container, *gldi.CairoDock, *gtk.Menu) { // *backendmenu.DockMenu {
	icon := gldi.NewIconFromNative(unsafe.Pointer(ic))
	container := gldi.NewContainerFromNative(unsafe.Pointer(cont))

	var dock *gldi.CairoDock
	if gldi.ObjectIsDock(container) {
		dock = container.ToCairoDock()
	}
	return icon, container, dock, togtk.Menu(unsafe.Pointer(cmenu))
}

//
//-----------------------------------------------------[ DESKLET C CALLBACKS ]--

//export onDeskletNew
func onDeskletNew(obj *C.GldiObject) C.gboolean {
	desklet := gldi.NewDeskletFromNative(unsafe.Pointer(obj))
	for _, call := range eventsDeskletNew {
		call(desklet)
	}
	return ActionLetPass
}

//export onDeskletDestroy
func onDeskletDestroy(obj *C.GldiObject) C.gboolean {
	desklet := gldi.NewDeskletFromNative(unsafe.Pointer(obj))
	for _, call := range eventsDeskletDestroy {
		call(desklet)
	}
	return ActionLetPass
}

//export onDeskletConfigure
func onDeskletConfigure(_ C.gpointer, cdesklet *C.CairoDesklet) C.gboolean {
	desklet := gldi.NewDeskletFromNative(unsafe.Pointer(cdesklet))
	for _, call := range eventsDeskletConfigure {
		call(desklet)
	}
	return ActionLetPass
}

//
//--------------------------------------------------------[ DOCK C CALLBACKS ]--

// TODO: maybe need to use the next func, I copied the exact behavior but need to report to confirm this func.
//export onDockIconMoved
func onDockIconMoved(_ C.gpointer, cicon *C.Icon, cdock *C.CairoDock) C.gboolean {
	icon := gldi.NewIconFromNative(unsafe.Pointer(cicon))
	dock := gldi.NewDockFromNative(unsafe.Pointer(cdock))
	for _, call := range eventsDockIconMoved {
		call(icon, dock)
	}
	return ActionLetPass
}

//export onDockInsertIcon
func onDockInsertIcon(_ C.gpointer, cicon *C.Icon, cdock *C.CairoDock) C.gboolean {
	icon := gldi.NewIconFromNative(unsafe.Pointer(cicon))
	dock := gldi.NewDockFromNative(unsafe.Pointer(cdock))
	for _, call := range eventsDockInsertIcon {
		call(icon, dock)
	}
	return ActionLetPass
}

//export onDockRemoveIcon
func onDockRemoveIcon(_ C.gpointer, cicon *C.Icon, cdock *C.CairoDock) C.gboolean {
	icon := gldi.NewIconFromNative(unsafe.Pointer(cicon))
	dock := gldi.NewDockFromNative(unsafe.Pointer(cdock))
	for _, call := range eventsDockRemoveIcon {
		call(icon, dock)
	}
	return ActionLetPass
}

//export onDockDestroy
func onDockDestroy(obj *C.GldiObject) C.gboolean {
	dock := gldi.NewDockFromNative(unsafe.Pointer(obj))
	for _, call := range eventsDockDestroy {
		call(dock)
	}
	return ActionLetPass
}

//export onModuleActivated
func onModuleActivated(_ C.gpointer, cModuleName *C.gchar, active C.gboolean) C.gboolean {
	name := C.GoString((*C.char)(cModuleName))
	for _, call := range eventsModuleActivated {
		call(name, active > 0)
	}
	return ActionLetPass
}

//export onModuleRegistered
func onModuleRegistered(_ C.gpointer, cModuleName *C.gchar, active C.gboolean) C.gboolean {
	name := C.GoString((*C.char)(cModuleName))
	for _, call := range eventsModuleRegistered {
		call(name, active > 0)
	}
	return ActionLetPass
}

//export onModuleInstanceDetached
func onModuleInstanceDetached(_ C.gpointer, instance *C.GldiModuleInstance, isDetached C.gboolean) C.gboolean {
	mi := gldi.NewModuleInstanceFromNative(unsafe.Pointer(instance))
	for _, call := range eventsModuleInstanceDetached {
		call(mi, isDetached > 0)
	}
	return ActionLetPass
}

//export onShortkeyUpdate
func onShortkeyUpdate(_ C.gpointer, _ C.gpointer) C.gboolean {
	for _, call := range eventsShortkeyChanged {
		call()
	}
	return ActionLetPass
}

//
//------------------------------------------------------------[ GO CALLBACKS ]--

// Registered events callbacks.
//
var (
	eventsContainerLeftClick     []OnContainerLeftClickFunc
	eventsContainerMiddleClick   []OnContainerMiddleClickFunc
	eventsContainerMouseScroll   []OnContainerMouseScrollFunc
	eventsContainerDropData      []OnContainerDropDataFunc
	eventsContainerMenuContainer []OnContainerMenuFunc
	eventsContainerMenuIcon      []OnContainerMenuFunc

	eventsDockIconMoved  []OnDockIconMovedFunc
	eventsDockInsertIcon []OnDockInsertIconFunc
	eventsDockRemoveIcon []OnDockRemoveIconFunc
	eventsDockDestroy    []OnDockDestroyFunc

	eventsDeskletNew       []OnDeskletFunc
	eventsDeskletDestroy   []OnDeskletFunc
	eventsDeskletConfigure []OnDeskletFunc

	eventsModuleActivated  []OnModuleFunc
	eventsModuleRegistered []OnModuleFunc

	eventsModuleInstanceDetached []OnModuleInstanceDetachedFunc

	eventsShortkeyNew     []OnShortkeyFunc
	eventsShortkeyDestroy []OnShortkeyFunc
	eventsShortkeyChanged []OnShortkeyFunc

	eventsWindowChangeFocus []OnWindowChangeFocusFunc
)

//
//-----------------------------------------------------------------[ MANAGER ]--

// Dock objects managers.
//
var (
	mgrContainer      = &objectManager{&C.myContainerObjectMgr}
	mgrDesklet        = &objectManager{&C.myDeskletObjectMgr}
	mgrDock           = &objectManager{&C.myDockObjectMgr}
	mgrModule         = &objectManager{&C.myModuleObjectMgr}
	mgrModuleInstance = &objectManager{&C.myModuleInstanceObjectMgr}
	mgrShortkey       = &objectManager{&C.myShortkeyObjectMgr}
	mgrWindow         = &objectManager{&C.myWindowObjectMgr}
)

// objectManager is a wrapper around a C dock manager.
//
type objectManager struct {
	Ptr *C.GldiObjectManager
}

func (o *objectManager) register(typ C.int, call unsafe.Pointer, run RunType) {
	C.gldi_object_register_notification(C.gpointer(o.Ptr),
		C.GldiNotificationType(typ),
		C.GldiNotificationFunc(call),
		C.gboolean(run), nil)
}

// Other sub events
// src/gldit/cairo-dock-flying-container.h:	NB_NOTIFICATIONS_FLYING_CONTAINER = NB_NOTIFICATIONS_CONTAINER
// src/gldit/cairo-dock-dialog-manager.h:	NB_NOTIFICATIONS_DIALOG = NB_NOTIFICATIONS_CONTAINER
