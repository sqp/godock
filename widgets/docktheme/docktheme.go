// Package docktheme provides a cairo-dock main theme load/save widget.
//
// Still use the themes.conf file with custom widgets, but as virtual (no temp).
//
// TODO:
//   Update what's needed after load/save.
//   Load add ratings.
//   Save add preview (change newComboBox call ?)
package docktheme

import (
	"github.com/sqp/godock/libs/cdglobal"
	"github.com/sqp/godock/libs/cdtype"

	"github.com/sqp/godock/widgets/confapplets"
	"github.com/sqp/godock/widgets/confbuilder"
	"github.com/sqp/godock/widgets/pageswitch"
)

const groupLoad = "Load theme"
const groupSave = "Save"

//-------------------------------------------------------[ WIDGET CORE LIST ]--

// DockTheme provides a dock main theme load/save widget.
//
type DockTheme struct {
	confbuilder.Grouper

	switcher *pageswitch.Switcher

	data confbuilder.Source
	log  cdtype.Logger
}

// New creates a DockTheme widget to load and save the main theme.
//
func New(data confbuilder.Source, log cdtype.Logger, switcher *pageswitch.Switcher) *DockTheme {
	file := data.DirShareData(cdglobal.FileConfigThemes)
	build, e := confbuilder.NewGrouper(data, log, file, "", "")
	if log.Err(e, "Load Keyfile "+file) {
		return nil
	}

	widgetLoad := func(build *confbuilder.Builder, key *confbuilder.Key) {
		w := confapplets.NewLoaded(data, log, nil, confapplets.ListThemes)
		getValue := func() interface{} { return w.Selected().GetName() }
		build.AddKeyWidget(w, key, getValue)
	}

	widgetSave := func(build *confbuilder.Builder, key *confbuilder.Key) {
		build.NewComboBoxFilled(key, true, false, data.ListDockThemeSave)
	}

	hack := func(build *confbuilder.Builder) {
		key, e := build.GetKey(groupLoad, "chosen theme").SetCustom(widgetLoad)
		if !log.Err(e, "get key=", groupLoad, "::", "chosen theme") {
			key.IsAlignedVertical = true
		}

		_, e = build.GetKey(groupSave, "theme name").SetCustom(widgetSave)
		log.Err(e, "get key=", groupSave, "::", "theme name")
	}

	return &DockTheme{
		Grouper:  *build.BuildAll(switcher, hack),
		switcher: switcher,
		data:     data,
		log:      log,
	}
}

//
//--------------------------------------------------------------------[ SAVE ]--

// Save activates load or save action.
//
func (widget *DockTheme) Save() {
	switch widget.switcher.Selected() {
	case groupLoad:

		themeName, e := widget.GetKey(groupLoad, "package").String()
		widget.log.Err(e, "get key")

		if themeName == "" {
			themeName, e = widget.GetKey(groupLoad, "chosen theme").String()
			widget.log.Err(e, "get key")
		}

		if themeName == "" {
			widget.log.NewErr("no theme selected", "load dock theme")
			return
		}

		useBehaviour, e := widget.GetKey(groupLoad, "use theme behaviour").Bool()
		widget.log.Err(e, "get key")

		useLaunchers, e := widget.GetKey(groupLoad, "use theme launchers").Bool()
		widget.log.Err(e, "get key")

		e = widget.data.CurrentThemeLoad(themeName, useBehaviour, useLaunchers)
		if !widget.log.Err(e, "load dock theme") {
			// 		cairo_dock_reload_gui ();
		}

	case groupSave:
		themeName, e := widget.GetKey(groupSave, "theme name").String()
		widget.log.Err(e, "get key")

		saveBehaviour, e := widget.GetKey(groupSave, "save current behaviour").Bool()
		widget.log.Err(e, "get key")

		saveLaunchers, e := widget.GetKey(groupSave, "save current launchers").Bool()
		widget.log.Err(e, "get key")

		needPackage, e := widget.GetKey(groupSave, "package").Bool()
		widget.log.Err(e, "get key")

		dirPackage, e := widget.GetKey(groupSave, "package dir").String()
		widget.log.Err(e, "get key")

		if dirPackage == "Home directory" {
			dirPackage = ""
		}

		e = widget.data.CurrentThemeSave(themeName, saveBehaviour, saveLaunchers, needPackage, dirPackage)
		widget.log.Err(e, "save dock theme")

		// 				cairo_dock_set_status_message (NULL, _("Theme has been saved"));
		// 				_fill_treeview_with_themes (pThemesWidget);
		// 				_fill_combo_with_user_themes (pThemesWidget);
		// 				cairo_dock_gui_select_in_combo_full (pThemesWidget->pCombo, cNewThemeName, TRUE);
		// 			}

	}
}
