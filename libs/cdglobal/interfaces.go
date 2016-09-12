package cdglobal

import "unsafe"

// Crypto defines a way to decrypt and encrypt strings.
//
type Crypto interface {
	DecryptString(str string) string
	EncryptString(str string) string
}

// Shortkeyer defines the interface to keyboard shortkeys.
//
type Shortkeyer interface {
	// ConfFilePath returns the shortkey conf file path.
	//
	ConfFilePath() string

	// Demander returns the shortkey Demander.
	//
	Demander() string

	// Description returns the shortkey description.
	//
	Description() string

	// GroupName returns the shortkey group name.
	//
	GroupName() string

	// IconFilePath returns the shortkey icon file path.
	//
	IconFilePath() string

	// KeyName returns the config key name.
	//
	KeyName() string

	// KeyString returns the shortkey key reference as string.
	//
	KeyString() string

	// Rebind rebinds the shortkey.
	//
	Rebind(keystring, description string) bool

	// Success returns the shortkey success.
	//
	Success() bool
}

// Window defines the interface to desktop windows.
//
type Window interface {

	// ToNative returns the pointer to the native object.
	//
	ToNative() unsafe.Pointer

	// CanMinMaxClose returns whether the window can do those actions.
	//
	CanMinMaxClose() (bool, bool, bool)

	// Class returns the window class name.
	//
	Class() string

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

	// IsTransientWin returns whether the window is a transient (modal) for another.
	//
	IsTransientWin() bool

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

	// GetTransientWin returns the parent window for a transient (modal).
	//
	GetTransientWin() Window

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
