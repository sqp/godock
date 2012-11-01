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

//------------------------------------------------------------------------------
// Poller
//------------------------------------------------------------------------------

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

func New(callCheck func()) *Poller {
	poller := &Poller{
		callCheck: callCheck,
		ticker:    new(time.Ticker),
		restart:   make(chan bool),
	}
	return poller
}

//----------------------------------------------------------------[ OLD ]--

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

// Set polling interval time, in minutes.
//
func (poller *Poller) SetInterval(delay int) {
	poller.delay = time.Duration(delay)
}

// Get the restart channel. You will need to lock it in a select loop to have a real
// polling routine.
//
func (poller *Poller) GetRestart() chan bool {
	return poller.restart
}

//------------------------------------------------------------------[ ACTION ]--

// Start a new ticker and directly launch first check routine. 
//
func (poller *Poller) Start() *time.Ticker {
	go poller.checkRoutine() // Always check.

	poller.active = true
	if poller.delay > 0 {
		poller.ticker = time.NewTicker(poller.delay * time.Minute)
	}
	return poller.ticker
}

// Restart polling ticker. This will send an event on the restart channel.
//
func (poller *Poller) Restart() {
	poller.ticker.Stop()
	poller.restart <- true // send our restart event.
}

// Check action. Launch PreCheck, OnCheck and PostCheck callbacks.
//
func (poller *Poller) checkRoutine() {
	if poller.started != nil { // Pre check call.
		poller.started()
	}

	poller.callCheck() // Data check call. Does the real polling job.

	if poller.finished != nil { // Post check call.
		poller.finished()
	}
}
