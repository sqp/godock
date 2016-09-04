// Package uptoshare uploads files to one-click hosting sites.
package uptoshare

import (
	"github.com/robfig/config" // Config parser.

	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/files"
	"github.com/sqp/godock/libs/net/upload"

	"errors"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// HistoryFile defines the name of the default history file.
//
const HistoryFile = "uptoshare_history.txt"

// FileType defines the type of a backend.
//
type FileType int

// File and backends types supported.
//
const (
	FileTypeUnknown FileType = iota
	FileTypeText
	FileTypeImage
	FileTypeVideo
	FileTypeFile
)

// String returns a human readable file type.
//
func (f FileType) String() string {
	str, ok := map[FileType]string{
		FileTypeText:  "text",
		FileTypeImage: "image",
		FileTypeVideo: "video",
		FileTypeFile:  "file",
	}[f]

	if ok {
		return str
	}

	return "unknown"
}

//
//-----------------------------------------------------------[ SERVICES HOST ]--

// Sender defines the common backend upload interface.
//
type Sender interface {
	SetName(string)
	SetConfig(*upload.Config)
	Send(r io.Reader, size int64, filename string) Links
}

// Host describes an upload service to implement a Sender.
//
type Host struct {
	upload.Poster
	Parse func(*Host, string) Links
}

// Send sends the file over the network to the hosting website.
//
func (h *Host) Send(r io.Reader, size int64, filename string) Links {
	data, e := h.Post(r, size, filename)
	if e != nil {
		return linkErr(e, h.Name())
	}
	return h.Parse(h, data)
}

//
//-------------------------------------------------------------------[ LINKS ]--

// Links contains the result data of an upload.
//
type Links map[string]string

// NewLinks creates links list.
//
// An optional argument will be used as the main "link" url value.
//
func NewLinks(link ...string) Links {
	if len(link) > 0 {
		return Links{"link": link[0]}
	}
	return make(Links)
}

// Add tests if link isn't empty before adding it to the list.
//
func (li Links) Add(key, link string) Links {
	if link != "" {
		li[key] = link
	}
	return li
}

//
//----------------------------------------------------------------[ UPLOADER ]--

// Uploader manages an upload queue to send files to one-click hosting websites.
//
type Uploader struct {
	LimitRate     int
	PostAnonymous bool
	FileForAll    bool

	// Log provides a common logger for the uptoshare service.
	Log cdtype.Logger

	queue   chan string
	active  bool
	upFile  Sender
	upImage Sender
	upText  Sender
	upVideo Sender

	actionPre  func()
	actionPost func()
	onResult   func(Links)

	history     []Links
	historyMax  int
	historyFile string
}

// New creates an Uploader.
//
func New() *Uploader {
	return &Uploader{
		queue: make(chan string, 10),
	}
}

//
//----------------------------------------------------------------[ SETTINGS ]--

// SetPreCheck sets the pre-upload action.
//
func (up *Uploader) SetPreCheck(call func()) {
	up.actionPre = call
}

// SetPostCheck sets the post-upload action.
//
func (up *Uploader) SetPostCheck(call func()) {
	up.actionPost = call
}

// SetOnResult sets the result return method.
//
func (up *Uploader) SetOnResult(call func(Links)) {
	up.onResult = call
}

// SiteImage sets the image upload backend.
//
func (up *Uploader) SiteImage(site string) error {
	return up.setSite(&up.upImage, backendImage, site)
}

// SiteText sets the text upload backend.
//
func (up *Uploader) SiteText(site string) error {
	return up.setSite(&up.upText, backendText, site)
}

// SiteVideo sets the text upload backend.
//
func (up *Uploader) SiteVideo(site string) error {
	return up.setSite(&up.upVideo, backendVideo, site)
}

// SiteFile sets the text upload backend.
//
func (up *Uploader) SiteFile(site string) error {
	e := up.setSite(&up.upFile, backendFile, site)
	up.Log.Err(e, "set SiteFile")
	return e
}

func (up *Uploader) setSite(call *Sender, backends map[string]Sender, site string) error {
	switch {
	case site == "" || site == "None":
		return nil

	case site == "-> file hosting":
		*call = up.upFile
		if up.upFile == nil {
			up.Log.Info("-> file hosting", "no upFile")
		}
		return nil
	}

	backend, ok := backends[site]
	if !ok {
		return errors.New("backend not found:" + site)
	}
	*call = backend
	backend.SetName(site)
	backend.SetConfig(&upload.Config{})
	return nil
}

//
//-----------------------------------------------------------------[ HISTORY ]--

// SetHistoryFile sets the location of the history file.
//
func (up *Uploader) SetHistoryFile(file string) {
	up.historyFile = file
	up.loadHistory()
}

// SetHistorySize sets the size of the history.
//
func (up *Uploader) SetHistorySize(nb int) {
	up.historyMax = nb
	up.trimHistory()
}

// ListHistory returns the history content.
//
func (up *Uploader) ListHistory() []Links {
	return up.history
}

func (up *Uploader) addHistory(list map[string]string) {
	up.history = append(up.history, list)
	up.trimHistory()
	up.saveHistory()
}

func (up *Uploader) trimHistory() {
	if len(up.history) > up.historyMax {
		up.history = up.history[len(up.history)-up.historyMax:]
	}
}

//
func (up *Uploader) loadHistory() error {
	up.history = nil
	c, e := config.Read(up.historyFile, config.DEFAULT_COMMENT, config.ALTERNATIVE_SEPARATOR, false, false)
	if up.Log.Err(e, "load uptoshare history") {
		return e
	}

	for _, group := range c.Sections() {
		if group != "DEFAULT" {
			links := NewLinks()
			up.history = append(up.history, links)

			opts, _ := c.Options(group)
			for _, key := range opts {
				str, _ := c.String(group, key)
				links.Add(key, str)
			}
		}
	}
	return nil
}

//
func (up *Uploader) saveHistory() {
	if _, e := os.Stat(filepath.Dir(up.historyFile)); e != nil {
		os.Mkdir(filepath.Dir(up.historyFile), os.ModePerm)
	}

	conf := config.New(config.DEFAULT_COMMENT, config.ALTERNATIVE_SEPARATOR, false, false)
	for _, hist := range up.ListHistory() {
		group := hist["link"]
		conf.AddSection(group)
		for key, link := range hist {
			conf.AddOption(group, key, link)
		}
	}
	up.Log.Err(conf.WriteFile(up.historyFile, 0644, ""), "save uptoshare history")
}

//
//------------------------------------------------------------------[ UPLOAD ]--

// Upload data to the configured server for file type.
//
func (up *Uploader) Upload(data string) {
	if data == "" {
		return
	}
	up.queue <- data

	if !up.active { // Only one worker.
		up.active = true
		if up.actionPre != nil {
			up.actionPre()
		}

		for len(up.queue) > 0 {
			data := <-up.queue
			links := up.uploadOne(data)

			if _, ok := links["error"]; !ok {
				up.addHistory(links)
			}
			if up.onResult != nil {
				up.onResult(links)
			}

		}

		if up.actionPost != nil {
			up.actionPost()
		}
		up.active = false
	}
}

func (up *Uploader) uploadOne(data string) Links {
	var fileType FileType
	var filePath string
	var isFile bool

	// File reference from desktop drop. Some cleaning to do.
	if strings.HasPrefix(data, "file://") {
		data = FileNameFromURI(data)
	}

	// Get full file path if needed.
	abs, e := filepath.Abs(data)
	if e == nil && files.IsExist(abs) {
		data = abs
	}

	// Test file and get type.
	clean := filepath.Clean(data)
	switch {
	case files.IsExist(clean): // use input as file location.
		isFile = true
		filePath = clean
		fileType = getFileType(filePath)
		if fileType == FileTypeUnknown {
			up.Log.NewWarn(filePath, "file type unknown, uploaded as 'file'")
			fileType = FileTypeFile
		}

	default: // use input as text as default/fallback.
		fileType = FileTypeText
	}

	up.Log.Debug("file upload", "type:", fileType, "path:", filePath)

	// Get the sender for the type.
	sender := up.getSender(&fileType)
	if sender == nil {
		return linkWarn("nothing to do with " + filePath)
	}

	// Get the data reader.
	var rdr io.Reader
	var size int64
	if isFile {
		var close func() error
		rdr, size, close, e = files.Reader(data)
		if e != nil {
			return linkErr(e, "open file:"+data)
		}
		defer func() { up.Log.Err(close(), "file close") }()

	} else {
		rdr = strings.NewReader(data)
		size = int64(len(data))
	}

	// And try to send.
	list := sender.Send(rdr, size, filePath)

	if len(list) == 0 {
		return linkWarn("upload: nothing returned for " + filePath)
	}

	// Should be a valid link. Add common info.
	list["file"] = filePath
	list["type"] = strconv.Itoa(int(fileType))
	list["date"] = time.Now().Format("20060102 15:04:05") // Time isn't used for now. Just display something readable. "YMD H:M:S"
	return list
}

// Get the sender for the type.
func (up *Uploader) getSender(typ *FileType) Sender {
	if up.FileForAll { // Forced upload as file.
		*typ = FileTypeFile
	}
	return map[FileType]Sender{
		FileTypeFile:  up.upFile,
		FileTypeImage: up.upImage,
		FileTypeText:  up.upText,
		FileTypeVideo: up.upVideo,
	}[*typ]
}

func getFileType(filePath string) FileType {
	mimetype := mime.TypeByExtension(filepath.Ext(filePath))
	switch {
	case strings.HasPrefix(mimetype, "video") || strings.HasSuffix(filePath, ".ogv"):
		return FileTypeVideo

	case strings.HasPrefix(mimetype, "image"):
		return FileTypeImage

	case strings.HasPrefix(mimetype, "text"):
		return FileTypeText
	}
	// Log.Info("no valid type", mimetype, filePath)
	return FileTypeUnknown
}

//
//------------------------------------------------------------------[ COMMON ]--

// findLink in str. Begin is the start of the link to match, like http://...
// End is the block to match after the link.
//
func findLink(str, begin, end string) string {
	istart := strings.Index(str, begin)

	if istart > 0 {
		iend := strings.Index(str[istart:], end)
		if iend > 0 {
			return str[istart : istart+iend]
		}
	}
	return ""
}

func findPrefix(str, prefix, suffix string) string {
	if txt := findLink(str, prefix, suffix); txt != "" {
		return txt[len(prefix):]
	}
	return ""
}

func linkErr(e error, str string) Links {
	return Links{"error": str + ": " + e.Error()}
}

func linkWarn(str string) Links {
	return Links{"error": str}
}

// FileNameFromURI strips file:// in front of a file path.
//
// TODO: check nothing is missing from the original call : g_filename_from_uri
//
func FileNameFromURI(str string) string {
	str = strings.Replace(str, "%20", " ", -1)
	return strings.TrimPrefix(str, "file://")
}

/*

//
//----------------------------------------------------------[ UPLOAD METHODS ]--

// Execute curl upload using curl command.
//
func curlExec(url string, limitRate int, fileref, filepath string, opts ...string) (string, error) {
	args := []string{
		"-s", // silent mode.
		"-L", url,
		"--connect-timeout", "5",
		"--retry", "2",
		"--limit-rate", strconv.Itoa(limitRate) + "k",
		"-H", "Expect:",
		"-F", fileref + `=@"` + filepath + `"`,
	}

	for _, opt := range opts {
		args = append(args, "-F", opt)
	}

	return curlExecArgs(args...)
}

func curlExecArgs(args ...string) (string, error) {
	body, e := exec.Command("curl", args...).CombinedOutput()

	if len(body) == 0 {
		return "", errors.New("curl output empty")
	}

	return string(body), e
}

*/
