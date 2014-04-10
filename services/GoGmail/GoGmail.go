// Package GoGmail is a mail checker applet for the Cairo-Dock project.
package GoGmail

import (
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/dock" // Connection to cairo-dock.
	"github.com/sqp/godock/libs/log"  // Display info in terminal.

	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var logger *log.Log

// Applet data and controlers.
//
type Applet struct {
	*dock.CDApplet

	// Main interfaces.
	render RendererMail
	data   Mailbox
	conf   *mailConf

	// Only one menu can be opened, and we want to be sure we end up on the good
	// action in case a few settings might have changed (ex: monitor closed)
	menuOpened []int

	// Local variables.
	err error // Buffer for last error to prevent displaying it twice.
}

// NewApplet create a new applet instance.
//
func NewApplet() dock.AppletInstance {
	app := &Applet{CDApplet: dock.NewCDApplet()} // Icon controler and interface to cairo-dock.

	app.defineActions()

	// Prepare mailbox with the display callback that will receive update info.
	app.data = NewFeed(app.updateDisplay)

	// The poller will check for new mails on a timer.
	poller := app.AddPoller(app.data.Check)

	// Display a small emblem during the polling, and clear it after.
	poller.SetPreCheck(func() { app.SetEmblem(app.FileLocation("img", "go-down.svg"), cdtype.EmblemTopLeft) })
	poller.SetPostCheck(func() { app.SetEmblem("none", cdtype.EmblemTopLeft) })

	logger = app.Log
	return app
}

// Init load user configuration if needed and initialise applet.
//
func (app *Applet) Init(loadConf bool) {
	app.LoadConfig(loadConf, &app.conf) // Load config will crash if fail. Expected.

	// Reset data to be sure our display will be refreshed.
	app.data.Clear()
	app.data.LoadLogin(app.FileLocation(loginLocation))
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
	app.SetDefaults(dock.Defaults{
		Shortkeys:      []string{app.conf.ShortkeyOpen, app.conf.ShortkeyCheck},
		Label:          "Mail unchecked",
		Icon:           icon,
		Templates:      []string{DialogTemplate},
		PollerInterval: dock.PollerInterval(app.conf.UpdateDelay*60, defaultUpdateDelay),
		Commands: dock.Commands{
			"mailClient": dock.NewCommandStd(app.conf.MailClientAction+1, app.conf.MailClientName, app.conf.MailClientClass)}, // Add 1 to action as we don't provide the none option.
		Debug: app.conf.Debug})

	// Create the renderer.
	switch app.conf.Renderer {
	case QuickInfo:
		app.render = NewRenderedQuick(app.CDApplet)

	case EmblemSmall, EmblemLarge:
		app.render = NewRenderedSVG(app.CDApplet, app.conf.Renderer)

	default: // NoDisplay case, but using default to be sure we have a valid renderer.
		app.render = NewRenderedNone()
	}

	// // Check libnotify library.
	// if app.conf.DialogType == dialogNotify && popUp == nil {
	// 	logger.Info("Can't use Desktop Notifications dialogs. Applet compiled without library support")
	// 	app.conf.DialogType = dialogInternal
	// }
}

//
//------------------------------------------------------------------[ EVENTS ]--

// DefineEvents set applet events callbacks.
//
func (app *Applet) DefineEvents() {

	// Left click: try to launch configured action.
	//
	app.Events.OnClick = func() {
		app.testAction(app.Actions.ID(app.conf.ActionClickLeft))
	}

	// Middle click: try to launch configured action.
	//
	app.Events.OnMiddleClick = func() {
		app.testAction(app.Actions.ID(app.conf.ActionClickMiddle))
	}

	// Right click menu. Provide actions list or registration request.
	//
	app.Events.OnBuildMenu = func() {
		haveApp, _ := app.HaveMonitor()
		switch {
		case !app.data.IsValid(): // No running loop =  no registration. User will do as expected !
			app.menuOpened = menuRegister

		case haveApp: // Monitored application opened.
			app.menuOpened = menuFull[1:] // Drop "Open client" option, already provided by the dock.

		default:
			app.menuOpened = menuFull
		}

		app.BuildMenu(app.menuOpened)
	}

	// Menu entry selected. Launch the expected action.
	//
	app.Events.OnMenuSelect = func(numEntry int32) {
		app.Actions.Launch(app.menuOpened[numEntry])
	}

	// User is providing his login informations, save to disk.
	//
	app.Events.OnAnswerDialog = func(button int32, data interface{}) {
		app.data.SaveLogin(data.(string))
		app.Actions.Launch(ActionCheckMail) // CheckMail will launch a check and reset the timer.
	}

	// Launch action configured for given shortkey.
	//
	app.Events.OnShortkey = func(key string) {
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
	app.Actions.Add(
		&dock.Action{
			ID: ActionNone,
			// Icontype: 2,
			Menu: cdtype.MenuSeparator,
		},
		&dock.Action{
			ID:   ActionOpenClient,
			Name: "Open mail client",
			Icon: "gtk-open",
			Call: func() { app.actionOpenClient() },
		},
		&dock.Action{
			ID:       ActionCheckMail,
			Name:     "Check now",
			Icon:     "gtk-refresh",
			Call:     func() { app.actionCheckMail() },
			Threaded: true,
		},
		&dock.Action{
			ID:       ActionShowMails,
			Name:     "Show mail dialog",
			Icon:     "gtk-media-forward",
			Call:     func() { app.actionShowMails() },
			Threaded: true,
		},
		&dock.Action{
			ID:       ActionRegister,
			Name:     "Set account",
			Icon:     "gtk-media-forward",
			Call:     func() { app.actionRegister() },
			Threaded: true,
		},
	)
}

// Test login infos before launching an action. Redirect to the the registration
// if failed.
//
func (app *Applet) testAction(id int) {
	if app.data.IsValid() {
		app.Actions.Launch(id)
	} else {
		app.Actions.Launch(ActionRegister) // No running loop = no registration. User must comply !
	}
}

// Open defined mail application or webpage. Manage application visibility if
// the user activated the application monitoring option.
//
func (app *Applet) actionOpenClient() {
	app.LaunchCommand("mailClient")
}

// Send the refresh event to the poller. It will reset our timer and
// restart the loop.  that will launch a check.
//
func (app *Applet) actionCheckMail() {
	app.Poller().Restart() // Should trigger a app.data.Check()
}

// Show dialog with informations on last mails.
//
func (app *Applet) actionShowMails() {
	app.mailPopup(app.conf.DialogNbMail, "ListMailsManual")
}

// Request login informations from user. Popup an AskText dialog.
//
func (app *Applet) actionRegister() {
	text := ""
	if !app.data.IsValid() {
		text = "No account configured.\n\n"
	}
	app.AskText(text+"Please enter your login in the format username:password", "")
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
		logger.Err(e, "Check mail")
		if app.err == nil || e.Error() != app.err.Error() { // Error buffer, dont warn twice the same information.
			app.render.Error(e)
			app.ShowDialog("Mail check error"+e.Error(), int32(app.conf.DialogTimer))
			// app.PopUp("Mail check error", e.Error())
			app.err = e
		}

	case first:
		app.Log.Debug("  * First check", delta)

	case delta > 0:
		app.Log.Debug("  * Count changed", delta)
		app.sendAlert(delta)

	case delta == 0:
		app.Log.Debug("  * ", "no change")
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
		app.mailPopup(min(delta, app.conf.DialogNbMail), "ListMailsNew")
	}
	if app.conf.AlertAnimName != "" {
		app.Animate(app.conf.AlertAnimName, int32(app.conf.AlertAnimDuration))
	}
	if app.conf.AlertSoundEnabled {
		sound := app.conf.AlertSoundFile
		if len(sound) == 0 {
			logger.Info("No sound file configured")
			return
		}
		if !filepath.IsAbs(sound) && sound[0] != []byte("~")[0] { // Check for relative path.
			sound = app.FileLocation(sound)
		}

		logger.Err(exec.Command("paplay", sound).Start(), "Play sound")
		// if e := exec.Command("paplay", sound).Start(); e != nil {
		//~ exec.Command("aplay", sound).Start()
		// }
	}
}

// Show dialog with information for the given number of mails. Can display an
// additional comment about mails being new if the second param is set to true.
//
func (app *Applet) mailPopup(nb int, template string) {
	// feed := app.data.Data().(*Feed)
	feed := app.data.(*Feed)

	// Prepare data for template formater.
	feed.New = nb
	feed.Plural = feed.New > 1
	max := min(feed.New, len(feed.Mail))
	feed.MailsNew = make([]*Email, max)
	for i := 0; i < max; i++ {
		feed.MailsNew[i] = feed.Mail[i]
	}

	// if app.conf.DialogType == dialogInternal {
	text, e := app.ExecuteTemplate(DialogTemplate, template, feed)
	if !logger.Err(e, "Template ListMailsNew") {
		// Remove a last EOL if any (from a template range).
		if text[len(text)-1] == '\n' {
			text = text[:len(text)-1]
		}

		dialog := map[string]interface{}{
			"message":     text,
			"use-markup":  true,
			"time-length": int32(app.conf.DialogTimer),
		}
		logger.Err(app.PopupDialog(dialog, nil), "popup")
		// app.ShowDialog(text, int32(app.conf.DialogTimer))
	}
	// } else {
	// 	if nb == 1 {
	// 		logger.Err(popUp(feed.Mail[0].AuthorName, feed.Mail[0].Title, app.FileLocation("icon"), app.conf.DialogTimer*1000), "libnotify")
	// 	} else {
	// 		title, eTit := app.ExecuteTemplate(DialogTemplate, "TitleCount", feed)
	// 		logger.Err(eTit, "Template TitleCount")
	// 		text, eTxt := app.ExecuteTemplate(DialogTemplate, "ListMails", feed)
	// 		logger.Err(eTxt, "Template ListMails")
	// 		logger.Err(popUp(title, text, app.FileLocation("icon"), app.conf.DialogTimer*1000), "Libnotify")
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
func NewRenderedSVG(app dock.RenderSimple, typ string) RendererMail {
	size := strings.Split(string(typ), " ")[0]

	source, err := ioutil.ReadFile(app.FileLocation("img", size+".svg"))
	if err != nil {
		return NewRenderedNone()
	}

	rs := &RenderedSVG{
		RenderSimple: app,
		pathDefault:  app.FileLocation("img", "gmail-icon.svg"),
		pathTemp:     app.FileLocation("img", "temp.svg"),
		pathError:    app.FileLocation("img", "gmail-error-"+size+".svg"),
		iconSource:   string(source),
	}
	return rs
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

//
//---------------------------------------------------------------[ LIBNOTIFY ]--

// DISABLED FOR NOW

// #L+[Internal dialog;Desktop notifications] Dialog type
// DialogType=Desktop notifications

// // libnotify call is currently stored as a global so libnotify.go can be
// // removed if needed. Need to see the doc about optional dependencies building
// // for better handling.
// //
// var popUp func(title, msg, icon string, duration int) error // store  if enabled.

// // Open a popup on the configured notification systme. Valid options are internal
// // or libnotify.
// //
// func (app *Applet) PopUp(title, msg string) {
// 	if app.conf.DialogType == dialogInternal {
// 		app.ShowDialog(msg, int32(app.conf.DialogTimer))
// 	} else {
// 		var e error
// 		if popUp == nil {
// 			e = errors.New("Applet was compiled with library support disabled")
// 		} else {
// 			e = popUp(title, msg, app.FileLocation("icon"), app.conf.DialogTimer*1000)
// 			//~ DEBUG("notify", e==nil, e)
// 		}
// 		logger.Err(e, "libnotify")
// 	}
// }

//
//------------------------------------------------------------------[ COMMON ]--

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
