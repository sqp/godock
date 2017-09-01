// Package confown manages the own config of the dock and its GUI
//
// This allows the configuration of the new dock settings.
package confown

import (
	"github.com/sqp/godock/libs/cdglobal"
	"github.com/sqp/godock/libs/cdtype"       // Logger type.
	"github.com/sqp/godock/libs/config"       // Config parser.
	"github.com/sqp/godock/libs/files"        // Files operations.
	"github.com/sqp/godock/libs/gldi/globals" // Global variables.

	"io/ioutil"
	"path/filepath"
)

const (
	// GuiFilename is the name of the gui config file in the appdata dir.
	GuiFilename = "rework.conf"

	// GuiGroup is the name displayed in the config for the gui own config page.
	GuiGroup = "GUI Settings"

	// TmpFile is the path to the tempfile for the config tests (used for the diff).
	TmpFile = "cairo-dock-test.conf"
)

// SeparatorWheelType defines the action when a separator receives a wheel
// scroll event.
type SeparatorWheelType int

// Actions when a separator receives a wheel scroll event.
const (
	SeparatorWheelInactive    SeparatorWheelType = iota // Do nothing
	SeparatorWheelChangeRange                           // Change desktop but do not cycle.
	SeparatorWheelChangeLoop                            // Change desktop and cycle between first and last.
)

// Current is the user own config live settings (what is currently active).
//
var Current = ConfigSettings{}

// ConfigSettings defines new dock options.
// This GUI config page will often be referred as "own config".
//
type ConfigSettings struct {
	ConfigGUI `group:"GUI Settings"`

	File string `conf:"-"` // File location, not saved.
	log  cdtype.Logger
}

// ConfigGUI defines the options the user can set about the GUI itself.
//
type ConfigGUI struct {
	SeparatorWheelChangeDesktop SeparatorWheelType

	OnStartDebug        bool
	OnStartWebMon       bool
	CrashDisplayColored bool
	CrashRecovery       bool

	// TODO have more persons make tests on saving and remove those.
	SaveEditor  string
	SaveEnabled bool

	TmplReport cdtype.Template `default:"report"`
}

// Load loads the own config settings.
//
func (cs *ConfigSettings) Load() (*ConfigSettings, error) {
	file := cs.File // Backup the file path.
	log := cs.log

	_, _, listErr, e := config.Load(cs.log, file, globals.DirShareData(), globals.DirShareData(), &cs, config.GetKey)
	cs.File = file // Force value of file every time as it's set to blank by unmarshal (obj recreated).
	cs.log = log

	if cs.log.Err(e, "confown load file") {
		return nil, e
	}

	// Display non fatal errors.
	for _, e := range listErr {
		cs.log.Err(e, "confown load parsing")
	}

	return cs, nil
}

// ToVirtual returns whether the save is virtual or not (only prints).
//
func (cs *ConfigSettings) ToVirtual(file string) bool {
	// save disabled and no editor and not own conf.
	return !cs.SaveEnabled && cs.SaveEditor == "" && cs.File != file
}

// Init will try to load the own config data from the file, and create it if missing.
//
func Init(log cdtype.Logger, file string, e error) {
	if log.Err(e, "confown init get dir") {
		return
	}
	if Current.File != "" {
		return
	}

	// Create file if needed.
	if !files.IsExist(file) {
		orig := globals.DirShareData(cdglobal.ConfigDirDefaults, GuiFilename)
		cdtype.InitConf(log, orig, file)
	}

	// Create our user settings
	Current = ConfigSettings{
		File: file,
		log:  log,
	}
	cs, e := Current.Load()
	Current = *cs
	log.Err(e, "confown init load")
}

// PathFile returns the path to the own config's config file.
//
func PathFile() string {
	return Current.File
}

// SaveFile is the current GUI save call to check whether it can be safely used
// according to user settings. May move at some point.
//
func SaveFile(log cdtype.Logger, file, content string) (tofile bool, e error) {
	isOwn := filepath.Base(file) == GuiFilename
	tofile = Current.SaveEnabled || isOwn // force save for own config page.
	switch {
	case tofile:
		e = ioutil.WriteFile(file, []byte(content), 0600)

	case Current.SaveEditor == "":

	default:
		tmpfile, e := ioutil.TempFile("", TmpFile)
		if e != nil {
			return tofile, e
		}
		defer tmpfile.Close()

		// TODO: remove tempfile.

		_, e = tmpfile.WriteString(content)
		if e != nil {
			return tofile, e
		}
		e = log.ExecCmd(Current.SaveEditor, file, tmpfile.Name()).Start()
	}

	return tofile, e
}
