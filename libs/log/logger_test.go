package log_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/dock/confown"
	"github.com/sqp/godock/libs/log"
)

func NewTestLogger(hist cdtype.LogOut) *log.Log {
	logger := log.NewLog(hist)
	logger.SetName("test") // displayed prefix

	logger.SetTimeFormat("")     // no time for tests.
	logger.SetColorName(nocolor) // and no color.
	logger.SetColorInfo(nocolor)
	logger.SetColorDebug(nocolor)
	logger.SetColorDEV(nocolor)
	logger.SetColorWarn(nocolor)
	logger.SetColorError(nocolor)
	return logger
}

func nocolor(str string) string { return str }

// Example of a very common case of chained tests in go:
//
func Example() {
	logger := NewTestLogger(log.Logs) // The logger with its common history.

	testChain := func(isErrGet, isErrParse, isErrUse bool) error {
		data, e := GetSomeData(isErrGet)
		if logger.Err(e, "Get data") { // when we need to keep or forward the error.
			return e
		}

		parsed, e := ParseMyData(data, isErrParse)
		logger.Err(e, "Parse data", "(don't block)") // when we just want to output it.

		result, e := UseData(parsed, isErrParse, isErrUse)
		if logger.Err(e, "Use data") { // used as a simple test.
			return e
		}
		logger.Info("Data used", result)
		return nil
	}

	for _, test := range []struct {
		isErrGet, isErrParse, isErrUse bool
	}{
		{true, false, false},
		{false, true, false},
		{false, false, true},
		{false, false, false},
	} {
		testChain(test.isErrGet, test.isErrParse, test.isErrUse)
	}

	// Output:
	// test [error] Get data : get fail
	// test [error] Parse data (don't block) : parse fail
	// test [Data used] have half data
	// test [error] Use data : use fail
	// test [Data used] everything is fine
}

//
//---------------------------------------------------------------[ MOCK DATA ]--

func GetSomeData(isErrGet bool) (string, error) {
	if isErrGet {
		return "", errors.New("get fail")
	}
	return "get success", nil
}

func ParseMyData(data string, isErrParse bool) (string, error) {
	if isErrParse {
		return "", errors.New("parse fail")
	}
	return "parse success", nil
}

func UseData(parsed string, isErrParse, isErrUse bool) (string, error) {
	if isErrUse {
		return "", errors.New("use fail")
	}
	if isErrParse {
		return "have half data", nil
	}
	return "everything is fine", nil
}

//
//-----------------------------------------------------------[ MORE EXAMPLES ]--

// ExampleMore shows more usage of the logger.
func ExampleMore() {
	logger := NewTestLogger(log.Logs)
	logger.SetLogOut(log.Logs)

	prevSize := len(logger.LogOut().List())
	printDebug(logger)

	logger.SetName("")
	logger.DEV("printed", len(logger.LogOut().List())-prevSize)
	// Output:
	// test [Info] with 3 args
	// test [Infof] with 1
	// test [DEV] for temp reports
	// test [warning] Warn : err_msg
	// test [warning] Warnf : msg
	// test [warning] NewWarn : msg
	// test [error] Err : err_msg
	// test [error] Errorf : msg
	// test [error] NewErr : msg
	// test [Debug] visible because true
	// test [Debugf] (*log.Log).Debugf use fmt to format
	// [printed] 11
}

// ExampleNoHist tests the logger without history.
func ExampleNoHist() {
	logger := NewTestLogger(nil)
	printDebug(logger)
	// Output:
	// test [Info] with 3 args
	// test [Infof] with 1
	// test [DEV] for temp reports
	// test [warning] Warn : err_msg
	// test [warning] Warnf : msg
	// test [warning] NewWarn : msg
	// test [error] Err : err_msg
	// test [error] Errorf : msg
	// test [error] NewErr : msg
	// test [Debug] visible because true
	// test [Debugf] (*log.Log).Debugf use fmt to format
}

func printDebug(logger *log.Log) {
	logger.Info("Info", "with", 3, "args")
	logger.Infof("Infof", "with %d", 1)
	logger.DEV("DEV", "for temp reports")

	e := errors.New("err_msg")
	logger.Warn(e, "Warn")
	logger.Warnf("Warnf", "msg")
	logger.NewWarn("NewWarn", "msg")
	logger.Err(e, "Err")
	logger.Errorf("Errorf", "msg")
	logger.NewErr("NewErr", "msg")

	logger.SetDebug(true)
	logger.Debug("Debug", "visible because", logger.GetDebug())
	logger.Debugf("Debugf", "(%T).Debugf use %s to format", logger, "fmt")
	logger.SetDebug(false)
	logger.Debug("Debug", "hidden")
	logger.Debugf("Debugf", "also")
}

// ExampleForwardTerm tests the logger with a custom terminal.
func ExampleForwardTerm() {
	var term mockFeeder
	logger := NewTestLogger(log.NewHistory(&term))

	logger.Info("double")
	hist := logger.LogOut().(*log.History)
	hist.SetTerminal(&term)

	hist.SetDelay(time.Millisecond)     // Set history to one millisec.
	<-time.After(50 * time.Millisecond) // Now our next message should clear the history.

	logger.Info("spam")
	hist.Write([]byte("hist.Write\n")) // to all buffers (stdout, hist and term)

	fmt.Println("in history:")
	fmt.Println("history size", len(hist.List()))
	for _, msg := range hist.List() { // print history.
		fmt.Print(msg.Text)
	}

	fmt.Println("in buffer:")
	fmt.Print(strings.Replace(string(term), "\r\n", "\n", -1)) // But the other buffer has everything.

	// Output:
	// test [double]
	// test [spam]
	// hist.Write
	// in history:
	// history size 2
	// test [spam]
	// hist.Write
	// in buffer:
	// test [double]
	// test [spam]
	// hist.Write
}

type mockFeeder string

func (m *mockFeeder) Feed(str string) { *m += mockFeeder(str) }

// TestRecover shows and tests the logger recovery.
func TestRecover(t *testing.T) {
	confown.Current.CrashRecovery = true
	logger := NewTestLogger(nil)

	logger.GoTry(func() { panic("safe") })
	<-time.After(time.Second) // wait async

	crash := func() { panic("crash or not") } // hidden panic so fail doesn't appear unreachable.
	defer logger.Recover()
	crash()
	t.Fail() // not called
}

// ExampleExec shows all exec modes available from the logger.
func ExampleExec() {
	logger := NewTestLogger(nil)
	e := logger.ExecCmd("echo", "ExecCmd", "done").Run()
	if logger.Err(e, "ExecCmd") {
		return
	}

	logger = NewTestLogger(log.Logs)
	e = logger.ExecCmd("echo", "ExecCmd", "term").Run()
	if logger.Err(e, "ExecCmd") {
		return
	}

	e = logger.ExecShow("echo", "ExecShow", "done")
	if logger.Err(e, "ExecShow") {
		return
	}

	result, e := logger.ExecSync("echo", "ExecSync", "done")
	if logger.Err(e, "ExecSync") {
		return
	}
	fmt.Print(result)

	cmd, e := logger.ExecShlex(`echo "ExecShlex done"`)
	if logger.Err(e, "ExecShlex") {
		return
	}
	e = cmd.Run()
	if logger.Err(e, "ExecShlex") {
		return
	}

	e = logger.PlaySound("/usr/share/cairo-dock/plug-ins/Sound-Effects/on-click.wav")
	if logger.Err(e, "PlaySound") {
		return
	}

	// error paths
	logger.ExecShlex("bad ExecShlex \" ")
	e = logger.PlaySound("")
	logger.Err(e, "PlaySound no file")
	log.CmdPlaySound = nil
	e = logger.PlaySound("/usr/share/cairo-dock/plug-ins/Sound-Effects/on-click.wav")
	logger.Err(e, "PlaySound no commands")
	logger.SetTimeFormat(" ")
	logger.FormatMsg(cdtype.LevelUnknown, "sender", "msg")

	defer logger.Recover() // recover unused path

	// async last
	e = logger.ExecAsync("echo", "ExecAsync", "done")
	if logger.Err(e, "ExecAsync") {
		return
	}

	<-time.After(time.Second) // wait async

	// Output:
	// ExecCmd done
	// ExecCmd term
	// ExecShow done
	// ExecSync done
	// ExecShlex done
	// test [error] parse command args bad ExecShlex "  : EOF found when expecting closing quote
	// test [error] PlaySound no file : empty file path
	// test [error] PlaySound no commands : can't find any command in: []
	// ExecAsync done
}
