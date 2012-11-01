/*

	Provided as external file so it can be removed, here it a simple libnotify integration.
	Will move to the Go dock API or the Cairo-Dock core, depending on fabounet.
*/
package main

import (
	"errors"
	"github.com/lenormf/go-notify/notify" // libnotify
	"github.com/sqp/godock/libs/log"      // Display info in terminal.
)

func init() {
	popUp = func(title, text, icon string, duration int) (e error) {
		notify.Init("Cairo-Dock")
		hello := notify.NotificationNew(title,
			text,
			icon)

		if hello == nil {
			return errors.New("Unable to create")
		}
		hello.SetTimeout(int32(duration))

		log.Debug("Libnotify send message")
		if e = hello.Show(); e.Error() == "" { // e == nil // Lib seem to send empty errors...
			//~ log.Debug("Libnotify message sent", e == nil, e)
			notify.UnInit()
			return nil
		}

		return e
	}
}

//~ log.Err(, "Notification")

//~ notify.NotificationSetTimeout(hello, DELAY)
// hello.SetTimeout(3000)
//~ if !log.Err(notify.NotificationShow(hello), "Notification") {
//~ time.Sleep(DELAY * 1000000)
// hello.Close()
//~ notify.NotificationClose(hello)
