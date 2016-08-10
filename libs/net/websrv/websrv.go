// Package websrv manages the default optional web service for the program with variable subservices.
package websrv

import (
	"github.com/braintree/manners"

	"github.com/sqp/godock/libs/cdtype"

	"errors"
	"net/http"
	"strconv"
	"strings"
)

const defaultPort = 15610

var (
	// Service manages the default web service for the program.
	Service = NewSrv(defaultPort)
)

type service struct {
	started bool
	call    http.HandlerFunc
}

// Srv defines a web service handling subservices.
//
type Srv struct {
	list map[string]*service
	Host string
	Port int
}

// NewSrv creates a new web service managing multiple
// You better use the already created Service var.
//
func NewSrv(port int) *Srv {
	return &Srv{
		list: make(map[string]*service),
		Port: port,
	}
}

// URL returns the service location: host:port
//
func (s *Srv) URL() string {
	return s.Host + ":" + strconv.Itoa(s.Port)
}

// Register registers the service matching the given prefix key.
//
func (s *Srv) Register(key string, call http.HandlerFunc, log cdtype.Logger) error {
	_, ok := s.list[key]
	if ok {
		return errors.New("register webservice: key already exist")
	}
	s.list[key] = &service{call: call}
	return nil
}

// Unregister unregisters the service matching the given prefix key.
//
func (s *Srv) Unregister(key string) error {
	_, ok := s.list[key]
	if !ok {
		return errors.New("unregister webservice: key not found")
	}

	e := s.Stop(key)
	if e != nil {
		return e
	}

	delete(s.list, key)
	return nil
}

// Start starts the registered service matching the given prefix key.
//
func (s *Srv) Start(key string) error {
	app, ok := s.list[key]
	if !ok {
		return errors.New("start webservice: key not found")
	}

	started := false
	for _, v := range s.list {
		started = started || v.started
	}

	app.started = true

	if started {
		return nil
	}

	manners.NewServer()
	go manners.ListenAndServe(s.URL(), s)
	return nil
}

// Stop stops the registered service matching the given prefix key.
//
func (s *Srv) Stop(key string) error {
	app, ok := s.list[key]
	if !ok {
		return errors.New("stop webservice: key not found")
	}

	app.started = false

	started := false
	for _, v := range s.list {
		started = started || v.started
	}

	if !started {
		manners.Close()
	}
	return nil
}

// type HandlerFunc func( http.ResponseWriter, *http.Request)

// ServeHTTP forwards the web call to the services matching the url prefix.
//
func (s *Srv) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	url := req.URL.String()
	for prefix, app := range s.list {
		if url == "/"+prefix || strings.HasPrefix(url, "/"+prefix+"/") {
			if app.call != nil {
				app.call(rw, req)
			}
			return
		}
	}
	// s.log.Err(req.URL.String(), "WebService: Wrong address")
}
