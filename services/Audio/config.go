package Audio

// const defaultUpdateDelay = 3
// const historyFile = "appdata/uptoshare_history.txt"

const VolumeMax = 65535

// EmblemAction is the position of the "upload in progress" emblem.
// const EmblemAction = cdtype.EmblemTopRight

//------------------------------------------------------------------[ CONFIG ]--

type appletConf struct {
	groupIcon          `group:"Icon"`
	groupConfiguration `group:"Configuration"`
	groupActions       `group:"Actions"`
}

type groupIcon struct {
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
