// Package term is a simple table formatter with colors for console.
package term

import (
	colors "github.com/sqp/godock/libs/log/color"

	"fmt"
	"strconv"
)

//
//----------------------------------------------------------[ TABLE FORMATER ]--

var tableBorder = []string{
	"┌", "┬", "┐",
	"├", "┼", "┤",
	"└", "┴", "┘",
}

const (
	tableBorderTop = 0
	tableBorderMid = 1
	tableBorderBot = 2
)

// TableFormater format colored console display as table.
//
type TableFormater struct {
	lines []*Line
	cols  []ColInfo
	max   map[int]int
}

// NewTableFormater create a TableFormater with some columns.
//
func NewTableFormater(columns []ColInfo) *TableFormater {
	return &TableFormater{
		cols: columns,
		max:  make(map[int]int),
	}
}

// Separator add a separator line.
//
func (lf *TableFormater) Separator() {
	line := lf.Line()

	line.separator = true
}

// Line create and append a new line to format.
//
func (lf *TableFormater) Line() *Line {
	line := NewLine(lf.max)
	lf.lines = append(lf.lines, line)
	return line
}

// Print the table content in console output.
//
func (lf *TableFormater) Print() {
	lf.printSeparator(tableBorderTop, true)

	for _, line := range lf.lines {
		if line.separator {
			lf.printSeparator(tableBorderMid, false)
			continue
		}

		format := "│"
		args := []interface{}{}
		for row := range lf.cols {
			format += " %" +
				lf.cols[row].left() + // negative sign if needed.
				strconv.Itoa(lf.rowSize(row)+line.colorDelta[row]) + // size = default + delta.
				"s │"
			args = append(args, line.content[row])
		}
		fmt.Printf(format+"\n", args...)
	}
	lf.printSeparator(tableBorderBot, false)
}

func (lf *TableFormater) rowSize(row int) int {
	if lf.cols[row].Size == 0 {
		return lf.max[row]
	}
	return lf.cols[row].Size
}

func (lf *TableFormater) printSeparator(pos int, withTitle bool) {
	out := tableBorder[3*pos]
	first := true
	for id, col := range lf.cols {
		if first {
			first = false
		} else {
			out += tableBorder[3*pos+1]
		}

		if withTitle {
			out += col.Title
			for i := len(col.Title); i < lf.rowSize(id)+2; i++ {
				out += "─"
			}
		} else {
			for i := 1; i < lf.rowSize(id)+3; i++ {
				out += "─"
			}
		}
	}
	fmt.Println(out + tableBorder[3*pos+2])
}

//
//-----------------------------------------------------------[ COLUMN CONFIG ]--

// ColInfo is the configuration of a table column.
//
type ColInfo struct {
	Size  int
	Left  bool
	Title string
}

func (info ColInfo) left() string {
	if info.Left {
		return "-"
	}
	return ""
}

//
//-----------------------------------------------------------[ LINE FORMATER ]--

// Line is a table line with .
//
type Line struct {
	content    map[int]string
	colorDelta map[int]int
	max        map[int]int
	separator  bool
}

// NewLine create a new line.
//
func NewLine(max map[int]int) *Line {
	return &Line{
		content:    make(map[int]string),
		colorDelta: make(map[int]int),
		max:        max,
	}
}

// Set text content for given row.
//
func (line *Line) Set(row int, text string) *Line {
	line.testmax(row, len(text))
	line.content[row] = text
	return line
}

// Colored set colored text content for given row.
//
func (line Line) Colored(row int, color, text string) {
	origsize := len(text)
	line.testmax(row, origsize) // Size of text without formating.
	line.content[row] = colors.Colored(text, color)
	line.colorDelta[row] += len(line.content[row]) - origsize
}

func (line Line) testmax(col, size int) {
	if size > line.max[col] {
		line.max[col] = size
	}
}
