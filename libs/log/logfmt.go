package log

import (
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/text/color"
	"github.com/sqp/godock/libs/text/strhelp"

	"fmt"
	"time"
)

// Fmt formats log messages.
//
type Fmt struct {
	// UseColor   bool
	timeFormat string
	colorName  func(string) string
	colorInfo  func(string) string
	colorDebug func(string) string
	colorDEV   func(string) string
	colorWarn  func(string) string
	colorError func(string) string
}

// NewFmt creates a logging formater.
//
func NewFmt() *Fmt {
	return &Fmt{
		timeFormat: "15:04:05",
		colorName:  color.Yellow,
		colorInfo:  color.Green,
		colorDebug: color.Magenta,
		colorDEV:   color.TxtBright,
		colorWarn:  color.Yellow,
		colorError: color.Red,
	}
}

// FormatMsg returns a formatted standard message with endline.
//
func (f *Fmt) FormatMsg(level cdtype.LogLevel, sender string, msg string, more ...interface{}) string {
	return f.Format(f.LevelColor(level), sender, msg, more...)
}

// FormatErr returns a formatted error message with endline.
//
func (f *Fmt) FormatErr(e error, level cdtype.LogLevel, sender string, msg ...interface{}) string {
	var (
		str     = fmt.Sprint(msg...)
		colfunc func(string) string
		info    string
	)
	switch level {
	case cdtype.LevelWarn:
		colfunc = f.colorWarn
		info = "warning"
	case cdtype.LevelError:
		colfunc = f.colorError
		info = "error"
	}
	return f.Format(colfunc, sender, info, str, ":", e)
}

// Format returns a formatted message in the given color with endline.
//
func (f *Fmt) Format(colfunc func(string) string, sender, msg string, more ...interface{}) string {
	args := []interface{}{}
	if f.timeFormat != "" {
		args = append(args, time.Now().Format(f.timeFormat))
	}

	if sender != "" {
		args = append(args, f.colorName(sender))
	}

	args = append(args, []interface{}{
		strhelp.Bracket(colfunc(msg)),
	}...)

	return fmt.Sprintln(append(args, more...)...)
}

// LevelColor returns the field color formater for the level.
//
func (f *Fmt) LevelColor(level cdtype.LogLevel) func(string) string {
	switch level {
	case cdtype.LevelInfo:
		return f.colorInfo
	case cdtype.LevelDebug:
		return f.colorDebug
	case cdtype.LevelDEV:
		return f.colorDEV
	}
	return func(str string) string { return str }
}

//
//--------------------------------------------------------------------[ SET ]--

// SetTimeFormat sets the time format displayed as first log argument.
//
func (f *Fmt) SetTimeFormat(format string) {
	f.timeFormat = format
}

// SetColorName sets the formater used to display the sender as second log argument.
//
func (f *Fmt) SetColorName(callFormat func(string) string) {
	f.colorName = callFormat
}

// SetColorInfo sets the formater used to color the message as third log argument.
//
func (f *Fmt) SetColorInfo(callFormat func(string) string) {
	f.colorInfo = callFormat
}

// SetColorDebug sets the formater used to color the message as third log argument.
//
func (f *Fmt) SetColorDebug(callFormat func(string) string) {
	f.colorDebug = callFormat
}

// SetColorDEV sets the formater used to color the message as third log argument.
//
func (f *Fmt) SetColorDEV(callFormat func(string) string) {
	f.colorDEV = callFormat
}

// SetColorWarn sets the formater used to color the message as third log argument.
//
func (f *Fmt) SetColorWarn(callFormat func(string) string) {
	f.colorWarn = callFormat
}

// SetColorError sets the formater used to color the message as third log argument.
//
func (f *Fmt) SetColorError(callFormat func(string) string) {
	f.colorError = callFormat
}
