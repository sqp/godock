// Package versions prints API and dependencies versions.
package versions

import (
	"github.com/sqp/godock/libs/cdglobal"     // Dock types.
	"github.com/sqp/godock/libs/text/color"   // Colored text.
	"github.com/sqp/godock/libs/text/tablist" // Format table.

	"fmt"
	"runtime"
)

var (
	// name stores the printed name for the program or backend.
	name = "missing, use versions.SetName"

	// Dock stores dock and deps versions.
	Dock = []Field{}
)

// SetName sets the name of the program or backend used.
//
func SetName(backendName string) {
	name = backendName
}

// Field defines a printable field with its version text.
//
type Field struct{ K, V string }

// Fields returns the list of fields to format.
//
func Fields() []Field {
	list := append([]Field{{name, cdglobal.AppVersion}}, Dock...)

	list = append(list, []Field{
		{"OS", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)},
		{"Compiler", runtime.Version()},
		{"BuildMode", cdglobal.BuildMode},
	}...)

	if cdglobal.BuildDate != "" {
		list = append(list,
			Field{"Date", cdglobal.BuildDate},
			Field{"CommitID", cdglobal.GitHash},
			Field{"FilesEdited", cdglobal.BuildNbEdited},
		)
	}
	return list
}

//
//-------------------------------------------------------------------[ PRINT ]--

// Print prints versions numbers.
//
func Print() {
	fmt.Println(Format(Fields()))
}

// Format formats fields content with colored output.
//
func Format(fields []Field) (out string) {
	lf := tablist.NewFormater(
		tablist.NewColLeft(0, ""),
		tablist.NewColLeft(0, fmt.Sprintf("[ %s  %s ]", fields[0].K, fields[0].V)),
	)

	for _, field := range fields[1:] {
		lf.AddLine().
			Colored(0, color.FgGreen, field.K).
			Set(1, field.V)
	}
	return lf.Sprint()
}
