package videodl

import (
	"github.com/sqp/godock/libs/net/websrv"

	"encoding/json"
	"html"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Entry defines a list entry to provide as json
//
type Entry struct {
	Name   string // name of the object
	IsDir  bool
	Mode   os.FileMode
	Viewed bool
	URL    string
	Fail   bool
}

func (m *Manager) WebRegister() {
	e := websrv.Service.Register(WebPath, m.ServeHTTP, m.log)
	m.log.Err(e, "WebRegister")
}

func (m *Manager) WebUnregister() {
	e := websrv.Service.Unregister(WebPath)
	m.log.Err(e, "WebUnregister")
}

func (m *Manager) WebStart() {
	if !m.EnabledWeb || m.StartedWeb {
		return
	}
	e := websrv.Service.Start(WebPath)
	if !m.log.Err(e, "WebStart") {
		m.StartedWeb = true
	}
}

func (m *Manager) WebStop() {
	if !m.StartedWeb {
		return
	}

	e := websrv.Service.Stop(WebPath)
	if !m.log.Err(e, "WebStop") {
		m.StartedWeb = false
	}
}

func (m *Manager) WebAutoStart() func() {
	if !m.StartedWeb {
		e := websrv.Service.Start(WebPath)
		m.log.Err(e, "WebAutoStart: start")
	}
	return func() {
		if !m.StartedWeb {
			e := websrv.Service.Stop(WebPath)
			m.log.Err(e, "WebAutoStart: stop")
		}
	}
}

func (m *Manager) SetStartedWeb(b bool) {
	if b {
		m.WebStart()
	} else {
		m.WebStop()
	}
}
func (m *Manager) SetJSWindowOption(str string) {
	m.JSWindowOption = str
}

func (m *Manager) WebURL() string {
	url := "http://" // TODO: need https
	if websrv.Service.Host == "" {
		url += "localhost"
	}
	return url + websrv.Service.URL() + "/" + WebPath
}

// type HandlerFunc func( http.ResponseWriter, *http.Request)

func (m *Manager) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	m.log.Info(req.URL.String())
	url := req.URL.String()[len(WebPath)+1:]
	switch {
	case url == "" || url == "/":
		m.webVideoIndex(rw, req)

	case strings.HasPrefix(url, "/add"): // remote add video from js link.
		m.webVideoAdd(rw, req)

	case strings.HasPrefix(url, "/linkadd"): // js script to add video (to copy to address bar).
		m.webLinkAdd(rw, req)

	case strings.HasPrefix(url, "/viewfile"): // stream video over http.
		m.webVideoViewFile(rw, req)

	case strings.HasPrefix(url, "/fileinfo"):
		m.webVideoFileInfo(rw, req)

	case strings.HasPrefix(url, "/viewed"):
		m.webVideoViewed(rw, req)

	case strings.HasPrefix(url, "/editquality"):
		m.webVideoEditQuality(rw, req)

	case strings.HasPrefix(url, "/openvideo"):
		m.webVideoOpenVideo(rw, req)

	case strings.HasPrefix(url, "/listremoveone"):
		m.webVideoRemoveOne(rw, req, &m.history.List)

	case strings.HasPrefix(url, "/list"):
		m.webVideoList(rw, req, m.history.List)

	case strings.HasPrefix(url, "/queue"):
		m.webVideoList(rw, req, m.history.Queue)

	default:
		m.log.NewErr(req.URL.String(), "webservice: bad address call=")
	}
}

// webVideoAdd handles the folders creation.
func (m *Manager) webVideoAdd(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Query()["url"]
	if len(url) == 0 {
		m.log.NewErr("url missing", "webVideoAdd")
		return
	}

	m.log.Debug("webVideoAdd", url[0])
	m.Download(url[0])
	w.Write([]byte(`<script language="javascript" type="text/javascript"> window.close(); </script>`))
}

func (m *Manager) webLinkAdd(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`javascript:u=document.location.href;t=document.title;` +
		`dock="http://localhost:` + strconv.Itoa(websrv.Service.Port) + `/video/add/?title="+escape(t)+'&url='+escape(u);` +
		`void(window.open(dock,'_blank','` + m.JSWindowOption + "'));"))
}

func (m *Manager) webVideoIndex(w http.ResponseWriter, r *http.Request) {
	// http.ServeFile(w, r, indexHTML())
	w.Write(indexHTML())
	m.log.Info("index called")
}

func (m *Manager) webVideoViewFile(w http.ResponseWriter, r *http.Request) {
	vid := m.findVideo(r)
	if vid == nil {
		return
	}
	m.log.Info("viewfile", vid.Name)

	http.ServeFile(w, r, m.FilePath(*vid))
}

func (m *Manager) webVideoList(w http.ResponseWriter, r *http.Request, list []*Video) {
	entries := make([]Entry, len(list), len(list))
	for k, v := range list {
		entries[k].Name = v.Name
		entries[k].IsDir = false // v.IsDir()
		entries[k].Mode = 0644
		entries[k].Viewed = v.Viewed
		entries[k].URL = html.EscapeString(v.URL)
		entries[k].Fail = v.Fail
	}
	e := json.NewEncoder(w).Encode(&entries)
	m.log.Err(e, "webservice: json.Encode videodl list")
}

func (m *Manager) webVideoRemoveOne(w http.ResponseWriter, r *http.Request, list *[]*Video) {
	vid := m.findVideo(r)
	if vid == nil {
		m.log.NewErr("listremoveone = not found", "webservice", r.URL.String())
		return
	}

	e := m.history.Remove(vid, list)
	m.log.Err(e, "listremoveone", "webservice")
}

func (m *Manager) webVideoEditQuality(w http.ResponseWriter, r *http.Request) {
	m.log.Info("video: editquality", r.URL)
	vid := m.findVideo(r)
	if vid == nil {
		return
	}

	m.getQuality(vid)
}

func (m *Manager) webVideoFileInfo(w http.ResponseWriter, r *http.Request) {
	m.log.Debug("video: fileinfo", r.URL)
	vid := m.findVideo(r)
	if vid == nil {
		return
	}

	args := argsFileInfo(r)

	strviewed, ok := args["viewed"]
	if ok {
		view := false
		if strviewed == "true" {
			view = true
		}
		vid.Viewed = view
	}

	e := json.NewEncoder(w).Encode(vid)
	m.log.Err(e, "webservice: json.Encode videodl fileinfo")
}

func argsFileInfo(r *http.Request) map[string]string {
	ret := make(map[string]string)
	for _, k := range []string{"viewed"} {
		v := r.URL.Query()[k]
		if len(v) != 0 {
			ret[k] = v[0]
		}
	}
	return ret
}

func (m *Manager) webVideoViewed(w http.ResponseWriter, r *http.Request) {
	m.log.Debug("video: viewed", r.URL)
	vid := m.findVideo(r)
	if vid == nil {
		return
	}
	strviewed := r.URL.Query()["viewed"]
	if len(strviewed) == 0 {
		m.log.NewErr("viewed missing", "webVideoViewed")
		return
	}
	view := false
	if strviewed[0] == "true" {
		view = true
	}

	vid.Viewed = view

	if vid.Viewed {
		w.Write([]byte("true"))
	} else {
		w.Write([]byte("false"))
	}
}
func (m *Manager) webVideoOpenVideo(w http.ResponseWriter, r *http.Request) {
	m.log.Debug("video: OpenVideo", r.URL)
	vid := m.findVideo(r)
	if vid == nil {
		return
	}
	m.log.ExecAsync(m.cmdOpenVideo, m.FilePath(*vid))
}

func (m *Manager) findVideo(r *http.Request) *Video {

	m.log.Debug("video: viewed", r.URL)
	urls := r.URL.Query()["url"]
	if len(urls) == 0 {
		m.log.NewErr("url missing", "webVideoViewed")
		return nil
	}

	url := html.UnescapeString(urls[0])

	vid := m.history.Find(url)
	if vid == nil {
		m.log.NewErr("no ref found", "webVideoViewed")
		return nil
	}
	path := r.URL.Query()["path"]
	if len(path) == 0 {
		m.log.NewErr("path missing", "webVideoViewed")
		return nil
	}
	return vid
}
