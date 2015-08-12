// Package conficons provides an icons list and configuration widget.
package conficons

import (
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/godock/libs/cdtype"

	"github.com/sqp/godock/widgets/cfbuild"
	"github.com/sqp/godock/widgets/cfbuild/cftype"
	"github.com/sqp/godock/widgets/cfbuild/datatype"
	"github.com/sqp/godock/widgets/common"
	"github.com/sqp/godock/widgets/confgui/btnaction"
	"github.com/sqp/godock/widgets/conficons/desktopclass"
	"github.com/sqp/godock/widgets/gtk/newgtk"
	"github.com/sqp/godock/widgets/pageswitch"
	"github.com/sqp/godock/widgets/welcome"

	"errors"
)

const iconSize = 24
const listIconsWidth = 200

//--------------------------------------------------------[ PAGE GUI ICONS ]--

// GuiIcons defines Icons configuration widget for currently actived cairo-dock Icons.
//
type GuiIcons struct {
	gtk.Paned

	icons  *List
	config cftype.Builder

	switcher *pageswitch.Switcher
	btn      btnaction.Tune // save button

	data cftype.Source
	log  cdtype.Logger
}

// New creates a GuiIcons widget to edit cairo-dock icons config.
//
func New(data cftype.Source, log cdtype.Logger, switcher *pageswitch.Switcher, btn btnaction.Tune) *GuiIcons {
	paned := newgtk.Paned(gtk.ORIENTATION_HORIZONTAL)
	widget := &GuiIcons{
		Paned:    *paned,
		switcher: switcher,
		btn:      btn,
		data:     data,
		log:      log,
	}
	widget.icons = NewList(widget, log)

	up := newgtk.ButtonFromIconName("go-up", gtk.ICON_SIZE_BUTTON)
	down := newgtk.ButtonFromIconName("go-down", gtk.ICON_SIZE_BUTTON)
	remove := newgtk.ButtonFromIconName("list-remove", gtk.ICON_SIZE_BUTTON)

	boxLeft := newgtk.Box(gtk.ORIENTATION_VERTICAL, 0)
	boxBtns := newgtk.Box(gtk.ORIENTATION_HORIZONTAL, 0)
	boxLeft.PackStart(widget.icons, true, true, 0)
	boxLeft.PackStart(boxBtns, false, false, 0)
	boxBtns.PackStart(up, false, false, 0)
	boxBtns.PackStart(down, false, false, 0)
	boxBtns.PackEnd(remove, false, false, 0)

	widget.Pack1(boxLeft, true, true)

	widget.SetPosition(listIconsWidth) // Paned position = list icons width.

	up.Connect("clicked", widget.actionSelected(datatype.Iconer.MoveBeforePrevious))
	down.Connect("clicked", widget.actionSelected(datatype.Iconer.MoveAfterNext))
	remove.Connect("clicked", widget.actionSelected(datatype.Iconer.RemoveFromDock))

	// widget.icons.Connect("row-inserted", func() { log.Info("row inserted") })
	// widget.icons.Connect("row-deleted", func() { log.Info("row deleted") })

	return widget
}

// actionSelected prepares a callback to act on the icon of the selected row.
//
func (widget *GuiIcons) actionSelected(call func(datatype.Iconer)) func() {
	return func() {
		ic, e := widget.icons.SelectedIcon()
		if e == nil {
			call(ic)
		}
	}
}

// Load loads the list of icons in the iconsList.
//
func (widget *GuiIcons) Load() {
	icons := widget.data.ListIcons()
	widget.icons.Load(icons)
}

// Selected returns the selected icon.
//
// func (widget *GuiIcons) Selected() datatype.Iconer {
// 	return widget.icons.Selected()
// }

// Select sets the selected icon based on its config file.
//
func (widget *GuiIcons) Select(conf string) bool {
	return widget.icons.Select(conf)
}

// ShowWelcome shows the welcome placeholder widget if nothing is displayed.
//
func (widget *GuiIcons) ShowWelcome(setBtn bool) {
	if widget.config == nil {
		widget.setCurrent(welcome.New(widget.data, widget.log))
		if setBtn {
			widget.btn.SetNone()
		}
	}
}

// Clear clears the widget data.
//
func (widget *GuiIcons) Clear() string {
	path := widget.icons.SelectedConf()
	widget.switcher.Clear()

	if widget.config != nil {
		widget.config.Destroy()
		widget.config = nil
	}
	widget.icons.Clear()
	return path
}

//
//-------------------------------------------------------[ SAVE CONFIG APPLET ]--

// Save saves the current page configuration
//
func (widget *GuiIcons) Save() {
	icon, e := widget.icons.SelectedIcon()
	if widget.config == nil || widget.log.Err(e, "SelectedIcon") {
		return
	}

	widget.config.Save()

	// we reload in case the items place has changed (icon's container, detached...).
	icon.Reload()

	// 	_items_widget_reload (CD_WIDGET (pItemsWidget));  // we reload in case the items place has changed (icon's container, dock orientation, etc).}
}

//
//-------------------------------------------------------[ CONTROL CALLBACKS ]--

// OnSelect reacts when a row is selected. Creates a new config for the icon.
//
func (widget *GuiIcons) OnSelect(icon datatype.Iconer, e error) {
	widget.switcher.Clear()

	if widget.config != nil {
		widget.config.Destroy()
		widget.config = nil
	}

	// Using the welcome widget as fallback for empty fields.
	if widget.log.Err(e, "OnSelect icon") || // shouldn't match.
		icon.ConfigPath() == "" { // for icon.ConfigGroup: datatype.KeyMainDock || datatype.GroupServices.

		widget.ShowWelcome(true)
		return
	}

	// Build a custom config widget from a dock config file.

	// Can be:
	//   field icon     Applet, Launcher, Subdock, Separator.
	//   field custom   TaskBar, Service (applet without icon).
	//   group          Desklets, Alt maindock.

	build, ok := cfbuild.NewFromFileSafe(
		widget.data,
		widget.log,
		icon.ConfigPath(),
		icon.OriginalConfigPath(),
		icon.GetGettextDomain())

	switch {
	case !ok: // Widget already build with an error message.
		widget.btn.SetNone()
		widget.setCurrent(build)
		return

	case icon.ConfigGroup() != "":
		build.BuildSingle(icon.ConfigGroup())

	case icon.IsLauncher():
		tweak := desktopclass.Tweak(build, widget.data, icon.GetClass())
		build.BuildAll(widget.switcher, tweak)

	default:
		build.BuildAll(widget.switcher)
	}

	widget.btn.SetSave()
	widget.setCurrent(build)
}

func (widget *GuiIcons) setCurrent(w cftype.Builder) {
	widget.config = w
	widget.Pack2(w, true, true)
}

//-------------------------------------------------------[ WIDGET ICONS LIST ]--

// ListControl forwards events to other widgets.
//
type ListControl interface {
	OnSelect(datatype.Iconer, error)
}

// List defines a dock icons management widget.
//
type List struct {
	gtk.ScrolledWindow // Main widget is the container. The ScrolledWindow will handle list scrollbars.
	list               *gtk.ListBox

	index map[*gtk.ListBoxRow]datatype.Iconer

	control ListControl // link to higher level widgets.
	log     cdtype.Logger
}

// NewList creates a dock icons management widget.
//
func NewList(control ListControl, log cdtype.Logger) *List {
	widget := &List{
		ScrolledWindow: *newgtk.ScrolledWindow(nil, nil),
		list:           newgtk.ListBox(),
		control:        control,
		log:            log,
		index:          make(map[*gtk.ListBoxRow]datatype.Iconer),
	}

	widget.list.Connect("row-selected", widget.onSelectionChanged)
	widget.Add(widget.list)

	return widget
}

// Load loads icons list in the widget.
//
func (widget *List) Load(icons *datatype.ListIcon) {
	widget.Clear()

	for _, iconContainer := range icons.Maindocks {
		widget.addBoxItem(iconContainer.Container, 0, true)
		widget.iconsDock(iconContainer.Icons, icons.Subdocks, 1)
	}

	widget.ShowAll()
}

// iconsDock adds a list of icons, and fills subdocks content if needed (recursive).
//
func (widget *List) iconsDock(icons []datatype.Iconer, subdocks map[string][]datatype.Iconer, indent int) {
	for _, icon := range icons {
		widget.addBoxItem(icon, indent, icon.IsStackIcon())

		if icon.IsStackIcon() {
			name, _ := icon.DefaultNameIcon()
			subicons, ok := subdocks[name]
			if ok {
				widget.iconsDock(subicons, subdocks, indent+1)
			}
		}
	}
}

// addBoxItem adds an item to the list.
//
func (widget *List) addBoxItem(icon datatype.Iconer, indent int, bold bool) *gtk.ListBoxRow {
	row := newgtk.ListBoxRow()
	box := newgtk.Box(gtk.ORIENTATION_HORIZONTAL, 0)
	row.Add(box)

	name, img := icon.DefaultNameIcon()

	box.Set("margin-start", 15*indent)
	if bold {
		name = common.Bold(name)
	}
	if img != "" {
		if pix, e := common.ImageNewFromFile(img, iconSize); !widget.log.Err(e, "Load icon") {
			box.PackStart(pix, false, false, 0)
		}
	}
	lbl := newgtk.Label(name)
	lbl.SetUseMarkup(true)
	box.PackStart(lbl, false, false, 0)

	widget.list.Add(row)
	widget.index[row] = icon

	return row
}

// Clear clears the widget data.
//
func (widget *List) Clear() {
	for box := range widget.index {
		widget.list.Remove(box)
	}
	widget.index = make(map[*gtk.ListBoxRow]datatype.Iconer)
}

//---------------------------------------------------------------[ SELECTION ]--

// SelectedIcon returns the iconer matching the selected row.
//
func (widget *List) SelectedIcon() (datatype.Iconer, error) {
	sel := widget.list.GetSelectedRow()
	if sel == nil {
		return nil, errors.New("no selection")
	}
	for box, icon := range widget.index {
		if box.Native() == sel.Native() {
			return icon, nil
		}
	}
	return nil, errors.New("no matching icon for selection")
}

// SelectedConf returns the path to icon config file for the selected row.
//
func (widget *List) SelectedConf() string {
	icon, e := widget.SelectedIcon()
	if e != nil {
		return ""
	}
	return icon.ConfigPath()
}

// Select sets the selected icon based on its config file.
//
func (widget *List) Select(conf string) bool {
	if conf != "" {
		for box, icon := range widget.index {
			if icon.ConfigPath() == conf {
				box.Activate()
				return true
			}
		}
	}
	return false
}

//-------------------------------------------------------[ ACTIONS CALLBACKS ]--

// Selected line has changed. Forward the call to the controler.
//
func (widget *List) onSelectionChanged(box *gtk.ListBox, row *gtk.ListBoxRow) {
	widget.control.OnSelect(widget.SelectedIcon())
}
