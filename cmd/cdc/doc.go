/*
cdc starts a custom version of cairo-dock with a new config GUI.

Usage:

	cdc command [arguments]

The commands are:

    debug       debug an active applet
    external    manage external applets
    upload      upload files or text to one-click hosting services
    build       build cairo-dock sources
    version     print cdc version

Use "cdc help [command]" for more information about a command.

Without command, the dock will be started with those arguments:

Backend:
  -c          Use Cairo backend.
  -o          Use OpenGL backend.
  -O          Use OpenGL backend with indirect rendering.
              There are very few case where this option should be used.
  -A          Ask again on startup which backend to use.
  -W          Work around some bugs in Metacity Window-Manager
              (invisible dialogs or sub-docks)

Desktop:
  -k          Lock the dock so that any modification is impossible for users.
  -e env      Force the dock to consider this environnement - use it with care.
  -a          Keep the dock above other windows.
  -s          Don't make the dock appear on all desktops.

Paths override:
  -d path     Use a custom config directory. Default: ~/.config/cairo-dock
  -S url      Address of a server with additional themes (overrides default).
  -M path     Ask the dock to load additionnal modules from this directory.
              (though it is unsafe for your dock to load unnofficial modules).

Debug:
  -w time     Wait for N seconds before starting; this is useful if you notice
              some problems when the dock starts with the session.
  -x appname  Exclude a given plug-in from activating (it is still loaded).
  -f          Safe mode: don't load any plug-ins.
  -l level    Log verbosity: debug,message,warning,critical,error (def=warning).
  -F          Force to display some output messages with colors.
  -D          Debug mode for the go part of the code (including applets).
  -N          Don't start Dbus and go applets services.
  -H          Http debug server: http://localhost:6987/debug/pprof

Versions:
  -v          Print gldi version.
  -vv         Print all versions.

This version still lacks some options and may not be considered usable for
everybody at the moment. But it also needs to be tested now.


Debug an active applet

Usage:

	cdc debug appletname [false|no|0]

Debug sets the debug state of an applet.

The first argument must be the applet name.

Options:
  false, no, 0    Disable debug.
  (default)       Enable debug.


Manage external applets

Usage:

	cdc external [-d path] [-r] [appletname...]

External lists, installs or removes Cairo-Dock external applets.

The action depends if applet names are provided, and on the -r flag:
  no applet name        Display the list of external applets.
  one or more           Install applet(s).
  one or more and -r    Remove applet(s).

Common flags:
  -d path      Use a custom config directory. Default: ~/.config/cairo-dock

Install flags:
  -r           Remove applets instead of install.
  -v           Verbose output for files extraction.

List matching flags:
  -s           Only match applets found on the applet server.
  -l           Only match applets found locally.

List display flags:
  -json        Print applet package data in JSON format.
  -f template  Set a specific format for the list, with the go template syntax.


You can use a template to format the result the way you want.
For example, to have just the applet name: -f '{{.DisplayedName}}'.
Everything is possible like '{{.DisplayedName}}  by {{.Author}}'.

For the full list of fields, see AppletPackage:
  http://godoc.org/github.com/sqp/godock/libs/packages#AppletPackage


Upload files or text to one-click hosting services

Usage:

	cdc upload fileorstring

Upload send data (raw text or file) to a one-click hosting service.

If your data start with / or file:// the file content will be sent.
Otherwise, the data is sent as raw text.

Note that you must have an active instance of the NetActivity applet.


Build cairo-dock sources

Usage:

	cdc build [-k] [-r] [-h] target

Build builds and install Cairo-Dock sources.

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
