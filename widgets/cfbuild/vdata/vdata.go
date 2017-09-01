// Package vdata provides a virtual data source for the config file builder.
//
// This data source can only be used for tests.
//
package vdata

import (
	"github.com/gotk3/gotk3/gtk"

	"github.com/sqp/godock/libs/cdglobal" // Dock types.
	"github.com/sqp/godock/libs/cdtype"   // Logger type.

	"github.com/sqp/godock/widgets/cfbuild/cftype"   // Types for config file builder usage.
	"github.com/sqp/godock/widgets/cfbuild/datatype" // Types for config file builder data source.
	"github.com/sqp/godock/widgets/gtk/keyfile"
)

//
//-----------------------------------------------------------------[ BUILDER ]--

// Sourcer extends the config source with the virtual source own methods.
//
type Sourcer interface {
	cftype.Source
	ClickedSave()
	ClickedQuit()
	SetBox(*gtk.Box)
	SetGrouper(cftype.Grouper)
	Select(string, ...string) bool
}

// New creates a virtual data source for the config file builder.
//
func New(log cdtype.Logger, win cftype.WinLike, saveCall func(cftype.Builder)) Sourcer {
	return &source{
		win:          win,
		saveCall:     saveCall,
		SourceCommon: datatype.SourceCommon{Log: log},
	}
}

type source struct {
	cftype.Grouper        // extends the grouper.
	datatype.SourceCommon // and the data source.

	box *gtk.Box       // but this is the main box.
	win cftype.WinLike // pointer to the parent window.

	saveCall func(cftype.Builder) // save button callback.
}

func (o *source) GetWindow() cftype.WinLike        { return o.win }
func (o *source) GrabWindowClass() (string, error) { return "windowclass", nil }
func (o *source) ClickedSave()                     { o.saveCall(o.Grouper) }
func (o *source) ClickedQuit()                     { gtk.MainQuit() }
func (o *source) SetBox(box *gtk.Box)              { o.box = box }
func (o *source) SetGrouper(g cftype.Grouper)      { o.Grouper = g }
func (o *source) Select(string, ...string) bool    { return false }
func (o *source) DecryptString(str string) string  { return str }
func (o *source) EncryptString(str string) string  { return str }

//
//-------------------------------------------------------------[ DATA SOURCE ]--

func (o *source) ListIconsMainDock() (list []datatype.Field)        { return IconsMainDock }
func (o *source) ListDocks(parent, subdock string) []datatype.Field { return Maindocks }
func (o *source) ListViews() map[string]datatype.Handbooker         { return Views }

func (o *source) Handbook(name string) datatype.Handbooker {
	for _, book := range Handbooks {
		if name == book.GetName() {
			return book
		}
	}
	return Handbooks[0]
}

func (o *source) DisplayMode() cdglobal.DisplayMode { return cdglobal.DisplayModeAll }

func (o *source) CreateMainDock() string { return "_MainDock_-2.conf" }

func (o *source) MainConfigFile() string                                                   { return "" }
func (o *source) MainConfigDefault() string                                                { return "" }
func (o *source) AppIcon() string                                                          { return "" }
func (o *source) DirShareData(path ...string) string                                       { return "" }
func (o *source) DirUserAppData(path ...string) (string, error)                            { return "", nil }
func (o *source) ListKnownApplets() map[string]datatype.Appleter                           { return nil }
func (o *source) ListDownloadApplets() (map[string]datatype.Appleter, error)               { return nil, nil }
func (o *source) ListIcons() *datatype.ListIcon                                            { return nil }
func (o *source) ListShortkeys() (list []cdglobal.Shortkeyer)                              { return nil }
func (o *source) ListScreens() (list []datatype.Field)                                     { return nil }
func (o *source) ListDialogDecorator() (list []datatype.Field)                             { return nil }
func (o *source) ListAnimations() (list []datatype.Field)                                  { return nil }
func (o *source) ListDeskletDecorations() (list []datatype.Field)                          { return nil }
func (o *source) ListDockThemeLoad() (map[string]datatype.Appleter, error)                 { return nil, nil }
func (o *source) ListDockThemeSave() []datatype.Field                                      { return nil }
func (o *source) CurrentThemeLoad(themeName string, useBehaviour, useLaunchers bool) error { return nil }
func (o *source) CurrentThemeSave(thm string, savBehav, savL, needP bool, dirP string) error {
	return nil
}

func (o *source) ManagerReload(name string, b bool, keyf *keyfile.KeyFile) {}
func (o *source) DesktopClasser(class string) datatype.DesktopClasser      { return nil }

// return desktopclass.Info(class)

// IconsMainDock defines the list of Icons of the virtual dock.
var IconsMainDock = []datatype.Field{{
	Key:  "/path/to/Audio.conf",
	Name: "Audio",
	Icon: "media-optical",
}, {
	Key:  "/path/to/Cpu.conf",
	Name: "Gnome Terminal",
	Icon: "system-run",
}, {
	Key:  "/path/to/NetActivity.conf",
	Name: "NetActivity",
	Icon: "network-workgroup",
}, {
	Key:  "/path/to/TVPlay.conf",
	Name: "TVPlay",
	Icon: "media-playback-pause",
}}

// Maindocks defines the list of maindocks for the virtual dock.
var Maindocks = []datatype.Field{
	{Key: datatype.KeyMainDock, Name: "Main dock"},
}

// Handbooks defines the list of applet handbooks for the virtual dock.
var Handbooks = []datatype.Handbooker{
	&datatype.HandbookSimple{
		Key:         "AppletName",
		Title:       "Handbook (A)",
		Description: "this is \nsome\ndescription text to explain the goal of the applet",
		Preview:     "/usr/share/cairo-dock/images/cairo-dock-logo.png",
		Author:      "Author Name",
	}, &datatype.HandbookSimple{
		Key:         "AnotherApplet",
		Title:       "Another Applet",
		Description: "with some other explanation",
		Preview:     "/usr/share/cairo-dock/images/preview-default.png",
		Author:      "Someone Else",
	}}

// Views defines the list of dock views for the virtual dock.
var Views = map[string]datatype.Handbooker{
	"ViewOne": &datatype.HandbookSimple{
		Key:         "ViewOne",
		Title:       "List view (A)",
		Description: "this is \nsome\ndescription text to explain how the view works",
		Preview:     "/usr/share/cairo-dock/images/cairo-dock-logo.png",
		Author:      "Author Name",
	},
	"ViewTwo": &datatype.HandbookSimple{
		Key:         "ViewTwo",
		Title:       "Another View",
		Description: "with some other explanation",
		Preview:     "/usr/share/cairo-dock/images/preview-default.png",
		Author:      "Someone Else",
	}}
