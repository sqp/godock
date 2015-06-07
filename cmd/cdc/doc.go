/*
cdc, Cairo-Dock Control, is a tool to manage a Cairo-Dock installation.
It can also embed and manage multiple applets if compiled with their support.
Most of the commands will require an active dock to work (with Dbus API).

Usage:

	cdc command [arguments]

The commands are:

    debug       debug change the debug state of an applet
    dock        dock starts a custom version of cairo-dock with a new config GUI.
    install     install external applet
    list        list external applets
    upload      upload data to one-click hosting service
    build       build a cairo-dock package
    version     print cdc version

Use "cdc help [command]" for more information about a command.


Debug change the debug state of an applet

Usage:

	cdc debug appletname [false|no|0]

Debug change the debug state of an applet.

The first argument must be the applet name.

Options:
  false, no, 0    Disable debug.
  (default)       Enable debug.
.


Dock starts a custom version of cairo-dock with a new config GUI.

Usage:

	cdc dock

Dock starts a custom version of cairo-dock with a new GUI.

Options:
  -c          Use Cairo backend.
  -o          Use OpenGL backend.
  -O          Use OpenGL backend with indirect rendering.
              There are very few case where this option should be used.
  -A          Ask again on startup which backend to use.
  -e env      Force the dock to consider this environnement - use it with care.

  -d path     Use a custom config directory. Default: ~/.config/cairo-dock
  -S url      Address of a server with additional themes (overrides default).

  -w time     Wait for N seconds before starting; this is useful if you notice
              some problems when the dock starts with the session.
  -x appname  Exclude a given plug-in from activating (it is still loaded).
  -f          Safe mode: don't load any plug-ins.
  -W          Work around some bugs in Metacity Window-Manager
              (invisible dialogs or sub-docks)
  -l level    Log verbosity (debug,message,warning,critical,error).
              Default is warning.
  -F          Force to display some output messages with colors.
  -k          Lock the dock so that any modification is impossible for users.
  -a          Keep the dock above other windows.
  -s          Don't make the dock appear on all desktops.
  -M path     Ask the dock to load additionnal modules from this directory.
              (though it is unsafe for your dock to load unnofficial modules).

  -N          Don't start the Dbus applets service.
  -H          Http debug server: http://localhost:6987/debug/pprof
  -D          Debug mode for the go part of the code (including applets).

  -v          Print version.

This version lacks a lot of options and may not be considered usable for
everybody at the moment.
.


Install external applet

Usage:

	cdc install [-v] appletname [appletname...]

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


Upload data to one-click hosting service

Usage:

	cdc upload fileorstring

Upload send data (raw text or file) to a one-click hosting service.

If your data start with / or file:// the file content will be sent.
Otherwise, the data is sent as raw text.

Note that you must have an active instance of the NetActivity applet.
.


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


Print cdc version

Usage:

	cdc version

Version prints the cdc version.


*/
package main
