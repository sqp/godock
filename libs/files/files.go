// Package files provides files operations helpers.
package files

import (
	"github.com/sqp/godock/widgets/gtk/keyfile" // Write config file.

	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// UpdateConfFile udates a key in a configuration file.
//
func UpdateConfFile(configFile, group, key string, value interface{}) error {
	pKeyF := keyfile.New()
	_, e := pKeyF.LoadFromFile(configFile, keyfile.FlagsKeepComments|keyfile.FlagsKeepTranslations)
	if e != nil {
		return e
	}

	// Get the file access rights to preserve them.
	fi, e2 := os.Stat(configFile)
	if e2 != nil {
		return e2
	}

	pKeyF.Set(group, key, value)
	_, content, _ := pKeyF.ToData()

	ioutil.WriteFile(configFile, []byte(content), fi.Mode())

	return nil
}

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
