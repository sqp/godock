package TVPlay

import "github.com/sqp/gupnp/upnptype"

type appletConf struct {
	groupIcon          `group:"Icon"`
	groupConfiguration `group:"Configuration"`
	groupActions       `group:"Actions"`
}

type groupIcon struct {
	Icon string `conf:"icon"`
	Name string `conf:"name"`
}

type groupConfiguration struct {
	VolumeDelta int
	SeekDelta   int

	PreferredRenderer string
	PreferredServer   string

	DialogEnabled bool
	DialogTimer   int
	AnimName      string
	AnimDuration  int

	WindowVisibility int
}

type groupActions struct {
	ActionClickMiddle string
	ActionMouseWheel  string

	ShortkeyMute         string
	ShortkeyVolumeUp     string
	ShortkeyVolumeDown   string
	ShortkeyPlayPause    string
	ShortkeyStop         string
	ShortkeySeekBackward string
	ShortkeySeekForward  string

	// Still hidden.
	Debug bool
}

// Actions available in right click menu.
//
var dockMenu = []int{
	int(upnptype.ActionToggleMute),
	int(upnptype.ActionVolumeUp),
	int(upnptype.ActionVolumeDown),
	int(upnptype.ActionNone),
	int(upnptype.ActionPlayPause),
	int(upnptype.ActionStop),
	int(upnptype.ActionSeekBackward),
	int(upnptype.ActionSeekForward),
}
