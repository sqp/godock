// Package gtktext provides helpers to work with strings.
//
// See:
//   https://developer.gnome.org/pango/stable/PangoMarkupFormat.html
//   http://www.pango.org/HelpOnLists
//
package gtktext

import "strings"

// Escape escapes a string to use as gtk text: &<>.
//
func Escape(msg string) string {
	msg = strings.Replace(msg, "&", "&amp;", -1) // Escape ampersand.
	msg = strings.Replace(msg, "<", "&lt;", -1)  // Escape <.
	msg = strings.Replace(msg, ">", "&gt;", -1)  // Escape >.
	return msg
}

// Big formats the text with the big size.
//
func Big(text string) string { return "<big>" + text + "</big>" }

// Small formats the text with the small size.
//
func Small(text string) string { return "<small>" + text + "</small>" }

// Bold formats the text with the bold font.
//
func Bold(text string) string { return "<b>" + text + "</b>" }

// Mono formats the text with the monospace font.
//
func Mono(text string) string { return "<tt>" + text + "</tt>" }

// URI formats a link with its text.
//
func URI(uri, text string) string { return "<a href=\"" + uri + "\">" + text + "</a>" }

// List converts a text to a bullet list (prepend " *" to lines )
func List(text string) string { sep := " * "; return sep + strings.Replace(text, "\n", "\n"+sep, -1) }
