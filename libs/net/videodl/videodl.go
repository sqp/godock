// Package videodl provides a video file downloader around the youtube-dl command (only for now).
package videodl

import (
	"github.com/sqp/godock/libs/cdglobal"
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/ternary"

	"os/exec"
	"strconv"
	"strings"
	"time"
)

// others possible backends :
// https://github.com/dwarvesf/glod-cli
// https://github.com/frou/yt2pod

// HistoryFile defines the name of the default history file.
//
const (
	HistoryFile     = "videodl_history.txt"
	WebPath         = "video"
	iconMenuMain    = "video-x-generic"
	iconMenuQuality = "video-x-generic"
	iconMenuTypeDL  = "video-x-generic"
)

// Videodl actions.
const (
	ActionOpenFolder = iota
	ActionCancelDownload
	ActionEnableDownload
	ActionEnableWeb
	ActionEditList
	ActionCount // Number of defined actions
)

const (
	groupQuality = 442
	groupTypeDL  = 443
)

// DownloadFunc defines a quality set function.
//
type DownloadFunc func(*Video, *Format)

// FuncPopupDialog defines a dialog display function.
//
type FuncPopupDialog func(cdtype.DialogData) error

// FuncFilterFormats defines a filter format test callback.
//
type FuncFilterFormats func([]*Format) []*Format

// FuncFilterFormatTest defines a filter format test callback.
//
type FuncFilterFormatTest func(*Format) bool

// FuncNewFiler defines the filer creation from a download backend.
//
type FuncNewFiler func(log cdtype.Logger, url string) (Filer, error)

//
//---------------------------------------------------------------[ VIDEODLER ]--

// VideoDler defines the backend interface.
//
type VideoDler interface {
	New(log cdtype.Logger, url string) (Filer, error)
	MenuQuality() []Quality
}

//
//-------------------------------------------------------------------[ FILER ]--

// Filer defines the usage of a video file downloader.
//
type Filer interface {
	Title() (string, error)
	DownloadCmd(path string, format *Format, progress *Progress) func() error
	Formats() ([]*Format, error)
}

//
//-----------------------------------------------------------------[ QUALITY ]--

// Quality defines the download quality setting.
//
type Quality int

// Quality settings.
const (
	QualityAsk Quality = iota
	QualityBestFound
	QualityBestPossible
	QualityCount // Number of quality actions defined.
)

func (q Quality) String() string {
	switch q {
	case QualityAsk:
		return "Ask quality"
	case QualityBestFound:
		return "Best found"
	case QualityBestPossible:
		return "Best possible"
	}
	return ""
}

func (q Quality) Tooltip() string {
	switch q {
	case QualityAsk:
		return "Use a popup to ask quality for every file"
	case QualityBestPossible:
		return "If available, this will merge the best audio and video streams into a single file."
	}
	return ""
}

//
//------------------------------------------------------------------[ TYPEDL ]--

type WebState int

// TypeDL settings.
const (
	WebStateDisabled WebState = iota
	WebStateStopped
	WebStateStarted
)

func (s WebState) Tooltip() string {
	return "The web service allows links forwarding directly from your browser\nand the web page to edit the download history."
}

type TypeDL int

// TypeDL settings.
const (
	TypeDLAll TypeDL = iota
	TypeDLAudio
	TypeDLVideo
	TypeDLVideoWithAudio
)

func (t TypeDL) String() string {
	switch t {
	case TypeDLAll:
		return "All files"
	case TypeDLAudio:
		return "Audio"
	case TypeDLVideo:
		return "Video"
	case TypeDLVideoWithAudio:
		return "Video with audio"
	}
	return ""
}

func (t TypeDL) Tooltip() string {
	switch t {
	case TypeDLAll:
		return "No filter on file type."
	case TypeDLAudio:
		return "Display only audio files (without video)"
	case TypeDLVideo:
		return "Display only video files (with video, and maybe audio)"
	case TypeDLVideoWithAudio:
		return "Display only files with audio and video"
	}
	return ""
}

//
//----------------------------------------------------------------[ PROGRESS ]--

type Progress struct {
	cur int64
	max int64
}

func NewProgress() *Progress { return &Progress{} }

func (p *Progress) SetMax(m int64) { p.max = m }

func (p *Progress) Write(data []byte) (n int, e error) {
	p.cur += int64(len(data))
	return 0, nil
}

//
//------------------------------------------------------------------[ FORMAT ]--

type Video struct {
	*Format

	Name      string
	URL       string
	DateAdded time.Time
	DateDone  *time.Time
	Fail      bool
	Viewed    bool
	Category  string

	filer Filer
}

func NewVideo(url string, filer Filer) (*Video, error) {
	title, e := filer.Title()
	if e != nil {
		return nil, e
	}
	return &Video{
		Name:      title,
		URL:       url,
		DateAdded: time.Now(),
		filer:     filer,
	}, nil
}

// Format defines the format of a media stream (audio or video).
//
type Format struct {
	Itag       int    // Key reference for the stream.
	Extension  string // Media extension (mp4, flv...)
	Resolution string // Video resolution.
	Size       int    // File size in MiB.
	// 	Note string // Unparsed informations provided by the backend.

	VideoEncoding string
	AudioEncoding string
	AudioBitrate  int

	needDeleteFile string // Set to extension to delete if active.
}

//
//-----------------------------------------------------------------[ BACKEND ]--

// BackendID represents the identifier of a backend.
//
type BackendID int

// Backends list.
const (
	BackendInternal BackendID = iota
	BackendYoutubeDL
)

//
//---------------------------------------------------------------[ CONTROLER ]--

// Downloader defines the usage of the video download manager.
//
type Downloader interface {
	// SetBackend sets the downloader backend ID.
	//
	SetBackend(ID BackendID)

	// SetPath sets the download location.
	//
	SetPath(path string)

	SetTypeDL(typ TypeDL)

	// SetQuality sets the default format quality.
	//
	SetQuality(quality Quality)

	// SetBlacklist sets the formats blacklist.
	//
	SetBlacklist(bl []string)

	SetCommands(openDir, openVideo, openWeb string)

	SetEditList(call func() error)

	// SetPreCheck sets the pre-upload action.
	//
	SetPreCheck(call func() error)

	// SetPostCheck sets the post-upload action.
	//
	SetPostCheck(call func() error)

	// IsActive returns whether a download is in progress or not.
	//
	IsActive() bool

	// FilterBlacklist provides a filter formats call to remove blacklisted file types.
	//
	FilterBlacklist() FuncFilterFormats

	SetEnabledWeb(WebState)
	SetJSWindowOption(string)

	WebRegister()
	WebUnregister()

	WebAutoStart() func()

	WebURL() string

	//----------------------------------------------------------------[ DOWNLOAD ]--

	SetEnabledDL(bool)

	// Download downloads a video file from the server at configured quality (can be ask).
	//
	Download(url string)

	// Enqueue enqueues an item to the download list.
	//
	Enqueue(*Video)
	EnqueueAndStart(*Video)
	Start()

	//-----------------------------------------------------------------[ ACTIONS ]--
	OpenFolder()
	CancelDownload()

	//------------------------------------------------------------[ DOCK HELPERS ]--
	Actions(firstID int, actionAdd func(...*cdtype.Action))
	Menu(cdtype.Menuer)
	DialogQuality(filterFormats FuncFilterFormats, callDialog FuncPopupDialog, callDL DownloadFunc, v *Video)
}

// Controler defines actions needed from a cdtype.AppBase.
//
type Controler interface {
	Action() cdtype.AppAction
	PopupDialog(cdtype.DialogData) error
}

// Manager defines a video file download manager.
//
type Manager struct {
	Path           string
	Quality        Quality
	TypeDL         TypeDL
	EnabledDL      bool
	EnabledWeb     bool
	StartedWeb     bool
	JSWindowOption string

	cmdOpenDir   string
	cmdOpenWeb   string
	cmdOpenVideo string

	blacklist []string

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
		control:      control,
		log:          log,
		backend:      YTDL{},
		history:      hist,
		cmdOpenDir:   cdglobal.CmdOpen,
		cmdOpenWeb:   cdglobal.CmdOpen,
		cmdOpenVideo: cdglobal.CmdOpen,
	}
	m.editList = func() error {
		m.SetStartedWeb(true)
		return m.log.ExecAsync(m.cmdOpenWeb, m.WebURL())
	}
	return m
}

//
//----------------------------------------------------------------[ SETTINGS ]--

// SetPath sets the download location.
//
func (m *Manager) SetPath(path string) {
	m.Path = path
}

// SetQuality sets the default format quality.
//
func (m *Manager) SetQuality(quality Quality) {
	m.Quality = quality
}

// SetBlacklist sets the formats blacklist.
//
func (m *Manager) SetBlacklist(bl []string) {
	m.blacklist = bl
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

// IsActive returns whether a download is in progress or not.
//
func (m *Manager) IsActive() bool { return m.active }

// FilterBlacklist provides a filter formats call to remove blacklisted file types.
//
func (m *Manager) FilterBlacklist() FuncFilterFormats {
	testBlacklist := func(form *Format) bool {
		for _, ext := range m.blacklist {
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
//----------------------------------------------------------------[ DOWNLOAD ]--

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
