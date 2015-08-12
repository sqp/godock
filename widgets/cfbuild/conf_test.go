package cfbuild_test

import (
	"github.com/stretchr/testify/assert"

	"github.com/sqp/godock/libs/log"
	"github.com/sqp/godock/widgets/cfbuild"
	"github.com/sqp/godock/widgets/cfbuild/cftype"
	"github.com/sqp/godock/widgets/cfbuild/valuer"
	"github.com/sqp/godock/widgets/cfbuild/vstorage"
	"github.com/sqp/godock/widgets/gtk/keyfile"

	"testing"
)

func TestValuerToBoth(t *testing.T) {
	group := "group"

	for _, conf := range []extendedStorage{
		vstorage.NewVirtual("", "").(extendedStorage),
		&cfbuild.CDConfig{
			KeyFile:     *keyfile.New(),
			BaseStorage: cftype.BaseStorage{File: ""},
		},
	} {
		testValuerToBoth(t, conf, group)
		testSetToValuer(t, conf, group)
		testGetSet(t, conf, group)
	}
}

func testGetSet(t *testing.T, v extendedStorage, group string) {
	name := "name"

	getSet := func(v cftype.Storage, group, name string, i, o interface{}) {
		v.Set(group, name+"2", i)
		log.Err(v.Get(group, name+"2", o), "get")
	}

	ib := true
	ob := false

	getSet(v, group, name, ib, &ob)
	v.SetBool(group, name, ib)
	rb, e := v.Bool(group, name)
	log.Err(e, "get bool")
	assert.Equal(t, ib, ob, "get/set bool")
	assert.Equal(t, ib, rb, "get/set bool")

	ii := 42
	oi := 0

	getSet(v, group, name, ii, &oi)
	v.SetInt(group, name, ii)
	ri, e := v.Int(group, name)
	log.Err(e, "get int")
	assert.Equal(t, ii, oi, "get/set int")
	assert.Equal(t, ii, ri, "get/set int")

	ifl := float64(4.2)
	ofl := float64(0)

	getSet(v, group, name, ifl, &ofl)
	v.SetFloat(group, name, ifl)
	rfl, e := v.Float(group, name)
	log.Err(e, "get float64")
	assert.Equal(t, ifl, ofl, "get/set float64")
	assert.Equal(t, ifl, rfl, "get/set float64")

	is := "42"
	os := ""

	getSet(v, group, name, is, &os)
	v.SetString(group, name, is)
	rs, e := v.String(group, name)
	log.Err(e, "get string")
	assert.Equal(t, is, os, "get/set string")
	assert.Equal(t, is, rs, "get/set string")

	ilb := []bool{true, true}
	olb := []bool{}

	getSet(v, group, name, ilb, &olb)
	v.SetListBool(group, name, ilb)
	rlb, e := v.ListBool(group, name)
	log.Err(e, "get list bool")
	assert.Equal(t, ilb, olb, "get/set list bool")
	assert.Equal(t, ilb, rlb, "get/set list bool")

	ili := []int{42, -1}
	oli := []int{}

	getSet(v, group, name, ili, &oli)
	v.SetListInt(group, name, ili)
	rli, e := v.ListInt(group, name)
	log.Err(e, "get list int")
	assert.Equal(t, ili, oli, "get/set list int")
	assert.Equal(t, ili, rli, "get/set list int")

	ilfl := []float64{4.2, -5.9}
	olfl := []float64{}

	getSet(v, group, name, ilfl, &olfl)
	v.SetListFloat(group, name, ilfl)
	rlfl, e := v.ListFloat(group, name)
	log.Err(e, "get list float")
	assert.Equal(t, ilfl, olfl, "get/set list float")
	assert.Equal(t, ilfl, rlfl, "get/set list float")

	ils := []string{"42", "or", "!"}
	ols := []string{}

	getSet(v, group, name, ils, &ols)
	v.SetListString(group, name, ils)
	rls, e := v.ListString(group, name)
	log.Err(e, "get list string")
	assert.Equal(t, ils, ols, "get/set list string")
	assert.Equal(t, ils, rls, "get/set list string")
}

func testSetToValuer(t *testing.T, conf extendedStorage, group string) {
	ii := 42
	ib := true
	it := float64(4.2)
	is := "42"
	ili := []int{42, -1}
	ilb := []bool{true, true}
	ilf := []float64{4.2, -5.9}
	ils := []string{"42", "or", "!"}

	conf.SetInt(group, "int", ii)
	conf.SetBool(group, "bool", ib)
	conf.SetFloat(group, "float", it)
	conf.SetString(group, "string", is)
	conf.SetListInt(group, "listint", ili)
	conf.SetListBool(group, "listbool", ilb)
	conf.SetListFloat(group, "listfloat", ilf)
	conf.SetListString(group, "liststring", ils)

	assert.Equal(t, ii, conf.Valuer(group, "int").Int(), "get/set int")
	assert.Equal(t, ib, conf.Valuer(group, "bool").Bool(), "get/set bool")
	assert.Equal(t, it, conf.Valuer(group, "float").Float(), "get/set float64")
	assert.Equal(t, is, conf.Valuer(group, "string").String(), "get/set string")
	assert.Equal(t, ili, conf.Valuer(group, "listint").ListInt(), "get/set list int")
	assert.Equal(t, ilb, conf.Valuer(group, "listbool").ListBool(), "get/set list bool")
	assert.Equal(t, ilf, conf.Valuer(group, "listfloat").ListFloat(), "get/set list float")
	assert.Equal(t, ils, conf.Valuer(group, "liststring").ListString(), "get/set list string")
}

func testValuerToBoth(t *testing.T, conf cftype.Storage, group string) {
	ii := 42
	ib := true
	it := float64(4.2)
	is := "42"
	ili := []int{42, -1}
	ilb := []bool{true, true}
	ilf := []float64{4.2, -5.9}
	ils := []string{"42", "or", "!"}

	vi := conf.Valuer(group, "int").(extendedValuer)
	vb := conf.Valuer(group, "bool").(extendedValuer)
	vf := conf.Valuer(group, "float").(extendedValuer)
	vs := conf.Valuer(group, "string").(extendedValuer)
	vli := conf.Valuer(group, "listint").(extendedValuer)
	vlb := conf.Valuer(group, "listbool").(extendedValuer)
	vlf := conf.Valuer(group, "listfloat").(extendedValuer)
	vls := conf.Valuer(group, "liststring").(extendedValuer)

	// Set from Valuer.

	vi.Set(ii)
	vb.Set(ib)
	vf.Set(it)
	vs.Set(is)
	vli.Set(ili)
	vlb.Set(ilb)
	vlf.Set(ilf)
	vls.Set(ils)

	// Test get from Valuer.

	assert.Equal(t, ii, vi.Int(), "get Valuer int")
	assert.Equal(t, ib, vb.Bool(), "get Valuer bool")
	assert.Equal(t, it, vf.Float(), "get Valuer float64")
	assert.Equal(t, is, vs.String(), "get Valuer string")
	assert.Equal(t, ili, vli.ListInt(), "get Valuer list int")
	assert.Equal(t, ilb, vlb.ListBool(), "get Valuer list bool")
	assert.Equal(t, ilf, vlf.ListFloat(), "get Valuer list float")
	assert.Equal(t, ils, vls.ListString(), "get Valuer list string")

	// Test get from Source with pointer.

	ob := false
	oi := 0
	ot := float64(0)
	os := ""
	olb := []bool{}
	oli := []int{}
	olf := []float64{}
	ols := []string{}

	log.Err(conf.Get(group, "int", &oi), "get int")
	log.Err(conf.Get(group, "bool", &ob), "get bool")
	log.Err(conf.Get(group, "float", &ot), "get float")
	log.Err(conf.Get(group, "string", &os), "get string")
	log.Err(conf.Get(group, "listint", &oli), "get listint")
	log.Err(conf.Get(group, "listbool", &olb), "get listbool")
	log.Err(conf.Get(group, "listfloat", &olf), "get listfloat")
	log.Err(conf.Get(group, "liststring", &ols), "get liststring")

	assert.Equal(t, ii, oi, "get Source Get int")
	assert.Equal(t, ib, ob, "get Source Get bool")
	assert.Equal(t, it, ot, "get Source Get float64")
	assert.Equal(t, is, os, "get Source Get string")
	assert.Equal(t, ilb, olb, "get Source Get list bool")
	assert.Equal(t, ili, oli, "get Source Get list int")
	assert.Equal(t, ilf, olf, "get Source Get list float")
	assert.Equal(t, ils, ols, "get Source Get list string")

	// Test get from Source directly.

	oi, _ = conf.Int(group, "int")
	ob, _ = conf.Bool(group, "bool")
	ot, _ = conf.Float(group, "float")
	os, _ = conf.String(group, "string")
	olb, _ = conf.ListBool(group, "listbool")
	oli, _ = conf.ListInt(group, "listint")
	olf, _ = conf.ListFloat(group, "listfloat")
	ols, _ = conf.ListString(group, "liststring")

	assert.Equal(t, ii, oi, "get Source fields int")
	assert.Equal(t, ib, ob, "get Source fields bool")
	assert.Equal(t, it, ot, "get Source fields float64")
	assert.Equal(t, is, os, "get Source fields string")
	assert.Equal(t, ili, oli, "get Source fields list int")
	assert.Equal(t, ilb, olb, "get Source fields list bool")
	assert.Equal(t, ilf, olf, "get Source fields list float")
	assert.Equal(t, ils, ols, "get Source fields list string")

}

type extendedValuer interface {
	valuer.Valuer
	Set(interface{})
}

type extendedStorage interface {
	cftype.Storage
	SetInt(group, key string, value int)
	SetBool(group, key string, value bool)
	SetFloat(group, key string, value float64)
	SetString(group, key string, value string)
	SetListInt(group, key string, value []int)
	SetListBool(group, key string, value []bool)
	SetListFloat(group, key string, value []float64)
	SetListString(group, key string, value []string)
}
