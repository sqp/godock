/*
Package godock is a library to build Cairo-Dock applets with the DBus connector.

This API just started to exist, so it will evolve, but not that much. You can
already consider starting an applet using it. All the base actions and callbacks
are provided by Cairo-Dock, and are in stable state, so everything provided will
remain in almost the same format. Only the name, args, of way to access methods
may evolve in the Golang implementation, but only to better suits the needs and
provide a great Cairo-Dock Golang API. Those changes could be made to fix some
issues you may encounter while using it, so feel free to use it and post your
comments on the forum.

Cairo-Dock : http://glx-dock.org/

Usefull links:
	* Main Cairo-Dock Applet API, with everything to build an applet: http://github.com/sqp/godock/libs/dock
	* The DBus connector, the methods to interact with the dock: http://github.com/sqp/godock/libs/dbus
	* The default types and events defined by the dock API: http://github.com/sqp/godock/libs/cdtype
Could also help applet developers:
	* A simple logging system, to get everything consistent: http://github.com/sqp/godock/libs/log
	* The common applet polling system, helps you handle a regular task: http://github.com/sqp/godock/libs/poller

Cairo-Dock Wiki:
	* DBus methods: http://www.glx-dock.org/ww_page.php?p=Control_your_dock_with_DBus&lang=en
	* DBus for applets: http://www.glx-dock.org/ww_page.php?p=Documentation&lang=en
This last link is mainly the python documentation, but could be considered as
almost valid for Go. It will be imported when the API will get a more stable look.

Copyright (C) 2012-2014 SQP  <sqp@glx-dock.org>
*/
package godock
