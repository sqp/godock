package weather

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

const DegreeSymbol = "°"

const (
	WeatherComURLBase        = "http://wxdata.weather.com/wxdata"
	WeatherComSuffixCurrent  = "/weather/local/%s?cc=*"
	WeatherComSuffixForecast = "/weather/local/%s?dayf=%d"
	WeatherComSuffixCelcius  = "&unit=m"
)

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

func (w *weatherCom) Get() error {

	if w.Config == nil {
		return ErrMissingConfObject
	}

	if w.LocationCode == "" {
		return ErrMissingLocationCode
	}

	wait := sync.WaitGroup{}
	for _, call := range []func() error{w.dlCurrent, w.dlForecast} {
		wait.Add(1)
		call := call
		go func() {
			e := call()
			if e != nil {
				println(e.Error())
			}
			wait.Done()
		}()
	}
	wait.Wait()

	return nil
}

func (w *weatherCom) download(query string, args ...interface{}) ([]byte, error) {
	format := WeatherComURLBase + query + unitTemp(w.UseCelcius)
	URL := fmt.Sprintf(format, args...)
	return Download(URL)
}

func (w *weatherCom) dlCurrent() error {
	if !w.DisplayCurrentIcon {
		return nil
	}

	data, e := w.download(WeatherComSuffixCurrent, w.LocationCode)
	if e != nil {
		return e
	}

	w.current = &Current{}

	e = xml.Unmarshal(data, w.current)
	if e != nil {
		return e
	}

	// Prepend degree symbol to unit if missing.
	w.current.UnitTemp = unitDegree(w.current.UnitTemp)

	upd, e := time.Parse("2/01/06 3:04 PM MST", w.current.UpdateTime)
	if e != nil {
		return e
	}
	w.current.Updated = upd
	w.current.UpdateHour = upd.Hour()
	w.current.UpdateMinute = upd.Minute()
	return nil
}

func (w *weatherCom) dlForecast() error {
	data, e := w.download(WeatherComSuffixForecast, w.LocationCode, w.NbDays)
	if e != nil {
		return e
	}

	w.forecast = &Forecast{}

	e = xml.Unmarshal(data, w.forecast)
	if e == nil {
		return e
	}
	w.forecast.UnitTemp = unitDegree(w.forecast.UnitTemp)

	return nil
}

func Download(URL string) ([]byte, error) {
	resp, e := http.Get(URL)
	if e != nil {
		return nil, e
	}
	body, e := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if e != nil {
		return nil, e
	}

	return body, nil
}

//
//-----------------------------------------------------------[ FIND LOCATION ]--

type Search struct {
	Ver string `xml:"ver,attr"`
	Loc []Loc  `xml:"loc"`
}

type Loc struct {
	Type string `xml:"type,attr"`
	Name string `xml:",chardata"`
	ID   string `xml:"id,attr"`
}

func FindLocation(locationName string) ([]Loc, error) {
	data, e := Download(WeatherComURLBase + "/search/search?where=" + locationName)
	if e != nil {
		return nil, e
	}

	var locations Search
	e = xml.Unmarshal(data, &locations)
	if e != nil {
		return nil, e
	}

	return locations.Loc, nil
}

func unitTemp(useCelcius bool) string {
	if useCelcius {
		return WeatherComSuffixCelcius
	}
	return ""
}
