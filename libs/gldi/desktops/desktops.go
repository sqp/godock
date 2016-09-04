// Package desktops manages and get desktop and screens info.
package desktops

/*
#cgo pkg-config: gldi

#include "cairo-dock-desktop-manager.h"          // g_desktopGeometry

static int screen_position_x(int i)	{ return g_desktopGeometry.pScreens[i].x; }
static int screen_position_y(int i)	{ return g_desktopGeometry.pScreens[i].y; }
static int screen_width_i(int i) 	{ return g_desktopGeometry.pScreens[i].width; }
static int screen_height_i(int i) 	{ return g_desktopGeometry.pScreens[i].height; }

// extern GldiDesktopGeometry g_desktopGeometry;

*/
import "C"

// CurrentDesktop returns the current desktop number.
//
func CurrentDesktop() int {
	return int(C.g_desktopGeometry.iCurrentDesktop)
}

// CurrentViewportX returns the current horizontal viewport number.
//
func CurrentViewportX() int {
	return int(C.g_desktopGeometry.iCurrentViewportX)
}

// CurrentViewportY returns the current vertical viewport number.
//
func CurrentViewportY() int {
	return int(C.g_desktopGeometry.iCurrentViewportY)
}

// Cycle switches to an adjacent desktop.
//
func Cycle(up, loop bool) {
	desk := CurrentDesktop()
	if up {
		desk++
		if desk >= NbDesktops() && loop {
			desk = 0
		}
	} else {
		desk--
		if desk < 0 && loop {
			desk = NbDesktops() - 1
		}
	}
	if 0 <= desk && desk < NbDesktops() {
		SetCurrent(desk, CurrentViewportX(), CurrentViewportY())
	}
}

// HeightAll returns the desktop total height.
//
func HeightAll() int {
	return int(C.g_desktopGeometry.Xscreen.height)
}

// HeightScreen returns the height of a desktop screen.
//
func HeightScreen(screenNb int) int {
	return int(C.screen_height_i(C.int(screenNb)))
}

// NbDesktops returns the number of desktops.
//
func NbDesktops() int {
	return int(C.g_desktopGeometry.iNbDesktops)
}

// NbScreens returns the number of desktop screens.
//
func NbScreens() int {
	return int(C.g_desktopGeometry.iNbScreens)
}

// NbViewportX returns the number of horizontal viewports.
//
func NbViewportX() int {
	return int(C.g_desktopGeometry.iNbViewportX)
}

// NbViewportY returns the number of vertical viewports.
//
func NbViewportY() int {
	return int(C.g_desktopGeometry.iNbViewportY)
}

// ScreenPosition returns the desktop screen position.
//
func ScreenPosition(i int) (int, int) {
	if 0 > i || i >= NbScreens() {
		return 0, 0
	}
	x := int(C.screen_position_x(C.int(i)))
	y := int(C.screen_position_y(C.int(i)))
	return x, y
}

// SetCurrent sets the desktop number displayed.
//
func SetCurrent(numDesktop, numViewportX, numViewportY int) {
	C.gldi_desktop_set_current(C.int(numDesktop), C.int(numViewportX), C.int(numViewportY))
}

// WidthAll returns the desktop total width.
//
func WidthAll() int {
	return int(C.g_desktopGeometry.Xscreen.width)
}

// WidthScreen returns the width of a desktop screen.
//
func WidthScreen(screenNb int) int {
	return int(C.screen_width_i(C.int(screenNb)))
}
