// Package versions prints API and dependencies versions.
package versions

import (
	"github.com/sqp/godock/libs/cdglobal"     // Dock types.
	"github.com/sqp/godock/libs/text/color"   // Colored text.
	"github.com/sqp/godock/libs/text/strhelp" // String helpers.

	"fmt"
	"runtime"
)

var (
	// Name defines the name printed for the program. By default its an applet.
	// Overridden with the dock build tag.
	//
	Name = "Applets API"

	// List stores the list of versions activated by build tags.
	//
	List Fields
)

// Field defines a printable field with its version text.
//
type Field struct{ K, V string }

// Fields defines a list of field to format.
//
type Fields []Field

// String formats the fields content with colored output.
//
func (fields Fields) String() string {
	fields = append([]Field{
		{Name, cdglobal.AppVersion},
		{"  go       ", fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)},
	}, List...)

	if cdglobal.BuildDate != "" {
		fields = append(fields,
			Field{"  Compiled ", cdglobal.BuildDate},
			Field{"  Git Hash ", cdglobal.GitHash},
		)
	}

	out := ""
	for _, line := range fields {
		out += strhelp.Bracket(color.Colored(line.K, color.FgGreen)) + " " + line.V + "\n"
	}
	return out
}

//
//-------------------------------------------------------------------[ PRINT ]--

// Print prints versions numbers.
//
func Print() {
	print(List.String())
}

// TestPrint prints versions numbers if the value is true.
//
func TestPrint(display bool) {
	if !display {
		return
	}
	Print()
}
