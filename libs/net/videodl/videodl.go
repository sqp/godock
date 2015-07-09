// Package videodl provides a video file downloader around the youtube-dl command (only for now).
package videodl

import (
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/text/strhelp"

	"os/exec"
)

// Videodl actions.
const (
	ActionOpenFolder = iota
	ActionCancelDownload
	ActionCount // Number of defined actions
)

// FuncDownload defines a quality set function.
//
type FuncDownload func(quality string)

// FuncPopupDialog defines a dialog display function.
//
type FuncPopupDialog func(cdtype.DialogData) error

// FuncFilterFormats defines a filter format test callback.
//
type FuncFilterFormats func([]Format) []Format

// FuncFilterFormatTest defines a filter format test callback.
//
type FuncFilterFormatTest func(Format) bool

// FuncNewFiler defines the filer creation from a download backend.
//
type FuncNewFiler func(log cdtype.Logger, url string) Filer

//
//-------------------------------------------------------------------[ FILER ]--

// Filer defines the usage of a video file downloader.
//
type Filer interface {
	Title() (string, error)
	DownloadCmd(path, quality string) *exec.Cmd
	Formats() ([]Format, error)
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

//
//------------------------------------------------------------------[ FORMAT ]--

// Format defines the format of a media stream (audio or video).
//
type Format struct {
	Code string // Key reference for the stream.
	Ext  string // Media extension (mp4, flv...)
	Res  string // Video resolution.
	Note string // Unparsed informations provided by the backend.
	Size string // Stream size in MiB.
}

//
//-----------------------------------------------------------------[ BACKEND ]--

// BackendID represents the identifier of a backend.
//
type BackendID int

// Backends list.
const (
	BackendYoutubeDL BackendID = iota
)

//
//---------------------------------------------------------------[ CONTROLER ]--

// Downloader defines the usage of the video download manager.
//
type Downloader interface {
	SetBackend(ID BackendID)

	// SetPath sets the download location.
	//
	SetPath(path string)

	// SetQuality sets the default format quality.
	//
	SetQuality(quality Quality)

	// SetBlacklist sets the formats blacklist.
	//
	SetBlacklist(bl []string)

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

	//----------------------------------------------------------------[ DOWNLOAD ]--

	// Download downloads a video file from the server at configured quality (can be ask).
	//
	Download(url string)

	// Enqueue enqueues an item to the download list.
	//
	Enqueue(call func() error)
	EnqueueStart(call func() error)
	Start()

	//-----------------------------------------------------------------[ ACTIONS ]--
	OpenFolder()
	CancelDownload()

	//------------------------------------------------------------[ DOCK HELPERS ]--
	Actions(firstID int, actionAdd func(...*cdtype.Action))
	MenuQuality(cdtype.Menuer)
	DialogQuality(filterFormats FuncFilterFormats, callDialog FuncPopupDialog, callDL FuncDownload, f Filer)
}

// Controler defines actions needed from a cdtype.AppBase.
//
type Controler interface {
	BuildMenu(menu cdtype.Menuer, actionIds []int)
	PopupDialog(cdtype.DialogData) error
}

// Manager defines a video file download manager.
//
type Manager struct {
	Path      string
	Quality   Quality
	blacklist []string

	newFiler FuncNewFiler

	firstID int // Position of first action, when used with other applets services.

	active     bool // true when downloading.
	actionPre  func() error
	actionPost func() error
	queue      chan func() error
	cmd        *exec.Cmd
	// onResult   func(Links)

	// history     []Links
	// historyMax  int
	// historyFile string

	control Controler
	log     cdtype.Logger
}

// NewManager creates a video file download manager.
//
func NewManager(control Controler, log cdtype.Logger) *Manager {
	return &Manager{
		control: control,
		log:     log,
		queue:   make(chan func() error, 10),
	}
}

//
//----------------------------------------------------------------[ SETTINGS ]--

// SetBackend sets the download location.
//
func (d *Manager) SetBackend(ID BackendID) {
	d.newFiler = NewYoutubeDL
}

// SetPath sets the download location.
//
func (d *Manager) SetPath(path string) {
	d.Path = path
}

// SetQuality sets the default format quality.
//
func (d *Manager) SetQuality(quality Quality) {
	d.Quality = quality
}

// SetBlacklist sets the formats blacklist.
//
func (d *Manager) SetBlacklist(bl []string) {
	d.blacklist = bl
}

// SetPreCheck sets the pre-upload action.
//
func (d *Manager) SetPreCheck(call func() error) {
	d.actionPre = call
}

// SetPostCheck sets the post-upload action.
//
func (d *Manager) SetPostCheck(call func() error) {
	d.actionPost = call
}

// SetOnResult sets the result return method.
//
// func (up *Uploader) SetOnResult(call func(Links)) {
// 	up.onResult = call
// }

// IsActive returns whether a download is in progress or not.
//
func (d *Manager) IsActive() bool { return d.active }

// FilterBlacklist provides a filter formats call to remove blacklisted file types.
//
func (d *Manager) FilterBlacklist() FuncFilterFormats {
	testBlacklist := func(form Format) bool {
		for _, ext := range d.blacklist {
			if form.Ext == ext {
				return false
			}
		}
		return true
	}

	return func(formats []Format) []Format {
		return FilterFormats(testBlacklist, formats)
	}
}

// FilterFormats filters a list of formats with the provided test.
//
func FilterFormats(filter FuncFilterFormatTest, formats []Format) []Format {
	var out []Format
	for _, form := range formats {
		if filter(form) {
			out = append(out, form)
		}
	}
	return out
}

//
//----------------------------------------------------------------[ DOWNLOAD ]--

// Download downloads a video file from the server at configured quality (can be ask).
//
func (d *Manager) Download(url string) {
	filer := d.newFiler(d.log, url)
	download := func(quality string) {
		d.EnqueueStart(d.runCmd(filer.DownloadCmd(d.Path, quality)))
	}

	switch d.Quality {
	case QualityAsk:
		go func() {
			d.DialogQuality(d.FilterBlacklist(), d.control.PopupDialog, download, filer)
		}()

	case QualityBestFound:
		download("best")

	case QualityBestPossible:
		download("")
	}
}

// Enqueue enqueues an item to the download list.
//
func (d *Manager) Enqueue(call func() error) {
	d.queue <- call
}

// EnqueueStart enqueues an item and starts downloading.
//
func (d *Manager) EnqueueStart(call func() error) {
	d.Enqueue(call)
	d.Start()
}

// Start starts downloading queued items.
//
func (d *Manager) Start() {
	if d.IsActive() || len(d.queue) == 0 { // Only one worker.
		return
	}
	go func() {
		d.active = true
		if d.actionPre != nil {
			d.log.Err(d.actionPre(), "actionPre")
		}
		for call := range d.queue {
			e := call()
			d.log.Err(e, "Download")

			// if _, ok := links["error"]; !ok {
			// 	d.addHistory(links)
			// }
			// if d.onResult != nil {
			// 	d.onResult(links)
			// }
		}

		if d.actionPost != nil {
			d.log.Err(d.actionPost(), "actionPost")
		}
		d.active = false
	}()
}

func (d *Manager) runCmd(cmd *exec.Cmd) func() error {
	return func() error {
		d.cmd = cmd
		e := d.cmd.Run()
		d.cmd = nil
		return e
	}
}

//
//-----------------------------------------------------------------[ ACTIONS ]--

// OpenFolder opens the destination folder.
//
func (d *Manager) OpenFolder() {
	e := d.log.ExecAsync("xdg-open", d.Path)
	d.log.Err(e, "open folder")
}

// CancelDownload stops the current download.
//
func (d *Manager) CancelDownload() {
	if d.cmd != nil {
		d.cmd.Process.Kill()
		d.cmd = nil
	}
}

//
//------------------------------------------------------------[ DOCK HELPERS ]--

// Actions fills an applet actions list with videodl actions.
//
func (d *Manager) Actions(firstID int, actionAdd func(acts ...*cdtype.Action)) {
	d.firstID = firstID
	actionAdd(
		&cdtype.Action{
			ID:      firstID + ActionOpenFolder,
			Name:    "Open video folder",
			Icon:    "folder",
			Menu:    cdtype.MenuEntry,
			Call:    d.OpenFolder,
			Tooltip: "WTF",
		},
		&cdtype.Action{
			ID:   firstID + ActionCancelDownload,
			Name: "Cancel download",
			Icon: "edit-undo",
			Menu: cdtype.MenuEntry,
			Call: d.CancelDownload,
		},
	)
}

// MenuQuality fills an applet actions list with videodl actions.
//
func (d *Manager) MenuQuality(menu cdtype.Menuer) {
	d.control.BuildMenu(menu, []int{d.firstID + ActionOpenFolder})

	if d.active {
		d.control.BuildMenu(menu, []int{d.firstID + ActionCancelDownload})
	}
	group := 42

	sub := menu.SubMenu("Quality: "+d.Quality.String(), "video-x-generic")
	sub.AddRadioEntry(
		QualityAsk.String(),
		d.Quality == QualityAsk,
		group,
		func() { d.Quality = QualityAsk },
	)

	sub.AddRadioEntry(
		QualityBestFound.String(),
		d.Quality == QualityBestFound,
		group,
		func() { d.Quality = QualityBestFound },
	)

	sub.AddRadioEntry(
		QualityBestPossible.String(),
		d.Quality == QualityBestPossible,
		group,
		func() { d.Quality = QualityBestPossible },
	).SetTooltipText("If available, this will merge the best audio and video streams into a single file.")
}

// DialogQuality prepares a dialog to ask the desired quality and download.
//
func (d *Manager) DialogQuality(filterFormats FuncFilterFormats, callDialog FuncPopupDialog, callDL FuncDownload, f Filer) {
	title, e := f.Title()
	if !d.log.Err(e, "youtube title") {
		d.log.Info("youtube title", title)
	}

	formats, e := f.Formats()
	if d.log.Err(e, "format") || len(formats) == 0 {
		return
	}

	formats = append(formats,
		Format{Res: QualityBestFound.String(), Code: "best"},
		Format{Res: QualityBestPossible.String()},
	)

	// Reduce list of formats.
	formats = filterFormats(formats)

	lastID := len(formats) - 1
	ids := ""
	sel := []Format{}
	for i := range formats {
		form := formats[lastID-i]                     // Reverse list order (best quality first).
		if form.Note == "" || form.Note == "(best)" { // TODO: improve.
			sel = append(sel, form)
			ids = strhelp.Separator(ids, ";", strhelp.Separator(form.Ext, ": ", form.Res))
		}
	}

	e = callDialog(cdtype.DialogData{
		Message: title + "\n\nSelect quality:",
		// UseMarkup: true,
		Widget: cdtype.DialogWidgetList{
			Values:       ids,
			InitialValue: int32(0),
		},
		Buttons: "ok;cancel",
		Callback: cdtype.DialogCallbackValidInt(func(id int) {
			if id < len(sel) { // Shouldn't be out of range, but just to be safe...
				callDL(sel[id].Code)
			}
		}),
	})
	d.log.Err(e, "popup")
}
