// Use of this source code is governed by a GPL v3 license. See LICENSE file.

package main

import (
	"github.com/sqp/godock/libs/log"

	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"
)

// Commands lists the available commands and help topics.
// The order here is the order in which they are printed by 'cdc help'.
var Commands = CommandList{
	cmdExternal,
	cmdRemote,
	cmdUpload,
	cmdBuild,
	cmdVersion,

	// helpGopath,
	// helpPackages,
	// helpRemote,
	// helpTestflag,
	// helpTestfunc,
}

var (
	// Default command, set by the backend (dock or service).
	cmdDefault *Command

	// Common logger for all services and actions.
	logger = &log.Log{}
)

func main() {
	// Get the command. Exit if fail.
	cmd, args := parseArgsDefault(os.Args, cmdDefault)

	// Free unused.
	Commands, cmdDefault = nil, nil

	// Set logger.
	logger.SetLogOut(log.Logs)
	logger.SetName(strings.TrimPrefix(os.Args[0], "./"))

	// And now we can start.
	cmd.Run(cmd, args)
}

//
//------------------------------------------------------------------[ COMMON ]--

// parseArgsDefault returns the command and its args parsed from startArgs.
// cmd will be used as default if no command name is provided by the user.
//
func parseArgsDefault(startArgs []string, cmd *Command) (*Command, []string) {
	// Add common help features to the default command.
	showHelpShort := cmdDefault.Flag.Bool("h", false, "")
	showHelpLong := cmdDefault.Flag.Bool("help", false, "")

	// Service is the applet start forward. Just run the default with all args.
	if len(startArgs) > 1 && startArgs[1] == "service" {
		return cmd, startArgs[2:]
	}

	// Strip default command name from args of found.
	if len(startArgs) > 1 && startArgs[1] == "dock" { // was cmdDefault.Name()
		startArgs = append(startArgs[:1], startArgs[2:]...)
	}

	// Parse flags as if the default command has been set.
	// No common flag have been set, it will help get the real command.
	cmd.Flag.Parse(startArgs[1:])
	// log.SetFlags(0)
	args := cmd.Flag.Args()

	if *showHelpShort || *showHelpLong {
		help(nil)
	}

	// No command found, return the default one.
	if len(args) == 0 {
		return cmd, startArgs[1:] // return args (nil) ?
	}

	action := findCommandShortcut(startArgs[1])

	// Help display and exit.
	if action == "help" {
		help(args[1:])
	}

	cmd = Commands.MustFind(action)

	// A command has been provided, restore common command arguments parsing.
	// fl := flag.NewFlagSet(startArgs[0], flag.ContinueOnError)
	// fl.Usage = usage
	// fl.Parse(startArgs[1:])
	// args = fl.Args()

	// Get command args, except for commands that will manage args themselves.
	if cmd.CustomFlags {
		args = startArgs[2:]
	} else {
		cmd.Flag.Usage = cmd.Usage
		cmd.Flag.Parse(startArgs[2:])
		args = cmd.Flag.Args()
	}

	// Get the command. Exit if fail.
	return cmd, args
}

func findCommandShortcut(name string) string {
	cmdShortcuts := map[string]string{
		"e":  "external",
		"h":  "help",
		"r":  "remote",
		"up": "upload",
	}
	newname, ok := cmdShortcuts[name]
	if ok {
		return newname
	}
	return name
}

//
//------------------------------------------------------------------[ COMMON ]--

func setPathAbsolute(path *string) {
	if *path != "" {
		newpath, e := filepath.Abs(*path)
		if e != nil {
			return
		}
		*path = newpath
	}
}

//--------------------------------------------------------------------[ HELP ]--

// documentation set by dock or service (whichever one is used).
var (
	usageHeader string
	usageFlags  *string
)

var usageTemplate = `

Usage:

	cdc command [arguments]

The commands are:
{{range .}}{{if .Runnable}}
    {{.Name | printf "%-11s"}} {{.Short}}{{end}}{{end}}

Use "cdc help [command]" for more information about a command.

`

// Additional help topics:
// {{range .}}{{if not .Runnable}}
//     {{.Name | printf "%-11s"}} {{.Short}}{{end}}{{end}}
//
// Use "cdc help [topic]" for more information about that topic.
//

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
	fulltemplate := usageHeader + usageTemplate
	if usageFlags != nil {
		fulltemplate += *usageFlags
	}
	tmpl(w, fulltemplate, Commands)
}

// help implements the 'help' command.
func help(args []string) {
	if len(args) == 0 {
		printUsage(os.Stdout)
		exit()
	}
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "usage: cdc help command\n\nToo many arguments given.\n")
		exit(2) // failed at 'cdc help'
	}

	arg := args[0]

	// 'cdc help documentation' generates doc.go.
	if arg == "documentation" {
		buf := new(bytes.Buffer)
		printUsage(buf)
		usage := &Command{Long: buf.String()}
		tmpl(os.Stdout, documentationTemplate, append([]*Command{usage}, Commands...))
		exit()
	}

	arg = findCommandShortcut(arg) // Enable shortcuts for actions help.

	for _, cmd := range Commands {
		if cmd.Name() == arg {
			tmpl(os.Stdout, helpTemplate, cmd)
			// not exit 2: succeeded at 'cdc help cmd'.
			exit()
		}
	}

	fmt.Fprintf(os.Stderr, "Unknown help topic %#q.  Run 'cdc help'.\n", arg)
	exit(2) // failed at 'cdc help cmd'
}

//
//-----------------------------------------------------------------[ COMMAND ]--

// CommandList defines a list of commands with find by name methods.
//
type CommandList []*Command

// Find returns the command associated with the given name.
//
func (cl CommandList) Find(name string) *Command {
	for _, cmd := range cl {
		if cmd.Name() == name && cmd.Run != nil {
			return cmd
		}
	}
	return nil
}

// MustFind returns the command associated with the given name. Exit if fail.
//
func (cl CommandList) MustFind(name string) *Command {
	cmd := cl.Find(name)
	if cmd == nil {
		fmt.Fprintf(os.Stderr, "cdc: unknown subcommand %q\nRun 'cdc help' for usage.\n", name)
		exit(2)
	}
	return cmd
}

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
	exit(2)
}

// Runnable reports whether the command can be run; otherwise
// it is a documentation pseudo-command such as importpath.
func (c *Command) Runnable() bool {
	return c.Run != nil
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
		os.Exit(status[0])
	}
	os.Exit(0)
}
