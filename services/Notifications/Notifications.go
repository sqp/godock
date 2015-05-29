// Package Notifications is a desktop notifications history applet for Cairo-Dock.
//
// requires a hacked version of the dbus api (that wont stop after eavesdropping a message).
//
package Notifications

// https://developer.gnome.org/notification-spec/

import (
	"github.com/godbus/dbus"

	"github.com/sqp/godock/libs/appdbus"
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/dock" // Connection to cairo-dock.

	"strconv"
	"strings"
)

// Applet data.
//
type Applet struct {
	cdtype.AppBase // Applet base and dock connection.

	conf   *appletConf
	notifs *Notifs
}

// NewApplet creates a new Notifications applet instance.
//
func NewApplet() cdtype.AppInstance {
	app := &Applet{AppBase: dock.NewCDApplet()} // Icon controler and interface to cairo-dock.

	app.notifs = &Notifs{}
	app.notifs.SetOnCount(app.UpdateCount)
	app.Log().Err(app.notifs.Start(), "notifications listener")

	app.defineActions()

	return app
}

// Init loads user configuration if needed and initialise applet.
//
func (app *Applet) Init(loadConf bool) {
	app.LoadConfig(loadConf, &app.conf) // Load config will crash if fail. Expected.

	app.notifs.MaxSize = app.conf.NotifSize
	app.notifs.Blacklist = app.conf.NotifBlackList

	// Fill config empty settings.
	if app.conf.NotifAltIcon == "" {
		app.conf.NotifAltIcon = app.FileLocation(defaultNotifAltIcon)
	}
	if app.conf.Icon == "" {
		app.conf.Icon = app.FileLocation("icon")
	}

	// Set defaults to dock icon: display and controls.
	app.SetDefaults(cdtype.Defaults{
		Icon:      app.conf.Icon,
		Label:     app.conf.Name,
		Templates: []string{"notif"},
		Debug:     app.conf.Debug})
}

//------------------------------------------------------------------[ EVENTS ]--

// DefineEvents sets applet events callbacks.
//
func (app *Applet) DefineEvents(events *cdtype.Events) {

	events.OnClick = app.ActionCallback(ActionShowAll)

	events.OnMiddleClick = app.ActionCallback(ActionClear)

	events.OnBuildMenu = app.BuildMenuCallback(menuUser)

	events.OnShortkey = func(key string) {
		// if key == app.conf.ShortkeyOpen {
		// }
		// if key == app.conf.ShortkeyCheck {
		// }
	}

	events.OnDropData = func(data string) {
		app.Log().Info("Grep " + data)
		// stream(data)
	}

}

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
			ID:       ActionShowAll,
			Name:     "Show messages",
			Icon:     "media-seek-forward",
			Call:     app.displayAll,
			Threaded: true,
		},
		&cdtype.Action{
			ID:       ActionClear,
			Name:     "Clear all",
			Icon:     "edit-clear",
			Call:     app.notifs.Clear,
			Threaded: true,
		},
	)
}

// UpdateCount shows the number of messages on the icon, and displays the
// alternate icon if count > 0.
//
func (app *Applet) UpdateCount(count int) {
	text := ""
	icon := ""
	switch {
	case count > 0:
		icon = app.conf.NotifAltIcon
		text = strconv.Itoa(count)

	case app.conf.Icon != "":
		icon = app.conf.Icon
	}
	app.SetQuickInfo(text)
	app.SetIcon(icon)
}

func (app *Applet) displayAll() {
	var msg string
	messages := app.notifs.List()
	if len(messages) == 0 {
		msg = "No recent notifications"
	} else {
		text, e := app.ExecuteTemplate("notif", "ListNotif", messages)
		app.Log().Err(e, "template")
		msg = strings.TrimRight(text, "\n")
	}

	app.PopupDialog(cdtype.DialogData{
		Message:   msg,
		UseMarkup: true,
		Buttons:   "edit-clear;cancel",
		Callback:  cdtype.DialogCallbackIsOK(app.ActionCallback(ActionClear)), // Clear notifs if the user press the 1st button.
	})

	// if self.config['clear'] else 4 + len(msg)/40 }  // if we're going to clear the history, show the dialog until the user closes it
}

//
//-----------------------------------------------------------[ NOTIFICATIONS ]--

// Notif defines a single Dbus notification.
//
type Notif struct {
	Sender, Icon, Title, Content string
	duration, ID                 uint32
}

// Notifs handles Dbus notifications management.
//
type Notifs struct {
	C         chan *dbus.Message
	MaxSize   int
	Blacklist []string

	messages  []*Notif
	callCount func(int)
}

const match = "type='method_call',path='/org/freedesktop/Notifications',member='Notify',eavesdrop='true'"

// List returns the list of notifications.
//
func (notifs *Notifs) List() []*Notif {
	return notifs.messages
}

// Clear resets the list of notifications.
//
func (notifs *Notifs) Clear() {
	notifs.messages = nil
	if notifs.callCount != nil {
		notifs.callCount(len(notifs.messages))
	}
}

// Add a new notifications to the list.
//
func (notifs *Notifs) Add(newtif *Notif) {
	if newtif == nil {
		return
	}

	for _, ignore := range notifs.Blacklist {
		if newtif.Sender == ignore {
			return
		}
	}

	if !notifs.replace(newtif) {
		notifs.messages = append(notifs.messages, newtif)
		if len(notifs.messages) > notifs.MaxSize {
			notifs.messages = notifs.messages[len(notifs.messages)-notifs.MaxSize:]
		}
	}

	if notifs.callCount != nil {
		notifs.callCount(len(notifs.messages))
	}
}

// try to replace an old notification (same id). Return true if replaced.
//
func (notifs *Notifs) replace(newtif *Notif) bool {
	// removed for now, ID was always 0.
	// for i, oldtif := range notifs.messages {
	// if oldtif.ID == newtif.ID {

	// 	// TODO:REMOVE !!!
	// 	println("replaced", oldtif.ID, newtif.ID)

	// 	notifs.messages[i] = newtif
	// 	return true
	// }
	// }
	return false
}

// SetOnCount sets the callback for notifications count change.
//
func (notifs *Notifs) SetOnCount(call func(int)) {
	notifs.callCount = call
}

// Start the message eavesdropping loop and forward notifs changes to the callback.
//
func (notifs *Notifs) Start() error {
	var e error
	notifs.C, e = appdbus.EavesDrop(match)
	if e != nil {
		return e
	}
	go notifs.Listen()
	return nil
}

// Listen to eavesdropped messages to find notifications..
//
func (notifs *Notifs) Listen() {
	for msg := range notifs.C {
		switch msg.Type {
		// case dbus.TypeSignal:

		case dbus.TypeMethodCall:
			// ensure we got a valid message
			if msg.Headers[dbus.FieldMember].Value().(string) == "Notify" && len(msg.Body) >= 8 {
				notifs.Add(messageToNotif(msg))
			}
		}
	}
}

// messageToNotif converts the dbus message to a notification.
//
func messageToNotif(message *dbus.Message) *Notif {
	newtif := &Notif{
		Sender:  message.Body[0].(string),
		ID:      message.Body[1].(uint32), // always 0 ??
		Icon:    message.Body[2].(string),
		Title:   message.Body[3].(string),
		Content: message.Body[4].(string),
		// duration: message.Body[7],
	}

	// Title too short (it's probably something we don't mind, like a notification that the volume has changed)
	if len(newtif.Title) < 2 {
		return nil
	}

	return newtif
}
