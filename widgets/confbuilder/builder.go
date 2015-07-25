package confbuilder

/*
	cGroupName = pGroupList[i];

	//\____________ On recupere les caracteristiques du groupe.
	cGroupComment = g_key_file_get_comment (pKeyFile, cGroupName, NULL, NULL);
	cIcon = NULL;
	cDisplayedGroupName = NULL;
	if (cGroupComment != NULL && *cGroupComment != '\0')  // extract the icon name/path, inside brackets [].
	{
		gchar *str = strrchr (cGroupComment, '[');
		if (str != NULL)
		{
			cIcon = str+1;
			str = strrchr (cIcon, ']');
			if (str != NULL)
				*str = '\0';
			str = strrchr (cIcon, ';');
			if (str != NULL)
			{
				*str = '\0';
				cDisplayedGroupName = str + 1;
			}
		}
	}
*/

import (
	"github.com/conformal/gotk3/gdk"
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/config"
	"github.com/sqp/godock/libs/text/tran"

	"github.com/sqp/godock/widgets/confbuilder/datatype"
	"github.com/sqp/godock/widgets/confsettings"

	"errors"
	"strings"
)

// Dock constants.
const (
	MarginGUI      = 4
	MarginIcon     = 6
	PreviewSizeMax = 200

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

// Modifier to show a widget according to the display backend.
const (
	WidgetCairoOnly  = '*'
	WidgetOpenGLOnly = '&'
)

//
//-------------------------------------------------------------[ WIDGET TYPE ]--

// WidgetType defines the type for a key and its widget.
//
type WidgetType byte

// Dock buildable widgets list.
const (
	WidgetCheckButton        WidgetType = 'b' // boolean in a button to tick.
	WidgetCheckControlButton            = 'B' // boolean in a button to tick, that will control the sensitivity of the next widget.
	WidgetIntegerSpin                   = 'i' // integer in a spin button.
	WidgetIntegerScale                  = 'I' // integer in an horizontal scale.
	WidgetIntegerSize                   = 'j' // pair of integers for dimansion WidthxHeight
	WidgetFloatSpin                     = 'f' // double in a spin button.
	WidgetFloatScale                    = 'e' // double in an horizontal scale.
	WidgetColorSelectorRGB              = 'c' // 3 doubles with a color selector (RGB).
	WidgetColorSelectorRGBA             = 'C' // 4 doubles with a color selector (RGBA).

	WidgetViewList                         = 'n' // list of views.
	WidgetThemeList                        = 'h' // list of themes in a combo, with preview and readme.
	WidgetAnimationList                    = 'a' // list of available animations.
	WidgetDialogDecoratorList              = 't' // list of available dialog decorators.
	WidgetDeskletDecorationListSimple      = 'O' // list of available desklet decorations.
	WidgetDeskletDecorationListWithDefault = 'o' // same but with the 'default' choice too.

	WidgetListDocks     = 'd' // list of existing docks.
	WidgetIconsList     = 'N' // list of icons of a dock.
	WidgetIconThemeList = 'w' // list of installed icon themes.
	WidgetScreensList   = 'r' // list of screens

	WidgetJumpToModuleSimple   = 'm' // a button to jump to another module inside the config panel.
	WidgetJumpToModuleIfExists = 'M' // same but only if the module exists.

	WidgetLaunchCommandSimple      = 'Z' // a button to launch a specific command.
	WidgetLaunchCommandIfCondition = 'G' // a button to launch a specific command with a condition.

	WidgetStringEntry      = 's' // a text entry.
	WidgetFileSelector     = 'S' // a text entry with a file selector.
	WidgetImageSelector    = 'g' // a text entry with a file selector, files are filtered to only display images.
	WidgetFolderSelector   = 'D' // a text entry with a folder selector.
	WidgetSoundSelector    = 'u' // a text entry with a file selector and a 'play' button, for sound files.
	WidgetShortkeySelector = 'k' // a text entry with a shortkey selector.
	WidgetClassSelector    = 'K' // a text entry with a class selector.
	WidgetPasswordEntry    = 'p' // a text entry, where text is hidden and the result is encrypted in the .conf file.
	WidgetFontSelector     = 'P' // a font selector button.

	WidgetListSimple                   = 'L' // a text list.
	WidgetListWithEntry                = 'E' // a combo-entry, that is to say a list where one can add a custom choice.
	WidgetNumberedList                 = 'l' // a combo where the number of the line is used for the choice.
	WidgetNumberedControlListSimple    = 'y' // a combo where the number of the line is used for the choice, and for controlling the sensitivity of the widgets below.
	WidgetNumberedControlListSelective = 'Y' // a combo where the number of the line is used for the choice, and for controlling the sensitivity of the widgets below; controlled widgets are indicated in the list : {entry;index first widget;nb widgets}.
	WidgetTreeViewSortSimple           = 'T' // a tree view, where lines are numbered and can be moved up and down.
	WidgetTreeViewSortAndModify        = 'U' // a tree view, where lines can be added, removed, and moved up and down.
	WidgetTreeViewMultiChoice          = 'V' // a tree view, where lines are numbered and can be selected or not.

	WidgetEmptyWidget = '_' // an empty GtkContainer, in case you need to build custom widgets.
	WidgetEmptyFull   = '<' // an empty GtkContainer, the same but using full available space.
	WidgetTextLabel   = '>' // a simple text label.

	WidgetLink      = 'W' // a simple text label.
	WidgetHandbook  = 'A' // a label containing the handbook of the applet.
	WidgetSeparator = 'v' // an horizontal separator.
	WidgetFrame     = 'F' // a frame. The previous frame will be closed.
	WidgetExpander  = 'X' // a frame inside an expander. The previous frame will be closed.
)

// Builder returns the builder call for the widget type.
//
func (typ WidgetType) Builder() func(build *Builder, key *Key) {
	switch typ {
	case WidgetCheckButton, // boolean
		WidgetCheckControlButton: // boolean qui controle le widget suivant
		return (*Builder).WidgetCheckButton

	case WidgetIntegerSpin, // integer in a spin button.
		WidgetIntegerScale, // integer in a HScale.
		WidgetIntegerSize:  // double integer WxH
		return (*Builder).WidgetInteger

	case WidgetFloatSpin, // float.
		WidgetFloatScale: // float in a HScale.
		return (*Builder).WidgetFloat

	case WidgetColorSelectorRGB, // float x3 avec un bouton de choix de couleur.
		WidgetColorSelectorRGBA: // float x4 avec un bouton de choix de couleur.
		return (*Builder).WidgetColorSelector

	case WidgetViewList: // List of dock views.
		return (*Builder).WidgetViewList

	case WidgetThemeList: // List themes in a combo, with preview and readme.
		return (*Builder).WidgetListTheme

	case WidgetAnimationList: // List of animations.
		return (*Builder).WidgetAnimationList

	case WidgetDialogDecoratorList: // liste des decorateurs de dialogue.
		return (*Builder).WidgetDialogDecoratorList

	case WidgetDeskletDecorationListSimple, // liste des decorations de desklet.
		WidgetDeskletDecorationListWithDefault: // idem mais avec le choix "defaut" en plus.
		return (*Builder).WidgetListDeskletDecoration

	case WidgetListDocks: // liste des docks existant.
		return (*Builder).WidgetDockList

	case WidgetIconThemeList:
		return (*Builder).WidgetIconThemeList

	case WidgetIconsList:
		return (*Builder).WidgetIconsList

	case WidgetScreensList:
		return (*Builder).WidgetScreensList

	// case WidgetJumpToModuleSimple, // bouton raccourci vers un autre module
	// 	WidgetJumpToModuleIfExists: // idem mais seulement affiche si le module existe.
	// 	 return (*Builder).WidgetJumpToModule

	case WidgetLaunchCommandSimple,
		WidgetLaunchCommandIfCondition:
		return (*Builder).WidgetLaunchCommand

	case WidgetListSimple, // a list of strings.
		WidgetListWithEntry, // a list of strings with an entry to add custom choices.

		WidgetNumberedList, // a list of numbered strings.
		// 	WidgetNumberedControlListSimple,           // a list of numbered strings whose current choice defines the sensitivity of the widgets below.
		WidgetNumberedControlListSelective: // a list of numbered strings whose current choice defines the sensitivity of the widgets below given in the list.
		return (*Builder).WidgetLists

	case WidgetTreeViewSortSimple, // N strings listed from top to bottom.
		WidgetTreeViewSortAndModify, // same with possibility to add/remove some.
		WidgetTreeViewMultiChoice:   // N strings that can be selected or not.
		return (*Builder).WidgetTreeView

	case WidgetFontSelector: // string avec un selecteur de font a cote du GtkEntry.
		return (*Builder).WidgetFontSelector

	case WidgetLink: // string avec un lien internet a cote.
		return (*Builder).WidgetLink

	case WidgetStringEntry, // string
		WidgetPasswordEntry,    // string de type "password", crypte dans le .build et cache dans l'UI (Merci Tofe !) ,-)
		WidgetFileSelector,     // string avec un selecteur de fichier a cote du GtkEntry.
		WidgetFolderSelector,   // string avec un selecteur de repertoire a cote du GtkEntry.
		WidgetSoundSelector,    // string avec un selecteur de fichier a cote du GtkEntry et un boutton play.
		WidgetShortkeySelector, // string avec un selecteur de touche clavier (Merci Ctaf !)
		WidgetClassSelector,    // string avec un selecteur de class (Merci Matttbe !)
		WidgetImageSelector:    // string with a file selector (results are filtered to only display images)
		return (*Builder).WidgetStrings

	case WidgetEmptyWidget: // Containers for custom widget.
	case WidgetEmptyFull:

	case WidgetTextLabel: // Just the text label.

		// int iFrameWidth = GPOINTER_TO_INT (g_object_get_data (G_OBJECT (pMainWindow), "frame-width"));
		// gtk_widget_set_size_request (pLabel, MIN (800, gldi_desktop_get_width() - iFrameWidth), -1);
		// gtk_label_set_justify (GTK_LABEL (pLabel), GTK_JUSTIFY_LEFT);
		return (*Builder).WidgetTextLabel

	case WidgetHandbook:
		return (*Builder).WidgetHandbook

	case WidgetFrame, WidgetExpander: // frames: simple or with expander.
		return (*Builder).WidgetFrame

	case WidgetSeparator:
		return (*Builder).WidgetSeparator
	}

	return func(build *Builder, _ *Key) {
		build.log.NewWarn("invalid type", "build widget")
	}
}

//
//---------------------------------------------------------------------[ KEY ]--

// Key defines a configuration entry.
//
type Key struct {
	Group             string
	Name              string
	Type              WidgetType
	NbElements        int
	AuthorizedValues  []string
	Text              string
	Tooltip           string
	IsAlignedVertical bool
	IsDefault         bool // true when a default text has been set (must be ignored). Match "ignore-value" in the C version.
	GetValues         []func() interface{}
	makeWidget        func(*Builder, *Key) // custom widget, replaces the use of Type.
}

// IsType returns whether the key type is one of the provided types.
//
func (key *Key) IsType(types ...WidgetType) bool {
	for _, test := range types {
		if key.Type == test {
			return true
		}
	}
	return false
}

// Bool returns the first value as bool.
//
func (key *Key) Bool() (bool, error) {
	if key == nil {
		return false, errors.New("missing key")
	}
	return key.GetValues[0]().(bool), nil
}

// String returns the first value as string.
//
func (key *Key) String() (string, error) {
	if key == nil {
		return "", errors.New("missing key")
	}
	return key.GetValues[0]().(string), nil
}

// SetCustom sets a custom key widget builder call.
//
func (key *Key) SetCustom(custom func(*Builder, *Key)) (*Key, error) {
	if key == nil {
		return nil, errors.New("missing key")
	}
	key.makeWidget = custom
	return key, nil
}

//
//-----------------------------------------------------------------[ BUILDER ]--

// Builder builds a Cairo-Dock configuration page.
//
type Builder struct {
	gtk.Box // Main container.

	pageScroll *gtk.ScrolledWindow // Page container.
	pageBox    *gtk.Box            // Was pGroupBox

	pFrameVBox *gtk.Box // Container for widgets in a frame.

	pKeyBox              *gtk.Box   // Box for the widget
	pLabel               *gtk.Label // Text on left.
	pWidgetBox           *gtk.Box   // Value widgets on the right.
	pAdditionalItemsVBox *gtk.Box

	iNbControlledWidgets int

	// FRAMES ONLY
	pFrame          *gtk.Frame
	pLabelContainer gtk.IWidget

	// Build keys. Using the same group id index (order as detected and used).
	buildGroups []string
	buildKeys   [][]*Key

	// extra data.
	Conf          *CDConfig
	data          datatype.Source
	log           cdtype.Logger
	win           *gtk.Window // Parent window.
	gettextDomain string
	originalConf  string // path to default config file.
}

// GetKey finds the key referenced by its config group and name.
//
func (build *Builder) GetKey(group, name string) *Key {
	for gid, testgroup := range build.buildGroups {
		if group == testgroup {
			for _, key := range build.buildKeys[gid] {
				if key.Name == name {
					return key
				}
			}
		}
	}
	return nil
}

// BuildPage builds a Cairo-Dock configuration page for the given group.
//
func (build *Builder) BuildPage(keys []*Key) *gtk.ScrolledWindow {

	// gconstpointer *data;
	// gsize length = 0;
	// gchar **pKeyList;

	// GtkWidget *pOneWidget;
	// GSList * pSubWidgetList;
	// GtkWidget *pSmallVBox;
	// GtkWidget *pEntry;
	// GtkWidget *pTable;
	// GtkWidget *pButtonAdd, *pButtonRemove;
	// GtkWidget *pButtonDown, *pButtonUp;
	// GtkWidget *pButtonFileChooser, *pButtonPlay;
	// GtkWidget *pScrolledWindow;
	// GtkWidget *pToggleButton=NULL;
	// GtkCellRenderer *rend;
	// GtkTreeIter iter;
	// GtkTreeSelection *selection;
	// GtkWidget *pBackButton;
	// GList *pControlWidgets = NULL;
	// int iFirstSensitiveWidget = 0, iNbSensitiveWidgets = 0;
	// gchar *cKeyName, *cKeyComment, **pAuthorizedValuesList;
	// const gchar *cUsefulComment, *cTipString;
	// CairoDockGroupKeyWidget *pGroupKeyWidget;
	// int j;
	// guint k, iNbElements;
	// char iType;
	// gboolean bValue, *bValueList;
	// int iValue, iMinValue, iMaxValue, *iValueList;
	// double fValue, fMinValue, fMaxValue, *fValueList;
	// gchar *cValue, **cValueList, *cSmallIcon=NULL;
	// GtkListStore *modele;
	// gboolean bAddBackButton;
	// GtkWidget *pPreviewBox;

	build.pageScroll, _ = gtk.ScrolledWindowNew(nil, nil)
	build.pageBox, _ = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, MarginGUI)
	build.pageBox.SetBorderWidth(MarginGUI)
	build.pageScroll.Add(build.pageBox)

	build.pFrameVBox = nil

	build.iNbControlledWidgets = 0

	// FRAMES ONLY
	build.pFrame = nil
	build.pLabelContainer = nil

	for _, key := range keys {
		// log.DEV(key.Name, string(key.Type), key.AuthorizedValues)

		build.pKeyBox = nil
		build.pLabel = nil
		build.pWidgetBox = nil
		build.pAdditionalItemsVBox = nil

		bFullSize := key.IsType(WidgetThemeList, WidgetViewList, WidgetEmptyFull, WidgetHandbook)

		if !key.IsType(WidgetFrame) && !key.IsType(WidgetExpander) && !key.IsType(WidgetSeparator) {
			// Create Key box.
			if key.IsType(WidgetThemeList, WidgetViewList) {
				build.pAdditionalItemsVBox, _ = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
				build.pKeyBox, _ = gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, MarginGUI)
				build.addWidget(build.pAdditionalItemsVBox, bFullSize, bFullSize, 0)
				build.pAdditionalItemsVBox.PackStart(build.pKeyBox, false, false, 0)

			} else {
				if key.IsAlignedVertical {
					build.log.Info("aligned /", strings.TrimSuffix(key.Name, "\n"))
					build.pKeyBox, _ = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, MarginGUI)
				} else {
					build.pKeyBox, _ = gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, MarginGUI)
				}

				build.addWidget(build.pKeyBox, bFullSize, bFullSize, 0)
			}

			if key.Tooltip != "" {
				build.pKeyBox.SetTooltipText(build.translate(key.Tooltip))
			}

			// 	if (pControlWidgets != NULL)
			// 	{
			// 		CDControlWidget *cw = pControlWidgets->data;
			// 		//g_print ("ctrl (%d widgets)\n", iNbControlledWidgets);
			// 		if (cw->pControlContainer == (pFrameVBox ? pFrameVBox : build))
			// 		{
			// 			//g_print ("ctrl (iNbControlledWidgets:%d, iFirstSensitiveWidget:%d, iNbSensitiveWidgets:%d)\n", iNbControlledWidgets, iFirstSensitiveWidget, iNbSensitiveWidgets);
			// 			cw->iNbControlledWidgets --;
			// 			if (cw->iFirstSensitiveWidget > 0)
			// 				cw->iFirstSensitiveWidget --;
			// 			cw->iNonSensitiveWidget --;

			// 			GtkWidget *w = (pAdditionalItemsVBox ? pAdditionalItemsVBox : pKeyBox);
			// 			if (cw->iFirstSensitiveWidget == 0 && cw->iNbSensitiveWidgets > 0 && cw->iNonSensitiveWidget != 0)  // on est dans la zone des widgets sensitifs.
			// 			{
			// 				//g_print (" => sensitive\n");
			// 				cw->iNbSensitiveWidgets --;
			// 				if (GTK_IS_EXPANDER (w))
			// 					gtk_expander_set_expanded (GTK_EXPANDER (w), TRUE);
			// 			}
			// 			else
			// 			{
			// 				//g_print (" => unsensitive\n");
			// 				if (!GTK_IS_EXPANDER (w))
			// 					gtk_widget_set_sensitive (w, FALSE);
			// 			}
			// 			if (cw->iFirstSensitiveWidget == 0 && cw->iNbControlledWidgets == 0)
			// 			{
			// 				pControlWidgets = g_list_delete_link (pControlWidgets, pControlWidgets);
			// 				g_free (cw);
			// 			}
			// 		}
			// 	}

			// Key description on the left.
			if key.Text != "" { // and maybe need to test different from  "loading..." ?
				build.pLabel, _ = gtk.LabelNew("")
				text := strings.TrimRight(build.translate(key.Text), ":") // dirty hack against ugly trailing colon.

				build.pLabel.SetMarkup(text)
				build.pLabel.SetHAlign(gtk.ALIGN_START)
				// margin-left
				// 		GtkWidget *pAlign = gtk_alignment_new (0., 0.5, 0., 0.);
				build.pKeyBox.PackStart(build.pLabel, false, false, 0)
			}

			// Key widgets on the right. In pWidgetBox, they will be stacked from left to right.
			if !key.IsType(WidgetTextLabel) {
				build.pWidgetBox, _ = gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, MarginGUI)
				build.pKeyBox.PackEnd(build.pWidgetBox, bFullSize, bFullSize, 0)
			}
		}

		// pSubWidgetList = NULL;

		// Build widget for the key, use default if not overridden.
		if key.makeWidget == nil {
			key.makeWidget = key.Type.Builder()
		}
		key.makeWidget(build, key)
	}

	return build.pageScroll
}

//
//--------------------------------------------------------------------[ SAVE ]--

// Save updates the configuration file with user changes.
//
func (build *Builder) Save() {
	for gid := range build.buildGroups {
		for _, key := range build.buildKeys[gid] {
			build.UpdateKeyfileWithKey(key)
		}
	}
	_, str, e := build.Conf.ToData()
	if !build.log.Err(e, "save: get keyfile data") {
		confsettings.SaveFile(build.Conf.File, str)
	}
}

// UpdateKeyfileWithKey updates the builder keyfile with key values.
//
func (build *Builder) UpdateKeyfileWithKey(key *Key) {
	if key.GetValues == nil {
		// log.DEV("KEY EMPTY")
	}

	switch key.Type {
	case WidgetCheckButton, // boolean
		WidgetCheckControlButton: // boolean qui controle le widget suivant
		if key.NbElements > 1 {
			bools := make([]bool, key.NbElements)
			for i, f := range key.GetValues {
				bools[i] = f().(bool)
			}
			build.Conf.KeyFile.SetBooleanList(key.Group, key.Name, bools)
		} else {
			build.Conf.KeyFile.SetBoolean(key.Group, key.Name, key.GetValues[0]().(bool))
		}

	case WidgetIntegerSpin, // integer
		WidgetIntegerScale, // integer in a HScale
		WidgetIntegerSize:  // double integer WxH
		if key.NbElements > 1 {
			ints := make([]int, key.NbElements)
			for i, f := range key.GetValues {
				ints[i] = f().(int)
			}
			build.Conf.KeyFile.SetIntegerList(key.Group, key.Name, ints)
		} else {
			build.Conf.KeyFile.SetInteger(key.Group, key.Name, key.GetValues[0]().(int))
		}

	case WidgetFloatSpin, // float.
		WidgetFloatScale: // float in a HScale.
		if key.NbElements > 1 {
			floats := make([]float64, key.NbElements)
			for i, f := range key.GetValues {
				floats[i] = f().(float64)
			}
			build.Conf.KeyFile.SetDoubleList(key.Group, key.Name, floats)
		} else {
			build.Conf.KeyFile.SetDouble(key.Group, key.Name, key.GetValues[0]().(float64))
		}

	case WidgetColorSelectorRGB, // float x3 avec un bouton de choix de couleur.
		WidgetColorSelectorRGBA: // float x4 avec un bouton de choix de couleur.
		value := key.GetValues[0]().(*gdk.RGBA)
		floats := value.Floats()

		if key.IsType(WidgetColorSelectorRGB) && len(floats) > 3 { // need only 3 values when no alpha.
			floats = floats[:3]
		}
		build.Conf.KeyFile.SetDoubleList(key.Group, key.Name, floats)

	case WidgetNumberedList, // a list of numbered strings.
		WidgetNumberedControlListSimple,    // a list of numbered strings whose current choice defines the sensitivity of the widgets below.
		WidgetNumberedControlListSelective: // a list of numbered strings whose current choice defines the sensitivity of the widgets below given in the list.
		value := key.GetValues[0]().(int)
		// log.DEV("NUMBERED LIST", key.Name, value)
		build.Conf.KeyFile.SetInteger(key.Group, key.Name, value)

	case WidgetTreeViewSortSimple, // N strings listed from top to bottom.
		WidgetTreeViewSortAndModify, // same with possibility to add/remove some.
		WidgetTreeViewMultiChoice:   // N strings that can be selected or not.
		value := key.GetValues[0]().([]string)
		// log.DEV("TREEVIEW", key.Name, value)
		if len(value) > 1 {
			build.Conf.KeyFile.SetStringList(key.Group, key.Name, value)
		} else if len(value) == 1 {
			build.Conf.KeyFile.SetString(key.Group, key.Name, value[0])
		}

	case WidgetStringEntry,
		WidgetPasswordEntry,
		WidgetFileSelector, WidgetFolderSelector, // selectors.
		WidgetSoundSelector, WidgetShortkeySelector,
		WidgetClassSelector, WidgetFontSelector,
		WidgetThemeList,                                      // themes list in a combo, with preview and readme.
		WidgetViewList, WidgetAnimationList, WidgetListDocks, // other filled lists.
		WidgetDialogDecoratorList, WidgetIconThemeList, // ...
		WidgetScreensList,

		WidgetListSimple, WidgetListWithEntry, // a list of strings.
		WidgetDeskletDecorationListSimple,      // desklet decorations list.
		WidgetDeskletDecorationListWithDefault, // idem mais avec le choix "defaut" en plus.
		WidgetIconsList,                        // main dock icons list.

		WidgetImageSelector:

		if key.IsDefault { // The default placeholder is active, no need to save.
			return
		}

		value := key.GetValues[0]().(string)
		if key.IsType(WidgetPasswordEntry) {
			// TODO: cairo_dock_encrypt_string(value, &value)
		}

		build.Conf.KeyFile.SetString(key.Group, key.Name, value)

		//

		//

		// shouldn't be saved, need to check.

	// case WidgetJumpToModuleSimple, // bouton raccourci vers un autre module
	// 	WidgetJumpToModuleIfExists: // idem mais seulement affiche si le module existe.

	// case WidgetLaunchCommandSimple,
	// 	WidgetLaunchCommandIfCondition:

	case WidgetLink: // string with internet.
	case WidgetEmptyWidget: // Containers for custom widget.
	case WidgetEmptyFull:
	case WidgetTextLabel: // Just the text label.
	case WidgetHandbook:
	case WidgetFrame, WidgetExpander:
	case WidgetSeparator:

	default:
		for i, f := range key.GetValues {
			build.log.Info("KEY NOT MATCHED", key.Name, i+1, "/", len(key.GetValues), "[", f(), "]")
		}
	}
}

//
//-----------------------------------------------------------------[ HELPERS ]--

func (build *Builder) translate(str string) string {
	return tran.Sloc(build.gettextDomain, str)
}

func (build *Builder) addWidget(child gtk.IWidget, expand, fill bool, padding uint) {
	if build.pFrameVBox != nil {
		build.pFrameVBox.PackStart(child, expand, fill, padding)
	} else {
		build.pageBox.PackStart(child, expand, fill, padding)
	}
}

// _pack_in_widget_box
func (build *Builder) addSubWidget(child gtk.IWidget) {
	build.pWidgetBox.PackStart(child, false, false, 0)
}

// AddKeyWidget adds a key widget to the page with its getValue call
//
// (was _pack_subwidget).
//
func (build *Builder) AddKeyWidget(child gtk.IWidget, key *Key, f func() interface{}) {
	key.GetValues = append(key.GetValues, f)

	// do {pSubWidgetList = g_slist_append (pSubWidgetList, pSubWidget);} while (0)
	build.pWidgetBox.PackStart(child, key.IsAlignedVertical, key.IsAlignedVertical, 0)
}

// _pack_hscale
func (build *Builder) addKeyScale(child *gtk.Scale, key *Key, f func() interface{}) {
	child.Set("width-request", 150)
	if len(key.AuthorizedValues) >= 4 {

		child.Set("value-pos", gtk.POS_TOP)
		// log.DEV("MISSING SubScale options", string(key.Type), key.AuthorizedValues)
		box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
		// 	GtkWidget * pAlign = gtk_alignment_new(1., 1., 0., 0.)
		labelLeft, _ := gtk.LabelNew(build.translate(key.AuthorizedValues[2]))
		// 	pAlign = gtk_alignment_new(1., 1., 0., 0.)
		labelRight, _ := gtk.LabelNew(build.translate(key.AuthorizedValues[3]))

		box.PackStart(labelLeft, false, false, 0)
		box.PackStart(child, false, false, 0)
		box.PackStart(labelRight, false, false, 0)

		build.AddKeyWidget(box, key, f)
	} else {
		child.Set("value-pos", gtk.POS_LEFT)
		build.AddKeyWidget(child, key, f)
	}
}

// addReset adds a reset value button.
// Requires a callback to restore defaults with the provided values (as strings).
//
func (build *Builder) addReset(key *Key, call func([]string)) {
	if build.originalConf == "" {
		return
	}

	back, _ := gtk.ButtonNewFromIconName("edit-clear", gtk.ICON_SIZE_MENU)
	build.addSubWidget(back)
	back.Connect("clicked", func() {
		conf, e := config.NewFromFile(build.originalConf)
		if build.log.Err(e, "load original conf") {
			return
		}
		str, e := conf.String(key.Group, key.Name)
		if build.log.Err(e, "get original conf key "+key.Group+" :: "+key.Name) {
			return
		}
		call(strings.Split(str, ";"))
	})
}
