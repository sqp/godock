/*
Package config is an automatic configuration loader for cairo-dock.

Config will fills the data of a struct from an INI config file with reflection.
Groups and keys in the file can be matched with the data struct by name or by a
special "conf" tag.

	GetKey  : Only parse using the field name. Names and keys need to be UpperCase.
	GetTag  ; Only parse using the "conf" tag of the field.
	GetBoth : Parse using both methods (tag is used when defined, name as fallback).

Parsing errors are stored in the Errors field.

Example for a single group

Load the data from the file and UnmarshalGroup a group.

	conf, e := config.NewFromFile(filepath) // Special conf reflector around the config file parser.
	if e != nil {
		return e
	}
	data := &groupConfiguration{}
	conf.UnmarshalGroup(data, groupName, config.GetKey)

Example with multiple groups

To load data from many groups splitted in according strucs, like applets config,
you have to define the main struct with a "group" tag that match the group in
the config file.

	data := &appletConf{}
	e := config.Load(filepath, data, config.GetBoth)

Structs data for the examples

This is an example of applet data with multiple groups and able to use the
lowercase keys of the Icon group.

	type appletConf struct {
		groupIcon          `group:"Icon"`
		groupConfiguration `group:"Configuration"`
	}

	type groupIcon struct {
		Icon string `conf:"icon"`
		Name string `conf:"name"`
	}

	type groupConfiguration struct {
		DisplayText   int
		DisplayValues int

		GaugeName string

		IconBroken  string
		VolumeStep  int
		StreamIcons bool

		// Still hidden.
		Debug bool
	}

*/
package config

import (
	"github.com/robfig/config" // Config parser.

	"github.com/sqp/godock/libs/cdtype"

	"bufio"
	"errors"
	"io"
	"reflect"
	"strings"
)

//
//--------------------------------------------------------------------[ KEYS ]--

// GetFieldKey is method to match config key name and struct field.
//
type GetFieldKey func(reflect.StructField) string

// GetKey is a GetFieldKey test that matches by the field name.
//
func GetKey(struc reflect.StructField) string {
	return struc.Name
}

// GetTag is a test GetFieldKey that matches by the struct tag is defined.
//
func GetTag(struc reflect.StructField) string {
	return struc.Tag.Get("conf")
}

// GetBoth is a GetFieldKey test that matches by the struct tag is defined,
// otherwise, use the field name.
//
func GetBoth(struc reflect.StructField) string {
	if tag := struc.Tag.Get("conf"); tag != "" {
		return tag // Got tag, use it.
	}
	return struc.Name // Else, use name.
}

//
//-----------------------------------------------------------------[ LOADING ]--

// Config file unmarshall. Parsing errors will be stacked in the Errors field.
//
type Config struct {
	config.Config
	Errors []error
}

// New creates a new Config parser.
//
func New() *Config {
	c := config.New(config.DEFAULT_COMMENT, config.ALTERNATIVE_SEPARATOR, false, false)
	return &Config{Config: *c}
}

// NewFromFile creates a new Config parser with reflection to fill fields.
//
func NewFromFile(filename string) (*Config, error) {
	c, e := config.ReadDefault(filename)
	if e != nil {
		return nil, e
	}
	return &Config{Config: *c}, nil
}

// NewFromReader creates a new Config parser with reflection to fill fields.
//
func NewFromReader(reader io.Reader) (*Config, error) {
	c := &Config{Config: *config.NewDefault()}
	e := c.Read(bufio.NewReader(reader))
	if e != nil {
		return nil, e
	}
	return c, nil
}

// Load config file and fills a config data struct.
//
//   First argument must be a the pointer to the data struct.
//   Second argument is the func to choose what key to load for each field.
//     Default methods provided: GetKey, GetTag, GetBoth.
//
func Load(filename string, v interface{}, fieldKey GetFieldKey) error {
	conf, e := NewFromFile(filename)
	if e != nil {
		return e
	}
	conf.Unmarshall(v, fieldKey)
	return nil
}

//
//---------------------------------------------------------------[ UNMARSHAL ]--

// Unmarshall fills a struct of struct with data from config.
// The First level is config group, matched by the key group.
// Second level is data fields, matched by the supplied GetFieldKey func.
//
func (c *Config) Unmarshall(v interface{}, fieldKey GetFieldKey) error {
	typ := reflect.Indirect(reflect.ValueOf(v)).Type().Elem() // Get the type of the struct behind the pointer.
	val := reflect.ValueOf(v).Elem()                          // ReflectValue of the config struct.

	val.Set(reflect.New(typ)) // Create a new empty struct.

	for i := 0; i < typ.NumField(); i++ { // Parsing all fields in grre.
		// log.Info("field", i, typ.Field(i).Name, typ.Field(i).Tag.Get("group"))
		if group := typ.Field(i).Tag.Get("group"); group != "" {
			c.unmarshalGroup(val.Elem().Field(i), group, fieldKey)
		}
	}

	// Get instance behind pointer. Not sure why I have to use 2x Elem()
	// maybe once to get inside the pointer and once inside the struct.
	// val = val.Elem().Elem()

	// // Parse struct to fill each group according to its tag.
	// typ := val.Type()
	// log.Info("kind", typ.Kind())
	// for i := 0; i < typ.NumField(); i++ { // Parsing all fields in type.
	// 	log.Info("field", i, typ.Field(i).Name, typ.Field(i).Tag.Get("group"))
	// 	if group := typ.Field(i).Tag.Get("group"); group != "" {
	// 		// log.Debug(typ.Field(i).Name, typ.Field(i).Tag.Get("group"))
	// 		c.UnmarshalGroup(val.Field(i), group, fieldKey)
	// 	}
	// }
	return nil
}

// UnmarshalGroup parse config to fill the struct provided with values from the
// given group in the file.
//
// The group param must match a group in the file with the format [MYGROUP]
//
func (c *Config) UnmarshalGroup(v interface{}, group string, fieldKey GetFieldKey) {
	conf := reflect.ValueOf(v).Elem()
	c.unmarshalGroup(conf, group, fieldKey)
}

// see UnmarshalGroup.
func (c *Config) unmarshalGroup(conf reflect.Value, group string, fieldKey GetFieldKey) {
	typ := conf.Type()
	for i := 0; i < typ.NumField(); i++ { // Parsing all fields in type.
		c.getField(conf.Field(i), group, fieldKey(typ.Field(i)))
	}
}

// Fill a single reflected field if it has the conf tag.
//
func (c *Config) getField(elem reflect.Value, group, key string) {
	var e error
	switch elem.Interface().(type) {

	case bool:
		var val bool
		val, e = c.Bool(group, key)
		elem.SetBool(val)

	case int, int32, cdtype.InfoPosition:
		var val int
		val, e = c.Int(group, key)
		elem.SetInt(int64(val))

	case string:
		var val string
		val, e = c.String(group, key)
		elem.SetString(val)

	case float64:
		var val float64
		val, e = c.Float(group, key)
		elem.SetFloat(val)

	case []string:
		var val string
		val, e = c.String(group, key)
		list := strings.Split(strings.TrimRight(val, ";"), ";")
		if list[len(list)-1] == "" {
			list = list[:len(list)-1]
		}
		elem.Set(reflect.ValueOf(list))

	default:
		c.logError(errors.New("Parse conf: wrong type: " + elem.Kind().String()))
	}
	if e != nil {
		c.logError(errors.New("Parse conf: " + e.Error()))
	}
}

// Test an error and append to the stack if needed.
//
func (c *Config) logError(e error) {
	if e != nil {
		c.Errors = append(c.Errors, e)
	}
}

//------------------------------------------------------------[ TEMP ]--

// Need to get access to the read function of our config library.
// this is just a copy with public access.

// Read config from a Reader.
func (c *Config) Read(buf *bufio.Reader) (err error) {
	var section, option string

	for {
		l, err := buf.ReadString('\n') // parse line-by-line
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		l = strings.TrimSpace(l)

		// Switch written for readability (not performance)
		switch {
		// Empty line and comments
		case len(l) == 0, l[0] == '#', l[0] == ';':
			continue

		// New section
		case l[0] == '[' && l[len(l)-1] == ']':
			option = "" // reset multi-line value
			section = strings.TrimSpace(l[1 : len(l)-1])
			c.AddSection(section)

		// No new section and no section defined so
		//case section == "":
		//return os.NewError("no section defined")

		// Other alternatives
		default:
			i := strings.IndexAny(l, "=:")

			switch {
			// Option and value
			case i > 0:
				i := strings.IndexAny(l, "=:")
				option = strings.TrimSpace(l[0:i])
				value := strings.TrimSpace(stripComments(l[i+1:]))
				c.AddOption(section, option, value)
			// Continuation of multi-line value
			case section != "" && option != "":
				prev, _ := c.RawString(section, option)
				value := strings.TrimSpace(stripComments(l))
				c.AddOption(section, option, prev+"\n"+value)

			default:
				return errors.New("could not parse line: " + l)
			}
		}
	}
	return nil
}

func stripComments(l string) string {
	// Comments are preceded by space or TAB
	for _, c := range []string{" ;", "\t;", " #", "\t#"} {
		if i := strings.Index(l, c); i != -1 {
			l = l[0:i]
		}
	}
	return l
}

//
