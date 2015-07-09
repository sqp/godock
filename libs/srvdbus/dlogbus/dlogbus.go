// Package dlogbus provides a Dbus service (and client) for a dock external launcher.
package dlogbus

import (
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"

	"github.com/sqp/godock/libs/cdtype" // Logger type.
	"github.com/sqp/godock/libs/srvdbus/dbuscommon"
	"github.com/sqp/godock/libs/srvdbus/dockbus"
	"github.com/sqp/godock/libs/srvdbus/dockpath" // Path to main dock dbus service.
)

// SrvObj is the Dbus object name for the service.
const SrvObj = "org.cairodock.DockLog"

// SrvPath is the Dbus path name for the service.
const SrvPath = "/org/cairodock/DockLog"

// Introspec is the Dbus introspect text with service methods.
const Introspec = `
<node>
	<interface name="` + SrvObj + `">
		<method name="Restart">
		</method>
	</interface>` +
	introspect.IntrospectDataString + `
</node> `

//
//------------------------------------------------------------------[ CLIENT ]--

// Client defines a Dbus client to connect to the dlogbus server.
//
type Client struct {
	*dbuscommon.Client
}

// Action sends an action to the dlogbus server.
//
func Action(action func(*Client) error) error {
	client, e := dbuscommon.GetClient(SrvObj, SrvPath)
	if e != nil {
		return e
	}
	return action(&Client{client}) // we have a server, launch the provided action.
}

// Restart sends the Restart action to the dlogbus server.
//
func (client *Client) Restart() error {
	return client.Call("Restart")
}

//
//------------------------------------------------------------------[ SERVER ]--

// Server defines a Dbus server that manage the state of a cdc program.
//
type Server struct {
	*dbuscommon.Server
	DockArgs []string
}

// NewServer creates a dlogbus server instance with cdc command args.
// Only one can be active.
//
func NewServer(dockArgs []string, log cdtype.Logger) *Server {
	return &Server{
		Server:   dbuscommon.NewServer(SrvObj, SrvPath, log),
		DockArgs: append([]string{"dock"}, dockArgs...),
	}
}

// DockStart starts the dock.
//
func (o *Server) DockStart() error {
	return o.Log.ExecAsync("cdc", o.DockArgs...)
}

// DockStop stops the dock.
//
func (o *Server) DockStop() error {
	dockpath.DbusPathDock = "/org/cdc/Cdc"
	return dockbus.Send(dockbus.DockQuit)
}

// DockRestart restarts the dock.
//
func (o *Server) DockRestart() error {
	o.Log.Info("build")
	e := o.Log.ExecShow("make", "dock")
	// e := o.Log.ExecShow("go", "install", "-tags", "dock log all", "github.com/sqp/godock/cmd/cdc")
	if e != nil {
		return e
	}

	e = o.DockStop()
	// if e != nil {
	// 	return e
	// }
	o.Log.Err(e, "StopDock")

	return o.DockStart()
}

//
//----------------------------------------------------------------[ DBUS API ]--

// Restart restarts the dock.
//
func (o *Server) Restart() *dbus.Error {
	// e :=
	o.DockRestart()
	// if e != nil {
	// 	return &dbus.Error{Name: e.Error()}
	// }
	return nil
}
