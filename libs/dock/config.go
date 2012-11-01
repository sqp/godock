package dock

import (
	"errors"
	"github.com/kless/goconfig/config"
	"reflect"
)

type Config struct {
	config.Config
	Errors []error
}

func NewConfig(filename string) (*Config, error) {
	c, e := config.ReadDefault(filename)
	if e != nil {
		return nil, e
	}
	return &Config{Config: *c}, nil
}

func (c *Config) logError(e error) {
	if e != nil {
		c.Errors = append(c.Errors, e)
	}
}

func (c *Config) Parse(group string, empty, conf interface{}) {
	//~ typ := reflect.TypeOf(conf)
	//~ term.Info("typ", typ) //, reflect.TypeOf(v).Kind())

	typ := reflect.TypeOf(empty)

	n := typ.NumField()
	elem := reflect.ValueOf(conf).Elem()

	for i := 0; i < n; i++ { // Parsing all fields in type.
		//~ field := typ.Field(i)
		field := elem.Field(i).Interface()
		//~ term.Info("XML Import Field mismatch", field.Name, elem.Field(i).Kind()) //, reflect.TypeOf(v).Kind())
		current := typ.Field(i)
		switch field.(type) {
		case bool:
			val, e := c.Bool(group, current.Name)
			c.logError(e)
			elem.Field(i).SetBool(val)

		case int:
			val, e := c.Int(group, current.Name)
			c.logError(e)
			elem.Field(i).SetInt(int64(val))

		case string:
			val, e := c.String(group, current.Name)
			c.logError(e)
			elem.Field(i).SetString(val)

		default:
			c.logError(errors.New("unknown field: " + current.Name))
		}

	}

	//~ if v, ok := parseMap[field.Name]; ok { // Got matching row in map
	//~ if elem.Field(i).Kind() == reflect.TypeOf(v).Kind() { // Types are compatible.
	//~ elem.Field(i).Set(reflect.ValueOf(v))
	//~ } else {
	//~ warn("XML Import Field mismatch", field.Name, elem.Field(i).Kind(), reflect.TypeOf(v).Kind())
	//~ }
	//~ }

}
