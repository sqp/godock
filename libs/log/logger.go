/*
Package log is a simple colored info and errors logger. Errors will be displayed only if they are
valid, so you can send all your errors without having to bother if they are
filled or not. You just have to use it with all errors you want to be displayed,
and drop the others.

	log.SetPrefix("test")

	log.Info("my topic in green", "my other", 2, "or", 3, "informations.")
	log.Info("Params", "number", `and type don't matter`)

	term.Warn(errors.New("field not found"), "Parse data")

Output:
	[test] 14:28:54 [my topic in green] my other 2 or 3 informations.
	[test] 14:28:54 [Params] number and type don't matter
	[test] 14:28:54 Warning Parse data : field not found



Very common case of chained tests:

	data, e := GetSomeData()
	if log.GetErr(e, "Get data") { // when we need to keep or forward the error.
		return e
	} else {
		parsed, e := ParseMyData(data)
		log.Err(e, "Parse data") // when we just want to output it.

		if !log.Err(AnalyzeData(parsed), "") { // used as a simple test.
			log.Info("Data analyze", "Everything is OK")
		}
	}
*/
package log

import (
	"log"
	"os"

	"fmt"
	"path"
	"runtime"
	"time"
)

//
//-----------------------------------------------------------[ TEXT WRAPPERS ]--

func Yellow(msg string) string  { return Colored(msg, FgYellow) }  // Yellow formatting of text.
func Magenta(msg string) string { return Colored(msg, FgMagenta) } // Magenta formatting of text.
func Green(msg string) string   { return Colored(msg, FgGreen) }   // Green formatting of text.
func Red(msg string) string     { return Colored(msg, FgRed) }     // Red formatting of text.

func Colored(msg, color string) string { // Colored formatting of text.
	if msg != "" {
		return color + msg + Reset
	}
	return ""
}

func Parenthesis(msg string) string { return addAround("(", msg, ")") } // Parenthesis added around text.
func Bracket(msg string) string     { return addAround("[", msg, "]") } // Brackets added around text.

func addAround(before, msg, after string) string {
	if msg != "" {
		return before + msg + after
	}
	return ""
}

//
//-----------------------------------------------------------------[ LOGGING ]--

var std = log.New(os.Stdout, "", log.Ltime) // log.Ldate
var debug bool

// SetPrefix set the prefix of the logger display.
//
func SetPrefix(pre string) {
	std.SetPrefix(Yellow("[" + pre + "] "))
	log.SetPrefix(Yellow("[" + pre + "] "))
	log.SetFlags(log.Ltime)
}

// SetDebug set the debug flag. If true all messages sent to the Debug function
// will be displayed.
//
func SetDebug(flag bool) {
	debug = flag
}

// Info displays normal informations on the standard output, with the first param in green.
//
func Info(msg string, more ...interface{}) {
	Render(FgGreen, msg, more...)
}

// To be used every time a usefull step is reached in your module activity. It
// will display the flood to the user only when the debug flag is enabled.
//
// Currently only flood the console, but other reporting methods could be
// implemented (file, special parser...).
//
func Debug(msg string, more ...interface{}) {
	if debug {
		Render(FgMagenta, msg, more...)
	}
}

// Render displays the msg argument in the given color. The colored message is
// passed with others to classic println.
//
func Render(color, msg string, more ...interface{}) {
	// println(, list...)
	list := append([]interface{}{time.Now().Format("15:04:05"), Yellow(caller()), Bracket(Colored(msg, color))}, more...)
	fmt.Println(list...)
	// log.Println(list...)
}

func caller() string {
	// var m runtime.MemStats
	// runtime.ReadMemStats(&m)
	// Info("mem", m.Alloc)

	_, file, _, _ := runtime.Caller(3) // (pc uintptr, file string, line int, ok bool)
	// Info("package", path.Base(path.Dir(file)))
	return path.Base(path.Dir(file))
}

//
//---------------------------------------------------------[ DEVELOPPER INFO ]--

// Dev displays normal colored informations on the standard output, To be used
// for temporary developer tests, so they could be easily tracked.
//
func DEV(msg string, more ...interface{}) {
	Render(Bright, msg, more...)
}

// DETAIL prints the detailled content of a variable.
// This is a convenience function for the developper, but not meant for
// production code. Its name is in full caps so it can be better seen and found.
//
func DETAIL(i interface{}) {
	log.Printf("%##v\n", i)
}

//
//---------------------------------------------------------[ ERROR REPORTING ]--

// Warn test and log the error as Warning type. Return true if an error was found.
//
func Warn(e error, msg string) (fail bool) {
	return l(e, Yellow("Warning"), msg)
}

// Err test and log the error as Error type. Return true if an error was found.
//
func Err(e error, msg string) (fail bool) {
	return l(e, Red("Error"), msg)
}

// GetErr test and logs the error, and return it for later use.
func GetErr(e error, msg string) error {
	if e != nil {
		std.Println(Red("Error"), msg, ":", e)
	}
	return e
}

func l(e error, level, msg string) (fail bool) {
	if e != nil {
		fail = true
		std.Println(level, msg, ":", e)
	}
	return fail
}

// Fatal will log the error and exit the program if an error was found.
//
func Fatal(e error, msg string) {
	if Err(e, msg) {
		os.Exit(2)
	}
}
