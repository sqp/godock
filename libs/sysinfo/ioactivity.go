package sysinfo

import (
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/dock"     // Connection to cairo-dock.
	"github.com/sqp/godock/libs/log"      // Display info in terminal.
	"github.com/sqp/godock/libs/packages" // ByteSize.

	// "fmt"
)

//
var DockGraphType = []string{"Line", "Plain", "Bar", "Circle", "Plain Circle"}

type FormatIO func(device string, in, out uint64) string

// Create a new data poller for disk activity monitoring.
//
type IOActivity struct {
	list     map[string]*stat
	interval uint64
	info     ITextInfo         // Paired values text renderer.
	app      dock.RenderSimple // Controler to the Cairo-Dock icon.

	FormatIcon  FormatIO
	FormatLabel FormatIO
	GetData     func() ([]value, error)
}

func NewIOActivity(app dock.RenderSimple) *IOActivity {
	return &IOActivity{
		list: make(map[string]*stat),
		app:  app,
	}
}

func (na *IOActivity) Settings(interval uint64, textPosition cdtype.InfoPosition, renderer, graphType int, gaugeTheme string, names ...string) {
	na.interval = interval

	na.list = make(map[string]*stat) // Clear list. Nothing must remain.
	na.app.AddDataRenderer("", 0, "")

	if len(names) > 0 {
		for _, name := range names {
			na.list[name] = &stat{}
		}

		switch textPosition { // Add text renderer info.
		case cdtype.InfoOnIcon:
			na.info = NewTextIcon(na.app)
			na.info.SetSeparator("\n")
			na.info.SetCallAppend(na.FormatIcon)
			na.info.SetCallFail(func(string) string { return "N/A" }) // NEED TRANSLATE GETTEXT

		case cdtype.InfoOnLabel:
			na.info = NewTextLabel(na.app)
			na.info.SetSeparator("\n")
			na.info.SetCallAppend(na.FormatLabel)
			na.info.SetCallFail(func(dev string) string { return dev + ": " + "N/A" }) // NEED TRANSLATE GETTEXT
			// na.info.SetCallFail(func(dev string) string { return fmt.Sprintf("%s: %s", dev, "N/A") }) // NEED TRANSLATE GETTEXT

		default:
			na.info = NewTextNil()
		}

		switch renderer {
		case 0:
			na.app.AddDataRenderer("gauge", 2*int32(len(na.list)), gaugeTheme)
		case 1:
			na.app.AddDataRenderer("graph", 2*int32(len(na.list)), DockGraphType[graphType])
		}
	} else {
		log.DEV("no na ffs")
		na.app.SetLabel("No na defined.")
	}
}

//
//-------------------------------------------------------------[ UPDATE DATA ]--

// Get and display activity information for configured network interfaces.
// Display on the Cairo-Dock icon:
//   RenderValues: gauge or graph
//   RenderText: quickinfo or label
//
func (na *IOActivity) Check() {
	na.Get()

	if len(na.list) == 0 {
		return
	}

	na.info.Clear()
	var values []float64

	for name, stat := range na.list {
		if in, out, ok := stat.Current(); ok {
			na.info.Append(name, stat.rateReadNow, stat.rateWriteNow)
			values = append(values, in, out)
		} else {
			na.info.Fail(name)
			values = append(values, 0, 0)
		}
	}

	na.info.Display()

	if len(values) > 0 {
		na.app.RenderValues(values...)
	}
}

func (net *IOActivity) Get() {
	// if len(net.list) == 0 {
	// 	return
	// }

	for _, stat := range net.list { // Reset our acquisition status for every field.
		stat.acquisitionOK = false
	}

	l, e := net.GetData()
	if log.Err(e, "get data") {
		return
	}

	for _, newv := range l {
		if stat, ok := net.list[newv.Field]; ok {
			stat.Set(newv.In, newv.Out, net.interval)
		} else {
			log.DEV("unknown", newv.Field)
		}
	}
}

//
//-----------------------------------------------------[ TEXT INFO CALLBACKS ]--

// Quick-info display callback. One line for each value. Zero are replaced by empty string.
//
func FormatIcon(dev string, in, out uint64) string {
	return FormatRate(in) + "\n" + FormatRate(out)
}

func FormatRate(size uint64) string {
	if size > 0 {
		return packages.ByteSize(size).String()
	}
	return ""
}

//
//-----------------------------------------------------[ TEXT INFO RENDERERS ]--

// ITextInfo is the interface for a paired value text renderer. Used with ....
//
type ITextInfo interface {
	Display()
	Clear()
	SetSeparator(sep string)

	Append(dev string, in, out uint64)
	SetCallAppend(call FormatIO)
	Fail(dev string)
	SetCallFail(call func(dev string) string)
}

type TextInfo struct {
	info       string
	sep        string
	callAppend FormatIO
	callFail   func(dev string) string
}

func (ti *TextInfo) Append(dev string, in, out uint64) {
	if ti.info != "" {
		ti.info += ti.sep
	}
	ti.info += ti.callAppend(dev, in, out)
}

func (ti *TextInfo) Fail(dev string) {
	if ti.info != "" {
		ti.info += ti.sep
	}
	ti.info += ti.callFail(dev)
}

func (ti *TextInfo) Clear() {
	ti.info = ""
}

func (ti *TextInfo) SetSeparator(sep string) {
	ti.sep = sep
}

func (ti *TextInfo) SetCallAppend(call FormatIO) {
	ti.callAppend = call
}

func (ti *TextInfo) SetCallFail(call func(dev string) string) {
	ti.callFail = call
}

type TextIcon struct {
	app dock.RenderSimple // Controler to the Cairo-Dock icon.
	TextInfo
}

func NewTextIcon(app dock.RenderSimple) *TextIcon {
	return &TextIcon{app: app}
}

func (ti *TextIcon) Display() {
	ti.app.SetQuickInfo(ti.info)
}

type TextLabel struct {
	app dock.RenderSimple // Controler to the Cairo-Dock icon.
	TextInfo
}

func NewTextLabel(app dock.RenderSimple) *TextLabel {
	return &TextLabel{app: app}
}

func (ti *TextLabel) Display() {
	ti.app.SetLabel(ti.info)
}

type TextNil struct {
	TextInfo
}

func NewTextNil() *TextNil {
	t := &TextNil{}
	t.callAppend = func(dev string, in, out uint64) string { return "" }
	t.callFail = func(dev string) string { return "" }
	return t
}

func (ti *TextNil) Display() {}
