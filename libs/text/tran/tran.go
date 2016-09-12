// Package tran translates user displayed text.
package tran

import (
	"github.com/gosexy/gettext"

	"github.com/sqp/godock/libs/cdglobal" // Global consts.
)

func init() {
	Scend(cdglobal.GettextPackageCairoDock, cdglobal.CairoDockLocaleDir, "UTF-8")
}

// Slate translates the given string in the dock domain.
//
func Slate(str string) string {
	return gettext.Gettext(str)
}

// Splug translates the given string in the applet domain.
//
func Splug(str string) string {
	return gettext.DGettext(cdglobal.GettextPackagePlugins, str)
}

// Sloc translates the given string using another domain name.
//
func Sloc(domain, str string) string {
	return gettext.DGettext(domain, str)
}

// Scend transcends your program with its gettext settings.
//
func Scend(pkg, dir, codeset string) {
	gettext.SetLocale(gettext.LcAll, "")
	gettext.BindTextdomain(pkg, dir)
	gettext.Textdomain(pkg)
	gettext.BindTextdomainCodeset(pkg, codeset)
}
