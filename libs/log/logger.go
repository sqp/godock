/*
Simple colored info and errors logger. Errors will be displayed only if they are
valid, so you can send all your errors without having to bother if they are
filled or not. You just have to use it with all errors you want to be displayed,
and drop the others.

	log.SetPrefix("test")

	log.Info("my topic in green", "my other", 2, "or", 3, "informations.")
	log.Info("Params", "number", `and type don't matter`)

	term.Warn(errors.New("field not found"), "Parse data")

Output:
	[test] 2012/10/28 14:28:54 [my topic in green my other 2 or 3 informations.]
	[test] 2012/10/28 14:28:54 [Params number and type don't matter]
	[test] 2012/10/28 14:28:54 Warning Parse data : field not found



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
)

// Message helpers 
func Colored(msg, color string) string { return color + msg + Reset }

func Yellow(msg string) string  { return FgYellow + msg + Reset }
func Magenta(msg string) string { return FgMagenta + msg + Reset }
func Green(msg string) string   { return FgGreen + msg + Reset }
func Red(msg string) string     { return FgRed + msg + Reset }

func Bracket(str string) string {
	if str != "" {
		return "(" + str + ")"
	}
	return ""
}

// Logger

func SetPrefix(pre string) {
	std.SetPrefix(Yellow("[" + pre + "] "))
	log.SetPrefix(Yellow("[" + pre + "] "))
}

// Displays normal informations on the standard output, with the first param in green.
//
func Info(msg string, more ...interface{}) {
	render(FgGreen, msg, more...)
}

// Displays normal colored informations on the standard output, To be used for
// temporary developer comments, so they could be easily tracked.
//
func DEV(msg string, more ...interface{}) {
	render(FgGreen, msg, more...)
}

// To be used every time a usefull step is reached in your module activity. It 
// will display the flood to the user only when the debug flag is enabled.
// 
// Currently only flood the console, but other reporting methods could be
// implemented (file, special parser...).
//
func Debug(msg string, more ...interface{}) {
	render(FgMagenta, msg, more...)
}

func render(color, msg string, more ...interface{}) {
	var first string = color + msg + Reset
	var list []interface{}
	list = append(list, first)
	list = append(list, more...)
	std.Println(list)
}

var std = log.New(os.Stdout, "", log.Ldate|log.Ltime)

// Errors testing and reporting.
//
func Warn(e error, msg string) (fail bool) {
	return l(e, Yellow("Warning"), msg)
}

func Err(e error, msg string) (fail bool) {
	return l(e, Red("Error"), msg)
}

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

// Test error: log and quit.
//
func Fatal(e error, msg string) {
	if Err(e, msg) {
		os.Exit(2)
	}
}

//~ func Show(msg string) {
//~ log.Println(FgGreen + msg + Reset)
//~ }

//~ func Warning(msg string, more... interface{}) {
//~ render(FgYellow, msg, more...)
//~ }
//~ 
//~ func Error(msg string, more... interface{}) {
//~ render(FgRed, msg, more...)
//~ }
