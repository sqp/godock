// Package confbuilder builds a cairo-dock configuration widget from its config file.
package confbuilder

import (
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/godock/libs/cdtype"

	"github.com/sqp/godock/widgets/confbuilder/datatype"
	"github.com/sqp/godock/widgets/gtk/keyfile"
	"github.com/sqp/godock/widgets/pageswitch"

	"strconv"
	"strings"
)

// CairoConfig builds Cairo-Dock configuration page widgets.
//
type CairoConfig struct {
	keyfile.KeyFile
	File string
}

// LoadFile loads a Cairo-Dock configuration file as *CairoConfig.
func LoadFile(configFile string) (*CairoConfig, error) {
	pKeyF := keyfile.New()

	_, e := pKeyF.LoadFromFile(configFile, keyfile.FlagsKeepComments|keyfile.FlagsKeepTranslations) // (bool, error)
	if e != nil {
		// pKeyF.Free()
		return nil, e
	}
	conf := &CairoConfig{
		KeyFile: *pKeyF,
		File:    configFile,
	}
	return conf, nil
}

// List lists keys defined in the configuration file.
//
func (conf *CairoConfig) List(cGroupName string) (list []*Key) {
	_, keys, _ := conf.GetKeys(cGroupName) // (uint64, []string, error)
	for _, cKeyName := range keys {
		cKeyComment, _ := conf.GetComment(cGroupName, cKeyName)

		if key := ParseKeyComment(cKeyComment); key != nil {
			if key.Type == '[' { // on gere le bug de la Glib, qui rajoute les nouvelles cles apres le commentaire du groupe suivant !
				// log.DEV("LIBC BUG, DETECTED", cKeyComment) // often seem to be a [gtk-convert]
				continue
			}

			key.Group = cGroupName
			key.Name = cKeyName
			list = append(list, key)
		}
	}
	return
}

// Builder returns a builder ready to create a configuration gui for the keyfile.
//
func (conf *CairoConfig) Builder(source datatype.Source, log cdtype.Logger, win *gtk.Window, gettextDomain string) *Grouper {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	build := &Builder{
		Box:           *box,
		Conf:          conf,
		data:          source,
		log:           log,
		win:           win,
		gettextDomain: gettextDomain,
	}
	return &Grouper{*build}
}

// Grouper builds config pages from the Builder.
//
type Grouper struct{ Builder }

// NewGrouper creates a config page builder from the file.
//
func NewGrouper(source datatype.Source, log cdtype.Logger, win *gtk.Window, configFile, gettextDomain string) (*Grouper, error) {
	keyfile, e := LoadFile(configFile)
	if e != nil {
		return nil, e
	}
	return keyfile.Builder(source, log, win, gettextDomain), nil
}

// BuildSingle builds a single page config for the given group.
//
func (build *Grouper) BuildSingle(group string) *Grouper {
	build.PackStart(build.BuildPage(group), true, true, 0)
	build.ShowAll()
	return build
}

// BuildAll builds a Cairo-Dock configuration gui page directly from file.
// Too much magic to be described here. Need to find the links.
//
func (build *Grouper) BuildAll(switcher *pageswitch.Switcher) *Grouper { //(build *Builder, e error) {

	first := true
	_, groups := build.Conf.GetGroups()
	for _, group := range groups {
		w := build.BuildPage(group)

		switcher.AddPage(&pageswitch.Page{
			Name:    group,
			OnShow:  func() { build.PackStart(w, true, true, 0); w.ShowAll() },
			OnHide:  func() { build.Remove(w) },
			OnClear: func() { w.Destroy() }})

		if first {
			switcher.Activate(group)
			first = false
		}
	}

	// Single group, hide the switcher. Multi groups, display it.
	switcher.Set("visible", len(groups) > 1)

	return build
}

// KeyFiler defines the interface to recognise a grouper (provides its KeyFile).
//
type KeyFiler interface {
	KeyFile() *keyfile.KeyFile
}

// KeyFile returns the pointer to the internal KeyFile.
//
func (build *Grouper) KeyFile() *keyfile.KeyFile {
	return &build.Conf.KeyFile
}

//
//-----------------------------------------------------------------[ PARSING ]--

// ParseKeyComment parse comments for a key.
//
func ParseKeyComment(cKeyComment string) *Key {
	cUsefulComment := strings.TrimLeft(cKeyComment, "# \n")  // remove #, spaces, and endline from start.
	cUsefulComment = strings.TrimRight(cUsefulComment, "\n") // remove endline from end.

	if len(cKeyComment) < 2 {
		// log.DEV("dropped comment", cKeyComment)
		return nil
	}

	key := &Key{Type: cUsefulComment[0]}
	cUsefulComment = cUsefulComment[1:]

	for i, c := range cUsefulComment {
		if c != '-' && c != '+' && c != ' ' {
			if c == WidgetCairoOnly {
				// If opengl, need drop key

			} else if c == WidgetOpenGLOnly {
				// If !opengl, need drop key

			} else {
				// Try to detect a value indicating the number of elements.
				key.NbElements, _ = strconv.Atoi(string(cUsefulComment[i:]))

				// Try to get authorized values between square brackets.
				if c == '[' {
					values := cUsefulComment[i+1 : strings.Index(cUsefulComment, "]")]
					i += len(values) + 1

					key.AuthorizedValues = strings.Split(values, ";")
				}

				// End of arguments at the start .
				cUsefulComment = cUsefulComment[i:]
				break
			}
		}
	}

	if key.NbElements == 0 {
		key.NbElements = 1
	}

	cUsefulComment = strings.TrimLeft(cUsefulComment, "]1234567890") // Remove last bits of possible arguments.
	cUsefulComment = strings.TrimLeft(cUsefulComment, " ")           // Remove separator.

	// log.DEV("parsed", string(iType), iNbElements, cUsefulComment)

	// Special widget alignment with a trailing slash.
	if strings.HasSuffix(cUsefulComment, "/") {
		cUsefulComment = strings.TrimSuffix(cUsefulComment, "/")
		key.IsAlignedVertical = true
	}

	// Get tooltip.
	toolStart := strings.IndexByte(cUsefulComment, '{')
	toolEnd := strings.IndexByte(cUsefulComment, '}')
	if toolStart > 0 && toolEnd > 0 && toolStart < toolEnd {
		key.Tooltip = cUsefulComment[toolStart+1 : toolEnd]
		cUsefulComment = cUsefulComment[:toolStart-1]
	}

	key.Text = cUsefulComment

	return key
}
