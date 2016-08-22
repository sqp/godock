package weather

import (
	"github.com/sqp/godock/libs/net/download"

	"fmt"
	"sync"
	"time"
)

// DegreeSymbol needs to be added before unit temperature (C or F).
const DegreeSymbol = "°"

// URL for backend data.
const (
	WeatherComURLBase        = "http://wxdata.weather.com/wxdata"
	WeatherComSuffixCurrent  = "/weather/local/%s?cc=*"
	WeatherComSuffixForecast = "/weather/local/%s?dayf=%d"
	WeatherComSuffixCelcius  = "&unit=m"
)

// URL for webpages to open.
const (
	URLWeatherComHour    = "http://www.weather.com/weather/hourbyhour/graph/%s"
	URLWeatherComToday   = "http://www.weather.com/weather/today/%s"
	URLWeatherComTomorow = "http://www.weather.com/weather/tomorrow/%s"
	URLWeatherComDetail  = "http://www.weather.com/weather/wxdetail%d/%s"
)

// weatherCom downloads weather data from weather.com and implements Weather
//
type weatherCom struct {
	*Config

	current  *Current
	forecast *Forecast

	Err error
}

func (w *weatherCom) Current() *Current      { return w.current }
func (w *weatherCom) Forecast() *Forecast    { return w.forecast }
func (w *weatherCom) SetConfig(conf *Config) { w.Config = conf }
func (w *weatherCom) Clear()                 { w.current = nil; w.forecast = nil }

// WebpageURL returns the webpage url to open to the user.
//
func (w *weatherCom) WebpageURL(numDay int) string {
	switch {
	case numDay == -1:
		return fmt.Sprintf(URLWeatherComHour, w.LocationCode)

	case numDay == 0:
		return fmt.Sprintf(URLWeatherComToday, w.LocationCode)

	case numDay == 1:
		return fmt.Sprintf(URLWeatherComTomorow, w.LocationCode)
	}

	return fmt.Sprintf(URLWeatherComDetail, numDay, w.LocationCode) // ?dayNum=%d
}

//
//----------------------------------------------------------------[ GET DATA ]--

func (w *weatherCom) Get() chan error {
	chane := make(chan error, 2)
	defer close(chane)

	if w.Config == nil {
		chane <- ErrMissingConfObject
		return chane
	}

	if w.LocationCode == "" {
		chane <- ErrMissingLocationCode
		return chane
	}

	wait := sync.WaitGroup{}
	for _, call := range []func() error{w.dlCurrent, w.dlForecast} {
		wait.Add(1)
		call := call
		go func() {
			chane <- call()
			wait.Done()
		}()
	}
	wait.Wait()
	return chane
}

func (w *weatherCom) dlCurrent() error {
	if !w.DisplayCurrentIcon {
		return nil
	}

	cur := &Current{}
	url := fmt.Sprintf(WeatherComURLBase+WeatherComSuffixCurrent+unitTemp(w.UseCelcius), w.LocationCode)
	e := download.XML(url, cur)
	if e != nil {
		return e
	}

	// Received data is valid, update it.
	w.current = cur

	// Prepend degree symbol to unit if missing.
	cur.UnitTemp = unitDegree(cur.UnitTemp)

	// Parse sun time.
	cur.TxtSunrise, cur.TxtSunset, cur.IsNight, e = FormatSun(cur.Sunrise, cur.Sunset, w.Time24H)

	// Parse update time.
	upd, e := time.Parse("1/2/06 3:04 PM MST", cur.UpdateTime)
	if e != nil {
		return e
	}
	cur.TxtUpdateTime = upd.Format(timeFormat(w.Time24H))

	return nil
}

func (w *weatherCom) dlForecast() error {
	w.forecast = &Forecast{}
	url := fmt.Sprintf(WeatherComURLBase+WeatherComSuffixForecast+unitTemp(w.UseCelcius), w.LocationCode, w.NbDays+1)
	e := download.XML(url, w.forecast)
	if e != nil {
		return e
	}
	w.forecast.UnitTemp = unitDegree(w.forecast.UnitTemp)

	// Parse day number and sun time.
	for i := range w.forecast.Days {
		day := &w.forecast.Days[i]
		date, e := time.Parse("Jan 2", day.Date)
		if e != nil {
			return e
		}
		day.MonthDay = date.Day()

		day.TxtSunrise, day.TxtSunset, _, e = FormatSun(day.Sunrise, day.Sunset, w.Time24H)
		if e != nil {
			return e
		}

		for i := range day.Part {
			if day.Part[i].WeatherDescription == "" {
				day.Part[i].WeatherDescription = ValueMissing
			}
		}
	}

	return nil
}

//
//-----------------------------------------------------------[ FIND LOCATION ]--

// Search represents the xml base struct for the search location query.
//
type Search struct {
	Ver string `xml:"ver,attr"`
	Loc []Loc  `xml:"loc"`
}

// Loc represents a location found in the search location query.
//
type Loc struct {
	Type string `xml:"type,attr"`
	Name string `xml:",chardata"`
	ID   string `xml:"id,attr"`
}

// FindLocation asks the server the list of locations matching the given name.
//
func FindLocation(locationName string) ([]Loc, error) {
	var search Search
	e := download.XML(WeatherComURLBase+"/search/search?where="+locationName, &search)
	if e != nil {
		return nil, e
	}

	return search.Loc, nil
}

func unitTemp(useCelcius bool) string {
	if useCelcius {
		return WeatherComSuffixCelcius
	}
	return ""
}
