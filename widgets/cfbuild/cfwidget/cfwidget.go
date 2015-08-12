package cfwidget

import (
	"github.com/sqp/godock/widgets/cfbuild/cftype" // Types for config file builder usage.
)

// Maker returns the make widget call for the key type.
//
func Maker(key *cftype.Key) func() {
	makeWidget, ok := Types[key.Type]
	if !ok {
		return nil
	}
	return func() { makeWidget(key) }
}

// Types defines widget build methods by key type.
//
var Types = map[cftype.KeyType]func(key *cftype.Key){
	cftype.KeyTextLabel: Text,
	cftype.KeyLink:      Link,
	cftype.KeySeparator: Separator,
	cftype.KeyFrame:     Frame,
	cftype.KeyExpander:  Frame,

	cftype.KeyBoolButton:        CheckButton,   // bool
	cftype.KeyBoolCtrl:          CheckButton,   // bool controling the next widget.
	cftype.KeyIntSpin:           IntegerSpin,   // int in a spin button.
	cftype.KeyIntScale:          IntegerScale,  // int in a HScale.
	cftype.KeyIntSize:           IntegerSize,   // int pair spin buttons WxH.
	cftype.KeyFloatSpin:         Float,         // float.
	cftype.KeyFloatScale:        Float,         // float in a HScale.
	cftype.KeyColorSelectorRGB:  ColorSelector, // float x3 avec un bouton de choix de couleur.
	cftype.KeyColorSelectorRGBA: ColorSelector, // float x4 avec un bouton de choix de couleur.
	cftype.KeyFontSelector:      FontSelector,  // string avec un selecteur de font a cote du GtkEntry.

	cftype.KeyStringEntry:      Strings,
	cftype.KeyPasswordEntry:    Strings,
	cftype.KeyFileSelector:     Strings,
	cftype.KeyFolderSelector:   Strings,
	cftype.KeySoundSelector:    Strings,
	cftype.KeyShortkeySelector: Strings,
	cftype.KeyClassSelector:    Strings,
	cftype.KeyImageSelector:    Strings,

	cftype.KeyListSimple:       Lists,
	cftype.KeyListEntry:        Lists,
	cftype.KeyListNumbered:     Lists,
	cftype.KeyListNbCtrlSimple: Lists,
	cftype.KeyListNbCtrlSelect: Lists,

	cftype.KeyTreeViewSortSimple:  TreeView,
	cftype.KeyTreeViewSortModify:  TreeView,
	cftype.KeyTreeViewMultiChoice: TreeView,

	cftype.KeyLaunchCmdSimple: LaunchCommand,
	cftype.KeyLaunchCmdIf:     LaunchCommand,

	cftype.KeyEmptyWidget: Nil, // Containers for custom widget.
	cftype.KeyEmptyFull:   Nil,

	// Dock only.

	cftype.KeyListThemeApplet: ListThemeApplet, // List themes in a combo, with preview and readme.
	cftype.KeyListDocks:       ListDock,        // liste des docks existant.
	cftype.KeyListViews:       ListView,        // List of dock views.

	cftype.KeyListThemeDesktopIcon: ListThemeDesktopIcon,

	cftype.KeyListAnimation:          ListAnimation,         // List of animations.
	cftype.KeyListDialogDecorator:    ListDialogDecorator,   // liste des decorateurs de dialogue.
	cftype.KeyListDeskletDecoSimple:  ListDeskletDecoration, // liste des decorations de desklet.
	cftype.KeyListDeskletDecoDefault: ListDeskletDecoration, // idem mais avec le choix "defaut" en plus.
	cftype.KeyListIconsMainDock:      ListIconsMainDock,
	cftype.KeyListScreens:            ListScreens,

	cftype.KeyHandbook: Handbook,

	//  WidgetJumpToModuleSimple:JumpToModule,
	// 	WidgetJumpToModuleIfExists:JumpToModule,
}

// Nil defines a nil widget builder.
//
func Nil(key *cftype.Key) {}
