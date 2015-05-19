package dlogbus

import (
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"

	"github.com/sqp/godock/libs/appdbus"
	"github.com/sqp/godock/libs/cdtype" // Logger type.
	"github.com/sqp/godock/libs/srvdbus"
	"github.com/sqp/godock/libs/srvdbus/dbuscommon"
)

// SrvObj is the Dbus object name for the service.
const SrvObj = "org.cairodock.DockLog"

// SrvPath is the Dbus path name for the service.
const SrvPath = "/org/cairodock/DockLog"

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

type Client struct {
	*dbuscommon.Client
}

func Action(action func(*Client) error) error {
	client, e := dbuscommon.GetClient(SrvObj, SrvPath)
	if e != nil {
		return e
	}
	return action(&Client{client}) // we have a server, launch the provided action.
}

func (client *Client) Restart() error {
	return client.Call("Restart")
}

//
//------------------------------------------------------------------[ SERVER ]--

type Server struct {
	*dbuscommon.Server
	DockArgs []string
}

// NewServer creates a dlog dbus session.
//
func NewServer(dockArgs []string, log cdtype.Logger) *Server {
	return &Server{
		Server:   dbuscommon.NewServer(SrvObj, SrvPath, log),
		DockArgs: append([]string{"dock"}, dockArgs...),
	}
}

func (o *Server) DockStart() error {
	return o.Log.ExecAsync("cdc", o.DockArgs...)
}

func (o *Server) DockStop() error {
	appdbus.DbusPathDock = "/org/cdc/Cdc"
	return srvdbus.Action((*srvdbus.Client).StopDock)
}

func (o *Server) DockRestart() error {
	o.Log.Info("build")
	e := o.Log.ExecShow("go", "install", "-tags", "dock log all", "github.com/sqp/godock/cmd/cdc")
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

func (o *Server) Restart() *dbus.Error {
	// e :=
	o.DockRestart()
	// if e != nil {
	// 	return &dbus.Error{Name: e.Error()}
	// }
	return nil
}
