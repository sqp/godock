package videodl

import (
	"github.com/sqp/godock/libs/cdglobal"
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/text/strhelp"

	"path/filepath"
	"strings"
)

func (m *Manager) testFiler(vid *Video) bool {
	if vid.filer != nil {
		return true
	}
	var e error
	vid.filer, e = m.backend.New(m.log, vid.URL)
	m.log.Err(e, "videodl: get remote data")
	return e == nil
}

//
//----------------------------------------------------------------[ SETTINGS ]--

// SetBackend sets the download location.
//
func (m *Manager) SetBackend(ID BackendID) {
	switch ID {
	case BackendInternal:
		m.backend = YTDL{}

	case BackendYoutubeDL:
		m.backend = YoutubeDL{}
	}
}

func (m *Manager) SetCommands(openDir, openVideo, openWeb string) {
	m.cmdOpenDir = strhelp.First(openDir, cdglobal.CmdOpen)
	m.cmdOpenVideo = strhelp.First(openVideo, cdglobal.CmdOpen)
	m.cmdOpenWeb = strhelp.First(openWeb, cdglobal.CmdOpen)
}

func (m *Manager) SetEnabledDL(b bool) {
	m.EnabledDL = b
}

func (m *Manager) SetEnabledWeb(s WebState) {
	m.EnabledWeb = s != WebStateDisabled
	m.SetStartedWeb(s == WebStateStarted)
}

// SetTypeDL sets the default download type.
//
func (m *Manager) SetTypeDL(typ TypeDL) {
	m.TypeDL = typ
}

// SetEditList updates the edit list call action.
//
func (m *Manager) SetEditList(call func() error) {
	m.editList = call
}

func (m *Manager) FilePath(vid Video) string {
	title := strings.Replace(vid.Name, "/", "-", -1)
	return filepath.Join(m.Path, title+"."+vid.Extension)
}

//
//----------------------------------------------------------[ ACTIONS DEFINE ]--

// Actions fills an applet actions list with videodl actions.
//
func (m *Manager) Actions(firstID int, actionAdd func(acts ...*cdtype.Action)) {
	m.firstID = firstID
	actionAdd(
		&cdtype.Action{
			ID:   firstID + ActionOpenFolder,
			Name: "Open video folder",
			Icon: "folder",
			Menu: cdtype.MenuEntry,
			Call: m.OpenFolder,
		},
		&cdtype.Action{
			ID:   firstID + ActionCancelDownload,
			Name: "Cancel download",
			Icon: "edit-undo",
			Menu: cdtype.MenuEntry,
			Call: m.CancelDownload,
		},
		&cdtype.Action{
			ID:   firstID + ActionEnableDownload,
			Name: "Enable download",
			Menu: cdtype.MenuCheckBox,
			Call: m.ToggleEnableDownload,
			Bool: &m.EnabledDL,
		},
		&cdtype.Action{
			ID:      firstID + ActionEnableWeb,
			Name:    "Enable web service",
			Menu:    cdtype.MenuCheckBox,
			Bool:    &m.StartedWeb,
			Call:    m.ToggleEnableWeb,
			Tooltip: WebStateDisabled.Tooltip(),
		},
		&cdtype.Action{
			ID:       firstID + ActionEditList,
			Name:     "Edit list",
			Icon:     "media-playlist-repeat",
			Menu:     cdtype.MenuEntry,
			Call:     func() { go m.log.Err(m.editList()) },
			Tooltip:  "Note that this will enable the web service\nYou may have to stop it manually when not needed anymore if you prefer.",
			Threaded: true,
		},
	)
}

//
//------------------------------------------------------------[ ACTIONS CALL ]--

// OpenFolder opens the destination folder.
//
func (m *Manager) OpenFolder() {
	e := m.log.ExecAsync(m.cmdOpenDir, m.Path)
	m.log.Err(e, "open folder")
}

// CancelDownload stops the current download.
//
func (m *Manager) CancelDownload() {
	if m.cmd != nil {
		m.cmd.Process.Kill()
		m.cmd = nil
	}
}

func (m *Manager) ToggleEnableWeb() {
	m.SetStartedWeb(!m.StartedWeb)
}

func (m *Manager) ToggleEnableDownload() {
	m.EnabledDL = !m.EnabledDL
	if m.EnabledDL {
		m.Start()
	}
}

//
//--------------------------------------------------------------------[ MENU ]--

// MenuQuality returns the list of available streams and formats for the video.
//
func (m *Manager) MenuQuality(menu cdtype.Menuer, qual *Quality, list []Quality) {
	if len(list) == 0 {
		return
	}

	sub := menu.AddSubMenu("Quality: "+qual.String(), iconMenuQuality)

	for _, q := range list {
		stq := q // force static for the callback.
		sub.AddRadioEntry(
			q.String(),
			*qual == q,
			groupQuality,
			func() { *qual = stq },
		).SetTooltipText(q.Tooltip())
	}
}

// MenuTypeDL fills the menu with a submenu to select TypeDL (audio, video or both).
//
func (m *Manager) MenuTypeDL(menu cdtype.Menuer, typ *TypeDL) {
	sub := menu.AddSubMenu("File Type: "+typ.String(), iconMenuTypeDL)

	for _, t := range []TypeDL{TypeDLAll, TypeDLAudio, TypeDLVideo, TypeDLVideoWithAudio} {
		stt := t // force static for the callback.
		sub.AddRadioEntry(
			t.String(),
			*typ == t,
			groupTypeDL,
			func() { *typ = stt },
		).SetTooltipText(t.Tooltip())
	}
}

//
//------------------------------------------------------------------[ DIALOG ]--

func (m *Manager) DialogCategory(filterFormats FuncFilterFormats, callDialog FuncPopupDialog, callDL DownloadFunc, vid *Video) {
	ids := []string{"", "Go Game", "Info"}
	sel := 0
	for k, v := range ids {
		if m.category == v {
			sel = k
			break
		}
	}

	e := callDialog(cdtype.DialogData{
		Message: vid.Name + "\n\nSelect category:",
		Widget: cdtype.DialogWidgetList{
			Values:       strings.Join(ids, ";"),
			InitialValue: sel,
		},
		Buttons: "ok;cancel",
		Callback: cdtype.DialogCallbackValidInt(func(id int) {
			if id < len(ids) { // Shouldn't be out of range, but just to be safe...
				m.category = ids[id]
				vid.Category = m.category
				m.DialogQuality(filterFormats, callDialog, callDL, vid)
			}
		}),
	})
	m.log.Err(e, "popup")
}
