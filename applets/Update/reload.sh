#!/bin/sh
#~ make
#DIR_APPLET="~/.config/cairo-dock/third-party/"
#go build -o $DIR_APPLET$1 github.com/sqp/godock/applets/$1/src &&
dbus-send --session --dest=org.cairodock.CairoDock /org/cairodock/CairoDock org.cairodock.CairoDock.ActivateModule string:$1 boolean:false &&
dbus-send --session --dest=org.cairodock.CairoDock /org/cairodock/CairoDock org.cairodock.CairoDock.ActivateModule string:$1 boolean:true
