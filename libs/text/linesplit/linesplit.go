// Package linesplit provides a Writer with a callback on every line.
package linesplit

import "strings"

// Writer defines a line splitter Writer.
//
// The buffer will be flushed on Close.
//   Separator: \n
//
type Writer struct {
	str      string
	lineCall func(string)
}

// NewWriter creates a Writer forwarding every line received to callback.
//
func NewWriter(callback func(string)) *Writer {
	return &Writer{lineCall: callback}
}

// Write forwards every line received to the callback (separator \n).
//
func (g *Writer) Write(p []byte) (n int, err error) {
	split := strings.Split(string(string(p)), "\n")
	split[0] = g.str + split[0]

	// Flush all lines except the last one (not finished).
	for _, s := range split[:len(split)-1] {
		g.lineCall(s)
	}

	// Store the remaining as our current line buffer.
	g.str = split[len(split)-1]

	return len(p), nil
}

// Close flushes the remaining buffer content.
//
func (g *Writer) Close() {
	if g.str != "" {
		g.lineCall(g.str)
	}
}
