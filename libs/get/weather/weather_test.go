package weather_test

import (
	"github.com/stretchr/testify/assert"

	"github.com/sqp/godock/libs/get/weather"

	"testing"
)

func TestDockbus(t *testing.T) {
	wea := weather.New()
	locations, e := weather.FindLocation("paris")
	assert.NoError(t, e, "weather.FindLocation")
	assert.NotEmpty(t, locations, "locations")

	wea.SetConfig(&weather.Config{
		LocationCode:       locations[0].ID,
		NbDays:             10,
		DisplayCurrentIcon: true,
		Time24H:            true,
		UseCelcius:         true,
	})
	for e := range wea.Get() {
		assert.NoError(t, e, "weather.Get")
	}
	cur := wea.Current()
	assert.NotNil(t, cur, "Current")

	fc := wea.Forecast()
	assert.True(t, len(fc.Days) > 10, "Forecast")

	assert.Equal(t, "Paris-Montsouris, 75, FR", cur.Observatory, "Observatory")
	assert.Equal(t, "Paris, 75, France", cur.LocName, "LocName")
	assert.Equal(t, "°C", cur.UnitTemp, "UnitTemp")
	assert.Equal(t, "km/h", cur.UnitSpeed, "UnitSpeed")
	assert.NotEmpty(t, cur.WeatherIcon, "WeatherIcon")
	assert.NotEmpty(t, cur.MoonIcon, "MoonIcon")
}
