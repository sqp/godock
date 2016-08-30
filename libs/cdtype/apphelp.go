package cdtype

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//
//--------------------------------------------------------------[ NEW APPLET ]--

// Applets stores registered applets creation calls.
//
var Applets = make(ListApps)

// ListApps defines a list of applet creation func, indexed by applet name.
//
type ListApps map[string]NewAppletFunc

// NewAppletFunc defines an applet creation function.
//
type NewAppletFunc func(base AppBase, events *Events) AppInstance

// Register registers an applet name with its new func.
//
func (l ListApps) Register(name string, callnew NewAppletFunc) {
	l[name] = callnew
}

// Unregister unregisters an applet name.
//
func (l ListApps) Unregister(name string) {
	delete(l, name)
}

// GetNewFunc gets the applet creation func for the name.
//
func (l ListApps) GetNewFunc(name string) NewAppletFunc {
	return l[name]
}

//
//----------------------------------------------------------------[ DURATION ]--

// Duration converts a time duration to seconds.
//
// Really basic, so you have to recreate one every time.
// Used by the auto config parser with tags "unit", "default" and "min".
//
type Duration struct {
	value      int
	multiplier int
}

// NewDuration creates an time duration helper.
//
func NewDuration(value int) *Duration {
	if value < 1 {
		value = 1
	}
	return &Duration{value: value, multiplier: 1}
}

// Value gets the time duration in seconds.
//
func (i *Duration) Value() int {
	return i.value * i.multiplier
}

// SetDefault sets a default duration value.
//
func (i *Duration) SetDefault(def int) {
	if i.value == 0 {
		i.value = def
	}
}

// SetMin sets a min duration value.
//
func (i *Duration) SetMin(min int) {
	if min < 1 {
		min = 1
	}
	if i.value < min {
		i.value = min
	}
}

// SetUnit sets the time unit multiplier.
//
func (i *Duration) SetUnit(unitTime string) error {
	switch strings.ToLower(unitTime) {
	case "":
		return nil

	case "s", "second":
		i.multiplier = 1
		return nil

	case "m", "minute":
		i.multiplier = 60
		return nil

	case "h", "hour":
		i.multiplier = 3600
		return nil
	}
	return errors.New("Duration.SetUnit: unknown unit=" + unitTime)
}

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
	Shortkey  string       // value
	Desc      string       // tag "desc"
	ActionID  int          // tag "action". Will be converted to Call at SetDefaults.
	Call      func()       // Simple callback when triggered.
	CallE     func() error // Error returned callback. If set, used first.
}

// TestKey tests if a shortkey registered the given key name.
// Launch the registered CallE or Call and returns true if found.
//
// Error returned are callbacks errors.
// It can only happen when the key was matched and CallE returned something.
//
func (sa *Shortkey) TestKey(key string) (matched bool, cberr error) {
	switch {
	case sa.Shortkey != key: // Wrong key.

	case sa.CallE != nil:
		cberr = sa.CallE()
		matched = true

	case sa.Call != nil:
		sa.Call()
		matched = true
	}
	return matched, cberr
}
