// Package GoGmail is a mail checker applet for Cairo-Dock.
package GoGmail

import (
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/dock" // Connection to cairo-dock.
	"github.com/sqp/godock/libs/ternary"

	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var log cdtype.Logger

// Applet data and controlers.
//
type Applet struct {
	cdtype.AppBase // Applet base and dock connection.

	// Main interfaces.
	render RendererMail
	data   Mailbox
	conf   *mailConf

	// Local variables.
	err error // Buffer for last error to prevent displaying it twice.
}

// NewApplet create a new applet instance.
//
func NewApplet() cdtype.AppInstance {
	app := &Applet{AppBase: dock.NewCDApplet()} // Icon controler and interface to cairo-dock.
	log = app.Log()

	app.defineActions()

	// Prepare mailbox with the display callback that will receive update info.
	app.data = NewFeed(app.updateDisplay)

	// The poller will check for new mails on a timer, with a small emblem during the polling.
	poller := app.AddPoller(app.data.Check)
	poller.SetPreCheck(func() { app.SetEmblem(app.FileLocation("img", "go-down.svg"), cdtype.EmblemTopLeft) })
	poller.SetPostCheck(func() { app.SetEmblem("none", cdtype.EmblemTopLeft) })

	return app
}

// Init load user configuration if needed and initialise applet.
//
func (app *Applet) Init(loadConf bool) {
	app.LoadConfig(loadConf, &app.conf) // Load config will crash if fail. Expected.

	// Reset data to be sure our display will be refreshed.
	app.data.Clear()
	app.data.LoadLogin(app.FileDataDir(loginLocation))
	app.err = nil

	// Define the mail client action.
	if app.conf.MailClientName == "" { //  Set default to webpage if not provided.
		app.conf.MailClientAction = MailClientLocation
		app.conf.MailClientName = app.conf.DefaultMonitorName
	}

	// Fill config empty settings.
	if app.conf.AlertSoundFile == "" {
		app.conf.AlertSoundFile = app.conf.DefaultAlertSoundFile
	}
	var icon string
	if app.conf.Icon != "" && app.conf.Renderer != EmblemSmall && app.conf.Renderer != EmblemLarge { // User selected icon.
		icon = app.conf.Icon
	}

	// Set defaults to dock icon: display and controls.
	app.SetDefaults(cdtype.Defaults{
		Label:          "Mail unchecked",
		Icon:           icon,
		Templates:      []string{DialogTemplate},
		PollerInterval: cdtype.PollerInterval(app.conf.UpdateDelay*60, defaultUpdateDelay),
		Commands: cdtype.Commands{
			"mailClient": cdtype.NewCommandStd(app.conf.MailClientAction+1, app.conf.MailClientName, app.conf.MailClientClass)}, // Add 1 to action as we don't provide the none option.
		Shortkeys: []cdtype.Shortkey{
			{"Actions", "ShortkeyOpen", "Open mail client", app.conf.ShortkeyOpen},
			{"Actions", "ShortkeyCheck", "Check mails now", app.conf.ShortkeyCheck}},
		Debug: app.conf.Debug})

	// Create the renderer.
	switch app.conf.Renderer {
	case QuickInfo:
		app.render = NewRenderedQuick(app)

	case EmblemSmall, EmblemLarge:
		var e error
		app.render, e = NewRenderedSVG(app, app.conf.Renderer)
		log.Err(e, "renderer svg")

	default: // NoDisplay case, but using default to be sure we have a valid renderer.
		app.render = NewRenderedNone()
	}
}

//
//------------------------------------------------------------------[ EVENTS ]--

// DefineEvents set applet events callbacks.
//
func (app *Applet) DefineEvents(events *cdtype.Events) {

	// Left click: try to launch configured action.
	//
	events.OnClick = func() {
		app.testAction(app.ActionID(app.conf.ActionClickLeft))
	}

	// Middle click: try to launch configured action.
	//
	events.OnMiddleClick = func() {
		app.testAction(app.ActionID(app.conf.ActionClickMiddle))
	}

	// Right click menu. Provide actions list or registration request.
	//
	events.OnBuildMenu = func(menu cdtype.Menuer) {
		haveApp, _ := app.HaveMonitor()
		switch {
		case !app.data.IsValid(): // No running loop =  no registration. User will do as expected !
			app.BuildMenu(menu, menuRegister)

		case haveApp: // Monitored application opened.
			app.BuildMenu(menu, menuFull[1:]) // Drop "Open client" option, already provided as window action by the dock.

		default:
			app.BuildMenu(menu, menuFull)
		}
	}

	// Launch action configured for given shortkey.
	//
	events.OnShortkey = func(key string) {
		if key == app.conf.ShortkeyOpen {
			app.testAction(ActionOpenClient)
		}
		if key == app.conf.ShortkeyCheck {
			app.testAction(ActionCheckMail)
		}
	}
}

//
//-----------------------------------------------------------------[ ACTIONS ]--

// Define applet actions. Order must match actions const declaration order.
//
func (app *Applet) defineActions() {
	app.ActionAdd(
		&cdtype.Action{
			ID:   ActionNone,
			Menu: cdtype.MenuSeparator,
		},
		&cdtype.Action{
			ID:   ActionOpenClient,
			Name: "Open mail client",
			Icon: "document-open",
			Call: app.actionOpenClient,
		},
		&cdtype.Action{
			ID:       ActionCheckMail,
			Name:     "Check now",
			Icon:     "view-refresh",
			Call:     app.actionCheckMail,
			Threaded: true,
		},
		&cdtype.Action{
			ID:       ActionShowMails,
			Name:     "Show mail dialog",
			Icon:     "media-seek-forward",
			Call:     app.actionShowMails,
			Threaded: true,
		},
		&cdtype.Action{
			ID:       ActionRegister,
			Name:     "Set account",
			Icon:     "media-seek-forward",
			Call:     app.actionRegister,
			Threaded: true,
		},
	)
}

// Test login infos before launching an action. Redirect to the the registration
// if failed.
//
func (app *Applet) testAction(id int) {
	if app.data.IsValid() {
		app.ActionLaunch(id)
	} else {
		app.ActionLaunch(ActionRegister) // No running loop = no registration. User must comply !
	}
}

// Open defined mail application or webpage. Manage application visibility if
// the user activated the application monitoring option.
//
func (app *Applet) actionOpenClient() {
	app.CommandLaunch("mailClient")
}

// Send the refresh event to the poller. It will reset our timer and
// restart the loop.  that will launch a check.
//
func (app *Applet) actionCheckMail() {
	app.Poller().Restart() // Should trigger a app.data.Check()
}

// Show dialog with informations on last mails.
// Infinite duration as it's from a user request.
//
func (app *Applet) actionShowMails() {
	app.mailPopup(app.conf.DialogNbMail, 0, "ListMailsManual")
}

// Request login informations from user. Popup an AskText dialog.
// Save to disk and try to get new data if confirmed.
//
func (app *Applet) actionRegister() {
	text := ternary.String(app.data.IsValid(), "", "No account configured.\n\n")
	e := app.PopupDialog(cdtype.DialogData{
		Message: text + "Please enter your login in the format username:password",
		Buttons: "ok;cancel",
		Widget: cdtype.DialogWidgetText{
			Editable: true,
			Visible:  true,
		},
		Callback: func(button int, data interface{}) {
			str, ok := data.(string)
			if !ok || (button != cdtype.DialogButtonFirst && button != cdtype.DialogKeyEnter) {
				return
			}
			app.data.SaveLogin(str)
			app.ActionLaunch(ActionCheckMail) // CheckMail will launch a check and reset the timer.
		},
	})

	log.Err(e, "popup")
}

//
//-----------------------------------------------------------[ MAIL HANDLING ]--

// Update display callback. Receives mail check result with new messages count
// and polling error status.
//
// Update checked time and, if needed, send info or error to renderer and user
// alerts.
//
func (app *Applet) updateDisplay(delta int, first bool, e error) {
	eventTime := time.Now().String()[11:19]
	label := "Checked: " + eventTime
	switch {
	case e != nil:
		label = "Update Error: " + eventTime + "\n" + e.Error() // Error time is refreshed.
		log.Err(e, "Check mail")
		if app.err == nil || e.Error() != app.err.Error() { // Error buffer, dont warn twice the same information.
			app.render.Error(e)
			app.ShowDialog("Mail check error"+e.Error(), int32(app.conf.DialogTimer))
			// app.PopUp("Mail check error", e.Error())
			app.err = e
		}

	case first:
		log.Debug("  * First check", delta)

	case delta > 0:
		log.Debug("  * Count changed", delta)
		app.sendAlert(delta)

	case delta == 0:
		log.Debug("  * ", "no change")
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
func (app *Applet) sendAlert(delta int) {
	if app.conf.AlertDialogEnabled {
		nb := ternary.Min(delta, app.conf.DialogNbMail)
		app.mailPopup(nb, app.conf.DialogTimer, "ListMailsNew")
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

		log.ExecAsync("paplay", sound)
		// if e := exec.Command("paplay", sound).Start(); e != nil {
		//~ exec.Command("aplay", sound).Start()
		// }
	}
}

// Show dialog with information for the given number of mails. Can display an
// additional comment about mails being new if the second param is set to true.
//
func (app *Applet) mailPopup(nb, duration int, template string) {
	// feed := app.data.Data().(*Feed)
	feed := app.data.(*Feed)

	// Prepare data for template formater.
	feed.New = nb
	feed.Plural = feed.New > 1
	max := ternary.Min(feed.New, len(feed.Mail))
	feed.MailsNew = make([]*Email, max)
	for i := 0; i < max; i++ {
		feed.MailsNew[i] = feed.Mail[i]
	}

	// if app.conf.DialogType == dialogInternal {
	text, e := app.ExecuteTemplate(DialogTemplate, template, feed)
	if log.Err(e, "Template ListMailsNew") {
		return
	}

	e = app.PopupDialog(cdtype.DialogData{
		Message:    strings.TrimRight(text, "\n"), // Remove last EOL if any (from template range).
		TimeLength: duration,
		UseMarkup:  true,
		Buttons:    "document-open;cancel",
		Callback:   cdtype.DialogCallbackIsOK(app.ActionCallback(ActionOpenClient)), // Open mail client if the user press the 1st button.
		// Widget:     interface{} ,
	})
	log.Err(e, "popup")

	// } else {
	// 	if nb == 1 {
	// 		log.Err(popUp(feed.Mail[0].AuthorName, feed.Mail[0].Title, app.FileLocation("icon"), app.conf.DialogTimer*1000), "libnotify")
	// 	} else {
	// 		title, eTit := app.ExecuteTemplate(DialogTemplate, "TitleCount", feed)
	// 		log.Err(eTit, "Template TitleCount")
	// 		text, eTxt := app.ExecuteTemplate(DialogTemplate, "ListMails", feed)
	// 		log.Err(eTxt, "Template ListMails")
	// 		log.Err(popUp(title, text, app.FileLocation("icon"), app.conf.DialogTimer*1000), "Libnotify")
	// 	}
	// }

	// app.PopUp("Gmail", text)

	return
}

//
//---------------------------------------------------------------[ RENDERERS ]--

// RenderedNone is a stub. Used for the none choice and as a fallback for SVG
// renderer if it failed to load its data.
//
type RenderedNone struct{}

// NewRenderedNone create a new null renderer.
//
func NewRenderedNone() *RenderedNone {
	return &RenderedNone{}
}

// Set counter value.
func (rs *RenderedNone) Set(count int) {}

// Error display.
func (rs *RenderedNone) Error(e error) {}

// RenderedQuick displays mail count on the icon QuickInfo.
//
type RenderedQuick struct {
	dock.RenderSimple // Controler to the Cairo-Dock icon.
	pathDefault       string
}

// NewRenderedQuick create a new text renderer for quick-info.
//
func NewRenderedQuick(app dock.RenderSimple) *RenderedQuick {
	return &RenderedQuick{
		RenderSimple: app,
		pathDefault:  app.FileLocation("img", "gmail-icon.svg"),
	}
}

// Set counter value.
//
func (rs *RenderedQuick) Set(count int) {
	info := ""
	if count > 0 {
		info = strconv.Itoa(count)
	}
	rs.SetQuickInfo(info)
}

// Error display.
//
func (rs *RenderedQuick) Error(e error) {
	rs.SetQuickInfo("N/A")
}

// RenderedSVG displays mail count in a hacked svg icon. The icon is rewritten
// with the new value on every change. In case of file loading problem, a new
// RenderedNone will be returned, so a valid renderer will always be provided.
//
type RenderedSVG struct {
	dock.RenderSimple // Controler to the Cairo-Dock icon.
	pathDefault       string
	pathTemp          string
	pathError         string
	iconSource        string
}

// NewRenderedSVG create a new SVG image renderer.
//
func NewRenderedSVG(app dock.RenderSimple, typ string) (RendererMail, error) {
	size := strings.Split(string(typ), " ")[0]

	source, e := ioutil.ReadFile(app.FileLocation("img", size+".svg"))
	if e != nil {
		return NewRenderedNone(), e
	}

	rs := &RenderedSVG{
		RenderSimple: app,
		pathDefault:  app.FileLocation("img", "gmail-icon.svg"),
		// pathTemp:   app.FileLocation("img", "temp.svg"),
		pathError:  app.FileLocation("img", "gmail-error-"+size+".svg"),
		iconSource: string(source),
	}

	f, et := ioutil.TempFile("", "cairo-dock-gogmail-icon-") // Need to create a new temp file
	if et != nil {
		return NewRenderedNone(), e
	}

	rs.pathTemp = f.Name()
	f.Close()
	// TODO: remove tempfile.

	return rs, nil
}

// Set counter value.
//
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

// Error display.
//
func (rs *RenderedSVG) Error(e error) {
	rs.SetIcon(rs.pathError)
}
