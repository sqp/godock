// Hacked from the Go command list.go file by SQP.
// Use of this source code is governed by a GPL v3 license. See LICENSE file.
// Original work was Copyright 2011 The Go Authors with a BSD-style license

package main

import (
	"github.com/sqp/godock/libs/cdglobal"
	"github.com/sqp/godock/libs/packages"
	"github.com/sqp/godock/libs/text/color"
	"github.com/sqp/godock/libs/text/tablist"

	"bufio"
	"encoding/json"
	"io"
	"os"
	"strings"
	"text/template"
)

var nl = []byte{'\n'}

var cmdExternal = &Command{
	UsageLine: "external [-d path] [-r] [appletname...]",
	Short:     "manage external applets",
	Long: `
External lists, installs or removes Cairo-Dock external applets.

The action depends if applet names are provided, and on the -r flag:
  no applet name        Display the list of external applets.
  one or more           Install applet(s).
  one or more and -r    Remove applet(s).

Common flags:
  -d path      Use a custom config directory. Default: ~/.config/cairo-dock

Install flags:
  -r           Remove applets instead of install.
  -v           Verbose output for files extraction.

List matching flags:
  -s           Only match applets found on the applet server.
  -l           Only match applets found locally.

List display flags:
  -json        Print applet package data in JSON format.
  -f template  Set a specific format for the list, with the go template syntax.  


You can use a template to format the result the way you want.
For example, to have just the applet name: -f '{{.DisplayedName}}'.
Everything is possible like '{{.DisplayedName}}  by {{.Author}}'.

For the full list of fields, see AppletPackage:
  http://godoc.org/github.com/sqp/godock/libs/packages#AppletPackage
`,
}

// TODO:
// add output gob

func init() {
	cmdExternal.Run = runList // break init cycle
}

var listUserDir = cmdExternal.Flag.String("d", "", "")
var listRemove = cmdExternal.Flag.Bool("r", false, "")
var listVerbose = cmdExternal.Flag.Bool("v", false, "")

var listServer = cmdExternal.Flag.Bool("s", false, "")
var listLocal = cmdExternal.Flag.Bool("l", false, "")

var listJSON = cmdExternal.Flag.Bool("json", false, "")
var listFmt = cmdExternal.Flag.String("f", "", "")

func runList(cmd *Command, args []string) {

	setPathAbsolute(listUserDir) // Ensure we have an absolute path for the config dir.

	if len(args) > 0 { // List of applets names provided (at least one).
		installOrRemoveApplets(args, *listRemove)
		return
	}

	packages, e := listPackages()
	if logger.Err(e, "get packages list") {
		return
	}

	// Print formated list.
	switch {
	case *listJSON:
		printJSON(packages)

	case *listFmt != "":
		printTemplate(packages)

	default:
		printConsole(packages)
	}
}

//
//----------------------------------------------------------------[ INSTALL ]--

func installOrRemoveApplets(list []string, remove bool) {
	externalUserDir, e := packages.DirAppletsExternal(*listUserDir)
	exitIfFail(e, "get config dir") // Ensure we have the config dir.

	var packs packages.AppletPackages
	var action func(appname string, pack *packages.AppletPackage) bool

	if remove {
		action = func(appname string, pack *packages.AppletPackage) bool {
			e := pack.Uninstall(externalUserDir)
			return testErr(e, "uninstall", "Applet removed", appname)
		}
		packs, e = packages.ListFromDir(externalUserDir, packages.TypeUser, packages.SourceApplet)

	} else { // install.
		options := ""
		if *listVerbose {
			options = "v" // Tar command verbose option.
		}
		action = func(appname string, pack *packages.AppletPackage) bool {
			pack.SrvTag = cdglobal.AppletsDirName + "/" + cdglobal.AppletsServerTag
			e := pack.Install(externalUserDir, options)
			return testErr(e, "install", "Applet installed", appname)
		}
		packs, e = packages.ListDistant(cdglobal.AppletsDirName + "/" + cdglobal.AppletsServerTag)
	}
	exitIfFail(e, "get applets list") // Ensure we have the server list.

	failed := false
	for _, appname := range list {
		pack := packs.Get(strings.Title(appname)) // Applets are using a CamelCase format. This will help lazy users
		if pack != nil {
			failed = failed || !action(appname, pack)
		}
	}
	if failed || len(list) == 0 {
		println("use list command to get the list of valid applets names")
	}
}

func testErr(e error, msgFail, msgOK, appname string) bool {
	if logger.Err(e, msgFail) {
		return false
	}
	logger.Info(msgOK, appname)
	return true
}

//
//---------------------------------------------------------------[ LIST DATA ]--

// Listing arguments: get data.

func listPackages() (list packages.AppletPackages, e error) {

	// List server only.
	if *listServer {
		return packages.ListDistant(cdglobal.AppletsDirName + "/" + cdglobal.AppletsServerTag)
	}

	// Get applets dir.
	externalUserDir, e := packages.DirAppletsExternal(*listUserDir)
	if e != nil {
		return nil, e
	}

	// List local only.
	if *listLocal {
		return packages.ListFromDir(externalUserDir, packages.TypeUser, packages.SourceApplet)
	}

	// List default (merged both).
	packs, e := packages.ListDownloadApplets(externalUserDir)
	if e != nil {
		return nil, e
	}
	return packages.ListDownloadSort(packs), nil
}

//
//---------------------------------------------------------------[ FORMATERS ]--

// Format applet packages using simple table formater.
//
func printConsole(list packages.AppletPackages) {
	lf := tablist.NewFormater(
		tablist.NewColRight(0, "Inst"),
		tablist.NewColLeft(0, "[ Applet ]"),
		tablist.NewColLeft(0, "Category"),
	)

	for _, pack := range list {
		line := lf.AddLine()
		if pack.Type == packages.TypeUser {
			line.Colored(0, color.FgGreen, " * ")
		}
		if pack.Type == packages.TypeInDev {
			line.Colored(0, color.FgYellow, " * ")
		}

		line.Set(1, pack.DisplayedName)
		cat, _ := packages.FormatCategory(pack.Category)
		line.Set(2, cat)
	}
	lf.Print()
}

// Format applet packages using json encoder.
//
func printJSON(list packages.AppletPackages) {
	out := newCountingWriter(os.Stdout)
	defer out.w.Flush()
	for _, p := range list {
		b, err := json.MarshalIndent(p, "", "\t")
		if logger.Err(err, "printJSON") {
			out.Flush()
			exit(1)
		}
		out.Write(b)
		out.Write(nl)
	}
}

// Format applet packages using golang template formater.
//
func printTemplate(list packages.AppletPackages) {
	tmpl, err := template.New("main").Parse(*listFmt)
	exitIfFail(err, "")
	out := newCountingWriter(os.Stdout)
	defer out.w.Flush()
	for _, p := range list {
		out.Reset()
		if err := tmpl.Execute(out, p); logger.Err(err, "printTemplate") {
			out.Flush()
			exit(1)
		}
		if out.Count() > 0 {
			out.w.WriteRune('\n')
		}
	}
}

//---------------------------------------------------------[ COUNTING WRITER ]--

// CountingWriter counts its data, so we can avoid appending a newline
// if there was no actual output.
//
type CountingWriter struct {
	w     *bufio.Writer
	count int64
}

func newCountingWriter(w io.Writer) *CountingWriter {
	return &CountingWriter{
		w: bufio.NewWriter(w),
	}
}

// Write add text to the writer.
//
func (cw *CountingWriter) Write(p []byte) (n int, err error) {
	cw.count += int64(len(p))
	return cw.w.Write(p)
}

// Flush the writer.
//
func (cw *CountingWriter) Flush() {
	cw.w.Flush()
}

// Reset the writer.
//
func (cw *CountingWriter) Reset() {
	cw.count = 0
}

// Count the writer data.
//
func (cw *CountingWriter) Count() int64 {
	return cw.count
}
