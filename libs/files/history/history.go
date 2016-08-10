// Package history provides the core part for services history data management.
package history

import (
	"github.com/sqp/godock/libs/cdglobal"

	"os"
	"path"
)

// AppletLike defines a small applet like interface.
type AppletLike interface {
	FileDataDir(...string) string
}

// History defines the core part for services history data management.
//
type History struct {
	File string // Path to history file.
	Max  int    // max number of items in history.

	load func() error
	save func() error
	trim func()
}

// New creates a new history manager to save applets data.
//
func New(app AppletLike, filename string) *History {
	h := &History{
		load: func() error { return nil },
		save: func() error { return nil },
		trim: func() {},
		Max:  -1,
	}
	h.SetHistoryFile(app.FileDataDir(cdglobal.DirUserAppData, filename))
	return h
}

// SetHistoryFile sets the location of the history file.
//
func (h *History) SetHistoryFile(file string) {
	h.File = file
	h.Load()
}

// SetHistorySize sets the size of the history. It will trim history size if needed.
//   -1  : unlimited
//    0  : disabled
//  >=1  : limit set.
//
func (h *History) SetHistorySize(nb int) {
	h.Max = nb
	h.trim()
}

// SetFuncs sets callback funcs.
//
func (h *History) SetFuncs(load, save func() error, trim func()) {
	h.load = load
	h.save = save
	h.trim = trim
}

// Load loads the history data.
//
func (h *History) Load() error {
	return h.load()
}

// Save saves the history data.
//
func (h *History) Save() error {
	if _, e := os.Stat(path.Dir(h.File)); e != nil {
		e = os.Mkdir(path.Dir(h.File), os.ModePerm)
		if e != nil {
			return e
		}
	}
	return h.save()
}
