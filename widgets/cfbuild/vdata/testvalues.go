package vdata

import (
	"github.com/sqp/godock/libs/ternary"

	"github.com/sqp/godock/widgets/cfbuild/cftype"   // Types for config file builder usage.
	"github.com/sqp/godock/widgets/cfbuild/datatype" // Types for config file builder data source.
)

// TestValues updates the key data with test values.
//
func TestValues(key *cftype.Key) {
	switch key.Type {
	case cftype.KeyBoolButton,
		cftype.KeyBoolCtrl:

		if key.NbElements > 1 {
			val := key.Value().ListBool()
			newval := make([]bool, len(val))
			for i, v := range val {
				newval[i] = !v
			}
			key.ValueSet(newval)

		} else {
			val := key.Value().Bool()
			newval := !val
			key.ValueSet(newval)
		}

	case cftype.KeyIntSpin,
		cftype.KeyIntSize,
		cftype.KeyIntScale:

		if key.NbElements > 1 {
			val := key.Value().ListInt()
			newval := make([]int, len(val))
			for i, v := range val {
				newval[i] = v + 1 + i
			}
			key.ValueSet(newval)
		} else {
			val := key.Value().Int()
			newval := val + 1
			key.ValueSet(newval)
		}

	case cftype.KeyFloatSpin,
		cftype.KeyFloatScale:

		if key.NbElements > 1 {
			val := key.Value().ListFloat()
			newval := make([]float64, len(val))
			for i, v := range val {
				newval[i] = v + 0.1 + float64(i) // hack +0.1 +i
			}
			key.ValueSet(newval)
		} else {
			val := key.Value().Float()
			newval := val + 0.1
			key.ValueSet(newval)
		}

	case cftype.KeyColorSelectorRGB,
		cftype.KeyColorSelectorRGBA:

		val := key.Value().ListFloat()
		newval := make([]float64, len(val))
		for i := range val {
			newval[i] = val[i] + 0.1
			if newval[i] > 1 {
				newval[i]--
			}
		}
		if len(newval) == 3 {
			newval = append(newval, 1)
		}
		key.ValueSet(newval)

	case cftype.KeyLink:

		val := key.Value().String()
		list := []string{urlDock, urlGoTour}
		newval := cycleNextString(list, val, key)
		key.ValueSet(newval)

	case cftype.KeyStringEntry,
		cftype.KeyFileSelector, cftype.KeyFolderSelector,
		cftype.KeyImageSelector, cftype.KeySoundSelector,
		cftype.KeyShortkeySelector, cftype.KeyClassSelector,
		cftype.KeyPasswordEntry,
		cftype.KeyListEntry:

		val := key.Value().String()
		newval := cycleNextString(otherStrings, val, nil)
		key.ValueSet(newval)

	case cftype.KeyFontSelector:

		val := key.Value().String()
		list := []string{"Arial 8", "Monospace 8"}
		newval := cycleNextString(list, val, key)
		key.ValueSet(newval)

	case cftype.KeyTreeViewSortSimple,
		cftype.KeyTreeViewSortModify:

		val := key.Value().ListString()
		newval := reverseStrings(val)
		key.ValueSet(newval)

	case cftype.KeyListNumbered,
		cftype.KeyListNbCtrlSimple,
		cftype.KeyListNbCtrlSelect:

		val := key.Value().Int()
		step := ternary.Int(key.IsType(cftype.KeyListNbCtrlSelect), 3, 1)
		newval := cycleNextID(len(key.AuthorizedValues), val, step)
		key.ValueSet(newval)

	case cftype.KeyListSimple:
		val := key.Value().String()
		newval := cycleNextString(key.AuthorizedValues, val, key)
		key.ValueSet(newval)

	case cftype.KeyListDocks:

		val := key.Value().String()
		list := []string{datatype.KeyMainDock, datatype.KeyNewDock}
		newval := cycleNextString(list, val, key)
		key.ValueSet(newval)

	case cftype.KeyListThemeApplet:

		val := key.Value().String()
		list := []string{"Turbo-night-fuel[0]", "Sound-Mono[0]"}
		newval := cycleNextString(list, val, key)
		key.ValueSet(newval)

	case cftype.KeyHandbook:

		val := key.Value().String()
		list := datatype.ListHandbooksKeys(Handbooks)
		newval := cycleNextString(list, val, key)
		key.ValueSet(newval)

	case cftype.KeyListViews:

		val := key.Value().String()
		books := key.Source().ListViews()
		list := datatype.IndexHandbooksKeys(books)
		newval := cycleNextString(list, val, key)
		key.ValueSet(newval)

	case cftype.KeyListIconsMainDock:

		val := key.Value().String()
		fields := key.Source().ListIconsMainDock()
		newval := cycleNextField(fields, val, key)
		key.ValueSet(newval)

	case cftype.KeyListThemeDesktopIcon:

		val := key.Value().String()
		fields := key.Source().ListThemeDesktopIcon()
		newval := cycleNextField(fields, val, key)
		key.ValueSet(newval)

	}
}

// update strings data.
var otherStrings = []string{"value changed", "this is an edit", "also updated", "other text"}

// update link data.
const urlDock = "http://glx-dock.org/"
const urlGoTour = "https://tour.golang.org/"

func reverseStrings(list []string) []string {
	size := len(list)
	newval := make([]string, size)
	for i, v := range list {
		newval[size-1-i] = v
	}
	if size%2 == 1 { // Update the middle one if needed.
		if len(newval[size/2]) < 12 {
			newval[size/2] += "-"
		} else {
			newval[size/2] = newval[size/2][:11]
		}
	}
	return newval
}

func cycleNextID(size, id, step int) int {
	return ternary.Int((id+1)*step < size, id+1, 0)
}

func cycleNextString(list []string, current string, key *cftype.Key) string {
	newID := findID(list, current, key)
	return list[cycleNextID(len(list), newID, 1)]
}

func cycleNextField(fields []datatype.Field, current string, key *cftype.Key) string {
	newID := datatype.ListFieldsIDByName(fields, current, key.Log())
	return fields[cycleNextID(len(fields), newID, 1)].Key
}

func findID(list []string, current string, key *cftype.Key) int {
	for i, str := range list {
		if current == str {
			return i
		}
	}
	if key != nil {
		key.Log().NewErr("not found", "list findID", current, list)
	}
	return 0
}
