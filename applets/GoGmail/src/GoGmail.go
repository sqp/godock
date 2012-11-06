package main

import (
	"errors"
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/dock"   // Connection to cairo-dock.
	"github.com/sqp/godock/libs/log"    // Display info in terminal.
	"github.com/sqp/godock/libs/poller" // Polling timing handler.
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Only one menu can be opened, and we want to be sure we end up on the good
// action in case a few settings might have changed (ex: monitor closed)
//~ var menuOpened int // unused

//---------------------------------------------------------------[ MAIN CALL ]--

// Program launched. Create and activate applet.
//
func main() {
	app := NewAppletGmail()
	dock.StartApplet(app.CDApplet, app, app.poller)
}

//------------------------------------------------------------------[ APPLET ]--

// Applet data and controlers.
//
type AppletGmail struct {
	*dock.CDApplet

	// Main interfaces.
	render RendererMail
	data   Mailbox
	poller *poller.Poller
	conf   *mailConf

	// Local variables.
	err error // Buffer last error to prevent displaying it twice.
}

// Create a new applet instance.
//
func NewAppletGmail() *AppletGmail {
	app := &AppletGmail{
		CDApplet: dock.Applet(), // Icon controler and interface to cairo-dock.
	}

	// Prepare mailbox with the display callback that will receive update info.
	onResult := func(i int, e error) { app.updateDisplay(i, e) }
	app.data = NewFeed(app.FileLocation(loginLocation), onResult)

	// The poller will check for new mails on a timer.
	app.poller = poller.New(func() { app.data.Check() })

	// Set updates callbacks pre and post check: Displays a small emblem during
	// the polling, and clears it after.
	app.poller.SetPreCheck(func() { app.SetEmblem(app.FileLocation("img", "go-down.svg"), cdtype.EmblemTopLeft) })
	app.poller.SetPostCheck(func() { app.SetEmblem("none", cdtype.EmblemTopLeft) })

	return app
}

// Config loading.
//
func (app *AppletGmail) getConfig() {
	app.conf = &mailConf{}

	loaded, e := dock.NewConfig(app.ConfFile)
	log.Fatal(e, "Load config")
	loaded.Parse("Icon", MailIcon{}, &app.conf.MailIcon)
	loaded.Parse("Configuration", MailConfig{}, &app.conf.MailConfig)
	loaded.Parse("Actions", MailActions{}, &app.conf.MailActions)
}

// Load user configuration if needed and initialise applet.
//
func (app *AppletGmail) Init(loadConf bool) {
	if loadConf {
		app.getConfig()
	}
	log.SetDebug(app.conf.Debug)

	// Reset data to be sure our display will be refreshed.
	app.data.Clear()
	app.err = nil

	// Fill config empty settings.
	if app.conf.MonitorName == "" {
		app.conf.MonitorName = app.conf.DefaultMonitorName
	}
	if app.conf.AlertSoundFile == "" {
		app.conf.AlertSoundFile = app.conf.DefaultAlertSoundFile
	}
	if app.conf.UpdateDelay == 0 {
		app.conf.UpdateDelay = defaultUpdateDelay
	}

	// Set defaults to dock icon: display and controls.
	def := cdtype.Defaults{
		Shortkeys: []string{app.conf.ShortkeyOpen, app.conf.ShortkeyCheck},
		Label:     "Mail unchecked",
	}
	if app.conf.Icon != "" && app.conf.Renderer != EmblemSmall && app.conf.Renderer != EmblemLarge { // User selected icon.
		def.Icon = app.conf.Icon
	}
	if app.conf.MonitorEnabled {
		def.MonitorName = app.conf.MonitorName
	}
	app.SetDefaults(def)

	// Create the renderer.
	switch app.conf.Renderer {
	case QuickInfo:
		app.render = NewRenderedQuick(app.CDApplet)

	case EmblemSmall, EmblemLarge:
		app.render = NewRenderedSVG(app.CDApplet, app.conf.Renderer)

	default: // NoDisplay case, but using default to be sure we have a valid renderer.
		app.render = NewRenderedNone()
	}

	// Configure the mail polling loop.
	app.poller.SetInterval(app.conf.UpdateDelay)
}

// Reset all settings and restart timer.
//
func (app *AppletGmail) Reload(confChanged bool) {
	log.Debug("Reload module")
	app.Init(confChanged)
	app.action(ActionCheckMail) // CheckMail recheck and reset the timer.
}

// End: Nothing to do ? Need to check DBus API.

//------------------------------------------------------------------[ EVENTS ]--

// Define applet events callbacks.
//
func (app *AppletGmail) DefineEvents() {

	// Left click: launch configured action for current user mode.
	//
	app.Events.OnClick = func() {
		app.action(app.conf.ActionClickLeft)
	}

	// Middle click: launch configured action for current user mode.
	//
	app.Events.OnMiddleClick = func() {
		app.action(app.conf.ActionClickMiddle)
	}

	// Right click menu. Provide actions list or registration request.
	//
	app.Events.OnBuildMenu = func() {
		menu := []string{""} // First entry is a separator.
		if app.data.IsValid() {
			menu = append(menu, ActionOpenClient, ActionCheckMail, ActionShowMails, "", "Change account")
		} else { // No running loop =  no registration. User will do as expected !
			menu = append(menu, "Set account")
		}
		app.PopulateMenu(menu...)
	}

	// Menu entry selected. Launch the expected action.
	//
	app.Events.OnMenuSelect = func(numEntry int32) {
		switch numEntry {
		case 1:
			app.action(ActionOpenClient)
		case 2:
			app.action(ActionCheckMail)
		case 3:
			app.action(ActionShowMails)
		case 5:
			app.askLogin()
		}
	}

	// User is providing his login informations, save to disk.
	//
	app.Events.OnAnswerDialog = func(button int32, data interface{}) {
		app.data.SaveLogin(data.(string))
		app.action(ActionCheckMail) // CheckMail will launch a check and reset the timer.
	}

	app.Events.OnShortkey = func(key string) {
		if key == app.conf.ShortkeyOpen {
			//~ app.Launch(app.conf.ShortkeyOneAction)
			app.action(ActionOpenClient)
		}
		if key == app.conf.ShortkeyCheck {
			app.action(ActionCheckMail)
		}
	}
}

//-----------------------------------------------------------------[ ACTIONS ]--

func (app *AppletGmail) action(action string) {
	// No running loop = no registration. User must comply !
	if !app.data.IsValid() {
		app.askLogin()
		return
	}

	switch action {
	case ActionOpenClient:
		app.actionOpenClient()
	case ActionShowMails:
		app.mailPopup(app.conf.DialogNbMailActionShow, false)
	case ActionCheckMail:
		// Send the refresh event to the poller. It will reset our timer and 
		// restart the loop.  that will launch a check.
		app.poller.Restart() // Send us to app.data.Check()
	}
}

// Open defined mail application or webpage. Manage application visibility if
// the user activated the application monitoring option.
//
func (app *AppletGmail) actionOpenClient() {
	haveMonitor, hasFocus := app.HaveMonitor()
	switch {
	case app.conf.MonitorEnabled && haveMonitor: // Application monitored
		app.ShowAppli(!hasFocus)
	case strings.HasPrefix(strings.ToLower(app.conf.MonitorName), "http"): // URL
		exec.Command("xdg-open", app.conf.MonitorName).Start()
	default: // Application
		exec.Command(app.conf.MonitorName).Start()
	}
}

// Show dialog with information for the given number of mails. Can display an
// additional comment about mails being new if the second param is set to true.
//
func (app *AppletGmail) mailPopup(nb int, new bool) {
	feed := app.data.Data().(*Feed)
	text := ""
	if new {
		text += strconv.Itoa(nb) + " new mails for  ::  "
	}
	text += feed.Title + "\n\n"

	if app.conf.DialogType == dialogNotify { // Temp test. libnotify seem to have a maximum size for messages.
		nb = 5
	}

	for i, mail := range feed.Mail {
		if i < nb {
			text += mail.AuthorName + "  ::  " + mail.Title + "\n"
		}
	}
	app.PopUp("Gmail", text)

	//~ app.ShowDialog(text, int32(app.conf.DialogTimer))
}

func (app *AppletGmail) askLogin() {
	text := ""
	if !app.data.IsValid() {
		text = "No account configured.\n\n"
	}
	app.AskText(text+"Please enter your login in the format username:password", "")
}

//-----------------------------------------------------------[ MAIL HANDLING ]--

// Update display callback. Receives mail check result with new messages count
// and polling error status.
//
// Update checked time and, if needed, send info or error to renderer and user
// alerts.
//
func (app *AppletGmail) updateDisplay(delta int, e error) {
	eventTime := time.Now().String()[11:19]
	label := "Checked: " + eventTime
	switch {
	case e != nil:
		label = "Update Error: " + eventTime + "\n" + e.Error() // Error time is refreshed.
		log.Err(e, "Check mail")
		if app.err == nil || e.Error() != app.err.Error() { // Error buffer, dont warn twice the same information.
			app.render.Error(e)
			app.PopUp("Mail check error", e.Error())
			app.err = e
		}

	case delta > 0:
		log.Debug("  * Count changed", delta)
		app.sendAlert(delta)

	case delta == 0:
		log.Debug("", " * no change")
	}

	switch {
	case e == nil && app.err != nil: // Error disapeared. Cleaning buffer and refresh display.
		app.render.Set(app.data.Count())
		app.err = nil

	case delta != 0: // Refresh display only if changed.
		app.render.Set(app.data.Count())
	}
	app.SetLabel(label)
}

// Mail count changed. Check if we need to warn the user.
//
func (app *AppletGmail) sendAlert(delta int) {
	if app.conf.AlertDialogEnabled {
		//~ app.mailPopup(math.Min(delta, app.conf.AlertDialogMaxNbMail), true)

		// TODO: need use  min 
		app.mailPopup(app.conf.AlertDialogMaxNbMail, true)
	}
	if app.conf.AlertAnimName != "" {
		app.Animate(app.conf.AlertAnimName, int32(app.conf.AlertAnimDuration))
	}
	if app.conf.AlertSoundEnabled {
		sound := app.conf.AlertSoundFile
		if len(sound) == 0 {
			log.Info("No sound file configured")
			return
		}
		if !filepath.IsAbs(sound) && sound[0] != []byte("~")[0] { // Check for relative path.
			sound = app.FileLocation(sound)
		}

		//~ exec.Command("paplay", sound).Start()
		if e := exec.Command("paplay", sound).Start(); e != nil {
			//~ exec.Command("aplay", sound).Start()
		}
	}
}

//---------------------------------------------------------------[ RENDERERS ]--

// RenderedNone is a stub. Used for the none choice and as a fallback for SVG 
// renderer if it failed to load its data.
//
type RenderedNone struct{}

func NewRenderedNone() *RenderedNone {
	return &RenderedNone{}
}
func (rs *RenderedNone) Set(count int) {}
func (rs *RenderedNone) Error(e error) {}

// RenderedQuick displays mail count on the icon QuickInfo.
//
type RenderedQuick struct {
	*dock.CDApplet // base applet should be enough, we only need FileLocation and SetIcon.
	pathDefault    string
}

func NewRenderedQuick(app *dock.CDApplet) *RenderedQuick {
	return &RenderedQuick{
		CDApplet:    app,
		pathDefault: app.FileLocation("img", "gmail-icon.svg"),
	}
}

func (rs *RenderedQuick) Set(count int) {
	info := ""
	if count > 0 {
		info = strconv.Itoa(count)
	}
	rs.SetQuickInfo(info)
}

func (rs *RenderedQuick) Error(e error) {
	rs.SetQuickInfo("N/A")
}

// RenderedQuick displays mail count on a hacked svg icon. The icon is rewritten
// with the new value on every change. In case of file loading problem, a new
// RenderedNone will be returned, so a valid renderer will always be provided.
//
type RenderedSVG struct {
	*dock.CDApplet // base applet should be enough, we only need FileLocation and SetIcon.
	pathDefault    string
	pathTemp       string
	pathError      string
	iconSource     string
}

func NewRenderedSVG(app *dock.CDApplet, typ string) RendererMail {
	size := strings.Split(string(typ), " ")[0]

	source, err := ioutil.ReadFile(app.FileLocation("img", size+".svg"))
	if err != nil {
		return NewRenderedNone()
	}

	rs := &RenderedSVG{
		CDApplet:    app,
		pathDefault: app.FileLocation("img", "gmail-icon.svg"),
		pathTemp:    app.FileLocation("img", "temp.svg"),
		pathError:   app.FileLocation("img", "gmail-error-"+size+".svg"),
		iconSource:  string(source),
	}

	return rs
}

func (rs *RenderedSVG) Set(count int) {
	if count == 0 { // No mail -> default icon.
		rs.SetIcon(rs.pathDefault)
	} else { // Build custom SVG.
		newfile := []byte(strings.Replace(rs.iconSource, "STRING_COUNTER", strconv.Itoa(count), -1))
		err := ioutil.WriteFile(rs.pathTemp, newfile, os.ModePerm)
		if err == nil {
			rs.SetIcon(rs.pathTemp)
		} else {
			rs.Error(err)
		}
	}
}

func (rs *RenderedSVG) Error(e error) {
	rs.SetIcon(rs.pathError)
}

//---------------------------------------------------------------[ LIBNOTIFY ]--

// libnotify call is currently stored as a global so libnotify.go can be
// removed if needed. Need to see the doc about optional dependencies building
// for better handling.
//
var popUp func(title, msg, icon string, duration int) error // store  if enabled.

// Call libnotify popup if enabled.
//
func (app *AppletGmail) popUpNotify(title, msg, icon string, duration int) {

}

// Open a popup on the configured notification systme. Valid options are internal
// or libnotify.
//
func (app *AppletGmail) PopUp(title, msg string) {
	if app.conf.DialogType == dialogInternal {
		app.ShowDialog(msg, int32(app.conf.DialogTimer))
	} else {
		var e error
		if popUp == nil {
			e = errors.New("Applet was compiled with library support disabled")
		} else {
			e = popUp(title, msg, app.FileLocation("icon"), app.conf.DialogTimer*1000)
			//~ DEBUG("notify", e==nil, e)
		}
		log.Err(e, "libnotify")
	}
}
