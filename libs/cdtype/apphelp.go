package cdtype

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"text/template"
)

//
//----------------------------------------------------------------[ TEMPLATE ]--

// TemplateDir defines the default templates dir name in applets dir.
var TemplateDir = "templates"

// Template defines a template formatter.
// Can be used in applet config struct with a "default" tag.
//
type Template struct {
	*template.Template // extends the basic template.
	FilePath           string
}

// NewTemplate creates a template with the given file path.
//
func NewTemplate(name, appdir string) (*Template, error) {
	paths := []string{
		filepath.Join(appdir, TemplateDir, name+".tmpl"), // short template name in templates dir.
		filepath.Join(appdir, TemplateDir, name),         // template name with ext in templates dir.
		filepath.Join(name),                              // full path to exact filename.
	}

	for _, path := range paths {
		_, err := os.Stat(path)
		if err != nil && !os.IsExist(err) {
			continue
		}
		loaded, e := template.ParseFiles(path)
		if e != nil {
			return nil, e
		}
		return &Template{
			FilePath: path,
			Template: loaded,
		}, nil
	}
	return nil, errors.New("template not found:" + name)
}

// ToString executes a template function with the given data.
//
func (t *Template) ToString(funcName string, data interface{}) (string, error) {
	if t.Template == nil {
		return "", errors.New("template not found: " + t.FilePath)
	}
	buff := bytes.NewBuffer(nil) // NewBuffer([]byte(""))
	e := t.ExecuteTemplate(buff, funcName, data)
	return buff.String(), e
}

//
//----------------------------------------------------------------[ SHORTKEY ]--

// Shortkey defines mandatory informations to register a shortkey.
// Can be used in applet config struct with a "desc" tag.
//
type Shortkey struct {
	ConfGroup string
	ConfKey   string
	Desc      string
	Shortkey  string
}

// ShortkeyAction groups a shortkey with its ActionID or callback.
//
// Action type can either be:
//   int             an ActionID.
//   func()          a simple callback.
//   func() error    a callback with possible error to log.
//
type ShortkeyAction struct {
	Action   interface{}
	Shortkey Shortkey
}

// TestKey tests if a shortkey registered the given key name.
// Launch the registered action and returns true if found.
// (Whether an error is (triggerred and logged) or not, the key was matched and called.)
//
func (sa *ShortkeyAction) TestKey(key string) (bool, error) {
	if sa.Shortkey.Shortkey != key {
		return false, nil
	}
	switch act := sa.Action.(type) {
	case func():
		act()
		return true, nil

	case func() error:
		e := act()
		return true, e
	}
	return false, nil
}
