package tran

import (
	// "code.google.com/p/gettext-go/gettext"
	"github.com/gosexy/gettext"
)

func Slate(str string) string {
	return gettext.Gettext(str)
}

func Scend(pkg, dir, codeset string) {
	// gettext.BindTextdomain(pkg, dir, nil)
	// gettext.Textdomain(pkg)

	gettext.BindTextdomain(pkg, dir)
	gettext.BindTextdomainCodeset(pkg, codeset)
	gettext.Textdomain(pkg)

}
