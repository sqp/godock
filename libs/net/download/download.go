// Package download gets content from internet.
package download

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
)

// XML gets data from url and unmarshal as XML to data.
//
func XML(URL string, data interface{}) error {
	resp, e := http.Get(URL)
	if e != nil {
		return e
	}
	body, e := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if e != nil {
		return e
	}

	return xml.Unmarshal(body, data)
}
