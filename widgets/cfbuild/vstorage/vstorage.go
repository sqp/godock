// Package vstorage provides a virtual config storage for the config file builder.
//
package vstorage

import (
	"github.com/sqp/godock/widgets/cfbuild/cftype" // Types for config file builder usage.
	"github.com/sqp/godock/widgets/cfbuild/valuer" // Converts interface value.

	"errors"
	"reflect"
)

// vcf implements cftype.Storage.
type vcf struct {
	cftype.BaseStorage // filepath and build.

	list map[string]map[string]*interface{}
	keys []*cftype.Key
}

// NewVirtual creates a virtual config. Implements Storage.
//
func NewVirtual(filePath, fileDefault string, keys ...*cftype.Key) cftype.Storage {
	return &vcf{
		BaseStorage: cftype.BaseStorage{
			File:    filePath,
			Default: fileDefault,
		},
		list: make(map[string]map[string]*interface{}),
		keys: keys,
	}
}

//
//--------------------------------------------------------------[ MANAGEMENT ]--

func (c *vcf) List(group string) []*cftype.Key { return c.keys }
func (c *vcf) ToData() (uint64, string, error) { return 0, "", nil }

func (c *vcf) GetGroups() (uint64, []string) {
	index := make(map[string]struct{})
	var groups []string
	for _, key := range c.keys {
		_, ok := index[key.Group]
		if !ok {
			index[key.Group] = struct{}{}
			groups = append(groups, key.Group)
		}
	}
	return 0, groups
}

// Valuer creates a valuer for the key matching group and name.
//
func (c *vcf) Valuer(group, name string) valuer.Valuer {
	c.testK(group, name)
	return valuer.New(c.list[group][name])
}

func (c *vcf) Default(group, name string) (valuer.Valuer, error) {
	return nil, errors.New("default not implemented on vstorage")
}

//
//---------------------------------------------------------------------[ GET ]--

func (c *vcf) Bool(g, k string) (v bool, e error)           { e = c.Get(g, k, &v); return v, nil }
func (c *vcf) Int(g, k string) (v int, e error)             { e = c.Get(g, k, &v); return v, nil }
func (c *vcf) Float(g, k string) (v float64, e error)       { e = c.Get(g, k, &v); return v, nil }
func (c *vcf) String(g, k string) (v string, e error)       { e = c.Get(g, k, &v); return v, nil }
func (c *vcf) ListBool(g, k string) (v []bool, e error)     { e = c.Get(g, k, &v); return v, nil }
func (c *vcf) ListInt(g, k string) (v []int, e error)       { e = c.Get(g, k, &v); return v, nil }
func (c *vcf) ListFloat(g, k string) (v []float64, e error) { e = c.Get(g, k, &v); return v, nil }
func (c *vcf) ListString(g, k string) (v []string, e error) { e = c.Get(g, k, &v); return v, nil }

func (c *vcf) Get(g, k string, val interface{}) error {
	cur, ok := c.list[g][k]
	if ok {
		switch ptr := val.(type) {
		case *bool:
			*ptr = (*cur).(bool)

		case *int:
			*ptr = (*cur).(int)

		case *float64:
			*ptr = (*cur).(float64)

		case *string:
			*ptr = (*cur).(string)

		case *[]bool:
			*ptr = (*cur).([]bool)

		case *[]int:
			*ptr = (*cur).([]int)

		case *[]float64:
			*ptr = (*cur).([]float64)

		case *[]string:
			*ptr = (*cur).([]string)

		default:
			println("vstorage Get. bad type for key:", reflect.TypeOf(val), k)
		}

	} else {
		println("vstorage Get. no match for key:", k)
	}
	return nil
}

//
//---------------------------------------------------------------------[ SET ]--

func (c *vcf) Set(g, k string, val interface{}) error    { c.testK(g, k); *c.list[g][k] = val; return nil }
func (c *vcf) SetBool(g, k string, value bool)           { c.testK(g, k); *c.list[g][k] = value }
func (c *vcf) SetInt(g, k string, value int)             { c.testK(g, k); *c.list[g][k] = value }
func (c *vcf) SetFloat(g, k string, value float64)       { c.testK(g, k); *c.list[g][k] = value }
func (c *vcf) SetString(g, k string, value string)       { c.testK(g, k); *c.list[g][k] = value }
func (c *vcf) SetListBool(g, k string, value []bool)     { c.testK(g, k); *c.list[g][k] = value }
func (c *vcf) SetListInt(g, k string, value []int)       { c.testK(g, k); *c.list[g][k] = value }
func (c *vcf) SetListFloat(g, k string, value []float64) { c.testK(g, k); *c.list[g][k] = value }
func (c *vcf) SetListString(g, k string, value []string) { c.testK(g, k); *c.list[g][k] = value }

//
//------------------------------------------------------------------[ COMMON ]--

func (c *vcf) testK(group, name string) {
	if c.list[group] == nil {
		c.list[group] = make(map[string]*interface{})
	}
	if _, ok := c.list[group][name]; !ok {
		var val interface{}
		c.list[group][name] = &val
	}
}
