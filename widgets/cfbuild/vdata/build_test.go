package vdata_test

import (
	"github.com/conformal/gotk3/gtk"

	"github.com/stretchr/testify/assert"

	"github.com/sqp/godock/widgets/cfbuild/cftype"
	"github.com/sqp/godock/widgets/cfbuild/vdata"
	"github.com/sqp/godock/widgets/pageswitch" // Switcher for config pages.

	"testing"
)

func TestConfig(t *testing.T) {
	gtk.Init(nil)

	build := vdata.TestInit(vdata.New(nil, nil), vdata.PathTestConf())
	if build == nil {
		return
	}
	build.BuildAll(pageswitch.New())

	assert.Equal(t, countChanged(t, build), 0, "Build unchanged")

	build.KeyWalk(vdata.TestValues)
	assert.Equal(t, countChanged(t, build), 32, "Build full change")
}

func countChanged(t *testing.T, build cftype.Builder) int {
	count := 0
	build.KeyWalk(func(key *cftype.Key) {
		lold, e := key.Storage().Default(key.Group, key.Name) // TODO: should crash with vstorage.
		assert.NoError(t, e, "get storage default", key.Group, key.Name)
		if e == nil && key.ValueState(lold).IsChanged() {
			count++
		}
	})
	return count
}
