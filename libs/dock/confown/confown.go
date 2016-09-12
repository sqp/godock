// Package confown manages the own config of the GUI
//
// This allows the configuration of the GUI itself.
package confown

import (
	"github.com/sqp/godock/libs/cdtype"       // Logger type.
	"github.com/sqp/godock/libs/config"       // Config parser.
	"github.com/sqp/godock/libs/files"        // Files operations.
	"github.com/sqp/godock/libs/gldi/globals" // Global variables.

	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	// GuiFilename is the name of the gui config file in the appdata dir.
	GuiFilename = "rework.conf"

	// GuiGroup is the name displayed in the config for the gui own config page.
	GuiGroup = "GUI Settings"

	// TmpFile is the path to the tempfile for the config tests (used for the diff).
	TmpFile = "/tmp/cairo-dock-test.conf"
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

// Settings is the user own config live settings (what is currently active).
//
var Settings = ConfigSettings{}

// ConfigSettings defines the options the user can set about the GUI itself.
// This GUI config page will often be referred as "own config".
//
type ConfigSettings struct {
	SeparatorWheelChangeDesktop SeparatorWheelType

	// TODO fix all widget key saving and remove those.
	SaveEditor  string
	SaveEnabled bool

	File string `conf:"-"` // File location, not saved.
	log  cdtype.Logger
}

// Load loads the own config settings.
//
func (cs *ConfigSettings) Load() error {
	file := cs.File // Backup the file path.
	e := config.GetFromFile(cs.log, file, func(cfg cdtype.ConfUpdater) {
		// Special conf reflector around the config file parser.
		liste := cfg.UnmarshalGroup(cs, GuiGroup, config.GetBoth)
		for _, e := range liste {
			cs.log.Err(e, "confown parse")
		}
	})
	cs.File = file // Force value of file every time, it's set to blank by unmarshal.

	return e
}

// ToVirtual returns whether the save is virtual or not (only prints).
//
func (cs *ConfigSettings) ToVirtual(file string) bool {
	// save disabled and no editor and not own conf.
	return !cs.SaveEnabled && cs.SaveEditor == "" && cs.File != file
}

// Init will try to load the own config data from the file, and create it if missing.
//
func Init(log cdtype.Logger, file string, e error) error {
	if e != nil {
		return e
	}

	// Create file if needed.
	if !files.IsExist(file) {
		source := globals.DirShareData(GuiFilename)
		e := files.CopyFile(source, file, os.FileMode(0644))
		if e != nil {
			return e
		}
		log.Info("created config file", file)
	}

	// Create our user settings
	Settings = ConfigSettings{
		File: file,
		log:  log,
	}
	return Settings.Load()
}

// PathFile returns the path to the own config's config file.
//
func PathFile() string {
	return Settings.File
}

// SaveFile is the current GUI save call to check whether it can be safely used
// according to user settings. May move at some point.
//
func SaveFile(file, content string) (tofile bool, e error) {
	isOwn := filepath.Base(file) == GuiFilename
	tofile = Settings.SaveEnabled || isOwn // force save for own config page.
	switch {
	case tofile:
		e = ioutil.WriteFile(file, []byte(content), 0600)

	case Settings.SaveEditor == "":

	default:
		ioutil.WriteFile(TmpFile, []byte(content), 0600)
		e = exec.Command(Settings.SaveEditor, file, TmpFile).Start()
	}

	return tofile, e
}
