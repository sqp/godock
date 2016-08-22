package upload

import (
	"io"
	"io/ioutil"
	"net/url"
)

//
//------------------------------------------------------[ SERVICE INTERFACES ]--

// A Poster sends reader data to one defined hosting service.
//
type Poster interface {
	Post(r io.Reader, size int64, filename string) (string, error)
	UpBaser
}

// UpBaser provides configuraton to the Poster.
//
type UpBaser interface {
	SetName(name string)
	Name() string
	GetURL(file string) string
	SetConfig(*Config)
}

//
//------------------------------------------------------------[ SERVICE BASE ]--

// Config defines upload settings.
//
type Config struct {
	Anonymous bool
	LimitRate int
}

// UpBase provides the core for an upload service.
//
type UpBase struct {
	*Config                        // User config.
	name       string              // Website name. Filled at set backend.
	CallGetURL func(string) string // Default set to return a simple url.
}

// NewBaseURL creates an upload service base with the URI.
//
func NewBaseURL(uri string) UpBase {
	return UpBase{
		CallGetURL: func(string) string {
			return uri
		},
	}
}

// NewBaseCB creates an upload service base with the given get URI callback.
//
func NewBaseCB(cb func(file string) string) UpBase {
	return UpBase{
		CallGetURL: cb,
	}
}

// SetConfig sets the service config pointer to the backend.
//
func (sb *UpBase) SetConfig(cfg *Config) { sb.Config = cfg }

// SetName sets the service name.
// (Done at service init with its reference, so the text name is only referenced
// in the backend list)
//
func (sb *UpBase) SetName(name string) { sb.name = name }

// Name returns the service text name.
//
func (sb *UpBase) Name() string { return sb.name }

// GetURL gives the service download URL for the given filename.
//
func (sb *UpBase) GetURL(file string) string { return sb.CallGetURL(file) }

//
//------------------------------------------------------------[ HTTP REQUEST ]--

// NewRequester creates a http request Poster with the given method (POST / GET).
//
func NewRequester(method, uri string) *AsRequestHTTP {
	return &AsRequestHTTP{
		Method: method,
		UpBase: NewBaseURL(uri),
	}
}

// AsRequestHTTP implements a Poster with simple http request (POST / PUT).
//
type AsRequestHTTP struct {
	UpBase
	Method string
}

// Post send content to the hosting service and return the parsed result.
//
func (m *AsRequestHTTP) Post(r io.Reader, size int64, file string) (string, error) {
	return Reader(m.Method, m.GetURL(file), r, size)
}

//
//---------------------------------------------------------------[ MULTIPART ]--

// NewMultiparter creates a http multipart Poster with the given fields.
// Args must be provided by pairs of key, value.
//
func NewMultiparter(uri, fileRef string, args ...string) *AsMultipart {
	return &AsMultipart{
		UpBase:  NewBaseURL(uri),
		FileRef: fileRef,
		Options: argsToOptions(args),
	}
}

// AsMultipart implements a Poster with multipart POST.
//
type AsMultipart struct {
	UpBase
	FileRef string                                   // Field used by the site as file name source.
	Options map[string]string                        // Needed site options with other fields values.
	PrePost func(hm *AsMultipart, file string) error // Optional call before the Post.
}

// Post send content to the hosting service and return the parsed result.
//
func (m *AsMultipart) Post(r io.Reader, size int64, file string) (string, error) {
	if m.PrePost != nil {
		e := m.PrePost(m, file)
		if e != nil {
			return "", e
		}
	}
	return ReaderMultipart(m.GetURL(file), m.FileRef, file, m.LimitRate, r, size, m.Options)
}

//
//----------------------------------------------------------[ HTTP POST FORM ]--

// NewPostForm creates a form Poster with the given fields.
// Args must be provided by pairs of key, value.
//
func NewPostForm(uri, fileRef string, args ...string) *AsPostFormHTTP {
	return &AsPostFormHTTP{
		UpBase:     NewBaseURL(uri),
		ContentRef: fileRef,
		Options:    argsToOptions(args),
	}
}

// AsPostFormHTTP implements a Poster with http POST form fields.
//
type AsPostFormHTTP struct {
	UpBase
	ContentRef string                                          // Field used by the site as file name source.
	Options    map[string]string                               // One shot options with fields values.
	PrePost    func(hp *AsPostFormHTTP, content *string) error // Optional call before the Post.
}

// Post send content to the hosting service and return the parsed result.
//
// Used by text backends, file may sometimes be a raw string to upload instead of a file path.
//
func (m *AsPostFormHTTP) Post(r io.Reader, size int64, file string) (string, error) {
	if m.PrePost != nil {
		e := m.PrePost(m, &file)
		if e != nil {
			return "", e
		}
	}
	values := make(url.Values)
	for k, v := range m.Options {
		values.Set(k, v)
	}
	content, e := ioutil.ReadAll(r)
	if e != nil {
		return "", e
	}
	values.Set(m.ContentRef, string(content))

	return PostForm(m.GetURL(""), values)
}

//

func argsToOptions(args []string) map[string]string {
	options := make(map[string]string)
	for i := 0; i < len(args)/2; i++ {
		options[args[2*i]] = args[2*i+1]
	}
	return options
}
