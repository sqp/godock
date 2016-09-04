// Package window manages desktop windows.
package window

/*
#cgo pkg-config: gldi

#include "cairo-dock-windows-manager.h"
*/
import "C"

import "unsafe"

//
//---------------------------------------------------------------[ INTERFACE ]--

// Type defines the interface to desktop windows.
//
type Type interface {

	// ToNative returns the pointer to the native object.
	//
	ToNative() unsafe.Pointer

	// CanMinMaxClose returns whether the window can do those actions.
	//
	CanMinMaxClose() (bool, bool, bool)

	// IsAbove returns whether the window is above.
	//
	IsAbove() bool // could split OrBelow but seem unused

	// IsActive returns whether the window is active.
	//
	IsActive() bool

	// IsFullScreen returns whether the window is full screen.
	//
	IsFullScreen() bool

	// IsHidden returns whether the window is hidden.
	//
	IsHidden() bool

	// IsMaximized returns whether the window is maximized.
	//
	IsMaximized() bool

	// IsOnCurrentDesktop returns whether the window is on current desktop.
	//
	IsOnCurrentDesktop() bool

	// IsOnDesktop returns whether the window is the given desktop and viewport number.
	//
	IsOnDesktop(desktopNumber, viewPortX, viewPortY int) bool

	// IsSticky returns whether the window is sticky.
	//
	IsSticky() bool

	// NumDesktop returns window desktop number. Can be -1.
	//
	NumDesktop() int

	// StackOrder returns window stack order.
	//
	StackOrder() int

	// ViewPortX returns window horizontal viewport number.
	//
	ViewPortX() int

	// ViewPortY returns window vertical viewport number.
	//
	ViewPortY() int

	// XID returns the window X identifier.
	//
	XID() uint

	// Close closes window.
	//
	Close()

	// Kill kills the window.
	//
	Kill()

	// Lower lowers the window.
	//
	Lower()

	// Maximize maximizes the window.
	//
	Maximize(full bool)

	// Minimize minimizes the window.
	//
	Minimize()

	// MoveToCurrentDesktop moves the window to the current desktop.
	//
	MoveToCurrentDesktop()

	// MoveToDesktop move the window to the given desktop and viewport number.
	//
	MoveToDesktop(desktopNumber, viewPortX, viewPortY int)

	// SetAbove sets the window above.
	//
	SetAbove(above bool)

	// SetFullScreen sets the window full screen state.
	//
	SetFullScreen(full bool)

	// SetSticky sets the window sticky state.
	//
	SetSticky(sticky bool)

	// Show shows the window.
	//
	Show()

	// SetVisibility sets the window visibility.
	//
	SetVisibility(show bool)

	// ToggleVisibility toggles the window visibility.
	//
	ToggleVisibility()
}

//
//------------------------------------------------------------[ WINDOW ACTOR ]--

// window defines a dock window actor.
//
type window struct {
	Ptr *C.GldiWindowActor
}

// NewFromNative wraps a dock window actor from C pointer.
//
func NewFromNative(p unsafe.Pointer) Type {
	if p == nil {
		return nil
	}
	return &window{(*C.GldiWindowActor)(p)}
}

// Get

func (o *window) ToNative() unsafe.Pointer { return unsafe.Pointer(o.Ptr) }
func (o *window) IsActive() bool           { return o.Ptr == C.gldi_windows_get_active() }
func (o *window) IsFullScreen() bool       { return gobool(o.Ptr.bIsFullScreen) }
func (o *window) IsHidden() bool           { return gobool(o.Ptr.bIsHidden) }
func (o *window) IsMaximized() bool        { return gobool(o.Ptr.bIsMaximized) }
func (o *window) IsOnCurrentDesktop() bool { return gobool(C.gldi_window_is_on_current_desktop(o.Ptr)) }
func (o *window) IsSticky() bool           { return gobool(C.gldi_window_is_sticky(o.Ptr)) }
func (o *window) StackOrder() int          { return int(o.Ptr.iStackOrder) }
func (o *window) NumDesktop() int          { return int(o.Ptr.iNumDesktop) }
func (o *window) ViewPortX() int           { return int(o.Ptr.iViewPortX) }
func (o *window) ViewPortY() int           { return int(o.Ptr.iViewPortY) }
func (o *window) XID() uint                { return uint(C.gldi_window_get_id(o.Ptr)) }

func (o *window) CanMinMaxClose() (bool, bool, bool) {
	var bCanMinimize, bCanMaximize, bCanClose C.gboolean
	C.gldi_window_can_minimize_maximize_close(o.Ptr, &bCanMinimize, &bCanMaximize, &bCanClose)
	return gobool(bCanMinimize), gobool(bCanMaximize), gobool(bCanClose)
}

func (o *window) IsAbove() bool { // could split OrBelow but seem unused.
	var isAbove, isBelow C.gboolean
	C.gldi_window_is_above_or_below(o.Ptr, &isAbove, &isBelow)
	return gobool(isAbove)
}

func (o *window) IsOnDesktop(desktopNumber, viewPortX, viewPortY int) bool {
	return gobool(C.gldi_window_is_on_desktop(o.Ptr, C.int(desktopNumber), C.int(viewPortX), C.int(viewPortY)))
}

// Actions

func (o *window) Close()                  { C.gldi_window_close(o.Ptr) }
func (o *window) Kill()                   { C.gldi_window_kill(o.Ptr) }
func (o *window) Lower()                  { C.gldi_window_lower(o.Ptr) }
func (o *window) Maximize(full bool)      { C.gldi_window_maximize(o.Ptr, cbool(full)) }
func (o *window) Minimize()               { C.gldi_window_minimize(o.Ptr) }
func (o *window) MoveToCurrentDesktop()   { C.gldi_window_move_to_current_desktop(o.Ptr) }
func (o *window) SetAbove(above bool)     { C.gldi_window_set_above(o.Ptr, cbool(above)) }
func (o *window) SetFullScreen(full bool) { C.gldi_window_set_fullscreen(o.Ptr, cbool(full)) }
func (o *window) SetSticky(sticky bool)   { C.gldi_window_set_sticky(o.Ptr, cbool(sticky)) }
func (o *window) Show()                   { C.gldi_window_show(o.Ptr) }

func (o *window) MoveToDesktop(desktopNumber, viewPortX, viewPortY int) {
	C.gldi_window_move_to_desktop(o.Ptr, C.int(desktopNumber), C.int(viewPortX), C.int(viewPortY))
}

func (o *window) SetVisibility(show bool) {
	if show {
		o.Show()
	}
	o.Minimize()
}

func (o *window) ToggleVisibility() {
	if o.IsActive() {
		o.Minimize()
	} else {
		o.Show()
	}
}

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

// MISSING
// gboolean bDisplayed;  /// not used yet...
// gboolean bDemandsAttention;
// GtkAllocation windowGeometry;
// gint iViewPortX, iViewPortY;
// gint iStackOrder;
// gchar *cClass;
// gchar *cWmClass;
// gchar *cName;
// gchar *cLastAttentionDemand;
// gint iAge;  // age of the window (a mere growing integer).
// gboolean bIsTransientFor;  // TRUE if the window is transient (for a parent window).
