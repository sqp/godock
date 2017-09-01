// Package appinfo provides an applets list with edit applet info widget.
//
package appinfo

import (
	"github.com/gotk3/gotk3/gtk"

	"github.com/sqp/godock/libs/cdglobal"     // Global consts.
	"github.com/sqp/godock/libs/cdtype"       // Applets types.
	"github.com/sqp/godock/libs/packages"     // Dock package.
	"github.com/sqp/godock/libs/text/gtktext" // Text format gtk.
	"github.com/sqp/godock/libs/text/tran"    // Translate.

	"github.com/sqp/godock/widgets/appletlist"
	"github.com/sqp/godock/widgets/cfbuild"        // The config file builder.
	"github.com/sqp/godock/widgets/cfbuild/cftype" // Types for config file builder usage.
	"github.com/sqp/godock/widgets/cfbuild/datatype"
	"github.com/sqp/godock/widgets/cfbuild/newkey" // Create config file builder keys.
	"github.com/sqp/godock/widgets/confapplets"
	"github.com/sqp/godock/widgets/gtk/newgtk"
	"github.com/sqp/godock/widgets/pageswitch" // Switcher for config pages.

	"path/filepath"
	"strings"
)

//
//-------------------------------------------------[ WIDGET APPLETS DOWNLOAD ]--

// GUIControl is the interface to the main GUI and data source.
//
type GUIControl interface {
	ListKnownApplets() map[string]datatype.Appleter
}

// ConfApplet provides an applets list with edit applet info widget.
//
type ConfApplet struct {
	gtk.Box
	applets confapplets.ListInterfaceBase
	edit    *EditInfo

	source cftype.Source
	log    cdtype.Logger

	Applets *map[string]datatype.Appleter // List of applets known by the Dock.
}

// New creates a widget to list cairo-dock applets and themes.
//
func New(source cftype.Source, log cdtype.Logger, switcher *pageswitch.Switcher) *ConfApplet {
	widget := &ConfApplet{
		Box:    *newgtk.Box(gtk.ORIENTATION_HORIZONTAL, 0),
		edit:   NewEdit(log),
		source: source,
		log:    log,
	}

	build := widget.edit.NewBuilder(source, log, switcher)

	widget.applets = appletlist.NewListEditApp(widget, log)

	widget.PackStart(widget.applets, false, false, 0)
	widget.PackStart(build, true, true, 4)
	widget.ShowAll()
	return widget
}

// NewLoaded creates a widget to list cairo-dock applets and themes and loads data.
//
func NewLoaded(source cftype.Source, log cdtype.Logger, switcher *pageswitch.Switcher) *ConfApplet {
	w := New(source, log, switcher)
	w.Load()
	return w
}

// Load list of applets in the appletlist.
//
func (widget *ConfApplet) Load() {
	applets := widget.source.ListKnownApplets()
	widget.applets.Load(applets)
}

// Save saves current applet info (user clicked the save button).
//
func (widget *ConfApplet) Save() {
	if widget.edit.pack == nil {
		return
	}
	edits := widget.edit.Edited()
	pack, e := widget.edit.pack.SaveUpdated(edits)
	if !widget.log.Err(e, "save appinfo") {
		widget.edit.SetPack(pack)
	}
}

// Clear clears the widget data.
//
func (widget *ConfApplet) Clear() {
	widget.applets.Clear()
}

//
//--------------------------------------------------[ LIST CONTROL CALLBACKS ]--

// OnSelect reacts when a row is selected. Sets the new selected package.
//
func (widget *ConfApplet) OnSelect(selpack datatype.Appleter) {
	pack, e := packages.NewAppletPackageUser(widget.log,
		filepath.Dir(selpack.Dir()),
		selpack.GetName(),
		cdtype.PackTypeUser,
		packages.SourceApplet)

	if widget.log.Err(e, "appinfo NewAppletPackageUser") {
		pack = &packages.AppletPackage{
			DisplayedName: "Error loading package info",
			Description:   e.Error(),
		}
	}
	widget.edit.SetPack(pack)
}

// SetControlInstall forwards the list controler to the menu for updates.
//
func (widget *ConfApplet) SetControlInstall(ctrl appletlist.ControlInstall) {}

//
//---------------------------------------------------------------[ EDIT INFO ]--

// EditInfo defines an applet info edit widget.
//
type EditInfo struct {
	Title         *cftype.Key
	Author        *cftype.Key
	Version       *cftype.Key
	Category      *cftype.Key
	MultiInstance *cftype.Key
	ActAsLauncher *cftype.Key
	Description   *cftype.Key
	Name          *cftype.Key
	Icon          *cftype.Key
	Path          *cftype.Key

	pack *packages.AppletPackage
	dir  string // package dir.
}

// NewEdit creates a welcome widget with informations about the program.
//
func NewEdit(log cdtype.Logger) *EditInfo {
	var (
		group      = packages.SourceApplet.Group()
		categories = cdtype.Categories()
		versions   = []string{
			tran.Slate("No change"),
			tran.Slate("Upgrade micro"),
			tran.Slate("Upgrade minor"),
			tran.Slate("Upgrade major"),
		}
	)

	ed := &EditInfo{
		Title: newkey.TextLabel(group, "txt_title",
			gtktext.Bold(gtktext.Big(tran.Slate("Edit applet info"))),
		),
		Author: newkey.StringEntry(group,
			packages.AppInfoAuthor.Key(),
			packages.AppInfoAuthor.Translated(),
		),
		Version: newkey.ListNumbered(group,
			packages.AppInfoVersion.Key(),
			packages.AppInfoVersion.Translated(),
			versions...,
		),
		Category: newkey.ListNumbered(group,
			packages.AppInfoCategory.Key(),
			packages.AppInfoCategory.Translated(),
			categories...,
		),
		MultiInstance: newkey.Bool(group,
			packages.AppInfoMultiInstance.Key(),
			packages.AppInfoMultiInstance.Translated(),
		),
		ActAsLauncher: newkey.Bool(group,
			packages.AppInfoActAsLauncher.Key(),
			packages.AppInfoActAsLauncher.Translated(),
		),
		Description: newkey.TextArea(group,
			packages.AppInfoDescription.Key(),
			packages.AppInfoDescription.Translated(),
			log,
		),
		Name: newkey.StringEntry(group,
			packages.AppInfoTitle.Key(),
			tran.Slate("Optional")+" - "+packages.AppInfoTitle.Translated(),
		),
		Icon: newkey.StringEntry(group,
			packages.AppInfoIcon.Key(),
			tran.Slate("Optional")+" - "+packages.AppInfoIcon.Translated(),
		),
	}

	ed.Path = newkey.CustomButton(group,
		"Path",
		" ", // Label can't be empty to allow creation of the key label widget.
		newkey.Call{Label: tran.Slate("Open dir"), Func: func() {
			if ed.dir != "" {
				log.ExecAsync(cdglobal.CmdOpen, ed.dir)
			}
		}},
	)

	// Extra settings.
	ed.Author.IsAlignedVertical = true
	ed.Version.Tooltip = "The version will also be upgraded in the default config file."
	ed.Name.Tooltip = "Set only if you want to use a name other than applet directory name."
	ed.Icon.Tooltip = "Set only if you want to use an icon name other than 'icon'"

	return ed
}

// NewBuilder creates the builder widget to edit applet info.
//
func (ed *EditInfo) NewBuilder(source cftype.Source, log cdtype.Logger, switcher *pageswitch.Switcher) cftype.Grouper {
	group := packages.SourceApplet.Group()
	addkeys := cfbuild.TweakAddGroup(group, // Displayed fields and order.
		ed.Title,
		newkey.Separator(group, "sep_title"),
		ed.Version,
		ed.Category,
		ed.Author,
		ed.Description,
		ed.MultiInstance,
		ed.ActAsLauncher,
		ed.Name,
		ed.Icon,
		newkey.Separator(group, "sep_path"),
		ed.Path,
	)

	setvalues := func(build cftype.Builder) { // Set empty values for startup.
		st := build.Storage()
		st.Set(group, packages.AppInfoAuthor.Key(), "")
		st.Set(group, packages.AppInfoVersion.Key(), "")
		st.Set(group, packages.AppInfoCategory.Key(), "")
		st.Set(group, packages.AppInfoActAsLauncher.Key(), []bool{false})
		st.Set(group, packages.AppInfoMultiInstance.Key(), []bool{false})
		st.Set(group, packages.AppInfoDescription.Key(), "")
		st.Set(group, packages.AppInfoTitle.Key(), "")
		st.Set(group, packages.AppInfoIcon.Key(), "")
	}

	build := cfbuild.NewVirtual(source, log, "", "", "")
	build.BuildAll(switcher, addkeys, setvalues)

	ed.Path.Label().SetSelectable(true)
	// ed.Path.Label().SetLineWrapMode(gtk.WRAP_WORD)

	return build
}

// SetPack applies package data in the editor.
//
func (ed *EditInfo) SetPack(pack *packages.AppletPackage) {
	ed.Title.Label().SetMarkup(gtktext.Bold(gtktext.Big(pack.DisplayedName)) + " v" + pack.Version)
	ed.Author.ValueSet(pack.Author)
	ed.Version.ValueSet(0) // 0 = unchanged.
	ed.Category.ValueSet(int(pack.Category))
	ed.MultiInstance.ValueSet(pack.IsMultiInstance)
	ed.ActAsLauncher.ValueSet(pack.ActAsLauncher)
	ed.Description.ValueSet(strings.Replace(pack.Description, "\\n", "\n", -1))
	ed.Name.ValueSet(pack.Title)
	ed.Icon.ValueSet(pack.Icon)
	ed.Path.Label().SetLabel(pack.Path)
	ed.pack = pack
	ed.dir = pack.Path
}

// Edited returns the list of edited fields.
//
func (ed *EditInfo) Edited() map[packages.AppInfoField]interface{} {
	edits := map[packages.AppInfoField]interface{}{}

	if newval := ed.Version.Value().Int(); newval != 0 {
		edits[packages.AppInfoVersion] = newval
	}

	if newval := cdtype.CategoryType(ed.Category.Value().Int()); newval != ed.pack.Category {
		edits[packages.AppInfoCategory] = newval
	}

	if newval := ed.Author.Value().String(); newval != ed.pack.Author {
		edits[packages.AppInfoAuthor] = newval
	}

	if newval := ed.Description.Value().String(); newval != ed.pack.Description {
		edits[packages.AppInfoDescription] = strings.Replace(newval, "\n", "\\n", -1)
	}

	if newval := ed.ActAsLauncher.Value().Bool(); newval != ed.pack.ActAsLauncher {
		edits[packages.AppInfoActAsLauncher] = newval
	}

	if newval := ed.MultiInstance.Value().Bool(); newval != ed.pack.IsMultiInstance {
		edits[packages.AppInfoMultiInstance] = newval
	}

	if newval := ed.Name.Value().String(); newval != ed.pack.Title {
		edits[packages.AppInfoTitle] = newval
	}

	if newval := ed.Icon.Value().String(); newval != ed.pack.Icon {
		edits[packages.AppInfoIcon] = newval
	}
	return edits
}
