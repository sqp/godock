// Package websrv manages the default optional web service for the program with variable subservices.
package websrv

import (
	"errors"

	"github.com/braintree/manners"      // Restartable webserver.
	_ "github.com/mkevac/debugcharts"   // Register monitoring charts.
	"github.com/sqp/godock/libs/cdtype" // Logger type.
	// Secure handling crashs.
	"fmt"
	"net/http"
	"net/http/pprof" // Web service for pprof.
	"strconv"
	"strings"
)

const (
	// PathPprof is the registered url for the monitoring pprof service.
	PathPprof = "debug/pprof"

	// PathCharts is the registered url for the monitoring charts service.
	PathCharts = "debug/charts"
)

var (
	// DefaultHost is the default hostname used for the web service.
	DefaultHost = "localhost"

	// DefaultPort is the default port used for the web service.
	DefaultPort = 15610

	// Service manages the default web service for the program.
	Service *Srv
)

// Init creates the default Service with a logger for monitoring.
//
func Init(log cdtype.Logger) {
	Service = NewSrv(DefaultHost, DefaultPort, log)
}

type service struct {
	started bool
	call    http.HandlerFunc
	log     cdtype.Logger
}

// Srv defines a web service handling subservices.
//
type Srv struct {
	Host string // Where to
	Port int    // Listen.

	list map[string]*service // Registered services.
	log  cdtype.Logger       // Logger.
}

// NewSrv creates a new web service managing multiple
// You better use the already created Service var.
//
func NewSrv(host string, port int, log cdtype.Logger) *Srv {
	srv := &Srv{
		Host: host,
		Port: port,
		list: make(map[string]*service),
		log:  log,
	}
	srv.Register(PathPprof, pprof.Index, log)
	srv.Register(PathCharts, http.DefaultServeMux.ServeHTTP, log)

	return srv
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
		return fmt.Errorf("register web service: key already exist: %s", key)
	}
	if call == nil {
		return errors.New("register web service: callback is nil")
	}
	s.list[key] = &service{
		call: call,
		log:  log,
	}
	return nil
}

// Unregister unregisters the service matching the given prefix key.
//
func (s *Srv) Unregister(key string) error {
	_, ok := s.list[key]
	if !ok {
		return fmt.Errorf("register web service: key not found: %s", key)
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
	svc, ok := s.list[key]
	if !ok {
		return fmt.Errorf("start web service: key not found: %s", key)
	}

	isListening := s.needListen()
	svc.started = true

	if isListening {
		return nil
	}
	s.log.GoTry(func() {
		e := manners.ListenAndServe(s.URL(), s)
		svc.log.Err(e, "start web server")
	})

	return nil
}

// Stop stops the registered service matching the given prefix key.
//
func (s *Srv) Stop(key string) error {
	svc, ok := s.list[key]
	if !ok {
		return fmt.Errorf("stop web service: key not found: %s", key)
	}
	if !svc.started {
		return nil
	}

	svc.started = false

	if !s.needListen() {
		manners.Close()
	}
	return nil
}

// ServeHTTP forwards the web call to the services matching the url prefix.
//
func (s *Srv) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	url := req.URL.String()
	for prefix, svc := range s.list {
		if strings.HasPrefix(url, "/"+prefix) {
			if svc.started {
				svc.log.Debug("served", req.RemoteAddr+" asked "+url+"  UA="+req.UserAgent())
				defer s.log.Recover()
				svc.call(rw, req)
			}
			// Should we log refused because inactive?
			return
		}
	}
	s.log.Debug("refused", req.RemoteAddr+" asked "+url+"  UA="+req.UserAgent())
}

// IsMonitored returns whether monitoring pages are active or not.
//
func (s *Srv) IsMonitored() bool {
	return s.list["debug/pprof"].started
}

// SetMonitored sets if monitoring pages are active or not.
//
func (s *Srv) SetMonitored(setActive bool) {
	svc := s.list["debug/pprof"]

	svc.log.Debug("websrv.SetMonitored", s.Host+":", s.Port, s.IsMonitored())
	switch {
	case setActive && !svc.started:
		s.Start("debug/pprof")
		s.Start("debug/charts")

	case !setActive && svc.started:
		s.Stop("debug/pprof")
		s.Stop("debug/charts")

	default:
		svc.log.Errorf("webservice.SetMonitored", "current state:%v  new state:%v", setActive, svc.started)
	}
}

func (s *Srv) needListen() bool {
	for _, service := range s.list {
		if service.started {
			return true
		}
	}
	return false
}
