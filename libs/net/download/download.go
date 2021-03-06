// Package download gets content from internet.
package download

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"
)

// Get gets data from url.
//
func Get(uri string) ([]byte, error) {
	return Default.Get(uri)
}

// Reader gets a reader for the url.
//
func Reader(uri string) (io.ReadCloser, error) {
	return Default.Reader(uri)
}

// XML gets data from url and unmarshal as XML to data.
//
func XML(uri string, data interface{}) error {
	return Default.XML(uri, data)
}

// JSON gets data from url and unmarshal as JSON to data.
//
func JSON(uri string, data interface{}) error {
	return Default.JSON(uri, data)
}

// Default defines default settings for the package.
//
var Default = Header{}

// Header represents a HTTP downloader with custom header settings.
//
type Header map[string]string

// Reader gets a reader for the url.
//
func (h Header) Reader(uri string) (io.ReadCloser, error) {
	request, e := http.NewRequest("GET", uri, nil)
	if e != nil {
		return nil, e
	}
	for k, v := range h {
		request.Header.Add(k, v)
	}

	// Try to get data from source.
	resp, e := new(http.Client).Do(request)
	if e != nil {
		return nil, e
	}
	return resp.Body, nil
}

// Get gets data from url.
//
func (h Header) Get(uri string) ([]byte, error) {
	reader, e := h.Reader(uri)
	if e != nil {
		return nil, e
	}
	body, e := ioutil.ReadAll(reader)
	reader.Close()
	if e != nil {
		return nil, e
	}
	return body, e
}

// XML gets data from url and unmarshal as XML to data.
//
func (h Header) XML(uri string, data interface{}) error {
	body, e := h.Get(uri)
	if e != nil {
		return e
	}
	return xml.Unmarshal(body, data)
}

// JSON gets data from url and unmarshal as JSON to data.
//
func (h Header) JSON(uri string, data interface{}) error {
	body, e := h.Get(uri)
	if e != nil {
		return e
	}
	return json.Unmarshal(body, data)
}
