// Package backendevents provides access to the dock clicks and window events.
package backendevents

// #cgo pkg-config: gldi
// #include "cairo-dock-windows-manager.h"            // GldiWindowActor
/*

extern gboolean onLeftClick   (gpointer data, Icon *pClickedIcon, GldiContainer *pClickedContainer, guint iButtonState);
extern gboolean onMiddleClick (gpointer data, Icon *pClickedIcon, GldiContainer *pClickedContainer);
extern gboolean onMouseScroll (gpointer data, Icon *pClickedIcon, GldiContainer *pClickedContainer, int iDirection);
extern gboolean onDropData    (gpointer data, gchar* cReceivedData, Icon *pClickedIcon, double fPosition, GldiContainer *pClickedContainer);
extern gboolean onChangeFocus (gpointer data, GldiWindowActor *window);

*/
import "C"

import (
	"github.com/sqp/godock/libs/gldi"
	"github.com/sqp/godock/libs/gldi/globals"

	"unsafe"
)

//
//---------------------------------------------------------[ REGISTER EVENTS ]--

var events = newEventForward()

// Eventser defines events to register on the dock interface.
//
type Eventser interface {
	OnLeftClick(icon *gldi.Icon, container *gldi.Container, btnState uint) bool
	OnMiddleClick(icon *gldi.Icon, container *gldi.Container) bool
	OnMouseScroll(icon *gldi.Icon, container *gldi.Container, scrollUp bool) bool
	OnDropData(icon *gldi.Icon, container *gldi.Container, data string) bool
	OnChangeFocus(*gldi.WindowActor) bool
}

// Register registers an Eventser to receive dock events.
//
func Register(recall Eventser) {
	if len(events.backends) == 0 {
		C.gldi_object_register_notification(C.gpointer(globals.ContainerObjectMgr.Ptr),
			C.GldiNotificationType(globals.NotifClickIcon),
			C.GldiNotificationFunc(unsafe.Pointer(C.onLeftClick)),
			C.gboolean(globals.RunFirst), nil)

		C.gldi_object_register_notification(C.gpointer(globals.ContainerObjectMgr.Ptr),
			C.GldiNotificationType(globals.NotifMiddleClickIcon),
			C.GldiNotificationFunc(unsafe.Pointer(C.onMiddleClick)),
			C.gboolean(globals.RunFirst), nil)

		C.gldi_object_register_notification(C.gpointer(globals.ContainerObjectMgr.Ptr),
			C.GldiNotificationType(globals.NotifScrollIcon),
			C.GldiNotificationFunc(unsafe.Pointer(C.onMouseScroll)),
			C.gboolean(globals.RunFirst), nil)

		C.gldi_object_register_notification(C.gpointer(globals.ContainerObjectMgr.Ptr),
			C.GldiNotificationType(globals.NotifDropData),
			C.GldiNotificationFunc(unsafe.Pointer(C.onDropData)),
			C.gboolean(globals.RunFirst), nil) // was first to intercept dropped net packages.

		C.gldi_object_register_notification(C.gpointer(globals.WindowObjectMgr.Ptr),
			C.GldiNotificationType(globals.NotifWindowActivated),
			C.GldiNotificationFunc(unsafe.Pointer(C.onChangeFocus)),
			C.gboolean(globals.RunAfter), nil)
	}

	events.backends = append(events.backends, recall)
}

//
//----------------------------------------------------------[ FAN OUT EVENTS ]--

type eventForward struct {
	backends []Eventser
}

func newEventForward() *eventForward {
	return &eventForward{}
}

func (o *eventForward) onLeftClick(icon *gldi.Icon, container *gldi.Container, btnState uint) bool {
	for _, caller := range o.backends {
		b := caller.OnLeftClick(icon, container, btnState)
		if b {
			return true
		}
	}
	return false
}

func (o *eventForward) onMiddleClick(icon *gldi.Icon, container *gldi.Container) bool {
	for _, caller := range o.backends {
		b := caller.OnMiddleClick(icon, container)
		if b {
			return true
		}
	}
	return false
}

func (o *eventForward) onMouseScroll(icon *gldi.Icon, container *gldi.Container, scrollUp bool) bool {
	for _, caller := range o.backends {
		b := caller.OnMouseScroll(icon, container, scrollUp)
		if b {
			return true
		}
	}
	return false
}

func (o *eventForward) onDropData(icon *gldi.Icon, container *gldi.Container, data string) bool {
	for _, caller := range o.backends {
		b := caller.OnDropData(icon, container, data)
		if b {
			return true
		}
	}
	return false
}

func (o *eventForward) onChangeFocus(win *gldi.WindowActor) bool {
	for _, caller := range o.backends {
		b := caller.OnChangeFocus(win)
		if b {
			return true
		}
	}
	return false
}

//
//-------------------------------------------------------------[ C CALLBACKS ]--

//export onLeftClick
func onLeftClick(data C.gpointer, pClickedIcon *C.Icon, pClickedContainer *C.GldiContainer, iButtonState C.guint) C.gboolean {
	icon := gldi.NewIconFromNative(unsafe.Pointer(pClickedIcon))
	container := gldi.NewContainerFromNative(unsafe.Pointer(pClickedContainer))
	return cbool(events.onLeftClick(icon, container, uint(iButtonState)))
}

//export onMiddleClick
func onMiddleClick(data C.gpointer, pClickedIcon *C.Icon, pClickedContainer *C.GldiContainer) C.gboolean {
	icon := gldi.NewIconFromNative(unsafe.Pointer(pClickedIcon))
	container := gldi.NewContainerFromNative(unsafe.Pointer(pClickedContainer))
	return cbool(events.onMiddleClick(icon, container))
}

//export onMouseScroll
func onMouseScroll(data C.gpointer, pClickedIcon *C.Icon, pClickedContainer *C.GldiContainer, iDirection C.int) C.gboolean {
	icon := gldi.NewIconFromNative(unsafe.Pointer(pClickedIcon))
	container := gldi.NewContainerFromNative(unsafe.Pointer(pClickedContainer))
	return cbool(events.onMouseScroll(icon, container, iDirection == C.GDK_SCROLL_UP))
}

//export onDropData
func onDropData(data C.gpointer, cReceivedData *C.gchar, pClickedIcon *C.Icon, fPosition C.double, pClickedContainer *C.GldiContainer) C.gboolean {
	icon := gldi.NewIconFromNative(unsafe.Pointer(pClickedIcon))
	container := gldi.NewContainerFromNative(unsafe.Pointer(pClickedContainer))
	str := C.GoString((*C.char)(cReceivedData)) // don't free it's a const.
	return cbool(events.onDropData(icon, container, str))
}

//export onChangeFocus
func onChangeFocus(data C.gpointer, window *C.GldiWindowActor) C.gboolean {
	win := gldi.NewWindowActorFromNative(unsafe.Pointer(window))
	return cbool(events.onChangeFocus(win))
}

//
//-----------------------------------------------------------------[ HELPERS ]--

func cbool(b bool) C.gboolean {
	if b {
		return C.gboolean(1)
	}
	return C.gboolean(0)
}
