// +build !gtk
// +build !dock

package clipboard

import "github.com/atotto/clipboard"

func init() {
	Write = clipboard.WriteAll
	Read = clipboard.ReadAll
}
