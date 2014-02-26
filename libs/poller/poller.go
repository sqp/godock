/*
Poller is a dedicated task with a ticker that handles regular polling actions.
It does not start a loop, just handles the ticker and restart channels. You
will have to get those with Start and GetRestart and use them in a loop. This
job is generaly done by the dock.StartApplet action, so you better use it or
hack/extend it to your needs.

Display and user information related to the result of the check must be made
using some return callback at the end of the check task.

Display and user information related to the check action itself, like
displaying an activity emblem during the check, should be done using the
PreCheck and PostCheck callbacks.

The goal is to keep each part separated and dedicated to one task. If we
split each role and keep it agnostic of others, we can have easier debuging
and evolution of our applets:
  * The poller send timing events.
  * The check task pull its data and send the results on a OnResult callback.
  * The OnResult callback sorts the data and dispatch it to the renderers or alert interfaces.
  * Renderer interfaces displays to the user informations or alerts the way he prefers.
*/
package poller

import (
	"time"
)

//------------------------------------------------------------------[ POLLER ]--

type Poller struct {
	// Callbacks in this order.
	started   func() // Action to execute before data polling.
	callCheck func() // Action data polling.
	finished  func() // Action to execute after data polling.

	// Ticker settings.
	delay   int  // Interval between checks in second.
	enabled bool // true if the poller should be active.
	active  bool // true if the poller is really active.

	// ticker  *time.Ticker
	restart chan bool // restart channel to forward user requests.
}

func New(callCheck func()) *Poller {
	poller := &Poller{
		callCheck: callCheck,
		enabled:   true,
		// ticker:    new(time.Ticker),
		restart: make(chan bool),
	}
	return poller
}

//---------------------------------------------------------------------[ OLD ]--

// Check if polling is active.
//
//~ func (poller *Poller) Active() bool {
//~ return poller.active
//~ }

//----------------------------------------------------------------[ SETTINGS ]--

// Set callback actions to launch before the polling job.
//
func (poller *Poller) SetPreCheck(onStarted func()) {
	poller.started = onStarted
}

// Set callback actions to launch after the polling job.
//
func (poller *Poller) SetPostCheck(onFinished func()) {
	poller.finished = onFinished
}

// Set polling interval time, in seconds. You can add a default value as second
// argument to be sure you will have a positive value (> 0).
//
func (poller *Poller) SetInterval(delay ...int) int {
	for _, d := range delay {
		if d > 0 {
			poller.delay = d
			return d
		}
	}
	poller.delay = 3600 * 24 // Failed to provide a valid value. Set check interval to a day.
	return poller.delay
}

// Get the restart channel. You will need to lock it in a select loop to have a real
// polling routine.
//
func (poller *Poller) ChanRestart() chan bool {
	return poller.restart
}

//------------------------------------------------------------------[ ACTION ]--

// Start a new ticker and directly launch first check routine.
//
// func (poller *Poller) Start() *time.Ticker {
// 	poller.checkRoutine() // Always check.

// 	log.DEV("end checkroutine")

// 	poller.active = true
// 	if poller.delay > 0 {
// 		// poller.ticker = time.NewTicker(poller.delay * time.Minute)
// 		poller.ticker = time.NewTicker(time.Duration(poller.delay) * time.Second)

// 	}
// 	return poller.ticker
// }

func (poller *Poller) ChanEndTimer() <-chan time.Time {
	if poller.enabled && poller.delay > 0 {
		poller.active = true
		return time.After(time.Duration(poller.delay) * time.Second)
	}
	return nil
}

// Restart polling ticker. This will send an event on the restart channel.
//
func (poller *Poller) Restart() {
	if poller.enabled {
		// poller.Stop()
		poller.restart <- true // send our restart event.
		// poller.enabled = true
		// poller.active = true
	}
}

// Stop the polling ticker.
//
func (poller *Poller) Stop() {
	if poller.active {
		poller.enabled = false
		poller.active = false
	}
}

// Check action. Launch PreCheck, OnCheck and PostCheck callbacks.
//
func (poller *Poller) Action() {
	if poller.started != nil { // Pre check call.
		poller.started()
	}

	poller.callCheck() // Data check call. Does the real polling job.

	if poller.finished != nil { // Post check call.
		poller.finished()
	}
}
