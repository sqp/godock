package weather

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"
)

type Weather interface {
	WebpageURL(numDay int) string
	Get() error
	Current() *Current
	Forecast() *Forecast
	SetConfig(*Config)
	Clear()
}

var (
	ErrMissingConfObject   = errors.New("weather #1: " + trans("conf object missing"))
	ErrMissingLocationCode = errors.New("weather #2: " + trans("location missing"))
	ValueMissing           = trans("N/A")
)

type BackendID int

const (
	BackendWeatherCom BackendID = iota
)

// Other possible backends.

// https://github.com/schachmat/wego            -- terminal weather client       --  forecast.io account ||  Worldweatheronline account
// https://github.com/jfrazelle/weather         -- Weather via the command line  --  forecast.io API
// https://github.com/sramsay/wu                -- fast command-line application --  http://www.wunderground.com/weather/api/.
// https://github.com/benschw/weather-go        -- demo app                      --  openweathermap
// https://github.com/msoap/yandex-weather-cli  --  Command line interface       --  Yandex weather service

//

func New(backend BackendID) Weather {
	return &weatherCom{}
}

type Config struct {
	LocationCode       string // hidden in conf
	UseCelcius         bool
	DisplayCurrentIcon bool // current weather.
	NbDays             int  // forecast (next days).
}

//
//--------------------------------------------------------[ TRANSLATION TODO ]--

func trans(str string) string { return str }

//
//-----------------------------------------------------------------[ CURRENT ]--

// Feed contains Gmail inbox data. Some fields are disabled because they are
// unused. They could be enabled simply by uncommenting them.
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

	Updated      time.Time // parsed from UpdateTime for templating.
	UpdateHour   int
	UpdateMinute int
}

func (wc *Current) IsNight() bool {
	sunrise, e := time.Parse("3:04 PM", wc.Sunrise)
	if e != nil {
		println("weather parse time sunrise: " + e.Error())
		return false
	}

	sunset, e := time.Parse("3:04 PM", wc.Sunset)
	if e != nil {
		println("weather parse time sunset: " + e.Error())
		return false
	}

	badnow := time.Now()
	now := time.Date(0, 1, 1, badnow.Hour(), badnow.Minute(), 0, 0, time.UTC)

	return now.Before(sunrise) || now.After(sunset)
}

//
//----------------------------------------------------------------[ FORECAST ]--

type Forecast struct {
	Ver          string `xml:"ver,attr"`
	Ur           string `xml:"head>ur"`
	UnitDistance string `xml:"head>ud"`
	UnitPressure string `xml:"head>up"`
	UnitSpeed    string `xml:"head>us"`
	UnitTemp     string `xml:"head>ut"`
	Locale       string `xml:"head>locale"`
	Form         string `xml:"head>form"`
	Suns         string `xml:"loc>suns"`
	Tm           string `xml:"loc>tm"`
	Dnam         string `xml:"loc>dnam"`
	Lon          string `xml:"loc>lon"`
	Lat          string `xml:"loc>lat"`
	Zone         string `xml:"loc>zone"`
	Sunr         string `xml:"loc>sunr"`

	UpdateTime string `xml:"dayf>lsup"`
	Days       []Day  `xml:"dayf>day"`
}

type Day struct {
	Date     string `xml:"dt,attr"`
	DayName  string `xml:"t,attr"`
	DayCount string `xml:"d,attr"`
	Sunrise  string `xml:"sunr"`
	Sunset   string `xml:"suns"`
	TempMin  string `xml:"low"`
	TempMax  string `xml:"hi"`

	Part []Part `xml:"part"`
}

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

func (wc *Forecast) Format(dayNum int) string {
	part := wc.DayPart(dayNum, false)
	if part == nil {
		return fmt.Sprintf("data missing forecast day: %d  (have %d)", dayNum, len(wc.Days))
	}

	if part.WeatherDescription == "" {
		part.WeatherDescription = ValueMissing
	}

	day := wc.Days[dayNum]

	// Format in tab-separated columns with a tab size of 4.
	buf := bytes.NewBuffer(nil)
	w := &tabwriter.Writer{}
	w.Init(buf, 0, 4, 0, '\t', 0)

	fmt.Fprintf(w, " %s\t%s%s -> %s%s\n", trans("Temperature"), display(day.TempMin), wc.UnitTemp, display(day.TempMax), wc.UnitTemp)
	fmt.Fprintf(w, " %s\t%s%%\n", trans("Precipitation probability"), display(part.PrecipitationProba))
	fmt.Fprintf(w, " %s\t%s%s (%s)\n", trans("Wind"), display(part.WindSpeed), wc.UnitSpeed, display(part.WindDirection))
	fmt.Fprintf(w, " %s\t%s%%\n", trans("Humidity"), display(part.Humidity))
	fmt.Fprintf(w, " %s - %s\t%s - %s", trans("Sunrise"), trans("Sunset"), display(day.Sunrise), display(day.Sunset))
	w.Flush()

	return fmt.Sprintf("<big><b>%s</b>, %s - <b>%s</b></big>\n\n<tt>%s</tt>",
		day.DayName, day.Date, part.WeatherDescription, buf.String())
}

func display(str string) string {
	if str == "" || str == "N" {
		return "?"
	}
	return str
}

// Prepend degree symbol to unit if missing.
func unitDegree(str string) string {
	if !strings.HasPrefix(str, DegreeSymbol) {
		return DegreeSymbol + str
	}
	return str
}

func AlignTab(s string) string {
	buf := bytes.NewBuffer(nil)
	w := &tabwriter.Writer{}
	w.Init(buf, 0, 4, 1, '\t', 0)
	fmt.Fprintln(w, s)
	w.Flush()
	return buf.String()
}
