package cfbuild

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
	"github.com/gotk3/gotk3/gtk"

	"github.com/sqp/godock/libs/cdglobal"     // Dock types.
	"github.com/sqp/godock/libs/cdtype"       // Logger type.
	"github.com/sqp/godock/libs/dock/confown" // New dock own settings.
	"github.com/sqp/godock/libs/text/tran"    // Translate.

	"github.com/sqp/godock/widgets/cfbuild/cfprint"  // Print config file builder keys.
	"github.com/sqp/godock/widgets/cfbuild/cftype"   // Types for config file builder usage.
	"github.com/sqp/godock/widgets/cfbuild/cfwidget" // Widgets for config file builder.
	"github.com/sqp/godock/widgets/gtk/newgtk"       // Create widgets.

	"strings"
)

//
//-----------------------------------------------------------------[ BUILDER ]--

// builder builds a Cairo-Dock configuration page.
//
type builder struct {
	gtk.Box // Main container.

	pageBox *gtk.Box // box for the page currently builded. Was pGroupBox

	pFrame     *gtk.Frame // frames only.
	pFrameVBox *gtk.Box   // Container for widgets in a frame.

	NbControlled int // was iNbControlledWidgets

	// Build keys.
	buildGroups []string                 // Order as detected and used by BuildAll.
	buildKeys   map[string][]*cftype.Key // Using group name as index.

	postSave func() // optional post save action.

	// extra data.
	conf          cftype.Storage
	data          cftype.Source
	log           cdtype.Logger
	gettextDomain string
	originalConf  string // path to default config file.
}

// NewBuilder creates a configuration page builder from a config storage.
//
func NewBuilder(source cftype.Source, log cdtype.Logger, conf cftype.Storage, originalConf, gettextDomain string) cftype.Builder {
	return &builder{
		Box:           *newgtk.Box(gtk.ORIENTATION_VERTICAL, 0),
		conf:          conf,
		data:          source,
		log:           log,
		originalConf:  originalConf,
		gettextDomain: gettextDomain,
	}
}

func (build *builder) Log() cdtype.Logger        { return build.log }
func (build *builder) Storage() cftype.Storage   { return build.conf }
func (build *builder) Source() cftype.Source     { return build.data }
func (build *builder) BoxPage() *gtk.Box         { return build.pageBox }
func (build *builder) SetFrame(frame *gtk.Frame) { build.pFrame = frame }
func (build *builder) SetFrameBox(box *gtk.Box)  { build.pFrameVBox = box }
func (build *builder) SetNbControlled(nb int)    { build.NbControlled = nb }
func (build *builder) SetPostSave(call func())   { build.postSave = call }
func (build *builder) Groups() []string          { return build.buildGroups }

// Free removes back references so that objects can be freed.
//
func (build *builder) Free() {
	build.Storage().SetBuilder(nil)
	for _, keys := range build.buildKeys {
		for _, key := range keys {
			key.Builder = nil
		}
	}
}

// Clear clears the builder for another use.
//
func (build *builder) Clear() {
	build.Storage().SetBuilder(nil)
	for _, keys := range build.buildKeys {
		for _, key := range keys {
			key.Builder = nil
		}
	}
}

// AddGroup adds a group with optional keys.
//
func (build *builder) AddGroup(group string, keys ...*cftype.Key) {
	build.buildGroups = append(build.buildGroups, group)
	if build.buildKeys == nil {
		build.buildKeys = make(map[string][]*cftype.Key)
	}

	build.AddKeys(group, keys...)
}

// AddKeys adds one or many keys to an existing group.
//
func (build *builder) AddKeys(group string, keys ...*cftype.Key) {
	for _, key := range keys {
		if key == nil {
			build.Log().Debug("builder add nil key", "group:", group)
			continue
		}
		build.buildKeys[group] = append(build.buildKeys[group], key)
		key.SetBuilder(build)
	}
}

//
//---------------------------------------------------------[ KEYS INTERACTION]--

// KeyAction acts on a key if found. Key access errors will just be logged.
//
func (build *builder) KeyAction(group, name string, action func(*cftype.Key)) bool {
	for _, key := range build.buildKeys[group] {
		if key.Name == name {
			action(key)
			return true
		}
	}
	build.log.Errorf("builder key action", "get key=%s  ::  %s", group, name)
	return false
}

// KeyWalk runs the given call on all keys in the defined group order.
//
func (build *builder) KeyWalk(call func(*cftype.Key)) {
	for _, group := range build.buildGroups {
		for _, key := range build.buildKeys[group] {
			call(key)
		}
	}
}

func (build *builder) KeyBool(g, k string) (v bool)     { build.keyGet(g, k, &v); return v }
func (build *builder) KeyInt(g, k string) (v int)       { build.keyGet(g, k, &v); return v }
func (build *builder) KeyFloat(g, k string) (v float64) { build.keyGet(g, k, &v); return v }
func (build *builder) KeyString(g, k string) (v string) { build.keyGet(g, k, &v); return v }

func (build *builder) keyGet(group, name string, val interface{}) {
	build.KeyAction(group, name, func(key *cftype.Key) { key.ValueGet(val) })
}

//
//-------------------------------------------------------------------[ BUILD ]--

// BuildPage builds a Cairo-Dock configuration widget for the given group.
//
func (build *builder) BuildPage(group string) cftype.GtkWidgetBase {
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

	build.pageBox = newgtk.Box(gtk.ORIENTATION_VERTICAL, cftype.MarginGUI)
	build.pageBox.SetBorderWidth(cftype.MarginGUI)

	build.pFrameVBox = nil

	build.NbControlled = 0

	// FRAMES ONLY
	build.pFrame = nil

	for _, key := range build.buildKeys[group] {
		build.buildKey(key)
	}
	pageScroll := newgtk.ScrolledWindow(nil, nil)
	pageScroll.Add(build.pageBox)
	return pageScroll
}

// buildKey builds a Cairo-Dock configuration widget with the given keys.
//
func (build *builder) buildKey(key *cftype.Key) {
	if key.DisplayMode != cdglobal.DisplayModeAll &&
		key.DisplayMode != build.Source().DisplayMode() {

		key.Log().Debug("DisplayMode key dropped", key.Group, key.Name)
		return
	}

	makeWidget := key.MakeWidget()
	if makeWidget == nil {
		makeWidget = cfwidget.Maker(key)
		if makeWidget == nil {
			key.Log().Errorf("no make widget call", "type=%s  [key=%s / %s]\n", key.Type, key.Group, key.Name)
			return
		}
	}

	key.Log().Debug(key.Group, "/", key.Name, key.Type, key.AuthorizedValues)

	fullSize := key.IsType(cftype.KeyListThemeApplet, cftype.KeyListViews, cftype.KeyEmptyFull, cftype.KeyHandbook)

	if !key.IsType(cftype.KeyFrame, cftype.KeyExpander, cftype.KeySeparator) {
		// Create Key box.
		if key.IsType(cftype.KeyListThemeApplet, cftype.KeyListViews) {
			keybox := newgtk.Box(gtk.ORIENTATION_HORIZONTAL, cftype.MarginGUI)
			addvbox := newgtk.Box(gtk.ORIENTATION_VERTICAL, 0)
			key.SetKeyBox(keybox)
			key.SetAdditionalItemsVBox(addvbox)

			// Pack the keybox as first item of the big box.
			build.PackWidget(addvbox, fullSize, fullSize, 0)
			addvbox.PackStart(keybox, false, false, 0)

		} else {
			if key.IsAlignedVertical {
				key.Log().Info("IsAlignedVertical", key.Group, "/", key.Name)
				key.SetKeyBox(newgtk.Box(gtk.ORIENTATION_VERTICAL, cftype.MarginGUI))
			} else {
				key.SetKeyBox(newgtk.Box(gtk.ORIENTATION_HORIZONTAL, cftype.MarginGUI))
			}

			build.PackWidget(key.KeyBox(), fullSize, fullSize, 0)
		}

		if key.Tooltip != "" {
			key.KeyBox().SetTooltipText(build.Translate(key.Tooltip))
		}

		// 	if (pControlWidgets != NULL)
		// 	{
		// 		CDControlWidget *cw = pControlWidgets->data;
		// 		//g_print ("ctrl (%d widgets)\n", NbControlled);
		// 		if (cw->pControlContainer == (pFrameVBox ? pFrameVBox : build))
		// 		{
		// 			//g_print ("ctrl (NbControlled:%d, iFirstSensitiveWidget:%d, iNbSensitiveWidgets:%d)\n", NbControlled, iFirstSensitiveWidget, iNbSensitiveWidgets);
		// 			cw->NbControlled --;
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
		// 			if (cw->iFirstSensitiveWidget == 0 && cw->NbControlled == 0)
		// 			{
		// 				pControlWidgets = g_list_delete_link (pControlWidgets, pControlWidgets);
		// 				g_free (cw);
		// 			}
		// 		}
		// 	}

		// Key description on the left.
		if key.Text != "" { // and maybe need to test different from  "loading..." ?
			label := newgtk.Label("")
			key.SetLabel(label)
			text := strings.TrimRight(build.Translate(key.Text), ":") // dirty hack against ugly trailing colon.

			label.SetMarkup(text)
			label.SetHAlign(gtk.ALIGN_START)
			// 		GtkWidget *pAlign = gtk_alignment_new (0., 0.5, 0., 0.);
			key.KeyBox().PackStart(key.Label(), false, false, 2)
		}

		// Key widgets on the right. In WidgetBox, they will be stacked from left to right.
		if !key.IsType(cftype.KeyTextLabel) {
			box := newgtk.Box(gtk.ORIENTATION_HORIZONTAL, cftype.MarginGUI)
			key.SetWidgetBox(box)
			key.KeyBox().PackEnd(box, fullSize, fullSize, 0)
		}
	}

	// Build widget for the key, use the default for the type if not overridden.
	makeWidget()
}

//
//--------------------------------------------------------------------[ SAVE ]--

// Save updates the configuration file with user changes.
//
func (build *builder) Save() {
	if confown.Current.ToVirtual(build.conf.FilePath()) {
		cfprint.Updated(build)
		return
	}

	build.Log().DEV("save real")

	build.KeyWalk((*cftype.Key).UpdateStorage)

	_, str, e := build.conf.ToData()
	if build.log.Err(e, "save: get keyfile data") {
		return
	}

	tofile, e := confown.SaveFile(build.Log(), build.conf.FilePath(), str)
	if !build.Log().Err(e, "save config") && build.postSave != nil && tofile {
		build.postSave()
	}
}

//
//-----------------------------------------------------------------[ HELPERS ]--

// Translate translates the given string using the builder domain.
//
func (build *builder) Translate(str string) string {
	return tran.Sloc(build.gettextDomain, str)
}

//
//-----------------------------------------------------------------[ PACKING ]--

// PackWidget packs a widget in the main box.
//
func (build *builder) PackWidget(child gtk.IWidget, expand, fill bool, padding uint) {
	if build.pFrameVBox != nil {
		build.pFrameVBox.PackStart(child, expand, fill, padding)
	} else {
		build.pageBox.PackStart(child, expand, fill, padding)
	}
}

//
//----------------------------------------------------------[ BUILDER TWEAKS ]--

// TweakAddGroup creates a tweak callback to add a group with keys to a builder.
//
func TweakAddGroup(group string, keys ...*cftype.Key) func(cftype.Builder) {
	return func(build cftype.Builder) {
		build.AddGroup(group, keys...)
	}
}

// TweakAddKeys creates a tweak callback to add keys to an existing builder group.
//
func TweakAddKeys(group string, keys ...*cftype.Key) func(cftype.Builder) {
	return func(build cftype.Builder) {
		build.AddKeys(group, keys...)
	}
}

// TweakKeyAction creates a tweak callback to edit a key of an existing builder.
//
func TweakKeyAction(group, name string, actions ...func(*cftype.Key)) func(cftype.Builder) {
	return func(build cftype.Builder) {
		for _, act := range actions {
			build.KeyAction(group, name, act)
		}
	}
}

// TweakKeyMakeWidget creates a tweak callback to set a key widget builder.
//
func TweakKeyMakeWidget(group, name string, call func(*cftype.Key)) func(cftype.Builder) {
	return TweakKeyAction(group, name, func(key *cftype.Key) {
		key.SetMakeWidget(call)
	})
}

// TweakKeySetAlignedVertical creates a tweak callback to set a key widget alignment.
//
func TweakKeySetAlignedVertical(group, name string) func(cftype.Builder) {
	return TweakKeyAction(group, name, func(key *cftype.Key) {
		key.IsAlignedVertical = true
	})
}

// TweakKeySetLabelSelectable creates a tweak callback to set the key label text selectable.
// Only valid after build.
//
func TweakKeySetLabelSelectable(group, name string) func(cftype.Builder) {
	return TweakKeyAction(group, name, func(key *cftype.Key) {
		key.Label().SetSelectable(true)
	})
}
