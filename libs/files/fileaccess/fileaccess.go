// Package fileaccess serialize files access.
package fileaccess

import (
	"github.com/sqp/godock/libs/cdtype" // Logger type.

	"sync"
)

//
//--------------------------------------------------------------[ FILE MUTEX ]--

var access = sync.Mutex{}

// Lock locks and prevents concurrent access to config files.
// Could be improved, but it may be safer to use for now.
//
func Lock(log cdtype.Logger) {
	log.Debug("files.Access", "Lock")
	access.Lock()
}

// Unlock releases the access to config files.
//
func Unlock(log cdtype.Logger) {
	log.Debug("files.Access", "Unlock")
	access.Unlock()
}
