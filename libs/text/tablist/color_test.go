package tablist_test

import (
	"github.com/stretchr/testify/assert"

	"github.com/sqp/godock/libs/text/color"
	"github.com/sqp/godock/libs/text/tablist"

	"testing"
)

func TestTableColor(t *testing.T) {
	tablist.GroupColor = color.FgRed
	// Create a table with some columns.
	lf := tablist.NewFormater(
		tablist.NewColRight(0, "column 0"),
	)
	lf.AddGroup(0, "group")

	assert.Equal(t, 1, lf.Count(), "Count")
	groupColor := color.FgRed + "group" + color.Reset
	output := `┌column ┐
├` + groupColor + `──┤
└───────┘`
	assert.Equal(t, output, lf.Sprint(), "Count")

}
