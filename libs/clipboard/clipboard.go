// Package clipboard provides interaction with the desktop clipboard.
//
// Dependencies
//
// tag gtk or dock :
//   go binding: github.com/conformal/gotk3.
//
// default :
//   external dependency: xclip or xsel.
//
package clipboard

var (
	// Write sets the clipboard content.
	Write func(string) error

	// Read returns the clipboard content.
	Read func() (string, error)
)
