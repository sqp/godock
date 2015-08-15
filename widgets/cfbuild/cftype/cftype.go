package cftype

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"github.com/sqp/godock/libs/cdtype" // Logger type.
	"github.com/sqp/godock/libs/helper/cast"
	"github.com/sqp/godock/libs/ternary"

	"github.com/sqp/godock/widgets/cfbuild/datatype"
	"github.com/sqp/godock/widgets/cfbuild/valuer" // Converts interface value.
	"github.com/sqp/godock/widgets/pageswitch"
)

// DesktopEntry defines the group name for launchers.
const DesktopEntry = "Desktop Entry"

// Display constants.
const (
	MarginGUI  = 4
	MarginIcon = 6

	// GTK_ICON_SIZE_MENU         = 16
	// CAIRO_DOCK_TAB_ICON_SIZE   = 24 // 32
	// CAIRO_DOCK_FRAME_ICON_SIZE = 24

	DefaultTextColor = 0.6 // light grey
)

// Dock icon types.
const (
	UserIconLauncher = iota
	UserIconStack
	UserIconSeparator
)

//
//----------------------------------------------------------------[ KEY TYPE ]--

// Modifier to show a widget according to the display backend.
const (
	FlagCairoOnly  = '*'
	FlagOpenGLOnly = '&'
)

// KeyType defines the type for a key and its widget.
//
type KeyType byte

// Dock buildable widgets list. Unused (HJQRqxz)
const (
	// Display.

	KeyTextLabel   KeyType = '>' // a simple text label.
	KeyLink        KeyType = 'W' // a simple text label.
	KeySeparator   KeyType = 'v' // an horizontal separator.
	KeyFrame       KeyType = 'F' // a frame. The previous frame will be closed.
	KeyExpander    KeyType = 'X' // a frame inside an expander. The previous frame will be closed.
	KeyEmptyWidget KeyType = '_' // an empty GtkContainer, in case you need to build custom widgets.
	KeyEmptyFull   KeyType = '<' // an empty GtkContainer, the same but using full available space.

	// Basic types.

	KeyBoolButton        KeyType = 'b' // boolean in a button to tick.
	KeyBoolCtrl          KeyType = 'B' // boolean in a button to tick, that will control the sensitivity of the next widget.
	KeyIntSpin           KeyType = 'i' // integer in a spin button.
	KeyIntScale          KeyType = 'I' // integer in an horizontal scale.
	KeyIntSize           KeyType = 'j' // pair of integer spin for size WidthxHeight.
	KeyFloatSpin         KeyType = 'f' // double in a spin button.
	KeyFloatScale        KeyType = 'e' // double in an horizontal scale.
	KeyColorSelectorRGB  KeyType = 'c' // 3 doubles with a color selector (RGB).
	KeyColorSelectorRGBA KeyType = 'C' // 4 doubles with a color selector (RGBA).

	// Strings.

	KeyStringEntry      KeyType = 's' // a text entry.
	KeyPasswordEntry    KeyType = 'p' // a text entry, where text is hidden and the result is encrypted in the .conf file.
	KeyFileSelector     KeyType = 'S' // a text entry with a file selector.
	KeyImageSelector    KeyType = 'g' // a text entry with a file selector, files are filtered to only display images.
	KeyFolderSelector   KeyType = 'D' // a text entry with a folder selector.
	KeySoundSelector    KeyType = 'u' // a text entry with a file selector and a 'play' button, for sound files.
	KeyShortkeySelector KeyType = 'k' // a text entry with a shortkey selector.
	KeyClassSelector    KeyType = 'K' // a text entry with a class selector.
	KeyFontSelector     KeyType = 'P' // a font selector button.

	// Multi strings.

	KeyListSimple          KeyType = 'L' // a text list.
	KeyListEntry           KeyType = 'E' // a combo-entry, that is to say a list where one can add a custom choice.
	KeyListNumbered        KeyType = 'l' // a combo where the number of the line is used for the choice.
	KeyListNbCtrlSimple    KeyType = 'y' // a combo where the number of the line is used for the choice, and for controlling the sensitivity of the widgets below.
	KeyListNbCtrlSelect    KeyType = 'Y' // a combo where the number of the line is used for the choice, and for controlling the sensitivity of the widgets below; controlled widgets are indicated in the list : {entry;index first widget;nb widgets}.
	KeyTreeViewSortSimple  KeyType = 'T' // a tree view, where lines are numbered and can be moved up and down.
	KeyTreeViewSortModify  KeyType = 'U' // a tree view, where lines can be added, removed, and moved up and down.
	KeyTreeViewMultiChoice KeyType = 'V' // a tree view, where lines are numbered and can be selected or not.

	KeyLaunchCmdSimple KeyType = 'Z' // a button to launch a specific command.
	KeyLaunchCmdIf     KeyType = 'G' // a button to launch a specific command with a condition.

	// Dock.

	KeyListViews              KeyType = 'n' // list of available views.
	KeyListAnimation          KeyType = 'a' // list of available animations.
	KeyListDialogDecorator    KeyType = 't' // list of available dialog decorators.
	KeyListDeskletDecoSimple  KeyType = 'O' // list of available desklet decorations.
	KeyListDeskletDecoDefault KeyType = 'o' // same but with the 'default' choice too.

	KeyListThemeApplet      KeyType = 'h' // list of themes in a combo, with preview and readme (gauges, sound...).
	KeyListThemeDesktopIcon KeyType = 'w' // list of installed icon themes.

	KeyListDocks         KeyType = 'd' // list of existing docks.
	KeyListIconsMainDock KeyType = 'N' // list of icons in the maindock.
	KeyListScreens       KeyType = 'r' // list of screens

	KeyHandbook KeyType = 'A' // a label containing the handbook of the applet.

	KeyJumpToModuleSimple   KeyType = 'm' // a button to jump to another module inside the config panel.
	KeyJumpToModuleIfExists KeyType = 'M' // same but only if the module exists.
)

var keyName = map[KeyType]string{

	KeyTextLabel:   "Text Label",
	KeyLink:        "Link",
	KeySeparator:   "Separator",
	KeyFrame:       "Frame",
	KeyExpander:    "Expander",
	KeyEmptyWidget: "Empty Widget",
	KeyEmptyFull:   "Empty Widget (expand)",

	KeyBoolButton:        "Bool Button",
	KeyBoolCtrl:          "Bool Ctrl",
	KeyIntSpin:           "Int Spin",
	KeyIntScale:          "Int Scale",
	KeyIntSize:           "Int Size",
	KeyFloatSpin:         "Float Spin",
	KeyFloatScale:        "Float Scale",
	KeyColorSelectorRGB:  "ColorSelector RGB",
	KeyColorSelectorRGBA: "ColorSelector RGBA",

	KeyStringEntry:      "String Entry",
	KeyPasswordEntry:    "Password Entry",
	KeyFileSelector:     "File Selector",
	KeyImageSelector:    "Image Selector",
	KeyFolderSelector:   "Folder Selector",
	KeySoundSelector:    "Sound Selector",
	KeyShortkeySelector: "Shortkey Selector",
	KeyClassSelector:    "Class Selector",
	KeyFontSelector:     "Font Selector",

	KeyListSimple:          "List Simple",
	KeyListEntry:           "List Entry",
	KeyListNumbered:        "List Numbered",
	KeyListNbCtrlSimple:    "List Nb Ctrl Simple",
	KeyListNbCtrlSelect:    "List Nb Ctrl Select",
	KeyTreeViewSortSimple:  "TreeView SortSimple",
	KeyTreeViewSortModify:  "TreeView SortModify",
	KeyTreeViewMultiChoice: "TreeView MultiCheck",

	KeyLaunchCmdSimple: "Launch Command",
	KeyLaunchCmdIf:     "Launch Command If",

	// ...

	KeyListDocks:            "List Docks",
	KeyListIconsMainDock:    "List Icons MainDock",
	KeyListThemeDesktopIcon: "List Theme Icons",
	KeyListScreens:          "List Screens",

	KeyListViews:              "List Views", // Strings
	KeyListThemeApplet:        "List Theme Applet",
	KeyListAnimation:          "List Animation",
	KeyListDialogDecorator:    "List DialogDecorator ",
	KeyListDeskletDecoSimple:  "List DeskletDeco",
	KeyListDeskletDecoDefault: "List DeskletDeco +Def",

	KeyJumpToModuleSimple:   "Jump To Module Simple",
	KeyJumpToModuleIfExists: "Jump To Module If Exists",

	KeyHandbook: "Handbook",
}

// String returns the type as string.
//
func (typ KeyType) String() string {
	str, ok := keyName[typ]
	if ok {
		return str
	}
	return string([]byte{byte(typ)})
}

// New creates a key of the given type.
//
func (typ KeyType) New(build Builder, group, name string) *Key {
	return &Key{
		Builder: build,
		Group:   group,
		Name:    name,
		Type:    typ,
	}
}

//
//-----------------------------------------------------------[ CONFIG SOURCE ]--

// Grouper builds config pages from the Builder.
//
// Custom tweaks can be applied on loaded keys before the build.
//
// When adding keys manually, ensure you use AddGroup/AddKeys,
// or set the storage access fields if you need widgets that use them.
//
// cftype.Key....NewKey(build, group, name, label)
// build.NewKey...(group, name, ...)
//
type Grouper interface {
	Builder // Extends the Builder.

	// BuildSingle builds a single page config for the given group.
	//
	BuildSingle(group string, tweaks ...func(Builder)) Grouper

	// BuildAll builds a dock configuration widget with all groups.
	//
	BuildAll(switcher *pageswitch.Switcher, tweaks ...func(Builder)) Grouper

	// BuildGroups builds a dock configuration widget with the given groups.
	//
	BuildGroups(switcher *pageswitch.Switcher, groups []string, tweaks ...func(Builder)) Grouper
}

// Builder builds a Cairo-Dock configuration page.
//
type Builder interface {
	//-------------------------------------------------[ INTERNAL INTERFACES ]--
	//
	// Storage gives access to the builder storage.
	//
	Storage() Storage

	// Storage gives access to the dock source.
	//
	Source() Source

	// Log gives access to the builder logger.
	//
	Log() cdtype.Logger

	//--------------------------------------------------------------[ COMMON ]--
	//
	// Save updates the configuration file with user changes.
	//
	Save()

	// SetPostSave sets a post save call, triggered after successful real saves.
	//
	SetPostSave(func())

	// Translate translates the given string using the builder domain.
	//
	Translate(str string) string

	// Free frees the builder internal pointers (cyclic references).
	//
	// Already called by the Grouper, but mandatory for other customs.
	//
	Free()

	//---------------------------------------------------------------[ BUILD ]--
	//
	// AddGroup adds a group with optional keys.
	//
	AddGroup(group string, keys ...*Key)

	// AddKeys adds one or many keys to an existing group.
	//
	AddKeys(group string, keys ...*Key)

	// Groups lists the configured build groups.
	//
	Groups() []string

	// BuildPage builds a Cairo-Dock configuration page for the given group.
	//
	BuildPage(group string) GtkWidgetBase

	//-----------------------------------------------------[ KEYS INTERACTION]--
	//
	// KeyAction acts on a key if found. Key access errors will just be logged.
	//
	KeyAction(group, name string, action func(*Key)) bool

	// KeyWalk runs the given call on all keys in the group and key build order.
	//
	KeyWalk(call func(*Key))

	// KeyBool returns the key value as boolean.
	//
	KeyBool(group, name string) (val bool)

	// KeyInt returns the key value as int.
	//
	KeyInt(group, name string) (val int)

	// KeyFloat returns the key value as float64.
	//
	KeyFloat(group, name string) (val float64)

	// KeyString returns the key value as string.
	//
	KeyString(group, name string) (val string)

	//------------------------------------------------------[ WIDGET PACKING ]--
	//--- Widget packing for custom and internal widgets use.
	//
	// PackWidget packs a widget in the page main box.
	//
	PackWidget(child gtk.IWidget, expand, fill bool, padding uint)

	// PackSubWidget packs a widget in the current subwidget box.
	//
	PackSubWidget(child gtk.IWidget)
	// (was _pack_in_widget_box)

	// PackKeyWidget sets get/set value calls for the widget and can pack widgets.
	//
	PackKeyWidget(key *Key, getValue func() interface{}, setValue func(interface{}), child ...gtk.IWidget)
	// (was _pack_subwidget).

	//------------------------------------------------------[ BOX MANAGEMENT ]--
	//
	Label() *gtk.Label

	BoxPage() *gtk.Box

	SetFrame(*gtk.Frame)

	SetFrameBox(*gtk.Box)

	SetNbControlled(nb int)

	//-----------------------------------------------------[ FROM GTK WIDGET ]--
	//
	GtkWidgetBase
	Connect(string, interface{}, ...interface{}) (glib.SignalHandle, error)
	Remove(gtk.IWidget)
	PackStart(gtk.IWidget, bool, bool, uint)
}

// Storage defines the storage access format.
//
type Storage interface {
	SetBuilder(build Builder)
	FilePath() string
	FileDefault() string

	GetGroups() (uint64, []string)
	List(group string) []*Key

	ToData() (uint64, string, error)

	// Get data.
	//
	Valuer(group, name string) valuer.Valuer           // Current value.
	Default(group, name string) (valuer.Valuer, error) // Default value.

	Get(group, key string, value interface{}) error
	Int(group, key string) (int, error)
	Bool(group, key string) (bool, error)
	Float(group, key string) (float64, error)
	String(group, key string) (string, error)
	ListInt(group, key string) (list []int, e error)
	ListBool(group, key string) (list []bool, e error)
	ListFloat(group, key string) (list []float64, e error)
	ListString(group, key string) (list []string, e error)

	// Set data. Only one method.
	//
	Set(group, key string, value interface{}) error
}

// Source extends the data source with GetWindow needed to use the builder.
//
type Source interface {
	datatype.Source
	GetWindow() *gtk.Window
}

// GtkWidgetBase extends the gtk.IWidget interface with widgets methods.
//
type GtkWidgetBase interface {
	gtk.IWidget
	ShowAll()
	Destroy()
}

//
//---------------------------------------------------------------------[ KEY ]--

type keyDataModel interface {
	UpdateRowKey(oldkey, newkey string)
}

// Key defines a configuration entry.
//
type Key struct {
	Builder // extend the builder for widget building needs.

	Type              KeyType  // Type of key, for the value type, build method and options.
	Group             string   // Group for the key. Match the config group and switcher page.
	Name              string   // Name for the key. Match the config name.
	NbElements        int      // number of values stored in the key.
	AuthorizedValues  []string //
	Text              string   // label for the key.
	Tooltip           string   // mouse over tooltip text.
	IsAlignedVertical bool     // orientation for the key widget box.
	IsDefault         bool     // true when a default text has been set (must be ignored). Match "ignore-value" in the C version.

	dataModel   keyDataModel       // act on the gtk data model.
	widgetValue func() interface{} // get values from the widget.
	widsetValue func(interface{})  // set values to the widget.
	makeWidget  func(*Key)         // Custom widget builder, overriding the default for key type.
}

// IsType returns whether the key type is one of the provided types.
//
func (key *Key) IsType(types ...KeyType) bool {
	for _, test := range types {
		if key.Type == test {
			return true
		}
	}
	return false
}

// SetBuilder sets the key builder. Mandatory if not using AddGroup or AddKeys.
//
func (key *Key) SetBuilder(build Builder) *Key {
	key.Builder = build
	if key.NbElements == 0 {
		key.NbElements = 1
	}
	return key
}

// MakeWidget returns the custom make widget call for the key if set.
//
func (key *Key) MakeWidget() func() {
	if key.makeWidget == nil {
		return nil
	}
	return func() { key.makeWidget(key) }
}

// SetMakeWidget sets a custom MakeWidget call for the key.
//
func (key *Key) SetMakeWidget(call func(*Key)) {
	key.makeWidget = call
}

// SetWidGetValue sets the get value to the widget callback.
//
func (key *Key) SetWidGetValue(getValue func() interface{}) {
	key.widgetValue = getValue
}

// SetWidSetValue sets the set value to the widget callback.
//
func (key *Key) SetWidSetValue(setValue func(interface{})) {
	key.widsetValue = setValue
}

//
//------------------------------------------------------------------[ VALUES ]--

// Value returns an interface to the key value.
//
// Before the build, the storage value will be used.
// After the build, the widget value will be used.
//
func (key *Key) Value() valuer.Valuer {
	if key.widgetValue != nil { // After the build.
		val := key.widgetValue()

		return valuer.New(&val) // Need better valuer.
	}
	return key.Storage().Valuer(key.Group, key.Name)
}

// ValueGet gets the key value.
//
// A pointer to the value must be used to allow the value to be assigned.
//
// Before the build, the storage value will be used.
// After the build, the widget value will be used.
//
func (key *Key) ValueGet(val interface{}) error {

	if key.widgetValue != nil { // After the build.
		switch ptr := val.(type) {
		case *bool:
			*ptr = (key.widgetValue()).(bool)

		case *int:
			*ptr = (key.widgetValue()).(int)

		case *float64:
			*ptr = (key.widgetValue()).(float64)

		case *string:
			*ptr = (key.widgetValue()).(string)

		case *[]bool:
			*ptr = (key.widgetValue()).([]bool)

		case *[]int:
			*ptr = (key.widgetValue()).([]int)

		case *[]float64:
			*ptr = (key.widgetValue()).([]float64)

		case *[]string:
			*ptr = (key.widgetValue()).([]string)

		// case nil: // TODO: test.
		// 	*ptr = key.widgetValue()

		default:
			key.Log().NewErr("bad type", "key ValueGet", key.Group, key.Name)
			return nil // TODO: return error.
		}

		return nil
	}

	// Before the build, return the storage value.
	return key.Storage().Get(key.Group, key.Name, val)
}

// ValueSet sets the key value.
//
// Before the build, the value will be set to the storage.
// After the build, the value will be set to the widget.
//
// Values will be copied from the widget to the storage with UpdateStorage.
// The file will only be updated with build.Save (which calls UpdateStorage).
//
func (key *Key) ValueSet(val interface{}) error {
	if key.widsetValue != nil { // After the build.
		key.widsetValue(val)
		return nil
	}
	// Before the build, or a static widget.
	return key.Storage().Set(key.Group, key.Name, val)
}

//
//-------------------------------------------------------------[ VALUE STATE ]--

// ValueState defines the update status of the key values.
type ValueState int

// List of value status.
const (
	StateBothEmpty ValueState = iota
	StateAdded                // Value added to new.
	StateRemoved              // Value removed from new.
	StateUnchanged            // Same values.
	StateEdited               // Different values.
)

// ValueStateField represents a value comparison for the field.
//
type ValueStateField struct {
	State ValueState
	Old   string
	New   string
}

// ValueStateList defines a list of ValueStateField.
//
type ValueStateList []ValueStateField

// IsChanged returns whether the value has changed.
//
func (list ValueStateList) IsChanged() bool {
	for _, st := range list {
		switch st.State {
		case StateAdded, StateRemoved, StateEdited:
			return true
		}
	}
	return false
}

// ValueState starts the comparison between old and new values.
// Only compares the presence or absence of values, not the content.
//
func (key *Key) ValueState(previous valuer.Valuer) ValueStateList {
	sizeOld := previous.Count()
	sizeNew := key.Value().Count()

	if sizeOld == 0 && sizeNew == 0 {
		return ValueStateList{{State: StateBothEmpty}}
	}

	count := ternary.Max(sizeOld, sizeNew)
	if count > key.NbElements && !key.IsType(KeyTreeViewSortSimple, KeyTreeViewSortModify, KeyTreeViewMultiChoice) {
		println("key.ValueState higher dropped", key.Type.String(), key.Name, ":", count, "/", key.NbElements)
		return ValueStateList{valueCompare(previous.Sprint(), key.Value().Sprint())}
	}

	flags := make(ValueStateList, count)

	for i := range flags {
		switch {
		case sizeOld <= i: // Added fields at the end.
			flags[i] = ValueStateField{
				State: StateAdded,
				New:   key.Value().SprintI(i),
			}

		case sizeNew <= i: // Same with removed.
			flags[i] = ValueStateField{
				State: StateRemoved,
				Old:   previous.SprintI(i),
			}

		default: // Values in both lists, pack first, and leave untested here.
			flags[i] = valueCompare(previous.SprintI(i), key.Value().SprintI(i))
		}
	}

	return flags
}

func valueCompare(strold, strnew string) ValueStateField {
	return ValueStateField{
		State: StateUnchanged + ValueState(cast.BoolToInt(strold != strnew)),
		Old:   strold,
		New:   strnew,
	}
}

//
//--------------------------------------------------------------[ KEY UPDATE ]--

// UpdateStorage updates the storage with values from the widget.
//
func (key *Key) UpdateStorage() {
	switch key.Type {
	case KeyBoolButton, // boolean
		KeyBoolCtrl, // boolean qui controle le widget suivant

		KeyIntSpin,    // integer
		KeyIntScale,   // integer in a HScale
		KeyIntSize,    // double integer WxH
		KeyFloatSpin,  // float.
		KeyFloatScale: // float in a HScale.

		key.Storage().Set(key.Group, key.Name, key.widgetValue())

	case KeyColorSelectorRGB, // float x3 avec un bouton de choix de couleur.
		KeyColorSelectorRGBA: // float x4 avec un bouton de choix de couleur.
		// // value := key.GetValues[0]().(*gdk.RGBA)
		// // floats := value.Floats()
		// floats := key.Value().ListFloat()
		// // if key.IsType(KeyColorSelectorRGB) && len(floats) > 3 { // need only 3 values when no alpha.
		// // 	floats = floats[:3]
		// // }
		key.Storage().Set(key.Group, key.Name, key.widgetValue())

	case KeyListNumbered, // a list of numbered strings.
		KeyListNbCtrlSimple, // a list of numbered strings whose current choice defines the sensitivity of the widgets below.
		KeyListNbCtrlSelect: // a list of numbered strings whose current choice defines the sensitivity of the widgets below given in the list.
		key.Storage().Set(key.Group, key.Name, key.widgetValue())

	case KeyTreeViewSortSimple, // N strings listed from top to bottom.
		KeyTreeViewSortModify,  // same with possibility to add/remove some.
		KeyTreeViewMultiChoice: // N strings that can be selected or not.
		key.Storage().Set(key.Group, key.Name, key.widgetValue())
		// value := key.GetValues[0]().([]string)
		// log.DEV("TREEVIEW", key.Name, value)
		// if len(value) > 1 {
		// 	key.Storage().Set(key.Group, key.Name, value)
		// } else if len(value) == 1 {
		// 	key.Storage().Set(key.Group, key.Name, value[0])
		// }

	case KeyStringEntry, KeyPasswordEntry,
		KeyFileSelector, KeyFolderSelector, // selectors.
		KeySoundSelector, KeyShortkeySelector,
		KeyClassSelector, KeyFontSelector,
		KeyImageSelector,

		KeyListSimple, KeyListEntry, // a list of strings.
		KeyListDeskletDecoSimple,  // desklet decorations list.
		KeyListDeskletDecoDefault, // idem mais avec le choix "defaut" en plus.
		KeyListIconsMainDock,      // main dock icons list.

		KeyListThemeApplet,             // themes list in a combo, with preview and readme.
		KeyListViews, KeyListAnimation, // other filled lists.
		KeyListDialogDecorator, KeyListThemeDesktopIcon, // ...
		KeyListScreens:

		value := key.Value().String()

		if key.IsType(KeyPasswordEntry) {
			// TODO: cairo_dock_encrypt_string(value, &value)
		}

		key.Storage().Set(key.Group, key.Name, value)

		//

	case KeyListDocks:
		// Only get the real dock name on update as it's only needed for the save.
		value := testNewMainDock(key)
		key.Storage().Set(key.Group, key.Name, value)

		//

		// shouldn't be saved, need to check.

	// case WidgetJumpToModuleSimple, // bouton raccourci vers un autre module
	// 	WidgetJumpToModuleIfExists: // idem mais seulement affiche si le module existe.

	// case WidgetLaunchCommandSimple,
	// 	WidgetLaunchCommandIfCondition:

	case KeyLink: // both fields that could now be updated as they have a get/set.
	case KeyHandbook:

	case KeyTextLabel: // Just the text label.
	case KeySeparator:
	case KeyFrame, KeyExpander:
	case KeyEmptyWidget, KeyEmptyFull: // Containers for custom widget.

	default:
		// for i, f := range key.GetValues {
		// 	key.Log().Info("KEY NOT MATCHED", key.Name, i+1, "/", len(key.GetValues), "[", f(), "]")
		// }
	}
}

//
//------------------------------------------------------------[ NEW MAINDOCK ]--

// testNewMainDock creates a maindock to hold the icon if it was moved (dock not found).
//
func testNewMainDock(key *Key) string {
	value := key.Value().String()
	if value != datatype.KeyNewDock {
		return value
	}

	println("UPDATE DOCK", value)

	// Reset the value to storage value.
	// It will be only updated if a new dock is really to create.
	newvalue, e := key.Storage().String(key.Group, key.Name)
	key.Log().Err(e, "testNewMainDock get storage", key.Name)

	//
	testNewDock := func(test func(key *Key) bool) func(key *Key) {
		return func(key *Key) {
			if test != nil && !test(key) {
				old, _ := key.Builder.Storage().String("Icon", "dock name")
				key.Log().DEV("detached, TODO: need to restore dock key", old)
				return
			}

			// Just update the real name of the new dock.
			// This would prevent the option to readd a new dock (not useful twice).
			// But as we're moving an icon, the whole list and the current page will
			// be reloaded after the save.

			newvalue = key.Source().CreateMainDock()
			// key.ModelUpdateRowKey(datatype.KeyNewDock, newname)
		}
	}

	if !key.KeyAction("Icon", "dock name", testNewDock(testAppletNotDetached)) {

		// Create new dock if needed for other icons.
		if !key.KeyAction(DesktopEntry, "Container", testNewDock(nil)) {
			println("problem with newdock, didn't match an applet nor a simple icon ?")
		}
	}
	return newvalue
}

// testAppletNotDetached tests the detached state to know if we need a new dock.
//
func testAppletNotDetached(key *Key) bool {
	return !key.KeyBool("Desklet", "initially detached")
}

//
//------------------------------------------------------------[ BASE STORAGE ]--

// BaseStorage provides a common base for Storage.
//
type BaseStorage struct {
	File    string // path to active config file.
	Default string // path to default config file.

	Build Builder
}

// SetBuilder sets the storage builder.
//
func (conf *BaseStorage) SetBuilder(build Builder) {
	conf.Build = build
}

// FilePath returns the path to the config file.
//
func (conf *BaseStorage) FilePath() string {
	return conf.File
}

// FileDefault returns the path to the config file.
//
func (conf *BaseStorage) FileDefault() string {
	return conf.Default
}
