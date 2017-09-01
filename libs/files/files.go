// Package files provides files operations helpers.
package files

import (
	"github.com/sqp/godock/libs/cdglobal" // Dock types.

	"archive/tar"
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

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

//
//-------------------------------------------------------------------[ WRITE ]--

// Save write a file to disk at given location.
//
func Save(path string, data []byte, mode os.FileMode) error {
	return ioutil.WriteFile(path, data, mode)
}

// SetLastModif save download date file for a package.
//
func SetLastModif(path ...string) error {
	lastmodif := filepath.Join(append(path, "last-modif")...)
	content := time.Now().Format("20060102")
	return Save(lastmodif, []byte(content), cdglobal.FileMode)
}

//
//--------------------------------------------------------------[ UNCOMPRESS ]--

// UnTarGz extracts a tar gz reader to disk at given location.
//
// thanks to github.com/verybluebot/tarinator-go.
func UnTarGz(topath string, source io.ReadCloser) error {
	defer source.Close()
	var e error
	source, e = gzip.NewReader(source)
	if e != nil {
		return e
	}

	tarBallReader := tar.NewReader(source)

	for {
		header, e := tarBallReader.Next()
		if e != nil {
			if e == io.EOF {
				break
			}
			return e
		}

		filename := filepath.Join(topath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			e = os.MkdirAll(filename, os.FileMode(header.Mode)) // or use 0755 if you prefer

			if e != nil {
				return e
			}

		case tar.TypeReg:
			writer, e := os.Create(filename)

			if e != nil {
				return e
			}

			io.Copy(writer, tarBallReader)

			e = os.Chmod(filename, os.FileMode(header.Mode))

			if e != nil {
				return e
			}

			writer.Close()
		default:
			return fmt.Errorf("Unable to untar type: %c in file %s", header.Typeflag, filename)
		}
	}
	return nil
}
