// Package files provides files operations helpers.
package files

import (
	"github.com/sqp/godock/libs/cdtype"         // ConfUpdater
	"github.com/sqp/godock/widgets/gtk/keyfile" // Write config file.

	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

//
//------------------------------------------------------------[ CONF UPDATER ]--

// Access prevents concurrent access to config files.
// Could be improved, but it may be safer to use for now.
//
var Access = sync.Mutex{}

// UpdateConfFile udates a key in a configuration file.
//
func UpdateConfFile(configFile, group, key string, value interface{}) error {
	cu, e := NewConfUpdater(configFile)
	if e != nil {
		return e
	}
	cu.Set(group, key, value)
	return cu.Save()
}

// NewConfUpdater creates a ConfUpdater for the given config file (full path).
//
func NewConfUpdater(configFile string) (cdtype.ConfUpdater, error) {
	// Ensure the file exists and get the file access rights to preserve them.
	fi, e := os.Stat(configFile)
	if e != nil {
		return nil, e
	}

	Access.Lock()

	flags := keyfile.FlagsKeepComments | keyfile.FlagsKeepTranslations
	pKeyF, e := keyfile.NewFromFile(configFile, flags)
	if e != nil {
		Access.Unlock()
		return nil, e
	}

	return &confUpdate{
		KeyFile:  *pKeyF,
		filePath: configFile,
		fileMode: fi.Mode(),
	}, nil
}

type confUpdate struct {
	keyfile.KeyFile             // Provides the Set method.
	filePath        string      // Full path to config file.
	fileMode        os.FileMode // File access rights.
}

func (cu *confUpdate) Save() error {
	defer cu.Cancel()
	_, content, e := cu.ToData()
	if e != nil {
		return e
	}
	return ioutil.WriteFile(cu.filePath, []byte(content), cu.fileMode)
}

func (cu *confUpdate) Cancel() {
	cu.KeyFile.Free()
	Access.Unlock()
}

//
//--------------------------------------------------------------------[ COPY ]--

// CopyDir copies files recursively from source to destination dir.
//
func CopyDir(src, dest string) {
	VisitFile := func(fp string, fi os.FileInfo, err error) error {
		subdir := strings.TrimPrefix(fp, src)
		switch {
		case subdir == "", err != nil:

		case fi.IsDir():
			os.MkdirAll(dest+subdir, fi.Mode())

		default:
			// log.Err(
			CopyFile(fp, dest+subdir, fi.Mode())
		}
		return nil
	}

	filepath.Walk(src, VisitFile)
}

// CopyFile copies source file to destination.
//
func CopyFile(source string, dest string, mode os.FileMode) (err error) {
	reader, err := os.Open(source)
	if err != nil {
		return err
	}
	defer reader.Close()

	writer, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.Copy(writer, reader)

	if err == nil {
		err = os.Chmod(dest, mode)
	}
	return
}

// IsExist checks whether a file or directory exists.
// It returns false when the file or directory does not exist.
func IsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}
