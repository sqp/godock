// Hacked from the Go command list.go file by SQP.
// Use of this source code is governed by a GPL v3 license. See LICENSE file.
// Original work was Copyright 2011 The Go Authors with a BSD-style license

package main

import (
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/packages"
	"github.com/sqp/godock/libs/text/color"
	"github.com/sqp/godock/libs/text/tablist"

	"bufio"
	"encoding/json"
	"io"
	"os"
	"text/template"
)

var nl = []byte{'\n'}

var cmdList = &Command{
	UsageLine: "list [-d] [-l] [-f format] [-json]",
	Short:     "list external applets",
	Long: `
List lists Cairo-Dock external applets with installed state.

The -d flag will only match applets found on the applet market.

The -l flag will only match applets found locally.

The -f flag specifies an specific format for the list,
using the syntax of applets template.  You can use it to
format the result the way you want. For example, to have just
the applet name: -f '{{.DisplayedName}}'.  Everything is
possible like '{{.DisplayedName}}  by {{.Author}}'. The struct
being passed to the template is:

  type AppletPackage struct {
	DisplayedName string      // name of the applet
	Author        string      // author(s)
	Description   string
	Category      int
	Version       string
	ActAsLauncher bool

	Type          PackageType // type of applet : installed, user, distant...
	Path          string      // complete path of the package.
	LastModifDate int         // date of latest changes in the package.
	Size float64              // size in Mo

	// On server only.
	CreationDate  int         // date of creation of the package.
  }

The -json flag causes the applet package data to be printed in JSON format
instead of using the template format.
	`,
}

// TODO:
// check args unused
// add output gob

func init() {
	cmdList.Run = runList // break init cycle
	// cmdList.Flag.Var(buildCompiler{}, "compiler", "") // ??
}

var listLocal = cmdList.Flag.Bool("l", false, "")
var listDistant = cmdList.Flag.Bool("d", false, "")

var listFmt = cmdList.Flag.String("f", "", "")
var listJSON = cmdList.Flag.Bool("json", false, "")

func runList(cmd *Command, args []string) {

	// Listing arguments: get data.
	var listPackages packages.AppletPackages
	var e error
	switch {
	case *listDistant:
		listPackages, e = packages.ListDistant(cdtype.AppletsDirName + "/" + cdtype.AppletsServerTag)
		logger.Err(e, "get packages list from server")
	case *listLocal:
		// Get applets dir.
		dir, e := packages.DirExternal()
		if e != nil {
			return
		}
		listPackages, e = packages.ListFromDir(dir, packages.TypeUser, packages.SourceApplet)
		logger.Err(e, "get packages list from external dir")

	default:
		listPackages, e = packages.ListDownloadSorted(cdtype.AppletsServerTag)
		logger.Err(e, "get packages list from server")
	}

	// Print formated list.
	switch {
	case *listJSON:
		printJSON(listPackages)

	case *listFmt != "":
		printTemplate(listPackages)

	default:
		printConsole(listPackages)

	}
}

//-----------------------------------------------------[ FORMATERS ]--

// Format applet packages using simple table formater.
//
func printConsole(list packages.AppletPackages) {
	lf := tablist.NewTableFormater([]tablist.ColInfo{
		{0, false, "Inst"},
		{0, true, "[ Applet ]"},
		{0, true, "Category"},
	})

	for _, pack := range list {
		line := lf.Line()
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

//-----------------------------------------------------[ COUNTING WRITER ]--

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
