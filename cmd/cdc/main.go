// Use of this source code is governed by a GPL v3 license. See LICENSE file.

package main

import (
	"github.com/sqp/godock/libs/log"
	"github.com/sqp/godock/libs/srvdbus"

	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"text/template"
	"unicode"
	"unicode/utf8"
)

// A Command is an implementation of a cdc command
//
type Command struct {
	// Run runs the command.
	// The args are the arguments after the command name.
	Run func(cmd *Command, args []string)

	// UsageLine is the one-line usage message.
	// The first word in the line is taken to be the command name.
	UsageLine string

	// Short is the short description shown in the 'cdc help' output.
	Short string

	// Long is the long message shown in the 'cdc help <this-command>' output.
	Long string

	// Flag is a set of flags specific to this command.
	Flag flag.FlagSet

	// CustomFlags indicates that the command will do its own
	// flag parsing.
	CustomFlags bool
}

// Name returns the command's name: the first word in the usage line.
func (c *Command) Name() string {
	name := c.UsageLine
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

// Usage print the command usage.
func (c *Command) Usage() {
	fmt.Fprintf(os.Stderr, "usage: %s\n\n", c.UsageLine)
	fmt.Fprintf(os.Stderr, "%s\n", strings.TrimSpace(c.Long))
	os.Exit(2)
}

// Runnable reports whether the command can be run; otherwise
// it is a documentation pseudo-command such as importpath.
func (c *Command) Runnable() bool {
	return c.Run != nil
}

// Commands lists the available commands and help topics.
// The order here is the order in which they are printed by 'cdc help'.
var commands = []*Command{
	cmdBuild,
	// cmdConfig,
	cmdDebug,
	cmdInstall,
	cmdList,
	// cmdRun,
	// cmdTest,
	cmdRestart,
	cmdService,
	cmdUpload,
	cmdVersion,

	// helpGopath,
	// helpPackages,
	// helpRemote,
	// helpTestflag,
	// helpTestfunc,
}

var logger = &log.Log{}

var exitStatus = 0
var exitMu sync.Mutex

func setExitStatus(n int) {
	exitMu.Lock()
	if exitStatus < n {
		exitStatus = n
	}
	exitMu.Unlock()
}

func main() {
	flag.Usage = usage
	flag.Parse()
	// log.SetFlags(0)

	args := flag.Args()
	if len(args) < 1 {
		usage()
	}

	logger.SetName(os.Args[0])

	if args[0] == "help" {
		help(args[1:])
		return
	}

	for _, cmd := range commands {
		if cmd.Name() == args[0] && cmd.Run != nil {
			cmd.Flag.Usage = func() { cmd.Usage() }
			if cmd.CustomFlags {
				args = args[1:]
			} else {
				cmd.Flag.Parse(args[1:])
				args = cmd.Flag.Args()
			}
			cmd.Run(cmd, args)
			exit()
			return
		}
	}

	fmt.Fprintf(os.Stderr, "cdc: unknown subcommand %q\nRun 'cdc help' for usage.\n", args[0])
	setExitStatus(2)
	exit()
}

var usageTemplate = `cdc, Cairo-Dock Control, is a tool to manage a Cairo-Dock installation.
It can also embed and manage multiple applets if compiled with their support.

Usage:

	cdc command [arguments]

The commands are:
{{range .}}{{if .Runnable}}
    {{.Name | printf "%-11s"}} {{.Short}}{{end}}{{end}}

Use "cdc help [command]" for more information about a command.

Additional help topics:
{{range .}}{{if not .Runnable}}
    {{.Name | printf "%-11s"}} {{.Short}}{{end}}{{end}}

Use "cdc help [topic]" for more information about that topic.

`

var helpTemplate = `{{if .Runnable}}usage: cdc {{.UsageLine}}

{{end}}{{.Long | trim}}
`

var documentationTemplate = `
/*
{{range .}}{{if .Short}}{{.Short | capitalize}}

{{end}}{{if .Runnable}}Usage:

	cdc {{.UsageLine}}

{{end}}{{.Long | trim}}


{{end}}*/
package main
`

// tmpl executes the given template text on data, writing the result to w.
func tmpl(w io.Writer, text string, data interface{}) {
	t := template.New("top")
	t.Funcs(template.FuncMap{"trim": strings.TrimSpace, "capitalize": capitalize})
	template.Must(t.Parse(text))
	if err := t.Execute(w, data); err != nil {
		panic(err)
	}
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToTitle(r)) + s[n:]
}

func printUsage(w io.Writer) {
	tmpl(w, usageTemplate, commands)
}

// help implements the 'help' command.
func help(args []string) {
	if len(args) == 0 {
		printUsage(os.Stdout)
		// not exit 2: succeeded at 'cdc help'.
		return
	}
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "usage: cdc help command\n\nToo many arguments given.\n")
		os.Exit(2) // failed at 'cdc help'
	}

	arg := args[0]

	// 'cdc help documentation' generates doc.go.
	if arg == "documentation" {
		buf := new(bytes.Buffer)
		printUsage(buf)
		usage := &Command{Long: buf.String()}
		tmpl(os.Stdout, documentationTemplate, append([]*Command{usage}, commands...))
		return
	}

	for _, cmd := range commands {
		if cmd.Name() == arg {
			tmpl(os.Stdout, helpTemplate, cmd)
			// not exit 2: succeeded at 'cdc help cmd'.
			return
		}
	}

	fmt.Fprintf(os.Stderr, "Unknown help topic %#q.  Run 'cdc help'.\n", arg)
	os.Exit(2) // failed at 'cdc help cmd'
}

//
//-----------------------------------------------------[ DBUS CLIENT ACTIONS ]--

type clientAction func(*srvdbus.Client, []string) error

func clientSend(call clientAction, args []string) error {
	srv, e := srvdbus.GetServer()
	switch {
	case e != nil:

	case srv == nil:
		e = errors.New("no service found")

	default:
		e = call(srv, args)
	}
	return e
}

func clientSendLogged(action string, call clientAction, args []string) error {
	return logger.GetErr(clientSend(call, args), action)
}

//
//------------------------------------------------------------------[ ERRORS ]--

func usage() {
	printUsage(os.Stderr)
	exit(2)
}

// Test for error and crash if needed.
// //
func exitIfFail(e error, msg string) {
	if logger.Err(e, msg) {
		exit(1)
	}
}

func exit(status ...int) {
	if len(status) > 0 {
		setExitStatus(status[0])
	}
	os.Exit(exitStatus)
}
