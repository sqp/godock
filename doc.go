/*
Package godock is a library to build Cairo-Dock applets with the DBus connector.

External applets are small programs spawned by the dock to manage dedicated icons.
Once started, the program can use his icon to display informations and receive
user interactions like mouse events and keyboard shortcuts.

The project is now also able to build a reworked version of cairo-dock. It's a
binding for the C gldi library, the heart of the dock, and a rewrite in go of
all the dock code that wasn't in that library (loader, events, menu, the big
GUI...). Besides a few missing options, it should be usable by advanced users,
as it has been the only dock/panel used by its dev for months.

The goals of the project is to remain as much compatible with the original dock
as possible, at least on the global behavior and for configuration files.
But it will try to improve the user experience with a new GUI more focused on
current needs.

Install packages

Pre-build packages for Debian, Ubuntu and Archlinux are available.

    https://software.opensuse.org//download.html?project=home%3ASQP%3Acairo-dock-go&package=cairo-dock-rework

Install from source

Documentation for install, build and package creation is on the dist page, with
pre build packages repositories (Debian, Ubuntu, Archlinux).

    http://godoc.org/github.com/sqp/godock/dist

Usage

Once installed (more than an applet), you can run the cdc command to start a
dock (if enabled), or interact with the dock and its new applets.

    http://godoc.org/github.com/sqp/godock/cmd/cdc

Applet creation and documentation

If you want to create a new applet or learn more about them, most of their work
is defined in the cdtype package, with their common types, actions, events...

    http://godoc.org/github.com/sqp/godock/libs/cdtype

Current golang thread on cairo-dock forum: http://glx-dock.org/bg_topic.php?t=7638

Cairo-Dock : http://glx-dock.org/


Copyright (C) 2012-2016 SQP  <sqp@glx-dock.org>

License: Released under GPLv3
*/
package godock
