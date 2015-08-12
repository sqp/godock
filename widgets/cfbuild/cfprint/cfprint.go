// Package cfprint prints config data to the console in a table.
package cfprint

import (
	"github.com/sqp/godock/libs/ternary"
	"github.com/sqp/godock/libs/text/color"
	"github.com/sqp/godock/libs/text/tablist"

	"github.com/sqp/godock/widgets/cfbuild/cftype" // Types for config file builder usage.

	"strconv"
)

// Printer rows.
const (
	RowType = iota
	RowName
	RowOld
	RowNew
)

//
//-----------------------------------------------------------------[ DEFAULT ]--

// Default prints the configuration compared to its default values.
//
func Default(build cftype.Builder, showAll bool) {
	cols := []tablist.ColInfo{
		tablist.NewColLeft(0, "Type"),
		tablist.NewColLeft(0, "Name"),
	}

	if showAll {
		cols = append(cols,
			tablist.NewColLeft(0, "Default"),
		)
	}
	cols = append(cols,
		tablist.NewColLeft(0, "Current"),
	)

	lf := &lineFeed{
		TableFormater: *tablist.NewFormater(cols...),
	}
	if showAll {
		lf.valuePrint = lf.valueDefault
		build.KeyWalk(lf.Add)
	} else {
		build.KeyWalk(lf.addTest)
	}

	lf.Print()

	if lf.countChanged > 0 {
		build.Log().Info("changed", lf.countChanged, "/", lf.countChangeable)
	} else {
		build.Log().Info("nothing changed")
	}
}

func (lf *lineFeed) valueDefault(key *cftype.Key, line tablist.Liner) {
	flag := false

	def, e := key.Storage().Default(key.Group, key.Name)
	if e != nil {
		println("default: ", e.Error())
	}

	valStatus := key.ValueState(def)

	for i, st := range valStatus {
		curline := lf.testNewLine(line, i)
		switch st.State {

		case cftype.StateBothEmpty:
			line.Set(RowOld, "**EMPTY**")

		case cftype.StateUnchanged:
			curline.Set(RowOld, st.Old)

		case cftype.StateEdited:
			curline.Set(RowOld, st.Old)
			if st.New == "" {
				curline.Colored(RowNew, color.FgMagenta, "**EMPTY**")
			} else {
				curline.Colored(RowNew, color.FgGreen, st.New)
			}
			flag = true

		case cftype.StateAdded:
			curline.Set(RowOld, "**EMPTY**")
			curline.Colored(RowNew, color.FgYellow, st.New)
			flag = true

		case cftype.StateRemoved:
			curline.Set(RowOld, st.Old)
			curline.Colored(RowNew, color.FgMagenta, "**EMPTY**")
			flag = true
		}
	}

	if flag {
		lf.countChanged++
	}
}

func (lf *lineFeed) addTest(key *cftype.Key) {
	def, e := key.Storage().Default(key.Group, key.Name)
	if e != nil {
		println("default: ", e.Error())
		return
	}

	valStatus := key.ValueState(def)
	if !valStatus.IsChanged() {
		return
	}

	lf.countChanged++

	lf.valuePrint = func(key *cftype.Key, line tablist.Liner) {
		for i, st := range valStatus {
			curline := lf.testNewLine(line, i)
			switch {
			// case cftype.StateUnchanged:
			// 	curline.Set(RowOld, st.Old)

			case st.State == cftype.StateEdited && st.New != "":
				curline.Set(RowOld, st.New)

			case st.State == cftype.StateAdded:
				curline.Colored(RowOld, color.FgYellow, st.New)

			case st.State == cftype.StateEdited, st.State == cftype.StateRemoved:
				curline.Colored(RowOld, color.FgMagenta, "**EMPTY**")
			}
		}
	}

	lf.Add(key)
}

//
//-------------------------------------------------------------[ LINE FEEDER ]--

// Updated prints the configuration compared to its storage values.
//
func Updated(build cftype.Builder) {
	// build.Log().DEV("save to virtual")
	lf := &lineFeed{
		TableFormater: *tablist.NewFormater(
			tablist.NewColLeft(0, "Type"),
			tablist.NewColLeft(0, "Name"),
			tablist.NewColLeft(0, "Old value"),
			tablist.NewColLeft(0, "New value"),
		),
	}
	lf.valuePrint = lf.valueUpdated

	build.KeyWalk(lf.Add)
	lf.Print()

	if lf.countChanged > 0 {
		build.Log().Info("changed", lf.countChanged, "/", lf.countChangeable)
	} else {
		build.Log().Info("nothing changed")
	}
}

func (lf *lineFeed) valueUpdated(key *cftype.Key, line tablist.Liner) {
	flag := false
	older := key.Storage().Valuer(key.Group, key.Name)
	valStatus := key.ValueState(older)

	for i, st := range valStatus {
		curline := lf.testNewLine(line, i)
		switch st.State {

		case cftype.StateBothEmpty:
			line.Colored(RowOld, color.FgMagenta, "**EMPTY**")
			line.Colored(RowNew, color.BgRed, "  ==  ")

		case cftype.StateUnchanged:
			curline.Set(RowOld, st.Old)
			curline.Colored(RowNew, color.BgRed, "  ==  ")

		case cftype.StateEdited:
			curline.Set(RowOld, st.Old)
			st.New = ternary.String(st.New == "", "**EMPTY**", st.New)
			curline.Colored(RowNew, color.FgGreen, st.New)
			flag = true

		case cftype.StateAdded:
			curline.Set(RowOld, "**EMPTY**")
			curline.Colored(RowNew, color.FgGreen, st.New)
			flag = true

		case cftype.StateRemoved:
			curline.Set(RowOld, st.Old)
			curline.Colored(RowNew, color.FgGreen, "**EMPTY**")
			flag = true
		}
	}

	if flag {
		lf.countChanged++
	}
}

//
//-------------------------------------------------------------[ LINE FEEDER ]--

type lineFeed struct {
	tablist.TableFormater

	valuePrint func(key *cftype.Key, line tablist.Liner)

	lastGroup       string
	hasFrame        bool
	countChanged    int
	countChangeable int
}

// Add adds a key to the printer.
//
func (lf *lineFeed) Add(key *cftype.Key) {
	if key.Group != lf.lastGroup {
		lf.AddGroup(RowName, key.Group)
		lf.lastGroup = key.Group
		lf.hasFrame = false
	}

	var line tablist.Liner
	var title string
	switch {
	case key.IsType(cftype.KeyFrame, cftype.KeyExpander):
		line = lf.AddEmptyFilled()
		lf.hasFrame = true
		line.Set(RowType, key.Type.String())
		if len(key.AuthorizedValues) == 0 {
			title = "[*FRAME NO TITLE*]"
		} else {
			title = key.AuthorizedValues[0]
		}

	case key.IsType(cftype.KeySeparator):
		line = lf.AddEmptyFilled()
		line.Set(RowType, key.Type.String())
		title = lf.indent() + "---------"

	case key.IsType(cftype.KeyTextLabel, cftype.KeyLaunchCmdSimple):
		line = lf.AddEmptyFilled()
		line.Set(RowType, key.Type.String())
		title = lf.indent() + key.Name

	default:
		line = lf.AddLine()
		lf.valuePrint(key, line)
		line.Colored(RowType, color.FgGreen, key.Type.String())
		title = lf.indent() + key.Name
		lf.countChangeable++
	}

	line.Set(RowName, title)
}

// indent returns 2 space of indent for keys in a frame.
//
func (lf *lineFeed) indent() string {
	return ternary.String(lf.hasFrame, "  ", "")
}

// testNewLine creates a new line if the valueID is greater than 0.
//
func (lf *lineFeed) testNewLine(line tablist.Liner, valueID int) tablist.Liner {
	if valueID > 0 {
		reline := lf.AddLine()
		reline.Set(RowType, "    +value "+strconv.Itoa(valueID+1))
		reline.Set(RowName, lf.indent()+"   -----")
		return reline
	}
	return line
}
