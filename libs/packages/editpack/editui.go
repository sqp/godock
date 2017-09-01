package editpack

import (
	"github.com/gizak/termui"

	"github.com/sqp/godock/libs/cdtype" // Logger type.
	"github.com/sqp/godock/libs/packages"
	"github.com/sqp/godock/libs/ternary"
	"github.com/sqp/godock/libs/text/tran"

	"fmt"
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
	edits   map[packages.AppInfoField]interface{}
	log     cdtype.Logger
}

// New creates a new edit info console GUI.
//
func New(log cdtype.Logger, packs packages.AppletPackages, appwidth, fieldswidth int) *Editor {
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

	ed.applets.BorderLabel = formatTitle(tran.Slate("Applets"))
	ed.applets.Width = appwidth + 3 // offset 1 for text and 2 for border
	ed.applets.BorderBottom = false
	ed.applets.BorderFg = termui.ColorCyan
	ed.applets.ItemFgColor = termui.ColorYellow

	ed.fields.BorderLabel = formatTitle(tran.Slate("Fields"))
	ed.fields.X = ed.applets.Width
	ed.fields.Width = fieldswidth + 2      // offset 1 for text and 1 for border
	ed.fields.Height = len(fields) + 1 + 2 // offset 1 for the last blank line and 2 for borders.
	ed.fields.BorderLeft = false
	ed.fields.BorderBottom = false
	ed.fields.BorderFg = termui.ColorCyan

	ed.appinfo.BorderLabel = formatTitle(tran.Slate("Values"))
	ed.appinfo.X = ed.fields.X + ed.fields.Width
	ed.appinfo.Height = ed.fields.Height
	ed.appinfo.BorderLeft = false
	ed.appinfo.BorderRight = false
	ed.appinfo.BorderBottom = false
	ed.appinfo.BorderFg = termui.ColorCyan

	ed.locked.BorderLabel = formatTitle(tran.Slate("Details"))
	ed.locked.X = ed.applets.Width
	ed.locked.Y = ed.fields.Height - 1 // offset 1 for border
	ed.locked.Height = 6
	ed.locked.BorderBottom = false
	ed.locked.BorderLeft = false
	ed.locked.BorderRight = false
	ed.locked.BorderFg = termui.ColorCyan

	ed.desc.BorderLabel = formatTitle(tran.Slate("Description"))
	ed.desc.X = ed.applets.Width
	ed.desc.BorderBottom = false
	ed.desc.BorderLeft = false
	ed.desc.BorderRight = false
	ed.desc.BorderFg = termui.ColorCyan

	ed.title.BorderLabel = formatTitle(tran.Slate("Edit applet info"))
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
func (ed *Editor) SetField(txtFields []string, id int) int {
	// Ensure selected item is in range.
	id = ternary.Max(firstField, ternary.Min(id, len(ed.fields.Items)-1))

	// Copy fields list.
	ed.fields.Items = make([]string, len(txtFields))
	copy(ed.fields.Items, txtFields)

	// Highlight selected.
	ed.fields.Items[id] = colored(colorSelected, ed.fields.Items[id])
	return id
}

// SetPack sets the current package number.
//
func (ed *Editor) SetPack(appnames []string, appID int) int {
	// Ensure selected item is in range.
	appID = ternary.Max(0, ternary.Min(appID, len(ed.packs)-1))

	// Copy applets names list.
	ed.applets.Items = make([]string, len(appnames))
	copy(ed.applets.Items, appnames)

	// Show color on selected and display.
	ed.applets.Items[appID] = colored(colorSelected, ed.applets.Items[appID])
	ed.showPack(appID)
	return appID
}

//
//-----------------------------------------------------------------[ DISPLAY ]--

func (ed *Editor) showPack(appID int) {
	appID = ternary.Max(0, ternary.Min(appID, len(ed.packs)-1))
	pack := ed.packs[appID]

	ed.appinfo.Items = []string{
		"",
		pack.Version,
		strconv.Itoa(int(pack.Category)) + " : " + pack.Category.Translated(),
		pack.Author,
		formatBool(pack.ActAsLauncher),
		formatBool(pack.IsMultiInstance),
		pack.Title,
		pack.Icon,
	}

	ed.locked.Text = fmt.Sprintf(lockedFmt, pack.Path, pack.FormatSize(), pack.Type)

	ed.desc.Text = strings.Replace(pack.Description, "\\n", "\n", -1)

	// Update edited fields.
	for k, v := range ed.edits {
		line := findField(k)
		color := colorEdited
		switch k {
		case packages.AppInfoVersion:
			ed.appinfo.Items[line], _ = packages.FormatNewVersion(pack.Version, v.(int))
			switch v.(int) { // Other colors when upgrading minor and major.
			case 2:
				color = colorEditMinor

			case 3:
				color = colorEditMajor
			}

		case packages.AppInfoCategory:
			cat := cdtype.CategoryType(v.(int)).String()
			ed.appinfo.Items[line] = strconv.Itoa(v.(int)) + " : " + cat

		case packages.AppInfoActAsLauncher, packages.AppInfoMultiInstance:
			ed.appinfo.Items[line] = formatBool(v.(bool))
		}

		// Add color to field to indicate it's modified.
		ed.appinfo.Items[line] = colored(color, ed.appinfo.Items[line])
	}
}

func colored(color, text string) string { return "[" + text + "](" + color + ")" }
func formatBool(val bool) string        { return ternary.String(val, "Yes", "No") }
func formatTitle(text string) string    { return "[ " + text + " ]" }

// expandTexts finds the largest text in the list, and expand them all to that max size.
//
func expandTexts(texts []string) []string {
	max := 0
	for _, str := range texts {
		max = ternary.Max(max, len(str))
	}
	format := fmt.Sprintf(" %%-%ds ", max)
	for i, str := range texts {
		texts[i] = fmt.Sprintf(format, str)
	}
	return texts
}
