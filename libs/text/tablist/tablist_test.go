package tablist_test

import "github.com/sqp/godock/libs/text/tablist"

func Example() {
	// Create a table with some columns.
	lf := tablist.NewFormater(
		tablist.NewColRight(0, "column 0"),
		tablist.NewColLeft(0, "column 1"),
		tablist.NewColLeft(0, "column 2"),
	)

	// Add a few lines of data.
	for _, item := range []struct {
		c0, c1, c2 string
	}{
		{"content", "this is on the left", "also on the left"},
		{"aligned", "with an empty cell", "for"},
		{"on the right", "", "a simple table"},
	} {
		line := lf.AddLine()
		line.Set(0, item.c0) // You better use constants to refer to your column ID.
		line.Set(1, item.c1)
		line.Set(2, item.c2)
	}

	// Prints the result to console.
	lf.Print()

	// Output:
	// ┌column 0──────┬column 1─────────────┬column 2──────────┐
	// │      content │ this is on the left │ also on the left │
	// │      aligned │ with an empty cell  │ for              │
	// │ on the right │                     │ a simple table   │
	// └──────────────┴─────────────────────┴──────────────────┘
}
