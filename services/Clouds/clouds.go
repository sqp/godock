// Package Clouds is a weather monitoring applet for Cairo-Dock.
//
// It's almost a clone of the weather app to show weather forecast icons.
//
// Added:
//   Autodetect location based on IP.
//   Shortcuts: show dialog today, show tomorrow, open Webpage, recheck, set location.
//   Editable template.
//
//
// Possible problem (to confirm):
//   External applets left and middle clicks disabled when a subdock is set.
//     For me, even when the subdock is removed, clicks actions aren't restored.
//     This can be tested by disabling the subdock (set forecast days to 0).
//
// Dropped, because it's impossible to do for an external app:
//     (but it could be possible to try in dock mode.
//      The problem would be the config file differences between both)
//
//   Always visible option and its background color.
//   Render desklet in 3D
//   Sub-dock view type.
//
package Clouds

import (
	"github.com/sqp/godock/libs/get/weather"
	"github.com/sqp/godock/libs/net/iplocation"

	"github.com/sqp/godock/libs/cdapplet" // Applet base.
	"github.com/sqp/godock/libs/cdglobal"
	"github.com/sqp/godock/libs/cdtype" // Applet types.
	"github.com/sqp/godock/libs/ternary"

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

// reenable show current dialog when show on icon is disabled (need to dl data)
// maybe other backends (at least one to provide a fallback and comparisons)

func trans(str string) string { return str }

//
//-------------------------------------------------------------------[ CONST ]--

const (
	minUpdateDelay     = 5         // in minutes.
	defaultUpdateDelay = 15        // in minutes.
	defaultTheme       = "Classic" // dir located in applet "themes" subdir.

	IconWork    = "EmblemWork.svg" // file located in applet "img" subdir.
	IconMissing = "default-icon"

	EmblemWork = cdtype.EmblemTopLeft
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
	DialogTemplate     cdtype.Template `default:"weather"`
}

type groupActions struct {
	ShortkeyShowCurrent  cdtype.Shortkey `desc:"Show current conditions dialog"`
	ShortkeyShowTomorrow cdtype.Shortkey `desc:"Show conditions for tomorrow"`
	ShortkeyOpenWeb      cdtype.Shortkey `desc:"Open webpage"`
	ShortkeyRecheck      cdtype.Shortkey `desc:"Recheck now"`
	ShortkeySetLocation  cdtype.Shortkey `desc:"Set location"`
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
	app.defineActions()

	app.Poller().Add(app.Check)
	app.Poller().SetPreCheck(func() { app.SetEmblem(app.FileLocation("img", IconWork), EmblemWork) })
	app.Poller().SetPostCheck(func() { app.SetEmblem("none", EmblemWork) })

	return app
}

// Init load user configuration if needed and initialise applet.
//
func (app *Applet) Init(loadConf bool) {
	app.LoadConfig(loadConf, &app.conf) // Load config will crash if fail. Expected.

	app.weather.Clear()

	// Share the conf with the weather service.
	app.weather.SetConfig(&app.conf.Config)

	// if app.conf.WeatherTheme == "" {
	app.conf.WeatherTheme = defaultTheme
	// }
	app.conf.WeatherTheme = app.ThemePath(app.conf.WeatherTheme)

	interval := cdtype.PollerInterval(app.conf.UpdateDelay*60, defaultUpdateDelay*60)
	if interval < minUpdateDelay*60 {
		interval = minUpdateDelay * 60
	}

	if app.conf.LocationCode == "" {
		app.DetectLocation()
	}

	// Set defaults to dock icon: display and controls.
	app.SetDefaults(cdtype.Defaults{
		Label:          app.conf.Name,
		Icon:           app.conf.Icon,
		PollerInterval: interval,
		ShortkeyActions: []cdtype.ShortkeyAction{
			{ActionShowCurrent, app.conf.ShortkeyShowCurrent},
			{ActionShowTomorrow, app.conf.ShortkeyShowTomorrow},
			{ActionOpenWebpage, app.conf.ShortkeyOpenWeb},
			{ActionRecheckNow, app.conf.ShortkeyRecheck},
			{ActionSetLocation, app.conf.ShortkeySetLocation},
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
		var items []int
		if app.weather.Current() != nil || app.weather.Forecast() != nil {
			if app.weather.Current() != nil {
				items = append(items, ActionShowCurrent)
			}
			items = append(items, ActionOpenWebpage, ActionRecheckNow)
		}
		items = append(items, ActionSetLocation)
		app.Action().BuildMenu(menu, items)
	}
}

//
//-----------------------------------------------------------------[ ACTIONS ]--

// List of actions defined in this applet.
// Actions order in this list must match the order in defineActions.
//
const (
	ActionNone = iota
	ActionShowCurrent
	ActionShowTomorrow
	ActionOpenWebpage
	ActionRecheckNow
	ActionSetLocation
)

// Define applet actions.
// Actions order in this list must match the order of defined actions numbers.
//
func (app *Applet) defineActions() {
	// app.Action().SetMax(1)
	app.Action().Add(
		&cdtype.Action{
			ID:   ActionNone,
			Menu: cdtype.MenuSeparator,
		},
		&cdtype.Action{
			ID:   ActionShowCurrent,
			Name: "Show current conditions",
			Icon: "dialog-information",
			Call: app.DialogWeatherCurrent,
		},
		&cdtype.Action{
			ID:   ActionShowTomorrow,
			Name: "Show conditions for tomorrow",
			Icon: "dialog-information",
			Call: func() { app.DialogWeatherForecast("1") },
		},
		&cdtype.Action{
			ID:   ActionOpenWebpage,
			Name: "Open webpage",
			Icon: "go-jump",
			Call: func() { app.OpenWeb(0) },
		},
		&cdtype.Action{
			ID:   ActionRecheckNow,
			Name: "Recheck now",
			Icon: "view-refresh",
			Call: app.Check,
		},
		&cdtype.Action{
			ID:   ActionSetLocation,
			Name: "Set location",
			Icon: "user-home",
			Call: func() { app.AskLocationText("") },
		},
	)
}

//
//-----------------------------------------------------------------[ WEATHER ]--

func (app *Applet) Check() {
	var (
		fail  int
		count int
		errs  []string
	)
	for e := range app.weather.Get() {
		count++
		if app.Log().Err(e, "get data") {
			fail++
			errs = append(errs, e.Error())
		}
	}
	if fail == 0 {
		app.Draw()
	} else {
		all := ternary.String(fail == count, " All failed", "")
		msg := "Get weather errors:" + all + "\n" + strings.Join(errs, "\n")
		app.ShowDialog(msg, app.conf.DialogDuration)
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
func (app *Applet) SetLocationCode(locationCode, locationName string) {
	// Reset weather data from previous location.
	app.weather.Clear()

	app.conf.LocationCode = locationCode

	//  Autodetect location if missing. Won't be saved.
	if locationCode == "" {
		app.DetectLocation()
	}

	cu, e := app.UpdateConfig()
	if app.Log().Err(e, "UpdateConfig") {
		return
	}

	cu.Set("Configuration", "LocationCode", locationCode)
	cu.Set("Configuration", "LocationName", locationName)
	e = cu.Save()
	if app.Log().Err(e, "UpdateConfig") {
		return
	}

	app.Log().Info("Updated LocationID", locationCode)
	app.Poller().Restart()
}

func (app *Applet) DetectLocation() {
	loc, e := iplocation.Get()
	if app.Log().Err(e, "autodetect location") {
		return
	}
	locations, e := weather.FindLocation(loc.City + ", " + loc.Country)
	if app.Log().Err(e, "FindLocation") || len(locations) == 0 {
		return
	}
	app.conf.LocationCode = locations[0].ID // do not save, just set live value.
	app.Log().Debug("autodetect location", locations[0].Name)
}

//
//-----------------------------------------------------------------[ DISPLAY ]--

// Draw new data on icon.
//
func (app *Applet) Draw() {

	// Show current info.
	if app.weather.Current() == nil {
		app.Log().NewErr("missing Current data")
	} else if app.conf.DisplayCurrentIcon {
		cur := app.weather.Current()
		app.SetLabel(cur.LocName)
		if app.conf.DisplayTemperature {
			info := fmt.Sprintf("%d%s", cur.TempReal, cur.UnitTemp)
			app.SetQuickInfo(info)
			app.Log().Debug(info, cur.WeatherDescription)
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
	message, e := app.weather.Current().Format(&app.conf.DialogTemplate)
	if app.Log().Err(e, "template current") {
		return
	}

	app.Log().Debug("message", message)

	// Show a dialog with the current conditions.
	e = app.PopupDialog(cdtype.DialogData{
		Message:    message,
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

	message, e := app.weather.Forecast().Format(&app.conf.DialogTemplate, dayNum, app.conf.Time24H)
	if app.Log().Err(e, "template forecast") {
		// 	return
	}
	part := app.weather.Forecast().DayPart(dayNum, false)

	// Show a dialog with the forecast info.
	e = app.PopupDialog(cdtype.DialogData{
		Message:    message,
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
		Message: trans("Enter your location:") + "\n\n" +
			trans("Leave empty to autodetect."),
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
	if locstr == "" {
		app.SetLocationCode("", "*AUTODETECT*")
		return
	}
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
			app.SetLocationCode(locations[id].ID, locations[id].Name)
		}),
	})
	app.Log().Err(e, "popup AskLocation")
}
