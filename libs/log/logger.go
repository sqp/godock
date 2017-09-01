/*
Package log is a simple colored info and errors logger.

Errors will be displayed only if they are valid, so you can send all your errors
without having to bother if they are filled or not.
You just have to use it with all errors you want to be displayed and sort the others.

98% code coverage in examples.
*/
package log

import (
	"github.com/google/shlex" // Parse exec command.

	"github.com/sqp/godock/libs/cdtype"         // Logger type.
	"github.com/sqp/godock/libs/dock/confown"   // New dock own settings.
	"github.com/sqp/godock/libs/log/crash"      // Crash parser.
	"github.com/sqp/godock/libs/packages/build" // Pkgbuild counters.

	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Logs provides a default history logger.
var Logs = NewHistory()

// CmdPlaySound defines simple sound players command and args.
var CmdPlaySound = [][]string{
	{"paplay", "--client-name=cairo-dock"},
	{"aplay"},
	{"play"},
}

//
//-------------------------------------------------------------[ MAIN LOGGER ]--

// Log is a simple colored info and errors logger.
//
type Log struct {
	cdtype.LogFmt // extends the formater

	name    string        // line prefix.
	isDebug bool          // active flood or not.
	out     cdtype.LogOut // forwarder optional
}

// NewLog creates a logger with the forwarder.
//
func NewLog(out cdtype.LogOut) *Log {
	logFmt, isFmt := out.(cdtype.LogFmt)
	if isFmt {
		return &Log{out: out, LogFmt: logFmt}
	}
	return &Log{LogFmt: NewFmt()}
}

// SetName set the displayed and forwarded name for the logger.
//
func (l *Log) SetName(name string) cdtype.Logger {
	l.name = name
	return l
}

// SetLogOut connects the optional forwarder to the logger.
//
func (l *Log) SetLogOut(out cdtype.LogOut) cdtype.Logger {
	l.out = out
	return l
}

// LogOut returns the optional forwarder of the logger.
//
func (l *Log) LogOut() cdtype.LogOut {
	return l.out
}

// SetDebug change the debug state of the logger.
// Only enable or disable messages send with the Debug command.
//
func (l *Log) SetDebug(debug bool) cdtype.Logger {
	l.isDebug = debug
	return l
}

// GetDebug gets the debug state of the logger.
//
func (l *Log) GetDebug() bool {
	return l.isDebug
}

// Debug is to be used every time a useful step is reached in your module
// activity. It will display the flood to the user only when the debug flag is
// enabled.
//
func (l *Log) Debug(msg string, more ...interface{}) {
	if l.GetDebug() {
		l.writeMsg(cdtype.LevelDebug, msg, more...)
	}
}

// Debugf log a new debug message with arguments formatting.
//
func (l *Log) Debugf(title, format string, args ...interface{}) {
	if l.GetDebug() {
		l.Debug(title, fmt.Sprintf(format, args...))
	}
}

// Info displays normal informations on the standard output, with the first param in green.
//
func (l *Log) Info(msg string, more ...interface{}) {
	l.writeMsg(cdtype.LevelInfo, msg, more...)
}

// Infof log a new info with arguments formatting.
//
func (l *Log) Infof(msg, format string, args ...interface{}) {
	l.Info(msg, fmt.Sprintf(format, args...))
}

// DEV is like Info, but to be used by the dev for his temporary tests.
//
func (l *Log) DEV(msg string, more ...interface{}) {
	l.writeMsg(cdtype.LevelDEV, msg, more...)
}

// Warn test and log the error as warning type. Return true if an error was found.
//
func (l *Log) Warn(e error, args ...interface{}) (fail bool) {
	return l.writeErr(e, cdtype.LevelWarn, args...)
}

// Warnf log a new error with arguments formatting.
//
func (l *Log) Warnf(msg, format string, args ...interface{}) {
	e := fmt.Errorf(format, args...)
	l.Warn(e, msg)
}

// NewWarn log a new warning.
//
func (l *Log) NewWarn(msg string, args ...interface{}) {
	e := errors.New(fmt.Sprint(args...))
	l.Warn(e, msg)
}

// Err test and log the error as Error type. Return true if an error was found.
//
func (l *Log) Err(e error, args ...interface{}) (fail bool) {
	return l.writeErr(e, cdtype.LevelError, args...)
}

// NewErr log a new error.
//
func (l *Log) NewErr(msg string, args ...interface{}) {
	e := errors.New(fmt.Sprint(args...))
	l.Err(e, msg)
}

// Errorf log a new error with arguments formatting.
//
func (l *Log) Errorf(msg, format string, args ...interface{}) {
	e := fmt.Errorf(format, args...)
	l.Err(e, msg)
}

// Write forward the stream to the connected logger.
//
func (l *Log) Write(p []byte) (n int, err error) {
	l.out.Raw(l.name, string(p))
	// print(string(p))
	return len(p), nil
}

func (l *Log) writeMsg(level cdtype.LogLevel, msg string, more ...interface{}) {
	if l.out == nil {
		io.WriteString(os.Stdout, l.FormatMsg(level, l.name, msg, more...))
		return
	}
	l.out.Msg(level, l.name, msg, more...)
}

func (l *Log) writeErr(e error, level cdtype.LogLevel, args ...interface{}) (fail bool) {
	if e == nil {
		return false
	}

	// Use ln func to ensure spaces are added between args and remove new line.
	msg := strings.TrimSuffix(fmt.Sprintln(args...), "\n")

	if l.out == nil {
		io.WriteString(os.Stdout, l.FormatErr(e, level, l.name, msg))
		return true
	}

	l.out.Err(e, level, l.name, msg)
	return true
}

//
//--------------------------------------------------------[ EXECUTE COMMANDS ]--

// ExecShow run a command with output forwarded to console and wait.
//
func (l *Log) ExecShow(command string, args ...string) error {
	return l.ExecCmd(command, args...).Run()
}

// ExecAsync run a command with output forwarded to console but don't wait for its completion.
// Errors will be logged.
//
func (l *Log) ExecAsync(command string, args ...string) error {
	e := l.ExecCmd(command, args...).Start()
	l.Err(e, "execute failed "+command)
	return e
}

// ExecSync run a command with and grab the output to return it when finished.
//
func (l *Log) ExecSync(command string, args ...string) (string, error) {
	out, e := exec.Command(command, args...).CombinedOutput() // Output()

	// if l.Err(e, "ExecSync: "+strings.Join(args, " ")) {
	// 	args = append([]string{command}, args...)
	// }
	return string(out), e
}

// ExecCmd provides a generic command with output forwarded to console.
//
func (l *Log) ExecCmd(command string, args ...string) *exec.Cmd {
	cmd := exec.Command(command, args...)

	if l.out != nil {
		cmd.Stdout = l // we have a special forwarder, so we will act as a writer to forward the data with a sender name.
		cmd.Stderr = l // TODO: need to split std and err streams.
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd
}

// ExecShlex parse the command with shlex before returning an ExecCmd.
//
func (l *Log) ExecShlex(command string, args ...string) (*exec.Cmd, error) {
	cmds, e := shlex.Split(command)
	if l.Err(e, "parse command args", command) {
		return nil, e
	}
	command = cmds[0]
	args = append(cmds[1:], args...)
	l.Debug("launch", command, args)
	return l.ExecCmd(command, args...), nil
}

// PlaySound plays a sound file.
//
func (l *Log) PlaySound(soundFile string) error {
	if soundFile == "" {
		return errors.New("empty file path")
	}

	for _, args := range CmdPlaySound {
		cmdname, e := exec.LookPath(args[0])
		if e == nil {
			args := append(args[1:], soundFile)
			l.Debug("PlaySound", cmdname, args)
			return l.ExecAsync(cmdname, args...)
		}
	}
	return fmt.Errorf("can't find any command in: %X", CmdPlaySound)
}

//
//-------------------------------------------------------------[ LOG HISTORY ]--

type feeder interface {
	Feed(string)
}

// History provides an history for the Log system.
//
type History struct {
	Fmt                   // extends the formater
	term  feeder          // terminal to forward (optional).
	msgs  []cdtype.LogMsg // list of messages.
	mutex sync.RWMutex    // list mutex.
	delay time.Duration   // history duration.
}

// NewHistory creates a logging history with an optional forwarder.
//
func NewHistory(optionalFeeder ...feeder) *History {
	hist := &History{
		Fmt:   *NewFmt(),
		delay: 12 * time.Hour,
	}
	if len(optionalFeeder) > 0 {
		hist.SetTerminal(optionalFeeder[0])
	}
	return hist
}

// SetTerminal sets the optional terminal forwarder.
//
func (hist *History) SetTerminal(f feeder) {
	hist.term = f
}

// SetDelay sets how long messages are stored in history.
//
func (hist *History) SetDelay(d time.Duration) {
	hist.delay = d
}

// List returns the log messages saved.
//
func (hist *History) List() []cdtype.LogMsg {
	hist.mutex.Lock()
	defer hist.mutex.Unlock()
	return hist.msgs
}

// Raw logs a raw data message.
//
func (hist *History) Raw(sender, msg string) {
	hist.newMsg(cdtype.LogMsg{
		Text:   msg,
		Sender: sender})
}

// Msg logs a standard message.
//
func (hist *History) Msg(level cdtype.LogLevel, sender, msg string, more ...interface{}) {
	hist.newMsg(cdtype.LogMsg{
		Text:   hist.Format(hist.LevelColor(level), sender, msg, more...),
		Sender: sender})
}

// Err logs an error message.
//
func (hist *History) Err(e error, level cdtype.LogLevel, sender, msg string) {
	hist.newMsg(cdtype.LogMsg{
		Text:   hist.FormatErr(e, level, sender, msg),
		Sender: sender})
}

// Write saves the log into his history and send it back to the default output.
// If a terminal is defined, it will have the data forwarded too.
//
func (hist *History) Write(p []byte) (n int, err error) {
	hist.newMsg(cdtype.LogMsg{Text: string(p), Sender: "dock"})
	return len(p), nil
}

// save and display message.
func (hist *History) newMsg(msg cdtype.LogMsg) {

	// Forward to standard output.
	io.WriteString(os.Stdout, msg.Text)

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
func (hist *History) removeOld() int {
	var lastindex int
	var datecomp = time.Now().Add(-hist.delay)
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
//----------------------------------------------------------------[ RECOVERY ]--

// Recover from crash. Use with defer before a dangerous action.
//
func (l *Log) Recover() {
	r := recover()
	if r == nil {
		return
	}
	printCrash(r, false)
}

// GoTry launch a secured go routine. Panic will be recovered and logged.
//
func (l *Log) GoTry(call func()) {
	go func() {
		defer func() {
			r := recover()
			if r != nil {
				printCrash(r, true)
			}
		}()
		call()
	}()
}

func printCrash(r interface{}, showCaller bool) {
	io.WriteString(os.Stdout, crash.Parse(r, true)+"\n")
	if !confown.Current.CrashRecovery {
		// TODO: quit nicely?
		os.Exit(1)
	}

	if build.Current.File != "" {
		build.Current.IncreaseCrash()
	}

	io.WriteString(os.Stdout, "==========Recovered\nThe dock state is now unknown\nYou may have to restart it to restore the expected behavior"+"\n\n")
}
