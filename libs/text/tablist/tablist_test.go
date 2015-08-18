package tablist_test

import "github.com/sqp/godock/libs/text/tablist"

func Example() {
	// Create a table with some columns.
	lf := tablist.NewFormater(
		tablist.NewColRight(0, "column 0"),
		tablist.NewColLeft(0, "column 1"),
		tablist.NewColLeft(23, "column 2"),
	)

	// Add a few lines of data.
	for _, item := range []struct {
		c0, c1, c2 string
	}{
		{"content", "this is on the left", "also left"},
		{"aligned", "with an empty cell", "with fixed size"},
		{"on the right", "", "can truncate the content too large"},
	} {
		line := lf.AddLine()
		line.Set(0, item.c0) // You better use constants to refer to your column ID.
		line.Set(1, item.c1)
		line.Set(2, item.c2)
	}

	lf.AddSeparator()

	sepInfo := lf.AddLine()
	sepInfo.Set(1, "AddSeparator")
	sepInfo.Set(2, "separator visible above")

	// Disable the group color for the example test.
	tablist.GroupColor = ""
	lf.AddGroup(1, "AddGroup")

	lineInfo := lf.AddLine()
	lineInfo.Set(1, "AddLine")
	lineInfo.Set(2, "a basic line")

	emptyFilled := lf.AddEmptyFilled()
	emptyFilled.Set(1, "AddEmptyFilled")
	emptyFilled.Set(2, "fills empty cells")

	// Prints the result to console.
	lf.Print()

	// Output:
	// ┌column 0──────┬column 1─────────────┬column 2─────────────────┐
	// │      content │ this is on the left │ also left               │
	// │      aligned │ with an empty cell  │ with fixed size         │
	// │ on the right │                     │ can truncate the conten │
	// ├──────────────┼─────────────────────┼─────────────────────────┤
	// │              │ AddSeparator        │ separator visible above │
	// ├──────────────┼AddGroup─────────────┼─────────────────────────┤
	// │              │ AddLine             │ a basic line            │
	// │──────────────│ AddEmptyFilled      │ fills empty cells       │
	// └──────────────┴─────────────────────┴─────────────────────────┘
}
