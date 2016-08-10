// Package Clouds is a weather monitoring applet for Cairo-Dock.
//
// It's almost a clone of the weather app to show weather forecast icons.
//
// Added:
//   Autodetect location based on IP.
//   Shortcuts: show dialog, open Webpage, recheck
//
// Possible problem (to confirm):
//   External applets left and middle clicks disabled when
//
// Dropped, because it's impossible to do for an external app:
//
//   Always visible option and its background color
//   Render desklet in 3D?
//   Sub-dock view
//
package Clouds

import (
	"github.com/sqp/godock/libs/get/weather"
	"github.com/sqp/godock/libs/net/iplocation"

	"github.com/sqp/godock/libs/cdapplet" // Applet base.
	"github.com/sqp/godock/libs/cdglobal"
	"github.com/sqp/godock/libs/cdtype" // Applet types.

	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

//
//--------------------------------------------------------------------[ TODO ]--

// translations : dialog and gui default for strings
// theme location : conf and real
// reduce timer when fail on first try?

func trans(str string) string { return str }

//
//-------------------------------------------------------------------[ CONST ]--

const (
	minUpdateDelay     = 5  // in minutes.
	defaultUpdateDelay = 15 // in minutes.
	defaultTheme       = "Classic"

	EmblemWork  = cdtype.EmblemTopLeft
	IconWork    = "EmblemWork.svg"
	IconMissing = "default-icon"
)

//------------------------------------------------------------------[ CONFIG ]--

type appletConf struct {
	cdtype.ConfGroupIconBoth `group:"Icon"`
	weather.Config           `group:"Configuration"`
	groupConfig              `group:"Configuration"`
	groupActions             `group:"Actions"`
}

type groupConfig struct {
	UpdateDelay        int
	DialogDuration     int
	DisplayNights      bool
	DisplayTemperature bool
	WeatherTheme       string
	Renderer           string
	Desket3D           bool
}

type groupActions struct {
	ShortkeyShowCurrent string
	ShortkeyOpenWeb     string
	ShortkeyRecheck     string
	Debug               bool
}

//
//------------------------------------------------------------------[ APPLET ]--

// Applet data and controlers.
//
type Applet struct {
	cdtype.AppBase // Applet base and dock connection.

	conf    *appletConf
	weather weather.Weather
}

// NewApplet create a new applet instance.
//
func NewApplet() cdtype.AppInstance {
	app := &Applet{
		AppBase: cdapplet.New(), // Icon controler and interface to cairo-dock.
		weather: weather.New(weather.BackendWeatherCom),
	}

	app.Poller().Add(app.Check)
	app.Poller().SetPreCheck(func() { app.SetEmblem(app.FileLocation("img", IconWork), EmblemWork) })
	app.Poller().SetPostCheck(func() { app.SetEmblem("none", EmblemWork) })

	return app
}

// Init load user configuration if needed and initialise applet.
//
func (app *Applet) Init(loadConf bool) {
	app.LoadConfig(loadConf, &app.conf) // Load config will crash if fail. Expected.

	// Share the conf with the weather service.
	app.weather.SetConfig(&app.conf.Config)

	if app.conf.LocationCode == "" {
		loc, e := iplocation.Get()
		if !app.Log().Err(e, "autodetect location") {
			locations, e := weather.FindLocation(loc.City + ", " + loc.Country)
			if !app.Log().Err(e, "FindLocation") && len(locations) > 0 {
				app.conf.LocationCode = locations[0].ID // do not save, just set live value.
				app.Log().Info("autodetect location", locations[0].Name)
			}
		}
	}

	// if app.conf.WeatherTheme == "" {
	app.conf.WeatherTheme = defaultTheme
	// }
	app.conf.WeatherTheme = app.ThemePath(app.conf.WeatherTheme)

	interval := cdtype.PollerInterval(app.conf.UpdateDelay*60, defaultUpdateDelay*60)
	if interval < minUpdateDelay*60 {
		interval = minUpdateDelay * 60
	}

	// Set defaults to dock icon: display and controls.
	app.SetDefaults(cdtype.Defaults{
		Label:          app.conf.Name,
		Icon:           app.conf.Icon,
		PollerInterval: interval,
		Templates:      []string{"weather"},
		Shortkeys: []cdtype.Shortkey{
			{"Actions", "ShortkeyShowCurrent", "Show current conditions dialog", app.conf.ShortkeyShowCurrent},
			{"Actions", "ShortkeyOpenWeb", "Open Webpage", app.conf.ShortkeyOpenWeb},
			{"Actions", "ShortkeyRecheck", "Recheck now", app.conf.ShortkeyRecheck},
		},
		Debug: app.conf.Debug,
	})
}

//
//------------------------------------------------------------------[ EVENTS ]--

// DefineEvents set applet events callbacks.
//
func (app *Applet) DefineEvents(events *cdtype.Events) {

	// Left and middle click: show current weather dialog.
	// The left click is unavailable when the subdock is opened.
	//
	events.OnClick = func(int) { app.DialogWeatherCurrent() }
	events.OnMiddleClick = app.DialogWeatherCurrent

	// Subicon click: show weather forecast dialog for that day.
	//
	events.OnSubClick = func(dayNum string, _ int) { app.DialogWeatherForecast(dayNum) }

	// Right click menu. Provide actions list or registration request.
	//
	events.OnBuildMenu = func(menu cdtype.Menuer) {
		if app.conf.LocationCode != "" {
			if app.weather.Current() != nil {
				menu.AddEntry(trans("Show current conditions"), "dialog-information", app.DialogWeatherCurrent)
			}
			menu.AddEntry(trans("Open webpage"), "go-jump", func() { app.OpenWeb(0) })

			menu.AddEntry(trans("Recheck now"), "view-refresh", app.Check)
		}
		menu.AddEntry(trans("Set location"), "user-home", func() { app.AskLocationText("") })
	}

	// Launch action configured for given shortkey.
	//
	events.OnShortkey = func(key string) {
		switch key {
		case app.conf.ShortkeyShowCurrent:
			app.DialogWeatherCurrent()

		case app.conf.ShortkeyOpenWeb:
			app.OpenWeb(0)

		case app.conf.ShortkeyRecheck:
			app.Check()
		}
	}
}

//
//-----------------------------------------------------------------[ WEATHER ]--

func (app *Applet) Check() {
	e := app.weather.Get()
	if !app.Log().Err(e, "weather") {
		app.Draw()
	}
}

func (app *Applet) ThemePath(themeName string) string {
	return app.FileLocation("themes", themeName)
	// return filepath.Join(pathtoDockData/plug-ins/weather/themes, themeName)
}

// WeatherIcon returns the full path to the current weather theme icon.
//
func (app *Applet) WeatherIcon(icon string) string {
	if icon == "" {
		return IconMissing
	}
	return filepath.Join(app.conf.WeatherTheme, icon+".png")
}

func (app *Applet) OpenWeb(numDay int) {
	app.Log().ExecAsync(cdglobal.CmdOpen, app.weather.WebpageURL(numDay))
}

// SetLocationCode updates the config file with the new location and ...
// (TODO: need reload, to check).
//
func (app *Applet) SetLocationCode(locationCode string) {
	// Reset weather data from previous location.
	app.weather.Clear()

	app.conf.LocationCode = locationCode

	cu, e := app.UpdateConfig()
	if app.Log().Err(e, "UpdateConfig") {
		return
	}

	cu.Set("Configuration", "LocationCode", locationCode)
	e = cu.Save()
	if app.Log().Err(e, "UpdateConfig") {
		return
	}

	app.Log().Info("Updated LocationID", locationCode)
	app.Poller().Restart()
}

//
//-----------------------------------------------------------------[ DISPLAY ]--

// Draw new data on icon.
//
func (app *Applet) Draw() {

	// Show current info.
	if app.weather.Current() == nil {
		app.Log().NewErr("weather: missing Current data")
	} else if app.conf.DisplayCurrentIcon {
		cur := app.weather.Current()
		app.SetLabel(cur.LocName)
		if app.conf.DisplayTemperature {
			info := fmt.Sprintf("%d%s", cur.TempReal, cur.UnitTemp)
			app.SetQuickInfo(info)
			app.Log().Info("weather info", info)
		}

		// if errordata{...}

		icon := cur.WeatherIcon
		// if cur.IsNight() {
		// 	icon = strconv.Itoa(cur.MoonIcon)
		// }

		icon = app.WeatherIcon(icon)
		app.SetIcon(icon)
		app.Log().Debug("weather icon", icon)
	}

	// Show forecast.
	app.RemoveSubIcons()
	for i, day := range app.weather.Forecast().Days {
		for _, part := range day.Part {
			if !app.conf.DisplayNights && part.Period != "d" {
				continue
			}
			key := strconv.Itoa(i)
			iconName := app.WeatherIcon(part.WeatherIcon)
			app.AddSubIcon(day.DayName, iconName, key)
			icon := app.SubIcon(key)
			if icon == nil {
				app.Log().NewErr("missing day="+key, "weather get subicon")

			} else if app.conf.DisplayTemperature {
				icon.SetQuickInfo(day.TempMin + " / " + day.TempMax)
			}
		}
	}
}

//
//------------------------------------------------------------------[ DIALOG ]--

// DialogWeatherCurrent shows the current weather details.
//
func (app *Applet) DialogWeatherCurrent() {
	// test running ?
	// e := app.ShowDialog(tran(("Data are being fetched, please re-try in a few seconds.")), 30)

	// TODO: test failed
	if app.weather.Current() == nil {
		// e := app.ShowDialog(tran(("No data available\nRetrying now...")), 30)
		return
	}

	liststr, e := app.Template().Execute("weather", "ListCurrent", nil)
	if app.Log().Err(e, "Template Current") {
		return
	}

	var (
		cur  = app.weather.Current()
		out  string
		pre  string
		post string
		ok   bool
		list = strings.Split(liststr, ",")
		args = map[string]string{
			"tempReal": trans("Temperature"),
			"tempFelt": trans("Feels like"),
			"wind":     trans("Wind"),
			"humidity": trans("Humidity"),
			"pressure": trans("Pressure"),
			"sun":      trans("Sunrise") + " - " + trans("Sunset"),
		}
	)
	for _, key := range list {
		post, e = app.Template().Execute("weather", key, cur)
		app.Log().Err(e, "Template Current")
		pre, ok = args[key]
		if ok {
			out += pre + "\t"
		}
		out += post + "\n"
	}

	out = weather.AlignTab(out)    // align our tabs.
	out = strings.Trim(out, " \n") // trim spaces and endlines.

	// Show a dialog with the current conditions.
	e = app.PopupDialog(cdtype.DialogData{
		Message:    "<tt>" + out + "</tt>",
		Icon:       app.WeatherIcon(app.weather.Current().WeatherIcon),
		TimeLength: app.conf.DialogDuration,
		UseMarkup:  true,
	})
	app.Log().Err(e, "DialogWeatherCurrent")
}

// DialogWeatherForecast shows the weather details for one of the next days.
//
func (app *Applet) DialogWeatherForecast(ref string) {
	// Ensure we have a valid day reference and all its data.
	if app.weather.Forecast() == nil {
		return
	}
	dayNum, e := strconv.Atoi(ref)
	if e != nil {
		app.Log().NewErr("bad icon reference: ["+ref+"]", "DialogWeatherForecast")
		return
	}

	part := app.weather.Forecast().DayPart(dayNum, false)
	if part == nil {
		return
	}

	// Show a dialog with the forecast info.
	e = app.PopupDialog(cdtype.DialogData{
		Message:    app.weather.Forecast().Format(dayNum),
		Icon:       app.WeatherIcon(part.WeatherIcon),
		TimeLength: app.conf.DialogDuration,
		UseMarkup:  true,
	})
	app.Log().Err(e, "DialogWeatherCurrent")
}

// AskLocationText asks the user his location name.
// If confirmed, continues the selection process
//
func (app *Applet) AskLocationText(deftxt string) {
	e := app.PopupDialog(cdtype.DialogData{
		Message: trans("Enter your location:"),
		Widget: cdtype.DialogWidgetText{
			InitialValue: deftxt,
			Editable:     true,
			Visible:      true,
		},
		Buttons:  "ok;cancel",
		Callback: cdtype.DialogCallbackValidString(app.AskLocationConfirm),
	})
	app.Log().Err(e, "popup AskLocation")
}

// AskLocationConfirm search the list of cities matching the provided name
// and shows a dialog with the list of locations for the user to choose.
// When validated, the config file is updated and ...(TODO: need reload, to check).
// If no data is found, the user is sent back to the AskLocationText dialog with
// the text he provided.
//
func (app *Applet) AskLocationConfirm(locstr string) {
	locations, e := weather.FindLocation(locstr)
	if app.Log().Err(e, "FindLocation") {
		app.ShowDialog("Find location: "+e.Error(), 10)
		return
	}

	if len(locations) == 0 { // Try again.
		app.AskLocationText(locstr)
		return
	}

	var ids []string
	for _, loc := range locations {
		ids = append(ids, loc.Name)
	}

	e = app.PopupDialog(cdtype.DialogData{
		Message: trans("Select your location:"),
		Widget: cdtype.DialogWidgetList{
			Values: strings.Join(ids, ";"),
		},
		Buttons: "ok;cancel",
		Callback: cdtype.DialogCallbackValidInt(func(id int) {
			app.SetLocationCode(locations[id].ID)
		}),
	})
	app.Log().Err(e, "popup AskLocation")
}
