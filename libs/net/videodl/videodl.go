// Package videodl provides a video file downloader around the youtube-dl command (only for now).
package videodl

import (
	"github.com/sqp/godock/libs/cdtype"

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

// Config defines the videodl service configuration.
// Can be used directly as applet config.
//
type Config struct {
	BackendID BackendID
	Path      string
	Quality   Quality
	TypeDL    TypeDL
	EnabledDL bool
	// EnabledWeb     bool
	// StartedWeb     bool
	JSWindowOption string

	CmdOpenDir   string
	CmdOpenWeb   string
	CmdOpenVideo string

	Blacklist []string
}

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

// Tooltip returns the tooltip text displayed for the quality setting.
//
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
//----------------------------------------------------------------[ WEBSTATE ]--

// WebState defines the state of the web service.
//
type WebState int

// TypeDL settings.
const (
	WebStateDisabled WebState = iota
	WebStateStopped
	WebStateStarted
)

// Tooltip returns the tooltip text displayed for the web service state.
//
func (s WebState) Tooltip() string {
	return "The web service allows links forwarding directly from your browser\nand the web page to edit the download history."
}

//
//------------------------------------------------------------------[ TYPEDL ]--

// TypeDL defines the media type filtering.
//
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

// Tooltip returns the tooltip text displayed for the media type filter.
//
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

// Progress will try to provide download progress report.
//
type Progress struct {
	cur int64
	max int64
}

// NewProgress creates a download progress reporter.
//
func NewProgress() *Progress { return &Progress{} }

// SetMax sets the expected size of the download.
//
func (p *Progress) SetMax(m int64) { p.max = m }

// Write implements io.Writer to count download progress.
//
func (p *Progress) Write(data []byte) (n int, e error) {
	p.cur += int64(len(data))
	return 0, nil
}

//
//------------------------------------------------------------------[ FORMAT ]--

// Video defines a video
//
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

// NewVideo creates a video for the remote url.
//
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
	// SetConfig sets the config data for the manager.
	//
	SetConfig(*Config)

	// SetEditList updates the edit list call action.
	//
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

	//-----------------------------------------------------------------[ WEB ]--

	WebRegister()
	WebUnregister()

	WebAutoStart() func()

	// SetEnabledWeb sets the web service state.
	//
	SetEnabledWeb(WebState)

	// WebURL formats the web service base url.
	//
	WebURL() string

	//------------------------------------------------------------[ DOWNLOAD ]--

	// Download downloads a video file from the server at configured quality (can be ask).
	//
	Download(url string)

	// Enqueue enqueues an item to the download list.
	//
	Enqueue(*Video)
	EnqueueAndStart(*Video)
	Start()

	//-------------------------------------------------------------[ ACTIONS ]--

	OpenFolder()
	CancelDownload()

	//--------------------------------------------------------[ DOCK HELPERS ]--

	Actions(firstID int, actionAdd func(...*cdtype.Action))
	Menu(cdtype.Menuer)
	DialogQuality(filterFormats FuncFilterFormats, callDialog FuncPopupDialog, callDL DownloadFunc, v *Video)
}
