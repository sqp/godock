// Package dockbus provides a Dbus client for the main dock Dbus service.
//
// Use with caution when gldi is enabled as the client and server should be on
// the same thread.
//
// Most functions in this package are unused, and some of them were made to
// support widget/confbuilder/datadbus, the GUI data source from Dbus, which is
// unused and unpublished (may be around 50% features, with missing API
// evolutions).
//
// So if you need to use this package, which isn't grouped in the applet API due
// to lack of use, and ideas to provide a common set for Dbus+Gldi, let us know
// your needs, and if current methods are fine for you.
//
package dockbus

import (
	"github.com/sqp/godock/libs/log"
	"github.com/sqp/godock/libs/packages"
	"github.com/sqp/godock/libs/srvdbus/dbuscommon"
	"github.com/sqp/godock/libs/srvdbus/dockpath" // Path to main dock dbus service.

	"errors"
	"path/filepath"
	"sort"
	"strings"
)

//------------------------------------------------------------[ DOCK ACTIONS ]--

// DockAdd adds an item to the Dock.
//
//   Launcher from desktop file:      "type":"Launcher", "config-file":"application://vlc.desktop"
//   Launcher custom:                 "type":"Launcher", "name":"Top 10", "command":"xterm -e top", "icon":"go-home.png"
//   Stack icon (SubDock container)   "type":"Stack-Icon", "name":"my sub-dock", "icon":"folder.png"
//   Separator                        "type":"Separator"
//   Module                           "type":"Module", "module":"clock"
//   MainDock                         "type":"Dock"
//
// Optional arguments:
//   Icon relative position                   "order":5
//   Icon location (main or subdock name)     "container":"_MainDock_"
//   Launcher application class               "class":"gjiten"
//
func DockAdd(args map[string]interface{}) func(*Client) error {
	return func(client *Client) error { return client.Call("Add", dbuscommon.ToMapVariant(args)) }
}

// DockRemove removes an item from the Dock.
//
//   Launcher                                  "type=Launcher & class=vlc"
//   Second main dock (and all its content)    "type=Dock & name=_MainDock_-2"
//   Module                                    "type=Module & name=clock"
//   Instance of a module                      "type=Module-Instance & config-file=clock.conf"
//
func DockRemove(arg string) func(*Client) error {
	return func(client *Client) error { return client.Call("Remove", arg) }
}

// func (cda *CdDbus) ActivateModule(module string, state bool) {
// 	busDock().Call(DbusInterfaceDock+".ActivateModule", 0, module, state)
// 	// return cda.launch(base, "ActivateModule", interface{}(module), interface{}(state))
// }

// DockReboot reload the current theme of the Dock, as if you had quitted and restarted the dock.
//
func DockReboot(client *Client) error {
	return client.Call("Reboot")
}

// DockQuit sends the Quit action to the dock dbus.
//
func DockQuit(client *Client) error {
	return client.Call("Quit")
}

// DockShow sets the dock visibility: 0 = HIDE, 1 = SHOW, 2 = TOGGLE.
// If you have several docks, it will show/hide all of them.
//
func DockShow(mode int32) func(*Client) error {
	return func(client *Client) error { return client.Call("ShowDock", mode) }
}

// ShowDesklet TODO: need to complete this part.
//
func ShowDesklet(mode int32) func(*Client) error {
	return func(client *Client) error { return client.Call("ShowDeslet", mode) }
}

// IconReload reloads an icon settings from disk.
//
//   "type=Module & name=weather"
//   "config-file=full_path_to_config_or_desktop_file"
//
func IconReload(arg string) func(*Client) error {
	return func(client *Client) error { return client.Call("Reload", arg) }
}

// DockProperties gets properties of different parts of the dock.
// API may change for this function. Need to figure out the best way to return the data.
//
//   "type=Launcher & class=firefox"
//   "type=Module"
//   "type=Module & name=clock"
//   "type=Desklet"
//
//   var name, icon string
//   for _, t := range vars {
//   	for k, v := range t {
//   		if k == "icon" {
//   			log.Info(mod, v)
//   		}
//   	}
//   }
func DockProperties(arg string) ([]map[string]interface{}, error) {
	client, e := NewClient()
	if log.Err(e, "dbus GetProperties") {
		return nil, e
	}

	var list []map[string]interface{}
	e = client.Get("GetProperties", []interface{}{&list}, arg)
	return list, e
}

// //--------------------------------------------------[ GET SPECIAL PROPERTIES ]--

// AppletAdd adds an applet instance referenced by its name.
//
func AppletAdd(name string) func(*Client) error {
	return DockAdd(map[string]interface{}{"type": "Module", "module": name})
}

// AppletRemove removes an applet instance referenced by its config file.
//
func AppletRemove(configFile string) func(*Client) error {
	return DockRemove("type=Module-Instance & config-file=" + configFile)
}

// AppletInstances asks the dock the list of active instances for an applet.
// Instances references are full paths to their config files.
//
func AppletInstances(name string) ([]string, error) {
	query := "type=Module & name=" + strings.Replace(name, "-", " ", -1)
	vars, e := DockProperties(query)
	if e != nil {
		return nil, e
	}
	if len(vars) == 0 {
		return nil, errors.New("applet not found: " + name)
	}

	val, ok := vars[0]["instances"]
	if !ok {
		return nil, errors.New("missing key instances")
	}

	instances, ok := val.([]string)
	if !ok {
		return nil, errors.New("bad type for instances list")
	}
	return instances, nil
}

//--------------------------------------------------[ GET SPECIAL PROPERTIES ]--

// Dock icon types.
const (
	IconTypeApplet    = "Applet"
	IconTypeLauncher  = "Launcher"
	IconTypeSeparator = "Separator"
	IconTypeSubDock   = "Stack-icon"
	IconTypeTaskbar   = "Taskbar"
)

// CDIcon defines a dock icon properties.
//
type CDIcon struct {
	// DisplayedName string      // name of the package

	Name     string
	Xid      uint32
	Position int32
	Type     string // Applet, Launcher, Stack-icon, Separator
	// TODO compare
	// Type          PackageType // type of package : installed, user, distant...
	QuickInfo  string
	Container  string
	Command    string
	Order      float64
	ConfigFile string
	Icon       string
	Class      string
	Module     string
}

// FormatName return the user readable name for the icon.
//
func (icon *CDIcon) FormatName() (name string) {
	switch icon.Type {
	case IconTypeApplet:
		name = icon.Module
	case IconTypeSeparator:
		name = "--------"
	default:
		name = icon.Name
		// log.DEV(name, icon)
	}
	return
}

// DefaultNameIcon returns improved name and image for the icon if possible.
// (can fix those for applets using the given list)
//
func (icon *CDIcon) DefaultNameIcon(applets map[string]*packages.AppletPackage) (name, img string) {

	name = icon.FormatName()
	if icon.Type == IconTypeApplet {
		if pack, ok := applets[icon.Module]; ok {
			name = pack.Title
		}
	}

	switch {
	case icon.Type == IconTypeApplet:
		if pack, ok := applets[icon.Module]; ok {
			img = pack.Icon
		} else {
			log.Info("module not found for icon", icon.Module)
		}

	case icon.Icon != "":
		img = icon.Icon

		// case icon.Type != IconTypeSeparator:
		// log.Info("no ICON", icon.Type, icon.FormatName())
	}
	return
}

//

// interface

// func (icon *CDIcon) MainConf() string {
// 	path, _ := packages.MainConf()
// 	return path
// }

// ConfigPath returns the full path to the icon config file.
//
func (icon *CDIcon) ConfigPath() string {
	switch icon.Type {
	case IconTypeApplet, IconTypeTaskbar:
		return icon.ConfigFile

	case IconTypeLauncher, IconTypeSeparator, IconTypeSubDock:
		if dir, e := packages.DirLaunchers(); !log.Err(e, "config launchers") {
			return filepath.Join(dir, icon.ConfigFile)
		}
	}
	return ""
}

// IsTaskbar returns whether the icon belongs to the taskbar or not.
//
func (icon *CDIcon) IsTaskbar() bool {
	return icon.Type == IconTypeTaskbar
}

//

// ListIcons asks the dock the list of active icons.
//
// TODO: add argument for advanced queries.
// would be cool to have argument list.
//
func ListIcons() (list []*CDIcon) {
	// TODO: should log or return error.
	iconsInfo, _ := DockProperties("type=Icon")

	for _, props := range iconsInfo {
		pack := &CDIcon{}
		for k, v := range props {
			switch k {
			case "name":
				pack.Name = v.(string)
			case "Xid":
				pack.Xid = v.(uint32)
			case "position":
				pack.Position = v.(int32)
			case "type":
				pack.Type = v.(string)
			case "quick-info":
				pack.QuickInfo = v.(string)
			case "container":
				pack.Container = v.(string)
			case "command":
				pack.Command = v.(string)
			case "order":
				pack.Order = v.(float64)
			case "config-file":
				pack.ConfigFile = v.(string)
			case "icon":
				pack.Icon = v.(string)
			case "class":
				pack.Class = v.(string)
			case "module":
				pack.Module = v.(string)
			default:
				log.Info("ListIcons key not found: "+k, v)
			}
		}
		// if pack.Name == "" {
		// 	log.DEV("*****NONAME", pack.Type, pack.ConfigFile)
		// 	// } else {
		// 	// 	log.DEV(pack.Name, pack.Order)
		// }
		list = append(list, pack)
	}
	return
}

//----------------------------------------------------------[ ICONS BY ORDER ]--

// IconsByOrder defines a list of icons that can be sorted on the order field.
//
type IconsByOrder []*CDIcon

// Len returns the size of the list.
//
func (a IconsByOrder) Len() int {
	return len(a)
}

// Swap swaps the position of two icons.
//
func (a IconsByOrder) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// Less compares the order of two icons.
//
func (a IconsByOrder) Less(i, j int) bool {
	return a[i].Order < a[j].Order
}

// ListIconsOrdered builds the list of dock icons sorted by container and order.
//
func ListIconsOrdered() map[string]IconsByOrder {
	list := make(map[string]IconsByOrder)
	for _, icon := range ListIcons() {
		list[icon.Container] = append(list[icon.Container], icon)
	}

	for container := range list {
		sort.Sort(IconsByOrder(list[container]))
	}
	return list
}

// func ListLaunchers() (list []*CDIcon) {
// 	for _, props := range DockProperties("type=Launcher") {
// 		pack := &CDIcon{}
// 		for k, v := range props {
// 			switch k {
// 			case "name":
// 				pack.Name = v.Value().(string)
// 			case "Xid":
// 				pack.Xid = v.Value().(uint32)
// 			case "position":
// 				pack.Position = v.Value().(int32)
// 			case "type":
// 				pack.Type = v.Value().(string)
// 			case "quick-info":
// 				pack.QuickInfo = v.Value().(string)
// 			case "container":
// 				pack.Container = v.Value().(string)
// 			case "command":
// 				pack.Command = v.Value().(string)
// 			case "order":
// 				pack.Order = v.Value().(float64)
// 			case "config-file":
// 				pack.ConfigFile = v.Value().(string)
// 			case "icon":
// 				pack.Icon = v.Value().(string)
// 			case "class":
// 				pack.Class = v.Value().(string)
// 			case "module":
// 				pack.Module = v.Value().(string)
// 			default:
// 				log.Info("ListIcons key not found: "+k, v)
// 			}
// 		}
// 		// if pack.Name == "" {
// 		// 	log.DEV("*****NONAME", pack.Type, pack.ConfigFile)
// 		// 	// } else {
// 		// 	// 	log.DEV(pack.Name, pack.Order)
// 		// }
// 		list = append(list, pack)
// 	}
// 	return
// }

// InfoApplet asks the dock all informations about an applet.
//
func InfoApplet(name string) *packages.AppletPackage {
	// TODO: should log or return error.
	vars, _ := DockProperties("type=Module & name=" + name)
	for _, props := range vars {
		return parseApplet(props)
	}
	return nil
}

// ListKnownApplets asks the dock informations about all known applets.
//
func ListKnownApplets() map[string]*packages.AppletPackage {
	list := make(map[string]*packages.AppletPackage)

	// TODO: should log or return error.
	vars, _ := DockProperties("type=Module")
	for _, props := range vars {
		pack := parseApplet(props)

		// println(pack.DisplayedName)

		if pack.DisplayedName != "" {
			list[pack.DisplayedName] = pack
			// list = append(list, pack)
			// log.Info("----------------")
			// log.DETAIL(pack)
		}
	}
	return list
}

func parseApplet(props map[string]interface{}) *packages.AppletPackage {
	pack := &packages.AppletPackage{}
	for k, v := range props {
		switch k {
		case "type": // == "Module"

		case "name":
			pack.DisplayedName = v.(string)

		case "title":
			pack.Title = v.(string)

		case "author":
			pack.Author = v.(string)

		case "instances":
			if instances, ok := v.([]string); ok {
				pack.Instances = instances
			}

		case "icon":
			pack.Icon = v.(string)

		case "description":
			pack.Description = v.(string)

		case "is-multi-instance":
			pack.IsMultiInstance = v.(bool)

		case "category":
			if cat, ok := v.(uint32); ok {
				pack.Category = int(cat)
			}

		case "preview":
			pack.Preview = v.(string)

		case "module-type":
			pack.ModuleType = int(v.(uint32))

		default:
			log.Info("parseApplet field unmatched", k, v)
		}
	}
	return pack
}

// type AppletPackage struct {
// 	DisplayedName string      // name of the package
// 	Type          PackageType // type of package : installed, user, distant...
// 	Path          string      // complete path of the package.
// 	LastModifDate string      `conf:"last modif"` // date of latest changes in the package.

// 	// On server only.
// 	CreationDate int     `conf:"creation"` // date of creation of the package.
// 	Size         float64 `conf:"size"`     // size in Mo
// 	// Rating int

// 	Author      string `conf:"author"` // author(s)
// 	Description string `conf:"description"`
// 	Category    int    `conf:"category"`

// 	Version       string `conf:"version"`
// 	ActAsLauncher bool   `conf:"act as launcher"`

// }

//
//------------------------------------------------------------------[ CLIENT ]--

// Client defines a Dbus client connected to the dock server.
//
type Client struct {
	*dbuscommon.Client
}

// NewClient creates a Client connected to the dock server.
//
func NewClient() (*Client, error) {
	cl, e := dbuscommon.GetClient(dockpath.DbusObject, string(dockpath.DbusPathDock))
	if e != nil {
		return nil, e
	}
	return &Client{cl}, nil
}

// Send sends an action to the dock server.
//
func Send(action func(*Client) error) error {
	client, e := NewClient()
	if e != nil {
		return e
	}
	return action(client) // we have a server, launch the provided action.
}

// Restart sends the Restart action to the dock dbus.
//
// func Restart(client *Client) error { return client.Call("Restart") }
