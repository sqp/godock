package cdtype_test

// The configuration loading can be fully handled by LoadConfig if the config
// struct is created accordingly.
//
// Main conf struct: one nested struct for each config page with its name
// referred to in the "group" tag.
//
// Sub conf structs: one field for each setting you need with the keys and types
// referring to those in your .conf file loaded by the dock. As fields are
// filled by reflection, they must be public and start with an uppercase letter.
//
// If the name of your field is the same as in the config file, just declare it.
//
// If you have different names, or need to refer to a lowercase field in the file,
// you must add a "conf" tag with the name of the field to match in the file.
//
// The applet source config file will be demo/src/config.go
//
// The matching configuration file will better be created by copying an existing
// applet config, and adapting it to your needs.
// More documentation about the dock config file can be found there:
//
//   cftype keys:       http://godoc.org/github.com/sqp/godock/widgets/cfbuild/cftype#KeyType
//   config test file:  https://raw.githubusercontent.com/sqp/godock/master/test/test.conf
//   cairo-dock wiki:   http://www.glx-dock.org/ww_page.php?p=Documentation&lang=en#27-Building%20the%20Applet%27s%20Configuration%20Window

//
//-----------------------------------------------------------[ src/config.go ]--

import (
	"github.com/sqp/godock/libs/cdapplet"
	"github.com/sqp/godock/libs/cdtype"

	"fmt"
)

//
//---------------------------------------------------------------[ constants ]--

// Emblem position for polling activity.
const emblemAction = cdtype.EmblemTopRight

// Define a list of commands references.
const (
	cmdClickLeft = iota
	cmdClickMiddle
)

// Define a list of actions references.
const (
	ActionNone = iota
	ActionOpenThing
	ActionEditThing
)

// LocalId defines an int like used as constant reference, parsed by the conf.
type LocalId int

//
//------------------------------------------------------------------[ config ]--

//
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
	LocalId LocalId

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

//
//---------------------------------------------------------------------[ doc ]--

func Example_config() {}

func testApplet(callnew cdtype.NewAppletFunc) {
	// This is not a valid way to start an applet.
	// Only used here to trigger the loading and output check.
	base := cdapplet.New()
	app := cdapplet.Start(callnew, base)
	fmt.Println(app != nil, app.Name())
}
