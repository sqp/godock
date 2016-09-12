// Package editpack builds a console interface to edit applets definition info.
package editpack

import (
	"github.com/gizak/termui"

	"github.com/sqp/godock/libs/cdtype" // Logger type.
	"github.com/sqp/godock/libs/config" // Config parser.
	"github.com/sqp/godock/libs/packages"
	"github.com/sqp/godock/libs/ternary"

	"errors"
	"path/filepath"
)

const infoText = "[PgUp](fg-bold)/[PgDn](fg-bold):Applet" +
	" - [Up](fg-bold)/[Down](fg-bold):Field" +
	" - [Left](fg-bold)/[Right](fg-bold):Value" +
	" - [F7](fg-bold):Save - [F10](fg-bold):Quit"

var fields = []string{
	"",
	"Name           ",
	"Author         ",
	"Version        ",
	"Category       ",
	"Act as launcher",
	"Multi instance ",
}

var locked = []string{
	"Path  ",
	"Size  ",
	"Type  ",
}

const (
	fieldBlank = iota
	fieldName
	fieldAuthor
	fieldVersion
	fieldCategory
	fieldActAsLauncher
	fieldMultiInstance
)

// PacksExternal lists applets packages in the given dir.
//
func PacksExternal(log cdtype.Logger, dir string) (packages.AppletPackages, error) {
	packs, e := packages.ListFromDir(log, dir, packages.TypeUser, packages.SourceApplet)
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

	// Applet info.
	ed := New(log, packs)

	selected := fieldVersion
	appID := 0

	ed.SetField(selected)
	ed.SetPack(appID)

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
				ed.SetPack(appID)
				termui.Render(ed.appinfo, ed.locked, ed.desc)

			} else {
				ed.locked.Text = "SAVE FAILED: " + e.Error()
			}

		case "<up>", "<down>":
			delta := ternary.Int(key == "<down>", 1, -1)
			selected = ed.SetField(selected + delta)
			termui.Render(ed.fields, ed.locked, ed.desc)

		case "<previous>", "<next>":
			ed.edits = nil
			delta := ternary.Int(key == "<next>", 1, -1)
			appID = ed.SetPack(appID + delta)
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
		ed.edits = make(map[int]interface{})
	}
	switch field {
	case fieldVersion:
		ed.setVersion(delta)

	case fieldCategory:
		ed.setInt(field, delta, ed.packs[appID].Category, 0, 7)

	case fieldActAsLauncher:
		ed.setBool(field, ed.packs[appID].ActAsLauncher)

	case fieldMultiInstance:
		ed.setBool(field, ed.packs[appID].IsMultiInstance)
	}

	// Apply changes.
	ed.showPack(appID)
}

func (ed *Editor) setBool(field int, def bool) {
	_, ok := ed.edits[field]
	if ok {
		delete(ed.edits, field)
	} else {
		ed.edits[field] = !def
	}
}

func (ed *Editor) setInt(field, delta, def, min, max int) {
	val := -1
	old, ok := ed.edits[field]
	if ok {
		val = old.(int) + delta
		if val == def {
			delete(ed.edits, field)
			return
		}
	} else {
		val = def + delta
	}
	ed.edits[field] = ternary.Max(min, ternary.Min(val, max))
}

func (ed *Editor) setVersion(delta int) {
	var val int
	old, ok := ed.edits[fieldVersion]
	switch {
	case ok:
		val = old.(int) + delta
		if val <= 0 {
			delete(ed.edits, fieldVersion)
			return
		}

	case delta < 1:
		return

	default:
		val = delta
	}
	ed.edits[fieldVersion] = ternary.Min(val, 3)
}

//
//--------------------------------------------------------------------[ SAVE ]--

func (ed *Editor) save(appID int) error {
	pack := ed.packs[appID]

	if len(ed.edits) == 0 || pack == nil {
		return nil
	}

	group := pack.Source.Group()
	newver := ""

	// Update version first in applet config file.
	// When an upgrade is requested, we have to change versions in both files,
	// so the config upgrade becomes mandatory.
	if delta, upVer := ed.edits[fieldVersion]; upVer {
		var e error
		newver, e = packages.FormatNewVersion(pack.Version, delta.(int))
		if e != nil {
			return e
		}
		filename := filepath.Join(pack.Path, pack.DisplayedName+".conf")
		e = config.SetFileVersion(ed.log, filename, "Icon", pack.Version, newver)
		if e != nil {
			return e
		}
	}

	// Update edited fields.
	filename := filepath.Join(pack.Path, pack.Source.File())
	e := config.SetToFile(ed.log, filename, func(cfg cdtype.ConfUpdater) (e error) {
		for k, v := range ed.edits {
			switch k {
			case fieldVersion:
				e = cfg.Set(group, "version", newver)

			case fieldCategory:
				e = cfg.Set(group, "category", v.(int))

			case fieldActAsLauncher:
				e = cfg.Set(group, "act as launcher", v.(bool))

			case fieldMultiInstance:
				e = cfg.Set(group, "multi-instance", v.(bool))
			}
			if e != nil {
				return e
			}
		}
		return nil
	})
	if e != nil {
		return e
	}

	// Reload data from disk.
	pack, e = packages.NewAppletPackageUser(ed.log,
		filepath.Dir(pack.Path), filepath.Base(pack.Path),
		pack.Type, pack.Source)

	if e != nil {
		return e
	}
	ed.packs[appID] = pack
	ed.edits = nil
	return nil
}
