// Package docktheme provides a cairo-dock main theme load/save widget.
//
// Still use the themes.conf file with custom widgets, but as virtual (no temp).
//
// TODO:
//   Update what's needed after load/save.
//   Load add ratings.
//   Save add preview (change newComboBox call ?)
//
package docktheme

import (
	"github.com/sqp/godock/libs/cdglobal"
	"github.com/sqp/godock/libs/cdtype"

	"github.com/sqp/godock/widgets/cfbuild"
	"github.com/sqp/godock/widgets/cfbuild/cftype"
	"github.com/sqp/godock/widgets/cfbuild/cfwidget"
	"github.com/sqp/godock/widgets/confapplets"
	"github.com/sqp/godock/widgets/pageswitch"
)

const groupLoad = "Load theme"
const groupSave = "Save"

//-------------------------------------------------------[ WIDGET CORE LIST ]--

// DockTheme provides a dock main theme load/save widget.
//
type DockTheme struct {
	cftype.Grouper

	switcher *pageswitch.Switcher

	data cftype.Source
	log  cdtype.Logger
}

// New creates a DockTheme widget to load and save the main theme.
//
func New(data cftype.Source, log cdtype.Logger, switcher *pageswitch.Switcher) (cftype.Grouper, bool) {
	file := data.DirShareData(cdglobal.FileConfigThemes)
	build, ok := cfbuild.NewFromFileSafe(data, log, file, "", "")
	if !ok {
		return build, false
	}

	// Widget building.
	keyLoad := func(key *cftype.Key) {
		w := confapplets.NewLoaded(data, log, nil, confapplets.ListThemes)
		getValue := func() interface{} { return w.Selected().GetName() }
		key.PackKeyWidget(getValue, nil, w)
	}

	keySave := func(key *cftype.Key) {
		cfwidget.PackComboBoxWithListField(key, true, false, data.ListDockThemeSave)
	}

	build.BuildAll(switcher,
		cfbuild.TweakKeySetAlignedVertical(groupLoad, "chosen theme"),
		cfbuild.TweakKeyMakeWidget(groupLoad, "chosen theme", keyLoad),
		cfbuild.TweakKeyMakeWidget(groupSave, "theme name", keySave),
	)

	// Builder update keys.
	return &DockTheme{
		Grouper:  build,
		switcher: switcher,
		data:     data,
		log:      log,
	}, true
}

//
//--------------------------------------------------------------------[ SAVE ]--

// Save activates load or save action.
//
func (widget *DockTheme) Save() {
	switch widget.switcher.Selected() {
	case groupLoad:
		// Use package location before the selected theme.
		themeName := widget.KeyString(groupLoad, "package")
		if themeName == "" {
			themeName = widget.KeyString(groupLoad, "chosen theme")
		}

		if themeName == "" {
			widget.log.NewErr("no theme selected", "load dock theme")
			return
		}

		useBehaviour := widget.KeyBool(groupLoad, "use theme behaviour")
		useLaunchers := widget.KeyBool(groupLoad, "use theme launchers")

		e := widget.data.CurrentThemeLoad(themeName, useBehaviour, useLaunchers)
		if !widget.log.Err(e, "load dock theme") {
			// 		cairo_dock_reload_gui ();
		}

	case groupSave:
		themeName := widget.KeyString(groupSave, "theme name")
		saveBehaviour := widget.KeyBool(groupSave, "save current behaviour")
		saveLaunchers := widget.KeyBool(groupSave, "save current launchers")
		needPackage := widget.KeyBool(groupSave, "package")
		dirPackage := widget.KeyString(groupSave, "package dir")

		if dirPackage == "Home directory" {
			dirPackage = ""
		}

		e := widget.data.CurrentThemeSave(themeName, saveBehaviour, saveLaunchers, needPackage, dirPackage)
		widget.log.Err(e, "save dock theme")

		// 				cairo_dock_set_status_message (NULL, _("Theme has been saved"));
		// 				_fill_treeview_with_themes (pThemesWidget);
		// 				_fill_combo_with_user_themes (pThemesWidget);
		// 				cairo_dock_gui_select_in_combo_full (pThemesWidget->pCombo, cNewThemeName, TRUE);
		// 			}

	}
}
