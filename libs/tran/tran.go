package tran

import (
	// "code.google.com/p/gettext-go/gettext"
	"github.com/gosexy/gettext"
)

// Slate translates the given string.
//
func Slate(str string) string {
	return gettext.Gettext(str)
}

// Scend transcends your program with its gettext settings.
//
func Scend(pkg, dir, codeset string) {
	// gettext.BindTextdomain(pkg, dir, nil)
	// gettext.Textdomain(pkg)

	gettext.BindTextdomain(pkg, dir)
	gettext.BindTextdomainCodeset(pkg, codeset)
	gettext.Textdomain(pkg)

}
