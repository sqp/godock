package build

import (
	"github.com/shirou/gopsutil/process"

	"github.com/sqp/godock/libs/cdglobal"
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/config" // Logger and update file.
	// Logger type.
	"github.com/sqp/godock/libs/files" // Files operations.
	"github.com/sqp/godock/libs/gldi/globals"

	"encoding/json"
	"os"
	"time"
)

const (
	// GroupHidden defines a group that should be hidden.
	GroupHidden = "Hidden" // TODO move to config as common
)

// Current is the user build config live settings (what is currently active).
//
var Current Config

// Counter counts builds and crashs.
//
type Counter struct {
	OK    uint
	Fail  uint
	Crash uint
}

// Config defines the options the user can set about the GUI itself.
// This GUI config page will often be referred as "own config".
//
type Config struct {
	Counter     map[string]Counter `conf:"-"`
	CounterDB   []byte
	TotalUptime int // in seconds

	File string `conf:"-"` // File location, not saved.
	log  cdtype.Logger
}

// Load loads the own config settings.
//
func (cs *Config) Load() error {
	file := cs.File // Backup the file path.
	e := config.GetFromFile(cs.log, file, func(cfg cdtype.ConfUpdater) {
		// Special conf reflector around the config file parser.
		listErr := cfg.UnmarshalGroup(cs, GroupHidden, config.GetBoth)
		for _, e := range listErr {
			cs.log.Err(e, "confown parse")
		}
	})
	cs.File = file // Force value of file every time, it's set to blank by unmarshal.

	// Usable counter.
	e = json.Unmarshal(cs.CounterDB, &cs.Counter)
	if e != nil {
		return e
	}

	return e
}

// Init will try to load the build config data from the file, and create it if missing.
//
func Init(log cdtype.Logger, file string, e error) {
	if log.Err(e, "pkgbuild init get dir") {
		return
	}
	if Current.File != "" {
		return
	}

	// Create file if needed.
	if !files.IsExist(file) {
		orig := globals.DirShareData(cdglobal.ConfigDirDefaults, cdglobal.FileBuildSource)
		cdtype.InitConf(log, orig, file)
	}

	// Create our user settings
	Current = Config{
		File:    file,
		log:     log,
		Counter: make(map[string]Counter),
	}
	e = Current.Load()
	log.Err(e, "pkgbuild init load file")
}

// Today returns today's counters in read only.
//
func (cs *Config) Today() (Counter, bool) {
	today := time.Now().Format("2006-01-02")
	bc, ok := cs.Counter[today]
	return bc, ok
}

// IncreaseCounter updates the build counter with a new success or fail.
//
func (cs *Config) IncreaseCounter(success bool) error {
	return cs.updateCounter(func(bc *Counter) {
		if success {
			bc.OK++
		} else {
			bc.Fail++
		}
		cs.log.Debug("Increase Build Counter", bc)
	})
}

// IncreaseCrash increase the crash counter for the day.
//
func (cs *Config) IncreaseCrash() error {
	return cs.updateCounter(func(bc *Counter) {
		bc.Crash++
		cs.log.Debug("Increase Crash Counter", bc)
	})
}

func (cs *Config) updateCounter(call func(*Counter)) error {
	return config.SetToFile(cs.log, cs.File, func(cfg cdtype.ConfUpdater) (e error) {
		today := time.Now().Format("2006-01-02")

		bc, _ := cs.Counter[today]
		call(&bc)
		cs.Counter[today] = bc

		cs.CounterDB, e = json.Marshal(cs.Counter)
		if e != nil {
			return e
		}

		// return e

		return cfg.Set(GroupHidden, "CounterDB", cs.CounterDB)
		// return cfg.MarshalGroup(cs, GroupHidden, config.GetBoth)
	})
}

// IncreaseUptime updates the uptime counter with the current process uptime.
// Use on program close.
//
func (cs *Config) IncreaseUptime() error {

	return config.SetToFile(cs.log, cs.File, func(cfg cdtype.ConfUpdater) (e error) {
		uptime := cfg.Valuer(GroupHidden, "TotalUptime").Int()
		psUptime, e := ProcessUptime()
		if e != nil {
			return e
		}
		cs.log.Info("Current uptime:", uptime, "Increased uptime by:", ProcessUptimeToSeconds(psUptime), "total uptime:", int32(uptime)+int32(ProcessUptimeToSeconds(psUptime)))
		return cfg.Set(GroupHidden, "TotalUptime", int32(uptime)+int32(ProcessUptimeToSeconds(psUptime)))
	})
}

// ProcessUptime returns the time when the process started.
//
func ProcessUptime() (time.Time, error) {
	ps, e := process.NewProcess(int32(os.Getpid()))
	if e != nil {
		return time.Time{}, e
	}

	tmpUptime, e := ps.CreateTime()
	if e != nil {
		return time.Time{}, e
	}
	return time.Unix(tmpUptime/1000, 0), nil
}

// ProcessUptimeToSeconds converts a ProcessUptime to a number of seconds.
//
func ProcessUptimeToSeconds(t time.Time) int {
	return -int(t.Sub(time.Now()).Seconds())
}
