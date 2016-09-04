// Package current provides access to the dock current theme variables.
package current

/*
#cgo pkg-config: gldi

#include "cairo-dock-applications-manager.h"       // myTaskbarParam
#include "cairo-dock-dock-manager.h"               // myDocksParam

*/
import "C"

import "github.com/sqp/godock/libs/gldi/globals" // Global consts.

// DockIsLocked returns if the dock is locked or not.
//
func DockIsLocked() bool {
	return Docks.LockAll() || globals.FullLock
}

//
//-------------------------------------------------------------[ DOCKS PARAM ]--

// Docks gives access to the dock myDocksParam.
//
var Docks DocksParamType

// DocksParamType is a pseudo wrapper around the dock myDocksParam.
//
type DocksParamType struct{}

// FrameMargin defines a margin. TODO complete doc.
//
func (DocksParamType) FrameMargin() int { return int(C.myDocksParam.iFrameMargin) }

// LineWidth defines the dock border line width.
//
func (DocksParamType) LineWidth() int { return int(C.myDocksParam.iDockLineWidth) }

// LockAll gets and sets the lock all state.
//
func (DocksParamType) LockAll(b ...bool) bool {
	if len(b) > 0 {
		C.myDocksParam.bLockAll = cbool(b[0])
	}
	return gobool(C.myDocksParam.bLockAll)
}

// LockIcons gets and sets the user lock icon state.
//
func (DocksParamType) LockIcons(b ...bool) bool {
	if len(b) > 0 {
		C.myDocksParam.bLockIcons = cbool(b[0])
	}
	return gobool(C.myDocksParam.bLockIcons)
}

// Radius defines the radius of dock corners.
//
func (DocksParamType) Radius() int { return int(C.myDocksParam.iDockRadius) }

// ShowSubDockOnClick returns whether we need to show the subdock when icon is clicked or not.
//
func (DocksParamType) ShowSubDockOnClick() bool { return gobool(C.myDocksParam.bShowSubDockOnClick) }

//
//-------------------------------------------------------------[ ICONS PARAM ]--

// Icons gives access to the dock myIconsParam.
//
var Icons IconsParamType

// IconsParamType is a pseudo wrapper around the dock myIconsParam.
//
type IconsParamType struct{}

// Amplitude defines the default icon pmplitude.
//
func (IconsParamType) Amplitude() float64 { return float64(C.myIconsParam.fAmplitude) }

// Gap defines the default icon gap.
//
func (IconsParamType) Gap() int { return int(C.myIconsParam.iIconGap) }

// Width defines the default icon width.
//
func (IconsParamType) Width() int { return int(C.myIconsParam.iIconWidth) }

// Height defines the default icon height.
//
func (IconsParamType) Height() int { return int(C.myIconsParam.iIconHeight) }

// LabelSize defines the default icon label size.
//
func (IconsParamType) LabelSize() int { return int(C.myIconsParam.iLabelSize) }

// RevolveSeparator TODO FIND USAGE.
//
func (IconsParamType) RevolveSeparator() bool { return gobool(C.myIconsParam.bRevolveSeparator) }

// SeparatorWidth defines the default separator width.
//
func (IconsParamType) SeparatorWidth() int { return int(C.myIconsParam.iSeparatorWidth) }

// SeparatorHeight defines the default separator height.
//
func (IconsParamType) SeparatorHeight() int { return int(C.myIconsParam.iSeparatorHeight) }

//
//-----------------------------------------------------------[ TASKBAR PARAM ]--

// Taskbar gives access to the dock myTaskbarParam.
//
var Taskbar = TaskbarParamType{}

// TaskbarParamType is a pseudo wrapper around the dock myTaskbarParam.
//
type TaskbarParamType struct{}

// ActionOnMiddleClick returns the defined middle click action.
//
func (TaskbarParamType) ActionOnMiddleClick() int { return int(C.myTaskbarParam.iActionOnMiddleClick) }

// MinimizeOnClick returns whether applications should be minimized on click.
//
func (TaskbarParamType) MinimizeOnClick() bool { return gobool(C.myTaskbarParam.bMinimizeOnClick) }

// MixLauncherAppli returns whether applications and launchers are mixed.
//
func (TaskbarParamType) MixLauncherAppli() bool { return gobool(C.myTaskbarParam.bMixLauncherAppli) }

// OverWriteXIcons returns whether default X (desktop) icons are overwritten.
//
func (TaskbarParamType) OverWriteXIcons() bool { return gobool(C.myTaskbarParam.bOverWriteXIcons) }

// PresentClassOnClick returns whether we need to show class windows when icon is clicked or not.
//
func (TaskbarParamType) PresentClassOnClick() bool {
	return gobool(C.myTaskbarParam.bPresentClassOnClick)
}

// struct _CairoTaskbarParam {
// 	gboolean bShowAppli;
// 	gboolean bGroupAppliByClass;
// 	gint iAppliMaxNameLength;
// 	gboolean bHideVisibleApplis;
// 	gdouble fVisibleAppliAlpha;
// 	gboolean bAppliOnCurrentDesktopOnly;
// 	gboolean bDemandsAttentionWithDialog;
// 	gint iDialogDuration;
// 	gchar *cAnimationOnDemandsAttention;
// 	gchar *cAnimationOnActiveWindow;
// 	gint iMinimizedWindowRenderType;
// 	gchar *cOverwriteException;
// 	gchar *cGroupException;
// 	gchar *cForceDemandsAttention;
// 	CairoTaskbarPlacement iIconPlacement;
// 	gchar *cRelativeIconName;
// 	gboolean bSeparateApplis;
// 	} ;

// var TaskbarParam TaskbarParamType

// var TaskbarParam = &TaskbarParamType{&C.myTaskbarParam}
// Ptr *C.CairoTaskbarParam

// func (o *TaskbarParamType) ActionOnMiddleClick() int { return int(o.Ptr.iActionOnMiddleClick) }
// func (o *TaskbarParamType) OverWriteXIcons() bool    { return gobool(o.Ptr.bOverWriteXIcons) }

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
