// Package dbuscommon provides a common dbus server and client base to extend.
package dbuscommon

import (
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"

	"github.com/sqp/godock/libs/cdtype" // Logger type.

	"errors"
)

//
//------------------------------------------------------------------[ CLIENT ]--

// Client is a Dbus client to connect to the internal Dbus server.
//
type Client struct {
	dbus.BusObject
	srvObj string
}

// GetClient return a connection to the active instance of the internal Dbus
// service if any. Return nil, nil if none found.
// InterfacePath is an optional string to provide if the object use an interface
// path different from SrvObj
//
func GetClient(SrvObj, SrvPath string, InterfacePath ...string) (*Client, error) {
	conn, ec := dbus.SessionBus()
	if ec != nil {
		return nil, ec
	}

	reply, e := conn.RequestName(SrvObj, dbus.NameFlagDoNotQueue)
	if e != nil {
		return nil, e
	}
	conn.ReleaseName(SrvObj)

	if reply == dbus.RequestNameReplyPrimaryOwner { // no active instance.
		return nil, errors.New("no service found")
	}

	if len(InterfacePath) == 0 { // Set default interface path = object name.
		InterfacePath = []string{SrvObj}
	}

	// Found active instance, return client.
	return &Client{conn.Object(SrvObj, dbus.ObjectPath(SrvPath)), InterfacePath[0]}, nil
}

// Call calls a method on a Dbus object.
//
func (cl *Client) Call(method string, args ...interface{}) error {
	return cl.BusObject.Call(cl.srvObj+"."+method, 0, args...).Err
}

func (cl *Client) Get(method string, args ...interface{}) ([]interface{}, error) {
	call := cl.BusObject.Call(cl.srvObj+"."+method, 0, args...)
	return call.Body, call.Err
}

// func (cl *Client) Go(method string, args ...interface{}) error {
// 	return cl.BusObject.Go(SrvObj+"."+method, dbus.FlagNoReplyExpected, nil, args...).Err
// }

//
//------------------------------------------------------------------[ SERVER ]--

// Server is a Dbus server with applets service management.
//
type Server struct {
	Conn   *dbus.Conn          // Dbus connection.
	Events <-chan *dbus.Signal // Dbus incoming signals channel.
	Log    cdtype.Logger

	srvObj  string
	srvPath string
}

// NewServer creates a Dbus service.
//
func NewServer(srvObj, srvPath string, log cdtype.Logger) *Server {
	conn, c, e := SessionBus()
	if log.Err(e, "DBus Connect") {
		return nil
	}

	load := &Server{
		Conn:    conn,
		Events:  c,
		Log:     log,
		srvObj:  srvObj,
		srvPath: srvPath,
	}

	return load
}

// Start will try to start and manage the applets server.
// You must provide the applet arguments used to launch the applet.
// If a server was already active, the applet start request is forwarded and
// no loop will be started, the function just return with the error if any.
//
func (load *Server) Start(obj interface{}, introspec string) (bool, error) {
	reply, e := load.Conn.RequestName(load.srvObj, dbus.NameFlagDoNotQueue)
	if e != nil {
		return false, e
	}

	if reply != dbus.RequestNameReplyPrimaryOwner {
		return false, nil
	}

	// logger.Err(Export(s, load.conn), "export")

	// Everything OK, we can register our Dbus methods.
	e = load.Conn.Export(obj, dbus.ObjectPath(load.srvPath), load.srvObj)
	load.Log.Err(e, "register service object")

	e = load.Conn.Export(introspect.Introspectable(introspec), dbus.ObjectPath(load.srvPath), "org.freedesktop.DBus.Introspectable")
	load.Log.Err(e, "register service introspect")

	return true, nil
}

//
//------------------------------------------------------------------[ COMMON ]--

// SessionBus creates a Dbus session with a listening chan.
//
func SessionBus() (*dbus.Conn, chan *dbus.Signal, error) {
	conn, e := dbus.SessionBus()
	if e != nil {
		return nil, nil, e
	}

	c := make(chan *dbus.Signal, 10)
	conn.Signal(c)
	return conn, c, nil
}

// EavesDrop registers to receive Dbus events for custom parsing.
//
func EavesDrop(match string) (chan *dbus.Message, error) {
	conn, e := dbus.SessionBus()
	if e != nil {
		return nil, e
	}
	e = conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, match).Err
	if e != nil {
		return nil, e
	}
	c := make(chan *dbus.Message, 10)
	conn.Eavesdrop(c)
	return c, nil
}

// ToMapVariant recasts a list of args to map[string]dbus.Variant as requested by the DBus API.
//
func ToMapVariant(input map[string]interface{}) map[string]dbus.Variant {
	vars := make(map[string]dbus.Variant)
	for k, v := range input {
		vars[k] = dbus.MakeVariant(v)
	}
	return vars
}

// ToMapInterface recasts a map of dbus.Variant to a map of interface.
//
func ToMapInterface(input map[string]dbus.Variant) map[string]interface{} {
	out := make(map[string]interface{}, len(input))
	for i, v := range input {
		out[i] = v.Value()
	}
	return out
}
