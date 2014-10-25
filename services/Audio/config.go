package Audio

import "github.com/sqp/godock/libs/cdtype"

// VolumeMax is the pulseaudio max value for speakers and channels volumes.
const VolumeMax = 65535

// EmblemMuted is the position of the "upload in progress" emblem.
const EmblemMuted = cdtype.EmblemTopRight

// DefaultIconMuted is the default emblem icon for muted streams.
const DefaultIconMuted = "muted.svg"

//------------------------------------------------------------------[ CONFIG ]--

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
	DisplayText   int
	DisplayValues int

	GaugeName string
	// GaugeRotate bool

	IconBroken  string
	VolumeStep  int
	StreamIcons bool
}

type groupActions struct {
	LeftAction   int
	MiddleAction int
	// MiddleCommand string

	MixerCommand  string
	MixerClass    string
	MixerShortkey string

	// Still hidden.
	Debug bool
}
