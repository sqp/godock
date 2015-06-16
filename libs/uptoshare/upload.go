// Package uptoshare uploads files to one-click hosting sites.
package uptoshare

/*
// #include <stdlib.h>                 // free
// #include <glib-2.0/glib/gstdio.h>   // g_filename_from_uri
// #cgo pkg-config: glib-2.0
import "C"
*/

import (
	"github.com/robfig/config" // Config parser.

	"github.com/sqp/godock/libs/cdtype"

	"mime"
	"path"
	"strconv"
	"strings"

	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"

	// "unsafe"
	// "path/filepath"
	// "bytes"
	// "fmt"
	// "io"
	// "mime/multipart"
)

/*
// FileNameFromURI is a wrapper around g_filename_from_uri to convert a filepath to UTF-8.
//
func FileNameFromURI(str string) string {
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))
	cFilePath := C.g_filename_from_uri((*C.gchar)(cstr), nil, nil)
	defer C.free(unsafe.Pointer(cFilePath))
	return C.GoString((*C.char)(cFilePath))
}
*/

// FileNameFromURI strips file:// in front of a file path.
//
// TODO: check that we can safely use that really simplified version, removing
// C and glib-2.0 dependency.
//
func FileNameFromURI(str string) string {
	return strings.TrimPrefix(str, "file://")
}

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

// CallUpload defines the format to use for a backend uploader.
//
type CallUpload func(filePath, localDir string, anonymous bool, limitRate int) Links

var (
	backendImage = map[string]CallUpload{
		"Imagebin.ca":         ImageBinCa,
		"ImageShack.us":       ImageShackUs,
		"Imgur.com":           ImgurCom,
		"pix.Toile-Libre.org": PixToileLibreOrg,
		"Postimage.org":       PostimageOrg,
		"Uppix.com":           UppixCom,
	}

	backendText = map[string]CallUpload{
		"Codepad.org":          CodePadOrg,
		"Pastebin.com":         PasteBinCom,
		"Pastebin.mozilla.org": PasteBinMozillaOrg,
		"Paste-ubuntu.com":     PasteUbuntuCom,
	}

	backendVideo = map[string]CallUpload{
		"VideoBin.org": VideoBinOrg,
	}

	backendFile = map[string]CallUpload{
		"dl.free.fr": DlFreeFr,
	}
)

// Links contains the result data of an upload.
//
type Links map[string]string

// NewLinks creates an empty links list.
//
func NewLinks() Links {
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
	upFile  CallUpload
	upImage CallUpload
	upText  CallUpload
	upVideo CallUpload

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
	return setSite(&up.upImage, backendImage, site)
}

// SiteText sets the text upload backend.
//
func (up *Uploader) SiteText(site string) error {
	return setSite(&up.upText, backendText, site)
}

// SiteVideo sets the text upload backend.
//
func (up *Uploader) SiteVideo(site string) error {
	return setSite(&up.upVideo, backendVideo, site)
}

// SiteFile sets the text upload backend.
//
func (up *Uploader) SiteFile(site string) error {
	return setSite(&up.upFile, backendFile, site)
}

func setSite(call *CallUpload, backends map[string]CallUpload, site string) error {
	backend, ok := backends[site]
	if !ok {
		return errors.New("backend not found:" + site)
	}
	*call = backend
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
	if _, e := os.Stat(path.Dir(up.historyFile)); e != nil {
		os.Mkdir(path.Dir(up.historyFile), os.ModePerm)
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

	switch {
	case strings.HasPrefix(data, "file://"): // input is a file reference.
		isFile = true
		filePath = FileNameFromURI(data)
		fileType = getFileType(filePath)

	case data[0] == '/': // use input as file location.
		isFile = true
		filePath = data
		fileType = getFileType(filePath)

	default: // use input as text.
		fileType = FileTypeText
	}

	if fileType == FileTypeUnknown {
		fileType = FileTypeFile
		// 	cd_debug ("we'll consider this as an archive.");
	}

	up.Log.Debug("file upload", "type:", fileType, "path:", filePath)
	var call func(string, string, bool, int) Links

	if up.FileForAll { // Forced upload as file.
		call = up.upFile
		if fileType == FileTypeText && !isFile {
			return linkWarn("text entry as FFA option missing\nYou have to uncheck the send all as file option to send raw text")
		}
		fileType = FileTypeFile

	} else {
		switch fileType {
		case FileTypeFile:
			call = up.upFile

		case FileTypeImage:
			call = up.upImage

		case FileTypeText:
			if isFile {
				bytes, e := ioutil.ReadFile(filePath)
				if e != nil {
					return linkErr(e, "can't read file")
				}
				data = string(bytes)
			}
			filePath = data
			call = up.upText
			if filePath == "" {
				return linkWarn("FileTypeText: can't upload an empty string")
			}

		case FileTypeVideo:
			call = up.upVideo
		}
	}

	if call == nil {
		return linkWarn("nothing to do with " + filePath)
	}

	list := call(filePath, "", up.PostAnonymous, up.LimitRate)
	if len(list) == 0 {
		return linkWarn("upload: nothing returned for " + filePath)
	}

	list["file"] = filePath
	list["type"] = strconv.Itoa(int(fileType))
	list["date"] = time.Now().Format("20060102 15:04:05") // Time isn't used for now. Just display something readable. "YMD H:M:S"
	return list
}

func getFileType(filePath string) FileType {
	mimetype := mime.TypeByExtension(path.Ext(filePath))
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

//
//----------------------------------------------------------[ UPLOAD METHODS ]--

// Execute curl upload using curl command.
//
func curlExec(url string, limitRate int, fileref, filepath string, opts []string) (string, error) {
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

// Creates a new http POST request with optional extra params.
//
func postSimple(url string, values url.Values) (string, error) {
	r, e := http.PostForm(url, values)
	if e != nil {
		// Log.Info("error posting %s", e)
		return "", e
	}
	defer r.Body.Close()
	body, er := ioutil.ReadAll(r.Body)
	if er != nil {
		// Log.Info("error reading %s", er)
		return "", er
	}

	if len(body) == 0 {
		return "", errors.New("POST output empty")
	}

	return string(body), nil
}

//
//------------------------------------------------------------------[ UNUSED ]--

// func postFile(filename, target_url, opt string) (*http.Response, error) {
// 	body_buf := bytes.NewBufferString("")
// 	body_writer := multipart.NewWriter(body_buf)

// 	// use the body_writer to write the Part headers to the buffer
// 	_, err := body_writer.CreateFormFile("upfile", filename)
// 	if err != nil {
// 		fmt.Println("error writing to buffer")
// 		return nil, err
// 	}

// 	// the file data will be the second part of the body
// 	fh, err := os.Open(filename)
// 	if err != nil {
// 		fmt.Println("error opening file")
// 		return nil, err
// 	}
// 	// need to know the boundary to properly close the part myself.
// 	boundary := body_writer.Boundary()
// 	// close_string := fmt.Sprintf("\r\n--%s--\r\n", boundary)
// 	close_buf := bytes.NewBufferString(fmt.Sprintf("\r\n--%s--\r\n", boundary))

// 	// use multi-reader to defer the reading of the file data until writing to the socket buffer.
// 	request_reader := io.MultiReader(body_buf, fh, close_buf)
// 	fi, err := fh.Stat()
// 	if err != nil {
// 		fmt.Printf("Error Stating file: %s", filename)
// 		return nil, err
// 	}
// 	req, err := http.NewRequest("POST", target_url, request_reader)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Set headers for multipart, and Content Length
// 	req.Header.Add("Content-Type", "multipart/form-data;boundary="+boundary+opt)
// 	req.ContentLength = fi.Size() + int64(body_buf.Len()) + int64(close_buf.Len())

// 	return http.DefaultClient.Do(req)
// }

// // Creates a new file upload http POST request with optional extra params.
// //
// func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
// 	file, err := os.Open(path)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer file.Close()

// 	body := &bytes.Buffer{}
// 	writer := multipart.NewWriter(body)
// 	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
// 	if err != nil {
// 		return nil, err
// 	}
// 	_, err = io.Copy(part, file)

// 	for key, val := range params {
// 		_ = writer.WriteField(key, val)
// 	}
// 	err = writer.Close()
// 	if err != nil {
// 		return nil, err
// 	}

// 	return http.NewRequest("POST", uri, body)
// }
