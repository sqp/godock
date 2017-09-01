package AppTmpl

import "github.com/sqp/godock/libs/cdtype"

// EmblemTest is the position of the test emblem.
const EmblemTest = cdtype.EmblemTopRight

// DefaultIconTest is the default emblem icon for test.
const DefaultIconTest = "test.svg"

// Commands references.
const (
	cmdLeft = iota
)

//------------------------------------------------------------------[ CONFIG ]--

type appletConf struct {
	cdtype.ConfGroupIconBoth `group:"Icon"`
	groupConfiguration       `group:"Configuration"`
}

type groupConfiguration struct {
	GaugeName string
	Devices   []string

	LeftAction    int
	LeftCommand   string
	LeftClass     string
	MiddleAction  int
	MiddleCommand string

	// We can parse an int like type, used with iota definitions lists.
	// LocalId LocalId

	CommandOne string `default:"xterm"`

	// With a Duration, we can provide a default, set the unit and a min value.
	UpdateInterval cdtype.Duration `default:"60"`

	// The template is loaded from appletDir/cdtype.TemplateDir or absolute.
	DialogTemplate cdtype.Template `default:"myfile"`

	// The shortkey definition is filled, but the desc tag is still needed for
	//global shortcut config.
	//   (also when in external applet mode, desc is not forwarded to the "all
	//   shortkeys page". Known dock TODO).
	//
	//
	ShortkeyOpenThing cdtype.Shortkey `action:"1" desc:"Open that thing"`
	ShortkeyEditThing cdtype.Shortkey `action:"2" desc:"Edit that thing"`
}
