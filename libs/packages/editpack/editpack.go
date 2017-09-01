// Package editpack builds a console interface to edit applets definition info.
package editpack

import (
	"github.com/gizak/termui"

	"github.com/sqp/godock/libs/cdtype" // Logger type.
	"github.com/sqp/godock/libs/packages"
	"github.com/sqp/godock/libs/ternary"

	"errors"
)

const infoText = "[PgUp](fg-bold)/[PgDn](fg-bold):Applet" +
	" - [Up](fg-bold)/[Down](fg-bold):Field" +
	" - [Left](fg-bold)/[Right](fg-bold):Value" +
	" - [F7](fg-bold):Save - [F10](fg-bold):Quit"

const lockedFmt = `Path: %s
Size: %s
Type: %s`

const (
	firstField     = 1 // field 0 is a blank line.
	colorSelected  = "fg-white,bg-blue"
	colorEdited    = "fg-green"
	colorEditMinor = "fg-magenta"
	colorEditMajor = "fg-red"
)

var fields = []packages.AppInfoField{
	packages.AppInfoUnknown,
	packages.AppInfoVersion,
	packages.AppInfoCategory,
	packages.AppInfoAuthor,
	packages.AppInfoActAsLauncher,
	packages.AppInfoMultiInstance,
	packages.AppInfoTitle,
	packages.AppInfoIcon,
}

func findField(field packages.AppInfoField) int {
	for i, test := range fields {
		if field == test {
			return i
		}
	}
	return 0
}

// PacksExternal lists applets packages in the given dir.
//
func PacksExternal(log cdtype.Logger, dir string) (packages.AppletPackages, error) {
	packs, e := packages.ListFromDir(log, dir, cdtype.PackTypeUser, packages.SourceApplet)
	if e != nil {
		return nil, e
	}
	if len(packs) == 0 {
		return nil, errors.New("no packages found in " + dir)
	}
	return packs, nil
}

// Start starts the edit applet console GUI.
//
func Start(log cdtype.Logger, packs packages.AppletPackages) error {
	e := termui.Init()
	if e != nil {
		return e
	}
	defer termui.Close()

	// Get applet names, and align text.
	txtApplets := make([]string, len(packs))
	for i, pack := range packs {
		txtApplets[i] = pack.DisplayedName
	}
	txtApplets = expandTexts(txtApplets)

	// Get translated fields, and align text.
	txtFields := make([]string, len(fields))
	for i, field := range fields[firstField:] {
		txtFields[i+firstField] = field.Translated()
	}
	txtFields = expandTexts(txtFields)

	// Applet info.
	var (
		ed       = New(log, packs, len(txtApplets[0]), len(txtFields[firstField]))
		selected = ed.SetField(txtFields, 0)
		appID    = ed.SetPack(txtApplets, 0)
	)

	// calculate layout and render.
	ed.resize()

	termui.Handle("/sys/wnd/resize", func(termui.Event) { ed.resize() })

	termui.Handle("/sys/kbd", func(evt termui.Event) {
		key := evt.Data.(termui.EvtKbd).KeyStr
		switch key {
		case "<f10>", "C-c", "C-q":
			termui.StopLoop()

		case "<f7>":
			e := ed.save(appID)
			if e == nil {
				// Refresh data to clear updated color.
				ed.SetPack(txtApplets, appID)
				termui.Render(ed.appinfo, ed.locked, ed.desc)

			} else {
				ed.locked.Text = "SAVE FAILED: " + e.Error()
				termui.Render(ed.locked)
			}

		case "<up>", "<down>":
			delta := ternary.Int(key == "<down>", 1, -1)
			selected = ed.SetField(txtFields, selected+delta)
			termui.Render(ed.fields, ed.locked, ed.desc)

		case "<previous>", "<next>": // PgDn, PgUp
			ed.edits = nil
			delta := ternary.Int(key == "<next>", 1, -1)
			appID = ed.SetPack(txtApplets, appID+delta)
			termui.Render(ed.applets, ed.appinfo, ed.locked, ed.desc, ed.title)

		case "<left>", "<right>":
			delta := ternary.Int(key == "<right>", 1, -1)
			ed.editValue(appID, selected, delta)
			termui.Render(ed.applets, ed.appinfo, ed.locked, ed.desc, ed.title)

			// default:
			// 	ed.title.BorderLabel = key
			// 	termui.Render(ed.title)
		}

	})

	termui.Loop()
	return nil
}

//
//--------------------------------------------------------------[ SET VALUES ]--

func (ed *Editor) editValue(appID, field, delta int) {
	if ed.edits == nil {
		ed.edits = make(map[packages.AppInfoField]interface{})
	}
	f := fields[field]
	switch f {
	case packages.AppInfoVersion:
		ed.setVersion(delta)

	case packages.AppInfoCategory:
		ed.setInt(f, delta, int(ed.packs[appID].Category), 0, 7)

	case packages.AppInfoActAsLauncher:
		ed.setBool(f, ed.packs[appID].ActAsLauncher)

	case packages.AppInfoMultiInstance:
		ed.setBool(f, ed.packs[appID].IsMultiInstance)
	}

	// Apply changes.
	ed.showPack(appID)
}

func (ed *Editor) setBool(field packages.AppInfoField, def bool) {
	_, ok := ed.edits[field]
	if ok {
		delete(ed.edits, field)
	} else {
		ed.edits[field] = !def
	}
}

func (ed *Editor) setInt(field packages.AppInfoField, delta, def, min, max int) {
	old, ok := ed.edits[field]
	val := 0
	if ok {
		val = old.(int) + delta
	} else {
		val = def + delta
	}
	val = ternary.Max(min, ternary.Min(val, max))

	if ok && val == def {
		delete(ed.edits, field)
	}
	if val != def {
		ed.edits[field] = val
	}
}

func (ed *Editor) setVersion(delta int) {
	var val int
	old, ok := ed.edits[packages.AppInfoVersion]
	switch {
	case ok:
		val = old.(int) + delta
		if val <= 0 {
			delete(ed.edits, packages.AppInfoVersion)
			return
		}

	case delta < 1:
		return

	default:
		val = delta
	}
	ed.edits[packages.AppInfoVersion] = ternary.Min(val, 3)
}

//
//--------------------------------------------------------------------[ SAVE ]--

func (ed *Editor) save(appID int) error {
	pack := ed.packs[appID]
	if pack == nil {
		return nil
	}

	newpack, e := pack.SaveUpdated(ed.edits)
	if e != nil {
		return e
	}
	ed.packs[appID] = newpack
	ed.edits = nil
	return nil
}
