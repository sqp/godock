package cdtype_test

// The configuration loading can be fully handled by LoadConfig if the config
// struct is created accordingly.
//
// Main conf struct: one nested struct for each config page with its name refered
// to in the "group" tag.
//
// Sub conf structs: one field for each setting you need with the keys and types
// refering to those in your .conf file loaded by the dock. As fields are filled by
// reflection, they must be public and start with an uppercase letter.
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

//
type appletConf struct {
	groupIcon          `group:"Icon"`
	groupConfiguration `group:"Configuration"`
}

type groupIcon struct {
	Name string `conf:"name"`
	Icon string `conf:"icon"`
}

type groupConfiguration struct {
	GaugeName      string
	UpdateInterval int
	Devices        []string

	LeftAction    int
	LeftCommand   string
	LeftClass     string
	MiddleAction  int
	MiddleCommand string

	Debug bool
}

//
//---------------------------------------------------------------------[ doc ]--

func Example_config() {}
