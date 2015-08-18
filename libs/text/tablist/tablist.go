// Package tablist is a simple table formatter with colors for console.
package tablist

import (
	"github.com/sqp/godock/libs/text/color"

	"fmt"
	"strconv"
	"strings"
)

//
//----------------------------------------------------------[ TABLE FORMATER ]--

// GroupColor defines the common group text color.
//
var GroupColor = color.FgMagenta

var tableBorder = []string{
	"┌", "┬", "┐",
	"├", "┼", "┤",
	"└", "┴", "┘",
}

var sepVertical = []string{"│", "│", "│"}

const sepHorizontal = "─"

const (
	tableBorderTop = 0
	tableBorderMid = 1
	tableBorderBot = 2
)

func tabSep(pos int) []string {
	return tableBorder[3*pos:]
}

// TableFormater format colored console display as table.
//
type TableFormater struct {
	Base
	lines []Liner
}

// NewFormater create a TableFormater with some columns.
//
func NewFormater(columns ...ColInfo) *TableFormater {
	return &TableFormater{
		Base: *NewBase(columns),
	}
}

// Count returns the number of lines.
//
func (lf *TableFormater) Count() int {
	return len(lf.lines)
}

// Print prints the table content in console output.
//
func (lf *TableFormater) Print() {
	fmt.Println(lf.Sprint())
}

// Sprint returns the table content text as if printed on console.
//
func (lf *TableFormater) Sprint() string {
	var content string
	for _, line := range lf.lines {
		content += line.Sprint() + "\n"
	}
	return lf.WalkSprint(tabSep(tableBorderTop), lf.sprintSeparator) +
		"\n" +
		content +
		lf.WalkSprint(tabSep(tableBorderBot), lf.sprintFooter)
}

// AddLine create and append a new line to format.
//
func (lf *TableFormater) AddLine() Liner {
	line := NewLine(lf.Base)
	lf.lines = append(lf.lines, line)
	return line
}

// AddEmptyFilled create and append a new line to format that fills empty fields.
//
func (lf *TableFormater) AddEmptyFilled() Liner {
	line := newEmptyFilled(lf.Base)
	lf.lines = append(lf.lines, line)
	return line
}

// AddGroup create and append a new group line to format.
//
func (lf *TableFormater) AddGroup(row int, name string) {
	line := newGroup(lf.Base)
	line.Colored(row, GroupColor, name)
	lf.lines = append(lf.lines, line)
}

// AddSeparator add a separator line.
//
func (lf *TableFormater) AddSeparator() {
	lf.lines = append(lf.lines, newSeparator(lf.Base, lf))
}

//
//--------------------------------------------------------------------[ BASE ]--

// Base defines the TableFormater and the Line base with columns definition and
// the column size map.
//
type Base struct {
	cols []ColInfo
	max  map[int]int
}

// NewBase creates a base for TableFormater with some columns.
//
func NewBase(columns []ColInfo) *Base {
	return &Base{
		cols: columns,
		max:  make(map[int]int),
	}
}

// WalkSprint runs the given call on each cell of the line and returns the line
// printable content.
//
//   sep[0] is added at the begining of the line.
//   sep[1] is added between each cell.
//   sep[2] is added at the end of the line.
// The result if like this:
//   sep0 cell0 sep1 cell1 sep1 cell2 sep1 cell3 sep2
//
func (o *Base) WalkSprint(sep []string, call func(int, ColInfo) string) string {
	var out string
	for id, col := range o.cols {
		if id > 0 {
			out += sep[1]
		}
		out += call(id, col)
	}
	return fmt.Sprint(sep[0] + out + sep[2])
}

// WalkSprintf runs the given get format call on each cell and prints the line.
//
func (o *Base) WalkSprintf(sep []string, call func(int, ColInfo) (format, content string)) string {
	var (
		args   []interface{}
		format string
	)
	for id, col := range o.cols {
		cformat, ccontent := call(id, col)
		if id > 0 {
			format += sep[1]
		}
		format += cformat
		args = append(args, ccontent)
	}
	return fmt.Sprintf(sep[0]+format+sep[2], args...)
}

// ColSize returns the size of the given column.
//
func (o *Base) ColSize(col int) int {
	if o.cols[col].Size == 0 {
		return o.max[col]
	}
	return o.cols[col].Size
}

// SetColSize sets the size of the given column (if it is larger than before).
//
func (o *Base) SetColSize(col, size int) {
	if size > o.max[col] {
		o.max[col] = size
	}
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

// NewColLeft creates a column align on the left.
//
func NewColLeft(size int, title string) ColInfo {
	return ColInfo{
		Size:  size,
		Left:  true,
		Title: title,
	}
}

// NewColRight creates a column align on the right.
//
func NewColRight(size int, title string) ColInfo {
	return ColInfo{
		Size:  size,
		Title: title,
	}
}

//
//----------------------------------------------------------------[ LINE API ]--

// Liner defines the common line API.
//
type Liner interface {
	Sprint() string
	Set(row int, text string) *Line
	Colored(row int, newcolor, text string)
}

//
//--------------------------------------------------------------------[ LINE ]--

// Line defines a table printer basic line.
//
type Line struct {
	Base
	content    map[int]string
	colorDelta map[int]int
}

// NewLine create a new line.
//
func NewLine(base Base) *Line {
	return &Line{
		content:    make(map[int]string),
		colorDelta: make(map[int]int),
		Base:       base,
	}
}

// Sprint prints the line content.
//
func (line *Line) Sprint() string {
	return line.WalkSprintf(sepVertical, line.CellData)
}

// CellData returns fmt format and argument for the cell.
//
func (line *Line) CellData(id int, _ ColInfo) (format, content string) {
	return line.rowFormat(id), line.content[id]
}

func (line *Line) rowFormat(col int) (format string) {
	sign := ""
	if line.Base.cols[col].Left {
		sign = "-"
	}
	size := strconv.Itoa(line.ColSize(col) + line.colorDelta[col]) //  default + delta.
	return " %" + sign + size + "." + size + "s "                  // force min and max size to fill exactly the cell.
}

// Set text content for given row.
//
func (line *Line) Set(row int, text string) *Line {
	line.SetColSize(row, len(text))
	line.content[row] = text
	return line
}

// Colored set colored text content for given row.
//
func (line Line) Colored(row int, newcolor, text string) {
	origsize := len(text)
	line.SetColSize(row, origsize) // Size of text without formating.
	if newcolor == "" {
		line.content[row] = text
	} else {
		line.content[row] = color.Colored(text, newcolor)
		line.colorDelta[row] += len(line.content[row]) - origsize
	}
}

func (line *Line) dash(row int) (out string) {
	return strings.Repeat(sepHorizontal, line.ColSize(row)-len(line.content[row])+line.colorDelta[row]+2)
}

//
//---------------------------------------------------------------[ SEPARATOR ]--

type separator struct{ Line }

func newSeparator(base Base, lf *TableFormater) *separator { return &separator{*NewLine(base)} }

func (o *separator) Sprint() string {
	return o.WalkSprint(tabSep(tableBorderMid), o.sprintFooter)
}

//
//------------------------------------------------------------[ EMPTY FILLED ]--

type emptyFilled struct{ Line }

func newEmptyFilled(base Base) *emptyFilled { return &emptyFilled{*NewLine(base)} }

func (o *emptyFilled) Sprint() string {
	return o.WalkSprint(sepVertical, o.sprintEmptyFilled)
}

//
//-------------------------------------------------------------------[ GROUP ]--

type group struct{ Line }

func newGroup(base Base) *group { return &group{*NewLine(base)} }

func (o *group) Sprint() string {
	return o.WalkSprint(tabSep(tableBorderMid), o.sprintContentDash)
}

//
//----------------------------------------------------------[ BASE FORMATERS ]--

func (o *Base) sprintSeparator(id int, col ColInfo) (out string) {
	max := o.ColSize(id) + 2
	if len(col.Title) > max {
		out = col.Title[:max]
	} else {
		out = col.Title
	}
	return out + strings.Repeat(sepHorizontal, max-len(out))
}

func (o *Base) sprintFooter(id int, col ColInfo) (out string) {
	return out + strings.Repeat(sepHorizontal, o.ColSize(id)+2-len(out))
}

//
//----------------------------------------------------------[ LINE FORMATERS ]--

func (line *Line) sprintContentDash(id int, col ColInfo) (out string) {
	return line.content[id] + line.dash(id)
}

func (line *Line) sprintEmptyFilled(id int, col ColInfo) (out string) {
	if len(line.content[id]) > 0 {
		return fmt.Sprintf(line.CellData(id, col))
	}
	return line.dash(id)
}
