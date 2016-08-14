package videodl

import (
	"github.com/sqp/godock/libs/cdglobal"
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/ternary"
	"github.com/sqp/godock/libs/text/strhelp"

	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Controler defines actions needed from a cdtype.AppBase.
//
type Controler interface {
	Action() cdtype.AppAction
	PopupDialog(cdtype.DialogData) error
}

// Manager defines a video file download manager.
//
type Manager struct {
	*Config

	EnabledWeb bool
	StartedWeb bool

	newFiler FuncNewFiler

	backend VideoDler

	firstID int // Position of first action, when used with other applets services.

	active     bool // true when downloading.
	actionPre  func() error
	actionPost func() error
	cmd        *exec.Cmd

	history *HistoryVideo

	progress *Progress

	category string // currently selected category (group / dir).

	editList func() error

	control Controler
	log     cdtype.Logger
}

// NewManager creates a video file download manager.
//
func NewManager(control Controler, log cdtype.Logger, hist *HistoryVideo) *Manager {
	m := &Manager{
		control: control,
		log:     log,
		backend: YTDL{},
		history: hist,
		Config: &Config{
			CmdOpenDir:   cdglobal.CmdOpen,
			CmdOpenWeb:   cdglobal.CmdOpen,
			CmdOpenVideo: cdglobal.CmdOpen},
	}
	m.editList = func() error {
		m.SetStartedWeb(true)
		return m.log.ExecAsync(m.CmdOpenWeb, m.WebURL())
	}
	return m
}

//
//----------------------------------------------------------------[ SETTINGS ]--

// SetConfig sets the config data for the manager.
//
func (m *Manager) SetConfig(conf *Config) {
	m.Config = conf

	switch m.Config.BackendID {
	case BackendInternal:
		m.backend = YTDL{}

	case BackendYoutubeDL:
		m.backend = YoutubeDL{}
	}

	m.CmdOpenDir = strhelp.First(m.CmdOpenDir, cdglobal.CmdOpen)
	m.CmdOpenVideo = strhelp.First(m.CmdOpenVideo, cdglobal.CmdOpen)
	m.CmdOpenWeb = strhelp.First(m.CmdOpenWeb, cdglobal.CmdOpen)
}

// SetEnabledWeb sets the web service state.
//
func (m *Manager) SetEnabledWeb(s WebState) {
	m.EnabledWeb = s != WebStateDisabled
	m.SetStartedWeb(s == WebStateStarted)
}

// SetPreCheck sets the pre-upload action.
//
func (m *Manager) SetPreCheck(call func() error) {
	m.actionPre = call
}

// SetPostCheck sets the post-upload action.
//
func (m *Manager) SetPostCheck(call func() error) {
	m.actionPost = call
}

// SetTypeDL sets the default download type.
//
// func (m *Manager) SetTypeDL(typ TypeDL) {
// 	m.TypeDL = typ
// }

// SetEditList updates the edit list call action.
//
func (m *Manager) SetEditList(call func() error) {
	m.editList = call
}

// FilePath returns the full path to the video file.
//
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
	e := m.log.ExecAsync(m.CmdOpenDir, m.Path)
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

// ToggleEnableWeb toggles the status of the web service.
//
func (m *Manager) ToggleEnableWeb() {
	m.SetStartedWeb(!m.StartedWeb)
}

// ToggleEnableDownload toggles the status of the download activity (dl/pause).
//
func (m *Manager) ToggleEnableDownload() {
	m.EnabledDL = !m.EnabledDL
	if m.EnabledDL {
		m.Start()
	}
}

//
//--------------------------------------------------------------------[ MENU ]--

// Menu fills an applet actions list with videodl actions.
//
func (m *Manager) Menu(menu cdtype.Menuer) {
	m.control.Action().BuildMenu(menu, []int{m.firstID + ActionOpenFolder, m.firstID + ActionEnableDownload})

	if m.active {
		m.control.Action().BuildMenu(menu, []int{m.firstID + ActionCancelDownload})
	}

	subTitle := "Video Download"
	if !m.EnabledDL {
		subTitle += " (paused)"
	}
	sub := menu.AddSubMenu(subTitle, iconMenuMain)
	if m.history.Queued() > 0 {
		sub.AddEntry("Queued: "+strconv.Itoa(m.history.Queued()), "emblem-downloads", nil)
		sub.AddSeparator()
	}

	// sub.AddEntry("Edit list", "media-playlist-repeat", func() { m.log.ExecAsync(m.cmdOpenWeb, m.WebURL()) }).
	// 	SetTooltipText("Note that this will enable the web service\nYou may have to stop it manually when not needed anymore if you prefer.")
	m.control.Action().BuildMenu(sub, []int{m.firstID + ActionEditList, m.firstID + ActionEnableWeb})

	m.MenuTypeDL(sub, &m.TypeDL)
	m.MenuQuality(sub, &m.Quality, m.backend.MenuQuality())
}

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

// DialogQuality prepares a dialog to ask the desired quality and download.
//
func (m *Manager) DialogQuality(filterFormats FuncFilterFormats, callDialog FuncPopupDialog, callDL DownloadFunc, vid *Video) {
	if vid.Name == "" {
		var e error
		vid.Name, e = vid.filer.Title()
		if m.log.Err(e, "videodl: missing title") {
			return
		}
	}

	formats, e := vid.filer.Formats()
	if m.log.Err(e, "format") || len(formats) == 0 {
		return
	}

	// formats = append(formats,
	// Format{Res: QualityBestFound.String(), Code: "best"},
	// Format{Res: QualityBestPossible.String()},
	// )

	// Reduce list of formats.
	formats = filterFormats(formats)

	var ids []string
	for _, form := range formats {
		size := "?"
		if form.Size > 0 {
			size = strconv.Itoa(form.Size)
		}

		ids = append(ids, size+" MB\t"+
			form.Extension+"\t"+form.Resolution+"\t"+
			form.AudioEncoding+"\t"+form.VideoEncoding,
		)
	}

	ids = append(ids, "Category: "+m.category)
	if vid.Format != nil {
		if vid.Format.needDeleteFile == "" {
			vid.Format.needDeleteFile = vid.Extension
		} else {
			vid.Format.needDeleteFile = ""
		}
		ids = append(ids, "Delete current file: "+ternary.String(vid.Format.needDeleteFile != "", "Yes", "No"))

	}

	e = callDialog(cdtype.DialogData{
		Message: vid.Name + "\n\nSelect quality:",
		// UseMarkup: true,
		Widget: cdtype.DialogWidgetList{
			Values:       strings.Join(ids, ";"),
			InitialValue: 0,
		},
		Buttons: "ok;cancel",
		Callback: cdtype.DialogCallbackValidInt(func(id int) {
			switch {
			case id > len(formats)+1:
				m.log.NewErr("ID out of range", "videodl DialogQuality")

			case id == len(formats)+1:
				if vid.Format == nil {
					m.log.NewErr("select delete file but format missing", "videodl DialogQuality")
				} else {
					m.DialogQuality(filterFormats, callDialog, callDL, vid)
				}

			case id == len(formats):
				m.DialogCategory(filterFormats, callDialog, callDL, vid)

			default:
				callDL(vid, formats[id])
			}
		}),
	})
	m.log.Err(e, "popup")
}

// DialogCategory prepares a dialog to ask the category assigned to next downloads.
//
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

//
//----------------------------------------------------------------[ DOWNLOAD ]--

// IsActive returns whether a download is in progress or not.
//
func (m *Manager) IsActive() bool { return m.active }

// Download downloads a video file from the server at configured quality (can be ask).
//
func (m *Manager) Download(url string) {
	filer, e := m.backend.New(m.log, url)
	if m.log.Err(e, "videodl init", url) {
		return
	}
	vid, e := NewVideo(url, filer)
	vid.Category = m.category

	if m.log.Err(e, "videodl: add to queue (get title)") {
		return
	}
	m.getQuality(vid)
}

func (m *Manager) getQuality(vid *Video) {
	download := func(vid *Video, format *Format) {
		// invert need delete (default = true).
		if vid.Format != nil {
			switch {
			case format.Extension == vid.Extension: // no post deletion, the current file will be overwritten.

			case vid.Format.needDeleteFile != "": // set to false
				format.needDeleteFile = ""

			default: // set to true
				format.needDeleteFile = vid.Extension
			}

			e := m.history.Remove(vid, &m.history.List)
			m.log.Err(e, "videodl: getQuality remove from list")
		}

		vid.Format = format
		m.EnqueueAndStart(vid)
	}

	if !m.testFiler(vid) {
		return
	}

	switch m.Quality {
	case QualityAsk:
		go func() {
			m.DialogQuality(m.FilterBlacklist(), m.control.PopupDialog, download, vid)
		}()

		// case QualityBestFound:
		// 	download("best")

		// case QualityBestPossible:
		// 	download("")
	}
}

// Enqueue enqueues an item to the download list.
//
func (m *Manager) Enqueue(vid *Video) {
	e := m.history.Add(vid)
	m.log.Err(e, "videodl: save data")
}

// EnqueueAndStart enqueues an item and starts downloading.
//
func (m *Manager) EnqueueAndStart(vid *Video) {
	m.Enqueue(vid)
	m.Start()
}

// Start starts downloading queued items.
//
func (m *Manager) Start() {
	if m.IsActive() || m.history.Queued() == 0 { // Only one worker.
		return
	}
	go func() {
		m.active = true
		if m.actionPre != nil {
			m.log.Err(m.actionPre(), "actionPre")
		}
		defer func() {
			if m.actionPost != nil {
				m.log.Err(m.actionPost(), "actionPost")
			}
			m.active = false
		}()

		var e error
		for {
			vid, ok := m.history.Next()
			if !ok || !m.EnabledDL {
				return
			}

			if !m.testFiler(vid) {
				return
			}

			if vid.filer == nil {
				vid.filer, e = m.backend.New(m.log, vid.URL)
				if m.log.Err(e, "videodl: get data") {
					return
				}
			}

			m.progress = NewProgress()
			e = vid.filer.DownloadCmd(m.Path, vid.Format, m.progress)()
			if m.log.Err(e, "Download") {
				vid.Fail = true
			} else if vid.Format.needDeleteFile != "" {
				m.log.Info("to delete", vid.Category, vid.Name, vid.needDeleteFile)
			}

			// Remove from queue.
			// Whether it was a success or not, we cannot chain loop over the same item.
			e = m.history.Done()
			m.log.Err(e, "videodl: save data")
		}
	}()
}

//
//-----------------------------------------------------------------[ FILTERS ]--

// FilterBlacklist provides a filter formats call to remove blacklisted file types.
//
func (m *Manager) FilterBlacklist() FuncFilterFormats {
	testBlacklist := func(form *Format) bool {
		for _, ext := range m.Blacklist {
			if form.Extension == ext {
				return false
			}
		}
		return true
	}

	testTypeDL := func(form *Format) bool {
		switch m.TypeDL {
		case TypeDLAudio:
			return form.AudioEncoding != "" && form.VideoEncoding == ""

		case TypeDLVideo:
			return form.VideoEncoding != ""

		case TypeDLVideoWithAudio:
			return form.VideoEncoding != "" && form.AudioEncoding != ""
		}
		return true
	}

	return func(formats []*Format) []*Format {
		return FilterFormats(formats, testBlacklist, testTypeDL)
	}
}

// FilterFormats filters a list of formats with the provided test.
//
func FilterFormats(formats []*Format, filters ...FuncFilterFormatTest) []*Format {
	var out []*Format
	for _, form := range formats {
		ok := true
		for _, f := range filters {
			ok = ok && f(form)
		}
		if ok {
			out = append(out, form)
		}
	}
	return out
}

//
//------------------------------------------------------------------[ COMMON ]--

// testFiler tests if the video filer is valid.
// It tries to rebuild it if missing, and only returns false if failed
// (error is logged).
//
func (m *Manager) testFiler(vid *Video) bool {
	if vid.filer != nil {
		return true
	}
	var e error
	vid.filer, e = m.backend.New(m.log, vid.URL)
	m.log.Err(e, "videodl: get remote data")
	return e == nil
}
