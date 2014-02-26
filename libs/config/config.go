package config

import (
	"github.com/robfig/config" // Config parser.

	// "github.com/sqp/godock/libs/log" // Display info in terminal.

	"errors"
	"reflect"
	"strings"

	"bufio"
	"io"
)

// Config file unmarshall. Parsing errors will be stacked in the Errors field.
//
type Config struct {
	config.Config
	Errors []error
}

// Config parser with reflection to fill fields.
//
func NewFromFile(filename string) (*Config, error) {
	c, e := config.ReadDefault(filename)
	if e != nil {
		return nil, e
	}
	return &Config{Config: *c}, nil
}

// Config parser with reflection to fill fields.
//
func NewFromReader(reader io.Reader) (*Config, error) {
	c := &Config{Config: *config.NewDefault()}
	e := c.Read(bufio.NewReader(reader))
	if e != nil {
		return nil, e
	}
	return c, nil
}

// Load config file and fills config a data struct.
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

// Method to choose which config key name to use to fill the struct field.
//
type GetFieldKey func(reflect.StructField) string

// Use field name as config source.
//
func GetKey(struc reflect.StructField) string {
	return struc.Name
}

// Use field tag as config source.
//
func GetTag(struc reflect.StructField) string {
	return struc.Tag.Get("conf")
}

// If a tag is defined use it as config source, otherwise, use the field name.
//
func GetBoth(struc reflect.StructField) string {
	if tag := struc.Tag.Get("conf"); tag != "" {
		return tag // Got tag, use it.
	}
	return struc.Name // Else, use name.
}

// Fill a struct of struct with data from config. First level is config group,
// matched by the key group. Second level is data fields, matched by the supplied
// GetFieldKey func.
//
func (c *Config) Unmarshall(v interface{}, fieldKey GetFieldKey) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return errors.New("non-pointer passed to Unmarshal")
	}

	// Get instance behind pointer. Not sure why I have to use 2x Elem()
	// maybe once to get inside the pointer and once inside the struct.

	val = val.Elem().Elem()

	// Parse struct to fill each group according to its tag.
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ { // Parsing all fields in type.
		if group := typ.Field(i).Tag.Get("group"); group != "" {
			// log.Debug(typ.Field(i).Name, typ.Field(i).Tag.Get("group"))
			c.UnmarshalGroup(val.Field(i), group, fieldKey)
		}
	}
	return nil
}

// Parse config to fill the conf param with values from the file in a specific
// group.
//
// The group param must match a group in the file with the format [MYGROUP]
//
func (c *Config) UnmarshalGroup(conf reflect.Value, group string, fieldKey GetFieldKey) {
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

	case int:
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
//
//
//
//
//
// OLD. Will DEPRECATE SOON !
//
// func (c *Config) Parse(group string, empty, conf interface{}) {
// 	//~ typ := reflect.TypeOf(conf)
// 	//~ term.Info("typ", typ) //, reflect.TypeOf(v).Kind())

// 	typ := reflect.TypeOf(empty)

// 	n := typ.NumField()
// 	elem := reflect.ValueOf(conf).Elem()

// 	for i := 0; i < n; i++ { // Parsing all fields in type.
// 		//~ field := typ.Field(i)
// 		field := elem.Field(i).Interface()
// 		//~ term.Info("XML Import Field mismatch", field.Name, elem.Field(i).Kind()) //, reflect.TypeOf(v).Kind())
// 		current := typ.Field(i)
// 		switch field.(type) {
// 		case bool:
// 			val, e := c.Bool(group, current.Name)
// 			c.logError(e)
// 			elem.Field(i).SetBool(val)

// 		case int:
// 			val, e := c.Int(group, current.Name)
// 			c.logError(e)
// 			elem.Field(i).SetInt(int64(val))

// 		case string:
// 			val, e := c.String(group, current.Name)
// 			c.logError(e)
// 			elem.Field(i).SetString(val)

// 		default:
// 			c.logError(errors.New("unknown field: " + current.Name))
// 		}

// 	}

// 	//~ if v, ok := parseMap[field.Name]; ok { // Got matching row in map
// 	//~ if elem.Field(i).Kind() == reflect.TypeOf(v).Kind() { // Types are compatible.
// 	//~ elem.Field(i).Set(reflect.ValueOf(v))
// 	//~ } else {
// 	//~ warn("XML Import Field mismatch", field.Name, elem.Field(i).Kind(), reflect.TypeOf(v).Kind())
// 	//~ }
// 	//~ }

// }
