package iplocation_test

import (
	"github.com/stretchr/testify/assert"

	"github.com/sqp/godock/libs/net/iplocation"

	"testing"
)

func TestLocation(t *testing.T) {
	loc, e := iplocation.Get()

	if !assert.NoError(t, e, "iplocation.Get") &&
		!assert.NotNil(t, loc, "Get loc") {
		return
	}
	assert.NotEmpty(t, loc.Country, "loc.Country")
	assert.NotEmpty(t, loc.CountryCode, "loc.CountryCode")
	assert.NotEmpty(t, loc.Region, "loc.Region")
	assert.NotEmpty(t, loc.RegionName, "loc.RegionName")
	assert.NotEmpty(t, loc.City, "loc.City")
	assert.NotEmpty(t, loc.Lat, "loc.Lat")
	assert.NotEmpty(t, loc.Lon, "loc.Lon")
	assert.NotEmpty(t, loc.Timezone, "loc.Timezone")
	assert.NotEmpty(t, loc.ISP, "loc.ISP")
	assert.NotEmpty(t, loc.ORG, "loc.ORG")
}
