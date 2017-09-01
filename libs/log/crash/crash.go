package crash

import (
	"github.com/maruel/panicparse/stack"

	"github.com/sqp/godock/libs/dock/confown"
	"github.com/sqp/godock/libs/text/color"

	"bytes"
	"fmt"
	"os"
	"runtime"
)

// Palette defines a color palette based on ANSI code.
//
var Palette = stack.Palette{
	EOLReset:               color.Reset,
	RoutineFirst:           color.FgMagenta,
	CreatedBy:              color.FgCyan, //   ansi.LightBlack,
	Package:                color.Bright, //   ansi.ColorCode("default+b"),
	SourceFile:             color.Reset,
	FunctionStdLib:         color.FgGreen,
	FunctionStdLibExported: color.FgGreen,
	FunctionMain:           color.FgYellow,
	FunctionOther:          color.FgRed,
	FunctionOtherExported:  color.FgRed,
	Arguments:              color.Reset,
}

// Parse returns a shorter colored parsed crash message.
//
func Parse(r interface{}, showCaller bool) string {
	deb := make([]byte, 1024)
	for {
		n := runtime.Stack(deb, false)
		if n < len(deb) {
			deb = deb[:n]
			break
		}
		deb = make([]byte, 2*len(deb))
	}

	if !confown.Current.CrashDisplayColored {
		return fmt.Sprintf("%s\n%s\n", r, deb)
	}

	out := fmt.Sprintf("\n=====================================[ dock panic ]==\n%s\n\n", r)

	in := bytes.NewBuffer(deb)
	goroutines, e := stack.ParseDump(in, os.Stdout)
	if e != nil {
		return out
	}

	// Improve values display.
	stack.Augment(goroutines)

	buckets := stack.SortBuckets(stack.Bucketize(goroutines, stack.AnyValue))
	srcLen, pkgLen := stack.CalcLengths(buckets, false)
	if len(buckets) == 1 {
		buckets[0].Stack.Calls = buckets[0].Stack.Calls[3:]
		if showCaller {
			out += fmt.Sprint(Palette.BucketHeader(&buckets[0], false, false))
		}
		out += fmt.Sprint(Palette.StackLines(&buckets[0].Signature, srcLen, pkgLen, false))

	} else {
		for _, bucket := range buckets {
			out += fmt.Sprint(Palette.BucketHeader(&bucket, false, len(buckets) > 1))
			out += fmt.Sprint(Palette.StackLines(&bucket.Signature, srcLen, pkgLen, false))
		}
	}

	return out
}
