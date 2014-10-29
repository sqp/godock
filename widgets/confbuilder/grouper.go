// Package confbuilder builds a cairo-dock configuration widget from its config file.
package confbuilder

import (
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/godock/libs/log"

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
				log.DEV("LIBC BUG, DETECTED")
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
func (conf *CairoConfig) Builder(source datatype.Source, win *gtk.Window) *Grouper {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	build := &Builder{
		Box:  *box,
		Conf: conf,
		data: source,
		win:  win,
	}
	return &Grouper{*build}
}

// Grouper builds config pages from the Builder.
//
type Grouper struct{ Builder }

// NewGrouper creates a config page builder from the file.
//
func NewGrouper(source datatype.Source, win *gtk.Window, configFile string) (*Grouper, error) {
	keyfile, e := LoadFile(configFile)
	if e != nil {
		return nil, e
	}
	return keyfile.Builder(source, win), nil
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

// ParseKeyComment parse comments for a key.
//
func ParseKeyComment(cKeyComment string) *Key {
	//  gchar ***pAuthorizedValuesList, gboolean *bAligned, const gchar **cTipString

	cUsefulComment := strings.TrimLeft(cKeyComment, "# \n")  // remove #, spaces, and endline from start.
	cUsefulComment = strings.TrimRight(cUsefulComment, "\n") // remove endline from end.

	if len(cKeyComment) < 2 {
		// log.DEV("dropped comment", cKeyComment)
		return nil
	}

	key := &Key{Type: cUsefulComment[0]}
	cUsefulComment = cUsefulComment[1:]

	for i, c := range cUsefulComment {
		if c != '-' && c != '+' {
			if c != '*' && c != '&' { // NEED TO IMPROVE TEST OPENGL & CAIRO

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

	// if (*cUsefulComment == WIDGET_CAIRO_ONLY)
	// {
	// 	if (g_bUseOpenGL)
	// 		return NULL;
	// 	cUsefulComment ++;
	// }
	// else if (*cUsefulComment == WIDGET_OPENGL_ONLY)
	// {
	// 	if (! g_bUseOpenGL)
	// 		return NULL;
	// 	cUsefulComment ++;
	// }

	// //\______________ On recupere l'alignement.
	// int len = strlen (cUsefulComment);
	// if (cUsefulComment[len - 1] == '\n')
	// {
	// 	len --;
	// 	cUsefulComment[len] = '\0';
	// }
	// if (cUsefulComment[len - 1] == '/')
	// {
	// 	cUsefulComment[len - 1] = '\0';
	// 	*bAligned = FALSE;
	// }
	// else
	// {
	// 	*bAligned = TRUE;
	// }

	// Get tooltip.
	toolStart := strings.IndexByte(cUsefulComment, '{')
	toolEnd := strings.IndexByte(cUsefulComment, '}')
	if toolStart > 0 && toolEnd > 0 && toolStart < toolEnd {
		key.Tooltip = cUsefulComment[toolStart+1 : toolEnd-1]
		cUsefulComment = cUsefulComment[:toolStart-1]
	}

	key.Text = cUsefulComment

	return key
}
