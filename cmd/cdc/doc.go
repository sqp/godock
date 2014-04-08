/*
cdc, Cairo-Dock Control, is a tool to manage a Cairo-Dock installation.
It can also embed and manage multiple applets if compiled with their support.

Usage:

	cdc command [arguments]

The commands are:

    build       build a cairo-dock package
    install     install external applet
    list        list external applets
    restart     restart the dock or one or more applet
    service     start service with the dock or an applet
    version     print cdc version

Use "cdc help [command]" for more information about a command.

Build a cairo-dock package

Usage:

	cdc build [-k] [-r] [-h] target

Build builds and install a Cairo-Dock package.

Targets:
  c or core        Build the dock core.
  p or plug-ins    Build all plug-ins.
  applet name      Use the name of the applet directory in cairo-dock-plug-ins.
                   Plug-ins must have been installed first.

Flags:
  -k               Keep build dir unchanged before build (no make clean).
  -r               Reload your target after build.
  -h               Hide the make install flood if any.
  -g               Graphical mode. Use gksudo to request password.

Options:
  -s               Sources directory. Default is current dir.
  -j               Specifies the number of jobs (commands) to run simultaneously.
                   Default = all availables processors.


Install external applet

Usage:

	cdc install [-d] [-l] [-f format] [-json]

Install download and install a Cairo-Dock external applets from the repository.

Flags:
  -v               Verbose output for files extraction.


List external applets

Usage:

	cdc list [-d] [-l] [-f format] [-json]

List lists Cairo-Dock external applets with installed state.

The -d flag will only match applets found on the applet market.

The -l flag will only match applets found locally.

The -f flag specifies an specific format for the list,
using the syntax of applets template.  You can use it to
format the result the way you want. For example, to have just
the applet name: -f '{{.DisplayedName}}'.  Everything is
possible like '{{.DisplayedName}}  by {{.Author}}'. The struct
being passed to the template is:

  type AppletPackage struct {
	DisplayedName string      // name of the applet
	Author        string      // author(s)
	Description   string
	Category      int
	Version       string
	ActAsLauncher bool

	Type          PackageType // type of applet : installed, user, distant...
	Path          string      // complete path of the package.
	LastModifDate int         // date of latest changes in the package.
	Size float64              // size in Mo

	// On server only.
	CreationDate  int         // date of creation of the package.
  }

The -json flag causes the applet package data to be printed in JSON format
instead of using the template format.


Restart the dock or one or more applet

Usage:

	cdc restart [appletname...]

Restart restarts the Cairo-Dock instance or external applets.

Without any argument, all your dock will be restarted.

If one or more applet name is provided, they will be restarted.

Note that only external applets will benefit from a simple applet restart if you modified the code.


Start service with the dock or an applet

Usage:

	cdc service [applet arguments list]

Service handle the loading of the dock or its own packed applets.

Options:
  dock        Start the dock. This allow to relaunch the dock with its output
              in the same location.
  list        List active applets instances handled by the service.
  stop        Stop the dock.

The service option can also be called with options to start an applet. Those
options are provided by the dock when starting an applet. It only work for
applets actually packed in this program.

To enable it, use a shell script in place of the applet binary to forward them:
  !/bin/sh
  cdc service $*
.


Print cdc version

Usage:

	cdc version

Version prints the cdc version.


*/
package main
