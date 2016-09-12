package editpack

import (
	"github.com/gizak/termui"

	"github.com/sqp/godock/libs/cdtype" // Logger type.
	"github.com/sqp/godock/libs/packages"
	"github.com/sqp/godock/libs/ternary"

	"strconv"
	"strings"
)

//
//------------------------------------------------------------------[ EDITOR ]--

// Editor defines the edit info console GUI.
//
type Editor struct {
	applets *termui.List
	fields  *termui.List
	appinfo *termui.List
	locked  *termui.Par
	desc    *termui.Par
	title   *termui.Par
	packs   packages.AppletPackages
	edits   map[int]interface{}
	log     cdtype.Logger
}

// New creates a new edit info console GUI.
//
func New(log cdtype.Logger, packs packages.AppletPackages) *Editor {
	ed := &Editor{
		applets: termui.NewList(),
		fields:  termui.NewList(),
		appinfo: termui.NewList(),
		locked:  termui.NewPar(""),
		desc:    termui.NewPar(""),
		title:   termui.NewPar(infoText),
		packs:   packs,
		log:     log,
	}

	ed.applets.ItemFgColor = termui.ColorYellow
	ed.applets.BorderLabel = "[ Applets ]"
	ed.applets.Width = 20 // TODO autodetect.
	ed.applets.BorderBottom = false
	ed.applets.BorderFg = termui.ColorCyan

	ed.fields.BorderLabel = "[ Fields ]"
	ed.fields.X = ed.applets.Width
	ed.fields.Height = 6 + 2 + 2 // last 2 for blank lines around text
	ed.fields.Width = len(fields[1]) + 2
	ed.fields.BorderLeft = false
	ed.fields.BorderBottom = false
	ed.fields.BorderFg = termui.ColorCyan

	ed.appinfo.BorderLabel = "[ Value ]"
	ed.appinfo.X = ed.fields.X + ed.fields.Width
	ed.appinfo.Height = 6 + 2 + 2 // last 2 for blank lines around text
	ed.appinfo.BorderLeft = false
	ed.appinfo.BorderRight = false
	ed.appinfo.BorderBottom = false
	ed.appinfo.BorderFg = termui.ColorCyan

	ed.locked.BorderLabel = "[ Details ]"
	ed.locked.X = ed.applets.Width
	ed.locked.Y = ed.fields.Height - 1 // offset 1 for border
	ed.locked.Height = 6
	ed.locked.BorderBottom = false
	ed.locked.BorderLeft = false
	ed.locked.BorderRight = false
	ed.locked.BorderFg = termui.ColorCyan

	ed.desc.BorderLabel = "[ Description ]"
	ed.desc.X = ed.applets.Width
	ed.desc.BorderBottom = false
	ed.desc.BorderLeft = false
	ed.desc.BorderRight = false
	ed.desc.BorderFg = termui.ColorCyan

	ed.title.BorderLabel = "[ Edit applet info ]"
	ed.title.Height = 2
	ed.title.TextFgColor = termui.ColorWhite
	ed.title.BorderBottom = false
	ed.title.BorderLeft = false
	ed.title.BorderRight = false
	ed.title.BorderFg = termui.ColorCyan

	return ed
}

func (ed *Editor) resize() {
	termui.Body.Width = termui.TermWidth()

	ed.applets.Height = termui.TermHeight() - 2 + 1 // offset 1 for border
	ed.desc.Y = ed.locked.Y + ed.locked.Height - 1  // offset 1 for border
	ed.desc.Height = termui.TermHeight() - ed.desc.Y - 2
	ed.title.Y = termui.TermHeight() - 2

	ed.title.Width = termui.TermWidth() + 1                     // offset 1 for border
	ed.appinfo.Width = termui.TermWidth() - ed.appinfo.X + 1    // offset 1 for border
	ed.locked.Width = termui.TermWidth() - ed.applets.Width + 1 // offset 1 for border
	ed.desc.Width = termui.TermWidth() - ed.applets.Width + 1   // offset 1 for border
	ed.desc.WrapLength = ed.desc.Width

	termui.Render(ed.applets, ed.fields, ed.appinfo, ed.locked, ed.desc, ed.title)
}

// SetField sets the current field number.
//
func (ed *Editor) SetField(id int) int {
	ed.fields.Items = make([]string, len(fields))
	for i, str := range fields {
		ed.fields.Items[i] = str
	}
	// Highlight selected.
	id = ternary.Max(1, ternary.Min(id, len(ed.fields.Items)-1))
	ed.fields.Items[id] = "[" + ed.fields.Items[id] + "](fg-white,bg-blue)"
	return id
}

// SetPack sets the current package number.
//
func (ed *Editor) SetPack(appID int) int {
	appID = ternary.Max(0, ternary.Min(appID, len(ed.packs)-1))

	var names []string
	for _, pack := range ed.packs {
		names = append(names, pack.DisplayedName)
	}
	names[appID] = "[" + names[appID] + "](fg-white,bg-blue)"

	ed.applets.Items = names

	ed.showPack(appID)
	return appID
}

//
//-----------------------------------------------------------------[ DISPLAY ]--

func (ed *Editor) showPack(appID int) {
	appID = ternary.Max(0, ternary.Min(appID, len(ed.packs)-1))
	pack := ed.packs[appID]
	catstr, _ := packages.FormatCategory(pack.Category)

	ed.appinfo.BorderLabel = pack.Path
	ed.appinfo.Items = []string{
		"",
		pack.DisplayedName,
		pack.Author,
		pack.Version,
		strconv.Itoa(pack.Category) + " : " + catstr,
		formatBool(pack.ActAsLauncher),
		formatBool(pack.IsMultiInstance),
	}

	ed.locked.Text = "this\nis\ndetails"

	ed.desc.Text = strings.Replace(pack.Description, "\\n", "\n", -1)

	// Update edited fields.
	for k, v := range ed.edits {
		switch k {
		case fieldVersion:
			ed.appinfo.Items[k], _ = packages.FormatNewVersion(pack.Version, v.(int))
			switch v.(int) {
			case 2:
				ed.appinfo.Items[k] = "[" + ed.appinfo.Items[k] + "](fg-magenta)"
				continue

			case 3:
				ed.appinfo.Items[k] = "[" + ed.appinfo.Items[k] + "](fg-red)"
				continue
			}

		case fieldCategory:
			catstr, _ := packages.FormatCategory(v.(int))
			ed.appinfo.Items[k] = strconv.Itoa(v.(int)) + " : " + catstr

		case fieldActAsLauncher, fieldMultiInstance:
			ed.appinfo.Items[k] = formatBool(v.(bool))
		}

		// Add color to field to indicate it's modified.
		ed.appinfo.Items[k] = "[" + ed.appinfo.Items[k] + "](fg-green)"
	}
}

func formatBool(val bool) string {
	return ternary.String(val, "Yes", "No")
}
