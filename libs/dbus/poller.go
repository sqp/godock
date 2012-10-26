/* This is a part of the external applet for Cairo-Dock

Copyright : (C) 2012 by SQP
E-mail : sqp@glx-dock.org

This program is free software; you can redistribute it and/or
modify it under the terms of the GNU General Public License
as published by the Free Software Foundation; either version 3
of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.
http://www.gnu.org/licenses/licenses.html#GPL */

package dbus

import (
"time"
"log"
)
//------------------------------------------------------------------------------
// Poller
//------------------------------------------------------------------------------

// Create a dedicated task with a ticker that handles regular polling actions.
//
type Poller struct {
	// Callbacks.
	callCheck func() // Action to execute on tick.
	started   func() // Action to execute before data polling.
	finished  func() // Action to execute before data polling.

	// Ticker settings.
	delay   time.Duration // Ticker time.
	active  bool
	ticker  *time.Ticker
	restart chan bool // restart channel to forward user requests.
}

func NewPoller(callCheck func()) *Poller {
	poller := &Poller{
		callCheck: callCheck,
		ticker:    new(time.Ticker),
		restart:   make(chan bool),
	}
	return poller
}

// Set callback actions for started and finished mail polling events. Allow you
// to update display during mail connection.
//
func (poller *Poller) SetCallDisplay(onStarted, onFinished func()) {
	poller.started = onStarted
	poller.finished = onFinished
}

// Set polling interval time, in minutes.
//
func (poller *Poller) SetInterval(delay int) {
	poller.delay = time.Duration(delay)
}

// Start a new ticker and directly launch first call. Also returns the restart channel.
func (poller *Poller) NewTicker() (*time.Ticker, chan bool) {
	if poller.Active() {
		poller.ticker.Stop()
	}

	poller.Check()
	poller.active = true
	poller.ticker = time.NewTicker(poller.delay * time.Minute)
	return poller.ticker, poller.restart
}

// Check if polling is active.
//
func (poller *Poller) Active() bool {
	return poller.active
}

// Action to launch on tick.
//
func (poller *Poller) Check() {
	go poller.check()
}

func (poller *Poller) check() {
	if poller.started != nil {
	poller.started()
}
	poller.callCheck()
	if poller.finished != nil {
	poller.finished()
}
}

// Restart polling ticker.
//
func (poller *Poller) Restart() {
	log.Println("should restart")
	poller.restart <- true // send our restart event.
}
