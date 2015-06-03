package tran

import (
	"github.com/gosexy/gettext"
)

// Slate translates the given string.
//
func Slate(str string) string {
	return gettext.Gettext(str)
}

// Sloc translates the given string with another package name.
//
func Sloc(domain, str string) string {
	return gettext.DGettext(domain, str)
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
