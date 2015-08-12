// Package valuer stores and converts a pointer to interface{}.
package valuer

import "fmt"

// Valuer defines an access to a data stored as interface{}.
//
type Valuer interface {
	Int() int
	Bool() bool
	Float() float64
	String() string
	ListInt() []int
	ListBool() []bool
	ListFloat() []float64
	ListString() []string

	Count() int
	Sprint() string
	SprintI(id int) string
}

//
//------------------------------------------------------------------[ VALUER ]--

type valuer struct {
	val *interface{}
}

// New creates an access to a data stored as interface{}.
//
func New(v *interface{}) Valuer {
	return &valuer{v}
}

func (o *valuer) Int() int             { return (*o.val).(int) }
func (o *valuer) Bool() bool           { return (*o.val).(bool) }
func (o *valuer) Float() float64       { return (*o.val).(float64) }
func (o *valuer) String() string       { return (*o.val).(string) }
func (o *valuer) ListInt() []int       { return (*o.val).([]int) }
func (o *valuer) ListBool() []bool     { return (*o.val).([]bool) }
func (o *valuer) ListFloat() []float64 { return (*o.val).([]float64) }
func (o *valuer) ListString() []string { return (*o.val).([]string) }

func (o *valuer) GetI(id int) interface{} {
	switch ptr := (*o.val).(type) {
	case bool, int, float64, string:
		return ptr

	case []bool:
		if id < len(ptr) {
			return ptr[id]
		}

	case []int:
		if id < len(ptr) {
			return ptr[id]
		}

	case []float64:
		if id < len(ptr) {
			return ptr[id]
		}

	case []string:
		if id < len(ptr) {
			return ptr[id]
		}

	}
	println("valuer GetI. bad type, out of range or storage problem:", id, *o.val)
	return nil
}

func (o *valuer) Count() int {
	switch ptr := (*o.val).(type) {
	case bool, int, float64:
		return 1

	case string:
		if ptr == "" {
			return 0
		}
		return 1

	case []bool:
		return len(ptr)

	case []int:
		return len(ptr)

	case []float64:
		return len(ptr)

	case []string:
		return len(ptr)

	}
	println("valuer Count. bad type:", *o.val)
	return 0
}

func (o *valuer) Sprint() string {
	return fmt.Sprint(*o.val)
}

func (o *valuer) SprintI(id int) string {
	return fmt.Sprint(o.GetI(id))
}

// func (o *valuer) Get(v interface{}) { v = *o.val }

// Unused
func (o *valuer) Set(v interface{}) { *o.val = v }

// func (o *valuer) SetBool(v bool)       { *o.val = v }
// func (o *valuer) SetInt(v int)         { *o.val = v }
// func (o *valuer) SetFloat(v float64)   { *o.val = v }
// func (o *valuer) SetString(v string)   { *o.val = v }
