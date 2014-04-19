/*
Package log is a simple colored info and errors logger.

Errors will be displayed only if they are valid, so you can send all your errors
without having to bother if they are filled or not.
You just have to use it with all errors you want to be displayed and drop the others.

    outdated !!!
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
	"github.com/sqp/godock/libs/cdtype"

	"log"
	"os"
	"os/exec"

	"fmt"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"
)

//
//-----------------------------------------------------------[ TEXT WRAPPERS ]--

// Yellow formatting of text.
func Yellow(msg string) string { return Colored(msg, FgYellow) }

// Magenta formatting of text.
func Magenta(msg string) string { return Colored(msg, FgMagenta) }

// Green formatting of text.
func Green(msg string) string { return Colored(msg, FgGreen) }

// Red formatting of text.
func Red(msg string) string { return Colored(msg, FgRed) }

func Colored(msg, color string) string { // Colored formatting of text.
	if msg == "" {
		return ""
	}
	return color + msg + Reset
}

// Parenthesis added around text if any.
func Parenthesis(msg string) string { return addAround("(", msg, ")") }

// Brackets added around text if any.
func Bracket(msg string) string { return addAround("[", msg, "]") }

func addAround(before, msg, after string) string {
	if msg == "" {
		return ""
	}
	return before + msg + after
}

// Format returns a formatted message in the given color with endline.
//
func Format(color, sender, msg string, more ...interface{}) string {
	list := append([]interface{}{time.Now().Format("15:04:05"), Yellow(sender), Bracket(Colored(msg, color))}, more...)
	return fmt.Sprintln(list...)
}

// FormatErr returns a formatted error message with endline.
//
func FormatErr(e string, sender string, msg ...interface{}) string {
	str := fmt.Sprintln(msg...) // adds an undesired \n at the end, removed next line.
	return Format(FgRed, sender, "error", str[:len(str)-1], ":", e)
}

//
//-------------------------------------------------------------[ MAIN LOGGER ]--

// Log is a simple colored info and errors logger.
//
type Log struct {
	name          string
	debug         bool
	cdtype.LogOut // forwarder
}

// NewLog creates a logger with the forwarder.
//
func NewLog(out cdtype.LogOut) *Log {
	return &Log{LogOut: out}
}

// SetName set the displayed and forwarded name for the logger.
//
func (l *Log) SetName(name string) {
	l.name = name
}

// SetLogOut connects the optional forwarder to the logger.
//
func (l *Log) SetLogOut(out cdtype.LogOut) {
	l.LogOut = out
}

// SetDebug change the debug state of the applet.
// Only enable or disable messages send with the Debug command.
//
func (l *Log) SetDebug(debug bool) {
	l.debug = debug
}

// Debug is to be used every time a usefull step is reached in your module
// activity. It will display the flood to the user only when the debug flag is
// enabled.
//
func (l *Log) Debug(msg string, more ...interface{}) {
	if l.debug {
		if l.LogOut != nil {
			l.LogOut.Debug(l.name, msg, more...)
		} else {
			println("manual debug", msg)
			l.Render(FgMagenta, msg, more...)
		}

		// l.Render(FgMagenta, msg, more...)
	}
}

// Info displays normal informations on the standard output, with the first param in green.
//
func (l *Log) Info(msg string, more ...interface{}) {
	if l.LogOut != nil {
		l.LogOut.Info(l.name, msg, more...)
	} else {
		l.Render(FgGreen, msg, more...)
	}
}

// Render displays the msg argument in the given color.
// The colored message is passed with others to classic println.
//
func (l *Log) Render(color, msg string, more ...interface{}) {
	print(Format(color, l.name, msg, more...))
}

// Warn test and log the error as warning type. Return true if an error was found.
//
func (l *Log) Warn(e error, msg ...string) (fail bool) {
	if e != nil {
		l.NewWarn(e.Error(), msg...)
	}
	return e != nil
}

// NewWarn log a new warning.
//
func (l *Log) NewWarn(e string, msg ...string) {
	l.Render(FgYellow, "warning", msg, ":", e)
}

// Err test and log the error as Error type. Return true if an error was found.
//
func (l *Log) Err(e error, msg ...interface{}) (fail bool) {
	return l.GetErr(e, msg...) != nil
}

// NewErr log a new error.
//
func (l *Log) NewErr(e string, msg ...interface{}) {
	if l.LogOut != nil {
		l.LogOut.Err(e, l.name, msg...)
	} else {
		print(FormatErr(e, l.name, msg...))
	}
}

// GetErr test and logs the error, and return it for later use.
//
func (l *Log) GetErr(e error, msg ...interface{}) error {
	if e != nil {
		l.NewErr(e.Error(), msg...)
	}
	return e
}

// Write forward the stream to the connected logger.
//
func (l *Log) Write(p []byte) (n int, err error) {
	l.LogOut.Raw(l.name, string(p))
	// print(string(p))
	return len(p), nil
}

//
//--------------------------------------------------------[ EXECUTE COMMANDS ]--

// ExecShow run a command with output forwarded to console and wait.
//
func (l *Log) ExecShow(command string, args ...string) error {
	return l.execCmd(command, args...).Run()
}

// ExecAsync run a command with output forwarded to console but don't wait for its completion.
// Errors will be logged.
//
func (l *Log) ExecAsync(command string, args ...string) error {
	return l.GetErr(l.execCmd(command, args...).Start(), "execute failed "+command)
}

// ExecSync run a command with and grab the output to return it when finished.
//
func (l *Log) ExecSync(command string, args ...string) (string, error) {
	out, e := exec.Command(command, args...).Output()
	if l.Err(e, "ExecSync: "+strings.Join(args, " ")) {
		args = append([]string{command}, args...)
	}
	return string(out), e
}

func (l *Log) execCmd(command string, args ...string) *exec.Cmd {
	cmd := exec.Command(command, args...)

	if l.LogOut != nil {
		cmd.Stdout = l // we have a special forwarder, so we will act as a writer to forward the data with a sender name.
		cmd.Stderr = l // TODO: need to split std and err streams.
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd
}

//
//-------------------------------------------------------------[ LOG HISTORY ]--

type feeder interface {
	Feed(string)
}

// LogMsg defines a single log message.
//
type LogMsg struct {
	Text   string
	Sender string
	Time   time.Time
}

// LogHistory provides an history for the Log system.
//
type LogHistory struct {
	term  feeder
	msgs  []LogMsg
	mutex sync.RWMutex
}

// NewLogHistory creates a logging history with an optional forwarder.
//
func NewLogHistory(optionalFeeder ...feeder) *LogHistory {
	hist := &LogHistory{}
	if len(optionalFeeder) > 0 {
		hist.SetTerminal(optionalFeeder[0])
	}
	return hist
}

// SetTerminal sets the optional terminal forwarder.
//
func (hist *LogHistory) SetTerminal(f feeder) {
	hist.term = f
}

// List returns the log messages saved.
//
func (hist *LogHistory) List() []LogMsg {
	hist.mutex.Lock()
	defer hist.mutex.Unlock()
	return hist.msgs
}

// Raw logs a raw data message.
//
func (hist *LogHistory) Raw(sender, msg string) {
	hist.newMsg(LogMsg{
		Text:   msg,
		Sender: sender})
}

// Debug logs a message of type debug.
//
func (hist *LogHistory) Debug(sender, msg string, more ...interface{}) {
	hist.newMsg(LogMsg{
		Text:   Format(FgMagenta, sender, msg, more...),
		Sender: sender})
}

// Info logs a message of type info.
//
func (hist *LogHistory) Info(sender, msg string, more ...interface{}) {
	hist.newMsg(LogMsg{
		Text:   Format(FgGreen, sender, msg, more...),
		Sender: sender})
}

// Err logs a message of type error.
//
func (hist *LogHistory) Err(e string, sender string, msg ...interface{}) {
	hist.newMsg(LogMsg{
		Text:   FormatErr(e, sender, msg...),
		Sender: sender})
}

// Write saves the log into his history and send it back to the default output.
// If a terminal is defined, it will have the data forwarded too.
//
func (hist *LogHistory) Write(p []byte) (n int, err error) {
	hist.newMsg(LogMsg{Text: string(p), Sender: "dock"})
	return len(p), nil
}

// save and display message.
func (hist *LogHistory) newMsg(msg LogMsg) {

	// Forward to standard output.
	// TODO: should reenable!
	print(msg.Text)

	// Display in the window if any.
	if hist.term != nil {
		hist.term.Feed(strings.Replace(msg.Text, "\n", "\r\n", -1))
	}

	// Save to history.
	msg.Time = time.Now()

	hist.mutex.Lock()
	defer hist.mutex.Unlock()

	hist.msgs = append(hist.msgs, msg)
	hist.msgs = hist.msgs[hist.removeOld():]
}

// return index of first valid message.
//
func (hist *LogHistory) removeOld() int {
	var lastindex int
	var datecomp = time.Now().Add(-12 * time.Hour)
	for i, msg := range hist.msgs {
		if msg.Time.Before(datecomp) {
			lastindex = i + 1
			// println("to drop", msg.Text)
		} else {
			// if msg.Time > datecomp {
			return lastindex
		}
	}
	return 0
}

//
//--------------------------------------------------------------[  OLD LOGGING ]--

var std = log.New(os.Stdout, "", log.Ltime) // log.Ldate
var debug bool

var Logs = &LogHistory{}

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

// Debug is to be used every time a usefull step is reached in your module
// activity. It will display the flood to the user only when the debug flag is
// enabled.
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

// DEV displays normal colored informations on the standard output, To be used
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
