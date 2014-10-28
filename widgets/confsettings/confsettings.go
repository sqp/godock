// Package confsettings manages the own config of the GUI
//
// This allows the configuration of the GUI itself.
package confsettings

import (
	"github.com/sqp/godock/libs/config" // config parser.

	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	// GuiFilename is the name of the gui config file in the appdata dir.
	GuiFilename = "gui.conf"

	// GuiGroup is the name displayed in the config for the gui own config page.
	GuiGroup = "GUI Settings"

	// TmpFile is the path to the tempfile for the config tests (used for the diff).
	TmpFile = "/tmp/cairo-dock-test.conf"
)

// User own config live settings (what is currently active).
var (
	Settings *ConfigSettings
)

// ConfigSettings defines the options the user can set about the GUI itself.
// This GUI config page will often be refered as "own config".
//
type ConfigSettings struct {

	// TODO fix all widget keys and remove those.
	SaveEditor  string `conf:"save editor"`
	SaveEnabled bool   `conf:"save enabled"`

	File string
}

func (cs *ConfigSettings) loadConf(file string) error {
	conf, e := config.NewFromFile(file) // Special conf reflector around the config file parser.
	if e != nil {
		return e
	}
	conf.UnmarshalGroup(cs, GuiGroup, config.GetTag)
	// TODO: need to forward conf.Errors
	cs.File = file // Force value of file everytime, it's set to blank by unmarshal.
	return nil
}

// Init will try to load the own config data from the file, and create it if missing.
//
func Init(dirAppdata string, e error) error {
	if e != nil {
		return e
	}

	file := filepath.Join(dirAppdata, GuiFilename)

	// Create file if needed.
	if _, e = os.Stat(file); e != nil {
		e = ioutil.WriteFile(file, []byte(guiConfDefault()), os.FileMode(0644))
		if e != nil {
			return e
		}
	}

	// Create our user settings
	Settings = &ConfigSettings{}
	return Settings.loadConf(file)
}

// PathFile returns the path to the own config's config file.
//
func PathFile() string {
	return Settings.File
}

// SaveFile is the current GUI save call to check whether it can be safely used
// according to user settings. May move at some point.
//
func SaveFile(file, content string) {
	isOwn := filepath.Base(file) == GuiFilename

	if Settings.SaveEnabled || isOwn { // force save for own config page.
		ioutil.WriteFile(file, []byte(content), 0600)
		if isOwn {
			Settings.loadConf(file)
		}

	} else {
		ioutil.WriteFile(TmpFile, []byte(content), 0600)
		exec.Command(Settings.SaveEditor, file, TmpFile).Run()

		// TODO: need return error
	}
}
