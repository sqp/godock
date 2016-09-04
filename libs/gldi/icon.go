package gldi

/*
#cgo pkg-config: gldi
#include <stdlib.h>                              // free

#include "cairo-dock-animations.h"               // gldi_icon_request_animation
#include "cairo-dock-applications-manager.h"     // cairo_dock_get_appli_icon
#include "cairo-dock-applet-facility.h"          // cairo_dock_set_image_on_icon
#include "cairo-dock-backends-manager.h"         // gldi_dialogs_remove_on_icon
#include "cairo-dock-class-manager.h"            // cairo_dock_get_class_command
#include "cairo-dock-class-icon-manager.h"       // myClassIconObjectMgr
#include "cairo-dock-desktop-manager.h"          // gldi_desktop_present_class
#include "cairo-dock-data-renderer.h"            // cairo_dock_render_new_data_on_icon
#include "cairo-dock-dock-facility.h"            // cairo_dock_show_subdock
#include "cairo-dock-icon-facility.h"            // gldi_icon_set_name
#include "cairo-dock-launcher-manager.h"         // CAIRO_DOCK_ICON_TYPE_IS_LAUNCHER
#include "cairo-dock-overlay.h"                  // cairo_dock_add_overlay_from_image
#include "cairo-dock-module-manager.h"           // CAIRO_DOCK_MODULE_CAN_DESKLET
#include "cairo-dock-module-instance-manager.h"  // ModuleInstance
#include "cairo-dock-separator-manager.h"        // CAIRO_DOCK_ICON_TYPE_IS_SEPARATOR
#include "cairo-dock-stack-icon-manager.h"       // GLDI_OBJECT_IS_STACK_ICON


extern CairoDock *g_pMainDock;
extern gchar     *g_cCurrentLaunchersPath;


static gboolean IconIsSeparator    (Icon *icon) { return CAIRO_DOCK_ICON_TYPE_IS_SEPARATOR(icon); }
static gboolean IconIsSeparatorAuto(Icon *icon) { return CAIRO_DOCK_IS_AUTOMATIC_SEPARATOR(icon); }
static gboolean IconIsLauncher     (Icon *icon) { return CAIRO_DOCK_ICON_TYPE_IS_LAUNCHER(icon); }
static gboolean IconIsStackIcon    (Icon *icon) { return CAIRO_DOCK_ICON_TYPE_IS_CONTAINER(icon); }


// from cairo-dock-icon-facility.h
static Icon* _icons_get_any_without_dialog() {
	return gldi_icons_get_without_dialog (g_pMainDock?g_pMainDock->icons:NULL);
}

static void render_new_data_on_icon (Icon *pIcon, double *pNewValues) {
	GldiContainer *pContainer = pIcon->pModuleInstance->pContainer;

	cairo_t *pDrawContext = cairo_create (pIcon->image.pSurface);
	cairo_dock_render_new_data_on_icon (pIcon, pContainer, pDrawContext, pNewValues);
	cairo_destroy (pDrawContext);
}


*/
import "C"

import (
	"github.com/gotk3/gotk3/glib"

	"github.com/sqp/godock/libs/cdtype"            // Dock types.
	"github.com/sqp/godock/libs/gldi/desktopclass" // XDG desktop class info.
	"github.com/sqp/godock/libs/gldi/shortkeys"    // Keyboard shortkeys.
	"github.com/sqp/godock/libs/gldi/window"       // Desktop windows control.
	"github.com/sqp/godock/libs/ternary"           // Ternary operators.

	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"unsafe"
)

//
//---------------------------------------------------------------[ INTERFACE ]--

// Icon defines the dock icon interface.
//
type Icon interface {

	// ToNative returns the pointer to the native object.
	//
	ToNative() unsafe.Pointer

	// DefaultNameIcon returns improved name and image for the icon if possible.
	//
	DefaultNameIcon() (name, img string)
	HasClass() bool
	GetClass() desktopclass.Info
	GetName() string
	GetInitialName() string
	GetFileName() string
	GetDesktopFileName() string
	GetParentDockName() string
	GetCommand() string
	GetContainer() *Container

	// ConfigPath gives the full path to the icon config file.
	//
	ConfigPath() string
	GetIgnoreQuickList() bool

	X() float64
	Y() float64
	DrawX() float64
	DrawY() float64
	Width() float64
	Height() float64
	RequestedWidth() int
	RequestedHeight() int
	RequestedDisplayWidth() int
	RequestedDisplayHeight() int
	IsPointed() bool
	SetPointed(val bool)
	SetWidth(val float64)

	SetHeight(val float64)
	SetScale(val float64)
	InsertRemoveFactor() float64
	Scale() float64
	XAtRest() float64
	SetX(f float64)
	SetY(f float64)
	SetDrawX(f float64)
	SetDrawY(f float64)
	SetXAtRest(f float64)
	SetAllocatedSize(w, h int)
	SetWidthFactor(f float64)
	SetHeightFactor(f float64)
	SetOrientation(f float64)
	SetAlpha(f float64)
	Order() float64
	IconExtent() (int, int)
	IsApplet() bool

	// IsAppli returns whether the icon manages an application. CAIRO_DOCK_IS_APPLI
	//
	IsAppli() bool

	// IsClassIcon returns whether the icon .
	// GLDI_OBJECT_IS_CLASS_ICON / CAIRO_DOCK_ICON_TYPE_IS_CLASS_CONTAINER
	//
	IsClassIcon() bool

	//
	IsDetachableApplet() bool

	// IsMultiAppli returns whether the icon manages multiple applications. CAIRO_DOCK_IS_MULTI_APPLI
	//
	IsMultiAppli() bool

	// IsTaskbar returns whether the icon belongs to the taskbar or not.
	//
	IsTaskbar() bool
	IsSeparator() bool
	IsSeparatorAuto() bool
	IsLauncher() bool
	IsStackIcon() bool // CAIRO_DOCK_ICON_TYPE_IS_CONTAINER

	RemoveFromDock()
	WriteContainerNameInConfFile(newdock string)
	ModuleInstance() *ModuleInstance
	GetSubDock() *CairoDock

	// SubDockIsVisible returns whether the subdock is visible or not.
	//
	SubDockIsVisible() bool
	RemoveSubdockEmpty()

	// LaunchCommand starts the command associated with the icon.
	//
	LaunchCommand(log cdtype.Logger, args ...string) bool

	// ShowSubdock shows the icon subdock.
	//
	ShowSubdock(parentDock *CairoDock)

	// DesktopPresentClass can use scale applet to show class windows.
	//
	DesktopPresentClass() bool

	RemoveDialogs()

	IsDemandingAttention() bool

	RequestAttention(animation string, nbRounds int)

	StopAttention()

	ClassIsInhibited() bool

	InhibiteClass(class string)

	DeinhibiteClass()

	Window() window.Type

	// GetInhibitor returns the icon that inhibits the current one (has registered the class).
	//
	GetInhibitor(b bool) Icon

	RemoveIconsFromSubdock(dest *CairoDock)

	// TODO: may have to move.
	SubDockIcons() []Icon

	// SetLabel sets the label of an icon.
	// If it has a sub-dock, it is renamed (the name is possibly altered to stay unique).
	// The label buffer is updated too.
	//
	SetLabel(str string)

	// SetQuickInfo sets the quick-info of an icon.
	// This is a small text (a few characters) that is superimposed on the icon.
	//
	SetQuickInfo(str string)

	SetIcon(str string) error

	AddNewDataRenderer(attr DataRendererAttributer)

	AddDataRendererWithText(attr DataRendererAttributer, dataRendererText DataRendererTextFunc)

	Render(values ...float64) error

	DataRendererTextFuncPercent(values ...float64) []string

	RemoveDataRenderer()

	// AddOverlayFromImage adds an overlay on the icon.
	//
	AddOverlayFromImage(iconPath string, position cdtype.EmblemPosition)

	// RemoveOverlayAtPosition removes an overlay on the icon.
	//
	RemoveOverlayAtPosition(position cdtype.EmblemPosition)

	RequestAnimation(animation string, rounds int)

	Redraw()

	// MoveAfterIcon moves the icon position after the given icon.
	//
	MoveAfterIcon(container *CairoDock, target Icon)

	// CallbackActionWindow returns a func to use as gtk callback.
	// On event, it will test if the icon still has a valid window and launch the
	// provided action on this window.
	//
	CallbackActionWindow(call func(window.Type)) func()

	// CallbackActionSubWindows is the same as CallbackActionWindow but launch the
	// action on all subdock windows.
	//
	CallbackActionSubWindows(call func(window.Type)) func()

	CallbackActionWindowToggle(call func(window.Type, bool), getvalue func(window.Type) bool) func()

	GetPrevNextClassMateIcon(next bool) Icon
}

//
//------------------------------------------------------[ DATA RENDERER TEXT ]--

// DataRendererTextFunc defines a data renderer text format func.
//
type DataRendererTextFunc func(...float64) []string

//
//--------------------------------------------------------------------[ ICON ]--

// dockIcon defines a gldi icon.
//
type dockIcon struct {
	Ptr *C.Icon

	dataRendererText DataRendererTextFunc // optional data renderer (when set, this will replace the C data rendering)
}

// NewIconFromNative wraps a gldi icon from C pointer.
//
func NewIconFromNative(p unsafe.Pointer) Icon {
	if p == nil {
		return nil
	}
	return &dockIcon{
		Ptr:              (*C.Icon)(p),
		dataRendererText: nil,
	}
}

// CreateDummyLauncher creates an empty icon.
//
func CreateDummyLauncher(name, iconPath, command, quickinfo string, order float64) Icon {
	var qi *C.gchar
	if quickinfo != "" {
		qi = gchar(quickinfo)
	}

	c := C.cairo_dock_create_dummy_launcher(gchar(name), gchar(iconPath), gchar(command), qi, C.double(order))
	return NewIconFromNative(unsafe.Pointer(c))
}

// IconsGetAnyWithoutDialog finds an icon usable to display a dialog.
//
func IconsGetAnyWithoutDialog() Icon {
	return NewIconFromNative(unsafe.Pointer(C._icons_get_any_without_dialog()))
}

// GetAppliIcon returns the icon managing the window.
//
func GetAppliIcon(win window.Type) Icon {
	c := C.cairo_dock_get_appli_icon(win.ToNative())
	return NewIconFromNative(unsafe.Pointer(c))
}

func (o *dockIcon) ToNative() unsafe.Pointer      { return unsafe.Pointer(o.Ptr) }
func (o *dockIcon) ClassIsInhibited() bool        { return C.cairo_dock_class_is_inhibited(o.Ptr.cClass) > 0 }
func (o *dockIcon) DeinhibiteClass()              { C.cairo_dock_deinhibite_class(o.Ptr.cClass, o.Ptr) }
func (o *dockIcon) DesktopPresentClass() bool     { return C.gldi_desktop_present_class(o.Ptr.cClass) > 0 }
func (o *dockIcon) DrawX() float64                { return float64(o.Ptr.fDrawX) }
func (o *dockIcon) DrawY() float64                { return float64(o.Ptr.fDrawY) }
func (o *dockIcon) GetCommand() string            { return C.GoString((*C.char)(o.Ptr.cCommand)) }
func (o *dockIcon) GetDesktopFileName() string    { return C.GoString((*C.char)(o.Ptr.cDesktopFileName)) }
func (o *dockIcon) GetFileName() string           { return C.GoString((*C.char)(o.Ptr.cFileName)) }
func (o *dockIcon) GetIgnoreQuickList() bool      { return gobool(o.Ptr.bIgnoreQuicklist) }
func (o *dockIcon) GetInitialName() string        { return C.GoString((*C.char)(o.Ptr.cInitialName)) }
func (o *dockIcon) GetName() string               { return C.GoString((*C.char)(o.Ptr.cName)) }
func (o *dockIcon) GetParentDockName() string     { return C.GoString((*C.char)(o.Ptr.cParentDockName)) }
func (o *dockIcon) GetSubDock() *CairoDock        { return NewDockFromNative(unsafe.Pointer(o.Ptr.pSubDock)) }
func (o *dockIcon) HasClass() bool                { return o.Ptr.cClass != nil }
func (o *dockIcon) Height() float64               { return float64(o.Ptr.fHeight) }
func (o *dockIcon) InsertRemoveFactor() float64   { return float64(o.Ptr.fInsertRemoveFactor) }
func (o *dockIcon) IsApplet() bool                { return o.Ptr != nil && o.Ptr.pModuleInstance != nil }
func (o *dockIcon) IsAppli() bool                 { return o.Ptr != nil && o.Ptr.pAppli != nil }              // CAIRO_DOCK_IS_APPLI
func (o *dockIcon) IsClassIcon() bool             { return ObjectIsManagerChild(o, &C.myClassIconObjectMgr) } // GLDI_OBJECT_IS_CLASS_ICON / CAIRO_DOCK_ICON_TYPE_IS_CLASS_CONTAINER
func (o *dockIcon) IsDemandingAttention() bool    { return gobool(o.Ptr.bIsDemandingAttention) }
func (o *dockIcon) IsLauncher() bool              { return gobool(C.IconIsLauncher(o.Ptr)) }
func (o *dockIcon) IsPointed() bool               { return gobool(o.Ptr.bPointed) }
func (o *dockIcon) IsSeparator() bool             { return gobool(C.IconIsSeparator(o.Ptr)) }
func (o *dockIcon) IsSeparatorAuto() bool         { return gobool(C.IconIsSeparatorAuto(o.Ptr)) }
func (o *dockIcon) IsStackIcon() bool             { return gobool(C.IconIsStackIcon(o.Ptr)) } // CAIRO_DOCK_ICON_TYPE_IS_CONTAINER
func (o *dockIcon) IsTaskbar() bool               { return o.IsAppli() && !o.IsLauncher() && !o.IsApplet() }
func (o *dockIcon) Order() float64                { return float64(o.Ptr.fOrder) }
func (o *dockIcon) Redraw()                       { C.cairo_dock_redraw_icon(o.Ptr) }
func (o *dockIcon) RemoveDialogs()                { C.gldi_dialogs_remove_on_icon(o.Ptr) }
func (o *dockIcon) RemoveFromDock()               { C.cairo_dock_trigger_icon_removal_from_dock(o.Ptr) }
func (o *dockIcon) RequestedDisplayHeight() int   { return int(o.Ptr.iRequestedDisplayHeight) }
func (o *dockIcon) RequestedDisplayWidth() int    { return int(o.Ptr.iRequestedDisplayWidth) }
func (o *dockIcon) RequestedHeight() int          { return int(o.Ptr.iRequestedHeight) }
func (o *dockIcon) RequestedWidth() int           { return int(o.Ptr.iRequestedWidth) }
func (o *dockIcon) Scale() float64                { return float64(o.Ptr.fScale) }
func (o *dockIcon) SetAlpha(f float64)            { o.Ptr.fAlpha = C.gdouble(f) }
func (o *dockIcon) SetDrawX(f float64)            { o.Ptr.fDrawX = C.gdouble(f) }
func (o *dockIcon) SetDrawY(f float64)            { o.Ptr.fDrawY = C.gdouble(f) }
func (o *dockIcon) SetHeight(val float64)         { o.Ptr.fHeight = C.gdouble(val) }
func (o *dockIcon) SetHeightFactor(f float64)     { o.Ptr.fHeightFactor = C.gdouble(f) }
func (o *dockIcon) SetOrientation(f float64)      { o.Ptr.fOrientation = C.gdouble(f) }
func (o *dockIcon) SetPointed(val bool)           { o.Ptr.bPointed = cbool(val) }
func (o *dockIcon) SetScale(val float64)          { o.Ptr.fScale = C.gdouble(val) }
func (o *dockIcon) SetWidth(val float64)          { o.Ptr.fWidth = C.gdouble(val) }
func (o *dockIcon) SetWidthFactor(f float64)      { o.Ptr.fWidthFactor = C.gdouble(f) }
func (o *dockIcon) SetX(f float64)                { o.Ptr.fX = C.gdouble(f) }
func (o *dockIcon) SetXAtRest(f float64)          { o.Ptr.fXAtRest = C.gdouble(f) }
func (o *dockIcon) SetY(f float64)                { o.Ptr.fY = C.gdouble(f) }
func (o *dockIcon) ShowSubdock(parent *CairoDock) { C.cairo_dock_show_subdock(o.Ptr, parent.ToNative()) }
func (o *dockIcon) StopAttention()                { C.gldi_icon_stop_attention(o.Ptr) }
func (o *dockIcon) Width() float64                { return float64(o.Ptr.fWidth) }
func (o *dockIcon) Window() window.Type           { return window.NewFromNative(unsafe.Pointer(o.Ptr.pAppli)) }
func (o *dockIcon) X() float64                    { return float64(o.Ptr.fX) }
func (o *dockIcon) XAtRest() float64              { return float64(o.Ptr.fXAtRest) }
func (o *dockIcon) Y() float64                    { return float64(o.Ptr.fY) }

func (o *dockIcon) AddOverlayFromImage(iconPath string, position cdtype.EmblemPosition) {
	var cstr *C.gchar
	if iconPath != "" {
		cstr = (*C.gchar)(C.CString(iconPath))
		defer C.free(unsafe.Pointer((*C.char)(cstr)))
	}
	// last arg was 'myApplet' to identify the overlays set by the Dbus plug-in (since the plug-in can't be deactivated, 'myApplet' is constant).
	C.cairo_dock_add_overlay_from_image(o.Ptr, cstr, C.CairoOverlayPosition(position), C.gpointer(o.Ptr))
}

func (o *dockIcon) CallbackActionSubWindows(call func(window.Type)) func() {
	return func() {
		for _, ic := range o.SubDockIcons() {
			if ic.IsAppli() {
				call(ic.Window())
			}
		}
	}
}

func (o *dockIcon) CallbackActionWindow(call func(window.Type)) func() {
	return func() {
		if o.IsAppli() {
			call(o.Window())
		}
	}
}

func (o *dockIcon) CallbackActionWindowToggle(call func(window.Type, bool), getvalue func(window.Type) bool) func() {
	return o.CallbackActionWindow(func(win window.Type) {
		v := getvalue(win)
		call(win, !v)
	})
}

func (o *dockIcon) ConfigPath() string {
	switch {
	case o.IsApplet():
		return o.ModuleInstance().GetConfFilePath()

	case o.IsStackIcon(), o.IsLauncher() || o.IsSeparator():
		dir := C.GoString((*C.char)(C.g_cCurrentLaunchersPath))
		// dir := globals.CurrentLaunchersPath()
		return filepath.Join(dir, o.GetDesktopFileName())
	}
	return ""
}

func (o *dockIcon) DefaultNameIcon() (name, img string) {
	switch {
	case o.IsApplet():
		vc := o.ModuleInstance().Module().VisitCard()
		return vc.GetTitle(), vc.GetIconFilePath()

	case o.IsSeparator():
		return "--------", ""

	case o.IsLauncher(), o.IsStackIcon(), o.IsAppli(), o.IsClassIcon():
		name := o.GetClass().Name()
		if name != "" {
			return name, o.GetFileName() // o.GetClassInfo(ClassIcon)
		}
		return ternary.String(o.GetInitialName() != "", o.GetInitialName(), o.GetName()), o.GetFileName()

	}
	return o.GetName(), o.GetFileName()
}

func (o *dockIcon) GetClass() desktopclass.Info {
	return desktopclass.Info(C.GoString((*C.char)(o.Ptr.cClass)))
}

func (o *dockIcon) GetContainer() *Container {
	if o.Ptr == nil || o.Ptr.pContainer == nil {
		return nil
	}
	return NewContainerFromNative(unsafe.Pointer(o.Ptr.pContainer))
}

// func (o *dockIcon) GetIconType() int {
// 	return int(C.cairo_dock_get_icon_type(o.Ptr))
// }

func (o *dockIcon) GetInhibitor(b bool) Icon {
	c := C.cairo_dock_get_inhibitor(o.Ptr, cbool(b))
	return NewIconFromNative(unsafe.Pointer(c))
	// 			pNewActiveIcon = cairo_dock_get_inhibitor (pNewActiveIcon, FALSE);
}

func (o *dockIcon) GetPrevNextClassMateIcon(next bool) Icon {
	c := C.cairo_dock_get_prev_next_classmate_icon(o.ToNative(), cbool(next))
	return NewIconFromNative(unsafe.Pointer(c))
}

func (o *dockIcon) IconExtent() (int, int) {
	var width, height C.int
	C.cairo_dock_get_icon_extent(o.Ptr, &width, &height)
	return int(width), int(height)
}

func (o *dockIcon) InhibiteClass(class string) {
	cstr := (*C.gchar)(C.CString(class))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	C.cairo_dock_inhibite_class(cstr, o.Ptr)
}

//
func (o *dockIcon) IsDetachableApplet() bool {
	return o.IsApplet() &&
		o.ModuleInstance().Module().VisitCard().GetContainerType()&C.CAIRO_DOCK_MODULE_CAN_DESKLET > 0
}

func (o *dockIcon) IsMultiAppli() bool { // CAIRO_DOCK_IS_MULTI_APPLI
	return o.Ptr.pSubDock != nil &&
		(o.IsLauncher() ||
			o.IsClassIcon() ||
			(o.IsApplet() && o.GetClass() != ""))
}

func (o *dockIcon) LaunchCommand(log cdtype.Logger, args ...string) bool {
	cmd := o.GetCommand()
	switch {
	case cmd == "":

	case cmd[0] == '<': // Launch as shortkey.
		success := shortkeys.Trigger(cmd)
		if success {
			return true
		}

		// try also as exec.
		exec, e := log.ExecShlex(cmd, args...)
		if log.Err(e, "parse command", cmd) {
			return false
		}

		e = exec.Run()
		return e == nil

	default: // Exec command.
		exec, e := log.ExecShlex(cmd, args...)
		if log.Err(e, "parse command", cmd) {
			return false
		}

		e = exec.Run()
		if log.Err(e, "launch command", cmd, strings.Join(args, " ")) {
			// try also as shortkey.
			return shortkeys.Trigger(cmd)
		}

		return true
	}

	return false
}

// cairo-dock-core/src/gldit/cairo-dock-icon-factory.h
// #define CAIRO_DOCK_IS_AUTOMATIC_SEPARATOR(icon) (CAIRO_DOCK_ICON_TYPE_IS_SEPARATOR (icon) && (icon)->cDesktopFileName == NULL)
// #define CAIRO_DOCK_IS_USER_SEPARATOR(icon) (CAIRO_DOCK_ICON_TYPE_IS_SEPARATOR (icon) && (icon)->cDesktopFileName != NULL)

func (o *dockIcon) ModuleInstance() *ModuleInstance {
	if !o.IsApplet() {
		return nil
	}
	return NewModuleInstanceFromNative(unsafe.Pointer(o.Ptr.pModuleInstance))
}

func (o *dockIcon) MoveAfterIcon(container *CairoDock, target Icon) {
	C.cairo_dock_move_icon_after_icon(container.Ptr, o.Ptr, target.ToNative())
}

func (o *dockIcon) RemoveIconsFromSubdock(dest *CairoDock) {
	C.cairo_dock_remove_icons_from_dock(o.Ptr.pSubDock, dest.Ptr)
}

func (o *dockIcon) RemoveOverlayAtPosition(position cdtype.EmblemPosition) {
	C.cairo_dock_remove_overlay_at_position(o.Ptr, C.CairoOverlayPosition(position), C.gpointer(o.Ptr))
}

func (o *dockIcon) RemoveSubdockEmpty() {
	if o.Ptr.pSubDock != nil && o.Ptr.pSubDock.icons == nil {
		o.Ptr.pSubDock = nil
	}
}

func (o *dockIcon) RequestAnimation(animation string, rounds int) {
	cstr := (*C.gchar)(C.CString(animation))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	C.gldi_icon_request_animation(o.Ptr, cstr, C.int(rounds))
}

func (o *dockIcon) RequestAttention(animation string, nbRounds int) {
	cstr := (*C.gchar)(C.CString(animation))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	C.gldi_icon_request_attention(o.Ptr, cstr, C.int(nbRounds))
}

func (o *dockIcon) SetAllocatedSize(w, h int) {
	o.Ptr.iAllocatedWidth = C.gint(w)
	o.Ptr.iAllocatedHeight = C.gint(h)
}

func (o *dockIcon) SetIcon(str string) error {
	if o.Ptr.image.pSurface == nil {
		return errors.New("icon has no image.pSurface")
	}

	var cstr *C.gchar
	if str != "" {
		cstr = (*C.gchar)(C.CString(str))
		defer C.free(unsafe.Pointer((*C.char)(cstr)))
	}
	ctx := C.cairo_create(o.Ptr.image.pSurface)
	C.cairo_dock_set_image_on_icon(ctx, cstr, o.Ptr, o.GetContainer().Ptr) // returns gboolean
	C.cairo_destroy(ctx)

	return nil
}

func (o *dockIcon) SetLabel(str string) {
	var cstr *C.gchar
	if str != "" {
		cstr = (*C.gchar)(C.CString(str))
		defer C.free(unsafe.Pointer((*C.char)(cstr)))
	}
	C.gldi_icon_set_name(o.Ptr, cstr)
}

func (o *dockIcon) SetQuickInfo(str string) {
	var cstr *C.gchar
	if str != "" {
		cstr = (*C.gchar)(C.CString(str))
		defer C.free(unsafe.Pointer((*C.char)(cstr)))
	}
	C.gldi_icon_set_quick_info(o.Ptr, cstr)
}

// TODO: may have to move.
func (o *dockIcon) SubDockIcons() []Icon {
	if o.Ptr == nil || o.Ptr.pSubDock == nil {
		return nil
	}
	clist := glib.WrapList(uintptr(unsafe.Pointer(o.Ptr.pSubDock.icons)))
	return goListIcons(clist)
}

func (o *dockIcon) SubDockIsVisible() bool {
	dock := o.GetSubDock()
	if dock == nil {
		return false
	}
	cont := dock.ToContainer()
	if cont == nil {
		return false
	}
	return gobool(C.gtk_widget_get_visible(cont.Ptr.pWidget))
}

func (o *dockIcon) WriteContainerNameInConfFile(newdock string) {
	cstr := (*C.gchar)(C.CString(newdock))
	defer C.free(unsafe.Pointer((*C.char)(cstr)))
	C.gldi_theme_icon_write_container_name_in_conf_file(o.Ptr, cstr)
}

//
//----------------------------------------------------------[ DATA RENDERING ]--

func (o *dockIcon) AddNewDataRenderer(attr DataRendererAttributer) {
	cAttr, free := attr.ToAttribute()
	C.cairo_dock_add_new_data_renderer_on_icon(o.Ptr, o.GetContainer().Ptr, cAttr)
	free()
}

func (o *dockIcon) AddDataRendererWithText(attr DataRendererAttributer, dataRendererText DataRendererTextFunc) {
	cAttr, free := attr.ToAttribute()
	C.cairo_dock_add_new_data_renderer_on_icon(o.Ptr, o.GetContainer().Ptr, cAttr)
	o.dataRendererText = dataRendererText
	free()
}

func (o *dockIcon) DataRendererTextFuncPercent(values ...float64) []string {
	list := make([]string, len(values))
	for i, val := range values {
		list[i] = fmt.Sprintf("%.1f%%%%", val*100)

		//snprintf (cFormatBuffer, iBufferLength, fValue < .0995 ? "%.1f%%" : (fValue < 1 ? " %.0f%%" : "%.0f%%"), fValue * 100.);

	}
	return list
}

func (o *dockIcon) RemoveDataRenderer() {
	C.cairo_dock_remove_data_renderer_on_icon(o.Ptr)
	o.dataRendererText = nil
}

func (o *dockIcon) Render(values ...float64) error {
	if o.Ptr.image.pSurface == nil {
		return errors.New("Render: icon has no image.pSurface")
	}
	if o.GetContainer() == nil {
		return errors.New("Render: icon has no container")
	}

	// if o.dataRendererText != nil {
	// 	strlist := o.dataRendererText(values...)
	// 	charlist := make([]*C.gchar, len(values))
	// 	for i, val := range strlist {
	// 		charlist[i] = (*C.gchar)(C.CString(val))
	// 	}

	// 	// points to the first element of the slice array (good enough for C).
	// 	o.Ptr.pDataRenderer.pFormatData = C.gpointer(&charlist[0])

	// 	defer func() {
	// 		C.free_list_gchar(&charlist[0], C.int(len(values)))
	// 		o.Ptr.pDataRenderer.pFormatData = nil
	// 	}()
	// }

	list := make([]C.double, len(values))
	for i, val := range values {
		list[i] = C.double(val)
	}

	C.render_new_data_on_icon(o.Ptr, &list[0]) // points to the first element of the slice array (good enough for C).
	return nil
}
