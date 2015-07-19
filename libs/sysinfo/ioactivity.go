package sysinfo

import (
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/text/bytesize"

	"fmt"
)

// percent provides a couple of text and value to render.
//
type percent struct {
	text  string
	value float64
}

// RenderPercent provides a simple icon/label text renderer with values in percent.
//
type RenderPercent struct {
	text   RenderOne
	values []percent
	App    cdtype.RenderSimple
	Texts  map[cdtype.InfoPosition]RenderOne
	cbText func(string) error // Display callback.

	DisplayText   cdtype.InfoPosition
	DisplayValues int
	GaugeTheme    string
	GraphType     cdtype.RendererGraphType
}

// Settings is a all in one method to apply applet settings.
//
func (rp *RenderPercent) Settings(textPosition cdtype.InfoPosition, renderer int, graphType cdtype.RendererGraphType, gaugeTheme string) {
	rp.DisplayValues = renderer
	rp.GraphType = graphType
	rp.GaugeTheme = gaugeTheme

	rp.text = rp.Texts[textPosition]

	switch textPosition { // Add text renderer info.
	case cdtype.InfoOnIcon:
		rp.cbText = rp.App.SetQuickInfo

	case cdtype.InfoOnLabel:
		rp.cbText = rp.App.SetLabel

	default:
		rp.cbText = nilText
	}
}

// SetSize sets the number of values to render on icon.
// Mandatory with and after Settings.
//
func (rp *RenderPercent) SetSize(size int) {
	rp.App.DataRenderer().Remove()

	switch {
	case rp.DisplayValues == 0:
		rp.App.DataRenderer().Gauge(size, rp.GaugeTheme)

	case rp.DisplayValues == 1:
		rp.App.DataRenderer().Graph(size, rp.GraphType)
	}
}

// Append adds a value to the renderer.
//
func (rp *RenderPercent) Append(str string, value float64) {
	rp.values = append(rp.values, percent{text: str, value: value})
}

// Display renders and displays the provided values.
//
func (rp *RenderPercent) Display() {
	if len(rp.values) == 0 {
		return
	}

	var values []float64
	rp.text.Clear()
	for _, v := range rp.values {
		values = append(values, v.value)
		rp.text.Append(v.text, v.value*100)
	}
	rp.App.DataRenderer().Render(values...)
	rp.cbText(rp.text.Text())
}

// Clear resets the internal data.
//
func (rp *RenderPercent) Clear() {
	rp.text.Clear()
	rp.values = nil
}

// stub for text rendering.
func nilText(string) error { return nil }

//
//-----------------------------------------------------------[ SINGLE VALUES ]--

// RenderOne provides a simple icon/label text renderer with values in percent.
//
type RenderOne struct {
	info     string // info to display.
	Sep      string // separator between lines.
	ShowPre  bool   // text pre value.
	ShowPost bool   // text post value.
}

// Append adds new data values to the renderer.
// The value must be in the 0..100 range.
//
func (ro *RenderOne) Append(str string, value float64) {
	if ro.info != "" {
		ro.info += ro.Sep
	}
	if ro.ShowPre && str != "" {
		ro.info += str + " "
	}

	ro.info += formatPercent(value)

	if ro.ShowPost && str != "" {
		ro.info += " " + str
	}
}

// Text returns the text to display.
//
func (ro *RenderOne) Text() string {
	return ro.info
}

// Clear resets the internal text.
//
func (ro *RenderOne) Clear() {
	ro.info = ""
}

// format value as percent.
//
func formatPercent(value float64) string {
	format := ""
	switch {
	case value < 0:
		return "N/A"
	case value < 10:
		format = "%1.1f%%"
	default:
		format = "%2.f%%"
	}
	return fmt.Sprintf(format, value)
}

//
//----------------------------------------------------------[ COUPLED VALUES ]--

// FormatIO is a text format method for IOActivity.
//
type FormatIO func(device string, in, out uint64) string

// IOActivity extract delta IO stats from stacking system counters.
//
type IOActivity struct {
	Log cdtype.Logger

	list     []*stat
	interval uint64
	info     ITextInfo           // Paired values text renderer.
	app      cdtype.RenderSimple // Controler to the Cairo-Dock icon.

	FormatIcon  FormatIO
	FormatLabel FormatIO
	GetData     func() ([]Value, error)
}

// NewIOActivity create a new data store for IO activity monitoring.
//
func NewIOActivity(app cdtype.RenderSimple) *IOActivity {
	return &IOActivity{
		app: app,
	}
}

// Settings is a all in one method to configure your IOActivity.
//
func (ioa *IOActivity) Settings(interval uint64, textPosition cdtype.InfoPosition, renderer int, graphType cdtype.RendererGraphType, gaugeTheme string, names ...string) {
	ioa.interval = interval

	ioa.list = []*stat{} // Clear list. Nothing must remain.
	ioa.app.DataRenderer().Remove()

	if len(names) > 0 {
		for _, name := range names {
			ioa.list = append(ioa.list, &stat{name: name})
		}

		switch textPosition { // Add text renderer info.
		case cdtype.InfoOnIcon:
			ioa.info = NewTextIcon(ioa.app)
			ioa.info.SetSeparator("\n")
			ioa.info.SetCallAppend(ioa.FormatIcon)
			ioa.info.SetCallFail(func(string) string { return "N/A" }) // NEED TRANSLATE GETTEXT

		case cdtype.InfoOnLabel:
			ioa.info = NewTextLabel(ioa.app)
			ioa.info.SetSeparator("\n")
			ioa.info.SetCallAppend(ioa.FormatLabel)
			ioa.info.SetCallFail(func(dev string) string { return dev + ": " + "N/A" }) // NEED TRANSLATE GETTEXT
			// ioa.info.SetCallFail(func(dev string) string { return fmt.Sprintf("%s: %s", dev, "N/A") }) // NEED TRANSLATE GETTEXT

		default:
			ioa.info = NewTextNil()
		}

		switch renderer {
		case 0:
			ioa.app.DataRenderer().Gauge(2*len(ioa.list), gaugeTheme)
		case 1:
			ioa.app.DataRenderer().Graph(2*len(ioa.list), graphType)
		}
	} else {
		// log.DEV("no na ffs")
		ioa.app.SetLabel("No device defined.")
	}
}

//
//-------------------------------------------------------------[ UPDATE DATA ]--

// Check pull and display activity information for configured interfaces.
// Display on the Cairo-Dock icon:
//   RenderValues: gauge or graph
//   RenderText: quickinfo or label
//
func (ioa *IOActivity) Check() {
	ioa.Get()

	if len(ioa.list) == 0 {
		return
	}

	ioa.info.Clear()
	var values []float64

	for _, stat := range ioa.list {
		if in, out, ok := stat.Current(); ok {
			ioa.info.Append(stat.name, stat.rateReadNow, stat.rateWriteNow)
			values = append(values, in, out)
		} else {
			ioa.info.Fail(stat.name)
			values = append(values, 0, 0)
		}
	}

	ioa.info.Display()

	if len(values) > 0 {
		ioa.app.DataRenderer().Render(values...)
	}
}

// Get new data from source.
//
func (ioa *IOActivity) Get() {
	// if len(ioa.list) == 0 {
	// 	return
	// }

	for _, stat := range ioa.list { // Reset our acquisition status for every field.
		stat.acquisitionOK = false
	}

	l, e := ioa.GetData()
	if ioa.Log.Err(e, "get data") {
		return
	}

	for _, newv := range l {
		if st := ioa.find(newv.Field); st != nil {
			st.Set(newv.In, newv.Out, ioa.interval)
		} else {
			// log.DEV("unknown", newv.Field)
		}
	}
}

// find returns the stat matching the given reference.
//
func (ioa *IOActivity) find(name string) *stat {
	for _, st := range ioa.list {
		if st.name == name {
			return st
		}
	}
	return nil
}

//
//-----------------------------------------------------[ TEXT INFO CALLBACKS ]--

// FormatIcon is a Quick-info display callback. One line for each value.
// Zero are replaced by empty string.
//
func FormatIcon(dev string, in, out uint64) string {
	return FormatRate(in) + "\n" + FormatRate(out)
}

// FormatRate format value to string, or nothing if 0.
//
func FormatRate(size uint64) string {
	if size > 0 {
		return bytesize.ByteSize(size).String()
	}
	return ""
}

//
//--------------------------------------------------------[ TEXT INFO COMMON ]--

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

// TextInfo defines the base data for a paired value text renderer.
//
type TextInfo struct {
	info       string
	sep        string
	callAppend FormatIO
	callFail   func(dev string) string
}

// Append adds new data values to the renderer.
//
func (ti *TextInfo) Append(dev string, in, out uint64) {
	if ti.info != "" {
		ti.info += ti.sep
	}
	ti.info += ti.callAppend(dev, in, out)
}

// Fail adds a data error to the renderer.
//
func (ti *TextInfo) Fail(dev string) {
	if ti.info != "" {
		ti.info += ti.sep
	}
	ti.info += ti.callFail(dev)
}

// Clear resets the internal text.
//
func (ti *TextInfo) Clear() {
	ti.info = ""
}

// SetSeparator declares the text separator to add between text values.
//
func (ti *TextInfo) SetSeparator(sep string) {
	ti.sep = sep
}

// SetCallAppend declares the text formatter callback for each value.
//
func (ti *TextInfo) SetCallAppend(call FormatIO) {
	ti.callAppend = call
}

// SetCallFail declares the error formatter callback.
//
func (ti *TextInfo) SetCallFail(call func(dev string) string) {
	ti.callFail = call
}

//
//-----------------------------------------------------[ TEXT INFO RENDERERS ]--

// TextIcon renders a paired value text on icon quickinfo.
//
type TextIcon struct {
	app cdtype.RenderSimple // Controler to the Cairo-Dock icon.
	TextInfo
}

// NewTextIcon creates a new paired value text renderer on icon quickinfo.
//
func NewTextIcon(app cdtype.RenderSimple) *TextIcon {
	return &TextIcon{app: app}
}

// Display renders data on icon quickinfo.
//
func (ti *TextIcon) Display() {
	ti.app.SetQuickInfo(ti.info)
}

// TextLabel renders a paired value text on icon label.
//
type TextLabel struct {
	app cdtype.RenderSimple // Controler to the Cairo-Dock icon.
	TextInfo
}

// NewTextLabel creates a new paired value text renderer on icon label.
//
func NewTextLabel(app cdtype.RenderSimple) *TextLabel {
	return &TextLabel{app: app}
}

// Display renders data on icon label.
//
func (ti *TextLabel) Display() {
	ti.app.SetLabel(ti.info)
}

// TextNil provides a dumb interface compatible with a paired value text renderer.
//
type TextNil struct {
	TextInfo
}

// NewTextNil creates a dumb interface compatible with paired value text renderer.
//
func NewTextNil() *TextNil {
	t := &TextNil{}
	t.callAppend = func(dev string, in, out uint64) string { return "" }
	t.callFail = func(dev string) string { return "" }
	return t
}

// Display will do nothing on the nil renderer.
//
func (ti *TextNil) Display() {}
