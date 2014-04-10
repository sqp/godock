// Package cdtype defines main types and constants for Cairo-Dock applets.
package cdtype

// Events you can connect with the cairo-dock applet. They are better set in the
// mandatory DefineEvents call of your applet. All those events are optional but
// it's better to find something meaningful to assign to them to improve your
// applet utility.
//
// Use with something like:
//    app.Events.OnClick = func () {app.onClick()}
//    app.Events.OnDropData = func (data string) {app.openWebpage(data)}
//
// Reload event is optional. Here is the default call if you want to override it.
//
// 	app.Events.Reload = func(confChanged bool) {
// 		log.Debug("Reload module")
// 		app.Init(confChanged)
//  	if app.Poller() != nil {
// 			app.Poller().Restart() // send our restart event.
//  	}
// 	}
//
type Events struct {
	// Action when the user clicks on the icon.
	OnClick func()

	// Action when the user use the middle-click on the icon.
	OnMiddleClick func()

	// Action when the user use the right-click on the icon, opening the menu.
	OnBuildMenu func()

	// Action when the user click on an item of the menu added in OnBuildMenu.
	OnMenuSelect func(itemid int32)

	// Action when the user use the mouse wheel on the icon.
	OnScroll func(scrollUp bool)

	// Action when the user drop data on the icon.
	OnDropData func(data string)

	// ??
	OnAnswer func(data interface{})

	// Action when the user answers a dialog you raised beforehand.
	OnAnswerDialog func(button int32, data interface{})

	// Action when the user triggers a registered shortkey.
	OnShortkey func(key string)

	// Action when the focus of the managed window change.
	OnChangeFocus func(active bool)

	// Action when a reload applet event is triggered from the dock.
	Reload func(bool)

	// Action when the quit applet event is triggered from the dock.
	End func()
}

// SubEvents work the same as main event with an additional argument for the id
// of the clicked icon.
//
type SubEvents struct {
	// Action when the user clicks on the subicon.
	OnSubClick func(state int32, icon string)

	// Action when the user use the middle-click on the subicon.
	OnSubMiddleClick func(icon string)

	// Action when the user use the right-click on the subicon, opening the menu.
	OnSubBuildMenu func(icon string)

	// Action when the user click on an item of the menu added in OnSubBuildMenu.
	OnSubMenuSelect func(numEntry int32, icon string)

	// Action when the user use the mouse wheel on the subicon.
	OnSubScroll func(scrollUp bool, icon string)

	// Action when the user drop data on the subicon.
	OnSubDropData func(data string, icon string)
}

// DockProperties defines basic informations about a dock icon.
//
type DockProperties struct {
	X      int32 // Distance from the left of the screen.
	Y      int32 // Distance from the bottom of the screen.
	Width  int32 // Width of the icon.
	Height int32 // Height of the icon.

	Orientation uint32 // Dock orientation.
	Container   uint32 // Container type

	HasFocus bool   // True if the monitored window has the cursor focus.
	Xid      uint64 // Xid of the monitored window. Value > 0 if a window is monitored.
}

// ScreenPosition refers to the border of the screen the dock is attached to.
type ScreenPosition int32

const (
	ScreenBottom ScreenPosition = iota // Dock in the bottom.
	ScreenTop                          // Dock in the top.
	ScreenRight                        // Dock in the right.
	ScreenLeft                         // Dock in the left.
)

// ContainerType is the type of container that manages the applet.
//
type ContainerType int32

const (
	ContainerDock    ContainerType = iota // Applet in a dock.
	ContainerDesklet                      // Applet in a desklet.
)

// InfoPosition is the location to render text data for an applet.
//
type InfoPosition int32

const (
	InfoNone    = iota // don't display anything.
	InfoOnIcon         // display info on the icon (as quick-info).
	InfoOnLabel        // display on the label of the icon.
)

// EmblemPosition is the location where an emblem is displayed.
//
type EmblemPosition int32

const (
	EmblemTopLeft     EmblemPosition = iota // Emblem in top left.
	EmblemBottomLeft                        // Emblem in bottom left.
	EmblemBottomRight                       // Emblem in bottom right.
	EmblemTopRight                          // Emblem in top right.
	EmblemMiddle                            // Emblem in the middle.
	EmblemBottom                            // Emblem in the bottom.
	EmblemTop                               // Emblem in the top.
	EmblemRight                             // Emblem in the right.
	EmblemLeft                              // Emblem in the left.
)

// type EmblemModifier int32

// const (
// 	EmblemPersistent EmblemModifier = 0
// 	EmblemPrint      EmblemModifier = 9
// )

// MenuItemType is the type of menu entry to create.
//
type MenuItemType int32

const (
	MenuEntry       MenuItemType = iota // Simple menu text entry.
	MenuSubMenu                         // Not working.
	MenuSeparator                       // Menu separator.
	MenuCheckBox                        // Not working.
	MenuRadioButton                     // Not working.
)

// MenuItemId
// const (
// 	MainMenuId = 0
// )

// DialogKey

const (
	DialogKeyEnter  = -1 // Answer when the user press enter.
	DialogKeyEscape = -2 // Answer when the user press escape.
)
