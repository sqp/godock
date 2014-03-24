package term

import (
	"github.com/sqp/godock/libs/log"

	"fmt"
	"strconv"
)

//-----------------------------------------------------------------------
// TableFormater
//-----------------------------------------------------------------------

var tableBorder = [][]string{
	{"┌", "┬", "┐"},
	{"├", "┼", "┤"},
	{"└", "┴", "┘"},
}

const (
	tableBorderTop = 0
	tableBorderMid = 1
	tableBorderBot = 2
)

// Format colored console display as table.
type TableFormater struct {
	lines []*Line
	cols  []ColInfo
	max   map[int]int
}

// Create a formater with some columns.
func NewTableFormater(columns []ColInfo) *TableFormater {
	return &TableFormater{
		cols: columns,
		max:  make(map[int]int),
	}
}

// Add a separator line.
func (lf *TableFormater) Separator() {
	line := lf.Line()

	line.separator = true
}

// Create a new line to format.
func (lf *TableFormater) Line() *Line {
	line := NewLine(lf.max)
	lf.lines = append(lf.lines, line)
	return line
}

func (lf *TableFormater) Print() {
	lf.printSeparator(tableBorderTop, true)

	for _, line := range lf.lines {
		if line.separator {
			lf.printSeparator(tableBorderMid, false)
			continue
		}

		format := "│"
		args := []interface{}{}
		for row, _ := range lf.cols {
			format += " %" + lf.cols[row].left() + strconv.Itoa(lf.rowSize(row)+line.colorDelta[row]) + "s │" // using size = default + delta.
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
	out := tableBorder[pos][0]
	first := true
	for id, col := range lf.cols {
		//~ cur := make([]byte, lf.rowSize(id) + 4)
		//~ cur[0] = '+'
		if first {
			first = false
		} else {
			out += tableBorder[pos][1]
		}

		if withTitle {
			out += col.Title
			for i := len(col.Title); i < lf.rowSize(id)+2; i++ {
				out += "─"
			}
		} else {
			for i := 1; i < lf.rowSize(id)+3; i++ {
				//~ cur[i] = '⎼'
				out += "─"
			}
		}
		//~ out += string(cur)
	}
	fmt.Println(out + tableBorder[pos][2])
}

//-----------------------------------------------------------------------
// Column configuration
//-----------------------------------------------------------------------

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

//------------------------------------------------------------------------------
// Line formating.
//------------------------------------------------------------------------------

type Line struct {
	content    map[int]string
	colorDelta map[int]int
	max        map[int]int
	separator  bool
}

func NewLine(max map[int]int) *Line {
	return &Line{
		content:    make(map[int]string),
		colorDelta: make(map[int]int),
		max:        max,
	}
}

// Chainable
func (line *Line) Set(row int, text string) *Line {
	line.testmax(row, len(text))
	line.content[row] = text
	return line
}

func (line Line) Colored(row int, color, text string) {
	origsize := len(text)
	line.testmax(row, origsize) // Size of text without formating.
	line.content[row] = log.Colored(text, color)
	line.colorDelta[row] += len(line.content[row]) - origsize
}

func (line Line) testmax(col, size int) {
	if size > line.max[col] {
		line.max[col] = size
	}
}
