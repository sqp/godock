#!/bin/sh
# use module name as argument to reload it.
dbus-send --session --dest=org.cairodock.CairoDock /org/cairodock/CairoDock org.cairodock.CairoDock.ActivateModule string:$1 boolean:false &&
dbus-send --session --dest=org.cairodock.CairoDock /org/cairodock/CairoDock org.cairodock.CairoDock.ActivateModule string:$1 boolean:true
