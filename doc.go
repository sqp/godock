/*
Package godock is a library to build Cairo-Dock applets with the DBus connector.

External applets are small programs spawned by the dock to manage dedicated icons.
Once started, the program can use his icon to display informations and receive
user interactions like mouse events and keyboard shortcuts.

The best way to get started is to install the cdc command which also handles applets:

	go get -u -tags 'DiskActivity DiskFree GoGmail NetActivity Update' github.com/sqp/godock/cmd/cdc

Then you can install the applets you need (you may have to restart your dock).

	cd $GOPATH/src/github.com/sqp/godock/applets/DiskActivity
	make link

	cd $GOPATH/src/github.com/sqp/godock/applets/DiskFree
	make link

	cd $GOPATH/src/github.com/sqp/godock/applets/GoGmail
	make link

	cd $GOPATH/src/github.com/sqp/godock/applets/NetActivity
	make link

	cd $GOPATH/src/github.com/sqp/godock/applets/Update
	make link

If you want to create a new applet or learn more about them, follow those links:
	Main Cairo-Dock Applet API               http://godoc.org/github.com/sqp/godock/libs/dock
	DBus methods to interact with the dock   http://godoc.org/github.com/sqp/godock/libs/appdbus
	Common types and applet events           http://godoc.org/github.com/sqp/godock/libs/cdtype
	Applets code repository (real examples)  http://github.com/sqp/godock/tree/master/services
	Applets data repository (config files)   http://github.com/sqp/godock/tree/master/applets

Current golang thread on cairo-dock: http://glx-dock.org/bg_topic.php?t=7638

Cairo-Dock : http://glx-dock.org/


Copyright (C) 2012-2014 SQP  <sqp@glx-dock.org>
*/
package godock
