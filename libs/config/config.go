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

This is an example of applet data with the common Icon group (Name, Debug, and
optional Icon).

	type appletConf struct {
		cdtype.ConfGroupIconBoth `group:"Icon"`
		groupConfiguration       `group:"Configuration"`
	}

	type groupConfiguration struct {
		DisplayText   int
		DisplayValues int

		GaugeName string

		IconBroken  string
		VolumeStep  int
		StreamIcons bool
	}

*/
package config

import (
	"github.com/robfig/config" // Config parser.

	"github.com/sqp/godock/libs/cdtype"

	"bufio"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
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
	config.Config // Extends the real config.

	Errors    []error
	appdir    string             // applet dir for templates loading.
	shortkeys []*cdtype.Shortkey // shortkeys found.
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

// Load loads a config file and fills a config data struct.
//
// Returns parsed defaults data, the list of parsing errors, and the main error
// if the load failed (file missing / not readable).
//
//   filename   Full path to the config file.
//   appdir     Application directory, to find templates.
//   v          The pointer to the data struct.
//   fieldKey   Func to choose what key to load for each field.
//              Usable methods provided: GetKey, GetTag, GetBoth.
//
func Load(filename, appdir string, v interface{}, fieldKey GetFieldKey) (cdtype.Defaults, []error, error) {
	conf, e := NewFromFile(filename)
	if e != nil {
		return cdtype.Defaults{}, nil, e
	}
	conf.appdir = appdir
	def := conf.Unmarshall(v, fieldKey)
	def.Shortkeys = conf.shortkeys
	return def, conf.Errors, nil
}

//
//---------------------------------------------------------------[ UNMARSHAL ]--

// Unmarshall fills a struct of struct with data from config.
// The First level is config group, matched by the key group.
// Second level is data fields, matched by the supplied GetFieldKey func.
//
func (c *Config) Unmarshall(v interface{}, fieldKey GetFieldKey) cdtype.Defaults {
	typ := reflect.Indirect(reflect.ValueOf(v)).Type().Elem() // Get the type of the struct behind the pointer.
	val := reflect.ValueOf(v).Elem()                          // ReflectValue of the config struct.

	val.Set(reflect.New(typ)) // Create a new empty struct.

	def := cdtype.Defaults{Commands: cdtype.Commands{}} // Empty defaults to gather groups auto set defaults.

	// Range over the first level of fields to find struct with tag "group".
	for i := 0; i < typ.NumField(); i++ { // Parsing all fields in grre.
		// log.Info("field", i, typ.Field(i).Name, typ.Field(i).Tag.Get("group"))
		if group := typ.Field(i).Tag.Get("group"); group != "" {
			// Get user data from the group.
			c.unmarshalGroup(val.Elem().Field(i), group, fieldKey)

			// Get applet defaults from the group if it's public and provides a ToDefaults method.
			if val.Elem().Field(i).CanInterface() {
				uncast := val.Elem().Field(i).Interface()
				getDef, ok := uncast.(cdtype.ToDefaultser)
				if ok {
					getDef.ToDefaults(&def)
				}
			}
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
	return def
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
		c.getField(conf.Field(i), group, fieldKey(typ.Field(i)), typ.Field(i).Tag)
	}
}

// Fill a single reflected field if it has the conf tag.
//
func (c *Config) getField(elem reflect.Value, group, key string, tag reflect.StructTag) {
	if key == "-" { // Disabled config key.
		return
	}

	// tagInt makes the call only if there is a valid value.
	tagInt := func(str string, call func(int)) {
		if str == "" {
			return
		}
		def, e := strconv.Atoi(str)
		if !c.testerr(e, group, key, "tag int") {
			call(def)
		}
	}

	switch elem.Interface().(type) {

	case bool:
		val, e := c.Bool(group, key)
		c.testerr(e, group, key, "bool value")
		elem.SetBool(val)

	case int, int32, int64, cdtype.InfoPosition, cdtype.RendererGraphType:
		val, e := c.Int(group, key)
		c.testerr(e, group, key, "int value")
		elem.SetInt(int64(val))

	case cdtype.Duration:
		val, e := c.Int(group, key)
		c.testerr(e, group, key, "Duration value")

		dur := cdtype.NewDuration(val)
		e = dur.SetUnit(tag.Get("unit"))
		c.testerr(e, group, key, "Duration unit")

		tagInt(tag.Get("default"), dur.SetDefault)
		tagInt(tag.Get("min"), dur.SetMin)

		elem.Set(reflect.ValueOf(*dur))

	case string:
		val, e := c.String(group, key)
		c.testerr(e, group, key, "string value")
		if val == "" {
			val = tag.Get("default")
		}
		elem.SetString(val)

	case float64:
		val, e := c.Float(group, key)
		c.testerr(e, group, key, "float64 value")
		elem.SetFloat(val)

	case []string:
		val, e := c.String(group, key)
		c.testerr(e, group, key, "[]string value")
		list := strings.Split(strings.TrimRight(val, ";"), ";")
		if list[len(list)-1] == "" {
			list = list[:len(list)-1]
		}
		elem.Set(reflect.ValueOf(list))

	case *cdtype.Shortkey:
		val, e := c.String(group, key)

		c.testerr(e, group, key, "Shortkey value")
		sk := &cdtype.Shortkey{
			ConfGroup: group,
			ConfKey:   key,
			Desc:      tag.Get("desc"),
			Shortkey:  val,
		}
		tagInt(tag.Get("action"), func(id int) {
			sk.ActionID = id
		})
		elem.Set(reflect.ValueOf(sk))
		c.shortkeys = append(c.shortkeys, sk)

	case cdtype.Template:
		name, e := c.String(group, key)
		c.testerr(e, group, key, "Template value")
		if name == "" {
			name = tag.Get("default")
		}
		if e == nil {
			tmpl, e := cdtype.NewTemplate(name, c.appdir)

			if e == nil {
				elem.Set(reflect.ValueOf(*tmpl))
			} else {
				elem.Set(reflect.ValueOf(cdtype.Template{FilePath: name}))
			}
		}

	default:
		if elem.Kind().String() == "int" { // Int like types often used as ref types.
			val, e := c.Int(group, key)
			c.testerr(e, group, key, "int like value")
			elem.SetInt(int64(val))

		} else {
			c.adderr(group, key, "unknown type: %s", elem.Kind().String())
		}
	}
}

func (c *Config) testerr(e error, group, key, msg string, args ...interface{}) bool {
	if e == nil {
		return false
	}
	c.adderr(group, key, msg+": %s", append(args, e.Error())...)
	return true
}

func (c *Config) adderr(group, key, msg string, args ...interface{}) {
	args = append([]interface{}{group, key}, args...)
	c.Errors = append(c.Errors, fmt.Errorf("config: %s / %s -- "+msg, args...))
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
