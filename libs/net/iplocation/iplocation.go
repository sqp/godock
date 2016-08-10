// Package iplocation detects your location based on your IP.
package iplocation

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// URL points to the location of the IP locale info API.
//
const URL = "http://ip-api.com/json"

// Data will hold the result of the query to get the IP address of the caller.
//
type Data struct {
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	ISP         string  `json:"isp"`
	ORG         string  `json:"org"`
	AS          string  `json:"as"`
	Message     string  `json:"message"`
	Query       string  `json:"query"`
}

// Get will get the location details based on current IP location.
//
func Get() (*Data, error) {
	response, e := http.Get(URL)
	if e != nil {
		return nil, e
	}
	defer response.Body.Close()

	result, e := ioutil.ReadAll(response.Body)
	if e != nil {
		return nil, e
	}

	r := &Data{}
	e = json.Unmarshal(result, &r)
	if e != nil {
		return nil, e
	}
	return r, nil
}
