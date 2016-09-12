// Package files provides files operations helpers.
package files

import (
	"github.com/sqp/godock/libs/cdtype" // Logger type.

	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

//
//--------------------------------------------------------------[ FILE MUTEX ]--

var access = sync.Mutex{}

// AccessLock locks and prevents concurrent access to config files.
// Could be improved, but it may be safer to use for now.
//
func AccessLock(log cdtype.Logger) {
	log.Debug("files.Access", "Lock")
	access.Lock()
}

// AccessUnlock releases the access to config files.
//
func AccessUnlock(log cdtype.Logger) {
	log.Debug("files.Access", "Unlock")
	access.Unlock()
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

// Reader returns a Reader to the given file, with its size and close call.
//
func Reader(filePath string) (r io.Reader, size int64, close func() error, e error) {
	f, e := os.Open(filePath)
	if e != nil {
		return nil, 0, nil, e
	}

	rdr := bufio.NewReader(f)

	stat, e := f.Stat()
	if e != nil {
		return nil, 0, nil, e
	}
	if stat.Size() == 0 {
		return nil, 0, nil, errors.New("empty file")
	}
	return rdr, stat.Size(), f.Close, nil
}
