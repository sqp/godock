// Package iplocation detects your location based on your IP.
package iplocation

import (
	"github.com/sqp/godock/libs/net/download"
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
	r := &Data{}
	e := download.JSON(URL, r)
	if e != nil {
		return nil, e
	}
	return r, nil
}
