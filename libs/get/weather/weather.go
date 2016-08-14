// Package weather provides a data source for weather informations.
package weather

import (
	"github.com/sqp/godock/libs/cdtype"

	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"
)

// Weather defines the usage of the weather data source.
//
type Weather interface {
	WebpageURL(numDay int) string
	Get() chan error
	Current() *Current
	Forecast() *Forecast
	SetConfig(*Config)
	Clear()
}

// Weather errors.
var (
	ErrMissingConfObject   = errors.New("weather #1: " + trans("conf object missing"))
	ErrMissingLocationCode = errors.New("weather #2: " + trans("location missing"))
	ValueMissing           = trans("N/A")
)

// BackendID represents the reference for a backend source of weather data.
//
type BackendID int

// Weather backends.
const (
	BackendWeatherCom BackendID = iota
)

// Other possible backends.

// https://github.com/schachmat/wego            -- terminal weather client       --  forecast.io account ||  Worldweatheronline account
// https://github.com/jfrazelle/weather         -- Weather via the command line  --  forecast.io API
// https://github.com/sramsay/wu                -- fast command-line application --  http://www.wunderground.com/weather/api/.
// https://github.com/benschw/weather-go        -- demo app                      --  openweathermap
// https://github.com/msoap/yandex-weather-cli  --  Command line interface       --  Yandex weather service

// New creates a new weather data source for the given backend.
//
func New(backend BackendID) Weather {
	return &weatherCom{}
}

// Config defines the weather data source configuration.
//
type Config struct {
	LocationCode       string // hidden in conf
	UseCelcius         bool   // format temp celcius or fahrenheit
	Time24H            bool   // format time 24H or 12H (AM/PM).
	DisplayCurrentIcon bool   // current weather.
	NbDays             int    // forecast (next days).
}

//
//--------------------------------------------------------[ TRANSLATION TODO ]--

func trans(str string) string { return str }

//
//-----------------------------------------------------------------[ CURRENT ]--

// Current contains weather data for the current day.
//
type Current struct {
	UnitDistance string `xml:"head>ud"`
	UnitPressure string `xml:"head>up"`
	UnitSpeed    string `xml:"head>us"`
	UnitTemp     string `xml:"head>ut"`

	LocName string `xml:"loc>dnam"`
	Sunrise string `xml:"loc>sunr"`
	Sunset  string `xml:"loc>suns"`

	UpdateTime          string `xml:"cc>lsup"`
	Observatory         string `xml:"cc>obst"`
	TempReal            int    `xml:"cc>tmp"`
	TempFelt            int    `xml:"cc>flik"`
	WeatherDescription  string `xml:"cc>t"`
	WeatherIcon         string `xml:"cc>icon"`
	Pressure            string `xml:"cc>bar>r"`
	PressureDescription string `xml:"cc>bar>d"`
	WindSpeed           string `xml:"cc>wind>s"`
	WindDirection       string `xml:"cc>wind>t"`
	Humidity            string `xml:"cc>hmid"`
	Visibility          string `xml:"cc>vis"`
	MoonIcon            int    `xml:"cc>moon>icon"`
	UVi                 int    `xml:"cc>uv>i"`
	UVDescription       string `xml:"cc>uv>t"`
	// <dewp>10</dewp>

	// Template stuff
	Template      *cdtype.Template // template for the field formater.
	TxtUpdateTime string           // formated (h24) UpdateTime
	TxtSunrise    string           // formated (h24) Sunrise
	TxtSunset     string           // formated (h24) Sunset
	IsNight       bool
}

// Format returns the template formatted string for today.
//
func (wc *Current) Format(template *cdtype.Template) (string, error) {
	wc.Template = template
	return FormatTemplate(template, "Current", wc)
}

// Fields format a list of fields from the template.
//
func (wc *Current) Fields(list ...string) (string, error) {
	return FormatField(wc.Template, list, wc)
}

//
//----------------------------------------------------------------[ FORECAST ]--

// Forecast defines weather forecast data for the following days.
//
type Forecast struct {
	Ver          string `xml:"ver,attr"`
	Ur           string `xml:"head>ur"`
	UnitDistance string `xml:"head>ud"`
	UnitPressure string `xml:"head>up"`
	UnitSpeed    string `xml:"head>us"`
	UnitTemp     string `xml:"head>ut"`
	Locale       string `xml:"head>locale"`
	Form         string `xml:"head>form"`
	Tm           string `xml:"loc>tm"`
	Dnam         string `xml:"loc>dnam"`
	Lon          string `xml:"loc>lon"`
	Lat          string `xml:"loc>lat"`
	Zone         string `xml:"loc>zone"`

	UpdateTime string `xml:"dayf>lsup"`
	Days       []Day  `xml:"dayf>day"`

	// Template stuff
	Template *cdtype.Template // template for the field formater.
	*Day                      // requested day
	*Part                     // requested part of day
}

// Day defines weather forecast data for one of the following day.
//
type Day struct {
	Date     string `xml:"dt,attr"`
	DayName  string `xml:"t,attr"`
	DayCount string `xml:"d,attr"`
	Sunrise  string `xml:"sunr"`
	Sunset   string `xml:"suns"`
	TempMin  string `xml:"low"`
	TempMax  string `xml:"hi"`

	Part []Part `xml:"part"`

	// Template stuff
	MonthDay   int    // day number in the month.
	TxtSunrise string // formated (h24) Sunrise
	TxtSunset  string // formated (h24) Sunset
}

// Part defines weather forecast data for one part (day or night) of a following day.
//
type Part struct {
	Period             string `xml:"p,attr"`
	WeatherDescription string `xml:"t"`
	WindDegree         string `xml:"wind>d"`
	WindSpeed          string `xml:"wind>s"`
	WindDirection      string `xml:"wind>t"`
	Humidity           string `xml:"hmid"`
	WeatherIcon        string `xml:"icon"`
	PrecipitationProba string `xml:"ppcp"`
	WeatherShortDesc   string `xml:"bt"`
	Gust               string `xml:"wind>gust"`
}

// DayPart returns a part (day or night) weather forecast data.
//
func (wc *Forecast) DayPart(dayNum int, getNight bool) *Part {
	if dayNum >= len(wc.Days) {
		return nil
	}

	search := "d"
	if getNight {
		search = "n"
	}

	for i, part := range wc.Days[dayNum].Part {
		if part.Period == search {
			return &wc.Days[dayNum].Part[i]
		}
	}
	return nil
}

// Format returns the template formatted string for the given day.
//
func (wc *Forecast) Format(template *cdtype.Template, dayNum int, time24H bool) (string, error) {
	wc.Part = wc.DayPart(dayNum, false)
	if wc.Part == nil {
		return "", fmt.Errorf("data missing forecast day: %d  (have %d)", dayNum, len(wc.Days))
	}

	wc.Template = template
	wc.Day = &wc.Days[dayNum]

	return FormatTemplate(template, "Forecast", wc)
}

// Fields format a list of fields from the template.
//
func (wc *Forecast) Fields(list ...string) (string, error) {
	return FormatField(wc.Template, list, wc)
}

//
//------------------------------------------------------------------[ FORMAT ]--

func display(str string) string {
	if str == "" || str == "N" {
		return "?"
	}
	return str
}

// Prepend degree symbol to unit if missing.
//
func unitDegree(str string) string {
	if !strings.HasPrefix(str, DegreeSymbol) {
		return DegreeSymbol + str
	}
	return str
}

// FormatTemplate formats the given data object with the given template function name.
//
func FormatTemplate(template *cdtype.Template, funcName string, data interface{}) (string, error) {
	text, e := template.ToString(funcName, data)
	return strings.Trim(text, " \n"), e
}

// FormatField formats a list of fields with the given template.
//
// It's a callback from the template, where the list of fields to display is set.
// With this method, we can translate titles, align data and allow custom
// templating for the user.
//
func FormatField(template *cdtype.Template, list []string, data interface{}) (string, error) {
	buf := bytes.NewBuffer(nil)
	tw := &tabwriter.Writer{}

	tw.Init(buf, 0, 4, 1, '\t', 0)
	needNL := false

	for _, key := range list {
		if needNL {
			fmt.Fprintln(tw)
		}
		needNL = true
		title := ""
		switch key {
		case "tempReal":
			title = trans("Temperature")

		case "tempFelt":
			title = trans("Feels like")

		case "wind":
			title = trans("Wind")

		case "humidity":
			title = trans("Humidity")

		case "pressure":
			title = trans("Pressure")

		case "sun":
			title = trans("Sunrise") + " - " + trans("Sunset")

		case "tempDay":
			title = trans("Temperature")

		case "precipitation":
			title = trans("Precipitation probability")
		}

		fmt.Fprint(tw, title+":\t")
		e := template.ExecuteTemplate(tw, key, data)
		if e != nil {
			return "", e
		}
	}
	tw.Flush()
	return strings.Trim(buf.String(), " \n"), nil
}

// FormatSun formats sunrise and sunset time and tells if it's the night.
//
func FormatSun(sunrise, sunset string, time24H bool) (sunrise2, sunset2 string, isNight bool, e error) {
	sunriseT, e := time.Parse("3:04 PM", sunrise)
	if e != nil {
		return "", "", false, e
	}
	sunsetT, e := time.Parse("3:04 PM", sunset)
	if e != nil {
		return "", "", false, e
	}

	badnow := time.Now()
	now := time.Date(0, 1, 1, badnow.Hour(), badnow.Minute(), 0, 0, time.UTC)

	isNight = now.Before(sunriseT) || now.After(sunsetT)
	timeF := timeFormat(time24H)
	return sunriseT.Format(timeF), sunsetT.Format(timeF), isNight, nil
}

func timeFormat(time24H bool) string {
	if time24H {
		return "15:04"
	}
	return "3:04 PM"
}
