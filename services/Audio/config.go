package Audio

import "github.com/sqp/godock/libs/cdtype"

// VolumeMax is the pulseaudio max value for speakers and channels volumes.
const VolumeMax = 65535

// EmblemMuted is the position of the "upload in progress" emblem.
const EmblemMuted = cdtype.EmblemTopRight

// DefaultIconMuted is the default emblem icon for muted streams.
const DefaultIconMuted = "muted.svg"

// Commands references.
const (
	cmdMixer = iota
)

//------------------------------------------------------------------[ CONFIG ]--

type appletConf struct {
	cdtype.ConfGroupIconBoth `group:"Icon"`
	groupConfiguration       `group:"Configuration"`
	groupActions             `group:"Actions"`
}

type groupConfiguration struct {
	DisplayText   int
	DisplayValues int

	GaugeName string
	// GaugeRotate bool

	IconBroken  string
	VolumeStep  int64
	StreamIcons bool
}

type groupActions struct {
	LeftAction   int
	MiddleAction int
	// MiddleCommand string

	MixerCommand  string
	MixerClass    string
	MixerShortkey *cdtype.Shortkey

	ShortkeyAllMute     *cdtype.Shortkey
	ShortkeyAllIncrease *cdtype.Shortkey
	ShortkeyAllDecrease *cdtype.Shortkey
}
